package validation_test

// From pull request #17 (the second review's follow-up), kept with engine walton.5.

import (
	"strings"
	"testing"

	cyaml "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/validation"
)

// A scheme that concludes ¬S for a negative S concludes S (docs/contracts.md,
// contract 4). The argument inference builds is an instance of its scheme,
// and so is the same argument written by hand; the exception it carries is
// the folded statement too.
func TestFoldedInstanceIsAnInstanceOfItsScheme(t *testing.T) {
	theory := `
language:
  denies/1: "It is denied that %s."
  withdrawn/1: "The denial of %s is withdrawn."
  q/1: "%s is q."
argument_schemes:
  - id: deny
    variables: [S]
    premises: [denies(S)]
    exceptions: ["¬S"]
    conclusions: ["¬S"]
issue_schemes:
  negation: [S, "¬S"]
`
	// built by inference
	ag, err := cyaml.Import(strings.NewReader(theory + "assumptions: [\"denies(¬q(a))\"]\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ag.Infer(); err != nil {
		t.Fatal(err)
	}
	if ag.Statements["q(a)"] == nil || ag.Statements["¬¬q(a)"] != nil {
		t.Fatalf("inference did not name the conclusion q(a)")
	}
	for _, p := range validation.Validate(ag) {
		if p.Category == validation.ARGUMENT {
			t.Errorf("the argument inference built is reported: %s: %s %s", p.Id, p.Description, p.Expression)
		}
	}
	// written by hand, with a conclusion that is not the folded instance
	wrong := theory + `
statements:
  denies(¬q(a)): The denial.
  q(a): Q.
  q(b): Another.
  cut: Undercutter.
assumptions: ["denies(¬q(a))"]
arguments:
  authored:
    scheme: deny
    parameters: ["¬q(a)"]
    premises: ["denies(¬q(a))"]
    conclusion: q(b)
    undercutter: cut
  link:
    premises: [q(a)]
    conclusion: cut
`
	ag, err = cyaml.Import(strings.NewReader(wrong))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range validation.Validate(ag) {
		if p.Category == validation.ARGUMENT && p.Id == "authored" && strings.Contains(p.Description, "conclusion does not instantiate") {
			found = true
		}
	}
	if !found {
		t.Errorf("an argument for q(b) passed as an instance of ¬S for S = ¬q(a)")
	}
	ag, err = cyaml.Import(strings.NewReader(strings.Replace(wrong, "conclusion: q(b)", "conclusion: q(a)", 1)))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range validation.Validate(ag) {
		if p.Category == validation.ARGUMENT {
			t.Errorf("the hand-written folded instance is reported: %s: %s %s", p.Id, p.Description, p.Expression)
		}
	}
}
