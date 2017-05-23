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
}

func bashLanguage() *Language {
	return &Language{
		Name:               "bash",
		Interpreter:        "/bin/bash",
		Bang:               "#!/bin/bash",
		Comment:            "#",
		AssignmentOperator: "=",
		FlagStopOnError:    "-e",
		FlagEvaluateScript: "-c",
		FileExtension:      ".sh",
	}
}

func pythonLanguage() *Language {
	return &Language{
		Name:               "python",
		Interpreter:        "/usr/bin/python",
		Bang:               "#!/usr/bin/python",
		Comment:            "#",
		AssignmentOperator: " = ",
		FlagEvaluateScript: "-c",
		FileExtension:      ".py",
		ExecOpPrefix:       "import os; os.system(\"",
		ExecOpSuffix:       "\")",
	}
}

func javaScriptLanguage() *Language {
	return &Language{
		Name:               "javascript",
		Interpreter:        "/usr/bin/osascript",
		Bang:               "#!/usr/bin/osascript -l JavaScript",
		Comment:            "//",
		AssignmentOperator: " = ",
		VariableKeyword:    "var ",
		UseTempFile:        true,
		FileExtension:      ".js",
		ExecOpPrefix:       "ObjC.import('stdlib'); $.system(\"",
		ExecOpSuffix:       "\");",
	}
}

func rubyLanguage() *Language {
	return &Language{
		Name:               "ruby",
		Interpreter:        "/usr/bin/ruby",
		Bang:               "#!/usr/bin/ruby",
		Comment:            "#",
		AssignmentOperator: " = ",
		VariableKeyword:    "$",
		FlagEvaluateScript: "-e",
		FileExtension:      ".rb",
		ExecOpPrefix:       "`",
		ExecOpSuffix:       "`",
	}
}
