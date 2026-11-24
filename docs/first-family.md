# The first family: practical reasoning and the consequences

*Version 2, 2026-11-24, after C2 (proposal D004-D075) was applied; version 1 the same day. The inventory of stage 2's slice (plan §0): what it includes, what it leaves out and why, what each scheme still needs, and which proposal closes it. Item C of the critical review's execution sheet asks for this inventory; each proposal that lands updates it.*

## The slice

The first family is two of the scheme families (`wag/data/schemes/families.json`): **F10, practical reasoning**, and **F15, consequences, alternatives, threat and fear**. Together they have 28 schemes in `walton.yml`, three of them the answer schemes C2 added. The game plays four of them.

| Family | Scheme | Compendium | In the game | Native diagnostics | Decisions still to apply |
|---|---|---|---|---|---|
| F10 | practical_inference | 22.1 | — | — | — (C2) |
| F10 | practical_necessary_condition | 22.2 | — | — (C2 states P and removes A2) | — (C2) |
| F10 | practical_sufficient_condition | 22.3 | — | — (C2 states P and removes A2) | — (C2) |
| F10 | practical_reasoning1 | 22.4 | AS:practical_reasoning | — | D003, D007; D004, D005, D070, D088, D089, D091 in the game's copy (C5) |
| F10 | practical_reasoning2 | 22.1 | AS:practical_reasoning_goal | — | D003, D007; D004, D005, D006, D070, D088, D089, D091 in the game's copy (C5) |
| F10 | better_means | — | — | — | — (an answer scheme, C2) |
| F10 | practical_goal | 22.5 | — | — | D023, D063 |
| F10 | end_and_means_1 | 22.6.1 | — | — | D022, D023 |
| F10 | end_and_means_2 | 22.6.2 | — | — | — |
| F10 | end_and_means_3 | 22.6.3 | — | — | — |
| F10 | two_person_practical | 23 | — | — | D063 |
| F10 | practical_analogy_positive | 8.1 | — | — | — |
| F10 | practical_analogy_negative | 8.2 | — | — | — |
| F10 | negative_practical | 26.2 | — | — | — |
| F10 | waste | 24 | — | — | D033, D063 |
| F10 | sunk_costs | 25 | — | — | D031 |
| F15 | positive_consequences | 33.1 | AS:consequences | — | D008, D009 |
| F15 | negative_consequences | 33.2 | AS:negative_consequences | — | D008, D009 |
| F15 | negative_consequences_reasoning | 33.3 | — | — | — |
| F15 | negative_consequences_prudential | 33.4 | — | — | D069 (C016) |
| F15 | pragmatic_alternative | 34 | — | — | D063 |
| F15 | threat | 35 | — | — | D060, D063 |
| F15 | ad_baculum | 35.1 | — | — | D063 |
| F15 | fear | 36 | — | — | D063 |
| F15 | danger | 37 | — | — | D062, D063 |
| F15 | slippery_slope_base_case | 49 | — | — | — |
| F15 | bad_outweighs_good | — | — | — | — (an answer scheme, C2) |
| F15 | good_outweighs_bad | — | — | — | — (an answer scheme, C2) |

The diagnostics are the engine's own (its automatic-inference profile): a variable that no premise binds. The two condition schemes had the slice's only ones, as the critical review of 2026-11-23 found, and C2 removed them on 2026-11-24; outside the slice, `ad_hominem` (S) and `regulative_rule` (A) have the same problem and go with their families.

Decisions that apply to the slice as a whole: D063 (practical conclusions meet `should_be_performed`, with bridges for conclusions that name an agent), D064 (C003, an answer scheme from a consequence of the other sign that matters more: applied by C2), D065 (C005 and C021, how serious and on whom), D066 (the action itself: mitigation, another action, feasibility), D067 (rights and values stay apart: no change), D069 (harms of not acting), D070 (a side effect that outweighs), D072 (C022, what happens anyway), D074 (C011, the appeal to consequences as an objection), D075 (answer schemes). D008 (modal consequences) and D009 (the game's question texts) concern the consequence schemes; D007 (the wording) and D009 the game's data.

## The first profile

What the bounded pilot (stage 3) can play and measure, once stage 2 is done.

- **Included:** the four schemes the game plays: `practical_reasoning1`, `practical_reasoning2`, `positive_consequences`, `negative_consequences`; and the answer schemes D075 names for them, in `walton.yml` since C2: `bad_outweighs_good` (C003's consequence of the other sign, D064, which is also the side effect that outweighs, D070), `good_outweighs_bad` and `better_means`.
- **Left out, and said so:** the other 21 schemes of the two families. Their decided changes are applied all the same (proposals C3 and C4 below), so that the family's formulas are closed; they are not yet questions and moves of the game, and a pilot that reports on the first profile reports nothing about them. Taking one into the game is the next slice's step (plan §6, order 2).
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

One backup edge touches the slice: positive consequences answers the opposite-consequences question. The decisions change this table, and C2 has made their formulas in `walton.yml` (step C5 brings them to the game): D004 gives practical_reasoning1 the conflicting-goal exception, so CQ:other_goals targets it instead of the conclusion; D070 renames the side-effect exception to a side effect that outweighs, answered by its answer scheme; D088 and D089 scope the possibility and better-means questions to the agent; D065 adds how serious and on whom to the consequence schemes; D074 adds the objection that a consequence argument is offered for a claim's truth. Each question of the included schemes then gets the CQ answer contract's cases (item F): a legitimate answer accepted, an irrelevant, circular or mismatched one refused, and the other challenges and independent support surviving.

## Where the game's copies differ from the canonical formulas

The game's `core.yml` holds its own copies of the four schemes. `positive_consequences` and `negative_consequences` there lack the canonical exceptions (`bad_outweighs(A,G)` and `good_outweighs(A,G)`, from D055); `practical_reasoning1` and `practical_reasoning2` there lack what C2 gave the canonical ones on 2026-11-24: the agent P, the goal premise, the possibilities apart, the scoped exceptions and `bad_outweighs(A,G)` for the side effect. The catalogue records each difference (game-differs), and its check warns where a game question targets the old formula. Step C5 brings the game's copies and its NASCQ to the canonical formulas of the slice once the proposals have landed, with the game's tests and the replay of the saved games.

## The proposals that close the slice

| Step | Decisions | Schemes |
|---|---|---|
| C2, practical reasoning: **done 2026-11-24** ([proposal D004-D075](proposals/D004-D075.md)) | D004, D005, D030 (with D063 for the two condition schemes and practical inference), D064 and D070, D075's first answer schemes, and the owner's D088-D091; D003 and D006 change nothing in `walton.yml` and are recorded so | practical_reasoning1, practical_reasoning2, practical_necessary_condition, practical_sufficient_condition, practical_inference, and the new answer schemes |
| C3, consequences | D008, D065, D066, D069, D072, D074 (D064 was applied with C2); D067 changes nothing | positive_consequences, negative_consequences, negative_consequences_prudential, and new schemes |
| C4, the rest | D063's conclusions and bridges, D022, D023, D031, D033 | fear, danger, threat, ad_baculum, pragmatic_alternative, practical_goal, two_person_practical, waste, sunk_costs, end_and_means_1 |
| C5, the game | D007, D009, D075's backup edges | the four game schemes and their answer schemes |

Each proposal goes the way of process.md §5: the proposal file and its view, `wag-cli --proposal`, the cleaning's gate and the registry with it applied, independent review until it is clean, then `edits.py`, the catalogue and the register. The decisions are taken (most under the owner's standing rule of 2026-11-22); a departure from a decision's text goes to the owner as a choice of its own.
