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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// generic formatter type
type formatter struct {
	language *Language

	// path to binary
	binPath string
}

// initialize the formatter to handle shell scripts
func newFormatter(path string, l *Language) *formatter {
	return &formatter{
		language: l,
		binPath:  path,
	}
}

// format a single shell file on disk
func (f *formatter) formatPath(path string) error {

	var cLog = Log.WithField("prefix", "formatPath")
	cLog.Debug("formatting: ", path)
	return nil
}

// walk the zeus directory and run formatPath on all files
func (f *formatter) formatzeusDir() error {

	var cLog = Log.WithField("prefix", "formatzeusDir")

	info, err := os.Stat(scriptDir)
	if err != nil {
		cLog.WithError(err).Error("path does not exist")
		return err
	}
	if !info.IsDir() {
		return errors.New("scriptDir path is not a directory")
	}

	return filepath.Walk(scriptDir, func(path string, info os.FileInfo, err error) error {

		// no recursion for now
		if info.IsDir() {
			return nil
		}

		if err != nil {
			cLog.WithError(err).Error("error walking zeus directory")
			return err
		}

		err = f.formatPath(path)
		if err != nil && !os.IsNotExist(err) {
			cLog.WithError(err).Error("failed to format path: " + path)
			return err
		}
		return nil
	})
}

/*
 *	Utils
 */

// truncate file and seek to the beginning
func empty(f *os.File) error {
	if err := f.Truncate(0); err != nil {
		return err
	}
	_, err := f.Seek(0, 0)
	return err
}

// run the formatter for all files in the zeus dir
// calculates runtime and displays error
func (f *formatter) formatCommand() {

	var (
		start = time.Now()
		err   = f.formatzeusDir()
	)
	if err != nil {
		l.Println("error formatting: ", err)
	}
	l.Println(printPrompt()+"formatted zeus directory in ", time.Now().Sub(start))
}

// watch the zeus dir changes and run format on write event
func (f *formatter) watchScriptDir(eventID string) {

	// dont add a new watcher when the event exists
	projectData.Lock()
	for _, e := range projectData.fields.Events {
		if e.Name == "formatter watcher" {
			projectData.Unlock()
			return
		}
	}
	projectData.Unlock()

	err := addEvent(newEvent(scriptDir, fsnotify.Write, "formatter watcher", "", eventID, "internal", func(event fsnotify.Event) {

		// check if its a valid script
		if strings.HasSuffix(event.Name, ".sh") {

			// ignore further WRITE events while formatting a script
			blockWriteEvent()

			// format script
			err := f.formatPath(event.Name)
			if err != nil {
				Log.WithError(err).Error("failed to format file")
			}
		}
	}))
	if err != nil {
		Log.Error("failed to watch path: ", scriptDir)
	}
}
