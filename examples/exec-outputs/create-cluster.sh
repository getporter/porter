#!/usr/bin/env bash

# This pretends to create a kubernetes cluster
# by generating a dummy kubeconfig file
mkdir -p /root/.kube
cat <<EOF >> /root/.kube/config
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: abc123==
    server: https://127.0.0.1:8443
  name: minikube
contexts:
- context:
    cluster: minikube
    user: minikube
  name: minikube
current-context: minikube
kind: Config
preferences: {}
users:
- name: minikube
  user:
    client-certificate-data: abc123=
    client-key-data: abc123==
EOF
