/*
 *  ZEUS - A Powerful Build System
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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/mgutz/ansi"
)

var (

	// ErrUnknownCommand occurs when the command requested is not known to zeus
	ErrUnknownCommand = errors.New("unknown command")

	// global readline instance
	rl *readline.Instance
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
		historyFileName = workingDir + "/zeus/zeus_history"
	}

	// prepare readline
	rl, err = readline.NewEx(&readline.Config{
		Prompt:          printPrompt(),
		AutoComplete:    completer,
		HistoryLimit:    conf.HistoryLimit,
		HistoryFile:     historyFileName,
		Listener:        listener,
		InterruptPrompt: "\nBye.",
	})
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
	print(cp.colorCommandOutput)

	switch line {
	case exitCommand:
		l.Println(cp.colorText + "Bye.")
		clearProcessMap()
		os.Exit(0)

	case helpCommand:

		clearScreen()

		l.Println(cp.colorText + asciiArt + ansi.Reset + "\n")
		l.Println(cp.colorText + "Project Name: " + cp.colorPrompt + filepath.Base(workingDir) + cp.colorText + "\n")

		if conf.PrintBuiltins {
			printBuiltins()
		}
		printCommands()

	case infoCommand:
		printProjectInfo()

	case formatCommand:
		f.formatCommand()

	case "zeus": // prevent spawning a new interactive shell

	case globalsCommand:
		listGlobals()

	case configCommand:
		printConfiguration()

	case dataCommand:
		printProjectData()

	case versionCommand:
		l.Println(version)

	case clearCommand:

		clearScreen()

		l.Println(cp.colorText + asciiArt + ansi.Reset + "\n")
		l.Println(cp.colorText + "Project Name: " + cp.colorPrompt + filepath.Base(workingDir) + cp.colorText + "\n")

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
		case milestonesCommand:
			handleMilestonesCommand(args)
		case helpCommand:
			handleHelpCommand(args)
		case colorsCommand:
			handleColorsCommand(args)
		case authorCommand:
			handleAuthorCommand(args)
		case keysCommand:
			handleKeysCommand(args)

		default:
			// check if its a commandchain
			if strings.Contains(line, p.separator) {
				executeCommandChain(line)
				return
			}

			// remove the command name from the slice
			args = args[1:]

			// try to find the command in the commands map
			cmd, ok := commands[commandName]
			if !ok {

				// check if its an alias
				if command, ok := projectData.Aliases[commandName]; ok {

					executeCommand(command)

					// reset counters
					numCommands = 0
					currentCommand = 0
					return

				}

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

			numCommands = getTotalCommandCount(cmd)

			// run the command
			err := cmd.Run(args)
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}

			// reset counters
			numCommands = 0
			currentCommand = 0
		}
	}
}
