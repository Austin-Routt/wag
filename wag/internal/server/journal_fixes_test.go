package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/llm"
	"github.com/yourorg/wag/internal/nascq"
)

// The corrections after the independent check of 90d2327
// (docs/g-journal.md, "Corrections after the independent check"): cases G11
// to G17, G10b, G5f to G5i, G4c, G3c and Gn, each test's outcome the plan's,
// written before the tests ran.

// requests is a copy of the requests the provider was sent.
func requests(p *provider) []llm.Request {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]llm.Request(nil), p.reqs...)
}

// copyStore copies the data folder as it is now: what a server restarted on
// it would find.
func copyStore(t *testing.T, store string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS(store)); err != nil {
		t.Fatal(err)
	}
	return dir
}

// playedIn is the time the game clock of the server's game counts as played.
func playedIn(srv *Server, id string) time.Duration {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return time.Since(srv.sessions[id].GS.StartedAt)
}

// decode reads JSON with its numbers as they are written.
func decode(t *testing.T, b []byte) map[string]any {
	t.Helper()
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var m map[string]any
	if err := d.Decode(&m); err != nil {
		t.Fatal(err)
	}
	return m
}

// editExport edits a game folder's export.
func editExport(t *testing.T, dir string, edit func(map[string]any)) {
	t.Helper()
	p := filepath.Join(dir, "export.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	m := decode(t, b)
	edit(m)
	out, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, out, 0o644); err != nil {
		t.Fatal(err)
	}
}

// editTrace edits a game folder's trace, event by event: keep false drops
// the event.
func editTrace(t *testing.T, dir string, edit func(i int, e map[string]any) (keep bool)) {
	t.Helper()
	p := filepath.Join(dir, "trace.jsonl")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for i, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		e := decode(t, []byte(l))
		if !edit(i, e) {
			continue
		}
		x, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, string(x))
	}
	if err := os.WriteFile(p, []byte(strings.Join(out, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func moderated(second Seat) NewGameRequest {
	req := debate(second)
	req.Moderated = true
	return req
}

func stateOf(t *testing.T, ts *httptest.Server, id string) State {
	t.Helper()
	var st State
	if code := call(t, ts, "GET", "/api/games/"+id, nil, &st); code != http.StatusOK {
		t.Fatalf("the game %s: %d", id, code)
	}
	return st
}

// G11: the moderator rejects the model's held move; the model's next turn
// starts afresh: one message, the board, which names the moderator's note.
// After four rejections the model is still asked.
func TestG11AHeldMoveRejectedStartsTheTurnAfresh(t *testing.T) {
	store := t.TempDir()
	_, ts, p := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, moderated(Seat{Name: "Ann", Kind: "human"}))
	for i := 1; i <= 5; i++ {
		mr := post(t, ts, "/api/games/"+id+"/ai", map[string]any{})
		if mr.State.Pending == nil || p.calls() != i {
			t.Fatalf("turn %d: held %v, %d calls", i, mr.State.Pending != nil, p.calls())
		}
		if i > 1 {
			m := requests(p)[i-1].Messages
			if len(m) != 1 || !strings.Contains(m[0].Content, "NOTE-"+itoa(i-1)) {
				t.Fatalf("turn %d: the request has %d messages, the board naming the note: %v", i, len(m), len(m) > 0 && strings.Contains(m[0].Content, "NOTE-"+itoa(i-1)))
			}
		}
		post(t, ts, "/api/games/"+id+"/moderate", map[string]any{"approve": false, "note": "NOTE-" + itoa(i)})
	}
	for _, d := range ofType(events(t, filepath.Join(store, folder(t, store), "trace.jsonl")), "decision") {
		if d["attempt"] != float64(1) || d["feedback"] != nil {
			t.Fatalf("a decision of a turn after a rejection: %v", d)
		}
	}
}

// G12: a moderated game stopped while the model's reply is held, resumed:
// the move is held again, /ai asks nothing, and the moderator's approval
// plays it.
func TestG12AHeldMoveIsHeldAgainAfterAResume(t *testing.T) {
	store := t.TempDir()
	_, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, moderated(Seat{Name: "Ann", Kind: "human"}))
	mr := post(t, ts, "/api/games/"+id+"/ai", map[string]any{})
	if mr.State.Pending == nil {
		t.Fatal("the model's move is not held")
	}
	held := mr.State.Pending
	stopped := copyStore(t, store)
	_, ts2, p2, id2 := resumed(t, stopped, "m-1", script(replyPass))
	st := stateOf(t, ts2, id2)
	if st.Pending == nil || st.Pending.Player != "M" || !sameJSON(st.Pending.Action, held.Action) || st.Pending.Description != held.Description {
		t.Fatalf("after the resume: held %+v, want %+v", st.Pending, held)
	}
	if code, msg := ai(t, ts2, id2); code == http.StatusOK || !strings.Contains(msg, "awaits the moderator") || p2.calls() != 0 {
		t.Fatalf("/ai while the move is held: %d %q, %d calls", code, msg, p2.calls())
	}
	mr = post(t, ts2, "/api/games/"+id2+"/moderate", map[string]any{"approve": true})
	if mr.Result["accepted"] != true || mr.State.Pending != nil || mr.State.Active != "Ann" {
		t.Fatalf("approved: %v, %s to move", mr.Result, mr.State.Active)
	}
	h := exportOfFolder(t, filepath.Join(stopped, folder(t, stopped))).History
	if len(h) != 1 || h[0].PlayerID != "M" || h[0].Moderation != "approved" || !sameJSON(h[0].Action, held.Action) {
		t.Fatalf("the move played: %+v", h)
	}

	// Also checked: an export that holds a move the trace's decision did not
	// hold, or holds it for a player whose turn it is not, is refused.
	for name, edit := range map[string]func(dir string){
		"not held": func(dir string) {
			editTrace(t, dir, func(_ int, e map[string]any) bool {
				if e["type"] == "decision" {
					delete(e["result"].(map[string]any), "pending")
				}
				return true
			})
			editExport(t, dir, func(m map[string]any) {
				for _, d := range m["traces"].([]any) {
					delete(d.(map[string]any)["result"].(map[string]any), "pending")
				}
			})
		},
	} {
		dir := copyStore(t, store)
		edit(filepath.Join(dir, folder(t, dir)))
		_, ts3, _ := journalServer(t, dir, "m-1", script(replyPass))
		code, out := resumeHTTP(t, ts3, folder(t, dir))
		if msg, _ := out["error"].(string); code == http.StatusOK || !strings.Contains(msg, "holds a move for the moderator that the game does not") {
			t.Errorf("%s: %d %q", name, code, msg)
		}
	}
	// The held move given to a player whose last decision was held too
	// (M's claim, held then approved), whose turn it is not: Ann's move is
	// held.
	store = t.TempDir()
	_, ts, _ = journalServer(t, store, "m-1", script(replyClaim))
	id = newGame(t, ts, moderated(Seat{Name: "Ann", Kind: "human"}))
	post(t, ts, "/api/games/"+id+"/ai", map[string]any{})
	post(t, ts, "/api/games/"+id+"/moderate", map[string]any{"approve": true})
	undercut := map[string]any{"type": "attack", "attack_type": "undercutter", "target_arg_id": "a1", "scheme_id": "AS:expert_opinion",
		"answers_cq_id": "CQ:gain_discounted", "variable_bindings": map[string]string{"D": "the subject at hand", "W": "a reviewer"}}
	if mr := post(t, ts, "/api/games/"+id+"/move", map[string]any{"action": undercut}); mr.State.Pending == nil || mr.State.Pending.Player != "Ann" {
		t.Fatalf("Ann's move is not held: %v", mr.Result)
	}
	dir := copyStore(t, store)
	editExport(t, filepath.Join(dir, folder(t, dir)), func(m map[string]any) {
		m["state"].(map[string]any)["pending"].(map[string]any)["player"] = "M"
	})
	_, ts4, _ := journalServer(t, dir, "m-1", script(replyPass))
	code, out := resumeHTTP(t, ts4, folder(t, dir))
	if msg, _ := out["error"].(string); code == http.StatusOK || !strings.Contains(msg, "it is not their turn") {
		t.Errorf("a held move of the player whose turn it is not: %d %q", code, msg)
	}
}

// G12b: the same with a random player's held move: it is held again, and
// after the moderator's rejection the random player chooses as in the game
// that did not stop.
func TestG12bAComputerPlayersHeldMoveIsHeldAgain(t *testing.T) {
	store := t.TempDir()
	_, ts, _ := journalServer(t, store, "m-1", script(replyClaim, replyPass))
	id := newGame(t, ts, moderated(Seat{Name: "R", Kind: "random"}))
	post(t, ts, "/api/games/"+id+"/ai", map[string]any{})
	post(t, ts, "/api/games/"+id+"/moderate", map[string]any{"approve": true})
	mr := post(t, ts, "/api/games/"+id+"/ai", map[string]any{})
	if mr.State.Pending == nil || mr.State.Pending.Player != "R" {
		t.Fatal("R's move is not held")
	}
	held := mr.State.Pending
	stopped := copyStore(t, store)
	// The game that did not stop.
	post(t, ts, "/api/games/"+id+"/moderate", map[string]any{"approve": false})
	mr = post(t, ts, "/api/games/"+id+"/ai", map[string]any{})
	if mr.State.Pending == nil {
		t.Fatal("R's next move is not held")
	}
	next := mr.State.Pending
	if sameJSON(next.Action, held.Action) {
		t.Fatal("the fixture does not tell the draws apart: R's next move is the one rejected")
	}
	// Resumed.
	_, ts2, p2, id2 := resumed(t, stopped, "m-1", script(replyPass))
	st := stateOf(t, ts2, id2)
	if st.Pending == nil || st.Pending.Player != "R" || !sameJSON(st.Pending.Action, held.Action) {
		t.Fatalf("after the resume: held %+v, want %+v", st.Pending, held)
	}
	post(t, ts2, "/api/games/"+id2+"/moderate", map[string]any{"approve": false})
	mr = post(t, ts2, "/api/games/"+id2+"/ai", map[string]any{})
	if mr.State.Pending == nil || !sameJSON(mr.State.Pending.Action, next.Action) || p2.calls() != 0 {
		t.Fatalf("R's next move after the resume: %+v, want %+v (%d calls)", mr.State.Pending, next, p2.calls())
	}
}

// G13: a game with a game clock, resumed twice after stops of a few
// seconds: the time played counted is the time played, not the stops.
func TestG13TwoResumesCountOnlyTheTimePlayed(t *testing.T) {
	store := t.TempDir()
	_, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	req := debate(Seat{Name: "Ann", Kind: "human"})
	limit := 600
	req.TimeLimitSec = &limit
	id := newGame(t, ts, req)
	if code, msg := ai(t, ts, id); code != http.StatusOK {
		t.Fatalf("M's move: %d %q", code, msg)
	}
	stop1 := copyStore(t, store)
	time.Sleep(1500 * time.Millisecond) // the server is down
	_, ts2, _, id2 := resumed(t, stop1, "m-1", script(replyPass))
	post(t, ts2, "/api/games/"+id2+"/move", map[string]any{"action": map[string]string{"type": "abstain"}})
	stop2 := copyStore(t, stop1)
	time.Sleep(1500 * time.Millisecond) // down again
	srv3, _, _, id3 := resumed(t, stop2, "m-1", script(replyPass))
	if played := playedIn(srv3, id3); played > time.Second {
		t.Fatalf("after two resumes the game clock counts %v as played", played)
	}
	// Also checked: the first resume came after a decision, and the events
	// after it go on in order.
	evs := events(t, filepath.Join(stop2, folder(t, stop2), "trace.jsonl"))
	if len(ofType(evs, "decision")) != 2 {
		t.Fatalf("%d decisions", len(ofType(evs, "decision")))
	}
	for i := 2; i < len(evs); i++ {
		if a, b := evs[i-1]["order"].(float64), evs[i]["order"].(float64); b <= a {
			t.Fatalf("the events' order: %v then %v", a, b)
		}
	}
}

// G10b: the game clock counts the time played its export records.
func TestG10bTheTimePlayedIsTheExports(t *testing.T) {
	store := t.TempDir()
	_, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	req := debate(Seat{Name: "Ann", Kind: "human"})
	limit := 600
	req.TimeLimitSec = &limit
	id := newGame(t, ts, req)
	ai(t, ts, id)
	stopped := copyStore(t, store)
	editExport(t, filepath.Join(stopped, folder(t, stopped)), func(m map[string]any) {
		if _, ok := m["played_ns"]; !ok {
			t.Fatal("the export records no time played")
		}
		m["played_ns"] = json.Number("300000000000")
	})
	srv2, _, _, id2 := resumed(t, stopped, "m-1", script(replyPass))
	if played := playedIn(srv2, id2); played < 299*time.Second || played > 302*time.Second {
		t.Fatalf("the game clock counts %v, not five minutes", played)
	}
}

// G14: a move clock of one second; the save fails after the provider's
// answer and succeeds two seconds later: the kept reply is played as the
// move, not voided.
func TestG14AKeptReplyIsNotVoidedByTheMoveClock(t *testing.T) {
	store := t.TempDir()
	_, ts, p := journalServer(t, store, "m-1", script(replyClaim))
	req := debate(Seat{Name: "Ann", Kind: "human"})
	move := 1
	req.MoveTimeSec = &move
	id := newGame(t, ts, req)
	unblock := blockGames(t, store)
	if code, msg := ai(t, ts, id); code == http.StatusOK || !strings.Contains(msg, "the reply is played once the game is saved") {
		t.Fatalf("the unsaved turn: %d %q", code, msg)
	}
	time.Sleep(2 * time.Second)
	unblock()
	if code, msg := ai(t, ts, id); code != http.StatusOK || p.calls() != 1 {
		t.Fatalf("once saved: %d %q, %d calls", code, msg, p.calls())
	}
	h := exportOfFolder(t, filepath.Join(store, folder(t, store))).History
	if len(h) != 1 || h[0].TimedOut || h[0].Action.Type != "support" || len(h[0].ArgIDs) == 0 {
		t.Fatalf("the kept reply's move: %+v", h)
	}
}

// retryingServer is journalServer with retries for the profile's calls.
func retryingServer(t *testing.T, store string, retries int, answer func(int, llm.Request) (*llm.Response, error)) (*httptest.Server, *provider) {
	t.Helper()
	srv, ts := newTestServerWith(t)
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	p := &provider{answer: answer}
	c := llm.WithProvider(llm.Profile{Name: "m", Provider: "mock", Retries: retries, Params: llm.Params{Model: "m-1"}}, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		n := int(p.n.Add(1))
		p.mu.Lock()
		p.reqs = append(p.reqs, req)
		p.mu.Unlock()
		return p.answer(n, req)
	}})
	c.Sleep = llm.NoWait
	srv.AddModel(c)
	return ts, p
}

// G15: retries allowed; the first try fails with a server error, and its
// record cannot be saved: no second try is sent.
func TestG15NoRetryWhileTheGameCannotBeSaved(t *testing.T) {
	store := t.TempDir()
	var unblock func()
	ts, p := retryingServer(t, store, 2, func(n int, _ llm.Request) (*llm.Response, error) {
		if n == 1 {
			unblock = blockGames(t, store) // the disk fails now
			return nil, &llm.Error{Provider: "mock", Kind: llm.KindServer, Message: "overloaded"}
		}
		return &llm.Response{Text: replyClaim, FinishReason: "stop", Model: "m-1"}, nil
	})
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	if code, msg := ai(t, ts, id); code == http.StatusOK || p.calls() != 1 || !strings.Contains(msg, "could not be saved") {
		t.Fatalf("the turn: %d %q, %d calls", code, msg, p.calls())
	}
	unblock()
	if code, msg := ai(t, ts, id); code != http.StatusOK || p.calls() != 2 {
		t.Fatalf("once saved: %d %q, %d calls", code, msg, p.calls())
	}
}

// G16: in a live game whose time is up, /ai asks nothing, and the game is
// over; resumed, it is over at once, and /ai asks nothing. A game not over
// whose export records more time played than its game clock gives is over
// once resumed.
func TestG16AGameOverByItsClockStaysOver(t *testing.T) {
	store := t.TempDir()
	_, ts, p := journalServer(t, store, "m-1", script(replyClaim, replyPass))
	req := debate(Seat{Name: "Ann", Kind: "human"})
	limit := 1
	req.TimeLimitSec = &limit
	id := newGame(t, ts, req)
	if code, msg := ai(t, ts, id); code != http.StatusOK {
		t.Fatalf("M's move: %d %q", code, msg)
	}
	post(t, ts, "/api/games/"+id+"/move", map[string]any{"action": map[string]string{"type": "abstain"}})
	time.Sleep(1100 * time.Millisecond)
	code, msg := ai(t, ts, id)
	if st := stateOf(t, ts, id); code != http.StatusOK || p.calls() != 1 || !st.Over {
		t.Fatalf("live, the time up: %d %q, %d calls, over %v", code, msg, p.calls(), st.Over)
	}
	_, ts2, p2, id2 := resumed(t, copyStore(t, store), "m-1", script(replyPass))
	if st := stateOf(t, ts2, id2); !st.Over {
		t.Fatal("resumed, the game over by its clock is not over")
	}
	if code, _ := ai(t, ts2, id2); code == http.StatusOK || p2.calls() != 0 {
		t.Fatalf("resumed: /ai %d, %d calls", code, p2.calls())
	}

	store = t.TempDir()
	_, ts, _ = journalServer(t, store, "m-1", script(replyClaim))
	limit = 600
	id = newGame(t, ts, req)
	ai(t, ts, id)
	stopped := copyStore(t, store)
	editExport(t, filepath.Join(stopped, folder(t, stopped)), func(m map[string]any) { m["played_ns"] = json.Number("601000000000") })
	_, ts3, p3, id3 := resumed(t, stopped, "m-1", script(replyPass))
	if st := stateOf(t, ts3, id3); !st.Over {
		t.Fatal("resumed with its time up, the game is not over")
	}
	if code, _ := ai(t, ts3, id3); code == http.StatusOK || p3.calls() != 0 {
		t.Fatalf("resumed with its time up: /ai %d, %d calls", code, p3.calls())
	}

	// Also checked: the time runs out while the model answers: its move is
	// not played, and it is not asked again.
	store = t.TempDir()
	_, ts, p = journalServer(t, store, "m-1", func(int, llm.Request) (*llm.Response, error) {
		time.Sleep(1100 * time.Millisecond)
		return &llm.Response{Text: replyClaim, FinishReason: "stop", Model: "m-1"}, nil
	})
	one := 1
	sreq := debate(Seat{Name: "Ann", Kind: "human"})
	sreq.TimeLimitSec = &one
	id = newGame(t, ts, sreq)
	code, msg = ai(t, ts, id)
	if st := stateOf(t, ts, id); code != http.StatusOK || p.calls() != 1 || !st.Over || len(st.Log) != 0 {
		t.Fatalf("the time up while the model answers: %d %q, %d calls, over %v, %d moves", code, msg, p.calls(), st.Over, len(st.Log))
	}

	// Also checked: a move held for the moderator when the time was up is
	// not played, and is no longer held.
	store = t.TempDir()
	_, ts, _ = journalServer(t, store, "m-1", script(replyClaim))
	mreq := moderated(Seat{Name: "Ann", Kind: "human"})
	mreq.TimeLimitSec = &limit
	id = newGame(t, ts, mreq)
	if mr := post(t, ts, "/api/games/"+id+"/ai", map[string]any{}); mr.State.Pending == nil {
		t.Fatal("the model's move is not held")
	}
	stopped = copyStore(t, store)
	editExport(t, filepath.Join(stopped, folder(t, stopped)), func(m map[string]any) { m["played_ns"] = json.Number("601000000000") })
	_, ts4, _, id4 := resumed(t, stopped, "m-1", script(replyPass))
	if st := stateOf(t, ts4, id4); !st.Over || st.Pending != nil || len(exportOfFolder(t, filepath.Join(stopped, folder(t, stopped))).History) != 0 {
		t.Fatalf("resumed with its time up and a move held: over %v, held %v", st.Over, st.Pending)
	}
}

// Found while correcting W5: the moderator's approval of a held move after
// the game's time is up ends the game without the move, and the move before
// keeps its own verdict (it was given the approval's).
func TestTheVerdictIsTheMovesItMade(t *testing.T) {
	store := t.TempDir()
	srv, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	req := moderated(Seat{Name: "Ann", Kind: "human"})
	limit := 1
	req.TimeLimitSec = &limit
	id := newGame(t, ts, req)
	// The game's clock is the test's: it moves only when the test moves it,
	// so the one-second limit is not reached while the requests run, however
	// loaded the machine (a gate on two toolchains at once missed it).
	var clockMu sync.Mutex
	now := time.Now()
	srv.mu.Lock()
	srv.sessions[id].GS.SetClock(func() time.Time { clockMu.Lock(); defer clockMu.Unlock(); return now })
	srv.mu.Unlock()
	post(t, ts, "/api/games/"+id+"/ai", map[string]any{})
	post(t, ts, "/api/games/"+id+"/moderate", map[string]any{"approve": false, "note": "no"})
	if mr := post(t, ts, "/api/games/"+id+"/ai", map[string]any{}); mr.State.Pending == nil {
		t.Fatal("the model's second move is not held")
	}
	clockMu.Lock()
	now = now.Add(1100 * time.Millisecond)
	clockMu.Unlock()
	if mr := post(t, ts, "/api/games/"+id+"/moderate", map[string]any{"approve": true}); !mr.State.Over {
		t.Fatal("the approval after the time is up does not end the game")
	}
	h := exportOfFolder(t, filepath.Join(store, folder(t, store))).History
	if len(h) != 1 || h[0].Rejection == "" || h[0].Moderation != "rejected: no" {
		t.Fatalf("the moves: %+v", h)
	}
	// A record that gave the rejected move the approval's verdict, as WAG did
	// until this correction, loads with the move rejected.
	b, err := os.ReadFile(filepath.Join(store, folder(t, store), "export.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := decode(t, b)
	m["history"].([]any)[0].(map[string]any)["Moderation"] = "approved"
	if b, err = json.Marshal(m); err != nil {
		t.Fatal(err)
	}
	var st State
	if code := call(t, ts, "POST", "/api/load", string(b), &st); code != http.StatusOK {
		t.Fatalf("the record the old build wrote does not load: %d", code)
	}
	var e savedGame
	if err := json.Unmarshal(export(t, ts, st.ID), &e); err != nil {
		t.Fatal(err)
	}
	if len(e.History) != 1 || !strings.HasPrefix(e.History[0].Rejection, "rejected by the moderator") {
		t.Fatalf("loaded: %+v", e.History)
	}
}

// G17: the start's build edited (another commit): the game resumes, its
// start keeps the edited build, and the resume records this server's. The
// start's pins are compared where it recorded them; a start that recorded
// none (begun before they were) keeps none.
func TestG17AnotherBuildResumesAndIsRecorded(t *testing.T) {
	store := t.TempDir()
	_, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	ai(t, ts, id)
	edited := func(edit func(start map[string]any)) string {
		dir := copyStore(t, store)
		editTrace(t, filepath.Join(dir, folder(t, dir)), func(i int, e map[string]any) bool {
			if i == 0 {
				edit(e)
			}
			return true
		})
		return dir
	}
	start := func(dir string) map[string]any {
		return events(t, filepath.Join(dir, folder(t, dir), "trace.jsonl"))[0]
	}

	dir := edited(func(s map[string]any) { s["build"].(map[string]any)["vcs.revision"] = "another commit" })
	resumed(t, dir, "m-1", script(replyPass))
	s := start(dir)
	rs, _ := s["resumed"].([]any)
	if s["build"].(map[string]any)["vcs.revision"] != "another commit" || len(rs) != 1 ||
		!sameJSON(rs[0].(map[string]any)["build"], buildInfo()) || rs[0].(map[string]any)["pins"] == nil {
		t.Fatalf("the start's build %v; resumed %v", s["build"], s["resumed"])
	}

	dir = edited(func(s map[string]any) { s["pins"].(map[string]any)["engine_sha256"] = strings.Repeat("0", 64) })
	_, ts2, _ := journalServer(t, dir, "m-1", script(replyPass))
	if code, out := resumeHTTP(t, ts2, folder(t, dir)); code == http.StatusOK || !strings.Contains(out["error"].(string), "pins") {
		t.Fatalf("another engine's pins: %d %v", code, out["error"])
	}

	dir = edited(func(s map[string]any) { delete(s, "pins") })
	resumed(t, dir, "m-1", script(replyPass))
	if s := start(dir); s["pins"] != nil {
		t.Fatalf("a start that recorded no pins has %v", s["pins"])
	}
}

// G5f: a game id this server does not give, or a game's folder or file that
// is a link, is refused.
func TestG5fCraftedIdsAndLinksAreRefused(t *testing.T) {
	store := t.TempDir()
	_, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	ai(t, ts, id)
	for _, crafted := range []string{"../x", "", "007"} {
		dir := copyStore(t, store)
		game := filepath.Join(dir, folder(t, dir))
		editTrace(t, game, func(_ int, e map[string]any) bool { e["game"] = crafted; return true })
		editExport(t, game, func(m map[string]any) {
			m["state"].(map[string]any)["id"] = crafted
			for _, d := range m["traces"].([]any) {
				d.(map[string]any)["game"] = crafted
			}
		})
		_, ts2, _ := journalServer(t, dir, "m-1", script(replyPass))
		if code, out := resumeHTTP(t, ts2, folder(t, dir)); code == http.StatusOK || !strings.Contains(out["error"].(string), "not one this server gives") {
			t.Errorf("the id %q: %d %v", crafted, code, out["error"])
		}
	}

	dir := copyStore(t, store)
	elsewhere := t.TempDir()
	if err := os.CopyFS(filepath.Join(elsewhere, "g"), os.DirFS(filepath.Join(dir, folder(t, dir)))); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, folder(t, dir))); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(elsewhere, "g"), filepath.Join(dir, "games", "link")); err != nil {
		t.Skipf("no symbolic links here: %v", err)
	}
	_, ts2, _ := journalServer(t, dir, "m-1", script(replyPass))
	if code, out := resumeHTTP(t, ts2, "games/link"); code == http.StatusOK || !strings.Contains(out["error"].(string), "not a game's folder") {
		t.Errorf("a link to a game's folder elsewhere: %d %v", code, out["error"])
	}

	dir = copyStore(t, store)
	export := filepath.Join(dir, folder(t, dir), "export.json")
	if err := os.Rename(export, filepath.Join(elsewhere, "export.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(elsewhere, "export.json"), export); err != nil {
		t.Fatal(err)
	}
	_, ts3, _ := journalServer(t, dir, "m-1", script(replyPass))
	if code, out := resumeHTTP(t, ts3, folder(t, dir)); code == http.StatusOK || !strings.Contains(out["error"].(string), "a link, or not a file") {
		t.Errorf("an export that is a link: %d %v", code, out["error"])
	}
}

// G5g: G4's folder (two moves) with, in turn, another move list in the
// export, a call made after more moves than the export has, the start's
// request edited, and the export's seed edited: each refused, naming it.
func TestG5gEachRecordThatDiffersIsNamed(t *testing.T) {
	store := t.TempDir()
	_, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	ai(t, ts, id)
	post(t, ts, "/api/games/"+id+"/move", map[string]any{"action": map[string]string{"type": "abstain"}})
	for name, c := range map[string]struct {
		export func(map[string]any)
		trace  func(i int, e map[string]any)
		says   string
	}{
		"another move list": {export: func(m map[string]any) {
			b := m["history"].([]any)[0].(map[string]any)["Action"].(map[string]any)["variable_bindings"].(map[string]any)
			b["G"] = "higher emissions"
		}, says: "made the move 1"},
		"a call after more moves": {trace: func(_ int, e map[string]any) {
			if e["type"] == "call" {
				e["turn"] = json.Number("5")
			}
		}, says: "was made after 5 moves, the export has 2"},
		"the start's request": {trace: func(i int, e map[string]any) {
			if i == 0 {
				e["request"].(map[string]any)["max_rounds"] = json.Number("5")
			}
		}, says: "the trace's request is not the export's"},
		"the export's seed": {export: func(m map[string]any) {
			m["request"].(map[string]any)["seed"] = json.Number("8")
		}, says: "the trace's seed is 7, the export's 8"},
		// Also checked: a call, or a decision, of another game.
		"a call of another game": {trace: func(_ int, e map[string]any) {
			if e["type"] == "call" {
				e["game"] = "99"
			}
		}, says: "call c0 is of the game 99"},
		"a decision of another game": {trace: func(_ int, e map[string]any) {
			if e["type"] == "decision" {
				e["game"] = "99"
			}
		}, says: "decision d0 is of the game 99"},
	} {
		dir := copyStore(t, store)
		game := filepath.Join(dir, folder(t, dir))
		if c.export != nil {
			editExport(t, game, c.export)
		}
		if c.trace != nil {
			editTrace(t, game, func(i int, e map[string]any) bool { c.trace(i, e); return true })
		}
		_, ts2, _ := journalServer(t, dir, "m-1", script(replyPass))
		code, out := resumeHTTP(t, ts2, folder(t, dir))
		if msg, _ := out["error"].(string); code == http.StatusOK || !strings.Contains(msg, c.says) {
			t.Errorf("%s: %d %q", name, code, msg)
		}
	}
}

// G5h: a complete game's folder under annotation/ is not a game's folder.
func TestG5hAGameFolderOutsideGamesIsRefused(t *testing.T) {
	store := t.TempDir()
	_, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	ai(t, ts, id)
	f := folder(t, store)
	copied := filepath.Join(store, "annotation", filepath.Base(f))
	if err := os.CopyFS(copied, os.DirFS(filepath.Join(store, f))); err != nil {
		t.Fatal(err)
	}
	_, ts2, _ := journalServer(t, store, "m-1", script(replyPass))
	code, out := resumeHTTP(t, ts2, "annotation/"+filepath.Base(f))
	if msg, _ := out["error"].(string); code == http.StatusOK || !strings.Contains(msg, "is not a game's folder of the data folder") {
		t.Fatalf("%d %q", code, msg)
	}
}

// G4c: a reply planted for the current turn, whose prompt is another
// board's, is not played: the provider is asked.
func TestG4cAReplyForAnotherBoardIsNotPlayed(t *testing.T) {
	store := t.TempDir()
	_, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	ai(t, ts, id)
	post(t, ts, "/api/games/"+id+"/move", map[string]any{"action": map[string]string{"type": "abstain"}})
	srv2, ts2, p2, id2 := resumed(t, store, "m-1", script(replyPass))
	var planted string
	func() {
		srv2.mu.Lock()
		defer srv2.mu.Unlock()
		ss := srv2.sessions[id2]
		if len(ss.Calls) == 0 {
			t.Fatal("the resumed game has no call to plant again")
		}
		c := ss.Calls[0] // M's first call, on the first board
		resp := *c.Response
		resp.Text = replyPass
		c.Response = &resp
		c.ID, c.Order, c.Turn = "c"+itoa(len(ss.Calls)), ss.order+1, len(ss.GS.MoveHistory)
		ss.order++
		ss.Calls = append(ss.Calls, c)
		planted = c.ID
	}()
	if code, msg := ai(t, ts2, id2); code != http.StatusOK || p2.calls() != 1 {
		t.Fatalf("M's turn: %d %q, %d calls", code, msg, p2.calls())
	}
	for _, d := range ofType(events(t, filepath.Join(store, folder(t, store), "trace.jsonl")), "decision") {
		if d["parent"] == planted || d["recovered"] == true {
			t.Fatalf("the planted reply is played: %v", d)
		}
	}
}

// G3c: the model's first reply refused, its second call failed: the next
// turn starts afresh, one message, attempt 1; live, and resumed.
func TestG3cAFailedCallAfterARefusedReplyEndsTheTurn(t *testing.T) {
	store := t.TempDir()
	_, ts, p := journalServer(t, store, "m-1", func(n int, _ llm.Request) (*llm.Response, error) {
		switch n {
		case 1:
			return &llm.Response{Text: replyNoMove, FinishReason: "stop", Model: "m-1"}, nil
		case 2:
			return nil, &llm.Error{Provider: "mock", Kind: llm.KindAuth, Message: "the key"}
		}
		return &llm.Response{Text: replyClaim, FinishReason: "stop", Model: "m-1"}, nil
	})
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	if code, _ := ai(t, ts, id); code == http.StatusOK || p.calls() != 2 {
		t.Fatalf("the failing turn: %d, %d calls", code, p.calls())
	}
	stopped := copyStore(t, store)
	check := func(name string, p *provider, calls int, dir string) {
		t.Helper()
		r := requests(p)
		if len(r) != calls || len(r[calls-1].Messages) != 1 {
			t.Fatalf("%s: %d requests, not %d, the last with one message", name, len(r), calls)
		}
		decisions := ofType(events(t, filepath.Join(dir, folder(t, dir), "trace.jsonl")), "decision")
		if len(decisions) == 0 {
			t.Fatalf("%s: no decision", name)
		}
		if last := decisions[len(decisions)-1]; last["attempt"] != float64(1) || last["move"] != float64(0) {
			t.Fatalf("%s: the turn's decision: %v", name, last)
		}
	}
	if code, msg := ai(t, ts, id); code != http.StatusOK {
		t.Fatalf("live: %d %q", code, msg)
	}
	check("live", p, 3, store)
	_, ts2, p2, id2 := resumed(t, stopped, "m-1", script(replyClaim))
	if code, msg := ai(t, ts2, id2); code != http.StatusOK {
		t.Fatalf("resumed: %d %q", code, msg)
	}
	check("resumed", p2, 1, stopped)
}

// G5i: G5d's folder without its second call: the recorded reply was used
// and refused, so the provider is asked once, for the second attempt, with
// the conversation so far.
func TestG5iEveryRecordedReplyRefusedGoesOnWithTheConversation(t *testing.T) {
	store := t.TempDir()
	srv, ts, _ := journalServer(t, store, "m-1", script(replyRule, replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	stopped := stopAt(t, srv, 2, store)
	ai(t, ts, id)
	dir := filepath.Join(stopped, folder(t, stopped))
	calls := 0
	editTrace(t, dir, func(_ int, e map[string]any) bool {
		if e["type"] == "call" {
			calls++
			return calls == 1
		}
		return true
	})
	if calls != 2 {
		t.Fatalf("the trace has %d calls", calls)
	}
	_, ts2, p2, id2 := resumed(t, stopped, "m-1", script(replyClaim))
	if code, msg := ai(t, ts2, id2); code != http.StatusOK || p2.calls() != 1 {
		t.Fatalf("the resumed turn: %d %q, %d calls", code, msg, p2.calls())
	}
	m := requests(p2)[0].Messages
	if len(m) != 3 || m[1].Content != replyRule || !strings.HasPrefix(m[2].Content, "That move was rejected: ") {
		t.Fatalf("the second attempt's conversation: %+v", m)
	}
	decisions := ofType(events(t, filepath.Join(dir, "trace.jsonl")), "decision")
	if len(decisions) == 0 {
		t.Fatal("no decision")
	}
	if last := decisions[len(decisions)-1]; last["attempt"] != float64(2) || last["recovered"] == true || last["feedback"] != m[2].Content {
		t.Fatalf("the second attempt's decision: %v", last)
	}
}

// Gn: after a resume of game 1 in a new server, a new game gets id 2.
func TestGnANewGameAfterAResumeGetsTheNextID(t *testing.T) {
	store := t.TempDir()
	_, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	ai(t, ts, id)
	_, ts2, _, id2 := resumed(t, copyStore(t, store), "m-1", script(replyPass))
	if id2 != id {
		t.Fatalf("resumed as %s", id2)
	}
	if next := newGame(t, ts2, debate(Seat{Name: "Ann", Kind: "human"})); next != "2" {
		t.Fatalf("a new game after the resume of game %s gets the id %s", id2, next)
	}

	// Also checked: game 2 resumed in a new server, whose load of it took
	// the id 1 for a moment: a new game gets id 3, and game 2 stays.
	store = t.TempDir()
	_, ts, _ = journalServer(t, store, "m-1", script(replyClaim))
	newGame(t, ts, NewGameRequest{Mode: "classic", Seats: []Seat{{Name: "Ann", Kind: "human"}, {Name: "Bob", Kind: "human"}}}) // not saved: no model seat
	id = newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	ai(t, ts, id)
	_, ts3, _, id3 := resumed(t, copyStore(t, store), "m-1", script(replyPass))
	if id3 != "2" {
		t.Fatalf("resumed as %s", id3)
	}
	if next := newGame(t, ts3, debate(Seat{Name: "Ann", Kind: "human"})); next != "3" {
		t.Fatalf("a new game after the resume of game 2 gets the id %s", next)
	}
	if st := stateOf(t, ts3, "2"); len(st.Log) != 1 {
		t.Fatalf("game 2 after a new game: %d moves", len(st.Log))
	}
}

// drawAgent records the board (the moves made) of each draw.
type drawAgent struct {
	at    *int
	drawn []int
}

func (a *drawAgent) Name() string { return "draws" }
func (a *drawAgent) ChooseAction(*game.Observation) nascq.ProposedAction {
	a.drawn = append(a.drawn, *a.at)
	return nascq.ProposedAction{Type: "abstain"}
}

// The redraw: a computer player draws again on each board it chose on, the
// last board included, and not for its pass after its tries.
func TestRedrawChoosesAgainOnEachRecordedBoard(t *testing.T) {
	srv, ts, _ := journalServer(t, t.TempDir(), "m-1", script(replyPass))
	id := newGame(t, ts, debate(Seat{Name: "R", Kind: "random"}))
	srv.mu.Lock()
	gs, seats := srv.sessions[id].GS, srv.sessions[id].Seats
	srv.mu.Unlock()
	rd := newRedraw(gs, seats, 7, []Trace{
		{Player: "R", Turn: 0, Attempt: 1},
		{Player: "M", Turn: 1, Attempt: 1},
		{Player: "R", Turn: 2, Attempt: 1},
		{Player: "R", Turn: 3, Attempt: maxAgentTries + 1}, // the pass after five refused tries
		{Player: "R", Turn: 4, Attempt: 2},                 // the last board
	})
	if rd.agents[0] != nil || rd.agents[1] == nil {
		t.Fatalf("the agents: %v", rd.agents)
	}
	var at int
	a := &drawAgent{at: &at}
	rd.agents[1] = a
	for at = 0; at <= 4; at++ {
		rd.draw(gs, at)
	}
	if !sameJSON(a.drawn, []int{0, 2, 4}) {
		t.Fatalf("drawn on the boards %v", a.drawn)
	}
}

// A record no turn leaves is not taken for a turn in progress: the model
// is asked afresh. The records are crafted (a well-formed game never writes
// them): a call of another seat, or of another turn, among the turn's; a
// decision after a reply none used; an attempt after the last call's.
func TestTurnInProgressTakesOnlyATurnsRecord(t *testing.T) {
	srv, ts, _ := journalServer(t, t.TempDir(), "m-1", script(replyPass))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	srv.mu.Lock()
	defer srv.mu.Unlock()
	ss := srv.sessions[id]
	req := llm.Request{Messages: []llm.Message{{Role: "user", Content: ss.Prompt()}}}
	call := func(id, player string, turn, attempt int) Call {
		return Call{Format: TraceFormat, Type: "call", ID: id, Game: ss.ID, Player: player, Turn: turn, ForAttempt: attempt,
			Attempt: llm.Attempt{N: 1, Request: req, Response: &llm.Response{Text: replyNoMove}}}
	}
	refused := func(parent string, attempt int) Trace {
		return Trace{Format: TraceFormat, Type: "decision", Game: ss.ID, Player: "M", Parent: parent, Attempt: attempt, Move: -1,
			Result: TraceResult{Error: "the reply holds no move"}}
	}
	for name, c := range map[string]struct {
		calls     []Call
		decisions []Trace
		attempt   int // 0: the turn starts afresh
	}{
		"a turn's record (the first reply refused, the second unused)": {
			[]Call{call("c0", "M", 0, 1), call("c1", "M", 0, 2)}, []Trace{refused("c0", 1)}, 2},
		"a call of another seat among the turn's": {
			[]Call{call("c0", "M", 0, 1), call("c1", "Ann", 0, 2), call("c2", "M", 0, 2)}, []Trace{refused("c0", 1)}, 0},
		"a call of another turn among the turn's": {
			[]Call{call("c0", "M", 0, 1), call("c1", "M", 5, 2), call("c2", "M", 0, 2)}, []Trace{refused("c0", 1)}, 0},
		"a decision after a reply none used": {
			[]Call{call("c0", "M", 0, 1), call("c1", "M", 0, 2)}, []Trace{refused("c1", 2)}, 0},
		"an attempt after the last call's": {
			[]Call{call("c0", "M", 0, 1), call("c1", "M", 0, 4), call("c2", "M", 0, 2)}, []Trace{refused("c0", 1)}, 0},
	} {
		ss.Calls, ss.Traces = c.calls, c.decisions
		tr := ss.turnInProgress("M")
		switch {
		case c.attempt == 0 && tr != nil:
			t.Errorf("%s: taken for a turn in progress, at attempt %d", name, tr.attempt)
		case c.attempt != 0 && (tr == nil || tr.attempt != c.attempt || len(tr.msgs) != 3):
			t.Errorf("%s: %+v", name, tr)
		}
	}
	ss.Calls, ss.Traces = nil, nil
}

// Also checked with G5e: a recorded reply whose request had another
// parameter (a temperature the profile does not give) answered another
// request: it is not played, and the provider is asked.
func TestG5eAReplyToOtherParametersIsNotPlayed(t *testing.T) {
	store := t.TempDir()
	srv, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	stopped := stopAt(t, srv, 1, store)
	ai(t, ts, id)
	edited := 0
	editTrace(t, filepath.Join(stopped, folder(t, stopped)), func(_ int, e map[string]any) bool {
		if e["type"] == "call" {
			e["request"].(map[string]any)["temperature"] = json.Number("0.5")
			edited++
		}
		return true
	})
	if edited != 1 {
		t.Fatalf("%d calls edited", edited)
	}
	_, ts2, p2, id2 := resumed(t, stopped, "m-1", script(replyClaim))
	if code, msg := ai(t, ts2, id2); code != http.StatusOK || p2.calls() != 1 {
		t.Fatalf("the resumed turn: %d %q, %d calls", code, msg, p2.calls())
	}
	for _, d := range ofType(events(t, filepath.Join(stopped, folder(t, stopped), "trace.jsonl")), "decision") {
		if d["recovered"] == true {
			t.Fatalf("the reply to other parameters is played: %v", d)
		}
	}
}
