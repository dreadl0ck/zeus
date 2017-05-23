#!/bin/bash

# {zeus}
# arguments:
# description: start data race detection tests
# dependencies: clean
# outputs:
# help: 
# {zeus}

go test -v -race
