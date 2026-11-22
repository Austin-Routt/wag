# The Carneades-4 engine in `carneades-4/` and `GoCHR/`

Since 2026-11-22 WAG uses the engine of the schemes library,
[`Austin-Routt/walton-argumentation-schemes-for-carneades-4`](https://github.com/Austin-Routt/walton-argumentation-schemes-for-carneades-4),
which keeps Carneades-4 and the GoCHR rule engine together in its `third_party/`.
`wag/go.mod` reaches both through `replace` lines:

```
replace github.com/carneades/carneades-4 => ../carneades-4
replace github.com/hfried/GoCHR => ../GoCHR
```

The two must move together: the library's Carneades does not compile against upstream GoCHR.

| | |
|---|---|
| Library commit copied here | `024b86b` (library 0.4.1, engine 4.3.1+walton.8) |
| Upstream base of both | Carneades-4 `d19d543` (2024-10-15); GoCHR as the library's `sources.lock` records |
| Local patches | `tools/carneades-patches/`, applied in name order |

**The local patches.** The library lacks three fixes. Each patch is also meant for the library itself; delete it here once the library has taken it.

1. `00-empty-cumulative.patch`: a cumulative argument with no premises weighs 1.0, as in the linked weighing function, not 0/0 (NaN). From WAG's old fork (`8d1728d`).
2. `01-yaml-scalars.patch`: the YAML writer quotes every string with Go's ASCII quoting, so that text with backslashes (Kialo's `\(` maths), line breaks, control characters, DEL, C1 controls, noncharacters, `yes` or `123` comes back unchanged. Every valid Unicode character is tested. It replaces the JSON escaping of the old fork (`0005b19`), which failed on DEL, C1 controls and U+FFFE/U+FFFF and turned NEL into a space. Written by the external review of 2026-11-22.
3. `02-yaml-invalid-utf8.patch`: invalid UTF-8 becomes U+FFFD before quoting, instead of `\xNN`, which YAML would read back as a different valid character.

**WAG's fork regressions.** `carneades-4/src/test/fork_regression_test.go` (five tests from the old fork) is kept in the copy and runs on the library's engine.

**What changed for WAG.** The library's graph keeps two records the old fork did not: which assumptions a scheme added by default (`Defaults`) and which statements are questioned (`Questioned`). WAG's checkpoint, retraction and moderator code now keeps them consistent (`wag/internal/carneades`; a test fails if the engine's graph gains a field the checkpoint does not know). Thirty-three seeded reference games (`wag/internal/game/testdata/reference-digests.txt`) and 120 more seeded games gave the same move digests on the old fork and on the library's engine. The scripts that call the `carneades` command now see malformed input refused and exit statuses that tell invalid input, an inconsistent context, a limit and a cancellation apart.

**Checks.** `tools/engine-tests.sh` runs GoCHR's and Carneades' own tests (with the five fork regressions) against the local GoCHR; CI runs it. WAG's own suite passes even without the engine patches, so this is the check that catches a lost one.

**Updating.** Run `tools/sync_carneades_fork.sh <library checkout> <commit>`, then `tools/engine-tests.sh` and `go test ./...` in `wag/`, and record the commit above. The old fork (branch `kialo-research`, commit `995768e`) stays in use by the Kialo argumentation project until it reruns its grid of 342 labellings on this engine.

**Licence:** Carneades-4 and GoCHR keep their own licences (`carneades-4/LICENSE`, `GoCHR/LICENSE`); Carneades is MPL-2.0, and the patched files are published under it.
