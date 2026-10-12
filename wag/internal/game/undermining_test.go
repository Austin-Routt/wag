package game

// Regression tests for defect #1 of the 2026-10-06 review: Trusted mode
// assumed every premise of a move, and Carneades labels an assumption In
// whatever is argued against it, so an undermining attack could never
// succeed (possible(carbon_tax) stayed In after CQ:possibility, and the
// conclusion was unchanged).

import (
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

func possiblePremise(t *testing.T, gs *GameState, main string) string {
	t.Helper()
	for _, p := range gs.Adapter.Graph.Arguments[main].Premises {
		if strings.HasPrefix(p.Stmt.Id, "possible(") {
			return p.Stmt.Id
		}
	}
	t.Fatal("no possible(...) premise")
	return ""
}

func TestMovePremisesArePresumedNotAssumed(t *testing.T) {
	gs := makeCoreGame(t)
	before := gs.Adapter.AssumptionCount()
	r, _ := gs.ExecuteTurn(prSupportAction())
	if !r.Accepted {
		t.Fatalf("support rejected: %s", r.Error)
	}
	if n := gs.Adapter.AssumptionCount(); n != before {
		t.Fatalf("the move added %d assumption(s): %v", n-before, gs.Adapter.Graph.Assumptions[before:])
	}
	main := gs.Adapter.Graph.Arguments[r.ArgIDs[0]]
	if len(r.Presumptions) != len(main.Premises) { // practical_reasoning1: 4 premises + 3 declared assumptions
		t.Fatalf("%d presumptions for %d premises", len(r.Presumptions), len(main.Premises))
	}
	for _, p := range main.Premises {
		found := false
		for _, a := range p.Stmt.Args {
			if carneades.ArgumentRole(a) == carneades.RolePresumption && a.Metadata["owner"] == "Alice" {
				found = true
			}
		}
		if !found {
			t.Errorf("premise %s has no presumption owned by Alice", p.Stmt.Id)
		}
		if p.Stmt.Label != caes.In {
			t.Errorf("premise %s is %v, want In while unattacked", p.Stmt.Id, p.Stmt.Label)
		}
	}
	if main.Conclusion.Label != caes.In {
		t.Fatalf("conclusion %v, want In", main.Conclusion.Label)
	}
	if r.ScoreDelta != 7 { // 10 points x weight 0.7: presumptions and exception arguments earn nothing
		t.Fatalf("support scored %.2f, want 7", r.ScoreDelta)
	}
}

func TestUnderminingCanSucceedAndBeAnswered(t *testing.T) {
	gs := makeCoreGame(t)
	r, _ := gs.ExecuteTurn(prSupportAction())
	main := r.ArgIDs[0]
	concl := gs.Adapter.Graph.Arguments[main].Conclusion
	prem := possiblePremise(t, gs, main)
	r2, err := gs.ExecuteTurn(nascq.ProposedAction{Type: "attack", AttackType: "undermining", SchemeID: "AS:sign",
		AnswersCQID: "CQ:possibility", TargetArgID: main, TargetPremiseID: prem,
		VariableBindings: map[string]string{"A": "no_parliament_majority", "B": "tax_cannot_pass"}})
	if err != nil || !r2.Accepted {
		t.Fatalf("undermining rejected: %v %+v", err, r2)
	}
	if l := gs.Adapter.Graph.Statements[prem].Label; l != caes.Out {
		t.Fatalf("undermined premise %s is %v, want Out", prem, l)
	}
	if l := gs.Adapter.Graph.Statements["¬"+prem].Label; l != caes.In {
		t.Fatalf("its negation is %v, want In", l)
	}
	if concl.Label != caes.Out {
		t.Fatalf("the undermined argument's conclusion is %v, want Out", concl.Label)
	}
	// The presumer answers with a stronger argument for the premise (0.8 > 0.6): reinstated.
	if _, err := gs.Adapter.AddArgumentDirect("Alice", "credible_source", nil, prem, 0.8); err != nil {
		t.Fatal(err)
	}
	gs.Adapter.Evaluate()
	if gs.Adapter.Graph.Statements[prem].Label != caes.In || concl.Label != caes.In {
		t.Fatalf("after backing: premise %v, conclusion %v, want In and In", gs.Adapter.Graph.Statements[prem].Label, concl.Label)
	}
}
