#!/bin/sh
# Replace carneades-4/ with a commit of the shared Carneades-4 fork.
#   tools/sync_carneades_fork.sh /path/to/carneades-fork 995768e
set -eu
fork="$1"; commit="$2"
root="$(cd "$(dirname "$0")/.." && pwd)"
git -C "$fork" rev-parse --verify "$commit^{commit}" >/dev/null
rm -rf "$root/carneades-4"
mkdir "$root/carneades-4"
git -C "$fork" archive "$commit" | tar -x -C "$root/carneades-4"
echo "carneades-4/ now holds fork commit $(git -C "$fork" rev-parse --short "$commit"); update CARNEADES_FORK.md and run go test ./... in wag/"
