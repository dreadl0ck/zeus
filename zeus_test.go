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

	// switch to testing mode
	testingMode = true

	// parse tests dir on startup
	zeusDir = "tests"

	Convey("When Starting main", t, func() {

		go main()

		time.Sleep(500 * time.Millisecond)

		commandMutex.Lock()
		So(len(commands), ShouldBeGreaterThan, 0)
		commandMutex.Unlock()

		configMutex.Lock()
		So(conf, ShouldNotBeNil)
		configMutex.Unlock()

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
	Convey("Testing aliases", t, func() {
		handleLine("alias asdfsdf")
		handleLine("alias set testAlias test")
		handleLine("alias remove testAlias")
		handleLine("alias")
	})
}

func TestConfig(t *testing.T) {
	Convey("Testing config", t, func() {
		handleLine("config asdfasdf")
		handleLine("config set WebInterface true")
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
	Convey("Testing milestones", t, func() {
		handleLine("milestones")
		handleLine("milestones asdfasd")
		handleLine("milestones add testMilestone 12-12-2012")
		handleLine("milestones set testMilestone 50")
		handleLine("milestones remove testMilestone")
	})
}

func TestDeadlines(t *testing.T) {
	Convey("Testing deadline", t, func() {
		handleLine("deadline")
		handleLine("deadline asdfasd")
		handleLine("deadline set 12-12-2012")
		handleLine("deadline remove")
	})
}

func TestEvents(t *testing.T) {
	Convey("Testing events", t, func() {
		handleLine("events")
		handleLine("events asdfasd")

		Log.Info("adding event for tests dir")
		handleLine("events add WRITE tests error")

		// event creation is async. wait a little bit.
		time.Sleep(100 * time.Millisecond)

		Log.Info("removing event for tests dir")
		handleLine("events remove tests")
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
			handleLine(cmd)
		}
	})
}

func TestSanitzer(t *testing.T) {
	Convey("Testing sanitizer", t, func() {
		sanitizeFile("tests/error.sh")
		l.Println(sanitizeField("# @zeus-chain: clean -> configure", "zeus-chain"))
		l.Println(sanitizeField("# zeus-chain: clean -> configure", "zeus-chain"))
		l.Println(sanitizeField("# @zeus-chain clean -> configure", "zeus-chain"))
		l.Println(sanitizeField("# zeus-chain clean -> configure", "zeus-chain"))
	})
}

func TestColors(t *testing.T) {
	Convey("Testing colors", t, func() {
		handleLine("colors")
		handleLine("colors asdfasdf")
		handleLine("colors light")
		handleLine("colors off")
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
	Convey("Testing author command", t, func() {
		handleLine("author")
		handleLine("author asdfasdf")
		handleLine("author set Test Author")
		handleLine("author remove")
	})
}

func TestKeybindings(t *testing.T) {
	Convey("Testing keybindings", t, func() {
		handleLine("keys")
		handleLine("keys asdafsdf")
		handleLine("keys set Ctrl-S git status")
		handleLine("keys set Ctrl-H help")
		handleLine("keys remove Ctrl-H")
	})
}

func TestProjectData(t *testing.T) {
	handleLine("data")
}

func TestBootstrap(t *testing.T) {
	// bootstrapCommand()
}

func TestParser(t *testing.T) {

}

func TestDependencies(t *testing.T) {
	Convey("Testing Dependencies", t, func() {

		// create bin/dependency1
		handleLine("dependency1")
		_, err := os.Stat("bin/dependency1")
		So(err, ShouldBeNil)

		// create bin/dependency2
		handleLine("dependency2")
		_, err = os.Stat("bin/dependency2")
		So(err, ShouldBeNil)

	})
}
