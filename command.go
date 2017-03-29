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
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dreadl0ck/readline"
	"github.com/mgutz/ansi"
)

var (
	// ErrEmptyName means the script has an empty name. thats cant be correct
	ErrEmptyName = errors.New("script has an empty name - wtf")

	// ErrTooManyArguments means there are too many arguments
	ErrTooManyArguments = errors.New("too many arguments")

	// ErrNotEnoughArguments means there are not enough arguments
	ErrNotEnoughArguments = errors.New("missing arguments, command incomplete")

	// ErrInvalidArgumentType means the argument type does not match the expected type
	ErrInvalidArgumentType = errors.New("invalid argument type")

	// ErrInvalidArgumentLabel means the argument label does not match the expected label
	ErrInvalidArgumentLabel = errors.New("invalid argument label")

	// ErrInvalidDependency means the named dependency command does not exist
	ErrInvalidDependency = errors.New("invalid dependency")

	lineErr = regexp.MustCompile("line\\s[1-9]+")
)

type commandChain []*command

// command represents a parsed script in memory
type command struct {

	// the path where the script resides
	path string

	// commandName
	name string

	// arguments that the command needs
	args []*commandArg

	// parameters that can be set, for calling commands with arguments in commandChains
	params []string

	// short help text
	help string

	// manual text
	manual string

	// commandChain contains commands that will be executed before the command runs
	commandChain commandChain

	// completer for interactive shell
	PrefixCompleter *readline.PrefixCompleter

	// buildNumber
	buildNumber bool

	// if the command depends on other command(s)
	// add them here and they will be executed prior to execution of the current command
	// if their named output files to not exist
	dependencies []string

	// output file(s) of the command
	// if the file exists the command will not be executed
	outputs []string
}

// Run executes the command
func (c *command) Run(args []string) error {

	if len(c.outputs) != 0 {
		// check if named outputs exist
		for _, output := range c.outputs {

			Log.Debug("checking output: ", output)

			_, err := os.Stat(output)
			if err == nil {
				// file exists, skip it
				Log.WithFields(logrus.Fields{
					"commandName": c.name,
					"output":      output,
				}).Info("skipping command because its output exists")
				return nil
			}
		}
	}

	// check if there are dependencies for the current command
	if len(c.dependencies) != 0 {

		for _, dep := range c.dependencies {

			numCommands++

			var (
				cmd *command
				ok  bool
				err error
			)

			// handle args
			fields := strings.Fields(dep)
			if len(fields) == 0 {
				Log.Error("empty fields")
				return ErrInvalidDependency
			}

			// look up command name
			if cmd, ok = commands[fields[0]]; !ok {
				return ErrInvalidDependency
			}

			if len(cmd.outputs) != 0 {

				var outputMissing bool

				// check if all named outputs exist
				for _, output := range cmd.outputs {
					_, err := os.Stat(output)
					if err != nil {
						outputMissing = true
					}
				}

				// there is at least one output missing
				// execute the command
				if outputMissing {

					// pass args if there are any
					if len(fields) > 1 {
						err = cmd.Run(fields[1:])
					} else {
						err = cmd.Run([]string{})
					}
					if err != nil {
						Log.WithError(err).Error("failed to execute dependency command")
						return err
					}
				}
			}
		}
	}

	var (
		argc         = len(args)
		requiredArgs = len(c.args)
		cLog         = Log.WithField("prefix", "runCommand"+strings.ToTitle(c.name))
		start        = time.Now()
	)

	cLog.WithFields(logrus.Fields{
		"name":   c.name,
		"args":   args,
		"params": c.params,
	}).Debug("running command")

	// check if parameters are set on the command
	// in this case ignore the arguments from the commandline and pass the predefined ones
	if len(c.params) > 0 {
		Log.Debug("found predefined params: ", c.params)
		args = c.params
		argc = len(c.params)
	}

	// check args
	if argc != requiredArgs {
		if argc > requiredArgs {
			return ErrTooManyArguments
		}
		cLog.Info("expected: ", getArgumentString(c.args))
		return ErrNotEnoughArguments
	}

	// execute build chain commands
	if len(c.commandChain) > 0 {
		for _, cmd := range c.commandChain {

			// dont pass the args down the commandChain
			err := cmd.Run([]string{})
			if err != nil {
				cLog.WithError(err).Error("failed to execute " + cmd.name)
				return err
			}
		}
	}

	// make script executable
	err := os.Chmod(c.path, 0700)
	if err != nil {
		cLog.WithError(err).Fatal("failed to make script executable")
	}

	var (
		cmd    *exec.Cmd
		script string
	)

	// read the contents of this commands script
	target, err := ioutil.ReadFile(c.path)
	if err != nil {
		l.Fatal(err)
	}

	Log.Debug("injecting args into script: ", args, " cmd: ", c.name)

	// parse arguments and add them to the script
	var argBuf bytes.Buffer
	for i, a := range args {
		if i < len(c.args) {

			// handle argument labels
			if strings.Contains(a, "=") {
				if !strings.HasPrefix(a, c.args[i].name+"=") {
					Log.Error("expected label: " + c.args[i].name + ", got: " + a)
					return ErrInvalidArgumentLabel
				}
				a = strings.TrimPrefix(a, c.args[i].name+"=")
			}

			if !validArgType(a, c.args[i].argType) {
				cLog.WithError(ErrInvalidArgumentType).WithFields(logrus.Fields{
					"value":   a,
					"argName": c.args[i].name,
				}).Error("expected type: ", c.args[i].argType.String())
				return ErrInvalidArgumentType
			}
			argBuf.WriteString(c.args[i].name + "=" + a + "\n")
		}
	}

	// prepend projectGlobals if not empty
	if len(globalsContent) > 0 {

		// add the globals, append argument buffer and then append script contents
		script = string(append(append(globalsContent, argBuf.Bytes()...), target...))

		// create command instance and pass new script to bash
		if conf.StopOnError {
			cmd = exec.Command(p.interpreter, []string{"-e", "-c", script}...)
		} else {
			cmd = exec.Command(p.interpreter, "-c", script)
		}
	} else {

		// add argument buffer and then append script contents
		script = string(append(argBuf.Bytes(), target...))

		// create command instance
		// no globals - only execute target script
		if conf.StopOnError {
			cmd = exec.Command(p.interpreter, []string{"-e", "-c", script}...)
		} else {
			cmd = exec.Command(p.interpreter, "-c", script)
		}
	}

	if conf.Debug {
		printScript(script)
	}

	// set up environment
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = cWriter
	cmd.Env = os.Environ()

	currentCommand++

	if c.buildNumber {
		projectData.BuildNumber++
		projectData.update()
	}

	l.Print(cp.colorText)
	l.Println(printPrompt() + "[" + strconv.Itoa(currentCommand) + "/" + strconv.Itoa(numCommands) + "] executing " + cp.colorPrompt + c.name + ansi.Reset)

	// lets go
	err = cmd.Start()
	if err != nil {
		cLog.WithError(err).Fatal("failed to start command: " + c.name)
	}

	// add to processMap
	id := processID(randomString())
	addProcess(id, c.name, cmd.Process)

	// wait for command to finish execution
	err = cmd.Wait()
	if err != nil {

		// when are no globals, read the command script directly and print it with line numbers to stdout for easy debugging
		if script == "" {
			scriptBytes, err := ioutil.ReadFile(c.path)
			if err != nil {
				cLog.WithError(err).Error("failed to read script")
			}
			script = string(scriptBytes)
		}

		if conf.DumpScriptOnError {
			dumpScript(script)
		}

		cLog.WithError(err).Error("failed to wait for command: " + c.name)
		return err
	}

	// after command has finished running, remove from processMap
	deleteProcess(id)

	// print stats
	l.Println(printPrompt()+"["+strconv.Itoa(currentCommand)+"/"+strconv.Itoa(numCommands)+"] finished "+cp.colorPrompt+c.name+cp.colorText+" in"+cp.colorPrompt, time.Now().Sub(start), ansi.Reset)

	return nil
}

/*
 *	Utils
 */

// addCommand parses the script at path, adds it to the commandMap and sets up the shell completer
// if force is set to true the command will parsed again even when it already has been parsed
func addCommand(path string, force bool) error {

	var (
		cLog = Log.WithField("prefix", "addCommand")

		// create parse job
		job = p.AddJob(path, false)
	)

	if !force {

		commandName := strings.TrimSuffix(filepath.Base(path), f.fileExtension)
		commandMutex.Lock()
		_, ok := commands[commandName]
		commandMutex.Unlock()

		if ok {
			return nil
		}
	}

	// create new command instance
	cmd, err := job.newCommand(path)
	if err != nil {
		return err
	}

	// job done
	p.RemoveJob(job)

	commandMutex.Lock()

	// Add the completer.
	completer.Children = append(completer.Children, cmd.PrefixCompleter)
	// add to command map
	commands[cmd.name] = cmd

	commandMutex.Unlock()

	cLog.Debug("added " + cmd.name + " to the command map")

	return nil
}

// newCommand creates a new command instance for the script at path
// a parseJob will be created
func (job *parseJob) newCommand(path string) (*command, error) {

	var (
		cLog = Log.WithField("prefix", "newCommand")
	)

	// parse the script
	d, err := p.parseScript(path, job)
	if err != nil {
		if !job.silent {
			cLog.WithFields(logrus.Fields{
				"path": path,
			}).Debug("Parse error")
		}
		return nil, err
	}

	// get build chain
	commandChain, err := job.getCommandChain(d.parsedCommands)
	if err != nil {
		return nil, err
	}

	// get name for command
	name := strings.TrimSuffix(strings.TrimPrefix(path, zeusDir+"/"), f.fileExtension)
	if name == "" {
		return nil, ErrEmptyName
	}

	return &command{
		path:            path,
		name:            name,
		args:            d.args,
		manual:          d.manual,
		help:            d.help,
		commandChain:    commandChain,
		PrefixCompleter: readline.PcItem(name),
		buildNumber:     d.buildNumber,
		dependencies:    d.dependencies,
		outputs:         d.outputs,
	}, nil
}

// assemble a commandChain with a list of parsed commands and their arguments
func (job *parseJob) getCommandChain(parsedCommands [][]string) (commandChain commandChain, err error) {

	var cLog = Log.WithFields(logrus.Fields{
		"prefix":         "getCommandChain",
		"parsedCommands": parsedCommands,
	})

	cLog.Debug("creating commandChain, job.commands: ", job.commands)

	// empty commandChain is OK
	for _, args := range parsedCommands {

		var count int

		// check if there are repetitive targets in the chain - this is not allowed to prevent cycles
		for _, c := range job.commands {

			// check if the key (commandName) is already there
			if c[0] == args[0] {
				count++
			}
		}

		if count > p.recursionDepth {
			cLog.WithFields(logrus.Fields{
				"count":          count,
				"path":           job.path,
				"parsedCommands": parsedCommands,
				"job.commands":   job.commands,
			}).Fatal("CYCLE DETECTED! -> ", args[0], " appeared more than ", p.recursionDepth, " times - thats invalid.")
		}

		job.commands = append(job.commands, args)

		// check if command has already been parsed
		commandMutex.Lock()
		cmd, ok := commands[args[0]]
		commandMutex.Unlock()

		if !ok {

			// add new command
			cmd, err = job.newCommand(zeusDir + "/" + args[0] + f.fileExtension)
			if err != nil {
				if !job.silent {
					cLog.WithError(err).Debug("failed to create command")
				}
				return commandChain, err
			}

			commandMutex.Lock()

			// add the completer
			completer.Children = append(completer.Children, cmd.PrefixCompleter)

			// add to command map
			commands[args[0]] = cmd

			commandMutex.Unlock()

			cLog.Debug("added " + cmd.name + " to the command map")
		}

		cLog.Debug("adding command to build chain: ", args)

		// this command has argument parameters in its commandChain
		// set them on the command
		if len(args) > 1 {

			cLog.WithFields(logrus.Fields{
				"command": args[0],
				"params":  args[1:],
			}).Debug("setting parameters")

			// creating a hard copy of the struct here,
			// otherwise params would be set for every execution of the command
			cmd = &command{
				name:            cmd.name,
				path:            cmd.path,
				params:          args[1:],
				args:            cmd.args,
				manual:          cmd.manual,
				help:            cmd.help,
				commandChain:    cmd.commandChain,
				PrefixCompleter: cmd.PrefixCompleter,
				buildNumber:     cmd.buildNumber,
			}
		}

		// append command to build chain
		commandChain = append(commandChain, cmd)
	}
	return
}

// parse and execute a given commandChain string
func executeCommandChain(chain string) {

	var (
		cLog = Log.WithField("prefix", "executeCommandChain")
		job  = p.AddJob(chain, false)
	)

	commandList := parseCommandChain(chain)
	commandChain, err := job.getCommandChain(commandList)
	if err != nil {
		cLog.WithError(err).Error("failed to get command chain")
	}

	p.RemoveJob(job)

	numCommands = countCommandChain(commandChain)

	for _, c := range commandChain {
		err := c.Run([]string{})
		if err != nil {
			cLog.WithError(err).Error("failed to execute " + c.name)
		}
	}
}

// walk all scripts in the zeus dir and setup commandMap and globals
func findCommands() {

	var (
		cLog    = Log.WithField("prefix", "findCommands")
		start   = time.Now()
		scripts []string
		wg      sync.WaitGroup
		// keep track of scripts that couldn't be parsed
		parseErrors      = make(map[string]error, 0)
		parseErrorsMutex = &sync.Mutex{}
	)

	// walk zeus directory and initialize scripts
	err := filepath.Walk(zeusDir, func(path string, info os.FileInfo, err error) error {

		// ignore self
		if path != zeusDir {

			// ignore sub directories
			if info.IsDir() {
				return filepath.SkipDir
			}

			// check if its a valid script
			if strings.HasSuffix(path, f.fileExtension) {

				// check for globals script
				// the globals script wont be parsed for zeus header fields
				if strings.HasPrefix(strings.TrimPrefix(path, zeusDir+"/"), "globals") {

					g, err := ioutil.ReadFile(zeusDir + "/globals.sh")
					if err != nil {
						l.Fatal(err)
					}

					// add newline to prevent parse errors
					globalsContent = append(g, []byte("\n")...)
					return nil
				}

				scripts = append(scripts, path)
			}
		}

		return nil
	})
	if err != nil {
		cLog.WithError(err).Fatal("failed to walk zeus directory")
	}

	wg.Add(1)

	// first half
	go func() {
		for _, path := range scripts[:len(scripts)/2] {
			err := addCommand(path, false)
			if err != nil {
				Log.WithError(err).Debug("failed to add command")
				parseErrorsMutex.Lock()
				parseErrors[path] = err
				parseErrorsMutex.Unlock()
			}
		}
		wg.Done()
	}()

	wg.Add(1)

	if conf.Debug {
		time.Sleep(500 * time.Millisecond)
		l.Println("--------------------------------------------------------------------------")
	}

	// second half
	go func() {
		for _, path := range scripts[len(scripts)/2:] {
			err := addCommand(path, false)
			if err != nil {
				Log.WithError(err).Debug("failed to add command")
				parseErrorsMutex.Lock()
				parseErrors[path] = err
				parseErrorsMutex.Unlock()
			}
		}
		wg.Done()
	}()

	wg.Wait()

	for path, err := range parseErrors {
		Log.WithError(err).Error("failed to parse: ", path)
	}
	if len(parseErrors) > 0 {
		println()
	}

	l.Println(cp.colorText+"initialized "+cp.colorPrompt, len(commands), cp.colorText+" commands in: "+cp.colorPrompt, time.Now().Sub(start), ansi.Reset+"\n")

	// check if custom command conflicts with builtin name
	for _, name := range builtins {
		if _, ok := commands[name]; ok {
			cLog.Error("command ", name, " conflicts with a builtin command. Please choose a different name.")
		}
	}

	var commandCompletions []readline.PrefixCompleterInterface
	for _, c := range commands {
		commandCompletions = append(commandCompletions, readline.PcItem(c.name))
	}

	// add all commands to the completer for the help page
	for _, c := range completer.Children {
		if string(c.GetName()) == "help " {
			c.SetChildren(commandCompletions)
		}
	}
}

// run an alias command (allows shell commands)
// first checks for zeus commands then passes it to the shell
func executeCommand(command string) {

	s := strings.Split(command, " ")
	if len(s) > 0 {

		// check if first command is known to zeus
		// check user commands
		if _, ok := commands[s[0]]; ok {
			executeCommandChain(command)
			return
		}

		// check builtins
		for _, b := range builtins {
			if b == s[0] {
				executeCommandChain(command)
				return
			}
		}

		if conf.PassCommandsToShell {

			// not an alias - pass to shell
			var err error
			if len(s) > 1 {
				err = passCommandToShell(s[0], s[1:])
			} else {
				err = passCommandToShell(s[0], []string{})
			}
			if err != nil {
				l.Println(err)
			}
		}
	}
}
