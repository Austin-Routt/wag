package game

import (
	"sort"
	"strings"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/yourorg/wag/internal/nascq"
)

// The independent check of F1, F2 and F4 (docs/c5-item-f.md, section K): the
// engine keys a statement by its formula read with the negation identity, and
// the game names it so.

// claimGame is item F's board with an objective of the claim given.
func claimGame(t *testing.T, claim string) *fgame {
	t.Helper()
	g := newFGame(t)
	cfg := g.gs.Config
	cfg.Objective, cfg.Claim = ObjectiveClaim, claim
	g.gs = NewGame(cfg, g.gs.Adapter, g.gs.NASCQ, []string{"Pro", "Con"})
	return g
}

// K2: NewGame makes a claim written otherwise than the engine keys it under
// the key, a position of its issue.
func TestK2TheClaimAsTheEngineKeysIt(t *testing.T) {
	for _, c := range []struct{ claim, key string }{{"¬ p", "¬p"}, {"¬(p)", "¬p"}, {"¬¬p", "p"}, {"¬ ¬ p", "p"},
		{"¬¬" + worthDF, worthDF}} {
		g := claimGame(t, c.claim)
		if g.gs.Claim != c.key {
			t.Errorf("K2 %s: the claim is %q, want %q", c.claim, g.gs.Claim, c.key)
			continue
		}
		if s := g.gs.Adapter.Graph.Statements[c.key]; s == nil || s.Id != c.key || s.Issue == nil {
			t.Errorf("K2 %s: the statement %q is %+v", c.claim, c.key, s)
		}
		if errs := g.gs.CheckInvariants(); len(errs) > 0 {
			t.Errorf("K2 %s: %v", c.claim, errs)
		}
	}
}

// K5 and K5b: with the claim written with two negations, S5 reads the claim as
// the engine keys it, the bridges' exception: sign against its negation and
// Other for it are refused, before a bridge and after one.
func TestK5ADoublyNegatedClaimIsTheException(t *testing.T) {
	signAgainst := about("attack", "AS:sign", "¬"+worthDF, vb("A", "a_pilot_is_planned"))
	otherFor := support(nascq.OtherSchemeID, vb("C", worthDF, "P1", "a pilot is planned"))
	for _, claim := range []string{"¬¬" + worthDF, "¬(¬" + worthDF + ")"} {
		for _, a := range []nascq.ProposedAction{signAgainst, otherFor} {
			g := claimGame(t, claim)
			g.refuse("Con", a, "S5")
			if !strings.Contains(g.gs.LastRejection, nascq.FormOnly) {
				t.Errorf("K5 %s, %s: no form-only note: %s", claim, a.SchemeID, g.gs.LastRejection)
			}
		}
	}
	for _, a := range []nascq.ProposedAction{signAgainst, otherFor} {
		g := claimGame(t, "¬¬"+worthDF)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		g.refuse("Con", a, "S5")
		g.inForce(bp)
	}
}

// K7 and K7b: in a game with no profile, the claim's variable is in the
// bridge's exception; S5 bounds that statement, after the bridge and before
// one (each variable read as a constant of its own).
func TestK7AVariableInTheException(t *testing.T) {
	const claim, exc = "should_be_performed(X)", "worth_finding_out_first(X,demand_falls)"
	signFor := about("support", "AS:sign", exc, vb("A", "a_pilot_is_planned"))
	newGame := func(t *testing.T) *fgame {
		g := newFGameNoProfile(t)
		if _, err := g.gs.Adapter.MakeIssue(claim, caes.PE); err != nil {
			t.Fatal(err)
		}
		return g
	}
	// A board whose keys hold a variable fails the game's admission check
	// (play's), which a game with a profile never meets: K7's moves are made
	// without it (docs/c5-item-f.md, K's correction).
	unchecked := func(g *fgame, player string, a nascq.ProposedAction) string {
		g.turnOf(player)
		r, err := g.gs.ExecuteTurn(a)
		if err != nil || !r.Accepted || len(r.ArgIDs) == 0 {
			t.Fatalf("%s %s: %v %q", a.Type, a.SchemeID, err, g.gs.LastRejection)
		}
		return r.ArgIDs[0]
	}
	t.Run("K7", func(t *testing.T) {
		g := newGame(t)
		b := unchecked(g, "Pro", about("support", "AS:bridge_possible_good", claim, vb("G", "demand_falls")))
		var prem string
		for _, p := range g.gs.Adapter.Graph.Arguments[b].Premises {
			if strings.HasPrefix(p.Stmt.Id, "may_be_worth_performing(") {
				prem = p.Stmt.Id
			}
		}
		unchecked(g, "Pro", about("support", "AS:consequences_modal", prem, vb()))
		g.inForce(b)
		g.refuse("Con", signFor, "S5: "+exc+" is the exception")
		g.inForce(b)
	})
	t.Run("K7b", func(t *testing.T) {
		g := newGame(t)
		g.gs.Adapter.AddStatement(exc, exc)
		g.refuse("Con", signFor, "S5: "+exc+" is the exception")
		if !strings.Contains(g.gs.LastRejection, nascq.FormOnly) {
			t.Errorf("K7b: no form-only note: %s", g.gs.LastRejection)
		}
	})
}

// A9.8d and A9.11b (docs/c5-item-f.md, added from the check's W3 and N5): Other
// concluding the exception as a direct move, which only S5's second layer
// refuses before any bridge; and the moves offered against the exception's
// negation.
func TestItemFExceptionFormSecondLayer(t *testing.T) {
	t.Run("A9.8d", func(t *testing.T) {
		g := newFGame(t)
		g.gs.Adapter.AddStatement(worthDF, worthDF)
		g.refuse("Con", support(nascq.OtherSchemeID, vb("C", worthDF, "P1", "a pilot is planned")), "S5: an argument of Other cannot conclude "+worthDF)
		if !strings.Contains(g.gs.LastRejection, nascq.FormOnly) {
			t.Errorf("A9.8d: no form-only note: %s", g.gs.LastRejection)
		}
	})
	t.Run("A9.11b", func(t *testing.T) {
		neg := "¬" + worthDF
		g := claimGame(t, neg)
		var against []string
		for _, a := range g.gs.NASCQ.EnumerateStatementActions(g.gs.Bridge(), []string{neg}) {
			if a.TargetStmtID == neg && a.Type == "attack" {
				against = append(against, a.SchemeID)
			}
		}
		sort.Strings(against)
		if got, want := strings.Join(against, " "), "AS:expert_opinion AS:position_to_know"; got != want {
			t.Errorf("A9.11b: against %s: %q, want %q", neg, got, want)
		}
	})
}
