#!/usr/bin/env bash

set -euo pipefail
export REGISTRY=${REGISTRY:-$USER}
export REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )"
export PORTER_HOME=${PORTER_HOME:-$REPO_DIR/bin}
export NAMESPACE="$(head /dev/urandom | tr -dc a-z0-9 | head -c 10 ; echo '')"
export KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}
# Run tests in a temp directory
export TEST_DIR=/tmp/porter/wordpress
mkdir -p ${TEST_DIR}
pushd ${TEST_DIR}
trap popd EXIT

# Verify a bundle with dependencies
cp ${REPO_DIR}/build/testdata/bundles/wordpress/porter.yaml .

# Substitute REGISTRY in for invocation image and bundle tag
sed -i "s/porter-wordpress:latest/${REGISTRY}\/porter-wordpress:latest/g" porter.yaml
sed -i "s/deislabs\/porter-wordpress-bundle/${REGISTRY}\/porter-wordpress-bundle/g" porter.yaml

${PORTER_HOME}/porter build

# Create temp file for install output, to search after action has completed
install_log=$(mktemp)
sensitive_value=${RANDOM}-value

# Piping both stderr and stdout to log as debug logs may flow via stderr
${PORTER_HOME}/porter install --insecure --cred ci \
    --param wordpress-password="${sensitive_value}" \
    --param namespace=$NAMESPACE \
    --param wordpress-name="porter-ci-wordpress-$NAMESPACE" \
    --param mysql-name="porter-ci-mysql-$NAMESPACE" \
    --debug 2>&1 | tee ${install_log}

# Be sure that sensitive data is masked
if cat ${install_log} | grep -q "${sensitive_value}"; then
  echo "ERROR: Sensitive parameter value (wordpress-password) not masked in console output"
  exit 1
fi

echo "Verifing bundle outputs..."
# TODO: porter bundle show wordpress to see outputs listing
${PORTER_HOME}/porter outputs list -b wordpress | grep "wordpress-password"
${PORTER_HOME}/porter output show -n wordpress-password -b wordpress | grep "${sensitive_value}"

cat ${PORTER_HOME}/claims/wordpress.json

${PORTER_HOME}/porter uninstall --insecure --cred ci --debug
kubectl delete ns $NAMESPACE

# Publish bundle
${PORTER_HOME}/porter publish
