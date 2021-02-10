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
	"strconv"
	"strings"

	"github.com/dreadl0ck/readline"
)

// command header
// used in scripts to supply information for ZEUS
type commandData struct {

	// one line Description text
	Description string `yaml:"description"`

	// scripting language of the command
	Language string `yaml:"language"`

	// Help page text
	Help string `yaml:"help"`

	// Arguments
	Arguments []string `yaml:"arguments"`

	// Dependencies
	Dependencies []string `yaml:"dependencies"`

	// outputs
	Outputs []string `yaml:"outputs"`

	// increase buildnumber on each execution
	BuildNumber bool `yaml:"buildNumber"`

	// execute command in a detached screen session
	Async bool `yaml:"async"`

	// Exec is the script to run when executed
	Exec string `yaml:"exec"`

	// Path allows to set a custom path for the command
	Path string `yaml:"path"`

	// Hidden controls if the command is shown the menu
	Hidden bool `yaml:"hidden"`

	// CanModifyPrompt controls if the command can modify the zeus prompt
	CanModifyPrompt bool `yaml:"canModifyPrompt"`

	// UseBase sets the base configuration to use for this command
	UseBase string `yaml:"useBase"`
}

// initialize a command from a commandData instance
// returns if command does already exist
func (d *commandData) init(commandsFile *CommandsFile, name string) error {

	// check if a path is set and an exec section specified
	// that's invalid - print debug info and return an error
	if d.Path != "" && d.Exec != "" {

		c, err := ioutil.ReadFile(commandsFilePath)
		if err != nil {
			return err
		}
		var (
			highlightLine  int
			commandStarted bool
		)
		for index, line := range strings.Split(string(c), "\n") {
			if strings.Contains(line, name+":") {
				commandStarted = true
			}
			if commandStarted {
				if strings.Contains(line, d.Path) {
					highlightLine = index
					break
				}
			}
		}

		printCodeSnippet(string(c), commandsFilePath, highlightLine)
		return errors.New("command " + name + " has custom path set, but specifies an exec action")
	}

	// check if the current command is a dependency and abort if true
	for index, dep := range d.Dependencies {
		fields := strings.Fields(dep)
		if len(fields) >= 1 {
			if fields[0] == name {
				c, err := ioutil.ReadFile(commandsFilePath)
				if err != nil {
					return err
				}
				var (
					highlightLine  int
					commandStarted bool
				)
				for index, line := range strings.Split(string(c), "\n") {
					if strings.Contains(line, name+":") {
						commandStarted = true
					}
					if commandStarted {
						if strings.Contains(line, "- "+name) {
							highlightLine = index
							break
						}
					}
				}

				printCodeSnippet(string(c), commandsFilePath, highlightLine)
				return errors.New("command " + name + " has itself as dependency at index: " + strconv.Itoa(index) + " This will result in a loop")
			}
		}
	}

	// assemble commands args
	args, err := commandsFile.validateArgs(d.Arguments)
	if err != nil {
		return errors.New("command " + name + ": " + err.Error())
	}

	var lang string
	if d.Language == "" {
		lang = commandsFile.Language
	} else {
		lang = d.Language
	}

	// create command
	cmd := &command{
		path:        d.Path,
		name:        name,
		args:        args,
		description: d.Description,
		help:        d.Help,
		hidden:      d.Hidden,
		// PrefixCompleter: readline.PcItem(name,
		// 	readline.PcItemDynamic(func(path string) (res []string) {

		// 		// fmt.Println("\npath:", path)

		// 		var allRequiredArgsSet = true
		// 		for _, a := range args {
		// 			if !strings.Contains(path, a.name+"=") {
		// 				res = append(res, a.name+"=")
		// 				if !a.optional {
		// 					allRequiredArgsSet = false
		// 				}
		// 			}
		// 		}

		// 		if !allRequiredArgsSet {
		// 			return
		// 		}

		// 		if allRequiredArgsSet && strings.HasSuffix(path, commandChainSeparator+" ") {
		// 			// return all available commands
		// 			cmdMap.Lock()
		// 			defer cmdMap.Unlock()
		// 			for name := range cmdMap.items {
		// 				res = append(res, name)
		// 			}
		// 			//					fmt.Println("allRequiredArgsSet:", allRequiredArgsSet, "commands result:", res)
		// 			return
		// 		}
		// 		if allRequiredArgsSet {
		// 			res = append(res, "->")
		// 		}
		// 		return
		// 	}),
		// ),
		PrefixCompleter: readline.PcItem(name,

			// completer for current commands arguments
			readline.PcItemDynamic(func(path string) (res []string) {
				var allRequiredArgsSet = true
				for _, a := range args {
					if !strings.Contains(path, a.name+"=") {
						res = append(res, a.name+"=")
						if !a.optional {
							allRequiredArgsSet = false
						}
					}
				}
				if allRequiredArgsSet {
					res = append(res, commandChainSeparator)
				}
				// l.Println("\npath:", path)
				// l.Println("result:", res)
				return
			},

				// completer for next command names
				readline.PcItemDynamic(func(path string) (res []string) {

					// return all available commands
					cmdMap.Lock()
					defer cmdMap.Unlock()
					for name := range cmdMap.items {
						res = append(res, name)
					}
					// l.Println("\npath:", path)
					// l.Println("result:", res)
					return
				},

					// completer for next commands args
					readline.PcItemDynamic(func(path string) (res []string) {

						slice := strings.Split(path, commandChainSeparator)
						if len(slice) == 0 {
							return
						}

						cmdArgSlice := strings.Fields(slice[len(slice)-1])
						if len(cmdArgSlice) == 0 {
							return
						}

						cmdMap.Lock()
						c, ok := cmdMap.items[cmdArgSlice[0]]
						if !ok {
							cmdMap.Unlock()
							return
						}
						cmdMap.Unlock()

						// return the next commands completer?
						// return c.PrefixCompleter.Callback(path)

						var allRequiredArgsSet = true
						for _, a := range c.args {
							if !strings.Contains(path, a.name+"=") {
								res = append(res, a.name+"=")
								if !a.optional {
									allRequiredArgsSet = false
								}
							}
						}
						if allRequiredArgsSet {
							res = append(res, commandChainSeparator)
						}

						// l.Println("\npath:", path)
						// l.Println("result:", res)
						return
					}),
				),
			),
		),
		buildNumber:     d.BuildNumber,
		dependencies:    d.Dependencies,
		outputs:         d.Outputs,
		exec:            d.Exec,
		async:           d.Async,
		language:        lang,
		canModifyPrompt: d.CanModifyPrompt,
		useBase:         d.UseBase,
	}

	if lang == "go" && d.Exec != "" {
		return errors.New("when using Go, use of the exec field is not allowed. Please a create a file in the scripts folder instead")
	}

	// replace globals in outputs
	for i, o := range cmd.outputs {
		out, err := commandsFile.replaceGlobals(o)
		if err != nil {
			return err
		}
		cmd.outputs[i] = out
	}

	// disable completion for hidden commands
	if d.Hidden {
		cmd.PrefixCompleter = readline.NewPrefixCompleter()
	}

	if d.Exec == "" {
		if d.Path == "" {
			l, err := cmd.getLanguage()
			if err != nil {
				return err
			}
			cmd.path = scriptDir + "/" + name + l.FileExtension
		}
	}

	var exists bool

	// update the completer if a completion exists
	completer.Lock()
	for i, c := range completer.Children {
		if string(cmd.PrefixCompleter.GetName()) == string(c.GetName()) {
			exists = true
			// update completer
			completer.Children[i] = cmd.PrefixCompleter
		}
	}

	// add to completer if none exists
	if !exists {
		completer.Children = append(completer.Children, cmd.PrefixCompleter)
	}
	completer.Unlock()

	// add to command map
	cmdMap.Lock()
	cmdMap.items[cmd.name] = cmd
	cmdMap.Unlock()

	Log.WithField("prefix", "parseCommandsFile").Debug("added " + cp.CmdName + cmd.name + cp.Reset + " to the command map")

	// if debug {
	// 	cmd.dump()
	// }

	return nil
}
