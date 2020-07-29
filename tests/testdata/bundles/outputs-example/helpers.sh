#!/usr/bin/env bash
set -euo pipefail

dump-config() {
  echo '{"user": "sally"}'
}

dump() {
  echo $1
}

# Call the requested function and pass the arguments as-is
"$@"
