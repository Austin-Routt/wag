package nascq

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Test fixtures (plan format)
// ---------------------------------------------------------------------------

// Minimal NASCQ dataset: 5 schemes, 7 CQs, edges.
// Subset of the plan's Appendix A MVP (10 schemes, 15 CQs).
// Uses the plan's JSON format (map-keyed schemes/CQs, separate edge arrays).

const testSchemesJSON = `{
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

const testQuestionsJSON = `{
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

const testEdgesJSON = `{
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

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func loadTestNASCQ(t *testing.T) *NASCQ {
	t.Helper()
	n, err := LoadFromBytes(
		[]byte(testSchemesJSON),
		[]byte(testQuestionsJSON),
		[]byte(testEdgesJSON),
	)
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}
	return n
}

// ---------------------------------------------------------------------------
// Tests: Loading
// ---------------------------------------------------------------------------

func TestLoadFromBytes(t *testing.T) {
	n := loadTestNASCQ(t)

	if len(n.Schemes) != 5 {
		t.Errorf("scheme count = %d, want 5", len(n.Schemes))
	}
	if len(n.Questions) != 7 {
		t.Errorf("question count = %d, want 7", len(n.Questions))
	}
	if len(n.AskEdges) != 7 {
		t.Errorf("ask edge count = %d, want 7", len(n.AskEdges))
	}
	if len(n.BackupEdges) != 4 {
		t.Errorf("backup edge count = %d, want 4", len(n.BackupEdges))
	}
}

func TestSchemeFieldsLoaded(t *testing.T) {
	n := loadTestNASCQ(t)

	pr := n.GetScheme("AS:practical_reasoning")
	if pr == nil {
		t.Fatal("AS:practical_reasoning not found")
	}
	if pr.CarneadesID != "practical_reasoning1" {
		t.Errorf("CarneadesID = %q, want %q", pr.CarneadesID, "practical_reasoning1")
	}
	if pr.Name != "Practical Reasoning" {
		t.Errorf("Name = %q, want %q", pr.Name, "Practical Reasoning")
	}
	if pr.SemanticType != TypeAction {
		t.Errorf("SemanticType = %q, want %q", pr.SemanticType, TypeAction)
	}
	if pr.ThesisIndependent {
		t.Error("practical_reasoning should NOT be thesis-independent")
	}
	if pr.BaseWeight != 0.70 {
		t.Errorf("BaseWeight = %v, want 0.70", pr.BaseWeight)
	}
	if len(pr.Variables) != 3 {
		t.Errorf("variable count = %d, want 3", len(pr.Variables))
	}
	if len(pr.PremiseTemplatesEN) != 3 {
		t.Errorf("premise template count = %d, want 3", len(pr.PremiseTemplatesEN))
	}
	if pr.ConclusionTemplateEN == "" {
		t.Error("conclusion template is empty")
	}
}

func TestCQFieldsLoaded(t *testing.T) {
	n := loadTestNASCQ(t)

	cq := n.GetCQ("CQ:side_effects")
	if cq == nil {
		t.Fatal("CQ:side_effects not found")
	}
	if cq.Name != "Side Effects Question" {
		t.Errorf("Name = %q, want %q", cq.Name, "Side Effects Question")
	}
	if cq.SemanticType != TypePossibility {
		t.Errorf("SemanticType = %q, want %q", cq.SemanticType, TypePossibility)
	}
	if cq.CaesMapping.PositiveAnswerType != "undercutter" {
		t.Errorf("PositiveAnswerType = %q, want %q", cq.CaesMapping.PositiveAnswerType, "undercutter")
	}
	if cq.CaesMapping.NegativeAnswerType != "support" {
		t.Errorf("NegativeAnswerType = %q, want %q", cq.CaesMapping.NegativeAnswerType, "support")
	}
	if cq.CaesMapping.Target != "argument" {
		t.Errorf("Target = %q, want %q", cq.CaesMapping.Target, "argument")
	}
}

func TestCQPossibilityUnderminingMapping(t *testing.T) {
	n := loadTestNASCQ(t)

	cq := n.GetCQ("CQ:possibility")
	if cq == nil {
		t.Fatal("CQ:possibility not found")
	}
	if cq.CaesMapping.NegativeAnswerType != "undermining" {
		t.Errorf("NegativeAnswerType = %q, want %q", cq.CaesMapping.NegativeAnswerType, "undermining")
	}
	if cq.CaesMapping.Target != "premise" {
		t.Errorf("Target = %q, want %q", cq.CaesMapping.Target, "premise")
	}
	if cq.CaesMapping.TargetPremisePattern != "possible({Action})" {
		t.Errorf("TargetPremisePattern = %q, want %q",
			cq.CaesMapping.TargetPremisePattern, "possible({Action})")
	}
}

func TestThesisIndependentSchemes(t *testing.T) {
	n := loadTestNASCQ(t)

	// expert_opinion and analogy should be thesis-independent.
	eo := n.GetScheme("AS:expert_opinion")
	if eo == nil || !eo.ThesisIndependent {
		t.Error("AS:expert_opinion should be thesis-independent")
	}
	an := n.GetScheme("AS:analogy")
	if an == nil || !an.ThesisIndependent {
		t.Error("AS:analogy should be thesis-independent")
	}

	// practical_reasoning should NOT be thesis-independent.
	pr := n.GetScheme("AS:practical_reasoning")
	if pr == nil || pr.ThesisIndependent {
		t.Error("AS:practical_reasoning should NOT be thesis-independent")
	}

	stats := n.Statistics()
	if stats.ThesisIndependentCount != 2 {
		t.Errorf("thesis-independent count = %d, want 2", stats.ThesisIndependentCount)
	}
}

// ---------------------------------------------------------------------------
// Tests: Ask index queries
// ---------------------------------------------------------------------------

func TestGetAskCQs(t *testing.T) {
	n := loadTestNASCQ(t)

	// Practical reasoning should have 4 CQs.
	cqs := n.GetAskCQs("AS:practical_reasoning")
	if len(cqs) != 4 {
		t.Errorf("AS:practical_reasoning ask CQ count = %d, want 4", len(cqs))
		for _, cq := range cqs {
			t.Logf("  CQ: %s", cq.ID)
		}
	}

	// Verify specific CQs are present.
	cqIDs := make(map[string]bool)
	for _, cq := range cqs {
		cqIDs[cq.ID] = true
	}
	for _, expectedCQ := range []string{"CQ:other_means", "CQ:other_goals", "CQ:side_effects", "CQ:possibility"} {
		if !cqIDs[expectedCQ] {
			t.Errorf("missing expected CQ %q for AS:practical_reasoning", expectedCQ)
		}
	}

	// Expert opinion should have 1 CQ.
	eoCQs := n.GetAskCQs("AS:expert_opinion")
	if len(eoCQs) != 1 {
		t.Errorf("AS:expert_opinion ask CQ count = %d, want 1", len(eoCQs))
	}
	if len(eoCQs) > 0 && eoCQs[0].ID != "CQ:expert_credibility" {
		t.Errorf("expected CQ:expert_credibility, got %q", eoCQs[0].ID)
	}

	// Nonexistent scheme should return empty.
	unknownCQs := n.GetAskCQs("AS:nonexistent")
	if len(unknownCQs) != 0 {
		t.Errorf("nonexistent scheme should return 0 CQs, got %d", len(unknownCQs))
	}
}

// ---------------------------------------------------------------------------
// Tests: Backup index queries
// ---------------------------------------------------------------------------

func TestGetBackupSchemes(t *testing.T) {
	n := loadTestNASCQ(t)

	// CQ:side_effects has 1 explicit backup (cause_to_effect) + 2 thesis-independent.
	backups := n.GetBackupSchemes("CQ:side_effects")
	if len(backups) != 3 {
		t.Errorf("CQ:side_effects backup count = %d, want 3", len(backups))
		for _, s := range backups {
			t.Logf("  backup: %s (TI=%v)", s.ID, s.ThesisIndependent)
		}
	}

	// Verify cause_to_effect is in the list.
	foundCTE := false
	for _, s := range backups {
		if s.ID == "AS:cause_to_effect" {
			foundCTE = true
		}
	}
	if !foundCTE {
		t.Error("CQ:side_effects should have AS:cause_to_effect as backup")
	}

	// CQ:expert_credibility has NO explicit backup edges, but should still
	// get the 2 thesis-independent schemes.
	ecBackups := n.GetBackupSchemes("CQ:expert_credibility")
	if len(ecBackups) != 2 {
		t.Errorf("CQ:expert_credibility backup count = %d, want 2 (thesis-independent only)",
			len(ecBackups))
		for _, s := range ecBackups {
			t.Logf("  backup: %s (TI=%v)", s.ID, s.ThesisIndependent)
		}
	}
	// Both should be thesis-independent.
	for _, s := range ecBackups {
		if !s.ThesisIndependent {
			t.Errorf("backup %q for CQ:expert_credibility should be thesis-independent", s.ID)
		}
	}
}

func TestGetBackupSchemesNoDuplicates(t *testing.T) {
	n := loadTestNASCQ(t)

	// A thesis-independent scheme with an explicit backup edge should
	// NOT appear twice in the results.
	// In our fixture, expert_opinion is thesis-independent but has no
	// explicit backup edges. Let's check any CQ's result for duplicates.
	for cqID := range n.Questions {
		backups := n.GetBackupSchemes(cqID)
		seen := make(map[string]bool)
		for _, s := range backups {
			if seen[s.ID] {
				t.Errorf("CQ %q: duplicate backup scheme %q", cqID, s.ID)
			}
			seen[s.ID] = true
		}
	}
}

// ---------------------------------------------------------------------------
// Tests: Bipartite validation
// ---------------------------------------------------------------------------

func TestValidateBipartite(t *testing.T) {
	n := loadTestNASCQ(t)
	errors := n.ValidateBipartite()
	if len(errors) > 0 {
		t.Errorf("bipartite validation should pass, got %d errors:", len(errors))
		for _, e := range errors {
			t.Logf("  %s", e)
		}
	}
}

func TestValidateBipartiteDetectsBadEdge(t *testing.T) {
	// Create edges referencing nonexistent nodes.
	badEdges := `{
		"ask_edges": [
			{ "cq": "CQ:nonexistent", "scheme": "AS:practical_reasoning" }
		],
		"backup_edges": [
			{ "scheme": "AS:nonexistent_scheme", "cq": "CQ:side_effects" }
		]
	}`
	n, err := LoadFromBytes(
		[]byte(testSchemesJSON),
		[]byte(testQuestionsJSON),
		[]byte(badEdges),
	)
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	errors := n.ValidateBipartite()
	if len(errors) != 2 {
		t.Errorf("expected 2 validation errors, got %d", len(errors))
		for _, e := range errors {
			t.Logf("  %s", e)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests: Statistics
// ---------------------------------------------------------------------------

func TestStatistics(t *testing.T) {
	n := loadTestNASCQ(t)
	s := n.Statistics()

	if s.SchemeCount != 5 {
		t.Errorf("SchemeCount = %d, want 5", s.SchemeCount)
	}
	if s.QuestionCount != 7 {
		t.Errorf("QuestionCount = %d, want 7", s.QuestionCount)
	}
	if s.AskEdgeCount != 7 {
		t.Errorf("AskEdgeCount = %d, want 7", s.AskEdgeCount)
	}
	if s.BackupEdgeCount != 4 {
		t.Errorf("BackupEdgeCount = %d, want 4", s.BackupEdgeCount)
	}
	if s.ThesisIndependentCount != 2 {
		t.Errorf("ThesisIndependentCount = %d, want 2", s.ThesisIndependentCount)
	}
}

// ---------------------------------------------------------------------------
// Tests: Error cases
// ---------------------------------------------------------------------------

func TestLoadInvalidSchemesJSON(t *testing.T) {
	_, err := LoadFromBytes([]byte("not json"), []byte(testQuestionsJSON), []byte(testEdgesJSON))
	if err == nil {
		t.Error("expected error for invalid schemes JSON")
	}
}

func TestLoadInvalidQuestionsJSON(t *testing.T) {
	_, err := LoadFromBytes([]byte(testSchemesJSON), []byte("not json"), []byte(testEdgesJSON))
	if err == nil {
		t.Error("expected error for invalid questions JSON")
	}
}

func TestLoadInvalidEdgesJSON(t *testing.T) {
	_, err := LoadFromBytes([]byte(testSchemesJSON), []byte(testQuestionsJSON), []byte("not json"))
	if err == nil {
		t.Error("expected error for invalid edges JSON")
	}
}

// ---------------------------------------------------------------------------
// Tests: Python prototype format compatibility
// ---------------------------------------------------------------------------

func TestLoadPrototypeFormat(t *testing.T) {
	// Simulate the Python prototype's array-based format.
	protoSchemes := `{
		"argumentation_schemes": [
			{
				"id": "AS:test_scheme",
				"name": "Test Scheme",
				"thesis_independent": false,
				"semantic_type": "action",
				"base_weight": 0.5,
				"variables": {},
				"premise_templates_en": ["Test premise"],
				"conclusion_template_en": "Test conclusion"
			}
		]
	}`
	protoQuestions := `{
		"critical_questions": [
			{
				"id": "CQ:test_cq",
				"name": "Test CQ",
				"semantic_type": "existence",
				"question_template_en": "Test question?",
				"caes_mapping": {
					"positive_answer_type": "support",
					"negative_answer_type": "undercutter",
					"target": "argument"
				}
			}
		]
	}`
	protoEdges := `{
		"edges": [
			{
				"id": "E001",
				"from": "CQ:test_cq",
				"to": "AS:test_scheme",
				"type": "ask"
			}
		]
	}`

	n, err := LoadFromBytes([]byte(protoSchemes), []byte(protoQuestions), []byte(protoEdges))
	if err != nil {
		t.Fatalf("LoadFromBytes (prototype format) failed: %v", err)
	}

	if len(n.Schemes) != 1 {
		t.Errorf("scheme count = %d, want 1", len(n.Schemes))
	}
	if len(n.Questions) != 1 {
		t.Errorf("question count = %d, want 1", len(n.Questions))
	}
	if len(n.AskEdges) != 1 {
		t.Errorf("ask edge count = %d, want 1", len(n.AskEdges))
	}

	// Verify index was built.
	cqs := n.GetAskCQs("AS:test_scheme")
	if len(cqs) != 1 {
		t.Errorf("ask CQ count for test_scheme = %d, want 1", len(cqs))
	}
}
