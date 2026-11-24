package annotate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/llm"
)

const oneItem = `{"discussion": {"thesis_claim_id": "1.1"}, "items": [{"kialo_claim_id": "1.2", "text": "A."}]}`

// dataset writes a dataset folder from name → content ("" for a directory).
func dataset(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if content == "" {
			if err := os.MkdirAll(p, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// A malformed reply keeps the decoder's own message, which is what the model
// is told when it retries: a trailing comma is not a nesting problem.
func TestDecoderMessageIsKept(t *testing.T) {
	_, err := jsonObject(`{"instances": [],}`)
	if err == nil || !strings.Contains(err.Error(), "invalid character") || strings.Contains(err.Error(), "nesting") {
		t.Fatalf("reply: %v", err)
	}
	_, err = FromDir(dataset(t, map[string]string{"items.json": `{"items": [],}`}))
	if err == nil || !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("items.json: %v", err)
	}
}

// Nesting deeper than encoding/json allows is refused before the duplicate
// scan, whose recursion has no limit of its own: without the pre-check a
// reply of a few million "[" overflows the stack on Go 1.24, a fatal error.
func TestDeepNestingIsRefusedWithoutCrash(t *testing.T) {
	deep := `{"a": ` + strings.Repeat("[", 1<<21) + `}`
	if _, err := jsonObject(deep); err == nil || !strings.Contains(err.Error(), "depth") {
		t.Fatalf("deep reply: %v", err)
	}
	ok := `{"a": ` + strings.Repeat("[", 9000) + strings.Repeat("]", 9000) + `}`
	if _, err := jsonObject(ok); err != nil {
		t.Fatalf("nesting within the limit: %v", err)
	}
}

func TestInvalidUTF8IsRefused(t *testing.T) {
	var v any
	if err := checkedJSON([]byte("{\"a\": \"\xff\"}"), &v); err == nil || !strings.Contains(err.Error(), "UTF-8") {
		t.Fatalf("invalid UTF-8: %v", err)
	}
}

// encoding/json matches a struct's fields case-insensitively, so a second
// spelling of a field would silently replace the first.
func TestVariantSpellingsAreRefused(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"dataset.json guidelines": {"items.json": oneItem, "guidelines.md": "G", "dataset.json": `{"guidelines": "guidelines.md", "Guidelines": ""}`},
		"dataset.json gold":       {"items.json": oneItem, "dataset.json": `{"gold": "missing.json", "GOLD": ""}`},
		"items.json items":        {"items.json": `{"items": [{"kialo_claim_id": "1.2", "text": "A."}], "Items": null}`},
		"items.json long s":       {"items.json": `{"items": [{"kialo_claim_id": "1.2", "text": "A."}], "itemſ": null}`},
		"an item's ID":            {"items.json": `{"items": [{"kialo_claim_id": "1.2", "Kialo_Claim_Id": "2.2", "text": "A."}]}`},
		"gold items":              {"items.json": oneItem, "gold.json": `{"items": [{"kialo_claim_id": "1.2", "instances": []}], "ITEMS": null}`},
	} {
		if _, err := FromDir(dataset(t, files)); err == nil || !strings.Contains(err.Error(), "another spelling") {
			t.Errorf("%s: %v", name, err)
		}
	}
	// A member that is no field's variant is left alone.
	if _, err := FromDir(dataset(t, map[string]string{"items.json": `{"items": [{"kialo_claim_id": "1.2", "text": "A.", "Text2": "x"}], "note": 1}`})); err != nil {
		t.Fatalf("an unrelated member: %v", err)
	}
}

func TestDatasetFilesRefuseDuplicateMembers(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"dataset.json": {"items.json": oneItem, "dataset.json": `{"name": "a", "name": "b"}`},
		"items.json":   {"items.json": `{"items": [], "items": [{"kialo_claim_id": "1.2", "text": "A."}]}`},
	} {
		if _, err := FromDir(dataset(t, files)); err == nil || !strings.Contains(err.Error(), "duplicate JSON member") {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// An ID of spaces is no ID, as in an annotation file: a run is never given
// claims it would then refuse to read back.
func TestWhitespaceItemIDIsRefused(t *testing.T) {
	_, err := FromDir(dataset(t, map[string]string{"items.json": `{"items": [{"kialo_claim_id": "  ", "text": "A."}]}`}))
	if err == nil || !strings.Contains(err.Error(), "no kialo_claim_id") {
		t.Fatalf("whitespace ID: %v", err)
	}
}

// A file that exists but cannot be read is an error, not an absent default.
func TestUnreadableInputsAreRefused(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"dataset.json":           {"items.json": oneItem, "dataset.json": ""},
		"the default guidelines": {"items.json": oneItem, "guidelines.md": ""},
		"the default gold":       {"items.json": oneItem, "gold.json": ""},
	} {
		if _, err := FromDir(dataset(t, files)); err == nil {
			t.Errorf("%s as a directory was accepted", name)
		}
	}
	if _, err := FromDir(dataset(t, map[string]string{"items.json": oneItem})); err != nil {
		t.Fatalf("absent defaults: %v", err)
	}
}

// Reviews and verdicts are read as strictly as annotations: an entry that
// names no claim, a second entry for a claim, or a field spelt two ways
// makes the run damaged, never silently dropped or replaced.
func TestReviewFilesAreReadStrictly(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	ids := d.IDs()[:2]
	x, y := ids[0], ids[1]
	newRun := func(t *testing.T) string {
		pa := llm.Profile{Name: "a", Provider: "mock", Params: llm.Params{Model: "m"}}
		pr := llm.Profile{Name: "judge", Provider: "mock", Params: llm.Params{Model: "m"}}
		dir := filepath.Join(t.TempDir(), "run")
		if _, err := New(dir, d, Settings{Annotators: []llm.Profile{pa}, Reviewer: &pr, Items: ids}, nil); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	open := func(t *testing.T, dir string) *Run {
		r, err := Open(dir, func(string, string) (*Dataset, error) { return d, nil })
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	at := `"at": "2026-11-24T00:00:00Z"`
	for _, tc := range []struct{ name, file, content, want string }{
		{"two verdicts", "human.json", `[{"item": "` + x + `", "choice": "a", ` + at + `}, {"item": "` + x + `", "choice": "none", ` + at + `}]`, "two verdicts for " + x},
		{"a verdict with no claim", "human.json", `[{"choice": "a", ` + at + `}]`, "verdict 1 names no claim"},
		{"a null verdict", "human.json", `[null]`, "verdict 1 names no claim"},
		{"a verdict's claim spelt two ways", "human.json", `[{"item": "` + x + `", "Item": "` + y + `", "choice": "a", ` + at + `}]`, "another spelling"},
		{"a verdict's claim given twice", "human.json", `[{"item": "` + x + `", "item": "` + y + `", "choice": "a", ` + at + `}]`, "duplicate JSON member"},
		{"two reviews", "model-judge.json", `[{"item": "` + x + `", "reviewer": "judge"}, {"item": "` + x + `", "reviewer": "judge"}]`, "two reviews of " + x},
		{"a review with no claim", "model-judge.json", `[{"reviewer": "judge"}]`, "review 1 names no claim"},
		{"a review's claim spelt two ways", "model-judge.json", `[{"item": "` + x + `", "ITEM": "` + y + `"}]`, "another spelling"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := newRun(t)
			if err := os.WriteFile(filepath.Join(dir, "review", tc.file), []byte(tc.content), 0o644); err != nil {
				t.Fatal(err)
			}
			r := open(t, dir)
			err := r.checkInputs()
			if !r.Damaged() || err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("damaged %v, %v; want %q", r.Damaged(), err, tc.want)
			}
		})
	}
	// Valid files still load: one verdict for each of two claims.
	dir := newRun(t)
	valid := `[{"item": "` + x + `", "choice": "a", ` + at + `}, {"item": "` + y + `", "choice": "none", ` + at + `}]`
	if err := os.WriteFile(filepath.Join(dir, "review", "human.json"), []byte(valid), 0o644); err != nil {
		t.Fatal(err)
	}
	if r := open(t, dir); r.Damaged() || r.Status().Verdicts != 2 {
		t.Fatalf("valid verdicts: damaged %v, %+v", r.Damaged(), r.Status())
	}
}
