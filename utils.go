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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v2"

	gosxnotifier "github.com/deckarep/gosx-notifier"
	"github.com/fsnotify/fsnotify"
)

var (
	// prompt for the interactive shell
	zeusPrompt  = "zeus"
	signalMutex = &sync.Mutex{}
)

// dump the currently executed script in case of an error
func dumpScript(script string, e error) {

	var (
		t            = "# Timestamp: " + time.Now().Format(timestampFormat) + "\n"
		errString    = "# Error: " + e.Error() + "\n\n"
		dumpFileName = "zeus_error_dump.sh"
	)

	f, err := os.OpenFile(dumpFileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	if err != nil {
		Log.WithError(err).Error("failed to open dump file")
		return
	}
	defer f.Close()

	f.WriteString("#!/bin/bash\n#\n")
	f.WriteString("# ZEUS Error Dump\n")
	f.WriteString(t)
	f.WriteString(errString)
	f.WriteString(script)
	Log.Debug("script dumped: ", dumpFileName)
}

// print the current script to stdout
// adds line numbers
func printScript(script, name string) {

	fmt.Println(" |---------------------------------------------------------------------------------------------|")
	fmt.Println("     Script: " + name)
	fmt.Println(" |---------------------------------------------------------------------------------------------|")
	for i, s := range strings.Split(script, "\n") {

		var lineNumber string
		switch true {
		case i > 9:
			lineNumber = strconv.Itoa(i) + " "
		case i > 99:
			lineNumber = strconv.Itoa(i)
		default:
			lineNumber = strconv.Itoa(i) + "  "
		}
		fmt.Println(" "+lineNumber, s)
	}
	fmt.Println(" |---------------------------------------------------------------------------------------------|")
}

// handle OS SIGNALS for a clean exit and clean up all spawned processes
func handleSignals() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGHUP, syscall.SIGQUIT)

	// var signalLock sync.Mutex
	go func() {

		sig := <-c

		Log.Debug("received SIGNAL: ", sig)

		// lock the mutex
		signalMutex.Lock()

		// pass signal to all spawned procs
		passSignalToProcs(sig)

		// return to interactive shell
		return
	}()
}

// pad the input string up to the given number of space characters
func pad(in string, length int) string {
	if len(in) < length {
		return fmt.Sprintf("%-"+strconv.Itoa(length)+"s", in)
	}
	return in
}

// create a readable string from a commandChain
// example: clean -> build name=testBuild -> install
func formatcommandChain(commands commandChain) (out string) {

	for i, cmd := range commands {

		out += cmd.name

		// check if command has params set
		if len(cmd.params) > 0 {
			for _, p := range cmd.params {
				out += " " + p
			}
		}

		// if not last elem
		if !(i == len(commands)-1) {
			out += " -> "
		}
	}
	return
}

// ClearScreen prints ANSI escape to flush screen
func clearScreen() {
	print("\033[H\033[2J")
}

// count total length of a commandchain
func countCommandChain(chain commandChain) int {
	count := 0
	for _, cmd := range chain {
		count++
		if len(cmd.commandChain) > 0 {
			count += countCommandChain(cmd.commandChain)
		}
	}
	return count
}

func getTotalCommandCount(c *command) int {
	return 1 + countCommandChain(c.commandChain)
}

// print the prompt for the interactive shell
func printPrompt() string {
	colorProfileMutex.Lock()
	defer colorProfileMutex.Unlock()
	return cp.prompt + zeusPrompt + " » " + cp.text
}

// pass the command to the underlying shell
// arguments that contain string literals " or ' will be grouped before passing them to shell
func passCommandToShell(commandName string, args []string) error {

	// handle string literals
	args = fixArgs(args)

	var cmd *exec.Cmd

	// if there are arguments pass them
	if len(args) > 0 {
		cmd = exec.Command(commandName, args...)
	} else {
		cmd = exec.Command(commandName)
	}

	// setup environment
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// fix args in case there is a string literal in there
// this will cause all strings in arguments to be passed as one to the shell
// example:
// ["git", "commit", "-m", "'what", "the", "hell'"] -> ["git", "commit", "-m", "'what the hell'"]
func fixArgs(args []string) []string {

	var (
		fixed         = []string{}
		insideLiteral bool
		literalIndex  int
	)

	// range arguments until there appears a starting string literal
	// from there on concatenate all following fields to the current one
	// when the closing tag appears concatenation is stopped
	for _, a := range args {

		if insideLiteral {
			fixed[literalIndex] += " " + a
		} else {
			fixed = append(fixed, a)
		}

		if isStartTag(a) {
			insideLiteral = true
			literalIndex = len(fixed) - 1
		} else if isEndTag(a) {
			insideLiteral = false
		}
	}

	return fixed
}

// check if the string literal starts
func isStartTag(s string) bool {
	if strings.HasPrefix(s, "\"") || strings.HasPrefix(s, "'") {
		return true
	}
	return false
}

// check if the string literal ends
func isEndTag(s string) bool {
	if strings.HasSuffix(s, "\"") || strings.HasSuffix(s, "'") {
		return true
	}
	return false
}

// check if its a valid command chain
func validCommandChain(args []string, silent bool) bool {

	var (
		chain = strings.Join(args, " ")
		job   = p.AddJob(chain, silent)
	)

	commandList, err := parseCommandChain(chain)
	if err != nil {
		Log.WithError(err).Error("failed to parse command chain")
		return false
	}

	defer p.RemoveJob(job)

	_, err = job.getCommandChain(commandList, nil)
	if err != nil {
		if !silent {
			Log.WithError(err).Error("failed to get command chain")
		}
		return false
	}

	return true
}

// handle help shell command
func handleHelpCommand(args []string) {

	if len(args) < 2 {
		printHelpUsageErr()
		return
	}

	if c, ok := commands[args[1]]; ok {
		l.Println("\n" + c.manual)
		return
	}

	printHelpUsageErr()
}

func printHelpUsageErr() {
	l.Println(ErrInvalidUsage)
	l.Println("usage: help <command>")
}

// check if the argument type matches the expected one
func validArgType(in string, k reflect.Kind) bool {

	var err error

	switch k {
	case reflect.Bool:
		_, err = strconv.ParseBool(in)
	case reflect.Float64:
		_, err = strconv.ParseFloat(in, 64)
	case reflect.String:
	case reflect.Int:
		_, err = strconv.ParseInt(in, 64, 0)
	default:
		return false
	}

	if err == nil {
		return true
	}
	return false
}

func showNote(text, subtitle string) {

	note := gosxnotifier.NewNotification(text)
	note.Title = "ZEUS"
	note.Subtitle = subtitle

	// optionally, set a group which ensures only one notification is ever shown replacing previous notification of same group id
	note.Group = "com.zeus"

	// optionally, set a sender icon
	note.Sender = "com.apple.Terminal"

	// optionally, specify a url or bundleid to open should the notification be clicked
	note.Link = "http://" + hostName + ":" + strconv.Itoa(conf.PortWebPanel)

	// optionally, an app icon
	// note.AppIcon = "gopher.png"

	// optionally, a content image
	// note.ContentImage = "gopher.png"

	err := note.Push()
	if err != nil {
		Log.WithError(err).Error("error pushing notification")
	}
}

// pass the args to the OSX open command
func open(args ...string) {
	err := exec.Command("open", args...).Run()
	if err != nil {
		Log.WithError(err).Error("failed to open: ", args)
	}
}

// generate a 8byte random string
func randomString() string {

	var rb = make([]byte, 8)

	_, err := rand.Read(rb)
	if err != nil {
		Log.WithError(err).Fatal(ErrReadingRandomString)
	}

	return hex.EncodeToString(rb)
}

func watchScripts(eventID string) {

	// dont add a new watcher when the event exists
	projectDataMutex.Lock()
	for _, e := range projectData.Events {
		if e.Name == "script watcher" {
			projectDataMutex.Unlock()
			return
		}
	}
	projectDataMutex.Unlock()

	err := addEvent(newEvent(zeusDir, fsnotify.Write, "script watcher", f.fileExtension, eventID, "internal", func(e fsnotify.Event) {

		Log.Debug("change event: ", e.Name)

		if e.Name == zeusDir+"/globals.sh" {

			c, err := ioutil.ReadFile(zeusDir + "/globals.sh")
			if err != nil {
				Log.WithError(err).Error("failed to read globals")
				return
			}
			Log.Info("updated globals")
			globalsContent = c
		} else {
			err := addCommand(e.Name, true)
			if err != nil {
				Log.WithError(err).Error("failed to parse command: ", e.Name)
			}
		}

	}))
	if err != nil {
		Log.WithError(err).Error("failed to watch script headers")
	}
}

// dump datastructure as YAML - useful for debugging
func dumpYAML(i interface{}) {
	out, err := yaml.Marshal(i)
	if err != nil {
		log.Println("ERROR: failed to marshal to YAML:", err)
		return
	}

	fmt.Println(string(out))
}

// print file content with linenumbers to stdout - useful for debugging
func printFileContents(data []byte) {
	l.Println("| ------------------------------------------------------------ |")
	for i, line := range strings.Split(string(data), "\n") {
		l.Println(pad(strconv.Itoa(i+1), 3), line)
	}
	l.Println("| ------------------------------------------------------------ |")
}

func printCompletions(previous string) {

	switch previous {
	case bootstrapCommand:
		fmt.Println("file dir")
		return
	case makefileCommand:
		fmt.Println("migrate")
		return
	}

	// print builtins
	var completions = []string{
		helpCommand,
		formatCommand,
		dataCommand,
		aliasCommand,
		configCommand,
		versionCommand,
		updateCommand,
		infoCommand,
		colorsCommand,
		authorCommand,
		builtinsCommand,
		makefileCommand,
		gitFilterCommand,
		createCommand,
	}

	for _, name := range completions {
		if previous == name || previous == bootstrapCommand {
			return
		}
	}

	// bootstrap is available when there's no dir or zeusfile
	fmt.Print("bootstrap ")

	// check for zeusfile
	var (
		zeusfile = new(Zeusfile)
		contents []byte
		err      error
	)

	contents, err = ioutil.ReadFile("Zeusfile.yml")
	if err != nil {
		contents, err = ioutil.ReadFile("Zeusfile")
	}
	if err == nil {

		// replace tabs with spaces
		contents = bytes.Replace(contents, []byte("\t"), []byte("    "), -1)

		// unmarshal data
		err = yaml.Unmarshal(contents, zeusfile)
		if err != nil {
			fmt.Println()
			return
		}

		for name := range zeusfile.Commands {
			if name == previous {
				return
			}
			completions = append(completions, name)
		}
	} else {

		// check for zeusDir
		files, err := ioutil.ReadDir(zeusDir)
		if err != nil {
			fmt.Println()
			return
		}

		for _, stat := range files {
			if strings.HasSuffix(stat.Name(), ".sh") {
				fileName := strings.TrimSuffix(filepath.Base(stat.Name()), ".sh")
				if fileName != "globals" {
					if fileName == previous {
						return
					}
					completions = append(completions, fileName)
				}
			}
		}
	}

	// print result
	for _, name := range completions {
		fmt.Print(name + " ")
	}
	fmt.Println()
}
