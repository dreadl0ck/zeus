# ZEUS TODOs

    ________ ____  __ __  ______  __  ___
    \___   // __ \|  |  \/  ___/  \\/^//^\
     /    /\  ___/|  |  /\___ \   //  \\ /
    /_____ \\___  >____//____  >  \\  /^\\
          \/    \/           \/   /\\/\ //\
                An Electrifying Build System

- keep eventID when reloading events
- watch scripts and parse again on WRITE event to make changing the headers possible without restarting the interactive shell
- remove logfile, generate reports
- improve tests and add more assertions

- add zeus create < commandname > to bootstrap single command
- add date format config option

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

- support more scripting languages
- use go-fuzz for tests