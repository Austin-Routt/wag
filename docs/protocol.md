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

**What the parser accepts.**
- **Where the move is.** It takes the last JSON object that is an envelope (with `action`, or `move`) or a move (with `type`). A reply may think aloud, quote examples, or use a code block.
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
2. **`call`** (`c0`, `c1`, …): one call of a model seat to its provider.
   - The request as sent: the system prompt, the **whole conversation** so far (the board, and on a retry the rejected reply and the reason), and the generation parameters.
   - The answer: its text, finish reason, the model (version) that served it, response and request IDs, token usage (a count the provider does not report is absent, not zero).
   - Or the error, classified: rate limit (with the delay asked for), authentication, invalid request, server, timeout, blocked, empty.
   - The latency, and how long the call waited first. A call retried after a rate limit is one event per try (`n`).
3. **`decision`** (`d0`, `d1`, …): one decision of a seat (human, language model, model seat, rule bot), accepted or rejected.
   - **Who.** The player, side, seat kind; for a model seat its profile, provider, model and the model that answered; the turn (moves made before it: the attempts at one move share it) and the attempt number.
   - **The input.** For a language model, the board as the turn began, its SHA-256 and the protocol version; on a retry, the `feedback` it was given. A model seat's decision has the call it came from as `parent`.
   - **The output.** The raw reply; what it said of its reasoning (`context`, `thoughts`, `summary`, `expect`, `alternatives`, `confidence`); for a human, the optional *Why this move?* note.
   - **The move and its result.** The move as validated (references resolved, attack type and answer filled in); accepted or why not; the arguments made; the score change.
   - **What it did.** The claim's label before and after; the effects ("a1 (Alice's) no longer holds"); each prediction against the actual label; what each named alternative would have done (legal or not, the claim's label after, the score change), from trying it on the game and undoing it.
4. **`game_end`** (`e`), once the game is over: the outcome (winner, verdict, reason, scores), every statement's final label, and the verdict under each context preset's weights.

The export keeps the decisions without their prompts, the testers' notes, every request that tried to change the game, and the verdicts. `wag-llm replay` plays a trace file again with the recorded answers and checks that the game sends the same requests and makes the same decisions.

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
