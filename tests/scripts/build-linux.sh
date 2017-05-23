#!/bin/bash

# {zeus}
# arguments:
# description: build project for linux amd64
# dependencies: configure
# outputs:
# buildNumber: true
# help: compile binary for linux amd64 into buildDir
# {zeus}

echo "building for linux amd64"
rice embed-go
GOOS=linux GOARCH=amd64 godep go build -o ${buildDir}/zeus-linux

