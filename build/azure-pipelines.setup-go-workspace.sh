#!/usr/bin/env bash

set -xeuo pipefail

# Create GOPATH/bin and add it to our PATH so that
# installed go binaries are available
mkdir -p /home/vsts/go/bin/
echo "##vso[task.prependpath]/home/vsts/go/bin/"
