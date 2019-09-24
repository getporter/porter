#!/usr/bin/env bash

set -xeuo pipefail

mkdir -p "$GOBIN"
mkdir -p "$GOPATH/pkg"
mkdir -p "$MODULE_PATH"
shopt -s extglob
mv !(gopath) "$MODULE_PATH"
echo "##vso[task.prependpath]$GOBIN"
echo "##vso[task.prependpath]$GOROOT/bin"