#!/bin/bash

# {zeus}
# description: test script for dependencies
# outputs: 
#     - bin/dependency1
# arguments: 
#     - bla:String
# help: |
#     zeus dependency1 script
#     this script creates the file bin/dependency1
# {zeus}

# make sure bin dir exists
mkdir -p bin

echo "creating bin/dependency1"
touch bin/dependency1
