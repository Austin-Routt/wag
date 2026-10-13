package carneades

import (
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// A rebuttal of a rebuttal argues for the rebutted statement again: no ¬¬p,
// one issue [p, ¬p], and the strongest position wins, in every evaluation
// (Carneades walks a Go map; with ¬p a position of two issues the outcome
// changed from run to run).
func TestRebuttingARebuttalArguesForTheOriginalPosition(t *testing.T) {
	for run := 0; run < 30; run++ {
		a := NewAdapter()
		for _, id := range []string{"p", "base", "q", "r"} {
			a.AddStatement(id, id)
		}
		for _, id := range []string{"base", "q", "r"} {
			a.AddAssumption(id)
		}
		if _, err := a.AddArgumentDirect("A", "s", []string{"base"}, "p", 0.5); err != nil {
			t.Fatal(err)
		}
		reb, err := a.AddRebuttal("B", "s", "p", []string{"q"}, 0.6, caes.PE)
		if err != nil {
			t.Fatal(err)
		}
		if got := a.Graph.Arguments[reb[0]].Conclusion.Id; got != "¬p" {
			t.Fatalf("rebuttal concludes %s, want ¬p", got)
		}
		counter, err := a.AddRebuttal("A", "s", "¬p", []string{"r"}, 0.9, caes.PE)
		if err != nil {
			t.Fatal(err)
		}
		if got := a.Graph.Arguments[counter[0]].Conclusion.Id; got != "p" {
			t.Fatalf("counter-rebuttal concludes %s, want p", got)
		}
		if _, ok := a.Graph.Statements["¬¬p"]; ok || len(a.Graph.Issues) != 1 {
			t.Fatalf("want one issue and no ¬¬p; issues=%d", len(a.Graph.Issues))
		}
		a.Evaluate()
		if p, n := a.Graph.Statements["p"].Label, a.Graph.Statements["¬p"].Label; p != caes.In || n != caes.Out {
			t.Fatalf("run %d: p=%v ¬p=%v, want In and Out (0.9 > 0.6)", run, p, n)
		}
	}
}

// Undermining an undermined premise's negation reinstates the premise the
// same way, and a target that is one of several positions of an issue is
// refused rather than given a second issue.
func TestOpposingPositionRules(t *testing.T) {
	a := NewAdapter()
	for _, id := range []string{"p", "x", "y", "base"} {
		a.AddStatement(id, id)
	}
	a.AddAssumption("base")
	if _, err := a.AddUndermining("B", "s", "p", []string{"base"}, 0.6, caes.PE); err != nil {
		t.Fatal(err)
	}
	back, err := a.AddUndermining("A", "s", "¬p", []string{"base"}, 0.7, caes.PE)
	if err != nil {
		t.Fatal(err)
	}
	if got := a.Graph.Arguments[back[0]].Conclusion.Id; got != "p" {
		t.Fatalf("undermining ¬p concludes %s, want p", got)
	}
	if neg := a.Graph.Statements["¬p"]; neg.Text != "It is not the case that: p" {
		t.Fatalf("negation text %q", neg.Text)
	}

	three := &caes.Issue{Id: "i3", Positions: []*caes.Statement{a.Graph.Statements["x"], a.Graph.Statements["y"], a.Graph.Statements["base"]}}
	a.Graph.Issues["i3"] = three
	for _, s := range three.Positions {
		s.Issue = three
	}
	if _, err := a.AddRebuttal("B", "s", "x", []string{"base"}, 0.6, caes.PE); err == nil {
		t.Fatal("rebutting one of three positions of an issue was accepted")
	}
}
