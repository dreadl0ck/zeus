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
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	rice "github.com/GeertJohan/go.rice"
	"github.com/dreadl0ck/readline"
	"github.com/mgutz/ansi"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (

	// Log instance for internal logs
	Log = newAtomicLogger()

	// logging instance for terminal UI
	l = log.New(os.Stdout, "", 0)

	// all available build target commands
	// does not include the built-ins!
	// command names mapped to command structs
	cmdMap = newCommandMap()

	// readline auto completion
	completer = newAtomicCompleter()

	// assets folder
	assetBox *rice.Box

	// configuration
	conf *config

	// project data
	projectData *data

	// shell formatter
	f = newFormatter("path/to/your/formatter", bashLanguage())

	g = &globals{
		Vars: make(map[string]string, 0),
	}

	debug        bool
	asciiArt     string
	asciiArtYAML string
	workingDir   string

	// status info
	s = &status{
		recursionMap: make(map[string]int, 0),
	}

	// running a test?
	testingMode bool

	promptFilePath = "zeus/.zeus_prompt"
)

type atomicLogger struct {
	*logrus.Logger
	sync.RWMutex
}

func newAtomicLogger() *atomicLogger {
	return &atomicLogger{
		logrus.New(),
		sync.RWMutex{},
	}
}

func initZeus() {

	var (
		err             error
		flagCompletions = flag.String("completions", "", "get available command completions")
		flagWorkDir     = flag.String("C", "", "set work directory to start from")
		flagHelp        = flag.Bool("h", false, "print zeus help and exit")
	)

	// set up formatter
	Log.Formatter = &prefixed.TextFormatter{}

	if runtime.GOOS == "windows" {
		Log.Fatal("windows is not (yet) supported.")
	}

	assetBox = rice.MustFindBox("assets")
	asciiArt, err = assetBox.String("ascii_art.txt")
	if err != nil {
		Log.WithError(err).Fatal("failed to get ascii_art.txt from rice box")
	}
	asciiArtYAML, err = assetBox.String("ascii_art.yml")
	if err != nil {
		Log.WithError(err).Fatal("failed to get ascii_art.yml from rice box")
	}

	// add version number
	asciiArtYAML += version + "\n#\n\n"

	if len(os.Args) > 1 {
		if os.Args[1] == bootstrapCommand || os.Args[1] == "init" { // allow init command as well, similar to other tools like git, go mod etc
			runBootstrapCommand()

			// remove bootstrap arg
			os.Args = []string{os.Args[0]}
		}
	}

	if len(os.Args) > 2 {
		if os.Args[1] == "makefile" && os.Args[2] == "migrate" {
			migrateMakefile(zeusDir)
			os.Exit(0)
		}
	}

	flag.Parse()

	if *flagWorkDir != "" {
		if strings.HasPrefix(*flagWorkDir, "~") {
			usr, err := user.Current()
			if err != nil {
				log.Fatal("unable to get current user for expanding the ~ character: ", err)
			}
			*flagWorkDir = filepath.Join(usr.HomeDir, strings.TrimPrefix(*flagWorkDir, "~"))
			fmt.Println("expanded ~:", *flagWorkDir)
		}
		if strings.HasPrefix(*flagWorkDir, "$HOME") {
			usr, err := user.Current()
			if err != nil {
				log.Fatal("unable to get current user for expanding $HOME: ", err)
			}
			*flagWorkDir = filepath.Join(usr.HomeDir, strings.TrimPrefix(*flagWorkDir, "$HOME"))
			fmt.Println("expanded $HOME:", *flagWorkDir)
		}
		err := os.Chdir(*flagWorkDir)
		if err != nil {
			log.Fatal("failed to change dir: ", err)
		}

		fullPath, err := os.Getwd()
		if err != nil {
			log.Fatal("failed to obtain full path of current working directory: ", err)
		}

		projectDir = fullPath
	} else {
		projectDir, err = os.Getwd()
		if err != nil {
			log.Fatal("failed to set project directory on startup: ", err)
		}
	}

	if *flagCompletions != "" {
		printCompletions(*flagCompletions)
		os.Exit(0)
	}

	if *flagHelp {
		printHelp()
	}

	stat, err := os.Stat(scriptDir)
	if err != nil {
		if stat, err = os.Stat(commandsFilePath); err != nil {
			Log.WithError(err).Error("no " + scriptDir + " directory or CommandsFile found.")
			Log.Info("run 'zeus bootstrap' to create a default setup, or 'zeus makefile migrate' if you want to migrate from a GNU Makefile.")
			os.Exit(1)
		} else {
			// make sure its a file
			if stat.IsDir() {
				Log.Fatal(commandsFilePath + " is not a file")
			}
		}
	} else {
		// make sure its a directory
		if !stat.IsDir() {
			Log.Fatal("zeus/ is not a directory")
		}
	}
}

func main() {

	initZeus()

	var (
		cLog           = Log.WithField("prefix", "main")
		err            error
		configWarnings []string
	)

	// look for project data
	projectData, err = parseProjectData()
	if err != nil {
		cLog.WithError(err).Debug("error looking for project data")
		projectData = newData()
	}

	// look for project config
	conf, configWarnings, err = parseProjectConfig()
	if err != nil {
		cLog.WithError(err).Debug("failed to parse project config")
		cLog.Info("initializing default configuration")

		conf = newConfig()
		conf.update()
	}

	initColorProfile()

	// load persisted events from project data
	loadEvents()

	projectData.Lock()

	// validate aliases
	for name := range projectData.fields.Aliases {
		err = validateAlias(name)
		if err != nil {
			Log.WithError(err).Fatal("failed to validate alias: ", name)
		}

		// add to completer
		completer.Children = append(completer.Children, readline.PcItem(name))
	}

	projectData.Unlock()

	// get debug value from config
	debug = conf.fields.Debug

	// handle debug mode for logger
	if debug {
		Log.Level = logrus.DebugLevel
	}

	if conf.fields.DisableTimestamps {
		formatter := new(prefixed.TextFormatter)
		formatter.DisableTimestamp = true
		Log.Formatter = formatter
	}

	// disable colors
	if !conf.fields.Colors {

		print(cp.Reset)

		// lock once
		cp.Lock()
		cp = colorsOffProfile().parse()

		Log.Formatter = &prefixed.TextFormatter{
			DisableColors:    true,
			DisableTimestamp: conf.fields.DisableTimestamps,
		}

		ansi.DisableColors(true)
	} else {
		// load colored ascii art
		asciiArt, err = assetBox.String("ascii_art_color.txt")
		if err != nil {
			Log.WithError(err).Fatal("failed to get ascii_art_color.txt from rice box")
		}
	}

	// set working directory
	workingDir, err = os.Getwd()
	if err != nil {
		cLog.WithError(err).Fatal("failed to get current directory name")
	}

	// only execute when using the interactive shell
	if len(os.Args) == 1 {
		printProjectHeader()
	}

	// print config warnings
	for _, w := range configWarnings {
		Log.Warn(w)
	}

	// start watchers when running in interactive mode
	if conf.fields.Interactive {

		// watch config for changes
		go conf.watch("")

		if conf.fields.AutoFormat {
			// watch zeus directory for changes
			go f.watchScriptDir("")
		}
	}

	// print makefile command overview
	if conf.fields.MakefileOverview {
		printMakefileCommandOverview()
	}

	if conf.fields.ProjectNamePrompt {
		// set shell prompt to project name
		zeusPrompt = filepath.Base(workingDir)
	}

	// check if a CommandsFile for the project exists
	cmdFile, err := parseCommandsFile(commandsFilePath, false)
	if err != nil {
		Log.Error("failed to parse commandsFile: ", err, "\n")
		os.Exit(1)
	}

	// handle commandsFile extension
	cmdFile.handleExtension()

	// handle commandsFile inclusion
	cmdFile.handleInclusion()

	// watch commandsFile for changes in interactive mode
	if conf.fields.Interactive {
		go watchCommandsFile(commandsFilePath, "")
	}

	// handle commandline arguments
	handleArgs(os.Args, cmdFile)

	// check if interactive mode is enabled in the config
	if conf.fields.Interactive {

		if conf.fields.WebInterface {
			go StartWebListener(true)
		}

		// handle OS Signals
		// all child processes need to be killed when theres an error
		handleSignals(cmdFile)

		// start interactive mode and start reading from stdin
		err = readlineLoop(cmdFile)
		if err != nil {
			cLog.WithError(err).Fatal("failed to read user input")
		}
	} else {
		printProjectHeader()
		if conf.fields.PrintBuiltins {
			printBuiltins()
		}
		printCommands()
	}
}

func printHelp() {
	l.Println("ZEUS - An Electrifying Build System")
	l.Println("author: dreadl0ck@protonmail.ch")
	l.Println("for documentation see: https://github.com/dreadl0ck/zeus")
	l.Println("run 'zeus bootstrap' to create a default ZEUS setup")
	l.Println("or 'zeus makefile migrate' if you want to migrate from a GNU Makefile.")
	l.Println("to get an overview over an existing ZEUS project, run 'zeus help' or 'zeus' for the interactive shell.")
	os.Exit(0)
}

// print the project ascii art and project infos
func printProjectHeader() {

	// print ascii art
	clearScreen()
	l.Println(cp.Text + asciiArt + "v" + version)

	if !conf.fields.Quiet {
		if conf.fields.Debug {
			l.Println(cp.Text + pad("Project Name", 14) + cp.Prompt + filepath.Base(workingDir) + cp.Text + "\n")
		}
		printAuthor()
	}

	printTodoCount()
	if projectData.fields.BuildNumber > 0 {
		l.Println(cp.Text + pad("BuildNumber", 14) + cp.Prompt + strconv.Itoa(projectData.fields.BuildNumber) + cp.Text)
	}
	if projectData.fields.Deadline != "" {
		l.Println(pad("Deadline", 14) + cp.Prompt + projectData.fields.Deadline + cp.Text)
	}

	// project infos
	listMilestones()
}

// handle commandline arguments
func handleArgs(args []string, cmdFile *CommandsFile) {

	// strip commandline flags
	for i, elem := range args {
		if strings.HasPrefix(elem, "-C=") {
			// delete i
			args = append(args[:i], args[i+1:]...)
			break
		}
		if elem == "-C" {
			// delete i and i+1
			args = append(args[:i], args[i+2:]...)
			break
		}
	}

	var cLog = Log.WithField("prefix", "handleArgs")

	if len(args) > 1 {

		var validCommand bool

		switch args[1] {
		case helpCommand:
			if conf.fields.PrintBuiltins {
				printBuiltins()
			}
			printCommands()

		case formatCommand:
			f.formatCommand()
		case dataCommand:
			printProjectData()

		case aliasCommand:
			if len(args) == 2 {
				printAliases()
				return
			}

			handleAliasCommand(args[2:])

		case configCommand:
			handleConfigCommand(args[2:])

		case versionCommand:
			l.Println(version)
		case updateCommand:
			updateZeus()
		case infoCommand:
			printProjectInfo()

		case colorsCommand:

			if len(args) == 3 {
				handleColorsCommand(args[1:])
			} else {
				printColorsUsageErr()
			}

		case authorCommand:
			handleAuthorCommand(args[1:])

		case builtinsCommand:
			printBuiltins()

		case makefileCommand:
			handleMakefileCommand(args[1:])
		case gitFilterCommand:
			handleGitFilterCommand(args[1:])

		case createCommand:
			handleCreateCommand(args[1:])
			os.Exit(0)

		default:
			handleSignals(cmdFile)
			cmdMap.Lock()

			// check if the command exists
			if cmd, ok := cmdMap.items[args[1]]; ok {
				cmdMap.Unlock()

				validCommand = true

				count, err := getTotalDependencyCount(cmd)
				if err != nil {
					l.Println(err)
					return
				}

				s.Lock()
				s.numCommands = count
				s.Unlock()

				shellBusy = true
				err = cmd.Run(args[2:], cmd.async)
				if err != nil {
					cLog.WithError(err).Error("failed to execute " + cmd.name)
					cleanup(cmdFile)
					os.Exit(1)
				}
				shellBusy = false
			} else {
				cmdMap.Unlock()
			}

			// check if its a commandChain supplied with "" or ''
			if strings.Contains(args[1], commandChainSeparator) {
				fields := strings.Split(args[1], commandChainSeparator)
				if cmdChain, ok := validCommandChain(fields, false); ok {
					shellBusy = true
					cmdChain.exec(fields)
					shellBusy = false
				} else {
					l.Println("invalid commandChain")
				}
				return
			}

			// check if its an alias
			if command, ok := projectData.fields.Aliases[args[1]]; ok {
				handleLine(command)
				os.Exit(0)
			}

			if !validCommand {
				if !testingMode {
					cLog.Fatal("unknown command: ", args[1])
				}
			}
		}
		if !testingMode {
			os.Exit(0)
		}
	}
}
