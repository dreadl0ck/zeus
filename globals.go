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
	"io/ioutil"
	"strconv"
	"sync"
)

type globals struct {

	// mapped variable names to values
	Vars map[string]string

	sync.RWMutex
}

// print the contents of all globals on stdout
func listGlobals() {

	g.Lock()
	defer g.Unlock()

	if len(g.Vars) > 0 {

		w := 20

		l.Println("\n" + cp.Prompt + pad("name", w) + "value")
		for name, val := range g.Vars {
			l.Println(cp.Text+pad(name, w), val)
		}

		ls.Lock()
		defer ls.Unlock()
		for name, lang := range ls.items {
			code, err := ioutil.ReadFile(zeusDir + "/globals/globals" + lang.FileExtension)
			if err == nil {
				l.Println("\n" + cp.Prompt + name)
				l.Println(cp.Text + string(code))
			}
		}
	} else {
		l.Println("no globals defined.")
	}
}

// generate global variables for a given language
// returns a string
func generateGlobals(lang *Language) (out string) {

	g.Lock()
	defer g.Unlock()

	// initialize global variables
	for name, value := range g.Vars {

		var valString = true
		// check if its a boolean
		if _, err := strconv.ParseBool(value); err == nil {
			// value goot as it is
			valString = false
		}
		// check if its an integer
		if _, err := strconv.ParseInt(value, 10, 0); err == nil {
			// value good as it is
			valString = false
		}
		// check if its a string
		if valString {
			value = "\"" + value + "\""
		}

		out += lang.VariableKeyword + name + lang.AssignmentOperator + value + lang.LineDelimiter + "\n"
	}

	return
}
