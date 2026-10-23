package registry

import (
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/carneades"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

func scheme(id string, concl, prem, assum, exc []string) *caes.Scheme {
	return &caes.Scheme{Id: id, Conclusions: concl, Premises: prem, Assumptions: assum, Exceptions: exc}
}

// The same pattern is found whatever the variables are called and however
// the premises are ordered; a variant and an extension are told apart.
func TestPatterns(t *testing.T) {
	a := scheme("a", []string{"S"}, []string{"expert(W,D)", "asserts(W,S)", "in_domain(S,D)"}, nil, []string{"biased(W)"})
	b := scheme("b", []string{"P"}, []string{"in_domain(P,F)", "asserts(E,P)", "expert(E,F)"}, nil, []string{"biased(E)"})
	c := scheme("c", []string{"S"}, []string{"position_to_know(W,D)", "asserts(W,S)", "in_domain(S,D)"}, nil, []string{"biased(W)"})
	d := scheme("d", []string{"S"}, []string{"expert(W,D)", "asserts(W,S)", "in_domain(S,D)", "truthful(W)"}, nil, []string{"biased(W)"})
	e := scheme("e", []string{"S"}, []string{"expert(W,D)", "asserts(W,S)"}, []string{"in_domain(S,D)"}, []string{"biased(W)"})
	// Two premises with the same predicate: the numbering must not depend
	// on their order.
	f := scheme("f", []string{"c(X,Y)"}, []string{"p(X,Z)", "p(Z,Y)"}, nil, nil)
	g := scheme("g", []string{"c(A,B)"}, []string{"p(M,B)", "p(A,M)"}, nil, nil)
	if PatternOf(a).Canonical != PatternOf(b).Canonical || PatternOf(f).Canonical != PatternOf(g).Canonical {
		t.Fatalf("the same pattern differs:\n%s\n%s\n%s\n%s", PatternOf(a).Canonical, PatternOf(b).Canonical, PatternOf(f).Canonical, PatternOf(g).Canonical)
	}
	got := map[string]string{}
	for _, r := range Compare([]Pattern{PatternOf(a), PatternOf(b), PatternOf(c), PatternOf(d), PatternOf(e)}) {
		got[r.A+"-"+r.B] = r.Kind
	}
	want := map[string]string{"a-b": "identical", "a-c": "one-formula", "a-d": "extends", "a-e": "roles", "b-c": "one-formula", "b-d": "extends", "b-e": "roles"}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: %q, want %q (all: %v)", k, got[k], v, got)
		}
	}
	fam := &Families{Families: []Family{{ID: "F1", Generic: "a", Members: []Member{{Scheme: "a", Relation: "generic"}, {Scheme: "b", Relation: "alias"},
		{Scheme: "c", Relation: "variant"}, {Scheme: "d", Relation: "specialization"}}}}}
	_, _, findings := Check([]*caes.Scheme{a, b, c, d, e}, fam, false)
	var msgs []string
	for _, f := range findings {
		msgs = append(msgs, f.String())
	}
	// e moves a premise of a to the assumptions: undeclared, with a, b and
	// (close) d.
	if len(findings) != 3 || !strings.Contains(strings.Join(msgs, " "), "a ~ e: roles") || !strings.Contains(strings.Join(msgs, " "), "d ~ e: close") {
		t.Fatalf("findings: %v", msgs)
	}
	for _, f := range findings {
		if f.Error {
			t.Errorf("an error: %v", f)
		}
	}
	_, _, findings = Check([]*caes.Scheme{a, b}, nil, false)
	if len(findings) != 1 || !findings[0].Error {
		t.Fatalf("an undeclared duplicate: %v", findings)
	}
}

// The game's own schemes are all different patterns.
func TestCoreSchemesAreUnique(t *testing.T) {
	ad, err := carneades.NewAdapterFromYAML(data.CoreYAML)
	if err != nil {
		t.Fatal(err)
	}
	_, rels, findings := Check(ad.Graph.Theory.ArgSchemes, nil, false)
	for _, f := range findings {
		if f.Error {
			t.Error(f)
		}
	}
	for _, r := range rels {
		t.Logf("%s %s %s: %s", r.A, r.Kind, r.B, r.Detail)
	}
}

// Distance counts an added formula once and a changed one twice, the
// conclusion included; Relate names an added exception an extension.
func TestDistance(t *testing.T) {
	base := PatternOf(scheme("pos", []string{"should_be_performed(A)"}, []string{"brings_about(A,G)", "good(G)"}, nil, []string{"bad_outweighs(A,G)"}))
	added := PatternOf(scheme("x", []string{"should_be_performed(A)"}, []string{"brings_about(A,G)", "good(G)"}, nil, []string{"bad_outweighs(A,G)", "forbidden(A)"}))
	turned := PatternOf(scheme("y", []string{"¬should_be_performed(A)"}, []string{"brings_about(A,G)", "good(G)"}, nil, []string{"bad_outweighs(A,G)"}))
	if d, _, adds := Distance(base, added); d != 1 || strings.Join(adds, ",") != "exception forbidden/1" {
		t.Errorf("an added exception: %d %v", d, adds)
	}
	if r, ok := Relate(base, added); !ok || r.Kind != "extends" {
		t.Errorf("an added exception relates as %+v %v", r, ok)
	}
	if d, lacks, adds := Distance(base, turned); d != 2 || len(lacks) != 1 || len(adds) != 1 || !strings.HasPrefix(adds[0], "conclusion ¬") {
		t.Errorf("a negated conclusion: %d %v %v", d, lacks, adds)
	}
}
