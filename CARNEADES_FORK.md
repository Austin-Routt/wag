# The Carneades-4 copy in `carneades-4/`

`carneades-4/` is not upstream Carneades. It is the shared research fork used by both this project (WAG) and the Kialo argumentation project. `wag/go.mod` reaches it through `replace github.com/carneades/carneades-4 => ../carneades-4`.

| | |
|---|---|
| Upstream | `carneades/carneades-4` at `d19d543` (2024-10-15) |
| Fork branch | `kialo-research` |
| Fork commit copied here | `995768e` |

The fork's commits on top of upstream:

1. `8d1728d` — Validator: no more crashes or silent skips; `Inconsistent()` fixed; cumulative weighing no longer returns NaN. Five regression tests, each failing on upstream.
2. `404d7d2` — `go test` passes and means something.
3. `0005b19` — YAML strings are escaped on export; `check` and `eval` exit 1 on problems.
4. `995768e` — The GopherJS command takes WAG's compile fix and stays behind the `js` build constraint.

None of them changes a labelling:
* the Kialo project's grid of 342 labellings is identical before and after;
* 40 seeded WAG games give the same move digests with the old bundled copy and with this one.

WAG's previous copy was upstream, with CRLF line endings and two local edits (the GopherJS fix, now commit 4, and the removed dead package, now in commit 2).

**Updating:** run `tools/sync_carneades_fork.sh <path to a fork checkout> <commit>`, then run `go test ./...` in `wag/`, and record the new commit in the table above.

**Licence:** Carneades is MPL-2.0. The fork's modified files are published under MPL-2.0 with the fork.
