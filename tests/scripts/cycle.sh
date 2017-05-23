#!/bin/bash

# {zeus}
# description: produce a cycle by calling the cycle command
# arguments: 
#     - bla:String
# help: |
#     zeus cycle script
#     this script produces a cycle
#     this is a test for the zeus cycle detection
#     the cycle is created when cycle has cycle2 as dependency
#     this should result in a parse error
# {zeus}

echo "Im a cycle"
echo "test"

# use global func
yolo
