#!/bin/bash

# ---------------------------------------------------------------------- #
# @zeus-chain: reset
# @zeus-help: start tests
# ---------------------------------------------------------------------- #
# zeus test script
# ---------------------------------------------------------------------- #

echo "starting tests"

go test -v -coverprofile coverage.out -cover

exit 0

if [[ $? == 0 ]]; then
	go tool cover -html=coverage.out
fi