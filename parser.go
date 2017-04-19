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
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
)

var (
	// regex for a VALID zeus header field
	validzeusHeaderField = regexp.MustCompile(`#+[[:space:]]+@zeus(-([a-z]+))+:+`)

	// regex for an INVALID zeus header field
	invalidzeusHeaderField = regexp.MustCompile("#*[[:space:]]*@*zeus-([a-z]+):*[[:space:]]*")

	// ErrDuplicateFields means a header field appeared twice
	ErrDuplicateFields = errors.New("duplicate zeus header fields")

	// ErrDuplicateArgumentNames means the name for an argument was reused
	ErrDuplicateArgumentNames = errors.New("duplicate argument name")

	// ErrInvalidHeaderType means the header field type does not exist
	ErrInvalidHeaderType = errors.New("invalid header field type")
)

// parser handles parsing of the script headers
// it contains syntactic language elements
// and manages a concurrently executed pool of jobs
// synchronization is provided through locking access to the job pool
type parser struct {

	// scripting language to use
	language string

	// path to interpreter
	interpreter string

	// identifier for script type
	bang string

	// comment identifier
	comment string

	// available header fields
	zeusFieldChain        string
	zeusFieldHelp         string
	zeusFieldArgs         string
	zeusFieldBuildNumber  string
	zeusFieldDependencies string
	zeusFieldOutputs      string

	// separator for build chain commands
	separator string

	// jobs
	jobs map[string]*parseJob

	// locking for map access
	mutex sync.Mutex
	// inputChannel chan string

	// limit for recursion level
	recursionDepth int
}

// create a new parser instance
func newParser() *parser {

	return &parser{
		language:    "bash",
		interpreter: "/bin/bash",
		bang:        "#!/bin/bash",
		comment:     "#",

		zeusFieldChain:        "zeus-chain",
		zeusFieldHelp:         "zeus-help",
		zeusFieldArgs:         "zeus-args",
		zeusFieldBuildNumber:  "zeus-build-number",
		zeusFieldDependencies: "zeus-deps",
		zeusFieldOutputs:      "zeus-outputs",

		separator:      "->",
		jobs:           map[string]*parseJob{},
		mutex:          sync.Mutex{},
		recursionDepth: 1,
	}
}

// commandData represents the information retrieved by parsing a command script
type commandData struct {
	// Help text
	Help string `yaml:"help"`

	// Args for command
	Args string `yaml:"args"`

	// Chain contains the targets commandChain
	Chain string `yaml:"chain"`

	// Manual text
	Manual string `yaml:"manual"`

	// BuildNumber yes or no
	BuildNumber bool `yaml:"buildNumber"`

	// Dependencies for the current target
	Dependencies []string `yaml:"deps"`

	// Outputs of the current target
	Outputs []string `yaml:"out"`

	// Run when executed
	Run string `yaml:"run"`
}

// argument types
const (
	argTypeString = "String"
	argTypeInt    = "Int"
	argTypeBool   = "Bool"
	argTypeFloat  = "Float"
)

// a command argument has a name and a type and a value
type commandArg struct {

	// argument label
	name string

	// argument type
	argType reflect.Kind

	// optionals are allowed, they can have default values
	optional     bool
	defaultValue string

	// value after parsing argument input from commandline
	value string
}

// parse script and return commandData
// a zeus header looks like this:
// # ----------------------------------------------------------------------------------- #
// # @zeus-chain: command1 arg1 arg2 -> command2 -> command3
// # @zeus-help: help text for command
// # @zeus-args: arg1 arg2 arg3
// # ----------------------------------------------------------------------------------- #
// # manual entry text
// # ----------------------------------------------------------------------------------- #
func (p *parser) parseScript(path string, job *parseJob) (*commandData, error) {

	var (
		cLog = Log.WithFields(logrus.Fields{
			"prefix": "parseScript",
			"path":   path,
		})

		commandName     = strings.TrimSuffix(filepath.Base(path), ".sh")
		helpFieldCount  = 0
		argsFieldCount  = 0
		chainFieldCount = 0
		separatorCount  int

		d = new(commandData)
	)

	// get file contents
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return d, err
	}

	// range line by line
	for c, line := range strings.Split(string(contents), "\n") {

		if c == 0 {
			// first line. make sure theres a shebang
			if line != p.bang {
				if conf.FixParseErrors {
					sanitizeFile(path)
					return p.parseScript(path, job)
				}
				Log.Fatal("first line does not contain a shebang.")
			}
		}

		// check if its a comment. only comments can be used as header fields
		if strings.HasPrefix(line, p.comment) {

			if separatorCount > 1 {

				// check if its the closing tag
				if strings.HasPrefix(line, "# --------------------") {
					return d, nil
				}

				d.Manual += line + "\n"
				continue
			}

			switch true {

			// parse help field
			case strings.Contains(line, p.zeusFieldHelp):

				helpFieldCount++

				if !validzeusHeaderField.MatchString(line) {
					if conf.FixParseErrors {
						sanitizeFile(path)
						return p.parseScript(path, job)
					}
					Log.Fatal("invalid zeus-help header field in line ", c, " : ", line)
				}
				d.Help = strings.TrimSpace(trimZeusPrefix(line))
				break

			// parse args field
			case strings.Contains(line, p.zeusFieldArgs):

				argsFieldCount++

				if !validzeusHeaderField.MatchString(line) {
					if conf.FixParseErrors {
						sanitizeFile(path)
						return p.parseScript(path, job)
					}
					Log.Fatal("invalid zeus-args header field in line ", c, " : ", line)
				}

				d.Args = strings.TrimSpace(trimZeusPrefix(line))

				break

			// parse chain field
			case strings.Contains(line, p.zeusFieldChain):

				chainFieldCount++

				if !validzeusHeaderField.MatchString(line) {
					if conf.FixParseErrors {
						sanitizeFile(path)
						return p.parseScript(path, job)
					}
					Log.Fatal("invalid zeus-chain header field in line ", c, " : ", line)
				}

				d.Chain = strings.TrimSpace(trimZeusPrefix(line))

				break

			// parse multiline help entry (20 dashes minimum)
			case strings.HasPrefix(line, "# --------------------"):
				separatorCount++

			case strings.Contains(line, p.zeusFieldBuildNumber):
				d.BuildNumber = true

			case strings.Contains(line, p.zeusFieldDependencies):
				for _, dep := range strings.Split(strings.TrimSpace(trimZeusPrefix(line)), ",") {
					d.Dependencies = append(d.Dependencies, dep)
				}

			case strings.Contains(line, p.zeusFieldOutputs):
				for _, output := range strings.Split(strings.TrimSpace(trimZeusPrefix(line)), ",") {
					d.Outputs = append(d.Outputs, output)
				}

			default:

				// check if line might be a zeus header field
				if strings.HasPrefix(line, p.comment+" @") {
					// check if its a header field that does not exist
					if !validHeaderType(line) {
						Log.WithError(ErrInvalidHeaderType).WithFields(logrus.Fields{
							"line": line,
							"file": path,
						}).Fatal("invalid header field type")
					}
				}

				continue
			}
		}
	}

	// check for duplicate fields
	if argsFieldCount > 1 || chainFieldCount > 1 || helpFieldCount > 1 {
		cLog.WithFields(logrus.Fields{
			"argsFieldCount":  argsFieldCount,
			"chainFieldCount": chainFieldCount,
			"helpFieldCount":  helpFieldCount,
			"commandName":     commandName,
		}).Error()
		return d, ErrDuplicateFields
	}

	return d, nil
}

func validHeaderType(line string) bool {

	slice := strings.Split(line, ":")
	if len(slice) == 0 {
		// not a zeus header field - ignore
		return true
	}

	fieldType := strings.TrimPrefix(slice[0], p.comment+" @")

	Log.Info("checking type: ", fieldType)

	switch fieldType {
	case p.zeusFieldArgs, p.zeusFieldBuildNumber, p.zeusFieldChain, p.zeusFieldDependencies, p.zeusFieldHelp, p.zeusFieldOutputs:
		return true
	default:
		return false
	}

}

// parse the command chain string
func parseCommandChain(line string) (parsedCommands [][]string) {

	var (
		// trim whitespace and zeus prefix
		// then get commands separated by parser separator
		cmds = strings.Split(strings.TrimSpace(trimZeusPrefix(line)), p.separator)

		cLog = Log.WithFields(logrus.Fields{
			"prefix": "parseCommandChain",
			"cmds":   cmds,
		})
	)

	cLog.Debug("starting to parse command chain")

	if len(cmds) > 0 {

		// when there are no commands specified, the resulting slice from the split contains 1 empty string
		if cmds[0] == "" {
			return
		}

		// range them
		for _, name := range cmds {

			// get arguments for commands
			var args = strings.Fields(name)

			if len(args) == 0 {
				Log.Fatal(ErrEmptyName)
			}

			parsedCommands = append(parsedCommands, args)

			cLog.WithFields(logrus.Fields{
				"command": args[0],
				"args":    args,
			}).Debug("found command")
		}
	}

	return
}

// trim the zeus prefix from the beginning of a line
func trimZeusPrefix(line string) string {
	return string(validzeusHeaderField.ReplaceAll([]byte(line), []byte("")))
}
