# Step C5: the game takes the first family

**Status:** a plan. Version 1 (3bc7f0c) was reviewed independently on 2026-11-26 (in the
workspace, verify/c5/review-v1); this is version 2, which follows the review's findings once each
was checked. Its choices are registered as D094, decided under the owner's instruction of
2026-11-25 to go with the recommended choices and reversible before the work they govern. Step C5
of [first-family.md](first-family.md): the game's own data takes what proposals D004-D075 (C2),
D008-D074 (C3) and D022-D093 (C4) gave `walton.yml`, through the generator that D007 and D009 name.

## What is already decided for the game

C5 applies decisions, most of them already worded for the game; this plan does not reopen them.

- **C2's game paragraph** (proposal D004-D075, *What else changes*): AS:practical_reasoning and
  AS:practical_reasoning_goal take the goal premise and the scoped exceptions; CQ:other_goals and
  CQ:conflicting_goals target `incompatible_goal(P,G)` (D004); CQ:side_effect targets
  `bad_outweighs(A,G)`, with `bad_outweighs_good` as its backup edge (D070); CQ:other_means gets
  `better_means` (D089); CQ:goal_possible and CQ:possibility target `possible_goal(G)` and
  `possible_bring_about(P,A)` (D088); CQ:best_means targets `bring_about_more_effectively(P,S2,A)`.
  Each backup edge has an answer that must pass, one that must be refused and one bound to the
  wrong target. The game refuses an answer whose goal premise is not the answered argument's,
  whose A2 is A or not P's action, or whose P is not the agent **the game's profile** names, which
  also names the situation compared (D089).
- **C3's game paragraph** (proposal D008-D074, *What else changes*): the two consequence schemes take
  the new exceptions, with game questions for `too_slight` and the bearer pair (D065); D074's
  objection is a question about the argument itself; the modal variants and the bridges join the
  NASCQ, with a question for `worth_finding_out_first` (`matters_to_decision` is a premise,
  challenged as any premise is); the question protocol of D003 and D075 (who answers an asked
  question, and what holds while it is pending); answers at the new exceptions get answer schemes
  or stay Other. 33.4's `fails_to_avert` and `harm_reduction` enter the game with a profile that
  includes them, not at C5. Once the game has the modal variants and the bridges, `to_wag.py` plays
  a claim's instance that feeds its primary one, keyed by instance.
- **The entries' recorded route for 33/CQ3** (the game-differs discrepancies of
  positive_consequences and negative_consequences, which D009 adopts as "the changes the entries
  list"): the game keeps the rebuttal route at the conclusion, CQ:side_effect on AS:consequences and
  CQ:opposite_consequences on AS:negative_consequences, with its backup edge from AS:consequences;
  and adds a question at the exception, `bad_outweighs(A,G)` and `good_outweighs(A,G)`, for the
  weighing route (D055), answered by C003's answer schemes. The weighing questions need ids of their
  own: an edge backs a question, whatever scheme asks it, and `bad_outweighs_good` cannot conclude
  the conclusion-route answer (S9).
- **D009:** the question texts come from the catalogue's paraphrases, with the changes the entries
  list (among them: the evidence question sits on 33/CQ2, at the premise `brings_about(A,G)`);
  CQ:field is kept; each generated question has a test that its target is one of its scheme's,
  covering the catalogue version, the target's role and type, the substitution, the instance, the
  polarity, the response's effect and its replay. **D007:** practical reasoning in the game is the
  book's ch. 9 version; ch. 3's further questions are derived questions with chapter and page.
  **D006:** a mere alternative answers only a claim that the action is necessary or the best, so
  CQ:other_means asks for a better means.
- **D008:** a bridge is an entry condition, not a weight: equal weights, no multiplier.
- **D053's declaration:** a bare claim that a consequence could be investigated does not establish
  `worth_finding_out_first`; the critic must show that finding out first is timely, reliable and
  worth its cost.
- **Workstreams E and I** (catalogue-alignment.md): the game's fields go into the entries with the
  status `generated`; the derived fields get rules that the checker re-derives; the generator's
  first output reproduces today's NASCQ and core.yml except where the catalogue records a reason,
  and the game's tests pass unchanged; core.yml's formulas come from the formula layer.

## What the game's data does today

From an inventory of the code (2026-11-26, verify/c5/inventory.md, checked by the review):

- **Every field is read.** `caes_mapping.negative_answer_type` is shown to players (app.js), and is
  already wrong where a question is asked at different loci; `semantic_type` serves display and a
  lint warning. Other's questions (CQ:relevance and the rest) are defined in code, outside the files.
- **Order is behaviour.** Edge order fixes the order of the offered moves, so the bots' moves and
  the 33 golden games; the order of premises in core.yml matters too.
- **Bytes are identity.** The SHA-256 of core.yml and the three NASCQ files is every saved game's
  `data_sha256` (`e7141ec8dea379bc` today); a game under other data replays only read-only, and
  `wag-llm replay` needs `--data` with its own files. The bytes go into annotation datasets, so a
  run made before a change can be read but not continued. Both copies of the batch-2 gold pin the
  NASCQ files' hashes (validate.py, R16).
- **core.yml need not equal `walton.yml`;** the catalogue records the differences (game-differs:
  practical_reasoning1 12, practical_reasoning2 13, the consequence schemes 3 each, sign 1, analogy
  10). Its language (39 templates and their negations) is the game's own.
- **The catalogue's direction is from the game:** `--catalogue-gen` writes each entry's `game` part
  from the NASCQ, and the check requires them equal.

## C5.1: the generator, with no change of meaning

Workstream I.1 and E.1-E.2 for the nine game schemes.

1. **Tests first.** The game tests that edit the data by text anchors (lint_test.go and the like)
   are made structured, so that a change of format leaves them unchanged (E.3).
2. **The fields.** Each game scheme's entry holds its NASCQ record and the game's copy of its
   formulas and English, with the status `generated` (I.1): the copy is a recorded override of the
   formula layer, with its reason, until the family is taken up. The shared questions are one file,
   `wag/data/catalogue/game/questions.yml`; the edge order is data, `game/order.yml`; core.yml's
   comments, language and the scheme kept for old games, `game/core.yml.in`.
3. **The rules (I.2-I.3).** A question's locus follows from its target's role (an assumption is
   asked at the premise), its pattern from the target formula; each exception today's data has is
   recorded with its reason (CQ:other_goals at the conclusion, the questions asked of the argument).
   The checker re-derives them. `caes_mapping` stays recorded in C5.1 (C5.2 derives it).
4. **The generator.** `wag-cli --game-gen` writes the NASCQ files and core.yml; core.yml's header
   carries the hash of the generator's inputs (the catalogue version D009's test covers). The check
   requires the committed files to equal its output; `--catalogue-gen` keeps the entries' game parts
   as the source.
5. **Reproduction.** The pre-C5 files are frozen in testdata with their hash. A test requires the
   generator's output to equal them as JSON and YAML trees (every field, the edge order, the
   premise order), not through the loader.
6. **The switch.** The committed files become the generator's output in its format: the bytes change
   once, the meaning not (**choice a**). In the same commit both gold copies are re-pinned, with a
   history entry; the batch-2 replay's report takes the new hash; the docs say how to replay an old
   game or trace (`--data` with the frozen files). A hashless export, which today replays as
   current, opens as legacy (adverse case 11).
7. **Review and gates** on C5.1 alone; no game test changes but those listed with their reason.

## C5.2: the slice, through the catalogue

One change of data identity for all of it (**choice a**), made on a branch-local sequence whose
steps each pass the generator's checks.

- **Formulas.** The four game schemes come from `walton.yml`, their overrides removed (game-differs
  to 0); sign, analogy and the other three keep theirs. English: the new predicates take
  `walton.yml`'s templates, agreement-free since C4, with negations written; the game's other
  templates stay, each difference recorded (**choice f**).
- **The profile** (D089, C2's paragraph): every game records at its start the agent, the actions it
  deliberates with their actors and aliases (D091, D092), and the situation compared. In a mode
  that sets a claim it comes from the claim's case; in the others it is asked when the game is
  created, with a default the board shows (**choice b**). The game binds P (bots skip bound
  variables); a move about an action of another actor binds that actor; the game refuses a P that
  does not perform A, aliases included. GameConfig, export, load and the trace carry the profile;
  /api/games, app.js, the model players' prompt and `to_wag.py` (from case.json) follow.
- **Questions:** C2's paragraph for practical reasoning; for the consequence schemes the entries'
  two routes, the weighing questions under new ids backed by C003's schemes, the evidence question
  at the premise, and new questions for `too_slight` and each side's bearer exception (D065). D006's
  text for CQ:other_means. `caes_mapping`'s effect derived from each edge's locus, and shown per
  edge.
- **D074** (**choice d**): in the game a consequence scheme cannot conclude a factual claim (S9), so
  the misplaced appeal meets only Other's arguments; the question about the argument itself is
  Other's CQ:relevance, its text naming the case, with the direct undercutter; the consequence
  schemes get no question for it (a reasoned departure from C3's wording).
- **D009's scope** (**choice e**): the slice's questions take the paraphrases; a question shared with
  a scheme outside the slice is split where their paraphrases differ (CQ:likelihood: sign keeps its
  own), not deferred.
- **New kinds of scheme.** The answer schemes, the modal variants and the bridges join the NASCQ
  with a role (answer, modal, bridge): an answer scheme is offered only as an answer, a bridge as a
  move for or against a claim whose first premise a modal argument supports; TestCoreDataLoading
  allows a scheme with no question where D075's reasoned none applies; the context presets weigh the
  new schemes as their categorical counterparts, the bridges with no multiplier (D008; **choice
  g**).
- **D053** (**choice c**): an answer scheme concludes `worth_finding_out_first(A,G)` from the
  declaration's conditions (an investigation timely and reliable enough to settle whether A brings
  about G; its costs and risks justify finding out first), proposed for `walton.yml` through the
  proposal process; at that question the game accepts it and the source-based schemes that assert
  the exception itself, not Other.
- **D003's question protocol:** who answers an asked question, and what holds while it is pending,
  written as the game's rule and tested (the critic's exceptions stay the critic's).
- **The gold:** validate.py's question table follows the new edges; the labels it then refuses
  (13292.127's `22.1 CQ5`, and the like) are recoded by a script, and the gold re-pinned.
- **The replays:** the 33 golden games stay on the frozen pre-C5 data, which `RunGame` takes as
  bytes, and keep their digests; a new golden set is made on the C5.2 data (**choice h**). The
  batch-2 replay plays by instance (C3) and reports the claims C2 to C4 left unplayed that now play
  (13292.104 and .123's better means, 13292.127 and .43's weighings, the modal claims).

## Item F: the answer contract, per route

D075 and the critical review's row F ask it per question and answer route. The routes of the
included schemes: each backup edge (the three answer schemes, sign's three, positive consequences'
rebuttal edge); the thesis-independent schemes (expert opinion, position to know), which back every
question; Other; a question asked alone; the direct rebuttal where an answer route is removed. For
each, the expected outcome is written before it runs (review item I) and paired with a planted
defect the test must catch:

- **Accepted:** the legitimate answer, at the right target and locus, with its burden.
- **Refused, with the rule that refuses it** (S5, S8, S9, Q1, the cycle check, the profile's P, the
  A2 rule), leaving the graph, the open questions and the journal unchanged: an irrelevant answer,
  a circular one, one bound to the wrong target, and for `better_means` an A2 that is A or its
  alias, another P or another G.
- **Surviving:** the target's other questions, its other exceptions and independent support.

The first-family adverse cases of the critical review of 2026-11-23: **5**, CQ:possibility, an
assumption, asked alone ties the presumption, while an exception's question alone is refused (Q1)
and a tie leaves the exception out; **6**, `bad_outweighs_good` defeats practical reasoning and
positive consequences on the same A and G, not another G, not negative consequences; `too_slight`
defeats both sides; `worth_finding_out_first` defeats both bridges about A and G, not the modal or
categorical arguments, not a bridge about another G (D053); **7**, a negative-consequences answer at
`bad_outweighs` is refused (S9); `better_means` with A2 = A, or another P or G, is refused; a
source-based or Other answer's relevance is challenged in play (CQ:field, CQ:relevance), not
refused, and the plan says so; **8**, undermining `matters_more(H,G)` restores the target with its
other questions still open, and independent support stays in. Two tests C3 promised: a modal
argument alone does not support `should_be_performed(A)`; a bridge played with its premise
challenged.

## Order of work

1. This plan and D094, checked independently.
2. C5.1, in the order above; gates on both toolchains; independent review; push.
3. The answer scheme for `worth_finding_out_first`, a proposal for `walton.yml` through the
   process.
4. C5.2, one data identity, with the profile, the questions, the new schemes, the protocol, item F's
   tests, the gold's recode and the replays; gates; independent review; push.

## Not in C5

The other five game schemes' changes (sign, analogy, expert opinion, position to know, values) wait
for their families; their fields move to the catalogue in C5.1 all the same. 33.4's
`fails_to_avert` and `harm_reduction` wait for a profile that includes them (C3). The contexts'
weights for today's schemes stay as they are (calibrated, plan §6).
