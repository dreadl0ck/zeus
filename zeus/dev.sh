#!/bin/bash

# ---------------------------------------------------------------------- #
# @zeus-chain: clean -> configure
# @zeus-help: start the dev mode
# ---------------------------------------------------------------------- #
# zeus development mode script
# this script compiles the binary without assets
# ---------------------------------------------------------------------- #

rm -f rice-box.go
godep go install