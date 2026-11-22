package yaml

import (
	"bytes"
	"strings"
	"testing"
)

func TestRejectLossyNativeInput(t *testing.T) {
	for name, body := range map[string]string{
		"object premise":         "arguments: {a: {conclusion: c, premises: [{statement: p, role: reply}]}}",
		"numeric premise":        "arguments: {a: {conclusion: c, premises: [7]}}",
		"scalar premises":        "arguments: {a: {conclusion: c, premises: p}}",
		"object role":            "arguments: {a: {conclusion: c, premises: {reply: {statement: p}}}}",
		"attack direction":       "arguments: {a: {conclusion: c, premises: [p], direction: attack}}",
		"pro flag":               "arguments: {a: {conclusion: c, premises: [p], pro: false}}",
		"duplicate argument key": "arguments: {a: {conclusion: c}, a: {conclusion: p}}",
		"null argument":          "arguments: {a: null}",
		"null issue":             "issues: {i: null}",
		"null scheme":            "argument_schemes: [null]",
		"unknown top level":      "assumptons: [p]",
	} {
		t.Run(name, func(t *testing.T) {
			g, err := Import(strings.NewReader("statements: {p: Premise, c: Conclusion}\n" + body))
			if err == nil || g != nil {
				t.Fatalf("malformed/unsupported native input returned a graph: %v, %v", g, err)
			}
		})
	}
	for _, input := range []string{
		"statements: {p: null}", "statements: {p: [text]}",
		"statements: {p: {txt: Typo}}", "statements: {p: {label: unknown}}",
		"statements: {p: one, p: two}",
	} {
		if g, err := Import(strings.NewReader(input)); err == nil || g != nil {
			t.Fatalf("invalid statement was accepted: %s", input)
		}
	}
}

func TestNativePremisesAndWeightSurviveImport(t *testing.T) {
	for _, premises := range []string{"[p]", "{reason: p}"} {
		g, err := Import(strings.NewReader("statements: {p: Premise, c: Conclusion}\narguments: {a: {conclusion: c, premises: " + premises + ", weight: 0.7}}"))
		if err != nil {
			t.Fatal(err)
		}
		a := g.Arguments["a"]
		if len(a.Premises) != 1 || a.Premises[0].Stmt != g.Statements["p"] || a.Weight != 0.7 {
			t.Fatalf("premise/weight lost: %+v", a)
		}
		if got := g.GroundedLabelling()[g.Statements["c"]].String(); got != "out" {
			t.Fatalf("an unaccepted premise must not become an unconditional conclusion: %s", got)
		}
		var output bytes.Buffer
		Export(&output, g)
		if _, err := Import(&output); err != nil {
			t.Fatalf("native export must satisfy the strict input profile: %v", err)
		}
	}
}
