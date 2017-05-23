#!/bin/bash

# {zeus}
# @zeus-dependencies: cycle arg1
# description: produce a cycle by calling the cycle command
# arguments:
#     - test:Bool
#     - asdf:String
# help: |
#     zeus cycle2 script
#     this script produces a cycle
#     this is a test for the zeus cycle detection
#     the cycle is created when cycle has cycle2 as dependency
#     this should result in a parse error
# {zeus}

echo "I am cycle2, lets call cycle"
