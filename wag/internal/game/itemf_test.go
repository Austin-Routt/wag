package game

// Item F (docs/c5-item-f.md): the answer contract per route, on the data the
// game ships. Each case's outcome was written in that page, and committed,
// before its test ran; the cases are named as there.

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
)

// fgame is item F's board: the shipped data with Other on, any premise open
// to challenge, no turn limit, Pro and Con, and the profile of the page.
type fgame struct {
	t  *testing.T
	gs *GameState
}

func itemfProfile() *Profile {
	p := DefaultProfile()
	p.Actors = append(p.Actors, ProfileActor{ID: "farmers", Text: "farmers"})
	p.Actions = []ProfileAction{
		{ID: "tax_meat", Actor: DefaultAgent, Aliases: []string{"meat_tax"}},
		{ID: "subsidy_plants", Actor: DefaultAgent},
		{ID: "farmers_strike", Actor: "farmers"},
	}
	return p
}

func newFGame(t *testing.T) *fgame {
	t.Helper()
	ad, err := carneades.NewAdapterFromYAML(data.CoreYAML)
	if err != nil {
		t.Fatal(err)
	}
	n, err := nascq.LoadFromBytes(data.SchemesJSON, data.CQsJSON, data.EdgesJSON)
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultGameConfig()
	cfg.MaxTurns = 0
	cfg.AbstentionThreshold = 1000
	cfg.AllowOther = true
	cfg.AnyPremise = true
	cfg.Profile = itemfProfile()
	return &fgame{t: t, gs: NewGame(cfg, ad, n, []string{"Pro", "Con"})}
}

// turnOf passes turns until the player named is to move.
func (g *fgame) turnOf(player string) {
	g.t.Helper()
	for i := 0; g.gs.ActivePlayer().ID != player; i++ {
		if i > 2 {
			g.t.Fatalf("%s never gets the turn", player)
		}
		if r, err := g.gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"}); err != nil || !r.Accepted {
			g.t.Fatalf("a pass: %v %+v", err, r)
		}
	}
}

// play makes a move that must be accepted and returns its main argument.
func (g *fgame) play(player string, a nascq.ProposedAction) string {
	g.t.Helper()
	g.turnOf(player)
	r, err := g.gs.ExecuteTurn(a)
	if err != nil || !r.Accepted {
		g.t.Fatalf("%s: %s %s %s (%v) refused: %v %q", player, a.Type, a.AnswersCQID, a.SchemeID, a.VariableBindings, err, g.gs.LastRejection)
	}
	if errs := g.gs.CheckInvariants(); len(errs) > 0 {
		g.t.Fatalf("after %s's %s %s: %v", player, a.Type, a.AnswersCQID, errs)
	}
	if len(r.ArgIDs) == 0 {
		return ""
	}
	return r.ArgIDs[0]
}

// refuse makes a move that must be refused by the rule named (the start of
// the refusal), leaving the game as state() renders it, and the turn, as
// they were: a rule's refusal (S, P, Q, G) adds one record to the move
// history, with its penalty, and the adapter's and the cycle check's add
// none.
func (g *fgame) refuse(player string, a nascq.ProposedAction, rule string) {
	g.t.Helper()
	g.turnOf(player)
	before, n := g.state(), len(g.gs.MoveHistory)
	r, err := g.gs.ExecuteTurn(a)
	if err != nil {
		g.t.Fatalf("%s: %s %s %s: %v", player, a.Type, a.AnswersCQID, a.SchemeID, err)
	}
	if r.Accepted || !strings.HasPrefix(g.gs.LastRejection, rule) {
		g.t.Fatalf("%s: %s %s %s (%v): accepted %v, %q; want a refusal by %s", player, a.Type, a.AnswersCQID, a.SchemeID,
			a.VariableBindings, r.Accepted, g.gs.LastRejection, rule)
	}
	if after := g.state(); after != before {
		g.t.Fatalf("the refusal changed the game:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	// A rule's refusal (S, P, Q, G) is recorded once, with its penalty; the
	// adapter's refusal and the cycle check's leave no record and cost
	// nothing (the independent check of step 7, W-3).
	added := g.gs.MoveHistory[n:]
	if ruleRefusal(rule) {
		if len(added) != 1 || added[0].Rejection == "" || added[0].ScoreDelta != g.gs.Config.Scoring.InvalidActionPenalty {
			g.t.Fatalf("a rule's refusal is recorded once, with its penalty: %+v", added)
		}
	} else if len(added) != 0 {
		g.t.Fatalf("the refusal recorded %+v", added)
	}
	if g.gs.ActivePlayer().ID != player {
		g.t.Fatalf("the refusal passed the turn")
	}
}

// state is the game as a refusal must leave it: every statement with its
// label and text, the issues, the assumptions, the questioned statements,
// the open questions, every argument with its conclusion, premises,
// undercutter and weight, the profile, and the number of accepted moves.
func (g *fgame) state() string {
	var b strings.Builder
	gr := g.gs.Adapter.Graph
	ids := make([]string, 0, len(gr.Statements))
	for id := range gr.Statements {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		fmt.Fprintf(&b, "s %s %v %q\n", id, gr.Statements[id].Label, gr.Statements[id].Text)
	}
	var issues []string
	for id, is := range gr.Issues {
		var pos []string
		for _, p := range is.Positions {
			pos = append(pos, p.Id)
		}
		issues = append(issues, fmt.Sprintf("%s %v %v", id, pos, is.Standard))
	}
	sort.Strings(issues)
	assumed := append([]string(nil), gr.Assumptions...)
	sort.Strings(assumed)
	questioned := append([]string(nil), gr.Questioned...)
	sort.Strings(questioned)
	fmt.Fprintf(&b, "issues %v\nassumptions %v\nquestioned %v\n", issues, assumed, questioned)
	for _, q := range g.gs.Questions {
		fmt.Fprintf(&b, "question %+v\n", q)
	}
	ids = ids[:0]
	for id := range gr.Arguments {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		a := gr.Arguments[id]
		var prem []string
		for _, p := range a.Premises {
			if p.Stmt != nil {
				prem = append(prem, p.Stmt.Id)
			}
		}
		uc := ""
		if a.Undercutter != nil {
			uc = a.Undercutter.Id
		}
		fmt.Fprintf(&b, "a %s -> %s %v uc=%s w=%v\n", id, a.Conclusion.Id, prem, uc, a.Weight)
	}
	pj, _ := json.Marshal(g.gs.Profile)
	accepted := 0
	for _, m := range g.gs.MoveHistory {
		if m.Rejection == "" {
			accepted++
		}
	}
	fmt.Fprintf(&b, "profile %s\naccepted moves %d\n", pj, accepted)
	return b.String()
}

func (g *fgame) label(stmt string) caes.Label {
	if s := g.gs.Adapter.Graph.Statements[stmt]; s != nil {
		return s.Label
	}
	return caes.Out
}

// in reports whether the statement is In.
func (g *fgame) in(stmt string) bool { return g.label(stmt) == caes.In }

// mustIn and mustNotIn check a statement's label.
func (g *fgame) mustIn(stmt string) {
	g.t.Helper()
	if !g.in(stmt) {
		g.t.Fatalf("%s is %v, want In", stmt, g.label(stmt))
	}
}

func (g *fgame) mustNotIn(stmt string) {
	g.t.Helper()
	if g.in(stmt) {
		g.t.Fatalf("%s is In", stmt)
	}
}

// mustOut checks that a statement is on the board and Out (not Undecided).
func (g *fgame) mustOut(stmt string) {
	g.t.Helper()
	if s := g.gs.Adapter.Graph.Statements[stmt]; s == nil || s.Label != caes.Out {
		g.t.Fatalf("%s is %v, want Out", stmt, g.label(stmt))
	}
}

// ruleRefusal tells whether a refusal is a rule's (S0 to S9, P1 to P5, Q1,
// Q2, G1), which the history records with its penalty.
func ruleRefusal(rule string) bool {
	return len(rule) >= 2 && strings.ContainsRune("SPQG", rune(rule[0])) && rule[1] >= '0' && rule[1] <= '9'
}

// answeredBy checks that the questions named, asked of the argument, are
// closed by the argument given.
func (g *fgame) answeredBy(target, by string, cqs ...string) {
	g.t.Helper()
	found := 0
	for _, q := range g.gs.Questions {
		for _, cq := range cqs {
			if q.TargetArgID == target && q.CQID == cq {
				found++
				if q.AnsweredBy != by {
					g.t.Fatalf("%s of %s is answered by %q, not %s", cq, target, q.AnsweredBy, by)
				}
			}
		}
	}
	if found != len(cqs) {
		g.t.Fatalf("%d of the questions %v were asked of %s", found, cqs, target)
	}
}

// inForce and outOfForce check an argument's weight (standing).
func (g *fgame) inForce(names ...string) {
	g.t.Helper()
	for _, id := range names {
		if !g.gs.standing(id) {
			g.t.Fatalf("argument %s is out of force", id)
		}
	}
}

func (g *fgame) outOfForce(names ...string) {
	g.t.Helper()
	for _, id := range names {
		if g.gs.Adapter.Graph.Arguments[id] == nil {
			g.t.Fatalf("argument %s is not on the board", id)
		}
		if g.gs.standing(id) {
			g.t.Fatalf("argument %s is in force", id)
		}
	}
}

func (g *fgame) conclusion(arg string) string {
	g.t.Helper()
	a := g.gs.Adapter.Graph.Arguments[arg]
	if a == nil {
		g.t.Fatalf("argument %s is not on the board", arg)
	}
	return a.Conclusion.Id
}

// The page's moves.
const (
	shouldTax = "should_be_performed(tax_meat)"
	brings1   = "brings_about(tax_meat,emissions_fall)"
	bad1      = "bad_outweighs(tax_meat,emissions_fall)"
	slight1   = "too_slight(tax_meat,emissions_fall)"
	worthDF   = "worth_finding_out_first(tax_meat,demand_falls)"
)

func vb(kv ...string) map[string]string {
	m := map[string]string{}
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return m
}

func support(scheme string, b map[string]string) nascq.ProposedAction {
	return nascq.ProposedAction{Type: "support", SchemeID: scheme, VariableBindings: b}
}

func about(typ, scheme, stmt string, b map[string]string) nascq.ProposedAction {
	return nascq.ProposedAction{Type: typ, SchemeID: scheme, TargetStmtID: stmt, VariableBindings: b}
}

func ask(cq, arg string) nascq.ProposedAction {
	return nascq.ProposedAction{Type: "question", AnswersCQID: cq, TargetArgID: arg}
}

func (g *fgame) c1() string {
	return g.play("Pro", support("AS:consequences", vb("A", "tax_meat", "G", "emissions_fall")))
}
func (g *fgame) c2() string {
	return g.play("Pro", support("AS:consequences", vb("A", "tax_meat", "G", "diets_healthier")))
}
func (g *fgame) c3() string {
	return g.play("Pro", support("AS:consequences", vb("A", "tax_meat", "G", "demand_falls")))
}
func (g *fgame) pr() string {
	return g.play("Pro", support("AS:practical_reasoning", vb("A", "tax_meat", "S1", "high_emissions", "S2", "low_emissions",
		"G", "emissions_fall", "V", "sustainability")))
}
func (g *fgame) pg() string {
	return g.play("Pro", support("AS:practical_reasoning_goal", vb("A", "tax_meat", "S1", "high_emissions", "S2", "low_emissions",
		"G", "emissions_fall")))
}
func (g *fgame) n1() string {
	return g.play("Con", support("AS:negative_consequences", vb("A", "tax_meat", "G", "food_costlier")))
}
func (g *fgame) n3() string {
	return g.play("Con", support("AS:negative_consequences", vb("A", "tax_meat", "G", "demand_falls")))
}
func (g *fgame) nm() string {
	return g.play("Con", support("AS:negative_consequences_modal", vb("A", "tax_meat", "G", "food_costlier")))
}
func (g *fgame) mPlus(gv string) string {
	return g.play("Pro", support("AS:consequences_modal", vb("A", "tax_meat", "G", gv)))
}
func (g *fgame) bPlus(gv string) string {
	return g.play("Pro", support("AS:bridge_possible_good", vb("A", "tax_meat", "G", gv)))
}
func (g *fgame) mMinus() string {
	return g.play("Con", support("AS:negative_consequences_modal", vb("A", "tax_meat", "G", "demand_falls")))
}
func (g *fgame) bMinus() string {
	return g.play("Con", support("AS:bridge_possible_harm", vb("A", "tax_meat", "G", "demand_falls")))
}

func betterMeans(pr string, b map[string]string, comparison string) nascq.ProposedAction {
	a := answer("attack", "CQ:other_means", "AS:better_means", pr, b)
	a.Comparison = comparison
	return a
}

// A. The backup edges.
func TestItemFBackupEdges(t *testing.T) {
	t.Run("A1.1", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		ans := g.play("Con", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", c1, vb("H", "food_costlier")))
		if got := g.conclusion(ans); got != bad1 {
			t.Fatalf("the answer concludes %s", got)
		}
		g.mustIn(bad1)
		g.outOfForce(c1)
		g.mustOut(shouldTax)
	})
	t.Run("A1.2", func(t *testing.T) {
		g := newFGame(t)
		c1, pr := g.c1(), g.pr()
		g.play("Con", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", pr, vb("H", "food_costlier")))
		g.mustIn(bad1)
		g.outOfForce(c1, pr)
		g.mustOut(shouldTax)
	})
	t.Run("A1.2b", func(t *testing.T) {
		// A1.2 from the other side: the answer at c1 reaches pr through the
		// exception the two share (the check of step 8, N-7).
		g := newFGame(t)
		c1, pr := g.c1(), g.pr()
		g.play("Con", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", c1, vb("H", "food_costlier")))
		g.mustIn(bad1)
		g.outOfForce(c1, pr)
		g.mustOut(shouldTax)
	})
	t.Run("A1.3", func(t *testing.T) {
		g := newFGame(t)
		c1, c2 := g.c1(), g.c2()
		g.play("Con", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", c1, vb("H", "food_costlier")))
		g.outOfForce(c1)
		g.inForce(c2)
		g.mustIn(shouldTax)
	})
	t.Run("A1.3b", func(t *testing.T) {
		g := newFGame(t)
		c1, n1 := g.c1(), g.n1()
		g.play("Con", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", c1, vb("H", "food_costlier")))
		g.outOfForce(c1)
		g.inForce(n1)
	})
	t.Run("A1.4", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.refuse("Con", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", c1, vb("H", "food_costlier", "G", "diets_healthier")), "S9")
	})
	t.Run("A1.5", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.refuse("Con", answer("attack", "CQ:side_effect", "AS:bad_outweighs_good", c1, vb("H", "food_costlier")), "S5")
		g.refuse("Con", answer("attack", "CQ:too_slight", "AS:bad_outweighs_good", c1, vb("H", "food_costlier")), "S5")
	})
	t.Run("A1.6", func(t *testing.T) {
		g := newFGame(t)
		g.c1()
		g.refuse("Con", support("AS:bad_outweighs_good", vb("A", "tax_meat", "H", "food_costlier", "G", "emissions_fall")), "S1")
		g.refuse("Con", about("support", "AS:bad_outweighs_good", bad1, vb("H", "food_costlier")), "S1")
	})
	t.Run("A1.7", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.refuse("Con", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", c1, vb()), "S2")
	})
	t.Run("A1.7b", func(t *testing.T) {
		// Added after the first run (G-unbound): a blank H names nothing.
		g := newFGame(t)
		c1 := g.c1()
		// A value with no letter or digit names nothing either (the check
		// of step 7, W-1): punctuation, a zero-width space, a byte order mark.
		for _, h := range []string{"", "  ", "?", "--", string(rune(0x200b)), string(rune(0xfeff))} {
			g.refuse("Con", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", c1, vb("H", h)), "S2")
		}
		g.refuse("Pro", support("AS:consequences", vb("A", "tax_meat", "G", " ")), "S2")
	})
	t.Run("A1.8", func(t *testing.T) {
		g := newFGame(t)
		n1 := g.n1()
		g.refuse("Pro", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", n1, vb("H", "emissions_fall")), "S3")
	})
	t.Run("A1.9", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.refuse("Con", answer("attack", "CQ:bad_outweighs", "AS:negative_consequences", c1, vb("A", "tax_meat", "G", "food_costlier")), "S5")
	})
	t.Run("A1.10", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		ans := g.play("Con", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", c1, vb("H", "food_costlier")))
		g.play("Pro", about("attack", nascq.OtherSchemeID, "matters_more(food_costlier,emissions_fall)",
			vb("P1", "lower emissions matter more than dearer food")))
		g.outOfForce(ans)
		g.mustOut(bad1)
		g.inForce(c1)
		g.mustIn(shouldTax)
		g.play("Con", ask("CQ:likelihood", c1))
	})

	t.Run("A2.1", func(t *testing.T) {
		g := newFGame(t)
		n1 := g.n1()
		g.play("Pro", answer("attack", "CQ:good_outweighs", "AS:good_outweighs_bad", n1, vb("H", "emissions_fall")))
		g.mustIn("good_outweighs(tax_meat,food_costlier)")
		g.outOfForce(n1)
	})
	t.Run("A2.2", func(t *testing.T) {
		g := newFGame(t)
		n1, nm := g.n1(), g.nm()
		g.play("Pro", answer("attack", "CQ:good_outweighs", "AS:good_outweighs_bad", n1, vb("H", "emissions_fall")))
		g.outOfForce(n1, nm)
	})
	t.Run("A2.3", func(t *testing.T) {
		g := newFGame(t)
		n1 := g.n1()
		g.refuse("Pro", answer("attack", "CQ:opposite_consequences", "AS:good_outweighs_bad", n1, vb("H", "emissions_fall")), "S5")
	})
	t.Run("A2.4", func(t *testing.T) {
		g := newFGame(t)
		n1 := g.n1()
		g.refuse("Pro", answer("attack", "CQ:good_outweighs", "AS:good_outweighs_bad", n1, vb("H", "emissions_fall", "G", "diets_healthier")), "S9")
	})
	t.Run("A2.5", func(t *testing.T) {
		g := newFGame(t)
		n1 := g.n1()
		g.refuse("Pro", answer("attack", "CQ:good_outweighs", "AS:consequences", n1, vb("A", "tax_meat", "G", "emissions_fall")), "S5")
	})

	t.Run("A3.1", func(t *testing.T) {
		g := newFGame(t)
		pr := g.pr()
		g.play("Con", betterMeans(pr, vb("A2", "subsidy_plants"), "sooner"))
		g.mustIn("realize_more_effectively(decider,emissions_fall,tax_meat)")
		g.outOfForce(pr)
	})
	for _, c := range []struct {
		name, rule, comparison string
		b                      map[string]string
	}{
		{"A3.2", "P4", "sooner", vb("A2", "tax_meat")},
		{"A3.3", "P4", "sooner", vb("A2", "meat_tax")},
		{"A3.4", "P3", "sooner", vb("A2", "farmers_strike")},
		{"A3.5", "P4", "", vb("A2", "subsidy_plants")},
		{"A3.6", "S9", "sooner", vb("A2", "subsidy_plants", "G", "diets_healthier")},
		{"A3.7", "P2", "sooner", vb("A2", "subsidy_plants", "P", "farmers")},
		{"A3.8", "S2", "sooner", vb()},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newFGame(t)
			pr := g.pr()
			g.refuse("Con", betterMeans(pr, c.b, c.comparison), c.rule)
		})
	}
	t.Run("A3.9", func(t *testing.T) {
		g := newFGame(t)
		pr, c1 := g.pr(), g.c1()
		g.play("Con", betterMeans(pr, vb("A2", "subsidy_plants"), "sooner"))
		g.outOfForce(pr)
		g.inForce(c1)
		g.mustIn(shouldTax)
	})

	t.Run("A4.1", func(t *testing.T) {
		g := newFGame(t)
		mp := g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		mm := g.mMinus()
		bm := g.bMinus()
		c1 := g.c1()
		g.play("Con", answer("attack", "CQ:worth_finding_out_first", "AS:finding_out_first", bp, vb("I", "pilot_study")))
		g.mustIn(worthDF)
		g.outOfForce(bp, bm)
		g.inForce(mp, mm, c1)
		g.mustIn("may_be_worth_performing(tax_meat,demand_falls)")
		g.mustIn("may_be_unwise_to_perform(tax_meat,demand_falls)")
		g.mustIn(shouldTax)
	})
	t.Run("A4.2", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		g.refuse("Con", answer("attack", "CQ:worth_finding_out_first", "AS:finding_out_first", bp, vb()), "S2")
	})
	t.Run("A4.2b", func(t *testing.T) {
		// Added after the first run (G-unbound): a blank I names no
		// investigation, the bare claim of knowability D053 refuses.
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		for _, i := range []string{"", "?", string(rune(0x200b))} {
			g.refuse("Con", answer("attack", "CQ:worth_finding_out_first", "AS:finding_out_first", bp, vb("I", i)), "S2")
		}
		pr := g.pr()
		// A blank action is no declared action: P1 refuses it first.
		for _, a2 := range []string{" ", "?"} {
			g.refuse("Con", betterMeans(pr, vb("A2", a2), "sooner"), "P1")
		}
		// A declaration whose id names nothing is refused (the check of step
		// 7, N-6).
		decl := betterMeans(pr, vb("A2", string(rune(0x200b))), "sooner")
		decl.Declare = []nascq.ActionDecl{{ID: string(rune(0x200b)), Actor: DefaultAgent}}
		g.refuse("Con", decl, "P1")
		// Where an action's name is written as the atom a blank value makes
		// (x_empty), the blank value still names no action (the check of
		// step 7, N-6).
		g2 := newFGame(t)
		g2.gs.Profile.Actions = append(g2.gs.Profile.Actions, ProfileAction{ID: "x empty", Actor: DefaultAgent})
		pr2 := g2.pr()
		g2.refuse("Con", betterMeans(pr2, vb("A2", " "), "sooner"), "P1")
	})
	t.Run("A4.3", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		// The answer form is refused by the question's own check, before the
		// move is placed (the game's check after placement is a second layer).
		g.refuse("Con", answer("attack", "CQ:worth_finding_out_first", nascq.OtherSchemeID, bp, vb("P1", "a study could settle it")),
			`S5: CQ "CQ:worth_finding_out_first" refuses Other as an answer`)
		g.refuse("Con", about("support", nascq.OtherSchemeID, worthDF, vb("P1", "a study could settle it")), "S5")
		g.refuse("Con", support(nascq.OtherSchemeID, vb("P1", "a study could settle it", "C", worthDF)), "S5")
		g.refuse("Con", support(nascq.OtherSchemeID, vb("P1", worthDF, "C", "the tax can wait")), "S5")
		g.play("Pro", about("attack", nascq.OtherSchemeID, worthDF, vb("P1", "the decision cannot wait")))
		g.refuse("Con", about("attack", nascq.OtherSchemeID, "¬"+worthDF, vb("P1", "a study could settle it")), "S5")
	})
	t.Run("A4.4", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		exp := g.play("Con", answer("attack", "CQ:worth_finding_out_first", "AS:expert_opinion", bp, vb("W", "dr_lee", "D", "food_economics")))
		g.mustIn(worthDF)
		g.outOfForce(bp)
		g.play("Pro", ask("CQ:expertise", exp))
		g.outOfForce(exp)
		g.mustOut(worthDF)
		g.inForce(bp)
	})
	t.Run("A4.5", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		g.mPlus("prices_rise")
		b2 := g.bPlus("prices_rise")
		g.play("Con", answer("attack", "CQ:worth_finding_out_first", "AS:finding_out_first", bp, vb("I", "pilot_study")))
		g.outOfForce(bp)
		g.inForce(b2)
	})
	t.Run("A4.6", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		g.mMinus()
		bp := g.bPlus("demand_falls")
		bm := g.bMinus()
		g.c1()
		ans := g.play("Con", answer("attack", "CQ:worth_finding_out_first", "AS:finding_out_first", bp, vb("I", "pilot_study")))
		q := ask(nascq.CQPremiseAcceptable, ans)
		q.TargetPremiseID = "settles_whether(pilot_study,tax_meat,demand_falls)"
		g.play("Pro", q)
		g.outOfForce(ans)
		g.mustOut(worthDF)
		g.inForce(bp, bm)
	})
	t.Run("A4.7", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		g.refuse("Con", ask("CQ:worth_finding_out_first", bp), "Q1")
	})
	t.Run("A4.8", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		g.refuse("Con", answer("attack", "CQ:worth_finding_out_first", "AS:sign", bp, vb("A", "a_pilot_is_planned")), "S5")
		g.refuse("Con", answer("attack", "CQ:worth_finding_out_first", "AS:analogy", bp, vb("C", "the_sugar_tax")), "S5")
	})
	t.Run("A4.9", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		q := ask(nascq.CQPremiseAcceptable, bp)
		q.TargetPremiseID = "matters_to_decision(tax_meat,demand_falls)"
		g.play("Con", q)
		g.outOfForce(bp)
		by := g.play("Pro", about("support", nascq.OtherSchemeID, "matters_to_decision(tax_meat,demand_falls)",
			vb("P1", "whether demand falls decides whether the tax works")))
		g.inForce(bp)
		g.answeredBy(bp, by, nascq.CQPremiseAcceptable)
	})
	t.Run("A4.10", func(t *testing.T) {
		// The board needs the statement (first run: S7 with m+ alone).
		g := newFGame(t)
		g.mPlus("demand_falls")
		g.bPlus("demand_falls")
		g.refuse("Pro", about("support", "AS:consequences_modal", shouldTax, vb("A", "tax_meat", "G", "demand_falls")), "S9")
	})
	t.Run("A4.11", func(t *testing.T) {
		g := newFGame(t)
		g.bPlus("demand_falls")
		g.mustOut("may_be_worth_performing(tax_meat,demand_falls)")
		g.mustNotIn(shouldTax)
		g.mPlus("demand_falls")
		g.mustIn(shouldTax)
	})

	t.Run("A5.1", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		ans := g.play("Con", answer("attack", "CQ:side_effect", "AS:sign", c1, vb("A", "food_banks_report_strain")))
		if got := g.conclusion(ans); got != "¬"+shouldTax {
			t.Fatalf("the answer concludes %s", got)
		}
	})
	t.Run("A5.2", func(t *testing.T) {
		g := newFGame(t)
		pr := g.pr()
		g.play("Con", answer("attack", "CQ:possibility", "AS:sign", pr, vb("A", "the_budget_is_frozen")))
		g.mustOut("possible_bring_about(decider,tax_meat)")
		g.outOfForce(pr)
	})
	t.Run("A5.3", func(t *testing.T) {
		g := newFGame(t)
		pg := g.pg()
		g.play("Con", answer("attack", "CQ:goal_possible", "AS:sign", pg, vb("A", "emissions_rose_last_year")))
		g.mustOut("possible_goal(emissions_fall)")
		g.outOfForce(pg)
	})
	t.Run("A5.4", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.refuse("Con", answer("attack", "CQ:likelihood", "AS:sign", c1, vb("A", "the_market_is_flat")), "S5")
	})

	t.Run("A6.1", func(t *testing.T) {
		g := newFGame(t)
		n1 := g.n1()
		ans := g.play("Pro", answer("attack", "CQ:opposite_consequences", "AS:consequences", n1, vb("A", "tax_meat", "G", "emissions_fall")))
		if got := g.conclusion(ans); got != shouldTax {
			t.Fatalf("the answer concludes %s", got)
		}
	})
	t.Run("A6.2", func(t *testing.T) {
		g := newFGame(t)
		n1 := g.n1()
		g.refuse("Pro", answer("attack", "CQ:opposite_consequences", "AS:negative_consequences", n1, vb("A", "tax_meat", "G", "jobs_lost")), "S5")
	})
}

// newFGameNoProfile is item F's board without a profile, as a game made
// before step C5.2: no rule of the profile (P1 to P5) applies, and a move
// binds practical reasoning's P itself.
func newFGameNoProfile(t *testing.T) *fgame {
	t.Helper()
	g := newFGame(t)
	g.gs.Config.Profile, g.gs.Profile = nil, nil
	return g
}

// Step C5.3 (D099) (docs/c5-item-f.md, A7 to A9): the backup edge from
// negative consequences to CQ:side_effect, P5, the offers, D095 (3)'s list
// bounding every route to the bridges' exception, and the question's text.
// A8 and A9 but A9.4 ran first on the data before the switch; A7 and A9.4
// came with the switch, which changed the data.
func TestItemFStepC53(t *testing.T) {
	strain := vb("A", "food_banks_report_strain")
	worthRefused := "S5: " + worthDF + " is the exception"
	costlier := vb("A", "tax_meat", "G", "food_costlier")
	t.Run("A7.1", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		ans := g.play("Con", answer("attack", "CQ:side_effect", "AS:negative_consequences", c1, costlier))
		if got := g.conclusion(ans); got != "¬"+shouldTax {
			t.Fatalf("the answer concludes %s", got)
		}
		// A rebuttal: c1 stays in force, and the weighing, c1's exception,
		// is not established by it (the check of the plan, Q2).
		g.inForce(c1)
		g.mustNotIn(bad1)
	})
	t.Run("A7.2", func(t *testing.T) {
		g := newFGame(t)
		pr := g.pr()
		ans := g.play("Con", answer("attack", "CQ:side_effect", "AS:negative_consequences", pr, costlier))
		if got := g.conclusion(ans); got != "¬"+shouldTax {
			t.Fatalf("the answer concludes %s", got)
		}
		g.inForce(pr)
	})
	t.Run("A7.3", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.refuse("Con", answer("attack", "CQ:side_effect", "AS:negative_consequences", c1, vb("A", "subsidy_plants", "G", "food_costlier")), "S9")
	})
	t.Run("A7.4", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.refuse("Con", answer("attack", "CQ:side_effect", "AS:consequences", c1, costlier), "S5")
	})
	t.Run("A7.5", func(t *testing.T) {
		g := newFGame(t)
		n1 := g.n1()
		g.refuse("Pro", answer("attack", "CQ:opposite_consequences", "AS:negative_consequences", n1, vb("A", "tax_meat", "G", "jobs_lost")), "S5")
	})
	t.Run("A9.4", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		text := g.gs.QuestionText("CQ:worth_finding_out_first", bp)
		if !strings.HasSuffix(text, "If so, how?") || strings.Contains(text, "Which way?") {
			t.Fatalf("the question reads %q", text)
		}
	})
	t.Run("A8.1", func(t *testing.T) {
		g := newFGame(t)
		pr := g.pr()
		g.refuse("Con", answer("attack", "CQ:side_effect", "AS:sign", pr, strain), "P5")
	})
	t.Run("A8.1b", func(t *testing.T) {
		g := newFGame(t)
		pr := g.pr()
		g.refuse("Pro", answer("support", "CQ:side_effect", "AS:sign", pr, vb("A", "emissions_fell_where_meat_was_taxed")), "P5")
	})
	t.Run("A8.2", func(t *testing.T) {
		g := newFGame(t)
		pg := g.pg()
		g.refuse("Con", answer("attack", "CQ:side_effect", "AS:sign", pg, strain), "P5")
	})
	t.Run("A8.3", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		ans := g.play("Con", answer("attack", "CQ:side_effect", "AS:sign", c1, strain))
		if got := g.conclusion(ans); got != "¬"+shouldTax {
			t.Fatalf("the answer concludes %s", got)
		}
	})
	t.Run("A8.4", func(t *testing.T) {
		g := newFGameNoProfile(t)
		pr := g.play("Pro", support("AS:practical_reasoning", vb("A", "tax_meat", "S1", "high_emissions", "S2", "low_emissions",
			"G", "emissions_fall", "V", "sustainability", "P", DefaultAgent)))
		ans := g.play("Con", answer("attack", "CQ:side_effect", "AS:sign", pr, strain))
		if got := g.conclusion(ans); got != "¬"+shouldTax {
			t.Fatalf("the answer concludes %s", got)
		}
	})
	t.Run("A8.5", func(t *testing.T) {
		// The answers made with sign to CQ:side_effect that the game offers
		// about the argument, by polarity.
		signAnswers := func(g *fgame, arg string) map[string]bool {
			m := map[string]bool{}
			for _, a := range g.gs.EnumerateActions([]string{arg}) {
				if a.TargetArgID == arg && a.AnswersCQID == "CQ:side_effect" && a.SchemeID == "AS:sign" {
					m[a.Type] = true
				}
			}
			return m
		}
		g := newFGame(t)
		c1, pr := g.c1(), g.pr()
		if m := signAnswers(g, pr); len(m) != 0 {
			t.Errorf("the answers P5 refuses are offered at pr: %v", m)
		}
		if m := signAnswers(g, c1); !m["attack"] || !m["support"] {
			t.Errorf("A5.1's answers are not offered at c1: %v", m)
		}
		for _, a := range g.gs.BuildObservation().AvailableActions {
			if a.TargetArgID == pr && a.AnswersCQID == "CQ:side_effect" && a.SchemeID == "AS:sign" {
				t.Errorf("offered to the computer players: %+v", a)
			}
		}
		h := newFGameNoProfile(t)
		prh := h.play("Pro", support("AS:practical_reasoning", vb("A", "tax_meat", "S1", "high_emissions", "S2", "low_emissions",
			"G", "emissions_fall", "V", "sustainability", "P", DefaultAgent)))
		if m := signAnswers(h, prh); !m["attack"] || !m["support"] {
			t.Errorf("without a profile, sign's answers at pr are not offered: %v", m)
		}
	})

	t.Run("A8.6", func(t *testing.T) {
		g := newFGame(t)
		pr := g.pr()
		r := g.play("Con", nascq.ProposedAction{Type: "attack", AttackType: "rebuttal", SchemeID: "AS:sign", TargetArgID: pr, VariableBindings: strain})
		if got := g.conclusion(r); got != "¬"+shouldTax {
			t.Fatalf("the rebuttal concludes %s", got)
		}
	})
	t.Run("A8.7", func(t *testing.T) {
		g := newFGame(t)
		g.pr()
		r := g.play("Con", about("attack", "AS:sign", shouldTax, strain))
		if got := g.conclusion(r); got != "¬"+shouldTax {
			t.Fatalf("the argument concludes %s", got)
		}
	})

	t.Run("A9.1", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		g.bPlus("demand_falls")
		g.refuse("Con", about("support", "AS:sign", worthDF, vb("A", "a_pilot_is_planned")), worthRefused)
		g.refuse("Con", about("support", "AS:analogy", worthDF, vb("C", "the_sugar_tax")), worthRefused)
	})
	t.Run("A9.1b", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		g.bPlus("demand_falls")
		g.play("Pro", about("attack", nascq.OtherSchemeID, worthDF, vb("P1", "the decision cannot wait")))
		g.refuse("Con", about("attack", "AS:sign", "¬"+worthDF, vb("A", "a_pilot_is_planned")), worthRefused)
	})
	t.Run("A9.2", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		src := g.play("Con", about("support", "AS:position_to_know", worthDF, vb("W", "the_farm_ministry", "D", "farm_markets")))
		if got := g.conclusion(src); got != worthDF {
			t.Fatalf("the source's argument concludes %s", got)
		}
		g.mustIn(worthDF)
		g.outOfForce(bp)
	})
	t.Run("A9.2b", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		src := g.play("Con", about("support", "AS:expert_opinion", worthDF, vb("W", "dr_lee", "D", "food_economics")))
		if got := g.conclusion(src); got != worthDF {
			t.Fatalf("the expert's argument concludes %s", got)
		}
		g.mustIn(worthDF)
		g.outOfForce(bp)
	})
	// An undercutter of b+ answering no question, made with the scheme given.
	undercut := func(bp, scheme, ground string) nascq.ProposedAction {
		b := map[string]string{
			nascq.OtherSchemeID: "P1",
			"AS:sign":           "A",
			"AS:analogy":        "C",
		}[scheme]
		return nascq.ProposedAction{Type: "attack", AttackType: "undercutter", SchemeID: scheme, TargetArgID: bp, Ground: ground,
			VariableBindings: vb(b, "the_vote_comes_first")}
	}
	t.Run("A9.3", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		g.play("Con", undercut(bp, nascq.OtherSchemeID, nascq.GroundOther))
		g.outOfForce(bp)
		g.mustNotIn(worthDF)
		if m := g.gs.MoveHistory[len(g.gs.MoveHistory)-1]; m.Action.Ground != nascq.GroundOther {
			t.Fatalf("the move's record keeps the ground %q", m.Action.Ground)
		}
	})
	t.Run("A9.3b", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		for _, ground := range []string{"", nascq.GroundKnowability} {
			g.refuse("Con", undercut(bp, nascq.OtherSchemeID, ground), "P6")
			if !strings.Contains(g.gs.LastRejection, `CQ "CQ:worth_finding_out_first"`) {
				t.Errorf("ground %q: the refusal does not name the question: %s", ground, g.gs.LastRejection)
			}
		}
	})
	t.Run("A9.3c", func(t *testing.T) {
		for _, scheme := range []string{"AS:sign", "AS:analogy"} {
			g := newFGame(t)
			g.mPlus("demand_falls")
			bp := g.bPlus("demand_falls")
			g.refuse("Con", undercut(bp, scheme, ""), "P6")
			g.play("Con", undercut(bp, scheme, nascq.GroundOther))
			g.outOfForce(bp)
		}
	})
	t.Run("A9.3d", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		g.play("Con", undercut(bp, nascq.OtherSchemeID, nascq.GroundOther))
		u := g.gs.Adapter.Graph.Arguments[bp].Undercutter
		if u == nil {
			t.Fatal("b+ has no undercut statement")
		}
		g.refuse("Con", about("support", "AS:sign", u.Id, vb("A", "the_vote_comes_first")), "P6")
	})
	t.Run("A9.3e", func(t *testing.T) {
		g := newFGameNoProfile(t)
		g.play("Pro", support("AS:consequences_modal", vb("A", "tax_meat", "G", "demand_falls")))
		bp := g.play("Pro", support("AS:bridge_possible_good", vb("A", "tax_meat", "G", "demand_falls")))
		g.play("Con", undercut(bp, nascq.OtherSchemeID, ""))
		g.outOfForce(bp)
	})
	t.Run("A9.3f", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.play("Con", undercut(c1, nascq.OtherSchemeID, ""))
		g.outOfForce(c1)
	})
	t.Run("A9.5", func(t *testing.T) {
		// The schemes offered for and against the exception, sorted.
		check := func(what string, ts []nascq.ActionTemplate) {
			t.Helper()
			offered := map[string][]string{}
			for _, a := range ts {
				if a.TargetStmtID == worthDF {
					offered[a.Type] = append(offered[a.Type], a.SchemeID)
				}
			}
			sort.Strings(offered["support"])
			sort.Strings(offered["attack"])
			if got, want := strings.Join(offered["support"], " "), "AS:expert_opinion AS:position_to_know"; got != want {
				t.Errorf("%s, for the exception: %q, want %q", what, got, want)
			}
			if got, want := strings.Join(offered["attack"], " "), "AS:analogy AS:expert_opinion AS:other AS:position_to_know AS:sign"; got != want {
				t.Errorf("%s, against the exception: %q, want %q", what, got, want)
			}
		}
		g := newFGame(t)
		g.mPlus("demand_falls")
		g.bPlus("demand_falls")
		// The moves about the statement, as the board lists them when it is
		// chosen (the server's statement moves).
		check("the statement's moves", g.gs.NASCQ.EnumerateStatementActions(g.gs.Bridge(), []string{worthDF}))
		// The computer players' menu lists a statement an argument concludes.
		g.play("Con", about("support", "AS:position_to_know", worthDF, vb("W", "the_farm_ministry", "D", "farm_markets")))
		check("the computer players' menu", g.gs.BuildObservation().AvailableActions)
	})
	t.Run("A9.7", func(t *testing.T) {
		g := newFGame(t)
		var ids []string
		for _, s := range g.gs.NASCQ.GetBackupSchemes("CQ:worth_finding_out_first") {
			if !s.Free {
				ids = append(ids, s.ID)
			}
		}
		sort.Strings(ids)
		if got, want := strings.Join(ids, " "), "AS:expert_opinion AS:finding_out_first AS:position_to_know"; got != want {
			t.Fatalf("D095 (3)'s list is %q, want %q", got, want)
		}
	})
	t.Run("A9.6", func(t *testing.T) {
		g := newFGame(t)
		g.mPlus("demand_falls")
		g.bPlus("demand_falls")
		g.gs.Adapter.AddStatement("placed_premise", "a placed premise")
		id, err := g.gs.Adapter.AddArgumentDirect("Con", "sign", []string{"placed_premise"}, worthDF, 0.5)
		if err != nil {
			t.Fatal(err)
		}
		g.gs.Adapter.SetArgumentMeta(id, carneades.RoleKey, carneades.RoleMain)
		for scheme, want := range map[string]string{
			"AS:sign":             "S5: an argument of Argument from Sign cannot conclude " + worthDF,
			nascq.OtherSchemeID:   "S5: an argument of Other cannot conclude " + worthDF,
			"AS:position_to_know": "",
		} {
			got := g.gs.refusedException([]string{id}, &nascq.ValidatedAction{ProposedAction: nascq.ProposedAction{SchemeID: scheme}})
			if (want == "") != (got == "") || !strings.HasPrefix(got, want) {
				t.Errorf("%s: %q, want %q", scheme, got, want)
			}
		}
	})
}

// B. Other and the thesis-independent schemes.
func TestItemFOther(t *testing.T) {
	other := func(p string) map[string]string { return vb("P1", p) }
	t.Run("B1", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.play("Con", answer("attack", "CQ:too_slight", nascq.OtherSchemeID, c1, other("the fall is tiny")))
		g.mustIn(slight1)
		g.outOfForce(c1)
	})
	t.Run("B2", func(t *testing.T) {
		g := newFGame(t)
		c3, n3 := g.c3(), g.n3()
		g.play("Con", answer("attack", "CQ:too_slight", nascq.OtherSchemeID, c3, other("demand hardly moves")))
		g.mustIn("too_slight(tax_meat,demand_falls)")
		g.outOfForce(c3, n3)
	})
	t.Run("B3", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		o := g.play("Con", answer("attack", "CQ:too_slight", nascq.OtherSchemeID, c1, other("the fall is tiny")))
		g.play("Pro", ask(nascq.CQRelevance, o))
		g.outOfForce(o)
		g.mustOut(slight1)
		g.inForce(c1)
	})
	t.Run("B4", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		// Refused by the adapter before the graph's cycle check (first run).
		g.refuse("Con", answer("attack", "CQ:bad_outweighs", nascq.OtherSchemeID, c1, other(bad1)), "adapter error: the premise")
	})
	t.Run("B4b", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.refuse("Con", answer("attack", "CQ:bad_outweighs", nascq.OtherSchemeID, c1, other(shouldTax)), "move creates a cycle")
	})
	t.Run("B5", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		exp := g.play("Con", answer("attack", "CQ:likelihood", "AS:expert_opinion", c1, vb("W", "dr_lee", "D", "climate_science")))
		g.mustOut(brings1)
		g.outOfForce(c1)
		g.play("Pro", ask("CQ:field", exp))
		g.outOfForce(exp)
		g.inForce(c1)
	})
	// B6 and B6b are played as recorded after the checks of steps 7 (E-1)
	// and 8 (W-1): c1 first, so that the Other argument concludes c1's
	// premise, whose presumption a pending question ties, so that it holds
	// only by the arguments for it.
	t.Run("B6", func(t *testing.T) {
		// The expert's argument is on the board when the undercut comes, so
		// that the support already there must survive it (Q2-C).
		g := newFGame(t)
		c1 := g.c1()
		g.play("Con", ask("CQ:likelihood", c1))
		g.mustOut(brings1)
		wish := g.play("Pro", support(nascq.OtherSchemeID, vb("P1", "lower emissions would be good", "C", brings1)))
		if got := g.conclusion(wish); got != brings1 {
			t.Fatalf("the Other argument concludes %s, not c1's premise", got)
		}
		exp := g.play("Pro", about("support", "AS:expert_opinion", brings1, vb("W", "dr_lee", "D", "climate_science")))
		g.mustIn(brings1)
		g.inForce(c1, exp)
		g.play("Con", answer("attack", nascq.CQRelevance, nascq.OtherSchemeID, wish,
			other("that an outcome is desirable is no evidence that it will happen")))
		g.outOfForce(wish) // Q2-A
		if u := g.gs.Adapter.Graph.Arguments[c1].Undercutter; u == nil || u.Label == caes.In {
			t.Fatalf("c1's undercutter is %+v: the undercutter reached c1", u)
		} // Q2-B: c1 itself is untouched
		g.inForce(exp) // Q2-C: the support already there stands
		g.mustIn(brings1)
		g.inForce(c1)
	})
	t.Run("B6b", func(t *testing.T) {
		// The undercut comes while the Other argument is the premise's only
		// support: the premise goes Out with it, and comes In again with
		// support added after.
		g := newFGame(t)
		c1 := g.c1()
		g.play("Con", ask("CQ:likelihood", c1))
		wish := g.play("Pro", support(nascq.OtherSchemeID, vb("P1", "lower emissions would be good", "C", brings1)))
		if got := g.conclusion(wish); got != brings1 {
			t.Fatalf("the Other argument concludes %s, not c1's premise", got)
		}
		g.mustIn(brings1)
		g.play("Con", answer("attack", nascq.CQRelevance, nascq.OtherSchemeID, wish,
			other("that an outcome is desirable is no evidence that it will happen")))
		g.outOfForce(wish)
		g.mustOut(brings1) // the premise stood only on the undercut inference
		if u := g.gs.Adapter.Graph.Arguments[c1].Undercutter; u == nil || u.Label == caes.In {
			t.Fatalf("c1's undercutter is %+v: the undercutter reached c1", u)
		}
		g.play("Pro", about("support", "AS:expert_opinion", brings1, vb("W", "dr_lee", "D", "climate_science")))
		g.mustIn(brings1)
		g.inForce(c1)
		g.outOfForce(wish)
	})
}

// C. Questions asked alone.
func TestItemFQuestionsAlone(t *testing.T) {
	t.Run("C1", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		before := map[string]caes.Label{}
		for id, s := range g.gs.Adapter.Graph.Statements {
			before[id] = s.Label
		}
		g.play("Con", ask("CQ:likelihood", c1))
		g.mustOut(brings1)
		g.outOfForce(c1)
		for id, l := range before {
			if id != brings1 && id != shouldTax && g.label(id) != l {
				t.Fatalf("%s went from %v to %v", id, l, g.label(id))
			}
		}
	})
	t.Run("C2", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.play("Con", ask("CQ:likelihood", c1))
		g.play("Con", ask("CQ:evidence", c1))
		exp := g.play("Pro", about("support", "AS:expert_opinion", brings1, vb("W", "dr_lee", "D", "climate_science")))
		g.mustIn(brings1)
		g.inForce(c1)
		g.answeredBy(c1, exp, "CQ:likelihood", "CQ:evidence")
	})
	t.Run("C3", func(t *testing.T) {
		g := newFGame(t)
		pr := g.pr()
		g.play("Con", ask("CQ:possibility", pr))
		g.mustOut("possible_bring_about(decider,tax_meat)")
		g.outOfForce(pr)
	})
	t.Run("C4", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.mPlus("demand_falls")
		bp := g.bPlus("demand_falls")
		for _, cq := range []string{"CQ:too_slight", "CQ:bad_outweighs", "CQ:gain_discounted"} {
			g.refuse("Con", ask(cq, c1), "Q1")
		}
		g.refuse("Con", ask("CQ:worth_finding_out_first", bp), "Q1")
	})
	t.Run("C5", func(t *testing.T) {
		g := newFGame(t)
		c1, n1 := g.c1(), g.n1()
		g.refuse("Con", ask("CQ:side_effect", c1), "Q1")
		g.refuse("Pro", ask("CQ:opposite_consequences", n1), "Q1")
	})
	t.Run("C6", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.play("Con", answer("attack", "CQ:too_slight", nascq.OtherSchemeID, c1, vb("P1", "the fall is tiny")))
		g.play("Pro", about("attack", nascq.OtherSchemeID, slight1, vb("P1", "the fall is large")))
		g.mustNotIn(slight1)
		g.inForce(c1)
	})
	t.Run("C7", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		g.play("Con", ask("CQ:likelihood", c1))
		g.refuse("Con", ask("CQ:likelihood", c1), "Q2")
	})
}

// D. The direct rebuttal where an answer route is removed.
func TestItemFDirectRebuttal(t *testing.T) {
	neg := vb("A", "tax_meat", "G", "food_costlier")
	t.Run("D1", func(t *testing.T) {
		// Refused by S5 before step C5.3's switch; since it, by the owner's
		// answer D099 (1) (the backup edge), accepted, a rebuttal: A7.1's
		// outcome (docs/c5-item-f.md, D).
		g := newFGame(t)
		c1 := g.c1()
		ans := g.play("Con", answer("attack", "CQ:side_effect", "AS:negative_consequences", c1, neg))
		if got := g.conclusion(ans); got != "¬"+shouldTax {
			t.Fatalf("the answer concludes %s", got)
		}
		g.inForce(c1)
		g.mustNotIn(bad1)
	})
	t.Run("D2 and D3", func(t *testing.T) {
		g := newFGame(t)
		c1 := g.c1()
		r := g.play("Con", about("attack", "AS:negative_consequences", shouldTax, neg))
		if got := g.conclusion(r); got != "¬"+shouldTax {
			t.Fatalf("the rebuttal concludes %s", got)
		}
		g.mustNotIn(bad1)
		g.inForce(c1)
	})
}

// E. The profile.
func TestItemFProfile(t *testing.T) {
	t.Run("E1", func(t *testing.T) {
		g := newFGame(t)
		g.refuse("Pro", support("AS:consequences", vb("A", "farmers_strike", "G", "emissions_fall")), "P3")
	})
	t.Run("E2", func(t *testing.T) {
		g := newFGame(t)
		g.refuse("Pro", support("AS:consequences", vb("A", "ban_meat", "G", "emissions_fall")), "P1")
	})
	t.Run("E3", func(t *testing.T) {
		g := newFGame(t)
		c := g.play("Pro", support("AS:consequences", vb("A", "meat_tax", "G", "emissions_fall")))
		if got := g.conclusion(c); got != shouldTax {
			t.Fatalf("the argument concludes %s", got)
		}
	})
}
