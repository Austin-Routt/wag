#!/bin/sh
# Replace carneades-4/ and GoCHR/ with the engine of the schemes library
# (Austin-Routt/walton-argumentation-schemes-for-carneades-4, third_party/) at
# a commit, then apply WAG's local engine patches (tools/carneades-patches/,
# in name order) and keep WAG's fork regressions (carneades-4/src/test/
# fork_regression_test.go).
#   tools/sync_carneades_fork.sh /path/to/walton-library 024b86b
# A patch the library has taken no longer applies: delete it here.
set -eu
lib="$1"; commit="$2"
root="$(cd "$(dirname "$0")/.." && pwd)"
git -C "$lib" rev-parse --verify "$commit^{commit}" >/dev/null
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
git -C "$lib" archive "$commit" third_party | tar -x -C "$tmp"
for p in "$root"/tools/carneades-patches/*.patch; do
  [ -e "$p" ] || continue
  (cd "$tmp" && patch --batch -p1 --forward < "$p") || { echo "patch $(basename "$p") does not apply: has the library taken it?" >&2; exit 1; }
done
keep="$root/carneades-4/src/test/fork_regression_test.go"
cp "$keep" "$tmp/fork_regression_test.go"
rm -rf "$root/carneades-4" "$root/GoCHR"
cp -R "$tmp/third_party/carneades-4" "$root/carneades-4"
cp -R "$tmp/third_party/GoCHR" "$root/GoCHR"
cp "$tmp/fork_regression_test.go" "$keep"
echo "carneades-4/ and GoCHR/ now hold library commit $(git -C "$lib" rev-parse --short "$commit") with the local patches;"
echo "update CARNEADES_FORK.md, then run tools/engine-tests.sh and go test ./... in wag/"
