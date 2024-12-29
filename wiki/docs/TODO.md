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

## TODO

- indicate in overview if there is a help text available for a command

- export all targets to scripts function!
  - proper script editing with syntax highlighting etc after bootstrapping in cmds file
- deps as arrays of arrays
  - everything in an array can be parallelized

- dump creation for commands with different working dir
- docs and examples for async command usage
- docs for writing commands in Go
- feedback jan path completion and shell passthrough behavior
- remove go rice and use go:embed instead

- globals should start with an uppercase letter
- fix argument / chain / path completion
- add pitfalls to README (stopOnError in bash)
- add typescript support
- add install-completions command

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