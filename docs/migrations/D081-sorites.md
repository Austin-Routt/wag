# D081: source dispositions and property-scoped sorites

## Accepted source dispositions

An independent AI reviewer compared the narrow repairs against the supplied
Walton, Reed and Macagno (2008) source, under the user's delegation to resolve
remaining decisions. This is AI source review, not human verification or a
claim that test success proves source interpretation. The supplied PDF's SHA-256
is `d010348014e8b80fa561b68e5fac732cce0b3905fb0b1d522dd6fd2578a69634`.

| Record | Printed source | Disposition |
| --- | --- | --- |
| `moral_justification_ad_populum_excuse` | Entry 4.6, p. 312 | Accept moving `morally_acceptable_action(X,A)` from premises to the empty conclusion. |
| `sacrifice` | Entry 20, p. 322 | Accept `great_sacrifice(S,A,V)` in place of its undeclared X argument. |
| `rhetorical_oppositions_normative` | Entry 12.1, p. 318 | Accept contrary treatment X2 in the fourth argument. Preserve WAG's treatment-first argument roles. |
| `sorites_slippery_slope` | Entry 51, pp. 340–341 | Revise the earlier binary repair: the chain must carry the same property P as the initial and inductive premises and conclusion. |

The first three executable repairs were already present. Their disposition and
stale source-link descriptions are corrected here. The only new executable
scheme change in this patch is sorites. The recipe's existing `owner decision`
category denotes authorized decisions; the reason fields explicitly retain the
delegated, independent AI provenance rather than claiming personal source
inspection by the owner. These decisions approve the stated repairs only.

## Sorites change

Replace `reapplication_sequence(A2,A)` with
`reapplication_sequence(A2,A,P)` and language key `/2` with `/3`. Its new wording
names both endpoints and the property. The other premises remain
`beyond_contention(A2,P)` and `all_has(A2,P,A1)`; the conclusion remains
`may_have(A,P)`.

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
