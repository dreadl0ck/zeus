#!/bin/bash

# {zeus}
# dependencies: dependency1 bla=test
# description: test script for dependencies
# outputs: 
#     - bin/dependency2
# arguments: 
#     - bla:String
# help: |
#     zeus dependency2 script
#     this script creates the file bin/dependency2
#     and requires bin/dependency1 to exist
# {zeus}

# make sure bin dir exists
mkdir -p bin

echo "creating bin/dependency2"
touch bin/dependency2
