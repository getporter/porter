#!/usr/bin/env bash

set -euo pipefail

create-cluster() {
echo "Creating the cluster..."
# This pretends to create a kubernetes cluster
# by generating a dummy kubeconfig file
mkdir -p /home/nonroot/.kube
cat <<EOF >> /home/nonroot/.kube/config
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
}

ensure-config() {
    if [ ! -f "/home/nonroot/.kube/config" ]; then
      echo "kubeconfig not found"
      exit 1
    fi
}

ensure-users() {
    if [ ! -f "/cnab/app/users.json" ]; then
      echo "users.json not found"
      exit 1
    fi
}

generate-users() {
    ensure-config
    echo '{"users": ["sally"]}' > users.json
    cat users.json
}

add-user() {
  ensure-config
  ensure-users
  # Do this in two steps because jq doesn't have overwrite functionality
  cat users.json | jq ".users += [\"$1\"]" > users.json.tmp
  # Using cp instead of mv because if the bundle is executed with Kubernetes
  # you can't move a file that was volume mounted in k8s. That only works with Docker.
  cp users.json.tmp users.json
}

dump-users() {
    ensure-config
    ensure-users
    cat users.json
}

uninstall() {
  ensure-config
  echo 'Uninstalling Cluster...'
}

# Call the requested function and pass the arguments as-is
"$@"