# The scheme registry: one pattern, one name

This directory holds what WAG uses to keep its argumentation schemes accurate and distinct as the set grows. Background: the owner's points 3–5 of 2026-10-13.

| File | What |
|---|---|
| `walton-clean/walton.yml` | The Walton–Carneades collection (MPL-2.0, see `walton-clean/LICENSE`), cleaned so that Carneades 4 loads every scheme: the upstream 111 schemes, and four more. Syntax, typos, variable declarations and English templates are fixed. The cleaning changes no scheme's content; the owner's decisions of 2026-10-22 change five upstream schemes and add four (CHANGELOG, "Changes decided by the owner"). Three formalize compendium entries that the upstream collection has only in its category files: `negative_consequences_reasoning` (33.3, under an id the master already used), `whole_into_parts` (10.2) and `full_slippery_slope` (53). The fourth, `precedent_slippery_slope` (50), was written for WAG, since no upstream file formalizes it. The five changed: the consequence schemes get the exception `other_consequences(A)` for the book's question about opposite consequences; the slippery slope's base case concludes `¬should_be_performed(A)`; `negative_value` gets its own conclusion and a corrected template; the slippery slope's source year follows the book. |
| `walton-clean/CHANGELOG.md`, `changes.json`, `walton.diff` | Every edit (282: 196 cleaning edits, and 86 for the owner's decisions), with its reason and category. Four repairs and five renamed IDs need the owner's approval (CHANGELOG, "Repairs that need the owner's approval" and "Id map"); the audit of 2026-10-13 approves them. |
| `walton-clean/id-migrations.json` | The five renamed scheme IDs, old to new, so that a file using an upstream ID still reads (`validate.py` uses it). |
| `walton-clean/walton.proposals.yml`, `proposals-check.txt` | The same file with four further proposals that change a premise. With them, all 115 schemes pass the Carneades validator. Not applied. |
| `walton-clean/tools/`, `walton-clean/check/` | The cleanup as code: the edits (`tools/edits.py`, `tools/proposals.py`), the checker (`check/main.go`), and `check/run.sh`. Run `sh walton-clean/check/run.sh ORIGINAL` on the upstream `walton.yml` (its SHA-256 is in the CHANGELOG). It rebuilds `walton.yml`, `changes.json`, `walton.proposals.yml`, `walton.diff` and the CHANGELOG, and reruns every check (`clean-check.txt`, `proposals-check.txt`, `check/*.txt`). It reproduces the committed files byte for byte. |
| `nascq-edges.md` | Every answer the game's NASCQ offers to every critical question, played on a sample argument: what each concludes, its English reading, and whether that is what the question's locus and the answer's polarity call for. It is made by `wag-cli --edges`, and a test keeps it current. This is the table to review the scheme–question links against. |
| `links.json` | For each of the 128 entries of the book's compendium (Walton, Reed & Macagno 2008, ch. 9: 60 schemes, 68 subtypes), which `walton.yml` scheme formalizes it and how closely. Also maps each premise, the conclusion and each critical question to its formula. Labels only: the book's text is not reproduced. |
| `../../wag/data/schemes/families.json` | 21 scheme families: generic scheme, specializations, variants, subtypes, counterparts and context-specific forms (133 members). |

## How uniqueness is checked

`wag-cli --registry FILE.yml --families data/schemes/families.json [--strict]` (package `internal/registry`). It reduces each scheme to its **pattern**: its formulas by role (conclusion, premises, assumptions, exceptions), with the variables renamed in a canonical order. Then it compares every pair:

| Relation | Meaning | Reported as |
|---|---|---|
| identical | The same formulas under two names. | Error, unless the family file declares them aliases. |
| roles | The same predicates, some in another role (a premise in one, an assumption in the other). | Warning, unless a family declares the pair. |
| one-formula | The same but for one formula: a variant. | Warning, unless a family declares the pair. |
| extends | One has every formula of the other, and more: a specialization. | Warning, unless a family declares the pair. |
| same-form | The same inference (premises to conclusion, predicates abstracted), with one premise predicate swapped, such as expert opinion and position to know. | Warning, unless a family declares the pair. |
| close | Same conclusion, at least 45% of the formulas shared (`--close`). | Warning, unless a family declares the pair. |

Predicate names carry the meaning. Two schemes with the same shape but different predicates are relatives, not duplicates. A family records how they relate, and the check makes sure no close relative goes undeclared.

**Results, 2026-10-13.**

| File | Result |
|---|---|
| `walton-clean/walton.yml` | 115 schemes, 114 distinct patterns. With the families, every related pair is declared. One error: `division_generic` is a copy of `composition_generic`, while the book's division (10.1) runs from the whole to the parts. That is a defect to fix in the collection; the check reports it. |
| WAG's `core.yml` | 10 Carneades schemes, 10 patterns. Related pairs, all declared in the families: credible source, expert opinion and position to know (same form); value-based and goal-based practical reasoning (close). The game's NASCQ uses 9 of them: `credible_source` stays in `core.yml` so that games played before 2026-10-13 replay, and no NASCQ scheme maps to it. |

## Adding a scheme to the game

1. **Copy the scheme.** Copy it from `walton-clean/walton.yml` into `wag/data/schemes/core.yml`, with English templates for its predicates and their negations (`¬p/n`).
2. **Add its NASCQ entry.** In `wag/data/nascq/schemes.json`, give its variables (type, description, example), sentences, cue words and weight. Give an `answer_variable` if its conclusion is a bare statement variable.
3. **Link its critical questions.** Take them from the compendium entry in `links.json`. Add the questions in `critical_questions.json`, and link them in `edges.json`:
   - an ask edge per question, with its locus (premise, exception, conclusion, argument) and pattern;
   - a backup edge where a scheme can back an answer.
4. **Check it.**
   - `wag-cli --lint` checks the data.
   - `wag-cli --registry data/schemes/core.yml --families data/schemes/families.json` checks that the scheme repeats no pattern and that its family is declared.
   - Add the scheme to its family in `families.json`.
5. **Test it.** Write a game test that plays it: a support, an attack on each question, and an argument for or against a statement it concludes. See `internal/game/statement_test.go` and `audit_test.go`.

Goal-based practical reasoning (`practical_reasoning2`) and negative consequences (`negative_consequences`) were added this way on 2026-10-13.
