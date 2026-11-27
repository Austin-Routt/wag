# Annotation records, format 2: the identity contract

Status: decided on 2026-11-24 (D092, under the standing rule of 2026-11-22), from the independent
reviews of 2026-11-24 and the developer's assessment of them (wag-workspace
`verify/r-review-2026-11-24/ASSESSMENT-2.md`: "identity before the first gold draft"). It is stage 2's
identity contract (plan §0): the first gold (D086) is drafted in this format, so that the owner's
verification refers to objects that stay the same.

Format 1 is the batch-2 format of 2026-10-13 (`kialo-batch2/guidelines.md` §4 and §5, until this
contract). Its files stay readable: `annotator-A.json`, `annotator-B.json`, the gold as it was before
this contract (`kialo-batch2/gold-format1.json`, WAG 96c3c5a) and the annotation files of runs made
before. They are checked by the format-1 rules and scored by the same method as before, against
today's gold: a file identical to the format-1 gold scores 0.94 on scheme, 0.88 on instances and 0.96
on target_formula against the gold D039 recoded. Checking an old file against a later `walton.yml`
says nothing about it, so a check of the format-1 gold that
depends on the formulas holds only while `walton.yml` is the one it was made with (WAG 96c3c5a's);
agreement with it does not depend on `walton.yml`.

What format 2 adds, and why:

| What | Why (the reviews of 2026-11-24) |
|---|---|
| An id for every argument instance, which edits do not change | A target must name the inference it challenges, not only the claim; one claim can hold several inferences, and an instance can become secondary |
| Exact targets: claim, instance, part and the formula itself | A claim ID and a part do not say which premise of which inference is attacked |
| The case record: its contexts, its actors, its actions and who performs each, with aliases | Who performs an action comes from a mapping, not from the action's name (D063, D091); an answer must stay in its target's context and agent |
| Four kinds of missing value kept apart | Unknown, absent, not applicable and false must not collapse into one null; a derived value is not stored |
| Revisions: the inputs a file was checked against, and an item's revision for its verification | Stable ids survive edits; revisions say which content was checked or verified |
| Agreement on the exact target, the question, the bindings and all instances | Agreement on eight coarse fields does not measure them |

## 1. The files of a dataset

| File | What | Format |
|---|---|---|
| `items.json` | The claims, with their Kialo context | unchanged |
| `guidelines.md` | What annotators are told; §4 gives the output format | format 2 since 2026-11-24 |
| `lexicon.json` | The shared atoms, each with its English | unchanged |
| `case.json` | The case record (§3): given to annotators with the lexicon | new, `wag-case/1` |
| `gold.json` | The adjudicated annotation | format 2 since 2026-11-24 |
| `dataset.json` | The manifest: names the files, and `format`, the format the guidelines ask for | `format` new |

A dataset's format is its manifest's `format`: `wag-annotation/2`, or `wag-annotation/1` for a dataset
whose own guidelines ask for format 1. A manifest that says nothing means its gold's format, or format
2 (the built-in guidelines') for a dataset without a gold. A gold in another format than the one its
manifest names is refused when the dataset is read. A run records the
format it was made in (`run.json`, `settings.format`); a run made before this contract has none and is
format 1, so its annotations are checked as they were made.

The typing of predicate positions (which argument of a predicate is an agent and which an action,
§3) is `wag/data/annotation/positions.json`, until the catalogue types each scheme's variables
(plan §2).

## 2. An annotation file

```json
{"format": "wag-annotation/2", "annotator": "gold", "case": "kialo-13292",
 "inputs": {"items.json": "<sha256>", "lexicon.json": "<sha256>", "case.json": "<sha256>", "walton.yml": "<sha256>", …},
 "status": "development material (D086): no item is verified by a person",
 "items": [ … ]}
```

- `format` is required in format 2. A file without it, or with `wag-annotation/1`, is format 1.
- `case` names the case record the file's actors, actions and contexts refer to (its `id`). A run
  writes it in its annotation files when its dataset has a case record.
- `inputs` pins every input the gold was checked against, by the SHA-256 of their bytes: the
  dataset's `items.json`, `lexicon.json` and `case.json`; `walton.yml`, its `id-migrations.json` and
  `links.json`; `positions.json`; and the game's `schemes.json`, `critical_questions.json` and
  `edges.json`, which give the questions R2 reads. Required in a gold; a run's annotation files
  leave it out, because the run pins its inputs. A change to a pinned file is applied with the gold
  checked again and its pins updated in the same commit: `validate.py --repin` writes them when
  nothing but the pins is wrong.
- `history`, optional: the file's migrations, each with its date, by whom and what changed. A
  change to an item's coding is recorded in the item (`adjudication.recoded`). An entry written by
  the owner's verification (`verify_gold.py`, D086) also holds `verification_verdicts`: each verdict
  it applied or recorded, `{"item": "<claim>", "choice": "gold" | "none" | …, "at": "<the verdict's
  time, RFC 3339, UTC>", "item_rev": "<sha256>" | null}`, null only for a verdict other than *gold*
  given on a gold the script could not place, whose revision is unknown. A later import compares a
  verdict's time with these, so that an older verdict cannot undo a newer one
  ([d086-first-gold.md](d086-first-gold.md) §5).
- `items`: one per claim; a gold has every claim of `items.json`, in its order.

## 3. The case record (`case.json`)

```json
{"format": "wag-case/1", "id": "kialo-13292", "discussion": "13292",
 "thesis": {"claim": "13292.1", "action": "tax_meat"},
 "default_context": "present",
 "contexts": [{"id": "present", "text": "…"}],
 "actors": [{"id": "governments", "text": "…"}],
 "actions": [{"id": "tax_meat", "actor": "governments", "aliases": []}],
 "status": "…"}
```

- **Contexts.** The decision situations the discussion argues about. An instance is in its
  `context`, or in the case's `default_context` if it names none. A claim about another situation
  used as evidence (an analogy's source case) is an instance in the present context whose formulas
  name that case; a context of its own is for an argument made about another decision.
- **Actors.** Who can act in the case. Each is an atom of the lexicon.
- **Actions.** The actions the discussion deliberates, each with the one actor who would perform it.
  An action is an atom of the lexicon. `aliases` lists other atoms for the same action; no atom is
  the alias of two actions, both an action and an alias, or an actor's id and an action's or an
  alias (since 2026-11-25: an actor's id names the actor). What others would do in response
  (people switching to other protein) is a consequence, a state, not a declared action.
- **What an alias is for.** A file writes each action by one atom. Aliases let other vocabularies be
  read as the same action: who performs it (R14), whether a better means is another course (R9) and
  whether a cause is an action the case deliberates (R4) read an alias as its action, and so do
  agreement and the choice of claims to review (§11), where an action stands: in the binding of a
  variable in an action's place of the instance's scheme, and in a formula's places that
  `positions.json` types as an action's or, within an instance, that its scheme fills with an action
  variable (`sunk_costs(A,C)`'s A, which `positions.json` leaves untyped), so that a binding and the
  formulas it fills read alike. A goal or a consequence spelled like an alias keeps its name (since
  2026-11-25): read as the action, it would make an annotation that puts the action in a goal's place
  agree with one that names the goal. A reference is exact: a target's
  formula (R13), an answer's conclusion (R8) and a challenge's formula (R10) are written as the
  instance they refer to has them, so an alias there is refused.
- The thesis's action is a declared action; its actor is the agent the case deliberates for. The
  thesis is a claim of the dataset, and every actor, action and alias an atom of its lexicon
  (checked when the dataset is read, and by `validate.py`).
- **A dataset without a case record** (one made from a Kialo capture, say) is checked by the Go check
  without R14 and R15: there are no declared actors, actions or contexts to check against.
  (`validate.py`, batch 2's checker, needs one.) Its annotators are told
  so: they bind agents and actions to the atoms the claims name, and give no instance a context.

Typed positions (`positions.json`): `possible_bring_about(P,A)` puts an agent in its first place and
an action in its second; `should_be_performed(A)` an action in its only place; and so on for every
predicate of the practical and consequence schemes. The table was derived from `walton.yml` and
checked: no variable of any scheme is both an agent and an action.

## 4. An item

| Field | Required | What; when it is null |
|---|---|---|
| `kialo_claim_id` | yes | The claim |
| `literal` | yes | What the claim says (guidelines §2) |
| `instances` | yes, at least one | §5; exactly one is primary |
| `retired` | no (none) | Ids of the item's own instances removed from it; none is used again |
| `role` | yes | `kind`; `move` (optional: `side_effect`, `alternative_means`); `cq`, null exactly when `kind` is not `cq_answer`; `polarity`, `for` or `against`, null only for the thesis; `confidence` |
| `target` | yes | §6. `kialo_parent` and `kialo_relation` are the claim's first Kialo parent and its relation, as `items.json` has them; they and `actual` are null only for the thesis, which has no Kialo parent. `other_kialo_parents` is there exactly when the claim has more than one Kialo parent (R7) |
| `evidence` | yes | `status`; `urls`, a list (empty when it cites none) |
| `challenges` | yes | §7 |
| `notes` | yes | `text`; `candidate_pattern`, null when it names none |
| `adjudication` | in a gold | `changed_fields`, `reason`; `recoded`, each later change with its date, by whom and why |
| `verified` | no (not verified) | §9 |

`re_parented` is not recorded: it is derived (the actual target claim is not the Kialo parent), as
the tooling has computed it since 2026-11-24. A format-2 item that records it is refused (R16).
Fields this contract does not name are allowed and not read.

## 5. Instances

| Field | Required | What |
|---|---|---|
| `id` | yes | `<claim>#<n>`: the item's claim ID, `#`, and a positive number (no leading zero) |
| `primary` | yes | `true` for exactly one instance of the item, a lone instance too. An annotator's record lists it first, as `<claim>#1`; a gold keeps the ids its history gave them (R12) |
| `context` | no (the case's default) | A context of the case |
| `scheme` | yes | `walton_id`, `compendium` (null if none), `wag_id` (null if none), `fit`, `confidence`, `alternatives`; `description` optional |
| `bindings` | yes | Every variable of the scheme, each to an atom; in a fragment, a variable of formulas it does not state is `{"unstated": "<why>"}` |
| `premises` | yes | As in format 1: `role`, `formula`, `en`, `status`; `note`, `cite` optional |
| `conclusion` | yes | `formula` (null only for `other` and `none`, and for a fragment whose conclusion names a variable it leaves unstated: no formula can be written without it), `en` |
| `comparison` | in `better_means` | `more surely`, `sooner` or `at lower cost` (R9) |
| `by` | no | `"the discussion"` on an instance that is not the claim's own but the argument the discussion needs for the claim (§13, D096); absent, the instance is the claim's |
| `label`, `note` | no | Free text |

**An id is the instance's, whatever happens to its text.** It is given when the instance is made: the
next number after every number the item has used, its retired ones included. An edit of the
instance (its scheme, bindings, formulas, English, fit or confidence) keeps its id; an instance that
is removed goes to the item's `retired`, and its number is not given again. Ids are the annotator's
own: annotator A's `13292.108#1` and the gold's need not be the same instance. A target refers to an
instance of the same file. A target in a claim whose annotation the annotator cannot see (one not
annotated yet, or any earlier one in an independent first pass) names its primary instance
`<claim>#1`, by the rule that an annotator's record numbers its primary instance `#1`; another
instance of it is `{"unresolved": why}`.

## 6. Targets

```json
"target": {"kialo_parent": "13292.108", "kialo_relation": "con",
           "actual": {"claim": "13292.108", "instance": "13292.108#1", "part": "exception",
                      "formula": "realize_more_effectively(governments,environmental_harm_from_meat_reduced,tax_meat)"}}
```

- `claim`: the claim targeted, any claim of the discussion (guidelines §2 item 5).
- `instance`: the id of an instance of that claim, in the same file; any of its instances, the
  secondary ones included.
- `part`: `premise` (a premise or an assumption), `exception`, `conclusion` or `inference`.
- `formula`: exactly a formula of that instance in that part, written as the instance has it: one of
  its premises or assumptions, one of its exceptions, or its conclusion. Null exactly when the part
  is `inference`, which has none.
- When the annotator cannot tell which instance, part or formula is meant, the field is
  `{"unresolved": "<why>"}`. An unresolved target is development material: it is kept, it is not
  scored on the fields it leaves open (§10), and a verified item has none (§9).
- An entry of `other_kialo_parents` that names a part (`premise`, `exception`, `conclusion`) names its
  `instance` and `formula` the same way.

## 7. Challenges

Each challenge (guidelines §2 item 7) names the instance it is about, by its `instance` id, and a
formula of that instance: `{"cq": "33.2 CQ1", "wag_cq": "CQ:likelihood", "type": "likelihood",
"instance": "13292.7#1", "formula": "brings_about(tax_meat,food_less_affordable_for_poor)"}`. The instance is
one of the claim's own, or the instance its target names. `cq` is null when the book prints no
question for it (a type of challenge of WAG's own: "no printed CQ"), `wag_cq` when the game has no
such question; `formula` is null when the challenge is about the instance as a whole (its
circumstances, say), and it then names the claim's primary instance unless it says otherwise.

## 8. Missing values: four kinds, kept apart

| Kind | Written as | Where |
|---|---|---|
| Not applicable | `null` | Only where §4–§7 say: `role.cq` when the role is not `cq_answer`; the thesis's `polarity`, `kialo_parent`, `kialo_relation` and `actual`; the formula of an `inference` target, or of a challenge about a whole instance; a challenge's `cq` when the book prints no question for it, and its `wag_cq` when the game has none; `compendium`, `wag_id` and `candidate_pattern` when there are none; the conclusion formula of `other` and `none`, and of a fragment whose conclusion names a variable it leaves unstated (the absence is recorded once, in that variable's `{"unstated"}`) |
| Unknown | `{"unresolved": "<why>"}` | A target's `instance`, `part` or `formula` |
| Absent | `{"unstated": "<why>"}` | A fragment's binding of a variable of formulas it does not state; and `urls: []`, an empty list, when a claim cites nothing |
| False | `false` | Only a real yes or no: `primary` |

A field the format requires is never left out to mean any of these. A derived value (re-parenting) is
not stored.

## 9. Revisions and verification

- An **input's revision** is the SHA-256 of its bytes; a gold pins its inputs (§2).
- A **claim's text revision** is the SHA-256 of its text in `items.json`. The gold pins `items.json`
  as a whole; when the pin changes, the items whose text changed are found by comparing their texts,
  and a verification records the text it was made on.
- An **item's revision**, `item_rev`, is the SHA-256 of the item without its `adjudication` and
  `verified` fields, in canonical JSON (keys sorted, no spaces, UTF-8; `validate.py`'s `item_rev`).
- `verified`, when present, records a person's verification (D045, D086): `{"by": "owner", "date":
  "YYYY-MM-DD", "item_rev": "<sha256>", "text_sha256": "<sha256>"}`. `by` is `owner`, who verifies
  the first gold (D086). It holds only while the item's revision and the claim's text are the ones
  verified: any later change to either makes the verification stale, and the check says so. A
  verified item has no unresolved value. Until the owner verifies an item, it is development
  material, not gold (D086).

## 10. The rules

`validate.py` checks a format-2 file with these rules, besides R2, R4 and R7–R11 (R1 is format 1's; R3,
R5 and R6 of the adjudication are coding conventions, not checks, and R13 replaces R5's
primary-only targets). The field contract (R16) is checked first: a malformed item is reported once,
never as a traceback, and no other item's rule reads it as a target. The Go check
(`annotate.Collection.CheckIn`, given the annotator's own annotations) reports the problems of the
rules marked Go on each annotator's annotations, in the claim's view and the report's counts (a
reviewer's adjudications are not checked); it refuses nothing (a reply that cannot be read is what a
run refuses). A target whose claim the
annotator has not annotated cannot be looked up, and is not reported.

| Rule | It checks | Go |
|---|---|---|
| R12 | Every instance has an id `<claim>#<n>` of its own claim, unique in the item and not retired (a retired id is the item's own); an annotator's record (one with no `adjudication`) lists its primary instance first, as `<claim>#1` | yes |
| R13 | A target, and each `other_kialo_parents` entry that names a part (any part, the inference included), names an instance of its claim, and a formula of that instance in that part; a format-1 target names the primary instance's | yes |
| R14 | An agent position is bound to an actor of the case, an action position to an action or an alias of one, and every action of an instance with an agent is its agent's | yes (with a case record) |
| R15 | An instance's context is one of the case's, and a claim's primary instance is in its target instance's context | yes (with a case record) |
| R16 | Every field of §4–§8, each of its type; nulls only where §8 allows them, markers only where it allows them; no `re_parented`; `kialo_parent` and `kialo_relation` as `items.json` has them; a gold names its case and pins every input it is checked against | in part: bindings (every variable, a fragment's unstated ones marked, and only formulas it binds), the target and polarity nulls, an empty target, `re_parented`, one primary instance, a fragment's conclusion |
| R17 | A `verified` item's revision and its claim's text are the ones verified, `by` is the owner, and it has no unresolved value | no |
| R18 | An instance whose fit is `explicit` has every premise-role formula (`role: premise`) explicit; an assumption or exception may be supplied (guidelines §5 item 6; since D086 round 1's draft 2, 2026-11-27) | yes |
| R19 | An inferred premise cites (`cite`) another claim of the discussion, the one that says it: a claim `items.json` lists, of the batch or of a claim's context (its parents, ancestors, siblings and children) (guidelines §2 item 3; since the same draft; the context's claims since the independent review of that draft) | no: the guidelines ask an annotator to cite the claim, but the output format (guidelines §4) has no `cite` field, so a model's record has none to check |
| R20 | An instance by the discussion (§13, D096) is not primary; its fit is `reconstructed`; no premise of it is `explicit`, and at least one is `inferred` (cited, R19); it concludes the statement of the claim it is recorded at (that claim's primary conclusion) or its negation; an item of the batch targets it. A premise of another instance that cites the claim it is recorded at may not take a formula that only the discussion's instance has: the reconstruction is not a claim, and cannot be cited (since D086 round 1's draft 3, 2026-11-27) | in part: not primary, fit, no explicit premise, the conclusion |

Rules that change in format 2:
- **R4** (D039): a step whose cause is only proposed or possible is `cause_to_effect_modal`. Format 2
  checks it by the record, not by a premise's status: a `cause_to_effect` whose cause is an action
  the case deliberates (which has not occurred), or the effect a modal step of the same claim says
  only may occur, is refused. Format 1 keeps the rule of the N8 note.
- **R7**: `other_kialo_parents` is there exactly when the claim has more than one Kialo parent, and
  the parents, each with its relation, are `items.json`'s, in any order, each as often as it has it:
  the source's pro or con, not the interpreted polarity (`role.polarity`), which may differ (since
  2026-11-25, after an external review). Each claim id and relation is a string (R16). The Go check
  reads them as `validate.py` does since 2026-11-26 (an independent check found it took a number for
  a claim id and did not check when the list is recorded).
- **R9**: A2 is another action than A as the case maps them, so an alias of A is not another course;
  the goal premise compared is the targeted instance's.
- **R10**: a challenge names an instance, its own or its target's, and a formula of that instance.
- **R2** (since step C5.2, 2026-11-27): a challenge's `wag_cq` is checked as a role's `cq` is. Where the game asks
  questions at its formula (at a premise or an exception, those whose pattern has its predicate; at the
  conclusion, those asked there), it names one of them, and a null is refused. Where the game asks none
  there, it names a question the game asks of the scheme at a formula of the game's own form, one
  `walton.yml`'s scheme lacks (analogy's `relevant_differences`), not one asked at another formula of the
  scheme, at its conclusion or of the whole argument; and none when the game asks the scheme none.
- **R16** (since step C5.2): an instance's `wag_id` is the game's scheme for its `walton.yml` scheme, as the
  pinned `schemes.json` maps them, or null when the game plays none.

## 11. Agreement

Agreement (`agreement.py`; the Go measure, `annotate.Agree`) compares, besides the eight fields of
format 1 (scheme, fit, role, polarity, target claim, target part, re-parenting, evidence):

| Field | Value compared |
|---|---|
| `cq` | The question answered (`role.cq`), or none |
| `target_instance` | Which inference of which claim the claim targets: the target claim with `primary`, or with `secondary:` and that instance's scheme and conclusion, read in the annotator's own file (ids are the annotator's own), so that two targets of different claims disagree even when each is its claim's primary inference (since 2026-11-26; before, the slot alone was compared, and 13292.29's two readings agreed). It measures the targeting decision, not how the target claim itself is coded, which its own fields measure: a different coding of a claim's primary instance does not spread to the claims that target it. A secondary target is named by that instance's scheme and conclusion, so a different coding of it does show in the claims that target it. A format-1 target is its claim's primary instance |
| `target_formula` | The exact formula targeted; a format-1 conclusion target is its instance's conclusion. It does depend on how the target claim is coded: the formula attacked is that coding's |
| `bindings` | The primary instance's bindings, every variable and its atom |
| `instances` | Every instance's scheme and conclusion: how the claim was divided into inferences. An instance by the discussion (§13) is left out: it is not how the claim itself was divided (since 2026-11-27) |

A value that is unresolved, or that cannot be read in the annotator's file (a target claim it did not
annotate), is left out of that field's agreement and counted apart (`unresolved`): missing data is
neither agreement nor disagreement. Only what a field itself leaves open makes it unresolved: an
unresolved instance leaves `target_instance` out. An unresolved formula or part leaves the relation
unsettled (D102, since 2026-11-27): the claim's `role`, `cq`, `target_part`, `target_instance` and
`target_formula` are all left out and counted, while its scheme, fit, polarity and target claim are
compared as before (`agreement.py`'s `RELATION_FIELDS`, the Go `FieldIn`). In choosing the claims to
review, an unresolved value differs from any other, so such a claim is reviewed against an annotator
who settled the relation. With a
case record, an alias is read as its action first where an action stands (§3), in agreement and in
the choice of claims to review alike.

These fields are also the ones a run compares to decide which claims the reviewer reviews (*Review
the disagreements*): a claim whose annotations differ in their exact target or bindings is reviewed,
which it was not before this contract. On batch 2's two model annotations that is 48 claims of 50,
where the eight coarse fields chose 21 (the bindings and the instances above all): more than twice
the reviewer's calls. The reviewer's reply names the fields it decides; `instances` is one of them since this
contract, and `target`, `role` and `premises` cover the target's instance and formula, the question
and the bindings. The reviewer adjudicates each claim in a call of its own, and names a target's
instance as the annotation it follows does, which its adjudication of the target claim may not
follow. Its targets are read in its adjudications: a primary target (`<claim>#1` in every
annotator's numbering) as such, a target at a secondary instance as unresolved, and a target in a
claim it did not review as unresolved too.

## 12. Limits

- **Source spans.** Format 2 does not align premises with spans of the claim's text; `literal` and
  each premise's `status` say what the text states. Spans would be needed for training argument
  mining (the review's interpretation layer) and are left for when that task is taken up.
- **One target per claim.** A claim that answers several arguments at once (OC7) still records one
  target, plus its other Kialo parents.
- **Record checks, not the engine.** The engine does not know actors or contexts: an argument bound
  to the wrong agent is refused by these checks, not by Carneades. The game takes the mapping at step
  C5.
- **Evidence revisions** belong to the evidence records (D076, item E1), not to this format.
- **Agreement on aliases** depends on the case record listing them; an atom used by one annotator for
  an action the case does not list is simply another atom. It also depends on walton.yml typing the
  variables (`positions.json`): an instance of a scheme walton.yml lacks (`other`) has no variable
  known to stand for an action, so its bindings keep their spelling, though its formulas' places that
  `positions.json` types are still read. A formula outside an instance, a target's, is read by
  `positions.json` alone: an alias in a place only the target instance's scheme makes an action's
  (`sunk_costs(A,C)`'s A) keeps its spelling there.
- **The case record, like the lexicon, is drawn from the gold for batch 2**, so it shows annotators
  which atoms the gold treats as actions. That is fine for development material; the pilot's
  (stage 3) must not be drawn from its gold, as its lexicon must not.
- **Open for the first gold's draft (D086)**, which the migration left as coded:
  - four targets that format 1 forced onto a claim's primary instance (13292.12, .141, .182 and
    .198; `adjudication.md`, R5) may name the causal step they answer, now that a target can name a
    secondary instance (13292.12's, 13292.9#1, is a secondary instance itself since proposal
    D008-D074 made the bridge 13292.9's primary, 2026-11-25, but not the causal step);
  - the gold's four fragments (13292.6, .11, .108 and .218) bind every variable of the scheme, where
    guidelines §5 item 6 leaves those of formulas the fragment does not state unstated: annotators
    who follow the guidelines disagree with the gold on their bindings until one of the two changes.

  Round 1 of the first gold's draft (2026-11-27, `d086-first-gold.md` §10) names the causal step for
  all four targets, and applies §5 item 6 to 13292.6, .11 and .218; 13292.108 keeps every variable
  bound, since the thesis has no argument and the fragment carries the environmental argument that
  eleven targets point into: a question for the other reviewer and the owner. Two more targets that
  format 1 moved for the same reason came to light: 13292.202, a mitigation, which the draft puts at
  the conclusion of the argument that cites the harm (13292.127#2), where both annotators put it; and
  13292.208, at 13292.104#1's better-means premise, which both annotators put at 13292.104#2's
  circumstances, the premise its note says it contradicts (round 2 takes it up).

## 13. The argument the discussion needs (D096)

Some claims state only part of an argument that no claim states as a whole, and other claims answer
its other parts. In the first gold, 13292.108 says only that meat production and consumption harm the
environment; the thesis has no argument of its own; seven claims answer the environmental argument
for the tax at its conclusion or at premises nobody states. D096 (the owner, 2026-11-27) keeps two
layers apart: what a claim states, and the argument the discussion needs.

**Where it is recorded.** The argument is an instance of the claim whose statement it concludes (here
the thesis, `13292.1#2`), marked `"by": "the discussion"`: it is that claim's argument as the
discussion gives it, not something the claim's text says (the claim may be a bare thesis), and not an
argument of the claim that states one of its premises. Its id is an ordinary instance id of that
claim (§5), so targets, agreement and the replay name it as any other instance. It is not primary;
the claim's own primary instance stays what the claim states.

**What it holds.** Its scheme and bindings as for any instance, fit `reconstructed`. A premise that a
claim of the discussion states is `inferred`, citing that claim (R19); the rest are `supplied`, with a
note of where the discussion presupposes them; none is `explicit`, since no claim states the argument.
At least one premise is cited: a reconstruction rests on what the discussion says. Guidelines §5
item 4's limit (one supplied link of three in a claim's own practical reasoning) is for what a claim
states; the discussion's argument is recorded because no claim states it, and its parts are cited
where claims state them.

**The claims around it.** A claim that states one of its premises is a fragment of it (guidelines
§5 item 6: a fragment of its parent's argument), backing that premise (`premise_backing`), as
13292.11 backs 13292.118's circumstances. A claim that answers another part targets it as any
instance (§6). The challenge records that list the questions it is open to are in the record of the
claim it is recorded at (R10). A premise elsewhere may cite the claim that states a part, never the
argument: a reconstruction is not a claim (R20).

**Its checks (R20).** `validate.py`: not primary; fit `reconstructed`; no explicit premise; at
least one inferred; it concludes the statement of the claim it is recorded at, or its negation; an
item of the batch targets it; and no premise of another instance cites that claim for a formula only
the reconstruction has. The Go check (§10): the field's value, not primary, the fit, no explicit
premise and the conclusion.

**Agreement and verification.** The `instances` field leaves it out (§11): it is not how the claim
itself was divided. Targets at it are compared as targets at any secondary instance (by its scheme
and conclusion). It is part of the item's revision (§9), so a verification of the claim it is
recorded at covers it, and a change to it makes that verification stale.

**The replay.** `to_wag.py` plays an instance by the discussion as an argument for its conclusion,
with its scheme, by the side that conclusion takes (the thesis's statement: Pro; its negation: Con):
the thesis's before any claim, another claim's right after that claim's own move. Claims that target
it are played at it. The report lists it apart from the claims' moves; the claims' counts are
unchanged by it.

