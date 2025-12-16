#!/usr/bin/env bash
set -euo pipefail

action=${1:-}

case "$action" in
  install)
    echo "Installing via helper script (with ./ prefix)"
    ;;
  upgrade)
    echo "Upgrading via helper script (with ./ prefix)"
    ;;
  uninstall)
    echo "Uninstalling via helper script (with ./ prefix)"
    ;;
  *)
    echo "Unknown action: $action"
    exit 1
    ;;
esac
