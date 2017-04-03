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
	"strings"

	"github.com/dreadl0ck/readline"
)

// ErrInvalidAlias means there is a name conflict with an existing command
var ErrInvalidAlias = errors.New("invalid alias")

func printAliasCommandErr() {
	Log.Error(ErrInvalidUsage)
	Log.Info("usage: alias [remove <name>] [set <name> <command>]")
}

// check if an alias name conflicts with builtin user defined command names
func validateAlias(name string) error {

	// check for conflict with builtin
	for _, b := range builtins {
		if b == name {
			Log.Error("alias ", name, " conflicts with builtin: ", b)
			return ErrInvalidAlias
		}
	}

	// check for conflict with user command
	if command, ok := commands[name]; ok {
		Log.Error("alias ", name, " conflicts with command: ", command.path)
		return ErrInvalidAlias
	}

	return nil
}

// add an alias to project data and shell completer
func addAlias(name, command string) {

	err := validateAlias(name)
	if err != nil {
		Log.WithError(err).Error("failed to validate alias: ", name)
		return
	}

	// add to project data
	projectDataMutex.Lock()
	projectData.Aliases[name] = command
	projectDataMutex.Unlock()

	projectData.update()

	// add to completer
	completerLock.Lock()
	completer.Children = append(completer.Children, readline.PcItem(name, nil))
	completerLock.Unlock()
}

func deleteAlias(name string) {
	projectDataMutex.Lock()
	delete(projectData.Aliases, name)
	projectDataMutex.Unlock()
}

// print alias names to stdout
func printAliases() {

	var maxLen int

	projectDataMutex.Lock()
	for name := range projectData.Aliases {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	for name, command := range projectData.Aliases {
		l.Println(pad(name, maxLen+1), "=", command)
	}

	projectDataMutex.Unlock()
}

// handle alias shell command
func handleAliasCommand(args []string) {

	if len(args) < 2 {
		printAliases()
		return
	}

	if len(args) < 3 {
		printAliasCommandErr()
		return
	}

	switch args[1] {
	case "set":
		if len(args) < 3 {
			printAliasCommandErr()
			return
		}
		addAlias(args[2], strings.Join(args[3:], " "))
	case "remove":
		deleteAlias(args[2])
	default:
		printAliasCommandErr()
	}
}
