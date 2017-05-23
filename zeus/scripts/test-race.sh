#!/bin/bash

# {zeus}
# dependencies: clean
# description: start data race detection tests
# help: zeus data race detection test script
# {zeus}

go test -v -race
