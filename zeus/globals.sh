#!/bin/bash

# ---------------------------------------------------------------------- #
# zeus globals script
# if this script exists, it will be prepended to the commandscript
# 
# so everything defined here will be available to all commands
# this script is NOT being parsed for zeus header fields
# ---------------------------------------------------------------------- #

TEST=testglobalvar
TEST2=testglobalvar2

function yolo() {
	echo "yoloyolo yoo"
	say yolo
}
