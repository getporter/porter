#!/usr/bin/env bash

set -euo pipefail
export REGISTRY=${REGISTRY:-$USER}
export REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )"
export PORTER_HOME=${PORTER_HOME:-$REPO_DIR/bin}
# Run tests in a temp directory
export TEST_DIR=/tmp/porter/hello
mkdir -p ${TEST_DIR}
pushd ${TEST_DIR}
trap popd EXIT

# Verify our default template bundle
${PORTER_HOME}/porter create

# Substitute REGISTRY in for invocation image and bundle tag
sed -i "s/porter-hello:latest/${REGISTRY}\/porter-hello:latest/g" porter.yaml
sed -i "s/deislabs\/porter-hello-bundle/${REGISTRY}\/porter-hello-bundle/g" porter.yaml

${PORTER_HOME}/porter build
${PORTER_HOME}/porter install --insecure --debug
cat ${PORTER_HOME}/claims/HELLO.json
${PORTER_HOME}/porter upgrade --insecure --debug
cat ${PORTER_HOME}/claims/HELLO.json
${PORTER_HOME}/porter uninstall --insecure --debug

# Publish bundle
${PORTER_HOME}/porter publish