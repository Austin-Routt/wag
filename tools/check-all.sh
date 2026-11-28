#!/usr/bin/env bash
# Every check of CI (.github/workflows/ci.yml), run on this working tree with
# the go and python3 on PATH. Keep the two in step: a step added to CI is added
# here. The review bundle's run-offline.sh runs this (tools/review-bundle.sh).
#
#   bash tools/check-all.sh UPSTREAM_WALTON_YML [LOGDIR]
#
# UPSTREAM_WALTON_YML is walton.yml of the schemes library the cleaning was
# written for (SHA-256 1976202ac6bd538526ea75488659ec9880641ec03534d6e098f6f6dd1bcc099c;
# CI fetches it, the review bundle carries it). Each step's output is in
# LOGDIR (a new temporary directory if none is given). The exit status is 0
# only if every step passes; the last step also requires the tree unchanged.
set -uo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
[ $# -ge 1 ] || { echo "usage: bash tools/check-all.sh UPSTREAM_WALTON_YML [LOGDIR]" >&2; exit 2; }
[ -f "$1" ] || { echo "error: no file $1" >&2; exit 2; }
upstream="$(cd "$(dirname "$1")" && pwd)/$(basename "$1")"
logs="${2:-$(mktemp -d)}"
mkdir -p "$logs" && logs="$(cd "$logs" && pwd)"
fail=0
step() {
	local name=$1
	shift
	if "$@" > "$logs/$name.log" 2>&1; then
		echo "PASS $name"
	else
		echo "FAIL $name ($logs/$name.log)"
		fail=1
	fi
}
echo "== $(go version); $(python3 --version 2>&1); commit $(git -C "$root" rev-parse HEAD 2>/dev/null || echo unknown)"

# The Go job.
cd "$root/wag" || exit 2
step vet go vet ./...
step test go test -count=1 ./...
step evidence go run ./cmd/wag-evidence ../examples/evidence/bundle.json
step engine bash -c 'cd .. && bash tools/engine-tests.sh'
step lint go run ./cmd/wag-cli --lint
step catalogue-gen bash -c 'go run ./cmd/wag-cli --catalogue-gen && git diff --exit-code -- data/catalogue data/nascq data/schemes/core.yml ../docs/decisions.md ../docs/candidates.md'
step catalogue-check go run ./cmd/wag-cli --catalogue-check
step registry bash -c 'python3 ../tools/test_registry_gate.py && bash ../tools/check-registry.sh'

# The Python job.
cd "$root" || exit 2
step validate python3 -I docs/kialo-batch2/validate.py
step test-validate python3 -I docs/kialo-batch2/test_validate.py
step test-agreement python3 -I docs/kialo-batch2/test_agreement.py
step test-verify-gold python3 -I docs/kialo-batch2/test_verify_gold.py
step test-to-wag-inference python3 -I docs/kialo-batch2/test_to_wag_inference.py
step test-to-wag-other python3 -I docs/kialo-batch2/test_to_wag_other_reporting.py
step test-to-wag-unresolved python3 -I docs/kialo-batch2/test_to_wag_unresolved.py
step test-to-wag-discussion python3 -I docs/kialo-batch2/test_to_wag_discussion.py
step rebuild bash -c 'sh docs/scheme-registry/walton-clean/check/run.sh "$1" && python3 -B tools/check_d084_migration.py && python3 -B tools/check_d081_migration.py && { test -z "$(git status --porcelain)" || { echo "the tree changed:"; git status --porcelain; exit 1; }; }' rebuild "$upstream"

echo "ALL $([ $fail = 0 ] && echo PASS || echo FAIL) (logs in $logs)"
exit $fail
