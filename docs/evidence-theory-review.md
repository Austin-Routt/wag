# Evidence theory and WAG's next implementation slice

Review date: 2026-11-22. Inspected baseline: `6375ac2b29d04cc5a5602c10cd33bcb247dd3bbc`.

This synthesis combines three independent AI source reviews of seven supplied works. It supports the five-record design in [evidence-records.md](evidence-records.md), with stronger identity, scope and assessment contracts. The accompanying native Go offline tool targets the **provisional v0.1** [evidence contract](evidence-contract.md). Its result concerns record integrity and procedural consideration. Native game/UI/capture-client integration, factual verification and empirical validation remain separate work. This increment changes no scheme formulas.

The [source registry](evidence-theory-sources.json) records exact PDF hashes, bibliographic/version caveats, printed/PDF locators, reviewer reading scopes and limitations. The books were read selectively; no full-book review is claimed. Papers and books remain outside this repository under D043. This document contains paraphrases and locators, not reproduced passages.

## 1. What the sources contribute

| ID | Supplied work | Scope and use |
|---|---|---|
| E01 | Dundee Argumentation Research Group, *The Argument Interchange Format (AIF) Specification*, 8 November 2011 | All 12 supplied pages reviewed. An abstract interchange ontology; not a complete evidence or current AIFdb specification. |
| E02 | Douglas Walton, *Argument Evaluation and Evidence*, Springer, copyright 2016 | Targeted chapters on evidence, Carneades, inquiry and evaluation. Describes pre-version-4 Carneades and acknowledges incomplete evaluation methods. |
| E03 | Douglas Walton, *Character Evidence: An Abductive Theory*, Springer, 2006 | Selected sections on explanations, alternative accounts, character and source attacks. |
| E04 | Johannes P. Wallner, *Computational Argumentation: Reasoning, Dynamics, and Supporting Explainability*, IJCAI-24, pp.8583–8588 | All six pages reviewed; an overview, not an audit of every cited theorem or implementation. |
| E05 | Douglas Walton, *Legal Argumentation and Evidence*, Penn State University Press, 2002 | Selected sections on relevance, credibility, proof and evidential networks. A historical theoretical account, not current legal guidance. |
| E06 | Douglas Walton, *When Expert Opinion Evidence Goes Wrong*, 2019 | Supplied 28-page manuscript reviewed. Its pagination and some editorial details differ from the publisher record for volume27, pp.369–401. |
| E07 | Douglas Walton, *Witness Testimony Evidence*, Cambridge University Press | Selected testimony, corroboration and account sections. Supplied front matter gives print2007 and copyright2008; preserve that distinction and the publisher metadata caveat. |

### Evidence connects material to a target through an inference

E05 distinguishes the evidential basis, the inference drawn and its use for an issue in a dialogue. E02 distinguishes broad uses of evidence from discipline-specific criteria. The resulting design principle is relational: a source is offered **for a particular proposition, under a particular warrant and policy**. A URL or file does not establish every assertion it contains. [E05 pp.200–205/PDF218–223; E02 pp.252–254/PDF263–265.]

Keep these steps distinguishable: the bytes contain a passage; a source asserts a proposition; its observation or report is credible; the assertion supports a premise; that premise contributes to an argument. A recording can substantiate that someone made an allegation while leaving the allegation unsubstantiated. Definition and context links need their own roles.

### Integrity, credibility, relevance and sufficient support differ

The books distinguish relevance, admissibility, weight and sufficiency, and separate a sensor's readout from what happened in the world. A matching hash supports byte integrity relative to a recorded digest; it does not establish origin, calibration, interpretation or truth. OCR, transcription and translation are fallible derivatives. [E05 pp.19–22/PDF37–40; E02 pp.270–271/PDF281–282.]

E06 and E07 distinguish competence, domain fit, honesty, access, memory, consistency and bias. An honest witness can be mistaken; an expert in one domain may not be qualified in another. An assessment should therefore name its source/assertion, target, domain, time, criterion and reasons. It should not silently assign global credibility to every statement by a person or website. An LLM output is an attributable model output, not automatically eyewitness or expert evidence. [E06 manuscript pp.5–9/PDF5–9; E07 pp.37–44,56–61/PDF57–64,76–81.]

### Agreement is not independent corroboration

E07 discusses collusion, common error, double counting and agreeing imperfect instruments. Linked premises, separate supporting arguments and serial inference chains have different functions. Multiple citations may still derive from one observation. [E07 pp.33–37,56–58,80–87/PDF53–57,76–78,100–107; E05 pp.338–342/PDF356–360.]

Preserve common-source and transformation dependencies. Unknown independence is not established independence. Two LLM summaries of one article, or ten copies of a press release, should not acquire automatic corroboration credit. The sources motivate this distinction; they do not supply a universally validated numerical aggregation rule for WAG.

### Questioning, burdens and revision require state

E02 and E06 distinguish assumptions requiring the proponent's answer from exceptions the challenger must substantiate. A burden includes a bearer, target and standard within a procedure. An undercutter can defeat a route to a conclusion without proving its negation. [E02 pp.133–134,155/PDF144–145,166; E06 manuscript pp.6–9/PDF6–9.]

This supports per-CQ NASCQ contracts: which object is questioned, what response is required, what evidence binds to it, and what follows from silence, refusal or an inadequate answer. A missing source and an unanswered challenge are different states. E02 also permits correction and reopening in defeasible inquiry, and requires a detection/completeness warrant for reasoning from missing evidence. Failed retrieval alone cannot establish that a claim is false. [E02 pp.219,221–224,228–233/PDF230,232–235,239–244.]

### Explanations, interchange and computational evaluation have limits

E03 separates coherent narratives from evidentially anchored explanations and asks whether alternatives and further investigation remain. A generated Kialo playthrough is therefore a reconstruction, not an observed dialogue or proof of the claims. [E03 pp.52–60,167–168/PDF64–72,178–179.]

E01 distinguishes information from applications of inference, conflict and preference schemes. It permits scheme-to-scheme edges and directional conflict, including attacks on an inference. Edge roles cannot always be inferred from endpoint types. WAG's restricted AIF importer is a declared projection; it should not be described as general lossless AIF/AIFdb support. [E01 pp.2–6/PDF3–7.]

E04 separates construction of arguments, acceptability evaluation and conclusions. It also separates learning, enforcement of a desired outcome, stability under specified additions, and explanation. Winning a game or learning to produce an accepted conclusion does not validate source truth. Formal stability requires a declared semantics and range of future changes; passing tests supports only the tested claims. [E04 pp.8583–8586/PDF1–4.]

## 2. Source discrepancies and limits

The literature must be checked critically before becoming executable reference cases:

- **Historical engine boundary:** E02 p.153/PDF164 describes Carneades versions1–3 and says the next version is not implemented. It cannot specify the current vendored Carneades-4 runtime.
- **Unexecuted example:** E02 pp.169–170/PDF180–181 says its Leonardo case was not evaluated through CAS; pp.172–173/PDF183–184 acknowledges unresolved cumulative modeling. A new reconstruction needs independently stated inputs and outcomes. The artwork's actual attribution is not established by this review.
- **Probability error:** E02 p.245/PDF256 misleadingly relates the conjunction fallacy to applying Bayesian rules and invokes independence for the conjunction bound. Set inclusion gives `P(A ∩ B) ≤ P(A), P(B)` without independence. The probability of combined evidence is also distinct from its support for a hypothesis. This passage does not justify rejecting probabilistic models wholesale.
- **Question-policy qualification:** broad default-on-question statements in E02 p.131/PDF142, E06 manuscript p.6 and E07 p.61/PDF81 require comparison with differentiated assumption/exception treatment. Do not encode one universal question effect.
- **Unexplained bean color:** E03 p.58/PDF70 places a red bean near a bag described as containing white beans; p.59/PDF71 does not reconcile this color mismatch or explicitly change the individual bean to white. Treat this as a likely textual error or an unaddressed exception. A fixture assuming an all-white bag must treat the red observation as contrary evidence; label any corrected white-bean fixture separately.
- **Manuscript diagram errors:** E06 pp.7–8 confuses the expertise-supporting PhD evidence with the reliability exception, and p.23 confuses identifiers and figure numbering. The supplied figures and prose need reconciliation before executable translation. No claim is made that the unchecked publisher PDF has identical errors. Its illustrative case and incomplete dialogue record do not establish measured accuracy. [E06 pp.23–26.]
- **AIF definition/prose mismatch:** E01's definition requires exactly one conclusion-role successor, while subsequent prose says at most one. Its local predecessor/successor condition also cannot alone exclude a scheme-only cycle. Adopt explicit profile constraints and test them; do not assert the stronger global property follows from the local rule. [E01 pp.3–4/PDF4–5.]

These findings limit particular claims; they do not invalidate the entire works. Case studies, philosophical theories, specifications and empirical validation have different evidential roles.

## 3. WAG design decisions derived from the review

The following are **WAG engineering recommendations**, not a schema dictated by any one paper. Details belong in [the contract](evidence-contract.md).

1. **Keep five record families.** Preserve artifacts, derivatives, bindings, assessments and events. Distinguish retrieval identity from shared blob storage; identical bytes can have different provenance.
2. **Pin the binding, not just the document.** Consent identifies the exact interpreted claim, representation/selector, target, polarity/role, game context and policy revision. A changed interpretation or target cannot inherit consent because the raw PDF is unchanged.
3. **Make consideration procedural.** The first offline reducer records two distinct participants' consent and withdrawal for consideration only. It does not stipulate a proposition, verify facts or change engine labels/weights. Moderator authority and ordinary assessment are different operations.
4. **Preserve revision and uncertainty.** Failed, partial, unsupported and unattempted captures can be valid records. They are not successful acquisitions or negative findings. Withdrawal removes its own route; history and other support remain.
5. **Preserve assessments and dependencies.** Record who assessed which version, what they checked against, the result and limitations. Avoid a whole-document verified flag, global source score or automatic numerical aggregation.
6. **Keep source content untrusted.** Quoting a page alone does not prove instruction isolation. The offline validator must not fetch arbitrary URLs or execute source content; future player/tool boundaries require their own enforcement and tests.

At the inspected baseline, the game's evidence requirement still checks nonempty citation strings and stores metadata (`wag/internal/game/turn.go`). The strict citation parser's rejection of non-text sources is useful, but it is not evidence validation. The v0.1 offline tool is a new bounded capability, not retroactive validation of old games.

Kialo's current import is a selected graph projection. Preserve raw captures and citation-extraction provenance alongside it; do not infer real speakers, chronological turns, assent or source-content coverage from reconstructed pro/con roles. Existing partial move digests are not comprehensive evidence or player-observation digests.

No formulas are changed by this evidence increment. Existing decided work on question wording, witness-domain direction and account membership (D009, D034, D035) remains subject to its normal proposal, review and integration process.

## 4. Verification now; empirical validation next

The v0.1 acceptance boundary is **structural/integrity and procedural correctness**, with independently specified positive and adverse cases. Passing it means valid for the declared record contract, not factually verified. Failed acquisition and unresolved assessment must remain representable.

| Contract test | Expected distinction |
|---|---|
| Duplicate keys/IDs, broken references, source/derivative cycle | Refuse malformed or inconsistent records. |
| Wrong bytes, length or unfetched LFS pointer | Refuse the purported payload; preserve an honestly failed acquisition record. |
| Same artifact, changed target/derivative/context/policy | No inherited consent. |
| One actor repeats consent; another remains silent | No bilateral consideration outcome. |
| Consent withdrawal and replay | Remove only the affected route; deterministic active state and retained history. |
| Positive assessment without authorized consent | Assessment does not grant admission or factual truth. |
| Source saysP versusP; context/definition versus support | Distinct interpretation and binding roles. |

These are required test categories, not a claim that this document itself executed them. Use the implementation's validation receipts for actual results. Later runtime integration must separately test CQ effects, argument/statement targets, hard-assumption avoidance, dependent recomputation and preservation of independent support.

Empirical evaluation needs a bounded, stratified pilot with independently produced annotations and held-out sources. Include original and derivative media, inaccessible captures, misleading paraphrases, conflicting witnesses, common-source copies and unresolved cases. Measure extraction fidelity, relevance, support, procedural behavior and agreement separately. Majority vote is an agreement signal, not the sole correctness reference. Required human verification under D045 remains required.

Meta-arguments can retain public reasons about an annotation or evidential judgment: exact target/version, rule, source region, assumptions and defeaters. They organize review but cannot certify themselves through mutual endorsement. E04's faithful/spurious abstraction distinction supports checking that a compressed explanation preserves the relevant result; a fluent rationale is not a proof certificate or a record of private model reasoning. [E04 p.8586/PDF4.]

## 5. Scope and next steps

Finish and independently review the offline contract first, then connect a bounded native-game profile and run the small pilot. Expand aggregation, narrative reasoning, interchange and richer dialogue only through declared mappings and measured failures. Avoid treating a growing check count as comprehensive framework validation.

D077's private research-retention instruction, D078's history decision, D079's reserved questions and D045's human-verification requirement remain unchanged. The supplied theory establishes no blanket rights conclusion. The repository's integration record documents earlier reviewer overreach in overriding decisions and claiming unimplemented controls; this review does not repeat those status or authority claims. The offline tool imposes no new capture gate. Source integrity, procedural consent, human review, implementation status and empirical correctness remain distinct records.
