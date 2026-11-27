package annotate

import "testing"

// Case A4 (docs/d086-first-gold.md §6; D102): FieldIn leaves an unsettled
// relation out as agreement.py does (cases A1-A3 in test_agreement.py).
func TestUnsettledRelationLeftOut(t *testing.T) {
	mark := map[string]any{"unresolved": "a background condition of the premise"}
	parent := Annotation{"kialo_claim_id": "c.0", "instances": []any{map[string]any{
		"id": "c.0#1", "primary": true,
		"scheme":     map[string]any{"walton_id": "positive_consequences", "fit": "reconstructed"},
		"premises":   []any{map[string]any{"role": "premise", "formula": "brings_about(a,g)"}},
		"conclusion": map[string]any{"formula": "should_be_performed(a)"},
	}}}
	item := func(formula, part, instance any) Annotation {
		return Annotation{"kialo_claim_id": "c.1",
			"instances": []any{map[string]any{"id": "c.1#1", "primary": true,
				"scheme":     map[string]any{"walton_id": "verbal_classification", "fit": "reconstructed"},
				"conclusion": map[string]any{"formula": "instance(o,g)"}}},
			"role": map[string]any{"kind": "premise_backing", "polarity": "for", "cq": nil},
			"target": map[string]any{"kialo_parent": "c.0", "actual": map[string]any{
				"claim": "c.0", "instance": instance, "part": part, "formula": formula}}}
	}
	relation := []string{"role", "cq", "target_part", "target_instance", "target_formula"}
	for _, tc := range []struct {
		name string
		a    Annotation
	}{
		{"A1 formula unresolved", item(mark, "premise", "c.0#1")},
		{"A2 part unresolved", item("brings_about(a,g)", mark, "c.0#1")},
	} {
		set := map[string]Annotation{"c.0": parent, "c.1": tc.a}
		for _, f := range relation {
			if v := FieldIn(tc.a, f, set); v != Unresolved {
				t.Errorf("%s: %s = %q, want unresolved", tc.name, f, v)
			}
		}
		for f, want := range map[string]string{"scheme": "verbal_classification", "fit": "reconstructed",
			"polarity": "for", "target_claim": "c.0"} {
			if v := FieldIn(tc.a, f, set); v != want {
				t.Errorf("%s: %s = %q, want %q", tc.name, f, v, want)
			}
		}
	}
	// A3: an unresolved instance alone leaves only the instance out.
	a3 := item("brings_about(a,g)", "premise", mark)
	set := map[string]Annotation{"c.0": parent, "c.1": a3}
	if v := FieldIn(a3, "target_instance", set); v != Unresolved {
		t.Errorf("A3: target_instance = %q, want unresolved", v)
	}
	for f, want := range map[string]string{"role": "premise_backing", "target_part": "premise",
		"target_formula": "brings_about(a,g)"} {
		if v := FieldIn(a3, f, set); v != want {
			t.Errorf("A3: %s = %q, want %q", f, v, want)
		}
	}
	// The resolved control: every relation field compared.
	ok := item("brings_about(a,g)", "premise", "c.0#1")
	set = map[string]Annotation{"c.0": parent, "c.1": ok}
	for _, f := range relation {
		if v := FieldIn(ok, f, set); v == Unresolved {
			t.Errorf("control: %s unresolved", f)
		}
	}
}
