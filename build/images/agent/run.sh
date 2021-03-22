#!/usr/bin/env sh

set -euo pipefail

# Copy user-defined porter configuration into PORTER_HOME
if test -n "$(find /porter-config -maxdepth 1 -name 'config.*' -print -quit)"; then
  echo "loading porter configuration..."
  cp -L /porter-config/config.* /root/.porter/
  ls | grep config.*
  cat /root/.porter/config.*
fi

# Print the version of porter we are using for this run
porter version

# Execute the command passed
echo "porter $@"
porter $@
