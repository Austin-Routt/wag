package validation_test

import (
	"strings"
	"testing"

	cyaml "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/validation"
)

func TestExceptionValidationRejectsNongroundBinding(t *testing.T) {
	// The malformed parameters create X -> Y -> Z -> Y. Validation must
	// reject nonground bindings before substituting an exception; following
	// that variable chain is neither required nor safe for invalid input.
	text := `
language:
  p/1: "%s is a premise."
  q/1: "%s is a conclusion."
  e/1: "%s is an exception."
argument_schemes:
  - id: s
    variables: [X,Y,Z]
    premises: [p(X), p(Y), p(Z)]
    conclusions: [q(X)]
    exceptions: [e(X)]
statements:
  p(a): Premise A.
  p(b): Premise B.
  p(c): Premise C.
  q(a): Conclusion A.
  e(a): Exception A.
  target: Undercutter.
arguments:
  authored:
    scheme: s
    parameters: [Y,Z,Y]
    premises: [p(a), p(b), p(c)]
    conclusion: q(a)
    undercutter: target
  cut:
    premises: [e(a)]
    conclusion: target
`
	ag, err := cyaml.Import(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	if probs := validation.Validate(ag); !exceptionProblemFor(probs, "authored") {
		t.Fatalf("nonground exception bindings accepted: %+v", probs)
	}
}
