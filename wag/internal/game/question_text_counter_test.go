package game

import (
	"reflect"
	"testing"

	"github.com/yourorg/wag/internal/nascq"
)

func TestQuestionTextOtherCounter(t *testing.T) {
	for _, tc := range []struct {
		name string
		seed nascq.ProposedAction
		want string
	}{
		{
			name: "existing_free_text_target",
			seed: free("Schools should stay open.", "Smaller cohorts limit crowding."),
			want: "Is there a stronger reason against the conclusion “Schools should stay open”?",
		},
		{
			name: "existing_structured_target",
			seed: nascq.ProposedAction{Type: "support", SchemeID: "AS:consequences",
				VariableBindings: map[string]string{"A": "reduce cohort sizes", "G": "lower transmission"}},
			want: "Is there a stronger reason against the conclusion “‘reduce cohort sizes’ should be done”?",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gs := newDataGame(t, "Alice", "Bob")
			gs.Config.AllowOther = true
			gs.NASCQ.EnableOther()
			seed := play(t, gs, tc.seed)
			target := gs.Adapter.Graph.Arguments[seed].Conclusion.Id
			id := play(t, gs, nascq.ProposedAction{
				Type: "support", SchemeID: nascq.OtherSchemeID, TargetStmtID: target,
				VariableBindings: map[string]string{"P1": "There is an independent supporting reason."},
			})
			before := gs.Adapter.ArgumentBindings(id)
			if _, ok := before["C"]; ok {
				t.Fatal("fixture should have no recorded C binding")
			}
			if got := gs.QuestionText(nascq.CQCounter, id); got != tc.want {
				t.Errorf("QuestionText\n got %q\nwant %q", got, tc.want)
			}
			if got := gs.Adapter.ArgumentBindings(id); !reflect.DeepEqual(got, before) {
				t.Errorf("rendering changed recorded bindings: before %v, after %v", before, got)
			}
			if got := gs.Adapter.Graph.Arguments[id].Conclusion.Id; got != target {
				t.Errorf("rendering changed conclusion: got %q, want %q", got, target)
			}
		})
	}

	t.Run("existing_C_binding_is_preserved", func(t *testing.T) {
		gs := newDataGame(t, "Alice", "Bob")
		gs.Config.AllowOther = true
		gs.NASCQ.EnableOther()
		id := play(t, gs, free("The stored conclusion.", "A supporting premise."))
		// A historical/imported Other argument may already carry a C binding.
		arg := gs.Adapter.Graph.Arguments[id]
		arg.Scheme.Variables = []string{"C"}
		arg.Parameters = []string{"recorded_wording"}
		const want = "Is there a stronger reason against the conclusion ‘recorded wording’?"
		if got := gs.QuestionText(nascq.CQCounter, id); got != want {
			t.Errorf("QuestionText = %q, want %q", got, want)
		}
		if got := gs.Adapter.ArgumentBindings(id)["C"]; got != "recorded_wording" {
			t.Errorf("rendering changed C binding to %q", got)
		}
	})
}

func TestQuestionTextAnalogyCaseBinding(t *testing.T) {
	// The historical playable questions include C as the analogous case,
	// not as the argument's conclusion.
	gs := newShippedGame(t, prevEdges, "Alice", "Bob")
	id := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:analogy",
		VariableBindings: map[string]string{"A": "schools can remain open", "C": "another city"}})
	const want = "Does the case of ‘another city’ differ from this one in relevant ways?"
	if got := gs.QuestionText("CQ:difference", id); got != want {
		t.Errorf("QuestionText = %q, want %q", got, want)
	}
}
