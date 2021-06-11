#!/usr/bin/env bash
set -euo pipefail


# REGISTRY, PERMALINK and VERSION must be set before calling this script
# It is intended to only be executed by make publish

if [[ "$PERMALINK" == *latest ]]; then
  docker push $REGISTRY/porter:$VERSION
  docker push $REGISTRY/porter-agent:$VERSION
  docker push $REGISTRY/workshop:$VERSION

  docker push $REGISTRY/porter:$PERMALINK
  docker push $REGISTRY/porter-agent:$PERMALINK
  docker push $REGISTRY/workshop:$PERMALINK
else
  docker push $REGISTRY/porter:$PERMALINK
  docker push $REGISTRY/porter-agent:$PERMALINK
  docker push $REGISTRY/workshop:$PERMALINK
fi
