# WAG review, part 3: the game restated, a browser prototype, Kialo, and the Walton collection

Date: 2026-10-13. This follows `review-2026-10-13.md` and `review-2026-10-13-part2.md`. It answers the questions of the third message. How to run and check the prototype is in `prototype.md`.

| Question | Short answer | Section |
|---|---|---|
| Do I understand the game? | The restatement below is how the code now works: natural language in, an official scheme out, then the scheme's variables, its critical questions through ask edges, and answers through backup edges. The prototype lets you check that visually. | 1 |
| What did "players' own words, scheme names and 'Round'" mean? | Display changes only: the board shows statements in the words players typed into the scheme's variables, not as atoms like `should_be_performed(carbon_tax)`. It meant nothing about bypassing schemes. | 2.1 |
| How were "several proposed code fixes wrong"? | That was about the code the three AI reviewers proposed in `CLI_Reviews.md`, not your six patches, which were verified and applied. Examples with evidence are in 2.2. | 2.2 |
| An "Other" scheme behind a flag? | Done: `AllowOther`, with three generic critical questions. *Export* lists Other arguments for mining. | 3.1 |
| Evidence? | Done: each scheme says whether sources are required, recommended or not needed. `RequireEvidence` enforces "required". | 3.2 |
| A moderator option? | Done: `Moderated`. The moderator approves or rejects moves, optionally sets the weight, and can accept or reject statements as in the CAF. | 3.3 |
| Can I make a visual prototype for Windows, and is a server–browser model a good idea? | Yes and yes: `wag-server.exe`, or `run-wag-server.bat` from the repository. | 4 |
| Is the Kialo visualizer helpful? | Yes, as data and as a layout. WAG imports Kialo discussions as games of "Other" arguments. The board now uses Kialo's claim-centred pros/cons layout. | 5 |
| The Walton collection | The "expanded" zip is byte-identical to the earlier one. It does not load in Carneades as it stands. The audit and a 10-step cleanup plan are in `walton-audit/`. That is the next phase, after you have checked the prototype. | 6 |

## 1. The game, restated

A move in WAG goes like this. Each step is enforced by the code, not only displayed.

1. **The player writes the argument in natural language.**
   - The text is kept with the argument and shown as a quote on the board and in the history.
   - It has no logical force of its own.
2. **The player chooses the scheme it instantiates** from the official list of Walton schemes in the NASCQ.
   - Today that is 7 schemes: practical reasoning, positive consequences, values, sign, expert opinion, position to know, analogy.
   - "Other" is offered only when the game allows it.
   - The page can suggest schemes, but the player chooses.
3. **The player breaks the argument into the scheme's variables.**
   - Practical reasoning has `S1` (current circumstances), `A` (action), `S2` (resulting state), `G` (goal) and `V` (value). Each variable has a semantic type: state, action, goal, value, source, domain and so on.
   - The engine instantiates the Carneades scheme with those values. The result is the premises, the conclusion and the exceptions, as statements in the graph.
   - The scheme's English language templates (`schemes/core.yml`) render each statement in the player's words.
4. **The opponent questions the argument through the critical questions of that scheme.**
   - The NASCQ's **ask edges** link a scheme to its CQs.
   - Each edge says what the question is about (its locus):
     - a premise, which the proponent must defend;
     - an exception, which the questioner must prove;
     - the conclusion, a counterargument;
     - the inference, an undercutter.
   - A CQ that is not linked to the argument's scheme cannot be asked of it (rule S3).
5. **An answer is itself an instance of a scheme.**
   - The NASCQ's **backup edges** say which schemes can back an answer to which CQ. Thesis-independent schemes (expert opinion, position to know, analogy) can back any answer (rule S5).
   - The answer fills the backing scheme's answer variable. The player fills in the rest of its variables.
   - A negative answer attacks; a positive answer defends.
   - The answer is a new argument with its own CQs, so a dispute can go as deep as the players take it. Each level runs through the NASCQ links of the scheme used there.
6. **Carneades labels the graph after every move.**
   - Every statement becomes In, Out or Undecided.
   - Issues resolve by their proof standard, and undercut arguments are out of force.
   - The mode decides what the labels are worth: points, the verdict on a central claim, or surviving questions.

**Semantic matching.** The NASCQ paper (Pyon & Ri 2022) ties backup schemes to the *semantic nature of the thesis* an answer must prove. A question that asks for another action is answered with an action scheme. Thesis-independent schemes back any thesis.

WAG uses this in two places:

- **Backup edges carry it.** The data lint warns about a thesis-dependent backup edge between different semantic types (`wag-cli --lint`). In the prototype, the scheme list for an answer shows only the linked and the thesis-independent schemes, each marked as such, with its semantic type.
- **Scheme suggestions for free text are keyword cues (`nascq.Suggest`), not semantic understanding.**
  - The better design follows the paper: first classify the semantic nature of the claim (an action? a fact? a classification? a value judgement?), then offer the schemes that conclude that kind of thesis, then fill their variables.
  - An LLM or an argument-mining classifier can do this behind the same `/api/suggest` endpoint. That is future work, and it gains a lot from the NASCQ expansion: with 7 schemes there is little to choose from.

The prototype's first screens show this sequence: free text, *Which scheme fits?*, the scheme card, the variables, the preview. Its CQ panels show ask and backup edges at work. `prototype.md` §2.1 walks through it.

## 2. The two confusing statements

### 2.1 "The board now uses players' own words, scheme names and 'Round'"

This was shorthand for three **display** fixes. The CLI reviewers had reported the problems they fix:

- **"Own words".** A statement used to be printed as a Carneades atom such as `should_be_performed(carbon_tax)`, or as "a1 is undercut." It is now printed through the scheme's English template, filled with the values the player typed into the variables: *Action carbon tax should be performed.* The words are the player's; the structure is still the scheme's.
- **"Scheme names".** The board says *Practical Reasoning* instead of the internal ids `AS:practical_reasoning` or `practical_reasoning1`.
- **"Round".** The counter labelled "Turn N" counted rounds, where every player moves once. It is now labelled "Round N".

It did not mean that players argue in free text without a scheme. That only happens with the optional "Other" scheme (3.1).

### 2.2 "Several of the proposed code fixes were wrong"

This refers to the code the three AI reviewers (Claude, DeepSeek, Gemini) proposed in `CLI_Reviews.md`, which you sent in the second message. **It does not refer to your six patches of 2026-10-12.** Those were applied and verified in part 1: determinism, exception roles, presumptions, exact rollback, the issue-aware cycle check, and the switch to the shared Carneades fork. `TestCoreUndercutterViaCQ` failed in 3 of 10 runs before them and passes 30 of 30 after.

The reviewers' observations of the CLI were accurate, and most were fixed (part 2, §1.3). Several of their diagnoses and patches were not:

| Proposal in `CLI_Reviews.md` | Why it is wrong | Evidence |
|---|---|---|
| "Carneades doesn't set `IsUndercutter` on the generated undercutter" | It does: `InstantiateScheme` creates exception statements with `IsUndercutter: true`. | `carneades-4/src/engine/caes/caes.go:675`; a probe showed the flag set |
| An off-by-one in `advanceTurn` | The move is recorded before the turn advances. | the reviewed commit, `turn.go` |
| `//go:embed ../../data/...` to embed the data | Go forbids `..` in embed patterns, so this does not compile. WAG embeds from package `data` instead. | the Go spec for `//go:embed` |
| Thesis modes that change labels directly, or add the claim as an assumption | Assumptions are In unconditionally, so the claim could never be attacked; changing labels by hand bypasses Carneades. | part 2, §1.3; the debate mode now uses `MakeIssue` and a proof standard |
| "Defend" templates that conclude the scheme's own conclusion | A defence must conclude the attacked premise, or deny the exception. Concluding the scheme's own statement was the original defect. | part 2, §1.3; `answerTemplates` |
| Labellers that count votes | They replace Carneades' grounded labelling and proof standards with something else. | part 2, §1.3 |
| "Fixed-seed determinism bug", "zero-weight bug" | Neither is a bug: the ties come from the dominant strategy, and a rebuttal that loses under PE scores 0 by design. | probes in part 1 |

One proposal was sound and is implemented: delta scoring, which is WAG's `live` scoring.

## 3. New options

### 3.1 The "Other" scheme (`AllowOther`, off by default)

- **What it is.** "Other" (`AS:other`, Carneades scheme `other`) lets a player write one to six premises and a conclusion in their own words.
- **Its questions.** It has no Walton CQs, so it gets three generic ones:
  - *Is premise n acceptable?* (premise locus, one per premise);
  - *Are the premises relevant to the conclusion?* (an undercutter);
  - *Is there a counterargument?* (a rebuttal).
- **Answers.** It is thesis-independent, so it can back the answer to any CQ. Its own CQs have no backup edges yet, so the thesis-independent schemes answer them: expert opinion, position to know, analogy, and Other itself. Free premises have no semantic type, so there is nothing yet to match a thesis-dependent scheme against.
- **Mining.** Its arguments are kept apart in the game export (`other_arguments`). The intended use is to collect arguments that do not fit the current schemes, and later classify them or propose new schemes.
- **Turning it off.** With the flag off, the scheme does not exist in the game's NASCQ, and moves naming it fail rule S1.

### 3.2 Evidence (`RequireEvidence`, off by default)

- **The scheme data.** Each scheme in `nascq/schemes.json` carries `evidence: required | recommended | ""` and a note saying what counts.
  - **Required:** expert opinion and position to know. A source-based argument without its source is just an assertion.
  - **Recommended:** consequences, sign and analogy.
- **On the board.** Sources are stored on the argument and shown on its card.
- **Enforcement.** With the flag on, a "required" scheme played without a source is rejected under rule E1 and costs the usual penalty.
- **What is not checked.** WAG does not judge a source's quality. That is the moderator's job (3.3), or the opponent's, through the scheme's own CQs: *Is E an expert?*, *Is E's assertion based on evidence?*.

### 3.3 Moderator (`Moderated`, off by default)

This follows the CAF of Gordon & Walton (2006): "at some point a decision will be made to either accept or reject some statement at issue … Accepted statements hold and rejected statements do not hold, no matter what arguments have been made or what proof standards apply." The moderator plays the CAF's decision-maker, the audience of *Formalizing Informal Logic*.

- **Holding moves.** Every move is validated, then held as *pending*. The moderator either:
  - approves it, optionally setting the argument's weight in place of the scheme's default; or
  - rejects it with a note, for example for an irrelevant source. The usual invalid-move penalty applies.
- **Deciding statements.** The moderator can rule on any statement at any time:
  - **accept:** the statement is assumed (a Carneades assumption, In unconditionally);
  - **reject:** its opposite is assumed;
  - **clear:** the ruling is withdrawn.
- **Recording.** Rulings appear in the history and on the statement, so anyone can check them against the arguments, as the CAF requires.
- **Limit.** In the prototype the moderator is whoever ticks *I am the moderator*. There are no accounts (`prototype.md` §3).

## 4. The prototype, and why server–browser

**Yes, a server–browser model is the right choice for now:**

- **One engine.** The browser shows the game and the Go server plays it. The rules, validation, labelling and scoring that the CLI and the tests exercise are the same code.
  - A second implementation in JavaScript would drift.
  - Compiling the engine to WebAssembly is blocked without changing the fork. Carneades' `caes` package includes `swichr.go`, which calls SWI-Prolog through `MakeSWIPrologCmd`. That function is defined only for Linux, macOS and Windows, so `GOOS=js GOARCH=wasm` does not build.
- **Windows without installs.** `wag-server.exe` is one 7 MB file with the page and the data built in. It opens the browser on start.
- **Several viewers.** Players and a moderator can watch the same game in several tabs, or on several machines with `--addr`.
- **An API for AI players.** An LLM or Python player uses the same HTTP API as the page.

The costs: games live in the server's memory, and there are no accounts. Both are fine for validation and need work before any shared deployment.

`prototype.md` explains how to run it and what to check, step by step. These flows were checked in a headless browser:

- setup;
- arguing with free text, a suggestion and a preview;
- the AI's reply;
- CQ panels;
- a defence that denies an exception with an expert;
- the AI undermining that defence;
- the NASCQ network;
- required sources;
- a moderator hold and approval with a weight;
- statement rulings;
- the Kialo imports.

## 5. Kialo

**Is the visualizer helpful? Yes, in two ways.**

1. **Data.** Its five captured discussions (`data/<id>/graph.json`) are exactly the input a converter needs:
   - claims with text and author;
   - support and attack relations between placements;
   - impact votes;
   - archive flags.

   `internal/kialo` converts a discussion into a WAG game:
   - every claim becomes a statement;
   - every pro becomes an "Other" argument for its parent, and every con an "Other" argument for the parent's negation;
   - the weight is the mean impact vote divided by 4 (0.5 with no votes);
   - claims are presumed;
   - the thesis becomes the central claim, at issue under PE;
   - sides alternate by depth;
   - archived relations and cycles are skipped and reported.

   All five discussions import. The largest, *Does God exist?*, gives 3,690 arguments in under a second. Carneades labels its thesis Out.

   Every imported argument then has the generic CQs, so a player can press any Kialo claim. Reclassifying an imported argument into a Walton scheme is the argument-mining step this enables. It is not built yet; it gains most from the NASCQ expansion.
2. **Layout.** "WAG's visualization could be more streamlined": agreed. The WAG board now follows the visualizer's reader layout:
   - one selected claim at a time, with breadcrumbs;
   - *For it* and *Against it* columns;
   - click a premise to make it the focus;
   - an outline map on the left.

   **What WAG adds** is the structure inside each card: the scheme, the premises with their labels, why the argument is out of force, and the CQ chips with their state.

   **What can be streamlined next:**
   - collapse premises and CQs until a card is selected;
   - draw the map as a tree or sunburst, as the visualizer does;
   - reuse its archived-placement and comments views for imported games.

   These are worth deciding after you have looked at the prototype.

## 6. The Walton collection and the NASCQ expansion (next phase)

The audit is in `walton-audit/REPORT.md`, with the data in the same folder. In brief:

- **The upload.** The "expanded" `walton-argumentation-schemes-for-carneades-4-master` is byte-identical to the one sent earlier (same SHA-256).
- **It does not load.** `walton.yml` has 111 schemes but fails to load in Carneades (a YAML error at line 733). After six one-line syntax fixes:
  - the validator reports 137 problems;
  - 7 schemes build no argument;
  - 4 build non-ground statements.
- **Content errors.** For example:
  - falsification concludes ¬A;
  - witness testimony's `in_domain(D,S)`;
  - division copies composition.
- **Coverage.** It covers 58 of the SPR's 60 schemes; precedent and full slippery slope are missing.
- **Critical questions.** The SPR gives 138 CQs for 45 of them:
  - 102 map to formulas that already exist, 18 of which need repair;
  - 23 need new formulas;
  - 13 need no formula (conclusion or inference loci).

  The draft loci reproduce the ones WAG already has.
- **One mismatch in WAG.** WAG's `AS:expert_opinion` instantiates `credible_source`, not the collection's `expert_opinion`.

**The plan** (`REPORT.md` §6):

1. Make `walton.yml` the single source.
2. Make it load.
3. Make it instantiate cleanly, gated on zero validator problems.
4. Make its language templates pass WAG's lint.
5. Fix the content errors.
6. Normalize ids, with a map from the old ones.
7. Check against Walton, Reed & Macagno (2008), ch. 9.
8. Add and repair the CQ formulas.
9. Write the NASCQ entries: schemes with answer variables, descriptions, cues and evidence, then ask edges with loci, then backup edges.
10. Gate each batch with `wag-cli --lint`.

**Backup edges are the main design work**, since they encode the semantic matching of §1. They need your review per scheme family.

**One input would help:** a copy of chapter 9 of the 2008 book, or of the official CQ lists you want to count as authoritative. The SPR is AI-written, and its catalogue was reconstructed from this same YAML, so it cannot check the YAML.
