# Annotation records, format 2: the identity contract

Status: decided on 2026-11-24 (D092, under the standing rule of 2026-11-22), from the independent
reviews of 2026-11-24 and the developer's assessment of them (wag-workspace
`verify/r-review-2026-11-24/ASSESSMENT-2.md`: "identity before the first gold draft"). It is stage 2's
identity contract (plan §0): the first gold (D086) is drafted in this format, so that the owner's
verification refers to objects that stay the same.

Format 1 is the batch-2 format of 2026-10-13 (`kialo-batch2/guidelines.md` §4 and §5, until this
contract). Its files stay readable: `annotator-A.json`, `annotator-B.json`, the gold as it was before
this contract (`kialo-batch2/gold-format1.json`, WAG 96c3c5a) and the annotation files of runs made
before. They are checked by the format-1 rules and scored as before.

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
 "inputs": {"items.json": "<sha256>", "lexicon.json": "<sha256>", "case.json": "<sha256>", "walton.yml": "<sha256>"},
 "status": "development material (D086): no item is verified by a person",
 "items": [ … ]}
```

- `format` is required. A file without it is format 1.
- `case` names the case record the file's actors, actions and contexts refer to (its `id`).
- `inputs` pins the inputs the file was checked against, by the SHA-256 of their bytes. Required in a
  gold; a run's annotation files leave it out, because the run pins its inputs. A change to a pinned
  file is applied with the gold checked again and its pin updated in the same commit.
- `history`, optional: the file's migrations, each with its date, by whom and what changed. A
  change to an item's coding is recorded in the item (`adjudication.recoded`).
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
  An action is an atom of the lexicon. `aliases` lists other atoms for the same action: an alias
  stands for its action in every check, and no atom is the alias of two actions or both an action
  and an alias. What others would do in response (people switching to other protein) is a
  consequence, a state, not a declared action.
- The thesis's action is a declared action; its actor is the agent the case deliberates for.

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
| `retired` | no (none) | Ids of instances removed from the item; none is used again |
| `role` | yes | `kind`; `move` (optional: `side_effect`, `alternative_means`); `cq`, null exactly when `kind` is not `cq_answer`; `polarity`, `for` or `against`, null only for the thesis; `confidence` |
| `target` | yes | §6. `kialo_parent`, `kialo_relation` and `actual` are null only for the thesis, which has no Kialo parent. `other_kialo_parents` is required exactly when the claim has more than one Kialo parent (R7) |
| `evidence` | yes | `status`; `urls`, a list (empty when it cites none) |
| `challenges` | yes | §7 |
| `notes` | yes | `text`; `candidate_pattern`, null when it names none |
| `adjudication` | in a gold | `changed_fields`, `reason`; `recoded`, each later change with its date, by whom and why |
| `verified` | no (not verified) | §9 |

`re_parented` is not recorded: it is derived (the actual target claim is not the Kialo parent), as
the tooling has computed it since 2026-11-24. A format-2 item that records it is refused (R16).

## 5. Instances

| Field | Required | What |
|---|---|---|
| `id` | yes | `<claim>#<n>`: the item's claim ID, `#`, and a positive number (no leading zero) |
| `primary` | yes | `true` for exactly one instance of the item |
| `context` | no (the case's default) | A context of the case |
| `scheme` | yes | `walton_id`, `compendium` (null if none), `wag_id` (null if none), `fit`, `confidence`, `alternatives`; `description` optional |
| `bindings` | yes | Each variable of the scheme to an atom; in a fragment, a variable of formulas it does not state is `{"unstated": "<why>"}` |
| `premises` | yes | As in format 1: `role`, `formula`, `en`, `status`; `note`, `cite` optional |
| `conclusion` | yes | `formula` (null only for `other` and `none`), `en` |
| `comparison` | in `better_means` | `more surely`, `sooner` or `at lower cost` (R9) |
| `label`, `note` | no | Free text |

**An id is the instance's, whatever happens to its text.** It is given when the instance is made: the
next number after every number the item has used, its retired ones included. An edit of the
instance (its scheme, bindings, formulas, English, fit or confidence) keeps its id; an instance that
is removed goes to the item's `retired`, and its number is not given again. Ids are the annotator's
own: annotator A's `13292.108#1` and the gold's need not be the same instance. A target refers to an
instance of the same file.

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
- `formula`: exactly a formula of that instance in that part: one of its premises or assumptions, one
  of its exceptions, or its conclusion. Null exactly when the part is `inference`, which has none.
- When the annotator cannot tell which instance, part or formula is meant, the field is
  `{"unresolved": "<why>"}`. An unresolved target is development material: it is kept, it is not
  scored on the fields it leaves open (§10), and a verified item has none (§9).
- An entry of `other_kialo_parents` that names a part (`premise`, `exception`, `conclusion`) names its
  `instance` and `formula` the same way.

## 7. Challenges

Each challenge (guidelines §2 item 7) names the instance it is about, by its `instance` id, and a
formula of that instance: `{"cq": "33.2 CQ1", "wag_cq": "CQ:likelihood", "type": "likelihood",
"instance": "13292.9#1", "formula": "brings_about(tax_meat,tax_revenue_reduced)"}`. The instance is
one of the claim's own, or the instance its target names. `wag_cq` is null when the game has no such
question; `formula` is null when the challenge is about the instance as a whole (its circumstances,
say), and it then names the claim's primary instance unless it says otherwise.

## 8. Missing values: four kinds, kept apart

| Kind | Written as | Where |
|---|---|---|
| Not applicable | `null` | Only where §4–§7 say: `role.cq` when the role is not `cq_answer`; the thesis's `polarity`, `kialo_parent`, `kialo_relation` and `actual`; the formula of an `inference` target, or of a challenge about a whole instance; `compendium`, `wag_id`, `wag_cq` and `candidate_pattern` when there are none; the conclusion formula of `other` and `none` |
| Unknown | `{"unresolved": "<why>"}` | A target's `instance`, `part` or `formula` |
| Absent | `{"unstated": "<why>"}` | A fragment's binding of a variable of formulas it does not state; and `urls: []`, an empty list, when a claim cites nothing |
| False | `false` | Only a real yes or no: `primary` |

A field the format requires is never left out to mean any of these. A derived value (re-parenting) is
not stored.

## 9. Revisions and verification

- An **input's revision** is the SHA-256 of its bytes; a gold pins its inputs (§2).
- A **claim's text revision** is the SHA-256 of its text in `items.json`. The gold pins `items.json`
  as a whole; when the pin changes, the items whose text changed are found by comparing their texts.
- An **item's revision**, `item_rev`, is the SHA-256 of the item without its `adjudication` and
  `verified` fields, in canonical JSON (keys sorted, no spaces, UTF-8; `validate.py`'s `item_rev`).
- `verified`, when present, records a person's verification (D045, D086): `{"by": "owner", "date":
  "YYYY-MM-DD", "item_rev": "<sha256>"}`. It holds only while the item's revision is the one
  verified: any later change to the item makes the verification stale, and the check says so. A
  verified item has no unresolved value. Until the owner verifies an item, it is development
  material, not gold (D086).

## 10. The rules

`validate.py` checks a format-2 file with these rules, besides R2–R11 (R1 is format 1's); the Go check
(`annotate.Collection.Check`) applies those marked Go to every annotation a run receives, using the
annotator's own annotations for the cross-claim parts.

| Rule | It checks | Go |
|---|---|---|
| R12 | Every instance has an id `<claim>#<n>` of its own claim, unique in the item and not retired | yes |
| R13 | A target, and each `other_kialo_parents` entry that names a part, names an instance of its claim, and a formula of that instance in that part; a format-1 target names the primary instance's | yes |
| R14 | An agent position is bound to an actor of the case, an action position to an action or an alias of one, and every action of an instance with an agent is that agent's | yes |
| R15 | An instance's context is one of the case's, and a claim's primary instance is in its target instance's context | yes |
| R16 | Nulls only where §8 allows them, markers only where it allows them, no `re_parented`, every required field present; a gold pins its inputs and they are the files it is checked against | in part (nulls, markers, `re_parented`) |
| R17 | A `verified` item's revision is the one verified, and it has no unresolved value | no |

Rules that change in format 2:
- **R4** (D039): a `cause_to_effect` instance does not supply its `has_occurred` premise: a step whose
  cause is only proposed or possible is `cause_to_effect_modal`. Format 1 keeps the rule of the N8 note.
- **R9**: A2 is another action than A as the case maps them, so an alias of A is not another course;
  the goal premise compared is the targeted instance's.
- **R10**: a challenge names an instance, its own or its target's, and a formula of that instance.

## 11. Agreement

Agreement (`agreement.py`; the Go measure, `annotate.Agree`) compares, besides the eight fields of
format 1 (scheme, fit, role, polarity, target claim, target part, re-parenting, evidence):

| Field | Value compared |
|---|---|
| `cq` | The question answered (`role.cq`), or none |
| `target_instance` | The targeted instance's scheme and conclusion, read in the annotator's own file: ids are the annotator's own, so they are compared by what they name. A format-1 target names its claim's primary instance |
| `target_formula` | The exact formula targeted; a format-1 conclusion target is its instance's conclusion |
| `bindings` | The primary instance's bindings, every variable and its atom |
| `instances` | Every instance's scheme and conclusion: how the claim was divided into inferences |

A value that is unresolved, or that cannot be read in the annotator's file (a target claim it did not
annotate), is left out of that field's agreement and counted apart (`unresolved`): missing data is
neither agreement nor disagreement. With a case record, aliases are read as their actions first.
These fields are also the ones a run compares to decide which claims the reviewer reviews (*Review
the disagreements*): a claim whose annotations differ in their exact target or bindings is reviewed,
which it was not before this contract.

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
  an action the case does not list is simply another atom.
