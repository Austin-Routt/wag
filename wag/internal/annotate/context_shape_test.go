package annotate

import (
	"encoding/json"
	"strings"
	"testing"
)

// Cases SC7-SC10 (docs/d086-first-gold.md §6; F5 of the external review of
// exchange 3): Go refuses, when it reads a dataset, a claim's Kialo context
// the checks could not read, and a claim with no text; it reads the context
// by its names as written, as R19 and validate.py do. validate.py's side is
// SC1-SC6 (docs/kialo-batch2/test_validate.py, SourceShape).

// absentField marks a field contextDataset leaves out.
const absentField = "<absent>"

// contextDataset reads a dataset of two claims, the thesis 1.1 and 1.2, whose
// fields are well formed but for those set (absentField leaves one out).
func contextDataset(t *testing.T, set map[string]any) (*Dataset, error) {
	t.Helper()
	second := map[string]any{"kialo_claim_id": "1.2", "text": "B.",
		"kialo_parents": []any{map[string]any{"claim_id": "1.1", "relation": "pro"}},
		"ancestors":     []any{map[string]any{"claim_id": "1.1", "relation_of_child": "pro", "text": "A."}},
		"siblings":      []any{}, "children": []any{}}
	for k, v := range set {
		if v == absentField {
			delete(second, k)
		} else {
			second[k] = v
		}
	}
	b, err := json.Marshal(map[string]any{
		"discussion": map[string]any{"thesis_claim_id": "1.1"},
		"items": []any{map[string]any{"kialo_claim_id": "1.1", "text": "A.", "kialo_parents": []any{},
			"ancestors": []any{}, "siblings": []any{}, "children": []any{}}, second}})
	if err != nil {
		t.Fatal(err)
	}
	return FromDir(dataset(t, map[string]string{"items.json": string(b)}))
}

func TestSC7AContextTheChecksCannotReadIsRefused(t *testing.T) {
	for _, f := range []string{"kialo_parents", "ancestors", "siblings", "children"} {
		for _, v := range []any{"not a list", 1, true, map[string]any{"claim_id": "1.1"}, "", map[string]any{}, 0, false} {
			_, err := contextDataset(t, map[string]any{f: v})
			if want := "items.json: 1.2: " + f + " is not a list"; err == nil || err.Error() != want {
				t.Errorf("%s = %#v: %v, want %q", f, v, err, want)
			}
		}
	}
	good := map[string]any{"claim_id": "1.1", "relation": "pro"}
	bad := []any{"1.1", nil}
	for _, key := range []string{"claim_id", "relation"} {
		for _, v := range []any{absentField, nil, 3} {
			e := map[string]any{"claim_id": "1.1", "relation": "pro", "location_id": "1.9"}
			if v == absentField {
				delete(e, key)
			} else {
				e[key] = v
			}
			bad = append(bad, e)
		}
	}
	for _, e := range bad {
		_, err := contextDataset(t, map[string]any{"kialo_parents": []any{good, e}})
		want := "items.json: 1.2: kialo_parents entry 2 is not a Kialo parent (an object with a text claim_id and relation)"
		if err == nil || err.Error() != want {
			t.Errorf("parent %#v: %v, want %q", e, err, want)
		}
	}
	// An entry of ancestors, siblings or children that is not a claim is
	// R19's to skip (CT16): the dataset is read.
	if _, err := contextDataset(t, map[string]any{"siblings": []any{"junk", map[string]any{"claim_id": "1.7"}}}); err != nil {
		t.Errorf("a sibling that is not a claim: %v", err)
	}
}

func TestSC8AContextNullOrAbsentIsNone(t *testing.T) {
	d, err := contextDataset(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := d.kialoParents("1.2"); !ok || strings.Join(got, ", ") != "1.1 pro" {
		t.Fatalf("the well-formed parents: %v %v", got, ok)
	}
	for _, f := range []string{"kialo_parents", "ancestors", "siblings", "children"} {
		for _, v := range []any{nil, absentField} {
			d, err := contextDataset(t, map[string]any{f: v})
			if err != nil {
				t.Errorf("%s = %v: %v", f, v, err)
				continue
			}
			if got, ok := d.kialoParents("1.2"); f == "kialo_parents" && (!ok || len(got) != 0) {
				t.Errorf("%s = %v: parents %v %v, want none", f, v, got, ok)
			}
		}
	}
	// A dataset not made from Kialo: its claims have none of the four.
	d, err = FromDir(dataset(t, map[string]string{"items.json": `{"discussion": {}, "items": [{"kialo_claim_id": "e.1", "text": "An argument.", "debate": "A debate"}]}`}))
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := d.kialoParents("e.1"); !ok || len(got) != 0 {
		t.Errorf("a claim with no context: parents %v %v, want none", got, ok)
	}
}

func TestSC9TheContextIsReadByItsNamesAsWritten(t *testing.T) {
	// Parents written only under another spelling are not the claim's parents,
	// well formed or not (before, R7 read the first and skipped the claim on
	// the second).
	for _, v := range []any{[]any{map[string]any{"claim_id": "1.1", "relation": "pro"}}, "junk"} {
		d, err := contextDataset(t, map[string]any{"kialo_parents": absentField, "Kialo_Parents": v})
		if err != nil {
			t.Errorf("Kialo_Parents = %#v: %v", v, err)
			continue
		}
		if got, ok := d.kialoParents("1.2"); !ok || len(got) != 0 {
			t.Errorf("Kialo_Parents = %#v: parents %v %v, want none", v, got, ok)
		}
	}
	// So with a parent's claim_id spelt Claim_ID: it has none, as validate.py reads it.
	_, err := contextDataset(t, map[string]any{"kialo_parents": []any{map[string]any{"Claim_ID": "1.1", "relation": "pro"}}})
	if want := "items.json: 1.2: kialo_parents entry 1 is not a Kialo parent (an object with a text claim_id and relation)"; err == nil || err.Error() != want {
		t.Errorf("Claim_ID: %v, want %q", err, want)
	}
}

func TestSC10AClaimHasAText(t *testing.T) {
	for _, v := range []any{absentField, nil, 3} {
		_, err := contextDataset(t, map[string]any{"text": v})
		if want := "items.json: item 2 (1.2) has no text"; err == nil || err.Error() != want {
			t.Errorf("text = %#v: %v, want %q", v, err, want)
		}
	}
	// The lone spelling Text is read as the text (TestALoneSpellingLoads).
	d, err := contextDataset(t, map[string]any{"text": absentField, "Text": "B."})
	if err != nil || d.Items[1].Text != "B." {
		t.Errorf("Text: %v %+v", err, d)
	}
}
