package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/yourorg/wag/internal/llm"
)

// A game with a model seat is saved after every change (autosave). A save
// that fails is shown, not only printed, the folder is named only once a save
// has succeeded, and the model is asked nothing until the game is saved.

type modelGameFixture struct {
	srv   *Server
	ts    *httptest.Server
	store string
	id    string
	calls *atomic.Int32
}

// modelGame starts, on a server with a data folder, a game of Ann (a person)
// against M, a model that passes.
func modelGame(t *testing.T) modelGameFixture {
	t.Helper()
	srv, ts := newTestServerWith(t)
	store := t.TempDir()
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	calls := &atomic.Int32{}
	srv.AddModel(llm.WithProvider(llm.Profile{Name: "m", Provider: "mock", Params: llm.Params{Model: "m-1"}}, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		calls.Add(1)
		return &llm.Response{Text: `{"action": {"type": "abstain"}}`}, nil
	}}))
	var st State
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Seats: []Seat{{Name: "Ann", Kind: "human"}, {Name: "M", Kind: "model", Model: "m"}}}, &st); code != http.StatusOK {
		t.Fatalf("new game: %d", code)
	}
	return modelGameFixture{srv, ts, store, st.ID, calls}
}

// blockGames makes the data folder's games/ a file, so that no game can be
// saved; the returned function puts the folder back.
func blockGames(t *testing.T, store string) func() {
	t.Helper()
	games := filepath.Join(store, "games")
	if err := os.RemoveAll(games); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(games, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.Remove(games); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(games, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAGameThatCannotBeSavedSaysSoAndItsModelWaits(t *testing.T) {
	f := modelGame(t)
	unblock := blockGames(t, f.store)
	mr := post(t, f.ts, "/api/games/"+f.id+"/move", map[string]any{"action": map[string]string{"type": "abstain"}})
	if mr.State.SaveError == "" || !mr.State.Unsaved || mr.State.SavedTo != "" || mr.State.Active != "M" {
		t.Fatalf("after Ann's move: %q unsaved %v saved to %q, %s to move", mr.State.SaveError, mr.State.Unsaved, mr.State.SavedTo, mr.State.Active)
	}
	var e map[string]string
	if code := call(t, f.ts, "POST", "/api/games/"+f.id+"/ai", map[string]any{}, &e); code != http.StatusConflict ||
		!strings.Contains(e["error"], "model seats wait") || !strings.Contains(e["error"], "Save now") || f.calls.Load() != 0 {
		t.Fatalf("the model's turn: %d %v, %d calls", code, e, f.calls.Load())
	}
	// Another computer sees that the game is not saved, not where or why.
	keys := f.srv.sessions[f.id].keysView()
	code, body := remote(t, f.srv, "GET", "/api/games/"+f.id, keys.Seats[0].Key, nil)
	var st State
	if code != http.StatusOK || json.Unmarshal(body, &st) != nil || !st.Unsaved || st.SaveError != "" || st.SavedTo != "" || strings.Contains(string(body), f.store) {
		t.Fatalf("remote: %d %s", code, body)
	}
	if code, body := remote(t, f.srv, "POST", "/api/games/"+f.id+"/ai", keys.Seats[0].Key, map[string]any{}); code != http.StatusConflict ||
		!strings.Contains(string(body), "model seats wait") || strings.Contains(string(body), f.store) || f.calls.Load() != 0 {
		t.Fatalf("remote ai: %d %s", code, body)
	}
	// Save now: refused while the folder is blocked, then done.
	if code := call(t, f.ts, "POST", "/api/games/"+f.id+"/save", map[string]any{}, &e); code != http.StatusConflict || !strings.Contains(e["error"], "could not be saved") {
		t.Fatalf("save while blocked: %d %v", code, e)
	}
	unblock()
	var saved struct{ State State }
	if code := call(t, f.ts, "POST", "/api/games/"+f.id+"/save", map[string]any{}, &saved); code != http.StatusOK ||
		saved.State.SaveError != "" || saved.State.Unsaved || saved.State.SavedTo == "" {
		t.Fatalf("save: %d %+v", code, saved.State)
	}
	for _, name := range []string{"trace.jsonl", "export.json"} {
		b, err := os.ReadFile(filepath.Join(saved.State.SavedTo, name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), f.store) {
			t.Fatalf("%s names the data folder", name)
		}
	}
	if mr := post(t, f.ts, "/api/games/"+f.id+"/ai", map[string]any{}); f.calls.Load() != 1 || mr.State.Unsaved {
		t.Fatalf("the model's turn after the save: %d calls, %+v", f.calls.Load(), mr.State)
	}
	// Saved again, by a game that knows its folder: neither the files nor
	// another computer's export name it.
	for _, name := range []string{"trace.jsonl", "export.json"} {
		if b, _ := os.ReadFile(filepath.Join(saved.State.SavedTo, name)); strings.Contains(string(b), f.store) {
			t.Fatalf("%s names the data folder", name)
		}
	}
	if code, body := remote(t, f.srv, "GET", "/api/games/"+f.id+"/export", keys.Seats[0].Key, nil); code != http.StatusOK || strings.Contains(string(body), f.store) {
		t.Fatalf("remote export: %d, names the folder: %v", code, strings.Contains(string(body), f.store))
	}
	// Save now is the owner's.
	if code, _ := remote(t, f.srv, "POST", "/api/games/"+f.id+"/save", keys.Seats[0].Key, map[string]any{}); code != http.StatusForbidden {
		t.Fatalf("remote save: %d", code)
	}
}

// A save that fails after one that succeeded is shown with the folder of the
// last good save; the model's turn tries to save again first.
func TestAFailedSaveAfterAGoodOneIsShown(t *testing.T) {
	f := modelGame(t)
	// A move through the text protocol is saved too.
	mr := post(t, f.ts, "/api/games/"+f.id+"/reply", map[string]any{"reply": `{"action": {"type": "abstain"}}`})
	dir := mr.State.SavedTo
	if dir == "" || mr.State.Unsaved {
		t.Fatalf("first save: %+v", mr.State)
	}
	blocked := filepath.Join(dir, "export.json.tmp")
	if err := os.Mkdir(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
	var e map[string]string
	if code := call(t, f.ts, "POST", "/api/games/"+f.id+"/note", map[string]string{"target": "game", "kind": "idea", "text": "x", "by": "me"}, nil); code != http.StatusOK {
		t.Fatalf("note: %d", code)
	}
	var st State
	call(t, f.ts, "GET", "/api/games/"+f.id, nil, &st)
	if st.SavedTo != dir || st.SaveError == "" || !st.Unsaved {
		t.Fatalf("after a failed save: %q %q %v", st.SavedTo, st.SaveError, st.Unsaved)
	}
	if code := call(t, f.ts, "POST", "/api/games/"+f.id+"/ai", map[string]any{}, &e); code != http.StatusConflict || f.calls.Load() != 0 {
		t.Fatalf("ai while unsaved: %d %v", code, e)
	}
	if err := os.Remove(blocked); err != nil {
		t.Fatal(err)
	}
	if mr := post(t, f.ts, "/api/games/"+f.id+"/ai", map[string]any{}); f.calls.Load() != 1 || mr.State.Unsaved || mr.State.SaveError != "" {
		t.Fatalf("ai after the folder is fixed: %d calls, %+v", f.calls.Load(), mr.State)
	}
	b, err := os.ReadFile(filepath.Join(dir, "export.json"))
	if err != nil || !strings.Contains(string(b), `"x"`) {
		t.Fatalf("the export lacks the note: %v", err)
	}
}

// A game without a model seat is not saved, and Save now has nothing to do.
func TestAGameWithoutAModelIsNotSaved(t *testing.T) {
	srv, ts := newTestServerWith(t)
	if err := srv.SetStore(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	var st State
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Seats: twoHumans}, &st)
	mr := post(t, ts, "/api/games/"+st.ID+"/move", map[string]any{"action": map[string]string{"type": "abstain"}})
	if mr.State.SavedTo != "" || mr.State.Unsaved {
		t.Fatalf("%+v", mr.State)
	}
	if code := call(t, ts, "POST", "/api/games/"+st.ID+"/save", map[string]any{}, nil); code != http.StatusOK {
		t.Fatalf("save: %d", code)
	}
}
