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
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/elliotchance/orderedmap"
	"github.com/gen2brain/beeep"
	"github.com/mgutz/ansi"
)

var (
	// prompt for the interactive shell
	zeusPrompt  = "zeus"
	signalMutex = &sync.Mutex{}

	// ErrNoLineNumberFound means there was no line number in the error message
	ErrNoLineNumberFound = errors.New("no line number found in error string")
)

// dump the currently executed script to disk
func dumpScript(script, language string, e error, stdErr string) {

	stat, err := os.Stat(zeusDir + "/dumps")
	if err != nil {
		err := os.Mkdir(zeusDir+"/dumps", 0700)
		if err != nil {
			Log.WithError(err).Error("failed to create dumps directory")
			return
		}
	} else {
		if !stat.IsDir() {
			Log.Error("dumpScript: " + zeusDir + "/dumps is a file")
			return
		}
	}

	lang, err := ls.getLang(language)
	if err != nil {
		Log.WithError(err).Error("failed to get lang")
		return
	}

	var stdErrOutputComment string
	for _, line := range strings.Split(stdErr, "\n") {
		stdErrOutputComment += lang.Comment + " " + line + "\n"
	}

	var (
		t            = lang.Comment + " Timestamp: " + time.Now().Format(timestampFormat) + "\n"
		errString    = lang.Comment + " Error: " + e.Error() + "\n" + lang.Comment + " StdErr: \n" + stdErrOutputComment + "\n\n"
		dumpFileName = zeusDir + "/dumps/error_dump" + lang.FileExtension
	)

	f, err := os.OpenFile(dumpFileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	if err != nil {
		Log.WithError(err).Error("failed to open dump file")
		return
	}
	defer f.Close()

	f.WriteString(lang.Bang + "\n" + lang.Comment + "\n")
	f.WriteString(lang.Comment + " ZEUS Error Dump\n")
	f.WriteString(t)
	f.WriteString(errString)
	f.WriteString(script)
	Log.Debug("script dumped: ", dumpFileName)
}

// print the complete script to stdout
// adds line numbers and optionally highlight a line
// when no line shall be highlighted pass -1
func printScript(contents, path string, highlightLine int) {

	fmt.Println("\n" + cp.Reset + " |---------------------------------------------------------------------------------------------|")
	fmt.Println("     Script: " + path)
	fmt.Println(" |---------------------------------------------------------------------------------------------|")
	for i, s := range strings.Split(contents, "\n") {

		var lineNumber string
		switch true {
		case i > 9:
			lineNumber = strconv.Itoa(i) + " "
		case i > 99:
			lineNumber = strconv.Itoa(i)
		default:
			lineNumber = strconv.Itoa(i) + "  "
		}

		if i == highlightLine {
			fmt.Println(" "+ansi.Red+lineNumber, s+cp.Reset)
		} else {
			fmt.Println(" "+lineNumber, s)
		}
	}
	fmt.Println(" |---------------------------------------------------------------------------------------------|" + cp.Text)
}

// print a code snippet to stdout
// adds line numbers and optionally highlight a line
// when no line shall be highlighted pass -1
// when a value for highlight line is supplied,
// $scope lines before and after the line will be printed
func printCodeSnippet(contents, path string, highlightLine int) {

	var (
		rangeStart int
		rangeEnd   int
	)

	conf.Lock()
	scope := conf.fields.CodeSnippetScope
	conf.Unlock()

	if highlightLine > 0 {
		rangeStart = highlightLine - scope
		rangeEnd = highlightLine + scope
	}

	fmt.Println("\n" + cp.Reset + " |---------------------------------------------------------------------------------------------|")
	fmt.Println("     File: " + path)
	fmt.Println(" |---------------------------------------------------------------------------------------------|")
	for i, s := range strings.Split(contents, "\n") {

		line := i + 1

		if line < rangeStart || line > rangeEnd {
			continue
		}

		var lineNumber string
		switch true {
		case line > 9:
			lineNumber = strconv.Itoa(line) + " "
		case line > 99:
			lineNumber = strconv.Itoa(line)
		default:
			lineNumber = strconv.Itoa(line) + "  "
		}

		if line == highlightLine {
			fmt.Println(" "+ansi.Red+lineNumber, s+cp.Reset)
		} else {
			fmt.Println(" "+lineNumber, s)
		}
	}
	fmt.Println(" |---------------------------------------------------------------------------------------------|" + cp.Text)
}

// handle OS SIGNALS for a clean exit and clean up all spawned processes
func handleSignals(cmdFile *CommandsFile) {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGHUP, syscall.SIGQUIT)

	// var signalLock sync.Mutex
	go func() {

		sig := <-c

		// lock the mutex
		signalMutex.Lock()
		defer signalMutex.Unlock()

		// pass signal to all spawned processes
		passSignalToProcs(sig)

		if shellBusy {
			// user is currently running a command, reattach the signal handler after handling it.
			go handleSignals(cmdFile)

			// and stay in the zeus shell
			return
		} else {
			// use is in the zeus shell - run cleanup and exit.
			fmt.Println(sig)
			cleanup(cmdFile)
			os.Exit(0)
		}
	}()
}

// pad the input string up to the given number of space characters
func pad(in string, length int) string {
	if len(in) < length {
		return fmt.Sprintf("%-"+strconv.Itoa(length)+"s", in)
	}
	return in
}

// create a readable string from a dependency array
// example: clean -> build name=testBuild -> install
func formatDependencies(deps []string) (out string) {

	for i, cmd := range deps {

		out += cmd

		// if not last elem
		if !(i == len(deps)-1) {
			out += " -> "
		}
	}
	return
}

// ClearScreen prints ANSI escape to flush screen
func clearScreen() {
	print("\033[H\033[2J")
}

// extract a line number from an stdErr buffer
// returns the first occurence of a line containing the 'line' keyword followed by a number
// used to highlight a line in a script when an error during execution occurs
func extractLineNumFromError(errMsg, errLineNumberSymbol string) (int, error) {

	// split buffer by newlines
	for _, line := range strings.Split(errMsg, "\n") {

		// check line contains 'line' keyword
		if strings.Contains(line, errLineNumberSymbol) {

			numExp := regexp.MustCompile("[0-9]+")

			// split string by keyword
			// the line number will appear after the keyword
			slice := strings.Split(line, errLineNumberSymbol)
			if len(slice) > 1 {

				// look for first number match after 'line' keyword
				res := numExp.FindAllString(slice[1], 1)
				if len(res) > 0 {

					// convert result string to integer
					i, err := strconv.Atoi(res[0])
					if err != nil {
						return 0, err
					}
					return i, nil
				}
			}
		}
	}
	return 0, ErrNoLineNumberFound
}

// count total length of the commands dependencies
func countDependencies(deps []string) (int, error) {

	if len(deps) == 0 {
		return 0, nil
	}

	count := 0
	for _, dep := range deps {

		fields := strings.Fields(dep)
		if len(fields) == 0 {
			return 0, ErrEmptyDependency
		}

		// lookup
		cmd, err := cmdMap.getCommand(fields[0])
		if err != nil {
			return 0, errors.New("invalid dependency: " + err.Error())
		}

		err = s.incrementRecursionCount(cmd.name)
		if err != nil {
			return 0, err
		}

		count++
		if len(cmd.dependencies) > 0 {
			c, err := countDependencies(cmd.dependencies)
			if err != nil {
				return 0, err
			}
			count += c
		}
	}
	return count, nil
}

func getTotalDependencyCount(c *command) (int, error) {
	count, err := countDependencies(c.dependencies)
	return count + 1, err
}

// print the prompt for the interactive shell
func printPrompt() string {
	return cp.Prompt + zeusPrompt + " » " + cp.Text
}

// pass the command to the bash
func passCommandToShell(commandName string, args []string) error {

	cmd := exec.Command("/bin/bash", "-e", "-c", commandName+" "+strings.Join(args, " "))

	// setup environment
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

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

// handle help shell command
func handleHelpCommand(args []string) {

	if len(args) < 2 {
		printHelpUsageErr()
		return
	}

	if c, ok := cmdMap.items[args[1]]; ok {

		if c.help != "" {
			l.Println("\n" + c.help)
		} else {
			l.Println("no help text available.")
		}
		return
	}

	l.Println("unknown command:", args[1])
}

func printHelpUsageErr() {
	l.Println(ErrInvalidUsage)
	l.Println("usage: help <command>")
}

// check if the argument type matches the expected one
func validArgType(in string, k reflect.Kind) error {

	var err error

	switch k {
	case reflect.Bool:
		_, err = strconv.ParseBool(in)
	case reflect.Int:
		_, err = strconv.ParseInt(in, 10, 0)
	case reflect.Float64:
		_, err = strconv.ParseFloat(in, 10)
	case reflect.String:

		// check if input is explicitely marked as string
		if strings.HasPrefix(in, "\"") && strings.HasSuffix(in, "\"") {
			return nil
		} else if strings.HasPrefix(in, "'") && strings.HasSuffix(in, "'") {
			return nil
		}

		// you could pass everything as a string
		// so lets check if its not something else...
		_, err = strconv.ParseBool(in)
		if err == nil {
			return errors.New("got bool but want string")
		}
		_, err = strconv.ParseInt(in, 10, 0)
		if err == nil {
			return errors.New("got int but want string")
		}
		_, err = strconv.ParseFloat(in, 10)
		if err == nil {
			return errors.New("got float but want string")
		}

		// all good
		return nil
	default:
		return errors.New("recevied unknown type")
	}

	return err
}

// display an OS notification
func showNote(text, subtitle string) {
	err := beeep.Notify("ZEUS", text+":"+subtitle, "")
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

	// read random bytes
	_, err := rand.Read(rb)
	if err != nil {
		Log.WithError(err).Fatal(ErrReadingRandomString)
	}

	// return hex string
	return hex.EncodeToString(rb)
}

// print file content with linenumbers to stdout - useful for debugging
func printFileContents(data []byte) {
	l.Println("| ------------------------------------------------------------ |")
	for i, line := range strings.Split(string(data), "\n") {
		l.Println(pad(strconv.Itoa(i+1), 3), line)
	}
	l.Println("| ------------------------------------------------------------ |")
}

// print available completions for the bash-completion package
func printCompletions(previous string) {

	switch previous {
	case makefileCommand:
		fmt.Println("migrate")
		return
	}

	// print builtins
	var completions = []string{
		helpCommand,
		// bootstrapCommand,
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
		generateCommand,
		editCommand,
	}

	for _, name := range completions {
		if previous == name || previous == bootstrapCommand {
			return
		}
	}

	// check for commandsFile
	var (
		commandsFile = new(CommandsFile)
		contents     []byte
		err          error
	)

	contents, err = ioutil.ReadFile(commandsFilePath)
	if err == nil {

		// unmarshal data
		err = yaml.Unmarshal(contents, commandsFile)
		if err != nil {
			fmt.Println()
			return
		}

		for name := range commandsFile.Commands {
			if name == previous {
				return
			}
			completions = append(completions, name)
		}
	} else {

		// bootstrap is available when there's no zeusDir or commandsFile
		fmt.Print("bootstrap ")

		// read scripts
		files, err := ioutil.ReadDir(scriptDir)
		if err != nil {
			return
		}

		// filter completions
		for _, stat := range files {
			fileName := strings.TrimSuffix(filepath.Base(stat.Name()), filepath.Ext(stat.Name()))
			if fileName != "globals" {
				if fileName == previous {
					return
				}
				completions = append(completions, fileName)
			}
		}
	}

	// print result
	for _, name := range completions {
		fmt.Print(name + " ")
	}
	fmt.Println()
}

// wire up environment for spawned commands
// connect stdin, stdout, stderr and pass environment
func wireEnv(cmd *exec.Cmd) {
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
}

/*
 * YAML Utils
 */

// extract YAML field name
// returns an empty string if the field is not valid
func extractYAMLField(line string) string {
	return strings.TrimSuffix(strings.TrimSpace(yamlField.FindString(line)), ":")
}

func getYAMLFieldPosition(fieldName string) (line int, col int, err error) {

	b, err := ioutil.ReadFile(commandsFilePath)
	if err != nil {
		return line, col, err
	}

	for index, l := range strings.Split(string(b), "\n") {
		if strings.Contains(l, fieldName+":") {
			line = index
			col = countLeadingSpace(l)
		}
	}

	return
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

// remove duplicates element keeping the left-most ones
func stripArrayRight(array []string) (strip []string) {
	var stripMap = orderedmap.NewOrderedMap()
	for _, element := range array {
		if _, ok := stripMap.Get(element); !ok {
			stripMap.Set(element, nil)
			strip = append(strip, element)
		}
	}
	return
}
