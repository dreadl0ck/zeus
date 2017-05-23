#!/bin/bash

# {zeus}
# description: prepare JS and CSS and move assets into place
# help: Prepare JS and CSS
# {zeus}

echo "copying LICENSE and README.md to wiki/docs"

# copy readme and license into wiki/docs to make them available for the web-wiki
# when started in the zeus directory
cp LICENSE wiki/docs
cp README.md wiki/docs

jsobfus -d frontend/src/js/:frontend/dist/js
sasscompile -d frontend/src/sass:frontend/dist/css
