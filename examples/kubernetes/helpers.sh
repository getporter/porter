#!/usr/bin/env bash
set -euo pipefail

dump-ip() {
  echo You will find the service at: $1
}

# Call the requested function and pass the arguments as-is
"$@"
