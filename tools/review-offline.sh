#!/usr/bin/env bash
# Run every check of WAG's CI on the review bundle this file is in, with no
# network: the Go modules from offline/goproxy, the upstream walton.yml from
# offline/upstream. Made by tools/review-bundle.sh (as run-offline.sh).
#
#   bash run-offline.sh
set -uo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
[ -d "$here/repo/.git" ] && [ -d "$here/offline/goproxy" ] || { echo "error: run this in a review bundle (repo/, offline/)" >&2; exit 2; }
cache="$(mktemp -d)"
trap 'chmod -R u+w "$cache" 2>/dev/null; rm -rf "$cache"' EXIT
export GOPROXY="file://$here/offline/goproxy" GOSUMDB=off GONOSUMDB= GOPRIVATE= GOFLAGS= \
	GOMODCACHE="$cache" GOTOOLCHAIN=local GOWORK=off
bash "$here/repo/tools/check-all.sh" "$here/offline/upstream/walton.yml" "$here/logs"
