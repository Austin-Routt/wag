package carneades

import (
	"errors"
	"os"
	"strings"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	caesYaml "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/validation"
	"github.com/yourorg/wag/data"
)

// The critical review of 2026-11-23 (probes/adapter_admission_probe_test.go)
// showed an authored argument whose premise is not its scheme's premise, and
// the assumptions p and ¬p, labelled In. Its probe passed by reproducing that
// acceptance; these tests reverse it: such graphs are refused, not labelled.

const probeTheory = `language:
  p/1: "P %s"
  q/1: "Q %s"
  r/1: "R %s"
argument_schemes:
  - id: s
    variables: [X]
    premises: [p(X)]
    conclusions: [r(X)]
`

// admissionError returns the *AdmissionError of a load, failing the test if
// the graph was admitted or refused for another reason.
func admissionError(t *testing.T, yml string) *AdmissionError {
	t.Helper()
	a, err := NewAdapterFromYAML([]byte(yml))
	if err == nil {
		labels, _ := a.Evaluate()
		t.Fatalf("admitted, labelled %v", labels)
	}
	var ae *AdmissionError
	if !errors.As(err, &ae) {
		t.Fatalf("refused for another reason: %v", err)
	}
	// The reader is the same door.
	if _, err2 := NewAdapterFromYAMLReader(strings.NewReader(yml)); err2 == nil || err2.Error() != err.Error() {
		t.Fatalf("the reader: %v, the bytes: %v", err2, err)
	}
	return ae
}

// caesImport is the engine's import alone, without admission.
func caesImport(yml string) (*caes.ArgGraph, error) { return caesYaml.Import(strings.NewReader(yml)) }

func hasProblem(ps []validation.Problem, cat validation.Category, id, description string) bool {
	for _, p := range ps {
		if p.Category == cat && p.Id == id && strings.Contains(p.Description, description) {
			return true
		}
	}
	return false
}

func TestAuthoredArgumentWithTheWrongPremiseIsRefused(t *testing.T) {
	ae := admissionError(t, probeTheory+`statements:
  q(a): Q
  r(a): R
assumptions: [q(a)]
arguments:
  a1:
    scheme: s
    parameters: [a]
    premises: [q(a)]
    conclusion: r(a)
`)
	if !hasProblem(ae.Problems, validation.ARGUMENT, "a1", "premises do not instantiate") || ae.Context != nil {
		t.Fatalf("%v", ae)
	}
	if !strings.Contains(ae.Error(), "argument a1: premises do not instantiate") {
		t.Fatalf("message: %s", ae.Error())
	}
}

// The control: the same graph with the scheme's premise is admitted and
// labelled as before.
func TestAWellFormedAuthoredGraphIsAdmitted(t *testing.T) {
	a, err := NewAdapterFromYAML([]byte(probeTheory + `statements:
  p(a): P
  r(a): R
assumptions: [p(a)]
arguments:
  a1:
    scheme: s
    parameters: [a]
    premises: [p(a)]
    conclusion: r(a)
`))
	if err != nil {
		t.Fatal(err)
	}
	labels, err := a.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	if labels[a.Graph.Statements["r(a)"]] != caes.In {
		t.Fatalf("r(a) is %v", labels[a.Graph.Statements["r(a)"]])
	}
	if err := a.Admit(); err != nil {
		t.Fatalf("Admit after loading: %v", err)
	}
}

func TestWrongParametersAndConclusionsAreRefused(t *testing.T) {
	for name, c := range map[string]struct{ arg, want string }{
		"another parameter":          {"parameters: [b]\n    premises: [p(a)]\n    conclusion: r(a)", "premises do not instantiate"},
		"another conclusion":         {"parameters: [a]\n    premises: [p(a)]\n    conclusion: r(b)", "conclusion does not instantiate"},
		"too many parameters":        {"parameters: [a, b]\n    premises: [p(a)]\n    conclusion: r(a)", "number of parameters"},
		"no parameters, no instance": {"premises: [p(a)]\n    conclusion: r(b)", "do not instantiate the scheme with one substitution"},
	} {
		t.Run(name, func(t *testing.T) {
			ae := admissionError(t, probeTheory+"statements:\n  p(a): P\n  r(a): R\n  r(b): R\nassumptions: [p(a)]\narguments:\n  a1:\n    scheme: s\n    "+c.arg+"\n")
			if !hasProblem(ae.Problems, validation.ARGUMENT, "a1", c.want) {
				t.Fatalf("%v", ae)
			}
		})
	}
}

func TestContradictoryAssumptionsAreRefused(t *testing.T) {
	ae := admissionError(t, "statements:\n  p: P\n  ¬p: Not P\nassumptions: [p, ¬p]\n")
	if ae.Context == nil || len(ae.Problems) != 0 || !strings.Contains(ae.Error(), "the assumptions are not consistent") {
		t.Fatalf("%v", ae)
	}
	// Two positions of one issue, both assumed.
	ae = admissionError(t, "statements:\n  italy: Italy.\n  greece: Greece.\nissues:\n  i1:\n    positions: [italy, greece]\n    standard: PE\nassumptions: [italy, greece]\n")
	if ae.Context == nil {
		t.Fatalf("%v", ae)
	}
	// One of them is consistent.
	if _, err := NewAdapterFromYAML([]byte("statements:\n  italy: Italy.\n  greece: Greece.\nissues:\n  i1:\n    positions: [italy, greece]\n    standard: PE\nassumptions: [italy]\n")); err != nil {
		t.Fatal(err)
	}
}

func TestMalformedStatementsAndAssumptionsAreRefused(t *testing.T) {
	ae := admissionError(t, "statements:\n  p(X): A variable.\n")
	if !hasProblem(ae.Problems, validation.STATEMENT, "", "not a ground atomic formula") {
		t.Fatalf("%v", ae)
	}
	ae = admissionError(t, "statements:\n  p: P\nassumptions: [\"q(X)\"]\n")
	if !hasProblem(ae.Problems, validation.ASSUMPTION, "", "not a ground atomic formula") {
		t.Fatalf("%v", ae)
	}
}

// An argument naming a scheme with exceptions carries them (the engine's
// contract 6); a request to instantiate the scheme builds them.
func TestAuthoredArgumentsCarryTheirSchemesExceptions(t *testing.T) {
	theory := `language:
  p/1: "P %s"
  e/1: "E %s"
  r/1: "R %s"
argument_schemes:
  - id: s
    variables: [X]
    premises: [p(X)]
    exceptions: [e(X)]
    conclusions: [r(X)]
statements:
  p(a): P
  r(a): R
assumptions: [p(a)]
`
	ae := admissionError(t, theory+"arguments:\n  a1:\n    scheme: s\n    parameters: [a]\n    premises: [p(a)]\n    conclusion: r(a)\n")
	if !hasProblem(ae.Problems, validation.ARGUMENT, "a1", "does not carry the scheme's exceptions") {
		t.Fatalf("%v", ae)
	}
	a, err := NewAdapterFromYAML([]byte(theory + "arguments:\n  a1:\n    scheme: s\n    parameters: [a]\n"))
	if err != nil {
		t.Fatal(err)
	}
	labels, _ := a.Evaluate()
	if labels[a.Graph.Statements["r(a)"]] != caes.In {
		t.Fatalf("the instantiated argument: r(a) is %v", labels[a.Graph.Statements["r(a)"]])
	}
}

// A scheme pattern that is not a term would let any premise through, since
// the engine's validator skips it; an argument of that scheme is refused.
func TestAnArgumentOfASchemeWithAnUnreadablePatternIsRefused(t *testing.T) {
	ae := admissionError(t, `argument_schemes:
  - id: s
    variables: [X]
    premises: ["p(X"]
    conclusions: [r(X)]
statements:
  q(a): Q
  r(a): R
assumptions: [q(a)]
arguments:
  a1:
    scheme: s
    parameters: [a]
    premises: [q(a)]
    conclusion: r(a)
`)
	if !hasProblem(ae.Problems, validation.SCHEME, "s", "not a term") {
		t.Fatalf("%v", ae)
	}
	// The scheme alone, with no argument of it, is the lint's to report.
	if _, err := NewAdapterFromYAML([]byte("argument_schemes:\n  - id: s\n    variables: [X]\n    premises: [\"p(X\"]\n    conclusions: [r(X)]\n")); err != nil {
		t.Fatal(err)
	}
}

// The theory is not what admission checks: WAG's language declares negated
// predicates, which the engine's validator reports.
func TestTheShippedTheoryIsAdmittedThoughTheValidatorReportsItsLanguage(t *testing.T) {
	a, err := NewAdapterFromYAML(data.CoreYAML)
	if err != nil {
		t.Fatal(err)
	}
	language := 0
	for _, p := range validation.Validate(a.Graph) {
		if p.Category == validation.LANGUAGE {
			language++
		}
	}
	if language == 0 {
		t.Fatal("fixture: the validator no longer reports the shipped language; admission's exclusion of the theory may be revisited")
	}
	if ps := ProfileProblems(a.Graph); len(ps) != 0 {
		t.Fatalf("%v", ps)
	}
}

// answerTheory has WAG's two shapes of an answer scheme (D075): one that
// concludes its answer variable, and one that concludes a one-place predicate
// of it; and a scheme that concludes the negation of a variable, which is not
// an answer scheme.
const answerTheory = `language:
  asserts/2: "%s asserts %s"
  holds_in/1: "%s holds in the situation"
  sign_of/2: "%s is a sign of %s"
  p/0: "P"
  q/0: "Q"
  denies/1: "%s is denied"
  related/2: "%s is related to %s"
argument_schemes:
  - id: expert
    variables: [W, S]
    premises: ["asserts(W, S)"]
    conclusions: [S]
  - id: sign
    variables: [A, B]
    premises: ["sign_of(A, B)"]
    conclusions: [holds_in(B)]
  - id: deny
    variables: [S]
    premises: [denies(S)]
    conclusions: ["¬S"]
  - id: nay
    variables: [S]
    premises: [denies(S)]
    conclusions: ["not(S)"]
  - id: pair
    variables: [A, B]
    premises: ["sign_of(A, B)"]
    conclusions: ["related(A, B)"]
`

// answerStatements declares every statement the answers below use.
const answerStatements = `statements:
  p: P
  ¬p: Not P
  q: Q
  holds_in(p): P holds
  "asserts(w,p)": W asserts P
  "asserts(w,not(p))": W asserts not P
  "sign_of(s,p)": S is a sign of P
  "sign_of(s,not(p))": S is a sign of not P
  denies(p): P is denied
  "sign_of(p,s)": P is a sign of S
`

// Authored input can spell an opposing statement ¬X as not(X) inside a bound
// term. It cannot grant itself the runtime answer-fold waiver through metadata.
// Genuine runtime answers are checked separately below.
func TestAnswersConcludeTheClaimTheyAnswer(t *testing.T) {
	for name, c := range map[string]struct {
		arg    string
		admits bool
	}{
		"the claim":                              {"scheme: expert\n    parameters: [w, p]\n    premises: [\"asserts(w,p)\"]\n    conclusion: p", true},
		"the opposing position":                  {"scheme: expert\n    parameters: [w, \"not(p)\"]\n    premises: [\"asserts(w,not(p))\"]\n    conclusion: ¬p", true},
		"another claim":                          {"scheme: expert\n    parameters: [w, \"not(p)\"]\n    premises: [\"asserts(w,not(p))\"]\n    conclusion: q", false},
		"the position, not the bound":            {"scheme: expert\n    parameters: [w, \"not(p)\"]\n    premises: [\"asserts(w,not(p))\"]\n    conclusion: p", false},
		"a sign, folded":                         {"scheme: sign\n    parameters: [s, p]\n    premises: [\"sign_of(s,p)\"]\n    conclusion: p\n    meta: {wag_answer_variable: B}", false},
		"a sign, folded, opposing":               {"scheme: sign\n    parameters: [s, \"not(p)\"]\n    premises: [\"sign_of(s,not(p))\"]\n    conclusion: ¬p\n    meta: {wag_answer_variable: B}", false},
		"a sign folded, not marked as an answer": {"scheme: sign\n    parameters: [s, p]\n    premises: [\"sign_of(s,p)\"]\n    conclusion: p", false},
		"a sign folded, marked for A":            {"scheme: sign\n    parameters: [s, p]\n    premises: [\"sign_of(s,p)\"]\n    conclusion: p\n    meta: {wag_answer_variable: A}", false},
		"a sign of another claim, marked":        {"scheme: sign\n    parameters: [s, p]\n    premises: [\"sign_of(s,p)\"]\n    conclusion: q\n    meta: {wag_answer_variable: B}", false},
		"a sign, its own conclusion":             {"scheme: sign\n    parameters: [s, p]\n    premises: [\"sign_of(s,p)\"]\n    conclusion: holds_in(p)", true},
		"a sign of another claim":                {"scheme: sign\n    parameters: [s, p]\n    premises: [\"sign_of(s,p)\"]\n    conclusion: q", false},
		"a denial concludes its negation":        {"scheme: deny\n    parameters: [p]\n    premises: [\"denies(p)\"]\n    conclusion: ¬p", true},
		"a denial does not conclude S":           {"scheme: deny\n    parameters: [p]\n    premises: [\"denies(p)\"]\n    conclusion: p", false},
		"a two-place conclusion is not folded":   {"scheme: pair\n    parameters: [p, s]\n    premises: [\"sign_of(p,s)\"]\n    conclusion: p", false},
		"a not(S) conclusion is a negation":      {"scheme: nay\n    parameters: [p]\n    premises: [\"denies(p)\"]\n    conclusion: p", false},
		"the opposing position, a wrong premise": {"scheme: expert\n    parameters: [w, \"not(p)\"]\n    premises: [\"asserts(w,p)\"]\n    conclusion: ¬p", false},
	} {
		t.Run(name, func(t *testing.T) {
			yml := answerTheory + answerStatements + "arguments:\n  a1:\n    " + c.arg + "\n"
			if c.admits {
				if _, err := NewAdapterFromYAML([]byte(yml)); err != nil {
					t.Fatal(err)
				}
				return
			}
			ae := admissionError(t, yml)
			want := "conclusion does not instantiate"
			if strings.Contains(name, "premise") {
				want = "premises do not instantiate"
			}
			if !hasProblem(ae.Problems, validation.ARGUMENT, "a1", want) {
				t.Fatalf("%v", ae)
			}
		})
	}
}

// The engine's message is what ProfileProblems allows for answers; if the
// engine changes it, answers are refused (closed), and this test says why.
func TestTheEnginesConclusionMessage(t *testing.T) {
	a, err := caesImport(probeTheory + "statements:\n  p(a): P\n  r(b): R\narguments:\n  a1:\n    scheme: s\n    parameters: [a]\n    premises: [p(a)]\n    conclusion: r(b)\n")
	if err != nil {
		t.Fatal(err)
	}
	if !hasProblem(validation.Validate(a), validation.ARGUMENT, "a1", ConclusionProblem) {
		t.Fatalf("the engine's message changed: %v", validation.Validate(a))
	}
}

// Admit reports what the graph has become: a graph that was admitted and
// then given contradictory assumptions is not.
func TestAdmitChecksTheGraphAsItStands(t *testing.T) {
	a, err := NewAdapterFromYAML([]byte("statements:\n  p: P\n  ¬p: Not P\nassumptions: [p]\n"))
	if err != nil {
		t.Fatal(err)
	}
	a.Graph.Assumptions = append(a.Graph.Assumptions, "¬p")
	var ae *AdmissionError
	if err := a.Admit(); !errors.As(err, &ae) || ae.Context == nil {
		t.Fatalf("%v", err)
	}
	if err := (*Adapter)(nil).Admit(); err == nil {
		t.Fatal("a nil adapter is admitted")
	}
}

func TestAdmissionErrorListsAtMostFiveProblems(t *testing.T) {
	var ps []validation.Problem
	for i := 0; i < 7; i++ {
		ps = append(ps, validation.Problem{Category: validation.ARGUMENT, Id: string(rune('a' + i)), Description: "wrong."})
	}
	msg := (&AdmissionError{Problems: ps}).Error()
	if !strings.Contains(msg, "argument e: wrong;") || strings.Contains(msg, "argument f") || !strings.HasSuffix(msg, "and 2 more") {
		t.Fatal(msg)
	}
}

// The problems are listed in a stable order: by category (statements, then
// arguments, then assumptions), then by the object's ID.
func TestAdmissionProblemsAreInAStableOrder(t *testing.T) {
	yml := probeTheory + "statements:\n  q(a): Q\n  r(a): R\n  p(X): A variable\nassumptions: [q(a), \"q(Y)\"]\narguments:\n"
	for _, id := range []string{"a3", "a1", "a2"} {
		yml += "  " + id + ":\n    scheme: s\n    parameters: [a]\n    premises: [q(a)]\n    conclusion: r(a)\n"
	}
	first := admissionError(t, yml).Error()
	for i := 0; i < 20; i++ {
		if again := admissionError(t, yml).Error(); again != first {
			t.Fatalf("%s\nthen\n%s", first, again)
		}
	}
	st, a1, a2, a3, as := strings.Index(first, "statement:"), strings.Index(first, "argument a1:"), strings.Index(first, "argument a2:"),
		strings.Index(first, "argument a3:"), strings.Index(first, "assumption:")
	if st < 0 || !(st < a1 && a1 < a2 && a2 < a3 && a3 < as) {
		t.Fatal(first)
	}
}

// answerForm is false, and does not panic, for an argument without a
// parameter for each of its scheme's variables, or without a scheme or a
// conclusion.
func TestAnswerFormNeedsAParameterPerVariable(t *testing.T) {
	s := &caes.Scheme{Id: "expert", Variables: []string{"W", "S"}, Conclusions: []string{"S"}}
	for name, arg := range map[string]*caes.Argument{
		"one parameter": {Scheme: s, Parameters: []string{"w"}, Conclusion: &caes.Statement{Id: "w"}},
		"no scheme":     {Parameters: []string{"w", "p"}, Conclusion: &caes.Statement{Id: "p"}},
		"no conclusion": {Scheme: s, Parameters: []string{"w", "p"}},
		"no argument":   nil,
		"the answer":    {Scheme: s, Parameters: []string{"w", "not(p)"}, Conclusion: &caes.Statement{Id: "¬p"}},
	} {
		if got := answerForm(arg, true); got != (name == "the answer") {
			t.Errorf("%s: %v", name, got)
		}
	}
}

// An argument from a scheme that is not an answer scheme, concluding the
// bare action rather than its scheme's should_be_performed of it, is not
// folded: only answers are (the review of 2026-11-24, R4).
func TestANonAnswerSchemeIsNotFolded(t *testing.T) {
	// The frozen pre-C5.2 data, whose positive_consequences has no
	// exceptions: an authored argument of a scheme with exceptions must carry
	// their links (TestAuthoredArgumentsCarryTheirSchemesExceptions' concern, not this one's).
	core, err := os.ReadFile("../../data/frozen/afd3335a23f12f41/schemes/core.yml")
	if err != nil {
		t.Fatal(err)
	}
	yml := string(core) + `
statements:
  "brings_about(tax_meat,less_emissions)": Taxing meat brings about fewer emissions.
  good(less_emissions): Fewer emissions are good.
  tax_meat: Meat is taxed.
assumptions: ["brings_about(tax_meat,less_emissions)", good(less_emissions)]
arguments:
  a1:
    scheme: positive_consequences
    parameters: [tax_meat, less_emissions]
    premises: ["brings_about(tax_meat,less_emissions)", good(less_emissions)]
    conclusion: tax_meat
`
	ae := admissionError(t, yml)
	if !hasProblem(ae.Problems, validation.ARGUMENT, "a1", "conclusion does not instantiate") {
		t.Fatalf("%v", ae)
	}
	// Its own conclusion is admitted.
	ok := strings.Replace(strings.Replace(yml, "conclusion: tax_meat", "conclusion: should_be_performed(tax_meat)", 1),
		"tax_meat: Meat is taxed.", "should_be_performed(tax_meat): Meat should be taxed.", 1)
	if _, err := NewAdapterFromYAML([]byte(ok)); err != nil {
		t.Fatal(err)
	}
}
