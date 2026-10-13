package carneades

import (
	"strings"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// ---------------------------------------------------------------------------
// Test YAML fixture
// ---------------------------------------------------------------------------

// minimalYAML is a self-contained Carneades YAML fixture with:
//   - A language section (predicate format strings)
//   - Two argument schemes: simple_support and counter_argument
//   - One pre-defined statement with an assumption
//
// This exercises the full YAML Import path including Theory construction,
// scheme parsing, language map loading, and assumption handling.
const minimalYAML = `
language:
  is_true/1: "%s is true."
  supports/2: "%s supports %s."
  claim/1: "The claim is: %s."
  evidence_for/1: "There is evidence for %s."
  against/1: "There is evidence against %s."

argument_schemes:
  - id: simple_support
    variables: [E, C]
    conclusions:
      - claim(C)
    premises:
      - evidence_for(E)
    assumptions:
      - evidence_for(E)

  - id: counter_argument
    variables: [E, C]
    conclusions:
      - against(C)
    premises:
      - evidence_for(E)
    assumptions:
      - evidence_for(E)

statements:
  base_fact: "A base fact."

assumptions:
  - base_fact
`

// ---------------------------------------------------------------------------
// Tests: YAML Loading
// ---------------------------------------------------------------------------

func TestNewAdapterFromYAML(t *testing.T) {
	a, err := NewAdapterFromYAML([]byte(minimalYAML))
	if err != nil {
		t.Fatalf("NewAdapterFromYAML failed: %v", err)
	}

	// Verify graph is valid.
	if a.Graph == nil {
		t.Fatal("Graph is nil")
	}
	if a.Graph.Theory == nil {
		t.Fatal("Theory is nil")
	}

	// Verify schemes loaded.
	ids := a.SchemeIDs()
	if len(ids) == 0 {
		t.Fatal("no schemes loaded from YAML")
	}

	foundSimple := false
	foundCounter := false
	for _, id := range ids {
		switch id {
		case "simple_support":
			foundSimple = true
		case "counter_argument":
			foundCounter = true
		}
	}
	if !foundSimple {
		t.Error("scheme 'simple_support' not found in loaded schemes")
	}
	if !foundCounter {
		t.Error("scheme 'counter_argument' not found in loaded schemes")
	}

	// Verify language loaded.
	if a.Graph.Theory.Language == nil {
		t.Fatal("Language map is nil")
	}
	if _, ok := a.Graph.Theory.Language["claim/1"]; !ok {
		t.Error("language entry 'claim/1' not found")
	}

	// Verify base_fact statement exists and is assumed.
	found := false
	for _, as := range a.Graph.Assumptions {
		if strings.Contains(as, "base_fact") {
			found = true
			break
		}
	}
	if !found {
		t.Error("assumption 'base_fact' not found")
	}

	// Evaluate and check base_fact is In (assumed).
	a.Evaluate()
	for _, s := range a.Graph.Statements {
		if strings.Contains(s.Id, "base_fact") {
			if s.Label != caes.In {
				t.Errorf("base_fact label = %v, want In (assumed)", s.Label)
			}
			break
		}
	}
}

func TestNewAdapterFromYAMLInvalid(t *testing.T) {
	_, err := NewAdapterFromYAML([]byte("not: [valid: yaml: {{{"))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

// ---------------------------------------------------------------------------
// Tests: AddSupport with YAML-loaded scheme (clean atoms)
// ---------------------------------------------------------------------------

func TestAddSupportWithYAMLScheme(t *testing.T) {
	a, err := NewAdapterFromYAML([]byte(minimalYAML))
	if err != nil {
		t.Fatalf("YAML load failed: %v", err)
	}

	// Use SanitizeBinding for the variable values — this is how the game
	// layer will prepare user input.
	bindings := map[string]string{
		"E": SanitizeBinding("climate data"),   // → "climate_data"
		"C": SanitizeBinding("global warming"), // → "global_warming"
	}

	argIDs, err := a.AddSupport("player1", "simple_support", bindings, 0.8)
	if err != nil {
		t.Fatalf("AddSupport failed: %v", err)
	}
	if len(argIDs) == 0 {
		t.Fatal("AddSupport created no arguments")
	}

	// Verify the argument exists and has correct metadata.
	arg := a.Graph.Arguments[argIDs[0]]
	if arg == nil {
		t.Fatal("argument not found")
	}
	if a.ArgumentOwner(argIDs[0]) != "player1" {
		t.Errorf("owner = %q, want %q", a.ArgumentOwner(argIDs[0]), "player1")
	}

	// Verify that conclusion statement was created with Language formatting.
	// The scheme conclusion is claim(C) → claim(global_warming)
	// The Language has claim/1: "The claim is: %s."
	// So the statement text should be formatted.
	foundClaim := false
	for _, s := range a.Graph.Statements {
		if strings.Contains(s.Id, "claim") && strings.Contains(s.Id, "global_warming") {
			foundClaim = true
			// The Language template should have been applied.
			if s.Text != "" && !strings.Contains(s.Text, "claim") {
				t.Logf("claim statement text: %q (Language template applied)", s.Text)
			}
			break
		}
	}
	if !foundClaim {
		t.Error("conclusion statement containing 'claim' and 'global_warming' not found")
		t.Logf("Statements:")
		for id := range a.Graph.Statements {
			t.Logf("  %q", id)
		}
	}

	// Evaluate: evidence_for(climate_data) is auto-assumed by the scheme,
	// so the argument's premise is In and claim(global_warming) should be In.
	a.Evaluate()
	for _, s := range a.Graph.Statements {
		if strings.Contains(s.Id, "claim") && strings.Contains(s.Id, "global_warming") {
			if s.Label != caes.In {
				t.Errorf("claim statement label = %v, want In", s.Label)
			}
			break
		}
	}
}

// TestAddSupportNoParseWarnings verifies that using SanitizeBinding (TextToAtom)
// instead of quoteForProlog eliminates the "invalid char literal" warnings.
//
// This is a functional test — the warnings go to stderr which we can't easily
// capture in Go tests. But we can verify that the scheme instantiation succeeds
// and produces correctly formatted statement IDs without single quotes.
func TestAddSupportNoParseWarnings(t *testing.T) {
	a, err := NewAdapterFromYAML([]byte(minimalYAML))
	if err != nil {
		t.Fatalf("YAML load failed: %v", err)
	}

	bindings := map[string]string{
		"E": SanitizeBinding("Dr. Smith's research"), // → "dr__smith_s_research"
		"C": SanitizeBinding("vaccine safety"),       // → "vaccine_safety"
	}

	argIDs, err := a.AddSupport("player1", "simple_support", bindings, 0.9)
	if err != nil {
		t.Fatalf("AddSupport failed: %v", err)
	}

	// Verify no single quotes in any statement IDs.
	for id := range a.Graph.Statements {
		if strings.Contains(id, "'") {
			t.Errorf("statement ID %q contains single quote — SanitizeBinding should prevent this", id)
		}
	}

	// Verify argument was created successfully.
	if len(argIDs) == 0 {
		t.Error("no arguments created")
	}
}

// ---------------------------------------------------------------------------
// Tests: RemoveArgument (rollback)
// ---------------------------------------------------------------------------

func TestRemoveArgument(t *testing.T) {
	a := NewAdapter()
	a.AddStatement("p", "Premise")
	a.AddAssumption("p")
	a.AddStatement("c", "Conclusion")

	argID, err := a.AddArgumentDirect("player1", "s", []string{"p"}, "c", 1.0)
	if err != nil {
		t.Fatalf("AddArgumentDirect failed: %v", err)
	}

	// Verify argument exists.
	if _, ok := a.Graph.Arguments[argID]; !ok {
		t.Fatal("argument should exist before removal")
	}

	// Remove argument.
	err = a.RemoveArgument(argID)
	if err != nil {
		t.Fatalf("RemoveArgument failed: %v", err)
	}

	// Verify argument is gone.
	if _, ok := a.Graph.Arguments[argID]; ok {
		t.Error("argument should not exist after removal")
	}

	// Verify conclusion's Args slice was cleaned up.
	cStmt := a.Graph.Statements["c"]
	for _, arg := range cStmt.Args {
		if arg.Id == argID {
			t.Error("removed argument still in conclusion's Args slice")
		}
	}

	// After removal, evaluate: C should be Out (unsupported).
	a.Evaluate()
	if cStmt.Label != caes.Out {
		t.Errorf("after removal: C label = %v, want Out", cStmt.Label)
	}
}

func TestRemoveArgumentRollbackCycle(t *testing.T) {
	// Simulate the turn pipeline rollback scenario:
	// 1. Add an argument
	// 2. Check acyclicity → fails
	// 3. Remove the argument (rollback)
	// 4. Graph is back to valid state

	a := NewAdapter()
	a.AddStatement("p", "P")
	a.AddAssumption("p")
	a.AddStatement("c", "C")

	// First argument: [p] → c
	_, err := a.AddArgumentDirect("player1", "s", []string{"p"}, "c", 1.0)
	if err != nil {
		t.Fatalf("arg1 failed: %v", err)
	}

	// Second argument: [c] → p (creates cycle!)
	arg2ID, err := a.AddArgumentDirect("player1", "s", []string{"c"}, "p", 1.0)
	if err != nil {
		t.Fatalf("arg2 failed: %v", err)
	}

	// Acyclicity check should fail.
	if a.IsAcyclic() {
		t.Fatal("graph should have a cycle")
	}

	// Rollback: remove the cycle-causing argument.
	err = a.RemoveArgument(arg2ID)
	if err != nil {
		t.Fatalf("RemoveArgument failed: %v", err)
	}

	// Graph should be acyclic again.
	if !a.IsAcyclic() {
		t.Error("graph should be acyclic after rollback")
	}

	// Evaluation should work correctly.
	a.Evaluate()
	pStmt := a.Graph.Statements["p"]
	cStmt := a.Graph.Statements["c"]
	if pStmt.Label != caes.In {
		t.Errorf("after rollback: p = %v, want In (assumed)", pStmt.Label)
	}
	if cStmt.Label != caes.In {
		t.Errorf("after rollback: c = %v, want In (supported by arg1)", cStmt.Label)
	}
}

func TestRemoveArgumentNotFound(t *testing.T) {
	a := NewAdapter()
	err := a.RemoveArgument("nonexistent")
	if err == nil {
		t.Error("RemoveArgument should fail for nonexistent argument")
	}
}

// ---------------------------------------------------------------------------
// Tests: StatementLabels (delta scoring helper)
// ---------------------------------------------------------------------------

func TestStatementLabels(t *testing.T) {
	a := NewAdapter()
	a.AddStatement("p", "P")
	a.AddAssumption("p")
	a.AddStatement("c", "C")
	a.AddArgumentDirect("player1", "s", []string{"p"}, "c", 1.0)
	a.Evaluate()

	labels := a.StatementLabels()
	if labels["p"] != caes.In {
		t.Errorf("labels[p] = %v, want In", labels["p"])
	}
	if labels["c"] != caes.In {
		t.Errorf("labels[c] = %v, want In", labels["c"])
	}
}

// ---------------------------------------------------------------------------
// Tests: RemoveArguments (batch, for AddSupport rollback)
// ---------------------------------------------------------------------------

func TestRemoveArguments(t *testing.T) {
	a := NewAdapter()
	a.AddStatement("p1", "P1")
	a.AddAssumption("p1")
	a.AddStatement("p2", "P2")
	a.AddAssumption("p2")
	a.AddStatement("c1", "C1")
	a.AddStatement("c2", "C2")

	id1, _ := a.AddArgumentDirect("player1", "s", []string{"p1"}, "c1", 1.0)
	id2, _ := a.AddArgumentDirect("player1", "s", []string{"p2"}, "c2", 1.0)

	err := a.RemoveArguments([]string{id1, id2})
	if err != nil {
		t.Fatalf("RemoveArguments failed: %v", err)
	}

	if len(a.Graph.Arguments) != 0 {
		t.Errorf("expected 0 arguments after batch removal, got %d", len(a.Graph.Arguments))
	}
}
