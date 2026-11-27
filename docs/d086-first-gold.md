# The first gold (D086): plan, version 1

Status: plan, 2026-11-27, before any drafting; the developer's under the owner's instruction to go with
the recommended choices. Stage 3's first item ([plan.md](plan.md) §4, item 1; D086, D045, D092).

D086 (the owner, 2026-11-24): "The developer and the other independent reviewer (ChatGPT) make it: the
developer first, then back and forth until they converge. The owner verifies it, which is D045's
verification by a person. Until the owner has verified an item, it is development material, not gold."
The plan (§4): "Drawn from batch 2's claims, starting where annotators A and B, the reference or a model
disagree, with some claims where they agree, since two models can share a mistake. … the owner
verifies, and the dashboard's verdicts are the record."

## 1. What is made

The first gold is batch 2's `gold.json` (format 2, [annotation-records.md](annotation-records.md)),
item by item: an item is gold once the owner has verified it, and only while it is the revision
verified (R17). No new file is made. The draft changes the items of a round in `gold.json`, each change
recorded in its `adjudication.recoded` (`date`, `by`, `fields`, `what`, the reason included); the
built-in copy (`wag/data/annotation/kialo-13292-batch2`) stays equal to it, as a test requires.

## 2. The rounds

Round 1 (this plan), 18 claims:

| Why | Claims |
|---|---|
| Annotators A and B disagree on three or more of the core fields (scheme, fit, role, polarity, the target's claim, instance, part and formula, evidence, the question) | 13292.233, .29, .192, .204, .33, .46 |
| Left open for this draft by the identity contract ([annotation-records.md](annotation-records.md), "Open for the first gold's draft"): targets forced onto a primary instance | 13292.12, .141, .182, .198 |
| The same: fragments that bind every variable, where guidelines §5 item 6 leaves unstated those of formulas the fragment does not state | 13292.6, .11, .108, .218 |
| Practical reasoning against positive consequences, the batch's commonest scheme disagreement ([adjudication.md](kialo-batch2/adjudication.md) §2.3) | 13292.147 |
| Claims where A, B and the reference agree on every core field, drawn at random (Python's `random.Random(86).sample`, from the 22 such claims less 13292.108, already above) | 13292.1, .61, .45 |

Later rounds take the other claims where the annotators disagree (round 2) and then the rest
(round 3), each the same way. "A model" in D086: no run of other models on batch 2 is in the
repository; the owner's runs may add one (§5), and a claim where a model disagrees with the draft is
looked at again before it is verified.

## 3. The draft

For each claim of the round, the developer:

1. codes it from the claim, its Kialo context (`items.json`), the guidelines as they stand, the case
   record, the lexicon and `walton.yml`, the book's pages cited by page (the book's text stays out of
   the repository, D043), before reading the reference's coding of it again;
2. compares that coding with the reference's, annotator A's and annotator B's, field by field, and
   settles each difference by a rule of the guidelines or a source, or marks it for the review (§4);
3. writes the result into `gold.json`: the fields changed, and an `adjudication.recoded` entry
   (`by`: "developer (D086 draft 1)") that names each change and its reason, or says the reference's
   coding is kept and why; an item with nothing changed gets an entry too, so that every item of the
   round says it was drafted;
4. keeps `verified` absent: a draft is development material.

The round is checked as the gold always is: `validate.py` (R2-R17), `test_validate.py`, the
agreement figures (`agreement.py`) and the game's replay (`to_wag.py`, `wag-report.json`,
`wag-game.json`) made again, with the README's account of what changed; the file's `history` gets
the round's entry, and its `status` says how many items are verified.

## 4. The review (the other independent reviewer)

A package for the owner to give the other reviewer, in the workspace (not the repository, for its
working notes): for each claim, its text and Kialo context, the guidelines' rules it turns on (the
repository's text), the draft's coding in plain words and as JSON, the reference's, A's and B's
codings, the developer's reasons, and the questions the draft leaves open; with the files it needs
(`gold.json`, `guidelines.md`, `case.json`, `lexicon.json`, `items.json`, `walton.yml`). Book pages are
cited, not quoted.

The reply comes back through the owner. Each of its claims is checked against the repository before
it is accepted (as every outside review is); an accepted correction is a new recoding (draft 2, …),
and a rejected one is answered with the reason. The round has converged when the reviewer and the
developer agree on every item, or the owner is given the items they still disagree on, each with
both readings.

## 5. The owner's verification

The owner verifies in the dashboard (`wag-server`, *Annotate*), as D086 says: a run on the built-in
batch-2 dataset with the round's claims; for each claim, *Show the reviewer and the gold*, and the
verdict *The gold is right* when the drafted item is right (*None is right* or *Unsure* otherwise,
with the reason in the note). A run needs an annotator profile, but no model has to be asked: a
verdict can be given before any annotation, and a model's annotations, if the owner has some made,
are "a model" of §2.

Each verdict records the dataset's input hashes, the gold's among them (`review/human.json`,
`inputs_sha256`). A new script writes the verifications from it, run by the developer on the
repository with the run's `review/human.json`, which the owner sends from the data folder:

`python3 -I docs/kialo-batch2/verify_gold.py HUMAN.json` reads the verdicts and, for each verdict *The gold is right*
(`choice` "gold") whose recorded gold hash is the hash of `gold.json` as it is now (with LF or CRLF
line endings: a Windows checkout has CRLF, and the dataset already compares both), writes the item's
`verified` (`by` "owner", `date` the verdict's, `item_rev` and `text_sha256` computed now) into
`gold.json` and the built-in copy, and lists what it wrote and what it did not, and why: a verdict
given on another version of the gold (the item may have changed since: verify it again), a claim not
in the gold, a verdict of another choice. It writes nothing when any other input the verdicts recorded
differs from the dataset's now (the claim's text, the guidelines or `walton.yml` could differ), and
says which. It changes no other field, writes the file as the recoding scripts and `validate.py
--repin` do (`json.dumps(…, indent=1, ensure_ascii=False)`), and computes `item_rev` with
`validate.py`'s own function. Its test, `test_verify_gold.py`, runs in CI beside `test_validate.py`.

The verifications are committed with the run's `review/human.json` copied into the workspace as the
record (D086: "the dashboard's verdicts are the record"); the file's `status` says how many items are
verified. A verified item that a later change touches is stale (R17) until the owner verifies it
again.

## 6. Cases (outcomes written before the tests)

| Case | Setup | Expected |
|---|---|---|
| V1 | A run whose `review/human.json` has the verdict *gold* on two claims, given on the gold as it is | Both items get `verified` (`by` owner, the verdict's date, the item's revision and its text's hash now); `validate.py` passes (R17); nothing else in the file changes, byte for byte apart from the two fields |
| V2 | The same verdict, given on another version of the gold (its recorded hash differs) | Nothing written for it; it is listed as given on another version |
| V3 | A verdict *none*, *unsure* or an annotator's name | Nothing written; listed |
| V4 | A verdict on a claim not in the gold | Nothing written; listed |
| V5 | The verdicts' recorded hash of `items.json` differs from the dataset's now | Nothing written at all; the command names the input |
| V5b | V1 with the verdicts' gold hash that of the same `gold.json` with CRLF line endings | As V1 |
| V6 | V1, then an item verified is recoded | `validate.py` refuses it as stale (R17) |
| V7 | V1 run twice | The second run writes nothing new and changes no byte |
| V8 | The built-in copy | Equal to `gold.json` after V1 (the existing test) |
| V9 | A `human.json` that is not a list of verdicts, or a verdict without `inputs_sha256` | Refused, naming the file; nothing written |

Each rule of the command gets a planted defect that one case must catch.

## 7. Out of this plan

Batch 2 recoded under the guidelines' version 4 (plan §4, item 2), other discussions and the probe set
(items 3 and 4), the bounded pilot (stage 3) and D087's order; and any change to `walton.yml`, which
goes through a proposal as always.
