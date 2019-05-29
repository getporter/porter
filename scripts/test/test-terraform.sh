#!/usr/bin/env bash

set -euo pipefail
export REGISTRY=${REGISTRY:-$USER}
export REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )"
export PORTER_HOME=${PORTER_HOME:-$REPO_DIR/bin}
# Run tests in a temp directory
export TEST_DIR=/tmp/porter/terraform
mkdir -p ${TEST_DIR}
pushd ${TEST_DIR}
trap popd EXIT

# Populate cnab/app with terraform assets
cp -r ${REPO_DIR}/build/testdata/bundles/terraform/cnab .

# Copy in the terraform porter manifest
cp ${REPO_DIR}/build/testdata/bundles/terraform/porter.yaml .
sed -i "s/porter-terraform:latest/${REGISTRY}\/porter-terraform:latest/g" porter.yaml

porter_output=$(mktemp)

${PORTER_HOME}/porter build

${PORTER_HOME}/porter install --insecure --debug --param file_contents='foo!' 2>&1 | tee ${porter_output}

if ! cat ${porter_output} | grep -q 'content:  "" => "foo!"'; then
  echo "ERROR: File contents not created properly"
  exit 1
fi

# TODO: enable when status supported
# ${PORTER_HOME}/porter status --debug | grep -q 'content = foo!'

# TODO: enable when upgrade supported
# ${PORTER_HOME}/porter upgrade --insecure --debug --param file_contents='bar!'

# TODO: enable when status supported
# ${PORTER_HOME}/porter status --debug | grep -q 'content = bar!'

cat ${PORTER_HOME}/claims/porter-terraform.json

${PORTER_HOME}/porter uninstall --insecure --debug --param file_contents='foo!'
