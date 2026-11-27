package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/llm"
)

// G: a model's paid reply is kept before it is used, and a game resumed from
// its folder goes on from its record (docs/g-journal.md, cases G1 to G10;
// G6 is wag-llm's, which replays). Each test's outcome is the plan's, written
// before the tests ran.

const (
	replyClaim  = `{"summary":"State the claim.","action":{"type":"support","scheme_id":"AS:consequences","variable_bindings":{"A":"tax meat","G":"lower emissions"},"declare":[{"id":"tax meat","actor":"decider"}]}}`
	replyRule   = `{"summary":"A goal of punctuation.","action":{"type":"support","scheme_id":"AS:consequences","variable_bindings":{"A":"tax meat","G":"?"},"declare":[{"id":"tax meat","actor":"decider"}]}}`
	replyNoMove = `I have nothing to say.`
	replyPass   = `{"summary":"Nothing more.","action":{"type":"abstain"}}`
)

// provider is the model's fake provider: it answers each call with what
// answer gives for it (n counts from 1), and keeps the requests.
type provider struct {
	mu     sync.Mutex
	n      atomic.Int32
	reqs   []llm.Request
	answer func(n int, req llm.Request) (*llm.Response, error)
}

func (p *provider) calls() int { return int(p.n.Load()) }

// script answers the replies in turn, the last one again when they run out.
func script(replies ...string) func(int, llm.Request) (*llm.Response, error) {
	return func(n int, _ llm.Request) (*llm.Response, error) {
		return &llm.Response{Text: replies[min(n, len(replies))-1], FinishReason: "stop", Model: "m-1"}, nil
	}
}

// journalServer is a server on the data folder store with the model
// profile "m" (of the model named), answered by a new fake provider.
func journalServer(t *testing.T, store, model string, answer func(int, llm.Request) (*llm.Response, error)) (*Server, *httptest.Server, *provider) {
	t.Helper()
	srv, ts := newTestServerWith(t)
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	p := &provider{answer: answer}
	c := llm.WithProvider(llm.Profile{Name: "m", Provider: "mock", Retries: -1, Params: llm.Params{Model: model}}, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		n := int(p.n.Add(1))
		p.mu.Lock()
		p.reqs = append(p.reqs, req)
		p.mu.Unlock()
		return p.answer(n, req)
	}})
	c.Sleep = llm.NoWait
	srv.AddModel(c)
	return srv, ts, p
}

// debate is a debate of M (the model, pro, first to move) against the
// second seat, of four rounds, in which a pass concedes nothing.
func debate(second Seat) NewGameRequest {
	rounds, concede := 4, false
	return NewGameRequest{Mode: "debate", Seats: []Seat{{Name: "M", Kind: "model", Model: "m"}, second}, MaxRounds: &rounds, AllowOther: true,
		ConcedeOnPass: &concede, Seed: 7, Profile: game.DefaultProfile()}
}

func newGame(t *testing.T, ts *httptest.Server, req NewGameRequest) string {
	t.Helper()
	var st State
	if code := call(t, ts, "POST", "/api/games", req, &st); code != http.StatusOK {
		t.Fatalf("new game: %d", code)
	}
	return st.ID
}

func ai(t *testing.T, ts *httptest.Server, id string) (int, string) {
	t.Helper()
	var out map[string]any
	code := call(t, ts, "POST", "/api/games/"+id+"/ai", map[string]any{}, &out)
	msg, _ := out["error"].(string)
	return code, msg
}

// folder is the game's folder in the data folder, relative to it.
func folder(t *testing.T, store string) string {
	t.Helper()
	dirs, _ := filepath.Glob(filepath.Join(store, "games", "*"))
	if len(dirs) != 1 {
		t.Fatalf("the data folder holds %d games: %v", len(dirs), dirs)
	}
	return filepath.ToSlash(filepath.Join("games", filepath.Base(dirs[0])))
}

func resumeHTTP(t *testing.T, ts *httptest.Server, path string) (int, map[string]any) {
	t.Helper()
	var out map[string]any
	code := call(t, ts, "POST", "/api/store/resume", map[string]string{"path": path}, &out)
	return code, out
}

// resumed resumes the game of the folder in a new server on the store, and
// returns the server, its provider and the game's id.
func resumed(t *testing.T, store, model string, answer func(int, llm.Request) (*llm.Response, error)) (*Server, *httptest.Server, *provider, string) {
	t.Helper()
	srv, ts, p := journalServer(t, store, model, answer)
	code, st := resumeHTTP(t, ts, folder(t, store))
	if code != http.StatusOK {
		t.Fatalf("resume: %d %v", code, st["error"])
	}
	id, _ := st["id"].(string)
	return srv, ts, p, id
}

// events reads a trace file's events.
func events(t *testing.T, path string) []map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		var e map[string]any
		if err := json.Unmarshal([]byte(l), &e); err != nil {
			t.Fatal(err)
		}
		out = append(out, e)
	}
	return out
}

func ofType(evs []map[string]any, typ string) []map[string]any {
	var out []map[string]any
	for _, e := range evs {
		if e["type"] == typ {
			out = append(out, e)
		}
	}
	return out
}

func replyOf(call map[string]any) string {
	r, _ := call["response"].(map[string]any)
	s, _ := r["text"].(string)
	return s
}

func exportOfFolder(t *testing.T, dir string) savedGame {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "export.json"))
	if err != nil {
		t.Fatal(err)
	}
	var g savedGame
	if err := json.Unmarshal(b, &g); err != nil {
		t.Fatal(err)
	}
	return g
}

// stopAt stops the model's turn after the reply of the nth call, as the
// process stopping would: the data folder as it is then is copied, and its
// copy, returned, is what a server restarted on it finds (the server goes
// on, and saves its record of the stopped turn, which a stopped process
// would not).
func stopAt(t *testing.T, srv *Server, n int32, store string) string {
	t.Helper()
	snapshot := t.TempDir()
	var k atomic.Int32
	srv.stopAfterReply = func() bool {
		if k.Add(1) != n {
			return false
		}
		if err := os.CopyFS(snapshot, os.DirFS(store)); err != nil {
			t.Error(err)
		}
		return true
	}
	return snapshot
}

// G1, then G2: the reply is on the disk before its move, and a resumed game
// plays it without asking the provider.
func TestG1G2AReplyKeptIsPlayedAfterTheStop(t *testing.T) {
	store := t.TempDir()
	srv, ts, p := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	// A note before the stop, so that the export has an event to carry over.
	if code := call(t, ts, "POST", "/api/games/"+id+"/note", map[string]string{"target": "game", "kind": "note", "text": "before the stop", "by": "tester"}, nil); code != http.StatusOK {
		t.Fatalf("the note: %d", code)
	}
	stopped := stopAt(t, srv, 1, store)
	if code, msg := ai(t, ts, id); code == http.StatusOK || !strings.Contains(msg, "stopped after the model's reply") || p.calls() != 1 {
		t.Fatalf("the stopped turn: %d %q, %d calls", code, msg, p.calls())
	}
	store = stopped
	dir := filepath.Join(store, folder(t, store))
	calls := ofType(events(t, filepath.Join(dir, "trace.jsonl")), "call")
	if len(calls) != 1 || replyOf(calls[0]) != replyClaim {
		t.Fatalf("G1: the folder's calls: %v", calls)
	}
	if g := exportOfFolder(t, dir); len(g.History) != 0 {
		t.Fatalf("G1: the export has %d moves", len(g.History))
	}

	_, ts2, p2, id2 := resumed(t, store, "m-1", script(replyPass))
	if id2 != id {
		t.Fatalf("G2: the resumed game is %s, not %s", id2, id)
	}
	if code, msg := ai(t, ts2, id2); code != http.StatusOK || p2.calls() != 0 {
		t.Fatalf("G2: the model's turn: %d %q, %d calls", code, msg, p2.calls())
	}
	evs := events(t, filepath.Join(dir, "trace.jsonl"))
	calls, decisions := ofType(evs, "call"), ofType(evs, "decision")
	if len(calls) != 1 || len(decisions) != 1 || decisions[0]["recovered"] != true || decisions[0]["parent"] != calls[0]["id"] ||
		decisions[0]["reply"] != replyClaim || decisions[0]["move"] != float64(0) {
		t.Fatalf("G2: the trace: %d calls, decisions %v", len(calls), decisions)
	}
	if g := exportOfFolder(t, dir); len(g.History) != 1 || g.History[0].PlayerID != "M" || g.History[0].Action.SchemeID != "AS:consequences" {
		t.Fatalf("G2: the moves: %+v", g.History)
	}
	start := evs[0]
	if rs, _ := start["resumed"].([]any); len(rs) != 1 {
		t.Fatalf("G2: the start records %v resumes", start["resumed"])
	}
	for i := 1; i < len(evs); i++ {
		if evs[i]["order"].(float64) <= evs[i-1]["order"].(float64) {
			t.Fatalf("G2: the trace's events %d and %d are out of order: %v, %v", i-1, i, evs[i-1]["order"], evs[i]["order"])
		}
	}
	var what []string
	for _, e := range exportOfFolder(t, dir).Events {
		what = append(what, e.What)
	}
	if strings.Join(what, " ") != "note resume ai" {
		t.Fatalf("G2: the export's events: %v", what)
	}
}

// G2b: a game resumed from a folder of another name goes on saving into
// it, and makes no other.
func TestG2bAResumedGameSavesIntoItsFolder(t *testing.T) {
	store := t.TempDir()
	srv, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	store = stopAt(t, srv, 1, store)
	ai(t, ts, id)
	if err := os.Rename(filepath.Join(store, folder(t, store)), filepath.Join(store, "games", "renamed")); err != nil {
		t.Fatal(err)
	}
	_, ts2, _, id2 := resumed(t, store, "m-1", script(replyPass))
	if code, msg := ai(t, ts2, id2); code != http.StatusOK {
		t.Fatalf("the resumed turn: %d %q", code, msg)
	}
	if f := folder(t, store); f != "games/renamed" || len(exportOfFolder(t, filepath.Join(store, f)).History) != 1 {
		t.Fatalf("the game saves into %s", f)
	}
}

// G3 and G3b: a failed call, or a reply that came with an error, is not
// played: the resumed turn asks the provider once.
func TestG3AFailedCallIsAskedAgain(t *testing.T) {
	for name, first := range map[string]func() (*llm.Response, error){
		"G3, no reply": func() (*llm.Response, error) {
			return nil, &llm.Error{Provider: "mock", Kind: llm.KindAuth, Message: "the provider refused the key"}
		},
		"G3b, a reply and an error": func() (*llm.Response, error) {
			return &llm.Response{Text: replyClaim, Model: "m-1"}, &llm.Error{Provider: "mock", Kind: llm.KindAuth, Message: "the budget refused it"}
		},
	} {
		t.Run(name, func(t *testing.T) {
			store := t.TempDir()
			_, ts, p := journalServer(t, store, "m-1", func(n int, _ llm.Request) (*llm.Response, error) { return first() })
			id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
			if code, _ := ai(t, ts, id); code == http.StatusOK || p.calls() != 1 {
				t.Fatalf("the failing turn: %d, %d calls", code, p.calls())
			}
			_, ts2, p2, id2 := resumed(t, store, "m-1", script(replyClaim))
			if code, msg := ai(t, ts2, id2); code != http.StatusOK || p2.calls() != 1 {
				t.Fatalf("the resumed turn: %d %q, %d calls", code, msg, p2.calls())
			}
			evs := events(t, filepath.Join(store, folder(t, store), "trace.jsonl"))
			calls, decisions := ofType(evs, "call"), ofType(evs, "decision")
			if len(calls) != 2 || len(decisions) != 1 || decisions[0]["parent"] != calls[1]["id"] || decisions[0]["attempt"] != float64(1) ||
				decisions[0]["recovered"] == true {
				t.Fatalf("the trace: calls %v, decisions %v", calls, decisions)
			}
		})
	}
}

// G4 and G4b: a reply a decision used is not used again, nor a reply of
// another move, planted in the game's calls.
func TestG4AUsedReplyIsNotUsedAgain(t *testing.T) {
	for _, plant := range []bool{false, true} {
		store := t.TempDir()
		_, ts, p := journalServer(t, store, "m-1", script(replyClaim))
		id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
		if code, msg := ai(t, ts, id); code != http.StatusOK || p.calls() != 1 {
			t.Fatalf("M's move: %d %q", code, msg)
		}
		post(t, ts, "/api/games/"+id+"/move", map[string]any{"action": map[string]string{"type": "abstain"}})
		srv2, ts2, p2, id2 := resumed(t, store, "m-1", script(replyPass))
		planted := ""
		if plant { // G4b: an unused reply of M's first move, as the last call
			func() {
				srv2.mu.Lock()
				defer srv2.mu.Unlock()
				ss := srv2.sessions[id2]
				if len(ss.Calls) == 0 {
					t.Fatal("G4b: the resumed game has no call to plant again")
				}
				c := ss.Calls[0]
				c.ID, c.Order = "c"+itoa(len(ss.Calls)), ss.order+1
				ss.order++
				ss.Calls = append(ss.Calls, c)
				planted = c.ID
			}()
		}
		if code, msg := ai(t, ts2, id2); code != http.StatusOK || p2.calls() != 1 {
			t.Fatalf("planted %v: M's second move: %d %q, %d calls", plant, code, msg, p2.calls())
		}
		for _, d := range ofType(events(t, filepath.Join(store, folder(t, store), "trace.jsonl")), "decision") {
			if planted != "" && d["parent"] == planted {
				t.Fatalf("the planted reply is played: %v", d)
			}
		}
	}
}

// G5: a trace of another game is refused, saying which.
func TestG5ATraceOfAnotherGameIsRefused(t *testing.T) {
	store := t.TempDir()
	_, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	if code, _ := ai(t, ts, id); code != http.StatusOK {
		t.Fatal("M's move")
	}
	dir := filepath.Join(store, folder(t, store))
	trace, err := os.ReadFile(filepath.Join(dir, "trace.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]struct{ old, new, says string }{
		"another game":  {`"game":"` + id + `"`, `"game":"99"`, "the trace is of the game 99"},
		"other data":    {`"data_sha256":"`, `"data_sha256":"0`, "scheme data"},
		"another move":  {`\"G\":\"lower emissions\"`, `\"G\":\"higher emissions\"`, ""},
		"another reply": {`"reply":"{\"summary\":\"State the claim.\"`, `"reply":"{\"summary\":\"State it.\"`, "decision d0 is not the trace's"},
	} {
		edited := strings.ReplaceAll(string(trace), c.old, c.new)
		if edited == string(trace) {
			t.Fatalf("%s: nothing edited", name)
		}
		if err := os.WriteFile(filepath.Join(dir, "trace.jsonl"), []byte(edited), 0o644); err != nil {
			t.Fatal(err)
		}
		_, ts2, _ := journalServer(t, store, "m-1", script(replyPass))
		code, out := resumeHTTP(t, ts2, folder(t, store))
		msg, _ := out["error"].(string)
		if code == http.StatusOK || !strings.Contains(msg, "the trace is not the export's") || !strings.Contains(msg, c.says) {
			t.Errorf("%s: %d %q", name, code, msg)
		}
	}
}

// G5b: the recorded reply is played and refused; the next attempt asks the
// provider once, with the recorded conversation, the reply and the reason.
func TestG5bARecoveredReplyRefusedGoesOn(t *testing.T) {
	store := t.TempDir()
	srv, ts, _ := journalServer(t, store, "m-1", script(replyNoMove))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	store = stopAt(t, srv, 1, store)
	ai(t, ts, id)
	_, ts2, p2, id2 := resumed(t, store, "m-1", script(replyClaim))
	if code, msg := ai(t, ts2, id2); code != http.StatusOK || p2.calls() != 1 {
		t.Fatalf("the resumed turn: %d %q, %d calls", code, msg, p2.calls())
	}
	m := p2.reqs[0].Messages
	if len(m) != 3 || m[1].Content != replyNoMove || !strings.HasPrefix(m[2].Content, "That move was rejected: ") {
		t.Fatalf("the second attempt's conversation: %+v", m)
	}
	decisions := ofType(events(t, filepath.Join(store, folder(t, store), "trace.jsonl")), "decision")
	if len(decisions) != 2 || decisions[0]["recovered"] != true || decisions[1]["recovered"] == true || decisions[1]["attempt"] != float64(2) {
		t.Fatalf("the decisions: %v", decisions)
	}
}

// G5c: a server whose profile for the seat names another model refuses the
// resume, naming the profiles.
func TestG5cAnotherProfileRefusesTheResume(t *testing.T) {
	store := t.TempDir()
	srv, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	store = stopAt(t, srv, 1, store)
	ai(t, ts, id)
	_, ts2, _ := journalServer(t, store, "m-2", script(replyPass))
	code, out := resumeHTTP(t, ts2, folder(t, store))
	if msg, _ := out["error"].(string); code == http.StatusOK || !strings.Contains(msg, "profiles") {
		t.Fatalf("%d %q", code, msg)
	}
}

// G5d and G5e: after a reply refused by a rule (recorded in the history),
// the turn's second reply is played as the second attempt without asking
// the provider; unless the conversation it answered is not the turn's.
func TestG5dAReplyAfterARefusalByARule(t *testing.T) {
	for _, edit := range []bool{false, true} {
		store := t.TempDir()
		srv, ts, _ := journalServer(t, store, "m-1", script(replyRule, replyClaim))
		id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
		store = stopAt(t, srv, 2, store)
		ai(t, ts, id)
		dir := filepath.Join(store, folder(t, store))
		if g := exportOfFolder(t, dir); len(g.History) != 1 || g.History[0].Rejection == "" {
			t.Fatalf("the first reply is not refused by a rule, in the history: %+v", g.History)
		}
		if edit { // G5e: the second call answered another conversation
			b, _ := os.ReadFile(filepath.Join(dir, "trace.jsonl"))
			s := strings.Replace(string(b), "The game is as before.", "The game is not as before.", 1)
			if s == string(b) {
				t.Fatal("G5e: nothing edited")
			}
			os.WriteFile(filepath.Join(dir, "trace.jsonl"), []byte(s), 0o644)
		}
		_, ts2, p2, id2 := resumed(t, store, "m-1", script(replyClaim))
		want := map[bool]int{false: 0, true: 1}[edit]
		if code, msg := ai(t, ts2, id2); code != http.StatusOK || p2.calls() != want {
			t.Fatalf("edited %v: the resumed turn: %d %q, %d calls", edit, code, msg, p2.calls())
		}
		evs := events(t, filepath.Join(dir, "trace.jsonl"))
		calls, decisions := ofType(evs, "call"), ofType(evs, "decision")
		if len(calls) < 2 || len(decisions) == 0 {
			t.Fatalf("edited %v: the trace: %d calls, %d decisions", edit, len(calls), len(decisions))
		}
		last := decisions[len(decisions)-1]
		if last["attempt"] != float64(2) || last["move"] != float64(1) || (last["recovered"] == true) == edit ||
			(last["parent"] == calls[1]["id"]) == edit {
			t.Fatalf("edited %v: the second attempt's decision: %v", edit, last)
		}
		// The recovered decision's prompt is the board as the turn began, and
		// its feedback why the first reply was refused, as a live turn's.
		board := calls[0]["request"].(map[string]any)["messages"].([]any)[0].(map[string]any)["content"]
		why, _ := decisions[0]["result"].(map[string]any)["error"].(string)
		if !edit && (last["prompt"] != board || why == "" || last["feedback"] != rejectedFeedback(why)) {
			t.Fatalf("the recovered decision's prompt %.60q (want %.60q), feedback %q (want %q)", last["prompt"], board, last["feedback"], rejectedFeedback(why))
		}
	}
}

// G7: a save that fails right after the provider's answer holds the move;
// once a save succeeds, the reply kept in memory is played, with no new call.
func TestG7AReplyThatCannotBeSavedWaits(t *testing.T) {
	store := t.TempDir()
	_, ts, p := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	unblock := blockGames(t, store)
	if code, msg := ai(t, ts, id); code == http.StatusOK || !strings.Contains(msg, "the reply is played once the game is saved") || p.calls() != 1 {
		t.Fatalf("the unsaved turn: %d %q, %d calls", code, msg, p.calls())
	}
	if code, msg := ai(t, ts, id); code == http.StatusOK || !strings.Contains(msg, "model seats wait") || p.calls() != 1 {
		t.Fatalf("while unsaved: %d %q, %d calls", code, msg, p.calls())
	}
	unblock()
	if code, msg := ai(t, ts, id); code != http.StatusOK || p.calls() != 1 {
		t.Fatalf("once saved: %d %q, %d calls", code, msg, p.calls())
	}
	decisions := ofType(events(t, filepath.Join(store, folder(t, store), "trace.jsonl")), "decision")
	if len(decisions) != 1 || decisions[0]["recovered"] != true || decisions[0]["move"] != float64(0) {
		t.Fatalf("the decision: %v", decisions)
	}
}

// G8: a save that stopped between the trace and the export: the decision
// the export does not have is dropped, and its reply played again.
func TestG8ASaveStoppedHalfWay(t *testing.T) {
	store := t.TempDir()
	srv, ts, p := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	stopped := stopAt(t, srv, 1, store)
	ai(t, ts, id)
	dir := filepath.Join(store, folder(t, store))
	before, err := os.ReadFile(filepath.Join(stopped, folder(t, stopped), "export.json"))
	if err != nil {
		t.Fatal(err)
	}
	srv.stopAfterReply = nil
	if code, msg := ai(t, ts, id); code != http.StatusOK || p.calls() != 1 {
		t.Fatalf("the move played on: %d %q, %d calls", code, msg, p.calls())
	}
	played := exportOfFolder(t, dir).History
	if err := os.WriteFile(filepath.Join(dir, "export.json"), before, 0o644); err != nil {
		t.Fatal(err)
	}
	_, ts2, p2, id2 := resumed(t, store, "m-1", script(replyPass))
	evs := events(t, filepath.Join(dir, "trace.jsonl"))
	rs, _ := evs[0]["resumed"].([]any)
	if len(rs) != 1 || rs[0].(map[string]any)["dropped_decisions"] != float64(1) || len(ofType(evs, "decision")) != 0 {
		t.Fatalf("the resume: %v, %d decisions", evs[0]["resumed"], len(ofType(evs, "decision")))
	}
	if code, msg := ai(t, ts2, id2); code != http.StatusOK || p2.calls() != 0 {
		t.Fatalf("the resumed turn: %d %q, %d calls", code, msg, p2.calls())
	}
	evs = events(t, filepath.Join(dir, "trace.jsonl"))
	again := exportOfFolder(t, dir).History
	if len(ofType(evs, "call")) != 1 || len(ofType(evs, "decision")) != 1 || len(again) != len(played) ||
		!sameJSON(again[0].Action, played[0].Action) || !sameJSON(again[0].ArgIDs, played[0].ArgIDs) {
		t.Fatalf("after: %d calls, %d decisions, moves %+v, want %+v", len(ofType(evs, "call")), len(ofType(evs, "decision")), again, played)
	}
}

// playOut lets the computer and model seats play until the game is over.
func playOut(t *testing.T, ts *httptest.Server, id string) []game.Move {
	t.Helper()
	for i := 0; i < 200; i++ {
		var st State
		call(t, ts, "GET", "/api/games/"+id, nil, &st)
		if st.Over {
			var e savedGame
			b := export(t, ts, id)
			if err := json.Unmarshal(b, &e); err != nil {
				t.Fatal(err)
			}
			return e.History
		}
		if code, msg := ai(t, ts, id); code != http.StatusOK {
			t.Fatalf("move %d: %d %q", i, code, msg)
		}
	}
	t.Fatal("the game does not end")
	return nil
}

func export(t *testing.T, ts *httptest.Server, id string) []byte {
	t.Helper()
	resp, err := direct.Get(ts.URL + "/api/games/" + id + "/export")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var b strings.Builder
	buf := make([]byte, 1<<16)
	for {
		n, err := resp.Body.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return []byte(b.String())
}

// G9: a model seat against a random player, stopped after a model's reply
// in a later turn and resumed in a new server, makes the moves of the same
// game played without a stop.
func TestG9AComputerPlayerGoesOnAsWithoutTheStop(t *testing.T) {
	random := Seat{Name: "R", Kind: "random"}
	_, ts, _ := journalServer(t, t.TempDir(), "m-1", script(replyClaim, replyPass))
	whole := playOut(t, ts, newGame(t, ts, debate(random)))

	store := t.TempDir()
	srv, ts1, _ := journalServer(t, store, "m-1", script(replyClaim, replyPass))
	id := newGame(t, ts1, debate(random))
	snapshot := stopAt(t, srv, 2, store)
	stopped := false
	for i := 0; i < 10 && !stopped; i++ {
		code, msg := ai(t, ts1, id)
		stopped = code != http.StatusOK && strings.Contains(msg, "stopped after the model's reply")
	}
	store = snapshot
	at := len(exportOfFolder(t, filepath.Join(store, folder(t, store))).History)
	_, ts2, _, id2 := resumed(t, store, "m-1", script(replyPass))
	again := playOut(t, ts2, id2)
	if !stopped || len(whole) < 4 || len(again) != len(whole) {
		t.Fatalf("stopped %v; %d moves, then %d", stopped, len(whole), len(again))
	}
	after := 0
	for _, m := range whole[at:] {
		if m.PlayerID == "R" {
			after++
		}
	}
	if after == 0 {
		t.Fatalf("the random player makes no move after the stop (at move %d of %d)", at, len(whole))
	}
	for i := range whole {
		if whole[i].PlayerID != again[i].PlayerID || !sameJSON(whole[i].Action, again[i].Action) {
			t.Fatalf("move %d: %s %+v, then %s %+v", i, whole[i].PlayerID, whole[i].Action, again[i].PlayerID, again[i].Action)
		}
	}
}

// G10: a resumed game keeps its clocks: the move clock starts again, the
// game clock counts the time played before the stop; its trace's start is
// unchanged.
func TestG10AResumedGameKeepsItsClocks(t *testing.T) {
	store := t.TempDir()
	srv, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	req := debate(Seat{Name: "Ann", Kind: "human"})
	limit, move := 600, 60
	req.TimeLimitSec, req.MoveTimeSec = &limit, &move
	id := newGame(t, ts, req)
	store = stopAt(t, srv, 1, store)
	ai(t, ts, id)
	dir := filepath.Join(store, folder(t, store))
	startBefore := events(t, filepath.Join(dir, "trace.jsonl"))[0]
	g := exportOfFolder(t, dir)
	at := time.Now()
	srv2, _, _, id2 := resumed(t, store, "m-1", script(replyPass))
	srv2.mu.Lock()
	gs := srv2.sessions[id2].GS
	cfg, started, turn := gs.Config, gs.StartedAt, gs.TurnStartedAt
	srv2.mu.Unlock()
	srv.mu.Lock()
	played := g.ExportedAt.Sub(srv.sessions[id].Created)
	srv.mu.Unlock()
	if cfg.TimeLimit != 10*time.Minute || cfg.MoveTimeLimit != time.Minute || turn.Before(at) ||
		started.After(at.Add(-played).Add(time.Second)) || started.Before(at.Add(-played).Add(-time.Second)) {
		t.Fatalf("clocks %v %v; started %v (want about %v), the turn %v", cfg.TimeLimit, cfg.MoveTimeLimit, started, at.Add(-played), turn)
	}
	startAfter := events(t, filepath.Join(dir, "trace.jsonl"))[0]
	delete(startAfter, "resumed")
	if !sameJSON(startBefore, startAfter) {
		t.Fatalf("the trace's start changed:\n%v\n%v", startBefore, startAfter)
	}
}

// The resume takes a game's folder of the data folder, from this computer
// only, as the data folder's other actions do.
func TestResumeTakesOnlyAGameFolderFromThisComputer(t *testing.T) {
	store := t.TempDir()
	srv, ts, _ := journalServer(t, store, "m-1", script(replyClaim))
	id := newGame(t, ts, debate(Seat{Name: "Ann", Kind: "human"}))
	stopAt(t, srv, 1, store)
	ai(t, ts, id)
	os.MkdirAll(filepath.Join(store, "annotation", "run"), 0o755)
	for _, path := range []string{"games", "annotation/run", "games/../annotation/run", "../" + filepath.Base(store), "games/none"} {
		if code, out := resumeHTTP(t, ts, path); code == http.StatusOK {
			t.Errorf("%s: resumed: %v", path, out["id"])
		}
	}
	s2, _ := journalServerOnly(t, store)
	if code, _ := remote(t, s2, "POST", "/api/store/resume", "", map[string]string{"path": folder(t, store)}); code == http.StatusOK {
		t.Fatalf("another computer resumed the game: %d", code)
	}
	if code, out := resumeHTTP(t, ts, folder(t, store)); code == http.StatusOK || !strings.Contains(out["error"].(string), "is open in this server") {
		t.Fatalf("the game open in this server, resumed again: %d %v", code, out["error"])
	}
}

// journalServerOnly is a server on the data folder with the model profile
// "m", without an HTTP test server.
func journalServerOnly(t *testing.T, store string) (*Server, *provider) {
	t.Helper()
	srv, _, p := journalServer(t, store, "m-1", script(replyPass))
	return srv, p
}

// The trace reader takes only a trace the server writes: its format, a
// game's start first and once, and the four kinds of event.
func TestReadTraceTakesOnlyATrace(t *testing.T) {
	start := `{"format":"wag-trace/2","type":"game_start","id":"g","order":0,"game":"1"}`
	call := `{"format":"wag-trace/2","type":"call","id":"c0","order":1,"game":"1"}`
	for name, c := range map[string]struct {
		lines []string
		ok    bool
	}{
		"a trace":                {[]string{start, call}, true},
		"empty":                  {nil, false},
		"another format":         {[]string{strings.Replace(start, "wag-trace/2", "wag-trace/1", 1), call}, false},
		"an unknown event":       {[]string{start, `{"format":"wag-trace/2","type":"note","id":"n","order":2,"game":"1"}`}, false},
		"no start first":         {[]string{call, start}, false},
		"a second start":         {[]string{start, call, start}, false},
		"a line that is not one": {[]string{start, "{"}, false},
	} {
		_, err := readTrace([]byte(strings.Join(c.lines, "\n")))
		if (err == nil) != c.ok {
			t.Errorf("%s: %v", name, err)
		}
	}
}
