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
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
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
	zeusDir           = "zeus"
	projectConfigPath string

	// regex for matching first level JSON keys from config file contents
	jsonField = regexp.MustCompile("\\s\"(\\s)*[A-Z]?(.|\\s)*\":")

	configMutex = &sync.Mutex{}
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
	PortWebPanel        int
	PortGlueServer      int
	ExitOnInterrupt     bool
	DisableTimestamps   bool
	PrintBuiltins       bool
	StopOnError         bool
	DumpScriptOnError   bool
	DateFormat          string
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
		PortWebPanel:        8080,
		ExitOnInterrupt:     true,
		DisableTimestamps:   false,
		PrintBuiltins:       true,
		StopOnError:         true,
		DumpScriptOnError:   true,
		// german date format
		DateFormat: "02-01-2006",
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

	contents, err := validateConfigJSON(globalConfigPath)
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

// check for unknown fields in the config
// since JSON simply ignores them
func validateConfigJSON(path string) ([]byte, error) {

	c, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var (
		items      = configItems()
		foundField bool
	)

	for i, line := range strings.Split(string(c), "\n") {
		field := jsonField.FindString(line)
		if field != "" {
			field = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(field), "\""), "\":")
			for _, item := range items {
				if field == strings.TrimSpace(string(item.GetName())) {
					foundField = true
				}
			}
			if !foundField {
				Log.Warn("line "+strconv.Itoa(i)+": unknown config field: ", field)
			}
			foundField = false
		}
	}

	return c, nil
}

// parse the local project JSON config
func parseProjectConfig() (*config, error) {

	projectConfigPath = zeusDir + "/zeus_config.json"

	var c = new(config)

	stat, err := os.Stat(projectConfigPath)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, ErrConfigFileIsADirectory
	}

	contents, err := validateConfigJSON(projectConfigPath)
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

	if len(args) < 2 {
		printConfiguration()
		return
	}

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

	configMutex.Lock()
	defer configMutex.Unlock()

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

// remove config event
func cleanConfigEvent() string {

	var id string
	eventLock.Lock()
	for _, e := range projectData.Events {
		if e.Name == "config event" {
			id = e.ID
		}
	}
	eventLock.Unlock()

	if id != "" {
		removeEvent(id)
	}

	return id
}

// remove formatter event
func cleanFormatterEvent() string {

	var id string
	eventLock.Lock()
	for _, e := range projectData.Events {
		if e.Name == "formatter event" {
			id = e.ID
		}
	}
	eventLock.Unlock()

	if id != "" {
		removeEvent(id)
	}

	return id
}

// watch and reload config on changes
func (c *config) watch(eventID string) {

	Log.Debug("watching config at " + projectConfigPath)

	err := addEvent(newEvent(projectConfigPath, fsnotify.Write, "config event", ".json", eventID, "internal", func(event fsnotify.Event) {

		Log.Info("config watcher event: ", event.Name)

		b, err := validateConfigJSON(projectConfigPath)
		if err != nil {
			Log.WithError(err).Fatal("failed to read config")
		}

		configMutex.Lock()
		err = json.Unmarshal(b, c)
		if err != nil {
			Log.WithError(err).Error("config parse error")
		}
		configMutex.Unlock()

		// handle updated values
		c.handle()
	}))
	if err != nil {
		Log.WithError(err).Fatal("projectConfig watcher failed")
	}
}

// get type and current vlaue information for a given field on the config struct
func (c *config) getFieldInfo(field string) string {

	configMutex.Lock()
	defer configMutex.Unlock()

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

	configMutex.Lock()

	// check if the named field exists on the struct
	f := reflect.Indirect(reflect.ValueOf(c)).FieldByName(field)

	configMutex.Unlock()

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

	configMutex.Lock()
	defer configMutex.Unlock()

	if c.Debug {
		Log.Level = logrus.DebugLevel
	} else {
		Log.Level = logrus.InfoLevel
	}

	// enable dumping the script on error when the auto formatter is enabled
	if c.AutoFormat {
		c.DumpScriptOnError = true
	}

	// disable colors if requested
	if !c.Colors {
		Log.Formatter = &prefixed.TextFormatter{
			DisableColors: true,
		}
	} else {
		Log.Formatter = &prefixed.TextFormatter{}
	}

	if !c.AutoFormat {
		cleanFormatterEvent()
	}
}
