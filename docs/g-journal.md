# G: a model's paid reply, kept before it is used

Stage 2 ends when, among other things, "every recorded event replays without asking a model again"
(plan.md, stage 2). Recording-1 of the external review's acceptance matrix asks for the case it
names: a model's reply has been paid for, and the process stops before the move it gives is
accepted; on resuming, the reply is recovered, and the model is not asked (and paid) again. This
page records how games with a model seat stand today, what changes, and the cases, each with its
outcome written before its test ran.

## How it stands (2026-11-27)

- **Annotation runs already do this.** Each call goes to an append-only log (`calls.jsonl`), and a
  reopened run takes a reply from the log, checked as a new one is, instead of asking again
  (llm.md).
- **A game with a model seat does not.** The seat's turn (`server/llm.go`, `modelTurn`):
  - asks the provider (`Complete`);
  - records the call in memory (`addCall`, from the attempt callback);
  - parses the reply and plays its move (`play`);
  - then saves (`autosave`): the trace file and the export, each replaced whole and synced.
- **The reply is on disk only after its move.** If the process stops between the provider's answer
  and that save, the reply is lost.
- **A saved game is reopened from its export alone** (`/api/load`). Its `trace.jsonl` is never read
  back, so the reopened game:
  - asks the model again for the turn;
  - pays again for a reply already received;
  - has a trace that lacks the calls made before it stopped.

## What changes

1. **Kept before used.** A call's record, with the raw reply or the error, is saved as soon as the
   provider answers, before the reply is parsed or its move played. If that save fails, the reply
   stays in memory, the game says so, and the move waits until a save succeeds: a model seat waits
   today when a save fails.
2. **Reopened with its trace.** A game in the data folder is resumed from its folder
   (`POST /api/store/resume` with the folder's path, from this computer only, as the store's other
   actions). The server reads its `export.json` and its `trace.jsonl`. The trace must be the
   export's: the same game, data and moves. Otherwise the resume is refused, saying which. The
   calls and decisions recorded carry over into the resumed game, whose trace continues them, and
   which saves into the same folder. Loading an export alone (`/api/load`) stays as it is: it brings
   no trace, and starts a new record.
3. **A reply used once.** When a model seat is to move, and the game's calls (carried over by a
   resume, or kept in memory after a failed save) hold a call for this move (the same turn: the
   number of moves made) whose reply no decision used, the seat plays that reply and asks the
   provider nothing. The decision records that the reply was recovered (`recovered`) and has the
   call as its parent. If the reply's move is refused, the turn goes on as a live turn does. The
   conversation for its next attempt is the recorded call's request, then the reply, then the
   reason for the refusal. These are not used again:
   - a call that failed (no reply);
   - a reply a decision already used;
   - a reply given for another turn.

## Cases (outcomes written before the tests)

The provider is a fake that counts its calls and answers with a fixed move; the process's stop is a
test seam that ends the turn after the reply is recorded and before its move is played.

| Case | Setup | Expected |
|---|---|---|
| G1 | A game with a model seat on its turn; the turn stops after the provider's answer | The game's folder holds the call with its reply; the export has no move for it |
| G2 | G1's folder, loaded into a new server; the model seat's turn | The provider is not called (its count is unchanged). The move is the recorded reply's. The trace has the call once and a decision with `recovered`, whose parent is the call |
| G3 | As G2, but the recorded call failed (an error, no reply) | The provider is called again, once |
| G4 | A game whose model seat's reply was used (its move played), saved and loaded | No reply is reused: the next model turn calls the provider |
| G5 | G1's folder, with a trace of another game (another id, data or move list) | The resume is refused, saying which |
| G5b | G1's folder, with its moves recorded as played but its last call's reply refused as a move (a second attempt was due) | The resumed turn plays the recorded reply, is refused, and asks the provider once, with the recorded conversation, the reply and the refusal's reason |
| G6 | G2's game, exported and replayed (`wag-llm --replay`) | It replays without asking a model: every decision's reply is in the trace |
| G7 | A save that fails right after the provider's answer | The move waits; the game says why; once a save succeeds, the move is played from the reply in memory, with no new call |

## Planted defects (each must be caught)

| Defect | Caught by |
|---|---|
| The call is saved after the move, not before | G1 |
| The trace is not read on load | G2 |
| A failed call is taken as a reply | G3 |
| A used reply is used again | G4 |
| A reply of another turn is used | G2 (a reply for turn n+1 is planted and must not be used at turn n) |
| A trace of another game is accepted | G5 |
