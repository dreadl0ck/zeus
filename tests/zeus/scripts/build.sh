#!/bin/bash

# {zeus}
# arguments:
# description: build project for current OS
# dependencies: configure
# outputs:
# buildNumber: true
# help: compile binary for current OS into buildDir
# {zeus}

echo "building $binaryName for current OS"
rice embed-go
godep go build -o ${buildDir}/$binaryName
