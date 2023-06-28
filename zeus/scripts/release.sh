#!/bin/bash

read -p "[INFO] new Version ${version} correct? Hit [ENTER] to release."

echo "[INFO] creating new git tag"
git tag -a "v${version}" -m "release v${version}"

echo "[INFO] pushing new git tag"
git push origin "v${version}"

echo "[INFO] running goreleaser"
goreleaser --clean

echo "[INFO] done"