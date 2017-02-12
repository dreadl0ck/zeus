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

import (
	"errors"
	"path/filepath"

	"github.com/mgutz/ansi"
)

var (
	// global ANSI terminal color profile
	cp *colorProfile

	// ErrUnknownColorProfile means the color profile does not exist
	ErrUnknownColorProfile = errors.New("unkown color profile")
)

type colorProfile struct {
	colorText          string
	colorPrompt        string
	colorCommandOutput string
	colorCommandName   string
	colorCommandChain  string
}

func printColorsUsageErr() {
	Log.Error(ErrInvalidUsage)
	Log.Info("usage: colors <default | dark | light>")
}

func darkProfile() *colorProfile {
	return &colorProfile{
		colorText:          ansi.Black,
		colorPrompt:        ansi.Blue,
		colorCommandOutput: ansi.White,
		colorCommandName:   ansi.Blue,
		colorCommandChain:  ansi.White,
	}
}

func lightProfile() *colorProfile {
	return &colorProfile{
		colorText:          ansi.Black,
		colorPrompt:        ansi.White,
		colorCommandOutput: ansi.Black,
		colorCommandName:   ansi.White,
		colorCommandChain:  ansi.White,
	}
}

func defaultProfile() *colorProfile {
	return &colorProfile{
		colorText:          ansi.Green,
		colorPrompt:        ansi.Red,
		colorCommandOutput: ansi.White,
		colorCommandName:   ansi.Red,
		colorCommandChain:  ansi.White,
	}
}

// handle colors shell command
func handleColorsCommand(args []string) {

	if len(args) < 2 {
		printColorsUsageErr()
		return
	}

	profile := args[1]

	switch profile {
	case "dark":
		cp = darkProfile()
	case "light":
		cp = lightProfile()
	case "default":
		cp = defaultProfile()
	default:
		Log.Error(ErrUnknownColorProfile)
		return
	}
	Log.Info("color profile set to: ", profile)

	conf.ColorProfile = profile
	conf.update()

	if rl != nil {
		rl.SetPrompt(printPrompt())
		clearScreen()

		l.Println(cp.colorText + asciiArt + ansi.Reset + "\n")
		l.Println(cp.colorText + "Project Name: " + cp.colorPrompt + filepath.Base(workingDir) + cp.colorText + "\n")

		printBuiltins()
		printCommands()
	}
}
