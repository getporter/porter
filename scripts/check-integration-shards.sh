#!/usr/bin/env bash
# Fail when a tests/integration/*_test.go file is in no shard, or is listed
# in more than one, or a shard lists a file that doesn't exist.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
shards="$root/.github/integration-shards.json"

actual="$(cd "$root/tests/integration" && ls *_test.go | sed 's/\.go$//' | sort)"
listed="$(jq -r '.[].files[]' "$shards" | sort)"

status=0
if dups="$(echo "$listed" | uniq -d)" && [ -n "$dups" ]; then
  echo "Listed in more than one shard:" >&2; echo "$dups" >&2; status=1
fi
if missing="$(comm -23 <(echo "$actual") <(echo "$listed" | uniq))" && [ -n "$missing" ]; then
  echo "In no shard (add to .github/integration-shards.json):" >&2; echo "$missing" >&2; status=1
fi
if extra="$(comm -13 <(echo "$actual") <(echo "$listed" | uniq))" && [ -n "$extra" ]; then
  echo "Listed but no such file:" >&2; echo "$extra" >&2; status=1
fi
exit $status
