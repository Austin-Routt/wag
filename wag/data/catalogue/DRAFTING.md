# Drafting and verifying a catalogue entry

How to write the work fields of an entry in `schemes/<id>.yml`, for a model or a person. The format, the statuses and the checks are in the [README](README.md); this guide says what to write in each field. Two entries are drafted as models of the style: [`expert_opinion`](schemes/expert_opinion.yml) and [`cause_to_effect`](schemes/cause_to_effect.yml).

## Rules

1. **Never copy the book.** Write every text in your own words. Short labels are fine ("Expertise Question", "Major Premise"), but not the book's sentences, and not lightly reworded ones either. Do not reuse the book's examples. The copy check is a floor, not the test.
2. **Never edit a source field**: `walton`, `game`, `formulas`, or the `game` part of a question. `--catalogue-gen` rewrites them. If a source is wrong, record a discrepancy (below).
3. **The book decides what a scheme says and asks; `walton.yml` decides the formulas; the NASCQ decides the game's questions** (the README's "who decides what"). When they disagree, record it and say what the catalogue does about it. Do not settle it by quietly writing the text to fit one source.
4. **Set the status of each item you write** to `drafted by WHO, YYYY-MM-DD`. A model writes its tool's name, for example `drafted by Claude Code, 2026-10-22`. Only a person who has checked an item against the book sets `verified by WHO, DATE`. A person who thinks an item is wrong sets `disputed by WHO, DATE: REASON`. Leave an item you have not touched as `generated` or `todo`.
5. **Add one line to `history`** for each session of work on the entry: date, who, what.
6. **Run the checker with the book's text** before you stop. Fix every error, and every copy warning:

   ```sh
   go run ./cmd/wag-cli --catalogue-check --book BOOK.txt,../docs/book-catalogue/catalogue.json
   ```

   `BOOK.txt` is the transcription of the compendium (ch. 9), which is kept outside the repository. Any plain-text file of the book will do.

## Reading an entry's sources

- **The book:** the compendium entries in `book` (and the entry named in `cqs_printed_under`, if any), and the chapters that discuss the scheme. The book's variable letters often differ from `walton.yml`'s (the book's E, S, A are `walton.yml`'s W, D, S for expert opinion). The catalogue writes `walton.yml`'s letters.
- **`walton.yml`:** the formulas in the entry, and their English readings (`wag-cli --catalogue-show ID` prints them; they often repeat the book, so do not copy them). The upstream collection (`walton-argumentation-schemes-for-carneades-4-master`) also has 13 category files, which hold older or different versions of some schemes. When one differs from `walton.yml`, record it as a discrepancy of kind `other`.
- **The game:** `wag/data/nascq/` (schemes, critical questions, ask and backup edges) and `wag/data/schemes/core.yml`. The `game` part of each question is the NASCQ's.
- **The Kialo gold:** `docs/kialo-batch2/gold.json` codes 59 claims of discussion 13292; the claims' texts are in `items.json`.

## The fields

### Book links, families and premise maps

These are generated from `links.json` and `families.json`, which a model drafted. Check them while you read the book.
- **Right:** leave them `generated`. A person verifying the entry marks them verified.
- **Wrong:** correct the item, set it to drafted, and say what you changed in its `note`.

### `summary`

**Two sentences:**
1. When the scheme applies: the situation and what the arguer has to go on.
2. What it concludes, and what makes it this scheme rather than its neighbours.

Use the formulas' variables where they help ("an expert W in domain D"). No book wording.

### `critical_questions`

There is one item per question. Book questions are `<entry>/<label>`; game questions that match no book question are `game/<NASCQ id>`.

**`targets` and `undercuts`.** Decide what the question attacks:
- **One or more formulas, with their roles:** `premise`, `assumption`, `exception` or `conclusion`. The checker holds each to `walton.yml`.
- **The inference itself:** set `undercuts: true` and give no targets. Use this when the question asks whether the premises, even if true, support the conclusion here: a missing link, a better explanation, a special circumstance.
- **Something no formula encodes:** leave the targets empty, and record a `book-walton` discrepancy saying which formula is missing.

The checker warns about every book question that has neither targets nor `undercuts`, until it is decided.

**`paraphrase`.** The question in our own words, with `walton.yml`'s variables: one sentence, ending in a question mark. It must ask what the book's question asks, no more. If the book's question is vague, say so in `note`, not by sharpening the paraphrase.

**`answer`.** What counts as an answer, and what each answer does. Write it in this shape:

> `<who carries the burden>. Against: <what shows the negative>, which <undermines | undercuts | rebuts> the argument. For: <what shows the positive>.`

The burden follows the game's rules (`docs/game-rules.md` §2), which follow Carneades:

| Target | Who carries the burden | A negative answer |
|---|---|---|
| A premise or assumption | Asking shifts the burden: the proponent must back it. | Undermines |
| An exception | Whoever raises it must prove it: it is raised with an argument, not asked. | Undercuts |
| The conclusion | Whoever raises a counterclaim must prove it. | Rebuts |
| The inference (`undercuts`) | Asking puts the argument out of force until the proponent answers. | Undercuts |

For example:

> The proponent, once asked. Against: evidence that W has no training, record or standing in D, which undermines `expert(W,D)`. For: W's qualifications, publications or record in D.

**Questions only `walton.yml` implies.** An assumption or exception that neither a book nor a game question is about can get its own question, with:
- the id `walton/<predicate>` (for example `walton/valid` for `valid(R)`);
- one target;
- no `book` or `game` part.

Draft these only for assumptions and exceptions. A premise can always be challenged in the game, so it does not need one.

**`note`.** Keep the generated note (the book's label for the question). Add anything a verifier should know.

### `contrasts`

There is one item per close relative; the generator lists the family's generic scheme and the registry's close pairs.

**Write one or two sentences** that say when to choose this scheme rather than the other, by what the text gives you:

> "Choose expert_opinion when the source's authority comes from knowledge of a field (credentials, research); choose position_to_know when it comes from access to the facts (being there, holding the records)."

**Drop a contrast only if it is wrong.** If the relative is not close in any useful way, set the status to drafted and say so in the text: "Not a close relative: …".

**Add a contrast** for a relative that the annotators confuse, if you know of one.

### `examples`

Give two or three examples, and one near miss.
- **Use Kialo claims where you can:** the batch-2 gold first (set `kialo` to the claim's id), then other claims of discussion 13292. Quote the claim's text without its markdown links. You may shorten it with "…".
- **Otherwise,** write a short example of your own, one or two sentences on an everyday or policy topic. Never use the book's examples.

**`coding`** gives the bindings, then which premises the text states and which the coder supplies, then the conclusion:

```
W = the 2020 Nature study; D = the external costs of food; S = greenhouse-gas costs are highest for animal products.
Stated: asserts(W,S). Supplied: expert(W,D), in_domain(S,D). Conclusion: S.
```

- **The gold's bindings.** Keep them unless you disagree; if you do, say why in the coding.
- **Exceptions.** Do not list them as premises: they are what a critic would raise.

**A near miss** (`near_miss: true`) is a text that looks like this scheme but is another one. Its coding names the scheme it is, and why:

```
Near miss: it cites a source, but for what she saw, not for knowledge of a field; position_to_know (W = the inspector).
```

### `discrepancies`

The generator records the ones the sources show:
- `partial-link`: the scheme formalizes the book entry only in part;
- `not-in-book`;
- `game-differs`: the game's copy of the scheme differs from `walton.yml`;
- `link-formula`.

**For each of these**, write the `decision`: what the catalogue does about it, and what should change, and where (the formula, the link, the game). Set the status to drafted.

**Add the ones you find:**
- **`book-walton`:** the book and `walton.yml` say different things. For example:
  - a premise or qualifier is lost (modal words like "probably", time indices);
  - a variable is bound differently;
  - a book question has no formula;
  - a formula has no book basis.
- **`other`:** an upstream category file's version of the scheme differs from `walton.yml`, or anything else.

Give `about` a short key (`CQ2`, `premise 2`, `C_applying-rules-to-cases/2`). In `detail`, say what differs, in our words. A decision you cannot make yet starts with "Open:".

### `classification`

Place the scheme in Walton and Macagno's tree (2015, p. 22; reproduced in Aberdein 2020, Table 4). The `group` must be one of the tree's groups under its heading:

| `source` | `reasoning` | `epistemic` | `group` (one of) |
|---|---|---|---|
| `dependent` | (empty) | (empty) | `position to know`, `ad hominem`, `popular acceptance` |
| `independent` | `practical` | (empty) | `instrumental`, `values`, `value-based` |
| `independent` | `epistemic` | `discovery` | `establishing rules`, `finding entities` |
| `independent` | `epistemic` | `application` | `based on cases`, `defeasible rule-based`, `chained` |

The tree names these schemes under its groups:
- **Discovery.** Establishing rules: random sample to population, best explanation. Finding entities: sign, ignorance.
- **Application.** Based on cases: established rule, verbal classification, cause to effect. Defeasible rule-based: example, analogy, precedent. Chained: gradualism, precedent slippery slope, sorites slippery slope.
- **Practical.** Instrumental: practical reasoning, action to motive. Values: values, fairness. Value-based: positive or negative consequences, with waste, threat and sunk costs under them.
- **Source-dependent.** Position to know: expert opinion; position to know, with witness testimony. Ad hominem: direct; circumstantial, with inconsistent commitment; attacking credibility, with bias and poisoning the well. Popular acceptance: popular opinion, popular practice.

**When the tree does not list the scheme,** place it by its nearest listed relative, and say so in `note`: "Not in the tree; placed with …, because …". The checker allows a group only under its own heading.

### `specialises`

`specialises` names the scheme this one is a domain or narrower version of. Examples:
- witness testimony specialises position to know;
- practical_reasoning2 specialises practical_reasoning1, if it does.

Its `note` says what the narrowing adds. If it specialises nothing, leave `scheme` empty, say so in the `note`, and set the status to drafted.

### `dialogue_types`

`proper` lists the dialogue types in which the scheme is a legitimate move. `suspect` lists those in which it is suspect or fallacious. The names are:
- `persuasion`
- `inquiry`
- `information_seeking`
- `negotiation`
- `deliberation`
- `eristic`

**Start from the table** in [`docs/dialogue-types.md` §7](../../../docs/dialogue-types.md#7-arguments-whose-standing-depends-on-the-type), which cites *The New Dialectic* (ND) by page.

**For a scheme the table does not cover, reason from each type's goal:**
- Practical schemes (consequences, values, practical reasoning) answer "what should we do?". They belong in deliberation, negotiation, and persuasion about action.
- Epistemic schemes (cause, sign, example) give evidence. They belong in inquiry, persuasion and deliberation.

The `note` says which parts ND states, with pages, and which are our extrapolation. Do not cite a page you have not checked.

## Verifying

A person who verifies an entry does the following:

1. Reads the compendium entry and the chapters that discuss the scheme.
2. Checks each item: links, premise maps, questions, targets, paraphrases, answers, contrasts, examples, classification, dialogue types and discrepancies.
3. Sets each item to `verified by WHO, DATE`, or `disputed by WHO, DATE: REASON`, correcting what is wrong.
4. Adds a line to `history`.
5. Runs the checker.

Verifying an entry should take 20 to 30 minutes with the book open (plan §2).

A model may check drafts, as the independent check of 2026-10-22 did. A model's check is recorded in `history`, not as `verified`: only a person verifies.
