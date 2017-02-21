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
	ansistrip "c0de/ansistrip"
	"io"
	"os"
	"time"

	"log"
)

var (

	// logging instance
	l = log.New(os.Stdout, "", 0)

	// path to the zeus logfile
	pathLogfile = "zeus/zeus.log"

	// format for TimeStamp in logfiles
	timestampFormat = "[Mon Jan 2 15:04:05 2006]"
)

// initialize logging to a file and to stdout
// returns the logfile handle and an error
func logToFile() (*os.File, error) {

	// open logfile
	f, err := os.OpenFile(pathLogfile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	if err != nil {
		return nil, err
	}

	if conf.LogToFileColor {

		// set logger output to MultiWriter
		l.SetOutput(io.MultiWriter(f, os.Stdout))
	} else {
		// write into strip ansi writer
		l.SetOutput(io.MultiWriter(os.Stdout, ansistrip.New(f)))
	}

	f.WriteString(time.Now().Format(timestampFormat) + "\n")

	return f, nil
}
