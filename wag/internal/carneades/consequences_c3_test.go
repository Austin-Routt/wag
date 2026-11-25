package carneades

import (
	"os"
	"sort"
	"strings"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// The consequence schemes as decided in proposal D008-D074 (decisions D008,
// D053, D055, D065, D069, D072 and D074; 2026-11-25), in the engine, on the
// committed walton.yml (or WAG_C3_LIBRARY). Every argument has the same
// weight, those that prove an exception or undercut an argument included:
// the bridge is an entry condition, not a weight (D008).
const w = 0.6

func c3(t *testing.T) *Adapter {
	t.Helper()
	path := os.Getenv("WAG_C3_LIBRARY")
	if path == "" {
		path = "../../../docs/scheme-registry/walton-clean/walton.yml"
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	a, err := NewAdapterFromYAML(b)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func c3lab(a *Adapter, id string) caes.Label {
	if s := a.Graph.Statements[id]; s != nil {
		return s.Label
	}
	return caes.Undecided
}

func c3assume(a *Adapter, ids ...string) {
	for _, s := range ids {
		a.AddStatement(s, s)
		a.AddAssumption(s)
	}
}

func c3arg(t *testing.T, a *Adapter, owner, scheme string, b map[string]string, concl string) string {
	t.Helper()
	ids, err := a.AddSchemeArgument(owner, scheme, b, concl, w)
	if err != nil {
		t.Fatal(err)
	}
	return ids[0]
}

// c3reason is a reason of owner's own for stmt, assumed.
func c3reason(a *Adapter, owner, stmt string) string {
	reason := owner + "_reason_for_" + stmt
	c3assume(a, reason)
	return reason
}

// c3argue makes a statement In by an argument of owner's from a reason of
// its own.
func c3argue(t *testing.T, a *Adapter, owner, stmt string) {
	t.Helper()
	reason := c3reason(a, owner, stmt)
	a.AddStatement(stmt, stmt)
	if _, err := a.AddArgumentDirect(owner, "other", []string{reason}, stmt, w); err != nil {
		t.Fatal(err)
	}
}

func c3undercut(a *Adapter, id string) caes.Label {
	if u := a.Graph.Arguments[id].Undercutter; u != nil {
		return u.Label
	}
	return caes.Undecided
}

// c3exceptions lists the exceptions an argument was instantiated with.
func c3exceptions(a *Adapter, id string) []string {
	var got []string
	if u := a.Graph.Arguments[id].Undercutter; u != nil {
		for _, x := range u.Args {
			if ArgumentRole(x) == RoleException && len(x.Premises) == 1 {
				got = append(got, x.Premises[0].Stmt.Id)
			}
		}
	}
	sort.Strings(got)
	return got
}

// D008: a possible consequence cannot fill a categorical premise or
// conclusion; the bridge needs its premise; with it, the thesis is reached.
func TestConsequencesModalCannotFillCategorical(t *testing.T) {
	a := c3(t)
	c3assume(a, "may_bring_about(tax,g)", "good(g)")
	c3arg(t, a, "pro", "positive_consequences", map[string]string{"A": "tax", "G": "g"}, "should_be_performed(tax)")
	c3arg(t, a, "pro", "positive_consequences_modal", map[string]string{"A": "tax", "G": "g"}, "may_be_worth_performing(tax,g)")
	a.Evaluate()
	if c3lab(a, "brings_about(tax,g)") == caes.In || c3lab(a, "should_be_performed(tax)") == caes.In {
		t.Fatalf("a possible consequence filled the categorical premise or conclusion: brings_about=%v should=%v",
			c3lab(a, "brings_about(tax,g)"), c3lab(a, "should_be_performed(tax)"))
	}
	if c3lab(a, "may_be_worth_performing(tax,g)") != caes.In {
		t.Fatalf("the modal conclusion is %v", c3lab(a, "may_be_worth_performing(tax,g)"))
	}
	c3arg(t, a, "pro", "bridge_from_possible_good", map[string]string{"A": "tax", "G": "g"}, "should_be_performed(tax)")
	a.Evaluate()
	if got := c3lab(a, "should_be_performed(tax)"); got == caes.In {
		t.Fatalf("the bridge reached the thesis without its premise (an entry condition): %v", got)
	}
	c3assume(a, "matters_to_decision(tax,g)")
	a.Evaluate()
	if got := c3lab(a, "should_be_performed(tax)"); got != caes.In {
		t.Fatalf("the bridge with its premise did not reach the thesis: %v", got)
	}
	t.Logf("modal alone: thesis not In; bridge without its premise: not In; with it: In")
}

// D008: each bridge's exception names A and G: defeating one side's bridge
// leaves the other's when they cite different consequences. The exception
// takes a bridge away, not the decision: the thesis goes by the arguments
// left (proposal D008-D074, version 5 as amended).
func TestConsequencesBridgesAreSided(t *testing.T) {
	a := c3(t)
	c3assume(a, "may_bring_about(tax,cut)", "good(cut)", "matters_to_decision(tax,cut)",
		"may_bring_about(tax,corners)", "bad(corners)", "matters_to_decision(tax,corners)")
	c3arg(t, a, "pro", "positive_consequences_modal", map[string]string{"A": "tax", "G": "cut"}, "may_be_worth_performing(tax,cut)")
	pro := c3arg(t, a, "pro", "bridge_from_possible_good", map[string]string{"A": "tax", "G": "cut"}, "should_be_performed(tax)")
	c3arg(t, a, "con", "negative_consequences_modal", map[string]string{"A": "tax", "G": "corners"}, "may_be_unwise_to_perform(tax,corners)")
	con := c3arg(t, a, "con", "bridge_from_possible_harm", map[string]string{"A": "tax", "G": "corners"}, "¬should_be_performed(tax)")
	// Pro shows that it is worth finding out first whether the tax brings
	// about Con's consequence.
	c3argue(t, a, "pro", "worth_finding_out_first(tax,corners)")
	a.Evaluate()
	if c3undercut(a, con) != caes.In || c3undercut(a, pro) == caes.In {
		t.Fatalf("Con's bridge undercut=%v, Pro's=%v", c3undercut(a, con), c3undercut(a, pro))
	}
	if got := c3lab(a, "should_be_performed(tax)"); got != caes.In {
		t.Fatalf("the thesis is %v: Pro's bridge stands, and the exception does not hold the decision back", got)
	}
}

// D008: the bridge is no weight: a bridged possible harm against a certain
// good of the same weight is a tie; without the bridge's premise the good wins.
func TestConsequencesEqualWeightsTie(t *testing.T) {
	for _, matters := range []bool{true, false} {
		a := c3(t)
		c3assume(a, "brings_about(tax,health)", "good(health)", "may_bring_about(tax,corners)", "bad(corners)")
		if matters {
			c3assume(a, "matters_to_decision(tax,corners)")
		}
		c3arg(t, a, "pro", "positive_consequences", map[string]string{"A": "tax", "G": "health"}, "should_be_performed(tax)")
		c3arg(t, a, "con", "negative_consequences_modal", map[string]string{"A": "tax", "G": "corners"}, "may_be_unwise_to_perform(tax,corners)")
		c3arg(t, a, "con", "bridge_from_possible_harm", map[string]string{"A": "tax", "G": "corners"}, "¬should_be_performed(tax)")
		a.Evaluate()
		got, neg := c3lab(a, "should_be_performed(tax)"), c3lab(a, "¬should_be_performed(tax)")
		t.Logf("matters=%v: should=%v ¬should=%v", matters, got, neg)
		if matters && (got == caes.In || neg == caes.In) {
			t.Errorf("equal weights did not tie: %v / %v", got, neg)
		}
		if !matters && got != caes.In {
			t.Errorf("without the bridge's premise the certain good did not win: %v", got)
		}
	}
}

// D055, kept: one weighing claim defeats a categorical and a modal argument
// about the same A and G (it compares magnitudes, not likelihood).
func TestConsequencesOutweighingIsShared(t *testing.T) {
	a := c3(t)
	c3assume(a, "brings_about(tax,corners)", "may_bring_about(tax,corners)", "bad(corners)",
		"brings_about(tax,health)", "good(health)", "matters_more(health,corners)")
	cat := c3arg(t, a, "con1", "negative_consequences", map[string]string{"A": "tax", "G": "corners"}, "¬should_be_performed(tax)")
	mod := c3arg(t, a, "con2", "negative_consequences_modal", map[string]string{"A": "tax", "G": "corners"}, "may_be_unwise_to_perform(tax,corners)")
	c3arg(t, a, "pro", "good_outweighs_bad", map[string]string{"A": "tax", "H": "health", "G": "corners"}, "good_outweighs(tax,corners)")
	a.Evaluate()
	if c3undercut(a, cat) != caes.In || c3undercut(a, mod) != caes.In {
		t.Fatalf("categorical undercut=%v, modal undercut=%v", c3undercut(a, cat), c3undercut(a, mod))
	}
}

// D065, D069, D072 and D008: each new exception, once argued, undercuts the
// arguments that name it and no other.
func TestConsequencesExceptionsReachOnlyTheirArguments(t *testing.T) {
	type c struct {
		scheme, concl string
		b             map[string]string
		premises      []string
	}
	cases := []struct {
		exception string
		hit, miss c
	}{
		{"too_slight(tax,price)",
			c{"negative_consequences", "¬should_be_performed(tax)", map[string]string{"A": "tax", "G": "price"}, []string{"brings_about(tax,price)", "bad(price)"}},
			c{"negative_consequences", "¬should_be_performed(tax)", map[string]string{"A": "tax", "G": "jobs"}, []string{"brings_about(tax,jobs)", "bad(jobs)"}}},
		{"gain_discounted_by_bearer(tax,health)",
			c{"positive_consequences_modal", "may_be_worth_performing(tax,health)", map[string]string{"A": "tax", "G": "health"}, []string{"may_bring_about(tax,health)", "good(health)"}},
			c{"positive_consequences", "should_be_performed(tax)", map[string]string{"A": "tax", "G": "climate"}, []string{"brings_about(tax,climate)", "good(climate)"}}},
		{"harm_discounted_by_bearer(tax,prices)",
			c{"negative_consequences_modal", "may_be_unwise_to_perform(tax,prices)", map[string]string{"A": "tax", "G": "prices"}, []string{"may_bring_about(tax,prices)", "bad(prices)"}},
			c{"negative_consequences", "¬should_be_performed(tax)", map[string]string{"A": "tax", "G": "jobs"}, []string{"brings_about(tax,jobs)", "bad(jobs)"}}},
		{"fails_to_avert(levy,crisis)",
			c{"negative_consequences_prudential", "should_be_performed(levy)", map[string]string{"X": "levy", "C": "crisis"}, []string{"not_doing(levy)", "bad_consequence(levy,crisis)", "strongly_occur(levy,crisis)"}},
			c{"negative_consequences_prudential", "should_be_performed(levy)", map[string]string{"X": "levy", "C": "strike"}, []string{"not_doing(levy)", "bad_consequence(levy,strike)", "strongly_occur(levy,strike)"}}},
		{"worth_finding_out_first(tax,corners)",
			c{"bridge_from_possible_harm", "¬should_be_performed(tax)", map[string]string{"A": "tax", "G": "corners"}, []string{"may_be_unwise_to_perform(tax,corners)", "matters_to_decision(tax,corners)"}},
			c{"bridge_from_possible_harm", "¬should_be_performed(tax)", map[string]string{"A": "tax", "G": "jobs"}, []string{"may_be_unwise_to_perform(tax,jobs)", "matters_to_decision(tax,jobs)"}}},
		{"preventable_otherwise(testing,drugs)",
			c{"harm_reduction", "should_be_performed(testing)", map[string]string{"A": "testing", "E": "drugs"}, []string{"happens_anyway(testing,drugs)", "less_harmful_per_occurrence(testing,drugs)"}},
			c{"harm_reduction", "should_be_performed(testing)", map[string]string{"A": "testing", "E": "drink"}, []string{"happens_anyway(testing,drink)", "less_harmful_per_occurrence(testing,drink)"}}},
		{"no_net_reduction(testing,drugs)",
			c{"harm_reduction", "should_be_performed(testing)", map[string]string{"A": "testing", "E": "drugs"}, []string{"happens_anyway(testing,drugs)", "less_harmful_per_occurrence(testing,drugs)"}},
			c{"harm_reduction", "should_be_performed(testing)", map[string]string{"A": "testing", "E": "drink"}, []string{"happens_anyway(testing,drink)", "less_harmful_per_occurrence(testing,drink)"}}},
	}
	for _, x := range cases {
		a := c3(t)
		c3assume(a, x.hit.premises...)
		c3assume(a, x.miss.premises...)
		hit := c3arg(t, a, "p1", x.hit.scheme, x.hit.b, x.hit.concl)
		miss := c3arg(t, a, "p2", x.miss.scheme, x.miss.b, x.miss.concl)
		c3argue(t, a, "critic", x.exception)
		a.Evaluate()
		if c3undercut(a, hit) != caes.In || c3undercut(a, miss) == caes.In {
			t.Errorf("%s: undercuts %s %v (want In) and %s %v (want not In)", x.exception, x.hit.scheme, c3undercut(a, hit), x.miss.scheme, c3undercut(a, miss))
		}
	}
}

// Every scheme the proposal changes or adds, with exactly the exceptions it
// gives that scheme (33.3's unchanged): each exception, argued by a critic,
// undercuts an argument of the scheme, whose conclusion was In before.
// Scheme by scheme, so that an exception missing from one scheme that should
// have it, or present on one that should not, fails here.
func TestConsequencesEveryExceptionOnEveryScheme(t *testing.T) {
	type shape struct{ premises, assumptions, conclusions []string }
	cases := []struct {
		scheme, concl string
		b             map[string]string
		exceptions    []string
		shape         shape
	}{
		{"positive_consequences", "should_be_performed(tax)", map[string]string{"A": "tax", "G": "g"},
			[]string{"bad_outweighs(tax,g)", "too_slight(tax,g)", "gain_discounted_by_bearer(tax,g)"},
			shape{[]string{"brings_about(A,G)", "good(G)"}, nil, []string{"should_be_performed(A)"}}},
		{"negative_consequences", "¬should_be_performed(tax)", map[string]string{"A": "tax", "G": "g"},
			[]string{"good_outweighs(tax,g)", "too_slight(tax,g)", "harm_discounted_by_bearer(tax,g)"},
			shape{[]string{"brings_about(A,G)", "bad(G)"}, nil, []string{"¬should_be_performed(A)"}}},
		{"positive_consequences_modal", "may_be_worth_performing(tax,g)", map[string]string{"A": "tax", "G": "g"},
			[]string{"bad_outweighs(tax,g)", "too_slight(tax,g)", "gain_discounted_by_bearer(tax,g)"},
			shape{[]string{"may_bring_about(A,G)", "good(G)"}, nil, []string{"may_be_worth_performing(A,G)"}}},
		{"negative_consequences_modal", "may_be_unwise_to_perform(tax,g)", map[string]string{"A": "tax", "G": "g"},
			[]string{"good_outweighs(tax,g)", "too_slight(tax,g)", "harm_discounted_by_bearer(tax,g)"},
			shape{[]string{"may_bring_about(A,G)", "bad(G)"}, nil, []string{"may_be_unwise_to_perform(A,G)"}}},
		{"bridge_from_possible_good", "should_be_performed(tax)", map[string]string{"A": "tax", "G": "g"},
			[]string{"worth_finding_out_first(tax,g)"},
			shape{[]string{"may_be_worth_performing(A,G)", "matters_to_decision(A,G)"}, nil, []string{"should_be_performed(A)"}}},
		{"bridge_from_possible_harm", "¬should_be_performed(tax)", map[string]string{"A": "tax", "G": "g"},
			[]string{"worth_finding_out_first(tax,g)"},
			shape{[]string{"may_be_unwise_to_perform(A,G)", "matters_to_decision(A,G)"}, nil, []string{"¬should_be_performed(A)"}}},
		{"negative_consequences_prudential", "should_be_performed(tax)", map[string]string{"X": "tax", "C": "g"},
			[]string{"bad_outweighs(tax,g)", "fails_to_avert(tax,g)"},
			shape{[]string{"not_doing(X)", "bad_consequence(X,C)"}, []string{"strongly_occur(X,C)"}, []string{"should_be_performed(X)"}}},
		{"negative_consequences_reasoning", "¬should_be_performed(tax)", map[string]string{"A": "tax", "B": "g"},
			[]string{"good_outweighs(tax,g)"},
			shape{[]string{"agent_bring_about(A,B)", "bad_for_goals(B)"}, nil, []string{"¬should_be_performed(A)"}}},
		{"harm_reduction", "should_be_performed(tax)", map[string]string{"A": "tax", "E": "g"},
			[]string{"preventable_otherwise(tax,g)", "no_net_reduction(tax,g)"},
			shape{[]string{"happens_anyway(A,E)", "less_harmful_per_occurrence(A,E)"}, nil, []string{"should_be_performed(A)"}}},
	}
	for _, x := range cases {
		a := c3(t)
		// The scheme as the proposal gives it: a modal variant or a bridge
		// concluding another formula is caught here, not only by the
		// catalogue's checks (the check of version 5, W2).
		sch := a.schemeMap[x.scheme]
		if sch == nil {
			t.Errorf("%s is not in the library", x.scheme)
			continue
		}
		for _, f := range []struct {
			what      string
			got, want []string
		}{{"premises", sch.Premises, x.shape.premises}, {"assumptions", sch.Assumptions, x.shape.assumptions},
			{"conclusions", sch.Conclusions, x.shape.conclusions}} {
			if strings.Join(f.got, " ") != strings.Join(f.want, " ") {
				t.Errorf("%s has the %s %v, want %v", x.scheme, f.what, f.got, f.want)
			}
		}
		id := c3arg(t, a, "p", x.scheme, x.b, x.concl)
		want := append([]string{}, x.exceptions...)
		sort.Strings(want)
		if got := c3exceptions(a, id); strings.Join(got, " ") != strings.Join(want, " ") {
			t.Errorf("%s has the exceptions %v, want %v", x.scheme, got, want)
			continue
		}
		for _, ex := range x.exceptions {
			a := c3(t)
			id := c3arg(t, a, "p", x.scheme, x.b, x.concl)
			for _, p := range a.Graph.Arguments[id].Premises {
				a.AddAssumption(p.Stmt.Id)
			}
			a.Evaluate()
			if got := c3lab(a, x.concl); got != caes.In {
				t.Errorf("%s: its conclusion is %v with its premises assumed", x.scheme, got)
				continue
			}
			c3argue(t, a, "critic", ex)
			a.Evaluate()
			if c3undercut(a, id) != caes.In || c3lab(a, x.concl) == caes.In {
				t.Errorf("%s: %s argued, undercutter %v, conclusion %v", x.scheme, ex, c3undercut(a, id), c3lab(a, x.concl))
			}
		}
	}
}

// D053, the bearer flip (the review of version 2, E1): the same G, good for
// Pro and bad for Con; that it falls on the polluters who cause the harm
// discounts it as a harm only. Proved against Con, it leaves Pro standing.
func TestConsequencesBearerIsSided(t *testing.T) {
	a := c3(t)
	c3assume(a, "brings_about(carbon_tax,firms_pay)", "good(firms_pay)", "bad(firms_pay)")
	pro := c3arg(t, a, "pro", "positive_consequences", map[string]string{"A": "carbon_tax", "G": "firms_pay"}, "should_be_performed(carbon_tax)")
	con := c3arg(t, a, "con", "negative_consequences", map[string]string{"A": "carbon_tax", "G": "firms_pay"}, "¬should_be_performed(carbon_tax)")
	c3argue(t, a, "pro", "harm_discounted_by_bearer(carbon_tax,firms_pay)")
	a.Evaluate()
	if c3undercut(a, con) != caes.In || c3undercut(a, pro) == caes.In {
		t.Fatalf("Con undercut=%v, Pro undercut=%v", c3undercut(a, con), c3undercut(a, pro))
	}
	t.Logf("Con's argument undercut, Pro's standing; thesis %v", c3lab(a, "should_be_performed(carbon_tax)"))
}

// D053's declarations, one example per side: a neutral exception, proved for
// a reason that favours either side, defeats both sides' arguments about the
// same A and G: too_slight the categorical and modal ones,
// worth_finding_out_first the bridges.
func TestConsequencesNeutralExceptionsPerSide(t *testing.T) {
	for _, side := range []string{"pro", "con"} {
		a := c3(t)
		c3assume(a, "brings_about(tax,g)", "may_bring_about(tax,g)", "good(g)", "bad(g)")
		ids := []string{
			c3arg(t, a, "pro", "positive_consequences", map[string]string{"A": "tax", "G": "g"}, "should_be_performed(tax)"),
			c3arg(t, a, "con", "negative_consequences", map[string]string{"A": "tax", "G": "g"}, "¬should_be_performed(tax)"),
			c3arg(t, a, "pro", "positive_consequences_modal", map[string]string{"A": "tax", "G": "g"}, "may_be_worth_performing(tax,g)"),
			c3arg(t, a, "con", "negative_consequences_modal", map[string]string{"A": "tax", "G": "g"}, "may_be_unwise_to_perform(tax,g)"),
		}
		c3argue(t, a, side, "too_slight(tax,g)") // the side's own argument and reason; the statement is the same
		a.Evaluate()
		for _, id := range ids {
			if c3undercut(a, id) != caes.In {
				t.Errorf("too_slight proved by %s: argument %s is not undercut", side, id)
			}
		}

		a = c3(t)
		c3assume(a, "may_bring_about(tax,g)", "good(g)", "bad(g)", "matters_to_decision(tax,g)")
		c3arg(t, a, "pro", "positive_consequences_modal", map[string]string{"A": "tax", "G": "g"}, "may_be_worth_performing(tax,g)")
		c3arg(t, a, "con", "negative_consequences_modal", map[string]string{"A": "tax", "G": "g"}, "may_be_unwise_to_perform(tax,g)")
		pro := c3arg(t, a, "pro", "bridge_from_possible_good", map[string]string{"A": "tax", "G": "g"}, "should_be_performed(tax)")
		con := c3arg(t, a, "con", "bridge_from_possible_harm", map[string]string{"A": "tax", "G": "g"}, "¬should_be_performed(tax)")
		c3argue(t, a, side, "worth_finding_out_first(tax,g)")
		a.Evaluate()
		if c3undercut(a, pro) != caes.In || c3undercut(a, con) != caes.In {
			t.Errorf("worth_finding_out_first proved by %s: Pro's bridge %v, Con's %v", side, c3undercut(a, pro), c3undercut(a, con))
		}
	}
}

// D074, as the owner refined it on 2026-11-25: the objection that
// consequences are no reason to think a claim true undercuts the specific
// argument it answers (guidelines OC18), not every argument about the same
// action, as an exception on A would. A legitimate argument about believing
// (OC18-c) stands beside the misplaced one (OC1-b), and the thesis it
// supports stays In.
func TestConsequencesAppealUndercutsOneArgument(t *testing.T) {
	a := c3(t)
	c3assume(a, "brings_about(believing_in_god,psychosocial_benefits)", "good(psychosocial_benefits)",
		"brings_about(believing_in_god,comfort_in_hard_times)", "good(comfort_in_hard_times)")
	misplaced := c3arg(t, a, "p1", "positive_consequences", map[string]string{"A": "believing_in_god", "G": "psychosocial_benefits"}, "should_be_performed(believing_in_god)")
	legitimate := c3arg(t, a, "p2", "positive_consequences", map[string]string{"A": "believing_in_god", "G": "comfort_in_hard_times"}, "should_be_performed(believing_in_god)")
	reason := c3reason(a, "critic", "usefulness_no_sign_of_truth")
	if _, err := a.AddUndercutter("critic", "other", misplaced, []string{reason}, w); err != nil {
		t.Fatal(err)
	}
	a.Evaluate()
	if c3undercut(a, misplaced) != caes.In || c3undercut(a, legitimate) == caes.In {
		t.Fatalf("misplaced undercut=%v, legitimate undercut=%v", c3undercut(a, misplaced), c3undercut(a, legitimate))
	}
	if got := c3lab(a, "should_be_performed(believing_in_god)"); got != caes.In {
		t.Fatalf("the legitimate argument's conclusion is %v", got)
	}
}

// D008: a bridge's exception reaches the bridge alone: a categorical argument
// about the same A and G (and the modal argument the bridge rests on) stand.
// Whether A brings about G at all is the categorical argument's premise,
// questioned by 33/CQ1, not by the bridge's exception.
func TestConsequencesBridgeExceptionSparesCategorical(t *testing.T) {
	a := c3(t)
	c3assume(a, "brings_about(tax,g)", "may_bring_about(tax,g)", "good(g)", "matters_to_decision(tax,g)")
	cat := c3arg(t, a, "pro1", "positive_consequences", map[string]string{"A": "tax", "G": "g"}, "should_be_performed(tax)")
	mod := c3arg(t, a, "pro2", "positive_consequences_modal", map[string]string{"A": "tax", "G": "g"}, "may_be_worth_performing(tax,g)")
	bridge := c3arg(t, a, "pro2", "bridge_from_possible_good", map[string]string{"A": "tax", "G": "g"}, "should_be_performed(tax)")
	c3argue(t, a, "critic", "worth_finding_out_first(tax,g)")
	a.Evaluate()
	if c3undercut(a, bridge) != caes.In || c3undercut(a, cat) == caes.In || c3undercut(a, mod) == caes.In {
		t.Fatalf("bridge undercut=%v, categorical=%v, modal=%v", c3undercut(a, bridge), c3undercut(a, cat), c3undercut(a, mod))
	}
	if got := c3lab(a, "should_be_performed(tax)"); got != caes.In {
		t.Fatalf("the categorical argument's conclusion is %v", got)
	}
}

// D008, as amended (the check of version 5, W1): a categorical argument that
// asserts the link meets a bridged possibility of the same G on the other
// side. The bridges' exception does not reach the categorical argument: proved
// alone, it would leave that side the winner, favouring whoever claims the
// link more firmly. A reply addressed to the categorical argument challenges
// its brings_about(A,G) instead (OC15), which its proponent must support:
// unsupported, neither argument stands.
func TestConsequencesBridgeExceptionAgainstAFirmClaim(t *testing.T) {
	for _, challenged := range []bool{false, true} {
		a := c3(t)
		c3assume(a, "good(g)", "may_bring_about(tax,g)", "bad(g)", "matters_to_decision(tax,g)")
		if !challenged {
			c3assume(a, "brings_about(tax,g)")
		}
		cat := c3arg(t, a, "pro", "positive_consequences", map[string]string{"A": "tax", "G": "g"}, "should_be_performed(tax)")
		c3arg(t, a, "con", "negative_consequences_modal", map[string]string{"A": "tax", "G": "g"}, "may_be_unwise_to_perform(tax,g)")
		br := c3arg(t, a, "con", "bridge_from_possible_harm", map[string]string{"A": "tax", "G": "g"}, "¬should_be_performed(tax)")
		c3argue(t, a, "critic", "worth_finding_out_first(tax,g)")
		a.Evaluate()
		should, not := c3lab(a, "should_be_performed(tax)"), c3lab(a, "¬should_be_performed(tax)")
		if c3undercut(a, br) != caes.In || c3undercut(a, cat) == caes.In {
			t.Fatalf("challenged=%v: bridge undercut=%v, categorical undercut=%v", challenged, c3undercut(a, br), c3undercut(a, cat))
		}
		if !challenged && should != caes.In {
			t.Errorf("the link unchallenged: the categorical argument should win alone (should=%v)", should)
		}
		if challenged && (should == caes.In || not == caes.In) {
			t.Errorf("the link challenged and unsupported: neither side should stand (should=%v, not=%v)", should, not)
		}
	}
}
