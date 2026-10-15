package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/llm"
	"github.com/yourorg/wag/internal/nascq"
)

// TraceFormat names the format of the trace file: a stream of events,
// one JSON object a line (docs/protocol.md, "What is recorded";
// docs/protocol/trace.schema.json).
//
//	game_start  the game: its configuration, seats and model profiles,
//	            seed, scheme data hash, protocol, build
//	call        one call of a model seat to its provider: the whole
//	            conversation sent, the parameters, the answer or the error,
//	            usage, latency, finish reason, request ID
//	decision    one decision of a seat (Trace): what it saw, what it
//	            replied, what it did, and what that did
//	game_end    the outcome, the final labels, the verdicts under each
//	            context (only once the game is over)
//
// Events have an id (g, cN, dN, e) and an order; a decision of a model seat
// has the call its reply came from as parent.
const TraceFormat = "wag-trace/2"

// Trace is one decision of a seat, a decision event of the game's trace
// file: what the player saw (the prompt, for a language-model seat), what it
// said (its reply, its own account of its reasoning, the predictions it
// made), what it did, and what that did. A rejected move is a decision too.
// Traces are for debugging the game and the prompt, and for training and
// evaluating players.
type Trace struct {
	Format string    `json:"format"`
	Type   string    `json:"type"` // "decision"
	ID     string    `json:"id"`   // dN, N its index among the game's decisions
	Parent string    `json:"parent,omitempty"`
	Order  int       `json:"order"` // its place in the trace file
	Game   string    `json:"game"`
	Seq    int       `json:"seq"`  // its index among the game's decisions
	Turn   int       `json:"turn"` // the moves made before it: the attempts at one move share it
	Move   int       `json:"move"` // the index of the move in the history, -1 if none was recorded
	Round  int       `json:"round"`
	Player string    `json:"player"`
	Side   string    `json:"side,omitempty"` // pro or con, in a debate
	Seat   string    `json:"seat"`           // human, llm, model, claude, random, greedy
	Time   time.Time `json:"time"`
	// Profile, Provider and Model are a model seat's configuration;
	// ServedModel is the model (version) the provider says answered.
	Profile     string `json:"profile,omitempty"`
	Provider    string `json:"provider,omitempty"`
	Model       string `json:"model,omitempty"`
	ServedModel string `json:"served_model,omitempty"`
	Attempt     int    `json:"attempt"` // 1 for the first try at a turn
	Protocol    string `json:"protocol,omitempty"`
	// PromptSHA256 identifies the prompt; Prompt is the text itself, kept for
	// language-model seats: the board as the turn began, on every attempt.
	// Feedback is what the seat was told on a retry (why its previous move
	// was rejected); the call event (Parent) has the whole conversation.
	PromptSHA256 string `json:"prompt_sha256,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	Feedback     string `json:"feedback,omitempty"`
	Reply        string `json:"reply,omitempty"` // the raw reply of a language-model seat
	// Reasoning is what the reply says of its reasoning (context, thoughts,
	// summary, expect, alternatives, confidence), without the action: the
	// model's own account, which need not be how it chose (docs/protocol.md).
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
	return &Trace{Format: TraceFormat, Type: "decision", Game: ss.ID, Turn: len(gs.MoveHistory), Move: -1, Round: gs.CurrentTurn, Player: p.ID,
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
	t.ID = "d" + strconv.Itoa(t.Seq)
	ss.order++
	t.Order = ss.order
	ss.Traces = append(ss.Traces, *t)
}

// Call is one call of a model seat to its provider, a call event of the
// trace file: the request as sent (the system prompt, the whole
// conversation, the parameters), and the answer or the error, with the
// usage, the latency and the provider's identifiers. A call that failed and
// was tried again (a rate limit) is one event per try (N).
type Call struct {
	Format string `json:"format"`
	Type   string `json:"type"` // "call"
	ID     string `json:"id"`   // cN
	Order  int    `json:"order"`
	Game   string `json:"game"`
	Player string `json:"player"`
	Turn   int    `json:"turn"` // as the decisions' Turn
	// ForAttempt is the decision attempt (Trace.Attempt) the call is for.
	ForAttempt int `json:"attempt"`
	llm.Attempt
}

// addCall records a call and returns its ID. Caller holds s.mu.
func (ss *Session) addCall(a llm.Attempt, turn, attempt int) string {
	ss.order++
	c := Call{Format: TraceFormat, Type: "call", ID: "c" + strconv.Itoa(len(ss.Calls)), Order: ss.order, Game: ss.ID,
		Turn: turn, ForAttempt: attempt, Attempt: a}
	if p := ss.GS.ActivePlayer(); p != nil {
		c.Player = p.ID
	}
	ss.Calls = append(ss.Calls, c)
	return c.ID
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

// GameStart is the first event of a trace file: everything needed to set
// the game up again (wag-llm replay does).
type GameStart struct {
	Format     string            `json:"format"`
	Type       string            `json:"type"` // "game_start"
	ID         string            `json:"id"`   // "g"
	Order      int               `json:"order"`
	Game       string            `json:"game"`
	Time       time.Time         `json:"time"`
	Title      string            `json:"title"`
	Source     string            `json:"source,omitempty"`
	Protocol   string            `json:"protocol"`
	DataSHA256 string            `json:"data_sha256"`
	Seed       int64             `json:"seed"`
	Request    *NewGameRequest   `json:"request,omitempty"`
	Config     game.GameConfig   `json:"config"`
	Seats      []Seat            `json:"seats"`
	Profiles   []llm.Profile     `json:"profiles,omitempty"` // the model seats' profiles (no keys: they are not in profiles)
	System     string            `json:"system_prompt"`      // the model seats' system prompt, unless a profile replaces it
	Build      map[string]string `json:"build"`
	Contexts   []game.Context    `json:"contexts,omitempty"`
	Players    []map[string]any  `json:"players"`
}

// GameEnd is the last event of the trace file of a game that is over.
type GameEnd struct {
	Format   string            `json:"format"`
	Type     string            `json:"type"` // "game_end"
	ID       string            `json:"id"`   // "e"
	Order    int               `json:"order"`
	Game     string            `json:"game"`
	Moves    int               `json:"moves"`
	Outcome  *game.Outcome     `json:"outcome"`
	Labels   map[string]string `json:"labels"` // every statement's label at the end
	Verdicts []Verdict         `json:"verdicts,omitempty"`
}

// gameStart describes the game for its trace file. Caller holds s.mu.
func (s *Server) gameStart(ss *Session) GameStart {
	g := GameStart{Format: TraceFormat, Type: "game_start", ID: "g", Game: ss.ID, Time: ss.Created.UTC(), Title: ss.Title, Source: ss.Source,
		Protocol: ProtocolVersion, DataSHA256: ss.dataHash, Seed: ss.Seed, Request: ss.Request, Config: ss.GS.Config,
		Seats: ss.Seats, System: modelSystem, Build: buildInfo(), Contexts: ss.contexts}
	seen := map[string]bool{}
	for _, seat := range ss.Seats {
		if c, _ := s.model(seat); c != nil && !seen[c.Profile.Name] {
			seen[c.Profile.Name] = true
			g.Profiles = append(g.Profiles, c.Profile)
		}
	}
	for _, p := range ss.GS.Players {
		g.Players = append(g.Players, map[string]any{"id": p.ID, "side": ss.GS.Side(p.ID)})
	}
	return g
}

// writeTraces writes the game's trace file (TraceFormat): its start, its
// calls and decisions in the order they happened, and its end if it is
// over. Caller holds s.mu.
func (s *Server) writeTraces(w http.ResponseWriter, ss *Session) {
	w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=wag-traces-"+ss.ID+".jsonl")
	s.encodeTraces(w, ss)
}

// encodeTraces writes the game's trace file to w. Caller holds s.mu.
func (s *Server) encodeTraces(w io.Writer, ss *Session) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s.gameStart(ss))
	i, j := 0, 0
	for i < len(ss.Traces) || j < len(ss.Calls) {
		if j == len(ss.Calls) || i < len(ss.Traces) && ss.Traces[i].Order < ss.Calls[j].Order {
			_ = enc.Encode(ss.Traces[i])
			i++
		} else {
			_ = enc.Encode(ss.Calls[j])
			j++
		}
	}
	if o := ss.GS.Outcome(); o != nil && o.Over {
		labels := map[string]string{}
		for id, l := range ss.GS.Adapter.StatementLabels() {
			labels[id] = labelName(l)
		}
		_ = enc.Encode(GameEnd{Format: TraceFormat, Type: "game_end", ID: "e", Order: ss.order + 1, Game: ss.ID,
			Moves: len(ss.GS.MoveHistory), Outcome: o, Labels: labels, Verdicts: ss.verdicts(ss.contexts)})
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
