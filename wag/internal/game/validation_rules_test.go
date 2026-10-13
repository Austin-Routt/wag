package game

import (
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// templatesOn returns the enumerated attack templates against argID.
func templatesOn(gs *GameState, argID string) []nascq.ActionTemplate {
	var out []nascq.ActionTemplate
	for _, a := range gs.BuildObservation().AvailableActions {
		if a.Type == "attack" && a.TargetArgID == argID {
			out = append(out, a)
		}
	}
	return out
}

func fromTemplate(tmpl nascq.ActionTemplate, bindings map[string]string) nascq.ProposedAction {
	return nascq.ProposedAction{Type: tmpl.Type, AttackType: tmpl.AttackType, SchemeID: tmpl.SchemeID,
		AnswersCQID: tmpl.AnswersCQID, TargetArgID: tmpl.TargetArgID, TargetPremiseID: tmpl.TargetPremiseID,
		VariableBindings: bindings}
}

// "Is it actually possible to perform action A?" undermines possible(A), as
// its target_premise_pattern says, when played from the enumerated template
// (as agents and the CLI wizard play it). It used to undermine the first
// premise, current_circumstances(S1).
func TestPossibilityQuestionUnderminesPossible(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	r, _ := gs.ExecuteTurn(prSupportAction())
	main := r.ArgIDs[0]
	var tmpl *nascq.ActionTemplate
	for _, a := range templatesOn(gs, main) {
		if a.AnswersCQID == "CQ:possibility" && a.SchemeID == "AS:sign" {
			a := a
			tmpl = &a
		}
	}
	if tmpl == nil || tmpl.TargetPremiseID != "possible(carbon_tax)" {
		t.Fatalf("possibility template: %+v", tmpl)
	}
	tmpl.TargetPremiseID = "" // as a hand-written action would leave it: validation resolves it
	r2, err := gs.ExecuteTurn(fromTemplate(*tmpl, map[string]string{"A": "no_majority", "B": "cannot_pass"}))
	if err != nil || !r2.Accepted {
		t.Fatalf("undermining rejected: %v %+v", err, r2)
	}
	if got := gs.Adapter.Graph.Arguments[r2.ArgIDs[0]].Conclusion.Id; got != "¬possible(carbon_tax)" {
		t.Fatalf("the possibility question concluded %s", got)
	}
	if rec := gs.MoveHistory[len(gs.MoveHistory)-1].Action; rec.TargetPremiseID != "possible(carbon_tax)" || rec.AttackType != "undermining" {
		t.Fatalf("history records %+v, want the resolved premise and attack type", rec)
	}
	if l := gs.Adapter.Graph.Arguments[main].Conclusion.Label; l != caes.Out {
		t.Fatalf("undermined conclusion is %v, want Out", l)
	}
}

// Malformed moves fail validation (and so are penalized and recorded like any
// invalid move) instead of reaching the adapter.
func TestMalformedMovesFailValidation(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	r, _ := gs.ExecuteTurn(prSupportAction())
	main := r.ArgIDs[0]
	presumption := r.Presumptions[0]
	sign := map[string]string{"A": "a", "B": "b"}
	cases := []struct {
		name   string
		action nascq.ProposedAction
		want   string
	}{
		{"unknown type", nascq.ProposedAction{Type: "bogus", SchemeID: "AS:sign", VariableBindings: sign}, "S0"},
		{"attack without target", nascq.ProposedAction{Type: "attack", AnswersCQID: "CQ:likelihood", SchemeID: "AS:sign", VariableBindings: sign}, "S0"},
		{"attack without CQ or type", nascq.ProposedAction{Type: "attack", TargetArgID: main, SchemeID: "AS:sign", VariableBindings: sign}, "S0"},
		{"support with a CQ", nascq.ProposedAction{Type: "support", AnswersCQID: "CQ:possibility", TargetArgID: main, SchemeID: "AS:sign", VariableBindings: sign}, "S0"},
		{"CQ against a presumption", nascq.ProposedAction{Type: "attack", AnswersCQID: "CQ:side_effect", TargetArgID: presumption, SchemeID: "AS:sign", VariableBindings: sign}, "S3"},
		{"wrong attack type for the CQ", nascq.ProposedAction{Type: "attack", AttackType: "rebuttal", AnswersCQID: "CQ:side_effect", TargetArgID: main, SchemeID: "AS:sign", VariableBindings: sign}, "S3"},
		{"undermining without a premise", nascq.ProposedAction{Type: "attack", AttackType: "undermining", TargetArgID: main, SchemeID: "AS:sign", VariableBindings: sign}, "S8"},
	}
	for _, c := range cases {
		before := gs.ActivePlayer().Score
		res, err := gs.ExecuteTurn(c.action)
		if err != nil {
			t.Fatal(err)
		}
		if res.Accepted || !strings.HasPrefix(res.Error, c.want) {
			t.Errorf("%s: accepted=%v error=%q, want a %s rejection", c.name, res.Accepted, res.Error, c.want)
		}
		if gs.ActivePlayer().Score != before+gs.Config.Scoring.InvalidActionPenalty {
			t.Errorf("%s: not penalized", c.name)
		}
	}
	if errs := gs.CheckInvariants(); len(errs) > 0 {
		t.Fatal(errs)
	}
}

func TestEnumeratedUnderminingTemplatesAreValid(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.ExecuteTurn(prSupportAction())
	for _, a := range gs.BuildObservation().AvailableActions {
		if a.CaesAttackType != "undermining" {
			continue
		}
		b := map[string]string{}
		for v := range a.Variables {
			b[v] = "x"
		}
		if _, err := gs.NASCQ.Validate(fromTemplate(a, b), gs.bridge); err != nil {
			t.Errorf("enumerated template %+v fails validation: %v", a, err)
		}
	}
}

// A round in which every player abstained ends the game at the threshold even
// when a rejected move came before an abstention (the round check used to read
// the last len(Players) history records, which include rejected moves).
func TestAllAbstainRoundCountsDespiteRejectedMoves(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.AbstentionThreshold = 1
	gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})                             // Alice
	gs.ExecuteTurn(nascq.ProposedAction{Type: "support", SchemeID: "AS:nonexistent"}) // Bob, rejected
	r, _ := gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})                     // Bob
	if !r.Terminated || gs.ConsecAllAbstain != 1 {
		t.Fatalf("all abstained a round: terminated=%v ConsecAllAbstain=%d", r.Terminated, gs.ConsecAllAbstain)
	}

	// A round with a move resets the count.
	gs = newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.AbstentionThreshold = 2
	gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	gs.ExecuteTurn(prSupportAction())
	gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	if gs.ConsecAllAbstain != 0 || gs.Terminated {
		t.Fatalf("after a round with a move: ConsecAllAbstain=%d terminated=%v", gs.ConsecAllAbstain, gs.Terminated)
	}
}
