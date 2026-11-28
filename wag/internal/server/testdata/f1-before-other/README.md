# f1-before-other

`game.export.json` was recorded by WAG at commit 6916369 (2026-11-28), whose S5 looked for the
question that refuses Other only at the bridges on the board: an inquiry whose claim is
`worth_finding_out_first(tax_meat,demand_falls)`, the bridges' exception, Other allowed, and one
move, Other concluding that formula as a direct move (Other reuses the statement on the board), which
that build accepted. It was made with the server's own export of a game played move by move (the test
helper `exportOf`, run in a worktree of 6916369), so it carries no build stamp.

Builds after f38d189 refuse the move in S5's second layer ("an argument of Other cannot conclude"),
by the exception's form, and the loader names the change (docs/c5-item-f.md, A9.8d and A9.13b;
`TestA913AGameRecordedBeforeTheBoundByForm`).
