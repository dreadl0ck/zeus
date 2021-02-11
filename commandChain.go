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
	"sync"

	"github.com/mgutz/ansi"
)

type status struct {
	recursionMap   map[string]int
	numCommands    int
	currentCommand int
	sync.RWMutex
}

func (s *status) reset() {
	// reset counters
	s.Lock()
	s.numCommands = 0
	s.currentCommand = 0
	s.recursionMap = make(map[string]int, 0)
	s.Unlock()
}

func (s *status) incrementRecursionCount(commandName string) error {

	conf.Lock()
	limit := conf.fields.RecursionDepth
	conf.Unlock()

	s.Lock()
	defer s.Unlock()

	if val, ok := s.recursionMap[commandName]; ok {
		if val == limit {
			return errors.New("recursion limit for command " + commandName + " reached.")
		}
		s.recursionMap[commandName]++
		Log.Debug("incremented recursion count for command "+ansi.Red+commandName+cp.Reset+" to: ", s.recursionMap[commandName])
	} else {
		s.recursionMap[commandName] = 1
		Log.Debug("adding " + ansi.Red + commandName + cp.Reset + " to recursionMap")
	}
	return nil
}

type commandChain []*command

// create a readable string from a commandChain
// example: clean -> build name=testBuild -> install
func (cmdChain commandChain) String() (out string) {

	for i, cmd := range cmdChain {

		out += cmd.name

		// if not last elem
		if !(i == len(cmdChain)-1) {
			out += " -> "
		}
	}
	return
}

// parse and execute a given commandChain string
func (cmdChain commandChain) exec(cmds []string) {

	defer s.reset()

	// set numCommands counter
	for _, c := range cmdChain {
		count, err := getTotalDependencyCount(c)
		if err != nil {
			Log.WithError(err).Error("failed to get dependency count")
			return
		}
		s.Lock()
		s.numCommands += count
		s.Unlock()
	}

	// exec and pass args
	for i, c := range cmdChain {
		err := c.Run(strings.Fields(cmds[i])[1:], c.async)
		if err != nil {
			Log.WithError(err).Error("failed to execute " + c.name)
			return
		}
	}
}

// check if its a valid command chain
// returns an initialized commandChain with all the commands
// and a boolean inidicating wheter its valid or not
func validCommandChain(commands []string, quiet bool) (commandChain, bool) {

	var (
		cmdChain     commandChain
		recursionMap = make(map[string]int, 0)
	)

	if len(commands) < 1 {
		l.Println("invalid command chain: empty")
		return nil, false
	}

	conf.Lock()
	maxRecursion := conf.fields.RecursionDepth
	conf.Unlock()

	for index, entry := range commands {

		fields := strings.Fields(entry)
		if len(fields) > 0 {

			// check if command exists
			cmd, err := cmdMap.getCommand(fields[0])
			if err != nil {
				if !quiet {
					l.Println(err)
				}
				return nil, false
			}

			// validate args
			_, err = cmd.parseArguments(fields[1:])
			if err != nil {
				l.Println(err)
				return nil, false
			}

			if val, ok := recursionMap[cmd.name]; ok {
				if val == maxRecursion {
					l.Println("recursion limit", maxRecursion, "reached for command:", cmd.name)
					return nil, false
				}
				recursionMap[cmd.name]++
			} else {
				recursionMap[cmd.name] = 1
			}

			// add to commandChain
			cmdChain = append(cmdChain, cmd)
		} else {
			l.Println("empty field at index:", index)
			return nil, false
		}
	}

	return cmdChain, true
}
