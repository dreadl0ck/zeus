#!/bin/bash
#
# ZEUS Error Dump
# Timestamp: [Tue May 23 20:31:08 2017]
# Error: exit status 1

#!/bin/bash
binaryName=zeus
buildDir=bin

#!/bin/bash

function sayHello() {
	echo "hello world!"
	echo "global var test: $testVar"
}


go test -v -race