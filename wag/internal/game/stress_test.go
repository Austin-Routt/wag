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
	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
	"testing"
)

// =========================================================================
// The shipped data (package data embeds the files under data/). These used
// to be a separate, drifted copy whose possibility question had no
// target_premise_pattern.
// =========================================================================

var (
	stressYAML        = string(data.CoreYAML)
	stressSchemesJSON = string(data.SchemesJSON)
	stressCQsJSON     = string(data.CQsJSON)
	stressEdgesJSON   = string(data.EdgesJSON)
)

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
