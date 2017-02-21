# ZEUS TODOs

    ________ ____  __ __  ______  __  ___
    \___   // __ \|  |  \/  ___/  \\/^//^\
     /    /\  ___/|  |  /\___ \   //  \\ /
    /_____ \\___  >____//____  >  \\  /^\\
          \/    \/           \/   /\\/\ //\
                An Electrifying Build System

- add date format config option
- finish implementing Dependencies
- add @zeus-async header for parallel builds
- watch scripts and parse again on WRITE event
- fix globals.sh generation when migrating makefiles
- events: add support for filetypes
- arguments: accept name=val syntax

## COMING SOON

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

- markdown wiki
- Zeusfile
- markdown reports
- goconvey tests
- webUI with live stats, controls and current report
- support more scripting languages