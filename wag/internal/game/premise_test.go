package game

import (
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// With AnyPremise, the premises of practical reasoning that none of its five
// critical questions is about can be questioned and undermined through the
// generic premise question; the one Possibility is about cannot.
func TestAnyPremise(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config = DefaultGameConfig()
	gs.Config.AnyPremise = true
	gs.NASCQ.EnablePremiseQuestion()
	a1 := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{"A": "ask her out", "S1": "single", "S2": "dating", "G": "a relationship", "V": "happiness"}})

	about := map[string]bool{}
	for _, tmpl := range gs.NASCQ.EnumerateActions(gs.Bridge(), []string{a1}) {
		if tmpl.AnswersCQID == nascq.CQPremiseAcceptable {
			about[tmpl.LocusStmtID] = true
		}
	}
	if len(about) != 6 || about["possible(ask_her_out)"] || !about["current_circumstances(single)"] {
		t.Fatalf("the generic question is about %v", about)
	}

	// Asking it about the premise Possibility covers is refused.
	r, err := gs.ExecuteTurn(nascq.ProposedAction{Type: "question", AnswersCQID: nascq.CQPremiseAcceptable, TargetArgID: a1, TargetPremiseID: "possible(ask_her_out)"})
	if err != nil || r.Accepted || !strings.Contains(r.Error, "ask that one") {
		t.Fatalf("question about a covered premise: %+v %v", r, err)
	}
	// Undermining the circumstances with a witness knocks a1 out.
	att := answer("attack", nascq.CQPremiseAcceptable, "AS:position_to_know", a1, map[string]string{"W": "her friend", "D": "her life"})
	att.TargetPremiseID = "current_circumstances(single)"
	play(t, gs, att)
	if c := gs.Adapter.Graph.Statements["should_be_performed(ask_her_out)"]; c.Label != caes.Out {
		t.Fatalf("the claim is %v after its circumstances were undermined", c.Label)
	}
	if got := gs.Say("¬current_circumstances(single)"); got != "The current situation is not ‘single’." {
		t.Fatalf("negation reads %q", got)
	}
}

// Statements read with the players' words quoted, nested statements in
// double quotes and, a level deeper, shortened in brackets; on a card the
// argument's own conclusion reads C.
func TestEnglish(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config = DefaultGameConfig()
	a1 := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:expert_opinion",
		VariableBindings: map[string]string{"W": "the IPCC", "D": "climate science", "S": "it rains more"}})
	_ = a1
	for id, want := range map[string]string{
		"credible_source(the_ipcc,climate_science)":          "‘The IPCC’ is a credible expert on ‘climate science’.",
		"¬in_domain(side_effects(a_tax,the_status_quo),law)": "“Doing ‘a tax’ in the situation ‘the status quo’ would have bad side effects” is not a matter of ‘law’.",
		"asserts(w,not(in_domain(side_effects(a,b),d)))":     "‘W’ says that “[doing ‘a’ in the situation ‘b’ would have …] is not a matter of ‘d’”.",
	} {
		if got := gs.Say(id); got != want {
			t.Errorf("Say(%s)\n got %q\nwant %q", id, got, want)
		}
	}
	arg := gs.Adapter.Graph.Arguments[a1]
	var texts []string
	for _, p := range arg.Premises {
		texts = append(texts, gs.SayIn(a1, p.Stmt.Id))
	}
	if !strings.Contains(strings.Join(texts, " | "), "‘The IPCC’ says that C.") {
		t.Fatalf("premises on the card: %v", texts)
	}
}

// A negated conclusion is C on its card even where a premise nests it as
// not(...).
func TestSayInNegatedConclusion(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config = DefaultGameConfig()
	a1 := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{"A": "ask her out", "S1": "single", "S2": "dating", "G": "a relationship", "V": "happiness"}})
	a2 := play(t, gs, answer("attack", "CQ:possibility", "AS:position_to_know", a1, map[string]string{"W": "her friend", "D": "her plans"}))
	var texts []string
	for _, p := range gs.Adapter.Graph.Arguments[a2].Premises {
		texts = append(texts, gs.SayIn(a2, p.Stmt.Id))
	}
	if got := strings.Join(texts, " | "); !strings.Contains(got, "‘Her friend’ says that C.") || !strings.Contains(got, "C is a matter of ‘her plans’.") {
		t.Fatalf("premises: %s", got)
	}
}

// Every context preset loads and weighs only schemes that exist.
func TestContexts(t *testing.T) {
	cs, err := LoadContexts(data.ContextsJSON)
	if err != nil || len(cs) == 0 {
		t.Fatalf("contexts: %v %v", cs, err)
	}
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.NASCQ.EnableOther()
	for _, c := range cs {
		for id := range c.Weights {
			if id != "*" && gs.NASCQ.GetScheme(id) == nil {
				t.Errorf("context %s weighs unknown scheme %s", c.ID, id)
			}
		}
		switch c.Standard {
		case "PE", "CCE", "BRD":
		default:
			t.Errorf("context %s: proof standard %q", c.ID, c.Standard)
		}
	}
}

// The same dispute under other weights: ClaimUnder re-weighs the arguments
// made with schemes, without changing the game.
func TestClaimUnder(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.MaxTurns = 0
	gs.Config.Objective = ObjectiveClaim
	play(t, gs, prSupportAction()) // practical reasoning, 0.70
	// A rebuttal-like argument against the claim from an expert (0.70): a tie
	// under PE leaves the claim out; weigh the expert lower and it is in.
	play(t, gs, nascq.ProposedAction{Type: "attack", TargetStmtID: "should_be_performed(carbon_tax)", SchemeID: "AS:expert_opinion",
		VariableBindings: map[string]string{"W": "an economist", "D": "taxation"}})
	before := gs.Adapter.StatementLabels()
	cases := map[string]map[string]float64{
		"out": {"AS:expert_opinion": 0.7},
		"in":  {"AS:expert_opinion": 0.4},
	}
	for want, w := range cases {
		if got := gs.ClaimUnder(w); got != want {
			t.Errorf("under %v the claim is %s, want %s", w, got, want)
		}
	}
	if got := gs.ClaimUnder(map[string]float64{"*": 0.5, "AS:practical_reasoning": 0.9}); got != "in" {
		t.Errorf("practical reasoning weighed 0.9 against 0.5: %s", got)
	}
	for id, l := range gs.Adapter.StatementLabels() {
		if before[id] != l {
			t.Fatalf("ClaimUnder changed %s", id)
		}
	}
	if u := gs.UnknownSchemes(map[string]float64{"*": 1, "AS:sign": 1, "AS:nope": 1}); len(u) != 1 || u[0] != "AS:nope" {
		t.Fatalf("unknown schemes: %v", u)
	}
}
