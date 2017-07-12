#!/bin/bash

# remove assets to force rice debug stubs
rm -f rice-box.go

# install binary
go install
