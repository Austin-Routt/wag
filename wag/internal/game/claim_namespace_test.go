package game

import (
	"errors"
	"testing"
)

// U2 and U3 (docs/c5-item-f.md): NewGame makes no claim in the engine's
// undercutter namespace, read as the engine reads it with its negations taken
// off, and makes a claim of another form.
func TestU2U3NewGameAndTheUndercuttersNamespace(t *testing.T) {
	for _, claim := range []string{"undercut(a11)", "undercut( a11 )", "undercut (a11)", "¬undercut(a11)", "¬¬undercut(a11)", "¬(undercut(a11))"} {
		if err := CheckClaimNamespace(claim); !errors.Is(err, ErrReservedClaim) {
			t.Errorf("%s: %v, want the reserved claim refused", claim, err)
		}
		g := newFGame(t)
		cfg := g.gs.Config
		cfg.Objective, cfg.Claim = ObjectiveClaim, claim
		g.gs = NewGame(cfg, g.gs.Adapter, g.gs.NASCQ, []string{"Pro", "Con"})
		if g.gs.Claim != "" || g.gs.Adapter.Graph.Statements["undercut(a11)"] != nil {
			t.Fatalf("%s: the claim was made: %q", claim, g.gs.Claim)
		}
		g.refuse("Pro", about("support", "AS:sign", "undercut(a11)", vb("A", "a_pilot_is_planned")), "S7")
	}
	for _, claim := range []string{"undercut_cost(a11)", "undercut(a11,b)", "undercut", worthDF} {
		if err := CheckClaimNamespace(claim); err != nil {
			t.Errorf("%s: %v", claim, err)
		}
		g := newFGame(t)
		cfg := g.gs.Config
		cfg.Objective, cfg.Claim = ObjectiveClaim, claim
		g.gs = NewGame(cfg, g.gs.Adapter, g.gs.NASCQ, []string{"Pro", "Con"})
		if g.gs.Claim == "" {
			t.Errorf("%s: the claim was not made", claim)
		}
	}
}
