#!/usr/bin/env bash
set -euo pipefail

install() {
  echo "Run 'porter invoke --action=help' next to see sfctl in action"
}

upgrade() {
  echo World is now at 2.0
}

uninstall() {
  echo Goodbye World
}

# Call the requested function and pass the arguments as-is
"$@"
