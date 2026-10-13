package game

// Answers to critical questions defend as well as attack, at any depth: a
// positive answer backs the premise (or denies the exception) the question is
// about, and is itself an argument whose critical questions can be answered,
// so each answer protects the one above it until it is knocked out.

import (
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// play executes an action that must be accepted and returns its main argument.
func play(t *testing.T, gs *GameState, a nascq.ProposedAction) string {
	t.Helper()
	r, err := gs.ExecuteTurn(a)
	if err != nil || !r.Accepted {
		t.Fatalf("%s %s/%s on %s rejected: %v %+v", a.Type, a.AnswersCQID, a.SchemeID, a.TargetArgID, err, r)
	}
	if errs := gs.CheckInvariants(); len(errs) > 0 {
		t.Fatalf("after %s %s on %s: %v", a.Type, a.AnswersCQID, a.TargetArgID, errs)
	}
	return r.ArgIDs[0]
}

func answer(typ, cq, scheme, target string, bindings map[string]string) nascq.ProposedAction {
	return nascq.ProposedAction{Type: typ, AnswersCQID: cq, SchemeID: scheme, TargetArgID: target, VariableBindings: bindings}
}

func TestDefenceInDepth(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.MaxTurns = 0
	label := func(id string) caes.Label { return gs.Adapter.Graph.Statements[id].Label }
	const shouldTax, possibleTax = "should_be_performed(carbon_tax)", "possible(carbon_tax)"

	// 1. Alice: the carbon tax should be performed.
	a1 := play(t, gs, prSupportAction())
	if label(shouldTax) != caes.In {
		t.Fatal("the unchallenged argument's conclusion is not In")
	}

	// 2. Bob: "Is it possible?" — no, says an expert: ¬possible(carbon_tax).
	// The answer variable S is bound to the answer, so the expert's premises
	// are about it.
	b1 := play(t, gs, answer("attack", "CQ:possibility", "AS:expert_opinion", a1,
		map[string]string{"W": "dr_no", "D": "politics"}))
	arg := gs.Adapter.Graph.Arguments[b1]
	if arg.Conclusion.Id != "¬"+possibleTax || !gs.bridge.PremiseExists(b1, "asserts(dr_no,not(possible(carbon_tax)))") {
		t.Fatalf("expert's answer: %s from %v", arg.Conclusion.Id, arg.Premises)
	}
	if rec := gs.MoveHistory[len(gs.MoveHistory)-1].Action.VariableBindings["S"]; rec != "¬"+possibleTax {
		t.Fatalf("the history records S = %q, want the answer", rec)
	}
	if label(possibleTax) != caes.Out || label(shouldTax) != caes.Out {
		t.Fatalf("after the undermining: possible %v, should %v; want Out, Out", label(possibleTax), label(shouldTax))
	}

	// 3. Alice, a level down: "Is dr_no trustworthy?" — no, he is biased,
	// says someone in a position to know. The exception undercuts b1, so the
	// premise of a1 stands again.
	a2 := play(t, gs, answer("attack", "CQ:trustworthiness", "AS:position_to_know", b1,
		map[string]string{"W": "journalist", "D": "media"}))
	if c := gs.Adapter.Graph.Arguments[a2].Conclusion.Id; c != "biased(dr_no)" {
		t.Fatalf("trustworthiness answer concludes %s", c)
	}
	if gs.Adapter.Graph.Arguments[b1].Weight != 0 || label(possibleTax) != caes.In || label(shouldTax) != caes.In {
		t.Fatalf("after the undercut: b1 weight %v, possible %v, should %v; want 0, In, In",
			gs.Adapter.Graph.Arguments[b1].Weight, label(possibleTax), label(shouldTax))
	}

	// 4. Bob, another level down: "Is the journalist in a position to know?"
	// — no, says an expert. a2 loses its premise, so b1 and the undermining
	// are back.
	play(t, gs, answer("attack", "CQ:position", "AS:expert_opinion", a2,
		map[string]string{"W": "press_council", "D": "media"}))
	if label("biased(dr_no)") != caes.Out || label(possibleTax) != caes.Out || label(shouldTax) != caes.Out {
		t.Fatalf("after the counter-undermining: biased %v, possible %v, should %v; want Out, Out, Out",
			label("biased(dr_no)"), label(possibleTax), label(shouldTax))
	}

	// 5. Alice defends a1's premise directly: a positive answer to the
	// possibility question, backed by an expert as strong as Bob's. A tie
	// leaves both positions Out under PE, so the premise is still not In.
	play(t, gs, answer("support", "CQ:possibility", "AS:expert_opinion", a1,
		map[string]string{"W": "dr_yes", "D": "politics"}))
	if label(possibleTax) != caes.Out || label("¬"+possibleTax) != caes.Out {
		t.Fatalf("a tie: possible %v, ¬possible %v; want Out, Out", label(possibleTax), label("¬"+possibleTax))
	}

	// 6. Bob abstains. 7. Alice re-backs her journalist's premise with an
	// expert (a positive answer to the position question): 0.70 against the
	// press council's 0.70 is a tie, so the premise is Out and a2 still fails.
	gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	play(t, gs, answer("support", "CQ:position", "AS:expert_opinion", a2,
		map[string]string{"W": "media_scholar", "D": "media"}))
	if label("biased(dr_no)") != caes.Out {
		t.Fatalf("biased is %v after a tied defence, want Out", label("biased(dr_no)"))
	}
	t.Logf("scores: Alice %.1f, Bob %.1f", gs.Players[0].Score, gs.Players[1].Score)
}

// A positive answer to an exception question denies the exception: it makes
// E an issue, which a stronger denial wins and a tie leaves Out, so the
// argument stands either way.
func TestExceptionDefence(t *testing.T) {
	for _, c := range []struct {
		name, defender string
		wantBiased     caes.Label
	}{
		{"stronger denial", "AS:expert_opinion", caes.Out}, // 0.70 > 0.65
		{"equal denial", "AS:position_to_know", caes.Out},  // 0.65 = 0.65: a tie
		{"weaker denial", "AS:analogy", caes.In},           // 0.55 < 0.65
	} {
		t.Run(c.name, func(t *testing.T) {
			gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
			b1 := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:expert_opinion",
				VariableBindings: map[string]string{"W": "dr_x", "D": "economics", "S": "tax_works"}})
			play(t, gs, answer("attack", "CQ:trustworthiness", "AS:position_to_know", b1,
				map[string]string{"W": "journalist", "D": "media"}))
			if gs.Adapter.Graph.Statements["tax_works"].Label != caes.Out {
				t.Fatal("the undercut argument's conclusion is not Out")
			}
			bindings := map[string]string{"W": "auditor", "D": "media"}
			if c.defender == "AS:analogy" {
				bindings = map[string]string{"C": "earlier_audit"}
			}
			d := play(t, gs, answer("support", "CQ:trustworthiness", c.defender, b1, bindings))
			if got := gs.Adapter.Graph.Arguments[d].Conclusion.Id; got != "¬biased(dr_x)" {
				t.Fatalf("the denial concludes %s", got)
			}
			biased := gs.Adapter.Graph.Statements["biased(dr_x)"].Label
			wantConcl := caes.In
			if c.wantBiased == caes.In {
				wantConcl = caes.Out
			}
			if biased != c.wantBiased || gs.Adapter.Graph.Statements["tax_works"].Label != wantConcl {
				t.Fatalf("biased %v, tax_works %v; want %v, %v", biased, gs.Adapter.Graph.Statements["tax_works"].Label, c.wantBiased, wantConcl)
			}
		})
	}
}

// Attacking an argument that is already defeated earns nothing, and neither
// does attacking one's own.
func TestAttacksOnDefeatedOrOwnArgumentsScoreNothing(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	a1 := play(t, gs, prSupportAction())
	r, _ := gs.ExecuteTurn(answer("attack", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "dr_no", "D": "politics"}))
	if r.ScoreDelta <= 0 {
		t.Fatalf("a successful undermining scored %v", r.ScoreDelta)
	}
	r, _ = gs.ExecuteTurn(answer("attack", "CQ:possibility", "AS:position_to_know", a1, map[string]string{"W": "own_doubt", "D": "politics"}))
	if !r.Accepted || r.ScoreDelta != 0 {
		t.Fatalf("Alice attacking her own argument: accepted %v, score %v; want accepted, 0", r.Accepted, r.ScoreDelta)
	}
	r, _ = gs.ExecuteTurn(answer("attack", "CQ:side_effect", "AS:sign", a1, map[string]string{"A": "riots", "S1": "x"}))
	if !r.Accepted || r.ScoreDelta != 0 {
		t.Fatalf("Bob attacking the defeated a1: accepted %v (%s), score %v; want accepted, 0", r.Accepted, r.Error, r.ScoreDelta)
	}
}

// The proleptic bonus rewards answering a question about one's own argument
// before anyone has raised it, and only then.
func TestProlepticBonus(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	gs.Config.Scoring.ProlepticBonus = 5
	a1 := play(t, gs, prSupportAction())
	gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
	r, _ := gs.ExecuteTurn(answer("support", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "dr_yes", "D": "politics"}))
	plain := basePoints(caes.In, gs.Config.Scoring) * 0.70
	if !r.Accepted || r.ScoreDelta != plain+5 {
		t.Fatalf("proleptic defence scored %v, want %v", r.ScoreDelta, plain+5)
	}
	// Bob raises another question; Alice's answer to it is not proleptic.
	gs.ExecuteTurn(answer("attack", "CQ:best_means", "AS:expert_opinion", a1, map[string]string{"W": "dr_b", "D": "policy"}))
	r, _ = gs.ExecuteTurn(answer("support", "CQ:best_means", "AS:position_to_know", a1, map[string]string{"W": "minister", "D": "policy"}))
	if !r.Accepted || r.ScoreDelta >= 5 {
		t.Fatalf("a reply to a raised question scored %v (accepted %v, %s), want no bonus", r.ScoreDelta, r.Accepted, r.Error)
	}
}

// Two players with the same name get distinct IDs, so neither's attacks on
// the other count as self-attacks.
func TestDuplicatePlayerNamesGetDistinctIDs(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "A", "A", "A")
	ids := map[string]bool{}
	for _, p := range gs.Players {
		ids[p.ID] = true
	}
	if len(ids) != 3 || !ids["A"] || !ids["A#2"] || !ids["A#3"] {
		t.Fatalf("player IDs %v", ids)
	}
	a1 := play(t, gs, prSupportAction())
	r, _ := gs.ExecuteTurn(answer("attack", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "dr_no", "D": "politics"}))
	if r.ScoreDelta <= 0 {
		t.Fatalf("the second A's attack on the first's argument scored %v", r.ScoreDelta)
	}
}
