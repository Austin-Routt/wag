# D081: source dispositions and property-scoped sorites

## Accepted source dispositions

ChatGPT's review of stages 0 and 1 compared the narrow repairs with Walton,
Reed and Macagno (2008); a second, independent reviewer checked each against
the book's page images (the private workspace's
`verify/r-stage-review-2026-11-22/`, report B). Both are model checks, not a
person's verification, and passing tests do not prove a reading of the source.
D081 is decided under the owner's standing rule of 2026-11-22, after the
proposal [D081-D084](../proposals/D081-D084.md) was checked. The PDF's SHA-256
is `d010348014e8b80fa561b68e5fac732cce0b3905fb0b1d522dd6fd2578a69634`.

| Record | Printed source | Disposition |
| --- | --- | --- |
| `moral_justification_ad_populum_excuse` | Entry 4.6, p. 312 | Accept moving `morally_acceptable_action(X,A)` from premises to the empty conclusion. |
| `sacrifice` | Entry 20, p. 322 | Accept `great_sacrifice(S,A,V)` in place of its undeclared X argument. |
| `rhetorical_oppositions_normative` | Entry 12.1, p. 318 | Accept contrary treatment X2 in the fourth argument. Preserve WAG's treatment-first argument roles. |
| `sorites_slippery_slope` | Entry 51, pp. 340–341 | Revise the earlier binary repair: the chain carries the property P that the base and inductive premises and the conclusion are about. The printed chain premise names no property; scoping it to P is our reading. |

The first three executable repairs were already present. Their disposition and
stale source-link descriptions are corrected here. The only new executable
scheme change is sorites. The recipe's `owner decision` category is right for
all four: the owner's standing rule makes a checked recommendation the owner's
decision, and the reason fields name the review it came from, not a reading of
the book by the owner. These decisions approve the stated repairs only.

## Sorites change

Replace `reapplication_sequence(A2,A)` with
`reapplication_sequence(A2,A,P)` and language key `/2` with `/3`. Its template,
in our own words, keeps the book's sense: a chain of modus ponens steps from the
clear case through the grey area, each step concerning the property. The other
premises remain `beyond_contention(A2,P)` and `all_has(A2,P,A1)`; the conclusion
remains `may_have(A,P)`.

A chain about wealth cannot establish a stepwise classification concerning
baldness merely because its endpoint names coincide. The binary representation
could not express that distinction. Native CHR fixtures use independent facts
for the initial case, local inductive step and chain. They require a matching
property and both endpoints, and reject a different property or unrelated
endpoint. A comparative run against the old library encodes its chain using the
old two-argument representation; that run overgenerates the different-property
case. This documents representational information loss rather than claiming
that the old engine was given a property argument that its schema lacked.

Run from `wag/`:

```sh
go test ./internal/carneades -run '^TestD081NativePropertyScope$' -count=1 -v
```

Set `WAG_MIGRATION_LIBRARY` to an absolute pre-D081 cleaned file to reproduce the
old different-property failure with the same test. Leave it unset for normal CI.

`python3 -B tools/check_d081_migration.py` checks the four scoped dispositions,
linked formulas, generated catalogue and identical annotation copy. Add
`--before /path/to/pre-D081-walton.yml` at this transition to verify the entire
parsed library changes only the sorites premise and its key/template. A later
independent semantic migration is expected to invalidate that strict comparison.

## Data migration

Keep old datasets immutable. New runs use the changed collection hash. A binary
chain atom alone does not identify its property, so an automatic migration must
not guess P from nearby facts, choose an arbitrary property, or add a global
default. A derived migration needs recorded source-argument provenance that
uniquely establishes the intended property, followed by an audited binding and
updated statement references. If that provenance is absent or ambiguous, retain
the original and mark the new representation as requiring reannotation.

This remains an explicit chain-evidence premise. It does not establish a full
proof of each link, identify an exact vagueness threshold, or certify an entire
slippery-slope taxonomy. Those are distinct dataset and validation tasks.
