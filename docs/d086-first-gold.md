# The first gold (D086): plan, version 2

Status: plan, 2026-11-27, written before any drafting; the developer's under the owner's instruction to
go with the recommended choices. Round 1's first draft followed the same day (§10), and its second, from the first exchange of the review (§11), corrected after an independent check (§12). Phase 3's first item ([plan.md](plan.md) §4, item 1; D086, D045, D092).
Version 1 (8fcd001) was checked independently the same day; the check found four errors, seven
warnings and ten notes, each verified against the repository and taken up here (§8).

D086 (the owner, 2026-11-24): "The developer and the other independent reviewer (ChatGPT) make it: the
developer first, then back and forth until they converge. The owner verifies it, which is D045's
verification by a person. Until the owner has verified an item, it is development material, not gold."
The plan (§4): "Drawn from batch 2's claims, starting where annotators A and B, the reference or a model
disagree, with some claims where they agree, since two models can share a mistake. … the owner
verifies, and the dashboard's verdicts are the record."

## 1. What is made

The first gold is batch 2's `gold.json` (format 2, [annotation-records.md](annotation-records.md)),
item by item: an item is gold once the owner has verified it, and only while it is the revision
verified (R17). No new format is made. The draft changes the items of a round in `gold.json`, each
change recorded in its `adjudication.recoded` (`date`, `by`, `fields`, `what`, the reason included);
the built-in copy (`wag/data/annotation/kialo-13292-batch2`) stays equal to it, as a test requires.

One new file records where each claim stands: `docs/kialo-batch2/d086-rounds.json`, the rounds, their
claims, and for each claim its latest draft, the review's state (§4) and the revision of the item the
developer and the reviewer agreed on, or the owner chose. The verifications are written only for an
item agreed on at the revision it has (§5).

## 2. The rounds

"Where they disagree" is measured on the ten core fields (the scheme, its fit, the role, the polarity,
the target's claim, instance, part and formula, the evidence, the question) among three codings:
annotator A's, annotator B's and the reference (`gold.json` before the draft), by `agreement.py`
(A with B, A with the reference, B with the reference; each read in its own format and with the case
record, unresolved values left out). A field is in disagreement for a claim when any two of the three
differ on it.

Round 1, 28 claims:

| Why | Claims |
|---|---|
| The three differ on three or more fields (the number in brackets) | 13292.127 (8), .12 (6), .233 (5), .104 (5), .43 (5), .29 (4), .192 (4), .147 (4), .182 (4), .198 (4), .202 (4), .25 (4), .174 (4), .123 (4), .204 (3), .33 (3), .46 (3), .194 (3), .141 (3), .8 (3), .5 (3) |
| Left open for this draft by the identity contract ([annotation-records.md](annotation-records.md), "Open for the first gold's draft") and not above: fragments that bind every variable, where guidelines §5 item 6 leaves unstated those of formulas the fragment does not state | 13292.6, .11, .108, .218 |
| Controls: where the three agree on every field, drawn at random | 13292.1, .30, .36 |

The contract's other open items, the four targets that format 1 forced onto a claim's primary
instance (13292.12, .141, .182, .198), are in the first row. The controls are
`random.Random(86).sample(population, 3)` (Python 3), the population the ten claims where the three
agree on every field, less 13292.108 (in the second row), in `items.json`'s order: 13292.1, .7, .115,
.118, .180, .36, .137, .3, .30, .56.

The claims where the reference differs from both annotators (13292.127, .43, .174, .123, .104, .25,
.5 among them) are mostly the developer's own recodings since the batch was adjudicated (D055, the
practical-reasoning and consequences proposals, step C5.2): no one else has checked them, which is
why the three-way measure takes them in.

Round 2 takes the claims where two of the three differ on one or two fields (13292.135, .2, .4, .224,
.145, .14, .169, .208, .190, .235, .13, .9, .61, .121, .45) and round 3 the rest (13292.7, .115,
.118, .180, .137, .3, .56), each the same way.

"A model" in D086: no run of another model on batch 2 is in the repository. The first pilot
(`docs/kialo-pilot/pilot.json`, 20 of the claims) is not used to choose: it was made before the
guidelines, under other conventions (the game's schemes of the time, targets without formulas). A run
the owner makes (§5) with a model is "a model": a claim where it disagrees with the draft on a core
field is looked at again before it is verified.

## 3. The draft

For each claim of the round, the developer:

1. codes it from the claim, its Kialo context (`items.json`), the guidelines as they stand, the case
   record, the lexicon and `walton.yml`, the book's pages cited by page (the book's text stays out of
   the repository, D043), before reading the reference's coding of it again, and records that reading
   in the workspace before step 2;
2. compares that coding with the reference's, annotator A's and annotator B's, field by field, and
   settles each difference by a rule of the guidelines or a source, or marks it for the review (§4);
3. writes the result into `gold.json`: the fields changed, and an `adjudication.recoded` entry
   (`by`: "developer (D086 draft 1)") that names each change and its reason, or says the reference's
   coding is kept and why; an item with nothing changed gets an entry too, so that every item of the
   round says it was drafted;
4. keeps `verified` absent: a draft is development material.

The developer's reading in step 1 is not independent of the reference, and the plan does not claim
it is: the developer made most of the reference's recodings, the case record and the lexicon are drawn
from the gold (annotation-records.md §12), the guidelines' examples give some of its codings (13292.61's
analogy, §3), and the rule cases code 13292.29, .204 and .233. The independence D086 asks for is the
other reviewer's and the owner's.

A draft that changes an item outside the round, or would leave one broken (a target whose formula is
gone, R13; a challenge record that names a formula no longer there, R10), changes that item in the
same draft, and the item joins the round: it is reviewed and verified with it. Where it can, a draft
keeps the formulas that other items target.

The round is checked as the gold always is: `validate.py` (R2-R19) and `test_validate.py`; the Go
tests of the annotation package (`go test ./internal/annotate/`), two of which pin the gold
(`TestRuleCases` compares the rule cases' codings with it, `TestAgreementFieldsAcrossFormats` pins the
agreement of the format-1 and format-2 golds and two targets), updated with each change and its reason;
the rule cases (`rule-cases.yml`) changed where the gold's coding of their claims changes; and the
game's replay (`to_wag.py`, `wag-report.json`, `wag-game.json`) made again, with the README's account
of what changed. The annotators' agreement (`agreement.json`, A with B) does not depend on the gold and
does not change. The file's `history` gets the round's entry, and its `status` says how many items are
verified.

The replay plays each claim's primary instance and a secondary one that carries it, nothing else: a
claim whose target is another secondary instance (a causal step, say) is not played (a "cascade").
The gold says what a claim answers; the replay, what the game can play of it. A target moved onto a
causal step makes its claim a cascade, which the README's account names as a gap of the replay.

## 4. The review (the other independent reviewer)

A package for the owner to give the other reviewer, in the workspace (not the repository, for its
working notes): for each claim, its text and Kialo context, the guidelines' rules it turns on (the
repository's text), the draft's coding in plain words and as JSON, the reference's, A's and B's
codings, the developer's reasons, and the questions the draft leaves open; with the files it needs
(`gold.json`, `guidelines.md`, `case.json`, `lexicon.json`, `items.json`, `walton.yml`,
`docs/scheme-registry/links.json` and the game's `nascq/*.json`, which R2 and R16 read). Book pages are
cited, not quoted; `walton.yml` goes as the repository has it (its templates that still repeat the book
are D043's open item, and public already).

The reply comes back through the owner. Each of its claims is checked against the repository before
it is accepted (as every outside review is); an accepted correction is a new recoding (draft 2, …),
and a rejected one is answered with the reason.

Convergence is per item. `d086-rounds.json` records, for each claim, the review's state: not yet
reviewed; objections (each open one named); agreed, with the revision of the item agreed on (its
`item_rev`); or to the owner. A round has at most three exchanges (draft 1, its review, draft 2, its
review, draft 3, its review). An item still disputed after the third goes to the owner with both
readings; the owner's choice is written as the next draft and recorded as agreed by the owner's choice,
and is then verified like any other item. An item agreed on and changed later is not agreed any more
until it is reviewed again.

The third review closes the round for each of its items (made precise on 2026-11-28, before that
review came back). An item the reviewer agrees on is agreed. An item the reviewer objects to is
agreed by the reviewer and the developer when the developer adopts the reviewer's correction as
given: the fields the reviewer states, as stated, and nothing else but what the guidelines' rules
write from them (a formula, a binding), each such detail named in the exchange's record, where the
owner sees it when verifying. A correction the developer rejects, or adopts in part, or one whose
recoding needs a choice the reviewer did not make, leaves the item disputed: it goes to the owner
with both readings. So after the third review no item of the round waits for another review.

## 5. The owner's verification

The owner verifies in the dashboard (`wag-server`, *Annotate*), as D086 and the plan say.

- **The run.** The built-in dataset *Kialo batch 2*, with the round's claims listed in the new
  *Claims* field (the claim ids, separated by commas or spaces; empty, the first N as before), and an
  annotator: the Models tab's *Test service (no network)* needs no key and asks no model. A run with a
  model is "a model" of §2.
- **What is shown.** For each claim, *Show the reviewer and the gold* shows the gold's card and,
  under it, the whole item: every instance with its premises, their English and status, the role, the
  target, the evidence and its links, the challenges, the notes and the adjudication's reasons. The
  verdict *The gold is right* verifies the whole item as shown, every field its revision covers; *None
  is right* or *Unsure* otherwise, with the reason in the note. The package's plain account of each item
  is the reading aid. Within a round, the owner verifies an item's target before the item, where both
  are in the round.
- **A stale page.** A verdict is sent with the input hashes of the page it was given on; the server
  refuses a verdict whose page shows other data than the run has now (the server restarted on another
  build, say), and the dashboard says to reload the claim. A verdict sent without them (by another
  client of the API) is taken as before.

Each verdict records the dataset's input hashes, the gold's among them (`review/human.json`,
`inputs_sha256`). The owner sends the run's `review/human.json` from the data folder, and the developer
runs:

`python3 -I docs/kialo-batch2/verify_gold.py HUMAN.json`

For each verdict, the script finds the gold it was given on, by its recorded hash (of the file with LF
or with CRLF line endings: `.gitattributes` keeps both copies LF, but a checkout made before it may
have CRLF), in the working tree or in the repository's history, and compares the item's revision then
with its revision now. A verdict holds of the item as it is now when the two are equal, whatever else
in the file has changed since (other items' drafts, other verifications).

- It writes the item's `verified` (`by` "owner", `date` the verdict's, a UTC date, `item_rev` and
  `text_sha256` computed now) for a verdict *The gold is right* (`choice` "gold") that holds of the
  item now, when the item is agreed in `d086-rounds.json` at the revision it has, its coding waits for
  no decision of the owner (`pending`, §11), and it has no unresolved value (R17).
- It removes `verified` for a verdict of another choice that holds of a verified item now: the owner's
  later verdict takes the verification back.
- It records a verdict of another choice on an item not verified in the history entry only, as a
  receipt (the command says "receipt recorded"). The item does not change, but the verdict's time is
  kept, so that an older verdict *gold* imported later cannot verify the item.
- A verdict of another choice whose gold it cannot place (one the repository does not hold, one that
  is not valid JSON, or none recorded) it treats as any verdict of that choice, since such a verdict
  can only leave the item unverified: it records the verdict, or takes a verification back, by the
  order below, with the item's revision unknown, and lists it, to verify the item again.
- It lists every verdict it does not write, and why:
  - the item changed since the verdict;
  - a verdict *gold* given on a gold it cannot place (as above), which it neither writes nor records;
  - a claim not agreed in a round, or not in the gold, or whose coding waits for a decision of the
    owner;
  - a verdict older than one already recorded for its item, or not ordered with it (below);
  - a verdict the gold already records (its claim, choice and time).
- It writes nothing at all when:
  - any other input the verdicts recorded differs from the repository's now (the claims' texts, the
    guidelines, `walton.yml`, …: the round's clarifications of the guidelines are made before the
    owner's build, not between it and the script's run);
  - an input is one the built-in dataset does not have (a folder dataset's "guidelines.md (built
    in)", say), or a verdict lacks one of the inputs' hashes;
  - the two copies of the gold differ, or the gold is not valid JSON, or two of its items name one
    claim;
  - the gold's history holds a receipt that is not one (`verification_verdicts`, below), or an item's
    `verified` does not have R17's fields;
  - `d086-rounds.json` cannot be read, or a claim's state in it is not one (a review state, the
    revision agreed on exactly when it is agreed, the decisions it waits for by name);
  - `items.json` is not the claims of the dataset;
  - the file is not a list of verdicts, or a JSON object in it repeats a member;
  - a verdict lacks its claim, choice, time or inputs, or its time is not an RFC 3339 time with a
    time zone, or not one of the calendar's (years 1 to 9999, in UTC too);
  - two verdicts name one claim.

It reads every file whole and checks it before it writes anything, so that a malformed input stops it
with a message (a malformed gold a verdict was given on only lists that verdict, as above), never with
half a file written or a trace of the program.

It writes `gold.json` and the built-in copy as the recoding scripts and `validate.py --repin` do
(`json.dumps(…, indent=1, ensure_ascii=False)`), changing only the `verified` fields, the file's
`status` (how many items are verified) and one `history` entry. The entry names what was verified,
taken back or recorded, and holds the verdicts' times (`verification_verdicts`,
[annotation-records.md](annotation-records.md) §2).
Its `item_rev` is a copy of `validate.py`'s (which runs its whole check when imported); the tests run
`validate.py` on the gold the script wrote, and R17 compares the two. Its test, `test_verify_gold.py`,
runs in CI beside `test_validate.py`.

The verdicts the script used are committed in the repository, `docs/kialo-batch2/verdicts/`, with
their notes (the plan's "the dashboard's verdicts are the record"; the owner is told that they become
public). A verified item that a later change touches is stale (R17) until the owner verifies it
again. R13 refuses a verified item whose target lost its formula, and R2 one whose question no longer
applies there; a target recoded with both kept leaves the verification standing, since what was
verified still holds of it.

### Order: the verdicts' times

The dashboard stamps each verdict with its time (RFC 3339, to the nanosecond, with a time zone). A
verification's `date` is the UTC date of that time.

The script records each verdict it applies or records in the history entry's
`verification_verdicts`: the claim, the choice, the time in UTC and the item's revision (null for a
verdict whose gold it cannot place). It compares each new verdict with the latest so recorded for its
item:

- **An older verdict is not written.** An old file imported again cannot undo a newer verdict.
- **A verdict already recorded is not written again.** One whose claim, choice and time a receipt
  holds changes nothing, so that a file imported again changes no byte, whatever its order.
- **Verdicts of different choices at the same time are not ordered.** The item is left unverified,
  whichever comes first: a verdict *gold* at that time is not written, and a verdict of another
  choice at that time takes a verification back. This lasts until a newer verdict. Of two recorded
  verdicts at one time, the one other than *gold* counts as the latest.
- **A verification's time is recorded** when the latest verdict the history records for its item is
  a *gold* of the verification's date, at the revision verified: the verdict that wrote it, or a
  confirmation of it. Otherwise (made by hand, made before the times were kept, or made by hand
  after a later verdict was recorded) it has only its date. An older receipt *gold* of the same date
  and revision does not record its time: that verification may have been taken back and made again.
- **A confirmation records its time.** A verdict *gold* on a verification whose time is recorded,
  later the same day, records its time, so that a verdict between the two cannot take the
  verification back.
- **A verification with only a date cannot be ordered with a verdict of that day.** A verdict of an
  earlier day is older, and is not written. A verdict of another choice on that day is not written.
  Nor is a verdict *gold* on that day that would leave the verification as it is: it is not recorded,
  since its receipt would seem to record the verification's time. A verdict of a later day is
  ordered after it: a *gold* writes the verification anew, its time recorded; one of another choice
  takes it back. All this whatever else the history records for the item.

These receipts keep the order of the imports; they do not authenticate the owner. The run's
`human.json`, committed in `verdicts/`, stays the record of the verdicts themselves.

## 6. Cases (outcomes written before the tests)

The script:

| Case | Setup | Expected |
|---|---|---|
| V1 | Verdicts *gold* on two claims agreed in a round, given on the gold as it is | Both items get `verified` (`by` owner, the verdict's UTC date, the item's revision and its text's hash now); `validate.py` passes (R17); the file is byte for byte the file before but for the two fields, the `status` and one `history` entry; the built-in copy equal to it (the test's own output) |
| V2 | A verdict given on an earlier gold, in which the item differs | Not written; listed: the item changed since |
| V3 | A verdict given on an earlier gold, in which the item is the same (another item differs) | Written, as V1 |
| V4 | A verdict given on a gold the repository does not hold | Not written; listed |
| V5 | Verdicts *none*, *unsure*, an annotator's name, *reviewer*, and an empty choice with a note, on items not verified | The items unchanged. One history entry names the five and records their times; each listed as "receipt recorded" |
| V6 | V1, then a verdict *none* on one of the two, given on the gold V1 wrote | Its `verified` removed, listed; the other kept |
| V7 | A verdict on a claim not in the gold (the dashboard cannot make one; the script still checks) | Not written; listed |
| V8 | The verdicts' recorded hash of `items.json` differs from the repository's now | Nothing written at all; the command names the input |
| V9 | The recorded gold hash is that of the same gold with CRLF line endings; and, separately, another input's is the CRLF one | As V1 |
| V10 | V1, then the item recoded | `validate.py` refuses it as stale (R17) |
| V11 | V1 run twice | The second run writes nothing and changes no byte |
| V12 | A `human.json` that is not a list of verdicts; a verdict without `inputs_sha256`; two verdicts on one claim | Refused, naming the file; nothing written |
| V13 | The built-in copy differs from `gold.json` | Refused; nothing written |
| V14 | A verdict *gold* on an item with an unresolved value | Not written; listed |
| V15 | A verdict *gold* on a claim not agreed in a round, and on one agreed at another revision than it has | Not written; listed |
| V16 | A verdict whose inputs lack the gold; one that records an input the built-in dataset does not have | The first listed and not written; the second: nothing written at all, the input named |
| V17 | A verdict without one input's hash (`items.json`); one that records only the gold's | Refused; nothing written |
| V18 | A verdict time with an impossible date (`2026-99-99garbage`); a date without a time (`2026-11-28`) | Refused; nothing written |
| V19 | A verdict *gold* at 23:30 on 28 November at UTC−6 | Verified, its date 2026-11-29 |
| V20 | A file in which a verdict repeats a member (`"choice"` twice) | Refused; nothing written |
| V21 | A verification of the 30th, then a verdict *none* of the 28th | Not written; the gold unchanged |
| V22 | A verdict *gold* of the 28th, a *none* of the 30th, then the *gold* of the 28th again | The third not written; the item unverified |
| V23 | A verdict *none* of the 30th on an item not verified, then a verdict *gold* of the 28th from another run | The first recorded, the item unchanged; the second not written |
| V24 | *gold* of the 27th, *none* of the 28th, *none* of the 30th, then *gold* of the 29th | The last not written; the item unverified |
| V25 | A verification at 10:00, its confirmation at 15:00 the same day, then a *none* at 14:00 | The *none* not written; the item verified |
| V26 | A *gold* and a *none* at the same nanosecond, imported in either order; then each imported again | The item unverified after both, in either order; the repeats change no byte |
| V27 | A *gold*, then a *none* one nanosecond later | The verification taken back |
| V28 | A *none* of the 28th, then a *gold* of the 30th | Verified |
| V29 | The same *none* imported twice | The second changes no byte |
| V30 | A verification that records only its date: a *none* earlier the same day; separately, a *none* of the 28th recorded, a verification of the 30th made by hand, then a *none* of the 29th | Neither written; the verification stands |
| V31 | One run: a *gold* on one claim, a *none* and an *unsure* on two claims not verified | One history entry naming the verification and the two recorded verdicts, with three times |
| V32 | A gold that repeats a member; separately, a verdict given on an earlier gold (given with `--also`) that is not valid JSON | The first refused, nothing written; the second not written, listed |
| V33 | A verification at a time, then removed by hand; a *none* at the same time; then the *gold* imported again (added with the code: V26 does not reach a verdict that ties with a *gold* on an item no longer verified) | The *none* recorded; the *gold* not written; the item unverified |
| V34 | A verdict *gold* on a claim agreed in its round whose coding waits for an owner's decision (`d086-rounds.json`, `pending`); a *none* on such a claim; a `pending` that is not a list of names (added with draft 2, §11) | The *gold* not written, listed with the decisions it waits for; the *none* recorded as for any item; the malformed rounds file refused, nothing written |
| V35 | A verification of the 28th made by hand (no receipt); a *gold* at 08:00 that day; then a *none* at 12:00 that day (V35-V42 written after the independent review of draft 2, before their tests) | Neither written, each listed; the gold unchanged by both; the verification stands |
| V36 | A *gold* at 08:00 on the 28th (it verifies), a *none* at 12:00 (it takes the verification back), a verification of the 28th made by hand, then a *none* at 13:00 | The last not written, listed; the verification stands |
| V37 | A verification of the 28th made by hand; a *gold* of the 29th; then a *none* later on the 29th | The *gold* writes the verification anew, dated the 29th, its receipt recorded; the *none* takes it back |
| V38 | A *none* of the 30th given on a gold the repository does not hold, on an item not verified; then a *gold* of the 28th from another run | The *none* recorded, its `item_rev` null, listed to verify again; the *gold* not written; the item unverified |
| V39 | A verification at 10:00 on the 28th, then a *none* of the 30th given on a gold the repository does not hold; separately, on items not verified, a *gold* given on such a gold, a *none* given on a gold that is not valid JSON, and a *none* that recorded no gold | The *none* takes the verification back, listed to verify again; the *gold* neither written nor recorded; the other two recorded, each `item_rev` null, listed to verify again |
| V40 | A *none* and an *unsure* at one time on an item not verified, imported in turn, then each again | Two receipts; the repeats change no byte, each listed as already recorded |
| V41 | A gold whose history holds `verification_verdicts` null, or a number, or a receipt whose `item_rev` is a list, or a receipt *gold* whose `item_rev` is null, or a receipt whose time leaves the calendar in UTC; one in which an item's `verified` does not have R17's fields (its `item_rev` a list; a string); one with two items for one claim | Each refused, naming what is wrong; nothing written; no trace of the program |
| V42 | A verdict time that leaves the calendar in UTC (`0001-01-01T00:00:00+23:59`, `9999-12-31T23:59:59-23:59`); a rounds file nested deeper than the parser reads | Refused; nothing written; no trace of the program |
| V43 | A *gold* at 10:00 on the 28th (it verifies); the verification rewritten by hand with the date of the 30th; then a *none* at 10:00 on the 30th (V43-V46 written after the planted defects of the fixes, which four of the rules survived, before their tests) | The *none* not written, listed: the latest receipt is of another date, so the verification has only its date; the verification stands |
| V44 | A *gold* at 10:00 on the 28th (it verifies); the item recoded, agreed at its new revision and verified by hand that day; then a *none* at 12:00 that day, on the item as it is | The *none* not written, listed: the latest receipt is of another revision; the verification stands |
| V45 | A *gold* and a *none* at one time on the 28th (the item left unverified); a verification of that day made by hand; then a *none* later that day | The last not written, listed: of the two receipts at one time the *none* counts as the latest, so the verification has only its date; the verification stands |
| V46 | A gold whose history holds a receipt without its time, or one with a field more | Refused, naming the receipt; nothing written; no trace of the program |
| V47 | A verdict time whose offset is outside RFC 3339's ranges (`+00:60`, `-00:60`, `+01:99`; V47-V49 written after ChatGPT's review of exchange 2, its patch 01, before the tests are added) | Refused, naming the offset; nothing written; no trace of the program |
| V48 | A *none* of the year 0001 recorded; then a *gold* of 2026 | The receipt written with a four-digit year (`0001-01-01T00:00:00Z`) and read back by the next import; the item verified |
| V49 | A gold holding a number outside the finite range (`1e999`) | Refused, naming the number; the gold and its built-in copy unchanged; no trace of the program |
| V50 | A gold holding an integer too large for a float (`1` and 400 zeros), which Go's reader refuses (V50-V51 written after the independent review of step 1, before their tests) | Refused, naming the number; the gold and its built-in copy unchanged; no trace of the program |
| V51 | A number outside the finite range (`1e999`) in the verdicts file, or in the rounds file | Refused, naming the file; nothing written; no trace of the program |

The dashboard:

| Case | Setup | Expected |
|---|---|---|
| D1 | A run made with a list of claims in the *Claims* field | The run holds those claims, in that order; an unknown claim is refused and no run is made; with the field empty, the first N as before |
| D2 | A claim of the built-in batch-2 dataset, the gold shown | Under the gold's card, the whole item: each instance's id and premises, the challenges, the evidence's links, the notes and the adjudication's reasons |
| D3 | A verdict sent with the input hashes of the page | Taken when they are the run's; refused, with nothing saved, when one differs; taken as before when none are sent |

Each rule of the script and each of the three changes to the dashboard gets a planted defect that one
case must catch.

The replay (`to_wag.py`; added after ChatGPT's reviews of exchange 2, before the tests: its patches 03
and 04):

| Case | Setup | Expected |
|---|---|---|
| T1 | A claim coded Other, attacking a premise, played with Other to its own annotated conclusion | `as_intended` |
| T2 | The same, supporting the premise | `as_intended` |
| T3 | A claim whose scheme the game lacks, played with Other | `coverage_gap`, as before |
| T4 | A named question's answer that the game refuses, then played with Other at the statement | `failed`: the question's route was lost |
| T5 | A claim coded Other that plays a free new argument (`other-new`) | `coverage_gap`, as before |
| T6 | A claim coded Other whose executed conclusion is not its own annotated one | `failed`; corrected with F3 of the external review of exchange 3 (below) to `conclusion_changed` |
| T7 | A claim coded Other with no conclusion formula | `failed` |
| T8 | T7, the executed argument showing no conclusion either (T8 written after step 1's planted defects, which one rule survived, before its test) | `failed`: two missing conclusions are not the same conclusion |
| T9 | A claim coded Other whose annotated conclusion is spelled with spaces, or with `not(...)` for `¬`, played with Other to the same formula (T9 written after the independent review of step 1, before its test) | `as_intended`: the formulas are compared, not their spelling |
| U1 | A target whose formula, or part, or instance is `{"unresolved": why}` (each alone) | Not played: no move, `not_played`, the markers listed in the report |
| U2 | All three unresolved | As U1; the report keeps the annotation's scheme, conclusion and question |
| U3 | A bare denial (scheme `none`) whose target formula is unresolved | Not played: no challenge |
| U4 | A named question's answer whose target instance is unresolved | Not played: not answered at the target claim's primary argument |
| U5 | The same claim with its target resolved | Played as before |
| U6 | An inference target, its formula null (not unresolved) | Played as before, as an undercutter |
| U7 | A target whose instance, part or formula is neither text, null nor `{"unresolved": why}`: a blank reason, `{"unstated": why}`, a list, a marker with a second key (U7-U8 written after the independent review of step 1, before their tests) | Not played: no move, `not_played`, the fields listed in the report; the replay goes on. `validate.py` refuses each (R16) |
| U8 | U4's claim with its target resolved | Answered through its question at the target's argument: played `cq`, `as_intended` |

Agreement (D102: an unsettled relation is not scored):

| Case | Setup | Expected |
|---|---|---|
| A1 | One file's item has its target formula unresolved; the other's is resolved | That item's role, question and target part, instance and formula left out and counted as unresolved; its scheme, fit, polarity and target claim compared as before |
| A2 | The same with the target part unresolved | As A1 |
| A3 | Only the target instance unresolved (a model's "not yet annotated") | Only the target instance left out, as before; role and question compared |
| A4 | The Go agreement (`FieldIn`) on A1-A3 | The same as `agreement.py` |
| A5 | A marker whose reason is only characters that Go does not trim (U+001C to U+001F) (A5-A7 written after the independent review of step 1, before their tests) | A marker, as in Go (`strings.TrimSpace`): in `agreement.py`, in `validate.py` and in the replay |
| A6 | An item whose target is not an object (text, a list) | D102's check finds no unsettled relation, as Go's does; no trace of the program |
| A7 | The role counts, and the pilot's role comparison, of an item whose relation is unsettled | Its role counted as unresolved (`?`) and left out of the pilot's role agreement |

The argument the discussion needs (D096; `docs/annotation-records.md` §13; written before the tests
and the code):

| Case | Setup | Expected |
|---|---|---|
| RC1 | The gold with the thesis's instance by the discussion (13292.1#2) and its claims recoded | `validate.py` passes; the Go check reports nothing on it |
| RC2 | `by` with another value (`"the claim"`, `true`) | Refused (R16), in `validate.py` and the Go check |
| RC3 | `by` on the claim's primary instance | Refused (R20), in both |
| RC4 | Its fit `explicit` or `fragment` | Refused (R20), in both |
| RC5 | One of its premises `explicit` | Refused (R20), in both |
| RC6 | None of its premises `inferred` (all supplied) | Refused (R20), in `validate.py` |
| RC7 | It concludes neither the claim's statement nor its negation | Refused (R20), in both |
| RC8 | No item of the batch targets it | Refused (R20), in `validate.py` |
| RC9 | Another item's inferred premise cites the claim it is recorded at, for a formula only the reconstruction has | Refused (R20), in `validate.py`; citing that claim for a formula of its own instance is allowed |
| RC10 | Agreement: two files whose items differ only in an instance by the discussion | The `instances` field agrees (it is left out), in `agreement.py` and the Go measure; a target at it is compared by its scheme and conclusion, as at a secondary instance |
| RC11 | The replay: a thesis with an instance by the discussion, and a claim at one of its premises | The instance is played first, by Pro, with its scheme, for the main claim; the claim is played at it, not as a cascade; the report lists the instance apart, and the claims' counts are those of the claims |
| RC12 | The replay: an instance by the discussion at a claim other than the thesis | Played right after that claim's own move, for that claim's statement |
| RC13 | The replay: an instance by the discussion that the game refuses | Reported as not played, with the error; the claims at it are cascades, as at any instance not on the board |
| RC14 | The replay: a bare denial at the premise a fragment states, which is the premise it backs in another claim's instance (13292.135 and .137 at 13292.108's circumstances, which 13292.1#2 has) (RC14 and RC15 written after the replay from e1d6aa6 showed the two failing, before their tests and the code) | A challenge of that premise in the argument the fragment backs, where it is a premise on the board, not in the fragment's own Other argument, whose conclusion it is; a second denial of the same premise is a duplicate |
| RC15 | The replay: a move at a premise of a fragment that is not the premise it backs (13292.192 at 13292.6's stated link, whose goal is narrower than that of the premise it backs) | Not moved to the argument the fragment backs: played as before (that statement is not on the board) |

The English of a gold and the templates of an applied proposal (the independent review of D097's
application, W6: a changed template passed every check once re-pinned; written before the tests and
the code):

| Case | Setup | Expected |
|---|---|---|
| EN1 | The gold, 13292.61#1's conclusion re-rendered (it had the game's phrasing) | `validate.py` passes R21: the English of every positive formula over lexicon atoms is `walton.yml`'s template for it, each atom in single quotes as the lexicon gives it |
| EN2 | One word changed in a checked English (an exception's) | Refused (R21), naming the item, the instance, the formula and the English the template gives |
| EN3 | `walton.yml`'s `interference/2` template changed in a copy, the gold not rendered again (the review's mutation) | Refused (R21) at each English that renders it; `--repin` renews no pin |
| EN4 | A negated formula (`¬should_be_performed(tax_meat)`, "should not be done") | Not checked: `walton.yml` has no template for a negation |
| EN5 | A nested formula (`based_on_evidence(asserts(W,S))`) or a bare proposition atom | Not checked |
| EN6 | An Other instance's conclusion over lexicon atoms (13292.182's exception) | Checked as any instance's |
| EN7 | A file that is not a gold (an annotator's) with English of its own | Not refused by R21: the rule is the gold's, whose English the tooling writes |
| PC1 | An applied proposal whose template is not `walton.yml`'s (D040-D071-D097 with `interference/2` changed in a copy) | `wag-cli --proposal`: an error, the proposal is applied but `walton.yml`'s template is not as proposed |
| PC2 | The same template in a proposal not applied | No such error: it proposes the template |
| PC3 | Every committed applied proposal | No such error (`TestCommittedProposals`) |

The independent review of D096's application (written before their tests and the code):

| Case | Setup | Expected |
|---|---|---|
| RC16 | The replay: a claim whose statement is a negation (¬S), with an instance by the discussion concluding S | Played as an argument for S (support S), by the side opposite the claim's: the move follows the instance's own conclusion |
| RC17 | The replay: an instance by the discussion concluding neither its claim's statement nor its negation | Not played; the report says why (R20 refuses it in a gold) |
| RC18 | An instance by the discussion whose inferred premise cites another claim at which a reconstruction is recorded, for a formula only that reconstruction has | Refused (R20), as for any other instance |
| RC19 | The only target at an instance by the discussion is its own claim's | Refused (R20): another item must target it |
| RC20 | The replay: a fragment whose own Other move the game refuses, and a denial of the premise it backs | The denial is made in the argument the fragment backs, as when the fragment's move is accepted (RC14) |
| RC21 | A formula with a character Go's `strings.Fields` does not take for space (U+001C) | Refused by `validate.py`, as by the Go check: both read the same characters as space |
| CT1-CT5 | The Go check (format 2) on an inferred premise whose `cite` is missing, null, a number, the claim itself, or a claim the dataset does not list | Refused (R19), as `validate.py` refuses it |
| CT6-CT8 | The same, citing a claim of the batch, a claim of a claim's context (a sibling the dataset lists), or a known claim that does not say the premise | Accepted: the check is of the cited identity, not of what the claim says |
| CT9 | A format-1 file with the same premise | No R19: format 1 keeps the rules it was made with |

The independent reviews of 28 November: of D098 and R21 (verify/d098/REVIEW-batch.md) and the
check of the D096 fixes and draft 3, part 4 (verify/d096/FIXCHECK.md); written before their tests
and the code:

| Case | Setup | Expected |
|---|---|---|
| EN8 | One word changed in the English of a checked formula of arity 1 (`bad(G)` or `good(G)`), and of arity 3 (13292.61#1's conclusion) | Refused (R21), each |
| EN9 | One word changed in an explicit premise's English, and in an assumption's | Refused (R21), each |
| EN10 | One word changed in the English of a formula whose atom has digits (13292.224#1's `expert(...)`) | Refused (R21) |
| EN11 | A checked formula written with a space after its comma, its English changed | Refused (R21): the formula is read without its spaces, as Go reads it |
| EN12 | A checked English changed in its case only (its first letter), or in its final period only | Refused (R21), each |
| EN13 | A copy of `walton.yml` whose templates use `%[n]s` followed by `%s`, and `%%`; and one whose template `Render` refuses (`%[0]s`) | R21 renders them as Go's `language.Render` does (`%[3]s, %[1]s then %s` gives the third, first and second arguments; `%%` a literal `%`), and checks nothing where `Render` refuses; Go's `Render` gives the same strings |
| CT10 | The Go check on a dataset whose items' context lists do not name another item of the batch, a premise citing that item's claim | Accepted: a claim of the batch is known |
| CT11 | The same, citing a claim listed only among an item's children | Accepted |
| CT12 | The same, citing a claim listed only among an item's `kialo_parents`, and one listed only among its ancestors | Accepted, each |
| CT13 | A premise of role exception, and one of role assumption, inferred, whose `cite` is the claim itself | Refused (R19), each: premises of any role |
| CT14 | The same premise in an instance that is not the primary one | Refused (R19) |
| CT15 | No dataset: an inferred premise with no `cite`, one citing the claim itself, and one citing an unknown claim | The first two refused (R19); the third accepted (without a dataset no claim is unknown) |
| CT16 | A dataset whose item's siblings list is `["junk", {"claim_id": "13292.888888"}]`, a premise citing 13292.888888 | Accepted, as by `validate.py`: an entry that is not a claim is skipped, the others kept |
| RC22 | The replay: a claim whose statement is a negation (¬S), with an instance by the discussion concluding ¬S | Played as an argument for ¬S (support ¬S), by the claim's own side |
| RC23 | `validate.py` on an inferred premise whose `cite` is a list, and one whose `cite` is an object | An R19 error line each, no traceback |
| RS1 | `TestRuleCases` on a case that names its target's instance (OC6-e, 13292.127#2; OC6-f, 13292.12#3; OC10-a, 13292.127#1), the gold's target moved to another instance of the same claim | Refused: the case's instance is compared |
| PC4 | An applied proposal whose template is not `walton.yml`'s, where a later applied proposal carries `walton.yml`'s template | No error; the view marks the template changed since the proposal was applied, naming the later proposal |
| PC5 | The same with no later applied proposal carrying it | An error, as PC1 |

Decision D098 (OC6 and OC10; rule cases in `rule-cases.yml`, added with the recode that applies it):
OC6-e, 13292.202, a mitigation at the conclusion of the argument that cites the harm, the starvation
argument 13292.127#2; OC6-f, 13292.45, an amended action at the conclusion of the consequence
argument 13292.12#3, not at the causal step 13292.12#1 that predicts the harm (as 13292.194);
OC10-a, 13292.204, a denial of the link a weighing relies on, at the weighing's own premise
(13292.127#1), not at the consequence argument that also has it.

The external review of exchange 3, F5 (a malformed `items.json`: `validate.py` stopped with a
traceback on a context field that is a text, and read `""`, `0` or `{}` as an empty list; Go skipped
a field that is not a list, and R7 for a claim whose `kialo_parents` it could not read). ChatGPT's
patch 03, with the changes of the workspace's check of it (`verify/chatgpt-x3-2026-11-28/
CHECK-annotation.md` §3); written before their tests and the code. "A later claim" is 13292.115, the
sixth:

| Case | Setup | Expected |
|---|---|---|
| SC1 | `validate.py` with an `items.json` whose `ancestors`, `siblings` or `children` is a text, a number, `true`, an object, `""`, `0`, `false` or `{}`, at the thesis (13292.1) and at a later claim | Refused, each: one line naming the claim and the field, exit 1, no traceback |
| SC2 | The same field null, or absent, at either claim | Accepted: no claim of its context |
| SC3 | `validate.py` with `kialo_parents` absent, null, a text, a number, an object, `""`, `0`, `false` or `{}`, at the thesis and at a later claim | Refused, each, with one line: `validate.py` finds the thesis by its empty list (`[]`) and R7 reads every parent |
| SC4 | `validate.py` with a `kialo_parents` entry that is a text, null, or an object whose `claim_id` or `relation` is absent, null or a number | Refused, each, with one line naming the entry |
| SC5 | `validate.py` with `items.json`'s `items` not a list, or holding a text; an item with no `kialo_claim_id`, a number, or one of spaces (a tab and U+00A0, which Go trims); two items with one id; an item whose `text` is absent or a number | Refused, each, with one line, exit 1, no traceback |
| SC5b | `validate.py` with an `items.json` that is not JSON (added with the code, after the cases above) | Refused with one line, exit 1, no traceback |
| SC6 | `validate.py` on CT16's dataset: a claim's siblings `["junk", {"claim_id": "13292.888888"}]`, and 13292.25#1's inferred premise citing 13292.888888 | Accepted, as by Go (CT16): an entry that is not a claim is skipped, the others kept; without that sibling, refused (R19) |
| SC7 | Go reading a dataset (`FromDir`) whose claim's `ancestors`, `siblings`, `children` or `kialo_parents` is a text, a number, `true`, an object, `""` or `{}`, or whose Kialo parent is a text, null, or an object whose `claim_id` or `relation` is absent, null or a number | Refused when the dataset is read, naming the claim and the field: R7 no longer skips a claim whose parents it cannot read |
| SC8 | Go reading a dataset whose claims have these four fields null or absent, and one whose claims have none of them (a dataset not made from Kialo) | Read: no Kialo parent, no claim of context |
| SC9 | Go: R7 on a claim whose parents are written only under `Kialo_Parents` (another spelling) | Not its Kialo parents: Go reads the context by its names as written, as R19 and `validate.py` do (before, R7 read any spelling and R19 one) |
| SC10 | Go reading a dataset whose claim's `text` is absent, null or a number; and one whose only text is spelt `Text` | The first three refused when the dataset is read, each saying the claim has no text (before, an absent text was read as empty, and a number was reported as a missing id); the last read, as `TestALoneSpellingLoads` has it |

Go reads a claim with no `kialo_parents` as having no Kialo parent (SC8), where `validate.py`
refuses it (SC3): Go reads any dataset, and finds the thesis by the discussion's
`thesis_claim_id`; `validate.py` checks the Kialo batch, whose every claim has the list. A dataset
that a build of 12a15b7 or earlier read, skipping a malformed field, is refused by name; such a
dataset is not WAG's (its Kialo import writes every list, and the built-in dataset passes).

The external review of exchange 3, F3 (a move the game took in the intended scheme and role that
concludes another proposition: 13292.224 and .233, played with expert opinion, conclude the statement
they answer, since the engine binds expert opinion's answer variable to it, `carneades/answers.go`,
and the replay labelled them `as_intended`). ChatGPT's patch 01 labelled them `failed`, which the
README keeps for a move the game refused; with the changes of the workspace's check of it
(`verify/chatgpt-x3-2026-11-28/CHECK-annotation.md` §1), a label of its own. Written before their
tests and the code:

| Case | Setup | Expected |
|---|---|---|
| T10 | A claim with a named scheme the game has, played for or against a statement, whose argument concludes another formula than its annotated conclusion | `conclusion_changed`, the row giving both formulas (`conclusion_change`) |
| T11 | The same, its argument concluding the annotated conclusion | `as_intended` |
| T12 | The same, its annotated conclusion null | `failed`, as T7: nothing to compare |
| T13 | An answer through its question, with a named scheme and with Other, whose argument concludes another formula | `conclusion_changed`, each |
| T14 | T13, its argument concluding the annotated conclusion | `as_intended`, each |
| T15 | An answer at an exception, against the argument and for it, its annotated conclusion the exception and its negation (R8) | `as_intended`, each: the engine concludes the exception for an attack and its negation for a defence |
| T16 | A new argument (no target statement), concluding its annotated conclusion, and another | `as_intended`; `conclusion_changed` |
| T17 | A coverage gap played with Other at its target, its argument concluding another formula than the claim's own, and the same formula; a challenge | Labelled `coverage_gap`, with `conclusion_preserved` false, and true; a challenge has no `conclusion_preserved` (it makes no argument) |
| T18 | The batch's replay | 13292.224 and .233 `conclusion_changed`; every other row's label as before; the summary counts the label |
| T19 | An undercutter of an inference (OC18), its argument concluding `undercut(<its target>)` | `as_intended`, not compared: no `conclusion_preserved` |

T6 is corrected: a claim coded Other whose argument concludes another formula than its own is
`conclusion_changed` (it was `failed`): a changed conclusion has one label, for Other as for a named
scheme. T7 and T8, with no conclusion to compare, stay `failed`. U8's fixture is corrected: its
capturing server answered a premise question with `undercut(a1)`, and its answer concluded an
unrelated formula, where the engine concludes the premise or its negation (`answers.go`); U8's
expected outcome does not change.

The external review of exchange 3, F6 (the agreement on the target instance names a secondary target
by its scheme and conclusion, so two different reconstructions of one claim with the same scheme and
conclusion agree: 13292.5 targeting 13292.1#2, the environmental goal, in one file and 13292.1#3, the
health goal, in another, both `13292.1 secondary:practical_reasoning2:should_be_performed(tax_meat)`).
The check's recommendation (`verify/chatgpt-x3-2026-11-28/CHECK-annotation.md` §4): keep
`target_instance` as the historical measure, and add a versioned field, `target_instance_v2`, which
names a secondary target by its structure: its scheme, its conclusion, its premises (each role and
formula, sorted) and who made it (`by`). Written before their tests and the code:

| Case | Setup | Expected |
|---|---|---|
| TS1 | Two files whose item targets a secondary instance of one claim, with one scheme and conclusion and different premises (the review's 13292.5) | `target_instance` agrees, as before; `target_instance_v2` disagrees |
| TS2 | The same secondary instance in both files, its premises in another order, a formula written with spaces | `target_instance_v2` agrees |
| TS3 | Both files targeting the claim's primary instance, coded differently | Both fields agree (`13292.1 primary`): a claim's own coding is its own fields' business |
| TS4 | A secondary target with a premise whose formula is `{"unresolved": why}` | `target_instance_v2` unresolved (left out, counted); `target_instance` compared |
| TS5 | A secondary target that is an instance by the discussion, and one of the claim's own with the same scheme, conclusion and premises | `target_instance_v2` disagrees (`by`) |
| TS6 | An item whose relation is unsettled (D102) | `target_instance_v2` left out, as `target_instance` |
| TS7 | The Go measure (`FieldIn`) on TS1-TS6, and on the batch's two annotators | The same values as `agreement.py`; `agreement.json` regenerated by the README's command holds the field, and the Go figures equal it |

## 7. Out of this plan

Batch 2 recoded under the guidelines' version 4 (plan §4, item 2), other discussions and the probe set
(items 3 and 4), the bounded pilot (stage 3) and D087's order; any change to `walton.yml`, which goes
through a proposal as always; and the replay's playing of the secondary instances other claims target
(§3).

## 8. Changes from version 1

The independent check of version 1 (in the workspace, `verify/d086/`):

| Finding | Taken up |
|---|---|
| E-1 The controls were drawn from where A and B agree (22 claims), not where A, B and the reference agree (11); 13292.61 and .45 differ from the reference on the target's formula; the draw depended on an unstated order | The population as the plan says, in `items.json`'s order, stated (§2); .61 and .45 go to round 2 |
| E-2 The claims where the reference differs from the annotators were left out | The three-way measure (§2): ten more claims in round 1 |
| E-3 The dashboard shows the gold's primary instance only, not all a verification covers | The whole item under the gold's card (§5, D2) |
| E-4 A run of the round's claims cannot be made in the dashboard, and the script would write a verdict on a claim no one drafted | The *Claims* field (D1); the script writes only items agreed in a round at their revision (V15) |
| W-1 Any change to the gold voided the verdicts not yet written | The verdict's gold found by its hash and compared item by item (V2, V3, V4) |
| W-2 A verdict given on a stale page records data the owner never saw | The page's hashes sent with the verdict (D3) |
| W-3 Missing cases | V5, V6, V9, V12-V16; V8 of version 1 reads the test's own output (V1) |
| W-4 Items outside the round; R17 and recoded targets | §3 (items outside join the round), §5 (the order; what R13 and R2 refuse) |
| W-5 The Go tests and rule cases that pin the gold; the replay's cascades | §3 |
| W-6 The draft is not independent of the reference | §3, said |
| W-7 Convergence undefined; the package lacked files R2 and R16 read | §4: per item, at most three exchanges, then the owner; the files added |
| N-1 Phase 3, not stage 3; N-2 `.gitattributes`; N-3 the record public; N-4 `item_rev` a copy; N-5 UTC dates; N-6 the test service; N-7 `agreement.json`; N-8 the first pilot; N-9 `walton.yml` and D043 | The status line, §5, §3, §2 and §4 |
| N-10 The script and its tests were written during the check | They are rewritten for this version and checked with it |

## 9. Results (the cases above, run after this version was committed, 0a09e32)

- **The rounds.** `d086_rounds.py gold.json --init` wrote `d086-rounds.json` from the gold before any
  draft (SHA-256 `fc929c7e…`): the lists of §2, 28, 15 and 7 claims, the controls 13292.1, .30 and .36;
  run again without `--init`, it finds the file the same.
- **The script.** V1-V16 pass, and three more cases: V3 with the earlier gold found in the repository's
  history rather than given; a rounds file it cannot read (refused, nothing written); and the
  repository's rounds file (every claim of `items.json` in one round, readable by the script). 25
  planted defects, one for each rule, are all caught. Corrected on the way: the first run's record
  marked every kill as the script raising, since the runner took the test file's name for the
  script's; read again, one defect is caught only by the script raising (a verdict without its inputs,
  whose check removed leaves the script to fail on the missing field, which the case still catches).
- **The dashboard.** D1 and D3 pass in the server's tests (`TestD1ARunOfTheClaimsListed`,
  `TestD3AVerdictGivenOnAStalePage`, which also refuses a page with an input fewer), and D1-D3 in the
  browser: the claims field makes the run listed and refuses an unknown claim, the whole item shows
  every instance id, challenge, link, the notes and the adjudication's reason, and a verdict carries
  the page's hashes and is refused on a stale page with nothing saved. Eight planted defects (four in
  the page, four in the server) are all caught, none by a panic. Corrected on the way: in the first
  run the browser check stopped when the run it made lacked the claim it opens (the defect "the claims
  field not sent" had no verdict); it now records that and goes on with a run made through the API,
  and catches it.

## 10. Round 1, draft 1 (2026-11-27)

The developer's readings came first (in the workspace, `verify/d086/round1-readings.md`: the 16 claims
of version 1's round, then the 12 new to version 2's), then the comparison with the reference, A and B.
`recode_d086_r1.py` wrote the draft into the gold, each item with its `adjudication.recoded` entry;
`d086-rounds.json` has every claim of round 1 at draft 1, not reviewed. Its first version (6e9b28c) was
checked independently before any review (the workspace's `verify/d086/DRAFT1-CHECK.md`: three errors,
five warnings, eight notes, each confirmed against the files), and so were its corrections (b888983;
`verify/d086/FIXCHECK-draft1.md`: three errors, two warnings, notes); the corrected script runs again
from the reference each time, so draft 1 is one record. What the corrections changed is listed at the
end.

- **Changed (8).** 13292.12 at 13292.9#2's `causes(tax_meat,meat_industry_harmed)` and 13292.141 at
  13292.2#2's `causes(meat_production,significant_ghg_emissions)`, both the book's 28 CQ1 (a mechanism
  offered for the link), for; 13292.182 and .198 at 13292.180#2's exception
  `interference(reduced_meat_consumption)` (28 CQ3), against, their own arguments concluding it (R8);
  13292.202, a mitigation, at the conclusion of the argument that cites the harm it mitigates,
  13292.127#2's `¬should_be_performed(tax_meat)` (OC6): each names what it answers, as both annotators
  aimed it, where format 1's rule of a primary instance's formulas had kept it from that. 13292.6, .11
  and .218 bind only the variables of the formulas they state, list only those formulas, and leave
  their conclusions, which name the action, null; their challenge records at formulas they no longer
  list are left to the arguments they are fragments of (13292.108, .118), which record the same
  challenges.
- **Joined (3).** 13292.9, .2 and .180 (rounds 2 and 3 by the measure) had challenge records that named
  the retargeted claims at their old places. As §3 has it, an item a draft changes joins the round:
  those records are put right, the missing ones at the causal steps added (13292.9#2's 28 CQ1,
  13292.180#2's 28 CQ3), and the three are drafted and reviewed with round 1 (`d086-rounds.json` marks
  them `joined`); otherwise their codings are kept, but for the confidence of their capped causal steps
  (below).
- **Kept (20),** each with its reason. Where the developer's reading differed from the reference, the
  guidelines decide for the reference (13292.46: the argument from rules fits what the claim states;
  13292.8: 13292.115 is a consequence argument, not practical reasoning, so 13292.8 is no fragment of
  it; 13292.36: the claim adds a consequence of its own, so an independent argument, as all three
  codings have it, not backing of 13292.7's premise), or the question goes to the review (13292.194:
  the developer read it at 13292.12#3; the reference and both annotators have 13292.12#1's conclusion;
  question 3). 13292.108 keeps every variable bound (question 1).
- **Confidence.** Nine causal steps whose confidence the N8 note capped "until the first gold's draft"
  (13292.9#2 and #3, .12#1 and #2, .36#2 and #3, .180#2, .192#1 and #2) have it set on their merits
  (confidence is the estimate that a careful annotator would choose the same, guidelines §3): both annotators coded
  the same nine steps (with `cause_to_effect`, before D039), and §5 item 11 prescribes the modal form,
  so 0.55; 13292.192#2, whose link the text states only as a possibility ("may still occur"), which the
  scheme's firm link does not express, 0.5. Fit stays `reconstructed` (§2 item 2). Their notes say
  whether the cause is the proposed tax or an effect predicted to follow from it. 13292.14's two steps
  keep the cap until round 2.
- **Notes.** Notes and challenge notes that told a coding's history in dated additions, described a
  coding the item no longer has, or said that the game asks a question at the conclusion only "until
  workstream E" (it asks it there by design since step C5.2, the rebuttal route beside the exception's
  question) are rewritten as one current account; the history is in the recoded entries and
  `gold-format1.json`. So are the premise notes that did so (13292.9#1's, 13292.174#1's), or that said
  something no longer true or loose (13292.127#2's, 13292.108#1's); the boilerplate of exception and assumption notes, which
  names the decision a premise comes from ("Since 2026-11-24: ..."), is kept, as everywhere in the
  gold. 13292.147 names candidate C017. The gold's status says who made the material:
  language models (the batch's adjudication, and the developer's drafts), no item verified by a person
  yet; `verify_gold.py` writes the same words when no verification is left.
- **Checks.** `validate.py` (R2-R17) passes. Each item's record names exactly the fields its diff against
  the reference changed (`verify/d086/records_check.py`, which finds the one omission of the first
  version). `d086_rounds.py` gives the same rounds from the reference; the agreement of annotators A and
  B is unchanged. The tests that named items the draft changed now plant the same defects elsewhere:
  `test_validate.py`'s R13 test plants its missing formula in another instance of 13292.9, and its R14
  test picks an item that binds the agent and into which nothing points (it finds 13292.190); their Go
  counterparts use 13292.108, the gold's fully bound fragment, 13292.45 for the agent, and 13292.9#1;
  `TestFormat2OnePrimaryAndFragmentConclusions` still plants in 13292.6. The two rules still catch their
  planted defects (`verify/d086/mutants-d086-r1-tests.py`). If question 1 goes against the draft,
  the fully bound fragment cases need a constructed fragment. `rename()` leaves a fragment's null
  conclusion alone, and no longer reads it as a formula of the instance. `TestAgreementFieldsAcrossFormats`
  pins the new figures of the format-1 gold against the format-2 one (each change above accounted for)
  and the target-instance test lists the five new moves. `test_verify_gold.py`'s V15 sets the round
  state it tests instead of reading the repository's, and V6b checks the status written when every
  verification is taken back. The replay: 13292.12, .182, .198 and .202, played before, are cascades
  now, with 13292.141 (a cascade before too: the premise it targeted was not on the board): 42 claims
  played instead of 46; the fragments play as before; the thesis stays out.
- **The corrections of the first version (6e9b28c).** 13292.108's note and record counted eight targets
  that §5 item 6 would leave without a formula: seven (the four at the stated circumstances keep
  theirs). 13292.11's record omitted its instance note. The records of 13292.9, .2 and .180 still named
  the retargeted claims (hence Joined). 13292.141 read as data (28 CQ2), though it offers a mechanism
  and cites nothing (28 CQ1, as 13292.2's own record has it). 13292.202 now names the argument both
  annotators aimed at (question 2 asks the other way). 13292.12's record at its interference no longer
  says that 13292.45 and .194 answer there. 13292.25's note now says that the claim names a way to
  avoid the cost (its premise), and argues the cost too slight, not avoided by a substitute. The caps,
  the status, the challenge notes and the fragments' reasons for their unstated variables (13292.6's
  S1, 13292.218's A) as above.
- **The corrections of the corrections (b888983).** 13292.127#2's starvation premise said that the
  children engage it, which stopped being true when 13292.202 moved: it now says which reply engages
  what. This section said the developer's reading matched where it did not (13292.36, .194): now
  listed above. 13292.204's reference coding came the same way as 13292.202's (D055's recoding moved
  B's answer to CQ:likelihood at the starvation argument's premise to the weighing's premise): its note
  and record say so, and question 2 puts both to the review. The causal steps' notes said "only
  proposed" where the cause is an effect predicted to follow from the tax. The nine confidences now
  rest on what confidence means (guidelines §3): 0.55, and 0.5 for 13292.192#2. Smaller: two premise notes told a history
  or said loosely what bears on them (13292.174#1, 13292.108#1), 13292.180's note named a reply to
  13292.182 as one to both substitutes, and some records named their changes only by implication.

Open for the review and the owner, each said in the items' notes:

1. 13292.108: does a fragment whose parent has no argument carry that argument, every variable bound,
   or does §5 item 6 apply to it too, leaving seven of the eleven targets into it without a formula?
   Either way the guidelines should say.
2. 13292.202 and 13292.204, the replies to the weighing claim 13292.127. When D055 made the weighing
   13292.127's primary, format 1's rule of a primary instance's formulas moved both from the starvation
   argument (now 13292.127#2), where both annotators had put them, to the weighing (13292.127#1):
   13292.202, a mitigation, from the argument's conclusion to the starvation link, and 13292.204, which
   denies that link, from the argument's premise (B's answer to CQ:likelihood there) to the same link
   as the weighing's premise (a premise attack). The draft moves 13292.202 back, by OC6 (a mitigation
   at the conclusion of the argument it answers), and keeps 13292.204 at the weighing, by OC10 (a reply
   to a weighing claim attacks that claim's own instance: its comparison or the consequence it relies
   on). Is that right for each, and which rule governs a reply that both mitigates or denies a harm and
   answers a weighing claim?
3. 13292.194 (and 13292.45, round 2): a mitigation at 13292.12's predicted harm (its primary step's
   conclusion; the reference's and both annotators'), at 13292.12#3 (the argument that cites the harm
   against the tax), or a 28 CQ3 answer at 13292.12#1's interference (the mechanism blocked)? The draft
   keeps the first, where the reference and both annotators put it; the developer's own reading was the
   second (13292.12#3 is the argument that cites the harm against the tax, as 13292.127#2 is for
   13292.202). It does not take the third: the measures are the agent's own action (case.json), unlike
   the substitutes of 13292.182 and .198, which are what people would do.
4. 28 CQ1 or CQ2: the draft reads a mechanism offered for a causal link as answering CQ1 (13292.12,
   13292.141) and data offered for it as answering CQ2 (13292.46's beef figures bear on 13292.2#2's
   CQ2). Nothing checks the choice: the book's two questions sit at the same formula. Should the
   guidelines say so?
5. A fragment's premises: the draft lists only the formulas a fragment states, not the parent's others
   whose variables it binds (13292.6 lists no `possible_goal(G)`), nor a formula the claim leaves to
   the parent (13292.218 names the tax only as "it": the transition is 13292.118's). §2 item 3 says no
   premise may go unlisted because it is implicit; the draft reads §5 item 6 as the rule for a
   fragment, whose implicit premises are its parent's to list. Should the guidelines say so?
6. The replay's cascades: the replay plays each claim's primary instance and a secondary one that
   carries it, so a claim at another secondary instance is not played. Four of the five target causal
   steps, which the game does not play yet (`cause_to_effect` and its modal form are a coverage gap):
   a replay that played those steps could play them only with Other, whose premises and exceptions are
   not on the board, so they become playable when the game takes the causal schemes. 13292.202's
   target, the starvation argument 13292.127#2, is a consequence argument the game has: playing it is
   a change to `to_wag.py`, outside this plan (§7). Is there a reason the gold should not target
   secondary instances? (Corrected in §11: 13292.202 has scheme `none` and targets a conclusion,
   where the replay plays no claim without a scheme.)

Taken up in round 2, not in this draft: 13292.208's target, which format 1 also moved (to 13292.104#1's
better-means premise; both annotators put it at 13292.104#2's circumstances
`current_circumstances(farming_industry_not_meat_is_problem)`, which its note says it contradicts,
though the note's last sentence reads it at the moved target); 13292.121's role
kind `side_effect`, format 1's label (§5 item 3 makes it a value of `role.move`); 13292.14's capped
causal steps; and the notes of 13292.7, .115 and .121 that still say "until workstream E".

## 11. Round 1, draft 2 (2026-11-27)

**The review.** The other reviewer's item review of draft 1 (in the workspace, `reviews/2026-11-27/`;
checked in `verify/chatgpt-d086-2026-11-27/REVIEW.md`) agreed with 21 of the 31 items and objected to
10. The developer accepted six objections (one without its optional part), partly accepted three,
and agreed with the direction of the tenth, which needs a rule of the owner's first. (Corrected after
the independent check of draft 2, §12: this paragraph first counted 13292.8 as accepted.) The reviewer's addendum of the same day
reconciled the two partial answers on conditions, and listed which agreed items the owner's
decisions affect (`verify/ethix-2026-11-27/REVIEW.md` §2, checked).

**The draft.** `recode_d086_r1_d2.py` writes draft 2 from draft 1, each changed item with its
`adjudication.recoded` entry:

- **Accepted (6).**
  - 13292.25: the ease premise is supplied, the claim's hedge ("could be as small as") in its note;
    confidence no longer said to carry the hedge.
  - 13292.104:
    - #2's fit is reconstructed (its goal premise is supplied, §5 item 6);
    - #1's comparison is supplied, its note saying how it is reconstructed from the claim's
      diagnosis and that the fairness reason is not it;
    - #1's link is explicit: the claim's own "the appropriate solution", projected to 13292.108's
      goal. (It first cited 13292.13, which names the problems but does not say that regulating
      solves them: §12.)
  - 13292.43: the tax's effect on the goods it ranks is supplied (presupposed), their evaluation
    explicit.
  - 13292.127: the starvation link is supplied in both instances, since 13292.7 grounds only the
    dearer food; #1's evaluation is explicit, as #2's.
  - 13292.192: its second causal step is primary, since its conclusion reaches the target (§5
    item 5). The note says the chain shows only that some deforestation may occur, not that it does
    not slow.
  - 13292.194: at 13292.12#3's conclusion, the argument that cites the harm, as 13292.202 is (OC6).
    Question 3 is answered so. The reviewer's optional second relation, the measures as interference
    with 13292.12#1's mechanism, is not coded: a claim has one target, and the claim proposes a
    changed policy, not a blocker under the tax. 13292.12's challenge at that interference names the
    measures.
- **Partly accepted (3).**
  - 13292.8: the note says the backing of 13292.115's premise is partial, and the target is the
    nearest formula the format offers. Not taken: the reviewer's unresolved or approximate target.
    The format has no approximate target, and an unresolved one could not be verified (R17). How to
    target a background condition that backs a premise in part is a question for the guidelines.
  - On the reconciliation's conditions, 13292.123:
    - `better_means` kept, for 13292.118's own goal (D075);
    - its comparison and its link to the goal are supplied, in both instances, and scoped: subsidy
      removal corrects the distortion 13292.11 names, not the externalities 13292.218 names;
    - the reviewer's reading, a proposal of another course not said to do better (C017; OC6), is
      recorded as the primary's alternative.
  - On the reconciliation's conditions, 13292.218:
    - the action is named ("it") but not represented in the fragment;
    - the circumstances it elaborates are inferred from 13292.11, which states them;
    - its stated link stays explicit, its note giving the projection to 13292.118's
      `higher_meat_price` and its scope: the price rise that internalizes the named externalities, not
      any higher price.
- **Refinements of agreed items (2).** 13292.11 says the claim presents the subsidies as existing
  ("due to subsidies"). 13292.6 records `cause_to_effect_modal` as an alternative.
- **Corrected after the independent check (2, §12).** These agreed items' records were made false by
  draft 2's own changes:
  - 13292.204's note said the premise it attacks is inferred from 13292.7, which draft 2 made
    supplied;
  - 13292.12's challenge and notes placed 13292.194 at the predicted harm, which draft 2 moved it
    from; its quotation of 13292.9 also left out "trillion-dollar".

  Both go back to not reviewed.
- **To the owner (1).** 13292.108, unchanged: question 1.
- **Agreed at their revision (17).** The other agreed items, unchanged.

**The rounds file.** `d086-rounds.json` records:

- each claim's first exchange (`exchanges`: the draft reviewed, the reviewer's and the developer's
  dispositions, the record);
- the 13 changed claims at draft 2, not reviewed;
- the 17 agreed at their revision;
- 13292.108 with the owner;
- the owner's decisions 17 claims wait for (`pending`). `verify_gold.py` writes no verification while
  a claim waits (V34). The decisions:
  - **Q1** (13292.108, .5, .6, .104, .127, .147, .174): whether 13292.108 keeps its source fragment
    beside a separately identified reconstruction of the environmental argument. If so, the targets
    into its conclusion and its supplied premises move to the reconstruction, since a fragment
    carries no conclusion (13292.6#1, .11#1, .218#1).
  - **G2** (13292.2, .9, .11, .12, .36, .141, .180, .182, .192, .198): whether the causal exception
    names its effect, `interference(E1,E2)`, through a proposal. If so, every causal instance's
    exception changes, and so do the answers at 13292.180#2's.

  Three claims of round 2 depend on them too. Their `pending` is recorded when round 2 is drafted:
  13292.13 on Q1 (its target moves) and 13292.4 and .14 on G2 (they have causal instances).

The review's other questions for the owner change no item of this draft: the convention of OC6 and
OC10, C5's questions, E1's choices and the pilot's external benchmark.

**Checks.**

- `validate.py` gains two rules from the review's findings, with planted-defect tests:
  - R18, explicit fit (guidelines §5 item 6), which 13292.104#2 broke; it is in the Go check too;
  - R19, an inferred premise cites the claim that says it (§2 item 3), which 13292.104#1, .43#1,
    .123#1 and .127#1 broke. Since §12 it accepts every claim `items.json` lists, its context
    included.
  Both pass now.
- Each changed item's record names exactly the fields its diff changed (in the workspace,
  `verify/d086/records_check.py`: 13 items, no mismatch).
- Each quotation in the round's notes is the words of the claim it quotes (in the workspace,
  `verify/d086/rv/quotes_check.py`). What remains unmatched are labels, schematic forms
  ("A only for X"), `walton.yml`'s English and one term. Round 2's draft takes two misquotations
  found there (13292.4, .235).
- `TestAgreementFieldsAcrossFormats` pins target_instance 0.84 (13292.194) and bindings 0.92
  (13292.192's primary).
- In the replay 13292.194, a new argument before, is a cascade, since its target, 13292.12#3, is a
  secondary instance the replay does not play: 41 claims played.

**Questions for the next revision of the guidelines,** which the review raised and this draft does
not settle:

- an S2 close to the action restated (13292.104#2, §5 item 4);
- a fragment that elaborates a formula its context states (13292.218, §5 item 6 with §2 item 3);
- a background condition that backs a premise only in part, which no target can name (13292.8).

**A correction of draft 1's question 6 (§10).** It said that playing 13292.202's target was a change
to `to_wag.py`. But 13292.202 has scheme `none` and targets a conclusion. The replay plays a claim
without a scheme only as a challenge, when it denies a premise of an argument on the board (README),
so it would not play 13292.202 at 13292.127#2's conclusion even if it played that instance. It is
reported as a cascade, because that instance is not played. (The reviewer's finding; this note said
"in no case" until the independent check, §12.)

**Next.**

- Exchange 2, with the reviewer, on the 13 changed items; 13292.123 and .218 are measured against the
  conditions the reviewer set.
- The owner's decisions: Q1 and G2 first, since 17 claims wait for them.

## 12. The independent check of step 1 and draft 2 (2026-11-27)

An independent check of 0daebfe..db72226 (in the workspace, `verify/d086/rv/`) confirmed the
regeneration of draft 2, the rounds file, the R18 and R19 findings, the replay and the tests. It
found these, each verified against the code before it was fixed:

- **The verification import** (`verify_gold.py`).
  - **E1.** A verification made by hand counted as timed whenever any receipt *gold* had its date
    and revision. Two orders then let a verdict of that day take it back. A *gold* the same day was
    recorded as a confirmation, and a later *none* took the verification back. Or a verification
    was taken back, then made again by hand, and a *none* that day took it back. §5's rule now
    reads the item's latest receipt.
  - **W3.** A newer verdict whose gold could not be placed left no receipt, so an older *gold*
    still verified the item.
  - **N1.** Two verdicts of different choices at one time added a receipt on every import.
  - **W5.** Malformed inputs stopped the script with a trace of the program.
  - **N5.** Some messages were wrong.
  - **N2.** The tests had gaps.

  The rules and cases V35-V42 were committed first (3cabca9). Then the code and its tests.
- **The rules.** R19 refused a claim of a claim's context that `items.json` lists (W2). The R19 row
  of `annotation-records.md` gave the wrong reason why Go does not check it (N4).
- **Draft 2.** It made the records of two agreed items false: 13292.204's account of its premise
  (E2), and 13292.12's of where 13292.194 is (W1). It left three provenance statuses inconsistent
  (W7): 13292.123#2's link was still inferred from a claim that states only the goal; 13292.104#1
  cited a claim that does not say its link; 13292.11's explicit premise had a note headed
  "presupposed". Three quotations were not the claims' words (N6: 13292.218, .8, .104), and the
  developer found a fourth (13292.12). Two dispositions counted as accepted were partial (N3:
  13292.8, and 13292.194 without its optional part). The fixes are in §11.
- **Docs.** §5 did not mention `pending` or every input that stops the import (W4). The correction
  of question 6 overstated (W6). "17 claims wait" counted round 1 only (N7).
- **The replay.** It was built from a tree with uncommitted changes, so its build names 8c64ee9,
  modified (N8). The replay is now made by a server built from a clean checkout of the commit that
  holds the code.

**Planted defects.** 27 planted defects of the fixes were tried (in the workspace,
`verify/d086/rv/mutants-rv.py`). On the first run, five survived:

- four were rules no case reached: the receipt's date and revision, the latest of two receipts at
  one time, and a receipt's fields. Cases V43-V46 were written for them (a3b9141), before their
  tests;
- one, R19 without siblings, does the same as the correct code on today's `items.json`, where every
  sibling outside the batch is also listed as a child. It was replaced by R19 without children,
  which the test now catches with a claim listed only as a child (13292.149).

On the second run all 27 were caught.

**Results.** `test_verify_gold.py` 50 tests, `test_validate.py` 91, `test_agreement.py` 5 and
`test_to_wag_inference.py` 10 pass. `validate.py` passes the gold. The draft-2 script reproduces the
gold from draft 1, its records match the diff (13 items), and `d086_rounds.py` gives the same rounds.

## 13. The owner's decisions (2026-11-27)

Asked the decision batch's questions, each with a recommendation, the owner replied: "Yes. Based on
your analysis and recommendations, please proceed to close all remaining gaps." They are in the
register as D096-D101:

- **D096 (question 1): two layers.** 13292.108 keeps what it states; a separately identified
  reconstruction of the environmental argument, marked as the discussion's, takes the targets on the
  parts it does not state (13292.5, .147 and .174 at its conclusion; 13292.6, .13, .104 and .127 at
  its supplied premises). How a reconstruction is identified goes through a proposal.
- **D097 (G2): the causal exception names its effect,** `interference(E1,E2)`, through a proposal
  and `edits.py`; the gold's causal instances and challenges are recoded.
- **D098: the OC6 and OC10 convention** goes into the guidelines (§3.1).
- **D099: C5's questions,** answered as the other reviewer answered them (`c5-game.md`).
- **D100: E1's choices,** as implemented.
- **D101: EthiX first** as the pilot's external benchmark, a private track; no model is run on it
  until the owner sets a budget.

**Next.** Round 1's draft 3 takes in D096, D097 and D098: the proposals first, each checked
independently, then the recode, the rounds file (the `pending` claims released) and exchange 3 with
the other reviewer. The owner's timed verification session follows draft 3: a verdict is tied to the
guidelines and `walton.yml` it was given on (§5), and these decisions change both. Exchange 2 went to
the reviewer in one package, as the owner asked (wag-workspace
`verify/d086/package/WAG-review-2026-11-27.zip`).

## 14. After exchange 2: ChatGPT's two reviews and the owner's answer (2026-11-27)

ChatGPT answered exchange 2 (11 items agreed; objections to 13292.8 and .104), then reviewed the
developer's check of that answer together with WAG's argument from rules. Both reviews are checked in
the workspace (`verify/chatgpt-d086-x2-2026-11-27`, `verify/chatgpt-rule-2026-11-27`). The owner
answered the developer's recommendations: "Based on your analysis, please proceed to close all
remaining gaps. Ensure that everything is correct with appropriate validators and independent
reviewers." In the register:

- **D102: 13292.8's relation stays unresolved** until a connecting argument is justified and reviewed:
  not played, not scored as a relation, not verified, kept in every count; the guidelines say the
  same of any claim that states only a background condition of a premise.
- **D101 amended: the EthiX protocol's rule is version 2's** (a paired debate bootstrap; at most one
  item newly wrong in each field on the agreed gold).

**Step 1 (code, before draft 3).** The verification import's strict input (ChatGPT's patch 01, cases
V47-V49); the replay's label for a claim coded Other (patch 03, T1-T8); the replay's guard on an
unresolved target (patch 04, U1-U6): today a formula marker stops the replay with a traceback, an
instance marker plays the claim at its target claim's primary argument, and a part marker plays it as
a free argument; agreement's rule for an unsettled relation (D102, A1-A4).

**Draft 3** takes in, besides D096, D097 and D098:
- 13292.8 (D102): its target's formula unresolved; its role `premise_backing`, with no question,
  since without a formula R2 asks for none.
- 13292.104#2: positive consequences. Guidelines §5 item 4 already decides it: practical reasoning
  only with an S2 distinct from the action, and "a proposal with a purpose is
  `positive_consequences`". Draft 2's note called it a question for the guidelines: wrong.
- 13292.192#2's note: `causes(E1,E2)` is a tendency, strong or weak (D036); what the attack lacks is
  the net-effect premise.
- R19's citation in the output format (ChatGPT's patch 02, its Go check), in the same revision of the
  guidelines as D098.
- D096: a reconstruction is not a claim an item can cite; its parts cite the claims that state them,
  or are supplied.

**The independent check of step 1 (7c8679d).** It found the Go agreement test's figures stale (fixed in
8431a50) and the replay's report not made again (made from a clean build, ebd2d72), and five defects or
gaps, each fixed with cases written first (8e900ec) and tests that fail without the fix:
- the import read an integer beyond a float's range, which Go's reader then refuses (V50); non-finite
  numbers in the verdicts and the rounds files had no case (V51, already refused);
- Python read a marker's reason with `str.strip`, which removes U+001C-U+001F where Go's
  `strings.TrimSpace` keeps them; `validate.py`, `agreement.py` and the replay now strip Go's white space
  (A5);
- the replay's guard knew only well-formed markers: a blank reason, an `unstated` marker or a list
  stopped the replay or was played (U7; U8 is U4's resolved control);
- D102's check stopped on a target that is not an object (A6), and the role counts and the pilot's
  role comparison read the role without D102 (A7);
- the Other label had no case for two spellings of one conclusion (T9).

**What the gold pins.** `validate.py` pins ten inputs, not the guidelines. A verdict still holds only
of the guidelines and `walton.yml` it was given on, since `verify_gold.py` compares every input but
the gold with the repository's (§5); a verification already written depends on the item's revision
and its claim's text alone (R17).

## 15. Round 1, draft 3 (2026-11-27)

Draft 3 takes in exchange 2 and the owner's decisions, part by part, each checked by `validate.py`
and recorded in the items it changes (`adjudication.recoded`) and in `d086-rounds.json`.

**Part 1: the causal exception names its effect (D071, D097; `recode_d097.py`).** Proposal
D040-D071-D097, version 2, written from the independent review of version 1 (which found D071, the same
decision of 2026-11-22, left out, and book words in the summary), is applied: `walton.yml`'s
`cause_to_effect` and `cause_to_effect_modal` have `interference(E1,E2)`, that another causal factor in
the case keeps, or would keep, E1 from bringing about E2; `cause_to_effect` cites its source (D040). In
the gold: the 17 causal instances' exceptions name their own cause and effect; the 6 challenges at
28 CQ3 name the new formula, and 13292.36's no longer cites 13292.24, a safeguard added to the policy
(an action amendment, OC6), as interference; 13292.182 and .198 answer 13292.180#2 at
`interference(reduced_meat_consumption,more_anemia_cases)`, which their Other instances conclude, so
an answer about anemia no longer undercuts 13292.192#1's step from the same cause to more crop
production (G2). The engine shows it (`causal_d097_test.go`): an answer about one effect undercuts that
step only, a second argument for the effect from another cause stands, and on the frozen format-1
`walton.yml` the one-place exception undercuts both. Guidelines §5 item 11 states the line between a
factor that keeps the cause from its effect (the exception), a safeguard added to the policy (OC6;
rule case OC6-d) and a factor that only weakens the effect (28 CQ1, as part 1 wrote it; corrected in part 3:
such a factor answers neither CQ1 nor the exception). Twelve items change: of the ten
round-1 claims that waited on G2, the seven agreed (13292.2, .9, .36, .141, .180, .182, .198) go back to
review at draft 3, with 13292.11, .12 and .192 (draft 2, not reviewed); 13292.4 and .14 are round 2's,
not yet drafted.

**Part 2: the argument the discussion needs (D096; `recode_d096.py`).** The format first
(`annotation-records.md` §13, rule R20, cases RC1-RC15 above, each written before its test and code):
an argument that no claim states as a whole, which claims answer, is an instance of the claim whose
statement it concludes, marked `"by": "the discussion"`, not primary, fit `reconstructed`, its premises
cited to the claims that state them or supplied. `validate.py` and the Go check hold it to that; the
`instances` field of agreement leaves it out; the replay plays it before the claims that answer it.
In the gold the environmental argument for the tax is the thesis's instance 13292.1#2: goal-based
practical reasoning, its circumstances cited to 13292.108 and its efficacy premise to 13292.14, the
rest supplied, with the four challenge records of the whole argument. 13292.108 is what it states, a
fragment backing those circumstances (`premise_backing`), as 13292.11 backs 13292.118's; the seven
claims that answer the argument's other parts (13292.5, .147 and .174 at its conclusion, 13292.13 at
its efficacy premise, 13292.6 at its goal link, 13292.104 and .127 at its exceptions) target
13292.1#2; the claims at the circumstances (13292.2, .4, .135, .137) keep their targets. 13292.104's
goal premise, cited to 13292.108, which does not state the goal, is supplied (R20). Thirteen items
change; the seven claims that waited on Q1 are released, and 13292.1, .5, .147 and .174, agreed at
draft 1, go back to review at draft 3, with .6, .104, .108, .123 and .127. The replay from a clean
build (6b44138) plays 13292.1#2 first, by Pro, with goal-based practical reasoning; it is out, for
13292.13's unanswered question about its efficacy premise, as 13292.108's argument was. 13292.108 is
played as a fragment (27 claims as intended, 4 fragments). The first replay of the recoded gold
showed 13292.135 and .137, which deny what 13292.108 states, challenged in the fragment's own Other
argument, where the formula is a conclusion: cases RC14-RC15 were written, and a move at the premise a
fragment backs is now made in the argument it backs, so the two are played as before (a challenge
and its duplicate).

**Part 3: mitigations and replies to a weighing (D098; `recode_d098.py`), and the review of part 1's
application.** Guidelines §3.1 states the convention both reviewers applied: a measure the agent adds
so that a cited harm does not follow is at the conclusion of the argument that cites the harm against
A, by its instance id (OC6; rule cases OC6-e and OC6-f), and what those affected would do is coded by
what it says of the harm; a denial of the link a weighing relies on is at the weighing's own premise
(OC10). 13292.45 moves to 13292.12#3's conclusion, as 13292.194 (a cascade in the replay now: 40 claims
played). The independent review of part 1's application (verdict: approve with changes) found the line
"a factor that only weakens the effect is 28 CQ1's" wrong: CQ1 asks about the generalization, and a
factor that only makes the effect smaller answers neither it nor the exception (at a cited
consequence, `too_slight`, OC9); the guidelines, the catalogue entries and the proposal (version 2b)
say so, and the old line is a retired wording. It also found that nothing tied the gold's English to
`walton.yml`'s templates: `validate.py`'s R21 and the proposal check do now (cases EN1-EN7, PC1-PC3),
and 13292.61#1's conclusion, in the game's phrasing, is re-rendered. The template's wording ("keeps, or
would keep") narrows the decided "interferes with": put to the owner as a choice of its own. 13292.12
and .194 (round 1) are at draft 3, not reviewed; 13292.45 and .61 are round 2's, and 13292.9, of round 2,
joined round 1 and is at draft 3, not reviewed.

**Part 4: exchange 2's objections and verdicts, and the review of D096's application
(`recode_d086_r1_d3.py`, 2026-11-28).** Exchange 2 agreed with 11 of its 13 items and objected to two,
both accepted. 13292.8 states a condition 13292.115's premise presupposes, excess consumption for the
tax to reduce, not the premise: its relation is unresolved by the owner's rule D102 (`premise_backing`,
its target's formula `{"unresolved": why}`, no question), so it is not played (39 claims played), not
scored as a relation and not verified. 13292.104#2 restated the action as its S2, which §5 item 4 does
not allow: it is positive consequences (`brings_about(A,G)` explicit, `good(G)` supplied), its
challenge records following (the feasibility record at the better means, whose premise it is).
13292.192's notes say that the modal form carries the source's "may" and that `causes(E1,E2)` is a
tendency of any strength (D036). `d086-rounds.json` records exchange 2's verdicts; the four agreed items
unchanged since (13292.25, .218, .43, .204) are agreed at that revision. The guidelines' output format
asks for `cite` on an inferred premise, and the Go check reads it (R19, cases CT1-CT9; ChatGPT's patch 02,
written again). The independent review of D096's application (approve with changes) found two latent
errors and gaps in R20's checks and tests: the replay plays a reconstruction for its own conclusion,
R20's citation and target rules have no gap, formulas are compared with Go's space set, and 13292.1#2's
alternatives are written for it (cases RC16-RC21). With 13292.104#2's practical reasoning, the target
both annotators gave 13292.208 is gone: 13292.104#2's circumstances,
`current_circumstances(farming_industry_not_meat_is_problem)`, which §10 left to round 2; round 2 finds
where 13292.208 answers (13292.104#1's better-means premise, where it is now, or another).

