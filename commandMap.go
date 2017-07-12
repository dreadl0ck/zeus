package main

import (
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dreadl0ck/readline"
	"github.com/mgutz/ansi"
)

type commandMap struct {
	items map[string]*command
	sync.RWMutex
}

func newCommandMap() *commandMap {
	return &commandMap{
		items: make(map[string]*command, 0),
	}
}

// flush commandMap
// and remove all command completions
func (cm *commandMap) flush() {

	cm.Lock()
	defer cm.Unlock()

	var (
		names       []string
		indices     []int
		newChildren []readline.PrefixCompleterInterface
	)
	for name := range cm.items {
		names = append(names, name)
	}

	// reset cmdMap
	cmdMap = newCommandMap()

	// remove all command completions
	completer.Lock()
	for i, comp := range completer.Children {
		for _, n := range names {
			if strings.TrimSpace(string(comp.GetName())) == n {
				indices = append(indices, i)
			}
		}
	}

	var addChild = true
	for i, comp := range completer.Children {
		for _, index := range indices {
			if i == index {
				addChild = false
			}
		}
		if addChild {
			newChildren = append(newChildren, comp)
		}

		// reset
		addChild = true
	}

	completer.Children = newChildren
	completer.Unlock()
}

func (cm *commandMap) length() int {
	cm.Lock()
	defer cm.Unlock()
	return len(cm.items)
}

func (cm *commandMap) init(start time.Time) {

	cLog := Log.WithField("prefix", "cmdMap.init")

	cm.Lock()
	defer cm.Unlock()

	// only print info when using the interactive shell
	if len(os.Args) == 1 {
		if len(cm.items) == 1 {
			l.Println(cp.Text+"initialized "+cp.Prompt, "1", cp.Text+" command in: "+cp.Prompt, time.Now().Sub(start), ansi.Reset+"\n")
		} else {
			l.Println(cp.Text+"initialized "+cp.Prompt, len(cmdMap.items), cp.Text+" commands in: "+cp.Prompt, time.Now().Sub(start), ansi.Reset+"\n")
		}
	}

	// check if custom command conflicts with builtin name
	for _, name := range builtins {
		if _, ok := cm.items[name]; ok {
			cLog.Error("command ", name, " conflicts with a builtin command. Please choose a different name.")
		}
	}

	var commandCompletions []readline.PrefixCompleterInterface
	for _, c := range cm.items {
		commandCompletions = append(commandCompletions, readline.PcItem(c.name))
	}

	// add all commands to the completer for the help page
	for _, c := range completer.Children {
		if string(c.GetName()) == "help " {
			c.SetChildren(commandCompletions)
		}
	}
}

// retrieve a command instance by passing a command string
func (cm *commandMap) getCommand(name string) (*command, error) {

	cmdMap.Lock()
	defer cmdMap.Unlock()

	if cmd, ok := cmdMap.items[name]; ok {

		// return command instance
		return cmd, nil
	}
	return nil, errors.New(ErrUnknownCommand.Error() + ": " + ansi.Red + name + cp.Text)
}
