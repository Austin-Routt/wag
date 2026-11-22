# The outcomes of 2026-11-22: the external review's dispositions (ChatGPT's
# convergence review), adopted under the owner's standing rule where the
# developer agrees, changed where it does not (the reason is given), and
# left open where a person must verify. Data only; applied by apply_decisions.py.

S2 = "stage 2: walton.yml through edits.py, after a proposal check and an independent review (plan §0)"
APPLY = {
    "scheme": S2,
    "vocabulary": S2,
    "game": "the NASCQ generator and the game's data (plan §0, stage 2)",
    "annotation": "the guidelines, the gold and the annotation tools (plan §0, stage 2)",
    "metadata": "walton.yml's meta through edits.py (plan §0, stage 2)",
    "project": "docs/plan.md and docs/process.md",
}

D = {
"D002": "Sign's exception names the finding and the proposition it would indicate: more_reliable(A,B), as the schemes library has it, so that an alternative explanation proven against one sign argument does not defeat another. The alternative event is the answer's witness.",
"D003": "The questions stay exceptions, the critic's to support. Who must answer a question that is asked is a rule of the dialogue (the request, its answer, and what holds while it is pending), set by the game's question protocol with D075, not by turning exceptions into assumptions.",
"D004": "practical_reasoning1 gets a conflicting-goal exception, with the predicate practical_reasoning2 already has (incompatible_goal), scoped to the agent and the goal, so that the first question and CQ:other_goals target it. (The review proposed an unnamed objection; reusing the existing predicate keeps one formula for one idea.)",
"D005": "Both practical reasoning schemes state the agent's goal as a premise, goal_of_agent(P,G), the schemes library's predicate. The conclusion stays should_be_performed(A) (D059, D063).",
"D006": "No. An equally good alternative is not a reason against doing A; it answers only a claim that A is necessary or the best (the necessary-condition schema, realize_more_effectively). practical_reasoning2's conclusion is unchanged.",
"D007": "The book's ch. 9 version, the compendium, which is the catalogue's authority (plan §1). Ch. 3's further questions are kept as derived questions, each with its source (D023). (The review left the chapter unchosen.)",
"D008": "Yes, as schemes of their own: modal variants of the consequence schemes, as cause to effect has cause_to_effect_modal, not a modality field. Likelihood stays a question (33/CQ1, CQ:likelihood). (The review left variant or field open and added an adequacy assumption; WAG's precedent is a scheme of its own, and the likelihood question already exists.)",
"D009": "Yes: the generated NASCQ takes its question texts from the catalogue's paraphrases, with the changes the entries list, except the backup edge that lets negative consequences answer the side-effect question, which D070 and D075 replace. CQ:field is kept. Every generated question gets a test that its formula target is one of its scheme's.",
"D010": "Yes: position_to_know gets an exception for a source who is sincere but not reliable on the claim, scoped to the claim's domain, beside dishonest(W).",
"D011": "No. popular_opinion keeps reason_in_favor(A), as the book concludes. Where a game needs it to meet A, a bridge scheme connects them, with its own exception (D075's rule for bridges).",
"D012": "The duplicate goes: one acceptance premise stays, named so that it reads as accepted, not as true.",
"D013": "No: general acceptance is a premise, as the compendium prints it (4). A context that presumes it is a later option of the game, not a change to the scheme.",
"D014": "from_example generalizes from the case: premise has_properties(A,P1,P2), conclusion usually_properties(P1,P2), with the typical-example assumption, as the schemes library has it after its check against the book (6.1).",
"D015": "Yes: alternative_cognitive gets an exception for an alternative not considered, the critic's to show.",
"D016": "No: not required is not should not be done. alternative_normative concludes the negation of is_required(B), which meets only is_required(B).",
"D017": "Exclusiveness is stated in the premise, as the book's either-or premise reads and the schemes library has it (exclusive_requirement_alternatives(A,B)): the proponent's to show, not an assumption.",
"D018": "Yes: the target of correlation's second question is widened to a reason against coincidence, with a predicate of its own and its migration recorded.",
"D019": "Yes: defeasible modus ponens gets an exception for the case at hand, with a template of our own; its questions are labelled derived, not the book's (D021).",
"D020": "Yes: defeasibly_implies(A,B) is its rule premise.",
"D022": "Both, as the book prints both readings (22.6.1): a negative variant (a bad end condemns its means) and a graded one, as the schemes library has them.",
"D023": "Yes, as derived questions, each with its source scheme and question recorded.",
"D024": "established_rule gets the two exceptions, an overriding rule and an exceptional case, as the schemes library has them (55.1). rule (55.2) keeps its form; questions from 55.1 apply to it only as derived questions. (The review left the second half vague.)",
"D025": "Yes: a variant naming the agent and the action type, beside the generic form, as the schemes library's established_rule_actor.",
"D026": "Yes: an evaluative and a prescriptive form, as the book prints both (55.2).",
"D030": "Yes, the means: the schema separates possible_goal and possible_bring_about(P,X), as the schemes library has it, so that no variable is left unbound. Its conclusion follows D063.",
"D031": "Yes: the commitment form becomes a scheme of its own, as the schemes library's sunk_costs_precommitment (compendium 25).",
"D032": "An exception for a doubtful classification, doubtful_classification(F,G).",
"D033": "An exception, the critic's burden, as D055's outweighing exceptions are: reassessment_against_continuing(A,X), as the schemes library has it.",
"D034": "Yes: in_domain(S,D) in witness_testimony too, so that one game question asks about both.",
"D035": "Yes: the testimony is a variable of its own (account_contains(T,S)).",
"D036": "The first: causes(E1,E2) is read as a tendency, strong or weak, and each scheme's first question asks whether it is strong enough for its conclusion, as the book's question on the strength of the causal generalization does. The modal schemes cover what may happen (D008). (The review proposed links typed by strength; the book carries strength in its question, and a strength argument would change every causal scheme.)",
"D037": "No to will_occur supporting has_occurred. will_occur may support may_occur only through a scheme of its own, in the same context. Chained steps meet through D038.",
"D038": "A chain of causal steps backs brings_about(A,G) only through an explicit scheme that keeps the steps' modality and context (a bridge under D075); otherwise it stays a secondary instance, as the gold codes it.",
"D039": "Yes, in a new version of the gold, with guidelines item 11 and validate.py's R4 changed with it; the version before is kept and scores against it stay comparable.",
"D040": "Yes: cause_to_effect cites the book's entry 28 (p. 328), and the reference it prints (A Pragmatic Theory of Fallacy, 1995, pp. 140-141) as cited there.",
"D042": "Rewrite division to run from the whole to the parts (has_property(A,P) to all_parts_property(A,P)), with its own heredity assumption; composition stays as it is. The registry's tolerated error goes.",
"D043": "No book text in the public repository, as plan §1 says: the transcription of pp. 1-138 moves to the private workspace; the copy check keeps working from a file of hashed word sequences, which holds no text; quoted questions become paraphrases; walton.yml's templates that repeat the book are rewritten in our words. Rewriting the public history is D078. (The review allowed scoped excerpts in public; that contradicts plan §1.)",
"D044": "MPL-2.0 for WAG's own code, the licence of Carneades and walton.yml, added at the first release. Third-party material (the book's labels and paraphrases aside, Kialo data, upstream files) keeps its own terms. The documents' and data's licence is D080.",
"D045": "A person verifies, as process §2 says; tier 1 is the first set. Models check (the drafted and checked statuses), and work may go on from checked entries, its results marked provisional, before a person has verified them (plan §0). (The review proposed that model review verify; that contradicts plan §1-2.)",
"D046": "Yes, for bounded runs after the integrity fixes of 2026-11-22: each run has hard caps on calls, tokens and cost, and stop conditions, set by the owner for that run; runs with the owner's keys run on the owner's machine (process §3).",
"D047": "The threshold stays (three sightings in two discussions or games) as the rule for bringing a candidate to the owner, who decides; archived and strained sightings count, and are reported apart from clean ones.",
"D048": "A negotiation profile of its own, scored per issue (a vector, not one number), when Track D reaches negotiation.",
"D049": "An experimental profile of its own, played by model seats on fictional scenarios; people only by choice; with plan §8's content limits, a moderator, and synthetic transcripts labelled as such.",
"D050": "Real discourse first, as plan §8 says: Kialo and other discussions whose licence allows research use, each recorded with its licence; WAG's own game traces only as a second source. (The review proposed WAG's traces first; finding dialogue types in our own synthetic play would be circular.)",
"D051": "The denominators are set now: assigned and attempted items, the moves players intended, the moves made, matches, and unknowns. Target rates are set after the first pilot measures them, not before.",
"D054": "Yes: falsification concludes the negation of H, and its assumption reads sufficient_correlation(H,A), as verification's does.",
"D056": "Two patterns of their own, labelled WAG extensions (not compendium schemes), through the proposal path (process §5).",
"D057": "Only questions at a premise, assumption, exception or conclusion locus, by explicit backup edges, and CQ:other_case by the counter-analogy; no blanket thesis independence. Argument-locus questions need an explicit target and a fitting rule (D075).",
"D058": "Yes: the exception compares the precedent case with the current one, naming both, as D001's two-case analogy does.",
"D060": "Yes: cited_bad_consequences(X,C), position_to_bring_about(C), and assert_bad(X,C) with a template that prints C before X (indexed templates, available since 2026-11-22). Its conclusion follows D063.",
"D062": "Yes: danger concludes about the action X: the negation of should_be_performed(X), by D063. Its templates are corrected with it.",
"D063": "Recommendations of an action that name no agent but the one addressed conclude should_be_performed(A) or its negation, as D059 and D029 did: fear, danger, threat, the pragmatic argument from alternatives and the disjunctive threat, among others. Conclusions that name an agent (help, distress, two-person practical reasoning, practical_goal's should_do(P,X), waste) keep the agent, and a bridge scheme with its own exception (the agent is not the one deliberating) connects them to should_be_performed(A). Requirements and permissions stay apart (D016). (The review proposed sharing only identical propositions and bridging the rest; this applies that rule with D059's precedent.)",
"D064": "C003 goes forward (add: a scheme that concludes D055's exceptions from a consequence of the other sign that matters more, the first answer scheme under D075). C004, C010, C013, C019, C025 and C026 stay candidates until their premises are written; no summing of claims.",
"D065": "C005 and C021 go forward as questions on the consequence schemes (how serious, and on whom); the likelihood question already exists. C008 (precaution) and C024 stay candidates.",
"D066": "Mitigation is an amended action (C006); another action with the same good at less cost (C017) is a separate alternative; feasibility (C018) is a challenge to the premise.",
"D067": "Rights and values stay apart from consequences: C007 is argued with the value schemes; a value that trumps whatever the consequences needs a context rule of its own, later.",
"D068": "A scheme of its own, with a prospective conclusion that is never a practical claim, through the proposal path.",
"D069": "C015 stays a coding rule (OC13: positive consequences of the action, with the harm's not happening as G), which D059 and the guidelines of D055 already give; it does not become a scheme. C016 (would doing the action avert the harm) becomes a question of 33.4; C020 is covered by OC12; C023 (doing against allowing) stays a candidate. (The review proposed a new omission scheme through D027; that duplicates 33.4 and OC13, which the owner decided.)",
"D070": "Yes: a plain side effect is argued as negative consequences, a rebuttal. The exception is renamed for a side effect claimed to outweigh the reasons for the action, with the comparison as a premise of its answer (D075), so that old weak side-effect facts are not read as the stronger claim. The gold's two side-effect answers are recoded in a new version.",
"D071": "Yes: interference(E1,E2) in cause_to_effect and cause_to_effect_modal.",
"D072": "Yes: a scheme of its own, with an exception for the outcome being preventable after all, through the proposal path (C022).",
"D073": "Yes, in a tested replay command of the game: the moving side is composed along the target chain to the thesis; ambiguous, multi-parent and cyclic cases are flagged, not guessed; every argument of a claim is played; a replay is a reconstruction, not an observed dialogue.",
"D074": "A question or undercutter for the consequence schemes, raised when a consequence argument is offered for the truth of a claim; it goes forward once C011's sketch and examples pass review.",
}

# Decisions that stay open: a person must verify (process §2). What the
# models found is added to the question.
KEEP_OPEN = {
"D021": "Checked on 2026-11-22 by two models (the external review and the developer's check), not yet by a person: pp. 364-367 treat the form and give the extra premise that the rule has no exception (p. 366), and print no numbered questions for it. If a person confirms, its questions are derived (D019).",
"D027": "Checked on 2026-11-22 by two models, not yet by a person: 33.3's first premise on p. 332 carries the parenthetical and the conclusion stays negative. Whatever the reading, no omission scheme is added: 33.4 and OC13 cover not acting (D069).",
"D041": "Checked on 2026-11-22 by two models, not yet by a person: the compendium prints 2007 (p. 311). Once a person confirms, walton.yml cites 2007, as the slope's year was corrected; the publisher's later printing (2008) is noted in the CHANGELOG.",
}

NEW = [
{"id": "D075", "title": "Answering a question with an argument", "category": "game", "blocks": ["game"], "priority": 1,
 "question": "How does an argument answer a critical question whose answer is an exception or an assumption? Today a backup edge may carry a sentence (a bridge) that lets an argument stand for an answer it does not conclude, and nothing on the board can be challenged. The external reviews of 2026-11-22 (Fable, ChatGPT) and WAG's review of 15 October (part 6 §4, the schemes bad_side_effect and better_means, never registered) propose answer schemes: schemes whose conclusion is the answer.",
 "docs": ["docs/review-2026-10-15-part6.md §4", "docs/game-rules.md", "the external reviews of 2026-11-22 (wag-workspace reviews/2026-11-22)"],
 "decided": "Answer schemes. An answer scheme concludes the exception or assumption it answers; its premises name the witness (the harm, the other action) and bind every variable of what it answers; it has its questions (the book's, derived ones, or a reasoned none). A free-text bridge is a lint error. Each backup edge has a test answer that must pass, one that must be refused, and one bound to the wrong target. The first answer schemes are C003's (D064), the side effect that outweighs (D070) and another means that does better, with part 6's sketches repaired (their side effect left a variable unbound). D009, D038, D057 and D063's bridges follow this rule.",
 "recorded": "docs/decisions.yml; to apply: " + S2},
{"id": "D076", "title": "Evidence as records", "category": "game", "blocks": ["game", "annotation"], "priority": 2,
 "question": "A player's evidence is a list of strings, and the evidence rule checks only that one is there. What does WAG record about a source, and what does admitting it mean?",
 "docs": ["the external reviews of 2026-11-22 (ChatGPT's evidence design)", "wag/internal/server/protocol.go"],
 "decided": "A source is a record: what it is (a URL, a file, a book page), the passage or region used, the claim drawn from it, the premise it is offered for, each party's assessment, and whether it was admitted, by whom and under which rule. Consent and a moderator's admission are commitments, not verification. Built in slices from stage 2; the four game profiles of the review (cooperative, adversarial, moderated, automated) wait for Track D. A moderator's decision on a statement is a hard assumption today, and is labelled so.",
 "recorded": "docs/decisions.yml; to apply: the game's moves and export (plan §0, stage 2)"},
{"id": "D077", "title": "Evidence captured for research", "category": "project", "blocks": ["annotation"], "priority": 2,
 "question": "Should the evidence cited in the data (the links of Kialo claims, the sources players cite) be captured whole, and where is it kept?",
 "docs": ["the owner's notes of 2026-11-22", "wag-workspace README.md"],
 "decided": "Yes, for research purposes, under the fair use doctrine (the owner, 2026-11-22): the whole source as cited, and the part each reader used, with every assessment. It is kept in the private workspace, large files in Git LFS, never in the public repository. First the links of the claims a pilot annotates. Terms of use, privacy and injected instructions are D079.",
 "recorded": "wag-workspace README.md; docs/plan.md"},
{"id": "D078", "title": "Rewriting the public history", "category": "project", "blocks": ["release"], "priority": 2,
 "question": "The public repository's history still holds the deleted PDF of The New Dialectic (commit f891090) and the transcription of pp. 1-138 of Argumentation Schemes. Should the history be rewritten to remove them? It changes every commit's identity and breaks other clones.",
 "docs": ["docs/decisions.yml D043", "README.md"]},
{"id": "D079", "title": "Terms of use, privacy and injected instructions in captured evidence", "category": "project", "blocks": ["release"], "priority": 3,
 "question": "Captured pages come with sites' terms of use, may hold personal data, and may carry instructions aimed at the models that read them. How are these handled? The owner's note of 2026-11-22: to be reviewed and decided later; capture for research goes on meanwhile (D077).",
 "docs": ["the owner's notes of 2026-11-22", "docs/decisions.yml D077"]},
{"id": "D080", "title": "A licence for WAG's documents and data", "category": "project", "blocks": ["release"], "priority": 3,
 "question": "Under which licence may others use WAG's documents, catalogue and annotated data (the code is MPL-2.0, D044)? Kialo data and the book keep their own terms.",
 "docs": ["docs/decisions.yml D044"]},
{"id": "D081", "title": "Four repairs awaiting approval", "category": "scheme", "blocks": ["verification"], "priority": 2,
 "question": "The cleaning of walton.yml made four repairs that still await the owner's approval (CHANGELOG, \"Repairs that need the owner's approval\"): rhetorical_oppositions_normative, sorites_slippery_slope (and its template), moral_justification_ad_populum_excuse and sacrifice. The schemes library has its own forms of the first two. Approve, revise from the library's form, or revert each?",
 "docs": ["docs/scheme-registry/walton-clean/CHANGELOG.md", "docs/scheme-registry/walton-clean/changes.json"]},
{"id": "D082", "title": "What a full NASCQ means", "category": "project", "blocks": ["game"], "priority": 1,
 "question": "The goal is that every scheme and question be playable in some context. What does full mean, and how is it measured?",
 "docs": ["docs/plan.md", "the external reviews of 2026-11-22 (Fable's addendum, ChatGPT's NASCQ review)"],
 "decided": "Two things. Within a versioned set of schemes and questions, every question can be asked and every answer the set offers has a tested contract (target, binding, effect, refusal, replay): this is checked. Beyond it, completeness is measured, never proven: the plan's two measures, and two more: the share of the answers found in coded data and in play that can be played as meant, and the share of answers that should be refused that the game refuses.",
 "recorded": "docs/plan.md (the goal)"},
]

NEW.append(
{"id": "D083", "title": "Candidates kept after the decisions of 2026-11-22", "category": "scheme", "blocks": ["play"], "priority": 3,
 "question": "D064, D065 and D069 sent some candidates forward and kept these for later: C004 (a balance of all consequences), C010 (the lesser evil), C013 (the greater good), C019 (accrual), C025 (no one else made worse off), C026 (the chooser's own weighing), C008 (precaution), C024 (a certain harm against a speculative one) and C023 (doing against allowing). Which go forward, and as what, once their premises are written?",
 "docs": ["docs/candidates.md", "docs/decisions.yml D064, D065, D069"]})

REC = "docs/decisions.yml"
# Candidate outcomes (process.md §8: reject, merge, domain, add), or a new decision.
CANDIDATES = {
"C003": ("add", "", REC + " D064, D075: the first answer scheme, proposed in stage 2"),
"C005": ("merge", "positive_consequences and negative_consequences, as a question (how serious)", REC + " D065"),
"C021": ("merge", "positive_consequences and negative_consequences, as a question (on whom)", REC + " D065"),
"C006": ("merge", "the guidelines' action_amendment role (kialo-batch2/guidelines.md)", REC + " D066"),
"C017": ("merge", "practical reasoning's better-means exception, answered by D075's answer scheme", REC + " D066"),
"C018": ("merge", "practical reasoning's possibility assumption, possible(A); against a consequence argument, a challenge to brings_about(A,G)", REC + " D066"),
"C007": ("merge", "the value schemes (positive_value, negative_value)", REC + " D067"),
"C009": ("add", "", REC + " D068: a scheme of its own, through the proposal path"),
"C015": ("merge", "the guidelines' rule OC13 (kialo-batch2/guidelines.md §3.1)", REC + " D069"),
"C016": ("merge", "negative_consequences_prudential, as a question", REC + " D069"),
"C020": ("merge", "the guidelines' rule OC12 (kialo-batch2/guidelines.md §3.1)", REC + " D069"),
"C022": ("add", "", REC + " D072: a scheme of its own, through the proposal path"),
"C011": ("merge", "the consequence schemes, as a question or undercutter, once its sketch passes review", REC + " D074"),
}
REASSIGN = {c: "D083" for c in ["C004", "C010", "C013", "C019", "C025", "C026", "C008", "C024", "C023"]}
CANDIDATES["C014"] = ("add", "", REC + " D063: the rule, applied to walton.yml in stage 2")
