#!/bin/bash

# {zeus}
# arguments:
# description: install to $PATH
# dependencies: clean -> configure
# outputs:
# help: Install the application to the default system location
# {zeus}

echo "installing zeus"
rice embed-go
godep go install

