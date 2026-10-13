# The WAG text protocol: how a language model plays

Any language model, or a person at a chat window, can play a WAG seat:
1. The server writes the game as a prompt in plain English.
2. The player replies in free text that ends with one JSON object: the move.

Nothing else is needed: no tools, no function calling, no particular model.

There are three ways to use it.

| Way | How |
|---|---|
| By hand, any chatbot | In the browser, set a seat to **Language model (copy and paste)**. When it is that seat's turn, the panel shows the prompt with a *Copy* button. Paste the model's whole reply into the box below it and press *Play the reply*. |
| Claude through the API | Start the server with `--claude` (and `ANTHROPIC_API_KEY` set, or an `ant auth login` profile), then set a seat to **Claude (API)**. The server sends the prompt and plays the reply. A rejected move goes back to the model with the reason, up to three times; then the seat passes. Each move costs API credits. `--claude-model` picks the model (default `claude-opus-5-5`) and `--claude-effort` the effort (default `medium`). |
| Your own program | `GET /api/games/{id}/prompt` returns the prompt as text. `POST /api/games/{id}/reply` with `{"reply": "<the model's text>"}` plays the move it contains. The response holds the result, the parsed action and the new state. |

## The prompt

The prompt is self-contained. In order, it gives:

1. **Who you are.** Your side (for or against the claim, or points), the mode, the round and the scores.
2. **How the game works.** Five lines: schemes, critical questions, attack and defence, asking, labels.
3. **The claim**, with its label.
4. **The arguments on the board.** For each:
   - its owner, scheme, weight and whether it holds (and if not, why);
   - which question it answers;
   - its conclusion C and premises, each with its label;
   - any exception that holds;
   - the owner's own words and sources.

   A large game is cut to 24 arguments: those on the claim and the most recent.
5. **What happened.** Every move, with its effects ("The claim is now OUT", "a1 (Alice's) no longer holds"). Only the last 30 moves are shown.
6. **Any question you must answer**, and why your last move was rejected, if it was.
7. **The critical questions you can answer or ask.** They are listed for the arguments you owe answers about and the most recent ones. For each question:
   - its ID and text, and what it is about;
   - what an attack would argue and what a defence would argue;
   - the schemes that can back the answer;
   - the `target_premise_id` to give, when the question is about a premise.
8. **The schemes.** For each:
   - its ID and description;
   - its parts, with type and an example;
   - how it reads, as sentences with `{X}` placeholders;
   - what it assumes and its exceptions;
   - whether sources are required.
9. **The reply format.**

The English is the board's: players' words in ‘single quotes’, nested statements in “double quotes”, an argument's own conclusion as C in its premises.

## The reply

Free text, ending with one JSON object:

```json
{"type": "attack", "target_arg_id": "a13", "answers_cq_id": "CQ:field",
 "target_premise_id": "in_domain(side_effects(ask_her_out,i_am_single),energy_policy)",
 "scheme_id": "AS:position_to_know",
 "variable_bindings": {"W": "a dating coach", "D": "dating"},
 "text": "Whether asking someone out has side effects is not a question of energy policy.",
 "evidence": []}
```

| Field | Meaning |
|---|---|
| `type` | `support`: a new argument (no target) or a defence (a target and a question). `attack`, `question`, `abstain` (pass), `resign`. |
| `scheme_id` | The scheme of the argument (not for a question). |
| `target_arg_id`, `answers_cq_id` | The argument and the critical question answered or asked. |
| `target_premise_id` | The premise the question is about, when the prompt lists one. |
| `variable_bindings` | The scheme's parts as short phrases. Leave out the answer part: the game fills it in. |
| `text` | The argument in your own words (shown on the board). |
| `evidence` | Sources, one string each. |

**What the parser accepts.**
- **Where the move is.** It takes the last JSON object with a `type`, so a reply may think aloud, quote examples, or put the move in a code block.
- **Other spellings.**

  | Instead of | It also accepts |
  |---|---|
  | `scheme_id` | `scheme` |
  | `target_arg_id` | `target` |
  | `answers_cq_id` | `cq`, `question` |
  | `target_premise_id` | `premise` |
  | `variable_bindings` | `bindings`, `slots`, `parts` |
  | `evidence` | `sources` |
  | type `abstain` | `pass` |
  | type `support` | `argue`, `defend` |
  | `AS:position_to_know` | `position_to_know` (missing `AS:` / `CQ:` prefixes are added) |

**Rejections.** A move the game rejects costs the usual penalty, and the seat moves again. The next prompt says why the move was rejected. The Claude seat gets the reason in the same conversation.

## Why text, and not the JSON view

The CLI's `--p1 json` pipe gives programs the full JSON view, including every legal action template. That suits programs. It does not suit language models, for three reasons:
- The templates use Carneades atoms such as `in_domain(side_effects(a,b),d)`.
- There can be hundreds of templates.
- Nothing in them says what a move would mean.

The text protocol gives the same choices grouped by argument and question, each with the sentence an attack or a defence would argue. That is what a player needs to pick a relevant move and fill the parts in sensibly.

## What is recorded

The export of a game (*Export*, `GET /api/games/{id}/export`) records:
- every request that tried to change the game, played or not (`events`);
- for Claude seats, every reply the model gave and why any was rejected (`llm`);
- the testers' notes (`notes`).

Together these let a game played by a model be reviewed afterwards, move by move.
