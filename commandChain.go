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
	"strings"
	"sync"
)

type status struct {
	numCommands    int
	currentCommand int
	sync.RWMutex
}

type commandChain []*command

// execute a commandChain directly
func (cmdChain commandChain) exec() {

	s.Lock()
	s.numCommands = countDependencies(cmdChain)
	s.Unlock()

	for _, c := range cmdChain {
		err := c.Run([]string{}, c.async)
		if err != nil {
			Log.WithError(err).Error("failed to execute " + c.name)
		}
	}

	// reset counters
	s.Lock()
	s.numCommands = 0
	s.currentCommand = 0
	s.Unlock()
}

// parse and execute a given commandChain string
func parseAndExecuteCommandChain(chain string) {

	var (
		cLog = Log.WithField("prefix", "parseAndExecuteCommandChain")
		cmds = strings.Split(chain, commandChainSeparator)
	)

	if len(cmds) < 1 {
		l.Println("invalid command chain")
		return
	}

	firstCommand := strings.Fields(cmds[0])
	p, err := getParserForScript(firstCommand[0])
	if err != nil {
		Log.WithError(err).Error("failed to get parser for command:" + firstCommand[0])
		return
	}

	job := p.AddJob(chain, false)
	defer p.RemoveJob(job)

	commandList, err := job.parseCommandChain(chain)
	if err != nil {
		cLog.WithError(err).Error("failed to parse command chain")
		return
	}

	cmdChain, err := job.getCommandChain(commandList, nil)
	if err != nil {
		cLog.WithError(err).Error("failed to get command chain")
		return
	}

	s.Lock()
	s.numCommands = countDependencies(cmdChain)
	s.Unlock()

	for _, c := range cmdChain {
		err := c.Run([]string{}, c.async)
		if err != nil {
			cLog.WithError(err).Error("failed to execute " + c.name)
		}
	}

	// reset counters
	s.Lock()
	s.numCommands = 0
	s.currentCommand = 0
	s.Unlock()
}

// check if its a valid command chain
// returns an initialized commandChain with all the commands
// and a boolean inidicating wheter its valid or not
func validCommandChain(args []string, silent bool) (commandChain, bool) {

	var (
		chain = strings.Join(args, " ")
		cmds  = strings.Split(chain, commandChainSeparator)
	)

	if len(cmds) < 1 {
		l.Println("invalid command chain")
		return nil, false
	}

	firstCommand := strings.Fields(cmds[0])

	p, err := getParserForScript(firstCommand[0])
	if err != nil {
		Log.WithError(err).Error("failed to get parser for command:" + firstCommand[0])
		return nil, false
	}

	job := p.AddJob(chain, silent)

	commandList, err := job.parseCommandChain(chain)
	if err != nil {
		Log.WithError(err).Error("failed to parse command chain")
		return nil, false
	}

	defer p.RemoveJob(job)

	cmdChain, err := job.getCommandChain(commandList, nil)
	if err != nil {
		if !silent {
			Log.WithError(err).Error("failed to get command chain")
		}
		return nil, false
	}

	return cmdChain, true
}
