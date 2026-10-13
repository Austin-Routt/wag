package game

// env.go provides a gym-style environment wrapper around GameState.
//
// The interface follows the OpenAI Gym / Gymnasium convention:
//   env.Reset()         → observation
//   env.Step(action)    → (observation, reward, done, truncated, info)
//   env.Render()        → string
//
// This supports:
//   - RL training agents (observation/reward loop)
//   - MCP tool integration (Reset/Step/Render as tool calls)
//   - CLI display (RenderState as shared utility)
//   - Agent-vs-agent harness (wrap RunGame or use Step directly)

import (
	"fmt"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// EnvConfig holds all parameters needed to create game environments.
// Passed to NewEnv; used by Reset to create fresh games.
type EnvConfig struct {
	Game        GameConfig
	YAMLBytes   []byte // Carneades YAML scheme definitions
	SchemesJSON []byte // NASCQ AS-node definitions
	CQsJSON     []byte // NASCQ CQ-node definitions
	EdgesJSON   []byte // NASCQ ask/backup edges
	PlayerNames []string
}

// Env is a gym-style environment wrapper around GameState.
// Thread-safe for single-threaded use. For concurrent access, add a mutex.
type Env struct {
	Config EnvConfig
	gs     *GameState
}

// StepResult is the return value of Env.Step, following the Gymnasium convention:
//
//	Observation — game state after the action
//	Reward      — score delta for the player who just acted
//	Done        — true if the game has terminated (any reason)
//	Truncated   — true if termination was due to max turns (not natural end)
//	Info        — extra metadata (accepted, error, player, arg_ids, etc.)
type StepResult struct {
	Observation *Observation
	Reward      float64
	Done        bool
	Truncated   bool
	Info        map[string]interface{}
}

// NewEnv creates an environment with the given configuration.
// Call Reset() before Step() to initialize a game.
func NewEnv(config EnvConfig) *Env {
	return &Env{Config: config}
}

// Reset creates a fresh game and returns the initial observation.
// Can be called multiple times to start new games.
func (e *Env) Reset() (*Observation, error) {
	adapter, err := carneades.NewAdapterFromYAML(e.Config.YAMLBytes)
	if err != nil {
		return nil, fmt.Errorf("YAML load: %w", err)
	}
	n, err := nascq.LoadFromBytes(e.Config.SchemesJSON, e.Config.CQsJSON, e.Config.EdgesJSON)
	if err != nil {
		return nil, fmt.Errorf("NASCQ load: %w", err)
	}
	e.gs = NewGame(e.Config.Game, adapter, n, e.Config.PlayerNames)
	return e.gs.BuildObservation(), nil
}

// Step processes one action and returns the result in gym format.
// The reward is the score delta for the player who just acted.
func (e *Env) Step(action nascq.ProposedAction) (*StepResult, error) {
	if e.gs == nil {
		return nil, fmt.Errorf("env not initialized; call Reset() first")
	}

	playerID := e.gs.ActivePlayer().ID
	result, err := e.gs.ExecuteTurn(action)
	if err != nil {
		return nil, err
	}

	obs := e.gs.BuildObservation()

	truncated := result.Terminated && e.gs.Truncated

	info := map[string]interface{}{
		"accepted":  result.Accepted,
		"abstained": result.Abstained,
		"player":    playerID,
		"arg_ids":   result.ArgIDs,
	}
	if result.Error != "" {
		info["error"] = result.Error
	}
	if result.Terminated {
		info["termination_reason"] = result.TerminationReason
		info["outcome"] = result.Outcome
	}

	return &StepResult{
		Observation: obs,
		Reward:      result.ScoreDelta,
		Done:        result.Terminated,
		Truncated:   truncated,
		Info:        info,
	}, nil
}

// GameState returns the underlying game state. Use for inspection only.
func (e *Env) GameState() *GameState { return e.gs }

// Render returns a human-readable string representation of the current state.
func (e *Env) Render() string {
	if e.gs == nil {
		return "(no game in progress)\n"
	}
	return RenderState(e.gs)
}

// =========================================================================
// RenderState — shared rendering utility
// =========================================================================

// RenderState produces a text representation of the game state.
// Used by Env.Render, HumanAgent, and the CLI main loop.
func RenderState(gs *GameState) string {
	var b strings.Builder

	// Header: the round (every player moves once a round), the mode, the
	// claim of a debate and its verdict so far, the scores.
	active := gs.ActivePlayer()
	fmt.Fprintf(&b, "\n══════════════════════════════════════════════\n")
	mode := ""
	if gs.Config.Mode != "" {
		mode = " | " + gs.Config.Mode
	}
	side := ""
	if sd := gs.Side(active.ID); sd != "" {
		side = " (" + sd + ")"
	}
	fmt.Fprintf(&b, "  Round %d%s | %s's turn%s\n", gs.CurrentTurn, mode, active.Name, side)
	if gs.Config.Objective == ObjectiveClaim {
		if gs.Claim == "" {
			fmt.Fprintf(&b, "  Claim: (pro's first argument will state it)\n")
		} else {
			fmt.Fprintf(&b, "  Claim: %s — %s\n", gs.Say(gs.Claim), labelString(gs.Adapter.Graph.Statements[gs.Claim].Label))
		}
	}
	fmt.Fprintf(&b, "  Scores:")
	for _, p := range gs.Players {
		marker := " "
		if p.ID == active.ID {
			marker = "►"
		}
		gone := ""
		if p.Resigned {
			gone = " (resigned)"
		}
		fmt.Fprintf(&b, "  %s%s: %.2f%s", marker, p.Name, p.Score, gone)
	}
	fmt.Fprintf(&b, "\n══════════════════════════════════════════════\n")

	open := 0
	for _, q := range gs.Questions {
		if !q.Open() {
			continue
		}
		if open == 0 {
			fmt.Fprintf(&b, "\n  Open questions:\n")
		}
		open++
		due := ""
		if gs.Config.AnswerWithin > 0 {
			due = fmt.Sprintf("  (%s must answer within %d turn(s))", q.Owner, q.TurnsLeft)
		}
		fmt.Fprintf(&b, "    ? %s asks %s about %s: %s%s\n", q.Asker, q.Owner, q.TargetArgID,
			gs.QuestionText(q.CQID, q.TargetArgID), due)
	}

	// Arguments in the graph: every argument a player put forward (attacks
	// included), with its premises and their labels. Exception, presumption
	// and question arguments are structure, shown through the labels and the
	// open questions.
	undercuts := map[*caes.Statement][]string{} // undercutter statement -> the arguments it undercuts
	for _, id := range sortedArgIDs(gs) {
		if u := gs.Adapter.Graph.Arguments[id].Undercutter; u != nil {
			undercuts[u] = append(undercuts[u], id)
		}
	}
	shown := 0
	for _, argID := range sortedArgIDs(gs) {
		arg := gs.Adapter.Graph.Arguments[argID]
		if !onBoard(arg) {
			continue
		}
		if shown == 0 {
			fmt.Fprintf(&b, "\n  Arguments in play:\n")
		}
		shown++

		conclusionText := gs.Say(arg.Conclusion.Id)
		if targets := undercuts[arg.Conclusion]; arg.Conclusion.IsUndercutter && len(targets) > 0 {
			conclusionText = "Undercuts " + strings.Join(targets, ", ") + ": the inference does not hold."
		}
		force := "in force"
		if arg.Weight == 0 {
			force = "out of force"
		}
		fmt.Fprintf(&b, "    [%s] %s\n", argID, conclusionText)
		fmt.Fprintf(&b, "         %-9s | %s | %s | %s\n",
			labelString(arg.Conclusion.Label), gs.SchemeName(arg), ownerOf(arg), force)
		for _, p := range arg.Premises {
			if p.Stmt != nil {
				fmt.Fprintf(&b, "           ← %-9s %s\n", labelString(p.Stmt.Label), gs.Say(p.Stmt.Id))
			}
		}
	}
	if shown == 0 {
		fmt.Fprintf(&b, "\n  Arguments: (none yet)\n")
	}

	return b.String()
}

// onBoard reports whether an argument is shown on the board: one a player
// put forward, not the structure of a scheme (exception arguments,
// presumptions) or a question.
func onBoard(arg *caes.Argument) bool {
	role := carneades.ArgumentRole(arg)
	return arg != nil && arg.Conclusion != nil && role != carneades.RoleException && role != carneades.RolePresumption &&
		role != carneades.RoleQuestion && !(role == "" && arg.Conclusion.IsUndercutter)
}

// labelString returns a human-friendly label with visual indicator.
func labelString(l caes.Label) string {
	switch l {
	case caes.In:
		return "In  ✓"
	case caes.Out:
		return "Out ✗"
	default:
		return "Undecided"
	}
}
