# Answers to part 5 (2026-10-13)

The owner's six points of 2026-10-13, the compendium (book chapter 9) and a third-party audit of the Kialo pilot. Each section answers one point: the short answer, what changed, where it is, and what is still open.

| Point | Short answer | Where |
|---|---|---|
| 1. Argue for or against a premise | Yes, and any other statement too. The challenge stays as a different move. | `nascq/validation.go` (S9), `carneades/statements.go`, the browser's *for · against · challenge* |
| 2. The text protocol | Protocol 1 gave the history but only part of the legal moves and of the help. Protocol 2 gives all of both, with your context, thoughts and summary, plus checkable predictions and simulated alternatives. Every decision is traced. A model played two games through it. | `docs/protocol.md`, `server/protocol.go`, `server/traces.go`, `docs/protocol/` |
| 3. Convert Kialo by hand | Batch 2: 50 claims, two independent annotators, agreement measured, then adjudicated and played as a WAG game. | §3, `docs/kialo-batch2/` |
| 4. `walton.yml` is the authority | Agreed. It is cleaned so all 111 schemes load; a few repairs need your approval. Expert opinion now uses its `expert_opinion`. | `docs/scheme-registry/walton-clean/` |
| 5. Any scheme, but each pattern once; families | Both practical-reasoning variants are now in the game, with negative consequences. A registry checks that no pattern appears under two names, and that close relatives are declared in a family. | `internal/registry`, `data/schemes/families.json`, `docs/scheme-registry/README.md` |
| 6. Weights | Presets stay. Any game can set its own weights, presets can be saved and loaded, and the end of a game shows the verdict under each preset. | Setup → *Scheme weights*; `--contexts` |

## 1. Arguing for or against a premise

**Short answer.** Yes: a premise is a statement, and any statement can now be argued for or against as the claim can. Carneades treats every statement alike. A Kialo discussion is exactly that: pros and cons on any claim, at any depth. The generic question "Is it true that …?" was the wrong tool for that, and it is gone. I kept it as a different move, the **Challenge** ("Why should we accept that …?"), for the following reason.

**Why both.** They do different things to the burden of proof:
- **Challenge.** It costs the challenger nothing. The premise stops counting until its author argues for it. Its presumption is withdrawn (Gordon, Prakken & Walton 2007: asking about an assumption is enough). Hamblin's and Walton & Krabbe's "why?" moves work the same way.
- **Argument against.** The challenger takes on a burden: an argument for ¬P that must outweigh the case for P under the proof standard.

Dropping the challenge would mean a premise could fall only to a counter-argument. That removes the asymmetry that makes presumptive schemes defeasible.

**What was wrong before.**
- **Direct moves on parts of an argument already existed** (an attack with an attack type; a support naming a premise), but they had two problems:
  - The browser did not offer them.
  - Any scheme could be used. The argument's conclusion was forced to the premise whatever the scheme concluded. Practical reasoning could "conclude" that something is possible.

**What changed.**
- **Moves about a statement.**
  - A move names `target_stmt_id`: support argues for the statement, attack argues for its opposite.
  - It is enumerated for bots, offered in the browser and the prompt, traced, and replayable.
- **S9, the scheme must fit.**
  - Other, and the schemes whose conclusion is a statement variable (expert opinion, position to know, sign, analogy), conclude any statement: the variable is bound to it.
  - Any other scheme fits only a statement its conclusion unifies with. The statement fixes the variables, so the premises are about the same action. Practical reasoning or positive consequences can argue for "‘tax meat’ should be done" with A = ‘tax meat’; giving another A is refused.
  - Negative consequences can argue against it.
  - Nothing but a source, an analogy or Other can argue that something is possible.
  - The same rule now holds for an answer to a critical question whose backing scheme has the answer's form.
- **The challenge.**
  - It is answered by any argument for the premise, not only by a defence that names the question.
  - On a premise one of the scheme's own questions is about (possibility), *challenge* asks that question.
- **Two engine bugs the new moves exposed, both fixed and tested:**
  - **Negation.** A scheme concluding ¬p (negative consequences) made ¬p a statement unrelated to p, so both could be IN. A statement and its negation are now always one issue.
  - **Conclusion questions.** A scheme without an answer variable that backs an answer concludes the answer instead of its own conclusion. Its questions about "its conclusion" were still offered, with the answers reversed. They no longer apply to such an argument (`nascq.OwnConclusion`).

**Limit.** A free-text statement (Other, or an imported Kialo claim) can be argued only by Other and the source-based schemes. Practical reasoning cannot conclude "We should tax meat" until the claim is stated in the scheme's form, `should_be_performed(tax_meat)`. The batch-2 converter does that for the thesis. A move to restate a free-text claim as a formula (moderator-approved) would be the general fix.

## 2. The text protocol

**Did protocol 1 give the model what the browser gives? Partly.**

| Item | Protocol 1 | Protocol 2 |
|---|---|---|
| History | The last 30 moves, with effects. | The last 40; questions name their premise; the player's own last move with its effects and how its predictions fared. |
| Legal moves | Critical questions for up to 10 arguments; no moves about statements. | Every critical question of up to 40 arguments; the challenges; every statement with the schemes that can argue for or against it, and the parts it fixes; new argument, pass, resign. |
| Help | Question text; what an attack and a defence would argue; the backing schemes. | Also: which schemes the NASCQ links to the question (🔗 in the browser), the thesis-independent ones, semantic types, cue words, the parts an answer fixes, whether sources are required in this game. |
| Rules | Five lines. | Also: how an issue is decided (strongest argument per side, ties fail, the CCE and BRD margins); what wins a debate and when it ends; what live scores mean. |
| Reply format | Free text, then a move. | Your `context`, `thoughts` and `summary`, plus `expect`, `alternatives` and `confidence`, and the move as `action`. The JSON Schema is in `docs/protocol/reply.schema.json`. |

**Your reply schema.** I adopted it as you described it, and added three checkable fields. They matter because of the next point.

**Are there better ideas? Yes: record what can be checked, not only what the model says.**
- **Stated reasoning is not a trace.** A model's stated reasoning is a rationale, not a record of how it chose (Turpin et al. 2023; Lanham et al. 2023). Fine for reading a game and debugging the prompt; not ground truth for training or judging.
- **Predictions (`expect`)** are checked against the labels after the move, and reported back to the player. They test whether the model understands the game's semantics, separately from whether it argues well. With `confidence`, they measure calibration.
- **Alternatives** are tried on the game and undone, so every decision with alternatives yields preference pairs with their actual effects: training data for DPO-style or value models, at no extra play.
- **Rejections with reasons, outcomes** (winner, final claim label, the verdict under each preset), **human "why" notes**, the **prompt's hash**, **protocol version**, **seed**, **build** and **scheme-data hash** are all recorded. Every decision of every seat is a `wag-trace/1` line.

`docs/protocol.md` lists experiments these support: model against rule bot, model against model, and prompt ablations, measured by rejection rate, prediction accuracy, win rate by side, and human ratings.

**A model played two games through the protocol alone.** You have no LLM to test with, so a model (a subagent with no access to the code) played two games:
- game A: protected bike lanes, as pro against the greedy bot, 6 moves;
- game B: fare-free buses, both sides, 10 moves.

It used every kind of move. **No move was rejected.** Its report, and what I did about each item:

| Finding | Fix |
|---|---|
| Statement numbers shifted when an exception was argued (s9 became s10). | Numbering was visiting exception arguments. Numbers are now permanent; new statements get new numbers. |
| The decision rule was not stated; it learned by playing that 0.65 + 0.60 does not beat 0.70. | Stated in the prompt. The claim shows why it has its label ("for it a26 (0.65), against it a37 (0.70)"). |
| The other side's negated conclusions had no reference. | They read [¬s3]; "¬s3" is accepted. |
| "no →/yes →" confused polarity ("Would it have bad side effects? no → … would have bad side effects"). | Now "attack argues: … / defence argues: …". |
| Predictions were "checked" but never reported. | The next prompt reports them. |
| "sources: required" was not enforced. | It says "required in this game" only when the game requires sources. |
| The prompt grew to 52 KB. | The repeated scheme lists are folded: the thesis-independent schemes are named once. About 42 KB at the same point (47 KB before). |
| A challenge's history line did not name its premise. | It does. |
| Ann's score fell to 0 after +7 and +14. | Debate scores are live: points go when an argument falls. The prompt says so. |

**A semantic bug the replay exposed.** Replaying game B showed a positive-consequences argument that answered "Other Means" being offered its own "Side Effects" question, with reversed answers. That is the second engine bug in §1, fixed.

The audit (N9) also rejects that backing: a good consequence of another action does not show it is the better means. That link is removed from the data. `wag-cli --lint` now warns about every backup scheme that cannot conclude the answer it backs. The two left are the values scheme backing "Other Goals" and "Legitimate Value": your call.

Samples are in `docs/protocol/samples/`: game B's traces without prompts, and the current prompt for its tenth move.

## 3. Converting Kialo by hand: batch 2

*(See below: filled in from the double annotation.)*

## 4. `walton.yml` as the authority

Agreed. The scheme file comes first, the book's compendium second.

**The cleanup.** As uploaded, `walton.yml` does not load: there is a YAML error at line 733. A conservative cleanup (`docs/scheme-registry/walton-clean/`, MPL-2.0 like the original) fixed syntax (12 edits), typos (105), variable declarations (24) and English templates (46), without changing any scheme's content. Every edit is in `CHANGELOG.md` and `changes.json`. Now all 111 schemes load, instantiate and read in English.

**Needs your approval.**
- **Four repairs that change a formula:**
  - `moral_justification_ad_populum_excuse` had no conclusion;
  - `sacrifice`'s undeclared X;
  - `rhetorical_oppositions_normative`'s fourth argument;
  - `sorites_slippery_slope`'s chain premise.
- **Five renamed IDs:** `popular_practise`, `epistemic_ignorence`, `treat`, `exeptional`, `exeption_to_rule`.
- **Four schemes still fail one Carneades validator rule** ("every variable occurs in a premise"): `ad_hominem`, the two practical necessary/sufficient condition schemes, and `regulative_rule`. `walton.proposals.yml` fixes them (14 lines; all 111 then pass). Not applied, because it changes premises.

**Defects in the collection's content**, from the compendium comparison. Not fixed: they are the author's to decide.
- `division_generic` is a copy of `composition_generic`. Division should run from the whole to the parts.
- `from_example` and `common_folk_ad_populum` swap premise and conclusion.
- `negative_value` concludes "retaining" the goal.
- Falsification (32.2) concludes ¬A, the event, instead of ¬H, the hypothesis.
- `inconsistent_commitment`'s exception restates its own premise.
- Seven schemes order arguments differently from their templates.
- 20 zero-argument predicates are one statement shared by every argument of the scheme.

**The game's own schemes.** Five matched `walton.yml` exactly. The other two:
- **Expert opinion** used Wyner's `credible_source`. It now uses `walton.yml`'s `expert_opinion`, the book's scheme 2, whose six critical questions are its three premises, its assumption and two exceptions. The game's six questions now point at those formulas. The Backup Evidence question is about the assumption `based_on_evidence(asserts(W,S))`, so asking it shifts the burden as the book intends. Games recorded earlier still load.
- **Value-based practical reasoning** matches 22.4 only in part. It is Atkinson & Bench-Capon's formalization, as in `walton.yml`.

## 5. Any scheme, each pattern once, and families

**Both practical-reasoning variants.**
- Goal-based practical reasoning (`practical_reasoning2`: Walton 2015, the book's 22.1) is now in the game, beside the value-based scheme.
- Its questions come from 22.1, mapped to `walton.yml`'s formulas:
  - existing: other means, best means, possibility, side effects;
  - new: conflicting goals, intervening actions (audit N13), and goal possibility.
- Negative consequences (33.2) came with it (audit N1–N3), with the likelihood and evidence questions and a new opposite-consequences question. It concludes ¬should_be_performed(A), so a con can now argue directly against "A should be done".

**Uniqueness: the registry** (`internal/registry`, `wag-cli --registry`).
- **Pattern.** It reduces each scheme to its formulas by role, with variables renamed canonically. Two schemes with the same pattern are one scheme under two names, which is an error.
- **Close relatives.** It also finds:
  - schemes that differ in one formula;
  - schemes that extend another;
  - schemes that have the same form with a different predicate (expert opinion, credible source, position to know);
  - schemes that share most formulas (the two practical-reasoning variants).

  A family must declare each such pair.
- **Families.**
  - `data/schemes/families.json`: 21 families and 133 members, drafted from the book's groupings and the collection's category files.
  - Relations: generic, specialization, variant, subtype, counterpart, context-specific, related, alias.
  - This is the generic/specific structure you described.
- **Result on the cleaned collection.**
  - 111 schemes, 110 distinct patterns.
  - Every related pair is declared.
  - The one error is the duplicate division scheme.
- **Result on the game.** 10 schemes, 10 patterns, nothing undeclared.

`docs/scheme-registry/README.md` gives the five steps for adding a scheme.

## 6. Weights

Agreed: an argument's weight is the audience's assessment (Carneades), so it depends on the discussion and on who judges it.
- **The presets stay.**
- **In the setup screen,** *Scheme weights* starts from the chosen preset; any weight can be changed for the game. The game records them (`config.weights`).
- **Presets as files.** *Save as a preset…* writes them as a preset file. *Load a preset…* reads one, and `wag-server --contexts FILE` offers it to everyone. So new presets are made by playing, not by editing code.
- **At the end of a game,** the result panel and the export show how the claim would stand under each preset's weights, without changing the game (`game.ClaimUnder`). A verdict that flips between presets depends on the audience; one that holds under all of them does not.

## 7. A critical review of the audit

The audit is careful, and most of its findings hold. What was done with them:
- **Separate fields.** §7.1's separate fields are the batch-2 guidelines.
- **The policy-pro rule.** §7.2's rule (consequences versus practical reasoning) is the guidelines' policy.
- **Proposals adopted.**
  - N1–N3: negative consequences with its questions.
  - N9: no positive-consequences backing for better means.
  - N13: intervening actions, in goal-based practical reasoning only, as the audit advises, with its yes/no mapping tested.
- **Tests.** §7.8's tests are `internal/game/audit_test.go`: items 1, 2, 5, 6 and 8. Items 3 and 4 need scheme changes you have not decided; item 7 is the importer's.

**Where it is limited.**
- **It could not see the game's code.** So it could not check what it rightly worried about: polarity, and how an answer binds. The tests now do. They found two engine bugs the audit had no way to see (§1): negation without an issue, and conclusion questions on answer arguments.
- **"New" scheme.** It reads the pilot's negative consequences as claimed to be a "new" scheme. The pilot meant new to WAG; it was in `walton.yml`.
- **Data versioning.** It does not raise it, but research on recorded games needs it: changing the data changes which recorded moves are legal. Game B no longer replays past move 9, after N9. Exports now record the scheme data's hash, and loading says when it differs.

**Where I would go further than the audit.**
- **Analogy (N12).** It needs separate source and target propositions. Changing `walton.yml`'s `analogy` is a scheme-level decision; the guidelines keep the source proposition in notes meanwhile.
- **Cause to effect (N8).** It asserts the cause has occurred. A policy debate needs a modal form for what has not happened. That is a new scheme variant, to be declared in the causal family.

## 8. Decisions for you

1. The four `walton.yml` repairs and five ID renames (§4); the four premise proposals; the content defects (§4).
2. Whether the values scheme should keep backing "Other Goals" and "Legitimate Value" (lint warning, §2).
3. N12 (analogy) and N8 (modal cause to effect): a scheme change each.
4. The book's text in the repository.
   - The compendium transcription (chapter 9, verbatim) is not committed. It is given to you as a file; `links.json` keeps only the links.
   - The turn-4 `docs/book-catalogue` quotes the book's pp. 1–138 verbatim. If the repository is or will be public, consider replacing it the same way.

## 9. Checks

- **Tests.**
  - `go test ./...` passes, as does `-race` on the server, NASCQ, registry and adapter packages.
  - New tests: statement moves, S9, the challenge, the audit's §7.8, the registry, protocol 2, traces, custom weights and verdicts.
- **Static checks.** `go vet` and staticcheck 2025.1.1 report nothing.
- **Data checks.**
  - `wag-cli --lint`: 0 errors, 2 warnings (§2).
  - `wag-cli --registry data/schemes/core.yml --families data/schemes/families.json`: 0 findings.
- **Browser.** A Playwright walkthrough found no errors. It covered:
  - saving a preset;
  - arguing against a premise from its card;
  - arguing for the claim with A fixed;
  - challenging a premise;
  - the verdicts under each preset;
  - the traces with a human's "why".
- **The language-model playtest** (§2).
