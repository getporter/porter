#!/usr/bin/env bash
set -euo pipefail

install() {
  mkdir -p /cnab/app/outputs
  echo "topsecret" >> /cnab/app/outputs/mysql-root-password
  echo "moresekrets" >> /cnab/app/outputs/mysql-password
}

ping() {
  echo ping
}

# Call the requested function and pass the arguments as-is
"$@"
