# Argumentation schemes in Walton, Reed & Macagno (2008), pp. i–138

A page-referenced catalogue of every scheme the four scanned PDFs state explicitly (premises and conclusion), with their critical questions (CQs), draft CQ loci, and a comparison with WAG's NASCQ data and the walton.yml collection. The machine-readable version is `catalogue.json` in this directory.

## Scope and method

- **Source:** Walton, Reed & Macagno, *Argumentation Schemes* (Cambridge UP, 2008). The excerpt covers the front matter, chapters 1–3, and chapter 4 up to p. 138 (in sec. 4.4, Arguments from Inconsistency). Ethotic, circumstantial ad hominem and bias schemes start on p. 140 or later, so they are **not** in the excerpt.
- **Page mapping:** Part1 pdf N = book p. N−10 (from pdf 11; contents pp. v–ix = pdf 3–7); Part2 = pdf+24 (pp. 25–68); Part3 = pdf+68 (pp. 69–104); Part4 = pdf+104 (pp. 105–138). Checked against the running heads.
- **Method:** I read every page from book p. 1 to p. 138, plus the contents pages, as images, and transcribed from the images. Blurred indices were re-rendered at 220–400 dpi. The OCR layer was used only to locate items and as a cross-check (`qa_ocr_crosscheck.txt`).
- **Conventions:**
  - Text is as printed; book typos are kept and noted. Italics are not reproduced.
  - Lowered indices are written as Unicode subscripts (Sᵢ, Bₙ, aₖ₋₁, C₀). Digits printed inline stay inline (C1, f1, S1, t1).
  - Premise labels are given as printed. '—' means the line is unlabelled.
- **Draft locus**, after Gordon & Walton:
  - **premise** or **assumption**: the proponent bears the burden. *Premise* restates a stated premise; *assumption* is an unstated one.
  - **exception**: the questioner bears the burden.
  - **conclusion**: the CQ points to a counterargument (a rebuttal).
  - **argument**: the CQ challenges the inference with no statement behind it (an undercutter).
  - Where the book itself comments on a CQ's burden or role, the comment is quoted. Every locus is a draft.

**Counts:**
- 75 stated scheme entries, in 22 families. Variants and restatements are separate entries, e.g. expert opinion versions I–IV and the 2002 form; 8 entries are quoted from other authors.
- 22 distinct CQ lists with 75 distinct CQs. Counting repeated lists, 87 CQs are printed.
- 21 expert-opinion subquestions.
- Draft loci of the distinct CQs: argument 3, assumption 11, conclusion 7, exception 19, premise 35.

## 1. Schemes table

| ID | Book name | Book p. | PDF | CQs | WAG | walton.yml |
|---|---|---|---|---|---|---|
| S01 | Argument from Sign | 10 | Part1 p.20 | – | AS:sign | sign |
| S02 | Argument from Position to Know | 13 | Part1 p.23 | 3 | AS:position_to_know | position_to_know |
| S03 | Argument from Expert Opinion | 14, 15 | Part1 p.24, Part1 p.25 | 6 | AS:expert_opinion | expert_opinion, credible_source |
| S04 | argument from position to know (as recalled on p. 17) | 17 | Part1 p.27 | 3 | AS:position_to_know | position_to_know |
| S05 | Appeal to Expert Opinion (Version II) | 19 | Part1 p.29 | →S03 | AS:expert_opinion | expert_opinion |
| S06 | Appeal to Expert Opinion (Version III) | 19, 20 | Part1 p.29, Part1 p.30 | →S03 | AS:expert_opinion | expert_opinion |
| S07 | Appeal to Expert Opinion (Version IV) | 20 | Part1 p.30 | →S03 | AS:expert_opinion | expert_opinion |
| S08 | argument from analogy: Beardsley's formalization | 49 | Part2 p.25 | – | (AS:analogy) | – |
| S09 | argument from analogy: Burbridge's scheme (Primary Subject, Analogue, Targeted Predicate) | 50 | Part2 p.26 | – | (AS:analogy) | – |
| S10 | Burbridge's objection to an argument from analogy | 51 | Part2 p.27 | – | (AS:analogy) | – |
| S11 | deductive argument by analogy (Waller; 'deduction from principle') | 53, 54 | Part2 p.29, Part2 p.30 | – | (AS:analogy) | – |
| S12 | inductive argument from analogy (Waller) | 54 | Part2 p.30 | – | (AS:analogy) | – |
| S13 | Argument from Analogy: Version 1 | 56, 62 | Part2 p.32, Part2 p.38 | 4 | AS:analogy | analogy |
| S14 | Guarini's core scheme for argument from analogy | 57 | Part2 p.33 | – | (AS:analogy) | – |
| S15 | Guarini's second (extended) scheme for argument from analogy | 58 | Part2 p.34 | – | (AS:analogy) | – |
| S16 | Argument from Analogy: Version 2 | 58 | Part2 p.34 | →S13 | AS:analogy | analogy |
| S17 | Deductive Form of Argument from Classification | 66 | Part2 p.42 | – | – | verbal_classification |
| S18 | Scheme for Defeasible Argument from a Verbal Classification | 68 | Part2 p.44 | 2 | – | verbal_classification |
| S19 | Argument from Analogy Based on Classification | 70 | Part3 p.2 | – | (AS:analogy) | – |
| S20 | Scheme for Argument from an Established Rule | 71 | Part3 p.3 | 3 | – | established_rule |
| S21 | Scheme for Argument from an Exceptional Case | 72 | Part3 p.4 | 3 | – | exeptional |
| S22 | Refutation Scheme for Argument from Precedent | 73, 74 | Part3 p.5, Part3 p.6 | 3 | – | – |
| S23 | Positive Scheme for Practical Argument from Analogy | 74 | Part3 p.6 | – | (AS:analogy) | practical_analogy_positive |
| S24 | Negative Scheme for Practical Argument from Analogy | 74 | Part3 p.6 | – | (AS:analogy) | practical_analogy_negative |
| S25 | argument from position to know (Walton, 2002) | 89, 90 | Part3 p.21, Part3 p.22 | 3 | AS:position_to_know | position_to_know |
| S26 | argument from witness testimony | 90, 91 | Part3 p.22, Part3 p.23 | 5 | (AS:position_to_know) | witness_testimony |
| S27 | argument from expert opinion (Walton, 2002) | 91, 92, 93 | Part3 p.23, Part3 p.24, Part3 p.25 | 6 | AS:expert_opinion | expert_opinion |
| S28 | argumentation scheme for practical reasoning (Goal Premise / Means Premise) | 95, 96 | Part3 p.27, Part3 p.28 | 5 | AS:practical_reasoning | practical_inference, practical_reasoning2 |
| S29 | Necessary-Condition Schema | 95, 96 | Part3 p.27, Part3 p.28 | →S28 | AS:practical_reasoning | practical_necessary_condition |
| S30 | Sufficient-Condition Schema | 96 | Part3 p.28 | →S28 | AS:practical_reasoning | practical_sufficient_condition |
| S31 | two-person practical reasoning | 97 | Part3 p.29 | 5 | (AS:practical_reasoning) | two_person_practical |
| S32 | negative form of argument from knowledge | 98 | Part3 p.30 | – | – | – |
| S33 | general form of ad ignorantiam (argument from ignorance) | 98 | Part3 p.30 | – | – | ignorance |
| S34 | epistemic argument from ignorance | 99 | Part3 p.31 | – | – | epistemic_ignorence |
| S35 | negative practical reasoning (practical subspecies of ad ignorantiam) | 100 | Part3 p.32 | – | – | negative_practical |
| S36 | Argument from Positive Consequences | 101, 102 | Part3 p.33, Part3 p.34 | 3 | AS:consequences | positive_consequences |
| S37 | Argument from Negative Consequences | 101, 102 | Part3 p.33, Part3 p.34 | →S36 | (AS:consequences) | negative_consequences |
| S38 | Argument from Negative Consequences (Prudential Inference) | 101, 104 | Part3 p.33, Part3 p.36 | →S36 | (AS:consequences) | negative_consequences_prudential |
| S39 | scheme for danger appeal | 103 | Part3 p.35 | – | – | danger |
| S40 | scheme for fear appeal | 104 | Part3 p.36 | – | – | fear |
| S41 | argument from fear appeal (threat structure) | 105 | Part4 p.1 | – | – | treat |
| S42 | disjunctive form of practical reasoning (indirect threat) | 105 | Part4 p.1 | – | – | pragmatic_alternative |
| S43 | disjunctive ad baculum threat | 106 | Part4 p.2 | – | – | ad_baculum |
| S44 | disjunctive syllogism | 106 | Part4 p.2 | – | – | alternative_cognitive |
| S45 | Positive Form of Argument from Opposites | 107 | Part4 p.3 | – | – | rhetorical_oppositions_descriptive |
| S46 | Negative Form of Argument from Opposites | 107 | Part4 p.3 | – | – | rhetorical_oppositions_descriptive |
| S47 | argument from need for help | 109 | Part4 p.5 | 3 | – | help |
| S48 | argument from distress | 110 | Part4 p.6 | 4 | – | distress |
| S49 | argument from plea for excuse | 111 | Part4 p.7 | 3 | – | plea_for_excuse |
| S50 | argument from composition | 113 | Part4 p.9 | 1 | – | composition_generic |
| S51 | argument from division | 113, 114 | Part4 p.9, Part4 p.10 | 1 | – | division_generic |
| S52 | Scheme for Argument from Gradualism | 114 | Part4 p.10 | – | – | gradualism |
| S53 | Scheme for the Sorites Slippery Slope Argument | 115 | Part4 p.11 | – | – | sorites_slippery_slope |
| S54 | verbal classification slippery slope argument | 115, 116 | Part4 p.11, Part4 p.12 | – | – | verbal_slippery_slope |
| S55 | Scheme for Argument from Precedent Slippery Slope | 116, 117 | Part4 p.12, Part4 p.13 | 3 | – | – |
| S56 | Argument from Vagueness of a Verbal Classification | 118, 119 | Part4 p.14, Part4 p.15 | 3 | – | verbal_classification_vagueness |
| S57 | Scheme for Argument from Arbitrariness of a Verbal Classification | 120 | Part4 p.16 | 3 | – | verbal_classification_arbitrariness |
| S58 | Seneca's form of argument from popular opinion | 123 | Part4 p.19 | – | – | pop_scheme |
| S59 | pop scheme (S1) | 123, 124 | Part4 p.19, Part4 p.20 | 3 | – | pop_scheme |
| S60 | pop scheme (S2) | 123 | Part4 p.19 | →S59 | – | pop_scheme |
| S61 | ad populum scheme (pop scheme) as a presumptive form of argument | 124, 125 | Part4 p.20, Part4 p.21 | →S59 | – | popular_opinion |
| S62 | 2.1. Position-to-Know Ad Populum Argument | 126 | Part4 p.22 | – | – | position_to_know_ad_populum |
| S63 | 2.2. Expert Opinion Ad Populum Argument | 126 | Part4 p.22 | – | – | expert_opinion_ad_populum |
| S64 | 2.3. Deliberation Ad Populum Argument | 127 | Part4 p.23 | – | – | deliberation_ad_populum |
| S65 | 2.4. Moral Justification (Excuse Subtype) Ad Populum Argument | 127, 128 | Part4 p.23, Part4 p.24 | – | – | moral_justification_ad_populum_excuse |
| S66 | 2.5. Moral Justification Ad Populum Argument | 128 | Part4 p.24 | – | – | moral_justification_ad_populum |
| S67 | 2.6. Common Folks Ad Populum Argument | 129 | Part4 p.25 | – | – | common_folks_group_ad_populum, common_folk_ad_populum |
| S68 | 2.7. Rhetoric of Belonging Ad Populum Argument | 130 | Part4 p.26 | – | – | rhetoric_ad_populum |
| S69 | 2.8. Snob Appeal Ad Populum Argument | 130 | Part4 p.26 | – | – | snob_appeal_ad_populum |
| S70 | 2.9. Appeal to Vanity Ad Populum Argument | 131 | Part4 p.27 | – | – | appeal_to_vanity_ad_populum |
| S71 | argument from popular practice (scheme from popular practice) | 131 | Part4 p.27 | – | – | popular_practise |
| S72 | argument from commitment | 134, 135 | Part4 p.30, Part4 p.31 | 3 | – | commitment1, commitment2 |
| S73 | argument from sunk costs | 135, 136 | Part4 p.31, Part4 p.32 | – | – | sunk_costs, waste |
| S74 | argument from inconsistent commitment | 136, 137 | Part4 p.32, Part4 p.33 | 6 | – | inconsistent_commitment |
| S75 | argument from pragmatic inconsistency | 138 | Part4 p.34 | – | – | pragmatic_inconsistency |

In the WAG column, parentheses mean a related WAG scheme, not a match. '→S03' means the entry uses another entry's CQ list.

## 2. Schemes in detail

### Chapter 1. Basic Tools in the State of the Art

#### S01 · Argument from Sign (p. 10)

*sec. 1. Introducing Argumentation Schemes · PDF Part1 p.20 · source: taken from Walton (1996, p. 49)*

- **Minor Premise**: Given data represented as statement A is true in this situation.
- **Major (Generalization) Premise**: Statement B is generally indicated as true when its sign, A, is true, in this kind of situation.
- **Conclusion** ⇒ Therefore, B is true in this situation.
- Variables: A, B

Critical questions: none given in the excerpt.

Notes:
- No critical questions are given for this scheme in the excerpt (the compendium version is in ch. 9, p. 329, outside the excerpt).
- p. 10: the major premise is a defeasible generalization; 'The tracks could have been planted on the trail by tricksters'; argument from sign 'is closely related to abductive inference, or inference to the best explanation'; 'There could be other explanations'.

Mapping: WAG **AS:sign** (comparison in sec. 4); walton.yml: `sign`.
walton.yml note: walton.yml sign (source Walton 2006, FCA pp. 113-114): finding_true(A), indicated_true(B,A); assumption strong_correlation(A,B); 0-ary exception more_reliable.

#### S02 · Argument from Position to Know (p. 13)

*sec. 2. Argument from Position to Know and Expert Opinion · PDF Part1 p.23 · source: given by Walton (1996, p. 61)*

- **Major Premise**: Source a is in a position to know about things in a certain subject domain S containing proposition A.
- **Minor Premise**: a asserts that A (in domain S) is true (false).
- **Conclusion** ⇒ A is true (false).
- Variables: a, S, A

Critical questions ('as indicated by Walton (1996)').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Is a in a position to know whether A is true (false)? (p. 13) | premise | Restates the major premise (the proponent must show it). The book says this CQ 'is directed against the major premise: if the source is not in a position to know, the whole argument collapses' (p. 89-90). |
| CQ2 | Is a an honest (trustworthy, reliable) source? (p. 13) | exception (alt. assumption) | Whoever alleges dishonesty must back it up. The book says this of the parallel trustworthiness question for expert opinion (p. 33: without evidence the question 'is merely an innuendo'). |
| CQ3 | Did a assert that A is true (false)? (p. 13) | premise | Restates the minor premise (that a asserted A). |

What the book says about these CQs:
- CQ1: p. 17: CQ1 and P1 'are closely related and might be characterized as a single premise'.
- CQ2: p. 17: 'for an argument from position to know to be successful, it is necessary for an audience to accept that a is honest'; CQ2 'is clearly distinct and forms a second premise'. p. 90: the assertion 'may be unreliable because the source is dishonest'.
- CQ3: p. 17: P2 and CQ3 are similar, but CQ3 is 'aiming for the very words a spoke, as opposed to a paraphrasing or interpretation'. p. 90: 'his words have been distorted'.

Mapping: WAG **AS:position_to_know** (comparison in sec. 4); walton.yml: `position_to_know`.
walton.yml note: walton.yml position_to_know (source Walton 2002, p. 46) has the same formulas as WAG core.yml: position_to_know(W,D), asserts(W,S), in_domain(S,D); exception dishonest(W).

#### S03 · Argument from Expert Opinion (p. 14, 15)

*sec. 2. Argument from Position to Know and Expert Opinion · PDF Part1 p.24, Part1 p.25 · source: the analysis given by Walton (1997, p. 210)*

- **Major Premise**: Source E is an expert in subject domain S containing proposition A.
- **Minor Premise**: E asserts that proposition A (in domain S) is true (false).
- **Conclusion** ⇒ A may plausibly be taken to be true (false).
- Variables: E, S, A

Critical questions ('the six basic critical questions matching the appeal to expert opinion, as indicated by Walton (1997, p. 223)' (p. 15); listed again pp. 32-33 (with the variants noted); subquestions pp. 31 and 92-93).

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| 1. Expertise Question | How credible is E as an expert source? (p. 15) | premise | Proponent's burden: 'you assume, as part of the argument, that the source is credible' (p. 33). Covered by the major premise (Expertise Premise in version IV). Variant: p. 32: 'How credible (knowledgeable) is E as an expert source?' |
| 2. Field Question | Is E an expert in the field that A is in? (p. 15) | premise | Proponent's burden (p. 33: 'You assume that the expert is an expert in the field of the claim made'). Covered by the major premise ('domain S containing proposition A'); Field Premise in version IV. |
| 3. Opinion Question | What did E assert that implies A? (p. 15) | premise | Proponent's burden (p. 33: the expert 'has said something ... from which the claim can be extracted'). Covered by the minor premise; Opinion Premise in version IV. |
| 4. Trustworthiness Question | Is E personally reliable as a source? (p. 15) | exception | Questioner's burden: 'to mount such an attack, the critic has to produce some fairly substantial evidence. Otherwise, the question is merely an innuendo' (p. 33). Variant: p. 33: 'Is E personally reliable as a source – for example, is E biased?' |
| 5. Consistency Question | Is A consistent with what other experts assert? (p. 15) | exception (alt. conclusion) | Questioner's burden: 'that needs to be shown by telling us what the other experts have in fact said' (p. 33). |
| 6. Backup Evidence Question | Is E's assertion based on evidence? (p. 15) | assumption (alt. argument) | Proponent's burden: 'You assume that the expert's assertion was based on some evidence within the field of his or her expertise' (p. 33). Not among the premises of versions I/II (Backup Evidence Premise in IV). |

What the book says about these CQs:
- 1. Expertise Question: p. 33: CQs 1, 2, 3 and 6 'can be seen as implicit premises' (merely asking makes the argument default; the burden stays on the proponent); 4 and 5 'seem to have a positive burden of proof attached to the side of the questioner'.
- 2. Field Question: p. 27: in the Bradshaw example the field CQ and the domain variable S diagnose the same fault twice ('a kind of duplication').
- 4. Trustworthiness Question: Subquestions (Walton, 1997, p. 217), p. 31: 'Subquestion 1: Is E biased? / Subquestion 2: Is E honest? / Subquestion 3: Is E conscientious?'
- 5. Consistency Question: p. 33: 'then that is an argument against the claim. It is a rebuttal, especially if what the other experts say represents the generally accepted opinion in the field.' See also p. 92 under CQ 6.
- 6. Backup Evidence Question: p. 92: 'The last two questions [5 and 6] are directed against the plausibility of the link between premises and conclusion ... Both these attacks weaken the acceptability of the conclusion, but they do not compromise the argument's validity.'

Notes:
- p. 19 calls this 'version I'.
- p. 14: 'Argument from expert opinion is a subtype of the more general argumentation scheme for argument from position to know'.

Mapping: WAG **AS:expert_opinion** (comparison in sec. 4); walton.yml: `expert_opinion`, `credible_source`.
walton.yml note: walton.yml expert_opinion (source Walton 1997, pp. 211-225): expert(W,D), in_domain(S,D), asserts(W,S); exceptions untrustworthy(W), inconsistent_with_other_experts(S); assumption based_on_evidence(asserts(W,S)). Its loci match the book's p. 33 burdens one to one. WAG's AS:expert_opinion instead uses credible_source (Wyner 2016).

#### S04 · argument from position to know (as recalled on p. 17) (p. 17)

*sec. 3. Critical Questions · PDF Part1 p.27 · source: (recalls p. 13) · name from text: 'recall the scheme capturing argument from position to know introduced earlier'*

- **(P1)**: a is in a position to know whether A is true (false).
- **(P2)**: a asserts that A is true (false).
- **(C)** ⇒ Therefore, A is true (false).
- Variables: a, A

Critical questions. Same questions as S02.

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| (CQ1) | Is a in a position to know whether A is true (false)? (p. 17) | premise | Restates the major premise (the proponent must show it). The book says this CQ 'is directed against the major premise: if the source is not in a position to know, the whole argument collapses' (p. 89-90). |
| (CQ2) | Is a an honest (trustworthy, reliable) source? (p. 17) | assumption (alt. exception) | The book's p. 17 reading: honesty is a necessary assumption the audience must accept 'in addition to the premises P1 and P2', so a linked premise the proponent carries. |
| (CQ3) | Did a assert that A is true (false)? (p. 17) | premise | Restates the minor premise (that a asserted A). |

What the book says about these CQs:
- (CQ1): p. 17: CQ1 and P1 'are closely related and might be characterized as a single premise'.
- (CQ2): p. 17: 'for an argument from position to know to be successful, it is necessary for an audience to accept that a is honest'; CQ2 'is clearly distinct and forms a second premise'. p. 90: the assertion 'may be unreliable because the source is dishonest'.
- (CQ3): p. 17: P2 and CQ3 are similar, but CQ3 is 'aiming for the very words a spoke, as opposed to a paraphrasing or interpretation'. p. 90: 'his words have been distorted'.

Notes:
- Wording differs from p. 13: P1 uses the CQ1 wording ('whether A is true (false)') and drops the domain S.
- p. 17: P1 'is left implicit (and ... is probably assumed by the hearer by Grice's Principle of Charity)'; 'The complete set of linked premises employed in a scheme is thus the union of those given as premises and (the propositional content of) those listed as critical questions.'

Mapping: WAG **AS:position_to_know**; walton.yml: `position_to_know`.

#### S05 · Appeal to Expert Opinion (Version II) (p. 19)

*sec. 4. Enthymemes, Schemes, and Critical Questions · PDF Part1 p.29 · source: reformulation of Walton (1997, p. 210)*

- **Major Premise**: Source E is an expert in subject domain S containing proposition A.
- **Minor Premise**: E asserts that proposition A (in domain S) is true (false).
- **Conditional Premise**: If source E is an expert in a subject domain S containing proposition A, and E asserts that proposition A is true (false), then A may plausibly be taken to be true (false).
- **Conclusion** ⇒ A may plausibly be taken to be true (false).
- Variables: E, S, A

Critical questions: see S03.

Notes:
- The conditional premise 'expresses the Toulmin warrant'; the form is 'defeasible modus ponens' (p. 19).
- Used 'with accompanying critical questions' (p. 20), i.e. the six CQs of S03. p. 21: version II is the approach adopted in Araucaria.

Mapping: WAG **AS:expert_opinion**; walton.yml: `expert_opinion`.

#### S06 · Appeal to Expert Opinion (Version III) (p. 19, 20)

*sec. 4. Enthymemes, Schemes, and Critical Questions · PDF Part1 p.29, Part1 p.30*

- **Major Premise**: Source E is an expert in subject domain S containing proposition A.
- **Minor Premise**: E asserts that proposition A (in domain S) is true (false).
- **Conditional Premise**: If source E is an expert in a subject domain S containing proposition A, and E asserts that proposition A is true (false), and E is credible as an expert source, and E is an expert in the field A is in, and E asserted A, or a statement that implies A, and E is personally reliable as a source, and A is consistent with what other experts assert, and E's assertion is based on evidence, then A may plausibly be taken to be true (false).
- **Conclusion** ⇒ A may plausibly be taken to be true (false).
- Variables: E, S, A

Critical questions: see S03.

Notes:
- The CQs are built into the conditional premise (p. 19: 'building the critical questions into the argumentation scheme'). p. 20: it 'makes the conditional premise seem cumbersome and hard to remember'.

Mapping: WAG **AS:expert_opinion**; walton.yml: `expert_opinion`.

#### S07 · Appeal to Expert Opinion (Version IV) (p. 20)

*sec. 4. Enthymemes, Schemes, and Critical Questions · PDF Part1 p.30*

- **Major Premise**: Source E is an expert in subject domain S containing proposition A.
- **Minor Premise**: E asserts that proposition A (in domain S) is true (false).
- **Conditional Premise**: If source E is an expert in a subject domain S containing proposition A, and E asserts that proposition A is true (false), then A may plausibly be taken to be true (false).
- **Expertise Premise**: E is credible as an expert source.
- **Field Premise**: E is an expert in the field that A is in.
- **Opinion Premise**: E asserted A, or made a statement that implies A.
- **Trustworthiness Premise**: E is personally reliable as a source.
- **Consistency Premise**: A is consistent with what other experts assert.
- **Backup Evidence Premise**: E's assertion is based on evidence.
- **Conclusion** ⇒ A may plausibly be taken to be true (false).
- Variables: E, S, A

Critical questions: see S03.

Notes:
- p. 20: 'In version IV, all the critical questions have been built in as premises. Now the argumentation scheme is complete by itself, and we don't need the device of critical questions any longer, or so it would seem.'
- Useful for WAG: each premise is the statement behind one CQ (Expertise, Field, Opinion, Trustworthiness, Consistency, Backup Evidence).

Mapping: WAG **AS:expert_opinion**; walton.yml: `expert_opinion`.

### Chapter 2. Argumentation Schemes for Arguments from Analogy, Classification, and Precedent

#### S08 · argument from analogy: Beardsley's formalization (p. 49)

*sec. 2. Argument from Analogy as Treated in Logic Textbooks · PDF Part2 p.25 · source: Beardsley, 1950, p. 107 · kind: quoted from the literature · name from text: 'His attack on argument from analogy derives from the following formalization of the argumentation scheme (ibid.)'*

- **—**: X has certain characteristics a, b, c . . .
- **—**: Y has the characteristics a, b, c . . .
- **—**: But Y also has the characteristics x, y, z . . .
- **—** ⇒ Therefore, X has the characteristics x, y, z.
- Variables: X, Y, a, b, c, x, y, z

Critical questions: none given in the excerpt.

Notes:
- Beardsley holds arguments of this form fallacious (p. 49).

Mapping: no WAG scheme (related: AS:analogy); walton.yml: none.

#### S09 · argument from analogy: Burbridge's scheme (Primary Subject, Analogue, Targeted Predicate) (p. 50)

*sec. 2. Argument from Analogy as Treated in Logic Textbooks · PDF Part2 p.26 · source: Burbridge, Within Reason (1990), pp. 12, 13 · kind: quoted from the literature · name from table column headed 'Scheme' (p. 50)*

- **Premise 1**: PS is like A in S¹ . . . Sⁿ
- **Premise 2**: A has TP.
- **Conclusion** ⇒ So, PS has TP.
- Variables: PS (Primary Subject), A (Analogue), TP (Targeted Predicate), S (Similarities)

Critical questions: none given in the excerpt.

Example: The universe and a machine are similar in that both are divided into an intricate pattern of parts and subparts. / A machine has a maker. / Therefore, the universe has a maker.

Notes:
- Premise 1 has no final full stop in the print.

Mapping: no WAG scheme (related: AS:analogy); walton.yml: none.

Uncertain readings: Premise 1 ends in 'S' . . . S"' in the scan: the marks after S are read as superscripts 1 and n (S¹ . . . Sⁿ), not certain.

#### S10 · Burbridge's objection to an argument from analogy (p. 51)

*sec. 2. Argument from Analogy as Treated in Logic Textbooks · PDF Part2 p.27 · source: Burbridge (1990), p. 31 · kind: quoted from the literature; a counter-scheme (objection) · name from text: 'The objection that can be raised, showing the unacceptability of the reasoning from analogy, takes the following form (p. 31)'*

- **Premise 1**: The Objector has assumed that because PS is the same as A, it shares the targeted predicate TP.
- **Premise 2**: PS and A are not identical but analogues, different though similar.
- **Premise 3**: The difference between PS and A is more closely connected to TP than the similarity is.
- **Premise 4** ⇒ Therefore, TP does not apply to PS.
- Variables: PS, A, TP

Critical questions: none given in the excerpt.

Notes:
- The conclusion is printed with the label 'Premise 4'.
- p. 51: 'Burbridge's evaluation of argument from analogy by the device of raising objections is comparable to the method of critical questions outlined earlier.' It matches the book's CQ3 on differences (S13).

Mapping: no WAG scheme (related: AS:analogy); walton.yml: none.

#### S11 · deductive argument by analogy (Waller; 'deduction from principle') (p. 53, 54)

*sec. 3. Is Argument from Analogy Deductive or Inductive? · PDF Part2 p.29, Part2 p.30 · source: Waller, 2001, p. 201 · kind: quoted from the literature · name from text: 'According to Waller (2001, p. 201), deductive arguments by analogy have the following form'*

- **—**: We both agree with case a.
- **—**: The most plausible reason for believing a is the acceptance of principle C.
- **—**: C implies b (b is a case that fits under principle C).
- **—** ⇒ Therefore, consistency requires the acceptance of b.
- Variables: a, b, C

Critical questions: none given in the excerpt.

Notes:
- p. 54: Waller classifies it as 'a deduction from principle'. The scheme itself is printed on p. 54; the introducing sentence is on p. 53.

Mapping: no WAG scheme (related: AS:analogy); walton.yml: none.

#### S12 · inductive argument from analogy (Waller) (p. 54)

*sec. 3. Is Argument from Analogy Deductive or Inductive? · PDF Part2 p.30 · source: Waller, 2001, p. 202 · kind: quoted from the literature · name from text: 'Waller offers (ibid.) the following structure to represent the form of inductive arguments from analogy'*

- **—**: D has characteristics e, f, g, and h.
- **—**: E also has characteristics e, f, g, and h.
- **—**: D also has characteristic k.
- **—**: Having characteristics e, f, g, and h is relevant to having characteristic k.
- **—** ⇒ Therefore, E will probably also have characteristic k.
- Variables: D, E, e, f, g, h, k

Critical questions: none given in the excerpt.

Notes:
- The relevance premise anticipates the book's 'Relevant Similarity Premise' of version 2 (S16).

Mapping: no WAG scheme (related: AS:analogy); walton.yml: none.

#### S13 · Argument from Analogy: Version 1 (p. 56, 62)

*sec. 4. The Schemes for Argument from Analogy (CQs in 5. Argument from Analogy as a Defeasible Form of Argument) · PDF Part2 p.32, Part2 p.38 · source: Walton, 1989, p. 258*

- **Major Premise**: Generally, case C1 is similar to case C2.
- **Minor Premise**: Proposition A is true (false) in case C1.
- **Conclusion** ⇒ Proposition A is true (false) in case C2.
- Variables: C1 (source case), C2 (target case), A

Critical questions ('The following critical questions matching version one of the scheme for argument from analogy are given by Walton (1989, p. 258)' (p. 62); heading 'Critical Questions for Argument from Analogy').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Is A true (false) in C1? (p. 62) | premise | Restates the minor premise. |
| CQ2 | Are C1 and C2 similar, in the respects cited? (p. 62) | premise | Restates the major premise. |
| CQ3 | Are there important differences (dissimilarities) between C1 and C2? (p. 62) | exception (alt. argument) | The questioner must point to the differences. |
| CQ4 | Is there some other case C3 that is also similar to C1 except that A is false (true) in C3? (p. 62) | conclusion (alt. exception) | Points to a counteranalogy, an argument for the opposite conclusion (a rebuttal). |

What the book says about these CQs:
- CQ1: p. 63: 'CQ1 and CQ2 are redundant' (they restate premises; cf. Verheij's first role, p. 62).
- CQ2: p. 63: redundant. p. 65: the factors of the wild-animals example 'correspond to critical questions CQ2 and CQ3'.
- CQ3: p. 63: 'CQ3 could be either an undercutter or a defeater, perhaps depending on how important the differences are.'
- CQ4: p. 63: 'CQ4 looks more like a defeater ... The strategy for attacking an argument from analogy suggested by CQ4 is to produce a counteranalogy ... This attack would seem to be a defeater.'

Example: I infer that a new pair of shoes will wear well on the grounds that I got good wear from other shoes previously purchased from the same store. (Walton, 1989, p. 258)

Notes:
- p. 86: 'Our conclusion is to recommend version one of the argumentation scheme for argument from analogy, reserving application of version two to cases where relevance needs to be decided.'

Mapping: WAG **AS:analogy** (comparison in sec. 4); walton.yml: `analogy`.
walton.yml note: walton.yml analogy (source Walton 2006, FCA pp. 96-97) has the same formulas as WAG: similar_case(C), in_case(A,C); exceptions relevant_differences(C), more_on_point(A,C). Only one case variable; the target case is implicit.

#### S14 · Guarini's core scheme for argument from analogy (p. 57)

*sec. 4. The Schemes for Argument from Analogy · PDF Part2 p.33 · source: Guarini, 2004, p. 161 · kind: quoted from the literature · name from text: 'The first one, which he calls the core scheme, has the following form (Guarini, 2004, p. 161)'*

- **—**: a has features f1, f2, . . . , fn.
- **—**: b has features f1, f2, . . . , fn.
- **—** ⇒ a and b should be treated or classified in the same way with respect to f1, f2, . . . , fn.
- Variables: a, b, f1 ... fn

Critical questions: none given in the excerpt.

Notes:
- No 'Therefore' is printed; the third line is taken as the conclusion.
- p. 57: Guarini classifies the DeGraffenreid v. General Motors case as 'deductive argument from analogy that depends on a principle but does not require relevance'.

Mapping: no WAG scheme (related: AS:analogy); walton.yml: none.

#### S15 · Guarini's second (extended) scheme for argument from analogy (p. 58)

*sec. 4. The Schemes for Argument from Analogy · PDF Part2 p.34 · source: Guarini, 2004, p. 162 · kind: quoted from the literature · name from text: 'He postulates a second scheme for argument from analogy by adding two further steps to the core scheme (2004, p. 162)'*

- **—**: a has features f1, f2, . . . , fn.
- **—**: b has features f1, f2, . . . , fn.
- **—**: a and b should be treated or classified in the same way with respect to f1, f2, . . . , fn.
- **—**: a is X in virtue of f1, f2, . . . , fn.
- **—** ⇒ Therefore, b is X.
- Variables: a, b, X, f1 ... fn

Critical questions: none given in the excerpt.

Notes:
- Only the two added lines are printed on p. 58 ('a is X in virtue of f1, f2, . . . , fn.' / 'Therefore, b is X.'); the first three lines are the core scheme (S14), and the book says 'all five steps' make up the scheme.
- p. 58: Guarini takes f1 ... fn to be 'relevant similarities' but leaves 'relevant' out of the scheme.

Mapping: no WAG scheme (related: AS:analogy); walton.yml: none.

#### S16 · Argument from Analogy: Version 2 (p. 58)

*sec. 4. The Schemes for Argument from Analogy · PDF Part2 p.34 · source: the book's own proposal (version one plus a relevance premise)*

- **Major Premise**: Generally, case C1 is similar to case C2.
- **Relevant Similarity Premise**: The similarity between C1 and C2 observed so far is relevant to the further similarity that is in question.
- **Minor Premise**: Proposition A is true (false) in case C1.
- **Conclusion** ⇒ Proposition A is true (false) in case C2.
- Variables: C1, C2, A

Critical questions: see S13 (given for version one only).

Notes:
- No CQ list is given for version 2; the p. 62 list is explicitly for version one.

Mapping: WAG **AS:analogy**; walton.yml: `analogy`.

#### S17 · Deductive Form of Argument from Classification (p. 66)

*sec. 6. Arguments from Classification · PDF Part2 p.42 · kind: deductive*

- **—**: All F's can be classified as G's.
- **—**: a is an F.
- **—** ⇒ Therefore, a is a G.
- Variables: a (a thing or individual), F, G (classes of things)

Critical questions: none given in the excerpt.

Notes:
- p. 66: 'This form of argument is deductively valid'.

Mapping: no WAG scheme; walton.yml: `verbal_classification`.
walton.yml note: walton.yml verbal_classification: instance(O,F), subclass(F,G) -> instance(O,G) has this deductive shape, with no exceptions.

#### S18 · Scheme for Defeasible Argument from a Verbal Classification (p. 68)

*sec. 6. Arguments from Classification · PDF Part2 p.44 · source: Walton, 1996, p. 54*

- **Major Premise**: If some particular thing a can be classified as falling under verbal category C, then a has property F (in virtue of such a classification).
- **Minor Premise**: a can be classified as falling under verbal category C.
- **Conclusion** ⇒ a has property F.
- Variables: a, C, F

Critical questions.

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Does a definitely have F, or is there room for doubt? (p. 68) | argument (alt. conclusion) | Raises doubt about whether the generalization carries over to this case, without a statement behind it (an undercutter). |
| CQ2 | Can the verbal classification (in the second premise) be said to hold strongly, or is it one of those weak classifications that is subject to doubt? (p. 68) | premise | Explicitly targets the second (minor) premise. |

What the book says about these CQs:
- CQ1: p. 68: 'This can be done by attacking the generalization or by questioning whether it fits the particular thing at issue. The two critical questions focus on the application of the generalization to the concrete case, and on its strength.'

Mapping: no WAG scheme; walton.yml: `verbal_classification`.
walton.yml note: walton.yml verbal_classification is the deductive class-membership form (no defeasibility, no exceptions); definition_to_verbal_classification (Walton 2006 FCA p. 129) is a different ch. 9 scheme.

#### S19 · Argument from Analogy Based on Classification (p. 70)

*sec. 6. Arguments from Classification · PDF Part3 p.2 · source: specified by Govier (1992, p. 274) · kind: quoted from the literature*

- **—**: The analogue has features a, b, and c.
- **—**: The primary subject has features a, b, and c.
- **—**: It is by virtue of features a, b, and c that the analogue is properly classified as W.
- **—** ⇒ So, the primary subject ought to be classified as W.
- Variables: a, b, c, W

Critical questions: none given in the excerpt.

Notes:
- p. 69: 'a special argumentation scheme that combines argument from analogy with argument from classification'.

Mapping: no WAG scheme (related: AS:analogy); walton.yml: none.

#### S20 · Scheme for Argument from an Established Rule (p. 71)

*sec. 7. Arguments Based on Rules and Classifications · PDF Part3 p.3 · source: Walton, 1996, p. 92*

- **Major Premise**: If carrying out types of actions including the state of affairs A is the established rule for x, then (unless the case is an exception), x must carry out A.
- **Minor Premise**: Carrying out types of actions including the state of affairs A is the established rule for a.
- **Conclusion** ⇒ Therefore, a must carry out A.
- Variables: A, x, a

Critical questions (heading 'Critical Questions for Argument from an Established Rule').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Does the rule require carrying out types of actions that include A as an instance? (p. 71) | premise | Questions what the rule requires, i.e. the minor premise and the major premise's antecedent (proponent's burden). |
| CQ2 | Are there other established rules that might conflict with, or override, this one? (p. 71) | exception (alt. conclusion) | The questioner must cite the conflicting rule. |
| CQ3 | Is this case an exceptional one, that is, could there be extenuating circumstances or an excuse for noncompliance? (p. 71) | exception (alt. conclusion) | The major premise itself reserves '(unless the case is an exception)'; the questioner must cite the excuse (cf. Argument from an Exceptional Case, S21). |

What the book says about these CQs:
- CQ3: p. 71: 'The last critical question concerns the problem of the exceptional case ... The counterargument is a kind of rebuttal of the implicit assumption that the case in question represents an instance of the rule.'

Notes:
- p. 71: 'The critical questions are directed against the generalization in the rule.'
- It 'belongs to a different kind of dialogue, that of deliberation' (p. 71).

Mapping: no WAG scheme; walton.yml: `established_rule`.
walton.yml note: walton.yml established_rule (source Walton 1995, p. 147): has_conclusion(R,C), applicable(R); assumption valid(R). Generic rule application, no exception formula; regulative_rule (Grennan) has a 0-ary exception excuse.

#### S21 · Scheme for Argument from an Exceptional Case (p. 72)

*sec. 7. Arguments Based on Rules and Classifications · PDF Part3 p.4 · source: Walton, 1996, p. 92*

- **Major Premise**: If the case of x is an exception, then the established rule can be waived in the case of x.
- **Minor Premise**: The case of a is an exception.
- **Conclusion** ⇒ Therefore, the established rule can be waived in the case of a.
- Variables: x, a

Critical questions (heading 'Critical Questions for Argument from an Exceptional Case').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Is the case of a a recognized type of exception? (p. 72) | premise | Questions the minor premise. |
| CQ2 | If it is not a recognized case, can evidence why the established rule does not apply to it be given? (p. 72) | premise | Asks the proponent for evidence for the minor premise when the exception is not a recognized type. |
| CQ3 | If it is a borderline case, can comparable cases be cited? (p. 72) | assumption | Asks the proponent to support a borderline exception with comparable cases (argument from analogy). |

What the book says about these CQs:
- CQ3: p. 72: 'The last critical question indicates how argument from an exceptional case and its rebuttals are based on argument from analogy'.

Notes:
- p. 72: 'The critical questions stem from the classification.'

Mapping: no WAG scheme; walton.yml: `exeptional`.
walton.yml note: walton.yml exeptional (source Walton 1995, p. 148): rule_can_waived(X), case_exception(A) -> rule_waived(A); case_exception(A) is both premise and assumption (audit defect).

#### S22 · Refutation Scheme for Argument from Precedent (p. 73, 74)

*sec. 8. Argument from Precedent and Practical Argument from Analogy · PDF Part3 p.5, Part3 p.6 · source: Walton, 1996, p. 94*

- **Major Premise**: Generally, according to the established rule, if x has property F, then x also has property G.
- **Minor Premise**: In this legitimate case, a has property F but does not have property G.
- **Conclusion** ⇒ Therefore, an exception to the rule must be recognized, and the rule must be appropriately modified or qualified.
- Variables: x, a, F, G

Critical questions (heading 'Critical Questions for Argument from Precedent').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Does the established rule really apply to this case? (p. 73) | premise | Questions the minor premise (that the cited case falls under the rule). |
| CQ2 | Is the case cited legitimate, or can it be explained as only an apparent violation of the rule? (p. 73) | premise (alt. exception) | Questions 'In this legitimate case' in the minor premise. |
| CQ3 | Can the case cited be dealt with under an already recognized category of exception that does not require a change in the rule? (p. 73) | exception (alt. argument) | The defender of the rule cites an existing exception category; the premises stand but the conclusion (modify the rule) is blocked. |

Notes:
- p. 73: 'It leads to a conclusion that the rule must be modified.' 'The following critical questions represent strategies to defend the argument from established rule by attacking the opponent's argument.' p. 74: 'These critical questions represent possible moves to defeat an argument from precedent.'

Mapping: no WAG scheme; walton.yml: none.
walton.yml note: No match. walton.yml exeption_to_rule is upstream Carneades 'precedent' (similar_case(C), rule_of_case(R,C), has_conclusion(R,S); exceptions relevant_differences(R,C), inapplicable_rule(R)): case-based rule transfer, not this refutation scheme (rule must be modified).

#### S23 · Positive Scheme for Practical Argument from Analogy (p. 74)

*sec. 8. Argument from Precedent and Practical Argument from Analogy · PDF Part3 p.6 · source: Walton, 1989, p. 257*

- **—**: The right thing to do in S1 was to carry out A.
- **—**: S2 is similar to S1.
- **—** ⇒ Therefore, the right thing to do in S2 is carry out A.
- Variables: S1, S2, A

Critical questions: none given in the excerpt.

Notes:
- No CQs in the excerpt (ch. 9 p. 316 has the compendium version).

Mapping: no WAG scheme (related: AS:analogy); walton.yml: `practical_analogy_positive`.

#### S24 · Negative Scheme for Practical Argument from Analogy (p. 74)

*sec. 8. Argument from Precedent and Practical Argument from Analogy · PDF Part3 p.6 · source: Walton, 1989, p. 257*

- **—**: The wrong thing to do in S1 was to carry out A.
- **—**: S2 is similar to S1.
- **—** ⇒ Therefore, the wrong thing to do in S2 is carry out A.
- Variables: S1, S2, A

Critical questions: none given in the excerpt.

Mapping: no WAG scheme (related: AS:analogy); walton.yml: `practical_analogy_negative`.

### Chapter 3. Knowledge-Related, Practical, and Other Schemes

#### S25 · argument from position to know (Walton, 2002) (p. 89, 90)

*sec. 1. Arguments from Knowledge · PDF Part3 p.21, Part3 p.22 · source: Walton, 2002, p. 46 · name from text: 'The argument from position to know takes the following form (Walton, 2002, p. 46)'*

- **Major Premise**: Source a is in a position to know about things in a certain subject domain S containing proposition A.
- **Minor Premise**: a asserts that A is true (false).
- **Conclusion** ⇒ A is true (false).
- Variables: a, S, A

Critical questions. Same questions as S02.

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Is a in a position to know whether A is true (false)? (p. 90) | premise | Restates the major premise (the proponent must show it). The book says this CQ 'is directed against the major premise: if the source is not in a position to know, the whole argument collapses' (p. 89-90). |
| CQ2 | Is a an honest (trustworthy, reliable) source? (p. 90) | exception (alt. assumption) | Whoever alleges dishonesty must back it up. The book says this of the parallel trustworthiness question for expert opinion (p. 33: without evidence the question 'is merely an innuendo'). |
| CQ3 | Did a assert that A is true (false)? (p. 90) | premise | Restates the minor premise (that a asserted A). |

What the book says about these CQs:
- CQ1: p. 17: CQ1 and P1 'are closely related and might be characterized as a single premise'.
- CQ2: p. 17: 'for an argument from position to know to be successful, it is necessary for an audience to accept that a is honest'; CQ2 'is clearly distinct and forms a second premise'. p. 90: the assertion 'may be unreliable because the source is dishonest'.
- CQ3: p. 17: P2 and CQ3 are similar, but CQ3 is 'aiming for the very words a spoke, as opposed to a paraphrasing or interpretation'. p. 90: 'his words have been distorted'.

Notes:
- Differs from p. 13 only in the minor premise ('a asserts that A is true (false).' without '(in domain S)').
- p. 89-90: 'The first critical question is directed against the major premise: if the source is not in a position to know, the whole argument collapses. The remaining two critical questions focus on the problem of the assertion: it may be unreliable because the source is dishonest, or because his words have been distorted.'

Mapping: WAG **AS:position_to_know**; walton.yml: `position_to_know`.

#### S26 · argument from witness testimony (p. 90, 91)

*sec. 1. Arguments from Knowledge · PDF Part3 p.22, Part3 p.23 · source: Prakken, Reed, and Walton, 2003, p. 35 · name from text: 'The following scheme represents the general form of the argument from witness testimony (Prakken, Reed, and Walton, 2003, p. 35)'*

- **Position-to-Know Premise**: Witness W is in position to know whether A is true or not.
- **Truth-Telling Premise**: Witness W is telling the truth (as W knows it).
- **Statement Premise**: Witness W states that A is true (false).
- **Conclusion** ⇒ A may be plausibly taken to be true (false).
- Variables: W, A

Critical questions.

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Is what the witness said internally consistent? (p. 91) | assumption (alt. exception) | Checkable from the testimony itself; the proponent offering the testimony can be expected to show it hangs together. |
| CQ2 | Is what the witness said consistent with the known facts of the case (based on evidence apart from what the witness testified to)? (p. 91) | exception | The questioner must cite the conflicting facts. |
| CQ3 | Is what the witness said consistent with what other witnesses have (independently) testified to? (p. 91) | exception | The questioner must cite the conflicting testimony. |
| CQ4 | Is there some kind of bias that can be attributed to the account given by the witness? (p. 91) | exception | Bias has to be shown by evidence (cf. p. 33 on trustworthiness). |
| CQ5 | How plausible is the statement A asserted by the witness? (p. 91) | conclusion (alt. exception) | Evidence that A is implausible is evidence against A itself (a rebuttal). |

Notes:
- p. 90-91: 'The reliability of the source being the most important condition of the scheme, the critical questions are directed against this major premise. Consistency of the statement with the rest of the testimony, with the known facts, and with other testimony, along with source bias, are the most important factors used to evaluate the trustworthiness of the witness and, consequently, the plausibility of his statement.'
- p. 90: 'The truth-telling premise is an important difference from the previous scheme.'

Mapping: no WAG scheme (related: AS:position_to_know); walton.yml: `witness_testimony`.
walton.yml note: walton.yml witness_testimony: position_to_know(W,D), in_domain(D,S) [arguments reversed vs. other schemes], believes(W,S), asserts(W,S); assumption internally_consistent(S) (CQ1); exceptions inconsistent_with_known_facts(S) (CQ2), inconsistent_with_other_witnesses(S) (CQ3), biased(W) (CQ4), implausible(S) (CQ5). Its loci match the drafts here except CQ5 (exception there, conclusion here).

#### S27 · argument from expert opinion (Walton, 2002) (p. 91, 92, 93)

*sec. 1. Arguments from Knowledge · PDF Part3 p.23, Part3 p.24, Part3 p.25 · source: Walton, 2002, p. 50 · name from text: 'This is the form of the argument from expert opinion (Walton, 2002, p. 50)'*

- **Major Premise**: Source E is an expert in subject domain S containing proposition A.
- **Minor Premise**: E asserts that proposition A is true (false).
- **Conclusion** ⇒ A is true (false).
- Variables: E, S, A

Critical questions ('a more extensive and detailed account of the critical questions and subquestions (Walton, 1997, pp. 211-225)' (p. 92)). Same questions as S03.

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| 1. Expertise Question | How credible is E as an expert source? (p. 92) | premise | Proponent's burden: 'you assume, as part of the argument, that the source is credible' (p. 33). Covered by the major premise (Expertise Premise in version IV). Variant: p. 32: 'How credible (knowledgeable) is E as an expert source?' |
| 2. Field Question | Is E an expert in the field that A is in? (p. 92) | premise | Proponent's burden (p. 33: 'You assume that the expert is an expert in the field of the claim made'). Covered by the major premise ('domain S containing proposition A'); Field Premise in version IV. |
| 3. Opinion Question | What did E assert that implies A? (p. 92) | premise | Proponent's burden (p. 33: the expert 'has said something ... from which the claim can be extracted'). Covered by the minor premise; Opinion Premise in version IV. |
| 4. Trustworthiness Question | Is E personally reliable as a source? (p. 92) | exception | Questioner's burden: 'to mount such an attack, the critic has to produce some fairly substantial evidence. Otherwise, the question is merely an innuendo' (p. 33). Variant: p. 33: 'Is E personally reliable as a source – for example, is E biased?' |
| 5. Consistency Question | Is A consistent with what other experts assert? (p. 92) | exception (alt. conclusion) | Questioner's burden: 'that needs to be shown by telling us what the other experts have in fact said' (p. 33). |
| 6. Backup Evidence Question | Is E's assertion based on evidence? (p. 92) | assumption (alt. argument) | Proponent's burden: 'You assume that the expert's assertion was based on some evidence within the field of his or her expertise' (p. 33). Not among the premises of versions I/II (Backup Evidence Premise in IV). |

What the book says about these CQs:
- 1. Expertise Question: p. 33: CQs 1, 2, 3 and 6 'can be seen as implicit premises' (merely asking makes the argument default; the burden stays on the proponent); 4 and 5 'seem to have a positive burden of proof attached to the side of the questioner'.
- 2. Field Question: p. 27: in the Bradshaw example the field CQ and the domain variable S diagnose the same fault twice ('a kind of duplication').
- 4. Trustworthiness Question: Subquestions (Walton, 1997, p. 217), p. 31: 'Subquestion 1: Is E biased? / Subquestion 2: Is E honest? / Subquestion 3: Is E conscientious?'
- 5. Consistency Question: p. 33: 'then that is an argument against the claim. It is a rebuttal, especially if what the other experts say represents the generally accepted opinion in the field.' See also p. 92 under CQ 6.
- 6. Backup Evidence Question: p. 92: 'The last two questions [5 and 6] are directed against the plausibility of the link between premises and conclusion ... Both these attacks weaken the acceptability of the conclusion, but they do not compromise the argument's validity.'

Subquestions (pp. 92–93):
- 1. Expertise Question: **1.1** What is E's name, job or official capacity, location, and employer? **1.2** What degrees, professional qualifications, or certification by licensing agencies does E hold? **1.3** Can testimony of peer experts in the same field be given to support E's competence? **1.4** What is E's record of experience or other indications of practiced skill in S? **1.5** What is E's record of peer-reviewed publications or contributions to knowledge in S?
- 2. Field Question: **2.1** Is the field of expertise cited in the appeal a genuine area of knowledge or an area of technical skill that supports a claim to knowledge? **2.2** If E is an expert in a field closely related to the field cited in the appeal, how close is the relationship between the expertise in the two fields? **2.3** Is the issue one where expert knowledge in any field is directly relevant to deciding the issue? **2.4** Is the field of expertise cited an area where there are changes in techniques or rapid developments in new knowledge? If so, is the expert up to date in these developments?
- 3. Opinion Question: **3.1** Was E quoted as asserting A? Was a reference to the source of the quote given, and can it be verified that E actually said A? **3.2** If E did not say A exactly, then what did E assert, and how was A inferred? **3.3** If the inference to A was based on more than one premise, could one premise have come from E and the other from a different expert? If so, is there evidence of disagreement between what the two experts (separately) asserted? **3.4** Is what E asserted clear? If not, was the process of interpretation of what A said by the respondent who used E's opinion justified? Are other interpretations plausible? Could important qualifications have been left out?
- 4. Trustworthiness Question: **4.1** Is E biased? **4.2** Is E honest? **4.3** Is E conscientious?
- 5. Consistency Question: **5.1** Does A have general acceptance in S? **5.2** If not, can E explain why not, and give reasons why there is good evidence for A?
- 6. Backup Evidence Question: **6.1** What is the internal evidence the expert herself used to arrive at this opinion as her conclusion? **6.2** If there is external evidence (e.g., physical evidence reported independently of the expert), can the expert deal with this adequately? **6.3** Can it be shown that the opinion given is one that is scientifically verifiable?

Notes:
- Differs from S03: minor premise without '(in domain S)', conclusion 'A is true (false).' instead of 'A may plausibly be taken to be true (false).'
- Subquestion 3.4 is printed 'what A said' (sic; E is meant).
- p. 92: 'the trustworthiness attack undermines the argument on the basis of source bias or dishonesty.' p. 92: 'The last two questions [5 and 6] are directed against the plausibility of the link between premises and conclusion ... Both these attacks weaken the acceptability of the conclusion, but they do not compromise the argument's validity.'

Mapping: WAG **AS:expert_opinion**; walton.yml: `expert_opinion`.

#### S28 · argumentation scheme for practical reasoning (Goal Premise / Means Premise) (p. 95, 96)

*sec. 2. Practical Reasoning · PDF Part3 p.27, Part3 p.28 · source: (after Walton 1990; variants Walton 1997, p. 164) · name from text: 'In the argumentation scheme for practical reasoning presented here ...'*

- **Goal Premise**: Bringing about Sₙ is my goal.
- **Means Premise**: In order to bring about Sₙ, I need to bring about Sᵢ.
- **Conclusion** ⇒ Therefore, I need to bring about Sᵢ.
- Variables: S₀, S₁, . . . Sₙ (states of affairs, ordered temporally), Sₙ (goal), Sᵢ (means)

Critical questions ('Following the analysis of Walton (1990) once again, five critical questions are given' (p. 96)).

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| Other-Means Question | Are there alternative possible actions to bring about Sᵢ that could also lead to the goal? (p. 96) | exception | The questioner has to name the alternative; its existence does not follow from the premises. |
| Best-Means Question | Is Sᵢ the best (or most favorable) of the alternatives? (p. 96) | exception (alt. premise) | The questioner must show a better alternative. |
| Other-Goals Question | Do I have goals other than Sᵢ whose achievement is preferable and that should have priority? (p. 96) | conclusion (alt. exception) | Points to a counterargument: practical reasoning from a goal with priority, against bringing about Sᵢ. |
| Possibility Question | Is it possible to bring about Sᵢ in the given circumstances? (p. 96) | assumption (alt. premise) | Proponent's burden; not stated in the Goal/Means scheme. In the schemata it is the Practicality Premise (N4/S4). |
| Side Effects Question | Would bringing about Sᵢ have known bad consequences that ought to be taken into account? (p. 96) | exception (alt. conclusion) | The questioner must cite the known bad consequences. |

What the book says about these CQs:
- Other-Means Question: p. 96: 'Asking any one of these five critical questions can defeat the argumentation, at least temporarily. Once the proponent gives an adequate answer to the question, however, the burden shifts back to the respondent.' Read literally, that treats all five as burdens on the proponent (premise/assumption). In the necessary/sufficient schemata the Alternatives and Selection Premises (N2, N3 / S2, S3) state what this CQ asks.
- Best-Means Question: Stated as the Selection Premise (N3/S3: 'the most acceptable ... condition') in the schemata.
- Other-Goals Question: As printed, 'goals other than Sᵢ' (Sᵢ is the means, Sₙ the goal). In two-person practical reasoning the matching CQ1 speaks of goals 'that might conflict with the goal of realizing A' (p. 97).
- Side Effects Question: Stated as the Side-Effects Premise (N5/S5) in the schemata. Bad consequences could also ground an argument from negative consequences against Sᵢ (a rebuttal).

Notes:
- The rational agent is 'I'/'my'; the conclusion expresses a 'practical ought' (p. 95).
- The five CQs are printed after the necessary- and sufficient-condition schemata but use this scheme's Sᵢ notation.

Mapping: WAG **AS:practical_reasoning** (comparison in sec. 4); walton.yml: `practical_inference`, `practical_reasoning2`.
walton.yml note: walton.yml practical_inference (sources Walton 1990, 1997 p. 164): have_goal(G), action_realize_goal(A,G) -> carry_out_action(A); exceptions other_goals_conflict(G), more_efficient_alternative_actions(A,G), incompatible_consequences(A,G); assumptions possible_action(A), possible_goal(G). Closest to this scheme and its five CQs. practical_reasoning2 (instrumental, Walton 2015) is related; practical_reasoning1 (WAG's) is the value-based Atkinson & Bench-Capon scheme, not in the excerpt.

#### S29 · Necessary-Condition Schema (p. 95, 96)

*sec. 2. Practical Reasoning · PDF Part3 p.27, Part3 p.28 · source: (Walton, 1997, p. 164: 'two variants')*

- **(N1) (Goal Premise)**: My goal is to bring about A.
- **(N2) (Alternatives Premise)**: I reasonably consider on the given information that bringing about at least one of [B₀, B₁, . . . , Bₙ] is necessary to bring about A.
- **(N3) (Selection Premise)**: I have selected one member Bᵢ as an acceptable, or as the most acceptable necessary condition for A.
- **(N4) (Practicality Premise)**: Nothing unchangeable prevents me from bringing about Bᵢ as far as I know.
- **(N5) (Side-Effects Premise)**: Bringing about A is more acceptable to me than not bringing about Bᵢ.
- **(Conclusion)** ⇒ Therefore, it is required that I bring about Bᵢ.
- Variables: A (goal), B₀ ... Bₙ (alternatives), Bᵢ (selected)

Critical questions: see S28.

Notes:
- The premise names are printed in italics in parentheses after each premise, e.g. '(N1) My goal is to bring about A (Goal Premise).'
- Notation switches from S (states) to A/B here.

Mapping: WAG **AS:practical_reasoning**; walton.yml: `practical_necessary_condition`.

#### S30 · Sufficient-Condition Schema (p. 96)

*sec. 2. Practical Reasoning · PDF Part3 p.28 · source: (Walton, 1997, p. 164)*

- **(S1) (Goal Premise)**: My goal is to bring about A.
- **(S2) (Alternatives Premise)**: I reasonably consider on the given information that each one of [B₀, B₁, . . . , Bₙ] is sufficient to bring about A.
- **(S3) (Selection Premise)**: I have selected one member Bᵢ as an acceptable, or as the most acceptable sufficient condition for A.
- **(S4) (Practicality Premise)**: Nothing unchangeable prevents me from bringing about Bᵢ as far as I know.
- **(S5) (Side-Effects Premise)**: Bringing about A is more acceptable to me than not bringing about Bᵢ.
- **(no conclusion printed)**
- Variables: A (goal), B₀ ... Bₙ (alternatives), Bᵢ (selected)

Critical questions: see S28.

Notes:
- NO CONCLUSION IS PRINTED for this schema in the excerpt: the text continues after (S5) with 'Practical reasoning is put forward ...'. By analogy with N it would be 'Therefore, it is required that I bring about Bᵢ', but that line is not in the book here.

Mapping: WAG **AS:practical_reasoning**; walton.yml: `practical_sufficient_condition`.

#### S31 · two-person practical reasoning (p. 97)

*sec. 2. Practical Reasoning · PDF Part3 p.29 · source: Walton, 1997, p. 163 · name from text: 'an interesting example of two-person practical reasoning, which takes the following form (Walton, 1997, p. 163)'*

- **—**: X (the agent) intends to realize A, and tells Y (the expert) this.
- **—**: As Y sees the situation, B is a necessary (sufficient) condition for carrying out A, and Y tells X this.
- **—** ⇒ Therefore, X should carry out B, unless he has better reasons not to.
- Variables: X (the agent), Y (the expert), A, B

Critical questions (heading 'Critical Questions').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Does X have other goals (of higher priority) that might conflict with the goal of realizing A? (p. 97) | conclusion (alt. exception) | A conflicting goal of higher priority grounds a counterargument (rebuttal). |
| CQ2 | Are there alternative means available to X (other than B) for carrying out A? (p. 97) | exception | The questioner must name the alternative means. |
| CQ3 | Would carrying out B have known side effects that might conflict with X's other goals? (p. 97) | exception (alt. conclusion) | The questioner must cite the known side effects. |
| CQ4 | Is it possible for X to bring about B? (p. 97) | assumption | Proponent's burden; not stated in the premises. |
| CQ5 | Are other actions, in addition to B, required for X to bring about A? (p. 97) | exception (alt. premise) | The questioner must name the further required actions (cf. walton.yml practical_reasoning2 intervening_actions). |

Notes:
- p. 97: 'The practical form of the argument from expert opinion is two-person practical reasoning.' Example: the Lorenzo's Oil case.

Mapping: no WAG scheme (related: AS:practical_reasoning); walton.yml: `two_person_practical`.
walton.yml note: walton.yml two_person_practical (source Walton 1997, p. 163): tells_intend(A,X,B), situation_necessiy(B,Y,X,A); exceptions alternative_means, goals_higher_priority, side_effect_other_act; assumption possible_action(Y). No formula for CQ5. The audit found two exceptions with arguments in the wrong order.

#### S32 · negative form of argument from knowledge (p. 98)

*sec. 3. Lack-of-Knowledge Arguments · PDF Part3 p.30 · source: Walton, 1997, p. 95 · name from text: 'The negative form of argument from knowledge, in fact, may be considered a subspecies of the argumentum ad ignorantiam (Walton, 1997, p. 95)'*

- **(Truth Condition)**: All the propositions in E's knowledge base K on subject S are true.
- **(Completeness Condition)**: All the propositions on subject S not in E's knowledge base K are false.
- **—**: E is asked whether A is true.
- **—**: E replies that A is not true (as far as E knows).
- **—** ⇒ Therefore, A is false.
- Variables: E, K, S, A

Critical questions: none given in the excerpt.

Notes:
- The condition names are printed in italics in parentheses after the first two lines.
- p. 98: 'The main distinctive feature of this scheme is the completeness condition' (epistemic closure).

Mapping: no WAG scheme; walton.yml: none.

#### S33 · general form of ad ignorantiam (argument from ignorance) (p. 98)

*sec. 3. Lack-of-Knowledge Arguments · PDF Part3 p.30 · source: Walton, 1995, p. 150 · name from text: 'This topic is an instance of the general form of ad ignorantiam ... The form of argument is the following (Walton, 1995, p. 150)'*

- **Major Premise**: If A were true, then A would be known to be true.
- **Minor Premise**: It is not the case that A is known to be true.
- **Conclusion** ⇒ Therefore, A is not true.
- Variables: A

Critical questions: none given in the excerpt.

Example: I do not know that there is a skunk in the cabin. / Therefore, it is false that there is a skunk in the cabin. (Walton, 1995, p. 150)

Notes:
- No CQs in the excerpt.

Mapping: no WAG scheme; walton.yml: `ignorance`.
walton.yml note: walton.yml ignorance (source Walton 1995, p. 150): would_be_known(S), ¬known(S) -> ¬S; exception uninvestigated(S).

#### S34 · epistemic argument from ignorance (p. 99)

*sec. 3. Lack-of-Knowledge Arguments · PDF Part3 p.31 · source: Walton, 1992, p. 386 · name from text: 'The reasoning is from expectations, and epistemic argument from ignorance (Walton, 1992, p. 386)'*

- **—**: It has not been established that all the true propositions in D are contained in K.
- **—**: A is a special type of proposition such that if A were true, A would normally or usually be expected to be in K.
- **—**: A is in D.
- **—**: A is not in K.
- **—**: For all A in D, A is either true or false.
- **—** ⇒ Therefore, it is plausible to presume that A is false (subject to further investigations in D).
- Variables: D (domain), K (knowledge base), A

Critical questions: none given in the excerpt.

Notes:
- Applied to 'reasoning from normal expectations' (Leona Helmsley example, Walton 1996, p. 85).

Mapping: no WAG scheme; walton.yml: `epistemic_ignorence`.

#### S35 · negative practical reasoning (practical subspecies of ad ignorantiam) (p. 100)

*sec. 3. Lack-of-Knowledge Arguments · PDF Part3 p.32 · source: Walton, 1996, p. 86 · name from text: 'This negative practical reasoning follows the following scheme (Walton, 1996, p. 86)'*

- **—**: I do not know whether A is true or not.
- **—**: I have to act on the presumption that A is true or not true.
- **—**: If I act on the presumption that A is true, and A is not true, consequences B will follow.
- **—**: If I act on the presumption that A is not true, and A is true, consequences C will follow.
- **—**: Consequences B (C) are more serious than consequences C (B).
- **—** ⇒ Therefore, I act on the presumption that A is not true (true).
- Variables: A, B, C

Critical questions: none given in the excerpt.

Notes:
- Example: the firearms safety rule (Walton, 1996, p. 86).

Mapping: no WAG scheme; walton.yml: `negative_practical`.

#### S36 · Argument from Positive Consequences (p. 101, 102)

*sec. 4. Arguments from Consequences · PDF Part3 p.33, Part3 p.34 · source: Walton, 1995, p. 155; Walton, 2000, p. 141*

- **—**: If A is brought about, then good consequences will occur.
- **—** ⇒ Therefore, A should be brought about.
- Variables: A

Critical questions (heading 'Critical Questions' (p. 102), printed after the three consequence schemes (S36-S38) and not tied to one; CQ3's '(bad as opposed to good, for example)' fits the positive form).

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | How strong is the likelihood that the cited consequences will (may, must) occur? (p. 102) | premise | Questions the conditional premise that bringing about A leads to the consequences; the proponent must support it. |
| CQ2 | What evidence supports the claim that the cited consequences will (may, must) occur, and is it sufficient to support the strength of the claim adequately? (p. 102) | premise (alt. argument) | Asks the proponent for the evidence behind the same premise (proponent's burden). |
| CQ3 | Are there other opposite consequences (bad as opposed to good, for example) that should be taken into account? (p. 102) | conclusion | Points to a counterargument from the opposite consequences (a rebuttal). |

Notes:
- p. 101: 'Argument from consequences has a positive and a negative form'.

Mapping: WAG **AS:consequences** (comparison in sec. 4); walton.yml: `positive_consequences`.
walton.yml note: walton.yml positive_consequences (same formulas as WAG): brings_about(A,G), good(G) -> should_be_performed(A); no exceptions.

#### S37 · Argument from Negative Consequences (p. 101, 102)

*sec. 4. Arguments from Consequences · PDF Part3 p.33, Part3 p.34 · source: Walton, 1995, p. 155; Walton, 2000, p. 141*

- **—**: If I (an agent) bring about (don't bring about) A, then B will occur.
- **—**: B is a bad outcome (from the point of view of my goals).
- **—** ⇒ Therefore, I should not (practically speaking) bring about A.
- Variables: A, B

Critical questions: see S36 (p. 102 list).

Mapping: no WAG scheme (related: AS:consequences); walton.yml: `negative_consequences`.

#### S38 · Argument from Negative Consequences (Prudential Inference) (p. 101, 104)

*sec. 4. Arguments from Consequences (restated in 5. Fear and Danger Appeals, p. 104) · PDF Part3 p.33, Part3 p.36 · source: Walton, 1995, p. 155; Walton, 2000, p. 141 (p. 104: Walton, 2000, p. 123)*

- **—**: You were considering not doing A
- **—**: But, if you do not do A, some consequence B, which will be very bad for you. will occur or is likely to occur.
- **—** ⇒ Therefore, you ought to reconsider, and (other things being equals) you ought (prudentially) to do A.
- Variables: A, B

Critical questions: see S36 (p. 102 list).

Restated p. 104: You were considering not doing A. / But, if you don't do A, some consequence B, which will be very bad for you, will occur or is likely to occur. ⇒ Therefore, you ought to reconsider, and (other things being equal) you ought (prudentially) to do A.

Notes:
- p. 101 prints 'not doing A' without a full stop, 'for you. will occur' (full stop for comma) and 'other things being equals' (sic). The p. 104 restatement has 'A.', 'don't', 'for you, will' and 'equal'.
- p. 104: 'The argument from negative consequences ... has as a subtype the prudential scheme.'

Mapping: no WAG scheme (related: AS:consequences); walton.yml: `negative_consequences_prudential`.

#### S39 · scheme for danger appeal (p. 103)

*sec. 5. Fear and Danger Appeals · PDF Part3 p.35 · source: Walton, 2000, p. 173 · name from text: 'Here is the scheme for danger appeal (Walton, 2000, p. 173)'*

- **—**: If you (the respondent) bring about A, then B will occur.
- **—**: B is a danger to you.
- **—** ⇒ Therefore, (on balance) you should not bring about A.
- Variables: A, B

Critical questions: none given in the excerpt.

Notes:
- p. 102: 'the danger appeal argument may be considered as a special kind of argument from negative consequence based on a claim about the respondent's suffering serious harm or loss of life.'

Mapping: no WAG scheme; walton.yml: `danger`.

#### S40 · scheme for fear appeal (p. 104)

*sec. 5. Fear and Danger Appeals · PDF Part3 p.36 · source: Walton, 2000, p. 22 · name from text: 'The example fits the scheme for fear appeal (Walton, 2000, p. 22)'*

- **—**: If you do not bring about A, then D will occur.
- **—**: D is very bad for you.
- **—**: Therefore, you ought to stop D if possible.
- **—**: But the only way for you to stop D is to bring about A.
- **—** ⇒ Therefore, you ought to bring about A.
- Variables: A, D

Critical questions: none given in the excerpt.

Notes:
- A chained form: the third line is an intermediate conclusion ('Therefore, you ought to stop D if possible.').

Mapping: no WAG scheme; walton.yml: `fear`.

#### S41 · argument from fear appeal (threat structure) (p. 105)

*sec. 5. Fear and Danger Appeals · PDF Part4 p.1 · source: Walton, 1995, p. 157 · name from text: 'These conditions underlie the structure of the argument from fear appeal (Walton, 1995, p. 157)'*

- **—**: If you bring about A, some cited bad consequences, B, will follow.
- **—**: I am in a position to bring about B.
- **—**: I hereby assert that in fact I will see to it that B occurs if you bring about A.
- **—** ⇒ Therefore, you had better not bring about A.
- Variables: A, B

Critical questions: none given in the excerpt.

Notes:
- Follows the three conditions of the speech act of making a threat (preparatory, sincerity, essential; Walton, 2000, pp. 113, 114).

Mapping: no WAG scheme; walton.yml: `treat`.
walton.yml note: walton.yml 'treat' ('Argument From Treat', sic; source Walton 1995, p. 157) encodes this scheme.

#### S42 · disjunctive form of practical reasoning (indirect threat) (p. 105)

*sec. 5. Fear and Danger Appeals · PDF Part4 p.1 · source: Walton, 2000, p. 142 · name from text: 'The argument can be put in disjunctive form as an instance of practical reasoning (Walton, 2000, p. 142)'*

- **—**: You (the respondent) must bring about A or B will occur.
- **—**: B is bad or undesirable, from your point of view.
- **—** ⇒ Therefore, you should (ought to, practically speaking) bring about A.
- Variables: A, B

Critical questions: none given in the excerpt.

Notes:
- p. 106: 'appeal to threat may also assume the form of argument from alternatives'; p. 107 calls the pattern of the Ten Commandments example 'the scheme of argument from alternatives, the practical counterpart of the disjunctive syllogism', without restating it. This disjunctive form is the closest stated scheme.

Mapping: no WAG scheme; walton.yml: `pragmatic_alternative`.
walton.yml note: walton.yml pragmatic_alternative (source Walton 2000, p. 142): alternatives(A,B), undesirable(B) -> do_other(A).

#### S43 · disjunctive ad baculum threat (p. 106)

*sec. 5. Fear and Danger Appeals · PDF Part4 p.2 · source: Walton, 2000, p. 140 · name from text: 'The argument of the disjunctive ad baculum threat is, therefore, a subspecies of the argumentation scheme just presented (Walton, 2000, p. 140)'*

- **—**: You (the respondent) must bring about A or I (the proponent) will undertake to see to it that B will occur.
- **—**: B is bad or undesirable, from your point of view.
- **—** ⇒ Therefore, you should (ought to, practically speaking) bring about A.
- Variables: A, B

Critical questions: none given in the excerpt.

Mapping: no WAG scheme; walton.yml: `ad_baculum`.

#### S44 · disjunctive syllogism (p. 106)

*sec. 6. Arguments from Alternatives and Opposites · PDF Part4 p.2 · kind: deductive · name from text: 'They may in some instances have the form of the disjunctive syllogism ... It may be represented as below'*

- **—**: p∨q
- **—**: ¬q
- **—** ⇒ Therefore, p.
- Variables: p, q

Critical questions: none given in the excerpt.

Notes:
- p. 106: 'In argumentative discourse, this scheme may be used for a pragmatic argument'.

Mapping: no WAG scheme; walton.yml: `alternative_cognitive`.
walton.yml note: Closest: walton.yml alternative_cognitive (Kienpointner 1992, p. 306): can_be_case(A,B), plausibly_not_case(A) -> plausibly_case(B), a defeasible disjunctive syllogism.

#### S45 · Positive Form of Argument from Opposites (p. 107)

*sec. 6. Arguments from Alternatives and Opposites · PDF Part4 p.3 · source: (example: Rhetoric, Book II, Chapter 23, 1397a)*

- **—**: The opposite of subject S has the property P.
- **—** ⇒ Therefore, S has the property not-P (the opposite of property P).
- Variables: S, P

Critical questions: none given in the excerpt.

Notes:
- p. 107: 'Argument from opposites is a very common scheme of wide generality, treated more fully in Chapter 7.'

Mapping: no WAG scheme; walton.yml: `rhetorical_oppositions_descriptive`.
walton.yml note: Closest: rhetorical_oppositions_descriptive/normative (ch. 9 #12, p. 318): the opposite term receives the opposite relation/treatment. They build no argument in Carneades (primed variables). The opposition-* schemes (#11) oppose predicates of one subject, a different pattern.

#### S46 · Negative Form of Argument from Opposites (p. 107)

*sec. 6. Arguments from Alternatives and Opposites · PDF Part4 p.3*

- **—**: The opposite of subject S has the property not-P.
- **—** ⇒ Therefore, S has the property P (the opposite of property not-P).
- Variables: S, P

Critical questions: none given in the excerpt.

Mapping: no WAG scheme; walton.yml: `rhetorical_oppositions_descriptive`.

#### S47 · argument from need for help (p. 109)

*sec. 7. Pleas for Help and Excuses · PDF Part4 p.5 · source: Walton, 1997, p. 104 · name from text: 'The argument in this case follows the scheme of the argument from need for help (Walton, 1997, p. 104)'*

- **—**: For all x and y, y ought to help x, if x is in a situation where x needs help, and y can help, and y's giving help would not be too costly for y.
- **—**: x is in a situation where some action A by y would help x.
- **—**: y can carry out A.
- **—**: y's carrying out A would not be too costly for y – that is, the negative side effects would not be too great, as y sees it.
- **—** ⇒ Therefore, y ought to carry out A.
- Variables: x, y, A

Critical questions ('The following critical questions (Walton, 1997b, p. 155)').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Would the proposed action A really help x? (p. 109) | premise | Questions the second premise. |
| CQ2 | Is it possible for x to carry out A? (p. 109) | premise | Questions the third premise ('y can carry out A'). Printed 'for x' (the helper is y). |
| CQ3 | Would there be negative side effects of carrying out A that would be too great? (p. 109) | premise | Questions the fourth premise, which states that the side effects would not be too great. |

Notes:
- p. 109: the CQs 'are focused on the practical reasoning in the argument'.
- p. 108 gives ten points defining empathy (Walton, 1997b, pp. 73, 74) as the basis of appeals to pity; they are conditions, not a premise/conclusion scheme.

Mapping: no WAG scheme; walton.yml: `help`.

#### S48 · argument from distress (p. 110)

*sec. 7. Pleas for Help and Excuses · PDF Part4 p.6 · source: Walton, 1997b, p. 195 · name from text: 'This form of reasoning is a subscheme of the need for help, called argument from distress (Walton, 1997b, p. 195)'*

- **—**: Individual x is in distress (is suffering).
- **—**: If y brings about A, it will relieve or help to relieve this distress.
- **—** ⇒ Therefore, y ought to bring about A.
- Variables: x, y, A

Critical questions ('These are the critical questions, similar to the general scheme (Walton, 1997b, p. 155)').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Is x really in distress? (p. 110) | premise | Questions the first premise. |
| CQ2 | Will y's bringing about A really help or relieve this distress? (p. 110) | premise | Questions the second premise. |
| CQ3 | Is it possible for y to bring about A? (p. 110) | assumption | Proponent's burden; not stated in the premises. |
| CQ4 | Would negative side effects of y's bringing about A be too great? (p. 110) | exception (alt. conclusion) | Not stated in the premises; the questioner must cite the side effects. |

Mapping: no WAG scheme; walton.yml: `distress`.

#### S49 · argument from plea for excuse (p. 111)

*sec. 7. Pleas for Help and Excuses · PDF Part4 p.7 · source: Walton, 1997, pp. 154, 156 · name from text: 'This is the scheme of the argument from plea for excuse, with its relative critical questions (Walton, 1997, pp. 154, 156)'*

- **—**: Normally, rule R requires or forbids a type of action or inaction T, which carries with it a sanction (penalty) S.
- **—**: I (the pleader) have committed T.
- **—**: But I can cite special circumstances that constitute an excuse, E.
- **—** ⇒ Therefore, in this instance, I ought to be exempted from S.
- Variables: R, T, S, E

Critical questions (heading 'Critical Questions').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Does E fall under one of the recognized categories of excuses for this type of case, and, if so, can this inclusion be justified in this case? (p. 111) | premise | Questions the third premise (that the circumstances constitute an excuse). |
| CQ2 | If E does not fall under a recognized category, then what about this case is special that justifies the claim to exemption? (p. 111) | premise | Asks the pleader to support the same premise when E is not a recognized category. |
| CQ3 | If E does not fall under a recognized category, would it set a precedent, and if so, would this pose a problem in future cases? (p. 111) | conclusion (alt. exception) | Points to a counterargument against granting the exemption (a precedent slippery slope, S55). |

Mapping: no WAG scheme; walton.yml: `plea_for_excuse`.

#### S50 · argument from composition (p. 113)

*sec. 8. Composition and Division Arguments · PDF Part4 p.9 · source: Walton, 1989, p. 130 · name from text: 'The general scheme of the argument from composition is the following (Walton, 1989, p. 130)'*

- **—**: All the parts of X have property Y.
- **—** ⇒ Therefore, X has property Y.
- Variables: X (aggregate), Y (property)

Critical questions.

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Is property Y compositionally hereditary with regard to aggregate X? That is, when every part that composes X has property Y, does X (the whole) have property Y? (p. 113) | assumption (alt. argument) | The unstated warrant of the inference; the proponent must be able to support it. |

What the book says about these CQs:
- CQ1: p. 113: 'The critical aspect of these topics is the nature of the analogy between the aggregate and its elements and the relevant properties for this relation'.

Mapping: no WAG scheme; walton.yml: `composition_generic`.

#### S51 · argument from division (p. 113, 114)

*sec. 8. Composition and Division Arguments · PDF Part4 p.9, Part4 p.10 · source: Walton, 1989, p. 130 · name from text: 'In this case, the following scheme is applied (Walton, 1989, p. 130)'*

- **—**: X has property Y.
- **—** ⇒ Therefore, all the parts of X have property Y.
- Variables: X (aggregate), Y (property)

Critical questions (heading 'Critical Question').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Is property Y divisionally hereditary with regard to aggregate X? That is, when X (the whole) has property Y, does every part that composes X have property Y? (p. 114) | assumption (alt. argument) | The unstated warrant of the inference; the proponent must be able to support it. |

What the book says about these CQs:
- CQ1: p. 113: in the heavy-machine example 'the critical question fails, determining the argument's defeat'.

Example: This machine is heavy. / Therefore, all the parts of this machine are heavy. (Walton, 1989, p. 130)

Mapping: no WAG scheme; walton.yml: `division_generic`.
walton.yml note: walton.yml division_generic is a verbatim copy of composition_generic (parts -> whole), so it does not encode this scheme.

#### S52 · Scheme for Argument from Gradualism (p. 114)

*sec. 9. Slippery Slope Arguments · PDF Part4 p.10 · source: Walton, 1999, p. 93*

- **—**: Proposition A is true (acceptable to the respondent).
- **—**: There is an intervening sequence of propositions, B₁, B₂, . . . Bₙ₋₁, Bₙ, C, such that the following conditionals are true: If A, then B₁; if B₁, then B₂, . . . ; if Bₙ₋₁, then Bₙ; if Bₙ, then C.
- **—**: The conditional "If A then C" is not, by itself, acceptable to the respondent, nor are shorter sequences (than the one specified in the second premise) from A to C acceptable to the respondent.
- **—** ⇒ Therefore, the proposition C is true (acceptable to the respondent).
- Variables: A, B₁ ... Bₙ, C

Critical questions: none given in the excerpt.

Notes:
- p. 114: 'All slippery slope arguments are based on an underlying form of reasoning called argument from gradualism'.
- The citation '(Walton, 1999, p. 93)' is printed after the conclusion; the example before it is cited Walton, 1995, p. 95.

Mapping: no WAG scheme; walton.yml: `gradualism`.

#### S53 · Scheme for the Sorites Slippery Slope Argument (p. 115)

*sec. 9. Slippery Slope Arguments · PDF Part4 p.11 · source: Walton, 1992a, p. 56*

- **Initial Base Premise**: It is clearly beyond contention that aₖ has P.
- **General Inductive Premise**: If aₖ has P, then aₖ₋₁ has P.
- **Reapplication Sequence**: A sequence of modus ponens subarguments linking premises and conclusions from the clear area through the grey area.
- **Conclusion** ⇒ aᵢ may have P, for all we know (or can prove).
- Variables: aₖ, aₖ₋₁, aᵢ, P

Critical questions: none given in the excerpt.

Notes:
- Example: the heap (Walton, 1996a, p. 237).

Mapping: no WAG scheme; walton.yml: `sorites_slippery_slope`.

Uncertain readings: Subscript of a in the conclusion: read as i (aᵢ); the glyph is small and could be 1.

#### S54 · verbal classification slippery slope argument (p. 115, 116)

*sec. 9. Slippery Slope Arguments · PDF Part4 p.11, Part4 p.12 · source: Walton, 1995, p. 160 · name from text: 'a verbal classification slippery slope argument ... The structure of the scheme is represented here, following Walton (1995, p. 160)'*

- **—**: Individual a₁ has property F.
- **—**: If a₁ has F, then a₂ has F.
- **—**: Property F is vague, and so generally, if aᵢ has F, then you can't deny that the next closely neighboring individual aₙ has F.
- **—**: But quite clearly it is false that aₙ has F.
- **—** ⇒ Therefore, you can't truly say that a₁ has F.
- Variables: a₁, a₂, aᵢ, aₙ, F

Critical questions: none given in the excerpt.

Notes:
- p. 115-116: the chain ends in modus tollens, 'compared to the strategy of reductio ad absurdum'. Example: Ted and Marcia on whether a fetus is a person (Walton, 1995, p. 106).

Mapping: no WAG scheme; walton.yml: `verbal_slippery_slope`.

Uncertain readings: Third premise: 'the next closely neighboring individual aₙ' is printed with subscript n as far as the scan shows (aⱼ would be expected).

#### S55 · Scheme for Argument from Precedent Slippery Slope (p. 116, 117)

*sec. 9. Slippery Slope Arguments · PDF Part4 p.12, Part4 p.13 · source: Walton, 1996, p. 98*

- **Claim to Exceptional Status Premise**: Case C₀ is claimed to be an exception to the rule R (an excusable or exceptional case).
- **Related Cases Premise**: Case C₀ is similar to case C₁, that is, if C₀ is held to be an exception, then C₁ must be held to be an exception too (in order to be consistent in treating equal cases alike). A sequence of similar pairs of cases {Cᵢ, Cⱼ} binds us by case-to-case consistency to the series C₀, C₁, . . . Cₙ.
- **Intolerable Outcome Premise**: Treating case Cₙ as an exception to the rule R would be intolerable (for the various kinds of reasons that could be relevant).
- **Conclusion** ⇒ Case C₀ cannot be judged to be an exception to the rule (an excusable or exceptional case).
- Variables: C₀ ... Cₙ, R

Critical questions (heading 'Critical Questions for Argument from Precedent Slippery Slope').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Would C₀ set a precedent? (p. 117) | premise | Questions the Related Cases Premise (that C₀ binds us to C₁). |
| CQ2 | What is the evidence showing why each of the cited intervening precedents would occur? (p. 117) | premise (alt. argument) | Asks the proponent for evidence for the sequence in the Related Cases Premise. |
| CQ3 | Is Cₙ as intolerable as it is portrayed? (p. 117) | premise | Questions the Intolerable Outcome Premise. |

What the book says about these CQs:
- CQ2: p. 117: 'The doubtfulness of inferring the conclusion from the premise is shown by the second question. It rebuts the argument by showing that one of its necessary conditions fails to apply.'

Notes:
- p. 117: 'Argument from precedent slippery slope is a counterargument to argument from precedent.'
- Example: Satanism as a religion (Walton, 1995, p. 100).

Mapping: no WAG scheme; walton.yml: none.
walton.yml note: Not in walton.yml (the audit lists AS50 Precedent Slippery Slope as missing).

Uncertain readings: Subscripts on C are blurred: read as C₀ in the Claim, Related Cases and Conclusion lines and in CQ1; in the Intolerable Outcome Premise the subscript looks like 0 but n is expected and is written here; in CQ3 it looks like 'r.' and is read as n. Series and pairs (C₀, C₁, . . . Cₙ; {Cᵢ, Cⱼ}) are clear enough.

#### S56 · Argument from Vagueness of a Verbal Classification (p. 118, 119)

*sec. 10. Attacking Verbal Classification and Slippery Slope Arguments · PDF Part4 p.14, Part4 p.15 · source: Walton, 1996, pp. 102, 103*

- **—**: If an argument A occurs in a context of dialogue that requires a certain level of precision, but some property F that occurs in A is defined in a way that is too vague to meet the requirements of that level of precision, then A ought to be rejected as deficient.
- **—**: A occurs in a context of dialogue that requires a certain level of precision that is appropriate to that context.
- **—**: Some property F that occurs in A is defined in a way that is too vague to meet the requirements of the level of precision appropriate to the context.
- **—** ⇒ Therefore, A ought be rejected as deficient.
- Variables: A (an argument), F (a property)

Critical questions (heading 'Critical Questions for Argument from Vagueness of a Verbal Classification').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Does the context of dialogue in which A occurs demand some particular level of precision in the key terms used? (p. 119) | premise | Questions the second premise. |
| CQ2 | Is some property F that occurs in A too vague to meet the proper level of standard of precision? (p. 119) | premise | Questions the third premise. |
| CQ3 | Why is this degree of vagueness a problem in relation to the dialogue in which A was advanced? (p. 119) | assumption (alt. argument) | Asks the proponent to show that the vagueness matters for this dialogue (backing for the conditional premise). |

Notes:
- The conclusion is printed 'A ought be rejected' (sic; 'ought to be' in the first premise).
- p. 120: these refutation schemes 'are therefore classified as meta-dialectical argumentation schemes'.

Mapping: no WAG scheme; walton.yml: `verbal_classification_vagueness`.
walton.yml note: walton.yml verbal_classification_vagueness is titled 'Argument From Definition to Verbal Classification' (wrong title) but cites Walton 1996, pp. 102-103; its F is undeclared.

#### S57 · Scheme for Argument from Arbitrariness of a Verbal Classification (p. 120)

*sec. 10. Attacking Verbal Classification and Slippery Slope Arguments · PDF Part4 p.16 · source: Walton, 1996, p. 104*

- **—**: If an argument A occurs in a context of dialogue that requires a certain level of nonarbitrary definition for a key property F that occurs in A, and F is defined in an arbitrary way in A, then A ought to be rejected as deficient.
- **—**: A occurs in a context of dialogue that requires a nonarbitrary definition for a key property F that occurs in A.
- **—**: Some property F that occurs in A is defined in a way that is arbitrary.
- **—** ⇒ Therefore, A ought to be rejected as deficient.
- Variables: A (an argument), F (a property)

Critical questions (heading 'Critical Questions for Scheme from Arbitrariness of a Verbal Classification').

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Does the context of dialogue in which A occurs require a nonarbitrary definition of F? (p. 120) | premise | Questions the second premise. |
| CQ2 | Is some property F that occurs in A defined in an arbitrary way? (p. 120) | premise | Questions the third premise. |
| CQ3 | Why is arbitrariness of definition a problem in the context of dialogue in which A was advanced? (p. 120) | assumption (alt. argument) | Asks the proponent to show that arbitrariness matters for this dialogue. |

What the book says about these CQs:
- CQ3: p. 120: 'The most interesting of the following critical questions regard the problem of the context of dialogue'.

Mapping: no WAG scheme; walton.yml: `verbal_classification_arbitrariness`.
walton.yml note: walton.yml verbal_classification_arbitrariness: premise needed_nonarbitrary((A,F) does not parse, so Carneades drops it.

### Chapter 4. Arguments from Generally Accepted Opinions, Commitment, and Character

#### S58 · Seneca's form of argument from popular opinion (p. 123)

*sec. 1. Arguments from Popular Opinion · PDF Part4 p.19 · source: Walton, 1999, p. 147 · kind: stated inline in prose (in quotation marks) · name from text: 'Seneca articulated this scheme, giving it an argument form (Walton, 1999, p. 147)'*

- **—**: Everybody believes that A;
- **—** ⇒ therefore, A is true.
- Variables: A

Critical questions: none given in the excerpt.

Notes:
- Printed in running text: '"Everybody believes that A; therefore, A is true."'

Mapping: no WAG scheme; walton.yml: `pop_scheme`.

#### S59 · pop scheme (S1) (p. 123, 124)

*sec. 1. Arguments from Popular Opinion · PDF Part4 p.19, Part4 p.20 · source: Walton, 1999, p. 224 · name from text: 'This generic pattern of argument is called the pop scheme (Walton, 1999, p. 224)'*

- **S1**: Everybody (in a particular reference group) accepts that A.
- **—** ⇒ Therefore, A is true (or you should accept A).
- Variables: A

Critical questions ('These are the critical questions for the pop scheme ad populum argument (Walton, 1989, p. 89)' (p. 124)).

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Does a large majority of the cited reference group accept A as true? (p. 124) | premise (alt. exception) | It questions the scheme's own premise. |
| CQ2 | Is there other relevant evidence available that would support the assumption that A is not true? (p. 124) | conclusion | Points to evidence for not-A (a rebuttal); the opponent bears the burden (p. 124). |
| CQ3 | What reason is there for thinking that the view of this large majority is likely to be right? (p. 124) | assumption (alt. argument) | The implicit premise that the majority is likely to be right; the proponent bears the burden (p. 124). |

What the book says about these CQs:
- CQ1: p. 124: 'The first critical question attacks the premise ... While in the first two critical points of the topic the burden of proof is on the opponent who tries to undermine the reasoning, in the last case the obligation to provide satisfactory evidence is on the proponent'. Under the burden rule alone, the book's placing of the burden on the opponent would make CQ1 an exception; flagged for review.
- CQ2: p. 124: 'the second is directed against the force of the conclusion. The latter, in fact, is only plausible: if contrary evidence is presented, the whole argument is likely to be refused.'
- CQ3: p. 124: 'The last question criticizes the connection between premise and conclusion, having as its target the implicit premise of the scheme: the fact that the majority is likely to be right.'

Notes:
- p. 123: S1 and S2 'are identified in Freeman (1995, p. 70) with the bandwagon appeal and the abandon-ship fallacy, respectively.'
- The CQs speak of 'a large majority', matching the presumptive form (S61) more than 'Everybody'; p. 124: the appeal 'may refer to a large majority in the cited reference group rather than to the whole group'.

Mapping: no WAG scheme; walton.yml: `pop_scheme`.
walton.yml note: walton.yml pop_scheme: reference_group_accepted(G,A) -> true(A) and should_accept(A) (two conclusions; S2 not encoded).

#### S60 · pop scheme (S2) (p. 123)

*sec. 1. Arguments from Popular Opinion · PDF Part4 p.19 · source: Walton, 1999, p. 224 · name from text (as S59)*

- **S2**: Everybody (in a particular reference group) rejects A.
- **—** ⇒ Therefore, A is false (or you should reject A).
- Variables: A

Critical questions: see S59.

Mapping: no WAG scheme; walton.yml: `pop_scheme`.

#### S61 · ad populum scheme (pop scheme) as a presumptive form of argument (p. 124, 125)

*sec. 2. Variants of the Basic Form · PDF Part4 p.20, Part4 p.21 · name from text: 'we can reconstruct the ad populum scheme, or pop scheme, in a different guise as a presumptive form of argument'*

- **—**: If a large majority in a particular reference group G accepts A as true (false), then there exists a defeasible presumption in favor of (against) A.
- **—**: A large majority accepts A as true (false).
- **—** ⇒ Therefore, there exists a presumption in favor of (against) A.
- Variables: G, A

Critical questions: see S59.

Notes:
- The scheme is printed on p. 125; the introducing sentence ends p. 124.

Mapping: no WAG scheme; walton.yml: `popular_opinion`.
walton.yml note: Closest: walton.yml popular_opinion: generally_true(A), favor(A) -> reason_in_favor(A); its 'assumption:' key (sic) is ignored by Carneades.

#### S62 · 2.1. Position-to-Know Ad Populum Argument (p. 126)

*sec. 2. Variants of the Basic Form · PDF Part4 p.22*

- **—**: Everybody in this group G accepts A.
- **—**: This group is in a special position to know that A is true.
- **—** ⇒ Therefore, A is (plausibly) true.

Critical questions: see none given for the subtypes in the excerpt.

Mapping: no WAG scheme; walton.yml: `position_to_know_ad_populum`.

#### S63 · 2.2. Expert Opinion Ad Populum Argument (p. 126)

*sec. 2. Variants of the Basic Form · PDF Part4 p.22*

- **—**: Everybody in this group G accepts A.
- **—**: G is a group of experts in a domain of knowledge.
- **—** ⇒ Therefore, A is true.

Critical questions: see none given for the subtypes in the excerpt.

Mapping: no WAG scheme; walton.yml: `expert_opinion_ad_populum`.

#### S64 · 2.3. Deliberation Ad Populum Argument (p. 127)

*sec. 2. Variants of the Basic Form · PDF Part4 p.23*

- **—**: Group G has deliberated intelligently and extensively on whether to accept proposition A.
- **—**: Everybody in G accepts A.
- **—** ⇒ Therefore, A is (plausibly) true.

Critical questions: see none given for the subtypes in the excerpt.

Mapping: no WAG scheme; walton.yml: `deliberation_ad_populum`.

#### S65 · 2.4. Moral Justification (Excuse Subtype) Ad Populum Argument (p. 127, 128)

*sec. 2. Variants of the Basic Form · PDF Part4 p.23, Part4 p.24*

- **—**: Everybody in a (morally good) group G does x (or accepts proposition A as a policy).
- **—**: Doing x (or accepting A) shows that x (or policy A) is an acceptable norm of conduct for G.
- **—**: I (the speaker) am a member of G.
- **—** ⇒ Therefore, my doing x (or accepting A) is morally justified as an acceptable action (or policy).

Critical questions: see none given for the subtypes in the excerpt.

Mapping: no WAG scheme; walton.yml: `moral_justification_ad_populum_excuse`.
walton.yml note: walton.yml moral_justification_ad_populum_excuse has an empty conclusions: list (no argument is built).

#### S66 · 2.5. Moral Justification Ad Populum Argument (p. 128)

*sec. 2. Variants of the Basic Form · PDF Part4 p.24*

- **—**: Everybody who is good, or who represents a group G with good qualities, accepts policy P.
- **—**: Your goal is (or should be) to be a good person, or a member of a group with good qualities.
- **—** ⇒ Therefore, you should accept P.

Critical questions: see none given for the subtypes in the excerpt.

Mapping: no WAG scheme; walton.yml: `moral_justification_ad_populum`.

#### S67 · 2.6. Common Folks Ad Populum Argument (p. 129)

*sec. 2. Variants of the Basic Form · PDF Part4 p.25*

- **—**: I (the speaker) am an ordinary person, that is, I share a common background with the members of this audience (group G).
- **—**: You (the respondent) are a member of this audience (group G).
- **—** ⇒ Therefore, you should accept what I say.

Critical questions: see none given for the subtypes in the excerpt.

Mapping: no WAG scheme; walton.yml: `common_folks_group_ad_populum`, `common_folk_ad_populum`.
walton.yml note: common_folks_group_ad_populum (group_background(G), respondent_member(G) -> 0-ary accept_saying) matches the two premises; common_folk_ad_populum (0-ary being_credible -> background) does not.

#### S68 · 2.7. Rhetoric of Belonging Ad Populum Argument (p. 130)

*sec. 2. Variants of the Basic Form · PDF Part4 p.26*

- **—**: Everybody in this group G accepts A.
- **—**: Being a member of this group G is highly valued by you (the respondent).
- **—**: If you do not accept A, you will be excluded from this group G.
- **—** ⇒ Therefore, you should accept A.

Critical questions: see none given for the subtypes in the excerpt.

Mapping: no WAG scheme; walton.yml: `rhetoric_ad_populum`.

#### S69 · 2.8. Snob Appeal Ad Populum Argument (p. 130)

*sec. 2. Variants of the Basic Form · PDF Part4 p.26*

- **—**: Everybody in this group G accepts A (or has some property P, or possesses some object O).
- **—**: This group G is elite, that is, everyone who belongs to it has prestige.
- **—**: Prestige is an important goal for you (the respondent)
- **—**: If you accept A (acquire property P or buy object O), then you will become a member of the group G.
- **—** ⇒ Therefore, you ought to accept A (etc.).

Critical questions: see none given for the subtypes in the excerpt.

Notes:
- The third premise has no final full stop in the print.

Mapping: no WAG scheme; walton.yml: `snob_appeal_ad_populum`.

#### S70 · 2.9. Appeal to Vanity Ad Populum Argument (p. 131)

*sec. 2. Variants of the Basic Form · PDF Part4 p.27*

- **—**: Everybody in this group of admired (popular) people G accepts A (possesses A, etc.).
- **—**: If you carry out action x, then you will belong to this group G.
- **—** ⇒ Therefore, you should carry out action x.

Critical questions: see none given for the subtypes in the excerpt.

Mapping: no WAG scheme; walton.yml: `appeal_to_vanity_ad_populum`.

#### S71 · argument from popular practice (scheme from popular practice) (p. 131)

*sec. 2. Variants of the Basic Form · PDF Part4 p.27 · name from text: 'This is the form of the scheme from popular practice.'*

- **—**: If a large majority of a particular reference group G does A, or acts as though A is the right (or an acceptable) thing to do, then A is a prudent course of action.
- **—**: A large majority of a particular reference group G acts as though A is the right thing to do.
- **—** ⇒ Therefore, A is a prudent course of action.
- Variables: G, A

Critical questions: none given in the excerpt.

Notes:
- p. 131: 'The argument from popularity also has a practical form'. Example: the sailboat race (Walton, 1996, p. 84).

Mapping: no WAG scheme; walton.yml: `popular_practise`.
walton.yml note: walton.yml popular_practise (source Walton 2006, FCA p. 93): popular_practise(A), to_think_acceptable(A) -> acceptable_in_case(A).

#### S72 · argument from commitment (p. 134, 135)

*sec. 3. Argument from Commitment · PDF Part4 p.30, Part4 p.31 · source: Walton, 1996, p. 56 · name from text: 'The scheme for argument from commitment is the following (Walton, 1996, p. 56)'*

- **Major Premise**: If arguer a has committed herself to proposition A at some point in a dialogue, then it may be inferred that she is also committed to proposition B, should the question of whether B is true become an issue later in the dialogue.
- **Minor Premise**: Arguer a has committed herself to proposition A at some point in a dialogue.
- **Conclusion** ⇒ At some later point in the dialogue, where the issue of B arises, arguer a may be said to be committed to proposition B.
- Variables: a, A, B

Critical questions ('These are the critical questions corresponding to argument from commitment (Walton, 1996, p. 56)' (p. 135)).

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | Is a really committed to A, and if so, what evidence supports the claim that she is so committed? (p. 135) | premise | Questions the minor premise and asks for its evidence. |
| CQ2 | If the evidence for commitment is indirect or weak, could there also be contrary evidence, or at least room for the rebuttal that this case is an exception? (p. 135) | exception (alt. conclusion) | The questioner cites contrary evidence or the exception. |
| CQ3 | Is the proposition A, as cited in the premise, identical to the proposition A as cited in the conclusion? If not, what exactly is the nature of the relationship between the two propositions? (p. 135) | premise (alt. argument) | Questions the link from A to B stated in the major premise (proponent's burden). |

Notes:
- CQ3 is printed as asking about 'the proposition A as cited in the conclusion', although the printed conclusion is about B.
- Example: Bob and Ed on the union dispute (Walton, 1996c, p. 55).

Mapping: no WAG scheme; walton.yml: `commitment1`, `commitment2`.
walton.yml note: walton.yml commitment1: committed(A,X), also_committed(X,Y) -> commined_to(A,Y) (typo in the conclusion predicate); exception commitment_contrary_evidence (CQ2); assumption commitment_support_evidence (CQ1). commitment2 is a variant.

#### S73 · argument from sunk costs (p. 135, 136)

*sec. 3. Argument from Commitment · PDF Part4 p.31, Part4 p.32 · source: Walton, 2002, p. 489 · name from text: 'Argument from sunk costs is represented by the following scheme (Walton, 2002, p. 489)'*

- **—**: There is a choice at t2 between A and not-A.
- **—**: At t2 I am pre-committed to A because of what I did or committed myself to at t1.
- **—** ⇒ Therefore, I should choose A.
- Variables: t1 (time of pre-commitment), t2 (time of decision), A

Critical questions: none given in the excerpt.

Notes:
- The scheme is printed on p. 136; it is introduced at the foot of p. 135. Example: Susan's Ph.D. (argument from waste, p. 10 and p. 135).

Mapping: no WAG scheme; walton.yml: `sunk_costs`, `waste`.
walton.yml note: walton.yml sunk_costs cites the same source (Walton 2002, p. 489) but encodes a different form: sunk_costs(A,C), too_high_to_waste(C) -> should_be_performed(A); assumption feasible(A); exception future_losses_outweigh(A).

#### S74 · argument from inconsistent commitment (p. 136, 137)

*sec. 4. Arguments from Inconsistency · PDF Part4 p.32, Part4 p.33 · source: Walton, 1998, p. 220 · name from text: 'The following is the scheme for argument from inconsistent commitment (Walton, 1998, p. 220)'*

- **—**: a is committed to proposition A (generally, or in virtue of what she has said in the past).
- **—**: a is committed to proposition ¬A, which is the conclusion of the argument α that a presently advocates.
- **—** ⇒ Therefore, a's argument α should not be accepted.
- Variables: a, A, α

Critical questions ('These are the critical questions (Walton, 1998, p. 220)' (p. 137)).

| Label | Question (p.) | Draft locus | Reason |
|---|---|---|---|
| CQ1 | What are the propositions alleged to be practically inconsistent, and are they practically inconsistent? (p. 137) | premise | Attacks the alleged incompatibility of the two commitments; the burden is on the proponent (p. 137). |
| CQ2 | If the identified propositions are not practically (pragmatically) inconsistent, as things stand, are there at least some grounds for a claim of practical inconsistency that can be evaluated from the textual evidence of the discourse? (p. 137) | premise | Same: the proponent must show grounds for the inconsistency (p. 137). |
| CQ3 | Even if there is not an explicit practical inconsistency, what is the connection between the pair of propositions alleged to be inconsistent? (p. 137) | premise (alt. assumption) | Same: the proponent must show the connection (p. 137). |
| CQ4 | If there is a practical inconsistency that can be identified as the focus of the attack, how serious a flaw is it? Could the apparent conflict be resolved or explained without destroying the consistency of the commitment in the dialogue? (p. 137) | exception (alt. argument) | The opponent argues that the conflict can be resolved or explained (p. 137: 'the opponent attacks the link between premise and conclusion'). |
| CQ5 | Does it follow from a's inconsistent commitment that his argument should not be accepted? (p. 137) | argument (alt. conclusion) | Challenges the inference itself, with no statement behind it (an undercutter). |
| CQ6 | Is the conclusion the weaker claim that a's credibility is open to question or the stronger claim that the conclusion of his argument is false? (p. 137) | argument (alt. conclusion) | Challenges how strong a conclusion the premises support (an undercutter). |

What the book says about these CQs:
- CQ1: p. 137: 'The first three critical questions ... attack the incompatibility of the two commitments ... The burden of proof is on the proponent ... In the fourth question, the opponent attacks the link between premise and conclusion ... The last two questions are directed against the force of the conclusion'.

Notes:
- The CQs speak of 'practical inconsistency', the subject of the next scheme (S75).

Mapping: no WAG scheme; walton.yml: `inconsistent_commitment`.

#### S75 · argument from pragmatic inconsistency (p. 138)

*sec. 4. Arguments from Inconsistency · PDF Part4 p.34 · source: Walton, 1998, p. 218 · name from text: 'The following scheme represents the general form of the argument from pragmatic inconsistency (Walton, 1998, p. 218)'*

- **—**: a advocates argument α, which has proposition A as its conclusion.
- **—**: a has carried out an action, or set of actions, that imply that a is personally committed to ¬A (the opposite or negation of A).
- **—** ⇒ Therefore, a's argument α should not be accepted.
- Variables: a, α, A

Critical questions: none given in the excerpt.

Notes:
- p. 138 mentions 'The second critical question of the scheme' and 'The distinctive critical question', but no CQ list for this scheme is printed in the excerpt (it ends at p. 138). It may refer to the p. 137 list (S74) or to a list on p. 139; not verifiable here.

Mapping: no WAG scheme; walton.yml: `pragmatic_inconsistency`.

## 3. Named or cited only (not stated in the excerpt)

| Book p. | PDF | Name | Kind | Context | Stated elsewhere in excerpt as |
|---|---|---|---|---|---|
| 3 | Part1 p.13 | argument from sign, argument from example, argument from commitment, argument from position to know, argument from expert opinion, ad hominem argument, argument from analogy, argument from precedent, argument from gradualism, several types of slippery slope argument | scheme list | Walton (1996) list of 'twenty-six defeasible argumentation schemes' | sign S01, commitment S72, position to know S02, expert opinion S03, analogy S13, precedent S22, gradualism S52, slippery slopes S53-S55; example and ad hominem are not stated in the excerpt |
| 4 | Part1 p.14 | compendium of sixty-five schemes (Chapter 9) | scheme list | 'In Chapter 9, a compendium of schemes has been produced that presents sixty-five schemes.' (The table of contents lists 60 numbered entries, some with subtypes.) | – |
| 10 | Part1 p.20 | abductive inference, or inference to the best explanation | scheme | Argument from sign 'is closely related to' it | – |
| 10 | Part1 p.20 | modus ponens; inductive forms (statistical conclusion from data) | deductive/inductive form | examples of deductive and inductive schemes | – |
| 10 | Part1 p.20 | argument from waste (Perelman and Olbrechts-Tyteca, 1969), 'more usually called the argument from sunk costs' | scheme | Susan's Ph.D. example (Walton, 1996, p. 81); continues on p. 11 | sunk costs S73 (p. 136); argument from waste not stated |
| 12 | Part1 p.22 | argument from example, argument from commitment, argument from position to know, argument from expert opinion, argument from analogy, argument from precedent, argument from gradualism, slippery slope argument | scheme list | presumptive schemes presented and analyzed by Walton (1996); argument from sign also listed | as for p. 3 |
| 12 | Part1 p.22 | modus ponens, disjunctive syllogism | deductive form | deductively valid forms | disjunctive syllogism S44 (p. 106) |
| 16 | Part1 p.26 | modus brevis | argument form | the enthymematic form analysed as a separate argument form (Sadock, 1977) | – |
| 19 | Part1 p.29 | defeasible modus ponens | scheme form | Appeal to Expert Opinion (Version II) 'has the form we could call defeasible modus ponens'; also p. 77 (DMP, Verheij 2003a) | – |
| 21 | Part1 p.31 | circumstantial argument against the person (circumstantial ad hominem) | scheme | Walton, 1996, p. 58; 'the premise usually left implicit in circumstantial ad hominem arguments' | – |
| 22 | Part1 p.32 | affirming the consequent | fallacy | example on the scale from good to bad argument | – |
| 26 | Part2 p.2 | appeal to unqualified authority / argumentum ad verecundiam | fallacy | Hurley (2000, p. 139); Bradshaw and tobacco examples; also p. 29 | expert opinion S03 |
| 34 | Part2 p.10 | argument from lack of knowledge | scheme | 'This defeasible notion of sufficiency incorporates an argument from lack of knowledge.' | lack-of-knowledge arguments S32-S35 |
| 42 | Part2 p.18 | argument from position to know, argument from consequences, argument from appearance, schemes for rule application (precedent-based reasoning, argument from analogy) | scheme list | schemes cited by Prakken (2005) for legal reasoning | position to know S02, consequences S36-S38, precedent S22, analogy S13; argument from appearance not stated |
| 43 | Part2 p.19 | argument from verbal classification, argument from precedent, argument from an established rule | scheme list | chapter introduction (pp. 43-45) | S18, S22, S20 |
| 43 | Part2 p.19 | case-based reasoning (CBR) | reasoning type | argument from analogy is its foundation | – |
| 51 | Part2 p.27 | fallacy of faulty analogy | fallacy | Johnson and Blair (1983) | – |
| 68 | Part2 p.44 | fallacy of begging the question; question-begging epithet | fallacy | verbal classification 'heresy' example (Walton, 1996, p. 54) | – |
| 74 | Part3 p.6 | slippery slope and its variants | scheme family | based on practical argument from analogy (pp. 74-75) | S52-S55 |
| 77 | Part3 p.9 | defeasible modus ponens (DMP) | scheme form | 'the thesis of Verheij (2003a) ... that all of the defeasible argumentation schemes have this form' | – |
| 83 | Part3 p.15 | appeal to witness testimony; argument from expert opinion (combined with analogy) | scheme | argument from analogy is 'commonly combined with other forms of argument' | S26, S03 |
| 84 | Part3 p.16 | post hoc fallacy (leaping from a correlation to a causal conclusion) | fallacy | compared with deceptive argument from analogy; continues p. 85 | – |
| 85 | Part3 p.17 | argument from appearance | scheme | Prakken (2003) example 'this object looks like an affidavit; therefore, it is an affidavit' (an instance, not a scheme) | – |
| 85 | Part3 p.17 | abductive reasoning ('a very weak form of abductive reasoning') | reasoning type | striking-similarity copyright cases | – |
| 87 | Part3 p.19 | argument from ignorance / argumentum ad ignorantiam; negative scheme 'the expert source does not know that the proposition is true; therefore, (plausibly) it is false' | scheme | sketched in prose pp. 87-88; also 'the proposition may be falsified because its opposite, contradictory, or contrary is known to be true' (p. 88) | S32-S35 |
| 94 | Part3 p.26 | phronesis (practical wisdom); practical inferences | reasoning type | Aristotle; Walton 1990 | S28 |
| 99 | Part3 p.31 | reasoning from normal expectations | scheme | one of two applications of ad ignorantiam; example only (Leona Helmsley) | epistemic argument from ignorance S34 |
| 100 | Part3 p.32 | causal argumentation schemes / causal inference | scheme family | 'a topic treated in Chapter 5' (outside the excerpt); also p. 94 | – |
| 100 | Part3 p.32 | appeal to pity; danger and pity appeals | scheme family | pp. 100-101 | need for help S47, distress S48 |
| 102 | Part3 p.34 | fear or threat appeal; scare tactics | scheme | pp. 102, 104-106 | S39-S43 |
| 106 | Part4 p.2 | modus tollens | deductive form | argument from negative consequences as its 'plausible counterpart'; also pp. 115-116 | – |
| 107 | Part4 p.3 | argument from alternatives ('the practical counterpart of the disjunctive syllogism') | scheme | Ten Commandments example; the scheme is not restated here | closest: disjunctive practical reasoning S42 (p. 105) |
| 108 | Part4 p.4 | plea from excuse; arguments from need for help and distress | scheme | section introduction | S47-S49 |
| 112 | Part4 p.8 | fallacy of composition and division | fallacy | section introduction | S50, S51 |
| 115 | Part4 p.11 | bald man argument; heap | example | instances of sorites | S53 |
| 115 | Part4 p.11 | hypothetical syllogism rule; reductio ad absurdum | deductive form | verbal classification slippery slope | – |
| 121 | Part4 p.17 | argumentum ad populum; argumentum ad hominem | fallacy/scheme | chapter introduction | ad populum S58-S70; ad hominem not stated |
| 121 | Part4 p.17 | endoxic argument (appeal to popular opinion / informed public opinion) | scheme | chapter introduction | S58-S61 |
| 121 | Part4 p.17 | ethotic argumentation (ethos) | scheme | chapter introduction; sec. 5 'Ethotic Arguments' starts p. 140, outside the excerpt | – |
| 122 | Part4 p.18 | argument from popularity as a fallacy (Port Royal logic, Arnauld 1964, p. 287): 'The majority hold this opinion; therefore, it is the truest.' | fallacy | quoted as 'not a valid inference' | S58-S61 |
| 123 | Part4 p.19 | argument from judgment (Boethius): 'what seems true to everyone or the many or the wise should not be gain-said' | maxim | Boethius, 1978, p. 54 | – |
| 123 | Part4 p.19 | bandwagon appeal; abandon-ship fallacy | fallacy | Freeman (1995, p. 70), identified with pop scheme S1 and S2 | S59, S60 |
| 134 | Part4 p.30 | arguments from character (ethos); arguments from popularity, from expert opinion | scheme family | close relation between them | – |
| 135 | Part4 p.31 | argument from sunk cost; circumstantial ad hominem arguments (based on commitment) | scheme | 'The general scheme of argument from commitment is the basis for several other schemes' | sunk costs S73 |
| 135 | Part4 p.31 | argument from waste | scheme | Susan's decision (comparison of consequences) | – |
| 136 | Part4 p.32 | ad hominem argument; circumstantial type of ad hominem (based on an allegation of inconsistency) | scheme | argument from commitment 'has often been equated with the ad hominem argument' | – |

Table-of-contents entries outside the excerpt (Part1 pdf 4–7):

- Later sections: 4.5 Ethotic Arguments (p. 140); 4.6 Circumstantial Ad Hominem (p. 147); 4.7 Argument from Bias (p. 154); 4.8 Ad Hominem Strategies to Rebut a Personal Attack (p. 160); 5.2 Argument from Cause to Effect (p. 168); 5.3 Argument from Effect to Cause (p. 169); 5.4 Argument from Correlation to Cause (p. 173); 5.5 Cases in Point (p. 175).
- Ch. 9 compendium, numbered schemes; 'stated as' gives the matching entry in this excerpt:

| No. | Ch. 9 title | p. | Stated in excerpt as |
|---|---|---|---|
| 1 | Argument from Position to Know | 309 | S02, S04, S25 |
| 2 | Argument from Expert Opinion | 310 | S03, S05-S07, S27 |
| 3 | Argument from Witness Testimony | 310 | S26 |
| 4 | Argument from Popular Opinion (and Subtypes) | 311 | S58-S70 |
| 5 | Argument from Popular Practice | 314 | S71 |
| 6 | Argument from Example | 314 | **not stated** |
| 7 | Argument from Analogy | 315 | S13, S16 |
| 8 | Practical Reasoning from Analogy | 316 | S23, S24 |
| 9 | Argument from Composition | 316 | S50 |
| 10 | Argument from Division | 317 | S51 |
| 11 | Argument from Oppositions | 317 | possibly S45, S46 (uncertain) |
| 12 | Rhetorical Argument from Oppositions | 318 | possibly S45, S46 (uncertain) |
| 13 | Argument from Alternatives | 318 | possibly S42, S44 (uncertain) |
| 14 | Argument from Verbal Classification | 319 | S18 |
| 15 | Argument from Definition to Verbal Classification | 319 | **not stated** |
| 16 | Argument from Vagueness of a Verbal Classification | 320 | S56 |
| 17 | Argument from Arbitrariness of a Verbal Classification | 320 | S57 |
| 18 | Argument from Interaction of Act and Person | 321 | **not stated** |
| 19 | Argument from Values | 321 | **not stated** |
| 20 | Argument from Sacrifice | 322 | **not stated** |
| 21 | Argument from the Group and Its Members | 322 | **not stated** |
| 22 | Practical Reasoning | 323 | S28-S30 |
| 23 | Two-Person Practical Reasoning | 325 | S31 |
| 24 | Argument from Waste | 326 | **not stated** |
| 25 | Argument from Sunk Costs | 326 | S73 |
| 26 | Argument from Ignorance | 327 | S33 |
| 27 | Epistemic Argument from Ignorance | 328 | S34 |
| 28 | Argument from Cause to Effect | 328 | **not stated** |
| 29 | Argument from Correlation to Cause | 328 | **not stated** |
| 30 | Argument from Sign | 329 | S01 |
| 31 | Abductive Argumentation Scheme | 329 | **not stated** |
| 32 | Argument from Evidence to a Hypothesis | 331 | **not stated** |
| 33 | Argument from Consequences | 332 | S36-S38 |
| 34 | Pragmatic Argument from Alternatives | 333 | possibly S42 (Walton 2000, p. 142; uncertain) |
| 35 | Argument from Threat | 333 | possibly S41 or S43 (uncertain) |
| 36 | Argument from Fear Appeal | 333 | S40 or S41 (uncertain which) |
| 37 | Argument from Danger Appeal | 334 | S39 |
| 38 | Argument from Need for Help | 334 | S47 |
| 39 | Argument from Distress | 334 | S48 |
| 40 | Argument from Commitment | 335 | S72 |
| 41 | Ethotic Argument | 336 | **not stated** |
| 42 | Generic Ad Hominem | 336 | **not stated** |
| 43 | Pragmatic Inconsistency | 336 | S75 |
| 44 | Argument from Inconsistent Commitment | 337 | S74 |
| 45 | Circumstantial Ad Hominem | 337 | **not stated** |
| 46 | Argument from Bias | 338 | **not stated** |
| 47 | Bias Ad Hominem | 338 | **not stated** |
| 48 | Argument from Gradualism | 339 | S52 |
| 49 | Slippery Slope Argument | 339 | **not stated** |
| 50 | Precedent Slippery Slope Argument | 340 | S55 |
| 51 | Sorites Slippery Slope Argument | 340 | S53 |
| 52 | Verbal Slippery Slope Argument | 341 | S54 |
| 53 | Full Slippery Slope Argument | 341 | **not stated** |
| 54 | Argument for Constitutive-Rule Claims | 342 | **not stated** |
| 55 | Argument from Rules | 343 | possibly S20 (uncertain) |
| 56 | Argument for an Exceptional Case | 344 | S21 |
| 57 | Argument from Precedent | 344 | S22 |
| 58 | Argument from Plea for Excuse | 345 | S49 |
| 59 | Argument from Perception | 345 | **not stated** |
| 60 | Argument from Memory | 346 | **not stated** |

## 4. WAG vs the book

Compared files: `wag/data/nascq/schemes.json`, `critical_questions.json`, `edges.json` (ask edges with locus and pattern) and `wag/data/schemes/core.yml`, read from the working tree at build time (file times: `nascq/schemes.json` 2026-10-13 14:08:03 UTC; `nascq/critical_questions.json` 2026-10-13 14:08:03 UTC; `nascq/edges.json` 2026-10-13 05:22:08 UTC; `schemes/core.yml` 2026-10-13 14:07:04 UTC).

- While this catalogue was being made, `schemes.json`, `critical_questions.json` and core.yml's language templates were reworded (uncommitted). `edges.json` and the core.yml formulas equal HEAD `cbf69b3`.
- The tables show the current question text and, where it differs, the HEAD text in parentheses.
- The notes were written against the current text; if the wording changes again, re-check them.
- 'Draft' is the locus drafted in this catalogue.

### AS:practical_reasoning — Practical Reasoning (carneades `practical_reasoning1`)

Book entries: S28 (argumentation scheme for practical reasoning (Goal Premise / Means Premise), p. 95, 96), S29 (Necessary-Condition Schema, p. 95, 96), S30 (Sufficient-Condition Schema, p. 96), S31 (two-person practical reasoning, p. 97).

**Premises**
- This is a different scheme. WAG's carneades_id practical_reasoning1 is Atkinson & Bench-Capon's value-based scheme. Its four premises are: the current situation is S1; doing A in S1 would lead to S2; goal G would be achieved in S2; achieving G would promote value V.
- The excerpt states Walton's goal/means scheme instead (p. 95): 'Goal Premise: Bringing about Sₙ is my goal.' / 'Means Premise: In order to bring about Sₙ, I need to bring about Sᵢ.' It also gives the five-premise necessary- and sufficient-condition schemata (pp. 95-96), with Goal, Alternatives, Selection, Practicality and Side-Effects premises.
- The book has no circumstances premise and no value premise. The value-related formulas in core.yml have no counterpart in the excerpt: would_promote_value, the assumptions legitimate_value(V) and worthy_goal(G), and the exception promote_more_effectively(V,A).
- The book's agent speaks in the first person ('my goal', 'I need to').
- WAG does not represent the book's distinction between Bᵢ as a necessary condition and Bᵢ as a sufficient condition.

**Conclusion.** WAG: '{A} should be done'. Book: 'Therefore, I need to bring about Sᵢ.', a 'practical ought' (p. 95); the N-schema has 'Therefore, it is required that I bring about Bᵢ.' The S-schema prints no conclusion.

**Critical questions**

| Book CQ | WAG CQ | Status | WAG locus / pattern | Draft locus | Note |
|---|---|---|---|---|---|
| Other-Means Question: Are there alternative possible actions to bring about Sᵢ that could also lead to the goal? | CQ:other_means: Could an action other than {A} also achieve {G}? (HEAD: Are there alternative actions to {A} that also achieve {G}?) | different wording | exception `realize_more_effectively({G},{A})` | exception | WAG's current text asks only whether another action could also achieve the goal, as the book does. Its exception pattern, realize_more_effectively({G},{A}), says something stronger: another action would achieve the goal MORE effectively. That is the Best-Means question. So the question text fits the book but the pattern does not: an alternative that is merely as good answers the CQ with yes without making the exception true. |
| Best-Means Question: Is Sᵢ the best (or most favorable) of the alternatives? | CQ:best_means: Is {A} the best way to achieve {G}? | different wording | exception `bring_about_more_effectively({S2},{A})` | exception | Same meaning. WAG's pattern bring_about_more_effectively({S2},{A}) is close to other_means' realize_more_effectively({G},{A}), so the two WAG exceptions overlap. |
| Other-Goals Question: Do I have goals other than Sᵢ whose achievement is preferable and that should have priority? | CQ:other_goals: Are there goals other than {G} that should come first? (HEAD: Are there goals other than {G} that should have priority?) | different wording | conclusion | conclusion | WAG asks about goals other than {G}, the goal. The book prints 'goals other than Sᵢ', the means, and adds 'whose achievement is preferable'. Two-person CQ1 speaks of goals that 'might conflict with the goal of realizing A'. Both WAG and the draft put this CQ at the conclusion. |
| Possibility Question: Is it possible to bring about Sᵢ in the given circumstances? | CQ:possibility: Is {A} actually possible? (HEAD: Is it actually possible to perform action {A}?) | different wording | premise `possible({A})` | assumption | WAG drops 'in the given circumstances', although the scheme has a circumstances variable, S1. WAG's edge gives locus 'premise', but possible(A) is an assumption in core.yml. Both mean the proponent bears the burden, so this is a WAG labelling inconsistency, not a conflict with the book. |
| Side Effects Question: Would bringing about Sᵢ have known bad consequences that ought to be taken into account? | CQ:side_effect: Would doing {A} have bad side effects? (HEAD: Would performing {A} have undesirable side effects?) | different wording | exception `side_effects({A},{S1})` | exception | WAG drops 'known' and 'that ought to be taken into account'. 'Known' matters: it makes the questioner cite actual consequences, which fits the exception locus that WAG and the draft share. |
| two-person CQ5: 'Are other actions, in addition to B, required for X to bring about A?' (p. 97) | – | missing in WAG | – | – | This CQ belongs only to the two-person variant, and WAG has no such scheme. |

**Other**
- p. 96: 'Asking any one of these five critical questions can defeat the argumentation, at least temporarily. Once the proponent gives an adequate answer ... the burden shifts back to the respondent.' This passage treats all five CQs as putting the burden on the proponent, i.e. as premises or assumptions. WAG, like Carneades, makes three of them exceptions; the excerpt does not settle which reading is right.
- No book CQ corresponds to the core.yml assumptions legitimate_value(V) and worthy_goal(G) or to the exception promote_more_effectively(V,A). WAG has no CQ edges to them either.

### AS:consequences — Argument from Positive Consequences (carneades `positive_consequences`)

Book entries: S36 (Argument from Positive Consequences, p. 101, 102), S37 (Argument from Negative Consequences, p. 101, 102), S38 (Argument from Negative Consequences (Prudential Inference), p. 101, 104).

**Premises**
- The book's positive form (p. 101) has a single premise: 'If A is brought about, then good consequences will occur.'
- WAG has two premises: 'Doing {A} brings about {G}' and '{G} is good'. WAG names the consequence and asserts its goodness separately. That is the shape of the book's NEGATIVE form: '... then B will occur.' / 'B is a bad outcome (from the point of view of my goals).'

**Conclusion.** Book: 'Therefore, A should be brought about.' WAG: '{A} should be done'. Same meaning.

**Critical questions**

| Book CQ | WAG CQ | Status | WAG locus / pattern | Draft locus | Note |
|---|---|---|---|---|---|
| CQ1: How strong is the likelihood that the cited consequences will (may, must) occur? | CQ:likelihood: How likely is it that the claimed link actually holds? (HEAD: How likely is it that the claimed effect will actually occur?) | different wording | premise `brings_about({A},{G})` | premise | WAG uses one generic question, 'the claimed link', for both consequences and sign. It drops the book's modality '(may, must)'. Both WAG and the draft put it at the brings_about premise. |
| CQ2: What evidence supports the claim that the cited consequences will (may, must) occur, and is it sufficient to support the strength of the claim adequately? | CQ:evidence: Is there enough evidence that doing {A} brings about {G}? (HEAD: Is there sufficient evidence for the claimed connection between {A} and {G}?) | different wording | argument | premise | WAG's 'enough evidence' covers the book's sufficiency clause. The loci differ. WAG uses argument (an undercutter). The draft uses premise, because the CQ asks the proponent for the evidence behind the premise; argument is its alternative. |
| CQ3: Are there other opposite consequences (bad as opposed to good, for example) that should be taken into account? | CQ:side_effect: Would doing {A} have bad side effects? (HEAD: Would performing {A} have undesirable side effects?) | different wording | conclusion | conclusion | The book asks about 'other opposite consequences (bad as opposed to good, for example) that should be taken into account'. WAG reuses its practical-reasoning side-effects question, which is narrower: 'opposite consequences' covers more than side effects. Both put it at the conclusion. |

**Other**
- WAG has a backup edge from AS:consequences to CQ:other_means. The book's consequences CQs include no other-means question, so this edge is a WAG design choice, not book content.
- The book's negative and prudential forms (S37, S38) have no WAG scheme.

### AS:values — Argument from Values (carneades `positive_value`)

Book entries: none in the excerpt.

**Premises**
- Not in the excerpt. Argument from values appears only in the table of contents (ch. 9, no. 19, p. 321, outside pp. i-138).
- WAG's formulas follow Bench-Capon (2003) and match walton.yml positive_value, which is titled 'Variant 1: Positive Value'.

**Conclusion.** Not in the excerpt.

**Critical questions**

| Book CQ | WAG CQ | Status | WAG locus / pattern | Draft locus | Note |
|---|---|---|---|---|---|
| – | CQ:legitimate_value: Is {V} really a positive value for {A}? | not in excerpt | premise `judged_positive_value({V},{A})` | – | Cannot be checked against pp. i-138. |
| – | CQ:value_effect: Does {V} really support the goal {G} for {A}? (HEAD: Does the value {V} really bear on the goal {G} of {A}?) | not in excerpt | premise `positive_affects({V},{G},{A})` | – | Cannot be checked against pp. i-138. |

**Other**
- The only values in the excerpt are the 'Three Values' of the wild-animals example (LLit, MSec, MProd; Chorley and Bench-Capon, p. 65). They are used to compare cases in argument from analogy, not as a scheme.

### AS:sign — Argument from Sign (carneades `sign`)

Book entries: S01 (Argument from Sign, p. 10).

**Premises**
- Minor premise. Book: 'Given data represented as statement A is true in this situation.' WAG: '{A} is observed'. Close.
- Major premise. Book: 'Statement B is generally indicated as true when its sign, A, is true, in this kind of situation.' WAG: '{B} is indicated when {A} holds'. WAG drops 'in this kind of situation' and 'generally', the word that marks the generalization as defeasible.
- WAG's assumption strong_correlation(A,B) is not part of the book's scheme.

**Conclusion.** Book: 'Therefore, B is true in this situation.' WAG: '{B} holds in the current situation'. Same.

**Critical questions**

| Book CQ | WAG CQ | Status | WAG locus / pattern | Draft locus | Note |
|---|---|---|---|---|---|
| – | CQ:likelihood: How likely is it that the claimed link actually holds? (HEAD: How likely is it that the claimed effect will actually occur?) | not in excerpt | premise `strong_correlation({A},{B})` | – | The excerpt gives no CQs for sign; the compendium version is on p. 329. |
| – | CQ:other_event: Could something other than {B} better explain why {A} is observed? (HEAD: Is there another event that would more reliably account for {A}?) | not in excerpt | argument | – | WAG's other-explanation question fits p. 10 ('There could be other explanations'; tracks 'planted ... by tricksters'), but the book prints no CQ. |

**Other**
- WAG has backup edges from AS:sign to CQ:side_effect and CQ:possibility. Both are practical-reasoning questions, and the excerpt gives no basis for them.

### AS:expert_opinion — Argument from Expert Opinion (carneades `credible_source`)

Book entries: S03 (Argument from Expert Opinion, p. 14, 15), S05 (Appeal to Expert Opinion (Version II), p. 19), S06 (Appeal to Expert Opinion (Version III), p. 19, 20), S07 (Appeal to Expert Opinion (Version IV), p. 20), S27 (argument from expert opinion (Walton, 2002), p. 91, 92, 93).

**Premises**
- Major premise. Book: 'Source E is an expert in subject domain S containing proposition A.' WAG: '{W} is a credible expert on {D}', plus a separate premise '{S} is a matter of {D}'. WAG folds the Expertise CQ ('credible') into the premise and splits off domain membership. The split matches version IV, which has separate Expertise and Field Premises.
- Minor premise. Book: 'E asserts that proposition A (in domain S) is true (false).' WAG: '{W} says that {S}'. WAG drops the negative reading, '(false)'.
- The variable letters clash. In the book S is the DOMAIN and A the proposition; in WAG S is the proposition and D the domain.
- WAG's carneades_id is credible_source (Wyner 2016), not walton.yml expert_opinion (Walton 1997). expert_opinion's loci follow the book's p. 33 burdens one to one, with based_on_evidence as an assumption.

**Conclusion.** Book (1997): 'A may plausibly be taken to be true (false).' Book (2002, p. 91): 'A is true (false).' WAG: '{S}'. WAG drops 'plausibly' and the (false) form.

**Critical questions**

| Book CQ | WAG CQ | Status | WAG locus / pattern | Draft locus | Note |
|---|---|---|---|---|---|
| 1. Expertise Question: How credible is E as an expert source? | CQ:expertise: How credible is {W} as an expert on {D}? (HEAD: How credible is {W} as an expert source in {D}?) | different wording | premise `credible_source({W},{D})` | premise | WAG adds the domain ('on {D}'). Both WAG and the draft treat it as a premise (p. 33: an implicit premise). |
| 2. Field Question: Is E an expert in the field that A is in? | CQ:field: Is {S} really a matter of {D}? (HEAD: Is the claim {S} within the field {D}?) | different wording | premise `in_domain({S},{D})` | premise | The two ask different things. The book asks whether the EXPERT's field is A's field. WAG asks whether the CLAIM belongs to D. Because WAG split the premise, its field CQ covers only the claim-in-D part, and whether W is an expert in D falls under CQ:expertise. Together, the two WAG CQs cover the book's question. |
| 3. Opinion Question: What did E assert that implies A? | CQ:opinion: Did {W} really say that {S}? (HEAD: Did {W} really assert {S}?) | different wording | premise `asserts({W},{S})` | premise | The two ask different things. The book's question is open: 'What did E assert that implies A?'. Subquestions 3.1-3.4 cover quotation, inference and interpretation. WAG's question is yes/no: did W really say S. That is the shape of the position-to-know CQ3, and it loses the indirect 'implies'. |
| 4. Trustworthiness Question: Is E personally reliable as a source? | CQ:trustworthiness: Is {W} honest and unbiased? (HEAD: Is {W} reliable as a source, honest and unbiased?) | different wording | exception `biased({W})` | exception | The book asks 'Is E personally reliable as a source?', with three subquestions: biased? honest? conscientious? WAG's current text asks only two of them, honest and unbiased; it drops 'personally reliable' and 'conscientious'. WAG's edge targets only biased({W}). credible_source in core.yml also has the exception dishonest(W), but no CQ edge reaches it for this scheme, although the WAG question mentions honesty. Both WAG and the draft treat the CQ as an exception (p. 33). |
| 5. Consistency Question: Is A consistent with what other experts assert? | CQ:consistency: Do other credible sources agree that {S}? (HEAD: Is {S} consistent with what other credible sources assert?) | different wording | exception `other_credible_sources_disagree({S})` | exception | Two wording differences. The book compares with other experts, WAG with other credible sources. The book asks whether A is CONSISTENT with them, WAG whether they AGREE. Sources can be consistent with A by saying nothing about it, so a 'no' to WAG's question need not mean that they disagree. Yet WAG's exception pattern is other_credible_sources_disagree({S}). Both treat the CQ as an exception. The book (p. 33) also calls it 'a rebuttal', so conclusion is the alternative. |
| 6. Backup Evidence Question: Is E's assertion based on evidence? | CQ:backup_evidence: Is what {W} says based on evidence? (HEAD: Is the assertion of {W} based on evidence?) | same | argument | assumption | The meaning is the same. The loci differ. WAG uses argument (an undercutter). On p. 33 the book treats this CQ as an implicit premise, i.e. an assumption the proponent must support. On p. 92 it groups CQs 5 and 6 as attacks on 'the link between premises and conclusion', which supports WAG's choice. The two passages disagree. |
| subquestions 1.1-6.3 (pp. 92-93) and trustworthiness subquestions (p. 31) | – | missing in WAG | – | – | There are 21 subquestions on pp. 92-93, and 4.1-4.3 are also printed on p. 31. WAG has none; they could serve as finer CQ nodes or as hints. |

### AS:position_to_know — Argument from Position to Know (carneades `position_to_know`)

Book entries: S02 (Argument from Position to Know, p. 13), S04 (argument from position to know (as recalled on p. 17), p. 17), S25 (argument from position to know (Walton, 2002), p. 89, 90).

**Premises**
- Major premise. Book: 'Source a is in a position to know about things in a certain subject domain S containing proposition A.' WAG: '{W} is in a position to know about {D}', plus '{S} is a matter of {D}' (the same split as expert opinion). The p. 17 form instead reads 'a is in a position to know whether A is true (false).', with no domain at all.
- Minor premise. Book: 'a asserts that A (in domain S) is true (false).' (p. 13) or 'a asserts that A is true (false).' (p. 89). WAG: '{W} says that {S}'. WAG drops (false).

**Conclusion.** Book: 'A is true (false).' WAG: '{S}'. The same, apart from the negative form.

**Critical questions**

| Book CQ | WAG CQ | Status | WAG locus / pattern | Draft locus | Note |
|---|---|---|---|---|---|
| CQ1: Is a in a position to know whether A is true (false)? | CQ:position: Is {W} really in a position to know about {D}? (HEAD: Is {W} in a position to know about {D}?) | different wording | premise `position_to_know({W},{D})` | premise | The book asks whether a is in a position to know 'whether A is true (false)', i.e. about the proposition. WAG asks about {D}, the domain. Both treat it as a premise. |
| CQ2: Is a an honest (trustworthy, reliable) source? | CQ:trustworthiness: Is {W} honest and unbiased? (HEAD: Is {W} reliable as a source, honest and unbiased?) | different wording | exception `dishonest({W})` | exception | The book asks 'Is a an honest (trustworthy, reliable) source?'. WAG reuses the expert-opinion node, which asks about honesty and BIAS. It adds 'unbiased', which none of the book's position-to-know CQs mention, and drops 'trustworthy, reliable'. WAG's pattern dishonest({W}) matches the book's honesty. WAG and the draft treat the CQ as an exception; the p. 17 reading would make it an assumption. |
| CQ3: Did a assert that A is true (false)? | CQ:opinion: Did {W} really say that {S}? (HEAD: Did {W} really assert {S}?) | different wording | premise `asserts({W},{S})` | premise | The same question, apart from (false). Both treat it as a premise. |
| – | CQ:field: Is {S} really a matter of {D}? (HEAD: Is the claim {S} within the field {D}?) | extra in WAG | premise `in_domain({S},{D})` | – | None of the book's three position-to-know CQs corresponds to it. The premise it questions does exist, inside the book's major premise ('domain S containing proposition A'). |

### AS:analogy — Argument from Analogy (carneades `analogy`)

Book entries: S13 (Argument from Analogy: Version 1, p. 56, 62), S16 (Argument from Analogy: Version 2, p. 58).

**Premises**
- Book version 1: 'Major Premise: Generally, case C1 is similar to case C2.' / 'Minor Premise: Proposition A is true (false) in case C1.' WAG: '{C} is a similar case' / '{A} is true in the case of {C}'.
- WAG has only one case variable, the source; the target case is left implicit.
- WAG drops 'Generally', the defeasibility marker, and (false).
- Book version 2 adds a 'Relevant Similarity Premise' (p. 58); WAG has none. The book recommends version 1 and reserves version 2 for cases where relevance must be decided (p. 86).

**Conclusion.** Book: 'Proposition A is true (false) in case C2.' WAG: '{A}'.

**Critical questions**

| Book CQ | WAG CQ | Status | WAG locus / pattern | Draft locus | Note |
|---|---|---|---|---|---|
| CQ1: Is A true (false) in C1? | CQ:truth: Is {A} really true in the case of {C}? (HEAD: Is {A} really true in the case {C}?) | different wording | premise `in_case({A},{C})` | premise | The same question, apart from (false). Both treat it as a premise; the book calls it redundant (p. 63). |
| CQ2: Are C1 and C2 similar, in the respects cited? | – | missing in WAG | – | premise | WAG has no ask edge to the similarity premise similar_case(C). The book calls this CQ redundant (p. 63), but it is the only CQ on the major premise, and p. 65 ties the case factors to CQ2 and CQ3. |
| CQ3: Are there important differences (dissimilarities) between C1 and C2? | CQ:difference: Does the case of {C} differ from this one in relevant ways? (HEAD: Are there relevant differences between {C} and the present case?) | different wording | exception `relevant_differences({C})` | exception | The book asks about 'important differences (dissimilarities)', WAG about relevant ones. Both treat the CQ as an exception. The book says it could be an undercutter or a defeater (p. 63). |
| CQ4: Is there some other case C3 that is also similar to C1 except that A is false (true) in C3? | CQ:other_case: Is there a case more on point than {C} in which {A} is false? (HEAD: Is there another case, more on point than {C}, in which {A} is false?) | different wording | exception `more_on_point({A},{C})` | conclusion | The book asks for 'some other case C3 that is also similar to C1 except that A is false (true) in C3', i.e. a counteranalogy. WAG requires a case 'more on point' than {C}, which is stronger than anything the book states. WAG's locus is the exception more_on_point. The book (p. 63) says CQ4 'looks more like a defeater', i.e. a rebuttal at the conclusion, and the draft follows that. |

### Differences that matter most for WAG

- **1. Practical reasoning is a different scheme.**
  - WAG's AS:practical_reasoning is the Atkinson & Bench-Capon value-based scheme, with circumstances, states and values. The excerpt has Walton's goal/means scheme (p. 95) and the necessary- and sufficient-condition schemata (pp. 95–96).
  - The five CQs are the same set in both.
  - WAG's other_means text now matches the book (does some other action also achieve the goal?). Its pattern, realize_more_effectively, encodes 'a better alternative', which is the Best-Means question; it overlaps best_means' pattern.
  - WAG's value formulas (legitimate_value, worthy_goal, promote_more_effectively) have no basis in the excerpt.
- **2. Expert opinion uses Wyner's credible_source, not walton.yml `expert_opinion`.**
  - The book's Field and Opinion questions are not what WAG asks.
    - Field: the book asks whether *E* is an expert in A's field; WAG asks whether the claim is a matter of D.
    - Opinion: the book asks the open 'What did E assert that implies A?'; WAG asks the yes/no 'did W really say S'.
  - Trustworthiness now asks only 'honest and unbiased'. It has lost the book's main question, 'personally reliable as a source', and the third subquestion, 'conscientious'.
  - The dishonest(W) exception has no CQ edge.
  - Consistency asks whether other sources *agree*, but the exception it maps to is that they *disagree*.
  - Backup evidence is an implicit premise (an assumption) on p. 33 but an attack on the link on p. 92. WAG chose argument.
  - walton.yml `expert_opinion` matches the book's p. 33 burdens one to one.
- **3. Analogy is missing CQ2.** WAG has no CQ on the similarity premise ('Are C1 and C2 similar, in the respects cited?').
- **4. Analogy CQ4 differs.** WAG's other_case adds 'more on point', which the book does not have. The book calls CQ4 a defeater (a counteranalogy), i.e. a rebuttal, where WAG uses an exception.
- **5. Position to know has an extra CQ.** WAG asks CQ:field, which is not among the book's three CQs. WAG's shared trustworthiness question adds bias, which the book's honesty question does not mention.
- **6. Values and sign cannot be checked.**
  - Argument from values is not in the excerpt at all (ch. 9, p. 321).
  - Sign is stated (p. 10) without CQs, so WAG's sign CQs and its backup edges to side_effect/possibility cannot be checked.
  - WAG's sign premise drops 'generally'.
- **7. Positive consequences has a different shape and wording.**
  - The book has one premise, 'If A is brought about, then good consequences will occur.'
  - CQ3 asks about 'opposite consequences', where WAG reuses its side-effects question.
  - CQ2 (evidence) is drafted here as a premise; WAG uses argument.

## 5. What the book itself says about CQ burden and role

- **p. 17** (position to know): 'the propositional content of the critical questions forms necessary assumptions if the argument scheme is to successfully carry the burden of proof ... it is necessary for an audience to accept that a is honest'. 'The complete set of linked premises employed in a scheme is thus the union of those given as premises and (the propositional content of) those listed as critical questions.'
- **p. 32**: a defeasible argument can be attacked 'by an attack on a premise, by a counterargument with an opposite conclusion, or by an argument attacking the inference rule'. Some critical questions 'function as implicit premises, while others function as starting points for finding rebuttals. The crucial difference is that the latter have a burden of proof attached for the questioner, while the former do not.'
- **p. 33** (expert opinion): CQs 1, 2, 3 and 6 are implicit premises. 4 (trustworthiness) and 5 (consistency) put 'a positive burden of proof' on the questioner; 5 'is a rebuttal'.
- **p. 62**: Verheij's four roles of CQs: question a premise; point to exceptional situations; set conditions for the proper use of a scheme; point to other arguments that might attack the scheme.
- **p. 63** (analogy): CQ1 and CQ2 are redundant; CQ4 is a defeater; CQ3 is an undercutter or a defeater. **p. 85**: CQs that merely ask whether a premise holds 'can be classified as redundant'; the rest are 'counterarguments of various kinds', some undercutters, some rebuttals.
- **p. 89–90** (position to know): CQ1 is 'directed against the major premise'; CQ2 and CQ3 'focus on the problem of the assertion'.
- **p. 92** (expert opinion): consistency and backup evidence 'are directed against the plausibility of the link between premises and conclusion'.
- **p. 96** (practical reasoning): asking any of the five CQs defeats the argument 'at least temporarily' until the proponent answers.
- **p. 117** (precedent slippery slope): CQ2 'rebuts the argument by showing that one of its necessary conditions fails to apply'.
- **p. 124** (pop scheme): CQ1 attacks the premise and CQ2 the force of the conclusion, with the burden on the opponent for both. CQ3 targets the implicit premise 'that the majority is likely to be right', with the burden on the proponent.
- **p. 137** (inconsistent commitment): CQs 1–3 attack the incompatibility, with the burden on the proponent. In CQ4 the opponent attacks the link. CQs 5–6 are 'directed against the force of the conclusion'.

## 6. Uncertain readings and print errors

**Uncertain readings.** Everything else was read with confidence from the images.
- S09 (p. 50): Premise 1 ends in 'S' . . . S"' in the scan: the marks after S are read as superscripts 1 and n (S¹ . . . Sⁿ), not certain.
- S53 (p. 115): Subscript of a in the conclusion: read as i (aᵢ); the glyph is small and could be 1.
- S54 (p. 115, 116): Third premise: 'the next closely neighboring individual aₙ' is printed with subscript n as far as the scan shows (aⱼ would be expected).
- S55 (p. 116, 117): Subscripts on C are blurred: read as C₀ in the Claim, Related Cases and Conclusion lines and in CQ1; in the Intolerable Outcome Premise the subscript looks like 0 but n is expected and is written here; in CQ3 it looks like 'r.' and is read as n. Series and pairs (C₀, C₁, . . . Cₙ; {Cᵢ, Cⱼ}) are clear enough.
- S28 (p. 96): in the five practical-reasoning CQs the subscript of S often prints like a comma ('S,'). It is clearly Sᵢ in the Possibility Question and is read as Sᵢ throughout.

**Print errors and oddities, kept as printed:**
- S30 (p. 96): the Sufficient-Condition Schema has no conclusion line.
- S28 (p. 96): the Other-Goals Question asks about goals 'other than Sᵢ' (the means).
- S38 (p. 101): 'not doing A' has no full stop; 'very bad for you. will occur'; 'other things being equals'. The p. 104 restatement corrects these.
- S27 (p. 93): subquestion 3.4 reads 'interpretation of what A said' (E is meant).
- S47 (p. 109): CQ2 reads 'Is it possible for x to carry out A?' (y is the helper).
- S56 (p. 119): the conclusion reads 'A ought be rejected as deficient.'
- S72 (p. 135): CQ3 compares 'proposition A ... as cited in the conclusion', but the conclusion is about B.
- S10 (p. 51): Burbridge's conclusion is labelled 'Premise 4'.
- S54 (p. 116): 'the next closely neighboring individual aₙ' (aⱼ expected).
- S09, S69: a premise ends without a full stop (p. 50, p. 130).
- Contents vs text: sec. 1.5 is titled 'Argument Diagramming Tools' in the contents (p. v) but '5. PEDAGOGY' on p. 21.

## 7. Stated forms not catalogued as schemes

- p. 27 (Part2 p.3): Cleaned-Up Version of the Bradshaw Example (an instance of expert opinion)
- p. 28 (Part2 p.4): Cleaned-Up Version of the Tobacco Example (an instance)
- p. 46 (Part2 p.22): Rules 1-3 and the analogy-warranting rule AWR (Weinreb, 2005, p. 24): legal rules, not schemes; Rules 1-3 again p. 75
- p. 65 (Part2 p.41): Four Factors and Three Values (Chorley and Bench-Capon, 2004, p. 94): case factors and values
- p. 69 (Part3 p.1): The Warranting Rule Argument in the Drug-Sniffing Dog Example (an instance of argument from classification); expanded p. 76, qualified rules pp. 78-79
- p. 90 (Part3 p.22): 'Witness A states that Peter shot George. Therefore, Peter shot George.' (Verheij, 2000): an instance
- p. 99 (Part3 p.31): 'I do not know that there is a skunk in the cabin. Therefore, it is false ...' (Walton, 1995, p. 150): an instance of S33
- p. 105 (Part4 p.1): Preparatory, sincerity and essential conditions of the speech act of making a threat (Walton, 2000, pp. 113, 114)
- p. 108 (Part4 p.4): Ten points defining the empathy behind appeals to pity (Walton, 1997b, pp. 73, 74)
- p. 113 (Part4 p.9): Two rules for aggregates: 'For any aggregate there is some property that does not compose it. / ... does not divide it.'
