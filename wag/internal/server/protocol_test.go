package server

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"

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
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Seats: []Seat{{"Ann", "llm"}, {"Bob", "human"}}}, &st)
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
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Seats: []Seat{{"Ann", "claude"}, {"Bob", "human"}}}, nil); code != http.StatusBadRequest {
		t.Fatalf("a Claude seat on a server without Claude: %d", code)
	}
}

// A Claude seat asks the model, plays its move, and sends a rejected move
// back with the reason; its replies go into the export.
func TestClaudeSeat(t *testing.T) {
	srv, ts := newTestServerWith(t)
	calls := 0
	srv.SetClaude(&Claude{Model: "fake", Effort: "low", askFn: func(_ context.Context, msgs []anthropic.BetaMessageParam) (string, anthropic.BetaMessageParam, error) {
		calls++
		if calls == 1 {
			return `{"type":"support","scheme_id":"AS:nonexistent"}`, anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock("x")), nil
		}
		if len(msgs) != 3 {
			t.Errorf("the retry does not carry the rejection: %d messages", len(msgs))
		}
		return `Fine. {"type":"support","scheme_id":"AS:consequences","variable_bindings":{"A":"tax meat","G":"cleaner rivers"}}`, anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock("x")), nil
	}})
	var st State
	call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Seats: []Seat{{"Claude", "claude"}, {"Bob", "human"}}}, &st)
	if st.ActiveAI != "claude" || st.CanMove {
		t.Fatalf("Claude's seat: %+v", st)
	}
	mr := post(t, ts, "/api/games/"+st.ID+"/ai", map[string]interface{}{})
	if mr.Result["accepted"] != true || len(mr.State.Arguments) != 1 || calls != 2 {
		t.Fatalf("Claude's move: %+v after %d calls", mr.Result, calls)
	}
	var exp struct {
		LLM []LLMRecord `json:"llm"`
	}
	call(t, ts, "GET", "/api/games/"+st.ID+"/export", nil, &exp)
	if len(exp.LLM) != 1 || len(exp.LLM[0].Turns) != 2 || exp.LLM[0].Turns[0].Rejected == "" {
		t.Fatalf("export of Claude's replies: %+v", exp.LLM)
	}
}
