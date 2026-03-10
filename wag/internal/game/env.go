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
	"sort"
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

	truncated := false
	if result.Terminated && strings.Contains(result.TerminationReason, "max turns") {
		truncated = true
	}

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

	// Header.
	active := gs.ActivePlayer()
	fmt.Fprintf(&b, "\n══════════════════════════════════════════════\n")
	fmt.Fprintf(&b, "  Turn %d | %s's turn\n", gs.CurrentTurn, active.Name)
	fmt.Fprintf(&b, "  Scores:")
	for _, p := range gs.Players {
		marker := " "
		if p.ID == active.ID {
			marker = "►"
		}
		fmt.Fprintf(&b, "  %s%s: %.2f", marker, p.Name, p.Score)
	}
	fmt.Fprintf(&b, "\n══════════════════════════════════════════════\n")

	// Arguments in the graph.
	if len(gs.Adapter.Graph.Arguments) == 0 {
		fmt.Fprintf(&b, "\n  Arguments: (none yet)\n")
	} else {
		fmt.Fprintf(&b, "\n  Arguments in play:\n")

		// Sort argument IDs for stable display order.
		argIDs := make([]string, 0, len(gs.Adapter.Graph.Arguments))
		for id := range gs.Adapter.Graph.Arguments {
			argIDs = append(argIDs, id)
		}
		sort.Strings(argIDs)

		for _, argID := range argIDs {
			arg := gs.Adapter.Graph.Arguments[argID]
			if arg == nil {
				continue
			}

			// Skip undercutter-conclusion arguments in display
			// (they're internal mechanics, not player-facing claims).
			if arg.Conclusion != nil && arg.Conclusion.IsUndercutter {
				continue
			}

			conclusionText := "(nil)"
			label := "?"
			if arg.Conclusion != nil {
				conclusionText = arg.Conclusion.Id
				if arg.Conclusion.Text != "" && arg.Conclusion.Text != arg.Conclusion.Id {
					conclusionText = arg.Conclusion.Text
				}
				label = labelString(arg.Conclusion.Label)
			}

			owner := "system"
			if arg.Metadata != nil {
				if o, ok := arg.Metadata["owner"].(string); ok {
					owner = o
				}
			}

			schemeID := ""
			if arg.Scheme != nil {
				schemeID = arg.Scheme.Id
			}

			fmt.Fprintf(&b, "    [%s] %s\n", argID, conclusionText)
			fmt.Fprintf(&b, "         Label: %-10s  Scheme: %-25s  Owner: %s\n",
				label, schemeID, owner)
		}
	}

	return b.String()
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
