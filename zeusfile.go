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
		return err
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
		var job = p.AddJob(randomString(), false)

		// get build chain
		commandChain, err := job.getCommandChain(parseCommandChain(d.Chain), zeusfile)
		if err != nil {
			return err
		}

		// assemble commands args
		args, err := validateArgs(d.Args)
		if err != nil {
			return err
		}

		// create command
		cmd := &command{
			path:            "",
			name:            name,
			args:            args,
			manual:          d.Manual,
			help:            d.Help,
			commandChain:    commandChain,
			PrefixCompleter: readline.PcItem(name),
			buildNumber:     d.BuildNumber,
			dependencies:    d.Dependencies,
			outputs:         d.Outputs,
			runCommand:      d.Run,
		}

		// job done
		p.RemoveJob(job)

		commandMutex.Lock()

		// add parameter labels to completer
		for _, arg := range cmd.args {
			cmd.PrefixCompleter.Children = append(cmd.PrefixCompleter.Children, readline.PcItem(arg.name+"="))
		}

		// Add the completer.
		completer.Children = append(completer.Children, cmd.PrefixCompleter)

		// add to command map
		commands[cmd.name] = cmd
		commandMutex.Unlock()

		Log.Debug("added " + cmd.name + " to the command map")
	}

	l.Println(cp.colorText+"initialized "+cp.colorPrompt, len(commands), cp.colorText+" commands from Zeusfile in: "+cp.colorPrompt, time.Now().Sub(start), ansi.Reset+"\n")

	// watch file for changes in interactive mode
	if conf.Interactive {
		go watchZeusfile(path, "")
	}

	return nil
}

// watch zeus file for changes and parse again
func watchZeusfile(path, eventID string) {

	// dont add a new watcher when the event exists
	projectDataMutex.Lock()
	for _, e := range projectData.Events {
		if e.Name == "zeusfile watcher" {
			projectDataMutex.Unlock()
			return
		}
	}
	projectDataMutex.Unlock()

	Log.Debug("watching zeusfile at ", path)

	err := addEvent(newEvent(path, fsnotify.Write, "zeusfile watcher", "", eventID, "internal", func(e fsnotify.Event) {

		Log.Debug("received zeusfile event")

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
	projectDataMutex.Lock()
	for id, e := range projectData.Events {
		if e.Name == "zeusfile watcher" {
			eventID = id
		}
	}
	projectDataMutex.Unlock()

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
		} else {
			filename = "Zeusfile"
		}
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
		var job = p.AddJob(randomString(), false)

		// validate build chain
		_, err = job.getCommandChain(parseCommandChain(d.Chain), zeusfile)
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

	l.Println(cp.colorText+"migrated "+cp.colorPrompt, len(commands), cp.colorText+" commands from Zeusfile in: "+cp.colorPrompt, time.Now().Sub(start), ansi.Reset+"\n")

	return os.Rename(filename, "Zeusfile_old.yml")
}
