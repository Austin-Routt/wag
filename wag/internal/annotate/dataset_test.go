package annotate

import (
	"bytes"
	"os"
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
		if p := c.Check(d.Gold[id], d); len(p) > 0 {
			t.Errorf("%s: %v", id, p)
		}
	}
	bad := Annotation{"instances": []any{map[string]any{"primary": true, "scheme": map[string]any{"walton_id": "sign", "fit": "explicit"},
		"bindings": map[string]any{"A": "x"}, "premises": []any{map[string]any{"role": "premise", "formula": "finding_true(y)"}},
		"conclusion": map[string]any{"formula": "true_in_situation(B)"}}}, "role": map[string]any{"kind": "guess"}}
	if p := c.Check(bad, d); len(p) < 4 {
		t.Fatalf("a bad annotation: %v", p)
	}
}
