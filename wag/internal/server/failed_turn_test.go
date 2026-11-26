package server

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/yourorg/wag/internal/fsutil"
	"github.com/yourorg/wag/internal/llm"
)

// Follow-ups to the independent review of this round (R1, R8): a model turn
// that fails after a paid reply is saved before the model can be asked again,
// or the game is unsaved and its model waits; Save now works on a game open
// read-only.

// failingModelGame starts, on a server with a data folder, a game of Ann
// against M, a model whose first reply carries no move and whose later calls
// fail (the provider refuses the key); Ann has passed, and the game is saved.
func failingModelGame(t *testing.T) (modelGameFixture, *atomic.Int32) {
	t.Helper()
	srv, ts := newTestServerWith(t)
	store := t.TempDir()
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	calls := &atomic.Int32{}
	srv.AddModel(llm.WithProvider(llm.Profile{Name: "m", Provider: "mock", Retries: -1, Params: llm.Params{Model: "m-1"}}, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		if calls.Add(1) == 1 {
			return &llm.Response{Text: "PAID-REPLY-ONE, with no move in it"}, nil
		}
		return nil, &llm.Error{Provider: "mock", Kind: llm.KindAuth, Message: "the provider refused the key"}
	}}))
	var st State
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Seats: []Seat{{Name: "Ann", Kind: "human"}, {Name: "M", Kind: "model", Model: "m"}}}, &st); code != http.StatusOK {
		t.Fatalf("new game: %d", code)
	}
	if mr := post(t, ts, "/api/games/"+st.ID+"/move", map[string]any{"action": map[string]string{"type": "abstain"}}); mr.State.SavedTo == "" || mr.State.Unsaved {
		t.Fatalf("Ann's move is not saved: %+v", mr.State)
	}
	return modelGameFixture{srv: srv, ts: ts, store: store, id: st.ID}, calls
}

// The model's paid reply is on the disk after its turn fails, and the state
// pollers saw thinking has changed.
func TestAFailedModelTurnIsSaved(t *testing.T) {
	f, calls := failingModelGame(t)
	var before State
	call(t, f.ts, "GET", "/api/games/"+f.id, nil, &before)
	var e map[string]string
	if code := call(t, f.ts, "POST", "/api/games/"+f.id+"/ai", map[string]any{}, &e); code != http.StatusConflict || calls.Load() != 2 {
		t.Fatalf("the model's turn: %d %v, %d calls", code, e, calls.Load())
	}
	var now State
	call(t, f.ts, "GET", "/api/games/"+f.id, nil, &now)
	b, err := os.ReadFile(filepath.Join(now.SavedTo, "trace.jsonl"))
	if err != nil || !strings.Contains(string(b), "PAID-REPLY-ONE") || now.Unsaved {
		t.Fatalf("the paid reply is not on the disk (%v), unsaved %v", err, now.Unsaved)
	}
	var un map[string]any
	call(t, f.ts, "GET", "/api/games/"+f.id+"?since="+itoa(before.Version+1), nil, &un)
	if un["unchanged"] == true || un["thinking"] == true {
		t.Fatalf("a poller that saw the seat thinking sees no change: %v", un["unchanged"])
	}
}

// When that save fails, the game says so, and the model is asked nothing
// until the reply is on the disk.
func TestAFailedModelTurnThatCannotBeSavedHoldsTheModel(t *testing.T) {
	f, calls := failingModelGame(t)
	var st State
	call(t, f.ts, "GET", "/api/games/"+f.id, nil, &st)
	blocked := filepath.Join(st.SavedTo, "trace.jsonl.tmp")
	if err := os.Mkdir(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
	call(t, f.ts, "POST", "/api/games/"+f.id+"/ai", map[string]any{}, nil)
	call(t, f.ts, "GET", "/api/games/"+f.id, nil, &st)
	if !st.Unsaved || st.SaveError == "" || calls.Load() != 2 {
		t.Fatalf("after the failed turn: unsaved %v %q, %d calls", st.Unsaved, st.SaveError, calls.Load())
	}
	var e map[string]string
	if code := call(t, f.ts, "POST", "/api/games/"+f.id+"/ai", map[string]any{}, &e); code != http.StatusConflict || !strings.Contains(e["error"], "model seats wait") || calls.Load() != 2 {
		t.Fatalf("the model was asked again while its reply is not on the disk: %d %v, %d calls", code, e, calls.Load())
	}
	if err := os.Remove(blocked); err != nil {
		t.Fatal(err)
	}
	if code := call(t, f.ts, "POST", "/api/games/"+f.id+"/save", map[string]any{}, nil); code != http.StatusOK {
		t.Fatalf("save now: %d", code)
	}
	if b, _ := os.ReadFile(filepath.Join(st.SavedTo, "trace.jsonl")); !strings.Contains(string(b), "PAID-REPLY-ONE") {
		t.Fatal("the paid reply is not on the disk after Save now")
	}
}

// A reply pasted for a seat played through the text protocol is in the
// record, played or not, and so it is saved.
func TestAReplyThatDoesNotParseIsSaved(t *testing.T) {
	f := modelGame(t)
	var e map[string]string
	if code := call(t, f.ts, "POST", "/api/games/"+f.id+"/reply", map[string]any{"reply": "PASTED-REPLY with no move"}, &e); code != http.StatusBadRequest {
		t.Fatalf("reply: %d %v", code, e)
	}
	var st State
	call(t, f.ts, "GET", "/api/games/"+f.id, nil, &st)
	b, err := os.ReadFile(filepath.Join(st.SavedTo, "trace.jsonl"))
	if st.SavedTo == "" || err != nil || !strings.Contains(string(b), "PASTED-REPLY") {
		t.Fatalf("the reply is not saved: %q %v", st.SavedTo, err)
	}
}

// A game open read-only takes notes, so its save can fail: Save now works on
// it (the review's R8).
func TestSaveNowWorksOnAGameOpenReadOnly(t *testing.T) {
	f := modelGame(t)
	post(t, f.ts, "/api/games/"+f.id+"/move", map[string]any{"action": map[string]string{"type": "abstain"}})
	var e Export
	if code := call(t, f.ts, "GET", "/api/games/"+f.id+"/export", nil, &e); code != http.StatusOK {
		t.Fatalf("export: %d", code)
	}
	e.DataHash = "0123456789abcdef"
	code, st := loadHTTP(t, f.ts, &e, "?inspect=1")
	if code != http.StatusOK || st["inspect"] == nil {
		t.Fatalf("inspect: %d %v", code, st["error"])
	}
	id := st["id"].(string)
	unblock := blockGames(t, f.store)
	var nr struct{ State State }
	if code := call(t, f.ts, "POST", "/api/games/"+id+"/note", map[string]string{"target": "game", "kind": "idea", "text": "x", "by": "me"}, &nr); code != http.StatusOK || nr.State.SaveError == "" {
		t.Fatalf("note: %d, save error %q", code, nr.State.SaveError)
	}
	unblock()
	var out struct{ State State }
	if code := call(t, f.ts, "POST", "/api/games/"+id+"/save", map[string]any{}, &out); code != http.StatusOK || out.State.Unsaved {
		t.Fatalf("Save now on a game open read-only: %d %+v", code, out.State)
	}
	// It still takes no move.
	var me map[string]string
	if code := call(t, f.ts, "POST", "/api/games/"+id+"/move", map[string]any{"action": map[string]string{"type": "abstain"}}, &me); code != http.StatusConflict {
		t.Fatalf("a move on a game open read-only: %d %v", code, me)
	}
}

// A save whose folder could not be synced is not done until a later save has
// synced it (the review's R2): Save now syncs games/ again.
func TestSaveNowSyncsTheFolderAFailedSaveCouldNot(t *testing.T) {
	f := modelGame(t)
	games := filepath.Join(f.store, "games")
	old := fsutil.SyncDir
	fail := true
	var synced []string
	fsutil.SyncDir = func(d string) error {
		synced = append(synced, d)
		if fail && d == games {
			return errors.New("I/O error")
		}
		return old(d)
	}
	t.Cleanup(func() { fsutil.SyncDir = old })
	mr := post(t, f.ts, "/api/games/"+f.id+"/move", map[string]any{"action": map[string]string{"type": "abstain"}})
	if !mr.State.Unsaved || !strings.Contains(mr.State.SaveError, "could not be synced") {
		t.Fatalf("the first save: %+v", mr.State)
	}
	var e map[string]string
	if code := call(t, f.ts, "POST", "/api/games/"+f.id+"/save", map[string]any{}, &e); code != http.StatusConflict {
		t.Fatalf("Save now while games/ cannot be synced: %d %v", code, e)
	}
	fail, synced = false, nil
	var saved struct{ State State }
	if code := call(t, f.ts, "POST", "/api/games/"+f.id+"/save", map[string]any{}, &saved); code != http.StatusOK || saved.State.Unsaved {
		t.Fatalf("Save now: %d %+v", code, saved.State)
	}
	found := false
	for _, d := range synced {
		found = found || d == games
	}
	if !found {
		t.Fatalf("Save now reported the game saved without syncing %s: synced %v", games, synced)
	}
}

// A request that fails changes no state, so pollers see no change; but if
// the save it makes fails, the state says so, and pollers see that.
func TestAFailedSaveOnAFailedRequestReachesPollers(t *testing.T) {
	f := modelGame(t)
	var st State
	call(t, f.ts, "GET", "/api/games/"+f.id, nil, &st)
	call(t, f.ts, "POST", "/api/games/"+f.id+"/reply", map[string]any{"reply": "no move here"}, nil)
	var un map[string]any
	call(t, f.ts, "GET", "/api/games/"+f.id+"?since="+itoa(st.Version), nil, &un)
	if un["unchanged"] != true {
		t.Fatalf("a refused reply that was saved changed the state: %v", un["unchanged"])
	}
	unblock := blockGames(t, f.store)
	defer unblock()
	call(t, f.ts, "POST", "/api/games/"+f.id+"/reply", map[string]any{"reply": "no move here either"}, nil)
	var now map[string]any
	call(t, f.ts, "GET", "/api/games/"+f.id+"?since="+itoa(st.Version), nil, &now)
	if now["unchanged"] == true || now["unsaved"] != true {
		t.Fatalf("pollers do not see the failed save: %v %v", now["unchanged"], now["unsaved"])
	}
}

// A pasted reply whose move cannot be played (a move awaits the moderator)
// is in the record, and so it is saved.
func TestAReplyThatCannotBePlayedIsSaved(t *testing.T) {
	srv, ts := newTestServerWith(t)
	if err := srv.SetStore(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	srv.AddModel(llm.WithProvider(llm.Profile{Name: "m", Provider: "mock", Params: llm.Params{Model: "m-1"}}, &llm.Mock{}))
	var st State
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "cq", Moderated: true, Seats: []Seat{{Name: "Ann", Kind: "human"}, {Name: "M", Kind: "model", Model: "m"}}}, &st); code != http.StatusOK {
		t.Fatalf("new game: %d", code)
	}
	pr := map[string]any{"type": "support", "scheme_id": "AS:practical_reasoning", "text": "Tax carbon to cut emissions.",
		"variable_bindings": map[string]string{"A": "carbon tax", "S1": "status quo", "S2": "lower emissions", "G": "clean air", "V": "health"},
		"declare":           []map[string]string{{"id": "carbon tax", "actor": "decider"}}}
	if mr := post(t, ts, "/api/games/"+st.ID+"/move", map[string]any{"action": pr}); mr.Result["pending"] != true || mr.State.SavedTo == "" {
		t.Fatalf("the move is not held, or not saved: %v %q", mr.Result, mr.State.SavedTo)
	}
	var e map[string]string
	if code := call(t, ts, "POST", "/api/games/"+st.ID+"/reply", map[string]any{"reply": `REPLY-WHILE-PENDING {"action": {"type": "abstain"}}`}, &e); code != http.StatusConflict {
		t.Fatalf("reply while a move is held: %d %v", code, e)
	}
	var now State
	call(t, ts, "GET", "/api/games/"+st.ID, nil, &now)
	if b, err := os.ReadFile(filepath.Join(now.SavedTo, "trace.jsonl")); err != nil || !strings.Contains(string(b), "REPLY-WHILE-PENDING") {
		t.Fatalf("the refused reply is not saved: %v", err)
	}
}
