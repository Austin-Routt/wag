package annotate

import (
	"encoding/json"
	"strings"
	"testing"
)

// Cases SC11-SC14 and SC16 (docs/d086-first-gold.md §6; the independent check
// of F3, F5 and F6): items.json is read by its names as written, as
// validate.py reads it; null and {} are no list of items; R19 reads a context
// entry's claim_id as R7 does; R7 reports a claim that items.json gives no
// Kialo parent while its record names one. validate.py's side is in
// docs/kialo-batch2/test_validate.py (SourceShape, SC15 and SC16).

func TestSC11ItemsThatAreNoList(t *testing.T) {
	for _, body := range []string{`null`, `{}`, `{"items": null}`, `[]`, `{"items": {}}`, `{"items": "x"}`} {
		_, err := FromDir(dataset(t, map[string]string{"items.json": body}))
		if err == nil {
			t.Errorf("%s: read", body)
		}
	}
	for _, body := range []string{`null`, `{}`, `{"items": null}`} {
		_, err := FromDir(dataset(t, map[string]string{"items.json": body}))
		if want := "items.json: its items are not a list of objects"; err == nil || err.Error() != want {
			t.Errorf("%s: %v, want %q", body, err, want)
		}
	}
}

func TestSC12ItemsNamesAsWritten(t *testing.T) {
	item := `{"kialo_claim_id": "1.2", "text": "Meat is bad for the climate."}`
	for name, body := range map[string]string{
		"Items":                  `{"Items": [` + item + `]}`,
		"KIALO_CLAIM_ID":         `{"items": [{"KIALO_CLAIM_ID": "1.2", "text": "Meat is bad."}]}`,
		"Kelvin sign":            `{"items": [{"\u212aialo_claim_id": "1.2", "text": "Meat is bad."}]}`,
		"Text":                   `{"items": [{"kialo_claim_id": "1.2", "Text": "Meat is bad."}]}`,
		"text and TEXT":          `{"items": [{"kialo_claim_id": "1.2", "text": "a", "TEXT": "b"}]}`,
		"items and the long s":   `{"items": [` + item + `], "item\u017f": []}`,
		"kialo_claim_id twice":   `{"items": [{"kialo_claim_id": "1.2", "Kialo_Claim_Id": "1.3", "text": "a"}]}`,
		"discussion and another": `{"discussion": {}, "Discussion": {}, "items": [` + item + `]}`,
	} {
		if _, err := FromDir(dataset(t, map[string]string{"items.json": body})); err == nil {
			t.Errorf("%s: read", name)
		}
	}
	// As written, it is read.
	d, err := FromDir(dataset(t, map[string]string{"items.json": `{"items": [` + item + `]}`}))
	if err != nil || len(d.Items) != 1 || d.Items[0].Text != "Meat is bad for the climate." {
		t.Errorf("as written: %v %+v", err, d)
	}
}

func TestSC16IdsAsBothToolsReadThem(t *testing.T) {
	read := func(ids ...string) error {
		var items []any
		for _, id := range ids {
			items = append(items, map[string]any{"kialo_claim_id": id, "text": "A claim."})
		}
		b, _ := json.Marshal(map[string]any{"items": items})
		_, err := FromDir(dataset(t, map[string]string{"items.json": string(b)}))
		return err
	}
	for _, id := range []string{"\u00a0", "\u3000", "\t\u00a0"} {
		if err := read(id); err == nil || !strings.Contains(err.Error(), "has no kialo_claim_id") {
			t.Errorf("%q: %v", id, err)
		}
	}
	if err := read("\u001c"); err != nil {
		t.Errorf("U+001C: %v", err)
	}
	if err := read(" 1.2", "1.2"); err != nil {
		t.Errorf("an id with a space beside the same without: %v", err)
	}
	// Not UTF-8: refused, by Go's reader as by validate.py.
	if _, err := FromDir(dataset(t, map[string]string{"items.json": "{\"items\": [{\"kialo_claim_id\": \"1.2\", \"text\": \"\xff\"}]}"})); err == nil {
		t.Error("not UTF-8: read")
	}
}

func TestSC13R19ReadsClaimIDAsWritten(t *testing.T) {
	d, c := builtin(t)
	// 13292.218#1's inferred premise cites 13292.777777, which only a context
	// entry of the first claim names.
	cite := func(entry map[string]any) []string {
		set := clone(t, d.Gold)
		a := set["13292.218"]
		for _, x := range inst(a, 0)["premises"].([]any) {
			if pm := x.(map[string]any); pm["status"] == "inferred" {
				pm["cite"] = "13292.777777"
			}
		}
		dd := *d
		dd.Items = append([]Item(nil), d.Items...)
		var raw map[string]any
		if err := json.Unmarshal(dd.Items[0].Raw, &raw); err != nil {
			t.Fatal(err)
		}
		raw["siblings"] = []any{entry}
		b, _ := json.Marshal(raw)
		dd.Items[0].Raw = b
		var r19 []string
		for _, p := range c.CheckIn(a, &dd, set, 2) {
			if strings.HasPrefix(p, "R19 ") {
				r19 = append(r19, p)
			}
		}
		return r19
	}
	if got := cite(map[string]any{"claim_id": "13292.777777"}); len(got) != 0 {
		t.Errorf("claim_id: %v", got)
	}
	if got := cite(map[string]any{"CLAIM_ID": "13292.777777"}); len(got) != 1 {
		t.Errorf("CLAIM_ID alone: %v, want R19", got)
	}
	// Two spellings: claim_id is read (the cited claim), whatever the other holds.
	if got := cite(map[string]any{"claim_id": "13292.777777", "Claim_Id": 5}); len(got) != 0 {
		t.Errorf("claim_id beside another spelling: %v", got)
	}
	if got := cite(map[string]any{"claim_id": "13292.1", "CLAIM_ID": "13292.777777"}); len(got) != 1 {
		t.Errorf("another spelling holding the cited claim: %v, want R19", got)
	}
}

func TestSC14R7AClaimItemsJSONGivesNoParent(t *testing.T) {
	d, c := builtin(t)
	for _, parents := range []any{nil, absentField, []any{}} {
		dd := *d
		dd.Items = append([]Item(nil), d.Items...)
		var raw map[string]any
		if err := json.Unmarshal(dd.Items[5].Raw, &raw); err != nil {
			t.Fatal(err)
		}
		if parents == absentField {
			delete(raw, "kialo_parents")
		} else {
			raw["kialo_parents"] = parents
		}
		b, _ := json.Marshal(raw)
		dd.Items[5].Raw = b
		claim := dd.Items[5].ID
		a := d.Gold[claim]
		var r7 []string
		for _, p := range c.CheckIn(a, &dd, d.Gold, 2) {
			if strings.HasPrefix(p, "R7 ") {
				r7 = append(r7, p)
			}
		}
		if len(r7) != 1 || !strings.Contains(r7[0], "items.json gives none") {
			t.Errorf("%s with kialo_parents %#v: %v", claim, parents, r7)
		}
	}
}
