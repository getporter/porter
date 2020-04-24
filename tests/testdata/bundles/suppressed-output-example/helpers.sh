#!/usr/bin/env bash

install() {
  echo '{"greeting": "Hello World!"}'
}

log-error() {
  >&2 echo "Error!"
}

uninstall() {
  echo "Farewell World!"
}

# Call the requested function and pass the arguments as-is
"$@"