# Protocol samples

On 2026-10-13 a language model played two games through the text protocol alone. It saw only the prompts, and its replies were the protocol 2 JSON. Neither game had a rejected move.

| File | What |
|---|---|
| `playtest-b.traces.jsonl` | Game B's decisions (fare-free buses; the model played both sides, 10 moves), as `wag-trace/1` lines with the prompts left out. Each line has the reply, the stated `context`, `thoughts` and `summary`, the predictions checked against the labels after the move, and the alternatives tried on the game. |
| `prompt-game-b-move10.txt` | The prompt the current build writes for game B's tenth move (Bob to move), after replaying the first nine moves. Compare it with the prompts the model saw, which are summarized in `docs/review-2026-10-13-part5.md`. |

The game cannot be replayed past move 9 on the current data. Move 10 backed "Other Means" with positive consequences, and the data no longer allows that (the audit's N9). An export now records the hash of the scheme data it was played with (`data_sha256`), and loading says when the data differ.

These lines are `wag-trace/1`, the format before 2026-10-15. `wag-trace/2` (`../trace.schema.json`) keeps the same decision fields and adds `type`, `id`, `order`, `turn`, `feedback` and the model seat's profile; the file also has `game_start`, `call` and `game_end` events.
