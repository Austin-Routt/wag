package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"
)

// Claude plays a seat through the Claude API, with the text protocol: the
// seat's prompt (Session.Prompt) goes to the model, its reply is parsed
// (ParseReply) and played. A rejected move goes back to the model with the
// reason, in the same conversation, up to three times; then the seat passes.
// Credentials are the SDK's: ANTHROPIC_API_KEY, or a profile from
// `ant auth login`.
type Claude struct {
	Model  string // e.g. claude-opus-5-5
	Effort string // low, medium, high, xhigh, max
	client anthropic.Client
	// askFn, if set, replaces the API (tests).
	askFn func(ctx context.Context, msgs []anthropic.BetaMessageParam) (string, anthropic.BetaMessageParam, error)
}

// NewClaude makes the Claude player the server offers as a seat.
func NewClaude(model, effort string) *Claude {
	if model == "" {
		model = "claude-opus-5-5"
	}
	if effort == "" {
		effort = "medium"
	}
	return &Claude{Model: model, Effort: effort, client: anthropic.NewClient()}
}

const claudeSystem = `You play one seat of WAG, a two-player argumentation game built on Walton's argumentation schemes and the Carneades model of argument. Play well and honestly: make arguments that are true and relevant to the discussion, in plain words, and pick the moves that help your side. Every reply ends with exactly one JSON object: your move, in the format the prompt gives.`

// Turn is one exchange with the model: what it replied, the move read from
// it, and why the game rejected the move (if it did).
type Turn struct {
	Reply    string               `json:"reply"`
	Action   nascq.ProposedAction `json:"action"`
	Rejected string               `json:"rejected,omitempty"`
}

// ask sends the conversation and returns the model's reply, as text and as
// the assistant message that continues the conversation.
func (c *Claude) ask(ctx context.Context, msgs []anthropic.BetaMessageParam) (string, anthropic.BetaMessageParam, error) {
	if c.askFn != nil {
		return c.askFn(ctx, msgs)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	resp, err := c.client.Beta.Messages.New(ctx, anthropic.BetaMessageNewParams{
		Model:        anthropic.Model(c.Model),
		MaxTokens:    16000,
		System:       []anthropic.BetaTextBlockParam{{Text: claudeSystem}},
		Messages:     msgs,
		OutputConfig: anthropic.BetaOutputConfigParam{Effort: anthropic.BetaOutputConfigEffort(c.Effort)},
		// A declined request is served again by a model the server picks
		// for the refusal's category, instead of failing.
		Fallbacks: anthropic.BetaFallbacksParamUnion{OfDefault: constant.ValueOf[constant.Default]()},
		Betas:     []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
	})
	if err != nil {
		return "", anthropic.BetaMessageParam{}, fmt.Errorf("Claude API: %w", err)
	}
	if resp.StopReason == anthropic.BetaStopReasonRefusal {
		return "", anthropic.BetaMessageParam{}, fmt.Errorf("Claude declined to move (%s)", resp.StopDetails.Category)
	}
	var text strings.Builder
	for _, b := range resp.Content {
		if t, ok := b.AsAny().(anthropic.BetaTextBlock); ok {
			text.WriteString(t.Text)
		}
	}
	return text.String(), resp.ToParam(), nil
}

// LLMRecord is what a Claude seat said on its way to one move, kept for the
// export.
type LLMRecord struct {
	Player string `json:"player"`
	Model  string `json:"model"`
	Round  int    `json:"round"`
	Turns  []Turn `json:"turns"`
	Error  string `json:"error,omitempty"`
}

// claudeTurn lets the Claude seat to move play: it asks the model for a move
// and plays it, sending a rejected move back with the reason, up to four
// times, then passes. It is called without s.mu and with ss.thinking set; it
// takes s.mu itself around every access to the game, and not during the API
// calls, so other games and the pollers are not held up.
func (s *Server) claudeTurn(ss *Session) (*game.TurnResult, error) {
	s.mu.Lock()
	rec := LLMRecord{Player: ss.GS.ActivePlayer().ID, Model: s.claude.Model, Round: ss.GS.CurrentTurn}
	msgs := []anthropic.BetaMessageParam{anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(ss.Prompt()))}
	s.mu.Unlock()
	keep := func(err error) {
		if err != nil {
			rec.Error = err.Error()
		}
		s.mu.Lock()
		ss.LLM = append(ss.LLM, rec)
		s.mu.Unlock()
	}
	for attempt := 0; attempt < 4; attempt++ {
		reply, assistant, err := s.claude.ask(context.Background(), msgs)
		if err != nil {
			keep(err)
			return nil, err
		}
		turn := Turn{Reply: reply}
		msgs = append(msgs, assistant)
		action, perr := ParseReply(reply)
		turn.Action = action
		if perr != nil {
			turn.Rejected = perr.Error()
		} else {
			s.mu.Lock()
			res, err := ss.GS.ExecuteTurn(action)
			s.mu.Unlock()
			if err != nil {
				rec.Turns = append(rec.Turns, turn)
				keep(err)
				return nil, err
			}
			if res.Accepted || res.Abstained || res.Pending {
				rec.Turns = append(rec.Turns, turn)
				keep(nil)
				return res, nil
			}
			turn.Rejected = res.Error
		}
		rec.Turns = append(rec.Turns, turn)
		msgs = append(msgs, anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(
			"That move was rejected: "+turn.Rejected+"\n\nThe game is as before. Reply with a legal move: one JSON object at the end.")))
	}
	s.mu.Lock()
	res, err := ss.GS.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	s.mu.Unlock()
	keep(err)
	return res, err
}
