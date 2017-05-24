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
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	// zeusHeaderTag for embedding YAML in script headers
	zeusHeaderTag         = "{zeus}"
	commandChainSeparator = "->"

	// global parser store
	ps = &parserStore{
		items: map[string]*parser{
			"bash":       newParser(bashLanguage()),
			"python":     newParser(pythonLanguage()),
			"javascript": newParser(javaScriptLanguage()),
			"ruby":       newParser(rubyLanguage()),
			"lua":        newParser(luaLanguage()),
		},
	}

	// ErrDuplicateArgumentNames means the name for an argument was reused
	ErrDuplicateArgumentNames = errors.New("duplicate argument name")
)

// thread safe store for all parsers
type parserStore struct {
	items map[string]*parser
	sync.Mutex
}

// parser handles parsing of the script headers
// it contains syntactic language elements
// and manages a concurrently executed pool of jobs
// synchronization is provided through locking access to the job pool
type parser struct {

	// scripting language to use
	language *Language

	// job pool
	jobs map[parseJobID]*parseJob

	// locking for map access
	sync.RWMutex
}

// create a new bash parser instance
func newParser(lang *Language) *parser {
	return &parser{
		language: lang,
		jobs:     map[parseJobID]*parseJob{},
	}
}

// parse script and return commandData
func (p *parser) parseScript(path string, job *parseJob) (*commandData, error) {

	var (
		h             = new(commandData)
		foundStartTag bool
		buffer        bytes.Buffer
	)

	// get file handle
	f, err := os.Open(path)
	if err != nil {
		return h, err
	}

	// create reader
	// and read file contents line by line
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		switch {
		case err == io.EOF:

			// Warn if there no closing zeus tag
			if foundStartTag {
				Log.Error("couldn't find closing header tag in script " + path)
			}
			return h, nil
		case err != nil:
			return h, err
		}

		if strings.Contains(line, zeusHeaderTag) && foundStartTag {

			// trim trailing newline
			contents := bytes.TrimSuffix(buffer.Bytes(), []byte{'\n'})

			err = validateHeader(contents, path)
			if err != nil {
				return nil, err
			}

			// parse buffer and return
			err := yaml.Unmarshal(contents, h)
			if err != nil {
				printScript(buffer.String(), path)
				return h, err
			}

			return h, nil
		}
		if foundStartTag {
			buffer.WriteString(strings.TrimPrefix(strings.TrimPrefix(line, p.language.Comment), " "))
		}
		if strings.Contains(line, zeusHeaderTag) {
			foundStartTag = true
			continue
		}
	}
}

// get parser instance for script by name or by path
func getParserForScript(name string) (*parser, error) {

	// look at fileExtension
	ext := filepath.Ext(name)
	if ext != "" {
		ps.Lock()
		defer ps.Unlock()

		for _, p := range ps.items {
			if p.language.FileExtension == ext {
				return p, nil
			}
		}
		return nil, ErrUnsupportedLanguage
	}

	cmdMap.Lock()
	defer cmdMap.Unlock()

	if cmd, ok := cmdMap.items[name]; ok {

		// check if language is set on command
		if cmd.language != "" {
			ps.Lock()
			defer ps.Unlock()

			if p, ok := ps.items[cmd.language]; ok {
				return p, nil
			}
			cmd.dump()
			return nil, errors.New(ErrUnsupportedLanguage.Error() + ": " + cmd.language)
		}
	}

	return nil, ErrUnknownCommand
}
