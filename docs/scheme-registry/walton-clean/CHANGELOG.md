# walton.yml cleanup: changelog

Input: `walton.yml` from `walton-argumentation-schemes-for-carneades-4-master` (2068 lines, 111 schemes, SHA-256 `1976202ac6bd538526ea75488659ec9880641ec03534d6e098f6f6dd1bcc099c`). Output: `walton.yml` (2480 lines, after a 5-line notice of the modification that the licence asks for, `tools/header.txt`): the same 111 schemes in the same order, 31 of them changed by the owner's decisions (`ad_baculum`, `analogy`, `appeal_to_vanity_ad_populum`, `cause_to_effect`, `danger`, `end_and_means_1`, `fear`, `group_2`, `moral_justification_ad_populum_excuse`, `negative_consequences`, `negative_consequences_prudential`, `negative_value`, `positive_consequences`, `practical_analogy_negative`, `practical_analogy_positive`, `practical_goal`, `practical_inference`, `practical_necessary_condition`, `practical_reasoning1`, `practical_reasoning2`, `practical_sufficient_condition`, `pragmatic_alternative`, `rhetorical_oppositions_normative`, `rule_physical_world_1`, `sacrifice`, `slippery_slope_base_case`, `slippery_slope_inductive_step`, `sorites_slippery_slope`, `threat`, `two_person_practical`, `waste`), and 19 added by them (`bad_outweighs_good`, `better_means`, `bridge_from_agent`, `bridge_from_continuing`, `bridge_from_possible_good`, `bridge_from_possible_harm`, `cause_to_effect_modal`, `end_and_means_bad`, `end_and_means_less_bad`, `finding_out_first`, `full_slippery_slope`, `good_outweighs_bad`, `harm_reduction`, `negative_consequences_modal`, `negative_consequences_reasoning`, `positive_consequences_modal`, `precedent_slippery_slope`, `sunk_costs_precommitment`, `whole_into_parts`); 130 in all. See [Changes decided by the owner](#changes-decided-by-the-owner). Every edit below is also in `changes.json` and is produced by `tools/apply_edits.py` from the list in `tools/edits.py`; `check/run.sh` rebuilds everything. Line numbers are those of the original file.

## Gate

`check/main.go` (Go, the Carneades-4 at `carneades-4`, the repository's root), one process per file. Since 2026-11-22 that is the schemes library's engine with WAG's fixes ([`CARNEADES_FORK.md`](../../../CARNEADES_FORK.md)). Check 1 also decodes with yaml.v2 `UnmarshalStrict`, which rejects unknown keys (`assumption:`) and duplicate keys. The engine's own import is now strict too. When it refuses a file, the check reads the file again leniently, as the engine before 2026-11-22 did (a duplicate key keeps its last value, an unknown field is dropped), so that checks 2–4 and the semantic diff still run; the import error stays in the report.

| | original | syntax fixes only | **cleaned** | cleaned + 2 proposals (not applied) |
|---|---|---|---|---|
| 1 YAML / import error | yaml: line 733: mapping values are not allowe | yaml: unmarshal errors: line 56: key "biased/…; read leniently | none | none |
| 1 strict-decode errors | 1 | 3 | 0 | 0 |
| 1 stderr during import | none | yes | none | none |
| 2 validator problems (before / after instantiation) | – | 142 / 155 | 5 / 5 | 0 / 0 |
| 3 schemes failing InstantiateScheme checks | – | 34 of 111 | 0 of 130 | 0 of 130 |
| 4 language problems | – | 93 | 0 | 0 |
| schemes passing checks 2–4 | – | 66 of 111 | 128 of 130 | 130 of 130 |
| **GATE** | **FAIL** | **FAIL** | **FAIL** | **PASS** |

(The validator count before instantiation for the syntax-only file includes one entry for the term-parser errors printed to stderr, and 4 for the reserved names `is` (D084) and `true` (repaired by the cleaning), which the engine's validator has reported since 2026-11-22; the audit's 137 is the same list without these 5.)

**The cleaned file does not pass the whole gate.** Checks 1, 3 and 4 pass for all 130 schemes, and check 2 passes for 128. Two fail one validator rule only: every variable must occur in a premise. In `ad_hominem` and `regulative_rule`, a variable occurs only in the conclusion or an exception. Fixing that means adding or changing a premise, which the rules for this cleanup forbid. See [Schemes that cannot pass without a change of meaning](#schemes-that-cannot-pass-without-a-change-of-meaning). D084 renamed the reserved predicate `is/2` to `is_a/2` in `group_2`, `rule_physical_world_1` and the language key, preserving the template and argument order. The engine reported three related reserved-name problems before this migration; they are now gone. These schemes still instantiate correctly (check 3), and WAG's loader does not run the validator.

WAG's own lint (`wag-cli --lint --data DIR`, with the cleaned file as `schemes/core.yml` and a copy of WAG's NASCQ data) reports **0 errors and 1 warning** for the cleaned file. The warning is `no template for ¬/1`: WAG's convention is a `¬p/n` key per negated predicate, and without one a negation reads "it is not the case that …". WAG's lint no longer loads the syntax-only file, since the engine's import became strict (`check/wag-lint-syntax-only.txt`; before 2026-11-22 it reported 58 language findings). Its NASCQ findings come from WAG's data, which targets `core.yml`, and are not about this file.

`check/semantic-diff.txt` compares what Carneades loads from the syntax-only file and from the cleaned file, after the renames below. The only scheme-level differences are the syntax fixes that restore a dropped formula (`popular_opinion`, `verbal_classification_arbitrariness`), the three variable declarations, the four D081 repairs accepted under delegated authority, and the owner's decisions (below; the added scheme is listed, not compared). The six primed-variable schemes are identical apart from the variable names.

## Diff size

| category | recorded edits | original lines touched |
|---|---|---|
| syntax | 12 | 12 |
| typo | 107 | 99 |
| variable | 24 | 22 |
| template | 46 | 44 |
| repair — needs approval | 0 | 0 |
| repair — owner decision | 167 | 141 |
| addition — owner decision | 355 | 38 |
| **total** | **711** | **318** distinct lines (a line can carry edits of several categories) |

Unified diff against the original (`walton.diff`): 263 lines removed, 674 added. That is 258 lines changed in place, 5 deleted (the duplicate keys `biased/1` and `good/1`, `posible/1`, and the conclusion that was listed as a premise), 2 inserted (the templates `cannot_present_sametime/3` and `data_set/2`); 522 of the recorded edits are authorized decisions (including the D081 conclusion move, below).

Typo edits by kind: predicate names (incl. 1 key deleted) 48, template words 26, sources 22, titles 7, scheme ids 4. The `popular_practise` rename also renames the scheme id (counted under predicate names).

## Changes decided by the owner

These change what schemes say, add one, or apply an authorized vocabulary migration. Earlier decisions were made by the owner on the date given. D081 was independently reviewed by an AI reviewer against the supplied book under the owner's delegated decision authority, not by a human annotator; its scoped repairs are described below. `tools/edits.py` records authorized choices as "repair — owner decision" and "addition — owner decision". Those category names do not claim that the owner personally checked every source page.

| scheme | change | decided | why |
|---|---|---|---|
| `negative_consequences_reasoning` (new) | added from `A_reasoning/3_practical-reasoning.yml`, lines 511–524, a second block with the id `negative_consequences`; variables `[A,B]`; conclusion `¬should_be_performed(A)` in place of the block's `not_bring_about(A)`; exception `other_consequences(A)` | 2026-10-22 | It formalizes compendium 33.3, Reasoning from Negative Consequences, which had no scheme: the master keeps only the first block of a duplicated id. The new id avoids the clash. `not_bring_about` is threat's predicate and meets no other consequence scheme, so the arguments could not rebut one another. Its two templates are in our own words, because the category file's repeat the book's premises. |
| `whole_into_parts` (new) | added from `A_reasoning/2_inductive-reasoning.yml`, under its upstream id | 2026-10-22 | It formalizes compendium 10.2, Division of the Whole into its Parts, which `walton.yml` lacked. Its formulas follow the book as printed (three parts stand for the book's n; only the positive reading). Its four templates are in our own words, because the category file's repeat the book. |
| `full_slippery_slope` (new) | added from `A_reasoning/5_casual-reasoning.yml`, under its upstream id; source filled in; `accept_all_steps` and `led_to_step` given the arguments `(C0,C2)` | 2026-10-22 | It formalizes compendium 53, the Full Slippery Slope Argument, which `walton.yml` lacked. The category file's source is empty, so the compendium's reference is used. Its two 0-ary atoms are about the series of cases at hand, and a 0-ary atom is one statement for every instance ("0-ary atoms", below). Its five templates are in our own words. |
| `precedent_slippery_slope` (new) | written for WAG: premises `claimed_exception(C0,R)`, `sets_precedent(C0,C1,R)`, `precedent_chain(C1,Cn,R)`, `intolerable_exception(Cn,R)`; conclusion `¬case_exception(C0)` | 2026-10-22 | It formalizes compendium 50, the Precedent Slippery Slope Argument, which no upstream file formalizes. One formula per printed premise, with the Related Cases Premise split into the first pair and the chain so that the book's CQ1 and CQ2 each have one. The intolerable case is taken to be the end of the series, Cn, as a slippery slope needs; the book prints C0 there and in CQ3 (the compendium, p. 340; the ch. 3 scan, p. 116, is unclear but looks the same), which can also be read as "treating C0 as an exception would be intolerable because of where it leads", a reason the chain premises then carry. A verifier with the book should confirm the reading. The conclusion negates `case_exception`, the premise of the argument for an exceptional case (`exceptional`), so that the slope meets the argument it counters (p. 117). Its templates are our own. |
| `cause_to_effect_modal` (new) | written for WAG: premises `may_occur(E1)`, `causes(E1,E2)`; exception `interference(E1)`; conclusion `may_occur(E2)` | 2026-10-22 | The modal cause to effect of audit N8. Compendium 28 has a weak reading beside the strong one (the cause occurs or might; the effect will or might follow); `cause_to_effect` formalizes only the strong reading and asserts that the cause has occurred, so the causal step of a policy not yet adopted had no scheme. It keeps `cause_to_effect`'s causal premise and exception, so the book's three questions target the same formulas, and uses `may_occur` for both the cause and the effect, so that hypothetical steps chain. The causal premise is not weakened, though the book's weak reading has a weak generalization too; the catalogue records that as open. Its template is our own. |
| `positive_consequences`, `negative_consequences` | exception `other_consequences(A)` added; the template `other_consequences/1` rewritten | 2026-10-22 | The book's question about other, opposite consequences (33/CQ3) is printed for all of 33.1–33.4; `walton.yml` had its formula only in `negative_consequences_prudential` (33.4). An exception that holds defeats the argument whatever the weights, so the template now says that the opposite consequences outweigh the ones cited (it had said only that some exist, ungrammatically). Replaced on 2026-10-23 by the sided exceptions of D055 (below). |
| `slippery_slope_base_case` | conclusion `negative_consequences(A)` → `¬should_be_performed(A)` | 2026-10-22 | No other scheme uses `negative_consequences(A)`, so a slippery slope never met the consequence and practical reasoning schemes; the book's conclusion (49) is that the first step should not be taken. |
| `slippery_slope_base_case`, `slippery_slope_inductive_step` | source year 1999 → 1992 | 2026-10-22 | The book's compendium cites the same pages (49, pp. 93, 95) as 1992, as do `walton.yml`'s three other citations of the same book, and booksellers' listings date the Vale Press edition 1992. The book's ch. 4 (p. 114) cites p. 93, for gradualism, as "Walton, 1999"; its reference list, not checked, may give that as another work. |
| `negative_value` | conclusion `reason_retaining_goal(V,G)` → `reason_retracting_goal(V,G)` (new template); the `negative_affects/3` template rewritten | 2026-10-22 | Both said the opposite of the book's 19.2: the conclusion was copied from `positive_value`, and the template said a bad value supports commitment to the goal. Both templates are in our own words. |
| `analogy` | variables `[A,C]` → `[P1,P2,C1,C2]`; premises `similar_to(C1,C2)`, `in_case(P1,C1)`, `corresponds(P1,P2)` (new template); conclusion `P2`; exception `difference_against(C1,C2,P2)` (new template) in place of `relevant_differences(C)`; `more_on_point(A,C)` removed | 2026-10-22 | Decision D001 (audit N12; WAG's proposal D001-D028-D029, version 3, after two independent reviews). The book's scheme 7 uses one letter for a claim made relative to each case; with one proposition, the premise asserted the target's claim of the source case, false whenever the two claims differ. The difference exception names the claim and tells against it, so it defeats only the argument it is about. The book's CQ3, a case like the source in which the claim failed, is a counter-analogy that rebuts the conclusion, so it needs no formula. |
| `practical_analogy_positive`, `practical_analogy_negative` | variables `[A1,A2,X]` → `[A1,A2,X1,X2]`; premises `right_to_do(A1,X1)` or `wrong_to_do(A1,X1)`, `similar_to(A1,A2)`, `corresponds(X1,X2)`; conclusion `should_be_performed(X2)` or `¬should_be_performed(X2)` in place of `carry_out_right(A2,X)` and `carry_out_wrong(A2,X)`; exception `difference_against(A1,A2,X2)` or `difference_for(A1,A2,X2)` (new templates); `right_to_do/2` and `wrong_to_do/2` rewritten | 2026-10-22 | Decisions D028 and D029, with D001. The book's 8.1 and 8.2 use one action letter made relative to each situation. The old conclusions met no other scheme; now the two rebut each other and meet the practical schemes about the same action. The book prints no questions for them; the sided exception lets a difference against one side leave the other standing when both argue from one past situation. The templates are in our own words. |
| `positive_consequences`, `negative_consequences`, `negative_consequences_reasoning`, `negative_consequences_prudential` | exception `other_consequences(A)` (or `(X)`) → `bad_outweighs(A,G)`, `good_outweighs(A,G)`, `good_outweighs(A,B)` and `bad_outweighs(X,C)` (new templates); in `negative_consequences_prudential`, assumption `strongly_occur(X)` → `strongly_occur(X,C)` (new template) and conclusion `prudential(X)` → `should_be_performed(X)` | 2026-10-23 | Decisions D055 and D059 (WAG's proposal D055-D059, version 5, after four independent reviews and a final check). The one exception for the book's 33/CQ3 was shared by the arguments for and against an action, so one argument that opposite consequences outweigh defeated every consequence argument about the action, on both sides. Each scheme now has an exception on its own side that names the consequence it cites, and so defeats only the argument it is about. The prudential form's likelihood assumption names the consequence too, and its conclusion meets the practical schemes about X. `other_consequences/1`, `prudential/1` and `strongly_occur/1` are left as orphan templates. The templates are in our own words. |
| `practical_reasoning1`, `practical_reasoning2` | variable P; premise `goal_of_agent(P,G)` (new template); assumptions `possible(A)` → `possible_bring_about(P,A)` and, in `practical_reasoning2`, `possible(G)` → `possible_goal(G)`; exceptions `bring_about_more_effectively`, `realize_more_effectively` and `promote_more_effectively` take the agent P (new templates), `side_effects(A,S1)` → `bad_outweighs(A,G)`, and `incompatible_goal(G)` → `incompatible_goal(P,G)` (new template; added to `practical_reasoning1`) | 2026-11-24 | Decisions D004, D005, D070, D088, D089 and D091 (WAG's proposal D004-D075, version 3, after two independent reviews). The book's practical reasoning is an agent's, but no formula stated the goal or whose it was, so the questions about other goals, possibility and better means could not say whose. The goal's possibility and the agent's are two claims. Any side effect, however small, defeated the argument; the exception is now the claim that the bad consequences outweigh the goal, D055's formula, and a plain side effect is a rebuttal. The templates are in our own words, and the replaced ones are left as orphans. |
| `practical_inference` | variable P; premise `have_goal(G)` → `goal_of_agent(P,G)`; assumption `possible_action(A)` → `possible_bring_about(P,A)`; exceptions `other_goals_conflict(G)` → `incompatible_goal(P,G)`, `more_efficient_alternative_actions(A,G)` → `realize_more_effectively(P,G,A)`, `incompatible_consequences(A,G)` → `bad_outweighs(A,G)` and `incompatible_consequences(P,A,G)` (new template); conclusion `carry_out_action(A)` → `should_be_performed(A)` | 2026-11-24 | Decisions D063 and D091 (the same proposal). Its conclusion met no practical scheme. It takes practical reasoning's goal premise and exceptions, and keeps its own objection, that the action works against its goal, stated of the agent. |
| `practical_necessary_condition`, `practical_sufficient_condition` | variables `[X,A,A2,P]` → `[X,G,P,S]`; the premises state the agent P, the set of means S and the goal G (`goal_of_agent(P,G)`, `alternative_action(P,S,G)` or `alternative_action_sufficient(P,S,G)`, `necessary_action(P,X,S,G)` or `sufficient_action(P,X,S,G)`, `practicality_action(P,X)`, `more_acceptable_bringing(P,G,X)`); assumptions `possible_goal(G)`, `acceptable_action(P,X)`, `possible_bring_about(P,X)`; exceptions `realize_more_effectively(P,G,X)`, in the necessary scheme `goal_achievable_without_any(P,G,S)`, `bad_outweighs(X,G)` and `incompatible_goal(P,G)`; conclusion `bringing_about_action(X)` → `should_be_performed(X)`; new templates | 2026-11-24 | Decisions D030, D090, D004, D063 and D091 (the same proposal). The agent and the other goal A2 were in no premise, so the validator failed both schemes, and `possible_action(A)` was about the goal. The sufficient scheme's alternatives premise repeated the necessary one, where the book's ch. 3 (p. 96) says each means would be enough. Read as a best choice, a strictly better means defeats the recommendation, and the necessary scheme's own objection is that the goal can be reached without any means of the set. The conclusion now meets the practical schemes about X. |
| `bad_outweighs_good`, `good_outweighs_bad` (new) | written for WAG: premises `brings_about(A,H)`, `bad(H)` or `good(H)`, `matters_more(H,G)` (new template); conclusion `bad_outweighs(A,G)` or `good_outweighs(A,G)` | 2026-11-24 | Decisions D064 and D075 (candidate C003; the same proposal). Answer schemes: each concludes a consequence scheme's sided exception (D055) from a consequence of the other sign that matters more, so that the weighing is an argument that can itself be questioned. `bad_outweighs_good` is also practical reasoning's side effect that outweighs the goal (D070). After Walton, The New Dialectic (1998), pp. 159, 170; not in the compendium. |
| `better_means` (new) | written for WAG: premises `goal_of_agent(P,G)`, `possible_bring_about(P,A2)`, `brings_about(A2,G)`, `more_effective_means(P,A2,G,A)` (new template); conclusion `realize_more_effectively(P,G,A)` | 2026-11-24 | Decisions D075 and D089 (the same proposal). The answer scheme for another means that does better, more surely, sooner or at lower cost, for the same agent and goal, which the agent can take. After the book's questions on other means (22.1, 22.4; ch. 3, p. 96); not in the compendium. |
| `positive_consequences`, `negative_consequences`, `negative_consequences_prudential` | exceptions added: `too_slight(A,G)` and `gain_discounted_by_bearer(A,G)` to `positive_consequences`, `too_slight(A,G)` and `harm_discounted_by_bearer(A,G)` to `negative_consequences`, `fails_to_avert(X,C)` to `negative_consequences_prudential` (new templates) | 2026-11-25 | Decisions D065, D069 and D053 (WAG's proposal D008-D074, version 5 as amended, after independent reviews, with the owner's declarations under D053). The consequence schemes had no exception for a consequence too slight to count, or one that counts for less because of who bears it, and the prudential form had none for an action that would not avert the harm. `too_slight` is shared by the arguments for and against the action that cite the same consequence: the owner declared it neutral as to polarity, with its scope (D053). Who bears a consequence is not neutral, so the bearer exceptions are sided, one on each side. `negative_consequences_reasoning` (33.3) is unchanged. The templates are in our own words. |
| `positive_consequences_modal`, `negative_consequences_modal` (new) | written for WAG: premises `may_bring_about(A,G)` (new template) and `good(G)` or `bad(G)`; the exceptions of the categorical scheme of the same side; conclusion `may_be_worth_performing(A,G)` or `may_be_unwise_to_perform(A,G)` (new templates) | 2026-11-25 | Decision D008 (the same proposal). A claim that an action may bring about a consequence could only be coded with `brings_about(A,G)`, which asserts more than the claim does. The modal variants assert the possibility only, and they conclude no recommendation: a bridge carries them to it. After the compendium's 33.1 and 33.2. |
| `bridge_from_possible_good`, `bridge_from_possible_harm` (new) | written for WAG: premises the modal variant's conclusion and `matters_to_decision(A,G)`; exception `worth_finding_out_first(A,G)`; conclusion `should_be_performed(A)` or `¬should_be_performed(A)` (new templates) | 2026-11-25 | Decisions D008 and D053 (the same proposal; the owner declared the exception neutral as to polarity, with its scope, after the review of version 5 as amended). A possibility reaches the recommendation only through an explicit warrant: its arguer supports the premise that the chance matters to the decision, given what G would be, and the bridge has an exception of its own, that it is worth finding out first whether A would bring about G, so that until then the possibility does not count. The exception takes the bridge away, not the decision. The bridge is an entry condition, not a weight. On the pattern of decisions D011 and D063; not in the compendium. |
| `harm_reduction` (new) | written for WAG: premises `happens_anyway(A,E)` and `less_harmful_per_occurrence(A,E)`; exceptions `preventable_otherwise(A,E)` and `no_net_reduction(A,E)`; conclusion `should_be_performed(A)` (new templates) | 2026-11-25 | Decision D072 (the same proposal; candidate C022). That E happens whether or not A is done, and that A makes each occurrence less harmful, had no scheme: it is not the prudential form (33.4), since E happens anyway, nor a weighing of consequences. The form that steers what is coming anyway is not covered. Not in the compendium. |
| `fear`, `danger`, `threat`, `ad_baculum`, `pragmatic_alternative`, `appeal_to_vanity_ad_populum` | conclusions `should_be_performed(X)` (fear, ad_baculum, pragmatic_alternative with A, the appeal to vanity) or `¬should_be_performed(X)` (danger, threat), for `ought_to_bring_about`, `should_not(Y)`, `not_bring_about`, `do_other` and `carry_out_action`; threat's premises `cited_bad_consequences(X,C)` and `position_to_bring_about(C)`, and `assert_bad(X,C)`'s template printing C first (indexed) | 2026-11-26 | Decisions D063, D060 and D062 (WAG's proposal D022-D093, after an independent review). Each recommends an action to the one addressed but concluded a predicate no other scheme concluded or used, so in the engine an appeal to fear for an action never met an argument from negative consequences against it; Walton reads these forms as species of argument from consequences, whose conclusion is a practical ought (Scare Tactics, 2000, pp. 140-143, 173, 191). Threat put the action where the threatened consequence belongs, and danger concluded about the danger, not the action (2008, pp. 333-334). The old conclusion templates are kept as orphans. |
| `practical_goal`, `two_person_practical`, `waste`, `end_and_means_1` | `practical_goal`: goal premise `goal_of_agent(P,G)`, exceptions `realize_more_effectively(P,G,X)`, `incompatible_goal(P,G)`, `bad_outweighs(X,G)`; `two_person_practical`: conclusion `should_do(A,Y)`, assumption `possible_bring_about(A,Y)`, exceptions `realize_more_effectively(A,X,Y)`, `incompatible_goal(A,X)`, `bad_outweighs(Y,X)`; `waste`: assumption `possible_goal(X)`, exception `reassessment_against_continuing(A,X)`; `end_and_means_1`: exception `means_bad_outweighs(X,Y)` | 2026-11-26 | Decisions D023, D033, D063 and D093 (the same proposal). The argument from goal borrows practical reasoning's questions in its formulas (D023); two-person practical reasoning asks four of the book's five questions in them, the fifth keeping no formula, where its exceptions named the advisor for the action and left the means out (D093); waste's second question is the critic's exception (D033) and its assumption a goal's possibility; end and means borrows the consequences' question of the opposite value, in a formula of its own, since a means may be a state (D023). |
| `sunk_costs_precommitment`, `end_and_means_bad`, `end_and_means_less_bad` (new) | the commitment form of sunk costs: premises `choice_at(T2,A)` and `precommitted(P,A,T1,T2)`, conclusion `should_do(P,A)`; the negative and graded readings of 22.6.1: `means_to_end(X,Y)`, `bad(Y)`, so `bad(X)`, with the exception `means_good_outweighs(X,Y)`; and `means_to_end(X,Y)`, `good(Y)`, `bad(X)`, so `less_bad_as_means(X,Y)` (new templates) | 2026-11-26 | Decisions D031 and D022, with D093's borrowed question for the negative reading (the same proposal). As the schemes library has them; 25 and 22.6.1 in the compendium (2008, pp. 135-136, 325-327; Walton 2002, pp. 488-491). |
| `bridge_from_agent`, `bridge_from_continuing` (new) | written for WAG: premises `should_do(P,A)` and `deliberating_agent(P,A)`; and `ought_to_continue(P,G)`, `continues_trying(A,G)` and `deliberating_agent(P,A)`, with the assumption `possible_bring_about(P,A)` and the exceptions `realize_more_effectively(P,G,A)` and `bad_outweighs(A,G)`; conclusion `should_be_performed(A)` (new templates) | 2026-11-26 | Decisions D063 and D093 (the same proposal). A conclusion that names an agent reaches the recommendation only when the decision is that agent's, a premise its arguer supports, for the exception D063's text names (an exception nobody raises would let an agent nobody checked pass); continuing binds to an action that continues the effort, a means for the goal, which practical reasoning's questions test. Not in the compendium. |
| `finding_out_first` (new) | written for WAG: premises `settles_whether(I,A,G)`, `done_in_time(I,A)` and `worth_its_costs(I,A,G)`; conclusion `worth_finding_out_first(A,G)` (new templates) | 2026-11-26 | Decision D095 (3), taken by the owner (WAG's proposal D095, version 2, after an independent review). The answer scheme for the bridges' exception: it concludes the exception from the conditions of the owner's declaration under D053, that an investigation I would settle whether A brings about G, reliably enough, before the decision on A must be taken, at a cost and risk worth bearing, so that the game can tell a made-out answer from a bare claim of knowability. No premise makes the decision wait. After Walton, The New Dialectic (1998), p. 203, and Fundamentals of Critical Argumentation (2006), p. 85; not in the compendium. |

## D081 repairs accepted after independent source review

These narrow repairs were accepted under user-delegated decision authority after an independent AI source review of the supplied book: entry 4.6 p. 312, entry 20 p. 322, entry 12.1 p. 318, and entry 51 pp. 340–341. The original binary sorites repair required the additional property scope shown here. This is not human verification or approval of every other interpretation in these records; see `docs/migrations/D081-sorites.md`.

| scheme | line | before | after | why |
|---|---|---|---|---|
| moral_justification_ad_populum_excuse | 495, 497 | `conclusions:` empty; premise `morally_acceptable_action(X,A)` | conclusion `morally_acceptable_action(X,A)` | no conclusion, so Carneades built no argument. The formula's template starts "Therefore, …", and the category file B/4 has it as the conclusion |
| sacrifice | 908 | `great_sacrifice(S,A,X)` | `great_sacrifice(S,A,V)` | X was undeclared (non-ground statement) and V occurred only in the conclusion; the template's third argument is "the value %s" |
| rhetorical_oppositions_normative | 768 | `if_treatment_required(X,A,A',X)` | `if_treatment_required(X,A,A2,X2)` | the 4th argument is "the contrary treatment %s" (X'), not X. Otherwise X' occurs only in the conclusion (validator error). The A'→A2 rename on the same line is a plain variable fix |
| sorites_slippery_slope | 1557, 306 | premise `reapplication_sequence` (0-ary); intermediate binary repair `reapplication_sequence(A2,A)` | `reapplication_sequence(A2,A,P)` with a template naming both endpoints and P | the chain links clear case A2 to A for the same property P. Binding only the endpoints could reuse a chain about a different property |

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
- **11** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_necessary_condition, practical_sufficient_condition)
  - before: _(none)_
  - after: `  acceptable_action/2: "For %s, %s is an acceptable thing to do."`
- **13** [insert after] · addition — owner decision · template for full_slippery_slope (53), in our own words (the category file's repeats the book) (used by: full_slippery_slope)
  - before: _(none)_
  - after: `  accept_all_steps/2: "Opinion runs so that whoever comes to accept one case in the series from %s to %s will be carried on to accept the next."`
- **17** · template · template has no %s for arity 1; no scheme uses the predicate, so the arity is set to fit the template instead of inventing an argument (used by: no scheme)
  - before: `  action_fit_quality/1: "Description of action actually fit the quality."`
  - after: `  action_fit_quality/0: "Description of action actually fit the quality."`
- **21** [insert after] · addition — owner decision · template for the 33.3 scheme added below, in our own words (the category file's repeats the book) (used by: negative_consequences_reasoning)
  - before: _(none)_
  - after: `  agent_bring_about/2: "Bringing about %s would lead to %s."`
- **26** · typo · misspelled predicate (used by: composition_generic, division_generic)
  - before: `  all_parts_propert/2: "All the parts %s have the property %s."`
  - after: `  all_parts_property/2: "All the parts %s have the property %s."`
- **27** · typo · template word 'commined' (used by: commitment1)
  - before: `  also_committed/2: "Generally, when an arguer is committed to %s, it can be inferred that he is also commined to %s."`
  - after: `  also_committed/2: "Generally, when an arguer is committed to %s, it can be inferred that he is also committed to %s."`
- **28** · typo · template word 'commiued' (OCR) (used by: commitment2)
  - before: `  also_committed_questioned/3: "If arguer %s has commiued herself to proposition %s at some point in a dialogue, then it may be inferred that she is also committed to proposition %s, should the question of whether it is true become an issue later in the dialogue."`
  - after: `  also_committed_questioned/3: "If arguer %s has committed herself to proposition %s at some point in a dialogue, then it may be inferred that she is also committed to proposition %s, should the question of whether it is true become an issue later in the dialogue."`
- **29** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_necessary_condition)
  - before: _(none)_
  - after: `  alternative_action/3: "%s can reasonably hold that at least one of the means %s is needed to bring about %s."`
- **29** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_sufficient_condition)
  - before: _(none)_
  - after: `  alternative_action_sufficient/3: "%s can reasonably hold that each of the means %s would be enough to bring about %s."`
- **38** · typo · misspelled predicate (used by: pragmatic_inconsistency)
  - before: `  arguement_not_accepted/2: "Therefore, %s's argument %s should not be accepted."`
  - after: `  argument_not_accepted/2: "Therefore, %s's argument %s should not be accepted."`
- **40** · typo · template word 'wil l' (used by: threat)
  - before: `  assert_bad/2: "I hereby assert that in fact I wil l see to it that %s occurs if you bring about %s."`
  - after: `  assert_bad/2: "I hereby assert that in fact I will see to it that %s occurs if you bring about %s."`
- **40** · repair — owner decision · D060: the threat names the threatened consequence C where the book does (35). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093) The template, in our own words. (used by: threat)
  - before: `  assert_bad/2: "I hereby assert that in fact I will see to it that %s occurs if you bring about %s."`
  - after: `  assert_bad/2: "I will bring about %[2]s if you go ahead with %[1]s."`
- **42** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses attack_well_supported with 0 argument(s) and the template has 0 %s (used by: ad_hominem)
  - before: `  attack_well_supported: "Allegration made in character attack premise evidence are well supported."`
  - after: `  attack_well_supported/0: "Allegration made in character attack premise evidence are well supported."`
- **42** · typo · template word 'Allegration' (used by: ad_hominem)
  - before: `  attack_well_supported/0: "Allegration made in character attack premise evidence are well supported."`
  - after: `  attack_well_supported/0: "Allegation made in character attack premise evidence are well supported."`
- **43** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses background with 0 argument(s) and the template has 0 %s (used by: common_folk_ad_populum)
  - before: `  background: "The speaker is an ordinary person, that is, share a common background with the audience."`
  - after: `  background/0: "The speaker is an ordinary person, that is, share a common background with the audience."`
- **46** [insert after] · addition — owner decision · template for the 33.3 scheme added below, in our own words (the category file's repeats the book) (used by: negative_consequences_reasoning)
  - before: _(none)_
  - after: `  bad_for_goals/1: "%s would harm the agent's goals."`
- **49** [insert after] · repair — owner decision · template for the consequence schemes' sided exceptions (D055, D059), in our own words (used by: bad_outweighs_good, bridge_from_continuing, negative_consequences_prudential, positive_consequences, positive_consequences_modal, practical_goal, practical_inference, practical_necessary_condition, practical_reasoning1, practical_reasoning2, practical_sufficient_condition, two_person_practical)
  - before: _(none)_
  - after: `  bad_outweighs/2: "The action %s has bad consequences that outweigh %s."`
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
- **60** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_reasoning1, practical_reasoning2)
  - before: _(none)_
  - after: `  bring_about_more_effectively/3: "There is a way for %s to reach the state %s more effectively than %s."`
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
- **87** [insert after] · addition — owner decision · template for the formulas of proposal D022-D093, in our own words (used by: sunk_costs_precommitment)
  - before: _(none)_
  - after: `  choice_at/2: "%s, there is a choice between %s and not doing so."`
- **88** · template · cited_bad_consequences is used with 1 argument (X); the second %s printed '%!s(MISSING)' (used by: threat)
  - before: `  cited_bad_consequences/1: "If you bring about %s, some cited bad consequences, %s, will follow."`
  - after: `  cited_bad_consequences/1: "If you bring about %s, some cited bad consequences will follow."`
- **88** [insert after] · addition — owner decision · template for precedent_slippery_slope (50), in our own words (used by: precedent_slippery_slope)
  - before: _(none)_
  - after: `  claimed_exception/2: "%s is put forward as an exception to the rule %s."`
- **88** [insert after] · repair — owner decision · template for the formulas of proposal D022-D093, in our own words (used by: threat)
  - before: _(none)_
  - after: `  cited_bad_consequences/2: "If you go ahead with %s, the bad consequence cited, %s, will follow."`
- **91** · typo · misspelled predicate (used by: circumstantial_ad_hominem)
  - before: `  commited_opposite/2: "%s is personally committed to the opposite of %s, as shown by commitments expressed in her/ his personal actions or personal circumstances expressing such commitments."`
  - after: `  committed_opposite/2: "%s is personally committed to the opposite of %s, as shown by commitments expressed in her/ his personal actions or personal circumstances expressing such commitments."`
- **97** · template · the template has 3 %s and both schemes use conflict_goals(P,A2,A) with 3 arguments (used by: no scheme)
  - before: `  conflict_goals/2: "%s has some other goals %s that conflict with %s." `
  - after: `  conflict_goals/3: "%s has some other goals %s that conflict with %s." `
- **98** [insert after] · addition — owner decision · template for the formulas of proposal D022-D093, in our own words (used by: bridge_from_continuing)
  - before: _(none)_
  - after: `  continues_trying/2: "%s would continue the effort to realize %s."`
- **102** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses correctly_report with 0 argument(s) and the template has 0 %s (used by: rule_physical_world_2)
  - before: `  correctly_report: "Oflicial rulelebooks correctly report the rules in force."`
  - after: `  correctly_report/0: "Oflicial rulelebooks correctly report the rules in force."`
- **102** · typo · template words 'Oflicial rulelebooks' (OCR) (used by: rule_physical_world_2)
  - before: `  correctly_report/0: "Oflicial rulelebooks correctly report the rules in force."`
  - after: `  correctly_report/0: "Official rulebooks correctly report the rules in force."`
- **103** [insert after] · repair — owner decision · template for the analogy family's two-case form (D001, D028, D029), in our own words (used by: analogy, practical_analogy_negative, practical_analogy_positive)
  - before: _(none)_
  - after: `  corresponds/2: "In the other case, the counterpart of %s is %s."`
- **110** [insert after] · template · abductive_backward uses data_set(D,A) with 2 arguments; only data_set/1 had a template (used by: abductive_backward, abductive_forward)
  - before: _(none)_
  - after: `  data_set/2: "%s is a set of data or supposed facts in case %s."`
- **112** [insert after] · addition — owner decision · template for the formulas of proposal D022-D093, in our own words (used by: bridge_from_agent, bridge_from_continuing)
  - before: _(none)_
  - after: `  deliberating_agent/2: "It falls to %s to decide whether to go ahead with %s."`
- **115** [insert after] · repair — owner decision · template for the analogy family's two-case form (D001, D028, D029), in our own words (used by: analogy, practical_analogy_positive)
  - before: _(none)_
  - after: `  difference_against/3: "%s and %s differ in a way that tells against %s."`
- **115** [insert after] · repair — owner decision · template for the analogy family's two-case form (D001, D028, D029), in our own words (used by: practical_analogy_negative)
  - before: _(none)_
  - after: `  difference_for/3: "%s and %s differ in a way that tells in favour of %s."`
- **118** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses do_obligator with 0 argument(s) and the template has 0 %s (used by: regulative_rule)
  - before: `  do_obligator: "One must not do what is prohibite -obligator-."`
  - after: `  do_obligator/0: "One must not do what is prohibite -obligator-."`
- **118** · typo · template words 'prohibite -obligator-' (cf. obligatory/1: 'prohibited -obligatory-') (used by: regulative_rule)
  - before: `  do_obligator/0: "One must not do what is prohibite -obligator-."`
  - after: `  do_obligator/0: "One must not do what is prohibited -obligatory-."`
- **119** [insert after] · addition — owner decision · template for the answer scheme's premises (D095), in our own words (used by: finding_out_first)
  - before: _(none)_
  - after: `  done_in_time/2: "%s can be done before the decision on action %s must be taken."`
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
- **138** [insert after] · repair — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: negative_consequences_prudential)
  - before: _(none)_
  - after: `  fails_to_avert/2: "Doing %s would fail to avert %s."`
- **139** · template · favor is used with 1 argument (A); both %s meant A, the second printed '%!s(MISSING)' (used by: popular_opinion)
  - before: `  favor/1: "If %s is generally accepted as truei that gives a reason in favor of %s."`
  - after: `  favor/1: "If %s is generally accepted as truei that gives a reason in favor of it."`
- **139** · typo · template word 'truei' (used by: popular_opinion)
  - before: `  favor/1: "If %s is generally accepted as truei that gives a reason in favor of it."`
  - after: `  favor/1: "If %s is generally accepted as true that gives a reason in favor of it."`
- **142** [insert after] · repair — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: positive_consequences, positive_consequences_modal)
  - before: _(none)_
  - after: `  gain_discounted_by_bearer/2: "In deciding on action %s, the gain %s counts for less because of whom it would go to."`
- **144** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_necessary_condition)
  - before: _(none)_
  - after: `  goal_achievable_without_any/3: "%s could reach %s without using any of the means %s."`
- **146** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: better_means, practical_goal, practical_inference, practical_necessary_condition, practical_reasoning1, practical_reasoning2, practical_sufficient_condition)
  - before: _(none)_
  - after: `  goal_of_agent/2: "One of the goals of %s is %s."`
- **147** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses goal_prestige with 0 argument(s) and the template has 0 %s (used by: snob_appeal_ad_populum)
  - before: `  goal_prestige: "Prestige is an important goal for you."`
  - after: `  goal_prestige/0: "Prestige is an important goal for you."`
- **150** [delete] · syntax · duplicate key good/1, identical to line 149 (used by: end_and_means_1, end_and_means_3, end_and_means_less_bad, good_outweighs_bad, positive_consequences, positive_consequences_modal)
  - before: `  good/1: "%s is good."`
  - after: _(none)_
- **151** · typo · template word 'represen ts' (OCR) (used by: moral_justification_ad_populum)
  - before: `  good_group/2: "Everybody who is good, or who represen ts a group %s with good qualities, accepts policy %s."`
  - after: `  good_group/2: "Everybody who is good, or who represents a group %s with good qualities, accepts policy %s."`
- **152** [insert after] · repair — owner decision · template for the consequence schemes' sided exceptions (D055, D059), in our own words (used by: good_outweighs_bad, negative_consequences, negative_consequences_modal, negative_consequences_reasoning)
  - before: _(none)_
  - after: `  good_outweighs/2: "The action %s has good consequences that outweigh %s."`
- **164** [insert after] · addition — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: harm_reduction)
  - before: _(none)_
  - after: `  happens_anyway/2: "Whether or not %s is done, %s will happen."`
- **164** [insert after] · repair — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: negative_consequences, negative_consequences_modal)
  - before: _(none)_
  - after: `  harm_discounted_by_bearer/2: "In deciding on action %s, the harm %s counts for less because of whom it would fall on."`
- **185** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_inference)
  - before: _(none)_
  - after: `  incompatible_consequences/3: "For %s, %s would have consequences that work against %s."`
- **186** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_goal, practical_inference, practical_necessary_condition, practical_reasoning1, practical_reasoning2, practical_sufficient_condition, two_person_practical)
  - before: _(none)_
  - after: `  incompatible_goal/2: "Another goal of %s is incompatible with %s."`
- **194** · typo · misspelled predicate (used by: verbal_slippery_slope)
  - before: `  individuals_propoerty/3: "If %s, has %s then, %s has it too."`
  - after: `  individuals_property/3: "If %s, has %s then, %s has it too."`
- **195** · typo · template word 'vinue' (OCR) (used by: inconsistent_commitment)
  - before: `  initial_commitment/2: "%s has claimed or indicated that he is committed to proposition %s. Generally, or by vinue of what he has said in the past."`
  - after: `  initial_commitment/2: "%s has claimed or indicated that he is committed to proposition %s. Generally, or by virtue of what he has said in the past."`
- **195** [insert after] · addition — owner decision · template for full_slippery_slope (53), in our own words (the category file's repeats the book) (used by: full_slippery_slope)
  - before: _(none)_
  - after: `  initial_presumption/1: "At the outset, %s is presumed acceptable."`
- **197** · repair — owner decision · template for the two-place exception (D071, D097), in our own words (used by: cause_to_effect, cause_to_effect_modal)
  - before: `  interference/1: "An event has occurred which interferes with event %s."`
  - after: `  interference/2: "Another causal factor in this case keeps, or would keep, %s from bringing about %s."`
- **199** [insert after] · addition — owner decision · template for precedent_slippery_slope (50), in our own words (used by: precedent_slippery_slope)
  - before: _(none)_
  - after: `  intolerable_exception/2: "Making an exception of %s to the rule %s would be intolerable."`
- **200** · template · malformed verb '% i' printed '%!i(string=...)s' (used by: no scheme)
  - before: `  is/2: "% is %s."`
  - after: `  is/2: "%s is %s."`
- **200** · repair — owner decision · D084: is/2 is reserved for GoCHR assignment; use is_a/2 for the ordinary two-place predicate, preserving its argument order and English wording. No coextensivity premise or other source reinterpretation is added. (used by: group_2, rule_physical_world_1)
  - before: `  is/2: "%s is %s."`
  - after: `  is_a/2: "%s is %s."`
- **204** [insert after] · addition — owner decision · template for full_slippery_slope (53), in our own words (the category file's repeats the book) (used by: full_slippery_slope)
  - before: _(none)_
  - after: `  leads_to_next/3: "In the chain of cases from %s through %s to %s, every case brings on the following one, through causal, precedent or analogical links."`
- **204** [insert after] · addition — owner decision · template for full_slippery_slope (53), in our own words (the category file's repeats the book) (used by: full_slippery_slope)
  - before: _(none)_
  - after: `  led_to_step/2: "Accepting a case in the series from %s to %s will in fact carry people on to the next one."`
- **205** [insert after] · addition — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: harm_reduction)
  - before: _(none)_
  - after: `  less_harmful_per_occurrence/2: "Doing %s would make each occurrence of %s less harmful."`
- **205** [insert after] · addition — owner decision · template for the formulas of proposal D022-D093, in our own words (used by: end_and_means_less_bad)
  - before: _(none)_
  - after: `  less_bad_as_means/2: "Being a means to the good end %[2]s would make %[1]s less bad."`
- **206** [insert after] · addition — owner decision · template for the answer schemes (D064, D075, D089), in our own words (used by: bad_outweighs_good, good_outweighs_bad)
  - before: _(none)_
  - after: `  matters_more/2: "%s matters more than %s."`
- **206** [insert after] · addition — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: bridge_from_possible_good, bridge_from_possible_harm)
  - before: _(none)_
  - after: `  matters_to_decision/2: "The chance that action %s brings about %s is large enough, given what that consequence would be, to bear on whether the action should be performed."`
- **206** [insert after] · addition — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: bridge_from_possible_harm, negative_consequences_modal)
  - before: _(none)_
  - after: `  may_be_unwise_to_perform/2: "Action %s may be unwise to perform because of %s."`
- **206** [insert after] · addition — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: bridge_from_possible_good, positive_consequences_modal)
  - before: _(none)_
  - after: `  may_be_worth_performing/2: "Action %s may be worth performing for the sake of %s."`
- **206** [insert after] · addition — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: negative_consequences_modal, positive_consequences_modal)
  - before: _(none)_
  - after: `  may_bring_about/2: "The action %s may bring about %s."`
- **207** [insert after] · addition — owner decision · template for cause_to_effect_modal (audit N8), in our own words (used by: cause_to_effect_modal)
  - before: _(none)_
  - after: `  may_occur/1: "An event %s has happened, or could happen."`
- **207** [insert after] · repair — owner decision · template for the formulas of proposal D022-D093, in our own words (used by: end_and_means_1)
  - before: _(none)_
  - after: `  means_bad_outweighs/2: "As a means to the good end %[2]s, %[1]s would have bad consequences that matter more than that end."`
- **208** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses means_end with 1 argument(s) and the template has 1 %s (used by: end_and_means_2)
  - before: `  means_end: "Means %s should be regarded as an end."`
  - after: `  means_end/1: "Means %s should be regarded as an end."`
- **208** [insert after] · addition — owner decision · template for the formulas of proposal D022-D093, in our own words (used by: end_and_means_bad)
  - before: _(none)_
  - after: `  means_good_outweighs/2: "As a means to the bad end %[2]s, %[1]s would have good consequences that matter more than that end."`
- **213** · typo · template word 'Appearence' (used by: no scheme)
  - before: `  misleading/1: "Appearence of the objects looking like it could be classified under %s is misleading for some reasons."`
  - after: `  misleading/1: "Appearance of the objects looking like it could be classified under %s is misleading for some reasons."`
- **216** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_necessary_condition, practical_sufficient_condition)
  - before: _(none)_
  - after: `  more_acceptable_bringing/3: "%s would rather bring about %s than leave %s undone."`
- **217** [insert after] · addition — owner decision · template for the answer schemes (D064, D075, D089), in our own words (used by: better_means)
  - before: _(none)_
  - after: `  more_effective_means/4: "For %s, %s would realize %s more effectively than %s: more surely, sooner or at lower cost."`
- **218** · typo · template word 'effecient' (used by: no scheme)
  - before: `  more_efficient_alternative_actions/2: " More effecient actions bringing about %s and that would also brin about %s."`
  - after: `  more_efficient_alternative_actions/2: " More efficient actions bringing about %s and that would also brin about %s."`
- **218** · typo · template word 'brin' (used by: no scheme)
  - before: `  more_efficient_alternative_actions/2: " More efficient actions bringing about %s and that would also brin about %s."`
  - after: `  more_efficient_alternative_actions/2: " More efficient actions bringing about %s and that would also bring about %s."`
- **219** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses more_important_ends with 0 argument(s) and the template has 0 %s (used by: end_and_means_2)
  - before: `  more_important_ends: "Ends are more important than the means used to achieve them."`
  - after: `  more_important_ends/0: "Ends are more important than the means used to achieve them."`
- **221** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses more_reliable with 0 argument(s) and the template has 0 %s (used by: sign)
  - before: `  more_reliable: "There are more reliable account for the sign."`
  - after: `  more_reliable/0: "There are more reliable account for the sign."`
- **227** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_necessary_condition)
  - before: _(none)_
  - after: `  necessary_action/4: "%s chose %s, among the means %s, as an acceptable or the best means needed for %s."`
- **229** · repair — owner decision · the template said that a bad value supports commitment to the goal, the opposite of the book's 19.2; rewritten in our own words (used by: negative_value)
  - before: `  negative_affects/3: "The fact that value %s is negative affects the interpretation and therefore the evaluation of goal %s of agent %s, if value is bad, it supports commitment to goal."`
  - after: `  negative_affects/3: "That value %s is negative bears on how goal %s of agent %s is judged: a bad value counts against commitment to the goal."`
- **230** [insert after] · addition — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: harm_reduction)
  - before: _(none)_
  - after: `  no_net_reduction/2: "Doing %s would not lower the harm that %s does overall."`
- **231** · typo · template word 'thte' (used by: negative_normal_expectations)
  - before: `  normal_true/1: "If thte situation were normal, %s would he true."`
  - after: `  normal_true/1: "If the situation were normal, %s would he true."`
- **231** · typo · template word 'he' for 'be' (OCR) (used by: negative_normal_expectations)
  - before: `  normal_true/1: "If the situation were normal, %s would he true."`
  - after: `  normal_true/1: "If the situation were normal, %s would be true."`
- **231** [insert after] · addition — owner decision · template for full_slippery_slope (53), in our own words (the category file's repeats the book) (used by: full_slippery_slope)
  - before: _(none)_
  - after: `  not_acceptable_step/1: "%s cannot be accepted after all, whatever the initial presumption."`
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
- **250** · repair — owner decision · D062: danger concludes about the action, not the danger (37). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093) The template, in our own words. (used by: danger)
  - before: `  occur/2: "If the respondent bring about %s then %s will occur."`
  - after: `  occur/2: "If the respondent goes ahead with %s, then %s will occur."`
- **251** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses oficcial_rulebook with 3 argument(s) and the template has 3 %s (used by: rule_physical_world_2)
  - before: `  oficcial_rulebook: "In the official rulebook for %s -a rule governed institution, such as a game- it is stated that %s counts as %s."`
  - after: `  oficcial_rulebook/3: "In the official rulebook for %s -a rule governed institution, such as a game- it is stated that %s counts as %s."`
- **251** · typo · misspelled predicate (used by: rule_physical_world_2)
  - before: `  oficcial_rulebook/3: "In the official rulebook for %s -a rule governed institution, such as a game- it is stated that %s counts as %s."`
  - after: `  official_rulebook/3: "In the official rulebook for %s -a rule governed institution, such as a game- it is stated that %s counts as %s."`
- **252** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093) The template, in our own words. (used by: fear)
  - before: `  only_way_prevent/2: "But the only way for you to prevent %s is bring about %s."`
  - after: `  only_way_prevent/2: "The only way for you to prevent %s is %s."`
- **255** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses other_consequences with 1 argument(s) and the template has 1 %s (used by: no scheme)
  - before: `  other_consequences: "There are some good or bad consequences of doing %s should be taken into account."`
  - after: `  other_consequences/1: "There are some good or bad consequences of doing %s should be taken into account."`
- **255** · repair — owner decision · the exception now carries the book's 33/CQ3 in the three consequence schemes (below), and an exception that holds defeats the argument whatever the weights; so it must say that the opposite consequences outweigh the cited ones, not only that some exist. The old template was also ungrammatical. In our own words (used by: no scheme)
  - before: `  other_consequences/1: "There are some good or bad consequences of doing %s should be taken into account."`
  - after: `  other_consequences/1: "Doing %s also has consequences of the opposite kind that outweigh the ones cited."`
- **256** · template · the predicate has 2 arguments (A,C); 'A %s' was the variable name plus a stray %s ('A is true') (used by: negative_practical)
  - before: `  other_consequences_presumption/2: "If I act on the presum ption that %s is not true, and A %s true, consequences %s will follow."`
  - after: `  other_consequences_presumption/2: "If I act on the presum ption that %s is not true, and it is true, consequences %s will follow."`
- **256** · typo · template word 'presum ption' (OCR) (used by: negative_practical)
  - before: `  other_consequences_presumption/2: "If I act on the presum ption that %s is not true, and it is true, consequences %s will follow."`
  - after: `  other_consequences_presumption/2: "If I act on the presumption that %s is not true, and it is true, consequences %s will follow."`
- **260** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses ought_to_continue with 2 argument(s) and the template has 2 %s (used by: bridge_from_continuing, waste)
  - before: `  ought_to_continue: "Therefore, %s ought to continue trying to realize %s."`
  - after: `  ought_to_continue/2: "Therefore, %s ought to continue trying to realize %s."`
- **260** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093) The template, in our own words. (used by: bridge_from_continuing, waste)
  - before: `  ought_to_continue/2: "Therefore, %s ought to continue trying to realize %s."`
  - after: `  ought_to_continue/2: "%s ought to continue trying to realize %s."`
- **261** [insert after] · addition — owner decision · template for whole_into_parts (10.2), in our own words (the category file's repeats the book) (used by: whole_into_parts)
  - before: _(none)_
  - after: `  part_is/2: "The part %s has %s."`
- **262** · typo · misspelled predicate (used by: part_in_whole)
  - before: `  part_propert/2: "%s is %s." `
  - after: `  part_property/2: "%s is %s." `
- **262** [insert after] · addition — owner decision · template for whole_into_parts (10.2), in our own words (the category file's repeats the book) (used by: whole_into_parts)
  - before: _(none)_
  - after: `  part_whole/5: "%s can have %s only if at least one of %s, %s and %s has it."`
- **277** · template · popular_practise is used with 1 argument (A); both %s meant A, the second printed '%!s(MISSING)' (used by: popular_practice)
  - before: `  popular_practise/1: "%s is a popular practice among those who are familiar with what is acceptable or not in regard to %s."`
  - after: `  popular_practise/1: "%s is a popular practice among those who are familiar with what is acceptable or not in regard to it."`
- **277** · typo · misspelled predicate and scheme id ('practise' for the noun 'practice') (used by: popular_practice)
  - before: `  popular_practise/1: "%s is a popular practice among those who are familiar with what is acceptable or not in regard to it."`
  - after: `  popular_practice/1: "%s is a popular practice among those who are familiar with what is acceptable or not in regard to it."`
- **277** [insert after] · repair — owner decision · template for the formulas of proposal D022-D093, in our own words (used by: threat)
  - before: _(none)_
  - after: `  position_to_bring_about/1: "I am in a position to bring about %s."`
- **278** [delete] · typo · 'posible' is a misspelling of 'possible'; renamed in its only use (line 1105), which then uses the existing template possible/1 (line 284); a second possible/1 key would be a duplicate (used by: no scheme)
  - before: `  posible/1: "%s is possible."`
  - after: _(none)_
- **281** · template · the template has 2 %s and the scheme uses position_to_know_group(G,A) (used by: position_to_know_ad_populum)
  - before: `  position_to_know_group/1: "Group %s is in a special position to know that %s is true."`
  - after: `  position_to_know_group/2: "Group %s is in a special position to know that %s is true."`
- **290** [insert after] · addition — owner decision · template for precedent_slippery_slope (50), in our own words (used by: precedent_slippery_slope)
  - before: _(none)_
  - after: `  precedent_chain/3: "From %s to %s runs a chain of cases, each so like the next that an exception for one commits us to an exception for the next under the rule %s."`
- **290** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_necessary_condition, practical_sufficient_condition)
  - before: _(none)_
  - after: `  practicality_action/2: "As far as %s can tell, nothing that cannot be changed stands in the way of bringing about %s."`
- **290** [insert after] · addition — owner decision · template for the formulas of proposal D022-D093, in our own words (used by: sunk_costs_precommitment)
  - before: _(none)_
  - after: `  precommitted/4: "What %[1]s did or undertook %[3]s is a commitment, binding %[4]s, to %[2]s."`
- **295** [insert after] · addition — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: harm_reduction)
  - before: _(none)_
  - after: `  preventable_otherwise/2: "A course open to whoever would do %s could prevent %s."`
- **295** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093) The template, in our own words. (used by: fear)
  - before: `  prevent/1: "Therefore, you ought to prevent %s if possible."
  preventable_otherwise/2: "A course open to whoever would do %s could prevent %s."`
  - after: `  prevent/1: "You ought to prevent %s if you can."
  preventable_otherwise/2: "A course open to whoever would do %s could prevent %s."`
- **296** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_reasoning1)
  - before: _(none)_
  - after: `  promote_more_effectively/3: "There is a way for %s to promote the value %s more effectively than %s."`
- **297** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses proper_sequence with 0 argument(s) and the template has 0 %s (used by: gradualism)
  - before: `  proper_sequence: "Sequence of proposition is dealt with in the appropriate order."`
  - after: `  proper_sequence/0: "Sequence of proposition is dealt with in the appropriate order."`
- **299** · typo · misspelled predicate (used by: verbal_classification_vagueness)
  - before: `  property_precion_needed/1: "%s occurs in a context of dialogue that requires a certain level of precision that is appropriate for that context."`
  - after: `  property_precision_needed/1: "%s occurs in a context of dialogue that requires a certain level of precision that is appropriate for that context."`
- **302** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses prudential with 1 argument(s) and the template has 1 %s (used by: no scheme)
  - before: `  prudential: "Therefore, you ought to reconsider other things being equal and you ought prudentially to do %s."`
  - after: `  prudential/1: "Therefore, you ought to reconsider other things being equal and you ought prudentially to do %s."`
- **304** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: better_means, bridge_from_continuing, practical_goal, practical_inference, practical_necessary_condition, practical_reasoning1, practical_reasoning2, practical_sufficient_condition, two_person_practical)
  - before: _(none)_
  - after: `  realize_more_effectively/3: "There is a way for %s to realize the goal %s more effectively than %s."`
- **306** · repair — owner decision · D081, decided under the owner's standing rule of 2026-11-22 on ChatGPT's stage review, with the book's page images checked independently (proposal D081-D084): Walton, Reed and Macagno (2008), printed pp. 340-341, entry 51: template for the property-scoped reapplication_sequence(A2,A,P) (line 1557) (used by: sorites_slippery_slope)
  - before: `  reapplication_sequence: "A sequence of modus ponens subarguments, a premise set linking premises and conclusions from the clear area through the grey area."`
  - after: `  reapplication_sequence/3: "A sequence of modus ponens subarguments, a premise set linking premises and conclusions from the clear area through the grey area."`
- **306** · repair — owner decision · D081, decided under the owner's standing rule of 2026-11-22 on ChatGPT's stage review, with the book's page images checked independently (proposal D081-D084): Walton, Reed and Macagno (2008), printed pp. 340-341, entry 51: the template names both ends of the chain and its property, in our own words (used by: sorites_slippery_slope)
  - before: `  reapplication_sequence/3: "A sequence of modus ponens subarguments, a premise set linking premises and conclusions from the clear area through the grey area."`
  - after: `  reapplication_sequence/3: "A chain of modus ponens steps leads from %s, in the clear area, through the grey area to %s, each step concerning %s."`
- **308** [insert after] · repair — owner decision · template for negative_value's new conclusion, in our own words (used by: negative_value)
  - before: _(none)_
  - after: `  reason_retracting_goal/2: "%s is a reason to give up commitment to goal %s."`
- **309** · typo · misspelled predicate (used by: memory)
  - before: `  reasonable_belive/1: "It is reasonable to believe %s."`
  - after: `  reasonable_believe/1: "It is reasonable to believe %s."`
- **310** [insert after] · repair — owner decision · template for the formulas of proposal D022-D093, in our own words (used by: waste)
  - before: _(none)_
  - after: `  reassessment_against_continuing/2: "Leaving aside past losses that cannot be recovered, a reassessment of the costs and benefits from now on goes against %s continuing to try to realize %s."`
- **312** · typo · template word 'recieves' (used by: opposition-normative)
  - before: `  recieves_evaluation/2: "%s recieves the evaluation %s." `
  - after: `  recieves_evaluation/2: "%s receives the evaluation %s." `
- **312** · typo · misspelled predicate (used by: opposition-normative)
  - before: `  recieves_evaluation/2: "%s receives the evaluation %s." `
  - after: `  receives_evaluation/2: "%s receives the evaluation %s." `
- **313** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses reference_class with 0 argument(s) and the template has 0 %s (used by: retroductive)
  - before: `  reference_class: "Reference class of other agents who also might have brought about this event and also have same character is sufficient."`
  - after: `  reference_class/0: "Reference class of other agents who also might have brought about this event and also have same character is sufficient."`
- **321** · repair — owner decision · the template nearly repeated the book's premise and had a stray comma; rewritten in our own words for the two-case form (D028, D029) (used by: practical_analogy_positive)
  - before: `  right_to_do/2: "The right thing to do in %s, was to carry out action %s."`
  - after: `  right_to_do/2: "In %s, %s was the right course."`
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
- **332** [insert after] · addition — owner decision · template for precedent_slippery_slope (50), in our own words (used by: precedent_slippery_slope)
  - before: _(none)_
  - after: `  sets_precedent/3: "Making an exception of %s would commit us to making one of %s as well, since like cases must be treated alike under the rule %s."`
- **332** [insert after] · addition — owner decision · template for the answer scheme's premises (D095), in our own words (used by: finding_out_first)
  - before: _(none)_
  - after: `  settles_whether/3: "%s would settle, reliably enough for the decision, whether action %s brings about %s."`
- **337** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses should_not with 1 argument(s) and the template has 1 %s (used by: no scheme)
  - before: `  should_not: "Therefore you should not bring about %s."`
  - after: `  should_not/1: "Therefore you should not bring about %s."`
- **344** · typo · template word 'caryying' (used by: two_person_practical)
  - before: `  situation_necessiy/4: "As %s sets the situation, %s is a necessary or sufficient condition for caryying out %s, and tells %s this."`
  - after: `  situation_necessiy/4: "As %s sets the situation, %s is a necessary or sufficient condition for carrying out %s, and tells %s this."`
- **344** · typo · misspelled predicate (used by: two_person_practical)
  - before: `  situation_necessiy/4: "As %s sets the situation, %s is a necessary or sufficient condition for carrying out %s, and tells %s this."`
  - after: `  situation_necessity/4: "As %s sets the situation, %s is a necessary or sufficient condition for carrying out %s, and tells %s this."`
- **344** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093) The template, in our own words. (used by: two_person_practical)
  - before: `  situation_necessity/4: "As %s sets the situation, %s is a necessary or sufficient condition for carrying out %s, and tells %s this."`
  - after: `  situation_necessity/4: "In the view of %s, %s would be a necessary or sufficient condition for realizing %s, a view reported to %s."`
- **346** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses speakers_correct with 0 argument(s) and the template has 0 %s (used by: rule_mental_world)
  - before: `  speakers_correct: "Speakers of a language are generally correct about word definitions in that language."`
  - after: `  speakers_correct/0: "Speakers of a language are generally correct about word definitions in that language."`
- **350** · template · special_proposition is used with 2 arguments (A,K); the first was written as the letter A (used by: epistemic_ignorance)
  - before: `  special_proposition/2: "A is a special type of proposition such that if it were true, it would normally or usually be expected to be in %s."`
  - after: `  special_proposition/2: "%s is a special type of proposition such that if it were true, it would normally or usually be expected to be in %s."`
- **353** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093) The template, in our own words. (used by: waste)
  - before: `  stop_realize_wasted/2: "If %s stops trying to realize %s now, all previous efforts will be wasted."`
  - after: `  stop_realize_wasted/2: "If %s stopped trying to realize %s now, all the previous efforts would be wasted."`
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
- **357** [insert after] · repair — owner decision · template for the consequence schemes' sided exceptions (D055, D059), in our own words (used by: negative_consequences_prudential)
  - before: _(none)_
  - after: `  strongly_occur/2: "If %s is not done, %s is likely to follow."`
- **359** · typo · misspelled predicate (used by: no scheme)
  - before: `  succesful_account/2: "%s , is the account that explains %s most successfully."`
  - after: `  successful_account/2: "%s , is the account that explains %s most successfully."`
- **360** · typo · misspelled predicate (used by: abductive_forward)
  - before: `  succesful_diagram/1: "The most plausible -strongest- argument is represented by %s."`
  - after: `  successful_diagram/1: "The most plausible -strongest- argument is represented by %s."`
- **361** · typo · template word 'availablle' (used by: abductive_backward)
  - before: `  sufficient_account/2: "%s itself as an explanation of %s apart from other alternative explanations availablle so far, is sufficient."`
  - after: `  sufficient_account/2: "%s itself as an explanation of %s apart from other alternative explanations available so far, is sufficient."`
- **362** [insert after] · repair — owner decision · template for practical reasoning's agent-scoped formulas and the answer schemes (D004, D005, D030, D064, D075, D088-D091), in our own words (used by: practical_sufficient_condition)
  - before: _(none)_
  - after: `  sufficient_action/4: "%s chose %s, among the means %s, as an acceptable or the best means that would be enough for %s."`
- **364** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses sufficient_evidence with 0 argument(s) and the template has 0 %s (used by: negative_practical)
  - before: `  sufficient_evidence: "search for the evidence progress is sufficient."`
  - after: `  sufficient_evidence/0: "search for the evidence progress is sufficient."`
- **366** · template · language key without /arity is never used by Carneades (it looks up name/arity); every scheme uses synonymous with 0 argument(s) and the template has 0 %s (used by: rule_physical_world_1)
  - before: `  synonymous: "Whenever expressions are referentially coextensive, they are synonymous."`
  - after: `  synonymous/0: "Whenever expressions are referentially coextensive, they are synonymous."`
- **367** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093) The template, in our own words. (used by: two_person_practical)
  - before: `  tells_intend/3: "%s intends to realize %s, and tells %s this."`
  - after: `  tells_intend/3: "%s, intending to realize %s, told %s so."`
- **370** [insert after] · repair — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: negative_consequences, negative_consequences_modal, positive_consequences, positive_consequences_modal)
  - before: _(none)_
  - after: `  too_slight/2: "In deciding on action %s, %s is too slight a consequence to count."`
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
- **390** [insert after] · addition — owner decision · template for whole_into_parts (10.2), in our own words (the category file's repeats the book) (used by: whole_into_parts)
  - before: _(none)_
  - after: `  whole_is/2: "The whole %s has %s."`
- **390** [insert after] · addition — owner decision · template for whole_into_parts (10.2), in our own words (the category file's repeats the book) (used by: whole_into_parts)
  - before: _(none)_
  - after: `  whole_of/4: "%s is made up of the parts %s, %s and %s."`
- **390** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093) The template, in our own words. (used by: waste)
  - before: `  waste_bad_thing/2: "If all %s 's previous attempts to realize %s are wasted, that would be a bad thing."
  whole_is/2: "The whole %s has %s."
  whole_of/4: "%s is made up of the parts %s, %s and %s."`
  - after: `  waste_bad_thing/2: "If all the previous efforts of %s to realize %s were wasted, that would be a bad thing."
  whole_is/2: "The whole %s has %s."
  whole_of/4: "%s is made up of the parts %s, %s and %s."`
- **392** [insert after] · addition — owner decision · template for the consequences' new formulas (D008, D065, D069, D072), in our own words (used by: bridge_from_possible_good, bridge_from_possible_harm, finding_out_first)
  - before: _(none)_
  - after: `  worth_finding_out_first/2: "It is worth finding out first whether action %s would bring about %s, and until then that possibility should not count in deciding on the action: it can be found out in time, reliably enough to settle the question, at a cost and risk worth bearing."`
- **393** [insert after] · addition — owner decision · template for the answer scheme's premises (D095), in our own words (used by: finding_out_first)
  - before: _(none)_
  - after: `  worth_its_costs/3: "The costs and risks of %s are worth bearing to know whether action %s brings about %s before that possibility counts in the decision."`
- **402** · repair — owner decision · the template nearly repeated the book's premise and had a stray comma; rewritten in our own words for the two-case form (D028, D029) (used by: practical_analogy_negative)
  - before: `  wrong_to_do/2: "The wrong thing to do in %s, was to carry out action %s."`
  - after: `  wrong_to_do/2: "In %s, %s was the wrong course."`

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

- **495** [insert after] · repair — owner decision · D081, decided under the owner's standing rule of 2026-11-22 on ChatGPT's stage review, with the book's page images checked independently (proposal D081-D084): Walton, Reed and Macagno (2008), printed p. 312, entry 4.6: empty conclusions: Carneades built no argument; the 'Therefore, ...' formula listed as the first premise is the conclusion (as in the category file B/4)
  - before: _(none)_
  - after: `      - morally_acceptable_action(X,A)`
- **497** [delete] · repair — owner decision · D081, decided under the owner's standing rule of 2026-11-22 on ChatGPT's stage review, with the book's page images checked independently (proposal D081-D084): Walton, Reed and Macagno (2008), printed p. 312, entry 4.6: moved to conclusions (line 495); it was the scheme's conclusion listed as a premise
  - before: `      -  morally_acceptable_action(X,A)`
  - after: _(none)_

### appeal_to_vanity_ad_populum

- **525** · repair — owner decision · D063: a recommendation to the one addressed concludes should_be_performed(A) or its negation, the claim the consequence and practical-reasoning schemes conclude, so that it meets them. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: `      - carry_out_action(X)`
  - after: `      - should_be_performed(X)`

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
- **646** · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `    variables: [A1,A2,X]`
  - after: `    variables: [A1,A2,X1,X2]`
- **648** · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `      - carry_out_right(A2,X)`
  - after: `      - should_be_performed(X2)`
- **650** · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `      - right_to_do(A1,X)`
  - after: `      - right_to_do(A1,X1)`
- **651** [insert after] · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: _(none)_
  - after: `      - corresponds(X1,X2)`
- **651** [insert after] · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: _(none)_
  - after: `    exceptions:`
- **651** [insert after] · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: _(none)_
  - after: `      - difference_against(A1,A2,X2)`

### practical_analogy_negative

- **657** · typo · source: 'lnformal' (OCR l for I)
  - before: `              Douglas Walton, lnformal Logic New York: Cambridge University Press, 1989. p.257.`
  - after: `              Douglas Walton, Informal Logic New York: Cambridge University Press, 1989. p.257.`
- **658** · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `    variables: [A1,A2,X]`
  - after: `    variables: [A1,A2,X1,X2]`
- **660** · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `      - carry_out_wrong(A2,X)`
  - after: `      - ¬should_be_performed(X2)`
- **662** · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `      - wrong_to_do(A1,X)`
  - after: `      - wrong_to_do(A1,X1)`
- **663** [insert after] · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: _(none)_
  - after: `      - corresponds(X1,X2)`
- **663** [insert after] · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: _(none)_
  - after: `    exceptions:`
- **663** [insert after] · repair — owner decision · D028, D029: the book's practical analogies use one action letter made relative to each situation; now an action in each situation (X1, X2) joined by corresponds(X1,X2), and the conclusion that the target action should (not) be performed, so that the two rebut each other and meet the practical schemes (carry_out_right and carry_out_wrong met nothing). The book prints no questions for them; they get a difference exception, sided: against the action in the positive, in its favour in the negative. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: _(none)_
  - after: `      - difference_for(A1,A2,X2)`

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

### whole_into_parts

- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `  - id: whole_into_parts`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `    meta:`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      title: Argument From Division - Division of the Whole into its Parts`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      source: >`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `              Ch. Perelman and L. Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation.`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `              Notre Dame, Ind.: University of Notre Dame Press, 1969, pp. 231-233,234-241.`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `    variables: [X,X1,X2,X3,A]`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `    conclusions:`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      - whole_is(X,A)`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `    premises:`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      - whole_of(X,X1,X2,X3)`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      - part_whole(X,A,X1,X2,X3)`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      - part_is(X1,A)`
- **705** [insert after] · addition — owner decision · compendium 10.2, Division of the Whole into its Parts, formalized only in the category file A_reasoning/2_inductive-reasoning.yml; brought in by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `     `

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
- **768** · repair — owner decision · D081, decided under the owner's standing rule of 2026-11-22 on ChatGPT's stage review, with the book's page images checked independently (proposal D081-D084): Walton, Reed and Macagno (2008), printed p. 318, entry 12.1: the 4th argument repeats X where the template says 'the contrary treatment %s', i.e. X2; otherwise X2 occurs only in the conclusion (validator error). Preserve WAG's treatment-first variable roles.
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
- **872** · typo · source: a soft hyphen (U+00AD) in place of the page range's dash
  - before: `               Journal of Logic and Computation 13 (2003), pp.429­448. `
  - after: `               Journal of Logic and Computation 13 (2003), pp. 429-448. `
- **874** · typo · source: 'lnformal' (OCR l for I)
  - before: `               between Parties without a Consensus about Values." lnformal Logic 22 (2003a), pp. 231 -245. `
  - after: `               between Parties without a Consensus about Values." Informal Logic 22 (2003a), pp. 231 -245. `

### negative_value

- **884** · syntax · unquoted title containing ': ' is a YAML error
  - before: `      title: Argument From Values - Variant 2: Negative Value`
  - after: `      title: "Argument From Values - Variant 2: Negative Value"`
- **887** · typo · source: a soft hyphen (U+00AD) in place of the page range's dash
  - before: `               Journal of Logic and Computation 13 (2003), pp.429­448. `
  - after: `               Journal of Logic and Computation 13 (2003), pp. 429-448. `
- **889** · typo · source: 'lnformal' (OCR l for I)
  - before: `               between Parties without a Consensus about Values." lnformal Logic 22 (2003a), pp. 231 -245. `
  - after: `               between Parties without a Consensus about Values." Informal Logic 22 (2003a), pp. 231 -245. `
- **892** · repair — owner decision · the conclusion was copied from positive_value and says the opposite of the book's 19.2 (the negative value is a reason to retract commitment to the goal)
  - before: `      - reason_retaining_goal(V,G)`
  - after: `      - reason_retracting_goal(V,G)`

### sacrifice

- **901** · typo · source: 'Trtalise' (OCR)
  - before: `               Ch. Perelman and L. Olbrechts-Tyteca, The New Rhetoric: A Trtalise on Argumentation. `
  - after: `               Ch. Perelman and L. Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation. `
- **908** · repair — owner decision · D081, decided under the owner's standing rule of 2026-11-22 on ChatGPT's stage review, with the book's page images checked independently (proposal D081-D084): Walton, Reed and Macagno (2008), printed p. 322, entry 20: undeclared X made the premise non-ground, and the value V occurred only in the conclusion; the template's third argument is 'the value %s', i.e. V
  - before: `      - great_sacrifice(S,A,X)`
  - after: `      - great_sacrifice(S,A,V)`

### group_2

- **933** · repair — owner decision · D084: is/2 is reserved for GoCHR assignment; use is_a/2 for the ordinary two-place predicate, preserving its argument order and English wording. No coextensivity premise or other source reinterpretation is added.
  - before: `      - is(M,Q)`
  - after: `      - is_a(M,Q)`

### practical_inference

- **949** · repair — owner decision · D063: practical inference concluded carry_out_action(A), which no practical scheme meets; now should_be_performed(A); D091: it takes practical reasoning's goal premise, possibility and exceptions, and keeps its own objection that the action works against its goal, scoped to the agent. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `    variables: [A,G]`
  - after: `    variables: [A,G,P]`
- **951** · repair — owner decision · D063: practical inference concluded carry_out_action(A), which no practical scheme meets; now should_be_performed(A); D091: it takes practical reasoning's goal premise, possibility and exceptions, and keeps its own objection that the action works against its goal, scoped to the agent. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - carry_out_action(A)`
  - after: `      - should_be_performed(A)`
- **953** · repair — owner decision · D063: practical inference concluded carry_out_action(A), which no practical scheme meets; now should_be_performed(A); D091: it takes practical reasoning's goal premise, possibility and exceptions, and keeps its own objection that the action works against its goal, scoped to the agent. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - have_goal(G)`
  - after: `      - goal_of_agent(P,G)`
- **956** · repair — owner decision · D063: practical inference concluded carry_out_action(A), which no practical scheme meets; now should_be_performed(A); D091: it takes practical reasoning's goal premise, possibility and exceptions, and keeps its own objection that the action works against its goal, scoped to the agent. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - other_goals_conflict(G)`
  - after: `      - incompatible_goal(P,G)`
- **957** · repair — owner decision · D063: practical inference concluded carry_out_action(A), which no practical scheme meets; now should_be_performed(A); D091: it takes practical reasoning's goal premise, possibility and exceptions, and keeps its own objection that the action works against its goal, scoped to the agent. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - more_efficient_alternative_actions(A,G)`
  - after: `      - realize_more_effectively(P,G,A)`
- **958** · repair — owner decision · D063: practical inference concluded carry_out_action(A), which no practical scheme meets; now should_be_performed(A); D091: it takes practical reasoning's goal premise, possibility and exceptions, and keeps its own objection that the action works against its goal, scoped to the agent. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - incompatible_consequences(A,G)`
  - after: `      - bad_outweighs(A,G)`
- **958** [insert after] · repair — owner decision · D063: practical inference concluded carry_out_action(A), which no practical scheme meets; now should_be_performed(A); D091: it takes practical reasoning's goal premise, possibility and exceptions, and keeps its own objection that the action works against its goal, scoped to the agent. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - incompatible_consequences(P,A,G)`
- **960** · repair — owner decision · D063: practical inference concluded carry_out_action(A), which no practical scheme meets; now should_be_performed(A); D091: it takes practical reasoning's goal premise, possibility and exceptions, and keeps its own objection that the action works against its goal, scoped to the agent. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `     - possible_action(A)`
  - after: `     - possible_bring_about(P,A)`

### practical_necessary_condition

- **973** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `    variables: [X,A,A2,P]`
  - after: `    variables: [X,G,P,S]`
- **975** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - bringing_about_action(X)`
  - after: `      - should_be_performed(X)`
- **977** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - goal_is_action(A)`
  - after: `      - goal_of_agent(P,G)`
- **978** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - alternative_action(X,A)`
  - after: `      - alternative_action(P,S,G)`
- **979** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - necessary_action(X,A)`
  - after: `      - necessary_action(P,X,S,G)`
- **980** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - practicality_action(X)`
  - after: `      - practicality_action(P,X)`
- **981** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - more_acceptable_bringing(A,X)`
  - after: `      - more_acceptable_bringing(P,G,X)`
- **983** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - alternative_means(A,X)`
  - after: `      - realize_more_effectively(P,G,X)`
- **983** [insert after] · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - goal_achievable_without_any(P,G,S)`
- **984** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - side_effect_action(P,X)`
  - after: `      - bad_outweighs(X,G)`
- **985** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - conflict_goals(P,A2,A)`
  - after: `      - incompatible_goal(P,G)`
- **987** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `     - possible_action(A)`
  - after: `     - possible_goal(G)`
- **988** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `     - acceptable_action(X)`
  - after: `     - acceptable_action(P,X)`
- **989** · repair — owner decision · D030, D090: the condition schemes' agent P and other goal A2 were in no premise, and possible_action(A) was about the goal; now, as the schemes library repairs them, the agent in each premise, the set of means S apart from the chosen means X, G for the goal, possible_goal(G) and possible_bring_about(P,X); the best-choice reading, a strictly better means defeating the recommendation, realize_more_effectively(P,G,X), and the objection that the goal can be reached without any means of the set, goal_achievable_without_any(P,G,S); D091: the side effect is bad_outweighs(X,G); D004: the conflicting goal incompatible_goal(P,G); D063: they conclude should_be_performed(X), which meets the practical schemes. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `     - possible_agent(X,P)`
  - after: `     - possible_bring_about(P,X)`

### practical_sufficient_condition

- **1001** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `    variables: [X,A,A2,P]`
  - after: `    variables: [X,G,P,S]`
- **1003** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - bringing_about_action(X)`
  - after: `      - should_be_performed(X)`
- **1005** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - goal_is_action(A)`
  - after: `      - goal_of_agent(P,G)`
- **1006** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - alternative_action(X,A)`
  - after: `      - alternative_action_sufficient(P,S,G)`
- **1007** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - sufficient_action(X,A)`
  - after: `      - sufficient_action(P,X,S,G)`
- **1008** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - practicality_action(X)`
  - after: `      - practicality_action(P,X)`
- **1009** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - more_acceptable_bringing(A,X)`
  - after: `      - more_acceptable_bringing(P,G,X)`
- **1011** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - alternative_means(A,X)`
  - after: `      - realize_more_effectively(P,G,X)`
- **1012** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - side_effect_action(P,X)`
  - after: `      - bad_outweighs(X,G)`
- **1013** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - conflict_goals(P,A2,A)`
  - after: `      - incompatible_goal(P,G)`
- **1015** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `     - possible_action(A)`
  - after: `     - possible_goal(G)`
- **1016** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `     - acceptable_action(X)`
  - after: `     - acceptable_action(P,X)`
- **1017** · repair — owner decision · D030, D090 for 22.3, the necessary-condition schema's twin: the same repair; its alternatives premise was the necessary-condition one, where the book's ch. 3 (p. 96) says each means would be enough, and it has no mere-alternative exception, which its own premise granted. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `     - possible_agent(X,P)`
  - after: `     - possible_bring_about(P,X)`

### practical_goal

- **1024** · typo · source: 'lntelligence' (OCR l for I)
  - before: `              Artificial lntelligence and Law 11 (2003), p. 169.`
  - after: `              Artificial Intelligence and Law 11 (2003), p. 169.`
- **1031** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: `      - person_goal(P,G)`
  - after: `      - goal_of_agent(P,G)`
- **1031** [insert after] · repair — owner decision · D023: a question borrowed from a related scheme, its source recorded in the proposal. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `    exceptions:`
- **1031** [insert after] · repair — owner decision · D023: a question borrowed from a related scheme, its source recorded in the proposal. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - realize_more_effectively(P,G,X)`
- **1031** [insert after] · repair — owner decision · D023: a question borrowed from a related scheme, its source recorded in the proposal. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - incompatible_goal(P,G)`
- **1031** [insert after] · repair — owner decision · D023: a question borrowed from a related scheme, its source recorded in the proposal. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - bad_outweighs(X,G)`

### end_and_means_1

- **1037** · typo · source: 'Trtatise' (OCR)
  - before: `              Chaim Perelman and Lucie Olbrechts-Tyteca, The New Rhetoric: A Trtatise on Argumentation `
  - after: `              Chaim Perelman and Lucie Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation `
- **1044** [insert after] · repair — owner decision · D023: a question borrowed from a related scheme, its source recorded in the proposal. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `    exceptions:`
- **1044** [insert after] · repair — owner decision · D023: a question borrowed from a related scheme, its source recorded in the proposal. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - means_bad_outweighs(X,Y)`

### end_and_means_bad

- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `  - id: end_and_means_bad`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    meta:`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      title: Ends and Means - a Bad End Condemns Its Means`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      source: >`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        Walton, Douglas and Reed, Chris and Macagno, Fabrizio (2008). Argumentation Schemes. Cambridge`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        University Press, ch. 9, entry 22.6.1 (its negative reading), p. 325. Chaim Perelman and Lucie`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation, Notre Dame, Ind.: University of`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        Notre Dame Press, 1969, pp. 273-278. As the schemes library has it (decision D022), with the`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        question of the opposite value that end and means 1 borrows (D093).`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    variables: [X,Y]`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    conclusions:`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - bad(X)`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    premises:`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - means_to_end(X,Y)`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - bad(Y)`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    exceptions:`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - means_good_outweighs(X,Y)`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, the negative reading of 22.6.1; its exception borrows the question of the opposite value, as end and means 1 does (D093(g)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: _(none)_

### end_and_means_less_bad

- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `  - id: end_and_means_less_bad`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `    meta:`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      title: Ends and Means - a Good End Makes a Bad Means Less Bad`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      source: >`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `        Walton, Douglas and Reed, Chris and Macagno, Fabrizio (2008). Argumentation Schemes. Cambridge`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `        University Press, ch. 9, entry 22.6.1 (its graded reading), p. 325. Chaim Perelman and Lucie`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `        Olbrechts-Tyteca, The New Rhetoric: A Treatise on Argumentation, Notre Dame, Ind.: University of`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `        Notre Dame Press, 1969, pp. 273-278. As the schemes library has it (decision D022).`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `    variables: [X,Y]`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `    conclusions:`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - less_bad_as_means(X,Y)`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `    premises:`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - means_to_end(X,Y)`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - good(Y)`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - bad(X)`
- **1045** [insert after] · addition — owner decision · D022: a variant of end and means 1, one reading of 22.6.1. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: _(none)_

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

- **1080** · repair — owner decision · D063: a recommendation to the one addressed concludes should_be_performed(A) or its negation, the claim the consequence and practical-reasoning schemes conclude, so that it meets them. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: `      - should_carry_out(A,Y)`
  - after: `      - should_do(A,Y)`
- **1083** · typo · misspelled predicate
  - before: `      - situation_necessiy(B,Y,X,A)`
  - after: `      - situation_necessity(B,Y,X,A)`
- **1085** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: `      - alternative_means(A,X)`
  - after: `      - realize_more_effectively(A,X,Y)`
- **1086** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: `      - goals_higher_priority(A,X)`
  - after: `      - incompatible_goal(A,X)`
- **1087** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: `      - side_effect_other_act(B,A)`
  - after: `      - bad_outweighs(Y,X)`
- **1089** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: `     - possible_action(Y)`
  - after: `     - possible_bring_about(A,Y)`

### waste

- **1105** · typo · misspelled predicate 'posible'
  - before: `     - posible(X)`
  - after: `     - possible(X)`
- **1105** · repair — owner decision · D093: found in drafting step C4 and in the review of its first version: practical reasoning's formulas for the goal and the questions, the goal's possibility, and the templates of the changed schemes. Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: `     - possible(X)`
  - after: `     - possible_goal(X)`
- **1105** [insert after] · repair — owner decision · D033: the reassessment from now on (24, CQ2), an exception, the critic's. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `    exceptions:`
- **1105** [insert after] · repair — owner decision · D033: the reassessment from now on (24, CQ2), an exception, the critic's. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - reassessment_against_continuing(A,X)`

### bridge_from_agent

- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `  - id: bridge_from_agent`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    meta:`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      title: From What an Agent Should Do to the Action (bridge)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      source: >`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        WAG's bridge (decision D063, refined 2026-11-22; D093): an explicit warrant, with a premise its`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        arguer supports, from a conclusion that names an agent, should_do(P,A), to the recommendation`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        should_be_performed(A), when it falls to P to decide whether to go ahead with A. Not in the`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        compendium.`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    variables: [P,A]`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    conclusions:`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - should_be_performed(A)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    premises:`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - should_do(P,A)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - deliberating_agent(P,A)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: _(none)_

### bridge_from_continuing

- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `  - id: bridge_from_continuing`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    meta:`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      title: From Continuing to Try to the Action (bridge)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      source: >`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        WAG's bridge (decision D063, refined 2026-11-22: continuing an activity binds to the continuing`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        action, not only to its goal; D093): from waste's conclusion, that an agent ought to continue trying`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        to realize a goal, to the recommendation of an action that continues the effort, when the decision`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        is the agent's, with practical reasoning's questions for the means it chooses. Not in the`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `        compendium.`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    variables: [P,G,A]`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    conclusions:`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - should_be_performed(A)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    premises:`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - ought_to_continue(P,G)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - continues_trying(A,G)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - deliberating_agent(P,A)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    assumptions:`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - possible_bring_about(P,A)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `    exceptions:`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - realize_more_effectively(P,G,A)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: `      - bad_outweighs(A,G)`
- **1106** [insert after] · addition — owner decision · D063: the bridge from a conclusion that names an agent to should_be_performed(A), with a premise its arguer supports, deliberating_agent(P,A), for the exception D063's text names (D093(e)); the continuing bridge asks practical reasoning's questions of the means it chooses (D093(f)). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093); D093: Decided under the owner's instruction of 2026-11-25 to go with the recommended choices (proposal D022-D093)
  - before: _(none)_
  - after: _(none)_

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

### negative_consequences_prudential

- **1289** · repair — owner decision · D055, D059: the prudential form's likelihood assumption named only the action, so two warnings about one action shared it; its conclusion prudential(X) met no other scheme. Now strongly_occur(X,C), the sided exception bad_outweighs(X,C), and the conclusion should_be_performed(X), which meets the practical schemes about X. Owner decision of 2026-10-23 (proposal D055-D059, version 5)
  - before: `      - prudential(X)`
  - after: `      - should_be_performed(X)`
- **1294** · repair — owner decision · D055, D059: the prudential form's likelihood assumption named only the action, so two warnings about one action shared it; its conclusion prudential(X) met no other scheme. Now strongly_occur(X,C), the sided exception bad_outweighs(X,C), and the conclusion should_be_performed(X), which meets the practical schemes about X. Owner decision of 2026-10-23 (proposal D055-D059, version 5)
  - before: `      - other_consequences(X)`
  - after: `      - bad_outweighs(X,C)`
- **1294** [insert after] · repair — owner decision · D069: the prudential form (33.4) gains fails_to_avert(X,C), the critic's to support. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - fails_to_avert(X,C)`
- **1296** · repair — owner decision · D055, D059: the prudential form's likelihood assumption named only the action, so two warnings about one action shared it; its conclusion prudential(X) met no other scheme. Now strongly_occur(X,C), the sided exception bad_outweighs(X,C), and the conclusion should_be_performed(X), which meets the practical schemes about X. Owner decision of 2026-10-23 (proposal D055-D059, version 5)
  - before: `     -  strongly_occur(X)`
  - after: `     -  strongly_occur(X,C)`

### pragmatic_alternative

- **1305** · repair — owner decision · D063: a recommendation to the one addressed concludes should_be_performed(A) or its negation, the claim the consequence and practical-reasoning schemes conclude, so that it meets them. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: `      - do_other(A) `
  - after: `      - should_be_performed(A) `

### treat → threat

- **1310** · typo · scheme id 'treat' for 'threat' (title: Argument From Threat)
  - before: `  - id: treat`
  - after: `  - id: threat`
- **1312** · typo · title 'Treat' for 'Threat'
  - before: `      title: Argument From Treat `
  - after: `      title: Argument From Threat `

### threat

- **1318** · repair — owner decision · D063: a recommendation to the one addressed concludes should_be_performed(A) or its negation, the claim the consequence and practical-reasoning schemes conclude, so that it meets them. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: `      - not_bring_about(X)`
  - after: `      - ¬should_be_performed(X)`
- **1320** · repair — owner decision · D060: the threat names the threatened consequence C where the book does (35). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: `      - cited_bad_consequences(X)`
  - after: `      - cited_bad_consequences(X,C)`
- **1321** · repair — owner decision · D060: the threat names the threatened consequence C where the book does (35). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: `      - position_to_know(X)`
  - after: `      - position_to_bring_about(C)`

### ad_baculum

- **1326** · typo · title 'Treat' for 'Threat'
  - before: `      title: Argument From Disjunctive Ad Baculum Treat`
  - after: `      title: Argument From Disjunctive Ad Baculum Threat`
- **1331** · repair — owner decision · D063: a recommendation to the one addressed concludes should_be_performed(A) or its negation, the claim the consequence and practical-reasoning schemes conclude, so that it meets them. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: `      - do_other(X)`
  - after: `      - should_be_performed(X)`

### fear

- **1343** · repair — owner decision · D063: a recommendation to the one addressed concludes should_be_performed(A) or its negation, the claim the consequence and practical-reasoning schemes conclude, so that it meets them. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: `      - ought_to_bring_about(X)`
  - after: `      - should_be_performed(X)`

### danger

- **1357** · repair — owner decision · D062: danger concludes about the action, not the danger (37). Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: `      - should_not(Y)`
  - after: `      - ¬should_be_performed(X)`

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
- **1557** · repair — owner decision · D081, decided under the owner's standing rule of 2026-11-22 on ChatGPT's stage review, with the book's page images checked independently (proposal D081-D084): Walton, Reed and Macagno (2008), printed pp. 340-341, entry 51: the chain links clear case A2 to A for the same property P. The former binary repair bound A but could reuse chain evidence concerning a different property.
  - before: `      - reapplication_sequence`
  - after: `      - reapplication_sequence(A2,A,P)`

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

### full_slippery_slope

- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `  - id: full_slippery_slope`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `    meta:`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `      title: Full Slippery Slope Argument`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `      source: >`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `             Douglas Walton, Slippery Slope Arguments. Newport News, Va.: Vale Press, 1992, pp. 200, 203.`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `    variables: [C0,C1,C2]`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `    conclusions:`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `      - not_acceptable_step(C0)`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `    premises:`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `      - initial_presumption(C0)`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `      - leads_to_next(C0,C1,C2)`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `      - accept_all_steps(C0,C2)`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `    assumptions:`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: `      - led_to_step(C0,C2)`
- **1573** [insert after] · addition — owner decision · compendium 53, the Full Slippery Slope Argument, formalized only in the category file A_reasoning/5_casual-reasoning.yml; brought in by the owner's decision of 2026-10-22, with the compendium's reference as its source and arguments (C0,C2) for its two 0-ary atoms, which are about the series at hand
  - before: _(none)_
  - after: _(none)_

### precedent_slippery_slope

- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `  - id: precedent_slippery_slope`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `    meta:`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `      title: Precedent Slippery Slope Argument`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `      source: >`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `             Douglas Walton, A Pragmatic Theory of Fallacy. Tuscaloosa and London: University of Alabama Press,`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `             1995, p. 159. Douglas Walton, Slippery Slope Arguments. Newport News, Va.: Vale Press, 1992, pp. 93, 95.`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `    variables: [C0,C1,Cn,R]`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `    conclusions:`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `      - ¬case_exception(C0)`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `    premises:`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `      - claimed_exception(C0,R)`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `      - sets_precedent(C0,C1,R)`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `      - precedent_chain(C1,Cn,R)`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: `      - intolerable_exception(Cn,R)`
- **1573** [insert after] · addition — owner decision · compendium 50, the Precedent Slippery Slope Argument, formalized nowhere upstream; written for WAG by the owner's decision of 2026-10-22, one formula per printed premise (the Related Cases Premise split into the first pair and the chain), the end case Cn as the intolerable one (the book prints C0; see the comment above), and the conclusion ¬case_exception(C0), which meets the argument for an exceptional case it counters (p. 117)
  - before: _(none)_
  - after: _(none)_

### rule_physical_world_1

- **1576** · typo · title 'Worl'
  - before: `      title: Argument For Constitutive Rule Claims - Physical Worl Premise Version 1 `
  - after: `      title: Argument For Constitutive Rule Claims - Physical World Premise Version 1 `
- **1584** · repair — owner decision · D084: is/2 is reserved for GoCHR assignment; use is_a/2 for the ordinary two-place predicate, preserving its argument order and English wording. No coextensivity premise or other source reinterpretation is added.
  - before: `      - is(W,D)`
  - after: `      - is_a(W,D)`
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

### analogy

- **1719** · repair — owner decision · D001 (audit N12): the single proposition A made the premise assert the target's claim of the source case, false whenever the two claims differ; now a claim in each case (P1, P2) joined by corresponds(P1,P2), similarity between the two cases, and a difference exception that tells against the claim carried over and names it. The book's CQ3, a case like the source in which the claim failed, is a counter-analogy that rebuts, so more_on_point goes. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `    variables: [A,C]`
  - after: `    variables: [P1,P2,C1,C2]`
- **1720** · repair — owner decision · D001 (audit N12): the single proposition A made the premise assert the target's claim of the source case, false whenever the two claims differ; now a claim in each case (P1, P2) joined by corresponds(P1,P2), similarity between the two cases, and a difference exception that tells against the claim carried over and names it. The book's CQ3, a case like the source in which the claim failed, is a counter-analogy that rebuts, so more_on_point goes. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `    conclusions: [A]`
  - after: `    conclusions: [P2]`
- **1722** · repair — owner decision · D001 (audit N12): the single proposition A made the premise assert the target's claim of the source case, false whenever the two claims differ; now a claim in each case (P1, P2) joined by corresponds(P1,P2), similarity between the two cases, and a difference exception that tells against the claim carried over and names it. The book's CQ3, a case like the source in which the claim failed, is a counter-analogy that rebuts, so more_on_point goes. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `      - similar_case(C)`
  - after: `      - similar_to(C1,C2)`
- **1723** · repair — owner decision · D001 (audit N12): the single proposition A made the premise assert the target's claim of the source case, false whenever the two claims differ; now a claim in each case (P1, P2) joined by corresponds(P1,P2), similarity between the two cases, and a difference exception that tells against the claim carried over and names it. The book's CQ3, a case like the source in which the claim failed, is a counter-analogy that rebuts, so more_on_point goes. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `      - in_case(A,C)`
  - after: `      - in_case(P1,C1)`
- **1723** [insert after] · repair — owner decision · D001 (audit N12): the single proposition A made the premise assert the target's claim of the source case, false whenever the two claims differ; now a claim in each case (P1, P2) joined by corresponds(P1,P2), similarity between the two cases, and a difference exception that tells against the claim carried over and names it. The book's CQ3, a case like the source in which the claim failed, is a counter-analogy that rebuts, so more_on_point goes. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: _(none)_
  - after: `      - corresponds(P1,P2)`
- **1725** · repair — owner decision · D001 (audit N12): the single proposition A made the premise assert the target's claim of the source case, false whenever the two claims differ; now a claim in each case (P1, P2) joined by corresponds(P1,P2), similarity between the two cases, and a difference exception that tells against the claim carried over and names it. The book's CQ3, a case like the source in which the claim failed, is a counter-analogy that rebuts, so more_on_point goes. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `      - relevant_differences(C)`
  - after: `      - difference_against(C1,C2,P2)`
- **1726** [delete] · repair — owner decision · D001 (audit N12): the single proposition A made the premise assert the target's claim of the source case, false whenever the two claims differ; now a claim in each case (P1, P2) joined by corresponds(P1,P2), similarity between the two cases, and a difference exception that tells against the claim carried over and names it. The book's CQ3, a case like the source in which the claim failed, is a counter-analogy that rebuts, so more_on_point goes. Owner decision of 2026-10-22 (proposal D001-D028-D029, version 3)
  - before: `      - more_on_point(A,C)`
  - after: _(none)_

### cause_to_effect

- **1740** [insert after] · repair — owner decision · D040: cause_to_effect cites the book's entry 28 (p. 328) and the reference it prints, as cause_to_effect_modal cites the reference. Decided under the owner's standing rule of 2026-11-22 (proposal D040-D071-D097)
  - before: _(none)_
  - after: `      source: >`
- **1740** [insert after] · repair — owner decision · D040: cause_to_effect cites the book's entry 28 (p. 328) and the reference it prints, as cause_to_effect_modal cites the reference. Decided under the owner's standing rule of 2026-11-22 (proposal D040-D071-D097)
  - before: _(none)_
  - after: `              Walton, Douglas and Reed, Chris and Macagno, Fabrizio (2008). Argumentation Schemes.`
- **1740** [insert after] · repair — owner decision · D040: cause_to_effect cites the book's entry 28 (p. 328) and the reference it prints, as cause_to_effect_modal cites the reference. Decided under the owner's standing rule of 2026-11-22 (proposal D040-D071-D097)
  - before: _(none)_
  - after: `              Cambridge University Press, p. 328 (scheme 28), citing Douglas Walton, A Pragmatic Theory of`
- **1740** [insert after] · repair — owner decision · D040: cause_to_effect cites the book's entry 28 (p. 328) and the reference it prints, as cause_to_effect_modal cites the reference. Decided under the owner's standing rule of 2026-11-22 (proposal D040-D071-D097)
  - before: _(none)_
  - after: `              Fallacy. Tuscaloosa and London: University of Alabama Press, 1995, pp. 140-141.`
- **1748** · repair — owner decision · D071, D097: the exception named only the cause, so an answer that something blocks one effect of a cause undercut every causal step from that cause (the engine test of proposal D055-D059's second review, 2026-11-22; the first gold's G2, 2026-11-27); now interference(E1,E2), as the book's 28/CQ3 is about the effect in the case at hand. Decided under the owner's standing rule of 2026-11-22 and by the owner on 2026-11-27 (proposal D040-D071-D097)
  - before: `      - interference(E1)`
  - after: `      - interference(E1,E2)`

### cause_to_effect_modal

- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `  - id: cause_to_effect_modal`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `    meta:`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `      title: Argument from Cause to Effect - Modal Form`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `      source: >`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `             Douglas Walton, A Pragmatic Theory of Fallacy. Tuscaloosa and London: University of Alabama Press,`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `             1995, pp. 140-141.`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `    variables: [E1,E2]`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `    conclusions:`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `      - may_occur(E2)`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `    premises:`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `      - may_occur(E1)`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `      - causes(E1,E2)`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `    exceptions:`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: `      - interference(E1)`
- **1749** [insert after] · addition — owner decision · audit N8: compendium 28's weak reading (the cause occurs or might occur; the effect might occur), which cause_to_effect does not formalize; written for WAG by the owner's decision of 2026-10-22, with cause_to_effect's causal premise and exception, and may_occur for the cause and the effect, so that hypothetical steps chain
  - before: _(none)_
  - after: _(none)_
- **1749** · repair — owner decision · D071, D097: the exception named only the cause, so an answer that something blocks one effect of a cause undercut every causal step from that cause (the engine test of proposal D055-D059's second review, 2026-11-22; the first gold's G2, 2026-11-27); now interference(E1,E2), as the book's 28/CQ3 is about the effect in the case at hand. Decided under the owner's standing rule of 2026-11-22 and by the owner on 2026-11-27 (proposal D040-D071-D097)
  - before: `
  - id: cause_to_effect_modal
    meta:
      title: Argument from Cause to Effect - Modal Form
      source: >
             Douglas Walton, A Pragmatic Theory of Fallacy. Tuscaloosa and London: University of Alabama Press,
             1995, pp. 140-141.
    variables: [E1,E2]
    conclusions:
      - may_occur(E2)
    premises:
      - may_occur(E1)
      - causes(E1,E2)
    exceptions:
      - interference(E1)
`
  - after: `
  - id: cause_to_effect_modal
    meta:
      title: Argument from Cause to Effect - Modal Form
      source: >
             Douglas Walton, A Pragmatic Theory of Fallacy. Tuscaloosa and London: University of Alabama Press,
             1995, pp. 140-141.
    variables: [E1,E2]
    conclusions:
      - may_occur(E2)
    premises:
      - may_occur(E1)
      - causes(E1,E2)
    exceptions:
      - interference(E1,E2)
`

### negative_consequences

- **1896** [insert after] · addition — owner decision · the book's 33/CQ3 (other, opposite consequences) applies to 33.1-33.4; walton.yml had its formula, the exception other_consequences, only in negative_consequences_prudential
  - before: _(none)_
  - after: `    exceptions:`
- **1896** [insert after] · addition — owner decision · the book's 33/CQ3 (other, opposite consequences) applies to 33.1-33.4; walton.yml had its formula, the exception other_consequences, only in negative_consequences_prudential
  - before: _(none)_
  - after: `      - other_consequences(A)`
- **1896** · repair — owner decision · D055: the exception other_consequences(A), shared by the arguments for and against A, let one argument that opposite consequences outweigh defeat every consequence argument about A on both sides; now an exception on the scheme's own side that names the consequence it cites, so that it defeats only the argument it is about. Owner decision of 2026-10-23 (proposal D055-D059, version 5)
  - before: `      - bad(G)
    exceptions:
      - other_consequences(A)`
  - after: `      - bad(G)
    exceptions:
      - good_outweighs(A,G)`
- **1896** [insert after] · repair — owner decision · D065: the consequence schemes gain too_slight(A,G) (how serious), neutral as to polarity (D053, declared by the owner), and the bearer exception of their side (on whom), a sided flip. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - too_slight(A,G)`
- **1896** [insert after] · repair — owner decision · D065: the consequence schemes gain too_slight(A,G) (how serious), neutral as to polarity (D053, declared by the owner), and the bearer exception of their side (on whom), a sided flip. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - harm_discounted_by_bearer(A,G)`

### negative_consequences_reasoning

- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `  - id: negative_consequences_reasoning`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `    meta:`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      title: Argument From Consequences - Reasoning From Negative Consequences`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      source: >`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `             Douglas Walton, A Pragmatic Theory of Fallacy.`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `             Tuscaloosa and London: University of Alabama Press, 1995. pp. 155-156.`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `             Douglas Walton, Scare Tactics. Dordrecht: Kluwer, 2000, p. 123.`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `    variables: [A,B]`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `    conclusions:`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      - ¬should_be_performed(A)`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `    premises:`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      - agent_bring_about(A,B)`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      - bad_for_goals(B)`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `    exceptions:`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      - other_consequences(A)`
- **1897** [insert after] · addition — owner decision · compendium 33.3 (Reasoning from Negative Consequences), formalized in the category file A_reasoning/3_practical-reasoning.yml under the duplicate id negative_consequences and missing from the master; added under a new id by the owner's decision of 2026-10-22
  - before: _(none)_
  - after: `      `
- **1897** · repair — owner decision · D055: the exception other_consequences(A), shared by the arguments for and against A, let one argument that opposite consequences outweigh defeat every consequence argument about A on both sides; now an exception on the scheme's own side that names the consequence it cites, so that it defeats only the argument it is about. Owner decision of 2026-10-23 (proposal D055-D059, version 5)
  - before: `      
  - id: negative_consequences_reasoning
    meta:
      title: Argument From Consequences - Reasoning From Negative Consequences
      source: >
             Douglas Walton, A Pragmatic Theory of Fallacy.
             Tuscaloosa and London: University of Alabama Press, 1995. pp. 155-156.
             Douglas Walton, Scare Tactics. Dordrecht: Kluwer, 2000, p. 123.
    variables: [A,B]
    conclusions:
      - ¬should_be_performed(A)
    premises:
      - agent_bring_about(A,B)
      - bad_for_goals(B)
    exceptions:
      - other_consequences(A)
      `
  - after: `      
  - id: negative_consequences_reasoning
    meta:
      title: Argument From Consequences - Reasoning From Negative Consequences
      source: >
             Douglas Walton, A Pragmatic Theory of Fallacy.
             Tuscaloosa and London: University of Alabama Press, 1995. pp. 155-156.
             Douglas Walton, Scare Tactics. Dordrecht: Kluwer, 2000, p. 123.
    variables: [A,B]
    conclusions:
      - ¬should_be_performed(A)
    premises:
      - agent_bring_about(A,B)
      - bad_for_goals(B)
    exceptions:
      - good_outweighs(A,B)
      `

### positive_consequences

- **1925** [insert after] · addition — owner decision · the book's 33/CQ3 (other, opposite consequences) applies to 33.1-33.4; walton.yml had its formula, the exception other_consequences, only in negative_consequences_prudential
  - before: _(none)_
  - after: `    exceptions:`
- **1925** [insert after] · addition — owner decision · the book's 33/CQ3 (other, opposite consequences) applies to 33.1-33.4; walton.yml had its formula, the exception other_consequences, only in negative_consequences_prudential
  - before: _(none)_
  - after: `      - other_consequences(A)`
- **1925** · repair — owner decision · D055: the exception other_consequences(A), shared by the arguments for and against A, let one argument that opposite consequences outweigh defeat every consequence argument about A on both sides; now an exception on the scheme's own side that names the consequence it cites, so that it defeats only the argument it is about. Owner decision of 2026-10-23 (proposal D055-D059, version 5)
  - before: `      - good(G)
    exceptions:
      - other_consequences(A)`
  - after: `      - good(G)
    exceptions:
      - bad_outweighs(A,G)`
- **1925** [insert after] · repair — owner decision · D065: the consequence schemes gain too_slight(A,G) (how serious), neutral as to polarity (D053, declared by the owner), and the bearer exception of their side (on whom), a sided flip. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - too_slight(A,G)`
- **1925** [insert after] · repair — owner decision · D065: the consequence schemes gain too_slight(A,G) (how serious), neutral as to polarity (D053, declared by the owner), and the bearer exception of their side (on whom), a sided flip. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - gain_discounted_by_bearer(A,G)`

### bad_outweighs_good

- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `  - id: bad_outweighs_good`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    meta:`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      title: Weighing a Bad Consequence against a Good One`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      source: >`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        WAG's answer scheme (decisions D064 and D075; candidate C003, docs/candidates.md),`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        after Walton's comparative weighing of good and bad consequences: Douglas Walton,`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        The New Dialectic: Conversational Contexts of Argument. University of Toronto Press,`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        1998, pp. 159, 170. Not in the compendium.`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    variables: [A,H,G]`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    conclusions:`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - bad_outweighs(A,G)`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    premises:`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - brings_about(A,H)`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - bad(H)`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - matters_more(H,G)`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: _(none)_

### good_outweighs_bad

- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `  - id: good_outweighs_bad`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    meta:`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      title: Weighing a Good Consequence against a Bad One`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      source: >`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        WAG's answer scheme (decisions D064 and D075; candidate C003, docs/candidates.md),`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        after Walton's comparative weighing of good and bad consequences: Douglas Walton,`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        The New Dialectic: Conversational Contexts of Argument. University of Toronto Press,`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        1998, pp. 159, 170. Not in the compendium.`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    variables: [A,H,G]`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    conclusions:`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - good_outweighs(A,G)`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    premises:`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - brings_about(A,H)`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - good(H)`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - matters_more(H,G)`
- **1926** [insert after] · addition — owner decision · D064, D075: candidate C003's answer schemes, which conclude the consequence schemes' sided exceptions (D055) from a consequence of the other sign that matters more; bad_outweighs_good is also D070's side effect that outweighs. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: _(none)_

### positive_consequences_modal

- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `  - id: positive_consequences_modal`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    meta:`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      title: Argument from Possible Positive Consequences`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      source: >`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        WAG's modal variant of the argument from positive consequences (decision D008), after Douglas`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        Walton, A Pragmatic Theory of Fallacy, The University of Alabama Press, Tuscaloosa and London, 1995,`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        pp. 155-156, and the questions on consequences in Walton, Reed and Macagno, Argumentation Schemes,`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        2008, pp. 332-333, which allow that a consequence will, may or must occur.`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    variables: [A,G]`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    conclusions:`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - may_be_worth_performing(A,G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    premises:`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - may_bring_about(A,G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - good(G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    exceptions:`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - bad_outweighs(A,G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - too_slight(A,G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - gain_discounted_by_bearer(A,G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: _(none)_

### negative_consequences_modal

- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `  - id: negative_consequences_modal`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    meta:`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      title: Argument from Possible Negative Consequences`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      source: >`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        WAG's modal variant of the argument from negative consequences (decision D008), after Douglas`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        Walton, A Pragmatic Theory of Fallacy, The University of Alabama Press, Tuscaloosa and London, 1995,`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        pp. 155-156, and the questions on consequences in Walton, Reed and Macagno, Argumentation Schemes,`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        2008, pp. 332-333, which allow that a consequence will, may or must occur.`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    variables: [A,G]`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    conclusions:`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - may_be_unwise_to_perform(A,G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    premises:`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - may_bring_about(A,G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - bad(G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    exceptions:`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - good_outweighs(A,G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - too_slight(A,G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - harm_discounted_by_bearer(A,G)`
- **1926** [insert after] · addition — owner decision · D008: a consequence the claim says only may follow cannot fill brings_about(A,G); the modal variant keeps the modality in its premise and its conclusion, which names the consequence, and meets the thesis only through its bridge. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: _(none)_

### bridge_from_possible_good

- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `  - id: bridge_from_possible_good`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    meta:`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      title: From a Possible Good Consequence to the Action (bridge)`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      source: >`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        WAG's bridge (decision D008, refined 2026-11-25), on the pattern of decisions D011 and D063: an`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        explicit warrant, with a premise its arguer supports and an exception of its own, from the modal`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        conclusion of positive_consequences_modal to the recommendation. Not in the compendium.`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    variables: [A,G]`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    conclusions:`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - should_be_performed(A)`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    premises:`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - may_be_worth_performing(A,G)`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - matters_to_decision(A,G)`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    exceptions:`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - worth_finding_out_first(A,G)`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: _(none)_

### bridge_from_possible_harm

- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `  - id: bridge_from_possible_harm`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    meta:`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      title: From a Possible Bad Consequence to Not Acting (bridge)`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      source: >`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        WAG's bridge (decision D008, refined 2026-11-25), on the pattern of decisions D011 and D063: an`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        explicit warrant, with a premise its arguer supports and an exception of its own, from the modal`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        conclusion of negative_consequences_modal to the rejection. Not in the compendium.`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    variables: [A,G]`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    conclusions:`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - ¬should_be_performed(A)`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    premises:`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - may_be_unwise_to_perform(A,G)`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - matters_to_decision(A,G)`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    exceptions:`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - worth_finding_out_first(A,G)`
- **1926** [insert after] · addition — owner decision · D008: the bridge from a modal conclusion to should_be_performed(A) or its negation, with a premise its arguer supports (the chance matters, given what G would be) and an exception of its own (it is worth finding out first), both naming A and G; an entry condition, not a weight. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: _(none)_

### harm_reduction

- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `  - id: harm_reduction`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    meta:`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      title: Argument from What Happens Anyway (harm reduction)`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      source: >`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        WAG's scheme (decision D072, refined 2026-11-25; candidate C022), not in the compendium: an outcome`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `        that happens whether or not the action is done, made less harmful each time by the action.`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    variables: [A,E]`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    conclusions:`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - should_be_performed(A)`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    premises:`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - happens_anyway(A,E)`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - less_harmful_per_occurrence(A,E)`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `    exceptions:`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - preventable_otherwise(A,E)`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: `      - no_net_reduction(A,E)`
- **1926** [insert after] · addition — owner decision · D072: the argument from what happens anyway: E happens whether or not A is done, and A makes each occurrence less harmful; its exceptions, a course open to whoever would do A could prevent E, and the harm overall would not fall. Decided under the owner's standing rule of 2026-11-22, refined by the owner on 2026-11-25 (proposal D008-D074, version 5)
  - before: _(none)_
  - after: _(none)_

### finding_out_first

- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `  - id: finding_out_first`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `    meta:`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `      title: Argument for Finding Out First`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `      source: >`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `        WAG's answer scheme (decision D095 (3), 2026-11-26), on the owner's declaration under D053`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `        (2026-11-25) of the bridges' exception worth_finding_out_first(A,G): it concludes the exception`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `        from the conditions the declaration names. Not in the compendium. After Douglas Walton, The New`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `        Dialectic: Conversational Contexts of Argument, University of Toronto Press, Toronto, 1998, p. 203`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `        (a shift to information-seeking within a critical discussion), and Douglas Walton, Fundamentals of`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `        Critical Argumentation, Cambridge University Press, Cambridge, 2006, p. 85 (delaying a decision to`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `        collect information).`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `    variables: [A,G,I]`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `    conclusions:`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `      - worth_finding_out_first(A,G)`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `    premises:`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `      - settles_whether(I,A,G)`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `      - done_in_time(I,A)`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: `      - worth_its_costs(I,A,G)`
- **1926** [insert after] · addition — owner decision · D095 (3): WAG's answer scheme for the bridges' exception worth_finding_out_first(A,G), from the conditions of the owner's declaration under D053 (2026-11-25): the investigation would settle the question, in time, at a cost and risk worth bearing. Decided by the owner on 2026-11-26 (proposal D095)
  - before: _(none)_
  - after: _(none)_

### practical_reasoning1

- **1935** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `    variables: [A,S1,S2,G,V]`
  - after: `    variables: [A,S1,S2,G,V,P]`
- **1938** [insert after] · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - goal_of_agent(P,G)`
- **1946** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - possible(A)`
  - after: `      - possible_bring_about(P,A)`
- **1948** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - bring_about_more_effectively(S2,A)`
  - after: `      - bring_about_more_effectively(P,S2,A)`
- **1949** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - realize_more_effectively(G,A)`
  - after: `      - realize_more_effectively(P,G,A)`
- **1950** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - promote_more_effectively(V,A)`
  - after: `      - promote_more_effectively(P,V,A)`
- **1951** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - side_effects(A,S1)`
  - after: `      - bad_outweighs(A,G)`
- **1951** [insert after] · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - incompatible_goal(P,G)`

### practical_reasoning2

- **1959** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `    variables: [A,S1,S2,G]`
  - after: `    variables: [A,S1,S2,G,P]`
- **1962** [insert after] · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - goal_of_agent(P,G)`
- **1967** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - possible(A)`
  - after: `      - possible_bring_about(P,A)`
- **1968** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - possible(G)`
  - after: `      - possible_goal(G)`
- **1970** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - bring_about_more_effectively(S2,A)`
  - after: `      - bring_about_more_effectively(P,S2,A)`
- **1971** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - realize_more_effectively(G,A)`
  - after: `      - realize_more_effectively(P,G,A)`
- **1973** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - side_effects(A,S1)`
  - after: `      - bad_outweighs(A,G)`
- **1974** · repair — owner decision · D005: practical reasoning states the agent's goal, goal_of_agent(P,G), which no formula stated; D004: the conflicting-goal exception is scoped to the agent and the goal, in both schemes, for one formula; D088: the goal's possibility and the agent's are two claims, possible_goal(G) and possible_bring_about(P,A); D089: the exceptions that another action would do better name the agent; D070, D091: the side-effect exception side_effects(A,S1) let any side effect defeat the argument however small, and is now the claim that A's bad consequences outweigh G, D055's bad_outweighs(A,G), a plain side effect being a rebuttal. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: `      - incompatible_goal(G)`
  - after: `      - incompatible_goal(P,G)`

### better_means

- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `  - id: better_means`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    meta:`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      title: Argument from a Better Means`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      source: >`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        WAG's answer scheme (decisions D075 and D089), the sketch better_means of WAG's review of`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        2026-10-15 (part 6, §4), repaired; after the book's questions on other means and efficiency`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        (Walton, Reed and Macagno, Argumentation Schemes, 2008: 22.1 and 22.4, and ch. 3, p. 96).`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `        Not in the compendium.`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    variables: [A,A2,G,P]`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    conclusions:`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - realize_more_effectively(P,G,A)`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `    premises:`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - goal_of_agent(P,G)`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - possible_bring_about(P,A2)`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - brings_about(A2,G)`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: `      - more_effective_means(P,A2,G,A)`
- **1975** [insert after] · addition — owner decision · D075, D089: the answer scheme for another means that does better, concluding realize_more_effectively(P,G,A), the agent bound by the answered argument's goal premise and able to take the other means. Decided under the owner's standing rule of 2026-11-22 and the owner's answers of 2026-11-24 (proposal D004-D075, version 3)
  - before: _(none)_
  - after: _(none)_

### exeption_to_rule → exception_to_rule

- **1976** · typo · scheme id 'exeption'
  - before: `  - id: exeption_to_rule`
  - after: `  - id: exception_to_rule`

### slippery_slope_base_case

- **1997** · repair — owner decision · the book's compendium cites these pages (49, pp. 93, 95) as 1992, as do walton.yml's three other citations of the same book, and booksellers' listings date the Vale Press edition 1992; the book's ch. 4 (p. 114) cites p. 93 for gradualism as 'Walton, 1999', which the book's reference list (not checked) may give as another work
  - before: `        Douglas Walton, Slippery Slope Arguments, Vale Press, Newport News, 1999, pp. 93, 95.`
  - after: `        Douglas Walton, Slippery Slope Arguments, Vale Press, Newport News, 1992, pp. 93, 95.`
- **2000** · repair — owner decision · the conclusion is used by no other scheme, so a slippery slope never meets the consequence and practical reasoning schemes' should_be_performed(A); the book's conclusion (49) is that A should not be brought about
  - before: `      - negative_consequences(A)`
  - after: `      - ¬should_be_performed(A)`

### slippery_slope_inductive_step

- **2009** · repair — owner decision · the book's compendium cites these pages (49, pp. 93, 95) as 1992, as do walton.yml's three other citations of the same book, and booksellers' listings date the Vale Press edition 1992; the book's ch. 4 (p. 114) cites p. 93 for gradualism as 'Walton, 1999', which the book's reference list (not checked) may give as another work
  - before: `        Douglas Walton, Slippery Slope Arguments, Vale Press, Newport News, 1999, pp. 93, 95.`
  - after: `        Douglas Walton, Slippery Slope Arguments, Vale Press, Newport News, 1992, pp. 93, 95.`

### sunk_costs_precommitment

- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `  - id: sunk_costs_precommitment`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `    meta:`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      title: Argument from Sunk Costs (the commitment form)`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      source: >`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `        Walton, Douglas and Reed, Chris and Macagno, Fabrizio (2008). Argumentation Schemes. Cambridge`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `        University Press, pp. 135-136, and ch. 9, entry 25, pp. 326-327. Douglas Walton, 'The Sunk Costs`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `        Fallacy or Argument from Waste', Argumentation 16 (2002), pp. 488-489. As the schemes library has it`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `        (decision D031).`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `    variables: [P,A,T1,T2]`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `    conclusions:`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - should_do(P,A)`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `    premises:`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - choice_at(T2,A)`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: `      - precommitted(P,A,T1,T2)`
- **2033** [insert after] · addition — owner decision · D031: the commitment form of sunk costs (25), a scheme of its own. Decided under the owner's standing rule of 2026-11-22 (proposal D022-D093)
  - before: _(none)_
  - after: _(none)_

## Schemes that cannot pass without a change of meaning

Both instantiate correctly (every premise, assumption and exception, a conclusion, all ground), and their templates pass. They fail only the validator rule that each variable must occur in a premise. The rule exists because Carneades applies schemes by matching premises. `InstantiateScheme` with explicit parameters, which WAG uses, does not need it.

| scheme | variable(s) in no premise | where they occur | proposal (in `tools/proposals.py`, not applied) |
|---|---|---|---|
| `ad_hominem` | S (the argument) | conclusion `bad_not_accepted(A,S)`, exception `stronger_than_character(S)` | add premise `advocates(A,S)` ("%s advocates argument %s."), which the conclusion "A's argument S" presupposes. Walton's scheme has the argument only in the conclusion, so this adds a premise |
| `regulative_rule` | A (agent) | conclusion `must_do(A,X)` | Grennan's scheme names the agent only in the conclusion. Proposal: apply the norm premise to A, `do_obligator(A)` ("%s must not do what is prohibited -obligatory-."). This turns the general norm into a claim about A |

With these 5 edits (`walton.proposals.yml`), the two schemes pass the validator (`proposals-check.txt`). The validator reports 0 problems.

`practical_necessary_condition` and `practical_sufficient_condition` failed the same rule until 2026-11-24: their agent P and other goal A2 were in no premise. Proposal D004-D075 (decisions D030 and D090) states the agent in each premise and replaces A2's exception with `incompatible_goal(P,G)`, so both pass; the repairs this file proposed for them were withdrawn then.

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
- `threat` (was `treat`): the premise `position_to_know(X)` ("I am in position to bring about %s") shares its name with the position-to-know predicate `position_to_know/2`. Other predicates used with two arities: `data_set/1`–`/2`, `excuse/0`–`/1`, `group_member/1`–`/2`, `cannot_present_sametime/2`–`/3`, and `relevant_differences/1`–`/2` until the analogy decision of 2026-10-22 left `relevant_differences/1` unused. They are different statements and easy to confuse.
- `abductive_backward`: `data_set(D,A)` has 2 arguments, but `abductive_forward` has `data_set(D)`. I added a `data_set/2` template ("%s is a set of data or supposed facts in case %s."); the role of A there is a guess.
- `exception_to_rule` (was `exeption_to_rule`): the content is upstream Carneades `precedent` with a new id and title. The README says Walton renamed the scheme, but the formulas were not changed. Check against the book (AS57). Its exception `relevant_differences(R,C)` puts the rule where a case belongs (WAG's decision D058).
- Templates I fitted to the formula as used, where the original template suggests the formula may have lost an argument: `strongly_occur(X)` (template had the consequences C too; since D055, on 2026-10-23, the formula is `strongly_occur(X,C)`), `cited_bad_consequences(X)` (the consequences; C occurs in `assert_bad(X,C)`), `excuse` in `regulative_rule` (template had the agent; `excuse/1` is taken by `plea_for_excuse`).
- `biased/1` had two templates. The loaded one, "Arguer %s is biased.", is kept so loading is unchanged. It is also used for witnesses and sources (`credible_source`, `witness_testimony`), where upstream's deleted "%s is biased." fits better.
- `waste` now uses `possible(X)`, whose existing template reads "Action %s is possible.", though X is a goal. `practical_reasoning2`'s `possible(G)` had the same mismatch until 2026-11-24, when it became `possible_goal(G)` (D088); the deleted `posible/1` read "%s is possible.".
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
- **Cleaned file, used but in no premise** (the validator counts these as unused): `ad_hominem` S; `regulative_rule` A. See above.
- **Undeclared variables in the original:** A in `part_in_whole`, F in `verbal_classification_vagueness`, X in `sacrifice` (repaired to V), F in `verbal_slippery_slope` (declared). The cleaned file has none.

## Left alone on purpose

- Grammar and wording in templates (e.g. "you will a member", "There are evidence", "If %s than %s", "an typical"). Only misspelled words and OCR splits were fixed.
- `common_folk_ad_populum` id, `do_obligator` name, `rule_can_waived` name: not misspellings.
- Titles beginning "Argumentation From …" (`sacrifice`, `regulative_rule`), empty sources (`retroductive`, `cause_to_effect`, `defeasible_modus_ponens`, `verbal_classification`), and stray spaces in sources.
- A space before `(` in `carry_out_act (A,X)` and `data_account (A1,D)`, and the double space in `-  vagueness_rejected(A)`: the parser accepts them.
- The stray " s" in the `data_account/2` template, and the soft hyphen in "pp.429­448".
- `langugage:` occurs only in the category file `B_source-based-arguments/1_…yml`, not in `walton.yml`.
- Negated formulas have no `¬p/n` WAG templates (WAG lint warning, see Gate).
- Orphan templates (no scheme uses them; 55): `acceptable_action/1`, `action_fit_quality/0`, `alternative_action/2`, `alternative_means/2`, `another_category/2`, `bribed/1`, `bring_about_more_effectively/2`, `bringing_about_action/1`, `cannot_opposite_evaluation/3`, `carry_out_action/1`, `carry_out_right/2`, `carry_out_wrong/2`, `cited_bad_consequences/1`, `classified_under/1`, `conflict_goals/3`, `defeasibly_implies/2`, `do_other/1`, `goal_is_action/1`, `goals_higher_priority/2`, `have_goal/1`, `incompatible_consequences/2`, `incompatible_goal/1`, `misleading/1`, `more_acceptable_bringing/2`, `more_efficient_alternative_actions/2`, `more_on_point/2`, `necessary_action/2`, `negative_consequences/1`, `not_bring_about/1`, `other_consequences/1`, `other_goals_conflict/1`, `ought_to_bring_about/1`, `person_goal/2`, `position_to_know/1`, `positive_consequences/1`, `possible/1`, `possible_action/1`, `possible_agent/2`, `practicality_action/1`, `promote_more_effectively/2`, `prudential/1`, `realize_more_effectively/2`, `relevant_differences/1`, `should_carry_out/2`, `should_not/1`, `side_effect_action/2`, `side_effect_other_act/2`, `side_effects/2`, `strongly_occur/1`, `successful_account/2`, `sufficient_action/2`, `true_false/2`, `trustworthy/1`, `would_achieve/2`, `would_demote_value/2`.

