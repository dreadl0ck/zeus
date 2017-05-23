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
	"sync"

	"gopkg.in/yaml.v2"
)

var globalsFromZeusfile bool

type globals struct {

	// mapped variable names to values
	Vars map[string]string `yaml:"variables"`

	sync.RWMutex
}

// check for globals script
func parseGlobals() {

	if !globalsFromZeusfile {
		globalsContent, err := ioutil.ReadFile(zeusDir + "/globals.yml")
		if err != nil {
			Log.WithError(err).Error("failed to read globals")
			return
		}

		g.Lock()
		defer g.Unlock()

		err = yaml.Unmarshal(globalsContent, g)
		if err != nil {
			Log.WithError(err).Error("failed to unmarshal globals")
			return
		}
	}
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

		ps.Lock()
		defer ps.Unlock()
		for name, p := range ps.items {
			code, err := ioutil.ReadFile(zeusDir + "/globals" + p.language.FileExtension)
			if err == nil {
				l.Println("\n" + cp.Prompt + name)
				l.Println(cp.Text + string(code))
			}
		}
	} else {
		l.Println("no globals defined.")
	}
}
