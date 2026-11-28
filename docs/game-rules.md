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
| premise P | possibility: `possible_bring_about(P,A)` | argues ¬P: undermines | argues P: backs the premise |
| exception E | other means: `realize_more_effectively(P,G,A)` | argues E: undercuts, through the scheme's exception argument | argues ¬E: E becomes an issue, decided by weight; a tie leaves E Out, so the argument stands |
| conclusion C | side effects (a further bad consequence, D095 (1)) | argues ¬C: rebuts | argues C: supports |
| argument | another event (sign's: a better explanation of the sign) | argues `undercut(a)`: undercuts | argues ¬`undercut(a)`: the undercutter becomes an issue |

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
- A free-text **bridge** on a backup edge (`"bridge": "why"`) no longer lets
  a scheme back answers of another form (D075, since 2026-11-22): a note
  cannot supply the inference from the scheme's conclusion to the answer.
  Lint refuses a bridge, and an answer of another form needs a typed answer
  scheme whose premises establish it. The shipped data has never declared a
  bridge; the audit of 2026-10-13 (§4) showed why: until then an edge stood
  for the link silently. Value arguments backing "are there other goals that
  should come first?" and "is V really a positive value?" concluded answers
  their premises do not support; those edges are gone.
- A saved game in which a bridge let a move through (possible only on custom
  data) cannot be replayed: loading it stops at that move and says so. Other
  data (`--data`) does not help; a build from before 2026-11-22 replays it.
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

Analogy is not thesis-independent since 2026-10-15: its formula had one
proposition for the source case and the target (`in_case(A,C)` ⇒ `A`), so
as the answer to any question it would say the answer held in another case.
It still argues for or against a statement. Since 2026-10-22 `walton.yml`'s
analogy has a claim in each case (`P1`, `P2`; decision D001), and the game
takes that form when its data is generated from the catalogue (workstream
E); how far it then reaches as an answer is decision D057. Declared
contraries, such as `right_to_do` and `wrong_to_do`, are not enforced: both
can be accepted at once (decision D053).

The game binds this variable to the answer, so the premises are about the
answer (`asserts(W, not(possible(A)))`) and the player does not fill it in.
The other schemes conclude something more specific, such as
`should_be_performed(A)`, and back only answers of that form (above).

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

**Validity checks.** NASCQ checks S0 to S9 and Q1 (`nascq.Validate`); S2 refuses a
variable left out or bound to a value that names nothing, with no letter or digit (blank,
punctuation alone, invisible characters, a number not written in digits such as "½"), and in a game
with a profile P1 refuses such an action first, as the profile's check refuses an actor, an action
or an alias that names nothing (since step C5.2's item F: an answer whose H or I names nothing would
make its exception out without saying what outweighs or what would settle it). Other's premises and
conclusion are free text, and S2 refuses one that names nothing likewise; its premises after the
first may be left blank. The terminal asks again for such a value, and a terminal seat whose input
has ended resigns. A game recorded with one of these before 2026-11-27 replays with a build of
2026-11-26 or earlier (with the scheme data it was played with), and the loader says so. The game
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

**The question protocol** (D003, D075). Who must answer an asked question is
a rule of the dialogue, not a change of who proves an exception:

- **Who asks what.** Any player may ask, about any argument on the board, a
  question its scheme asks at a premise or of the inference, and, where any
  premise may be challenged, `CQ:premise_acceptable` of any premise. A
  question about an exception or the conclusion is not asked alone (Q1): an
  exception is the critic's to support (D003), and a counterclaim is argued.
  The same question about the same argument and premise is asked once (Q2).
- **While it is pending.** The question is an argument for the negative
  answer, with no premise, owned by the asker, at the presumption's weight.
  On a premise it ties that premise's presumption, so the premise is out; on
  the inference it puts the argument out of force. Nothing else changes: the
  argument's other premises, its exceptions and the other arguments stand.
- **Who answers, and how.** The proponent of the argument asked about (its
  owner) must answer, and any player's answer closes the question: a defence
  that names the question, or any argument for the premise asked about. The
  answer outweighs the question's presumption, so the premise or the
  inference holds again until it is attacked; the answer can itself be
  questioned and attacked.
- **When it is closed.** A question is closed when it is answered. With
  `AnswerWithin` set (the `cq` mode sets 1), an owner who lets that many of
  their own turns pass without answering loses the game; otherwise open
  questions are shown, not enforced.
- **What may answer.** A scheme on the question's backup edges, a
  thesis-independent scheme (they back every question), or Other, unless the
  question refuses Other (D095 (3)): then only its typed answers may, and a
  well-made Other is refused too (S5), whether it answers the question or
  argues the exception as a move about the statement (or against its
  negation). The schemes that back such a question bound every route to its
  exception (D099 (4)): its answer scheme and the source-based schemes argue
  for the exception, as an answer or by a move about it; sign and analogy,
  which conclude any statement, are refused there as Other is (S5), and the
  board does not offer them. Arguing against the exception is not bounded,
  and an undercutter of the question's argument is kept: it concludes neither
  the exception nor anything it presumes. The bound holds by the exception's
  form, before and without a bridge (since 2026-11-28; F1 of the external
  review of exchange 3): a statement that matches the exception pattern of a
  question that refuses Other is bounded whether or not a bridge on the board
  has it as its exception, in any game whose data has the question, with or
  without a profile, and also where no bridge in the game could take it
  (another actor's action). So in an inquiry whose claim is the exception,
  only expert opinion and position to know argue for the claim before any
  bridge. Until then the bound held only at a bridge on the board, so sign
  could make the statement In first and the bridge, made later, stood
  defeated. S5's refusal says when it bounds by the form alone ("no bridge on
  the board has it as its exception yet"). A game that recorded such a move
  replays with a build of commit 6916369 or earlier, and the loader says so.
  An answer scheme (one that concludes
  the formula of the question it answers, D094 (g)) is offered only as an
  answer, on its backup edges, never as a move of its own (S1). Since step
  C5.3's switch, negative consequences backs CQ:side_effect (D099 (1)), as
  positive consequences backs CQ:opposite_consequences: a further bad
  consequence answers the side-effect question of positive consequences and
  practical reasoning, a rebuttal at the conclusion; asked no question, it is
  a direct rebuttal, as before.

This formalizes the game's behaviour since October 2026, but for the two
rules step C5.2 adds (an answer scheme only as an answer, a question that
refuses Other), which no reference game reaches, so no golden game moves.
Step C5.3's rules (P5 below, and the bound on the exception) also change what
the board and the computer players are offered, so 24 of the 33 golden games
on the shipped data were made again, each first changed where its menu lost
two offers (docs/c5-game.md, C5.3). At step C5.3's switch the new backup edge
added one offer to some menus, and 23 of the 33 were made again, each first
changed where its menu gained that offer, every earlier move the same.

**The profile** (D094 (b); D089, D091, D092, D063). A game made since
step C5.2 has a profile, a discussion imported from Kialo included: the
agent who decides, the actors, the actions argued about (each with its one
actor and its aliases), and the situation the alternatives are compared in.
It comes from a case record (`--case`, the new-game form, `"case"` in
`POST /api/games`), or the agent and the situation are given when the game
is created; otherwise it is the default, an agent called `decider`, in the
present situation. The board shows it under *Who decides*. A claim that
recommends an action is written with the action's id (an alias in it reads
as the action). Its rules come before and after NASCQ's (S0 to S9):

- **P1.** A variable of type action is bound to a declared action, by its
  id or an alias, names compared as atoms, as the board makes statements
  of them (`tax meat`, `Tax-Meat` and `tax_meat` are one action; actors
  too); an alias is read as its action. A move that first argues about an
  action declares it, with its actor:
  `"declare": [{"id": "ban_meat", "actor": "governments", "aliases": []}]`.
  A move declares only actions it binds, so a question declares none. The
  declaration is recorded with the move and joins the profile only if the
  move is accepted. A declaration without an actor, with an actor the
  profile does not know, or of a name already declared as another actor's
  is refused; one of a name already declared as the same actor's declares
  nothing new but the aliases it adds. The board and the terminal ask whose
  a new action is, the agent's unless the player chooses another actor. The
  values a move's target fixes are checked too; the scheme and values a
  question carries are not its own (S0 sets them aside).
- **P2.** A variable of type agent (practical reasoning's P, `better_means`'
  P) is the agent: the game binds it, the board shows it as fixed, and a
  move that binds it to anyone else is refused.
- **P3.** The actions argued are the agent's (their actor, aliases
  included): a recommendation names no agent but the one addressed (D063),
  so a P that does not perform A is refused, and so is any other scheme's
  move about another actor's action.
- **P4.** A `better_means` answer's alternative A2 is another of the agent's
  actions, not A and not one of A's aliases; the move records how it is
  better (`"comparison": "surely"`, `"sooner"` or `"cheaper"`; D089) and is
  refused without it. Its P, G and A are the answered exception's (S9), so
  its goal premise is the answered argument's.

- **P5.** An answer the data offers but WAG's first profile leaves out
  until its warrant is reviewed is refused, and the board and the computer
  players are not offered it: sign's answer to CQ:side_effect of practical
  reasoning (`AS:practical_reasoning`, `AS:practical_reasoning_goal`), of
  either polarity (D099 (3)). Sign's answer to positive consequences'
  side-effect question is played as before (D009), and a game with no
  profile plays both.

- **P6.** An undercutter of an argument whose question refuses Other (a
  bridge, whose question asks whether the consequence is worth finding out
  first), made by a move that answers no question (a direct undercutter, or
  an argument for the argument's undercut statement), declares its ground:
  `"ground": "other"`, for an attack on something other than whether the
  consequence could be found out first (D099 (4)). That is the question's,
  answered with the schemes that back it, so a move that declares no ground,
  or `"ground": "knowability"`, is refused and pointed to the question. The
  game reads formulas, not the free text, so the declaration is recorded
  with the move, and whether the ground is truly another is the review's
  (D053). An undercutter of any other argument declares nothing.

P1 (the declarations) and P2 are checked before NASCQ's rules, P1 (the
values), P3, P4 and P5 after them, and P6 once the move is placed (it reads
what the move's arguments conclude, by whatever route); a preview checks a
move as a turn does.
The situation is shown and recorded, not checked move by move (D089).
Computer players argue about the agent's actions, and declare the new ones
they name.

A game made before step C5.2 has no profile and plays without these rules.
Its export, a reference game's and a Kialo import's included, loads and
replays as it was played. A trace a model player recorded before then is
replayed through a new game (`wag-llm --replay`), which has a profile, so it
replays with the build it was recorded with, and with the data it was
played with (`--data` and the frozen files, wag/data/frozen).

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

The game clock is read when a move is made, and before a model seat is asked
for one: once the time is up the game ends, and the move is not played (nor a
move held for the moderator, whose approval then ends the game without it).
A game resumed from the data folder counts the time played before it stopped,
not the time it was stopped (`docs/g-journal.md`).

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
| inquiry | claim, two players together: no sides, no winner | live | 20 rounds, a round in which both pass, or a resignation |

Every setting can also be changed on its own:

- `--scoring`, `--target`, `--time`, `--move-time` and `--timeout-loses`;
- `--answer-within`, `--objective`, `--claim`, `--standard` and `--concede`;
- `--undercut-bonus` and `--proleptic-bonus`.

In tournaments (`--games`), `--alternate` swaps the seats every other game.

**Self-play** is two seats played by the same kind of agent
(`--p1 greedy --p2 greedy`). **Solo** is one seat arguing both sides.
**Inquiry** is two players examining a claim together, as the first evidence
slice's players do (D076; `docs/e1-evidence.md`): neither has a side or wins,
and the result is the verdict.

**Dialogue type.** In Walton's terms every mode but `inquiry` is a persuasion
dialogue, or none:
- `debate` is the one-sided form: pro must prove the claim, and con need only
  doubt it.
- `cq` adds a strict duty to answer.
- `solo` is one person deliberating with a devil's advocate on a practical
  claim, or a one-person inquiry on a factual one.
- `inquiry` takes the inquiry's relation and its result, not its rules: two
  fellow investigators, with no side, find the claim proven, disproven or
  not settled (In, Out or Undecided; `dialogue-types.md` §5.2), under the
  game's standard of proof; the inquiry's high standard of proof and
  evidential priority are not imposed.
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
- `HumanAgent` is the CLI wizard: Argue; For or against a statement (no
  question named); Attack, Defend and Question through a critical question;
  Pass and Resign.
- Language models play through the text protocol (`docs/protocol.md`): a
  prompt in English, a free-text reply ending in one JSON move. In the browser
  a seat can be a language model (copy and paste) or Claude (the API).
- `JSONAgent` (`--p1 json`) exchanges one `game.View` line for one action line.
  The View includes:
  - the mode, objective, claim, verdict and side;
  - the open questions;
  - every legal action, with its locus and answer variable;
  - the outcome, once the game is over.

## 8. Evidence in an inquiry (E1)

An inquiry may be played with a **record set** ([`e1-evidence.md`](e1-evidence.md)): sources,
artifacts, representations, claims and dependencies captured before play, a `wag-evidence/0.2`
bundle without the records the game writes ([`evidence-contract.md`](evidence-contract.md)).
`wag-server` reads the record set `NAME` from its data folder's `evidence/NAME/records.json`, with
its payloads under `evidence/NAME/`; a request names it (`"evidence": "NAME"`). Only an inquiry of
two seats, each a person or a model (a person relaying a chatbot counts as a model), not
moderated, takes one: a rule bot cannot consent (D076). The game checks the record set in full when
it is created or loaded, writes the context and its actors (`game:context`, profile
`wag-inquiry/1`, policy `two-party-consideration/1`, the players `game:player:<id>`, the game
`game:wag`), and records the record set's name and SHA-256 in the trace's start and the export.
The board shows the sources, claims, dependencies and bindings as text.

**Acts.** Three moves act on evidence. Each takes the turn and is a move, not a pass; one that
breaks a rule is refused with the usual penalty and leaves no record.

- An **offer** (`evidence_offer`) binds a claim of the record set to a statement of an argument
  that a support or attack move made: a premise, an assumption (one of its scheme's assumptions;
  an Other argument has none), an exception, its conclusion, or its inference (whose statement is
  `undercut(A)`, whether or not the argument has been undercut yet), with a role (`support`,
  `oppose`, `context` or `definition`), an interpretation and a warrant. The game writes the target (`game:t<i>`, by
  `game:wag`, with the argument's revision, `wag-argument-revision/1`) and the binding (`game:b<i>`,
  by the player), `<i>` being the act's place in the history. The offer is its player's consent.
  An offer identical to one made is refused; any other change makes another binding, with its own
  revision.
- A **consent** (`evidence_consent`) names a binding and its exact revision, in full or by its
  first 12 digits; a consent for another revision is stale.
- A **withdrawal** (`evidence_withdraw`) takes back one's own standing consent.

A binding is **admitted for consideration** while both players' consents to its revision stand.
Admission sets no label, weight or assumption: it is not a finding of fact (D076). An act's records
are dated at the act, and never before the records already made (a clock set back, or a game
resumed on another computer, refuses no act). In a game with a record set, a move other than an
act that carries an act's parts (`offer`, `binding_id`, `revision`) is refused; a game without a
record set ignores those fields, and an act's type there is an unknown move, as before E1.

**Citing a binding.** A support or attack move may cite admitted bindings (`evidence_refs`),
besides its citation strings, which keep their meaning. Each must agree with what the move's
argument concludes:

| The move's argument concludes | A cited binding must be |
|---|---|
| S (an argument for S; an answer at an exception S) | `support`, its target's proposition S |
| the opposite of S (a rebuttal, an undermining of premise S, a defence at exception S, an attack on S) | `oppose`, its target's proposition S |
| `undercut(A)` (a direct undercutter of A) | `oppose`, at A's inference |
| the opposite of `undercut(A)` (a defence of A's inference) | `support`, at A's inference |

Context and definition bindings ground no move; a question, a pass, a resignation or an act cites
none. *Require sources* is met by a citation string or by a cited binding that agrees.

**Withdrawal** ends the admission. A move that cited the binding stays, in force as before; the
board shows its reference withdrawn, and an argument each of whose cited bindings was withdrawn,
citing no string, as no longer grounded by admitted evidence. The mark changes no label.

**Load and resume.** The record set is read again from the evidence folder: its SHA-256 must be
the game's, its payloads must check, and the server's pins must be the game's; each act is replayed
with the time it was made and must make or name the revisions recorded. A game with a record set
is not opened read-only.
