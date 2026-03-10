package game

import (
	"fmt"

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
		// Invalid action: penalize and record.
		player.Score += gs.Config.Scoring.InvalidActionPenalty

		gs.MoveHistory = append(gs.MoveHistory, Move{
			TurnNumber: gs.CurrentTurn,
			PlayerID:   player.ID,
			Action:     action,
			ScoreDelta: gs.Config.Scoring.InvalidActionPenalty,
		})

		// Invalid actions don't consume a turn — the player can try again.
		// For CLI Phase 1, we return the error so the caller can prompt again.
		return &TurnResult{
			Accepted: false,
			Error:    err.Error(),
		}, nil
	}

	// ---------------------------------------------------------------
	// Phase 3: Weight assignment (Trusted mode = scheme base_weight)
	// ---------------------------------------------------------------
	weight := validated.Weight

	// ---------------------------------------------------------------
	// Phase 4: Snapshot labels, apply mutation, validate acyclicity
	// ---------------------------------------------------------------
	labelsBefore := gs.Adapter.StatementLabels()

	argIDs, err := gs.applyAction(validated, player.ID, weight)
	if err != nil {
		return &TurnResult{
			Accepted: false,
			Error:    fmt.Sprintf("adapter error: %v", err),
		}, nil
	}

	// Acyclicity check (WAG game rule — Carneades handles cycles but WAG forbids them).
	if !gs.Adapter.IsAcyclic() {
		// Rollback: remove the arguments we just created.
		gs.Adapter.RemoveArguments(argIDs)
		return &TurnResult{
			Accepted: false,
			Error:    "move creates a cycle in the argument graph",
		}, nil
	}

	// ---------------------------------------------------------------
	// Phase 5: Evaluation
	// ---------------------------------------------------------------
	gs.Adapter.Evaluate()
	labelsAfter := gs.Adapter.StatementLabels()

	// ---------------------------------------------------------------
	// Phase 6: Scoring (Phase 1 formula)
	// ---------------------------------------------------------------
	scoreDelta := gs.computeScore(argIDs, validated, labelsBefore, labelsAfter)
	player.Score += scoreDelta
	player.ConsecutiveAbstentions = 0
	player.SchemeUsageCounts[action.SchemeID]++

	// ---------------------------------------------------------------
	// Phase 7: Record move, advance turn
	// ---------------------------------------------------------------
	gs.MoveHistory = append(gs.MoveHistory, Move{
		TurnNumber:   gs.CurrentTurn,
		PlayerID:     player.ID,
		Action:       action,
		ArgIDs:       argIDs,
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
		ScoreDelta:        scoreDelta,
		Terminated:        reason != "",
		TerminationReason: reason,
	}, nil
}

// -------------------------------------------------------------------------
// Action application: translates ValidatedAction into adapter calls
// -------------------------------------------------------------------------

// applyAction translates a NASCQ-validated action into Carneades adapter
// operations. For support actions, it delegates to AddSupport which uses
// InstantiateScheme. For attacks, it manually creates premise statements
// (from the NASCQ-instantiated English templates) and calls the appropriate
// adapter attack method.
//
// The premise creation for attacks works as follows:
//  1. Each instantiated English premise is converted to a clean Prolog atom
//     via carneades.TextToAtom (e.g., "carbon tax causes recession" →
//     "carbon_tax_causes_recession").
//  2. The atom is added as a Statement in the Carneades graph.
//  3. The atom is added as an Assumption (so it's labeled In during evaluation).
//  4. The atom IDs are collected and passed to the attack method.
//
// This approach is pragmatic for Phase 1. In Phase 2+, attacks could also
// use InstantiateScheme for the backup scheme.
func (gs *GameState) applyAction(
	va *nascq.ValidatedAction,
	ownerID string,
	weight float64,
) ([]string, error) {

	switch {
	// -------------------------------------------------------------------
	// Support: delegate to InstantiateScheme via adapter
	// -------------------------------------------------------------------
	case va.Type == "support":
			argIDs, err := gs.Adapter.AddSupport(ownerID, va.CarneadesSchemeID,
				va.VariableBindings, weight)
			if err != nil {
				return nil, err
			}
			// Trusted mode policy: assume premises of the main argument(s).
			// Exception arguments (whose conclusion IsUndercutter) are skipped —
			// assuming their premises would make exceptions true by default,
			// which would defeat the main argument immediately.
			for _, argID := range argIDs {
				arg := gs.Adapter.Graph.Arguments[argID]
				if arg == nil {
					continue
				}
				if arg.Conclusion != nil && arg.Conclusion.IsUndercutter {
					continue // skip exception arguments
				}
				for _, p := range arg.Premises {
					if p.Stmt != nil {
						gs.Adapter.AddAssumption(p.Stmt.Id)
					}
				}
			}
			return argIDs, nil

	// -------------------------------------------------------------------
	// Attacks: create premises manually, then call adapter attack method
	// -------------------------------------------------------------------
	case va.Type == "attack":
		// Create premise statements from instantiated English templates.
		premiseIDs, err := gs.createAttackPremises(va)
		if err != nil {
			return nil, fmt.Errorf("creating attack premises: %w", err)
		}

		switch va.CaesAttackType {
		case "undercutter":
			return gs.Adapter.AddUndercutter(ownerID, va.CarneadesSchemeID,
				va.TargetArgID, premiseIDs, weight)

		case "rebuttal":
			// Find the target argument's conclusion ID for the rebuttal.
			targetArg := gs.Adapter.Graph.Arguments[va.TargetArgID]
			if targetArg == nil || targetArg.Conclusion == nil {
				return nil, fmt.Errorf("target argument %q has no conclusion", va.TargetArgID)
			}
			return gs.Adapter.AddRebuttal(ownerID, va.CarneadesSchemeID,
				targetArg.Conclusion.Id, premiseIDs, weight, caes.PE)

		case "undermining":
			// Resolve target premise: if not specified, pick the first
			// non-assumed premise of the target argument.
			premiseID := va.TargetPremiseID
			if premiseID == "" {
				targetArg := gs.Adapter.Graph.Arguments[va.TargetArgID]
				if targetArg != nil {
					for _, p := range targetArg.Premises {
						if p.Stmt != nil {
							premiseID = p.Stmt.Id
							break
						}
					}
				}
			}
			if premiseID == "" {
				return nil, fmt.Errorf("no premise found to undermine in argument %q", va.TargetArgID)
			}
			return gs.Adapter.AddUndermining(ownerID, va.CarneadesSchemeID,
				premiseID, premiseIDs, weight, caes.PE)

		default:
			return nil, fmt.Errorf("unknown CAES attack type %q", va.CaesAttackType)
		}

	default:
		return nil, fmt.Errorf("unknown action type %q", va.Type)
	}
}

// createAttackPremises creates Statement nodes in the graph for each of the
// attack argument's premises (derived from NASCQ template instantiation).
// Each premise is assumed true so it contributes weight to the attack.
func (gs *GameState) createAttackPremises(va *nascq.ValidatedAction) ([]string, error) {
	premiseIDs := make([]string, 0, len(va.InstPremises))
	for _, premText := range va.InstPremises {
		// Convert English text to a clean Prolog atom for the statement ID.
		atomID := carneades.TextToAtom(premText)
		// AddStatement creates or returns existing; AddAssumption makes it In.
		gs.Adapter.AddStatement(atomID, premText)
		gs.Adapter.AddAssumption(atomID)
		premiseIDs = append(premiseIDs, atomID)
	}
	return premiseIDs, nil
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
) float64 {
	cfg := gs.Config.Scoring
	score := 0.0

	// Base points for each argument created.
	for _, argID := range argIDs {
		arg := gs.Adapter.Graph.Arguments[argID]
		if arg == nil || arg.Conclusion == nil {
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

		// Check if ALL players abstained this round.
		roundStart := len(gs.MoveHistory) - len(gs.Players)
		if roundStart < 0 {
			roundStart = 0
		}
		allAbstained := true
		count := 0
		for i := roundStart; i < len(gs.MoveHistory); i++ {
			if !gs.MoveHistory[i].Abstained {
				allAbstained = false
				break
			}
			count++
		}
		// Only count as all-abstain if we had a full round of moves.
		if allAbstained && count >= len(gs.Players) {
			gs.ConsecAllAbstain++
		} else {
			gs.ConsecAllAbstain = 0
		}
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
