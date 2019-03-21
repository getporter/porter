#!/usr/bin/env bash

KUBECONFIG=${KUBECONFIG:-$(HOME)/.kube/config}
DUFFLE_HOME=${DUFFLE_HOME:-bin/.duffle}
PORTER_HOME=${PORTER_HOME:-bin}

# Verify a bundle with dependencies
cp build/testdata/bundles/wordpress/porter.yaml .
sed -i "s/porter-wordpress:latest/${REGISTRY}\/porter-wordpress:latest/g" porter.yaml
./bin/porter build
./bin/porter install PORTER-WORDPRESS -f bundle.json --credentials ci --insecure --home ${DUFFLE_HOME}
