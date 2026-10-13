package game

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// ExecuteTurn processes one player's action through the complete turn pipeline.
//
// Pipeline phases (from WAG plan §16):
//
//	Phase 0: Parse action type (abstain vs. move)
//	Phase 2: NASCQ structural validation (S1–S8)
//	Phase 3: Weight assignment (scheme base_weight in Trusted mode)
//	Phase 4: Apply mutation to graph, check acyclicity, rollback if needed
//	Phase 5: Run GroundedLabelling evaluation
//	Phase 6: Compute score delta
//	Phase 7: Record move, advance turn
//	Phase 8: Check termination conditions
//
// Phase 1 (moderator review) and Phase 7 (WebSocket broadcast) are skipped
// in Phase 1 of the project (CLI trusted mode).
//
// Returns a TurnResult indicating whether the move was accepted, the score
// delta, any arguments created, and whether the game has terminated.
func (gs *GameState) ExecuteTurn(action nascq.ProposedAction) (*TurnResult, error) {
	r, err := gs.executeTurn(action)
	if err == nil {
		gs.LastRejection = ""
		if !r.Accepted {
			gs.LastRejection = r.Error
		}
	}
	return r, err
}

func (gs *GameState) executeTurn(action nascq.ProposedAction) (*TurnResult, error) {
	if gs.Terminated {
		return nil, fmt.Errorf("game is already terminated: %s", gs.TerminationReason)
	}

	player := gs.ActivePlayer()

	// ---------------------------------------------------------------
	// Phase 0: Handle abstain
	// ---------------------------------------------------------------
	if action.Type == "abstain" {
		player.ConsecutiveAbstentions++

		// Record move.
		gs.MoveHistory = append(gs.MoveHistory, Move{
			TurnNumber: gs.CurrentTurn,
			PlayerID:   player.ID,
			Action:     action,
			Abstained:  true,
		})

		gs.advanceTurn()
		reason := gs.checkTermination()

		return &TurnResult{
			Accepted:          true,
			Abstained:         true,
			Terminated:        reason != "",
			TerminationReason: reason,
		}, nil
	}

	// ---------------------------------------------------------------
	// Phase 2: NASCQ structural validation
	// ---------------------------------------------------------------
	validated, err := gs.NASCQ.Validate(action, gs.bridge)
	if err != nil {
		return gs.rejectInvalid(player, action, err.Error()), nil
	}

	// ---------------------------------------------------------------
	// Phase 3: Weight assignment (Trusted mode = scheme base_weight)
	// ---------------------------------------------------------------
	weight := validated.Weight

	// ---------------------------------------------------------------
	// Phase 4: Snapshot labels, apply mutation, validate acyclicity
	// ---------------------------------------------------------------
	labelsBefore := gs.Adapter.StatementLabels()
	targetStanding := gs.standing(validated.TargetArgID)
	checkpoint := gs.Adapter.Checkpoint() // a rejected move must leave no trace (review defect #5)

	argIDs, presumptions, err := gs.applyAction(validated, player.ID, weight)
	if err != nil {
		gs.Adapter.Restore(checkpoint) // the move may have failed half-way
		return &TurnResult{
			Accepted: false,
			Error:    fmt.Sprintf("adapter error: %v", err),
		}, nil
	}

	// Acyclicity check (WAG game rule — Carneades handles cycles but WAG forbids them).
	if !gs.Adapter.IsAcyclic() {
		// Rollback: restore the graph exactly as it was before the move. Removing
		// the move's arguments alone left its statements, issues, assumptions and
		// undercutter pointers behind.
		gs.Adapter.Restore(checkpoint)
		return &TurnResult{
			Accepted: false,
			Error:    "move creates a cycle in the argument graph",
		}, nil
	}

	// G1: a move must not restate an argument already in the graph (same
	// scheme, conclusion and premises, whoever made it). Repeating a support
	// move used to score it again every turn.
	if dup := gs.duplicateArgument(argIDs); dup != "" {
		gs.Adapter.Restore(checkpoint)
		return gs.rejectInvalid(player, action, fmt.Sprintf("G1: this argument is already in the graph as %s", dup)), nil
	}

	// ---------------------------------------------------------------
	// Phase 5: Evaluation
	// ---------------------------------------------------------------
	gs.Adapter.Evaluate()
	labelsAfter := gs.Adapter.StatementLabels()

	// ---------------------------------------------------------------
	// Phase 6: Scoring (Phase 1 formula)
	// ---------------------------------------------------------------
	scoreDelta := gs.computeScore(argIDs, validated, labelsBefore, labelsAfter, targetStanding)
	player.Score += scoreDelta
	player.ConsecutiveAbstentions = 0
	gs.roundHadMove = true
	player.SchemeUsageCounts[action.SchemeID]++

	// ---------------------------------------------------------------
	// Phase 7: Record move, advance turn
	// ---------------------------------------------------------------
	gs.MoveHistory = append(gs.MoveHistory, Move{
		TurnNumber:   gs.CurrentTurn,
		PlayerID:     player.ID,
		Action:       validated.ProposedAction, // with the resolved attack type and undermined premise
		ArgIDs:       argIDs,
		Presumptions: presumptions,
		Weight:       weight,
		ScoreDelta:   scoreDelta,
		LabelsBefore: labelsBefore,
		LabelsAfter:  labelsAfter,
	})

	gs.advanceTurn()

	// ---------------------------------------------------------------
	// Phase 8: Termination check
	// ---------------------------------------------------------------
	reason := gs.checkTermination()

	return &TurnResult{
		Accepted:          true,
		ArgIDs:            argIDs,
		Presumptions:      presumptions,
		ScoreDelta:        scoreDelta,
		Terminated:        reason != "",
		TerminationReason: reason,
	}, nil
}

// -------------------------------------------------------------------------
// Action application: translates ValidatedAction into adapter calls
// -------------------------------------------------------------------------

// applyAction translates a NASCQ-validated action into Carneades adapter
// operations and returns the arguments of the move and the presumption
// arguments made for its premises.
//
// A plain support instantiates its scheme with InstantiateScheme. Every move
// with a target (an attack or a defence) answers a critical question about
// it: Adapter.AnswerQuestion makes an instance of the backup scheme whose
// conclusion is the answer, decided by the locus (premise, exception,
// conclusion or the inference) and the polarity.
//
// Premises are presumed, not assumed (review defect #1). Trusted mode used to
// call AddAssumption on every premise of a move, and Carneades labels an
// assumption In whatever is argued against it, so an undermining attack could
// never succeed. Now every premise of a move gets a presumption argument owned
// by the mover (weight GameConfig.PresumptionWeight). The assumptions that
// InstantiateScheme adds for a scheme's declared assumptions (e.g.
// possible(A)) are retracted and presumed in the same way. Exception
// arguments' premises are neither assumed nor presumed: an exception holds
// only if someone argues for it.
func (gs *GameState) applyAction(
	va *nascq.ValidatedAction,
	ownerID string,
	weight float64,
) ([]string, []string, error) {
	var argIDs []string
	var err error
	switch {
	case va.TargetArgID == "" && va.Type == "support":
		nAssumptions := gs.Adapter.AssumptionCount()
		if argIDs, err = gs.Adapter.AddSupport(ownerID, va.CarneadesSchemeID, va.VariableBindings, weight); err != nil {
			return nil, nil, err
		}
		gs.Adapter.RetractAssumptionsSince(nAssumptions)
	case va.TargetArgID != "":
		argIDs, err = gs.Adapter.AnswerQuestion(carneades.Answer{
			Owner:          ownerID,
			SchemeID:       va.CarneadesSchemeID,
			Bindings:       va.VariableBindings,
			Weight:         weight,
			TargetArgID:    va.TargetArgID,
			Locus:          va.Locus.Kind,
			LocusStmtID:    va.LocusStmtID,
			Positive:       va.IsDefence(),
			AnswerVariable: va.AnswerVariable,
			Standard:       caes.PE,
		})
		if err != nil {
			return argIDs, nil, err
		}
		if v := va.AnswerVariable; v != "" { // record what the adapter bound it to
			b := make(map[string]string, len(va.VariableBindings)+1)
			for k, x := range va.VariableBindings {
				b[k] = x
			}
			b[v] = gs.Adapter.Graph.Arguments[argIDs[0]].Conclusion.Id
			va.VariableBindings = b
		}
	default:
		return nil, nil, fmt.Errorf("unknown action type %q", va.Type)
	}

	var presumptions []string
	for _, argID := range argIDs {
		arg := gs.Adapter.Graph.Arguments[argID]
		if arg == nil || carneades.ArgumentRole(arg) != carneades.RoleMain {
			continue // exception arguments: their premise holds only if argued for
		}
		for _, p := range arg.Premises {
			if p.Stmt == nil {
				continue
			}
			pid, err := gs.Adapter.AddPresumption(ownerID, p.Stmt.Id, gs.Config.presumptionWeight())
			if err != nil {
				return argIDs, presumptions, err
			}
			presumptions = appendNew(presumptions, pid)
		}
	}
	return argIDs, presumptions, nil
}

// rejectInvalid penalizes and records an invalid move. An invalid move does
// not end the turn: the player can try again.
func (gs *GameState) rejectInvalid(player *Player, action nascq.ProposedAction, msg string) *TurnResult {
	player.Score += gs.Config.Scoring.InvalidActionPenalty
	gs.MoveHistory = append(gs.MoveHistory, Move{
		TurnNumber: gs.CurrentTurn,
		PlayerID:   player.ID,
		Action:     action,
		ScoreDelta: gs.Config.Scoring.InvalidActionPenalty,
	})
	return &TurnResult{Accepted: false, Error: msg}
}

// duplicateArgument returns the ID of an argument that was in the graph before
// the move and that one of the move's arguments restates: a main argument with
// the same scheme, conclusion and premises. It returns "" if there is none.
func (gs *GameState) duplicateArgument(moveArgs []string) string {
	key := func(arg *caes.Argument) string {
		prem := make([]string, 0, len(arg.Premises))
		for _, p := range arg.Premises {
			if p.Stmt != nil {
				prem = append(prem, p.Stmt.Id)
			}
		}
		sort.Strings(prem)
		scheme := ""
		if arg.Scheme != nil {
			scheme = arg.Scheme.Id
		}
		return scheme + "|" + arg.Conclusion.Id + "|" + strings.Join(prem, ",")
	}
	isNew := make(map[string]bool, len(moveArgs))
	keys := map[string]bool{}
	for _, id := range moveArgs {
		isNew[id] = true
		if arg := gs.Adapter.Graph.Arguments[id]; arg != nil && arg.Conclusion != nil && carneades.ArgumentRole(arg) == carneades.RoleMain {
			keys[key(arg)] = true
		}
	}
	for _, id := range sortedArgIDs(gs) {
		arg := gs.Adapter.Graph.Arguments[id]
		if isNew[id] || arg.Conclusion == nil || carneades.ArgumentRole(arg) != carneades.RoleMain {
			continue
		}
		if keys[key(arg)] {
			return id
		}
	}
	return ""
}

func appendNew(list []string, id string) []string {
	for _, x := range list {
		if x == id {
			return list
		}
	}
	return append(list, id)
}

// -------------------------------------------------------------------------
// Scoring (Phase 1)
// -------------------------------------------------------------------------

// computeScore calculates the Phase 1 score delta for a move.
//
// Formula: score_delta = base_points[conclusion_label] × weight + flat_bonus
//
// Where:
//   - conclusion_label is the label of the new argument's conclusion AFTER evaluation
//   - weight is the scheme's base_weight
//   - flat_bonus is +UndercutterBonus if the undercutter flipped the target from In/Undecided to Out
//
// Phase 2 adds: scheme_multiplier, consistency_multiplier, chain_multiplier,
// repetition_penalty, and retroactive delta scoring. Not implemented here.
func (gs *GameState) computeScore(
	argIDs []string,
	va *nascq.ValidatedAction,
	labelsBefore, labelsAfter map[string]caes.Label,
	targetStanding bool,
) float64 {
	cfg := gs.Config.Scoring
	score := 0.0

	if va.Type == "attack" {
		// An attack on one's own argument is legal (the SRD lets self-attacks
		// count against a player's consistency, which is not implemented yet)
		// but earns nothing: undercutting your own argument used to score
		// 7 + 15.
		if gs.Adapter.ArgumentOwner(va.TargetArgID) == gs.ActivePlayer().ID {
			return 0
		}
		// Nor does an attack on an argument that was already defeated (its
		// weight 0 or its conclusion Out): it changes nothing. Undercutting
		// an argument that was already Out used to score 7 every time.
		if !targetStanding {
			return 0
		}
	}

	// Proleptic defence (SRD §16.4): a positive answer to a critical question
	// about one's own argument that nobody has raised yet.
	if va.IsDefence() && va.AnswersCQID != "" && cfg.ProlepticBonus != 0 &&
		gs.Adapter.ArgumentOwner(va.TargetArgID) == gs.ActivePlayer().ID && !gs.raised(va.TargetArgID, va.AnswersCQID) {
		score += cfg.ProlepticBonus
	}

	// Base points for each argument the player put forward. Exception
	// arguments (part of the scheme) and presumptions earn nothing: an
	// exception's conclusion is the mover's own undercutter.
	for _, argID := range argIDs {
		arg := gs.Adapter.Graph.Arguments[argID]
		if arg == nil || arg.Conclusion == nil || carneades.ArgumentRole(arg) != carneades.RoleMain {
			continue
		}
		label := labelsAfter[arg.Conclusion.Id]
		score += basePoints(label, cfg) * va.Weight
	}

	// Undercutter bonus: if the target argument's conclusion flipped to Out.
	if va.CaesAttackType == "undercutter" && va.TargetArgID != "" {
		targetArg := gs.Adapter.Graph.Arguments[va.TargetArgID]
		if targetArg != nil && targetArg.Conclusion != nil {
			cid := targetArg.Conclusion.Id
			before := labelsBefore[cid]
			after := labelsAfter[cid]
			if (before == caes.In || before == caes.Undecided) && after == caes.Out {
				score += cfg.UndercutterBonus
			}
		}
	}

	return score
}

// standing reports whether the argument is in force: it exists, its weight
// is above 0 (not undercut, its premises In) and its conclusion is not Out.
func (gs *GameState) standing(argID string) bool {
	arg := gs.Adapter.Graph.Arguments[argID]
	return arg != nil && arg.Weight > 0 && arg.Conclusion != nil && arg.Conclusion.Label != caes.Out
}

// raised reports whether an accepted move has already answered the critical
// question about the argument, either way.
func (gs *GameState) raised(argID, cqID string) bool {
	for _, m := range gs.MoveHistory {
		if len(m.ArgIDs) > 0 && m.Action.TargetArgID == argID && m.Action.AnswersCQID == cqID {
			return true
		}
	}
	return false
}

// basePoints returns the Phase 1 base points for a given label.
func basePoints(label caes.Label, cfg ScoringConfig) float64 {
	switch label {
	case caes.In:
		return cfg.BasePointsIn
	case caes.Undecided:
		return cfg.BasePointsUndecided
	default: // caes.Out
		return cfg.BasePointsOut
	}
}

// -------------------------------------------------------------------------
// Turn management
// -------------------------------------------------------------------------

// advanceTurn moves to the next player and increments the turn counter
// when a full round completes. Also tracks consecutive all-abstain rounds.
func (gs *GameState) advanceTurn() {
	gs.ActivePlayerIdx++

	// If we've gone through all players, that's one full round.
	if gs.ActivePlayerIdx >= len(gs.Players) {
		gs.ActivePlayerIdx = 0
		gs.CurrentTurn++

		// Every turn ends with an accepted move or an abstention (a rejected
		// move does not end the turn), so all players abstained this round
		// exactly when no move was accepted in it. The history cannot be used
		// for this: it also records rejected moves, so a rejected move followed
		// by an abstention hid an all-abstain round.
		if gs.roundHadMove {
			gs.ConsecAllAbstain = 0
		} else {
			gs.ConsecAllAbstain++
		}
		gs.roundHadMove = false
	}
}

// -------------------------------------------------------------------------
// Termination
// -------------------------------------------------------------------------

// checkTermination evaluates termination conditions in priority order.
// Returns the reason string if the game should end, or "" to continue.
//
// Conditions (from WAG plan §18):
//  1. Max turns reached
//  2. All-abstain threshold exceeded
//
// Proof-standard termination (Phase 2+) is not implemented here.
func (gs *GameState) checkTermination() string {
	// Condition 1: Max turns reached.
	if gs.Config.MaxTurns > 0 && gs.CurrentTurn > gs.Config.MaxTurns {
		gs.Terminated = true
		gs.TerminationReason = fmt.Sprintf("max turns reached (%d)", gs.Config.MaxTurns)
		return gs.TerminationReason
	}

	// Condition 2: All players abstained for threshold consecutive rounds.
	if gs.Config.AbstentionThreshold > 0 && gs.ConsecAllAbstain >= gs.Config.AbstentionThreshold {
		gs.Terminated = true
		gs.TerminationReason = fmt.Sprintf("all players abstained for %d consecutive rounds",
			gs.ConsecAllAbstain)
		return gs.TerminationReason
	}

	return ""
}
