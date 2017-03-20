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
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
)

var (
	// path for the project data JSON
	projectDataPath string
)

// zeus project data written to disk
type data struct {
	BuildNumber int

	// project deadline
	Deadline string

	// project milestones
	Milestones []*milestone

	// alias names mapped to commands
	Aliases map[string]string

	// mapping from watched path to the corresponding event
	Events map[string]*Event

	Author string

	// keys mapped to commands
	KeyBindings map[string]string
}

func newData() *data {
	return &data{
		BuildNumber: 0,
		Deadline:    "",
		Milestones:  make([]*milestone, 0),
		Aliases:     make(map[string]string, 0),
		Events:      make(map[string]*Event, 0),
		Author:      "",
		KeyBindings: make(map[string]string, 0),
	}
}

// update project data on disk
func (d *data) update() {

	// make it pretty
	b, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		Log.WithError(err).Fatal("failed to marshal zeus data")
	}

	f, err := os.OpenFile(projectDataPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		Log.WithError(err).Fatal("failed to open zeus data")
	}

	disableWriteEventMutex.Lock()
	disableWriteEvent = true
	disableWriteEventMutex.Unlock()

	_, err = f.Write(b)
	if err != nil {
		Log.WithError(err).Fatal("failed to write zeus data")
	}

	disableWriteEventMutex.Lock()
	disableWriteEvent = false
	disableWriteEventMutex.Unlock()
}

// parse the project data JSON
func parseProjectData() (*data, error) {

	projectDataPath = zeusDir + "/zeus_data.json"
	var d = new(data)

	_, err := os.Stat(projectDataPath)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadFile(projectDataPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, d)
	if err != nil {
		Log.WithError(err).Error("failed to unmarshal zeus data - invalid JSON")
		return nil, err
	}

	return d, nil
}

// load user events from projectData and create the watchers
func loadEvents() {

	eventLock.Lock()
	for _, e := range projectData.Events {

		// skip loading of internal watchers from project data
		if e.Path == zeusDir || e.Path == projectConfigPath {
			delete(projectData.Events, e.ID)
			continue
		}

		Log.WithFields(logrus.Fields{
			"command": e.Command,
		}).Debug("EVENT: ", e)

		fields := strings.Fields(e.Command)

		Log.Warn("LOADING EVENT: ", e.Command, " path: ", e.Path)

		// addEvent will create a new eventID so we need to clean up the entry for the previous one
		delete(projectData.Events, e.ID)

		// copy values from struct
		var (
			path    = e.Path
			op      = e.Op
			command = e.Command
		)

		go func() {

			err := addEvent(path, op, func(event fsnotify.Event) {

				Log.Debug("event fired, name: ", event.Name, " path: ", path)

				// validate commandChain
				if validCommandChain(fields, false) {
					executeCommandChain(command)
				} else {

					Log.Debug("passing chain to shell")

					// its a shell command
					if len(fields) > 1 {
						passCommandToShell(fields[0], fields[1:])
					} else {
						passCommandToShell(fields[0], []string{})
					}
				}

			}, e.Name, e.FileExtension, command)
			if err != nil {
				Log.Error("failed to watch path: ", path)
			}
		}()
	}
	eventLock.Unlock()
}
