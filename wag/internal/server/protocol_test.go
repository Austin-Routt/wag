package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/wag/internal/llm"
	"github.com/yourorg/wag/internal/nascq"
)

// A reply may think aloud, quote examples and wrap the move in a code block:
// the last JSON object with a type is the move, under the usual field names
// or common variations of them.
func TestParseReply(t *testing.T) {
	for reply, want := range map[string]nascq.ProposedAction{
		`I could pass ({"type":"abstain"}), but I'll attack.` + "\n```json\n" +
			`{"type": "attack", "target": "a13", "cq": "field", "premise": "in_domain(x,y)", "scheme": "position_to_know", "slots": {"W": "a dating coach", "D": "dating"}, "text": "Dating is not energy policy {really}.", "sources": "my experience"}` + "\n```": {
			Type: "attack", TargetArgID: "a13", AnswersCQID: "CQ:field", TargetPremiseID: "in_domain(x,y)", SchemeID: "AS:position_to_know",
			VariableBindings: map[string]string{"W": "a dating coach", "D": "dating"}, Text: "Dating is not energy policy {really}.", Evidence: []string{"my experience"}},
		`{"type":"pass"}`: {Type: "abstain"},
		`Defending: {"type":"defend","target_arg_id":"a1","answers_cq_id":"CQ:possibility","scheme_id":"AS:expert_opinion","variable_bindings":{"W":"x","D":"y"}}`: {
			Type: "support", TargetArgID: "a1", AnswersCQID: "CQ:possibility", SchemeID: "AS:expert_opinion", VariableBindings: map[string]string{"W": "x", "D": "y"}},
	} {
		got, err := ParseReply(reply)
		if err != nil {
			t.Fatalf("%q: %v", reply, err)
		}
		if got.Type != want.Type || got.TargetArgID != want.TargetArgID || got.AnswersCQID != want.AnswersCQID || got.SchemeID != want.SchemeID ||
			got.TargetPremiseID != want.TargetPremiseID || got.Text != want.Text || len(got.VariableBindings) != len(want.VariableBindings) ||
			len(got.Evidence) != len(want.Evidence) {
			t.Errorf("%q\n got %+v\nwant %+v", reply, got, want)
		}
		for k, v := range want.VariableBindings {
			if got.VariableBindings[k] != v {
				t.Errorf("%q: %s = %q, want %q", reply, k, got.VariableBindings[k], v)
			}
		}
	}
	if _, err := ParseReply("I pass."); err == nil {
		t.Error("a reply without a move was read as one")
	}
}

// An LLM seat plays through the prompt and a pasted reply; a rejected reply
// says why.
func TestPromptAndReply(t *testing.T) {
	ts := newTestServer(t)
	var st State
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Seats: []Seat{{Name: "Ann", Kind: "llm"}, {Name: "Bob", Kind: "human"}}}, &st)
	if st.SeatKind != "llm" || !st.CanMove {
		t.Fatalf("an llm seat is a seat that moves through the API: %+v", st)
	}
	resp, err := http.Get(ts.URL + "/api/games/" + st.ID + "/prompt")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	prompt := string(b)
	for _, want := range []string{"You are Ann.", "You argue FOR the claim", "AS:practical_reasoning — Practical Reasoning", "also assumed:", "## Your reply"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt lacks %q", want)
		}
	}
	reply := `My opening. {"type":"support","scheme_id":"AS:practical_reasoning","variable_bindings":{"A":"tax meat","S1":"meat is cheap","S2":"less meat is eaten","G":"lower emissions","V":"a stable climate"},"text":"Tax meat to cut emissions."}`
	var mr struct {
		Result map[string]interface{} `json:"result"`
		State  State                  `json:"state"`
	}
	call(t, ts, "POST", "/api/games/"+st.ID+"/reply", map[string]string{"reply": reply}, &mr)
	if mr.Result["accepted"] != true || len(mr.State.Arguments) != 1 || mr.State.Arguments["a1"].Text != "Tax meat to cut emissions." {
		t.Fatalf("reply: %+v", mr.Result)
	}
	if code := call(t, ts, "POST", "/api/games/"+st.ID+"/reply", map[string]string{"reply": "I give up."}, nil); code != http.StatusBadRequest {
		t.Fatalf("a reply without a move: %d", code)
	}
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Seats: []Seat{{Name: "Ann", Kind: "claude"}, {Name: "Bob", Kind: "human"}}}, nil); code != http.StatusBadRequest {
		t.Fatalf("a Claude seat on a server without Claude: %d", code)
	}
}

// A model seat asks its model, plays its move, and sends a rejected move
// back with the reason, in the same conversation; its replies go into the
// export, and every call and decision into the trace file, with the
// conversation each call sent. A rate limit is retried and traced.
func TestModelSeat(t *testing.T) {
	srv, ts := newTestServerWith(t)
	mock := &llm.Mock{}
	calls := 0
	mock.Fn = func(req llm.Request) (*llm.Response, error) {
		calls++
		switch calls {
		case 1:
			return nil, &llm.Error{Provider: "mock", Kind: llm.KindRateLimit, Status: 429, RetryAfter: time.Millisecond}
		case 2:
			return &llm.Response{Text: `{"type":"support","scheme_id":"AS:nonexistent"}`, Model: "mock-1", FinishReason: "stop"}, nil
		}
		if len(req.Messages) != 3 || req.Messages[1].Role != "assistant" || !strings.HasPrefix(req.Messages[2].Content, "That move was rejected") {
			t.Errorf("the retry does not carry the conversation: %+v", req.Messages)
		}
		in := 900
		return &llm.Response{Text: `Fine. {"type":"support","scheme_id":"AS:consequences","variable_bindings":{"A":"tax meat","G":"cleaner rivers"}}`,
			Model: "mock-1", FinishReason: "stop", Usage: llm.Usage{InputTokens: &in}}, nil
	}
	srv.AddModel(llm.WithProvider(llm.Profile{Name: "test-model", Provider: "mock", Params: llm.Params{Model: "mock-1"}}, mock))
	var st State
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Seats: []Seat{{Name: "Pro", Kind: "model", Model: "test-model"}, {Name: "Bob", Kind: "human"}}}, &st)
	if st.ActiveAI != "model" || st.CanMove {
		t.Fatalf("the model's seat: %+v", st)
	}
	mr := post(t, ts, "/api/games/"+st.ID+"/ai", map[string]interface{}{})
	if mr.Result["accepted"] != true || len(mr.State.Arguments) != 1 || calls != 3 {
		t.Fatalf("the model's move: %+v after %d calls", mr.Result, calls)
	}
	var exp struct {
		LLM []LLMRecord `json:"llm"`
	}
	call(t, ts, "GET", "/api/games/"+st.ID+"/export", nil, &exp)
	if len(exp.LLM) != 1 || len(exp.LLM[0].Turns) != 2 || exp.LLM[0].Turns[0].Rejected == "" || exp.LLM[0].Profile != "test-model" {
		t.Fatalf("export of the model's replies: %+v", exp.LLM)
	}

	events := traceEvents(t, ts, st.ID)
	var types []string
	for _, e := range events {
		types = append(types, e["type"].(string)+":"+e["id"].(string))
	}
	if got := strings.Join(types, " "); got != "game_start:g call:c0 call:c1 decision:d0 call:c2 decision:d1" {
		t.Fatalf("events: %s", got)
	}
	start := events[0]
	if start["data_sha256"] == "" || start["protocol"] != ProtocolVersion || len(start["profiles"].([]any)) != 1 {
		t.Fatalf("game_start: %v", start)
	}
	c0, c1, d0, c2, d1 := events[1], events[2], events[3], events[4], events[5]
	if c0["error"].(map[string]any)["kind"] != "rate_limit" || c1["n"] != 2.0 || c1["error"] != nil {
		t.Fatalf("the retried call: %v / %v", c0, c1)
	}
	if d0["parent"] != "c1" || d1["parent"] != "c2" || d0["turn"] != d1["turn"] || d1["attempt"] != 2.0 {
		t.Fatalf("decisions: %v / %v", d0, d1)
	}
	if d0["prompt"] != d1["prompt"] || d0["prompt"] == "" || d1["feedback"] == "" || d1["served_model"] != "mock-1" {
		t.Fatal("every attempt keeps the board it started from; a retry says what it was told")
	}
	req := c2["request"].(map[string]any)
	if len(req["messages"].([]any)) != 3 || req["system"] != modelSystem || req["model"] != "mock-1" {
		t.Fatalf("the call's request: %v", req)
	}
	if c2["response"].(map[string]any)["usage"].(map[string]any)["input_tokens"] != 900.0 {
		t.Fatalf("usage: %v", c2["response"])
	}

	// The "claude" seat kind is the profile named claude.
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Seats: []Seat{{Name: "C", Kind: "claude"}, {Name: "B", Kind: "human"}}}, nil); code != http.StatusBadRequest {
		t.Fatalf("claude without the profile: %d", code)
	}
	srv.AddModel(llm.WithProvider(llm.Profile{Name: "claude", Provider: "mock", Params: llm.Params{Model: "m"}}, &llm.Mock{}))
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Seats: []Seat{{Name: "C", Kind: "claude"}, {Name: "B", Kind: "human"}}}, &st); code != http.StatusOK {
		t.Fatalf("claude with the profile: %d", code)
	}
	if mr := post(t, ts, "/api/games/"+st.ID+"/ai", map[string]interface{}{}); mr.Result["abstained"] != true {
		t.Fatalf("the mock passes: %+v", mr.Result)
	}
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Seats: []Seat{{Name: "C", Kind: "model", Model: "nope"}, {Name: "B", Kind: "human"}}}, nil); code != http.StatusBadRequest {
		t.Fatalf("an unknown profile: %d", code)
	}
}

// traceEvents reads a game's trace file.
func traceEvents(t *testing.T, ts *httptest.Server, id string) []map[string]any {
	t.Helper()
	resp, err := http.Get(ts.URL + "/api/games/" + id + "/traces")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out []map[string]any
	dec := json.NewDecoder(resp.Body)
	for dec.More() {
		var e map[string]any
		if err := dec.Decode(&e); err != nil {
			t.Fatal(err)
		}
		if e["format"] != TraceFormat {
			t.Fatalf("format %v", e["format"])
		}
		out = append(out, e)
	}
	return out
}

// Protocol 2: the reply is an envelope around the move, statements can be
// named by reference, and each decision is traced with its predictions
// checked and its alternatives tried.
func TestReplyEnvelopeAndTraces(t *testing.T) {
	ts := newTestServer(t)
	var st State
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Seats: []Seat{{Name: "Ann", Kind: "llm"}, {Name: "Bob", Kind: "llm"}}}, &st)
	play := func(reply string) map[string]interface{} {
		t.Helper()
		var mr struct {
			Result map[string]interface{} `json:"result"`
		}
		call(t, ts, "POST", "/api/games/"+st.ID+"/reply", map[string]string{"reply": reply}, &mr)
		return mr.Result
	}
	r := play(`{"action":{"type":"support","scheme_id":"AS:practical_reasoning","variable_bindings":{"A":"tax meat","S1":"meat is cheap","S2":"less meat is eaten","G":"lower emissions","V":"a stable climate"}},"summary":"Opening."}`)
	if r["accepted"] != true {
		t.Fatalf("opening: %+v", r)
	}
	resp, err := http.Get(ts.URL + "/api/games/" + st.ID + "/prompt")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	prompt := string(b)
	for _, want := range []string{"wag-protocol/2", "[s1]", "### For or against a statement", "### Challenges", "linked: AS:sign (causality)", "thesis-independent scheme: AS:expert_opinion", "attack argues:", `(an attack answers "no", a defence "yes")`, `"thoughts"`} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt lacks %q", want)
		}
	}
	// s1 is the claim; s2 is the first premise of a1. Bob argues against
	// it, predicts the claim goes out, and names an alternative.
	r = play("Thinking...\n```json\n" + `{"context":["a1 rests on the circumstances"],"thoughts":["if the circumstances fall, a1 falls"],
		"summary":"Undermine the circumstances.","expect":[{"stmt":"claim","label":"out"},{"stmt":"s2","label":"in"}],
		"alternatives":[{"action":{"type":"question","target_arg_id":"a1","answers_cq_id":"CQ:premise_acceptable","target_premise_id":"s2"},"why_not":"a question is weaker"}],
		"confidence":0.7,
		"action":{"type":"against","statement":"s2","scheme":"position_to_know","bindings":{"W":"a butcher","D":"meat prices"}}}` + "\n```")
	if r["accepted"] != true {
		t.Fatalf("argument against s2: %+v", r)
	}
	var traces []Trace
	resp, err = http.Get(ts.URL + "/api/games/" + st.ID + "/traces")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(func() string { b, _ := io.ReadAll(resp.Body); return string(b) }()), "\n") {
		var tr Trace
		if err := json.Unmarshal([]byte(line), &tr); err != nil {
			t.Fatal(err)
		}
		if tr.Type == "decision" {
			traces = append(traces, tr)
		}
	}
	if len(traces) != 2 {
		t.Fatalf("%d traces", len(traces))
	}
	tr := traces[1]
	if tr.Player != "Bob" || tr.Protocol != ProtocolVersion || tr.PromptSHA256 == "" || tr.Move != 1 || tr.Reasoning == nil ||
		tr.Reasoning.Summary != "Undermine the circumstances." || *tr.Reasoning.Confidence != 0.7 {
		t.Fatalf("trace: %+v", tr)
	}
	if !strings.HasPrefix(tr.Action.TargetStmtID, "current_circumstances(") || tr.Action.Type != "attack" {
		t.Fatalf("the reference was not resolved: %+v", tr.Action)
	}
	if len(tr.ExpectCheck) != 2 || !tr.ExpectCheck[0].OK || tr.ExpectCheck[1].OK || tr.ExpectCheck[1].Actual != "out" {
		t.Fatalf("expectations: %+v", tr.ExpectCheck)
	}
	if len(tr.AlternativeOutcomes) != 1 || !tr.AlternativeOutcomes[0].Legal || tr.ClaimBefore != "in" || tr.ClaimAfter != "out" {
		t.Fatalf("alternatives %+v, claim %s → %s", tr.AlternativeOutcomes, tr.ClaimBefore, tr.ClaimAfter)
	}
}
