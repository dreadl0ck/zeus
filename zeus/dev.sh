#!/bin/bash

# ---------------------------------------------------------------------- #
# @zeus-chain: clean -> configure
# @zeus-help: start the dev mode
# ---------------------------------------------------------------------- #
# zeus development mode script
# this script compiles the binary without assets
# ---------------------------------------------------------------------- #

# remove assets to force rice debug stubs
rm -f rice-box.go

# install binary
go install
