#!/usr/bin/env bash
set -euo pipefail

install() {
  echo Hello World
  exit 1
}

upgrade() {
  echo World 2.0
}

uninstall() {
  echo Goodbye World
}

# Call the requested function and pass the arguments as-is
"$@"
