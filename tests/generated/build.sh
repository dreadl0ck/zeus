#!/bin/bash


test='bla'

#!/bin/bash

function yolo() {
	echo "yolo"
}
#!/bin/bash

# {zeus}
# arguments:
# description: prepare JS and CSS and move assets into wiki/docs
# dependencies: 
# outputs:
# help: Generate Javascript and CSS
# {zeus}

echo "copying LICENSE and README.md"

cp -f LICENSE wiki/docs
cp -f README.md wiki/docs

jsobfus -d frontend/src/js/:frontend/dist/js
sasscompile -d frontend/src/sass:frontend/dist/css

echo "building $binaryName for current OS"
rice embed-go
godep go build -o ${buildDir}/$binaryName
