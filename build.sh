#!/bin/bash

MOD_NAME="rails"
VERSION_VAR_PATH="$MOD_NAME/internal/utils.Version"

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENTRY_POINT="$PROJECT_ROOT/cmd/$MOD_NAME"
DIST_DIR="$PROJECT_ROOT/dist"
BIN_DIR="$DIST_DIR/bin"

# Stop on error
set -e
set -o pipefail

# Clean dist directory
if [ -d "$DIST_DIR" ]; then
  rm -rf "$DIST_DIR"
fi
mkdir -p "$DIST_DIR"
echo "Cleaned dist directory."

# Set the version if it's not set
if [ -z "$VERSION" ]; then
  VERSION="vX.X.X"
fi
echo "Output version set to $VERSION"

# Function to build a binary, given a GOOS and GOARCH, and return the output path.
build() {
  export GOOS="$1"; export GOARCH="$2"
  OUTPUT_PATH="$BIN_DIR/$MOD_NAME-$1-$2"
  if [ "$1" == "windows" ]; then
    OUTPUT_PATH="$OUTPUT_PATH.exe"
  fi
  go build -ldflags="-X '$VERSION_VAR_PATH=$VERSION'" -o "$OUTPUT_PATH" "$ENTRY_POINT"
  echo "$OUTPUT_PATH" # Return the output path
}

# Run tailwindcss build
TAIL_INPUT="./internal/input.css"
TAIL_OUTPUT="./internal/public/css/output.css"
TAIL_CONFIG_PATH="./internal/tailwind.config.js"
npx tailwindcss --config "$TAIL_CONFIG_PATH" -i "$TAIL_INPUT" -o "$TAIL_OUTPUT" --minify

# If -dev flag is set, build the binary for the current platform and capture the output path without printing.
if [ "$1" == "-dev" ]; then
  BIN_PATH=$(build "$(go env GOOS)" "$(go env GOARCH)" > /dev/null)
  echo "Successfully built $MOD_NAME for $(go env GOOS) $(go env GOARCH)."
  exit 0
fi

# Copy and create a .zip release for each platform, including the LICENSE, README, binary, and /internal/public directory.
for PLATFORM in linux windows darwin; do
  for ARCH in amd64; do
    # Build the binary and capture the output path
    BIN_PATH=$(build "$PLATFORM" "$ARCH")

    # Set the output zip file path
    ZIP_PATH="$DIST_DIR/$MOD_NAME-$VERSION-$PLATFORM-$ARCH.zip"
    
    # Zip the files
    zip -j "$ZIP_PATH" "./LICENSE.md" "./README.adoc" "$BIN_PATH"
    cd "$PROJECT_ROOT/internal"
    zip -r "$ZIP_PATH" "./public"
    cd "$PROJECT_ROOT"
    
    echo "Zipped $MOD_NAME for $PLATFORM $ARCH."
  done
done
