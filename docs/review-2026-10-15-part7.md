# Part 7 (2026-10-15): the first real annotation run, reviewed

The owner ran the dashboard's annotation with Gemini 2.5 Flash on the free tier and sent the two run folders. This part covers four things:
- what happened in the runs, and why the optional review did nothing;
- why the page kept jumping to the top;
- how the 14 annotations compare with the gold, claim by claim, and what the differences mean;
- the owner's questions on prompting, on "none", on a verified catalogue with an "Other" option, on evaluation, and on whether the data is enough.

The plan that follows from it is [`plan.md`](plan.md). The central [`README.md`](../README.md) sums up where the project stands.

## 1. The runs

| Run | Claims | Annotator | Reviewer | Calls | Annotated | Failed |
|---|---|---|---|---|---|---|
| `2026-10-15_112656` | 5 | Sara = `gemini-2.5-flash` | Sara (the same) | 5 | 5 | 0 |
| `2026-10-15_113207` | 50 | the same | the same | 24 | 14 | 2 (quota) |

- **Prompt size.** Every call sent about 41,000 input tokens: the guidelines plus five resources (section 5).
- **Thinking.** Gemini also spent 850–7,600 thinking tokens per call, and answered in 6–62 s.
- **Caching.** Eight of the 19 answered calls reused Gemini's implicit cache of the common prefix (39,846 cached tokens).

**Why the run stopped.**
1. The free tier of Gemini 2.5 Flash allows 20 requests a day (the owner's AI Studio page: RPD 21/20).
2. The two runs made 19 successful calls; the model test on the Models tab probably made the 20th.
3. The next call got HTTP 429 `RESOURCE_EXHAUSTED`, "limit: 20 … retry in 7h16m".
4. Our client still retried after two minutes, three times, and then moved on to the next claim. That made eight more refused calls in twelve minutes.

The per-minute limits (5 requests, 250,000 tokens) were reached too, but never refused a call. The calls were slower than the 7 s spacing, and about 41,000 tokens × 5 is just under 250,000.

**Why the review did nothing.**
1. Review compared annotators: it took only claims annotated by two or more, and by default only those where they differed.
2. Both runs had one annotator, so there was nothing to review.
3. The buttons were disabled, and nothing on the page said why.
4. The reviewer was also the same model as the annotator. A model checking its own output tends to agree with itself, so even a working review would have said little.

**Why the page jumped to the top.** While a run works, the Annotate page reloads its state every 2 s and rebuilt the whole page. The rebuild reset three things:
- the claims list (always);
- the page's scroll, whenever the page got shorter for a moment;
- any open section and any half-typed verdict note.

A headless browser reproduced it: in 5 s the page went from 598 px to 352 px and the claims list from 300 px to 0.

**Fixed in this commit:**
- **Polling.** The page redraws only the run list and the run's pane, and only when something changed. It keeps:
  - the scroll of the page and of the lists;
  - open sections;
  - a note being typed.

  Checked in the same headless browser: nothing moves while the run works.
- **Review with one annotator.** The reviewer checks each annotation, field by field. With one annotator the button reads *Have the reviewer check each annotation*. The page says why a review button is disabled, and warns when the reviewer is also an annotator.
- **Quota used up.** When the provider asks for more than two minutes, the call is not retried. The run stops that annotator, and the message starts with "quota used up: gemini asks to wait 7h…". `report.json` is written when every pass ends; the second run's report was stale.
- **Spacing.** The default spacing for Gemini is now 13 s (5 requests a minute), not 7 s. The docs give the October 2026 free-tier limits.
- **Checkers.** `validate.py` no longer crashes on an instance without a conclusion formula. It and the Go check treat `f(a, b)` and `f(a,b)` as the same formula.

## 2. The 14 annotations against the gold

**Fields that agree with the gold** (14 claims):

| Field | Agree | | Field | Agree |
|---|---|---|---|---|
| polarity | 14 | | role | 9 |
| evidence | 13 | | scheme | 8 |
| re_parented | 13 | | fit | 8 |
| target claim | 12 | | | |
| target part | 10 | | | |

For comparison, the two batch-2 annotators agreed on the scheme for 40 of 50 claims (0.80).

**Before reading the numbers, three cautions:**
1. **The gold is not ground truth.** Annotators A and B were two runs of one language model, and the adjudicator was the same model. No person has coded or reviewed a claim. So Gemini's score measures agreement with that model's reading.
2. **The two batch-2 annotators share a model.** Their 0.80 is an upper bound for one model agreeing with itself, not a human benchmark.
3. **Fourteen claims are a handful.** One claim moves a field by 7 points.

| Claim | Text (shortened) | Gemini | Gold | Reading |
|---|---|---|---|---|
| .1 | Governments should tax meat (thesis) | none; targets itself | none; no target | Agree on "none". Targeting itself is a prompt gap: nothing says the thesis has no target. |
| .13 | A tax would not resolve the problems | cq_answer (likelihood) on the thesis's implied argument | premise_attack on 108's reconstructed argument, re-parented | **Both are allowed.** The guidelines' "bare denials" rule permits premise_attack or cq_answer. Gemini's target is an argument the annotation never states (the thesis has none), so the check rejects it. This is a design gap (section 4). |
| .7 | A sin tax makes food unaffordable | negative consequences, *explicit*; evidence *assertion* | *reconstructed*; *definition* | **Model errors.** "Bad" is not stated, so the instance is reconstructed. Its own note says the link explains "sin tax", which the guidelines code as definition evidence. |
| .9 | Revenue lost by harming the meat industry | negative consequences, *explicit* | *reconstructed* | Model error, same as .7. |
| .108 | Meat is bad for the environment | none | fragment of practical reasoning | **Guidelines conflict.** "A bare assertion is none" (§2) against "a claim stating only S1 is a fragment of its parent's argument" (§3). Here the parent is the thesis, which has no argument. |
| .115 | The tax would combat unhealthy consumption | agree | | |
| .118 | The tax makes prices reflect real costs | positive consequences | practical reasoning (goal) | A boundary the guidelines decide by a fine rule (goal, or good outcome?). Missed, but the prompt gives no contrast between the two schemes. |
| .25 | The burden could be small | none, premise_attack | other, premise_attack | **Both are allowed.** The guidelines say "fit none or other" for bare denials. |
| .33 | Cheaper proteins exist | none, premise_attack on bad(G) | other, cq_answer (likelihood) on brings_about | Scheme: both allowed. Role: the gold's reasoning (avoidable, not mild) is subtle and defensible. |
| .180 | Meat is the easiest iron source → anaemia | negative consequences, premise_backing | negative consequences, independent_pro | Defensible either way. The claim has two Kialo parents, and Gemini missed the second (rule R7). |
| .36 | Poor communities need cheap food; food security falls | cause to effect, premise_backing | negative consequences, independent_pro | Gemini made the gold's secondary causal step primary. The gold itself names premise_backing as the alternative role. The rule for choosing the primary instance is unclear. |
| .127 | Starvation is worse than over-consumption's harms | none, cq_answer "33.2 CQ3: How bad are the consequences?" | negative consequences, independent_pro | **Invented critical question.** The book's CQ3 for consequences asks whether opposite consequences should be taken into account. The question Gemini quoted is not in the book, and could not be in the prompt. Under the real CQ3, a cq_answer reading is arguable: the gold may be the one to revisit. |
| .14 | Tariffs reduce consumption | agree on every field; omits the exception `interference(tax_meat)` | | Formal error: a walton.yml exception is missing. |
| .61 | Tobacco tax analogy | agree on every field; omits two exceptions | | Formal error, as .14. |

**What the six scheme disagreements are:**

| Kind | Claims |
|---|---|
| Both readings allowed by the guidelines | .25, .33 |
| A guidelines conflict | .108 |
| A boundary the prompt does not explain | .118 |
| No rule for which instance is primary | .36 |
| An invented critical question, where the gold is arguable | .127 |

**Clear model errors** are of other kinds:
- fit and evidence on .7 and .9;
- three missing exceptions (.14, .36, .61);
- one invented CQ (.127);
- one missed second parent (.180).

**Validation.** `validate.py` reported 19 problems:
- One is an artifact: the run checked 14 of the 50 claims, which fails the order check.
- Five were only spaces after commas in formulas. The checkers now ignore spaces.
- Thirteen are real:

  | Problem | Count | Claims |
  |---|---|---|
  | missing walton.yml exceptions | 4 | .14, .36 (twice), .61 |
  | the hypothetical-cause rule R4 not applied | 3 | .14, .36 (twice) |
  | the critical-question rule R2 | 3 | .36's target has a CQ, so the claim is a cq_answer; .127 cites a CQ for `bad(G)`, which has none |
  | a target outside any stated argument | 1 | .13 |
  | the thesis targeting itself (R1) | 1 | .1 |
  | a second Kialo parent left out (R7) | 1 | .180 |

**What this says.** The disagreements come mostly from the task design, not from the model:
- guidelines that allow two answers or are silent;
- a scheme collection without critical questions, contrasts or examples;
- formulas copied by hand.

A stronger model would not fix these. Better instructions and better data would.

## 3. The questions

### Why are some claims "none", even the thesis?

In the guidelines, "none" means "this text gives no reason". The gold says the same about the thesis, so Gemini and the gold agree. The thesis "Governments should put a tax on meat" states a conclusion without a reason. By itself it is not an argument.

But "none" is the wrong label for it, and the owner's instinct is right in substance:
- **The thesis is a conclusion form.** It has the form `should_be_performed(tax_meat)`, the conclusion of the practical-reasoning family.
- **Its pros and cons supply the arguments.** The pros give the premises of that family (goals, consequences, values); the cons attack them.
- **Coding it as a full instance would invent premises.** The text does not state them, and the guidelines forbid that (R3). So the fix is not to make the thesis an instance.
- **What to record instead.** Give the thesis its conclusion formula and its scheme family. Create an explicit node, "the practical reasoning for the thesis", that its bare pros (like .108) are fragments of, and that claims like .13 can target.

"None" also covered three other things in this run:
- **.25, .33:** reasons that fit no catalogue scheme. They should be `other`, but the guidelines allow either.
- **.108:** a bare premise of the thesis's implicit argument, which should be a fragment.
- **.127:** a weighing of consequences.

`plan.md` §3 splits "none" into separate codes.

### Is the model given every scheme, with its pattern and examples?

No.

**What the prompt has** (132,565 characters, about 41,000 tokens):

| Part | Characters | What it is |
|---|---|---|
| guidelines | 13,400 | Version 2, with the version 3 changes appended as a list of edits (§5), not merged |
| `walton.yml` | 84,000 | 111 schemes: formulas and English templates. Critical questions appear only where Carneades encodes them as assumptions or exceptions: 54 of 111 schemes have any. |
| `lexicon.json` | 9,600 | The shared atoms |
| `schemes.json`, `critical_questions.json`, `edges.json` | 25,600 | The game's 9 schemes and 24 critical questions |

**What it does not have:**
- the book's critical questions (153 for 109 schemes), or which formula each is about (`links.json`, 197 links);
- a description of any scheme in plain words;
- any example, and any contrast between close schemes (consequences against practical reasoning, cause to effect against sign);
- the scheme families;
- any coded claim to imitate.

**Consequences:**
- The model sees 57 schemes with no critical questions at all.
- It is told to cite the book's CQ labels, which it has never seen; it made one up.
- The guidelines contradict themselves in places, and a list of edits at the end is easy to miss.

### Should we give the model a verified catalogue, with an "Other" option?

Yes. It is the right next step, with four refinements:
1. **One catalogue, three sources aligned.**
   - The book decides what a scheme says and its critical questions.
   - `walton.yml` decides the formulas.
   - The game's NASCQ becomes a part of the catalogue, generated from it, not a separate truth.
   - Where the sources disagree, the catalogue records the disagreement and the decision.
2. **"Other" is structured, and earned.** An "other" answer must:
   - name the closest catalogue schemes and why each fails;
   - give a pattern with variables and the critical questions it would face.

   Candidates go to a queue. One that recurs across claims and discussions becomes a *derived* scheme after a person reviews it. This keeps the rule "no new schemes for one-off patterns".
3. **Do not put the whole catalogue in every call.** With critical questions, contrasts and examples, it would be well over 100,000 tokens per call. Two smaller steps:
   - pick candidates from a one-line-per-scheme index;
   - then code the claim against the full entries of a few candidates.

   Whether one large prompt does better can then be measured rather than assumed.
4. **The model chooses; code writes.** The model chooses the scheme, binds the variables, and says which premises the text states. The code writes the formulas from `walton.yml`. That removes the copy errors (spaces, missing exceptions) and shortens the answers.

**Time.**
- **Drafting.** The skeleton can be generated from what exists (`walton.yml`, `links.json`, the NASCQ, the families). A language model can draft the rest in a few sessions: paraphrased critical questions, descriptions, contrasts, examples.
- **Verification is the slow part.** It must be done by a person with the book:
  - about 20–30 minutes a scheme;
  - roughly 8–10 hours for the 20 schemes that policy debates use most;
  - 35–55 hours for all 109.

  Start with those 20; let the data decide the order of the rest. The batch-2 gold uses 13 schemes.

### How to debug, test, evaluate and refine

The full plan is [`plan.md`](plan.md) §5. In short:
- **Score against acceptable answers, not one.** The gold records the alternatives the guidelines allow.
- **Tag every disagreement with its cause:** model error, guidelines gap, both allowed, catalogue gap, gold error, data. The verdicts in the dashboard become the record.
- **Tune on a development set; report on a held-out test set.** Fifty claims are easy to over-fit a prompt to.
- **Change one thing at a time** on the same claims and model (ablations). `run.json` already records the prompt's hash.
- **Re-score without calls.** `calls.jsonl` holds every request and answer, so a checker or metric can be changed and re-run for free.
- **Use a scheme probe set:** two short positive examples and one near miss per scheme. It measures recognition scheme by scheme, cheaply.
- **Use different models.** At least two providers annotate, and a third reviews.

### Is the annotation data enough?

**For debugging the pipeline, yes. For evaluating models or shaping the catalogue, no.** Today it is:
- one discussion, on one policy topic;
- 50 claims, using 13 of 111 schemes;
- a gold made by one model and its adjudication, with no person involved;
- coded under guidelines that changed during the adjudication.

**What it needs:**
1. A person's review of the 50 claims, starting with the 20 where the batch-2 annotators or Gemini disagreed with the gold. This is the first real gold.
2. Re-coding after the catalogue and guidelines version 4 (the gold is versioned with the guidelines).
3. Four to nine more discussions of other kinds, about 50 claims each:
   - empirical and causal;
   - ethical;
   - heavy on expert testimony;
   - legal.
4. The probe set.
