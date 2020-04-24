#!/usr/bin/env bash
set -euo pipefail

install() {
  echo Using Magic Password: $1
}

upgrade() {
  echo World is now at 2.0
}

uninstall() {
  echo Goodbye World
}

# Call the requested function and pass the arguments as-is
"$@"
