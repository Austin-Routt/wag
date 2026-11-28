package annotate

import (
	"encoding/json"
	"strings"
	"testing"
)

// Cases CT1-CT9 (docs/d086-first-gold.md §6): R19 in the Go check, as
// validate.py checks it. An inferred premise cites another claim of the
// discussion; the check reads the citation, not what the claim says.
func TestCitationCheck(t *testing.T) {
	d, c := builtin(t)
	// The fixture: 13292.218#1's inferred circumstances, cited to 13292.11.
	inferred := func(a Annotation) map[string]any {
		for _, x := range inst(a, 0)["premises"].([]any) {
			if pm := x.(map[string]any); pm["status"] == "inferred" {
				return pm
			}
		}
		t.Fatal("13292.218#1 has no inferred premise")
		return nil
	}
	r19 := func(out []string) []string {
		var got []string
		for _, p := range out {
			if strings.HasPrefix(p, "R19 ") {
				got = append(got, p)
			}
		}
		return got
	}
	if got := r19(c.CheckIn(clone(t, d.Gold)["13292.218"], d, clone(t, d.Gold), 2)); len(got) != 0 {
		t.Fatalf("the gold's 13292.218: %v", got)
	}
	for _, tc := range []struct {
		name    string
		cite    any
		omit    bool
		context bool
		refused bool
	}{
		{name: "CT1 missing", omit: true, refused: true},
		{name: "CT2 null", cite: nil, refused: true},
		{name: "CT3 a number", cite: 13292.11, refused: true},
		{name: "CT4 the claim itself", cite: "13292.218", refused: true},
		{name: "CT5 a claim the dataset does not list", cite: "13292.999999", refused: true},
		{name: "CT6 a claim of the batch", cite: "13292.11"},
		{name: "CT7 a claim of a claim's context", cite: "13292.888888", context: true},
		{name: "CT8 a known claim that does not say it", cite: "13292.1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			set := clone(t, d.Gold)
			a := set["13292.218"]
			pm := inferred(a)
			if tc.omit {
				delete(pm, "cite")
			} else {
				pm["cite"] = tc.cite
			}
			dd := *d
			if tc.context {
				dd.Items = append([]Item(nil), d.Items...)
				var raw map[string]any
				if err := json.Unmarshal(dd.Items[0].Raw, &raw); err != nil {
					t.Fatal(err)
				}
				raw["siblings"] = append(raw["siblings"].([]any), map[string]any{"claim_id": "13292.888888", "text": "a sibling"})
				b, err := json.Marshal(raw)
				if err != nil {
					t.Fatal(err)
				}
				dd.Items[0].Raw = b
			}
			out := c.CheckIn(a, &dd, set, 2)
			if got := r19(out); (len(got) != 0) != tc.refused {
				t.Fatalf("R19 %v, want refused %v; all: %v", got, tc.refused, out)
			}
			// CT9: a format-1 check knows no R19.
			if got := r19(c.CheckIn(a, &dd, set, 1)); len(got) != 0 {
				t.Fatalf("format 1: %v", got)
			}
		})
	}
}
