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
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// Test main entrypoint
// must be executed prior to other tests
// because it handles command and config parsing
func TestMain(t *testing.T) {

	Log.Info("TEST MAIN")

	// switch to testing mode
	testingMode = true

	// parse tests dir on startup
	zeusDir = "tests"

	// ignore zeusfile in the project dir for now, it will be tested separately with TestZeusfile()
	zeusfilePath = ""

	Convey("When Starting main", t, func(c C) {

		// remove project data rom previous test runs
		os.Remove("tests/zeus_data.json")

		go main()

		time.Sleep(500 * time.Millisecond)

		commandMutex.Lock()
		c.So(len(commands), ShouldBeGreaterThan, 0)
		commandMutex.Unlock()

		// config should be initialized
		configMutex.Lock()
		c.So(conf, ShouldNotBeNil)
		// enable debug mode
		conf.Debug = true
		configMutex.Unlock()

		// manipulate Zeusfile path to not use the ZEUS projects Zeusfile for the tests
		zeusfilePath = "tests/Zeusfile.yml"

		go StartWebListener(false)

		time.Sleep(500 * time.Millisecond)

		// glueServerMutex.Lock()
		// So(glueServer, ShouldNotBeNil)
		// glueServerMutex.Unlock()

		// socketstoreMutex.Lock()
		// So(socketstore, ShouldNotBeNil)
		// socketstoreMutex.Unlock()
	})
}

func TestCommandlineArgs(t *testing.T) {

	var commands = []string{
		"config",
		"help",
		"builtins",
		"format",
		"version",
		"colors",
		"author",
		"makefile",
		"info",
		"clean",
		"data",
	}

	mutex := &sync.Mutex{}

	Convey("Testing commandline args", t, func() {

		for _, cmd := range commands {
			mutex.Lock()
			os.Args = []string{"zeus", cmd}
			mutex.Unlock()
			handleArgs()
		}
	})
}

func TestAliases(t *testing.T) {
	Convey("Testing aliases", t, func(c C) {
		handleLine("alias asdfsdf")
		c.So(projectData.Aliases, ShouldBeEmpty)
		handleLine("alias set testAlias test")
		c.So(len(projectData.Aliases), ShouldEqual, 1)
		c.So(projectData.Aliases["testAlias"], ShouldEqual, "test")
		handleLine("alias remove testAlias")
		c.So(len(projectData.Aliases), ShouldEqual, 0)
		handleLine("alias")
	})
}

func TestConfig(t *testing.T) {
	Convey("Testing config", t, func(c C) {
		handleLine("config asdfasdf")
		handleLine("config set WebInterface true")
		c.So(conf.WebInterface, ShouldBeTrue)
		handleLine("config get WebInterface")
		handleLine("config")
	})
}

func TestCommands(t *testing.T) {
	Convey("Testing commands", t, func() {
		handleLine("help")
		handleLine("help asdfasd")
		handleLine("help test")
	})
}

func TestMilestones(t *testing.T) {
	Convey("Testing milestones", t, func(c C) {
		handleLine("milestones")
		handleLine("milestones asdfasd")
		handleLine("milestones add testMilestone 12-12-2012")
		c.So(projectData.Milestones, ShouldNotBeEmpty)
		handleLine("milestones set testMilestone 50")
		c.So(projectData.Milestones[0].PercentComplete, ShouldEqual, 50)
		handleLine("milestones remove testMilestone")
		c.So(projectData.Milestones, ShouldBeEmpty)
	})
}

func TestDeadlines(t *testing.T) {
	Convey("Testing deadline", t, func(c C) {
		handleLine("deadline")
		c.So(projectData.Deadline, ShouldBeEmpty)
		handleLine("deadline asdfasd")
		handleLine("deadline set 12-12-2012")
		c.So(projectData.Deadline, ShouldEqual, "12-12-2012")
		handleLine("deadline remove")
		c.So(projectData.Deadline, ShouldBeEmpty)
	})
}

func printEvents() {
	for id, e := range projectData.Events {
		Log.Info("ID: " + id + ", Name: " + e.Name + ", Command: " + e.Command)
	}
}

func TestEvents(t *testing.T) {
	Convey("Testing events", t, func(c C) {

		handleLine("events")

		func() {
			projectDataMutex.Lock()
			defer projectDataMutex.Unlock()

			printEvents()

			// there should be only the config and the script or zeusfile watcher event
			c.So(len(projectData.Events), ShouldEqual, 2)
		}()

		handleLine("events asdfasd")

		Log.Info("adding event for tests dir")

		handleLine("events add WRITE tests .xyz error")

		// event creation is async. wait a little bit.
		time.Sleep(100 * time.Millisecond)

		var id string

		func() {
			projectDataMutex.Lock()
			defer projectDataMutex.Unlock()

			printEvents()

			c.So(len(projectData.Events), ShouldEqual, 3)
		}()

		for eID, e := range projectData.Events {
			if e.Path == "tests" {
				id = eID
			}
		}

		Log.Info("removing event for tests dir")

		handleLine("events remove " + id)

		time.Sleep(100 * time.Millisecond)

		func() {
			projectDataMutex.Lock()
			defer projectDataMutex.Unlock()

			c.So(len(projectData.Events), ShouldEqual, 2)
		}()
	})
}

func TestShell(t *testing.T) {
	Convey("Testing the interactive shell", t, func() {
		commands := []string{
			"help",
			"info",
			"format",
			"globals",
			"config",
			"data",
			"version",
			"clear",
			"builtins",
			"clean",
		}

		for _, cmd := range commands {
			Log.Info("testing builtin: ", cmd)
			handleLine(cmd)
		}
	})
}

func TestSanitzer(t *testing.T) {
	Convey("Testing sanitizer", t, func(c C) {
		sanitizeFile("tests/error.sh")

		c.So(sanitizeField("# @zeus-chain: clean -> configure", "zeus-chain"), ShouldEqual, "# @zeus-chain: clean -> configure")
		c.So(sanitizeField("# zeus-chain: clean -> configure", "zeus-chain"), ShouldEqual, "# @zeus-chain: clean -> configure")
		c.So(sanitizeField("# @zeus-chain clean -> configure", "zeus-chain"), ShouldEqual, "# @zeus-chain: clean -> configure")
		c.So(sanitizeField("# zeus-chain clean -> configure", "zeus-chain"), ShouldEqual, "# @zeus-chain: clean -> configure")
	})
}

func TestColors(t *testing.T) {
	Convey("Testing colors", t, func() {
		handleLine("colors")
		handleLine("colors asdfasdf")
		handleLine("colors light")
		handleLine("colors dark")
		handleLine("colors default")
	})
}

func TestCompleters(t *testing.T) {
	Convey("Testing completers", t, func() {
		directoryCompleter("")
		fileCompleter("")
	})
}

func TestMakefileMigration(t *testing.T) {
	Convey("Testing makefile migration", t, func() {
		os.Remove("tests/zeus")
		zeusDir = "tests/zeus"
		migrateMakefile()
		zeusDir = "zeus"
		os.Remove("tests/zeus")
		zeusDir = "tests"
	})
}

func TestAuthorCommand(t *testing.T) {
	Convey("Testing author command", t, func(c C) {
		handleLine("author")
		handleLine("author asdfasdf")
		c.So(projectData.Author, ShouldBeEmpty)
		handleLine("author set Test Author")
		c.So(projectData.Author, ShouldEqual, "Test Author")
		handleLine("author remove")
		c.So(projectData.Author, ShouldBeEmpty)
	})
}

func TestKeybindings(t *testing.T) {
	Convey("Testing keybindings", t, func(c C) {
		handleLine("keys")
		handleLine("keys asdafsdf")
		c.So(projectData.KeyBindings, ShouldBeEmpty)
		handleLine("keys set Ctrl-S git status")
		c.So(projectData.KeyBindings, ShouldNotBeEmpty)
		handleLine("keys set Ctrl-H help")
		c.So(projectData.KeyBindings, ShouldHaveLength, 2)
		handleLine("keys remove Ctrl-H")
		c.So(projectData.KeyBindings, ShouldHaveLength, 1)
		handleLine("keys remove Ctrl-S")
		c.So(projectData.KeyBindings, ShouldBeEmpty)
	})
}

func TestProjectData(t *testing.T) {
	handleLine("data")
}

func TestDependencies(t *testing.T) {
	Convey("Testing Dependencies", t, func(c C) {

		// create bin/dependency1
		handleLine("dependency1")
		_, err := os.Stat("bin/dependency1")
		c.So(err, ShouldBeNil)

		// create bin/dependency2
		handleLine("dependency2")
		_, err = os.Stat("bin/dependency2")
		c.So(err, ShouldBeNil)

	})
}

func TestZeusfile(t *testing.T) {
	Convey("Testing Zeusfile parsing", t, func(c C) {
		err := parseZeusfile("Zeusfile.yml")
		c.So(err, ShouldBeNil)
	})

	var eventID string

	// remove zeusfile watcher
	projectDataMutex.Lock()
	for id, e := range projectData.Events {
		if e.Name == "zeusfile watcher" {
			eventID = id
		}
	}
	projectDataMutex.Unlock()

	removeEvent(eventID)

	// @TODO: test migration & bootstrapping
}

// func TestBootstrapFile(t *testing.T) {
// 	bootstrapCommand()
// }

// func TestBootstrapDir(t *testing.T) {
// 	bootstrapCommand()
// }
