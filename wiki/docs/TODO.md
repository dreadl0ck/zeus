# ZEUS TODOs

    ________ ____  __ __  ______  __  ___
    \___   // __ \|  |  \/  ___/  \\/^//^\
     /    /\  ___/|  |  /\___ \   //  \\ /
    /_____ \\___  >____//____  >  \\  /^\\
          \/    \/           \/   /\\/\ //\
                An Electrifying Build System

---------------------------------------------------------------------------------------
    COMMIT:
---------------------------------------------------------------------------------------

This release adds support for using multiple scripting languages (python, javascript, ruby, bash) + adding custom languages in the config.

Also custom color profiles have been implemented and can be configured in the config.
The new YAML script header is lot easier to read and simplifies the parsing logic & header extension.

Two new builtins, have been added:
1) edit builtin, for quickly editing files on the commandline
2) generate builtin, for generating standalone versions of the build scripts

The globals have been refactored, to support a multi language setup.

This version intruduces a new directory structure:
1) all files that belong to zeus have moved into the zeus directory
2) all scripts moved into the zeus/scripts directory

The tests were improved, as well as the README, preview gif and overview graphic.
Addionally lot of code cleanup & renaming took place, come have a look!

## readline

- fix dynamic command chain completer
- fix path completion bug
- fix argument completion

## next up

- remove headers completely and supply script information in Zeusfile ???
- remove globals.yml -> supply global vars always in Zeusfile?
- dont parse globals for every command execution
- always populate execScript field when parsing, to avoid reading the file contents for every execution? This would require updating all dependency instances of the command as well... switch to looking up dep commands and pass args via Run func ?
- pass dependencies as array to YAML? would be cleaner with a large number of deps...

- make formatter modular: add as field to parser, to allow using a specific formatter for each language
- add sloc and make cloc optional in config
- create commands for starting the js & scss watchers, remove code from project
- release sasscompile and jstool -> add examples for event usage to README
- create core package, make it importable and add godocs

- make date optional for milestones
- improve git filter: parse output and format correctly + add colors + commit hashes, options: author=, date=, subject=, grep=
- improve create builtin: bootstrap headers interactively + bootstrap zeusfile entry

- generate reports
- improve tests
- improve parse error feedback

- integrate config-bob & vault
- integrate fstree & fsdiff as builtins
- add encrypted storage
- add scripting engine for using builtins during scripts


## future plans

- SVG & ascii dependency tree
- buildserver daemon
- web panel for all projects on localhost @ zeus.build