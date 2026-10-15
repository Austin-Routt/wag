# Kialo batch 2: double annotation of "Should Governments Put A Tax On Meat?"

The second round of hand-formalizing Kialo claims as Walton scheme instances, as the audit of the first pilot recommended (§7.7):
- the pilot's 20 claims and 30 more;
- coded independently by two annotators;
- agreement measured field by field;
- disagreements adjudicated;
- the result played as a WAG game.

| File | What |
|---|---|
| `guidelines.md` | The annotation guidelines: version 2 (the audit's separate fields, the policies, the output format), and the version 3 changes the adjudication found (§5). |
| `items.json` | The 50 claims, top-down, with their Kialo context (parents, ancestors, siblings, children, votes). |
| `annotator-A.json`, `annotator-B.json` | The two independent annotations. |
| `agreement.py`, `agreement.json` | Agreement per field (Cohen's kappa), at the level of scheme families, on premises and bindings, and against the first pilot. |
| `gold.json` | The adjudicated annotations, each disagreement resolved with its reason. |
| `adjudication.md` | The disagreements by kind, how they were resolved, and what they say about the guidelines. |
| `lexicon.json`, `validate.py` | The shared atoms of the discussion, and the check that every formula instantiates `walton.yml`, every target exists and the guidelines' rules hold. |
| `to_wag.py`, `wag-report.json`, `wag-game.json` | The gold annotations played as a WAG game through the server, move by move, each with what the annotation intended and what was played. The game loads in the browser (*Load an exported game*). |

## Method, and what it does and does not show

- **Annotators.** Two annotators coded every claim. They were language-model agents, briefed alike, with no access to each other's work or to the pilot. They used the guidelines, `walton.yml` (cleaned), the compendium links, the scheme families and the game's NASCQ.
- **Agreement.** It is reported per field, because the audit's point was that one label (scheme) hid several decisions:
  - scheme, fit, role, polarity, actual target, re-parenting, evidence;
  - premises and bindings where the scheme agrees.
- **Adjudication.** Each disagreement was resolved against the guidelines and the sources, with the reason recorded. The adjudicator was a third language-model agent, not a person. No human expert has reviewed the gold annotations yet. They are a consistent, checked reading of the discussion, not a validated gold standard (`docs/review-2026-10-15-part6.md`, §5).
- **One discussion.** All 50 claims come from one Kialo discussion. Nothing here shows how the schemes and guidelines carry over to other topics.
- **Playing it.** The converter plays the result. Every move goes through the game's validation, including the scheme-fit rule (S9).

## Checking and replaying, from a clean checkout

```sh
python3 -I validate.py                      # gold.json: walton.yml instantiation, targets, rules R1–R7, lexicon
python3 -I validate.py annotator-A.json     # any annotation file (an annotator's, or a model's)
python3 -I agreement.py annotator-A.json annotator-B.json --families ../../wag/data/schemes/families.json
# the game: start wag-server, then
python3 -I to_wag.py gold.json items.json http://127.0.0.1:8080 --report wag-report.json --export wag-game.json
```

`validate.py` reads only files in the repository. The compendium links are in `docs/scheme-registry/links.json`, which gives labels only, without the book's text. It needs Python 3 with PyYAML.

## How the annotations went into the game

`wag-report.json` keeps, for each claim, what the annotation intended (`intended`: scheme, role, target, question) apart from what was played (`executed`), and says how they relate (`integration`). On the data of 2026-10-15:

| Integration | Claims | Meaning |
|---|---|---|
| as intended | 20 | Played with its scheme, in its role: 16 arguments for or against a statement or new, 1 answer to a critical question, 3 challenges of a premise. |
| fragment | 3 | The claim states a premise of its target's argument (the scheme is the target's), so it supports that premise, with Other. |
| coverage gap | 15 | The game lacks its scheme, so it was played with Other: the annotation is not at fault. The schemes: 5 `other`, 3 verbal classification, 2 cause to effect, 2 rule, and one each of ends and means, practical analogy, and practical necessary condition. |
| failed | 6 | The game has the scheme, but the intended answer to a critical question cannot be made with it. 2 negative consequences answering "side effects" (an exception); 3 consequences or goal-based reasoning answering "other means"; 1 value argument answering "good consequences to weigh against". In each, the scheme's conclusion is not the answer: the bridge the audit's N4 and §4 discuss. |
| duplicate | 1 | A challenge of a premise another claim already challenged. |
| cascade | 2 | Its target was itself played with Other, so the premise it attacks is not on the board. |
| not played | 2 | No scheme and no move: an amendment, and a causal claim on a missing premise. |

Played with Other is therefore not one thing. It covers coverage gaps, fragments and failures, and only the failures count against the scheme–question links.

## Annotating with a model

`wag-llm annotate` (`docs/llm.md`) has any model annotate these claims under these guidelines, one claim per call, top-down. It writes `annotator-NAME.json` in the annotators' format, every call in `calls.jsonl`, and the hashes of everything the model saw in `run.json`. Run it from the `wag` directory:

```sh
./wag-llm annotate --model gemini:gemini-2.5-flash,output=json,min_interval=12s --out runs/annotate-flash
python3 -I ../docs/kialo-batch2/validate.py runs/annotate-flash/annotator-gemini_gemini-2.5-flash.json
python3 -I ../docs/kialo-batch2/agreement.py runs/annotate-flash/annotator-gemini_gemini-2.5-flash.json ../docs/kialo-batch2/gold.json
```

Each call sends the guidelines, `walton.yml`, the lexicon and the game's NASCQ files: about 130,000 characters (35,000–40,000 tokens). Use `--context` to send less, and `--limit` to try a few claims first. Results are in `docs/review-2026-10-13-part5.md` §3 (batch 2) and `docs/review-2026-10-15-part6.md` (this round).
