package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/game"
)

// E1.0 (docs/e1-evidence.md): the cooperative game, mode inquiry. Each
// test's outcome is the plan's, written before the tests ran.

// The opening argument of the tests: practical reasoning for "tax meat".
var claimAction = json.RawMessage(`{"type":"support","scheme_id":"AS:consequences","variable_bindings":{"A":"tax meat","G":"lower emissions"},"declare":[{"id":"tax meat","actor":"decider"}]}`)

// An undercutter of a1, answering its Discounted Gain Question.
var undercutAction = json.RawMessage(`{"type":"attack","attack_type":"undercutter","target_arg_id":"a1","scheme_id":"AS:expert_opinion","answers_cq_id":"CQ:gain_discounted","variable_bindings":{"D":"the subject at hand","W":"a reviewer"}}`)

var passAction = json.RawMessage(`{"type":"abstain"}`)

func inquiry(seats int, rounds *int) NewGameRequest {
	req := NewGameRequest{Mode: "inquiry", AllowOther: true, Seed: 7, Profile: game.DefaultProfile(), MaxRounds: rounds}
	for _, name := range []string{"A", "B", "C"}[:seats] {
		req.Seats = append(req.Seats, Seat{Name: name, Kind: "human"})
	}
	return req
}

// play makes the player to move play a; it fails the test if the move is
// not accepted.
func play(t *testing.T, call func(path string, body any) moveReply, id string, a json.RawMessage) moveReply {
	t.Helper()
	mr := call("/api/games/"+id+"/move", map[string]any{"action": a})
	if mr.Result["accepted"] != true {
		t.Fatalf("the move %s is not accepted: %v", a, mr.Result)
	}
	return mr
}

// C1: an inquiry of two seats: no sides; the outcome has the verdict and no
// winner, before the end and after it.
// C3: A argues and B passes; then both pass in the next round: over, the
// result the claim's label.
func TestC1C3AnInquiryHasNoSidesAndEndsWithTheVerdict(t *testing.T) {
	srv, ts := newTestServerWith(t)
	var st State
	if code := call(t, ts, "POST", "/api/games", inquiry(2, nil), &st); code != http.StatusOK {
		t.Fatalf("new inquiry: %d", code)
	}
	poster := func(path string, body any) moveReply { return post(t, ts, path, body) }
	for _, p := range st.Players {
		if p.Side != "" {
			t.Fatalf("C1: %s has the side %q", p.ID, p.Side)
		}
	}
	if mr := play(t, poster, st.ID, claimAction); mr.State.Outcome != nil || mr.State.Verdict == "" {
		t.Fatalf("C1: the board before the end: outcome %+v, verdict %q", mr.State.Outcome, mr.State.Verdict)
	}
	// Before the end the board shows the verdict and no outcome; the game's
	// own outcome is provisional (what the result would be if it ended now).
	srv.mu.Lock()
	o := srv.sessions[st.ID].GS.Outcome()
	srv.mu.Unlock()
	if o.Winner != "" || o.WinningSide != "" || o.Verdict == "" {
		t.Fatalf("C1: the outcome before the end: %+v", o)
	}
	play(t, poster, st.ID, passAction) // B
	if mr := play(t, poster, st.ID, passAction); mr.State.Over {
		t.Fatal("C3: over after A's pass alone")
	}
	mr := play(t, poster, st.ID, passAction) // B: both passed this round
	o = mr.State.Outcome
	if !mr.State.Over || o == nil || o.Winner != "" || o.WinningSide != "" || o.Draw || o.Verdict == "" || o.Claim == "" {
		t.Fatalf("C3: over %v, outcome %+v", mr.State.Over, o)
	}
}

// C2: an inquiry is two players.
func TestC2AnInquiryIsTwoPlayers(t *testing.T) {
	ts := newTestServer(t)
	for _, n := range []int{1, 3} {
		var out map[string]any
		code := call(t, ts, "POST", "/api/games", inquiry(n, nil), &out)
		if msg, _ := out["error"].(string); code != http.StatusBadRequest || !strings.Contains(msg, "an inquiry is two players") {
			t.Errorf("an inquiry of %d seats: %d %v", n, code, out["error"])
		}
	}
}

// C4: a limit of one round; A argues and B attacks: over at the limit, the
// verdict as result, no winner.
func TestC4AnInquiryEndsAtItsRoundLimit(t *testing.T) {
	ts := newTestServer(t)
	one := 1
	var st State
	if code := call(t, ts, "POST", "/api/games", inquiry(2, &one), &st); code != http.StatusOK {
		t.Fatalf("new inquiry: %d", code)
	}
	poster := func(path string, body any) moveReply { return post(t, ts, path, body) }
	play(t, poster, st.ID, claimAction)
	mr := play(t, poster, st.ID, undercutAction)
	o := mr.State.Outcome
	if !mr.State.Over || o == nil || o.Winner != "" || o.WinningSide != "" || o.Verdict == "" {
		t.Fatalf("C4: over %v, outcome %+v", mr.State.Over, o)
	}
}

// C5: A argues; B resigns: over, the verdict as result, no winner and no
// loser.
func TestC5AResignationEndsTheInquiry(t *testing.T) {
	ts := newTestServer(t)
	var st State
	if code := call(t, ts, "POST", "/api/games", inquiry(2, nil), &st); code != http.StatusOK {
		t.Fatalf("new inquiry: %d", code)
	}
	poster := func(path string, body any) moveReply { return post(t, ts, path, body) }
	play(t, poster, st.ID, claimAction)
	mr := play(t, poster, st.ID, json.RawMessage(`{"type":"resign"}`))
	o := mr.State.Outcome
	if !mr.State.Over || o == nil || o.Winner != "" || o.Loser != "" || o.WinningSide != "" || o.Verdict == "" {
		t.Fatalf("C5: over %v, outcome %+v", mr.State.Over, o)
	}
}

// C6: an inquiry requested with a setting that makes a winner or a loser is
// refused, naming it.
func TestC6AnInquiryRefusesWinnersAndLosers(t *testing.T) {
	ts := newTestServer(t)
	fifty, one, yes := 50.0, 1, true
	for name, c := range map[string]struct {
		edit func(*NewGameRequest)
		says string
	}{
		"a target score":      {func(r *NewGameRequest) { r.TargetScore = &fifty }, "target score"},
		"an answer deadline":  {func(r *NewGameRequest) { r.AnswerWithin = &one }, "answer deadline"},
		"conceding on a pass": {func(r *NewGameRequest) { r.ConcedeOnPass = &yes }, "concedes"},
	} {
		req := inquiry(2, nil)
		c.edit(&req)
		var out map[string]any
		code := call(t, ts, "POST", "/api/games", req, &out)
		if msg, _ := out["error"].(string); code != http.StatusBadRequest || !strings.Contains(msg, c.says) {
			t.Errorf("C6, %s: %d %q", name, code, msg)
		}
	}
}

// C7: a model seat's turn in an inquiry: its prompt and system prompt name
// no side and no winner, and say the players examine the claim together.
// C8: in a debate, the system prompt is the one before E1, byte for byte.
func TestC7C8AModelSeatsPromptsInAnInquiryAndADebate(t *testing.T) {
	_, ts, p := journalServer(t, t.TempDir(), "m-1", script(replyPass))
	req := inquiry(2, nil)
	req.Seats[0] = Seat{Name: "M", Kind: "model", Model: "m"}
	id := newGame(t, ts, req)
	if code, msg := ai(t, ts, id); code != http.StatusOK {
		t.Fatalf("the model's turn: %d %q", code, msg)
	}
	r := requests(p)[0]
	prompt := r.Messages[0].Content
	for _, s := range []string{r.System, prompt} {
		if !strings.Contains(s, "together") || strings.Contains(s, "help your side") || strings.Contains(s, "You argue FOR") ||
			strings.Contains(s, "You argue AGAINST") || strings.Contains(s, "pro wins") || strings.Contains(s, "for points") ||
			strings.Contains(s, "your side") || strings.Contains(s, "other side may argue") || strings.Contains(s, "In a debate") {
			t.Fatalf("C7: an inquiry's prompt or system prompt names a side or a winner:\n%s", s)
		}
	}
	// The seat's role, said where a debate's prompt gives the side.
	if !strings.Contains(prompt, "You and the other player examine the claim together") {
		t.Fatalf("C7: the prompt does not give the seat's role in an inquiry:\n%s", prompt)
	}
	_, ts2, p2 := journalServer(t, t.TempDir(), "m-1", script(replyPass))
	id2 := newGame(t, ts2, debate(Seat{Name: "Ann", Kind: "human"}))
	if code, msg := ai(t, ts2, id2); code != http.StatusOK {
		t.Fatalf("the debate's model turn: %d %q", code, msg)
	}
	sum := sha256.Sum256([]byte(requests(p2)[0].System))
	if got := hex.EncodeToString(sum[:]); got != "0e53d0948a0fec168752f979d9b13d70b0c7a5f0cad948a500247ddef73a008f" {
		t.Fatalf("C8: a debate's system prompt changed: %s", got)
	}
}

// C9: an inquiry with a rule bot and no record set is created, and played to
// its end.
func TestC9AnInquiryWithRuleBots(t *testing.T) {
	ts := newTestServer(t)
	req := inquiry(2, nil)
	req.Seats = []Seat{{Name: "R", Kind: "random"}, {Name: "G", Kind: "greedy"}}
	var st State
	if code := call(t, ts, "POST", "/api/games", req, &st); code != http.StatusOK {
		t.Fatalf("C9: new inquiry with rule bots: %d", code)
	}
	for i := 0; i < 200 && !st.Over; i++ {
		mr := post(t, ts, "/api/games/"+st.ID+"/ai", map[string]any{})
		st = mr.State
	}
	if !st.Over || st.Outcome == nil || st.Outcome.Winner != "" || st.Outcome.Verdict == "" {
		t.Fatalf("C9: over %v, outcome %+v", st.Over, st.Outcome)
	}
}
