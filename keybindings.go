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
	"errors"
	"strings"

	"github.com/chzyer/readline"
)

var (
	// global listener for key events
	listener = readline.FuncListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {

		if key > 26 {
			return
		}

		// Log.Info(key)

		if keyName, ok := keyMap[key]; ok {
			if chain, ok := projectData.KeyBindings[keyName]; ok {
				executeCommand(chain)
			}
		}

		return
	})

	// mapped runes to keyComb strings
	keyMap = map[rune]string{
		1:  "Ctrl-A",
		2:  "Ctrl-B",
		5:  "Ctrl-E",
		6:  "Ctrl-F",
		7:  "Ctrl-G",
		8:  "Ctrl-H",
		9:  "Ctrl-I",
		10: "Ctrl-J",
		11: "Ctrl-K",
		12: "Ctrl-L",
		13: "Ctrl-M",
		14: "Ctrl-N",
		15: "Ctrl-O",
		16: "Ctrl-P",
		17: "Ctrl-Q",
		18: "Ctrl-R",
		19: "Ctrl-S",
		20: "Ctrl-T",
		21: "Ctrl-U",
		22: "Ctrl-V",
		23: "Ctrl-W",
		24: "Ctrl-X",
		25: "Ctrl-Y",
	}

	// ErrInvalidKeyComb means the KeyComb does not exist
	ErrInvalidKeyComb = errors.New("invalid key combination")
)

// handle keys shell command
func handleKeysCommand(args []string) {

	if len(args) < 2 {
		printKeybindings()
		return
	}

	if len(args) < 3 {
		printKeybindingsCommmandUsageErr()
		return
	}

	var ok bool

	// check if KeyComb exists
	for _, s := range keyMap {
		if s == args[2] {
			ok = true
		}
	}
	if !ok {
		Log.Error(ErrInvalidKeyComb)
		return
	}

	if args[1] == "set" {

		if len(args) < 4 {
			printKeybindingsCommmandUsageErr()
			return
		}

		projectData.KeyBindings[args[2]] = strings.Join(args[3:], " ")
		projectData.update()

		Log.Info("key binding added")
	} else if args[1] == "remove" {
		delete(projectData.KeyBindings, args[2])
		projectData.update()

		Log.Info("key binding removed")
	} else {
		printKeybindingsCommmandUsageErr()
	}
}

func printKeybindingsCommmandUsageErr() {
	Log.Error(ErrInvalidUsage)
	Log.Info("usage: keys [set <KeyComb> <commandChain>] [remove <KeyComb>]")
}

func printKeybindings() {
	for key, cmd := range projectData.KeyBindings {
		l.Println(key, "=", cmd)
	}
}
