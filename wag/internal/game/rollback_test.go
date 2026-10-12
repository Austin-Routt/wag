package game

// Regression test for review defect #5 in the turn pipeline: a move rejected
// as cyclic must leave the graph exactly as it was.

import (
	"testing"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
)

func TestRejectedCyclicMoveLeavesNoTrace(t *testing.T) {
	gs := makeCoreGame(t)
	r, _ := gs.ExecuteTurn(prSupportAction())
	main := r.ArgIDs[0]
	concl := gs.Adapter.Graph.Arguments[main].Conclusion.Id

	attack := nascq.ProposedAction{Type: "attack", AttackType: "undercutter", SchemeID: "AS:sign",
		AnswersCQID: "CQ:side_effect", TargetArgID: main,
		VariableBindings: map[string]string{"A": "rising_unemployment", "B": "economic_harm"}}
	va, err := gs.NASCQ.Validate(attack, gs.bridge)
	if err != nil {
		t.Fatal(err)
	}
	// Make the attack's first premise depend on the attacked conclusion, so the
	// undercutter closes a cycle: conclusion <- main <- undercutter <- attack <- premise <- conclusion.
	p0 := carneades.TextToAtom(va.InstPremises[0])
	gs.Adapter.AddStatement(p0, va.InstPremises[0])
	if _, err := gs.Adapter.AddArgumentDirect("setup", "s", []string{concl}, p0, 0.5); err != nil {
		t.Fatal(err)
	}
	gs.Adapter.Evaluate()
	before := gs.Adapter.Fingerprint()

	r2, err := gs.ExecuteTurn(attack)
	if err != nil {
		t.Fatal(err)
	}
	if r2.Accepted || r2.Error != "move creates a cycle in the argument graph" {
		t.Fatalf("cyclic move not rejected: %+v", r2)
	}
	if after := gs.Adapter.Fingerprint(); after != before {
		t.Fatalf("the rejected move left residue.\nbefore:\n%s\nafter:\n%s", before, after)
	}
}
