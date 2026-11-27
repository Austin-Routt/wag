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

## Amendments before the tests ran (2026-11-27)

Writing the code showed points the plan above gets wrong or leaves open. They are settled here, before
any test of G ran; the cases and planted defects below are added on the same terms (outcomes written
first).

1. **Which reply is for this move.** The plan takes a reply to be for this move when the number of
   moves made is its call's turn. But a seat's move refused by a rule is recorded in the history
   with its penalty (`rejectInvalid`), so the number grows during the seat's own turn, while the
   turn's calls keep its start. A reply is for this move when the moves made since its call are only
   that seat's refused moves.
2. **A reply that came with an error is a failed call.** The client keeps a provider's answer that
   came with an error (a partial or blocked answer, or one the budget refused: `Complete`). The live
   turn does not play it, and neither does a resumed one.
3. **The turn goes on as the live turn would.** The turn in progress is read from its record: its
   calls, attempt by attempt, and the decisions that used them. The replies no decision used are
   played in their order, each only if the conversation it answered is the one the turn has then
   (the same request); otherwise the provider is asked. When every recorded reply was used and
   refused, the next attempt asks the provider with the conversation so far (the refused reply and
   the reason), or the seat passes after its fourth. A failed call ends the turn, as it does live:
   the next turn starts afresh.
4. **A save that stopped half way.** A save writes the trace, then the export. A process that stops
   between the two leaves a trace with decisions the export does not have, and perhaps a move. The
   export lists its decisions. The resume keeps those, drops the trace's later decisions, and keeps
   all the calls: a dropped decision's reply is played again (3), and a dropped move of a person is
   made again by its player. The trace's start records the resume: its time, the last event kept,
   and how many decisions were dropped (`resumed`).
5. **The same game, set up the same way.** The trace's start must be the one this server would write
   for the game: the same build, protocol, system prompt, model profiles and context presets, the
   same request, seats, seed and clocks. Otherwise the resume is refused, naming what differs: a
   trace has one start, and all its decisions were made under it. So a resumed game keeps its
   clocks, unlike a loaded one: the move clock starts again for the turn in progress, and the game
   clock counts the time played before the stop (from the start to the export's last save).
6. **Computer players choose as they would have.** A random or greedy player draws from its own
   random numbers, seeded at the start. In a resumed game, each draws again, on the boards it drew
   on, what it drew before the stop, so that it goes on as in a game that did not stop, and the
   game replays (`wag-llm replay`).
7. **G6's game.** `wag-llm replay` plays games whose seats are all models or computer players, so G6
   plays a model seat against a random player, not G2's game (a person against a model). Its test
   is in wag-llm, whose server has no test seam: the stop is a copy of the game's folder taken while
   the provider answers the seat's second attempt (its first reply refused), which is the folder a
   process stopped then leaves.
8. **An existing test changes.** `TestAFailedModelTurnThatCannotBeSavedHoldsTheModel` expected the
   model to be asked twice in a turn whose saves fail: a reply that cannot be played, then a call
   that fails. Under "kept before used", the turn stops after the first reply, which cannot be
   saved: the model is asked once, and again only once that reply is saved and played.
9. **"A reply of another turn" is planted in the game's calls.** A trace with a call for a later
   move than its export's is refused at the resume (G5), so that defect cannot reach the turn
   through a resume. G4b plants the reply in the game's record instead.

The browser lists the data folder's games, each with Resume (the endpoint of 2).

### Cases added

| Case | Setup | Expected |
|---|---|---|
| G3b | As G3, but the recorded call has a reply and an error | The reply is not played; the provider is called again, once |
| G4b | As G4, with a reply recorded for the seat's previous move and used by no decision (planted in the game's calls) | The next model turn calls the provider |
| G5c | G1's folder, resumed by a server whose profile for the seat names another model | The resume is refused, naming the profiles |
| G5d | A game whose model seat's first reply is refused by a rule (recorded in the history with its penalty), and whose turn stops after its second reply | The resumed turn plays the second reply as the second attempt; the provider is not called |
| G5e | As G5d, but the recorded second call answered another conversation (its request edited in the trace) | The recorded second reply is not played; the provider is called once, for the second attempt |
| G8 | G1's game played on in the same server (the kept reply's move made, and saved), then its export put back as it was before that move: the save stopped between the trace and the export | The resume drops the decision the export does not have and says so in the trace's start; the seat plays the call's reply again without asking the provider; the moves are those of the game played on, and the trace has the call once |
| G9 | A game of a model seat against a random player, stopped after a model's reply in a later turn, resumed in a new server and played to its end; and the same game played without a stop | The two games make the same moves |
| G10 | A game with clocks (a move clock; a game clock), stopped and resumed | It keeps both clocks: the move clock starts again, and the game clock has the time played before the stop counted; its trace's start is unchanged |

### Planted defects added

| Defect | Caught by |
|---|---|
| The number of moves alone tells which reply is for this move | G5d |
| A reply that came with an error is played | G3b |
| A recorded reply is played whatever conversation it answered | G5e |
| The trace's decisions the export does not have are kept | G8 |
| The resume does not compare the trace's start with this server's | G5c |
| Computer players start their random numbers again | G9 |
| The clocks are dropped | G10 |

### Checks added before they ran (2026-11-27)

Written while the planted defects were being prepared, before these checks ran:

- G2 also checks that the trace's events are in strictly increasing order, and that the export
  keeps the events recorded before the stop and records the resume.
- G2b: G1's folder renamed (`games/renamed`), then resumed and played on. Expected: the game goes on
  saving into that folder, and makes no other.
- G3 also checks that no decision uses the failed call, and that the new reply's decision is the
  turn's first attempt; G4b, that no decision uses the planted reply.

| Defect | Caught by |
|---|---|
| A resumed game saves into a folder named anew | G2b |
| The export's events are not carried over | G2 |
| The order of events starts again | G2 |

## Results (recorded after the tests ran, 2026-11-27)

- **Every case passed as written**, with the checks added before they ran: G1, G2, G2b, G3, G3b, G4,
  G4b, G5 (another game, other data, another move, another reply), G5b, G5c, G5d, G5e, G6, G7, G8,
  G9 and G10 (`server/journal_test.go`; G6 in `wag-llm/resume_replay_test.go`).
- **The existing test changed as amendment 8 said**: `TestAFailedModelTurnThatCannotBeSavedHoldsTheModel`
  asks the model once while the reply cannot be saved, and now also checks that once it is saved the
  kept reply is played (`recovered`) before the model is asked again.
- **Also tested:** the resume takes only a game's folder of the data folder, from this computer
  only, and not a game open in the server (`TestResumeTakesOnlyAGameFolderFromThisComputer`). The
  browser was checked by hand: a server restarted on the data folder lists the game under *Saved in
  your data folder*, and *Resume* opens it with its move and record.
- **Corrected after the planted defects ran** (each correction made to a test, then all the
  defects run again):
  - "The call is saved after the move" was not caught by G1 at first (G7 and the failed-turn test
    caught it): the test's stop ended the turn, but the server then saved its record of the stopped
    turn, which a stopped process would not. The stop now copies the data folder at that moment,
    and the tests resume from the copy, the folder a process stopped then leaves. G1 then catches
    the defect, with G2, G2b, G5b to G5d, G8 and G10.
  - "The trace is not read on load" was caught partly by a crash (G4b took a call the defect had
    removed); the tests now fail on their checks there.
  - "The export's events are not carried over" then survived: the copy is taken before any event is
    recorded, so there was nothing to carry. G2 now has a note posted before the stop.
- The trace reader takes only a trace the server writes: its format, a game's start first and once,
  the four kinds of event (`TestReadTraceTakesOnlyATrace`).
- The planted defects and their runs are in the workspace (`verify/g`).

## Corrections after the independent check of 90d2327 (2026-11-27), before their tests ran

The check (in the workspace, `verify/g/CHECK-g.md`) found 3 errors, 7 warnings and 15 notes. The
corrections, each with its case, written before the case ran:

1. **A held move ends the turn in progress** (E1). A model seat's reply held for the moderator is not
   a refused reply: when the moderator rejects it, the seat's next turn starts afresh, with the
   board and the moderator's note, as before G, and the moderator's rejections are not the seat's
   attempts.
2. **A held move is held again after a resume** (E2). The trace's last decision, held for the
   moderator when the export was saved, is held again: the model is not asked, and a computer player
   does not choose again.
3. **The game clock counts the time played, across resumes** (E3). Each save records the time played
   so far (`played_ns` in the export); a resume counts that, not the time since the game was created.
4. **A move clock does not run while the game cannot be saved** (W3). A model seat waits while the
   game cannot be saved; when a save succeeds, its move clock starts again.
5. **No retry is sent while the game cannot be saved** (W4). A call's tries stop as soon as one's
   record cannot be saved.
6. **A game over by its clock stays over** (W5). A resume applies the game clock: a game whose time
   was up is over at once. A model seat is not asked once the game's time is up.
7. **The build is recorded, not compared** (W6). A trace's start keeps the build that began the game,
   and each resume records its own (`resumed[].build`); the start's other fields, its pins included,
   must still be this server's. The protocol, the system prompt, the model profiles, the context
   presets and the engine's and data's pins say what a build would change in play; `go run`
   records no commit, and an untracked file changes a built binary's.
8. **What a resume takes** (N1, N2, N3). A game's id must be one this server makes (digits); the
   game's folder and its two files must be a folder and files, not links. A recorded reply for the
   first attempt of a turn on the current board (no move since its call) must have answered this
   board: its prompt must be the one the board gives now.
9. **A recorded failed call replays** (N7): `wag-llm replay` goes on past a failed call it replays,
   as the game did.

### Cases added

| Case | Setup | Expected |
|---|---|---|
| G11 | A moderated debate; the model's legal reply held; the moderator rejects it with a note | The model's next request is one message, the board, which names the note; after four rejections the model is still asked |
| G12 | A moderated debate stopped while the model's reply is held, resumed | The move is held again; `/ai` asks nothing (a move awaits the moderator); the moderator's approval plays it |
| G12b | The same with a random player's held move | The same move is held again; after the moderator's rejection, the random player chooses as in the game that did not stop |
| G13 | A game with a game clock, resumed twice after stops of a few seconds | The time played counted is the time played (under a second), not the stops |
| G10b | G10 with the export's time played set to five minutes | The resumed game clock counts five minutes |
| G14 | A move clock of one second; the save fails after the provider's answer; it succeeds two seconds later | The kept reply is played as the move, not voided |
| G15 | Retries allowed; the first try fails with a server error and its save fails | No second try is sent; the provider is called once |
| G16 | A game over by its game clock, resumed | Over at once; `/ai` asks nothing. In a live game whose time is up, `/ai` asks nothing either |
| G17 | G1's folder with its start's build edited (another commit), resumed | Resumed; the start keeps the edited build; the resume records this server's |
| G5f | G1's folder with a game id `../x`, or as a link to a folder elsewhere | Refused, each |
| G5g | G4's folder with: the export's move edited (another move list); a call's turn after the export's moves; the start's request edited; the export's seed edited | Refused, each, naming it |
| G5h | A complete game folder copied under `annotation/` | Refused: not a game's folder |
| G4c | A reply planted for the current turn whose prompt is another board's | Not played: the provider is asked |
| G3c | The model's first reply refused, its second call failed | The next turn starts afresh: a one-message request, attempt 1 |
| G5i | G5d's folder with the second call removed (every recorded reply used and refused) | The provider is asked once, for the second attempt, with the conversation so far |
| G6b | G3's game (a failed call, then the model asked again), replayed | It replays |
| Gn | After resuming game 1 in a new server, a new game | Gets id 2 |

G5d also checks the recovered decision's prompt (the board as the turn began) and feedback; the
redraw has a unit test (a pass is not drawn again; the last board is).
