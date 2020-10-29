#!/usr/bin/env bash
set -euo pipefail

ping() {
  echo ping
}

# Call the requested function and pass the arguments as-is
"$@"
