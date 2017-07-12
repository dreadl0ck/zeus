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

	yaml "gopkg.in/yaml.v1"
)

// bootstrap basic zeus setup
// useful when starting from scratch
func runBootstrapCommand() {

	err := os.MkdirAll(scriptDir, 0700)
	if err != nil {
		Log.WithError(err).Fatal("failed to create zeus directory")
	}

	f, err := os.Create(commandsFilePath)
	if err != nil {
		Log.WithError(err).Fatal("failed to create CommandsFile")
	}
	defer f.Close()
	f.WriteString(asciiArtYAML + `

# default language
language: bash

# globals for all commands
globals:

# command data
commands:
    
    # build the binary
    build:
        description: build project
        dependencies:
            - clean
        buildNumber: true
        exec: |
            echo "build the binary"

    # clean up the mess
    clean:
        description: clean up to prepare for build
        exec: rm -rf bin/*
    
    # perform install
    install:
        dependencies:
            - clean
        description: install to $PATH
        help: Install the application to the default system location
        exec: |
            echo "perform install"
`)
}

func printCreateCommandUsageErr() {
	l.Println("usage:")
	l.Println("zeus create [<language> <commandName>] [script <all> | <commandName>]")
}

// bootstrap a single new command
// either append to CommandsFile or create a new script
// then drop into editor
func handleCreateCommand(args []string) {

	if len(args) < 3 {
		printCreateCommandUsageErr()
		return
	}

	switch args[1] {

	// create script <all> | <commandName>
	case "script":
		switch args[2] {
		case "all":
			err := createAllScripts()
			if err != nil {
				Log.WithError(err).Error("failed to create scripts from commands.yml")
			}
		default:
			// check if command exists
			cmdMap.Lock()
			if cmd, ok := cmdMap.items[args[2]]; ok {
				cmdMap.Unlock()

				// get commandData from commandsFile
				var commandsFile = newCommandsFile()

				contents, err := ioutil.ReadFile(commandsFilePath)
				if err != nil {
					l.Println("unable to read " + commandsFilePath + ": " + err.Error())
					return
				}

				err = yaml.Unmarshal(contents, commandsFile)
				if err != nil {
					l.Println("failed to unmarshal commandsFile: " + err.Error())
					return
				}

				if d, ok := commandsFile.Commands[args[2]]; ok {
					err = createScript(d, cmd.name)
					if err != nil {
						l.Println("failed to create script: " + err.Error())
					}
				} else {
					l.Println("could not find " + args[2] + " in commandsFile: " + err.Error())
				}

				err = stripExecSectionFromCommandsFile(cmd.name)
				if err != nil {
					l.Println("failed to strip exec section from commandsFile: " + err.Error())
				}
				return
			}
			cmdMap.Unlock()
			l.Println("command " + args[2] + " does not exist!")
			return
		}

	// create <lang> <commandName>
	default:
		lang, err := ls.getLang(args[1])
		if err != nil {
			l.Println("getLang: ", err)
			return
		}

		// check if command exists
		cmdMap.Lock()
		if _, ok := cmdMap.items[args[2]]; ok {
			cmdMap.Unlock()
			l.Println("command " + args[2] + " exists!")
			return
		}
		cmdMap.Unlock()

		// check if there's a CommandsFile
		_, err = os.Stat(commandsFilePath)
		if err == nil {

			// append command to CommandsFile
			f, err := os.OpenFile(commandsFilePath, os.O_APPEND|os.O_WRONLY, 0744)
			if err != nil {
				Log.WithError(err).Error("failed to open CommandsFile for writing")
				return
			}

			f.WriteString("\n" + `    ` + args[2] + `:
        language: ` + lang.Name + `
        description:
        help:
        arguments:
        dependencies:
        outputs:
        exec: implement ` + args[2] + ` command
`)

			err = f.Close()
			if err != nil {
				l.Println(err)
				return
			}
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

			f.WriteString(lang.Bang + "\n")

			err = f.Close()
			if err != nil {
				l.Println(err)
				return
			}

			l.Println("created zeus command at " + filename)
		}

		// the WRITE event will cause the commandsFile to be parsed again - this happens async
		// lets wait a little bit...
		time.Sleep(120 * time.Millisecond)

		// start editor
		handleEditCommand([]string{"edit", args[2]})
	}
}

func stripExecSectionFromCommandsFile(commandName string) error {

	c, err := ioutil.ReadFile(commandsFilePath)
	if err != nil {
		return err
	}

	var (
		commandStarted               bool
		execStarted                  bool
		b                            bytes.Buffer
		offsetCommandNamesAndGlobals int
	)

	// iterate over contents line by line
	for _, line := range strings.Split(string(c), "\n") {

		if strings.Contains(line, commandName+":") {
			commandStarted = true
		}

		if offsetCommandNamesAndGlobals == 0 {
			if commandStarted {
				offsetCommandNamesAndGlobals = countLeadingSpace(line)
			}
		}

		if commandStarted {

			// check if next command started, for every line after the command started
			if !strings.Contains(line, commandName+":") {
				leadingSpace := countLeadingSpace(line)
				if leadingSpace == offsetCommandNamesAndGlobals {
					commandStarted = false
					b.WriteString(line + "\n")
					continue
				}
			}

			field := extractYAMLField(line)
			if field == "exec" {
				execStarted = true
			}

			// check if a new YAML field started
			if execStarted && field != "" && field != "exec" {
				execStarted = false
			}

			if execStarted {
				// skip line
				continue
			}
		}

		b.WriteString(line + "\n")
	}

	f, err := os.OpenFile(commandsFilePath, os.O_WRONLY|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer f.Close()

	blockWriteEvent()

	// update commandsFile
	f.WriteString(b.String())

	return nil
}
