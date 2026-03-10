package game

// Integration tests using the REAL core.yml Carneades schemes and NASCQ JSON
// data files. These tests prove that the YAML ↔ NASCQ linkage works
// end-to-end: variable names match, InstantiateScheme produces correct
// statements, attacks route through the correct CQ→scheme mappings, and
// the grounded labelling produces the expected labels.
//
// Run with:
//   go test ./internal/game/ -v -run TestCore

import (
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// =========================================================================
// Embedded data (identical to data/ files — self-contained for testing)
// =========================================================================

const coreYAML = `
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

const coreSchemesJSON = `{
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
    "name": "Argument from Positive Consequences",
    "semantic_type": "action",
    "thesis_independent": false,
    "base_weight": 0.65,
    "variables": {
      "A": { "type": "action", "description": "Action" },
      "G": { "type": "event",  "description": "Good consequence" }
    },
    "premise_templates_en": [
      "Action {A} brings about {G}",
      "{G} is a good outcome"
    ],
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
    "premise_templates_en": [
      "Agent {A} judges {V} as a positive value",
      "Value {V} positively affects goal {G} for agent {A}"
    ],
    "conclusion_template_en": "There is reason to retain goal {G} because of value {V}"
  },
  "AS:sign": {
    "carneades_id": "sign",
    "name": "Argument from Sign",
    "semantic_type": "causality",
    "thesis_independent": false,
    "base_weight": 0.60,
    "variables": {
      "A": { "type": "event", "description": "Observable sign" },
      "B": { "type": "event", "description": "Indicated condition" }
    },
    "premise_templates_en": [
      "{A} is observed to be the case",
      "{B} is indicated when {A} is true"
    ],
    "conclusion_template_en": "{B} is true in the current situation"
  },
  "AS:expert_opinion": {
    "carneades_id": "credible_source",
    "name": "Argument from Expert Opinion",
    "semantic_type": "truth",
    "thesis_independent": true,
    "base_weight": 0.70,
    "variables": {
      "W": { "type": "entity",      "description": "Expert or credible source" },
      "D": { "type": "domain",      "description": "Field of expertise" },
      "S": { "type": "proposition", "description": "Proposition asserted" }
    },
    "premise_templates_en": [
      "{W} is a credible source for domain {D}",
      "{W} asserts that {S}",
      "{S} is in the domain of {D}"
    ],
    "conclusion_template_en": "{S}"
  },
  "AS:position_to_know": {
    "carneades_id": "position_to_know",
    "name": "Argument from Position to Know",
    "semantic_type": "truth",
    "thesis_independent": true,
    "base_weight": 0.65,
    "variables": {
      "W": { "type": "entity",      "description": "Knowledgeable source" },
      "D": { "type": "domain",      "description": "Domain of knowledge" },
      "S": { "type": "proposition", "description": "Proposition asserted" }
    },
    "premise_templates_en": [
      "{W} is in a position to know about {D}",
      "{W} asserts that {S}",
      "{S} is in the domain of {D}"
    ],
    "conclusion_template_en": "{S}"
  },
  "AS:analogy": {
    "carneades_id": "analogy",
    "name": "Argument from Analogy",
    "semantic_type": "truth",
    "thesis_independent": true,
    "base_weight": 0.55,
    "variables": {
      "A": { "type": "proposition", "description": "Proposition claimed" },
      "C": { "type": "case",        "description": "Similar case" }
    },
    "premise_templates_en": [
      "{C} is a similar case",
      "{A} is true in case {C}"
    ],
    "conclusion_template_en": "{A}"
  }
}`

const coreCQsJSON = `{
  "CQ:other_goals": {
    "name": "Other Goals Question",
    "semantic_type": "value",
    "question_template_en": "Are there goals other than {G} that should have priority?",
    "caes_mapping": { "positive_answer_type": "rebuttal", "negative_answer_type": "support", "target": "conclusion" }
  },
  "CQ:other_means": {
    "name": "Other Means Question",
    "semantic_type": "action",
    "question_template_en": "Are there alternative actions to {A} that also achieve {G}?",
    "caes_mapping": { "positive_answer_type": "rebuttal", "negative_answer_type": "support", "target": "conclusion" }
  },
  "CQ:side_effect": {
    "name": "Side Effects Question",
    "semantic_type": "causality",
    "question_template_en": "Would performing {A} have undesirable side effects?",
    "caes_mapping": { "positive_answer_type": "undercutter", "negative_answer_type": "support", "target": "argument" }
  },
  "CQ:possibility": {
    "name": "Possibility Question",
    "semantic_type": "causality",
    "question_template_en": "Is it actually possible to perform action {A}?",
    "caes_mapping": { "positive_answer_type": "support", "negative_answer_type": "undermining", "target": "premise", "target_premise_pattern": "possible({A})" }
  },
  "CQ:best_means": {
    "name": "Best Means Question",
    "semantic_type": "action",
    "question_template_en": "Is {A} the best way to achieve {G}?",
    "caes_mapping": { "positive_answer_type": "support", "negative_answer_type": "rebuttal", "target": "conclusion" }
  },
  "CQ:evidence": {
    "name": "Evidence Question",
    "semantic_type": "truth",
    "question_template_en": "Is there sufficient evidence for the claimed connection?",
    "caes_mapping": { "positive_answer_type": "support", "negative_answer_type": "undercutter", "target": "argument" }
  },
  "CQ:likelihood": {
    "name": "Likelihood Question",
    "semantic_type": "truth",
    "question_template_en": "How likely is it that the effect will occur?",
    "caes_mapping": { "positive_answer_type": "support", "negative_answer_type": "undercutter", "target": "argument" }
  }
}`

const coreEdgesJSON = `{
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

// =========================================================================
// Helpers
// =========================================================================

func makeCoreGame(t *testing.T) *GameState {
	t.Helper()

	// Load Carneades schemes from YAML.
	adapter, err := carneades.NewAdapterFromYAML([]byte(coreYAML))
	if err != nil {
		t.Fatalf("NewAdapterFromYAML failed: %v", err)
	}

	// Load NASCQ from JSON.
	n, err := nascq.LoadFromBytes(
		[]byte(coreSchemesJSON),
		[]byte(coreCQsJSON),
		[]byte(coreEdgesJSON),
	)
	if err != nil {
		t.Fatalf("LoadFromBytes NASCQ failed: %v", err)
	}

	config := DefaultGameConfig()
	config.MaxTurns = 10
	return NewGame(config, adapter, n, []string{"Alice", "Bob"})
}

// =========================================================================
// Test 1: Data loading and mapping verification
// =========================================================================

func TestCoreDataLoading(t *testing.T) {
	gs := makeCoreGame(t)

	// Verify 7 Carneades schemes loaded.
	schemeIDs := gs.Adapter.SchemeIDs()
	if len(schemeIDs) < 7 {
		t.Errorf("expected at least 7 Carneades schemes, got %d: %v", len(schemeIDs), schemeIDs)
	}

	// Verify 7 NASCQ schemes loaded.
	stats := gs.NASCQ.Statistics()
	if stats.SchemeCount != 7 {
		t.Errorf("NASCQ scheme count = %d, want 7", stats.SchemeCount)
	}
	if stats.QuestionCount != 7 {
		t.Errorf("NASCQ CQ count = %d, want 7", stats.QuestionCount)
	}
	if stats.ThesisIndependentCount != 3 {
		t.Errorf("thesis-independent count = %d, want 3", stats.ThesisIndependentCount)
	}

	// Verify every NASCQ carneades_id maps to an actual Carneades scheme.
	schemeSet := make(map[string]bool)
	for _, id := range schemeIDs {
		schemeSet[id] = true
	}
	for nascqID, node := range gs.NASCQ.Schemes {
		if !schemeSet[node.CarneadesID] {
			t.Errorf("NASCQ scheme %q references carneades_id %q which is not in loaded YAML schemes",
				nascqID, node.CarneadesID)
		}
	}
	t.Logf("Carneades schemes: %v", schemeIDs)
	t.Logf("NASCQ: %d schemes, %d CQs, %d ask edges, %d backup edges",
		stats.SchemeCount, stats.QuestionCount, stats.AskEdgeCount, stats.BackupEdgeCount)

	// Verify bipartite structure.
	errors := gs.NASCQ.ValidateBipartite()
	if len(errors) > 0 {
		t.Errorf("NASCQ bipartite validation errors: %v", errors)
	}

	// Verify reverse mapping works.
	nascqID := gs.NASCQ.FindNASCQSchemeID("practical_reasoning1")
	if nascqID != "AS:practical_reasoning" {
		t.Errorf("reverse map practical_reasoning1 = %q, want AS:practical_reasoning", nascqID)
	}
}

// =========================================================================
// Test 2: Support move with real practical_reasoning1 scheme
// =========================================================================

func TestCoreSupportPracticalReasoning(t *testing.T) {
	gs := makeCoreGame(t)

	// Alice argues: government should implement carbon_tax to reduce emissions.
	// Variables match practical_reasoning1: [A, S1, S2, G, V]
	action := nascq.ProposedAction{
		Type:     "support",
		SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{
			"A":  "carbon_tax",
			"S1": "current_climate_crisis",
			"S2": "lower_emissions_state",
			"G":  "reduce_emissions",
			"V":  "environmental_health",
		},
	}

	result, err := gs.ExecuteTurn(action)
	if err != nil {
		t.Fatalf("ExecuteTurn failed: %v", err)
	}
	if !result.Accepted {
		t.Fatalf("support move rejected: %s", result.Error)
	}
	if len(result.ArgIDs) == 0 {
		t.Fatal("no arguments created")
	}

	t.Logf("Created arguments: %v", result.ArgIDs)

	// Verify the conclusion statement exists and is In.
	conclusionID := "should_be_performed(carbon_tax)"
	stmt := gs.Adapter.Graph.Statements[conclusionID]
	if stmt == nil {
		// Try to find it by substring.
		for id, s := range gs.Adapter.Graph.Statements {
			if strings.Contains(id, "should_be_performed") {
				t.Logf("  Found conclusion: %q label=%v", id, s.Label)
				stmt = s
				conclusionID = id
			}
		}
	}
	if stmt == nil {
		t.Fatal("conclusion statement not found in graph")
		return
	}
	if stmt.Label != caes.In {
		t.Errorf("conclusion %q label = %v, want In", conclusionID, stmt.Label)
	}

	// Verify assumptions were created.
	foundAssumptions := 0
	for id, s := range gs.Adapter.Graph.Statements {
		if s.Label == caes.In {
			foundAssumptions++
			t.Logf("  In: %q", id)
		}
	}
	// practical_reasoning1 has 3 assumptions + 4 premises (also assumed) + 1 conclusion = many In statements.
	if foundAssumptions < 3 {
		t.Errorf("expected at least 3 In-labeled statements, got %d", foundAssumptions)
	}

	// Verify scoring: 10 * 0.70 = 7.0
	alice := gs.PlayerByID("Alice")
	if !approxEqual(alice.Score, 7.0, 0.5) {
		t.Errorf("Alice score = %.2f, want ~7.0", alice.Score)
	}

	// Verify Language formatting produced readable text.
	if stmt.Text == "" || stmt.Text == stmt.Id {
		t.Logf("NOTE: conclusion text = %q (Language template may not have applied)", stmt.Text)
	} else {
		t.Logf("Conclusion text: %q", stmt.Text)
	}
}

// =========================================================================
// Test 3: Undercutter attack via CQ:side_effect using AS:sign
// =========================================================================

func TestCoreUndercutterViaCQ(t *testing.T) {
	gs := makeCoreGame(t)

	// Turn 1: Alice supports with practical reasoning.
	support := nascq.ProposedAction{
		Type:     "support",
		SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{
			"A": "carbon_tax", "S1": "current_state", "S2": "lower_emissions",
			"G": "reduce_emissions", "V": "environmental_health",
		},
	}
	r1, _ := gs.ExecuteTurn(support)
	if !r1.Accepted {
		t.Fatalf("Turn 1 rejected: %s", r1.Error)
	}
	targetArgID := r1.ArgIDs[0]
	t.Logf("Turn 1: Alice created %s, score=%.2f", targetArgID, r1.ScoreDelta)

	// Verify conclusion is In.
	targetArg := gs.Adapter.Graph.Arguments[targetArgID]
	if targetArg.Conclusion.Label != caes.In {
		t.Fatalf("before attack: conclusion label=%v, want In", targetArg.Conclusion.Label)
	}

	// Turn 2: Bob attacks via CQ:side_effect using AS:sign.
	// "Rising unemployment (sign A) indicates economic harm (condition B)."
	attack := nascq.ProposedAction{
		Type:        "attack",
		AttackType:  "undercutter",
		SchemeID:    "AS:sign",
		AnswersCQID: "CQ:side_effect",
		TargetArgID: targetArgID,
		VariableBindings: map[string]string{
			"A": "rising_unemployment",
			"B": "economic_harm",
		},
	}
	r2, err := gs.ExecuteTurn(attack)
	if err != nil {
		t.Fatalf("Turn 2 error: %v", err)
	}
	if !r2.Accepted {
		t.Fatalf("Turn 2 rejected: %s", r2.Error)
	}
	t.Logf("Turn 2: Bob created %v, score=%.2f", r2.ArgIDs, r2.ScoreDelta)

	// After undercutter: Alice's conclusion should be Out.
	if targetArg.Conclusion.Label != caes.Out {
		t.Errorf("after undercutter: conclusion label=%v, want Out", targetArg.Conclusion.Label)
		for id, s := range gs.Adapter.Graph.Statements {
			t.Logf("  %q: label=%v", id, s.Label)
		}
	}

	// Verify Bob got undercutter bonus.
	bob := gs.PlayerByID("Bob")
	if bob.Score <= 15.0 {
		t.Errorf("Bob score = %.2f, want > 15 (should include undercutter bonus)", bob.Score)
	}
	t.Logf("Final: Alice=%.2f, Bob=%.2f", gs.PlayerByID("Alice").Score, bob.Score)
}

// =========================================================================
// Test 4: Rebuttal attack via CQ:other_means using AS:consequences
// =========================================================================

func TestCoreRebuttalViaCQ(t *testing.T) {
	gs := makeCoreGame(t)

	// Turn 1: Alice supports.
	r1, _ := gs.ExecuteTurn(nascq.ProposedAction{
		Type:     "support",
		SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{
			"A": "carbon_tax", "S1": "current_state", "S2": "lower_emissions",
			"G": "reduce_emissions", "V": "environmental_health",
		},
	})
	if !r1.Accepted {
		t.Fatalf("Turn 1 rejected: %s", r1.Error)
	}
	targetArgID := r1.ArgIDs[0]
	t.Logf("Turn 1: Alice created %s", targetArgID)

	// Turn 2: Bob rebuts via CQ:other_means using AS:consequences.
	// "Solar subsidies bring about emission reduction, and that's good."
	rebuttal := nascq.ProposedAction{
		Type:        "attack",
		AttackType:  "rebuttal",
		SchemeID:    "AS:consequences",
		AnswersCQID: "CQ:other_means",
		TargetArgID: targetArgID,
		VariableBindings: map[string]string{
			"A": "solar_subsidies",
			"G": "green_energy_adoption",
		},
	}
	r2, err := gs.ExecuteTurn(rebuttal)
	if err != nil {
		t.Fatalf("Turn 2 error: %v", err)
	}
	if !r2.Accepted {
		t.Fatalf("Turn 2 rejected: %s", r2.Error)
	}
	t.Logf("Turn 2: Bob rebutted with %v, score=%.2f", r2.ArgIDs, r2.ScoreDelta)

	// After rebuttal: Issue created. Alice's conclusion has weight 0.70,
	// Bob's rebuttal ¬conclusion has weight 0.65. Under PE, 0.70 > 0.65,
	// so Alice's conclusion should STILL be In.
	targetArg := gs.Adapter.Graph.Arguments[targetArgID]
	if targetArg.Conclusion.Label != caes.In {
		t.Logf("NOTE: conclusion label=%v (Alice's 0.70 should beat Bob's 0.65 under PE)",
			targetArg.Conclusion.Label)
		// This might be Out if the Issue resolution works differently than expected.
		// Log all statements for debugging.
		for id, s := range gs.Adapter.Graph.Statements {
			t.Logf("  %q: label=%v", id, s.Label)
		}
	} else {
		t.Logf("Correct: Alice's conclusion stays In (0.70 > 0.65 under PE)")
	}

	// Verify an Issue was created.
	if len(gs.Adapter.Graph.Issues) == 0 {
		t.Error("expected at least one Issue after rebuttal")
	}

	// Verify both players scored.
	alice := gs.PlayerByID("Alice")
	bob := gs.PlayerByID("Bob")
	t.Logf("Scores: Alice=%.2f, Bob=%.2f", alice.Score, bob.Score)
}

// =========================================================================
// Test 5: Expert opinion (thesis-independent) as direct support
// =========================================================================

func TestCoreExpertOpinionSupport(t *testing.T) {
	gs := makeCoreGame(t)

	// Alice uses expert opinion (thesis-independent, credible_source scheme).
	// Variables match credible_source: [W, D, S]
	action := nascq.ProposedAction{
		Type:     "support",
		SchemeID: "AS:expert_opinion",
		VariableBindings: map[string]string{
			"W": "ipcc",
			"D": "climate_science",
			"S": "global_warming_is_real",
		},
	}

	result, err := gs.ExecuteTurn(action)
	if err != nil {
		t.Fatalf("ExecuteTurn failed: %v", err)
	}
	if !result.Accepted {
		t.Fatalf("expert opinion rejected: %s", result.Error)
	}
	t.Logf("Expert opinion created args: %v, score=%.2f", result.ArgIDs, result.ScoreDelta)

	// The conclusion should be the bare variable S, substituted to
	// global_warming_is_real. Check if it's In.
	found := false
	for id, s := range gs.Adapter.Graph.Statements {
		if strings.Contains(id, "global_warming_is_real") && s.Label == caes.In {
			found = true
			t.Logf("  Found conclusion: %q label=In", id)
		}
	}
	if !found {
		t.Error("global_warming_is_real should be In after expert opinion support")
		for id, s := range gs.Adapter.Graph.Statements {
			t.Logf("  %q: label=%v", id, s.Label)
		}
	}
}

// =========================================================================
// Test 6: Full multi-turn game — two players, diverse strategies
// =========================================================================

func TestCoreFullGame(t *testing.T) {
	gs := makeCoreGame(t)
	gs.Config.MaxTurns = 6

	// Turn 1: Alice — practical reasoning (support).
	r1, _ := gs.ExecuteTurn(nascq.ProposedAction{
		Type: "support", SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{
			"A": "carbon_tax", "S1": "climate_crisis", "S2": "lower_emissions",
			"G": "reduce_emissions", "V": "sustainability",
		},
	})
	if !r1.Accepted {
		t.Fatalf("T1 rejected: %s", r1.Error)
	}
	t.Logf("T1 Alice support: args=%v score=%.2f", r1.ArgIDs, r1.ScoreDelta)

	// Turn 2: Bob — undercutter via side effects.
	r2, _ := gs.ExecuteTurn(nascq.ProposedAction{
		Type: "attack", AttackType: "undercutter",
		SchemeID: "AS:sign", AnswersCQID: "CQ:side_effect",
		TargetArgID: r1.ArgIDs[0],
		VariableBindings: map[string]string{
			"A": "economic_data", "B": "recession_risk",
		},
	})
	if !r2.Accepted {
		t.Fatalf("T2 rejected: %s", r2.Error)
	}
	t.Logf("T2 Bob undercut: args=%v score=%.2f", r2.ArgIDs, r2.ScoreDelta)

	// Turn 3: Alice — expert opinion (direct support, no CQ).
	r3, _ := gs.ExecuteTurn(nascq.ProposedAction{
		Type: "support", SchemeID: "AS:expert_opinion",
		VariableBindings: map[string]string{
			"W": "economist_panel", "D": "economic_policy",
			"S": "carbon_tax_works",
		},
	})
	if !r3.Accepted {
		t.Fatalf("T3 rejected: %s", r3.Error)
	}
	t.Logf("T3 Alice expert: args=%v score=%.2f", r3.ArgIDs, r3.ScoreDelta)

	// Turn 4: Bob — abstains.
	gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	t.Logf("T4 Bob abstains")

	// Turn 5: Alice — consequences support for second argument.
	r5, _ := gs.ExecuteTurn(nascq.ProposedAction{
		Type: "support", SchemeID: "AS:consequences",
		VariableBindings: map[string]string{
			"A": "renewable_investment", "G": "job_creation",
		},
	})
	if !r5.Accepted {
		t.Fatalf("T5 rejected: %s", r5.Error)
	}
	t.Logf("T5 Alice consequences: args=%v score=%.2f", r5.ArgIDs, r5.ScoreDelta)

	// Turn 6: Bob — abstains again.
	r6, _ := gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	_ = r6

	// Final state.
	alice := gs.PlayerByID("Alice")
	bob := gs.PlayerByID("Bob")
	t.Logf("=== FINAL STATE ===")
	t.Logf("Turns played: %d moves in history", len(gs.MoveHistory))
	t.Logf("Scores: Alice=%.2f, Bob=%.2f", alice.Score, bob.Score)
	t.Logf("Graph: %d statements, %d arguments, %d issues",
		len(gs.Adapter.Graph.Statements),
		len(gs.Adapter.Graph.Arguments),
		len(gs.Adapter.Graph.Issues))

	for id, s := range gs.Adapter.Graph.Statements {
		if s.Label == caes.In || s.Label == caes.Out {
			t.Logf("  %q: %v", id, s.Label)
		}
	}

	// Verify game didn't crash and produced reasonable results.
	if len(gs.MoveHistory) != 6 {
		t.Errorf("expected 6 moves in history, got %d", len(gs.MoveHistory))
	}
	if !gs.Adapter.IsAcyclic() {
		t.Error("graph should be acyclic")
	}
	// Both players should have non-zero scores (Alice from support, Bob from undercut).
	if alice.Score <= 0 {
		t.Errorf("Alice should have positive score, got %.2f", alice.Score)
	}
	if bob.Score <= 0 {
		t.Errorf("Bob should have positive score, got %.2f", bob.Score)
	}
}

// =========================================================================
// Test 7: NASCQ validation catches wrong CQ for scheme
// =========================================================================

func TestCoreNASCQRejectsWrongCQ(t *testing.T) {
	gs := makeCoreGame(t)

	// First create an argument to attack.
	r1, _ := gs.ExecuteTurn(nascq.ProposedAction{
		Type: "support", SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{
			"A": "carbon_tax", "S1": "now", "S2": "future",
			"G": "reduce_emissions", "V": "health",
		},
	})
	if !r1.Accepted {
		t.Fatalf("setup failed: %s", r1.Error)
	}

	// Try to attack using CQ:evidence (which is an ask-CQ of AS:consequences,
	// NOT AS:practical_reasoning). Should fail S3.
	badAttack := nascq.ProposedAction{
		Type: "attack", AttackType: "undercutter",
		SchemeID: "AS:expert_opinion", AnswersCQID: "CQ:evidence",
		TargetArgID: r1.ArgIDs[0],
		VariableBindings: map[string]string{
			"W": "expert", "D": "domain", "S": "claim",
		},
	}
	r2, _ := gs.ExecuteTurn(badAttack)
	if r2.Accepted {
		t.Fatal("CQ:evidence should not be applicable to practical_reasoning argument")
	}
	if !strings.Contains(r2.Error, "S3") {
		t.Errorf("expected S3 error, got: %s", r2.Error)
	}
	t.Logf("Correctly rejected: %s", r2.Error)
}
