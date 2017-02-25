#!/bin/bash

# ---------------------------------------------------------------------- #
# @zeus-chain: clean
# @zeus-help: start tests
# ---------------------------------------------------------------------- #
# zeus test script
# 
# ---------------------------------------------------------------------- #

echo "starting tests"

go test -v -coverprofile coverage.out -cover
if [[ $? == 0 ]]; then
	go tool cover -html=coverage.out
fi