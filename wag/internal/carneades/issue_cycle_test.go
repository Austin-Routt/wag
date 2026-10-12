package carneades

import (
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// Regression test for review defect #6: IsAcyclic followed premises and
// undercutters but not issues, so a cycle through a rebuttal passed the
// check, and Carneades then left its statements undecided.
//
//	p <- [base]          q <- [p]          ¬p <- [q]   (issue {p, ¬p})
//
// p's label depends on ¬p's argument, which depends on q, which depends on p.
func TestCycleThroughAnIssueIsDetected(t *testing.T) {
	a := NewAdapter()
	for _, id := range []string{"base", "p", "q"} {
		a.AddStatement(id, id)
	}
	a.AddAssumption("base")
	if _, err := a.AddArgumentDirect("Alice", "s", []string{"base"}, "p", 0.6); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddArgumentDirect("Alice", "s", []string{"p"}, "q", 0.6); err != nil {
		t.Fatal(err)
	}
	if !a.IsAcyclic() {
		t.Fatal("no cycle yet")
	}
	if _, err := a.AddRebuttal("Bob", "s", "p", []string{"q"}, 0.7, caes.PE); err != nil {
		t.Fatal(err)
	}
	a.Evaluate()
	if a.Graph.Statements["p"].Label != caes.Undecided || a.Graph.Statements["q"].Label != caes.Undecided {
		t.Fatalf("expected the cycle to leave p and q undecided, got %v and %v",
			a.Graph.Statements["p"].Label, a.Graph.Statements["q"].Label)
	}
	if a.IsAcyclic() {
		t.Fatal("the cycle through the issue {p, ¬p} was not detected")
	}
}

// Sharing an issue alone is not a cycle: a rebuttal from independent premises.
func TestAnIssueAloneIsNotACycle(t *testing.T) {
	a := NewAdapter()
	for _, id := range []string{"b1", "b2", "p"} {
		a.AddStatement(id, id)
	}
	a.AddAssumption("b1")
	a.AddAssumption("b2")
	if _, err := a.AddArgumentDirect("Alice", "s", []string{"b1"}, "p", 0.6); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddRebuttal("Bob", "s", "p", []string{"b2"}, 0.7, caes.PE); err != nil {
		t.Fatal(err)
	}
	if !a.IsAcyclic() {
		t.Fatal("a rebuttal from independent premises was reported as a cycle")
	}
	a.Evaluate()
	if a.Graph.Statements["p"].Label != caes.Out || a.Graph.Statements["¬p"].Label != caes.In {
		t.Fatalf("p %v, ¬p %v; want Out and In", a.Graph.Statements["p"].Label, a.Graph.Statements["¬p"].Label)
	}
}
