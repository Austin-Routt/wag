package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/game"
)

// Who may do what in a game (access.go). The critical review of 2026-11-23
// (repro/remote_moderator_probe_test.go) showed another computer, with no
// credential, deciding a statement as the moderator; its probe passed by
// reproducing that. These tests reverse it, and check every other role.

// remote sends a request as another computer (198.51.100.23, which reaches
// this one as wag.example), with the game key key if not empty.
func remote(t *testing.T, s *Server, method, path, key string, body any) (int, []byte) {
	t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	r := httptest.NewRequest(method, "http://wag.example:8080"+path, rd)
	r.RemoteAddr = "198.51.100.23:1234"
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	if key != "" {
		r.Header.Set(KeyHeader, key)
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w.Code, w.Body.Bytes()
}

// moderatedGame starts, as the owner, a moderated game of two people with a
// statement p to decide on, and returns it with its keys.
func moderatedGame(t *testing.T, s *Server) (*Session, KeysView) {
	t.Helper()
	req := NewGameRequest{Mode: "classic", Moderated: true, Seats: []Seat{{Name: "A", Kind: "human"}, {Name: "B", Kind: "human"}}}
	cfg, err := req.config(s.contexts)
	if err != nil {
		t.Fatal(err)
	}
	gs, err := s.newState(cfg, []string{"A", "B"})
	if err != nil {
		t.Fatal(err)
	}
	gs.Adapter.AddStatement("p", "A synthetic claim")
	s.mu.Lock()
	ss := s.add(gs, req.Seats, "access", "new", 1)
	s.mu.Unlock()
	return ss, ss.keysView()
}

func TestAnotherComputerDecidesOnlyWithTheModeratorsKey(t *testing.T) {
	s, _ := newTestServerWith(t)
	ss, keys := moderatedGame(t, s)
	decide := map[string]string{"statement": "p", "decision": "accepted"}
	for name, key := range map[string]string{"no key": "", "a seat's key": keys.Seats[0].Key, "a wrong key": "x" + keys.Moderator} {
		code, body := remote(t, s, "POST", "/api/games/"+ss.ID+"/decide", key, decide)
		if code != http.StatusForbidden || len(ss.GS.MoveHistory) != 0 {
			t.Fatalf("%s: %d %s, %d moves", name, code, body, len(ss.GS.MoveHistory))
		}
		code, _ = remote(t, s, "POST", "/api/games/"+ss.ID+"/moderate", key, map[string]any{"approve": true})
		if code != http.StatusForbidden {
			t.Fatalf("%s: moderate %d", name, code)
		}
	}
	code, body := remote(t, s, "POST", "/api/games/"+ss.ID+"/decide", keys.Moderator, decide)
	if code != http.StatusOK || len(ss.GS.MoveHistory) != 1 || ss.GS.MoveHistory[0].PlayerID != game.ModeratorID {
		t.Fatalf("the moderator's key: %d %s", code, body)
	}
}

func TestAnotherComputerMovesOnlyForItsSeatOnItsTurn(t *testing.T) {
	s, _ := newTestServerWith(t)
	ss, keys := moderatedGame(t, s)
	ss.GS.Config.Moderated = false // moves take effect at once
	abstain := map[string]any{"action": map[string]string{"type": "abstain"}}
	a, b := keys.Seats[0], keys.Seats[1]
	if a.Name != "A" || b.Name != "B" || ss.GS.ActivePlayer().ID != "A" {
		t.Fatalf("fixture: %v, %s to move", keys.Seats, ss.GS.ActivePlayer().ID)
	}
	for name, key := range map[string]string{"B's key on A's turn": b.Key, "the moderator's key": keys.Moderator} {
		for _, action := range []string{"move", "reply"} {
			body := any(abstain)
			if action == "reply" {
				body = map[string]string{"reply": `{"action": {"type": "abstain"}}`}
			}
			code, out := remote(t, s, "POST", "/api/games/"+ss.ID+"/"+action, key, body)
			if code != http.StatusForbidden || len(ss.GS.MoveHistory) != 0 {
				t.Fatalf("%s, %s: %d %s", name, action, code, out)
			}
		}
	}
	if code, out := remote(t, s, "POST", "/api/games/"+ss.ID+"/move", a.Key, abstain); code != http.StatusOK || ss.GS.ActivePlayer().ID != "B" {
		t.Fatalf("A's key on A's turn: %d %s", code, out)
	}
	if code, _ := remote(t, s, "POST", "/api/games/"+ss.ID+"/move", a.Key, abstain); code != http.StatusForbidden {
		t.Fatalf("A's key on B's turn: %d", code)
	}
	if code, out := remote(t, s, "POST", "/api/games/"+ss.ID+"/move", b.Key, abstain); code != http.StatusOK {
		t.Fatalf("B's key on B's turn: %d %s", code, out)
	}
}

// A key is for its game: it does nothing in another.
func TestAKeyIsForItsOwnGame(t *testing.T) {
	s, _ := newTestServerWith(t)
	one, keys := moderatedGame(t, s)
	two, _ := moderatedGame(t, s)
	for _, key := range []string{keys.Moderator, keys.Seats[0].Key} {
		if code, _ := remote(t, s, "GET", "/api/games/"+two.ID, key, nil); code != http.StatusForbidden {
			t.Fatalf("read: %d", code)
		}
		if code, _ := remote(t, s, "POST", "/api/games/"+two.ID+"/decide", key, map[string]string{"statement": "p", "decision": "accepted"}); code != http.StatusForbidden || len(two.GS.MoveHistory) != 0 {
			t.Fatalf("decide: %d", code)
		}
		if code, _ := remote(t, s, "GET", "/api/games/"+one.ID, key, nil); code != http.StatusOK {
			t.Fatalf("its own game: %d", code)
		}
	}
	// The list shows another computer the games it has a key for.
	var list []struct{ ID string }
	_, body := remote(t, s, "GET", "/api/games", keys.Seats[1].Key, nil)
	if json.Unmarshal(body, &list) != nil || len(list) != 1 || list[0].ID != one.ID {
		t.Fatalf("list: %s", body)
	}
}

// Without a key, another computer learns nothing: not whether a game exists,
// nor the list, and it starts, loads and imports no game.
func TestAnotherComputerWithoutAKeyLearnsNothing(t *testing.T) {
	s, _ := newTestServerWith(t)
	ss, keys := moderatedGame(t, s)
	_, exists := remote(t, s, "GET", "/api/games/"+ss.ID, "", nil)
	code, missing := remote(t, s, "GET", "/api/games/999", "", nil)
	if code != http.StatusForbidden || !bytes.Equal(exists, missing) {
		t.Fatalf("%d: %s, then %s", code, exists, missing)
	}
	if _, body := remote(t, s, "GET", "/api/games", "", nil); strings.TrimSpace(string(body)) != "[]" {
		t.Fatalf("list: %s", body)
	}
	for _, c := range []struct{ method, path string }{
		{"POST", "/api/games"}, {"POST", "/api/load"}, {"GET", "/api/kialo"}, {"POST", "/api/kialo/import"},
	} {
		for _, key := range []string{"", keys.Moderator} {
			code, body := remote(t, s, c.method, c.path, key, map[string]any{})
			if code != http.StatusForbidden || !strings.Contains(string(body), "computer that runs the server") {
				t.Fatalf("%s %s with key %q: %d %s", c.method, c.path, key, code, body)
			}
		}
	}
	// The keys are the owner's to share.
	for _, key := range []string{keys.Moderator, keys.Seats[0].Key} {
		if code, _ := remote(t, s, "GET", "/api/games/"+ss.ID+"/keys", key, nil); code != http.StatusForbidden {
			t.Fatalf("keys with a key: %d", code)
		}
	}
	// The schemes and suggestions are for everyone.
	if code, _ := remote(t, s, "GET", "/api/meta", "", nil); code != http.StatusOK {
		t.Fatalf("meta: %d", code)
	}
}

// A page of another site whose name resolves to this computer is not the
// owner: its requests come from this computer but name another host.
func TestAPageOfAnotherSiteIsNotTheOwner(t *testing.T) {
	s, _ := newTestServerWith(t)
	ss, _ := moderatedGame(t, s)
	for _, host := range []string{"evil.example:8080", "wag.example"} {
		r := httptest.NewRequest("POST", "http://"+host+"/api/games/"+ss.ID+"/decide", strings.NewReader(`{"statement":"p","decision":"accepted"}`))
		r.RemoteAddr = "127.0.0.1:5555"
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != http.StatusForbidden || len(ss.GS.MoveHistory) != 0 {
			t.Fatalf("%s: %d %s", host, w.Code, w.Body)
		}
	}
	// Another computer that names this one localhost is another computer.
	for _, host := range []string{"localhost:8080", "127.0.0.1:8080"} {
		r := httptest.NewRequest("POST", "http://"+host+"/api/games/"+ss.ID+"/decide", strings.NewReader(`{"statement":"p","decision":"accepted"}`))
		r.RemoteAddr = "198.51.100.23:1234"
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != http.StatusForbidden || len(ss.GS.MoveHistory) != 0 {
			t.Fatalf("another computer as %s: %d %s", host, w.Code, w.Body)
		}
	}
	// This computer, named as this computer, is the owner, as before.
	for _, host := range []string{"localhost:8080", "127.0.0.1:8080", "[::1]:8080"} {
		r := httptest.NewRequest("GET", "http://"+host+"/api/games/"+ss.ID+"/role", nil)
		r.RemoteAddr = "127.0.0.1:5555"
		if strings.HasPrefix(host, "[") {
			r.RemoteAddr = "[::1]:5555"
		}
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"owner"`) {
			t.Fatalf("%s: %d %s", host, w.Code, w.Body)
		}
	}
}

// A key writes notes in its holder's name, whatever name the note gives.
func TestANoteFromAKeyIsInItsHoldersName(t *testing.T) {
	s, _ := newTestServerWith(t)
	ss, keys := moderatedGame(t, s)
	note := map[string]string{"target": "game", "kind": "idea", "text": "a remark", "by": "moderator"}
	if code, body := remote(t, s, "POST", "/api/games/"+ss.ID+"/note", keys.Seats[1].Key, note); code != http.StatusOK {
		t.Fatalf("%d %s", code, body)
	}
	note["by"] = "B"
	if code, body := remote(t, s, "POST", "/api/games/"+ss.ID+"/note", keys.Moderator, note); code != http.StatusOK {
		t.Fatalf("%d %s", code, body)
	}
	if len(ss.Notes) != 2 || ss.Notes[0].By != "B" || ss.Notes[1].By != "moderator" {
		t.Fatalf("notes %+v", ss.Notes)
	}
}

// Each key reads the game and says who it is; the owner gets the keys, one
// for the moderator and one for each seat a person plays.
func TestRolesAndTheOwnersKeys(t *testing.T) {
	s, ts := newTestServerWith(t)
	var st State
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Seats: []Seat{{Name: "Ann", Kind: "human"}, {Name: "Bot", Kind: "greedy"}, {Name: "Pat", Kind: "llm"}}}, &st); code != http.StatusOK {
		t.Fatalf("new game: %d", code)
	}
	var keys KeysView
	if code := call(t, ts, "GET", "/api/games/"+st.ID+"/keys", nil, &keys); code != http.StatusOK {
		t.Fatalf("keys: %d", code)
	}
	if keys.Moderator == "" || len(keys.Seats) != 2 || keys.Seats[0].Name != "Ann" || keys.Seats[1].Name != "Pat" ||
		keys.Seats[0].Key == keys.Seats[1].Key || keys.Seats[0].Key == keys.Moderator || len(keys.Moderator) < 26 {
		t.Fatalf("keys %+v", keys)
	}
	for key, want := range map[string]RoleView{keys.Moderator: {Role: "moderator"}, keys.Seats[0].Key: {Role: "seat", Seat: "Ann"}, keys.Seats[1].Key: {Role: "seat", Seat: "Pat"}} {
		code, body := remote(t, s, "GET", "/api/games/"+st.ID+"/role", key, nil)
		var got RoleView
		if code != http.StatusOK || json.Unmarshal(body, &got) != nil || got != want {
			t.Fatalf("role: %d %s, want %+v", code, body, want)
		}
		for _, path := range []string{"", "/export", "/traces", "/prompt"} {
			if code, _ := remote(t, s, "GET", "/api/games/"+st.ID+path, key, nil); code != http.StatusOK {
				t.Fatalf("read %q: %d", path, code)
			}
		}
		if code, _ := remote(t, s, "POST", "/api/games/"+st.ID+"/preview", key, map[string]any{"action": map[string]string{"type": "abstain"}}); code != http.StatusOK {
			t.Fatalf("preview: %d", code)
		}
	}
	// No key is no role, though the bot's seat has no key either.
	if code, _ := remote(t, s, "GET", "/api/games/"+st.ID, "", nil); code != http.StatusForbidden {
		t.Fatalf("no key on a game with a bot: %d", code)
	}
	// An action the roles do not name is the owner's.
	if code, _ := remote(t, s, "GET", "/api/games/"+st.ID+"/somethingnew", keys.Moderator, nil); code != http.StatusForbidden {
		t.Fatalf("an unknown action: %d", code)
	}
	// A key lets a bot take its turn.
	call(t, ts, "POST", "/api/games/"+st.ID+"/move", map[string]any{"action": map[string]string{"type": "abstain"}}, nil)
	if code, body := remote(t, s, "POST", "/api/games/"+st.ID+"/ai", keys.Seats[1].Key, map[string]any{}); code != http.StatusOK {
		t.Fatalf("ai: %d %s", code, body)
	}
}

// The keys are in no state, export or trace.
func TestKeysAreNotInTheGamesRecords(t *testing.T) {
	s, ts := newTestServerWith(t)
	ss, keys := moderatedGame(t, s)
	all := []string{keys.Moderator}
	for _, k := range keys.Seats {
		all = append(all, k.Key)
	}
	for _, path := range []string{"", "/export", "/traces", "/prompt", "/role"} {
		req, _ := http.NewRequest("GET", ts.URL+"/api/games/"+ss.ID+path, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var b bytes.Buffer
		b.ReadFrom(resp.Body)
		resp.Body.Close()
		for _, k := range all {
			if strings.Contains(b.String(), k) {
				t.Fatalf("%q holds a key", path)
			}
		}
	}
	var list bytes.Buffer
	resp, _ := http.Get(ts.URL + "/api/games")
	list.ReadFrom(resp.Body)
	resp.Body.Close()
	for _, k := range all {
		if strings.Contains(list.String(), k) {
			t.Fatal("the list holds a key")
		}
	}
}
