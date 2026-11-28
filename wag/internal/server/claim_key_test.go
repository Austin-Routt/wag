package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// K1 (docs/c5-item-f.md, section K; the independent check of F1, F2 and F4,
// E2): a new game whose claim is written otherwise than the engine keys it is
// made under the key, and the server keeps answering. Such a claim made the
// new game's request fail while it held the server's lock, so no other
// request was answered.
func TestK1AClaimAsTheEngineKeysIt(t *testing.T) {
	s, err := New(builtinData(), "")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	// Closed only if the server still answers: a request stuck on the lock
	// would keep Close waiting.
	defer func() {
		if !t.Failed() {
			ts.Close()
		}
	}()
	client := &http.Client{Transport: &http.Transport{Proxy: nil}, Timeout: 20 * time.Second}
	for _, c := range []struct{ claim, key string }{{"¬ p", "¬p"}, {"¬(p)", "¬p"}, {"¬¬p", "p"}, {"¬ ¬ p", "p"},
		{"¬¬worth_finding_out_first(tax_meat,demand_falls)", "worth_finding_out_first(tax_meat,demand_falls)"}} {
		b, _ := json.Marshal(NewGameRequest{Mode: "inquiry", Claim: c.claim, Seats: twoHumans})
		resp, err := client.Post(ts.URL+"/api/games", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Errorf("K1 %s: the new game: %v", c.claim, err)
		} else {
			var st State
			err := json.NewDecoder(resp.Body).Decode(&st)
			resp.Body.Close()
			if err != nil || resp.StatusCode != http.StatusOK || st.Claim != c.key || st.Statements[c.key] == nil {
				t.Errorf("K1 %s: HTTP %d %v, the claim %q (want %q) on the board: %v", c.claim, resp.StatusCode, err, st.Claim, c.key, st.Statements[c.key] != nil)
			}
		}
		resp, err = client.Get(ts.URL + "/api/games")
		if err != nil {
			t.Fatalf("K1 %s: then the list of games: %v", c.claim, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("K1 %s: then the list of games: HTTP %d", c.claim, resp.StatusCode)
		}
	}
}

// K4: building the state of a session whose claim is not on the board does
// not fail, and leaves the claim out of the roots.
func TestK4AClaimNotOnTheBoard(t *testing.T) {
	s, ts := newTestServerWith(t)
	var st State
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "inquiry", Claim: "p", Seats: twoHumans}, &st); code != http.StatusOK {
		t.Fatalf("new game: %d", code)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ss := s.sessions[st.ID]
	ss.GS.Claim = "q_not_on_the_board"
	var got *State
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("K4: building the state failed: %v", r)
			}
		}()
		got = ss.buildState()
	}()
	if got != nil {
		for _, r := range got.Roots {
			if r == "q_not_on_the_board" {
				t.Errorf("K4: the claim is among the roots %v", got.Roots)
			}
		}
	}
}
