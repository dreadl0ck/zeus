#!/bin/bash

# {zeus}
# dependencies: clean -> configure
# description: start the dev mode
# help: |
#     zeus development mode script
#     this script compiles the binary without assets
# {zeus}

# remove assets to force rice debug stubs
rm -f rice-box.go

# install binary
go install
