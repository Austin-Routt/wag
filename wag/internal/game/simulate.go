package game

import (
	"fmt"

	"github.com/yourorg/wag/internal/nascq"
)

// Simulate plays action as the active player, returns what ExecuteTurn
// returned, and puts the game back exactly as it was: the graph (through an
// adapter checkpoint, then a fresh evaluation, so labels and argument
// weights are as before), every player's state (score, flat points,
// abstentions, scheme usage, resignation), the turn and its clock, the
// history, the questions, the claim and the termination state. Agents that
// look ahead (GreedyAgent, or a planner) use it to try moves.
func (gs *GameState) Simulate(action nascq.ProposedAction) (*TurnResult, error) {
	return gs.SimulateThen(action, nil)
}

// SimulateThen is Simulate that also calls inspect, if not nil, on the game
// as the move left it, before putting it back: a planner can evaluate the
// position there (inspect must not change the game).
func (gs *GameState) SimulateThen(action nascq.ProposedAction, inspect func(*GameState)) (*TurnResult, error) {
	snap := gs.takeSnapshot()
	cp := gs.Adapter.Checkpoint()

	// A simulated move is played, not held for the moderator.
	moderated := gs.Config.Moderated
	gs.Config.Moderated = false
	defer func() { gs.Config.Moderated = moderated }()

	result, err := gs.ExecuteTurn(action)
	if inspect != nil {
		inspect(gs)
	}

	gs.Adapter.Restore(cp)
	gs.Adapter.Evaluate()
	gs.restoreSnapshot(snap)
	return result, err
}

// Reading is how a move's main argument would read on the board: its
// premises (the conclusion inside them as "C", see SayIn) and its
// conclusion, with every variable bound, the answer variable included.
type Reading struct {
	Premises   []string
	Conclusion string
}

// Read plays action in a simulation (Simulate) and returns how its main
// argument reads, or why the move would be rejected.
func (gs *GameState) Read(action nascq.ProposedAction) (*Reading, error) {
	var out *Reading
	r, err := gs.SimulateThen(action, func(g *GameState) {
		n := len(g.MoveHistory)
		if n == 0 || len(g.MoveHistory[n-1].ArgIDs) == 0 {
			return
		}
		arg := g.Adapter.Graph.Arguments[g.MoveHistory[n-1].ArgIDs[0]]
		if arg == nil || arg.Conclusion == nil {
			return
		}
		out = &Reading{Conclusion: g.Say(arg.Conclusion.Id)}
		for _, p := range arg.Premises {
			if p.Stmt != nil {
				out.Premises = append(out.Premises, g.SayIn(arg.Id, p.Stmt.Id))
			}
		}
	})
	if err != nil {
		return nil, err
	}
	if !r.Accepted {
		return nil, fmt.Errorf("%s", r.Error)
	}
	return out, nil
}
