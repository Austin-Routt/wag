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
	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// -------------------------------------------------------------------------
// Configuration
// -------------------------------------------------------------------------

// ScoringConfig holds the Phase 1 scoring parameters.
//
// Phase 1 formula: score_delta = BasePoints[label] × weight + flat_bonus
// Phase 2 adds multipliers and retroactive delta scoring (not implemented here).
type ScoringConfig struct {
	BasePointsIn         float64 // points when conclusion is In (default: 10)
	BasePointsUndecided  float64 // points when conclusion is Undecided (default: 2)
	BasePointsOut        float64 // points when conclusion is Out (default: 0)
	UndercutterBonus     float64 // flat bonus for flipping target to Out (default: 15)
	InvalidActionPenalty float64 // penalty for structurally invalid move (default: -5)
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

// GameConfig holds all parameters for a game instance.
type GameConfig struct {
	MaxTurns            int           // 0 = unlimited
	AbstentionThreshold int           // consecutive all-abstain rounds to trigger end (default: 2)
	Scoring             ScoringConfig // scoring parameters

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
	ArgIDs       []string            // argument IDs created (empty if abstained)
	Presumptions []string            // presumption arguments created for the move's premises
	Weight       float64             // scheme base_weight used
	ScoreDelta   float64             // points awarded this turn
	LabelsBefore map[string]caes.Label // statement labels before the move
	LabelsAfter  map[string]caes.Label // statement labels after the move
}

// -------------------------------------------------------------------------
// TurnResult (returned to caller after each turn)
// -------------------------------------------------------------------------

// TurnResult communicates the outcome of ExecuteTurn to the caller.
type TurnResult struct {
	Accepted          bool     // true if the move was valid and applied
	Abstained         bool     // true if the player abstained
	ArgIDs            []string // IDs of arguments created
	Presumptions      []string // IDs of the presumption arguments created for the move's premises
	ScoreDelta        float64  // points awarded this turn
	Error             string   // non-empty if the move was rejected
	Terminated        bool     // true if the game ended after this turn
	TerminationReason string   // why the game ended
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
	bridge  *graphBridge        // implements nascq.GraphQuerier

	CurrentTurn     int  // 1-indexed turn number
	ActivePlayerIdx int  // index into Players slice
	MoveHistory     []Move
	ConsecAllAbstain int  // consecutive ROUNDS where all players abstained
	Terminated      bool
	TerminationReason string
}

// NewGame creates a game state from the given configuration, adapter, NASCQ,
// and player names. The adapter should already have schemes loaded (via YAML
// or programmatic setup). The NASCQ should already have schemes, CQs, and
// edges loaded.
//
// Player order: players take turns in the order provided. Players[0] goes first.
func NewGame(config GameConfig, adapter *carneades.Adapter, n *nascq.NASCQ, playerNames []string) *GameState {
	players := make([]*Player, len(playerNames))
	for i, name := range playerNames {
		players[i] = newPlayer(name, name) // use name as both ID and display name
	}

	gs := &GameState{
		Config:      config,
		Players:     players,
		Adapter:     adapter,
		NASCQ:       n,
		CurrentTurn: 1,
	}
	gs.bridge = &graphBridge{gs: gs}
	return gs
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
