package chr

import (
	"errors"
	"testing"

	. "github.com/hfried/GoCHR/src/engine/terms"
)

func propStore(t *testing.T, rules ...[3][]string) *RuleStore {
	t.Helper()
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.Unique = true
	for i, r := range rules {
		if err := rs.AddRule(string(rune('a'+i)), r[0], nil, nil, r[1]); err != nil {
			t.Fatal(err)
		}
	}
	if !isPropagationProgram(rs) {
		t.Fatal("not a propagation program")
	}
	return rs
}

func rule(heads, body []string) [3][]string { return [3][]string{heads, body, nil} }

// Modus ponens with a bare-variable premise, in both orders, with a fact of
// another arity under the premise's functor: one firing, no limit.
func TestPropagationVariableHeadAndArity(t *testing.T) {
	for _, heads := range [][]string{{"implies(A,B)", "A"}, {"A", "implies(A,B)"}} {
		for _, goals := range [][]string{
			{"go", "implies(p,q)", "p", "implies(x,y,z)"},
			{"implies(x,y,z)", "p", "implies(p,q)"},
		} {
			rs := propStore(t, rule(heads, []string{"B", "argument(mp,[A,B])"}))
			ok, store, err := rs.Infer(goals, 50)
			if !ok || err != nil {
				t.Errorf("%v with %v: ok=%v err=%v", heads, goals, ok, err)
				continue
			}
			if count(store, "q") != 1 || count(store, "argument(mp,[p,q])") != 1 {
				t.Errorf("%v with %v: want q and one argument, got %v", heads, goals, store)
			}
		}
	}
}

// A bare variable that no other head binds matches any constraint.
func TestPropagationUnboundVariableHead(t *testing.T) {
	rs := propStore(t, rule([]string{"X", "trigger"}, []string{"seen(X)"}))
	_, store, err := rs.Infer([]string{"p(a)", "q", "trigger"}, 50)
	if err == nil {
		t.Fatal("seen(X) matches X again: the program has no fixpoint and must reach the limit")
	}
	if count(store, "seen(p(a))") != 1 || count(store, "seen(q)") != 1 || count(store, "seen(trigger)") != 0 {
		t.Errorf("got %v", store)
	}
	rs = propStore(t, rule([]string{"X", "trigger"}, []string{"done"}))
	if ok, store, err := rs.Infer([]string{"p(a)", "q", "trigger"}, 50); !ok || err != nil || count(store, "done") != 1 {
		t.Errorf("got ok=%v err=%v store=%v", ok, err, store)
	}
}

// Each rule fires once per tuple: a chain of three rules ends, and every rule
// listed after a pair that derive each other's premise still fires.
func TestPropagationFairAndFinite(t *testing.T) {
	rs := propStore(t,
		rule([]string{"part(X)"}, []string{"whole(X)"}),
		rule([]string{"whole(X)"}, []string{"part(X)"}),
		rule([]string{"part(X)", "r(X)"}, []string{"s(X)"}),
	)
	ok, store, err := rs.Infer([]string{"part(a)", "r(a)"}, 1000)
	if !ok || err != nil || count(store, "whole(a)") != 1 || count(store, "s(a)") != 1 {
		t.Errorf("got ok=%v err=%v store=%v", ok, err, store)
	}
}

// Heads match distinct constraints, as in the general solver.
func TestPropagationDistinctConstraints(t *testing.T) {
	rs := propStore(t, rule([]string{"p(X)", "p(Y)"}, []string{"pair(X,Y)"}))
	_, store, _ := rs.Infer([]string{"p(a)"}, 50)
	if count(store, "pair(a,a)") != 0 {
		t.Errorf("one constraint matched two heads: %v", store)
	}
	_, store, _ = rs.Infer([]string{"p(a)", "p(b)"}, 50)
	if count(store, "pair(a,b)") != 1 || count(store, "pair(b,a)") != 1 {
		t.Errorf("want pair(a,b) and pair(b,a), got %v", store)
	}
}

// The limit is exact: a program needing max applications completes, one
// needing more stops with ErrLimit.
func TestPropagationLimit(t *testing.T) {
	rs := propStore(t, rule([]string{"p(X)"}, []string{"q(X)"}))
	goals := []string{"p(a)", "p(b)", "p(c)"}
	if ok, _, err := rs.Infer(goals, 3); !ok || err != nil {
		t.Errorf("three applications under a limit of three: ok=%v err=%v", ok, err)
	}
	if _, _, err := rs.Infer(goals, 2); !errors.Is(err, ErrLimit) {
		t.Errorf("three applications under a limit of two: %v", err)
	}
	rs = propStore(t, rule([]string{"t(X)"}, []string{"t(f(X))"}), rule([]string{"t(X)"}, []string{"t(g(X))"}))
	if _, _, err := rs.Infer([]string{"t(a)"}, 1000); !errors.Is(err, ErrLimit) {
		t.Errorf("a binary tree of terms: %v", err)
	}
}

// Rules with a deleted head or a guard keep the general solver.
func TestPropagationOnlyForItsProfile(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.Unique = true
	rs.AddRule("g", []string{"p(X)"}, nil, []string{"X > 1"}, []string{"q(X)"})
	if isPropagationProgram(rs) {
		t.Error("a guard does not leave the profile")
	}
	rs = MakeRuleStore()
	rs.AddRule("p", []string{"p(X)"}, nil, nil, []string{"q(X)"})
	if isPropagationProgram(rs) {
		t.Error("multiset semantics do not leave the general solver")
	}
}
