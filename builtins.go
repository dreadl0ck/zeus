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

// used to indicate that no output shall be printed into the terminal
var shellBusy bool

// constants for builtin names
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
	gitFilterCommand  = "git-filter"
	todoCommand       = "todo"
	updateCommand     = "update"
	procsCommand      = "procs"
	editCommand       = "edit"
	generateCommand   = "generate"
)

// mapped builtin names to description
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
	gitFilterCommand:  "filter git log output",
	todoCommand:       "manage todos",
	updateCommand:     "update zeus version",
	procsCommand:      "manage spawned processes",
	editCommand:       "edit scripts",
	generateCommand:   "generate a standalone version of the script",
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

	l.Println()
	l.Println(cp.Text + "builtins")

	// print
	for _, name := range names {
		description := builtins[name]
		l.Println(cp.CmdName + pad(name, width) + cp.Text + description)
	}
	l.Println()
}

// print all available commands
func printCommands() {

	cmdMap.Lock()
	defer cmdMap.Unlock()

	if len(cmdMap.items) == 0 {
		return
	}

	var sortedCommandKeys []string

	// copy command names into array for sorting
	for key, cmd := range cmdMap.items {
		// do not display hidden commands
		if cmd.hidden {
			continue
		}
		sortedCommandKeys = append(sortedCommandKeys, key)
	}

	// sort alphabetically
	sort.Strings(sortedCommandKeys)

	// print them
	l.Println(cp.Text + "commands")
	printSortedCommandKeys(sortedCommandKeys)
	l.Println("")
}

func printSortedCommandKeys(sortedCommandKeys []string) {

	maxLen := 14

	for i, key := range sortedCommandKeys {

		var (
			lastElem = i == len(sortedCommandKeys)-1
			cmd      = cmdMap.items[key]
		)

		if conf.fields.Quiet {
			var deps string
			if len(cmd.dependencies) > 0 {
				deps = cp.CmdFields + " [" + formatDependencies(cmd.dependencies) + "]"
			}
			if lastElem {
				l.Print(cp.Text + "└─── " + cp.CmdName + cmd.name + " " + getArgumentString(cmd.args) + deps)
			} else {
				l.Print(cp.Text + "├─── " + cp.CmdName + cmd.name + " " + getArgumentString(cmd.args) + deps)
			}

		} else {

			if lastElem {
				l.Print(cp.Text + "└─── " + cp.CmdName + cmd.name + " " + getArgumentString(cmd.args) + cp.Text)
			} else {
				l.Print(cp.Text + "├─── " + cp.CmdName + cmd.name + " " + getArgumentString(cmd.args) + cp.Text)
			}

			if cmd.path != "" {
				printLine(pad("path", maxLen)+cp.CmdFields+cmd.path, lastElem, !(len(cmd.dependencies) > 0) && !(len(cmd.outputs) > 0) && !cmd.async && !cmd.buildNumber && !(len(cmd.description) > 0))
			}

			if len(cmd.dependencies) > 0 {
				printLine(pad("dependencies", maxLen)+cp.CmdFields+formatDependencies(cmd.dependencies), lastElem, !(len(cmd.outputs) > 0) && !cmd.async && !cmd.buildNumber && !(len(cmd.description) > 0))
			}

			if len(cmd.outputs) > 0 {
				printLine(pad("outputs", maxLen)+cp.CmdFields+strings.Join(cmd.outputs, ", "), lastElem, !cmd.async && !cmd.buildNumber && !(len(cmd.description) > 0))
			}

			if cmd.async {
				printLine(cp.CmdFields+"async", lastElem, !cmd.buildNumber && !(len(cmd.description) > 0))
			}

			if cmd.buildNumber {
				printLine(cp.CmdFields+"buildNumber", lastElem, !(len(cmd.description) > 0))
			}

			if len(cmd.description) > 0 {
				printLine(pad("description", maxLen)+cp.CmdFields+cmd.description, lastElem, true)
			}

			if !lastElem {
				l.Println("|")
			}
		}

	}
}

func printLine(line string, lastElem, lastItem bool) {
	switch {
	case lastElem && lastItem:
		l.Println(cp.Text + "     └─── " + line + cp.Text)
	case lastItem:
		l.Println(cp.Text + "|    └─── " + line + cp.Text)
	case lastElem:
		l.Println(cp.Text + "     ├─── " + line + cp.Text)
	default:
		l.Println(cp.Text + "|    ├─── " + line + cp.Text)
	}
}

// format commandArg map into human readable string
func getArgumentString(args map[string]*commandArg) string {

	if len(args) == 0 {
		return ""
	}

	var (
		requiredArgs string
		optionalArgs string
		count        = 1
	)

	for _, arg := range args {
		var t = cp.CmdArgType + strings.Title(arg.argType.String())
		if arg.optional {
			if arg.defaultValue != "" {
				t += "?" + cp.CmdOutput + " =" + arg.defaultValue
			} else {
				t += "?"
			}
		}
		if arg.optional {
			optionalArgs += cp.CmdArgs + arg.name + cp.Text + ":" + t + cp.Text + ", "
		} else {
			requiredArgs += cp.CmdArgs + arg.name + cp.Text + ":" + t + cp.Text + ", "
		}
		count++
	}

	if optionalArgs == "" {
		return cp.Text + "(" + strings.TrimSuffix(requiredArgs, ", ") + cp.Text + ")"
	}
	return cp.Text + "(" + requiredArgs + strings.TrimSuffix(optionalArgs, ", ") + cp.Text + ")"
}

func printGitFilterCommandUsageErr() {
	l.Println("invalid usage")
	l.Println("usage: git-filter [keyword]")
}

func handleGitFilterCommand(args []string) {

	l.Println()

	w := 35
	l.Println(cp.Prompt + pad("time", w) + pad("author", 41) + "subject")
	out, err := exec.Command("git", "log", "--pretty=format:"+cp.Text+pad("[%ci]", 13)+pad("%cn", w)+"%s").CombinedOutput()
	if err != nil {
		l.Println(err)
		return
	}

	if len(args) < 2 {
		// print all
		l.Println(string(out))
		return
	}

	if len(args) > 2 {
		printGitFilterCommandUsageErr()
		return
	}

	// filter output for lines containing the keyword
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

// print todo overview
func printTodos() {

	conf.Lock()
	defer conf.Unlock()

	contents, err := ioutil.ReadFile(conf.fields.TodoFilePath)
	if err != nil {
		if conf.fields.Debug {
			l.Println(err)
		}
		return
	}

	var index int

	for _, line := range strings.Split(string(contents), "\n") {
		if strings.HasPrefix(line, "#") {
			l.Println("\n" + cp.Prompt + line + "\n")
		}
		if strings.HasPrefix(line, "- ") {
			index++
			l.Println(cp.CmdOutput + pad(strconv.Itoa(index)+")", 4) + strings.TrimPrefix(line, "- "))
		}
	}
}

// print amount of tasks in todo file
func printTodoCount() {

	conf.Lock()
	defer conf.Unlock()

	if len(conf.fields.TodoFilePath) == 0 {
		return
	}

	contents, err := ioutil.ReadFile(conf.fields.TodoFilePath)
	if err != nil {
		if conf.fields.Debug {
			l.Println(err)
		}
		return
	}

	var count int

	for _, line := range strings.Split(string(contents), "\n") {
		if strings.HasPrefix(line, "- ") {
			count++
		}
	}

	l.Println(cp.Text + pad("TODOs", 14) + cp.Prompt + strconv.Itoa(count))
}

// manage todos
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

		f, err := os.OpenFile(conf.fields.TodoFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
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

		contents, err := ioutil.ReadFile(conf.fields.TodoFilePath)
		if err != nil {
			l.Println(err)
			return
		}

		f, err := os.OpenFile(conf.fields.TodoFilePath, os.O_RDWR|os.O_TRUNC, 0600)
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

// run go get -u to get the latest ZEUS build from github
func updateZeus() {

	cmd := exec.Command("go", "get", "-u", "-v", "github.com/dreadl0ck/zeus")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	l.Println("current version:", version)
	l.Println("updating zeus...")
	err := cmd.Run()
	if err != nil {
		Log.WithError(err).Fatal("failed to update zeus")
	}

	l.Println("zeus updated!")
	cleanup(nil)
	os.Exit(0)
}

func printEditCommandUsageErr() {
	l.Println("invalid usage")
	l.Println("usage: edit <command>")
}

// start editor to edit files in the interactive shell
func handleEditCommand(args []string) {

	if len(args) < 2 {
		printEditCommandUsageErr()
		return
	}

	var path string

	// get editor from config
	conf.Lock()
	editor := conf.fields.Editor
	conf.Unlock()

	cmdMap.Lock()
	defer cmdMap.Unlock()

	switch args[1] {
	case "config":
		path = zeusDir + "/config.yml"
	case "commands":
		path = commandsFilePath
	case "todo":
		conf.Lock()
		path = conf.fields.TodoFilePath
		conf.Unlock()
	case "data":
		path = zeusDir + "/data.yml"
	case "globals":
		if len(args) > 2 {

			lang, err := ls.getLang(args[1])
			if err != nil {
				l.Println(err)
				return
			}

			path = zeusDir + "/globals/globals" + lang.FileExtension
		} else {
			path = commandsFilePath
		}
	default:
		// check if its a valid command
		if cmd, ok := cmdMap.items[args[1]]; ok {

			// command has a path set?
			if cmd.path == "" {
				path = commandsFilePath
			} else {
				path = cmd.path

				if _, err := os.Stat(cmd.path); err != nil {
					path = commandsFilePath
				}
			}
		} else {
			l.Println("invalid command:", args[1])
			return
		}
	}

	// command is in commandsFile? let's start the editor at the correct position!
	if path == commandsFilePath {

		// check if editor supports starting at a position
		// it could be supplied as a path, so we just check the path suffix
		switch {
		case strings.HasSuffix(editor, "vim"):
			// find position of command in commands file
			line, col, err := getYAMLFieldPosition(args[1])
			if err != nil {
				l.Println(err)
				return
			}
			args = append(args, "+call cursor("+strconv.Itoa(line+1)+","+strconv.Itoa(col)+")")
		case strings.HasSuffix(editor, "micro"):
			// find position of command in commands file
			line, col, err := getYAMLFieldPosition(args[1])
			if err != nil {
				l.Println(err)
				return
			}
			args = append(args, "+"+strconv.Itoa(line+1)+":"+strconv.Itoa(col))
		default: // not supported
		}
	}

	var editorArgs []string

	// prepend args
	if len(args) > 2 {
		if args[2] == "line" {
			if !(len(args) > 3) {
				printEditCommandUsageErr()
				return
			}
			editorArgs = append(editorArgs, "+"+args[3]+":0")
		} else {
			editorArgs = append(editorArgs, args[2:]...)
		}
	}

	// append path
	editorArgs = append(editorArgs, path)

	Log.Debug(editor, " ", editorArgs)

	cmd := exec.Command(editor, editorArgs...)
	wireEnv(cmd)

	shellBusy = true

	err := cmd.Run()
	if err != nil {
		//Log.WithError(err).Error("edit command failed: using vim as fallback")

		// try vim as fallback
		cmd = exec.Command("vim", path)
		wireEnv(cmd)

		err = cmd.Run()
		if err != nil {
			Log.WithError(err).Error("edit command failed: fix editor in config")
		}
	}

	shellBusy = false
}

func printGenerateCommandUsageErr() {
	l.Println(ErrInvalidUsage)
	l.Println("usage: generate <outputName> <commandChain>")
}
