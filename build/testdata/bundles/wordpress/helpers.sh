#!/usr/bin/env bash
set -euo pipefail

install() {
  mkdir -p /cnab/app/outputs
  echo "topsecret-blog" >> /cnab/app/outputs/wordpress-password
}

ping() {
  echo ping
}

# Call the requested function and pass the arguments as-is
"$@"
