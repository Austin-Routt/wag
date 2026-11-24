# Language models in WAG: the dashboard, providers, traces, annotation

WAG's seats can be played by a language model through its API: Gemini, Claude, OpenAI, DeepSeek, or a model served locally (LM Studio, llama.cpp, SGLang, vLLM, Ollama). Another harness, such as Guidance or DSPy, plugs in as an external command.

Every seat gets the same prompt (`docs/protocol.md`), and the game's own parser and validator judge every reply. Every call and every decision is traced: the whole conversation sent, the answer, usage, latency, the move and what it did. A game can be replayed from its trace.

The code is in `wag/internal/llm`, an interface with one adapter per provider. The game does not import any provider's types.

## The dashboard: the easy way, on Windows too

Everything below can be done in the browser, with no command line and no environment variables.

1. **Start WAG.**
   - On Windows, double-click `wag/run-wag-server.bat` (it needs Go), or `wag-server.exe`.
   - Anywhere else: `go run ./cmd/wag-server` in `wag/`.

   The browser opens at http://127.0.0.1:8080.
2. **Connect a model, in the Models tab.**
   1. Choose the service (Google Gemini, Anthropic Claude, OpenAI, DeepSeek, a local server, …).
   2. Paste your API key. Tick *Remember on this computer* if it should survive a restart.
   3. Click *Find models*, search, and pick one.
   4. *Test it*, then *Add player*.
3. **Play.** In *Play → New game*, every player is a seat: against you, a rule bot, or another model.
4. **Annotate.** In the *Annotate* tab:
   1. Choose a dataset: the built-in Kialo batch (50 claims, with a gold annotation), a captured Kialo discussion (*New dataset from a Kialo discussion*), or your own folder.
   2. Tick one or more annotators, and optionally a reviewer.
   3. Create the run. Then either:
      - *Annotate the next claim*, to look at each claim as it comes;
      - or *Annotate all the rest*, which runs in the server, so you can close the page.
   4. *Review the disagreements* has the reviewer compare the annotations, field by field. With one annotator there is nothing to compare: the button is *Have the reviewer check each annotation*, and the reviewer judges each field of the one annotation, giving its own where it disagrees. Choose a reviewer other than the annotators if you can: a model checking its own work tends to agree with itself.
   - **Calls and limits.** Each claim costs one call per annotator, plus one if it is reviewed. Each call sends about 41,000 tokens (the guidelines and the scheme collection), and Gemini 2.5 Flash also spends about 850–7,600 thinking tokens on it (in our first run). On the free tier (5 requests a minute, 20 a day, in October 2026), 50 claims with one annotator take three days. When the provider says a quota is used up for longer than two minutes, the run stops that annotator at once and says how long to wait; *Annotate all the rest* carries on later. An annotator whose replies cannot be used for three claims in a row (each reply is sent back once, with the reason), or a reviewer whose replies cannot be read for three claims in a row, is stopped too, and says why: more calls would most likely fail the same way. Look at the replies in `calls.jsonl`, change the model, its settings or the prompt, then start again.
   5. Open a claim: the annotations side by side, the fields they differ on, the problems the check finds. Give your verdict. The reviewer and the gold stay hidden until you choose to see them.
   6. The agreement table compares the annotators with each other, with the gold, and with the reviewer.

**Keys.**
- A key typed in the dashboard goes to the WAG server on your computer and stays there: in memory until it stops, or in `keys.json` in your data folder if you tick *Remember*. That file is not encrypted; *Forget it* deletes it.
- The dashboard never shows a key again (only its last four characters). No game, trace, run or player file contains it.
- The model and annotation pages answer only requests from this computer.

**Where everything goes: your data folder.** It is `WAG` in your home folder (for example `C:\Users\you\WAG`), unless you start the server with `--store DIR`. The dashboard shows it, with an *Open folder* button.

| In the data folder | What |
|---|---|
| `players.json` | The players you added (service, model, settings; no keys). |
| `keys.json` | The keys you asked to remember (only then). |
| `games/DATE_gameN/` | Every game with a model seat, saved after each move: `trace.jsonl` (the trace file, below) and `export.json` (loads in the browser), each replaced whole and synced to the disk. The game bar's *Saved ✓* opens it. A model's replies are saved even when its turn then fails, and automatic play pauses after a failed turn instead of asking the model again. If the game cannot be saved, the game bar says so instead, with why, and *Save now*; meanwhile its model seats are asked nothing, since their replies cannot be had again for nothing. Every move and *Save now* try again, and a folder that could not be synced is synced again before the game counts as saved. |
| `annotation/RUN/` | One folder per annotation run; its `README.txt` says what each file is. `run.json` holds the settings, the models and the inputs' hashes. `annotations/NAME.json` holds each annotator's annotations, in the guidelines' format, which `validate.py` and `agreement.py` read. `calls.jsonl` holds every call. `review/model-NAME.json` holds the reviewer's decisions, `review/human.json` your verdicts, and `report.json` validation and agreement. A folder made by `wag-llm annotate` also has `budget.jsonl`, a line for each invocation's budget ([below](#a-runs-budget-d046)). |
| `datasets/NAME/` | Your datasets: `items.json` (claims in the batch-2 format), and optionally `guidelines.md`, `gold.json`, `lexicon.json`. |

The command line below (`wag-llm`) does the same, writing to the folders you give it.

### Continuing and reassessing an annotation run

A run preserves the input hashes and annotator profiles with which it began. Both the dashboard and the command line refuse to continue it under a changed model, provider, endpoint, system prompt, output schema or generation parameters. Schema contents are saved, so replacing a file at the same path cannot change an existing run. Timeout, spacing, retry and key settings can change. A legacy schema run that did not save its schema bytes stays readable but needs a new run for further calls.

Gold annotations used only for scoring may be corrected or removed without changing the annotator's identity. A file also supplied as claims or resources remains an annotation input even if the dataset calls it gold. A drifted run can still be inspected and reassessed when its dataset remains loadable. Its report identifies the current evaluation input hashes and build; a later human verdict records that context too. Historical `run.json` input hashes are not replaced with the new ones.

A dataset is read strictly, so that nothing ambiguous decides a result in silence: a JSON file with a duplicate member, a second spelling of a field (`Gold` beside `gold`), invalid UTF-8 or nesting deeper than 10000 levels is refused; so are a claim or annotation with no ID, or an ID of spaces, and two with the same ID; and a file `dataset.json` names, or a default file that exists but cannot be read, must be readable. A saved run whose dataset no longer loads is reported, with the reason, until its folder is fixed.

If an existing saved annotation, model review or human-verdict file cannot be decoded, names no claim, holds two entries for one claim or spells a field two ways, the run opens with an explicit warning and refuses further annotation or verdict writes until the file is recovered. It does not silently treat the damaged result as unfinished work. The dashboard does not keep such a run open, so it is read again once the file is repaired, without a restart. Unreadable lines in the append-only call log are counted separately. Reported served-model counts distinguish annotator and reviewer roles, include format retries, and label responses with no provider-reported model as unknown. These are provider reports, not proof of the model's implementation or independence.

A reviewer's reply is checked before it counts: the claim it names, if it names one, is the one it was asked about; it decides at least one field or gives an adjudicated annotation, which must read as an annotation of that claim; each field's choice is one of the annotations it was shown, or NONE; its confidence is between 0 and 1; and no key is spelt two ways. A reply that fails is a failed review, which the next review does again. A reopened run checks a reply it would take from `calls.jsonl` the same way: one that fails is not taken, and its claim is reviewed again. Reviews already in the reviewer's file are not checked again, so a run made before 24 November 2026 keeps the reviews it saved.

When a file of a run cannot be written (a full disk, a folder that cannot be written to), the run asks the models nothing new: the annotators and the reviewer finish the call under way, and its retries stop. A call already with a provider finishes and is logged; one waiting for its turn or for a retry stops. `wag-llm annotate` and the dashboard say which record failed and what the files lack, and so does `report.json`. What was received is kept in memory, counted as not saved rather than done; the dashboard's *Write them now*, or the next start, writes it before anything is asked, and a start does not begin until it can. The command line tries once more before it ends: what it still cannot write is lost then, except a reply that reached `calls.jsonl`, which a continued run takes from there instead of asking again, exactly as it came (a reply the run refused is not taken). A verdict that cannot be written is not kept: give it again. A verdict that is written but whose folder cannot be synced is kept, as the file keeps it, and the dashboard says it may not survive a power loss yet; the folder is synced again with the next verdict, or with *Sync it now*. Lines of `calls.jsonl` are written whole: a write that fails part way is cut off, and half a line left by a crash is ended before the next is written, so that it costs only itself. A record counts as written once its bytes are synced to the disk, and its folder too (the folder's entry for a new or replaced file, which a power loss could otherwise undo; Windows journals these itself). The other files are replaced whole, through a temporary file. A run made from a Windows checkout, whose pinned inputs have CRLF line endings, continues on the same files with LF ones.

## The command line: quick start with Gemini

1. **Make a key.** Create an API key in Google AI Studio.
2. **Put the key in your shell's environment, nowhere else.** Never in a file of the repository, a profile, a command-line flag or a chat.
   ```sh
   export GEMINI_API_KEY=…            # in your shell, or your shell's profile
   ```
   WAG reads `GEMINI_API_KEY`, then `GOOGLE_API_KEY`. Profiles name the variable (`api_key_env`), never the key, and refuse a field named like a key. Traces never contain it.
3. **Build the tools.**
   ```sh
   cd wag
   go build -o wag-llm ./cmd/wag-llm
   go build -o wag-server ./cmd/wag-server
   ```
4. **Check the key and the model.**
   ```sh
   ./wag-llm models gemini                          # the models your key can use
   ./wag-llm ping --model gemini:gemini-2.5-flash   # one short request: text, model version, usage, latency
   ```
   The model name here is an example: use one that `models` lists.
5. **Play in the browser.** A Gemini seat appears in *New game → Players*.
   ```sh
   ./wag-server --model flash=gemini:gemini-2.5-flash,min_interval=13s,output=json
   ```
6. **Play headless, collecting traces.**
   ```sh
   ./wag-llm play --model flash=gemini:gemini-2.5-flash,min_interval=13s,output=json \
       --pro flash --con greedy --games 5 --swap --out runs/flash-vs-greedy
   ./wag-llm play --model flash=gemini:gemini-2.5-flash,min_interval=13s,output=json \
       --pro flash --con flash --games 3 --title "Should governments tax meat?" --out runs/flash-self
   ./wag-llm report runs/flash-vs-greedy             # wins and rejected moves by seat and side
   ./wag-llm replay runs/flash-self/<file>.trace.jsonl
   ```
   `replay` uses the recorded answers, so it makes no calls. It checks that the game sent the same requests and made the same decisions. It refuses a trace played with other scheme data (give that data with `--data`) and uses the context presets the trace recorded.

### The free tier: privacy and limits

On the Gemini API's unpaid tier, Google may use prompts and answers to improve its products, and people may review them (Gemini API Additional Terms, `https://ai.google.dev/gemini-api/terms`). Send only public, non-confidential material.

WAG's prompts hold the rules, the board, the schemes, and the topic and arguments of the game. Kialo discussions are public; your own unpublished material may not be. A paid (billed) project is under different terms. Check your project's status and the current terms before you send research material.

Free-tier limits are low and change; AI Studio shows your project's. They are counted per minute (requests and tokens) and per day (requests). On 2026-10-15, Gemini 2.5 Flash's free tier allowed 5 requests and 250,000 tokens a minute, and 20 requests a day.

- **Spacing calls.** `min_interval` spaces a profile's calls: `13s` stays under 5 requests a minute. The dashboard proposes 13 seconds for Gemini.
- **Retries.** A rate-limited call is retried after the delay Gemini asks for, up to `retries` times (default 3). Every try is traced. If Gemini asks for more than two minutes, a daily quota is used up: the call is not retried, and an annotation run stops that annotator.
- **What a game costs.** A WAG prompt is 4,000–12,000 tokens. A move takes one call, or more after a rejected reply, and a retry resends the conversation so far. Twenty requests a day is one short game; a paid tier or another model is needed for real runs.
- **Thinking.** Gemini 2.5 Flash thinks before it answers, by default as long as it likes (up to about 7,600 thinking tokens per annotation in our first run). `thinking_budget` in a profile caps it (0 turns it off); whether that costs accuracy has not been measured.

### A run's budget (D046)

`wag-llm play` and `wag-llm annotate` count what each invocation uses, and
can limit it. **No limit is the default** (the owner, 2026-11-23): a local
model, a prepaid plan or a free tier may need none. `--unlimited` says so
explicitly; it cannot be combined with a limit. The limits cover every client
of the invocation together, both annotators of a run, say.

- `--max-calls N` stops the invocation before its N+1th **client attempt**.
  WAG's retries are attempts. The cap covers calls through WAG's client; it
  does not count a provider's hidden internal work or calls made inside a
  `command` profile's program.
- `--max-tokens N` refuses a call when its **estimated reservation**, plus
  accounted and in-flight usage, would exceed N. The estimate uses supplied
  system/message text bytes, supplied structured-output schema bytes, and
  `max_output_tokens`. It is a proxy, not a tokenizer or a proven billed-token
  ceiling: provider framing, hidden prompt additions and provider behavior can
  differ. Every estimated call needs a positive `max_output_tokens`. Where a
  provider's output limit does not cover reasoning it reports apart (check
  its terms), one call can use more than its estimate, and that overrun stops
  a token- or cost-limited invocation; `--max-calls` has no such problem.
- Gemini estimates additionally require a fixed, nonnegative `thinking_budget`
  within the provider field's integer range. Missing reasoning usage retains
  that allowance unless a coherent reported total supplies it. Whether a model
  honors a requested thinking setting remains a provider/model contract.
- `--max-cost X` uses `--prices FILE`, whose currency and rates the owner chooses.
  Rates are per million tokens, by profile name, then model, then `"*"`:
  `{"gemini-2.5-flash":{"input":0.3,"output":2.5}}`. Every entry must explicitly
  provide finite, nonnegative input and output rates; explicit zero is allowed.
  This is a price-table estimate, not an invoice ceiling. It does not discover
  current prices, tier changes, cache charges or other provider fees.
  Rates are resolved from the configured profile/model, also when a provider
  answers with another model (a fallback); native callers that override a
  request's model must keep that price mapping aligned themselves. `--prices`
  without `--max-cost` prices what is used without limiting it.

A profile's `extra` fields can replace its messages or output limit. Therefore
nonempty `extra` is refused when token/cost estimation is enabled. Runs with no
limit or a call cap alone keep their custom fields.

**What is counted.** Every invocation counts its client attempts, the tokens
the providers report, their price-table cost when `--prices` is given, and the
attempts that reported no usage, failed ones included. An unusable answer
(empty, refused or blocked) still fails the call, and the usage the provider
reported for it is counted and kept in the trace: a reasoning model can use its
whole output limit thinking and answer nothing. Reasoning is counted as each
provider reports it: Gemini's thoughts apart from the answer; other providers'
reasoning within the completion, unless it is larger than the completion, which
it then cannot be part of (xAI reports it apart). A reported total larger than
its parts is counted whole, the excess as output; a total smaller than the parts
that are known is malformed. Without an input count, the total less a complete
output is the input; with neither, the total is priced at the dearer rate.
Without a token or cost limit, an attempt with no reported usage adds no tokens,
and the count says how many there were. Usage that a given price table has no
rate for is counted as unpriced, with no cost.

Under a token or cost limit, coherent reported usage replaces the corresponding
estimates after a successful attempt. Missing components retain conservative
fallbacks; failed attempts retain their reservation because they may be
billed. Negative or inconsistent usage cannot create credit: the checker
charges at least the reservation and any larger nonnegative observations,
records the problem and stops later attempts. Usage exceeding the reservation
is also retained and stops later attempts; a cost is compared with its
reservation only under a cost limit. The actual response and usage remain in
the trace alongside the accounting error. Already dispatched concurrent calls
cannot be undone, and a reported overrun demonstrates why estimates are not
hard billing guarantees.

A refused call or accounting violation has error kind `budget`. Annotation
stops the affected annotator; `play` saves its trace/export and starts no further
game.

**The record.** At the end the invocation prints what it used and appends one
line to `budget.jsonl` in its output folder: the command, its start and end,
its limits or `"unlimited": true` (with `"default": true` when no limit was
given, rather than `--unlimited`), the price table's file name and SHA-256,
what was counted (`spent`: `calls`, `tokens`, `cost`, `unreported`,
`unpriced`), whether a call was refused before it was sent (`refused`), and
why an accounting violation stopped later attempts (`halted`). Each invocation
has its own allowance, a resumed annotation run included, and its own line, so
a folder's lines sum to what `wag-llm` spent in it; the dashboard records no
budget yet. `--unlimited` with `--max-calls 0` is accepted: 0 is no limit.

**Keys.** A profile whose provider needs an API key, or that names one
(`api_key_env`, or a key typed into the dashboard), may be billed. With no
limit by default, the invocation prints one note naming those profiles;
`--unlimited` drops it. A continued annotation run calls, and names, the
annotators and reviewer saved in it, whatever the invocation's `--annotators`.
As D046 says, runs with the owner's keys run on the owner's machine. No live
API call was needed to test these controls: the tests use mock providers and
local test servers. The dashboard does not take a budget yet (stage 2): its
runs have no limit, and no record beyond their traces.

## Profiles

A profile is a configured model. Give profiles on the command line, repeatable:

```
--model [NAME=]PROVIDER:MODEL[,option=value…]
```

Or put them in a file and pass `--models profiles.json`:

```json
{"profiles": [
  {"name": "flash", "provider": "gemini", "model": "gemini-2.5-flash",
   "temperature": 0.7, "output": "json", "min_interval": "13s"},
  {"name": "flash-nothink", "provider": "gemini", "model": "gemini-2.5-flash", "thinking_budget": 0},
  {"name": "claude", "provider": "anthropic", "model": "claude-opus-5-5", "effort": "medium"},
  {"name": "gpt", "provider": "openai", "model": "gpt-5", "effort": "low"},
  {"name": "deepseek", "provider": "deepseek", "model": "deepseek-chat"},
  {"name": "local", "provider": "lmstudio", "model": "qwen3-8b", "output": "json"},
  {"name": "llama", "provider": "llamacpp", "base_url": "http://127.0.0.1:8081/v1", "model": "any"},
  {"name": "guided", "provider": "command", "command": ["python3", "my_guidance_player.py"]}
]}
```

The model names are examples; use the ones your provider lists.

| Field | Meaning |
|---|---|
| `name` | What seats call it. |
| `provider` | `wag-llm providers` lists them; see below. |
| `model` | The provider's model name. |
| `base_url` | The endpoint, for a local or other compatible server. |
| `api_key_env` | The environment variable that holds the key, if not the provider's usual one. |
| `temperature`, `top_p`, `top_k`, `max_output_tokens`, `seed` | Sent only when set; otherwise the provider's defaults apply. |
| `effort` | How much the model reasons: Gemini's thinking level (`minimal`, `low`, `medium`, `high`), Claude's effort (`low` … `max`), OpenAI's `reasoning_effort`. |
| `thinking_budget` | Gemini 2.5's thinking budget in tokens (`0` turns it off where the model allows; `-1` lets the model decide). |
| `output` | `text` (default); `json` (the provider's JSON mode: the reply is the JSON envelope alone); `schema` (JSON constrained to `schema_file`, where the provider supports it). |
| `extra` | Fields added to the request as they are (an engine's own options, e.g. llama.cpp's `grammar`). |
| `system` | Replaces the seat's system prompt (to compare wordings). |
| `timeout` | Bounds one call (default `5m`). |
| `min_interval` | Spaces the profile's calls. |
| `retries` | Retries after a rate limit, a server error or a timeout (default 3; `-1` for none). Errors that will not pass (a bad key, a blocked answer) are not retried. |
| `command` | For the `command` provider. |
| `replies` | For the `mock` provider. |

Gemini's thoughts are never requested (`include_thoughts` stays off). Only how much the model thinks is configurable, and the trace records the thinking tokens it reports. A reasoning block that some local engines put before the answer (`<think>…</think>`) is removed before parsing and is not kept.

## Providers

| Provider | What | Key (environment) |
|---|---|---|
| `gemini` | Gemini Developer API, through Google's GenAI SDK (`google.golang.org/genai` v1.73.0, pinned) | `GEMINI_API_KEY` or `GOOGLE_API_KEY` |
| `anthropic` | The Claude API, through Anthropic's SDK | `ANTHROPIC_API_KEY`, or an `ant auth login` profile |
| `openai` | OpenAI's Chat Completions API | `OPENAI_API_KEY` |
| `deepseek` | DeepSeek's API (OpenAI-compatible) | `DEEPSEEK_API_KEY` |
| `lmstudio`, `llamacpp`, `sglang`, `vllm`, `ollama` | Local servers' OpenAI-compatible endpoints. Default URLs: `localhost:1234`, `8080`, `30000`, `8000` and `11434`, each with `/v1`. | none |
| `openai-compatible` | Any other compatible endpoint (`base_url` required) | `api_key_env`, if it needs one |
| `command` | An external program, for harnesses: Guidance, DSPy, a constrained decoder, a model without a server | — |
| `mock` | Scripted answers, for tests and dry runs | — |

- **Compatible servers differ.** WAG sends only the parameters a profile sets. A server that rejects one says so in the trace (an `invalid_request` error).
- **Port clash.** llama.cpp's default port, 8080, is also `wag-server`'s. Run one of them elsewhere (`--addr 127.0.0.1:8090`, or `base_url`).
- **The command protocol.** The program gets the request as JSON on its standard input: `system`, `messages` (each with a role and content) and the parameters. It answers on its standard output, with either a JSON object (`{"text": "…", "model": "…", "finish_reason": "…", "usage": {"input_tokens": …}}`) or plain text. A non-zero exit is an error, reported with its standard error.

## Traces

Every game has a trace file (`wag-trace/2`): the browser's *Traces* link, or `wag-llm play`'s `.trace.jsonl` files. `docs/protocol.md` ("What is recorded") and `docs/protocol/trace.schema.json` describe it.

The events, in order:
1. `game_start`: the setup, the seats' profiles, the seed, the scheme data's hash, the build.
2. `call`: every call, with the whole conversation sent, the answer or the classified error, usage, latency, and request and response IDs.
3. `decision`: every decision, with the board, the reply, the stated reasoning, the move, its result and effects, predictions checked, and alternatives tried.
4. `game_end`: the outcome, every final label, and the verdicts under each context.

`wag-llm play` also writes each game's export, which the browser can load, a line of `summary.jsonl` for each game, and a line of `budget.jsonl` for each invocation ([its budget](#a-runs-budget-d046)). The summary line holds:
- the winner;
- the decisions, how many were rejected and how many could not be parsed;
- the calls and how many failed, by kind;
- tokens and call time;
- prediction accuracy, and the share of named alternatives that were legal.

## Testing without a key, and with one

- **Without a key: `go test ./internal/llm ./internal/server ./cmd/wag-llm`.** It runs:
  - each adapter through its real SDK or HTTP code, against a local fake of the provider's API (request shape, key in a header and never in the URL or a trace, parsing, usage, rate limits with the delay asked for, blocked answers);
  - the model seat's conversation and retries;
  - the trace events;
  - a headless game, its replay, and a tampered trace that must not replay.
- **With a key, on purpose only.** `WAG_LIVE_TESTS=gemini WAG_LIVE_MODEL_GEMINI=gemini-2.5-flash go test ./internal/llm -run TestLive -v` sends one short request per provider named. It is skipped unless `WAG_LIVE_TESTS` is set.

## Using the traces as data

**What they are good for.**
- **Evaluation.** Legal-move rate, rejections by reason, prediction accuracy (`expect_check`), win rates by side and model (`wag-llm report`).
- **Debugging prompts and rules.** A rejected move's reason and the conversation that followed.
- **Supervised data of moves the validator accepted**, with their effects.
- **Preference pairs** (DPO-style): the move made against the alternatives the model named, each tried on the game. Treat these with care: a legal alternative is not a worse argument because it scored less.

**What they are not.**
- **Not ground truth about reasoning.** The `thoughts` and `summary` a model writes are its own account, not a record of how it chose (`docs/protocol.md`). Do not train a model to imitate them as if they were.
- **Not evidence of argument quality.** A move the game accepts is legal: its scheme fits, its question applies. It is not shown to be true or relevant. Winning a WAG game under one proof standard and one set of weights is not arguing well.
- **Not a reward on their own.** A policy trained on game score alone can learn to exploit the rules: repeated undercutting, `Other` arguments, burden-of-proof wins. Add human judgements of relevance, truth, scheme fit and question handling before any RL.

**Keeping a dataset clean.**
- **Version.** Record the data hash, protocol and build with every game; `game_start` does. Do not mix games played on different scheme data without saying so.
- **Splits.** Split by topic or discussion, not by game or move, so that one discussion's arguments are not in both training and test data.
- **Failed calls.** Keep the failed calls: rate limits, blocked answers and unparseable replies are data about the providers, not noise to drop silently.
