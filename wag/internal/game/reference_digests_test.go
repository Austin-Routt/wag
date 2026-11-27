package game

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// Reference games for an engine change.
//
// referenceGames is a fixed table of seeded computer-against-computer games in
// the five modes that do not depend on the wall clock. The move digest, the
// scores, the number of moves and the way each game ended are stored in
// testdata/reference-digests.txt. A different build of Carneades (the fork, the
// schemes library's engine, a later version of either) must reproduce every
// line for these bounded checks; the digest is not a complete experiment identity.
//
// The file was written on WAG 7f784b0 with the fork (carneades-4 995768e) and
// reproduced with the schemes library's engine 4.3.1+walton.8 (024b86b).
// WAG_UPDATE_REFERENCE=1 go test ./internal/game -run TestReferenceDigests
// rewrites it; read the diff before committing it.

type referenceGame struct {
	mode   string
	claim  string // debate only
	turns  int    // maximum rounds
	p1, p2 string // "random" or "greedy"
	s1, s2 int64
}

func referenceGames() []referenceGame {
	var gs []referenceGame
	for seed := int64(1); seed <= 8; seed++ {
		gs = append(gs, referenceGame{"classic", "", 12, "random", "random", seed, seed + 100})
	}
	for seed := int64(1); seed <= 4; seed++ {
		gs = append(gs,
			referenceGame{"classic", "", 6, "greedy", "random", seed, seed + 100},
			referenceGame{"infinite", "", 10, "random", "random", seed, seed + 100},
			referenceGame{"target", "", 12, "random", "random", seed, seed + 100},
			referenceGame{"cq", "", 12, "random", "random", seed, seed + 100},
			referenceGame{"debate", "should_be_performed(carbon_tax)", 10, "random", "random", seed, seed + 100},
		)
	}
	gs = append(gs,
		referenceGame{"classic", "", 5, "greedy", "greedy", 1, 2},
		referenceGame{"infinite", "", 5, "greedy", "random", 3, 4},
		referenceGame{"target", "", 5, "random", "greedy", 5, 6},
		referenceGame{"cq", "", 5, "greedy", "random", 7, 8},
		referenceGame{"debate", "should_be_performed(carbon_tax)", 5, "greedy", "greedy", 9, 10},
	)
	return gs
}

func (g referenceGame) agent(name, kind string, seed int64) Agent {
	if kind == "greedy" {
		return NewGreedyAgent(name, seed, 2)
	}
	return NewRandomAgent(name, seed, 0.15, 0.55) // wag-cli's random player
}

func (g referenceGame) play(t *testing.T) string {
	t.Helper()
	return g.playOn(t, []byte(coreYAML), []byte(coreSchemesJSON), []byte(coreCQsJSON), []byte(coreEdgesJSON), nil)
}

// playOn plays the game on the data given, with the profile given (nil: none,
// as games made before step C5.2).
func (g referenceGame) playOn(t *testing.T, core, schemes, cqs, edges []byte, profile *Profile) string {
	t.Helper()
	cfg, err := ModeConfig(g.mode)
	if err != nil {
		t.Fatal(err)
	}
	cfg.MaxTurns = g.turns
	if g.claim != "" {
		cfg.Objective, cfg.Claim = ObjectiveClaim, g.claim
	}
	cfg.Profile = profile
	r := RunGame(cfg, DefaultRunnerConfig(), core, schemes, cqs, edges,
		[]Agent{g.agent("Alice", g.p1, g.s1), g.agent("Bob", g.p2, g.s2)})
	if r.Panicked {
		t.Fatalf("%+v panicked: %s", g, r.PanicMessage)
	}
	return fmt.Sprintf("%s turns=%d %s/%d-%s/%d moves=%d rejected=%d alice=%.4f bob=%.4f winner=%s verdict=%q reason=%q digest=%s",
		g.mode, g.turns, g.p1, g.s1, g.p2, g.s2, r.MoveCount, r.InvalidCount, r.Scores["Alice"], r.Scores["Bob"],
		r.Winner, r.Verdict, reasonKind(r.Reason), r.MoveDigest)
}

func TestReferenceDigests(t *testing.T) {
	const path = "testdata/reference-digests.txt"
	var got []string
	for _, g := range referenceGames() {
		got = append(got, g.play(t))
	}
	if os.Getenv("WAG_UPDATE_REFERENCE") != "" {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(strings.Join(got, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d reference games to %s", len(got), path)
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (WAG_UPDATE_REFERENCE=1 writes it)", err)
	}
	// A checkout with CRLF line endings (Windows) reads the same digests.
	want := strings.Split(strings.TrimRight(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n"), "\n")
	if len(want) != len(got) {
		t.Fatalf("%d reference games stored, %d played", len(want), len(got))
	}
	differ := 0
	for i := range got {
		if got[i] != want[i] {
			differ++
			t.Errorf("game %d differs from the reference\n got  %s\n want %s", i+1, got[i], want[i])
		}
	}
	if differ == 0 {
		t.Logf("%d reference games reproduced", len(got))
	}
}
