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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	ErrFailedToReadCommandsFile = errors.New("failed to read commandsFile")
)

// CommandsFile contains globals and commands for the main zeus configuration file commands.yml
type CommandsFile struct {

	// Override default language bash
	Language string `yaml:"language"`

	// global vars for all commands
	Globals map[string]string `yaml:"globals"`

	// command data
	Commands map[string]*commandData `yaml:"commands"`

	// script to call when starting zeus
	StartupHook string `yaml:"startupHook"`

	// script to call when exiting zeus
	ExitHook string `yaml:"exitHook"`

	// commandsFile that is extended by the current commandsFile.
	// commands from this file will be executed within the CURRENT zeus directory.
	Extends string `yaml:"extends"`

	// commandsFile that is extended by the current commandsFile.
	// commands from this file will be executed within the ORIGINAL zeus directory.
	Includes string `yaml:"includes"`
}

func newCommandsFile() *CommandsFile {
	return &CommandsFile{
		Language: "bash",
		Globals:  make(map[string]string, 0),
		Commands: make(map[string]*commandData, 0),
	}
}

// parse and initialize all commands from the CommandsFile
func parseCommandsFile(path string, flush bool) (*CommandsFile, error) {

	var (
		start        = time.Now()
		commandsFile = newCommandsFile()
	)

	// read file contents
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		Log.Debug(err)
		return nil, errors.New(ErrFailedToReadCommandsFile.Error() + ": " + err.Error())
	}

	// unmarshal YAML
	err = yaml.UnmarshalStrict(contents, commandsFile)
	if err != nil {
		i, lineErr := extractLineNumFromError(err.Error(), "line")
		if lineErr == ErrNoLineNumberFound {
			i = -1
		} else if lineErr != nil {
			l.Println("failed to retrieve line number in which the error occurred:", lineErr)
			i = -1
		}
		if !shellBusy {
			printCodeSnippet(string(contents), commandsFilePath, i)
		}
		return nil, err
	}

	// catch attempts to use includes and extends at the same time
	// TODO: add support for this in the future
	if commandsFile.Extends != "" && commandsFile.Includes != "" {
		return nil, errors.New("commandsFile options 'extends' and 'includes' can't be used in combination at the moment")
	}

	// check if language is supported
	_, err = ls.getLang(commandsFile.Language)
	if err != nil {
		return nil, errors.New(commandsFilePath + ": " + err.Error() + ": " + ansi.Red + commandsFile.Language + cp.Text)
	}

	if flush {
		// flush command map
		cmdMap.flush()
		g = &globals{
			Vars: make(map[string]string, 0),
		}
	}

	if len(commandsFile.Globals) > 0 {
		if g != nil {
			for k, v := range commandsFile.Globals {
				g.Vars[k] = v
			}
		} else {
			// init
			g = &globals{
				Vars: commandsFile.Globals,
			}
		}

		// resolve environment vars used in the globals
		// must be in format ${VAR}
		for k, v := range g.Vars {
			g.Vars[k] = resolveEnvironment(v)
		}
	}

	// initialize commands
	for name, d := range commandsFile.Commands {
		if d != nil {
			err = d.init(commandsFile, name)
			if err != nil {
				return nil, errors.New("failed to init command: " + err.Error())
			}
		}
	}

	// handle base configurations
	// since this allows commands to cross reference each other, this must be done after all commands have been initialized.
	cmdMap.Lock()
	defer cmdMap.Unlock()

	// get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// set working directory for all commands that are from the current commandsFile
	for name := range commandsFile.Commands {
		if cmd, ok := cmdMap.items[name]; ok {
			if cmd.workingDir == "" {
				cmd.workingDir = wd
			}
		}
	}

	for _, cmd := range cmdMap.items {
		if cmd.extends != "" {
			if baseCmd, ok := cmdMap.items[cmd.extends]; ok {

				// handle arguments
				// save old args
				oldArgs := cmd.args
				// overwrite args with base args
				cmd.args = baseCmd.args
				// add the args of the current command again. this will allow to overwrite args from the base if desired.
				for n, a := range oldArgs {
					cmd.args[n] = a
				}

				// prepend outputs from base command
				cmd.outputs = append(baseCmd.outputs, cmd.outputs...)

				// if no description is provided for the current command, use the one from the base command.
				if cmd.description == "" {
					cmd.description = baseCmd.description
				}

				// if no help text is provided for the current command, use the one from the base command.
				if cmd.help == "" {
					cmd.help = baseCmd.help
				}

				// prepend deps from base command
				cmd.dependencies = append(baseCmd.dependencies, cmd.dependencies...)

				cmd.canModifyPrompt = baseCmd.canModifyPrompt
				cmd.buildNumber = baseCmd.buildNumber
				cmd.language = baseCmd.language
				cmd.async = baseCmd.async

				// handle exec action
				if cmd.exec == "" && baseCmd.exec != "" {
					cmd.exec = baseCmd.exec
				}
				if cmd.path == "" && baseCmd.path != "" {
					cmd.path = baseCmd.path
				}
			} else {
				return nil, errors.New("base command not found: " + cmd.extends)
			}
		}
	}

	// only print info when using the interactive shell
	if len(os.Args) == 1 {
		if conf.fields.Debug {
			l.Println(cp.Text+"initialized "+cp.Prompt, len(cmdMap.items), cp.Text+" commands from CommandsFile in: "+cp.Prompt, time.Now().Sub(start), cp.Reset+"\n")
		}
	}

	// invoke startupHook if set
	if commandsFile.StartupHook != "" {
		out, err := exec.Command(commandsFile.StartupHook).CombinedOutput()
		if err != nil {

			// cleanup without calling the exitHook to avoid confusion
			// in case the startupHook fails, and the exitHook would fail as well due to that.
			cleanup(nil)

			fmt.Println(string(out))
			log.Fatal("startupHook failed: ", err)
		}
		modifyPrompt()
	}

	return commandsFile, nil
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

var lastCommandsFileError error

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
		//time.Sleep(100 * time.Millisecond)
		//l.Println()

		Log.Debug("received commandsFile WRITE event: ", e.Name)

		cmdFile, err := parseCommandsFile(path, true)
		if !shellBusy {
			if err != nil {
				// flush command map
				cmdMap.flush()
				g = &globals{
					Vars: make(map[string]string, 0),
				}
				Log.WithError(err).Error("failed to parse commandsFile")
			}
		} else {
			if err != nil {
				// flush command map
				cmdMap.flush()
				g = &globals{
					Vars: make(map[string]string, 0),
				}
				// shell is currently busy. store the error to present it to the user once the shell is free again.
				lastCommandsFileError = err
			} else {
				// commandsFile was parsed successfully in the background. Make sure previous error is cleared.
				lastCommandsFileError = nil
			}
		}

		if cmdFile != nil {

			// handle commandsFile extension
			cmdFile.handleExtension()

			// handle commandsFile inclusion
			cmdFile.handleInclusion()
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
			err = commandsFile.createScript(d, name)
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

	_, err = parseCommandsFile(commandsFilePath, true)
	return err
}

func (c *CommandsFile) createScript(d *commandData, name string) error {

	// set default language
	if d.Language == "" {
		d.Language = "bash"
	}

	lang, err := ls.getLang(d.Language)
	if err != nil {
		return errors.New(err.Error() + ": " + d.Language)
	}

	// check commands args
	_, err = c.validateArgs(d.Arguments)
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

func (c *CommandsFile) handleExtension() {
	// handle commandsFile extension
	if c.Extends != "" {

		// get current working directory
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		// change into extension directory
		p := strings.TrimSuffix(filepath.Dir(c.Extends), "zeus")
		err = os.Chdir(p)
		if err != nil {
			log.Fatal(err)
		}

		// check if a CommandsFile for the project exists
		// parse and flush it, in order to use it as base for the actual commandsfile of this project
		_, err = parseCommandsFile(commandsFilePath, true)
		if err != nil {
			Log.Error("failed to parse commandsFile: ", err, "\n")
			os.Exit(1)
		}

		// move back into actual root directory
		err = os.Chdir(wd)
		if err != nil {
			log.Fatal(err)
		}

		// update workingDirs on loaded commands to point to the current directory
		for _, cmd := range cmdMap.items {
			cmd.workingDir = wd
		}

		// check if a CommandsFile for the project exists and collect the commands
		c, err = parseCommandsFile(commandsFilePath, false)
		if err != nil {
			Log.Error("failed to parse commandsFile: ", err, "\n")
			os.Exit(1)
		}
	}
}

func (c *CommandsFile) handleInclusion() {
	// handle commandsFile inclusion
	if c.Includes != "" {

		// get current working directory
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		// change into extension directory
		p := strings.TrimSuffix(filepath.Dir(c.Includes), "zeus")
		err = os.Chdir(p)
		if err != nil {
			log.Fatal(err)
		}

		// check if a CommandsFile for the project exists
		// parse and flush it, in order to use it as base for the actual commandsFile of this project
		_, err = parseCommandsFile(commandsFilePath, true)
		if err != nil {
			Log.Error("failed to parse commandsFile: ", err, "\n")
			os.Exit(1)
		}

		// move back into actual root directory
		err = os.Chdir(wd)
		if err != nil {
			log.Fatal(err)
		}

		// check if a CommandsFile for the project exists and collect the commands
		c, err = parseCommandsFile(commandsFilePath, false)
		if err != nil {
			Log.Error("failed to parse commandsFile: ", err, "\n")
			os.Exit(1)
		}
	}
}