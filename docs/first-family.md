# The first family: practical reasoning and the consequences

*Version 3, 2026-11-25, after C3 (proposal D008-D074) was applied; version 2 on 2026-11-24, after C2 (proposal D004-D075); version 1 the same day. The inventory of stage 2's slice (plan §0): what it includes, what it leaves out and why, what each scheme still needs, and which proposal closes it. Item C of the critical review's execution sheet asks for this inventory; each proposal that lands updates it.*

## The slice

The first family is two of the scheme families (`wag/data/schemes/families.json`): **F10, practical reasoning**, and **F15, consequences, alternatives, threat and fear**. Together they have 39 schemes in `walton.yml`, three of them the answer schemes C2 added, five the schemes C3 added (the modal variants of 33.1 and 33.2, their two bridges, and the argument from what happens anyway) and five the schemes C4 added (the negative and graded readings of ends and means, the commitment form of sunk costs, and two bridges from conclusions that name an agent), and one the answer scheme for the bridges' exception that D095 added (`finding_out_first`). The game plays four of them.

| Family | Scheme | Compendium | In the game | Native diagnostics | Decisions still to apply |
|---|---|---|---|---|---|
| F10 | practical_inference | 22.1 | — | — | — (C2) |
| F10 | practical_necessary_condition | 22.2 | — | — (C2 states P and removes A2) | — (C2) |
| F10 | practical_sufficient_condition | 22.3 | — | — (C2 states P and removes A2) | — (C2) |
| F10 | practical_reasoning1 | 22.4 | AS:practical_reasoning | — | D003, D007; D004, D005, D070, D088, D089, D091 in the game's copy (C5) |
| F10 | practical_reasoning2 | 22.1 | AS:practical_reasoning_goal | — | D003, D007; D004, D005, D006, D070, D088, D089, D091 in the game's copy (C5) |
| F10 | better_means | — | — | — | — (an answer scheme, C2) |
| F10 | practical_goal | 22.5 | — | — | — (C4) |
| F10 | end_and_means_1 | 22.6.1 | — | — | — (C4) |
| F10 | end_and_means_2 | 22.6.2 | — | — | — |
| F10 | end_and_means_3 | 22.6.3 | — | — | — |
| F10 | two_person_practical | 23 | — | — | — (C4) |
| F10 | practical_analogy_positive | 8.1 | — | — | — |
| F10 | practical_analogy_negative | 8.2 | — | — | — |
| F10 | negative_practical | 26.2 | — | — | — |
| F10 | waste | 24 | — | — | — (C4) |
| F10 | sunk_costs | 25 | — | — | — (D031 by a scheme of its own, C4) |
| F15 | positive_consequences | 33.1 | AS:consequences | — | D009; D065 and D095's weighing question in the game's copy (C5) |
| F15 | negative_consequences | 33.2 | AS:negative_consequences | — | D009; D065 and D095's weighing question in the game's copy (C5) |
| F15 | negative_consequences_reasoning | 33.3 | — | — | — |
| F15 | negative_consequences_prudential | 33.4 | — | — | — (C3) |
| F15 | pragmatic_alternative | 34 | — | — | — (C4) |
| F15 | threat | 35 | — | — | — (C4) |
| F15 | ad_baculum | 35.1 | — | — | — (C4) |
| F15 | fear | 36 | — | — | — (C4) |
| F15 | danger | 37 | — | — | — (C4) |
| F15 | slippery_slope_base_case | 49 | — | — | — |
| F15 | bad_outweighs_good | — | — | — | — (an answer scheme, C2) |
| F15 | good_outweighs_bad | — | — | — | — (an answer scheme, C2) |
| F15 | positive_consequences_modal | 33.1 (in part) | — | — | — (a modal variant, C3; the game at C5) |
| F15 | negative_consequences_modal | 33.2 (in part) | — | — | — (a modal variant, C3; the game at C5) |
| F15 | bridge_from_possible_good | — | — | — | — (a bridge, C3; the game at C5) |
| F15 | bridge_from_possible_harm | — | — | — | — (a bridge, C3; the game at C5) |
| F15 | harm_reduction | — | — | — | — (C3) |
| F15 | finding_out_first | — | — | — | — (an answer scheme, D095; the game at C5) |
| F10 | end_and_means_bad | 22.6.1 (in part) | — | — | — (a variant, C4) |
| F10 | end_and_means_less_bad | 22.6.1 (in part) | — | — | — (a variant, C4) |
| F10 | sunk_costs_precommitment | 25 | — | — | — (C4) |
| F10 | bridge_from_agent | — | — | — | — (a bridge, C4) |
| F10 | bridge_from_continuing | — | — | — | — (a bridge, C4) |

The diagnostics are the engine's own (its automatic-inference profile): a variable that no premise binds. The two condition schemes had the slice's only ones, as the critical review of 2026-11-23 found, and C2 removed them on 2026-11-24; outside the slice, `ad_hominem` (S) and `regulative_rule` (A) have the same problem and go with their families.

Decisions that apply to the slice as a whole: D063 (practical conclusions meet `should_be_performed`, with bridges for conclusions that name an agent), D064 (C003, an answer scheme from a consequence of the other sign that matters more: applied by C2), D065 (C005 and C021, how serious and on whom: applied by C3), D066 (the action itself: mitigation, another action, feasibility: coding rules, C3), D067 (rights and values stay apart: no change), D069 (harms of not acting: applied by C3), D070 (a side effect that outweighs), D072 (C022, what happens anyway: applied by C3), D074 (C011, the appeal to consequences as an objection: an undercutter of the argument it answers, by C3), D075 (answer schemes). D008 (modal consequences: applied by C3) and D009 (the game's question texts) concern the consequence schemes; D007 (the wording) and D009 the game's data.

## The first profile

What the bounded pilot (stage 3) can play and measure, once stage 2 is done.

- **Included:** the four schemes the game plays: `practical_reasoning1`, `practical_reasoning2`, `positive_consequences`, `negative_consequences`; and the answer schemes D075 names for them, in `walton.yml` since C2: `bad_outweighs_good` (C003's consequence of the other sign, D064, which is also the side effect that outweighs, D070), `good_outweighs_bad` and `better_means`; and the modal variants of the two consequence schemes with their bridges, in `walton.yml` since C3 (D008), which join the game at step C5, with `finding_out_first`, the answer scheme for the bridges' exception (D095, in `walton.yml` since 2026-11-26).
- **Left out, and said so:** the other 27 schemes of the two families. Their decided changes are applied all the same (proposals C3 and C4 below), so that the family's formulas are closed; they are not yet questions and moves of the game, and a pilot that reports on the first profile reports nothing about them. Taking one into the game is the next slice's step (plan §6, order 2).
- **Outside the slice, still in the game:** expert opinion, position to know, sign, analogy, values and Other. They stay playable; their results are not the first family's, and a report keeps them apart.

## The questions of the included schemes, today

| Scheme | Question | Locus | Pattern |
|---|---|---|---|
| AS:practical_reasoning | CQ:other_goals | conclusion | — |
| AS:practical_reasoning | CQ:other_means | exception | realize_more_effectively(G,A) |
| AS:practical_reasoning | CQ:side_effect | exception | side_effects(A,S1) |
| AS:practical_reasoning | CQ:possibility | premise | possible(A) |
| AS:practical_reasoning | CQ:best_means | exception | bring_about_more_effectively(S2,A) |
| AS:practical_reasoning_goal | CQ:conflicting_goals | exception | incompatible_goal(G) |
| AS:practical_reasoning_goal | CQ:other_means | exception | realize_more_effectively(G,A) |
| AS:practical_reasoning_goal | CQ:best_means | exception | bring_about_more_effectively(S2,A) |
| AS:practical_reasoning_goal | CQ:possibility | premise | possible(A) |
| AS:practical_reasoning_goal | CQ:goal_possible | premise | possible(G) |
| AS:practical_reasoning_goal | CQ:intervening_actions | exception | intervening_actions(A,G) |
| AS:practical_reasoning_goal | CQ:side_effect | exception | side_effects(A,S1) |
| AS:consequences | CQ:evidence | argument | — |
| AS:consequences | CQ:side_effect | conclusion | — |
| AS:consequences | CQ:likelihood | premise | brings_about(A,G) |
| AS:negative_consequences | CQ:likelihood | premise | brings_about(A,G) |
| AS:negative_consequences | CQ:evidence | argument | — |
| AS:negative_consequences | CQ:opposite_consequences | conclusion | — |

One backup edge touches the slice: positive consequences answers the opposite-consequences question. The decisions change this table, and C2 has made their formulas in `walton.yml` (step C5 brings them to the game): D004 gives practical_reasoning1 the conflicting-goal exception, so CQ:other_goals targets it instead of the conclusion; D070 renames the side-effect exception to a side effect that outweighs, answered by its answer scheme, which D095 (1) gives a weighing question of its own while CQ:side_effect stays the rebuttal route at the conclusion; D088 and D089 scope the possibility and better-means questions to the agent; D065 adds how serious and on whom to the consequence schemes, in `walton.yml` since C3 (`too_slight(A,G)` and the bearer exception of each side), with game questions at step C5; D074's objection, that a consequence argument is offered for a claim's truth, undercuts the argument it answers: in the game it is Other's relevance question with the direct undercutter of that inference, from step C5, and no consequence scheme asks it (D095 (2)). Each question of the included schemes then gets the CQ answer contract's cases (item F): a legitimate answer accepted, an irrelevant, circular or mismatched one refused, and the other challenges and independent support surviving.

## Where the game's copies differ from the canonical formulas

The game's `core.yml` holds its own copies of the four schemes. `positive_consequences` and `negative_consequences` there lack the canonical exceptions (`bad_outweighs(A,G)` and `good_outweighs(A,G)`, from D055, and since C3 `too_slight(A,G)` and the bearer exception of their side); `practical_reasoning1` and `practical_reasoning2` there lack what C2 gave the canonical ones on 2026-11-24: the agent P, the goal premise, the possibilities apart, the scoped exceptions and `bad_outweighs(A,G)` for the side effect. The catalogue records each difference (game-differs), and its check warns where a game question targets the old formula. Step C5 brings the game's copies and its NASCQ to the canonical formulas of the slice once the proposals have landed, with the game's tests and the replay of the saved games.

## The proposals that close the slice

| Step | Decisions | Schemes |
|---|---|---|
| C2, practical reasoning: **done 2026-11-24** ([proposal D004-D075](proposals/D004-D075.md)) | D004, D005, D030 (with D063 for the two condition schemes and practical inference), D064 and D070, D075's first answer schemes, and the owner's D088-D091; D003 and D006 change nothing in `walton.yml` and are recorded so | practical_reasoning1, practical_reasoning2, practical_necessary_condition, practical_sufficient_condition, practical_inference, and the new answer schemes |
| C3, consequences: **done 2026-11-25** ([proposal D008-D074](proposals/D008-D074.md)) | D008, D065, D066, D069, D072, D074 (D064 was applied with C2); D067 changes nothing; D053's declarations, by the owner | positive_consequences, negative_consequences, negative_consequences_prudential, and the new schemes positive_consequences_modal, negative_consequences_modal, bridge_from_possible_good, bridge_from_possible_harm and harm_reduction |
| C4, the rest: **done 2026-11-26** ([proposal D022-D093](proposals/D022-D093.md)) | D063's conclusions and bridges, D022, D023, D031, D033, with D060, D062 and D093 | fear, danger, threat, ad_baculum, pragmatic_alternative, practical_goal, two_person_practical, waste, sunk_costs, end_and_means_1 |
| C5, the game ([plan](c5-game.md)) | D007, D009, D075's backup edges; game questions for C3's exceptions (too_slight, the bearer pair, worth_finding_out_first) and the weighing questions (D095 (1)); D074 through Other's relevance question (D095 (2)); the answer scheme for worth_finding_out_first in the game (in `walton.yml` since 2026-11-26, proposal D095), and Other refused there (D053, D095 (3), item F) | the four game schemes and their answer schemes; the modal variants and their bridges (D008) |

Each proposal goes the way of process.md §5: the proposal file and its view, `wag-cli --proposal`, the cleaning's gate and the registry with it applied, independent review until it is clean, then `edits.py`, the catalogue and the register. The decisions are taken (most under the owner's standing rule of 2026-11-22); a departure from a decision's text goes to the owner as a choice of its own.
