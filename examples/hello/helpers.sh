#!/usr/bin/env bash
set -euo pipefail

name=$(cat /cnab/app/foo/name.txt)

install() {
  echo "Hello, $name"
}

upgrade() {
  echo "Hello, $name"
}

uninstall() {
  echo "Goodbye, $name"
}

# Call the requested function and pass the arguments as-is
"$@"
