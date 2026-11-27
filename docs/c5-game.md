# Step C5: the game takes the first family

**Status:** C5.1 is applied (2026-11-26: c9935ce, 17fd22a and the switch). D095 is decided (the
owner, 2026-11-26), and its answer scheme for `worth_finding_out_first` is in `walton.yml`
(proposal D095, applied 2026-11-26). C5.2's steps 1 to 7 are done, the switch, the replay by instance and item F on 2026-11-27; step 8 (gates, an independent review, CI, the docs) is next. The
plan: version 1 (3bc7f0c) and version 2 (3471585) were each checked independently on
2026-11-26 (in the workspace, verify/c5); version 3 put D095 to the owner; version 4 had D095 as
decided and item F's cases from the independent review of 26 November; version 5 gave C5.2's order
of work, its data and the choices the decisions leave to it; this is version 5a, after version 5's
independent check. Its implementation
choices, within the decisions it applies, are D094. Three choices departed from the wording of
decided texts, so they were the owner's: D095. Step C5 of [first-family.md](first-family.md): the game's own data takes what proposals
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
  so one id cannot be both (S9 would refuse `bad_outweighs_good` at the conclusion). D095 (1)
  decides it (below).
- **D009:** the question texts come from the catalogue's paraphrases (all 35 game questions have
  one), with the changes the entries list, among them sign's own likelihood text and the evidence
  question at the premise `brings_about(A,G)`; CQ:field is kept; each generated question has a test
  that its target is one of its scheme's, covering the catalogue version, the target's role and
  type, the substitution, the instance, the polarity, the effect and the replay. **D007:** the
  game's practical reasoning is ch. 9's; ch. 3's further questions are derived questions in the
  catalogue, with chapter and page. **D006:** CQ:other_means asks for a better means.
- **D008:** a bridge is an entry condition, not a weight. **D053:** a bare claim that a consequence
  could be investigated does not establish `worth_finding_out_first` (the owner's declaration).
- **D095** (the owner, 2026-11-26), which departs from C2's and C3's game paragraphs where they
  differ: (1) CQ:side_effect is the rebuttal route (a further bad consequence, at the conclusion) for
  practical reasoning and positive consequences, as CQ:opposite_consequences is for negative
  consequences; a new weighing question asks `bad_outweighs(A,G)` of practical reasoning and
  positive consequences (one id: one formula, locus and burden), and another asks
  `good_outweighs(A,G)` of negative consequences, both backed by C003's answer schemes; an id that
  answer schemes back is asked at the same formula and kind of locus wherever it is asked, and the
  generator refuses anything else. (2) D074's objection is Other's relevance question, unchanged,
  with the direct undercutter of the particular inference; the consequence schemes get no question
  for it, and direct counterarguments stay available. (3) An answer scheme concludes
  `worth_finding_out_first(A,G)` from the owner's declared conditions, proposed for `walton.yml`
  through the proposal process; at that question the game accepts it and the source-based schemes
  that assert the exception, each with its own questions, and refuses Other, well made or not: a
  rule of WAG's first profile. (4) D093 (e), (f) and (g) are confirmed, with tests for a wrong
  actor, action, goal and context, unrelated consequences, and surviving independent support.
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

One change of data identity for all of it (D094 a). Version 5 of this plan gave C5.2's order of
work, the data at the switch, and the choices the decisions leave to the implementation; version
5a corrects it after its independent check (in the workspace, verify/c5/c52: 5 errors, 15 warnings
and 13 notes, each verified), and after the check of version 5a and the profile's code (5 errors, 9
warnings and 10 notes, each verified, VERIFY-5a). The choices no decision makes are recorded as a
refinement of D094, the developer's under the owner's mandate as D094 is; four questions go to the
owner, and the plan keeps the decided text until they are answered, but for question 2, where the
game applies the refinement's reading of D063 meanwhile (reversible).

### Order of work

As C5.1: the code first, with the data's bytes unchanged and each commit gated on both toolchains;
then the switch, one commit that changes the data's identity once; then the review.

1. **Roles** (D094 (g); done: 5e6ca71). A game scheme's NASCQ record names its role: an argument
   (no role, as today), an answer, a modal variant or a bridge. An answer scheme is offered only on
   its backup edges (S1 refuses it as a move of its own or about a statement) and weighs as the
   scheme of the argument it answers. A modal variant names the scheme it weighs as (`weighs_as`, its
   categorical counterpart) and a bridge names its modal variant, so both weigh as the counterpart,
   with no multiplier (D008); a context that sets the counterpart's weight sets theirs, in play and
   under another audience's weights (ClaimUnder). A bridge names its entry premise (`entry_premise`,
   the modal conclusion it carries), which the move does not presume: a modal argument, a move of
   its own made before or after, supports it. A question record can refuse Other (`refuse_other`,
   D095 (3)): S5 refuses an answer made with it, well made or not, before S9, and the question does
   not offer it; Other's move about the exception, or against its negation, is refused too
   (920adca). Lint refuses a role the game cannot play. Tests build their own small data in memory.
2. **The game's profile** (D094 (b); done: 920adca; below).
3. **D003's question protocol** written in game-rules.md as the game's rule (done: 5e6ca71): who may
   ask what, what a question does while it is pending, who answers it and how, when it is closed,
   and what may answer. It formalizes today's behaviour; no golden game moves.
4. **The generator's rules** (`wag-cli --game-gen`), each refusing a catalogue that breaks it, with
   a test that plants the fault (written for version 5 at 3373558, then for version 5a, which checks
   one formula and one text for every id, generates the text and adds (d) and (e); applied with the
   switch, since today's data breaks them): (a) one id is one question: every scheme that asks it asks it at the same
   formula and kind of locus (D095 (1) for a backed id; for every id by the refinement of D094); (b)
   its record's `caes_mapping` is derived from that locus (premise: undermining; exception or
   argument: undercutter; conclusion: rebuttal), so the effect shown is the edge's, and
   `target_premise_pattern` goes (the edges carry the pattern); (c) a game question's text is its
   entry's paraphrase with the scheme's variables braced, generated, not written: a paraphrase of a
   game question uses a variable's letter only for the variable (so "A grave consequence" becomes
   "One grave consequence"); (d) a role scheme's record agrees with its entry: an answer scheme
   concludes the formula of each question it backs, a modal variant's and a bridge's `weighs_as` is a
   game scheme, a bridge's `entry_premise` is one of its premises; (e) the templates of the predicates
   the game plays are walton.yml's (D094 (f)), which becomes an input of the generator, or a
   difference is recorded once, with its reason, in `game/`; new predicates get their negation
   templates written, as core.yml has one for every predicate. Done with the switch: the rules
   hold on its data, and each refuses its planted fault.
5. **The switch** (the data below): the catalogue's game parts, `game/questions.yml`,
   `game/order.yml` and `core.yml.in` changed, and the generated files with them, in one commit: the
   data's fingerprint changes once. In the same commit: the gold recoded where validate.py then
   refuses it, validate.py checking each challenge's `wag_cq` against the edges as it checks a role's
   `cq` (four records carry CQ:side_effect at `bad_outweighs(...)` and become CQ:bad_outweighs), and
   re-pinned, both copies; the frozen copy of the pre-C5.2 files (`wag/data/frozen/<fingerprint>`),
   as C5.1 did; the docs say how an old game or trace replays. Done (2026-11-27): the fingerprint
   went from `afd3335a23f12f41` (frozen, 1273555) to `092b6a6f48a672d2`. The gold's recode
   (`recode_c52.py`) found more than the four records: six roles named the book's question, or
   none, where the game now asks one (13292.25 now answers CQ:too_slight; 13292.12, .233 and
   13292.180's second function name CQ:modal_likelihood, 13292.127 CQ:bad_outweighs, 13292.43
   CQ:good_outweighs), and seventeen challenges named no game question where the game asks one; the
   ten instances of the schemes the game now plays name its scheme. `validate.py` checks a
   challenge's `wag_cq` both ways (a null where the game asks one is refused) and an instance's
   `wag_id`, and checks a format-1 file against the NASCQ it was made with
   (`wag/data/frozen/e7141ec8dea379bc`). The guidelines and the rule cases name the game's
   questions where it asks them. The batch-2 replay outputs stay those of the data before until
   step 6.
6. **The replays.** The 33 golden games stay on the frozen pre-C5 data and keep their digests (D094
   (h)): they are computer games without a profile, so the profile's rules do not apply to them, as
   to any record made before step 2. A new golden set is made on the new data, with a profile,
   covering each new route. `to_wag.py` plays the batch-2 gold by instance (a claim's secondary
   instances, a bridge with its modal argument), with the case's profile (done for the profile:
   the replay is unchanged, claim for claim); the batch-2 report is made again and names the claims
   C2 to C4 left unplayed that now play. The replay by instance done (2026-11-27): it reads the
   game's schemes from the server's data and sends a better means' comparison; on the new data 46 of
   49 claims are played (37 before) and 31 as intended (17), the three bridges with their modal
   arguments. An answer coded with Other is played through its question, with Other (the check of
   the switch, N-15). On the frozen data the script plays as the committed report of that data did,
   but for three such answers (13292.33, .182 and .198, now played through their question, as
   intended) and the `intended` fields of the five claims the recode changed (the recoded gold's). The new
   golden set is written with item F (step 7), whose cases are its routes. Old traces with model players replay with the pre-C5.2 build
   and `--data` with the frozen files; their prompts carry no profile section.
7. **Item F** (below), with the review's eleven new cases and the check's two, each outcome written
   first and paired with a planted defect. Done (2026-11-27): the cases and their outcomes in
   [`c5-item-f.md`](c5-item-f.md), committed before any test of them ran (5b84dc2); the tests
   (`itemf_test.go` on the shipped data, the engine's and `validate.py`'s for the schemes the game
   does not play yet) and 25 planted defects, each caught. The first runs found two outcomes the page
   had wrong (A4.10's board, B4's check) and one the game had wrong: a value that names nothing was
   accepted as a name (`x_empty`), for an answer's H or I, a consequence's G, and an action's A2 in a
   game without a profile; S2, P1 and the adapter now refuse it, and the terminal asks again. The
   independent check of step 7 (in the workspace, `verify/c5/c52/CHECK-step7.md`) found B6 testing
   nothing (played in the wrong order), punctuation and invisible values still binding, old records
   with such a value failing to load without a reason, and the golden set covering few new routes;
   each is fixed in the commit after b3b10ef. The new golden set: the reference games' table on the
   shipped data with the default profile, and a scripted game for each new route
   (TestShippedReferenceDigests, TestShippedReferenceRoutes), each file's first line pinning the data
   and the profile; the move digest now renders declarations and comparisons.
8. **Gates on both toolchains, an independent review, CI, and the docs** (README, prototype,
   game-rules, protocol.md, llm.md, the catalogue README).

### The game's profile

D094 (b), applying D089, D091 and D092; D063 says why no agent may pass unchecked. "The game's
profile" is this record; "WAG's first profile" in D095 (3) is the rule set it belongs to.

- **What it holds:** the agent (an actor), the actors (id and text), the actions the game
  deliberates (id, actor and aliases), and the situation the alternatives are compared in (a text,
  D089: the current one, fixed for the game). `GameConfig` carries it; the request recorded with the
  game holds it, so an export, a trace, a loaded game and a replay carry it with the moves.
- **Where it comes from:** a case record (`--case case.json` for wag-cli and wag-llm, the board's
  new-game form, `"case"` in `POST /api/games`; `to_wag.py` sends the batch's): its actors and
  actions, its agent the thesis action's actor, its situation the default context's text. Otherwise
  the agent and situation given when the game is created, or else the default the board shows: the
  agent `decider` ("the one who decides whether to take the actions argued about"), no action yet,
  the situation "the present one". A game's claim that recommends an action declares it as the
  agent's when the game is made; one that recommends another actor's declared action is refused.
- **Declaring an action:** a move that first argues about an action declares it (`declare`: its id,
  its actor, which every declaration names, D094 (b), and its aliases), recorded with the move and
  joining the profile only if the move is accepted; a move declares only actions it binds. The board
  and the terminal ask whose a new action is, the agent's unless the player chooses another actor;
  computer players declare the agent's actions they name; a model player's prompt lists the declared
  actions, the other actors' too, and the field. A re-declaration with the same actor adds only its
  aliases; with another actor it is refused. Names and actors are compared as atoms.
- **The rules** (game-rules.md), P1 (the declarations) and P2 before NASCQ's S0 to S9, the rest after
  them, each refusal saying which: **P1** an action variable is bound to a declared action, by id or
  alias, names compared as the board writes them (an alias reads as its action), the values a target
  fixes included; **P2** an agent variable is the agent, bound by the game; **P3** the actions argued
  are the agent's (D063: a recommendation names no agent but the one addressed; the refinement of
  D094 records this reading, which goes beyond D094 (b)'s "a P that does not perform A", and the
  owner may narrow it: question 2); **P4** a `better_means` answer's A2 is another of the agent's
  actions, not A or its alias, and the move records its comparison (surely, sooner or at lower cost;
  D089). Its P, G and A are the answered exception's (S9 unifies them).
- A game made before step 2 has no profile and plays without these rules. The situation is shown and
  recorded, not checked move by move (D089).

### The data at the switch

Formulas: walton.yml's, the four game schemes' overrides removed (an entry keeps only its meta in
`game.core`). New predicates take walton.yml's templates, with negation templates written; the
game's other templates stay, each difference recorded once (rule (e)); a template no scheme uses,
the schemes core.yml keeps for old games included, leaves core.yml (old games replay with the
frozen files).

| Game scheme | walton.yml | Role, weighs as | Notes |
|---|---|---|---|
| AS:practical_reasoning | practical_reasoning1 | argument | gains P (type agent) and `goal_of_agent(P,G)`; `possible_bring_about(P,A)` replaces `possible(A)` (D088); the scoped exceptions (D004, D089), `bad_outweighs(A,G)` replacing the side effect (D070, D091) |
| AS:practical_reasoning_goal | practical_reasoning2 | argument | the same, with `possible_goal(G)` |
| AS:consequences | positive_consequences | argument | gains `bad_outweighs`, `too_slight`, `gain_discounted_by_bearer` |
| AS:negative_consequences | negative_consequences | argument | gains `good_outweighs`, `too_slight`, `harm_discounted_by_bearer` |
| AS:consequences_modal (new) | positive_consequences_modal | modal, AS:consequences | concludes `may_be_worth_performing(A,G)` |
| AS:negative_consequences_modal (new) | negative_consequences_modal | modal, AS:negative_consequences | concludes `may_be_unwise_to_perform(A,G)` |
| AS:bridge_possible_good (new) | bridge_from_possible_good | bridge, AS:consequences_modal | entry premise `may_be_worth_performing({A},{G})` |
| AS:bridge_possible_harm (new) | bridge_from_possible_harm | bridge, AS:negative_consequences_modal | entry premise `may_be_unwise_to_perform({A},{G})` |
| AS:bad_outweighs_good (new) | bad_outweighs_good | answer | C003, D064 |
| AS:good_outweighs_bad (new) | good_outweighs_bad | answer | C003, D064 |
| AS:better_means (new) | better_means | answer | D089; P4 (`distinct` A2 and A; `comparison`) |
| AS:finding_out_first (new) | finding_out_first | answer | D095 (3) |

A modal variant's or bridge's record carries the base weight of the scheme it weighs as (what is
shown before a move is the weight that applies); an answer scheme's carries 0.5, never used, since
it weighs as the argument it answers. Types: P `agent`; A and A2 `action`; better_means' G `goal`
(the practical reasoning's); H and the consequence schemes' G `event`; I `entity`. The three
answer schemes without a question of their own record D075's reasoned none in their entries, as
finding_out_first's does, and TestCoreDataLoading allows it.

The other five game schemes (values, sign, expert opinion, position to know, analogy) keep their
formulas; D009's listed changes reach them: their question texts (below), the values scheme
described in the book's direction (a value favours a goal), and sign's premise keeping "generally".

| Question | Asked of | Locus and formula | Backed by |
|---|---|---|---|
| CQ:other_goals | practical reasoning | exception `incompatible_goal({P},{G})` (D004) | — |
| CQ:conflicting_goals | practical_reasoning_goal | exception `incompatible_goal({P},{G})` | — |
| CQ:other_means | both practical reasonings | exception `realize_more_effectively({P},{G},{A})`, D006's text | AS:better_means |
| CQ:best_means | both | exception `bring_about_more_effectively({P},{S2},{A})` | — |
| CQ:possibility | both | premise `possible_bring_about({P},{A})` (D088) | AS:sign |
| CQ:goal_possible | practical_reasoning_goal | premise `possible_goal({G})` | AS:sign |
| CQ:intervening_actions | practical_reasoning_goal | exception `intervening_actions({A},{G})` | — |
| CQ:side_effect | both practical reasonings (game-only questions), AS:consequences | conclusion (the rebuttal route, D095 (1)) | AS:sign |
| CQ:bad_outweighs (new) | both practical reasonings (22.4/CQ7, 22.1/CQ5), AS:consequences, AS:consequences_modal | exception `bad_outweighs({A},{G})` | AS:bad_outweighs_good |
| CQ:opposite_consequences | AS:negative_consequences | conclusion | AS:consequences |
| CQ:good_outweighs (new) | AS:negative_consequences, AS:negative_consequences_modal | exception `good_outweighs({A},{G})` | AS:good_outweighs_bad |
| CQ:too_slight (new) | both consequence schemes and both modal variants | exception `too_slight({A},{G})` (D065, D053) | — |
| CQ:gain_discounted (new) | AS:consequences, AS:consequences_modal | exception `gain_discounted_by_bearer({A},{G})` | — |
| CQ:harm_discounted (new) | AS:negative_consequences, AS:negative_consequences_modal | exception `harm_discounted_by_bearer({A},{G})` | — |
| CQ:likelihood | both consequence schemes | premise `brings_about({A},{G})` (33/CQ1) | — |
| CQ:evidence | both consequence schemes | premise `brings_about({A},{G})` (33/CQ2; D009) | — |
| CQ:modal_likelihood (new) | both modal variants | premise `may_bring_about({A},{G})` (33/CQ1's modal paraphrase) | — |
| CQ:modal_evidence (new) | both modal variants | premise `may_bring_about({A},{G})` (33/CQ2's modal paraphrase) | — |
| CQ:worth_finding_out_first (new) | both bridges | exception `worth_finding_out_first({A},{G})`; refuses Other | AS:finding_out_first |
| CQ:correlation (new) | AS:sign | premise `strong_correlation({A},{B})` (sign's own likelihood text, D009) | — |
| CQ:honesty (new) | AS:position_to_know | exception `dishonest({W})` (split from CQ:trustworthiness, which expert opinion keeps at `untrustworthy({W})`) | — |

The other questions keep their ids and edges, with their entries' paraphrases as texts. CQ:field
and CQ:opinion stay shared by expert opinion and position to know, each with one paraphrase
written for both. `matters_to_decision(A,G)`, the bridges' other premise, is challenged as any
premise is (CQ:premise_acceptable). D074 is Other's relevance question with the direct undercutter:
no data. D007's derived questions go into the practical reasoning entries as derived questions
(ch. 3's Other-Goals question, which asks which goal comes first, not which conflicts), each with
its chapter and page; the game does not ask them. The entries' game/CQ:evidence items go, 33/CQ2
taking the game part. Every question whose text changes has its answer words written again (some
new texts ask "what" or "who", not yes or no), with a test that each question has them.

### Choices within the decisions

- **One id is one question** (the refinement of D094): every scheme that asks an id asks it at one
  formula and one kind of locus, with one text. Where entries' paraphrases of one id ask of one
  formula at one locus in different words, they are given one paraphrase (drafted text, with a
  history line): CQ:bad_outweighs (which D095 (1) makes one id), CQ:side_effect, CQ:field (which
  D009 keeps), CQ:opinion, CQ:too_slight (one formula, declared neutral in polarity, D053). An id
  over two formulas is split, the scheme whose book question named it keeping it: CQ:likelihood
  (sign's becomes CQ:correlation, as D009 asks) and CQ:trustworthiness (position to know's becomes
  CQ:honesty).
- **No backup edge from negative consequences to CQ:side_effect**, as D009 says (D009 predates the
  bridges; the harm bridge has none either). D009's reason, D070's weighing reading of the side-effect question, has lapsed with D095 (1),
  whose rebuttal route is "as CQ:opposite_consequences is", which positive consequences backs: the
  owner decides (question 1). Meanwhile a further bad consequence argued by negative consequences
  is played as a direct rebuttal of the conclusion.
- **Sign's three backup edges stay**; on practical reasoning the side-effect route moves to the
  conclusion, so sign's answer there becomes a rebuttal, as it is on positive consequences already.
  Whether a sign fits a practical conclusion is for sign's family (question 3).
- **CQ:evidence moves to the premise** `brings_about(A,G)`, as D009 adopted: a premise question,
  the proponent's to answer, beside CQ:likelihood; one argument for the premise answers both.

### Questions for the owner

1. A backup edge from negative consequences to CQ:side_effect, the mirror of positive consequences
   at CQ:opposite_consequences (D009 excluded it for a reason D095 (1) has removed)?
2. P3: the actions argued in the game are the agent's (from D063), or only those of a scheme with an
   agent place (D094 (b)'s words)?
3. Sign's answer at practical reasoning's rebuttal route: kept, or left to sign's family?
4. Other's direct undercutter of a bridge whose premise is only that the consequence could be found
   out: D095 (2) keeps direct counterarguments available, while D095 (3) and D053 refuse a bare
   claim of knowability at the question. The game accepts it meanwhile (Other cannot conclude the
   exception, nor presume it, VERIFY-5a E-b); should the refusal reach it? Likewise sign and analogy,
   which no backup edge offers at that question but which can argue the exception by a move about
   it, as any statement can be argued (the check of the switch, N-12): should D095 (3)'s list (the
   answer scheme and the source-based schemes) bound every route to the exception?
5. A wording: CQ:worth_finding_out_first's text, from proposal D008-D074 (a decided text), ends
   "Which way?", the investigation its answer names (D095 (3)), while the board offers "yes" and "no"
   (the check of the switch, N-4). Keep it, or end it "If so, how?"

## Item F: the answer contract, per route

Per question and answer route (D075; the critical review's row F): each backup edge (the answer
schemes, four with D095 (3)'s; sign's three; positive consequences' rebuttal edge); the
thesis-independent schemes, which back every question; Other; a question asked alone; the direct
rebuttal where an answer route is removed. Each expected outcome is written before it runs and paired
with a planted defect the test must catch:

- **Accepted:** the legitimate answer, at the right target and locus, with its burden.
- **Refused, with the rule that refuses it first** (the profile's P1 declarations and P2 before
  NASCQ's S0 to S9, S5 before S9, Q1; P1's values, P3 and P4, the A2 rule, after them; the cycle
  check, S5 for Other at a refused exception and G1 once the move is placed), leaving the graph, the open questions and the journal unchanged: an
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
premise challenged. For `finding_out_first` (proposal D095's check): an answer whose I is unbound is
refused (S2); a condition challenged and not made out (each of its three is a premise of its own)
restores both bridges. From the check of version 5: Other's move about the exception, or against
its negation, is refused as its answer at the question is (done, 920adca); one argument for
`brings_about(A,G)` answers both CQ:likelihood and CQ:evidence; F-continuing, F-context and
G-endandmeans below concern schemes the game does not play in C5.2 (`bridge_from_continuing`,
`bridge_from_agent`, `end_and_means_bad`), so they are engine and validate.py tests.

The independent review of 26 November (in the workspace, reviews/2026-11-26 and
verify/r-external-2026-11-26) proposed 22 cases for D095. Ten are the cases above (its Q1-A, Q1-D,
Q3-A, Q3-B, Q3-D, Q3-E, Q3-F, E-agent, Replay-1, Replay-2). Eleven are added here: **Q1-B** and
**Q1-C**, each route of each consequence scheme and the shared weighing id, each tested; **Q1-E**, a
slight harm supports a con argument at the conclusion but does not establish `bad_outweighs(A,G)`;
**Q2-A**, D074's undercutter of an argument that infers a factual claim from desirable consequences;
**Q2-B**, with a second, legitimate argument about the same action, which the undercutter leaves
standing; **Q2-C**, independent support of the factual claim survives the undercutter; **Q3-C**, a
source-based answer at `worth_finding_out_first` is challenged in play through its own questions
(adverse case 7); **F-continuing**, the better-means exception of the continuing bridge;
**F-context**, a bridge that carries `should_do(P,A)` where A is another actor's is refused;
**G-endandmeans**, `end_and_means_bad`'s mirrored weighing in the right direction; **G-unbound**, a
weighing exception with an unbound X or Y is refused, not read as a wildcard. The twenty-second,
Recording-1 (a paid reply recovered after a crash, with no second call), belongs to G, the journal.

## Order of work

1. This plan and D094, checked independently; D095 put to the owner.
2. C5.1 (it depends on no part of D095); gates on both toolchains; independent review; push.
3. D095 decided (2026-11-26); the answer scheme for `worth_finding_out_first` applied the same day through
   the proposal process (proposal D095, version 2a); then C5.2 in the order above (version 5 of this plan,
   checked independently; version 5a after the check), one data identity; gates; independent review;
   push.

## Not in C5

The other five game schemes' changes wait for their families; their fields move to the catalogue in
C5.1, and in C5.2 only D009's listed changes reach them (their question texts, the values scheme's
description and sign's "generally"), with the refinement of D094's split of the questions shared over
two formulas. 33.4's `fails_to_avert` and `harm_reduction` wait for a profile that includes them. The
contexts' weights for today's schemes stay as they are (calibrated, plan §6).
