package game

import (
	"github.com/yourorg/wag/internal/nascq"
)

// Simulate plays action as the active player, returns what ExecuteTurn
// returned, and puts the game back exactly as it was: the graph (through an
// adapter checkpoint, then a fresh evaluation, so labels and argument
// weights are as before), every player's score, abstention count and scheme
// usage, the turn, the history, the termination state and LastRejection.
// Agents that look ahead (GreedyAgent, or a planner) use it to try moves.
func (gs *GameState) Simulate(action nascq.ProposedAction) (*TurnResult, error) {
	type playerState struct {
		score      float64
		abstains   int
		schemeUses map[string]int
	}
	players := make([]playerState, len(gs.Players))
	for i, p := range gs.Players {
		uses := make(map[string]int, len(p.SchemeUsageCounts))
		for k, v := range p.SchemeUsageCounts {
			uses[k] = v
		}
		players[i] = playerState{p.Score, p.ConsecutiveAbstentions, uses}
	}
	turn, active, history := gs.CurrentTurn, gs.ActivePlayerIdx, len(gs.MoveHistory)
	allAbstain, roundHadMove := gs.ConsecAllAbstain, gs.roundHadMove
	terminated, reason, rejection := gs.Terminated, gs.TerminationReason, gs.LastRejection
	cp := gs.Adapter.Checkpoint()

	result, err := gs.ExecuteTurn(action)

	gs.Adapter.Restore(cp)
	gs.Adapter.Evaluate()
	for i, p := range gs.Players {
		p.Score, p.ConsecutiveAbstentions, p.SchemeUsageCounts = players[i].score, players[i].abstains, players[i].schemeUses
	}
	gs.CurrentTurn, gs.ActivePlayerIdx, gs.MoveHistory = turn, active, gs.MoveHistory[:history]
	gs.ConsecAllAbstain, gs.roundHadMove = allAbstain, roundHadMove
	gs.Terminated, gs.TerminationReason, gs.LastRejection = terminated, reason, rejection
	return result, err
}
