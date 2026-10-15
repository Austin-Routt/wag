# Part 8 (2026-10-15): dialogue types, and claims merged by their text

> **Corrected after reading the primary source** (part 9, the same day). §1 below was written from Aberdein's survey alone. The owner then added Walton's *The New Dialectic* (1998), and checking §1 against it changed five points. Each is fixed in place and marked *(corrected in part 9)*:
> 1. **"Eight types"** is the list in Walton's 1997 table. The 1998 book keeps six main types: debate is mixed persuasion and eristic, and teaching is a kind of information-seeking. *(Refined in part 10: this is not a change of mind. Walton had already classed debate as mixed in 1982 and 1989, and A20 points to similar tables from 1990 and 1995.)*
> 2. **"A public debate is persuasion addressed to an audience"** is the 1997 table's "debate". The 1998 book calls political debate mixed discourse.
> 3. **Deliberation** does not need several options or the absence of a thesis. In the simplest case it reduces to one side for and one side against a proposed course of action. What it needs is a practical problem, shared goals, symmetric roles and a tentative joint decision.
> 4. **Kialo** policy debates are public persuasion about whether an action is advisable, not "debate" in the 1997 sense. *(Refined in part 10: earlier this said "a persuasion–deliberation hybrid". ND makes a hybrid of deliberation with conflicting basic goals; Kialo decides nothing jointly, so the book's p. 229 reading fits better.)*
> 5. **"Negotiation and eristic: not as modes"** is superseded by the owner's objective: every type, playable and findable. *(Part 10: the playable quarrel comes last.)*
>
> The full account is [`dialogue-types.md`](dialogue-types.md); the work it leads to is [plan §8](plan.md#8-track-d-dialogue-types).

The owner sent Andrew Aberdein's *Dialogue Types, Argumentation Schemes, and Mathematical Practice: Douglas Walton and Mathematics* (arXiv:2011.13845, 2020). The owner asked two things:
1. Should WAG classify games and theses by Walton's dialogue types? Is one of six definitive, and can types combine? Should the types be part of annotation and of choosing a game mode? What else in the paper matters?
2. Is it true that WAG merges different Kialo claims that have the same text, sometimes into a loop that makes it refuse the game, and should claims be keyed by ID?

**Short answers**

1. **Dialogue types.** Yes, but classify the *dialogue* (the game, or a Kialo discussion), not the thesis. And a type is only worth recording where it changes a rule.
   - **Not definitive.** The six types come from a grid that the paper itself says is not exhaustive. Types can be split or combined, and a real dialogue shifts between them and nests one inside another.
   - **WAG today.** Every WAG game is in effect a persuasion dialogue; its points modes match no type at all. Two context presets are named after types WAG does not implement: "Policy deliberation" and "Scientific inquiry".
   - **Annotation.** Record the type once per discussion. Per claim, a type is not worth adding yet.
   - **Other ideas.** Three more ideas from the paper improve the catalogue plan directly (§1.5).
2. **The merge bug.** Yes, the bug is real, and it is fixed in this commit: imported claims are now keyed by claim ID.
   - **Verified locally.** On the five discussions we have, two had different claims merged.
   - **The refusal reproduced.** A test makes the loop and the refusal happen on the old code.
   - **Your counts.** We could not check them here, but `wag-cli --kialo-check` reproduces them on your folder (§2).

## 1. Dialogue types

### 1.1 What the paper says

The paper reviews two of Walton's ideas, applied to mathematics: dialogue types and argumentation schemes. On dialogue types:

- **The six principal types.** Walton and Krabbe sort dialogues by two factors (its Table 1, after Walton & Krabbe 1995, p. 80):
  - the *initial situation*: a conflict, an open problem, or one party lacking information;
  - the *main goal*: a stable agreement, a practical decision, or a provisional accommodation.

  Six cells make sense, giving the six types; three are empty.

  | | Conflict | Open problem | Lack of information |
  |---|---|---|---|
  | Stable agreement | persuasion | inquiry | information-seeking |
  | Practical settlement | negotiation | deliberation | — |
  | Accommodation | eristic (quarrel) | — | — |
- **The list is not closed.** The paper states, citing Walton 1998, that the types can be subdivided or combined and that the list is not meant to be exhaustive. *(Corrected in part 9: the 1998 book gives the six with subtypes and treats combinations in its chapters 8–9; it does not itself say whether the list is exhaustive.)*
  - Walton's 1997 table (the paper's Table 2) lists eight types: it adds *debate* (persuading a third party) and *pedagogical*. It gives each type's individual goals, collective goal and benefits. *(Corrected in part 9: the 1998 book does not keep these two as main types. Political debate is mixed persuasion and eristic (p. 223), and pedagogical dialogue is a subtype of information-seeking (pp. 235–7).)*
  - The paper's own Table 3 adds *oracular* information-seeking, where the one who answers is a black box.
- **The type can change mid-dialogue (a dialectical shift).**
  - A shift can be productive. Krabbe describes a proof moving from inquiry to persuasion (verification) to information-seeking (teaching).
  - A shift can also be illicit, when it goes unnoticed and brings in tactics improper in the new type.
- **Dialogues nest.** A persuasion dialogue can sit inside an inquiry.
- **What is reasonable depends on the type.** What is fine in a negotiation may be improper in a persuasion dialogue.

### 1.2 Is one of six definitive? Can there be combinations?

**Not definitive.** The six are a systematic core, not a complete list. *(Corrected in part 9: an earlier version said Walton's lists "grew to eight". The eight are in a 1997 table; the 1998 book keeps six main types and treats the rest as subtypes or mixtures. Walton had already classed debate as mixed in 1982 and 1989, so the lists differ by presentation, not by a change of view.)*

**Combinations, yes, in three senses:**
1. **Nesting.** A persuasion dialogue can sit inside an inquiry, or an information-seeking exchange inside a persuasion.
2. **Shifts over time.** The type of a dialogue can change as it goes.
3. **Mixed settings.** A public debate is persuasion addressed to an audience. *(Corrected in part 9: that is the "debate" of Walton's 1997 table. In the 1998 book, as in Walton's work from 1982 on, political debate is mixed discourse: persuasion and eristic, usually with negotiation.)*

So one label per game is a simplification. The model that fits the literature:
- a *main* type for the whole dialogue;
- typed segments, for the nested and shifted parts;
- shifts recorded as events, licit or not.

### 1.3 The game, the thesis, the domain: three different things

| What | Examples | What it decides |
|---|---|---|
| **The dialogue's type** | persuasion, deliberation, inquiry | Roles, who bears the burden, what counts as winning (or succeeding together), which moves and schemes are proper, when it ends |
| **The thesis's kind** | practical ("governments *should* tax meat"), factual ("meat consumption causes…"), evaluative, classificatory | Its conclusion form, and so which scheme families can argue for it (§3 of part 7: the thesis is the conclusion of practical reasoning) |
| **The domain** | everyday, policy, science, law | Proof standards and scheme weights: WAG's context presets |

**The thesis does not fix the type.** "Governments should put a tax on meat" can be the subject of:
- a persuasion dialogue (two sides with opposite views, as on Kialo);
- a deliberation (a committee choosing among options);
- a negotiation (farmers and a ministry trading concessions).

The thesis's kind only narrows the types that make sense.

### 1.4 WAG today

| WAG | Initial situation | Goal | Closest type |
|---|---|---|---|
| `debate` mode | one thesis, two sides; pro bears the burden | pro wins if the thesis is In at the end | **persuasion**. With a moderator, a debate before a judge. |
| `cq` mode | the same, with points; an unanswered question loses | points | persuasion with strict duties to answer (a rigorous variant) |
| `solo` mode | one player argues both sides | the verdict; no winner | nearest to **inquiry** by one inquirer |
| `classic`, `infinite`, `target`, `timed`, `speed` | no shared thesis; each player scores their own arguments that stand | points | **none.** A competitive exercise that rewards arguments, not settling anything. |
| Kialo import | a public pro and con tree; readers vote on impact | persuade the audience | **public persuasion**. On a practical thesis, persuasion about whether an action is advisable (ND p. 229) (*corrected in parts 9 and 10*; was "debate"). |

**Two context presets name types the game does not play:**
- **"Policy deliberation"** is described as "a deliberation dialogue". But the game is adversarial, with a winner, and pro alone bears the burden. That is persuasion about a practical proposition.
  - *(Corrected in part 9.)* An earlier version said a deliberation must weigh several options with no side owning a thesis. That overstates Walton's 1998 account: deliberation typically compares alternatives (pp. 154–5), but need not have several on the table. A deliberation starts from a practical problem and shared goals. The parties are symmetric co-deliberators, and in the simplest case it reduces to one side for and one side against a proposed course of action (p. 162). It ends in a tentative joint decision.
  - The preset is now named "Policy question".
- **"Scientific inquiry"** is described as "an inquiry". But the game is adversarial, with a winner, under a higher proof standard. An inquiry is cooperative: both parties try to prove or disprove the claim. The preset is now named "Scientific question" (part 9).

The presets mix two dimensions: the dialogue type, and the domain's weights and proof standards.

### 1.5 Should it be part of game-mode selection?

Yes, as a setting separate from the domain preset, and only where it changes rules. A label that changes nothing invites a false sense of fidelity to Walton. By type:

- **Persuasion.** It is what WAG plays now. The setting would make it explicit.
- **Deliberation.** It is the type policy theses call for, and policy theses dominate Kialo. *(Corrected in part 9.)* It needs:
  - a practical problem and a shared goal agreed at the opening;
  - symmetric roles;
  - a joint, tentative decision as the outcome, with caution where harm is possible.

  One proposal argued for and against is the simplest case. Several options, each a position of one Carneades issue, is the general one (Carneades allows any number of positions; WAG's adapter handles only two today).

  The practical-reasoning, consequences and values schemes are its natural moves. This is days of work, not hours.
- **Inquiry.** It needs:
  - cooperative goals and shared scoring;
  - a higher standard by default;
  - success defined as the claim proven or disproven, with no concession or resignation.

  The `solo` mode is a start.
- **Information-seeking.** Already partly present: the question move, and requiring evidence. It is not worth a mode of its own.
- **Negotiation and eristic: not as modes.** *(Superseded in part 9: the owner's objective is every type, playable and findable. Negotiation becomes a mode with its own bargaining core, and eristic a detection exercise; see [plan §8](plan.md#8-track-d-dialogue-types). Updated in part 10: then a playable quarrel, last.)*
  - Negotiation is about resources and offers, not argumentation schemes. *(In part: its offers are backed and attacked with schemes, such as practical reasoning, threats and promises.)*
  - Eristic is worth *detecting* (a personal attack in a persuasion), not playing.

**One rule worth adopting early: scheme appropriateness by type.**
- This is Walton's pragmatic view of fallacy, where an argument is fallacious when it subverts the dialogue's goal.
- `walton.yml` has `threat`, `ad_baculum`, `fear`, `danger`, `help` and `distress`. These are legitimate in a negotiation and improper in a persuasion.
- Once a game declares its type, the validator or the moderator can flag them.
- The catalogue can carry "proper in" per scheme, drafted from the book.

### 1.6 Should it be part of annotation?

- **Per discussion: yes.** It is cheap: two fields filled once, by a person.
  - **The main dialogue type.** For Kialo it is always the same (a public debate), so it says little within Kialo. It matters as soon as other sources come in: deliberation transcripts, question-and-answer forums, peer review.
  - **The thesis's kind.** It is informative now. It predicts the families of the thesis's direct pros and cons. It also completes the fix for the thesis coded "none" (plan §3): the thesis gets its conclusion form and family.
- **Per claim: not now.** The existing roles already mark the usual shifts inside a persuasion:
  - `action_amendment` and `alternative_means` are deliberation moves;
  - `clarification` is information-seeking.

  A separate per-claim type would add a field with low agreement and no consumer yet. Map the roles to shift kinds in analysis. Add a per-claim tag only if a research question needs it.

  *(Part 9: that research question now exists. Finding every type in data needs typed segments and shifts, annotated at the level of a stretch of dialogue or a Kialo subtree, not each claim. See [plan §8](plan.md#8-track-d-dialogue-types), D3.)*

### 1.7 Other ideas in the paper to take up before going on

These concern the schemes, and they improve the catalogue plan ([`plan.md`](plan.md)) directly.

1. **Walton and Macagno's classification (2015)** (the paper's Table 4). It sorts schemes by a series of binary distinctions:
   - source-dependent or source-independent;
   - then practical or epistemic reasoning;
   - then, for epistemic reasoning, discovering a rule or entity, or applying a rule to a case.

   This is the decision procedure the annotation guidelines need for "which family, then which scheme". It is also a natural top level for the catalogue's one-line index. Walton and Macagno say it is incomplete (some linguistic arguments are missing), so our families sit under it, not in place of it.
2. **Defeasible modus ponens as the general scheme.** Walton argues that every defeasible scheme is a special case of it: data P; as a rule, if P then Q; so, presumably, Q. It has two critical questions: what backs the rule, and is this case an exception. It gives "other" a principled form:
   - **In annotation.** An "other" argument is coded as a defeasible modus ponens with its rule written out. It then has two critical questions instead of none, and a recurring rule is a candidate scheme.
   - **In the game.** The free "Other" scheme could take this form, so that its arguments can be asked for backing and exceptions, not only whether their premise is acceptable.

   `walton.yml` already has `defeasible_modus_ponens`.
3. **Hansen's definition of a scheme.** A scheme is a pattern of argument, made of sentential forms with variables, with at least one schematic constant, ending in a conclusion. This is a test that a derived scheme must pass before it joins the catalogue.
4. **Localisation.** Aberdein turns the ethotic scheme into an "ethotic mathematical" one (by changing one word), and suggests sharpening the expert-opinion questions for mathematics. A recurring "other" pattern may thus be a *domain version* of an existing scheme, not a new one. The catalogue needs a "specialises" link, and the candidate review needs that outcome.
5. **Illative neutrality and qualifiers.** The same scheme can be used deductively or defeasibly, with qualifiers such as "presumably" or "certainly" that differ from one instance to another. Annotation could record the qualifier. It bears on the book's modal cause to effect ("will (might) occur", audit N8) and on argument weights.
6. **Expert opinion against position to know** (expert witness against fact witness). A contrast the catalogue entries must make explicit.
7. **Commitment.** Walton and Krabbe's dialogues track each party's commitments: what they have conceded, and what they may retract. WAG has no commitment store. Schemes such as `inconsistent_commitment` and the circumstantial ad hominem depend on one. It would come with a proper persuasion protocol, later.

### 1.8 How helpful is the paper?

- **Useful as a guide** to the dialogue-type framework and to the classification of schemes, with the right primary sources:
  - Walton & Krabbe 1995;
  - Walton 1998;
  - Walton & Macagno 2015;
  - Walton, Reed & Macagno 2008.
- **Secondary.** It is one author's survey, aimed at mathematics; its mathematical content has no use in WAG today.
- **For definitions, go to the primary sources** before building rules on them.

### 1.9 Recommendation

| Step | When | Size |
|---|---|---|
| Put Walton and Macagno's tree, defeasible modus ponens for "other", Hansen's test, "specialises" links and "proper in" into the catalogue plan | Now (done in `plan.md`) | Planning |
| Record the dialogue type and the thesis's kind per dataset; the thesis's kind in guidelines version 4 | With phase 2 | Small |
| Record the dialogue type in game settings and exports (persuasion for every current mode); split the context presets into dialogue type and domain, and correct the two descriptions | After the catalogue, or whenever the game is next worked on | Small |
| Flag schemes improper for the game's type | After "proper in" is in the catalogue | Small |
| A deliberation mode (several options, Carneades issues with more than two positions) | After phase 2 | Days |
| An inquiry mode; commitment stores | Later | Days each |

The catalogue stays the first priority. Dialogue types change the game more than the annotation. *(Superseded in part 9: this table is replaced by [plan §8, Track D](plan.md#8-track-d-dialogue-types).)*

## 2. Claims merged by their text

### 2.1 Was it true?

**Yes.**
- **How the old import keyed statements.** The Kialo import (`internal/kialo/build.go`) made each claim a free-text statement, whose ID is a hash of the text: lowercased, spaces collapsed, a final period dropped (`carneades.FreeStatementID`).
- **What that did.** Two different claims with the same words became one statement, and everything argued about one counted for the other.
- **Why the loop check missed it.** The import's plan (`Plan`) checks for loops by claim ID, so it could not see a loop made by merging. The merged graph then failed the final check, and the game was refused with "the imported graph has a cycle".

**Measured on the five discussions we have** (`wag-cli --kialo-check`):

| Discussion | Claims | Relations imported | Claims sharing a text | Old import |
|---|---|---|---|---|
| 13292 | 102 | 86 | 0 | the same game |
| 13564 | 388 | 345 | 0 | the same game |
| 2629 | 8,319 | 3,690 | 113 | two claims merged into one statement |
| 27150 | 189 | 169 | 0 | the same game |
| 5637 | 3,253 | 437 | 6 | two claims merged into one statement |

- **What the two real merges joined.** Each joined cons of two different parents. So the arguments under one claim also counted for the other.
- **No refusal here.** None of the five was refused. A loop needs the duplicate to sit below the other in the same branch.
- **The refusal reproduced.** A test builds such a graph (a claim, a con of it, and a pro of the con with the same words as the claim). On the unchanged code the import fails with "the imported graph has a cycle".
- **Your counts.** We cannot check "297 of 461 records" and "486" without your corpus. As written they cannot both be counts of records, since 486 is more than 461. `wag-cli --kialo-check` on your folder gives:
  - the records with claims that share a text;
  - the records the old import refused;
  - the records that fail now (none expected).

### 2.2 The fix

| Part | What changed |
|---|---|
| Statement IDs | **One statement per claim, keyed by the claim's ID:** `claim_` plus the claim ID, with its dot turned into an underscore, so `13292.7` becomes `claim_13292_7`. The statement keeps the claim's text, and its metadata holds the claim ID. Its English description is its text, as for free text (`carneades.IsTextID`). |
| Free text | **Still one statement per wording.** This is what players expect when they type the same sentence twice. |
| Saved games | **They replay as they were played.** The import report now records `"keys": "claim"`. An export without it is rebuilt with claims merged by text. Checked on an export made with the old code: it loads and replays its moves. |
| The new command | **`wag-cli --kialo-check DIR`** imports every `<id>/graph.json` of a folder. For each discussion it reports the claims that share a text, whether it imports now, and what the old import made of it. Its exit status is 1 if any discussion fails. |
| Tests | Two claims with the same text stay two statements, and the loop case imports. An old-style rebuild of the same graph is refused, as before. The server imports, plays a move on a claim statement, exports, reloads, and reloads an old-style export with the claims merged. The check command has its own test. |

**What the fix does not change:**
- **Annotation datasets** (`Discussion.Items`) were always keyed by claim ID.
- **`to_wag.py`** plays a new game from the gold annotation, not an import.
