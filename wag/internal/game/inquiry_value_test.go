package game

import (
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// X4 (docs/e1-evidence.md, E1.7, A4; the outcome written before the test
// ran): in an inquiry, the greedy value of each player is its score less the
// other's, whether the claim is In or Out: the verdict favours neither.
func TestX4AGreedyBotTakesNoSideInAnInquiry(t *testing.T) {
	gs := modeGame(t, "inquiry", "A", "B")
	a1 := play(t, gs, prSupportAction())
	claim := gs.Adapter.Graph.Arguments[a1].Conclusion
	scores := func() (float64, float64) {
		var a, b float64
		for _, p := range gs.Players {
			if p.ID == "A" {
				a = p.Score
			} else {
				b = p.Score
			}
		}
		return a, b
	}
	was := claim.Label
	defer func() { claim.Label = was }()
	for _, l := range []caes.Label{caes.In, caes.Out} {
		claim.Label = l
		a, b := scores()
		if va, vb := Value(gs, "A"), Value(gs, "B"); va != a-b || vb != b-a {
			t.Fatalf("X4: with the claim %v, Value(A) %v and Value(B) %v; want %v and %v", l, va, vb, a-b, b-a)
		}
	}
}

// X7 (E1.7, B3), in the game: a game without a record set plays a move
// carrying the evidence fields without them, and refuses an act's type as
// an unknown type (S0), as before E1.
func TestX7AGameWithoutARecordSetReadsNoEvidenceField(t *testing.T) {
	gs := modeGame(t, "debate", "Pro", "Con")
	a := prSupportAction()
	a.BindingID, a.Revision, a.EvidenceRefs = "game:b2", "2", []string{"game:b2"}
	a.Offer = &nascq.EvidenceOffer{ClaimID: "claim-quote"}
	play(t, gs, a)
	if h := gs.MoveHistory[0].Action; h.BindingID != "" || h.Revision != "" || h.Offer != nil || h.EvidenceRefs != nil {
		t.Fatalf("X7: the recorded action %+v", h)
	}
	r, err := gs.ExecuteTurn(nascq.ProposedAction{Type: ActOffer})
	if err != nil || r.Accepted || !strings.Contains(r.Error, "S0: unknown action type") {
		t.Fatalf("X7: an offer in a debate: %+v %v", r, err)
	}
}
