package game

import (
	"fmt"
	"strings"

	"github.com/yourorg/wag/internal/nascq"
)

// Moderate gives the moderator's verdict on the held move (GameConfig.
// Moderated). Approved, the move is played, with weight in place of its
// scheme's base weight if weight > 0 (the SRD's Moderated mode: moderators
// weigh arguments). Rejected, it is recorded as an invalid move (with the
// usual penalty) and the same player moves again. The note is kept with the
// move.
func (gs *GameState) Moderate(approve bool, weight float64, note string) (*TurnResult, error) {
	if gs.Pending == nil {
		return nil, fmt.Errorf("no move awaits the moderator")
	}
	if weight < 0 || weight > 1 {
		return nil, fmt.Errorf("weight %v is not in [0, 1]", weight)
	}
	p := gs.Pending
	note = strings.TrimSpace(note)
	verdict := "approved"
	if !approve {
		verdict = "rejected"
	}
	if note != "" {
		verdict += ": " + note
	}
	if !approve {
		gs.Pending = nil
		msg := "rejected by the moderator"
		if note != "" {
			msg += ": " + note
		}
		r := gs.rejectInvalid(gs.ActivePlayer(), p.Action, msg)
		gs.MoveHistory[len(gs.MoveHistory)-1].Moderation = verdict
		gs.LastRejection = msg
		return r, nil
	}
	gs.moderating, gs.weightOverride = true, weight
	gs.TurnStartedAt = gs.now() // the moderator's time is not the player's
	r, err := gs.ExecuteTurn(p.Action)
	gs.moderating, gs.weightOverride = false, 0
	if err != nil {
		return nil, err
	}
	gs.Pending = nil
	if n := len(gs.MoveHistory); n > 0 {
		gs.MoveHistory[n-1].Moderation = verdict
	}
	return r, nil
}

// Decide records the moderator's decision on a statement: accepted, rejected
// or undecided (see carneades.Adapter.Decide). It is not a player's move:
// the turn does not pass. Under live scoring the scores follow the new
// labels.
func (gs *GameState) Decide(stmtID, decision, note string) (*TurnResult, error) {
	if !gs.Config.Moderated {
		return nil, fmt.Errorf("only a moderated game has a moderator")
	}
	if gs.Terminated {
		return nil, fmt.Errorf("game is already terminated: %s", gs.TerminationReason)
	}
	before := gs.Adapter.StatementLabels()
	id, err := gs.Adapter.Decide(stmtID, decision)
	if err != nil {
		return nil, err
	}
	gs.Adapter.Evaluate()
	deltas := map[string]float64{}
	if gs.Config.live() {
		deltas = gs.rescore()
	}
	gs.MoveHistory = append(gs.MoveHistory, Move{
		TurnNumber:   gs.CurrentTurn,
		PlayerID:     ModeratorID,
		Action:       nascq.ProposedAction{Type: "decide", Text: strings.TrimSpace(note)},
		ScoreDeltas:  deltas,
		Decision:     decision,
		DecisionStmt: id,
		LabelsBefore: before,
		LabelsAfter:  gs.Adapter.StatementLabels(),
	})
	return gs.ended(&TurnResult{Accepted: true, ScoreDeltas: deltas}), nil
}
