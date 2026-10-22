# Candidates

This is the candidates file of [`process.md`](process.md) §8. It lists schemes, questions and links that WAG does not have yet but that the literature, data or play have shown it may need. For each it records what is known, so that whoever implements it starts from the evidence rather than from scratch.

A candidate goes to a person once it has been seen at least three times in at least two discussions or games; whether that threshold is right is decision D047. The person decides one of four outcomes:
- reject it;
- merge it into an existing item;
- record it as a domain version of one;
- add it, drafted and checked like any other item.

Each candidate below names the decision in the [register](decisions.md) that settles it.

## 1. Argument a fortiori

**Decision:** D056. **Status:** candidate, sketched and engine-tested; it meets the threshold in Kialo.

**What it is.** A claim holds in one case, and another case is a *stronger* case for the same kind of claim, so the claim holds there all the more:
- "if a student may borrow ten books, a professor certainly may";
- negatively: "if jail was not justified for the graver act, still less for the lighter one".

The cases are not alike; they are ordered on the respect that matters. That is why analogy's form misstates it: its similarity premise is false, and a difference exception that does not say which way the difference points would defeat it by the very difference it relies on.

**Sources.**
- **The book:**
  - not one of the compendium's 60 schemes (checked against the transcription);
  - not mentioned in pp. 1–138 (OCR searched).
- **The literature** (not yet read here; read before writing its questions):
  - the classical argument from the more and the less;
  - a fortiori is one of the standard arguments of legal interpretation, and Walton, Sartor and Macagno (*Statutory Interpretation*, 2021) formalize interpretive arguments as schemes;
  - "A Refined Concept of A Fortiori Arguments for Argumentation Theory" proposes four critical questions for it.

**Evidence.**

| Source | Item | Note |
|---|---|---|
| Kialo | 5637.347 (parasitic twins, "isn't even fully formed") | fits cleanly (second review) |
| Kialo | 13292.186 (bug farming, "let alone a country") | negated form; fits cleanly |
| Kialo | 2629.8449 | found by the second review |
| Ours | proposal D001-D028-D029 examples 26–28 (a dog and a child in a hot car; a professor's loans; bread about to be thrown away) | rendered in [the proposal's view](proposals/D001-D028-D029.md) |
| Ours | the first review's R30 | analogy's form fails on it |

**The sketch** (proposal D001-D028-D029, version 3, candidates):

| Role | Formula |
|---|---|
| variables | P1, P2, C1, C2 |
| premises | `in_case(P1,C1)`, `stronger_case(C2,C1,P2)`, `corresponds(P1,P2)` |
| exception | `difference_against(C1,C2,P2)`, shared with analogy |
| conclusion | P2 |

- **Template:** `stronger_case/3`, "%s is a stronger case than %s for %s."
- **Question:** `walton/difference_against`, "Does a difference between the cases, other than the one the argument relies on, tell against the claim here?"

**What is known from the engine:**
- **It shares analogy's difference exception, rightly.** A difference that tells against the claim defeats an a fortiori and an analogy argued for the same claim from the same cases, and the stronger-case difference, which favours the claim, defeats neither. This is shown by the proposal's v3 engine test (`TestV3AFortioriBesideAnalogy`, in the private workspace) and by the second review's R13.
- **It depended on D001 being directional.** With version 2's exception, which only said that the difference "bears on" the claim, adding it would have forced a change to analogy (the second review's R10).

**What implementing it needs:**
- a `walton.yml` addition through `walton-clean/tools/edits.py` (ADDITIONS) and its template;
- a catalogue entry, with no book entry, in family F04 (analogy and case-based arguments);
- questions drafted from the literature;
- the game later, at workstream E.

**Open questions:**
1. **"A stronger case than C1 for P2" reads circularly.** P2 is already the claim about C2. Should the respect in which C2 is stronger be a variable of its own?
2. **"Other than the one it relies on"** lives only in the question's text. Is that enough, or does the exception need it in its formula?
3. **Can annotators code the negated form** (example 28) without confusing the two directions?
4. **Does it answer any question in the game, as analogy does?** It concludes a bare claim, so D057 applies to it too.

## 2. Refutation by parallel argument (refutation by logical analogy)

**Decision:** D056. **Status:** candidate, sketched and engine-tested; it meets the threshold in Kialo. D056 must first choose its form.

**What it is.** An argument is attacked by citing a parallel argument:
- of the same form;
- with plainly acceptable premises;
- with a plainly unacceptable conclusion.

The attacked argument's inference is thereby undercut ("by that reasoning, destroying an acorn would be as bad as felling an oak"). Its conclusion is about an argument, not about a case. That is why analogy, read as an argument about cases, cannot code it in annotation: there it can only rebut the conclusion, and the parallel argument it cites appears in no formula.

**Sources.**
- **The book:**
  - no scheme in the compendium;
  - pp. 54–55 discuss Govier's a priori analogy, the family this belongs to.
- **Names in the literature:**
  - Govier's "logical analogies";
  - Woods and Hudak's "parity of reasoning";
  - Hitchcock's "counterexampling parallel arguments";
  - Juthe's "refutation by parallel argument" (*Argumentation*), which describes it as duplicating the core of the target argument.

**Evidence.**

| Source | Item | Note |
|---|---|---|
| Kialo | 5637.12201 (acorns), answering 5637.12151 | proposal example 29 |
| Kialo | 2629.18591 (airplanes made by wind and chance) | ridicule by parity; proposal example 30 |
| Kialo | 5637.2296 (condoms) | fits cleanly; 5637.3072 answers it with a feature the parallel lacks, the scheme's exception |
| Kialo | 2629.22874 (Sherlock Holmes) | fits |
| Kialo | 2629.19533 (failing to have a child) | strain: a reductio of a principle, not a parallel argument |
| Kialo | 5637.3697 (emergency rooms) | partly a parity refutation |
| Ours | proposal examples 29–31; the first review's R5, R23, R39 | |

**The sketch** (proposal D001-D028-D029, version 3, candidates):

| Role | Formula |
|---|---|
| variables | A, the attacked argument on the board; B, the parallel argument |
| premises | `same_form(B,A)`, `acceptable_premises(B)`, `unacceptable_conclusion(B)` |
| exception | `form_difference(A,B)` (A relies on a feature that B lacks) |
| conclusion | `undercut(A)`, the attacked argument's undercutter in WAG's engine |

**What is known from the engine** (the second review, R9, R9b, R9c and R12):
- **It works as an answer at the argument locus.** It undercuts the attacked argument, and its own exception, once proven, reinstates it.
- **As a support move it never applies.** The adapter files its main argument as an exception argument, with no owner and no presumed premises.
- **The move check (S9) refuses it at the inference locus,** since it has no answer variable.
- **Bound to a description, A names no argument.** A must be bound to the attacked argument's id.
- **Analogy can already make this move in the game.** The game accepts analogy as a direct undercutter, with P2 bound to `undercut(A)`: `similar_to(B,A)`, `in_case(P1,B)` and `corresponds(P1,undercut(A))` conclude `undercut(A)`.

**The choice D056 must make:**
- **(a) Analogy played as an undercutter.**
  - No new scheme.
  - Analogy's templates must read well with an argument as a case.
  - Annotation needs a rule for coding parity refutations this way.
- **(b) A scheme of its own.**
  - It needs code: S9 must return `undercut(<target>)` at the inference locus, so that the move check binds A to the attacked argument's id.
  - It is used only on the answer path.
  - Its templates must be able to show an argument's text.
  - D057 could then bar analogy from the argument locus.

**Open questions:**
1. **"The premises of B are plainly acceptable"** strains when B is a pattern rather than an argument, and is false when B shares the very principle under attack (a reductio). Is the reductio of a principle (2629.19533) the same candidate or another?
2. **Ridicule by parity** (2629.18591) is often a weak or fallacious use. The scheme's exception is the defence, and the dialogue type should set its weight (eristic or persuasion).
3. **The book's a priori analogy** (pp. 54–55) is the source to read before writing its questions.

## 3. Found by the independent review of proposal D055-D059

**Decision:** D064; the link in 3.12 is D063. **Status:** recorded by the review of the consequence family's proposal (2026-10-22). Not yet researched. The review's file is `candidates.json` in the private workspace (`verify/p-D055-D059/`). Each Kialo text cited here was spot-checked against the discussions. "Active" leaves out archived claims.

| # | Kind | Candidate | Seen | Threshold | Review's confidence | Closest |
|---|---|---|---|---|---|---|
| 3.1 | scheme | Weighing an opposite consequence against a cited one | 9 active Kialo claims in 2 discussions; ND three times | met | high | D055's exceptions; `negative_practical` (26.2) |
| 3.2 | scheme or exception | Balance of consequences, a global cost-benefit claim | 4 active in 3 discussions; ND once | met | medium | sunk costs' `future_losses_outweigh(A)`; D033 |
| 3.3 | question | How serious is the cited consequence? | 4 or more in 2 discussions; ND; 2 gold challenge types | met | medium | 33/CQ1; the premise challenge |
| 3.4 | question or exception | Mitigation: can the action be done so that the harm does not follow? | 6 in 2 discussions (5 in the gold) | met | medium | practical reasoning's alternative-means exceptions; the gold's `action_amendment` role |
| 3.5 | question or exception | A right or value that rules the action out (or in) whatever its consequences | 4 active in 2 discussions, 3 more archived | met | medium | practical reasoning's `side_effects`; the value schemes |
| 3.6 | scheme | Precaution against a severe, irreversible risk | 4 active in 2 discussions, 3 archived in a third | met | medium | danger (37); 33.2 with a modal `brings_about` (D008) |
| 3.7 | scheme | Prediction from an agent's incentives | 5 active in 2 discussions | met | medium | abductive character to action (31.4); practical inference |
| 3.8 | scheme | The lesser evil | 3 active in 2 discussions; ND twice | met | medium | pragmatic argument from alternatives (34) |
| 3.9 | question or link | Is a consequence argument offered for the truth of a claim? (appeal to consequences) | 4 active in 1 discussion; ND once | not met (one discussion) | low | S9; the consequence entries' dialogue-type notes |
| 3.10 | question | Is the warning an indirect threat? | the book's two examples; not looked for in Kialo | not met | low | threat (35), ad baculum (35.1), 33.4 |
| 3.11 | scheme | The greater good: a harm as the necessary price of a greater good | 5 active in 1 discussion | not met (one discussion) | low | 3.1; D055's `good_outweighs` |
| 3.12 | link | The family's practical conclusions should meet `should_be_performed` | 9 schemes in walton.yml | — | high | registered as D063 |

### 3.1 Weighing an opposite consequence against a cited one

**What it is.** An argument that a consequence H of an action matters more than the consequence G that an argument of the other sign cites. The other argument then does not carry. It is the claim behind D055's exceptions, made with a scheme tied to H, rather than with Other.

**Evidence.**
- **ND:** p. 159 (weighing the worth of the good consequences against the seriousness of the bad), p. 168 (the greater good and the lesser evil decide), p. 170 (the bad can outweigh the good).
- **Kialo, active:**
  - 13292.127 and 13292.43 (both in the gold);
  - 5637.12239, 5637.12666, 5637.14334, 5637.14031;
  - 13292.55.
- **Kialo, archived:** 5637.357, 5637.1752, 5637.239, 5637.4449.
- **The gold:** 13292.127's secondary instance is coded Other with `more_serious_consequences/2`. walton.yml has that predicate only inside negative practical reasoning (26.2).

**Sketch.** It was gated in the review and tested in WAG's engine.
- `bad_outweighs_good [A,H,G]`: premises `brings_about(A,H)`, `bad(H)`, `matters_more(H,G)`; conclusion `bad_outweighs(A,G)`.
- `good_outweighs_bad`: the mirror, concluding `good_outweighs(A,G)`. The two are a polarity pair under D053.
- **Template:** `matters_more/2`, "%s matters more than %s."
- **Questions:** 33/CQ1 and 33/CQ2 at `brings_about(A,H)`, and a question about the comparison at `matters_more(H,G)`.

**What the engine showed.**
- The weighing argument's first premise is the same statement as the matching consequence argument's.
- Undermining that link takes down both, and the argument weighed against is reinstated.
- Argued with Other, the weighing claim would survive the loss of the consequence it relies on.

**Open questions.**
- Is it one scheme with D055's exception as its conclusion, or a question form of 33/CQ3?
- How is `matters_more` judged (severity, number affected, reversibility; see 3.3 and 3.6)?

### 3.2 Balance of consequences

**What it is.** "All told, the costs outweigh the benefits", or the reverse, naming no single consequence. Under D055's sided exceptions it must be argued once per argument of the other side (proposal D055-D059, example 17).

**Evidence.**
- **ND:** p. 170 (a favourable or unfavourable balance).
- **Kialo:**
  - 13564.400 (the benefit "might not be worth the costs");
  - 5637.12666 ("outweighs any unethical implications");
  - 5637.14336 (no net benefit);
  - 2629.17573 (a net positive effect).

**Sketch.** Two forms:
- **Side-level exceptions:** `net_cost(A)` on the positive consequence schemes and 33.4, and `net_benefit(A)` on the negative ones. They are sided, so one proof undercuts one whole side. Sunk costs' `future_losses_outweigh(A)` is the one-sided precedent.
- **A scheme** concluding `¬should_be_performed(A)` from `net_cost(A)`, weighed like any rebuttal.

Decide with D033.

### 3.3 How serious is the cited consequence?

**What it is.** The book's three questions cover likelihood, evidence and opposite consequences, but not how good or bad the cited consequence is, or for how many. ND names seriousness and number beside likelihood. In WAG `good(G)` and `bad(G)` are yes-or-no, and no question targets them. So a minimising reply ("the burden is small") is a bare challenge.

**Evidence.**
- **ND:** pp. 158–159.
- **The gold:** challenges of type "severity or value" with no printed question (13292.127, 13292.56).
- **Kialo:** 13292.25 ("as small as changing one's food consumption pattern"); 13292.58 and 13292.226.

**Sketch.** `walton/good` and `walton/bad`: "Is G good (bad) enough, and for enough people, to bear on whether A should be done?", at the premise.

### 3.4 Mitigation

**What it is.** A reply that keeps the action but adds a measure or narrows its scope, so that the bad consequence no longer follows. The gold codes it with the role `action_amendment`, with no scheme question behind it.

**Evidence.**
- **The gold:**
  - 13292.202 against 13292.127 (exempt poor countries);
  - 13292.235 against 13292.121;
  - 13292.194 and 13292.45 against 13292.12 (import safeguards);
  - 13292.24 against 13292.36 (tax credits or food stamps).
- **Kialo:** 5637.144 against 5637.84 (genetic selection prevented without a ban).

**Sketch.**
- **As an exception:** `avoidable(A,G)` on 33.2 and 33.3, "The action %s can be carried out so that %s does not follow".
- **As a question at `brings_about(A,G)`:** answered by an argument for the amended action.

**Open question:** does it undercut the argument, or redirect the issue to the amended action?

### 3.5 A right or value that trumps consequences

**What it is.** A consequence weighed against a right or a value rather than against another consequence: liberty against health effects, a right to life against hardship. D055's exceptions compare consequences only.

**Evidence.**
- **Kialo, active:** 13292.29 and 13292.55 (personal choice against health consequences); 5637.1259; 5637.14182.
- **Kialo, outside the active scope or archived:** 5637.3450, 5637.239, 5637.4449.

**Sketch.**
- **Sided exceptions on the consequence schemes:** `forbidden(A)` on the positive side, `required(A)` on the negative.
- **Or a value-priority premise** in a value-based scheme.

**Open question:** are rights weighed against consequences, or do they trump them?

### 3.6 Precaution

**What it is.** An action should not be taken because it risks a catastrophic or irreversible harm, even an unlikely one. Likelihood is conceded and outweighed by severity and irreversibility, which 33/CQ1 cannot express.

**Evidence.**
- **Kialo, active:** 13564.175 ("irreversible changes"), 13564.70, 13564.1159, 27150.203.
- **Kialo, archived:** 5637.12406, 5637.4449, 5637.239.

**Sketch.** `may_bring_about(A,G)`, `catastrophic(G)`, `irreversible(G)` ⇒ `¬should_be_performed(A)`.
- **Exceptions:** `negligible_risk(A,G)`, `precaution_costs_outweigh(A)`.
- **Or a question for 33.2:** "Is the consequence irreversible?", feeding 3.1's comparison.

### 3.7 Prediction from an agent's incentives

**What it is.** A prediction that agents will, or will not, do something because they expect it to bring them good, or bad, consequences. Predictive debates use it as support for a factual claim, not a recommendation.

**Evidence (Kialo):**
- 27150.193 (expected benefits will push stakeholders);
- 27150.233 with 27150.197;
- 13564.393 with 13564.48;
- 13564.400 and 13564.430.

**Sketch.** `agent(P)`, `can_do(P,A)`, `expects(P,brings_about(A,G))`, `good_for(P,G)` ⇒ `likely_to_do(P,A)`.
- **Exceptions:** `stronger_counter_incentive(P,A)`, `unable(P,A)`.
- It concludes a prediction, so it must not meet `should_be_performed`.

### 3.8 The lesser evil

**What it is.** Of two courses that both have bad consequences, take the one whose consequences are less bad. The comparison is between alternatives, not between the consequences of one action. ND ties it to the practical dilemma.

**Evidence.**
- **ND:** p. 160 (the dilemma), p. 168 (the lesser evil).
- **Kialo:** 5637.208; 5637.14334; 2629.7306.

**Sketch.** `either(A,B)`, `brings_about(A,G1)`, `brings_about(B,G2)`, `bad(G1)`, `bad(G2)`, `less_bad(G1,G2)` ⇒ `should_be_performed(A)`.
- **Exception:** `third_option(A,B)`: slipping between the horns (ND p. 161).
- It may merge into the pragmatic argument from alternatives (34).

### 3.9 Appeal to consequences, as a question

**What it is.** A consequence argument offered for the truth of a claim: bad consequences of believing P, or of P being true, used against P.
- S9 already keeps a consequence scheme from concluding a factual statement in the game.
- The catalogue's dialogue-type notes mark inquiry as suspect.
- But nothing lets a player raise the objection that Kialo users raise by name.

**Evidence.**
- **ND:** pp. 176–177.
- **Kialo:**
  - 2629.17573, attacked by 2629.18670;
  - 2629.3501, attacked by 2629.19027 (which names the fallacy);
  - 2629.10311.

**Sketch.** An undercutter at the argument locus: "Does the conclusion about what to do bear on whether P is true?"

### 3.10 Is the warning an indirect threat?

**What it is.** A prudential warning (33.4) conveys an indirect threat when the speaker can bring the bad consequence about. The book separates the two by the speaker's commitment to bringing it about.

**Evidence.**
- **The book:** p. 102 (an adviser's warning is not a threat); pp. 104–105 (two warnings read as indirect threats).
- **ND:** p. 116.

**Sketch.** A question for 33.4: "Is the speaker able to bring C about, and committed to it if X is not done?" If so, recode the argument as threat or ad baculum. Both need D060 and D063 first.

### 3.11 The greater good

**What it is.** A harm is justified because it is inseparable from, or necessary for, a greater good. It is a weighing with a necessity link.

**Evidence.**
- **Kialo:** 2629.1471, 2629.2856, 2629.944, 2629.7265, 2629.7306, all in one discussion.
- **Kialo, archived:** 2629.2119.

**Sketch.** It is perhaps 3.1 with an extra premise, `necessary_for(B,G)`.

### 3.12 The family's practical conclusions

Registered as D063: fear, danger, threat, the alternatives, waste, help, distress and two-person practical reasoning conclude predicates that meet nothing.

## Limits recorded elsewhere

These are not candidates yet. They are recorded in the catalogue's analogy entries:
- several source cases folded into one (Carneades does not add arguments up);
- proportional analogies;
- imagined sources;
- a precedent cited for its rule, which is `exception_to_rule` (whose exception reads a rule as a case: D058).
