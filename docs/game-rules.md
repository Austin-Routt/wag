# WAG rules and game modes (as implemented, 2026-10-13)

This is the rulebook of the code on this branch. The design reasons and the
review of the sources are in `review-2026-10-13-part2.md`.

## 1. The board

The board is a Carneades-4 argument graph. Every statement is labelled **In**,
**Out** or **Undecided** by Carneades' grounded labelling after every move.

- A statement with an argument for it that is in force, and with no issue, is
  In.
- An **issue** is a statement and its opposing position (`p`, `¬p`). Its
  proof standard decides between them by the weight of the best argument on
  each side.
  - PE: the stronger side wins.
  - CCE: the stronger side must also lead by more than 0.5.
  - BRD: as CCE, and the weaker side must also be below 0.3.
  - A tie, or a standard that is not met, leaves both positions Out.
- An argument is **in force** when its premises are In and its undercutter is
  not In. An undercutter that is In defeats its argument whatever the weights:
  Carneades does not weigh undercutters.

Arguments are instances of the schemes in `data/schemes/core.yml`. The NASCQ
(`data/nascq/*.json`, after Pyon & Ri 2022) decides three things:

- which critical questions (CQs) can be asked of an argument made with a scheme
  (ask edges);
- which schemes can back an answer to a CQ (backup edges; thesis-independent
  schemes, expert opinion and position to know, back every CQ);
- which part of the argument each CQ is about: its **locus**, given on the ask
  edge.

| Locus | Example | Negative answer (an **attack**) | Positive answer (a **defence**) |
|---|---|---|---|
| premise P | possibility: `possible(A)` | argues ¬P: undermines | argues P: backs the premise |
| exception E | side effects: `side_effects(A,S1)` | argues E: undercuts, through the scheme's exception argument | argues ¬E: E becomes an issue, decided by weight; a tie leaves E Out, so the argument stands |
| conclusion C | other goals | argues ¬C: rebuts | argues C: supports |
| argument | evidence | argues `undercut(a)`: undercuts | argues ¬`undercut(a)`: the undercutter becomes an issue |

**Any statement can be argued for or against.** A premise, a conclusion,
an exception someone raised, and the claim are all statements. A move about
a statement (`target_stmt_id`) argues for it, or for its opposite, as one
argues for or against the claim. A defeasible argument "can be attacked by
an attack on a premise, by a counterargument with an opposite conclusion,
or by an argument attacking the inference rule" (Walton, Reed & Macagno
2008, p. 32).

**The scheme must fit the statement (S9).** The scheme of such a move, of
a direct attack or defence of an argument's part, and of an answer to a
critical question, must be able to conclude the statement, or its opposite:

- Other, and the schemes whose conclusion is a statement variable (expert
  opinion, position to know, sign, analogy), conclude any statement. The
  variable is bound to it.
- Any other scheme fits a statement its conclusion unifies with, and the
  statement fixes the variables of its conclusion:
  - practical reasoning or positive consequences can argue for
    `should_be_performed(carbon_tax)` with A = carbon tax, and the player
    may not give another A;
  - negative consequences, which concludes `¬should_be_performed(A)`, can
    argue against it;
  - neither can conclude `possible(carbon_tax)`.

  The same holds for an answer to a critical question whose backing scheme
  concludes the answer's form: its premises must be about the same action.

**Challenge** (`GameConfig.AnyPremise`, on in every mode). A scheme's own
critical questions often point at its exceptions and its conclusion;
practical reasoning's five touch only one premise (possibility). With this
option, any premise that none of the scheme's own questions is about can be
challenged ("Why should we accept that “P”?", `CQ:premise_acceptable`).

- A challenge is a question: it shifts the burden of proof (below), and any
  argument for the premise answers it.
- To argue against the premise instead, argue against the statement.

**Arguments that answer a question.** The answering argument concludes the
answer, so its scheme must be able to conclude it (S9).
- A scheme with an answer variable has the answer in its premises: the
  expert asserts the answer itself.
- A scheme without one backs only the answers its conclusion can be. Positive
  consequences can answer "would doing A also have good consequences, to
  weigh against G?" (asked of a negative-consequences argument) against it,
  with "A should be done", about the same A. It cannot answer it in its
  favour.
- A backup edge may declare a **bridge** (`"bridge": "why"`): then the scheme
  backs answers of another form, its argument concludes the answer in place
  of its own conclusion, and the move records the bridge. The scheme's
  questions about its conclusion then do not apply to that argument
  (`nascq.OwnConclusion`); its other questions do.
- The shipped data declares no bridge. The audit of 2026-10-13 (§4) showed
  why: until then an edge stood for the link silently. Value arguments
  backing "are there other goals that should come first?" and "is V really a
  positive value?" concluded answers their premises do not support; those
  edges are gone.
- `wag-cli --lint` reports a backup edge whose scheme can conclude none of its
  question's answers. `wag-cli --edges` plays every answer the NASCQ offers to
  every question and prints what each concludes:
  `docs/scheme-registry/nascq-edges.md`.

**The literal answer.** Each question says which literal answer counts
against the argument (`answer_en` in `critical_questions.json`). "Would doing A
have bad side effects?" is answered "yes" by an attack. "Is A possible?" is
answered "no" by one. The browser's buttons and the prompt say which.

The answering argument is a full instance of the backup scheme: it has its own
premises and exceptions, and its own critical questions. Each answer
therefore protects the part above it until someone knocks it out, at any
depth. `TestDefenceInDepth` plays such a chain four levels deep.

A scheme whose conclusion is a proposition variable declares that variable
as `answer_variable`:

- S of expert opinion and of position to know;
- A of analogy;
- B of sign.

Analogy is not thesis-independent since 2026-10-15: its formula has one
proposition for the source case and the target (`in_case(A,C)` ⇒ `A`), so
as the answer to any question it would say the answer held in another case.
It still argues for or against a statement; the audit's N12 (a source and a
target case) is pending.

The game binds this variable to the answer, so the premises are about the
answer (`asserts(W, not(possible(A)))`) and the player does not fill it in.
The other schemes conclude something more specific, such as
`should_be_performed(A)`, and back only answers of that form, unless their
backup edge declares a bridge (above).

## 2. Moves

On your turn you make one move. A rejected move costs 5 points, and you
move again.

| Move | JSON | What it does |
|---|---|---|
| Argue | `{"type":"support","scheme_id":…,"variable_bindings":{…}}` | A new argument. Its premises are presumed: they hold until attacked or questioned. |
| Attack | `{"type":"attack","answers_cq_id":…,"target_arg_id":…,"scheme_id":…,…}` | A negative answer to a CQ about the target, backed by a scheme (see the table). Without a CQ, it is a direct rebuttal, undercutter or undermining (`attack_type`), with a scheme that fits (S9). |
| Defend | `{"type":"support","answers_cq_id":…,"target_arg_id":…,"scheme_id":…,…}` | A positive answer to a CQ about the target. Without a CQ, it directly backs a premise (`target_premise_id`) or the conclusion. |
| For / against a statement | `{"type":"support" or "attack","target_stmt_id":…,"scheme_id":…,…}` | An argument for the statement, or for its opposite, with a scheme that can conclude it (S9). An argument for a challenged premise answers the challenge. |
| Question | `{"type":"question","answers_cq_id":…,"target_arg_id":…}` | Asks the CQ without arguing, or challenges a premise (`CQ:premise_acceptable`). See below. |
| Pass | `{"type":"abstain"}` | Passes the turn. |
| Resign | `{"type":"resign"}` | Concedes the game. With three or more players on points, the others play on. |

**Validity checks.** NASCQ checks S0 to S9 and Q1 (`nascq.Validate`). The game
then adds three rules:

- G1: a move may not restate an argument already in the graph;
- Q2: a question that has already been raised about an argument may not be
  asked again;
- the graph must stay acyclic.

**Questions shift the burden of proof.** A question about a premise, or about
the inference, adds a premise-less argument for the negative answer, owned
by the asker. It has the weight of a presumption (0.01).

- On a premise, it ties the premise's presumption, so the premise is Out until
  the proponent backs it.
- On the inference, it puts the argument out of force until the proponent
  answers.

This is how Carneades treats asking about an assumption (Gordon, Prakken and
Walton 2007). A question about an exception or a conclusion is refused
(Q1): whoever raises an exception or a counterclaim must prove it, so it is
raised with an argument (an attack).

## 3. Scoring

| Model | Points |
|---|---|
| `move` (the SRD's snapshot) | Once, when the move is made: for each argument put forward, 10 × weight if its conclusion is In right after the move, 2 × weight if Undecided, plus 15 when an undercutter turns its target's conclusion Out. |
| `live` (the SRD's delta) | Recomputed from the labels after every move. For each statement you argued for: 10 (In) or 2 (Undecided) × the weight of your best argument for it that is in force. Plus the undercut bonus for each argument your undercutters currently defeat. Plus your flat points (penalties and proleptic bonuses). |

The undercut bonus is 15 in classic and 0 in the live presets. Live scoring
already counts a defeat, because the defeated argument loses its points, so
the bonus would count it twice. With the bonus, greedy self-play in the target
mode ended in two rounds, always won by the second seat (see the part 2
review, section 6).

In both models, three kinds of move earn nothing:

- an attack on your own argument (it is legal: belief revision, self-play);
- an attack on an argument already out of force;
- a question.

An optional proleptic bonus (`ScoringConfig.ProlepticBonus`, SRD §16.4)
rewards a defence of your own argument against a question nobody has raised
yet.

Under live scoring, a defeated argument loses its points, and a second
argument for a statement you already argued for adds nothing unless it is
stronger.

## 4. Objectives

- **points**: the highest score wins (a draw if tied); a player who forfeits
  loses.
- **claim** (a debate):
  - The central claim is `GameConfig.Claim`, or the conclusion of the first
    argument of the first seat.
  - It is at issue from the start, under `ClaimStandard`.
  - Even seats are pro and bear the burden of proof; odd seats are con.
  - Pro wins if the claim is In when the game ends, and con wins otherwise.
  - With one player it is solo belief revision: the result is the verdict,
    and there is no winner.

## 5. When the game ends

The game ends at the first of these conditions that is met.

| Condition | Setting | Result |
|---|---|---|
| A player resigns | always | they lose |
| A question about your argument goes unanswered for N of your turns | `AnswerWithin` | you lose |
| A move comes later than the move clock allows | `MoveTimeLimit` with `MoveTimeoutLoses` | you lose (without it, the turn passes) |
| A side passes while the verdict is against it | `ConcedeOnPass` (claim) | it concedes: pro retracts a claim it cannot carry; con concedes one it cannot shake |
| A score reaches the target | `TargetScore` | by the objective |
| The game clock runs out | `TimeLimit` | by the objective (truncated) |
| The round limit is reached | `MaxTurns` | by the objective (truncated) |
| Every player passes N rounds running | `AbstentionThreshold` | by the objective |

SRD §17.3's "proof standard satisfied" condition is deliberately left out. It
ends the game as soon as the claim meets its standard, which is usually right
after the opening argument. In this implementation the verdict decides the
winner when the game ends, and the concession rule ends the game when the
losing side has nothing more to say.

## 6. Modes (`wag-cli --mode`, `game.ModeConfig`)

| Mode | Objective | Scoring | Ends |
|---|---|---|---|
| classic | points | move | 20 rounds, or two all-pass rounds |
| infinite | points | live | a resignation, or two all-pass rounds |
| target | points | live | a score of 50 |
| timed | points | live | a 10-minute game clock |
| speed | points | live | 20 rounds; a move later than 60 s passes the turn |
| cq | points | live | 30 rounds; a question left unanswered on your next turn loses |
| debate | claim | live | 20 rounds, or a concession |
| solo | claim, one player | live | the player passes |

Every setting can also be changed on its own:

- `--scoring`, `--target`, `--time`, `--move-time` and `--timeout-loses`;
- `--answer-within`, `--objective`, `--claim`, `--standard` and `--concede`;
- `--undercut-bonus` and `--proleptic-bonus`.

In tournaments (`--games`), `--alternate` swaps the seats every other game.

**Self-play** is two seats played by the same kind of agent
(`--p1 greedy --p2 greedy`). **Solo** is one seat arguing both sides.

**Dialogue type.** In Walton's terms every mode is a persuasion dialogue, or
none:
- `debate` is the one-sided form: pro must prove the claim, and con need only
  doubt it.
- `cq` adds a strict duty to answer.
- `solo` is one person deliberating with a devil's advocate on a practical
  claim, or a one-person inquiry on a factual one.
- The points modes match no type.

The context presets are domains (weights and standards), not types. Other
types are planned: [`dialogue-types.md`](dialogue-types.md) §9,
[`plan.md`](plan.md) §8.

## 7. Agents and programs

- The rule bots fill a move's parts with the words already used on the board
  for that kind of part, or with neutral placeholders ("an expert", "the
  subject at hand"); sources are always placeholders. They test the rules;
  they do not understand the topic.
- `RandomAgent` plays a random legal move, and always answers the questions it
  owes in cq play.
- `GreedyAgent` tries up to 40 sampled templates, every answer it owes, and
  passing, each through `SimulateThen`. It plays the one that leaves it best
  placed (`game.Value`), choosing at random (seeded) between equally good
  moves:
  - a won game is worth 10⁶, a lost one −10⁶;
  - otherwise its lead in score;
  - plus 100 while the verdict favours its side.
- `HumanAgent` is the CLI wizard: Argue, Attack, Defend, Question, Pass and
  Resign.
- Language models play through the text protocol (`docs/protocol.md`): a
  prompt in English, a free-text reply ending in one JSON move. In the browser
  a seat can be a language model (copy and paste) or Claude (the API).
- `JSONAgent` (`--p1 json`) exchanges one `game.View` line for one action line.
  The View includes:
  - the mode, objective, claim, verdict and side;
  - the open questions;
  - every legal action, with its locus and answer variable;
  - the outcome, once the game is over.
