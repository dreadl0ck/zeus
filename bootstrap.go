/*
 *  ZEUS - A Powerful Build System
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
	f, err := os.Create("zeus/" + name)
	if err != nil {
		cLog.WithError(err).Fatal("failed to create file: ", "zeus/"+name)
	}
	defer f.Close()

	f.Write(content)

	return
}

// bootstrap basic zeus scripts
// useful when starting from scratch
func bootstrapCommand() {

	err := os.Mkdir("zeus", 0700)
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
