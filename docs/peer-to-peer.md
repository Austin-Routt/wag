# True peer-to-peer play: a note for later

*Noted on 2026-11-24 at the owner's request. Not scheduled: the plan's order stands ([plan §0](plan.md#0-from-2026-11-22-the-order-of-work)), and when to take this up is the owner's decision.*

## What there is today

WAG is a client and a server. One computer runs `wag-server`, which holds each game and its engine; the players' browsers send it moves and show what it returns ([prototype.md §3](prototype.md#3-known-limits-of-the-prototype)).

- **Without a moderator** (the default), the players are equals: each moves only for their own seat, on its turn, and the engine refuses an illegal move. Players on other computers play through the links the owner shares (*Share*: one key per seat a person plays, and one for the moderator of a moderated game).
- **The host is not a peer.** The host's computer runs the engine and holds the game: if it stops, the game stops. On the host's own computer the page can move for whichever seat is to move, so the host's player is trusted not to play for others. The links travel unencrypted over plain HTTP, so the network has to be trusted.

"True" peer-to-peer play would need no trusted host: every player runs the engine, every player can check every move, and nobody can move for another.

## What WAG already has that it would build on

- **A game replays from its moves.** An export's moves, replayed with the same scheme data (`data_sha256`), the recorded weights and the game's settings, give the same game; loading a game works this way (`internal/server/load.go`).
- **The engine is deterministic, and tested for it.** 33 seeded games, in the five modes that do not depend on the clock, must reproduce their stored move digests, scores and endings on any build of the engine (`internal/game/testdata/reference-digests.txt`; the digest is a bounded check, not a complete identity of a game), and the game's invariants include a fresh labelling equal to the applied one (`CheckInvariants`).
- **Model seats replay without asking the model again.** A trace keeps every call and reply, and `wag-llm replay` plays it back from them.
- **D073's journal and replay** (plan §0, stage 2) is to make one versioned record of a reading, a move and its outcome, every recorded event replayable. A peer-to-peer game's log should be that journal, not a second format.

## The idea: a shared log, each peer its own referee

Every player runs WAG. The game is its log: the opening (the scheme data's fingerprint, the engine's version, the settings, the seats and their public keys, the seed), then one entry per move. Every peer applies every entry with its own engine and checks that it reaches the same state.

- **Turns are the agreement.** Only the seat to move may add the next entry, so there is no general consensus to reach: there is one writer at a time.
- **Each entry is signed** with its seat's key (Ed25519, say), and carries the hash of the entry before it and a digest of the state it leads to. A forged, out-of-turn, replayed or reordered entry fails the check; two different continuations of one entry are a fork, which every peer sees.
- **A disagreement stops the game** at the first entry whose state digest a peer does not reach, and names it. Different scheme data is refused before play, as a load refuses it today.
- **A moderator,** in a moderated game, is a seat with a key whose signed decisions are entries.
- **A model seat** is played by one peer, who pays for its calls; its replies and calls go into the log, so the others check its moves by replay without asking the model.
- **Randomness** (who opens, the rule bots' seeds) comes from the opening's hash, or a commit and reveal, so that nobody chooses it alone.
- **Notes** are not moves: signed side entries that change no state.

## Ways to carry the log, simplest first

1. **Correspondence play, by files.** After a move, WAG writes the new signed entry to a small file (or text) the player sends by any means (mail, chat); the other player's WAG checks it and applies it. No network service, no address to open; slow games only. It reuses export and load almost as they are.
2. **An untrusted relay.** A small service that stores and forwards entries and runs no engine. It can delay or drop an entry, but cannot forge one; every peer still checks everything. It avoids the trouble of connecting two computers behind home routers.
3. **Direct connections.** WebRTC data channels between the players' pages, or libp2p between their servers. No relay for the moves, but a way to find each other is still needed (a rendezvous service, or copying an offer by hand), and connections behind some routers fail without a relay.
4. **The engine in the browser.** The engine compiled to WebAssembly, so that a player needs only a browser. The vendored engine does not build for `js/wasm` today (a platform file is missing, as for the BSDs), so this needs a build fix first, and the page would carry the engine and the scheme data.

## What it would cost, and what stays hard

- **Identity.** A key per player, made and kept on their computer; losing it loses the seat. Accounts would bring back a trusted service.
- **Clocks.** Timed and speed modes need a clock all peers trust. Without one, leave them out, or accept signed times within a tolerance.
- **Abandonment.** A player who stops answering cannot be ruled to have lost without a referee; the game can only be left open, or ended by the others' signed agreement.
- **Versions.** Peers must run the same engine and parser, or replay may differ (the replay notes in [protocol.md](protocol.md) list such changes). The opening names the engine's version, and a peer with another refuses.
- **Privacy.** Every peer holds the whole log, model conversations included; nothing of a peer's own machine (folders, profiles' files) belongs in it, as today's guest copies already leave out.

## Suggested steps, when scheduled

1. **After D073's journal** (stage 2): a test that two replays of one journal, on different machines and both Go versions, give the same digests at every entry.
2. **Signed entries and the opening:** keys, the hash chain, the state digest; validators that refuse a forged, out-of-turn, replayed or forked entry, and a game whose data or engine differ.
3. **Correspondence play** over files, with a browser check of two players on two computers.
4. **Live play** through a relay; then, if wanted, direct connections and the engine in the browser.

Each step would be checked as the rest of WAG is: tests that fail before the change, mutants, both Go versions, and an independent review.
