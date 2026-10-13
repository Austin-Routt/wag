package game

import (
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
