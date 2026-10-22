# Candidates

This is the candidates file of [`process.md`](process.md) §8. It lists schemes, questions and links that WAG does not have yet but that the literature, data or play have shown it may need. For each it records what is known, so that whoever implements it starts from the evidence rather than from scratch.

A candidate goes to a person once it has been seen at least three times in at least two discussions or games; whether that threshold is right is decision D030. The person decides one of four outcomes:
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

## Limits recorded elsewhere

These are not candidates yet. They are recorded in the catalogue's analogy entries:
- several source cases folded into one (Carneades does not add arguments up);
- proportional analogies;
- imagined sources;
- a precedent cited for its rule, which is `exception_to_rule` (whose exception reads a rule as a case: D058).
