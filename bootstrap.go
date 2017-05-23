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

import "os"
import "time"

/*
 *	Bootstrapping
 */

// create a file with name and content
func bootstrapScript(name string) {

	var cLog = Log.WithField("prefix", "bootstrapFile")

	cLog.Info("creating file: ", name)
	f, err := os.Create(scriptDir + "/" + name)
	if err != nil {
		cLog.WithError(err).Fatal("failed to create file: ", scriptDir+"/"+name)
	}
	defer f.Close()

	f.WriteString(`#!/bin/bash

# {zeus}
# description: description for ` + name + `
# arguments:
# dependencies:
# outputs:
# help: help text for ` + name + `
# {zeus}

echo "implement ` + name + `"
`)

	return
}

func printBootstrapCommandUsageErr() {
	l.Println("usage: zeus bootstrap <file | dir>")
}

// bootstrap basic zeus scripts
// useful when starting from scratch
func runBootstrapDirCommand() {

	err := os.MkdirAll(scriptDir, 0700)
	if err != nil {
		Log.WithError(err).Fatal("failed to create zeus directory")
	}

	bootstrapScript("clean.sh")
	bootstrapScript("build.sh")
	bootstrapScript("run.sh")
	bootstrapScript("test.sh")
	bootstrapScript("install.sh")
	bootstrapScript("bench.sh")
}

// bootstrap basic zeus scripts
// useful when starting from scratch
func runBootstrapZeusfileCommand() {

	f, err := os.Create("Zeusfile.yml")
	if err != nil {
		Log.WithError(err).Fatal("failed to create Zeusfile")
	}
	defer f.Close()

	f.WriteString(`############
# ZEUSFILE #
############

# globals for all build commands
globals:

# all commands
commands: 
    build:
        description: build project
        dependencies: clean
        buildNumber: true
        exec: 
	clean:
        description: clean up to prepare for build
        exec: rm -rf bin/*
    install:
        dependencies: clean
        description: install to $PATH
        help: Install the application to the default system location
        exec:
`)
}

func printCreateCommandUsageErr() {
	l.Println("usage:")
	l.Println("zeus create <language> <command>")
}

// bootstrap a single new command
// either append to Zeusfile or create a new script
// then drop into editor
func handleCreateCommand(args []string) {

	if len(args) < 3 {
		printCreateCommandUsageErr()
		return
	}

	cmdMap.Lock()

	// check if command exists
	if _, ok := cmdMap.items[args[2]]; ok {
		cmdMap.Unlock()
		l.Println("command " + args[2] + " exists!")
		return
	}
	cmdMap.Unlock()

	var lang *Language

	// look up language fileExtension and bang
	ps.Lock()
	for name, p := range ps.items {
		if name == args[1] {
			lang = p.language
		}
	}
	ps.Unlock()

	if lang == nil {
		l.Println("no parser for " + args[1])
		return
	}

	// check if there's a Zeusfile
	_, err := os.Stat(zeusfilePath)
	if err == nil {
		// append command to Zeusfile
		f, err := os.OpenFile(zeusfilePath, os.O_APPEND|os.O_WRONLY, 0744)
		if err != nil {
			Log.WithError(err).Error("failed to open Zeusfile for writing")
			return
		}

		f.WriteString("\n" + `    ` + args[2] + `:
        language: ` + lang.Name + `
        description:
        help:
        arguments:
        dependencies:
        outputs:
        exec:
`)
		f.Close()
	} else {
		filename := scriptDir + "/" + args[2] + lang.FileExtension

		// check if the script already exists
		_, err := os.Stat(filename)
		if err == nil {
			l.Println("file " + filename + " exists!")
			return
		}

		f, err := os.Create(filename)
		if err != nil {
			l.Println("failed to create file: ", err)
			return
		}

		f.WriteString(lang.Bang + `

# {zeus}
# description:
# arguments:
# dependencies:
# outputs:
# help:
# {zeus}

`)

		f.Close()
		l.Println("created zeus command at " + filename)

		// parse new command
		err = addCommand(filename, true)
		if err != nil {
			Log.WithError(err).Error("failed to add command")
		}
	}

	// parsing commands is async, wait a little
	time.Sleep(100 * time.Millisecond)

	// start editor
	handleEditCommand([]string{"edit", args[2]})
}
