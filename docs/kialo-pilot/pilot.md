# Kialo → Walton annotation pilot: “Should Governments Put A Tax On Meat?”

Pilot of hand-formalizing Kialo claims as instances of Walton argumentation schemes for WAG. Kialo discussion 13292, active edges only, annotated top-down (parents before children).

## Summary

- **Scope.** 20 claims: the thesis, all 6 of its direct children (3 pros, 3 cons; Kialo has no more), and the 13 direct children of 13292.7, 13292.13 and 13292.118.
- **Scheme fit (primary label, 19 arguments).** 6 fit an existing WAG scheme, 5 a walton.yml or book scheme, 1 a NEW scheme, and 7 only Other. None of the walton.yml/book fits can be played in WAG today.
- **Roles.** 5 are CQ answers, 4 back or undermine a premise directly, 9 are plain pros or rebuttals, and 1 fits no role. **5 of 19** act on something other than their Kialo parent. Of the 7 Kialo *pros* one level down, none is a plain new argument for its parent.
- **Confidence (self-estimated).** Scheme 0.51, role 0.56, both together 0.3–0.47. Only the three pros of the thesis and the most-voted con reach 0.75 or more on role.
- **Patterns.** 15 reusable patterns, 10 with a single instance: the inventory has not saturated after 20 claims.
- **What is hard.**
  1. The commonest con is an argument from negative consequences, and WAG has no such scheme (proposal N1).
  2. Many claims are bare assertions whose reasons sit in their Kialo children.
  3. Kialo places a claim under the claim it replies to, not under the argument part it attacks, so polarity and target must be recomputed.
  4. Practical reasoning needs two to four premises the annotator invents.
  5. Value judgements on outcomes (good(G), bad(G)) have no scheme.
- **Decisions requested.**
  1. The pros-of-a-policy rule (practical reasoning or consequences; P2/P3).
  2. Cons as rebuttals (P4).
  3. Bare denials as question moves (P5).
  4. Whether claims may be re-parented.
  5. The NASCQ bundle N1–N5.
  6. A double-annotation round before scaling.

**How to read a block.**
- *Confidence* is my estimate that a careful human annotator would make the same choice.
- *Supplied by the annotator* lists the bindings the claim does not state.
- English sentences use WAG's current templates, with bound phrases in ‘quotes’ as on the board. Templates of proposed schemes are drafts.
- N-numbers refer to the NASCQ proposals listed after the patterns.

## Claims

### — Thesis —

### 13292.1 · thesis · no votes

> Governments should put a tax on meat.

**Claim type:** policy action (practical ought).

**Reading.** A policy proposal put to deliberation: governments (unspecified, so read as any or all) ought to tax meat (rate, base and which meats unspecified). It is the claim at issue; it is not itself an argument.

**Scheme:** `n/a: claim at issue, statement should_be_performed(A)`, confidence **0.9**, n/a. Source: the shared conclusion of AS:practical_reasoning and AS:consequences (core.yml should_be_performed/1).
Typical pro schemes: AS:practical_reasoning, AS:consequences. Typical cons: walton.yml negative_consequences (rebuttal; book p. 101), missing from WAG; answers to the pros' CQs: CQ:side_effect, CQ:other_means, CQ:best_means, CQ:possibility, and an efficacy CQ WAG lacks.
Alternatives: value judgement ('meat consumption is bad') (0.05: the claim names an action and an agent; 'should put a tax' is a practical ought).

**Bindings** (statement): A = ‘tax meat’. *Supplied by the annotator:* none.

⇒ ‘tax meat’ should be done.

*Template fit:* Reads well, but the agent ('governments') is lost: WAG's schemes have no agent variable.

**NASCQ needed:** N1 AS:negative_consequences (rebuttal of a policy by a bad consequence).
*Nothing for the thesis itself. Its con side needs N1: WAG has no scheme that concludes ¬should_be_performed(A).*

**Pattern:** P1, thesis: policy action ← claim should_be_performed(A); pros by practical reasoning or positive consequences, cons by negative consequences.

**Problems:**
- Generic agent and no scope (which governments, which meats, what rate). Deeper claims argue the scope (tax only beef, exempt poor countries) as if they were pros or cons; they are really amendments of the action.
- WAG schemes have no agent variable, so 'governments' disappears from the formula.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### — Direct pros and cons of the thesis —

### 13292.13 · con of 13292.1 · impact 2.25 (4 votes)

> Taxes on meat would not resolve the problems associated with its production and consumption.

**Reading.** Denies that a meat tax would achieve what its supporters want: the environmental and health problems of meat would remain. The claim gives no reason. It is aimed at the means-end premise that the pro arguments share ('taxing meat would lead to less meat being produced and eaten / would solve these problems').
*Implicit premises:* The pro arguments rely on the tax solving the problems of meat production and consumption.

**Scheme:** `AS:other`, confidence **0.45**, only Other. Source: WAG user-defined scheme (EnableOther).
Alternatives: no scheme: a WAG question move on the efficacy premise of each pro (0.3: the best model of a bare denial, but a question is not an argument, and AS:practical_reasoning has no efficacy CQ (N5)); book p. 97 CQ5 'are other actions needed besides A?' (reading: the tax alone is not enough) (0.15: the claim says the tax 'would not resolve' the problems, not that it is insufficient on its own); NEW:ineffective_means ('A does not achieve G, so A should not be done') (0.1: not a Walton scheme; it is the negation of a premise, not a new inference).

**Role:** answer to CQ:likelihood of the pros' scheme, attack, locus premise: brings_about({A},{G}) on AS:consequences (edge exists), would_bring_about({A},{S1},{S2}) on AS:practical_reasoning (edge missing, N5). Confidence **0.5**. Real target: sibling pros 13292.108, .115, .118 (not the thesis). Kialo relation does **not** fit.
Alternatives: rebuttal of the thesis (enthymeme: 'a tax that does not solve the problems should not be imposed') (0.35).

**Bindings** (AS:other): P1 = ‘a meat tax would not resolve the problems of meat production and consumption’; C = ‘¬would_bring_about(tax meat, S1, less meat produced and eaten) of each pro’. *Supplied by the annotator:* C.

1. ‘a meat tax would not resolve the problems of meat production and consumption’

⇒ Doing ‘tax meat’ in the situation ‘meat production and consumption harm the environment’ would not lead to ‘less meat is produced and eaten’.

*Template fit:* The negated template reads well, but the premise only restates the conclusion: as an argument it is circular, which is why a question move fits better.

**NASCQ needed:** N5 efficacy question on practical reasoning: 'would the action really lead to S2?'; N13 CQ:additional_actions: 'are other actions needed besides A?'.
*N5 lets the denial be a question move on practical-reasoning pros. N13 is only for the 'not enough on its own' reading.*

**Pattern:** P5, con denying that the policy works ← question or undermining of the pros' means-end premise (bare denial).

**Problems:**
- Bare assertion: no reason given.
- Kialo attaches it to the thesis, but it targets a premise of each pro argument. One Kialo edge becomes several WAG moves, one per target argument.
- 'The problems' is vague: environment, health, animal welfare?
- Its Kialo children do not support it as stated. 13292.174 is a side effect, 13292.104 is an 'other means' objection, 13292.145 is an amended proposal.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.7 · con of 13292.1 · impact 2.2 (10 votes)

> A [sin tax](https://www.investopedia.com/terms/s/sin_tax.asp) on meat would make food more expensive, despite many people not being able to afford it.

**Reading.** A meat tax would raise food prices for people who already struggle to afford food. Making food less affordable for them is bad, so the tax should not be imposed.
*Implicit premises:* Meat is a significant part of the diet of people on low incomes. · Making food less affordable for the poor is a bad outcome (equity).

**Scheme:** `walton.yml negative_consequences`, confidence **0.8**, walton.yml / book scheme. Source: book p. 101, 'Argument from Negative Consequences'; walton.yml negative_consequences.
Alternatives: AS:values (V = fairness: a regressive tax) (0.1: the claim states a consequence, not a value); AS:other (0.1: only if N1 is rejected).

**Role:** rebuttal: an argument for ¬should_be_performed(tax meat). Confidence **0.75**.
Alternatives: answer to CQ:side_effect of each pro, attack (exception side_effects({A},{S1}) on AS:practical_reasoning; conclusion locus on AS:consequences) (0.25).

**Bindings** (AS:negative_consequences (N1)): A = ‘tax meat’; G = ‘food becoming more expensive for people who already cannot afford it’. *Supplied by the annotator:* none.

1. Doing ‘tax meat’ brings about ‘food becoming more expensive for people who already cannot afford it’.
2. ‘food becoming more expensive for people who already cannot afford it’ is bad.

⇒ ‘tax meat’ should not be done.

*Template fit:* Reads well.

**NASCQ needed:** N1 AS:negative_consequences (rebuttal of a policy by a bad consequence); N2 the three book CQs on AS:negative_consequences; N3 CQ:opposite_consequences (book CQ3 in the negative direction); N4 let negative/positive consequences back the side-effect and opposite-consequences CQs.

**Pattern:** P4, con of policy via a bad consequence ← argument from negative consequences, rebuttal.

**Problems:**
- The link only defines 'sin tax' (framing); it is not evidence.
- 'despite' is ungrammatical; read it as 'even though'.
- 'Food' overgeneralizes: the tax raises meat prices, not food prices in general. 13292.33 attacks exactly this.
- The most-voted claim (10 votes, mean impact 2.2 on 0-4).

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.9 · con of 13292.1 · impact 0.33 (3 votes)

> A tax on meat could actually reduce tax revenues by harming the trillion-dollar meat industry, which pays substantial taxes.

**Reading.** The tax would shrink a large industry that pays a lot of tax, so total tax revenue could fall. Losing revenue (and harming the industry) is bad, so the tax should not be imposed.
*Implicit premises:* Raising revenue is a goal of the tax, or losing revenue is bad. · The revenue lost from the shrinking industry exceeds what the new tax raises.

**Scheme:** `walton.yml negative_consequences`, confidence **0.7**, walton.yml / book scheme. Source: book p. 101; walton.yml negative_consequences.
Alternatives: attack on an implicit 'raises revenue' consequence (CQ:likelihood) (0.2: no pro claims revenue; the claim would attack a premise nobody made); walton.yml cause_to_effect (chain tax -> industry harm -> less tax paid) (0.1: the chain is the brings_about premise; the claim's point is that the outcome is bad).

**Role:** rebuttal: an argument for ¬should_be_performed(tax meat). Confidence **0.65**.
Alternatives: undermining of an unstated revenue goal; its child 13292.43 reads it that way ('revenue is not the primary reason') (0.3).

**Bindings** (AS:negative_consequences (N1)): A = ‘tax meat’; G = ‘lower total tax revenue, because the heavily taxed meat industry shrinks’. *Supplied by the annotator:* none.

1. Doing ‘tax meat’ brings about ‘lower total tax revenue, because the heavily taxed meat industry shrinks’.
2. ‘lower total tax revenue, because the heavily taxed meat industry shrinks’ is bad.

⇒ ‘tax meat’ should not be done.

*Template fit:* Reads well.

**NASCQ needed:** N1 AS:negative_consequences (rebuttal of a policy by a bad consequence); N2 the three book CQs on AS:negative_consequences.

**Pattern:** P4, con of policy via a bad consequence ← argument from negative consequences, rebuttal.

**Problems:**
- Hedged ('could').
- Two bad consequences in one claim: harm to the industry and lost revenue.
- It depends on an unstated goal (revenue) that no pro states.
- The 'trillion-dollar' figure is unsourced.
- Rated weak: mean impact 0.33 from 3 votes.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.108 · pro of 13292.1 · impact 3.0 (1 vote)

> Meat production and consumption are bad for the environment.

**Reading.** The claim only states the problem. The charitable completion: a tax would cut meat production and consumption, which would reduce environmental damage and promote environmental protection, so governments should tax meat.
*Implicit premises:* Taxing meat would lead to less meat being produced and eaten (S2). · Less meat production means less environmental damage (G). · Less environmental damage promotes environmental protection (V).

**Scheme:** `AS:practical_reasoning`, confidence **0.55**, existing WAG scheme. Source: NASCQ AS:practical_reasoning (Atkinson & Bench-Capon); book pp. 95-96.
Alternatives: AS:consequences (G = less environmental damage) (0.35: the claim states a circumstance, not a consequence; the consequence would be wholly implicit); walton.yml negative_value (0.1: concludes about retaining a goal, not about an action; and walton.yml's version has a wrong conclusion).

**Role:** new pro argument for the parent. Confidence **0.85**.

**Bindings** (AS:practical_reasoning): A = ‘tax meat’; S1 = ‘meat production and consumption harm the environment’; S2 = ‘less meat is produced and eaten’; G = ‘less environmental damage from meat’; V = ‘environmental protection’. *Supplied by the annotator:* A, S2, G, V.

1. The current situation is ‘meat production and consumption harm the environment’.
2. Doing ‘tax meat’ in the situation ‘meat production and consumption harm the environment’ would lead to ‘less meat is produced and eaten’.
3. The goal ‘less environmental damage from meat’ would be achieved in ‘less meat is produced and eaten’.
4. Achieving ‘less environmental damage from meat’ would promote the value ‘environmental protection’.

⇒ ‘tax meat’ should be done.

*Template fit:* Readable with the quotes. 'The current situation is <clause>' is clumsy, and 'achieved in <state>' is stilted.

**NASCQ needed:** N6 CQ:circumstances on practical reasoning's S1 premise.
*Nothing for the argument itself. Its Kialo children argue S1 (deforestation, emissions, soil for; 'less harm than other pollution' and 'can be produced sustainably' against). N6 is only needed if S1 is to be questioned rather than attacked.*

**Pattern:** P2, pro of policy stating the problem (S1) or the target state (S2/G) ← AS:practical_reasoning, the other premises implicit.

**Problems:**
- Only S1 is stated. Three of the four premises and the value are supplied by the annotator.
- 'Bad for the environment' is vague.
- Its own children argue the premise S1, not the inference.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.115 · pro of 13292.1 · impact 3.0 (1 vote)

> A tax on meat would combat unhealthy meat consumption.

**Reading.** The tax would reduce unhealthy (excessive) meat consumption, which is good for health, so governments should tax meat.
*Implicit premises:* Current meat consumption is unhealthily high (its child 13292.8 argues this). · Higher prices reduce consumption. · Better health is good.

**Scheme:** `AS:consequences`, confidence **0.5**, existing WAG scheme. Source: NASCQ AS:consequences; book p. 101.
Alternatives: AS:practical_reasoning (S1 = people eat unhealthy amounts of meat, S2 = less meat eaten, G = healthier diets, V = public health) (0.45: equally plausible. The claim states the consequence, so the simpler scheme was chosen; a guideline must decide (see P2/P3)).

**Role:** new pro argument for the parent. Confidence **0.85**.

**Bindings** (AS:consequences): A = ‘tax meat’; G = ‘less unhealthy meat consumption’. *Supplied by the annotator:* none.

1. Doing ‘tax meat’ brings about ‘less unhealthy meat consumption’.
2. ‘less unhealthy meat consumption’ is good.

⇒ ‘tax meat’ should be done.

*Template fit:* Reads well.

**NASCQ needed:** none.
*None. If a con of this pro offers a better means, N10 is needed, because CQ:other_means is asked only of practical reasoning.*

**Pattern:** P3, pro of policy stating a benefit ← AS:consequences.

**Problems:**
- 'Unhealthy meat consumption' presupposes S1, and it is ambiguous whether all meat consumption or only excess is meant.
- Consequences or practical reasoning is a coin toss without a guideline.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.118 · pro of 13292.1 · impact 2.0 (2 votes)

> Taxing meat would raise the consumer price to reflect the real costs of meat consumption.

**Reading.** The Pigouvian argument. Today's meat price leaves out costs borne by others (externalities, subsidies). A tax would raise the price to the real cost, and prices that reflect real costs are good (efficient, or fair: 'the polluter pays'). So tax.
*Implicit premises:* The consumer price of meat is below its real cost (S1; children 13292.11 and .218 argue this). · Prices that reflect real costs promote economic efficiency or fairness (V).

**Scheme:** `AS:practical_reasoning`, confidence **0.55**, existing WAG scheme. Source: NASCQ AS:practical_reasoning; book pp. 95-96.
Alternatives: AS:consequences (G = the consumer price reflects real costs) (0.4: also fits the wording; practical reasoning was chosen because the claim names a target state, and its children argue S1); AS:values (V = fairness, 'polluter pays') (0.05: no value is named).

**Role:** new pro argument for the parent. Confidence **0.85**.

**Bindings** (AS:practical_reasoning): A = ‘tax meat’; S1 = ‘the price of meat is below its real cost’; S2 = ‘a consumer price that reflects the real cost of meat’; G = ‘prices that reflect real costs’; V = ‘economic efficiency’. *Supplied by the annotator:* S1, V.

1. The current situation is ‘the price of meat is below its real cost’.
2. Doing ‘tax meat’ in the situation ‘the price of meat is below its real cost’ would lead to ‘a consumer price that reflects the real cost of meat’.
3. The goal ‘prices that reflect real costs’ would be achieved in ‘a consumer price that reflects the real cost of meat’.
4. Achieving ‘prices that reflect real costs’ would promote the value ‘economic efficiency’.

⇒ ‘tax meat’ should be done.

*Template fit:* S2 and G nearly coincide: the template forces a split the claim does not make. V is a guess.

**NASCQ needed:** N6 CQ:circumstances on practical reasoning's S1 premise.
*Nothing for the argument itself. N6 only if its S1 is to be questioned.*

**Pattern:** P2, pro of policy stating the problem (S1) or the target state (S2/G) ← AS:practical_reasoning, the other premises implicit.

**Problems:**
- The value is unstated and ambiguous: efficiency or fairness. That decides which value-based attacks apply.
- S1 is implicit, yet two of its three children back exactly that premise. The annotator must make S1 explicit, or those children have nothing to attach to.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### — Children of 13292.7 —

### 13292.25 · con of 13292.7 · impact 3.0 (5 votes)

> The additional burden could be as small as changing one's food consumption pattern.

**Reading.** The harm is minor: people can respond to the higher price by changing what they eat. So the burden is small, not an inability to afford food.
*Implicit premises:* Changing one's diet is a small cost. · Affordable substitutes exist.

**Scheme:** `AS:other`, confidence **0.45**, only Other. Source: WAG user-defined scheme.
Alternatives: NEW:minor_harm ('the consequence is small, so it is not bad') (0.2: a single instance; not a Walton scheme); AS:values (0.1: no value is invoked).

**Role:** undermines a premise of the parent: bad(G), 'the outcome is not (seriously) bad'. A direct undermining: the book's CQs (p. 102) have no question for this premise. Confidence **0.45**.
Alternatives: answer to CQ:likelihood on brings_about({A},{G}), attack: unaffordability will not happen, only a diet change (0.4).

**Bindings** (AS:other): P1 = ‘the extra cost can be avoided by changing one's diet’; P2 = ‘changing one's diet is a small burden’; C = ‘¬bad(food becoming more expensive for people who already cannot afford it)’. *Supplied by the annotator:* P2.

1. ‘the extra cost can be avoided by changing one's diet’
2. ‘changing one's diet is a small burden’

⇒ ‘food becoming more expensive for people who already cannot afford it’ is not bad.

*Template fit:* Readable, but the conclusion overstates the claim: it says the harm is small, not that it is no harm. A binary bad(G) cannot express degree.

**NASCQ needed:** N1 AS:negative_consequences (rebuttal of a policy by a bad consequence); N7 CQ:bad_outcome on the bad(G) premise.
*N1 is the parent's scheme. N7 is optional: a direct undermining needs no CQ.*

**Pattern:** P10, attack on a cited bad consequence's severity ← undermining of bad(G).

**Problems:**
- Hedged ('could be as small as').
- Ambiguous between severity and likelihood.
- 'The additional burden' is an anaphor: the claim cannot be read without the parent.
- Rated strong: 5 votes, mean impact 3.0.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.33 · con of 13292.7 · impact 2.8 (5 votes)

> Other sources of protein are currently cheaper than meat, providing alternatives to this added burden.

**Reading.** Cheaper protein sources exist, so people who cannot afford taxed meat can switch. The tax would therefore not make food unaffordable: only meat would become dearer.
*Implicit premises:* The substitutes are nutritionally adequate and available to the people concerned.

**Scheme:** `AS:other`, confidence **0.4**, only Other. Source: WAG user-defined scheme.
Alternatives: walton.yml cause_to_effect, via its exception interference(E1) (substitution interferes with price -> unaffordability) (0.25: only if the parent's brings_about premise is itself formalized as cause to effect); NEW:available_substitute (0.2: a single instance); book 'argument from alternatives' (pp. 106-107) (0.05: that scheme is a disjunctive practical inference toward an action).

**Role:** answer to CQ:likelihood of the parent's negative_consequences, attack, locus premise brings_about({A},{G}). Confidence **0.55**.
Alternatives: undermines bad(G): the burden is avoidable, so not bad (0.3).

**Bindings** (AS:other): P1 = ‘other protein sources are currently cheaper than meat’; P2 = ‘people who cannot afford taxed meat can switch to them’; C = ‘(answer) ¬brings_about(tax meat, food more expensive for people who cannot afford it)’. *Supplied by the annotator:* P2.

1. ‘other protein sources are currently cheaper than meat’
2. ‘people who cannot afford taxed meat can switch to them’

⇒ Doing ‘tax meat’ does not bring about ‘food becoming more expensive for people who already cannot afford it’.

*Template fit:* Reads well with the new negation template.

**NASCQ needed:** N1 AS:negative_consequences (rebuttal of a policy by a bad consequence); N2 the three book CQs on AS:negative_consequences.

**Pattern:** P9, attack on a cited bad consequence's likelihood (substitutes, mitigation) ← CQ:likelihood on brings_about(A,G).

**Problems:**
- Factual claim with no source.
- 'This added burden' is an anaphor.
- Protein is only part of 'food', and 13292.180 (iron) contests nutritional equivalence.
- Its Kialo 'pro' child 13292.56 reads more like an attack on it (the tax would hurt protein intake): relation noise one level down.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.180 · pro of 13292.7 · impact 3.0 (2 votes)

> Meat is the [easiest source of organic iron](https://health.clevelandclinic.org/how-to-add-more-iron-to-your-diet/) for one's diet. Any financial restrictions will only increase cases for anemia directly because of this.

*Also linked under:* 13292.121

**Reading.** Meat is the easiest source of iron. Making it dearer would reduce iron intake and increase anaemia: a further bad consequence of the tax.
*Implicit premises:* A meat tax reduces meat intake among those affected. · Those people do not replace the iron from other sources.

**Scheme:** `walton.yml negative_consequences`, confidence **0.65**, walton.yml / book scheme. Source: book p. 101.
Alternatives: AS:expert_opinion for sentence 1 (the Cleveland Clinic link) (0.2: a sub-argument for a premise, not the claim's own inference); walton.yml cause_to_effect (0.15: the inner causal link only).

**Role:** misplaced: a separate rebuttal of the thesis (a second bad consequence), not support for the parent's conclusion or premises. Confidence **0.45**. Real target: grandparent 13292.1 (thesis), not the parent. Kialo relation does **not** fit.
Alternatives: backs a premise of the parent: bad(G) (food becoming dearer is bad because it causes anaemia). No WAG or book scheme concludes bad(G), so it would be AS:other (0.35).

**Bindings** (AS:negative_consequences (N1)): A = ‘tax meat’; G = ‘more cases of anaemia’. *Supplied by the annotator:* none.

1. Doing ‘tax meat’ brings about ‘more cases of anaemia’.
2. ‘more cases of anaemia’ is bad.

⇒ ‘tax meat’ should not be done.

*Template fit:* Reads well; the factual premise about iron is not represented.

**NASCQ needed:** N1 AS:negative_consequences (rebuttal of a policy by a bad consequence).

**Pattern:** P11, a further bad consequence filed as a Kialo pro of a con ← separate negative-consequences rebuttal of the policy (re-parent).

**Problems:**
- Two arguments: a factual premise (with a source) and a causal consequence.
- The link carries the evidence for sentence 1, which is only an implicit premise of the scheme.
- Overstatement ('any', 'only', 'directly'); 'organic iron' presumably means haem iron.
- Multi-parent claim: it is also a pro of 13292.121. WAG has one argument per edge, so the same text becomes two arguments.
- Placement: Kialo files a new consequence as a pro of a con.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.36 · pro of 13292.7 · impact 1.5 (2 votes)

> In poor communities, people need the easiest access to all types of food. Taxes would add costs to food production and decrease the country's food security.

**Reading.** (i) People in poor communities need easy access to all foods (a need claim). (ii) The tax adds to food costs. (iii) It lowers the country's food security. As a pro of 13292.7: (i) says why dearer food is bad for the poor, (ii) supports the price effect, and (iii) extends the consequence.
*Implicit premises:* The tax raises food costs for poor communities.

**Scheme:** `AS:other`, confidence **0.4**, only Other. Source: WAG user-defined scheme.
Alternatives: walton.yml negative_consequences (G = lower food security) (0.35: fits sentence 2 read as a new consequence; then the role is a misplaced rebuttal, as in 13292.180); AS:values (V = access to food for the poor) (0.15: concludes 'reason to keep a goal', not 'the outcome is bad'); book 'argument from need for help' (pp. 108-109) (0.05: that scheme argues for helping, not against an action).

**Role:** backs premises of the parent: bad(G) (sentence 1) and brings_about({A},{G}) (sentence 2). Confidence **0.4**.
Alternatives: misplaced: a separate rebuttal of the thesis (lower food security) (0.35).

**Bindings** (AS:other): P1 = ‘people in poor communities need the easiest access to all types of food’; P2 = ‘a meat tax adds to the cost of food’; C = ‘food becoming more expensive for people who already cannot afford it is bad’. *Supplied by the annotator:* none.

1. ‘people in poor communities need the easiest access to all types of food’
2. ‘a meat tax adds to the cost of food’

⇒ ‘food becoming more expensive for people who already cannot afford it’ is bad.

*Template fit:* One Other argument cannot conclude two premises. Backing both bad(G) and brings_about(A,G) takes two moves.

**NASCQ needed:** N1 AS:negative_consequences (rebuttal of a policy by a bad consequence).
*N1 is the parent's scheme. Backing a premise directly needs no NASCQ edge, but no scheme other than Other concludes bad(G).*

**Pattern:** P15, Kialo pro of a negative-consequence con that says why the consequence is bad or likely ← backing of bad(G) / brings_about(A,G) with AS:other.

**Problems:**
- Two or three arguments in one claim.
- 'Taxes would add costs to food production' misdescribes a consumption tax.
- 'All types of food' versus a tax on meat only.
- Shifts from the poor to 'the country's food security'.
- Rated low: 1.5 from 2 votes.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.127 · pro of 13292.7 · impact 1.0 (1 vote)

> Starvation from decreased access to meat is a more severe threat than any consequence of high meat consumption.

**Reading.** Weighs consequences. The harm the tax would cause (starvation where meat becomes unaffordable) is worse than any harm of high meat consumption that the tax would reduce. So the parent's bad consequence outweighs the pros' benefits.
*Implicit premises:* The tax would cause starvation somewhere. · The benefits of the tax are (only) reductions in the harms of high meat consumption.

**Scheme:** `NEW:comparative_consequences`, confidence **0.4**, NEW scheme. Source: proposal N11: 'Doing A brings about B; B is worse than C, the harm A would remedy; so A should not be done.' No such scheme in book pp. i-135.
Alternatives: AS:values backing CQ:other_goals of the health pro 13292.115 (survival comes before health) (0.3: existing WAG edges, but it attaches to a different argument than Kialo's parent); AS:other (0.25: if N11 is rejected); walton.yml negative_practical (book p. 100, 'more serious consequences') (0.05: that scheme is about acting under ignorance).

**Role:** answer to the opposite-consequences CQ of the parent's negative_consequences (book p. 102 CQ3), defence, locus conclusion: yes there are good consequences, but they are outweighed. Confidence **0.4**.
Alternatives: answer to CQ:other_goals of 13292.115, attack, locus conclusion (a rebuttal), backed by AS:values (0.35); backs bad(G) of the parent (severity) (0.2).

**Bindings** (NEW:comparative_consequences (N11)): A = ‘tax meat’; B = ‘starvation where people lose access to meat’; C = ‘the harms of high meat consumption’. *Supplied by the annotator:* none.

1. Doing ‘tax meat’ brings about ‘starvation where people lose access to meat’.
2. ‘starvation where people lose access to meat’ is worse than ‘the harms of high meat consumption’.

⇒ ‘tax meat’ should not be done.

*Template fit:* Reads well (draft templates).

**NASCQ needed:** N1 AS:negative_consequences (rebuttal of a policy by a bad consequence); N3 CQ:opposite_consequences (book CQ3 in the negative direction); N11 NEW AS:comparative_consequences (weighing a bad consequence against the harm the action remedies).

**Pattern:** P12, weighing a bad consequence against the policy's benefits ← NEW comparative consequences (opposite-consequences CQ, defence), or AS:values on CQ:other_goals.

**Problems:**
- Hyperbole: the step from 'dearer' to 'starvation' is itself an unargued causal claim, which its children attack.
- 'Any consequence of high meat consumption' ignores the environmental pros, the thesis's strongest.
- Two placements are possible: a defence of 13292.7, or an attack on 13292.115. Kialo's single parent loses the second.
- Rated low: 1.0 from 1 vote.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### — Children of 13292.13 —

### 13292.14 · con of 13292.13 · impact 4.0 (1 vote)

> [Tariffs decrease consumption of goods](https://www.investopedia.com/articles/economics/08/tariff-trade-barrier-basics.asp), which would help decrease consumption of meats \(due to their higher price\).

**Reading.** Levies that raise prices reduce consumption (a sourced economic rule). A meat tax raises meat prices, so it would cut meat consumption, which helps solve the problems. So 13292.13 is false.
*Implicit premises:* A meat tax works on prices like a tariff. · Less meat consumption reduces the problems.

**Scheme:** `walton.yml cause_to_effect`, confidence **0.45**, walton.yml / book scheme. Source: walton.yml cause_to_effect; book ch. 5 p. 168, outside the pages available (unverified).
Alternatives: AS:expert_opinion (W = Investopedia, S = tariffs decrease consumption) (0.2: the link backs the generalization, a premise; the claim's own step is causal); AS:analogy (tariffs are like a meat tax) (0.2: the comparison is not drawn); walton.yml defeasible_modus_ponens (0.15: too generic).

**Role:** rebuttal of the parent: it argues the parent's negation, i.e. the pros' means-end premise. If the parent is modelled as a question, this is the defence answer to it (CQ:likelihood, locus premise). Confidence **0.6**.
Alternatives: answer to CQ:likelihood of the pros (defence, locus premise), attached to the pros rather than to 13292.13 (0.3).

**Bindings** (AS:cause_to_effect (N8; walton.yml templates)): E1 = ‘meat prices rise because of the tax’; E2 = ‘people eat less meat’. *Supplied by the annotator:* none.

1. An event ‘meat prices rise because of the tax’ has occurred.
2. Event ‘meat prices rise because of the tax’ causes event ‘people eat less meat’.

⇒ An event ‘people eat less meat’ will occur.  (As the answer: Doing ‘tax meat’ in the situation ‘…’ would lead to ‘less meat is produced and eaten’.)

*Template fit:* Poor: has_occurred is the wrong tense for a hypothetical policy (N8 redrafts it). The causal premise really states a generalization (higher prices -> lower demand), not the instance.

**NASCQ needed:** N8 AS:cause_to_effect, able to back CQ:likelihood; N5 efficacy question on practical reasoning: 'would the action really lead to S2?'.

**Pattern:** P6, defence of a policy's efficacy via a causal generalization ← cause to effect.

**Problems:**
- Tariffs are not taxes on domestic consumption: a category slip.
- The link carries the evidence for the generalization only.
- 'Help decrease consumption' is S2, not 'resolve the problems' (G); the step from less consumption to fewer problems is still implicit.
- In Kialo it attacks 13292.13; in WAG it defends a premise of the pros. The polarity inverts.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.61 · con of 13292.13 · no votes

> A [tax for tobacco](https://www.washingtonpost.com/national/health-science/cigarette-taxes-are-the-best-way-to-cut-smoking-scaring-big-tobacco/2017/10/21/fbf51d04-9f05-11e7-8ea1-ed975285475e_story.html?utm_term=.278c1ec80729) significantly reduced its consumption.

**Reading.** By analogy: taxing tobacco cut its consumption, and meat is a similar taxed consumer good, so a meat tax would cut meat consumption. Against 13292.13.
*Implicit premises:* Meat is relevantly similar to tobacco (price response, substitutes).

**Scheme:** `AS:analogy`, confidence **0.7**, existing WAG scheme. Source: NASCQ AS:analogy; book p. 56 (Version 1), CQs p. 62.
Alternatives: book 'practical argument from analogy' (p. 74: the right thing in S1 was A, S2 is similar, so...) (0.15: the claim is about effect, not rightness); AS:expert_opinion via the Washington Post link (0.1: backs the source-case premise only); argument from precedent (0.05: no rule or ruling is involved).

**Role:** rebuttal of the parent (argues the tax would work); equivalently a defence of the pros' means-end premise. Confidence **0.6**.
Alternatives: defence answer to the efficacy CQ on the pros (0.3).

**Bindings** (AS:analogy): C = ‘tobacco taxation’; A = ‘(answer variable, bound by the game) the negation of 13292.13: the meat tax would lead to less meat consumption’. *Supplied by the annotator:* none.

1. ‘tobacco taxation’ is a similar case.
2. “Doing ‘tax meat’ … would lead to ‘less meat is produced and eaten’” is true in the case of ‘tobacco taxation’.

⇒ Doing ‘tax meat’ in the situation ‘…’ would lead to ‘less meat is produced and eaten’.

*Template fit:* Breaks. The answer variable binds A to the claim about meat, so the source-case premise says something false about tobacco. What held for tobacco was 'taxing tobacco cut tobacco consumption' (N12).

**NASCQ needed:** N12 AS:analogy: separate the source-case proposition from the answer.
*No addition is needed for legality (AS:analogy is thesis-independent); N12 is needed for a correct reading.*

**Pattern:** P7, defence of a policy's efficacy via a similar past policy ← AS:analogy.

**Problems:**
- No comparison is stated: the reader supplies 'meat is like tobacco'.
- Relevant differences (addiction, price elasticity, substitutes) are the obvious attack (CQ:difference).
- The link carries the evidence for the source case.
- No votes.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.145 · con of 13292.13 · no votes

> Governments should put a tax on meat and use some of those tax revenues to encourage research into developing synthetic meats.

**Reading.** A counter-proposal that amends the thesis: tax meat and earmark part of the revenue for synthetic-meat research. As a con of 13292.13 it implies: even if the tax alone would not solve the problems, the tax plus funded alternatives would.
*Implicit premises:* Research funded by the tax would yield viable synthetic meat. · Synthetic meat would replace enough conventional meat to resolve the problems.

**Scheme:** `AS:practical_reasoning (for the compound action)`, confidence **0.4**, existing WAG scheme. Source: NASCQ AS:practical_reasoning.
Alternatives: AS:other (0.35: no inference is stated); AS:consequences (0.25: no consequence is stated either).

**Role:** none of the four roles fits: it amends the parent's action. The closest is a rebuttal of 13292.13 for the amended action; in Walton's terms, a concession to book p. 97 CQ5 ('are other actions needed besides A?') followed by a repair. Confidence **0.25**. Real target: a new action A' (tax plus earmarked research). Kialo relation does **not** fit.
Alternatives: a pro of the thesis: it entails 'governments should tax meat' (0.3).

**Bindings** (AS:practical_reasoning): A = ‘tax meat and fund synthetic-meat research from the revenue’; S1 = ‘a meat tax alone would not solve the problems of meat’; S2 = ‘viable synthetic meat replaces much conventional meat’; G = ‘the problems of meat production are solved’; V = ‘environmental protection’. *Supplied by the annotator:* S1, S2, G, V.

1. The current situation is ‘a meat tax alone would not solve the problems of meat’.
2. Doing ‘tax meat and fund synthetic-meat research from the revenue’ in the situation ‘…’ would lead to ‘viable synthetic meat replaces much conventional meat’.
3. The goal ‘the problems of meat production are solved’ would be achieved in ‘viable synthetic meat replaces much conventional meat’.
4. Achieving ‘the problems of meat production are solved’ would promote the value ‘environmental protection’.

⇒ ‘tax meat and fund synthetic-meat research from the revenue’ should be done.

*Template fit:* Reads, but every binding except A is invented, and the conclusion is not the thesis statement.

**NASCQ needed:** N13 CQ:additional_actions: 'are other actions needed besides A?'.
*N13 helps only partly. The real gap is in the data model: no move says that should_be_performed(A plus B) bears on should_be_performed(A).*

**Pattern:** P14, amended counter-proposal (A plus B) ← practical reasoning for a different action; no WAG move links it to A.

**Problems:**
- Restates the thesis and adds a clause; no reason is given.
- Only the Kialo relation (a con of a con) hints at its function.
- It changes the action under debate.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.174 · pro of 13292.13 · no votes

> Taxing meat might force producers to use cheaper and more dangerous means of producing meat.

**Reading.** A side effect: to keep costs down, producers would turn to cheaper and riskier production methods, a bad consequence of the tax. As a pro of 13292.13 it fits only loosely: it says the tax creates new problems, not that it fails to solve the old ones.
*Implicit premises:* Producers absorb the tax by cutting production costs. · Cheaper methods are more dangerous (to consumers, workers or animals).

**Scheme:** `walton.yml negative_consequences`, confidence **0.7**, walton.yml / book scheme. Source: book p. 101.
Alternatives: AS:other (0.15: only if N1 is rejected); book 'danger appeal' (p. 103) (0.1: it is not addressed to the hearer's own safety).

**Role:** misplaced: a rebuttal of the thesis (negative consequences), or an answer to CQ:side_effect of the pros (exception side_effects({A},{S1})); not support of 13292.13. Confidence **0.5**. Real target: grandparent 13292.1 (thesis), or CQ:side_effect of the pros. Kialo relation does **not** fit.
Alternatives: new pro argument for the parent: new problems mean the problems are not resolved (0.3).

**Bindings** (AS:negative_consequences (N1)): A = ‘tax meat’; G = ‘producers switching to cheaper and more dangerous production methods’. *Supplied by the annotator:* none.

1. Doing ‘tax meat’ brings about ‘producers switching to cheaper and more dangerous production methods’.
2. ‘producers switching to cheaper and more dangerous production methods’ is bad.

⇒ ‘tax meat’ should not be done.

*Template fit:* Reads well.

**NASCQ needed:** N1 AS:negative_consequences (rebuttal of a policy by a bad consequence); N4 let negative/positive consequences back the side-effect and opposite-consequences CQs.
*N4 is needed if it is attached to the pros as a side-effect answer.*

**Pattern:** P11, a further bad consequence filed as a Kialo pro of a con ← separate negative-consequences rebuttal of the policy (re-parent).

**Problems:**
- Hedged ('might'); 'force' overstates.
- No reason or source for the causal claim.
- Who is endangered is unspecified.
- The placement does not match the parent.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.104 · pro of 13292.13 · impact 0.0 (1 vote)

> Meat is not inherently problematic. What is problematic is the [animal farming industry](https://www.theguardian.com/commentisfree/2018/aug/25/veganism-intensively-farmed-meat-dairy-soya-maize), so the appropriate solution is to regulate this industry rather than penalize those who wish to consume meat.

**Reading.** The harms come from industrial animal farming, not from meat as such. So the right means is to regulate the industry rather than tax consumers, which would also spare consumers' freedom.
*Implicit premises:* Regulating the industry would solve the problems at least as well as a tax. · Penalizing consumers is unfair or illiberal.

**Scheme:** `AS:consequences (for the alternative action, as backup of CQ:other_means)`, confidence **0.4**, existing WAG scheme. Source: NASCQ AS:consequences; backup edge AS:consequences -> CQ:other_means exists.
Alternatives: AS:practical_reasoning for 'regulate the industry' (0.25: no backup edge to CQ:other_means); AS:values (V = consumer freedom: 'rather than penalize') (0.15: covers the last clause only); book verbal classification (p. 68) for sentence 1 (0.1: sentence 1 only); AS:other (0.1: fallback).

**Role:** misplaced: answer to CQ:other_means of the thesis's pros, attack, locus exception realize_more_effectively({G},{A}). Confidence **0.5**. Real target: the thesis's practical-reasoning pros (13292.108, .118), not the parent. Kialo relation does **not** fit.
Alternatives: new pro argument for the parent: a consumer tax targets the wrong thing, so it will not resolve the problems (0.35).

**Bindings** (AS:consequences (as backup of CQ:other_means on 13292.108)): A = ‘regulate the animal farming industry’; G = ‘less environmental damage from meat’. *Supplied by the annotator:* G (taken from the target pro).

1. Doing ‘regulate the animal farming industry’ brings about ‘less environmental damage from meat’.
2. ‘less environmental damage from meat’ is good.

⇒ Another action would achieve ‘less environmental damage from meat’ more effectively than ‘tax meat’.

*Template fit:* Reads well. The 'is good' premise is redundant, since the pro already grants it.

**NASCQ needed:** N10 CQ:other_means on AS:consequences.
*Nothing if the pros are practical reasoning. N10 if a pro is AS:consequences.*

**Pattern:** P8, con of policy via a better alternative action ← answer to CQ:other_means / CQ:best_means (exception locus).

**Problems:**
- Three claims and a value clause in one: meat is not the problem; the industry is; regulate instead; do not penalize consumers.
- The link (an op-ed) supports sentence 2.
- Placement: it attacks the thesis's pros but is filed as a pro of 13292.13. Its child 13292.208 attacks sentence 1 only.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### — Children of 13292.118 —

### 13292.169 · con of 13292.118 · no votes

> Taxing meat would be an ineffective means to achieve this objective.

**Reading.** A tax would not (efficiently) make the price reflect real costs. As written it is a bare denial of the means-end premise. With its child 13292.171 ('easier to remove subsidies') it becomes a better-means objection.
*Implicit premises:* (from the child) Removing subsidies would bring the price to its real cost more easily.

**Scheme:** `AS:other`, confidence **0.45**, only Other. Source: WAG user-defined scheme.
Alternatives: AS:consequences for 'remove meat subsidies', backing CQ:best_means / CQ:other_means (0.3: only with the child's content; the claim names no alternative); NEW:ineffective_means (0.15: see 13292.13).

**Role:** undermines would_bring_about({A},{S1},{S2}) of the parent: the efficacy question, which needs N5. Confidence **0.4**.
Alternatives: answer to CQ:best_means, attack, locus exception bring_about_more_effectively({S2},{A}) (the reading the child suggests) (0.4); answer to CQ:other_means (0.2).

**Bindings** (AS:other): P1 = ‘taxing meat would be an ineffective means to make the consumer price reflect real costs’; C = ‘(answer) ¬would_bring_about(tax meat, price below real cost, price reflects real cost)’. *Supplied by the annotator:* none.

1. ‘taxing meat would be an ineffective means to make the consumer price reflect real costs’

⇒ Doing ‘tax meat’ in the situation ‘the price of meat is below its real cost’ would not lead to ‘a consumer price that reflects the real cost of meat’.

*Template fit:* Reads well, but it is circular as an argument: better a question move (efficacy reading). The best-means reading cannot be a question (rule Q1: an exception must be argued).

**NASCQ needed:** N5 efficacy question on practical reasoning: 'would the action really lead to S2?'; N9 AS:consequences backs CQ:best_means.
*N5 for the efficacy reading, N9 for the best-means reading.*

**Pattern:** P5, con denying that the policy works ← question or undermining of the pros' means-end premise (bare denial).

**Problems:**
- Bare assertion; 'this objective' is an anaphor.
- The reading depends on whether the claim's own children may be used: a guideline decision.
- 'Ineffective' (does not work) versus 'less effective' (another means is better) are different CQs and loci.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.218 · pro of 13292.118 · impact 4.0 (1 vote)

> It would factor in [externalities](https://en.wikipedia.org/wiki/Externality) – such as costs to society due to climate change caused from its greenhouse gas emissions or its water pollution and infectious disease risks and so on.

**Reading.** Spells out the 'real costs': the tax would add external costs (climate change, water pollution, disease risk) to the price. It backs the parent's implicit S1 (today's price leaves these costs out) and its means-end premise (the tax would put them in).
*Implicit premises:* These external costs are not in the current price.

**Scheme:** `AS:other`, confidence **0.4**, only Other. Source: WAG user-defined scheme.
Alternatives: book verbal classification (p. 68): these costs fall under 'real costs' (0.25: the classification is presupposed, not argued); AS:expert_opinion (via its own child 13292.224, a 2020 Nature analysis) (0.15: that is the child's scheme, not this claim's); walton.yml from_example (0.1: the examples illustrate a category; they do not generalize); walton.yml cause_to_effect (0.1: no causal step is claimed).

**Role:** backs premises of the parent: would_bring_about({A},{S1},{S2}) and the implicit current_circumstances({S1}). Confidence **0.5**.
Alternatives: a clarification of the parent with no argumentative role (0.3).

**Bindings** (AS:other): P1 = ‘meat has external costs borne by society: climate change, water pollution, infectious-disease risk’; P2 = ‘a meat tax would add these external costs to the consumer price’; C = ‘would_bring_about(tax meat, price below real cost, price reflects real cost)’. *Supplied by the annotator:* none.

1. ‘meat has external costs borne by society: climate change, water pollution, infectious-disease risk’
2. ‘a meat tax would add these external costs to the consumer price’

⇒ Doing ‘tax meat’ in the situation ‘the price of meat is below its real cost’ would lead to ‘a consumer price that reflects the real cost of meat’.

*Template fit:* Reads well.

**NASCQ needed:** N6 CQ:circumstances on practical reasoning's S1 premise.
*A direct premise defence needs nothing; N6 only for question moves on S1.*

**Pattern:** P13, backing the circumstances or means-end premise with facts ← direct premise defence (AS:other; the link carries the evidence).

**Problems:**
- A list of examples ('and so on').
- The link is a definition (Wikipedia, 'Externality'), not evidence.
- The premise it backs (S1) is implicit in the parent, so the parent's annotator must have bound S1.
- Overlaps its sibling 13292.11: two different gaps between price and real cost (externalities, subsidies).

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

### 13292.11 · pro of 13292.118 · no votes

> The price of meat does not reflect its true cost due to [subsidies](https://www.perc.org/2011/02/14/big-meat-and-big-government/).

**Reading.** Subsidies push the price of meat below its true cost, so the current price does not reflect it. This backs the parent's implicit circumstance S1.
*Implicit premises:* Meat production is subsidized (the link is offered as evidence).

**Scheme:** `AS:other`, confidence **0.35**, only Other. Source: WAG user-defined scheme.
Alternatives: walton.yml cause_to_effect (subsidies cause lower prices; its child 13292.167 says exactly this) (0.3: the claim reports a present state and its cause; it does not predict an effect); AS:expert_opinion / AS:position_to_know (W = the linked PERC article) (0.2: the link carries the evidence, but the claim does not appeal to the source); AS:sign (0.05: no sign is observed).

**Role:** backs a premise of the parent: current_circumstances({S1}), 'the price of meat is below its real cost'. Confidence **0.65**.
Alternatives: double-edged: its content supports removing subsidies instead (CQ:other_means against the parent), as its attack 13292.123 says (0.2).

**Bindings** (AS:other): P1 = ‘meat production is subsidized’; P2 = ‘subsidies lower the price of meat below its true cost’; C = ‘current_circumstances(the price of meat is below its real cost)’. *Supplied by the annotator:* none.

1. ‘meat production is subsidized’
2. ‘subsidies lower the price of meat below its true cost’

⇒ The current situation is ‘the price of meat is below its real cost’.

*Template fit:* Reads well.

**NASCQ needed:** N6 CQ:circumstances on practical reasoning's S1 premise; N8 AS:cause_to_effect, able to back CQ:likelihood.
*N6 only for question moves; N8 if cause to effect is chosen.*

**Pattern:** P13, backing the circumstances or means-end premise with facts ← direct premise defence (AS:other; the link carries the evidence).

**Problems:**
- S1 and its cause in one sentence.
- The link carries the evidence.
- The cause it names (subsidies) argues as much for removing subsidies as for a tax.
- The premise it backs is implicit in the parent.

**Decision:** ☐ approve ☐ reject ☐ change: ________

---

## Patterns

Per instance, joint = min(scheme confidence, role confidence): the chance that an annotator matches both if the two choices go together (an upper bound). The product, the value if they were independent, is given as the lower bound. Pattern confidence = mean over its instances. With n <= 2 per pattern these are anecdotes, not estimates.

| id | pattern (what to approve) | instances | n | confidence (lower) | NASCQ additions |
|---|---|---|---|---|---|
| P1 | thesis: policy action ← claim should_be_performed(A); pros by practical reasoning or positive consequences, cons by negative consequences | 13292.1 | 1 | 0.9 (0.9) | yes: N1 |
| P2 | pro of policy stating the problem (S1) or the target state (S2/G) ← AS:practical_reasoning, the other premises implicit | 13292.108, 13292.118 | 2 | 0.55 (0.47) | none required |
| P3 | pro of policy stating a benefit ← AS:consequences | 13292.115 | 1 | 0.5 (0.42) | none required; conditional: N10 (only if this pattern is kept) |
| P4 | con of policy via a bad consequence ← argument from negative consequences, rebuttal | 13292.7, 13292.9 | 2 | 0.7 (0.53) | yes: N1, N2, N3, N4 |
| P5 | con denying that the policy works ← question or undermining of the pros' means-end premise (bare denial) | 13292.13, 13292.169 | 2 | 0.43 (0.21) | yes: N5; conditional: N13 (optional) |
| P6 | defence of a policy's efficacy via a causal generalization ← cause to effect | 13292.14 | 1 | 0.45 (0.27) | yes: N8, N5 |
| P7 | defence of a policy's efficacy via a similar past policy ← AS:analogy | 13292.61 | 1 | 0.6 (0.42) | yes: N12 |
| P8 | con of policy via a better alternative action ← answer to CQ:other_means / CQ:best_means (exception locus) | 13292.104 | 1 | 0.4 (0.2) | none required; conditional: N9 (best-means variant), N10 (if the pro is AS:consequences) |
| P9 | attack on a cited bad consequence's likelihood (substitutes, mitigation) ← CQ:likelihood on brings_about(A,G) | 13292.33 | 1 | 0.4 (0.22) | yes: N1, N2 |
| P10 | attack on a cited bad consequence's severity ← undermining of bad(G) | 13292.25 | 1 | 0.45 (0.2) | yes: N1; conditional: N7 (optional) |
| P11 | a further bad consequence filed as a Kialo pro of a con ← separate negative-consequences rebuttal of the policy (re-parent) | 13292.180, 13292.174 | 2 | 0.47 (0.32) | yes: N1 |
| P12 | weighing a bad consequence against the policy's benefits ← NEW comparative consequences (opposite-consequences CQ, defence), or AS:values on CQ:other_goals | 13292.127 | 1 | 0.4 (0.16) | yes: N3, N11 |
| P13 | backing the circumstances or means-end premise with facts ← direct premise defence (AS:other; the link carries the evidence) | 13292.218, 13292.11 | 2 | 0.38 (0.22) | none required; conditional: N6 (only for question moves) |
| P14 | amended counter-proposal (A plus B) ← practical reasoning for a different action; no WAG move links it to A | 13292.145 | 1 | 0.25 (0.1) | none required; conditional: N13 (partly), a data-model change (a move that amends the action), outside the NASCQ |
| P15 | Kialo pro of a negative-consequence con that says why the consequence is bad or likely ← backing of bad(G) / brings_about(A,G) with AS:other | 13292.36 | 1 | 0.4 (0.16) | yes: N1 |

## NASCQ proposals

Each is a concrete proposal for `schemes.json`, `critical_questions.json`, `edges.json` or `core.yml`. Sources are given; “from memory” marks what I could not check against the available pages.

**N1 · new_scheme: AS:negative_consequences (rebuttal of a policy by a bad consequence).** Source: Walton, Reed & Macagno 2008, p. 101 (scheme), p. 102 (CQs); walton.yml id negative_consequences (inherited from upstream Carneades, which loads). Used by 9 item(s).
*Check:* No core scheme concludes a negated statement today. Verify that a support move whose conclusion is ¬should_be_performed(A) joins the issue on the thesis (adapter.findOrCreateStatement keeps ¬ ids as raw keys).
```json
{
 "nascq": {
  "AS:negative_consequences": {
   "carneades_id": "negative_consequences",
   "name": "Argument from Negative Consequences",
   "semantic_type": "action",
   "thesis_independent": false,
   "base_weight": 0.65,
   "variables": {
    "A": {
     "type": "action",
     "description": "Action",
     "example_en": "introduce a carbon tax"
    },
    "G": {
     "type": "event",
     "description": "Bad consequence",
     "example_en": "higher heating bills for poor households"
    }
   },
   "premise_templates_en": [
    "Doing {A} brings about {G}",
    "{G} is bad"
   ],
   "conclusion_template_en": "{A} should not be done",
   "evidence": "recommended",
   "evidence_note_en": "a source showing the action brings about the consequence"
  }
 },
 "core_yml": {
  "scheme": {
   "id": "negative_consequences",
   "variables": [
    "A",
    "G"
   ],
   "conclusions": [
    "¬should_be_performed(A)"
   ],
   "premises": [
    "brings_about(A,G)",
    "bad(G)"
   ]
  },
  "language": {
   "bad/1": "%s is bad.",
   "¬bad/1": "%s is not bad."
  }
 }
}
```

**N2 · ask_edges: the three book CQs on AS:negative_consequences.** Source: book p. 102, CQ1-CQ3 (the same CQs as positive consequences). Used by 3 item(s).
```json
{
 "nascq": [
  {
   "cq": "CQ:likelihood",
   "scheme": "AS:negative_consequences",
   "locus": "premise",
   "pattern": "brings_about({A},{G})"
  },
  {
   "cq": "CQ:evidence",
   "scheme": "AS:negative_consequences",
   "locus": "argument"
  },
  {
   "cq": "CQ:opposite_consequences",
   "scheme": "AS:negative_consequences",
   "locus": "conclusion"
  }
 ]
}
```

**N3 · new_cq: CQ:opposite_consequences (book CQ3 in the negative direction).** Source: book p. 102 CQ3: 'Are there other opposite consequences (bad as opposed to good, for example) that should be taken into account?'. Used by 2 item(s).
*Note:* WAG already maps the same book CQ to CQ:side_effect on AS:consequences (locus conclusion). The alternative is one neutral CQ node for both directions.
```json
{
 "nascq": {
  "CQ:opposite_consequences": {
   "name": "Opposite Consequences Question",
   "semantic_type": "value",
   "question_template_en": "Would doing {A} also have good consequences that outweigh {G}?",
   "caes_mapping": {
    "positive_answer_type": "support",
    "negative_answer_type": "rebuttal",
    "target": "conclusion"
   }
  }
 }
}
```

**N4 · backup_edges: let negative/positive consequences back the side-effect and opposite-consequences CQs.** Source: design choice (Pyon & Ri backup edges); not in the book. Used by 2 item(s).
*Note:* Today CQ:side_effect is backed only by AS:sign and the thesis-independent schemes.
```json
{
 "nascq": [
  {
   "scheme": "AS:negative_consequences",
   "cq": "CQ:side_effect"
  },
  {
   "scheme": "AS:consequences",
   "cq": "CQ:opposite_consequences"
  },
  {
   "scheme": "AS:practical_reasoning",
   "cq": "CQ:opposite_consequences"
  }
 ]
}
```

**N5 · ask_edge: efficacy question on practical reasoning: 'would the action really lead to S2?'.** Source: NOT among the five CQs of book p. 96. It is the Atkinson & Bench-Capon value-based CQ 'does the action have the stated consequences?' (cited from memory, not verified here). Used by 3 item(s).
*Note:* A direct undermining needs no edge. The edge is needed for a question move, which is the natural model of a bare denial.
```json
{
 "nascq": {
  "cq": "CQ:likelihood",
  "scheme": "AS:practical_reasoning",
  "locus": "premise",
  "pattern": "would_bring_about({A},{S1},{S2})"
 }
}
```

**N6 · new_cq+ask_edge: CQ:circumstances on practical reasoning's S1 premise.** Source: Atkinson-style CQ 'are the believed circumstances true?' (from memory, unverified); not in book p. 96. Used by 4 item(s).
*Note:* Only needed for question moves; direct backing or undermining of S1 works without it.
```json
{
 "nascq": {
  "CQ:circumstances": {
   "name": "Circumstances Question",
   "semantic_type": "truth",
   "question_template_en": "Is the current situation really {S1}?",
   "caes_mapping": {
    "positive_answer_type": "support",
    "negative_answer_type": "undermining",
    "target": "premise",
    "target_premise_pattern": "current_circumstances({S1})"
   }
  },
  "ask_edge": {
   "cq": "CQ:circumstances",
   "scheme": "AS:practical_reasoning",
   "locus": "premise",
   "pattern": "current_circumstances({S1})"
  }
 }
}
```

**N7 · new_cq+ask_edge (optional): CQ:bad_outcome on the bad(G) premise.** Source: no book CQ; premise 2 of the scheme on p. 101. Used by 1 item(s).
```json
{
 "nascq": {
  "CQ:bad_outcome": {
   "name": "Bad Outcome Question",
   "semantic_type": "value",
   "question_template_en": "Is {G} really bad?",
   "caes_mapping": {
    "positive_answer_type": "support",
    "negative_answer_type": "undermining",
    "target": "premise",
    "target_premise_pattern": "bad({G})"
   }
  },
  "ask_edge": {
   "cq": "CQ:bad_outcome",
   "scheme": "AS:negative_consequences",
   "locus": "premise",
   "pattern": "bad({G})"
  }
 }
}
```

**N8 · new_scheme+ask_edges+backup_edge: AS:cause_to_effect, able to back CQ:likelihood.** Source: walton.yml id cause_to_effect (upstream; loads); book ch. 5 p. 168 is outside the pages available, so the wording is unverified; CQs from cq_drafts.json (SPR-derived, unverified). Used by 2 item(s).
*Note:* walton.yml's premise has_occurred(E1) has the wrong tense for a prediction about a policy; the template must allow 'would occur'.
```json
{
 "nascq": {
  "AS:cause_to_effect": {
   "carneades_id": "cause_to_effect",
   "name": "Argument from Cause to Effect",
   "semantic_type": "causality",
   "thesis_independent": false,
   "base_weight": 0.6,
   "variables": {
    "E1": {
     "type": "event",
     "description": "Cause"
    },
    "E2": {
     "type": "event",
     "description": "Effect"
    }
   },
   "premise_templates_en": [
    "{E1} occurs or would occur",
    "Generally, {E1} causes {E2}"
   ],
   "conclusion_template_en": "{E2} will occur"
  },
  "new_cqs": {
   "CQ:causal_strength": "How strong is the causal generalization that {E1} causes {E2}?",
   "CQ:interference": "Could other causal factors interfere with {E1} producing {E2}?"
  },
  "ask_edges": [
   {
    "cq": "CQ:causal_strength",
    "scheme": "AS:cause_to_effect",
    "locus": "premise",
    "pattern": "causes({E1},{E2})"
   },
   {
    "cq": "CQ:interference",
    "scheme": "AS:cause_to_effect",
    "locus": "exception",
    "pattern": "interference({E1})"
   }
  ],
  "backup_edge": {
   "scheme": "AS:cause_to_effect",
   "cq": "CQ:likelihood"
  }
 }
}
```

**N9 · backup_edge: AS:consequences backs CQ:best_means.** Source: design choice; mirrors the existing AS:consequences -> CQ:other_means edge. Used by 1 item(s).
*Note:* Today only thesis-independent schemes can back CQ:best_means.
```json
{
 "nascq": {
  "scheme": "AS:consequences",
  "cq": "CQ:best_means"
 }
}
```

**N10 · ask_edge (conditional): CQ:other_means on AS:consequences.** Source: NOT in the book's CQs for consequences (p. 102); needed only if pros of a policy are annotated as AS:consequences. Used by 1 item(s).
*Note:* The alternative is the guideline 'annotate pros of a policy as AS:practical_reasoning', which already has CQ:other_means.
```json
{
 "nascq": {
  "cq": "CQ:other_means",
  "scheme": "AS:consequences",
  "locus": "argument"
 }
}
```

**N11 · new_scheme (NEW): NEW AS:comparative_consequences (weighing a bad consequence against the harm the action remedies).** Source: none in book pp. i-135; the closest is the 'more serious consequences' premise of negative practical reasoning (p. 100), which is about acting under ignorance. Used by 1 item(s).
```json
{
 "nascq": {
  "AS:comparative_consequences": {
   "name": "Argument from Comparative Consequences",
   "semantic_type": "value",
   "thesis_independent": false,
   "base_weight": 0.55,
   "variables": {
    "A": {
     "type": "action",
     "description": "Action"
    },
    "B": {
     "type": "event",
     "description": "Bad consequence of doing A"
    },
    "C": {
     "type": "event",
     "description": "Bad state that doing A would remedy"
    }
   },
   "premise_templates_en": [
    "Doing {A} brings about {B}",
    "{B} is worse than {C}"
   ],
   "conclusion_template_en": "{A} should not be done"
  },
  "core_yml": {
   "premises": [
    "brings_about(A,B)",
    "worse(B,C)"
   ],
   "conclusions": [
    "¬should_be_performed(A)"
   ]
  },
  "ask_edges": [
   {
    "cq": "CQ:likelihood",
    "scheme": "AS:comparative_consequences",
    "locus": "premise",
    "pattern": "brings_about({A},{B})"
   },
   {
    "cq": "CQ:comparison",
    "scheme": "AS:comparative_consequences",
    "locus": "premise",
    "pattern": "worse({B},{C})",
    "new_cq_template_en": "Is {B} really worse than {C}?"
   }
  ],
  "backup_edges": [
   {
    "scheme": "AS:comparative_consequences",
    "cq": "CQ:opposite_consequences"
   },
   {
    "scheme": "AS:comparative_consequences",
    "cq": "CQ:other_goals"
   }
  ]
 }
}
```

**N12 · scheme_change: AS:analogy: separate the source-case proposition from the answer.** Source: book p. 56 (Version 1: 'A is true in C1, so A is true in C2'), CQs p. 62. Used by 1 item(s).
*Note:* With answer_variable A bound to the answer about meat, the premise 'A is true in the case of tobacco taxation' says something false about tobacco. Owner decision: it changes a scheme WAG already uses.
```json
{
 "nascq": {
  "AS:analogy": {
   "variables": {
    "A1": {
     "type": "proposition",
     "description": "What held in the similar case"
    }
   },
   "premise_templates_en": [
    "{C} is a similar case",
    "{A1} in the case of {C}"
   ]
  },
  "core_yml": {
   "premises": [
    "similar_case(C)",
    "in_case(A1,C)"
   ]
  }
 }
}
```

**N13 · new_cq+exception (optional): CQ:additional_actions: 'are other actions needed besides A?'.** Source: book p. 97 CQ5 of two-person practical reasoning ('Are other actions, in addition to B, required for X to bring about A?'); walton.yml practical_reasoning2 exception intervening_actions(A,G). Used by 2 item(s).
```json
{
 "nascq": {
  "CQ:additional_actions": {
   "name": "Additional Actions Question",
   "semantic_type": "action",
   "question_template_en": "Would other actions besides {A} be needed to achieve {G}?",
   "caes_mapping": {
    "positive_answer_type": "support",
    "negative_answer_type": "undercutter",
    "target": "argument"
   }
  },
  "core_yml": {
   "practical_reasoning1": {
    "exceptions_add": [
     "intervening_actions(A,G)"
    ]
   },
   "language": {
    "intervening_actions/2": "Other actions besides %s would be needed to achieve %s."
   }
  },
  "ask_edge": {
   "cq": "CQ:additional_actions",
   "scheme": "AS:practical_reasoning",
   "locus": "exception",
   "pattern": "intervening_actions({A},{G})"
  }
 }
}
```

## Statistics

| | count | claims |
|---|---|---|
| existing WAG scheme | 6 | 13292.108, 13292.115, 13292.118, 13292.61, 13292.145, 13292.104 |
| walton.yml or book scheme | 5 | 13292.7, 13292.9, 13292.180, 13292.14, 13292.174 |
| NEW scheme | 1 | 13292.127 |
| only Other | 7 | 13292.13, 13292.25, 13292.33, 13292.36, 13292.169, 13292.218, 13292.11 |

Arguments whose primary scheme is not in WAG but which have an existing WAG scheme as an alternative (confidence ≥ 0.2): 13292.180, 13292.127, 13292.14, 13292.169, 13292.11.

| role | count |
|---|---|
| cq_answer: answer to a CQ of a scheme (the parent's, or a sibling's when misplaced) | 5 |
| rebuttal: argument for the negation of the parent (or of the ancestor named in target) | 6 |
| new_pro: plain new argument for the parent | 3 |
| premise_attack: undermines a premise of the parent directly, with no CQ | 1 |
| premise_backing: backs a premise of the parent directly | 3 |
| none: no role of the vocabulary fits | 1 |

One level down, Kialo relation against role:

- Kialo **con**: premise_attack 1, cq_answer 2, rebuttal 2, none (misplaced) 1
- Kialo **pro**: rebuttal (misplaced) 2, premise_backing 3, cq_answer 1, cq_answer (misplaced) 1

Mean confidence over the 19 arguments: scheme 0.51, role 0.56, joint 0.3 (product) to 0.47 (min). Mean scheme confidence by fit: existing WAG scheme 0.52, walton.yml / book scheme 0.66, NEW scheme 0.4, only Other 0.41.

Bindings supplied by the annotator: mean 0.74 per argument, at most 4.

Problem tags: implicit_premises 14, kialo_relation_mismatch 5, link_evidence 5, multi_argument 4, vague_scope 3, bare_assertion 3, hedged 3, unsourced_fact 3, anaphora 3, link_definition 2, hyperbole 2, template_misfit 2, multi_parent 1.

## Lessons for the process

### Decide once, per pattern (owner)

- P2 vs P3: one rule for the pros of a policy. Option 1: always AS:practical_reasoning. It keeps CQ:other_means, CQ:side_effect and CQ:possibility open to the con side, but needs two to four bindings the annotator invents. Option 2: AS:consequences when only a consequence is stated. Fewer invented premises, but no other-means question without N10. In this pilot the two were close for all three pros (0.55 vs 0.35, 0.5 vs 0.45, 0.55 vs 0.4).
- P4: a con of a policy is a rebuttal by negative consequences (one Kialo edge, one argument, weighed against the pros), not an undercutter of every pro through CQ:side_effect. Recommended: the rebuttal. It matches Kialo's weighing of pros against cons, and WAG already gives CQ:side_effect a conclusion locus on AS:consequences.
- P5: a bare denial of a premise becomes a question move (the proponent must then back the premise), not an Other argument whose premise restates its conclusion.
- Misplacement (P11, P8, P5): may the WAG graph re-parent a claim, to the grandparent or to the sibling pros it really attacks? 5 of 19 arguments act on something other than their Kialo parent. If re-parenting is refused, they can only be Other arguments against the parent, which misstates them.
- P13: may a child back a premise that the parent's annotator had to make explicit (S1 of 13292.118)? This requires the parent's bindings to be fixed before its children are annotated, and frozen afterwards.
- Which NASCQ additions come first. N1-N4 (negative consequences, its CQs and backup edges) are needed by 8 of 19 arguments, as their own scheme or their parent's. N5 (the efficacy question) is needed by 3 more. N11 (comparative consequences) and N12 (an analogy change) each rest on a single instance: defer them.

### Decide per instance (annotator)

- The reading and its implicit premises: which consequence, which value (13292.118: efficiency or fairness?).
- Which premise a child attacks when the text allows two: severity or likelihood (13292.25), efficacy or a better means (13292.169).
- Whether a claim holds several arguments, and which one is primary.
- Whether the Kialo relation fits the content, and if not, what the real target is.
- How to phrase the bindings: A as a verb phrase ('tax meat'), states as clauses or noun phrases.

### What the annotation guidelines must say

- The unit: one Kialo claim, read against its parent's formalization. Resolve anaphors ('this burden', 'this objective') from the parent. Say whether a claim's own children may disambiguate it (13292.169). Recommended: annotate the literal claim and record the child-informed reading as an alternative.
- Implicit premises: list them, mark the bindings the annotator supplied, add only what the scheme needs, and never strengthen the claim. Here a mean of 0.74 bindings per argument (max 4) were supplied by the annotator.
- Several arguments in one claim: one primary argument per item, the others as sub-items (e.g. 13292.36#2), because a WAG move is one argument.
- Links: classify each as evidence (then a source-based sub-argument, usually AS:expert_opinion, for the premise it supports), as a definition or framing (ignore it), or as neither. Do not fetch links during annotation.
- Record the role twice: relative to the Kialo parent, and relative to the real target (an ancestor, the sibling pros, a premise) when they differ.
- Polarity: Kialo pro/con is relative to the parent's statement; WAG attack/defence is relative to the target argument. Under a con that is itself a premise attack, a Kialo con is a defence of that premise (13292.14, 13292.61). Work the polarity out; do not copy it.
- Hedges ('could', 'might'): keep the scheme, and note the hedge as exposure to CQ:likelihood.
- The role vocabulary needs more than the four roles of the brief. It needs 'undermines a premise of the parent directly (no CQ)', 'acts on an ancestor or on sibling arguments (misplaced)' and 'none' (amendments of the action).
- The confidence scale: 0.8 or more when one scheme is obvious; about 0.5 when two are plausible; 0.3 or less when forced. These are self-estimates until double annotation measures agreement.

### Where the WAG data model gets in the way

- Cons of the thesis against CQs of arguments. A Kialo con of the thesis that attacks a premise shared by all pros (13292.13) becomes one WAG move per pro argument (each needs target_arg_id), or a rebuttal that loses the premise it aims at.
- WAG has no negative-consequences scheme, so the commonest con pattern can only be an Other argument today.
- No scheme concludes that an outcome is good or bad (AS:values concludes 'a reason to keep a goal'). Backing or attacking good(G)/bad(G) (13292.25, 13292.36, one reading of 13292.180) is Other only. bad(G) is binary, so 'the burden is small' cannot be said.
- The answer-variable binding breaks analogy between policies (13292.61): the proposition that held in the source case is not the answer's proposition (N12).
- Practical reasoning's four premises force the annotator to invent S2, G and V when a claim states only S1 (13292.108), or to split one target state into S2 and G (13292.118).
- No agent or scope variables, and no move that amends the action. 13292.145 (tax plus earmarked research), and the scope claims deeper in the tree (tax only beef, exempt poor countries), become new theses.
- One argument per Kialo edge. A claim with two parents (13292.180) becomes two arguments with the same premises, and rule G1 (no restating) may reject the second.
- Rule Q1: a bare 'there is a better way' (13292.169, best-means reading) cannot be a question, because an exception must be argued. It has to be an Other argument whose premise restates its conclusion.
- Missing backup and ask edges: CQ:best_means has no backup scheme except the thesis-independent ones (N9), and CQ:other_means is asked only of practical reasoning (N10).
- Bare claims carry their reasons in their own Kialo children (13292.13, 13292.169), but a WAG answer must carry its premises in the same move. So a Kialo claim plus its pros may be one WAG argument, not two.

### Recommendations

- Before scaling, have a second annotator redo these 20 claims and about 30 more, and measure agreement on scheme and role separately. The confidences here are self-estimates.
- The pattern inventory has not saturated: 15 patterns for 19 arguments, 10 of them single instances. Approve patterns in batches after every 20-30 claims, not once up front.
- Annotate each discussion top-down and freeze a parent's bindings before annotating its children: a child's role depends on the parent's scheme and on its explicit premises.
- Treat Other as a legitimate label. Here it is the honest primary label for 7 of 19 arguments, mostly bare denials and unargued factual or evaluative premises. That is not a failure of effort.
- For bare claims, try 'claim plus its pro children' as the unit. The children are the premises of the claim's own WAG argument.
- Select by centrality, not votes, below the first level: Kialo impact votes are sparse (median 1 per annotated argument, 5 with none).

## Provenance

- kialo_graph: `<kialo visualizer>/data/13292/graph.json` sha256 `d68bdd07f929559b…`
- nascq_schemes: `wag/data/nascq/schemes.json` sha256 `310757dda01c71eb…`
- nascq_cqs: `wag/data/nascq/critical_questions.json` sha256 `a617b60cb0ae23f3…`
- nascq_edges: `wag/data/nascq/edges.json` sha256 `90aa8bdba88be685…`
- core_yml: `wag/data/schemes/core.yml` sha256 `b025ebb7b56bc1b2…`
- walton_yml: `walton-argumentation-schemes-for-carneades-4-master/walton.yml` sha256 `1976202ac6bd5385…`

The NASCQ English templates and core.yml language were being reworded in the working tree (uncommitted) while this pilot ran; edges.json was unchanged. Bindings use the reworded templates; the schemes, CQs and edges are the same as at HEAD.

Walton, Reed & Macagno (2008), OCR of pp. i-135. Pages cited: 56 and 62 (analogy), 68 (verbal classification), 74 (practical analogy), 95-97 (practical reasoning and its CQs), 100 (negative practical reasoning), 101-102 (consequences and CQs), 103 (danger appeal), 106-107 (alternatives), 108-109 (need for help). Ch. 5 (cause to effect, p. 168) is outside the available pages.

Kialo texts are quoted verbatim as data, links included; the links were not followed.
