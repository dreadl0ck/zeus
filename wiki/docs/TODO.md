# ZEUS TODOs

    ________ ____  __ __  ______  __  ___
    \___   // __ \|  |  \/  ___/  \\/^//^\
     /    /\  ___/|  |  /\___ \   //  \\ /
    /_____ \\___  >____//____  >  \\  /^\\
          \/    \/           \/   /\\/\ //\
                An Electrifying Build System

- add support for running shell commands on events
- improve tests
- add zeus create < commandname > to bootstrap single command
- add date format config option
- watch scripts and parse again on WRITE event
- events: add support for filetypes
- Zeusfile

- add @zeus-async header for parallel builds
- extract core into an importable package
- arguments: accept name=val syntax

- add option to include encrypted data, example: crypto keys, passwords, infos etc

```shell
# zeus > data put <name> <content>
# enter password:
# repeat password:
# data encrypted and stored!

# zeus > data read <name>
# enter password:
# <content>
```

- markdown reports
- support more scripting languages
- use go-fuzz for tests