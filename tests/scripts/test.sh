#!/bin/bash

# {zeus}
# arguments:
# description: run automated tests
# dependencies: clean
# outputs:
# help: 
# {zeus}

echo "starting tests"

go test -v -coverprofile coverage.out -cover

if [[ $? == 0 ]]; then
    go tool cover -html=coverage.out
fi

