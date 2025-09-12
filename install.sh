#!/bin/bash

# exit on error
set -e

if [ -n "$PYTHONHOME" ]; then
    export PKG_CONFIG_PATH="$PYTHONHOME/lib/pkgconfig:$PKG_CONFIG_PATH"
fi

echo "Installing llpyg..."

cd ./_xtool
llgo install ./...

cd ../
go install -v ./cmd/...

echo "llpyg is now available in your GOPATH."
