// WAG CLI — Waltonian Argumentation Game command-line interface.
//
// Usage:
//
//	go run ./cmd/wag-cli/ --p1 human --p2 random --turns 10
//
// Flags:
//
//	--p1      Player 1 type: "human", "random", or "greedy" (default: "human")
//	--p2      Player 2 type: "human", "random", or "greedy" (default: "random")
//	--turns   Maximum turns per game (default: 10)
//	--seed    Random seed for AI agents (default: 42)
//	--name1   Player 1 name (default: "Player1")
//	--name2   Player 2 name (default: "Player2")
//	--data    Directory with schemes/core.yml and nascq/{schemes,critical_questions,edges}.json
//	          (default: the data built into the binary)
//	--lint    Check the data (see game.LintData), print the findings, and exit:
//	          status 1 if there is an error
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"
)

func main() {
	// ---------------------------------------------------------------
	// Flag parsing
	// ---------------------------------------------------------------
	p1Type := flag.String("p1", "human", "Player 1 type: human, random, greedy")
	p2Type := flag.String("p2", "random", "Player 2 type: human, random, greedy")
	maxTurns := flag.Int("turns", 10, "Maximum turns (rounds)")
	seed := flag.Int64("seed", 42, "Random seed for AI agents")
	name1 := flag.String("name1", "Player1", "Player 1 display name")
	name2 := flag.String("name2", "Player2", "Player 2 display name")
	dataDir := flag.String("data", "", "directory with schemes/core.yml and nascq/*.json (default: built-in data)")
	lint := flag.Bool("lint", false, "check the data, print the findings and exit (status 1 on an error)")
	flag.Parse()

	files, err := loadData(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading data: %v\n", err)
		os.Exit(1)
	}
	if *lint {
		os.Exit(runLint(files))
	}

	// ---------------------------------------------------------------
	// Create agents
	// ---------------------------------------------------------------
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

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║     WAG — Waltonian Argumentation Game          ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Printf("  %s (%s) vs %s (%s)\n", *name1, *p1Type, *name2, *p2Type)
	fmt.Printf("  Max turns: %d  |  Seed: %d\n\n", *maxTurns, *seed)

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
				fmt.Printf("\n  ✗ Move rejected: %s\n  Try again.\n", result.Error)
				// Humans always get to retry. Rebuild observation
				// (score changed from penalty).
				obs = gs.BuildObservation()
				continue
			}

			// AI agent: force abstain after max retries.
			if retries >= maxRetries {
				abstain := nascq.ProposedAction{Type: "abstain"}
				result, _ = gs.ExecuteTurn(abstain)
				fmt.Printf("  %s forced to abstain after %d invalid attempts.\n",
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
	fmt.Println("\n╔══════════════════════════════════════════════════╗")
	fmt.Println("║                   GAME OVER                     ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")

	if gs.TerminationReason != "" {
		fmt.Printf("  Reason: %s\n", gs.TerminationReason)
	}

	fmt.Println("\n  Final Scores:")
	bestScore := -1e9
	winner := ""
	for _, p := range gs.Players {
		fmt.Printf("    %s: %.2f\n", p.Name, p.Score)
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
		fmt.Println("\n  Result: TIE!")
	} else {
		fmt.Printf("\n  Winner: %s!\n", winner)
	}

	fmt.Printf("\n  Graph: %d statements, %d arguments, %d issues\n",
		len(gs.Adapter.Graph.Statements),
		len(gs.Adapter.Graph.Arguments),
		len(gs.Adapter.Graph.Issues))
	fmt.Printf("  Moves: %d total\n\n", len(gs.MoveHistory))

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
	default:
		return nil, fmt.Errorf("unknown agent type %q (use human, random, or greedy)", agentType)
	}
}

// displayResult prints a brief summary after each accepted move.
func displayResult(result *game.TurnResult, agent game.Agent, gs *game.GameState) {
	// Don't print for human agents — they already see the state via ChooseAction.
	if _, isHuman := agent.(*game.HumanAgent); isHuman {
		if result.Abstained {
			fmt.Println("\n  You chose to abstain.")
		} else {
			fmt.Printf("\n  ✓ Move accepted! Arguments created: %v\n", result.ArgIDs)
			fmt.Printf("    Score delta: %+.2f\n", result.ScoreDelta)
		}
		return
	}

	// For AI agents, print a one-line summary.
	if result.Abstained {
		fmt.Printf("  %s abstains.\n", agent.Name())
	} else {
		fmt.Printf("  %s plays → args=%v  score=%+.2f\n",
			agent.Name(), result.ArgIDs, result.ScoreDelta)
	}

	if result.Terminated {
		fmt.Printf("  Game over: %s\n", result.TerminationReason)
	}
}

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
