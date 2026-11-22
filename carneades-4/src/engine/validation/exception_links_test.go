package validation_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
	cyaml "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/terms"
	"github.com/carneades/carneades-4/src/engine/validation"
)

func exceptionGraph(t *testing.T, params, exceptions, links string) *caes.ArgGraph {
	t.Helper()
	text := `
language:
  p/1: "%s is a premise."
  q/1: "%s is a conclusion."
  e/1: "%s is an exception."
  f/1: "%s is another exception."
argument_schemes:
  - id: s
    variables: [X]
    premises: [p(X)]
    conclusions: [q(X)]
    exceptions: ` + exceptions + `
statements:
  p(a): Premise A.
  q(a): Conclusion A.
  e(a): Exception A.
  e(b): Exception B.
  f(a): Other exception A.
  guard: An extra condition.
  target: Target undercutter.
  unrelated: Another undercutter.
arguments:
  authored:
    scheme: s
` + params + `    premises: [p(a)]
    conclusion: q(a)
    undercutter: target
` + links + `
assumptions: [p(a)]
`
	ag, err := cyaml.Import(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	return ag
}

func exceptionLink(id, premise, conclusion string) string {
	return fmt.Sprintf("  %s:\n    premises: [%s]\n    conclusion: %s\n", id, premise, conclusion)
}

func exceptionProblemFor(probs []validation.Problem, id string) bool {
	for _, p := range probs {
		if p.Category == validation.ARGUMENT && p.Id == id && strings.Contains(p.Description, "exceptions") {
			return true
		}
	}
	return false
}

func TestAuthoredArgumentsCarryEveryInstantiatedException(t *testing.T) {
	cases := []struct {
		name, exceptions, links string
		valid                   bool
	}{
		{"inert-undercutter", "[e(X)]", "", false},
		{"valid-direct-link", "[e(X)]", exceptionLink("cut", "e(a)", "target"), true},
		{"wrong-instance", "[e(X)]", exceptionLink("cut", "e(b)", "target"), false},
		{"wrong-target", "[e(X)]", exceptionLink("cut", "e(a)", "unrelated"), false},
		{"extra-prerequisite", "[e(X)]", exceptionLink("cut", "e(a), guard", "target"), false},
		{"missing-second-exception", "[e(X), f(X)]", exceptionLink("cut", "e(a)", "target"), false},
		{"all-exceptions", "[e(X), f(X)]", exceptionLink("cut1", "e(a)", "target") + exceptionLink("cut2", "f(a)", "target"), true},
	}
	for _, params := range []string{"    parameters: [a]\n", ""} {
		mode := "explicit"
		if params == "" {
			mode = "inferred"
		}
		for _, tc := range cases {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				ag := exceptionGraph(t, params, tc.exceptions, tc.links)
				beforeArgs, beforeStmts := len(ag.Arguments), len(ag.Statements)
				beforeParams := strings.Join(ag.Arguments["authored"].Parameters, ",")
				beforeCut := ag.Arguments["authored"].Undercutter
				probs := validation.Validate(ag)
				if tc.valid && len(probs) != 0 {
					t.Fatalf("complete authored argument rejected: %+v", probs)
				}
				if !tc.valid && !exceptionProblemFor(probs, "authored") {
					t.Fatalf("missing/misbound exception link accepted: %+v", probs)
				}
				if beforeArgs != len(ag.Arguments) || beforeStmts != len(ag.Statements) || beforeParams != strings.Join(ag.Arguments["authored"].Parameters, ",") || beforeCut != ag.Arguments["authored"].Undercutter {
					t.Fatal("validation rewrote an authored graph")
				}
			})
		}
	}
}

func TestExceptionLinkCannotUseDisabledWeight(t *testing.T) {
	ag := exceptionGraph(t, "    parameters: [a]\n", "[e(X)]", exceptionLink("cut", "e(a)", "target"))
	ag.Arguments["cut"].Scheme = &caes.Scheme{Id: "disabled", Weight: func(*caes.Argument, caes.Labelling) float64 { return 0 }}
	if probs := validation.Validate(ag); !exceptionProblemFor(probs, "authored") {
		t.Fatalf("a zero-weight exception link was accepted: %+v", probs)
	}
}

func TestExceptionLinkMustBelongToGraph(t *testing.T) {
	ag := exceptionGraph(t, "    parameters: [a]\n", "[e(X)]", exceptionLink("cut", "e(a)", "target"))
	delete(ag.Arguments, "cut")
	if probs := validation.Validate(ag); !exceptionProblemFor(probs, "authored") {
		t.Fatalf("a dangling exception link was accepted: %+v", probs)
	}
}

func TestInferredExceptionBindingsBacktrack(t *testing.T) {
	// Both premise permutations match q. Only X=b,Y=a matches the exception
	// links. A validator that takes the first premise substitution rejects
	// this complete graph; one that ignores binding consistency accepts its
	// deliberately altered counterpart below.
	text := `
language:
  p/1: "%s is a premise."
  e/1: "%s is an exception."
  f/1: "%s is another exception."
  q/0: "Conclusion."
argument_schemes:
  - id: symmetric
    variables: [X,Y]
    premises: [p(X), p(Y)]
    conclusions: [q]
    exceptions: [e(X), f(Y)]
statements:
  p(a): Premise A.
  p(b): Premise B.
  e(b): Exception B.
  f(a): Other exception A.
  f(b): Other exception B.
  q: Conclusion.
  target: Undercutter.
arguments:
  authored:
    scheme: symmetric
    premises: [p(a), p(b)]
    conclusion: q
    undercutter: target
  cut1:
    premises: [e(b)]
    conclusion: target
  cut2:
    premises: [f(a)]
    conclusion: target
`
	ag, err := cyaml.Import(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	if probs := validation.Validate(ag); len(probs) != 0 {
		t.Fatalf("valid alternative substitution rejected: %+v", probs)
	}
	ag.Arguments["cut2"].Premises[0].Stmt = ag.Statements["f(b)"]
	if probs := validation.Validate(ag); !exceptionProblemFor(probs, "authored") {
		t.Fatalf("exceptions using incompatible substitutions accepted: %+v", probs)
	}
}

func TestCompleteAuthoredExceptionsDefeatWithoutRewrite(t *testing.T) {
	for _, params := range []string{"    parameters: [a]\n", ""} {
		for _, shown := range []bool{false, true} {
			ag := exceptionGraph(t, params, "[e(X)]", exceptionLink("cut", "e(a)", "target"))
			if shown {
				ag.Assumptions = append(ag.Assumptions, "e(a)")
			}
			if probs := validation.Validate(ag); len(probs) != 0 {
				t.Fatalf("complete argument rejected: %+v", probs)
			}
			if err := ag.Infer(); err != nil {
				t.Fatal(err)
			}
			l := ag.GroundedLabelling()
			if got := l[ag.Statements["q(a)"]] == caes.In; got == shown {
				t.Fatalf("exception shown %v, conclusion In %v", shown, got)
			}
		}
	}
}

func TestGeneratedExceptionStructureSurvivesExport(t *testing.T) {
	// Multiple conclusions share their undercutter; multiple exceptions each
	// have a direct argument. Compound parameters and explicit negation must
	// still be bound correctly after an export/import cycle.
	text := `
language:
  p/1: "%s is a premise."
  d/1: "%s is a default."
  q/1: "%s is a conclusion."
  r/1: "%s is another conclusion."
  e/1: "%s is an exception."
  f/1: "%s is another exception."
argument_schemes:
  - id: s
    variables: [X]
    premises: [p(X)]
    assumptions: [d(X)]
    conclusions: [q(X), r(X)]
    exceptions: [e(X), "¬f(X)"]
assumptions:
  - p(pair(a,b))
  - e(pair(a,b))
`
	ag, err := cyaml.Import(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	if err := ag.Infer(); err != nil {
		t.Fatal(err)
	}
	if probs := validation.Validate(ag); len(probs) != 0 {
		t.Fatalf("complete generated structure rejected: %+v", probs)
	}
	var b bytes.Buffer
	cyaml.Export(&b, ag)
	reloaded, err := cyaml.Import(bytes.NewReader(b.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if probs := validation.Validate(reloaded); len(probs) != 0 {
		t.Fatalf("complete exported structure rejected: %+v", probs)
	}
	if err := reloaded.Infer(); err != nil {
		t.Fatal(err)
	}
	l := reloaded.GroundedLabelling()
	for _, name := range []string{"q(pair(a,b))", "r(pair(a,b))"} {
		st, ok := reloaded.Statements[terms.Normalize(name)]
		if !ok {
			t.Fatalf("no conclusion %s after round trip", name)
		}
		if l[st] == caes.In {
			t.Fatalf("exception did not defeat %s after round trip", name)
		}
	}
}
