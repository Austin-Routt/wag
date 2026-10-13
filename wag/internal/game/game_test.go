package game

// Integration tests for the WAG game state and turn pipeline.
//
// These tests exercise the FULL stack: NASCQ validation → adapter mutation →
// Carneades evaluation → scoring → termination. They prove that all four
// blocks work together correctly.
//
// Run with:
//   go test ./internal/game/ -v

import (
	"math"
	"testing"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// -------------------------------------------------------------------------
// Test fixtures: NASCQ JSON (same data as Block 3/4 tests)
// -------------------------------------------------------------------------

const testSchemesJSON = `{
  "AS:practical_reasoning": {
    "carneades_id": "practical_reasoning1",
    "name": "Practical Reasoning",
    "semantic_type": "action",
    "thesis_independent": false,
    "base_weight": 0.70,
    "variables": {
      "Agent": { "type": "entity", "description": "Who should act" },
      "Goal":  { "type": "goal",   "description": "What to achieve" },
      "Action": { "type": "action", "description": "What to do" }
    },
    "premise_templates_en": [
      "{Agent} has goal {Goal}",
      "{Action} is a means to achieve {Goal}"
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
      "Domain":      { "type": "domain",      "description": "Field" },
      "Proposition": { "type": "proposition", "description": "Claim" }
    },
    "premise_templates_en": [
      "{Expert} is an expert in {Domain}",
      "{Expert} asserts that {Proposition}"
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
      "If {Cause} occurs then {Effect} occurs",
      "In this case {Cause} occurs"
    ],
    "conclusion_template_en": "{Effect} will occur"
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
    "conclusion_template_en": "{Action} promotes {Value}"
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
    "question_template_en": "Are there alternatives to {Action}?",
    "caes_mapping": {
      "positive_answer_type": "rebuttal",
      "negative_answer_type": "support",
      "target": "conclusion"
    }
  },
  "CQ:other_goals": {
    "name": "Other Goals Question",
    "semantic_type": "value",
    "question_template_en": "Are there goals other than {Goal}?",
    "caes_mapping": {
      "positive_answer_type": "rebuttal",
      "negative_answer_type": "support",
      "target": "conclusion"
    }
  },
  "CQ:side_effects": {
    "name": "Side Effects Question",
    "semantic_type": "possibility",
    "question_template_en": "Would {Action} have side effects?",
    "caes_mapping": {
      "positive_answer_type": "undercutter",
      "negative_answer_type": "support",
      "target": "argument"
    }
  },
  "CQ:possibility": {
    "name": "Possibility Question",
    "semantic_type": "possibility",
    "question_template_en": "Is {Action} possible?",
    "caes_mapping": {
      "positive_answer_type": "support",
      "negative_answer_type": "undermining",
      "target": "premise"
    }
  },
  "CQ:expert_credibility": {
    "name": "Expert Credibility Question",
    "semantic_type": "property",
    "question_template_en": "Is {Expert} credible in {Domain}?",
    "caes_mapping": {
      "positive_answer_type": "support",
      "negative_answer_type": "undercutter",
      "target": "argument"
    }
  },
  "CQ:interference": {
    "name": "Interference Question",
    "semantic_type": "causality",
    "question_template_en": "Could something prevent {Cause} from causing {Effect}?",
    "caes_mapping": {
      "positive_answer_type": "undercutter",
      "negative_answer_type": "support",
      "target": "argument"
    }
  },
  "CQ:legitimate_value": {
    "name": "Legitimate Value Question",
    "semantic_type": "value",
    "question_template_en": "Is {Value} legitimate?",
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

// -------------------------------------------------------------------------
// Test helpers
// -------------------------------------------------------------------------

// makeTestAdapter creates an adapter with Carneades schemes whose IDs and
// variable names match the NASCQ test fixtures. This is the critical bridge:
// NASCQ variable names MUST match Carneades scheme variable names for
// AddSupport to work correctly.
func makeTestAdapter(t *testing.T) *carneades.Adapter {
	t.Helper()

	ag := caes.NewArgGraph()

	// Scheme matching NASCQ AS:practical_reasoning (carneades_id: "practical_reasoning1")
	// Variables MUST match NASCQ: Agent, Goal, Action
	s1 := &caes.Scheme{
		Id:          "practical_reasoning1",
		Variables:   []string{"Agent", "Goal", "Action"},
		Premises:    []string{"goal(Goal)"},
		Conclusions: []string{"should_do(Action)"},
		Assumptions: []string{"goal(Goal)"},
	}

	// Scheme matching NASCQ AS:cause_to_effect (carneades_id: "cause_to_effect")
	s2 := &caes.Scheme{
		Id:          "cause_to_effect",
		Variables:   []string{"Cause", "Effect"},
		Premises:    []string{"causes(Cause,Effect)"},
		Conclusions: []string{"will_occur(Effect)"},
		Assumptions: []string{"causes(Cause,Effect)"},
	}

	ag.Theory.ArgSchemes = append(ag.Theory.ArgSchemes, s1, s2)
	ag.Theory.InitSchemeIndex()

	a := carneades.NewAdapterWithGraph(ag)
	return a
}

// makeTestNASCQ loads the NASCQ test fixture.
func makeTestNASCQ(t *testing.T) *nascq.NASCQ {
	t.Helper()
	n, err := nascq.LoadFromBytes(
		[]byte(testSchemesJSON),
		[]byte(testQuestionsJSON),
		[]byte(testEdgesJSON),
	)
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}
	return n
}

// makeTestGame creates a complete game with 2 players (Alice and Bob),
// matching adapter schemes and NASCQ data, and default config.
func makeTestGame(t *testing.T) *GameState {
	t.Helper()
	adapter := makeTestAdapter(t)
	n := makeTestNASCQ(t)
	config := DefaultGameConfig()
	config.MaxTurns = 10
	return NewGame(config, adapter, n, []string{"Alice", "Bob"})
}

// approxEqual checks if two floats are within epsilon of each other.
func approxEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// -------------------------------------------------------------------------
// Tests
// -------------------------------------------------------------------------

func TestNewGame(t *testing.T) {
	gs := makeTestGame(t)

	if len(gs.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(gs.Players))
	}
	if gs.Players[0].ID != "Alice" {
		t.Errorf("player 0 ID = %q, want Alice", gs.Players[0].ID)
	}
	if gs.Players[1].ID != "Bob" {
		t.Errorf("player 1 ID = %q, want Bob", gs.Players[1].ID)
	}
	if gs.CurrentTurn != 1 {
		t.Errorf("initial turn = %d, want 1", gs.CurrentTurn)
	}
	if gs.ActivePlayerIdx != 0 {
		t.Errorf("initial active player = %d, want 0", gs.ActivePlayerIdx)
	}
	if gs.Terminated {
		t.Error("game should not be terminated at start")
	}
	if gs.ActivePlayer().ID != "Alice" {
		t.Errorf("active player = %q, want Alice", gs.ActivePlayer().ID)
	}
}

func TestSupportTurn(t *testing.T) {
	gs := makeTestGame(t)

	// Alice supports: practical reasoning for carbon tax.
	action := nascq.ProposedAction{
		Type:     "support",
		SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{
			"Agent":  "government",
			"Goal":   "reduce_emissions",
			"Action": "carbon_tax",
		},
	}

	result, err := gs.ExecuteTurn(action)
	if err != nil {
		t.Fatalf("ExecuteTurn failed: %v", err)
	}
	if !result.Accepted {
		t.Fatalf("move should be accepted, error: %s", result.Error)
	}
	if len(result.ArgIDs) == 0 {
		t.Fatal("expected at least one argument created")
	}

	// Verify the conclusion is labeled In after evaluation.
	// InstantiateScheme creates should_do(carbon_tax) as the conclusion.
	foundIn := false
	for id, stmt := range gs.Adapter.Graph.Statements {
		if stmt.Label == caes.In && id != "goal(reduce_emissions)" {
			// Skip the assumed premise; look for the conclusion.
			foundIn = true
			t.Logf("Conclusion %q labeled In", id)
		}
	}
	if !foundIn {
		t.Error("expected at least one non-assumed statement labeled In (the conclusion)")
		for id, stmt := range gs.Adapter.Graph.Statements {
			t.Logf("  %q: label=%v", id, stmt.Label)
		}
	}

	// Verify scoring: base_points[In] * weight = 10 * 0.7 = 7.0
	alice := gs.PlayerByID("Alice")
	if !approxEqual(alice.Score, 7.0, 0.1) {
		t.Errorf("Alice score = %.2f, want ~7.0", alice.Score)
	}
	if !approxEqual(result.ScoreDelta, 7.0, 0.1) {
		t.Errorf("score delta = %.2f, want ~7.0", result.ScoreDelta)
	}

	// Verify turn advanced to Bob.
	if gs.ActivePlayer().ID != "Bob" {
		t.Errorf("active player = %q, want Bob", gs.ActivePlayer().ID)
	}

	// Verify move history.
	if len(gs.MoveHistory) != 1 {
		t.Fatalf("move history length = %d, want 1", len(gs.MoveHistory))
	}
	if gs.MoveHistory[0].PlayerID != "Alice" {
		t.Errorf("move 0 player = %q, want Alice", gs.MoveHistory[0].PlayerID)
	}
}

func TestUndercutterFlipsLabel(t *testing.T) {
	gs := makeTestGame(t)

	// Turn 1: Alice supports with practical reasoning.
	support := nascq.ProposedAction{
		Type:     "support",
		SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{
			"Agent":  "government",
			"Goal":   "reduce_emissions",
			"Action": "carbon_tax",
		},
	}
	r1, err := gs.ExecuteTurn(support)
	if err != nil || !r1.Accepted {
		t.Fatalf("Turn 1 failed: err=%v accepted=%v error=%s", err, r1.Accepted, r1.Error)
	}
	targetArgID := r1.ArgIDs[0]
	t.Logf("Turn 1: Alice created argument %s", targetArgID)

	// Verify conclusion is In before attack.
	targetArg := gs.Adapter.Graph.Arguments[targetArgID]
	if targetArg == nil {
		t.Fatalf("target argument %q not found", targetArgID)
	}
	if targetArg.Conclusion.Label != caes.In {
		t.Fatalf("before attack: conclusion = %v, want In", targetArg.Conclusion.Label)
	}

	// Turn 2: Bob undercuts via CQ:side_effects using AS:cause_to_effect.
	attack := nascq.ProposedAction{
		Type:        "attack",
		AttackType:  "undercutter",
		SchemeID:    "AS:cause_to_effect",
		AnswersCQID: "CQ:side_effects",
		TargetArgID: targetArgID,
		VariableBindings: map[string]string{
			"Cause":  "carbon_tax",
			"Effect": "economic_recession",
		},
	}
	r2, err := gs.ExecuteTurn(attack)
	if err != nil || !r2.Accepted {
		t.Fatalf("Turn 2 failed: err=%v accepted=%v error=%s", err, r2.Accepted, r2.Error)
	}
	t.Logf("Turn 2: Bob created arguments %v", r2.ArgIDs)

	// After undercutter: Alice's conclusion should flip to Out.
	// Trace:
	//   1. Attack premises assumed In → undercutter argument applicable, weight > 0
	//   2. undercut statement becomes In (supported)
	//   3. Alice's argument GetWeight → 0.0 (undercut is In)
	//   4. Alice's conclusion unsupported → Out
	if targetArg.Conclusion.Label != caes.Out {
		t.Errorf("after undercutter: conclusion = %v, want Out", targetArg.Conclusion.Label)
		for id, stmt := range gs.Adapter.Graph.Statements {
			t.Logf("  %q: label=%v", id, stmt.Label)
		}
	}

	// Verify Bob's score includes undercutter bonus.
	// Score: base_points[In] * 0.65 (for undercut statement labeled In) + 15 (bonus)
	// = 10 * 0.65 + 15 = 6.5 + 15 = 21.5
	bob := gs.PlayerByID("Bob")
	if !approxEqual(bob.Score, 21.5, 0.5) {
		t.Errorf("Bob score = %.2f, want ~21.5", bob.Score)
	}

	// Verify graph is still acyclic.
	if !gs.Adapter.IsAcyclic() {
		t.Error("graph should remain acyclic")
	}
}

func TestAbstainAndTermination(t *testing.T) {
	gs := makeTestGame(t)
	gs.Config.AbstentionThreshold = 2

	// Round 1: Both abstain.
	for i := 0; i < 2; i++ {
		r, err := gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
		if err != nil {
			t.Fatalf("abstain %d failed: %v", i, err)
		}
		if !r.Accepted || !r.Abstained {
			t.Fatalf("abstain %d not accepted", i)
		}
	}
	// After round 1: ConsecAllAbstain = 1, not yet terminated.
	if gs.Terminated {
		t.Fatal("should not terminate after 1 round of all-abstain")
	}
	if gs.ConsecAllAbstain != 1 {
		t.Errorf("ConsecAllAbstain = %d, want 1", gs.ConsecAllAbstain)
	}

	// Round 2: Both abstain again → should trigger termination.
	for i := 0; i < 2; i++ {
		r, err := gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
		if err != nil {
			t.Fatalf("abstain round 2 turn %d failed: %v", i, err)
		}
		if i == 1 { // Last abstain of the round should trigger termination.
			if !r.Terminated {
				t.Error("game should terminate after 2 rounds of all-abstain")
			}
		}
	}

	if !gs.Terminated {
		t.Error("game should be terminated")
	}
	if gs.TerminationReason == "" {
		t.Error("termination reason should be set")
	}
	t.Logf("Termination reason: %s", gs.TerminationReason)
}

func TestInvalidActionRejected(t *testing.T) {
	gs := makeTestGame(t)

	// Submit an action with a nonexistent scheme → S1 failure.
	action := nascq.ProposedAction{
		Type:             "support",
		SchemeID:         "AS:nonexistent",
		VariableBindings: map[string]string{"X": "value"},
	}
	result, err := gs.ExecuteTurn(action)
	if err != nil {
		t.Fatalf("ExecuteTurn should not return error for invalid action: %v", err)
	}
	if result.Accepted {
		t.Fatal("invalid action should not be accepted")
	}
	if result.Error == "" {
		t.Error("error message should be set")
	}
	t.Logf("Rejection reason: %s", result.Error)

	// Verify penalty applied.
	alice := gs.PlayerByID("Alice")
	if alice.Score != gs.Config.Scoring.InvalidActionPenalty {
		t.Errorf("Alice score = %.2f, want %.2f (penalty)",
			alice.Score, gs.Config.Scoring.InvalidActionPenalty)
	}

	// Turn should NOT advance — player can retry.
	if gs.ActivePlayer().ID != "Alice" {
		t.Errorf("active player = %q, want Alice (invalid move doesn't consume turn)",
			gs.ActivePlayer().ID)
	}
}

func TestFullGameSequence(t *testing.T) {
	gs := makeTestGame(t)
	gs.Config.MaxTurns = 4

	// Turn 1: Alice supports.
	r1, _ := gs.ExecuteTurn(nascq.ProposedAction{
		Type:     "support",
		SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{
			"Agent": "government", "Goal": "reduce_emissions", "Action": "carbon_tax",
		},
	})
	if !r1.Accepted {
		t.Fatalf("turn 1 rejected: %s", r1.Error)
	}
	t.Logf("Turn 1 (Alice support): args=%v score=%.2f", r1.ArgIDs, r1.ScoreDelta)

	// Turn 2: Bob undercuts.
	r2, _ := gs.ExecuteTurn(nascq.ProposedAction{
		Type: "attack", AttackType: "undercutter",
		SchemeID: "AS:cause_to_effect", AnswersCQID: "CQ:side_effects",
		TargetArgID: r1.ArgIDs[0],
		VariableBindings: map[string]string{
			"Cause": "carbon_tax", "Effect": "job_losses",
		},
	})
	if !r2.Accepted {
		t.Fatalf("turn 2 rejected: %s", r2.Error)
	}
	t.Logf("Turn 2 (Bob undercutter): args=%v score=%.2f", r2.ArgIDs, r2.ScoreDelta)

	// Turn 3: Alice abstains.
	r3, _ := gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	if !r3.Accepted || !r3.Abstained {
		t.Fatal("turn 3 abstain failed")
	}
	t.Logf("Turn 3 (Alice abstain)")

	// Turn 4: Bob abstains.
	r4, _ := gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	if !r4.Accepted || !r4.Abstained {
		t.Fatal("turn 4 abstain failed")
	}
	t.Logf("Turn 4 (Bob abstain)")

	// After turn 4, we're at the end of round 3 (turns: 1=Alice, 2=Bob |
	// round boundary | 3=Alice, 4=Bob | round boundary → turn counter = 3).
	// MaxTurns is 4, currentTurn should now be 3. Let's check.
	// Actually the turn counter advances per round, not per individual turn.

	// Verify scores.
	alice := gs.PlayerByID("Alice")
	bob := gs.PlayerByID("Bob")
	t.Logf("Final scores: Alice=%.2f, Bob=%.2f", alice.Score, bob.Score)

	if alice.Score <= 0 {
		t.Errorf("Alice should have positive score from turn 1, got %.2f", alice.Score)
	}
	if bob.Score <= alice.Score {
		t.Errorf("Bob should outscore Alice (undercutter bonus), got Bob=%.2f Alice=%.2f",
			bob.Score, alice.Score)
	}

	// Verify move history.
	if len(gs.MoveHistory) != 4 {
		t.Errorf("move history length = %d, want 4", len(gs.MoveHistory))
	}

	// Log final graph state for debugging.
	t.Logf("Final graph: %d statements, %d arguments",
		len(gs.Adapter.Graph.Statements), len(gs.Adapter.Graph.Arguments))
	for id, stmt := range gs.Adapter.Graph.Statements {
		t.Logf("  stmt %q: label=%v text=%q", id, stmt.Label, stmt.Text)
	}
}

func TestTerminationMaxTurns(t *testing.T) {
	gs := makeTestGame(t)
	gs.Config.MaxTurns = 2 // Very short game: 2 rounds.

	// Play 4 abstains (2 rounds of 2 players). This should hit max turns.
	// But also hits all-abstain. Let's set threshold high so only max-turns fires.
	gs.Config.AbstentionThreshold = 100

	for i := 0; i < 4; i++ {
		r, _ := gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
		if r.Terminated {
			t.Logf("Terminated at individual turn %d: %s", i+1, r.TerminationReason)
			break
		}
	}

	if !gs.Terminated {
		t.Errorf("game should terminate at max turns; currentTurn=%d", gs.CurrentTurn)
	}
}

func TestTerminatedGameRejectsMoves(t *testing.T) {
	gs := makeTestGame(t)
	gs.Terminated = true
	gs.TerminationReason = "test termination"

	_, err := gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	if err == nil {
		t.Error("terminated game should reject moves with an error")
	}
}
