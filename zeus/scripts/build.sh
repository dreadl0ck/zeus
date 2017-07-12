#!/bin/bash

echo "building $name"
godep go build -o bin/$name
