#!/usr/bin/env bash

set -euo pipefail
export REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )"
export PORTER_HOME=${PORTER_HOME:-$REPO_DIR/bin}
# Run tests in a temp directory
export TEST_DIR=/tmp/porter/hello
mkdir -p ${TEST_DIR}
pushd ${TEST_DIR}
trap popd EXIT

# Verify our default template bundle
${PORTER_HOME}/porter create

${PORTER_HOME}/porter build
${PORTER_HOME}/porter install --debug
${PORTER_HOME}/porter installation show HELLO --debug
${PORTER_HOME}/porter upgrade --debug
${PORTER_HOME}/porter installation show HELLO --debug
${PORTER_HOME}/porter uninstall --debug
${PORTER_HOME}/porter installation show HELLO --debug

# Publish bundle
${PORTER_HOME}/porter publish --tag localhost:5000/porter-hello:v0.1.0
