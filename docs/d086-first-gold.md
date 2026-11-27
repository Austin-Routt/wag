# The first gold (D086): plan, version 2

Status: plan, 2026-11-27, written before any drafting; the developer's under the owner's instruction to
go with the recommended choices. Round 1's first draft followed the same day (§10). Phase 3's first item ([plan.md](plan.md) §4, item 1; D086, D045, D092).
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

The round is checked as the gold always is: `validate.py` (R2-R17) and `test_validate.py`; the Go
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
  item now, when the item is agreed in `d086-rounds.json` at the revision it has, and has no
  unresolved value (R17).
- It removes `verified` for a verdict of another choice that holds of a verified item now: the owner's
  later verdict takes the verification back.
- It lists every verdict it does not write, and why: the item changed since the verdict, or the gold it
  was given on is not in the repository; a claim not agreed in a round, or not in the gold; another
  choice on an item not verified; a verdict that recorded no gold.
- It writes nothing at all when any other input the verdicts recorded differs from the repository's
  now (the claims' texts, the guidelines, `walton.yml`, …: the round's clarifications of the guidelines
  are made before the owner's build, not between it and the script's run), or is one the built-in
  dataset does not have (a folder dataset's "guidelines.md (built in)", say); when the two copies of the
  gold differ; or when the file is not a list of verdicts, a verdict lacks its claim, choice, time or
  inputs, or two verdicts name one claim.

It writes `gold.json` and the built-in copy as the recoding scripts and `validate.py --repin` do
(`json.dumps(…, indent=1, ensure_ascii=False)`), changing only the `verified` fields, the file's
`status` (how many items are verified) and one `history` entry naming what was verified or taken back.
Its `item_rev` is a copy of `validate.py`'s (which runs its whole check when imported); the tests run
`validate.py` on the gold the script wrote, and R17 compares the two. Its test, `test_verify_gold.py`,
runs in CI beside `test_validate.py`.

The verdicts the script used are committed in the repository, `docs/kialo-batch2/verdicts/`, with
their notes (the plan's "the dashboard's verdicts are the record"; the owner is told that they become
public). A verified item that a later change touches is stale (R17) until the owner verifies it
again. R13 refuses a verified item whose target lost its formula, and R2 one whose question no longer
applies there; a target recoded with both kept leaves the verification standing, since what was
verified still holds of it.

## 6. Cases (outcomes written before the tests)

The script:

| Case | Setup | Expected |
|---|---|---|
| V1 | Verdicts *gold* on two claims agreed in a round, given on the gold as it is | Both items get `verified` (`by` owner, the verdict's UTC date, the item's revision and its text's hash now); `validate.py` passes (R17); the file is byte for byte the file before but for the two fields, the `status` and one `history` entry; the built-in copy equal to it (the test's own output) |
| V2 | A verdict given on an earlier gold, in which the item differs | Not written; listed: the item changed since |
| V3 | A verdict given on an earlier gold, in which the item is the same (another item differs) | Written, as V1 |
| V4 | A verdict given on a gold the repository does not hold | Not written; listed |
| V5 | Verdicts *none*, *unsure*, an annotator's name, *reviewer*, and an empty choice with a note, on an item not verified | Nothing written; each listed |
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

The dashboard:

| Case | Setup | Expected |
|---|---|---|
| D1 | A run made with a list of claims in the *Claims* field | The run holds those claims, in that order; an unknown claim is refused and no run is made; with the field empty, the first N as before |
| D2 | A claim of the built-in batch-2 dataset, the gold shown | Under the gold's card, the whole item: each instance's id and premises, the challenges, the evidence's links, the notes and the adjudication's reasons |
| D3 | A verdict sent with the input hashes of the page | Taken when they are the run's; refused, with nothing saved, when one differs; taken as before when none are sent |

Each rule of the script and each of the three changes to the dashboard gets a planted defect that one
case must catch.

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
   secondary instances?

Taken up in round 2, not in this draft: 13292.208's target, which format 1 also moved (to 13292.104#1's
better-means premise; both annotators put it at 13292.104#2's circumstances
`current_circumstances(farming_industry_not_meat_is_problem)`, which its note says it contradicts,
though the note's last sentence reads it at the moved target); 13292.121's role
kind `side_effect`, format 1's label (§5 item 3 makes it a value of `role.move`); 13292.14's capped
causal steps; and the notes of 13292.7, .115 and .121 that still say "until workstream E".
