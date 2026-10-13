package game

import (
	"encoding/json"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
)

// newShippedGame starts a game on the shipped data with the given NASCQ edges.
func newShippedGame(t *testing.T, edgesJSON []byte, players ...string) *GameState {
	t.Helper()
	ad, err := carneades.NewAdapterFromYAML(data.CoreYAML)
	if err != nil {
		t.Fatal(err)
	}
	n, err := nascq.LoadFromBytes(data.SchemesJSON, data.CQsJSON, edgesJSON)
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultGameConfig()
	cfg.MaxTurns = 12
	return NewGame(cfg, ad, n, players)
}

// everyCQAsksEverySchemeEdges keeps the shipped backup edges and asks every
// critical question of every scheme. Play on it reaches structures the shipped
// edges do not yet allow (a rebuttal of a rebuttal, an attack on an argument
// made with a thesis-independent scheme), which the NASCQ will allow once it
// grows.
func everyCQAsksEverySchemeEdges(t *testing.T) []byte {
	t.Helper()
	var schemes map[string]json.RawMessage
	var cqs map[string]json.RawMessage
	var edges struct {
		Ask    []map[string]string `json:"ask_edges"`
		Backup []map[string]string `json:"backup_edges"`
	}
	for _, x := range []struct {
		b []byte
		v interface{}
	}{{data.SchemesJSON, &schemes}, {data.CQsJSON, &cqs}, {data.EdgesJSON, &edges}} {
		if err := json.Unmarshal(x.b, x.v); err != nil {
			t.Fatal(err)
		}
	}
	edges.Ask = nil
	for _, s := range carneades.SortedKeys(schemes) {
		for _, q := range carneades.SortedKeys(cqs) {
			edges.Ask = append(edges.Ask, map[string]string{"cq": q, "scheme": s})
		}
	}
	out, err := json.Marshal(edges)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// playCheckingInvariants plays a game of random agents and checks the
// invariants after every accepted move, and that every rejected move leaves
// the graph exactly as it was.
func playCheckingInvariants(t *testing.T, gs *GameState, agents []Agent) {
	t.Helper()
	for it := 0; !gs.Terminated && it < 2000; it++ {
		before := gs.Adapter.Fingerprint()
		act := agents[gs.ActivePlayerIdx].ChooseAction(gs.BuildObservation())
		r, err := gs.ExecuteTurn(act)
		if err != nil {
			t.Fatal(err)
		}
		if !r.Accepted {
			if after := gs.Adapter.Fingerprint(); after != before {
				t.Fatalf("rejected move %+v (%s) changed the graph", act, r.Error)
			}
			if _, err := gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"}); err != nil {
				t.Fatal(err)
			}
		}
		if errs := gs.CheckInvariants(); len(errs) > 0 {
			t.Fatalf("after move %d (%+v): %v", len(gs.MoveHistory), act, errs)
		}
	}
}

func TestInvariantsHoldInRandomPlay(t *testing.T) {
	fixtures := []struct {
		name  string
		edges []byte
	}{
		{"shipped edges", data.EdgesJSON},
		{"every CQ asks every scheme", everyCQAsksEverySchemeEdges(t)},
	}
	games := 60
	if testing.Short() {
		games = 10
	}
	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			for i := 0; i < games; i++ {
				gs := newShippedGame(t, f.edges, "A", "B", "C")
				agents := []Agent{
					NewRandomAgent("A", int64(i), 0.05, 0.3),
					NewRandomAgent("B", int64(i+1000), 0.05, 0.3),
					NewRandomAgent("C", int64(i+2000), 0.05, 0.3),
				}
				playCheckingInvariants(t, gs, agents)
			}
		})
	}
}

// The batch statistics add up: wins by seat plus ties is the number of games,
// the move kinds count every accepted move.
func TestBatchStatisticsAddUp(t *testing.T) {
	br := RunBatch(6, DefaultGameConfig(), DefaultRunnerConfig(), data.CoreYAML, data.SchemesJSON, data.CQsJSON, data.EdgesJSON,
		func(i int) []Agent {
			return []Agent{NewRandomAgent("A", int64(i), 0.1, 0.5), NewRandomAgent("B", int64(i+99), 0.1, 0.5)}
		})
	seats, kinds := br.Ties, 0
	for _, n := range br.SeatWins {
		seats += n
	}
	for _, n := range br.MoveKinds {
		kinds += n
	}
	if br.Panics != 0 || seats != br.GamesPlayed || kinds != br.TotalMoves || len(br.AvgScores) != 2 {
		t.Fatalf("panics %d, seat wins + ties %d of %d games, move kinds %d of %d moves, %d mean scores",
			br.Panics, seats, br.GamesPlayed, kinds, br.TotalMoves, len(br.AvgScores))
	}
}
