# Part 9 (2026-10-15): *The New Dialectic*, and dialogue types as a second objective

The owner added Douglas Walton's *The New Dialectic: Conversational Contexts of Argument* (1998) to the repository and asked five things:
1. Review everything critically against Aberdein's paper (part 8) and what has been added to the repository.
2. WAG should be able to play a game in any established dialogue type. It should also be able to find and mine every type in data, negotiations and quarrels included, as it does schemes and critical questions. Is this feasible?
3. How does it change the plan, the README and the other documents? What needs to be updated, added or revised?
4. Should there be a definitive, clear, concise and comprehensive reference based on both texts?
5. Is there a clear, well-defined path to the objective?

The book was read in full: ten chapters and the notes. Page numbers below are its pages.

**Short answers**

1. **Part 8 mostly holds. Five points were wrong or are now out of date,** and all five are corrected in place, with a notice at the top (§1):
   1. "eight types";
   2. debate as persuasion;
   3. what deliberation needs;
   4. what kind of dialogue Kialo is;
   5. "negotiation and eristic: not as modes".

   The book adds a great deal that the paper does not have: commitments, the rules of each type, how to judge shifts, and how to tell the types apart (§2).
2. **Feasible: yes, both halves, at very different costs** (§3).
   - **Persuasion, deliberation and inquiry** fit the current engine.
   - **Information-seeking** needs a seat that holds knowledge.
   - **Negotiation** needs a new bargaining core beside the argument graph. It is the largest piece.
   - **Eristic** should be a detection exercise, not a game that rewards abuse. *(Updated in part 10: the owner decided that research needs a playable quarrel. It comes last, under safeguards.)*
   - **Finding the types in data** is an annotation task like the schemes: typed segments and shifts, with agreement measured. Kialo alone will not supply negotiations or quarrels.
3. **The changes** (§4):
   - a new Track D in the plan, with the scheme catalogue still first;
   - part 8 corrected;
   - the README updated;
   - two misnamed context presets renamed;
   - short pointers in the rulebook and the prototype guide.
4. **A reference: yes, written.** [`dialogue-types.md`](dialogue-types.md) is the one place for definitions, citing the book's pages. It is written in our own words.
5. **A path: yes.** [Plan §8](plan.md#8-track-d-dialogue-types) runs from D0, the reference (done), to D4, the corpora. Each step has its effort and a "done when".
   - **The first code** is to record the type in exports (D1).
   - **The first new rules** are commitment stores in persuasion (D2a), after reading Walton & Krabbe 1995.

**One urgent issue, outside the question:** `docs/The New Dialectic.pdf` is the whole of a copyrighted book, in a repository that is public. It should come out (§5). *(Deleted in part 10; still in the history.)*

## 1. Part 8 checked against the primary source

| Part 8 said | *The New Dialectic* says | Verdict |
|---|---|---|
| Six types, from Walton & Krabbe's grid of initial situation and goal | The same six main types, with subtypes (p. 31) | **Holds** |
| "Not definitive"; Walton's lists "grew to eight" | The eight are in Walton's 1997 table (A20 points to similar tables from 1990 and 1995). The 1998 book keeps six main types: debate is mixed persuasion and eristic, as Walton had already said in 1982 and 1989 (p. 223), and teaching is a subtype of information-seeking (pp. 235–7). The book does not say whether the six are exhaustive. | **Corrected.** The lists differ by presentation, not by a change of view. *(Part 10: an earlier version called it "wrong chronology: six, then eight, then six again", which the sources do not show.)* |
| Types combine by nesting, shifts and mixing | Yes. Chapters 8 and 9 give the vocabulary: external and internal shifts, sharp and gradual, functionally joined or an interruption, licit or illicit, mixed discourse. | **Holds, much extended** (reference §6) |
| "A public debate is persuasion addressed to an audience" | That is the 1997 table's "debate". The 1998 book calls political debate mixed discourse (p. 223). | **Corrected** |
| Classify the dialogue, not the thesis; the thesis does not fix the type | Agreed. The same practical question can be a town-hall deliberation or a partisan debate. Where basic goals conflict, it is a persuasion–deliberation hybrid (ch. 6, note 1). | **Holds** |
| Every WAG game is persuasion | Every game with a central claim (`debate`, `solo`) is; the points modes match no type. `debate` is the simpler one-sided form, where only pro must prove and con may merely doubt (p. 38). WAG has no *dispute* form. | **Holds, sharper** *(part 10: "every game" narrowed to the claim modes)* |
| "Policy deliberation" is not a deliberation, because the game has one fixed proposal and a winner, and deliberation weighs several options with no side owning a thesis | The conclusion holds; the reason does not. In the simplest case deliberation *is* one proposal argued for and against (p. 162, note 4). What the preset lacks is a practical problem, shared goals, symmetric roles and a tentative joint decision. | **Corrected.** The preset is renamed. |
| Kialo is "debate" | The book says nothing about Kialo. It allows a persuasion dialogue about whether an action is advisable (p. 229), and makes a hybrid of deliberation only where basic goals conflict (ch. 6, note 1). | **Corrected.** Our reading: public persuasion, on a practical thesis about action. *(Part 10: an earlier version said "a persuasion–deliberation hybrid", and put our reading in the book's column.)* |
| Negotiation and eristic: not as modes | The owner's new objective replaces this. Also, negotiation does use schemes: offers are backed by practical reasoning, promises and threats (ch. 4). | **Superseded.** Eristic stays a detection exercise. *(Updated in part 10: detection first, then a playable quarrel, last.)* |
| Flag schemes that are improper in the game's type ("proper in") | This is the book's central thesis: a fallacy is a technique misused against the goal of its dialogue (ch. 10). | **Holds, strengthened.** Reference §7 has the table to draft from. |
| WAG has no commitment store | Commitment is the basic idea behind every type (p. 31). Retraction is a key contrast between persuasion (fairly free) and inquiry (minimised, ideally eliminated; p. 70). | **Holds, and matters more.** It comes first in D2. |
| Type per discussion; not per claim | Still right for Kialo. But finding types in other data needs typed segments and shifts. | **Extended** (D3) |

## 2. What the book adds to the paper

The paper gave the map. The book gives the rules and the tests.

| Topic | What the book gives | Where it goes in WAG |
|---|---|---|
| **Commitment** | Assertions and concessions; light and dark side; retraction, and its limits by type (pp. 43–62, 70) | Commitment stores (D2a) |
| **Persuasion** | Four minimal rules (p. 43). A dispute against one-sided doubt (p. 38). The critical discussion and how it differs (pp. 46–9). Permissive and rigorous forms, one embeddable in the other (pp. 53–7). The maieutic function (pp. 57–9). Enthymemes (pp. 60–3). | D2a; the `debate` mode's place |
| **Inquiry** | Cumulative, with retraction minimised; evidential priority; no circularity; three stages; relevance by stage (pp. 33, 69–99) | D2c |
| **Negotiation** | Four subtypes; interests as dark-side commitments; Donohue's response rules; threats and their place; shifts out of deadlock (ch. 4) | D2e |
| **Information-seeking** | Asymmetric roles; presuppositions as commitments; direct and indirect answers; expert consultation and probing it (ch. 5) | D2d |
| **Deliberation** | The practical problem; the full practical inference; the simplest case is pro and contra; tentative closing; caution where harm is possible (tutiorism); joined to information-seeking (ch. 6) | D2b |
| **Eristic** | What a quarrel is for; a closed attitude; four subtypes; five identifying characteristics (ch. 7) | D2g, D3 |
| **Shifts** | A taxonomy, and a test for licit shifts: judged afterwards, by the first dialogue's goals; agreed or not; open or concealed (ch. 8) | D2f, D3 |
| **Mixed discourse** | The supposed type and the enacted type can differ. Evaluation is conditional. The speech event. (ch. 9) | The annotation fields `type_supposed`, `type_enacted` and `setting` |
| **Method** | The four-step method; the six relevance factors (pp. 249–55) | The identification guidelines (D3) |

**What the book does not give, and where to look:**
- **No formal protocols.** Those are in Walton & Krabbe 1995, which must be read before coding, and in later work (reference §10).
- **No coding manual and no agreement figures.** Reliability is ours to measure.
- **Little on how a negotiation's outcome should be scored.**

## 3. Is the objective feasible?

### 3.1 Playing every type

**Yes.** The requirement is a **protocol layer** above the argument engine. A type becomes a protocol with:
- roles and moves;
- what each move does to the commitment stores;
- the response duties and the stages;
- how it ends and what it yields.

The existing engine (schemes, critical questions, Carneades) stays the argument layer for every type that argues.

| Type | Fit | What is new | Effort |
|---|---|---|---|
| Persuasion, with commitments | High | Commitment stores, retraction, rule 1, a dispute form, stages | 1 week |
| Deliberation | High | A shared goal, symmetric roles, n-option issues, a tentative joint outcome | 1 week |
| Inquiry | High | One team, no retraction, required sources, stages | Days |
| Information-seeking | Medium | A seat that holds knowledge; presuppositions | 1 week |
| Negotiation | Low | A bargaining core: issues, private interests, offers, deadlock | 2–3 weeks |
| Shifts and embedding | Medium | A stack of dialogues; carrying results back | 1 week |
| Eristic | — | A detection exercise only *(part 10: then a playable quarrel, last, under safeguards; 1–2 weeks)* | Days |

**The risk is scope, not possibility.** Six protocols plus shifts is two to three months of focused work. The plan therefore adds one type at a time, each playable and tested before the next. Persuasion comes first, because every game today is persuasion, and most of the other types build on its commitment store.

### 3.2 Finding every type in data

**Yes, as an annotation task modelled on the schemes:**

| For schemes | For dialogue types |
|---|---|
| The scheme catalogue | A dialogue-type catalogue: one entry per type and subtype (D1) |
| Critical questions | Walton's identification questions, and the test for licit shifts (reference §8.1, §6.2) |
| One annotation per claim | One record per dialogue, per segment and per shift (reference §9.3) |
| Model annotators, a reviewer, people's verdicts | The same machinery, as a new task in the annotation package (D3) |
| Agreement per field | Kappa on the type; segment boundaries; shifts found; licit or not |

**What to expect.**
- **The main type of a dialogue,** where genres differ: probably good agreement.
- **Segments and shifts:** moderate.
- **Whether a shift was licit:** hardest, because it needs the whole context. Walton evaluates conditionally: *if* the speakers can be taken as having a critical discussion, *then* an argument commits a given fallacy, and the speaker can answer that the exchange was partly eristic (pp. 224–5). The annotation should allow such conditional verdicts.

**The data limit is real.** Kialo gives persuasion, much of it about action, with almost no negotiation, quarrels or inquiry. Plan §8 lists candidate corpora per type; none has been checked here for licence or fit.

### 3.3 What it is not

Walton's types are **normative models**, not a partition of real talk. A dialogue labelled "negotiation" is a claim about the goals and rules its participants were bound by, supported by evidence. It is not a fact read off the surface. The annotation records what was *supposed* and what was *enacted*, and allows mixtures. The game, by contrast, *does* enforce one type's rules at a time; that is what makes a type playable.

## 4. What changed in this round

| File | Change |
|---|---|
| [`docs/dialogue-types.md`](dialogue-types.md) | **New.** The reference: the classification and its sources; shared concepts; each type; shifts and mixtures; the arguments whose standing depends on the type, with their `walton.yml` ids; recognising types in text; WAG against the book; feasibility; a data model; further reading |
| [`docs/review-2026-10-15-part8.md`](review-2026-10-15-part8.md) | A correction notice, and the five points fixed in place |
| [`docs/plan.md`](plan.md) | **New §8, Track D:** principles, steps D0–D4 with effort and "done when", order, candidate corpora. **Also:** guidelines step 0 records setting and type, supposed and enacted; phase 5 points to Track D; decisions 1 (the PDF) and 6 (Track D's open choices) updated |
| [`README.md`](../README.md) | The second objective in the aim; status; issues (the PDF, dialogue types); roadmap; the document map |
| `wag/data/contexts.json` | "Policy deliberation" and "Scientific inquiry" renamed "Policy question" and "Scientific question". The descriptions say that the game is still persuasion. The ids (`policy`, `scientific`) and weights are unchanged, so saved games and exports are unaffected (the presets are not part of the data hash). The `game` and `server` tests pass. |
| [`docs/game-rules.md`](game-rules.md), [`docs/prototype.md`](prototype.md) | One paragraph each: the modes in Walton's terms; the presets are domains |

**No other code changed.** D1, recording the type in settings and exports, is the first code step. It is small, but it belongs with the catalogue skeleton, so that the type and the "proper in" field arrive together. *(Part 10: the owner put Track D after the Kialo gold. Only "proper in" and the guidelines' step 0 come earlier, with phases 1 and 2.)*

## 5. The book in the repository

- **What it is.** `docs/The New Dialectic.pdf` (17.5 MB, commit f891090) is a PDF of the whole book, published by University of Toronto Press in 1998 and in copyright.
- **Where it is.** The repository is public (checked on 2026-10-15), so anyone can download the book from it.
- **Why it is a problem.**
  - It goes against the project's own rule: no book text in the public repository ([plan §1](plan.md#1-principles)).
  - Publishing a whole copyrighted book is very unlikely to be allowed under fair use or fair dealing. This is not legal advice.
- **The options:**
  1. **Delete the file in a new commit.** It leaves the current tree, but stays in the history and in any clone or fork.
  2. **Remove it from the history** (for example with `git filter-repo`), then force-push. GitHub's guide on removing sensitive data also covers cached views. This rewrites every later commit, so only the owner should decide it.
  3. **Make the repository private.** Option 1 or 2 is still advisable.
- **What we did.** We did not delete it: the owner added it, and the choice of option is theirs. Option 1 can be done on request. *(Part 10: the owner chose option 1, and the file was deleted. Removing it from the history remains the owner's decision.)*
- **Nothing in WAG depends on the file.** The reference cites pages and paraphrases. The text extracted to read it stays outside the repository.

The same decision covers the transcribed scheme-book text already flagged in [plan §7](plan.md#7-decisions-the-owner-needs-to-make), item 1.
