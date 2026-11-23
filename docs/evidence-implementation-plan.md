# Evidence implementation and evaluation plan

Review base: `6375ac2b29d04cc5a5602c10cd33bcb247dd3bbc`.
Patch integration base: `f944f38281a5a2d67f756a5cf01b9907b600f8ee`;
the intervening commit adds API-run budgets without changing evidence records.
This refines stage 2 of [the plan](plan.md), using the
[literature review](evidence-theory-review.md). It does not re-sign or supersede
owner decisions. D045, D077, D078 and D079 remain as recorded. Capture does
not wait for a new permission gate introduced by this review.

## E0: current increment, offline integrity and procedural records

Implemented in the accompanying patch:

- A native Go `wag-evidence/0.1` record checker and CLI, with exact selectors,
  retained payload identity, transformation ancestry, attributed assertions,
  targets/bindings, assessments and dependency records.
- A deterministic two-party consent/withdrawal reducer for **consideration
  only**, with exact binding/context/source/target fingerprints.
- Strict decoding and independent adverse controls, including malformed
  Unicode, duplicate keys, missing zero-valued fields, native nonfinite values,
  traversal/symlink attempts, tampered payloads and unresolved LFS pointers.
- A bounded live reply-parser fix: ambiguous duplicate JSON member names are
  rejected before they can overwrite malformed evidence/action values.

The executable contract is [documented separately](evidence-contract.md).
The committed example is synthetic and rerun in CI. Offline passes do not
close stage 2, certify a live game, authenticate a person, judge a warrant or
establish factual truth. Current game citations retain their legacy meaning.

## E1: connect the first slice before an evidence-aware game pilot

Implement this bounded path with native fixtures before scaling API calls:

1. **Immutable source registry.** Write capture attempts and available original
   bytes, derivatives and representation IDs. Preserve failure states. Record
   the exact representation and excerpt shown to each annotator/player.
2. **Explicit evidence references.** Add a versioned action/export field for
   binding references. Keep old `evidence: [string]` as citations. Never
   silently promote a URL or a historical accepted move to captured, admitted
   or verified evidence. Identify old versus new profiles in imports/replay.
3. **Native target checks.** Resolve target IDs, arguments, roles and premises
   against the actual pinned graph and catalogue; verify actual run input
   hashes. A structurally valid sidecar cannot choose an unrelated graph or
   satisfy a different argument's requirement.
4. **Consent actions and export.** Record distinct participants' consent to the
   exact binding, allow discussion before admission, and preserve refusal,
   withdrawal and unresolved states. Define whether a move needs citation,
   consideration admission or another specified status. The current first
   policy supplies no hard assumption or factual endorsement.
5. **Source-grounded argument moves.** Where an accepted source assertion is
   used to support P, create the relevant testimony, expert, sign, observation
   or other scheme instance, with its warrant and CQ targets. Source reports
   and P remain separately challengeable. Not every artifact requires an
   expert-opinion scheme; the interpretation determines the scheme.
6. **Dependency-preserving revision.** Withdrawing one consent or defeating one
   route recomputes its consequences while preserving independent support,
   history and earlier assessments. Do not globally delete a proposition or
   mark a person untrustworthy from one scoped account criticism.
7. **Observation and replay.** Pin content/profile/engine/catalogue/NASCQ,
   the WAG build/parser and player settings. Replay the same records without refetching mutable URLs
   or asking a model again. Retain the actual source material available to a
   player, not merely an assertion that it read an entire linked document.

Acceptance cases include wrong actor/target/graph revision/polarity, stale
consent, changed OCR or source interpretation, withdrawal with a second support
route, a question about an exception, absence of an answer, invalid final
replies, and source instructions that try to become tool/policy instructions.
The controller must enforce the data boundary; prompt quotation alone is not
an acceptance test. These integration cases are **planned**, not already
executed by the offline checker.

E1's practical-reasoning slice needs evidence for empirical consequences,
causal links, values/preferences where asserted, relevant agent/action identity
and comparisons. A participant's declared preference or commitment can be the
appropriate source for that preference; it need not be external factual proof. Source scope must preserve modality, time, population,
scenario and uncertainty when those affect the inference. AS2 can add explicit
states/actions/goals/values later without replacing this provenance contract.

## E2: source-dependent families and full NASCQ contracts

Complete the already decided witness/expert changes through the normal proposal
path, including D009, D034 and D035; they are existing pending work, not new
regressions introduced by this patch. Do not replace the book's authority or
formula review procedure with examples from the supplied manuscripts.

For every supported CQ, version its target locus, burden bearer, triggering
act, answer options, evidence binding, response status and engine effect.
Questioned, unanswered, refused, unsupported, contradicted and resolved are
different states. An assumption challenge and an exception allegation may
have different effects. Silence and a failed search are not falsehood proofs.

Add paired native cases for exact testimony versus stronger paraphrase;
competence in the relevant versus another domain; reliable access versus
hearsay; authentic but false reporting; independent corroboration versus common
origin; credible conflicting accounts; bias about one account versus global
character condemnation; and a source challenge versus a rebuttal of P.
Use independent expected results with explicit profile assumptions.

Only after these contracts work should larger model-driven play be described
as testing the complete source-dependent family. Finite NASCQ completeness
within a versioned set and open-world coverage remain distinct measures.

## E3: multimodal fidelity and Kialo/API pilot

The retained Kialo graph and its external citations are separate evidence
objects. Preserve graph-export provenance, claim/version/location IDs, link
anchor/context, deduplicated artifact bytes and every distinct citation use.
Do not infer speakers, chronology or consent from reconstructed pro/con roles.

Build three linked datasets: source-grounded argument annotations, evidence
interpretations/assessments, and explicitly reconstructed playthroughs. Each
episode records what was observed, inferred or generated. An annotator's
unavailable source remains unavailable; it cannot claim a check it could not
perform. API annotation may proceed provisionally under D045; a human check is
not invented to upgrade its status.

Begin with a bounded text/PDF slice and failed-source controls, while retaining
other modalities. Add scanned PDF/image, audio/video and compound-page adapters
with modality-specific references. Verify page/region/time alignment and exact
displayed views against original material. Automatic extraction quality needs
its own measurement; checks against declared metadata alone do not establish it.

Sample before choosing the study size: estimate prevalence of source-dependent
arguments, inaccessible links, OCR/transcript failures, disagreements and cost.
Include expert/testimony, measurement/sign, statistical/causal, practical and
context/definition uses, with both supportive and contrary evidence. Retain
all attempted cases and failure denominators. Pilot confidence intervals and
paired comparisons determine a subsequent sample; a convenient fixed number
cannot establish the needed statistical power in advance.

Split by discussion, upstream source/common origin, near-duplicate content and
derivation lineage. Keep alternate reconstructions and model paraphrases with
their source group. A random edge split or two mirrors of the same article do
not supply independent train/test evidence. Use blinded first passes and
multiple model families where available; disclose shared models/prompts/tools.

## E4: review/meta-argumentation, aggregation and interchange

Represent public review reasons with exact object/version, source selector,
claim, warrant, alternatives, defeaters and disposition. The same Walton
schemes can express many such meta-arguments. Explicit provenance and a bounded
review protocol are essential; circular endorsement is not independent proof.
Compare flat review, structured justification and bounded metadialogue at
matched cost, including planted defects and genuinely ambiguous cases.

Do not introduce one global credibility score or treat a scheme weight as a
calibrated probability. Study dependence-aware aggregation, competing accounts,
abductive alternatives and proof standards as separate policies with held-out
cases. One source's problem need not destroy independent evidence. Probability
and argumentation can be complementary; E02's conjunction passage is not a
sound reason to reject probability theory.

For AIF, document the exact supported projection, retain raw input and publish
a loss report. General AIF permits directional conflicts, preferences and
scheme-to-scheme links. The current importer rejects several of those patterns
by design. Role-sensitive validation and round-trip tests need separate native
work; the 2011 specification is not a complete current AIFdb interchange
contract, and this evidence patch adds no general AIF importer.

## What to measure, and what serves as a reference

| Question | Check/reference | Limit |
|---|---|---|
| Did the captured bytes and selected view survive? | Hash/length, local payload, selector and derivative checks | Does not authenticate origin or truth |
| Does the record represent the source faithfully? | Independent passage/region comparison, alternatives and modality-specific annotation | Agreement alone is not ground truth |
| Does the source support this particular premise? | Source-linked interpretation and warrant review, positive/adverse cases | Requires semantic judgment |
| Are evidence/CQ actions legal and effective? | Native transition/replay fixtures from a separate specification | Offline admission passes are insufficient |
| Does added evidence change the result appropriately? | Independently expected defeat/reinstatement/dependency cases; sensitivity analysis | Engine labels are acceptance under chosen semantics |
| Are annotations and reviews useful? | Error precision/recall, omissions, corrections, abstention calibration, effort | Keep unresolved cases and model/human status visible |
| Does play improve? | Matched-cost policies/opponents and held-out tasks; valid and invalid attempts retained | Win rate and desired-outcome enforcement are not truth |
| Can it scale? | Capture/validation/replay latency, memory, unique bytes, citation uses, error rates | Small fixtures are not capacity benchmarks |

Thresholds belong to a preregistered study after the pilot estimates variability
and the cost of errors. Separate interpretation, inference, procedure and
engineering failures in reports. A larger number of checks is not itself
evidence that the entire framework is valid or complete.
