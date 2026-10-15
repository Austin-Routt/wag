# Part 6 (2026-10-15): the audit reviewed, the LLM layer, the evidence for the schemes

This part answers three things:
- the third-party audit of 2026-10-13, *Critical audit of WAG, NASCQ, Carneades-4, scheme data, game logic, and LLM research infrastructure*;
- concern 1: Gemini, and a general LLM interface with complete traces;
- concern 2: whether the Kialo-derived schemes are validated and integrated.

| Question | Answer | Where |
|---|---|---|
| Is the audit accurate? | Mostly. Its main finding is right, and so are most of its specific findings, which are fixed now. It has one factual error, misses one mechanism, and is half wrong about one edge. | §1 |
| Is it coherent? Does the author understand the system? | Yes, and yes. The code paths it describes are the real ones. | §1.3 |
| Are its suggestions good? | Most. Adopted now: 16. Waiting for your decision: 5. Changed: 3. | §1.4 |
| Do I need your API key? | **No.** Put it in `GEMINI_API_KEY` in your own shell. Never paste it in a chat, a file or a flag. | §2 |
| Gemini, and the other providers | Built: Gemini (Google's GenAI SDK), Claude, OpenAI, DeepSeek, LM Studio, llama.cpp, SGLang, vLLM, Ollama, any command (Guidance, DSPy), and a mock. One interface (`internal/llm`); full traces (`wag-trace/2`); headless games, replay, reports, annotation. | §2, `docs/llm.md` |
| Are the Kialo schemes validated and integrated? | They load, and every scheme–question link behaves as its locus says (153 of 153 answers). Two model annotators agree well on 50 claims. But all 50 come from one discussion, no person has reviewed them, and only 1 of 6 intended question answers could be played. They are **not validated**. | §3 |
| Can Gemini help validate them? | Yes, as one more annotator and as a player whose mistakes are informative. It cannot replace human review. | §5 |
| Without introducing errors? | Nobody can promise that. I can say what was checked, and what changed on purpose. | §6 |

## 1. The audit

### 1.1 Accurate, and now fixed

| The audit's finding | Checked against the code | What changed |
|---|---|---|
| §4: a backup edge silently stood for the link between a scheme's conclusion and the answer. | **True.** `fitAnswer` returned no error when the answer was not of the scheme's form. Its comment said the edge "stands for the link". | An answer must now be of its backing scheme's conclusion form (S9), unless the edge **declares a bridge** (`"bridge": "why"`). The move then records the bridge. The shipped data declares none. |
| §5.6: value arguments should not back "other goals" or "legitimate value". | **True.** `reason_retaining_goal(V,G)` is neither answer. The "legitimate value" backup was also circular. | Both edges are removed. If someone adds them back, `wag-cli --lint` reports an error: "S9 rejects every such answer". |
| §5.2: analogy should not back every question. | **True**, because of its formula. As the answer to any question, it says the answer "held in a similar case". | Analogy is no longer thesis-independent. It still argues for or against statements. N12 is pending (§4). |
| §5.3: the literal yes/no is easy to misread. | **Truer than the audit knew. This was a bug.** The browser's buttons read "Attack: answer no" for every question. For 9 of the 24 questions ("bad side effects?", "another means?", "other goals first?" …), an attack answers **yes**. | Each question now has `answer_en` (the literal answer against the argument, and the one for it). The buttons say it ("Attack: answer “yes”"), and so does the prompt. |
| §5.4: "outweigh" overstates. | **True of the wording.** | The question now reads "Would doing A also have good consequences, to weigh against G?" (but see §1.2). |
| §8.1: my earlier review overgeneralized about CCE and BRD. | **True.** | Corrected in `review-2026-10-13.md`. The claim holds between stock schemes, not in general. |
| §9: the traces were not the full exchange. A Claude retry stored only the rejection; there were no usage, latency, request ID or finish reason, no event IDs, and calls used `context.Background()`. | **All true.** | Trace v2 (§2): every call holds the whole conversation, the parameters, usage, latency, finish reason and IDs. Decisions keep the board on every attempt. Calls follow the server's lifetime and a per-call timeout. |
| §3: the docs say 9 schemes, but `core.yml` has 10. | **True.** `credible_source` is kept for old games. | Docs say "9 NASCQ schemes, 10 Carneades schemes", and why. |
| §3, §6.6: the check outputs name files that do not exist, and the cleanup tooling was not committed. | **True.** | The tooling is in `walton-clean/tools` and `walton-clean/check`. `check/run.sh` rebuilds `walton.yml`, `changes.json` and `walton.proposals.yml` **byte for byte** from the upstream file (SHA-256 checked), and rewrites the reports with the right names. |
| §10: `validate.py` needs a compendium file kept outside the repository. | **True.** | It reads `links.json` instead (labels only) and runs from a clean checkout. It also checks any annotator's file. |
| §10: the converter's fallbacks hide failures. | **True.** | Each claim records `intended` and `executed`, and an integration category (§3.3). |
| §9: the free tier's terms. | Relayed, not re-verified by me. | `docs/llm.md`, and the tools print it when a Gemini profile is used. |

### 1.2 Inaccurate, or missing a mechanism

1. **"Human adjudication is documented" (§10) is false.** The adjudicator was a third language-model agent. Our README did not say who adjudicated, which invited the error. It does now, and it says no person has reviewed the gold set.
2. **The answer variable already does much of what §5.2 and §6.1 ask.** For expert opinion, position to know and sign, the game binds the answer into the premises: `asserts(W, ¬possible(tax_meat))`, and "finding A indicates that the answer is true". So "the source must assert the actual answer proposition" is already true, by construction. "The domain must fit" is a premise with its own question (*Field*). Three of the six backup edges (sign's) were never the problem. The audit counts all six as unverified.
3. **"Consequences backing Opposite Consequences: reject" (§5.4, §13) is half right.**
   - Attacking a negative-consequences argument through that question concludes `should_be_performed(A)`. That is exactly positive consequences' conclusion, about the same A (S9 fixes it). Whether the good outweighs the bad is not asserted by the argument. Carneades decides it by weighing both sides under the proof standard. This direction is sound: "doing A brings about cleaner air; cleaner air is good; so A should be done" rebuts "A brings about higher bills; so A should not be done".
   - The defence direction concluded `¬should_be_performed(A)` from positive consequences. That was unsound.
   - The strict rule keeps the first and refuses the second. `nascq-edges.md` shows both.
4. **N4 (§5.5, §7.1) is split by the same rule.** "Side effects" is an exception of practical reasoning (`side_effects(A,S1)`), where negative consequences cannot conclude the answer. It is also the conclusion question of positive consequences, where an attack concludes `¬should_be_performed(A)`: exactly what negative consequences concludes. A bare N4 edge would be sound at the conclusion and unsound at the exception. Strict S9 would now enforce that split, so the audit's "reject a bare edge" and its acceptance gate ("unify, or name a bridge") are one rule. I have not added the edge: a statement move against `should_be_performed(A)` already makes the same argument.
5. **It could not run the tests.** That was its environment, which had Go 1.23. Here every check runs (§6).
6. **The phase order (§12) is stricter than needed.** It asks to freeze the semantics before any provider work. But every trace records the scheme data's hash, the protocol and the build, so games on different data can always be told apart and filtered. The semantic fixes that change what gets recorded came first anyway, in this round.
7. **"Deterministic replay" needs a definition.** A model's answers cannot be reproduced. The game, given the same answers, can. `wag-llm replay` does exactly that, and checks that every request the game sends is byte-identical to the recorded one.
8. **The SDK version.** The audit cites v1.73.0 as the latest. v1.74.0 came out on 13 October; I pinned v1.73.0, which is a week older.

### 1.3 Coherence, and whether the author understands the system

The audit is coherent: its decisions log matches its sections. The author read the right functions (`fitAnswer`, `OwnConclusion`, S5, the Claude retry loop) and described them correctly. It is careful about evidence: it separates what it observed from what the repository reports.

Its blind spots:
- the answer variable (§1.2, item 2);
- that CAES weighing, not the answering argument, decides "outweighs" (item 3).

Both are about how the NASCQ hands answers to Carneades. That is the subtlest part of the design, and our documentation of it was thin.

### 1.4 Its suggestions

**Adopted now (16):**
- strict answer fit with declared bridges;
- the value backups removed;
- analogy narrowed;
- literal answers;
- the "outweigh" wording;
- trace v2;
- the provider interface;
- Gemini through the official SDK, pinned;
- keys only from the environment;
- live tests only on request;
- Claude refactored behind the interface, with the retry trace fixed;
- an OpenAI-compatible adapter for local engines;
- `validate.py` reproducible;
- the cleanup tooling committed;
- counts corrected;
- the converter's intended and executed records.

**Adopted in a changed form (3):**
- **A fixture per edge.** It is generated, not hand-written: `wag-cli --edges` plays all 153 answers and writes `nascq-edges.md`, and a test fails if one misbehaves or the file is stale.
- **Capabilities instead of `thesis_independent`.** The answer variable already is the capability "concludes any proposition"; the flag now only means "backs every question".
- **Phase order.** See §1.2, item 6.

**Waiting for your decision (5):** N12, N8, the premise repairs, the five renamed IDs, and bridge schemes for the six failed answers. See §4.

## 2. Concern 1: language models through their APIs

**You do not need to give me your key, and should not.** The code reads it from the environment, on your machine:

```sh
export GEMINI_API_KEY=…     # in your shell
cd wag && go build -o wag-llm ./cmd/wag-llm && go build -o wag-server ./cmd/wag-server
./wag-llm models gemini                                   # the models your key can use
./wag-llm ping --model gemini:gemini-2.5-flash            # one short call: text, model version, usage, latency
./wag-server --model flash=gemini:gemini-2.5-flash,min_interval=7s,output=json     # a Gemini seat in the browser
./wag-llm play --model flash=gemini:gemini-2.5-flash,min_interval=7s,output=json \
    --pro flash --con greedy --games 5 --swap --out runs/flash-vs-greedy           # headless, traced
```

`docs/llm.md` is the full guide: profiles, every provider, the free tier's privacy terms and limits, traces, testing, and what the data is and is not good for.

**What was built.**
- **`internal/llm`.** One request and response format. Adapters:
  - Gemini, through `google.golang.org/genai` v1.73.0. Its thoughts are never requested; only the thinking level or budget is configurable.
  - Claude, through Anthropic's SDK. It is the old Claude seat, moved behind the interface.
  - OpenAI-compatible endpoints: OpenAI, DeepSeek, LM Studio, llama.cpp, SGLang, vLLM, Ollama, or any base URL.
  - An external command, for Guidance, DSPy or your own harness.
  - A mock.
- **Profiles.** A file or the command line, holding the generation parameters, JSON mode, timeouts, call spacing (`min_interval`) and retries. A rate-limited call is retried after the delay the provider asks for.
- **Model seats.** A seat can be played by any profile (`--model`, `--models`), in the browser or headless. The old `claude` seat is the profile named `claude`.
- **Trace v2.** One stream per game:
  - `game_start`: everything needed to set the game up again;
  - `call`: the whole conversation sent, parameters, answer or classified error, usage, latency, IDs;
  - `decision`: the board, the reply, its stated reasoning, the move, its result and effects, predictions checked, alternatives tried;
  - `game_end`: the outcome, every final label, the verdicts.

  `docs/protocol/trace.schema.json` describes it; real trace files validate against it.
- **`wag-llm`.**
  - `play`: headless games, with a trace file, export and summary per game; `--swap` alternates sides.
  - `replay`: plays a trace again with the recorded answers, and checks that the requests and decisions are identical.
  - `report`: wins and rejected moves by seat and side.
  - `annotate`: Kialo annotation, §5.
  - `models` and `ping`.
- **The prompt now states the topic** (`--title`), so a model can be given one.

**Not yet done, on purpose.** I made no live calls: I have no key, and the calls would use your quota. Every adapter is tested through its real SDK or HTTP code against a local fake of the provider's API. The tests check the request shape; that the key travels in a header and never in a URL or trace; parsing; usage; rate limits; and blocked answers. Your first `ping` is the live test. If a model rejects a parameter (a thinking level it does not support, say), the trace shows the provider's error.

## 3. Concern 2: are the schemes validated and integrated?

### 3.1 Four levels of "validated"

| Level | Status |
|---|---|
| (a) The schemes load and instantiate. | Yes: all 111 of `walton.yml` and all 10 of `core.yml`. Checked by `check/run.sh` and the game's tests. |
| (b) The game's scheme–question links do what their locus says. | Yes: 153 of 153 offered answers conclude what the locus and the answer's polarity call for (`nascq-edges.md`, and a test). The data lint finds 0 errors and 0 warnings. |
| (c) Independent annotators apply the schemes reliably. | **In part.** Two model annotators: scheme κ = 0.77, role 0.84, fit 0.60, re-parenting 0.51. That is one discussion, with models from one family. |
| (d) The instances are correct, and the patterns hold beyond one discussion. | **No.** No person has reviewed the gold set, and there is only one topic. |

### 3.2 Per scheme (batch 2: 50 claims, one discussion)

Agreement of the gold labels with the annotators is not a measure of correctness, since gold was built from them. The column *A and B agree* counts the claims where both independent annotators chose the gold scheme.

| Scheme (`walton.yml`) | Gold claims | A and B agree | In the game | My confidence |
|---|---|---|---|---|
| `negative_consequences` | 9 (+2 secondary) | 9 | 7 as intended; 2 failed (answering "side effects") | High that it fits these claims. Medium for its question links. |
| `practical_reasoning2` (goal-based) | 8 (4 of them fragments) | 8 | 3 as intended, 3 fragments, 2 failed ("other means") | Medium. It is often confused with positive consequences: B chose it for 4 claims that A and gold call consequences. |
| `positive_consequences` | 5 | 1 | 4 as intended, 1 failed | Low to medium: the main disagreement. |
| `cause_to_effect` | 4 (+13 secondary) | 4 | Not in the game: coverage gap | Low as used. It needs the hypothetical-cause variant (N8). |
| `verbal_classification` | 3 | 0 | Not in the game | Low. |
| `rule` | 2 | 1 | Not in the game | Too few. |
| `expert_opinion` | 2 (+2) | 2 | 2 as intended | Too few to say. |
| `analogy`, `positive_value`, and 3 others | 1 each | 1 each | Analogy as intended; the value argument failed | Too few. |
| `none` / `other` | 7 / 5 | 5 / 5 | Challenges, or Other | Honest no-fits. They show where Walton's catalogue does not reach Kialo's style. |

### 3.3 Integration with Carneades, the NASCQ and WAG

Playing the gold annotations (`to_wag.py`, re-run on the new data) gives:

| Category | Claims |
|---|---|
| As intended | 20 |
| Fragments | 3 |
| Coverage gaps (scheme not in the game) | 15 |
| Failed | 6 |
| Duplicate challenge | 1 |
| Cascade | 2 |
| Not played | 2 |

The totals are unchanged by this round's data fixes (44 played). The gold game never used the two edges that were removed.

**What this shows.**
- **The question layer has almost no positive evidence from real discussions.** Of the game's 35 ask edges and 4 backup edges, the gold game used one question answer as intended ("likelihood", answered by an expert), plus three premise challenges. Kialo replies are mostly counter-arguments: moves about a statement, which work well (16 as intended).
- **All six failures are the same gap: a bridge.** The claim's own scheme concludes something other than the answer.
  - Negative consequences answering "side effects": twice.
  - Goal-based practical reasoning or consequences answering "another means is better": three times.
  - A value argument answering "good consequences to weigh against": once.

  These are exactly the cases the audit's §4 and N4 discuss. They are evidence *for* two typed bridge schemes (§4), not for silent edges.
- **Semantic types do not decide anything.** They label schemes and questions, and the lint uses them. The logical check is formula unification (S9), which the 153-answer table exercises.

## 4. Decisions for you

Each is a change of scheme content. I propose; you decide.

1. **N12, analogy between two cases.** A new scheme beside `analogy`, which stays for old games:
   - premises `similar_cases(Cs,Ct)`, `holds_in(Ps,Cs)`, `corresponds(Ps,Pt)`;
   - exceptions `relevant_differences(Cs,Ct)`, `more_on_point(Pt,Ct)`;
   - conclusion `Pt`, with answer variable `Pt`.

   It would be thesis-independent. Questions: difference, truth in the source case, correspondence, a better case.
2. **N8, a prospective cause.** `cause_to_effect_prospective`:
   - premises `would_occur(E1)`, `causes(E1,E2)`;
   - exception `interference(E1)`;
   - conclusion `would_occur(E2)`.

   This is for policies not yet enacted; `cause_to_effect` stays for observed causes. It would cover 13 secondary instances in the gold set.

   *Added 2026-10-15, after your principle that Walton's text and the Walton–Carneades collection are the standard.* The collection already has causal schemes:
   - `cause_to_effect` (the book's 28), `correlation_to_cause` (29) and `sign` (30);
   - the abductive schemes (31);
   - the slippery-slope schemes, whose steps are `causes(E1,E2)` (49, 51, 52).

   The book's scheme 28 is already modal. Its minor premise has A that "occurs (might occur)", and its conclusion has B that "will (might) occur". The collection's formula keeps only the first reading: `has_occurred(E1)`, `causes(E1,E2)` ⇒ `will_occur(E2)`. So N8 is not a new scheme. It is the book's scheme 28, formalized with the "might" the collection dropped.

   The book's CQ1 (how strong is the causal generalization) and CQ2 (the evidence for it) are questions about `causes(E1,E2)`. Only CQ3 (interference) has a formula in the collection.
3. **Two bridge schemes, for the six failures.**
   - `bad_side_effect`: `brings_about(A,G)`, `bad(G)`, `not_the_goal(G,A)` ⇒ `side_effects(A,S)`.
   - `better_means`: `brings_about(A2,G)`, `better_means(A2,A,G)` ⇒ `realize_more_effectively(G,A)`.

   Each concludes the exception it answers, so S9 accepts it without a bridge declaration. Both patterns recur (2 and 3 times), but in one discussion.
4. **The premise repairs.** The audit's line is right: `practical_necessary_condition` and `practical_sufficient_condition` should keep the other goal (`A2`) rather than drop it. `ad_hominem` can be approved as is. `regulative_rule` only as an agent-specific variant.
5. **The five renamed IDs** (`popular_practise` → `popular_practice`, and four others). They have been in the cleaned `walton.yml` since 13 October, as have the four formula repairs; the audit approves both. The alias map it asks for exists now (`walton-clean/id-migrations.json`; `validate.py` reads it), so a file with an upstream ID still reads. The gold set uses none of the old spellings. Your approval makes them final.

## 5. Gathering further evidence

1. **A model as a third annotator, now.**
   ```sh
   ./wag-llm annotate --model gemini:gemini-2.5-flash,output=json,min_interval=12s --out runs/annotate-flash
   python3 -I ../docs/kialo-batch2/agreement.py runs/annotate-flash/annotations/gemini_gemini-2.5-flash.json ../docs/kialo-batch2/gold.json
   ```
   - Every call is traced, and `run.json` records the hash of everything the model saw.
   - Field-by-field agreement with gold shows where the guidelines are clear. Disagreements are where a person should look first.
   - Each call is about 35,000–40,000 tokens. On a free tier, try `--limit 5`, and send less with `--context`.
   - Kialo discussions are public. Even so, read the free tier's terms in `docs/llm.md`.
2. **Annotators that do not share a model.** Annotators A and B were models of one family. Add a model from another provider (Gemini, or a local model) as a third. Agreement across families is better evidence than agreement within one.
3. **People.** A blind, stratified review of about 30 gold items is the evidence the system most lacks:
   - high- and low-agreement items;
   - every scheme;
   - every failure category.

   It gives the error rate of the gold set.
4. **Other discussions.** Annotate 2–3 Kialo discussions on other topics under the frozen version-3 guidelines, held out from any tuning. A scheme is evidenced when it appears, and is agreed on, across topics.
5. **Counterexamples and near-misses** for each scheme, as fixtures (the audit's Phase 2). The edge table already covers the question links.
6. **Models playing.** `wag-llm play` traces show:
   - which schemes and questions models reach for;
   - why their moves are rejected (by rule, by scheme);
   - whether their predictions of the labels are right (`expect_accuracy`), that is, whether they understand the Carneades semantics.

   Systematic rejections point at prompt or data problems; correct, strategic use of a question is positive evidence for its link.
7. **A proposed bar for calling a scheme validated.**
   - At least 10 independent instances, across at least 2 topics.
   - κ ≥ 0.6 on that scheme between annotators of different model families.
   - At least 80% of a human reviewer's sample judged correct.
   - At least 80% of instances integrated as intended.
   - Its rows of `nascq-edges.md` reviewed by a person.

## 6. Without introducing errors?

**What was run**, all on this branch:
- the full Go test suite: every package passes, including new tests for the adapters, the model seat and trace events, headless play and replay (with a tampered trace that must not replay), annotation, strict answer fit with bridges, and the 153-answer table;
- the race detector on the concurrent packages;
- `go vet` and staticcheck, both clean;
- the data lint: 0 errors, 0 warnings;
- the registry check on `core.yml`: 10 schemes, 10 distinct patterns;
- `validate.py` on gold: OK;
- the gold game re-played: same totals as before;
- the scheme cleanup rebuilt byte for byte;
- a browser check of a model seat and the new answer buttons.

**What changed on purpose**, and may surprise:
- Answers whose scheme cannot conclude them are refused.
- Analogy no longer answers questions.
- The two value backups are gone.
- The trace format is v2.
- Seat kind `model` is new (`claude` still works).
- The prompt shows the topic and each question's literal answers.

Exports and traces from before these changes record a different data hash, and the loader says so if their moves no longer replay.

**What is untested**: live calls to every provider. Those need your key, or a local server, and happen on your machine.
