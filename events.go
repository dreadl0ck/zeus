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
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
)

var (
	// ignore the the write event when updating the config using the shell command
	disableWriteEvent      = false
	disableWriteEventMutex = &sync.Mutex{}

	// ErrInvalidEventType means the given event type string is invalid
	ErrInvalidEventType = errors.New("invalid fsnotify event type. available types are: WRITE | REMOVE | RENAME | CHMOD")

	// ErrInvalidUsage means the command was used incorrectly
	ErrInvalidUsage = errors.New("invalid usage")
)

// Event represents a watched path, along with an an action
// that will be performed when an operation of the specified type occurs
type Event struct {

	// Event Name
	Name string

	// Event ID
	ID string

	// Path to watch
	Path string

	// Operation type
	Op fsnotify.Op

	// Command to be executed upon event
	Command string

	// custom event handler func
	handler func(fsnotify.Event)

	// shutdown chan
	stopChan chan bool
}

func printEventsUsageErr() {
	Log.Error(ErrInvalidUsage)
	Log.Info("usage: events [add <type> <path> <commandChain>] [remove <path>]")
}

// handle events command
func handleEventsCommand(args []string) {

	if len(args) < 2 {
		listEvents()
		return
	}

	if len(args) < 3 {
		printEventsUsageErr()
		return
	}

	switch args[1] {
	case "remove":
		removeEvent(args[2])
	case "add":

		if len(args) < 5 {
			printEventsUsageErr()
			return
		}

		// check if event type is valid
		op, err := getEventType(args[2])
		if err != nil {
			Log.Error(err)
			return
		}

		// check if path exists
		_, err = os.Stat(args[3])
		if err != nil {
			Log.Error(err)
			return
		}

		chain := strings.Join(args[4:], " ")
		if validCommandChain(args[4:]) {
			Log.Info("adding command chain")
		} else {
			Log.Info("adding shell command")
		}

		fields := args[4:]

		go func() {
			err := addEvent(args[3], op, func(event fsnotify.Event) {

				Log.Debug("event fired, name: ", event.Name, " path: ", args[3])

				if validCommandChain(fields) {
					executeCommandChain(chain)
				} else {

					Log.Debug("passing chain to shell")

					// its a shell command
					if len(fields) > 1 {
						passCommandToShell(fields[0], fields[1:])
					} else {
						passCommandToShell(fields[0], []string{})
					}
				}

			}, "custom event", chain)
			if err != nil {
				Log.Error("failed to watch path: ", args[3])
			}

		}()

	default:
		printEventsUsageErr()
	}
}

// parse command type string and fsnotify type
func getEventType(event string) (fsnotify.Op, error) {

	switch event {
	case "WRITE":
		return fsnotify.Write, nil
	case "REMOVE":
		return fsnotify.Remove, nil
	case "RENAME":
		return fsnotify.Rename, nil
	case "CHMOD":
		return fsnotify.Chmod, nil
	default:
		return 0, ErrInvalidEventType
	}
}

// list all currently registered events
func listEvents() {

	w := 20

	l.Println(cp.colorPrompt + pad("name", w) + pad("ID", w) + pad("operation", w) + pad("command", w) + pad("path", w))
	for _, e := range projectData.Events {
		l.Println(cp.colorText + pad(e.Name, w) + pad(e.ID, w) + pad(e.Op.String(), w) + pad(e.Command, w) + pad(e.Path, w))
	}
}

// remove the event for the given path
func removeEvent(id string) {

	eventLock.Lock()
	defer eventLock.Unlock()

	// check if event exists
	if e, ok := projectData.Events[id]; ok {

		Log.Info("stopping handler")

		if e.stopChan != nil {
			// stop event handler
			e.stopChan <- true
		}

		Log.Info("deleting event")

		// delete event
		delete(projectData.Events, id)

		Log.Info("removed event with name ", e.Name)

		// update project data
		projectData.update()
		return
	}

	Log.Error("event with ID ", id, " does not exist")
}

// addEvent adds a watcher for path and register a handler that will fire if operation op occurs
// the command parameter contains the associated shell command / buildChain for user defined events
func addEvent(path string, op fsnotify.Op, handler func(fsnotify.Event), name, command string) error {

	var (
		cLog = Log.WithField("prefix", "addEvent")

		// create event
		e = &Event{
			Path:     path,
			Name:     name,
			ID:       randomString(),
			Op:       op,
			handler:  handler,
			stopChan: make(chan bool, 1),
			Command:  command,
		}
	)

	Log.WithField("path", path).Warn("############ adding event")

	// add to events
	eventLock.Lock()
	projectData.Events[e.ID] = e
	projectData.update()
	eventLock.Unlock()

	// init new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// listen for events
	done := make(chan bool)
	go func() {

		for {
			select {
			case event := <-watcher.Events:

				// cLog.WithFields(logrus.Fields{
				// 	"event": event,
				// 	"path":  path,
				// }).Debug("incoming event")

				// check operation type
				if event.Op == op {

					// check if write event was disabled.
					// example: when updating the config with the config command
					// revalidating the config is not necessary
					disableWriteEventMutex.Lock()
					if disableWriteEvent {
						disableWriteEvent = false
						disableWriteEventMutex.Unlock()
						cLog.Debug("ignoring WRITE event for path: ", path)
						continue
					}
					disableWriteEventMutex.Unlock()

					// fire handler
					handler(event)
				}
			case err := <-watcher.Errors:
				cLog.WithError(err).Fatal("watcher failed")
			case _ = <-e.stopChan:
				watcher.Close()
				done <- true
				return
			}
		}
	}()

	// add path to watcher
	err = watcher.Add(path)
	if err != nil {
		cLog.WithFields(logrus.Fields{
			"error": err,
			"path":  path,
		}).Error("failed to add path to watcher")
		e.stopChan <- true
		return err
	}

	// wait for it
	<-done

	return nil
}
