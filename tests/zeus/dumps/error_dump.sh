#!/bin/bash
#
# ZEUS Error Dump
# Timestamp: [Tue Feb 9 17:16:06 2021]
# Error: exit status 1
# StdErr: 
# 


#!/bin/bash

# {zeus}
# arguments:
# description: build project for current OS
# dependencies: configure
# outputs:
# buildNumber: true
# help: compile binary for current OS into buildDir
# {zeus}

echo "test: $test"

echo "building $binaryName for current OS"
rice embed-go
godep go build -o ${buildDir}/$binaryName
