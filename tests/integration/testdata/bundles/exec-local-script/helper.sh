#!/usr/bin/env bash
set -euo pipefail

action=${1:-}

case "$action" in
  install)
    echo "Installing via helper script"
    ;;
  upgrade)
    echo "Upgrading via helper script"
    ;;
  uninstall)
    echo "Uninstalling via helper script"
    ;;
  *)
    echo "Unknown action: $action"
    exit 1
    ;;
esac
