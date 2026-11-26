# Frozen game data

Each folder here is the game's scheme data as it was before a change of the data's identity, named by
its fingerprint: the `data_sha256` that a saved game, an export and a model trace record. A folder is
laid out as `wag/data` (`schemes/core.yml`, `nascq/*.json`, `contexts.json`), so a game played with
that data replays exactly with `--data` on it:

    go run ./cmd/wag-server --data data/frozen/e7141ec8dea379bc
    go run ./cmd/wag-llm replay TRACE --data data/frozen/e7141ec8dea379bc

The fingerprint covers `core.yml` and the three NASCQ files; `contexts.json` is there because `--data`
reads it, and a game replays with the weights it recorded. The files are never edited: a test requires
each folder's fingerprint to be its name.

| Folder | The data | Kept for |
| --- | --- | --- |
| `e7141ec8dea379bc` | the game's data from 2026-10-15 to step C5.1's switch on 2026-11-26 ([`docs/c5-game.md`](../../../docs/c5-game.md)), when the files began to be generated from the catalogue with the same values | the games, exports and traces made with it; the generator's reproduction test |
