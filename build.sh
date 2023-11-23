#!/bin/bash
DIST_DIR=dist/linux
SRC_DIR=src

# clean dist dir
if [ -d "$DIST_DIR" ]; then
    rm -rf $DIST_DIR
fi
mkdir $DIST_DIR

# build
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o $DIST_DIR/rails ./$SRC_DIR

# check if build was successful
if [ $? -eq 0 ]; then
    echo "Build successful."
else
    echo "Build failed."
    exit 1  # Exit script with an error status
fi
