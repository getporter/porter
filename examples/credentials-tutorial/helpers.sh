#!/usr/bin/env bash
set -euo pipefail

getUser() {
  url="https://api.github.com/user"
  if [[ "$GITHUB_USER" != "" ]]; then
    url="https://api.github.com/users/$GITHUB_USER"
  fi
  curl -s -H "Accept: application/vnd.github.v3+json" -H "Authorization: token $GITHUB_TOKEN" $url
}

# Call the requested function and pass the arguments as-is
"$@"
