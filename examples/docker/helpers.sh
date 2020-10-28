#!/usr/bin/env bash
set -euo pipefail

whalesay() {
  docker run --rm docker/whalesay:latest cowsay $1
}

install() {
  whalesay "Hello World"
}

upgrade() {
  whalesay "World 2.0"
}

uninstall() {
  whalesay "Goodbye World"
}

# Call the requested function and pass the arguments as-is
"$@"
