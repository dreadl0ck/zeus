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
	"strconv"
	"strings"
	"time"
)

var (
	// german date format
	dateFormat = "02-01-2006"
)

// milestone represents a project milestone
type milestone struct {
	Name            string
	Date            time.Time
	Description     string
	PercentComplete int
}

// create a new milestone instance
func newMilestone(name string, date time.Time, description []string) *milestone {
	return &milestone{
		Name:            name,
		Date:            date,
		Description:     strings.Join(description, " "),
		PercentComplete: 0,
	}
}

func printMilestoneUsageErr() {
	Log.Error(ErrInvalidUsage)
	Log.Info("usage: milestones [remove <name>] [set <name> <0-100>] [add <name> <date> [description]]")
}

// handle milestones shell command
func handleMilestonesCommand(args []string) {

	if len(args) < 2 {
		listMilestones()
		return
	}

	// check sub command
	switch args[1] {
	case "remove":
		if len(args) < 3 {
			printMilestoneUsageErr()
			return
		}
		removeMilestone(args[2])
		return
	case "set":
		if len(args) < 4 {
			printMilestoneUsageErr()
			return
		}
		setMilestone(args[2], args[3])
		return
	case "add":
		if len(args) < 3 {
			printMilestoneUsageErr()
			return
		}
		addMilestone(args[2:])
		return
	default:
		printMilestoneUsageErr()
	}
}

// create the status bar from the status integer
func getStatusBar(p int) string {

	res := "["
	for c := 1; c <= 10; c++ {

		if p < c*10 {
			res += "  "
		} else {
			res += "=="
		}
	}
	return res + "] " + strconv.Itoa(p) + "%"
}

// add a milestone to the project
func addMilestone(args []string) {

	if len(args) < 2 {
		printMilestoneUsageErr()
		return
	}

	// check if date is valid
	t, err := time.Parse(dateFormat, args[1])
	if err != nil {
		Log.WithError(err).Error("failed to parse date")
		return
	}

	// create milestone
	var m *milestone

	// check if theres a description
	if len(args) >= 3 {
		m = newMilestone(args[0], t, args[2:])
	} else {
		m = newMilestone(args[0], t, []string{})
	}

	projectData.Milestones = append(projectData.Milestones, m)
	projectData.update()
	Log.Info("added milestone ", args[0])
}

// update a milestones status
// valid values are 0-100
func setMilestone(name, percent string) {

	p, err := strconv.ParseInt(percent, 10, 0)
	if err != nil || p > 100 || p < 0 {
		printMilestoneUsageErr()
		return
	}

	var ok bool

	for i := range projectData.Milestones {
		if projectData.Milestones[i].Name == name {
			projectData.Milestones[i].PercentComplete = int(p)
			ok = true
		}
	}

	if !ok {
		Log.Info("unknown milestone: ", name)
		return
	}

	projectData.update()
}

// remove a milestone from project data
func removeMilestone(name string) {

	if name == "" {
		Log.Error("no name supplied")
		return
	}

	for i, m := range projectData.Milestones {
		if m.Name == name {
			projectData.Milestones = append(projectData.Milestones[:i], projectData.Milestones[i+1:]...)
			projectData.update()
			Log.Info("remove milestone ", name)
			return
		}
	}
}

// print all milestones to stdout
func listMilestones() {
	if len(projectData.Milestones) > 0 {

		l.Println(cp.colorText + "Milestones:")
		for i, m := range projectData.Milestones {
			if len(m.Description) > 0 {
				l.Println("#", i, getStatusBar(m.PercentComplete), "name:", m.Name, "date:", cp.colorPrompt+m.Date.Format(dateFormat)+cp.colorText, "description:", m.Description)
			} else {
				l.Println("#", i, getStatusBar(m.PercentComplete), "name:", m.Name, "date:", cp.colorPrompt+m.Date.Format(dateFormat)+cp.colorText)
			}
		}
		l.Println("")
	} else {
		l.Println("no milestones set.")
		l.Println("")
	}
}
