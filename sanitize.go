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
	"bytes"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
)

// santitize the file at path
// handles typos in zeus header fields and adds missing shebang
func sanitizeFile(path string) {

	var (
		buffer bytes.Buffer
		cLog   = Log.WithFields(logrus.Fields{
			"prefix": "sanitizeFile",
			"path":   path,
		})
	)

	// read file
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		cLog.WithError(err).Fatal("failed to read file")
	}

	// split line by line
	for c, line := range strings.Split(string(contents), "\n") {

		if c == 0 {
			if line != p.bang {
				cLog.Info("adding missing bang")
				buffer.WriteString(p.bang + "\n")
				continue
			}
		}

		// check if its a comment
		if strings.HasPrefix(line, "#") {

			switch true {
			case strings.Contains(line, p.zeusFieldChain):
				line = sanitizeField(line, p.zeusFieldChain)
			case strings.Contains(line, p.zeusFieldHelp):
				line = sanitizeField(line, p.zeusFieldHelp)
			case strings.Contains(line, p.zeusFieldArgs):
				line = sanitizeField(line, p.zeusFieldArgs)
			default: // normal comment
				break
			}
			cLog.Debug("sanitizing line with: ", line)
			buffer.WriteString(line + "\n")
		} else {
			buffer.WriteString(line + "\n")
		}
	}

	// open file for writing and truncate
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0700)
	if err != nil {
		cLog.WithError(err).Fatal("failed to open file")
	}
	defer f.Close()

	// write buffer
	f.Write(buffer.Bytes())
}

// sanitize the given header field and correct typos
// multiple or no occurrence of @, # and : should be handled
func sanitizeField(line, field string) string {
	return "# @" + field + ": " + string(invalidzeusHeaderField.ReplaceAll([]byte(line), []byte("")))
}
