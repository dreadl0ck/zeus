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
	"os"
	"strconv"
	"strings"

	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/dreadl0ck/readline"
	"github.com/mgutz/ansi"
)

// ErrNoFileExtension means the script does not have a file extension
var ErrNoFileExtension = errors.New("no file extension")

type parseJobID string

// parseJob represents a parse process for a specific script
// or a command parsed from a zeusfile
// it keeps track of all parsed commands to prevent cycles
// parseJobs can run concurrently
type parseJob struct {

	// path for script to parse, empty if its a command from zesufile
	path string

	// unique identifier
	id parseJobID

	// command array with arguments
	commands [][]string

	// log parse errors to stdout
	silent bool

	// jobs waiting for command currently being parsed by this job
	waiters []chan bool

	// pointer to the parser instance the job belongs to
	p *parser
}

// newJob returns a new parseJob for the given path
func newJob(path string, silent bool, p *parser) *parseJob {
	return &parseJob{
		path:     path,
		id:       parseJobID(randomString()),
		commands: make([][]string, 0),
		silent:   silent,
		p:        p,
	}
}

// parse the command chain string
func (job *parseJob) parseCommandChain(line string) ([][]string, error) {

	var (
		// trim zeus prefix then get commands separated by parser separator
		cmds           = strings.Split(line, commandChainSeparator)
		parsedCommands = [][]string{}
		cLog           = Log.WithFields(logrus.Fields{
			"prefix": "parseCommandChain",
			"path":   job.path,
		})
	)

	// trim whitespace on all fields
	for i, field := range cmds {
		cmds[i] = strings.TrimSpace(field)
	}

	cLog.WithField("cmds", cmds).Debug("parsing cmdChain")

	if len(cmds) > 0 {

		// when there are no commands specified
		// the resulting slice from the split contains 1 empty string
		if cmds[0] == "" {
			return parsedCommands, nil
		}

		// range them
		for i, name := range cmds {

			// get arguments for commands
			var args = strings.Fields(name)

			if len(args) == 0 {
				return nil, errors.New(ErrInvalidCommand.Error() + " at index: " + strconv.Itoa(i))
			}

			parsedCommands = append(parsedCommands, args)

			cLog.Debug("adding " + cp.CmdOutput + strings.Join(args, " ") + ansi.Reset + " to parsedCommands")
		}
	}

	return parsedCommands, nil
}

// newCommand creates a new command instance for the script at path
// a parseJob will be created
func (job *parseJob) newCommand(path string) (*command, error) {

	var (
		cLog = Log.WithField("prefix", "newCommand")
		ext  = filepath.Ext(path)
		lang string
	)

	if len(ext) == 0 {
		return nil, ErrNoFileExtension
	}

	// parse the script
	d, err := job.p.parseScript(path, job)
	if err != nil {
		if !job.silent {
			cLog.WithFields(logrus.Fields{
				"path": path,
			}).Debug("Parse error")
		}
		return nil, err
	}

	// assemble commands args
	args, err := validateArgs(d.Arguments)
	if err != nil {
		return nil, err
	}

	chain, err := job.parseCommandChain(d.Dependencies)
	if err != nil {
		return nil, err
	}

	// get build chain
	commandChain, err := job.getCommandChain(chain, nil)
	if err != nil {
		return nil, err
	}

	// get name for command
	name := strings.TrimSuffix(strings.TrimPrefix(path, scriptDir+"/"), job.p.language.FileExtension)
	if name == "" {
		return nil, ErrInvalidCommand
	}

	// determine language by looking at fileExtension
	ps.Lock()
	for name, p := range ps.items {
		if p.language.FileExtension == ext {
			lang = name
		}
	}
	ps.Unlock()

	return &command{
		path:        path,
		name:        name,
		language:    lang,
		args:        args,
		help:        d.Help,
		description: d.Description,
		PrefixCompleter: readline.PcItem(name,
			readline.PcItemDynamic(func(path string) (res []string) {
				for _, a := range args {
					if !strings.Contains(path, a.name+"=") {
						res = append(res, a.name+"=")
					}
				}
				return
			}),
		),
		buildNumber:  d.BuildNumber,
		dependencies: commandChain,
		outputs:      d.Outputs,
		async:        d.Async,
	}, nil
}

// assemble a commandChain with a list of parsed commands and their arguments
func (job *parseJob) getCommandChain(parsedCommands [][]string, zeusfile *Zeusfile) (commandChain commandChain, err error) {

	var cLog = Log.WithFields(logrus.Fields{
		"prefix": "getCommandChain",
		"path":   job.path,
	})

	cLog.WithField("parsedCommands", parsedCommands).Debug(cp.CmdArgType + "creating cmdChain" + ansi.Reset)

	cLog = cLog.WithField("job.commands", job.commands)

	// empty commandChain is OK
	for _, args := range parsedCommands {

		var count int

		// check if there are repetitive targets in the chain
		// to prevent cyclos
		for _, c := range job.commands {

			// check if the key (commandName) is already there
			if c[0] == args[0] {
				count++
			}
		}

		conf.Lock()
		recursionDepth := conf.fields.RecursionDepth
		conf.Unlock()

		if count > recursionDepth {
			cLog.WithFields(logrus.Fields{
				"count": count,
			}).Error("CYCLE DETECTED! -> ", args[0], " appeared more than ", recursionDepth, " times - thats invalid.")
			cleanup()
			os.Exit(1)
		}

		job.commands = append(job.commands, args)

		var jobPath = scriptDir + "/" + args[0] + job.p.language.FileExtension
		if zeusfile != nil {
			jobPath = "zeusfile." + args[0]
		}

		// check if command has already been parsed
		cmdMap.Lock()
		cmd, ok := cmdMap.items[args[0]]
		cmdMap.Unlock()

		if !ok {

			cmd, err = job.getCommand(jobPath, args, zeusfile)
			if err != nil {
				return commandChain, err
			}
		}

		cLog.Debug("adding " + cp.CmdOutput + strings.Join(args, " ") + ansi.Reset + " to cmdChain")

		// this command has argument parameters in its commandChain
		// set them on the command
		if len(args) > 1 {

			cLog.WithFields(logrus.Fields{
				"command": args[0],
				"params":  args[1:],
			}).Debug("setting parameters")

			// creating a hard copy of the struct here,
			// otherwise params would be set for every execution of the command
			cmd = &command{
				name:            cmd.name,
				path:            cmd.path,
				params:          args[1:],
				args:            cmd.args,
				description:     cmd.description,
				help:            cmd.help,
				dependencies:    cmd.dependencies,
				PrefixCompleter: cmd.PrefixCompleter,
				buildNumber:     cmd.buildNumber,
				language:        cmd.language,
			}
		}

		// append command to build chain
		commandChain = append(commandChain, cmd)

		if debug {
			cmd.dump()
		}
	}
	return
}

func (job *parseJob) getCommand(jobPath string, args []string, zeusfile *Zeusfile) (*command, error) {

	cLog := Log.WithField("prefix", "getCommand")

	// check if command is currently being parsed
	if job.p.JobExists(jobPath) {
		cLog.Warn("getCommandChain: JOB EXISTS: ", jobPath)
		job.p.WaitForJob(jobPath)

		// now the command is there
		cmdMap.Lock()
		cmd, ok := cmdMap.items[args[0]]
		cmdMap.Unlock()

		if !ok {
			return nil, errors.New("command " + args[0] + "not found after waiting for it - wtf")
		}

		return cmd, nil
	}

	if zeusfile != nil {

		d := zeusfile.Commands[args[0]]
		if d == nil {
			return nil, errors.New("invalid command in commandChain: " + args[0])
		}

		// assemble commands args
		arguments, err := validateArgs(d.Arguments)
		if err != nil {
			return nil, err
		}

		parsedCommands, err := job.parseCommandChain(d.Dependencies)
		if err != nil {
			return nil, err
		}

		chain, err := job.getCommandChain(parsedCommands, zeusfile)
		if err != nil {
			return nil, err
		}

		// create command
		cmd := &command{
			path:            "",
			name:            args[0],
			args:            arguments,
			description:     d.Description,
			help:            d.Help,
			PrefixCompleter: readline.PcItem(args[0]),
			buildNumber:     d.BuildNumber,
			dependencies:    chain,
			outputs:         d.Outputs,
			execScript:      d.Exec,
			async:           d.Async,
			language:        zeusfile.Language,
		}

		return cmd, nil
	}

	// add new command
	cmd, err := job.newCommand(scriptDir + "/" + args[0] + job.p.language.FileExtension)
	if err != nil {
		if !job.silent {
			cLog.WithError(err).Debug("failed to create command")
		}
		return nil, err
	}

	cmdMap.Lock()

	// add the completer
	completer.Children = append(completer.Children, cmd.PrefixCompleter)

	// add to command map
	cmdMap.items[args[0]] = cmd

	cmdMap.Unlock()

	cLog.Debug("added " + cp.CmdName + cmd.name + ansi.Reset + " to the command map")

	return cmd, nil
}

// thread safe
func (p *parser) AddJob(path string, silent bool) (job *parseJob) {

	job = newJob(path, silent, p)

	Log.WithFields(logrus.Fields{
		"ID":   job.id,
		"path": path,
	}).Debug("adding job")

	p.Lock()
	p.jobs[job.id] = job
	p.Unlock()

	return job
}

// RemoveJob removes a job from the parser
// thread safe
func (p *parser) RemoveJob(job *parseJob) {

	Log.WithFields(logrus.Fields{
		"ID":      job.id,
		"path":    job.path,
		"waiters": len(job.waiters),
	}).Debug("removing job")

	// notify waiters
	for _, c := range job.waiters {
		c <- true
	}

	p.Lock()
	delete(p.jobs, job.id)
	p.Unlock()
}

func (p *parser) JobExists(path string) bool {

	Log.WithFields(logrus.Fields{
		"path": path,
	}).Debug("job exists?")

	p.Lock()
	defer p.Unlock()

	for _, job := range p.jobs {
		if job.path == path {
			return true
		}
	}

	return false
}

// wait for a running parseJob
func (p *parser) WaitForJob(path string) {

	p.printJobs()

	c := make(chan bool)

	p.Lock()

	for _, job := range p.jobs {
		if job.path == path {

			// add channel to waiters
			job.waiters = append(job.waiters, c)

			Log.WithFields(logrus.Fields{
				"ID":   job.id,
				"path": job.path,
			}).Debug("waiting for job")

			p.Unlock()
			<-c

			Log.WithFields(logrus.Fields{
				"ID":   job.id,
				"path": job.path,
			}).Debug(ansi.Yellow + "job complete" + ansi.Reset)

			return
		}
	}
	p.Unlock()
}

func (p *parser) printJobs() {

	p.Lock()
	defer p.Unlock()

	l.Println(cp.Prompt + pad("ID", 20) + " path" + cp.Text)
	for _, job := range p.jobs {
		l.Println(pad(string(job.id), 20), job.path)
	}
}
