#!/bin/bash

# {zeus}
# arguments:
# description: clean up to prepare for build
# dependencies: 
# outputs:
# help: clears bin/ directory and deletes generated config & data in tests
# {zeus}

rm -rf bin/*
rm -rf tests/config.yml
rm -rf tests/data.yml
