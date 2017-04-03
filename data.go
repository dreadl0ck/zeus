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
	"os"
	"strings"
	"sync"

	yaml "gopkg.in/yaml.v2"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
)

var (
	// path for the project data YAML
	projectDataPath  string
	projectDataMutex sync.Mutex

	// ErrEmptyZeusData occurs when the zeus_data file is empty
	ErrEmptyZeusData = errors.New("zeus data file is empty")
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

	projectDataMutex.Lock()

	b, err := yaml.Marshal(d)
	if err != nil {
		projectDataMutex.Unlock()
		Log.WithError(err).Fatal("failed to marshal zeus data")
	}
	projectDataMutex.Unlock()

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

// parse the project data YAML
func parseProjectData() (*data, error) {

	projectDataPath = zeusDir + "/zeus_data.yml"
	var d = new(data)

	_, err := os.Stat(projectDataPath)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadFile(projectDataPath)
	if err != nil {
		return nil, err
	}

	if len(contents) == 0 {
		return nil, ErrEmptyZeusData
	}

	err = yaml.Unmarshal(contents, d)
	if err != nil {
		Log.WithError(err).Error("failed to unmarshal zeus data - invalid YAML: " + string(contents))
		return nil, err
	}

	return d, nil
}

// load user events from projectData and create the watchers
func loadEvents() {

	projectDataMutex.Lock()
	for _, e := range projectData.Events {

		// reload internal watchers from project data
		if e.Command == "internal" {
			// remove from projectData
			delete(projectData.Events, e.ID)
			reloadEvent(e)
			continue
		}

		Log.WithFields(logrus.Fields{
			"command": e.Command,
		}).Debug("EVENT: ", e)

		fields := strings.Fields(e.Command)

		Log.Info("loading event: ", e.Command, " path: ", e.Path)

		// addEvent will create a new eventID so we need to clean up the entry for the previous one
		delete(projectData.Events, e.ID)

		// copy values from struct
		var (
			path          = e.Path
			op            = e.Op
			command       = e.Command
			name          = e.Name
			fileExtension = e.FileExtension
		)

		go func() {

			err := addEvent(newEvent(path, op, name, fileExtension, "", command, func(event fsnotify.Event) {

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
			}))
			if err != nil {
				Log.Error("failed to watch path: ", path)
			}
		}()
	}
	projectDataMutex.Unlock()

	// event creating is async
	// wait a little bit to avoid duplicate internal events
	time.Sleep(50 * time.Millisecond)
}

// print the current project data as YAML to stdout
func printProjectData() {

	projectDataMutex.Lock()
	defer projectDataMutex.Unlock()

	// make it pretty
	b, err := yaml.Marshal(projectData)
	if err != nil {
		Log.WithError(err).Fatal("failed to marshal zeus project data to YAML")
	}

	l.Println(string(b))
}
