#!/usr/bin/env bash
set -euo pipefail

dump-account-key() {
  echo "Here is a the storage account key (base64 encoded) ==> $(echo $1 | base64)"
}

# Call the requested function and pass the arguments as-is
"$@"
