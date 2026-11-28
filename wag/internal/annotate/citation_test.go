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

// Cases CT10-CT16 (docs/d086-first-gold.md §6; the check of the D096 fixes,
// W4 and N1): each branch of R19's known claims, premises of any role and
// instance, no dataset, and a context list with an entry that is not a
// claim.
func TestCitationCheckBranches(t *testing.T) {
	d, c := builtin(t)
	r19 := func(out []string) []string {
		var got []string
		for _, p := range out {
			if strings.HasPrefix(p, "R19 ") {
				got = append(got, p)
			}
		}
		return got
	}
	// 13292.218#1's inferred premise, in a copy of the gold.
	fixture := func() (map[string]Annotation, Annotation, map[string]any) {
		set := clone(t, d.Gold)
		a := set["13292.218"]
		for _, x := range inst(a, 0)["premises"].([]any) {
			if pm := x.(map[string]any); pm["status"] == "inferred" {
				return set, a, pm
			}
		}
		t.Fatal("13292.218#1 has no inferred premise")
		return nil, nil, nil
	}
	// The dataset with every item's context lists emptied, then the lists
	// given set on its first item.
	contexts := func(lists map[string][]any) *Dataset {
		dd := *d
		dd.Items = append([]Item(nil), d.Items...)
		for i := range dd.Items {
			var raw map[string]any
			if err := json.Unmarshal(dd.Items[i].Raw, &raw); err != nil {
				t.Fatal(err)
			}
			for _, f := range []string{"kialo_parents", "ancestors", "siblings", "children"} {
				raw[f] = []any{}
				if l, ok := lists[f]; ok && i == 0 {
					raw[f] = l
				}
			}
			b, err := json.Marshal(raw)
			if err != nil {
				t.Fatal(err)
			}
			dd.Items[i].Raw = b
		}
		return &dd
	}
	accepted := func(t *testing.T, dd *Dataset, cite string) {
		t.Helper()
		set, a, pm := fixture()
		pm["cite"] = cite
		if got := r19(c.CheckIn(a, dd, set, 2)); len(got) != 0 {
			t.Fatalf("refused: %v", got)
		}
	}
	t.Run("CT10 a claim of the batch that no context lists", func(t *testing.T) {
		accepted(t, contexts(nil), "13292.11")
		// The same cite is refused once the batch does not have it either:
		// the fixture is what decides.
		set, a, pm := fixture()
		pm["cite"] = "13292.777777"
		if got := r19(c.CheckIn(a, contexts(nil), set, 2)); len(got) != 1 {
			t.Fatalf("an unknown claim: %v", got)
		}
	})
	t.Run("CT11 a claim listed only among the children", func(t *testing.T) {
		accepted(t, contexts(map[string][]any{"children": {map[string]any{"claim_id": "13292.777777"}}}), "13292.777777")
	})
	t.Run("CT12 a claim listed only among the Kialo parents, and only among the ancestors", func(t *testing.T) {
		accepted(t, contexts(map[string][]any{"kialo_parents": {map[string]any{"claim_id": "13292.777777", "relation": "pro"}}}), "13292.777777")
		accepted(t, contexts(map[string][]any{"ancestors": {map[string]any{"claim_id": "13292.777777"}}}), "13292.777777")
	})
	t.Run("CT13 premises of role exception and assumption", func(t *testing.T) {
		for _, role := range []string{"exception", "assumption"} {
			set, a, pm := fixture()
			pm["role"], pm["cite"] = role, "13292.218"
			if got := r19(c.CheckIn(a, d, set, 2)); len(got) != 1 {
				t.Fatalf("%s: %v", role, got)
			}
		}
	})
	t.Run("CT14 an instance that is not the primary one", func(t *testing.T) {
		set, a, _ := fixture()
		second := map[string]any{}
		b, _ := json.Marshal(inst(a, 0))
		if err := json.Unmarshal(b, &second); err != nil {
			t.Fatal(err)
		}
		second["id"], second["primary"] = "13292.218#2", false
		for _, x := range second["premises"].([]any) {
			if pm := x.(map[string]any); pm["status"] == "inferred" {
				pm["cite"] = "13292.218"
			}
		}
		a["instances"] = append(a["instances"].([]any), second)
		got := r19(c.CheckIn(a, d, set, 2))
		if len(got) != 1 || !strings.HasPrefix(got[0], "R19 13292.218#2's") {
			t.Fatalf("%v", got)
		}
	})
	t.Run("CT15 no dataset", func(t *testing.T) {
		for _, tc := range []struct {
			cite    any
			omit    bool
			refused bool
		}{{omit: true, refused: true}, {cite: "13292.218", refused: true}, {cite: "13292.777777"}} {
			_, a, pm := fixture()
			if tc.omit {
				delete(pm, "cite")
			} else {
				pm["cite"] = tc.cite
			}
			if got := checkCitations(a, nil); (len(got) != 0) != tc.refused {
				t.Fatalf("cite %v: %v, want refused %v", tc.cite, got, tc.refused)
			}
		}
	})
	t.Run("CT16 an entry that is not a claim", func(t *testing.T) {
		accepted(t, contexts(map[string][]any{"siblings": {"junk", map[string]any{"claim_id": "13292.777777"}}}), "13292.777777")
	})
}
