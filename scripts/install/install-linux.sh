#!/usr/bin/env bash
set -euo pipefail

# Installs the porter CLI for a single user.
# PORTER_HOME:        Location where Porter is installed (defaults to ~/.porter).
# PORTER_MIRROR:      Base URL where Porter assets, such as binaries and atom feeds, are downloaded.
#                     This lets you setup an internal mirror.
# PORTER_VERSION:     The version of Porter assets to download.
# PORTER_BINARY_URL:  The URL to porter binary (https:// or file:// for local binary)

export PORTER_HOME=${PORTER_HOME:-~/.porter}
export PORTER_MIRROR=${PORTER_MIRROR:-https://cdn.porter.sh}
PORTER_VERSION=${PORTER_VERSION:-latest}
PORTER_BINARY_URL=${PORTER_BINARY_URL:-"$PORTER_MIRROR/$PORTER_VERSION/porter-linux-amd64"}

echo "Installing porter@$PORTER_VERSION to $PORTER_HOME from $PORTER_BINARY_URL"

mkdir -p "$PORTER_HOME/runtimes"

curl -fsSLo "$PORTER_HOME/porter" "$PORTER_BINARY_URL" 
chmod +x "$PORTER_HOME/porter"
cp "$PORTER_HOME/porter" "$PORTER_HOME/runtimes/porter-runtime"
echo Installed "$("$PORTER_HOME"/porter version)"

"$PORTER_HOME/porter" mixin install exec --version "$PORTER_VERSION"

echo "Installation complete."
echo "Add porter to your path by adding the following line to your ~/.profile and open a new terminal:"
echo "export PATH=\$PATH:~/.porter"
