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

## v0.8.1

- BUG: write events corrupt with vim?
- add typescript support
- use gometalinter
- make formatter modular: add it as a field to language, to allow using a specific formatter for each language
- create core package, make it importable + add godoc badge
- add golang sloc as default and add clocPath to config
- create commands for starting the js & scss watchers and add events, remove code from project
- release sasscompile and jstool -> add examples for event usage to README
- make date optional for milestones -> pass params in value= form
- improve git filter: parse output and format correctly in a table view + add colors + commit hashes, options: author=, date=, subject=, grep=
- improve test coverage

## readline

- fix tab completion for commandChainSeparator
- fixed dynamic command chain completer
- fix path completion bug
- fix argument completion

## next up

- web panel for all projects on localhost @ zeus.build
- generate reports
- integrate config-bob & vault
- integrate fstree & fsdiff as builtins
- add encrypted storage
- SVG & ascii dependency tree
- buildserver daemon
- add plugin api for language specific packages with new builtins (deadcode linter etc)