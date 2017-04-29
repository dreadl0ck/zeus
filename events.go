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
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
)

var (
	// ignore the the write event when updating the config using the shell command
	disableWriteEvent      = false
	disableWriteEventMutex = &sync.Mutex{}

	// ErrInvalidEventType means the given event type string is invalid
	ErrInvalidEventType = errors.New("invalid fsnotify event type. available types are: WRITE | REMOVE | RENAME | CHMOD")

	// ErrInvalidUsage means the command was used incorrectly
	ErrInvalidUsage = errors.New("invalid usage")
)

// Event represents a watched path, along with an an action
// that will be performed when an operation of the specified type occurs
type Event struct {

	// Event Name
	Name string

	// Event ID
	ID string

	// Path to watch
	Path string

	// Operation type
	Op fsnotify.Op

	// optional File Type Extension
	// if empty the event will be fired for all file types
	FileExtension string

	// Command to be executed upon event
	Command string

	// custom event handler func
	handler func(fsnotify.Event)

	// shutdown chan
	stopChan chan bool
}

func printEventsUsageErr() {
	l.Println(ErrInvalidUsage)
	l.Println("usage: events [add <optype> <path> <filetype> <commandChain>] [remove <path>]")
}

// handle events command
func handleEventsCommand(args []string) {

	if len(args) < 2 {
		listEvents()
		return
	}

	if len(args) < 3 {
		printEventsUsageErr()
		return
	}

	switch args[1] {
	case "remove":
		removeEvent(args[2])
	case "add":
		registerEvent(args)

	default:
		printEventsUsageErr()
	}
}

func registerEvent(args []string) {
	if len(args) < 5 {
		printEventsUsageErr()
		return
	}

	// check if event type is valid
	op, err := getEventType(args[2])
	if err != nil {
		Log.Error(err)
		return
	}

	// check if path exists
	_, err = os.Stat(args[3])
	if err != nil {
		Log.Error(err)
		return
	}

	var (
		fields   []string
		filetype string
	)

	if strings.HasPrefix(args[4], ".") {
		fields = args[5:]
		filetype = args[4]
	} else {
		fields = args[4:]
	}

	if filetype != "" && len(fields) == 0 {
		Log.Error("no command supplied")
		return
	}

	if validCommandChain(fields, true) {
		Log.Info("adding command chain")
	} else {
		Log.Info("adding shell command")
	}

	chain := strings.Join(fields, " ")
	go func() {
		e := newEvent(args[3], op, "custom event", filetype, "", chain, func(event fsnotify.Event) {

			Log.Debug("event fired, name: ", event.Name, " path: ", args[3])

			if validCommandChain(fields, true) {
				executeCommandChain(chain)
			} else {

				// its a shell command
				if len(fields) > 1 {
					passCommandToShell(fields[0], fields[1:])
				} else {
					passCommandToShell(fields[0], []string{})
				}
			}
		})
		err := addEvent(e)
		if err != nil {
			Log.Error("failed to watch path: ", args[3])
		}
	}()
}

// parse command type string and fsnotify type
func getEventType(event string) (fsnotify.Op, error) {

	switch event {
	case "WRITE":
		return fsnotify.Write, nil
	case "REMOVE":
		return fsnotify.Remove, nil
	case "RENAME":
		return fsnotify.Rename, nil
	case "CHMOD":
		return fsnotify.Chmod, nil
	default:
		return 0, ErrInvalidEventType
	}
}

// list all currently registered events
func listEvents() {

	w := 20

	l.Println(cp.prompt + pad("name", w) + pad("ID", w) + pad("operation", w) + pad("command", w) + pad("filetype", w) + pad("path", w))
	for _, e := range projectData.Events {
		l.Println(cp.text + pad(e.Name, w) + pad(e.ID, w) + pad(e.Op.String(), w) + pad(e.Command, w) + pad(e.FileExtension, w) + pad(e.Path, w))
	}
}

// remove the event for the given path
func removeEvent(id string) {

	projectData.Lock()

	// check if event exists
	if e, ok := projectData.Events[id]; ok {

		if e.stopChan != nil {
			// stop event handler
			e.stopChan <- true
		}

		// delete event
		delete(projectData.Events, id)
		projectData.Unlock()

		Log.Debug("removed event with name ", e.Name)

		// update project data
		projectData.update()
		return
	}
	projectData.Unlock()

	Log.Error("event with ID ", id, " does not exist")
}

// create a new event
// if the supplied eventID is empty it will be generated
func newEvent(path string, op fsnotify.Op, name, filetype, eventID, command string, handler func(fsnotify.Event)) *Event {

	if eventID == "" {
		eventID = randomString()
	}

	// create event
	return &Event{
		Path:          path,
		Name:          name,
		ID:            eventID,
		Op:            op,
		handler:       handler,
		stopChan:      make(chan bool, 1),
		Command:       command,
		FileExtension: filetype,
	}
}

// addEvent takes an event, registers it, creates a watcher for the events path
// and registers a handler that will fire if operation op occurs
func addEvent(e *Event) error {

	var cLog = Log.WithField("prefix", "addEvent")
	Log.WithField("path", e.Path).Debug("adding event")

	// init new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// add to events
	projectData.Lock()
	projectData.Events[e.ID] = e
	projectData.Unlock()

	// update projectData on disk
	projectData.update()

	// listen for events
	done := make(chan bool)
	go func() {

		for {
			select {
			case event := <-watcher.Events:

				// cLog.WithFields(logrus.Fields{
				// 	"event": event,
				// 	"path":  path,
				// }).Debug("incoming event")

				// check operation type
				if event.Op == e.Op {

					if e.FileExtension != "" {
						if !strings.HasSuffix(event.Name, e.FileExtension) {
							Log.Debug("ignoring event because file type does not match: ", event.Name)
							continue
						}
					}

					// check if write event was disabled.
					// example: when updating the config with the config command
					// revalidating the config is not necessary
					disableWriteEventMutex.Lock()
					if disableWriteEvent {
						disableWriteEvent = false
						disableWriteEventMutex.Unlock()
						cLog.Debug("ignoring WRITE event for path: ", e.Path)
						continue
					}
					disableWriteEventMutex.Unlock()

					// fire handler
					e.handler(event)
				}
			case err := <-watcher.Errors:
				cLog.WithError(err).Fatal("watcher failed")
			case _ = <-e.stopChan:
				watcher.Close()
				done <- true
				return
			}
		}
	}()

	// add path to watcher
	err = watcher.Add(e.Path)
	if err != nil {
		cLog.WithFields(logrus.Fields{
			"error": err,
			"path":  e.Path,
		}).Error("failed to add path to watcher")
		e.stopChan <- true
		return err
	}

	// wait for it
	<-done

	return nil
}

// reload an internal event from project data
func reloadEvent(e *Event) {

	Log.Debug("reloading event: ", e.Name)

	switch e.Name {
	case "config watcher":
		go conf.watch(e.ID)
	case "formatter watcher":
		if conf.AutoFormat {
			go f.watchzeusDir(e.ID)
		}
	case "zeusfile watcher":
		if _, err := os.Stat(zeusfilePath); err != nil {
			go watchZeusfile("Zeusfile", e.ID)
		} else {
			go watchZeusfile(zeusfilePath, e.ID)
		}
	case "script watcher":
		go watchScripts(e.ID)
	default:
		Log.Warn("reload event called for an unknown event: ", e.Name)
	}
}
