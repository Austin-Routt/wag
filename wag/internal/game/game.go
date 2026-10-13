// Package game implements the Waltonian Argumentation Game (WAG) state
// machine. It wires together the Carneades adapter (argument evaluation)
// and the NASCQ layer (structural validation) into a playable turn-based
// game with scoring and termination conditions.
//
// Architecture (from the WAG plan, Part I §2):
//
//	Game Layer → turn sequencing, scoring, termination
//	  ↓ validates via
//	NASCQ Layer → structural move validation (S1–S8)
//	  ↓ translates via
//	Carneades Adapter → graph mutations + evaluation
//	  ↓ delegates to
//	Carneades-4 Engine → GroundedLabelling, proof standards
//
// This package is the top-level orchestrator. It imports both
// internal/carneades and internal/nascq but neither of those
// packages imports the other (no circular dependencies).
package game

import (
	"fmt"
	"time"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// -------------------------------------------------------------------------
// Configuration
// -------------------------------------------------------------------------

// Scoring models (ScoringConfig.Model).
const (
	// ScoringMove awards a move's points once, when it is made, from the
	// labels right after it: BasePoints[label of its conclusion] × weight,
	// plus the flat bonuses (the SRD's snapshot scoring). A later move that
	// defeats the argument does not take the points back.
	ScoringMove = "move"
	// ScoringLive recomputes every score after every move from the current
	// labels (the SRD's delta scoring): a player has, for each statement
	// they argued for, BasePoints[its label] × the weight of their best
	// argument for it that is in force, plus UndercutterBonus for each
	// argument their undercutters currently defeat, plus the flat points
	// (penalties and proleptic bonuses). Arguments that are defeated lose
	// their points, so the last mover does not keep points they no longer
	// deserve, and arguing the same statement again earns nothing more.
	ScoringLive = "live"
)

// ScoringConfig holds the scoring parameters.
//
// Formula: score_delta = BasePoints[label] × weight + flat_bonus, awarded
// once (ScoringMove) or kept up to date (ScoringLive).
type ScoringConfig struct {
	// Model is ScoringMove (the default, also for "") or ScoringLive.
	Model string

	BasePointsIn         float64 // points when conclusion is In (default: 10)
	BasePointsUndecided  float64 // points when conclusion is Undecided (default: 2)
	BasePointsOut        float64 // points when conclusion is Out (default: 0)
	UndercutterBonus     float64 // flat bonus for flipping target to Out (default: 15)
	InvalidActionPenalty float64 // penalty for structurally invalid move (default: -5)

	// ProlepticBonus is the SRD's flat bonus (§16.4) for answering a critical
	// question about one's own argument positively before anyone has raised
	// it (default: 0; the SRD suggests 5).
	ProlepticBonus float64
}

// DefaultScoringConfig returns the Phase 1 scoring parameters from the plan.
func DefaultScoringConfig() ScoringConfig {
	return ScoringConfig{
		BasePointsIn:         10.0,
		BasePointsUndecided:  2.0,
		BasePointsOut:        0.0,
		UndercutterBonus:     15.0,
		InvalidActionPenalty: -5.0,
	}
}

// Objectives (GameConfig.Objective).
const (
	// ObjectivePoints: the highest score wins (a draw if tied).
	ObjectivePoints = "points"
	// ObjectiveClaim: a debate about a central claim. Even seats argue for
	// it (pro, who bears the burden of proof), odd seats against it (con).
	// Pro wins if the claim is In at the end, con otherwise; a single player
	// argues both sides and gets the verdict only (solo belief revision).
	ObjectiveClaim = "claim"
)

// GameConfig holds all parameters for a game instance. The termination
// conditions combine: the game ends at the first one met (see
// checkTermination). ModeConfig returns the presets.
type GameConfig struct {
	// Mode is the name of the preset this configuration came from, for
	// display only.
	Mode string

	MaxTurns            int           // maximum number of rounds (every player moves once a round); 0 = unlimited
	AbstentionThreshold int           // consecutive all-abstain rounds to trigger end (default: 2)
	Scoring             ScoringConfig // scoring parameters

	// Objective is ObjectivePoints (the default, also for "") or
	// ObjectiveClaim.
	Objective string
	// Claim is the central claim of ObjectiveClaim, as a statement ID; ""
	// makes it the conclusion of pro's first argument.
	Claim string
	// ClaimStandard is the proof standard of the claim (the zero value is
	// PE): it is at issue from the start, so pro's arguments must meet it.
	// Carneades' CCE needs the best pro argument to outweigh the best con
	// argument by more than 0.5, and BRD also needs the con side below 0.3.
	ClaimStandard caes.Standard
	// ConcedeOnPass (ObjectiveClaim): a side that passes while the verdict
	// is against it concedes, as in persuasion dialogues (pro retracts a
	// claim it cannot carry; con concedes one it cannot shake).
	ConcedeOnPass bool

	// TargetScore > 0 ends the game when a player's score reaches it.
	TargetScore float64
	// TimeLimit > 0 ends the game when that much time has passed since it
	// began (the move made after it is not applied).
	TimeLimit time.Duration
	// MoveTimeLimit > 0 voids a move made later than that after the turn
	// began: the turn passes (an abstention), or with MoveTimeoutLoses the
	// player loses.
	MoveTimeLimit    time.Duration
	MoveTimeoutLoses bool
	// AnswerWithin > 0: a player who leaves a question about their argument
	// unanswered for that many of their own turns loses (critical-question
	// play). Questions are tracked, but not enforced, when it is 0.
	AnswerWithin int

	// AllowOther offers the user-defined scheme "Other" (free premises and
	// conclusion, generic critical questions; nascq.EnableOther).
	AllowOther bool
	// AnyPremise lets any premise be challenged ("why should we accept
	// that?"), not only the premises a scheme's own critical questions are
	// about (nascq.EnablePremiseQuestion). The modes turn it on. Any
	// statement can be argued for or against whatever this says.
	AnyPremise bool
	// Context names the context preset the game was set up from (Context),
	// and Weights are the scheme weights it sets (scheme ID, or "*" for
	// every scheme, to base weight).
	Context string
	Weights map[string]float64
	// RequireEvidence rejects (E1) an argument made with a scheme whose
	// evidence is "required" unless the move cites a source.
	RequireEvidence bool
	// Moderated holds every move for a moderator, who approves it (and may
	// set its weight) or rejects it, and who can accept or reject
	// statements (GameState.Moderate, GameState.Decide). Agents cannot play
	// the moderator: the runner and the CLI do not support it.
	Moderated bool

	// PresumptionWeight is the weight of the presumption that makes a move's
	// premises hold until attacked (see carneades.AddPresumption). Any argued
	// undermining (scheme weights are 0.55-0.70) outweighs it, and the mover
	// can reinstate the premise by backing it. 0 or less means
	// DefaultPresumptionWeight.
	PresumptionWeight float64
}

// DefaultPresumptionWeight is used when GameConfig.PresumptionWeight is not
// positive. It equals the presumption weight of the Kialo translator's primary
// profile.
const DefaultPresumptionWeight = 0.01

// EffectivePresumptionWeight is the weight of the presumptions that make
// premises hold until attacked: PresumptionWeight, or the default.
func (c GameConfig) EffectivePresumptionWeight() float64 { return c.presumptionWeight() }

func (c GameConfig) presumptionWeight() float64 {
	if c.PresumptionWeight > 0 {
		return c.PresumptionWeight
	}
	return DefaultPresumptionWeight
}

// DefaultGameConfig returns a sensible default configuration for testing.
func DefaultGameConfig() GameConfig {
	return GameConfig{
		MaxTurns:            20,
		AbstentionThreshold: 2,
		Scoring:             DefaultScoringConfig(),
	}
}

// -------------------------------------------------------------------------
// Player
// -------------------------------------------------------------------------

// Player represents a participant in the game.
type Player struct {
	ID                     string
	Name                   string
	Score                  float64
	SchemeUsageCounts      map[string]int // NASCQ scheme ID → times used
	ConsecutiveAbstentions int

	// Flat is the sum of the player's flat points (invalid-move penalties
	// and proleptic bonuses), which ScoringLive adds to the live score.
	Flat float64
	// Resigned players no longer move.
	Resigned bool
}

// newPlayer creates a Player with initialized maps.
func newPlayer(id, name string) *Player {
	return &Player{
		ID:                id,
		Name:              name,
		SchemeUsageCounts: make(map[string]int),
	}
}

// -------------------------------------------------------------------------
// Move (history record)
// -------------------------------------------------------------------------

// Move records a single turn's action and its effects on the game state.
// Stored in GameState.MoveHistory for replay, delta scoring, and debugging.
type Move struct {
	TurnNumber   int
	PlayerID     string
	Action       nascq.ProposedAction
	Abstained    bool
	ArgIDs       []string              // argument IDs created (empty if abstained)
	Presumptions []string              // presumption arguments created for the move's premises
	Weight       float64               // scheme base_weight used
	ScoreDelta   float64               // points awarded this turn
	LabelsBefore map[string]caes.Label // statement labels before the move
	LabelsAfter  map[string]caes.Label // statement labels after the move

	// ScoreDeltas is every player's score change through this move: the
	// mover's only under ScoringMove; anyone's under ScoringLive.
	ScoreDeltas map[string]float64
	// Scorable: the move's arguments can earn points (it is not an attack on
	// the mover's own argument or on one already defeated).
	Scorable bool
	TimedOut bool // a void late move, recorded as an abstention
	Resigned bool
	// Rejection is why an invalid move was rejected.
	Rejection string `json:",omitempty"`

	// Moderation is the moderator's verdict on a held move ("approved",
	// "rejected"), with their note.
	Moderation string
	// Decision and DecisionStmt record a moderator's decision on a statement
	// (Action.Type "decide", PlayerID ModeratorID).
	Decision     string
	DecisionStmt string
}

// ModeratorID is the PlayerID of the moderator's records in the history.
const ModeratorID = "moderator"

// PendingMove is a move held for the moderator.
type PendingMove struct {
	PlayerID    string               `json:"player"`
	Action      nascq.ProposedAction `json:"action"`
	SubmittedAt time.Time            `json:"submitted_at"`
}

// -------------------------------------------------------------------------
// TurnResult (returned to caller after each turn)
// -------------------------------------------------------------------------

// TurnResult communicates the outcome of ExecuteTurn to the caller.
type TurnResult struct {
	Accepted          bool     `json:"accepted"`                     // true if the move was valid and applied
	Abstained         bool     `json:"abstained,omitempty"`          // true if the player abstained
	ArgIDs            []string `json:"arg_ids,omitempty"`            // IDs of arguments created
	Presumptions      []string `json:"presumptions,omitempty"`       // IDs of the presumption arguments created for the move's premises
	ScoreDelta        float64  `json:"score_delta"`                  // points awarded this turn
	Error             string   `json:"error,omitempty"`              // non-empty if the move was rejected
	Terminated        bool     `json:"terminated,omitempty"`         // true if the game ended after this turn
	TerminationReason string   `json:"termination_reason,omitempty"` // why the game ended

	ScoreDeltas map[string]float64 `json:"score_deltas,omitempty"` // every player's score change (see Move.ScoreDeltas)
	TimedOut    bool               `json:"timed_out,omitempty"`    // the move came too late and was void
	Pending     bool               `json:"pending,omitempty"`      // the move is held for the moderator (GameConfig.Moderated)
	Outcome     *Outcome           `json:"outcome,omitempty"`      // the result, once the game has ended
}

// -------------------------------------------------------------------------
// GameState
// -------------------------------------------------------------------------

// GameState is the core game object. It owns the Carneades adapter, the
// NASCQ graph, and all game-level state (players, turns, scores, history).
//
// Thread safety: GameState is designed for single-threaded use (Phase 1 CLI).
// For Phase 3+ (concurrent web), add a mutex around ExecuteTurn.
type GameState struct {
	Config  GameConfig
	Players []*Player // turn order is Players[0], Players[1], ...

	Adapter *carneades.Adapter // argument graph + evaluation engine
	NASCQ   *nascq.NASCQ       // structural validation graph
	bridge  *graphBridge       // implements nascq.GraphQuerier

	// dataWeights are the schemes' base weights as the data gives them,
	// before a context or the game's own weights changed them (ClaimUnder).
	dataWeights map[string]float64

	CurrentTurn       int // 1-indexed turn number
	ActivePlayerIdx   int // index into Players slice
	MoveHistory       []Move
	ConsecAllAbstain  int  // consecutive ROUNDS where all players abstained
	roundHadMove      bool // a move was accepted in the current round
	Terminated        bool
	TerminationReason string
	// Truncated: the game was cut off by a limit (rounds or time) rather
	// than ended by play (gym's truncation).
	Truncated bool

	// LastRejection is why the last move was rejected, or "" if it was
	// accepted. The player is still to move after a rejection, so an agent
	// can read it (it is part of the JSON View) and try again.
	LastRejection string

	// Claim is the central claim of ObjectiveClaim, once set.
	Claim string
	// Questions are the critical questions asked without an argument
	// (question moves), answered or not, in the order asked.
	Questions []OpenQuestion
	// Loser is the player who forfeited the game (resigned, left a question
	// unanswered or ran out of time), if any.
	Loser string

	// Pending is the move held for the moderator, if any: no other move can
	// be made until the moderator approves or rejects it.
	Pending *PendingMove

	moderating     bool    // a held move is being applied
	weightOverride float64 // > 0: the weight the moderator gave the move being applied

	clock         func() time.Time
	StartedAt     time.Time // when the game began (SetClock resets it)
	TurnStartedAt time.Time // when the current turn began
}

// OpenQuestion is a critical question asked without an argument: the
// target's proponent must answer it (a defence with the same question).
type OpenQuestion struct {
	ArgID       string // the question argument (carneades.RoleQuestion)
	TargetArgID string
	CQID        string
	LocusStmtID string // the premise asked about; "" for the inference
	Asker       string
	Owner       string // the proponent of the target, who must answer
	AskedAt     int    // index of the question move in MoveHistory
	AnsweredBy  string // the argument that answered it; "" while open
	TurnsLeft   int    // the owner's turns left to answer (GameConfig.AnswerWithin)
}

// Open reports whether the question still awaits an answer.
func (q OpenQuestion) Open() bool { return q.AnsweredBy == "" }

// NewGame creates a game state from the given configuration, adapter, NASCQ,
// and player names. The adapter should already have schemes loaded (via YAML
// or programmatic setup). The NASCQ should already have schemes, CQs, and
// edges loaded.
//
// Player order: players take turns in the order provided. Players[0] goes first.
func NewGame(config GameConfig, adapter *carneades.Adapter, n *nascq.NASCQ, playerNames []string) *GameState {
	// A player's ID is their name, made unique: two players called "A" used
	// to share an ID, so each one's attacks on the other counted as
	// self-attacks and every score went to the first.
	players := make([]*Player, len(playerNames))
	used := map[string]bool{}
	for i, name := range playerNames {
		id := name
		for n := 2; used[id]; n++ {
			id = fmt.Sprintf("%s#%d", name, n)
		}
		used[id] = true
		players[i] = newPlayer(id, name)
	}

	gs := &GameState{
		Config:      config,
		Players:     players,
		Adapter:     adapter,
		NASCQ:       n,
		CurrentTurn: 1,
	}
	gs.bridge = &graphBridge{gs: gs}
	gs.SetClock(nil)
	if config.AllowOther {
		n.EnableOther()
	}
	if config.AnyPremise {
		n.EnablePremiseQuestion()
	}
	gs.dataWeights = map[string]float64{}
	for id, sc := range n.Schemes {
		gs.dataWeights[id] = sc.BaseWeight
	}
	if len(config.Weights) > 0 {
		gs.applyWeights(config.Weights)
	}
	if config.Objective == ObjectiveClaim && config.Claim != "" {
		if id, err := adapter.MakeIssue(config.Claim, config.ClaimStandard); err == nil {
			gs.Claim = id
			adapter.Evaluate()
		}
	}
	return gs
}

// SetClock sets the game's time source (nil: time.Now) and restarts the game
// and turn clocks from it. Tests inject a fake clock.
func (gs *GameState) SetClock(now func() time.Time) {
	if now == nil {
		now = time.Now
	}
	gs.clock = now
	gs.StartedAt = now()
	gs.TurnStartedAt = gs.StartedAt
}

func (gs *GameState) now() time.Time { return gs.clock() }

func (c GameConfig) live() bool { return c.Scoring.Model == ScoringLive }

// Side returns "pro" or "con" for a player of an ObjectiveClaim game with two
// or more players (even seats are pro), and "" otherwise.
func (gs *GameState) Side(playerID string) string {
	if gs.Config.Objective != ObjectiveClaim || len(gs.Players) < 2 {
		return ""
	}
	for i, p := range gs.Players {
		if p.ID == playerID {
			if i%2 == 0 {
				return "pro"
			}
			return "con"
		}
	}
	return ""
}

// ActivePlayer returns the player whose turn it currently is.
func (gs *GameState) ActivePlayer() *Player {
	return gs.Players[gs.ActivePlayerIdx]
}

// PlayerByID returns the player with the given ID, or nil if not found.
func (gs *GameState) PlayerByID(id string) *Player {
	for _, p := range gs.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// -------------------------------------------------------------------------
// GraphQuerier bridge
// -------------------------------------------------------------------------

// graphBridge adapts the carneades.Adapter to the nascq.GraphQuerier
// interface. This is the glue between the NASCQ validation layer (which
// needs to inspect the argument graph) and the Carneades adapter (which
// owns the graph).
//
// Thread safety: This bridge accesses adapter.Graph directly without
// locking. This is safe for Phase 1 (single-threaded CLI). For Phase 3+,
// add locked accessor methods to the adapter.
type graphBridge struct {
	gs *GameState
}

// Compile-time check that graphBridge implements nascq.GraphQuerier.
var _ nascq.GraphQuerier = (*graphBridge)(nil)

// ArgumentExists returns true if an argument with the given ID exists in
// the Carneades argument graph.
func (b *graphBridge) ArgumentExists(argID string) bool {
	return b.gs.Adapter.Graph.Arguments[argID] != nil
}

// ArgumentSchemeID returns the Carneades scheme ID for the given argument.
// Returns "" if the argument doesn't exist or has no scheme.
func (b *graphBridge) ArgumentSchemeID(argID string) string {
	arg := b.gs.Adapter.Graph.Arguments[argID]
	if arg == nil || arg.Scheme == nil {
		return ""
	}
	return arg.Scheme.Id
}

// PremiseExists returns true if the argument has a premise with the given
// statement ID.
func (b *graphBridge) PremiseExists(argID, premiseStmtID string) bool {
	arg := b.gs.Adapter.Graph.Arguments[argID]
	if arg == nil {
		return false
	}
	for _, p := range arg.Premises {
		if p.Stmt != nil && p.Stmt.Id == premiseStmtID {
			return true
		}
	}
	return false
}

// PremiseMatching returns the premise of the argument that a CQ's
// target_premise_pattern designates (see carneades.Adapter.PremiseMatching).
func (b *graphBridge) PremiseMatching(argID, pattern string) string {
	return b.gs.Adapter.PremiseMatching(argID, pattern)
}

// Bridge returns the graph querier that NASCQ validation needs for this game.
func (gs *GameState) Bridge() nascq.GraphQuerier { return gs.bridge }

// PremiseIDs returns the statement IDs of the argument's premises, in order.
func (b *graphBridge) PremiseIDs(argID string) []string {
	arg := b.gs.Adapter.Graph.Arguments[argID]
	if arg == nil {
		return nil
	}
	out := make([]string, 0, len(arg.Premises))
	for _, p := range arg.Premises {
		if p.Stmt != nil {
			out = append(out, p.Stmt.Id)
		}
	}
	return out
}

// ExceptionMatching returns the exception of the argument that an exception
// pattern designates (see carneades.Adapter.ExceptionMatching).
func (b *graphBridge) ExceptionMatching(argID, pattern string) string {
	return b.gs.Adapter.ExceptionMatching(argID, pattern)
}

// ArgumentConclusion returns the statement ID of the argument's conclusion.
func (b *graphBridge) ArgumentConclusion(argID string) string {
	if arg := b.gs.Adapter.Graph.Arguments[argID]; arg != nil && arg.Conclusion != nil {
		return arg.Conclusion.Id
	}
	return ""
}

// Statement describes a statement for a move about it (see
// carneades.Adapter.StatementUse).
func (b *graphBridge) Statement(stmtID string) nascq.StatementInfo {
	exists, premise, concluded, opposite := b.gs.Adapter.StatementUse(stmtID)
	return nascq.StatementInfo{Exists: exists, Premise: premise, Concluded: concluded, Opposite: opposite}
}

// Fits returns the values a scheme's conclusion variables take to conclude
// a statement (see carneades.Adapter.Fits).
func (b *graphBridge) Fits(schemeID, stmtID string, given map[string]string) (map[string]string, error) {
	return b.gs.Adapter.Fits(schemeID, stmtID, given)
}
