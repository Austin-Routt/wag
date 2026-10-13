// WAG CLI — Waltonian Argumentation Game command-line interface.
//
// Usage:
//
//	go run ./cmd/wag-cli/ --p1 human --p2 random --turns 10
//
// Flags:
//
//	--mode    Game mode (default: "classic"); --modes lists them: classic, infinite,
//	          target, timed, speed, cq, debate, solo (see game.ModeConfig)
//	--p1      Player 1 type: "human", "random", "greedy" or "json" (default: "human")
//	--p2      Player 2 type: "human", "random", "greedy", "json" or "none" (default:
//	          "random"; "none", and the solo mode, make a one-player game)
//	--turns   Maximum rounds (default: the mode's)
//	--seed    Random seed for AI agents (default: 42)
//	--name1   Player 1 name (default: "Player1")
//	--name2   Player 2 name (default: "Player2")
//	--data    Directory with schemes/core.yml and nascq/{schemes,critical_questions,edges}.json
//	          (default: the data built into the binary)
//	--lint    Check the data (see game.LintData), print the findings, and exit:
//	          status 1 if there is an error
//	--games   Play N games between two AI players (random/greedy) and print the
//	          statistics: wins by player and by seat, mean scores, the moves played,
//	          why the games ended; --alternate swaps the seats every other game
//
// The mode's settings can be changed one by one: --scoring move|live,
// --target N, --time 10m, --move-time 60s, --timeout-loses, --answer-within N,
// --objective points|claim, --claim STATEMENT, --standard PE|CCE|BRD,
// --concede, --undercut-bonus N, --proleptic-bonus N. Self-play is two seats of the same AI type (--p1 greedy --p2
// greedy); solo play is one player arguing both sides.
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
	"strings"

	caes "github.com/carneades/carneades-4/src/engine/caes"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"
)

func main() {
	// ---------------------------------------------------------------
	// Flag parsing
	// ---------------------------------------------------------------
	p1Type := flag.String("p1", "human", "Player 1 type: human, random, greedy, json")
	p2Type := flag.String("p2", "random", "Player 2 type: human, random, greedy, json, none (one player)")
	mode := flag.String("mode", "classic", "game mode (see --modes)")
	listModes := flag.Bool("modes", false, "list the game modes and exit")
	maxTurns := flag.Int("turns", 10, "maximum rounds (0: no limit; default: the mode's)")
	scoring := flag.String("scoring", "", "scoring model: move or live (default: the mode's)")
	target := flag.Float64("target", 0, "target score: the first to reach it wins")
	timeLimit := flag.Duration("time", 0, "game clock, e.g. 10m")
	moveTime := flag.Duration("move-time", 0, "move clock, e.g. 60s: a later move passes the turn")
	timeoutLoses := flag.Bool("timeout-loses", false, "a move later than the move clock loses the game")
	answerWithin := flag.Int("answer-within", 0, "a question about your argument left unanswered for this many of your turns loses")
	objective := flag.String("objective", "", "points or claim (default: the mode's)")
	claim := flag.String("claim", "", "the central claim of a debate, as a statement such as should_be_performed(carbon_tax)")
	standard := flag.String("standard", "", "proof standard of the claim: PE, CCE or BRD")
	concede := flag.Bool("concede", false, "a debater who passes while the verdict is against them concedes")
	undercutBonus := flag.Float64("undercut-bonus", 15, "flat points for an undercut that defeats its target")
	prolepticBonus := flag.Float64("proleptic-bonus", 0, "flat points for answering a question about your own argument before anyone asks it")
	seed := flag.Int64("seed", 42, "Random seed for AI agents")
	name1 := flag.String("name1", "Player1", "Player 1 display name")
	name2 := flag.String("name2", "Player2", "Player 2 display name")
	dataDir := flag.String("data", "", "directory with schemes/core.yml and nascq/*.json (default: built-in data)")
	lint := flag.Bool("lint", false, "check the data, print the findings and exit (status 1 on an error)")
	games := flag.Int("games", 0, "play this many AI-vs-AI games and print statistics")
	alternate := flag.Bool("alternate", false, "with --games: swap the seats every other game, as colours alternate in chess")
	flag.Parse()

	if *listModes {
		for _, m := range game.Modes() {
			fmt.Printf("  %-9s %s\n", m.Name, m.Description)
		}
		return
	}
	files, err := loadData(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading data: %v\n", err)
		os.Exit(1)
	}
	if *lint {
		os.Exit(runLint(files))
	}

	cfg, err := game.ModeConfig(*mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	set := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { set[f.Name] = true })
	if set["turns"] {
		cfg.MaxTurns = *maxTurns
	}
	if set["scoring"] {
		cfg.Scoring.Model = *scoring
	}
	if set["target"] {
		cfg.TargetScore = *target
	}
	if set["time"] {
		cfg.TimeLimit = *timeLimit
	}
	if set["move-time"] {
		cfg.MoveTimeLimit = *moveTime
	}
	if set["timeout-loses"] {
		cfg.MoveTimeoutLoses = *timeoutLoses
	}
	if set["answer-within"] {
		cfg.AnswerWithin = *answerWithin
	}
	if set["objective"] {
		cfg.Objective = *objective
	}
	if set["claim"] {
		cfg.Objective, cfg.Claim = game.ObjectiveClaim, *claim
	}
	if set["concede"] {
		cfg.ConcedeOnPass = *concede
	}
	if set["undercut-bonus"] {
		cfg.Scoring.UndercutterBonus = *undercutBonus
	}
	if set["proleptic-bonus"] {
		cfg.Scoring.ProlepticBonus = *prolepticBonus
	}
	if set["standard"] {
		std, ok := map[string]caes.Standard{"PE": caes.PE, "CCE": caes.CCE, "BRD": caes.BRD}[strings.ToUpper(*standard)]
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: unknown proof standard %q (PE, CCE or BRD)\n", *standard)
			os.Exit(1)
		}
		cfg.ClaimStandard = std
	}
	if msg := checkConfig(cfg); msg != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
		os.Exit(1)
	}
	if *mode == "solo" {
		*p2Type = "none"
	}

	if *games > 0 {
		os.Exit(runGames(files, cfg, *games, *seed, *p1Type, *p2Type, *name1, *name2, *alternate))
	}

	// ---------------------------------------------------------------
	// Create agents
	// ---------------------------------------------------------------
	// One program may play both json players (each View names the active
	// player), and two humans may share a terminal, but a human's prompts
	// would corrupt a json player's stream.
	types := map[string]bool{*p1Type: true, *p2Type: true}
	if *p1Type == "none" {
		fmt.Fprintln(os.Stderr, "Error: player 1 cannot be none")
		os.Exit(1)
	}
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
	agents := []game.Agent{agent1}
	names := []string{*name1}
	if *p2Type != "none" {
		agent2, err := makeAgent(*p2Type, *name2, *seed+1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating player 2: %v\n", err)
			os.Exit(1)
		}
		agents, names = append(agents, agent2), append(names, *name2)
	}

	fmt.Fprintln(say, "╔══════════════════════════════════════════════════╗")
	fmt.Fprintln(say, "║     WAG — Waltonian Argumentation Game          ║")
	fmt.Fprintln(say, "╚══════════════════════════════════════════════════╝")
	if len(agents) == 1 {
		fmt.Fprintf(say, "  %s (%s), alone\n", *name1, *p1Type)
	} else {
		fmt.Fprintf(say, "  %s (%s) vs %s (%s)\n", *name1, *p1Type, *name2, *p2Type)
	}
	fmt.Fprintf(say, "  Mode: %s  |  %s  |  Seed: %d\n\n", cfg.Mode, describeConfig(cfg), *seed)

	// ---------------------------------------------------------------
	// Create game environment
	// ---------------------------------------------------------------
	env := game.NewEnv(game.EnvConfig{
		Game:        cfg,
		YAMLBytes:   files.yaml,
		SchemesJSON: files.schemes,
		CQsJSON:     files.cqs,
		EdgesJSON:   files.edges,
		PlayerNames: names,
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

			if result.Accepted || result.Terminated {
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
	for _, p := range gs.Players {
		fmt.Fprintf(say, "    %s: %.2f\n", p.Name, p.Score)
	}
	o := gs.Outcome()
	if o.Claim != "" || cfg.Objective == game.ObjectiveClaim {
		claimText := "(never made)"
		if o.Claim != "" {
			claimText = gs.Say(o.Claim)
		}
		fmt.Fprintf(say, "\n  Claim: %s\n  Verdict: %s\n", claimText, o.Verdict)
	}
	switch {
	case o.Winner != "":
		fmt.Fprintf(say, "\n  Winner: %s!\n", gs.PlayerByID(o.Winner).Name)
	case len(gs.Players) > 1:
		fmt.Fprintln(say, "\n  Result: TIE!")
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

// checkConfig returns what is wrong with a configuration, or "".
func checkConfig(c game.GameConfig) string {
	switch {
	case c.Scoring.Model != "" && c.Scoring.Model != game.ScoringMove && c.Scoring.Model != game.ScoringLive:
		return fmt.Sprintf("unknown scoring model %q (move or live)", c.Scoring.Model)
	case c.Objective != "" && c.Objective != game.ObjectivePoints && c.Objective != game.ObjectiveClaim:
		return fmt.Sprintf("unknown objective %q (points or claim)", c.Objective)
	case c.MaxTurns < 0 || c.TargetScore < 0 || c.TimeLimit < 0 || c.MoveTimeLimit < 0 || c.AnswerWithin < 0:
		return "limits cannot be negative"
	}
	return ""
}

// describeConfig summarizes the termination conditions and scoring.
func describeConfig(c game.GameConfig) string {
	var parts []string
	if c.Objective == game.ObjectiveClaim {
		std := map[caes.Standard]string{caes.PE: "PE", caes.CCE: "CCE", caes.BRD: "BRD"}[c.ClaimStandard]
		parts = append(parts, "debate on a claim ("+std+")")
	}
	if c.MaxTurns > 0 {
		parts = append(parts, fmt.Sprintf("%d rounds", c.MaxTurns))
	} else {
		parts = append(parts, "no round limit")
	}
	if c.TargetScore > 0 {
		parts = append(parts, fmt.Sprintf("target %.0f", c.TargetScore))
	}
	if c.TimeLimit > 0 {
		parts = append(parts, "game clock "+c.TimeLimit.String())
	}
	if c.MoveTimeLimit > 0 {
		parts = append(parts, "move clock "+c.MoveTimeLimit.String())
	}
	if c.AnswerWithin > 0 {
		parts = append(parts, fmt.Sprintf("questions must be answered within %d turn(s)", c.AnswerWithin))
	}
	model := c.Scoring.Model
	if model == "" {
		model = game.ScoringMove
	}
	parts = append(parts, model+" scoring")
	return strings.Join(parts, ", ")
}

// displayResult prints a brief summary after each accepted move.
func displayResult(result *game.TurnResult, agent game.Agent, gs *game.GameState) {
	if n := len(gs.MoveHistory); n > 0 && (result.Accepted || result.TimedOut) {
		line := gs.DescribeMove(gs.MoveHistory[n-1])
		if result.ScoreDelta != 0 {
			line += fmt.Sprintf("  (%+.2f)", result.ScoreDelta)
		}
		fmt.Fprintf(say, "  %s\n", line)
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

// runGames plays n games between AI players (one, for "none" as the second
// type), with seeds derived from seed, prints the batch statistics and
// returns the exit status.
func runGames(f dataFiles, cfg game.GameConfig, n int, seed int64, p1, p2, name1, name2 string, alternate bool) int {
	for _, typ := range []string{p1, p2} {
		if typ != "random" && typ != "greedy" && !(typ == p2 && typ == "none") {
			fmt.Fprintf(os.Stderr, "Error: --games needs AI players (random or greedy), not %q\n", typ)
			return 1
		}
	}
	if name1 == name2 {
		name2 += "'"
	}
	if cfg.TimeLimit > 0 || cfg.MoveTimeLimit > 0 {
		fmt.Fprintln(os.Stderr, "Note: the clocks run in real time; AI players are fast, so they rarely matter")
	}
	br := game.RunBatch(n, cfg, game.DefaultRunnerConfig(), f.yaml, f.schemes, f.cqs, f.edges, func(i int) []game.Agent {
		a1, _ := makeAgent(p1, name1, seed+int64(2*i))
		if p2 == "none" {
			return []game.Agent{a1}
		}
		a2, _ := makeAgent(p2, name2, seed+int64(2*i+1))
		if alternate && i%2 == 1 {
			return []game.Agent{a2, a1}
		}
		return []game.Agent{a1, a2}
	})
	maxTurns := cfg.MaxTurns
	fmt.Printf("%d games (%d panicked), mode %s: %s; %d rounds at most (0: no limit)\n",
		br.GamesPlayed, br.Panics, cfg.Mode, describeConfig(cfg), maxTurns)
	if p2 == "none" {
		fmt.Printf("  %-12s mean score %7.2f\n", name1, br.AvgScores[name1])
	} else {
		for _, name := range []string{name1, name2} {
			fmt.Printf("  %-12s wins %4d   mean score %7.2f\n", name, br.Wins[name], br.AvgScores[name])
		}
		fmt.Printf("  ties %d; wins by seat: first %d, second %d\n", br.Ties, br.SeatWins[0], br.SeatWins[1])
	}
	fmt.Printf("  moves %d, abstentions %d, rejected %d; mean %.1f rounds and %.1f arguments per game\n",
		br.TotalMoves, br.TotalAbstains, br.TotalInvalid, br.AvgTurns, br.AvgArguments)
	printCounts("ended by", br.Reasons)
	printCounts("verdict", br.Verdicts)
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

// printCounts prints the counts of a statistic, most frequent first.
func printCounts(what string, m map[string]int) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return m[keys[i]] > m[keys[j]] || m[keys[i]] == m[keys[j]] && keys[i] < keys[j] })
	for _, k := range keys {
		fmt.Printf("  %s %-16s %d\n", what, k, m[k])
	}
}
