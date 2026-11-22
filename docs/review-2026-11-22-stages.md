# ChatGPT's review of stages 0 and 1 (22 November 2026): what was adopted

ChatGPT reviewed the engine swap, the decisions and the plan of 22 November
(commits `69472ae` to `dbfc6be`) and supplied eight patches against
`dbfc6be`, with decisions on every open entry of the register. The archive
is kept whole in the private workspace (`reviews/2026-11-22/5-*`). Five
independent reviewers checked it, one area each, one of them against the
book's page images (the workspace's `verify/r-stage-review-2026-11-22/`,
reports A to E, and `SUMMARY.md`).

The owner's instructions for this round (22 November): its decisions are
recommendations under the standing rule ([process.md](process.md) §5), not
decisions of their own; none overrides the owner's earlier instructions; the
public history is kept (D078); the code's MPL-2.0 and the documents' CC BY
4.0 apply now (D044, D080).

## What was checked

- The archive matches its manifest, and the eight patches give exactly the
  tree it names (`fcf649a`). On Go 1.24.13 and 1.27.4, vet, the 293
  top-level tests (400 with subtests), the engine's tests, lint, the
  catalogue's generator and check, the registry gate, the batch-2 gold and
  the walton.yml rebuild all pass, and the 33 reference games replay.
- Its findings against the work of 22 November held, in substance:
  - the registry step in CI passed when the tool crashed;
  - the rebuild only warned on a wrong upstream file;
  - core attach and the dashboard did not check a continued annotation
    run's profile, and the schema's bytes were not compared;
  - two annotator names could share one result file, and damaged results
    could be overwritten;
  - a malformed final move still let an earlier example in the reply be
    played;
  - D041's note of a later printing had no basis in the book.

## Adopted

| Patch | What | Changes made on adoption |
|---|---|---|
| 1 | CI on both Go versions; a registry gate that fails closed; the rebuild refuses a wrong upstream file | Tests for the import gate and for an unwritable report; usage errors on stderr; one shell step split so that `set -e` sees it |
| 2, 3 | D084 (`is` becomes `is_a`) and D081 (the sorites chain scoped to its property; three repairs approved) | First the proposal [D081-D084](proposals/D081-D084.md), checked with no errors or warnings, with the gate and the registry run on it; the chain template in our own words; a test of the three repairs' roles; the D081 checker also refuses a replaced form; the D084 test pins the old silent failure |
| 4 | Annotation runs keep their identity at every entry point; a malformed final move is refused | Brackets in prose are not JSON (the patch refused "[s3]" after a valid move); wrong-typed targets and bindings are refused by name; damaged runs are not cached by the dashboard; a few smaller fixes |
| 5 | A free-text bridge no longer stands in for an answer's conclusion; arguments record their game scheme | A clear error for a saved game a bridge let through (the patch said such games still load; they do not); lookup by the recorded ID first; leftovers removed |
| 8 | C001 and C002 take D056's outcome | `by` is the owner's standing rule, as D056 records |

In the register, adopted under the standing rule:
- D081 decided;
- D084 applied;
- refinements appended to D007, D008, D009, D011, D050, D051, D063, D069,
  D075 and D076. The weightiest is positive identity in bridges (D063,
  D011, D075): an exception nobody raises is presumed not to hold, so an
  exception alone would let an unknown agent through.

The models' readings of D021, D027 and D041 are added to their questions.
The review's constraints on the nine candidates of D083 are recorded as
their needs.

## Not adopted

- **Patch 6 as delivered.** It re-signed 28 entries as decided by ChatGPT
  "under user-delegated decision authority" and replaced, rather than
  amended, the text of 20 decisions the owner had already taken. Among
  them it:
  - removed D077's fair-use basis;
  - dropped D046's "on the owner's machine";
  - deleted D043's "no book text in the public repository";
  - reversed D045 (a person verifies);
  - decided D079 against the owner's note that its questions are reviewed
    later, making capture wait on controls;
  - closed the verifier entries D021, D027 and D041 without a person;
  - rewrote process §5 to grant a reviewer that authority;
  - said that evidence ingestion controls were already in place, when none
    exist;
  - decided candidates that are not "ready".

  Its sound parts are taken as listed above.
- **D036's adequacy policy.** The book carries a causal link's strength in a
  question, which D036 follows. The case the review worried about, a weak
  link offered for a categorical one, is D008's test.
- **Patch 7 as delivered.** The licences are added (D044, D080), but
  `evidence-policy.md` made capture wait on D079's controls and reread the
  owner's fair-use basis. Its record model is kept as a design note for
  stage 2 ([evidence-records.md](evidence-records.md)).

## Still open

| Entry | Kind | What it waits for |
|---|---|---|
| D021, D027, D041, D085 | verifier | A person confirms the models' readings. D085 is new: the sorites source's page, 151 in walton.yml, 251 in the book. |
| D079 | owner | Reserved by the owner for later review. |
| D083 | owner | Its candidates' sketches. |

Also open: a test of an undercutter played with a typed scheme, which needs
a scheme whose own conclusion answers an undercutter question. And
upstreaming WAG's three engine patches to the schemes library.
