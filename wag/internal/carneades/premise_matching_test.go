package carneades

import (
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

func TestPremiseMatching(t *testing.T) {
	a := NewAdapter()
	a.Graph.Theory.ArgSchemes = []*caes.Scheme{{
		Id:          "pr",
		Variables:   []string{"A", "G"},
		Premises:    []string{"wants(G)", "achieves(A,G)"},
		Assumptions: []string{"possible(A)"},
		Conclusions: []string{"should(A)"},
	}}
	a.Graph.Theory.InitSchemeIndex()
	a.buildSchemeMap()
	ids, err := a.AddSupport("p", "pr", map[string]string{"A": "tax", "G": "clean_air"}, 0.7)
	if err != nil {
		t.Fatal(err)
	}
	main := ids[0]
	for pattern, want := range map[string]string{
		"possible({A})":     "possible(tax)",
		"achieves({A},{G})": "achieves(tax,clean_air)",
		"possible(subsidy)": "possible(tax)", // no exact match: the unique possible/1 premise
		"achieves({A})":     "",              // wrong arity
		"unknown({A})":      "",
		"":                  "",
	} {
		if got := a.PremiseMatching(main, pattern); got != want {
			t.Errorf("PremiseMatching(%q) = %q, want %q", pattern, got, want)
		}
	}
	if got := a.PremiseMatching("nope", "possible({A})"); got != "" {
		t.Errorf("unknown argument matched %q", got)
	}

	// An argument without scheme bindings matches by predicate only if that is unambiguous.
	for _, id := range []string{"q(a)", "q(b)", "r(a)", "c"} {
		a.AddStatement(id, id)
	}
	direct, err := a.AddArgumentDirect("p", "x", []string{"q(a)", "q(b)", "r(a)"}, "c", 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if got := a.PremiseMatching(direct, "r({X})"); got != "r(a)" {
		t.Errorf("unique predicate: got %q", got)
	}
	if got := a.PremiseMatching(direct, "q({X})"); got != "" {
		t.Errorf("ambiguous predicate matched %q", got)
	}
}
