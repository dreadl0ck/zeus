#!/bin/bash

# {zeus}
# arguments:
# description: build race detection enabled binary
# dependencies: clean
# outputs:
# help: zeus build-race script
this script produces the zeus binary with race detection enabled

# {zeus}

go build -race
