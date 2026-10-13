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
	"github.com/yourorg/wag/data"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// =========================================================================
// The shipped data (package data embeds the files under data/)
// =========================================================================

var (
	coreYAML        = string(data.CoreYAML)
	coreSchemesJSON = string(data.SchemesJSON)
	coreCQsJSON     = string(data.CQsJSON)
	coreEdgesJSON   = string(data.EdgesJSON)
)

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
	if stats.QuestionCount != 20 {
		t.Errorf("NASCQ CQ count = %d, want 20", stats.QuestionCount)
	}
	for id := range gs.NASCQ.Schemes {
		if len(gs.NASCQ.GetAskCQs(id)) == 0 {
			t.Errorf("no critical question is asked of %s", id)
		}
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
// Test 4: The critical questions of practical reasoning, by locus
// =========================================================================

// "Are there goals other than G that should have priority?" is about the
// conclusion: answered with an argument from values, it rebuts. "Are there
// alternative actions?" is about the exception realize_more_effectively(G,A):
// answered with an argument from consequences, it undercuts.
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
	target := gs.Adapter.Graph.Arguments[targetArgID]

	// Turn 2: Bob rebuts through CQ:other_goals with AS:values.
	r2, _ := gs.ExecuteTurn(nascq.ProposedAction{
		Type:        "attack",
		AttackType:  "rebuttal",
		SchemeID:    "AS:values",
		AnswersCQID: "CQ:other_goals",
		TargetArgID: targetArgID,
		VariableBindings: map[string]string{
			"V": "economic_freedom", "A": "carbon_tax", "G": "low_prices",
		},
	})
	if !r2.Accepted {
		t.Fatalf("Turn 2 rejected: %s", r2.Error)
	}
	rebuttal := gs.Adapter.Graph.Arguments[r2.ArgIDs[0]]
	if want := "¬" + target.Conclusion.Id; rebuttal.Conclusion.Id != want {
		t.Fatalf("rebuttal concludes %s, want %s", rebuttal.Conclusion.Id, want)
	}
	if target.Conclusion.Issue == nil || target.Conclusion.Issue != rebuttal.Conclusion.Issue {
		t.Fatal("the rebuttal did not make the conclusion an issue")
	}
	// Under PE Alice's 0.70 beats Bob's 0.60.
	if target.Conclusion.Label != caes.In || rebuttal.Conclusion.Label != caes.Out {
		t.Fatalf("conclusion %v, rebuttal %v; want In and Out (0.70 > 0.60 under PE)",
			target.Conclusion.Label, rebuttal.Conclusion.Label)
	}

	// Turn 3: Alice abstains. Turn 4: Bob undercuts through CQ:other_means
	// with AS:consequences: its answer is the exception, which holds.
	gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	r4, _ := gs.ExecuteTurn(nascq.ProposedAction{
		Type:        "attack",
		SchemeID:    "AS:consequences",
		AnswersCQID: "CQ:other_means",
		TargetArgID: targetArgID,
		VariableBindings: map[string]string{
			"A": "solar_subsidies", "G": "green_energy_adoption",
		},
	})
	if !r4.Accepted {
		t.Fatalf("Turn 4 rejected: %s", r4.Error)
	}
	if got := gs.MoveHistory[len(gs.MoveHistory)-1].Action.AttackType; got != "undercutter" {
		t.Fatalf("CQ:other_means recorded as %q, want undercutter", got)
	}
	exc := gs.Adapter.Graph.Arguments[r4.ArgIDs[0]].Conclusion
	if exc.Id != "realize_more_effectively(reduce_emissions,carbon_tax)" || exc.Label != caes.In {
		t.Fatalf("other means answer concludes %s (%v), want the exception, In", exc.Id, exc.Label)
	}
	if target.Weight != 0 || target.Conclusion.Label != caes.Out {
		t.Fatalf("undercut argument: weight %v, conclusion %v; want 0 and Out", target.Weight, target.Conclusion.Label)
	}
	if errs := gs.CheckInvariants(); len(errs) > 0 {
		t.Fatal(errs)
	}
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
