#!/bin/bash

# -------------------------------------------------------------------------- #
# @zeus-help: prepare JS and CSS and move assets into place
# -------------------------------------------------------------------------- #
# Prepare JS and CSS
# -------------------------------------------------------------------------- #

echo "copying LICENSE and README.md"

cp LICENSE wiki/docs
cp README.md wiki/docs

jsobfus -d frontend/src/js/:frontend/dist/js
sasscompile -d frontend/src/sass:frontend/dist/css
