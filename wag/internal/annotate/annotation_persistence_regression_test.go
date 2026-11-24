package annotate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/wag/internal/llm"
)

// These are unresolved-regression probes, kept outside the application patch.
// All calls use the in-process mock provider; no credentials or network needed.
func TestAuditPersistenceFailureMustStopAndSurface(t *testing.T) {
	for _, failure := range []string{"annotation", "call-trace"} {
		t.Run(failure, func(t *testing.T) {
			ds, err := Builtin()
			if err != nil {
				t.Fatal(err)
			}
			d := ds[0]
			dir := filepath.Join(t.TempDir(), "run")
			items := d.IDs()[1:3]
			calls := 0
			provider := &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
				calls++
				b, err := json.Marshal(d.Gold[lastClaim(req)])
				if err != nil {
					t.Fatal(err)
				}
				return &llm.Response{Text: string(b)}, nil
			}}
			p := llm.Profile{Name: "solo", Provider: "mock", Params: llm.Params{Model: "m1"}}
			r, err := New(dir, d, Settings{Annotators: []llm.Profile{p}, Items: items}, map[string]*llm.Client{"solo": llm.WithProvider(p, provider)})
			if err != nil {
				t.Fatal(err)
			}
			blocked := filepath.Join(dir, "annotations", "solo.json.tmp")
			if failure == "call-trace" {
				blocked = filepath.Join(dir, "calls.jsonl")
			}
			if err := os.Mkdir(blocked, 0700); err != nil {
				t.Fatal(err)
			}
			err = r.Annotate(context.Background(), nil)
			st := r.Status()
			if err == nil || calls > 1 {
				t.Fatalf("persistence failure silently continued: error=%v calls=%d done=%d failed=%d last_error=%q", err, calls, st.Progress[0].Done, st.Progress[0].Failed, st.Progress[0].LastError)
			}
		})
	}
}

func TestAuditReviewSaveFailureMustStopAndSurface(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	dir := filepath.Join(t.TempDir(), "run")
	items := d.IDs()[1:3]
	pa := llm.Profile{Name: "solo", Provider: "mock", Params: llm.Params{Model: "m1"}}
	pr := llm.Profile{Name: "judge", Provider: "mock", Params: llm.Params{Model: "m2"}}
	copier := &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		b, _ := json.Marshal(d.Gold[lastClaim(req)])
		return &llm.Response{Text: string(b)}, nil
	}}
	calls := 0
	reviewer := &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		calls++
		return &llm.Response{Text: `{"fields":{"scheme":{"choice":"A","why":"fixture"}},"needs_human":true}`}, nil
	}}
	r, err := New(dir, d, Settings{Annotators: []llm.Profile{pa}, Reviewer: &pr, Items: items}, map[string]*llm.Client{"solo": llm.WithProvider(pa, copier), "judge": llm.WithProvider(pr, reviewer)})
	if err != nil {
		t.Fatal(err)
	}
	if err = r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(filepath.Join(dir, "review", "model-judge.json.tmp"), 0700); err != nil {
		t.Fatal(err)
	}
	err = r.Review(context.Background(), nil, false)
	if err == nil || calls > 1 {
		t.Fatalf("review persistence failure silently continued: error=%v calls=%d reviewed=%d", err, calls, r.Status().Reviewed)
	}
}
