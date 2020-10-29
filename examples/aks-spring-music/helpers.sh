#!/usr/bin/env bash
set -euo pipefail

uninstall() {
  echo App should be uninstalled here, but it is not
}

# Call the requested function and pass the arguments as-is
"$@"
