#!/bin/bash

# {zeus}
# dependencies: clean
# description: build race detection enabled binary
# help: |
#     zeus build-race script
#     this script produces the zeus binary with race detection enabled
# {zeus}

go build -race -o bin/zeus && bin/zeus
