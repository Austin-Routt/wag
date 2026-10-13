# WAG in the browser: running and checking the prototype

The prototype runs as a small local server (`wag-server`) that the browser talks to:

- **The server** holds the game engine. That is the same Go code as the CLI and the tests: NASCQ validation, Carneades labelling, modes and scoring.
- **The page** shows the game and sends the moves.

Nothing leaves your machine. By default the server listens on `127.0.0.1` only.

## 1. Run it on Windows

**A. The ready-made program (no installs).**
1. Put `wag-server.exe` in any folder and double-click it.
2. A console window says `WAG is running at http://127.0.0.1:8080/`, and your browser opens on that page.
3. Close the window, or press Ctrl+C in it, to stop.

Windows SmartScreen may warn about an unsigned program. Choose *More info → Run anyway*.

**B. From the repository (needs Go 1.24 or later from <https://go.dev/dl/>).**
- WAG builds against `..\carneades-4`, so check out the whole repository, not only `wag\`.
- Double-click `wag\run-wag-server.bat`, or run `wag\run-wag-server.ps1` in PowerShell. Either one runs `go run ./cmd/wag-server`.
- To build your own `.exe`, run `go build -o wag-server.exe ./cmd/wag-server` in `wag\`.

**Options.** Flags pass through the launchers as well.

| Flag | Default | What it does |
|---|---|---|
| `--addr` | `127.0.0.1:8080` | Where to listen. For another port: `--addr 127.0.0.1:9090`. Use `--addr 0.0.0.0:8080` only on a network you trust: seats have no passwords. |
| `--kialo-dir` | none | The visualizer's `data` folder, for example `--kialo-dir C:\kialo_visualizer_v10_release\data`. Its discussions are then listed for import. Without it, you can still upload a single `graph.json`. |
| `--data` | built-in | A folder with `schemes\core.yml`, `nascq\*.json` and `contexts.json`. Use it to try edited scheme data without rebuilding. |
| `--claude` | off | Offer **Claude (API)** seats. Needs `ANTHROPIC_API_KEY` set (or an `ant auth login` profile); every Claude move costs API credits. `--claude-model` (default `claude-opus-5-5`) and `--claude-effort` (default `medium`) tune it. |
| `--open=false` | open | Do not open the browser. |

On Windows, to use Claude: `set ANTHROPIC_API_KEY=sk-ant-...` in a command prompt, then `wag-server.exe --claude`.

Several browser tabs can show the same game, for example one per player and one for the moderator. Each tab updates within about 1.5 s.

## 2. What to check

Each step says what you should see. If what you see differs, press **⚑ Note** on the move, say what you expected, and send me the *Export*.

### 2.1 A debate

1. **Set up the game.**
   - On *New game*, pick **debate** and a **Context** (start with *Everyday discussion*).
   - Make the first seat *Human* (you argue for the claim) and the second one of:
     - *Language model (copy and paste)*: any chatbot;
     - *Claude (API)*: when the server runs with `--claude`;
     - *Rule bot: greedy*: legal moves, no understanding.
   - Click *Start the game*. The page opens on the **Dialogue**; the panel on the right is set to **New argument**.
2. **Say it in your own words**, for example: *I should ask her out: I am single, and dating could lead to a relationship, which would make me happy.*
3. **Pick the scheme.**
   - *Which scheme fits?* suggests schemes from cue words. It is a keyword heuristic; the choice is yours.
   - The scheme card shows its sentences with the parts as chips, what it assumes, and its exceptions ("unless …").
4. **Fill in its parts.** Each field shows an example, such as "e.g. introduce a carbon tax". The preview shows the argument exactly as the board will read it: "You argue that C: …, because 1. … 2. …", and what would change.
5. **Put it forward.** Your move appears as card #1:
   - the claim, with its label;
   - your words;
   - **because …**, which unfolds the premises, each with a coloured dot;
   - what the move changed.
6. **The other side answers.** Its card is indented under yours. It names the question it answers ("↳ the Side Effects question about Alice's a1 (an exception)"), the question (**Q**) and its answer (**A**), the scheme behind it, and the effects ("a1 (Alice's) no longer holds").
7. **Respond.** Press **Respond…** on any card. Its argument's critical questions appear as chips; pick one and choose:
   - *Attack: answer no*;
   - *Defend: answer yes*;
   - *Just ask*.

   The panel then offers only the schemes the NASCQ links to that question, plus the thesis-independent ones, and fills the answer in for you.
8. **Question any premise.** Every premise that no critical question covers has the *Premise Acceptability* question ("Is it true that “…”?"), so you can challenge, say, the circumstances of a practical-reasoning argument.
9. **The Board view** (top left) shows one claim at a time with its pros and cons, and the outline map.
10. **When the game ends**, a panel gives the winner and the reason.

### 2.2 A language model or Claude as a player

- **Language model (copy and paste).**
  - When its turn comes, the panel shows the seat's prompt (`docs/protocol.md`) with *Copy the prompt*.
  - Paste it into any chatbot, paste the whole reply back, and press *Play the reply*.
  - If the move is rejected, the next prompt says why. *or make the move yourself* switches to the ordinary panel.
- **Claude (API).**
  - It moves by itself ("Claude is thinking…"). Each move costs API credits.
  - Its replies, including rejected ones, are kept in the export.

### 2.2 Schemes & questions

The second tab shows the NASCQ itself:

- **The schemes** with their variables, semantic types, Carneades formulas and cue words;
- **The critical questions**, with what each is about (premise, exception, conclusion, inference) and its formula;
- **A network diagram.**
  - Coloured lines are ask edges (scheme → question), coloured by locus.
  - Dashed lines are backup edges (question → scheme).

Click a scheme or question to see its links. The data is what `wag/data/nascq/*.json` and `wag/data/schemes/core.yml` contain today: 7 Walton schemes and 20 questions, plus the generic premise question, and "Other" with its generic questions when enabled.

### 2.3 Options

- **Context.** Sets the proof standard of the claim, whether sources are required, and draft scheme weights (`wag/data/contexts.json`): everyday, policy deliberation, scientific inquiry, legal case, equal weights.
- **Any premise can be questioned** (on). It adds the generic premise question; turn it off to play the NASCQ's questions only.
- **Allow the "Other" scheme.** The scheme list gains *Other*, where you write up to six premises and a conclusion yourself.
  - Its questions are generic: is premise *n* acceptable, are the premises relevant, is there a counterargument.
  - *Export* lists every Other argument separately, for argument mining.
- **Require sources.** Expert opinion and position to know cannot be played without at least one source (rule E1). For consequences, sign and analogy, sources are recommended.
- **Moderator.**
  - Every move waits as *Awaiting the moderator*.
  - The moderator is a tab with *I am the moderator* ticked. They approve the move, optionally with their own weight, or reject it with a note.
  - On any statement they can also rule *Accept*, *Reject* or *Clear*. A ruling is a Carneades assumption, which the CAF reserves for statements the audience accepts.

### 2.4 Kialo

- **Import a discussion.** Use *Import a Kialo discussion* with an uploaded `graph.json`, or with `--kialo-dir`.
- **How it is converted.**
  - Every claim becomes a statement.
  - Every pro becomes an *Other* argument for its parent; every con becomes one against it.
  - Arguments are weighted by the mean impact vote.
  - The thesis is the central claim.
- **What to check.**
  - The labels Carneades computes for the imported debate.
  - That every imported argument now has critical questions you can press.
- **Known limit.** The 3,690-argument "Does God exist?" imports in about a second, but its map is long.

### 2.5 Saving and notes

- **⚑ Note** on any move, or **Note** in the top bar for the whole game: wrong, confusing, unreadable, good, an idea. Notes appear on the card and go into the export.
- **Export** downloads the game as JSON (`wag-export/2`). It holds:
  - the build;
  - the settings and seed;
  - every move, and every move the server rejected, with why;
  - your notes;
  - Claude's replies.
- **Load an exported game** (on *New game*) replays an export move by move, so a game can be looked at again, continued, or turned into a test.

## 3. Known limits of the prototype

- **Games live in memory.** Export and load to keep them.
- **No player accounts.** Whoever has the page open can move for the seat whose turn it is, or tick *I am the moderator*. That is fine on one machine, not on a network.
- **Few schemes.** Only 7 Walton schemes are in the NASCQ, and two of them differ from the book (`docs/review-2026-10-13-part4.md`, §5 and §11). Expanding it is the next phase.
- **Rule bots do not understand the topic.** They reuse words already on the board and cite neutral sources. Use a language model or Claude for real play.
- **The English is built from templates.** It reads well when the parts are short phrases like the examples. Long sentences inside parts make long statements.
- **Scheme suggestions are keyword cues.** A classifier, such as an LLM or a trained argument-mining model, would replace `nascq.Suggest` behind the same API.
- **The context weights are drafts.**
- **A Claude move takes a while.** The other games on the server keep working meanwhile; that game waits.

## 4. For developers

`internal/server` serves the page (embedded with `go:embed`) and a JSON API:

| Endpoint | What |
|---|---|
| `GET /api/meta` | Schemes, questions, links, modes |
| `GET /api/suggest?text=` | Scheme suggestions |
| `GET/POST /api/games` | List games, or start one |
| `GET /api/games/{id}?since=v` | Game state (`{"unchanged":true}` if still at version v) |
| `POST /api/games/{id}/move` | A move, `{"action": …}`, where the action is the CLI's `nascq.ProposedAction` JSON plus `text` and `evidence` |
| `GET /api/games/{id}/moves?target=a1` | Legal attacks, defences and questions on an argument |
| `POST /api/games/{id}/preview` | What a move would do, without playing it |
| `POST /api/games/{id}/ai` | Let an AI seat move (a rule bot, or Claude) |
| `GET /api/games/{id}/prompt`, `POST /api/games/{id}/reply` | The text protocol: the seat's prompt, and a reply to play (`docs/protocol.md`) |
| `POST /api/games/{id}/note` | A tester's note, `{"target": "move:3", "kind": "confusing", "text": "…", "by": "…"}` |
| `POST /api/load` | Start a game from an export and replay it |
| `POST /api/games/{id}/moderate`, `/decide` | Moderator actions |
| `GET /api/games/{id}/export` | The game: settings, moves, attempts, notes, LLM replies, Other arguments, build |
| `GET /api/kialo`, `POST /api/kialo/import` | Kialo discussions |

Requests that change a game must be `POST`s with `Content-Type: application/json`. That way other web pages cannot play in a local game. A language-model player uses the text protocol; a program can also use the JSON API like the page, or the CLI's `--p1 json` pipe.
