# Item F: the answer contract per route, written before the tests

Step C5.2's step 7 (docs/c5-game.md, *Item F*): for each question and answer route of the data the
game ships (fingerprint 092b6a6f48a672d2), the outcome a move must have, written here before any
test of it ran (2026-11-27), each case with the decision it follows. The tests
(`wag/internal/game/itemf_test.go`) play these cases on the shipped data; each is paired with a
planted defect it must catch (the last section). A case whose outcome turns out other than written
here is not rewritten silently: the change, and whether the game or this page was wrong, is
recorded below it.

*Refused* means: the move is refused with the rule named, and the graph (its statements with their
labels and texts, arguments, issues and assumptions), the open questions and the profile are as
before. The rules come in this order (game-rules.md): the profile's P1 for the move's declarations
and its own action values, and P2; NASCQ's S0 to S9 and Q1; P1 for the values a target fixes, P3
and P4; Q2 for a question; then, once the move is placed, the adapter's refusal (a premise that is
the argument's own conclusion), the cycle check, S5 for a refused exception and G1. A rule's refusal
(S, P, Q, G) is recorded in the move history once, with its penalty; the adapter's refusal and the
cycle check's leave no record and cost nothing. (Corrected after the independent check of step 7:
the page first put P1's own values after NASCQ, and said every refusal is recorded.) *Out of force* means the
argument does not count: an exception of it is In, a premise of it is Out, or a question about its
inference is pending.

## The board

The game is the shipped data with Other on, any premise open to challenge, no turn limit, two
players (Pro and Con), and a profile: the agent `decider`; the actors `decider` and `farmers`; the
actions `tax_meat` (the agent's, alias `meat_tax`), `subsidy_plants` (the agent's) and
`farmers_strike` (the farmers'). Moves are made by the side named; the other side passes between
two moves of one side.

| Name | Move | Arguments |
|---|---|---|
| c1 | Pro: AS:consequences, A `tax_meat`, G `emissions_fall` | concludes `should_be_performed(tax_meat)` |
| c2 | Pro: AS:consequences, A `tax_meat`, G `diets_healthier` | the same conclusion, another G |
| pr | Pro: AS:practical_reasoning, A `tax_meat`, S1 `high_emissions`, S2 `low_emissions`, G `emissions_fall`, V `sustainability` (P is the agent, bound by the game) | concludes `should_be_performed(tax_meat)` |
| pg | Pro: AS:practical_reasoning_goal, the same without V | concludes `should_be_performed(tax_meat)` |
| n1 | Con: AS:negative_consequences, A `tax_meat`, G `food_costlier` | concludes `¬should_be_performed(tax_meat)` |
| nm | Con: AS:negative_consequences_modal, A `tax_meat`, G `food_costlier` | concludes `may_be_unwise_to_perform(tax_meat,food_costlier)` |
| m+ | Pro: AS:consequences_modal, A `tax_meat`, G `demand_falls` | concludes `may_be_worth_performing(tax_meat,demand_falls)` |
| b+ | Pro: AS:bridge_possible_good, A `tax_meat`, G `demand_falls` | concludes `should_be_performed(tax_meat)` |
| m- | Con: AS:negative_consequences_modal, A `tax_meat`, G `demand_falls` | concludes `may_be_unwise_to_perform(tax_meat,demand_falls)` |
| b- | Con: AS:bridge_possible_harm, A `tax_meat`, G `demand_falls` | concludes `¬should_be_performed(tax_meat)` |
| c3 | Pro: AS:consequences, A `tax_meat`, G `demand_falls` | concludes `should_be_performed(tax_meat)` |
| n3 | Con: AS:negative_consequences, A `tax_meat`, G `demand_falls` | concludes `¬should_be_performed(tax_meat)` |

`demand_falls` is one G for both sides: a good to one (fewer emissions) and a harm to the other
(fewer farm jobs), so that both bridges share the exception
`worth_finding_out_first(tax_meat,demand_falls)`, and c3 and n3 the exception
`too_slight(tax_meat,demand_falls)`. A label asserted below is asserted on a board where no tie of
weights decides it.

## A. The backup edges

Each backup edge has an answer that must pass, one that must be refused, and one bound to the wrong
target (C2's game paragraph); what the answer leaves standing is tested with it.

### A1. AS:bad_outweighs_good at CQ:bad_outweighs (D095 (1), C003, D075)

| Case | Board, then the move | Expected |
|---|---|---|
| A1.1 | c1; Con answers CQ:bad_outweighs of c1 with AS:bad_outweighs_good, H `food_costlier` | accepted; the answer concludes `bad_outweighs(tax_meat,emissions_fall)`, In; c1 out of force; `should_be_performed(tax_meat)` Out |
| A1.2 | c1, pr; the same answer to pr's CQ:bad_outweighs (Q1-C, adverse case 6) | accepted; c1 and pr both out of force (one statement, the exception of both); `should_be_performed(tax_meat)` Out |
| A1.3 | c1, c2; A1.1's answer (adverse case 6) | accepted; c2 stays in force, so `should_be_performed(tax_meat)` stays In |
| A1.3b | c1, n1; A1.1's answer | accepted; n1 stays in force (its exception is `good_outweighs`) |
| A1.4 | c1; the answer with G `diets_healthier` (a wrong target) | refused, S9 |
| A1.5 | c1; the answer at CQ:side_effect, the conclusion's rebuttal route (Q1-D); and at CQ:too_slight | refused, S5 (no backup edge), both |
| A1.6 | c1; AS:bad_outweighs_good as a move of its own, and as a move about `bad_outweighs(tax_meat,emissions_fall)` | refused, S1, both (an answer scheme only answers) |
| A1.7 | c1; the answer with H unbound (G-unbound, in the game) | refused, S2 |
| A1.7b | (added after the first run) c1; the answer with H blank, and (after the check of step 7) H that names nothing: punctuation, a zero-width space, a byte order mark; Pro's AS:consequences with G blank | refused, S2, each |
| A1.8 | n1; Pro answers CQ:bad_outweighs of n1 (not asked of negative consequences) | refused, S3 |
| A1.9 | c1; Con answers CQ:bad_outweighs with AS:negative_consequences, G `food_costlier` (adverse case 7) | refused, S5 |
| A1.10 | A1.1, then Pro argues against `matters_more(food_costlier,emissions_fall)` with Other (adverse case 8) | accepted; the answer out of force; the exception Out; c1 in force again; `should_be_performed(tax_meat)` In; c1's other questions still open (CQ:likelihood asked is accepted) |

### A2. AS:good_outweighs_bad at CQ:good_outweighs (D095 (1), C003)

| Case | Board, then the move | Expected |
|---|---|---|
| A2.1 | n1; Pro answers CQ:good_outweighs of n1 with AS:good_outweighs_bad, H `emissions_fall` | accepted; `good_outweighs(tax_meat,food_costlier)` In; n1 out of force |
| A2.2 | n1, nm; A2.1's answer | accepted; n1 and nm both out of force (the shared exception) |
| A2.3 | n1; the answer at CQ:opposite_consequences, the conclusion's rebuttal route (Q1-B) | refused, S5 |
| A2.4 | n1; the answer with G `diets_healthier` | refused, S9 |
| A2.5 | n1; Pro answers CQ:good_outweighs with AS:consequences | refused, S5 |

### A3. AS:better_means at CQ:other_means (D089, D094 (b), the profile's P1 to P4)

| Case | Board, then the move | Expected |
|---|---|---|
| A3.1 | pr; Con answers CQ:other_means of pr with AS:better_means, A2 `subsidy_plants`, comparison sooner | accepted; `realize_more_effectively(decider,emissions_fall,tax_meat)` In; pr out of force |
| A3.2 | pr; A2 `tax_meat` | refused, P4 |
| A3.3 | pr; A2 `meat_tax` (tax_meat's alias) | refused, P4 |
| A3.4 | pr; A2 `farmers_strike` (another actor's action) | refused, P3 |
| A3.5 | pr; no comparison | refused, P4 |
| A3.6 | pr; G `diets_healthier` | refused, S9 |
| A3.7 | pr; P `farmers` | refused, P2 |
| A3.8 | pr; A2 unbound | refused, S2 |
| A3.9 | pr, c1; A3.1's answer | accepted; c1 stays in force (its exceptions do not include the better means), so `should_be_performed(tax_meat)` stays In |

### A4. AS:finding_out_first at CQ:worth_finding_out_first (D095 (3), D053, D008)

| Case | Board, then the move | Expected |
|---|---|---|
| A4.1 | m+, b+, m-, b-, c1; Con answers CQ:worth_finding_out_first of b+ with AS:finding_out_first, I `pilot_study` (Q3-A, Q3-E) | accepted; `worth_finding_out_first(tax_meat,demand_falls)` In; b+ and b- out of force; m+, m- and c1 in force, their conclusions In; `should_be_performed(tax_meat)` In (c1) |
| A4.2 | m+, b+; I unbound | refused, S2 |
| A4.2b | (added after the first run) m+, b+; I blank, and (after the check of step 7) I that names nothing; and pr with a better means whose A2 is blank or names nothing, and one that declares an action whose id names nothing | refused: S2 for I; P1 for A2 and for the declaration |
| A4.3 | m+, b+; Other as the answer; Other for the exception as a move about it; Other concluding it as a new move; Other presuming it as a premise; and, once Pro has argued against the exception with Other (accepted), Other against that negation (Q3-B) | refused, S5, every form |
| A4.4 | m+, b+; Con answers with AS:expert_opinion that the exception holds, then Pro asks CQ:expertise of that answer (Q3-C, adverse case 7) | both accepted; while the question is pending the expert's argument is out of force, the exception Out and b+ in force again |
| A4.5 | m+, b+, and a bridge about G `prices_rise` with its modal argument; A4.1's answer (Q3-D) | accepted; the bridge about `prices_rise` stays in force |
| A4.6 | A4.1, then Pro asks CQ:premise_acceptable of `settles_whether(pilot_study,tax_meat,demand_falls)`, one of the answer's three conditions | accepted; while pending, the answer is out of force, the exception Out, b+ and b- in force again |
| A4.7 | m+, b+; CQ:worth_finding_out_first asked alone (Q3-F) | refused, Q1 |
| A4.8 | m+, b+; Con answers with AS:sign, and with AS:analogy | refused, S5, both |
| A4.9 | m+, b+; Con asks CQ:premise_acceptable of `matters_to_decision(tax_meat,demand_falls)` (C3) | accepted; b+ out of force while pending; Pro's argument for the premise answers it and b+ is in force again |
| A4.10 | m+; Pro supports `should_be_performed(tax_meat)` with AS:consequences_modal (C3) | refused, S9 |
| A4.11 | b+ alone, then m+ (D008, D094 (g)) | b+ accepted with its entry premise `may_be_worth_performing(tax_meat,demand_falls)` Out (not presumed), `should_be_performed(tax_meat)` not In; after m+ it is In |

Recorded after the first run (2026-11-27): A4.10 was refused by S7, "statement
`should_be_performed(tax_meat)` is not on the board": with m+ alone no argument has put the statement
on the board, and S7 comes before S9. The page was wrong, not the game: a move about a statement
needs the statement. The case is played on m+ and b+ (b+ concludes it), and is refused by S9.

### A5. AS:sign's three backup edges (D009; owner's question 3)

| Case | Board, then the move | Expected |
|---|---|---|
| A5.1 | c1; Con answers CQ:side_effect of c1 with AS:sign | accepted, a rebuttal: the answer concludes `¬should_be_performed(tax_meat)` |
| A5.2 | pr; Con answers CQ:possibility of pr with AS:sign, against | accepted; `possible_bring_about(decider,tax_meat)` Out; pr out of force |
| A5.3 | pg; Con answers CQ:goal_possible of pg with AS:sign, against | accepted; `possible_goal(emissions_fall)` Out; pg out of force |
| A5.4 | c1; Con answers CQ:likelihood of c1 with AS:sign | refused, S5 |

### A6. AS:consequences at CQ:opposite_consequences (D009)

| Case | Board, then the move | Expected |
|---|---|---|
| A6.1 | n1; Pro answers CQ:opposite_consequences of n1 with AS:consequences, G `emissions_fall` | accepted, a rebuttal: the answer concludes `should_be_performed(tax_meat)` |
| A6.2 | n1; Pro answers it with AS:negative_consequences | refused, S5 |

## B. Other and the thesis-independent schemes

| Case | Board, then the move | Expected |
|---|---|---|
| B1 | c1; Con answers CQ:too_slight of c1 with Other (D053, D065) | accepted; `too_slight(tax_meat,emissions_fall)` In; c1 out of force |
| B2 | c3, n3; Con answers CQ:too_slight of c3 with Other (adverse case 6: too_slight defeats both sides) | accepted; c3 and n3 both out of force |
| B3 | B1, then Pro asks CQ:relevance of Con's Other argument (D074, Q2-A, adverse case 7) | accepted; Con's argument out of force while pending; the exception Out; c1 in force again |
| B4 | c1; Con answers CQ:bad_outweighs with Other whose premise is its own conclusion, `bad_outweighs(tax_meat,emissions_fall)` | refused: the move makes a cycle |
| B4b | c1; Con answers CQ:bad_outweighs with Other whose premise is `should_be_performed(tax_meat)`, the conclusion the answer defeats | refused: the move makes a cycle (through c1's exception) |
| B5 | c1; Con answers CQ:likelihood of c1 with AS:expert_opinion, against; then Pro asks CQ:field of it (adverse case 7) | both accepted; after the answer `brings_about(tax_meat,emissions_fall)` Out and c1 out of force; while CQ:field is pending the expert's argument is out of force and c1 in force again |
| B6 | (Q2-A to Q2-C) Pro: Other concluding `brings_about(tax_meat,emissions_fall)` from its being desirable; c1; Pro: AS:expert_opinion for `brings_about(tax_meat,emissions_fall)`; then Con undercuts the Other argument, answering CQ:relevance with Other | accepted; the Other argument out of force; c1 in force (Q2-B); `brings_about(tax_meat,emissions_fall)` In by the expert's argument (Q2-C) |

Recorded after the independent check of step 7 (2026-11-27): B6 as written played the Other argument
before c1, so its conclusion was a statement of its own words (the formula was not yet on the
board), not c1's premise, and the test passed whatever the game did about Q2-B and Q2-C. The page
was wrong. B6 is now played as: c1; Con asks CQ:likelihood, so that the premise holds only by the
arguments for it; Pro's Other argument for c1's premise; Con's undercutter, answering CQ:relevance
with Other (Q2-A: the Other argument out of force and the premise Out, while c1's own undercutter is
not In: Q2-B); Pro's AS:expert_opinion for the premise (Q2-C: the premise In and c1 in force, the
Other argument still out of force).

Recorded after the first run (2026-11-27): B4 was refused, but by the adapter, "the premise
`bad_outweighs(tax_meat,emissions_fall)` is the conclusion itself", before the graph's cycle check,
which B4b's longer cycle (through c1's exception) reaches. The game is right: the circular answer is
refused and the graph restored; the page named the wrong check. B4 now expects the adapter's
refusal.

## C. Questions asked alone (D003, D075; adverse case 5)

| Case | Board, then the move | Expected |
|---|---|---|
| C1 | c1; Con asks CQ:likelihood | accepted; `brings_about(tax_meat,emissions_fall)` Out while pending; c1 out of force; its exceptions and the other arguments as before |
| C2 | c1; Con asks CQ:likelihood and CQ:evidence; Pro supports `brings_about(tax_meat,emissions_fall)` with AS:expert_opinion | all accepted; the one argument answers both: the premise In, c1 in force |
| C3 | pr; Con asks CQ:possibility (an assumption) | accepted; it ties the presumption: `possible_bring_about(decider,tax_meat)` Out; pr out of force |
| C4 | c1, b+ with m+; CQ:too_slight, CQ:bad_outweighs, CQ:gain_discounted (of c1) and CQ:worth_finding_out_first (of b+) asked alone | refused, Q1, each |
| C5 | c1, n1; CQ:side_effect of c1 and CQ:opposite_consequences of n1 asked alone | refused, Q1, both |
| C6 | B1, then Pro argues against `too_slight(tax_meat,emissions_fall)` with Other (adverse case 5: a tie) | accepted; the exception not In; c1 in force |
| C7 | C1, then Con asks CQ:likelihood of c1 again | refused, Q2 |

## D. The direct rebuttal where an answer route is removed

| Case | Board, then the move | Expected |
|---|---|---|
| D1 | c1; Con answers CQ:side_effect of c1 with AS:negative_consequences, G `food_costlier` (owner's question 1) | refused, S5 |
| D2 | c1; Con argues the same against `should_be_performed(tax_meat)`, a move about the statement | accepted, a rebuttal |
| D3 | D2 (a slight harm; Q1-E) | `bad_outweighs(tax_meat,emissions_fall)` is not In, and c1 is not out of force: the con argument does not establish the weighing |

## E. The profile (D094 (b), D063, D095 (4))

| Case | Board, then the move | Expected |
|---|---|---|
| E1 | Pro: AS:consequences with A `farmers_strike` (E-agent: another actor's action) | refused, P3 |
| E2 | Pro: AS:consequences with A `ban_meat`, declared by no one and by no move | refused, P1 |
| E3 | Pro: AS:consequences with A `meat_tax` (the alias) | accepted; the argument is about `tax_meat` (the alias reads as its action) |

E1's outcome follows P3's reading of D063 (the actions argued are the agent's), which is the
owner's question 2 (c5-game.md): under D094 (b)'s narrower words a consequence argument has no agent
place, and E1 would be accepted.

D095 (4)'s wrong actor, action and goal are A3.4, A3.7, A3.2, A3.6, A1.4 and E1; its unrelated
consequences A1.3 and A4.5; its surviving independent support A1.3, A3.9, A4.1 and B6. Its
wrong context (the situation) is shown and recorded, not checked move by move (D089), so no move is
refused for it; F-context below is the bridge that carries another actor's recommendation.

## F. Schemes the game does not play in C5.2 (engine and validate.py)

`bridge_from_continuing`, `bridge_from_agent` and `end_and_means_bad` are walton.yml's, not the
game's, so their cases are tests of the Carneades engine on walton.yml's schemes and of
`validate.py` on annotation records.

| Case | Where | Expected |
|---|---|---|
| F-continuing | engine | an argument from `bridge_from_continuing` (P, G, A) is out of force once `realize_more_effectively(P,G,A)` is In: the better-means exception applies to the continuing bridge |
| F-context | validate.py | an instance of `bridge_from_agent` whose `deliberating_agent(P,A)` names an A that the case record gives to another actor is refused (R14) |
| G-endandmeans | engine | an argument from `end_and_means_bad` (X, Y) is out of force once `means_good_outweighs(X,Y)` is In, and stands otherwise; walton.yml's indexed sentences of `means_good_outweighs/2` and `means_bad_outweighs/2` render their places in order |
| G-unbound | validate.py, engine | an instance of `end_and_means_bad` with X or Y unbound is refused; the engine does not read an unbound variable as a wildcard |

Recorded after the first run (2026-11-27): F-continuing, G-endandmeans and F-context passed as
written, and G-unbound for an X or Y left out. A blank X or Y was another matter: the engine read it
as the atom `x_empty`, and so did the game, which accepted an answer of `bad_outweighs_good` with a
blank H (its premises then spoke of `x_empty`) and AS:consequences with a blank G; a blank I would
have let `finding_out_first` stand without naming any investigation, the bare claim D053 refuses.
The game was wrong: S2 checked that a binding was there, not that it named something. S2 and the
adapter now refuse a blank binding (A1.7b, A4.2b, and the engine's G-unbound).

## Not here

- Replay-1, the 33 reference games on the frozen data (TestReferenceDigests), and Replay-2, the new
  golden set on the shipped data with a profile (step 7's second part).
- Recording-1 (a paid reply recovered after a crash, with no second call) belongs to G, the journal.
- Every answer offered at every question concluding what its locus and polarity call for is
  TestEveryEdgeCase's (the 243 answers of `docs/scheme-registry/nascq-edges.md`).

## Planted defects

Each defect is planted on a copy of the tree, and the tests named must fail:

| Defect | Caught by |
|---|---|
| S1 lets an answer scheme make a move of its own | A1.6 |
| S2 reads a missing binding as empty | A1.7, A3.8, A4.2 |
| S3 lets a scheme answer a question its target is not asked | A1.8 |
| S5 skips the backup-edge check | A1.5, A1.9, A2.3, A2.5, A4.8, A5.4, A6.2, D1 |
| S5 lets Other answer the question that refuses it, or argue its exception | A4.3 |
| S9 takes the answer's own G instead of the exception's | A1.4, A2.4, A3.6 |
| S9 lets a modal argument support `should_be_performed(A)` | A4.10 |
| Q1 lets an exception's or conclusion's question be asked alone | A4.7, C4, C5 |
| Q2 lets a question be asked twice | C7 |
| P1 lets an undeclared action pass; an alias is not read as its action | E2, E3 |
| P2 lets a move bind P to another actor | A3.7 |
| P3 lets a move argue another actor's action | A3.4, E1 |
| P4 lets A2 be A or its alias, or a better means without its comparison | A3.2, A3.3, A3.5 |
| the cycle check is skipped | B4b |
| the adapter takes a premise that is the argument's own conclusion | B4 |
| a pending premise question does not tie the premise's presumption | C1, C3, A4.6, A4.9 |
| a pending inference question does not put the argument out of force | B3 |
| a bridge's entry premise is presumed | A4.11 |
| the shared exception is two statements (by scheme), not one | A1.2, A2.2, A4.1, B2 |
| an exception defeats arguments about another G | A1.3, A1.3b, A4.5 |
| S2 takes a blank binding | A1.7b, A4.2b |
| the adapter takes a blank binding | G-unbound (engine) |

Recorded after the run of the planted defects (2026-11-27): each row above was planted. One survived
at first: with NASCQ's refusal of Other as an answer removed, the game's own check after the move is
placed refused it, also by S5, so A4.3 now asks the answer form to be refused by the question's own
check ("refuses Other as an answer"). Two did not build at first (a variable left unused) and were
rewritten. The record is in the workspace (verify/c5/c52/itemf).

Corrected after the independent check of step 7 (2026-11-27):

- The engine's two defects (an exception statement made per scheme, and one that ignores its G) are
  caught by the engine's own admission check at each case's first move, not by the cases named. The
  shared exception is planted in the data instead, as an exception written with its places swapped
  in one scheme of `core.yml`: `bridge_from_possible_harm`'s, caught by A4.1;
  `negative_consequences`' `too_slight`, caught by B2; `negative_consequences_modal`'s
  `good_outweighs`, caught by A2.2.
- No planted defect was found that makes an exception defeat an argument about another G and
  nothing else: an exception is a statement keyed by its whole formula, and a defect in that key (the
  engine's second defect) fails at the engine's admission check first. A1.3, A1.3b and A4.5 assert
  it.
- The planted rules are S1, S2, S3, S5 and S9 of NASCQ's, not "S1 to S9".
- Rows added with the fixes of the check:

| Defect | Caught by |
|---|---|
| an argument for a premise answers no pending question about it | C2, A4.9 |
| S2 takes a value that names nothing (punctuation, invisible) as a name | A1.7b, A4.2b |
| P1 resolves an action value that names nothing | A4.2b |
| the adapter's AddSupport takes a value that names nothing | G-unbound (engine) |

