package game

import (
	"math/rand"
	"sort"
	"strings"

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
		"epa", "world_health_org", "eu_commission",
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

// neutralWords are what a bot says when the board has no word of the kind it
// needs: placeholders that fit any topic. A bot has no grasp of language;
// these at least do not drag in a topic nobody is discussing.
var neutralWords = map[string][]string{
	"entity":      {"an expert", "a researcher", "a witness", "a reviewer"},
	"domain":      {"the subject at hand", "this field"},
	"action":      {"the proposal", "the alternative"},
	"state":       {"the present situation", "the situation afterwards"},
	"goal":        {"the stated goal", "a better outcome"},
	"value":       {"the common good", "fairness", "safety"},
	"event":       {"a related event", "the observed sign"},
	"proposition": {"the opposite claim", "a further reason"},
	"case":        {"a comparable case", "an earlier case"},
	"default":     {"something else"},
}

// vocabulary is what a bot fills variables with: the words the players have
// already used for each kind of variable, then neutral placeholders. On an
// empty board (a bot opening the game) it is the example word pool.
type vocabulary map[string][]string

// vocabulary collects the words of the accepted moves, by variable type, in
// the order they were played.
func (gs *GameState) vocabulary() vocabulary {
	v := vocabulary{}
	seen := map[string]bool{}
	for _, m := range gs.MoveHistory {
		sc := gs.NASCQ.GetScheme(m.Action.SchemeID)
		if len(m.ArgIDs) == 0 || sc == nil {
			continue
		}
		for _, name := range carneades.SortedKeys(m.Action.VariableBindings) {
			spec, ok := sc.Variables[name]
			w := strings.TrimSpace(m.Action.VariableBindings[name])
			if !ok || w == "" || strings.ContainsAny(w, "()¬") || seen[spec.Type+"|"+w] {
				continue // an answer bound to a statement is not a word
			}
			seen[spec.Type+"|"+w] = true
			v[spec.Type] = append(v[spec.Type], w)
		}
	}
	return v
}

// word draws a word for a variable of type typ. Sources are always neutral:
// a bot that cited the other side's witness as its own expert would read as
// nonsense.
func (v vocabulary) word(rng *rand.Rand, typ string) string {
	if len(v) == 0 {
		return randomWord(rng, typ)
	}
	neutral, ok := neutralWords[typ]
	if !ok {
		neutral = neutralWords["default"]
	}
	var pool []string
	if typ != "entity" {
		pool = append(append(pool, v[typ]...), v[typ]...) // the board's words twice as likely
	}
	pool = append(pool, neutral...)
	return pool[rng.Intn(len(pool))]
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
	// In critical-question play, a question left unanswered loses: answer
	// one.
	if answers := dueAnswers(obs); len(answers) > 0 && obs.GameState.Config.AnswerWithin > 0 {
		return fill(a.rng, answers[a.rng.Intn(len(answers))], obs.GameState.vocabulary())
	}

	// Randomly abstain.
	if a.rng.Float64() < a.abstainProb {
		return nascq.ProposedAction{Type: "abstain"}
	}

	if len(obs.AvailableActions) == 0 {
		return nascq.ProposedAction{Type: "abstain"}
	}

	// Filter by type preference: arguments and defences, or attacks and
	// questions.
	var candidates []nascq.ActionTemplate
	for _, tmpl := range obs.AvailableActions {
		if tmpl.Type == "support" && a.rng.Float64() < a.supportBias {
			candidates = append(candidates, tmpl)
		} else if tmpl.Type != "support" && a.rng.Float64() >= a.supportBias {
			candidates = append(candidates, tmpl)
		}
	}
	if len(candidates) == 0 {
		// Fallback: pick from all available.
		candidates = obs.AvailableActions
	}

	// Pick a random action template.
	return fill(a.rng, candidates[a.rng.Intn(len(candidates))], obs.GameState.vocabulary())
}

// fill makes an action of the template with random variable bindings, drawn
// in sorted variable order so that the draws do not depend on map iteration
// order. The variables the game binds (Bound) are left to it.
func fill(rng *rand.Rand, tmpl nascq.ActionTemplate, words vocabulary) nascq.ProposedAction {
	var bindings map[string]string
	if len(tmpl.Variables) > 0 {
		bindings = make(map[string]string, len(tmpl.Variables))
		for _, varName := range carneades.SortedKeys(tmpl.Variables) {
			if !tmpl.Bound(varName) {
				bindings[varName] = words.word(rng, tmpl.Variables[varName].Type)
			}
		}
	}
	return tmpl.Action(bindings)
}

// dueAnswers returns the templates that answer an open question about one of
// the active player's arguments.
func dueAnswers(obs *Observation) []nascq.ActionTemplate {
	var out []nascq.ActionTemplate
	for _, q := range obs.GameState.Questions {
		if !q.Open() || q.Owner != obs.ActivePlayer.ID {
			continue
		}
		for _, t := range obs.AvailableActions {
			if t.Type == "support" && t.TargetArgID == q.TargetArgID && t.AnswersCQID == q.CQID {
				out = append(out, t)
			}
		}
	}
	return out
}

// -------------------------------------------------------------------------
// GreedyAgent (tries each action, picks highest score delta)
// -------------------------------------------------------------------------

// GreedyAgent looks one move ahead: it tries the available action templates
// (up to MaxCandidates of them, plus every answer to a question it owes)
// with trialsPerAction sets of random bindings, and passing, each through
// GameState.SimulateThen (which restores the game afterwards), and plays the
// one that leaves it best placed (Value). Equally good moves are chosen
// between at random (with the agent's seed, so a seeded agent is
// reproducible); it passes only when no move is better than passing.
//
// It used to score templates by a fixed estimate (base weight × 10, +15 for
// an undercutter) without trying them, so it preferred undercutters whether
// or not they would succeed or score; then by the mover's score delta, which
// under live scoring misses what a move takes from the others.
type GreedyAgent struct {
	name            string
	seed            int64
	rng             *rand.Rand
	trialsPerAction int // how many random binding sets to try per template

	// MaxCandidates caps the templates tried per turn (0: all): when there
	// are more, a random sample of that many is tried, in enumeration
	// order. Every argument offers attacks and defences for each of its
	// critical questions and backup schemes, so the enumeration grows
	// quickly with the graph.
	MaxCandidates int
}

// DefaultGreedyCandidates is NewGreedyAgent's MaxCandidates.
const DefaultGreedyCandidates = 40

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
		MaxCandidates:   DefaultGreedyCandidates,
	}
}

func (a *GreedyAgent) Name() string { return a.name }

// candidates returns the templates to try: all of them, or a random sample
// of MaxCandidates in enumeration order.
func (a *GreedyAgent) candidates(all []nascq.ActionTemplate) []nascq.ActionTemplate {
	if a.MaxCandidates <= 0 || len(all) <= a.MaxCandidates {
		return all
	}
	idx := a.rng.Perm(len(all))[:a.MaxCandidates]
	sort.Ints(idx)
	out := make([]nascq.ActionTemplate, len(idx))
	for i, j := range idx {
		out[i] = all[j]
	}
	return out
}

// Seed returns the seed the agent was created with (Seeded).
func (a *GreedyAgent) Seed() int64 { return a.seed }

func (a *GreedyAgent) ChooseAction(obs *Observation) nascq.ProposedAction {
	me := obs.ActivePlayer.ID
	gs := obs.GameState
	var best nascq.ProposedAction
	bestValue, found, ties := 0.0, false, 0
	try := func(action nascq.ProposedAction) {
		value := 0.0
		r, err := gs.SimulateThen(action, func(after *GameState) { value = Value(after, me) })
		if err != nil || r == nil || (!r.Accepted && !r.Abstained) {
			return
		}
		switch {
		case !found || value > bestValue+1e-9:
			best, bestValue, found, ties = action, value, true, 1
		case value >= bestValue-1e-9 && action.Type != "abstain":
			// An equally good move: each of the ties is as likely to be
			// played (reservoir sampling), so the bot does not answer every
			// argument with the first question it enumerates.
			ties++
			if a.rng.Intn(ties) == 0 {
				best = action
			}
		}
	}
	words := gs.vocabulary()
	tmpls := a.candidates(obs.AvailableActions)
	if due := dueAnswers(obs); len(due) > 0 {
		tmpls = append(append([]nascq.ActionTemplate{}, due...), tmpls...)
	}
	for _, tmpl := range tmpls {
		trials := a.trialsPerAction
		if len(tmpl.Variables) == 0 {
			trials = 1 // a question: nothing to vary
		}
		for trial := 0; trial < trials; trial++ {
			try(fill(a.rng, tmpl, words))
		}
	}
	try(nascq.ProposedAction{Type: "abstain"})
	if !found {
		return nascq.ProposedAction{Type: "abstain"}
	}
	return best
}

// Value is how well placed the player is: a won game is worth 1e6 and a lost
// one -1e6; otherwise their lead in score over the best of the others, plus
// 100 in a debate while the verdict favours their side.
func Value(gs *GameState, playerID string) float64 {
	if gs.Terminated {
		o := gs.Outcome()
		switch {
		case o.Winner == playerID:
			return 1e6
		case o.Winner != "" || o.Loser == playerID:
			return -1e6
		}
	}
	mine, others, seen := 0.0, 0.0, false
	for _, p := range gs.Players {
		if p.ID == playerID {
			mine = p.Score
		} else if !seen || p.Score > others {
			others, seen = p.Score, true
		}
	}
	v := mine - others
	if gs.Config.Objective == ObjectiveClaim && len(gs.Players) > 1 && gs.verdictFavours(playerID) {
		v += 100
	}
	return v
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

	// Enumerate available action templates: the answers to critical
	// questions, and the arguments for and against the statements on the
	// board.
	actions := gs.NASCQ.EnumerateActions(gs.bridge, argIDs)
	actions = append(actions, gs.NASCQ.EnumerateStatementActions(gs.bridge, gs.BoardStatements())...)

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
