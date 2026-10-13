# WAG — Waltonian Argumentation Game

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
| `data/` | The game data: `schemes/core.yml` (Carneades schemes and their English `language` templates) and `nascq/*.json` (NASCQ schemes, critical questions, edges). Package `data` embeds them. |
| `internal/carneades` | Adapter over the Carneades argument graph: moves as graph operations, checkpoints, labelling. |
| `internal/nascq` | NASCQ loading, move validation (S0–S8, Q1), action enumeration, data lint. |
| `internal/game` | Turns, modes, scoring, termination, invariants, agents (random, greedy, human, JSON), JSON view, batch runner. |
| `cmd/wag-cli` | Command-line game, data lint and tournaments. |

## Commands

```sh
go test ./...                      # all tests (about a minute); -short for a quick run
go test -race -short ./...
go test -run '^$' -fuzz FuzzTextToAtom -fuzztime 30s ./internal/carneades/

go run ./cmd/wag-cli                                   # you (human) against a random agent
go run ./cmd/wag-cli --p1 human --p2 human             # two people at one terminal
go run ./cmd/wag-cli --p1 greedy --p2 random --seed 7  # watch two agents
go run ./cmd/wag-cli --lint                            # check the data; exit 1 on an error
go run ./cmd/wag-cli --data ./data --lint              # check data files on disk (no rebuild)
go run ./cmd/wag-cli --games 50 --p1 greedy --p2 random  # tournament statistics
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

`../docs/review-2026-10-13.md` and `../docs/review-2026-10-13-part2.md`
explain the design, what is still missing and the balance measurements.
