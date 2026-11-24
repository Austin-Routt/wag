package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/annotate"
	"github.com/yourorg/wag/internal/llm"
)

// Follow-ups to the independent review of the integration
// (verify/r-critical-review-2026-11-23, INTEGRATION.md).

// otherWeightData is the built-in data with AS:analogy's base weight changed:
// other scheme data, on which the same moves are still legal.
func otherWeightData(t *testing.T) Data {
	t.Helper()
	var schemes map[string]map[string]any
	if err := json.Unmarshal(data.SchemesJSON, &schemes); err != nil {
		t.Fatal(err)
	}
	schemes["AS:analogy"]["base_weight"] = 0.9
	b, err := json.MarshalIndent(schemes, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return Data{YAML: data.CoreYAML, Schemes: b, CQs: data.CQsJSON, Edges: data.EdgesJSON, Contexts: data.ContextsJSON}
}

// A game opened read-only on other data is exported with its history as it
// was played (the replay's moves carry this data's weights).
func TestInspectedExportKeepsTheHistoryAsPlayed(t *testing.T) {
	sa, _ := newTestServerWith(t)
	played := reviewExport(t, sa, NewGameRequest{Mode: "classic", Seats: twoHumans})
	sb, err := New(otherWeightData(t), "")
	if err != nil {
		t.Fatal(err)
	}
	tb := httptest.NewServer(sb.Handler())
	defer tb.Close()
	code, st := loadHTTP(t, tb, played, "?inspect=1")
	if code != http.StatusOK {
		t.Fatalf("inspect: %d %v", code, st["error"])
	}
	var raw map[string]any
	call(t, tb, "GET", "/api/games/"+st["id"].(string)+"/export", nil, &raw)
	b, _ := json.Marshal(raw)
	var ex Export
	if err := json.Unmarshal(b, &ex); err != nil {
		t.Fatal(err)
	}
	if played.History[0].Weight == 0.9 {
		t.Fatal("fixture: the move was played at the other data's weight")
	}
	if ex.DataHash != played.DataHash || len(ex.History) != len(played.History) || ex.History[0].Weight != played.History[0].Weight ||
		ex.History[0].ScoreDelta != played.History[0].ScoreDelta {
		t.Fatalf("export %s: move 1 at weight %v (delta %v), played at %v (delta %v)", ex.DataHash, ex.History[0].Weight, ex.History[0].ScoreDelta,
			played.History[0].Weight, played.History[0].ScoreDelta)
	}
	// The re-export loads on the data it was played with, as the original.
	ta := httptest.NewServer(sa.Handler())
	defer ta.Close()
	if code, out := loadHTTP(t, ta, &ex, ""); code != http.StatusOK {
		t.Fatalf("the re-export on its own data: %d %v", code, out["error"])
	}
}

// The dashboard can write what a run's files lack, even when no claim is
// left to annotate (its start buttons are then disabled).
func TestTheDashboardWritesWhatIsNotSaved(t *testing.T) {
	srv, ts := newTestServerWith(t)
	store := t.TempDir()
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	ds, _ := annotate.Builtin()
	gold := ds[0].Gold
	srv.AddModel(llm.WithProvider(llm.Profile{Name: "one", Provider: "mock", Params: llm.Params{Model: "m-one"}}, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		m := claimRe.FindAllStringSubmatch(req.Messages[len(req.Messages)-1].Content, -1)
		b, _ := json.Marshal(gold[m[len(m)-1][1]])
		return &llm.Response{Text: string(b)}, nil
	}}))
	var made struct {
		Status annotate.Status `json:"status"`
	}
	if code := do(t, ts, "POST", "/api/annotate/runs", map[string]any{"dataset": ds[0].ID, "annotators": []string{"one"}, "count": 1}, &made); code != 200 {
		t.Fatalf("new run: %d", code)
	}
	id := made.Status.ID
	blocked := filepath.Join(store, "annotation", id, "annotations", "one.json.tmp")
	if err := os.MkdirAll(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
	status := func() annotate.Status {
		for i := 0; i < 200; i++ {
			var s struct {
				Status annotate.Status `json:"status"`
			}
			do(t, ts, "GET", "/api/annotate/runs/"+id, nil, &s)
			if s.Status.Busy == "" {
				return s.Status
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatal("the run does not finish")
		return annotate.Status{}
	}
	do(t, ts, "POST", "/api/annotate/runs/"+id+"/start", map[string]any{}, nil)
	st := status()
	if st.WriteError == "" || len(st.Unsaved) != 1 || st.NextItem != "" {
		t.Fatalf("fixture: %q %q next %q", st.WriteError, st.Unsaved, st.NextItem)
	}
	var e map[string]string
	if code := do(t, ts, "POST", "/api/annotate/runs/"+id+"/flush", map[string]any{}, &e); code != http.StatusConflict || !strings.Contains(e["error"], "cannot be written") {
		t.Fatalf("flush while blocked: %d %v", code, e)
	}
	if err := os.Remove(blocked); err != nil {
		t.Fatal(err)
	}
	var after struct {
		Status annotate.Status `json:"status"`
	}
	if code := do(t, ts, "POST", "/api/annotate/runs/"+id+"/flush", map[string]any{}, &after); code != 200 || after.Status.WriteError != "" ||
		len(after.Status.Unsaved) != 0 || after.Status.Progress[0].Done != 1 {
		t.Fatalf("flush: %d %+v", code, after.Status)
	}
	if code := do(t, ts, "GET", "/api/annotate/runs/"+id+"/flush", nil, nil); code != http.StatusMethodNotAllowed {
		t.Fatalf("GET flush: %d", code)
	}
}
