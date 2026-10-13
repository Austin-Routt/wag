// WAG CLI — Waltonian Argumentation Game command-line interface.
//
// Usage:
//
//	go run ./cmd/wag-cli/ --p1 human --p2 random --turns 10
//
// Flags:
//
//	--p1      Player 1 type: "human", "random", "greedy" or "json" (default: "human")
//	--p2      Player 2 type: "human", "random", "greedy" or "json" (default: "random")
//	--turns   Maximum turns per game (default: 10)
//	--seed    Random seed for AI agents (default: 42)
//	--name1   Player 1 name (default: "Player1")
//	--name2   Player 2 name (default: "Player2")
//	--data    Directory with schemes/core.yml and nascq/{schemes,critical_questions,edges}.json
//	          (default: the data built into the binary)
//	--lint    Check the data (see game.LintData), print the findings, and exit:
//	          status 1 if there is an error
//	--games   Play N games between two AI players (random/greedy) and print the
//	          statistics: wins by player and by seat, mean scores, the moves played
//
// A "json" player is a program on the other end of standard input and output
// (game.JSONAgent): it gets the game as one JSON line (game.View) whenever it
// is to move and answers with one JSON line (nascq.ProposedAction). With a
// json player, everything else the CLI prints goes to standard error.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"
)

func main() {
	// ---------------------------------------------------------------
	// Flag parsing
	// ---------------------------------------------------------------
	p1Type := flag.String("p1", "human", "Player 1 type: human, random, greedy, json")
	p2Type := flag.String("p2", "random", "Player 2 type: human, random, greedy, json")
	maxTurns := flag.Int("turns", 10, "Maximum turns (rounds)")
	seed := flag.Int64("seed", 42, "Random seed for AI agents")
	name1 := flag.String("name1", "Player1", "Player 1 display name")
	name2 := flag.String("name2", "Player2", "Player 2 display name")
	dataDir := flag.String("data", "", "directory with schemes/core.yml and nascq/*.json (default: built-in data)")
	lint := flag.Bool("lint", false, "check the data, print the findings and exit (status 1 on an error)")
	games := flag.Int("games", 0, "play this many AI-vs-AI games and print statistics")
	flag.Parse()

	files, err := loadData(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading data: %v\n", err)
		os.Exit(1)
	}
	if *lint {
		os.Exit(runLint(files))
	}
	if *games > 0 {
		os.Exit(runGames(files, *games, *maxTurns, *seed, *p1Type, *p2Type, *name1, *name2))
	}

	// ---------------------------------------------------------------
	// Create agents
	// ---------------------------------------------------------------
	// One program may play both json players (each View names the active
	// player), and two humans may share a terminal, but a human's prompts
	// would corrupt a json player's stream.
	types := map[string]bool{*p1Type: true, *p2Type: true}
	if types["json"] {
		if types["human"] {
			fmt.Fprintln(os.Stderr, "Error: a json player and a human player cannot share standard input and output")
			os.Exit(1)
		}
		say = os.Stderr // standard output carries the JSON
	}
	agent1, err := makeAgent(*p1Type, *name1, *seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating player 1: %v\n", err)
		os.Exit(1)
	}
	agent2, err := makeAgent(*p2Type, *name2, *seed+1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating player 2: %v\n", err)
		os.Exit(1)
	}
	agents := []game.Agent{agent1, agent2}

	fmt.Fprintln(say, "╔══════════════════════════════════════════════════╗")
	fmt.Fprintln(say, "║     WAG — Waltonian Argumentation Game          ║")
	fmt.Fprintln(say, "╚══════════════════════════════════════════════════╝")
	fmt.Fprintf(say, "  %s (%s) vs %s (%s)\n", *name1, *p1Type, *name2, *p2Type)
	fmt.Fprintf(say, "  Max turns: %d  |  Seed: %d\n\n", *maxTurns, *seed)

	// ---------------------------------------------------------------
	// Create game environment
	// ---------------------------------------------------------------
	env := game.NewEnv(game.EnvConfig{
		Game: game.GameConfig{
			MaxTurns:            *maxTurns,
			AbstentionThreshold: 3,
			Scoring:             game.DefaultScoringConfig(),
		},
		YAMLBytes:   files.yaml,
		SchemesJSON: files.schemes,
		CQsJSON:     files.cqs,
		EdgesJSON:   files.edges,
		PlayerNames: []string{*name1, *name2},
	})

	obs, err := env.Reset()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing game: %v\n", err)
		os.Exit(1)
	}

	// ---------------------------------------------------------------
	// Game loop
	// ---------------------------------------------------------------
	gs := env.GameState()
	maxRetries := 5

	for !gs.Terminated {
		activeIdx := gs.ActivePlayerIdx
		agent := agents[activeIdx]

		// Rebuild observation (may have changed from retries).
		obs = gs.BuildObservation()

		retries := 0
		for {
			action := agent.ChooseAction(obs)

			result, err := gs.ExecuteTurn(action)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n  Game error: %v\n", err)
				goto gameOver
			}

			if result.Accepted {
				displayResult(result, agent, gs)
				if result.Terminated {
					goto gameOver
				}
				break
			}

			// Move rejected.
			retries++
			_, isHuman := agent.(*game.HumanAgent)
			if isHuman {
				fmt.Fprintf(say, "\n  ✗ Move rejected: %s\n  Try again.\n", result.Error)
				// Humans always get to retry. Rebuild observation
				// (score changed from penalty).
				obs = gs.BuildObservation()
				continue
			}

			// AI agent: force abstain after max retries.
			if retries >= maxRetries {
				abstain := nascq.ProposedAction{Type: "abstain"}
				result, _ = gs.ExecuteTurn(abstain)
				fmt.Fprintf(say, "  %s forced to abstain after %d invalid attempts.\n",
					agent.Name(), retries)
				if result != nil && result.Terminated {
					goto gameOver
				}
				break
			}
			// AI retry: rebuild observation and try again.
			obs = gs.BuildObservation()
		}
	}

gameOver:
	// ---------------------------------------------------------------
	// Final results
	// ---------------------------------------------------------------
	fmt.Fprintln(say, "\n╔══════════════════════════════════════════════════╗")
	fmt.Fprintln(say, "║                   GAME OVER                     ║")
	fmt.Fprintln(say, "╚══════════════════════════════════════════════════╝")

	if gs.TerminationReason != "" {
		fmt.Fprintf(say, "  Reason: %s\n", gs.TerminationReason)
	}

	fmt.Fprintln(say, "\n  Final Scores:")
	bestScore := -1e9
	winner := ""
	for _, p := range gs.Players {
		fmt.Fprintf(say, "    %s: %.2f\n", p.Name, p.Score)
		if p.Score > bestScore {
			bestScore = p.Score
			winner = p.Name
		}
	}

	// Check tie.
	tieCount := 0
	for _, p := range gs.Players {
		if p.Score == bestScore {
			tieCount++
		}
	}
	if tieCount > 1 {
		fmt.Fprintln(say, "\n  Result: TIE!")
	} else {
		fmt.Fprintf(say, "\n  Winner: %s!\n", winner)
	}

	fmt.Fprintf(say, "\n  Graph: %d statements, %d arguments, %d issues\n",
		len(gs.Adapter.Graph.Statements),
		len(gs.Adapter.Graph.Arguments),
		len(gs.Adapter.Graph.Issues))
	fmt.Fprintf(say, "  Moves: %d total\n\n", len(gs.MoveHistory))

	_ = obs // suppress unused warning
}

// makeAgent creates an Agent based on the type string.
func makeAgent(agentType, name string, seed int64) (game.Agent, error) {
	switch agentType {
	case "human":
		return game.NewHumanAgent(name, os.Stdin, os.Stdout), nil
	case "random":
		return game.NewRandomAgent(name, seed, 0.15, 0.55), nil
	case "greedy":
		return game.NewGreedyAgent(name, seed, 3), nil
	case "json":
		return game.NewJSONAgent(name, os.Stdin, os.Stdout), nil
	default:
		return nil, fmt.Errorf("unknown agent type %q (use human, random, greedy or json)", agentType)
	}
}

// displayResult prints a brief summary after each accepted move.
func displayResult(result *game.TurnResult, agent game.Agent, gs *game.GameState) {
	// Don't print for human agents — they already see the state via ChooseAction.
	if _, isHuman := agent.(*game.HumanAgent); isHuman {
		if result.Abstained {
			fmt.Fprintln(say, "\n  You chose to abstain.")
		} else {
			fmt.Fprintf(say, "\n  ✓ Move accepted! Arguments created: %v\n", result.ArgIDs)
			fmt.Fprintf(say, "    Score delta: %+.2f\n", result.ScoreDelta)
		}
		return
	}

	// For AI agents, print a one-line summary.
	if result.Abstained {
		fmt.Fprintf(say, "  %s abstains.\n", agent.Name())
	} else {
		fmt.Fprintf(say, "  %s plays → args=%v  score=%+.2f\n",
			agent.Name(), result.ArgIDs, result.ScoreDelta)
	}

	if result.Terminated {
		fmt.Fprintf(say, "  Game over: %s\n", result.TerminationReason)
	}
}

// say receives the CLI's human-readable output.
var say io.Writer = os.Stdout

// dataFiles holds the game data: the Carneades YAML and the NASCQ files.
type dataFiles struct{ yaml, schemes, cqs, edges []byte }

// loadData reads the data from dir, or returns the built-in data if dir is "".
func loadData(dir string) (dataFiles, error) {
	if dir == "" {
		return dataFiles{data.CoreYAML, data.SchemesJSON, data.CQsJSON, data.EdgesJSON}, nil
	}
	var f dataFiles
	for _, x := range []struct {
		dst  *[]byte
		path string
	}{
		{&f.yaml, "schemes/core.yml"},
		{&f.schemes, "nascq/schemes.json"},
		{&f.cqs, "nascq/critical_questions.json"},
		{&f.edges, "nascq/edges.json"},
	} {
		b, err := os.ReadFile(filepath.Join(dir, x.path))
		if err != nil {
			return f, err
		}
		*x.dst = b
	}
	return f, nil
}

// runLint prints the data's lint findings and returns the exit status.
func runLint(f dataFiles) int {
	issues, err := game.LintData(f.yaml, f.schemes, f.cqs, f.edges)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data: %v\n", err)
		return 1
	}
	errors := 0
	for _, i := range issues {
		fmt.Println(i)
		if i.Error {
			errors++
		}
	}
	fmt.Printf("%d error(s), %d warning(s)\n", errors, len(issues)-errors)
	if errors > 0 {
		return 1
	}
	return 0
}

// runGames plays n games between two AI players, with seeds derived from
// seed, prints the batch statistics and returns the exit status.
func runGames(f dataFiles, n, maxTurns int, seed int64, p1, p2, name1, name2 string) int {
	for _, typ := range []string{p1, p2} {
		if typ != "random" && typ != "greedy" {
			fmt.Fprintf(os.Stderr, "Error: --games needs AI players (random or greedy), not %q\n", typ)
			return 1
		}
	}
	if name1 == name2 {
		name2 += "'"
	}
	cfg := game.GameConfig{MaxTurns: maxTurns, AbstentionThreshold: 3, Scoring: game.DefaultScoringConfig()}
	br := game.RunBatch(n, cfg, game.DefaultRunnerConfig(), f.yaml, f.schemes, f.cqs, f.edges, func(i int) []game.Agent {
		a1, _ := makeAgent(p1, name1, seed+int64(2*i))
		a2, _ := makeAgent(p2, name2, seed+int64(2*i+1))
		return []game.Agent{a1, a2}
	})
	fmt.Printf("%d games (%d panicked), %d turns at most\n", br.GamesPlayed, br.Panics, maxTurns)
	for _, name := range []string{name1, name2} {
		fmt.Printf("  %-12s wins %4d   mean score %7.2f\n", name, br.Wins[name], br.AvgScores[name])
	}
	fmt.Printf("  ties %d; wins by seat: first %d, second %d\n", br.Ties, br.SeatWins[0], br.SeatWins[1])
	fmt.Printf("  moves %d, abstentions %d, rejected %d; mean %.1f arguments per game\n",
		br.TotalMoves, br.TotalAbstains, br.TotalInvalid, br.AvgArguments)
	kinds := make([]string, 0, len(br.MoveKinds))
	for k := range br.MoveKinds {
		kinds = append(kinds, k)
	}
	sort.Slice(kinds, func(i, j int) bool {
		return br.MoveKinds[kinds[i]] > br.MoveKinds[kinds[j]] || br.MoveKinds[kinds[i]] == br.MoveKinds[kinds[j]] && kinds[i] < kinds[j]
	})
	for _, k := range kinds {
		fmt.Printf("  %6d  %s\n", br.MoveKinds[k], k)
	}
	if br.Panics > 0 {
		for _, m := range br.PanicMessages {
			fmt.Printf("  panic: %s\n", m)
		}
		return 1
	}
	return 0
}
