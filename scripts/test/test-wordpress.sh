#!/usr/bin/env bash

set -euo pipefail
export REGISTRY=${REGISTRY:-$USER}
export PORTER_HOME=${PORTER_HOME:-bin}
export KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}
# Run tests at the root of the repository
export TEST_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )"
pushd ${TEST_DIR}
trap popd EXIT

# Verify a bundle with dependencies
cp build/testdata/bundles/wordpress/porter.yaml .
sed -i "s/porter-wordpress:latest/${REGISTRY}\/porter-wordpress:latest/g" porter.yaml

${PORTER_HOME}/porter build
${PORTER_HOME}/porter install --insecure --cred ci --debug
cat ${PORTER_HOME}/claims/wordpress.json
