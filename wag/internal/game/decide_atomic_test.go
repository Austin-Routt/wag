package game

import (
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// A refused moderator decision leaves no trace, whatever the adapter did
// before refusing it: rejecting a statement that has no issue creates its
// complement, and then refuses when the issue the two would form already
// exists with other positions.
func TestDecideRefusedLeavesNoTrace(t *testing.T) {
	gs := newShippedGame(t, prevEdges, "Alice", "Bob")
	gs.Config.Moderated = true
	a := gs.Adapter
	a.AddStatement("p", "P")
	q := a.AddStatement("q", "Q")
	r := a.AddStatement("r", "R")
	is := &caes.Issue{Id: "issue:p", Positions: []*caes.Statement{q, r}, Standard: caes.PE}
	a.Graph.Issues[is.Id] = is
	q.Issue, r.Issue = is, is
	before, history := a.Fingerprint(), len(gs.MoveHistory)
	if _, err := gs.Decide("p", "rejected", "no"); err == nil {
		t.Fatal("a rejection whose complement cannot form an issue was accepted")
	}
	if _, ok := a.Graph.Statements["¬p"]; ok {
		t.Error("the refused decision left its complement in the graph")
	}
	if got := a.Fingerprint(); got != before {
		t.Errorf("a refused decision changed the graph\nbefore:\n%s\nafter:\n%s", before, got)
	}
	if len(gs.MoveHistory) != history {
		t.Error("a refused decision was recorded as a move")
	}
	// A misspelled decision, as the server passes it on, is refused too.
	if _, err := gs.Decide("p", "acceptd", "no"); err == nil {
		t.Error("an unknown decision was accepted")
	}
	if got := a.Fingerprint(); got != before || len(gs.MoveHistory) != history {
		t.Error("an unknown decision changed the game")
	}
	// The same decision on a statement that can take a complement still works.
	if _, err := gs.Decide("q", "accepted", "yes"); err != nil || len(gs.MoveHistory) != history+1 {
		t.Fatalf("a valid decision after a refused one: %v", err)
	}
}
