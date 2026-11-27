package game

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/nascq"
)

// The fixes of the independent check of plan version 5a and the profile's
// code (wag-workspace verify/c5/c52/CHECK-5a.md), each with the case that
// showed the fault.

// A declaration of an action the profile holds, with the same actor, adds
// only its new aliases: the board and a model that declare on every use are
// not refused (W-a, N-e). Actors and the agent are compared as atoms, as
// actions are.
func TestProfileRedeclarationAndAtoms(t *testing.T) {
	gs := profileGame(t, caseProfile(t))
	same := cons("Tax Meat")
	same.Declare = []nascq.ActionDecl{{ID: "Tax Meat", Actor: "Governments", Aliases: []string{"the meat levy"}}}
	id := play(t, gs, same)
	if c := gs.Adapter.Graph.Arguments[id].Conclusion.Id; c != "should_be_performed(tax_meat)" {
		t.Errorf("the redeclared action's argument concludes %s", c)
	}
	if a, ok := gs.Profile.Action("the meat levy"); !ok || a.ID != "tax_meat" || a.Actor != "governments" {
		t.Errorf("the new alias: %+v %v", a, ok)
	}
	if n := len(gs.Profile.Actions); n != len(caseProfile(t).Actions) {
		t.Errorf("a redeclaration added an action: %d actions", n)
	}
	// A declaration whose names name two declared actions is refused.
	two := cons("tax_meat")
	two.Declare = []nascq.ActionDecl{{ID: "tax_meat", Actor: "governments", Aliases: []string{"end subsidies"}}}
	refused(t, gs, two, "P1: end subsidies is already declared (the action end_subsidies)")
	// The agent written otherwise is the agent (P2).
	pr := nascq.ProposedAction{Type: "support", SchemeID: "AS:pr_agent",
		VariableBindings: map[string]string{"A": "tax_meat", "G": "lower_emissions", "P": "Governments"}}
	play(t, gs, pr)
}

// A question argues about no action: it declares none, and the scheme and
// values it may carry are set aside (S0), not checked against the profile
// (W-b, N-i).
func TestProfileQuestions(t *testing.T) {
	gs := profileGame(t, caseProfile(t))
	target := play(t, gs, cons("tax_meat"))
	q := nascq.ProposedAction{Type: "question", TargetArgID: target, AnswersCQID: "CQ:evidence",
		Declare: []nascq.ActionDecl{{ID: "ban_meat", Actor: "governments"}}}
	refused(t, gs, q, "P1: the move declares ban_meat but argues about no action")
	q.Declare = nil
	q.SchemeID, q.VariableBindings = "AS:consequences", map[string]string{"A": "whatever"}
	play(t, gs, q)
}

// A claim written with an alias, or with spaces, is the claim of its
// action's id: an argument for the action concludes it (W-c, N-f).
func TestProfileClaimWritten(t *testing.T) {
	base := profileGame(t, nil)
	for claim, want := range map[string]string{
		"should_be_performed(meat_tax)":     "should_be_performed(tax_meat)",
		"¬should_be_performed( Meat Tax )":  "¬should_be_performed(tax_meat)",
		"should_be_performed( ban meat )":   "should_be_performed(ban_meat)",
		"should_be_performed(tax_meat)":     "should_be_performed(tax_meat)",
		"lower emissions matter":            "lower emissions matter",
		"¬ should_be_performed(fish quota)": "¬should_be_performed(fish_quota)",
	} {
		cfg := DefaultGameConfig()
		cfg.MaxTurns, cfg.Objective, cfg.Claim, cfg.Profile = 0, ObjectiveClaim, claim, caseProfile(t)
		gs := NewGame(cfg, base.Adapter, base.NASCQ, []string{"Alice", "Bob"})
		if gs.Config.Claim != want {
			t.Errorf("%q is played as %q, want %q", claim, gs.Config.Claim, want)
		}
		if a := claimAction(want); a != "" {
			if _, ok := gs.Profile.Action(a); !ok {
				t.Errorf("%q: its action %s is not declared", claim, a)
			}
		}
	}
	cfg := DefaultGameConfig()
	cfg.MaxTurns, cfg.Objective, cfg.Claim, cfg.Profile = 0, ObjectiveClaim, "should_be_performed(meat_tax)", caseProfile(t)
	gs := NewGame(cfg, base.Adapter, base.NASCQ, []string{"Alice", "Bob"})
	play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:consequences",
		VariableBindings: map[string]string{"A": "meat_tax", "G": "less_meat_eaten"}})
	if gs.Claim != "should_be_performed(tax_meat)" {
		t.Errorf("the opening argument for the alias states the claim %q", gs.Claim)
	}
}

// The game validates a move as a turn would, without playing it: the
// profile's rules around NASCQ's, so that a preview accepts an agent place
// the game binds (W-d) and refuses an undeclared action.
func TestValidateAsATurn(t *testing.T) {
	gs := profileGame(t, caseProfile(t))
	va, err := gs.Validate(nascq.ProposedAction{Type: "support", SchemeID: "AS:pr_agent",
		VariableBindings: map[string]string{"A": "meat_tax", "G": "lower_emissions"}})
	if err != nil {
		t.Fatal(err)
	}
	if va.VariableBindings["P"] != "governments" || va.VariableBindings["A"] != "tax_meat" {
		t.Errorf("the validated bindings: %v", va.VariableBindings)
	}
	if _, err := gs.Validate(cons("ban_fish")); err == nil || !strings.Contains(err.Error(), "P1:") {
		t.Errorf("an undeclared action: %v", err)
	}
	if len(gs.MoveHistory) != 0 || gs.Profile.Agent != "governments" {
		t.Error("validating a move played it")
	}
}

// Every player sees the agent's place as fixed by the game, a person at the
// terminal included, who declares a new action with the actor they choose
// (E-d).
func TestHumanDeclaresNewActions(t *testing.T) {
	gs := profileGame(t, caseProfile(t))
	obs := gs.BuildObservation()
	var tmpl *nascq.ActionTemplate
	for i, a := range obs.AvailableActions {
		if a.SchemeID == "AS:pr_agent" && a.Type == "support" && a.TargetArgID == "" && a.TargetStmtID == "" {
			tmpl = &obs.AvailableActions[i]
		}
	}
	if tmpl == nil || tmpl.Fixed["P"] != "governments" {
		t.Fatalf("the agent's place is not fixed for the players: %+v", tmpl)
	}
	for _, v := range variableOrder(*tmpl) {
		if v == "P" {
			t.Error("the terminal asks for the agent")
		}
	}
	var out bytes.Buffer
	h := NewHumanAgent("Alice", strings.NewReader("2\n"), &out)
	decl := h.declarations(gs, *tmpl, map[string]string{"A": "ban fish", "G": "lower_emissions"})
	if len(decl) != 1 || decl[0].ID != "ban fish" || decl[0].Actor != "farmers" {
		t.Fatalf("the declarations: %+v\n%s", decl, out.String())
	}
	if !strings.Contains(out.String(), `"ban fish" is a new action: whose action is it?`) {
		t.Errorf("the terminal did not ask whose action it is:\n%s", out.String())
	}
	if d := h.declarations(gs, *tmpl, map[string]string{"A": "Tax-Meat"}); len(d) != 0 {
		t.Errorf("a declared action is declared again: %+v", d)
	}
	// Declared as the agent's, the move plays.
	h = NewHumanAgent("Alice", strings.NewReader("1\n"), &out)
	a := tmpl.Action(map[string]string{"A": "ban fish", "G": "lower_emissions"})
	a.Declare = h.declarations(gs, *tmpl, a.VariableBindings)
	play(t, gs, a)
	if act, ok := gs.Profile.Action("ban_fish"); !ok || act.Actor != "governments" {
		t.Errorf("the terminal's declaration: %+v", act)
	}
}
