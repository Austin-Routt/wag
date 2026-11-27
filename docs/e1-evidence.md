# E1: evidence records in the game (the evidence plan's step 1), version 3

Stage 2 ends when, among other things, evidence records are bound to the actual game, old
citation-only traces keep their meaning, and the specific CQ, admission and revision effects are
tested (plan.md §0, "Evidence increment"). This is step 1 of the
[evidence implementation plan](evidence-implementation-plan.md#step-1-connect-the-first-slice-before-an-evidence-aware-game-pilot)
("EIP" below): connect the first slice before an evidence-aware game pilot. This page plans it: what
stands, the choices, the steps, and each step's cases with their outcomes, written before the tests
run. "E1" names this step; the game's citation rule, whose refusals begin `E1:`, is called the
citation rule here.

Version 1 was checked independently (in the workspace, `verify/e1/CHECK-plan-v1.md`: 5 errors, 15
warnings, 14 notes); version 2, which answered it, was checked too (`CHECK-plan-v2.md`: 4 errors, 14
warnings, 13 notes). Version 3 answers both; [the last section](#what-versions-2-and-3-change) says
how, finding by finding.

## What stands (2026-11-27)

- **A move cites sources as strings** (`evidence: [string]`), kept as the argument's metadata and
  shown. With *Require sources* on, an argument from a scheme whose evidence is "required" (expert
  opinion, position to know) is refused without one (the citation rule, `turn.go`).
- **`wag-evidence/0.1` checks records offline** (`internal/evidence`, `wag-evidence`): sources,
  artifacts with their capture status, retained representations with their bytes' hash, attributed
  claims with exact selectors, targets, bindings, assessments, dependencies, and an ordered log of
  two participants' consents and withdrawals for an exact binding revision
  (`two-party-consideration/1`). It reads no live graph: a target and the context's pins are
  checked for form only (evidence-contract.md).
- **E1.1 is done** (2360463): a game records its inputs by content in its trace's start and its
  export: the engine (the non-test Go files and `go.mod` of both vendored modules, carneades-4 and
  GoCHR, read with LF line endings), the scheme file and the NASCQ data (each read with LF), tested by
  `internal/manifest/manifest_test.go` and `internal/server/pins_test.go`, with a GoCHR byte and CRLF
  copies.
- **No mode is cooperative.** Every mode has a winner, or in solo none (`modes.go`).
- **The owner's decisions that bound this:** D076 (a source is a record; consent is to the exact
  binding; admission is for consideration only: not a stipulation, a moderator's ruling, the
  engine's acceptance or a finding of fact; withdrawing a consent removes only its own route; the
  four game profiles of the review, cooperative, adversarial, moderated and automated, wait for
  Track D, and the first slice is two players under the two-party policy), evidence-records.md
  (lines 69-71: "Two players, cooperative"; lines 145-146: the other profiles wait for Track D), D077
  (capture for research, kept private), D079 (terms of use and injected instructions: open, no gate
  before the first pilot), D045 (a person verifies).

## The choices

Each is the developer's, under the owner's instruction to go with the recommended choices; those
marked **(owner)** set a game rule or read a decision, and are put to the owner with this plan, the
work going on with the recommendation meanwhile, as before (D094, D095).

1. **The slice is a cooperative game of two players (owner).** D076 puts the review's four profiles
   in Track D and fixes the first slice as two players under the two-party policy;
   evidence-records.md reads that slice as cooperative. *The automated profile* is the review's
   "Automated/mixed: deterministic checks and model reports supply inspectable assessments; policy
   determines what authority, if any, they receive" (reviews/2026-11-22, the convergence review's
   RESPONSE.md, its profile table): automatic assessments, which no step here makes, not players
   that are models. A new mode, `inquiry`, is the cooperative game:
   - Two players examine a claim together. Neither has a side; no one wins or loses; the result is
     the verdict, as in solo; scores are shown and decide nothing.
   - It ends when both players pass in the same round, at the round limit (20), when the game clock
     runs out, or when a player resigns: a resignation ends the inquiry, with the verdict and no
     loser.
   - It refuses the settings that make a winner or a loser: a target score, an answer deadline
     (`answer_within`), conceding on a pass, and losing on a late move; it takes the clocks (a late
     move passes the turn) and the other settings.
   - A model seat's prompt and system prompt in an inquiry say that the players examine the claim
     together, with no side and no winner. The system prompt of other games does not change: each
     game's start records the one it was played with, which a resume and a replay compare.
   - The board names no side and no winner in an inquiry; the terminal (`wag-cli`) plays one too.
     An inquiry is two players: the game refuses another number, at the server and the terminal.
   - An inquiry without a record set may have rule bots; one with a record set may not (choice 2).
2. **The record set comes with the game, from an evidence folder.** Sources, artifacts,
   representations, claims and dependencies are captured before play (D077), by the owner's tools.
   - **Where it lives.** A record set named `NAME` is the file `NAME/records.json` of an evidence
     folder, with its payloads under `NAME/` (payload paths are relative to it). `wag-server`'s
     evidence folder is its data folder's `evidence/`; `wag-llm play` and `wag-llm replay` take one
     with `--evidence DIR`. A request names the record set (`"evidence": "NAME"`); the trace's start
     and the export record the name and the SHA-256 of `records.json`'s bytes. `NAME` is a plain
     name (no separator, no `..`); the server refuses an evidence folder, or a `NAME` folder, that
     is a link, and wag-evidence's rule refuses a link below it.
   - **What it holds.** A `wag-evidence/0.2` bundle (choice 6) without the keys `context`,
     `targets`, `bindings`, `assessments` and `events` (the game makes those), decoded strictly
     (`evidence.Decode`: no unknown field, duplicate key or null). Its actors are its records'
     authors (the capture tools, the people who captured), and every record's author is one of
     them; no id, its actors' included, begins with `game:`, which the game's records and actors
     use. Artifacts of any capture status may be in it (evidence-records.md keeps partial and failed
     records as research records); a claim's binding shows its artifact's status.
   - **When it is checked.** In full (`evidence.Validate`: records, payload bytes, selectors), with
     the game's context and actors (choice 3), when a game is created, loaded or resumed and when a
     trace is replayed; a record set that fails is refused, naming the first error. During play an
     act is checked with `evidence.ValidateRecords`, which applies the same record and consent rules
     without reading a payload (a payload changed during a game is found at the next load). No
     payload bytes are served during play in this slice. `wag-evidence --record-set NAME/records.json`
     checks a record set alone, with a placeholder context of two players.
   - **Which games (owner, reading D076).** An inquiry of two seats, each a person (`human`), a model
     through its API (`model`, `claude`) or a person relaying a chatbot (`llm`, a model participant:
     the person relays, the chatbot answers); not moderated; on a server with an evidence folder. A
     rule bot (`random`, `greedy`) is not a player in D076's sense: it cannot consent. Otherwise the
     game is not created, saying why.
3. **The game writes the context and the actors.** The context's id is `game:context`, its profile
   `wag-inquiry/1`, its policy `two-party-consideration/1`, its participants the two players' actor
   ids, and its pins the game's own (E1.1). Each player is an actor: id `game:player:` and the
   player's id, kind `human` for a person and `model` for a model seat (`model`, `claude`, `llm`),
   label the seat's name. The game itself is a `tool` actor, `game:wag`, which authors the targets;
   a binding is authored by the player who offers it.
4. **Evidence acts are moves, and take the turn (owner).** Recommended because a model seat's reply
   is one move, and a turn's record (G) already handles moves; the alternative, acts off the turn as
   the moderator's decisions are, would let a seat make several acts and a move in one turn.
   - **The acts and their fields.** An offer (`"type": "evidence_offer"`, `"offer": {"claim_id",
     "argument_id", "kind", "statement", "role", "interpretation", "warrant"}`); a consent
     (`"evidence_consent"`, `"binding_id"`, `"revision"`); a withdrawal (`"evidence_withdraw"`,
     `"binding_id"`, `"revision"`: the revision consented to). A revision is given in full or as its
     first 12 hexadecimal digits, which must be those of the named binding's revision; the act
     records the full one.
   - **The rules of the turn.** An act counts as a move, not a pass: an inquiry does not end on a
     round of acts. A refused act is refused as any move is, with the usual penalty, which decides
     nothing in an inquiry, and leaves no record. Question deadlines and the move clock apply to an
     act as to any move.
   - **An offer is its player's consent (owner):** so that the other player's consent admits it; a
     player who offers material for consideration is taken to put it forward.
   - Silence is no consent, and there is no act to decline (choice 14 says what can be argued).
   - **The history's mark.** Each accepted act is recorded in the move history with a mark of its
     own (as the moderator's `decide` has): its act, binding, binding revision, an offer's argument
     revision, and its time. Replay, the board's log and the move's description read an accepted
     act as accepted, though it makes no argument. The trace and the export record it.
5. **The game makes the target from the live graph.** An offer names an argument a support or
   attack move made (not a question's, a presumption or an exception argument) and one of its
   statements by kind: a premise or an assumption (its scheme's premises, then its assumptions), an
   exception, its conclusion, or its inference, whose statement is the argument's undercutter
   statement (`undercut(A)`, which an undercutter of it concludes). Kinds `source` and `annotation`
   are not offered in this slice. The game writes the target, authored by `game:wag`: the
   argument's id and revision, the kind, and the statement's id as its proposition.
   - **The argument's revision** is `wag-argument-revision/1`: the SHA-256 of the JSON encoding of
     the algorithm's name, the move's NASCQ scheme id, the argument's premises (statement ids, in
     order), its conclusion, its exceptions (sorted) and the move's weight (the assigned weight, not
     the evaluated weight, which an undercutter sets to 0). A golden test pins it. In play an
     argument never changes once made, so an admission cannot go stale through the graph.
6. **Revisions depend only on recorded inputs.** A load, a resume and `wag-llm replay` rebuild the
   acts from the history and must find the revisions recorded, or a recorded consent would be
   refused as stale and a model asked (and paid) again.
   - A new contract version, `wag-evidence/0.2`, computes a binding's revision as
     `wag-binding-go-json/3`: as `/2`, without the records' `created_at`. A time says when a record
     was written, not what is consented to, and D076's list of what a consent is to has no time.
     `wag-evidence/0.1` bundles keep `/2` and their meaning.
   - The game's records take their ids from the history: `game:t<i>`, `game:b<i>` and `game:e<i>`
     for the target, the binding and the event made by the move at index `i`. Their time is the
     act's, recorded with its mark and reused when the act is replayed, so that an export dates each
     act when it was made.
   - Replaying an act checks that the revisions it makes or names are the recorded ones.
7. **A move may cite an admitted binding (owner).** A move may cite bindings (`evidence_refs`,
   besides the citation strings, which keep their meaning: a binding id or a URL written among the
   strings is a string). A cited binding must be admitted for consideration when the move is made,
   and must agree with what the move's argument concludes:

   | The move's argument concludes | A cited binding must be |
   |---|---|
   | S (a support of S, by target statement or as an answer; an answer at an exception, which concludes the exception S) | `support`, its target's proposition S |
   | the opposite of S (a rebuttal of an argument for S, an undermining of premise S, a defence at exception S, an attack on S by target statement) | `oppose`, its target's proposition S |
   | `undercut(A)` (a direct undercutter of A) | `oppose`, at A's inference |
   | the opposite of `undercut(A)` (a defence of A's inference) | `support`, at A's inference |
   | nothing (a question, a pass, a resignation, an evidence act) | none: they cite nothing |
   | a statement not yet on the board (the opening claim, a new statement) | none: no binding can target it; it cites strings |

   - The match is at the statement (owner): a binding offered for premise S of argument A grounds
     any argument for S, since the evidence is about S. The stricter reading of EIP item 3 (only a
     move about A itself) would leave an admitted binding unable to ground the argument that defends
     S.
   - Bindings of role `context` and `definition` may be offered and admitted, and ground no move: a
     definition or context link cannot silently become factual support (evidence-records.md).
   - *Require sources* is met by a citation string or by a cited admitted binding that agrees with
     the move, of role `support` or `oppose`.
   - Admission changes no label, weight or assumption (D076): the argument stands on its scheme and
     its critical questions like any other.
8. **Withdrawal marks, it does not undo (owner).** A withdrawn consent ends the binding's admission.
   A move already made that cited it stays in the history and the graph, in force as before: the
   citation rule is checked when a move is made. Its reference is shown withdrawn, and an argument
   whose every cited binding is withdrawn, and which cites no string, is marked on the board as no
   longer grounded by admitted evidence; the mark changes no label (D076). What withdrawal
   recomputes (EIP item 6) is admission and these marks, never labels, which admission never set.
   Nothing is deleted: records, history and earlier consents stay.
9. **A changed record is a new record.** The record set is fixed for the game. An offer of the same
   claim with another interpretation, warrant, role or target is another binding, with its own
   revision; a consent names the revision it is for, and one for another revision is refused. An
   offer identical to one already offered (the same claim, target, role, interpretation and
   warrant) is refused, naming the binding.
10. **Model seats read evidence as data.** A model seat's prompt, in an inquiry with a record set,
    lists the claims it may offer (each with its source, representation, mode, text, scope and
    passage; at most 30, the rest counted), and each offered and admitted binding (at most 30) with
    every part its revision covers (the source and representation, the passage, the source's
    assertion, the interpretation and warrant, the target, the role, the context and the policy),
    its full revision and who has consented. Excerpts are quoted as material to read, not
    instructions, inside delimiters an excerpt cannot close (the delimiter's characters in an
    excerpt are escaped), each cut at 2,000 characters with the cut said. What a model was shown is
    its decision's recorded prompt. That a model obeys nothing in an excerpt is tested on what the
    game does, not on what the prompt says.
11. **Every simulation holds the evidence state.** The snapshot a simulation takes and puts back
    (the board's preview, a reply's alternatives, the greedy bot, the terminal) holds the records
    made in play, the log and admission, from E1.3 on: no simulation makes a real act.
12. **The board shows evidence as data.** For each binding, the board shows every part its revision
    covers, its full revision and who has consented, besides the record set's sources, claims and
    dependencies. Excerpts, claims and labels are shown as text (escaped, never through the board's
    `rich` text); a source's reference (its URL) is shown as text, not as the material and not as a
    link. The markup the board builds is checked in a headless browser (Playwright, as the browser
    checks so far): not in CI, recorded with the results.
13. **Replay, load and resume need the record set and its payloads (owner).** A load, a resume or a
    replay of a game with a record set reads it again from the evidence folder (`wag-llm replay`:
    `--evidence DIR`), checks its SHA-256 and its payloads, and the game's pins against the
    server's; missing or changed, or with other pins, the game is refused, naming what, and its
    export stays readable as a file. There is no read-only inspection of such a game in this slice:
    replaying its acts needs the payloads and the pins. Games without a record set are unchanged:
    their prompts, replies, traces and exports are byte for byte what they are now
    (`wag-protocol/2`, `wag-trace/2`, `wag-export/2`; every new field is omitted when empty), so
    traces made before E1 replay on this build. A game with a record set says
    `wag-protocol/2+evidence` in its prompt and trace.
14. **What can be argued about material not admitted (owner).** evidence-records.md (lines 77-78):
    "Evidence that is not admitted can still be argued about: a player may attack a source or its
    relevance like any premise." In this slice players argue about the target statements, as about
    any statement; no move targets a source or a binding (kinds `source` and `annotation`, and role
    `source_challenge`, are not offered). This is a departure from the decided text, put to the
    owner, with the move that would meet it (an offer of role `source_challenge` at a target of kind
    `source`) left to the next step.
15. **Privacy.** A game with a record set carries captured excerpts in its trace, export and
    prompts: private data (D077), kept in the data folder, never in the public repository; the
    fixture is synthetic. A model seat sends excerpts to its provider, which D079 leaves open,
    without a gate before the first pilot.

## The steps

| Step | What | Done when |
|---|---|---|
| E1.0 | The cooperative game: mode `inquiry` in the game, the server and the terminal; its prompt, system prompt and board | C1 to C9 pass |
| E1.1 | The pins (done, 2360463) | Passed |
| E1.2 | The contract's `wag-evidence/0.2` and `ValidateRecords`; the record set: the evidence folder, the request, strict decoding, the checks, the game's context and actors; `wag-evidence --record-set`; the board's evidence panel | V1 to V4, R1 to R16 |
| E1.3 | Offers, consents and withdrawals as moves: validation, targets from the live graph, the argument revision, the log and admission; the history's mark with the act's time; the snapshot; the trace, the export and the move digest (evidence acts rendered only when present) | O1 to O19 |
| E1.4 | Moves that cite bindings: `evidence_refs`, the matching table, *Require sources*, withdrawal's marks | M1 to M14 |
| E1.5 | Model seats: the prompt's evidence section, the reply protocol's acts | S1 to S6 |
| E1.6 | Replay, load and resume with a record set; old games unchanged | Y1 to Y10 |
| E1.7 | A planted defect for every rule below; the docs (protocol.md, game-rules.md, evidence-records.md, evidence-contract.md, reply.schema.json, trace.schema.json, llm.md's `--evidence`, g-journal.md, the plan); gates on both toolchains; an independent review | Each defect caught |

## Cases (outcomes written before the tests)

The fixture: an inquiry on the shipped data, players A (first) and B, with a synthetic record set
`fixture` in a temporary evidence folder: two sources (an agency's report, a news article quoting
it), each with an original text representation, the report with an excerpt derivative, a
common-origin dependency between the two artifacts, and three claims (a quotation from the report's
excerpt, a paraphrase in the article, an observation). A's opening argument `a1` is AS:consequences
(positive consequences) for "tax meat", declaring `tax meat` with its actor (D094 (b)); its premises
are P = `brings_about(tax_meat,lower_emissions)` and V = `good(lower_emissions)`, its exceptions
`bad_outweighs`, `gain_discounted_by_bearer` and `too_slight`, and its claim
`should_be_performed(tax_meat)`.

### E1.0 The cooperative game

| Case | Setup | Expected |
|---|---|---|
| C1 | An inquiry of two seats | Neither player has a side; the outcome has the verdict and no winner, before and after the end |
| C2 | An inquiry of one seat, or three, at the server and at the terminal | Refused: an inquiry is two players |
| C3 | A argues and B passes; then both pass in the next round | Over; the result is the claim's label, no winner |
| C4 | A limit of one round; A argues and B attacks | Over at the limit; the verdict as result, no winner |
| C5 | A argues; B resigns | Over; the verdict as result; no winner and no loser |
| C6 | An inquiry requested with a target score, an answer deadline, conceding on a pass, or losing on a late move | Refused, each, naming it |
| C7 | A model seat's turn in an inquiry | Its prompt and system prompt name no side and no winner, and say the players examine the claim together |
| C8 | A model seat's turn in a debate, after E1 | Its system prompt is the one before E1, byte for byte |
| C9 | An inquiry with a rule bot and no record set | Created; played to its end |

### E1.2 The contract and the record set

| Case | Setup | Expected |
|---|---|---|
| V1 | A `wag-evidence/0.2` bundle; each record's time changed | The binding's revision unchanged; a 0.1 bundle's changes, as before |
| V2 | The same; the interpretation, the warrant, the role, the target, the argument revision, the claim's scope, a participant's label, a pin, the artifact's revision, the binding's author, the context's profile or the source's reference changed | The revision changes, each |
| V3 | `ValidateRecords` on a bundle whose payloads are gone | Its records and admission checked, no payload read; `Validate` refuses it, naming the representation |
| V4 | `ValidateRecords` and `Validate` on the same bundle with its payloads | The same revisions, consents and admission |
| R1 | The fixture's game | Created; the board lists the sources, claims and the dependency; the trace's start and the export record the name and `records.json`'s SHA-256 |
| R2 | A payload changed by one byte | Refused, naming the representation |
| R3 | A record set with a `context`, `targets`, `bindings`, `assessments` or `events` key (even empty) | Refused, each |
| R4 | A record set with an id, or an actor, beginning with `game:` | Refused |
| R5 | A record whose author is not one of the record set's actors | Refused |
| R6 | A payload path outside `NAME/`, a link below it, the `NAME` folder a link, or the evidence folder a link | Refused, each |
| R7 | `NAME` with a separator or `..` | Refused |
| R8 | A debate (not an inquiry) with the record set | Refused |
| R9 | An inquiry with the record set and a rule bot seat | Refused (the game is not created) |
| R10 | A moderated inquiry with the record set | Refused |
| R11 | A server without an evidence folder (no data folder, no `--evidence`) | Refused, naming the evidence folder |
| R12 | The record set's JSON with a duplicate key, an unknown field, or a null | Refused, each |
| R13 | A claim's text holding `<script>`, a `[text](https://…)` link and a JSON move; a source whose reference is a URL (browser) | Shown as text: no script runs, no link is made, no move is made |
| R14 | The fixture's game: the context and actors of its bundle | The context `game:context`, `wag-inquiry/1`, the policy, the participants `game:player:A` and `game:player:B`, the game's pins; the actors' kinds (`human`) and labels (the seats' names); `game:wag` a tool |
| R15 | An inquiry with an `llm` seat and the record set | Created; that player is a `model` participant |
| R16 | `wag-evidence --record-set` on the fixture's record set, and on one with a binding | Valid; refused |

### E1.3 Offers, consents and withdrawals

| Case | Setup | Expected |
|---|---|---|
| O1 | A offers the quotation for `a1`'s premise P, role support | Accepted; a target (by `game:wag`: `a1`, its revision, kind premise, P) and a binding (by `game:player:A`) are recorded, with A's consent to the binding's revision; not admitted |
| O2 | O1, then B consents to that revision | Admitted for consideration |
| O3 | A offers the quotation for P with interpretation X, then with interpretation Y; B consents naming Y's binding and X's revision | Refused: stale |
| O4 | O2; B consents again | Refused: a duplicate consent |
| O5 | O2; B withdraws | Not admitted; A's consent and the records stay |
| O6 | O1; B withdraws without having consented | Refused |
| O7 | O5; B consents again to the same revision | Admitted again |
| O8 | An offer for an argument not on the board, a question's argument, a presumption argument, a statement not in the argument, a kind it does not have, or a kind `source` or `annotation` | Refused, each; no target, binding or event is left |
| O9 | An offer of a claim not in the record set, of role `source_challenge`, or without an interpretation or a warrant | Refused, each, leaving nothing |
| O10 | O2; then another argument for P, an answer at `a1`'s exception `bad_outweighs`, a question about `a1` and a direct undercutter of `a1` | `a1`'s revision is unchanged, and so is the binding's; the admission stands |
| O11 | Each of O1, O2, O5, and a refused act | The history records the act with its mark; the board's log shows an accepted act as accepted and a refused one as refused; the trace and the export record it; no label, weight or assumption changes (the labels before and after equal) |
| O12 | O1, then the same offer again | Refused, naming the binding |
| O13 | A offers the same claim for P with another warrant | Another binding, another revision; B's consent to the first does not admit it |
| O14 | A round in which both players make acts only | Not over: an act is not a pass |
| O15 | B consents by the revision's first 12 digits; then by 12 digits of another binding's revision | Admitted, the full revision recorded; refused |
| O16 | A refused act | The usual penalty; no record |
| O17 | The argument revision of `a1` | Equal to the golden value: `wag-argument-revision/1` over AS:consequences, P and V in order, `should_be_performed(tax_meat)`, the three exceptions sorted, 0.65 |
| O18 | The board's preview of an act, and an act among a reply's alternatives | The evidence state after as before |
| O19 | An offer at `a1`'s inference | Its target's proposition is `undercut(a1)` |

### E1.4 Moves that cite bindings

| Case | Setup | Expected |
|---|---|---|
| M1 | O2 (admitted, support, P); B argues for P with expert opinion, citing the binding | Accepted; the argument records the reference |
| M2 | O1 (offered, not admitted); the same move | Refused: not admitted |
| M3 | O2; an undermining of P citing the binding | Refused: the role does not agree |
| M4 | O2; a move for another statement citing it | Refused: the target does not agree |
| M5 | *Require sources* on; expert opinion for P citing neither a string nor a binding | Refused by the citation rule; with the admitted support binding, accepted; with an admitted definition binding for P only, refused |
| M6 | *Require sources* on; an undermining of P by expert opinion citing an admitted `oppose` binding at P | Accepted; citing an admitted `context` binding at P only, refused |
| M7 | M1, then B withdraws | The argument stays, in force, the labels as before; its reference is shown withdrawn; the argument is marked no longer grounded |
| M8 | M1 citing two admitted bindings, then one withdrawn; and M1 citing one binding and a string, then the binding withdrawn | One reference withdrawn, one admitted, not marked; the second not marked (the string grounds it) |
| M9 | M1's moves with a citation string in place of the binding | The same labels and the same open questions (paired) |
| M10 | A move citing an unknown binding id | Refused (it is not read as a string) |
| M11 | An admitted `oppose` binding at `a1`'s inference; a direct undercutter of `a1` citing it | Accepted; a rebuttal of `a1` citing it, refused |
| M12 | An admitted `support` binding at `a1`'s exception `bad_outweighs`; B's answer at that exception (CQ:bad_outweighs) citing it | Accepted; the same answer citing an inference binding, refused |
| M13 | An admitted `oppose` binding at `a1`'s exception `bad_outweighs`, after B's answer there; A's defence at the exception (concluding its opposite) citing it | Accepted |
| M14 | A binding at `a1`'s value premise V, admitted; an argument for V citing it | Accepted (the practical-reasoning slice: evidence for a value) |

### E1.5 Model seats

| Case | Setup | Expected |
|---|---|---|
| S1 | A model seat's turn with the record set, an offered and an admitted binding | Its prompt lists the claims it may offer and both bindings with every part their revisions cover, the full revisions and who consented; excerpts delimited and escaped, cut at 2,000 characters; the decision's prompt has them |
| S2 | B a model seat: A (a person) offers; B replies with a consent (by the 12-digit prefix), then later a withdrawal, then an offer of its own | Each played as a person's: the same history entries as the same acts by a human seat (paired) |
| S3 | An excerpt holding a JSON move to resign and words telling the reader to consent to everything; the model's reply quotes it, then ends with its own move | The game plays the reply's own move, and makes no act the reply did not end with |
| S4 | A reply whose alternatives include an evidence act | The evidence state after the turn is as before it, but for the reply's own move |
| S5 | A reply citing a binding that is not admitted | Refused, as M2, and the model is told why |
| S6 | A malformed evidence act as the final reply (no revision, an unknown binding) | Refused, and the model is told why |

### E1.6 Replay, load and resume

| Case | Setup | Expected |
|---|---|---|
| Y1 | A game with O2, M1 and O5, exported and loaded | Replays: the same bindings, the revisions equal to the recorded ones, the same admission states and references; the records dated as the acts were made |
| Y2 | Y1 with a payload missing | Refused, naming the representation |
| Y3 | Y1 with `records.json` changed (a claim's text) | Refused: its SHA-256 |
| Y4 | A game of two model seats with the record set, stopped after a model's consent was recorded and before it was played (G), resumed | Resumes; the consent is played without asking the model; its revision is the recorded one |
| Y5 | Y4's trace replayed (`wag-llm replay --evidence DIR`) | Matches |
| Y6 | Y4's trace replayed without `--evidence` | Refused, naming the flag |
| Y7 | An export and a trace made before E1 (citations only), and a game begun before E1 resumed | Load, replay and resume as before; the citations stay citations; the trace replays on this build |
| Y8 | A new game whose citation strings are a source's reference URL and a binding's id | They stay citations |
| Y9 | Y1, Y4 and Y5 on a server or a replay with other pins (another engine) | Refused, each, naming the pins |
| Y10 | Y4 with a payload changed before the resume | Refused, naming the representation |

## Out of this step

Each is EIP's, and goes where it says:

- **Capture attempts written by WAG** (EIP item 1): the record set is made by the owner's tools
  before play; WAG's capture, with failure states written as such, comes with D077's pilot capture.
- **What each person was shown** (EIP item 1): a model's prompt records what it was shown; a
  person's board view is not recorded in this slice.
- **Moves about a source or a binding** (choice 14), and a decline act: the next step.
- **Assessments in play** (D076's "each party's assessment"): stage 2's boundary; Track D.
- **The source variable** (EIP item 5): the game does not check that an expert opinion's expert, or a
  witness, names the cited binding's source; the binding is shown with the argument.
- **Changed OCR** cannot happen in play: the record set is fixed for the game (O13 covers a changed
  interpretation).
- **"Absence of an answer"** (EIP's acceptance cases): an unanswered question is the game's (its
  deadlines), and evidence does not change it in this slice.
- **"Wrong graph revision" and "wrong context"** (EIP; plan.md §0): a target is written from the live
  graph and the context by the game, so neither can be given; a consent naming the revision of
  another target or context is refused as stale (O3).
- **Comparisons** of the practical-reasoning slice (EIP): M14 covers a value; a comparison needs D089's
  comparison schemes in play, which come with their family.
- **The other profiles** (adversarial, moderated, automated) wait for Track D (D076).

## Planted defects (each must be caught)

| Defect | Caught by |
|---|---|
| A side or a winner in an inquiry | C1, C3 |
| An inquiry of three players taken, at the server or the terminal | C2 |
| An inquiry that does not end when both pass | C3 |
| An inquiry that does not end at its round limit | C4 |
| A resignation in an inquiry names a loser | C5 |
| A target score or an answer deadline taken by an inquiry | C6 |
| The debate's side in an inquiry's prompt | C7 |
| The system prompt of other games changed | C8 |
| A rule bot refused in an inquiry without a record set | C9 |
| A time in a 0.2 revision | V1, Y1 |
| A part D076 lists left out of the revision | V2 |
| `ValidateRecords` reads a payload, or gives another result | V3, V4 |
| The record set's name or hash not recorded | R1, Y3 |
| A payload's hash not checked | R2, Y2, Y10 |
| A record the game makes taken from the record set | R3 |
| A `game:` id taken from the record set | R4 |
| A record by an author not of the record set | R5 |
| A link followed for the record set | R6 |
| A `NAME` with a separator taken | R7 |
| A record set taken by a debate, a rule bot's game or a moderated game | R8, R9, R10 |
| A record set taken without an evidence folder | R11 |
| The record set decoded loosely | R12 |
| An excerpt or a reference rendered as markup or a link | R13 |
| The context's or the actors' fields not the game's | R14 |
| An `llm` seat refused, or made a human participant | R15 |
| A record set with a binding passed alone | R16 |
| One consent admits | O1, M2 |
| The offer not counted as its player's consent | O2 |
| A consent counted for another revision | O3 |
| A duplicate consent counted | O4 |
| A withdrawal deletes the other consent or the records | O5 |
| A withdrawal without a consent accepted | O6 |
| A consent given again refused | O7 |
| The target written from the offer, not the live graph; a question's argument taken | O8 |
| A refused act leaves a record | O8, O9, O16 |
| `source_challenge` offered | O9 |
| The evaluated weight in the argument's revision | O10 |
| An accepted act read as rejected | O11 |
| A label changed by an act | O11, M7 |
| An identical offer accepted twice | O12 |
| A changed warrant keeps the old revision | O13 |
| An act counted as a pass | O14 |
| A 12-digit prefix of another binding taken | O15 |
| A refused act without the penalty | O16 |
| The argument revision's encoding changed | O17, Y1 |
| A simulation makes a real act | O18, S4 |
| The inference's proposition not `undercut(A)` | O19, M11 |
| An unadmitted binding grounds a move | M2 |
| A binding's role or target not checked against the move | M3, M4 |
| A definition or context binding meets *Require sources* | M5, M6 |
| An oppose binding cannot meet *Require sources* | M6 |
| A withdrawn binding takes the argument out of force | M7 |
| An argument with a string marked no longer grounded | M8 |
| A binding changes what a move does | M9 |
| An unknown binding id read as a citation | M10 |
| An inference binding grounds a rebuttal | M11 |
| An answer at an exception matched to the inference | M12 |
| A defence at an exception refused its oppose binding | M13 |
| An excerpt's words become a move or an act | S3 |
| A model's act played otherwise than a person's | S2 |
| A malformed act or an unadmitted citation played | S5, S6 |
| A recorded consent asked for again at a resume | Y4 |
| A replay without the record set taken | Y6 |
| An old game's prompt, trace or resume changed | Y7 |
| Old citations promoted to bindings | Y7, Y8 |
| Other pins taken for a game with a record set | Y9 |

## What versions 2 and 3 change

**The check of version 1** (`CHECK-plan-v1.md`), answered in version 2 and kept in version 3:
E-1 by choice 6; E-2 by choice 5 (the assigned weight); E-3 by choice 4's mark; E-4 by choice 1's
inquiry; E-5 by O1 and M2; W-1 by E1.1; W-2 by choice 3; W-3 by choice 2's games; W-4 by choice 4;
W-5 by choice 11; W-6 by strict decoding; W-7 by choice 10; W-8 by choice 7's table; W-9 by choice 8;
W-10 by choice 13; W-11 by choice 13's versions; W-12 by Y4 to Y6; W-13 and W-14 by the defect table
and "Out of this step"; W-15 by choice 12; the notes as the check of version 2 found them answered.

**The check of version 2** (`CHECK-plan-v2.md`):

| Finding | Disposition |
|---|---|
| E-1 read-only inspection cannot replay | Choice 13 (owner): no inspection of a game with a record set; refused, naming what; Y2, Y9 |
| E-2 `wag-llm` servers have no data folder | Choice 2: an evidence folder, `--evidence DIR`; R11; Y5, Y6 |
| E-3 answers at an exception | Choice 7's table by what the move concludes; M11 names a direct undercutter; M12, M13 |
| E-4 *Require sources* for attacks | Choice 7: any agreeing binding, support or oppose; M6 |
| W-1 participants, authors, the inference, the fields | Choices 3, 4 and 5; R14, O19 |
| W-2 the inquiry's rules | Choice 1 (owner): resignation, clocks, refused settings; C5, C6 |
| W-3 E1.0's prompt, system prompt, board, terminal | Choice 1; C2, C7, C8 |
| W-4 the record set's home and hash | Choice 2: `NAME/records.json` and its bytes' SHA-256; R1, Y3 |
| W-5 the argument revision's algorithm | Choice 5: `wag-argument-revision/1`, a golden test; O17 |
| W-6 one defect per rule | The defect table, with a case for each rule the check listed |
| W-7 EIP's omissions | "Out of this step"; M12 (a question about an exception), M14 (a value) |
| W-8 other simulations | Choice 11; O18 |
| W-9 "can still be argued about" | Choice 14 (owner): a departure, put to the owner |
| W-10 the board for a person's consent; R8's scope | Choice 12; R9 and C9 |
| W-11 an API that checks payloads once | Choice 2: `ValidateRecords`; V3, V4 |
| W-12 R13's harness; links | Choice 12: Playwright, recorded; no `rich`, no links; R13 |
| W-13 the fixture; M11; S2 | The fixture names AS:consequences and the declaration; M14 replaces M11's duplicate; S2 names who offers |
| W-14 the records' times | Choice 6: the act's time in its mark, reused; Y1 |
| N-1 P1 to P3 | "What stands" names the tests |
| N-2 the digest and the documents | E1.3 and E1.7's rows |
| N-3 the link rule's owner | Choice 2: the server checks the folders; R6 |
| N-4 failed and partial artifacts | Choice 2: they may be in a record set; the rule dropped |
| N-5 `llm` seats and rule bots | Choice 2 (owner); R9, R15 |
| N-6 D076's wording; "automated" | Choice 1, with the review's definition |
| N-7 a record set checked alone | Choice 2: `wag-evidence --record-set`; R16 |
| N-8 the context key; authors | Choice 2: keys refused even empty; authors the record set's; R3, R5 |
| N-9 the prompt's lists | Choice 10: at most 30 claims and 30 bindings |
| N-10 the dependency | Choice 12: shown; R1 |
| N-11 payload bytes in play | Choice 2: none served |
| N-12 the pins rule's scope | Choice 13: games with a record set; Y9 |
| N-13 new fields omitted when empty | Choice 13; Y7 |

## Results

The tables above stay as they were written before the tests ran; corrections and findings are
recorded here.

### E1.0, with the contract's 0.2 (V1 to V4)

- **C1 to C9 and V1 to V4 pass**: `internal/game/inquiry_test.go`, `internal/server/inquiry_test.go`,
  `cmd/wag-cli/inquiry_test.go`, `internal/evidence/format02_test.go`. V1 to V4 are E1.2's row,
  done with E1.0 because the contract's version comes first.
- **C1, before the end.** The board's state carries an outcome only once a game is over, in every
  mode (`state.go`), and the verdict throughout. The test as first written read the board's
  outcome before the end, which no game shows, and failed; it now checks, before the end, that the
  board shows the verdict and no outcome and that the game's provisional outcome (`Outcome()`)
  names no winner, and after the end the board's outcome, as before.
- **The terminal named a tie.** Played at the terminal, an inquiry ended with "Result: TIE!", and a
  series (`--games`) counted wins and ties. Fixed: the terminal says "Result: the claim's verdict
  (an inquiry: no one wins)", and a series gives the mean scores and the verdicts. The terminal's
  test now also plays an inquiry to its end and a series, and refuses an inquiry that loses on a
  late move (C6).
- **C6 at the server.** The server takes no setting for losing on a late move (only the terminal's
  `--timeout-loses` sets it); the game's check and the terminal refuse it.
- **An older test.** `TestInvariantsHoldInEveryMode` plays every mode, and expected every game of
  two players on a claim to have a winner or a draw, and tried three players in every mode but
  solo. It now plays the inquiry with two players and expects no winner, loser or draw.
- **The board**, checked in a headless browser (the workspace's `verify/e1/browser/inquiry.js`):
  no player's side shown before or after the end; the end's panel reads "An inquiry: no one wins;
  the result is the claim's label"; no page error. "Wins" before the end is the board's
  explanation of a statement under its proof standard ("it wins"), not a player's.
- **The name.** The mode is named `inquiry`, as choice 1 says. `game-rules.md` and
  `dialogue-types.md` §9.1 say what it takes from Walton's inquiry (two fellow investigators with
  no side; the claim proven, disproven or not settled) and what it does not (the inquiry's high
  standard of proof and evidential priority), as the context presets' renaming did for the types
  the game does not play. Put to the owner with choice 1.
- **Planted defects:** 27, one or more for each of C1 to C9, V1 to V4 and the terminal's result
  (the workspace's `verify/e1/mutants-e10.py` and its two runs' output), all caught. In the first
  run one survived: an inquiry's prompt saying "You play for points" passed C7, whose "together"
  the rules paragraph also has; C7 now checks the seat's role line, and the defect is caught. Two
  defects gave no verdict (they left a variable unused) and were rewritten to keep it used.
- **Not yet:** `wag-llm play --mode inquiry` names its seats Pro and Con, and `wag-llm report`
  counts sides; E1.6, which gives `wag-llm` its `--evidence`, names them otherwise.

### E1.2 to E1.4, with the load of E1.6 (Y1 to Y3, and Y9 for a load)

- **R1 to R16, O1 to O19, M1 to M14 pass** (`internal/server/recordset_test.go`,
  `acts_test.go`, `cmd/wag-evidence/recordset_test.go`); R13 in a headless browser (the
  workspace's `verify/e1/browser/evidence.js`): the record set's claims, a script and a link
  written in Markdown among them, and its references are shown as text; no script runs, no link
  is made, no move is made, no dialog opens.
- **Y1 to Y3, and Y9 for a load, pass** (`evidence_load_test.go`), with the code of choice 13 for a
  load and a resume: the record set is read again (its SHA-256 the game's, its payloads checked),
  the server's pins must be the game's, each act is replayed with its mark's time and must give its
  recorded revisions, and a game with a record set is not opened read-only. Done now because,
  without it, loading such a game replayed its acts as refused without saying so: the replay read
  a move's acceptance from its arguments, which an act does not make. Y4 to Y8 and Y10 (model
  seats, `wag-llm replay`, old games) come with E1.5 and E1.6.
- **M11's direct undercutter.** The shipped data gives `a1`'s scheme no question whose locus is
  its inference, so its direct undercutter is an undercutter that answers no question (`attack`,
  `undercutter`, expert opinion). The test as first written used an answer at the exception
  `gain_discounted_by_bearer`, which concludes the exception, and was corrected; O10 uses the same
  undercutter. The test now checks that the undercutter concludes `undercut(a1)`.
- **Beyond the cases:** M11's test also checks the table's fourth row (a defence of `a1`'s
  inference, an attack on `undercut(a1)`, grounded by a support binding at the inference, not by
  an oppose one); O8 also refuses an exception argument and a premise given as the conclusion.
- **Ids from the history.** A refused move takes a place in the history too, so an act after one
  makes `game:b<i>` with that place counted; the tests read a new binding's id from the act's mark.
- **R6** gained a case after a planted defect survived: `records.json` itself a link (a link below
  `NAME/`).
- **The board.** A person offers with the composer's *Offer evidence*, consents and withdraws with
  the evidence panel's buttons, and cites admitted bindings with the argument forms' check boxes;
  an argument's card shows its cited bindings, admitted or withdrawn, and the mark of an argument
  no longer grounded. Checked in the browser with the record set `r13`.
- **The move digest** renders an act's parts and its mark (without its time) and a move's cited
  bindings only when present: other games keep their digests.
- **Planted defects:** 75 in the code (the workspace's `verify/e1/mutants-e124.py`, two runs) and
  3 in the board's markup (`verify/e1/browser/r13_defects.py`), all caught. In the first run 70 of
  75 were caught: one survived for want of a case (`records.json` a link, added to R6), two were
  equivalent as written (kinds `source` and `annotation` were still refused by the statement's
  check; a refused act's record was left after a check that returns first) and two gave no
  verdict (an unused import, a repeated case); all five were rewritten and caught in the second
  run. One was caught only by a test's panic (O15 read an event it did not count); O15 now counts
  the events first, and the defect is caught without one.

### E1.5 and E1.6

- **S1 to S6 pass** (`internal/server/evidence_model_test.go`); **Y4 to Y10 pass**
  (`evidence_resume_test.go`, `evidence_before_test.go`, `cmd/wag-llm/evidence_test.go`), with Y1 to
  Y3 and Y9 for a load from E1.4's commit. Y9 is checked for a load, a resume and a replay; beyond
  the cases, a replay with a changed record set is refused, naming its SHA-256.
- **S2's pairing.** A binding's revision covers the participants' declared kinds (D076), and B is
  a model in one game and a person in the other, so the two games' revisions differ, each its own.
  The test pairs every other part (the history's entries, the acts' marks, the records) and checks
  that each mark's revision is its game's.
- **Y7** uses an export and two traces recorded by the build of 4ebff26, the plan's commit, before
  any of E1's code (`wag-llm play` with scripted profiles; `internal/server/testdata/e1-before/`):
  the export loads and the game stopped early resumes on this build, the finished game's trace
  replays, and their citations, a source's reference URL and binding ids among them, stay
  citations. A trace of a game stopped early (`--max-moves`) asks a replay for a call it does not
  have, under any build, so the replay takes the finished game.
- **`wag-llm`** (E1.0's "not yet"): `play` takes `--evidence DIR` and `--record-set NAME`; in an
  inquiry the seats are A and B, the summary records them as `first` and `second` with the verdict,
  and `report` counts no wins; `replay` takes `--evidence DIR`.
- **A correction to E1.0**, found while writing the prompt's evidence section: an inquiry's prompt
  still said that the reply's confidence was how likely the move "helps your side", named "the
  other side" as who argues a scheme's exceptions, and said that in a debate pro's first argument
  states the claim. C7 looked only for the system prompt's "help your side". They now read "helps
  the inquiry", "either player" and "In an inquiry, the first argument states the claim", and C7
  checks for "your side", "other side may argue" and "In a debate". A debate's prompt is unchanged
  byte for byte.
- A decision of such a game records `wag-protocol/2+evidence`, as its trace's start does.
- **Beyond the cases**, each outcome written before it ran. (a) `wag-llm play` and `report` of an
  inquiry (`cmd/wag-llm/inquiry_play_test.go`): the summary's `first` and `second` and no sides, the
  printed line naming the claim's label, the trace's record set and protocol, its replay, the report's
  rows with no wins, and `--record-set` without `--evidence` refused before the run's folder is made.
  (b) A reply with two faulty parts of one object was refused for whichever part Go's map order gave
  first, and a replay compares the refusal's reason (the next prompt quotes it), so such a reply could
  fail its own replay. Found in the offer's parts while writing them; the parser's older code read a
  move's variable bindings and an action's declaration the same way. All three now read their parts in
  the order of their names (`internal/server/parse_order_test.go`, each case read 64 times).
- **Planted defects** (`verify/e1/mutants-e156.py` in the owner's workspace; each run on a fresh copy of
  the tree). Run 1, 29 for S1 to S6, Y4 to Y10 and the corrected wording: 28 caught; one was not applied
  (its second edit named the wrong file) and was rewritten. Run 2, that one and 8 for (a) and (b): 8
  caught; "`--record-set` without `--evidence` taken" survived, since the server refuses such a game
  too, with a message that also names `--evidence`, but only after the run's folder is made: (a) now
  checks the folder. Run 3, that one: caught. In all, 37 of 37 caught, none by a panic.

### E1.7: the independent review, its corrections and their cases (written before the cases ran)

Two reviewers, each given the code and this plan and not the developer's conclusions, read E1 from
4ebff26 to 86a2f87: one the game and the records (A), one the server, the model seats, the board and
persistence (B). Their reports are in the workspace (`verify/e1/review/`). Each finding was checked
against the code before it was accepted; all were. A3 and B1 are one finding, as are A5 and B8.

1. **An Other argument's premises are premises** (A1). The split of an argument's premises into
   premises and assumptions read its scheme's premises, and an Other argument's scheme lists none:
   every premise of one counted as an assumption, so an offer at it as a premise was refused and as
   an assumption accepted. An Other argument has premises and no assumptions.
2. **Every offerable argument has an inference** (A2). An argument whose scheme has no exceptions, or
   an Other argument, has no undercutter statement until it is undercut or asked about, and an offer
   at its inference was refused. Its inference's statement is `undercut(A)` all the same, the id an
   undercutter of it concludes; the board's composer offers it too.
3. **An act is never dated before the records already made** (A3, B1). An act's time was the clock's,
   and the contract refuses an event dated before the one before it, or before its binding: with the
   clock behind the last act (a resume on another computer, a clock set back), a valid act was
   refused, and a load on a later clock replayed it as played and refused the whole game. An act's
   time is now the clock's or, if later, the latest time of the game's bindings and events.
4. **A rule bot takes no side in an inquiry** (A4). The greedy bot added a bonus when the verdict
   favoured its side, and read an inquiry's two players as con: both argued against the claim. In
   an inquiry the verdict adds nothing to its value.
5. **ValidateRecords checks the selectors' structure** (A6). It left the selector check out entirely,
   where only its reading of the payloads is to be left out. It now checks the rest: the kind, the
   offsets and the quote's presence, the declared pages, duration and region, stray coordinates.
6. **A record set's ids and name are identifiers** (B2). The prompt writes the ids of the record
   set's records and its name as they are, unquoted, and they were checked only as present, unique
   and not `game:`: an id or a name could carry a line of the prompt. A record set's record and actor
   ids are now letters, digits, `.`, `_`, `:` and `-`, beginning with a letter or a digit, at most
   128 characters (refused by `DecodeRecordSet`, naming the id); its name letters, digits, `.`, `_`
   and `-`, beginning with a letter or a digit, at most 64 (refused as not a plain name).
7. **Games without a record set are as they were before E1** (B3; choice 13). The reply parser read
   the acts' short names (`offer`, `consent`, `withdraw`, `withdrawal`) and the fields `offer`,
   `binding_id`, `binding`, `revision` and `evidence_refs` in every game: a debate's `withdraw` was
   refused by the evidence rules rather than as an unknown type, a stray `"revision": 2` refused the
   reply, a stray `binding_id` was recorded, and a trace made before E1 with such a reply did not
   replay. The parser now reads them only in a game with a record set; the server and the game
   ignore them in a game without one, and an act's type there is an unknown type (S0), as before.
8. **An act's parts on another move are refused** (A5, B8). In a game with a record set, a move that
   is not an act, carrying an `offer`, a `binding_id` or a `revision`, is refused, naming
   `evidence_refs` (how a move cites a binding), with the usual penalty and no record.
9. **A load or a resume applies the new game's checks** (B4). An export edited to a debate with the
   record set, or to an inquiry of three, loaded. A load and a resume now refuse an inquiry that
   `CheckInquiry` refuses and a record set in any game but an inquiry of two people or models that is
   not moderated, naming why, as a new game does.
10. **The evidence folder is checked as itself** (B5). Named with a trailing separator, a link passed
    the link check (`Lstat("link/")` follows it). The folder's name is cleaned first.
11. **Assumptions are offered as assumptions** (B6). The board's composer listed an argument's
    assumptions under kind premise, which the game refuses, and had no kind assumption; the prompt
    calls them premises. In a game with a record set, the state lists each argument's assumptions,
    the composer offers kind assumption with them and kind premise without them, and the prompt
    names them assumptions. A game without a record set is unchanged.
12. **Past 30 bindings, the prompt lists those the seat must weigh** (B7). It listed the first 30,
    hiding the newest. It now picks, newest first, the bindings that await the seat's consent (not
    admitted, not consented to by it), then the admitted ones, then the rest, up to 30, lists them in
    the order they were made, and counts the others.
13. **The nits** (B). A binding's entry in the prompt gives its claim's scope, which its revision
    covers. An inquiry's board says no debate's words: its empty board says "In an inquiry, the first
    argument states the claim.", and its pass help says the inquiry ends when both players pass in
    one round. `records.json` is read from the file checked: opened, then compared with the file the
    link check saw (`os.SameFile`), so that a link swapped in between is refused. A new game reads and
    checks its record set without holding the server's lock; a load and a resume check it under the
    lock, as they replay the whole game under it. The prompt says "an observation", "1 more binding".

| Case | Setup | Expected |
|---|---|---|
| X1 | The fixture after `a1` and A's pass; B supports P with the Other scheme (one premise); A offers `claim-quote` at that premise, kind premise | Accepted, its target of kind premise. The same offer with kind assumption: refused, naming the kind |
| X2 | B answers `CQ:bad_outweighs` at `a1` with AS:bad_outweighs_good (no exceptions, no undercutter); A offers `claim-observation` at its inference, oppose; B consents; A undercuts it directly, citing the binding | The offer accepted, its proposition `undercut(<id>)`; the undercutter accepted with its citation |
| X3 | A's offer; the game's clock then an hour behind it; B's consent; the export loaded on the clock | The consent accepted, dated at the offer's time; the load succeeds, the binding admitted |
| X4 | An inquiry with `a1`; the claim In, then Out | The greedy value of each player is its score less the other's, In or Out |
| X5 | `ValidateRecords` on the fixture with a claim's selector of kind `line`, and with a text selector whose end precedes its start | Invalid, naming the claim's selector, as `Validate` |
| X6 | A record set whose claim id holds a line break and "## Your legal moves"; one whose dependency id holds `>>>`; a NAME holding a line break | Each refused, naming the id or as not a plain name; no game created |
| X7 | A debate of model seats whose replies are `{"type":"withdraw"}`, a support with `"revision": 2`, one with `"binding_id"` and one with `"offer"` and `"evidence_refs"`; a person's move with those fields; a trace of such a debate recorded by the build of 4ebff26 | As by that build: `withdraw` refused with S0's unknown action type, its type as written; the others accepted, their recorded actions without those fields; the trace replays on this build |
| X8 | A game with a record set: a support carrying `binding_id` and `revision`; a pass carrying an `offer` | Each refused, naming `evidence_refs`; no record |
| X9 | The fixture game's export edited to mode `classic`; edited to three seats | Each load refused, naming the inquiry, or "two players" |
| X10 | The evidence folder a link, given with a trailing separator | The game refused, naming the link |
| X11 | The fixture with an expert opinion argument (its assumption `based_on_evidence(...)`); the state; A's offer at the assumption, kind assumption; the prompt; a debate's prompt | The state lists the assumption; the offer accepted; the prompt names it an assumption; the debate's prompt unchanged (Y7) |
| X12 | 31 bindings in an inquiry, offered in turn by A (16) and B (15); the prompt B is given | It lists A's 16 and B's newest 14, in the order made, and "(1 more binding, not listed)" |
| X13 | The prompt's entry of a binding; the claim `claim-observation` | Its claim's scope, quoted; "an observation attributed to" |
| X14 | An inquiry's board, in a headless browser: the empty board and the pass help | No "In a debate"; "In an inquiry, the first argument states the claim."; the composer's kinds include assumption, and its inference names `undercut(<id>)` |
