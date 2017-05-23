#!/bin/bash

# {zeus}
# dependencies: clean -> configure
# description: install binary to $PATH
# help: |
#     zeus install script
#     installs the zeus binary to $PATH
#     this is an example multiline manual text
#     it can be viewed by typing help install in the interactive zeus shell
# {zeus}

echo "installing zeus"
rice embed-go
godep go install
