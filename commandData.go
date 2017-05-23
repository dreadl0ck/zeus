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
	"strconv"
	"strings"
)

// command header
// used in scripts to supply information for ZEUS
type commandData struct {

	// one line Description text
	Description string `yaml:"description"`

	// scripting language of the command
	Language string `yaml:"language"`

	// Help page text
	Help string `yaml:"help"`

	// Arguments
	Arguments []string `yaml:"arguments"`

	// Dependencies
	Dependencies string `yaml:"dependencies"`

	// ouptuts
	Outputs []string `yaml:"ouputs"`

	// increase buildnumber on each execution
	BuildNumber bool `yaml:"buildNumber"`

	// execute command in a detached screen session
	Async bool `yaml:"async"`

	// Exec is the script to run when executed
	Exec string `yaml:"exec"`
}

// perform validation on the script header
// checks for duplicate and unknown header fields
func validateHeader(c []byte, path string) error {

	var (
		fields       = []string{"description", "help", "language", "arguments", "dependencies", "outputs", "buildNumber", "async", "exec"}
		parsedFields []string
		foundField   bool
	)

	for i, line := range strings.Split(string(c), "\n") {

		field := yamlField.FindString(line)
		if field != "" && !strings.HasPrefix(field, "    ") {
			field = strings.TrimSuffix(strings.TrimSpace(field), ":")
			for _, item := range fields {
				if field == strings.TrimSpace(string(item)) {
					for _, f := range parsedFields {
						if f == field {
							printScript(string(c), path)
							return errors.New("line " + strconv.Itoa(i) + ": duplicate header field: " + field)
						}
					}
					parsedFields = append(parsedFields, field)
					foundField = true
				}
			}
			if !foundField {
				printScript(string(c), path)
				return errors.New("line " + strconv.Itoa(i) + ": unknown header field: " + field)
			}
			foundField = false
		}
	}

	return nil
}
