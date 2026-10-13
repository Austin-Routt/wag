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
	// Clocks: the game clock ends the game before the move; a late move is
	// void (the turn passes, or the player loses).
	// ---------------------------------------------------------------
	now := gs.now()
	if gs.Config.TimeLimit > 0 && now.Sub(gs.StartedAt) >= gs.Config.TimeLimit {
		gs.terminate(fmt.Sprintf("time limit reached (%v)", gs.Config.TimeLimit))
		gs.Truncated = true
		return gs.ended(&TurnResult{Error: "time limit reached: the move was not played"}), nil
	}
	timedOut := false
	if gs.Config.MoveTimeLimit > 0 && now.Sub(gs.TurnStartedAt) > gs.Config.MoveTimeLimit && action.Type != "resign" {
		if gs.Config.MoveTimeoutLoses {
			gs.MoveHistory = append(gs.MoveHistory, Move{TurnNumber: gs.CurrentTurn, PlayerID: player.ID, Action: action, TimedOut: true})
			gs.forfeit(player, fmt.Sprintf("%s ran out of time (%v a move)", player.ID, gs.Config.MoveTimeLimit))
			return gs.ended(&TurnResult{TimedOut: true, Error: "out of time"}), nil
		}
		action, timedOut = nascq.ProposedAction{Type: "abstain"}, true
	}

	switch action.Type {
	case "resign":
		player.Resigned = true
		gs.MoveHistory = append(gs.MoveHistory, Move{TurnNumber: gs.CurrentTurn, PlayerID: player.ID, Action: action, Resigned: true})
		active := 0
		for _, p := range gs.Players {
			if !p.Resigned {
				active++
			}
		}
		if active < 2 || gs.Config.Objective == ObjectiveClaim {
			gs.forfeit(player, player.ID+" resigned")
			return gs.ended(&TurnResult{Accepted: true}), nil
		}
		gs.endTurn(player) // the others play on
		return gs.ended(&TurnResult{Accepted: true}), nil

	case "abstain":
		player.ConsecutiveAbstentions++
		gs.MoveHistory = append(gs.MoveHistory, Move{
			TurnNumber: gs.CurrentTurn,
			PlayerID:   player.ID,
			Action:     action,
			Abstained:  true,
			TimedOut:   timedOut,
		})
		// A side that passes while the verdict is against it concedes.
		if gs.Config.ConcedeOnPass && gs.Side(player.ID) != "" && !gs.verdictFavours(player.ID) {
			what := "conceded the claim"
			if gs.Side(player.ID) == "pro" {
				what = "passed without carrying the burden of proof"
			}
			gs.forfeit(player, fmt.Sprintf("%s (%s) %s", player.ID, gs.Side(player.ID), what))
			return gs.ended(&TurnResult{Accepted: true, Abstained: true, TimedOut: timedOut}), nil
		}
		gs.endTurn(player)
		return gs.ended(&TurnResult{Accepted: true, Abstained: true, TimedOut: timedOut}), nil
	}

	// ---------------------------------------------------------------
	// Phase 2: NASCQ structural validation
	// ---------------------------------------------------------------
	validated, err := gs.NASCQ.Validate(action, gs.bridge)
	if err != nil {
		return gs.rejectInvalid(player, action, err.Error()), nil
	}
	if validated.Type == "question" && gs.raised(validated.TargetArgID, validated.AnswersCQID) {
		return gs.rejectInvalid(player, action, fmt.Sprintf("Q2: %s has already been raised about %s",
			validated.AnswersCQID, validated.TargetArgID)), nil
	}

	// ---------------------------------------------------------------
	// Phase 3: Weight assignment (Trusted mode = scheme base_weight)
	// ---------------------------------------------------------------
	weight := validated.Weight

	// ---------------------------------------------------------------
	// Phase 4: Snapshot labels, apply mutation, validate acyclicity
	// ---------------------------------------------------------------
	labelsBefore := gs.Adapter.StatementLabels()
	// An attack on one's own argument is legal (self-play, belief revision)
	// but earns nothing; nor does an attack on an argument already out of
	// force (weight 0: undercut, or a premise Out), which changes nothing.
	// Undercutting one's own argument used to score 7 + 15, and undercutting
	// an argument that was already Out 7 every time.
	scorable := validated.Type != "question"
	if validated.Type == "attack" {
		scorable = gs.Adapter.ArgumentOwner(validated.TargetArgID) != player.ID && gs.standing(validated.TargetArgID)
	}
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

	// The opening argument of a debate states its claim, which is at issue
	// from then on.
	if gs.Config.Objective == ObjectiveClaim && gs.Claim == "" && validated.Type == "support" &&
		validated.TargetArgID == "" && (gs.Side(player.ID) != "con") {
		claim, err := gs.Adapter.MakeIssue(gs.Adapter.Graph.Arguments[argIDs[0]].Conclusion.Id, gs.Config.ClaimStandard)
		if err != nil {
			gs.Adapter.Restore(checkpoint)
			return &TurnResult{Accepted: false, Error: fmt.Sprintf("adapter error: %v", err)}, nil
		}
		gs.Claim = claim
	}

	// ---------------------------------------------------------------
	// Phase 5: Evaluation
	// ---------------------------------------------------------------
	gs.Adapter.Evaluate()
	labelsAfter := gs.Adapter.StatementLabels()

	// ---------------------------------------------------------------
	// Phase 6: Scoring
	// ---------------------------------------------------------------
	base, flat := 0.0, 0.0
	if scorable {
		base, flat = gs.computeScore(argIDs, validated, labelsBefore, labelsAfter)
	}
	player.Flat += flat
	var deltas map[string]float64
	gs.MoveHistory = append(gs.MoveHistory, Move{
		TurnNumber:   gs.CurrentTurn,
		PlayerID:     player.ID,
		Action:       validated.ProposedAction, // with the resolved attack type and undermined premise
		ArgIDs:       argIDs,
		Presumptions: presumptions,
		Weight:       weight,
		LabelsBefore: labelsBefore,
		LabelsAfter:  labelsAfter,
		Scorable:     scorable,
	})
	if gs.Config.live() {
		deltas = gs.rescore()
	} else {
		player.Score += base + flat
		deltas = map[string]float64{}
		if base+flat != 0 {
			deltas[player.ID] = base + flat
		}
	}
	m := &gs.MoveHistory[len(gs.MoveHistory)-1]
	m.ScoreDelta, m.ScoreDeltas = deltas[player.ID], deltas
	player.ConsecutiveAbstentions = 0
	gs.roundHadMove = true
	if validated.SchemeID != "" {
		player.SchemeUsageCounts[validated.SchemeID]++
	}

	// ---------------------------------------------------------------
	// Phase 7: Questions, then advance the turn
	// ---------------------------------------------------------------
	switch {
	case validated.Type == "question":
		gs.Questions = append(gs.Questions, OpenQuestion{
			ArgID:       argIDs[0],
			TargetArgID: validated.TargetArgID,
			CQID:        validated.AnswersCQID,
			LocusStmtID: validated.LocusStmtID,
			Asker:       player.ID,
			Owner:       gs.Adapter.ArgumentOwner(validated.TargetArgID),
			AskedAt:     len(gs.MoveHistory) - 1,
			TurnsLeft:   gs.Config.AnswerWithin,
		})
	case validated.IsDefence() && validated.AnswersCQID != "":
		for i := range gs.Questions {
			q := &gs.Questions[i]
			if q.Open() && q.TargetArgID == validated.TargetArgID && q.CQID == validated.AnswersCQID {
				q.AnsweredBy = argIDs[0]
			}
		}
	}
	gs.endTurn(player)

	return gs.ended(&TurnResult{
		Accepted:     true,
		ArgIDs:       argIDs,
		Presumptions: presumptions,
		ScoreDelta:   deltas[player.ID],
		ScoreDeltas:  deltas,
	}), nil
}

// endTurn ends the player's turn: the questions about their arguments that
// they still have not answered come closer to their deadline, and the turn
// passes on (unless the game is over).
func (gs *GameState) endTurn(player *Player) {
	current := len(gs.MoveHistory) - 1
	for i := range gs.Questions {
		q := &gs.Questions[i]
		if !q.Open() || q.Owner != player.ID || q.AskedAt == current {
			continue
		}
		q.TurnsLeft--
		if gs.Config.AnswerWithin > 0 && q.TurnsLeft <= 0 && gs.Loser == "" {
			gs.forfeit(player, fmt.Sprintf("%s left %s about %s unanswered", player.ID, q.CQID, q.TargetArgID))
			return
		}
	}
	gs.advanceTurn()
	gs.checkTermination()
}

// forfeit ends the game: the player loses.
func (gs *GameState) forfeit(player *Player, reason string) {
	gs.Loser = player.ID
	gs.terminate(reason)
}

func (gs *GameState) terminate(reason string) {
	gs.Terminated = true
	gs.TerminationReason = reason
}

// ended completes a TurnResult with the termination state and, if the game
// is over, its outcome.
func (gs *GameState) ended(r *TurnResult) *TurnResult {
	r.Terminated, r.TerminationReason = gs.Terminated, gs.TerminationReason
	if gs.Terminated {
		r.Outcome = gs.Outcome()
	}
	return r
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
	case va.Type == "question":
		id, err := gs.Adapter.RaiseQuestion(ownerID, va.TargetArgID, va.Locus.Kind, va.LocusStmtID,
			gs.Config.presumptionWeight(), caes.PE)
		if err != nil {
			return nil, nil, err
		}
		return []string{id}, nil, nil
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
	pen := gs.Config.Scoring.InvalidActionPenalty
	player.Score += pen
	player.Flat += pen
	gs.MoveHistory = append(gs.MoveHistory, Move{
		TurnNumber:  gs.CurrentTurn,
		PlayerID:    player.ID,
		Action:      action,
		ScoreDelta:  pen,
		ScoreDeltas: map[string]float64{player.ID: pen},
	})
	return &TurnResult{Accepted: false, Error: msg, ScoreDelta: pen, ScoreDeltas: map[string]float64{player.ID: pen}}
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

// computeScore returns the ScoringMove points of a scorable move: base, for
// the arguments it put forward, and flat, for its bonuses.
//
// Formula: score_delta = base_points[conclusion_label] × weight + flat_bonus
//
// Where:
//   - conclusion_label is the label of the new argument's conclusion AFTER evaluation
//   - weight is the scheme's base_weight
//   - flat_bonus is +UndercutterBonus if the undercutter flipped the target
//     from In/Undecided to Out, and +ProlepticBonus for a positive answer to
//     a critical question about one's own argument that nobody has raised
//
// The SRD's multipliers (scheme, consistency, chain, repetition) are not
// implemented.
func (gs *GameState) computeScore(
	argIDs []string,
	va *nascq.ValidatedAction,
	labelsBefore, labelsAfter map[string]caes.Label,
) (base, flat float64) {
	cfg := gs.Config.Scoring

	// Proleptic defence (SRD §16.4): a positive answer to a critical question
	// about one's own argument that nobody has raised yet.
	if va.IsDefence() && va.AnswersCQID != "" && cfg.ProlepticBonus != 0 &&
		gs.Adapter.ArgumentOwner(va.TargetArgID) == gs.ActivePlayer().ID && !gs.raised(va.TargetArgID, va.AnswersCQID) {
		flat += cfg.ProlepticBonus
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
		base += basePoints(label, cfg) * va.Weight
	}

	// Undercutter bonus: if the target argument's conclusion flipped to Out.
	if va.CaesAttackType == "undercutter" && va.TargetArgID != "" {
		targetArg := gs.Adapter.Graph.Arguments[va.TargetArgID]
		if targetArg != nil && targetArg.Conclusion != nil {
			cid := targetArg.Conclusion.Id
			before := labelsBefore[cid]
			after := labelsAfter[cid]
			if (before == caes.In || before == caes.Undecided) && after == caes.Out {
				flat += cfg.UndercutterBonus
			}
		}
	}

	return base, flat
}

// standing reports whether the argument is in force: it exists and its
// weight is above 0 (it is not undercut and its premises are In). Its
// conclusion may still be Out, outweighed or tied on an issue: attacking it
// then can change the verdict.
func (gs *GameState) standing(argID string) bool {
	arg := gs.Adapter.Graph.Arguments[argID]
	return arg != nil && arg.Weight > 0
}

// raised reports whether an accepted move has already raised the critical
// question about the argument: asked it, or answered it either way.
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
	defer func() { gs.TurnStartedAt = gs.now() }()
	for range gs.Players {
		gs.nextSeat()
		if !gs.ActivePlayer().Resigned {
			return
		}
	}
}

// nextSeat moves to the next seat, counting rounds.
func (gs *GameState) nextSeat() {
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

// checkTermination evaluates the termination conditions that hold between
// turns, in priority order, and returns the reason if the game ends (or has
// ended), "" otherwise. A forfeit (resignation, unanswered question, late
// move), the game clock and a concession end the game where they happen.
//
//  1. Target score reached (the highest score at or above it wins)
//  2. Fewer than two players left (the others resigned)
//  3. Max rounds reached
//  4. All-abstain threshold reached
func (gs *GameState) checkTermination() string {
	if gs.Terminated {
		return gs.TerminationReason
	}
	if t := gs.Config.TargetScore; t > 0 {
		for _, p := range gs.Players {
			if p.Score >= t {
				gs.terminate(fmt.Sprintf("target score %.0f reached", t))
				return gs.TerminationReason
			}
		}
	}
	active := 0
	for _, p := range gs.Players {
		if !p.Resigned {
			active++
		}
	}
	if len(gs.Players) > 1 && active < 2 {
		gs.terminate("the other players resigned")
		return gs.TerminationReason
	}

	if gs.Config.MaxTurns > 0 && gs.CurrentTurn > gs.Config.MaxTurns {
		gs.terminate(fmt.Sprintf("max rounds reached (%d)", gs.Config.MaxTurns))
		gs.Truncated = true
		return gs.TerminationReason
	}

	if gs.Config.AbstentionThreshold > 0 && gs.ConsecAllAbstain >= gs.Config.AbstentionThreshold {
		gs.terminate(fmt.Sprintf("all players abstained for %d consecutive rounds", gs.ConsecAllAbstain))
		return gs.TerminationReason
	}

	return ""
}
