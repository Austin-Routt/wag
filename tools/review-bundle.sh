#!/usr/bin/env bash
# A review bundle: one commit of WAG with everything its checks need offline,
# so that a reviewer with no network (no Go module proxy) can run every check
# CI runs (tools/check-all.sh), the Go suites of all four modules included.
#
#   bash tools/review-bundle.sh [COMMIT] [OUT.zip]
#
# COMMIT defaults to HEAD; OUT to ./wag-review-<commit>.zip. Making the bundle
# needs the network once: the Go modules (filled into a fresh module cache by
# building every package and test), and the upstream walton.yml the cleaning
# was written for (MPL-2.0, checked against its SHA-256). It needs git, go 1.24
# or newer, python3, curl and zip.
#
# The bundle holds:
#   repo/                  the commit, with one commit of history (git rev-parse HEAD names it)
#   offline/goproxy/       the module downloads, served as a file proxy (GOPROXY=file://...)
#   offline/upstream/      the upstream walton.yml (its licence: repo/docs/scheme-registry/walton-clean/LICENSE)
#   run-offline.sh         runs tools/check-all.sh with no network
#   README.md              what it is and how to run it
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
commit="$(git -C "$root" rev-parse --verify "${1:-HEAD}^{commit}")"
short="${commit:0:12}"
out="${2:-$PWD/wag-review-$short.zip}"
case "$out" in /*) ;; *) out="$PWD/$out" ;; esac
upstream_url=https://raw.githubusercontent.com/basakkurtuldu/walton-argumentation-schemes-for-carneades-4/master/walton.yml
upstream_sha=1976202ac6bd538526ea75488659ec9880641ec03534d6e098f6f6dd1bcc099c

work="$(mktemp -d)"
trap 'chmod -R u+w "$work" 2>/dev/null; rm -rf "$work"' EXIT
name="wag-review-$short"
b="$work/$name"
mkdir -p "$b/offline/upstream"

echo "-- the commit $commit"
git init -q "$b/repo"
git -C "$b/repo" fetch -q --depth 1 "file://$root" "$commit"
git -C "$b/repo" checkout -q --detach FETCH_HEAD
test "$(git -C "$b/repo" rev-parse HEAD)" = "$commit"

echo "-- the Go modules (every package and test of the four modules is built)"
export GOMODCACHE="$work/modcache" GOTOOLCHAIN=local GOWORK=off GOFLAGS=
(cd "$b/repo/wag" && go vet ./... && go test -count=1 -run '^$' ./... > /dev/null)
(cd "$b/repo/GoCHR" && go vet ./... && go test -count=1 -run '^$' ./... > /dev/null)
modfile="$work/carneades.go.mod"
cp "$b/repo/carneades-4/go.mod" "$modfile"
cp "$b/repo/carneades-4/go.sum" "${modfile%.mod}.sum"
go mod edit -replace "github.com/hfried/GoCHR=$b/repo/GoCHR" "$modfile"
(cd "$b/repo/carneades-4" && go vet -mod=mod -modfile="$modfile" ./... && go test -mod=mod -modfile="$modfile" -count=1 -run '^$' ./... > /dev/null)
(cd "$b/repo/docs/scheme-registry/walton-clean/check" && GOFLAGS=-mod=mod go vet ./... && GOFLAGS=-mod=mod go test -count=1 -run '^$' ./... > /dev/null)
test -z "$(git -C "$b/repo" status --porcelain)" || { echo "error: building changed the tree" >&2; git -C "$b/repo" status --porcelain >&2; exit 1; }
cp -R "$GOMODCACHE/cache/download" "$b/offline/goproxy"
chmod -R u+w "$b/offline/goproxy"

echo "-- the upstream walton.yml"
curl -fsSL -o "$b/offline/upstream/walton.yml" "$upstream_url"
echo "$upstream_sha  $b/offline/upstream/walton.yml" | sha256sum -c - > /dev/null

cp "$root/tools/review-offline.sh" "$b/run-offline.sh"
chmod +x "$b/run-offline.sh"
gover="$(sed -n 's/^go \([0-9.]*\)$/\1/p' "$b/repo/wag/go.mod")"
cat > "$b/README.md" <<EOF
# WAG review bundle: commit $commit

Made $(date -u +%Y-%m-%d) by tools/review-bundle.sh. It lets a reviewer with no network run every
check of WAG's CI on this commit, the Go suites of all four modules included.

    bash run-offline.sh

It needs bash, git, Go $gover or newer on PATH (GOTOOLCHAIN=local: no toolchain is downloaded),
and Python 3 with PyYAML. It sets GOPROXY to the file proxy in offline/goproxy, a fresh module
cache, and no checksum database; every module is checked against the commit's go.sum. It prints
PASS or FAIL for each step, writes each step's output to logs/, and exits 0 only if every step
passes. The commit: \`git -C repo rev-parse HEAD\`. The steps are tools/check-all.sh's.

offline/upstream/walton.yml is the schemes library's walton.yml that the cleaning was written for
(SHA-256 $upstream_sha), under the Mozilla Public License 2.0
(repo/docs/scheme-registry/walton-clean/LICENSE), from $upstream_url.

Modules in offline/goproxy: $(find "$b/offline/goproxy" -name '*.zip' | wc -l | tr -d ' ').
EOF

echo "-- the bundle"
rm -f "$out"
(cd "$work" && zip -qr "$out" "$name")
sha256sum "$out"
