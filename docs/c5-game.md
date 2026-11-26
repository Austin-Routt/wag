# Step C5: the game takes the first family

**Status:** a plan, drafted 2026-11-26, for independent review before work starts. Step C5 of
[first-family.md](first-family.md): the game's own data takes the formulas that proposals
D004-D075 (C2), D008-D074 (C3) and D022-D093 (C4) gave `walton.yml`, and the decisions that name
the NASCQ generator (D007, D009) are applied through one. This document says how, and in what order.

## What C5 must do

From first-family.md's table and the register:

- **D007:** the game uses the book's ch. 9 wording of practical reasoning; ch. 3's further questions
  are kept as derived questions, each with its chapter and page.
- **D009:** the game's question texts come from the catalogue's paraphrases, with the changes the
  entries list; every generated question gets a test that its formula target is one of its
  scheme's, which also covers the catalogue version, the target's role and type, the variable
  substitution, the argument instance, the polarity, the response's effect and its replay. Both
  D007 and D009 say: *to apply: the NASCQ generator and the game's data*.
- **The slice's formulas in the game:** the game's copies of practical_reasoning1 and 2 and of the
  two consequence schemes take `walton.yml`'s (C2's agent P, goal premise, the possibilities
  apart, the scoped exceptions, `bad_outweighs` for the side effect; C3's `too_slight` and the
  bearer exception of each side).
- **D075's backup edges:** the answer schemes `bad_outweighs_good`, `good_outweighs_bad` and
  `better_means` join the game and answer the questions whose exceptions they conclude. D009
  excepts the old edge by which negative consequences answered the side-effect question, which
  D070 and D075 replace; today's file no longer has it. The one edge of the slice it has, by which
  positive consequences answers the opposite-consequences question, goes too: once that question
  asks the exception `good_outweighs(A,G)` (D055), an argument concluding `should_be_performed(A)`
  cannot be its answer (S9), and `good_outweighs_bad` is; a further good consequence stays a
  rebuttal, a move of its own (OC2).
- **D008:** the modal variants and their bridges join the game.
- **New game questions:** for `too_slight`, the bearer pair and `worth_finding_out_first` (C3), and
  D074's question about the argument itself (an undercutter of an argument from consequences
  offered for a claim's truth).
- **Item F, the CQ answer contract:** for each question of the included schemes, a legitimate answer
  accepted; an irrelevant, circular or mismatched one refused; the other challenges and independent
  support surviving; and the refusal of a bare claim of knowability as an answer to
  `worth_finding_out_first` (D053's declaration).
- **The replay of the saved games**, and the game's tests.

## Where the game's data comes from today

The game reads `wag/data/nascq/{schemes,critical_questions,edges}.json` and
`wag/data/schemes/core.yml`, written by hand. The catalogue copies its game parts *from* them
(`internal/catalogue/generate.go`), and keeps only a scheme's game id and backup edges and a
question's game id, text, locus and pattern: too little to write them back (review part 13 §1).
Plan §6 names the fields the game needs and gives each a kind: derived by a rule, judged with
examples, calibrated by play. process.md: the game's data is to be *generated from the catalogue,
never by hand, once generated*.

What the files do in the code (an inventory of 2026-11-26, in the workspace, verify/c5):

- **Every field is read**, except `caes_mapping.target` and, outside a fallback no shipped edge
  takes, the rest of `caes_mapping`. `semantic_type` is read for display and a lint warning only.
- **Order is behaviour.** The order of the ask and backup edges fixes the order of the offered
  moves, so the bots' moves and the 33 golden games (`testdata/reference-digests.txt`). Key order
  in the files is not.
- **Bytes are identity.** The SHA-256 of core.yml and the three NASCQ files is every saved game's
  `data_sha256`; a game exported under another hash replays read-only. The bytes also go,
  verbatim, into annotators' prompts and annotation run identities, and `wag-llm replay` compares
  prompts byte for byte.
- **core.yml is not required to equal `walton.yml`.** The catalogue records each difference as a
  `game-differs` discrepancy (practical_reasoning1 12, practical_reasoning2 13, the consequence
  schemes 3 each, sign 1, analogy 10). Its language block (39 templates and their 39 negations)
  is the game's own: 34 of the 39 differ from `walton.yml`'s.

## Two parts

### C5.1: the generator, with no change of behaviour

1. **The game fields get a home in the catalogue.**
   - In each game scheme's entry, `game` grows from `{id, backs}` to also hold the NASCQ scheme
     record (name, semantic type, thesis independence, base weight, variables with type,
     description and example, premise and conclusion templates, description, cues, answer
     variable, evidence policy and its note) and the game's copy of the formulas (variables,
     premises, assumptions, exceptions, conclusions, meta). Each is recorded from today's files,
     with a status (`recorded`, the date, "from the NASCQ" or "from core.yml"), as plan §6 asks.
   - The NASCQ's questions are shared by several schemes (9 of 24 are asked of more than one), so
     they get one file, `wag/data/catalogue/game/questions.yml`: for each question its name,
     semantic type, template, answers and `caes_mapping`, recorded.
   - The order of the edges is behaviour, so it is data: `game/order.yml` lists the schemes in
     order and, for each, its questions in the order they are asked; then the backup edges in
     order. The generator checks that each listed ask edge is a game question of that scheme's
     entry, and each entry's game question is listed.
   - core.yml's language block and its comments, and the scheme it keeps for games before
     2026-10-13 (`credible_source`), go to `game/core.yml.in`, a template the generator fills
     with the schemes.
2. **The generator.** `wag-cli --game-gen` writes the three NASCQ files and core.yml from the
   catalogue. `--catalogue-check` and a test require the committed files to equal its output, as
   they require the catalogue to equal `--catalogue-gen`'s. The catalogue's check turns round: an
   entry's game parts are the source, and the files are checked against them.
3. **The first output reproduces today's.** A test parses today's files and the generator's
   output and requires them equal: every record and field, and the order of the edges. Then the
   committed files become the generator's output.
   - Its format is the generator's (two-space JSON, no hand alignment), so the bytes change once
     and `data_sha256` with them (from `e7141ec8dea379bc`). Byte identity would need the hand
     alignment of each block stored as data; the meaning, which the test pins, is what the
     reproduction is for. **A choice:** recommended, and reversible by storing the layout.
   - What the one change of bytes touches: saved games under the old hash open read-only (as any
     game under other data does); the golden games must not change, since they depend on the
     meaning and the order, not the bytes (C5.1 shows it); the batch-2 replay's report changes
     only its `data_sha256`; annotation runs record new input hashes from then on.
4. **Its own review and gates.** C5.1 is committed on its own: no game test may change except those
   that pin bytes or the file's text layout, each listed with its reason.

### C5.2: the slice, through the catalogue

Each change is a change of the catalogue's game fields, regenerated; the game's tests follow, each
change of a golden game explained.

- **The four game schemes take the canonical formulas.** core.yml's practical_reasoning1 and 2 and
  the two consequence schemes become `walton.yml`'s; their `game-differs` discrepancies go to 0.
  Their new predicates get templates in core.yml's language (the game's own, as for the others),
  with their negations where the board shows them.
- **The agent P.** Practical reasoning names its agent since C2 (`goal_of_agent(P,G)`,
  `possible_bring_about(P,A)`, the scoped exceptions). The game binds no agent today. **A
  choice:** recommended, the game binds P itself, from the thesis's agent, which a game records
  once, at its start (as the annotation's case record does, D063, D091, D092), not per move; a
  move about another agent's action names that agent, and the board says whose action it is.
  Alternatives: the player binds P on each move (more typing, and a mismatch the game must
  refuse), or the game leaves P out (the game's copy would differ from `walton.yml` again).
- **The questions' patterns** follow the formulas (first-family.md, the table of the slice's
  questions): `CQ:other_goals` asks `incompatible_goal(P,G)` at the exception (D004, practical
  reasoning 1 included); `CQ:side_effect` asks `bad_outweighs(A,G)` (D070), answered with
  `bad_outweighs_good`; `CQ:possibility` asks `possible_bring_about(P,A)` (D088);
  `CQ:other_means` asks `realize_more_effectively(P,G,A)` (D089), answered with `better_means`;
  the consequence schemes' opposite-consequences question asks their exception
  (`bad_outweighs`, `good_outweighs`; D055), answered with C003's answer schemes.
- **New questions:** `too_slight(A,G)` (D065), the bearer exception of each side (D069),
  `worth_finding_out_first(A,G)` of the bridges (D008, D053), and D074's question about the
  argument itself (an undercutter, D074), each with its text from the catalogue's paraphrase
  (D009) and its answers' effects.
- **The answer schemes and the modal variants join the game** as NASCQ schemes, with their
  fields recorded as judged, with examples (plan §6); the backup edges are the decided ones
  (D075: each answer scheme answers the questions whose exceptions it concludes; D070 removes the
  negative-consequences edge for the side effect).
- **D007 and D009 applied:** the question texts come from the catalogue's paraphrases (D009), and
  ch. 3's further questions of practical reasoning are kept as derived questions with chapter and
  page (D007), in the catalogue, whether or not the game asks them.
- **Item F.** For each question of the four schemes, the answer schemes and the modal variants and
  bridges, a test of the contract: the right target, locus and burden; a legitimate answer
  accepted; an irrelevant, a circular and a mismatched answer refused; the other challenges and
  independent support surviving; for `worth_finding_out_first`, a bare claim that the matter can
  be found out refused (D053's declaration: the investigation must be timely, reliable and worth
  its cost). The first-family adverse cases 5 to 8 of the critical review of 2026-11-23 are
  matched to these tests.
- **The replay.** The golden games are replayed on the new data and each change explained; the
  batch-2 replay plays the claims C2 to C4 left unplayed for want of the game's schemes (13292.104
  and .123's better means, 13292.127 and .43's weighings, the modal claims), and its report says
  which now play as intended.

## Order of work

1. This plan, reviewed independently; the choices put to the owner where they are not covered.
2. C5.1: the fields recorded, the generator, the reproduction test, the switch; gates on both
   toolchains; an independent review; push.
3. C5.2, in steps that each keep the game's tests green: the formulas and the agent; the questions'
   patterns; the answer schemes and their edges; the new questions; the modal variants and
   bridges; item F's tests; the replays. Gates and an independent review before the push of each.

## Not in C5

The other five game schemes' changes (sign, analogy, expert opinion, position to know, values)
wait for their families; their fields move to the catalogue in C5.1 all the same. Contexts'
weights (`data/contexts.json`) stay where they are: they are calibrated, not judged (plan §6).
