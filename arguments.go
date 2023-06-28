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
	"bytes"
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/mgutz/ansi"
)

var (
	// ErrDuplicateArgumentNames means the name for an argument was reused
	ErrDuplicateArgumentNames = errors.New("duplicate argument name")
)

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

// validate arguments string from CommandsFile
// and return the validatedArgs as map
func (c *CommandsFile) validateArgs(args []string) ([]*commandArg, error) {

	// init map
	var validatedArgs []*commandArg

	// empty string - empty args
	if len(args) == 0 {
		return nil, nil
	}

	// parse arg string
	for i, s := range args {

		if len(s) == 0 {
			return nil, errors.New("found empty argument at index: " + strconv.Itoa(i))
		}

		var (
			k            reflect.Kind
			slice        = strings.Split(s, ":")
			opt          bool
			defaultValue string
		)

		if len(slice) > 1 {

			// argument name may contain leading whitespace - trim it
			var argumentName = strings.TrimSpace(slice[0])

			// check for name conflicts with globals
			g.Lock()
			for name := range g.Vars {
				if argumentName == name {
					g.Unlock()
					listGlobals()
					return nil, errors.New("argument name " + argumentName + " conflicts with a global variable")
				}
			}
			g.Unlock()

			// check for duplicate argument names
			if a, ok := containsArg(validatedArgs, argumentName); ok {
				Log.Error("argument label ", a.name, " was used twice")
				return nil, ErrDuplicateArgumentNames
			}

			// check if there's a default value set
			defaultValSlice := strings.Split(slice[1], "=")
			if len(defaultValSlice) > 1 {
				if !strings.Contains(slice[1], "?") {
					return nil, errors.New("default values for mandatory arguments are not allowed: " + s + ", at index: " + strconv.Itoa(i))
				}
				slice[1] = strings.TrimSpace(defaultValSlice[0])
				defaultValue = defaultValSlice[1]
			}

			// check if its an optional arg
			if strings.HasSuffix(slice[1], "?") {
				slice[1] = strings.TrimSuffix(slice[1], "?")
				opt = true
			}

			// check if its a valid argType and set reflect.Kind
			switch slice[1] {
			case argTypeBool:
				k = reflect.Bool
			case argTypeFloat:
				k = reflect.Float64
			case argTypeString:
				k = reflect.String
			case argTypeInt:
				k = reflect.Int
			default:
				return nil, errors.New("invalid or missing argument type: " + s)
			}

			// add to validatedArgs
			validatedArgs = append(validatedArgs, &commandArg{
				name:         argumentName,
				argType:      k,
				optional:     opt,
				defaultValue: defaultValue,
			})
		} else {
			return nil, errors.New("invalid argument declaration: " + s)
		}
	}

	return validatedArgs, nil
}

// parse arguments array in the label=value format
// and return a code snippet that declares them in the language of the command
func (c *command) parseArguments(args []string) (string, map[string]string, error) {

	var (
		argBuf      bytes.Buffer
		occurrences = make(map[string]int, 0)

		// In order to be able to use argument values in command output names,
		// we need to return the current state of arguments.
		// It's not possible to use the c.args map for that, as these are shared in between multiple invocations of the same command,
		// and the argument values must  be determined again for every invocation.
		// For that reason, the argument names mapped to their current values will be returned as a separate map to the caller.
		argValues = map[string]string{}
	)

	// parse args
	for _, val := range args {

		// handle argument labels
		if strings.Contains(val, "=") {

			var (
				cmdArg *commandArg
				ok     bool
			)

			argSlice := strings.Split(val, "=")
			if len(argSlice) < 2 {
				return "", argValues, errors.New("invalid argument: " + val)
			}

			cmdArg, ok = containsArg(c.args, argSlice[0])
			if !ok {
				return "", argValues, errors.New(ErrInvalidArgumentLabel.Error() + ": " + ansi.Red + argSlice[0] + cp.Reset)
			}

			if _, ok := occurrences[argSlice[0]]; ok {
				occurrences[argSlice[0]]++
			} else {
				occurrences[argSlice[0]] = 1
			}

			if occurrences[argSlice[0]] > 1 {
				return "", argValues, errors.New("argument label appeared more than once: " + cmdArg.name)
			}

			argValue := strings.Join(argSlice[1:], "=")

			if err := validArgType(argValue, cmdArg.argType); err != nil {
				return "", argValues, errors.New(ErrInvalidArgumentType.Error() + ": " + err.Error() + ", label=" + cmdArg.name + ", value=" + argValue)
			}

			// temporarily set value on arg
			cmdArg.value = argValue
		} else {
			return "", argValues, errors.New("invalid argument: " + val)
		}
	}

	lang, err := c.getLanguage()
	if err != nil {
		return "", argValues, err
	}

	for _, arg := range c.args {
		if arg.value == "" {
			if arg.optional {
				if arg.defaultValue != "" {

					argValues[arg.name] = arg.defaultValue

					// default value has been set
					argBuf.WriteString(lang.VariableKeyword + arg.name + lang.AssignmentOperator + strings.TrimSpace(arg.defaultValue) + lang.LineDelimiter + "\n")
				} else {

					argValues[arg.name] = getDefaultValue(arg)

					// init empty optionals with default value for their type
					argBuf.WriteString(lang.VariableKeyword + arg.name + lang.AssignmentOperator + getDefaultValue(arg) + lang.LineDelimiter + "\n")
				}
			} else {
				// empty value and not optional - error
				return "", argValues, errors.New(c.name + " is missing an argument: " + ansi.Red + arg.name + ":" + strings.Title(arg.argType.String()) + cp.Reset)
			}
		} else {

			argValues[arg.name] = arg.value

			// write value into buffer
			argBuf.WriteString(lang.VariableKeyword + arg.name + lang.AssignmentOperator + arg.value + lang.LineDelimiter + "\n")
		}
	}

	// flush arg values before returning
	for _, arg := range c.args {
		arg.value = ""
	}

	return argBuf.String(), argValues, nil
}

func replaceArgs(input string, args map[string]string) (string, error) {
	var (
		dollar, startOfIdent bool
		name                 string
		names                []string
	)

	// replace variables used in ${} notation with global values
	for _, char := range input {
		if char == '$' {
			dollar = true
			continue
		}
		if dollar {
			if char == '{' {
				startOfIdent = true
				continue
			}
		}
		if char == '}' {
			// collect string
			n := name
			names = append(names, n)

			// reset state values
			name = ""
			dollar = false
			startOfIdent = false
		}
		if startOfIdent {
			name += string(char)
			continue
		}
	}

	for _, n := range names {
		if val, ok := args[n]; ok {
			input = strings.ReplaceAll(input, "${"+n+"}", val)
		} else {
			return "", errors.New(input + ": variable is not provided via globals or arguments: ${" + n + "}")
		}
	}

	return input, nil
}

func (c *CommandsFile) replaceGlobals(input string) string {
	var (
		dollar, startOfIdent bool
		name                 string
		names                []string
	)

	// replace variables used in ${} notation with global values
	for _, char := range input {
		if char == '$' {
			dollar = true
			continue
		}
		if dollar {
			if char == '{' {
				startOfIdent = true
				continue
			}
		}
		if char == '}' {
			// collect string
			n := name
			names = append(names, n)

			// reset state values
			name = ""
			dollar = false
			startOfIdent = false
		}
		if startOfIdent {
			name += string(char)
			continue
		}
	}

	for _, n := range names {
		if val, ok := g.Vars[n]; ok {
			input = strings.ReplaceAll(input, "${"+n+"}", val)
		}
	}

	return input
}

func resolveEnvironment(input string) string {
	var (
		dollar, startOfIdent bool
		name                 string
		names                []string
	)

	// replace variables used in ${} notation with global values
	for _, char := range input {
		if char == '$' {
			dollar = true
			continue
		}
		if dollar {
			if char == '{' {
				startOfIdent = true
				continue
			}
		}
		if char == '}' {
			// collect string
			n := name
			names = append(names, n)

			// reset state values
			name = ""
			dollar = false
			startOfIdent = false
		}
		if startOfIdent {
			name += string(char)
			continue
		}
	}

	for _, n := range names {
		input = strings.ReplaceAll(input, "${"+n+"}", os.Getenv(n))
	}

	return input
}

func containsArg(args []*commandArg, name string) (*commandArg, bool) {
	for _, a := range args {
		if a.name == name {
			return a, true
		}
	}
	return nil, false
}
