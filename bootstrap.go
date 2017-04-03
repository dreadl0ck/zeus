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

import "os"

/*
 *	Bootstrapping
 */

// create a file with name and content
func bootstrapFile(name string) {

	var cLog = Log.WithField("prefix", "bootstrapFile")

	content, err := assetBox.Bytes(name)
	if err != nil {
		cLog.WithError(err).Fatal("failed to get content for file: " + name)
	}

	cLog.Info("creating file: ", name)
	f, err := os.Create(zeusDir + "/" + name)
	if err != nil {
		cLog.WithError(err).Fatal("failed to create file: ", zeusDir+"/"+name)
	}
	defer f.Close()

	f.Write(content)

	return
}

func printBootstrapCommandUsageErr() {
	l.Println("usage: zeus bootstrap <file | dir>")
}

// bootstrap basic zeus scripts
// useful when starting from scratch
func runBootstrapDirCommand() {

	err := os.Mkdir(zeusDir, 0700)
	if err != nil {
		Log.WithError(err).Fatal("failed to create zeus directory")
	}

	bootstrapFile("clean.sh")
	bootstrapFile("build.sh")
	bootstrapFile("run.sh")
	bootstrapFile("test.sh")
	bootstrapFile("install.sh")
	bootstrapFile("bench.sh")
}

// bootstrap basic zeus scripts
// useful when starting from scratch
func runBootstrapFileCommand() {

	f, err := os.Create("Zeusfile.yml")
	if err != nil {
		Log.WithError(err).Fatal("failed to create Zeusfile")
	}
	defer f.Close()

	f.WriteString(`############
# ZEUSFILE #
############

# globals for all build commands
globals: |
    version=0.1

# all commands
commands: 
    build:
        chain: clean
        help: build project
        buildNumber: true
        run: 
	clean:
        help: clean up to prepare for build
        run: rm -rf bin/*
    install:
        chain: clean
        help: install to $PATH
        manual: Install the application to the default system location
        run:
`)
}

func printCreateCommandUsageErr() {
	l.Println("usage:")
	l.Println("zeus create <command>")
}

func handleCreateCommand(args []string) {

	if len(args) < 2 {
		printCreateCommandUsageErr()
		return
	}

	filename := zeusDir + "/" + args[1] + f.fileExtension

	_, err := os.Stat(filename)
	if err == nil {
		l.Println("file " + filename + " exists!")
		return
	}

	f, err := os.Create(filename)
	if err != nil {
		l.Println("failed to create file: ", err)
		return
	}
	defer f.Close()

	f.WriteString(`#!/bin/bash

# ---------------------------------------------------------------------- #
# @zeus-chain: 
# @zeus-help: 
# @zeus-args: 
# ---------------------------------------------------------------------- #
#
# ---------------------------------------------------------------------- #

echo "implement me!"
`)

	l.Println("created zeus command at " + filename)

	// parse new command
	err = addCommand(filename, true)
	if err != nil {
		Log.WithError(err).Error("failed to add command")
	}
}
