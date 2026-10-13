package game

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
)

// Two humans play through the CLI wizard: Alice supports practical reasoning
// (variables asked in the order the premises use them), Bob asks her the
// possibility question (shown with her words) and undermines possible(A)
// with an argument from sign, whose answer variable B the game binds to the
// answer (so he is asked for A only).
func TestHumanWizardSupportThenUndermining(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Alice", "Bob")
	var out bytes.Buffer
	alice := NewHumanAgent("Alice", strings.NewReader(strings.Join([]string{
		"1",                                                                                       // support
		"5",                                                                                       // Practical Reasoning (schemes are listed by NASCQ ID)
		"status quo", "carbon tax", "lower emissions", "reduce emissions", "environmental health", // S1 A S2 G V
	}, "\n")+"\n"), &out)
	act := alice.ChooseAction(gs.BuildObservation())
	if act.SchemeID != "AS:practical_reasoning" || act.VariableBindings["A"] != "carbon tax" || act.VariableBindings["S1"] != "status quo" {
		t.Fatalf("Alice's action: %+v\n%s", act, out.String())
	}
	if r, _ := gs.ExecuteTurn(act); !r.Accepted {
		t.Fatalf("rejected: %s", r.Error)
	}
	if !strings.Contains(out.String(), "therefore: Action carbon tax should be performed") {
		t.Errorf("no preview of the support:\n%s", out.String())
	}

	out.Reset()
	bob := NewHumanAgent("Bob", strings.NewReader(strings.Join([]string{
		"2",                         // attack
		"1",                         // the only target, Alice's argument
		"4",                         // the possibility question
		"1",                         // Argument from Sign
		"no parliamentary majority", // A (B is the answer)
	}, "\n")+"\n"), &out)
	act = bob.ChooseAction(gs.BuildObservation())
	shown := out.String()
	for _, want := range []string{
		"Is it actually possible to perform action carbon tax?",
		"therefore: It is not the case that action carbon tax is possible (against a premise of a1).",
		"no parliamentary majority is observed to be the case",
		"“it is not the case that action carbon tax is possible” is indicated when no parliamentary majority is true",
	} {
		if !strings.Contains(shown, want) {
			t.Errorf("wizard output lacks %q:\n%s", want, shown)
		}
	}
	r, _ := gs.ExecuteTurn(act)
	if !r.Accepted || gs.Adapter.Graph.Arguments[r.ArgIDs[0]].Conclusion.Id != "¬possible(carbon_tax)" {
		t.Fatalf("Bob's attack: %+v", r)
	}
	board := RenderState(gs)
	for _, want := range []string{"[" + r.ArgIDs[0] + "]", "Out ✗     Action carbon tax is possible."} {
		if !strings.Contains(board, want) {
			t.Errorf("board lacks %q:\n%s", want, board)
		}
	}
}
