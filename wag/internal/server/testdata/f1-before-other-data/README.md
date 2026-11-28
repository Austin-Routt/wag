# f1-before-other-data

Two games recorded on 2026-11-28 by the independent check of K, each with a wag-server built from
commit 6916369 and started with `--data` on one of the frozen data folders (`wag/data/frozen/`): an
inquiry whose claim is `worth_finding_out_first(tax_meat,demand_falls)`, the bridges' exception, and
one move, Ann's support of it with AS:sign before any bridge, which that build accepted. The binary was
built outside a checkout, so the exports carry no commit stamp; their `data_sha256` names the data.

- `092b6a6f48a672d2.export.json`: data that has the question refusing Other (CQ:worth_finding_out_first).
  This build refuses the move with that data too, by the exception's form.
- `afd3335a23f12f41.export.json`: data from before step C5.2, which has no such question; its data
  bounds nothing by the form, and this build loads the game on it.

docs/c5-item-f.md, A9.13c-A9.13e; `TestA913dTheFormBoundIsTheData`.
