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

import "time"

func printDeadlineUsageErr() {
	l.Println(ErrInvalidUsage)
	l.Println("usage: deadline [remove] [set <date>]")
}

// handle deadline shell command
func handleDeadlineCommand(args []string) {

	if len(args) < 2 {
		printDeadline()
		return
	}

	// check sub command
	switch args[1] {
	case "remove":
		removeDeadline()
		return
	case "set":
		if len(args) < 3 {
			printDeadlineUsageErr()
			return
		}
		addDeadline(args[2:])
		return
	default:
		printDeadlineUsageErr()
	}
}

// add a deadline to the projects local configuration
func addDeadline(args []string) {

	if len(args) < 1 {
		printDeadlineUsageErr()
		return
	}

	conf.Lock()
	format := conf.fields.DateFormat
	conf.Unlock()

	// check if date is valid
	t, err := time.Parse(format, args[0])
	if err != nil {
		Log.WithError(err).Error("failed to parse date")
		return
	}

	projectData.Lock()
	projectData.fields.Deadline = t.Format(conf.fields.DateFormat)
	projectData.Unlock()
	projectData.update()
	Log.Info("added deadline for ", args[0])
}

// remove the deadline from project data
func removeDeadline() {
	projectData.Lock()
	projectData.fields.Deadline = ""
	projectData.Unlock()
	projectData.update()
	Log.Info("removed deadline")
}

func printDeadline() {
	if projectData.fields.Deadline != "" {
		l.Println("Deadline: " + cp.Prompt + projectData.fields.Deadline + cp.Text + "\n")
	} else {
		l.Println("no deadline set.")
	}
}
