package game

import (
	"math/rand"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
)

// Agent is the interface for any entity that plays the WAG game.
// Both AI agents and human input adapters implement this interface.
//
// This follows the RL-style pattern: the agent receives an observation
// (game state + available actions) and returns an action to take.
type Agent interface {
	// Name returns the agent's display name.
	Name() string

	// ChooseAction selects a move given the current game state and the
	// list of structurally valid action templates. The agent must fill in
	// variable bindings and return a complete ProposedAction.
	//
	// If the agent wants to abstain, it should return a ProposedAction
	// with Type == "abstain".
	//
	// The agent may be called multiple times per turn if its chosen action
	// fails validation (the runner retries up to a limit).
	ChooseAction(obs *Observation) nascq.ProposedAction
}

// Observation is the RL-style observation passed to agents on each turn.
// It contains everything an agent needs to make a decision.
type Observation struct {
	GameState        *GameState             // full game state (read-only by convention)
	AvailableActions []nascq.ActionTemplate // structurally valid action skeletons
	ActivePlayer     *Player                // whose turn it is
	TurnNumber       int
	Scores           map[string]float64 // playerID → score
}

// -------------------------------------------------------------------------
// Word pool for random binding generation
// -------------------------------------------------------------------------

// wordPool provides domain-relevant words for random variable filling.
// Each semantic type has its own word list to produce somewhat coherent
// arguments (though still random).
var wordPool = map[string][]string{
	"action": {
		"carbon_tax", "solar_subsidies", "wind_investment",
		"regulation", "cap_and_trade", "reforestation",
		"nuclear_energy", "public_transit", "ev_mandate",
		"emission_ban", "recycling_program", "green_deal",
	},
	"state": {
		"current_crisis", "status_quo", "recession",
		"growth_period", "transition_phase", "emergency",
	},
	"goal": {
		"reduce_emissions", "economic_growth", "public_health",
		"energy_independence", "job_creation", "innovation",
		"biodiversity", "clean_air", "food_security",
	},
	"value": {
		"sustainability", "freedom", "equality",
		"prosperity", "justice", "security",
		"environmental_health", "human_dignity", "efficiency",
	},
	"event": {
		"temperature_rise", "sea_level_rise", "drought",
		"flooding", "crop_failure", "economic_downturn",
		"tech_breakthrough", "policy_change", "market_shift",
	},
	"entity": {
		"government", "ipcc", "world_bank",
		"epa", "who", "eu_commission",
		"harvard_study", "nasa", "un_panel",
	},
	"domain": {
		"climate_science", "economics", "public_health",
		"energy_policy", "environmental_law", "agriculture",
	},
	"proposition": {
		"warming_is_real", "carbon_causes_warming", "tax_reduces_emissions",
		"renewables_are_viable", "action_is_urgent", "costs_are_manageable",
	},
	"case": {
		"eu_carbon_market", "sweden_carbon_tax", "california_cap_trade",
		"china_solar_boom", "germany_energiewende", "norway_ev_policy",
	},
	// Fallback for any unrecognized type.
	"default": {
		"alpha", "beta", "gamma", "delta", "epsilon",
		"zeta", "eta", "theta", "iota", "kappa",
	},
}

// randomWord picks a random word from the pool for the given variable type.
func randomWord(rng *rand.Rand, varType string) string {
	words, ok := wordPool[varType]
	if !ok {
		words = wordPool["default"]
	}
	return words[rng.Intn(len(words))]
}

// -------------------------------------------------------------------------
// RandomAgent
// -------------------------------------------------------------------------

// RandomAgent picks a random action template and fills in random variable
// bindings from the word pool. It has a configurable abstain probability.
//
// This agent is designed for stress testing: if it can play hundreds of
// games without panics, the engine is robust.
type RandomAgent struct {
	name        string
	seed        int64
	rng         *rand.Rand
	abstainProb float64 // probability of abstaining (0.0-1.0)
	supportBias float64 // probability of choosing support over attack (0.0-1.0)
}

// NewRandomAgent creates a RandomAgent with the given random seed.
// abstainProb controls how often the agent abstains (e.g., 0.1 = 10%).
// supportBias controls preference for support vs attack (e.g., 0.5 = equal).
func NewRandomAgent(name string, seed int64, abstainProb, supportBias float64) *RandomAgent {
	return &RandomAgent{
		name:        name,
		seed:        seed,
		rng:         rand.New(rand.NewSource(seed)),
		abstainProb: abstainProb,
		supportBias: supportBias,
	}
}

func (a *RandomAgent) Name() string { return a.name }

// Seed returns the seed the agent was created with (Seeded).
func (a *RandomAgent) Seed() int64 { return a.seed }

func (a *RandomAgent) ChooseAction(obs *Observation) nascq.ProposedAction {
	// Randomly abstain.
	if a.rng.Float64() < a.abstainProb {
		return nascq.ProposedAction{Type: "abstain"}
	}

	if len(obs.AvailableActions) == 0 {
		return nascq.ProposedAction{Type: "abstain"}
	}

	// Filter by type preference.
	var candidates []nascq.ActionTemplate
	for _, tmpl := range obs.AvailableActions {
		if tmpl.Type == "support" && a.rng.Float64() < a.supportBias {
			candidates = append(candidates, tmpl)
		} else if tmpl.Type == "attack" && a.rng.Float64() >= a.supportBias {
			candidates = append(candidates, tmpl)
		}
	}
	if len(candidates) == 0 {
		// Fallback: pick from all available.
		candidates = obs.AvailableActions
	}

	// Pick a random action template.
	tmpl := candidates[a.rng.Intn(len(candidates))]

	// Fill in random variable bindings, in sorted variable order so that the
	// random draws do not depend on map iteration order.
	bindings := make(map[string]string)
	for _, varName := range carneades.SortedKeys(tmpl.Variables) {
		bindings[varName] = randomWord(a.rng, tmpl.Variables[varName].Type)
	}

	action := nascq.ProposedAction{
		Type:             tmpl.Type,
		AttackType:       tmpl.AttackType,
		SchemeID:         tmpl.SchemeID,
		AnswersCQID:      tmpl.AnswersCQID,
		TargetArgID:      tmpl.TargetArgID,
		TargetPremiseID:  tmpl.TargetPremiseID,
		VariableBindings: bindings,
	}

	return action
}

// -------------------------------------------------------------------------
// GreedyAgent (tries each action, picks highest score delta)
// -------------------------------------------------------------------------

// GreedyAgent evaluates each available action template by simulating it
// with random bindings, then picks the one that produces the highest
// score delta. Uses RemoveArgument for rollback after each trial.
//
// This is a simple 1-ply lookahead. For stress testing purposes, it
// exercises the rollback mechanism heavily.
type GreedyAgent struct {
	name            string
	seed            int64
	rng             *rand.Rand
	trialsPerAction int // how many random binding sets to try per template
}

// NewGreedyAgent creates a GreedyAgent with the given seed.
func NewGreedyAgent(name string, seed int64, trialsPerAction int) *GreedyAgent {
	if trialsPerAction < 1 {
		trialsPerAction = 1
	}
	return &GreedyAgent{
		name:            name,
		seed:            seed,
		rng:             rand.New(rand.NewSource(seed)),
		trialsPerAction: trialsPerAction,
	}
}

func (a *GreedyAgent) Name() string { return a.name }

// Seed returns the seed the agent was created with (Seeded).
func (a *GreedyAgent) Seed() int64 { return a.seed }

func (a *GreedyAgent) ChooseAction(obs *Observation) nascq.ProposedAction {
	if len(obs.AvailableActions) == 0 {
		return nascq.ProposedAction{Type: "abstain"}
	}

	// For each action template, generate random bindings and evaluate.
	bestScore := -1000.0
	var bestAction nascq.ProposedAction
	found := false

	for _, tmpl := range obs.AvailableActions {
		for trial := 0; trial < a.trialsPerAction; trial++ {
			bindings := make(map[string]string)
			for _, varName := range carneades.SortedKeys(tmpl.Variables) {
				bindings[varName] = randomWord(a.rng, tmpl.Variables[varName].Type)
			}

			action := nascq.ProposedAction{
				Type:             tmpl.Type,
				AttackType:       tmpl.AttackType,
				SchemeID:         tmpl.SchemeID,
				AnswersCQID:      tmpl.AnswersCQID,
				TargetArgID:      tmpl.TargetArgID,
				TargetPremiseID:  tmpl.TargetPremiseID,
				VariableBindings: bindings,
			}

			// Validate structurally (cheap check before simulation).
			_, err := obs.GameState.NASCQ.Validate(action, obs.GameState.bridge)
			if err != nil {
				continue // invalid, skip
			}

			// Use the validated weight as a heuristic score estimate.
			// Full simulation would require snapshot/rollback which is
			// too expensive for a simple heuristic agent.
			scheme := obs.GameState.NASCQ.GetScheme(tmpl.SchemeID)
			score := scheme.BaseWeight * 10.0 // rough estimate
			if tmpl.Type == "attack" && tmpl.CaesAttackType == "undercutter" {
				score += 15.0 // anticipate undercutter bonus
			}

			if score > bestScore {
				bestScore = score
				bestAction = action
				found = true
			}
		}
	}

	if !found {
		return nascq.ProposedAction{Type: "abstain"}
	}
	return bestAction
}

// -------------------------------------------------------------------------
// Helper: build Observation for agents
// -------------------------------------------------------------------------

// BuildObservation constructs the RL-style observation for the current
// player. It enumerates available actions from the NASCQ and collects
// game state information.
func (gs *GameState) BuildObservation() *Observation {
	// Collect argument IDs for action enumeration.
	argIDs := make([]string, 0, len(gs.Adapter.Graph.Arguments))
	for id, arg := range gs.Adapter.Graph.Arguments {
		if carneades.IsAttackable(arg) { // exception and presumption arguments are never targets
			argIDs = append(argIDs, id)
		}
	}
	carneades.SortArgumentIDs(argIDs) // the order of the enumerated actions must not depend on map order

	// Enumerate available action templates.
	actions := gs.NASCQ.EnumerateActions(gs.bridge, argIDs)

	// Collect scores.
	scores := make(map[string]float64, len(gs.Players))
	for _, p := range gs.Players {
		scores[p.ID] = p.Score
	}

	return &Observation{
		GameState:        gs,
		AvailableActions: actions,
		ActivePlayer:     gs.ActivePlayer(),
		TurnNumber:       gs.CurrentTurn,
		Scores:           scores,
	}
}
