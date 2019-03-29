#!/usr/bin/env bash

set -euo pipefail
export REGISTRY=${REGISTRY:-$USER}
export PORTER_HOME=${PORTER_HOME:-bin}
# Run tests at the root of the repository
export TEST_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )"
pushd ${TEST_DIR}
trap popd EXIT

# Verify our default template bundle
${PORTER_HOME}/porter create
sed -i "s/porter-hello:latest/${REGISTRY}\/porter-hello:latest/g" porter.yaml

${PORTER_HOME}/porter build
${PORTER_HOME}/porter install --insecure --debug
cat ${PORTER_HOME}/claims/HELLO.json
