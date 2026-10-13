package game

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
)

// snapshotOf renders everything Simulate must restore.
func snapshotOf(gs *GameState) string {
	s := fmt.Sprintf("%s|turn %d|active %d|abstain %d %v|term %v %q|digest %s|hist %d\n",
		gs.Adapter.Fingerprint(), gs.CurrentTurn, gs.ActivePlayerIdx, gs.ConsecAllAbstain, gs.roundHadMove,
		gs.Terminated, gs.TerminationReason, gs.MoveDigest(), len(gs.MoveHistory))
	for _, p := range gs.Players {
		s += fmt.Sprintf("%s %.6f %d %v\n", p.ID, p.Score, p.ConsecutiveAbstentions, p.SchemeUsageCounts)
	}
	labels := gs.Adapter.StatementLabels()
	for _, id := range carneades.SortedKeys(labels) {
		s += fmt.Sprintf("%s=%v ", id, labels[id])
	}
	for _, id := range sortedArgIDs(gs) {
		s += fmt.Sprintf("%s:%.4f ", id, gs.Adapter.Graph.Arguments[id].Weight)
	}
	return s
}

// Simulating any enumerated action leaves the game exactly as it was, and the
// simulated result is what playing the action then gives.
func TestSimulateRestoresTheGameExactly(t *testing.T) {
	games := 8
	if testing.Short() {
		games = 2
	}
	for game := 0; game < games; game++ {
		gs := newShippedGame(t, everyCQAsksEverySchemeEdges(t), "A", "B")
		agents := []Agent{NewRandomAgent("A", int64(game), 0.05, 0.4), NewRandomAgent("B", int64(game+50), 0.05, 0.4)}
		for move := 0; !gs.Terminated && move < 16; move++ {
			obs := gs.BuildObservation()
			before := snapshotOf(gs)
			for i, tmpl := range obs.AvailableActions {
				if i%3 != move%3 { // a third of the templates per move keeps this quick
					continue
				}
				b := map[string]string{}
				for v := range tmpl.Variables {
					b[v] = fmt.Sprintf("w%d", i)
				}
				if _, err := gs.Simulate(fromTemplate(tmpl, b)); err != nil {
					t.Fatal(err)
				}
				if after := snapshotOf(gs); after != before {
					t.Fatalf("game %d move %d: simulating %+v changed the game", game, move, tmpl)
				}
			}
			act := agents[gs.ActivePlayerIdx].ChooseAction(obs)
			sim, err := gs.Simulate(act)
			if err != nil {
				t.Fatal(err)
			}
			real, err := gs.ExecuteTurn(act)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(sim, real) {
				t.Fatalf("simulated %+v, played %+v", sim, real)
			}
			if !real.Accepted {
				gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
			}
			if errs := gs.CheckInvariants(); len(errs) > 0 {
				t.Fatal(errs)
			}
		}
	}
}
