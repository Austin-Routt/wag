#!/bin/sh
# Rebuild walton.yml (the cleaned collection) from the original, and re-run
# every check, writing the outputs that CHANGELOG.md cites.
#
#   sh check/run.sh ORIGINAL
#
# ORIGINAL is walton.yml of
# github.com/basakkurtuldu/walton-argumentation-schemes-for-carneades-4
# (MPL-2.0; the edits were made for SHA-256
# 1976202ac6bd538526ea75488659ec9880641ec03534d6e098f6f6dd1bcc099c).
# It needs Go (the Carneades-4 fork at the repository's root) and Python 3
# with PyYAML. Each walton-check run is its own process (the Carneades
# importer mutates package-level state).
set -eu
IN=$(cd "$(dirname "$1")" && pwd)/$(basename "$1")
HERE=$(cd "$(dirname "$0")" && pwd)   # walton-clean/check
OUT=$(dirname "$HERE")                # walton-clean
REPO=$(cd "$OUT/../../.." && pwd)
WANT=1976202ac6bd538526ea75488659ec9880641ec03534d6e098f6f6dd1bcc099c
GOT=$(sha256sum "$IN" | cut -d' ' -f1)
[ "$GOT" = "$WANT" ] || { echo "error: $IN is not the original the edits were written for (SHA-256 $GOT)" >&2; exit 2; }
cd "$HERE" || exit 2
GOFLAGS=-mod=mod go test -count=1 ./... || exit 2
GOFLAGS=-mod=mod go build -o walton-check . || exit 2

# A finding is exit 1; failed execution or output is an operational error.
# Expected findings must not hide a broken build, missing input or failed write.
check_one() {
  label="$1"; output="$2"; shift 2
  status=0
  "$@" > "$output" || status=$?
  if [ "$status" -gt 1 ]; then
    echo "$label: checker could not complete (exit $status)" >&2
    exit "$status"
  fi
  echo "$label: exit $status"
}

python3 -I "$OUT/tools/apply_edits.py" "$IN" "$HERE/walton.clean.yml" --json "$OUT/changes.json"
cat "$OUT/tools/header.txt" "$HERE/walton.clean.yml" > "$OUT/walton.yml"
python3 -I "$OUT/tools/apply_edits.py" "$IN" "$HERE/walton.syntax-only.yml" --syntax-only
python3 -I "$OUT/tools/apply_edits.py" "$IN" "$OUT/walton.proposals.yml" --with-proposals --json "$OUT/tools/proposals-changes.json"

cd "$OUT" || exit 2
check_one original check/original.txt check/walton-check -json check/original.json "$IN"
# The report names the file it read: not a path on this machine.
{ echo "walton-check walton.yml (the original)"; tail -n +2 check/original.txt; } > check/original.tmp
mv check/original.tmp check/original.txt
check_one syntax-only check/syntax-only.txt check/walton-check -json check/syntax-only.json -dump tools/dump-syntax-only.json check/walton.syntax-only.yml
check_one clean clean-check.txt check/walton-check -json check/clean.json -dump tools/dump-clean.json walton.yml
check_one proposals proposals-check.txt check/walton-check -json check/proposals.json walton.proposals.yml
python3 -I tools/semantic_diff.py tools/dump-syntax-only.json tools/dump-clean.json > check/semantic-diff.txt
DIFF=$(mktemp)
status=0
diff -u "$IN" check/walton.clean.yml > "$DIFF" || status=$?
[ "$status" -le 1 ] || { rm -f "$DIFF"; exit "$status"; }
sed "1s|.*|--- walton.yml (original)|; 2s|.*|+++ walton.yml (cleaned)|" "$DIFF" > walton.diff
rm -f "$DIFF"
# WAG's own lint: wag-cli --lint --data DIR reads DIR/schemes/core.yml and
# DIR/nascq/*.json; only its "language" findings are about the scheme file.
WAG="$REPO/wag"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
(cd "$WAG" && go build -o "$TMP/wag-cli" ./cmd/wag-cli)
for v in syntax-only clean; do
  mkdir -p "$TMP/$v/schemes" "$TMP/$v/nascq"; cp "$WAG"/data/nascq/*.json "$TMP/$v/nascq/"
done
cp check/walton.syntax-only.yml "$TMP/syntax-only/schemes/core.yml"
cp check/walton.clean.yml "$TMP/clean/schemes/core.yml"
for v in syntax-only clean; do
  status=0
  "$TMP/wag-cli" --lint --data "$TMP/$v" > "check/wag-lint-$v.txt" 2>&1 || status=$?
  [ "$status" -le 1 ] || { rm -rf "$TMP"; exit "$status"; }
  echo "wag lint $v: exit $status"
done
rm -rf "$TMP"
python3 -I tools/make_changelog.py "$OUT"
