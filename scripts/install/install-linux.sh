#!/usr/bin/env bash
set -euo pipefail
# Installs the Porter CLI for a single user.
# Environment Variables:
# PORTER_HOME:      Location where Porter is installed (defaults to ~/.porter).
# PORTER_MIRROR:    Base URL for downloading Porter assets (defaults to https://cdn.porter.sh).
# PORTER_VERSION:   The version of Porter assets to download (defaults to latest).
export PORTER_HOME=${PORTER_HOME:-~/.porter}
export PORTER_MIRROR=${PORTER_MIRROR:-https://cdn.porter.sh}
PORTER_VERSION=${PORTER_VERSION:-latest}
MONGO_IMAGE_FILE="mongo_image.tar"
echo "Installing porter@$PORTER_VERSION to $PORTER_HOME from $PORTER_MIRROR"
mkdir -p "$PORTER_HOME/runtimes"
if [[ -f "./porter-linux-amd64" ]]; then
    echo "Using existing porter-linux-amd64 binary from the current directory."
    cp "./porter-linux-amd64" "$PORTER_HOME/porter"
else
    echo "Downloading Porter from $PORTER_MIRROR/$PORTER_VERSION/porter-linux-amd64"
    curl -fsSLo "$PORTER_HOME/porter" "$PORTER_MIRROR/$PORTER_VERSION/porter-linux-amd64"
fi
chmod +x "$PORTER_HOME/porter"
cp "$PORTER_HOME/porter" "$PORTER_HOME/runtimes/porter-runtime"
echo "Installed Porter version: $("$PORTER_HOME/porter" version)"
if [[ -f "$MONGO_IMAGE_FILE" ]]; then
    echo "Loading MongoDB image from $MONGO_IMAGE_FILE..."
    docker load -i "$MONGO_IMAGE_FILE"
    if ! docker ps -q -f name=porter-mongodb-docker-plugin; then
        echo "Running MongoDB container..."
        docker run --name porter-mongodb-docker-plugin -d -p 27018:27017 -v mongodb_data:/data/db --restart always mongo:4.0-xenial
    else
        echo "MongoDB container is already running."
    fi
fi
echo "Installation complete."
echo "Add Porter to your PATH by adding the following line to your ~/.bashrc or ~/.bash_profile and open a new terminal:"
echo "export PATH=\$PATH:$PORTER_HOME"
