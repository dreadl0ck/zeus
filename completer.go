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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/dreadl0ck/readline"
)

var (
	// regex to the match a UNIX path
	shellPath = regexp.MustCompile("(([a-z]*[A-Z]*[0-9]*(_|-)*)*/*)*")

	// regex to match a command with a trailing UNIX path
	shellCommandWithPath = regexp.MustCompile("([a-z]*\\s*)*(([a-z]*[A-Z]*[0-9]*(_|-)*)*/*)*")

	// Enhanced dynamic completer for the interactive shell
	enhancedCompleter = readline.PcItemDynamic(enhancedTabCompleter)
)

type atomicCompleter struct {
	*readline.PrefixCompleter
	sync.RWMutex
}

func newAtomicCompleter() *atomicCompleter {
	return &atomicCompleter{
		newCompleter(),
		sync.RWMutex{},
	}
}

// Enhanced tab completer that handles command chains and argument completion
func enhancedTabCompleter(line string) []string {
	// Handle command chains by checking if we have the -> separator
	if strings.Contains(line, commandChainSeparator) {
		return handleCommandChainCompletion(line)
	}
	
	// Handle single command completion
	return handleSingleCommandCompletion(line)
}

// Handle completion for command chains (commands separated by ->)
func handleCommandChainCompletion(line string) []string {
	// Split by command chain separator
	chains := strings.Split(line, commandChainSeparator)
	
	// Get the last chain element (what we're currently completing)
	lastChain := strings.TrimSpace(chains[len(chains)-1])
	
	// If the last chain is empty, suggest command names
	if lastChain == "" {
		return getAvailableCommands()
	}
	
	// Parse the current command in the chain
	fields := strings.Fields(lastChain)
	if len(fields) == 0 {
		return getAvailableCommands()
	}
	
	commandName := fields[0]
	args := fields[1:]
	
	// Check if this is a valid command
	cmdMap.Lock()
	cmd, exists := cmdMap.items[commandName]
	cmdMap.Unlock()
	
	if !exists {
		// Command doesn't exist, suggest command names that start with the input
		return filterCommands(commandName)
	}
	
	// Complete arguments for this command
	return completeCommandArguments(cmd, args, lastChain)
}

// Handle completion for single commands (no chaining)
func handleSingleCommandCompletion(line string) []string {
	fields := strings.Fields(line)
	
	// If no fields, suggest all commands and builtins
	if len(fields) == 0 {
		result := getAvailableCommands()
		result = append(result, getBuiltinCommands()...)
		return result
	}
	
	commandName := fields[0]
	args := fields[1:]
	
	// Check if this is a builtin command
	if completion := handleBuiltinCompletion(commandName, args, line); completion != nil {
		return completion
	}
	
	// Check if this is a custom command
	cmdMap.Lock()
	cmd, exists := cmdMap.items[commandName]
	cmdMap.Unlock()
	
	if !exists {
		// Command doesn't exist, suggest commands and builtins that start with the input
		result := filterCommands(commandName)
		result = append(result, filterBuiltins(commandName)...)
		return result
	}
	
	// Complete arguments for this command
	return completeCommandArguments(cmd, args, line)
}

// Complete arguments for a specific command
func completeCommandArguments(cmd *command, args []string, fullLine string) []string {
	var suggestions []string
	
	// Track which arguments have been provided
	providedArgs := make(map[string]bool)
	var lastArg string
	
	// Parse existing arguments
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				providedArgs[parts[0]] = true
				lastArg = arg
			}
		} else {
			lastArg = arg
		}
	}
	
	// Check if we're in the middle of completing an argument value
	if strings.HasSuffix(fullLine, "=") || (lastArg != "" && strings.Contains(lastArg, "=") && !strings.HasSuffix(lastArg, " ")) {
		// We're completing an argument value
		var argName string
		if strings.HasSuffix(fullLine, "=") {
			// Get the argument name before the =
			beforeEquals := strings.TrimSuffix(fullLine, "=")
			parts := strings.Fields(beforeEquals)
			if len(parts) > 0 {
				argName = parts[len(parts)-1]
			}
		} else if strings.Contains(lastArg, "=") {
			parts := strings.SplitN(lastArg, "=", 2)
			argName = parts[0]
		}
		
		if argName != "" {
			return getArgumentValueSuggestions(cmd, argName)
		}
	}
	
	// Suggest missing arguments
	for _, cmdArg := range cmd.args {
		if !providedArgs[cmdArg.name] {
			suggestions = append(suggestions, cmdArg.name+"=")
		}
	}
	
	// Check if all required arguments are provided
	allRequiredProvided := true
	for _, cmdArg := range cmd.args {
		if !cmdArg.optional && !providedArgs[cmdArg.name] {
			allRequiredProvided = false
			break
		}
	}
	
	// If all required arguments are provided, suggest the command chain separator
	if allRequiredProvided {
		suggestions = append(suggestions, commandChainSeparator)
	}
	
	return suggestions
}

// Get argument value suggestions based on argument type
func getArgumentValueSuggestions(cmd *command, argName string) []string {
	// Find the argument definition
	var cmdArg *commandArg
	for _, arg := range cmd.args {
		if arg.name == argName {
			cmdArg = arg
			break
		}
	}
	
	if cmdArg == nil {
		return []string{}
	}
	
	// Provide suggestions based on argument type
	switch cmdArg.argType {
	case reflect.Bool:
		return []string{"true", "false"}
	case reflect.String:
		// For string arguments, suggest files/directories if it looks like a path argument
		if strings.Contains(strings.ToLower(argName), "path") || 
		   strings.Contains(strings.ToLower(argName), "file") ||
		   strings.Contains(strings.ToLower(argName), "dir") {
			return fileCompleter("")
		}
		// If there's a default value, suggest it
		if cmdArg.defaultValue != "" {
			return []string{cmdArg.defaultValue}
		}
		return []string{}
	case reflect.Int:
		// For int arguments, suggest some common values
		if cmdArg.defaultValue != "" {
			return []string{cmdArg.defaultValue}
		}
		return []string{"1", "10", "100"}
	case reflect.Float64:
		// For float arguments, suggest some common values
		if cmdArg.defaultValue != "" {
			return []string{cmdArg.defaultValue}
		}
		return []string{"0.0", "1.0", "10.0"}
	default:
		if cmdArg.defaultValue != "" {
			return []string{cmdArg.defaultValue}
		}
		return []string{}
	}
}

// Get all available custom commands
func getAvailableCommands() []string {
	var commands []string
	cmdMap.Lock()
	defer cmdMap.Unlock()
	for name, cmd := range cmdMap.items {
		if !cmd.hidden {
			commands = append(commands, name)
		}
	}
	return commands
}

// Filter commands based on prefix
func filterCommands(prefix string) []string {
	var matches []string
	cmdMap.Lock()
	defer cmdMap.Unlock()
	for name, cmd := range cmdMap.items {
		if !cmd.hidden && strings.HasPrefix(name, prefix) {
			matches = append(matches, name)
		}
	}
	return matches
}

// Get all builtin commands
func getBuiltinCommands() []string {
	return []string{
		exitCommand, helpCommand, infoCommand, clearCommand, formatCommand,
		globalsCommand, versionCommand, configCommand, createCommand, eventsCommand,
		milestonesCommand, gitFilterCommand, deadlineCommand, makefileCommand,
		dataCommand, aliasCommand, todoCommand, generateCommand, colorsCommand,
		authorCommand, updateCommand, builtinsCommand, keysCommand, editCommand,
		webCommand, procsCommand, wikiCommand,
		// Common shell commands
		"git", "ls", "cat", "rm", "tree", "mkdir", "touch", "micro",
	}
}

// Filter builtin commands based on prefix
func filterBuiltins(prefix string) []string {
	var matches []string
	builtins := getBuiltinCommands()
	for _, builtin := range builtins {
		if strings.HasPrefix(builtin, prefix) {
			matches = append(matches, builtin)
		}
	}
	return matches
}

// Handle completion for builtin commands
func handleBuiltinCompletion(commandName string, args []string, line string) []string {
	switch commandName {
	case helpCommand:
		if len(args) == 0 {
			return getAvailableCommands()
		}
		return nil
		
	case configCommand:
		if len(args) == 0 {
			return []string{"set", "get"}
		}
		if len(args) == 1 && (args[0] == "set" || args[0] == "get") {
			return getConfigItemNames()
		}
		return nil
		
	case createCommand:
		if len(args) == 0 {
			return getLanguageNames()
		}
		if len(args) == 1 && args[0] == "script" {
			return append([]string{"all"}, getAvailableCommands()...)
		}
		return nil
		
	case eventsCommand:
		if len(args) == 0 {
			return []string{"add", "remove"}
		}
		if len(args) == 1 && args[0] == "add" {
			return []string{"WRITE", "REMOVE", "CHMOD", "RENAME"}
		}
		if len(args) == 1 && args[0] == "remove" {
			return getEventIDs()
		}
		return nil
		
	case editCommand:
		if len(args) == 0 {
			commands := getAvailableCommands()
			commands = append(commands, "commands", "data", "config", "todo", "globals")
			return commands
		}
		if len(args) == 1 && args[0] == "globals" {
			return getLanguageNames()
		}
		return nil
		
	case procsCommand:
		if len(args) == 0 {
			return []string{"detach", "kill", "attach"}
		}
		if len(args) == 1 && args[0] == "detach" {
			return getAvailableCommands()
		}
		if len(args) == 1 && (args[0] == "kill" || args[0] == "attach") {
			return getPIDs()
		}
		return nil
		
	case colorsCommand:
		if len(args) == 0 {
			colors := []string{"off", "default"}
			colors = append(colors, getColorProfiles()...)
			return colors
		}
		return nil
		
	case generateCommand:
		if len(args) == 0 {
			return getAvailableCommands()
		}
		return nil
		
	case todoCommand:
		if len(args) == 0 {
			return []string{"add", "remove"}
		}
		if len(args) == 1 && args[0] == "remove" {
			return getTodoIndices()
		}
		return nil
		
	case aliasCommand, authorCommand, deadlineCommand:
		if len(args) == 0 {
			return []string{"set", "remove"}
		}
		return nil
		
	case milestonesCommand:
		if len(args) == 0 {
			return []string{"set", "remove", "add"}
		}
		return nil
		
	case makefileCommand:
		if len(args) == 0 {
			return []string{"migrate"}
		}
		return nil
		
	case keysCommand:
		if len(args) == 0 {
			return []string{"set", "remove"}
		}
		if len(args) == 1 && (args[0] == "set" || args[0] == "remove") {
			return getKeyCombinations()
		}
		return nil
		
	// Shell commands
	case "ls", "tree":
		return directoryCompleter(line)
	case "cat", "rm", "micro":
		return fileCompleter(line)
	case "git":
		if len(args) == 0 {
			return []string{"add", "status", "commit", "push", "pull", "branch", "checkout"}
		}
		return nil
	}
	
	return nil
}

// return a new default completer instance
func newCompleter() *readline.PrefixCompleter {
	// Use the enhanced dynamic completer for everything
	return readline.NewPrefixCompleter(enhancedCompleter)
}

/*
 *	Helper functions for specific completion types
 */

// assemble and return all items for config item completion
// also used for validating the config YAML for unknown fields
// if there's a key in the config that is not in here there will be a warning
func configItems() []readline.PrefixCompleterInterface {
	configNames := getConfigItemNames()
	var items []readline.PrefixCompleterInterface
	for _, name := range configNames {
		items = append(items, readline.PcItem(name))
	}
	return items
}

func getConfigItemNames() []string {
	return []string{
		"makefileOverview", "autoFormat", "fixParseErrors", "colors", "passCommandsToShell",
		"eebInterface", "interactive", "debug", "recursionDepth", "projectNamePrompt",
		"colorProfile", "historyFile", "historyLimit", "exitOnInterrupt", "disableTimestamps",
		"printBuiltins", "dumpScriptOnError", "stopOnError", "portWebPanel", "portGlueServer",
		"dateFormat", "todoFilePath", "editor", "codeSnippetScope", "quiet",
	}
}

func getLanguageNames() []string {
	var languages []string
	ls.Lock()
	defer ls.Unlock()
	for name := range ls.items {
		languages = append(languages, name)
	}
	return languages
}

func getColorProfiles() []string {
	var profiles []string
	conf.Lock()
	defer conf.Unlock()
	for name := range conf.fields.ColorProfiles {
		profiles = append(profiles, name)
	}
	return profiles
}

func getEventIDs() []string {
	var ids []string
	projectData.Lock()
	defer projectData.Unlock()
	for _, e := range projectData.fields.Events {
		ids = append(ids, e.ID)
	}
	return ids
}

func getPIDs() []string {
	var pids []string
	projectData.Lock()
	defer projectData.Unlock()
	for _, p := range processMap {
		pids = append(pids, strconv.Itoa(p.PID))
	}
	return pids
}

func getTodoIndices() []string {
	contents, err := ioutil.ReadFile(conf.fields.TodoFilePath)
	if err != nil {
		return []string{}
	}
	var indices []string
	var index int
	for _, line := range strings.Split(string(contents), "\n") {
		if strings.HasPrefix(line, "- ") {
			index++
			indices = append(indices, strconv.Itoa(index))
		}
	}
	return indices
}

func getKeyCombinations() []string {
	return []string{
		"Ctrl-A", "Ctrl-B", "Ctrl-E", "Ctrl-F", "Ctrl-G", "Ctrl-H", "Ctrl-I", "Ctrl-J",
		"Ctrl-K", "Ctrl-L", "Ctrl-M", "Ctrl-N", "Ctrl-O", "Ctrl-P", "Ctrl-Q", "Ctrl-R",
		"Ctrl-S", "Ctrl-T", "Ctrl-U", "Ctrl-V", "Ctrl-W", "Ctrl-X", "Ctrl-Y",
	}
}

/*
 *	Legacy completers (kept for compatibility)
 */

// complete eventIDs for removing events
func eventIDCompleter(path string) (res []string) {
	return getEventIDs()
}

// complete available commands
func commandCompleter(path string) (res []string) {
	return getAvailableCommands()
}

// complete available parser languages
func languageCompleter(path string) (res []string) {
	return getLanguageNames()
}

func colorProfileCompleter(path string) (res []string) {
	return getColorProfiles()
}

func todoIndexCompleter(path string) (res []string) {
	return getTodoIndices()
}

// complete PIDs for killing processes
func pIDCompleter(path string) (res []string) {
	return getPIDs()
}

// complete available filetypes for the event target directory
func fileTypeCompleter(path string) (res []string) {
	var (
		fields = strings.Fields(path)
		dir    string
	)

	if len(fields) > 2 {
		dir = fields[3]
	} else {
		return
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		Log.Error(err)
		return res
	}

	for _, f := range files {
		res = append(res, filepath.Ext(f.Name()))
	}

	// remove duplicates
	var (
		out []string
		ok  bool
	)

	for _, path := range res {
		for _, name := range out {
			if path == name {
				ok = true
			}
		}
		if !ok && path != "" {
			out = append(out, path)
		}
		ok = false
	}

	return out
}

// return available directories
func directoryCompleter(path string) (names []string) {
	files, dir := getFilesInDir(path)

	for _, f := range files {
		if dir == "./" {
			if f.IsDir() {
				names = append(names, strings.TrimPrefix(f.Name(), "./")+"/")
			}
			continue
		}
		if f.IsDir() {
			names = append(names, f.Name()+"/")
		}
	}
	return
}

func getFilesInDir(path string) (files []os.FileInfo, dir string) {
	var (
		fields = strings.Fields(path)
		fLen   = len(fields)
		err    error
	)
	if fLen < 2 {
		path = "./"
	} else {
		path = fields[fLen-1]
	}

	dir, _ = filepath.Split(path)
	files, err = ioutil.ReadDir(dir)
	if err != nil {
		if len(strings.Split(dir, "/")) > 1 {
			files, _ = ioutil.ReadDir(filepath.Base(dir))
		} else {
			dir = "./"
			files, _ = ioutil.ReadDir(dir)
		}
	}

	return
}

func fileCompleter(path string) (names []string) {
	files, dir := getFilesInDir(path)

	for _, f := range files {
		if dir == "./" {
			name := strings.TrimPrefix(f.Name(), "./")
			if f.IsDir() {
				names = append(names, name+"/")
				continue
			}
			names = append(names, name)
			continue
		}

		if f.IsDir() {
			names = append(names, f.Name()+"/")
			continue
		}
		names = append(names, f.Name())
	}
	return
}
