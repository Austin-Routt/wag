package carneades

import (
	"strings"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// The answer scheme for the bridges' exception (proposal D095, decision
// D095 (3)): finding_out_first concludes worth_finding_out_first(A,G) from
// the conditions of the owner's declaration under D053, on the committed
// walton.yml (or WAG_C3_LIBRARY), every argument at the same weight (c3).

// fofPremises are the answer's premises for A = tax, G = g, I = trial.
var fofPremises = []string{"settles_whether(trial,tax,g)", "done_in_time(trial,tax)", "worth_its_costs(trial,tax,g)"}

var fofBindings = map[string]string{"A": "tax", "G": "g", "I": "trial"}

// The scheme as the proposal gives it.
func TestFindingOutFirstAsProposed(t *testing.T) {
	a := c3(t)
	sch := a.schemeMap["finding_out_first"]
	if sch == nil {
		t.Fatal("finding_out_first is not in the library")
	}
	for _, f := range []struct {
		what      string
		got, want []string
	}{
		{"premises", sch.Premises, []string{"settles_whether(I,A,G)", "done_in_time(I,A)", "worth_its_costs(I,A,G)"}},
		{"assumptions", sch.Assumptions, nil},
		{"exceptions", sch.Exceptions, nil},
		{"conclusions", sch.Conclusions, []string{"worth_finding_out_first(A,G)"}},
	} {
		if strings.Join(f.got, " ") != strings.Join(f.want, " ") {
			t.Errorf("finding_out_first has the %s %v, want %v", f.what, f.got, f.want)
		}
	}
}

// fofBoard puts both modal arguments about tax and g on the board with their
// bridges, and a bridge about another consequence g2 and a categorical
// argument about g; it returns the arguments' ids.
func fofBoard(t *testing.T, a *Adapter) (pro, con, other, cat, proModal, conModal string) {
	t.Helper()
	c3assume(a, "may_bring_about(tax,g)", "good(g)", "bad(g)", "matters_to_decision(tax,g)",
		"may_bring_about(tax,g2)", "good(g2)", "matters_to_decision(tax,g2)", "brings_about(tax,g)")
	proModal = c3arg(t, a, "pro", "positive_consequences_modal", map[string]string{"A": "tax", "G": "g"}, "may_be_worth_performing(tax,g)")
	conModal = c3arg(t, a, "con", "negative_consequences_modal", map[string]string{"A": "tax", "G": "g"}, "may_be_unwise_to_perform(tax,g)")
	pro = c3arg(t, a, "pro", "bridge_from_possible_good", map[string]string{"A": "tax", "G": "g"}, "should_be_performed(tax)")
	con = c3arg(t, a, "con", "bridge_from_possible_harm", map[string]string{"A": "tax", "G": "g"}, "¬should_be_performed(tax)")
	c3arg(t, a, "pro", "positive_consequences_modal", map[string]string{"A": "tax", "G": "g2"}, "may_be_worth_performing(tax,g2)")
	other = c3arg(t, a, "pro", "bridge_from_possible_good", map[string]string{"A": "tax", "G": "g2"}, "should_be_performed(tax)")
	cat = c3arg(t, a, "pro", "positive_consequences", map[string]string{"A": "tax", "G": "g"}, "should_be_performed(tax)")
	return
}

// Offered by either side, with its three premises, the answer proves the
// exception, which defeats both bridges about the same A and G (D053's
// declaration, one example per side) and nothing else: not a bridge about
// another consequence, not the modal arguments, not a categorical argument
// about the same A and G.
func TestFindingOutFirstDefeatsBothBridgesPerSide(t *testing.T) {
	for _, side := range []string{"pro", "con"} {
		a := c3(t)
		pro, con, other, cat, proModal, conModal := fofBoard(t, a)
		c3assume(a, fofPremises...)
		ans := c3arg(t, a, side, "finding_out_first", fofBindings, "worth_finding_out_first(tax,g)")
		a.Evaluate()
		if got := c3lab(a, "worth_finding_out_first(tax,g)"); got != caes.In {
			t.Fatalf("offered by %s: the exception is %v", side, got)
		}
		if c3undercut(a, ans) == caes.In {
			t.Errorf("offered by %s: the answer itself is undercut", side)
		}
		if c3undercut(a, pro) != caes.In || c3undercut(a, con) != caes.In {
			t.Errorf("offered by %s: Pro's bridge undercut %v, Con's %v", side, c3undercut(a, pro), c3undercut(a, con))
		}
		for name, id := range map[string]string{"the bridge about g2": other, "the categorical argument": cat,
			"Pro's modal argument": proModal, "Con's modal argument": conModal} {
			if c3undercut(a, id) == caes.In {
				t.Errorf("offered by %s: %s is undercut", side, name)
			}
		}
	}
}

// Each condition is a premise its arguer supports: with any one of them not
// established, the exception is not proved and both bridges stand (a test
// that comes too late, one that would not settle the question, one not worth
// its cost). Here the condition is denied by the bridges' side, at the same
// weight as the answer's reason for it: under the preponderance of evidence a
// tie leaves the condition out, and with it the exception. MakeIssue puts the
// condition and its denial in one issue, as a premise attack in the game does
// (opposingPosition, answers.go); without the issue the two would be unrelated
// statements and the condition In on its own argument.
func TestFindingOutFirstNeedsEachCondition(t *testing.T) {
	for i, cond := range fofPremises {
		a := c3(t)
		pro, con, _, _, _, _ := fofBoard(t, a)
		for j, p := range fofPremises {
			if j != i {
				c3assume(a, p)
			}
		}
		if _, err := a.MakeIssue(cond, caes.PE); err != nil {
			t.Fatal(err)
		}
		c3argue(t, a, "critic", cond)
		c3argue(t, a, "pro", "¬"+cond)
		c3arg(t, a, "critic", "finding_out_first", fofBindings, "worth_finding_out_first(tax,g)")
		a.Evaluate()
		if got := c3lab(a, cond); got == caes.In {
			t.Fatalf("%s: a tie left the condition %v", cond, got)
		}
		if got := c3lab(a, "worth_finding_out_first(tax,g)"); got == caes.In {
			t.Errorf("%s not established: the exception is %v", cond, got)
		}
		if c3undercut(a, pro) == caes.In || c3undercut(a, con) == caes.In {
			t.Errorf("%s not established: Pro's bridge undercut %v, Con's %v", cond, c3undercut(a, pro), c3undercut(a, con))
		}
	}
}
