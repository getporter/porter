#!/usr/bin/env bash
set -euo pipefail

PORTER_MIRROR=${PORTER_MIRROR:-https://cdn.porter.sh}
PORTER_VERSION=${PORTER_VERSION:-latest}
MONGO_IMAGE_FILE="mongo_image.tar"
MONGO_IMAGE_URL="docker.io/library/mongo:8.0-noble"
BUNDLE_NAME="porter-air-gapped-install-$PORTER_VERSION"
DOWNLOAD_DIR="/tmp/$BUNDLE_NAME"

PORTER_BINARY_URL=${PORTER_BINARY_URL:-"$PORTER_MIRROR/$PORTER_VERSION/porter-linux-amd64"}
PORTER_SCRIPT_URL=${PORTER_SCRIPT_URL:-"$PORTER_MIRROR/$PORTER_VERSION/install-linux.sh"}

mkdir -p $DOWNLOAD_DIR

echo "Pulling MongoDB image..."
docker pull "$MONGO_IMAGE_URL"
echo "Saving MongoDB image to $MONGO_IMAGE_FILE..."
docker save -o "$DOWNLOAD_DIR/$MONGO_IMAGE_FILE" "$MONGO_IMAGE_URL"

echo "Downloading Porter Install Script from $PORTER_SCRIPT_URL"
curl -fsSLo "$DOWNLOAD_DIR/install-linux.sh" "$PORTER_SCRIPT_URL"

echo "Downloading Porter Binary from $PORTER_BINARY_URL"
curl -fsSLo "$DOWNLOAD_DIR/porter-linux-amd64" "$PORTER_BINARY_URL" 

cat << EOF > $DOWNLOAD_DIR/install-bundle.sh 
export PORTER_BINARY_URL="file://\$(realpath ./porter-linux-amd64)" 
export PORTER_SCRIPT_URL="file://\$(realpath ./install-linux.sh)" 
bash install-linux.sh
docker load -i "$MONGO_IMAGE_FILE"
if ! docker ps -q -f name=porter-mongodb-docker-plugin; then
  echo "Running MongoDB container..."
  docker run --name porter-mongodb-docker-plugin -d -p 27018:27017 -v mongodb_data:/data/db --restart always mongo:8.0-noble
else
  echo "MongoDB container is already running."
fi
EOF


TAR_FILE="$BUNDLE_NAME.tar.gz"
echo "Creating tarball: $TAR_FILE..."
tar -czvf "$TAR_FILE" -C /tmp "$(basename "$DOWNLOAD_DIR")" 

echo "Bundle created: $TAR_FILE"
echo "Usage Instructions:"
echo "tar -xzf $TAR_FILE -C ."
echo "cd $BUNDLE_NAME"
echo "bash install-bundle.sh"

exit 0

