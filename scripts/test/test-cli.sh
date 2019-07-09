#!/usr/bin/env bash

set -euo pipefail
export DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )"
export PORTER_HOME=${PORTER_HOME:-$REPO_DIR/bin}

${PORTER_HOME}/porter help
${PORTER_HOME}/porter version

${DIR}/test-hello.sh
# TODO: Temporarily disable the wordpress test because it relies on dependencies which are being rewritten
#${DIR}/test-wordpress.sh
${DIR}/test-terraform.sh
