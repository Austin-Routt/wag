// Copyright © 2026 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL was not distributed
// with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

package caes_test

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/terms"
)

func infer(t *testing.T, src string) *caes.ArgGraph {
	t.Helper()
	ag, err := yaml.Import(strings.NewReader(src))
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if err := ag.Infer(); err != nil {
		t.Fatalf("Infer: %v", err)
	}
	return ag
}

// argumentsFor lists, sorted, the schemes of the arguments for a conclusion.
func argumentsFor(ag *caes.ArgGraph, conclusion string) []string {
	var ids []string
	for _, a := range ag.Arguments {
		if a != nil && a.Conclusion != nil && a.Conclusion.Id == conclusion {
			ids = append(ids, a.Scheme.Id+"("+strings.Join(a.Parameters, ",")+")")
		}
	}
	sort.Strings(ids)
	return ids
}

// Set semantics for facts (docs/contracts.md, contract 2) removes repeated
// facts, not arguments: two schemes concluding one statement give two
// arguments, and one scheme applied to two facts gives two.
func TestSetSemanticsKeepsDistinctArguments(t *testing.T) {
	ag := infer(t, `
language:
  p/1: "%s is p"
  q/1: "%s is q"
  r/1: "%s is r"
argument_schemes:
  - id: from_p
    variables: [X]
    premises:
      - p(X)
    conclusions:
      - q(X)
  - id: from_r
    variables: [X]
    premises:
      - r(X)
    conclusions:
      - q(X)
assumptions:
  - p(a)
  - r(a)
  - p(b)
`)
	if got, want := argumentsFor(ag, "q(a)"), []string{"from_p(a)", "from_r(a)"}; strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("arguments for q(a): got %v, want %v", got, want)
	}
	if got := argumentsFor(ag, "q(b)"); len(got) != 1 {
		t.Errorf("arguments for q(b): got %v, want one", got)
	}
}

// Two schemes that derive each other's premise stop after one round, each with
// its one argument, and a scheme listed after them still fires (X-LOOP).
func TestMutuallyDerivingSchemesTerminate(t *testing.T) {
	ag := infer(t, `
language:
  part/1: "%s is a part"
  whole/1: "%s is a whole"
  noted/1: "%s is noted"
argument_schemes:
  - id: composition
    variables: [X]
    premises:
      - part(X)
    conclusions:
      - whole(X)
  - id: division
    variables: [X]
    premises:
      - whole(X)
    conclusions:
      - part(X)
  - id: later
    variables: [X]
    premises:
      - part(X)
    conclusions:
      - noted(X)
assumptions:
  - part(a)
`)
	if got := argumentsFor(ag, "whole(a)"); len(got) != 1 {
		t.Errorf("arguments for whole(a): got %v, want one", got)
	}
	if got := argumentsFor(ag, "part(a)"); len(got) != 1 {
		t.Errorf("arguments for part(a): got %v, want one", got)
	}
	if got := argumentsFor(ag, "noted(a)"); len(got) != 1 {
		t.Errorf("the scheme listed last did not fire: arguments for noted(a): %v", got)
	}
}

// tree derives two new terms from every term: within the profile, it reaches
// the rule limit long before the term-size limit.
const tree = `
language:
  t/1: "%s is a term"
argument_schemes:
  - id: left
    variables: [X]
    premises: [t(X)]
    conclusions: [t(l(X))]
  - id: right
    variables: [X]
    premises: [t(X)]
    conclusions: [t(r(X))]
assumptions: [t(a)]
`

const growing = `
language:
  t/1: "%s is a term"
argument_schemes:
  - id: grow
    variables: [X]
    premises:
      - t(X)
    conclusions:
      - t(f(X))
assumptions:
  - t(a)
`

// A limit is an error and leaves the graph as it was (docs/contracts.md,
// contract 3), whether the rule limit or the term size ends the run.
func TestLimitIsAnErrorAndGraphUnchanged(t *testing.T) {
	for name, src := range map[string]string{"tree": tree, "growing term": growing} {
		ag, err := yaml.Import(strings.NewReader(src))
		if err != nil {
			t.Fatalf("%s: import: %v", name, err)
		}
		args, stmts := len(ag.Arguments), len(ag.Statements)
		start := time.Now()
		err = ag.Infer()
		if !errors.Is(err, caes.ErrLimitReached) {
			t.Errorf("%s: Infer returned %v, want an error wrapping ErrLimitReached", name, err)
		}
		if caes.ExitStatus(err) != 4 {
			t.Errorf("%s: exit status %d, want 4", name, caes.ExitStatus(err))
		}
		if len(ag.Arguments) != args || len(ag.Statements) != stmts {
			t.Errorf("%s: the graph changed: %d arguments and %d statements before, %d and %d after",
				name, args, stmts, len(ag.Arguments), len(ag.Statements))
		}
		if d := time.Since(start); d > 20*time.Second {
			t.Errorf("%s: took %v", name, d)
		}
	}
}

// A cancelled context stops argument construction with ErrCancelled.
func TestInferContextCancelled(t *testing.T) {
	ag, err := yaml.Import(strings.NewReader(tree))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = ag.InferContext(ctx)
	if !errors.Is(err, caes.ErrCancelled) || caes.ExitStatus(err) != 5 {
		t.Errorf("InferContext with a cancelled context returned %v (exit %d), want ErrCancelled (exit 5)", err, caes.ExitStatus(err))
	}
	if len(ag.Arguments) != 0 {
		t.Errorf("a cancelled run added %d arguments", len(ag.Arguments))
	}
}

// A scheme the rule engine cannot read makes Infer fail as invalid input,
// naming the scheme, and leaves the graph unchanged (E-PARSE). Before, the
// premise `broken(` was read as `broken` and the scheme silently did nothing.
func TestMalformedSchemeIsInvalidInput(t *testing.T) {
	ag, err := yaml.Import(strings.NewReader(`
language:
  p/1: "%s is p"
  q/1: "%s is q"
argument_schemes:
  - id: fine
    variables: [X]
    premises:
      - p(X)
    conclusions:
      - q(X)
  - id: s15
    variables: [X]
    premises:
      - p(X)
      - "broken("
    conclusions:
      - q(X)
assumptions:
  - p(a)
`))
	if err != nil {
		t.Fatal(err)
	}
	err = ag.Infer()
	if !errors.Is(err, caes.ErrInvalidInput) || caes.ExitStatus(err) != 1 {
		t.Fatalf("Infer returned %v (exit %d), want invalid input (exit 1)", err, caes.ExitStatus(err))
	}
	if !strings.Contains(err.Error(), "s15") {
		t.Errorf("the error does not name the scheme: %v", err)
	}
	if len(ag.Arguments) != 0 {
		t.Errorf("%d argument(s) added", len(ag.Arguments))
	}
}

// The language applies inside a negation (E-NEGTEXT).
func TestNegationText(t *testing.T) {
	l := caes.Language{"p/1": "%s is p.", "goal/1": "The goal is %s."}
	for in, want := range map[string]string{
		"¬p(a)":    "It is not the case that a is p.",
		"¬goal(g)": "It is not the case that the goal is g.",
		"¬z(a)":    "¬z(a)",
	} {
		term, ok := terms.ReadString(in)
		if !ok {
			t.Fatalf("cannot read %s", in)
		}
		if got := l.Apply(term); got != want {
			t.Errorf("%s: got %q, want %q", in, got, want)
		}
	}
}

// Negated premises match, also through a statement variable (X-NEG).
func TestNegatedPremises(t *testing.T) {
	ag := infer(t, `
language:
  says/2: "%s says that %s."
  p/1: "%s is p."
  r/1: "%s is r."
argument_schemes:
  - id: s6a
    variables: [W,S]
    premises: ["says(W,S)"]
    conclusions: [S]
  - id: s6b
    variables: [X]
    premises: ["¬p(X)"]
    conclusions: [r(X)]
assumptions: ["says(ann,¬p(a))", "¬p(b)"]
`)
	for _, c := range []string{"r(a)", "r(b)"} {
		if got := argumentsFor(ag, c); len(got) != 1 {
			t.Errorf("arguments for %s: %v", c, got)
		}
	}
}

// A second Infer adds nothing: arguments are identified by scheme and
// canonical parameters, and undercutters have no scheme (E-IDEMP).
func TestInferTwice(t *testing.T) {
	ag := infer(t, `
language:
  p/2: "%s p %s"
  q/2: "%s q %s"
  r/1: "%s is r"
argument_schemes:
  - id: s
    variables: [X, Y]
    premises:
      - p(X,Y)
    exceptions:
      - r(X)
    conclusions:
      - q(X,Y)
assumptions:
  - p(a,b)
`)
	n := len(ag.Arguments)
	if err := ag.Infer(); err != nil {
		t.Fatal(err)
	}
	if len(ag.Arguments) != n {
		t.Errorf("a second Infer: %d arguments, then %d", n, len(ag.Arguments))
	}
}

// Outside the profile (contract 1): deletions and guards are unsupported, and
// arithmetic in a statement is invalid input, not computed.
func TestProfile(t *testing.T) {
	for _, c := range []struct {
		src  string
		kind error
	}{
		{`
language:
  n/1: "%s is a number"
argument_schemes:
  - id: count
    variables: [X]
    deletions: [n(X)]
    conclusions: [m(X)]
assumptions: [n(0)]
`, caes.ErrUnsupported},
		{`
language:
  p/2: "%s p %s"
  q/2: "%s q %s"
argument_schemes:
  - id: s4
    variables: [P,D]
    premises: ["p(P,D)"]
    conclusions: ["q(P,D)"]
assumptions:
  - p(joe,1990-05-17)
`, caes.ErrInvalidInput},
		{`
language:
  p/1: "%s p"
  q/1: "%s q"
argument_schemes:
  - id: s
    variables: [X]
    premises: ["p(X)"]
    conclusions: ["q(X+1)"]
assumptions: [p(1)]
`, caes.ErrInvalidInput},
	} {
		ag, err := yaml.Import(strings.NewReader(c.src))
		if err != nil {
			t.Fatal(err)
		}
		if err := ag.Infer(); !errors.Is(err, c.kind) {
			t.Errorf("want %v, got %v", c.kind, err)
		}
		if len(ag.Arguments) != 0 {
			t.Errorf("%d arguments added", len(ag.Arguments))
		}
	}
	// a quoted date, a hyphenated name and a negative number are constants
	ag := infer(t, `
language:
  p/2: "%s p %s"
  q/2: "%s q %s"
argument_schemes:
  - id: s4
    variables: [P,D]
    premises: ["p(P,D)"]
    conclusions: ["q(P,D)"]
assumptions:
  - 'p(joe,"1990-05-17")'
  - p(opposition-normative,-1)
  - 'p(joe,"it rains")'
`)
	for _, c := range []string{`q(joe,"1990-05-17")`, "q(opposition-normative,-1)", `q(joe,"it rains")`} {
		if got := argumentsFor(ag, c); len(got) != 1 {
			t.Errorf("arguments for %s: %v", c, got)
		}
	}
}

// Critical questions under `assumptions:` (contract 5): a questioned default
// is not assumed, a stipulation stays assumed, support restores the argument.
func TestQuestioned(t *testing.T) {
	theory := `
language:
  p/1: "%s is p"
  a/1: "%s is a"
  c/1: "%s is c"
  e/1: "%s is e"
argument_schemes:
  - id: s10
    variables: [X]
    premises: [p(X)]
    assumptions: [a(X)]
    conclusions: [c(X)]
  - id: s11
    variables: [X]
    premises: [e(X)]
    conclusions: [a(X)]
`
	for _, c := range []struct {
		graph string
		want  caes.Label
	}{
		{"assumptions: [p(x)]\n", caes.In},
		{"assumptions: [p(x)]\nquestioned: [a(x)]\n", caes.Out},
		{"assumptions: [p(x), a(x)]\nquestioned: [a(x)]\n", caes.In},
		{"assumptions: [p(x), e(x)]\nquestioned: [a(x)]\n", caes.In},
	} {
		ag := infer(t, theory+c.graph)
		l := ag.GroundedLabelling()
		if got := l[ag.Statements["c(x)"]]; got != c.want {
			t.Errorf("%q: c(x) is %v, want %v", c.graph, got, c.want)
		}
	}
}

// A request to instantiate a scheme is carried out (contract 6).
func TestInstantiationRequest(t *testing.T) {
	ag, err := yaml.Import(strings.NewReader(`
language:
  expert/2: "%s is an expert in %s"
  asserts/2: "%s asserts %s"
  untrustworthy/1: "%s is untrustworthy"
  safe/1: "%s is safe"
argument_schemes:
  - id: expert_opinion
    variables: [W,D,S]
    premises:
      - expert(W,D)
      - asserts(W,S)
    exceptions: [untrustworthy(W)]
    conclusions: [S]
arguments:
  r1:
    scheme: expert_opinion
    parameters:
      - ann
      - vaccines
      - safe(vaccine)
assumptions:
  - expert(ann,vaccines)
  - asserts(ann,safe(vaccine))
  - untrustworthy(ann)
`))
	if err != nil {
		t.Fatal(err)
	}
	if got := argumentsFor(ag, "safe(vaccine)"); len(got) != 1 {
		t.Fatalf("arguments for safe(vaccine): %v", got)
	}
	if err := ag.Infer(); err != nil {
		t.Fatal(err)
	}
	if got := argumentsFor(ag, "safe(vaccine)"); len(got) != 1 {
		t.Errorf("after Infer: %v", got)
	}
	l := ag.GroundedLabelling()
	if got := l[ag.Statements["safe(vaccine)"]]; got == caes.In {
		t.Errorf("the instantiated argument is not undercut by its exception: safe(vaccine) is %v", got)
	}
}
