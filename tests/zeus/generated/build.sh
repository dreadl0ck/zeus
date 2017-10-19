#!/bin/bash
# generated by ZEUS v0.8.1
# Timestamp: [Wed Jul 12 18:11:13 2017]


binaryName="zeus"
buildDir="bin"
version="0.8.1"

#!/bin/bash

function bash_greet() {
	echo "hello world from bash!"
	echo "ZEUS version: $version"
}

echo "[ZEUS v${version}] copying LICENSE and README.md"

cp -f LICENSE wiki/docs
cp -f README.md wiki/docs

echo "[ZEUS v${version}] minifying javscript and css"
jsobfus -d frontend/src/js/:frontend/dist/js
sasscompile -d frontend/src/sass:frontend/dist/css

echo "[ZEUS v${version}] building ${buildDir}/${binaryName} for current OS"
rice embed-go
godep go build -o ${buildDir}/${binaryName}
