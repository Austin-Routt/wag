package game

import (
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/nascq"
)

// These tests use stressYAML, stressSchemesJSON, stressCQsJSON, stressEdgesJSON
// from stress_test.go (same package, accessible in test binary).

func makeTestEnv(t *testing.T) *Env {
	t.Helper()
	return NewEnv(EnvConfig{
		Game:        GameConfig{MaxTurns: 6, AbstentionThreshold: 2, Scoring: DefaultScoringConfig()},
		YAMLBytes:   []byte(stressYAML),
		SchemesJSON: []byte(stressSchemesJSON),
		CQsJSON:     []byte(stressCQsJSON),
		EdgesJSON:   []byte(stressEdgesJSON),
		PlayerNames: []string{"Alice", "Bob"},
	})
}

func TestEnvResetCreatesGame(t *testing.T) {
	env := makeTestEnv(t)
	obs, err := env.Reset()
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
	if obs == nil {
		t.Fatal("observation should not be nil")
	}
	if obs.ActivePlayer.ID != "Alice" {
		t.Errorf("first player = %q, want Alice", obs.ActivePlayer.ID)
	}
	if obs.TurnNumber != 1 {
		t.Errorf("turn = %d, want 1", obs.TurnNumber)
	}
	if len(obs.AvailableActions) == 0 {
		t.Error("should have available actions after reset")
	}
	if env.GameState() == nil {
		t.Error("GameState() should not be nil after Reset")
	}
}

func TestEnvStepSupport(t *testing.T) {
	env := makeTestEnv(t)
	env.Reset()

	action := nascq.ProposedAction{
		Type:     "support",
		SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{
			"A": "carbon_tax", "S1": "now", "S2": "future",
			"G": "reduce_emissions", "V": "sustainability",
		},
	}

	result, err := env.Step(action)
	if err != nil {
		t.Fatalf("Step failed: %v", err)
	}

	if result.Reward <= 0 {
		t.Errorf("reward = %.2f, want > 0 for accepted support", result.Reward)
	}
	if result.Done {
		t.Error("game should not be done after one move")
	}
	if result.Info["accepted"] != true {
		t.Error("info should indicate accepted")
	}
	if result.Info["player"] != "Alice" {
		t.Errorf("info player = %v, want Alice", result.Info["player"])
	}
	if result.Observation == nil {
		t.Fatal("observation should not be nil")
	}
	// After Alice's move, it should be Bob's turn.
	if result.Observation.ActivePlayer.ID != "Bob" {
		t.Errorf("next player = %q, want Bob", result.Observation.ActivePlayer.ID)
	}
}

func TestEnvStepAbstain(t *testing.T) {
	env := makeTestEnv(t)
	env.Reset()

	result, err := env.Step(nascq.ProposedAction{Type: "abstain"})
	if err != nil {
		t.Fatalf("Step failed: %v", err)
	}
	if result.Reward != 0 {
		t.Errorf("reward = %.2f, want 0 for abstain", result.Reward)
	}
	if result.Info["abstained"] != true {
		t.Error("info should indicate abstained")
	}
}

func TestEnvStepTermination(t *testing.T) {
	env := NewEnv(EnvConfig{
		Game:        GameConfig{MaxTurns: 1, AbstentionThreshold: 100, Scoring: DefaultScoringConfig()},
		YAMLBytes:   []byte(stressYAML),
		SchemesJSON: []byte(stressSchemesJSON),
		CQsJSON:     []byte(stressCQsJSON),
		EdgesJSON:   []byte(stressEdgesJSON),
		PlayerNames: []string{"Alice", "Bob"},
	})
	env.Reset()

	// Two abstains = 1 round = max turns reached.
	env.Step(nascq.ProposedAction{Type: "abstain"})
	result, _ := env.Step(nascq.ProposedAction{Type: "abstain"})

	if !result.Done {
		t.Error("game should be done after max turns")
	}
	if !result.Truncated {
		t.Error("truncated should be true for max-turns termination")
	}
	reason, _ := result.Info["termination_reason"].(string)
	if reason == "" {
		t.Error("termination reason should be set")
	}
	t.Logf("Termination: %s", reason)
}

func TestEnvStepBeforeReset(t *testing.T) {
	env := NewEnv(EnvConfig{})
	_, err := env.Step(nascq.ProposedAction{Type: "abstain"})
	if err == nil {
		t.Error("Step before Reset should return error")
	}
}

func TestEnvResetMultiple(t *testing.T) {
	env := makeTestEnv(t)

	// First game.
	obs1, _ := env.Reset()
	env.Step(nascq.ProposedAction{
		Type: "support", SchemeID: "AS:consequences",
		VariableBindings: map[string]string{"A": "action1", "G": "goal1"},
	})

	// Second game — should be fresh.
	obs2, _ := env.Reset()
	gs2 := env.GameState()

	if len(gs2.Adapter.Graph.Arguments) != 0 {
		t.Error("second game should have empty argument graph")
	}
	if obs2.TurnNumber != 1 {
		t.Error("second game should start at turn 1")
	}
	_ = obs1 // used
}

func TestEnvRender(t *testing.T) {
	env := makeTestEnv(t)
	env.Reset()

	output := env.Render()
	if output == "" {
		t.Error("Render should produce non-empty output")
	}
	if !strings.Contains(output, "Alice") {
		t.Error("Render should mention player names")
	}
	if !strings.Contains(output, "Round 1") {
		t.Error("Render should mention the round")
	}
	t.Logf("Render output:\n%s", output)
}

func TestEnvRenderBeforeReset(t *testing.T) {
	env := NewEnv(EnvConfig{})
	output := env.Render()
	if !strings.Contains(output, "no game") {
		t.Errorf("Render before Reset should say no game, got: %q", output)
	}
}

func TestRenderStateShowsArguments(t *testing.T) {
	env := makeTestEnv(t)
	env.Reset()

	// Add an argument.
	env.Step(nascq.ProposedAction{
		Type: "support", SchemeID: "AS:consequences",
		VariableBindings: map[string]string{"A": "test_action", "G": "test_goal"},
	})

	output := env.Render()
	if !strings.Contains(output, "Arguments in play") {
		t.Error("should show arguments section")
	}
	// Should show at least one argument entry.
	if !strings.Contains(output, "[a") {
		t.Error("should show argument ID like [a1]")
	}
	t.Logf("After 1 move:\n%s", output)
}
