# Provisional offline evidence contract: wag-evidence/0.1 and 0.2

This is the first executable record contract for D076, informed by the
[literature review](evidence-theory-review.md). It is implemented by
`wag/internal/evidence` and `wag/cmd/wag-evidence`. It validates an offline
snapshot. The current game still accepts citation strings; this tool does
not fetch sources, call models, modify a game graph, assign weights or turn
admission into factual acceptance. Its version is deliberately provisional.

## Run and interpret the result

From `wag/`:

```sh
go run ./cmd/wag-evidence ../examples/evidence/bundle.json
go run ./cmd/wag-evidence -root /path/to/private/payloads /path/to/bundle.json
go run ./cmd/wag-evidence --record-set /path/to/evidence/NAME/records.json   # a game's record set, alone
go test -count=1 ./internal/evidence
```

The default payload root is the bundle's directory. Output is JSON containing
errors, warnings, exact binding fingerprints, active consents, bindings
admitted for consideration and identical-payload groups. Exit 0 means the
record contract passed; exit 1 means a decoding/contract failure; exit 2
means a usage, input-file or output-write failure. Decode errors go to stderr.
No failed log exposes a partly replayed admission as usable output.

`valid: true` is **not** a finding that the claims are true, the inference is
sound, the evidence is independently corroborated, or a live WAG graph accepts
the record. The output states this scope. Warnings name limitations of declared
media bounds, external graph references and claimed independence.

## Record meanings

The Go structs in `wag/internal/evidence/model.go` are the field contract.
The committed example is a complete machine-readable instance. All record
IDs are unique across the bundle. Every authored record names a declared
actor and an RFC3339 creation time. Actor kinds are human, model and tool;
these are declared identities, not authenticated credentials. Only players,
human or model, can be the context's participants (D076); a tool can author
records but cannot consent.

| Object | Required meaning |
|---|---|
| Context | Session and profile identity, the supported policy, two distinct participants (each a human or model player), declared SHA-256 pins for engine/catalogue/NASCQ inputs |
| Source | Descriptive source identity, kind, label and reference; a source is not globally trusted or automatically an expert |
| Artifact | Source identity, immutable revision, declared capture scope, capture status, omissions/failure where applicable |
| Representation | One retained byte sequence, artifact revision, original/derivative kind, capture/transformation method and parents |
| Claim | What is attributed to a source, exact representation/selector, scope, and quotation/paraphrase/observation/reconstruction mode |
| Target | Argument identity and revision, particular locus and proposition, as a supplied snapshot |
| Binding | Source claim, exact target, role, offered interpretation and an explicit inferential or use warrant |
| Assessment | Assessor, binding, distinct assessed dimensions, public rationale, basis and declared review/fact-check status |
| Dependency | A recorded relation among artifacts: common origin, derivation, claimed independence or unknown |
| Event | One participant's scoped consent or withdrawal for the exact binding fingerprint, in an ordered complete log, dated no earlier than its binding or the event before it |

The five design families remain artifact, derivative, binding, assessment and
admission event. The additional identity/claim/target records make their
references unambiguous and permit repeated uses without treating all uses as
one evidential judgment.

Context pins and target revisions are checked for form and included in the
binding identity. Their correspondence to the *actual* engine files, catalogue,
NASCQ and live graph must be established by the future adapter/run manifest.
This checker does not have those external objects. The synthetic example uses
explicitly synthetic identities and is not a receipt for a real WAG game.

## Capture and transformations

Capture statuses are `complete`, `partial`, `failed`, `unsupported` and
`unattempted`. Complete means complete within the declared scope and no known
omissions; the checker cannot independently establish that a dynamic website's
entire content was captured. Complete requires one original representation.
Partial requires an omission description and may honestly contain no body
bytes. Failed/unsupported needs a reason; unavailable states do not invent
retained content. A failed acquisition can therefore pass as an honest record.

Each retained representation has a relative local path, lowercase SHA-256,
explicit byte length and media type. The checker reads and hashes the bytes;
an unresolved Git LFS pointer cannot stand in for its payload. It rejects
static path escapes, observed symlink components and nonregular payloads, and
uses Go's confined directory handle for the final open. Files are never fetched
or executed. This is not a browser sandbox or a capture client.

Original representations have no transformation parents. OCR, transcript,
translation, excerpt and normalization records name retained parents and a
method; their ancestry must be acyclic. In this first profile, parents belong
to the same artifact revision. Combining several sources into a new synthetic
artifact is an explicitly unsupported transformation requiring an extension,
not permission to drop provenance. Citation/dependency relations among sources
are descriptive; they are not forced into this transformation DAG.

All records remain immutable for a run. A changed file must be checked again
before consumption. A prior receipt does not guarantee that the filesystem
has remained unchanged. The checker does not include an operating-system
race stress test or a general capture security audit.

## Selectors and reported assertions

Selector kinds are whole, text, page, time and region. Text uses zero-based,
half-open **Unicode code-point** offsets into an exact retained UTF-8 text
representation. It includes the exact quote. The same buffer checked against
the recorded hash and length is used to check the selector. Byte offsets and
UTF-16 code-unit offsets must be converted explicitly by the producer.

Page numbers start at 1. Times are integer milliseconds. Regions use positive
normalized width/height and coordinates within one declared page. Page count
and duration are declared metadata: the checker does not parse PDF pages,
decode images or measure audio/video duration. It emits that limitation rather
than reporting independent modality verification. Combined audiovisual
selectors, transcripts with word alignment and scanned-image quotation review
remain later extensions.

An exact quotation requires a text selector whose quote equals the claim text.
Quoting OCR or translation establishes correspondence with that derivative,
not accuracy of the OCR/translation or the original author's words. Paraphrase,
observation and reconstruction need their own modes and subsequent review.

The warrant field makes the offered inference inspectable. A nonempty warrant
does not make the inference valid. For example, an instrument report can be
offered for a temperature premise while its calibration and conditions remain
unresolved. A report of an allegation can be authentic while the allegation is
false. The checker deliberately supplies no truth field or automatic weight.

## Assessments, uncertainty and dependence

Assessment dimensions are authenticity, relevance, competence, observation
access, sincerity, probative force and independence. Their values are
supported, challenged, unknown and not applicable. These are attributed
assessments. Different parties' judgments coexist; a later assessment does not
overwrite an earlier one.

Fact-check status is not checked, supported, contradicted or inconclusive.
A claimed check needs a method, and a determination needs cited basis claims.
References are checked; the checker does not independently perform the check.
The general review level distinguishes unreviewed, model checked and human
checked. The latter requires a declared human author, but identity and actual
human involvement still require external attestation, and the report warns so
(`human_checked`). D045 remains unchanged.

No dependency record means unknown independence. Common origin and derived
relations remain explicit; an independence claim requires a rationale and
produces a warning that it has not been certified. Identical payload hashes
are reported as groups. They are not converted into multiple independent
supports, nor are different hashes sufficient to establish independence.
Copying, paraphrasing, collusion and common measurement error need broader
review than byte comparison.

## Consideration-only consent and replay

The sole implemented policy is `two-party-consideration/1`. Both distinct
context participants must consent. One party's silence, an assessment, a
moderator-like label, or duplicate consent by the same party is insufficient.
Events are numbered from 1 with no gaps, and their times do not go back along
that sequence; no event is dated before the binding it concerns. Each event ID
is unique. A participant
may withdraw only their own active consent. Duplicate active consent and
withdrawal without active consent are invalid transitions in this profile.

An event includes `binding_revision`. This is produced by the checker from
the exact binding, claim, target, selected representation and its retained
ancestry, relevant artifact/source records, the full context and the
participants' actor records. Change the interpretation, target, role, source
metadata, derivative, profile, participants, a participant's declared kind or
label, or input pins, and the prior fingerprint no longer applies. Assessments do not
grant consent and do not themselves change this consideration-only identity.

Fingerprints use the explicitly named `wag-binding-go-json/2` serialization
implemented in the Go code (sorted string map keys). Version 2 (2026-11-23)
added the participants' actor records; a consent pinned under version 1 is
stale and must be given again. A test pins the fixture's fingerprint, so a
change to it is a deliberate version change. This is not an assertion
of RFC canonical-JSON compatibility. Cross-language writers should consume the
checker-produced fingerprints or reproduce the named algorithm against pinned
fixtures. A future canonicalization migration needs a new algorithm version.

**Version 0.2** (`wag-evidence/0.2`, 2026-11-27) is the version a game writes
([`e1-evidence.md`](e1-evidence.md), choice 6). It is 0.1 with one change: a
binding's revision is `wag-binding-go-json/3`, the same serialization with
every record's `created_at` left out (the binding's, and its claim's,
target's, representations', artifacts' and sources'). A time says when a
record was written, not what a consent is to, and a game replayed, loaded or
resumed must find the revisions it recorded. A 0.1 bundle keeps `/2` and its
meaning; the checker takes both. In Go, `evidence.ValidateRecords` applies
the same record and consent rules as `evidence.Validate` without reading a
payload or checking a selector against its bytes, for records whose payloads
were checked when they were taken; it checks each selector's structure (its
kind, offsets, declared pages, duration and region, and stray coordinates) as
`Validate` does.

**A game's record set** ([`e1-evidence.md`](e1-evidence.md), choice 2) is a 0.2
bundle without the records a game writes: no `context`, `targets`, `bindings`,
`assessments` or `events` key, even empty; no id, its actors' included, beginning
with `game:`; every id a plain identifier (letters, digits, `.`, `_`, `:` and
`-`, beginning with a letter or a digit, at most 128 characters: a game's prompt
writes ids as they are); every record by one of its own actors
(`evidence.DecodeRecordSet`). Its name in an evidence folder is letters, digits,
`.`, `_` and `-`, beginning with a letter or a digit, at most 64 characters.
The game adds its context and actors (`evidence.GameBundle`). `wag-evidence
--record-set NAME/records.json` checks a record set alone, with the game's actors
and a placeholder context of two players whose pins are zeros; its payloads are
under `NAME/`.

First create a bundle with no events and inspect `binding_revisions`. Append
the participants' actual consent events using those values, then recheck.
Label synthetic event fixtures explicitly; never present their manufactured
consents as actual participants' decisions. A real capture does not acquire
consent merely to make it usable. Changed records receive new IDs; keep the
previous records/events for history.
Active old-revision admission does not authorize a new graph revision.

Withdrawal affects that binding's procedural admission. It does not delete
captured bytes, erase assessments, retract an unrelated participant's statement
or remove an independently supported argument. The offline tool has no live
graph; the future adapter must prove those effects with game-level tests.

## Decode contract and limits

JSON nested deeper than 10000 levels, encoding/json's own limit, is refused
before any recursive scan. JSON decoding rejects duplicate member names after
unescaping, unknown fields,
noncanonical field spelling, explicit null values, trailing JSON, invalid raw
UTF-8 and unpaired Unicode surrogates. Valid surrogate pairs and intentional
U+FFFD remain representable. Payload byte length and the applicable selector
coordinates must be explicit even when zero; unused selector coordinates must
be zero/empty. Empty optional collections may be omitted. No coercion turns
an arbitrary JSON object into evidence text.

The native `Validate` entry also rejects invalid UTF-8, NULs and nonfinite
numbers in programmatically constructed records; serialization errors cannot
produce a misleading fingerprint. Native typed values have already lost JSON
field-presence information, so file/API ingestion must use `Decode` first.

Payloads are read no further than one byte past their declared length, so a
larger file is a mismatch without being read to its end. Hashing streams. A
text representation is read into memory once per check, bounded the same way,
and a selection is found by walking its code points, so a check needs about
the payload's size in memory; a claim on a payload that failed its check is
not checked again. Reports are deterministic. Very large text and concurrent
live capture are not capacity-validated; no arbitrary size ceiling is treated
as a semantic rule, and capacity should be measured before scaling beyond the
bounded pilot.

The [implementation plan](evidence-implementation-plan.md) lists native game
integration, CQ protocols, modality checks and empirical studies separately.
