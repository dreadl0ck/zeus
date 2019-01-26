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
	cp = &ansiProfile{}

	// ErrUnknownColorProfile means the color profile does not exist
	ErrUnknownColorProfile = errors.New("unknown color profile")
)

// ANSI Escape Sequence Representation of a ColorProfile
// contains a mutex to make changes on the fly possible
// without a data race
type ansiProfile struct {
	Text       string
	Prompt     string
	CmdOutput  string
	CmdName    string
	CmdFields  string
	CmdArgs    string
	CmdArgType string
	Reset      string
	sync.RWMutex
}

// ColorProfile for terminal colors
// it contains the colors following the ansi package style format
// https://github.com/mgutz/ansi
type ColorProfile struct {
	Text       string `yaml:"Text"`
	Prompt     string `yaml:"Prompt"`
	CmdOutput  string `yaml:"CmdOutput"`
	CmdName    string `yaml:"CmdName"`
	CmdFields  string `yaml:"CmdFields"`
	CmdArgs    string `yaml:"CmdArgs"`
	CmdArgType string `yaml:"CmdArgType"`
}

func printColorsUsageErr() {
	conf.Lock()
	l.Println("current color profile: " + conf.fields.ColorProfile)
	conf.Unlock()
	l.Println("usage: colors [default | off" + getAvailableColorProfiles() + "]")
}

func getAvailableColorProfiles() (res string) {
	conf.Lock()
	defer conf.Unlock()
	for name := range conf.fields.ColorProfiles {
		res += " | " + name
	}
	return
}

// low contrast profile
func darkProfile() *ColorProfile {
	return &ColorProfile{
		Text:       "cyan",
		Prompt:     "blue",
		CmdOutput:  "white",
		CmdName:    "blue",
		CmdFields:  "yellow",
		CmdArgs:    "white+h",
		CmdArgType: "green",
	}
}

// high contrast profile
func lightProfile() *ColorProfile {
	return &ColorProfile{
		Text:       "white",
		Prompt:     "yellow",
		CmdOutput:  "white+h",
		CmdName:    "red",
		CmdFields:  "blue",
		CmdArgs:    "cyan",
		CmdArgType: "green",
	}
}

// default terminal color profile
func defaultProfile() *ColorProfile {
	return &ColorProfile{
		Text:       "green",
		Prompt:     "red",
		CmdOutput:  "white+h",
		CmdName:    "red",
		CmdFields:  "yellow",
		CmdArgs:    "red",
		CmdArgType: "cyan+h",
	}
}

// black terminal color profile
func blackProfile() *ColorProfile {
	return &ColorProfile{
		Text:       "black",
		Prompt:     "black",
		CmdOutput:  "black",
		CmdName:    "black",
		CmdFields:  "black",
		CmdArgs:    "black",
		CmdArgType: "black",
	}
}

// profile with disabled colors
func colorsOffProfile() *ColorProfile {
	return &ColorProfile{
		Text:       ansi.ColorCode("off"),
		Prompt:     ansi.ColorCode("off"),
		CmdOutput:  ansi.ColorCode("off"),
		CmdName:    ansi.ColorCode("off"),
		CmdFields:  ansi.ColorCode("off"),
		CmdArgs:    ansi.ColorCode("off"),
		CmdArgType: ansi.ColorCode("off"),
	}
}

// handle colors shell command
func handleColorsCommand(args []string) {

	if len(args) < 2 {
		printColorsUsageErr()
		return
	}

	var (
		err     error
		profile = args[1]
	)

	// lock to prevent a race on the global ansiProfile instance
	cp.Lock()
	switch profile {
	case "off":
		cp = colorsOffProfile().parse()
	case "default":
		cp = defaultProfile().parse()
	case "black":
		cp = blackProfile().parse()
	default:

		// lookup profile name in config
		conf.Lock()
		if p, ok := conf.fields.ColorProfiles[profile]; ok {
			conf.Unlock()
			cp = p.parse()
		} else {
			// no change to colorProfile - Unlock it
			cp.Unlock()
			conf.Unlock()
			Log.Error(ErrUnknownColorProfile)
			return
		}
	}
	Log.Info("color profile set to: ", profile)

	// update value in config
	conf.Lock()
	conf.fields.ColorProfile = profile
	if profile == "off" {
		conf.fields.Colors = false
		// load uncolored ascii art
		asciiArt, err = assetBox.String("ascii_art.txt")
		if err != nil {
			Log.WithError(err).Fatal("failed to get ascii_art.txt from rice box")
		}
	} else {
		conf.fields.Colors = true
		// load colored ascii art
		asciiArt, err = assetBox.String("ascii_art_color.txt")
		if err != nil {
			Log.WithError(err).Fatal("failed to get ascii_art_color.txt from rice box")
		}
	}
	conf.Unlock()

	blockWriteEvent()

	// update config on disk
	conf.update()

	readlineMutex.Lock()
	if rl != nil {
		rl.SetPrompt(printPrompt())
		readlineMutex.Unlock()
		clearScreen()

		l.Println(cp.Text + asciiArt + "v" + version)
		l.Println(cp.Text + "Project Name: " + cp.Prompt + filepath.Base(workingDir) + cp.Text + "\n")

		printBuiltins()
		printCommands()
	}
}

// init the current color profile from config
func initColorProfile() {

	conf.Lock()
	defer conf.Unlock()

	// look up current profile string from config
	profile := conf.fields.ColorProfile

	// lock to prevent a race on the global ansiProfile instance
	cp.Lock()
	switch profile {
	case "off":
		cp = colorsOffProfile().parse()
		cp.Reset = ""
	case "default":
		cp = defaultProfile().parse()
	case "black":
		cp = blackProfile().parse()
	default:
		if p, ok := conf.fields.ColorProfiles[profile]; ok {
			cp = p.parse()
		} else {
			// no change to colorProfile - unlock it
			cp.Unlock()
			Log.Error(ErrUnknownColorProfile, " : ", conf.fields.ColorProfile)
			return
		}
	}
}

// convert a ColorProfile to an ansiProfile
func (cp *ColorProfile) parse() *ansiProfile {
	return &ansiProfile{
		Text:       ansi.ColorCode(cp.Text),
		Prompt:     ansi.ColorCode(cp.Prompt),
		CmdArgs:    ansi.ColorCode(cp.CmdArgs),
		CmdArgType: ansi.ColorCode(cp.CmdArgType),
		CmdFields:  ansi.ColorCode(cp.CmdFields),
		CmdName:    ansi.ColorCode(cp.CmdName),
		CmdOutput:  ansi.ColorCode(cp.CmdOutput),
		Reset:      ansi.Reset,
	}
}
