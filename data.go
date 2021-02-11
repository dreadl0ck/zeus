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

	"github.com/fsnotify/fsnotify"
)

var (
	// path for the project data YAML
	projectDataPath string

	// ErrEmptyZeusData occurs when the zeus_data file is empty
	ErrEmptyZeusData = errors.New("zeus data file is empty")
)

// zeus project data written to disk
type data struct {
	fields *dataFields
	sync.RWMutex
}

type dataFields struct {
	BuildNumber int `yaml:"buildNumber"`

	// project deadline
	Deadline string `yaml:"deadline"`

	// project milestones
	Milestones []*milestone `yaml:"milestones"`

	// alias names mapped to commands
	Aliases map[string]string `yaml:"aliases"`

	// mapping from watched path to the corresponding event
	Events map[string]*Event `yaml:"events"`

	Author string `yaml:"author"`

	// keys mapped to commands
	KeyBindings map[string]string `yaml:"keyBindings"`
}

func newData() *data {
	return &data{
		fields: &dataFields{
			BuildNumber: 0,
			Deadline:    "",
			Milestones:  make([]*milestone, 0),
			Aliases:     make(map[string]string, 0),
			Events:      make(map[string]*Event, 0),
			Author:      "",
			KeyBindings: make(map[string]string, 0),
		},
	}
}

// update project data on disk
func (d *data) update() {

	d.Lock()
	defer d.Unlock()

	// marshal data
	b, err := yaml.Marshal(d.fields)
	if err != nil {
		Log.WithError(err).Error("failed to marshal zeus data")
		return
	}

	// get file handle
	f, err := os.OpenFile(projectDataPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		Log.WithError(err).Error("failed to open zeus data")
		return
	}

	// write to file
	_, err = f.Write(append([]byte(asciiArtYAML), b...))
	if err != nil {
		Log.WithError(err).Error("failed to write zeus data")
		return
	}
}

// parse the project data YAML
func parseProjectData() (*data, error) {

	projectDataPath = zeusDir + "/data.yml"

	// init default data
	var d = newData()

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

	err = yaml.Unmarshal(contents, d.fields)
	if err != nil {
		printFileContents(contents)
		Log.WithError(err).Fatal("failed to unmarshal zeus data - invalid YAML:")
		return nil, err
	}

	return d, nil
}

// load user events from projectData and create the watchers
func loadEvents() {

	projectData.Lock()
	for _, e := range projectData.fields.Events {

		// reload internal watchers from project data
		if e.Command == "internal" {
			// remove from projectData
			delete(projectData.fields.Events, e.ID)
			reloadEvent(e)
			continue
		}

		fields := strings.Fields(e.Command)

		Log.Info("loading event: ", e.Command, " path: ", e.Path)

		// addEvent will create a new eventID so we need to clean up the entry for the previous one
		delete(projectData.fields.Events, e.ID)

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
				if cmdChain, ok := validCommandChain(fields, true); ok {
					cmdChain.exec(fields)
				} else {

					Log.Debug("passing command to shell: ", fields)

					// its a shell command
					// ignoring the error here, because stdin, stdout and stderr will be wired up when the command gets executed
					// so the user will become aware of any errors that occur.
					if len(fields) > 1 {
						_ = passCommandToShell(fields[0], fields[1:])
					} else {
						_ = passCommandToShell(fields[0], []string{})
					}
				}
			}))
			if err != nil {
				Log.Error("failed to watch path: ", path)
			}
		}()
	}
	projectData.Unlock()

	// event creation is async
	// wait a little bit to avoid duplicate internal events
	time.Sleep(50 * time.Millisecond)
}

// print the current project data as YAML to stdout
func printProjectData() {

	projectData.Lock()
	defer projectData.Unlock()

	l.Println()

	// make it pretty
	b, err := yaml.Marshal(projectData.fields)
	if err != nil {
		Log.WithError(err).Fatal("failed to marshal zeus project data to YAML")
	}

	l.Println(string(b))
}
