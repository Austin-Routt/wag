# Plan: complete coverage and alignment of the scheme catalogue

**Started 2026-10-22, after tier 1 was drafted.** This plan expands [plan §2](plan.md#2-phase-1-the-aligned-catalogue), the aligned catalogue, from the 32 tier-1 schemes to every scheme. It lists what is left and in which order, and it states when the catalogue is complete and aligned.

**The catalogue's own README has the current figures:** [`wag/data/catalogue/README.md`](../wag/data/catalogue/README.md). The figures below were measured on 2026-10-22, from the committed catalogue and the upstream collection.

## 1. What "complete" and "aligned" mean

**Complete.**
- Every scheme of the book's compendium has an entry, and so does every scheme of `walton.yml`.
- Every entry has been drafted and then verified by a person.
- Every critical question of the book is in an entry, and is decided:
  - it targets the formulas it puts in doubt;
  - or it is marked as attacking the inference;
  - or the formula it needs is recorded as missing.
- Every assumption and exception of `walton.yml` has a question that asks about it.

**Aligned.**
- **Book and `walton.yml`.** Where they differ, the entry records the difference and the decision. No decision is left open.
- **Catalogue and `walton.yml`.** The formulas are `walton.yml`'s. A formalization that the upstream collection has but `walton.yml` lost is either brought in or recorded as rejected.
- **Catalogue and the game.** The NASCQ's schemes and questions are generated from the catalogue, and the game's tests pass unchanged (plan §6).
- **Catalogue and the other WAG data.**
  - `links.json` and `families.json` are generated from the catalogue, or retired.
  - The annotator's prompt and guidelines take their scheme descriptions from the catalogue (plan §3).

## 2. Where things stand (2026-10-22)

| Item | State |
|---|---|
| Entries | 115: one for each of `walton.yml`'s 112 schemes, and one for each of the 3 compendium schemes it lacks. All 109 compendium schemes are linked to an entry. |
| Tier 1 | 32 entries, drafted on 2026-10-22, checked by an independent model, not yet verified by a person. |
| Tier 2 | 83 entries, generated only. |
| Book questions | All 153 printed questions are in some entry. Several book questions target no formula and are not marked as undercutting; the checker lists them. |
| Questions for assumptions and exceptions | 10 assumptions and exceptions in 9 schemes had no question before tier 1 was drafted. Tier 1 adds `walton/` questions for its five. |
| Entries without any question | 47: the book prints none for them, and the game asks none. |
| Generated items still unverified | About 840: book links, families, premise maps and question targets. |
| Discrepancies | Before drafting, the generator found 17 partial links, 3 compendium schemes without a formalization, 1 `walton.yml` scheme not in the book, and 1 game difference. Drafting tier 1 added the book-versus-`walton.yml` differences it found. |

### What the upstream collection holds that `walton.yml` lacks

The upstream collection, `walton-argumentation-schemes-for-carneades-4-master`, has the master `walton.yml` and 13 category files. Against the master:
- 106 category-file schemes are identical to it.
- 6 differ.
- 4 are only in the category files.

What matters for alignment:

| Finding | Where | What it means |
|---|---|---|
| A second `negative_consequences` block, formalizing compendium 33.3 (reasoning from negative consequences: an agent's act leads to an outcome bad for the agent's goals) | `A_reasoning/3_practical-reasoning.yml` | It is lost from the master, apparently because its id is a duplicate. **Brought in on 2026-10-22** as `negative_consequences_reasoning` (C1 below). |
| A second `practical_necessary_condition` block, which is the sufficient-condition schema | same file | The same duplicate-id defect. The master already has it as `practical_sufficient_condition`, so nothing is lost. |
| `whole_into_parts` | `A_reasoning/2_inductive-reasoning.yml` | A formalization of compendium 10.2, division of the whole into its parts, which `walton.yml` lacks. |
| `full_slippery_slope` | `A_reasoning/5_casual-reasoning.yml` | A formalization of compendium 53, the full slippery slope, which `walton.yml` lacks. It has no source, and its formulas need checking (`accept_all_steps` and `led_to_step` have no arguments). |
| `opposition_descriptive_2`, `_3` | same file | The master's `opposition-descriptive_2`, `_3`, spelled with an underscore. Nothing is lost. |
| `commitment1` with `committed_to` | `B_source-based-arguments/2_…` | The master misspells it `commined_to`; the category file is right, and the cleaned `walton.yml` already has the correction (its CHANGELOG). |
| `part_in_whole` with variables `[Y,X,A]` | `A_reasoning/2_inductive-reasoning.yml` | The master declares `[Y,X,P]`, though its formulas use A and never P. The category file is right, and the cleaned `walton.yml` already has the correction. |

**The one compendium scheme with no formalization anywhere** is 50, the precedent slippery slope.

## 3. Workstreams

### A. A person verifies tier 1

**The work.** 32 entries, about 20 to 30 minutes each with the book: 10 to 15 hours. [`DRAFTING.md`](../wag/data/catalogue/DRAFTING.md) §"Verifying" says how.

**Before it starts:** the owner's decisions in workstream C that change tier-1 formulas (N8, N12). Otherwise the verifier would verify entries that are about to change.

### B. Draft tier 2

83 entries. Use the same pipeline as tier 1:
1. a source pack per scheme;
2. `DRAFTING.md`;
3. drafting agents by family, each in its own copy of the catalogue;
4. an independent model check against the book, the NASCQ and the upstream files;
5. a person's verification.

**Order, as plan §2 has it:**
1. the schemes annotators propose as near misses, and the "other" candidates;
2. then whole families, so that contrasts are written in both directions at once.

**Effort.**
- **Drafting:** four to six agent sessions.
- **The model check:** two sessions.
- **Verification by a person:** 30 to 40 hours.

### C. Decisions on the sources (the owner)

Each of these changes `walton.yml`, `links.json` or `families.json`. The catalogue records them as open discrepancies until they are decided.

1. **33.3: decided on 2026-10-22.** The upstream second `negative_consequences` block is in `walton.yml` as `negative_consequences_reasoning`:
   - variables filled in;
   - templates in our own words;
   - recorded in the cleaning's CHANGELOG under "Added by the owner's decision".

   `links.json` links 33.3 to it exactly, `families.json` names it, and the catalogue has it as a tier-1 entry.
2. **10.2 and 53:** bring `whole_into_parts` and `full_slippery_slope` into `walton.yml`, after checking their formulas; or keep the two entries without a formalization.
3. **50, the precedent slippery slope:** write a formalization, which must pass Hansen's definition (plan §2); or leave it unformalized.
4. **N8, the modal cause to effect:** a variant in the causal family for causes that might occur, or have not yet (`docs/review-2026-10-13-part5.md` §7).
5. **N12, analogy:** separate source and target propositions.
6. **The upstream defects** (duplicate ids; the master's typo and wrong variable list, already corrected in the cleaned file): record only. `walton.yml` is already right.
7. **`walton.yml`'s English templates.** Rewrite in our own words the 173 that repeat the book (plan §7, item 1). The catalogue does not store them, but the game and `wag-cli --catalogue-show` render them.

### D. Decide every book question and cover every assumption and exception

- **Book questions with no target.** For each, choose one:
  - targets;
  - "attacks the inference";
  - or a `book-walton` discrepancy naming the missing formula.

  The checker warns until it is decided.
- **Uncovered assumptions and exceptions.** Give each a `walton/<predicate>` question. There are five left after tier 1, in tier-2 schemes.
- **Uncovered premises: no new questions.** The game's challenge move lets anyone question any premise that no question covers (game rules §2).

### E. Generate the game's data from the catalogue

This is plan §6.

1. Write a generator from verified entries to the NASCQ's `schemes.json`, `critical_questions.json` and `edges.json`:
   - the question texts from the paraphrases;
   - the loci from the targets: an assumption is asked at the premise, as now;
   - the patterns from the formulas.
2. Its first output must reproduce today's NASCQ except where the catalogue records a reason. Known reasons:
   - two of expert opinion's questions narrow the book's;
   - the game's copy of `sign` lacks the exception `more_reliable`.
3. The game's tests must pass unchanged.

**Then** the game can take the other verified schemes, family by family, with their ask and backup edges linted as now.

### F. Retire the drafted link files

Once an entry's book links and families are verified, the catalogue is their authority. Then:
- generate `links.json` and `families.json` from the catalogue, so that the registry and other readers still find them;
- or move those readers to the catalogue and delete the files.

**Until then,** a correction made in the catalogue is also made in `links.json` or `families.json`, so that the generator agrees. The checker reports the cases where they differ.

### G. Tools

- **Checker.**
  - Warn when a contrast has no contrast back in the other entry.
  - Warn when a question printed once under a heading (such as 33/CQ1–3) reads differently in the entries that share it.
  - Count the open ("Open:") discrepancy decisions in the progress line.
- **A verification view in the dashboard.**
  - It shows an entry next to the book's text, which the verifier loads from a local file. The text is never stored or committed.
  - It has buttons to verify or dispute each item.
  - It writes the status and a history line.
- **Drafting helpers.**
  - Move the patch helper used for tier 1 into `wag-cli` (`--catalogue-apply`), so that people and models draft the same way.
  - Move the source-pack builder into `wag-cli` too. It reads the book's text from a local file given on the command line.
- **The annotator's prompt from the catalogue** (plan §3):
  - each scheme's summary, contrasts and near misses;
  - its questions in our own words.

### H. The book's chapters 4 to 8

Only chapter 9 (the compendium) and pp. 1–138 are transcribed, outside the repository. Chapters 4–8 (pp. 139–307) discuss:
- the causal schemes (ch. 5);
- the attack and rebuttal of schemes (ch. 6);
- their formal models (chs. 7–8).

Two ways to cover them:
- the person verifying reads them with the book open;
- or the owner transcribes them locally, outside the repository, so that the model check and the copy check can use them.

## 4. Order

1. **The owner decides C2 to C5,** which change formulas, before anyone verifies the entries they touch (C1 was decided on 2026-10-22). C6 and C7 can wait.
2. **A person verifies tier 1** (A).
3. **Generate the game's data** for the 9 game schemes from the verified tier 1 (E), and run the game's tests.
4. **Tier 2 in batches** (B), each with its questions decided and its assumptions and exceptions covered (D).
5. **Tools** (G) as the batches show the need. The verification view helps most before tier 2's verification starts.
6. **Retire the drafted link files** (F) when every entry's links are verified.
7. **Rewrite the English templates** (C7) before any public release of `walton.yml` with WAG.

## 5. Done when

- `wag-cli --catalogue-check --book …` shows:
  - no errors and no warnings;
  - every item verified;
  - no discrepancy decision beginning "Open:".
- The NASCQ is generated from the catalogue and the game's tests pass.
- `links.json` and `families.json` are generated from the catalogue, or gone.
- Nothing the repository keeps reproduces the book: the catalogue's copy check passes, and `walton.yml`'s English is our own.
