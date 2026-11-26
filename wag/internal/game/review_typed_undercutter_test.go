package game

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
	"gopkg.in/yaml.v2"
)

// This deliberately synthetic scheme tests the existing exception-locus
// path. It is not a proposed Walton record, a source approval, or support
// for the planned C001 direct inference-locus/argument-instance extension.
func TestReviewTypedExceptionUndercutterFitsExactTarget(t *testing.T) {
	const schemeID = "AS:test_typed_alternative"
	makeGame := func(t *testing.T, conclusion, bridge string) (*GameState, string) {
		t.Helper()
		core := addTemplate(t, prevCore, "test_comparison/3", "Synthetic comparison: %s against %s for %s.")
		core = withScheme(t, core, yaml.MapSlice{{Key: "id", Value: "test_typed_alternative"},
			{Key: "variables", Value: []string{"B", "A", "G"}}, {Key: "premises", Value: []string{"test_comparison(B,A,G)"}},
			{Key: "conclusions", Value: []string{conclusion}}})
		ad, err := carneades.NewAdapterFromYAML(core)
		if err != nil {
			t.Fatal(err)
		}
		var schemes map[string]json.RawMessage
		if err := json.Unmarshal(prevSchemes, &schemes); err != nil {
			t.Fatal(err)
		}
		schemes[schemeID] = json.RawMessage(`{"carneades_id":"test_typed_alternative","name":"Synthetic typed comparison (test only)","semantic_type":"proposition","base_weight":0.9,"variables":{"B":{"type":"action"},"A":{"type":"action"},"G":{"type":"goal"}},"premise_templates_en":["Synthetic comparison of {B} against {A} for {G}"],"conclusion_template_en":"An alternative achieves {G} more effectively than {A}"}`)
		schemeBytes, err := json.Marshal(schemes)
		if err != nil {
			t.Fatal(err)
		}
		var edges map[string][]map[string]string
		if err := json.Unmarshal(prevEdges, &edges); err != nil {
			t.Fatal(err)
		}
		edge := map[string]string{"scheme": schemeID, "cq": "CQ:other_means"}
		if bridge != "" {
			edge["bridge"] = bridge
		}
		edges["backup_edges"] = append(edges["backup_edges"], edge)
		edgeBytes, err := json.Marshal(edges)
		if err != nil {
			t.Fatal(err)
		}
		n, err := nascq.LoadFromBytes(schemeBytes, prevCQs, edgeBytes)
		if err != nil {
			t.Fatal(err)
		}
		cfg := DefaultGameConfig()
		cfg.MaxTurns = 0
		gs := NewGame(cfg, ad, n, []string{"Alice", "Bob"})
		return gs, play(t, gs, prSupportAction())
	}

	t.Run("exact exception is offered and defeats target", func(t *testing.T) {
		gs, target := makeGame(t, "realize_more_effectively(G,A)", "")
		offered := false
		for _, a := range gs.NASCQ.EnumerateActions(gs.Bridge(), []string{target}) {
			if a.Type == "attack" && a.SchemeID == schemeID && a.AnswersCQID == "CQ:other_means" {
				offered = true
				if a.Fixed["A"] != "carbon_tax" || a.Fixed["G"] != "reduce_emissions" {
					t.Fatalf("answer did not fix the target's exact action and goal: %v", a.Fixed)
				}
			}
		}
		if !offered {
			t.Fatal("a fitting typed exception answer was not offered")
		}
		id := play(t, gs, answer("attack", "CQ:other_means", schemeID, target, map[string]string{"B": "insulate homes"}))
		arg, attacked := gs.Adapter.Graph.Arguments[id], gs.Adapter.Graph.Arguments[target]
		if arg.Conclusion.Id != "realize_more_effectively(reduce_emissions,carbon_tax)" || arg.Conclusion.Label != caes.In {
			t.Fatalf("typed answer lost its exact exception conclusion: %+v", arg.Conclusion)
		}
		if len(arg.Premises) != 1 || arg.Premises[0].Stmt.Id != "test_comparison(insulate_homes,carbon_tax,reduce_emissions)" {
			t.Fatalf("typed answer premises are not about the target: %+v", arg.Premises)
		}
		if arg.Scheme == nil || arg.Scheme.Id != "test_typed_alternative" || arg.Metadata["nascq_scheme_id"] != schemeID || !gs.NASCQ.OwnConclusion(gs.Bridge(), id, schemeID) {
			t.Fatal("answer did not preserve its own typed scheme and conclusion")
		}
		if attacked.Undercutter == nil || attacked.Undercutter.Label != caes.In || attacked.Weight != 0 || attacked.Conclusion.Label != caes.Out {
			t.Fatalf("the established exception did not undercut its target: weight=%v conclusion=%v undercutter=%+v", attacked.Weight, attacked.Conclusion.Label, attacked.Undercutter)
		}
		last := gs.MoveHistory[len(gs.MoveHistory)-1]
		if last.Action.AttackType != "undercutter" || last.Bridge != "" {
			t.Fatalf("answer used a prose bridge or lost its undercutter classification: %+v", last)
		}
	})

	for _, tc := range []struct {
		name, conclusion, bridge string
		bindings                 map[string]string
	}{
		{"different action is refused", "realize_more_effectively(G,A)", "", map[string]string{"B": "insulate homes", "A": "plant trees"}},
		{"different goal is refused", "realize_more_effectively(G,A)", "", map[string]string{"B": "insulate homes", "G": "reduce costs"}},
		{"prose cannot bridge a different conclusion", "should_be_performed(B)", "the other action should be done, so it must be better", map[string]string{"B": "insulate homes", "A": "carbon_tax", "G": "reduce_emissions"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gs, target := makeGame(t, tc.conclusion, tc.bridge)
			before := gs.Adapter.Fingerprint()
			res, err := gs.ExecuteTurn(answer("attack", "CQ:other_means", schemeID, target, tc.bindings))
			if err != nil || res.Accepted || !strings.Contains(res.Error, "S9") {
				t.Fatalf("mismatched typed answer was not refused by S9: %+v %v", res, err)
			}
			if before != gs.Adapter.Fingerprint() || gs.Adapter.Graph.Arguments[target].Conclusion.Label != caes.In {
				t.Fatal("a refused typed answer changed the graph or defeated its target")
			}
			if tc.bridge != "" {
				for _, a := range gs.NASCQ.EnumerateActions(gs.Bridge(), []string{target}) {
					if a.SchemeID == schemeID && a.AnswersCQID == "CQ:other_means" {
						t.Fatal("a prose bridge made an incompatible answer available")
					}
				}
			}
		})
	}
}
