package annotate

import (
	"encoding/json"
	"strings"
	"testing"
)

// Cases RC1-RC5 and RC7 (docs/d086-first-gold.md §6; D096,
// docs/annotation-records.md §13): the Go check of an instance by the
// discussion, the environmental argument recorded at the thesis.
func TestDiscussionInstanceCheck(t *testing.T) {
	d, c := builtin(t)
	// plant puts the environmental argument at claim `at` as the discussion's
	// (or takes the one the gold has there) and returns the claim's record.
	plant := func(set map[string]Annotation, at string) (Annotation, map[string]any) {
		a := set[at]
		insts := a["instances"].([]any)
		for _, x := range insts {
			if in := x.(map[string]any); in["by"] == "the discussion" {
				return a, in
			}
		}
		var src map[string]any
		for _, x := range set["13292.1"]["instances"].([]any) {
			if in := x.(map[string]any); in["by"] == "the discussion" {
				src = in
			}
		}
		for _, x := range set["13292.108"]["instances"].([]any) {
			in := x.(map[string]any)
			sc := in["scheme"].(map[string]any)
			if src == nil && sc["walton_id"] == "practical_reasoning2" && in["conclusion"].(map[string]any)["formula"] == "should_be_performed(tax_meat)" {
				src = in
			}
		}
		if src == nil {
			t.Fatal("no environmental argument to plant")
		}
		var in map[string]any
		b, _ := json.Marshal(src)
		_ = json.Unmarshal(b, &in)
		in["id"], in["primary"], in["by"] = at+"#9", false, "the discussion"
		in["scheme"].(map[string]any)["fit"] = "reconstructed"
		for _, y := range in["premises"].([]any) {
			if pm := y.(map[string]any); pm["status"] == "explicit" {
				pm["status"], pm["cite"] = "inferred", "13292.108"
			}
		}
		a["instances"] = append(insts, in)
		return a, in
	}
	has := func(out []string, prefix, text string) bool {
		for _, p := range out {
			if strings.HasPrefix(p, prefix) && strings.Contains(p, text) {
				return true
			}
		}
		return false
	}
	none := func(out []string) bool { return !has(out, "R20 ", "") && !has(out, "R16 ", " by ") }

	set := clone(t, d.Gold)
	a, _ := plant(set, "13292.1")
	if out := c.CheckIn(a, d, set, 0); !none(out) {
		t.Fatalf("RC1: the discussion's instance is reported: %v", out)
	}
	for _, tc := range []struct {
		name, prefix, text string
		edit               func(a Annotation, in map[string]any)
	}{
		{"RC2 another value", "R16 ", "by", func(a Annotation, in map[string]any) { in["by"] = "the claim" }},
		{"RC2 not text", "R16 ", "by", func(a Annotation, in map[string]any) { in["by"] = true }},
		{"RC3 primary", "R20 ", "is primary", func(a Annotation, in map[string]any) {
			for _, x := range a["instances"].([]any) {
				x.(map[string]any)["primary"] = false
			}
			in["primary"] = true
		}},
		{"RC4 fragment", "R20 ", "fit", func(a Annotation, in map[string]any) { in["scheme"].(map[string]any)["fit"] = "fragment" }},
		{"RC4 explicit", "R20 ", "fit", func(a Annotation, in map[string]any) { in["scheme"].(map[string]any)["fit"] = "explicit" }},
		{"RC5 an explicit premise", "R20 ", "explicit", func(a Annotation, in map[string]any) {
			in["premises"].([]any)[0].(map[string]any)["status"] = "explicit"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			set := clone(t, d.Gold)
			a, in := plant(set, "13292.1")
			tc.edit(a, in)
			if out := c.CheckIn(a, d, set, 0); !has(out, tc.prefix, tc.text) {
				t.Fatalf("not reported (%s… %q): %v", tc.prefix, tc.text, out)
			}
		})
	}
	// RC7: at a claim whose statement it does not conclude.
	set = clone(t, d.Gold)
	a, _ = plant(set, "13292.2")
	if out := c.CheckIn(a, d, set, 0); !has(out, "R20 ", "neither") {
		t.Fatalf("RC7: a discussion's instance at 13292.2 concluding the thesis is not reported: %v", out)
	}
}

// Case RC10: the Go measure's instances field leaves an instance by the
// discussion out, as agreement.py does; a target at it is compared by its
// scheme and conclusion, as at a secondary instance.
func TestDiscussionInstanceAgreement(t *testing.T) {
	own := map[string]any{"id": "c.0#1", "primary": true,
		"scheme":     map[string]any{"walton_id": "positive_consequences", "fit": "reconstructed"},
		"conclusion": map[string]any{"formula": "should_be_performed(a)"}}
	disc := map[string]any{"id": "c.0#2", "primary": false, "by": "the discussion",
		"scheme":     map[string]any{"walton_id": "practical_reasoning2", "fit": "reconstructed"},
		"conclusion": map[string]any{"formula": "should_be_performed(a)"}}
	without := Annotation{"kialo_claim_id": "c.0", "instances": []any{own}}
	with := Annotation{"kialo_claim_id": "c.0", "instances": []any{own, disc}}
	if a, b := FieldOf(with, "instances"), FieldOf(without, "instances"); a != b {
		t.Fatalf("instances: %q with the discussion's instance, %q without", a, b)
	}
	item := Annotation{"kialo_claim_id": "c.1", "target": map[string]any{"actual": map[string]any{
		"claim": "c.0", "instance": "c.0#2", "part": "conclusion", "formula": "should_be_performed(a)"}}}
	set := map[string]Annotation{"c.0": with, "c.1": item}
	if got, want := FieldIn(item, "target_instance", set), "c.0 secondary:practical_reasoning2:should_be_performed(a)"; got != want {
		t.Fatalf("target_instance: %q, want %q", got, want)
	}
}
