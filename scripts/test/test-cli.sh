#!/usr/bin/env bash

set -euo pipefail
export DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )"
export PORTER_HOME=/tmp/porter

# Clean last test run
rm -fr ${PORTER_HOME}/* &>/dev/null | true
mkdir -p ${PORTER_HOME}/runtime

# Populate temporary porter home
BIN_DIR=${REPO_DIR}/bin/
cp ${BIN_DIR}/porter ${PORTER_HOME}/
cp ${BIN_DIR}/porter-runtime ${PORTER_HOME}/runtime
cp -R ${BIN_DIR}/mixins ${PORTER_HOME}/

${PORTER_HOME}/porter help
${PORTER_HOME}/porter version

${DIR}/test-hello.sh
