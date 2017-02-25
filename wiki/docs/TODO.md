# ZEUS TODOs

    ________ ____  __ __  ______  __  ___
    \___   // __ \|  |  \/  ___/  \\/^//^\
     /    /\  ___/|  |  /\___ \   //  \\ /
    /_____ \\___  >____//____  >  \\  /^\\
          \/    \/           \/   /\\/\ //\
                An Electrifying Build System

- finish implementing Dependencies
- add @zeus-async header for parallel builds
- add zeus create < commandname > to bootstrap single command
- extract core into an importable package
- add date format config option
- watch scripts and parse again on WRITE event
- events: add support for filetypes
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

- Zeusfile
- markdown reports
- support more scripting languages
- use go-fuzz for tests