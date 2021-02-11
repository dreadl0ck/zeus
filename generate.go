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
	"path/filepath"
	"strings"
	"time"
)

// Generate a standalone version of a single command or commandChain.
// If all commands are of the same language, generate a single script,
// if there are multiple scripting languages involved, generate a directory with all required scripts.
func handleGenerateCommand(args []string) {

	if len(args) < 3 {
		printGenerateCommandUsageErr()
		return
	}

	var (
		chain      commandChain
		ok         bool
		mixed      bool
		language   string
		arguments  []string
		outputName = zeusDir + "/generated/" + args[1]
	)

	// check if its a valid command chain
	if chain, ok = validCommandChain(args[2:], false); !ok {
		l.Println("invalid command chain")
		return
	}

	// make sure generated dir exists
	os.Mkdir(zeusDir+"/generated", 0744)

	// check if it contains multiple languages
	for _, cmd := range chain {
		if language == "" {
			language = cmd.language
		}
		if language != cmd.language {
			mixed = true
			break
		}
	}

	if mixed {

		// create a directory for multiple scripts
		err := os.Mkdir(outputName, 0744)
		if err != nil {
			Log.WithError(err).Error("failed to create directory for mixed commandChain")
			return
		}

		f, err := os.Create(outputName + "/run.sh")
		if err != nil {
			Log.WithError(err).Error("failed to create run script")
			return
		}

		err = f.Chmod(0700)
		if err != nil {
			Log.WithError(err).Error("failed to make run script executable")
			return
		}

		f.WriteString("#!/bin/bash\n\n")
		f.WriteString("# generated by ZEUS v" + version + " @ " + time.Now().String() + "\n")
		f.WriteString("# commandChain: " + strings.Join(args[2:], " ") + "\n")
		f.WriteString("\n./" + filepath.Base(chain[0].path) + "\n")
		f.Close()

		l.Println("generated " + outputName + "/run.sh")
	}

	chainSlice := strings.Split(strings.Join(args[3:], " "), commandChainSeparator)
	if len(chainSlice) > 0 {
		arguments = strings.Fields(chainSlice[0])
	}

	// iterate over the chain and generate scripts
	for i, cmd := range chain {
		f, lang, err := generateScript(outputName, mixed, cmd, arguments)
		if err != nil {
			l.Println(err)
			return
		}

		// inject call to next script
		if i < len(chain)-1 {

			nextCmd := chain[i+1]

			nextLang, err := nextCmd.getLanguage()
			if err != nil {
				l.Println(nextCmd.name + ": " + err.Error() + ": " + nextCmd.language)
				return
			}

			f.WriteString("\n" + lang.Comment + " execute next script: " + filepath.Base(nextCmd.path) + "\n")
			f.WriteString(lang.ExecOpPrefix + nextLang.Interpreter + " " + filepath.Base(nextCmd.path) + lang.ExecOpSuffix + "\n")
		}
		f.Close()
	}
}

// generate a single script
// if mixed each script will be put into the outputName directory
// if not, a single script file will be generated
// returns the fileDescriptor, language and an error
func generateScript(outputName string, mixed bool, cmd *command, args []string) (*os.File, *Language, error) {

	var arguments string

	// get language for current command
	lang, err := cmd.getLanguage()
	if err != nil {
		cmd.dump()
		return nil, nil, errors.New(cmd.name + ": " + err.Error() + ": " + cmd.language)
	}

	// handle arguments
	if len(cmd.args) > 0 {
		arguments, err = cmd.parseArguments(args)
		if err != nil {
			return nil, nil, err
		}
	}

	if mixed {
		// create new file in output directory
		outputName = outputName + "/" + cmd.name + lang.FileExtension
	}

	f, err := os.Create(outputName)
	if err != nil {
		return nil, nil, err
	}

	header := lang.Comment + " generated by ZEUS v" + version + "\n"
	header += lang.Comment + " Timestamp: " + time.Now().Format(timestampFormat) + "\n"

	// insert bang and args
	f.WriteString(lang.Bang + "\n" + header + "\n" + generateGlobals(lang) + "\n" + arguments + "\n")

	// add language specific global code
	code, err := ioutil.ReadFile(zeusDir + "/globals/globals" + lang.FileExtension)
	if err == nil {
		f.WriteString("\n")
		f.Write(code)
		f.WriteString("\n")
	}

	// add dependencies
	err = addDependencies(cmd, f, lang, outputName, mixed)
	if err != nil {
		return nil, nil, err
	}

	// add script of current command
	if cmd.exec == "" {
		c, err := ioutil.ReadFile(cmd.path)
		if err != nil {
			l.Println("failed to read: " + cmd.path)
			return nil, nil, err
		}
		f.Write(c)
	} else {
		f.WriteString(cmd.exec)
	}

	// make script executable
	err = os.Chmod(outputName, 0700)
	if err != nil {
		Log.Error("failed to make script executable")
		return nil, nil, err
	}

	l.Println("generated " + outputName)

	return f, lang, nil
}

func addDependencies(cmd *command, f *os.File, lang *Language, outputName string, mixed bool) error {

	for i, d := range cmd.dependencies {

		fields := strings.Fields(d)
		if len(fields) == 0 {
			return ErrEmptyDependency
		}

		// lookup
		dep, err := cmdMap.getCommand(fields[0])
		if err != nil {
			return errors.New("invalid dependency: " + err.Error())
		}

		// get language for current command
		depLang, err := dep.getLanguage()
		if err != nil {
			l.Println(dep.name+":", err, ", language:", dep.language)
			return err
		}

		if lang.Name == depLang.Name {
			// add to current script
			if dep.exec == "" {
				c, err := ioutil.ReadFile(dep.path)
				if err != nil {
					l.Println("failed to read: " + dep.path)
					return err
				}
				f.Write(c)
			} else {
				f.WriteString(dep.exec)
			}
		} else {
			// dependency script is in another language
			// generate a new script and inject a call to it
			generateScript(outputName, mixed, dep, []string{})

			fields := strings.Fields(cmd.dependencies[i+1])
			if len(fields) == 0 {
				return ErrEmptyDependency
			}

			// lookup
			nextDep, err := cmdMap.getCommand(fields[0])
			if err != nil {
				return err
			}

			f.WriteString("\n" + lang.Comment + " execute next script: " + filepath.Base(nextDep.path) + "\n")
			f.WriteString(lang.ExecOpPrefix + lang.Interpreter + " " + filepath.Base(nextDep.path) + lang.ExecOpSuffix + "\n")
		}

		f.WriteString("\n")
	}

	return nil
}
