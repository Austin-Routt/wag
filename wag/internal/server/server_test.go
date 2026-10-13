package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/nascq"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	_, ts := newTestServerWith(t)
	return ts
}

func newTestServerWith(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	s, err := New(Data{YAML: data.CoreYAML, Schemes: data.SchemesJSON, CQs: data.CQsJSON, Edges: data.EdgesJSON, Contexts: data.ContextsJSON}, "")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return s, ts
}

func call(t *testing.T, ts *httptest.Server, method, path string, body interface{}, out interface{}) int {
	t.Helper()
	var r io.Reader
	if body != nil {
		if s, ok := body.(string); ok {
			r = strings.NewReader(s)
		} else {
			b, _ := json.Marshal(body)
			r = bytes.NewReader(b)
		}
	}
	req, _ := http.NewRequest(method, ts.URL+path, r)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("%s %s: %v", method, path, err)
		}
	}
	return resp.StatusCode
}

// The interface is served, the metadata lists the schemes with their links.
func TestStaticAndMeta(t *testing.T) {
	ts := newTestServer(t)
	for _, p := range []string{"/", "/app.js", "/styles.css"} {
		resp, err := http.Get(ts.URL + p)
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("GET %s: %v %v", p, err, resp.Status)
		}
	}
	var meta struct {
		Schemes []SchemeMeta
		CQs     []CQMeta `json:"cqs"`
	}
	call(t, ts, "GET", "/api/meta", nil, &meta)
	if len(meta.Schemes) != 10 || len(meta.CQs) != 27 { // 9 Walton schemes and Other; 24 questions, the challenge and Other's two
		t.Fatalf("%d schemes, %d questions", len(meta.Schemes), len(meta.CQs))
	}
	for _, s := range meta.Schemes {
		if s.ID == "AS:expert_opinion" && (len(s.Asked) != 6 || s.Carneades == nil || s.Carneades.ID != "expert_opinion") {
			t.Fatalf("expert opinion meta: %+v", s)
		}
	}
}

// A game through the API: an argument, a preview, the move options of an
// argument, a question, a defence, the AI's move, and the export.
func TestGameFlow(t *testing.T) {
	ts := newTestServer(t)
	var st State
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "cq", Seats: []Seat{{"Alice", "human"}, {"Bob", "human"}}, AllowOther: true}, &st)
	if st.ID == "" || st.Active != "Alice" || !st.CanMove {
		t.Fatalf("new game: %+v", st)
	}
	pr := nascq.ProposedAction{Type: "support", SchemeID: "AS:practical_reasoning", Text: "Tax carbon to cut emissions.",
		VariableBindings: map[string]string{"A": "carbon tax", "S1": "status quo", "S2": "lower emissions", "G": "clean air", "V": "health"}}
	var p Preview
	call(t, ts, "POST", "/api/games/"+st.ID+"/preview", map[string]interface{}{"action": pr}, &p)
	if !p.OK || len(p.Premises) != 7 || len(p.Changes) == 0 {
		t.Fatalf("preview: %+v", p)
	}
	for _, x := range append(append([]string{p.Conclusion}, p.Premises...), p.Changes[0].Text) {
		if strings.ContainsAny(x, "{}_") || x == "" {
			t.Fatalf("preview reads %q", x)
		}
	}
	var mr moveReply
	mr = post(t, ts, "/api/games/"+st.ID+"/move", map[string]interface{}{"action": pr})
	if mr.Result["accepted"] != true || len(mr.State.Arguments) != 1 || mr.State.Arguments["a1"].Text != pr.Text {
		t.Fatalf("move: %+v", mr.Result)
	}
	var opts []MoveOption
	call(t, ts, "GET", "/api/games/"+st.ID+"/moves?target=a1", nil, &opts)
	// Its five critical questions, and the generic premise question about
	// each of the six premises none of them is about (not possible(A),
	// which the Possibility Question is about).
	generic := map[string]bool{}
	for _, o := range opts {
		if o.CQ == nascq.CQPremiseAcceptable {
			generic[o.Stmt] = true
		}
	}
	if len(opts) != 11 || len(generic) != 6 || generic["possible(carbon_tax)"] || !generic["would_bring_about(carbon_tax,status_quo,lower_emissions)"] {
		t.Fatalf("%d move options on a1, the generic question about %v", len(opts), generic)
	}
	q := nascq.ProposedAction{Type: "question", AnswersCQID: "CQ:possibility", TargetArgID: "a1"}
	mr = post(t, ts, "/api/games/"+st.ID+"/move", map[string]interface{}{"action": q})
	if len(mr.State.Questions) != 1 || mr.State.Statements["possible(carbon_tax)"].Label != "out" {
		t.Fatalf("question: %+v", mr.State.Questions)
	}
	cq := mr.State.Arguments["a1"].CQs
	found := false
	for _, c := range cq {
		found = found || (c.CQ == "CQ:possibility" && c.Open && c.AskedBy == "Bob")
	}
	if !found {
		t.Fatalf("the open question is not on a1's card: %+v", cq)
	}
	d := nascq.ProposedAction{Type: "support", AnswersCQID: "CQ:possibility", TargetArgID: "a1", SchemeID: nascq.OtherSchemeID,
		VariableBindings: map[string]string{"P1": "Parliament has a majority for it."}}
	mr = post(t, ts, "/api/games/"+st.ID+"/move", map[string]interface{}{"action": d})
	if mr.Result["accepted"] != true || mr.State.Statements["possible(carbon_tax)"].Label != "in" || len(mr.State.Questions) != 0 {
		t.Fatalf("defence: %+v", mr.Result)
	}
	if code := call(t, ts, "POST", "/api/games/"+st.ID+"/ai", map[string]interface{}{}, nil); code != http.StatusConflict {
		t.Fatalf("AI move on a human seat: %d", code)
	}
	var ex map[string]interface{}
	call(t, ts, "GET", "/api/games/"+st.ID+"/export", nil, &ex)
	if mined, _ := ex["other_arguments"].([]interface{}); len(mined) != 1 {
		t.Fatalf("export: %v other arguments", ex["other_arguments"])
	}
	var un map[string]interface{}
	call(t, ts, "GET", "/api/games/"+st.ID+"?since="+itoa(mr.State.Version), nil, &un)
	if un["unchanged"] != true {
		t.Fatalf("poll: %v", un)
	}
}

// moveReply is the answer to a move; post decodes each into a fresh one.
type moveReply struct {
	Result map[string]interface{}
	State  State
}

func post(t *testing.T, ts *httptest.Server, path string, body interface{}) moveReply {
	t.Helper()
	var r moveReply
	call(t, ts, "POST", path, body, &r)
	return r
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }

// An AI seat moves on request; a moderated game holds the move.
func TestAIAndModeration(t *testing.T) {
	ts := newTestServer(t)
	var st State
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Seats: []Seat{{"Bot", "greedy"}, {"Ann", "human"}}, Moderated: true, Seed: 1}, &st)
	if st.ActiveAI != "greedy" || st.CanMove {
		t.Fatalf("AI seat: %+v", st)
	}
	var mr moveReply
	mr = post(t, ts, "/api/games/"+st.ID+"/ai", map[string]interface{}{})
	if mr.Result["pending"] != true || mr.State.Pending == nil || mr.State.Pending.Description == "" {
		t.Fatalf("held AI move: %+v %+v", mr.Result, mr.State.Pending)
	}
	mr = post(t, ts, "/api/games/"+st.ID+"/moderate", map[string]interface{}{"approve": true, "weight": 0.4, "note": "ok"})
	if mr.Result["accepted"] != true || mr.State.Active != "Ann" || mr.State.Pending != nil {
		t.Fatalf("approved: %+v", mr.Result)
	}
	var a *ArgState
	for _, x := range mr.State.Arguments {
		a = x
	}
	if a == nil || a.BaseWeight != 0.4 || !strings.HasPrefix(a.Moderation, "approved") {
		t.Fatalf("approved argument: %+v", a)
	}
	stmt := a.Conclusion
	mr = post(t, ts, "/api/games/"+st.ID+"/decide", map[string]interface{}{"statement": stmt, "decision": "rejected"})
	if mr.State.Statements[stmt].Label != "out" || mr.State.Statements[stmt].Decision != "rejected" {
		t.Fatalf("decision: %+v", mr.State.Statements[stmt])
	}
}

// A Kialo graph.json posted to the import endpoint becomes a game.
func TestKialoImportEndpoint(t *testing.T) {
	ts := newTestServer(t)
	body := `{"id":"9","displayTitle":"Q?","thesis_claim_id":"9.1","claims":{"9.1":{"id":"9.1","text":"Yes."},"9.2":{"id":"9.2","text":"Because."}},
	 "edges":[{"location_id":"9.3","from_claim_id":"9.1","to_claim_id":"9.2","relation":"support","vote_vector":[0,0,0,0,1],"active":true}]}`
	var r struct {
		State  State
		Report map[string]interface{}
	}
	if code := call(t, ts, "POST", "/api/kialo/import", body, &r); code != 200 {
		t.Fatalf("import: %d", code)
	}
	if r.State.Verdict != "In" || len(r.State.Arguments) != 1 || r.Report["arguments"].(float64) != 1 {
		t.Fatalf("imported: verdict %s, %d arguments, %v", r.State.Verdict, len(r.State.Arguments), r.Report)
	}
	if code := call(t, ts, "POST", "/api/kialo/import", `{"nope":1}`, nil); code != 400 {
		t.Fatalf("bad import: %d", code)
	}
}

// A request that changes a game must be a JSON POST, which another web page
// cannot send to this server without a preflight it never grants.
func TestCrossSiteRequestsRefused(t *testing.T) {
	ts := newTestServer(t)
	var st State
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Seats: []Seat{{"Bot", "greedy"}, {"Ann", "human"}}, Seed: 1}, &st)
	if code := call(t, ts, "GET", "/api/games/"+st.ID+"/ai", nil, nil); code != http.StatusMethodNotAllowed {
		t.Fatalf("GET /ai: %d", code)
	}
	for _, path := range []string{"/api/games", "/api/games/" + st.ID + "/ai", "/api/games/" + st.ID + "/move", "/api/kialo/import"} {
		resp, err := http.Post(ts.URL+path, "text/plain", strings.NewReader(`{}`))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnsupportedMediaType {
			t.Fatalf("text/plain POST %s: %d", path, resp.StatusCode)
		}
	}
	var after State
	call(t, ts, "GET", "/api/games/"+st.ID, nil, &after)
	if len(after.Log) != 0 {
		t.Fatalf("a refused request changed the game: %+v", after.Log)
	}
}

// A game exported by a tester loads, replays move by move, and ends with the
// labels it had when it was exported.
func TestLoadReplaysAnExport(t *testing.T) {
	ts := newTestServer(t)
	b, err := os.ReadFile("testdata/user-game-4.json")
	if err != nil {
		t.Fatal(err)
	}
	var exp struct {
		State State `json:"state"`
	}
	if err := json.Unmarshal(b, &exp); err != nil {
		t.Fatal(err)
	}
	var st State
	if code := call(t, ts, "POST", "/api/load", string(b), &st); code != 200 {
		t.Fatalf("load: %d", code)
	}
	if len(st.Log) != len(exp.State.Log) || st.Round != exp.State.Round || st.Active != exp.State.Active {
		t.Fatalf("replayed %d moves to round %d (%s to move), exported %d to round %d (%s)",
			len(st.Log), st.Round, st.Active, len(exp.State.Log), exp.State.Round, exp.State.Active)
	}
	// The game was recorded when the game's expert opinion was Wyner's
	// credible_source; it is walton.yml's expert_opinion since 2026-10-13,
	// so the statements only that scheme made are not replayed (they were
	// unargued premises and exceptions; no label depends on them).
	retired := map[string]bool{"credible_source": true, "biased": true, "dishonest": true, "other_credible_sources_disagree": true}
	for id, s := range exp.State.Statements {
		got := st.Statements[id]
		if pred, _, _ := strings.Cut(strings.TrimPrefix(id, "¬"), "("); got == nil && retired[pred] {
			continue
		}
		if got == nil || got.Label != s.Label {
			t.Errorf("%s: replayed %+v, exported %s", id, got, s.Label)
		}
	}
	for i, p := range exp.State.Players {
		if st.Players[i].Score != p.Score {
			t.Errorf("%s scores %v, exported %v", p.ID, st.Players[i].Score, p.Score)
		}
	}
	// The replayed game exports what it needs to be loaded again.
	var again map[string]interface{}
	call(t, ts, "GET", "/api/games/"+st.ID+"/export", nil, &again)
	if again["request"] == nil || again["build"] == nil || again["format"] != "wag-export/2" {
		t.Fatalf("export lacks request, build or format: %v %v %v", again["request"], again["build"], again["format"])
	}
}

// Notes attach to a game and come back in its export.
func TestNotes(t *testing.T) {
	ts := newTestServer(t)
	var st State
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Seats: []Seat{{"Ann", "human"}, {"Bob", "human"}}}, &st)
	if code := call(t, ts, "POST", "/api/games/"+st.ID+"/note", map[string]string{"target": "move:1", "kind": "confusing", "text": "What does this mean?", "by": "Ann"}, nil); code != 200 {
		t.Fatalf("note: %d", code)
	}
	if code := call(t, ts, "POST", "/api/games/"+st.ID+"/note", map[string]string{"kind": "bogus", "text": "x"}, nil); code != 400 {
		t.Fatalf("bad note kind: %d", code)
	}
	var exp struct {
		Notes []Note `json:"notes"`
	}
	call(t, ts, "GET", "/api/games/"+st.ID+"/export", nil, &exp)
	if len(exp.Notes) != 1 || exp.Notes[0].Kind != "confusing" || exp.Notes[0].Target != "move:1" {
		t.Fatalf("notes: %+v", exp.Notes)
	}
}

// A context preset weighs the schemes as its audience would.
func TestContext(t *testing.T) {
	ts := newTestServer(t)
	var meta struct {
		Contexts []struct{ ID string } `json:"contexts"`
	}
	call(t, ts, "GET", "/api/meta", nil, &meta)
	if len(meta.Contexts) < 4 {
		t.Fatalf("contexts: %+v", meta.Contexts)
	}
	var st State
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Context: "scientific", Standard: "CCE",
		Seats: []Seat{{"Ann", "human"}, {"Bob", "human"}}}, &st)
	if st.Config.Context != "scientific" {
		t.Fatalf("config: %+v", st.Config)
	}
	mr := post(t, ts, "/api/games/"+st.ID+"/move", map[string]interface{}{"action": nascq.ProposedAction{Type: "support", SchemeID: "AS:analogy",
		VariableBindings: map[string]string{"C": "Denmark's fat tax", "A": "food taxes change diets"}}})
	if a := mr.State.Arguments["a1"]; a == nil || a.BaseWeight != 0.45 {
		t.Fatalf("an analogy in a scientific inquiry: %+v", a)
	}
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Context: "nonsense"}, nil); code != http.StatusBadRequest {
		t.Fatalf("unknown context: %d", code)
	}
}

// A game can set its own scheme weights on top of a context's; weights for
// unknown schemes, or out of (0, 1], are refused; when the game ends, the
// state says how the claim would stand under each preset's weights.
func TestCustomWeights(t *testing.T) {
	ts := newTestServer(t)
	var st State
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Context: "policy", Seats: []Seat{{"Ann", "human"}, {"Bob", "human"}},
		Weights: map[string]float64{"AS:sign": 0.9}}, &st)
	if st.Config.Weights["AS:sign"] != 0.9 || st.Config.Weights["AS:practical_reasoning"] != 0.7 {
		t.Fatalf("weights: %v", st.Config.Weights)
	}
	for _, w := range []map[string]float64{{"AS:nope": 0.5}, {"AS:sign": 1.5}, {"AS:sign": 0}} {
		if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Weights: w}, nil); code != http.StatusBadRequest {
			t.Errorf("weights %v: %d", w, code)
		}
	}
	mr := post(t, ts, "/api/games/"+st.ID+"/move", map[string]interface{}{"action": map[string]interface{}{"type": "support", "scheme_id": "AS:consequences",
		"variable_bindings": map[string]string{"A": "tax meat", "G": "cleaner rivers"}}})
	if mr.Result["accepted"] != true {
		t.Fatalf("opening: %+v", mr.Result)
	}
	mr = post(t, ts, "/api/games/"+st.ID+"/move", map[string]interface{}{"action": map[string]interface{}{"type": "resign"}})
	if !mr.State.Over || len(mr.State.Verdicts) < 2 || mr.State.Verdicts[0].Context != "this game" {
		t.Fatalf("verdicts: %+v", mr.State.Verdicts)
	}
}
