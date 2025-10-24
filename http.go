/*
 *  ZEUS - An Electrifying Build System
 *  Copyright (c) 2017 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"embed"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/desertbit/glue"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

//go:embed frontend/dist
var distFS embed.FS

var (
	webInterfaceRunning      bool
	webInterfaceRunningMutex = &sync.Mutex{}

	hostName        = "localhost"
	glueServer      *glue.Server
	glueServerMutex = &sync.Mutex{}

	socketstore      *SocketStore
	socketstoreMutex = &sync.Mutex{}

	// ErrReadingRandomString means reading the random data failed
	ErrReadingRandomString = errors.New("failed to read random data")
)

// router for the REST API
func createRouter() *httprouter.Router {

	r := httprouter.New()
	r.HandlerFunc("GET", "/files/:type/:file", serveFiles)
	r.HandlerFunc("GET", "/", serveHTTP)
	r.HandlerFunc("GET", "/quit", quitHandler)
	r.HandlerFunc("GET", "/wiki", wikiIndexHandler)
	r.HandlerFunc("GET", "/wiki/docs/:doc", wikiDocsHandler)
	r.HandlerFunc("GET", "/glue/ws", glueWebSocketHandler)
	r.HandlerFunc("POST", "/glue/ajax", glueAjaxHandler)

	return r
}

// glue handler for web sockets
var glueWebSocketHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	glueServer.ServeHTTP(w, r)
})

// glue handler for ajax requests
var glueAjaxHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	glueServer.ServeHTTP(w, r)
})

// StartWebListener fires up the web interface
func StartWebListener(openInBrowser bool) {

	var cLog = Log.WithField("prefix", "StartWebListener")

	// check if its already running
	webInterfaceRunningMutex.Lock()
	if webInterfaceRunning {
		webInterfaceRunningMutex.Unlock()

		if openInBrowser {
			if runtime.GOOS == "darwin" {
				open("http://" + hostName + ":" + strconv.Itoa(conf.fields.PortWebPanel))
			}
			return
		}
		return
	}
	webInterfaceRunning = true
	webInterfaceRunningMutex.Unlock()

	showNote("serving on "+strconv.Itoa(conf.fields.PortWebPanel), "starting server...")

	socketstoreMutex.Lock()
	socketstore = NewSocketStore()
	socketstoreMutex.Unlock()

	go runGlue()

	// init router
	r := createRouter()

	conf.Lock()
	if conf.fields.Debug {
		// start asset watchers for development
		go startJSWatcher()
		go startSassWatcher()
	}
	conf.Unlock()

	// listen and serve
	err := http.ListenAndServe(":"+strconv.Itoa(conf.fields.PortWebPanel), r)
	if err != nil {
		cLog.WithError(err).Error("failed to listen")
	}
}

// serve index page
var serveHTTP = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "invalid method, only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	c, err := distFS.ReadFile("frontend/dist/html/index.html")
	if err != nil {
		Log.WithError(err).Error("failed to serve index page")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	w.WriteHeader(200)
	w.Write(c)
})

// serve assets
var serveFiles = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	// get file name
	fileName := strings.TrimPrefix(r.RequestURI, "/files/")
	if strings.Contains(fileName, "?") {
		slice := strings.Split(fileName, "?")
		fileName = slice[0]
	}

	Log.Debug("serveFiles: ", fileName)

	// get file contents from embedded FS
	b, err := distFS.ReadFile("frontend/dist/" + fileName)
	if err != nil {
		Log.WithError(err).Error("unknown file")
		w.WriteHeader(404)
		w.Write([]byte("Not found"))
		return
	}

	// set Content-Length
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))

	// set Content-Type
	switch {
	case strings.HasPrefix(r.RequestURI, "/files/js"):
		w.Header().Set("Content-Type", "text/javascript")
	case strings.HasPrefix(r.RequestURI, "/files/css"):
		w.Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(r.RequestURI, ".mp4"):
		w.Header().Set("Content-Type", "video/mp4")
	case strings.HasSuffix(r.RequestURI, ".mov"):
		w.Header().Set("Content-Type", "video/quicktime")
	default:
		w.Header().Set("Content-Type", http.DetectContentType(b))
		Log.Debug("URI=", r.RequestURI, " CONTENT_TYPE=", http.DetectContentType(b))
	}

	w.WriteHeader(200)
	w.Write(b)
})

// run the glue server
func runGlue() {

	glueServerMutex.Lock()

	// create a new glue server
	glueServer = glue.NewServer(glue.Options{
		HTTPListenAddress: ":" + strconv.Itoa(conf.fields.PortGlueServer),
	})

	// release the glue server on defer
	// This will block new incoming connections
	// and close all current active sockets
	defer glueServer.Release()

	// set the glue event function to handle new incoming socket connections
	glueServer.OnNewSocket(socketstore.OnNewSocket)

	glueServerMutex.Unlock()

	// start
	err := glueServer.Run()
	if err != nil {
		Log.WithError(err).Error("glue server failed")
	}
}

/*
 *	Development
 */

// watch JS files and minify on changes
func startJSWatcher() {

	var (
		id   = processID(randomString())
		cLog = Log.WithFields(logrus.Fields{
			"prefix":    "startJSWatcher",
			"processID": id,
		})
		cmd = exec.Command("jsobfus",
			"-w",
			"frontend/src/js/:frontend/dist/js",
		)
	)
	cLog.Info("starting JS watcher")

	err := cmd.Start()
	if err != nil {
		cLog.WithError(err).Error("JavaScript watcher failed")
		return
	}

	addProcess(id, "jswatcher", cmd.Process, cmd.Process.Pid)
	defer deleteProcess(id)

	err = cmd.Wait()
	if err != nil {
		cLog.WithError(err).Error("js watcher failed")
	}
}

// start sass watcher and generate css on change
func startSassWatcher() {

	var (
		id   = processID(randomString())
		cLog = Log.WithFields(logrus.Fields{
			"prefix":    "startSassWatcher",
			"processID": id,
		})
		cmd = exec.Command("sasscompile",
			"-w",
			"frontend/src/sass:frontend/dist/css",
		)
	)

	cLog.Info("starting sass watcher")

	err := cmd.Start()
	if err != nil {
		cLog.WithError(err).Error("sass watcher failed")
		return
	}

	addProcess(id, "sasswatcher", cmd.Process, cmd.Process.Pid)
	defer deleteProcess(id)

	err = cmd.Wait()
	if err != nil {
		cLog.WithError(err).Error("sass watcher failed")
	}
}

// handle /quit route and exit application
var quitHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "invalid method, only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	Log.Info("exiting, Bye.")

	w.WriteHeader(200)
	w.Write([]byte("Bye."))

	showNote("Bye.", "stopping server and cleaning up")
	cleanup(nil)

	err := rl.Close()
	if err != nil {
		Log.WithError(err).Error("failed to close readline")
	}
	os.Exit(0)
})
