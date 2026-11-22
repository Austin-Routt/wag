package annotate

import (
	"context"
	"encoding/json"
	"github.com/yourorg/wag/internal/llm"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A run whose inputs have changed opens, so that it can still be read, given
// verdicts and reported on, but it cannot be continued, and saving it keeps
// the identity it was made with.
func TestConvergenceResumeRejectsChangedInputs(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	p := llm.Profile{Name: "m", Provider: "mock", Params: llm.Params{Model: "requested"}}
	r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{p}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	pinned := func() map[string]string {
		var f struct {
			Inputs map[string]string `json:"inputs_sha256"`
		}
		b, _ := os.ReadFile(filepath.Join(r.Dir, "run.json"))
		if err := json.Unmarshal(b, &f); err != nil {
			t.Fatal(err)
		}
		return f.Inputs
	}
	original := pinned()
	changed := *d
	changed.Hashes = map[string]string{}
	for k, v := range d.Hashes {
		changed.Hashes[k] = v
	}
	changed.Hashes["items.json"] = "changed"
	r2, err := Open(r.Dir, func(string, string) (*Dataset, error) { return &changed, nil })
	if err != nil {
		t.Fatalf("a run with changed inputs must still open for reading: %v", err)
	}
	if drift := r2.InputDrift(); !strings.Contains(drift, "items.json") {
		t.Errorf("drift %q does not name items.json", drift)
	}
	if err := r2.Annotate(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "create a new run") {
		t.Errorf("continued a run whose inputs changed: %v", err)
	}
	rep, err := r2.Report()
	if err != nil || !strings.Contains(rep.InputDrift, "items.json") {
		t.Errorf("report on a drifted run: %v, drift %q", err, rep.InputDrift)
	}
	if err := r2.saveRun(); err != nil {
		t.Fatal(err)
	}
	if got := pinned(); got["items.json"] != original["items.json"] || got["items.json"] == "changed" {
		t.Errorf("saving overwrote the run's identity: %v", got)
	}
}

func TestConvergenceResumeRejectsPromptDrift(t *testing.T) {
	ds, _ := Builtin()
	d := ds[0]
	p := llm.Profile{Name: "m", Provider: "mock"}
	r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{p}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	changed := *d
	changed.Guidelines += "\nNew instruction"
	r2, err := Open(r.Dir, func(string, string) (*Dataset, error) { return &changed, nil })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r2.InputDrift(), "system prompt") {
		t.Errorf("drift %q does not name the prompt", r2.InputDrift())
	}
	if err := r2.Annotate(context.Background(), nil); err == nil {
		t.Error("continued a run whose prompt changed")
	}
}

// The gold scores a run but is never shown to an annotator: correcting it
// does not stop a run.
func TestConvergenceGoldIsNotPartOfIdentity(t *testing.T) {
	ds, _ := Builtin()
	d := ds[0]
	r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{{Name: "m", Provider: "mock"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if d.GoldFile == "" {
		t.Fatal("the built-in dataset names no gold file")
	}
	changed := *d
	changed.Hashes = map[string]string{}
	for k, v := range d.Hashes {
		changed.Hashes[k] = v
	}
	changed.Hashes[d.GoldFile] = "corrected"
	r2, err := Open(r.Dir, func(string, string) (*Dataset, error) { return &changed, nil })
	if err != nil || r2.InputDrift() != "" {
		t.Fatalf("a gold correction stopped the run: %v %q", err, r2.InputDrift())
	}
	rep, err := r2.Report()
	if err != nil || rep.GoldSHA256 != "corrected" {
		t.Errorf("report gold %q, %v", rep.GoldSHA256, err)
	}
}

// A run made before inputs were pinned can be read, not continued.
func TestConvergenceLegacyRunIsReadOnly(t *testing.T) {
	ds, _ := Builtin()
	d := ds[0]
	r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{{Name: "m", Provider: "mock"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(r.Dir, "run.json")
	var f map[string]any
	b, _ := os.ReadFile(path)
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	delete(f, "inputs_sha256")
	delete(f, "annotator_system_prompt_sha256")
	b, _ = json.Marshal(f)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	r2, err := Open(r.Dir, func(string, string) (*Dataset, error) { return d, nil })
	if err != nil {
		t.Fatalf("a legacy run must open for reading: %v", err)
	}
	if !strings.Contains(r2.InputDrift(), "before its inputs were pinned") {
		t.Errorf("drift %q", r2.InputDrift())
	}
	if err := r2.Annotate(context.Background(), nil); err == nil {
		t.Error("continued a run with no pinned identity")
	}
}

// A damaged last line of calls.jsonl (a crash mid-write) is counted, and the
// report still comes out.
func TestConvergenceReportSurvivesDamagedCallLog(t *testing.T) {
	ds, _ := Builtin()
	d := ds[0]
	r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{{Name: "m", Provider: "mock"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	log := `{"role":"annotator","name":"m","call":{"response":{"model":"served-1"}}}` + "\n" + `{"role":"annotator","name":"m","call":{"resp`
	if err := os.WriteFile(filepath.Join(r.Dir, "calls.jsonl"), []byte(log), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, err := r.Report()
	if err != nil {
		t.Fatalf("a damaged call log stopped the report: %v", err)
	}
	if rep.UnreadableCallLines != 1 || rep.Annotators[0].ServedModels["served-1"] != 1 {
		t.Errorf("unreadable %d, served %v", rep.UnreadableCallLines, rep.Annotators[0].ServedModels)
	}
}

// Continuing a run under an annotator's name with another model is refused;
// operational settings may change.
func TestConvergenceProfileMustMatchTheRun(t *testing.T) {
	ds, _ := Builtin()
	d := ds[0]
	p := llm.Profile{Name: "m", Provider: "mock", Params: llm.Params{Model: "m1"}}
	r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{p}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	other := p
	other.Params.Model = "m2"
	if err := r.CheckProfile(other); err == nil {
		t.Error("accepted another model under the same name")
	}
	slower := p
	slower.Retries = 7
	slower.APIKeyEnv = "OTHER_KEY"
	if err := r.CheckProfile(slower); err != nil {
		t.Errorf("refused a change of retries and key source: %v", err)
	}
	if err := r.CheckProfile(llm.Profile{Name: "new", Provider: "mock"}); err != nil {
		t.Errorf("refused a profile the run does not have: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(r.Dir, "run.json"))
	if !strings.Contains(string(b), `"builds"`) {
		t.Error("run.json does not record the build")
	}
}

func TestConvergenceMixedCaseVariables(t *testing.T) {
	got := subst("last(Cn,Al,A1)", map[string]string{"Cn": "last_case", "Al": "agent", "A1": "first"})
	if got != "last(last_case,agent,first)" {
		t.Fatal(got)
	}
}
func TestConvergenceServedModelReport(t *testing.T) {
	ds, _ := Builtin()
	d := ds[0]
	p := llm.Profile{Name: "m", Provider: "mock", Params: llm.Params{Model: "alias"}}
	c := llm.WithProvider(p, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		b, _ := json.Marshal(d.Gold[lastClaim(req)])
		return &llm.Response{Text: string(b), Model: "actual-v2"}, nil
	}})
	r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{p}, Items: d.IDs()[:1]}, map[string]*llm.Client{"m": c})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(r.Dir, "report.json"))
	if !strings.Contains(string(b), "actual-v2") {
		t.Fatal("report omits served model:", string(b))
	}
	r2, err := Open(r.Dir, func(string, string) (*Dataset, error) { return d, nil })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r2.Report(); err != nil {
		t.Fatal(err)
	}
	b, _ = os.ReadFile(filepath.Join(r.Dir, "report.json"))
	if !strings.Contains(string(b), "actual-v2") {
		t.Fatal("reopened report omits served model")
	}
}

func TestConvergenceAssignedDenominator(t *testing.T) {
	a := map[string]Annotation{"one": {"role": map[string]any{"kind": "none"}}}
	b := map[string]Annotation{"one": {"role": map[string]any{"kind": "none"}}}
	p := AgreeAssigned("A", "B", a, b, []string{"one", "two", "three"})
	if p.Items != 1 || p.Coverage.Assigned != 3 || p.Coverage.Neither != 2 || p.Fields["role"].Agreement != 1 || p.AllAssignedAgreement["role"] != 1.0/3.0 {
		t.Fatalf("%+v", p)
	}
}

func TestConvergenceIndependentFirstPass(t *testing.T) {
	for _, independent := range []bool{false, true} {
		ds, _ := Builtin()
		d := ds[0]
		p := llm.Profile{Name: "m", Provider: "mock"}
		calls := 0
		c := llm.WithProvider(p, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
			history := strings.Contains(req.Messages[0].Content, "Your annotations so far")
			if history != (calls > 0 && !independent) {
				t.Errorf("independent=%v call=%d history=%v", independent, calls, history)
			}
			calls++
			b, _ := json.Marshal(d.Gold[lastClaim(req)])
			return &llm.Response{Text: string(b)}, nil
		}})
		r, err := New(filepath.Join(t.TempDir(), "r"), d, Settings{Annotators: []llm.Profile{p}, Items: d.IDs()[:2], IndependentFirstPass: independent}, map[string]*llm.Client{"m": c})
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Annotate(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
		reopened, err := Open(r.Dir, func(string, string) (*Dataset, error) { return d, nil })
		if err != nil {
			t.Fatal(err)
		}
		if reopened.Settings.IndependentFirstPass != independent {
			t.Fatal("mode lost")
		}
	}
}
