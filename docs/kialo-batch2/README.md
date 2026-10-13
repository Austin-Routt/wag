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
| `lexicon.json`, `validate.py` | The shared atoms of the discussion, and the check that every gold formula instantiates `walton.yml` and every target exists (it needs the compendium transcription, which is not in the repository). |
| `to_wag.py`, `wag-report.json`, `wag-game.json` | The gold annotations played as a WAG game through the server, move by move. The game loads in the browser (*Load an exported game*). |

**Method.**
- **Annotators.** Two annotators (language-model agents, briefed alike, with no access to each other's work or to the pilot) coded every claim. They used the guidelines, `walton.yml` (cleaned), the compendium links, the scheme families and the game's NASCQ.
- **Agreement.** It is reported per field, because the audit's point was that one label (scheme) hid several decisions:
  - scheme, fit, role, polarity, actual target, re-parenting, evidence;
  - premises and bindings where the scheme agrees.
- **Adjudication** resolved each disagreement against the guidelines and the sources, and records the reason.
- **Playing it.** The converter plays the result. Every move goes through the game's validation, including the new scheme-fit rule (S9).

Results are in `docs/review-2026-10-13-part5.md`, §3.
