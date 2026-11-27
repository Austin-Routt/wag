package server

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/annotate"
	"github.com/yourorg/wag/internal/llm"
)

// The dashboard's part in the owner's verification (docs/d086-first-gold.md
// §5 and §6, cases D1 and D3; D2 is the page's, checked in the browser).

func d086Server(t *testing.T) (*Server, string, *annotate.Dataset, func(method, path string, body any, out any) int) {
	t.Helper()
	srv, ts := newTestServerWith(t)
	store := t.TempDir()
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	ds, err := annotate.Builtin()
	if err != nil {
		t.Fatal(err)
	}
	srv.AddModel(llm.WithProvider(llm.Profile{Name: "one", Provider: "mock", Params: llm.Params{Model: "m-one"}}, &llm.Mock{Fn: func(llm.Request) (*llm.Response, error) {
		return nil, errors.New("not asked")
	}}))
	return srv, store, ds[0], func(method, path string, body any, out any) int { return do(t, ts, method, path, body, out) }
}

type madeRun struct {
	Status annotate.Status `json:"status"`
}

// D1: a run made with a list of claims holds those claims, in that order; an
// unknown claim is refused and no run is made; with no list, the first N.
func TestD1ARunOfTheClaimsListed(t *testing.T) {
	_, store, d, call := d086Server(t)
	ids := d.IDs()
	want := []string{ids[20], ids[3], ids[11]}
	var made madeRun
	if code := call("POST", "/api/annotate/runs", map[string]any{"dataset": d.ID, "annotators": []string{"one"}, "items": want}, &made); code != http.StatusOK {
		t.Fatalf("new run with a list: %d", code)
	}
	if got := made.Status.Settings.Items; !reflect.DeepEqual(got, want) {
		t.Fatalf("the run holds %v, not %v", got, want)
	}
	runs := func() int {
		es, err := os.ReadDir(filepath.Join(store, "annotation"))
		if err != nil {
			t.Fatal(err)
		}
		return len(es)
	}
	before := runs()
	if code := call("POST", "/api/annotate/runs", map[string]any{"dataset": d.ID, "annotators": []string{"one"}, "items": []string{ids[0], d.ID + "-no-such-claim"}}, nil); code != http.StatusBadRequest {
		t.Fatalf("a list with an unknown claim: %d", code)
	}
	if n := runs(); n != before {
		t.Fatalf("a refused run left a folder: %d runs, %d before", n, before)
	}
	made = madeRun{}
	if code := call("POST", "/api/annotate/runs", map[string]any{"dataset": d.ID, "annotators": []string{"one"}, "count": 2}, &made); code != http.StatusOK {
		t.Fatalf("new run of the first 2: %d", code)
	}
	if got := made.Status.Settings.Items; !reflect.DeepEqual(got, ids[:2]) {
		t.Fatalf("the first 2: %v", got)
	}
}

// D3: a verdict sent with the input hashes of the page it was given on is
// taken when they are the run's, refused with nothing saved when one
// differs, and taken as before when none are sent.
func TestD3AVerdictGivenOnAStalePage(t *testing.T) {
	_, _, d, call := d086Server(t)
	var made madeRun
	if code := call("POST", "/api/annotate/runs", map[string]any{"dataset": d.ID, "annotators": []string{"one"}, "count": 2}, &made); code != http.StatusOK {
		t.Fatalf("new run: %d", code)
	}
	id, items := made.Status.ID, made.Status.Settings.Items
	var view annotate.ItemView
	if code := call("GET", "/api/annotate/runs/"+id+"/item/"+items[0], nil, &view); code != http.StatusOK {
		t.Fatalf("the claim's view: %d", code)
	}
	if len(view.InputsSHA256) == 0 || !reflect.DeepEqual(view.InputsSHA256, d.Hashes) {
		t.Fatalf("the view shows the hashes %v, not the dataset's %v", view.InputsSHA256, d.Hashes)
	}
	var got annotate.ItemView
	if code := call("POST", "/api/annotate/runs/"+id+"/verdict", map[string]any{"item": items[0], "choice": "gold", "inputs_sha256": view.InputsSHA256}, &got); code != http.StatusOK || got.Verdict == nil || got.Verdict.Choice != "gold" {
		t.Fatalf("a verdict on the page as it is: %d %+v", code, got.Verdict)
	}
	stale := map[string]string{}
	for k, v := range view.InputsSHA256 {
		stale[k] = v
	}
	stale["gold.json"] = "0000000000000000000000000000000000000000000000000000000000000000"
	if code := call("POST", "/api/annotate/runs/"+id+"/verdict", map[string]any{"item": items[0], "choice": "none", "note": "on an old page", "inputs_sha256": stale}, nil); code != http.StatusConflict {
		t.Fatalf("a verdict on a stale page: %d, want %d", code, http.StatusConflict)
	}
	fewer := map[string]string{}
	for k, v := range view.InputsSHA256 {
		if k != "gold.json" {
			fewer[k] = v
		}
	}
	if code := call("POST", "/api/annotate/runs/"+id+"/verdict", map[string]any{"item": items[1], "choice": "none", "inputs_sha256": fewer}, nil); code != http.StatusConflict {
		t.Fatalf("a verdict on a page with an input fewer: %d, want %d", code, http.StatusConflict)
	}
	got = annotate.ItemView{}
	call("GET", "/api/annotate/runs/"+id+"/item/"+items[0], nil, &got)
	if got.Verdict == nil || got.Verdict.Choice != "gold" || got.Verdict.Note != "" {
		t.Fatalf("a refused verdict changed the claim's: %+v", got.Verdict)
	}
	b, err := os.ReadFile(filepath.Join(made.Status.Dir, "review", "human.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"choice": "gold"`) || strings.Contains(string(b), "on an old page") {
		t.Fatalf("human.json after a refused verdict: %s", b)
	}
	got = annotate.ItemView{}
	if code := call("POST", "/api/annotate/runs/"+id+"/verdict", map[string]any{"item": items[1], "choice": "unsure"}, &got); code != http.StatusOK || got.Verdict == nil || !reflect.DeepEqual(got.Verdict.InputsSHA256, d.Hashes) {
		t.Fatalf("a verdict sent without the page's hashes: %d %+v", code, got.Verdict)
	}
}
