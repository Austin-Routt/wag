# D084: ordinary classification predicate

## Decision and scope

Use `is_a/2` for WAG's ordinary two-place classification predicate. `is/2`
collides with GoCHR's assignment operation. This migration changes exactly the
language key and two formulas in the cleaned collection:

| Record | Role | Previous | New |
| --- | --- | --- | --- |
| `group_2` | conclusion | `is(M,Q)` | `is_a(M,Q)` |
| `rule_physical_world_1` | premise | `is(W,D)` | `is_a(W,D)` |

The template remains `%s is %s.`; argument order, all other formulas, weighting,
and critical questions remain the same. The original upstream source file is
unchanged. The cleaning recipe performs three targeted substitutions, then the
linked formula maps, generated catalogue and annotation copy are synchronized.

## Verification

`python3 tools/check_d084_migration.py` checks the executable collection, linked
formulas, catalogue and identical annotation copy. For the D084-only transition,
add `--before /path/to/pre-D084-walton.yml`: it compares the entire parsed YAML
against exactly the two formula replacements and one key rename. Do this before
applying another semantic migration; a later independent change must make that
strict comparison fail.

`go test ./internal/carneades -run '^TestD084Native' -count=1` from `wag/`
runs both D084 tests: native forward inference on positive and
mismatched/missing-premise cases, and the two-scheme bridge in which group_2's
conclusion feeds rule_physical_world_1. It loads the committed library and isolates the named scheme. These tests
do not substitute WAG's explicit `InstantiateScheme` path for the CHR path that
caused the compatibility concern.

## Existing datasets and games

Keep existing runs and source captures immutable. Start new runs with the new
catalogue hash. Do not resume an old run as if it used this catalogue. A derived
migration of an existing graph must record its old and new hashes and update all
references to the renamed statement IDs, including premises, assumptions,
issues and argument conclusions. Reject ambiguous collisions with an existing
`is_a` statement unless its identity and provenance are reconciled explicitly.
Do not globally replace the English word `is`, quoted evidence, or arbitrary
string substrings. This patch does not supply a blind historical graph rewriter.

The batch-2 gold keeps the old name: `docs/kialo-batch2/annotator-A.json` has
seven `is(...)` formulas and describes is/2 as a walton.yml predicate. They are
the historical record of what that annotator wrote against the catalogue of
the time, and stay as they are; `validate.py` passes. A new run uses `is_a`.

The rename removes a runtime-name collision; it does not certify the full source
interpretation of entry 54.1. In particular, foundation's additional
`referentially_coextensive` condition would be a distinct semantic change with
its own source review and fixtures. It is not silently imported by D084.

The cleaned collection still contains four declared variable-binding problems
in other records. A passing D084 check is not a claim that all 116 records pass
the automatic-inference validator.
