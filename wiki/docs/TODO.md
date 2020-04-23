# ZEUS TODOs

    ________ ____  __ __  ______  __  ___
    \___   // __ \|  |  \/  ___/  \\/^//^\
     /    /\  ___/|  |  /\___ \   //  \\ /
    /_____ \\___  >____//____  >  \\  /^\\
          \/    \/           \/   /\\/\ //\
                An Electrifying Build System

------------------------------------------------------------------------------------
    COMMIT:
------------------------------------------------------------------------------------

## v0.9

- both globals and arg values must be inserted into the code to make it runnable standalone (generate cmd)
    - code generation needs to be updated to generate setting the env vars at the beginning of the scripts
    - in bash accessed via $XXX, in python os.getenv("XXX") etc
    - before: global var access like any other variable; code generation unified  

### questions

- pass args AND globals via env?
- allow overwriting env vars via globals? 
- globals should start with an uppercase letter? or all uppercase?

### cleanup

- fix generate chains
- fix generateScript
- verify unit tests work
- update documentation

## General

- add install-completions command
- use filepath.Join instead of assembling paths with +
- add command to invoke golangci-lint
    - incorporate feedback from golangci-lint

- fix argument / chain / path completion
- add pitfalls to README (stopOnError in bash)
- add typescript support
- make formatter modular: add it as a field to language, to allow using a specific formatter for each language

- add edit data / source sub command
- [BUG] write events corrupt with vim?
- use gometalinter: add check command target
- create core package, make it importable + add godoc badge
- add golang sloc as default and add clocPath to config
- create commands for starting the js & scss watchers and add events, remove code from project
- release sasscompile and jstool -> add examples for event usage to README
- make date optional for milestones -> pass params in value= form
- improve git filter: parse output and format correctly in a table view + add colors + commit hashes, options: author=, date=, subject=, grep=
- improve test coverage

## readline

- fix tab completion for commandChainSeparator
- fix dynamic command chain completer

## next up

- web panel for all projects on localhost @ zeus.build
- generate reports
- integrate config-bob & vault
- integrate fstree & fsdiff as builtins
- add encrypted storage
- SVG & ascii dependency tree
- buildserver daemon
- add plugin api for language specific packages with new builtins (deadcode linter etc)