package caes_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
	cyaml "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/validation"
)

func readNegationGraph(t *testing.T, text string) *caes.ArgGraph {
	t.Helper()
	ag, err := cyaml.Import(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	return ag
}

func TestGeneratedNegationIdentityAndException(t *testing.T) {
	for _, shown := range []bool{false, true} {
		extra := ""
		if shown {
			extra = ", p"
		}
		ag := readNegationGraph(t, `
language:
  given/1: "Evidence %s."
  p/0: "Proposition p."
  q/0: "Conclusion q."
argument_schemes:
  - id: s
    variables: [X]
    premises: [given(X)]
    conclusions: [q]
    exceptions: ["¬X"]
assumptions: ["given(¬p)"`+extra+`]
`)
		if err := ag.Infer(); err != nil {
			t.Fatal(err)
		}
		if problems := validation.Validate(ag); len(problems) != 0 {
			t.Fatalf("generated structure invalid: %+v", problems)
		}
		if _, exists := ag.Statements["¬¬p"]; exists {
			t.Fatal("exception retained a separate double-negated identity")
		}
		labels := ag.GroundedLabelling()
		if got := labels[ag.Statements["q"]] == caes.In; got == shown {
			t.Fatalf("shown=%v: q in=%v", shown, got)
		}
		var b bytes.Buffer
		cyaml.Export(&b, ag)
		reloaded := readNegationGraph(t, b.String())
		if err := reloaded.Infer(); err != nil {
			t.Fatal(err)
		}
		if got := reloaded.GroundedLabelling()[reloaded.Statements["q"]] == caes.In; got == shown {
			t.Fatal("round trip changed exception result")
		}
	}
}

func TestGeneratedNegationConclusionChains(t *testing.T) {
	ag := readNegationGraph(t, `
language:
  given/1: "Evidence %s."
  p/0: "Proposition p."
  r/0: "Downstream result."
argument_schemes:
  - id: first
    variables: [X]
    premises: [given(X)]
    conclusions: ["¬X"]
  - id: second
    premises: [p]
    conclusions: [r]
assumptions: ["given(¬p)"]
`)
	if err := ag.Infer(); err != nil {
		t.Fatal(err)
	}
	if problems := validation.Validate(ag); len(problems) != 0 {
		t.Fatalf("generated conclusion rejected: %+v", problems)
	}
	l := ag.GroundedLabelling()
	for _, id := range []string{"p", "r"} {
		if l[ag.Statements[id]] != caes.In {
			t.Fatalf("%s did not receive the canonical conclusion's support", id)
		}
	}
}

func TestAuthoredNegationIdentity(t *testing.T) {
	for _, parameters := range []string{"    parameters: [\"¬p\"]\n", ""} {
		for _, conclusion := range []string{"q", "¬X"} {
			actual := "q"
			if conclusion == "¬X" {
				actual = "p"
			}
			text := fmt.Sprintf(`
language:
  given/1: "Evidence %%s."
  p/0: "Proposition p."
  q/0: "Conclusion q."
argument_schemes:
  - id: s
    variables: [X]
    premises: [given(X)]
    conclusions: ["%s"]
    exceptions: ["¬X"]
statements:
  given(¬p): Evidence.
  p: Proposition p.
  q: Conclusion q.
  target: Undercutter.
arguments:
  authored:
    scheme: s
%s    premises: ["given(¬p)"]
    conclusion: %s
    undercutter: target
  cut:
    premises: [p]
    conclusion: target
`, conclusion, parameters, actual)
			ag := readNegationGraph(t, text)
			if problems := validation.Validate(ag); len(problems) != 0 {
				t.Fatalf("parameters %q, conclusion %q: %+v", parameters, conclusion, problems)
			}
			ag.Arguments["cut"].Premises[0].Stmt = ag.Statements["given(¬p)"]
			if problems := validation.Validate(ag); len(problems) == 0 {
				t.Fatal("an unrelated witness was accepted")
			}
		}
	}
}

func TestAuthoredNegatedPremiseOrder(t *testing.T) {
	for _, patterns := range []string{"[\"¬X\", given(X)]", "[given(X), \"¬X\"]"} {
		for _, premises := range []string{"[p, \"given(¬p)\"]", "[\"given(¬p)\", p]"} {
			for _, params := range []string{"", "    parameters: [\"¬p\"]\n"} {
				text := fmt.Sprintf(`
language:
  given/1: "Evidence %%s."
  e/1: "Exception %%s."
  p/0: "Proposition p."
  q/0: "Conclusion q."
argument_schemes:
  - id: s
    variables: [X]
    premises: %s
    conclusions: [q]
    exceptions: [e(X)]
statements:
  given(¬p): Evidence.
  p: Proposition p.
  q: Conclusion q.
  e(¬p): Correct witness.
  e(p): Wrong witness.
  target: Undercutter.
arguments:
  authored:
    scheme: s
%s    premises: %s
    conclusion: q
    undercutter: target
  cut:
    premises: ["e(¬p)"]
    conclusion: target
`, patterns, params, premises)
				ag := readNegationGraph(t, text)
				if problems := validation.Validate(ag); len(problems) != 0 {
					t.Fatalf("patterns=%s premises=%s params=%q: %+v", patterns, premises, params, problems)
				}
				ag.Arguments["cut"].Premises[0].Stmt = ag.Statements["e(p)"]
				if problems := validation.Validate(ag); len(problems) == 0 {
					t.Fatal("premise permutation accepted a mismatched exception witness")
				}
			}
		}
	}
}

func TestNormalizeAPIGraphPreservesIdentityAndAtomicity(t *testing.T) {
	ag := caes.NewArgGraph()
	original := &caes.Statement{Id: "¬¬p", Text: "Original text."}
	ag.Statements[original.Id] = original
	ag.Assumptions = []string{"¬¬p"}
	ag.ExpectedLabeling["¬¬p"] = caes.In
	if err := ag.Infer(); err != nil {
		t.Fatal(err)
	}
	if ag.Statements["p"] != original || original.Id != "p" || original.Text != "Original text." {
		t.Fatal("normalization lost the existing node identity or text")
	}
	if ag.GroundedLabelling()[original] != caes.In {
		t.Fatal("normalized assumption not in")
	}
	if ag.ExpectedLabeling["p"] != caes.In || len(ag.ExpectedLabeling) != 1 {
		t.Fatal("expected labels lost the normalized identity")
	}
	bad := caes.NewArgGraph()
	a := &caes.Statement{Id: "p", Text: "First."}
	b := &caes.Statement{Id: "¬¬p", Text: "Second."}
	bad.Statements["p"] = a
	bad.Statements["¬¬p"] = b
	if err := bad.Infer(); !errors.Is(err, caes.ErrInvalidInput) {
		t.Fatalf("ambiguous distinct nodes accepted: %v", err)
	}
	if len(bad.Statements) != 2 || bad.Statements["p"] != a || bad.Statements["¬¬p"] != b || b.Id != "¬¬p" {
		t.Fatal("failed normalization changed the caller's graph")
	}
	if problems := validation.Validate(bad); len(problems) == 0 {
		t.Fatal("validator failed to report colliding identities")
	}
}

func TestNormalizedContextAndImport(t *testing.T) {
	ag := caes.NewArgGraph()
	ag.Assumptions = []string{"¬¬p", "¬p"}
	if err := ag.Infer(); !errors.Is(err, caes.ErrInconsistentContext) {
		t.Fatalf("contradictory normalized context accepted: %v", err)
	}
	if len(ag.Statements) != 0 {
		t.Fatal("failed context changed the graph")
	}
	if _, err := cyaml.Import(strings.NewReader("statements:\n  p: First.\n  ¬¬p: Second.\n")); err == nil {
		t.Fatal("import silently overwrote a normalized statement collision")
	}
	if _, err := cyaml.Import(strings.NewReader("tests:\n  in: [p]\n  out: [\"¬¬p\"]\n")); err == nil {
		t.Fatal("import silently overwrote conflicting normalized expected labels")
	}
	alias := readNegationGraph(t, "statements:\n  ¬¬p: Proposition p.\nassumptions: [p]\ntests:\n  in: [\"¬¬p\"]\n")
	if problems := validation.Validate(alias); len(problems) != 0 {
		t.Fatalf("expected-label identity rejected: %+v", problems)
	}
	if alias.ExpectedLabeling["p"] != caes.In {
		t.Fatal("import did not normalize expected-label identity")
	}
	for _, stipulated := range []bool{false, true} {
		text := "defaults: [\"¬¬p\"]\nquestioned: [p]\n"
		if stipulated {
			text += "assumptions: [p]\n"
		}
		ag := readNegationGraph(t, text)
		if err := ag.Infer(); err != nil {
			t.Fatal(err)
		}
		if got := ag.GroundedLabelling()[ag.Statements["p"]] == caes.In; got != stipulated {
			t.Fatalf("stipulated=%v: normalized default/stipulation got in=%v", stipulated, got)
		}
	}
}
