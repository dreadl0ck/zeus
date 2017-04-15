#!/bin/bash

# ---------------------------------------------------------------------- #
# @zeus-help: test script for dependencies
# @zeus-outputs: bin/dependency2
# @zeus-deps: dependency1 test-arg
# ---------------------------------------------------------------------- #
# zeus dependency2 script
# 
# this script creates the file bin/dependency2
# and requires bin/dependency1 to exist
# ---------------------------------------------------------------------- #

# make sure bin dir exists
mkdir -p bin

echo "creating bin/dependency2"
touch bin/dependency2
