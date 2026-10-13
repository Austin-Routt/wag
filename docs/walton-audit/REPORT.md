# Audit: Walton argumentation schemes for Carneades 4 (Kurtuldu collection)

Date: 2026-10-13. This is groundwork for building NASCQ entries for the full collection. It is analysis only: no scheme data in the repository was changed.

**Inputs**
- **The collection.** `walton-argumentation-schemes-for-carneades-4-master` as uploaded: `walton.yml` (2068 lines), 13 category files and `README.md`. The upload labelled "expanded" is byte-identical to the one uploaded earlier (same SHA-256).
- **Baselines.**
  - Upstream Carneades `carneades-4/examples/AGs/YAML/walton.yml` (24 schemes).
  - WAG's `wag/data/schemes/core.yml` and `wag/data/nascq/*.json`.
- **The SPR** (`SPR.md`, second version): an AI-written summary of Walton, Reed & Macagno (2008). **Everything taken from it is "from the SPR, unverified against the book".**
  - Its "Claude (Opus 4.8)" catalogs were reconstructed from this same YAML, so they cannot check the YAML independently.
  - For the scheme list and the CQs, I relied mainly on the Deepseek compendium (§2) and on the Muse sections (§5–6 and the "All 60 Schemes" list in §12).

**Method**
1. **Load check.** A throwaway Go program imports each file in its own process, because the importer mutates package globals.
   - It uses the Carneades YAML importer that WAG uses (`caesYaml.Import`).
   - It runs the Carneades validator.
   - It calls `InstantiateScheme` on every scheme with dummy constants.
2. **Patched copies.** For files that failed to load, I made minimal syntax fixes in copies and re-ran the check, to find the defects behind the first error.
3. **Structural analysis.** A line-based parser recorded file:line positions and checked formulas, variables, ids and language templates. For templates it applied WAG's `LintData` rule: the number of `%s` must equal the arity. It also diffed the master file against the category files and against upstream. Its formula lists match the Carneades loader's output exactly for all 111 schemes.
4. **SPR mapping.** I took the SPR's scheme list and CQs, drafted a locus for each CQ, and checked every draft formula against the scheme it names.

## 1. Summary numbers

| | |
|---|---|
| Schemes in `walton.yml` | **111** `- id:` entries, all ids distinct: 24 from upstream Carneades, 87 added |
| Language entries in `walton.yml` | 394 lines, 392 distinct keys (`biased/1` and `good/1` are duplicated) |
| Category files | 13 files, 118 entries, 109 distinct ids |
| Does `walton.yml` load in Carneades? | **No.** YAML error at line 733. It loads after 6 one-line syntax fixes |
| After those fixes | The validator reports 137 problems. `InstantiateScheme` builds **no argument** for 7 schemes, drops a premise in 1, and creates non-ground statements in 4. 14 "Functor not defined" messages |
| Category files that load | 8 of 13. `B/1` loads with an **empty** language, because its key is spelled `langugage:` |
| Defects for `walton.yml` in `defects.json` | 274 of 471. By severity: 6 load-blocking, 45 silent wrong instantiation, 43 WAG-lint errors, 63 lint warnings, 27 content or logic, 90 cosmetic |
| SPR scheme list | 60 numbered schemes, AS1–AS60. The README claims 106 in the book; the SPR says "~65 with subtypes" |
| Coverage | 58 of the 60 AS numbers. Missing: AS50 Precedent Slippery Slope and AS53 Full Slippery Slope |
| CQs in the SPR | 138 CQs for 45 `walton.yml` schemes; none for the other 66 |
| Draft CQ loci | 102 name an existing formula (18 of which need repair first), 23 need a new formula, 13 need no formula |

## 2. Load results

Each cell gives the scheme count, the number of language keys, validator problems, and the number of schemes for which `InstantiateScheme` builds no argument.

| File | Original | After minimal syntax patch |
|---|---|---|
| `walton.yml` | FAILS: `yaml: line 733: mapping values are not allowed in this context` | loads: 111, 392, 137, 7 |
| `A_reasoning/1_deductive-reasoning.yml` | loads: 2, 7, 0, 0 | — |
| `A_reasoning/2_inductive-reasoning.yml` | FAILS: line 57, `cannot unmarshal !!str into []string` | loads: 4, 14, 0, 0 |
| `A_reasoning/3_practical-reasoning.yml` | loads: 26, 104, 45, 1 (a shadowed duplicate) | — |
| `A_reasoning/4_abductive-reasoning.yml` | loads: 9, 36, 6, 0 | — |
| `A_reasoning/5_casual-reasoning.yml` | FAILS: line 90 | loads: 13, 35, 36, 6 |
| `B_source-based-arguments/1_arguments-from-position-to-know.yml` | loads: 12, **0**, 53, 0 | — |
| `B_source-based-arguments/2_argumen-from-commitment.yml` | FAILS: line 78 | loads: 5, 19, 0, 0 |
| `B_source-based-arguments/3_attacking-personal-credibility.yml` | loads: 5, 21, 9, 0 | — |
| `B_source-based-arguments/4_argumen-from-popular-acceptence.yml` | loads: 13, 39, 11, 0 | — |
| `C_applying-rules-to-cases/1_arguments-based-on-cases.yml` | loads: 10, 31, 0, 0 | — |
| `C_applying-rules-to-cases/2_defeasible-rule-based-arguments.yml` | FAILS: lines 46 and 61 | loads: 8, 29, 20, 0 |
| `C_applying-rules-to-cases/3_chained-arguments-connecting-rules-and-cases.yml` | loads: 2, 8, 6, 0 | — |
| `C_applying-rules-to-cases/4_verbal_classification_arguments.yml` | FAILS: lines 60 and 75 | loads: 9, 33, 20, 0 |
| Upstream Carneades `walton.yml` | loads: 24, 80, 0, 0 | — |
| WAG `core.yml` | loads: 7, 32, 0, 0 | — |

**Syntax patches to `walton.yml`:**
- :732 `- id:opposition-descriptive_3` → `- id: opposition-descriptive_3`.
- :869 and :884: quoted the titles that contain `: `.
- :691, :1587 and :1602: `-smaller_part_property(...)`, `-not_necessarily_true` and `-rule_changed` → `- ...`. Without the space, YAML reads a string where a list is expected.

**Silent failures** (the file loads, but Carneades builds the wrong thing):
- **No argument at all** (7 schemes):
  - Six schemes use primed variables `P'`, `X'`, `A'`, `R'`: `opposition-descriptive_1/2/3`, `opposition-normative` and `rhetorical_oppositions_normative/descriptive`. In the term parser `'` starts a quoted atom, so every formula that contains one is dropped, conclusions included.
  - `moral_justification_ad_populum_excuse` has an empty `conclusions:`.
- **Premise dropped.** `verbal_classification_arbitrariness`: `needed_nonarbitrary((A,F)` (:838) does not parse.
- **Non-ground statements**, from undeclared variables:
  - `part_in_whole` (A)
  - `verbal_classification_vagueness` (F)
  - `sacrifice` (X)
  - `verbal_slippery_slope` (F)
- **Keys ignored without an error:**
  - `popular_opinion` uses `assumption:` (:418), so `generally_accepted(A)` is dropped.
  - `B/1` uses `langugage:`, so its whole language is dropped.
- **Shadowed duplicates.** In `A/3`, the later of two schemes with the same id wins.

## 3. Defects (`defects.json`: file, line, scheme, severity, category, detail)

### 3.1 Identity
- **Duplicate ids.** `walton.yml` has none. `A/3` has two:
  - `practical_necessary_condition` at :249 and :277. The second is the sufficient-condition schema.
  - `negative_consequences` at :157 and :511.
- **Seven schemes in both `C/2` and `C/4`.**
- **Only in category files:** `whole_into_parts`, `full_slippery_slope`, and the underscore spellings of `opposition_descriptive_2/3`.
- **The master regressed in three schemes** that the category files have right:
  - `commitment1` concludes `commined_to(A,Y)`;
  - the excuse subtype's conclusion;
  - `part_in_whole`'s variables.
- **Against upstream.** 23 of the 24 upstream schemes are identical in content. Upstream `precedent` became `exeption_to_rule`, with only the id and title changed.

### 3.2 Language templates
- **29 keys without `/arity`.** Carneades never uses them, and `LintData` reports them as errors.
- **14 templates whose number of `%s` differs from the arity**, three of them with malformed verbs:
  - `bring_about_more_effectively/2` (`% more`)
  - `is/2`
  - `not_required/1` (`%is`)
- **Duplicate keys:** `biased/1` (with different texts) and `good/1`.
- **35 predicate/arity pairs without a template.**
- **20 0-ary atoms** used as premises, assumptions, exceptions or conclusions, e.g. `more_reliable`, `background`, `excuse`. After instantiation each is one statement shared by every argument of its scheme, so answering it once answers it for all of them.
- **19 orphan templates**, e.g. `committed_to/2` and `generally_accepted/1`.
- **6 predicates used with several arities**, e.g. `position_to_know/1` vs `/2` and `relevant_differences/1` vs `/2`.
- **`true/1`**: the parser reads `true` as the boolean constant.

### 3.3 Variables
- **Declared but unused:** P in `part_in_whole` and in `verbal_classification_vagueness`.
- **Used but undeclared:** see §2.
- **Primed variables:** see §2.
- **Variables that occur in no premise** (13 cases): this matters for the NASCQ, because a support move binds every variable.

### 3.4 Conclusions
- **Empty:** `moral_justification_ad_populum_excuse`.
- **0-ary, without a template:** `common_folk_ad_populum` (`background`), `common_folks_group_ad_populum` (`accept_saying`), `negative_normal_expectations` (`not_normal`).
- **Wrong:** `evidence_hypothesis_falsification` concludes `¬A`, its own premise restated; it should conclude `¬H`.
- **Two conclusions:** `pop_scheme`.
- **Bare variables** (14 schemes). These are the natural `answer_variable` and thesis-independent candidates for the NASCQ:
  - `S`: position_to_know, expert_opinion, credible_source, witness_testimony, ethotic1, exeption_to_rule
  - `¬S`: ethotic2, ignorance
  - `A`: analogy
  - `H`: abduction, evidence_hypothesis_verification
  - `¬A`: evidence_hypothesis_falsification
  - `B`: defeasible_modus_ponens
  - `C`: established_rule

### 3.5 Content and logic (manual review)
- `division_generic` is a verbatim copy of `composition_generic`.
- `witness_testimony` has `in_domain(D,S)`, the reverse of every other scheme's `in_domain(S,D)`. It is inherited from upstream.
- `sufficient_correlation(A,H)` in falsification vs `sufficient_correlation(H,A)` in verification.
- Errors in the opposition schemes.
- Premise and conclusion swapped in several ad populum schemes, and in `fear`, `bias_ad_hominem` and `from_example`.
- `exeptional`: `case_exception(A)` is both a premise and an assumption.
- `bias`: `not_taken_both(X,C)` is not linked to `biased(A)`.
- `two_person_practical`: two exceptions have their arguments in the wrong order.
- `exeption_to_rule`: the content is upstream `precedent` and does not match AS57 in the SPR. **Uncertain.**

### 3.6 Typos (65 for `walton.yml`)
- **Ids:** `exeptional`, `exeption_to_rule`, `epistemic_ignorence`, `popular_practise`, `treat`.
- **Titles:** "Ignorence", "Treat", "Physical Worl", among others.
- **File names:** `5_casual-reasoning.yml`, `argumen-`, `acceptence`.
- **Predicates:** `commined_to`, `oficcial_rulebook`, `posible`, `recieve(s)_evaluation`, `succesful_*`, among others.
- **OCR artifacts** in templates and sources.
- **Mojibake:** `Â¬presents(A,P)` in `A/5`.

### 3.7 Metadata
Empty sources:
- `retroductive`
- `cause_to_effect`
- `defeasible_modus_ponens`
- `verbal_classification`
- `full_slippery_slope`

## 4. Scheme count against the SPR (from the SPR, unverified against the book)

`book_comparison.json` maps the SPR's 60 numbered schemes to `walton.yml` ids.

**Missing from `walton.yml`:**
- AS50 Precedent Slippery Slope.
- AS53 Full Slippery Slope (only in `A/5`).
- Division of the Whole into its Parts (only in `A/2`).
- AS57 in the book's sense.
- The ad hominem subtypes.

**Extra:**
- From upstream: `credible_source`, `defeasible_modus_ponens`, `abduction`.
- `practical_reasoning2`.
- `practical_goal` and `end_and_means_1/2/3`.

## 5. Critical questions and draft loci (`cq_drafts.json`; all "draft, needs review")

CQ texts are paraphrased from the SPR, with section and AS number as provenance. Each draft follows Gordon & Walton's rule:

| The CQ… | Draft locus |
|---|---|
| puts the burden on the proponent | a premise or assumption |
| puts the burden on the questioner | an exception |
| points to a counterargument | the conclusion (a rebuttal) |
| challenges the inference with no formula behind it | the argument (an undercutter) |

Each CQ is marked `existing` (the scheme already has the formula, checked), `new` (the scheme needs a new formula) or `none` (conclusion or argument locus). Each also has a draft NASCQ pattern, e.g. `dishonest({W})`, and a repair note where needed.

**Totals:** 138 CQs for 45 schemes:
- 102 name an existing formula, 18 of which need repair first;
- 23 need a new formula;
- 13 need no formula.

**Formulas to repair** before an edge can name them:
- the 0-ary atoms `more_reliable`, `attack_well_supported` and `character_revelant`;
- the templates of `conflict_goals/3`, `strong_advocate/3`, `strongly_occur/1`, `popular_practise/1`, `based_on_evidence` (it uses `%v`), and the duplicate `biased/1`;
- the undeclared F in `verbal_classification_vagueness`;
- `needed_nonarbitrary((A,F)`;
- the argument order in `two_person_practical`;
- `posible`.

**Observations:**
- **Agreement with WAG's NASCQ.** The drafts reproduce the loci WAG already has for practical reasoning, sign, position to know, analogy and consequences.
- **`expert_opinion` vs `credible_source`.** `expert_opinion`'s six CQs map one to one onto its own formulas. WAG's `AS:expert_opinion` points at `credible_source`, which has no `based_on_evidence`; that is why WAG's backup-evidence question is about the inference.
- **Clean mappings:** `witness_testimony`, `analogy`, `cause_to_effect`, `correlation_to_cause`.
- **Weakest mappings:** `exeption_to_rule`, `memory`, `appearance`.

## 6. Recommended cleanup plan

1. **Pick one source file.** Make `walton.yml` the source and port the fixes the category files got right. Treat the category files as generated views.
2. **Make it load.** Fix the 6 syntax errors.
3. **Make it instantiate correctly.**
   - Check with yaml.v2 `UnmarshalStrict`, which catches `assumption:` and `langugage:`.
   - Rename the primed variables.
   - Fix the unparsable formula and the undeclared variables.
   - Give the 0-ary atoms arguments.
   - Fix the empty and double conclusions.

   The gate: zero validator problems, every scheme instantiates with all its premises, no non-ground statements, and no stderr output.
4. **Make the language pass `LintData`.**
5. **Fix the content errors of §3.5.**
6. **Normalize names.** Keep an old → new id map: WAG's `core.yml` uses `practical_reasoning1`, `positive_value`, `sign`, `credible_source`, `position_to_know` and `analogy`.
7. **Verify against Walton, Reed & Macagno (2008), ch. 9**, before writing NASCQ entries. Decide on AS57, AS50, AS53, division and the ad hominem subtypes, and on whether `AS:expert_opinion` should use `expert_opinion`.
8. **Add the CQ formulas.** Add the 23 new ones (an assumption where the proponent bears the burden, an exception where the questioner does) and repair the 18 flagged ones.
9. **Write the NASCQ entries:** schemes (with `answer_variable`, descriptions, cues and evidence), shared CQ nodes, ask edges with locus and pattern, then backup edges (a separate design decision).
10. **Gate each batch** with `wag-cli --lint` plus the instantiation check. Start with the schemes whose CQs map cleanly.

## 7. Uncertainties

- **The SPR.** It is AI-generated, and its Claude catalogs derive from this YAML. Its CQ wording, numbering and page references are unchecked against the book.
- **Placements.** Several placements and draft loci are judgment calls; `cq_drafts.json` notes the alternatives.
- **`exeption_to_rule`** vs AS57 is unresolved.
- **`B/1`'s language** could not be validated.
- **The typo list** is manual.

## Files

| File | Contents |
|---|---|
| `inventory.json` | Every scheme of `walton.yml` with its formulas, ignored keys and instantiation results |
| `defects.json` | All 471 findings |
| `crossfile.json` | The master file against the category files and against upstream |
| `category_ids.json` | The ids in each category file |
| `book_comparison.json` | The SPR's list mapped to `walton.yml` ids |
| `cq_drafts.json` | Per CQ: provenance, burden, draft locus, formula, status and pattern |
