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
	"fmt"
	"github.com/mgutz/ansi"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

type processID string

var (
	// process instances for all spawned commands, for cleaning up when we leave
	processMap      = make(map[processID]*Process, 0)
	processMapMutex = &sync.Mutex{}
)

// Process keeps track of an os.Process
type Process struct {

	// command name
	Name string

	// internal ID
	ID processID

	// OS PID
	PID int

	// underlying process
	Proc *os.Process
}

// add a process to the store
// thread safe
func addProcess(id processID, name string, p *os.Process, pid int) {
	processMapMutex.Lock()
	defer processMapMutex.Unlock()
	processMap[id] = &Process{
		Name: name,
		ID:   id,
		PID:  pid,
		Proc: p,
	}
}

// delete a process from the store
// thread safe
func deleteProcess(id processID) {
	processMapMutex.Lock()
	delete(processMap, id)
	processMapMutex.Unlock()
}

// delete a process from the store by its PID
// thread safe
func deleteProcessByPID(pid int) {
	processMapMutex.Lock()
	for id, p := range processMap {
		if p.PID == pid {
			delete(processMap, id)
		}
	}
	processMapMutex.Unlock()
}

// cleanup before we leave
// used only on unrecoverable errors
func cleanup(cmdFile *CommandsFile) {

	// reset terminal colors
	fmt.Print(ansi.Reset)

	// kill all spawned processes
	clearProcessMap()

	// invoke exitHook if set
	if cmdFile != nil {
		if cmdFile.ExitHook != "" {
			out, err := exec.Command(cmdFile.ExitHook).CombinedOutput()
			if err != nil {
				fmt.Println(string(out))
				log.Fatal("exitHook failed: ", err)
			}
		}
	}

	// close readline
	if rl != nil {
		err := rl.Close()
		if err != nil {
			Log.WithError(err).Error("failed to close readline")
		}
	}
}

// clean up the mess
func clearProcessMap() {

	// l.Println("processMap:", processMap)

	processMapMutex.Lock()
	defer processMapMutex.Unlock()

	// range processes
	for id, p := range processMap {
		if p.Proc != nil {

			Log.Debug("killing process with ID: "+id+" and PID:", p.Proc.Pid)

			// kill it
			err := p.Proc.Kill()
			if err != nil {
				Log.WithError(err).Debug("failed to kill process with ID: "+id+" and PID:", p.Proc.Pid)
			}
		}
	}
}

// clean up the mess
func passSignalToProcs(sig os.Signal) {

	// l.Println("processMap:", processMap)

	processMapMutex.Lock()
	defer processMapMutex.Unlock()

	// range processes
	for _, p := range processMap {
		if p.Proc != nil {

			Log.Debug("passing signal "+sig.String()+" to PID: ", p.Proc.Pid)

			err := p.Proc.Signal(sig)
			if err != nil {
				Log.WithError(err).Debug("failed to pass signal "+sig.String()+" to PID:", p.Proc.Pid)
			}
		}
	}
}

func printProcsCommandUsageErr() {
	l.Println(ErrInvalidUsage)
	l.Println("usage: procs [detach <command>] [attach <pid>] [kill <pid>]")
}

// manage spawned processes
func handleProcsCommand(args []string) {

	if len(args) < 3 {
		printProcs()
		return
	}

	switch args[1] {
	// detach any command async
	case "detach":
		if cmd, ok := cmdMap.items[args[2]]; ok {
			cmd.async = true
			err := cmd.Run(args[3:], true)
			if err != nil {
				Log.WithError(err).Error("failed to run command. args: ", args[3:])
			}
			time.Sleep(100 * time.Millisecond)
			cmd.async = false
		} else {
			l.Println("invalid command:", args[2])
		}
	// attach to a runnning async process with screen -r
	case "attach":
		cmd := exec.Command("screen", "-r", args[2])
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Env = os.Environ()
		err := cmd.Run()
		if err != nil {
			Log.WithError(err).Error("failed to attach to PID: ", args[2])
		}
	// kill a process by PID
	case "kill":
		pid, err := strconv.Atoi(args[2])
		if err != nil {
			Log.WithError(err).Error("invalid integer value: ", args[2])
			return
		}
		err = exec.Command("kill", args[2]).Run()
		if err != nil {
			Log.WithError(err).Error("failed to kill PID: ", args[2])
			return
		}
		deleteProcessByPID(pid)
	default:
		printProcsCommandUsageErr()
	}
}

func printProcs() {
	processMapMutex.Lock()
	defer processMapMutex.Unlock()

	l.Println(cp.Prompt + pad("ID", 20) + pad("PID", 10) + "Name")
	for _, p := range processMap {
		l.Println(cp.Text + pad(string(p.ID), 20) + pad(strconv.Itoa(p.PID), 10) + p.Name)
	}
}
