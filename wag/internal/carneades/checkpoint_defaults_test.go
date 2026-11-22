package carneades

import (
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// These tests need the schemes library's engine (4.3.1+walton.N), whose graph
// records which assumptions a scheme added (ArgGraph.Defaults) and which
// statements are questioned (ArgGraph.Questioned). Its contract 5 leaves the
// consistency of these fields to a caller that writes ArgGraph.Assumptions
// directly, as Restore, RetractAssumptionsSince and unassume do.

func adapterWithDefaultScheme() *Adapter {
	a := NewAdapter()
	a.Graph.Theory.ArgSchemes = append(a.Graph.Theory.ArgSchemes, &caes.Scheme{
		Id:          "test_default",
		Variables:   []string{"P", "C"},
		Premises:    []string{"premise(P)"},
		Assumptions: []string{"usual(P)"},
		Conclusions: []string{"conclusion(C)"},
	})
	a.Graph.Theory.InitSchemeIndex()
	a.buildSchemeMap()
	return a
}

// A rolled-back move leaves no default marker behind.
func TestRestoreForgetsTheDefaultsOfARolledBackMove(t *testing.T) {
	a := adapterWithDefaultScheme()
	cp := a.Checkpoint()
	if _, err := a.AddSupport("Bob", "test_default", map[string]string{"P": "x", "C": "y"}, 0.7); err != nil {
		t.Fatal(err)
	}
	if !a.Graph.Defaults["usual(x)"] {
		t.Fatalf("the engine did not mark the scheme's assumption as a default: %v", a.Graph.Defaults)
	}
	a.Restore(cp)
	if len(a.Graph.Defaults) != 0 || len(a.Graph.Questioned) != 0 {
		t.Fatalf("the rolled-back move left defaults %v, questioned %v", a.Graph.Defaults, a.Graph.Questioned)
	}
}

// The game retracts a scheme's assumptions and presumes them instead; the
// retraction takes their default markers with it.
func TestRetractAssumptionsSinceForgetsTheirDefaultMarkers(t *testing.T) {
	a := adapterWithDefaultScheme()
	n := a.AssumptionCount()
	if _, err := a.AddSupport("Bob", "test_default", map[string]string{"P": "x", "C": "y"}, 0.7); err != nil {
		t.Fatal(err)
	}
	if got := a.RetractAssumptionsSince(n); len(got) != 1 || got[0] != "usual(x)" {
		t.Fatalf("retracted %v, want [usual(x)]", got)
	}
	if len(a.Graph.Defaults) != 0 {
		t.Fatalf("the retracted assumption is still marked as a default: %v", a.Graph.Defaults)
	}
}

// A stipulated statement is still a stipulation after a rolled-back move that
// withdrew it and argued from a scheme assuming it by default. A stipulation
// does not yield to a stipulated contrary; a default does.
func TestRestoreKeepsAStipulationAStipulation(t *testing.T) {
	a := adapterWithDefaultScheme()
	a.AddStatement("usual(x)", "x is usual")
	a.AddAssumption("usual(x)")
	cp := a.Checkpoint()

	if _, err := a.Decide("usual(x)", DecisionUndecided); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddSupport("Bob", "test_default", map[string]string{"P": "x", "C": "y"}, 0.7); err != nil {
		t.Fatal(err)
	}
	a.Restore(cp)

	if a.Graph.Defaults["usual(x)"] {
		t.Errorf("after the rollback the stipulated usual(x) is marked as a scheme default")
	}
	a.AddStatement("¬usual(x)", "x is not usual")
	a.AddAssumption("¬usual(x)") // a second stipulation
	assumed := map[string]bool{}
	for _, s := range a.Graph.Assumed() {
		assumed[s] = true
	}
	if !assumed["usual(x)"] || !assumed["¬usual(x)"] {
		t.Errorf("two stipulations: assumed %v, want both (a stipulation does not yield)", a.Graph.Assumed())
	}
}
