#!/bin/bash

# ---------------------------------------------------------------------- #
# @zeus-chain: clean
# @zeus-help: build race detection enabled binary
# ---------------------------------------------------------------------- #
# zeus build-race script
# this script produces the zeus binary with race detection enabled
# ---------------------------------------------------------------------- #

go build -race -o bin/zeus && bin/zeus