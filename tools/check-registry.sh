#!/usr/bin/env bash
# Keep tool failures separate from the one explicitly tolerated registry debt.
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$root/wag"
go build -o "$tmp/wag-cli" ./cmd/wag-cli
status=0
"$tmp/wag-cli" --registry ../docs/scheme-registry/walton-clean/walton.yml \
  --families data/schemes/families.json > "$tmp/stdout" 2> "$tmp/stderr" || status=$?
cat "$tmp/stdout"
cat "$tmp/stderr" >&2
python3 "$root/tools/check-registry-result.py" --status "$status" "$tmp/stdout" "$tmp/stderr"
