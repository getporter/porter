#!/usr/bin/env bash
set -euo pipefail

# Print the directory parameter and copy it to the directory output
copy-dir() {
  cat /cnab/app/mydir/a.txt /cnab/app/mydir/sub/b.txt
  cp -r /cnab/app/mydir /cnab/app/outdir
}

noop() {
  echo noop
}

# Call the requested function and pass the arguments as-is
"$@"
