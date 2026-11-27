# The inputs the format-1 gold was made with

`gold-format1.json` is the batch-2 gold before format 2, kept as it was (decision D039; the identity
contract, D092, 2026-11-24). A kept version is checked against the inputs it was made with, not as
they are since, so that a later decision does not fail it:

- `walton.yml`: the scheme collection as it was until 2026-11-25 (SHA-256 bb1edbcc6953…), before
  proposal D008-D074 added exceptions to the consequence schemes, and added their modal variants.
  `validate.py` reads it for a format-1 file, and `TestFormat1IsCheckedAsMade` builds its checker
  on it. `test_validate.py` checks that it is that walton.yml (its hash), and runs its format-1
  tests on it.

- The game's NASCQ (`schemes.json`, `critical_questions.json`, `edges.json`): since step C5.2's switch
  (2026-11-27) the game asks questions at formulas where it asked none, which R2 would ask a kept
  version to name (13292.127 and 13292.43 name the book's questions). For a format-1 file
  `validate.py` reads the files frozen in `wag/data/frozen/e7141ec8dea379bc`, the game's data of
  2026-10-15 to step C5.1's switch: the NASCQ files of WAG 96c3c5a, where the format-1 gold was
  made, byte for byte. That folder is never edited, so no copy is kept here; `test_validate.py`
  checks the three files' hashes.

The other inputs a format-1 check reads (`items.json`, `lexicon.json`, `links.json`) have not
changed in ways that touch it: `links.json` gained two links on 2026-11-25, and the format-1 gold
passes with them. When one does change in a way that fails it, its copy is kept here too.

What the copy covers: `validate.py` reads it for every format-1 file, the format-1 gold (made with
it) and the annotators' files of 2026-10-13 (made with an earlier `walton.yml`, as before). The Go
side re-checks annotation runs recorded in format 1 with the dataset's `walton.yml` when it shows
them, since such runs pin no inputs: a limit of D092's "checked as they were made".
