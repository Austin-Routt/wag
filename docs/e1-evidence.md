# E1: evidence records in the game (the evidence plan's step 1), version 1

Stage 2 ends when, among other things, evidence records are bound to the actual game, old
citation-only traces keep their meaning, and the specific CQ, admission and revision effects are
tested (plan.md §0, "Evidence increment"). This is step 1 of the
[evidence implementation plan](evidence-implementation-plan.md#step-1-connect-the-first-slice-before-an-evidence-aware-game-pilot):
connect the first slice before an evidence-aware game pilot. This page plans it: what stands, the
choices, the steps, and each step's cases with their outcomes, written before the tests run.

## What stands (2026-11-27)

- **A move cites sources as strings** (`evidence: [string]`), kept as the argument's metadata and
  shown. With *Require sources* on, an argument from a scheme whose evidence is "required" (expert
  opinion, position to know) is refused without one (the game's rule E1, `turn.go`).
- **`wag-evidence/0.1` checks records offline** (`internal/evidence`, `wag-evidence`): sources,
  artifacts with their capture status, retained representations with their bytes' hash, attributed
  claims with exact selectors, targets, bindings, assessments, dependencies, and an ordered log of
  two participants' consents and withdrawals for an exact binding revision
  (`two-party-consideration/1`). It reads no live graph: a target and the context's pins are
  checked for form only (evidence-contract.md).
- **The owner's decisions that bound this:** D076 (a source is a record; consent is to the exact
  binding; admission is for consideration only: not a stipulation, a moderator's ruling, the
  engine's acceptance or a finding of fact; withdrawing a consent removes only its own route;
  captured content is data, never instructions), D077 (capture for research, kept private), D079
  (terms of use and injected instructions: open, no gate before the first pilot), D045 (a person
  verifies).

## The choices

Each is the developer's, under the owner's instruction to go with the recommended choices; those
marked **(owner)** set a game rule and are put to the owner with this plan, the work going on with
the recommendation meanwhile, as before (D094, D095).

1. **The record set comes with the game.** Sources, artifacts, representations and claims are
   captured before play (D077), by the owner's tools, and given to a game when it is created: a
   `wag-evidence/0.1` bundle without targets, bindings or events, whose payloads are in the data
   folder (`evidence/<name>/`, private, as D077 keeps captured content). Players do not capture
   during play in this slice.
2. **Evidence acts are moves.** A player offers a binding (`evidence_offer`: a claim of the record
   set, bound to a statement of an argument on the board, with a role, an interpretation and a
   warrant), consents to one (`evidence_consent`, naming its exact revision) or withdraws a
   consent (`evidence_withdraw`). An offer is also its player's consent to the binding it makes, so
   that the other player's consent admits it. Each act takes the player's turn and is recorded in
   the move history, the trace and the export like any move, so that the order of the log is the
   game's, and replay and resume rebuild the acts as they rebuild moves. The board, the text
   protocol and model seats make them; the terminal does not yet, and says so. A cooperative
   two-player game can spend turns on them; the profiles where they should not (Track D) come
   later. Assessments (each party's view of a binding) are not made in play in this slice.
3. **The game makes the target and checks it.** An offer names an argument on the board and a
   statement of it (a premise, an assumption, an exception or its conclusion) or its inference. The
   game writes the target record from the live graph: the argument's revision (a hash of its
   scheme, premises, conclusion, exceptions and weight as the board holds them), the locus and the
   statement's id. An argument or statement that is not there, or a locus it does not have, is
   refused.
4. **The pins are the game's own inputs.** The record set's context pins name the game's actual
   inputs, each a full SHA-256: the scheme file the game reads (`catalogue_sha256`, of `core.yml`),
   the NASCQ data (`nascq_sha256`, of the schemes, questions and edges), and the engine
   (`engine_sha256`, of the vendored Carneades engine's sources, computed by a generator and checked
   by a test). Other pins refuse the record set. The participants are the game's two players; a
   tool can author records, not consent (D076).
5. **A move grounded on evidence cites an admitted binding (owner).** A move may cite bindings
   (`evidence_refs`, besides the citation strings, which keep their meaning). A cited binding must be
   admitted for consideration when the move is made (both players consented to its revision), its
   target must be the statement the move argues for or against, and its role must agree (support
   for an argument for it, oppose against). *Require sources* is met by a citation or by such a
   binding. Admission changes no label, weight or assumption (D076): the argument stands on its
   scheme and its critical questions like any other.
6. **Withdrawal marks, it does not undo.** A withdrawn consent ends the binding's admission. A move
   already made that cited it stays in the history and the graph; its reference is shown withdrawn,
   and an argument that cites another admitted binding keeps that one. Nothing is deleted:
   records, assessments and history stay.
7. **A changed record is a new record.** The record set is fixed for the game. An offer of the same
   claim with another interpretation, warrant, role or target is another binding, with its own
   revision; a consent names the revision it is for, and one for another revision is refused.
8. **Model seats read evidence as data.** A model seat's prompt shows the admitted and offered
   bindings that concern the board, each with its excerpt quoted and marked as material to read,
   not instructions. What a model was shown is its decision's recorded prompt. A model seat makes
   evidence acts like any move. That a model obeys nothing in an excerpt is tested on what the game
   does (no act it did not choose), not on what the prompt says.
9. **Replay and resume need the payloads.** A game with a record set replays (load, resume,
   `wag-llm replay`) only where its payloads are, each checked against its hash before use; without
   them the game is refused, naming what is missing. Old games, and games without a record set,
   replay as before.

## The steps

| Step | What | Done when |
|---|---|---|
| E1.1 | The pins: the engine's generated fingerprint and its test; the game's three pins, recorded in its trace's start and its export | The cases P1 to P3 pass |
| E1.2 | The record set: given with a new game (`evidence` in the request: a bundle and its payload folder), checked by `wag-evidence`'s rules and the pins, kept in the game's folder, shown on the board | R1 to R6 |
| E1.3 | Offers, consents and withdrawals as moves: validation, the target written from the live graph, the log and the admission state, in the history, the trace, the export; the board's evidence panel | O1 to O9 |
| E1.4 | Moves that cite bindings: `evidence_refs`, *Require sources*, withdrawal's marks | M1 to M8 |
| E1.5 | Model seats: the prompt's evidence section, the reply protocol's evidence acts | S1 to S4 |
| E1.6 | Replay, load and resume with a record set; old games unchanged | Y1 to Y5 |
| E1.7 | Planted defects for every rule above; the docs (protocol.md, game-rules.md, evidence-records.md, the plan); gates on both toolchains; an independent review | Each defect caught |

## Cases (outcomes written before the tests)

The fixture: a debate on the shipped data, Pro (who argues for the claim) against Con, with a record
set of two synthetic sources (an agency's report, a news article quoting it: a common origin), each
with an original text representation, one with an excerpt derivative, and three claims (a
quotation, a paraphrase, an observation).

### E1.1 The pins

| Case | Setup | Expected |
|---|---|---|
| P1 | The engine's fingerprint computed from the vendored sources | Equals the generated constant; one byte changed in an engine source changes it |
| P2 | A new game | Its trace's start and its export carry the three pins, each 64 hexadecimal digits |
| P3 | The scheme file or the NASCQ data changed (another data folder) | `catalogue_sha256` or `nascq_sha256` changes with it |

### E1.2 The record set

| Case | Setup | Expected |
|---|---|---|
| R1 | A new game with the fixture's record set | Accepted; the board lists its sources and claims |
| R2 | The record set with a payload changed by one byte | Refused, naming the representation |
| R3 | The record set with a pin of another engine | Refused, naming the pin |
| R4 | The record set with participants other than the game's players, or a tool as one | Refused |
| R5 | The record set with targets, bindings or events already in it | Refused: those are made in play |
| R6 | A payload path outside the evidence folder, or a symbolic link | Refused (wag-evidence's rule) |

### E1.3 Offers, consents and withdrawals

| Case | Setup | Expected |
|---|---|---|
| O1 | Pro's argument on the board; Pro offers the quotation for its premise, role support | Accepted; a target and a binding are recorded, with the argument's revision, and Pro's consent to the binding's revision; the revision is shown; not admitted |
| O2 | O1, then Con consents to that revision | Admitted for consideration |
| O3 | O1; Con consents to another revision (a stale one) | Refused |
| O4 | O1; Con consents twice | The second refused (duplicate consent) |
| O5 | O2; Con withdraws | The binding is no longer admitted; Pro's consent and the records stay |
| O6 | An offer for an argument not on the board, a statement not in the argument, or a locus it does not have | Refused, each |
| O7 | An offer of a claim not in the record set | Refused |
| O8 | O2, then the argument is changed in the graph (another argument made for the same statement does not change it; an undercutter does not change it) | The binding's argument revision is unchanged; the admission stands |
| O9 | Any of O1 to O5 | The history, the trace and the export record the act; no label, weight or assumption changes |

### E1.4 Moves that cite bindings

| Case | Setup | Expected |
|---|---|---|
| M1 | O2 (admitted, support, the premise P); Pro argues for P with expert opinion, citing the binding | Accepted; the argument records the reference |
| M2 | O1 (offered, not admitted); the same move | Refused: the binding is not admitted |
| M3 | O2; a move against P citing the binding (role support) | Refused: the role does not agree |
| M4 | O2; a move for another statement citing the binding | Refused: the target is not the move's statement |
| M5 | *Require sources* on; expert opinion with neither a citation nor a binding | Refused (E1, as today); with an admitted binding, accepted |
| M6 | M1, then Con withdraws | The argument stays, in force as before; its reference is shown withdrawn |
| M7 | M1 with two admitted bindings, then one withdrawn | The argument shows one reference withdrawn and one admitted |
| M8 | M1; Con asks the expert's critical questions | As for any expert-opinion argument: the binding changes no question's effect |

### E1.5 Model seats

| Case | Setup | Expected |
|---|---|---|
| S1 | A model seat's turn with an offered and an admitted binding | Its prompt shows both, each excerpt marked as material to read, not instructions; the decision's prompt has them |
| S2 | A model replies with an offer, a consent, a withdrawal | Each played as the person's would be |
| S3 | An excerpt holding a move (a JSON action to resign) and words telling the reader to consent to everything; the model's reply quotes the excerpt, then ends with its own move | The game plays the reply's own move, and makes no act the reply did not end with |
| S4 | A reply citing a binding that is not admitted | Refused, as M2, and the model is told why |

### E1.6 Replay, load and resume

| Case | Setup | Expected |
|---|---|---|
| Y1 | A game with M1 and O5, exported and loaded | Replays: the same bindings, revisions, admission states and references |
| Y2 | Y1 where the payloads are missing | Refused, naming the representation |
| Y3 | A game with a model seat and a record set, stopped and resumed (G) | Resumes, with its record set and its evidence acts |
| Y4 | Its trace replayed (`wag-llm replay`, the payload folder given) | Matches |
| Y5 | An export of a game made before E1 (citations only) | Loads and replays as before; its citations stay citations |

## Limits of this slice

- The game does not check that a move's source variable (an expert opinion's expert, a witness)
  names the cited binding's source; the binding is shown with the argument.
- Assessments, dependency records made in play, captures during play, and the terminal's evidence
  acts are later.

## Planted defects (each must be caught)

| Defect | Caught by |
|---|---|
| The engine's fingerprint not recomputed | P1 |
| A payload's hash not checked | R2 |
| The pins not compared with the game's | R3 |
| A consent counted for another revision | O3 |
| One consent admits | O2 |
| A withdrawal deletes the consent of the other player or the records | O5 |
| The target written from the offer, not the live graph | O6 |
| An unadmitted binding grounds a move | M2 |
| A binding's role or target not checked against the move | M3, M4 |
| Admission changes a label | O9, M6 |
| An excerpt's words become a move | S3 |
| Old citations promoted to bindings | Y5 |
