#!/usr/bin/env bash
set -euo pipefail


# REGISTRY, PERMALINK and VERSION must be set before calling this script
# It is intended to only be executed by make publish

docker build -t $REGISTRY/porter:$VERSION -f build/images/client/Dockerfile .
docker build -t $REGISTRY/porter-agent:$VERSION --build-arg PORTER_VERSION=$VERSION --build-arg REGISTRY=$REGISTRY -f build/images/agent/Dockerfile build/images/agent
docker build -t $REGISTRY/workshop:$VERSION -f build/images/workshop/Dockerfile .

docker tag $REGISTRY/porter:$VERSION $REGISTRY/porter:$PERMALINK
docker tag $REGISTRY/porter-agent:$VERSION $REGISTRY/porter-agent:$PERMALINK
docker tag $REGISTRY/workshop:$VERSION $REGISTRY/workshop:$PERMALINK
