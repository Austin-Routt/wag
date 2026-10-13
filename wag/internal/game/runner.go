package game

import (
	"fmt"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
)

// GameResult captures the outcome of a completed game for analysis.
type GameResult struct {
	TurnsPlayed    int
	MoveCount      int // non-abstain moves
	AbstainCount   int
	InvalidCount   int // rejected moves (retried)
	Scores         map[string]float64
	Winner         string         // the winner's player ID (Outcome), or "tie" if there is none
	WinnerSeat     int            // index of the winner in turn order, or -1 for a tie
	Verdict        string         // ObjectiveClaim: the claim's final label
	MoveKinds      map[string]int // accepted moves by kind: "support/<scheme>" or "attack/<attack type>/<scheme>"
	Reason         string         // termination reason
	StatementCount int
	ArgumentCount  int
	IssueCount     int
	Panicked       bool // true if the game panicked (recovered)
	PanicMessage   string

	// Provenance: with these, a game can be replayed and checked.
	YAMLSHA256  string           // SHA-256 of the Carneades YAML
	NASCQSHA256 string           // SHA-256 of the NASCQ schemes, CQs and edges
	AgentSeeds  map[string]int64 // seed of every agent that implements Seeded, by player name
	MoveDigest  string           // GameState.MoveDigest() at the end of the game
}

// RunnerConfig controls the game runner behavior.
type RunnerConfig struct {
	MaxRetriesPerTurn int // max invalid actions before forcing abstain (default: 5)
}

// DefaultRunnerConfig returns sensible defaults.
func DefaultRunnerConfig() RunnerConfig {
	return RunnerConfig{
		MaxRetriesPerTurn: 5,
	}
}

// RunGame plays a complete game between the given agents using the provided
// configuration, Carneades YAML, and NASCQ data.
//
// The runner handles:
//   - Turn looping until termination
//   - Retrying invalid actions (up to MaxRetriesPerTurn, then force abstain)
//   - Panic recovery (the game result records panics without crashing)
//   - Statistics collection
//
// The yamlBytes and NASCQ JSON bytes are used to create a fresh game for
// each run, ensuring no state leakage between games.
func RunGame(
	gameConfig GameConfig,
	runnerConfig RunnerConfig,
	yamlBytes []byte,
	schemesJSON, cqsJSON, edgesJSON []byte,
	agents []Agent,
) (result *GameResult) {
	result = &GameResult{
		Scores:     make(map[string]float64),
		AgentSeeds: make(map[string]int64),
		MoveKinds:  make(map[string]int),
		WinnerSeat: -1,
	}
	result.YAMLSHA256, result.NASCQSHA256 = InputHashes(yamlBytes, schemesJSON, cqsJSON, edgesJSON)
	for _, a := range agents {
		if s, ok := a.(Seeded); ok {
			result.AgentSeeds[a.Name()] = s.Seed()
		}
	}

	// Panic recovery — record but don't crash.
	defer func() {
		if r := recover(); r != nil {
			result.Panicked = true
			result.PanicMessage = fmt.Sprintf("%v", r)
		}
	}()

	// Create fresh adapter from YAML.
	adapter, err := carneades.NewAdapterFromYAML(yamlBytes)
	if err != nil {
		result.Panicked = true
		result.PanicMessage = fmt.Sprintf("YAML load failed: %v", err)
		return
	}

	// Create fresh NASCQ from JSON.
	n, err := nascq.LoadFromBytes(schemesJSON, cqsJSON, edgesJSON)
	if err != nil {
		result.Panicked = true
		result.PanicMessage = fmt.Sprintf("NASCQ load failed: %v", err)
		return
	}

	// Build player names from agents.
	playerNames := make([]string, len(agents))
	for i, a := range agents {
		playerNames[i] = a.Name()
	}

	gs := NewGame(gameConfig, adapter, n, playerNames)

	// Main game loop.
	maxIterations := gameConfig.MaxTurns * len(agents) * (runnerConfig.MaxRetriesPerTurn + 1) * 2
	if maxIterations < 1000 {
		maxIterations = 1000
	}
	iteration := 0

	for !gs.Terminated && iteration < maxIterations {
		iteration++
		activeIdx := gs.ActivePlayerIdx
		agent := agents[activeIdx]

		// Build observation for the agent.
		obs := gs.BuildObservation()

		// Let the agent choose an action, with retries on failure.
		var turnResult *TurnResult
		for retry := 0; retry <= runnerConfig.MaxRetriesPerTurn; retry++ {
			action := agent.ChooseAction(obs)

			tr, err := gs.ExecuteTurn(action)
			if err != nil {
				// Game-level error (e.g., terminated) — break out.
				result.Reason = fmt.Sprintf("error: %v", err)
				goto done
			}

			if tr.Accepted {
				turnResult = tr
				if tr.Abstained {
					result.AbstainCount++
				} else {
					result.MoveCount++
					result.MoveKinds[moveKind(gs.MoveHistory[len(gs.MoveHistory)-1])]++
				}
				break
			}
			if tr.Terminated { // the game clock ran out, or a late move lost
				turnResult = tr
				break
			}

			// Action rejected — count and retry.
			result.InvalidCount++
			if retry == runnerConfig.MaxRetriesPerTurn {
				// Force abstain after max retries.
				abstain := nascq.ProposedAction{Type: "abstain"}
				tr, _ = gs.ExecuteTurn(abstain)
				turnResult = tr
				result.AbstainCount++
			}
		}

		if turnResult != nil && turnResult.Terminated {
			break
		}
	}

done:
	// Collect results.
	result.TurnsPlayed = gs.CurrentTurn - 1
	if result.TurnsPlayed < 1 {
		result.TurnsPlayed = len(gs.MoveHistory)
	}
	result.Reason = gs.TerminationReason
	if !gs.Terminated && result.Reason == "" {
		result.Reason = fmt.Sprintf("runner stopped after %d turns", iteration)
	}
	result.StatementCount = len(gs.Adapter.Graph.Statements)
	result.ArgumentCount = len(gs.Adapter.Graph.Arguments)
	result.IssueCount = len(gs.Adapter.Graph.Issues)
	result.MoveDigest = gs.MoveDigest()

	// Scores and winner, by the game's objective (Outcome).
	o := gs.Outcome()
	result.Verdict = o.Verdict
	for _, p := range gs.Players {
		result.Scores[p.ID] = p.Score
	}
	result.Winner, result.WinnerSeat = "tie", -1
	for i, p := range gs.Players {
		if p.ID == o.Winner {
			result.Winner, result.WinnerSeat = p.ID, i
		}
	}

	return
}

// BatchResult aggregates statistics across multiple game runs.
type BatchResult struct {
	GamesPlayed   int
	Panics        int
	PanicMessages []string
	Wins          map[string]int // playerName → win count
	Ties          int
	TotalMoves    int
	TotalAbstains int
	TotalInvalid  int
	AvgTurns      float64
	AvgStatements float64
	AvgArguments  float64
	MaxStatements int
	MaxArguments  int
	SeatWins      map[int]int        // wins by seat (index in turn order): a first- or last-mover advantage shows here
	AvgScores     map[string]float64 // playerName → mean final score
	MoveKinds     map[string]int     // accepted moves by kind, over all games (see GameResult.MoveKinds)
	Verdicts      map[string]int     // ObjectiveClaim: final verdicts ("In", "Out", ...)
	Reasons       map[string]int     // why the games ended
}

// RunBatch plays multiple games and collects aggregate statistics.
// This is the primary stress test mechanism.
func RunBatch(
	numGames int,
	gameConfig GameConfig,
	runnerConfig RunnerConfig,
	yamlBytes []byte,
	schemesJSON, cqsJSON, edgesJSON []byte,
	agentFactory func(gameIndex int) []Agent,
) *BatchResult {
	br := &BatchResult{
		Wins:      make(map[string]int),
		SeatWins:  make(map[int]int),
		AvgScores: make(map[string]float64),
		MoveKinds: make(map[string]int),
		Verdicts:  make(map[string]int),
		Reasons:   make(map[string]int),
	}

	totalTurns := 0
	totalStmts := 0
	totalArgs := 0

	for i := 0; i < numGames; i++ {
		agents := agentFactory(i)
		result := RunGame(gameConfig, runnerConfig, yamlBytes, schemesJSON, cqsJSON, edgesJSON, agents)

		br.GamesPlayed++
		if result.Panicked {
			br.Panics++
			br.PanicMessages = append(br.PanicMessages, result.PanicMessage)
			continue
		}

		if result.Winner == "tie" {
			br.Ties++
		} else {
			br.Wins[result.Winner]++
			br.SeatWins[result.WinnerSeat]++
		}
		for name, sc := range result.Scores {
			br.AvgScores[name] += sc
		}
		for k, n := range result.MoveKinds {
			br.MoveKinds[k] += n
		}
		if result.Verdict != "" {
			br.Verdicts[result.Verdict]++
		}
		br.Reasons[reasonKind(result.Reason)]++

		br.TotalMoves += result.MoveCount
		br.TotalAbstains += result.AbstainCount
		br.TotalInvalid += result.InvalidCount
		totalTurns += result.TurnsPlayed
		totalStmts += result.StatementCount
		totalArgs += result.ArgumentCount

		if result.StatementCount > br.MaxStatements {
			br.MaxStatements = result.StatementCount
		}
		if result.ArgumentCount > br.MaxArguments {
			br.MaxArguments = result.ArgumentCount
		}
	}

	completed := br.GamesPlayed - br.Panics
	if completed > 0 {
		br.AvgTurns = float64(totalTurns) / float64(completed)
		br.AvgStatements = float64(totalStmts) / float64(completed)
		br.AvgArguments = float64(totalArgs) / float64(completed)
		for name := range br.AvgScores {
			br.AvgScores[name] /= float64(completed)
		}
	}

	return br
}

// moveKind classifies an accepted move for the batch statistics.
func moveKind(m Move) string {
	a := m.Action
	switch {
	case m.Resigned:
		return "resign"
	case a.Type == "question":
		return "question/" + a.AnswersCQID
	case a.Type == "attack":
		return "attack/" + a.AttackType + "/" + a.SchemeID
	case a.TargetArgID != "":
		return "defence/" + a.SchemeID
	}
	return a.Type + "/" + a.SchemeID
}

// reasonKind is a termination reason without the players' names and numbers,
// for counting.
func reasonKind(r string) string {
	for _, k := range []string{"resigned", "unanswered", "ran out of time", "time limit", "target score",
		"max rounds", "abstained", "burden of proof", "conceded", "runner stopped", "error"} {
		if strings.Contains(r, k) {
			return k
		}
	}
	return r
}
