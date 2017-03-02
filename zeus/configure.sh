#!/bin/bash

# -------------------------------------------------------------------------- #
# @zeus-chain: 
# @zeus-args: 
# @zeus-help: prepare JS and CSS
# -------------------------------------------------------------------------- #
# Prepare JS and CSS
# -------------------------------------------------------------------------- #

jsobfus -d frontend/src/js/:frontend/dist/js
sasscompile -d frontend/src/sass:frontend/dist/css