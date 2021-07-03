#!/usr/bin/env bash
set -euo pipefail

# PERMALINK and VERSION must be set before calling this script
# It is intended to only be executed by make publish

if [[ "$PERMALINK" = "canary" ]]; then
  PORTER_PERMALINK=$PERMALINK
else
  PORTER_PERMALINK=$VERSION
fi

ls -R bin

sed -e "s|PORTER_PERMALINK:-latest|PORTER_PERMALINK:-$PORTER_PERMALINK|g" -e "s|PKG_PERMALINK:-latest|PKG_PERMALINK:-$PERMALINK|" scripts/install/install-mac.sh > bin/$VERSION/install-mac.sh
sed -e "s|PORTER_PERMALINK:-latest|PORTER_PERMALINK:-$PORTER_PERMALINK|g" -e "s|PKG_PERMALINK:-latest|PKG_PERMALINK:-$PERMALINK|" scripts/install/install-linux.sh > bin/$VERSION/install-linux.sh
sed -e "s|PORTER_PERMALINK:-latest|PORTER_PERMALINK:-$PORTER_PERMALINK|g" -e "s|PKG_PERMALINK:-latest|PKG_PERMALINK:-$PERMALINK|" scripts/install/install-mac-arm64.sh > bin/$VERSION/install-mac-arm64.sh
sed -e "s|PORTER_PERMALINK:-latest|PORTER_PERMALINK:-$PORTER_PERMALINK|g" -e "s|PKG_PERMALINK:-latest|PKG_PERMALINK:-$PERMALINK|" scripts/install/install-linux-arm64.sh > bin/$VERSION/install-linux-arm64.sh
sed -e "s|PORTER_PERMALINK='latest'|PORTER_PERMALINK='$PORTER_PERMALINK'|g" -e "s|PKG_PERMALINK='latest'|PKG_PERMALINK='$PERMALINK'|g" scripts/install/install-windows.ps1 > bin/$VERSION/install-windows.ps1
