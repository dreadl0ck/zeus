#!/bin/bash

# ---------------------------------------------------------------------- #
# @zeus-chain: clean -> configure
# @zeus-help: install binary to $PATH
# ---------------------------------------------------------------------- #
# zeus install script
# installs the zeus binary to $PATH
#
# this is an example multiline manual text
# it can be viewed by typing help install in the interactive zeus shell
# 
# great!
# ---------------------------------------------------------------------- #

echo "installing zeus"
rice embed-go
godep go install
