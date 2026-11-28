# f1-before

`game.export.json` was recorded by WAG at commit 6303759 (2026-11-28), whose S5 looked for the
question that refuses Other only at the bridges on the board: an inquiry whose claim is
`worth_finding_out_first(tax_meat,demand_falls)`, the bridges' exception, and one move, that claim
argued with sign before any bridge, which that build accepted. It was made with the server's own
export of a game played move by move (the test helper `exportOf`), so it carries no build stamp.

Builds after 6916369 refuse the move by the exception's form (docs/c5-item-f.md, A9.8 to A9.13;
docs/game-rules.md, S5), and the loader names the change and the build that replays the game
(`TestA913AGameRecordedBeforeTheBoundByForm`).
