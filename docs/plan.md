# The plan: one verified scheme catalogue, then annotation on it

Status: proposed on 2026-10-15, not started. The owner decides the open questions in §7 before work begins.

- **Why this plan,** from the first real model run: [`review-2026-10-15-part7.md`](review-2026-10-15-part7.md).
- **What Walton and Macagno's classification adds:** [`review-2026-10-15-part8.md`](review-2026-10-15-part8.md).
- **Dialogue types,** the second objective, are Track D (§8). It rests on the reference [`dialogue-types.md`](dialogue-types.md) and the review [`review-2026-10-15-part9.md`](review-2026-10-15-part9.md).
- **Where the project stands:** [`README.md`](../README.md).

## 1. Principles

- **One catalogue, with an authority for each part.** The owner's decision of 2026-10-15:
  - The book (Walton, Reed & Macagno, *Argumentation Schemes*, 2008) decides what a scheme says, and its critical questions.
  - The human-verified Carneades collection (`walton.yml`) decides the formulas.
  - The game's NASCQ is generated from the catalogue: a selection with game rules added, not a separate source.
  - Schemes found in Kialo data join as *derived* schemes, with their evidence.
- **Every field says who stands behind it:**
  - *drafted* (by a model);
  - *verified* (by a person, with a date);
  - *disputed* (with the reason).

  Where the book and `walton.yml` disagree (for example the book's modal cause to effect, audit N8), the entry records both and the decision.
- **No book text in the public repository.**
  - The catalogue holds our own paraphrases, the book's labels (scheme number, CQ number) and page references.
  - The book stays the owner's copy, outside the repository, for verification.
  - This is how `links.json` works already.
- **No new scheme for a one-off pattern.** A derived scheme needs recurring evidence and a person's review.
- **Measure before and after each change,** on the same claims and the same model (§5).

## 2. Phase 1: the aligned catalogue

### What an entry holds

One file per scheme, for example `catalogue/schemes/negative_consequences.yml`, so that a review is a readable diff.

| Field | Source | Done by |
|---|---|---|
| id, name, book number(s) and pages, family | `walton.yml`, `links.json`, `families.json` | generated |
| summary: when the scheme applies, in two sentences of our own | new | drafted, then verified |
| variables, premises, assumptions, exceptions, conclusion: formulas and English | `walton.yml` | generated |
| map from the book's premises to the formulas | `links.json` premise maps | generated, then verified |
| critical questions: book label, our paraphrase, locus (premise, assumption, exception, conclusion, argument), the formula it is about, the game's CQ id if any, what counts as an answer | `links.json` CQ maps (197), the NASCQ (24), new paraphrases | drafted, then verified |
| contrasts: "this, not X, when …", for each close relative | the registry's families and signatures, new text | drafted, then verified |
| examples: two or three short claims coded in full (Kialo claims with their ids where possible), and one near miss | new | drafted, then verified |
| discrepancies between book and `walton.yml`, and the decision | the audit, `links.json` partial links | drafted, then decided |
| place in Walton and Macagno's classification: source-dependent or not; practical or epistemic; discovering a rule or applying one | Walton & Macagno 2015 (part 8 §1.7) | drafted, then verified |
| "specialises": the scheme this one is a domain version of, if any | new | drafted, then verified |
| proper in: the dialogue types where the scheme is a legitimate move (a threat is, in a negotiation, and is not, in a persuasion) | the book, Walton's pragmatic theory of fallacy | drafted, then verified |
| status and history of each field | | |

### Where the sources stand today

| Source | What it has | State |
|---|---|---|
| The book's compendium (ch. 9) | 109 schemes (128 numbered entries, 19 of them headings only), 153 critical questions | Transcribed outside the repository |
| `docs/scheme-registry/walton-clean/walton.yml` | 111 schemes with formulas and English; 54 encode some CQs as assumptions or exceptions | Cleaned so that Carneades loads it; content unchanged |
| `docs/scheme-registry/links.json` | 93 exact and 13 partial links from the book to `walton.yml`, 3 book schemes without one; 197 links from a CQ to a formula | Drafted by a model, not verified |
| `wag/data/nascq/*.json` | 9 schemes, 24 CQs, 35 ask edges, 4 backup edges | The game's data, tested |
| `wag/data/schemes/families.json` | 21 families | Drafted |

`walton.yml` has 6 schemes the compendium does not list:
- `defeasible_modus_ponens`;
- `epistemic_ignorance`;
- `exception_to_rule`;
- `exceptional`;
- `popular_practice`;
- `threat`.

### Order of work

1. **Generate the skeleton of all entries** from the sources above, and a checker:
   - every formula is `walton.yml`'s;
   - every CQ's formula is one of its scheme's;
   - every game CQ appears;
   - no paraphrase reproduces the book's sentence.
2. **Tier 1: about 25 schemes.** These are the 13 schemes of the batch-2 gold and the close relatives needed for contrasts:
   - **practical reasoning:** practical reasoning 1 and 2, practical goal, end and means, necessary condition;
   - **consequences and values:** positive and negative consequences, positive and negative value;
   - **causes and signs:** cause to effect, correlation to cause, sign;
   - **cases and analogy:** analogy, the practical analogies, example;
   - **sources:** expert opinion, position to know, witness testimony, popular opinion;
   - **classification and rules:** verbal classification, definition to verbal classification, rule, established rule;
   - **alternatives and costs:** the alternatives, sunk costs and waste;
   - **slippery slope.**

   A model drafts every field; a person verifies them against the book.
3. **Tier 2: the rest,** in the order the data asks for them: the schemes that annotators propose as near misses, and the "other" candidates.

**The index's top level is Walton and Macagno's tree.** Our families sit under it, and the few schemes it leaves out (some linguistic ones) are placed by hand. **`defeasible_modus_ponens`** is in tier 1 as the general form "other" falls back to (§3). **A derived scheme must pass Hansen's definition:** a pattern of sentential forms with variables, at least one schematic constant, and a conclusion.

**Effort.**
- **Drafting:** a few working sessions with a model.
- **Verification:** about 20–30 minutes a scheme for a person with the book. That is 8–12 hours for tier 1, and 35–55 hours for everything.

**Done when:**
- every tier-1 entry is verified;
- the checker passes;
- the NASCQ is generated from the catalogue and the game's tests pass unchanged.

## 3. Phase 2: annotation, version 4

### The guidelines: one document, a procedure

The current guidelines are version 2 with the version 3 changes appended as a list of edits. They allow two answers in places and contradict each other in one. Version 4 is one document, in the order an annotator decides:
0. **Once per discussion, by a person:**
   - the setting, and the dialogue type, both supposed and enacted ([`dialogue-types.md`](dialogue-types.md) §9.3). A Kialo discussion is public persuasion addressed to readers; on a practical thesis, it is persuasion about whether an action is advisable (ND p. 229), using deliberation's arguments but deciding nothing jointly;
   - the thesis's kind: practical, factual, evaluative or classificatory.
1. **Does the claim give a reason?** If not, what is it?
   - a thesis or proposal: record its conclusion form and scheme family;
   - a bare premise of another argument: a fragment of that argument;
   - a clarification, a question, a meta-comment.
2. **What does it target?** The parent's conclusion, a premise, the inference, or another claim (re-parented).
3. **Which family, then which scheme?** Walk Walton and Macagno's distinctions first:
   - Does the argument rest on a source?
   - If not, is it about what to do, or about what is so?
   - If about what is so, does it find a rule or entity, or apply one to a case?

   Then use the catalogue's contrasts. Choose the primary instance by a stated rule.
4. **What does the text state, and what is supplied?** This decides the fit.
5. **Role, polarity, critical question.** Which CQ comes from the catalogue, by its label.
6. **Evidence.**

Each step has examples from the catalogue. Where two answers are equally right, the guidelines say so, and the gold records both (§5).

### "None", the thesis, and "Other"

- **"None" is split into separate codes:**
  - *no reason given*: a thesis, proposal or bare assertion; it gets its conclusion formula and scheme family;
  - *fragment*: a stated premise of another argument;
  - *not an argument*: a clarification, question or meta-comment.
- **The thesis gets an explicit implied argument.** "The practical reasoning for the thesis" is a node that bare pros complete and that claims like 13292.13 can target. Nothing unstated is filled in.
- **"Other" must justify itself.**
  - It is coded as a defeasible modus ponens with its rule written out ("as a rule, if P then Q"). It then has two critical questions: what backs the rule, and is this case an exception.
  - It names the two or three closest catalogue schemes and why each fails, and proposes:
    - a pattern with variables;
    - a conclusion form;
    - the critical questions the pattern would face.

  Proposals go to a candidates file. A candidate seen in at least three claims of at least two discussions goes to a person, who decides one of four outcomes:
  - reject it;
  - merge it into an existing scheme;
  - record it as a domain version of one;
  - add it as a derived scheme.

### The prompt and the answer

- **Two steps per claim, so a call is not 41,000 tokens:**
  1. The claim, its context, and the catalogue's index (one line per scheme: name, conclusion form, a cue). The model names up to five candidates, or "other".
  2. The full entries of those candidates (CQs, contrasts, examples). The model codes the claim.

  Stage 1's recall is measured on its own. A single large prompt (the whole catalogue, cached) is the comparison, not the default.
- **The model chooses; the code writes.**
  - The answer gives the scheme, the variable bindings, which premises the text states, the target and the role.
  - The code builds the formulas from `walton.yml` and checks the answer before saving it.
- **Constrained output.** A JSON schema with the catalogue's scheme ids, roles and CQ labels as enums. The model cannot name a scheme or a CQ that does not exist.
- **Fewer, cheaper calls:**
  - several sibling claims per call where the provider limits requests per day;
  - context caching for the fixed part;
  - the thinking budget as a measured setting.

**Done when:** version 4 beats version 3 on the development set (§5) with the same model, and the test set confirms it.

## 4. Phase 3: data

1. **A person reviews batch 2.** Start with the claims where annotators A and B, the gold, or a model disagree. The dashboard's verdicts are the record. This is the first gold any person stands behind.
2. **Re-code batch 2 under version 4 and the catalogue.** Gold files are versioned with the guidelines they follow.
3. **More discussions, of other kinds:** empirical and causal, ethical, heavy on expert testimony, legal.
   - Four to nine discussions of about 50 claims each, sampled across depth.
   - The dashboard already makes a dataset from a captured Kialo discussion.
4. **A probe set.** For each tier-1 scheme, two short claims that instantiate it and one near miss, written and verified. It measures recognition per scheme, cheaply, and it catches regressions.

## 5. Phase 4: evaluation and debugging

| Practice | How |
|---|---|
| Acceptable answers | The gold lists, per field, the answers the guidelines allow. Scores report exact and acceptable agreement. |
| A cause for every disagreement | Model error, guidelines gap, both allowed, catalogue gap, gold error, or data. The dashboard verdict gets these as choices, so the record builds up while reviewing. |
| Development and test sets | Prompts are tuned on the development claims only. The test claims are scored at the end, once per version. |
| One change at a time | Same claims, same model, one difference. `run.json` records the inputs' and the prompt's hashes. |
| Re-scoring for free | `calls.jsonl` holds every request and answer. A new checker or metric is re-run without calls; the replay provider re-runs the pipeline on recorded answers. |
| Per-scheme results | A confusion matrix of schemes, agreement at family level, recall on the probe set. |
| Validity | The share of answers the checker accepts, before and after one repair round. |
| Several models | At least two providers annotate, and a third reviews. A model never reviews itself. |
| Cost | Tokens, calls and seconds per claim, from the traces. |

## 6. Phase 5: back to the game

- **Data from the catalogue.**
  - The NASCQ's schemes and CQs are generated from the catalogue.
  - The game's 24 critical questions, already in our own words, are checked against the catalogue's.
- **Derived schemes join the game** once they are verified, with their ask and backup edges linted like the rest.
- **Model players use the same catalogue** for the schemes they may play.
- **Dialogue types:** Track D (§8). The presets were renamed on 2026-10-15: "Policy question" and "Scientific question", since the game does not play deliberation or inquiry.
- **"Other" in the game as a defeasible modus ponens,** so that its arguments can be asked for backing and exceptions.

## 7. Decisions the owner needs to make

1. **Book text in this public repository.**
   - **`docs/The New Dialectic.pdf`, the whole of Walton's 1998 book,** was added and then deleted on 2026-10-15. It is still in the history (commit f891090) until the history is rewritten ([part 9 §5](review-2026-10-15-part9.md#5-the-book-in-the-repository)).
   - `docs/book-catalogue/` transcribes the scheme book's pp. 1–138 as printed: 75 scheme entries and 75 distinct critical questions.
   - `docs/walton-audit/cq_drafts.json` quotes 8 critical questions of the compendium.
   - The game's data uses our own wording. `walton.yml` has a few quotations from its upstream authors.

   The options:
   - make the repository private;
   - or replace the quotations with labels, page references and paraphrases.

   Removing them from the history means rewriting the history and force-pushing, which only the owner should decide. (This is not legal advice.)
2. **A licence for WAG.** The repository has none, so others may not reuse it. Carneades and `walton.yml` are MPL-2.0.
3. **Who verifies the catalogue,** and whether tier 1 above is the right first set.
4. **Paid model access.** The free tier's 20 requests a day are one annotator on 20 claims. A batch on a paid tier costs cents to a few dollars per 50 claims at today's prices; check the provider's page.
5. **The "other" threshold** (three claims in two discussions) and who reviews candidates.
6. **Dialogue types (Track D, §8).** Decided by the owner on 2026-10-15:
   - every type, playable and findable;
   - a playable quarrel is needed for research, and comes last;
   - the focus stays on schemes, critical questions and the Kialo gold, so Track D waits for phase 3.

   What remains to decide:
   - how a negotiation is scored;
   - the quarrel's safeguards, and whether people play it or only model seats;
   - which corpora, under which licences.

## 8. Track D: dialogue types

**The objective,** set by the owner on 2026-10-15:
- WAG can play a game in any established dialogue type, the quarrel included.
- WAG can find every type in real discourse, negotiation and quarrels included, as it does for schemes.

**Its place.** The project's focus stays on argumentation schemes, critical questions, and a gold standard made from Kialo (phases 1–3, measured as phase 4 describes). Track D comes after them, except for the two small pieces that phases 1 and 2 need anyway (see Order). The playable quarrel is the last step of all.

**Where it is explained:**
- **What the types are:** [`dialogue-types.md`](dialogue-types.md).
- **Whether each part is feasible:** [`dialogue-types.md` §9.2](dialogue-types.md#92-can-wag-play-every-type-and-find-every-type-in-data).
- **Why this track:** [part 9](review-2026-10-15-part9.md).

### Principles

- **Definitions come from the sources.** Walton's *The New Dialectic* (1998) comes first. Walton & Krabbe's *Commitment in Dialogue* (1995) is read before any code on commitments or shifts.
- **A type is recorded only where it changes something:** a rule in the game, or a field in annotation.
- **Mixtures are normal.** A dialogue has several typed segments; shifts are events between them.
- **Persuasion is the base.** Other types come one at a time, each playable end to end and tested before the next.
- **The quarrel comes last, with safeguards.** Research needs a playable quarrel: to produce labelled quarrels, to study how other dialogues slide into one, and to test detectors. Detecting quarrels comes first (D2g); the playable quarrel (D2h) is built last, with:
  - model seats by default, and people only by their own choice;
  - fictional scenarios, with no real person or group as a target;
  - limits on content, and a moderator;
  - every transcript labelled as synthetic.

### Steps

| Step | What | Needs | Effort | Done when |
|---|---|---|---|---|
| **D0. Reference** | [`dialogue-types.md`](dialogue-types.md): the types, the shared concepts, shifts, the arguments whose standing depends on the type, cues, a data model | — | Done on 2026-10-15 | A person checks its page references against the book |
| **D1. Record the type** | (a) Game settings and exports carry `dialogue_type` (persuasion for every mode) and the persuasion form (one-sided, or a dispute). (b) Step 0 of the version 4 guidelines: setting, type supposed and enacted, thesis kind, initial situation. (c) A dialogue-type catalogue: one entry per type and subtype, drafted from the reference, then verified. (d) "Proper in" on the scheme entries (§2). | Phase 1's skeleton for (c) and (d); phase 2 for (b) | 1–2 sessions | Exports carry the type; the guidelines have step 0; the six type entries pass a checker |
| **D2a. Persuasion with commitments** | **Commitment stores** with assert, concede, retract and challenge. **Rules:** rule 1 as an option (premises must be the respondent's commitments, or be defended when challenged); the retraction rules. **A dispute variant:** both sides must prove. **Stages:** an opening that agrees the issue and the standard, and a closing. **Variants:** permissive and rigorous as settings. **Model players** get hidden positions, so that the maieutic outcome can be measured. The text protocol's next version carries all of this. | Walton & Krabbe 1995 read | 1 week | Existing games replay unchanged; each rule has a test; a model game is played under each variant |
| **D2b. Deliberation** | **Opening:** a practical problem and a shared goal. **Roles:** symmetric. **Proposals:** one in the simplest case, then several, as the positions of a Carneades issue (the adapter handles two today). **Outcome:** a joint, tentative decision. **A caution setting:** a presumption against a proposal with a possible serious harm. | D2a | 1 week | A three-option game between models; a Kialo practical thesis can be opened as a deliberation |
| **D2c. Inquiry** | **One team. No retraction. Sources required.** **Stages:** data, analysis, presentation. **Outcome:** proven, disproven, or open. **Export:** the result as an ordered proof. | D2a | Days | Two models cooperate to prove a claim, or fail; a retraction is refused with its reason |
| **D2d. Information-seeking** | **An answerer seat that holds knowledge:** a document, a scenario card, or a model. **Questions** with presuppositions; direct and indirect answers; challenging a presupposition. **Expert consultation,** probed through the expert-opinion critical questions. | D2a | 1 week | A questioner finds the facts on a hidden card; a loaded question is flagged |
| **D2e. Negotiation** | **The bargaining core:** issues and options, private interests and reservation values. **Moves:** offers, counter-offers, accept, reject; promises and threats. **Duties:** Donohue's response rules. **Arguments for offers** go through the schemes. **Deadlock,** and a shift to information-seeking. Distributive and integrative subtypes first. | D2a; the owner's decision on scoring | 2–3 weeks | Two models reach a deal or fail; the outcome is scored; a threat inside a persuasion game is flagged |
| **D2f. Shifts and embedding** | **A stack of dialogues:** a subdialogue opened by agreement; its results carried back when it is functionally joined. **A flag** on a move of another type made without an agreed shift. | Two of D2b–e | 1 week | A deliberation embeds an information-seeking subdialogue; an unagreed threat is flagged as an illicit shift |
| **D2g. Eristic, detection** | "Spot the shift": on real or recorded dialogues, mark where a quarrel begins and which move started it; scored against a gold | D3's gold | Days | It runs on 20 dialogues with a gold |
| **D2h. Eristic, playable (last)** | **A quarrel protocol after ND ch. 7.** It opens from a grievance in a scenario card. **Moves:** blame, counter-blame ("you too"), personal attack, straw man, change of topic, concession, reconciliation. **It ends** when the grievance has been aired (catharsis) or the quarrel is broken off. **Subtypes as settings:** natural, staged, group, sophistical. **A research variant:** a persuasion game in which one seat tries to provoke a quarrel and the other tries to keep to the issue. **Output:** transcripts labelled with their segments and shifts. The safeguards are in the principles above. | D2f, D2g, D3 | 1–2 weeks | Two model seats play a staged quarrel from a scenario card; D3's detector finds its shift points; every transcript passes the content limits |
| **D3. Identification** | **Guidelines:** Walton's questions, the cues, the cautions (reference §8). **Model annotators and a reviewer** on dialogues and segments: a new task in the annotation package. **People verify.** **Agreement:** kappa on the type; segment boundaries; shifts found; licit or not. | D1 | 1–2 weeks of code, then data work | Agreement measured on at least 30 dialogues across at least four types; a baseline classifier reported |
| **D4. Corpora** | At least one source per type (below) | The owner's decision on licences | Ongoing | Each of the six types has a source with at least 20 annotated dialogues |

### Order

Set by the owner on 2026-10-15: schemes, critical questions and the Kialo gold first.

1. **D0** is done.
2. **Only what phases 1 and 2 need anyway:**
   - **"Proper in" (D1 d)** is a field of the scheme catalogue (phase 1). It is drafted from the reference's §7.
   - **Step 0 of the guidelines (D1 b)** goes with phase 2. It is a few fields filled once per Kialo discussion.
3. **After the Kialo gold (phase 3):**
   1. D1 (a) and (c);
   2. D2a to D2f, in the order shown;
   3. D3 and D4, starting from Kialo's own data (persuasion, much of it about action), then other corpora.
4. **Last:** D2g, then D2h, the playable quarrel.

The scheme catalogue stays the first priority, because every type that argues uses it.

### Candidate corpora

None of these has been checked here for licence or fit.

| Type | Candidates |
|---|---|
| Persuasion | Kialo (in hand); ChangeMyView (Tan et al., 2016) |
| Deliberation | Meeting and town-hall transcripts (Kialo's practical theses carry deliberation's arguments, but are persuasion about action); DeliData (Karadzhov et al., 2023), group problem-solving that is close to inquiry |
| Inquiry | Peer-review and replication discussions; public inquiry reports |
| Information-seeking | QuAC (Choi et al., 2018); interview transcripts |
| Negotiation | Deal or No Deal (Lewis et al., 2017); CraigslistBargain (He et al., 2018); CaSiNo (Chawla et al., 2021) |
| Eristic | Conversations Gone Awry (Zhang et al., 2018); parliamentary question periods (Hansard) |
| Mixed | AIFdb's annotated debates (for example the Moral Maze radio debates, and the 2016 US presidential debates); Question Period |
