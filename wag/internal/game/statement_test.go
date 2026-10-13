package game

import (
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// Any statement on the board can be argued for or against, as the claim can,
// with a scheme that can conclude it (S9).
func TestArgueAboutStatements(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.MaxTurns = 0
	label := func(id string) caes.Label { return gs.Adapter.Graph.Statements[id].Label }
	const shouldTax, possibleTax = "should_be_performed(carbon_tax)", "possible(carbon_tax)"
	try := func(a nascq.ProposedAction) *TurnResult {
		t.Helper()
		r, err := gs.ExecuteTurn(a)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	a1 := play(t, gs, prSupportAction()) // Alice

	// Bob argues against a premise, as against a claim: a witness says the
	// tax is not possible. The premise, and with it a1, is out.
	against := nascq.ProposedAction{Type: "attack", TargetStmtID: possibleTax, SchemeID: "AS:position_to_know",
		VariableBindings: map[string]string{"W": "the finance minister", "D": "the budget"}}
	a2 := play(t, gs, against)
	if label(possibleTax) != caes.Out || label(shouldTax) != caes.Out {
		t.Fatalf("after the argument against %s: it is %v, the claim %v", possibleTax, label(possibleTax), label(shouldTax))
	}
	if c := gs.Adapter.Graph.Arguments[a2].Conclusion.Id; c != "¬"+possibleTax {
		t.Fatalf("the argument against concludes %s", c)
	}
	if m := gs.MoveHistory[len(gs.MoveHistory)-1]; m.Action.AttackType != "undermining" || m.Action.VariableBindings["S"] != "¬"+possibleTax {
		t.Fatalf("recorded as %+v", m.Action)
	}

	// Practical reasoning concludes what should be done, not what is
	// possible: it cannot argue for the premise.
	r := try(nascq.ProposedAction{Type: "support", TargetStmtID: possibleTax, SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{"A": "x", "S1": "y", "S2": "z", "G": "g", "V": "v"}})
	if r.Accepted || !strings.Contains(r.Error, "S9") {
		t.Fatalf("practical reasoning for %s: %+v", possibleTax, r)
	}
	// Nor can it argue against what should be done: no scheme of the core
	// set concludes ¬should_be_performed(A).
	r = try(nascq.ProposedAction{Type: "attack", TargetStmtID: shouldTax, SchemeID: "AS:consequences",
		VariableBindings: map[string]string{"G": "higher prices"}})
	if r.Accepted || !strings.Contains(r.Error, "S9") {
		t.Fatalf("consequences against %s: %+v", shouldTax, r)
	}
	// An argument from consequences for the claim: the statement fixes A,
	// which Bob need not give, and may not give otherwise.
	r = try(nascq.ProposedAction{Type: "support", TargetStmtID: shouldTax, SchemeID: "AS:consequences",
		VariableBindings: map[string]string{"A": "a fuel tax", "G": "cleaner air"}})
	if r.Accepted || !strings.Contains(r.Error, "S9") || !strings.Contains(r.Error, "leave A out") {
		t.Fatalf("consequences with another A: %+v", r)
	}
	a3 := play(t, gs, nascq.ProposedAction{Type: "support", TargetStmtID: shouldTax, SchemeID: "AS:consequences",
		VariableBindings: map[string]string{"G": "cleaner air"}})
	arg := gs.Adapter.Graph.Arguments[a3]
	if arg.Conclusion.Id != shouldTax || arg.Premises[0].Stmt.Id != "brings_about(carbon_tax,cleaner_air)" {
		t.Fatalf("a3 concludes %s from %s", arg.Conclusion.Id, arg.Premises[0].Stmt.Id)
	}

	// A statement that is not on the board, and a move naming both a
	// statement and an argument, are refused.
	r = try(nascq.ProposedAction{Type: "support", TargetStmtID: "possible(world_peace)", SchemeID: "AS:position_to_know",
		VariableBindings: map[string]string{"W": "w", "D": "d"}})
	if r.Accepted || !strings.Contains(r.Error, "S7") {
		t.Fatalf("a statement not on the board: %+v", r)
	}
	r = try(nascq.ProposedAction{Type: "support", TargetStmtID: possibleTax, TargetArgID: a1, SchemeID: "AS:position_to_know",
		VariableBindings: map[string]string{"W": "w", "D": "d"}})
	if r.Accepted || !strings.Contains(r.Error, "S0") {
		t.Fatalf("a statement and an argument: %+v", r)
	}

	// The direct moves on an argument are checked the same way: a rebuttal
	// with practical reasoning would conclude ¬should_be_performed(...).
	r = try(nascq.ProposedAction{Type: "attack", AttackType: "rebuttal", TargetArgID: a1, SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{"A": "x", "S1": "y", "S2": "z", "G": "g", "V": "v"}})
	if r.Accepted || !strings.Contains(r.Error, "S9") {
		t.Fatalf("a direct rebuttal with practical reasoning: %+v", r)
	}

	// Bob passes; Alice argues for the undermined premise with Other: the
	// issue is weighed again.
	gs.NASCQ.EnableOther()
	try(nascq.ProposedAction{Type: "abstain"})
	a4 := play(t, gs, nascq.ProposedAction{Type: "support", TargetStmtID: possibleTax, SchemeID: nascq.OtherSchemeID,
		VariableBindings: map[string]string{"P1": "Parliament passed a similar tax in 2008"}})
	if c := gs.Adapter.Graph.Arguments[a4].Conclusion.Id; c != possibleTax {
		t.Fatalf("the Other argument for %s concludes %s", possibleTax, c)
	}
	if errs := gs.CheckInvariants(); len(errs) > 0 {
		t.Fatal(errs)
	}
}

// A challenge ("why should we accept that?") is answered by any argument
// for the premise, not only by a defence through the question.
func TestChallengeAnsweredByAnArgumentForThePremise(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.MaxTurns, gs.Config.AnyPremise = 0, true
	gs.NASCQ.EnablePremiseQuestion()
	a1 := play(t, gs, prSupportAction())
	const circ = "current_circumstances(current_state)"
	play(t, gs, nascq.ProposedAction{Type: "question", AnswersCQID: nascq.CQPremiseAcceptable, TargetArgID: a1, TargetPremiseID: circ})
	if l := gs.Adapter.Graph.Statements[circ].Label; l != caes.Out {
		t.Fatalf("the challenged premise is %v", l)
	}
	if len(gs.Questions) != 1 || !gs.Questions[0].Open() || gs.Questions[0].Owner != "Alice" {
		t.Fatalf("open questions: %+v", gs.Questions)
	}
	a3 := play(t, gs, nascq.ProposedAction{Type: "support", TargetStmtID: circ, SchemeID: "AS:expert_opinion",
		VariableBindings: map[string]string{"W": "the statistics office", "D": "the economy"}})
	if q := gs.Questions[0]; q.AnsweredBy != a3 {
		t.Fatalf("the challenge is answered by %q, want %s", q.AnsweredBy, a3)
	}
	if l := gs.Adapter.Graph.Statements[circ].Label; l != caes.In {
		t.Fatalf("the premise argued for is %v", l)
	}
}

// An argument that answers a critical question concludes the answer: a
// consequences argument backing "is there a better means?" concludes that
// another action is better, so the consequences scheme's question about its
// conclusion (side effects of doing it) does not apply to it. The playtest
// of 2026-10-13 showed the prompt offering it, with the answers reversed.
// (The shipped data no longer links consequences to that question, after the
// audit's N9: a good consequence of another action does not show it is the
// better means. The edge is added back here to keep the rule tested.)
func TestConclusionQuestionsOfAnswerArguments(t *testing.T) {
	edges := strings.Replace(string(data.EdgesJSON), `"backup_edges": [`, `"backup_edges": [ { "scheme": "AS:consequences", "cq": "CQ:other_means" },`, 1)
	gs := newShippedGame(t, []byte(edges), "Alice", "Bob")
	gs.Config.MaxTurns = 0
	a1 := play(t, gs, prSupportAction())
	a2 := play(t, gs, answer("attack", "CQ:other_means", "AS:consequences", a1, map[string]string{"A": "subsidize insulation", "G": "lower emissions"}))
	if c := gs.Adapter.Graph.Arguments[a2].Conclusion.Id; !strings.HasPrefix(c, "realize_more_effectively(") {
		t.Fatalf("the answer concludes %s", c)
	}
	for _, tmpl := range gs.NASCQ.EnumerateActions(gs.Bridge(), []string{a2}) {
		if tmpl.Locus.Kind == nascq.LocusConclusion {
			t.Errorf("offered %s about the conclusion of %s", tmpl.AnswersCQID, a2)
		}
	}
	r, err := gs.ExecuteTurn(answer("attack", "CQ:side_effect", "AS:sign", a2, map[string]string{"A": "x"}))
	if err != nil || r.Accepted || !strings.Contains(r.Error, "does not apply") {
		t.Fatalf("side effects of %s: %+v %v", a2, r, err)
	}
	// Its other questions still apply, and its conclusion can be argued
	// against as a statement.
	play(t, gs, nascq.ProposedAction{Type: "attack", TargetStmtID: gs.Adapter.Graph.Arguments[a2].Conclusion.Id, SchemeID: "AS:position_to_know",
		VariableBindings: map[string]string{"W": "an energy auditor", "D": "home insulation"}})
}

// Negative consequences concludes ¬should_be_performed(A), so it can argue
// against what should be done, with A fixed by the statement; goal-based
// practical reasoning stands beside the value-based one, with its own
// questions.
func TestNegativeConsequencesAndGoalBasedReasoning(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.MaxTurns = 0
	const shouldTax = "should_be_performed(carbon_tax)"
	a1 := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:practical_reasoning_goal",
		VariableBindings: map[string]string{"A": "carbon tax", "S1": "rising emissions", "S2": "lower emissions", "G": "the 2030 target"}})
	if c := gs.Adapter.Graph.Arguments[a1].Conclusion.Id; c != shouldTax {
		t.Fatalf("a1 concludes %s", c)
	}
	cqs := map[string]bool{}
	for _, tmpl := range gs.NASCQ.EnumerateActions(gs.Bridge(), []string{a1}) {
		cqs[tmpl.AnswersCQID] = true
	}
	for _, cq := range []string{"CQ:conflicting_goals", "CQ:intervening_actions", "CQ:goal_possible", "CQ:possibility", "CQ:side_effect"} {
		if !cqs[cq] {
			t.Errorf("%s is not offered on goal-based practical reasoning (%v)", cq, cqs)
		}
	}
	a2 := play(t, gs, nascq.ProposedAction{Type: "attack", TargetStmtID: shouldTax, SchemeID: "AS:negative_consequences",
		VariableBindings: map[string]string{"G": "higher heating bills for poor households"}})
	arg := gs.Adapter.Graph.Arguments[a2]
	if arg.Conclusion.Id != "¬"+shouldTax || arg.Premises[0].Stmt.Id != "brings_about(carbon_tax,higher_heating_bills_for_poor_households)" {
		t.Fatalf("a2 concludes %s from %s", arg.Conclusion.Id, arg.Premises[0].Stmt.Id)
	}
	// 0.65 against 0.65: a tie under PE leaves the claim out.
	if l := gs.Adapter.Graph.Statements[shouldTax].Label; l != caes.Out {
		t.Fatalf("the claim is %v", l)
	}
	// Alice answers the Opposite Consequences question with positive
	// consequences: a rebuttal of a2, for the claim.
	play(t, gs, answer("attack", "CQ:opposite_consequences", "AS:consequences", a2, map[string]string{"A": "carbon tax", "G": "cleaner air"}))
	if got := gs.Say("¬" + shouldTax); got != "‘Carbon tax’ should not be done." {
		t.Fatalf("the negation reads %q", got)
	}
}

// An answer whose backing scheme concludes the answer's form must be about
// the same action: positive consequences rebutting a negative-consequences
// argument against the carbon tax must be about the carbon tax.
func TestAnswerFixesTheSchemesConclusion(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.MaxTurns = 0
	a1 := play(t, gs, prSupportAction())
	_ = a1
	a2 := play(t, gs, nascq.ProposedAction{Type: "attack", TargetStmtID: "should_be_performed(carbon_tax)", SchemeID: "AS:negative_consequences",
		VariableBindings: map[string]string{"G": "higher heating bills"}})
	r, err := gs.ExecuteTurn(answer("attack", "CQ:opposite_consequences", "AS:consequences", a2, map[string]string{"A": "a fuel subsidy", "G": "cheaper travel"}))
	if err != nil || r.Accepted || !strings.Contains(r.Error, "S9") {
		t.Fatalf("an answer about another action: %+v %v", r, err)
	}
	a3 := play(t, gs, answer("attack", "CQ:opposite_consequences", "AS:consequences", a2, map[string]string{"G": "cleaner air"}))
	if p := gs.Adapter.Graph.Arguments[a3].Premises[0].Stmt.Id; p != "brings_about(carbon_tax,cleaner_air)" {
		t.Fatalf("a3's premise: %s", p)
	}
}
