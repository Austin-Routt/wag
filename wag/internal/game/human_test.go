package game

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/nascq"
)

// Two humans play through the CLI wizard: Alice supports practical reasoning
// (variables asked in the order the premises use them), Bob asks her the
// possibility question (shown with her words) and undermines possible(A)
// with an argument from sign, whose answer variable B the game binds to the
// answer (so he is asked for A only).
func TestHumanWizardSupportThenUndermining(t *testing.T) {
	gs := newShippedGame(t, prevEdges, "Alice", "Bob")
	var out bytes.Buffer
	alice := NewHumanAgent("Alice", strings.NewReader(strings.Join([]string{
		"1",                                                                                       // support
		"6",                                                                                       // Practical Reasoning (schemes are listed by NASCQ ID)
		"status quo", "carbon tax", "lower emissions", "reduce emissions", "environmental health", // S1 A S2 G V
	}, "\n")+"\n"), &out)
	act := alice.ChooseAction(gs.BuildObservation())
	if act.SchemeID != "AS:practical_reasoning" || act.VariableBindings["A"] != "carbon tax" || act.VariableBindings["S1"] != "status quo" {
		t.Fatalf("Alice's action: %+v\n%s", act, out.String())
	}
	if r, _ := gs.ExecuteTurn(act); !r.Accepted {
		t.Fatalf("rejected: %s", r.Error)
	}
	if !strings.Contains(out.String(), "You argue that C: ‘Carbon tax’ should be done.") {
		t.Errorf("no preview of the support:\n%s", out.String())
	}

	out.Reset()
	bob := NewHumanAgent("Bob", strings.NewReader(strings.Join([]string{
		"3",                         // attack
		"1",                         // the only target, Alice's argument
		"4",                         // the possibility question
		"1",                         // Argument from Sign
		"no parliamentary majority", // A (B is the answer)
	}, "\n")+"\n"), &out)
	act = bob.ChooseAction(gs.BuildObservation())
	shown := out.String()
	for _, want := range []string{
		"Is ‘carbon tax’ actually possible?",
		"therefore: ‘Carbon tax’ is not possible (against a premise of a1).",
		"‘No parliamentary majority’ is observed.",
		"C is indicated when ‘no parliamentary majority’ holds.",
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
	for _, want := range []string{"[" + r.ArgIDs[0] + "]", "Out ✗     ‘Carbon tax’ is possible."} {
		if !strings.Contains(board, want) {
			t.Errorf("board lacks %q:\n%s", want, board)
		}
	}
}

// The wizard plays a move about a statement, naming no critical question:
// Bob argues against the premise that the tax is possible, with an argument
// from position to know (an undermining), as the browser client and the
// models can (docs/game-rules.md).
func TestHumanWizardArguesAboutAStatement(t *testing.T) {
	gs := newShippedGame(t, prevEdges, "Alice", "Bob")
	if r, _ := gs.ExecuteTurn(prSupportAction()); !r.Accepted {
		t.Fatalf("opening move refused: %s", r.Error)
	}
	const target = "possible(carbon_tax)"
	obs := gs.BuildObservation()
	var stmts, sides, schemes []string
	var chosen nascq.ActionTemplate
	seen := map[string]bool{}
	for _, tmpl := range obs.AvailableActions {
		if tmpl.TargetStmtID != "" && !seen[tmpl.TargetStmtID] {
			seen[tmpl.TargetStmtID] = true
			stmts = append(stmts, tmpl.TargetStmtID)
		}
	}
	for _, typ := range []string{"support", "attack"} {
		for _, tmpl := range obs.AvailableActions {
			if tmpl.TargetStmtID == target && tmpl.Type == typ {
				sides = append(sides, typ)
				break
			}
		}
	}
	for _, tmpl := range obs.AvailableActions {
		if tmpl.TargetStmtID == target && tmpl.Type == "attack" {
			schemes = append(schemes, tmpl.SchemeID)
			if tmpl.SchemeID == "AS:position_to_know" {
				chosen = tmpl
			}
		}
	}
	pick := func(list []string, want string) string {
		t.Helper()
		for i, x := range list {
			if x == want {
				return fmt.Sprint(i + 1)
			}
		}
		t.Fatalf("%s is not among %v", want, list)
		return ""
	}
	input := []string{"2", pick(stmts, target)} // for or against a statement; the tax's possibility
	if len(sides) == 2 {
		input = append(input, pick(sides, "attack")) // against it
	}
	input = append(input, pick(schemes, "AS:position_to_know"))
	values := map[string]string{"W": "the finance minister", "D": "the budget"}
	for _, v := range variableOrder(chosen) {
		input = append(input, values[v])
	}
	var out bytes.Buffer
	bob := NewHumanAgent("Bob", strings.NewReader(strings.Join(input, "\n")+"\n"), &out)
	act := bob.ChooseAction(obs)
	if act.Type != "attack" || act.TargetStmtID != target || act.SchemeID != "AS:position_to_know" || act.AnswersCQID != "" || act.TargetArgID != "" {
		t.Fatalf("Bob's move: %+v\n%s", act, out.String())
	}
	for _, want := range []string{"[2] For or against", "Which statement?", "therefore: ‘Carbon tax’ is not possible."} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("wizard output lacks %q:\n%s", want, out.String())
		}
	}
	r, _ := gs.ExecuteTurn(act)
	if !r.Accepted || gs.Adapter.Graph.Arguments[r.ArgIDs[0]].Conclusion.Id != "¬"+target {
		t.Fatalf("Bob's argument: %+v", r)
	}
	if m := gs.MoveHistory[len(gs.MoveHistory)-1]; m.Action.AttackType != "undermining" {
		t.Errorf("recorded as %+v", m.Action)
	}
}
