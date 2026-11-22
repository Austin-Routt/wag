package chr

import (
	"testing"

	. "github.com/hfried/GoCHR/src/engine/terms"
)

func negStore(t *testing.T, heads, body []string) *RuleStore {
	t.Helper()
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.Unique = true
	rs.ExplicitNegation = true
	if err := rs.AddRule("r", heads, nil, nil, body); err != nil {
		t.Fatal(err)
	}
	return rs
}

// A negated head matches a negated constraint, given as a goal or derived
// through a variable; ¬¬S is S; neg/1 is an ordinary predicate.
func TestExplicitNegation(t *testing.T) {
	rs := negStore(t, []string{"¬p(X)"}, []string{"r(X)"})
	if _, store, err := rs.Infer([]string{"¬p(a)", "neg(p(b))"}, 100); err != nil || count(store, "r(a)") != 1 || count(store, "r(b)") != 0 {
		t.Errorf("negated goal: %v %v", store, err)
	}

	rs = negStore(t, []string{"says(W,S)"}, []string{"S"})
	rs.AddRule("n", []string{"¬p(X)"}, nil, nil, []string{"r(X)"})
	if _, store, err := rs.Infer([]string{"says(ann,¬p(a))"}, 100); err != nil || count(store, "r(a)") != 1 {
		t.Errorf("through a statement variable: %v %v", store, err)
	}

	rs = negStore(t, []string{"q(X)"}, []string{"¬¬p(X)"})
	rs.AddRule("n", []string{"p(X)"}, nil, nil, []string{"r(X)"})
	if _, store, err := rs.Infer([]string{"q(a)"}, 100); err != nil || count(store, "p(a)") != 1 || count(store, "r(a)") != 1 {
		t.Errorf("double negation: %v %v", store, err)
	}

	rs = negStore(t, []string{"¬¬p(X)"}, []string{"r(X)"})
	if _, store, err := rs.Infer([]string{"p(a)"}, 100); err != nil || count(store, "r(a)") != 1 {
		t.Errorf("double negation in a head: %v %v", store, err)
	}
}

// Without the flag ¬ stays a built-in operator.
func TestNegationOffIsBuiltIn(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.Unique = true
	rs.AddRule("r", []string{"¬p(X)"}, nil, nil, []string{"r(X)"})
	if _, store, _ := rs.Infer([]string{"¬p(a)"}, 100); count(store, "r(a)") != 0 {
		t.Errorf("without ExplicitNegation a negated head matched: %v", store)
	}
}
