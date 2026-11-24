package carneades

import (
	caes "github.com/carneades/carneades-4/src/engine/caes"
	"testing"
)

// A rejected moderator request must not change the shared assumptions.
func TestDecideInvalidDecisionIsAtomic(t *testing.T) {
	for _, initial := range []string{DecisionAccepted, DecisionRejected} {
		t.Run(initial, func(t *testing.T) {
			a := NewAdapter()
			a.AddStatement("p", "P")
			if _, err := a.Decide("p", initial); err != nil {
				t.Fatal(err)
			}
			if _, err := a.Evaluate(); err != nil {
				t.Fatal(err)
			}
			before := a.Fingerprint()
			if _, err := a.Decide("p", "acceptd"); err == nil {
				t.Fatal("invalid decision accepted")
			}
			if got := a.Fingerprint(); got != before {
				t.Fatalf("rejected request mutated graph\nbefore:\n%s\nafter:\n%s", before, got)
			}
			if got := a.Decision("p"); got != initial {
				t.Fatalf("decision changed from %s to %s", initial, got)
			}
		})
	}
}

// Binary rejection has no unambiguous opposite in an issue with three positions.
func TestDecideUnrepresentableRejectionIsAtomic(t *testing.T) {
	a := NewAdapter()
	p := a.AddStatement("p", "P")
	q := a.AddStatement("q", "Q")
	r := a.AddStatement("r", "R")
	issue := &caes.Issue{Id: "three", Positions: []*caes.Statement{p, q, r}, Standard: caes.PE}
	a.Graph.Issues[issue.Id] = issue
	for _, s := range issue.Positions {
		s.Issue = issue
	}
	a.AddAssumption("p")
	before := a.Fingerprint()
	if _, err := a.Decide("p", DecisionRejected); err == nil {
		t.Fatal("ambiguous rejection accepted")
	}
	if got := a.Fingerprint(); got != before {
		t.Fatalf("failed rejection mutated graph\nbefore:\n%s\nafter:\n%s", before, got)
	}
}
