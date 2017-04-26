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
	"errors"
	"path/filepath"
	"sync"

	"github.com/mgutz/ansi"
)

var (
	// global ANSI terminal color profile
	cp *colorProfile

	// ErrUnknownColorProfile means the color profile does not exist
	ErrUnknownColorProfile = errors.New("unknown color profile")

	colorProfileMutex = &sync.Mutex{}
)

type colorProfile struct {
	text       string
	prompt     string
	cmdOutput  string
	cmdName    string
	cmdFields  string
	cmdArgs    string
	cmdArgType string
}

func printColorsUsageErr() {
	l.Println(ErrInvalidUsage)
	l.Println("usage: colors <default | dark | light>")
}

func darkProfile() *colorProfile {
	return &colorProfile{
		text:       ansi.Black,
		prompt:     ansi.Blue,
		cmdOutput:  ansi.LightWhite,
		cmdName:    ansi.Blue,
		cmdFields:  ansi.Yellow,
		cmdArgs:    ansi.LightBlack,
		cmdArgType: ansi.Green,
	}
}

func lightProfile() *colorProfile {
	return &colorProfile{
		text:       ansi.Black,
		prompt:     ansi.Green,
		cmdOutput:  ansi.Black,
		cmdName:    ansi.Blue,
		cmdFields:  ansi.Green,
		cmdArgs:    ansi.Cyan,
		cmdArgType: ansi.Green,
	}
}

func defaultProfile() *colorProfile {
	return &colorProfile{
		text:       ansi.Green,
		prompt:     ansi.Red,
		cmdOutput:  ansi.LightWhite,
		cmdName:    ansi.Red,
		cmdFields:  ansi.Yellow,
		cmdArgs:    ansi.Red,
		cmdArgType: ansi.Green,
	}
}

// handle colors shell command
func handleColorsCommand(args []string) {

	if len(args) < 2 {
		printColorsUsageErr()
		return
	}

	profile := args[1]

	colorProfileMutex.Lock()

	switch profile {
	case "dark":
		cp = darkProfile()
	case "light":
		cp = lightProfile()
	case "default":
		cp = defaultProfile()
	default:
		Log.Error(ErrUnknownColorProfile)
		colorProfileMutex.Unlock()
		return
	}
	colorProfileMutex.Unlock()

	Log.Info("color profile set to: ", profile)

	configMutex.Lock()
	conf.ColorProfile = profile
	configMutex.Unlock()

	conf.update()

	readlineMutex.Lock()
	if rl != nil {
		rl.SetPrompt(printPrompt())
		readlineMutex.Unlock()
		clearScreen()

		l.Println(cp.text + asciiArt + ansi.Reset + "\n")
		l.Println(cp.text + "Project Name: " + cp.prompt + filepath.Base(workingDir) + cp.text + "\n")

		printBuiltins()
		printCommands()
	}
}

func initColorProfile() {
	switch conf.ColorProfile {
	case "dark":
		cp = darkProfile()
	case "light":
		cp = lightProfile()
	case "default":
		cp = defaultProfile()
	default:
		Log.Fatal(ErrUnknownColorProfile, " : ", conf.ColorProfile)
	}
}
