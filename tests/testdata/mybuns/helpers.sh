#!/usr/bin/env bash
set -euo pipefail

install() {
  touch /cnab/app/outputs/tfstate
  if [[ "$LOG_LEVEL" == "11" ]]; then
    echo Hello, "$ROOT_USERNAME"!
  fi
}

makeMagic() {
  echo "$1" > /cnab/app/magic.txt
}

ensureMagic() {
  if ! test -f "magic.txt"; then
    echo "No magic detected"
    exit 1
  fi
}

upgrade() {
  if [[ "$LOG_LEVEL" == "11" ]]; then
    echo World 2.0
  fi
}

uninstall() {
  if [[ "$LOG_LEVEL" == "11" ]]; then
    echo Goodbye World
  fi
}

chaos_monkey() {
  if [[ "$1" == "true" ]]; then
    echo "a chaos monkey appears. you have died"
    exit 1
  fi

    echo "no chaos monkey appeared. you have lived"

}

# Call the requested function and pass the arguments as-is
"$@"
