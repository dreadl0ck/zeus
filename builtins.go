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
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
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
	createCommand     = "create"
	bootstrapCommand  = "bootstrap"
	zeusfileCommand   = "migrate-zeusfile"
	gitFilterCommand  = "git-filter"
	todoCommand       = "todo"
	updateCommand     = "update"
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
	webCommand:        "start web interface",
	wikiCommand:       "start web wiki ",
	createCommand:     "bootstrap single commands",
	zeusfileCommand:   "migrate zeusfile into a zeus directory",
	gitFilterCommand:  "filter git log output",
	todoCommand:       "manage todos",
	updateCommand:     "update zeus version",
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

	var (
		names []string
		width = 17
	)

	// assemble array for sorting
	for builtin := range builtins {
		names = append(names, builtin)
	}

	// sort alphabetically
	sort.Strings(names)

	l.Println(cp.colorText + "builtins:")

	// print
	for _, name := range names {
		description := builtins[name]
		l.Println(cp.colorCommandName + pad(name, width) + cp.colorText + description)
	}
	l.Println("")
}

// print all available commands
func printCommands() {

	commandMutex.Lock()
	defer commandMutex.Unlock()

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
			l.Print(cp.colorText + "├~» " + cp.colorCommandName + pad(cmd.name+" "+getArgumentString(cmd.args), 20) + cp.colorText)
		} else {
			l.Print(cp.colorText + "├~» " + cp.colorCommandName + pad(cmd.name, 20) + cp.colorText)
		}

		if len(cmd.commandChain) > 0 {
			l.Println(cp.colorText + "├──── " + pad("chain:", 18) + cp.colorCommandChain + formatcommandChain(cmd.commandChain))
		}

		// print help section
		l.Println(cp.colorText + "├──── " + pad("help:", 18) + cmd.help)
	}
	l.Println("")
}

// format argStr
func getArgumentString(args map[string]*commandArg) string {

	var (
		argStr = "("
		count  = 1
	)

	for _, arg := range args {
		var t = strings.Title(arg.argType.String())
		if arg.optional {
			if arg.defaultValue != "" {
				t += "? =" + arg.defaultValue
			} else {
				t += "?"
			}
		}
		if count == len(args) {
			argStr += arg.name + ":" + t + ")"
		} else {
			argStr += arg.name + ":" + t + ", "
		}
		count++
	}
	return argStr
}

// print the contents of globals.sh on stdout
func listGlobals() {

	if len(globalsContent) > 0 {
		c, err := ioutil.ReadFile(zeusDir + "/globals.sh")
		if err != nil {
			l.Fatal("failed to read globals: ", err)
		}
		l.Println(string(c))
	} else {
		l.Println("no globals defined.")
	}
}

func printGitFilterCommandUsageErr() {
	l.Println("invalid usage")
	l.Println("usage: git-filter [keyword]")
}

func handleGitFilterCommand(args []string) {

	out, err := exec.Command("git", "log", "--pretty=format:[%cd] author: %cn, subject: %s").CombinedOutput()
	if err != nil {
		l.Println(err)
		return
	}

	if len(args) < 2 {
		// print all
		l.Println(string(out))
		return
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, args[1]) {
			l.Println(line)
		}
	}
}

func printTodoCommandUsageErr() {
	l.Println("invalid usage")
	l.Println("usage: todo [add <task>] [remove <index>]")
}

func printTodos() {
	contents, err := ioutil.ReadFile(conf.TodoFilePath)
	if err != nil {
		l.Println(err)
		return
	}

	var index int

	for _, line := range strings.Split(string(contents), "\n") {
		if strings.HasPrefix(line, "- ") {
			index++
			l.Println(pad(strconv.Itoa(index)+")", 4) + strings.TrimPrefix(line, "- "))
		}
	}
}

func handleTodoCommand(args []string) {

	if len(args) < 2 {
		printTodos()
		return
	}

	if len(args) < 3 {
		printTodoCommandUsageErr()
		return
	}

	switch args[1] {
	case "add":

		l.Println("adding TODO ", args[2:])

		f, err := os.OpenFile(conf.TodoFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
		if err != nil {
			l.Println(err)
			return
		}
		defer f.Close()

		_, err = f.WriteString("\n- " + strings.Join(args[2:], " "))
		if err != nil {
			l.Println(err)
		}
	case "remove":

		l.Println("removing TODO ", args[2])

		i, err := strconv.Atoi(args[2])
		if err != nil {
			l.Println(err)
			return
		}

		contents, err := ioutil.ReadFile(conf.TodoFilePath)
		if err != nil {
			l.Println(err)
			return
		}

		f, err := os.OpenFile(conf.TodoFilePath, os.O_RDWR|os.O_TRUNC, 0600)
		if err != nil {
			l.Println(err)
			return
		}
		defer f.Close()

		var index int
		for _, line := range strings.Split(string(contents), "\n") {
			if strings.HasPrefix(line, "- ") {
				index++
			}
			if index != i {
				f.WriteString(line + "\n")
			}
		}
	}
}

func updateZeus() {
	cmd := exec.Command("go", "get", "-u", "github.com/dreadl0ck/zeus")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	l.Println("current version:", version)
	l.Println("updating zeus...")
	err := cmd.Run()
	if err != nil {
		Log.WithError(err).Fatal("failed to update zeus")
	}
	l.Println("zeus updated!")
}
