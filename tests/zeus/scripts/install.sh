#!/bin/bash

echo "installing zeus"
rice embed-go
godep go install
