# The scheme catalogue

One entry per argumentation scheme. Each entry aligns three sources:
- **the book:** Walton, Reed & Macagno, *Argumentation Schemes*, 2008, chapter 9's compendium;
- **the Carneades collection:** `walton.yml`;
- **the game's NASCQ.**

Every item that people or models write carries a status. This is phase 1 of [the plan](../../../docs/plan.md#2-phase-1-the-aligned-catalogue).

**Status on 2026-10-22: the skeleton.**
- **Generated:** every entry, from the sources.
- **Not yet done:** nothing is drafted or verified.
- **Next steps** (plan §2, order of work):
  1. A model drafts the 31 tier-1 entries.
  2. A person verifies them against the book.

| File | What |
|---|---|
| `schemes/<id>.yml` | 114 entries: one per `walton.yml` scheme (111), and one per compendium scheme that `walton.yml` lacks (3) |
| [`index.md`](index.md) | One line per entry, tier 1 first, with how many of its items are verified. Generated. |
| This README | The format, the statuses, the commands, the checks |

**Who decides what** (the owner's decision of 2026-10-15):

| Part of an entry | Decided by |
|---|---|
| What a scheme says, and its critical questions | The book |
| The formulas | `walton.yml` |
| The game's critical questions | The game's NASCQ, until it is generated from the catalogue (plan §6) |

## Commands

Run them from `wag/`:

```sh
go run ./cmd/wag-cli --catalogue-gen [--by "WHO"]   # (re)generate the skeleton; keeps what people and models wrote
go run ./cmd/wag-cli --catalogue-check [--book FILE,FILE]   # check everything; status 1 on an error
```

- **Where things are.** Both commands find the repository's root from the working directory; `--repo` and `--catalogue` override that.
- **`--catalogue-gen` is safe to rerun.** It rewrites only what the sources own (below), and writes `index.md`.
- **`--catalogue-check`** reports errors, which fail it, and warnings. It ends with a progress line.
- **A test keeps the committed catalogue current:** `internal/catalogue`, `TestCommittedCatalogueIsCurrent`. When a source changes, run `--catalogue-gen` and commit the result.

## An entry

`schemes/expert_opinion.yml`, shortened:

```yaml
id: expert_opinion
name: Argument from Expert Opinion
tier: 1
walton: {id: expert_opinion, title: ..., source: ...}       # source: walton.yml
game: {id: AS:expert_opinion}                                # source: the NASCQ
formulas:                                                    # source: walton.yml
  variables: [W, D, S]
  premises: [expert(W,D), in_domain(S,D), asserts(W,S)]
  ...
book:                                                        # links.json, then verified
- {entry: "2", title: Argument From Expert Opinion, pages: "310", match: exact, status: generated}
families: [...]                                              # families.json, then verified
premise_map:                                                 # links.json, then verified
- {entry: "2", label: Major Premise, formulas: [{formula: 'expert(W,D)', locus: premise}, ...], status: generated}
critical_questions:
- id: 2/CQ1                                                  # the book's entry and label
  book: {entry: "2", label: CQ1}
  game: {id: CQ:expertise, text: ..., locus: premise, pattern: expert(W,D)}   # source: the NASCQ
  targets: [{formula: 'expert(W,D)', locus: premise}]       # links.json, then verified
  paraphrase: ""                                             # to write, in our own words
  answer: ""                                                 # to write: what counts as an answer
  status: generated
summary: {text: "", status: todo}
contrasts: [{scheme: position_to_know, text: "", status: todo}, ...]
examples: []
discrepancies: []
classification: {source: "", reasoning: "", epistemic: "", group: "", status: todo}
specialises: {scheme: "", note: "", status: todo}
dialogue_types: {proper: [], suspect: [], note: "", status: todo}
history: [{date: "2026-10-22", by: wag-cli --catalogue-gen, what: skeleton generated ...}]
```

### The fields

| Field | What | Who owns it |
|---|---|---|
| `id`, `name` | The `walton.yml` id and title. For a compendium scheme that `walton.yml` lacks, an id made from the book's title. | Source |
| `tier` | 1: verified first (below). 2: the rest. | Set when the entry is made; then people |
| `walton` | The `walton.yml` scheme and its own citation (`meta.source`); `null` if none | Source |
| `game` | The NASCQ scheme that plays it, and the questions it can back an answer to (backup edges) | Source |
| `formulas` | Variables, premises, assumptions, exceptions and conclusions, as in `walton.yml` | Source |
| `book` | Each compendium entry the scheme formalizes: entry number, title, pages, the heading it sits under, how closely (`exact`, `partial`; `none` for a scheme `walton.yml` lacks), and the entry its questions are printed under, if it shares a list | Work: from `links.json` |
| `families` | Each family of `families.json` it belongs to, and the relation | Work: from `families.json` |
| `premise_map` | Each printed premise and conclusion, by its label, and the formulas that express it | Work: from `links.json` |
| `critical_questions` | Every question of the book's entries (id `<entry>/<label>`, from the entry the list is printed under), and every game question asked of the scheme. A game question sits on the book question that targets the same formula, or on its own (`game/<id>`). **`targets`:** the formulas the question puts in doubt, with their role. **`undercuts`:** the question attacks the inference itself. **`paraphrase`:** the question in our own words, with the scheme's variables. **`answer`:** what counts as an answer, and what each answer does. | Work, except `game`, which is source |
| `summary` | When the scheme applies, in two sentences of our own | Work |
| `contrasts` | "This, not X, when …" for each close relative. Generated as targets to write: the family's generic scheme (or, for the generic, every member), and the schemes the registry finds close in pattern. People add others. | Work |
| `examples` | Two or three short claims coded in full, Kialo claims with their ids where possible, and one near miss (`near_miss: true`) | Work |
| `discrepancies` | Where the sources disagree, and the decision. Kinds: `partial-link`, `no-formalization`, `not-in-book`, `game-differs`, `link-formula` (made by the generator from the sources); `book-walton`, `other` (written by people). | Work |
| `classification` | Walton and Macagno's tree. **`source`:** `dependent` or `independent`. **`reasoning`:** `practical` or `epistemic`. **`epistemic`:** `discovery` or `application`. **`group`:** the thematic group, in words. | Work |
| `specialises` | The scheme this one is a domain version of, if any | Work |
| `dialogue_types` | Where the scheme is a proper move, and where it is suspect ([`dialogue-types.md`](../../../docs/dialogue-types.md) §7). Types: `persuasion`, `inquiry`, `information_seeking`, `negotiation`, `deliberation`, `eristic`. | Work |
| `history` | Notable events. The history of each field is the git log of the file. | People |

**Source fields have no status.** They are copies, never edited here. The generator rewrites them, and the checker fails if they differ from the sources. To change one, change the source:
- `walton.yml` for formulas;
- the NASCQ for the game;

then regenerate.

**English readings of the formulas are not stored.** 173 of `walton.yml`'s 393 English templates share eight or more words in a row with the book's compendium, and copying them into 114 files would multiply that. The package renders them from `walton.yml` when a reader or a prompt needs them (`catalogue.English`). Rewriting those templates in our own words is part of the owner's decision on book text ([plan §7](../../../docs/plan.md#7-decisions-the-owner-needs-to-make), item 1).

### Statuses

Every work item has one:

| Status | Meaning |
|---|---|
| `generated` | Made by the generator from a source that is itself only drafted (`links.json`, `families.json`). Not yet checked by a person. |
| `todo` | Empty, to be written |
| `drafted by WHO, YYYY-MM-DD` | Written by a model or a person; not yet checked against the book |
| `verified by WHO, YYYY-MM-DD` | Checked against the book by a person |
| `disputed by WHO, YYYY-MM-DD: REASON` | A person found it wrong, and says why |

**The generator owns an item only while it is `generated` or `todo`.** Rerunning it then rewrites the item from the sources. Once you draft, verify or dispute an item, it is yours:
- the generator keeps it as it is;
- the checker still checks it.

So set the status before or with any edit. An edit to a `generated` item is overwritten at the next run.

**One part stays source even in your items:** the game part of a question (`game`). It is refreshed even in a kept question. A game question that you move from one book question to another stays where you put it.

## The checks

`--catalogue-check` fails (an error) when:
- **Coverage:**
  - a `walton.yml` scheme or compendium scheme has no entry, or an entry is not made from the sources;
  - a file does not read, has unknown keys, or is not named after its id.
- **Sources:**
  - a source field differs from its source: `walton.yml` for the formulas, the NASCQ for the game;
  - an item still `generated` differs from what the sources make (run `--catalogue-gen`).
- **Formulas:** a formula an item names is not one of its scheme's formulas, or not in the role given. This covers premise maps and question targets, and is the rule that every formula is `walton.yml`'s.
- **Missing items:**
  - a compendium entry, premise map, family or book question that the sources link to the scheme is missing;
  - a game question asked of the scheme is missing, or it is not the NASCQ's;
  - a discrepancy the sources show is missing.
- **Statuses and values:**
  - a status does not follow the grammar, or has a bad date;
  - an item marked drafted or verified has no text;
  - an example lacks its text or its coding;
  - a contrast or `specialises` names an unknown entry;
  - a classification, dialogue type or discrepancy kind is outside the vocabulary.
- **Copying:** a text of ours reproduces the book (below).
- **The index:** `index.md` is out of date.

**It warns:**
- when a book question targets no formula and is not marked as undercutting (to decide while verifying);
- when a game question targets another formula than its book question;
- when the sources disagree with each other;
- when the copy check did not run.

### The copy check: no text of ours reproduces the book

The texts checked are:
- summaries;
- question paraphrases and answers;
- contrasts;
- examples;
- the decisions on discrepancies;
- the notes.

Each is compared with the book's text, normalized: lower case, and every variable (E, A, S1) as one word, so that a copy with its variables renamed still matches.

| Found | Result |
|---|---|
| 8 words in a row shared with a book sentence | Error |
| A text of up to 25 words that is 85% the same words, in order, as a book sentence | Error |
| 70% the same, or 6 words in a row | Warning: rephrase it |

**The book's text is not in the repository.** The checker is given it with `--book`: plain-text files, or JSON whose `text` and `book_text` keys hold the book's sentences.
- **Without `--book`,** it uses `docs/book-catalogue/catalogue.json` while that file exists. That file covers the book's pp. 1–138 only.
- **For the compendium (ch. 9),** which most questions come from, give your own transcription. For example, the transcription the links were made from: one line per premise, conclusion and question.

## Tier 1

31 schemes, verified first: the 13 schemes of the batch-2 gold, their close relatives for contrasts, and defeasible modus ponens, the form "other" falls back to (plan §2).

| Group | Schemes |
|---|---|
| Practical reasoning | `practical_reasoning1`, `practical_reasoning2`, `practical_goal`, `end_and_means_1`, `practical_necessary_condition` |
| Consequences and values | `positive_consequences`, `negative_consequences`, `positive_value`, `negative_value` |
| Causes and signs | `cause_to_effect`, `correlation_to_cause`, `sign` |
| Cases and analogy | `analogy`, `practical_analogy_positive`, `practical_analogy_negative`, `from_example` |
| Sources | `expert_opinion`, `position_to_know`, `witness_testimony`, `popular_opinion` |
| Classification and rules | `verbal_classification`, `definition_to_verbal_classification`, `rule`, `established_rule` |
| Alternatives and costs | `alternative_cognitive`, `alternative_normative`, `sunk_costs`, `waste` |
| Slippery slope | `slippery_slope_base_case`, `slippery_slope_inductive_step` |
| The general form | `defeasible_modus_ponens` |

## What the skeleton found

Generated on 2026-10-22.

| | Count |
|---|---|
| Entries | 114: 111 `walton.yml` schemes and 3 compendium schemes that `walton.yml` lacks (10.2 division of the whole into its parts, 50 precedent slippery slope, 53 full slippery slope) |
| Compendium links | 111 links from 106 compendium entries: 94 exact, 17 partial. Only `defeasible_modus_ponens` has no compendium entry. The figure of six such schemes in earlier documents came from an older list; `links.json` links the other five. |
| Critical questions | 212. **From the book:** 200. **From the game:** 35, of which 23 were matched by formula to a book question and 12 stand on their own. 60 book questions target no formula yet. |
| Premise maps | 366 |
| Contrasts to write | 376 |
| Discrepancies from the sources | 22: 17 partial links, 3 compendium schemes without a formalization, 1 scheme not in the compendium, 1 game difference (the game's `sign` lacks the exception `more_reliable`) |
| Work items | 1,678: 846 generated (to verify), 832 to write |

**`links.json` holds up.** All 519 of its formula links, for premises and questions, name formulas of their scheme in the role it gives. 39 of them use the upstream spelling, which the cleaning changed, and the generator applies the cleaning's edits (`changes.json`) before matching:
- misspelled predicates, such as `succesful_diagram` and `commited_opposite`;
- primed variables (`P'`, now `P2`);
- one stray parenthesis.

**The sources disagree once.** `families.json` lists compendium 33.3 as having no `walton.yml` scheme, while `links.json` links it, partially, to `negative_consequences`. The checker warns; one of the two needs correcting.
