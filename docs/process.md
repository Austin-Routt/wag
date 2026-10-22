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
| Open decisions | the decision register **(to build, §5)** | the decider named in each |
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

## 5. The decision register (to build)

One machine-readable file of every open decision. It gathers:
- the "Open:" decisions in catalogue entries;
- [`plan.md`](plan.md) §7;
- [`catalogue-alignment.md`](catalogue-alignment.md) §3 C.

| Field | Values |
|---|---|
| id, question, where it was raised | |
| category | game rules; verification; annotation; shared vocabulary; needs the book's ch. 4–8; metadata |
| blocks | the work it holds up |
| decider | the owner, a verifier, data, or play |
| priority, status, date decided, and where the decision is recorded | |

The checker counts the register's entries and checks that every "Open:" in an entry is in it.

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
- **CI on every push (to build):** the tests, the catalogue check, the gold's validation, and a check that `check/run.sh` rebuilds `walton.yml` unchanged.
- **The status report (to build):** one command prints every figure the documents need:
  - entries and items by status;
  - open decisions;
  - schemes and questions in the game;
  - how well the schemes connect;
  - play counts.

## 8. How the set grows

New schemes, questions and links come from three places:
- the literature;
- data: Kialo's "other" candidates, [`plan.md`](plan.md) §3;
- play: "I wanted to say X and could not".

All three go through one candidates file. A candidate seen at least three times in at least two discussions or games goes to a person, who decides one of four outcomes:
- reject it;
- merge it into an existing item;
- record it as a domain version of one;
- add it, as a derived scheme, a question or a link rule, drafted and checked like any other item.
