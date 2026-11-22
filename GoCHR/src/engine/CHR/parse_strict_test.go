package chr

import (
	"testing"

	. "github.com/hfried/GoCHR/src/engine/terms"
)

// Rule text the parser cannot read in full is refused, and the rule is not
// added. The parser used to read "broken(" as broken and "p(X) q(Y)" as p(X).
func TestAddRuleRefusesMalformedText(t *testing.T) {
	CHRtrace = 0
	for _, text := range []string{"broken(", "p(X", "p(X))", "p(X),", "p(X) q(Y)", "p(a,)", "p(X).", "p(X); q", "", "42", "1.5", `"s"`} {
		for _, place := range []string{"head", "body"} {
			rs := MakeRuleStore()
			var err error
			if place == "head" {
				err = rs.AddRule("r", []string{text}, nil, nil, []string{"q(a)"})
			} else {
				err = rs.AddRule("r", []string{"p(X)"}, nil, nil, []string{text})
			}
			if err == nil || len(rs.CHRruleStore) != 0 {
				t.Errorf("%q in the %s: want an error and no rule, got err=%v and %d rule(s)", text, place, err, len(rs.CHRruleStore))
			}
		}
	}
}

// Well-formed text is accepted as before, including a term of the wrong kind
// in a head (a negation; such a rule never fires, PLAN.md 2.4).
func TestAddRuleAcceptsWellFormedText(t *testing.T) {
	CHRtrace = 0
	for _, text := range []string{"p(X)", "p(X), q(Y)", "go", "neg(p(X))", "¬p(X)", `p("a b")`, "argument(s,[X,Y])"} {
		rs := MakeRuleStore()
		if err := rs.AddRule("r", []string{text}, nil, nil, []string{"q(a)"}); err != nil {
			t.Errorf("%q in the head: %v", text, err)
		}
		if err := rs.AddRule("r2", []string{"p(X)"}, nil, nil, []string{text}); err != nil {
			t.Errorf("%q in the body: %v", text, err)
		}
	}
}

// A malformed goal is an error, and nothing is printed.
func TestInferRefusesMalformedGoal(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.AddRule("r", []string{"p(X)"}, nil, nil, []string{"q(X)"})
	var err error
	out := captureStdout(t, func() { _, _, err = rs.Infer([]string{"p(a", "p(b)"}, 100) })
	if err == nil {
		t.Errorf("Infer accepted the goal p(a")
	}
	if out != "" {
		t.Errorf("Infer printed %q", out)
	}
}

// When no rule fires, Infer returns the goals, not an empty store.
func TestNoRuleFiredKeepsGoals(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.AddRule("r", []string{"p(X)"}, nil, nil, []string{"q(X)"})
	ok, store, err := rs.Infer([]string{"r(a)", "s(b)"}, 100)
	if !ok || err != nil || count(store, "r(a)") != 1 || count(store, "s(b)") != 1 {
		t.Errorf("want r(a) and s(b) back, got ok=%v err=%v store=%v", ok, err, store)
	}
}
