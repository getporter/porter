#!/usr/bin/env bash
set -euo pipefail

# Return non-zero for a doc only build, and 0 for a builds that touch code.
DOCS_REGEX='(LICENSE|netlify.toml)|(\.md$)|(^docs/)|(^.github/)|(^.workshop/)'
if [[ -z "$(git diff --name-only HEAD HEAD~ | grep -vE $DOCS_REGEX)" ]]; then
  echo "This is a doc-only build"
  echo "##vso[task.setvariable variable=DOCS_ONLY;isOutput=true]true"
  exit 1
else
  echo "A full build must be run, code has been changed"
  echo "##vso[task.setvariable variable=DOCS_ONLY;isOutput=true]false"
  exit 0
fi
