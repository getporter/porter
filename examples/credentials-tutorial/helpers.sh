#!/usr/bin/env bash
set -euo pipefail

getUser() {
  curl -s -H "Accept: application/vnd.github.v3+json" -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
}

# Call the requested function and pass the arguments as-is
"$@"
