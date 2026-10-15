# WAG: the Waltonian Argumentation Game

WAG is a turn-based argumentation game, and a research workbench around it.

- **The game.** Players put forward arguments built from Walton's argumentation schemes. They attack, defend and question each other's arguments through the schemes' critical questions, to any depth.
- **What decides the moves.** A network of schemes and critical questions (the NASCQ, after Pyon & Ri 2022) decides which moves are legal. Carneades 4 labels every statement In, Out or Undecided after each move.
- **Who plays.** A seat can be a person, a rule bot, or a language model.
- **The research aims.**
  1. **Schemes.** Ground the game's schemes in real debate: annotate Kialo discussions as scheme instances, with models and people, and grow the scheme network from what the data shows.
  2. **Dialogue types** (added 2026-10-15). Play a game in any of Walton's dialogue types: persuasion, inquiry, deliberation, information-seeking, negotiation, and eristic as a detection exercise. Find every type, and the shifts between them, in real discourse. The reference is [`docs/dialogue-types.md`](docs/dialogue-types.md).

**Status on 2026-10-15.**
- **Working prototype:** the game, the browser dashboard, the model players, and a first annotation pipeline. They are tested and run on Windows.
- **Weak link:** the scheme knowledge they share. The book, the Carneades collection and the game's data are not yet one verified catalogue.
- **Annotation so far:** every annotation was made by a language model; no person has checked it.
- **Dialogue types:** a reference is written. The game plays only persuasion so far.

The next step is therefore not more code. It is to align and verify that catalogue, then annotate against it ([the plan](docs/plan.md)).

**Contents**
1. [How the pieces fit](#1-how-the-pieces-fit)
2. [Where things stand](#2-where-things-stand)
3. [Current issues](#3-current-issues)
4. [Where it is going](#4-where-it-is-going)
5. [Running it](#5-running-it)
6. [Every document](#6-every-document)
7. [Repository layout](#7-repository-layout)
8. [Sources and licences](#8-sources-and-licences)

## 1. How the pieces fit

```
  the book: Walton, Reed & Macagno (2008)        walton.yml: the Carneades collection
  what a scheme says, its critical questions     the formulas (111 schemes)
                       \                          /
                        links.json: the book ↔ walton.yml   (drafted, not verified)
                                    |
                     THE CATALOGUE (planned: one verified entry per scheme,
                                    |          plus schemes derived from data)
                  +-----------------+------------------+
                  |                                    |
     the game's NASCQ (9 schemes today)     annotation guidelines and prompts
                  |                                    |
     WAG games: people, bots, models        Kialo claims → model annotators → reviewer model
     Carneades labels every statement          → your verdicts → gold → new schemes
```

## 2. Where things stand

| Part | State | What works | What is missing | Read |
|---|---|---|---|---|
| Game engine | Working; tested | Moves validated by the NASCQ (rules S0–S9, Q1); Carneades labelling; eight modes; rule bots; batch runs | Only 9 of the 111 schemes are playable | [`wag/README.md`](wag/README.md), [`docs/game-rules.md`](docs/game-rules.md) |
| Browser dashboard (`wag-server`) | Prototype, used on Windows | **Play:** dialogue and board views. **Models:** type a key, find a model, add it as a player. **Annotate:** runs, model review, your verdicts. Everything is saved in a data folder (`~/WAG`). | A catalogue browser; causes for verdicts | [`docs/prototype.md`](docs/prototype.md), [`docs/llm.md`](docs/llm.md) |
| Model players and traces | Working: tested with scripted models, and in one real Gemini run | **Providers:** Gemini, Claude, OpenAI-compatible (including local servers), an external command, a mock, a replay. **Traces:** every request, answer, token count and decision. | Measured play quality | [`docs/llm.md`](docs/llm.md), [`docs/protocol.md`](docs/protocol.md) |
| Text protocol | Working | Any chatbot can play by copy and paste (protocol 2) | | [`docs/protocol.md`](docs/protocol.md) |
| Kialo import | Working | A captured discussion becomes a game (one statement per claim, keyed by claim ID), or an annotation dataset; `wag-cli --kialo-check` imports a whole folder | | [`wag/README.md`](wag/README.md), [part 8 §2](docs/review-2026-10-15-part8.md#2-claims-merged-by-their-text) |
| Scheme sources | Split | `walton.yml` cleaned so that Carneades loads all 111 schemes; the book linked to it; families and pattern signatures | One verified catalogue; critical questions and examples for most schemes | [`docs/scheme-registry/README.md`](docs/scheme-registry/README.md), [`docs/walton-audit/REPORT.md`](docs/walton-audit/REPORT.md) |
| Annotation | Pilot | One discussion, 50 claims: two model annotators, agreement, an adjudicated gold, validators | A gold a person stands behind; more discussions | [`docs/kialo-batch2/README.md`](docs/kialo-batch2/README.md) |
| Evaluation | Basic | Validation; per-field agreement and Cohen's kappa; a report per run; re-scoring from recorded calls | Acceptable answers, causes of disagreement, development and test sets, a probe set | [`docs/plan.md`](docs/plan.md) §5 |
| Dialogue types | Reference written | Every game is a persuasion dialogue: `debate` is the one-sided form | Commitment stores; the other five types; shifts; identification in data | [`docs/dialogue-types.md`](docs/dialogue-types.md), [plan §8](docs/plan.md#8-track-d-dialogue-types) |
| Carneades 4 | Working | A fork with validator and evaluation fixes, shared with the Kialo project | | [`CARNEADES_FORK.md`](CARNEADES_FORK.md) |

## 3. Current issues

In order of importance. Each issue links to the document that explains it.

1. **The scheme knowledge is split, and the annotating model sees only part of it.**
   - The model gets all 111 formulas. It gets critical questions only where Carneades encodes them as assumptions or exceptions (54 schemes), plus the game's 24 questions for its 9 schemes. It gets no descriptions, no examples, and no contrasts between close schemes.
   - It is asked to cite the book's critical questions, which it never sees. In the first run it made one up.
   - See [part 7 §3](docs/review-2026-10-15-part7.md#3-the-questions).
2. **The guidelines allow two answers in places, contradict each other once, and code four different things as "none".** "None" covers the thesis, bare premises, reasons that fit no scheme, and non-arguments. Most disagreements in the first Gemini run trace to this, not to the model. See [part 7 §2](docs/review-2026-10-15-part7.md#2-the-14-annotations-against-the-gold).
3. **The gold is one model's reading.**
   - Both batch-2 annotators and the adjudicator were runs of the same language model, and no person has reviewed them.
   - The data is 50 claims from one policy debate, using 13 of the 111 schemes.
   - See [part 7 §3](docs/review-2026-10-15-part7.md#is-the-annotation-data-enough).
4. **Models copy formulas by hand.** They add spaces and drop exceptions. The model should choose the scheme and its bindings, and the code should write the formulas. See [plan §3](docs/plan.md#3-phase-2-annotation-version-4).
5. **Cost.**
   - Each annotation call sends about 41,000 tokens, one claim per call.
   - Gemini 2.5 Flash's free tier allows 20 requests a day, so 50 claims take three days for one annotator.
   - See [`docs/llm.md`](docs/llm.md#the-free-tier-privacy-and-limits).
6. **Books in this public repository, and no licence.**
   - **`docs/The New Dialectic.pdf` is the whole of Walton's 1998 book,** a copyrighted work. It should be removed, and preferably purged from the history. See [part 9 §5](docs/review-2026-10-15-part9.md#5-the-book-in-the-repository).
   - `docs/book-catalogue/` transcribes the scheme book's schemes and critical questions as printed.
   - WAG itself has no licence.
   - All three need the owner's decision: [plan §7](docs/plan.md#7-decisions-the-owner-needs-to-make).
7. **Every game is a persuasion dialogue, in Walton's terms.**
   - WAG has no commitment stores, no retraction rules and no stages, and these are the base of every type in Walton's account.
   - The points modes match no dialogue type.
   - The second aim needs a protocol for each type and an annotation scheme for types and shifts. That is Track D, after the scheme catalogue.
   - See [`docs/dialogue-types.md`](docs/dialogue-types.md) §9 and [part 9](docs/review-2026-10-15-part9.md).
8. **Smaller known gaps.**
   - The book's cause to effect is modal; `walton.yml`'s is not (audit N8).
   - 3 book schemes have no `walton.yml` scheme, and 6 `walton.yml` schemes are not in the book.
   - 13 of the book-to-`walton.yml` links are partial.
   - See [part 6](docs/review-2026-10-15-part6.md) and [the registry](docs/scheme-registry/README.md).

**Fixed on 2026-10-15** (part 7 §1, part 8 §2, part 9 §4):
- the context presets "Policy deliberation" and "Scientific inquiry" are renamed "Policy question" and "Scientific question": the game plays neither a deliberation nor an inquiry;
- the Kialo import kept apart claims with the same words: it had merged them into one statement, and refused the game when the merge closed a loop;
- the Annotate page no longer jumps to the top while a run works;
- the review works with one annotator;
- a used-up daily quota stops a run with a plain message, instead of retrying;
- the report is written when a pass ends;
- Gemini's default spacing is 13 s;
- the checkers ignore spaces in formulas.

## 4. Where it is going

The detail, with the effort and the "done when" of each phase, is in [`docs/plan.md`](docs/plan.md).

| Phase | Goal | The work | Who |
|---|---|---|---|
| 1. Catalogue | One verified entry per scheme, aligned across the book, `walton.yml` and the NASCQ | **Generate** a skeleton from the existing sources. **Draft:** paraphrased critical questions with their loci, descriptions, contrasts, examples. **Verify:** about 25 schemes first, then the rest as the data needs them. | A model drafts; a person verifies with the book (about 20–30 minutes a scheme) |
| 2. Annotation version 4 | Annotations that follow the catalogue | **Guidelines** as one decision procedure, with "none" split and the thesis given its conclusion form. **"Other"** with a structured proposal. **Prompts:** first candidates from an index, then the full entries. Formulas written by code; constrained JSON. | Code and guidelines; measured against version 3 |
| 3. Data | A gold people stand behind, beyond one debate | **Review** batch 2 in the dashboard. **Re-code** it under version 4. **Add** four to nine discussions of other kinds, and a probe set per scheme. | People, helped by model annotators |
| 4. Evaluation, alongside | Know what each change does | Acceptable answers; a cause for every disagreement; development and test sets; one change at a time; re-scoring from recorded calls; several models | Code, then routine |
| 5. Back to the game | Play what the data shows | **Catalogue:** the NASCQ generated from it; derived schemes join once verified. | Code |
| Track D. Dialogue types | Play every type; find every type in data | **D0:** the reference (done). **D1:** the type in exports and guidelines, and a type catalogue. **D2:** persuasion with commitments, then deliberation, inquiry, information-seeking, negotiation, shifts, and eristic as detection. **D3:** identification with agreement measured. **D4:** corpora per type. | Code, then people with model annotators |

**The next three actions:**
1. The owner decides the open questions of [plan §7](docs/plan.md#7-decisions-the-owner-needs-to-make): the book PDF and the book text, a licence, who verifies, paid model access, and Track D's open choices.
2. Generate the catalogue skeleton and its checker.
3. Verify the first set of schemes while the version 4 guidelines are drafted.

## 5. Running it

**On Windows, without installing anything.**
1. Run `wag-server.exe` (built with the command below, or a copy you were given).
2. It opens the dashboard at `http://127.0.0.1:8080`.
3. To use a model:
   1. In **Models**, choose a service and paste your key.
   2. Find a model, test it, and add it as a player.
   3. Play it in **Play**, or annotate with it in **Annotate**.

Everything is saved in `C:\Users\you\WAG`. See [`docs/llm.md`](docs/llm.md) and [`docs/prototype.md`](docs/prototype.md).

**From source** (Go 1.24 or later):

```sh
cd wag
go run ./cmd/wag-server            # the dashboard; --store DIR for another data folder
go run ./cmd/wag-cli               # a game in the terminal
go run ./cmd/wag-llm --help        # headless model games and annotation runs
go run ./cmd/wag-cli --kialo-check DIR   # import every captured Kialo discussion in DIR; report problems
go test ./...                      # all tests
GOOS=windows GOARCH=amd64 go build ./cmd/wag-server   # the Windows executable
```

**Checking annotation files:**

```sh
python3 -I docs/kialo-batch2/validate.py FILE.json
python3 -I docs/kialo-batch2/agreement.py A.json B.json
```

`FILE.json` is any annotation file, such as a run's `annotations/NAME.json`.

## 6. Every document

**Start here**

| Document | What it is |
|---|---|
| This README | Where the project stands, its issues, its direction |
| [`docs/plan.md`](docs/plan.md) | The plan: the catalogue, annotation version 4, data, evaluation; the owner's open decisions |
| [`docs/review-2026-10-15-part7.md`](docs/review-2026-10-15-part7.md) | The first real model annotation run, reviewed; why "none"; what the model is given; whether the data is enough |
| [`docs/dialogue-types.md`](docs/dialogue-types.md) | **The reference on dialogue types:** what each type is, how types mix and shift, which arguments are proper in which type, how to recognise them, and what WAG needs to play and find them |
| [`docs/review-2026-10-15-part9.md`](docs/review-2026-10-15-part9.md) | *The New Dialectic* read in full: part 8 corrected; whether every type can be played and found; Track D; the book PDF |
| [`docs/review-2026-10-15-part8.md`](docs/review-2026-10-15-part8.md) | Walton's dialogue types from Aberdein's survey (corrected by part 9); Walton and Macagno's classification; the Kialo claim-merge bug, fixed |

**The game**

| Document | What it is |
|---|---|
| [`wag/README.md`](wag/README.md) | The Go module: its packages, commands and tests |
| [`docs/game-rules.md`](docs/game-rules.md) | The rulebook, as implemented: the board, moves, modes, scoring |
| [`docs/prototype.md`](docs/prototype.md) | Running and checking the browser prototype, step by step |
| [`docs/protocol.md`](docs/protocol.md), [`docs/protocol/`](docs/protocol/) | The text protocol by which a model or a person at a chat window plays; its JSON schemas and samples |

**Language models**

| Document | What it is |
|---|---|
| [`docs/llm.md`](docs/llm.md) | The dashboard, providers, profiles, keys, limits, traces, annotation runs and their files |

**Schemes**

| Document | What it is |
|---|---|
| [`docs/scheme-registry/README.md`](docs/scheme-registry/README.md) | The cleaned `walton.yml`, the book links (`links.json`), families, pattern signatures |
| [`docs/scheme-registry/nascq-edges.md`](docs/scheme-registry/nascq-edges.md) | Every critical question of the game, and every answer it allows |
| [`docs/walton-audit/REPORT.md`](docs/walton-audit/REPORT.md) | The audit of the Carneades Walton collection |
| [`docs/book-catalogue/catalogue.md`](docs/book-catalogue/catalogue.md) | The book's pp. 1–138 catalogued (contains book text: see issue 6) |

**Kialo annotation**

| Document | What it is |
|---|---|
| [`docs/kialo-pilot/pilot.md`](docs/kialo-pilot/pilot.md) | The first pilot: 20 claims |
| [`docs/kialo-batch2/README.md`](docs/kialo-batch2/README.md) | Batch 2: 50 claims, two annotators, agreement, gold, the game made from it |
| [`docs/kialo-batch2/guidelines.md`](docs/kialo-batch2/guidelines.md), [`adjudication.md`](docs/kialo-batch2/adjudication.md) | The guidelines (version 2, with version 3's changes) and how each disagreement was decided |

**Carneades**

| Document | What it is |
|---|---|
| [`CARNEADES_FORK.md`](CARNEADES_FORK.md) | The fork in `carneades-4/`: its commits, and how to update it |

**Reviews, in order** (each answers one round of the owner's questions and records the decisions):

| Document | What it covers |
|---|---|
| [`review-2026-10-13.md`](docs/review-2026-10-13.md) | Code review of the game; bugs fixed |
| [`part 2`](docs/review-2026-10-13-part2.md) | The game designs, the CLI reviews, Pyon & Ri's NASCQ |
| [`part 3`](docs/review-2026-10-13-part3.md) | The game restated, the browser prototype, Kialo, the Walton collection |
| [`part 4`](docs/review-2026-10-13-part4.md) | A test game, readability, model players, weights, contexts, the Kialo plan |
| [`part 5`](docs/review-2026-10-13-part5.md) | Six points: statements, protocol 2, the registry, the compendium, an audit of the pilot |
| [`part 6`](docs/review-2026-10-15-part6.md) | A third-party audit reviewed, the model layer, the evidence for the schemes |
| [`part 7`](docs/review-2026-10-15-part7.md) | The first real annotation run |
| [`part 8`](docs/review-2026-10-15-part8.md) | Dialogue types (Aberdein 2020); Kialo claims keyed by ID |
| [`part 9`](docs/review-2026-10-15-part9.md) | *The New Dialectic* (Walton 1998); dialogue types as a second aim |

## 7. Repository layout

| Path | What |
|---|---|
| `wag/` | The Go module |
| `wag/cmd/` | `wag-server` (the dashboard), `wag-cli` (terminal games, lint, tournaments), `wag-llm` (headless model games and annotation) |
| `wag/internal/` | `game`, `nascq`, `carneades`, `registry`, `kialo`, `llm`, `annotate`, `server` (with the web page) |
| `wag/data/` | The game's data (`nascq/`, `schemes/`) and the built-in annotation dataset (`annotation/`), embedded in the binaries |
| `docs/` | Everything above; `kialo-batch2/` also holds its data and Python checkers |
| `carneades-4/` | The Carneades 4 fork ([`CARNEADES_FORK.md`](CARNEADES_FORK.md)) |
| `tools/` | `sync_carneades_fork.sh` |

## 8. Sources and licences

- **Carneades 4:** MPL-2.0 (`carneades-4/LICENSE`).
- **The Walton collection for Carneades 4** (`walton.yml`, from `github.com/basakkurtuldu/walton-argumentation-schemes-for-carneades-4`): MPL-2.0 (`docs/scheme-registry/walton-clean/LICENSE`).
- **Walton, Reed & Macagno, *Argumentation Schemes* (Cambridge UP, 2008):** not included. The repository cites it by scheme and question labels and pages, except `docs/book-catalogue/` (issue 6).
- **Walton, *The New Dialectic* (University of Toronto Press, 1998):** cited by page in [`docs/dialogue-types.md`](docs/dialogue-types.md), in paraphrase. `docs/The New Dialectic.pdf` should not be in the repository (issue 6).
- **Aberdein, *Dialogue Types, Argumentation Schemes, and Mathematical Practice* (arXiv:2011.13845, 2020):** cited; not included.
- **Kialo discussions:** public; the datasets keep their claim ids.
- **WAG's own code and data:** no licence yet (issue 6).
