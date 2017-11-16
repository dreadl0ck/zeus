#!/bin/bash

if [[ $force != true ]]; then
	read -p "[INFO] Version ${version} correct? Hit [ENTER] to release."
fi

git tag -a "v${version}" -m "release v${version}"
git push origin "v${version}"

goreleaser

