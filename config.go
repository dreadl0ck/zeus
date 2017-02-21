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
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
)

var (
	// ErrConfigFileIsADirectory means the config file is a directory, thats wrong
	ErrConfigFileIsADirectory = errors.New("the config file is a directory")

	// ErrInvalidGlobalConfig means your global config sucks
	ErrInvalidGlobalConfig = errors.New("global configuration file is invalid")

	// ErrInvalidLocalConfig means your local config sucks
	ErrInvalidLocalConfig = errors.New("local configuration file is invalid")

	// path for global config file
	globalConfigPath = os.Getenv("HOME") + "/.zeus_config.json"

	// path for project config files
	projectConfigPath = "zeus/zeus_config.json"
	zeusDir           = "zeus"
)

// config contains configurable parameters
type config struct {
	MakefileOverview    bool
	AutoFormat          bool
	FixParseErrors      bool
	Colors              bool
	PassCommandsToShell bool
	WebInterface        bool
	Interactive         bool
	LogToFileColor      bool
	LogToFile           bool
	Debug               bool
	RecursionDepth      int
	ProjectNamePrompt   bool
	AllowUntypedArgs    bool
	ColorProfile        string
	HistoryFile         bool
	HistoryLimit        int
	ExitOnInterrupt     bool
	DisableTimestamps   bool
	PrintBuiltins       bool
	StopOnError         bool
	DumpScriptOnError   bool
}

// newConfig returns the default configuration in case there is no config file
func newConfig() *config {
	return &config{
		MakefileOverview:    true,
		AutoFormat:          false,
		FixParseErrors:      true,
		Colors:              true,
		PassCommandsToShell: true,
		WebInterface:        false,
		Interactive:         true,
		LogToFileColor:      false,
		LogToFile:           true,
		Debug:               false,
		RecursionDepth:      1,
		ProjectNamePrompt:   true,
		AllowUntypedArgs:    false,
		ColorProfile:        "default",
		HistoryFile:         true,
		HistoryLimit:        20,
		ExitOnInterrupt:     true,
		DisableTimestamps:   false,
		PrintBuiltins:       true,
		StopOnError:         true,
		DumpScriptOnError:   true,
	}
}

func printConfigUsageErr() {
	Log.Error(ErrInvalidUsage)
	Log.Info("usage: config [get <field>] [set <field> <value>]")
}

// parse the global JSON config
func parseGlobalConfig() (*config, error) {

	var c = new(config)

	stat, err := os.Stat(globalConfigPath)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, ErrConfigFileIsADirectory
	}

	contents, err := ioutil.ReadFile(globalConfigPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, c)
	if err != nil {
		Log.WithError(err).Fatal("failed to unmarshal confg - invalid JSON")
	}

	c.handle()

	return c, nil
}

// parse the local project JSON config
func parseProjectConfig() (*config, error) {

	var c = new(config)

	stat, err := os.Stat(projectConfigPath)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, ErrConfigFileIsADirectory
	}

	contents, err := ioutil.ReadFile(projectConfigPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, c)
	if err != nil {
		Log.WithError(err).Fatal("failed to unmarshal confg - invalid JSON")
	}

	c.handle()

	return c, nil
}

// handle config shell command
func handleConfigCommand(args []string) {

	switch args[1] {
	case "set":
		if len(args) < 4 {
			printConfigUsageErr()
			return
		}
		conf.setValue(args[2], args[3])
	case "get":
		if len(args) < 3 {
			printConfigUsageErr()
			return
		}
		Log.Info(conf.getFieldInfo(args[2]))
	default:
		Log.Error("invalid config command: ", args[1])
		printConfigUsageErr()
	}
}

// update config on disk
func (c *config) update() {

	// make it pretty
	b, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		Log.WithError(err).Fatal("failed to marshal config")
	}

	// open the config file write only and truncate if it exists
	f, err := os.OpenFile(projectConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		Log.WithError(err).Fatal("failed to open config")
	}

	// write to file
	_, err = f.Write(b)
	if err != nil {
		Log.WithError(err).Fatal("failed to write config")
	}
}

// watch and reload on changes
func (c *config) watch() {

	err := addEvent(projectConfigPath, fsnotify.Write, func(event fsnotify.Event) {

		// check if the event name is correct because watching the zeus dir will also result in an event for zeus/config.json
		if event.Name == projectConfigPath {

			b, err := ioutil.ReadFile(projectConfigPath)
			if err != nil {
				Log.WithError(err).Fatal("failed to read config")
			}

			err = json.Unmarshal(b, c)
			if err != nil {
				Log.WithError(err).Error("config parse error")
				return
			}
		}
	}, "")
	if err != nil {
		Log.WithError(err).Fatal("projectConfig watcher failed")
	}
}

// get type and current vlaue information for a given field on the config struct
func (c *config) getFieldInfo(field string) string {

	f := reflect.Indirect(reflect.ValueOf(c)).FieldByName(field)
	switch f.Kind() {
	case reflect.Bool:
		return "field type: " + f.Kind().String() + ", value: " + strconv.FormatBool(f.Bool())
	case reflect.Int:
		return "field type: " + f.Kind().String() + ", value: " + strconv.Itoa(int(f.Int()))
	default:
		Log.Error(f.Kind())
		return "unknown field"
	}
}

// set a config field to a specified value by its name
func (c *config) setValue(field, value string) {

	// check if the named field exists on the struct
	f := reflect.Indirect(reflect.ValueOf(c)).FieldByName(field)
	if !f.IsValid() {
		Log.Error("invalid config field: ", field)
		return
	}

	switch f.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			Log.WithError(err).Error("invalid boolean value: ", value)
			return
		}

		f.SetBool(b)

		Log.Info("set config field ", field, " to ", value)

	case reflect.Int:
		i, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			Log.WithError(err).Error("invalid integer value: ", value)
			return
		}

		f.SetInt(i)

		Log.Info("set config field ", field, " to ", value)
	default:
		Log.Error("unknown type: ", f.Kind())
		return
	}
	c.handle()
	c.update()
}

// handle the config by applying updated values
func (c *config) handle() {
	if c.Debug {
		Log.Level = logrus.DebugLevel
	} else {
		Log.Level = logrus.InfoLevel
	}

	// enable dumping the script on error when the auto formatter is enabled
	if c.AutoFormat {
		c.DumpScriptOnError = true
	}
}
