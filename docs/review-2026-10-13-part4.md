# WAG review, part 4: your test game, readability, LLM players, weights, contexts, and the Kialo plan

Date: 2026-10-13. This answers the fourth message. It covers the game you exported (`wag-game-4.json`), the book excerpt (Walton, Reed & Macagno 2008, pp. 1–138), and your questions.

| Question | Short answer | § |
|---|---|---|
| Why was the game hard to read and follow? | Three causes. The English templates assumed particular phrasings. Nested statements were repeated in full. The page put the board first and the history last. All three are reworked: the dialogue is now the main view. | 1, 2 |
| Why weren't `{S}` placeholders substituted? | The preview filled the scheme's templates before the game had bound the answer part. The preview now shows the argument exactly as the move builds it. | 1 |
| Why does the bot always cite "UN panel … energy policy"? | Its word list was a fixed climate-policy vocabulary from the original code, and ties always went to the first move it tried. Both are fixed. The bots are rule testers, though: use an LLM seat for real play. | 1 |
| Why these scheme weights? | They came with the original plan (Plans.md D6) and have no source. Walton gives schemes no weights. In Carneades a weight is the audience's assessment. | 3 |
| Context presets, like "Scientific"? | Yes. Five draft presets now set the proof standard, sources and an audience weighting. | 4 |
| Why couldn't you attack practical reasoning's premises? | Its five critical questions mostly point at exceptions and the conclusion. A generic premise question now covers every other premise. Separately, WAG's practical reasoning is not the book's scheme. | 5 |
| Is the game just drilling down through critical questions? | Depth is how a dispute gets resolved, but the game is also breadth: new pros and cons on the claim, where to fight, asking versus arguing, defending. The bot made it look like a single chain. | 6 |
| Can any LLM play from a text protocol? | Yes, now: `docs/protocol.md`, a copy-and-paste LLM seat for any chatbot, and a Claude seat through the API. | 7 |
| A better way to experiment and validate? | Four layers: rules, language, play quality, human sessions. Your exports become regression tests. | 8 |
| What data would help? | The new export, which records the build, every attempt, notes and LLM replies, plus one sentence on what you meant whenever something went wrong. | 9 |
| Convert Kialo by hand, adding schemes as we go? | Yes. Approve patterns rather than instances, with a running confidence and a promotion threshold. A 20-claim pilot shows what it takes. | 10 |

## 1. Your test game, move by move

`wag-game-4.json` now replays exactly: same labels, same scores (`internal/server/testdata/user-game-4.json`, `TestLoadReplaysAnExport`). Here is what went wrong in it, why, and what changed.

**#1. Alice argues.**
- **What went wrong.** The board showed "Performing ask a girl on a date in alone would bring about in a relationship" and "Sex would be realized in in a relationship".
- **Why.** The templates were built for noun phrases ("carbon tax") and placed your words without marking them.
- **Now.** Your words stand in single quotes, so any phrase reads: "Doing ‘ask a girl on a date’ in the situation ‘alone’ would lead to ‘in a relationship’." Every input also shows an example of what to type ("e.g. introduce a carbon tax", "e.g. emissions are rising").

**#2. Bob attacks with Side Effects, citing "UN panel … energy policy".**
- **Why the same move every time.** The greedy bot tries moves and keeps the best. Ties went to the first move it tried, so it always chose the first question and the first backing scheme.
- **Why that vocabulary.** Its words came from a fixed climate-policy list (`agent.go`: "un_panel", "energy_policy" …).
- **Now.** Ties are broken at random. The bots reuse the words already on the board and cite neutral sources ("an expert", "the subject at hand").
- **Still true.** Rule bots have no grasp of language. Their job is to exercise the rules; for meaningful opponents, see §7.

**#3. You attack the Field question.**
- **Your move was right.** This is exactly what the Field question is for. It was hard to read for two reasons:
  - The question nested the whole claim: "Is the claim performing … within the field energy policy?"
  - Your answer was rendered as "It is not the case that “…” is in the domain of energy policy".
- **Now.** The question reads "Is “Doing ‘ask a girl on a date’ in the situation ‘alone’ would have bad side effects” really a matter of ‘energy policy’?". The answer has its own negative form: "… is not a matter of ‘energy policy’". Every predicate now has a negative sentence.

**#3, continued. The `{S}` placeholders.**
- **What went wrong.** The preview listed "a man asserts that {S}" and "{S} is in the domain of asking girls out".
- **Why.** The preview filled the scheme's English templates before the game had bound the answer part S.
- **Now.** The preview simulates the move and shows the argument it builds, every part bound. It names the argument's own conclusion **C**, so the premises read "‘A man’ says that C" and "C is a matter of ‘asking girls out’" instead of repeating a sentence that already contains a sentence.

**#4. Bob undermines your witness with "World health org … economics".**
- **Why.** The same word list.
- **The scores.** The swings (+13.5, −13.5) were correct live scoring, but nothing explained them.
- **Now.** Each move in the history lists its effects: "The claim is now OUT (it was in)." "a1 (Alice's) no longer holds." "a13 (Bob's) holds again."

## 2. Readability: what changed

**The dialogue is the main view.** You found the history most helpful, so the history is now the page:
- Every move is a card. It shows:
  - who moved and what kind of move it was;
  - which question about whose argument it answers, with the question (**Q**) and the answer (**A**) as sentences;
  - the mover's own words;
  - **because …**, which unfolds the scheme's premises, each with its label;
  - whether the argument holds;
  - what the move changed, and the score change.
- A reply is indented under the move it answers.
- **Respond…** on any card lists that argument's critical questions, each with *Attack: answer no*, *Defend: answer yes* or *Just ask*.
- **⚑ Note** records a remark on that move (see §9).
- The claim-centred board is still there under **Board**.

**Plain words in the controls.**
- The move kinds: New argument, Attack, Defend, Ask, Pass, Resign. Each has a one-line explanation.
- What a question is about: "a premise: its author must defend it", "an exception: whoever raises it must argue it".
- Steps such as "Say it in your own words", "Pick the scheme it fits", "Fill in its parts: short phrases, as in the examples".

**The rules of the English** (`GameState.Say`; How it works tab):

| Mark | Means |
|---|---|
| ‘single quotes’ | a player's words |
| “double quotes” | a statement inside a statement |
| [brackets] | a statement nested deeper, shortened |
| C | on an argument's card, its own conclusion |

## 3. Why these weights?

They were in the initial commit (`9607261`, `data/nascq/schemes.json`), from the original plan (Plans.md, D6: "In Trusted mode, argument weight equals the scheme's base_weight"):

| Scheme | Weight |
|---|---|
| practical reasoning | 0.70 |
| expert opinion | 0.70 |
| consequences | 0.65 |
| position to know | 0.65 |
| values | 0.60 |
| sign | 0.60 |
| analogy | 0.55 |

Nothing in Walton, the CAF or the NASCQ paper supports these numbers. Walton's schemes are presumptive: an argument shifts a burden until a critical question is asked (book p. 96). They carry no numbers. Plans.md itself, in its review notes, calls constant weights "a design smell".

**What weights do in WAG.** They matter in two places only:
1. **When pro and con arguments meet on the same issue.** Under PE the heavier side wins; CCE and BRD add margins.
2. **Points:** 10 × weight.

They do not affect undercutting, undermining or questions.

**Where weights should come from.** Carneades treats a weight as the audience's assessment of an argument. So the defensible sources are:
- the **moderator**, who can already set a weight when approving a move;
- **votes**, as Kialo's impact votes are for imported discussions;
- a **context**: what an audience in that kind of discussion would grant (§4).

The "equal" preset (all 0.5) lets the structure alone decide, which is the cleanest setting for testing the rules.

## 4. Context presets

Yes, there should be context presets. Walton and Krabbe distinguish dialogue types (persuasion, inquiry, deliberation, …), and the CAF ties the proof standard to the kind of dispute. A preset now sets:
- the claim's proof standard;
- whether sources are required;
- whether a moderator is expected;
- the weights an audience in that context would give each scheme (`data/contexts.json`, chosen under **Context** on the setup page).

| Preset | Standard | Sources | Weights (drafts) |
|---|---|---|---|
| Everyday discussion | PE | optional | the schemes' defaults |
| Policy deliberation | PE | optional | goals, values and consequences highest |
| Scientific inquiry | CCE | required | expert opinion 0.75 and sign 0.65 highest; values 0.3 lowest |
| Legal case (criminal) | BRD | required, plus a moderator | witnesses, experts and precedent highest |
| Equal weights | PE | optional | all 0.5 |

**The weights are my drafts and need your judgment.** Two further things a context could set:
- the admissible schemes, for example no appeal to popularity in an inquiry;
- the admissible critical questions.

Both wait for the NASCQ expansion.

## 5. Attacking premises

**Why you couldn't.** WAG only offered moves through critical questions. Practical reasoning has five (other means, best means, other goals, possibility, side effects), the same five as the book (p. 96). Only possibility is about one of its premises. The others point at exceptions and at the conclusion. So the circumstances, the means-end link, the goal and the value could not be challenged.

**Why that is wrong.** The book says a defeasible argument "can be attacked by an attack on a premise, by a counterargument with an opposite conclusion, or by an argument attacking the inference rule" (p. 32). It also treats the critical questions as adding to the premises, not replacing a challenge to them (p. 17).

**Now.**
- Every premise that no question of its scheme is about gets the generic **Premise Acceptability** question: "Is it true that “…”?". You can ask it, or attack or defend the premise with a backing scheme.
- It is the option *Any premise can be questioned*, on by default. The NASCQ's own questions stay separate from it.
- On a practical-reasoning argument, that adds six challengeable premises; possibility keeps its own question.

**A larger issue the book raises.** WAG's practical reasoning is not Walton's scheme:
- WAG's is Atkinson and Bench-Capon's value-based scheme: circumstances S1, action A, resulting state S2, goal G, value V.
- The book's is the goal/means scheme (p. 95), with its necessary- and sufficient-condition forms (pp. 95–96).
- They share the five questions. But WAG's value premises, and its assumptions "V is a legitimate value" and "G is a worthy goal", have no counterpart in the book.
- The book also says asking any of the five questions defeats the argument until the proponent answers (p. 96). That puts the burden on the proponent. WAG, like Carneades, makes three of them exceptions, which the questioner must argue.

This is a decision for you, before the NASCQ is expanded: keep the value-based scheme, switch to the book's, or offer both. The catalogue (§11) lists every difference.

## 6. Is the game just drilling through critical questions?

Depth is how a dispute gets resolved. An argument stands until a question is raised against it. The answer is itself an argument with questions of its own, so the dispute goes down until one side runs out of answers. That is Walton's dialectic, and the game should keep it.

But it is not the whole game. A player also chooses:
- **Breadth.** New arguments for or against the claim, each weighed under the proof standard against the others. Kialo's pros and cons are mostly this.
- **Where to fight.** Any open question on any argument, not just the last move.
- **Ask or argue.** A question costs nothing and shifts the burden. An argument commits you to premises the other side can attack.
- **Defend rather than counter-attack.** A positive answer protects a premise from below.
- **When to stop.** In a debate, passing while the verdict is against you concedes.

Your game looked like a single chain for three reasons:
- The claim had one argument.
- The greedy bot always answers the latest argument with the move that flips the claim right now.
- Under live scoring, the last word pays.

**What would make play richer.**
- Opponents that argue about the topic (LLM seats, §7).
- Debates opened with several pros and cons, as imported Kialo discussions already are.
- A cost per move or a move budget, so that piling on is not free. This was listed as your decision in part 2.

## 7. A text protocol any LLM can play

Before this round, only the CLI's JSON pipe existed. Its legal moves are lists of Carneades atoms, with nothing to say what a move means: usable by a program, not by a language model.

Now there is the text protocol, `docs/protocol.md`:
- **The prompt.** The server writes the game as a self-contained prompt in plain English:
  - the rules in five lines;
  - the claim;
  - the arguments, with premises, labels and the players' words;
  - the history, with each move's effects;
  - the critical questions the player can answer or ask, each with what an attack and a defence would argue and the schemes allowed;
  - the schemes' parts, with examples;
  - the reply format.
- **The reply.** Free text ending in one JSON move. The parser is lenient: it takes the last move in the reply, and accepts common variations of the field names.

Three ways to use it:

| Way | What you do |
|---|---|
| **Language model (copy and paste)** | Set a seat to it. The panel shows the prompt with a Copy button; paste any chatbot's reply back. |
| **Claude (API)** | Start the server with `--claude` and an API key. The seat plays by itself. A rejected move goes back to the model with the reason, up to three times. Each move costs API credits. Defaults: `claude-opus-5-5`, effort `medium`; both can be set by flag. |
| **Your own program** | `GET /api/games/{id}/prompt`, then `POST /api/games/{id}/reply`. |

Every reply Claude gives, rejected ones included, goes into the export.

I have not run a Claude seat against the real API from here: that needs your key and spends your credits. It is tested with a stand-in model (`TestClaudeSeat`).

## 8. A better way to experiment and validate

Validate in four layers, each with its own instrument:

| Layer | Question | Instrument |
|---|---|---|
| Rules | Is every move legal and every label right? | The test suite: invariants, mode tests, differential labelling. **Every game you export can become a regression test**, as yours now is. |
| Language | Does every statement read well, whatever a player types? | `TestEnglish` and the scheme cards. Proposed: a gallery of every scheme's sentences, filled with awkward inputs, to review in one sitting. |
| Play quality | Do the moves make sense in a real debate? | LLM self-play on a fixed set of claims (the five Kialo theses are a ready set). Claude against Claude, or against you. Transcripts are reviewed with notes. |
| Experience | Can a person play it, and understand why things happen? | Short sessions with you (below). |

**A session that gives me what I need** (about 15 minutes):
1. Pick a claim you care about.
2. Play 6–10 moves against an LLM seat (any chatbot), or against Claude.
3. Press **⚑ Note** on anything wrong, confusing, unreadable or good. One line each: what you expected, and what happened.
4. **Export** and send me the file.

With the export I can replay the game exactly, see what you saw, and turn each note into a fix and a test.

## 9. What data helps

**The export now holds** (format `wag-export/2`):
- the build (commit, and whether the source was modified);
- the request that started the game, seed included, so it replays exactly;
- every move;
- every move the server received, played or rejected, with the reason;
- your notes;
- for Claude seats, every model reply;
- the state as rendered, with the English you saw.

**Beyond the export:**
- **Your intention.** For anything that surprised you, one sentence on what you meant to do. A note does this.
- **A screenshot**, when the problem is the layout.
- **For LLM seats,** which model played. Its reply is in the export when the server played it; paste it into a note when you copied it by hand.

**Exports from before this round.** They still load. Your game-4 export did; it is a test now.

## 10. Converting Kialo by hand

**Should we?** Yes. Walking real discussions is the best way to decide which schemes, questions and links the NASCQ needs first. It shows what arguments actually look like, and it builds the data we need for an argument-mining classifier later.

**The pilot.** I ran a pilot on 20 claims of "Should governments put a tax on meat?": the thesis, its 6 children, and 13 grandchildren. It is in `docs/kialo-pilot/pilot.md`. Each claim has:
- a reading, with its implicit premises;
- a scheme, with its source and a confidence;
- its role towards its parent;
- the parts filled in, as WAG's English;
- the NASCQ additions it needs;
- its pattern;
- an approve / reject box.

What it found:
- **Scheme fit, 19 arguments.**

  | Fit | Arguments |
  |---|---|
  | A WAG scheme | 6 |
  | A walton.yml or book scheme | 5, mostly negative consequences (p. 101) |
  | A new scheme | 1 |
  | Only "Other" | 7 (bare denials, and unargued facts or evaluations) |
- **Roles.**

  | Role | Arguments |
  |---|---|
  | Answer a critical question | 5 |
  | Back or undermine a premise directly | 4 |
  | Plain pro or rebuttal | 9 |
  | Fits no role | 1 |

  5 of the 19 act on something other than their Kialo parent: Kialo hangs a claim under the claim it replies to, not under the part it attacks.
- **Most needed.** An argument from negative consequences, with its book questions and backup edges (proposals N1–N4), would cover 8 of the 19. An "efficacy" challenge (N5, would the action work?) is now possible through the generic premise question.
- **Confidence** (self-estimated): scheme 0.51, role 0.56. 15 patterns for 19 arguments, 10 with a single instance, so the inventory has not settled.

**How to approve.** Approve patterns, not instances, with a running confidence:
- **A pattern** is: claim type + relation → scheme, question, role, and the NASCQ entry it needs. An example: "con of a policy ← negative consequences, as a rebuttal".
- **Each pattern carries a running score:** instances that fit, instances you rejected, and the mean confidence.
- **Promote a pattern into the shipped NASCQ** when it has at least 3 instances you approved, from at least 2 discussions, with none rejected and unresolved.
  - Before that, it lives in a *provisional* data file the game can load with `--data`, so it can be played.
- **You review in batches** of 20–30 claims, deciding patterns. I annotate instances and flag the ones that do not fit.
- **Agreement.** Before scaling, a second annotator should redo about 30 claims, to measure agreement on scheme and on role. That could be you, or a second model run blind.

So to your question: you need not wait for three or four confirmations before saying anything. Approve or reject the pattern as soon as you see it. The threshold only decides when an approved pattern ships.

**Decisions the pilot needs from you** (pilot.md, "Lessons"):
1. Pros of a policy: practical reasoning, or consequences when only a consequence is stated?
2. Cons of a policy as rebuttals by negative consequences?
3. Bare denials as question moves?
4. May a claim be re-attached to the argument it really attacks?
5. Approve N1–N5?

## 11. The book excerpt

`docs/book-catalogue/` catalogues the excerpt, read from the page images. The OCR text was used only to locate passages and to cross-check.
- **Coverage.** The excerpt runs to p. 138 (chapters 1–3 and the start of chapter 4).
- **Schemes.** 75 stated schemes in 22 families, each with:
  - the page and PDF location;
  - the premises and conclusion;
  - the critical questions, verbatim;
  - a drafted locus for each question, with the book's own words where it says who bears the burden (pp. 17, 32–33, 63, 92, 96, 124, 137);
  - the matching WAG and walton.yml schemes.
- **Not covered.** Argument from values, the causal schemes (ch. 5), the ethotic, ad hominem and bias schemes (p. 140 on), and the chapter 9 compendium. Please send:
  - pp. 139–200, which include chapter 5;
  - chapter 9 (about pp. 308–346), the compendium the NASCQ expansion should follow.

**WAG against the book** (catalogue §4):
1. **Practical reasoning** is a different scheme (§5).
2. **Expert opinion** uses Wyner's `credible_source`, not walton.yml's `expert_opinion`.
   - The book's Field question asks whether *E* is an expert in the claim's field, and its Opinion question asks "What did E assert that implies A?". WAG asks something else in both.
   - Trustworthiness has lost "personally reliable".
   - The `dishonest(W)` exception has no question.
   - walton.yml's version matches the book's burdens (p. 33) one to one.
3. **Analogy** lacks the book's CQ2 (are the cases similar in the respects cited?). WAG treats CQ4 as an exception where the book calls it a defeater.
4. **Position to know** has an extra Field question, and its trustworthiness question adds bias.
5. **Positive consequences** has two premises where the book has one, and reuses side effects for the book's "opposite consequences".
6. **Sign and values** cannot be checked against the excerpt.

These are the first items for the NASCQ expansion. All of them change what the game means, not how it reads, so I have left the data as it is for you to decide.

## 12. What changed in this round

- **English.** Quoted player words, negative templates, shortened nesting, C on cards, and shorter questions.
- **Preview.** It shows the argument as the move builds it, in the browser and the CLI. This fixes `{S}`.
- **Premise question.** Generic, for any premise no critical question covers (option, on by default).
- **Rule bots.** They reuse board words, cite neutral sources, and break ties at random.
- **History.** Each move lists its effects.
- **The page.** The dialogue view, Respond on every move, notes, plainer controls, and the scheme cards in the board's English.
- **Text protocol.** The LLM seat (copy and paste) and the Claude seat (API).
- **Saving and replaying.** Load an export, and a richer export: build, request, attempts, notes, LLM replies.
- **Contexts.** Five draft presets.
- **Documents.** The book catalogue and the Kialo pilot.
- **Build.** WAG now needs Go 1.24 to build from source, because the Anthropic SDK requires it. The ready-made `wag-server.exe` needs nothing.
