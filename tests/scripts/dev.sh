#!/bin/bash

# {zeus}
# arguments:
# description: 
# dependencies: clean -> configure
# outputs:
# help: zeus development mode script
clears bindata & installs to $GOPATH

# {zeus}

rm -f rice-box.go
go install

