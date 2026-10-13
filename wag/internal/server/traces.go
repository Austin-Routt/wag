package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"
)

// TraceFormat names the format of a decision trace.
const TraceFormat = "wag-trace/1"

// Trace is one decision of a seat, a line of the game's trace file: what the
// player saw (the prompt, for a language-model seat), what it said (its
// reply, its own account of its reasoning, the predictions it made), what it
// did, and what that did. A rejected move is a decision too. Traces are for
// debugging the game and the prompt, and for training and evaluating
// players (docs/protocol.md, "What is recorded"); docs/protocol/trace.schema.json
// describes them.
type Trace struct {
	Format   string    `json:"format"`
	Game     string    `json:"game"`
	Seq      int       `json:"seq"`  // its index among the game's traces
	Move     int       `json:"move"` // the index of the move in the history, -1 if none was recorded
	Round    int       `json:"round"`
	Player   string    `json:"player"`
	Side     string    `json:"side,omitempty"` // pro or con, in a debate
	Seat     string    `json:"seat"`           // human, llm, claude, random, greedy
	Model    string    `json:"model,omitempty"`
	Attempt  int       `json:"attempt"` // 1 for the first try at a turn
	Time     time.Time `json:"time"`
	Protocol string    `json:"protocol,omitempty"`
	// PromptSHA256 identifies the prompt; Prompt is the text itself, kept for
	// language-model seats. On a retry it is the message that said why the
	// move was rejected.
	PromptSHA256 string `json:"prompt_sha256,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	Reply        string `json:"reply,omitempty"` // the raw reply of a language-model seat
	// Reasoning is what the reply says of its reasoning (context, thoughts,
	// summary, expect, alternatives, confidence), without the action.
	Reasoning *Reply `json:"reasoning,omitempty"`
	Why       string `json:"why,omitempty"` // a human player's note on the move

	Action nascq.ProposedAction `json:"action"`
	Result TraceResult          `json:"result"`

	ClaimBefore string        `json:"claim_before,omitempty"`
	ClaimAfter  string        `json:"claim_after,omitempty"`
	Effects     []string      `json:"effects,omitempty"`
	ExpectCheck []ExpectCheck `json:"expect_check,omitempty"`
	// AlternativeOutcomes are what the moves the player considered and did
	// not make would have done, tried on the game before its move and undone.
	AlternativeOutcomes []AltOutcome `json:"alternative_outcomes,omitempty"`
}

// TraceResult is what the game made of the move.
type TraceResult struct {
	Accepted   bool     `json:"accepted"`
	Abstained  bool     `json:"abstained,omitempty"`
	Pending    bool     `json:"pending,omitempty"`
	Error      string   `json:"error,omitempty"`
	ArgIDs     []string `json:"arg_ids,omitempty"`
	ScoreDelta float64  `json:"score_delta,omitempty"`
}

// ExpectCheck compares a predicted label with the label after the move.
type ExpectCheck struct {
	Stmt      string `json:"stmt"`
	Predicted string `json:"predicted"`
	Actual    string `json:"actual"`
	OK        bool   `json:"ok"`
}

// AltOutcome is what an alternative move would have done.
type AltOutcome struct {
	Action     nascq.ProposedAction `json:"action"`
	Legal      bool                 `json:"legal"`
	Error      string               `json:"error,omitempty"`
	ClaimAfter string               `json:"claim_after,omitempty"`
	ScoreDelta float64              `json:"score_delta"`
}

// maxAlternatives is how many alternatives of one reply are tried.
const maxAlternatives = 3

// newTrace starts the trace of a decision of the player to move.
func (ss *Session) newTrace(seat string) *Trace {
	gs := ss.GS
	p := gs.ActivePlayer()
	return &Trace{Format: TraceFormat, Game: ss.ID, Move: -1, Round: gs.CurrentTurn, Player: p.ID,
		Side: gs.Side(p.ID), Seat: seat, Attempt: 1, Time: time.Now().UTC()}
}

// withPrompt records the prompt a language-model seat was given.
func (t *Trace) withPrompt(prompt string) *Trace {
	h := sha256.Sum256([]byte(prompt))
	t.Protocol, t.PromptSHA256, t.Prompt = ProtocolVersion, hex.EncodeToString(h[:]), prompt
	return t
}

// play plays a move for the player to move and records its trace: the
// alternatives the reply names are tried first, then the move, and the
// predictions are checked against the labels it leaves. rep, if not nil, is
// the reply the move came from. Caller holds s.mu.
func (ss *Session) play(action nascq.ProposedAction, t *Trace, rep *Reply) (*game.TurnResult, error) {
	gs := ss.GS
	r := newRefs(gs)
	action = r.resolveAction(action)
	if rep != nil {
		reasoning := *rep
		reasoning.Action = nascq.ProposedAction{}
		t.Reasoning = &reasoning
		for i, alt := range rep.Alternatives {
			if i == maxAlternatives {
				break
			}
			t.AlternativeOutcomes = append(t.AlternativeOutcomes, ss.tryAlternative(r.resolveAction(alt.Action)))
		}
	}
	t.ClaimBefore = ss.claimLabel()
	t.Action = action
	res, err := gs.ExecuteTurn(action)
	if err != nil {
		t.Result.Error = err.Error()
		ss.addTrace(t)
		return res, err
	}
	t.Result = TraceResult{Accepted: res.Accepted, Abstained: res.Abstained, Pending: res.Pending, Error: res.Error,
		ArgIDs: res.ArgIDs, ScoreDelta: res.ScoreDelta}
	t.ClaimAfter = ss.claimLabel()
	if n := len(gs.MoveHistory); n > 0 && (res.Accepted || res.Abstained) && !res.Pending {
		t.Move = n - 1
		t.Action = gs.MoveHistory[n-1].Action // as validated: resolved attack type and answer
		t.Effects = logEntry(gs, n-1, gs.MoveHistory[n-1], madeByIndex(gs)).Effects
	}
	if rep != nil && res.Accepted && !res.Pending {
		labels := gs.Adapter.StatementLabels()
		for _, e := range rep.Expect {
			id := e.Stmt
			if strings.EqualFold(id, "claim") {
				id = gs.Claim
			}
			id = r.resolve(id)
			c := ExpectCheck{Stmt: id, Predicted: e.Label, Actual: "absent"}
			if l, ok := labels[id]; ok {
				c.Actual = labelName(l)
			}
			c.OK = c.Actual == c.Predicted
			t.ExpectCheck = append(t.ExpectCheck, c)
		}
	}
	ss.addTrace(t)
	return res, nil
}

// tryAlternative plays a move on the game and undoes it.
func (ss *Session) tryAlternative(a nascq.ProposedAction) AltOutcome {
	out := AltOutcome{Action: a}
	r, err := ss.GS.SimulateThen(a, func(g *game.GameState) {
		if g.Claim != "" {
			if s := g.Adapter.Graph.Statements[g.Claim]; s != nil {
				out.ClaimAfter = labelName(s.Label)
			}
		}
	})
	switch {
	case err != nil:
		out.Error = err.Error()
	case !r.Accepted && !r.Abstained:
		out.Error = r.Error
	default:
		out.Legal, out.ScoreDelta = true, r.ScoreDelta
	}
	return out
}

func (ss *Session) claimLabel() string {
	if gs := ss.GS; gs.Claim != "" {
		if s := gs.Adapter.Graph.Statements[gs.Claim]; s != nil {
			return labelName(s.Label)
		}
	}
	return ""
}

func (ss *Session) addTrace(t *Trace) {
	t.Seq = len(ss.Traces)
	ss.Traces = append(ss.Traces, *t)
}

// madeByIndex maps each argument to the index of the move that made it.
func madeByIndex(gs *game.GameState) map[string]int {
	m := map[string]int{}
	for i, mv := range gs.MoveHistory {
		for _, id := range mv.ArgIDs {
			m[id] = i
		}
	}
	return m
}

// writeTraces writes the game's traces as JSON Lines.
func (ss *Session) writeTraces(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=wag-traces-"+ss.ID+".jsonl")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, t := range ss.Traces {
		_ = enc.Encode(t)
	}
}

// Verdict is the claim's label under a context preset's weights.
type Verdict struct {
	Context string `json:"context"`
	Name    string `json:"name"`
	Claim   string `json:"claim"`
}

// verdicts says how the claim would stand under each context preset's
// weights, and under the game's own: how much the outcome depends on who
// weighs the arguments (game.GameState.ClaimUnder). Nil without a claim.
func (ss *Session) verdicts(contexts []game.Context) []Verdict {
	gs := ss.GS
	if gs.Claim == "" {
		return nil
	}
	out := []Verdict{{Context: "this game", Name: "The weights of this game", Claim: ss.claimLabel()}}
	for _, c := range contexts {
		out = append(out, Verdict{Context: c.ID, Name: c.Name, Claim: gs.ClaimUnder(c.Weights)})
	}
	return out
}

// tracesWithoutPrompts are the traces for the export: the prompts are left
// out (their hashes stay), as they make up most of a game's size; the trace
// file (/traces) has them.
func (ss *Session) tracesWithoutPrompts() []Trace {
	out := make([]Trace, len(ss.Traces))
	for i, t := range ss.Traces {
		t.Prompt = ""
		out[i] = t
	}
	return out
}
