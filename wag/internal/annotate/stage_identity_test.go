package annotate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/llm"
)

func TestStageAttachedProfileCannotBypassIdentity(t *testing.T) {
	ds, _ := Builtin()
	d := ds[0]
	p := llm.Profile{Name: "ann", Provider: "mock", Params: llm.Params{Model: "same"}}
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{p}, Items: d.IDs()[:1]}, nil)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	changed := p
	changed.System = "A different annotation instruction"
	c := llm.WithProvider(changed, &llm.Mock{Fn: func(llm.Request) (*llm.Response, error) { calls++; return &llm.Response{Text: `{"instances":[]}`}, nil }})
	r.Attach(p.Name, c)
	err = r.Annotate(context.Background(), nil)
	if err == nil && calls != 0 {
		t.Fatalf("a mismatched attached profile made %d calls", calls)
	}
	if calls != 0 {
		t.Fatal("identity rejection happened after a provider call")
	}
}

func TestStageSchemaContentsPinnedAndReadableAfterRemoval(t *testing.T) {
	ds, _ := Builtin()
	d := ds[0]
	dir := t.TempDir()
	schema := filepath.Join(dir, "schema.json")
	original := `{"type":"object","required":["instances"]}`
	if err := os.WriteFile(schema, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	p := llm.Profile{Name: "ann", Provider: "mock", SchemaFile: schema, Params: llm.Params{Model: "m", Output: "schema"}}
	r, err := New(filepath.Join(dir, "run"), d, Settings{Annotators: []llm.Profile{p}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(r.Dir, "run.json"))
	var saved struct {
		Settings Settings `json:"settings"`
	}
	_ = json.Unmarshal(b, &saved)
	if len(saved.Settings.Annotators[0].Schema) == 0 {
		t.Error("actual schema content was not pinned")
	}
	if err := r.CheckProfile(p); err != nil {
		t.Fatalf("same schema refused: %v", err)
	}
	if err := os.WriteFile(schema, []byte(`{"type":"object","required":["other"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := r.CheckProfile(p); err == nil {
		t.Error("changed schema bytes at same path accepted")
	}
	if err := os.Remove(schema); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(r.Dir, func(string, string) (*Dataset, error) { return d, nil }); err != nil {
		t.Fatalf("inspection unnecessarily needs schema path: %v", err)
	}
}

func TestStageRemovingScoringGoldDoesNotChangeInputs(t *testing.T) {
	dir := t.TempDir()
	items := `{"items":[{"kialo_claim_id":"1.1","text":"Claim"}]}`
	if err := os.WriteFile(filepath.Join(dir, "items.json"), []byte(items), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "gold.json"), []byte(`{"items":[{"kialo_claim_id":"1.1","instances":[]}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := FromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	p := llm.Profile{Name: "ann", Provider: "mock"}
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{p}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "gold.json")); err != nil {
		t.Fatal(err)
	}
	loaded, err := Open(r.Dir, func(string, string) (*Dataset, error) { return FromDir(dir) })
	if err != nil {
		t.Fatal(err)
	}
	if loaded.InputDrift() != "" {
		t.Fatalf("scoring-only removal blocks annotation: %s", loaded.InputDrift())
	}
}

func TestStageGoldAliasCannotHideClaimInputChanges(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dataset.json"), []byte(`{"gold":"items.json"}`), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "items.json")
	if err := os.WriteFile(path, []byte(`{"items":[{"kialo_claim_id":"1.1","text":"Before","instances":[]}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := FromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{{Name: "ann", Provider: "mock"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"items":[{"kialo_claim_id":"1.1","text":"After","instances":[]}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Open(r.Dir, func(string, string) (*Dataset, error) { return FromDir(dir) })
	if err != nil {
		t.Fatal(err)
	}
	if loaded.InputDrift() == "" {
		t.Error("gold alias hid a changed claim supplied to the model")
	}
}

func TestStageCorruptAnnotationFileIsNotSilentlyResumed(t *testing.T) {
	ds, _ := Builtin()
	d := ds[0]
	p := llm.Profile{Name: "ann", Provider: "mock"}
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{p}, Items: d.IDs()[:1]}, nil)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(r.Dir, "annotations", "ann.json")
	broken := []byte(`{"items":[`)
	if err := os.WriteFile(path, broken, 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Open(r.Dir, func(string, string) (*Dataset, error) { return d, nil })
	if err != nil {
		t.Fatalf("inspection should still open: %v", err)
	}
	calls := 0
	loaded.Attach("ann", llm.WithProvider(p, &llm.Mock{Fn: func(llm.Request) (*llm.Response, error) { calls++; return &llm.Response{Text: `{"instances":[]}`}, nil }}))
	if err := loaded.Annotate(context.Background(), nil); err == nil {
		t.Error("corrupt saved annotations were treated as missing work")
	}
	after, _ := os.ReadFile(path)
	if calls != 0 || string(after) != string(broken) {
		t.Fatalf("corrupt source overwritten/called: calls=%d bytes=%s", calls, after)
	}
}

func TestStageReportRefreshesDriftAndPinsEvaluationInputs(t *testing.T) {
	ds, _ := Builtin()
	d := ds[0]
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{{Name: "ann", Provider: "mock"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	d.Hashes["items.json"] = "current-evaluation-input"
	rep, err := r.Report()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rep.InputDrift, "items.json") {
		t.Error("report claims no drift after its evaluation inputs changed")
	}
	b, _ := json.Marshal(rep)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	h, _ := m["evaluation_inputs_sha256"].(map[string]any)
	if h["items.json"] != "current-evaluation-input" {
		t.Error("report does not identify the input revision used by its current checker")
	}
}

func TestStageAttachedParameterMutationIsRejectedBeforeCalling(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	p := llm.Profile{Name: "ann", Provider: "mock", Params: llm.Params{Model: "alias", Extra: map[string]any{"mode": "original"}}}
	calls := 0
	c := llm.WithProvider(p, &llm.Mock{Fn: func(llm.Request) (*llm.Response, error) { calls++; return &llm.Response{Text: `{"instances":[]}`}, nil }})
	r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{p}, Items: d.IDs()[:1]}, map[string]*llm.Client{"ann": c})
	if err != nil {
		t.Fatal(err)
	}
	c.Profile.Extra["mode"] = "changed"
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Error("mutated attached parameters continued a run")
	}
	if calls != 0 {
		t.Fatal("changed parameters reached a provider")
	}
	if r.Settings.Annotators[0].Extra["mode"] != "original" {
		t.Fatal("caller mutation also changed the saved identity")
	}
}

func TestStageReviewerServedModelsRemainSeparateAfterOpen(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	p := llm.Profile{Name: "same", Provider: "mock", Params: llm.Params{Model: "requested-alias"}}
	r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{p}, Reviewer: &p}, nil)
	if err != nil {
		t.Fatal(err)
	}
	log := strings.Join([]string{
		`{"role":"annotator","name":"same","call":{"response":{"model":"annotation-v1"}}}`,
		`{"role":"reviewer","name":"same","call":{"response":{"model":"review-v2"}}}`,
		`{"role":"reviewer","name":"same","call":{"response":{}}}`,
		`{"role":"reviewer","name":"same","call":{"error":"unavailable"}}`,
		`{"role":"reviewer","name":"same","call":`,
	}, "\n")
	if err := os.WriteFile(filepath.Join(r.Dir, "calls.jsonl"), []byte(log), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Open(r.Dir, func(string, string) (*Dataset, error) { return d, nil })
	if err != nil {
		t.Fatal(err)
	}
	rep, err := loaded.Report()
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		Reviewer struct {
			Model    string         `json:"model"`
			Provider string         `json:"provider"`
			Served   map[string]int `json:"served_models"`
			Unknown  int            `json:"unknown_served_responses"`
		} `json:"reviewer"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Reviewer.Model != "requested-alias" || out.Reviewer.Provider != "mock" || out.Reviewer.Served["review-v2"] != 1 || out.Reviewer.Unknown != 1 {
		t.Fatalf("reviewer attribution: %s", b)
	}
	if len(out.Reviewer.Served) != 1 || len(rep.Annotators[0].ServedModels) != 1 || rep.Annotators[0].ServedModels["annotation-v1"] != 1 || rep.UnreadableCallLines != 1 {
		t.Fatalf("cross-role or damage attribution: %s", b)
	}
}

func TestStageLaterVerdictRecordsActualEvaluationContext(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{{Name: "ann", Provider: "mock"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(r.Dir, "run.json")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	d.Hashes["items.json"] = "later-evaluation"
	if err := r.SetVerdict(Verdict{Item: d.IDs()[0], Choice: "none", By: "reviewer"}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(r.Dir, "review", "human.json"))
	if err != nil {
		t.Fatal(err)
	}
	var out []struct {
		Inputs map[string]string `json:"inputs_sha256"`
		Drift  string            `json:"input_drift"`
		Build  string            `json:"assessment_build"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Inputs["items.json"] != "later-evaluation" || !strings.Contains(out[0].Drift, "items.json") || out[0].Build == "" {
		t.Fatalf("later verdict lost context: %s", b)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("later verdict rewrote the historical run identity")
	}
}

func TestStageProfileIdentityPreservesExactExtraNumbers(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	p := llm.Profile{Name: "ann", Provider: "mock", Params: llm.Params{Extra: map[string]any{"seed": json.Number("9007199254740993")}}}
	r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{p}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Open(r.Dir, func(string, string) (*Dataset, error) { return d, nil })
	if err != nil {
		t.Fatal(err)
	}
	if err := loaded.CheckProfile(p); err != nil {
		t.Fatal(err)
	}
	p.Extra = map[string]any{"seed": json.Number("9007199254740992")}
	if err := loaded.CheckProfile(p); err == nil {
		t.Fatal("distinct large numeric parameters collapsed to one identity")
	}
}

func TestStageDistinctAnnotatorsCannotShareResultStorage(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	for _, names := range [][]string{{"model/a", "model_a"}, {".../", "..._"}} {
		_, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{{Name: names[0], Provider: "mock"}, {Name: names[1], Provider: "mock"}}}, nil)
		if err == nil {
			t.Fatalf("%v can overwrite each other's annotation results", names)
		}
	}
	reviewer := llm.Profile{Name: "judge", Provider: "mock"}
	if _, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{{Name: "review:judge", Provider: "mock"}}, Reviewer: &reviewer}, nil); err == nil {
		t.Fatal("annotator and reviewer share a progress identity")
	}
}
