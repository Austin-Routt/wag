# WAG — Waltonian Argumentation Game

A turn-based argumentation game: players put forward arguments built from
Walton's argumentation schemes and attack each other's arguments by asking
critical questions. A NASCQ graph (schemes, critical questions, ask and
backup edges) decides which moves are legal; Carneades-4 (the fork in
`../carneades-4`) labels every statement In, Out or Undecided after each move.

## Layout

| Path | What |
|---|---|
| `data/` | The game data: `schemes/core.yml` (Carneades schemes and their English `language` templates) and `nascq/*.json` (NASCQ schemes, critical questions, edges). Package `data` embeds them. |
| `internal/carneades` | Adapter over the Carneades argument graph: moves as graph operations, checkpoints, labelling. |
| `internal/nascq` | NASCQ loading, move validation (S0–S8), action enumeration, data lint. |
| `internal/game` | Turns, scoring, termination, invariants, agents (random, greedy, human, JSON), JSON view, batch runner. |
| `cmd/wag-cli` | Command-line game, data lint and tournaments. |

## Commands

```sh
go test ./...                      # all tests (about 30 s); -short for a quick run
go test -race -short ./...
go test -run '^$' -fuzz FuzzTextToAtom -fuzztime 30s ./internal/carneades/

go run ./cmd/wag-cli                                   # you (human) against a random agent
go run ./cmd/wag-cli --p1 human --p2 human             # two people at one terminal
go run ./cmd/wag-cli --p1 greedy --p2 random --seed 7  # watch two agents
go run ./cmd/wag-cli --lint                            # check the data; exit 1 on an error
go run ./cmd/wag-cli --data ./data --lint              # check data files on disk (no rebuild)
go run ./cmd/wag-cli --games 50 --p1 greedy --p2 random  # tournament statistics
```

### Playing from another program (LLMs, Python)

`--p1 json` (or `--p2 json`) hands a player to the program on the other end of
standard input and output. Whenever that player is to move, the CLI writes the
game as one JSON line (`game.View`: board, scores, every legal action with its
question and what it concludes) and reads one JSON line back
(`nascq.ProposedAction`):

```json
{"type":"support","scheme_id":"AS:practical_reasoning","variable_bindings":{"A":"carbon tax","S1":"status quo","S2":"lower emissions","G":"reduce emissions","V":"public health"}}
{"type":"attack","scheme_id":"AS:sign","answers_cq_id":"CQ:possibility","target_arg_id":"a1","variable_bindings":{"A":"no majority","B":"the bill cannot pass"}}
{"type":"abstain"}
```

A rejected move is followed by a new view whose `last_rejection` says why.
Everything else the CLI prints goes to standard error.

## Rules as implemented

- **Moves.** *Support*: instantiate a scheme with your bindings; its premises
  are presumed (they hold until attacked). *Attack*: pick an argument, a
  critical question asked of its scheme, and a scheme that can back an answer;
  the question decides the attack: rebuttal (argue against the conclusion),
  undercutter (the inference does not apply), or undermining (argue against
  the premise the question names). *Abstain.*
- **Validity.** NASCQ checks S0–S8 (see `nascq.Validate`), no cycles, and no
  restating an argument already in the graph (G1). An invalid move costs 5
  points and the player moves again.
- **Score.** For each argument you put forward: 10 × weight if its conclusion
  is In right after the move, 2 × weight if Undecided; +15 when an undercutter
  turns its target's conclusion Out. Attacks on your own arguments score 0.
- **End.** After `--turns` rounds, or when every player abstains for three
  rounds running. Highest score wins.

The rules are a first version; see `../docs/review-2026-10-13.md` for what is
known to be missing (a central claim, defence moves, termination by the
claim's status) and for the balance measurements.
