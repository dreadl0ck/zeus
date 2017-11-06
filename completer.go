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
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/dreadl0ck/readline"
)

var (
	// regex to the match a UNIX path
	shellPath = regexp.MustCompile("(([a-z]*[A-Z]*[0-9]*(_|-)*)*/*)*")

	// regex to match a command with a trailing UNIX path
	shellCommandWithPath = regexp.MustCompile("([a-z]*\\s*)*(([a-z]*[A-Z]*[0-9]*(_|-)*)*/*)*")

	// completer for the the events add subcommand
	addEventCompleter = readline.PcItemDynamic(fileCompleter,
		readline.PcItemDynamic(fileTypeCompleter,
			readline.PcItemDynamic(commandCompleter),
		),
		readline.PcItemDynamic(commandCompleter),
	)
)

type atomicCompleter struct {
	*readline.PrefixCompleter
	sync.RWMutex
}

func newAtomicCompleter() *atomicCompleter {
	return &atomicCompleter{
		newCompleter(),
		sync.RWMutex{},
	}
}

// assemble and return all items for config item completion
// also used for validating the config YAML for unknown fields
// if there's a key in the config that is not in here there will be a warning
func configItems() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("makefileOverview", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("autoFormat", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("fixParseErrors", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("colors", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("passCommandsToShell", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("eebInterface", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("interactive", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("debug", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("recursionDepth"),
		readline.PcItem("projectNamePrompt", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("colorProfile"),
		readline.PcItem("historyFile", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("historyLimit"),
		readline.PcItem("exitOnInterrupt", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("disableTimestamps", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("printBuiltins", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("dumpScriptOnError", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("stopOnError", readline.PcItem("true"), readline.PcItem("false")),
		readline.PcItem("portWebPanel"),
		readline.PcItem("portGlueServer"),
		readline.PcItem("dateFormat"),
		readline.PcItem("todoFilePath"),
		readline.PcItem("editor"),
		readline.PcItem("codeSnippetScope"),
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
		readline.PcItem(exitCommand),
		readline.PcItem(helpCommand,
			readline.PcItemDynamic(commandCompleter),
		),
		readline.PcItem(infoCommand),
		readline.PcItem(clearCommand),
		readline.PcItem(formatCommand),
		readline.PcItem(globalsCommand),
		readline.PcItem(versionCommand),
		readline.PcItem(configCommand,
			readline.PcItem("set",
				configItems()...,
			),
			readline.PcItem("get",
				configItems()...,
			),
		),
		readline.PcItem(createCommand,
			readline.PcItemDynamic(languageCompleter),
			readline.PcItem("script",
				readline.PcItem("all"),
				readline.PcItemDynamic(commandCompleter),
			),
		),
		readline.PcItem(eventsCommand,
			readline.PcItem("add",
				readline.PcItem("WRITE",
					addEventCompleter,
				),
				readline.PcItem("REMOVE",
					addEventCompleter,
				),
				readline.PcItem("CHMOD",
					addEventCompleter,
				),
				readline.PcItem("RENAME",
					addEventCompleter,
				),
			),
			readline.PcItem("remove",
				readline.PcItemDynamic(eventIDCompleter),
			),
		),
		readline.PcItem(milestonesCommand,
			readline.PcItem("set"),
			readline.PcItem("remove"),
			readline.PcItem("add"),
		),
		readline.PcItem(gitFilterCommand),
		readline.PcItem(deadlineCommand,
			readline.PcItem("set"),
			readline.PcItem("remove"),
		),
		readline.PcItem(makefileCommand,
			readline.PcItem("migrate"),
		),
		readline.PcItem(dataCommand),
		readline.PcItem(aliasCommand,
			readline.PcItem("set"),
			readline.PcItem("remove"),
		),
		readline.PcItem(todoCommand,
			readline.PcItem("add"),
			readline.PcItem("remove",
				readline.PcItemDynamic(todoIndexCompleter),
			),
		),
		readline.PcItem(generateCommand,
			readline.PcItemDynamic(commandCompleter),
		),
		readline.PcItem(colorsCommand,
			readline.PcItem("off"),
			readline.PcItem("default"),
			readline.PcItemDynamic(colorProfileCompleter),
		),
		readline.PcItem(authorCommand,
			readline.PcItem("set"),
			readline.PcItem("remove"),
		),
		readline.PcItem(updateCommand),
		readline.PcItem(builtinsCommand),
		readline.PcItem(keysCommand,
			readline.PcItem("set",
				keyKombItems()...,
			),
			readline.PcItem("remove",
				keyKombItems()...,
			),
		),
		readline.PcItem(editCommand,
			readline.PcItemDynamic(commandCompleter),
			readline.PcItem("commands",
				readline.PcItem("line"),
			),
			readline.PcItem("data",
				readline.PcItem("line"),
			),
			readline.PcItem("config",
				readline.PcItem("line"),
			),
			readline.PcItem("todo",
				readline.PcItem("line"),
			),
			readline.PcItem("globals",
				readline.PcItemDynamic(languageCompleter),
			),
		),
		readline.PcItem(webCommand),
		readline.PcItem(procsCommand,
			readline.PcItem("detach",
				readline.PcItemDynamic(commandCompleter),
			),
			readline.PcItem("kill",
				readline.PcItemDynamic(pIDCompleter),
			),
			readline.PcItem("attach",
				readline.PcItemDynamic(pIDCompleter),
			),
		),
		readline.PcItem(wikiCommand),
		// completions for common shell commands
		readline.PcItem("git",
			readline.PcItem("add"),
			readline.PcItem("status"),
			readline.PcItem("commit"),
		),
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
		readline.PcItem("micro",
			readline.PcItemDynamic(fileCompleter),
		),
	)

	c.Dynamic = true
	return c
}

/*
 *	Custom Completers
 */

// complete eventIDs for removing events
func eventIDCompleter(path string) (res []string) {
	projectData.Lock()
	defer projectData.Unlock()
	for _, e := range projectData.fields.Events {
		res = append(res, e.ID)
	}
	return
}

// complete available commands
func commandCompleter(path string) (res []string) {
	cmdMap.Lock()
	defer cmdMap.Unlock()
	for name := range cmdMap.items {
		res = append(res, name)
	}
	return
}

// complete available parser languages
func languageCompleter(path string) (res []string) {
	ls.Lock()
	defer ls.Unlock()
	for name := range ls.items {
		res = append(res, name)
	}
	return
}

func colorProfileCompleter(path string) (res []string) {
	conf.Lock()
	defer conf.Unlock()
	for name := range conf.fields.ColorProfiles {
		res = append(res, name)
	}
	return
}

func todoIndexCompleter(path string) (res []string) {
	contents, err := ioutil.ReadFile(conf.fields.TodoFilePath)
	if err != nil {
		l.Println(err)
		return
	}
	var index int
	for _, line := range strings.Split(string(contents), "\n") {
		if strings.HasPrefix(line, "- ") {
			index++
			res = append(res, strconv.Itoa(index))
		}
	}
	return
}

// complete PIDs for killing processes
func pIDCompleter(path string) (res []string) {
	projectData.Lock()
	defer projectData.Unlock()
	for _, p := range processMap {
		res = append(res, strconv.Itoa(p.PID))
	}
	return
}

// complete available filetypes for the event target directory
func fileTypeCompleter(path string) (res []string) {

	var (
		fields = strings.Fields(path)
		dir    string
	)

	if len(fields) > 2 {
		dir = fields[3]
	} else {
		return
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		Log.Error(err)
		return res
	}

	for _, f := range files {
		res = append(res, filepath.Ext(f.Name()))
	}

	// remove duplicates
	var (
		out []string
		ok  bool
	)

	for _, path := range res {
		for _, name := range out {
			if path == name {
				ok = true
			}
		}
		if !ok && path != "" {
			out = append(out, path)
		}
		ok = false
	}

	return out
}

// return available directories
func directoryCompleter(path string) (names []string) {

	files, dir := getFilesInDir(path)

	for _, f := range files {
		if dir == "./" {
			if f.IsDir() {
				names = append(names, strings.TrimPrefix(f.Name(), "./")+"/")
			}
			continue
		}
		if f.IsDir() {
			names = append(names, f.Name()+"/")
		}
	}
	return
}

func getFilesInDir(path string) (files []os.FileInfo, dir string) {

	var (
		fields = strings.Fields(path)
		fLen   = len(fields)
		err    error
	)
	if fLen < 2 {
		path = "./"
	} else {
		path = fields[fLen-1]
	}

	dir, _ = filepath.Split(path)
	files, err = ioutil.ReadDir(dir)
	if err != nil {
		if len(strings.Split(dir, "/")) > 1 {
			files, _ = ioutil.ReadDir(filepath.Base(dir))
		} else {
			dir = "./"
			files, _ = ioutil.ReadDir(dir)
		}
	}

	return
}

func fileCompleter(path string) (names []string) {

	files, dir := getFilesInDir(path)

	for _, f := range files {
		if dir == "./" {
			name := strings.TrimPrefix(f.Name(), "./")
			if f.IsDir() {
				names = append(names, name+"/")
				continue
			}
			names = append(names, name)
			continue
		}

		if f.IsDir() {
			names = append(names, f.Name()+"/")
			continue
		}
		names = append(names, f.Name())
	}
	return
}
