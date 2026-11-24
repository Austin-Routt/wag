# Kialo batch 2: double annotation of "Should Governments Put A Tax On Meat?"

The second round of formalizing Kialo claims as Walton scheme instances, as the audit of the first pilot recommended (§7.7):
- the pilot's 20 claims and 30 more;
- coded independently by two annotators. Both annotators and the adjudicator were runs of one language model with the same guidelines; no person has coded or reviewed these claims yet (`../review-2026-10-15-part7.md`);
- agreement measured field by field;
- disagreements adjudicated;
- the result played as a WAG game.

| File | What |
|---|---|
| `guidelines.md` | The annotation guidelines: version 2 (the audit's separate fields, the policies, the output format), the rules on opposite consequences and weighing (§3.1, rules OC1–OC14, decisions D055 and D059), practical reasoning's agent, its answers and the side effect that outweighs (§3, §3.1; decisions D005, D006, D070 and D088-D091, 2026-11-24), and the version 3 changes the adjudication found (§5). |
| `rule-cases.yml` | Worked cases for each rule of §3.1, with their codings and the rule's retired wording. A test (`wag/internal/annotate`, `TestRuleCases`) checks the codings against `walton.yml` and the gold, that every rule has a case, and that no retired wording survives. Not given to annotators: some cases are claims of this batch. |
| `items.json` | The 50 claims, top-down, with their Kialo context (parents, ancestors, siblings, children, votes). |
| `annotator-A.json`, `annotator-B.json` | The two independent annotations. |
| `agreement.py`, `agreement.json` | Agreement per field (Cohen's kappa), at the level of scheme families, on premises and bindings, and against the first pilot. |
| `gold.json` | The adjudicated annotations, each disagreement resolved with its reason. A later recoding is recorded in the item's `adjudication.recoded`. |
| `recode_d055.py`, `recode_d055_review.py` | The recoding of the gold for decisions D055 and D059 (2026-10-23): the sided exceptions, the weighing claims 13292.127 and 13292.43, and the 33/CQ3 challenges split; then 13292.29 naming C007, as OC5 asks, after the independent review of the application. One-offs, kept as the record; each refuses to run twice. |
| `recode_c2.py` | The recoding of the gold for proposal D004-D075 (2026-11-24): practical reasoning's agent and goal premise, 13292.45 on the repaired necessary-condition scheme, the side effects that rank nothing as rebuttals (13292.174, .5), the better means (13292.104, .123, with 13292.208), 13292.147 as a proposal of another course, the weighing claims 13292.43 and 13292.127 with C003's answer schemes, and the challenge records. A one-off, kept as the record; it refuses to run twice. |
| `adjudication.md` | The disagreements by kind, how they were resolved, and what they say about the guidelines. |
| `lexicon.json`, `validate.py`, `test_validate.py` | The shared atoms of the discussion, and the check that every formula instantiates `walton.yml`, every target exists and the guidelines' rules hold (R1–R11; R8–R11 since 2026-11-24: a claim at an exception concludes it, or its negation when it defends the argument; a better means records its comparison and its target's goal; a challenge record names a formula of its claim or its target; an answer scheme's own argument is at an exception, unless it is a fragment backing a premise). `test_validate.py` plants a defect for each of R8–R11 and checks that it is caught. |
| `to_wag.py`, `wag-report.json`, `wag-game.json` | The gold annotations played as a WAG game through the server, move by move, each with what the annotation intended and what was played. The game loads in the browser (*Load an exported game*). |

## Method, and what it does and does not show

- **Annotators.** Two annotators coded every claim. They were language-model agents, briefed alike, with no access to each other's work or to the pilot. They used the guidelines, `walton.yml` (cleaned), the compendium links, the scheme families and the game's NASCQ.
- **Agreement.** It is reported per field, because the audit's point was that one label (scheme) hid several decisions:
  - scheme, fit, role, polarity, actual target, re-parenting, evidence;
  - premises and bindings where the scheme agrees.
  - Re-parenting is derived from the target (actual target ≠ Kialo parent, as the guidelines say the tooling computes it), not read from the annotators' own field: since 24 November 2026 in `agreement.py` and the Go measure alike. The annotators' own values, which followed an older reading (set whenever the target was a premise), agreed at 0.74 (κ 0.513); derived, they agree at 1.0. Their 13 disagreements were all that older reading (`adjudication.md`, R1).
- **Adjudication.** Each disagreement was resolved against the guidelines and the sources, with the reason recorded. The adjudicator was a third language-model agent, not a person. No human expert has reviewed the gold annotations yet. They are a consistent, checked reading of the discussion, not a validated gold standard (`docs/review-2026-10-15-part6.md`, §5).
- **One discussion.** All 50 claims come from one Kialo discussion. Nothing here shows how the schemes and guidelines carry over to other topics.
- **Playing it.** The converter plays the result. Every move goes through the game's validation, including the scheme-fit rule (S9).

## Checking and replaying, from a clean checkout

```sh
python3 -I validate.py                      # gold.json: walton.yml instantiation, targets, rules R1–R11, lexicon
python3 -I test_validate.py                 # R8-R11 catch the defects they are for
python3 -I validate.py annotator-A.json     # any annotation file (an annotator's, or a model's)
python3 -I agreement.py annotator-A.json annotator-B.json --pilot ../kialo-pilot/pilot.json --families ../../wag/data/schemes/families.json --json agreement.json
# the game: start wag-server, then
python3 -I to_wag.py gold.json items.json http://127.0.0.1:8080 --report wag-report.json --export wag-game.json
```

`validate.py` reads only files in the repository. The compendium links are in `docs/scheme-registry/links.json`, which gives labels only, without the book's text. It needs Python 3 with PyYAML.

## How the annotations went into the game

`wag-report.json` keeps, for each claim, what the annotation intended (`intended`: scheme, role, target, question) apart from what was played (`executed`), and says how they relate (`integration`). On the data of 2026-11-24, after the analogy family's two-case form (decisions D001, D028, D029), the consequence family's sided exceptions (D055, D059) and practical reasoning's proposal D004-D075, and before the game's data takes them (workstream E, step C5):

| Integration | Claims | Meaning |
|---|---|---|
| as intended | 21 | Played with its scheme, in its role: 17 arguments for or against a statement or new, 1 answer to a critical question, 3 challenges of a premise. |
| fragment | 3 | The claim states a premise of its target's argument (the scheme is the target's), so it supports that premise, with Other. |
| coverage gap | 15 | The game lacks its scheme, so it was played with Other: the annotation is not at fault. The schemes: 5 `other`, 3 verbal classification, 2 cause to effect, 2 rule, and one each of ends and means, practical analogy, and practical necessary condition. The practical analogy (13292.30) now concludes the thesis itself, `should_be_performed(tax_meat)`, so its Other argument supports the thesis's own statement. |
| failed | 1 | The game has the scheme, but the intended move cannot be made with it: 13292.61, coded with analogy's two-case form, which the game's `AS:analogy` (still A and C) does not take until its data is generated from the catalogue; until then it is played with Other. |
| duplicate | 1 | A challenge of a premise another claim already challenged. |
| cascade | 1 | Its target was itself played with Other, so the premise it attacks is not on the board. |
| not played | 7 | No scheme and no move: an amendment, and a causal claim on a missing premise. The two weighing claims, 13292.127 and 13292.43, and the two better means, 13292.104 and 13292.123, whose targets, the new exceptions, are not on the game's board until workstream E (step C5); and 13292.208, which attacks 13292.104's premise (by the script's own definitions a cascade). |

Played with Other is therefore not one thing. It covers coverage gaps, fragments and failures, and only the failures count against the scheme–question links.

Since 2026-10-23 the replay plays an answer at an exception as an argument that the exception holds. An annotation's polarity is its stance on the target's conclusion, so it flips at an exception. Before, the two answers about side effects and the three about other means (13292.174, .5, .104, .147, .123) were played against the exceptions they answer. The counts did not change, but the evaluation did: 13292.118's goal-based practical reasoning, in force before, is now undercut by 13292.123's other-means answer, and 13292.108's, out before for a premise, is now undercut by the two side-effect answers (13292.174, .5) and two other-means answers (13292.104, .147). All five answers are Other arguments, so their weights do not matter (D070). The thesis stays out.

Since 2026-10-23 the gold is recoded for D055 (`recode_d055.py`). 13292.127, a negative consequences argument before, is now a weighing claim against 13292.115's new exception, and 13292.43, a value argument before, one against 13292.9's; neither target is on the game's board yet, so neither is played. 13292.127's starvation argument, now its secondary instance, leaves the replayed debate, and 13292.204, which denies the starvation link, is played as a new Other argument that attacks nothing. The script labels it a coverage gap, though by its own definitions it is a cascade. The thesis stays out. The fourth independent review's replay of the same recoding gave the same result, claim by claim.

Since 2026-11-24 the gold is recoded for proposal D004-D075 (`recode_c2.py`). 13292.174 and 13292.5, side effects that rank nothing against 13292.108's goal, are rebuttals at its conclusion now, played as negative consequences against the thesis, and 13292.147, a proposal of another course, is played as positive consequences for rationing: three failures before, now as intended. 13292.104 and 13292.123 answer with `better_means` at the exception `realize_more_effectively(governments,G,tax_meat)`, which the game's copies of practical reasoning do not have until step C5, so they are not played, and neither is 13292.208, which attacks 13292.104's better-means premise; 13292.127 weighs against 13292.108's exception now, still not on the board. 13292.118's goal-based practical reasoning, undercut before by 13292.123's Other answer, is in force again; 13292.108 stays out for an unanswered premise question; the thesis stays out. The export of 2026-10-23 had the thesis in under the policy preset; replayed from the gold before this recoding, today's server gives out there too, so the change came with the server's changes since 2026-10-23, not with the recoding.

## Annotating with models

The dashboard's *Annotate* tab (`docs/llm.md`) has models annotate these claims under these guidelines, one claim per call, top-down, several models in parallel. A reviewer model compares their annotations, and you give your verdicts, claim by claim. The dataset is built into WAG (`wag/data/annotation/kialo-13292-batch2`, a copy of this folder, kept equal by a test).

Each run is a folder in your data folder (`annotation/RUN/`). Its `annotations/NAME.json` files are in the annotators' format, so the scripts here read them:

```sh
python3 -I validate.py ~/WAG/annotation/RUN/annotations/NAME.json
python3 -I agreement.py ~/WAG/annotation/RUN/annotations/NAME.json gold.json
```

The dashboard computes the same agreement figures itself; `report.json` has them, and the Go code reproduces `agreement.py`'s on the two annotators here. The command line does the same as the dashboard:

```sh
./wag-llm annotate --model a=gemini:gemini-2.5-flash,output=json,min_interval=12s --model b=lmstudio:qwen3-8b,output=json \
    --reviewer a --limit 5 --out runs/batch2-try
```

Each call sends the guidelines, `walton.yml`, the lexicon and the game's NASCQ files: about 130,000 characters (35,000–40,000 tokens). On a free tier, start with a few claims. Results are in `docs/review-2026-10-13-part5.md` §3 (batch 2) and `docs/review-2026-10-15-part6.md` (this round).
