# WAG — Waltonian Argumentation Game

The project's overview, status, issues and plan are in the central [`README.md`](../README.md).

A turn-based argumentation game: players put forward arguments built from
Walton's argumentation schemes, and attack, defend and question each other's
arguments through their critical questions, to any depth. A NASCQ graph
(schemes, critical questions, ask and backup edges, after Pyon & Ri 2022)
decides which moves are legal; Carneades-4 (the fork in `../carneades-4`)
labels every statement In, Out or Undecided after each move. Eight modes
(points races, timed and speed play, critical-question duels, debates on a
central claim, solo belief revision) are combinations of a few settings.

## Layout

| Path | What |
|---|---|
| `data/` | The game data: `schemes/core.yml` (Carneades schemes and their English `language` templates) and `nascq/*.json` (NASCQ schemes, critical questions, edges). Package `data` embeds them. Also `catalogue/`, the scheme catalogue: one entry per scheme, aligning the book, `walton.yml` and the NASCQ ([`data/catalogue/README.md`](data/catalogue/README.md)). |
| `internal/carneades` | Adapter over the Carneades argument graph: moves as graph operations, checkpoints, labelling. |
| `internal/nascq` | NASCQ loading, move validation (S0–S9, Q1), action enumeration (answers to critical questions, and arguments for or against any statement), data lint. |
| `internal/registry` | Scheme patterns and families: finds schemes that repeat a pattern under another name, and close relatives no family declares (`docs/scheme-registry/README.md`). |
| `internal/catalogue` | The scheme catalogue: its entry format and statuses, the generator that makes and merges the skeleton from the sources, the checker, and the copy check against the book's text ([`data/catalogue/README.md`](data/catalogue/README.md)). |
| `internal/game` | Turns, modes, scoring, termination, invariants, agents (random, greedy, human, JSON), JSON view, batch runner. |
| `internal/kialo` | Import of Kialo discussions (`graph.json` of the Kialo research visualizer) as games of "Other" arguments, one statement per claim (keyed by claim ID, so two claims with the same words stay apart). |
| `internal/server` | HTTP server and browser interface (`web/`, embedded): the dialogue and board views, moderation, notes, export and replay, the NASCQ, Kialo import; the text protocol for language models (`protocol.go`), model seats (`llm.go`) and their traces (`traces.go`). |
| `internal/llm` | The provider-neutral interface to language models: Gemini (Google's GenAI SDK), Claude (Anthropic's SDK), OpenAI-compatible endpoints (OpenAI, DeepSeek, LM Studio, llama.cpp, SGLang, vLLM, Ollama), an external command, a mock; profiles, retries, rate spacing, replay (`../docs/llm.md`). |
| `cmd/wag-cli` | Command-line game, data lint, the table of every scheme–question link (`--edges`), tournaments, and the catalogue's generator and checker. |
| `cmd/wag-llm` | Language models play WAG headless and annotate Kialo claims, with full traces; replay a trace; sum up runs (`../docs/llm.md`). |
| `cmd/wag-server` | The browser prototype: `go run ./cmd/wag-server` opens it at http://127.0.0.1:8080. |

## In the browser

```sh
go run ./cmd/wag-server                                  # serves http://127.0.0.1:8080 and opens it
go run ./cmd/wag-server --kialo-dir <visualizer>/data    # also offer the captured Kialo discussions
go run ./cmd/wag-server --claude                         # also offer Claude seats (ANTHROPIC_API_KEY; costs API credits)
go run ./cmd/wag-server --model gemini:gemini-2.5-flash,min_interval=13s  # a Gemini seat (GEMINI_API_KEY)
```

The **Models** tab connects a model service with a key typed in the browser,
finds a model and adds it as a player (a seat in any game); the **Annotate**
tab has players annotate Kialo claims, a reviewer compare them, and you give
your verdicts. Everything is saved in your data folder (`~/WAG`, or
`--store DIR`): `../docs/llm.md`, "The dashboard".

Building from source needs Go 1.24 or later (the Anthropic and Google SDKs require it).
A seat can be a person, a language model through the text protocol
(`../docs/protocol.md`: copy the prompt into any chatbot, paste the reply
back), a model through its API (Gemini, Claude, OpenAI, DeepSeek, a local
server: `../docs/llm.md`), or a rule bot (legal moves, placeholder words).

On Windows, double-click `run-wag-server.bat` (needs Go), or run a prebuilt
`wag-server.exe` (`GOOS=windows GOARCH=amd64 go build ./cmd/wag-server`).
`../docs/prototype.md` explains what to check, step by step.

## Commands

```sh
go test ./...                      # all tests (one to two minutes); -short for a quick run
go test -race -short ./...
go test -run '^$' -fuzz FuzzTextToAtom -fuzztime 30s ./internal/carneades/

go run ./cmd/wag-cli                                   # you (human) against a random agent
go run ./cmd/wag-cli --p1 human --p2 human             # two people at one terminal
go run ./cmd/wag-cli --p1 greedy --p2 random --seed 7  # watch two agents
go run ./cmd/wag-cli --lint                            # check the data; exit 1 on an error
go run ./cmd/wag-cli --data ./data --lint              # check data files on disk (no rebuild)
go run ./cmd/wag-cli --registry data/schemes/core.yml --families data/schemes/families.json  # one pattern, one name
go run ./cmd/wag-cli --games 50 --p1 greedy --p2 random  # tournament statistics
go run ./cmd/wag-cli --kialo-check <visualizer>/data   # import every captured Kialo discussion; report claims that share a text
go run ./cmd/wag-cli --catalogue-gen                   # (re)generate the scheme catalogue's skeleton, keeping what was written
go run ./cmd/wag-cli --catalogue-check [--book FILE]   # check the catalogue; exit 1 on an error
go run ./cmd/wag-cli --catalogue-show ID               # a walton.yml scheme's formulas with their English readings
go run ./cmd/wag-cli --modes                           # list the game modes
go run ./cmd/wag-cli --mode debate --p2 human          # debate a claim against a friend
go run ./cmd/wag-cli --mode cq --p1 greedy --p2 random # critical-question duel
go run ./cmd/wag-cli --mode solo                       # argue both sides yourself
go run ./cmd/wag-cli --mode classic --scoring live --answer-within 1  # mix the settings
go run ./cmd/wag-cli --mode debate --games 40 --p1 greedy --p2 greedy --alternate  # self-play
```

### Playing from another program (LLMs, Python)

`--p1 json` (or `--p2 json`) hands a player to the program on the other end of
standard input and output. Whenever that player is to move, the CLI writes the
game as one JSON line (`game.View`: board, scores, every legal action with its
question and what it concludes) and reads one JSON line back
(`nascq.ProposedAction`):

```json
{"type":"support","scheme_id":"AS:practical_reasoning","variable_bindings":{"A":"carbon tax","S1":"status quo","S2":"lower emissions","G":"reduce emissions","V":"public health"}}
{"type":"attack","scheme_id":"AS:sign","answers_cq_id":"CQ:possibility","target_arg_id":"a1","variable_bindings":{"A":"no majority"}}
{"type":"support","scheme_id":"AS:expert_opinion","answers_cq_id":"CQ:possibility","target_arg_id":"a1","variable_bindings":{"W":"treasury","D":"fiscal policy"}}
{"type":"question","answers_cq_id":"CQ:possibility","target_arg_id":"a1"}
{"type":"abstain"}
{"type":"resign"}
```

A rejected move is followed by a new view whose `last_rejection` says why.
Everything else the CLI prints goes to standard error.

## Rules as implemented

The rulebook is `../docs/game-rules.md`; in brief:

- **Moves.** *Argue*: instantiate a scheme; its premises are presumed (they
  hold until attacked). *Attack* or *defend*: answer a critical question about
  an argument negatively or positively, with a scheme that can back the
  answer; the question's locus decides what the answer argues (against or
  for a premise, an exception, the conclusion, or the inference). Answers
  are arguments with their own critical questions, so disputes go as deep
  as the players take them. *Question*: ask without answering; the premise
  (or inference) is out until the proponent answers. *Pass*, *resign*.
- **Validity.** NASCQ checks S0–S8 and Q1 (see `nascq.Validate`), no cycles,
  no restating an argument already in the graph (G1), no asking a question
  twice (Q2). An invalid move costs 5 points and the player moves again.
- **Score.** 10 × weight for an argument whose conclusion is In (2 × weight
  if Undecided), +15 for an undercut that defeats its target; either once
  per move (`move`) or kept up to date as labels change (`live`). Attacks
  on your own or already defeated arguments, and questions, score 0.
- **Objectives and end.** Points (highest score) or a central claim (pro
  must carry it under its proof standard; con wins otherwise). Games end
  by resignation, an unanswered question (cq), a concession (debate), a
  target score, a clock, a round limit or everyone passing, as the mode
  says.

Options: `AllowOther` (the user-defined "Other" scheme, for arguments no
Walton scheme fits), `AnyPremise` (the generic premise question for every
premise no critical question covers; on in every mode), `RequireEvidence`
(schemes marked `evidence: required` need a source), `Moderated` (a moderator
approves each move and can accept or reject statements), and a context
preset (`data/contexts.json`: proof standard, sources, draft scheme weights).

`../docs/review-2026-10-13.md`, `../docs/review-2026-10-13-part2.md` and
`../docs/review-2026-10-13-part3.md` and `../docs/review-2026-10-13-part4.md`
explain the design, what is still missing and the balance measurements.
`../docs/walton-audit/` audits the Walton scheme collection,
`../docs/book-catalogue/` catalogues the schemes and critical questions of
Walton, Reed & Macagno (2008), pp. 1–138, against WAG's, and
`../docs/kialo-pilot/` is a pilot of formalizing Kialo claims as schemes.
