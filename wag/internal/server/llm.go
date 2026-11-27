package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/llm"
	"github.com/yourorg/wag/internal/nascq"
)

// Model seats are played by a language model through an API: the seat's
// prompt (Session.Prompt) goes to the model of a profile (llm.Profile:
// Gemini, Claude, an OpenAI-compatible endpoint, a command), its reply is
// parsed (ParseReplyFull) and played. A rejected move goes back to the
// model with the reason, in the same conversation, up to three times; then
// the seat passes. Every call (llm.Attempt) and every decision is traced.

// modelSystem is a model seat's system prompt (a profile can replace it).
const modelSystem = `You play one seat of WAG, a two-player argumentation game built on Walton's argumentation schemes and the Carneades model of argument. Play well and honestly: make arguments that are true and relevant to the discussion, in plain words, and pick the moves that help your side. Every reply ends with exactly one JSON object: your move, in the format the prompt gives.`

// modelSystemInquiry is a model seat's system prompt in an inquiry, where the
// players examine the claim together (docs/e1-evidence.md, E1.0). Other
// games keep modelSystem, which their starts record.
const modelSystemInquiry = `You play one seat of WAG, a two-player argumentation game built on Walton's argumentation schemes and the Carneades model of argument. This game is an inquiry: you and the other player examine a claim together, with no side and no winner. Play well and honestly: make arguments that are true and relevant to the discussion, in plain words, and pick the moves that bring out what holds. Every reply ends with exactly one JSON object: your move, in the format the prompt gives.`

// system is the session's system prompt for its model seats.
func (ss *Session) system() string {
	if ss.GS.Config.Cooperative {
		return modelSystemInquiry
	}
	return modelSystem
}

// maxModelAttempts is how many replies a model seat may give for one move.
const maxModelAttempts = 4

// Turn is one exchange with the model: what it replied, the move read from
// it, and why the game rejected the move (if it did).
type Turn struct {
	Reply    string               `json:"reply"`
	Action   nascq.ProposedAction `json:"action"`
	Rejected string               `json:"rejected,omitempty"`
}

// LLMRecord is what a model seat said on its way to one move, kept for the
// export.
type LLMRecord struct {
	Player   string `json:"player"`
	Profile  string `json:"profile,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model"`
	Round    int    `json:"round"`
	Turns    []Turn `json:"turns"`
	Error    string `json:"error,omitempty"`
}

// AddModel offers seats played by the client's profile, under its name. A
// profile with the name of one already offered replaces it.
func (s *Server) AddModel(c *llm.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.models[c.Profile.Name] = c
	s.meta = nil
}

// SetClaude offers Claude seats (the seat kind "claude"): the profile
// "claude", through the Claude API.
func (s *Server) SetClaude(model, effort string) error {
	if model == "" {
		model = "claude-opus-5-5"
	}
	if effort == "" {
		effort = "medium"
	}
	c, err := llm.NewClient(llm.Profile{Name: "claude", Provider: "anthropic", Params: llm.Params{Model: model, Effort: effort}})
	if err != nil {
		return err
	}
	s.AddModel(c)
	return nil
}

// model returns the client that plays a seat, nil for a seat no model
// plays; an error for a model seat whose profile this server does not have.
func (s *Server) model(seat Seat) (*llm.Client, error) {
	name := seat.Model
	switch seat.Kind {
	case "claude":
		name = "claude"
	case "model":
	default:
		return nil, nil
	}
	if c := s.models[name]; c != nil {
		return c, nil
	}
	if seat.Kind == "claude" {
		return nil, fmt.Errorf("this server has no Claude seats: start it with --claude")
	}
	if name == "" {
		return nil, fmt.Errorf("a model seat names its profile (model); this server has %s", s.modelNames())
	}
	return nil, fmt.Errorf("this server has no model profile %q; it has %s (start it with --model or --models)", name, s.modelNames())
}

func (s *Server) modelNames() string {
	if len(s.models) == 0 {
		return "none"
	}
	var out []string
	for k := range s.models {
		out = append(out, k)
	}
	sort.Strings(out)
	return fmt.Sprint(out)
}

// modelTurn lets the model seat to move play: it asks the model for a move
// and plays it, sending a rejected move back with the reason, up to four
// replies, then passes. It is called without s.mu and with ss.thinking set;
// it takes s.mu itself around every access to the game, and not during the
// API calls, so other games and the pollers are not held up.
//
// Every call to the provider is traced (Session.addCall), with the whole
// conversation it sent; every reply is a decision (Trace), whose prompt is
// the board the turn started from and whose parent is the call it came
// from.
func (s *Server) modelTurn(ss *Session, c *llm.Client) (*game.TurnResult, error) {
	s.mu.Lock()
	gs := ss.GS
	seat := ss.Seats[gs.ActivePlayerIdx]
	rec := LLMRecord{Player: gs.ActivePlayer().ID, Profile: c.Profile.Name, Provider: c.Provider.Name(), Model: c.Profile.Model, Round: gs.CurrentTurn}
	// The game clock first: a seat is not asked to move in a game whose time
	// is up.
	if r := gs.CheckClock(); r != nil {
		s.mu.Unlock()
		return r, nil
	}
	prompt := ss.Prompt()
	turn := len(gs.MoveHistory)
	// No try is sent while the game cannot be saved: a try whose record
	// could not be saved stops the call's retries (llm.WithGate).
	ctx := llm.WithGate(s.ctx, func() error {
		s.mu.Lock()
		defer s.mu.Unlock()
		if ss.saveErr != "" {
			return errUnsaved
		}
		return nil
	})
	// The turn may be in progress: replies recorded for it that no decision
	// used (the game stopped, or could not be saved, after the provider
	// answered), or refused replies whose next attempt was not made. It goes
	// on from its record, and the model is asked only what it was not
	// (docs/g-journal.md).
	tr := ss.turnInProgress(gs.ActivePlayer().ID)
	s.mu.Unlock()
	keep := func(err error) {
		if err != nil {
			rec.Error = err.Error()
		}
		s.mu.Lock()
		ss.LLM = append(ss.LLM, rec)
		s.mu.Unlock()
	}
	msgs := []llm.Message{{Role: "user", Content: prompt}}
	feedback := ""
	start := 1
	if tr != nil {
		// The turn's start, the board it began on, and the conversation so
		// far: the earlier replies and why they were refused.
		turn, msgs, start = tr.turn, tr.msgs, tr.attempt
		prompt = msgs[0].Content
		if start > 1 {
			feedback = msgs[len(msgs)-1].Content
		}
	}
	for attempt := start; attempt <= maxModelAttempts; attempt++ {
		var callID string
		var resp *llm.Response
		req := c.Request(ss.system(), msgs)
		recorded := tr.reply(attempt, req)
		if recorded != nil {
			callID, resp = recorded.ID, recorded.Response
		} else {
			var err error
			resp, err = c.Complete(ctx, req, func(a llm.Attempt) {
				s.mu.Lock()
				callID = ss.addCall(a, turn, attempt)
				// Kept before used: the call's record, with the reply or the
				// error, is saved before the reply is read or played.
				s.autosave(ss)
				s.mu.Unlock()
			})
			if err != nil {
				keep(err)
				return nil, err
			}
			s.mu.Lock()
			unsaved := ss.saveErr != ""
			s.mu.Unlock()
			if unsaved {
				err := errors.New("the model's reply is kept, but the game could not be saved: the reply is played once the game is saved, and the model is not asked again")
				keep(err)
				return nil, err
			}
			if s.stopAfterReply != nil && s.stopAfterReply() {
				return nil, errStoppedAfterReply
			}
		}
		turnRec := Turn{Reply: resp.Text}
		s.mu.Lock()
		t := ss.newTrace(seat.Kind).withPrompt(prompt)
		s.mu.Unlock()
		t.Model, t.Profile, t.Provider, t.ServedModel = c.Profile.Model, c.Profile.Name, c.Provider.Name(), resp.Model
		t.Attempt, t.Reply, t.Feedback, t.Parent, t.Turn = attempt, resp.Text, feedback, callID, turn
		t.Recovered = recorded != nil
		rep, perr := ParseReplyFull(resp.Text)
		if perr != nil {
			turnRec.Rejected = perr.Error()
			t.Result.Error = perr.Error()
			s.mu.Lock()
			ss.addTrace(t)
			s.mu.Unlock()
		} else {
			turnRec.Action = rep.Action
			s.mu.Lock()
			res, err := ss.play(rep.Action, t, rep)
			s.mu.Unlock()
			if err != nil {
				rec.Turns = append(rec.Turns, turnRec)
				keep(err)
				return nil, err
			}
			if res.Accepted || res.Abstained || res.Pending || res.Terminated {
				rec.Turns = append(rec.Turns, turnRec)
				keep(nil)
				return res, nil
			}
			turnRec.Rejected = res.Error
		}
		rec.Turns = append(rec.Turns, turnRec)
		feedback = rejectedFeedback(turnRec.Rejected)
		msgs = append(msgs, llm.Message{Role: "assistant", Content: resp.Text}, llm.Message{Role: "user", Content: feedback})
	}
	s.mu.Lock()
	t := ss.newTrace(seat.Kind).withPrompt(prompt)
	t.Model, t.Profile, t.Provider, t.Attempt, t.Turn = c.Profile.Model, c.Profile.Name, c.Provider.Name(), maxModelAttempts+1, turn
	t.Why = fmt.Sprintf("%d replies were rejected: the seat passes", maxModelAttempts)
	res, err := ss.play(nascq.ProposedAction{Type: "abstain"}, t, nil)
	s.mu.Unlock()
	keep(err)
	return res, err
}

// errStoppedAfterReply is the stop a test makes between a provider's
// answer and its use (Server.stopAfterReply), as a crash would.
var errStoppedAfterReply = errors.New("the turn stopped after the model's reply was recorded")

// errUnsaved stops a model seat's call before a try is sent while the game
// cannot be saved.
var errUnsaved = errors.New("the game could not be saved: its model seats wait until it is, and are asked nothing meanwhile")

// rejectedFeedback is what a model seat is told when its move is refused,
// before its next attempt.
func rejectedFeedback(reason string) string {
	return "That move was rejected: " + reason + "\n\nThe game is as before. Reply with a legal move: one JSON object at the end."
}

// turnRecord is where a model seat's turn in progress stands, read from the
// game's record (docs/g-journal.md): the turn's start, the attempt to make
// next and the conversation for it, and, by attempt, the replies recorded
// for the turn that no decision used.
type turnRecord struct {
	turn, attempt int
	msgs          []llm.Message
	replies       map[int]*Call
}

// reply is the reply recorded for the attempt, if one is and it answered
// the request the turn sends now: the same conversation, system prompt and
// parameters. Otherwise the provider is asked.
func (tr *turnRecord) reply(attempt int, req llm.Request) *Call {
	if tr == nil || tr.replies[attempt] == nil {
		return nil
	}
	c := tr.replies[attempt]
	a, err1 := json.Marshal(c.Request)
	b, err2 := json.Marshal(req)
	if err1 != nil || err2 != nil || !bytes.Equal(a, b) {
		return nil
	}
	return c
}

// replied: the call has a reply that came without an error. A reply that
// came with one (a partial or blocked answer, or one the budget refused) is
// not played, live or recovered.
func replied(c *Call) bool { return c.Response != nil && c.Error == nil }

// turnInProgress reads the record of the model seat p's turn in progress:
// its last call, and the calls before it in the same turn, attempt by
// attempt, with the decisions that used them. The conversation goes on from
// the turn's prompt through the replies decisions used, and refused, each
// with the reason, as the live turn has it; the replies no decision used
// are played again in their attempts (turnRecord.reply). Nil when the turn
// starts afresh: no call is recorded; the moves made since the seat's last
// call are not only its own refused moves (the board moved on: another
// player's move, the moderator's, a move of its own); that call failed (a
// failed call ends a turn, live too); or the record is not one a turn
// leaves. Caller holds s.mu.
func (ss *Session) turnInProgress(p string) *turnRecord {
	h := ss.GS.MoveHistory
	last := -1
	for i := len(ss.Calls) - 1; i >= 0 && last < 0; i-- {
		if ss.Calls[i].Player == p {
			last = i
		}
	}
	if last < 0 || !replied(&ss.Calls[last]) {
		return nil
	}
	turn := ss.Calls[last].Turn
	if turn > len(h) {
		return nil
	}
	for _, m := range h[turn:] {
		if m.PlayerID != p || m.Rejection == "" {
			return nil
		}
	}
	// The turn's calls, back to its first: its first attempt's first try.
	first := last
	for first >= 0 && (ss.Calls[first].ForAttempt != 1 || ss.Calls[first].N != 1) {
		first--
	}
	if first < 0 {
		return nil
	}
	final := map[int]*Call{} // each attempt's last try: the one that answered
	for i := first; i <= last; i++ {
		c := &ss.Calls[i]
		if c.Player != p || c.Turn != turn {
			return nil
		}
		final[c.ForAttempt] = c
	}
	n := ss.Calls[last].ForAttempt
	if len(final) != n || final[1] == nil || len(final[1].Request.Messages) == 0 {
		return nil
	}
	if turn == len(h) && final[1].Request.Messages[0].Content != ss.Prompt() {
		// No move since the turn began: its first request was this board,
		// or the record is not this turn's.
		return nil
	}
	used := map[string]*Trace{}
	for i := range ss.Traces {
		if d := &ss.Traces[i]; d.Parent != "" {
			used[d.Parent] = d
		}
	}
	tr := &turnRecord{turn: turn, replies: map[int]*Call{}}
	msgs := []llm.Message{final[1].Request.Messages[0]} // the board the turn began on
	for a := 1; a <= n; a++ {
		c := final[a]
		if c == nil || !replied(c) {
			return nil
		}
		switch d := used[c.ID]; {
		case d == nil:
			if tr.attempt == 0 {
				tr.attempt = a
			}
			cc := *c // a copy: the turn reads it without the lock
			tr.replies[a] = &cc
		case d.Result.Pending:
			// Held for the moderator: the turn ended with it. Rejected, the
			// seat's next turn starts afresh, on the board with the
			// moderator's note, as a refused move does not.
			return nil
		case tr.attempt != 0:
			return nil // a decision after a reply none used: not a turn's record
		default:
			// Used, and refused (a move played would have moved the board
			// on): the conversation goes on with it and the reason.
			msgs = append(msgs, llm.Message{Role: "assistant", Content: c.Response.Text},
				llm.Message{Role: "user", Content: rejectedFeedback(d.Result.Error)})
		}
	}
	if tr.attempt == 0 {
		tr.attempt = n + 1 // every reply used and refused: the next attempt
	}
	tr.msgs = msgs
	return tr
}

// modelMeta describes the model profiles a server offers, for the browser.
func (s *Server) modelMeta() []map[string]any {
	var out []map[string]any
	for _, name := range sortedKeys(s.models) {
		c := s.models[name]
		out = append(out, map[string]any{"name": name, "provider": c.Provider.Name(), "model": c.Profile.Model, "effort": c.Profile.Effort})
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
