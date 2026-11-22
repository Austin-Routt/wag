package caes

import (
	"context"
	"errors"
	"testing"
)

func TestReviewAlreadyCancelledEmptyTheory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for i := 0; i < 30; i++ {
		ag := NewArgGraph()
		ag.Assumptions = []string{"p(a)"}
		if err := ag.InferContext(ctx); !errors.Is(err, ErrCancelled) {
			t.Fatalf("attempt %d: cancelled empty theory: %v", i, err)
		}
		if len(ag.Statements) != 0 {
			t.Fatal("cancelled run changed the graph")
		}
	}
}

func TestReviewGeneratedIssueRejectsAcceptedPositions(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		ag := NewArgGraph()
		p := IssueScheme{"accepts(X)", "rejects(X)"}
		ag.Theory.IssueSchemes["choice"] = &p
		ag.Assumptions = []string{"accepts(proposal)", "rejects(proposal)"}
		if explicit {
			for _, id := range ag.Assumptions {
				ag.Statements[id] = &Statement{Id: id, Text: id}
			}
		}
		before := len(ag.Statements)
		if err := ag.Infer(); !errors.Is(err, ErrInconsistentContext) {
			t.Errorf("explicit=%v: want inconsistent-context, got %v", explicit, err)
		}
		if len(ag.Statements) != before || len(ag.Issues) != 0 {
			t.Errorf("explicit=%v: failed run changed the graph", explicit)
		}
	}
}

func TestReviewInvalidIssueLeavesGraphUnchanged(t *testing.T) {
	ag := NewArgGraph()
	p := IssueScheme{"p(X)", "q("}
	ag.Theory.IssueSchemes["bad"] = &p
	ag.Assumptions = []string{"p(a)"}
	if err := ag.Infer(); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("want invalid-input, got %v", err)
	}
	if len(ag.Statements) != 0 || len(ag.Issues) != 0 {
		t.Fatal("invalid issue scheme changed the graph")
	}
}

func TestReviewGeneratedIssuePreservesDefaultPriority(t *testing.T) {
	for _, bothDefaults := range []bool{false, true} {
		ag := NewArgGraph()
		p := IssueScheme{"accepts(X)", "rejects(X)"}
		ag.Theory.IssueSchemes["choice"] = &p
		ag.Assumptions = []string{"accepts(proposal)", "rejects(proposal)"}
		ag.Defaults = map[string]bool{"accepts(proposal)": true}
		if bothDefaults {
			ag.Defaults["rejects(proposal)"] = true
		}
		if err := ag.Infer(); err != nil {
			t.Fatalf("default conflict is not inconsistent stipulation: %v", err)
		}
		l := ag.GroundedLabelling()
		if l[ag.Statements["accepts(proposal)"]] == In {
			t.Error("default remained accepted")
		}
		if got := l[ag.Statements["rejects(proposal)"]] == In; got == bothDefaults {
			t.Errorf("contrary accepted=%v with bothDefaults=%v", got, bothDefaults)
		}
	}
}

func TestReviewEmptyTheoryRejectsMalformedAssumption(t *testing.T) {
	ag := NewArgGraph()
	ag.Assumptions = []string{"p("}
	if err := ag.Infer(); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("want invalid-input: %v", err)
	}
	if len(ag.Statements) != 0 {
		t.Fatal("invalid run committed a statement")
	}
}

func TestReviewIssueClosureAndLateFailureAreAtomic(t *testing.T) {
	for _, derived := range []bool{false, true} {
		ag := NewArgGraph()
		p, q := IssueScheme{"p(X)", "r(X)"}, IssueScheme{"r(X)", "q(X)"}
		ag.Theory.IssueSchemes["pr"] = &p
		ag.Theory.IssueSchemes["rq"] = &q
		ag.Assumptions = []string{"p(a)", "q(a)"}
		for _, id := range ag.Assumptions {
			ag.Statements[id] = &Statement{Id: id, Text: "original " + id}
		}
		if derived {
			ag.Theory.ArgSchemes = []*Scheme{{Id: "bridge", Conclusions: []string{"r(a)"}}}
			ag.Theory.InitSchemeIndex()
		} else {
			ag.Statements["r(a)"] = &Statement{Id: "r(a)", Text: "original bridge"}
		}
		originalMap := ag.Statements
		original := ag.Statements["p(a)"]
		before := len(ag.Statements)
		if err := ag.Infer(); !errors.Is(err, ErrInconsistentContext) {
			t.Fatalf("derived=%v: want inconsistent context, got %v", derived, err)
		}
		if len(ag.Issues) != 0 || len(ag.Arguments) != 0 || len(ag.Statements) != before || len(originalMap) != before || ag.Statements["p(a)"] != original || original.Issue != nil || len(original.Args) != 0 || original.Text != "original p(a)" {
			t.Fatalf("derived=%v: failed transaction changed original graph", derived)
		}
	}
}

func TestReviewSuccessfulCommitPreservesPointers(t *testing.T) {
	ag := NewArgGraph()
	p := &Statement{Id: "p(a)", Text: "premise"}
	q := &Statement{Id: "q(a)", Text: "conclusion"}
	ag.Statements[p.Id] = p
	ag.Statements[q.Id] = q
	ag.Assumptions = []string{p.Id}
	ag.Theory.ArgSchemes = []*Scheme{{Id: "derive", Variables: []string{"X"}, Premises: []string{"p(X)"}, Conclusions: []string{"q(X)"}}}
	ag.Theory.InitSchemeIndex()
	if err := ag.Infer(); err != nil {
		t.Fatal(err)
	}
	if ag.Statements[p.Id] != p || ag.Statements[q.Id] != q || len(q.Args) != 1 || q.Args[0].Conclusion != q || q.Args[0].Premises[0].Stmt != p {
		t.Fatal("successful commit replaced existing statement identity or links")
	}
	a := q.Args[0]
	if err := ag.Infer(); err != nil {
		t.Fatal(err)
	}
	if q.Args[0] != a || ag.Arguments[a.Id] != a {
		t.Fatal("second inference replaced existing argument identity")
	}
}

func TestReviewIssueMatchingCanStopWithoutChangingInput(t *testing.T) {
	ag := NewArgGraph()
	p := IssueScheme{"p(X)", "q(X)"}
	ag.Statements["p(a)"] = &Statement{Id: "p(a)"}
	ag.Statements["q(a)"] = &Statement{Id: "q(a)"}
	work := ag.copyForInference()
	err := work.makeIssueWithWork("choice", p, &issueWork{ctx: context.Background(), remaining: 1})
	if !errors.Is(err, ErrLimitReached) || len(ag.Issues) != 0 || ag.Statements["p(a)"].Issue != nil {
		t.Fatalf("issue quota: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = work.makeIssueWithWork("choice", p, &issueWork{ctx: ctx, remaining: 100})
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("issue cancellation: %v", err)
	}
}

func TestReviewDefeatedDefaultCannotSeedAnotherDefault(t *testing.T) {
	ag := NewArgGraph()
	p := IssueScheme{"p", "q"}
	ag.Theory.IssueSchemes["choice"] = &p
	ag.Assumptions = []string{"p", "q"}
	ag.Defaults = map[string]bool{"p": true}
	ag.Theory.ArgSchemes = []*Scheme{
		{Id: "one", Premises: []string{"p"}, Assumptions: []string{"extra"}, Conclusions: []string{"r"}},
		{Id: "two", Premises: []string{"extra"}, Conclusions: []string{"s"}},
	}
	ag.Theory.InitSchemeIndex()
	if err := ag.Infer(); err != nil {
		t.Fatal(err)
	}
	if len(ag.Arguments) != 0 || ag.Statements["extra"] != nil || ag.Statements["s"] != nil {
		t.Fatal("defeated default seeded another scheme's default")
	}
}
