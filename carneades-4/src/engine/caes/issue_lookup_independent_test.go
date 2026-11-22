package caes

import (
	"context"
	"testing"
)

// Public API users may supply valid ground formula keys with noncanonical
// whitespace. A lookup optimization must preserve the scan's issue positions.
func TestIssueLookupPreservesNoncanonicalKeys(t *testing.T) {
	build := func(lookup bool) string {
		saved := issueLookup
		issueLookup = lookup
		defer func() { issueLookup = saved }()
		ag := NewArgGraph()
		s := IssueScheme{"accepts(X)", "rejects(X)"}
		ag.Theory.IssueSchemes["choice"] = &s
		for _, id := range []string{"accepts( a )", "rejects( a )"} {
			ag.Statements[id] = &Statement{Id: id, Text: id}
		}
		if err := ag.materializeIssues(context.Background()); err != nil {
			t.Fatal(err)
		}
		return describeIssues(ag)
	}
	scan, lookup := build(false), build(true)
	if scan == "" || lookup != scan {
		t.Fatalf("lookup lost valid API positions: scan=%q lookup=%q", scan, lookup)
	}
}

func TestIssueLookupPreservesScanVariants(t *testing.T) {
	tests := []struct {
		name      string
		patterns  IssueScheme
		ids       []string
		positions int
	}{
		{"aliases", IssueScheme{"accepts(X)", "rejects(X)"}, []string{"accepts(a)", "rejects(a)", "rejects( a )"}, 3},
		{"list_spacing", IssueScheme{"accepts(X)", "rejects(X)"}, []string{"accepts([a,b])", "rejects([a, b])"}, 2},
		{"numeric_types", IssueScheme{"accepts(X)", "rejects(X)"}, []string{"accepts(1)", "rejects(1)", "rejects(1.0)"}, 2},
		{"rounded_float_keys", IssueScheme{"accepts(X)", "rejects(X)"}, []string{"accepts(1.0000001)", "rejects(1.0000002)"}, 0},
		{"duplicate_patterns", IssueScheme{"accepts(X)", "rejects(X)", "rejects(X)"}, []string{"accepts(a)", "rejects(a)"}, 2},
		{"open_pattern", IssueScheme{"accepts(X)", "rejects(X)", "abstains(Y)"}, []string{"accepts(a)", "rejects(a)", "abstains(b)"}, 3},
		{"enumeration", IssueScheme{"p(X1)", "...", "p(X2)"}, []string{"p( a )", "p(b)", "p(c)", "q(a)"}, 3},
		{"substituted_double_negation", IssueScheme{"given(X)", "¬X"}, []string{"given(¬p)", "p", "¬p"}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			build := func(lookup bool) (string, int) {
				saved := issueLookup
				issueLookup = lookup
				defer func() { issueLookup = saved }()
				ag := NewArgGraph()
				ag.Theory.IssueSchemes["choice"] = &tc.patterns
				for _, id := range tc.ids {
					ag.Statements[id] = &Statement{Id: id, Text: id}
				}
				if err := ag.materializeIssues(context.Background()); err != nil {
					t.Fatal(err)
				}
				count := 0
				for _, issue := range ag.Issues {
					count += len(issue.Positions)
				}
				return describeIssues(ag), count
			}
			scan, nscan := build(false)
			lookup, nlookup := build(true)
			if scan != lookup || nscan != tc.positions || nlookup != tc.positions {
				t.Fatalf("want %d positions: scan=%q (%d), lookup=%q (%d)", tc.positions, scan, nscan, lookup, nlookup)
			}
		})
	}
}
