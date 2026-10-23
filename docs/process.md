# How the work is done

This is the system that keeps WAG's data work orderly: who may change what, how a change is checked, how play and data feed back, and where everything is kept. The owner set it up on 2026-10-22 ([review part 13](review-2026-10-22-part13.md)). The plan says *what* is built ([`plan.md`](plan.md), [`catalogue-alignment.md`](catalogue-alignment.md)). This document says *how*.

Parts not built yet are marked **(to build)**.

## 1. One source of truth for each kind of fact

| Fact | Kept in | Changed by |
|---|---|---|
| What a scheme says, its critical questions | the catalogue (`wag/data/catalogue/`), from the book | drafting and verifying entries |
| The formulas | `walton.yml`: WAG's formula layer, which starts from the upstream collection | `walton-clean/tools/edits.py` and `check/run.sh`, with every change in the CHANGELOG |
| The game's fields and links | the catalogue's game fields **(to build: workstream I)** | drafting and verifying, then data and play |
| The game's data (NASCQ, `core.yml`) | generated from the catalogue **(to build: workstream E)** | never by hand, once generated |
| How to code a claim (annotation rules) | the batch's guidelines (`docs/kialo-batch2/guidelines.md`); the rules on opposite consequences are numbered (§3.1, OC1–OC14), each with worked cases in `rule-cases.yml` | editing the rule there, once, with a case for it; a test checks the cases against `walton.yml` and the gold, and rejects a rule's retired wording in the guidelines, the cases, the catalogue's entries and the candidates registry (`TestRuleCases`; proposals and reviews are records and keep their wording) |
| Open decisions | the decision register, [`decisions.yml`](decisions.yml) (§5) | the decider named in each |
| The book's text and other private sources | the private workspace (§6) | the owner |
| Figures (counts, progress) | a generated status report **(to build, §7)** | nobody: documents link to it instead of copying numbers |

## 2. Statuses: who stands behind an item

Every work item in the catalogue carries one:

| Status | Meaning |
|---|---|
| `generated` | made from the sources by `wag-cli --catalogue-gen` |
| `todo` | to write |
| `drafted by WHO, DATE` | written by a person or a model |
| `checked by WHAT, DATE` **(to build)** | passed a check: the validators, plus a model check naming the model, how many checkers, and the brief's version |
| `verified by WHO, DATE` | a person checked it against the sources |
| `disputed by WHO: REASON` | someone, or a game, found a problem (§4) |

**How far a check counts depends on the risk.**
- **Low risk:** examples, contrasts and summaries. These may be accepted when model-checked, once the model checks' measured miss rate is low (§3).
- **High risk:** formulas, question targets and loci, link rules, and the game's fields. These need a person.

**Verifying** follows a rubric for each field, in [`DRAFTING.md`](../wag/data/catalogue/DRAFTING.md) §"Verifying" **(to write)**. A random sample of verified items, about one in ten, is verified again by someone else, and the checker reports the agreement **(to build)**.

## 3. Independent agents

Models draft and check. Each run follows this protocol.

1. **Roles.**
   - A *drafter* writes items.
   - A *checker* finds what is wrong.
   - When the two disagree, a *second checker*, then the owner, decides.
   - No agent checks an item it drafted.
2. **Briefs.** Each role has a brief, versioned in the WAG repository and free of book text **(to move from the workspace)**. A run names its brief and version.
3. **Inputs.** A command builds each scheme's source pack from the workspace, recording the hash of every file **(to build: `wag-cli --catalogue-pack`)**. Every agent of a run sees the same sources.
4. **Outputs.**
   - **Drafters** write patches (`wag-cli --catalogue-apply`, **to build**).
   - **Checkers** write findings in one JSON format: entry, item, severity, finding, evidence, fix. A command validates the format **(to build)**.
   - **Coverage.** Every item in scope gets a finding, or a note saying what was checked and found right.
5. **Independence.**
   - The checker sees the draft and the sources, not the drafter's reasoning.
   - Where possible it is another provider's model. WAG's LLM layer supports Gemini, OpenAI-compatible services and Anthropic.
   - Runs that use the owner's API keys run on the owner's machine, never in a cloud session.
6. **The ledger.** Every finding gets one disposition:
   - applied;
   - applied differently;
   - declined, with the reason.

   Each disposition also records the commit that carried it. Review documents take their counts from the ledger **(to build)**. Until it exists, counts are made by a script and stated with the script.
7. **Measured reliability (to build).**
   - **Seeded faults.** Known faults are planted in a copy of some entries, and the share each checker finds is its recall.
   - **Precision** is the share of its findings that are applied.
   - **A person verifies a random sample** of model-checked items.

   These figures set where model checks are accepted (§2).
8. **A work queue.**
   - Items are claimed before work starts (`drafting by WHO`), so that two agents never do the same work.
   - Batches are small enough to check.
   - Each run's cost (tokens, calls, time) is recorded.

### Review rounds

Since 2026-10-23, after the rounds on proposal D055-D059. Their record showed:
- **The formulas were settled at version 2.** The later rounds worked on the rule for annotators and the effects.
- **Every error after version 1 came from the fix of the round before.**
  - Two came from a reviewer's suggested text taken as written.
  - Two came from a rule stated in several places and changed in one.
- **The findings never ran out,** because each brief asked for new Kialo probes and new candidates.

The rounds now go like this.

1. **One full review.**
   - It reviews everything: the formulas, templates and questions, the examples, the effects and the rules for annotators.
   - It may test the rules on Kialo claims the proposal does not use.
   - The claims it uses become the probe set, listed in its report.
2. **Then verification rounds, scoped to the diff.** Each later brief gives the previous findings, the diff since the version reviewed, and this checklist:
   - **each finding:** made, made in part, not made or not needed, and whether it was made correctly;
   - **each changed passage:** every other place that states the same fact, found by searching for its key terms and for the rule's retired wording;
   - **every check green:**
     - the proposal check and the catalogue check;
     - `validate.py`;
     - the rule's worked cases (`TestRuleCases`);
     - the effects' scripts;
     - `go test`;
   - **the probe set only:** no new Kialo probes, and no search for candidates.
3. **Severity, and when it is done.**
   - **An error blocks:**
     - a false claim;
     - a defect in a formula, template or question;
     - a rule that tells annotators to code something wrongly;
     - a coding in the gold or the worked cases that does not follow the rules;
     - a code bug, or a test that passes when it should fail;
     - a wrong effect, or a missing one.
   - **Warnings and notes do not start a new version.** They go to the owner with the verdict, and the owner decides:
     - which to fix now;
     - which to record, in the register if they raise a question.
   - **A proposal is done** when a verification round finds no errors on the checklist and every check is green.
4. **Fixing a finding.**
   - **Reproduce it first:** run the check, or read the source it cites.
   - **Write the fix in our words.** A reviewer's suggested text is checked like any other: against the sources, the rule's cases and the gold. Record where it came from.
   - **A fix may not depart from the approved text** without the departure being put to the owner as such. The D055-D059 application set its scope rule by a claim's target, against the approved summary and the final check's fix. The gold's cases caught one fault in that rule before the commit, and the independent review of the application found the departure itself unfounded; the rule now follows the approved text.
   - **State each fact once.**
     - A rule lives in one place: for annotation rules, the guidelines, with an id. Other texts cite the id.
     - When a rule's wording changes, the old wording goes on the retired list, which a test enforces (§1).
   - **Effects that change data are scripts,** committed with the change. The numbers a proposal states come from running them.
   - **Run every check** after the edit, before the commit.
5. **Looking for candidates is its own brief,** after the review, against the registry and its relations (§8). A review notes a candidate it meets in passing, but is not asked to find them.

## 4. Feedback from play (to build)

People and models beta-testing WAG report through one channel:
- **Notes attach themselves to what they are on.** A note on a move, an argument or a question records its catalogue IDs: the scheme, the book question, the NASCQ question, the edge.
- **Every category leads to an action:**

  | Category | Action |
  |---|---|
  | wrong scheme or wrong question | review the item; if confirmed, mark it `disputed` |
  | unclear wording | review the paraphrase or template |
  | an illegal move was allowed | a bug in the rules or the data |
  | I wanted to say X and could not | a candidate scheme, question or link (§8) |
  | good | evidence that the item works |
- **Model players** get the same optional field in the text protocol, kept in their traces.
- **A report** (`wag-cli --play-report` over exported games and traces) totals notes per catalogue item. It also counts how often each scheme is played, each question asked, and each answer succeeds.
- **Version records.** Every game records the versions (hashes) of the catalogue and the NASCQ it ran on, so that results stay comparable.

## 5. The decision register

[`decisions.yml`](decisions.yml), made on 2026-10-22, is one machine-readable file of every open decision. [`decisions.md`](decisions.md) is its readable view, generated by `wag-cli --catalogue-gen`. The register gathers:
- the open decisions of the catalogue's entries: every discrepancy whose decision has a sentence beginning "Open";
- [`plan.md`](plan.md) §7;
- [`catalogue-alignment.md`](catalogue-alignment.md) §3 C;
- other documents, such as the registry's known duplicate.

A decision raised in several entries is one decision that lists them all. For example, the burden of practical reasoning's questions is raised in three entries.

| Field | Values |
|---|---|
| id | D and three digits, never reused |
| title, question | in our own words |
| category | what it is about: `scheme` (a scheme's formulas or questions), `vocabulary` (where schemes meet), `game` (the game's fields, texts, edges), `annotation`, `book` (a reading of the book), `metadata`, `project` |
| blocks | the work it holds up: `game`, `verification`, `annotation`, `play`, `release` |
| decider | `owner`, `verifier`, `data`, `play` |
| priority | 1 holds up the next steps of the order; 2, later work; 3 can wait. Proposed when the register was made, and confirmed by the owner on 2026-10-22. |
| status | `open` or `decided`; a decided one records the date, who decided, the decision, and where it is recorded |
| items, docs | where it is raised: catalogue discrepancies (entry, kind, about) and documents |

**What the checker does** (`wag-cli --catalogue-check`; CI runs it):
- every open catalogue decision is listed under an open decision of the register;
- every item the register names exists;
- its own vocabularies are followed;
- a decision that one side marks decided while the other still says "Open" is flagged;
- `decisions.md` is current;
- the progress line counts the open decisions by priority.

**When a decision is taken:**
1. Record the outcome in the register.
2. Update the entries it lists, so that they no longer say "Open".
3. Make the change in the source it concerns: `edits.py` for a formula, the guidelines for annotation.

### Proposals: examples before a scheme changes

A decision that changes a scheme's formulas is prepared as a proposal (the owner's decision of 2026-10-22, with D001). The proposal is written and checked first; `walton.yml` changes only once the owner has read it.

1. **The proposal file,** `docs/proposals/<decisions>.yml`, holds:
   - the schemes as they would be;
   - the templates they need, in our own words;
   - each scheme's questions after the change;
   - what else changes if it is decided;
   - examples, each with its variables bound to phrases;
   - optionally, candidates: schemes sketched beside the proposal but not proposed, to test that its vocabulary extends to neighbouring patterns. They are checked and rendered like the proposed schemes, and never applied to `walton.yml`.
2. **`wag-cli --proposal FILE`** checks it:
   - every variable is declared, used, in a premise, and bound in every example;
   - every predicate has a template;
   - every question targets a formula of its scheme, and a `walton/` question an assumption or exception, as in the catalogue;
   - a book label (`7/CQ1`) is one the compendium prints for an entry linked to the scheme, and no id is given twice;
   - questions are given only for the proposed and candidate schemes, each under its own heading;
   - an example challenges only a premise or assumption that no question of its scheme targets (`CQ:premise_acceptable`, with the formula as its target), as the game allows.

   It also writes the view, `<decisions>.md`. Every formula of every example is rendered in English there, under each question that is pressed against it.
3. **The minimum set of examples** for each proposed scheme (the check warns when one is missing):
   - two uses as an argument;
   - one use as a game answer, if the scheme concludes a bare proposition;
   - every question pressed at least once;
   - a near miss, coded with the scheme it really is.

   The scheme's negated form gets an example where it has one.
4. **The cleaning's gate and the registry** run on `walton.yml` with the proposal applied (`--proposal-walton PATH`, then `walton-check PATH` and `wag-cli --registry PATH`). Only the proposed schemes may differ from today's results.
5. **An independent reviewer checks it** (§3, review rounds): against the book, `walton.yml`, the game and Kialo, with examples of its own, and gives a verdict for each decision.
   - The first review is full.
   - The proposal is revised on its errors, and each later round verifies the diff against the checklist.
   - Warnings go to the owner with the verdict.
6. **The owner reads the view and the reviews.** If it is approved, the change is made through `edits.py`. The examples move into the catalogue's entries, the decision is closed in the register, and the proposal is marked `applied`, so that its view says its "Now" columns now equal the proposed ones. The first proposal, [D001-D028-D029](proposals/D001-D028-D029.md), went this way on 2026-10-22: two independent reviews, then a short check of the third version. The second, [D055-D059](proposals/D055-D059.md), on 2026-10-23: four independent reviews, then a final check of the fifth version. Its annotation rule moved into the guidelines (§1), where the final check's warnings were fixed; the proposal keeps the approved text as its record.

A test keeps every committed proposal valid and its view current.

## 6. The private workspace

`github.com/Austin-Routt/wag-workspace`, private, holds what the public repository may not or need not:
- the owner's copies of the book texts;
- the source packs;
- the Kialo data;
- drafts, findings and run records.
- **Each session restores it.** WAG's session-start hook (`.claude/hooks/session-start.sh`) clones or updates it next to WAG and sets `WAG_WORKSPACE`. Where the session cannot reach it, the hook says so, and the repository is attached with `add_repo`.
- **Durable work goes there;** the session scratchpad is for throwaway files.
- **`sync.py` saves the work.** It archives a folder, leaves out builds, scans for secrets and commits.
- **Tools and briefs move into WAG** as they become part of the process, so a contributor without the workspace can follow it.

## 7. Checks on every change

- **Before a commit.** Run the checks for what changed:
  - `go vet` and `go test`;
  - `wag-cli --catalogue-check --book …`;
  - `validate.py`;
  - `check/run.sh` after a formula change.
- **CI on every push and pull request** ([`.github/workflows/ci.yml`](../.github/workflows/ci.yml), since 2026-10-22):
  - `go vet`, `go test` and the data lint;
  - the catalogue, the register's view and the candidates' view equal what `--catalogue-gen` makes;
  - `--catalogue-check`. Its copy check reads only `docs/book-catalogue`, since the compendium's transcription is private;
  - the registry check, which tolerates only the known duplicate (D042 in the register);
  - the gold's `validate.py`, and the guidelines' worked cases (`TestRuleCases`, in `go test`);
  - the upstream `walton.yml` fetched and checked against its recorded SHA-256, and `check/run.sh` reproducing every committed file.
- **The status report (to build):** one command prints every figure the documents need:
  - entries and items by status;
  - open decisions;
  - schemes and questions in the game;
  - how well the schemes connect;
  - play counts.

## 8. How the set grows

New schemes, questions, exceptions and links come from three places:
- the literature;
- data: Kialo's "other" candidates, the gold, annotation runs ([`plan.md`](plan.md) §3);
- play: "I wanted to say X and could not".

All three go through one registry, [`candidates.yml`](candidates.yml). Its view, [`candidates.md`](candidates.md), is generated by `wag-cli --catalogue-gen`, and `--catalogue-check` checks both.

**What each candidate records.**
- **Its place:** an id (C and three digits, never reused); a kind (scheme, question, exception or link); its relation to what WAG has (below); and a pattern group within a family of `families.json`. Similar candidates sit side by side in their group's table, which shows what each compares, what it concludes, how it relates, what it comes closest to, how often it has been seen and how far it has got.
- **Its evidence:**
  - sources: the literature, and the book's and ND's pages, cited and paraphrased, never copied;
  - sightings in data, each with how well it fits (clean or with strain), a note in our own words, who found it and who verified it. Kialo claims are cited by id; their text stays in the private workspace.
- **What it is not:** the items it comes closest to and why each falls short; related candidates and how they relate.
- **Its form:** a sketch, checked as a proposal's candidates are (§5) and never applied; examples rendered in English, with a near miss; what WAG's engine showed; what implementing it needs; open questions.

**Its relation to what WAG has** (since 2026-10-23). Most patterns found in data are not new schemes but questions, premises or variants of schemes WAG has, so each candidate says which, with its base (the `walton.yml` schemes or candidates it is built from) and why:

| Relation | What it is | Its kind | Usual outcome |
|---|---|---|---|
| new | a pattern of its own: no scheme WAG has is its form | scheme | add |
| variant | a scheme WAG has with a premise, the conclusion or the modality changed | scheme | a domain version, or add it to the base's family |
| specialization | a scheme WAG has (or a candidate) with premises added | scheme | a domain version, or merge into the base |
| composition | two schemes or candidates chained or combined | scheme | merge, as a rule for coding the chain, or reject |
| addition | one premise, assumption, exception or question added to a scheme WAG has | question or exception | merge into the base |
| rule | a rule for annotation or the engine | link | the guidelines or the engine |

- **The formula check.** The check measures each sketch with the registry's patterns (`internal/registry`: predicates by role, the conclusion included; a changed formula counts two), against `walton.yml` and against the candidate's own base, and shows the nearest `walton.yml` scheme in the view. Only schemes that share a formula with the sketch count as near.
  - **An error:** a scheme whose sketch is within one formula of a `walton.yml` scheme. It is a question or exception of that scheme.
  - **Warnings:**
    - a new pattern within two formulas of a `walton.yml` scheme;
    - an addition more than one formula from its base;
    - a variant or specialization more than two changed formulas (4) from its base, or sharing nothing with it;
    - a composition that shares no formula with one of its bases.

    A warning says the sketch is not written on the formulas its relation names: write it on them, or reconsider the relation.
- **The literature check.** Before a candidate is registered, the book's compendium and text and ND are searched for the pattern, and what they say is recorded in its sources, or that nothing was found and what was searched for. A pattern the sources describe is not a discovery: its relation says how it stands to the schemes they give.

**Intake.** Every review brief, annotation run and play report records the candidates it meets in this shape, or one that converts to it. Looking for candidates is its own task, with its own brief (§3), not part of reviewing a proposal. A new sighting of an existing candidate is added to that candidate, not recorded as a new one. A claim that shows two patterns is listed under both, and each listing names the other (`also`); the check notes a claim listed twice without that. A sighting counts only once a second reader has read the text against its note (it is then *verified*).

**The stages,** computed by the check:
- **collecting:** fewer than three verified sightings in two discussions or games (the threshold; whether it is right is D047). A gap shown by analysis rather than by data, such as a scan of `walton.yml` or an engine test, is exempt, with the reason recorded.
- **drafting:** over the threshold, but the minimum is not complete.
- **ready:** the minimum is complete: the threshold; the closest items and why each falls short; and, for a scheme, question or exception, a sketch that passes the check, three examples of its use and a near miss.
- **decided:** a person has decided one of four outcomes:
  - reject it;
  - merge it into an existing item;
  - record it as a domain version of one;
  - add it, as a derived scheme, a question, an exception or a link rule, through a proposal (§5), drafted and checked like any other item.

**Decisions.** A candidate goes to a person through a decision in the register. There is one decision per pattern group where that is possible, so that similar candidates are decided together. A candidate over the threshold without a decision is a warning. The outcome is recorded on the candidate, and the decision is marked decided.
