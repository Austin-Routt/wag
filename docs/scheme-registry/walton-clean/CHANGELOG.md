# walton.yml cleanup: changelog

Input: `walton.yml` from `walton-argumentation-schemes-for-carneades-4-master` (2068 lines, 111 schemes, SHA-256 `1976202ac6bd538526ea75488659ec9880641ec03534d6e098f6f6dd1bcc099c`). Output: `walton.yml` (2067 lines, same 111 schemes in the same order, after a 5-line notice of the modification that the licence asks for, `tools/header.txt`). Every edit below is also in `changes.json` and is produced by `tools/apply_edits.py` from the list in `tools/edits.py`; `check/run.sh` rebuilds everything. Line numbers are those of the original file.

## Gate

`check/main.go` (Go, Carneades-4 fork at `carneades-4`, the repository's root), one process per file. Check 1 also decodes with yaml.v2 `UnmarshalStrict`, which rejects unknown keys (`assumption:`) and duplicate keys.

| | original | syntax fixes only | **cleaned** | cleaned + 4 proposals (not applied) |
|---|---|---|---|---|
| 1 YAML / import error | yaml: line 733: mapping values are not allowe | none | none | none |
| 1 strict-decode errors | 1 | 3 | 0 | 0 |
| 1 stderr during import | none | yes | none | none |
| 2 validator problems (before / after instantiation) | – | 138 / 147 | 17 / 17 | 0 / 0 |
| 3 schemes failing InstantiateScheme checks | – | 34 of 111 | 0 of 111 | 0 of 111 |
| 4 language problems | – | 93 | 0 | 0 |
| schemes passing checks 2–4 | – | 67 of 111 | 107 of 111 | 111 of 111 |
| **GATE** | **FAIL** | **FAIL** | **FAIL** | **PASS** |

(The validator count before instantiation for the syntax-only file includes one entry for the term-parser errors printed to stderr; the audit's 137 is the same list without it.)

**The cleaned file does not pass the whole gate.** Checks 1, 3 and 4 pass for all 111 schemes, and check 2 passes for 107. The other four fail one validator rule only: every variable must occur in a premise. In `ad_hominem`, `practical_necessary_condition`, `practical_sufficient_condition` and `regulative_rule`, a variable occurs only in the conclusion, an assumption or an exception. Fixing that means adding or changing a premise, which the rules for this cleanup forbid. See [Schemes that cannot pass without a change of meaning](#schemes-that-cannot-pass-without-a-change-of-meaning). With the four proposals there, the gate passes for all 111 (`proposals-check.txt`). These schemes still instantiate correctly (check 3), and WAG's loader does not run the validator.

WAG's own lint (`wag-cli --lint --data DIR`, with the cleaned file as `schemes/core.yml` and a copy of WAG's NASCQ data) reports **0 errors and 1 warning** for the cleaned file. The warning is `no template for ¬/1`: WAG's convention is a `¬p/n` key per negated predicate, and without one a negation reads "it is not the case that …". For the syntax-only file it reports 58 language findings (`check/wag-lint-*.txt`). Its NASCQ findings come from WAG's data, which targets `core.yml`, and are not about this file.

`check/semantic-diff.txt` compares what Carneades loads from the syntax-only file and from the cleaned file, after the renames below. The only scheme-level differences are the syntax fixes that restore a dropped formula (`popular_opinion`, `verbal_classification_arbitrariness`), the three variable declarations, and the four approval-needed repairs. The six primed-variable schemes are identical apart from the variable names.

## Diff size

| category | recorded edits | original lines touched |
|---|---|---|
| syntax | 12 | 12 |
| typo | 105 | 97 |
| variable | 24 | 22 |
| template | 46 | 44 |
| repair — needs approval | 7 | 6 |
| **total** | **194** | **169** distinct lines (a line can carry edits of several categories) |

Unified diff against the original (`walton.diff`): 166 lines removed, 165 added. That is 162 lines changed in place, 4 deleted (the duplicate keys `biased/1` and `good/1`, `posible/1`, and the conclusion that was listed as a premise) and 3 inserted (the templates `cannot_present_sametime/3` and `data_set/2`, and that conclusion under `conclusions:`).

Typo edits by kind: predicate names (incl. 1 key deleted) 48, template words 26, sources 20, titles 7, scheme ids 4. The `popular_practise` rename also renames the scheme id (counted under predicate names).

## Repairs that need the owner's approval

| scheme | line | before | after | why |
|---|---|---|---|---|
| moral_justification_ad_populum_excuse | 495, 497 | `conclusions:` empty; premise `morally_acceptable_action(X,A)` | conclusion `morally_acceptable_action(X,A)` | no conclusion, so Carneades built no argument. The formula's template starts "Therefore, …", and the category file B/4 has it as the conclusion |
| sacrifice | 908 | `great_sacrifice(S,A,X)` | `great_sacrifice(S,A,V)` | X was undeclared (non-ground statement) and V occurred only in the conclusion; the template's third argument is "the value %s" |
| rhetorical_oppositions_normative | 768 | `if_treatment_required(X,A,A',X)` | `if_treatment_required(X,A,A2,X2)` | the 4th argument is "the contrary treatment %s" (X'), not X. Otherwise X' occurs only in the conclusion (validator error). The A'→A2 rename on the same line is a plain variable fix |
| sorites_slippery_slope | 1557, 306 | premise `reapplication_sequence` (0-ary) | `reapplication_sequence(A2,A)`; template `…from the clear area -%s- through the grey area to %s.` | the conclusion's individual A occurred in no premise (validator error). The chain premise is the one that links the clear case A2 to A |

Three more edits change a declaration without touching a formula. They are recorded as `variable`, but they depart from "leave declared-but-unused variables", so the owner should see them. In `part_in_whole` (`[Y,X,P]`→`[Y,X,A]`, as in the category file A/2) and `verbal_classification_vagueness` (`[A,P]`→`[A,F]`), the unused P was the misnamed variable that the formulas use. Leaving it would keep a validator error and an extra dummy parameter. `verbal_slippery_slope` gets F declared; see the semantic concerns.

## Id map (old → new)

| old id | new id | reason |
|---|---|---|
| `popular_practise` | `popular_practice` | misspelling ('practise' for the noun 'practice'); also the predicate |
| `epistemic_ignorence` | `epistemic_ignorance` | misspelling |
| `treat` | `threat` | misspelling; the title is Argument From Threat |
| `exeptional` | `exceptional` | misspelling |
| `exeption_to_rule` | `exception_to_rule` | misspelling |

No other id changed. The ids WAG's `core.yml` uses (`practical_reasoning1`, `positive_value`, `sign`, `credible_source`, `position_to_know`, `analogy`) are untouched. Hyphenated ids (`opposition-descriptive_1/2/3`, `opposition-normative`) are valid and were left; the category files spell some of them with underscores.

### Predicate renames (old → new, everywhere they occur)

| old | new | lines |
|---|---|---|
| `posible` | `possible` | 1 |
| `all_parts_propert` | `all_parts_property` | 3 |
| `arguement_not_accepted` | `argument_not_accepted` | 2 |
| `beyond_contetion` | `beyond_contention` | 2 |
| `bring_relive` | `bring_relief` | 2 |
| `cannot_recieve_evaluation` | `cannot_receive_evaluation` | 2 |
| `certainity_no_need` | `certainty_no_need` | 2 |
| `character_revelant` | `character_relevant` | 2 |
| `commined_to` | `committed_to` | 1 |
| `commited_opposite` | `committed_opposite` | 2 |
| `easy_to_achive` | `easy_to_achieve` | 2 |
| `evidence_not_commited` | `evidence_not_committed` | 2 |
| `individuals_propoerty` | `individuals_property` | 2 |
| `not_has_propery` | `not_has_property` | 2 |
| `oficcial_rulebook` | `official_rulebook` | 2 |
| `part_propert` | `part_property` | 2 |
| `popular_practise` | `popular_practice` | 3 |
| `property_precion_needed` | `property_precision_needed` | 2 |
| `reasonable_belive` | `reasonable_believe` | 2 |
| `recieves_evaluation` | `receives_evaluation` | 2 |
| `situation_necessiy` | `situation_necessity` | 2 |
| `succesful_account` | `successful_account` | 1 |
| `succesful_diagram` | `successful_diagram` | 2 |
| `vaguness_recejted` | `vagueness_rejected` | 2 |
| `true` | `is_true` | 2 (syntax: `true` is the boolean constant to the term parser) |

## Every edit, grouped by scheme

Format: **line** · category · reason, then the line before and after. An inserted line has an empty "before", a deleted line an empty "after". When a line has several edits, each is listed in the order applied, so one row's "after" is the next row's "before".

### language (templates; "used by" names the schemes that use the predicate)

- **10** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses accept_saying with 0 argument(s) and the template has 0 %s (used by: common_folks_group_ad_populum)
  - before: `  accept_saying: "Therefore, you should accept what I say."`
  - after: `  accept_saying/0: "Therefore, you should accept what I say."`
- **17** · template · template has no %s for arity 1; no scheme uses the predicate, so the arity is set to fit the template instead of inventing an argument (used by: no scheme)
  - before: `  action_fit_quality/1: "Description of action actually fit the quality."`
  - after: `  action_fit_quality/0: "Description of action actually fit the quality."`
- **26** · typo · misspelled predicate (used by: composition_generic, division_generic)
  - before: `  all_parts_propert/2: "All the parts %s have the property %s."`
  - after: `  all_parts_property/2: "All the parts %s have the property %s."`
- **27** · typo · template word 'commined' (used by: commitment1)
  - before: `  also_committed/2: "Generally, when an arguer is committed to %s, it can be inferred that he is also commined to %s."`
  - after: `  also_committed/2: "Generally, when an arguer is committed to %s, it can be inferred that he is also committed to %s."`
- **28** · typo · template word 'commiued' (OCR) (used by: commitment2)
  - before: `  also_committed_questioned/3: "If arguer %s has commiued herself to proposition %s at some point in a dialogue, then it may be inferred that she is also committed to proposition %s, should the question of whether it is true become an issue later in the dialogue."`
  - after: `  also_committed_questioned/3: "If arguer %s has committed herself to proposition %s at some point in a dialogue, then it may be inferred that she is also committed to proposition %s, should the question of whether it is true become an issue later in the dialogue."`
- **38** · typo · misspelled predicate (used by: pragmatic_inconsistency)
  - before: `  arguement_not_accepted/2: "Therefore, %s's argument %s should not be accepted."`
  - after: `  argument_not_accepted/2: "Therefore, %s's argument %s should not be accepted."`
- **40** · typo · template word 'wil l' (used by: threat)
  - before: `  assert_bad/2: "I hereby assert that in fact I wil l see to it that %s occurs if you bring about %s."`
  - after: `  assert_bad/2: "I hereby assert that in fact I will see to it that %s occurs if you bring about %s."`
- **42** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses attack_well_supported with 0 argument(s) and the template has 0 %s (used by: ad_hominem)
  - before: `  attack_well_supported: "Allegration made in character attack premise evidence are well supported."`
  - after: `  attack_well_supported/0: "Allegration made in character attack premise evidence are well supported."`
- **42** · typo · template word 'Allegration' (used by: ad_hominem)
  - before: `  attack_well_supported/0: "Allegration made in character attack premise evidence are well supported."`
  - after: `  attack_well_supported/0: "Allegation made in character attack premise evidence are well supported."`
- **43** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses background with 0 argument(s) and the template has 0 %s (used by: common_folk_ad_populum)
  - before: `  background: "The speaker is an ordinary person, that is, share a common background with the audience."`
  - after: `  background/0: "The speaker is an ordinary person, that is, share a common background with the audience."`
- **50** · template · '%v' is not a %s verb (WAG lint error) (used by: expert_opinion)
  - before: `  based_on_evidence/1: "The assertion %v is based on evidence."`
  - after: `  based_on_evidence/1: "The assertion %s is based on evidence."`
- **51** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses being_credible with 0 argument(s) and the template has 0 %s (used by: common_folk_ad_populum)
  - before: `  being_credible: "Therefore, you ought to take what I say as being more credible or acceptable."`
  - after: `  being_credible/0: "Therefore, you ought to take what I say as being more credible or acceptable."`
- **54** · typo · misspelled predicate (used by: sorites_slippery_slope)
  - before: `  beyond_contetion/2: "It is clearly beyond contention that %s , has %s."`
  - after: `  beyond_contention/2: "It is clearly beyond contention that %s , has %s."`
- **55** [delete] · syntax · duplicate key biased/1 (also line 56); YAML keeps only the later one, so this line was never used; deleting it leaves the loaded language unchanged (used by: bias, credible_source, witness_testimony)
  - before: `  biased/1: "%s is biased."`
  - after: _(none)_
- **60** · template · malformed verb '% m' printed '%!m(string=...)ore' (used by: practical_reasoning1, practical_reasoning2)
  - before: `  bring_about_more_effectively/2: "There exists an action that would bring about % more effectively than %s."`
  - after: `  bring_about_more_effectively/2: "There exists an action that would bring about %s more effectively than %s."`
- **61** · typo · misspelled predicate ('relive' for 'relief') (used by: distress)
  - before: `  bring_relive/2: "Therefore, %s ought to bring about %s."`
  - after: `  bring_relief/2: "Therefore, %s ought to bring about %s."`
- **68** [insert after] · template · opposition-normative uses cannot_present_sametime(A,X,X2) with 3 arguments; only /2 had a template (used by: opposition-descriptive_1, opposition-normative)
  - before: _(none)_
  - after: `  cannot_present_sametime/3: "%s cannot present %s and %s at the same time and in the same respect."`
- **69** · typo · misspelled predicate (used by: opposition-normative)
  - before: `  cannot_recieve_evaluation/2: "%s cannot receive the evaluation %s at the same time and in the same respect."`
  - after: `  cannot_receive_evaluation/2: "%s cannot receive the evaluation %s at the same time and in the same respect."`
- **80** · typo · template word 'arguement' (used by: verbal_classification_vagueness)
  - before: `  certainity_no_need/2: "In arguement %s property %s doesn't need to be certain for the sake of the argument." `
  - after: `  certainity_no_need/2: "In argument %s property %s doesn't need to be certain for the sake of the argument." `
- **80** · typo · misspelled predicate (used by: verbal_classification_vagueness)
  - before: `  certainity_no_need/2: "In argument %s property %s doesn't need to be certain for the sake of the argument." `
  - after: `  certainty_no_need/2: "In argument %s property %s doesn't need to be certain for the sake of the argument." `
- **85** · typo · template word 'depent' (used by: circumstantial_ad_hominem)
  - before: `  character_issue/1: "Character is an issue in dialogue, and more specifically, %s 's argument depent on his/her credibility."`
  - after: `  character_issue/1: "Character is an issue in dialogue, and more specifically, %s 's argument depends on his/her credibility."`
- **87** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses character_revelant with 0 argument(s) and the template has 0 %s (used by: ad_hominem)
  - before: `  character_revelant: "In the type of dialogue in which the arguement was used issue of character revelant."`
  - after: `  character_revelant/0: "In the type of dialogue in which the arguement was used issue of character revelant."`
- **87** · typo · template word 'arguement' (used by: ad_hominem)
  - before: `  character_revelant/0: "In the type of dialogue in which the arguement was used issue of character revelant."`
  - after: `  character_revelant/0: "In the type of dialogue in which the argument was used issue of character revelant."`
- **87** · typo · template word 'revelant' (used by: ad_hominem)
  - before: `  character_revelant/0: "In the type of dialogue in which the argument was used issue of character revelant."`
  - after: `  character_revelant/0: "In the type of dialogue in which the argument was used issue of character relevant."`
- **87** · typo · misspelled predicate (used by: ad_hominem)
  - before: `  character_revelant/0: "In the type of dialogue in which the argument was used issue of character relevant."`
  - after: `  character_relevant/0: "In the type of dialogue in which the argument was used issue of character relevant."`
- **88** · template · cited_bad_consequences is used with 1 argument (X); the second %s printed '%!s(MISSING)' (used by: threat)
  - before: `  cited_bad_consequences/1: "If you bring about %s, some cited bad consequences, %s, will follow."`
  - after: `  cited_bad_consequences/1: "If you bring about %s, some cited bad consequences will follow."`
- **91** · typo · misspelled predicate (used by: circumstantial_ad_hominem)
  - before: `  commited_opposite/2: "%s is personally committed to the opposite of %s, as shown by commitments expressed in her/ his personal actions or personal circumstances expressing such commitments."`
  - after: `  committed_opposite/2: "%s is personally committed to the opposite of %s, as shown by commitments expressed in her/ his personal actions or personal circumstances expressing such commitments."`
- **97** · template · the template has 3 %s and both schemes use conflict_goals(P,A2,A) with 3 arguments (used by: practical_necessary_condition, practical_sufficient_condition)
  - before: `  conflict_goals/2: "%s has some other goals %s that conflict with %s." `
  - after: `  conflict_goals/3: "%s has some other goals %s that conflict with %s." `
- **102** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses correctly_report with 0 argument(s) and the template has 0 %s (used by: rule_physical_world_2)
  - before: `  correctly_report: "Oflicial rulelebooks correctly report the rules in force."`
  - after: `  correctly_report/0: "Oflicial rulelebooks correctly report the rules in force."`
- **102** · typo · template words 'Oflicial rulelebooks' (OCR) (used by: rule_physical_world_2)
  - before: `  correctly_report/0: "Oflicial rulelebooks correctly report the rules in force."`
  - after: `  correctly_report/0: "Official rulebooks correctly report the rules in force."`
- **110** [insert after] · template · abductive_backward uses data_set(D,A) with 2 arguments; only data_set/1 had a template (used by: abductive_backward, abductive_forward)
  - before: _(none)_
  - after: `  data_set/2: "%s is a set of data or supposed facts in case %s."`
- **118** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses do_obligator with 0 argument(s) and the template has 0 %s (used by: regulative_rule)
  - before: `  do_obligator: "One must not do what is prohibite -obligator-."`
  - after: `  do_obligator/0: "One must not do what is prohibite -obligator-."`
- **118** · typo · template words 'prohibite -obligator-' (cf. obligatory/1: 'prohibited -obligatory-') (used by: regulative_rule)
  - before: `  do_obligator/0: "One must not do what is prohibite -obligator-."`
  - after: `  do_obligator/0: "One must not do what is prohibited -obligatory-."`
- **121** · typo · template word 'achive' (used by: end_and_means_3)
  - before: `  easy_to_achive/1: "%s is easy to achive."`
  - after: `  easy_to_achive/1: "%s is easy to achieve."`
- **121** · typo · misspelled predicate (used by: end_and_means_3)
  - before: `  easy_to_achive/1: "%s is easy to achieve."`
  - after: `  easy_to_achieve/1: "%s is easy to achieve."`
- **130** · typo · misspelled predicate (used by: inconsistent_commitment)
  - before: `  evidence_not_commited/2: "There are evidence in this particular case shows that %s is not committed to %s." `
  - after: `  evidence_not_committed/2: "There are evidence in this particular case shows that %s is not committed to %s." `
- **132** · template · language key without /arity; regulative_rule uses 'excuse' with 0 arguments (excuse/1 is a different predicate, used by plea_for_excuse) (used by: plea_for_excuse, regulative_rule)
  - before: `  excuse: "%s has an adequate excuse or an overriding duty."`
  - after: `  excuse/0: "%s has an adequate excuse or an overriding duty."`
- **132** · template · excuse is used with 0 arguments, so the template may not contain %s (used by: plea_for_excuse, regulative_rule)
  - before: `  excuse/0: "%s has an adequate excuse or an overriding duty."`
  - after: `  excuse/0: "The agent has an adequate excuse or an overriding duty."`
- **139** · template · favor is used with 1 argument (A); both %s meant A, the second printed '%!s(MISSING)' (used by: popular_opinion)
  - before: `  favor/1: "If %s is generally accepted as truei that gives a reason in favor of %s."`
  - after: `  favor/1: "If %s is generally accepted as truei that gives a reason in favor of it."`
- **139** · typo · template word 'truei' (used by: popular_opinion)
  - before: `  favor/1: "If %s is generally accepted as truei that gives a reason in favor of it."`
  - after: `  favor/1: "If %s is generally accepted as true that gives a reason in favor of it."`
- **147** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses goal_prestige with 0 argument(s) and the template has 0 %s (used by: snob_appeal_ad_populum)
  - before: `  goal_prestige: "Prestige is an important goal for you."`
  - after: `  goal_prestige/0: "Prestige is an important goal for you."`
- **150** [delete] · syntax · duplicate key good/1, identical to line 149 (used by: end_and_means_1, end_and_means_3, positive_consequences)
  - before: `  good/1: "%s is good."`
  - after: _(none)_
- **151** · typo · template word 'represen ts' (OCR) (used by: moral_justification_ad_populum)
  - before: `  good_group/2: "Everybody who is good, or who represen ts a group %s with good qualities, accepts policy %s."`
  - after: `  good_group/2: "Everybody who is good, or who represents a group %s with good qualities, accepts policy %s."`
- **194** · typo · misspelled predicate (used by: verbal_slippery_slope)
  - before: `  individuals_propoerty/3: "If %s, has %s then, %s has it too."`
  - after: `  individuals_property/3: "If %s, has %s then, %s has it too."`
- **195** · typo · template word 'vinue' (OCR) (used by: inconsistent_commitment)
  - before: `  initial_commitment/2: "%s has claimed or indicated that he is committed to proposition %s. Generally, or by vinue of what he has said in the past."`
  - after: `  initial_commitment/2: "%s has claimed or indicated that he is committed to proposition %s. Generally, or by virtue of what he has said in the past."`
- **200** · template · malformed verb '% i' printed '%!i(string=...)s' (used by: group_2, rule_physical_world_1)
  - before: `  is/2: "% is %s."`
  - after: `  is/2: "%s is %s."`
- **208** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses means_end with 1 argument(s) and the template has 1 %s (used by: end_and_means_2)
  - before: `  means_end: "Means %s should be regarded as an end."`
  - after: `  means_end/1: "Means %s should be regarded as an end."`
- **213** · typo · template word 'Appearence' (used by: no scheme)
  - before: `  misleading/1: "Appearence of the objects looking like it could be classified under %s is misleading for some reasons."`
  - after: `  misleading/1: "Appearance of the objects looking like it could be classified under %s is misleading for some reasons."`
- **218** · typo · template word 'effecient' (used by: practical_inference)
  - before: `  more_efficient_alternative_actions/2: " More effecient actions bringing about %s and that would also brin about %s."`
  - after: `  more_efficient_alternative_actions/2: " More efficient actions bringing about %s and that would also brin about %s."`
- **218** · typo · template word 'brin' (used by: practical_inference)
  - before: `  more_efficient_alternative_actions/2: " More efficient actions bringing about %s and that would also brin about %s."`
  - after: `  more_efficient_alternative_actions/2: " More efficient actions bringing about %s and that would also bring about %s."`
- **219** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses more_important_ends with 0 argument(s) and the template has 0 %s (used by: end_and_means_2)
  - before: `  more_important_ends: "Ends are more important than the means used to achieve them."`
  - after: `  more_important_ends/0: "Ends are more important than the means used to achieve them."`
- **221** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses more_reliable with 0 argument(s) and the template has 0 %s (used by: sign)
  - before: `  more_reliable: "There are more reliable account for the sign."`
  - after: `  more_reliable/0: "There are more reliable account for the sign."`
- **231** · typo · template word 'thte' (used by: negative_normal_expectations)
  - before: `  normal_true/1: "If thte situation were normal, %s would he true."`
  - after: `  normal_true/1: "If the situation were normal, %s would he true."`
- **231** · typo · template word 'he' for 'be' (OCR) (used by: negative_normal_expectations)
  - before: `  normal_true/1: "If the situation were normal, %s would he true."`
  - after: `  normal_true/1: "If the situation were normal, %s would be true."`
- **239** · typo · misspelled predicate (used by: verbal_slippery_slope)
  - before: `  not_has_propery/2: "Therefore, you can't truly say that %s, has %s"`
  - after: `  not_has_property/2: "Therefore, you can't truly say that %s, has %s"`
- **242** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses not_necessarily_true with 0 argument(s) and the template has 0 %s (used by: rule_physical_world_1)
  - before: `  not_necessarily_true: "First premise not necessary true." `
  - after: `  not_necessarily_true/0: "First premise not necessary true." `
- **243** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses not_normal with 0 argument(s) and the template has 0 %s (used by: negative_normal_expectations)
  - before: `  not_normal: "Therefore, the situation is not normal."`
  - after: `  not_normal/0: "Therefore, the situation is not normal."`
- **245** · template · malformed verb '%i' printed '%!i(string=...)s' (used by: alternative_normative)
  - before: `  not_required/1: "%is not required."`
  - after: `  not_required/1: "%s is not required."`
- **251** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses oficcial_rulebook with 3 argument(s) and the template has 3 %s (used by: rule_physical_world_2)
  - before: `  oficcial_rulebook: "In the official rulebook for %s -a rule governed institution, such as a game- it is stated that %s counts as %s."`
  - after: `  oficcial_rulebook/3: "In the official rulebook for %s -a rule governed institution, such as a game- it is stated that %s counts as %s."`
- **251** · typo · misspelled predicate (used by: rule_physical_world_2)
  - before: `  oficcial_rulebook/3: "In the official rulebook for %s -a rule governed institution, such as a game- it is stated that %s counts as %s."`
  - after: `  official_rulebook/3: "In the official rulebook for %s -a rule governed institution, such as a game- it is stated that %s counts as %s."`
- **255** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses other_consequences with 1 argument(s) and the template has 1 %s (used by: negative_consequences_prudential)
  - before: `  other_consequences: "There are some good or bad consequences of doing %s should be taken into account."`
  - after: `  other_consequences/1: "There are some good or bad consequences of doing %s should be taken into account."`
- **256** · template · the predicate has 2 arguments (A,C); 'A %s' was the variable name plus a stray %s ('A is true') (used by: negative_practical)
  - before: `  other_consequences_presumption/2: "If I act on the presum ption that %s is not true, and A %s true, consequences %s will follow."`
  - after: `  other_consequences_presumption/2: "If I act on the presum ption that %s is not true, and it is true, consequences %s will follow."`
- **256** · typo · template word 'presum ption' (OCR) (used by: negative_practical)
  - before: `  other_consequences_presumption/2: "If I act on the presum ption that %s is not true, and it is true, consequences %s will follow."`
  - after: `  other_consequences_presumption/2: "If I act on the presumption that %s is not true, and it is true, consequences %s will follow."`
- **260** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses ought_to_continue with 2 argument(s) and the template has 2 %s (used by: waste)
  - before: `  ought_to_continue: "Therefore, %s ought to continue trying to realize %s."`
  - after: `  ought_to_continue/2: "Therefore, %s ought to continue trying to realize %s."`
- **262** · typo · misspelled predicate (used by: part_in_whole)
  - before: `  part_propert/2: "%s is %s." `
  - after: `  part_property/2: "%s is %s." `
- **277** · template · popular_practise is used with 1 argument (A); both %s meant A, the second printed '%!s(MISSING)' (used by: popular_practice)
  - before: `  popular_practise/1: "%s is a popular practice among those who are familiar with what is acceptable or not in regard to %s."`
  - after: `  popular_practise/1: "%s is a popular practice among those who are familiar with what is acceptable or not in regard to it."`
- **277** · typo · misspelled predicate and scheme id ('practise' for the noun 'practice') (used by: popular_practice)
  - before: `  popular_practise/1: "%s is a popular practice among those who are familiar with what is acceptable or not in regard to it."`
  - after: `  popular_practice/1: "%s is a popular practice among those who are familiar with what is acceptable or not in regard to it."`
- **278** [delete] · typo · 'posible' is a misspelling of 'possible'; renamed in its only use (line 1105), which then uses the existing template possible/1 (line 284); a second possible/1 key would be a duplicate (used by: practical_reasoning1, practical_reasoning2, waste)
  - before: `  posible/1: "%s is possible."`
  - after: _(none)_
- **281** · template · the template has 2 %s and the scheme uses position_to_know_group(G,A) (used by: position_to_know_ad_populum)
  - before: `  position_to_know_group/1: "Group %s is in a special position to know that %s is true."`
  - after: `  position_to_know_group/2: "Group %s is in a special position to know that %s is true."`
- **297** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses proper_sequence with 0 argument(s) and the template has 0 %s (used by: gradualism)
  - before: `  proper_sequence: "Sequence of proposition is dealt with in the appropriate order."`
  - after: `  proper_sequence/0: "Sequence of proposition is dealt with in the appropriate order."`
- **299** · typo · misspelled predicate (used by: verbal_classification_vagueness)
  - before: `  property_precion_needed/1: "%s occurs in a context of dialogue that requires a certain level of precision that is appropriate for that context."`
  - after: `  property_precision_needed/1: "%s occurs in a context of dialogue that requires a certain level of precision that is appropriate for that context."`
- **302** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses prudential with 1 argument(s) and the template has 1 %s (used by: negative_consequences_prudential)
  - before: `  prudential: "Therefore, you ought to reconsider other things being equal and you ought prudentially to do %s."`
  - after: `  prudential/1: "Therefore, you ought to reconsider other things being equal and you ought prudentially to do %s."`
- **306** · repair — needs approval · template for the repaired reapplication_sequence(A2,A) (line 1557) (used by: sorites_slippery_slope)
  - before: `  reapplication_sequence: "A sequence of modus ponens subarguments, a premise set linking premises and conclusions from the clear area through the grey area."`
  - after: `  reapplication_sequence/2: "A sequence of modus ponens subarguments, a premise set linking premises and conclusions from the clear area through the grey area."`
- **306** · repair — needs approval · template for the repaired reapplication_sequence(A2,A) (line 1557), in the file's '-...-' style (used by: sorites_slippery_slope)
  - before: `  reapplication_sequence/2: "A sequence of modus ponens subarguments, a premise set linking premises and conclusions from the clear area through the grey area."`
  - after: `  reapplication_sequence/2: "A sequence of modus ponens subarguments, a premise set linking premises and conclusions from the clear area -%s- through the grey area to %s."`
- **309** · typo · misspelled predicate (used by: memory)
  - before: `  reasonable_belive/1: "It is reasonable to believe %s."`
  - after: `  reasonable_believe/1: "It is reasonable to believe %s."`
- **312** · typo · template word 'recieves' (used by: opposition-normative)
  - before: `  recieves_evaluation/2: "%s recieves the evaluation %s." `
  - after: `  recieves_evaluation/2: "%s receives the evaluation %s." `
- **312** · typo · misspelled predicate (used by: opposition-normative)
  - before: `  recieves_evaluation/2: "%s receives the evaluation %s." `
  - after: `  receives_evaluation/2: "%s receives the evaluation %s." `
- **313** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses reference_class with 0 argument(s) and the template has 0 %s (used by: retroductive)
  - before: `  reference_class: "Reference class of other agents who also might have brought about this event and also have same character is sufficient."`
  - after: `  reference_class/0: "Reference class of other agents who also might have brought about this event and also have same character is sufficient."`
- **323** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses rule_changed with 0 argument(s) and the template has 0 %s (used by: rule_physical_world_2)
  - before: `  rule_changed: "The rule has changed since the publication of the rulebook." `
  - after: `  rule_changed/0: "The rule has changed since the publication of the rulebook." `
- **325** · typo · template word 'o f' (OCR) (used by: plea_for_excuse)
  - before: `  rule_require/3: "Normally, rule %s requires or forbids a type o f action or inaction %s, which carries with it a sanction - penalty- %s."`
  - after: `  rule_require/3: "Normally, rule %s requires or forbids a type of action or inaction %s, which carries with it a sanction - penalty- %s."`
- **328** · template · sacrifice_value is used with 2 arguments (A,V); the value V was printed as '%!(EXTRA ...)' (used by: sacrifice)
  - before: `  sacrifice_value/2: "%s has a great value."`
  - after: `  sacrifice_value/2: "%s has a great value %s."`
- **329** · template · the template has 3 %s and the scheme uses same_meaning(L,W,D) (used by: rule_mental_world)
  - before: `  same_meaning/2: "Speakers of %s believe that %s means the same as %s."`
  - after: `  same_meaning/3: "Speakers of %s believe that %s means the same as %s."`
- **337** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses should_not with 1 argument(s) and the template has 1 %s (used by: danger)
  - before: `  should_not: "Therefore you should not bring about %s."`
  - after: `  should_not/1: "Therefore you should not bring about %s."`
- **344** · typo · template word 'caryying' (used by: two_person_practical)
  - before: `  situation_necessiy/4: "As %s sets the situation, %s is a necessary or sufficient condition for caryying out %s, and tells %s this."`
  - after: `  situation_necessiy/4: "As %s sets the situation, %s is a necessary or sufficient condition for carrying out %s, and tells %s this."`
- **344** · typo · misspelled predicate (used by: two_person_practical)
  - before: `  situation_necessiy/4: "As %s sets the situation, %s is a necessary or sufficient condition for carrying out %s, and tells %s this."`
  - after: `  situation_necessity/4: "As %s sets the situation, %s is a necessary or sufficient condition for carrying out %s, and tells %s this."`
- **346** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses speakers_correct with 0 argument(s) and the template has 0 %s (used by: rule_mental_world)
  - before: `  speakers_correct: "Speakers of a language are generally correct about word definitions in that language."`
  - after: `  speakers_correct/0: "Speakers of a language are generally correct about word definitions in that language."`
- **350** · template · special_proposition is used with 2 arguments (A,K); the first was written as the letter A (used by: epistemic_ignorance)
  - before: `  special_proposition/2: "A is a special type of proposition such that if it were true, it would normally or usually be expected to be in %s."`
  - after: `  special_proposition/2: "%s is a special type of proposition such that if it were true, it would normally or usually be expected to be in %s."`
- **354** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses strong_advocate with 3 argument(s) and the template has 3 %s (used by: pragmatic_inconsistency)
  - before: `  strong_advocate: "%s advocate %s in a stron way indicating her personal commitment to %s."`
  - after: `  strong_advocate/3: "%s advocate %s in a stron way indicating her personal commitment to %s."`
- **354** · typo · template word 'stron' (used by: pragmatic_inconsistency)
  - before: `  strong_advocate/3: "%s advocate %s in a stron way indicating her personal commitment to %s."`
  - after: `  strong_advocate/3: "%s advocate %s in a strong way indicating her personal commitment to %s."`
- **357** · template · language key without /arity; negative_consequences_prudential uses strongly_occur(X) with 1 argument (used by: negative_consequences_prudential)
  - before: `  strongly_occur: "If you don't do %s consequences %s may strongly appear."`
  - after: `  strongly_occur/1: "If you don't do %s consequences %s may strongly appear."`
- **357** · template · strongly_occur is used with 1 argument (X); the template had a second %s (used by: negative_consequences_prudential)
  - before: `  strongly_occur/1: "If you don't do %s consequences %s may strongly appear."`
  - after: `  strongly_occur/1: "If you don't do %s consequences may strongly appear."`
- **359** · typo · misspelled predicate (used by: no scheme)
  - before: `  succesful_account/2: "%s , is the account that explains %s most successfully."`
  - after: `  successful_account/2: "%s , is the account that explains %s most successfully."`
- **360** · typo · misspelled predicate (used by: abductive_forward)
  - before: `  succesful_diagram/1: "The most plausible -strongest- argument is represented by %s."`
  - after: `  successful_diagram/1: "The most plausible -strongest- argument is represented by %s."`
- **361** · typo · template word 'availablle' (used by: abductive_backward)
  - before: `  sufficient_account/2: "%s itself as an explanation of %s apart from other alternative explanations availablle so far, is sufficient."`
  - after: `  sufficient_account/2: "%s itself as an explanation of %s apart from other alternative explanations available so far, is sufficient."`
- **364** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses sufficient_evidence with 0 argument(s) and the template has 0 %s (used by: negative_practical)
  - before: `  sufficient_evidence: "search for the evidence progress is sufficient."`
  - after: `  sufficient_evidence/0: "search for the evidence progress is sufficient."`
- **366** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses synonymous with 0 argument(s) and the template has 0 %s (used by: rule_physical_world_1)
  - before: `  synonymous: "Whenever expressions are referentially coextensive, they are synonymous."`
  - after: `  synonymous/0: "Whenever expressions are referentially coextensive, they are synonymous."`
- **372** · syntax · 'true' is the boolean constant to the Carneades term parser, so 'true/1' is not a predicate symbol (validator); predicate renamed is_true, same template (used by: pop_scheme)
  - before: `  true/1: "%s is true." `
  - after: `  is_true/1: "%s is true." `
- **378** · typo · template word 'generalzation' (used by: from_example)
  - before: `  typical_example/1: "Example %s is an typical example of the kinds of cases the generalzation covers."`
  - after: `  typical_example/1: "Example %s is an typical example of the kinds of cases the generalization covers."`
- **385** · typo · misspelled predicate (used by: verbal_classification_vagueness)
  - before: `  vaguness_recejted/1: " Therefore, %s ought to be rejected as deficient."`
  - after: `  vagueness_rejected/1: " Therefore, %s ought to be rejected as deficient."`
- **390** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses waste_bad_thing with 2 argument(s) and the template has 2 %s (used by: waste)
  - before: `  waste_bad_thing: "If all %s 's previous attempts to realize %s are wasted, that would be a bad thing."`
  - after: `  waste_bad_thing/2: "If all %s 's previous attempts to realize %s are wasted, that would be a bad thing."`

### popular_opinion

- **410** · typo · source: 'lnformal' (OCR l for I)
  - before: `              Douglas Walton, lnformal Logic. New York: Cambridge University Press, 1989, p. 89.`
  - after: `              Douglas Walton, Informal Logic. New York: Cambridge University Press, 1989, p. 89.`
- **418** · syntax · misspelled key: yaml.v2 silently ignores 'assumption:', so generally_accepted(A) was dropped
  - before: `    assumption:`
  - after: `    assumptions:`

### pop_scheme

- **431** · syntax · predicate 'true' renamed is_true (see line 372); the statement is unchanged
  - before: `      - true(A)`
  - after: `      - is_true(A)`

### moral_justification_ad_populum_excuse

- **495** [insert after] · repair — needs approval · empty conclusions: Carneades built no argument; the 'Therefore, ...' formula listed as the first premise is the conclusion (as in the category file B/4)
  - before: _(none)_
  - after: `      - morally_acceptable_action(X,A)`
- **497** [delete] · repair — needs approval · moved to conclusions (line 495); it was the scheme's conclusion listed as a premise
  - before: `      -  morally_acceptable_action(X,A)`
  - after: _(none)_

### popular_practise → popular_practice

- **569** · typo · misspelled predicate and scheme id ('practise' for the noun 'practice')
  - before: `  - id: popular_practise`
  - after: `  - id: popular_practice`
- **571** · typo · title: 'Practise' (the noun is 'practice')
  - before: `      title: Argument From Popular Practise`
  - after: `      title: Argument From Popular Practice`
- **579** · typo · misspelled predicate and scheme id ('practise' for the noun 'practice')
  - before: `      - popular_practise(A)`
  - after: `      - popular_practice(A)`

### practical_analogy_positive

- **645** · typo · source: 'lnformal' (OCR l for I)
  - before: `              Douglas Walton, lnformal Logic New York: Cambridge University Press, 1989. p.257.`
  - after: `              Douglas Walton, Informal Logic New York: Cambridge University Press, 1989. p.257.`

### practical_analogy_negative

- **657** · typo · source: 'lnformal' (OCR l for I)
  - before: `              Douglas Walton, lnformal Logic New York: Cambridge University Press, 1989. p.257.`
  - after: `              Douglas Walton, Informal Logic New York: Cambridge University Press, 1989. p.257.`

### composition_generic

- **669** · typo · source: 'Iogic' (OCR I for l)
  - before: `              Douglas Walton, Informal Iogic, New York: Cambridge University Press, 1989, p.130.`
  - after: `              Douglas Walton, Informal Logic, New York: Cambridge University Press, 1989, p.130.`
- **674** · typo · misspelled predicate
  - before: `      - all_parts_propert(A,P)`
  - after: `      - all_parts_property(A,P)`

### part_in_whole

- **684** · variable · A is used in 3 formulas but undeclared (non-ground statements); P is declared but used nowhere (validator: 'declared but not used'); the category file A_reasoning/2 declares [Y,X,A]
  - before: `    variables: [Y,X,P]`
  - after: `    variables: [Y,X,A]`
- **686** · typo · misspelled predicate
  - before: `      - part_propert(Y,A)`
  - after: `      - part_property(Y,A)`
- **691** · syntax · no space after '-': YAML reads the exceptions value as a string, not a list
  - before: `      -smaller_part_property(Y,A,X)`
  - after: `      - smaller_part_property(Y,A,X)`

### division_generic

- **702** · typo · misspelled predicate
  - before: `      - all_parts_propert(A,P)`
  - after: `      - all_parts_property(A,P)`

### opposition-descriptive_1

- **712** · variable · primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme
  - before: `    variables: [P,P',A]`
  - after: `    variables: [P,P2,A]`
- **714** · variable · primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme
  - before: `      - cannot_present_sametime(A,P')`
  - after: `      - cannot_present_sametime(A,P2)`
- **716** · variable · primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme
  - before: `      - presents_predicate(A,P,P')`
  - after: `      - presents_predicate(A,P,P2)`

### opposition-descriptive_2

- **725** · variable · primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme
  - before: `    variables: [P,P',A]`
  - after: `    variables: [P,P2,A]`
- **727** · variable · primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme
  - before: `      - must_present_sametime(A,P')`
  - after: `      - must_present_sametime(A,P2)`
- **729** · variable · primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme
  - before: `      - not_presents_predicate(A,P,P')`
  - after: `      - not_presents_predicate(A,P,P2)`

### opposition-descriptive_3

- **732** · syntax · no space after 'id:': YAML reads a plain string and the next line fails (the file's load error 'line 733: mapping values are not allowed')
  - before: `  - id:opposition-descriptive_3`
  - after: `  - id: opposition-descriptive_3`
- **734** · typo · title of opposition-descriptive_3 repeated 'Scheme 2' (same as opposition-descriptive_2)
  - before: `      title: Argument From Opposition - Descriptive Scheme 2`
  - after: `      title: Argument From Opposition - Descriptive Scheme 3`
- **738** · variable · primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme
  - before: `    variables: [P,P',A]`
  - after: `    variables: [P,P2,A]`
- **740** · variable · primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme
  - before: `      - must_present_sametime(A,P')`
  - after: `      - must_present_sametime(A,P2)`
- **742** · variable · primed variable P' does not parse (' starts a quoted atom); renamed P2 within the scheme
  - before: `      - must_present_complementary_contrary(A,P,P')`
  - after: `      - must_present_complementary_contrary(A,P,P2)`

### opposition-normative

- **751** · variable · primed variable X' does not parse (' starts a quoted atom); renamed X2 within the scheme
  - before: `    variables: [A,X,X']`
  - after: `    variables: [A,X,X2]`
- **753** · variable · primed variable X' does not parse (' starts a quoted atom); renamed X2 within the scheme
  - before: `      - cannot_recieve_evaluation(A,X')`
  - after: `      - cannot_recieve_evaluation(A,X2)`
- **753** · typo · misspelled predicate
  - before: `      - cannot_recieve_evaluation(A,X2)`
  - after: `      - cannot_receive_evaluation(A,X2)`
- **755** · variable · primed variable X' does not parse (' starts a quoted atom); renamed X2 within the scheme
  - before: `      - cannot_present_sametime(A,X,X')`
  - after: `      - cannot_present_sametime(A,X,X2)`
- **756** · variable · primed variable X' does not parse (' starts a quoted atom); renamed X2 within the scheme
  - before: `      - recieves_evaluation(A,X')`
  - after: `      - recieves_evaluation(A,X2)`
- **756** · typo · misspelled predicate
  - before: `      - recieves_evaluation(A,X2)`
  - after: `      - receives_evaluation(A,X2)`

### rhetorical_oppositions_normative

- **764** · variable · primed variable X' does not parse (' starts a quoted atom); renamed X2 within the scheme
  - before: `    variables: [X,X',A,A']`
  - after: `    variables: [X,X2,A,A']`
- **764** · variable · primed variable A' does not parse (' starts a quoted atom); renamed A2 within the scheme
  - before: `    variables: [X,X2,A,A']`
  - after: `    variables: [X,X2,A,A2]`
- **766** · variable · primed variable X' does not parse (' starts a quoted atom); renamed X2 within the scheme
  - before: `      - contrary_treatment_required(X',A')`
  - after: `      - contrary_treatment_required(X2,A')`
- **766** · variable · primed variable A' does not parse (' starts a quoted atom); renamed A2 within the scheme
  - before: `      - contrary_treatment_required(X2,A')`
  - after: `      - contrary_treatment_required(X2,A2)`
- **768** · variable · primed variable A' does not parse (' starts a quoted atom); renamed A2 within the scheme
  - before: `      - if_treatment_required(X,A,A',X)`
  - after: `      - if_treatment_required(X,A,A2,X)`
- **768** · repair — needs approval · the 4th argument repeats X where the template says 'the contrary treatment %s', i.e. X2; otherwise X2 occurs only in the conclusion (validator error)
  - before: `      - if_treatment_required(X,A,A2,X)`
  - after: `      - if_treatment_required(X,A,A2,X2)`

### rhetorical_oppositions_descriptive

- **777** · variable · primed variable R' does not parse (' starts a quoted atom); renamed R2 within the scheme
  - before: `    variables: [A,B,R,R']`
  - after: `    variables: [A,B,R,R2]`
- **779** · variable · primed variable R' does not parse (' starts a quoted atom); renamed R2 within the scheme
  - before: `      - converse_relation(B,R',A)`
  - after: `      - converse_relation(B,R2,A)`
- **781** · variable · primed variable R' does not parse (' starts a quoted atom); renamed R2 within the scheme
  - before: `      - relation_converse_relation(A,R,B,R')`
  - after: `      - relation_converse_relation(A,R,B,R2)`

### alternative_cognitive

- **788** · typo · source: 'rmd' (OCR for 'und')
  - before: `              Manfred Kienpointner, Alltagslogik, Struktur rmd Funktion von Argumentationmustern. `
  - after: `              Manfred Kienpointner, Alltagslogik, Struktur und Funktion von Argumentationmustern. `

### alternative_normative

- **802** · typo · source: 'rmd' (OCR for 'und')
  - before: `              Manfred Kienpointner, Alltagslogik, Struktur rmd Funktion von Argumentationmustern. `
  - after: `              Manfred Kienpointner, Alltagslogik, Struktur und Funktion von Argumentationmustern. `

### verbal_classification_vagueness

- **815** · typo · source: 'Schmes'
  - before: `                Douglas Walton, Argumentation Schmes for Presumptive Reasoning. `
  - after: `                Douglas Walton, Argumentation Schemes for Presumptive Reasoning. `
- **817** · variable · F is used in all 4 property formulas but undeclared (non-ground statements); P is declared but used nowhere (validator: 'declared but not used'); P is the misnamed F
  - before: `    variables: [A,P]`
  - after: `    variables: [A,F]`
- **819** · typo · misspelled predicate
  - before: `      -  vaguness_recejted(A)`
  - after: `      -  vagueness_rejected(A)`
- **822** · typo · misspelled predicate
  - before: `      - property_precion_needed(F)`
  - after: `      - property_precision_needed(F)`
- **825** · typo · misspelled predicate
  - before: `      - certainity_no_need(A,F)`
  - after: `      - certainty_no_need(A,F)`

### verbal_classification_arbitrariness

- **831** · typo · source: 'Schmes'
  - before: `               Douglas Walton, Argumentation Schmes for Presumptive Reasoning. `
  - after: `               Douglas Walton, Argumentation Schemes for Presumptive Reasoning. `
- **838** · syntax · unbalanced '((': the term does not parse and InstantiateScheme silently dropped this premise
  - before: `      - needed_nonarbitrary((A,F)`
  - after: `      - needed_nonarbitrary(A,F)`

### positive_value

- **869** · syntax · unquoted title containing ': ' is a YAML error (trailing space dropped)
  - before: `      title: Argument From Values - Variant 1: Positive Value `
  - after: `      title: "Argument From Values - Variant 1: Positive Value"`
- **874** · typo · source: 'lnformal' (OCR l for I)
  - before: `               between Parties without a Consensus about Values." lnformal Logic 22 (2003a), pp. 231 -245. `
  - after: `               between Parties without a Consensus about Values." Informal Logic 22 (2003a), pp. 231 -245. `

### negative_value

- **884** · syntax · unquoted title containing ': ' is a YAML error
  - before: `      title: Argument From Values - Variant 2: Negative Value`
  - after: `      title: "Argument From Values - Variant 2: Negative Value"`
- **889** · typo · source: 'lnformal' (OCR l for I)
  - before: `               between Parties without a Consensus about Values." lnformal Logic 22 (2003a), pp. 231 -245. `
  - after: `               between Parties without a Consensus about Values." Informal Logic 22 (2003a), pp. 231 -245. `

### sacrifice

- **901** · typo · source: 'Trtalise' (OCR)
  - before: `               Ch. Perelman and L. Olbrechts-Tyteca, The New Rhetoric: A Trtalise on Argumentation. `
  - after: `               Ch. Perelman and L. Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation. `
- **908** · repair — needs approval · undeclared X made the premise non-ground, and the value V occurred only in the conclusion; the template's third argument is 'the value %s', i.e. V
  - before: `      - great_sacrifice(S,A,X)`
  - after: `      - great_sacrifice(S,A,V)`

### practical_goal

- **1024** · typo · source: 'lntelligence' (OCR l for I)
  - before: `              Artificial lntelligence and Law 11 (2003), p. 169.`
  - after: `              Artificial Intelligence and Law 11 (2003), p. 169.`

### end_and_means_1

- **1037** · typo · source: 'Trtatise' (OCR)
  - before: `              Chaim Perelman and Lucie Olbrechts-Tyteca, The New Rhetoric: A Trtatise on Argumentation `
  - after: `              Chaim Perelman and Lucie Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation `

### end_and_means_2

- **1050** · typo · source: 'Trtatise' (OCR)
  - before: `              Chaim Perelman and Lucie Olbrechts-Tyteca, The New Rhetoric: A Trtatise on Argumentation `
  - after: `              Chaim Perelman and Lucie Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation `

### end_and_means_3

- **1063** · typo · source: 'Trtatise' (OCR)
  - before: `              Chaim Perelman and Lucie Olbrechts-Tyteca, The New Rhetoric: A Trtatise on Argumentation `
  - after: `              Chaim Perelman and Lucie Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation `
- **1070** · typo · misspelled predicate
  - before: `      - easy_to_achive(X)`
  - after: `      - easy_to_achieve(X)`

### two_person_practical

- **1083** · typo · misspelled predicate
  - before: `      - situation_necessiy(B,Y,X,A)`
  - after: `      - situation_necessity(B,Y,X,A)`

### waste

- **1105** · typo · misspelled predicate 'posible'
  - before: `     - posible(X)`
  - after: `     - possible(X)`

### negative_normal_expectations

- **1111** · typo · source: 'lgnorance' (OCR l for I)
  - before: `             Douglas Walton, Arguments from lgnorance. University Park: `
  - after: `             Douglas Walton, Arguments from Ignorance. University Park: `

### negative_practical

- **1127** · typo · source: 'lgnorance' (OCR l for I)
  - before: `             Douglas Walton, Arguments from lgnorance. University Park: `
  - after: `             Douglas Walton, Arguments from Ignorance. University Park: `

### epistemic_ignorence → epistemic_ignorance

- **1143** · typo · scheme id 'ignorence'
  - before: `  - id: epistemic_ignorence`
  - after: `  - id: epistemic_ignorance`
- **1145** · typo · title 'Ignorence'
  - before: `      title: Epistemic Argument From Ignorence`
  - after: `      title: Epistemic Argument From Ignorance`
- **1148** · typo · source: 'Quaerterly'
  - before: `              American Philosophical Quaerterly 29 (1992), p. 386.`
  - after: `              American Philosophical Quarterly 29 (1992), p. 386.`

### abductive_forward

- **1203** · typo · misspelled predicate
  - before: `      - succesful_diagram(C1)`
  - after: `      - successful_diagram(C1)`

### treat → threat

- **1310** · typo · scheme id 'treat' for 'threat' (title: Argument From Threat)
  - before: `  - id: treat`
  - after: `  - id: threat`
- **1312** · typo · title 'Treat' for 'Threat'
  - before: `      title: Argument From Treat `
  - after: `      title: Argument From Threat `

### ad_baculum

- **1326** · typo · title 'Treat' for 'Threat'
  - before: `      title: Argument From Disjunctive Ad Baculum Treat`
  - after: `      title: Argument From Disjunctive Ad Baculum Threat`

### distress

- **1388** · typo · misspelled predicate ('relive' for 'relief')
  - before: `      - bring_relive(B,X)`
  - after: `      - bring_relief(B,X)`

### commitment1

- **1408** · typo · misspelled predicate in commitment1's conclusion; committed_to/2 (line 96) is its template, and the category file B/2 has committed_to(A,Y)
  - before: `      - commined_to(A,Y)`
  - after: `      - committed_to(A,Y)`

### ad_hominem

- **1451** · typo · misspelled predicate
  - before: `     - character_revelant`
  - after: `     - character_relevant`

### pragmatic_inconsistency

- **1461** · typo · misspelled predicate
  - before: `      - arguement_not_accepted(A,S)`
  - after: `      - argument_not_accepted(A,S)`

### inconsistent_commitment

- **1482** · typo · misspelled predicate
  - before: `      - evidence_not_commited(A,S)`
  - after: `      - evidence_not_committed(A,S)`

### circumstantial_ad_hominem

- **1497** · typo · misspelled predicate
  - before: `      - commited_opposite(A,S)`
  - after: `      - committed_opposite(A,S)`

### sorites_slippery_slope

- **1550** · typo · source: 'lnformal' (OCR l for I)
  - before: `             Douglas Walton, "The Argument of the Beard." lnformal Logic 18 (1996), p. 151.`
  - after: `             Douglas Walton, "The Argument of the Beard." Informal Logic 18 (1996), p. 151.`
- **1555** · typo · misspelled predicate
  - before: `      - beyond_contetion(A2,P)`
  - after: `      - beyond_contention(A2,P)`
- **1557** · repair — needs approval · 0-ary chaining premise: the conclusion's individual A occurred in no premise (validator error); the chain links the clear case A2 to A
  - before: `      - reapplication_sequence`
  - after: `      - reapplication_sequence(A2,A)`

### verbal_slippery_slope

- **1565** · variable · F is used in individuals_propoerty(A1,F,A2) but undeclared (non-ground statement); declared
  - before: `    variables: [A1,A2,A,P]`
  - after: `    variables: [A1,A2,A,P,F]`
- **1567** · typo · misspelled predicate
  - before: `      - not_has_propery(A1,P)`
  - after: `      - not_has_property(A1,P)`
- **1570** · typo · misspelled predicate
  - before: `      - individuals_propoerty(A1,F,A2)`
  - after: `      - individuals_property(A1,F,A2)`

### rule_physical_world_1

- **1576** · typo · title 'Worl'
  - before: `      title: Argument For Constitutive Rule Claims - Physical Worl Premise Version 1 `
  - after: `      title: Argument For Constitutive Rule Claims - Physical World Premise Version 1 `
- **1587** · syntax · no space after '-': YAML reads a string where a list is expected
  - before: `      -not_necessarily_true`
  - after: `      - not_necessarily_true`

### rule_physical_world_2

- **1591** · typo · title 'Worl'
  - before: `      title: Argument For Constitutive Rule Claims - Physical Worl Premise Version 2`
  - after: `      title: Argument For Constitutive Rule Claims - Physical World Premise Version 2`
- **1599** · typo · misspelled predicate
  - before: `      - oficcial_rulebook(I,D,W)`
  - after: `      - official_rulebook(I,D,W)`
- **1602** · syntax · no space after '-': YAML reads a string where a list is expected
  - before: `      -rule_changed`
  - after: `      - rule_changed`

### rule_mental_world

- **1608** · typo · source: 'Wayrw' (OCR; cf. line 1635 'Wayne Grennan')
  - before: `             Wayrw Grennan, Informal Logic, Montreal: McGill-Queen's University Press, 1997, p. 196.`
  - after: `             Wayne Grennan, Informal Logic, Montreal: McGill-Queen's University Press, 1997, p. 196.`

### exeptional → exceptional

- **1645** · typo · scheme id 'exeptional'
  - before: `  - id: exeptional`
  - after: `  - id: exceptional`

### memory

- **1699** · typo · misspelled predicate
  - before: `      - reasonable_belive(M)`
  - after: `      - reasonable_believe(M)`

### exeption_to_rule → exception_to_rule

- **1976** · typo · scheme id 'exeption'
  - before: `  - id: exeption_to_rule`
  - after: `  - id: exception_to_rule`

## Schemes that cannot pass without a change of meaning

All four instantiate correctly (every premise, assumption and exception, a conclusion, all ground), and their templates pass. They fail only the validator rule that each variable must occur in a premise. The rule exists because Carneades applies schemes by matching premises. `InstantiateScheme` with explicit parameters, which WAG uses, does not need it.

| scheme | variable(s) in no premise | where they occur | proposal (in `tools/proposals.py`, not applied) |
|---|---|---|---|
| `ad_hominem` | S (the argument) | conclusion `bad_not_accepted(A,S)`, exception `stronger_than_character(S)` | add premise `advocates(A,S)` ("%s advocates argument %s."), which the conclusion "A's argument S" presupposes. Walton's scheme has the argument only in the conclusion, so this adds a premise |
| `practical_necessary_condition` | P (agent), A2 (other goal) | assumption `possible_agent(X,P)`, exceptions `side_effect_action(P,X)`, `conflict_goals(P,A2,A)` | the premises are first-person ("My goal…", "I have selected…"). Proposal: `goal_is_action(P,A)` ("The goal of %s is to bring about %s.") and `conflict_goals(P,A)` ("%s has some other goals that conflict with %s."), dropping A2 |
| `practical_sufficient_condition` | P, A2 | same | same |
| `regulative_rule` | A (agent) | conclusion `must_do(A,X)` | Grennan's scheme names the agent only in the conclusion. Proposal: apply the norm premise to A, `do_obligator(A)` ("%s must not do what is prohibited -obligatory-."). This turns the general norm into a claim about A |

With these 14 edits (`walton.proposals.yml`), the validator reports 0 problems and the gate passes for all 111 schemes (`proposals-check.txt`).

## Semantic concerns (left unchanged: these load and instantiate)

- `pop_scheme`: two conclusions, `is_true(A)` and `should_accept(A)`. Carneades builds two arguments that share the premises and the undercutter. The book's conclusion is a disjunction ("A is true, or you should accept A"). The owner should decide whether to split the scheme or keep one conclusion.
- `common_folk_ad_populum`: premise and conclusion look swapped. The premise `being_credible` reads "Therefore, you ought to take what I say as … more credible" and the conclusion `background` reads "The speaker is an ordinary person…". The scheme has no variables, so every instance is the same argument.
- `common_folks_group_ad_populum`: the conclusion `accept_saying` is 0-ary, so every instance concludes the same statement.
- `from_example` (`both_properties`), `fear` (`prevent(Y)`) and `bias_ad_hominem` (`morally_bad(A)`): a premise has a "Therefore, …" template, so premise and conclusion may be swapped.
- `division_generic` is a verbatim copy of `composition_generic` (parts → whole). Division should go from the whole to its parts; the category file A/2 has a `whole_into_parts` scheme.
- `opposition-normative`: the premise `receives_evaluation(A,X2)` and the conclusion `cannot_receive_evaluation(A,X2)` are about the same evaluation X2; the premise is presumably about X. The 3-place premise uses `cannot_present_sametime`, while the unused template `cannot_opposite_evaluation/3` ("If %s receives the evaluation %s. It cannot receive the opposite evaluation %s…") looks like the intended predicate. I added a `cannot_present_sametime/3` template worded like the /2 one, and did not change the formula.
- The other opposition schemes (`opposition-descriptive_1/2/3`, `rhetorical_oppositions_descriptive`; book p. 317–318) have the content errors the audit lists (REPORT.md 3.5). I renamed their primed variables and changed nothing else; they were not checked against the book.
- `evidence_hypothesis_falsification` concludes `¬A`, which restates its premise `event_false(A)`; modus tollens gives `¬H`. Its assumption is `sufficient_correlation(A,H)`, while verification has `sufficient_correlation(H,A)` (different statements).
- `witness_testimony`: `in_domain(D,S)` reverses the order every other scheme uses (`in_domain(S,D)`). It is inherited from upstream Carneades.
- `bias`: the conditional premise `not_taken_both(X,C)` is about X, not the biased arguer A, so X is otherwise unconstrained.
- `two_person_practical`: the exceptions `alternative_means(A,X)` and `side_effect_other_act(B,A)` do not fit the order of their templates' arguments ("alternative means of realizing %s other than %s"; "Carrying out %s … %s other goals").
- `exceptional` (was `exeptional`): `case_exception(A)` is both a premise and an assumption, and `rule_can_waived(X)` uses X, which no other formula shares.
- `verbal_slippery_slope`: I declared F, the property in `individuals_property(A1,F,A2)`. Every other formula uses P for the property, so F is probably P. As written, the conditional is not linked to P.
- `threat` (was `treat`): the premise `position_to_know(X)` ("I am in position to bring about %s") shares its name with the position-to-know predicate `position_to_know/2`. Other predicates used with two arities: `data_set/1`–`/2`, `excuse/0`–`/1`, `group_member/1`–`/2`, `cannot_present_sametime/2`–`/3`, `relevant_differences/1`–`/2`. They are different statements and easy to confuse.
- `abductive_backward`: `data_set(D,A)` has 2 arguments, but `abductive_forward` has `data_set(D)`. I added a `data_set/2` template ("%s is a set of data or supposed facts in case %s."); the role of A there is a guess.
- `exception_to_rule` (was `exeption_to_rule`): the content is upstream Carneades `precedent` with a new id and title. The README says Walton renamed the scheme, but the formulas were not changed. Check against the book (AS57).
- Templates I fitted to the formula as used, where the original template suggests the formula may have lost an argument: `strongly_occur(X)` (template had the consequences C too), `cited_bad_consequences(X)` (the consequences; C occurs in `assert_bad(X,C)`), `excuse` in `regulative_rule` (template had the agent; `excuse/1` is taken by `plea_for_excuse`).
- `biased/1` had two templates. The loaded one, "Arguer %s is biased.", is kept so loading is unchanged. It is also used for witnesses and sources (`credible_source`, `witness_testimony`), where upstream's deleted "%s is biased." fits better.
- `waste` now uses `possible(X)`, whose existing template reads "Action %s is possible.", though X is a goal. `practical_reasoning2`'s `possible(G)` has the same mismatch; the deleted `posible/1` read "%s is possible.".
- Titles: `verbal_classification_vagueness` has the title of `definition_to_verbal_classification`; `rule_mental_world` has an empty title; `slippery_slope_base_case` and `slippery_slope_inductive_step` share a title; `abduction` has no `meta`.

### 0-ary atoms

These 19 atoms stay 0-ary; only their language keys got `/0`. A 0-ary atom is one statement for every instance of its scheme, so answering it once answers it for all. That is right for a general principle and wrong for anything about the instance. I gave arguments only to `reapplication_sequence`, because the gate needed it. The others change which statement the scheme asserts and are left to the owner.

| atom | scheme, role | what it is about | suggestion |
|---|---|---|---|
| `goal_prestige` | snob_appeal_ad_populum, premise | the respondent ("you"), who has no variable | leave, or add a respondent variable |
| `background` | common_folk_ad_populum, conclusion | the speaker; the scheme has no variables | see concerns (swapped) |
| `being_credible` | common_folk_ad_populum, premise | the speaker; no variables | see concerns (swapped) |
| `accept_saying` | common_folks_group_ad_populum, conclusion | "what I say"; no variable | a variable for what is said |
| `more_important_ends` | end_and_means_2, premise | general principle | fine as 0-ary |
| `not_normal` | negative_normal_expectations, conclusion | "the situation"; no variable | e.g. `not_normal(A)` |
| `sufficient_evidence` | negative_practical, assumption | the search for evidence about A | e.g. `sufficient_evidence(A)` |
| `more_reliable` | sign, exception | "the sign" = A | `more_reliable(A)` (as in cq_drafts.json) |
| `reference_class` | retroductive, assumption | "this event"; no event variable | leave |
| `attack_well_supported` | ad_hominem, assumption | the character attack on A | `attack_well_supported(A)` |
| `character_relevant` | ad_hominem, assumption | "the argument" = S | `character_relevant(S)` |
| `proper_sequence` | gradualism, assumption | the sequence P1…P4 | e.g. `proper_sequence(P1,P4)` |
| `synonymous` | rule_physical_world_1, premise | general principle | fine as 0-ary |
| `not_necessarily_true` | rule_physical_world_1, exception | "First premise" = `is(W,D)` | `not_necessarily_true(W,D)` |
| `correctly_report` | rule_physical_world_2, premise | general principle | fine as 0-ary |
| `rule_changed` | rule_physical_world_2, exception | the rule of `official_rulebook(I,D,W)` | e.g. `rule_changed(I,D,W)` |
| `speakers_correct` | rule_mental_world, premise | general principle | fine as 0-ary |
| `do_obligator` | regulative_rule, premise | general norm ("One must…") | see the regulative_rule proposal |
| `excuse` | regulative_rule, exception | the agent's excuse (template had %s) | needs a name other than `excuse/1`, e.g. `has_excuse(A)` |

## Unused variables

- **Original, declared but used in no formula:** `part_in_whole` P and `verbal_classification_vagueness` P. Each was the misnamed form of a variable the formulas use but nobody declared (A and F respectively), so the declaration was corrected (lines 684, 817), not left. The primed variables P', X', A', R' (6 schemes) were declared, but the parser cannot read them; they were renamed P2, X2, A2, R2.
- **Cleaned file, declared but used in no formula:** 0.
- **Cleaned file, used but in no premise** (the validator counts these as unused): `ad_hominem` S; `practical_necessary_condition` P, A2; `practical_sufficient_condition` P, A2; `regulative_rule` A. See above.
- **Undeclared variables in the original:** A in `part_in_whole`, F in `verbal_classification_vagueness`, X in `sacrifice` (repaired to V), F in `verbal_slippery_slope` (declared). The cleaned file has none.

## Left alone on purpose

- Grammar and wording in templates (e.g. "you will a member", "There are evidence", "If %s than %s", "an typical"). Only misspelled words and OCR splits were fixed.
- `common_folk_ad_populum` id, `do_obligator` name, `rule_can_waived` name: not misspellings.
- Titles beginning "Argumentation From …" (`sacrifice`, `regulative_rule`), empty sources (`retroductive`, `cause_to_effect`, `defeasible_modus_ponens`, `verbal_classification`), and stray spaces in sources.
- A space before `(` in `carry_out_act (A,X)` and `data_account (A1,D)`, and the double space in `-  vagueness_rejected(A)`: the parser accepts them.
- The stray " s" in the `data_account/2` template, and the soft hyphen in "pp.429­448".
- `langugage:` occurs only in the category file `B_source-based-arguments/1_…yml`, not in `walton.yml`.
- Negated formulas have no `¬p/n` WAG templates (WAG lint warning, see Gate).
- Orphan templates (no scheme uses them; 13): `action_fit_quality/0`, `another_category/2`, `bribed/1`, `cannot_opposite_evaluation/3`, `classified_under/1`, `defeasibly_implies/2`, `misleading/1`, `positive_consequences/1`, `successful_account/2`, `true_false/2`, `trustworthy/1`, `would_achieve/2`, `would_demote_value/2`.

