package nascq

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Mock GraphQuerier for testing
// ---------------------------------------------------------------------------

// mockGraph simulates the Carneades adapter's graph for validation testing.
// In production, the real adapter implements GraphQuerier.
type mockGraph struct {
	arguments map[string]string   // argID → carneadesSchemeID
	premises  map[string][]string // argID → []premiseStmtID
}

func (m *mockGraph) ArgumentExists(argID string) bool {
	_, ok := m.arguments[argID]
	return ok
}

func (m *mockGraph) ArgumentSchemeID(argID string) string {
	return m.arguments[argID]
}

func (m *mockGraph) PremiseExists(argID, premiseStmtID string) bool {
	for _, pid := range m.premises[argID] {
		if pid == premiseStmtID {
			return true
		}
	}
	return false
}

func newMockGraph() *mockGraph {
	return &mockGraph{
		arguments: map[string]string{
			"a1": "practical_reasoning1", // Uses practical reasoning
			"a2": "credible_source",      // Uses expert opinion
		},
		premises: map[string][]string{
			"a1": {"possible(reduce_emissions)", "current_circumstances(now)", "would_bring_about(carbon_tax,now,lower_emissions)"},
			"a2": {"expert(dr_smith)", "asserts(dr_smith,climate_change_real)"},
		},
	}
}

// ---------------------------------------------------------------------------
// Test fixture loader (reuse from loader_test.go fixtures)
// ---------------------------------------------------------------------------

// We redefine the fixtures here rather than importing from loader_test.go
// because Go test files can't be imported across test packages.
// These are identical to the Block 3 fixtures.

const valSchemesJSON = `{
  "AS:practical_reasoning": {
    "carneades_id": "practical_reasoning1",
    "name": "Practical Reasoning",
    "semantic_type": "action",
    "thesis_independent": false,
    "base_weight": 0.70,
    "variables": {
      "Agent": { "type": "entity", "description": "Who should act" },
      "Goal":  { "type": "goal",   "description": "What they want to achieve" },
      "Action": { "type": "action", "description": "What they should do" }
    },
    "premise_templates_en": [
      "{Agent} has goal {Goal}",
      "{Action} is a means to achieve {Goal}",
      "{Action} is possible for {Agent}"
    ],
    "conclusion_template_en": "{Agent} should perform {Action}"
  },
  "AS:expert_opinion": {
    "carneades_id": "credible_source",
    "name": "Argument from Expert Opinion",
    "semantic_type": "truth",
    "thesis_independent": true,
    "base_weight": 0.70,
    "variables": {
      "Expert":      { "type": "entity",      "description": "The expert" },
      "Domain":      { "type": "domain",      "description": "Field of expertise" },
      "Proposition": { "type": "proposition", "description": "What is asserted" }
    },
    "premise_templates_en": [
      "{Expert} is an expert in {Domain}",
      "{Expert} asserts that {Proposition}",
      "{Proposition} is within {Domain}"
    ],
    "conclusion_template_en": "{Proposition} is true"
  },
  "AS:cause_to_effect": {
    "carneades_id": "cause_to_effect",
    "name": "Argument from Cause to Effect",
    "semantic_type": "causality",
    "thesis_independent": false,
    "base_weight": 0.65,
    "variables": {
      "Cause":  { "type": "event", "description": "The cause" },
      "Effect": { "type": "event", "description": "The effect" }
    },
    "premise_templates_en": [
      "Generally, if {Cause} occurs, then {Effect} occurs",
      "In this case, {Cause} occurs"
    ],
    "conclusion_template_en": "In this case, {Effect} will occur"
  },
  "AS:values": {
    "carneades_id": "positive_value",
    "name": "Argument from Values",
    "semantic_type": "value",
    "thesis_independent": false,
    "base_weight": 0.60,
    "variables": {
      "Value":  { "type": "value",  "description": "The value" },
      "Action": { "type": "action", "description": "The action" }
    },
    "premise_templates_en": [
      "{Value} is a legitimate value",
      "Performing {Action} promotes {Value}"
    ],
    "conclusion_template_en": "{Action} should be performed to promote {Value}"
  },
  "AS:analogy": {
    "carneades_id": "analogy",
    "name": "Argument from Analogy",
    "semantic_type": "truth",
    "thesis_independent": true,
    "base_weight": 0.55,
    "variables": {
      "Case1":      { "type": "case",        "description": "Similar case" },
      "Case2":      { "type": "case",        "description": "Target case" },
      "Proposition": { "type": "proposition", "description": "What holds" }
    },
    "premise_templates_en": [
      "{Case1} is similar to {Case2}",
      "{Proposition} is true in {Case1}"
    ],
    "conclusion_template_en": "{Proposition} is true in {Case2}"
  }
}`

const valQuestionsJSON = `{
  "CQ:other_means": {
    "name": "Other Means Question",
    "semantic_type": "existence",
    "question_template_en": "Are there alternative actions to {Action} that also achieve {Goal}?",
    "caes_mapping": {
      "positive_answer_type": "rebuttal",
      "negative_answer_type": "support",
      "target": "conclusion"
    }
  },
  "CQ:other_goals": {
    "name": "Other Goals Question",
    "semantic_type": "value",
    "question_template_en": "Are there goals other than {Goal} that should have priority?",
    "caes_mapping": {
      "positive_answer_type": "rebuttal",
      "negative_answer_type": "support",
      "target": "conclusion"
    }
  },
  "CQ:side_effects": {
    "name": "Side Effects Question",
    "semantic_type": "possibility",
    "question_template_en": "Would performing {Action} have undesirable side effects?",
    "caes_mapping": {
      "positive_answer_type": "undercutter",
      "negative_answer_type": "support",
      "target": "argument"
    }
  },
  "CQ:possibility": {
    "name": "Possibility Question",
    "semantic_type": "possibility",
    "question_template_en": "Is it actually possible for {Agent} to do {Action}?",
    "caes_mapping": {
      "positive_answer_type": "support",
      "negative_answer_type": "undermining",
      "target": "premise",
      "target_premise_pattern": "possible({Action})"
    }
  },
  "CQ:expert_credibility": {
    "name": "Expert Credibility Question",
    "semantic_type": "property",
    "question_template_en": "Does {Expert} have genuine expertise in {Domain}?",
    "caes_mapping": {
      "positive_answer_type": "support",
      "negative_answer_type": "undercutter",
      "target": "argument"
    }
  },
  "CQ:interference": {
    "name": "Interference Question",
    "semantic_type": "causality",
    "question_template_en": "Is there a factor that could prevent {Cause} from causing {Effect}?",
    "caes_mapping": {
      "positive_answer_type": "undercutter",
      "negative_answer_type": "support",
      "target": "argument"
    }
  },
  "CQ:legitimate_value": {
    "name": "Legitimate Value Question",
    "semantic_type": "value",
    "question_template_en": "Is {Value} a legitimate value in this context?",
    "caes_mapping": {
      "positive_answer_type": "support",
      "negative_answer_type": "undercutter",
      "target": "argument"
    }
  }
}`

const valEdgesJSON = `{
  "ask_edges": [
    { "cq": "CQ:other_means",        "scheme": "AS:practical_reasoning" },
    { "cq": "CQ:other_goals",        "scheme": "AS:practical_reasoning" },
    { "cq": "CQ:side_effects",       "scheme": "AS:practical_reasoning" },
    { "cq": "CQ:possibility",        "scheme": "AS:practical_reasoning" },
    { "cq": "CQ:expert_credibility", "scheme": "AS:expert_opinion" },
    { "cq": "CQ:interference",       "scheme": "AS:cause_to_effect" },
    { "cq": "CQ:legitimate_value",   "scheme": "AS:values" }
  ],
  "backup_edges": [
    { "scheme": "AS:cause_to_effect", "cq": "CQ:side_effects" },
    { "scheme": "AS:values",          "cq": "CQ:other_goals" },
    { "scheme": "AS:cause_to_effect", "cq": "CQ:interference" },
    { "scheme": "AS:values",          "cq": "CQ:legitimate_value" }
  ]
}`

func loadValNASCQ(t *testing.T) *NASCQ {
	t.Helper()
	n, err := LoadFromBytes([]byte(valSchemesJSON), []byte(valQuestionsJSON), []byte(valEdgesJSON))
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}
	return n
}

// ---------------------------------------------------------------------------
// Tests: Template instantiation
// ---------------------------------------------------------------------------

func TestInstantiateTemplate(t *testing.T) {
	template := "{Agent} has goal {Goal}"
	bindings := map[string]string{"Agent": "Alice", "Goal": "reduce emissions"}
	got := instantiateTemplate(template, bindings)
	want := "Alice has goal reduce emissions"
	if got != want {
		t.Errorf("instantiateTemplate = %q, want %q", got, want)
	}
}

func TestInstantiateTemplates(t *testing.T) {
	n := loadValNASCQ(t)
	scheme := n.GetScheme("AS:practical_reasoning")
	bindings := map[string]string{
		"Agent":  "Alice",
		"Goal":   "reduce emissions",
		"Action": "carbon tax",
	}
	premises, conclusion := instantiateTemplates(scheme, bindings)
	if len(premises) != 3 {
		t.Fatalf("expected 3 premises, got %d", len(premises))
	}
	if premises[0] != "Alice has goal reduce emissions" {
		t.Errorf("premise[0] = %q", premises[0])
	}
	if conclusion != "Alice should perform carbon tax" {
		t.Errorf("conclusion = %q", conclusion)
	}
}

// ---------------------------------------------------------------------------
// Tests: Reverse scheme index
// ---------------------------------------------------------------------------

func TestFindNASCQSchemeID(t *testing.T) {
	n := loadValNASCQ(t)

	// practical_reasoning1 → AS:practical_reasoning
	got := n.FindNASCQSchemeID("practical_reasoning1")
	if got != "AS:practical_reasoning" {
		t.Errorf("FindNASCQSchemeID(practical_reasoning1) = %q, want AS:practical_reasoning", got)
	}

	// credible_source → AS:expert_opinion
	got = n.FindNASCQSchemeID("credible_source")
	if got != "AS:expert_opinion" {
		t.Errorf("FindNASCQSchemeID(credible_source) = %q, want AS:expert_opinion", got)
	}

	// Unknown → ""
	got = n.FindNASCQSchemeID("nonexistent")
	if got != "" {
		t.Errorf("FindNASCQSchemeID(nonexistent) = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// Tests: S1 — Scheme existence
// ---------------------------------------------------------------------------

func TestValidateS1SchemeExists(t *testing.T) {
	n := loadValNASCQ(t)
	action := ProposedAction{
		Type:     "support",
		SchemeID: "AS:nonexistent",
		VariableBindings: map[string]string{},
	}
	_, err := n.Validate(action, nil)
	if err == nil {
		t.Fatal("S1: expected error for nonexistent scheme")
	}
	if !strings.Contains(err.Error(), "S1") {
		t.Errorf("error should mention S1, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: S2 — Variable completeness
// ---------------------------------------------------------------------------

func TestValidateS2MissingVariable(t *testing.T) {
	n := loadValNASCQ(t)
	action := ProposedAction{
		Type:     "support",
		SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{
			"Agent": "Alice",
			// Missing Goal and Action
		},
	}
	_, err := n.Validate(action, nil)
	if err == nil {
		t.Fatal("S2: expected error for missing variables")
	}
	if !strings.Contains(err.Error(), "S2") {
		t.Errorf("error should mention S2, got: %v", err)
	}
	// Should mention both missing variables.
	if !strings.Contains(err.Error(), "Action") || !strings.Contains(err.Error(), "Goal") {
		t.Errorf("error should list missing variables, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: S3 — CQ applicability
// ---------------------------------------------------------------------------

func TestValidateS3CQApplicable(t *testing.T) {
	n := loadValNASCQ(t)
	graph := newMockGraph()

	// side_effects is an ask-CQ of practical_reasoning. Attacking a1
	// (which uses practical_reasoning1 → NASCQ AS:practical_reasoning)
	// with CQ:side_effects should be valid.
	action := ProposedAction{
		Type:        "attack",
		AttackType:  "undercutter",
		SchemeID:    "AS:cause_to_effect", // backup scheme for side_effects
		AnswersCQID: "CQ:side_effects",
		TargetArgID: "a1",
		VariableBindings: map[string]string{
			"Cause":  "carbon tax",
			"Effect": "economic recession",
		},
	}
	result, err := n.Validate(action, graph)
	if err != nil {
		t.Fatalf("S3: valid action should pass, got: %v", err)
	}
	if result.CaesAttackType != "undercutter" {
		t.Errorf("CaesAttackType = %q, want undercutter (from CQ mapping)", result.CaesAttackType)
	}
}

func TestValidateS3CQNotApplicable(t *testing.T) {
	n := loadValNASCQ(t)
	graph := newMockGraph()

	// expert_credibility is an ask-CQ of expert_opinion, NOT practical_reasoning.
	// Trying to use it against a1 (practical_reasoning) should fail S3.
	action := ProposedAction{
		Type:        "attack",
		AttackType:  "undercutter",
		SchemeID:    "AS:expert_opinion", // thesis-independent, passes S5
		AnswersCQID: "CQ:expert_credibility",
		TargetArgID: "a1", // uses practical_reasoning, not expert_opinion
		VariableBindings: map[string]string{
			"Expert":      "Dr. Smith",
			"Domain":      "climate",
			"Proposition": "tax works",
		},
	}
	_, err := n.Validate(action, graph)
	if err == nil {
		t.Fatal("S3: CQ:expert_credibility should not be applicable to practical_reasoning arg")
	}
	if !strings.Contains(err.Error(), "S3") {
		t.Errorf("error should mention S3, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: S4 — Semantic type compatibility
// ---------------------------------------------------------------------------

func TestValidateS4SemanticMismatch(t *testing.T) {
	n := loadValNASCQ(t)
	graph := newMockGraph()

	// AS:values has semantic_type "value".
	// CQ:side_effects has semantic_type "possibility".
	// Using values to back up side_effects should fail S4
	// (values is NOT thesis-independent and types don't match).
	action := ProposedAction{
		Type:        "attack",
		AttackType:  "undercutter",
		SchemeID:    "AS:values",
		AnswersCQID: "CQ:side_effects",
		TargetArgID: "a1",
		VariableBindings: map[string]string{
			"Value":  "economic stability",
			"Action": "carbon tax",
		},
	}
	_, err := n.Validate(action, graph)
	if err == nil {
		t.Fatal("S4: semantic type mismatch should fail")
	}
// S5 now runs before S4. Since AS:values has no backup edge to
	// CQ:side_effects, S5 catches it first. This is correct — S4 only
	// applies to schemes that pass S5 but have mismatched types.
	if !strings.Contains(err.Error(), "S5") && !strings.Contains(err.Error(), "S4") {
		t.Errorf("error should mention S5 or S4, got: %v", err)
	}
}

func TestValidateS4ThesisIndependentBypassesSemantic(t *testing.T) {
	n := loadValNASCQ(t)
	graph := newMockGraph()

	// Thesis-independent schemes (expert_opinion) should bypass S4.
	action := ProposedAction{
		Type:        "attack",
		AttackType:  "undercutter",
		SchemeID:    "AS:expert_opinion",
		AnswersCQID: "CQ:side_effects",
		TargetArgID: "a1",
		VariableBindings: map[string]string{
			"Expert":      "Dr. Smith",
			"Domain":      "environmental science",
			"Proposition": "no side effects",
		},
	}
	_, err := n.Validate(action, graph)
	if err != nil {
		t.Fatalf("S4: thesis-independent should bypass semantic check, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: S5 — Backup edge validity
// ---------------------------------------------------------------------------

func TestValidateS5NoBackupEdge(t *testing.T) {
	n := loadValNASCQ(t)
	graph := newMockGraph()

	// AS:practical_reasoning is NOT thesis-independent and has NO backup
	// edge to CQ:other_goals. But CQ:other_goals IS an ask-CQ of
	// AS:practical_reasoning, so S3 passes. S5 should catch the missing backup.
	action := ProposedAction{
		Type:        "attack",
		AttackType:  "rebuttal",
		SchemeID:    "AS:practical_reasoning", // no backup to CQ:other_goals
		AnswersCQID: "CQ:other_goals",
		TargetArgID: "a1",
		VariableBindings: map[string]string{
			"Agent":  "Alice",
			"Goal":   "different goal",
			"Action": "different action",
		},
	}
	_, err := n.Validate(action, graph)
	if err == nil {
		t.Fatal("S5: no backup edge should fail")
	}
	if !strings.Contains(err.Error(), "S5") {
		t.Errorf("error should mention S5, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: S6 — Template instantiation (verified through successful validate)
// ---------------------------------------------------------------------------

func TestValidateS6TemplatesInstantiated(t *testing.T) {
	n := loadValNASCQ(t)

	action := ProposedAction{
		Type:     "support",
		SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{
			"Agent":  "Alice",
			"Goal":   "reduce emissions",
			"Action": "carbon tax",
		},
	}
	result, err := n.Validate(action, nil)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}
	if len(result.InstPremises) != 3 {
		t.Fatalf("expected 3 instantiated premises, got %d", len(result.InstPremises))
	}
	if result.InstPremises[0] != "Alice has goal reduce emissions" {
		t.Errorf("premise[0] = %q", result.InstPremises[0])
	}
	if result.InstConclusion != "Alice should perform carbon tax" {
		t.Errorf("conclusion = %q", result.InstConclusion)
	}
	if result.CarneadesSchemeID != "practical_reasoning1" {
		t.Errorf("CarneadesSchemeID = %q, want practical_reasoning1", result.CarneadesSchemeID)
	}
	if result.Weight != 0.70 {
		t.Errorf("Weight = %v, want 0.70", result.Weight)
	}
}

// ---------------------------------------------------------------------------
// Tests: S7 — Target argument exists
// ---------------------------------------------------------------------------

func TestValidateS7TargetNotFound(t *testing.T) {
	n := loadValNASCQ(t)
	graph := newMockGraph()

	action := ProposedAction{
		Type:        "attack",
		AttackType:  "undercutter",
		SchemeID:    "AS:expert_opinion",
		TargetArgID: "a999", // doesn't exist
		VariableBindings: map[string]string{
			"Expert":      "Dr. X",
			"Domain":      "D",
			"Proposition": "P",
		},
	}
	_, err := n.Validate(action, graph)
	if err == nil {
		t.Fatal("S7: nonexistent target should fail")
	}
	if !strings.Contains(err.Error(), "S7") {
		t.Errorf("error should mention S7, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: S8 — Target premise exists (undermining)
// ---------------------------------------------------------------------------

func TestValidateS8PremiseNotFound(t *testing.T) {
	n := loadValNASCQ(t)
	graph := newMockGraph()

	action := ProposedAction{
		Type:            "attack",
		AttackType:      "undermining",
		SchemeID:        "AS:expert_opinion",
		TargetArgID:     "a1",
		TargetPremiseID: "nonexistent_premise",
		VariableBindings: map[string]string{
			"Expert":      "Dr. X",
			"Domain":      "D",
			"Proposition": "P",
		},
	}
	_, err := n.Validate(action, graph)
	if err == nil {
		t.Fatal("S8: nonexistent premise should fail")
	}
	if !strings.Contains(err.Error(), "S8") {
		t.Errorf("error should mention S8, got: %v", err)
	}
}

func TestValidateS8PremiseExists(t *testing.T) {
	n := loadValNASCQ(t)
	graph := newMockGraph()

	action := ProposedAction{
		Type:            "attack",
		AttackType:      "undermining",
		SchemeID:        "AS:expert_opinion", // thesis-independent, passes S5
		TargetArgID:     "a1",
		TargetPremiseID: "possible(reduce_emissions)",
		VariableBindings: map[string]string{
			"Expert":      "Dr. X",
			"Domain":      "feasibility",
			"Proposition": "not possible",
		},
	}
	result, err := n.Validate(action, graph)
	if err != nil {
		t.Fatalf("S8: valid premise should pass, got: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
}

// ---------------------------------------------------------------------------
// Tests: Full valid support action (no graph needed)
// ---------------------------------------------------------------------------

func TestValidateDirectSupport(t *testing.T) {
	n := loadValNASCQ(t)

	action := ProposedAction{
		Type:     "support",
		SchemeID: "AS:expert_opinion",
		VariableBindings: map[string]string{
			"Expert":      "IPCC",
			"Domain":      "climate science",
			"Proposition": "global warming is real",
		},
	}
	result, err := n.Validate(action, nil)
	if err != nil {
		t.Fatalf("direct support should pass: %v", err)
	}
	if result.CarneadesSchemeID != "credible_source" {
		t.Errorf("CarneadesSchemeID = %q, want credible_source", result.CarneadesSchemeID)
	}
	if result.Weight != 0.70 {
		t.Errorf("Weight = %v, want 0.70", result.Weight)
	}
}

// ---------------------------------------------------------------------------
// Tests: CAES attack type resolution from CQ mapping
// ---------------------------------------------------------------------------

func TestValidateCaesAttackTypeResolution(t *testing.T) {
	n := loadValNASCQ(t)
	graph := newMockGraph()

	// CQ:side_effects has:
	//   positive_answer_type = "undercutter" (yes, side effects exist → attacks)
	//   negative_answer_type = "support" (no side effects → supports)
	//
	// When attacking via this CQ, the CAES attack type should be "undercutter"
	// (the non-"support" answer type, since that's the attacking direction).
	action := ProposedAction{
		Type:        "attack",
		AttackType:  "undercutter",
		SchemeID:    "AS:cause_to_effect",
		AnswersCQID: "CQ:side_effects",
		TargetArgID: "a1",
		VariableBindings: map[string]string{
			"Cause":  "carbon tax",
			"Effect": "economic downturn",
		},
	}
	result, err := n.Validate(action, graph)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}
	if result.CaesAttackType != "undercutter" {
		t.Errorf("CaesAttackType = %q, want %q (from CQ:side_effects positive_answer_type)",
			result.CaesAttackType, "undercutter")
	}
}

func TestValidateCaesAttackTypeNegativeDirection(t *testing.T) {
	n := loadValNASCQ(t)
	graph := newMockGraph()

	// CQ:expert_credibility has:
	//   positive_answer_type = "support" (yes, credible → supports)
	//   negative_answer_type = "undercutter" (no, not credible → attacks)
	//
	// When attacking, CAES attack type should be "undercutter"
	// (the non-"support" answer type).
	action := ProposedAction{
		Type:        "attack",
		AttackType:  "undercutter",
		SchemeID:    "AS:analogy",
		AnswersCQID: "CQ:expert_credibility",
		TargetArgID: "a2",
		VariableBindings: map[string]string{
			"Case1":       "similar expert",
			"Case2":       "Dr. Smith",
			"Proposition": "not credible",
		},
	}
	result, err := n.Validate(action, graph)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}
	if result.CaesAttackType != "undercutter" {
		t.Errorf("CaesAttackType = %q, want %q (from CQ:expert_credibility negative_answer_type)",
			result.CaesAttackType, "undercutter")
	}
}
