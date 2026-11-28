# The plan: one verified scheme catalogue, then annotation on it

Status: proposed on 2026-10-15. **Revised on 2026-11-22:** §0 gives the order of work from then on. **Phase 1 started on 2026-10-22:** the catalogue's skeleton and its checker exist ([`wag/data/catalogue/`](../wag/data/catalogue/README.md)); tier 1 is drafted and checked by models, and nothing is verified by a person yet. **Revised on 2026-10-22** after the owner's review of the goal ([part 13](review-2026-10-22-part13.md)). The owner's open questions are in §7.

**The goal,** set by the owner on 2026-10-22:
- WAG is to become a complete dialogue system, starting with argumentation-based dialogue.
- Every argumentation scheme and critical question is to be playable in some game context. The context (the dialogue type, the preset, and the NASCQ's fields and links) decides what a scheme can do there (§6).
- The literature gives the first playable set and its families. Kialo and other data, and play by people and models, refine and complete it.
- Completeness cannot be proven, only measured as it approaches. Four measures show it:
  - the share of arguments in new data that fit no playable scheme;
  - the share of play reports that say a move could not be made (§5, [`process.md`](process.md) §4);
  - the share of the answers to critical questions found in coded data and in play that can be played as meant (added 2026-11-22, D082);
  - the share of answers that should be refused (irrelevant, wrongly bound, unsupported) that the game refuses (D082).
- Within a versioned set of schemes and questions, completeness is checked, not measured: every question can be asked, and every answer the set offers has a tested contract (D082).

**Related documents:**
- **Why this plan,** from the first real model run: [`review-2026-10-15-part7.md`](review-2026-10-15-part7.md).
- **What Walton and Macagno's classification adds:** [`review-2026-10-15-part8.md`](review-2026-10-15-part8.md).
- **Dialogue types,** the second objective, are Track D (§8). It rests on the reference [`dialogue-types.md`](dialogue-types.md) and the review [`review-2026-10-15-part9.md`](review-2026-10-15-part9.md).
- **How the work is done:** statuses, agents, feedback from play, the decision register, the private workspace, the checks: [`process.md`](process.md).
- **Where the project stands:** [`README.md`](../README.md).

## 0. From 2026-11-22: the order of work

After the external reviews of 22 November 2026 (kept whole in the workspace, `reviews/2026-11-22`; checked in `verify/r-reviews-2026-11-22`, `verify/r-convergence-2026-11-22` and `verify/r-stage-review-2026-11-22`) the owner set the direction, and a standing rule for decisions ([`process.md`](process.md) §5). Where this section and the older ones differ, this section decides the order; the principles of §1 stand.

- **Data-driven development with model and human players, now.** API players and Kialo annotation produce the data that improves the game, the catalogue and a reusable dataset. Work may go on from entries models have checked, its results marked provisional; a person still verifies, and no accuracy claim or release rests on entries only models have checked (D045).
- **Full NASCQ** is a checked contract within a versioned set and four measured rates beyond it (the goal, above; D082).
- **The first gold** (D086, the owner, 2026-11-24), from batch 2's claims (§4): the developer drafts it, the other independent reviewer (ChatGPT) reviews it, back and forth until they converge, and the owner verifies it (D045). Until then it is development material, as batch 2's reference made by models is.
- **Then a first persuasion protocol** (D087, the owner, 2026-11-24): once the first gold and the bounded pilot (stage 3) have shown where WAG falls short, a small first persuasion protocol is the next experiment while the corpus grows (§8, Order).
- **Families as slices.** Each family goes through the whole path before the next: its decisions and formulas, the game's questions and answers (answer schemes, D075), annotation, replay and play. Within a slice the decisions come first.
- **Evidence** is a record, not a string (D076); the evidence cited in the data is captured whole for research, under the fair use doctrine, and kept in the private workspace, large files in Git LFS (D077). Terms of use, privacy and injected instructions are reviewed later (D079).
- **The engine** is the schemes library's ([`CARNEADES_FORK.md`](../CARNEADES_FORK.md)). The formulas stay WAG's own layer (§1); where WAG has decided nothing, the library's records are the draft (tier 2), and every difference is recorded.

| Stage | What | State |
|---|---|---|
| 0 | The decisions of 2026-11-22 in the register (63 decided, 8 open; D084, the reserved name `is` the engine swap brought to light, decided), this section, `process.md`. After ChatGPT's review of stages 0 and 1 ([what was adopted](review-2026-11-22-stages.md)): the owner's D078 (history kept), D044 and D080 (licences now); D081 decided; ten refinements; 6 open, four of them questions of fact for a person | done 2026-11-22 |
| 1 | The engine swap with WAG's fixes and the engine's tests in CI; indexed templates; annotation runs that keep their identity; sources that must be text. Closed after ChatGPT's review of stages 0 and 1: CI on both Go versions, a registry gate that fails closed; D084 and D081 applied through proposal [D081-D084](proposals/D081-D084.md); annotation identity at every entry point; a malformed final move refused, never replaced; typed answers instead of free-text bridges; MPL-2.0 and CC BY 4.0 | done 2026-11-22 |
| 2 | The first slice, practical reasoning and the consequences: D063 (with positive binding), D070 and the answer schemes (D075); the scheme fixes decided on 2026-11-22, through proposals and `edits.py`; one versioned game profile, its catalogue, formulas and engine pinned by content; the replay as a tested command (D073); one versioned record of a reading, a move and its outcome; evidence records and the first admission policy (D076, [evidence-records.md](evidence-records.md)); the [evidence implementation plan](evidence-implementation-plan.md), whose offline record checker and step 1, the records bound to the game (offers, consents, admissions and citations as moves), are built ([e1-evidence.md](e1-evidence.md)), with its owner-flagged choices put to the owner; API runs' budgets shown on mock clients first: no limit by default, optional limits, and a record of what each run used (D046); book text out of the public tree (D043) | in progress |
| 3 | A bounded Kialo and API pilot on the owner's machine, its spending recorded, with the limits the owner sets, if any (D046): independent first passes, a lexicon not drawn from the gold, the moves players could not make recorded, development and held-out discussions kept apart, the size planned from the gain to detect | after stage 2 |
| 4 | A small first persuasion protocol, scoped by what the first gold and the pilot show (D087, §8); the other families; tier 2 from the library's records; a person's blind sample of claims; AS2 and meta-argument trials, kept only if they do better than a simpler baseline at the same cost | later |

**Stage 2 is done when**, for its slice: no known contract violation; no input changed without its record (data, prompt, profile, model); every recorded event replays without asking a model again; the wrong actor, context, polarity and evidence revision are each refused by a test; what is unresolved or unavailable is stated, not guessed; and the sources and the code are reviewed at the exact versions used. These are finite checks; they do not claim that every argument people make can be expressed (after ChatGPT's review of stages 0 and 1).

**Stage 2 so far:**
- 2026-11-22: run budgets (D046). `wag-llm play` and `annotate` take limits on calls, tokens and cost, count retries, and refuse a call before it is sent if its estimate would pass a limit ([`llm.md`](llm.md#a-runs-budget-d046)). Tested on mock clients.
- 2026-11-23: the budgets hardened (the evidence-theory review's patch 04), and D046 amended by the owner: no limit by default, `--unlimited` to say so explicitly, `--max-calls`, `--max-tokens` or `--max-cost` when the owner wants one. Only the call cap is exact; token and cost limits are estimates, not billing ceilings. Every invocation counts its calls, the usage providers report and its price-table cost, prints them, and appends its limits, or none, with what it used, to `budget.jsonl` in its output folder; a resumed annotation run is a new invocation, with its own allowance and line. Malformed prices or usage cannot create credit, reasoning reported apart from the completion (xAI) is counted rather than refused, and an overrun stops later attempts while the trace keeps the response. Tested on mock clients and local test servers. The dashboard does not take a budget yet.
- 2026-11-24: the critical review's integrity repairs, with their follow-ups. A refused moderator decision leaves no trace. Datasets, annotations, model reviews and verdicts are read strictly. A load replays a game with the weights it recorded, refuses a game played with other scheme data or opens it read-only, and reads CRLF data as LF; `wag-llm replay` finds calls by player ID and uses the presets its trace recorded. Annotation runs record what they receive, or stop: a record that cannot be written stops new calls and is shown with what the files lack, it is written before any new call, and a reply that reached `calls.jsonl` is taken from there, not asked again ([`llm.md`](llm.md)). The dashboard's autosave of games with model seats, which only printed a failure, now shows it and holds the model seats until the game is saved; files are synced with their folders.
- 2026-11-24: the identity contract of annotation records, before the first gold (D092, [`annotation-records.md`](annotation-records.md)). Format 2 gives every argument instance an id that edits do not change; targets name the instance and the formula itself; a case record gives the decision context and maps each action to the one actor who would perform it, with its aliases; not applicable, unknown, absent and false are kept apart, and re-parenting is derived, not stored. The batch-2 gold is migrated to it, with D039's recoding of its eleven hypothetical causal steps (the version before kept). `validate.py` (R12–R17) refuses the wrong actor, the wrong context, an inexact or dangling target, a malformed record and a verification of another revision or another text of the claim; the Go check reports the first three (the actor and the context where the dataset has a case record) and part of the field contract on each annotator's annotations, in the claim's view and the run's report, and refuses nothing. Agreement adds the question, the exact target, the bindings and every instance, and leaves unresolved values out. An independent review found seven errors and eleven warnings, all fixed or recorded the same day; a check of the fixes found three errors and six warnings more, also fixed (`wag-workspace verify/identity-2026-11-24/`).
- 2026-11-26: the game's data from the catalogue (step C5.1, [`c5-game.md`](c5-game.md)). Each game scheme's entry records its NASCQ record and its `core.yml` block (with formulas only where the game's differ from `walton.yml`'s), and the catalogue's `game/` the shared questions, the order of the schemes and edges, and `core.yml`'s template. `wag-cli --catalogue-gen` generates the NASCQ files and `core.yml` from them, with `walton.yml`'s formulas, before making the entries, and the check requires the files to be its output. The values did not change: a test reproduces the files frozen before the switch as trees, and every answer to every question and the 33 reference games play the same. The bytes did, once: the data's fingerprint is now `afd3335a23f12f41`; the old data is kept in `wag/data/frozen/e7141ec8dea379bc` for the games and traces played with it; the batch-2 gold is re-pinned with a history entry; and a game loaded from an export that names no data stays a legacy record (the critical review's adverse case 11).
- 2026-11-27: the game takes the first family (step C5.2's switch, [`c5-game.md`](c5-game.md)). It plays 17 schemes and 34 questions: the four answer schemes (D075, D095), the modal variants of the consequences and their two bridges join the nine, and the consequences' exceptions of how serious and on whom, the outweighing exceptions and the bridges' exception are asked. The catalogue generates them by rules that refuse a game part breaking them: one id is one question (one formula, one kind of locus, one text, made from its paraphrase); an answer scheme concludes what it backs; a modal variant and a bridge weigh as a game scheme; the game's sentences are `walton.yml`'s, each difference recorded with its reason. A new game has a profile (D094 (b), since step 2). The data's fingerprint changed once, to `092b6a6f48a672d2`; the data before is kept in `wag/data/frozen/afd3335a23f12f41`, which the game's mechanics tests and the 33 reference games play. The batch-2 gold names the game's questions where the game now asks them (`recode_c52.py`): six roles and 21 challenge records; `validate.py` checks a challenge's game question as it checks a role's, a null where the game asks one included, and an instance's game scheme. The same day, item F ([`c5-item-f.md`](c5-item-f.md)): each question and answer route of the new data with the outcome a move must have, written and committed before its tests ran, then played on the data the game ships (and the schemes it does not play yet on the engine and in `validate.py`), each paired with a planted defect the tests catch. Its first runs found a value that names nothing read as a name (`x_empty`, or `x_` and a hash), which let an answer stand without naming what outweighs or what would settle the question; S2, P1, the profile's check and the adapter now refuse it, for Other's free text too, and the terminal asks again (and resigns for a seat whose input has ended). The new golden set plays the reference games' table and scripted games that reach each new question on the new data with a profile, its files pinning the data and the profile; the 33 games on the frozen data stay. The same day, G ([`g-journal.md`](g-journal.md)): a model's paid reply is saved before it is played, so that a game that stops, or cannot be saved, after the model answered does not ask (and pay) for the reply again. A game in the data folder is resumed from its folder with its trace (*Resume* in the browser); the trace must be the export's and must start as this server would start the game, a save that stopped between the trace and the export is mended, the model seat's turn in progress goes on from its record, and computer players draw again what they drew, so that the resumed game replays (`wag-llm replay`). Its cases and planted defects were written before the tests ran, and amended, before they ran, where the code showed the plan wrong. An independent check then found a held move taken for a refused reply, a resume losing a held move, and a second resume counting the stop as played, besides warnings; they were corrected the same day, with cases written before they ran (a move held for the moderator ends a turn and is held again after a resume; the export records the time played; no retry is sent and no move clock runs while a game cannot be saved; a game whose time is up asks no one; the build is recorded, not compared).
- 2026-11-27: the first gold begins (D086, [`d086-first-gold.md`](d086-first-gold.md)). Its plan, checked independently and revised (four errors, seven warnings), chooses round 1 by disagreement among annotator A, annotator B and the reference: 28 claims, three of them controls. The owner verifies in the dashboard, which now shows the whole gold item, runs the claims listed and refuses a verdict given on a stale page; `verify_gold.py` writes the verifications from the verdicts, item by item and only for items the developer and the other reviewer agreed on. Round 1's first draft is in the gold: four targets name the causal step they answer and a mitigation the argument it amends, three fragments leave unstated what they do not state, three items whose records named the moved claims join the round, and the rest keep the reference's coding with the reason recorded. An independent check of its first version found three errors and five warnings (a miscount, a record missing a field, stale records in three items; a mechanism read as data and a fifth moved target among the warnings); before any review they were corrected, or, for the fifth target (13292.208, round 2's), recorded for its round; a check of the corrections found three errors and two warnings more (a premise note left false by a move, readings the plan did not list as differing, a miscount; a reply whose history was not put to the review, causal steps called proposed where their cause is a predicted effect), also corrected. It goes to the other reviewer next.
- 2026-11-27: the first exchange of round 1's review (D086, [`d086-first-gold.md`](d086-first-gold.md) §11). The other reviewer agreed with 21 of the 31 items and objected to 10; draft 2 takes in the six objections accepted, the three partly accepted (two on the conditions of the reviewer's reconciliation), and two refinements, and 13292.108 goes to the owner (question 1). The rules the review found broken are checked (R18, explicit fit; R19, cited inferences). The owner's verification import keeps every verdict's time, so an older verdict cannot undo a newer one and verdicts of different choices at one time leave the item unverified, and it waits for the owner's decisions where a claim's coding depends on them (Q1, the two layers of 13292.108; G2, the causal exception's scope: 17 claims). The replay plays a claim at an argument's inference as an answer to its question or an undercut of it, and Other's counter-question names its conclusion. An independent check of the step ([`d086-first-gold.md`](d086-first-gold.md) §12) found the import's date-only guard lifted by a receipt that did not record the verification, a newer verdict whose gold was not found leaving no receipt, and two agreed items whose records draft 2 made false; each was verified and fixed, its cases written first (V35-V46), and 27 planted defects are caught. The owner then took every recommendation of the decision batch (D096-D101: two layers for 13292.108; the causal exception names its effect; the OC6 and OC10 convention in the guidelines; C5's questions; E1's choices; EthiX first, with no model run until a budget is set), and the second exchange went to the other reviewer in one package. Next: draft 3, from D096-D098, then the owner's timed verification session.
- 2026-11-28: step C5.3 (D099, [`c5-game.md`](c5-game.md)), the owner's answers to C5's questions. The code: P5 leaves sign's answer at practical reasoning's side-effect question out of the first profile until its warrant is reviewed; D095 (3)'s list (the answer scheme and the source-based schemes) bounds every route to the bridges' exception, in NASCQ and in the game; P6 has an undercutter of a bridge declare its ground; the board and the computer players are offered only the moves the game accepts. The switch (the game's data `a0352a9944190d8b`, the old data frozen): negative consequences answers CQ:side_effect, now a question of semantic type action (sign's edge to it, which then crosses types, marked `across: D009`); the bridges' question ends "If so, how?". Item F's cases A7 to A9 were written before their tests, each paired with a planted defect; the golden games were made again, every change traced to an offer added or removed; the batch-2 replay's moves did not change.

**Evidence increment (literature review of 2026-11-22).** [The theory review](evidence-theory-review.md) supports the record model and refines its semantics: source reports, interpretation/warrant, scope, dependence, assessment and procedural admission stay distinct. `wag-evidence/0.1` now checks offline records, retained bytes, selectors and two-party consent for consideration. Its pass does not verify facts, implement game admission, or close stage 2. Within stage 2, and so before the pilot, [step 1 of the evidence plan](evidence-implementation-plan.md#step-1-connect-the-first-slice-before-an-evidence-aware-game-pilot) binds records to the actual game, keeps old citation-only traces, and tests the specific CQ, admission and revision effects; this is the evidence item of stage 2 above, as the review recommended, not a new condition. The owner's existing capture, history and human-verification decisions remain unchanged; no new capture gate is introduced. **Step 1 is built** ([e1-evidence.md](e1-evidence.md), 2026-11-27): the inquiry mode, record sets, offers, consents and withdrawals as moves, citations of admitted bindings, model seats, replay, load and resume, each case written before its test and its rules checked by planted defects; its owner-flagged choices (1, 2, 4, 7, 8, 13, 14) are put to the owner.

**The critical review of 2026-11-23** (in the workspace, `reviews/2026-11-23`; checked in `verify/r-critical-review-2026-11-23`) proposed an execution sheet for the next increment: ten work items (A to J) and thirteen adverse cases. It refines the checks of stages 2 and 3 without changing the register or the order; its answer 2, on Track D's order, became the owner's D087. A (integrity repairs) and B (recording of annotation runs) were done on 2026-11-24, and D (admission of authored graphs: [`CARNEADES_FORK.md`](../CARNEADES_FORK.md), *Admission*) the same day. Still to do in stage 2 from it: the first family's formula closure, of which practical reasoning (C2) was applied on 2026-11-24 by proposal D004-D075, which states the agent P in the two condition schemes and removes their A2, and the consequences (C3) on 2026-11-25 by proposal D008-D074, which adds the modal variants, their bridges and the argument from what happens anyway, and the exceptions of how serious, on whom and whether acting averts the harm, and makes the appeal to consequences an undercutter of the argument it answers, and the rest (C4) on 2026-11-26 by proposal D022-D093, which makes the recommendations to the one addressed (fear, danger, threat and their kin) meet the other practical arguments and carries the conclusions that name an agent to them through bridges ([first-family.md](first-family.md); `ad_hominem`'s S and `regulative_rule`'s A go with their families), and the annotation records' identity contract on 2026-11-24 (D092); evidence E1; the CQ answer contract (done 2026-11-27, as step C5.2's item F); and D073's journal and replay (the journal of a model's paid replies done 2026-11-27, [`g-journal.md`](g-journal.md)). Remote play authorizes by role since 2026-11-24: another computer acts in a game only with a key its owner shared (a seat's, or the moderator's).

## 1. Principles

- **One catalogue, with an authority for each part.** The owner's decisions of 2026-10-15 and 2026-10-22:
  - The book (Walton, Reed & Macagno, *Argumentation Schemes*, 2008) decides what a scheme says, and its critical questions.
  - **The formulas are WAG's own layer,** `walton.yml`. It starts from the upstream Carneades collection, which is its origin, not its authority. Every change is recorded (the cleaning's CHANGELOG) and verified like any other item. Until 2026-10-22 this plan called the upstream collection human-verified and the authority; drafting tier 1 found defects in it, and five schemes have been written for WAG since.
  - **The game's fields and links** (§2, §6) are decided by logical coherence, consistency and examples of use. Data and play then evaluate, refine and validate them.
  - The game's NASCQ is generated from the catalogue: a selection with game rules added, not a separate source.
  - Schemes found in Kialo data or in play join as *derived* schemes, with their evidence.
- **Every field says who stands behind it:**
  - *drafted* (by a person or a model);
  - *checked* (by the validators and by models, named);
  - *verified* (by a person, with a date);
  - *disputed* (with the reason).

  How far a model's check counts depends on the field's risk and on the measured reliability of the checks ([`process.md`](process.md) §2–3). Where the book and `walton.yml` disagree (for example the book's modal cause to effect, audit N8, which the owner settled on 2026-10-22 with a second scheme), the entry records both and the decision.
- **No book text in the public repository.**
  - The catalogue holds our own paraphrases, the book's labels (scheme number, CQ number) and page references.
  - The owner's copies of the book texts are kept in the private workspace, `wag-workspace`, for verification and for the copy check ([`process.md`](process.md) §6).
  - This is how `links.json` works already.
- **One source of truth for each kind of fact.** Documents link to a generated status report instead of copying figures ([`process.md`](process.md) §1, §7).
- **No new scheme for a one-off pattern.** A derived scheme needs recurring evidence and a person's review.
- **Measure before and after each change,** on the same claims and the same model (§5).

## 2. Phase 1: the aligned catalogue

### What an entry holds

One file per scheme, for example `catalogue/schemes/negative_consequences.yml`, so that a review is a readable diff.

| Field | Source | Done by |
|---|---|---|
| id, name, book number(s) and pages, family | `walton.yml`, `links.json`, `families.json` | generated |
| summary: when the scheme applies, in two sentences of our own | new | drafted, then verified |
| variables, premises, assumptions, exceptions, conclusion: formulas. English readings are rendered from `walton.yml` when needed, not stored: many of its templates repeat the book (§7, item 1). | `walton.yml` | generated |
| map from the book's premises to the formulas | `links.json` premise maps | generated, then verified |
| critical questions: book label, our paraphrase, locus (premise, assumption, exception, conclusion, argument), the formula it is about, the NASCQ question it becomes (an existing one, or a new one), what counts as an answer | `links.json` CQ maps (197), the NASCQ (24), new paraphrases | drafted, then verified |
| the game's fields: for the scheme, its kind of claim, whether it can back any statement, its base weight, evidence policy, cues, variable types and English; for each question, what each answer does; the backup edges. Each is derived by a rule, judged with examples, or calibrated by data and play (§6) | the NASCQ (17 schemes), the rules, the examples | drafted, then verified; data and play refine them |
| contrasts: "this, not X, when …", for each close relative | the registry's families and signatures, new text | drafted, then verified |
| examples: two or three short claims coded in full (Kialo claims with their ids where possible), and one near miss | new | drafted, then verified |
| discrepancies between book and `walton.yml`, and the decision | the audit, `links.json` partial links | drafted, then decided |
| place in Walton and Macagno's classification: source-dependent or not; practical or epistemic; discovering a rule or applying one | Walton & Macagno 2015 (part 8 §1.7) | drafted, then verified |
| "specialises": the scheme this one is a domain version of, if any | new | drafted, then verified |
| proper in: the dialogue types where the scheme is a legitimate move (a threat is, in a negotiation, and is not, in a persuasion). It is the evidence for the weights each context gives the scheme (§6), not a mechanism of its own | the book, Walton's pragmatic theory of fallacy | drafted, then verified |
| status and history of each field | | |

### Where the sources stand today

| Source | What it has | State |
|---|---|---|
| The book's compendium (ch. 9) | 109 schemes (128 numbered entries, 19 of them headings only), 153 critical questions | Transcribed outside the repository |
| `docs/scheme-registry/walton-clean/walton.yml` | 130 schemes with formulas and English (111 upstream; compendium 33.3's, 10.2's and 53's, brought in from the upstream category files, and 50's and the modal form of 28, written for WAG, on 2026-10-22; three answer schemes written for WAG on 2026-11-24; five consequence schemes written for WAG on 2026-11-25, D008 and D072; five schemes of the first family on 2026-11-26, D022, D031 and D063; the answer scheme for the bridges' exception on 2026-11-26, D095); 70 encode some CQs as assumptions or exceptions (60 of the upstream 111) | Cleaned so that Carneades loads it; since 2026-10-22 WAG's formula layer, changed only by the owner's recorded decisions |
| `docs/scheme-registry/links.json` | 83 compendium entries linked exactly to `walton.yml` and 26 partly (120 links: 84 exact, 36 partial; recounted on 2026-11-26, after proposal D022-D093, whose restated conclusions are partial), none without a link; 212 CQ links, 161 of them naming a formula | Drafted by a model, not verified. All its formulas check against `walton.yml` (the catalogue's checker). |
| `wag/data/nascq/*.json` | 17 schemes, 34 CQs, 55 ask edges, 9 backup edges (since step C5.3's switch, 2026-11-28: negative consequences at CQ:side_effect; 8 from step C5.2, 2026-11-27; 9, 24, 35 and 4 before) | The game's data, generated from the catalogue, tested |
| `wag/data/schemes/families.json` | 21 families | Drafted |

Only one `walton.yml` scheme has no compendium entry: `defeasible_modus_ponens`. An earlier version of this plan counted six, from an older list. `links.json` links the other five:
- `epistemic_ignorance` and `exception_to_rule`, partly;
- `exceptional`, `popular_practice` and `threat`, exactly.

### Order of work

1. **Generate the skeleton of all entries** from the sources above, and a checker:
   - every formula is `walton.yml`'s;
   - every CQ's formula is one of its scheme's;
   - every game CQ appears;
   - no paraphrase reproduces the book's sentence.

   **Done on 2026-10-22.** [`wag/data/catalogue/`](../wag/data/catalogue/README.md) holds 114 entries and an index (116 since compendium 33.3 and the modal cause to effect got schemes of their own, the same day); `wag-cli --catalogue-gen` and `--catalogue-check` make and check them. The README there lists every check and what the skeleton found.
2. **Tier 1: 33 schemes** (first estimated at about 25; 31 until 33.3's scheme joined the consequences, and 32 until the modal cause to effect joined the causes, both on 2026-10-22), listed in the catalogue's README. These are the 13 schemes of the batch-2 gold and the close relatives needed for contrasts:
   - **practical reasoning:** practical reasoning 1 and 2, practical goal, end and means, necessary condition;
   - **consequences and values:** positive and negative consequences, reasoning from negative consequences, positive and negative value;
   - **causes and signs:** cause to effect and its modal form, correlation to cause, sign;
   - **cases and analogy:** analogy, the practical analogies, example;
   - **sources:** expert opinion, position to know, witness testimony, popular opinion;
   - **classification and rules:** verbal classification, definition to verbal classification, rule, established rule;
   - **alternatives and costs:** the alternatives, sunk costs and waste;
   - **slippery slope.**

   A model drafts every field; a person verifies them against the book.

   **Drafted and checked on 2026-10-22.** Four drafting agents followed [`DRAFTING.md`](../wag/data/catalogue/DRAFTING.md); three independent agents checked the drafts and the generated parts against the book, the NASCQ and the upstream files, and their 117 findings are applied ([part 12](review-2026-10-22-part12.md)). What remains, to complete coverage and alignment, is in [`catalogue-alignment.md`](catalogue-alignment.md).
3. **Tier 2: the rest,** in the order the data asks for them: the schemes that annotators propose as near misses, and the "other" candidates.

**The index's top level is Walton and Macagno's tree.** Our families sit under it, and the few schemes it leaves out (some linguistic ones) are placed by hand. **`defeasible_modus_ponens`** is in tier 1 as the general form "other" falls back to (§3). **A derived scheme must pass Hansen's definition:** a pattern of sentential forms with variables, at least one schematic constant, and a conclusion.

**Effort.**
- **Drafting:** a few working sessions with a model.
- **Verification:** about 20–30 minutes a scheme for a person with the book. That is 10–15 hours for tier 1, and 40–55 hours for everything.

**Done when:**
- every tier-1 entry is verified;
- the checker passes;
- the NASCQ is generated from the catalogue and the game's tests pass unchanged.

## 3. Phase 2: annotation, version 4

### The guidelines: one document, a procedure

The current guidelines are version 2 with the version 3 changes appended as a list of edits. They allow two answers in places and contradict each other in one. Version 4 is one document, in the order an annotator decides:
0. **Once per discussion, by a person:**
   - the setting;
   - the dialogue's dimensions ([`dialogue-types.md`](dialogue-types.md) §2.7): its aim (belief, action or relationship), its situation (conflict, open problem, asymmetric information), and any third party (audience or adjudicator);
   - the dialogue type, both supposed and enacted (§9.3). A Kialo discussion is public persuasion addressed to readers; on a practical thesis, it is persuasion about whether an action is advisable (ND p. 229), using deliberation's arguments but deciding nothing jointly;
   - the thesis's kind: practical, factual, evaluative or classificatory.
1. **Does the claim give a reason?** If not, what is it?
   - a thesis or proposal: record its conclusion form and scheme family;
   - a bare premise of another argument: a fragment of that argument;
   - a clarification, a question, a meta-comment.
2. **What does it target?** The parent's conclusion, a premise, the inference, or another claim (re-parented).
3. **Which family, then which scheme?** Walk Walton and Macagno's distinctions first:
   - Does the argument rest on a source?
   - If not, is it about what to do, or about what is so?
   - If about what is so, does it find a rule or entity, or apply one to a case?

   Then use the catalogue's contrasts. Choose the primary instance by a stated rule.
4. **What does the text state, and what is supplied?** This decides the fit.
5. **Role, polarity, critical question.** Which CQ comes from the catalogue, by its label.
6. **Evidence.**

Each step has examples from the catalogue. Where two answers are equally right, the guidelines say so, and the gold records both (§5).

### "None", the thesis, and "Other"

- **"None" is split into separate codes:**
  - *no reason given*: a thesis, proposal or bare assertion; it gets its conclusion formula and scheme family;
  - *fragment*: a stated premise of another argument;
  - *not an argument*: a clarification, question or meta-comment.
- **The thesis gets an explicit implied argument.** "The practical reasoning for the thesis" is a node that bare pros complete and that claims like 13292.13 can target. Nothing unstated is filled in.
- **"Other" must justify itself.**
  - It is coded as a defeasible modus ponens with its rule written out ("as a rule, if P then Q"). It then has two critical questions: what backs the rule, and is this case an exception.
  - It names the two or three closest catalogue schemes and why each fails, and proposes:
    - a pattern with variables;
    - a conclusion form;
    - the critical questions the pattern would face.

  Proposals go to a candidates file. A candidate seen in at least three claims of at least two discussions goes to a person, who decides one of four outcomes:
  - reject it;
  - merge it into an existing scheme;
  - record it as a domain version of one;
  - add it as a derived scheme.

### The prompt and the answer

- **Two steps per claim, so a call is not 41,000 tokens:**
  1. The claim, its context, and the catalogue's index (one line per scheme: name, conclusion form, a cue). The model names up to five candidates, or "other".
  2. The full entries of those candidates (CQs, contrasts, examples). The model codes the claim.

  Stage 1's recall is measured on its own. A single large prompt (the whole catalogue, cached) is the comparison, not the default.
- **The model chooses; the code writes.**
  - The answer gives the scheme, the variable bindings, which premises the text states, the target and the role.
  - The code builds the formulas from `walton.yml` and checks the answer before saving it.
- **Constrained output.** A JSON schema with the catalogue's scheme ids, roles and CQ labels as enums. The model cannot name a scheme or a CQ that does not exist.
- **Fewer, cheaper calls:**
  - several sibling claims per call where the provider limits requests per day;
  - context caching for the fixed part;
  - the thinking budget as a measured setting.

**Done when:** version 4 beats version 3 on the development set (§5) with the same model, and the test set confirms it.

## 4. Phase 3: data

1. **The first gold (D086).** Drawn from batch 2's claims, starting where annotators A and B, the reference or a model disagree, with some claims where they agree, since two models can share a mistake. The developer drafts, the other independent reviewer reviews, back and forth until they converge; the owner verifies, and the dashboard's verdicts are the record. This is the first gold a person stands behind.
2. **Re-code batch 2 under version 4 and the catalogue.** Gold files are versioned with the guidelines they follow.
3. **More discussions, of other kinds:** empirical and causal, ethical, heavy on expert testimony, legal.
   - Four to nine discussions of about 50 claims each, sampled across depth.
   - The dashboard already makes a dataset from a captured Kialo discussion.
   - Who wrote a claim and when, which the capture records, go in a private side file of the dataset, not in the claim's record: the annotator sends the whole record to the model's provider.
4. **A probe set.** For each tier-1 scheme, two short claims that instantiate it and one near miss, written and verified. It measures recognition per scheme, cheaply, and it catches regressions.

## 5. Phase 4: evaluation and debugging

| Practice | How |
|---|---|
| Acceptable answers | The gold lists, per field, the answers the guidelines allow. Scores report exact and acceptable agreement. |
| A cause for every disagreement | Model error, guidelines gap, both allowed, catalogue gap, gold error, or data. The dashboard verdict gets these as choices, so the record builds up while reviewing. |
| Development and test sets | Prompts are tuned on the development claims only. The test claims are scored at the end, once per version. |
| One change at a time | Same claims, same model, one difference. `run.json` records the inputs' and the prompt's hashes. |
| Re-scoring for free | `calls.jsonl` holds every request and answer. A new checker or metric is re-run without calls; the replay provider re-runs the pipeline on recorded answers. |
| Per-scheme results | A confusion matrix of schemes, agreement at family level, recall on the probe set. |
| Validity | The share of answers the checker accepts, before and after one repair round. |
| Several models | At least two providers annotate, and a third reviews. A model never reviews itself. |
| Cost | Tokens, calls and seconds per claim, from the traces. |
| Play | Per scheme and per question: how often played, asked, answered with success, and noted as wrong or missing; from exported games and traces ([`process.md`](process.md) §4). |

## 6. Phase 5: the game takes the catalogue

**What makes a scheme playable.** A scheme is playable in a context when:
1. **Formulas.** Its formulas are in the formula layer, and its entry is at least checked. It is verified before a release.
2. **Questions.** Each of its questions names the NASCQ question it becomes, with a locus and a pattern.
3. **Game fields.** They are set, each with a status (below).
4. **English.** Its English renders every premise, conclusion and question.
5. **Checks and play.** The lint and the game's tests pass, and scripted play exercises each of its ask and backup edges.
6. **Weight.** The context gives it a weight. A weight of zero means the move can be made but carries nothing there.

**Contexts.** A context sets:
- the dialogue type (Track D, §8);
- the proof standard;
- whether evidence is required;
- whether a moderator is expected;
- the weight an audience gives each scheme.

Today's presets (`data/contexts.json`) set all but the dialogue type. The catalogue's "proper in" field is the evidence for each context's weights.

**Fallacies need no system of their own.** In Walton's account, a fallacy is a bad use of an ordinary scheme, not a scheme of its own: the fallacious appeal to authority misuses argument from expert opinion. The context's weight and the scheme's own critical questions handle it. For example, ad hominem asks whether the person's character bears on the issue. A dialogue's ordinary rules may still refuse an illegal move, and ordinary argument and evidence checks may reject support; neither is a fallacy system (the owner's decision 4 of 2026-10-22, confirmed on 2026-11-24 with this addition).

Some faults are not misused schemes: begging the question, the loaded question, the straw man, equivocation. They are faults of moves and commitments, and need the dialogue's rules: commitment stores and the presuppositions of questions (Track D's D2a and D2d), and a check for circular arguments.

**The game's fields.** These are the NASCQ fields that the texts do not define ([part 13 §1](review-2026-10-22-part13.md#1-what-the-review-found)). Each becomes an item of the catalogue entry, with a status, a reason and example IDs. There are three kinds:

| Kind | Fields | How it is checked |
|---|---|---|
| **Derived by a rule** | the Carneades id; the variables; each question's locus and pattern (from its target); what each answer does (from the locus); whether the scheme can back any statement (its conclusion is a bare proposition, unless the formula ties it to a case, as in analogy); the answer variable; the English (from the formula layer) | the checker re-derives them; an exception to a rule is recorded with its reason |
| **Judged, with examples** | the kind of claim (action, value, causality, truth) of schemes and questions; variable types and examples; cues; the evidence policy; which NASCQ question a book question becomes; backup edges that the rules do not give | each judgment cites coded examples; the lint checks consistency |
| **Calibrated by data and play** | base weights; each context's weights | measured, then set |

**Which NASCQ question a book question becomes.** The NASCQ's questions already serve several schemes: 9 of its 24 are asked of more than one. Reuse an existing question when the book's question asks the same thing and its answers act the same way, since the NASCQ question fixes that for every scheme that asks it. `CQ:likelihood` serves both consequence schemes and sign on this basis. Otherwise make a new one. Each catalogue question records its choice.

**Links across schemes.** Which scheme can answer which question, and where one scheme's conclusion feeds another, is not in the sources. On 2026-10-22, 10 of `walton.yml`'s 388 premise-side formulas were the conclusion of some specific other scheme ([part 13 §1](review-2026-10-22-part13.md#1-what-the-review-found)).
- **Rules, not lists.**
  - Backup edges follow from rules over what a scheme concludes and what a question needs, plus the generic schemes that can back any statement.
  - Hand-written edges are allowed, each with its reason and examples.
  - A new scheme gets its links by declaring its predicates and kinds.
- **A shared vocabulary** of meeting points is kept in one place: `should_be_performed`, the causal predicates, the value predicates, and others as found. The checker reports how well the schemes connect.
- **Families are modules.** They organize the work and the vocabulary; they drive no mechanics.
- **Refined by data and play.** An answer in coded data that no edge allows becomes a candidate edge, and so does a "could not make the move" report.

**Order.**
1. The 9 game schemes' NASCQ is generated from the catalogue, and the first output must reproduce today's.
2. Tier 1's schemes follow, then the rest family by family, as data and play ask.
3. Derived schemes join the same way.
4. Model players use the same catalogue for the schemes they may play.
5. "Other" is played as a defeasible modus ponens, so that its arguments can be asked for backing and exceptions.

**Dialogue types:** Track D (§8). The presets were renamed on 2026-10-15 to "Policy question" and "Scientific question", since the game does not play deliberation or inquiry.

**Peer-to-peer play** (no trusted host: every player runs the engine and checks every move) is not scheduled. Ideas and the steps it would take are in [peer-to-peer.md](peer-to-peer.md), noted on 2026-11-24 at the owner's request; it would build on D073's journal (§0, stage 2).

<a name="7-decisions-the-owner-needs-to-make"></a>

## 7. Decision record and remaining implementation

The [decision register](decisions.md) is authoritative, and its view gives the
counts. The open entries are questions of fact for a person (D021, D027, D041,
D085), D079, which the owner reserved, and D083, decided when its candidates'
sketches are ready. A candidate's `ready` in the generated view means its
minimum is complete; the decision still judges its meaning and its `needs`.

The earlier questions in this section now have these recorded dispositions:

1. **Book text in the public tree (D043) and history (D078).** D043 assigns the
   transcription, source quotations and repeated templates to the current-tree
   cleanup in stage 2: retain private source material, use labels/page references
   and paraphrases publicly, and preserve the copy check with hashed word
   sequences. The implementation remains open. D078 preserves the existing
   public history. The deleted *New Dialectic* PDF therefore remains in that
   history; rewriting it is not a remaining choice in this plan. The measured
   inventory (173 of `walton.yml`'s 393 templates share eight or more words with
   the compendium) is in the [catalogue's README](../wag/data/catalogue/README.md)
   and [catalogue-alignment.md](catalogue-alignment.md).
2. **Licences (D044, D080).** The repository now includes MPL-2.0 for WAG's
   code and CC BY 4.0 for its own documents and annotations. The actual scope,
   mixed files and excluded third-party material are recorded in
   [LICENSING.md](../LICENSING.md). These choices are implemented, not open
   questions for the owner.
3. **Catalogue verification (D045).** Tier 1 remains the first set. Models can
   draft and check, and work can proceed provisionally from checked entries;
   a person supplies the required verified status. The formula/source
   dispositions already adopted are in the register and
   [walton-clean/CHANGELOG.md](scheme-registry/walton-clean/CHANGELOG.md).
4. **Paid model access (D046, amended by the owner on 2026-11-23).** Runs
   with the owner's keys run on the owner's machine. A run has no limit unless
   the owner sets one, and `--unlimited` says so explicitly. The CLI's call cap
   is exact, its token and cost limits are estimates, and each invocation
   records its limits and what it used, as described in §0 and
   [llm.md](llm.md#a-runs-budget-d046). Remaining: the dashboard's budget and
   record; an allowance shared by several invocations (their records can be
   summed, but nothing enforces a total); and a provider-side limit for any
   claim of a hard token or cost ceiling. Historical price or free-tier
   figures are not a current allowance or benchmark.
5. **Candidate threshold (D047).** Three verified sightings in two discussions or games
   trigger owner review. Archived and strained sightings count but are
   reported separately from clean sightings. Reaching the threshold is not
   semantic approval of a sketch.
6. **Dialogue types (D048–D050 and Track D).** Negotiation has its own profile
   and a per-issue score vector. The experimental quarrel profile uses the
   recorded fictional/model-seat and participation rules. Real discourse and
   WAG traces serve distinct purposes: discourse exposes unanticipated
   patterns; native traces test legality, answerability and the protocol.
   Reconstructed Kialo play stays distinct from observed dialogue and
   model-authored arguments. Real discourse comes first, each corpus recorded
   with its licence (D050). These are the owner's decisions D048-D050, for
   later implementation; candidate corpora still need source-specific
   assessment.
7. **Goal and authorities,** decided on 2026-10-22 ([part 13 §3](review-2026-10-22-part13.md#3-the-owners-decisions)),
   with CI on the public repository (D052, [`process.md`](process.md) §7). The goal remains that every scheme and critical
   question can be played in some appropriate context. The book governs its
   scheme meanings; WAG's fields and links are justified by coherence,
   examples and measured data/play. The library supplies draft records where
   WAG has no decision. Independent review, explicit provenance and the
   [process](process.md) apply; empirical consensus does not replace required
   source checks.

Open decisions and implementation debt have different statuses. In particular,
the open register entries do not count every unfinished stage-2 task, and
an adopted direction or a passing mock test does not establish full deployment
or empirical validation.

## 8. Track D: dialogue types

**The objective,** set by the owner on 2026-10-15:
- WAG can play a game in any established dialogue type, the quarrel included.
- WAG can find every type in real discourse, negotiation and quarrels included, as it does for schemes.

**Its place.** The project's focus stays on argumentation schemes, critical questions, and a gold standard made from Kialo (phases 1–3, measured as phase 4 describes). Track D comes after them, except for the two small pieces that phases 1 and 2 need anyway, and a small first persuasion protocol once the first gold and the pilot have shown where WAG falls short, while the corpus grows (D087; see Order). The playable quarrel is the last step of all.

**Where it is explained:**
- **What the types are:** [`dialogue-types.md`](dialogue-types.md).
- **Whether each part is feasible:** [`dialogue-types.md` §9.2](dialogue-types.md#92-can-wag-play-every-type-and-find-every-type-in-data).
- **Why this track:** [part 9](review-2026-10-15-part9.md).

### Principles

- **Definitions come from the sources.** Walton's *The New Dialectic* (1998) comes first. Walton & Krabbe's *Commitment in Dialogue* (1995) is read before any code on commitments or shifts.
- **A type is recorded only where it changes something:** a rule in the game, or a field in annotation.
- **Mixtures are normal.** A dialogue has several typed segments; shifts are events between them.
- **Dimensions as well as types.** Annotation records the aim, the situation and any third party, then the type (reference §2.7). A dialogue that fits no type keeps its dimensions instead of being forced into one.
- **A newly proposed type needs evidence, as a new scheme does.** A pattern no recognized type covers must recur, in at least three dialogues from at least two sources, before a person reviews it as a candidate. This Track D policy is separate from D047. Advice-solicitation is already a subtype in ND; team forming and exhortation remain unconfirmed candidates. Discovery and adjudication have literature-described status after the targeted source check in [reference §10](dialogue-types.md#10-further-reading), without becoming playable or changing this track's order. Examination is discussed in the reference as a subtype and in later work.
- **Persuasion is the base.** Other types come one at a time, each playable end to end and tested before the next.
- **The quarrel comes last, with safeguards.** Research needs a playable quarrel: to produce labelled quarrels, to study how other dialogues slide into one, and to test detectors. Detecting quarrels comes first (D2g); the playable quarrel (D2h) is built last, with:
  - model seats by default, and people only by their own choice;
  - fictional scenarios, with no real person or group as a target;
  - limits on content, and a moderator;
  - every transcript labelled as synthetic.

### Steps

| Step | What | Needs | Effort | Done when |
|---|---|---|---|---|
| **D0. Reference** | [`dialogue-types.md`](dialogue-types.md): the types, the shared concepts, shifts, the arguments whose standing depends on the type, cues, a data model | — | Drafted on 2026-10-15; human check pending | A person checks its page references against the book |
| **D1. Record the type** | (a) Game settings and exports carry the declared `dialogue_type` and profile, and the persuasion form where applicable (one-sided, or a dispute). A mode or practical topic alone does not establish a type; unsupported classifications remain explicit. (b) Step 0 of the version 4 guidelines: setting, type supposed and enacted, thesis kind, initial situation. (c) A dialogue-type catalogue: one entry per type and subtype, drafted from the reference, then verified. Proposed types (reference §2.7) get entries marked *candidate*, with their source and status. (d) "Proper in" on the scheme entries (§2). | Phase 1's skeleton for (c) and (d); phase 2 for (b) | 1–2 sessions | Exports carry the type; the guidelines have step 0; the six type entries pass a checker |
| **D2a. Persuasion with commitments** | **Commitment stores** with assert, concede, retract and challenge. **Rules:** rule 1 as an option (premises must be the respondent's commitments, or be defended when challenged); the retraction rules. **A dispute variant:** both sides must prove. **Stages:** an opening that agrees the issue and the standard, and a closing. **Variants:** permissive and rigorous as settings. **Model players** get hidden positions, so that the maieutic outcome can be measured. The text protocol's next version carries all of this. | Walton & Krabbe 1995 read | 1 week | Existing games replay unchanged; each rule has a test; a model game is played under each variant |
| **D2b. Deliberation** | **Opening:** a practical problem and a shared goal. **Roles:** symmetric. **Proposals:** one in the simplest case, then several, as the positions of a Carneades issue (the adapter handles two today). **Outcome:** a joint, tentative decision. **A caution setting:** a presumption against a proposal with a possible serious harm. | D2a | 1 week | A three-option game between models; a Kialo practical thesis can be opened as a deliberation |
| **D2c. Inquiry** | **First profile: strict cumulative inquiry. One team. No retraction within that profile. Sources required.** **Stages:** data, analysis, presentation. **Outcome:** proven, disproven, or open. **Export:** the result as an ordered proof. This is a chosen profile, not a universal definition of inquiry; a revisable profile needs separate rules for retained history and dependent support. | D2a | Days | Two models cooperate to prove a claim, or fail; a retraction in the strict profile is refused with its reason |
| **D2d. Information-seeking** | **An answerer seat that holds knowledge:** a document, a scenario card, or a model. **Questions** with presuppositions; direct and indirect answers; challenging a presupposition. **Expert consultation,** probed through the expert-opinion critical questions. | D2a | 1 week | A questioner finds the facts on a hidden card; a loaded question is flagged |
| **D2e. Negotiation** | **The bargaining core:** issues and options, private interests and reservation values. **Moves:** offers, counter-offers, accept, reject; promises and threats. **Duties:** Donohue's response rules. **Arguments for offers** go through the schemes. **Deadlock,** and a shift to information-seeking. Distributive and integrative subtypes first. | D2a; the owner's decision on scoring | 2–3 weeks | Two models reach a deal or fail; the outcome is scored; a threat inside a persuasion game is flagged |
| **D2f. Shifts and embedding** | **A stack of dialogues:** a subdialogue opened by agreement; its results carried back when it is functionally joined. **A flag** on a move of another type made without an agreed shift. | Two of D2b–e | 1 week | A deliberation embeds an information-seeking subdialogue; an unagreed threat is flagged as an illicit shift |
| **D2g. Eristic, detection** | "Spot the shift": on real or recorded dialogues, mark where a quarrel begins and which move started it; scored against a gold | D3's gold | Days | It runs on 20 dialogues with a gold |
| **D2h. Eristic, playable (last)** | **A quarrel protocol after ND ch. 7.** It opens from a grievance in a scenario card. **Moves:** blame, counter-blame ("you too"), personal attack, straw man, change of topic, concession, reconciliation. **It ends** when the grievance has been aired (catharsis) or the quarrel is broken off. **Subtypes as settings:** natural, staged, group, sophistical. **A research variant:** a persuasion game in which one seat tries to provoke a quarrel and the other tries to keep to the issue. **Output:** transcripts labelled with their segments and shifts. The safeguards are in the principles above. | D2f, D2g, D3 | 1–2 weeks | Two model seats play a staged quarrel from a scenario card; D3's detector finds its shift points; every transcript passes the content limits |
| **D3. Identification** | **Guidelines:** the dimensions first, then Walton's questions, the cues, the cautions (reference §2.7, §8). **Model annotators and a reviewer** on dialogues and segments: a new task in the annotation package. **People verify.** **Agreement:** kappa on each dimension and on the type; segment boundaries; shifts found; licit or not. **Candidates:** dialogues that fit no type, grouped by their dimensions. | D1 | 1–2 weeks of code, then data work | Agreement measured on at least 30 dialogues across at least four types; a baseline classifier reported |
| **D4. Corpora** | At least one source per type (below) | The owner's decision on licences | Ongoing | Each of the six types has a source with at least 20 annotated dialogues |

### Order

Set by the owner on 2026-10-15: schemes, critical questions and the Kialo gold first. Read by the owner on 2026-11-24 (D087): "the Kialo gold" is the first gold a person has verified (D086) and the bounded pilot (stage 3), not a complete corpus.

1. **D0's reference is drafted.** Its existing completion condition—a person checking the page references against the book—remains open.
2. **Only what phases 1 and 2 need anyway:**
   - **"Proper in" (D1 d)** is a field of the scheme catalogue (phase 1). It is drafted from the reference's §7.
   - **Step 0 of the guidelines (D1 b)** goes with phase 2. It is a few fields filled once per Kialo discussion.
3. **After the first gold and the pilot (D087):**
   1. a small first persuasion protocol, the next experiment while the corpus grows: the start of D2a, with what it needs of D1 (a), scoped by the deficiencies the gold and the pilot show. A run made by models is not gold;
   2. the rest of D1 (a), and D1 (c);
   3. the rest of D2a, then D2b to D2f, in the order shown;
   4. D3 and D4, starting from Kialo's own data (persuasion, much of it about action), then other corpora.
4. **Last:** D2g, then D2h, the playable quarrel.

The scheme catalogue stays the first priority, because every type that argues uses it.

### Candidate corpora

None of these has been checked here for licence or fit.

| Type | Candidates |
|---|---|
| Persuasion | Kialo (in hand); ChangeMyView (Tan et al., 2016) |
| Deliberation | Meeting and town-hall transcripts (Kialo's practical theses carry deliberation's arguments, but are persuasion about action); DeliData (Karadzhov et al., 2023), group problem-solving that is close to inquiry |
| Inquiry | Peer-review and replication discussions; public inquiry reports |
| Information-seeking | QuAC (Choi et al., 2018); interview transcripts |
| Negotiation | Deal or No Deal (Lewis et al., 2017); CraigslistBargain (He et al., 2018); CaSiNo (Chawla et al., 2021) |
| Eristic | Conversations Gone Awry (Zhang et al., 2018); parliamentary question periods (Hansard) |
| Mixed | AIFdb's annotated debates (for example the Moral Maze radio debates, and the 2016 US presidential debates); Question Period |
