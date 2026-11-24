package annotate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewAnnotationFilesRejectAmbiguousIdentities(t *testing.T) {
	for name, raw := range map[string]string{
		"duplicate claim":  `{"items":[{"kialo_claim_id":"1.1","instances":[]},{"kialo_claim_id":"1.1","instances":[]}]}`,
		"missing claim":    `{"items":[{"instances":[]}]}`,
		"nontext claim":    `{"items":[{"kialo_claim_id":3,"instances":[]}]}`,
		"null item":        `{"items":[null]}`,
		"missing items":    `{"annotator":"a"}`,
		"duplicate member": `{"items":[],"items":[{"kialo_claim_id":"1.1","instances":[]}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if got, err := parseAnnotations([]byte(raw)); err == nil {
				t.Fatalf("ambiguous annotation file accepted: %+v", got)
			}
		})
	}
	for _, raw := range []string{`{"items":[]}`, `{"items":null}`, `{"items":[{"kialo_claim_id":"1.1","instances":[]}]}`} {
		if _, err := parseAnnotations([]byte(raw)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReviewExplicitDatasetInputsMustExist(t *testing.T) {
	for _, field := range []string{"guidelines", "gold", "resources"} {
		t.Run(field, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "items.json"), []byte(`{"items":[{"kialo_claim_id":"1.1","text":"a"}]}`), 0600); err != nil {
				t.Fatal(err)
			}
			fields := map[string]any{field: "missing-file"}
			if field == "resources" {
				fields[field] = []string{"lexicon.json"}
			}
			manifest, _ := json.Marshal(fields)
			if err := os.WriteFile(filepath.Join(dir, "dataset.json"), manifest, 0600); err != nil {
				t.Fatal(err)
			}
			if d, err := FromDir(dir); err == nil {
				t.Fatalf("explicit missing %s silently replaced or omitted (has_gold=%v)", field, d.HasGold)
			}
		})
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "items.json"), []byte(`{"items":[{"kialo_claim_id":"1.1","text":"a"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if d, err := FromDir(dir); err != nil || d.Guidelines == "" || d.HasGold {
		t.Fatalf("default fallback changed: dataset=%+v err=%v", d, err)
	}
}

func TestReviewReplyDuplicateMembersRejected(t *testing.T) {
	for _, raw := range []string{
		`{"instances":false,"instances":[]}`,
		`{"instances":[],"kialo_claim_id":"wrong","kialo_claim_id":"1.1"}`,
		`{"instances":[{"scheme":{"walton_id":"none","walton_id":"sign"}}]}`,
		`{"instances":[],"note":{"x":1,"\u0078":2}}`,
	} {
		if a, err := parseAnnotation(raw, "1.1"); err == nil {
			t.Fatalf("duplicate members accepted: %+v", a)
		}
	}
	if _, err := parseAnnotation(`{"instances":[],"note":"{\"x\":1,\"x\":2}"}`, "1.1"); err != nil {
		t.Fatal(err)
	}
	if _, err := parseAnnotation(`{"instances":[],"note":`+strings.Repeat("[", 10001)+"0"+strings.Repeat("]", 10001)+"}", "1.1"); err == nil {
		t.Fatal("over-deep input accepted")
	}
}

func TestReviewFragmentCannotHideUnrelatedFormula(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	c, err := LoadCollection(d.WaltonYAML())
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(d.Gold["13292.108"])
	clone := func() Annotation {
		var a Annotation
		if err := json.Unmarshal(b, &a); err != nil {
			t.Fatal(err)
		}
		return a
	}
	if problems := c.Check(clone(), d); len(problems) != 0 {
		t.Fatalf("valid fragment rejected: %v", problems)
	}
	for name, mutate := range map[string]func(map[string]any){
		"premise": func(in map[string]any) {
			in["premises"].([]any)[0].(map[string]any)["formula"] = "unrelated_fact(anything)"
		},
		"conclusion": func(in map[string]any) {
			in["conclusion"].(map[string]any)["formula"] = "another_unrelated_fact(anything)"
		},
		"variable":                 func(in map[string]any) { in["bindings"].(map[string]any)["NOT_A_SCHEME_VARIABLE"] = "x" },
		"null undeclared variable": func(in map[string]any) { in["bindings"].(map[string]any)["NOT_A_SCHEME_VARIABLE"] = nil },
	} {
		t.Run(name, func(t *testing.T) {
			a := clone()
			mutate(a["instances"].([]any)[0].(map[string]any))
			if problems := c.Check(a, d); len(problems) == 0 {
				t.Fatal("invalid fragment passed")
			}
		})
	}
	// Fragments may still omit premises and a conclusion; they must not invent
	// unrelated formulas in the fields they do supply.
	a := clone()
	in := a["instances"].([]any)[0].(map[string]any)
	in["premises"] = []any{}
	in["conclusion"] = nil
	if problems := c.Check(a, d); len(problems) != 0 {
		t.Fatalf("partial fragment rejected: %v", problems)
	}
}
