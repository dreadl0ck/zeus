#!/bin/bash

echo "building for linux amd64"
rice embed-go
GOOS=linux GOARCH=amd64 godep go build -o ${buildDir}/zeus-linux
