package annotate

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
)

// The built-in dataset is a copy of docs/kialo-batch2 and of the cleaned
// walton.yml: they must stay equal.
func TestBuiltinIsTheDocsCopy(t *testing.T) {
	for emb, doc := range map[string]string{
		"annotation/kialo-13292-batch2/items.json":    "../../../docs/kialo-batch2/items.json",
		"annotation/kialo-13292-batch2/guidelines.md": "../../../docs/kialo-batch2/guidelines.md",
		"annotation/kialo-13292-batch2/lexicon.json":  "../../../docs/kialo-batch2/lexicon.json",
		"annotation/kialo-13292-batch2/gold.json":     "../../../docs/kialo-batch2/gold.json",
		"annotation/kialo-13292-batch2/case.json":     "../../../docs/kialo-batch2/case.json",
		"annotation/walton.yml":                       "../../../docs/scheme-registry/walton-clean/walton.yml",
	} {
		a, err := data.Annotation.ReadFile(emb)
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(doc)
		if err != nil {
			t.Skip("the docs are not here:", err)
		}
		if !bytes.Equal(a, b) {
			t.Errorf("wag/data/%s differs from %s: copy it again", emb, doc)
		}
	}
}

// The gold annotation passes the check, as it passes validate.py.
func TestGoldChecks(t *testing.T) {
	ds, err := Builtin()
	if err != nil || len(ds) == 0 {
		t.Fatal(err)
	}
	d := ds[0]
	if d.NItems != 50 || !d.HasGold || len(d.Gold) != 50 {
		t.Fatalf("%+v", d)
	}
	c, err := LoadCollection(d.WaltonYAML())
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range d.IDs() {
		// With the gold's own annotations, in which its exact targets are
		// read (format 2, docs/annotation-records.md).
		if p := c.CheckIn(d.Gold[id], d, d.Gold, 0); len(p) > 0 {
			t.Errorf("%s: %v", id, p)
		}
	}
	bad := Annotation{"instances": []any{map[string]any{"primary": true, "scheme": map[string]any{"walton_id": "sign", "fit": "explicit"},
		"bindings": map[string]any{"A": "x"}, "premises": []any{map[string]any{"role": "premise", "formula": "finding_true(y)"}},
		"conclusion": map[string]any{"formula": "true_in_situation(B)"}}}, "role": map[string]any{"kind": "guess"}}
	if p := c.Check(bad, d); len(p) < 4 {
		t.Fatalf("a bad annotation: %v", p)
	}
	// Spaces in a formula do not make it another formula (a model wrote
	// "brings_about(tax_meat, x)").
	var spaced Annotation
	b, _ := json.Marshal(d.Gold["13292.7"])
	_ = json.Unmarshal([]byte(strings.ReplaceAll(string(b), ",", ", ")), &spaced)
	if p := c.Check(spaced, d); len(p) > 0 {
		t.Fatalf("spaces after commas: %v", p)
	}
}

// A claim at an exception, an answer, concludes the exception it answers,
// or its negation when it defends the argument (D075; rule R8 of
// validate.py): an argument for another course, or against the action, is
// not an answer there.
func TestAnswerAtAnExceptionConcludesIt(t *testing.T) {
	ds, err := Builtin()
	if err != nil || len(ds) == 0 {
		t.Fatal(err)
	}
	d := ds[0]
	c, err := LoadCollection(d.WaltonYAML())
	if err != nil {
		t.Fatal(err)
	}
	const exc = "realize_more_effectively(governments, meat_price_reflects_real_costs, tax_meat)"
	answer := func(kind, part, concludes string, primary bool, more ...map[string]any) Annotation {
		insts := []any{map[string]any{"primary": primary, "scheme": map[string]any{"walton_id": "other", "fit": "reconstructed"},
			"bindings": map[string]any{}, "premises": []any{}, "conclusion": map[string]any{"formula": concludes}}}
		for _, m := range more {
			insts = append(insts, m)
		}
		return Annotation{
			"instances": insts,
			"role":      map[string]any{"kind": kind, "cq": "CQ:other_means", "polarity": "against"},
			"target": map[string]any{"kialo_parent": "13292.1", "actual": map[string]any{"claim": "13292.118", "part": part,
				"formula": exc}},
			"evidence": map[string]any{"status": "assertion"},
		}
	}
	r8 := func(a Annotation) bool {
		for _, p := range c.Check(a, d) {
			if strings.Contains(p, "an answer at the exception") {
				return true
			}
		}
		return false
	}
	other := map[string]any{"primary": false, "scheme": map[string]any{"walton_id": "other", "fit": "reconstructed"},
		"bindings": map[string]any{}, "premises": []any{}, "conclusion": map[string]any{"formula": "should_be_performed(abolish_meat_subsidies)"}}
	defend := func(concludes string) Annotation {
		a := answer("cq_answer", "exception", concludes, true)
		a["role"].(map[string]any)["polarity"] = "for"
		return a
	}
	secondFirst := answer("cq_answer", "exception", "should_be_performed(abolish_meat_subsidies)", true)
	secondFirst["instances"] = append([]any{map[string]any{"primary": false, "scheme": map[string]any{"walton_id": "other", "fit": "reconstructed"},
		"bindings": map[string]any{}, "premises": []any{}, "conclusion": map[string]any{"formula": exc}}}, secondFirst["instances"].([]any)...)
	for _, tc := range []struct {
		name string
		a    Annotation
		want bool
	}{
		{"concludes the exception", answer("cq_answer", "exception", "realize_more_effectively(governments,meat_price_reflects_real_costs,tax_meat)", true), false},
		{"its only instance, not marked primary", answer("cq_answer", "exception", "realize_more_effectively(governments,meat_price_reflects_real_costs,tax_meat)", false), false},
		{"the primary among others", answer("cq_answer", "exception", "realize_more_effectively(governments,meat_price_reflects_real_costs,tax_meat)", true, other), false},
		{"an argument for another course", answer("cq_answer", "exception", "should_be_performed(abolish_meat_subsidies)", true), true},
		{"bound to another agent", answer("cq_answer", "exception", "realize_more_effectively(the_city,meat_price_reflects_real_costs,tax_meat)", true), true},
		{"whatever the role", answer("rebuttal", "exception", "should_be_performed(abolish_meat_subsidies)", true), true},
		{"no conclusion", answer("cq_answer", "exception", "", true), true},
		{"a secondary that concludes it", answer("cq_answer", "exception", "should_be_performed(abolish_meat_subsidies)", true, map[string]any{
			"primary": false, "scheme": map[string]any{"walton_id": "other", "fit": "reconstructed"}, "bindings": map[string]any{},
			"premises": []any{}, "conclusion": map[string]any{"formula": exc}}), true},
		{"a secondary listed first that concludes it", secondFirst, true},
		{"a defence concluding the negation", defend("¬realize_more_effectively(governments,meat_price_reflects_real_costs,tax_meat)"), false},
		{"a defence concluding the exception", defend("realize_more_effectively(governments,meat_price_reflects_real_costs,tax_meat)"), true},
		// Two primaries are refused anyway; the first is the one read, as
		// validate.py reads it.
		{"two primaries, the first concluding it", answer("cq_answer", "exception", "realize_more_effectively(governments,meat_price_reflects_real_costs,tax_meat)", true, map[string]any{
			"primary": true, "scheme": map[string]any{"walton_id": "other", "fit": "reconstructed"}, "bindings": map[string]any{},
			"premises": []any{}, "conclusion": map[string]any{"formula": "should_be_performed(abolish_meat_subsidies)"}}), false},
		{"at a premise", answer("cq_answer", "premise", "should_be_performed(abolish_meat_subsidies)", true), false},
	} {
		if got := r8(tc.a); got != tc.want {
			t.Errorf("%s: R8 reported %v, want %v: %v", tc.name, got, tc.want, c.Check(tc.a, d))
		}
	}
	// An instance with no conclusion formula concludes nothing.
	none := answer("cq_answer", "exception", "", true)
	none["instances"].([]any)[0].(map[string]any)["conclusion"] = map[string]any{}
	if p := c.Check(none, d); !strings.Contains(strings.Join(p, "; "), "concludes nothing, not realize_more_effectively(") {
		t.Errorf("no conclusion: %v", p)
	}
}
