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
	// default path for Zeusfile
	zeusfilePath = "zeus/Zeusfile.yml"

	// ErrFailedToReadZeusfile occurs when the Zeusfile could not be read
	ErrFailedToReadZeusfile = errors.New("failed to read Zeusfile")
)

// Zeusfile contains globals and commands for the Zeusfile.yml
type Zeusfile struct {

	// Overrride default language bash
	Language string `yaml:"language"`

	// global vars for all commands
	Globals *globals `yaml:"globals"`

	// command data
	Commands map[string]*commandData `yaml:"commands"`
}

func newZeusfile() *Zeusfile {
	return &Zeusfile{
		Language: "bash",
		Globals:  &globals{},
		Commands: make(map[string]*commandData, 0),
	}
}

// parse and initialize all commands inside the Zeusfile
func parseZeusfile(path string) error {

	var (
		start    = time.Now()
		zeusfile = newZeusfile()
		p        *parser
		ok       bool
	)

	// read file contents
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		Log.Debug(err)
		return ErrFailedToReadZeusfile
	}

	// unmarshal YAML
	err = yaml.Unmarshal(contents, zeusfile)
	if err != nil {
		return err
	}

	ps.Lock()
	if p, ok = ps.items[zeusfile.Language]; !ok {
		return errors.New("Zeusfile: " + ErrUnsupportedLanguage.Error() + ": " + zeusfile.Language)
	}
	ps.Unlock()

	// initialize commands
	for name, d := range zeusfile.Commands {

		// create parse job
		var job = p.AddJob("zeusfile."+name, false)

		chain, err := job.parseCommandChain(d.Dependencies)
		if err != nil {
			return err
		}

		// get build chain
		commandChain, err := job.getCommandChain(chain, zeusfile)
		if err != nil {
			return errors.New("Zeusfile, command " + name + ": " + err.Error())
		}

		// assemble commands args
		args, err := validateArgs(d.Arguments)
		if err != nil {
			return errors.New("Zeusfile, command " + name + ": " + err.Error())
		}

		var lang string
		if d.Language == "" {
			lang = zeusfile.Language
		} else {
			lang = d.Language
		}

		// create command
		cmd := &command{
			path:        "",
			name:        name,
			args:        args,
			description: d.Description,
			help:        d.Help,
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
			dependencies: commandChain,
			outputs:      d.Outputs,
			execScript:   d.Exec,
			async:        d.Async,
			language:     lang,
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

		Log.WithField("prefix", "parseZeusfile").Debug("added " + cp.CmdName + cmd.name + ansi.Reset + " to the command map")
		if debug {
			cmd.dump()
		}
	}

	cmdMap.Lock()
	defer cmdMap.Unlock()

	// only print info when using the interactive shell
	if len(os.Args) == 1 {
		l.Println(cp.Text+"initialized "+cp.Prompt, len(cmdMap.items), cp.Text+" commands from Zeusfile in: "+cp.Prompt, time.Now().Sub(start), ansi.Reset+"\n")
	}

	// watch file for changes in interactive mode
	if conf.fields.Interactive {
		go watchZeusfile(path, "")
	}

	return nil
}

// watch zeus file for changes and parse again
func watchZeusfile(path, eventID string) {

	// don't add a new watcher when the event exists
	projectData.Lock()
	for _, e := range projectData.fields.Events {
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
	for id, e := range projectData.fields.Events {
		if e.Name == "zeusfile watcher" {
			eventID = id
		}
	}
	projectData.Unlock()

	removeEvent(eventID)

	// parse file
	var (
		start    = time.Now()
		zeusfile = newZeusfile()
		p        *parser
		ok       bool
	)

	contents, err := ioutil.ReadFile(zeusfilePath)
	if err != nil {
		l.Println("couldnt find Zeusfile.yml")
		return err
	}

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
	if len(zeusfile.Globals.Items) > 0 {

		// create globals
		f, err := os.Create(zeusDir + "/globals.yml")
		if err != nil {
			return err
		}

		g.Lock()
		c, err := yaml.Marshal(g)
		if err != nil {
			g.Unlock()
			Log.Error("failed to marshal globals")
			return err
		}
		g.Unlock()

		f.Write(c)
		f.Close()
	}

	ps.Lock()
	if p, ok = ps.items[zeusfile.Language]; !ok {
		ps.Unlock()
		return errors.New("Zeusfile: " + ErrUnsupportedLanguage.Error() + ": " + zeusfile.Language)
	}
	ps.Unlock()

	// initialize commands
	for name, d := range zeusfile.Commands {

		// create parse job
		var job = p.AddJob("zeusfile."+name, false)

		chain, err := job.parseCommandChain(d.Dependencies)
		if err != nil {
			return err
		}

		// validate build chain
		_, err = job.getCommandChain(chain, zeusfile)
		if err != nil {
			return err
		}

		// check commands args
		_, err = validateArgs(d.Arguments)
		if err != nil {
			return err
		}

		// job done
		p.RemoveJob(job)

		// create command script
		f, err := os.Create(scriptDir + "/" + name + p.language.FileExtension)
		if err != nil {
			return err
		}

		f.WriteString("#!/bin/bash\n\n")
		f.WriteString("# " + zeusHeaderTag + "\n")

		f.WriteString("# arguments:\n")
		for _, arg := range d.Arguments {
			f.WriteString("#     - " + arg + "\n")
		}

		f.WriteString("# description: " + d.Description + "\n")
		f.WriteString("# dependencies: " + d.Dependencies + "\n")

		f.WriteString("# outputs:\n")
		for _, out := range d.Outputs {
			f.WriteString("#     - " + out + "\n")
		}

		if d.BuildNumber == true {
			f.WriteString("# buildNumber: true\n")
		}
		if d.Async == true {
			f.WriteString("# async: true\n")
		}

		f.WriteString("# help: " + d.Help + "\n")
		f.WriteString("# " + zeusHeaderTag + "\n\n")
		f.WriteString(d.Exec + "\n")
		f.Close()
	}

	cmdMap.Lock()
	defer cmdMap.Unlock()

	l.Println(cp.Text+"migrated "+cp.Prompt, len(cmdMap.items), cp.Text+" commands from Zeusfile in: "+cp.Prompt, time.Now().Sub(start), ansi.Reset+"\n")

	return os.Rename(zeusfilePath, zeusDir+"/Zeusfile_old.yml")
}
