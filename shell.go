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
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"time"

	"github.com/dreadl0ck/readline"
	"github.com/mgutz/ansi"
)

var (

	// ErrUnknownCommand occurs when the command requested is not known to zeus
	ErrUnknownCommand = errors.New("unknown command")

	// global readline instance
	rl            *readline.Instance
	readlineMutex = &sync.Mutex{}
)

// readline loop for interactive mode
// when there's an unknown command it will be passed to the shell
func readlineLoop() error {

	if conf.PrintBuiltins {
		printBuiltins()
	}

	// print overview
	printCommands()

	var (
		historyFileName string
		err             error
	)

	if conf.HistoryFile {
		historyFileName = zeusDir + "/zeus_history"
	}

	conf.Lock()
	historyLimit := conf.HistoryLimit
	conf.Unlock()

	readlineMutex.Lock()
	// prepare readline
	rl, err = readline.NewEx(&readline.Config{
		Prompt:          printPrompt(),
		AutoComplete:    completer,
		HistoryLimit:    historyLimit,
		HistoryFile:     historyFileName,
		Listener:        listener,
		InterruptPrompt: "\nBye." + ansi.Reset,
	})
	readlineMutex.Unlock()
	if err != nil {
		return err
	}
	defer rl.Close()

	for {

		// read a line
		line, err := rl.Readline()
		if err != nil {

			if err == io.EOF {
				return nil
			}

			if err == readline.ErrInterrupt {

				if conf.ExitOnInterrupt {
					clearProcessMap()
					os.Exit(0)
				} else {
					Log.Info("ExitOnInterrupt is disabled, type 'exit' if you want to leave.")
					continue
				}
			}

			return fmt.Errorf("readline error: %v", err)
		}

		handleLine(line)
	}
}

// handle input line read by the readline instance
func handleLine(line string) {

	// trim
	line = strings.TrimSpace(line)

	// set the color
	print(cp.cmdOutput)

	switch line {
	case exitCommand:
		l.Println(cp.text + "Bye." + ansi.Reset)
		clearProcessMap()
		os.Exit(0)

	case helpCommand:

		clearScreen()

		l.Println(cp.text + asciiArt + ansi.Reset + "\n")
		l.Println(cp.text + "Project Name: " + cp.prompt + filepath.Base(workingDir) + cp.text + "\n")

		conf.Lock()
		if conf.PrintBuiltins {
			printBuiltins()
		}
		conf.Unlock()
		printCommands()

	case infoCommand:
		printProjectInfo()

	case zeusfileCommand:
		err := migrateZeusfile()
		if err != nil {
			Log.WithError(err).Error("failed to migrate zeusfile")
		}

	case formatCommand:
		f.formatCommand()

	case "zeus": // prevent spawning a new interactive shell

	case globalsCommand:
		listGlobals()

	case configCommand:
		conf.dump()

	case wikiCommand:
		go StartWebListener(false)
		open("http://" + hostName + ":" + strconv.Itoa(conf.PortWebPanel) + "/wiki")

	case webCommand:
		go StartWebListener(true)

	case dataCommand:
		printProjectData()

	case updateCommand:
		updateZeus()

	case versionCommand:
		l.Println(version)

	case clearCommand:

		clearScreen()

		l.Println(cp.text + asciiArt + ansi.Reset + "\n")
		l.Println(cp.text + "Project Name: " + cp.prompt + filepath.Base(workingDir) + cp.text + "\n")

	case builtinsCommand:
		printBuiltins()

	default:

		// split the input line
		args := strings.Fields(line)

		// skip if empty
		if len(args) == 0 {
			return
		}

		// get the command name
		commandName := args[0]

		switch commandName {
		case makefileCommand:
			handleMakefileCommand(args)
		case configCommand:
			handleConfigCommand(args)
		case eventsCommand:
			handleEventsCommand(args)
		case aliasCommand:
			handleAliasCommand(args)
		case deadlineCommand:
			handleDeadlineCommand(args)
		case gitFilterCommand:
			handleGitFilterCommand(args)
		case milestonesCommand:
			handleMilestonesCommand(args)
		case procsCommand:
			handleProcsCommand(args)
		case helpCommand:
			handleHelpCommand(args)
		case colorsCommand:
			handleColorsCommand(args)
		case authorCommand:
			handleAuthorCommand(args)
		case keysCommand:
			handleKeysCommand(args)
		case createCommand:
			handleCreateCommand(args)
		case todoCommand:
			handleTodoCommand(args)

		default:
			// check if its a commandchain
			if strings.Contains(line, p.separator) {
				executeCommandChain(line)
				return
			}

			// remove the command name from the slice
			args = args[1:]

			cmdMap.Lock()

			// try to find the command in the commands map
			cmd, ok := cmdMap.items[commandName]
			if !ok {
				cmdMap.Unlock()

				projectData.Lock()

				// check if its an alias
				if command, ok := projectData.Aliases[commandName]; ok {

					projectData.Unlock()
					handleLine(command)

					// reset counters
					numCommands = 0
					currentCommand = 0
					return
				}
				projectData.Unlock()

				// not an alias - pass to shell
				if conf.PassCommandsToShell {
					err := passCommandToShell(commandName, args)
					if err != nil {
						l.Println(err)
					}
				} else {
					l.Println(ErrUnknownCommand, ": ", commandName)
				}
				return
			}
			cmdMap.Unlock()

			numCommands = numCommands + getTotalCommandCount(cmd)

			// run the command
			err := cmd.Run(args, cmd.async)
			if err != nil {
				fmt.Printf("command "+cmd.name+" failed. error: %v\n", err)
			}

			if cmd.async {
				time.Sleep(100 * time.Millisecond)
			}

			// reset counters
			numCommands = 0
			currentCommand = 0
		}
	}
}
