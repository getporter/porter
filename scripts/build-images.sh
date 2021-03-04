#!/usr/bin/env bash
set -euo pipefail


# PERMALINK and VERSION must be set before calling this script
# It is intended to only be executed by make publish

docker build -t getporter/porter:$VERSION -f build/images/client/Dockerfile .
docker build -t getporter/workshop:$VERSION -f build/images/workshop/Dockerfile .

docker tag getporter/porter:$VERSION getporter/porter:$PERMALINK
docker tag getporter/workshop:$VERSION getporter/workshop:$PERMALINK
