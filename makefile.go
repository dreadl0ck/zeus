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
	"bytes"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// regular expressions to match various elements from a makefile
var (
	makefileTarget = regexp.MustCompile("^[^\\s]*([a-z]+):")
	makeTargetBody = regexp.MustCompile("^[\\s\\p{Zs}]+")

	global               = regexp.MustCompile("^[^\\s]*([A-Z]+)(\\s)*(=)")
	makefileVar          = regexp.MustCompile("\\$\\([A-Z]+(_[A-Z]+)*\\)")
	makefileShellCommand = regexp.MustCompile("\\$\\(shell(\\s*([\\w|\\S]*))*\\)")

	bashBuiltin = regexp.MustCompile("(@[a-z]*\\s+)")
	makeCommand = regexp.MustCompile("make\\s+[a-z0-9]+")
)

// print an overview of the available makefile commands to stdout
func printMakefileCommandOverview() {

	b, err := ioutil.ReadFile("Makefile")
	if err != nil {
		Log.WithError(err).Debug("unable to read Makefile")
		return
	}

	l.Println("available GNUMake Commands:")

	for _, line := range bytes.Split(b, []byte("\n")) {

		if makefileTarget.Match(line) && !bytes.Contains(line, []byte("\t")) {
			l.Println("~> " + strings.TrimSuffix(string(line), ":"))
		}
	}
	l.Println("")
}

// migrate Makefile into a zeus command folder
func migrateMakefile() {

	var (
		file            *os.File
		writeInProgress bool
		err             error
		perm            = os.FileMode(0700)
		dir             = zeusDir
	)

	Log.WithField("dir", dir).Info("Makefile migration started.")

	contents, err := ioutil.ReadFile("Makefile")
	if err != nil {
		if testingMode {
			contents, err = ioutil.ReadFile("tests/Makefile")
		} else {
			Log.WithError(err).Debug("unable to read Makefile")
			return
		}
	}

	// create dir
	err = os.Mkdir(dir, perm)
	if err != nil {
		Log.WithError(err).Error("failed to create: ", zeusDir)
		return
	}

	for _, line := range bytes.Split(contents, []byte("\n")) {

		if writeInProgress {

			// write empty lines to file
			// they can can be used used for formatting
			// which is perfectly valid in makefiles
			if len(line) == 0 {
				file.WriteString(string(line) + "\n")
				continue
			}

			// match everything preceded by whitespace
			if makeTargetBody.Match(line) {

				// trim whitespace
				line = bytes.TrimSpace(line)

				// replace '@builtin' with 'builtin'
				line = bashBuiltin.ReplaceAll(line, bytes.TrimPrefix(bashBuiltin.Find(line), []byte("@")))

				// replace $(VAR) with $VAR
				var step1 = bytes.TrimSuffix(bytes.TrimPrefix(makefileVar.Find(line), []byte("$(")), []byte(")")) // VAR

				// escaping the $ with another $ is important here,
				// otherwise the regex engine treats it as an expansion!
				var replace = append([]byte("$$"), step1...) // $VAR

				line = makefileVar.ReplaceAll(line, replace)

				// replace $(shell ...) commands with $(...)
				// replace $(shell command > test) with $(command > test)
				step1 = bytes.TrimPrefix(makefileShellCommand.Find(line), []byte("$(shell ")) // command > test)

				replace = append([]byte("$("), step1...) // $(command > test)
				line = makefileShellCommand.ReplaceAll(line, replace)

				// replace all make <command> with zeus <command>
				line = makeCommand.ReplaceAll(line, append([]byte("g"), makeCommand.Find(line)...))

				// convert if statements
				if bytes.HasSuffix(line, []byte("\\")) {

					if bytes.HasPrefix(line, []byte("if")) {
						line = bytes.TrimSuffix(line, []byte("\\"))
						line = append(line, []byte("then")...)
					} else if bytes.HasPrefix(line, []byte("then")) {
						// delete line with then
						line = []byte{}
					} else {
						line = bytes.TrimSuffix(line, []byte("\\"))
						line = bytes.TrimSpace(line)
						line = bytes.TrimSuffix(line, []byte(";"))
						line = append([]byte("\t"), line...)
					}
				} else if bytes.HasPrefix(line, []byte("fi;")) {
					line = []byte("fi")
				}

				// write to file
				file.WriteString(string(line) + "\n")
			} else {
				writeInProgress = false
			}
		}

		if makefileTarget.Match(line) && !makeTargetBody.Match(line) {

			l.Println("migrating target ~> " + strings.TrimSuffix(string(line), ":"))

			var (
				args       = strings.Split(string(line), " ")
				targetName = strings.TrimSuffix(args[0], ":")
				filename   = dir + "/" + targetName + f.fileExtension
			)

			file, err = os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
			if err != nil {
				Log.WithError(err).Error("failed to create file: ", filename, "permMode: ", perm)
			}
			defer file.Close()

			file.WriteString(p.bang + "\n\n")
			file.WriteString("# -------------------------------------------------------------------------- #" + "\n")
			file.WriteString("# @zeus-chain: " + strings.Join(args[1:], " -> ") + "\n")
			file.WriteString("# @zeus-args: " + "\n")
			file.WriteString("# @zeus-help: simple help text for command " + targetName + "\n")
			file.WriteString("# -------------------------------------------------------------------------- #" + "\n")
			file.WriteString("# manual text for command " + targetName + "\n")
			file.WriteString("# -------------------------------------------------------------------------- #" + "\n\n")

			writeInProgress = true
		}
	}

	// handle globals

	var globals string
	for _, line := range bytes.Split(contents, []byte("\n")) {
		if global.Match(line) {
			globals += string(line) + "\n"
		}
	}

	if len(globals) > 0 {
		f, err := os.Create(dir + "/globals.sh")
		if err != nil {
			Log.WithError(err).Error("failed to create globals file")
			return
		}
		defer f.Close()
		f.WriteString("#!/bin/bash\n")
		f.WriteString(globals + "\n")
		l.Println("created " + dir + "/globals.sh")
	}

	l.Println("migrated Makefile")
}

// handle makefile shell commands
func handleMakefileCommand(args []string) {

	if len(args) < 2 {
		printMakefileCommandOverview()
		return
	}

	if args[1] == "migrate" {
		migrateMakefile()
		return
	}

	Log.Error("unknown sub command: " + args[1])
}
