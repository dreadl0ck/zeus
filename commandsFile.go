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
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mgutz/ansi"
	yaml "gopkg.in/yaml.v2"
)

var (
	// default path for CommandsFile
	commandsFilePath = zeusDir + "/commands.yml"

	// ErrFailedToReadCommandsFile occurs when the CommandsFile could not be read
	ErrFailedToReadCommandsFile = errors.New("failed to read CommandsFile")
)

// CommandsFile contains globals and commands for the CommandsFile.yml
type CommandsFile struct {

	// Overrride default language bash
	Language string `yaml:"language"`

	// global vars for all commands
	Globals map[string]string `yaml:"globals"`

	// command data
	Commands map[string]*commandData `yaml:"commands"`
}

func newCommandsFile() *CommandsFile {
	return &CommandsFile{
		Language: "bash",
		Globals:  make(map[string]string, 0),
		Commands: make(map[string]*commandData, 0),
	}
}

// parse and initialize all commands from the CommandsFile
func parseCommandsFile(path string) error {

	var (
		start        = time.Now()
		commandsFile = newCommandsFile()
	)

	// read file contents
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		Log.Debug(err)
		return ErrFailedToReadCommandsFile
	}

	// unmarshal YAML
	err = yaml.Unmarshal(contents, commandsFile)
	if err != nil {
		i, lineErr := extractLineNumFromError(err.Error(), "line")
		if lineErr == ErrNoLineNumberFound {
			i = -1
		} else if lineErr != nil {
			l.Println("failed to retrieve line number in which the error ocurred:", lineErr)
			i = -1
		}
		if !editorProcRunning {
			printCodeSnippet(string(contents), commandsFilePath, i)
		}
		return err
	}

	// validate
	err = validateCommandsFile(contents)
	if err != nil {
		return err
	}

	// check if language is supported
	_, err = ls.getLang(commandsFile.Language)
	if err != nil {
		return errors.New(commandsFilePath + ": " + err.Error() + ": " + ansi.Red + commandsFile.Language + cp.Text)
	}

	// flush command map
	cmdMap.flush()

	if len(commandsFile.Globals) > 0 {
		g = &globals{
			Vars: commandsFile.Globals,
		}
	}

	// initialize commands
	for name, d := range commandsFile.Commands {
		if d != nil {
			err = d.init(commandsFile, name)
			if err != nil {
				return errors.New("failed to init command: " + err.Error())
			}
		}
	}

	cmdMap.Lock()
	defer cmdMap.Unlock()

	// only print info when using the interactive shell
	if len(os.Args) == 1 {
		if !conf.fields.Quiet {
			l.Println(cp.Text+"initialized "+cp.Prompt, len(cmdMap.items), cp.Text+" commands from CommandsFile in: "+cp.Prompt, time.Now().Sub(start), cp.Reset+"\n")
		}
	}

	return nil
}

// look for invalid fields in commandsFile
func validateCommandsFile(c []byte) error {

	var (
		fields = []string{
			"description",
			"help",
			"language",
			"arguments",
			"dependencies",
			"outputs",
			"buildNumber",
			"async",
			"exec",
			"globals",
			"path",
			"commands",
		}
		parsedFields                 []string
		foundField                   bool
		globalsStarted               bool
		commandsStarted              bool
		globalNames                  []string
		commandNames                 []string
		offsetCommandNamesAndGlobals int
	)

	// iterate over contents line by line
	for i, line := range strings.Split(string(c), "\n") {

		// determine current section
		if strings.Contains(line, "globals:") {
			globalsStarted = true
			continue
		} else if strings.Contains(line, "commands:") {
			commandsStarted = true
			globalsStarted = false
			continue
		}

		if offsetCommandNamesAndGlobals == 0 {
			if globalsStarted || commandsStarted {
				offsetCommandNamesAndGlobals = countLeadingSpace(line)
			}
		}

		// validate names of commands and custom global variable names
		leadingSpace := countLeadingSpace(line)
		if leadingSpace == offsetCommandNamesAndGlobals {

			// catch duplicate global variable names
			if globalsStarted {
				field := extractYAMLField(line)
				if field != "" {
					for _, name := range globalNames {
						if field == name {
							printCodeSnippet(string(c), commandsFilePath, i)
							return errors.New("line " + strconv.Itoa(i) + ": duplicate global name detected: " + name)
						}
					}
					globalNames = append(globalNames, field)
				}
			}

			// catch duplicate command names
			if commandsStarted {
				field := extractYAMLField(line)
				if field != "" {

					for _, name := range commandNames {
						if field == name {
							printCodeSnippet(string(c), commandsFilePath, i)
							return errors.New("line " + strconv.Itoa(i) + ": duplicate command name detected: " + name)
						}
					}
					commandNames = append(commandNames, field)
					// next command started - reset parsedFields
					parsedFields = []string{}
				}
			}
			continue

		} else if leadingSpace > offsetCommandNamesAndGlobals*2 {
			// ignore everything that contains a colon inside the 'exec' field
			continue
		}

		// check for unknown and duplicate fields + invalid combinations
		field := extractYAMLField(line)
		if field != "" {

			// check for unknown fields
			for _, item := range fields {
				if field == strings.TrimSpace(string(item)) {
					foundField = true
				}
			}
			if !foundField {
				printCodeSnippet(string(c), commandsFilePath, i)
				return errors.New("line " + strconv.Itoa(i) + ": unknown field: " + field)
			}
			foundField = false

			// check duplicate fields
			for _, f := range parsedFields {
				if f == field {
					printCodeSnippet(string(c), commandsFilePath, i)
					return errors.New("line " + strconv.Itoa(i) + ": duplicate field: " + field)
				}
			}
			parsedFields = append(parsedFields, field)
		}
	}

	return nil
}

// count prefix whitespace characters of a string
func countLeadingSpace(line string) int {
	i := 0
	for _, runeValue := range line {
		if runeValue == ' ' {
			i++
		} else {
			break
		}
	}
	return i
}

// watch zeus file for changes and parse again
func watchCommandsFile(path, eventID string) {

	// don't add a new watcher when the event exists
	projectData.Lock()
	for _, e := range projectData.fields.Events {
		if e.Name == "commandsFile watcher" {
			projectData.Unlock()
			return
		}
	}
	projectData.Unlock()

	Log.Debug("watching commandsFile at ", path)

	err := addEvent(newEvent(path, fsnotify.Write, "commandsFile watcher", ".yml", eventID, "internal", func(e fsnotify.Event) {

		// without sleeping every line written to stdout has the length of the previous line as offset
		// sleeping at least 100 millisecs seems to work - strange
		time.Sleep(100 * time.Millisecond)
		l.Println()

		Log.Debug("received commandsFile WRITE event: ", e.Name)

		if !editorProcRunning {
			printProjectHeader()
		}

		err := parseCommandsFile(path)
		if !editorProcRunning {
			if err != nil {
				Log.WithError(err).Error("failed to parse commandsFile")
			} else {
				printCommands()
			}
		}
	}))
	if err != nil {
		Log.WithError(err).Error("failed to watch commandsFile")
	}
}

func createAllScripts() error {

	// parse file
	var (
		commandsFile = newCommandsFile()
	)

	contents, err := ioutil.ReadFile(commandsFilePath)
	if err != nil {
		l.Println("unable to read " + commandsFilePath)
		return err
	}

	err = yaml.Unmarshal(contents, commandsFile)
	if err != nil {
		return err
	}

	// create scriptDir if necessary
	if _, err = os.Stat(scriptDir); err != nil {
		err = os.MkdirAll(scriptDir, 0700)
		if err != nil {
			Log.WithError(err).Fatal("failed to create " + scriptDir + " directory")
		}
	}

	// check if language is valid
	_, err = ls.getLang(commandsFile.Language)
	if err != nil {
		return errors.New("CommandsFile: " + err.Error() + ": " + commandsFile.Language)
	}

	// write commands to disk
	for name, d := range commandsFile.Commands {
		if d.Path == "" {
			err = createScript(d, name)
			if err != nil {
				l.Println(err)
				continue
			}

			err = stripExecSectionFromCommandsFile(name)
			if err != nil {
				return errors.New("failed to strip exec section from commandsFile: " + err.Error())
			}
		} else {
			l.Println("skipping command " + name + " because it has a custom path set")
		}
	}

	return parseCommandsFile(commandsFilePath)
}

func createScript(d *commandData, name string) error {

	// set default language
	if d.Language == "" {
		d.Language = "bash"
	}

	lang, err := ls.getLang(d.Language)
	if err != nil {
		return errors.New(err.Error() + ": " + d.Language)
	}

	// check commands args
	_, err = validateArgs(d.Arguments)
	if err != nil {
		return err
	}

	scriptName := scriptDir + "/" + name + lang.FileExtension

	// make sure the file does not already exist
	_, err = os.Stat(scriptName)
	if err == nil {
		return errors.New(scriptName + " already exists!")
	}

	// create command script
	f, err := os.Create(scriptName)
	if err != nil {
		return err
	}

	f.WriteString(lang.Bang + "\n\n")
	f.WriteString(d.Exec + "\n")
	f.Close()

	// flush exec field
	d.Exec = ""

	return nil
}
