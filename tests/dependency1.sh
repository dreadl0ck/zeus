#!/bin/bash

# ---------------------------------------------------------------------- #
# @zeus-help: test script for dependencies
# @zeus-outputs: bin/dependency1
# @zeus-args: argument:String
# ---------------------------------------------------------------------- #
# zeus dependency1 script
# 
# this script creates the file bin/dependency1
# ---------------------------------------------------------------------- #

echo "my arg: $argument"

# make sure bin dir exists
mkdir -p bin

echo "creating bin/dependency1"
touch bin/dependency1
