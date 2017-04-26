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
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

var (
	// ErrConfigFileIsADirectory means the config file is a directory, thats wrong
	ErrConfigFileIsADirectory = errors.New("the config file is a directory")

	// path for global config file
	globalConfigPath = os.Getenv("HOME") + "/.zeus_config.yml"

	// path for project config files
	zeusDir           = "zeus"
	projectConfigPath string

	// regex for matching top level YAML keys from config file contents
	yamlField = regexp.MustCompile("(\\s)*[a-z]?(.|\\s)*:")

	configMutex = &sync.Mutex{}
)

// config contains configurable parameters
type config struct {
	AutoFormat          bool   `yaml:"AutoFormat"`
	FixParseErrors      bool   `yaml:"FixParseErrors"`
	Colors              bool   `yaml:"Colors"`
	PassCommandsToShell bool   `yaml:"PassCommandsToShell"`
	WebInterface        bool   `yaml:"WebInterface"`
	Interactive         bool   `yaml:"Interactive"`
	Debug               bool   `yaml:"Debug"`
	RecursionDepth      int    `yaml:"RecursionDepth"`
	ProjectNamePrompt   bool   `yaml:"ProjectNamePrompt"`
	ColorProfile        string `yaml:"ColorProfile"`
	HistoryFile         bool   `yaml:"HistoryFile"`
	HistoryLimit        int    `yaml:"HistoryLimit"`
	PortWebPanel        int    `yaml:"PortWebPanel"`
	PortGlueServer      int    `yaml:"PortGlueServer"`
	ExitOnInterrupt     bool   `yaml:"ExitOnInterrupt"`
	DisableTimestamps   bool   `yaml:"DisableTimestamps"`
	PrintBuiltins       bool   `yaml:"PrintBuiltins"`
	MakefileOverview    bool   `yaml:"MakefileOverview"`
	StopOnError         bool   `yaml:"StopOnError"`
	DumpScriptOnError   bool   `yaml:"DumpScriptOnError"`
	DateFormat          string `yaml:"DateFormat"`
	TodoFilePath        string `yaml:"TodoFilePath"`
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
		Debug:               false,
		RecursionDepth:      1,
		ProjectNamePrompt:   true,
		ColorProfile:        "default",
		HistoryFile:         true,
		HistoryLimit:        20,
		PortWebPanel:        8080,
		ExitOnInterrupt:     true,
		DisableTimestamps:   false,
		PrintBuiltins:       false,
		StopOnError:         true,
		DumpScriptOnError:   true,
		// german date format
		DateFormat:   "02-01-2006",
		TodoFilePath: "TODO.md",
	}
}

func printConfigUsageErr() {
	l.Println(ErrInvalidUsage)
	l.Println("usage: config [get <field>] [set <field> <value>]")
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

	contents, err := validateConfig(globalConfigPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(contents, c)
	if err != nil {
		Log.WithError(err).Fatal("failed to unmarshal confg - invalid YAML")
	}

	c.handle()

	return c, nil
}

// check for unknown fields in the config
// since YAML simply ignores them and intializes them with their default values
func validateConfig(path string) ([]byte, error) {

	c, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var (
		items      = configItems()
		foundField bool
	)

	for i, line := range strings.Split(string(c), "\n") {
		field := yamlField.FindString(line)
		if field != "" {
			field = strings.TrimSuffix(strings.TrimSpace(field), ":")
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

// parse the local project YAML config
func parseProjectConfig() (*config, error) {

	projectConfigPath = zeusDir + "/zeus_config.yml"

	var c = new(config)

	stat, err := os.Stat(projectConfigPath)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, ErrConfigFileIsADirectory
	}

	contents, err := validateConfig(projectConfigPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(contents, c)
	if err != nil {
		Log.WithError(err).Fatal("failed to unmarshal confg - invalid YAML:")
		printFileContents(contents)
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

	b, err := yaml.Marshal(conf)
	if err != nil {
		Log.WithError(err).Fatal("failed to marshal config YAM:")
	}

	if _, err := os.Stat(zeusDir); err != nil {
		err = os.Mkdir(zeusDir, 0700)
		if err != nil {
			Log.WithError(err).Fatal("failed to create zeusDir")
		}
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

// remove formatter event
func cleanFormatterEvent() string {

	var id string
	projectDataMutex.Lock()
	for _, e := range projectData.Events {
		if e.Name == "formatter watcher" {
			id = e.ID
		}
	}
	projectDataMutex.Unlock()

	if id != "" {
		removeEvent(id)
	}

	return id
}

// watch and reload config on changes
func (c *config) watch(eventID string) {

	// dont add a new watcher when the event exists
	projectDataMutex.Lock()
	for _, e := range projectData.Events {
		if e.Name == "config watcher" {
			projectDataMutex.Unlock()
			return
		}
	}
	projectDataMutex.Unlock()

	Log.Debug("watching config at " + projectConfigPath)

	err := addEvent(newEvent(projectConfigPath, fsnotify.Write, "config watcher", ".yml", eventID, "internal", func(event fsnotify.Event) {

		Log.Debug("config watcher event: ", event.Name)

		b, err := validateConfig(projectConfigPath)
		if err != nil {
			Log.WithError(err).Fatal("failed to read config")
		}

		configMutex.Lock()
		err = yaml.Unmarshal(b, c)
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
	case reflect.String:
		return "field type: " + f.Kind().String() + ", value: " + f.String()
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
	case reflect.String:
		f.SetString(f.String())
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

	// this produces a data race
	// if c.Debug {
	// 	Log.Level = logrus.DebugLevel
	// } else {
	// 	Log.Level = logrus.InfoLevel
	// }

	// enable dumping the script on error when the auto formatter is enabled
	if c.AutoFormat {
		c.DumpScriptOnError = true
	}

	// this produces a data race
	// Log.Lock()
	// Log = logrus.New()

	// // disable colors if requested
	// if !c.Colors {
	// 	Log.Formatter = &prefixed.TextFormatter{
	// 		DisableColors: true,
	// 	}
	// } else {
	// 	Log.Formatter = &prefixed.TextFormatter{}
	// }

	if !c.AutoFormat {
		cleanFormatterEvent()
	}
}

// print the current configuration as JSON to stdout
func printConfiguration() {

	configMutex.Lock()
	defer configMutex.Unlock()

	l.Println()

	b, err := yaml.Marshal(conf)
	if err != nil {
		Log.WithError(err).Fatal("failed to marshal config to JSON")
	}

	l.Println(string(b))
}
