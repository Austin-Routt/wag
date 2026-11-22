package caes

import (
	"context"
	"errors"
	"testing"
)

// The first pattern fixes an anchor's bindings. Each other pattern describes
// alternative positions under those bindings; it is not a sequential join that
// commits to the first fresh-variable value encountered in a map iteration.
func TestIssueLaterVariablesAreLocalToEachCandidate(t *testing.T) {
	cases := []struct {
		name     string
		patterns IssueScheme
		ids      []string
		want     string
	}{
		{
			"fresh-variable",
			IssueScheme{"leads(X)", "trails(Y)"},
			[]string{"leads(x)", "trails(a)", "trails(b)", "trails(c)"},
			"leads(x) trails(a) trails(b) trails(c)",
		},
		{
			"anchor-bindings-retained",
			IssueScheme{"leads(X)", "trails(X,Y)"},
			[]string{"leads(a)", "leads(b)", "trails(a,u)", "trails(a,v)", "trails(b,u)", "trails(b,v)", "trails(c,u)"},
			"leads(a) trails(a,u) trails(a,v)\nleads(b) trails(b,u) trails(b,v)",
		},
		{
			"fresh-name-shared-by-alternative-patterns",
			IssueScheme{"leads(X)", "first(X,Y)", "second(X,Y)"},
			[]string{"leads(a)", "first(a,u)", "first(a,v)", "second(a,v)", "second(a,w)", "second(b,w)"},
			"first(a,u) first(a,v) leads(a) second(a,v) second(a,w)",
		},
		{
			"within-pattern-repetition-still-constrains",
			IssueScheme{"leads(X)", "pair(X,Y,Y)"},
			[]string{"leads(a)", "pair(a,u,u)", "pair(a,v,v)", "pair(a,u,v)", "pair(b,u,u)"},
			"leads(a) pair(a,u,u) pair(a,v,v)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, lookup := range []bool{false, true} {
				for order := 0; order < len(tc.ids); order++ {
					saved := issueLookup
					issueLookup = lookup
					ag := NewArgGraph()
					ag.Theory.IssueSchemes["choice"] = &tc.patterns
					for n := range tc.ids {
						id := tc.ids[(n+order)%len(tc.ids)]
						ag.Statements[id] = &Statement{Id: id, Text: id}
					}
					err := ag.materializeIssues(context.Background())
					issueLookup = saved
					if err != nil || describeIssues(ag) != tc.want {
						t.Fatalf("lookup=%v order=%d: err=%v\ngot %q\nwant %q", lookup, order, err, describeIssues(ag), tc.want)
					}
					if err := ag.Infer(); err != nil {
						t.Fatal(err)
					}
					if got := describeIssues(ag); got != tc.want {
						t.Fatalf("another inference changed issue closure: %q", got)
					}
				}
			}
		})
	}
}

// Every stipulated alternative must participate in the consistency check,
// including one that would lose a map-order race to another fresh binding.
func TestIssueLaterAlternativeCannotEscapeConsistency(t *testing.T) {
	for _, stipulated := range []string{"trails(a)", "trails(b)", "trails(c)"} {
		ag := NewArgGraph()
		patterns := IssueScheme{"leads(X)", "trails(Y)"}
		ag.Theory.IssueSchemes["choice"] = &patterns
		for _, id := range []string{"leads(x)", "trails(a)", "trails(b)", "trails(c)"} {
			ag.Statements[id] = &Statement{Id: id, Text: id}
		}
		ag.Assumptions = []string{"leads(x)", stipulated}
		if err := ag.Infer(); !errors.Is(err, ErrInconsistentContext) {
			t.Fatalf("stipulated alternative %s escaped its issue: %v", stipulated, err)
		}
		if len(ag.Issues) != 0 {
			t.Fatal("refusal published a partial issue")
		}
	}
}
