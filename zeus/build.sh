#!/bin/bash

# ---------------------------------------------------------------------- #
# @zeus-chain: clean
# @zeus-help: build project
# @zeus-args: name:String
# @zeus-build-number
# ---------------------------------------------------------------------- #
# zeus build script
# this script produces the zeus binary
#
# it will be be placed in bin/$name
# ---------------------------------------------------------------------- #

echo "building $name"
godep go build -o bin/$name
