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
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/chzyer/readline"
)

var (
	// regex to the match a UNIX path
	shellPath = regexp.MustCompile("(([a-z]*[A-Z]*[0-9]*(_|-)*)*/*)*")

	// regex to match a command with a trailing UNIX path
	shellCommandWithPath = regexp.MustCompile("([a-z]*\\s*)*(([a-z]*[A-Z]*[0-9]*(_|-)*)*/*)*")
)

// assemble and return all items for config item completion
func configItems() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("MakefileOverview", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("Report", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("AutoFormat", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("FixParseErrors", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("Colors", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("PassCommandsToShell", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("WebInterface", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("Interactive", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("LogToFileColor", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("LogToFile", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("Debug", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("RecursionDepth"),
		readline.PcItem("ProjectNamePrompt", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("AllowUntypedArgs", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("ColorProfile"),
		readline.PcItem("HistoryFile", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("HistoryLimit"),
		readline.PcItem("ExitOnInterrupt", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("DisableTimestamps", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("PrintBuiltins", readline.PcItem("true"), readline.PcItem("false")),
	}
}

// assemble and return all items for keycomb item completion
func keyKombItems() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("Ctrl-A"),
		readline.PcItem("Ctrl-B"),
		readline.PcItem("Ctrl-E"),
		readline.PcItem("Ctrl-F"),
		readline.PcItem("Ctrl-G"),
		readline.PcItem("Ctrl-H"),
		readline.PcItem("Ctrl-I"),
		readline.PcItem("Ctrl-J"),
		readline.PcItem("Ctrl-K"),
		readline.PcItem("Ctrl-L"),
		readline.PcItem("Ctrl-M"),
		readline.PcItem("Ctrl-N"),
		readline.PcItem("Ctrl-O"),
		readline.PcItem("Ctrl-P"),
		readline.PcItem("Ctrl-Q"),
		readline.PcItem("Ctrl-R"),
		readline.PcItem("Ctrl-S"),
		readline.PcItem("Ctrl-T"),
		readline.PcItem("Ctrl-U"),
		readline.PcItem("Ctrl-V"),
		readline.PcItem("Ctrl-W"),
		readline.PcItem("Ctrl-X"),
		readline.PcItem("Ctrl-Y"),
	}
}

// return a new default completer instance
func newCompleter() *readline.PrefixCompleter {
	c := readline.NewPrefixCompleter(
		readline.PcItem("git",
			readline.PcItem("add"),
			readline.PcItem("status"),
			readline.PcItem("commit"),
		),
		readline.PcItem("exit"),
		readline.PcItem("help"),
		readline.PcItem("info"),
		readline.PcItem("clear"),
		readline.PcItem("format"),
		readline.PcItem("globals"),
		readline.PcItem("version"),
		readline.PcItem("config",
			readline.PcItem("set",
				configItems()...,
			),
			readline.PcItem("get",
				configItems()...,
			),
		),
		readline.PcItem("events",
			readline.PcItem("add",
				readline.PcItem("WRITE",
					readline.PcItemDynamic(fileCompleter),
				),
				readline.PcItem("REMOVE",
					readline.PcItemDynamic(fileCompleter),
				),
				readline.PcItem("CHMOD",
					readline.PcItemDynamic(fileCompleter),
				),
				readline.PcItem("RENAME",
					readline.PcItemDynamic(fileCompleter),
				),
			),
			readline.PcItem("remove",
				readline.PcItemDynamic(fileCompleter),
			),
		),
		readline.PcItem("milestones",
			readline.PcItem("set"),
			readline.PcItem("remove"),
			readline.PcItem("add"),
		),
		readline.PcItem("deadline",
			readline.PcItem("set"),
			readline.PcItem("remove"),
		),
		readline.PcItem("makefile",
			readline.PcItem("migrate"),
		),
		readline.PcItem("data"),
		readline.PcItem("alias",
			readline.PcItem("set"),
			readline.PcItem("remove"),
		),
		readline.PcItem("colors",
			readline.PcItem("dark"),
			readline.PcItem("light"),
			readline.PcItem("default"),
		),
		readline.PcItem("author",
			readline.PcItem("set"),
			readline.PcItem("remove"),
		),
		readline.PcItem("builtins"),
		readline.PcItem("keys",
			readline.PcItem("set",
				keyKombItems()...,
			),
			readline.PcItem("remove",
				keyKombItems()...,
			),
		),
		// shell commands that need file/dir completion
		readline.PcItem("ls",
			readline.PcItemDynamic(directoryCompleter),
		),
		readline.PcItem("cat",
			readline.PcItemDynamic(fileCompleter),
		),
		readline.PcItem("rm",
			readline.PcItemDynamic(fileCompleter),
			readline.PcItem("-r",
				readline.PcItemDynamic(directoryCompleter),
			),
		),
		readline.PcItem("tree",
			readline.PcItemDynamic(directoryCompleter),
		),
		readline.PcItem("mkdir"),
		readline.PcItem("touch"),
	)

	c.Dynamic = true
	return c
}

/*
 *	Custom Completers
 */

func directoryCompleter(path string) []string {

	if shellCommandWithPath.MatchString(path) {
		// extract path from command
		paths := shellPath.FindAllString(path, -1)
		path = paths[len(paths)-1]
	} else {
		// search in current dir
		path = "./"
	}

	names := make([]string, 0)
	files, err := ioutil.ReadDir(path)
	if err != nil {

		// check if path is multilevel
		// otherwise read current directory
		// the error for reading the directory can be ignored
		// because when the path is invalid there will be no completions and an empty string array is returned
		// this behaviour is equivalent with the bash shell
		arr := strings.Split(path, "/")
		if len(arr) > 1 {
			// trim base
			path = strings.TrimSuffix(path, filepath.Base(path))
			files, _ = ioutil.ReadDir(path)
		} else {
			files, _ = ioutil.ReadDir("./")
		}
	}
	for _, f := range files {
		if f.IsDir() {
			names = append(names, f.Name()+"/")
		}
	}

	return names

}

func fileCompleter(path string) []string {

	if shellCommandWithPath.MatchString(path) {
		// extract path from command
		paths := shellPath.FindAllString(path, -1)
		path = paths[len(paths)-1]
	} else {
		// search in current dir
		path = "./"
	}

	names := make([]string, 0)
	files, err := ioutil.ReadDir(path)
	if err != nil {

		// check if path is multilevel
		// otherwise read current directory
		// the error for reading the directory can be ignored
		// because when the path is invalid there will be no completions and an empty string array is returned
		// this behaviour is equivalent with the bash shell
		arr := strings.Split(path, "/")
		if len(arr) > 1 {
			// trim base
			path = strings.TrimSuffix(path, filepath.Base(path))
			files, _ = ioutil.ReadDir(path)
		} else {
			files, _ = ioutil.ReadDir("./")
		}

	}
	for _, f := range files {
		if f.IsDir() {
			names = append(names, f.Name()+"/")
			continue
		}
		names = append(names, f.Name())
	}

	return names
}
