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

# Substitute REGISTRY in for the bundle tag
sed -i "s/getporter\/porter-hello/${REGISTRY}\/porter-hello/g" porter.yaml

${PORTER_HOME}/porter build
${PORTER_HOME}/porter install --debug
cat ${PORTER_HOME}/claims/HELLO.json
${PORTER_HOME}/porter upgrade --debug
cat ${PORTER_HOME}/claims/HELLO.json
${PORTER_HOME}/porter uninstall --debug

# Publish bundle
${PORTER_HOME}/porter publish
