#!/usr/bin/env bash

set -euo pipefail
export DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

${PORTER_HOME}/porter help
${PORTER_HOME}/porter version

${DIR}/test-hello.sh
${DIR}/test-wordpress.sh
${DIR}/test-terraform.sh
