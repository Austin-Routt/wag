# Adjudication of Kialo batch 2

`gold.json` holds one gold annotation per claim, in `items.json` order; each item's `adjudication` says what changed and why. `lexicon.json` holds the shared atoms. `python3 -I` validation passes: order; walton.yml IDs and exact bindings; every premise, assumption, exception and conclusion is the instantiated walton.yml formula; R1, R2, R4, R7; target formulas in the target's primary instance; every atom in the lexicon.

## 1. Counts

45 measured disagreements (`agreement.json`):

| Field | Disagreements | Resolved for A | Resolved for B | Neither |
|---|---|---|---|---|
| scheme | 10 | 8 (145, 2, 4, 46, 147, 194, 202, 235) | 2 (135, 8) | 0 |
| fit | 9 | 6 (2, 4, 46, 202, 224, 233) | 3 (104, 135, 29) | 0 |
| role | 7 | 0 | 7 (33, 192, 204, 233, 29, 182\*, 198\*) | 0 |
| polarity | 2 | 0 | 2 (46, 190) | 0 |
| target claim | 1 | 0 | 1 (29) | 0 |
| target part | 1 | 0 | 1 (192) | 0 |
| re_parented | 13 | 0 | 13 (all, by R1) | 0 |
| evidence | 2 | 0 | 2 (12, 194) | 0 |
| **total** | **45** | **14** | **31** | **0** |

\* B's kind is kept; the question (CQ:likelihood on 180's primary premise, not "28 CQ3") is neither's.

Some agreed codings changed under the rules (R6):
- **R1**: re_parented true → false in 6 items (11, 6, 182, 198, 204, 208).
- **R2**: role → cq_answer in 8 items.
  - The side_effect roles of 174 and 5 → CQ:side_effect.
  - The alternative_means roles of 104, 147 and 123 → CQ:other_means.
  - The premise_backing roles of 8 and 12 → CQ:likelihood (for), and of 141 → 14 CQ1.
  - The finer label is kept in `role.move`.
- **R5 targets**: 12, 141, 182 and 198 targeted a causal step that is not the target's primary instance; they now target the primary premise it supports.
- **R4**: 11 cause_to_effect steps with an un-enacted cause (9, 12, 14, 36, 180, 192) have fit reconstructed, confidence ≤ 0.5 and the N8 note. A's step in 6 was dropped: the primary fragment states the same link.
- **Secondary instances**:
  - Kept: A's in 104 (alternatives, 13.2) and 12 (negative consequences); B's in 8 (expert opinion) and 4 (step 1).
  - Replaced: B's opposition-normative in 123 (§2.7).
- **R5 lexicon**: 102 atoms. `reduced_meat_consumption` replaces five annotator atoms (less_meat_produced_and_consumed, reduced_meat_demand, …); `tax_meat` replaces meat_tax_imposed; `higher_meat_price` replaces three price atoms.

Overall, 34 items changed in some field; 16 are "agreed" apart from atom normalisation and R4 notes.

## 2. Kinds of disagreement and v3 changes

1. **Re-parenting (13, plus 6 agreed).** Set whenever the target was a premise.
   *v3:* `re_parented = (actual.claim ≠ Kialo parent)`, computed by the tooling.
2. **Premise attack vs CQ answer (6 role disagreements, 8 agreed roles changed).** Guidelines §2 (item 4) lists cq_answer but gives no test.
   *v3:* ship a table (scheme × formula → compendium CQ, game CQ) from the cq_map and edges.json; cq_answer exactly when the targeted formula is in it; side_effect and alternative_means become `role.move` values. Settle conclusion-located CQs: 121 (side_effect, though AS:consequences has CQ:side_effect) and 43 (cq_answer, CQ:opposite_consequences) are one situation coded two ways; I propose cq_answer for both.
3. **Practical reasoning vs positive consequences (145, 147, 194, 235; also 46, 202).** B coded practical_reasoning2 whenever a proposal had a purpose, taking S1 from the parent, supplying S2→G, or using the action's own description as S2.
   *v3:* practical reasoning only when S2 is a state distinct from the action and at most one of S1, A→S2, S2→G is supplied. A proposal with a purpose is positive_consequences. A bare proposal is `none` with role action_amendment.
4. **Which instance is primary (2, 4; targets in 12, 141, 182, 192, 198).**
   *v3:* the primary instance is the one whose conclusion reaches the target (for causal facts backing an evaluative premise, the classification; the chain is secondary). Targets must be formulas of the target's primary instance.
5. **Fit (9).** B marked 224 and 233 explicit with supplied premises; A marked 29 reconstructed with all premises stated.
   *v3:* explicit iff every `premise`-role formula is explicit.
6. **Small items.**
   - Amendment polarity (46, 190) is the stance on the target's conclusion; "none" is not a value.
   - A link explaining a term is `definition` (12, 194).
   - Atom drift (binding-word Jaccard 0.52 on same-scheme items): issue `lexicon.json` with the batch.
7. **Sources.**
   - walton.yml `opposition-normative` uses X2 in minor premise and conclusion, so every instance contradicts itself; fix: `receives_evaluation(A,X)`.
   - `to_wag.py` reads a target formula only from `part: "premise: <formula>"`; annotators and gold use `part` plus `formula`.

## 3. Candidate patterns

| Pattern | Claims | Covered? |
|---|---|---|
| Mitigation by substitution (a substitute makes the cited consequence avoidable → ¬brings_about) | 33, 182, 198; by adaptation (cost is small → ¬bad): 25 | No (`other`). Nearest: 22.1 from the affected party's view; `alternative_means/2`, an exception of 22.2/22.3 (F10). Best new-scheme candidate; CQs: access (206), adequacy (184), cost. |
| Weighing consequences (bad outweighs good) | 127 (explicit), 43 (values outweigh revenue) | No. `more_serious_consequences/2` occurs only in negative_practical (26.2, ignorance). The game asks it only as CQ:opposite_consequences. |
| Absent exposure (affected group does not consume the taxed good) | 204 | No; only defeasible_modus_ponens. |
| Hypothetical causal step | 11 instances in 9, 12, 14, 36, 180, 192 | Partly. cause_to_effect (F13) asserts has_occurred; a modal variant is needed (audit N8). |
| Action amendment: "A plus B" | 145, 194, 45, 235 | A move, not a scheme; coded 33.1, 22.2 or `none`. No game move. |
| Action amendment: "A only for X" | 46, 190, 202 | Same; coded 55.2, 22.1, `none`. |
| Alternative means | 104, 147, 123 | Yes: the 22.1 exception realize_more_effectively and CQ:other_means. |
| Side effect | 174, 5, 121 | Yes: side_effects(A,S1) and CQ:side_effect. |
| Relative magnitude / priority | 135; S1 of 190 | No; 135 coded `none`. |
| Incoherent policy (tax and subsidy) | 123 | 11.2 (F06), but walton.yml's encoding is broken. |
| Precedent elsewhere | 30 | Yes: practical_analogy_positive (8.1, F04/F10). |

## 4. Schemes the gold uses that the game lacks

| walton.yml scheme | Primary | All instances | Claims (primary) |
|---|---|---|---|
| cause_to_effect | 4 | 17 | 14, 12, 192, 141 (secondary in 9, 36, 180, 11, 2, 4) |
| verbal_classification | 3 | 3 | 2, 4, 8 |
| rule | 2 | 2 | 46, 29 |
| end_and_means_1 | 1 | 1 | 3 |
| practical_analogy_positive | 1 | 1 | 30 |
| practical_necessary_condition | 1 | 1 | 45 |
| alternative_normative | 0 | 1 | (104) |
| **total** | **12** | **26** | |

Of 50 primary instances, the game's schemes cover 26 (negative consequences 9, practical_reasoning2 8, positive consequences 5, expert opinion 2, analogy 1, value 1); 12 use the schemes above, 5 are `other`, 7 `none`. Most useful additions: a predictive cause_to_effect (17 instances) and verbal_classification (links causal facts to evaluative premises).
