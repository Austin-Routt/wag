package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/wag/internal/annotate"
	"github.com/yourorg/wag/internal/llm"
)

// do sends a request with the browser's headers and decodes the answer.
func do(t *testing.T, ts *httptest.Server, method, path string, body any, out any, hdr ...string) int {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, ts.URL+path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if out != nil {
		if err := json.Unmarshal(b, out); err != nil {
			t.Fatalf("%s %s: %s", method, path, b)
		}
	}
	return resp.StatusCode
}

// Keys typed in the dashboard: kept in memory, or remembered in keys.json
// only when asked; never sent back; players saved without them; a player
// waits for its key and plays once it has one.
func TestWorkbenchKeysAndPlayers(t *testing.T) {
	srv, ts := newTestServerWith(t)
	store := t.TempDir()
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")
	var sv struct {
		Services []Service `json:"services"`
		Store    string    `json:"store"`
	}
	do(t, ts, "GET", "/api/llm/services", nil, &sv)
	if len(sv.Services) < 8 || sv.Store != store || sv.Services[0].ID != "gemini" || sv.Services[0].KeySource != "" {
		t.Fatalf("services %+v", sv)
	}
	// A Gemini player without a key is kept, and waits.
	var pl struct {
		Players []PlayerView `json:"players"`
		Warning string       `json:"warning"`
	}
	if code := do(t, ts, "POST", "/api/llm/players", map[string]any{"service": "gemini", "model": "gemini-x", "name": "flash"}, &pl); code != 200 ||
		pl.Warning == "" || pl.Players[0].Ready {
		t.Fatalf("player without a key: %d %+v", code, pl)
	}
	// The key, remembered: the player can play; the key is in keys.json
	// only, and never in an answer.
	var after struct {
		Services []Service    `json:"services"`
		Players  []PlayerView `json:"players"`
	}
	raw := "TESTKEY-not-a-real-key-wxyz"
	do(t, ts, "POST", "/api/llm/key", map[string]any{"service": "gemini", "key": raw, "remember": true}, &after)
	if after.Services[0].KeySource != "remembered" || after.Services[0].KeyHint != "…wxyz" || !after.Players[0].Ready {
		t.Fatalf("after the key: %+v %+v", after.Services[0], after.Players)
	}
	keys, _ := os.ReadFile(filepath.Join(store, "keys.json"))
	players, _ := os.ReadFile(filepath.Join(store, "players.json"))
	if !strings.Contains(string(keys), raw) || strings.Contains(string(players), raw) {
		t.Fatalf("keys.json %s / players.json %s", keys, players)
	}
	if st, _ := os.Stat(filepath.Join(store, "keys.json")); st.Mode().Perm()&0o077 != 0 && os.PathSeparator == '/' {
		t.Fatalf("keys.json is readable by others: %v", st.Mode())
	}
	for _, p := range []string{"/api/llm/services", "/api/llm/players", "/api/meta"} {
		resp, _ := http.Get(ts.URL + p)
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if strings.Contains(string(b), raw) {
			t.Fatalf("%s sends the key back", p)
		}
	}
	// A new server on the same folder has the player and the key.
	srv2, err := New(srv.data, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := srv2.SetStore(store); err != nil {
		t.Fatal(err)
	}
	if srv2.models["flash"] == nil {
		t.Fatal("the remembered key and player did not come back")
	}
	// Forgotten: keys.json goes, the player waits again.
	do(t, ts, "POST", "/api/llm/forget", map[string]any{"service": "gemini"}, &after)
	if _, err := os.Stat(filepath.Join(store, "keys.json")); !os.IsNotExist(err) || after.Players[0].Ready {
		t.Fatalf("forget: %v %+v", err, after.Players)
	}
	// A key not remembered stays in memory.
	do(t, ts, "POST", "/api/llm/key", map[string]any{"service": "gemini", "key": raw}, &after)
	if _, err := os.Stat(filepath.Join(store, "keys.json")); !os.IsNotExist(err) || after.Services[0].KeySource != "dashboard" {
		t.Fatalf("a key kept in memory: %v %+v", err, after.Services[0])
	}
	// The test service lists its model, answers a test, and plays.
	var ms struct {
		Models []llm.ModelInfo `json:"models"`
	}
	do(t, ts, "POST", "/api/llm/models", map[string]any{"service": "mock"}, &ms)
	var test map[string]any
	do(t, ts, "POST", "/api/llm/test", map[string]any{"service": "mock", "model": "mock-1"}, &test)
	if len(ms.Models) != 1 || test["error"] != nil || test["text"] == "" {
		t.Fatalf("mock: %+v %+v", ms, test)
	}
	do(t, ts, "POST", "/api/llm/players", map[string]any{"service": "mock", "model": "mock-1", "name": "scripted"}, &pl)
	var st State
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Seats: []Seat{{Name: "P", Kind: "model", Model: "scripted"}, {Name: "C", Kind: "greedy"}}}, &st); code != 200 {
		t.Fatalf("a game with a dashboard player: %d", code)
	}
	if mr := post(t, ts, "/api/games/"+st.ID+"/ai", map[string]any{}); mr.State.SavedTo == "" {
		t.Fatal("a game with a model seat is not saved")
	} else if _, err := os.Stat(filepath.Join(mr.State.SavedTo, "trace.jsonl")); err != nil {
		t.Fatal(err)
	}
	do(t, ts, "POST", "/api/llm/players/remove", map[string]any{"name": "scripted"}, &pl)
	for _, p := range pl.Players {
		if p.Name == "scripted" {
			t.Fatal("not removed")
		}
	}
	var stored map[string]any
	do(t, ts, "GET", "/api/store", nil, &stored)
	if stored["dir"] != store || len(stored["games"].([]any)) != 1 {
		t.Fatalf("store %+v", stored)
	}
	if code := do(t, ts, "GET", "/api/store/file?path=../outside.txt", nil, nil); code != http.StatusBadRequest {
		t.Fatalf("a file outside the data folder: %d", code)
	}
}

// The workbench answers only requests from this computer, at localhost,
// with JSON from its own page.
func TestWorkbenchGuards(t *testing.T) {
	_, ts := newTestServerWith(t)
	body := map[string]any{"service": "gemini", "key": "x1234567890"}
	if code := do(t, ts, "POST", "/api/llm/key", body, nil, "Origin", "https://evil.example"); code != http.StatusForbidden {
		t.Fatalf("another site's request: %d", code)
	}
	req, _ := http.NewRequest("POST", ts.URL+"/api/llm/key", strings.NewReader(`{"service":"gemini","key":"x1234567890"}`))
	req.Header.Set("Content-Type", "text/plain")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("a form post: %d", resp.StatusCode)
	}
	req, _ = http.NewRequest("GET", ts.URL+"/api/llm/services", nil)
	req.Host = "attacker.example:8080"
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("another host name: %d", resp.StatusCode)
	}
	if code := do(t, ts, "POST", "/api/llm/key", body, nil, "Origin", ts.URL); code != 200 {
		t.Fatalf("the page's own request: %d", code)
	}
}

var claimRe = regexp.MustCompile(`"kialo_claim_id": "([^"]+)"`)

// An annotation run through the API: two annotators and a reviewer on the
// built-in dataset's first claims, step by step and then the rest, the
// report, a claim's view, a verdict; all in the run's folder.
func TestWorkbenchAnnotation(t *testing.T) {
	srv, ts := newTestServerWith(t)
	store := t.TempDir()
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	ds, _ := annotate.Builtin()
	gold := ds[0].Gold
	answer := func(changeScheme bool) *llm.Mock {
		return &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
			m := claimRe.FindAllStringSubmatch(req.Messages[len(req.Messages)-1].Content, -1)
			id := m[len(m)-1][1]
			b, _ := json.Marshal(gold[id])
			s := string(b)
			if changeScheme { // another primary scheme
				var a map[string]any
				_ = json.Unmarshal(b, &a)
				for _, x := range a["instances"].([]any) {
					if in := x.(map[string]any); in["primary"] == true {
						in["scheme"].(map[string]any)["walton_id"] = "sign"
					}
				}
				c, _ := json.Marshal(a)
				s = string(c)
			}
			if strings.Contains(req.System, "You review") {
				s = `{"fields": {"scheme": {"choice": "A"}}, "adjudicated": ` + string(b) + `, "confidence": 0.9}`
			}
			return &llm.Response{Text: s}, nil
		}}
	}
	for name, m := range map[string]*llm.Mock{"one": answer(false), "two": answer(true), "judge": answer(false)} {
		srv.AddModel(llm.WithProvider(llm.Profile{Name: name, Provider: "mock", Params: llm.Params{Model: "m-" + name}}, m))
	}
	var dl struct {
		Datasets []annotate.Dataset `json:"datasets"`
	}
	do(t, ts, "GET", "/api/annotate/datasets", nil, &dl)
	if len(dl.Datasets) != 1 || dl.Datasets[0].NItems != 50 {
		t.Fatalf("datasets %+v", dl)
	}
	var made struct {
		Status annotate.Status `json:"status"`
	}
	if code := do(t, ts, "POST", "/api/annotate/runs", map[string]any{"dataset": dl.Datasets[0].ID, "annotators": []string{"one", "two"},
		"reviewer": "judge", "count": 3}, &made); code != 200 || len(made.Status.Settings.Items) != 3 {
		t.Fatalf("new run: %d %+v", code, made)
	}
	id := made.Status.ID
	wait := func() annotate.Status {
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
	do(t, ts, "POST", "/api/annotate/runs/"+id+"/next", map[string]any{}, nil)
	if st := wait(); st.Progress[0].Done != 1 || st.NextItem != made.Status.Settings.Items[1] {
		t.Fatalf("one step: %+v", st)
	}
	do(t, ts, "POST", "/api/annotate/runs/"+id+"/start", map[string]any{}, nil)
	if st := wait(); st.Progress[0].Done != 3 || st.Progress[1].Done != 3 || st.Disagreed != 3 {
		t.Fatalf("all: %+v", st)
	}
	do(t, ts, "POST", "/api/annotate/runs/"+id+"/review", map[string]any{}, nil)
	if st := wait(); st.Reviewed != 3 {
		t.Fatalf("review: %+v", st)
	}
	item := made.Status.Settings.Items[0]
	var v annotate.ItemView
	do(t, ts, "POST", "/api/annotate/runs/"+id+"/verdict", map[string]any{"item": item, "choice": "one", "note": "right"}, &v)
	if v.Verdict == nil || v.Review == nil || len(v.Annotations) != 2 || len(v.Problems["two"]) == 0 || len(v.Problems["one"]) != 0 {
		t.Fatalf("item %+v", v)
	}
	var rep annotate.Report
	do(t, ts, "GET", "/api/annotate/runs/"+id+"/report", nil, &rep)
	if rep.Annotators[0].VsGold.Fields["scheme"].Agreement != 1 || rep.Verdicts["one"] != 1 || rep.Reviewer.Reviewed != 3 {
		t.Fatalf("report %+v", rep)
	}
	var list struct {
		Runs []map[string]any `json:"runs"`
	}
	do(t, ts, "GET", "/api/annotate/runs", nil, &list)
	if len(list.Runs) != 1 {
		t.Fatalf("runs %+v", list)
	}
	for _, f := range []string{"run.json", "README.txt", "calls.jsonl", "report.json", "annotations/one.json", "annotations/two.json",
		"review/model-judge.json", "review/human.json"} {
		if _, err := os.Stat(filepath.Join(store, "annotation", id, f)); err != nil {
			t.Errorf("no %s", f)
		}
	}
	var file []any
	if code := do(t, ts, "GET", "/api/store/file?path=annotation/"+id+"/review/human.json", nil, &file); code != 200 || len(file) != 1 {
		t.Fatalf("download: %d %v", code, file)
	}
	if code := do(t, ts, "GET", "/api/annotate/runs/..%2Fx", nil, nil); code != http.StatusNotFound {
		t.Fatalf("a run outside: %d", code)
	}
}

// A Kialo discussion becomes a dataset folder, which the dashboard offers.
func TestWorkbenchKialoDataset(t *testing.T) {
	srv, ts := newTestServerWith(t)
	store := t.TempDir()
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	graph := json.RawMessage(`{"id": "77", "displayTitle": "Should we?", "discussionUrl": "https://example.org/77", "thesis_claim_id": "77.1",
		"claims": {"77.1": {"id": "77.1", "text": "We should."}, "77.2": {"id": "77.2", "text": "It helps."}, "77.3": {"id": "77.3", "text": "It costs."}},
		"edges": [{"location_id": "77.11", "from_claim_id": "77.1", "to_claim_id": "77.2", "relation": "support", "active": true},
			{"location_id": "77.12", "from_claim_id": "77.1", "to_claim_id": "77.3", "relation": "attack", "active": true}]}`)
	var made map[string]any
	if code := do(t, ts, "POST", "/api/annotate/datasets/kialo", map[string]any{"graph": graph}, &made); code != 200 || made["dataset"] != "folder:kialo-77" {
		t.Fatalf("%d %v", code, made)
	}
	var dl struct {
		Datasets []annotate.Dataset `json:"datasets"`
	}
	do(t, ts, "GET", "/api/annotate/datasets", nil, &dl)
	if len(dl.Datasets) != 2 || dl.Datasets[1].ID != "folder:kialo-77" || dl.Datasets[1].NItems != 3 || dl.Datasets[1].HasGold {
		t.Fatalf("datasets %+v", dl.Datasets)
	}
	if _, err := os.Stat(filepath.Join(store, "datasets", "kialo-77", "items.json")); err != nil {
		t.Fatal(err)
	}
	if code := do(t, ts, "POST", "/api/annotate/datasets/kialo", map[string]any{"kialo_id": "../x"}, nil); code != http.StatusBadRequest {
		t.Fatalf("a discussion outside: %d", code)
	}
}
