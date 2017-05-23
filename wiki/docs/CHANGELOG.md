# ZEUS CHANGELOG

    ________ ____  __ __  ______  __  ___
    \___   // __ \|  |  \/  ___/  \\/^//^\
     /    /\  ___/|  |  /\___ \   //  \\ /
    /_____ \\___  >____//____  >  \\  /^\\
          \/    \/           \/   /\\/\ //\
                An Electrifying Build System

## 23.05.17 - v0.7.4

- renamed run field in zeusfile to exec
- merged dependencies and chains
- renamed zeus fields args and deps
- added edit builtin
- made colorprofiles configurable in zeus config
- improved README
- added scripts folder inside zeusDir + moved Zeusfile into zeusDir
- removed support for Zeusfile without yaml file extension
- removed global config
- renamed help field to description
- renamed manual field to help
- implemented support for python, ruby and javascript
- refactored globals
- refactor config & data to embedded fields to hide mutex in yaml
- added support for overwriting languages in config
- testing micro with auto formatter (works nice!)
- implemented generate builtin
- check if argument label gets shadowed by a global
- removed tab support for YAML
- updated tests
- make zeusfile multiple languages possible -> set default lang in zeusfile, specific lang on commands if desired
- code cleanup & comments
- updated graffle
- updated README PDF
- updated gif

## 26.04.17

- refactored UI
- improved zeus_error_dump
- implemented async header field
- added procs builtin
- added completions for help command
- fixed async parsing bug
- added TODO count to projectHeader
- refactored colorProfiles
- SEO optimizations for github page
- added downloadable binaries

## 19.04.17

- improved Zeusfile parse error feedback
- argument refactoring: added optional args with default values, declaring args as comma separated list, requiring argument labels, requiring data type on arguments
- fixed panic when using tab completion on an alias

## 15.04.17

- fixed generic markdown wiki
- implemented todo builtin
- fix keybindings mapping to builtins
- fixed periodic zeus_data corruption
- removed executeCommand func

## 14.04.17

- fixed command arguments after parsing Zeusfile
- add tab completion for parameter labels
- bootstrapped git-filter builtin

## 13.04.17

- changed default PrintBuiltins to false
- added note about using bashrc in globals
- fixed makefile migration bug
- fix duplicate completer option after re-parsing scripts
- added more details about makefile conversion to README
- added a fatal when executed on windows
- mentioned behaviour regarding dispatched processes in the README

## 10.04.17

- fixed parsing zeusDir command chain

## 03.04.17

- switched to YAML for config file and project data
- implemented Zeusfile handling
- added an example Zeusfile
- updated README, changelog and tests
- sorting builtins alphabetically
- added test 50% badge
- updated LICENSE file
- implemented zeusfile bootstrapping
- implemented zeusfile to zeusDir migration

## 29.03.17

- added header watcher event to watch scripts and parse again on WRITE event
- updated tests
- working on parse error feedback

## 27.03.17

- handling strings for manipulating config
- command arguments accept name=val syntax
- added dateFormat for deadline and milestones to config
- refactored addEvent to keep eventID and formatterID when reloading events

## 22.03.17

- added zeus create to bootstrap single command

## 20.03.17

- warn about unknown config fields
- added filetypes for events

## 13.03.17

- improved UI
- improved events
- implemented dependencies & outputs

## 22.02.17

- updated godeps
- added wiki
- bootstrapped webinterface
- bootstrapped tests

## 21.02.17

- added sh/fileutil package
- disabled auto formatting by default to avoid issues with IDEs
- enabling DumpScriptOnError when enabling auto formatting
- ignoring WRITE events when there is a formatting job ongoing
- updated README

## 16.02.17

- fixed globals.sh generation when migrating makefiles
- fixed makefile target migration for targets that include blank lines for formatting reasons
- updated README
- fixed updating loglevel after changing to debug mode in the shell

## 14.02.17

- project structure cleanup
- created zeus_overview graffle
- added StopOnError and DumpScriptOnError config options
- updated README
- updated gif
- updated ascii