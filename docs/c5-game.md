# Step C5: the game takes the first family

**Status:** C5.1 is applied (2026-11-26: c9935ce, 17fd22a and the switch); C5.2 waits for D095. The
plan: version 1 (3bc7f0c) and version 2 (3471585) were each checked independently on
2026-11-26 (in the workspace, verify/c5); this is version 3. Its implementation choices, within the
decisions it applies, are D094. Three choices depart from the wording of decided texts, so they are
the owner's, put as D095 with a recommendation each; C5.2's parts that depend on them wait for the
answer. Step C5 of [first-family.md](first-family.md): the game's own data takes what proposals
D004-D075 (C2), D008-D074 (C3) and D022-D093 (C4) gave `walton.yml`, through the generator that
D007 and D009 name.

## What is already decided for the game

- **C2's game paragraph** (proposal D004-D075, *What else changes*): AS:practical_reasoning and
  AS:practical_reasoning_goal take the goal premise and the scoped exceptions; CQ:other_goals and
  CQ:conflicting_goals target `incompatible_goal(P,G)` (D004); CQ:side_effect targets
  `bad_outweighs(A,G)`, with `bad_outweighs_good` as its backup edge (D070); CQ:other_means gets
  `better_means` (D089); CQ:goal_possible and CQ:possibility target `possible_goal(G)` and
  `possible_bring_about(P,A)` (D088); CQ:best_means targets `bring_about_more_effectively(P,S2,A)`.
  Each backup edge has an answer that must pass, one that must be refused and one bound to the
  wrong target. The game refuses an answer whose goal premise is not the answered argument's, whose
  A2 is A or not P's action, or whose P is not the agent the game's profile names, which also names
  the situation compared (D089); an answer records its comparison (more surely, sooner, at lower
  cost).
- **C3's game paragraph** (proposal D008-D074): the two consequence schemes take the new exceptions,
  with game questions for `too_slight` and the bearer pair (D065); D074's objection is a question
  about the argument itself; the modal variants and the bridges join the NASCQ, with a question for
  `worth_finding_out_first` (`matters_to_decision` is a premise); the question protocol of D003 and
  D075 (who answers an asked question, what holds while it is pending); answers at the new
  exceptions get answer schemes or stay Other. 33.4's `fails_to_avert` and `harm_reduction` enter the
  game with a profile that includes them. Once the bridges are in the game, `to_wag.py` plays by
  instance. A modal argument offered for `should_be_performed(A)` is refused (S9); a bridge is played
  with its premise challenged.
- **The entries' route for 33/CQ3** (the game-differs discrepancies of positive_consequences and
  negative_consequences, which D009 adopts): the rebuttal route stays at the conclusion,
  CQ:side_effect on AS:consequences and CQ:opposite_consequences on AS:negative_consequences, with
  its backup edge from AS:consequences; a question at the exception, `bad_outweighs(A,G)` and
  `good_outweighs(A,G)`, gives the weighing route (D055), answered by C003's answer schemes. With
  C2's wording above, this asks CQ:side_effect to be the weighing question on practical reasoning and
  the rebuttal question on positive consequences; an edge backs a question whatever scheme asks it,
  so one id cannot be both (S9 would refuse `bad_outweighs_good` at the conclusion): **D095 (1)**.
- **D009:** the question texts come from the catalogue's paraphrases (all 35 game questions have
  one), with the changes the entries list, among them sign's own likelihood text and the evidence
  question at the premise `brings_about(A,G)`; CQ:field is kept; each generated question has a test
  that its target is one of its scheme's, covering the catalogue version, the target's role and
  type, the substitution, the instance, the polarity, the effect and the replay. **D007:** the
  game's practical reasoning is ch. 9's; ch. 3's further questions are derived questions in the
  catalogue, with chapter and page. **D006:** CQ:other_means asks for a better means.
- **D008:** a bridge is an entry condition, not a weight. **D053:** a bare claim that a consequence
  could be investigated does not establish `worth_finding_out_first` (the owner's declaration).
- **Workstreams E and I** (catalogue-alignment.md): the game's fields go into the entries with the
  status `generated`; derived fields get rules the checker re-derives; the first output reproduces
  today's data except where the catalogue records a reason; the game's tests pass unchanged; the
  formulas come from the formula layer.

## What the game's data did before C5.1

From an inventory of the code (verify/c5/inventory.md), checked twice; since C5.1's switch the direction is the
other way (step 6):

- **Every field is read but `caes_mapping.target`.** `negative_answer_type` is shown to players
  (app.js) and is wrong where a question is asked at different loci. Other's questions are defined
  in code, outside the files and their hash.
- **Order is behaviour:** the order of the edges, and of the premises in core.yml.
- **Bytes are identity:** core.yml and the three NASCQ files make every saved game's `data_sha256`
  (`e7141ec8dea379bc`); a game under other data replays only read-only; `wag-llm replay` needs
  `--data` with its own files and compares prompts byte for byte; an annotation run made before a
  change can be read, not continued; both copies of the batch-2 gold pin the NASCQ files (R16).
- **core.yml need not equal `walton.yml`:** game-differs records the differences (practical
  reasoning 12 and 13, the consequence schemes 3 each, sign 1, analogy 10); expert opinion,
  position to know and positive value differ in their meta only.
- **The catalogue's direction is from the game:** `--catalogue-gen` writes the entries' game parts
  from the NASCQ (generate.go), and the check requires them equal (check.go).

## C5.1: the generator, with no change of meaning

1. **Tests first:** done (6395180): the tests that edited the data by its text edit it by its
   structure.
2. **The fields.** Each game scheme's entry holds its NASCQ record, its meta and, where the game's
   formulas differ from `walton.yml`'s (practical reasoning, the consequence schemes, sign,
   analogy), its formulas as a recorded override with its reason (status `generated`). The shared
   questions are one file, `wag/data/catalogue/game/questions.yml`; the edge order is data,
   `game/order.yml`; core.yml's comments, language and the scheme kept for old games,
   `game/core.yml.in`. Done: `--catalogue-gen` records them from today's files (each game scheme's
   entry has `game.nascq`, `game.core` and `game.status`; the reasons are its `game-differs`
   discrepancies), and the check requires them current.
3. **The rules.** A game question's locus is its target's role in the formulas the game plays (its
   override where there is one; an assumption is asked at the premise), its pattern that formula;
   where a question has several targets, its game part names the one asked; a question asked of the
   argument or of the conclusion records that. The checker re-derives them. How the book question's
   target in `walton.yml` maps to the game's formula is the game-differs record, as today. The
   generator applies the rule (asked at a premise, the pattern is a premise or an assumption of the
   formulas the game plays; at an exception, an exception; of the conclusion or the argument, none)
   and refuses a catalogue that breaks it; today's 35 questions keep it.
4. **The generator.** `wag-cli --game-gen` writes the NASCQ files and core.yml. core.yml's header
   carries the hash of a canonical form of the values the generator reads, so that drafting the
   entries' other parts changes nothing. The check requires the committed files to equal its
   output; `--catalogue-gen` and the check turn round in the same commit. Written: it reads only the
   catalogue (each game scheme's entry, its formulas where it records none of its own, and `game/`),
   refuses a catalogue whose game parts disagree with one another, and writes the header's
   fingerprint (`# catalogue: ...`, of the values it reads, `core.yml`'s template and its comments
   included). The header names `wag-cli --game-gen`, which `--catalogue-gen` runs; any change to its
   wording changes the data's fingerprint, so it waits for C5.2's change of identity. Since the switch
   `--catalogue-gen` runs it first, with
   `walton.yml`'s formulas (not the entries' copies, which a change to `walton.yml` would leave
   stale), then makes the entries from its output, so an edit to a game part is what the entries
   say; the check requires the game's files to be its output, and the entries and `game/` to be
   what their round trip gives.
5. **Reproduction.** The pre-C5 files are frozen in `wag/data/frozen/e7141ec8dea379bc`, a `--data`
   folder named by its fingerprint (done; a test requires the two to agree); a test requires the
   generator's output to equal them as JSON and YAML trees (every field, the order of the edges and
   of the premises), not through the loader. It passes: the JSON files' trees are equal with their
   keys in order, core.yml's with its mappings in any order (one block, expert opinion's, writes its
   formulas in another order of keys) and its lists in theirs. A second test plays both: every
   answer to every question (`wag-cli --edges`) and the 33 reference games end the same, move for
   move.
6. **The switch:** the bytes change once, the meaning not (D094 a). In the same commit both gold
   copies are re-pinned with a history entry, the batch-2 report takes the new hash, and the docs
   say how to replay an old game or trace (`--data` with the frozen files). No annotation run is in
   progress at the switch. A hashless export replays and is flagged as legacy, not as current
   (adverse case 11), and TestLoadReplaysAnExport keeps passing. Done: the data's fingerprint went
   from `e7141ec8dea379bc` to `afd3335a23f12f41`; the 33 reference games kept their digests; the
   batch-2 replay is the same but for its fingerprint, build and times; a game loaded from an export
   that names no data shows that it is a legacy record, and its export names none either
   (`legacy_replayed_with_data_sha256` says what it was replayed with).
7. **Review and gates** on C5.1 alone. Done: each commit passed the gates on both toolchains, and CI.
   The independent review (2026-11-26; the workspace's verify/c5/c51) found 2 errors, 7 warnings
   and 6 notes, each confirmed and fixed with a test that fails without the fix (C5.1-D; F5, the CI
   step, in 9e43ab0): a `carneades_id` must be its entry's scheme; `game.backs` and each question's
   game text are copied from `game/`, and the check says where to change them; `game.status` is the
   entry's; a comment under a marker stays in the template; the formulas are written in `core.yml`'s
   order of keys; another catalogue leaves the repository's game files alone; a continued legacy
   record names the data its later moves were played with; the check reads CRLF as LF; the generator
   refuses a formula with an undeclared variable and a backed question no scheme asks. None changes
   the data's bytes.

## C5.2: the slice, through the catalogue

One change of data identity for all of it (D094 a).

- **Formulas.** The four game schemes come from `walton.yml`, their overrides removed. The new
  predicates take `walton.yml`'s templates, with negations written; the game's other templates
  stay, each difference recorded (D094 f).
- **The profile** (D089; D094 b): every game has one: the agent, the actions it deliberates with
  their actors and aliases (D091, D092), and the situation compared. It comes from a case record
  when one is supplied (`to_wag.py` supplies the batch's case.json), and is otherwise asked when the
  game is created, with a default the board shows. A journaled move declares an action with its
  actor and aliases; a move about an undeclared action is refused, so no agent passes unchecked
  (D063). The game binds P; it refuses a P that does not perform A, aliases included, and the
  refusals of C2's paragraph; a `better_means` answer records its comparison. GameConfig, export,
  load and the trace carry the profile; /api/games, app.js, the model players' prompt and
  `to_wag.py` follow.
- **Questions:** C2's paragraph and the entries' two routes, with the weighing questions' ids as
  D095 (1) settles; D006's text for CQ:other_means; questions for `too_slight` and each side's
  bearer exception (D065); the modal variants' and the bridges' questions from their entries (new
  ids where their paraphrases differ, as the modal 33/CQ1's does from CQ:likelihood's), with the
  question for `worth_finding_out_first`; D007's derived questions in the catalogue;
  `caes_mapping`'s effect derived from each edge's locus and shown per edge. Sign's three backup
  edges keep their meaning or are decided with their must-pass answers.
- **D074 and D053's refusal:** as D095 (2) and (3) settle. CQ:relevance's text in code does not
  change.
- **New kinds of scheme** (D094 g): the answer schemes, the modal variants and the bridges join with
  a role. An answer scheme is offered only as an answer, and weighs as the schemes whose exception it
  concludes; a modal variant weighs as its categorical counterpart; a bridge is offered as a move for
  or against a claim `should_be_performed(A)`, its first premise supported by a modal argument then
  or later, at its modal variant's weight, with no multiplier (D008). TestCoreDataLoading allows a
  scheme with no question where D075's reasoned none applies.
- **D003's question protocol** written as the game's rule, formalizing today's behaviour where it
  exists; any change of behaviour is listed with the golden games it moves.
- **The gold:** validate.py's question table follows the new edges; the labels it then refuses are
  recoded by a script, and the gold re-pinned.
- **The replays:** the 33 golden games stay on the frozen pre-C5 data and keep their digests; a new
  set is made on the new data (D094 h). The batch-2 replay plays by instance and reports the claims
  C2 to C4 left unplayed that now play. Old traces with model players replay with the pre-C5.2
  build and `--data` with the frozen files.

## Item F: the answer contract, per route

Per question and answer route (D075; the critical review's row F): each backup edge (the answer
schemes, four with D095 (3)'s; sign's three; positive consequences' rebuttal edge); the
thesis-independent schemes, which back every question; Other; a question asked alone; the direct
rebuttal where an answer route is removed. Each expected outcome is written before it runs and paired
with a planted defect the test must catch:

- **Accepted:** the legitimate answer, at the right target and locus, with its burden.
- **Refused, with the rule that refuses it first** (S5 before S9; Q1; the cycle check; the
  profile's P; the A2 rule), leaving the graph, the open questions and the journal unchanged: an
  irrelevant answer, a circular one, one bound to the wrong target; for `better_means` an A2 that is
  A or its alias, another P or another G; Other at `worth_finding_out_first` (D095 (3)).
- **Surviving:** the target's other questions and exceptions, and independent support.

The critical review's first-family adverse cases: **5**, CQ:possibility (an assumption) asked alone
ties the presumption; an exception's question alone is refused (Q1); a tie leaves the exception out.
**6**, `bad_outweighs_good` defeats practical reasoning and positive consequences on the same A and
G, not another G, not negative consequences; `too_slight` defeats both sides; `worth_finding_out_first`
defeats both bridges about A and G, not the modal or categorical arguments, not a bridge about another
G (D053). **7**, a negative-consequences answer at `bad_outweighs` is refused (S5, not backed);
`better_means` with A2 = A, another P or another G is refused; a source-based or Other answer's
relevance is challenged in play (CQ:field, CQ:relevance), not refused. **8**, undermining
`matters_more(H,G)` restores the target with its other questions open, and independent support stays
in. C3's two tests: a modal argument for `should_be_performed(A)` refused (S9); a bridge with its
premise challenged.

## Order of work

1. This plan and D094, checked independently; D095 put to the owner.
2. C5.1 (it depends on no part of D095); gates on both toolchains; independent review; push.
3. When D095 is answered: the answer scheme for `worth_finding_out_first`, if (3) is taken, through
   the proposal process; then C5.2, one data identity; gates; independent review; push.

## Not in C5

The other five game schemes' changes wait for their families; their fields move to the catalogue in
C5.1. 33.4's `fails_to_avert` and `harm_reduction` wait for a profile that includes them. The
contexts' weights for today's schemes stay as they are (calibrated, plan §6).
