/*
 *  ZEUS - An Electrifying Build System
 *  Copyright (c) 2017 Philipp Mieden <dreadl0ck@protonmail.ch>
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
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
)

// contants for builtin names
const (
	exitCommand       = "exit"
	helpCommand       = "help"
	clearCommand      = "clear"
	keysCommand       = "keys"
	versionCommand    = "version"
	infoCommand       = "info"
	configCommand     = "config"
	formatCommand     = "format"
	colorsCommand     = "colors"
	builtinsCommand   = "builtins"
	aliasCommand      = "alias"
	globalsCommand    = "globals"
	deadlineCommand   = "deadline"
	milestonesCommand = "milestones"
	eventsCommand     = "events"
	dataCommand       = "data"
	makefileCommand   = "makefile"
	authorCommand     = "author"
	wikiCommand       = "wiki"
	webCommand        = "web"
)

var builtins = map[string]string{
	exitCommand:       "leave the interactive shell",
	helpCommand:       "print the command overview or the manualtext for a specific command",
	clearCommand:      "clear the terminal screen",
	infoCommand:       "print project info (lines of code + latest git commits)",
	formatCommand:     "run the formatter for all scripts",
	globalsCommand:    "print the current globals",
	configCommand:     "print or change the current config",
	deadlineCommand:   "print or change the deadline",
	milestonesCommand: "print, add or remove the milestones",
	versionCommand:    "print version",
	eventsCommand:     "print, add or remove events",
	dataCommand:       "print the current project data",
	aliasCommand:      "print, add or remove aliases",
	colorsCommand:     "change the current ANSI color profile",
	makefileCommand:   "show or migrate GNU Makefiles",
	authorCommand:     "print or change project author name",
	keysCommand:       "manage keybindings",
	builtinsCommand:   "print the builtins overview",
	webCommand:        "web",
	wikiCommand:       "wiki",
}

// executed when running the info command
// runs a count line of code and displays git info
func printProjectInfo() {

	cmd := exec.Command("cloc", "--exclude-dir=vendor,dist,node_modules,master,files", ".")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		Log.WithError(err).Info("running cloc failed.")
		return
	}

	cmd = exec.Command("git", "log", "-n", "5")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		Log.WithError(err).Info("running git log failed.")
		return
	}
}

// print built-in commands
func printBuiltins() {

	width := 15
	l.Println(cp.colorText + "builtins:")
	for builtin, description := range builtins {
		l.Println(cp.colorCommandName + pad(builtin, width) + cp.colorText + " (" + description + ")")
	}
	l.Println("")
}

// print all available commands
func printCommands() {

	var (
		sortedCommandKeys = make([]string, len(commands))
		index             = 0
	)

	// copy command names into array for sorting
	for key := range commands {
		sortedCommandKeys[index] = key
		index++
	}

	// sort alphabetically
	sort.Strings(sortedCommandKeys)

	// print them
	l.Println(cp.colorText + "commands:")
	for _, key := range sortedCommandKeys {

		cmd := commands[key]

		// check if there are arguments
		if len(cmd.args) != 0 {
			l.Println(cp.colorText + "├~» " + cp.colorCommandName + pad(cmd.name+" "+getArgumentString(cmd.args), 20) + cp.colorText)
		} else {
			l.Println(cp.colorText + "├~» " + cp.colorCommandName + pad(cmd.name, 20) + cp.colorText)
		}

		// print command chain if there is one
		if len(cmd.commandChain) > 0 {
			l.Println(cp.colorText + "├──── " + pad("chain:", 18) + cp.colorCommandChain + formatcommandChain(cmd.commandChain) + cp.colorText)
		}

		// print help section
		l.Println(cp.colorText + "├──── " + pad("help:", 18) + cmd.help)
	}
	l.Println("")
}

// format argStr
func getArgumentString(args []*commandArg) (argStr string) {
	for _, arg := range args {
		argStr += "[" + arg.name + ":" + arg.argType.String() + "] "
	}
	return
}

// print the current configuration as JSON to stdout
func printConfiguration() {

	configMutex.Lock()

	b, err := json.MarshalIndent(conf, "", "	")
	if err != nil {
		Log.WithError(err).Fatal("failed to marshal config to JSON")
	}

	configMutex.Unlock()
	l.Println(string(b))
}

// print the current project data as JSON to stdout
func printProjectData() {

	eventLock.Lock()
	defer eventLock.Unlock()

	// make it pretty
	b, err := json.MarshalIndent(projectData, "", "	")
	if err != nil {
		Log.WithError(err).Fatal("failed to marshal zeus project data to JSON")
	}

	l.Println(string(b))
}

// print the contents of globals.sh on stdout
func listGlobals() {

	if len(globalsContent) > 0 {
		c, err := ioutil.ReadFile("zeus/globals.sh")
		if err != nil {
			l.Fatal("failed to read globals: ", err)
		}
		l.Println(string(c))
	} else {
		l.Println("no globals defined.")
	}
}
