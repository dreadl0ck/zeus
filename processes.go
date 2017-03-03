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
	"net/http"
	"os"
	"sync"
	"time"
)

var (

	// process instances for all spawned commands, for cleaning up when we leave
	processMap      = make(map[processID]*Process, 0)
	processMapMutex = &sync.Mutex{}
)

// Process keeps track of an os.Process
type Process struct {
	Name string
	ID   string
	Proc *os.Process
}

type processID string

// add a process to the store
// thread safe
func addProcess(id processID, name string, p *os.Process) {
	processMapMutex.Lock()
	processMap[id] = &Process{
		Name: name,
		ID:   string(id),
		Proc: p,
	}
	processMapMutex.Unlock()
}

// delete a process from the store
// thread safe
func deleteProcess(id processID) {
	processMapMutex.Lock()
	delete(processMap, id)
	processMapMutex.Unlock()
}

// cleanup before we leave
func cleanup() {

	// gracefully shutdown prometheus
	c := &http.Client{}
	_, err := c.Post("http://localhost:9090/-/quit", "text", nil)
	if err == nil {
		time.Sleep(200 * time.Millisecond)
	}

	// kill all spawned processes
	clearProcessMap(nil)

	// close readline
	if rl != nil {
		err := rl.Close()
		if err != nil {
			Log.WithError(err).Error("failed to close readline")
		}
	}

	// close logfileHandle
	if logfileHandle != nil {
		err := logfileHandle.Close()
		if err != nil {
			Log.WithError(err).Error("failed to close logfileHandle")
		}
	}
}

// print all registered processes
func printProcessMap() {
	l.Println("processMap: ", len(processMap))
	for id, p := range processMap {
		if p.Proc != nil {
			l.Println("ID:", id, "PID:", p.Proc.Pid)
		} else {
			l.Println("ID:", id, "Process: <nil>")
		}
	}
}

// clean up the mess
func clearProcessMap(sig os.Signal) {

	// l.Println("processMap:", processMap)

	// range processes
	for id, p := range processMap {
		if p.Proc != nil {

			Log.Info("killing "+id+" PID:", p.Proc.Pid)

			// kill it
			var err error
			if sig == nil {
				err = p.Proc.Kill()
			} else {
				err = p.Proc.Signal(sig)
			}
			if err != nil {
				Log.WithError(err).Error("failed to kill "+id+" PID:", p.Proc.Pid)
			}
		}
	}
}
