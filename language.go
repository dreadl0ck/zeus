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
	"sync"
)

var (
	commandChainSeparator = "->"

	// global language store
	ls = &languageStore{
		items: map[string]*Language{
			"bash":       bashLanguage(),
			"python":     pythonLanguage(),
			"javascript": javaScriptLanguage(),
			"ruby":       rubyLanguage(),
			"lua":        luaLanguage(),
			"sh":         shellLanguage(),
			"zsh":        zshellLanguage(),
			"perl":       perlLanguage(),
		},
	}

	// ErrUnknownLanguage means there's no syntax definition for the desired language
	ErrUnknownLanguage = errors.New("unknown language")
)

// thread safe store for all languages
type languageStore struct {
	items map[string]*Language
	sync.Mutex
}

func (langStore *languageStore) getLang(name string) (*Language, error) {

	langStore.Lock()
	defer langStore.Unlock()

	if lang, ok := langStore.items[name]; ok {

		// return language instance
		return lang, nil
	}
	return nil, ErrUnknownLanguage
}

// Language describes interpreter and syntactic elements of a scripting language
type Language struct {
	Name string `yaml:"name"`

	// path to Interpreter
	Interpreter string `yaml:"interpreter"`

	// identifier for script type
	Bang string `yaml:"bang"`

	// Comment identifier
	Comment string `yaml:"comment"`

	// prefix when declaring variables i.e. 'var' keyword
	VariableKeyword string `yaml:"variableKeyword"`

	// assigment operator i.e. '='
	AssignmentOperator string `yaml:"assignmentOperator"`

	// line delimiter i.e. ';'
	LineDelimiter string `yaml:"lineDelimiter"`

	// flag for stopping execution after an error
	FlagStopOnError string `yaml:"flagStopOnError"`

	// flag for passing a script on the commandline
	FlagEvaluateScript string `yaml:"flagEvaluateScript"`

	// some interpreters (i.e. osascript) don't allow passing a multiline script for evaluation on the commandline
	// in this case a temporary script is generated on disk and passed to the interpreter for execution
	UseTempFile bool `yaml:"useTempFile"`

	ExecOpPrefix string `yaml:"execOpPrefix"`
	ExecOpSuffix string `yaml:"execOpSuffix"`

	// extension for filetype
	FileExtension string `yaml:"fileExtension"`

	CorrectErrLineNumber bool   `yaml:"correctErrLineNumber"`
	ErrLineNumberSymbol  string `yaml:"errLineNumberSymbol"`
}

func bashLanguage() *Language {
	return &Language{
		Name:                 "bash",
		Interpreter:          "/bin/bash",
		Bang:                 "#!/bin/bash",
		Comment:              "#",
		AssignmentOperator:   "=",
		FlagStopOnError:      "-e",
		FlagEvaluateScript:   "-c",
		FileExtension:        ".sh",
		CorrectErrLineNumber: false,
		ErrLineNumberSymbol:  "line",
	}
}

func shellLanguage() *Language {
	return &Language{
		Name:                 "sh",
		Interpreter:          "/bin/sh",
		Bang:                 "#!/bin/sh",
		Comment:              "#",
		AssignmentOperator:   "=",
		FlagStopOnError:      "-e",
		FlagEvaluateScript:   "-c",
		FileExtension:        ".sh",
		CorrectErrLineNumber: false,
		ErrLineNumberSymbol:  "line",
	}
}

func zshellLanguage() *Language {
	return &Language{
		Name:                 "zsh",
		Interpreter:          "/bin/zsh",
		Bang:                 "#!/bin/zsh",
		Comment:              "#",
		AssignmentOperator:   "=",
		FlagStopOnError:      "-e",
		FlagEvaluateScript:   "-c",
		FileExtension:        ".zsh",
		CorrectErrLineNumber: false,
		ErrLineNumberSymbol:  "", // TODO: no symbol for that, allow to use a regex for this task
	}
}

func pythonLanguage() *Language {
	return &Language{
		Name:                 "python",
		Interpreter:          "/usr/bin/python",
		Bang:                 "#!/usr/bin/python",
		Comment:              "#",
		AssignmentOperator:   " = ",
		FlagEvaluateScript:   "-c",
		FileExtension:        ".py",
		ExecOpPrefix:         "import os; os.system(\"",
		ExecOpSuffix:         "\")",
		CorrectErrLineNumber: true,
		ErrLineNumberSymbol:  "line",
	}
}

func javaScriptLanguage() *Language {
	return &Language{
		Name:                 "javascript",
		Interpreter:          "/usr/bin/osascript",
		Bang:                 "#!/usr/bin/osascript -l JavaScript",
		Comment:              "//",
		AssignmentOperator:   " = ",
		VariableKeyword:      "var ",
		UseTempFile:          true,
		FileExtension:        ".js",
		ExecOpPrefix:         "ObjC.import('stdlib'); $.system(\"",
		ExecOpSuffix:         "\");",
		CorrectErrLineNumber: false,
		ErrLineNumberSymbol:  "line",
	}
}

func rubyLanguage() *Language {
	return &Language{
		Name:                 "ruby",
		Interpreter:          "/usr/bin/ruby",
		Bang:                 "#!/usr/bin/ruby",
		Comment:              "#",
		AssignmentOperator:   " = ",
		VariableKeyword:      "$",
		FlagEvaluateScript:   "-e",
		FileExtension:        ".rb",
		ExecOpPrefix:         "`",
		ExecOpSuffix:         "`",
		CorrectErrLineNumber: true,
		ErrLineNumberSymbol:  "-e:",
	}
}

func luaLanguage() *Language {
	return &Language{
		Name:        "lua",
		Interpreter: "/usr/local/bin/lua",
		//Bang:               "#!/usr/local/bin/lua",
		Comment:              "--",
		AssignmentOperator:   " = ",
		VariableKeyword:      "local ",
		FlagEvaluateScript:   "-e",
		FileExtension:        ".lua",
		ExecOpPrefix:         "os.execute(\"",
		ExecOpSuffix:         "\")",
		CorrectErrLineNumber: true,
		ErrLineNumberSymbol:  "line",
	}
}

func perlLanguage() *Language {
	return &Language{
		Name:                 "perl",
		Interpreter:          "/usr/bin/perl",
		Bang:                 "#!/usr/bin/perl",
		Comment:              "#",
		AssignmentOperator:   " = ",
		LineDelimiter:        ";",
		VariableKeyword:      "$",
		FlagEvaluateScript:   "-e",
		FileExtension:        ".pl",
		ExecOpPrefix:         "system(\"",
		ExecOpSuffix:         "\")",
		CorrectErrLineNumber: true,
		ErrLineNumberSymbol:  "line",
	}
}
