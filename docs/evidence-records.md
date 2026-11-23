# Evidence records: the design and first offline checks for stage 2

A design note for D076 (evidence as records) and D077 (evidence captured for
research), the first slice of which is built in stage 2 ([plan.md](plan.md)
§0). It takes the record model of ChatGPT's review of stages 0 and 1 (22
November 2026), without that review's gating of capture, which the owner did
not adopt ([review-2026-11-22-stages.md](review-2026-11-22-stages.md)).

The [evidence-theory review](evidence-theory-review.md) grounds the refinements
below in seven supplied works. The provisional [evidence contract](evidence-contract.md)
now has a native Go offline checker and a consent-log reducer. These do not yet
connect to game execution, source capture, the dashboard or model prompts.
The existing game's E1 rule still checks citation presence. Stage 2 remains open.

## The owner's rules this follows

- Evidence cited in the data is captured whole, for research, under the
  fair use doctrine (the owner, 2026-11-22; D077). It is kept in the private
  workspace, large files in Git LFS, never in the public repository.
- Terms of use, privacy and instructions planted in captured pages are
  reviewed later and decided then (D079, open). Nothing here decides them,
  and capture does not wait for them.

## What a record holds

Five kinds of record, each with an id, a time and who or what made it.

1. **The captured artifact.** What was fetched or supplied (a URL, a file,
   a book page), when, how, its bytes' SHA-256, length and media type, and
   its state: complete, partial, failed, unsupported or unattempted. A failed or partial
   capture is recorded, not dropped.
2. **A derived representation.** OCR, a transcript, a translation, or a
   text cleaned of invalid bytes, each pointing to the original bytes and
   naming its method. A derivative never replaces the original: what a
   player or model was shown is one of these, and the record says which.
3. **A binding.** An attributed source claim, the exact retained representation
   and passage/region, the interpretation and warrant, and the particular
   argument/premise revision for which it is offered. Its role can be support,
   opposition, definition, context or a source challenge. Quoting a source
   establishes what it reports only to the extent the capture/attribution is
   reliable; accepting the reported proposition needs its own inference.
   A definition or context link cannot silently become factual support.
4. **An assessment.** One party's view, with separate fields:
   - authenticity;
   - relevance;
   - how strongly it supports what it is offered for;
   - whether its facts were checked.

   These are not one yes-or-no. Every assessment is kept.
   Where relevant, also record domain competence, observation/access,
   sincerity, source dependence and the scope of the assessment. Unknown,
   challenged and not applicable remain distinct. A model check does not
   acquire D045's human-verified status. A claimed fact check records its
   method; a supported or contradicted determination also cites basis claims.
   An inconclusive check may have no basis claims. The checker does not certify
   that assessment as correct.
5. **An admission event.** That a binding was admitted, or that an
   admission was withdrawn: by whom, for the exact binding revision,
   under which rule, with what effect in the game. Source identity, attributed
   claims, target snapshots and common-origin relations support these five
   families of records; they do not collapse them into one trusted-source flag.

Evidence is relational: material is offered for a target through an inference
and for a purpose. The same artifact can support one reading, fail to support
a stronger reading, and provide only background for another. Authenticity,
relevance, probative force and sufficient support are distinct questions.
See E02 and E05 in the [source registry](evidence-theory-sources.json).

## The first slice's admission policy (D076, refined 2026-11-22 and 2026-11-23)

- Two players, cooperative. A binding is admitted **for consideration** when
  both consent to its exact revision: source/representation ancestry, passage,
  source assertion, interpretation, warrant, argument target, role, context
  and policy. Agreement to the same document alone is insufficient.
  One acceptance is one party's scoped procedural commitment.
  Silence is not consent.
- Evidence that is not admitted can still be argued about: a player may
  attack a source or its relevance like any premise.
- Withdrawing a consent removes that route of admission only. A statement
  that has other, independent support keeps it.
- A changed artifact, derivative, interpretation, target or policy needs new
  consent. Consent to an old binding does not carry over. Retain the old
  records and use new identities for changed records.
- Admission for consideration does not assert a proposition, stipulate a
  premise, assign a weight or set a Carneades label. A future profile that
  permits stipulation must specify that separate act and its engine effect.
- A moderated game, when it comes (Track D), records the moderator's
  admission as a procedural decision, not as a finding of fact. Today a
  moderator's decision on a statement is a hard assumption, labelled so.

## Captured content is data

Text, images or recordings in a captured artifact are material to read,
never instructions. The required player/tool boundary prevents them from
calling tools, changing policy or reaching credentials. Quoting them in a
prompt alone does not establish that boundary. It must be enforced and tested
when capture-to-player integration is implemented, keeping exactly what each
model was shown. The offline checker fetches nothing and executes no evidence
content; it does not claim that future model integration is already protected.
This is the existing engineering requirement, not a decision of D079.

## Retention and reproducibility

- **Originals and derivatives are kept apart.** The record keeps both and
  says which is which. Invalid UTF-8 replaced by U+FFFD in an export, for
  instance, is a derivative.
- **An LFS pointer is not its payload.** Before a run uses an artifact, its
  bytes are checked against the recorded hash and length; a pointer that
  was never fetched is refused, even if the pointer itself is intact.
- **A hash is an integrity check relative to a record.** It does not establish
  the author's identity, authenticity, expertise, truth, entailment or a
  model's actual observation of the material. Those need separate evidence.
- **Common origins are retained.** Mirrored pages, quotations and model
  summaries may derive from one source. Missing dependency information means
  unknown independence. Agreement does not grant automatic corroboration
  credit; the first checker calculates no evidential probability or weight.
- **Acquisition failure is not contrary evidence.** Failed, unsupported,
  unattempted and partial records remain usable research records. A failure
  to retrieve a page does not establish that the page or a claimed fact never
  existed. An argument from absence needs its own search/detection warrant.

- **Removing a file is not purging it.** Removing a file from the workspace
  differs from purging an LFS object from the server, which needs the
  repository deleted or GitHub's support.

## What can be checked now

From `wag/`, run:

```sh
go run ./cmd/wag-evidence ../examples/evidence/bundle.json
go test ./internal/evidence ./internal/server
```

The first command checks the declared evidence contract, retained payloads,
selectors and the two-participant consent log. The sample is synthetic.
Neither command establishes empirical accuracy. [The implementation plan](evidence-implementation-plan.md)
separates completed offline checks from the remaining native integration and
data-study gates. Previous citation-string games retain their meaning.

## Not decided here

- D079: terms of use, privacy, and injected instructions as a policy
  question.
- The other game profiles (adversarial, moderated, automated), which wait
  for Track D.
- Which evidence a pilot captures first: per D077, the links of the claims
  it annotates.
