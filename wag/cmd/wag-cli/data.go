package main

// Embedded game data for the WAG CLI. These are the same 7-scheme core
// data files from Block 6, compiled into the binary so the CLI works
// without external data files.
//
// To customize: modify these constants and rebuild, or (Phase 2) add
// --yaml/--schemes/--cqs/--edges flags for file-based loading.

const embeddedYAML = `
language:
  current_circumstances/1: "%s is the case in the current circumstances."
  would_bring_about/3: "Performing %s in %s would bring about %s."
  would_be_realized/2: "%s would be realized in %s."
  would_promote_value/2: "Achieving the goal %s would promote the value %s."
  should_be_performed/1: "Action %s should be performed."
  legitimate_value/1: "%s is a legitimate value."
  worthy_goal/1: "%s is a worthy goal."
  possible/1: "Action %s is possible."
  bring_about_more_effectively/2: "There exists an action that would bring about %s more effectively than %s."
  realize_more_effectively/2: "There exists an action that would realize the goal %s more effectively than %s."
  promote_more_effectively/2: "There exists an action that would promote the value %s more effectively than %s."
  side_effects/2: "Performing %s in %s would have side-effects which demote some value."
  brings_about/2: "The action %s brings about %s."
  good/1: "%s is good."
  judged_positive_value/2: "Agent %s judges %s as a positive value."
  positive_affects/3: "Value %s positively affects goal %s for agent %s."
  reason_retaining_goal/2: "There is reason to retain goal %s because of value %s."
  finding_true/1: "%s is observed to be the case."
  indicated_true/2: "%s is indicated when %s is true."
  true_in_situation/1: "%s is true in the current situation."
  strong_correlation/2: "There is a strong correlation between %s and %s."
  credible_source/2: "%s is a credible source for the domain %s."
  asserts/2: "%s asserts that %s is true."
  in_domain/2: "%s is in the domain of %s."
  biased/1: "%s is biased."
  dishonest/1: "%s is dishonest."
  other_credible_sources_disagree/1: "Other credible sources disagree with %s."
  position_to_know/2: "%s is in a position to know about things in domain %s."
  similar_case/1: "%s is a similar case."
  in_case/2: "%s is true in case %s."
  relevant_differences/1: "There are relevant differences in case %s."
  more_on_point/2: "There is a case more on point than %s regarding %s."

argument_schemes:
  - id: practical_reasoning1
    meta:
      title: Argument from Value-Based Practical Reasoning
    variables: [A,S1,S2,G,V]
    conclusions:
      - should_be_performed(A)
    premises:
      - current_circumstances(S1)
      - would_bring_about(A,S1,S2)
      - would_be_realized(G,S2)
      - would_promote_value(G,V)
    assumptions:
      - legitimate_value(V)
      - worthy_goal(G)
      - possible(A)
    exceptions:
      - bring_about_more_effectively(S2,A)
      - realize_more_effectively(G,A)
      - promote_more_effectively(V,A)
      - side_effects(A,S1)

  - id: positive_consequences
    meta:
      title: Argument from Positive Consequences
    variables: [A,G]
    conclusions:
      - should_be_performed(A)
    premises:
      - brings_about(A,G)
      - good(G)

  - id: positive_value
    meta:
      title: Argument from Values
    variables: [V,A,G]
    conclusions:
      - reason_retaining_goal(V,G)
    premises:
      - judged_positive_value(V,A)
      - positive_affects(V,G,A)

  - id: sign
    meta:
      title: Argument from Sign
    variables: [A,B]
    conclusions:
      - true_in_situation(B)
    premises:
      - finding_true(A)
      - indicated_true(B,A)
    assumptions:
      - strong_correlation(A,B)

  - id: credible_source
    meta:
      title: Argument from Credible Source
    variables: [W,D,S]
    conclusions: [S]
    premises:
      - credible_source(W,D)
      - asserts(W,S)
      - in_domain(S,D)
    exceptions:
      - biased(W)
      - dishonest(W)
      - other_credible_sources_disagree(S)

  - id: position_to_know
    meta:
      title: Argument from Position to Know
    variables: [W,D,S]
    conclusions: [S]
    premises:
      - position_to_know(W,D)
      - asserts(W,S)
      - in_domain(S,D)
    exceptions:
      - dishonest(W)

  - id: analogy
    meta:
      title: Argument from Analogy
    variables: [A,C]
    conclusions: [A]
    premises:
      - similar_case(C)
      - in_case(A,C)
    exceptions:
      - relevant_differences(C)
      - more_on_point(A,C)
`

const embeddedSchemesJSON = `{
  "AS:practical_reasoning": {
    "carneades_id": "practical_reasoning1",
    "name": "Practical Reasoning",
    "semantic_type": "action",
    "thesis_independent": false,
    "base_weight": 0.70,
    "variables": {
      "A":  { "type": "action", "description": "Action to be performed" },
      "S1": { "type": "state",  "description": "Current circumstances" },
      "S2": { "type": "state",  "description": "Future state after action" },
      "G":  { "type": "goal",   "description": "Goal to be realized" },
      "V":  { "type": "value",  "description": "Value promoted by the goal" }
    },
    "premise_templates_en": [
      "In the current circumstances {S1}",
      "Performing {A} in {S1} would bring about {S2}",
      "Goal {G} would be realized in {S2}",
      "Goal {G} would promote value {V}"
    ],
    "conclusion_template_en": "Action {A} should be performed"
  },
  "AS:consequences": {
    "carneades_id": "positive_consequences",
    "name": "Positive Consequences",
    "semantic_type": "action",
    "thesis_independent": false,
    "base_weight": 0.65,
    "variables": {
      "A": { "type": "action", "description": "Action" },
      "G": { "type": "event",  "description": "Good consequence" }
    },
    "premise_templates_en": ["Action {A} brings about {G}", "{G} is a good outcome"],
    "conclusion_template_en": "Action {A} should be performed"
  },
  "AS:values": {
    "carneades_id": "positive_value",
    "name": "Argument from Values",
    "semantic_type": "value",
    "thesis_independent": false,
    "base_weight": 0.60,
    "variables": {
      "V": { "type": "value",  "description": "Value being invoked" },
      "A": { "type": "entity", "description": "Agent or judger" },
      "G": { "type": "goal",   "description": "Goal affected" }
    },
    "premise_templates_en": ["Agent {A} judges {V} as a positive value", "Value {V} positively affects goal {G} for agent {A}"],
    "conclusion_template_en": "There is reason to retain goal {G} because of value {V}"
  },
  "AS:sign": {
    "carneades_id": "sign",
    "name": "Argument from Sign",
    "semantic_type": "causality",
    "thesis_independent": false,
    "base_weight": 0.60,
    "variables": {
      "A": { "type": "event", "description": "Observable sign or indicator" },
      "B": { "type": "event", "description": "Indicated condition" }
    },
    "premise_templates_en": ["{A} is observed to be the case", "{B} is indicated when {A} is true"],
    "conclusion_template_en": "{B} is true in the current situation"
  },
  "AS:expert_opinion": {
    "carneades_id": "credible_source",
    "name": "Expert Opinion",
    "semantic_type": "truth",
    "thesis_independent": true,
    "base_weight": 0.70,
    "variables": {
      "W": { "type": "entity",      "description": "Expert or credible source" },
      "D": { "type": "domain",      "description": "Field of expertise" },
      "S": { "type": "proposition", "description": "Proposition asserted" }
    },
    "premise_templates_en": ["{W} is a credible source for domain {D}", "{W} asserts that {S}", "{S} is in the domain of {D}"],
    "conclusion_template_en": "{S}"
  },
  "AS:position_to_know": {
    "carneades_id": "position_to_know",
    "name": "Position to Know",
    "semantic_type": "truth",
    "thesis_independent": true,
    "base_weight": 0.65,
    "variables": {
      "W": { "type": "entity",      "description": "Knowledgeable source" },
      "D": { "type": "domain",      "description": "Domain of knowledge" },
      "S": { "type": "proposition", "description": "Proposition asserted" }
    },
    "premise_templates_en": ["{W} is in a position to know about {D}", "{W} asserts that {S}", "{S} is in the domain of {D}"],
    "conclusion_template_en": "{S}"
  },
  "AS:analogy": {
    "carneades_id": "analogy",
    "name": "Argument from Analogy",
    "semantic_type": "truth",
    "thesis_independent": true,
    "base_weight": 0.55,
    "variables": {
      "A": { "type": "proposition", "description": "Proposition claimed to hold" },
      "C": { "type": "case",        "description": "Similar case" }
    },
    "premise_templates_en": ["{C} is a similar case", "{A} is true in case {C}"],
    "conclusion_template_en": "{A}"
  }
}`

const embeddedCQsJSON = `{
  "CQ:other_goals":  { "name": "Other Goals Question",  "semantic_type": "value",     "question_template_en": "Are there goals other than {G} that should have priority?",          "caes_mapping": { "positive_answer_type": "rebuttal",    "negative_answer_type": "support",     "target": "conclusion" } },
  "CQ:other_means":  { "name": "Other Means Question",  "semantic_type": "action",    "question_template_en": "Are there alternative actions to {A} that also achieve {G}?",         "caes_mapping": { "positive_answer_type": "rebuttal",    "negative_answer_type": "support",     "target": "conclusion" } },
  "CQ:side_effect":  { "name": "Side Effects Question", "semantic_type": "causality", "question_template_en": "Would performing {A} have undesirable side effects?",                 "caes_mapping": { "positive_answer_type": "undercutter", "negative_answer_type": "support",     "target": "argument" } },
  "CQ:possibility":  { "name": "Possibility Question",  "semantic_type": "causality", "question_template_en": "Is it actually possible to perform action {A}?",                      "caes_mapping": { "positive_answer_type": "support",     "negative_answer_type": "undermining", "target": "premise" } },
  "CQ:best_means":   { "name": "Best Means Question",   "semantic_type": "action",    "question_template_en": "Is {A} the best way to achieve {G}?",                                 "caes_mapping": { "positive_answer_type": "support",     "negative_answer_type": "rebuttal",    "target": "conclusion" } },
  "CQ:evidence":     { "name": "Evidence Question",     "semantic_type": "truth",     "question_template_en": "Is there sufficient evidence for the claimed connection?",             "caes_mapping": { "positive_answer_type": "support",     "negative_answer_type": "undercutter", "target": "argument" } },
  "CQ:likelihood":   { "name": "Likelihood Question",   "semantic_type": "truth",     "question_template_en": "How likely is it that the claimed effect will actually occur?",        "caes_mapping": { "positive_answer_type": "support",     "negative_answer_type": "undercutter", "target": "argument" } }
}`

const embeddedEdgesJSON = `{
  "ask_edges": [
    { "cq": "CQ:other_goals",  "scheme": "AS:practical_reasoning" },
    { "cq": "CQ:other_means",  "scheme": "AS:practical_reasoning" },
    { "cq": "CQ:side_effect",  "scheme": "AS:practical_reasoning" },
    { "cq": "CQ:possibility",  "scheme": "AS:practical_reasoning" },
    { "cq": "CQ:best_means",   "scheme": "AS:practical_reasoning" },
    { "cq": "CQ:evidence",     "scheme": "AS:consequences" },
    { "cq": "CQ:side_effect",  "scheme": "AS:consequences" },
    { "cq": "CQ:likelihood",   "scheme": "AS:consequences" },
    { "cq": "CQ:likelihood",   "scheme": "AS:sign" }
  ],
  "backup_edges": [
    { "scheme": "AS:values",        "cq": "CQ:other_goals" },
    { "scheme": "AS:consequences",  "cq": "CQ:other_means" },
    { "scheme": "AS:sign",          "cq": "CQ:side_effect" },
    { "scheme": "AS:sign",          "cq": "CQ:possibility" }
  ]
}`
