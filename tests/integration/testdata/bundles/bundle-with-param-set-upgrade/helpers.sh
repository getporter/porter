#!/usr/bin/env bash
set -euo pipefail

install() {
  echo "install: $1"
}

upgrade() {
  echo "upgrade: $1"
}

uninstall() {
  echo "uninstall: $1"
}

"$@"
