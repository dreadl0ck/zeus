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
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/dreadl0ck/readline"
	"github.com/fsnotify/fsnotify"
	"github.com/mgutz/ansi"
	"gopkg.in/yaml.v2"
)

var (
	zeusfilePath = "Zeusfile.yml"

	// ErrFailedToReadZeusfile occurs when the Zeusfile could not be read
	ErrFailedToReadZeusfile = errors.New("failed to read Zeusfile")
)

// Zeusfile contains globals and commands for the Zeusfile.yml
type Zeusfile struct {
	Globals  string                  `yaml:"globals"`
	Commands map[string]*commandData `yaml:"commands"`
}

// parse and initialize all commands inside the Zeusfile
func parseZeusfile(path string) error {

	var (
		start    = time.Now()
		zeusfile = new(Zeusfile)
	)

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		Log.Debug(err)
		return ErrFailedToReadZeusfile
	}

	// replace all tabs with spaces
	contents = bytes.Replace(contents, []byte("\t"), []byte("    "), -1)

	err = yaml.Unmarshal(contents, zeusfile)
	if err != nil {
		return err
	}

	// check if there are globals
	if len(zeusfile.Globals) > 0 {
		globalsContent = append(globalsContent, []byte("#!/bin/bash\n"+zeusfile.Globals+"\n")...)
	}

	// initialize commands
	for name, d := range zeusfile.Commands {

		// create parse job
		var job = p.AddJob("zeusfile."+name, false)

		chain, err := job.parseCommandChain(d.Chain)
		if err != nil {
			return err
		}

		// get build chain
		commandChain, err := job.getCommandChain(chain, zeusfile)
		if err != nil {
			return errors.New("Zeusfile, command " + name + ": " + err.Error())
		}

		// assemble commands args
		args, err := validateArgs(d.Args)
		if err != nil {
			return errors.New("Zeusfile, command " + name + ": " + err.Error())
		}

		// create command
		cmd := &command{
			path:         "",
			name:         name,
			args:         args,
			manual:       d.Manual,
			help:         d.Help,
			commandChain: commandChain,
			PrefixCompleter: readline.PcItem(name,
				readline.PcItemDynamic(func(path string) (res []string) {
					for _, a := range args {
						if !strings.Contains(path, a.name+"=") {
							res = append(res, a.name+"=")
						}
					}
					return
				}),
			),
			buildNumber:  d.BuildNumber,
			dependencies: d.Dependencies,
			outputs:      d.Outputs,
			runCommand:   d.Run,
			async:        d.Async,
		}

		// job done
		p.RemoveJob(job)

		// Add the completer.
		completer.Lock()
		completer.Children = append(completer.Children, cmd.PrefixCompleter)
		completer.Unlock()

		// add to command map
		cmdMap.Lock()
		cmdMap.items[cmd.name] = cmd
		cmdMap.Unlock()

		Log.WithField("prefix", "parseZeusfile").Debug("added " + cp.cmdName + cmd.name + ansi.Reset + " to the command map")
		if debug {
			cmd.dump()
		}
	}

	cmdMap.Lock()
	defer cmdMap.Unlock()

	// only print info when using the interactive shell
	if len(os.Args) == 1 {
		l.Println(cp.text+"initialized "+cp.prompt, len(cmdMap.items), cp.text+" commands from Zeusfile in: "+cp.prompt, time.Now().Sub(start), ansi.Reset+"\n")
	}

	// watch file for changes in interactive mode
	if conf.Interactive {
		go watchZeusfile(path, "")
	}

	return nil
}

// watch zeus file for changes and parse again
func watchZeusfile(path, eventID string) {

	// dont add a new watcher when the event exists
	projectData.Lock()
	for _, e := range projectData.Events {
		if e.Name == "zeusfile watcher" {
			projectData.Unlock()
			return
		}
	}
	projectData.Unlock()

	Log.Debug("watching zeusfile at ", path)

	err := addEvent(newEvent(path, fsnotify.Write, "zeusfile watcher", "", eventID, "internal", func(e fsnotify.Event) {

		Log.Debug("received zeusfile event: ", e.Name)

		err := parseZeusfile(path)
		if err != nil {
			Log.WithError(err).Error("failed to parse zeusfile")
		}
	}))
	if err != nil {
		Log.WithError(err).Error("failed to watch zeusfile")
	}
}

func migrateZeusfile() error {

	// remove zeusfile watcher
	var eventID string

	// remove zeusfile watcher
	projectData.Lock()
	for id, e := range projectData.Events {
		if e.Name == "zeusfile watcher" {
			eventID = id
		}
	}
	projectData.Unlock()

	removeEvent(eventID)

	// parse file
	var (
		start    = time.Now()
		zeusfile = new(Zeusfile)
		filename string
	)

	contents, err := ioutil.ReadFile(zeusfilePath)
	if err != nil {
		contents, err = ioutil.ReadFile("Zeusfile")
		if err != nil {
			l.Println("couldnt find Zeusfile.yml or Zeusfile")
			return err
		}
		filename = "Zeusfile"
	} else {
		filename = zeusfilePath
	}

	// replace all tabs with spaces
	contents = bytes.Replace(contents, []byte("\t"), []byte("    "), -1)

	err = yaml.Unmarshal(contents, zeusfile)
	if err != nil {
		return err
	}

	// create zeus dir if necessary
	if _, err = os.Stat(zeusDir); err != nil {
		err = os.Mkdir(zeusDir, 0700)
		if err != nil {
			Log.WithError(err).Fatal("failed to create zeus directory")
		}
	}

	// check if there are globals
	if len(zeusfile.Globals) > 0 {
		// create globals.sh
		f, err := os.Create(zeusDir + "/globals.sh")
		if err != nil {
			return err
		}
		defer f.Close()
	}

	// initialize commands
	for name, d := range zeusfile.Commands {

		// create parse job
		var job = p.AddJob("zeusfile."+name, false)

		chain, err := job.parseCommandChain(d.Chain)
		if err != nil {
			return err
		}

		// validate build chain
		_, err = job.getCommandChain(chain, zeusfile)
		if err != nil {
			return err
		}

		// check commands args
		_, err = validateArgs(d.Args)
		if err != nil {
			return err
		}

		// job done
		p.RemoveJob(job)

		// create command script
		f, err := os.Create(zeusDir + "/" + name + ".sh")
		if err != nil {
			return err
		}

		f.WriteString("#!/bin/bash\n\n")
		f.WriteString("# ------------------------------------------------ #\n")
		f.WriteString("# @zeus-args: " + d.Args + "\n")
		f.WriteString("# @zeus-help: " + d.Help + "\n")
		f.WriteString("# @zeus-chain: " + d.Chain + "\n")
		f.WriteString("# @zeus-deps: " + strings.Join(d.Dependencies, ", ") + "\n")
		f.WriteString("# @zeus-outputs: " + strings.Join(d.Outputs, ", ") + "\n")
		if d.BuildNumber == true {
			f.WriteString("# @zeus-buildNumber: true\n")
		}
		f.WriteString("# ------------------------------------------------ #\n")
		f.WriteString("#" + d.Manual + "\n")
		f.WriteString("# ------------------------------------------------ #\n\n")
		f.WriteString(d.Run + "\n")
		f.Close()
	}

	cmdMap.Lock()
	defer cmdMap.Unlock()

	l.Println(cp.text+"migrated "+cp.prompt, len(cmdMap.items), cp.text+" commands from Zeusfile in: "+cp.prompt, time.Now().Sub(start), ansi.Reset+"\n")

	return os.Rename(filename, "Zeusfile_old.yml")
}
