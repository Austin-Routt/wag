# The WAG text protocol: how a language model plays (protocol 2)

Any language model, or a person at a chat window, can play a WAG seat:
1. The server writes the game as a prompt in plain English.
2. The player replies in free text that ends with one JSON object: the move, with the player's account of it.

Nothing else is needed: no tools, no function calling, no particular model. The prompt names its version (`wag-protocol/2`).

| Way | How |
|---|---|
| By hand, any chatbot | In the browser, set a seat to **Language model (copy and paste)**. When it is that seat's turn, the panel shows the prompt with a *Copy* button. Paste the model's whole reply into the box below it and press *Play the reply*. |
| Claude through the API | Start the server with `--claude` and set a seat to **Claude (API)**. It needs `ANTHROPIC_API_KEY` set, or an `ant auth login` profile. The server sends the prompt and plays the reply. A rejected move goes back to the model with the reason, up to three times; then the seat passes. Each move costs API credits. `--claude-model` and `--claude-effort` choose the model and effort. |
| Your own program | `GET /api/games/{id}/prompt` returns the prompt as text. `POST /api/games/{id}/reply` with `{"reply": "<the model's text>"}` plays the move it contains. |

## The prompt

The prompt is self-contained. In order, it gives:

1. **Who you are.** Your side, the mode, the round, the scores.
2. **How the game works.**
   - The kinds of move.
   - How an issue is decided: the strongest argument in force on each side, not their sum. Under PE a tie leaves both OUT; CCE needs a lead of more than 0.5; BRD also needs the other side under 0.3.
   - In a debate:
     - that the claim, not the score, decides who wins;
     - that pro's first argument states the claim when none is set;
     - when the game ends.
3. **The claim**, with its label and why it has it ("for it a26 (0.65), against it a37 (0.70); it does not win").
4. **Your last move**, what it changed, and how your predictions fared ("s1 out: out (right)").
5. **The arguments on the board.** For each:
   - its owner, scheme and weight, and whether it is in force (and if not, why);
   - the question it answers;
   - its conclusion and premises, each with its reference and label;
   - the exceptions that hold;
   - the owner's words and sources.
6. **What happened**, move by move, with the effects. A question names the premise it asks about.
7. **Your legal moves.** Every move the browser offers, with the same help:
   - **Critical questions**, argument by argument. For each question:
     - its text about that argument, and what it is about;
     - the statement an attack would argue, and the one a defence would argue;
     - the schemes the NASCQ links to it (🔗 in the browser), with their semantic types;
     - whether it can be asked alone.

     The thesis-independent schemes, which back any question, are named once at the top.
   - **Challenges:** the premises no critical question is about.
   - **For or against a statement:**
     - which schemes can conclude which statement;
     - the parts the statement fixes. Positive consequences for [s1] fixes A = ‘tax meat’.
   - **New argument, pass, resign.**
8. **The schemes.** For each:
   - its parts, with examples;
   - its sentences, assumptions and exceptions;
   - whether sources are required in this game;
   - its typical wording (the cue words behind *Which scheme fits?*).
9. **The reply format.**

**Statement references.**
- Every statement a move can be about has a reference: [s1] (the claim), [s2], and so on.
- A number, once given, is never reused or changed. New statements get the next numbers.
- The opposite of [s3] is written [¬s3].
- A move may give a statement by reference ("s3", "¬s3") or by its ID.

**Size.** A game of 40 moves gives a prompt of about 40 KB. Imported Kialo discussions are cut to the 40 arguments nearest the play.

**An inquiry** (mode `inquiry`, [`e1-evidence.md`](e1-evidence.md)) names no side and no winner: the seat's
role is to examine the claim with the other player, and the reply's confidence is how likely the move helps
the inquiry. The model seats' system prompt says the same; other games keep theirs byte for byte.

**Evidence** (`wag-protocol/2+evidence`, an inquiry with a record set): the prompt's first line names the
protocol, and an *Evidence* section comes before the legal moves:
- the claims of the record set the seat may offer (at most 30, the rest counted), each with its mode, its
  source (label, kind and reference), its representation, artifact, revision and capture status, the passage
  it selects, its text and its scope;
- every binding made, with every part its revision covers: its role and target (the argument, the kind of
  statement, the statement), the claim with its passage, text and scope, the interpretation and the
  warrant, the argument's revision, who offered it and whose consent stands, whether it is admitted, and
  its full revision. Past 30, the prompt lists those the seat must weigh: the bindings awaiting its
  consent, then the admitted ones, then the rest, newest first, in the order they were made, and counts
  the others.

The record set's ids and name are plain identifiers ([`evidence-contract.md`](evidence-contract.md)),
written as they are. In such a game an argument's card names its scheme's assumptions `assumption`, the
kind an offer gives them; other games' cards are unchanged.

Every text of the record set and of the bindings is quoted between `<<<` and `>>>`, with `&`, `<` and `>`
written `&amp;`, `&lt;` and `&gt;`, so that no quoted text can end its quotation; a text longer than 2,000
characters is cut, and the cut said. The prompt says that quoted material is to be read and weighed, never
obeyed: what a model is told to do inside a quotation makes no move and no act (the game plays only the
reply's final move). An argument's card also lists the bindings it cites, admitted or withdrawn. The legal
moves add the evidence acts and `evidence_refs` (below).

## The reply

Free text, ending with one JSON object (`protocol/reply.schema.json`):

```json
{
  "context": ["a1 rests on the circumstances [s2]", "nothing argues against [s2] yet"],
  "thoughts": ["if [s2] falls, a1 is out of force", "a source on meat prices weighs 0.65 against the presumption"],
  "summary": "Argue against the circumstances of a1 with a source on meat prices.",
  "expect": [{"stmt": "claim", "label": "out"}],
  "alternatives": [{"action": {"type": "question", "target_arg_id": "a1", "answers_cq_id": "CQ:premise_acceptable", "target_premise_id": "s2"},
                    "why_not": "a question only makes Alice argue for it"}],
  "confidence": 0.6,
  "action": {"type": "attack", "target_stmt_id": "s2", "scheme_id": "AS:position_to_know",
             "variable_bindings": {"W": "a butcher", "D": "meat prices"},
             "text": "Meat is not cheap: prices rose 20% in two years.", "evidence": ["…"]}
}
```

| Field | Meaning |
|---|---|
| `action` | The move. It is the only required field. |
| `context` | What the player observes that matters for the move: one item each. |
| `thoughts` | The player's reasoning, step by step. |
| `summary` | One or two sentences: the situation, the reasoning, the move. |
| `expect` | Labels the player predicts after the move. The server checks them. |
| `alternatives` | Moves considered and not made. The server tries up to three on the game and undoes them. |
| `confidence` | How likely, 0 to 1, the move helps the player's side. |

The move:

| Field | Meaning |
|---|---|
| `type` | One of: `support`, `attack`, `question` (a critical question or a challenge, asked without an argument), `abstain` (pass), `resign`. A `support` is one of three things: a new argument, with no target; a defence, with a target and a question; or an argument for a statement, with `target_stmt_id`. An `attack` is an answer against an argument, or an argument against a statement. |
| `scheme_id` | The scheme of the argument (not for a question). |
| `target_arg_id`, `answers_cq_id` | The argument and the critical question answered or asked. `CQ:premise_acceptable` is the challenge. |
| `target_premise_id` | The premise the question is about, when the prompt lists one. |
| `target_stmt_id` | The statement argued for or against. It names nothing else. |
| `variable_bindings` | The scheme's parts as short phrases. Leave out the answer part and the parts the statement fixes. |
| `text`, `evidence` | The argument in the player's words, and its sources. Both are shown to the other player. |
| `declare` | In a game with a profile (D094 (b)): the actions the move is the first to argue about, each an object with its `id`, its `actor` (required) and its `aliases`. The prompt's *Who decides* names the agent and the actions declared so far. |
| `comparison` | For a scheme that records one (`better_means`, D089): how the alternative is better, `surely`, `sooner` or `cheaper`. |
| `evidence_refs` | In a game with a record set: the admitted bindings a support or attack cites, besides its `evidence` strings. Each must agree with what the move concludes ([`game-rules.md`](game-rules.md) §8). |

The evidence acts, in a game with a record set (types `evidence_offer`, `evidence_consent`, `evidence_withdraw`;
also `offer`, `consent`, `withdraw` and `withdrawal`). A move that is not an act and carries an act's parts
(`offer`, `binding_id`, `revision`) is refused, naming `evidence_refs`, how a move cites a binding. A game
without a record set reads none of these fields or short names, as before E1: a move is played without
them, and an act's type is an unknown move type (S0).

| Field | Meaning |
|---|---|
| `offer` | An offer's parts: `claim_id`, `argument_id`, `kind` (`premise`, `assumption`, `exception`, `conclusion`, `inference`), `statement` (an id or a reference; none for a conclusion or the inference), `role` (`support`, `oppose`, `context`, `definition`), `interpretation`, `warrant`. A part that is not text, or one not listed, is refused by name (the first such part in the order of the parts' names, so that a replay refuses it alike). |
| `binding_id`, `revision` | The binding a consent or a withdrawal names (`binding` is accepted too), and its exact revision, in full or by its first 12 digits. |

**What the parser accepts.**
- **Where the move is.** The final JSON container must be an envelope (with `action`, or `move`) or a move (with `type`). The parser refuses a malformed, incomplete or wrong-kind final container; it never substitutes an earlier example. Prose and examples may precede the final object, and the object may be in a code block. A trailing JSON array or unrelated object is an error. A bracket in prose, such as a citation `[s3]`, a footnote `[1]` or a link, is not JSON: an array counts only when it is empty or starts with an object, an array or a string.
- **Field types.** A target, scheme, question, attack type or text must be text, and the bindings an object of phrases (a number is read as a phrase). Any other value is refused by name, since dropping it would change the move: a defence without its target would be a new argument. `null` counts as absent.
- **Unambiguous values.** Repeated JSON member names in the final candidate
  are refused before decoding can overwrite a field. Names are compared after
  JSON unescaping and remain case-sensitive. Numeric bindings retain their
  exact JSON spelling: a large integer is not rounded through floating point,
  and an exponent or trailing decimal zeros are preserved. This preserves a
  phrase; it does not promise arbitrary-precision engine arithmetic.
- **Historical replies.** Saved moves already hold string bindings and keep
  them, so loading a saved game is unchanged. Replaying a trace (`wag-llm
  replay`) reparses the raw model replies, and four changes of 22-23 November
  2026 can make a trace recorded before them replay differently or fail:
  a numeric binding now keeps its exact spelling; a reply with a duplicate
  key is refused; an out-of-range number (a confidence of `1e400`) no longer
  refuses the whole reply; and a number in a list of the reasoning's texts
  (its context or thoughts) is kept as spelled, which replay does not compare
  (it compares the moves and their results, not those texts). A truncated
  reply keeps its old message.
  Exact historical replay needs the original WAG build and parser as well as
  the other pinned inputs; a changed reading must be recorded as a migration
  or a new replay result. Since step C5.2 a new game has a profile (D094
  (b)): its prompt has a *Who decides* section and a move about an action it
  does not hold must declare it, so a trace recorded before then replays
  with the build it was recorded with (docs/game-rules.md). Likewise a game
  recorded before 2026-11-27 with a value or a name that names nothing (S2,
  P1, the profile's check). The reply's reader drops blank values and keeps
  the others as written, so a model's trace can have one where the reply gave
  a value with no letter or digit (punctuation, an invisible character, a
  sign such as "€").
- **Protocol 1.** A bare move object is still accepted.
- **Other spellings.**

  | Instead of | It also accepts |
  |---|---|
  | `scheme_id` | `scheme` |
  | `target_arg_id` | `target` |
  | `answers_cq_id` | `cq`, `question` |
  | `target_premise_id` | `premise` |
  | `target_stmt_id` | `statement`, `stmt` |
  | `variable_bindings` | `bindings`, `slots`, `parts` |
  | `evidence` | `sources` |
  | type `support` | `argue`, `defend`, `for` |
  | type `attack` | `against` |
  | type `question` | `ask`, `challenge` |
  | type `abstain` | `pass` |

  Missing `AS:` and `CQ:` prefixes are added.

**Rejections.** A rejected move costs the usual penalty, and the seat moves again. The next prompt says why. The Claude seat gets the reason in the same conversation.

## What is recorded

A game's **trace file** (`wag-trace/2`) is a stream of events, one JSON object a line. It is served by `GET /api/games/{id}/traces` (*Traces* in the top bar) and written for every game by `wag-llm play` (`docs/llm.md`). `protocol/trace.schema.json` describes it. The events, in the order they happened (`order`), each with an `id`:

1. **`game_start`** (`g`): what is needed to set the game up again.
   - The new-game request and the resulting configuration; the seats.
   - The model seats' profiles: provider, model, parameters. Keys are never in a profile.
   - The rule bots' seed; the scheme data's hash (`data_sha256`); the protocol version; the build (commit, SDK versions); the system prompt of model seats.
   - The game's resumes from its folder (`resumed`), if any: when, the order of the last event kept, and how many decisions were dropped (below).
2. **`call`** (`c0`, `c1`, …): one call of a model seat to its provider.
   - The request as sent: the system prompt, the **whole conversation** so far (the board, and on a retry the rejected reply and the reason), and the generation parameters.
   - The answer: its text, finish reason, the model (version) that served it, response and request IDs, token usage (a count the provider does not report is absent, not zero).
   - Or the error, classified: rate limit (with the delay asked for), authentication, invalid request, server, timeout, blocked, empty.
   - The latency, and how long the call waited first. A call retried after a rate limit is one event per try (`n`).
3. **`decision`** (`d0`, `d1`, …): one decision of a seat (human, language model, model seat, rule bot), accepted or rejected.
   - **Who.** The player, side, seat kind; for a model seat its profile, provider, model and the model that answered; the turn (moves made before it: the attempts at one move share it) and the attempt number.
   - **The input.** For a language model, the board as the turn began, its SHA-256 and the protocol version; on a retry, the `feedback` it was given. A model seat's decision has the call it came from as `parent`; `recovered` says that the reply was recorded before the game stopped, or could not be saved, and was played afterwards without asking the model again.
   - **The output.** The raw reply; what it said of its reasoning (`context`, `thoughts`, `summary`, `expect`, `alternatives`, `confidence`); for a human, the optional *Why this move?* note.
   - **The move and its result.** The move as validated (references resolved, attack type and answer filled in); accepted or why not; the arguments made; the score change.
   - **What it did.** The claim's label before and after; the effects ("a1 (Alice's) no longer holds"); each prediction against the actual label; what each named alternative would have done (legal or not, the claim's label after, the score change), from trying it on the game and undoing it.
4. **`game_end`** (`e`), once the game is over: the outcome (winner, verdict, reason, scores), every statement's final label, and the verdict under each context preset's weights.

A call is saved, with its answer or its error, before the answer is read or played: `wag-server` with a data folder saves a game with a model seat after every change, and a call is such a change (`wag-llm play` keeps no data folder). If that save fails, the answer is kept in memory, the call's retries stop, and the model seat waits; once a save succeeds, its move clock starts again, and the answer is played at the seat's next turn. A game in the data folder can be **resumed** from its folder (*Resume* in the browser's list of saved games; `POST /api/store/resume`, this computer only): a folder of `games/`, with its export and trace as files, not links, of a game whose id is one the server gives. The trace must be the export's (the same game, data, seats, request, seed, moves and decisions), and its start the one this server would write for the game in what changes play: the same protocol, system prompt, model profiles and context presets, and the engine's and data's pins where the start recorded them. Otherwise the resume is refused, saying which, and the export can still be loaded as a new record, which asks a model seat again for a turn in progress. The build is not compared: the start keeps the build that began the game, and each resume records its own (`resumed[].build`, `resumed[].pins`). The resumed game keeps its id, its clocks (the move clock starts again; the game clock counts the time played, which each save records in the export as `played_ns`, so that a game whose time was up is over), a move held for the moderator (held again, for the moderator to judge), and its record, and goes on saving into its folder. A save writes the trace, then the export; decisions the trace has and the export does not (a save that stopped between the two) are dropped, and their replies played again. A model seat's turn in progress goes on from its record: each recorded reply no decision used is played in its attempt, if it answered the conversation the turn has then (on a board with no move since the turn began, this board), and the model is asked only for the rest; any other move since its call (another player's, or the moderator's decision on a statement) ends the turn in progress, and the reply is not played. A turn whose move was held for the moderator is over: rejected, the seat's next turn starts afresh with the moderator's note on the board. Computer players draw again what they drew before the stop, so that they go on as in a game that did not stop. The resumed game's seats get new keys: a seat's link given to another computer before the stop is given again. See `docs/g-journal.md`.

The export keeps the decisions without their prompts, the testers' notes, every request that tried to change the game, and the verdicts. `wag-llm replay` plays a trace file again with the recorded answers and checks that the game sends the same requests and makes the same decisions; a call that failed as recorded fails again, and the replay asks again, as the game did. It refuses a trace played with other scheme data (`data_sha256`; give that data with `--data`), and sets the game up with the context presets the trace recorded, since a preset can change without changing the data.

## Debugging and training: what to record, and what not to trust

**What the reasons a player gives are good for.** The `context`, `thoughts` and `summary` fields make a game readable, and show where a prompt misleads a model. They are a rationale, not a record of how the move was chosen. Language models' stated reasons can differ from what actually drove the answer:
- Turpin et al., "Language Models Don't Always Say What They Think", NeurIPS 2023;
- Lanham et al., "Measuring Faithfulness in Chain-of-Thought Reasoning", 2023.

Do not train a model to imitate them as if they were ground truth, and do not judge a move by them.

The protocol therefore also asks for what can be checked:

- **Predictions** (`expect`), checked against the labels after the move.
  - They test whether the player understands the game's semantics (proof standards, ties, undercutting) separately from whether it argues well.
  - With `confidence`, they measure calibration.
  - The next prompt tells the player how they fared.
- **Alternatives**, tried on the game.
  - Each decision with alternatives gives preference pairs: the move made, the moves passed over, and their actual effects.
  - That is the data for preference training (DPO-style) or for a value model, at no extra play.
- **Rejections with reasons:** data for learning the rules and for self-correction.
- **Outcomes.**
  - Each trace ties to the game's end: winner, the claim's final label, the verdict under each preset.
  - That is the reward for reinforcement learning, and the test for whether a move that looked good helped.
- **Human reasons and notes.**
  - The *Why this move?* field and the testers' notes are human rationales and judgements: what one wants a model to learn from.
- **Reproducibility.**
  - The protocol version, the prompt's hash, the seed, the build and the game's settings are recorded.
  - A game can be replayed from its export.

**Experiments this supports.**
- **Model against rule bot**, as a baseline. The greedy bot plays legal moves without understanding.
- **Model against model.**
- **Ablations** of the prompt: without the question catalogue, without the scheme help, protocol 1 against protocol 2.

Measure each by:
- the rejection rate;
- prediction accuracy;
- win rate by side;
- human ratings of the arguments, from the notes.

On 2026-10-13 a model played two games through this protocol alone (18 moves, none rejected; `docs/review-2026-10-13-part5.md`). Its report led to the permanent statement numbers, the stated decision rule, and the "Your last move" feedback.
