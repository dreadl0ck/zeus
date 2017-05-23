#!/bin/bash

# {zeus}
# dependencies: configure
# outputs: bin/zeus
# description: build project
# arguments: 
#     - name:String
# buildNumber: true
# help: |
#     zeus build script
#     this script produces the zeus binary
#     it will be be placed in bin/$name
# {zeus}

echo "building $name"
godep go build -o bin/$name
