#!/usr/bin/env bash
set -euo pipefail

install() {
  echo Hello World
}

upgrade() {
  echo World 2.0
}

uninstall() {
  echo Goodbye World
}

zombies() {
  echo oh noes my brains
}

dump-myfile() {
  cat /cnab/app/myfile
}

# Call the requested function and pass the arguments as-is
"$@"
