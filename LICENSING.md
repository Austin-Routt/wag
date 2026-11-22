# Licences

Since 22 November 2026 WAG's own work is licensed: its code under the
Mozilla Public License 2.0 ([`LICENSE`](LICENSE)), and its documents and
annotations under Creative Commons Attribution 4.0 International
([`LICENSES/CC-BY-4.0.txt`](LICENSES/CC-BY-4.0.txt)). These are the owner's
decisions D044 and D080 ([`docs/decisions.md`](docs/decisions.md)). What WAG
does not own keeps its own terms; neither licence covers it.

A file's licence follows from what it is and where it came from, not only
from its directory. MPL-2.0 is a file-level licence: a file that contains
material from an MPL-2.0 work is MPL-2.0 as a whole. WAG's own text in such
a file is also offered under CC BY 4.0.

## MPL-2.0

- **WAG's code:**
  - the Go code in `wag/`;
  - the scripts in `tools/`, `docs/kialo-batch2/` and
    `docs/scheme-registry/walton-clean/tools/` and `check/`;
  - the web client in `wag/internal/server/web/`;
  - the CI workflows in `.github/`;
  - the build files.
- **The game's data, built into the program:** `wag/data/nascq/`,
  `wag/data/contexts.json` and `wag/data/schemes/`.
- **Every file that contains or derives from the Walton collection for
  Carneades 4,** itself MPL-2.0
  ([`docs/scheme-registry/walton-clean/LICENSE`](docs/scheme-registry/walton-clean/LICENSE)):
  - the cleaned `walton.yml`, `walton.proposals.yml` and the other outputs
    in `docs/scheme-registry/walton-clean/`;
  - its copy `wag/data/annotation/walton.yml`;
  - the game's `wag/data/schemes/core.yml`;
  - the catalogue's entries, `wag/data/catalogue/schemes/`;
  - the proposals, `docs/proposals/`;
  - `docs/scheme-registry/links.json`.

  WAG's own text in the catalogue and the proposals is also offered under
  CC BY 4.0.
- **The vendored engine,** `carneades-4/` and `GoCHR/`, under its own
  `LICENSE` files and notices, as the schemes library vendors it with WAG's
  patches ([`CARNEADES_FORK.md`](CARNEADES_FORK.md)).

For WAG's own source files, this is the notice of the licence's Exhibit A:
this Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, you can
obtain one at http://mozilla.org/MPL/2.0/.

## CC BY 4.0

- **WAG's documents:** the Markdown in the repository written for WAG
  (`README.md`, `docs/`), the decision register (`docs/decisions.yml`), the
  candidates (`docs/candidates.yml`) and the views generated from them.
- **WAG's annotations:**
  - the batch-2 gold, the annotators' files, the adjudication, the
    guidelines and the lexicon in `docs/kialo-batch2/` and
    `wag/data/annotation/kialo-13292-batch2/`;
  - the labels in `docs/kialo-pilot/`.
- **WAG's own text in the catalogue and the proposals,** besides MPL-2.0
  (above).

Attribute it to "the WAG project, https://github.com/Austin-Routt/wag", link
the licence, and say what you changed.

## Not WAG's, and not licensed by it

- **The book's text.** Walton, Reed and Macagno, *Argumentation Schemes*
  (Cambridge University Press, 2008), and other cited works. This covers
  the quotations in the repository; the transcription in
  `docs/book-catalogue/`; and the English templates of `walton.yml` where
  they repeat the book (the README, issue 6). They are cited for research
  and criticism. D043 takes the book's text out of the current tree.
- **Kialo's content.** The claims of Kialo discussions, quoted in the
  annotation items and the pilot's files, belong to their authors under
  Kialo's terms. WAG's labels on them are CC BY 4.0.
- **Captured evidence.** None is in this repository. It is kept in the
  private workspace, for research (D077).
- **The public history.** Earlier commits stay as they are (D078), the
  deleted book included; these licences do not cover it.

This file describes the licences; it is not a legal clearance of every
file. A release that redistributes files lists their sources.
