package game

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/carneades"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// ModeInfo describes a preset of ModeConfig.
type ModeInfo struct {
	Name        string
	Description string
}

// Modes lists the presets of ModeConfig. Each is a combination of the
// orthogonal settings of GameConfig (objective, scoring model, termination
// conditions), which can also be set one by one.
func Modes() []ModeInfo {
	return []ModeInfo{
		{"classic", "points, scored once per move; 20 rounds or two all-pass rounds"},
		{"infinite", "points, live scoring; no round limit: ends when a player resigns or all pass twice"},
		{"target", "points, live scoring; the first to reach the target score (default 50) wins"},
		{"timed", "points, live scoring; ends when the game clock (default 10m) runs out"},
		{"speed", "points, live scoring; a move later than the move clock (default 60s) passes the turn"},
		{"cq", "points, live scoring; a question about your argument left unanswered on your next turn loses"},
		{"debate", "a central claim: pro (first seat) bears the burden of proof and wins if the claim is In; a losing side that passes concedes"},
		{"solo", "one player argues both sides of a claim until they pass; the result is the verdict (belief revision)"},
	}
}

// ModeConfig returns the configuration of a preset (see Modes).
func ModeConfig(name string) (GameConfig, error) {
	c := DefaultGameConfig()
	c.Mode = name
	live := func() { c.Scoring.Model = ScoringLive }
	switch name {
	case "classic", "":
		c.Mode = "classic"
	case "infinite":
		live()
		c.MaxTurns = 0
	case "target":
		live()
		c.MaxTurns = 0
		c.TargetScore = 50
	case "timed":
		live()
		c.MaxTurns = 0
		c.TimeLimit = 10 * time.Minute
	case "speed":
		live()
		c.MoveTimeLimit = 60 * time.Second
	case "cq":
		live()
		c.MaxTurns = 30
		c.AnswerWithin = 1
	case "debate":
		live()
		c.Objective = ObjectiveClaim
		c.ConcedeOnPass = true
	case "solo":
		live()
		c.Objective = ObjectiveClaim
		c.AbstentionThreshold = 1
	default:
		names := make([]string, 0, len(Modes()))
		for _, m := range Modes() {
			names = append(names, m.Name)
		}
		return c, fmt.Errorf("unknown mode %q (modes: %s)", name, strings.Join(names, ", "))
	}
	return c, nil
}

// -------------------------------------------------------------------------
// Outcome
// -------------------------------------------------------------------------

// Outcome is the result of a game: final once it has ended, provisional
// (who would win if it ended now) before.
type Outcome struct {
	Over   bool
	Reason string
	// Winner is the winning player's ID; "" for a draw, or when nobody can
	// win (a solo game, whose result is the verdict).
	Winner string
	Draw   bool
	// Claim, Verdict and WinningSide are set for ObjectiveClaim: the claim,
	// its label ("In", "Out", "Undecided"; "Out" while no claim is made) and
	// "pro" or "con".
	Claim       string
	Verdict     string
	WinningSide string
	Loser       string // the player who forfeited, if any
	Scores      map[string]float64
}

// Outcome returns the game's result.
func (gs *GameState) Outcome() *Outcome {
	o := &Outcome{Over: gs.Terminated, Reason: gs.TerminationReason, Loser: gs.Loser, Scores: map[string]float64{}}
	for _, p := range gs.Players {
		o.Scores[p.ID] = p.Score
	}
	if gs.Config.Objective == ObjectiveClaim {
		o.Claim = gs.Claim
		o.Verdict = verdictString(caes.Out)
		if s := gs.Adapter.Graph.Statements[gs.Claim]; s != nil {
			o.Verdict = verdictString(s.Label)
		}
		if len(gs.Players) < 2 {
			return o // solo: the verdict is the result
		}
		switch {
		case gs.Loser != "" && gs.Side(gs.Loser) == "pro":
			o.WinningSide = "con"
		case gs.Loser != "":
			o.WinningSide = "pro"
		case o.Verdict == verdictString(caes.In):
			o.WinningSide = "pro"
		default: // the burden of proof is pro's
			o.WinningSide = "con"
		}
		o.Winner = gs.best(func(p *Player) bool { return gs.Side(p.ID) == o.WinningSide })
		return o
	}
	if gs.Loser != "" && len(gs.Players) == 2 {
		o.Winner = gs.best(func(p *Player) bool { return p.ID != gs.Loser })
		return o
	}
	o.Winner = gs.best(func(p *Player) bool { return p.ID != gs.Loser && !p.Resigned })
	o.Draw = o.Winner == ""
	return o
}

func verdictString(l caes.Label) string {
	switch l {
	case caes.In:
		return "In"
	case caes.Out:
		return "Out"
	}
	return "Undecided"
}

// best returns the ID of the eligible player with the highest score, or ""
// if two of them share it (or none is eligible).
func (gs *GameState) best(eligible func(*Player) bool) string {
	winner, top, tied := "", 0.0, false
	for _, p := range gs.Players {
		if !eligible(p) {
			continue
		}
		switch {
		case winner == "" || p.Score > top:
			winner, top, tied = p.ID, p.Score, false
		case p.Score == top:
			tied = true
		}
	}
	if tied {
		return ""
	}
	return winner
}

// verdictFavours reports whether the claim's current label favours the
// player's side: pro needs the claim In; con needs it not In.
func (gs *GameState) verdictFavours(playerID string) bool {
	in := false
	if s := gs.Adapter.Graph.Statements[gs.Claim]; s != nil {
		in = s.Label == caes.In
	}
	return in == (gs.Side(playerID) == "pro")
}

// -------------------------------------------------------------------------
// Live scoring
// -------------------------------------------------------------------------

// LiveScores computes every player's ScoringLive score from the current
// labels (see ScoringLive).
func (gs *GameState) LiveScores() map[string]float64 {
	cfg := gs.Config.Scoring
	ag := gs.Adapter.Graph
	best := map[string]map[*caes.Statement]float64{} // player -> statement argued for -> best value
	undercuts := map[string]map[string]bool{}        // player -> arguments their undercutters defeat
	for _, m := range gs.MoveHistory {
		if !m.Scorable || len(m.ArgIDs) == 0 {
			continue
		}
		if best[m.PlayerID] == nil {
			best[m.PlayerID], undercuts[m.PlayerID] = map[*caes.Statement]float64{}, map[string]bool{}
		}
		for _, id := range m.ArgIDs {
			arg := ag.Arguments[id]
			if arg == nil || arg.Conclusion == nil || carneades.ArgumentRole(arg) != carneades.RoleMain {
				continue
			}
			v := 0.0
			if arg.Weight > 0 { // in force: applicable and not undercut
				v = basePoints(arg.Conclusion.Label, cfg) * m.Weight
			}
			if old, ok := best[m.PlayerID][arg.Conclusion]; !ok || v > old {
				best[m.PlayerID][arg.Conclusion] = v
			}
		}
		if m.Action.AttackType == "undercutter" {
			main := ag.Arguments[m.ArgIDs[0]]
			target := ag.Arguments[m.Action.TargetArgID]
			if main != nil && main.Conclusion.Label == caes.In && target != nil && target.Weight == 0 &&
				target.Conclusion != nil && target.Conclusion.Label == caes.Out {
				undercuts[m.PlayerID][target.Id] = true
			}
		}
	}
	scores := make(map[string]float64, len(gs.Players))
	for _, p := range gs.Players {
		sum := p.Flat + float64(len(undercuts[p.ID]))*cfg.UndercutterBonus
		stmts := make([]*caes.Statement, 0, len(best[p.ID]))
		for s := range best[p.ID] {
			stmts = append(stmts, s)
		}
		sort.Slice(stmts, func(i, j int) bool { return stmts[i].Id < stmts[j].Id }) // a fixed order of addition
		for _, s := range stmts {
			sum += best[p.ID][s]
		}
		scores[p.ID] = sum
	}
	return scores
}

// rescore sets every player's score to their live score and returns the
// changes.
func (gs *GameState) rescore() map[string]float64 {
	live := gs.LiveScores()
	deltas := map[string]float64{}
	for _, p := range gs.Players {
		if d := live[p.ID] - p.Score; d != 0 {
			deltas[p.ID] = d
		}
		p.Score = live[p.ID]
	}
	return deltas
}

// -------------------------------------------------------------------------
// Snapshot (Simulate)
// -------------------------------------------------------------------------

// snapshot is the game-level state a move can change, besides the graph.
type snapshot struct {
	players []Player // copies, with their own SchemeUsageCounts maps
	turn    int
	active  int
	history int

	roundHadMove                  bool
	consecAllAbstain              int
	terminated, truncated         bool
	reason, rejection, claim, lsr string
	questions                     []OpenQuestion
	turnStartedAt                 time.Time
}

func (gs *GameState) takeSnapshot() *snapshot {
	s := &snapshot{
		turn: gs.CurrentTurn, active: gs.ActivePlayerIdx, history: len(gs.MoveHistory),
		roundHadMove: gs.roundHadMove, consecAllAbstain: gs.ConsecAllAbstain,
		terminated: gs.Terminated, truncated: gs.Truncated, reason: gs.TerminationReason, rejection: gs.LastRejection,
		claim: gs.Claim, lsr: gs.Loser, questions: append([]OpenQuestion{}, gs.Questions...),
		turnStartedAt: gs.TurnStartedAt,
	}
	for _, p := range gs.Players {
		c := *p
		c.SchemeUsageCounts = make(map[string]int, len(p.SchemeUsageCounts))
		for k, v := range p.SchemeUsageCounts {
			c.SchemeUsageCounts[k] = v
		}
		s.players = append(s.players, c)
	}
	return s
}

func (gs *GameState) restoreSnapshot(s *snapshot) {
	for i, p := range gs.Players {
		*p = s.players[i]
	}
	gs.CurrentTurn, gs.ActivePlayerIdx, gs.MoveHistory = s.turn, s.active, gs.MoveHistory[:s.history]
	gs.roundHadMove, gs.ConsecAllAbstain = s.roundHadMove, s.consecAllAbstain
	gs.Terminated, gs.Truncated, gs.TerminationReason, gs.LastRejection = s.terminated, s.truncated, s.reason, s.rejection
	gs.Claim, gs.Loser, gs.Questions = s.claim, s.lsr, s.questions
	gs.TurnStartedAt = s.turnStartedAt
}
