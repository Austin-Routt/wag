package game

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/nascq"
)

// Every action a View offers is a legal move once its variables are bound,
// in random play on the widest NASCQ fixture, and every View is valid JSON.
func TestViewActionsAreLegalMoves(t *testing.T) {
	gs := newShippedGame(t, everyCQAsksEverySchemeEdges(t), "A", "B")
	agents := []Agent{NewRandomAgent("A", 3, 0.05, 0.4), NewRandomAgent("B", 4, 0.05, 0.4)}
	for move := 0; !gs.Terminated && move < 14; move++ {
		v := gs.BuildView()
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		var back View
		if err := json.Unmarshal(b, &back); err != nil || len(back.Actions) != len(v.Actions) {
			t.Fatalf("view does not round-trip: %v", err)
		}
		for _, a := range v.Actions {
			bind := map[string]string{}
			for _, name := range a.VariableOrder {
				bind[name] = "x"
			}
			act := nascq.ProposedAction{Type: a.Type, AttackType: a.AttackType, SchemeID: a.SchemeID, AnswersCQID: a.AnswersCQID,
				TargetArgID: a.TargetArgID, TargetPremiseID: a.TargetPremiseID, VariableBindings: bind}
			if _, err := gs.NASCQ.Validate(act, gs.bridge); err != nil {
				t.Fatalf("offered action %+v is not legal: %v", a, err)
			}
			if len(a.VariableOrder) != len(a.Variables) {
				t.Fatalf("variable order %v for %d variables", a.VariableOrder, len(a.Variables))
			}
		}
		if r, _ := gs.ExecuteTurn(agents[gs.ActivePlayerIdx].ChooseAction(gs.BuildObservation())); !r.Accepted {
			gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
		}
	}
}

// A JSONAgent writes a View per decision and reads an action per line; a
// line that is not an action and a rejected move are both explained in the
// next View's last_rejection.
func TestJSONAgent(t *testing.T) {
	gs := newShippedGame(t, data.EdgesJSON, "Bot", "Other")
	in := strings.Join([]string{
		`{"type":"support","scheme_id":"AS:practical_reasoning","variable_bindings":{"A":"carbon tax","S1":"status quo","S2":"lower emissions","G":"reduce emissions","V":"health"}}`,
		`not json`,
		`{"type":"attack","scheme_id":"AS:sign","answers_cq_id":"CQ:likelihood","target_arg_id":"a1","variable_bindings":{"A":"x","B":"y"}}`,
		`{"type":"attack","scheme_id":"AS:sign","answers_cq_id":"CQ:possibility","target_arg_id":"a1","variable_bindings":{"A":"no majority","B":"cannot pass"}}`,
	}, "\n") + "\n"
	var out bytes.Buffer
	bot := NewJSONAgent("Bot", strings.NewReader(in), &out)
	play := func() *TurnResult {
		r, err := gs.ExecuteTurn(bot.ChooseAction(gs.BuildObservation()))
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	if r := play(); !r.Accepted {
		t.Fatalf("support: %s", r.Error)
	}
	gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"}) // Other
	if r := play(); r.Accepted {
		t.Fatal("a line that is not JSON was accepted")
	}
	if r := play(); r.Accepted || !strings.HasPrefix(r.Error, "S3") {
		t.Fatalf("an inapplicable CQ: %+v", r)
	}
	if r := play(); !r.Accepted {
		t.Fatalf("undermining: %s", r.Error)
	}
	if r := play(); !r.Abstained { // end of input
		t.Fatalf("end of input: %+v", r)
	}

	var views []View
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		var v View
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			t.Fatalf("not a View: %q", line)
		}
		views = append(views, v)
	}
	if len(views) != 5 {
		t.Fatalf("%d views, want 5", len(views))
	}
	if !strings.HasPrefix(views[2].LastRejection, "your line is not an action") {
		t.Errorf("view after the bad line: %q", views[2].LastRejection)
	}
	if !strings.HasPrefix(views[3].LastRejection, "S3") {
		t.Errorf("view after the rejected move: %q", views[3].LastRejection)
	}
	last := views[4]
	found := false
	for _, s := range last.Statements {
		found = found || (s.ID == "possible(carbon_tax)" && s.Label == "out" && s.Issue == "issue:possible(carbon_tax)")
	}
	if !found || len(last.Issues) != 1 {
		t.Errorf("the final view does not show the undermined premise: %+v", last.Statements)
	}
}
