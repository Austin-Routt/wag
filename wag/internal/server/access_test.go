package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/llm"
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

// thisMachinesAddress is a network address of this computer that is not a
// loopback address, if it has one.
func thisMachinesAddress(t *testing.T) net.IP {
	t.Helper()
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok && n.IP.To4() != nil && !n.IP.IsLoopback() {
			return n.IP
		}
	}
	t.Skip("no network address on this machine but loopback")
	return nil
}

// A server that listens on one address of the network (--addr <address>)
// is opened there by its owner, on this computer: that is the owner (the
// review's R3), as long as the Host is an IP address, not a name.
func TestThisComputerAtItsNetworkAddressIsTheOwner(t *testing.T) {
	lan := thisMachinesAddress(t)
	s, err := New(builtinData(), "")
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp", net.JoinHostPort(lan.String(), "0"))
	if err != nil {
		t.Skip(err)
	}
	ts := httptest.NewUnstartedServer(s.Handler())
	ts.Listener.Close()
	ts.Listener = ln
	ts.Start()
	defer ts.Close()
	var st State
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Seats: twoHumans}, &st); code != http.StatusOK {
		t.Fatalf("listening on %s, this computer's page at %s cannot start a game: %d", ln.Addr(), ts.URL, code)
	}
	var keys KeysView
	if code := call(t, ts, "GET", "/api/games/"+st.ID+"/keys", nil, &keys); code != http.StatusOK || keys.Moderator == "" {
		t.Fatalf("keys: %d", code)
	}
	// The workbench too.
	if code := call(t, ts, "GET", "/api/llm/services", nil, nil); code != http.StatusOK {
		t.Fatalf("the workbench at this computer's address: %d", code)
	}
	// A name for this computer is not the owner, even from this computer.
	req, _ := http.NewRequest("GET", ts.URL+"/api/games/"+st.ID+"/keys", nil)
	req.Host = "wag.example"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("this computer, named wag.example: %d", resp.StatusCode)
	}
}

// Who counts as this computer: a loopback client, or a client at the very
// address it reached; with a Host that is localhost or an IP address.
func TestWhatCountsAsThisComputer(t *testing.T) {
	at := func(remote, local, host string) bool {
		r := httptest.NewRequest("GET", "http://"+host+"/", nil)
		r.RemoteAddr = remote
		if local != "" {
			la, err := net.ResolveTCPAddr("tcp", local)
			if err != nil {
				t.Fatal(err)
			}
			r = r.WithContext(context.WithValue(r.Context(), http.LocalAddrContextKey, la))
		}
		return fromThisComputer(r)
	}
	for _, c := range []struct {
		remote, local, host string
		want                bool
	}{
		{"127.0.0.1:5000", "127.0.0.1:8080", "localhost:8080", true},
		{"[::1]:5000", "[::1]:8080", "[::1]:8080", true},
		{"127.0.0.1:5000", "127.0.0.1:8080", "0.0.0.0:8080", true},
		{"192.0.2.7:5000", "192.0.2.7:8080", "192.0.2.7:8080", true},
		{"192.0.2.7:5000", "192.0.2.7:8080", "wag.example:8080", false},
		{"192.0.2.9:5000", "192.0.2.7:8080", "192.0.2.7:8080", false},
		{"192.0.2.9:5000", "", "192.0.2.7:8080", false},
		{"192.0.2.9:5000", "0.0.0.0:8080", "192.0.2.7:8080", false},
		{"127.0.0.1:5000", "127.0.0.1:8080", "evil.example", false},
		{"127.0.0.1:5000", "127.0.0.1:8080", "LOCALHOST", false},
	} {
		if got := at(c.remote, c.local, c.host); got != c.want {
			t.Errorf("from %s to %s as %s: %v", c.remote, c.local, c.host, got)
		}
	}
}

// Another computer's trace and export name no file of the owner's: the model
// profiles keep only their name, provider and parameters, and the data folder
// and home folder, which an error may name, are placeholders (the review's
// R5). The owner's own copies keep everything.
func TestAGuestsTraceAndExportNameNoFileOfTheOwner(t *testing.T) {
	srv, ts := newTestServerWith(t)
	store := t.TempDir()
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	schema := filepath.Join(store, "move-schema.json")
	srv.AddModel(llm.WithProvider(llm.Profile{Name: "m", Provider: "mock", SchemaFile: schema, BaseURL: "http://127.0.0.1:1234/v1",
		Command: []string{filepath.Join(store, "model.sh")}, Retries: -1, Params: llm.Params{Model: "m-1"}},
		&llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
			return nil, &llm.Error{Provider: "mock", Kind: llm.KindAuth, Message: "cannot run " + filepath.Join(store, "model.sh")}
		}}))
	var st State
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Seats: []Seat{{Name: "Ann", Kind: "human"}, {Name: "M", Kind: "model", Model: "m"}}}, &st); code != http.StatusOK {
		t.Fatalf("new game: %d", code)
	}
	post(t, ts, "/api/games/"+st.ID+"/move", map[string]any{"action": map[string]string{"type": "abstain"}})
	call(t, ts, "POST", "/api/games/"+st.ID+"/ai", map[string]any{}, nil) // the model's turn fails, naming the folder
	keys := srv.sessions[st.ID].keysView()
	for _, what := range []string{"traces", "export"} {
		code, body := remote(t, srv, "GET", "/api/games/"+st.ID+"/"+what, keys.Seats[0].Key, nil)
		if code != http.StatusOK || strings.Contains(string(body), store) || strings.Contains(string(body), "127.0.0.1:1234") {
			i := strings.Index(string(body), store)
			t.Fatalf("%s for a seat's key: %d, names the folder at %d", what, code, i)
		}
		if what == "traces" && (!strings.Contains(string(body), `"name":"m"`) || !strings.Contains(string(body), "<data folder>")) {
			t.Fatalf("the guest's trace lost the profile or the error: %.300s", body)
		}
		// The owner's copy keeps the folder.
		req, _ := http.NewRequest("GET", ts.URL+"/api/games/"+st.ID+"/"+what, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var b bytes.Buffer
		b.ReadFrom(resp.Body)
		resp.Body.Close()
		if what == "traces" && !strings.Contains(b.String(), schema) {
			t.Fatalf("the owner's trace lost the profile's schema file")
		}
	}
}

// guestCopy finds the folder in either JSON spelling: an export escapes <, >
// and &, a trace does not.
func TestGuestCopyFindsBothSpellings(t *testing.T) {
	s := &Server{store: "/data/a&b"}
	amp := string([]byte{92}) + "u0026" // & as an export's JSON writes it
	got := string(s.guestCopy([]byte(`{"x":"/data/a` + amp + `b/game","y":"/data/a&b/trace"}`)))
	if strings.Contains(got, "/data/a") || strings.Count(got, "<data folder>") != 2 {
		t.Fatal(got)
	}
	// Nothing to hide: unchanged.
	if got := string((&Server{}).guestCopy([]byte(`{"x":"/"}`))); got != `{"x":"/"}` {
		t.Fatal(got)
	}
}
