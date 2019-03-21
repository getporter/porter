#!/usr/bin/env bash

REGISTRY=${REGISTRY:-$USER}
KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}
PORTER_HOME=${PORTER_HOME:-bin}

# Verify our default template bundle
./bin/porter create
sed -i "s/porter-hello:latest/${REGISTRY}\/porter-hello:latest/g" porter.yaml
./bin/porter build
./bin/porter install --insecure
