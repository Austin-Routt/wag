package kialo

import (
	"fmt"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"
)

// Players of an imported debate.
const (
	ProPlayer = "Pro"
	ConPlayer = "Con"
)

// Config is the game an imported discussion becomes: a debate on its thesis
// (pro first, live scoring, no round limit), with the user-defined scheme
// allowed, since every imported argument is one.
func Config() game.GameConfig {
	c, _ := game.ModeConfig("debate")
	c.Mode = "kialo"
	c.MaxTurns = 0
	c.AllowOther = true
	return c
}

// Build makes the game: one statement per claim, keyed by its claim ID
// (rep.Keys); one Other argument per relation, owned by the side its claim
// argues for, with Weight from its impact votes, its premise (the claim)
// presumed; and the thesis as the debate's claim, at issue. Players
// continue from there. The imported arguments are not moves: nobody scored
// them.
func Build(d *Discussion, rels []Relation, rep *Report, cfg game.GameConfig, ad *carneades.Adapter, n *nascq.NASCQ) (*game.GameState, error) {
	gs := game.NewGame(cfg, ad, n, []string{ProPlayer, ConPlayer})
	stmt := map[string]string{}
	statementOf := func(claimID string) string {
		if id, ok := stmt[claimID]; ok {
			return id
		}
		var id string
		if rep.Keys == KeysByText {
			id = ad.AddFreeStatement(d.Claims[claimID].Text)
		} else {
			id = ad.AddClaimStatement(claimID, d.Claims[claimID].Text)
		}
		stmt[claimID] = id
		return id
	}
	thesis := statementOf(d.ThesisClaimID)
	owner := map[string]string{"pro": ProPlayer, "con": ConPlayer}
	for _, r := range rels {
		parent := statementOf(r.From)
		child := statementOf(r.To)
		if parent == child {
			continue // KeysByText: the same text pro or con itself
		}
		ids, err := ad.AddFreeArgumentOn(owner[r.Side], []string{child}, parent, r.Relation == "attack", r.Weight)
		if err != nil {
			return nil, fmt.Errorf("relation %s: %v", r.LocationID, err)
		}
		ad.SetArgumentMeta(ids[0], "kialo_location", r.LocationID)
		ad.SetArgumentMeta(ids[0], "kialo_claim", r.To)
		ad.SetArgumentMeta(ids[0], "text", d.Claims[r.To].Text)
		if _, err := ad.AddPresumption(owner[r.Side], child, cfg.EffectivePresumptionWeight()); err != nil {
			return nil, err
		}
		rep.Arguments++
	}
	claim, err := ad.MakeIssue(thesis, cfg.ClaimStandard)
	if err != nil {
		return nil, err
	}
	gs.Claim = claim
	if !ad.IsAcyclic() {
		return nil, fmt.Errorf("the imported graph has a cycle")
	}
	ad.Evaluate()
	return gs, nil
}
