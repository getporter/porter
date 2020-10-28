#!/usr/bin/env bash
set -euo pipefail

dump-config() {
  echo '{"user": "sally"}'
}

dump() {
  echo $1
}

assert-output-value() {
  if [ "$1" != "$2" ]; then
    echo "'$1' does not match the expected output value of '$2'."
    exit 1
  fi
}

# Call the requested function and pass the arguments as-is
"$@"
