#!/usr/bin/env bash
set -euo pipefail

if [[ ! -f "/run/secrets/token" ]]; then
    echo "You forgot to use --secret id=token,src=secrets/token"
    exit 1
fi
