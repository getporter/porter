#!/usr/bin/env bash
set -euo pipefail


# PERMALINK and VERSION must be set before calling this script
# It is intended to only be executed by make publish

if [[ "$PERMALINK" == "latest" ]]; then
  docker push getporter/porter:$VERSION
  docker push getporter/workshop:$VERSION
  docker push getporter/porter:$PERMALINK
  docker push getporter/workshop:$PERMALINK
else
  docker push getporter/porter:$PERMALINK
  docker push getporter/workshop:$PERMALINK
fi
