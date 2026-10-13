package game

import (
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

func otherGame(t *testing.T, cfg func(*GameConfig)) *GameState {
	t.Helper()
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config = DefaultGameConfig()
	gs.Config.AllowOther = true
	if cfg != nil {
		cfg(&gs.Config)
	}
	gs.NASCQ.EnableOther()
	return gs
}

func free(c string, premises ...string) nascq.ProposedAction {
	b := map[string]string{"C": c}
	for i, p := range premises {
		b["P"+string(rune('1'+i))] = p
	}
	return nascq.ProposedAction{Type: "support", SchemeID: nascq.OtherSchemeID, VariableBindings: b}
}

// The user-defined scheme: free premises and conclusion, its generic
// questions, and answers with it to the questions of an official scheme.
func TestOtherScheme(t *testing.T) {
	gs := otherGame(t, nil)
	act := free("Governments should tax meat.", "Meat production causes a large share of emissions.", "Taxes reduce consumption.")
	act.Text = "Meat should be taxed: it causes emissions, and taxes cut consumption."
	a1 := play(t, gs, act)
	arg := gs.Adapter.Graph.Arguments[a1]
	if len(arg.Premises) != 2 || arg.Conclusion.Label != caes.In || arg.Metadata["text"] != act.Text {
		t.Fatalf("free argument: %d premises, conclusion %v, text %v", len(arg.Premises), arg.Conclusion.Label, arg.Metadata["text"])
	}
	if got := gs.Say(arg.Conclusion.Id); got != "Governments should tax meat." {
		t.Fatalf("Say = %q", got)
	}

	// Bob questions the second premise (the asker names it).
	p2 := arg.Premises[1].Stmt.Id
	play(t, gs, nascq.ProposedAction{Type: "question", AnswersCQID: nascq.CQPremiseAcceptable, TargetArgID: a1, TargetPremiseID: p2})
	if arg.Conclusion.Label != caes.Out {
		t.Fatal("the questioned free argument still stands")
	}
	// Alice answers with an expert.
	def := answer("support", nascq.CQPremiseAcceptable, "AS:expert_opinion", a1, map[string]string{"W": "oecd", "D": "tax policy"})
	def.TargetPremiseID = p2
	d := play(t, gs, def)
	if !gs.bridge.PremiseExists(d, "asserts(oecd,"+p2+")") {
		t.Fatalf("the expert's premises are not about the questioned premise: %v", gs.Adapter.Graph.Arguments[d].Premises)
	}
	if !strings.Contains(gs.Say("asserts(oecd,"+p2+")"), "Oecd asserts that “Taxes reduce consumption”") {
		t.Fatalf("Say: %q", gs.Say("asserts(oecd,"+p2+")"))
	}
	if arg.Conclusion.Label != caes.In {
		t.Fatal("the answered free argument is not In")
	}
	// Bob rebuts with his own scheme.
	r, _ := gs.ExecuteTurn(nascq.ProposedAction{Type: "attack", AnswersCQID: nascq.CQCounter, TargetArgID: a1, SchemeID: nascq.OtherSchemeID,
		VariableBindings: map[string]string{"P1": "A meat tax hits poor households hardest."}})
	if !r.Accepted || gs.Adapter.Graph.Arguments[r.ArgIDs[0]].Conclusion.Id != "¬"+arg.Conclusion.Id {
		t.Fatalf("free rebuttal: %+v", r)
	}

	// An Other answer to an official scheme's question.
	gs2 := otherGame(t, nil)
	b1 := play(t, gs2, prSupportAction())
	r, _ = gs2.ExecuteTurn(nascq.ProposedAction{Type: "attack", AnswersCQID: "CQ:possibility", TargetArgID: b1, SchemeID: nascq.OtherSchemeID,
		VariableBindings: map[string]string{"P1": "Parliament has no majority for it."}})
	if !r.Accepted || gs2.Adapter.Graph.Arguments[r.ArgIDs[0]].Conclusion.Id != "¬possible(carbon_tax)" {
		t.Fatalf("free undermining: %+v", r)
	}
	if errs := gs2.CheckInvariants(); len(errs) > 0 {
		t.Fatal(errs)
	}
	// Without AllowOther, there is no such scheme.
	gs3 := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	if r, _ := gs3.ExecuteTurn(free("x", "y")); r.Accepted || !strings.HasPrefix(r.Error, "S1") {
		t.Fatalf("Other without AllowOther: %+v", r)
	}
}

// Every offered move on a free argument validates (one set per premise).
func TestOtherTemplatesAreLegal(t *testing.T) {
	gs := otherGame(t, nil)
	play(t, gs, free("C1", "first premise", "second premise"))
	n := 0
	for _, a := range gs.BuildObservation().AvailableActions {
		if a.TargetArgID == "" {
			continue
		}
		n++
		b := map[string]string{"P1": "p"}
		for v := range a.Variables {
			b[v] = "x"
		}
		if _, err := gs.NASCQ.Validate(fromTemplate(a, b), gs.bridge); err != nil {
			t.Fatalf("%+v: %v", a, err)
		}
	}
	if n == 0 {
		t.Fatal("no moves on the free argument")
	}
}

func TestRequireEvidence(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.RequireEvidence = true
	act := nascq.ProposedAction{Type: "support", SchemeID: "AS:expert_opinion",
		VariableBindings: map[string]string{"W": "ipcc", "D": "climate", "S": "warming is real"}}
	r, _ := gs.ExecuteTurn(act)
	if r.Accepted || !strings.HasPrefix(r.Error, "E1") {
		t.Fatalf("expert opinion without a source: %+v", r)
	}
	act.Evidence = []string{" ", "IPCC AR6 SPM, A.1"}
	id := play(t, gs, act)
	ev, _ := gs.Adapter.Graph.Arguments[id].Metadata["evidence"].([]string)
	if len(ev) != 1 || ev[0] != "IPCC AR6 SPM, A.1" {
		t.Fatalf("evidence %v", ev)
	}
	for _, a := range gs.BuildView().Arguments {
		if a.ID == id && len(a.Evidence) != 1 {
			t.Fatalf("view evidence %v", a.Evidence)
		}
	}
}

// A moderated game holds each move for the moderator, who approves it (with
// a weight) or rejects it, and decides on statements.
func TestModeration(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.Moderated = true
	r, err := gs.ExecuteTurn(prSupportAction())
	if err != nil || !r.Pending || gs.Pending == nil || len(gs.Adapter.Graph.Arguments) != 0 {
		t.Fatalf("held move: %+v %v, %d arguments", r, err, len(gs.Adapter.Graph.Arguments))
	}
	if _, err := gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"}); err == nil {
		t.Fatal("a move was accepted while another awaits the moderator")
	}
	r, err = gs.Moderate(true, 0.9, "sound")
	if err != nil || !r.Accepted || gs.Pending != nil || gs.ActivePlayer().ID != "Bob" {
		t.Fatalf("approved: %+v %v", r, err)
	}
	a1 := r.ArgIDs[0]
	if m := gs.MoveHistory[len(gs.MoveHistory)-1]; m.Weight != 0.9 || m.Moderation != "approved: sound" {
		t.Fatalf("approved move: weight %v, moderation %q", m.Weight, m.Moderation)
	}

	// A rejected move costs the penalty and the same player moves again.
	r, _ = gs.ExecuteTurn(answer("attack", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "dr_no", "D": "politics"}))
	if !r.Pending {
		t.Fatalf("Bob's attack not held: %+v", r)
	}
	r, _ = gs.Moderate(false, 0, "no source")
	if r.Accepted || gs.ActivePlayer().ID != "Bob" || gs.Players[1].Score != gs.Config.Scoring.InvalidActionPenalty ||
		!strings.Contains(gs.LastRejection, "no source") {
		t.Fatalf("rejected: %+v, active %s, score %v", r, gs.ActivePlayer().ID, gs.Players[1].Score)
	}

	// The moderator rejects a premise: the argument falls; accepts it: it stands.
	label := func(id string) caes.Label { return gs.Adapter.Graph.Statements[id].Label }
	if _, err := gs.Decide("possible(carbon_tax)", "rejected", "no majority"); err != nil {
		t.Fatal(err)
	}
	if label("possible(carbon_tax)") != caes.Out || label("should_be_performed(carbon_tax)") != caes.Out {
		t.Fatal("a rejected premise still holds")
	}
	if _, err := gs.Decide("possible(carbon_tax)", "accepted", ""); err != nil {
		t.Fatal(err)
	}
	if label("possible(carbon_tax)") != caes.In || label("should_be_performed(carbon_tax)") != caes.In ||
		gs.Adapter.Decision("possible(carbon_tax)") != "accepted" {
		t.Fatal("an accepted premise does not hold")
	}
	if errs := gs.CheckInvariants(); len(errs) > 0 {
		t.Fatal(errs)
	}
	if !strings.HasPrefix(gs.DescribeMove(gs.MoveHistory[len(gs.MoveHistory)-1]), "The moderator accepts: Action carbon tax is possible") {
		t.Fatalf("describe: %q", gs.DescribeMove(gs.MoveHistory[len(gs.MoveHistory)-1]))
	}
}

func TestSuggest(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	for text, want := range map[string]string{
		"According to Dr. Lee, an expert in economics, the tax works.": "AS:expert_opinion",
		"We should tax meat in order to reduce emissions.":             "AS:practical_reasoning",
		"Smoke indicates a fire nearby.":                               "AS:sign",
		"Just as in Sweden, a carbon tax would work here.":             "AS:analogy",
	} {
		s := gs.NASCQ.Suggest(text)
		if len(s) == 0 || s[0].SchemeID != want {
			t.Errorf("Suggest(%q) = %+v, want %s first", text, s, want)
		}
	}
}
