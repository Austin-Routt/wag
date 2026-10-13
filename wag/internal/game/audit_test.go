package game

import (
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// The implementation tests the third-party audit of the Kialo pilot asks for
// before NASCQ changes are merged (audit §7.8, 2026-10-13). Items 3
// (cause to effect for a hypothetical policy) and 4 (analogy with separate
// source and target) need scheme changes the owner has not approved; item 7
// (Kialo parents) is the importer's.

// §7.8.1: negative consequences concludes the negation of the action positive
// consequences supports, and the two join one issue, even when neither is a
// reply to the other.
func TestAuditNegativeConsequencesJoinOneIssue(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.MaxTurns = 0
	pc := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:consequences", VariableBindings: map[string]string{"A": "tax meat", "G": "lower emissions"}})
	nc := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:negative_consequences", VariableBindings: map[string]string{"A": "tax meat", "G": "dearer food for the poor"}})
	p, n := gs.Adapter.Graph.Arguments[pc].Conclusion, gs.Adapter.Graph.Arguments[nc].Conclusion
	if n.Id != "¬"+p.Id || p.Issue == nil || p.Issue != n.Issue {
		t.Fatalf("%s and %s are not one issue (%v, %v)", p.Id, n.Id, p.Issue, n.Issue)
	}
	if p.Label == caes.In && n.Label == caes.In {
		t.Fatal("both positions are in")
	}
	if errs := gs.CheckInvariants(); len(errs) > 0 {
		t.Fatal(errs)
	}
}

// §7.8.2: each question of negative consequences has its locus and polarity:
// likelihood undermines brings_about(A,G), evidence undercuts the inference,
// opposite consequences rebuts the conclusion (for the action).
func TestAuditNegativeConsequencesQuestions(t *testing.T) {
	for _, c := range []struct {
		cq, scheme string
		bind       map[string]string
		concludes  string
	}{
		{"CQ:likelihood", "AS:expert_opinion", map[string]string{"W": "an economist", "D": "food prices"}, "¬brings_about(tax_meat,dearer_food_for_the_poor)"},
		{"CQ:evidence", "AS:expert_opinion", map[string]string{"W": "a statistician", "D": "evidence"}, "undercut("},
		{"CQ:opposite_consequences", "AS:consequences", map[string]string{"G": "lower emissions"}, "should_be_performed(tax_meat)"},
	} {
		gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
		gs.Config.MaxTurns = 0
		nc := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:negative_consequences", VariableBindings: map[string]string{"A": "tax meat", "G": "dearer food for the poor"}})
		a := play(t, gs, answer("attack", c.cq, c.scheme, nc, c.bind))
		if got := gs.Adapter.Graph.Arguments[a].Conclusion.Id; !strings.HasPrefix(got, c.concludes) {
			t.Errorf("%s attack concludes %s, want %s", c.cq, got, c.concludes)
		}
		if gs.Adapter.Graph.Arguments[nc].Weight != 0 && c.cq != "CQ:opposite_consequences" {
			t.Errorf("%s attack leaves %s in force", c.cq, nc)
		}
	}
}

// §7.8.5: "yes, something else would have to happen first" (the attack on
// the Intervening Actions question) argues the exception, which undercuts
// goal-based practical reasoning; "no" argues its negation and leaves the
// argument in force.
func TestAuditInterveningActions(t *testing.T) {
	for _, typ := range []string{"attack", "support"} {
		gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
		gs.Config.MaxTurns = 0
		pr := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:practical_reasoning_goal",
			VariableBindings: map[string]string{"A": "tax meat", "S1": "meat is cheap", "S2": "less meat is eaten", "G": "lower emissions"}})
		a := play(t, gs, answer(typ, "CQ:intervening_actions", "AS:expert_opinion", pr, map[string]string{"W": "a farm economist", "D": "agricultural policy"}))
		concl := gs.Adapter.Graph.Arguments[a].Conclusion.Id
		inForce := gs.Adapter.Graph.Arguments[pr].Weight > 0
		switch typ {
		case "attack":
			if concl != "intervening_actions(tax_meat,lower_emissions)" || inForce {
				t.Errorf("yes-answer concludes %s; %s in force: %v", concl, pr, inForce)
			}
		case "support":
			if concl != "¬intervening_actions(tax_meat,lower_emissions)" || !inForce {
				t.Errorf("no-answer concludes %s; %s in force: %v", concl, pr, inForce)
			}
		}
	}
}

// §7.8.6: a good consequence of another action does not show it is the
// better means: positive consequences cannot back the Other Means or Best
// Means questions (S5).
func TestAuditBestMeansNeedsComparativeEvidence(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.MaxTurns = 0
	pr := play(t, gs, prSupportAction())
	for _, cq := range []string{"CQ:other_means", "CQ:best_means"} {
		r, err := gs.ExecuteTurn(answer("attack", cq, "AS:consequences", pr, map[string]string{"A": "subsidize vegetables", "G": "lower emissions"}))
		if err != nil || r.Accepted || !strings.Contains(r.Error, "S5") {
			t.Errorf("%s backed by positive consequences: %+v %v", cq, r, err)
		}
	}
}

// §7.8.8: Other, and arguments for and against a premise, need no named
// scheme: Bob argues against a premise with his own premises.
func TestAuditOtherAndPremiseMovesNeedNoScheme(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.MaxTurns = 0
	gs.NASCQ.EnableOther()
	play(t, gs, prSupportAction())
	a := play(t, gs, nascq.ProposedAction{Type: "attack", TargetStmtID: "possible(carbon_tax)", SchemeID: nascq.OtherSchemeID,
		VariableBindings: map[string]string{"P1": "Parliament has voted the tax down twice"}})
	if c := gs.Adapter.Graph.Arguments[a].Conclusion.Id; c != "¬possible(carbon_tax)" {
		t.Fatalf("Other against a premise concludes %s", c)
	}
}
