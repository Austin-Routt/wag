# WAG review, part 2 (2026-10-13)

**Scope.** This part covers four new resources:

- `CLI_Reviews.md`: three AI reviews of CLI sessions;
- `SPR.md`: a sparse priming representation of Carneades-4;
- `Argumentation_Game.md`: game designs, ending with System Reference Document (SRD) v3.0;
- Pyon & Ri (2022), *A Network of Argumentation Schemes and Critical Questions*, Informal Logic 42(4).

It also answers your ten points about part 1 and describes the changes made in response. Commits `5e5258c`..HEAD on `claude/determined-wozniak-3v3h6a`; the rulebook is `game-rules.md`.

## Short answers

| Question | Answer |
|---|---|
| Are the resources helpful? | **Yes: the NASCQ paper most, then the CLI reviews, then the SRD.** Use the SPR only as an index: about half of its specific claims about Carneades are wrong (section 1.2). The paper settled the design of defence. The reviews' observations were accurate, but most of their fixes were not (section 1.3). The SRD is a good catalogue of options, with two defects that would break the game if implemented as written (section 1.4). |
| Do I understand the game? | Yes. Section 2 states it in one page, and `game-rules.md` is the rulebook of the code. |
| Are Carneades-4 and the NASCQ used correctly? | **Carneades: yes, with two deliberate departures.** Presumptions replace assumptions, because an assumption can never be undermined. Some answers are enthymemes (section 3.1). **NASCQ: the structure is used as the paper defines it; the semantics was incomplete.** The NASCQ does not say what part of an argument a question is about, or what an answer concludes. I added both (`locus`, `pattern` and `answer_variable`), and they need your review (section 3.2). |
| Did the system improve? | Yes. There are defence moves at any depth, questions that shift the burden of proof, critical questions for all seven schemes, eight game modes, live scoring, a central claim with sides and burden of proof, solo play, resignation and clocks. The CLI, the wizard, the agents and the JSON view were updated for all of it. There are 152 tests, all passing; `-race` and staticcheck are clean (section 4). |
| Is more needed? | Yes, mainly a domain review of the new data. Section 7 lists it. |

## 1. The resources, critically

### 1.1 Pyon & Ri (2022): the most useful

The paper gives the protocol what it lacked: an account of both answers to a critical question.

- **Polarity is fixed by the question, not by its grammar.** "The positive answer to a critical question is one that supports an argument fitting the scheme … the negative answer is one that attacks" (pp. 792–793). For example, "no, there is no third factor" is the positive answer. The code had treated only the attacking side as a move. Defence was impossible (part 1, section 2.2).
- **An answer is an argument.** "The argument consists of an answer to the question as the conclusion and support for the answer as the premises" (p. 789). The backup scheme therefore provides the premises, and the answer is the conclusion. That is what `Adapter.AnswerQuestion` now builds.
- **Positive answers are premises; negative answers rebut, undercut or undermine** (p. 793). The paper does not say which part of the argument a question concerns. Gordon, Prakken & Walton (2007) do: questions correspond to the scheme's premises, assumptions and exceptions. That gap is why the NASCQ data needed a locus per ask edge.
- **Thesis-independent schemes back every question** (§4). The code already did this.
- **Existence questions** such as "are there other goals?" or "other means?" are hard to answer positively, because the answer is a universal negative. The paper recommends asking specific questions instead: "is G′ preferable?" (pp. 794–795). Our practical-reasoning exceptions come from the Carneades YAML and are existential (`realize_more_effectively(G,A)`). A defence of them has to deny the existence of any better means. It is legal and is decided by weight, but you should decide whether to make these exceptions specific (section 7).

Limits of the paper for our purposes:

- It defines no game: no turns, scoring or termination.
- Fig. 4 is a network of 29 schemes and 39 questions. We implement 7 schemes and 20 questions.
- It leaves the mapping to a formal model (premise, assumption, exception) open.

### 1.2 SPR.md (Carneades-4): unreliable in the details that matter

Its overview (two engines, formats, proof standards, weighing functions) is fine. Its specific semantic claims often contradict the code, and WAG depends on exactly those. I re-checked these in the engine today:

- **A tie on an issue does not leave the positions Undecided.** `Issue.Resolve` makes every position Out unless one wins. Both sides of an equal-weight dispute lose.
- **Undercutters are not weighed.** `GetWeight` returns 0 for an argument whose undercutter is In, whatever the weights.
- **`InstantiateScheme` reports bad parameters only on standard error** (`fmt.Fprintf(os.Stderr, …)`). It returns no error.
- **The proof standards have fixed constants.** PE compares weights only. CCE needs a margin above 0.5. BRD also needs the loser below 0.3. The SPR's "SE" standard does not exist in carneades-4.

Earlier in this session, the code check also found the following. Your fork does not touch them, and WAG avoids them:

- the YAML export drops the theory;
- a second `Infer` call panics;
- `Labelling.init` re-keys the statements map.

**Verdict.** The SPR is a useful index of what exists, but not a reference for how anything behaves. The SRD inherits some of its errors (section 1.4).

### 1.3 CLI_Reviews.md: accurate observations, unsound fixes

The three reviewers (Claude, DeepSeek, Gemini) watched real CLI sessions. Their observations hold. Status of each one on this branch:

| Observation | Status |
|---|---|
| Exception arguments offered as attack targets | fixed in part 1 |
| Argument shown without its premises | fixed in part 1 |
| "Turn N" counts rounds | fixed: "Round N" |
| Atoms and gibberish (`carbon_tax`, "a1 is undercut.") | fixed: `GameState.Say` renders statements through the Language templates in the words typed, and quotes nested statements |
| Internal scheme IDs shown | fixed: NASCQ scheme names |
| Support makes disconnected islands | fixed for answers: defences and attacks conclude on existing statements, and answer variables tie premises to the answer. A plain argument is still a new island by design: it is how a claim enters. |
| Nobody answers critical questions | fixed: question move, defences, obligations in cq mode |
| Greedy bot exploits, multiple-argument scoring | addressed: live scoring counts each statement once per player; attacks on defeated or own arguments score 0; G1 forbids restating |
| A rebuttal that loses scores +0.00 | intended in move scoring. Under live scoring it earns its points if it later prevails. Rewarding "tension" regardless of outcome would pay for spam. |
| Forest display (trees per claim) | **not done**: the board is still a flat list |

Their proposed implementations were mostly unsound:

- **Claims that are false.** For example, "Carneades doesn't set IsUndercutter", an off-by-one in `advanceTurn`, and an `//go:embed ../../` path, which Go does not allow.
- **Thesis modes that break the engine.** They mutate the labelling, or add the claim as an assumption, which makes it unattackable.
- **Defend templates that conclude the scheme's own statement**, not the attacked premise. That was the original defect.
- **Vote-counting labellers that bypass Carneades.**

Their delta-scoring proposal was sound and is implemented as live scoring.

### 1.4 Argumentation_Game.md and SRD v3.0: a good catalogue with two traps

**Useful.**

- The three modes: Moderated, Trusted, Headless.
- CQ-directed support: "the new argument's conclusion is the positive answer to a specific CQ" (§12.2).
- The scoring models `snapshot | delta | end_game_only` (§16.1).
- The proleptic bonus (§16.4).
- Declared sides and consistency (§14).
- The termination precedence (§17).

**Implemented from it.**

- Snapshot scoring (as `move`) and delta scoring (as `live`).
- The proleptic bonus, optional.
- CQ-directed support, as defences.
- Sides, as the claim objective.
- The round, timer and abstention conditions.

**Two defects that would break the game if implemented as written.**

1. **§10.3 says a central argument with no premises is "unapplicable", so the claim starts Undecided.** In Carneades, an argument with no premises is applicable: it has no premises that could fail. With linked weighing it weighs 1, so the claim starts In, at full weight.
2. **§17.3 ends the game when the proof standard is met.** Under PE that happens as soon as pro's best argument outweighs con's best. Con's best is 0 before con has moved, so the game ends right after the opening argument. The verdict should decide the winner when the game ends. This implementation does that, and ends the game by a concession when the losing side has nothing more to say (`ConcedeOnPass`).

**Smaller issues.**

- The SRD's delta rule floors points (`floor(after) − floor(before)`), which loses fractions.
- It credits each argument separately, so restating a statement in new words multiplies points.
- It keeps the points of an argument that is out of force whenever its conclusion is In thanks to someone else.

Live scoring avoids all three.

## 2. The game, in one page

Two or more players build a Carneades argument graph from instances of Walton's schemes. A move is one of the following:

- **argue**: put forward a new argument; its premises are presumed;
- **attack**: give a negative answer to a critical question about an argument, backed by a scheme;
- **defend**: give a positive answer, backed by a scheme;
- **question**: ask without answering, which shifts the burden of proof to the proponent;
- **pass**;
- **resign**.

The NASCQ decides which questions apply to which arguments, which schemes may back their answers, and what part of the argument each question is about. Carneades labels every statement after every move.

An answer is itself an instance of a scheme, open to its own questions. So the dispute deepens along chains of questions. Each answer protects the premise above it until it is knocked out, and anything left unanswered when a question is pressed falls back by presumption.

The mode decides what winning means:

- the highest score, revised after every move or fixed when each move is made;
- the verdict on a central claim, which pro must carry;
- surviving the opponent's questions.

It also decides when the game ends. Solo play is one player arguing both sides; the verdict is their revised belief.

## 3. Are Carneades-4 and the NASCQ used correctly?

### 3.1 Carneades-4

| Use | Correct? |
|---|---|
| Grounded labelling after every move; invariants re-label three times to check stability | yes |
| Exceptions: an exception argument from E to `undercut(a)`, no scheme, so linked weighing (1 when E is In) | yes, the same structure `InstantiateScheme` builds |
| Rebuttals: `p` and `¬p` as an issue, PE by default, ties Out | yes |
| Claim standard CCE and BRD | yes, with Carneades' fixed α = 0.5 and β = 0.3. Base weights are 0.55–0.70, so under CCE any rebuttal that stays in force keeps the claim from being In: con wins unless pro defeats it. That is a heavy burden; use it on purpose. |
| Answers as scheme instances with `Parameters`, so their own questions can find their premises | yes |
| **Departure 1:** premises are *presumed* (an argument of weight 0.01), not *assumed* | Deliberate. A Carneades assumption is In unconditionally, so it could never be undermined. A presumption loses to any argued answer. A question ties it (0.01 against 0.01), so the premise is Out until backed, as an assumption is once questioned in Gordon, Prakken & Walton (2007). |
| **Departure 2:** a backup scheme without an answer variable (consequences, values, practical reasoning) still concludes the answer | An enthymeme: the scheme's own conclusion (`reason_retaining_goal(V,G)`) is the unstated step to the answer. Pyon & Ri's backup edges for these schemes mean chains of arguments. Modelling the chain explicitly would add a linking argument per answer. Say if you want it. |
| Undercutters unweighed | This is Carneades' semantics. WAG gives the proponent a weighed reply instead: denying the exception (`¬E`, an issue) or the undercutter (`¬undercut(a)`). Weighing undercutters themselves would need a change to the fork. I don't recommend it. |
| Avoided: CHR inference, YAML export, factorized weighing | Not used. |

### 3.2 The NASCQ

**Structure.** The code follows the paper:

- AS and CQ nodes;
- ask edges (CQ → AS) and backup edges (AS → CQ);
- thesis-independent schemes back every CQ;
- S5 admits only an explicit backup edge or a thesis-independent scheme;
- S4 (semantic types) became a data lint check, because after S5 it can never fail a move.

**Semantics.** Your question 8 asked: "aren't all CQs and their linkages defined in the NASCQ?" The linkages are, but **the NASCQ does not say two things the engine needs.** These are now data, checked by `wag-cli --lint`:

1. **The locus of each ask edge**: premise, exception, conclusion or argument, and for a premise or exception the formula, as a pattern over the scheme's variables. For example:
   - possibility → the premise `possible({A})` of practical reasoning;
   - trustworthiness → the exception `biased({W})` of expert opinion.

   The lint check fails if a pattern names no premise, assumption or exception of the Carneades scheme.
2. **The answer variable of a scheme**: the variable that states its conclusion (S, S, A, B for expert opinion, position to know, analogy, sign). The lint check fails if it is not what the Carneades scheme concludes.

**Coverage.**

- 20 critical questions and 25 ask edges.
- Every one of the 7 schemes can now be challenged, so the four lint warnings of part 1 are gone.
- Every question has backups: the three thesis-independent schemes, plus five explicit backup edges.

## 4. What changed

| Commit | Change |
|---|---|
| `5e5258c` | Ask edges name the locus of their question. Lint checks the loci against the Carneades schemes. |
| `09c3782` | Answers attack or defend, at any depth (`AnswerQuestion`). Answer variables. Critical questions for every scheme. Attacks on defeated arguments score 0. Optional proleptic bonus. Distinct IDs for players with the same name. The greedy agent is capped at 40 templates a turn. |
| `20f7a81` | Game modes: claim objective with sides and burden of proof, live scoring, question move with obligations, resignation, target score, game and move clocks, `Outcome`, `Simulate` snapshots, invariants for live scores and questions. |
| `62ed85f` | CLI `--mode` and per-setting flags. Wizard with Defend, Question and Resign. Board and JSON view in players' words. Agents for every mode. Runner by outcome. |
| this commit | No undercut bonus under live scoring (section 6). `--alternate`, `--undercut-bonus` and `--proleptic-bonus` for the CLI. A resigned player owes no answers. `game-rules.md`, this report, the README. |

**Validation added.**

- A four-level defence chain (`TestDefenceInDepth`).
- Exception denial: stronger, equal and weaker (`TestExceptionDefence`).
- Live scoring, step by step.
- Every mode's termination and outcome.
- Simulate restoring all mode state.
- The wizard's six moves.
- Agents answering their questions.
- **Random play in every mode, with all invariants checked after every move**, on a fake clock for the timed modes.

The full suite takes about a minute: `go test ./...`.

## 5. Your ten points

1. **Modes of play.** Implemented as orthogonal settings with eight presets: classic, infinite, target, timed, speed, cq, debate and solo (`game-rules.md` §6). Other termination conditions follow from Carneades and the dialogue literature, and are implemented:
   - **concession on the burden of proof**: pro passes without carrying the claim, or con passes while it stands, as in Prakken's persuasion dialogues;
   - **an unanswered question**: the burden shift of a critical question;
   - **resignation**;
   - **stalemate**: everyone passes.

   Not implemented:
   - the SRD's "proof standard met", for the reason in section 1.4;
   - a moderator vote;
   - running out of legal moves, which cannot happen because thesis-independent schemes back every question.

   One caution: a human at the CLI cannot be interrupted, so the clocks apply when the move is submitted. A real-time UI would enforce them while the player is typing.
2. **Defence in depth.** Yes, as you describe. A positive answer backs the premise, or denies the exception. It is a full argument with its own questions, so its premises can be attacked, defended and questioned in turn. Each level protects the one above until knocked out. Two asymmetries come from Carneades and Walton:
   - a tie leaves both sides Out, so a defence must outweigh the attack, not just match it;
   - an exception holds only if argued for, so asking about an exception without an argument is refused.
3. **All schemes should be challengeable, and CQs answerable by schemes.** Done for the seven schemes: 20 questions. Every question is answerable: the thesis-independent schemes back all of them, and five backup edges add others. Expanding the NASCQ toward Fig. 4's 29 schemes and 39 questions needs, for each new question:
   - its locus;
   - its pattern;
   - for a new scheme, its Carneades formulas, Language templates and answer variable.

   Lint catches structural mistakes. Semantic ones need a person.
4. **Last-mover advantage depends on the mode.** It does, and section 6 measures it. Between players of equal skill, any mode that ends at a cut-off (rounds or a clock) is won by whoever moves last. Arguments cost nothing and defeat is strict, so each move can undo the one before it. Modes that end by play are fair in random play: debates split 103 to 97. The fixes are seat alternation (`--alternate`, done), the undercut bonus removed under live scoring (done; target self-play is now 7 to 5), and making arguments cost something (your decision).
5. **Score revision depends on the mode.** `move` scores once; `live` keeps every score current, so defeated arguments lose their points. The scoring model is a setting of every mode.
6. **A central claim with sides, and other coherent modes.** The debate mode is the central-claim game. Section 8 lists other structures that are logically coherent with Carneades and Walton's dialogue types, and which of them exist.
7. **Binding a scheme's conclusion to an existing statement.** It depends on the scheme, as you said, and the NASCQ now records how. Schemes with an answer variable conclude the answer itself, with premises about it. The others conclude the answer as an enthymeme. Defence is not cheap:
   - an answer costs a scheme instance whose premises can be attacked and questioned;
   - G1 forbids restating;
   - live scoring pays a player once per statement, for their best argument for it.
8. **CQ linkages in the NASCQ, and Appendix A.** See section 3.2. The 20 questions here supersede the plan's Appendix A list. The mapping of each question to its locus is the part that needs review: by you or a domain expert, against Walton, Reed and Macagno (2008).
9. **Scoring by mode, and weighing undercutters.** Scoring is per mode. Undercutters stay unweighed, as Carneades defines them. The weighed reply is to deny the exception or the undercutter (section 3.1).
10. **Self-play.** Two kinds are supported:
    - **self-play**: two seats played by the same agent type, which an AI trains against itself with;
    - **solo**: one seat that argues both sides, whose result is the verdict, i.e. belief revision.

    Attacking your own argument is legal and scores nothing, so it can't be farmed.

## 6. Measurements: does the seat matter?

Tournaments are run with `wag-cli --mode M --games N --p1 X --p2 Y`.

- The greedy agent tries 40 sampled templates with 3 bindings each.
- Random-against-random runs are 200 games of up to 20 rounds.
- Runs with greedy players are 24 games of up to 12 rounds.

The table gives first-seat wins / second-seat wins / ties. It was run with the +15 undercut bonus in every mode. That was the default until this commit; the second table shows what happens without it.

| Mode | random v random (200 games) | greedy v random | random v greedy | greedy v greedy (self-play) | How games ended |
|---|---|---|---|---|---|
| classic (move scoring) | 88 / 110 / 2 | 24 / 0 | 0 / 24 | 5 / 19 | round limit |
| infinite (live), cut at the round limit | 84 / 114 / 2 | 24 / 0 | 0 / 24 | 0 / 24 | round limit |
| target 50 (live) | 91 / 108 / 1 | 24 / 0 | 0 / 24 | 0 / 24, in 2.0 rounds | target (183 of 200 random games) |
| cq (live) | 93 / 106 / 1 | 24 / 0 | 0 / 24 | 1 / 23 | round limit; 7 random games lost by an unanswered question |
| debate (live, PE) | 103 / 97 / 0 | 24 / 0 | 0 / 24 | 0 / 24, verdict Out | random: 100 concessions by con, 92 by pro; greedy self-play: round limit |

Solo: the random player ends In 107 times and Out 93 times. The greedy player ends In 24 times out of 24. With one player, `Value` has no verdict to aim at, so it maximises points, and points come from arguments that stand.

Runs without the undercut bonus (`--undercut-bonus 0`), which is now the default of the live modes:

| Mode | Matchup | first / second / ties |
|---|---|---|
| target 50 | greedy v greedy, 12 games | **7 / 5**, in 8.4 rounds (with the bonus: 0 / 24, in 2.0 rounds) |
| infinite, cut at 12 rounds | greedy v greedy, 12 games | 0 / 12 |
| infinite, cut at 20 rounds | random v random, 200 games | 93 / 105 / 2 (with the bonus: 84 / 114 / 2) |

**Findings.**

1. **Skill beats seat.** The greedy agent beats the random agent from either seat in every mode: 24 of 24 games, every time.
2. **Between equals, the seat decides.** The second seat, which moves last, wins:
   - nearly every greedy self-play game that ends at a round limit (classic 19 of 24, infinite 24 of 24, cq 23 of 24, debate 24 of 24);
   - every target game while the undercut bonus applied.

   Random play shows the same tilt, smaller: 53–58% for the second seat.

   **The cause is structural, not a bug.** Three properties combine:
   - a scheme turns any words into a new argument;
   - its premises are presumed;
   - defeat is strict (an In undercutter or a lost issue defeats outright).

   So every move can undo the one before it, and whoever moves last before a cut-off flips the result. In the debate self-play games, con made the last move and the claim ended Out in all 24.
3. **Games that end by play are fairer.** In random debates, which end by concession (100) or failure to carry the burden of proof (92), the seats split 103 / 97. Target games without the bonus split 7 / 5.
4. **The undercut bonus was the main cause in target mode.** Under live scoring, a defeat already costs its victim their points, so the +15 counted it twice. It made the responder reach 50 in two moves. It is now 0 in the live presets; classic, which scores once per move, keeps it.

**What to do about the rest.** These are decisions for you:

- **Alternate seats in tournaments.** `--alternate` does this, as colours alternate in chess.
- **Prefer modes that end by play** (debate with concession, cq, target) over round limits for single games.
- **Make arguments cost something**, so they stop being free. Options: moderated weights (the SRD's Moderated mode), an argument budget, a points cost per new premise, or commitment rules (SRD §14) that punish asserting a claim and its contrary.
- **Give the side with the burden of proof the closing reply in debates**, as many debate formats give it to the affirmative. That moves the advantage to pro rather than removing it.

## 7. What I need from you

1. **A review of the new data.**
   - The 25 loci and patterns in `data/nascq/edges.json`.
   - The four answer variables in `schemes.json`.
   - The wording of the 13 new questions.
   - Above all, the polarity calls: which answer counts as positive.
2. **The source texts for expanding the NASCQ**: the question lists of Walton, Reed & Macagno (2008) for the schemes of Fig. 4 you want next. With them I can add schemes and questions, and lint will check the structure.
3. **Decisions.**
   - Make the existential exceptions (other means, best means, other goals) specific, as Pyon & Ri recommend?
   - Model enthymematic answers as explicit chains?
   - Default proleptic bonus: 0 or 5?
   - Should debates use CCE or BRD for some topics?
   - Should points modes give the first seat the last word, or always alternate seats?
4. **For timed and speed play with people:** whether a real-time interface is in scope.

## 8. Other coherent modes

Each is a combination of Walton's dialogue types (Walton & Krabbe 1995) with Carneades' semantics.

| Mode | Structure | Status |
|---|---|---|
| Persuasion (debate) | One claim, pro with the burden of proof, con. Ends by concession or rounds; verdict decides. | **implemented** (`debate`) |
| Critical-question duel | Ask; the proponent must answer, and answers are asked in turn. Fails when unanswered. | **implemented** (`cq`) |
| Solo belief revision | One player, both sides; the result is the final belief. | **implemented** (`solo`) |
| Points race (open, timed, speed, target) | No claim; the score decides. | **implemented** |
| Team debate | Even seats pro, odd seats con. | **implemented** (the claim objective with four or more players) |
| Inquiry (cooperative) | All players try to *settle* a claim (In or Out) under BRD; they win together if it is settled, and lose if it is left open. | not implemented: an objective on `Outcome.Verdict` |
| Deliberation | Each player proposes an action, and practical-reasoning conclusions compete. The proposal left In wins, or nothing if none is. | not implemented: several claims, one per player |
| Burden handicap | The stronger player argues under CCE or BRD, the weaker under PE. | possible now through `--standard` |
| Commitment persuasion | Players' commitment stores (SRD §14). Asserting the negation of one's own commitment loses or costs. | not implemented: needs the consistency layer |
| Moderated | Weights come from moderators (SRD Moderated mode). | not implemented |
