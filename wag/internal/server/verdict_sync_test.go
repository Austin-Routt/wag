package server

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/annotate"
	"github.com/yourorg/wag/internal/fsutil"
	"github.com/yourorg/wag/internal/llm"
)

// A verdict in its file whose folder could not be synced is given, with a
// warning (the review's EV6): the answer is the claim with the verdict and
// the warning, the run's status says the verdicts are not synced yet, and
// the run's write-now action (flush) syncs the folder again.
func TestWorkbenchVerdictWhoseFolderCannotBeSynced(t *testing.T) {
	srv, ts := newTestServerWith(t)
	store := t.TempDir()
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	ds, _ := annotate.Builtin()
	srv.AddModel(llm.WithProvider(llm.Profile{Name: "one", Provider: "mock", Params: llm.Params{Model: "m-one"}}, &llm.Mock{Fn: func(llm.Request) (*llm.Response, error) {
		return nil, errors.New("not asked")
	}}))
	var made struct {
		Status annotate.Status `json:"status"`
	}
	if code := do(t, ts, "POST", "/api/annotate/runs", map[string]any{"dataset": ds[0].ID, "annotators": []string{"one"}, "count": 1}, &made); code != 200 {
		t.Fatalf("new run: %d", code)
	}
	id, item := made.Status.ID, made.Status.Settings.Items[0]
	review := filepath.Join(store, "annotation", id, "review")
	old := fsutil.SyncDir
	fail := true
	fsutil.SyncDir = func(d string) error {
		if fail && d == review {
			return errors.New("I/O error")
		}
		return old(d)
	}
	t.Cleanup(func() { fsutil.SyncDir = old })

	var got struct {
		annotate.ItemView
		Warning string `json:"warning"`
	}
	code := do(t, ts, "POST", "/api/annotate/runs/"+id+"/verdict", map[string]any{"item": item, "choice": "none", "note": "neither"}, &got)
	if code != http.StatusOK || got.Verdict == nil || got.Verdict.Choice != "none" || !strings.Contains(got.Warning, "power loss") {
		t.Fatalf("%d %+v %q", code, got.Verdict, got.Warning)
	}
	var s struct {
		Status annotate.Status `json:"status"`
	}
	do(t, ts, "GET", "/api/annotate/runs/"+id, nil, &s)
	if s.Status.Verdicts != 1 || !strings.Contains(s.Status.VerdictsNotSynced, "could not be synced") {
		t.Fatalf("status: %d verdicts, %q", s.Status.Verdicts, s.Status.VerdictsNotSynced)
	}
	if code := do(t, ts, "POST", "/api/annotate/runs/"+id+"/flush", map[string]any{}, nil); code != http.StatusConflict {
		t.Fatalf("write now while the folder cannot be synced: %d", code)
	}
	fail = false
	s.Status = annotate.Status{}
	if code := do(t, ts, "POST", "/api/annotate/runs/"+id+"/flush", map[string]any{}, &s); code != http.StatusOK || s.Status.ID != id || s.Status.VerdictsNotSynced != "" {
		t.Fatalf("write now: %d %q", code, s.Status.VerdictsNotSynced)
	}
	// A verdict written in full has no warning.
	got.Warning, got.Verdict = "", nil
	code = do(t, ts, "POST", "/api/annotate/runs/"+id+"/verdict", map[string]any{"item": item, "choice": "unsure"}, &got)
	if code != http.StatusOK || got.Verdict == nil || got.Verdict.Choice != "unsure" || got.Warning != "" {
		t.Fatalf("%d %+v %q", code, got.Verdict, got.Warning)
	}
}
