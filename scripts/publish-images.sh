#!/usr/bin/env bash
set -euo pipefail

# PERMALINK and VERSION must be set before calling this script
# It is intended to only be executed by make publish

if [[ "$PERMALINK" == "latest" ]]; then
  docker build --build-arg PERMALINK=$VERSION -t getporter/porter:$VERSION build/images/client
  docker build --build-arg PERMALINK=$VERSION -t getporter/workshop:$VERSION build/images/workshop

  docker tag getporter/porter:$VERSION getporter/porter:$PERMALINK
  docker tag getporter/workshop:$VERSION getporter/workshop:$PERMALINK

  docker push getporter/porter:$VERSION
  docker push getporter/workshop:$VERSION
  docker push getporter/porter:$PERMALINK
  docker push getporter/workshop:$PERMALINK
else
  docker build --build-arg PERMALINK=$PERMALINK -t getporter/porter:$PERMALINK build/images/client
  docker build --build-arg PERMALINK=$PERMALINK -t getporter/workshop:$PERMALINK build/images/workshop

  docker push getporter/porter:$PERMALINK
  docker push getporter/workshop:$PERMALINK
fi