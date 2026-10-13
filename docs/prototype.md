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

**B. From the repository (needs Go 1.22 or later from <https://go.dev/dl/>).**
- WAG builds against `..\carneades-4`, so check out the whole repository, not only `wag\`.
- Double-click `wag\run-wag-server.bat`, or run `wag\run-wag-server.ps1` in PowerShell. Either one runs `go run ./cmd/wag-server`.
- To build your own `.exe`, run `go build -o wag-server.exe ./cmd/wag-server` in `wag\`.

**Options.** Flags pass through the launchers as well.

| Flag | Default | What it does |
|---|---|---|
| `--addr` | `127.0.0.1:8080` | Where to listen. For another port: `--addr 127.0.0.1:9090`. Use `--addr 0.0.0.0:8080` only on a network you trust: seats have no passwords. |
| `--kialo-dir` | none | The visualizer's `data` folder, for example `--kialo-dir C:\kialo_visualizer_v10_release\data`. Its discussions are then listed for import. Without it, you can still upload a single `graph.json`. |
| `--data` | built-in | A folder with `schemes\core.yml` and `nascq\*.json`. Use it to try edited scheme data without rebuilding. |
| `--open=false` | open | Do not open the browser. |

Several browser tabs can show the same game, for example one per player and one for the moderator. Each tab updates within about 1.5 s.

## 2. What to check

Each step says what you should see. If what you see differs, that is a finding I want to hear about.

### 2.1 A debate against the computer

1. **Set up the game.**
   - On *New game*, pick **debate**.
   - Make the first seat *Human* (Alice, pro) and the second *Greedy AI* (Bob, con).
   - Click *Start*.
   - The board is empty. The composer on the right is set to **Argue**.
2. **Write the argument in your own words**, for example: *We should introduce a carbon tax in order to reduce emissions: it would lower emissions from the status quo, and that promotes environmental health.*
3. **Find the scheme.**
   - Click *Which scheme fits?*. It suggests schemes from cue words such as "in order to" or "should".
   - **This is a keyword heuristic, not semantic understanding.** The choice is yours.
4. **Choose the scheme** from the list of official Walton schemes, for example *Practical Reasoning*.
   - The scheme card shows the scheme's premises and conclusion with their variables (`S1`, `A`, `S2`, `G`, `V`).
   - It also says whether the scheme needs sources.
5. **Break the argument into the scheme's parts.**
   - Example: `S1` = the status quo, `A` = carbon tax, `S2` = lower emissions, `G` = reduced emissions, `V` = environmental health.
   - The preview below shows the statements the engine will create and how the labels would change.
6. **Put it forward.** The claim, *Action carbon tax should be performed*, now stands as the focus.
   - The argument card lists the premises, each with a coloured dot (green In, red Out, grey Undecided).
   - It also shows the scheme's critical questions as chips. These are the **ask edges** of the NASCQ.
7. **Bob answers.**
   - Typically Bob answers a critical question negatively. For example, he answers *Other Means* with an *Argument from Expert Opinion* that another action would do better.
   - The card then shows the exception that now holds, who argued it, and why the argument is out of force.
   - The history records the move and the change in score.
8. **Open the question.**
   - Click a question chip, then *Defend: answer for*, *Attack: answer against* or *Ask*.
   - The composer offers only the schemes the NASCQ links to that question (its **backup edges**), plus the thesis-independent ones (expert opinion, position to know, analogy). Each is marked as such.
   - The answer fills the scheme's answer variable for you. You fill in the rest.
9. **Defend.**
   - For example, deny the exception with an expert who says it is not so.
   - The preview shows the argument coming back into force. After you play it, Bob may attack your defence in turn (for example its premise *the expert is credible*).
   - Disputes go as deep as the players take them, each level through the critical questions of the scheme used there.
10. **Check the result.**
    - The top bar shows the scores and the claim's label.
    - When the game ends, a panel gives the winner and the reason.

### 2.2 Schemes & questions

The second tab shows the NASCQ itself:

- **The schemes** with their variables, semantic types, Carneades formulas and cue words;
- **The critical questions**, with what each is about (premise, exception, conclusion, inference) and its formula;
- **A network diagram.**
  - Coloured lines are ask edges (scheme → question), coloured by locus.
  - Dashed lines are backup edges (question → scheme).

Click a scheme or question to see its links. The data is what `wag/data/nascq/*.json` and `wag/data/schemes/core.yml` contain today: 7 Walton schemes, 20 questions, and "Other" with its 3 generic questions when enabled.

### 2.3 Options

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

## 3. Known limits of the prototype

- **No saving.** Games live in memory and are lost when the server stops. *Export* saves a game as JSON, but there is no import of saved games yet.
- **No player accounts.** Whoever has the page open can move for the seat whose turn it is, or tick *I am the moderator*. That is fine on one machine, not on a network.
- **Few schemes.** Only 7 Walton schemes are in the NASCQ. Expanding it is the next phase (`docs/walton-audit/REPORT.md`).
- **The English is faithful but clumsy in places.**
  - Templates read awkwardly when a variable is a noun phrase: "Performing carbon tax" where "introducing a carbon tax" was meant.
  - Answers about answers stack negations: "It is not the case that “it is not the case that …” is in the domain of …".
  - Improving the templates is part of the scheme cleanup.
- **Scheme suggestions are keyword cues.** A classifier, such as an LLM or a trained argument-mining model, would replace `nascq.Suggest` behind the same API.
- **The map is an outline of claims and arguments, not a drawn graph.** The NASCQ tab draws the scheme network; argument graphs are not drawn yet.

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
| `POST /api/games/{id}/ai` | Let an AI seat move |
| `POST /api/games/{id}/moderate`, `/decide` | Moderator actions |
| `GET /api/games/{id}/export` | The game, its moves and its Other arguments |
| `GET /api/kialo`, `POST /api/kialo/import` | Kialo discussions |

Requests that change a game must be `POST`s with `Content-Type: application/json`. That way other web pages cannot play in a local game. An LLM player can use the same API as the page, or the CLI's `--p1 json` pipe.
