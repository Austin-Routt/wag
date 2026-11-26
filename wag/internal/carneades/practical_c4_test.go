package carneades

import (
	"os"
	"sort"
	"strings"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// The rest of the first family as decided in proposal D022-D093 (decisions
// D022, D023, D031, D033, D060, D062, D063 and D093; 2026-11-26), in the
// engine, on the committed walton.yml (or WAG_C4_LIBRARY). Every argument has
// the same weight (w), those that prove an exception included.
func c4(t *testing.T) *Adapter {
	t.Helper()
	path := os.Getenv("WAG_C4_LIBRARY")
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

// c4shape is a scheme as the proposal gives it: its formulas, and an
// instance of it with its premises and assumptions (instantiated).
type c4shape struct {
	scheme, concl                                  string
	b                                              map[string]string
	premises, assumptions, exceptions, conclusions []string
}

var c4schemes = []c4shape{
	{"fear", "should_be_performed(brakes)", map[string]string{"X": "brakes", "Y": "crash"},
		[]string{"not_do_appear(X,Y)", "bad_for_you(Y)", "prevent(Y)", "only_way_prevent(Y,X)"}, nil, nil, []string{"should_be_performed(X)"}},
	{"danger", "¬should_be_performed(quarry)", map[string]string{"X": "quarry", "Y": "rocks"},
		[]string{"occur(X,Y)", "danger(Y)"}, nil, nil, []string{"¬should_be_performed(X)"}},
	{"threat", "¬should_be_performed(extension)", map[string]string{"X": "extension", "C": "lawsuit"},
		[]string{"cited_bad_consequences(X,C)", "position_to_bring_about(C)", "assert_bad(X,C)"}, nil, nil, []string{"¬should_be_performed(X)"}},
	{"ad_baculum", "should_be_performed(pay)", map[string]string{"X": "pay", "Y": "towed"},
		[]string{"alternative_bad(X,Y)", "undesirable(Y)"}, nil, nil, []string{"should_be_performed(X)"}},
	{"pragmatic_alternative", "should_be_performed(dyke)", map[string]string{"A": "dyke", "B": "flood"},
		[]string{"alternatives(A,B)", "undesirable(B)"}, nil, nil, []string{"should_be_performed(A)"}},
	{"appeal_to_vanity_ad_populum", "should_be_performed(carbon)", map[string]string{"G": "riders", "A": "frames", "X": "carbon"},
		[]string{"admired_people(G,A)", "group_action(X,G)"}, nil, nil, []string{"should_be_performed(X)"}},
	{"practical_goal", "should_do(maya,walk)", map[string]string{"P": "maya", "X": "walk", "G": "exercise"},
		[]string{"contributing_action(X,G)", "goal_of_agent(P,G)"}, nil,
		[]string{"realize_more_effectively(P,G,X)", "incompatible_goal(P,G)", "bad_outweighs(X,G)"}, []string{"should_do(P,X)"}},
	{"two_person_practical", "should_do(lena,train)", map[string]string{"A": "lena", "B": "coach", "X": "marathon", "Y": "train"},
		[]string{"tells_intend(A,X,B)", "situation_necessity(B,Y,X,A)"}, []string{"possible_bring_about(A,Y)"},
		[]string{"realize_more_effectively(A,X,Y)", "incompatible_goal(A,X)", "bad_outweighs(Y,X)"}, []string{"should_do(A,Y)"}},
	{"waste", "ought_to_continue(council,tram)", map[string]string{"A": "council", "X": "tram"},
		[]string{"stop_realize_wasted(A,X)", "waste_bad_thing(A,X)"}, []string{"possible_goal(X)"},
		[]string{"reassessment_against_continuing(A,X)"}, []string{"ought_to_continue(A,X)"}},
	{"sunk_costs_precommitment", "should_do(priya,race)", map[string]string{"P": "priya", "A": "race", "T1": "march", "T2": "october"},
		[]string{"choice_at(T2,A)", "precommitted(P,A,T1,T2)"}, nil, nil, []string{"should_do(P,A)"}},
	{"end_and_means_1", "good(cooking)", map[string]string{"X": "cooking", "Y": "eating_well"},
		[]string{"means_to_end(X,Y)", "good(Y)"}, nil, []string{"means_bad_outweighs(X,Y)"}, []string{"good(X)"}},
	{"end_and_means_bad", "bad(rigging)", map[string]string{"X": "rigging", "Y": "corrupt_mayor"},
		[]string{"means_to_end(X,Y)", "bad(Y)"}, nil, []string{"means_good_outweighs(X,Y)"}, []string{"bad(X)"}},
	{"end_and_means_less_bad", "less_bad_as_means(injection,recovery)", map[string]string{"X": "injection", "Y": "recovery"},
		[]string{"means_to_end(X,Y)", "good(Y)", "bad(X)"}, nil, nil, []string{"less_bad_as_means(X,Y)"}},
	{"bridge_from_agent", "should_be_performed(walk)", map[string]string{"P": "maya", "A": "walk"},
		[]string{"should_do(P,A)", "deliberating_agent(P,A)"}, nil, nil, []string{"should_be_performed(A)"}},
	{"bridge_from_continuing", "should_be_performed(contract)", map[string]string{"P": "council", "G": "tram", "A": "contract"},
		[]string{"ought_to_continue(P,G)", "continues_trying(A,G)", "deliberating_agent(P,A)"}, []string{"possible_bring_about(P,A)"},
		[]string{"realize_more_effectively(P,G,A)", "bad_outweighs(A,G)"}, []string{"should_be_performed(A)"}},
}

// Each scheme the proposal changes or adds has exactly the formulas it gives
// (so that a premise, an exception or a conclusion lost or added in the
// application fails here), each instance of it concludes when its premises
// and assumptions hold, and each of its exceptions, argued by a critic,
// undercuts it.
func TestPracticalC4Shapes(t *testing.T) {
	for _, x := range c4schemes {
		a := c4(t)
		sch := a.schemeMap[x.scheme]
		if sch == nil {
			t.Errorf("%s is not in the library", x.scheme)
			continue
		}
		for _, f := range []struct {
			what      string
			got, want []string
		}{{"premises", sch.Premises, x.premises}, {"assumptions", sch.Assumptions, x.assumptions},
			{"exceptions", sch.Exceptions, x.exceptions}, {"conclusions", sch.Conclusions, x.conclusions}} {
			if strings.Join(f.got, " ") != strings.Join(f.want, " ") {
				t.Errorf("%s has the %s %v, want %v", x.scheme, f.what, f.got, f.want)
			}
		}
		id := c3arg(t, a, "p", x.scheme, x.b, x.concl)
		var want []string
		for _, e := range x.exceptions {
			want = append(want, c4inst(e, x.b))
		}
		sort.Strings(want)
		if got := c3exceptions(a, id); strings.Join(got, " ") != strings.Join(want, " ") {
			t.Errorf("%s has the instantiated exceptions %v, want %v", x.scheme, got, want)
		}
		for _, p := range a.Graph.Arguments[id].Premises {
			a.AddAssumption(p.Stmt.Id)
		}
		a.Evaluate()
		if got := c3lab(a, x.concl); got != caes.In {
			t.Errorf("%s: its conclusion is %v with its premises assumed", x.scheme, got)
			continue
		}
		for _, ex := range want {
			a := c4(t)
			id := c3arg(t, a, "p", x.scheme, x.b, x.concl)
			for _, p := range a.Graph.Arguments[id].Premises {
				a.AddAssumption(p.Stmt.Id)
			}
			c3argue(t, a, "critic", ex)
			a.Evaluate()
			if c3undercut(a, id) != caes.In || c3lab(a, x.concl) == caes.In {
				t.Errorf("%s: %s argued, undercutter %v, conclusion %v", x.scheme, ex, c3undercut(a, id), c3lab(a, x.concl))
			}
		}
	}
}

// c4inst instantiates a formula's variables with the bindings.
func c4inst(f string, b map[string]string) string {
	name, args := f, ""
	if i := strings.Index(f, "("); i >= 0 {
		name, args = f[:i], f[i+1:len(f)-1]
	} else {
		return f
	}
	parts := strings.Split(args, ",")
	for i, p := range parts {
		if v, ok := b[p]; ok {
			parts[i] = v
		}
	}
	return name + "(" + strings.Join(parts, ",") + ")"
}

// D063: each recommendation to the one addressed concludes the claim the
// consequence schemes conclude, so an appeal to fear for an action meets an
// argument from negative consequences against it, and a threat against an
// action meets an argument from positive consequences for it: alone each
// decides, and at equal weight neither side's claim is in. The conclusion is
// the scheme's own, instantiated (the adapter takes the one it is given), so
// a scheme that concludes a predicate of its own, as each did before
// proposal D022-D093, fails here.
func TestPracticalC4ConclusionsMeetTheOtherSide(t *testing.T) {
	recommendations := []string{"fear", "danger", "threat", "ad_baculum", "pragmatic_alternative", "appeal_to_vanity_ad_populum"}
	for _, x := range c4schemes {
		if !in(recommendations, x.scheme) {
			continue
		}
		act := x.b[c4action(x)]
		for _, opposed := range []bool{false, true} {
			a := c4(t)
			sch := a.schemeMap[x.scheme]
			if sch == nil || len(sch.Conclusions) != 1 {
				t.Fatalf("%s: %v", x.scheme, sch)
			}
			concl := c4inst(sch.Conclusions[0], x.b)
			pro := concl == "should_be_performed("+act+")"
			if !pro && concl != "¬should_be_performed("+act+")" {
				t.Errorf("%s concludes %s, which meets no argument about %s", x.scheme, concl, act)
				break
			}
			id := c3arg(t, a, "p", x.scheme, x.b, concl)
			for _, p := range a.Graph.Arguments[id].Premises {
				a.AddAssumption(p.Stmt.Id)
			}
			if opposed {
				other, oc := "negative_consequences", "¬should_be_performed("+act+")"
				if !pro {
					other, oc = "positive_consequences", "should_be_performed("+act+")"
				}
				c3assume(a, "brings_about("+act+",g)", "good(g)", "bad(g)")
				c3arg(t, a, "o", other, map[string]string{"A": act, "G": "g"}, oc)
			}
			a.Evaluate()
			got, neg := c3lab(a, "should_be_performed("+act+")"), c3lab(a, "¬should_be_performed("+act+")")
			switch {
			case !opposed && pro && got != caes.In, !opposed && !pro && neg != caes.In:
				t.Errorf("%s alone: should=%v ¬should=%v", x.scheme, got, neg)
			case opposed && (got == caes.In || neg == caes.In):
				t.Errorf("%s against an argument of the other side at equal weight: should=%v ¬should=%v (want a tie)", x.scheme, got, neg)
			}
		}
	}
}

func in(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// c4action is the variable of a scheme's recommended action.
func c4action(x c4shape) string {
	c := x.conclusions[0]
	i := strings.Index(c, "(")
	return strings.TrimSuffix(c[i+1:], ")")
}

// D063's bridges: an agent's conclusion reaches the recommendation only with
// deliberating_agent(P,A), the arguer's premise (the refinement: positively
// the same agent, not an exception nobody raises), and the continuing bridge
// only with continues_trying(A,G) and its assumption possible_bring_about(P,A)
// too (D093(e) and (f)).
func TestPracticalC4Bridges(t *testing.T) {
	t.Run("from an agent", func(t *testing.T) {
		for _, deliberating := range []bool{false, true} {
			a := c4(t)
			c3assume(a, "contributing_action(walk,exercise)", "goal_of_agent(maya,exercise)")
			c3arg(t, a, "p", "practical_goal", map[string]string{"P": "maya", "X": "walk", "G": "exercise"}, "should_do(maya,walk)")
			c3arg(t, a, "p", "bridge_from_agent", map[string]string{"P": "maya", "A": "walk"}, "should_be_performed(walk)")
			if deliberating {
				c3assume(a, "deliberating_agent(maya,walk)")
			}
			a.Evaluate()
			if c3lab(a, "should_do(maya,walk)") != caes.In {
				t.Fatalf("the argument from goal did not conclude: %v", c3lab(a, "should_do(maya,walk)"))
			}
			if got := c3lab(a, "should_be_performed(walk)"); (got == caes.In) != deliberating {
				t.Errorf("deliberating_agent %v: the recommendation is %v", deliberating, got)
			}
		}
	})
	t.Run("from continuing", func(t *testing.T) {
		for _, c := range []struct {
			continues, deliberating, possible, want bool
		}{{false, true, true, false}, {true, false, true, false}, {true, true, false, false}, {true, true, true, true}} {
			a := c4(t)
			c3assume(a, "stop_realize_wasted(council,tram)", "waste_bad_thing(council,tram)", "possible_goal(tram)")
			c3arg(t, a, "p", "waste", map[string]string{"A": "council", "X": "tram"}, "ought_to_continue(council,tram)")
			c3arg(t, a, "p", "bridge_from_continuing", map[string]string{"P": "council", "G": "tram", "A": "contract"}, "should_be_performed(contract)")
			if c.continues {
				c3assume(a, "continues_trying(contract,tram)")
			}
			if c.deliberating {
				c3assume(a, "deliberating_agent(council,contract)")
			}
			if c.possible {
				c3assume(a, "possible_bring_about(council,contract)")
			}
			a.Evaluate()
			if got := c3lab(a, "should_be_performed(contract)"); (got == caes.In) != c.want {
				t.Errorf("continues_trying %v, deliberating_agent %v, possible_bring_about %v: the recommendation is %v", c.continues, c.deliberating, c.possible, got)
			}
		}
	})
	// A bridged recommendation meets the other side like any argument: at
	// equal weight, a tie.
	t.Run("a tie", func(t *testing.T) {
		a := c4(t)
		c3assume(a, "contributing_action(walk,exercise)", "goal_of_agent(maya,exercise)", "deliberating_agent(maya,walk)",
			"brings_about(walk,danger)", "bad(danger)")
		c3arg(t, a, "p", "practical_goal", map[string]string{"P": "maya", "X": "walk", "G": "exercise"}, "should_do(maya,walk)")
		c3arg(t, a, "p", "bridge_from_agent", map[string]string{"P": "maya", "A": "walk"}, "should_be_performed(walk)")
		c3arg(t, a, "c", "negative_consequences", map[string]string{"A": "walk", "G": "danger"}, "¬should_be_performed(walk)")
		a.Evaluate()
		if got, neg := c3lab(a, "should_be_performed(walk)"), c3lab(a, "¬should_be_performed(walk)"); got == caes.In || neg == caes.In {
			t.Errorf("should=%v ¬should=%v, want a tie", got, neg)
		}
	})
}

// D023, D033 and D093: each new exception reaches the arguments whose
// variables it names, and no other: the reassessment of one agent's goal
// leaves another goal's argument from waste standing, a conflicting goal of
// one agent leaves another agent's argument from goal, the outweighing of one
// means leaves another's, a better means for one agent leaves another agent's
// advice standing (23/CQ2, D089), a better means than one continuing
// action, or its outweighing side effects, leave the bridge to its rival
// standing (D093(f)), and a bad end's good means is judged alone (D093(g)).
func TestPracticalC4ExceptionsReachOnlyTheirArguments(t *testing.T) {
	type c struct {
		scheme, concl string
		b             map[string]string
		premises      []string
	}
	cases := []struct {
		exception string
		hit, miss c
	}{
		{"reassessment_against_continuing(council,tram)",
			c{"waste", "ought_to_continue(council,tram)", map[string]string{"A": "council", "X": "tram"}, []string{"stop_realize_wasted(council,tram)", "waste_bad_thing(council,tram)", "possible_goal(tram)"}},
			c{"waste", "ought_to_continue(council,museum)", map[string]string{"A": "council", "X": "museum"}, []string{"stop_realize_wasted(council,museum)", "waste_bad_thing(council,museum)", "possible_goal(museum)"}}},
		{"incompatible_goal(maya,exercise)",
			c{"practical_goal", "should_do(maya,walk)", map[string]string{"P": "maya", "X": "walk", "G": "exercise"}, []string{"contributing_action(walk,exercise)", "goal_of_agent(maya,exercise)"}},
			c{"practical_goal", "should_do(jo,walk)", map[string]string{"P": "jo", "X": "walk", "G": "exercise"}, []string{"contributing_action(walk,exercise)", "goal_of_agent(jo,exercise)"}}},
		{"bad_outweighs(train,marathon)",
			c{"two_person_practical", "should_do(lena,train)", map[string]string{"A": "lena", "B": "coach", "X": "marathon", "Y": "train"}, []string{"tells_intend(lena,marathon,coach)", "situation_necessity(coach,train,marathon,lena)", "possible_bring_about(lena,train)"}},
			c{"two_person_practical", "should_do(lena,swim)", map[string]string{"A": "lena", "B": "coach", "X": "marathon", "Y": "swim"}, []string{"tells_intend(lena,marathon,coach)", "situation_necessity(coach,swim,marathon,lena)", "possible_bring_about(lena,swim)"}}},
		{"means_bad_outweighs(cooking,eating_well)",
			c{"end_and_means_1", "good(cooking)", map[string]string{"X": "cooking", "Y": "eating_well"}, []string{"means_to_end(cooking,eating_well)", "good(eating_well)"}},
			c{"end_and_means_1", "good(gardening)", map[string]string{"X": "gardening", "Y": "eating_well"}, []string{"means_to_end(gardening,eating_well)", "good(eating_well)"}}},
		{"realize_more_effectively(lena,marathon,train)",
			c{"two_person_practical", "should_do(lena,train)", map[string]string{"A": "lena", "B": "coach", "X": "marathon", "Y": "train"}, []string{"tells_intend(lena,marathon,coach)", "situation_necessity(coach,train,marathon,lena)", "possible_bring_about(lena,train)"}},
			c{"two_person_practical", "should_do(sam,train)", map[string]string{"A": "sam", "B": "coach", "X": "marathon", "Y": "train"}, []string{"tells_intend(sam,marathon,coach)", "situation_necessity(coach,train,marathon,sam)", "possible_bring_about(sam,train)"}}},
		{"means_good_outweighs(rigging,corrupt_mayor)",
			c{"end_and_means_bad", "bad(rigging)", map[string]string{"X": "rigging", "Y": "corrupt_mayor"}, []string{"means_to_end(rigging,corrupt_mayor)", "bad(corrupt_mayor)"}},
			c{"end_and_means_bad", "bad(bribery)", map[string]string{"X": "bribery", "Y": "corrupt_mayor"}, []string{"means_to_end(bribery,corrupt_mayor)", "bad(corrupt_mayor)"}}},
		{"bad_outweighs(contract_old,tram)",
			c{"bridge_from_continuing", "should_be_performed(contract_old)", map[string]string{"P": "council", "G": "tram", "A": "contract_old"}, []string{"ought_to_continue(council,tram)", "continues_trying(contract_old,tram)", "deliberating_agent(council,contract_old)", "possible_bring_about(council,contract_old)"}},
			c{"bridge_from_continuing", "should_be_performed(contract_new)", map[string]string{"P": "council", "G": "tram", "A": "contract_new"}, []string{"ought_to_continue(council,tram)", "continues_trying(contract_new,tram)", "deliberating_agent(council,contract_new)", "possible_bring_about(council,contract_new)"}}},
		{"realize_more_effectively(council,tram,contract_old)",
			c{"bridge_from_continuing", "should_be_performed(contract_old)", map[string]string{"P": "council", "G": "tram", "A": "contract_old"}, []string{"ought_to_continue(council,tram)", "continues_trying(contract_old,tram)", "deliberating_agent(council,contract_old)", "possible_bring_about(council,contract_old)"}},
			c{"bridge_from_continuing", "should_be_performed(contract_new)", map[string]string{"P": "council", "G": "tram", "A": "contract_new"}, []string{"ought_to_continue(council,tram)", "continues_trying(contract_new,tram)", "deliberating_agent(council,contract_new)", "possible_bring_about(council,contract_new)"}}},
	}
	for _, x := range cases {
		a := c4(t)
		c3assume(a, x.hit.premises...)
		c3assume(a, x.miss.premises...)
		hit := c3arg(t, a, "p1", x.hit.scheme, x.hit.b, x.hit.concl)
		miss := c3arg(t, a, "p2", x.miss.scheme, x.miss.b, x.miss.concl)
		c3argue(t, a, "critic", x.exception)
		a.Evaluate()
		if c3undercut(a, hit) != caes.In || c3undercut(a, miss) == caes.In {
			t.Errorf("%s: undercuts %s %v (want In) and %s %v (want not In)", x.exception, x.hit.concl, c3undercut(a, hit), x.miss.concl, c3undercut(a, miss))
		}
		if c3lab(a, x.miss.concl) != caes.In {
			t.Errorf("%s: %s is %v, want In", x.exception, x.miss.concl, c3lab(a, x.miss.concl))
		}
	}
}
