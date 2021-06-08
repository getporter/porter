#!/usr/bin/env bash
set -euo pipefail

install() {
  echo Hello, $1
}

upgrade() {
  echo $1 2.0
}

uninstall() {
  echo Goodbye, $1
}

# Call the requested function and pass the arguments as-is
"$@"
