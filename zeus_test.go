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
	"syscall"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	// make sure backend is only started once
	running bool
)

// Test main entrypoint
// must be executed prior to other tests
// because it handles command and config parsing
func TestMain(t *testing.T) {

	if !running {

		running = true

		// switch to testing mode
		testingMode = true

		// parse tests dir on startup
		zeusDir = "tests/zeus"
		scriptDir = "tests/zeus/scripts"

		// ignore commandsFile in the project dir for now, it will be tested separately with TestCommandsFile()
		// commandsFilePath = ""
		// manipulate CommandsFile path to not use the ZEUS projects CommandsFile for the tests
		commandsFilePath = "tests/zeus/commands.yml"

		Convey("When Starting main", t, func(c C) {

			// remove project data from previous test runs
			os.Remove("tests/zeus/data.yml")

			go main()

			time.Sleep(500 * time.Millisecond)

			func() {
				cmdMap.Lock()
				defer cmdMap.Unlock()
				c.So(len(cmdMap.items), ShouldBeGreaterThan, 0)
			}()

			// config should be initialized
			func() {
				conf.Lock()
				defer conf.Unlock()
				c.So(conf, ShouldNotBeNil)
				// enable debug mode
				conf.fields.Debug = true
			}()

			go StartWebListener(false)

			time.Sleep(500 * time.Millisecond)
		})
	}
}

func TestCommandlineArgs(t *testing.T) {

	TestMain(t)

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

	TestMain(t)

	Convey("Testing aliases", t, func(c C) {
		handleLine("alias asdfsdf")
		c.So(projectData.fields.Aliases, ShouldBeEmpty)
		handleLine("alias set testAlias test")
		c.So(len(projectData.fields.Aliases), ShouldEqual, 1)
		c.So(projectData.fields.Aliases["testAlias"], ShouldEqual, "test")
		handleLine("alias remove testAlias")
		c.So(len(projectData.fields.Aliases), ShouldEqual, 0)
		handleLine("alias")
	})
}

func TestConfig(t *testing.T) {

	TestMain(t)

	Convey("Testing config", t, func(c C) {
		handleLine("config asdfasdf")
		handleLine("config set WebInterface true")
		c.So(conf.fields.WebInterface, ShouldBeTrue)
		handleLine("config get WebInterface")
		handleLine("config")
	})
}

func TestCommands(t *testing.T) {

	TestMain(t)

	Convey("Testing commands", t, func() {
		handleLine("help")
		handleLine("help asdfasd")
		handleLine("help test")
	})
}

func TestMilestones(t *testing.T) {

	TestMain(t)

	Convey("Testing milestones", t, func(c C) {
		handleLine("milestones")
		handleLine("milestones asdfasd")
		handleLine("milestones add testMilestone 12-12-2012")
		c.So(projectData.fields.Milestones, ShouldNotBeEmpty)
		handleLine("milestones set testMilestone 50")
		c.So(projectData.fields.Milestones[0].PercentComplete, ShouldEqual, 50)
		handleLine("milestones remove testMilestone")
		c.So(projectData.fields.Milestones, ShouldBeEmpty)
	})
}

func TestLanguages(t *testing.T) {

	TestMain(t)

	Convey("Testing multiple languages", t, func(c C) {
		handleLine("python src='asdf' dst='fdsa'")
		handleLine("ruby src='asdf' dst='fdsa'")
		// handleLine("lua src='asdf' dst='fdsa'")
		// handleLine("javascript src='asdf' dst='fdsa'")
	})
}

func TestDeadlines(t *testing.T) {

	TestMain(t)

	Convey("Testing deadline", t, func(c C) {
		handleLine("deadline")
		c.So(projectData.fields.Deadline, ShouldBeEmpty)
		handleLine("deadline asdfasd")
		handleLine("deadline set 12-12-2012")
		c.So(projectData.fields.Deadline, ShouldEqual, "12-12-2012")
		handleLine("deadline remove")
		c.So(projectData.fields.Deadline, ShouldBeEmpty)
	})
}

func TestEvents(t *testing.T) {

	TestMain(t)

	Convey("Testing events", t, func(c C) {

		// print events
		handleLine("events")

		// check number of events
		func() {
			projectData.Lock()
			defer projectData.Unlock()

			printEvents()

			// there should be only the config watcher event
			c.So(len(projectData.fields.Events), ShouldEqual, 2)
		}()

		handleLine("events asdfasd")

		Log.Info("adding event for tests dir")
		handleLine("events add WRITE tests .xyz error")

		// event creation is async. wait a little bit.
		time.Sleep(100 * time.Millisecond)

		// check number of events
		func() {
			projectData.Lock()
			defer projectData.Unlock()

			c.So(len(projectData.fields.Events), ShouldEqual, 3)
		}()

		projectData.Lock()

		var id string
		for eID, e := range projectData.fields.Events {
			if e.Path == "tests" {
				id = eID
			}
		}
		projectData.Unlock()

		Log.Info("removing event for tests dir")

		handleLine("events remove " + id)

		time.Sleep(100 * time.Millisecond)

		func() {
			projectData.Lock()
			defer projectData.Unlock()

			c.So(len(projectData.fields.Events), ShouldEqual, 2)
		}()
	})
}

func TestShell(t *testing.T) {

	TestMain(t)

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
			"edit",
			"todo",
			"git-filter",
			"procs",
			"asdfasdfasdfasdfasdfasdfas",
			"generate",
		}

		// execute builtins without parameters
		for _, cmd := range commands {
			Log.Info("testing builtin: ", cmd)
			handleLine(cmd)
		}
	})
}

func TestColors(t *testing.T) {

	TestMain(t)

	Convey("Testing colors", t, func() {

		// print colors
		handleLine("colors")

		// invalid input
		handleLine("colors asdfasdf")

		// switch some profiles
		handleLine("colors light")
		handleLine("colors dark")
		handleLine("colors default")
	})
}

func TestMakefileMigration(t *testing.T) {

	TestMain(t)

	Convey("Testing makefile migration", t, func() {

		// remove previous generated directory
		// os.Remove("tests/zeus/migration-test")

		// migrate test Makefile into tests/zeus
		migrateMakefile("tests/zeus/migration-test")

		// clean up
		// os.Remove("tests/zeus/migration-test")
	})
}

func TestAuthorCommand(t *testing.T) {

	TestMain(t)

	Convey("Testing author command", t, func(c C) {

		// print author
		handleLine("author")

		// invalid input
		handleLine("author asdfasdf")
		c.So(projectData.fields.Author, ShouldBeEmpty)

		// set a new author
		handleLine("author set Test Author")
		c.So(projectData.fields.Author, ShouldEqual, "Test Author")

		// remove author
		handleLine("author remove")
		c.So(projectData.fields.Author, ShouldBeEmpty)
	})
}

func TestKeybindings(t *testing.T) {

	TestMain(t)

	Convey("Testing keybindings", t, func(c C) {

		// print keybindings
		handleLine("keys")

		// invalid input
		handleLine("keys asdafsdf")
		c.So(projectData.fields.KeyBindings, ShouldBeEmpty)

		// add keybinding
		handleLine("keys set Ctrl-S git status")
		c.So(projectData.fields.KeyBindings, ShouldNotBeEmpty)

		// add a second keybinding
		handleLine("keys set Ctrl-H help")
		c.So(projectData.fields.KeyBindings, ShouldHaveLength, 2)

		// remove one
		handleLine("keys remove Ctrl-H")
		c.So(projectData.fields.KeyBindings, ShouldHaveLength, 1)

		// remove the other
		handleLine("keys remove Ctrl-S")
		c.So(projectData.fields.KeyBindings, ShouldBeEmpty)
	})
}

func TestProjectData(t *testing.T) {

	TestMain(t)

	// print project data
	handleLine("data")
}

func TestDependencies(t *testing.T) {

	TestMain(t)

	Convey("Testing Dependencies", t, func(c C) {

		// create tests/bin/dependency1
		handleLine("dependency1")
		_, err := os.Stat("tests/bin/dependency1")
		c.So(err, ShouldBeNil)

		// remove dependency1
		os.Remove("tests/bin/dependency1")

		// create tests/bin/dependency2
		handleLine("dependency2")
		_, err = os.Stat("tests/bin/dependency2")
		c.So(err, ShouldBeNil)

		// dependency1 should have been created
		_, err = os.Stat("tests/bin/dependency1")
		c.So(err, ShouldBeNil)

		// clean up
		os.Remove("tests/bin/dependency1")
		os.Remove("tests/bin/dependency2")
	})
}

func TestCommandsFile(t *testing.T) {

	TestMain(t)

	Convey("Testing CommandsFile parsing", t, func(c C) {

		// parse ZEUS project CommandsFile
		err := parseCommandsFile("zeus/commands.yml")
		c.So(err, ShouldBeNil)

		// event creation is async, wait a little bit
		time.Sleep(100 * time.Millisecond)

		// get commandsFile watcher eventID
		var eventID string

		projectData.Lock()
		for id, e := range projectData.fields.Events {
			if e.Name == "commandsFile watcher" {
				eventID = id
			}
		}
		projectData.Unlock()

		// event must exist
		c.So(eventID, ShouldNotBeEmpty)

		// clean up
		removeEvent(eventID)
	})
}

// func TestBootstrap(t *testing.T) {

// 	TestMain(t)

// 	Convey("Testing zeus bootstrapping", t, func(c C) {

// 		// make sure zeus dir does not exist
// 		os.Remove("tests/zeus/bootstrap-test")
// 	})
// }

func TestGenerate(t *testing.T) {

	TestMain(t)

	Convey("Testing standalone script generation", t, func(c C) {

		handleLine("generate")
		handleLine("generate build.sh build")
		handleLine("generate testChain.sh async -> optional bla=asdf req=asdfd -> error")

		os.Remove("tests/zeus/generated")
	})
}

// func TestCommandsFileMigration(t *testing.T) {

// 	TestMain(t)

// 	Convey("Testing CommandsFile to zeusDir migration", t, func(c C) {

// 		os.Remove("tests/zeus-migration-test")
// 		os.Remove("tests/CommandsFile.yml")
// 		os.Remove("tests/CommandsFile_old.yml")

// 		c.So(exec.Command("cp", "zeus/CommandsFile.yml", "tests/CommandsFile.yml").Run(), ShouldBeNil)

// 		zeusDir = "tests/zeus/zeus-migration-test"

// 		//c.So(migrateCommandsFile(), ShouldBeNil)

// 		zeusDir = "tests"

// 		// clean up
// 		os.Remove("tests/zeus/zeus-migration-test")
// 	})
// }

func TestProcesses(t *testing.T) {

	TestMain(t)

	Convey("Testing process handling", t, func(c C) {

		printProcsCommandUsageErr()

		// spawn async command
		handleLine("async")

		// kill it by passing SIGINT
		passSignalToProcs(syscall.SIGINT)

		// spawn async command again
		handleLine("async")

		// flush process map
		clearProcessMap()
	})
}

func TestCustomCompleters(t *testing.T) {

	TestMain(t)

	Convey("Testing custom completers", t, func(c C) {

		// test completion of eventIDs for removing events
		c.So(eventIDCompleter(""), ShouldNotBeEmpty)

		// test completion of available commands
		c.So(commandCompleter(""), ShouldNotBeEmpty)

		// complete available parser languages
		c.So(languageCompleter(""), ShouldNotBeEmpty)

		// complete available commands for chains
		//c.So(commandChainCompleter("d"), ShouldNotBeEmpty)

		c.So(colorProfileCompleter(""), ShouldNotBeEmpty)

		c.So(todoIndexCompleter(""), ShouldBeEmpty)

		// complete PIDs for killing processes
		// c.So(pIDCompleter(""), ShouldNotBeEmpty)

		// complete available filetypes for the event target directory
		c.So(fileTypeCompleter("events add WRITE tests/zeus/scripts"), ShouldNotBeEmpty)

		c.So(directoryCompleter(""), ShouldNotBeEmpty)
	})
}
