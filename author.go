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

import "strings"

func printAuthorUsageErr() {
	Log.Error(ErrInvalidUsage)
	Log.Info("usage: author [set <name>] [remove]")
}

func printAuthor() {
	if projectData.Author != "" {
		l.Println("Author: ", cp.colorPrompt+projectData.Author)
	}
}

// handle author shell command
func handleAuthorCommand(args []string) {
	if len(args) < 2 {
		printAuthor()
		return
	}

	switch args[1] {
	case "set":
		if len(args) < 3 {
			printAuthorUsageErr()
			return
		}
		projectData.Author = strings.Join(args[2:], " ")
		projectData.update()
	case "remove":
		projectData.Author = ""
		projectData.update()
	default:
		printAuthorUsageErr()
	}
}
