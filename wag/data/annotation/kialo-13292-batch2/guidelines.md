# Annotating Kialo claims as Walton scheme instances: guidelines, version 2

These guidelines replace the conventions of the first pilot (`docs/kialo-pilot/pilot.md`). They follow the third-party audit of that pilot (2026-10-13, §7) and the owner's decisions of the same date:
- the schemes of `walton.yml` (the Walton–Carneades collection) are the authority on scheme definitions, with the book's compendium (Walton, Reed & Macagno 2008, ch. 9) for the critical questions;
- the translation must be accurate, with real variables and correct semantics;
- there are no new schemes for one-off patterns.

Two annotators code the same items independently. Do not look at the other annotator's work, at the first pilot, or at any earlier annotation of these claims.

## 1. What you have

| File | What it is |
|---|---|
| `items.json` | The claims to annotate, top-down (parents first). For each claim: its text, its Kialo parent(s) and relation (pro/con), its ancestors up to the thesis, its siblings and children (for context), and its impact votes. |
| `walton.yml` | The scheme definitions: for each scheme, its `variables`, `premises`, `assumptions`, `exceptions` and `conclusions` as formulas, and English templates under `language`. Use these IDs and formulas. This is the cleaned file: syntax and typos fixed, scheme content unchanged. |
| `compendium.md` / `compendium.json` | The book's 60 schemes and 68 subtypes, with their premises, conclusions and critical questions as printed. Each links to the `walton.yml` scheme that formalizes it (`walton_links`), with a map from each CQ to the formula that expresses it. |
| `families.json` | Scheme families: a generic scheme with its specializations, variants, subtypes, counterparts and context-specific forms. Use it to choose the most specific scheme the text supports. |
| `schemes.json`, `critical_questions.json`, `edges.json`, `core.yml` | What the game (WAG) can play today. There are nine NASCQ schemes (`core.yml` has a tenth, `credible_source`, kept for old games), mapped to `walton.yml` by `carneades_id`, with their critical questions and the scheme→question (ask) and question→scheme (backup) links. |

WAG's schemes and their `walton.yml` IDs:

| WAG | `walton.yml` | Note |
|---|---|---|
| `AS:practical_reasoning` | `practical_reasoning1` | Value-based (Atkinson & Bench-Capon): S1, A, S2, G, V. |
| `AS:practical_reasoning_goal` | `practical_reasoning2` | Goal-based (Walton 2015; the book's 22.1): S1, A, S2, G. |
| `AS:consequences` | `positive_consequences` | Book 33.1. |
| `AS:negative_consequences` | `negative_consequences` | Book 33.2. Concludes ¬should_be_performed(A). |
| `AS:values` | `positive_value` | |
| `AS:sign` | `sign` | |
| `AS:expert_opinion` | `expert_opinion` | The book's scheme 2. The game used Wyner's `credible_source` until 2026-10-13, while this batch was annotated. |
| `AS:position_to_know` | `position_to_know` | |
| `AS:analogy` | `analogy` | |

Any `walton.yml` scheme may be used, whether or not WAG has it. That is the point: we learn which schemes the game needs.

## 2. The unit, and what to record

Annotate each claim as it functions in the discussion, in seven separate fields (audit §7.1). Never let one field stand in for another.

1. **Literal proposition (`literal`).**
   - What the claim says, with pronouns and references resolved ("it" → "a meat tax"), but not strengthened or weakened.
   - If the claim makes several assertions, list each.
2. **Scheme (`scheme`).** The inference the text expresses.
   - `fit` is one of:
     - `explicit`: the text states the premises and the conclusion follows by the scheme;
     - `reconstructed`: an enthymeme, where you supply at least one premise or the conclusion;
     - `fragment`: the claim states one premise of a larger argument made elsewhere, such as the circumstances of a parent's practical reasoning;
     - `none`: no scheme applies, as with a bare assertion, a definition, or a question.
   - Give the `walton.yml` ID, the compendium number when one matches, and the WAG ID when there is one.
   - Use `other` when a real inference fits no scheme. Describe it in a sentence; it is material for finding new schemes.
3. **Premises (`premises`).**
   - Every premise, assumption and exception of the scheme that the instance uses, as the `walton.yml` formula with your values, plus its English.
   - For each, a status:
     - `explicit`: the text says it;
     - `inferred`: the immediate context says it, such as the parent or a sibling (cite the claim ID);
     - `supplied`: you add it.
   - No premise may go unlisted because it is implicit.
4. **Role (`role`).** What the move does in the discussion:
   - `independent_pro`, `independent_con`: a new argument for or against its target's conclusion;
   - `rebuttal`: an argument for the opposite of the target's conclusion, made as a reply;
   - `undercutter`: the target's inference does not hold here (an exception, or the inference as a whole);
   - `premise_attack`: an argument against a premise of the target;
   - `premise_backing`: an argument for a premise of the target;
   - `cq_answer`: an answer to one of the target scheme's critical questions; give which question, and whether for or against;
   - `side_effect`: a further bad consequence of the action, offered as a reply;
   - `alternative_means`: another action is said to achieve the goal as well or better;
   - `action_amendment`: proposes a changed action ("A plus B", "A only for X");
   - `clarification`, `none`.
5. **Target (`target`).**
   - The Kialo parent, and separately what the claim actually attacks or supports (audit §7.3).
   - The actual target is a claim, plus which part of its argument:
     - the conclusion;
     - a premise, by its formula;
     - an exception, by its formula;
     - the inference.
   - It may be the parent, a sibling, an ancestor, or a premise you reconstructed for one of them.
   - Set `re_parented` when the actual target is not the Kialo parent's conclusion.
6. **Evidence status (`evidence`).**
   - One of:
     - `source`: it cites one;
     - `definition`;
     - `example`;
     - `assertion`: unsupported.
   - List the URLs.
   - A link shows that a source exists, not that it proves the claim (audit §7.6). Do not raise confidence because of a link or of votes.
7. **Challenges (`challenges`).**
   - The typed critical questions this instance is open to (audit §7.5):
     - likelihood, evidence, efficacy;
     - circumstances, other means, best means;
     - side effects, opposite consequences;
     - severity or value, analogy difference, causal interference;
     - source credibility.
   - Use the compendium's CQ labels (for example "33.2 CQ1") and the formula each one targets.

## 3. Policies

- **Pros of a policy (audit §7.2).**
  - Use positive consequences when the claim chiefly says that the action brings about an outcome that is good.
  - Use practical reasoning only when the claim or context states the circumstances S1, the transition from A to S2, the goal G realized in S2, and, for the value-based scheme, the value V. Use the goal-based scheme when there is no value.
  - When a claim states only S1, or only G, without the inference, code a `fragment` of its parent's argument. Do not code a full instance with three invented premises.
- **Cons of a policy.**
  - A bad consequence of the action is negative consequences: `¬should_be_performed(A)` from `brings_about(A,G)` and `bad(G)`.
  - Bind A to the action exactly as the thesis states it, so that the con and the thesis's pros meet on one issue (§3.1, OC13).
  - A further consequence offered as a reply, and a claim that one consequence outweighs another, follow §3.1.
- **Bare denials.** "The tax would not resolve the problems" is a `premise_attack` or a `cq_answer` against the efficacy premise of the arguments it denies, with `fit: none` or `other`. It is not a new scheme.
- **Analogy and the practical analogies** (`analogy`, `practical_analogy_positive`, `practical_analogy_negative`; two-case form since 2026-10-22, decisions D001, D028, D029).
  - Each case has its own claim or action, joined by `corresponds`: P1 in the source case C1 and P2 in the target case C2; X1 in situation A1 and X2 in A2. P2 says of C2 what P1 says of C1, and X2 is to A2 what X1 was to A1.
  - Bind P2 to the very statement the claim backs or attacks (13292.61 binds it to the efficacy premise that 13292.13 denies), and X2 to the thesis's action when the thesis is the conclusion (13292.30: `tax_meat`).
  - A recommendation drawn from a finding ("it worked there, so we should do it here") is two instances: analogy to the finding about the target, then consequences or practical reasoning.
  - **Which scheme.** A practical analogy cites the source for a verdict that an action taken there was right or wrong, whatever the word (right, wrong, a mistake, unethical, should not have been done), and concludes that verdict on the corresponding action here. Analogy covers every other claim: findings, predictions, deontic statuses that are not verdicts on doing something (entitled, obliged, permitted), and evaluations of things other than actions.
  - Bind A1 and A2 as situations ("the 2019 flood") and X1 and X2 as actions ("evacuating the town a day early"). For an imagined source, the base premise or `in_case` is the audience's considered judgment.
  - A case like the source in which the claim failed is a counter-analogy: code it as an analogy for the opposite claim, a rebuttal, from the third case to the target (C1 = the third case, C2 = the target, P2 = the opposite of the claim attacked). A difference between the cases that tells against the conclusion is the exception: `difference_against`, or `difference_for` in the negative practical analogy.
  - Several source cases, proportional comparisons, a fortiori arguments and refutations by parallel argument fit the form badly or not at all. Code what fits with strain and say so in `notes`; for the rest use the near miss or `other`, and name the pattern in `notes.candidate_pattern` (`docs/candidates.md`).
- **Causal claims.**
  - `cause_to_effect` asserts that the cause has occurred. For a policy not yet enacted, say so in `notes` and lower the confidence. Do not use it as if the tax had happened.
  - Code a chain (tax → higher price → less consumption) as separate steps, each its own instance in `instances`.
- **New patterns.** Do not invent schemes. If you see a recurring inference no scheme covers, name it in `notes.candidate_pattern`.
- **Several instances.** A claim can contain more than one argument. Give each its own entry in `instances`, and mark one as primary.
- **Uncertainty.** Give `confidence` (0 to 1) separately for scheme and role: your estimate that a careful annotator would choose the same. Write what you were unsure of in `notes`.

### 3.1 Opposite consequences and weighing (decisions D055 and D059, 2026-10-23)

Each consequence scheme has an exception on its own side. The exception names the consequence that the argument cites.

| Scheme | Exception | It says |
|---|---|---|
| `positive_consequences` (33.1) | `bad_outweighs(A,G)` | A has bad consequences that outweigh G. |
| `negative_consequences` (33.2) | `good_outweighs(A,G)` | A has good consequences that outweigh G. |
| `negative_consequences_reasoning` (33.3) | `good_outweighs(A,B)` | A has good consequences that outweigh B. |
| `negative_consequences_prudential` (33.4) | `bad_outweighs(X,C)` | X has bad consequences that outweigh C. 33.4 concludes `should_be_performed(X)`. |

The book's third question about consequences (33/CQ3, opposite consequences) has two routes. A further consequence that ranks nothing is a rebuttal (OC2). A claim that one consequence matters more than another argues the exception (OC3). The rules below decide which route a claim takes.

Terms:
- *the argument it answers*: the consequence argument that a claim replies to;
- *a weighing claim*: a claim that one consequence matters more than another;
- *the one who acts*: the agent who would do the action A.

Each rule has worked cases, with their codings, in `rule-cases.yml` (in `docs/kialo-batch2`; not given to annotators, because some cases are claims of this batch). A test checks that the codings follow `walton.yml` and the gold, that every rule has a case, and that no retired wording of a rule is used. What the exceptions cannot do is recorded once, in the catalogue entry for positive consequences (`wag/data/catalogue/schemes/positive_consequences.yml`, the discrepancy on 33/CQ3).

**OC1. Scope.** These rules are for consequence claims: claims that argue from what an action would bring about (33.1–33.4).
- A further bad consequence offered against practical reasoning answers its side-effects question, at `side_effects(A,S1)` (D070).
- What the claim concludes decides, not its target or the discussion's thesis.
  - A consequence claim that concludes something about an action (doing, believing, pursuing, steering) is coded with the scheme that fits it: a consequence scheme, or practical reasoning by §3.
    - This holds when its target is a claim about what is the case too. Its conclusion then meets nothing at the target; code its role and polarity as usual. Unless it proposes a changed or another action (`action_amendment`, `alternative_means`), it is then the appeal that candidate C011 objects to: name C011 in `notes`.
  - A claim that concludes what is the case from consequences (that God exists, because believing it does good) is not coded with a consequence scheme. Code it with `other`, and name C011 in `notes`.

**OC2. A further consequence that ranks nothing is a rebuttal.** A claim that cites a further consequence of the action, without ranking it against one already argued, is a consequence argument of its own sign at the conclusion of the argument it answers. This holds however grave the claim calls the consequence. It holds after a concession with a bare "but" ("poultry is healthier, but means more miserable lives").

**OC3. A ranking argues the exception.** A claim that names a consequence of the action and a consequence that the argument it answers cites, and says that one matters more than the other, argues that argument's exception.
- Words such as outweigh, worse, more serious, more pressing, not worth, the price of and net benefit are hints, not the test.
- "A life not worth living" ranks nothing. "Both emotional and physical harms", set against emotional harms, ranks without any of those words.

**OC4. How to code a weighing claim.**
- Use `other`, unless a source-based scheme fits (expert opinion, position to know). No scheme concludes the exceptions (candidate C003).
- Its conclusion is the exception of the argument it answers, with that argument's own A and G (A and B for 33.3; X and C for 33.4): `bad_outweighs(A,G)` against an argument for A, `good_outweighs(A,G)` against an argument against A.
- Its role is `cq_answer`. Its `cq` is the third question of that argument's scheme (`33.1 CQ3`, `33.2 CQ3`, `33.3 CQ3`, `33.4 CQ3`). Its target is that argument's exception. Its polarity is `against`, its stance on that argument's conclusion (§5 item 8).
- The argument it answers is the Kialo parent, if the parent cites the outweighed consequence. Otherwise re-parent the claim to the argument that cites it (§5 item 1).
- If no argument on the board cites the outweighed consequence, code the claim as a consequence argument of its own sign, or as backing of its parent's premise, and name the comparison in `notes`.

**OC5. Values and rights.**
- A value or a right named as something the action promotes or harms is a consequence ("cleaner rivers", "pupils losing their privacy"). A ranking with it argues the exception (OC3).
- A value or a right is set against consequences only when the claim says that it rules the action in or out whatever its consequences ("no one may be watched all day"). That is not the exception. Code it with the scheme that fits it (the value schemes, the argument from rules, 55, or `other`), and name candidate C007 in `notes`.
- Two values ranked against each other, with no consequence named, are not a weighing of consequences either (C007).

**OC6. Another course is not the exception.**
- A claim that another course is better compares two actions, not two consequences of one action ("adoption, though not perfect, is better than abortion"). Code it with `other`, and name candidate C010 or C017 in `notes`. Another course is a different action (adoption, regulation, rationing).
- The action's omission is not another course. A claim that a consequence of A matters more, or less, than a harm of not doing A weighs two consequences of A: the harm of not acting counts as a good consequence of A, its not happening (OC13). Code it by OC4, with the A and G of the argument it answers.

**OC7. A comparison with a class or a whole side** ("than any benefit of the curfew") weighs against every consequence argument of that class on the board: one weighing instance for each.
- Arguments of the class that are not consequence schemes are left out. Practical reasoning has no outweighing exception (D070).
- The gold records one target per claim. Until its format takes more, code one instance: against the Kialo parent's argument if it is in the class, else against the earliest argument of the class on the board.

**OC8. A concession plus a weighing** ("true, it cuts emissions, but the jobs matter more") gives two instances: the rebuttal (OC2) and the weighing (OC4). Mark one primary:
- the weighing, when the claim's words state the comparison;
- the opposite consequence, when they state that consequence;
- when they state both alike, the instance whose target the claim's last sentence addresses.

**OC9. A claim that a cited consequence is small, rare or minor,** and names no rival consequence, attacks `good(G)` or `bad(G)`, or the consequence's likelihood (33/CQ1). It does not argue the exception.

**OC10. A reply to a weighing claim** attacks that claim's own instance: its comparison (a premise of its `other` instance, such as `more_serious_consequences(...)`) or the consequence it relies on. Only a reply that nothing outweighs G argues against the exception itself, because that would cancel every other weighing claim against the argument.
- When the weighing is the claim's secondary instance, which the gold cannot target, the reply targets the claim's primary instance and names the comparison in `notes`, until the gold's format takes a target instance.
- When the weighing is argued from a source, the reply answers the source scheme's own questions.

**OC11. Split a compound consequence** ("cleaner air and fewer accidents") into one argument per consequence when a reply weighs only one of them.

**OC12. Code one consideration once:** as a gain (33.1) or as an averted harm (33.4), not both. Code the framing the claim leads with. A claim that states both alike is coded as the gain (33.1), which the game can play.

**OC13. The one who acts, and the action.** Bind A to the thesis's action, never to its omission or its opposite.
- When the thesis states a right, A is the action that grants or secures the right ("pregnant people having access to abortion services").
- In a policy debate, the one who acts is whoever would enact the policy as the thesis states it. The people it affects are others, even when the action's phrase names them.
- A reason against doing A is a negative consequence of A (33.2).
- A harm of not doing A that falls on the one who acts is the prudential form (33.4) for A, unless the claim also states the consideration as a gain (then OC12). Its templates address the one who acts as "you" (a government whose revenue falls without the tax); a first-person claim reads "I". The game does not have 33.4 yet, so it is played with `other`.
- A harm of not doing A that falls on anyone else, the reader included when the reader does not act, is positive consequences of A (33.1). Its G is the harm's not happening, phrased as a good state ("people in abusive relationships being able to leave their abusers"; candidate C015).

**OC14. A balance claim** that names no single consequence ("all told, the costs outweigh the benefits") is a comparison with a whole side: code it by OC7, and name candidate C004 in `notes`.

## 4. Output

One JSON file: `{"annotator": "A" | "B", "items": [ … ]}`, one item per claim in `items.json` order:

```json
{
  "kialo_claim_id": "13292.7",
  "literal": ["A meat tax would make meat less affordable for poor households."],
  "instances": [{
    "primary": true,
    "scheme": {"walton_id": "negative_consequences", "compendium": "33.2", "wag_id": "AS:negative_consequences",
               "fit": "reconstructed", "confidence": 0.8,
               "alternatives": [{"walton_id": "practical_reasoning1", "why_not": "…"}]},
    "bindings": {"A": "tax meat", "G": "meat becomes unaffordable for poor households"},
    "premises": [
      {"role": "premise", "formula": "brings_about(tax_meat, meat_unaffordable_for_poor)", "en": "Taxing meat brings about …", "status": "explicit"},
      {"role": "premise", "formula": "bad(meat_unaffordable_for_poor)", "en": "That is bad.", "status": "supplied"}
    ],
    "conclusion": {"formula": "¬should_be_performed(tax_meat)", "en": "Meat should not be taxed."}
  }],
  "role": {"kind": "independent_con", "confidence": 0.85, "cq": null, "polarity": "against"},
  "target": {"kialo_parent": "13292.1", "kialo_relation": "con", "actual": {"claim": "13292.1", "part": "conclusion"}, "re_parented": false},
  "evidence": {"status": "assertion", "urls": []},
  "challenges": [{"cq": "33.2 CQ1", "type": "likelihood", "formula": "brings_about(tax_meat, meat_unaffordable_for_poor)"}],
  "notes": {"text": "…", "candidate_pattern": null}
}
```

Formulas use `walton.yml` predicates, with the values in place of the variables. Write each value as a short snake_case atom (`tax_meat`); the English in `en` uses the `language` templates. `scheme.walton_id` is `"other"` or `"none"` when no scheme applies. `role.cq` names the question (a WAG ID such as `CQ:likelihood` if one exists, else the compendium label) when the role is `cq_answer`.

## 5. Version 3: changes after the batch-2 adjudication (2026-10-13)

The two annotators and the adjudicator found these gaps. Batch 3 applies them; `gold.json` already does (rules R1–R7 of the adjudication).

1. **Re-parenting.** `re_parented` is true exactly when the actual target claim is not the Kialo parent. Attacking a premise or exception of the parent's own argument is recorded in `target.actual.part` and `formula`, not as re-parenting. The tooling computes it.
2. **Targets.** The actual target may be any claim of the discussion, the parent's siblings included. A target premise or exception is given as `part` (premise, exception, conclusion, inference) plus `formula`, and the formula must be one of the target's primary instance. Further Kialo parents go in `target.other_kialo_parents`.
3. **CQ answer or direct attack.**
   - `cq_answer` exactly when the targeted formula is one that a critical question of the target's scheme is about.
   - `validate.py` builds the table of (scheme, formula) → (compendium CQ, game CQ) from `docs/scheme-registry/links.json` and `edges.json` each time it runs. Since D055 (2026-10-23) it maps 33/CQ3 to each consequence scheme's own exception.
   - Otherwise the role is `premise_attack`, `premise_backing` or `undercutter`.
   - `side_effect` and `alternative_means` become values of `role.move`, a finer label beside the kind.
   - A question about the conclusion (opposite consequences, side effects of a consequences argument) is a `cq_answer` too.
4. **Pros of a policy.**
   - Use practical reasoning only when S2 is a state distinct from the action, and the annotator supplies at most one of: S1, A→S2, S2→G.
   - A proposal with a purpose is `positive_consequences`.
   - A bare proposal is `none` with role `action_amendment`.
5. **The primary instance** is the one whose conclusion reaches the target. Causal facts that back an evaluative premise make the classification primary and the causal chain secondary.
6. **Fit.**
   - `explicit` when every premise-role formula is explicit.
   - A `fragment` binds only the variables of the formulas it states; the others are null.
7. **Exceptions and assumptions** are listed under `premises` with their `role` (exception, assumption).
8. **Polarity** is the stance on the target's conclusion; there is no "none". A link that explains a term is `definition` evidence.
9. **The shared vocabulary** (`lexicon.json`) is issued with the batch. One atom per entity or proposition; the action is always `tax_meat`.
10. **Challenge types** add possibility and conflicting goals.
11. **Causal steps of an un-enacted policy.** Use `cause_to_effect` only if nothing better fits, with fit `reconstructed` and the note "hypothetical cause (needs a modal causal scheme, audit N8)".
    - *Note of 2026-10-22.* `walton.yml` now has that modal scheme, `cause_to_effect_modal` (the owner's decision on audit N8).
      - **Batch 2:** its gold still codes its 11 such steps by the rule above, with confidence at most 0.5, so an annotation of batch 2 that is scored against the gold follows it too, until the gold is recoded. `validate.py`'s rule R4 changes with the recoding.
      - **Batch 3 and later:** code a causal step whose cause is only proposed or possible with `cause_to_effect_modal`, with fit and confidence as for any other instance and without the N8 note.
