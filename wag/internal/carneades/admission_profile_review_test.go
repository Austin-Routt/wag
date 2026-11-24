package carneades

import (
	"errors"
	"strings"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/validation"
	"github.com/yourorg/wag/data"
)

// Merely declaring metadata must not turn a non-answer scheme's practical
// conclusion into its bare action.
func TestReviewAuthoredAnswerMarkerCannotAuthorizeNonAnswerScheme(t *testing.T) {
	yml := string(data.CoreYAML) + `
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
    meta: {wag_answer_variable: A}
`
	plain := strings.Replace(yml, "    meta: {wag_answer_variable: A}\n", "", 1)
	if _, err := NewAdapterFromYAML([]byte(plain)); err == nil {
		t.Fatal("control: unmarked conclusion mismatch was not refused")
	}
	a, err := NewAdapterFromYAML([]byte(yml))
	if err != nil {
		var ae *AdmissionError
		if !errors.As(err, &ae) || !hasProblem(ae.Problems, validation.ARGUMENT, "a1", ConclusionProblem) {
			t.Fatalf("refused for an unrelated reason: %v", err)
		}
		return
	}
	native := 0
	for _, p := range validation.Validate(a.Graph) {
		if p.Category == validation.ARGUMENT && p.Id == "a1" && p.Description == ConclusionProblem {
			native++
		}
	}
	labels, err := a.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	if native != 1 || labels[a.Graph.Statements["tax_meat"]] != caes.In {
		t.Fatalf("counterexample changed: native conclusion problems=%d, bare action label=%v", native, labels[a.Graph.Statements["tax_meat"]])
	}
	t.Fatalf("metadata alone authorized a non-answer scheme: native conclusion mismatch=%d, WAG bare action=In, profile problems=%v", native, ProfileProblems(a.Graph))
}

func TestReviewRuntimeAnswerRetainsItsFold(t *testing.T) {
	for _, positive := range []bool{true, false} {
		name, premise, want := "positive", "sign_of(s,p)", "p"
		if !positive {
			name, premise, want = "opposing", "sign_of(s,not(p))", "¬p"
		}
		t.Run(name, func(t *testing.T) {
			a, err := NewAdapterFromYAML([]byte(answerTheory + answerStatements + "assumptions: [\"" + premise + "\"]\n"))
			if err != nil {
				t.Fatal(err)
			}
			ids, err := a.AnswerQuestion(Answer{Owner: "player", SchemeID: "sign", Bindings: map[string]string{"A": "s"},
				AnswerVariable: "B", Locus: LocusStatement, LocusStmtID: "p", Positive: positive, Weight: 1})
			if err != nil || len(ids) != 1 {
				t.Fatalf("runtime answer: ids=%v err=%v", ids, err)
			}
			arg := a.Graph.Arguments[ids[0]]
			if arg.Metadata[AnswerKey] != "B" || arg.Conclusion.Id != want {
				t.Fatalf("answer changed: metadata=%v conclusion=%s", arg.Metadata, arg.Conclusion.Id)
			}
			if err := a.Admit(); err != nil {
				t.Fatalf("genuine runtime answer refused: %v", err)
			}
			labels, err := a.Evaluate()
			if err != nil || labels[arg.Conclusion] != caes.In {
				t.Fatalf("runtime answer result: label=%v error=%v", labels[arg.Conclusion], err)
			}
		})
	}
}

// An issue scheme is an instruction to construct opposition. CheckContext
// examines its private closure; labelling the untouched graph must not ignore
// the opposition that check just relied on.
func TestReviewAuthoredIssueSchemeMustNotDisappearAtEvaluation(t *testing.T) {
	yml := `language:
  p/0: P
  q/0: Q
issue_schemes:
  choice: [p, q]
statements:
  p: P
  q: Q
arguments:
  ap:
    conclusion: p
    premises: []
  aq:
    conclusion: q
    premises: []
`
	a, err := NewAdapterFromYAML([]byte(yml))
	if err != nil {
		var ae *AdmissionError
		if !errors.As(err, &ae) || !hasProblem(ae.Problems, validation.ISCHEME, "choice", "unsupported") {
			t.Fatalf("refused for an unrelated reason: %v", err)
		}
		// Concrete issues remain supported without running inference.
		concrete := strings.Replace(yml, "issue_schemes:\n  choice: [p, q]\n", "issues:\n  choice:\n    positions: [p, q]\n    standard: PE\n", 1)
		control, err := NewAdapterFromYAML([]byte(concrete))
		if err != nil {
			t.Fatalf("concrete issue refused: %v", err)
		}
		labels, err := control.Evaluate()
		if err != nil || len(control.Graph.Issues) != 1 || labels[control.Graph.Statements["p"]] != caes.Out || labels[control.Graph.Statements["q"]] != caes.Out {
			t.Fatalf("concrete issue control: p=%v q=%v issues=%d error=%v", labels[control.Graph.Statements["p"]], labels[control.Graph.Statements["q"]], len(control.Graph.Issues), err)
		}
		return
	}
	labels, err := a.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	if labels[a.Graph.Statements["p"]] == caes.In && labels[a.Graph.Statements["q"]] == caes.In {
		native, err := caesImport(yml)
		if err != nil {
			t.Fatal(err)
		}
		if ps := validation.Validate(native); len(ps) != 0 {
			t.Fatalf("native fixture problems: %v", ps)
		}
		if err := native.Infer(); err != nil {
			t.Fatal(err)
		}
		nativeLabels := native.GroundedLabelling()
		if len(native.Issues) != 1 || nativeLabels[native.Statements["p"]] == caes.In && nativeLabels[native.Statements["q"]] == caes.In {
			t.Fatalf("native control changed: issues=%d p=%v q=%v", len(native.Issues), nativeLabels[native.Statements["p"]], nativeLabels[native.Statements["q"]])
		}
		t.Fatalf("declared mutually exclusive positions both In after admission: WAG issues=%d; CheckContext=%v; native issues=%d p=%v q=%v", len(a.Graph.Issues), a.Graph.CheckContext(), len(native.Issues), nativeLabels[native.Statements["p"]], nativeLabels[native.Statements["q"]])
	}
}
