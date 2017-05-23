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
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"gopkg.in/yaml.v2"
)

var (
	// ErrConfigFileIsADirectory means the config file is a directory, thats wrong
	ErrConfigFileIsADirectory = errors.New("the config file is a directory")

	// path for project config file
	projectConfigPath string

	// path for project config files
	zeusDir = "zeus"

	// path for command scripts
	scriptDir = zeusDir + "/scripts"

	// regex for matching YAML keys from config file contents
	configYamlField = regexp.MustCompile("^(\\s)*[A-Z]+(.|\\s)*:")

	// regex for matching YAML keys from script header
	yamlField = regexp.MustCompile("^(\\s)*[a-z]+(.|\\s)*:")
)

// config contains configurable parameters
type config struct {
	fields *configFields
	sync.RWMutex
}

type configFields struct {
	AutoFormat          bool                     `yaml:"AutoFormat"`
	FixParseErrors      bool                     `yaml:"FixParseErrors"`
	Colors              bool                     `yaml:"Colors"`
	PassCommandsToShell bool                     `yaml:"PassCommandsToShell"`
	WebInterface        bool                     `yaml:"WebInterface"`
	Interactive         bool                     `yaml:"Interactive"`
	Debug               bool                     `yaml:"Debug"`
	RecursionDepth      int                      `yaml:"RecursionDepth"`
	ProjectNamePrompt   bool                     `yaml:"ProjectNamePrompt"`
	ColorProfile        string                   `yaml:"ColorProfile"`
	HistoryFile         bool                     `yaml:"HistoryFile"`
	HistoryLimit        int                      `yaml:"HistoryLimit"`
	PortWebPanel        int                      `yaml:"PortWebPanel"`
	PortGlueServer      int                      `yaml:"PortGlueServer"`
	ExitOnInterrupt     bool                     `yaml:"ExitOnInterrupt"`
	DisableTimestamps   bool                     `yaml:"DisableTimestamps"`
	PrintBuiltins       bool                     `yaml:"PrintBuiltins"`
	MakefileOverview    bool                     `yaml:"MakefileOverview"`
	StopOnError         bool                     `yaml:"StopOnError"`
	DumpScriptOnError   bool                     `yaml:"DumpScriptOnError"`
	DateFormat          string                   `yaml:"DateFormat"`
	TodoFilePath        string                   `yaml:"TodoFilePath"`
	Editor              string                   `yaml:"Editor"`
	ColorProfiles       map[string]*ColorProfile `yaml:"ColorProfiles"`
	Languages           []*Language              `yaml:"Languages"`
}

// newConfig returns the default configuration in case there is no config file
func newConfig() *config {
	return &config{
		fields: &configFields{
			MakefileOverview:    false,
			AutoFormat:          false,
			FixParseErrors:      true,
			Colors:              true,
			PassCommandsToShell: true,
			WebInterface:        false,
			Interactive:         true,
			Debug:               false,
			RecursionDepth:      1,
			ProjectNamePrompt:   true,
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
			Editor:       "micro",
			ColorProfile: "default",
			ColorProfiles: map[string]*ColorProfile{
				"light": lightProfile(),
				"dark":  darkProfile(),
			},
		},
	}
}

func printConfigUsageErr() {
	l.Println(ErrInvalidUsage)
	l.Println("usage: config [get <field>] [set <field> <value>]")
}

// check for unknown fields in the config
// since YAML simply ignores them and intializes them with their default values
func validateConfig(path string) ([]byte, error) {

	c, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var (
		items        = configItems()
		parsedFields []string
		foundField   bool
	)

	for i, line := range strings.Split(string(c), "\n") {
		field := configYamlField.FindString(line)
		if field != "" && !strings.HasPrefix(field, "    ") {
			field = strings.TrimSuffix(strings.TrimSpace(field), ":")
			for _, item := range items {
				if field == strings.TrimSpace(string(item.GetName())) {
					for _, f := range parsedFields {
						if f == field {
							Log.Warn("line " + strconv.Itoa(i) + ": duplicate config field: " + field)
						}
					}
					parsedFields = append(parsedFields, field)
					foundField = true
				}
			}
			if !foundField && field != "rwmutex" && field != "ColorProfiles" && field != "Languages" {
				Log.Warn("line "+strconv.Itoa(i)+": unknown config field: ", field)
			}
			foundField = false
		}
	}

	return c, nil
}

// parse the local project YAML config
func parseProjectConfig() (*config, error) {

	projectConfigPath = zeusDir + "/config.yml"

	// init default config
	var c = newConfig()

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

	err = yaml.Unmarshal(contents, c.fields)
	if err != nil {
		Log.WithError(err).Fatal("failed to unmarshal confg - invalid YAML:")
		printFileContents(contents)
		return nil, err
	}

	c.handle()

	return c, nil
}

// handle config shell command
func handleConfigCommand(args []string) {

	if len(args) < 2 {
		conf.dump()
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

	c.Lock()
	defer c.Unlock()

	b, err := yaml.Marshal(c.fields)
	if err != nil {
		Log.WithError(err).Fatal("failed to marshal config YAML:")
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
	projectData.Lock()
	for _, e := range projectData.fields.Events {
		if e.Name == "formatter watcher" {
			id = e.ID
		}
	}
	projectData.Unlock()

	if id != "" {
		removeEvent(id)
	}

	return id
}

// watch and reload config on changes
func (c *config) watch(eventID string) {

	// dont add a new watcher when the event exists
	projectData.Lock()
	for _, e := range projectData.fields.Events {
		if e.Name == "config watcher" {
			projectData.Unlock()
			return
		}
	}
	projectData.Unlock()

	Log.Debug("watching config at " + projectConfigPath)

	err := addEvent(newEvent(projectConfigPath, fsnotify.Write, "config watcher", ".yml", eventID, "internal", func(event fsnotify.Event) {

		Log.Debug("config watcher event: ", event.Name)

		b, err := validateConfig(projectConfigPath)
		if err != nil {
			Log.WithError(err).Error("failed to read config")
			return
		}

		// lock config
		c.Lock()

		err = yaml.Unmarshal(b, c.fields)
		if err != nil {
			Log.WithError(err).Error("config parse error")
			c.Unlock()
			return
		}
		c.Unlock()

		// handle updated values
		c.handle()
	}))
	if err != nil {
		Log.WithError(err).Fatal("projectConfig watcher failed")
	}
}

// get type and current vlaue information for a given field on the config struct
func (c *config) getFieldInfo(field string) string {

	c.Lock()
	defer c.Unlock()

	f := reflect.Indirect(reflect.ValueOf(c.fields)).FieldByName(field)
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

	c.Lock()

	// check if the named field exists on the struct
	f := reflect.Indirect(reflect.ValueOf(c.fields)).FieldByName(field)

	c.Unlock()

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
		f.SetString(value)
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

	// lock config
	c.Lock()
	defer c.Unlock()

	// lock Logger
	Log.Lock()
	defer Log.Unlock()

	// handle debug mode
	// @TODO: this produces a data race
	// if c.Debug {
	// 	Log.Level = logrus.DebugLevel
	// } else {
	// 	Log.Level = logrus.InfoLevel
	// }

	// enable dumping the script on error when the auto formatter is enabled
	if c.fields.AutoFormat {
		c.fields.DumpScriptOnError = true
	}

	// disable colors if requested
	if !c.fields.Colors {

		// lock once
		cp.Lock()
		cp = colorsOffProfile().parse()

		Log.Formatter = &prefixed.TextFormatter{
			DisableColors:    true,
			DisableTimestamp: c.fields.DisableTimestamps,
		}
	}

	if !c.fields.AutoFormat {
		cleanFormatterEvent()
	}

	ps.Lock()
	defer ps.Unlock()

	// overwrite default languages with those from config
	for _, lang := range c.fields.Languages {
		ps.items[lang.Name] = newParser(lang)
	}
}

// print the current configuration as YAML to stdout
func (c *config) dump() {

	c.Lock()
	defer c.Unlock()

	l.Println()

	b, err := yaml.Marshal(c.fields)
	if err != nil {
		Log.WithError(err).Error("failed to marshal config to YAML")
	} else {
		l.Println(string(b))
	}
}
