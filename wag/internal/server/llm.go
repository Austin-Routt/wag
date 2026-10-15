package server

import (
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
	prompt := ss.Prompt()
	turn := len(gs.MoveHistory)
	ctx := s.ctx
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
	for attempt := 1; attempt <= maxModelAttempts; attempt++ {
		var callID string
		resp, err := c.Complete(ctx, c.Request(modelSystem, msgs), func(a llm.Attempt) {
			s.mu.Lock()
			callID = ss.addCall(a, turn, attempt)
			s.mu.Unlock()
		})
		if err != nil {
			keep(err)
			return nil, err
		}
		turnRec := Turn{Reply: resp.Text}
		s.mu.Lock()
		t := ss.newTrace(seat.Kind).withPrompt(prompt)
		s.mu.Unlock()
		t.Model, t.Profile, t.Provider, t.ServedModel = c.Profile.Model, c.Profile.Name, c.Provider.Name(), resp.Model
		t.Attempt, t.Reply, t.Feedback, t.Parent, t.Turn = attempt, resp.Text, feedback, callID, turn
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
			if res.Accepted || res.Abstained || res.Pending {
				rec.Turns = append(rec.Turns, turnRec)
				keep(nil)
				return res, nil
			}
			turnRec.Rejected = res.Error
		}
		rec.Turns = append(rec.Turns, turnRec)
		feedback = "That move was rejected: " + turnRec.Rejected + "\n\nThe game is as before. Reply with a legal move: one JSON object at the end."
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
