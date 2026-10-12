package game

// Regression tests for defects #3 and #4 of the 2026-10-06 review: games with
// the same seeds must be the same game, and AddSupport must return its main
// argument first. Both failed on the snapshot of 2026-10-06 because the order
// of Go map iteration leaked into the move enumeration, the random bindings
// and the list of created arguments.

import (
	"testing"

	"github.com/yourorg/wag/internal/nascq"
)

func playSeeded(t *testing.T, moves int) (*GameState, string) {
	t.Helper()
	gs := makeCoreGame(t)
	agents := []Agent{NewRandomAgent("Alice", 7, 0.0, 0.5), NewRandomAgent("Bob", 11, 0.0, 0.5)}
	for i := 0; i < moves && !gs.Terminated; i++ {
		a := agents[gs.ActivePlayerIdx]
		if _, err := gs.ExecuteTurn(a.ChooseAction(gs.BuildObservation())); err != nil {
			t.Fatalf("turn %d: %v", i, err)
		}
	}
	return gs, gs.MoveDigest()
}

func TestSameSeedsSameGame(t *testing.T) {
	_, first := playSeeded(t, 12)
	for run := 1; run < 30; run++ {
		if _, d := playSeeded(t, 12); d != first {
			t.Fatalf("run %d: move history differs from run 0 with the same seeds", run)
		}
	}
}

func TestRunGameRecordsProvenanceAndIsReproducible(t *testing.T) {
	cfg := DefaultGameConfig()
	cfg.MaxTurns = 6
	run := func() *GameResult {
		return RunGame(cfg, DefaultRunnerConfig(), []byte(coreYAML), []byte(coreSchemesJSON), []byte(coreCQsJSON),
			[]byte(coreEdgesJSON), []Agent{NewRandomAgent("Alice", 3, 0.1, 0.5), NewGreedyAgent("Bob", 5, 2)})
	}
	r0 := run()
	if r0.Panicked {
		t.Fatalf("panic: %s", r0.PanicMessage)
	}
	if r0.AgentSeeds["Alice"] != 3 || r0.AgentSeeds["Bob"] != 5 || len(r0.YAMLSHA256) != 64 || len(r0.NASCQSHA256) != 64 || r0.MoveDigest == "" {
		t.Fatalf("provenance not recorded: seeds=%v yaml=%q nascq=%q digest=%q", r0.AgentSeeds, r0.YAMLSHA256, r0.NASCQSHA256, r0.MoveDigest)
	}
	for i := 0; i < 20; i++ {
		if r := run(); r.MoveDigest != r0.MoveDigest || r.Scores["Alice"] != r0.Scores["Alice"] || r.Scores["Bob"] != r0.Scores["Bob"] {
			t.Fatalf("run %d differs: digest %s vs %s", i, r.MoveDigest, r0.MoveDigest)
		}
	}
}

func TestEnumeratedActionsHaveAStableOrder(t *testing.T) {
	gs := makeCoreGame(t)
	if r, _ := gs.ExecuteTurn(prSupportAction()); !r.Accepted {
		t.Fatalf("support rejected: %s", r.Error)
	}
	sig := func() string {
		s := ""
		for _, a := range gs.BuildObservation().AvailableActions {
			s += a.Type + a.SchemeID + a.AnswersCQID + a.TargetArgID + "|"
		}
		return s
	}
	first := sig()
	for i := 0; i < 50; i++ {
		if sig() != first {
			t.Fatal("the enumeration order of available actions changed between two calls")
		}
	}
}

func prSupportAction() nascq.ProposedAction {
	return nascq.ProposedAction{Type: "support", SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{"A": "carbon_tax", "S1": "current_state", "S2": "lower_emissions",
			"G": "reduce_emissions", "V": "environmental_health"}}
}

func TestAddSupportReturnsTheMainArgumentFirst(t *testing.T) {
	for i := 0; i < 30; i++ {
		gs := makeCoreGame(t)
		r, _ := gs.ExecuteTurn(prSupportAction())
		if !r.Accepted || len(r.ArgIDs) < 2 {
			t.Fatalf("support: %+v", r)
		}
		first := gs.Adapter.Graph.Arguments[r.ArgIDs[0]]
		if first.Conclusion == nil || first.Conclusion.IsUndercutter {
			t.Fatalf("run %d: ArgIDs[0]=%s concludes %v, not the scheme's conclusion", i, r.ArgIDs[0], first.Conclusion)
		}
		for _, id := range r.ArgIDs[1:] {
			if c := gs.Adapter.Graph.Arguments[id].Conclusion; c == nil || !c.IsUndercutter {
				t.Fatalf("run %d: %s after the main argument is not an exception argument", i, id)
			}
		}
	}
}
