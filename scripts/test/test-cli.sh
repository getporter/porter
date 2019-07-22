#!/usr/bin/env bash

set -euo pipefail
export DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )"
export PORTER_HOME=${PORTER_HOME:-$REPO_DIR/bin}

${PORTER_HOME}/porter help
${PORTER_HOME}/porter version

${DIR}/test-hello.sh
${DIR}/test-terraform.sh
