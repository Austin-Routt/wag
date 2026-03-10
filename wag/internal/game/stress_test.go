package game

// Stress tests using random and greedy agents playing complete games.
// These tests exercise the FULL engine under adversarial random play to
// find panics, deadlocks, state corruption, and edge cases that no
// hand-written test can discover.
//
// Run with:
//   go test ./internal/game/ -v -run TestStress
//   go test ./internal/game/ -v -run TestAgent

import (
    "fmt"
    "testing"
    "github.com/yourorg/wag/internal/carneades"
    "github.com/yourorg/wag/internal/nascq"
)

// =========================================================================
// Shared test data (same as core_integration_test.go)
// =========================================================================

const stressYAML = `
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
    variables: [A,G]
    conclusions:
      - should_be_performed(A)
    premises:
      - brings_about(A,G)
      - good(G)

  - id: positive_value
    variables: [V,A,G]
    conclusions:
      - reason_retaining_goal(V,G)
    premises:
      - judged_positive_value(V,A)
      - positive_affects(V,G,A)

  - id: sign
    variables: [A,B]
    conclusions:
      - true_in_situation(B)
    premises:
      - finding_true(A)
      - indicated_true(B,A)
    assumptions:
      - strong_correlation(A,B)

  - id: credible_source
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
    variables: [W,D,S]
    conclusions: [S]
    premises:
      - position_to_know(W,D)
      - asserts(W,S)
      - in_domain(S,D)
    exceptions:
      - dishonest(W)

  - id: analogy
    variables: [A,C]
    conclusions: [A]
    premises:
      - similar_case(C)
      - in_case(A,C)
    exceptions:
      - relevant_differences(C)
      - more_on_point(A,C)
`

const stressSchemesJSON = `{
  "AS:practical_reasoning": {
    "carneades_id": "practical_reasoning1",
    "name": "Practical Reasoning",
    "semantic_type": "action",
    "thesis_independent": false,
    "base_weight": 0.70,
    "variables": {
      "A":  { "type": "action", "description": "Action" },
      "S1": { "type": "state",  "description": "Current state" },
      "S2": { "type": "state",  "description": "Future state" },
      "G":  { "type": "goal",   "description": "Goal" },
      "V":  { "type": "value",  "description": "Value" }
    },
    "premise_templates_en": [
      "In circumstances {S1}",
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
      "G": { "type": "event",  "description": "Consequence" }
    },
    "premise_templates_en": ["Action {A} brings about {G}", "{G} is good"],
    "conclusion_template_en": "Action {A} should be performed"
  },
  "AS:values": {
    "carneades_id": "positive_value",
    "name": "Argument from Values",
    "semantic_type": "value",
    "thesis_independent": false,
    "base_weight": 0.60,
    "variables": {
      "V": { "type": "value",  "description": "Value" },
      "A": { "type": "entity", "description": "Agent" },
      "G": { "type": "goal",   "description": "Goal" }
    },
    "premise_templates_en": ["Agent {A} judges {V} positive", "{V} affects goal {G} for {A}"],
    "conclusion_template_en": "Retain goal {G} because of {V}"
  },
  "AS:sign": {
    "carneades_id": "sign",
    "name": "Argument from Sign",
    "semantic_type": "causality",
    "thesis_independent": false,
    "base_weight": 0.60,
    "variables": {
      "A": { "type": "event", "description": "Sign" },
      "B": { "type": "event", "description": "Condition" }
    },
    "premise_templates_en": ["{A} is observed", "{B} is indicated when {A} is true"],
    "conclusion_template_en": "{B} is true"
  },
  "AS:expert_opinion": {
    "carneades_id": "credible_source",
    "name": "Expert Opinion",
    "semantic_type": "truth",
    "thesis_independent": true,
    "base_weight": 0.70,
    "variables": {
      "W": { "type": "entity",      "description": "Expert" },
      "D": { "type": "domain",      "description": "Domain" },
      "S": { "type": "proposition", "description": "Claim" }
    },
    "premise_templates_en": ["{W} is credible in {D}", "{W} asserts {S}", "{S} is in {D}"],
    "conclusion_template_en": "{S}"
  },
  "AS:position_to_know": {
    "carneades_id": "position_to_know",
    "name": "Position to Know",
    "semantic_type": "truth",
    "thesis_independent": true,
    "base_weight": 0.65,
    "variables": {
      "W": { "type": "entity",      "description": "Source" },
      "D": { "type": "domain",      "description": "Domain" },
      "S": { "type": "proposition", "description": "Claim" }
    },
    "premise_templates_en": ["{W} knows about {D}", "{W} asserts {S}", "{S} is in {D}"],
    "conclusion_template_en": "{S}"
  },
  "AS:analogy": {
    "carneades_id": "analogy",
    "name": "Argument from Analogy",
    "semantic_type": "truth",
    "thesis_independent": true,
    "base_weight": 0.55,
    "variables": {
      "A": { "type": "proposition", "description": "Claim" },
      "C": { "type": "case",        "description": "Similar case" }
    },
    "premise_templates_en": ["{C} is a similar case", "{A} is true in {C}"],
    "conclusion_template_en": "{A}"
  }
}`

const stressCQsJSON = `{
  "CQ:other_goals":  { "name": "Other Goals",   "semantic_type": "value",     "question_template_en": "Other goals?",     "caes_mapping": { "positive_answer_type": "rebuttal",    "negative_answer_type": "support",     "target": "conclusion" } },
  "CQ:other_means":  { "name": "Other Means",   "semantic_type": "action",    "question_template_en": "Other means?",     "caes_mapping": { "positive_answer_type": "rebuttal",    "negative_answer_type": "support",     "target": "conclusion" } },
  "CQ:side_effect":  { "name": "Side Effects",  "semantic_type": "causality", "question_template_en": "Side effects?",    "caes_mapping": { "positive_answer_type": "undercutter", "negative_answer_type": "support",     "target": "argument" } },
  "CQ:possibility":  { "name": "Possibility",   "semantic_type": "causality", "question_template_en": "Possible?",        "caes_mapping": { "positive_answer_type": "support",     "negative_answer_type": "undermining", "target": "premise" } },
  "CQ:best_means":   { "name": "Best Means",    "semantic_type": "action",    "question_template_en": "Best means?",      "caes_mapping": { "positive_answer_type": "support",     "negative_answer_type": "rebuttal",    "target": "conclusion" } },
  "CQ:evidence":     { "name": "Evidence",       "semantic_type": "truth",     "question_template_en": "Evidence?",        "caes_mapping": { "positive_answer_type": "support",     "negative_answer_type": "undercutter", "target": "argument" } },
  "CQ:likelihood":   { "name": "Likelihood",    "semantic_type": "truth",     "question_template_en": "How likely?",      "caes_mapping": { "positive_answer_type": "support",     "negative_answer_type": "undercutter", "target": "argument" } }
}`

const stressEdgesJSON = `{
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
// Test: Action enumeration
// =========================================================================

func TestAgentActionEnumeration(t *testing.T) {
	gs := makeCoreGame(t)

	// Before any moves: only direct support actions available (no arguments to attack).
	obs := gs.BuildObservation()
	if len(obs.AvailableActions) == 0 {
		t.Fatal("expected at least direct support actions")
	}

	supportCount := 0
	attackCount := 0
	for _, a := range obs.AvailableActions {
		if a.Type == "support" {
			supportCount++
		} else {
			attackCount++
		}
	}
	t.Logf("Before any moves: %d support templates, %d attack templates",
		supportCount, attackCount)

	if supportCount != 7 { // one per scheme
		t.Errorf("expected 7 support templates (one per scheme), got %d", supportCount)
	}
	if attackCount != 0 {
		t.Errorf("expected 0 attack templates (no arguments yet), got %d", attackCount)
	}

	// After a support move: attack templates should appear.
	gs.ExecuteTurn(nascqAction("support", "AS:practical_reasoning", map[string]string{
		"A": "carbon_tax", "S1": "now", "S2": "future",
		"G": "reduce_emissions", "V": "sustainability",
	}))

	obs2 := gs.BuildObservation()
	attack2 := 0
	for _, a := range obs2.AvailableActions {
		if a.Type == "attack" {
			attack2++
		}
	}
	t.Logf("After 1 support: %d total templates (%d attack)", len(obs2.AvailableActions), attack2)
	if attack2 == 0 {
		t.Error("expected attack templates after an argument was created")
	}
}

// helper to build a ProposedAction concisely
func nascqAction(actionType, schemeID string, bindings map[string]string) nascq.ProposedAction {
	return nascq.ProposedAction{
		Type:             actionType,
		SchemeID:         schemeID,
		VariableBindings: bindings,
	}
}

// =========================================================================
// Test: RandomAgent plays a single complete game
// =========================================================================

func TestAgentRandomSingleGame(t *testing.T) {
	result := RunGame(
		GameConfig{MaxTurns: 8, AbstentionThreshold: 3, Scoring: DefaultScoringConfig()},
		DefaultRunnerConfig(),
		[]byte(stressYAML),
		[]byte(stressSchemesJSON), []byte(stressCQsJSON), []byte(stressEdgesJSON),
		[]Agent{
			NewRandomAgent("Alice", 42, 0.1, 0.6),
			NewRandomAgent("Bob", 99, 0.1, 0.4),
		},
	)

	if result.Panicked {
		t.Fatalf("Game panicked: %s", result.PanicMessage)
	}

	t.Logf("Game completed: %d turns, %d moves, %d abstains, %d invalid",
		result.TurnsPlayed, result.MoveCount, result.AbstainCount, result.InvalidCount)
	t.Logf("Scores: %v", result.Scores)
	t.Logf("Winner: %s (%s)", result.Winner, result.Reason)
	t.Logf("Graph: %d statements, %d arguments, %d issues",
		result.StatementCount, result.ArgumentCount, result.IssueCount)

	// Basic sanity checks.
	if result.MoveCount+result.AbstainCount == 0 {
		t.Error("no moves or abstains recorded")
	}
}

// =========================================================================
// Test: GreedyAgent plays a single complete game
// =========================================================================

func TestAgentGreedySingleGame(t *testing.T) {
	result := RunGame(
		GameConfig{MaxTurns: 8, AbstentionThreshold: 3, Scoring: DefaultScoringConfig()},
		DefaultRunnerConfig(),
		[]byte(stressYAML),
		[]byte(stressSchemesJSON), []byte(stressCQsJSON), []byte(stressEdgesJSON),
		[]Agent{
			NewGreedyAgent("GreedyAlice", 42, 3),
			NewRandomAgent("RandomBob", 99, 0.15, 0.5),
		},
	)

	if result.Panicked {
		t.Fatalf("Game panicked: %s", result.PanicMessage)
	}

	t.Logf("Greedy vs Random: %d turns, %d moves, %d abstains, %d invalid",
		result.TurnsPlayed, result.MoveCount, result.AbstainCount, result.InvalidCount)
	t.Logf("Scores: %v", result.Scores)
	t.Logf("Winner: %s", result.Winner)
}

// =========================================================================
// STRESS TEST: 200 random games — the ultimate engine validation
// =========================================================================

func TestStress200RandomGames(t *testing.T) {
	br := RunBatch(
		200,
		GameConfig{MaxTurns: 10, AbstentionThreshold: 3, Scoring: DefaultScoringConfig()},
		DefaultRunnerConfig(),
		[]byte(stressYAML),
		[]byte(stressSchemesJSON), []byte(stressCQsJSON), []byte(stressEdgesJSON),
		func(gameIndex int) []Agent {
			return []Agent{
				NewRandomAgent("P1", int64(gameIndex*2), 0.15, 0.55),
				NewRandomAgent("P2", int64(gameIndex*2+1), 0.15, 0.45),
			}
		},
	)

	t.Logf("=== STRESS TEST RESULTS (200 games) ===")
	t.Logf("Games played:  %d", br.GamesPlayed)
	t.Logf("Panics:        %d", br.Panics)
	t.Logf("Wins:          %v", br.Wins)
	t.Logf("Ties:          %d", br.Ties)
	t.Logf("Total moves:   %d", br.TotalMoves)
	t.Logf("Total abstains:%d", br.TotalAbstains)
	t.Logf("Total invalid: %d", br.TotalInvalid)
	t.Logf("Avg turns:     %.1f", br.AvgTurns)
	t.Logf("Avg statements:%.1f", br.AvgStatements)
	t.Logf("Avg arguments: %.1f", br.AvgArguments)
	t.Logf("Max statements:%d", br.MaxStatements)
	t.Logf("Max arguments: %d", br.MaxArguments)

	// THE critical assertion: zero panics in 200 games.
	if br.Panics > 0 {
		t.Errorf("PANICS DETECTED: %d out of %d games", br.Panics, br.GamesPlayed)
		for i, msg := range br.PanicMessages {
			t.Errorf("  Panic %d: %s", i+1, msg)
		}
	}

	// At least some games should have actual moves (not all abstains).
	if br.TotalMoves == 0 {
		t.Error("no moves made across 200 games — agents may be broken")
	}

	// Both players should win at least some games (no systematic bias).
	if len(br.Wins) < 2 && br.Ties < br.GamesPlayed/2 {
		t.Logf("WARNING: unbalanced wins: %v (ties: %d)", br.Wins, br.Ties)
	}
}

// =========================================================================
// STRESS TEST: 50 greedy-vs-random games
// =========================================================================

func TestStress50GreedyVsRandom(t *testing.T) {
	br := RunBatch(
		50,
		GameConfig{MaxTurns: 8, AbstentionThreshold: 3, Scoring: DefaultScoringConfig()},
		DefaultRunnerConfig(),
		[]byte(stressYAML),
		[]byte(stressSchemesJSON), []byte(stressCQsJSON), []byte(stressEdgesJSON),
		func(gameIndex int) []Agent {
			return []Agent{
				NewGreedyAgent("Greedy", int64(gameIndex*3), 2),
				NewRandomAgent("Random", int64(gameIndex*3+1), 0.2, 0.5),
			}
		},
	)

	t.Logf("=== GREEDY VS RANDOM (50 games) ===")
	t.Logf("Panics: %d", br.Panics)
	t.Logf("Wins:   %v", br.Wins)
	t.Logf("Ties:   %d", br.Ties)
	t.Logf("Avg turns: %.1f", br.AvgTurns)
	t.Logf("Avg args:  %.1f", br.AvgArguments)

	if br.Panics > 0 {
		t.Errorf("PANICS: %d", br.Panics)
		for _, msg := range br.PanicMessages {
			t.Errorf("  %s", msg)
		}
	}
}

// =========================================================================
// Test: Three-player game
// =========================================================================

func TestAgentThreePlayerGame(t *testing.T) {
	result := RunGame(
		GameConfig{MaxTurns: 6, AbstentionThreshold: 2, Scoring: DefaultScoringConfig()},
		DefaultRunnerConfig(),
		[]byte(stressYAML),
		[]byte(stressSchemesJSON), []byte(stressCQsJSON), []byte(stressEdgesJSON),
		[]Agent{
			NewRandomAgent("Alice", 1, 0.1, 0.7),
			NewRandomAgent("Bob", 2, 0.1, 0.3),
			NewRandomAgent("Carol", 3, 0.2, 0.5),
		},
	)

	if result.Panicked {
		t.Fatalf("3-player game panicked: %s", result.PanicMessage)
	}

	t.Logf("3-player game: %d turns, scores=%v, winner=%s",
		result.TurnsPlayed, result.Scores, result.Winner)

	if len(result.Scores) != 3 {
		t.Errorf("expected 3 player scores, got %d", len(result.Scores))
	}
}

// =========================================================================
// Test: Edge case — game with only abstains
// =========================================================================

func TestAgentAllAbstain(t *testing.T) {
	result := RunGame(
		GameConfig{MaxTurns: 20, AbstentionThreshold: 2, Scoring: DefaultScoringConfig()},
		DefaultRunnerConfig(),
		[]byte(stressYAML),
		[]byte(stressSchemesJSON), []byte(stressCQsJSON), []byte(stressEdgesJSON),
		[]Agent{
			NewRandomAgent("Passive1", 1, 1.0, 0.5), // 100% abstain
			NewRandomAgent("Passive2", 2, 1.0, 0.5), // 100% abstain
		},
	)

	if result.Panicked {
		t.Fatalf("all-abstain game panicked: %s", result.PanicMessage)
	}

	t.Logf("All-abstain: terminated=%v reason=%q moves=%d abstains=%d",
		result.Reason != "", result.Reason, result.MoveCount, result.AbstainCount)

	if result.MoveCount != 0 {
		t.Errorf("expected 0 moves in all-abstain game, got %d", result.MoveCount)
	}
	if result.Reason == "" {
		t.Error("all-abstain game should terminate with a reason")
	}
}

// =========================================================================
// Test: Observation structure
// =========================================================================

func TestAgentObservationStructure(t *testing.T) {
	gs := makeCoreGame(t)
	obs := gs.BuildObservation()

	if obs.ActivePlayer == nil {
		t.Fatal("ActivePlayer should not be nil")
	}
	if obs.ActivePlayer.ID != "Alice" {
		t.Errorf("ActivePlayer = %q, want Alice", obs.ActivePlayer.ID)
	}
	if obs.TurnNumber != 1 {
		t.Errorf("TurnNumber = %d, want 1", obs.TurnNumber)
	}
	if len(obs.Scores) != 2 {
		t.Errorf("Scores count = %d, want 2", len(obs.Scores))
	}
	if obs.GameState == nil {
		t.Error("GameState should not be nil")
	}

	// Verify action templates have populated fields.
	for _, a := range obs.AvailableActions {
		if a.SchemeID == "" {
			t.Error("action template has empty SchemeID")
		}
		if a.SchemeName == "" {
			t.Error("action template has empty SchemeName")
		}
		if len(a.Variables) == 0 {
			t.Errorf("action template %q has no variables", a.SchemeID)
		}
	}
}

// =========================================================================
// Test: Print a sample game trace for human review
// =========================================================================

func TestAgentSampleGameTrace(t *testing.T) {
	gameConfig := GameConfig{MaxTurns: 6, AbstentionThreshold: 3, Scoring: DefaultScoringConfig()}
	agents := []Agent{
		NewRandomAgent("Alice", 12345, 0.1, 0.6),
		NewRandomAgent("Bob", 67890, 0.1, 0.4),
	}

	// Create game manually so we can trace moves.
	adapter, _ := carneades.NewAdapterFromYAML([]byte(stressYAML))
	n, _ := nascq.LoadFromBytes([]byte(stressSchemesJSON), []byte(stressCQsJSON), []byte(stressEdgesJSON))
	gs := NewGame(gameConfig, adapter, n, []string{"Alice", "Bob"})

	t.Logf("=== SAMPLE GAME TRACE ===")
	maxIter := 100
	for iter := 0; iter < maxIter && !gs.Terminated; iter++ {
		agent := agents[gs.ActivePlayerIdx]
		obs := gs.BuildObservation()
		action := agent.ChooseAction(obs)

		result, err := gs.ExecuteTurn(action)
		if err != nil {
			t.Logf("  Error: %v", err)
			break
		}

		player := gs.Players[(gs.ActivePlayerIdx+len(gs.Players)-1)%len(gs.Players)]
		if result.Accepted {
			if result.Abstained {
				t.Logf("  Turn %d: %s ABSTAINS", gs.CurrentTurn, player.ID)
			} else {
				t.Logf("  Turn %d: %s %s via %s → args=%v score=%.2f",
					gs.CurrentTurn, player.ID, action.Type, action.SchemeID,
					result.ArgIDs, result.ScoreDelta)
			}
		} else {
			t.Logf("  Turn %d: %s REJECTED: %s", gs.CurrentTurn, player.ID, result.Error)
		}

		if result.Terminated {
			t.Logf("  GAME OVER: %s", result.TerminationReason)
		}
	}

	t.Logf("Final scores:")
	for _, p := range gs.Players {
		t.Logf("  %s: %.2f", p.ID, p.Score)
	}
	t.Logf("Graph: %d stmts, %d args",
		len(gs.Adapter.Graph.Statements), len(gs.Adapter.Graph.Arguments))

	// Verify the acyclic invariant held throughout.
	if !gs.Adapter.IsAcyclic() {
		t.Error("graph should be acyclic after random play")
	}
}

// needed for import
var _ = fmt.Sprintf
