package game

import (
	"testing"
)

// Supplied by the non-author reviewer: profile-specific audience weights
// cannot silently be applied to an unknown recorded historical identity.
func TestStage2HistoricalSchemeKeepsOriginalWeight(t *testing.T) {
	gs := newShippedGame(t, prevEdges, "Alice", "Bob")
	id := play(t, gs, prSupportAction())
	arg := gs.Adapter.Graph.Arguments[id]
	gs.Claim = arg.Conclusion.Id
	weights := map[string]float64{"AS:practical_reasoning": 0}
	if got := gs.ClaimUnder(weights); got != "out" {
		t.Fatalf("positive control: current weight was not applied: %s", got)
	}
	arg.Metadata["nascq_scheme_id"] = "AS:prior_profile_practical"
	if got := gs.ClaimUnder(weights); got != "in" {
		t.Fatalf("unknown historical identity acquired current profile weight: %s", got)
	}
	delete(arg.Metadata, "nascq_scheme_id")
	if got := gs.ClaimUnder(weights); got != "out" {
		t.Fatalf("legacy unambiguous identity lost its applicable weight: %s", got)
	}
}
