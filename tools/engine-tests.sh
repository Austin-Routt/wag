#!/usr/bin/env bash
# Tests of the vendored engine: GoCHR, then carneades-4 built against the local
# GoCHR (as the schemes library's tools/vv/engine-tests.sh does). WAG's own
# tests never run these, and only these catch a lost engine fix (the empty
# cumulative weight, the YAML scalars, WAG's fork regressions in src/test).
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
echo "-- GoCHR"
(cd "$root/GoCHR" && go build ./... && go vet ./... && go test -count=1 ./...)
echo "-- carneades-4 (against the local GoCHR)"
cp "$root/carneades-4/go.mod" "$tmp/go.mod"
cp "$root/carneades-4/go.sum" "$tmp/go.sum"
go mod edit -replace "github.com/hfried/GoCHR=$root/GoCHR" "$tmp/go.mod"
m=(-mod=mod -modfile="$tmp/go.mod")
(cd "$root/carneades-4" && go build "${m[@]}" ./... && go vet "${m[@]}" ./... && go test "${m[@]}" -count=1 ./...)
