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
MONGO_IMAGE_URL="docker.io/library/mongo:4.0-xenial"
BUNDLE_NAME="porter-offline-install-$PORTER_VERSION"
DOWNLOAD_DIR="/tmp/$BUNDLE_NAME"

mkdir -p $DOWNLOAD_DIR

echo "Pulling MongoDB image..."
docker pull "$MONGO_IMAGE_URL"
echo "Saving MongoDB image to $MONGO_IMAGE_FILE..."
docker save -o "$DOWNLOAD_DIR/$MONGO_IMAGE_FILE" "$MONGO_IMAGE_URL"
echo "Downloading Porter from $PORTER_MIRROR/$PORTER_VERSION/install-linux.sh"
curl -fsSLo "$DOWNLOAD_DIR/install-linux.sh" "$PORTER_MIRROR/$PORTER_VERSION/install-linux.sh"
echo "Downloading Porter from $PORTER_MIRROR/$PORTER_VERSION/porter-linux-amd64"
curl -fsSLo "$DOWNLOAD_DIR/porter-linux-amd64" "$PORTER_MIRROR/$PORTER_VERSION/porter-linux-amd64"

TAR_FILE="$BUNDLE_NAME.tar.gz"
echo "Creating tarball: $TAR_FILE..."
tar -czvf "$TAR_FILE" -C /tmp "$(basename "$DOWNLOAD_DIR")" 

echo "Bundle created: $TAR_FILE"
echo "Usage Instructions:"
echo "tar -xzf $TAR_FILE -C ."
echo "cd $BUNDLE_NAME"
echo "bash install-linux.sh"
exit 0
