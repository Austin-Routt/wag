# Evidence records: the design for stage 2

A design note for D076 (evidence as records) and D077 (evidence captured for
research), the first slice of which is built in stage 2 ([plan.md](plan.md)
§0). It takes the record model of ChatGPT's review of stages 0 and 1 (22
November 2026), without that review's gating of capture, which the owner did
not adopt ([review-2026-11-22-stages.md](review-2026-11-22-stages.md)).

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
   its state: complete, partial, failed or unsupported. A failed or partial
   capture is recorded, not dropped.
2. **A derived representation.** OCR, a transcript, a translation, or a
   text cleaned of invalid bytes, each pointing to the original bytes and
   naming its method. A derivative never replaces the original: what a
   player or model was shown is one of these, and the record says which.
3. **A binding.** Which passage or region of which artifact is offered for
   which premise of which argument, and in what role: to support it, to
   define a term, to give context, or to attack a source. A link that
   defines a term is not, by itself, evidence that the claim is true.
4. **An assessment.** One party's view, with separate fields:
   - authenticity;
   - relevance;
   - how strongly it supports what it is offered for;
   - whether its facts were checked.

   These are not one yes-or-no. Every assessment is kept.
5. **An admission event.** That a binding was admitted, or that an
   admission was withdrawn: by whom, for which revision of the artifact,
   under which rule, with what effect in the game.

## The first slice's admission policy (D076, refined 2026-11-22)

- Two players, cooperative. A binding is admitted when both consent to the
  same revision of the artifact. One acceptance is one party's commitment.
  Silence is not consent.
- Evidence that is not admitted can still be argued about: a player may
  attack a source or its relevance like any premise.
- Withdrawing a consent removes that route of admission only. A statement
  that has other, independent support keeps it.
- A changed revision of the artifact needs new consent. Consent to the old
  revision does not carry over.
- A moderated game, when it comes (Track D), records the moderator's
  admission as a procedural decision, not as a finding of fact. Today a
  moderator's decision on a statement is a hard assumption, labelled so.

## Captured content is data

Text, images or recordings in a captured artifact are material to read,
never instructions. Nothing in them can call a tool, change a player's
policy or model, or reach a credential. A prompt quotes them as quoted
material, and the record keeps exactly what each model was shown. This is
an engineering rule of how WAG works, needed for correct behaviour. It does
not decide D079's questions.

## Retention and reproducibility

- **Originals and derivatives are kept apart.** The record keeps both and
  says which is which. Invalid UTF-8 replaced by U+FFFD in an export, for
  instance, is a derivative.
- **An LFS pointer is not its payload.** Before a run uses an artifact, its
  bytes are checked against the recorded hash and length; a pointer that
  was never fetched is refused, even if the pointer itself is intact.
- **Removing a file is not purging it.** Removing a file from the workspace
  differs from purging an LFS object from the server, which needs the
  repository deleted or GitHub's support.

## Not decided here

- D079: terms of use, privacy, and injected instructions as a policy
  question.
- The other game profiles (adversarial, moderated, automated), which wait
  for Track D.
- Which evidence a pilot captures first: per D077, the links of the claims
  it annotates.
