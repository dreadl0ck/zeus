#!/bin/bash
#
# ZEUS Error Dump
# Timestamp: [Tue May 23 18:57:07 2017]
# Error: exit status 1

#!/bin/bash
testVar="testValue"
serverIP="192.168.1.1"
number=4

#!/bin/bash

function sayHello() {
	echo "hello world!"
	echo "global var test: $testVar"
}


go test -v -race