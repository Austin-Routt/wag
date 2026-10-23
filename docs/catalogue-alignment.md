# Plan: complete coverage and alignment of the scheme catalogue

**Started 2026-10-22, while tier 1 was drafted; updated the same day after its independent check ([part 12](review-2026-10-22-part12.md)), and revised after the owner's review of the goal ([part 13](review-2026-10-22-part13.md)): the game's fields, links, feedback from play and the process (workstreams I to M).** This plan expands [plan §2](plan.md#2-phase-1-the-aligned-catalogue), the aligned catalogue, from the 33 tier-1 schemes to every scheme. It lists what is left and in which order, and it states when the catalogue is complete and aligned.

**The catalogue's own README has the current figures:** [`wag/data/catalogue/README.md`](../wag/data/catalogue/README.md). The figures below were measured on 2026-10-22, from the committed catalogue and the upstream collection.

## 1. What "complete" and "aligned" mean

**Complete.**
- Every scheme of the book's compendium has an entry, and so does every scheme of `walton.yml`.
- Every entry has been drafted and then verified by a person.
- Every critical question of the book is in an entry, and is decided:
  - it targets the formulas it puts in doubt;
  - or it is marked as attacking the inference;
  - or the formula it needs is recorded as missing.
- Every assumption and exception of `walton.yml` has a question that asks about it.

**Aligned.**
- **Book and `walton.yml`.** Where they differ, the entry records the difference and the decision. No decision is left open.
- **Catalogue and `walton.yml`.** The formulas are `walton.yml`'s. A formalization that the upstream collection has but `walton.yml` lost is either brought in or recorded as rejected.
- **Catalogue and the game.** The NASCQ (schemes, questions, ask and backup edges, and every field the game uses) and the game's `core.yml` are generated from the catalogue, and the game's tests pass unchanged (plan §6).

**Playable** (added 2026-10-22, the owner's goal: plan, "The goal"). Every scheme and critical question is playable in some context, as plan §6 defines it. The literature gives the first set; data and play refine and complete it. Completeness is measured, not proven:
- the share of arguments in new data that fit no playable scheme;
- the share of play reports that a move could not be made.
- **Catalogue and the other WAG data.**
  - `links.json` and `families.json` are generated from the catalogue, or retired.
  - The annotator's prompt and guidelines take their scheme descriptions from the catalogue (plan §3).

## 2. Where things stand (2026-10-22)

| Item | State |
|---|---|
| Entries | 116, one for each of `walton.yml`'s 116 schemes. All 109 compendium schemes are linked to an entry, and since 2026-10-22 every one has a formalization. |
| Tier 1 | 33 entries, drafted on 2026-10-22. The first 32 were checked by three independent model agents, whose 117 findings are applied ([part 12](review-2026-10-22-part12.md)); `cause_to_effect_modal`, added later the same day (C4), had a check of its own, whose 28 findings are applied (part 12 §4). Not yet verified by a person. |
| Tier 2 | 83 entries, generated only. |
| Book questions | All 153 printed questions are in some entry. 48 book-question items target no formula and are not marked as undercutting: 2 in tier 1 (established rule, with the missing exceptions recorded), 46 in tier 2. The checker lists them. |
| Questions for assumptions and exceptions | 10 assumptions and exceptions in 9 schemes had no question before tier 1 was drafted. Tier 1 adds six `walton/` questions: its five, and `legitimate_value`, whose book question turned out to ask something else. |
| Entries without any question | 47: the book prints none for them, and the game asks none. |
| Generated items still unverified | 763: book links, families, premise maps and question targets. 164 of them are in tier 1, where the drafters and the checkers found them right; the rest are tier 2's. |
| Discrepancies | 126. Tier 1 holds 115: 61 between the book and `walton.yml`, 43 other, 7 partial links, 3 game differences, 1 not in the book. 57 of tier 1's decisions are open, for the owner or a verifier ([part 12 §5](review-2026-10-22-part12.md#5-what-is-left-open)). |

### What the upstream collection holds that `walton.yml` lacks

The upstream collection, `walton-argumentation-schemes-for-carneades-4-master`, has the master `walton.yml` and 13 category files. Against the master:
- 106 category-file schemes are identical to it.
- 6 differ.
- 4 are only in the category files.

What matters for alignment:

| Finding | Where | What it means |
|---|---|---|
| A second `negative_consequences` block, formalizing compendium 33.3 (reasoning from negative consequences: an agent's act leads to an outcome bad for the agent's goals) | `A_reasoning/3_practical-reasoning.yml` | It is lost from the master, apparently because its id is a duplicate. **Brought in on 2026-10-22** as `negative_consequences_reasoning` (C1 below). |
| A second `practical_necessary_condition` block, which is the sufficient-condition schema | same file | The same duplicate-id defect. The master already has it as `practical_sufficient_condition`, so nothing is lost. |
| `whole_into_parts` | `A_reasoning/2_inductive-reasoning.yml` | A formalization of compendium 10.2, division of the whole into its parts, which `walton.yml` lacked. **Brought in on 2026-10-22** (C2 below). |
| `full_slippery_slope` | `A_reasoning/5_casual-reasoning.yml` | A formalization of compendium 53, the full slippery slope, which `walton.yml` lacked. It had no source, and two 0-ary atoms (`accept_all_steps`, `led_to_step`). **Brought in on 2026-10-22,** with the compendium's reference as its source and arguments for the two atoms (C2 below). |
| `opposition_descriptive_2`, `_3` | same file | The master's `opposition-descriptive_2`, `_3`, spelled with an underscore. Nothing is lost. |
| `commitment1` with `committed_to` | `B_source-based-arguments/2_…` | The master misspells it `commined_to`; the category file is right, and the cleaned `walton.yml` already has the correction (its CHANGELOG). |
| `part_in_whole` with variables `[Y,X,A]` | `A_reasoning/2_inductive-reasoning.yml` | The master declares `[Y,X,P]`, though its formulas use A and never P. The category file is right, and the cleaned `walton.yml` already has the correction. |

**The one compendium scheme with no formalization anywhere** was 50, the precedent slippery slope. It got a scheme written for WAG on 2026-10-22 (C3 below).

## 3. Workstreams

### A. A person verifies tier 1

**The work.** 33 entries, about 20 to 30 minutes each with the book: 10 to 15 hours. [`DRAFTING.md`](../wag/data/catalogue/DRAFTING.md) §"Verifying" says how.

**How:** by the rubric for each field, with about one item in ten verified again by someone else, and the agreement reported ([`process.md`](process.md) §2; the rubric is to be written).

**Before it starts:** the owner's decisions in workstream C that change tier-1 formulas (N12, and the scheme changes listed in [part 12 §5](review-2026-10-22-part12.md#5-what-is-left-open)). Otherwise the verifier would verify entries that are about to change.

### B. Draft tier 2

83 entries. Use the same pipeline as tier 1:
1. a source pack per scheme;
2. `DRAFTING.md`;
3. drafting agents by family, each in its own copy of the catalogue;
4. an independent model check against the book, the NASCQ and the upstream files;
5. a person's verification.

**Order, as plan §2 has it:**
1. the schemes annotators propose as near misses, and the "other" candidates;
2. then whole families, so that contrasts are written in both directions at once.

**Effort.**
- **Drafting:** four to six agent sessions.
- **The model check:** two sessions.
- **Verification by a person:** 30 to 40 hours.

### C. Decisions on the sources (the owner)

Each of these changes `walton.yml`, `links.json` or `families.json`. The catalogue records them as open discrepancies until they are decided.

1. **33.3: decided on 2026-10-22.** The upstream second `negative_consequences` block is in `walton.yml` as `negative_consequences_reasoning`:
   - variables filled in;
   - templates in our own words;
   - recorded in the cleaning's CHANGELOG under "Changes decided by the owner".

   `links.json` links 33.3 to it exactly, `families.json` names it, and the catalogue has it as a tier-1 entry.

   **Also decided on 2026-10-22,** from what drafting tier 1 found (the same CHANGELOG section):
   - the new scheme and the slippery slope's base case conclude `¬should_be_performed(A)`, so that they meet the consequence and practical-reasoning arguments about the same action. The new scheme had threat's `not_bring_about(A)`; the base case had `negative_consequences(A)`, which no other scheme uses;
   - `negative_value` concludes `reason_retracting_goal(V,G)`, as the book has it, and its `negative_affects/3` template is corrected. Both had been copied from `positive_value` and said the opposite; 19.2 now links exactly;
   - the three consequence schemes get the exception `other_consequences(A)`, which `negative_consequences_prudential` already had, for the book's question about opposite consequences (33/CQ3). The game's copies of positive and negative consequences lack it until the game's data is generated from the catalogue (E). The batch-2 gold gains the exception in its 16 consequence instances. Since replaced by D055 (2026-10-23): one shared exception let one argument defeat every consequence argument about the action, so each scheme now has an exception on its own side that names its consequence (`bad_outweighs(A,G)`, `good_outweighs(A,G)`, `good_outweighs(A,B)`, `bad_outweighs(X,C)`), and 33.4 concludes `should_be_performed(X)` (D059);
   - the slippery slope's source year is the book's 1992, not 1999. The book's ch. 4 cites the same page as 1999; a verifier with its reference list should confirm.

   The `other_consequences/1` template was rewritten with the exception, to say that the opposite consequences outweigh the cited ones. It is an orphan template since D055.
2. **10.2 and 53: decided on 2026-10-22.** `whole_into_parts` and `full_slippery_slope` are in `walton.yml` under their upstream ids, checked against the book first.
   - **Both** follow the book as printed. Their templates are in our own words.
   - **`full_slippery_slope`** has the compendium's reference as its source. Its two 0-ary atoms take the series' ends, `(C0,C2)`, because they are about the series at hand (the cleaning's rule on 0-ary atoms).
   - **`links.json`** links 10.2 partly: three parts stand for the book's n, and only the positive reading is formalized. It links 53 exactly. `families.json` names both, and the catalogue's two placeholder entries gave way to real ones.
   - **Left open,** in the entries when they are drafted (tier 2):
     - whether `full_slippery_slope`'s conclusion, `not_acceptable_step(C0)`, should meet the practical schemes, as the base case's now does;
     - whether its assumption `led_to_step`, which repeats the Group Opinion premise, should stay.
3. **50, the precedent slippery slope: decided on 2026-10-22.** `precedent_slippery_slope` is written for WAG. It passes Hansen's definition: sentential forms with variables, predicate constants, a conclusion.

   | Book | Formula |
   |---|---|
   | Claim to Exceptional Status Premise | `claimed_exception(C0,R)` |
   | Related Cases Premise, the first pair (CQ1) | `sets_precedent(C0,C1,R)` |
   | Related Cases Premise, the chain (CQ2) | `precedent_chain(C1,Cn,R)` |
   | Intolerable Outcome Premise (CQ3) | `intolerable_exception(Cn,R)` |
   | Conclusion | `¬case_exception(C0)` |

   - **The conclusion** negates the premise of the argument for an exceptional case (`exceptional`). The book calls the slope a counterargument to argument from precedent (p. 117), and this lets it meet that argument.
   - **One reading is ours.** The book prints C0 in the Intolerable Outcome Premise and CQ3 (the ch. 3 scan is unclear). The formula takes the end of the series, Cn, as a slippery slope needs. `links.json` therefore links 50 partly, and a verifier with the book should confirm the reading.
4. **N8, the modal cause to effect: decided on 2026-10-22.** `cause_to_effect_modal` is written for WAG, for causes that might occur, or have not yet (`docs/review-2026-10-13-part5.md` §7): `may_occur(E1)` and `causes(E1,E2)`, with the exception `interference(E1)`, conclude `may_occur(E2)`. Its source is the one the book gives for 28.
   - **`cause_to_effect` keeps the strong reading** (`has_occurred`, `will_occur`). `links.json` links 28 to it exactly, as both forms of 41 link exactly, and to the new scheme partly, because its causal premise is not weakened, and `families.json` makes the new scheme a variant in F13. Both are tier-1 entries; their contrast says when to use which.
   - **The batch-2 gold** still codes its 11 hypothetical causal steps as `cause_to_effect` with the N8 note, as its guidelines (item 11) ask. Recoding them is left to the owner; later batches use the new scheme.
   - **Left open,** in the entries:
     - whether `causes(E1,E2)` should be read as a tendency, strong or weak, in every scheme, or a weak link get a formula of its own (the book's weak reading weakens the generalization too);
     - whether a causal chain should be able to back a consequence scheme's `brings_about(A,G)`, which no step concludes;
     - whether a `cause_to_effect` step should chain into a modal one (`will_occur(E)` meets no `may_occur(E)`);
     - whether the consequence schemes' `brings_about` and the slope's links should take a modal form too.
5. **N12, analogy:** separate source and target propositions. **Done on 2026-10-22** (decisions D001, D028, D029; [proposal](proposals/D001-D028-D029.md), version 3, after two independent reviews and a short check): analogy and the practical analogies have a claim or action in each case, joined by a correspondence premise, with a directional difference exception; the practical analogies conclude `should_be_performed(X2)` and its negation. The game takes the new form at workstream E; until then the catalogue records that the game's copy differs. Two patterns that fit neither form are candidates ([`candidates.md`](candidates.md), D056).
6. **The upstream defects** (duplicate ids; the master's typo and wrong variable list, already corrected in the cleaned file): record only. `walton.yml` is already right.
7. **`walton.yml`'s English templates.** Rewrite in our own words the 173 that repeat the book (plan §7, item 1). The catalogue does not store them, but the game and `wag-cli --catalogue-show` render them.
8. **The scheme changes that drafting tier 1 proposed.** These are the 57 open decisions recorded in the tier-1 entries, grouped in [part 12 §5](review-2026-10-22-part12.md#5-what-is-left-open):
   - exceptions `walton.yml` lacks;
   - the burden of practical reasoning's questions;
   - formula shapes;
   - citations;
   - the game's texts.

   Those that change tier-1 formulas come before tier 1's verification (A).

### D. Decide every book question and cover every assumption and exception

- **Book questions with no target.** For each, choose one:
  - targets;
  - "attacks the inference";
  - or a `book-walton` discrepancy naming the missing formula.

  The checker warns until it is decided.
- **Uncovered assumptions and exceptions.** Give each a `walton/<predicate>` question. There are five left after tier 1, in tier-2 schemes. Tier 1 added six, including `legitimate_value`, once its book question proved to ask something else.
- **Uncovered premises: no new questions.** The game's challenge move lets anyone question any premise that no question covers (game rules §2).

### E. Generate the game's data from the catalogue

This is plan §6.

1. Write a generator from the entries to the NASCQ's `schemes.json`, `critical_questions.json` and `edges.json`, and to the game's `core.yml`:
   - the question texts from the paraphrases;
   - the loci from the targets: an assumption is asked at the premise, as now;
   - the patterns from the formulas;
   - **every other field from the entry's game fields (I),** and each question's NASCQ question from its choice (J);
   - `core.yml`'s formulas from the formula layer, so that the game no longer keeps its own copy.
2. Its first output must reproduce today's NASCQ and `core.yml` except where the catalogue records a reason. Known reasons:
   - two of expert opinion's questions narrow the book's;
   - the game's copy of `sign` lacks the exception `more_reliable`.
3. The game's tests must pass unchanged.

The generator can be written before tier 1 is verified: reproducing today's data is its test. The game's data is regenerated from verified entries before a release.

**Then** the game can take the other schemes, family by family, with their ask and backup edges linted as now and exercised by scripted play (L).

### F. Retire the drafted link files

Once an entry's book links and families are verified, the catalogue is their authority. Then:
- generate `links.json` and `families.json` from the catalogue, so that the registry and other readers still find them;
- or move those readers to the catalogue and delete the files.

**Until then,** a correction made in the catalogue is also made in `links.json` or `families.json`, so that the generator agrees. The checker reports the cases where they differ.

### G. Tools

- **Checker.**
  - Warn when a contrast has no contrast back in the other entry.
  - Warn when a question printed once under a heading (such as 33/CQ1–3) reads differently in the entries that share it.
  - Count the open ("Open:") discrepancy decisions in the progress line. Done on 2026-10-22, with the decision register (M).
- **A verification view in the dashboard.**
  - It shows an entry next to the book's text, which the verifier loads from a local file. The text is never stored or committed.
  - It has buttons to verify or dispute each item.
  - It writes the status and a history line.
- **Drafting helpers.**
  - Move the patch helper used for tier 1 into `wag-cli` (`--catalogue-apply`), so that people and models draft the same way.
  - Move the source-pack builder into `wag-cli` too. It reads the book's text from a local file given on the command line.
- **The annotator's prompt from the catalogue** (plan §3):
  - each scheme's summary, contrasts and near misses;
  - its questions in our own words.

### H. The book's chapters 4 to 8

Only chapter 9 (the compendium) and pp. 1–138 are transcribed, outside the repository. Chapters 4–8 (pp. 139–307) discuss:
- the causal schemes (ch. 5);
- the attack and rebuttal of schemes (ch. 6);
- their formal models (chs. 7–8).

Two ways to cover them:
- the person verifying reads them with the book open;
- or the owner transcribes them locally, outside the repository, so that the model check and the copy check can use them.

### I. The game's fields in the catalogue

Added 2026-10-22 ([part 13](review-2026-10-22-part13.md)). The NASCQ fields that the texts do not define get a place in each entry, with a status, a reason and example IDs. Plan §6 lists them in three kinds: derived by a rule, judged with examples, and calibrated by data and play.

1. **Extend the entry format** (`entry.go`) and the generator. The fields of the 9 game schemes are generated from today's NASCQ, with the status `generated`.
2. **Write the rules** for the derived fields, each with positive and negative examples run as tests, and recorded exceptions (such as analogy's).
3. **The checker re-derives** the derived fields and reports any difference that has no recorded reason.
4. **Draft the judged fields** for tier 1, citing coded examples. The calibrated ones start from today's values.

### J. Each book question names its NASCQ question

35 of the catalogue's 222 questions have a game part. For each of the others, choose an existing NASCQ question or a new one, by the criterion in plan §6, and record the choice with a status. This happens with the drafting, scheme by scheme. The checker reports two problems:
- a NASCQ question whose schemes' questions ask different things;
- a question with no choice.

### K. Links across schemes

1. **A shared vocabulary.** One file lists the predicates where schemes meet: `should_be_performed`, the causal and value predicates, and the others as found, each with what it means and which schemes use it.
2. **Rules for backup edges,** over what a scheme concludes and what a question needs, with the generic schemes that can back any statement. Hand-written edges keep a reason and examples.
3. **How well the schemes connect.** The checker reports how many premise-side formulas some scheme concludes, and how many schemes conclude something nothing uses (10 of 388, and 91 of 102, on 2026-10-22).
4. **Refined by data and play.** An answer in coded data that no edge allows, or a "could not make the move" report, is a candidate edge.

### L. Feedback from play

As [`process.md`](process.md) §4 describes:
1. Notes attach to catalogue IDs, in categories that lead to actions, and model players get the same channel.
2. `wag-cli --play-report` totals the notes, and counts plays, questions asked and answers that succeed, per catalogue item.
3. Scripted model play exercises every ask and backup edge of each playable scheme.
4. Human sessions follow a short rubric per move.
5. Games record the versions of the catalogue and the NASCQ they ran on.

### M. The process and its infrastructure

As [`process.md`](process.md) describes:
1. **The private workspace** with its session-start hook. Done on 2026-10-22: `github.com/Austin-Routt/wag-workspace`, restored by the hook in a test from GitHub.
2. **CI on every push.** Done on 2026-10-22 ([`.github/workflows/ci.yml`](../.github/workflows/ci.yml)).
3. **The decision register,** gathering every open decision, sorted by category and by what it blocks. Done on 2026-10-22: [`decisions.yml`](decisions.yml), with its view [`decisions.md`](decisions.md), kept in step with the catalogue by the checker.
4. **The generated status report.** Documents link to it instead of copying figures.
5. **The agents' protocol:**
   - briefs moved into WAG;
   - `--catalogue-pack` and `--catalogue-apply` (G);
   - the findings format validated;
   - the ledger of dispositions;
   - a `checked by` status;
   - measured reliability, by seeded faults and a person's sample.

## 4. Order

Revised on 2026-10-22 ([part 13](review-2026-10-22-part13.md)): the infrastructure first, so that no work is lost or repeated; then the game's fields, so that verification covers them too.

1. **Protect and organize the work** (M):
   - the workspace, CI and the decision register (done 2026-10-22);
   - the status report;
   - the briefs and pack builder moved into WAG.
2. **The owner sorts the register** ([`decisions.md`](decisions.md)): confirms the proposed priorities, then decides first what blocks the game or verification. C5 and C8, which change formulas, come before anyone verifies the entries they touch. C1 to C4 were decided on 2026-10-22; C6 and C7 can wait.
3. **The game's fields and question choices for the 9 game schemes** (I, J), and **the generator** (E): its first output reproduces today's NASCQ and `core.yml`, and the game's tests pass.
4. **The verification protocol,** then **a person verifies tier 1** (A), game fields included. Model checks are measured at the same time (M).
5. **Feedback from play** (L), then beta play on the game's schemes by people and models.
6. **Tier 1's other schemes into the game,** with links by rule (K). **Tier 2 in batches** (B), each with its questions decided and its assumptions and exceptions covered (D). Data and play refine the rules and propose candidates.
7. **Tools** (G) as the batches show the need. The verification view helps most before tier 2's verification starts.
8. **Retire the drafted link files** (F) when every entry's links are verified.
9. **Rewrite the English templates** (C7) before any public release of `walton.yml` with WAG.

## 5. Done when

- `wag-cli --catalogue-check --book …` shows:
  - no errors and no warnings;
  - every item verified;
  - no discrepancy decision beginning "Open:".
- The NASCQ and `core.yml` are generated from the catalogue, every game field included, and the game's tests pass.
- Every scheme is playable in at least one context, or the register says why not yet.
- Every ask and backup edge has been exercised by scripted play. The play report has no open "wrong" note, and its rates meet the targets the owner sets once the first report exists.
- The register has no open decision that blocks the game or verification.
- `links.json` and `families.json` are generated from the catalogue, or gone.
- Nothing the repository keeps reproduces the book: the catalogue's copy check passes, and `walton.yml`'s English is our own.

Completeness keeps being measured after that, as data and play add schemes (§1, "Playable").
