#!/usr/bin/env bash
set -euo pipefail

# VERSION must be set before calling this script
# It is intended to only be executed by make publish
ls -R bin

sed -e "s|PORTER_VERSION:-latest|PORTER_VERSION:-$VERSION|g" scripts/install/install-mac.sh > bin/$VERSION/install-mac.sh
sed -e "s|PORTER_VERSION:-latest|PORTER_VERSION:-$VERSION|g" scripts/install/install-linux.sh > bin/$VERSION/install-linux.sh
sed -e "s|PORTER_VERSION='latest'|PORTER_VERSION='$VERSION'|g" scripts/install/install-windows.ps1 > bin/$VERSION/install-windows.ps1
