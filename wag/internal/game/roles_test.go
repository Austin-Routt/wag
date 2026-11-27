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

// bridgeGame is a game on the shipped data with walton.yml's modal variant of
// positive consequences and its bridge added, with their roles; refuse adds
// the bridge's question at its exception, which refuses Other (D095 (3)),
// and turns Other on.
func bridgeGame(t *testing.T, refuse bool) (*GameState, *carneades.Adapter, *nascq.NASCQ) {
	t.Helper()
	core := prevCore
	for k, tmpl := range map[string]string{
		"may_bring_about/2":         "The action %s may bring about %s.",
		"may_be_worth_performing/2": "Action %s may be worth performing for the sake of %s.",
		"matters_to_decision/2":     "The chance that action %s brings about %s bears on the decision.",
		"worth_finding_out_first/2": "It is worth finding out first whether action %s would bring about %s.",
	} {
		core = addTemplate(t, core, k, tmpl)
	}
	core = withScheme(t, core, yaml.MapSlice{{Key: "id", Value: "positive_consequences_modal"},
		{Key: "variables", Value: []string{"A", "G"}}, {Key: "premises", Value: []string{"may_bring_about(A,G)", "good(G)"}},
		{Key: "conclusions", Value: []string{"may_be_worth_performing(A,G)"}}})
	core = withScheme(t, core, yaml.MapSlice{{Key: "id", Value: "bridge_from_possible_good"},
		{Key: "variables", Value: []string{"A", "G"}},
		{Key: "premises", Value: []string{"may_be_worth_performing(A,G)", "matters_to_decision(A,G)"}},
		{Key: "exceptions", Value: []string{"worth_finding_out_first(A,G)"}},
		{Key: "conclusions", Value: []string{"should_be_performed(A)"}}})
	ad, err := carneades.NewAdapterFromYAML(core)
	if err != nil {
		t.Fatal(err)
	}
	var schemes map[string]json.RawMessage
	if err := json.Unmarshal(prevSchemes, &schemes); err != nil {
		t.Fatal(err)
	}
	vars := `{"A":{"type":"action"},"G":{"type":"event"}}`
	schemes["AS:consequences_modal"] = json.RawMessage(`{"carneades_id":"positive_consequences_modal","name":"Possible Positive Consequences",` +
		`"semantic_type":"action","base_weight":0.1,"role":"modal","weighs_as":"AS:consequences","variables":` + vars + `,` +
		`"premise_templates_en":["Doing {A} may bring about {G}","{G} is good"],"conclusion_template_en":"{A} may be worth doing for {G}"}`)
	schemes["AS:bridge_possible_good"] = json.RawMessage(`{"carneades_id":"bridge_from_possible_good","name":"From a Possible Good to the Action",` +
		`"semantic_type":"action","base_weight":0.2,"role":"bridge","weighs_as":"AS:consequences_modal",` +
		`"entry_premise":"may_be_worth_performing({A},{G})","variables":` + vars + `,` +
		`"premise_templates_en":["{A} may be worth doing for {G}","The chance matters"],"conclusion_template_en":"{A} should be done"}`)
	schemeBytes, err := json.Marshal(schemes)
	if err != nil {
		t.Fatal(err)
	}
	cqs, edges := prevCQs, prevEdges
	if refuse {
		var q map[string]json.RawMessage
		if err := json.Unmarshal(cqs, &q); err != nil {
			t.Fatal(err)
		}
		q["CQ:worth_finding_out_first"] = json.RawMessage(`{"name":"Finding Out First Question","semantic_type":"causality","refuse_other":true,
			"question_template_en":"Should finding out whether {A} brings about {G} come first?",
			"caes_mapping":{"positive_answer_type":"support","negative_answer_type":"undercutter","target":"argument"}}`)
		var e map[string][]json.RawMessage
		if err := json.Unmarshal(edges, &e); err != nil {
			t.Fatal(err)
		}
		e["ask_edges"] = append(e["ask_edges"], json.RawMessage(
			`{"cq":"CQ:worth_finding_out_first","scheme":"AS:bridge_possible_good","locus":"exception","pattern":"worth_finding_out_first({A},{G})"}`))
		if cqs, err = json.Marshal(q); err != nil {
			t.Fatal(err)
		}
		if edges, err = json.Marshal(e); err != nil {
			t.Fatal(err)
		}
	}
	n, err := nascq.LoadFromBytes(schemeBytes, cqs, edges)
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultGameConfig()
	cfg.MaxTurns = 0
	cfg.AllowOther = refuse
	return NewGame(cfg, ad, n, []string{"Alice", "Bob"}), ad, n
}

// A bridge's entry premise is not presumed: a modal argument supports it,
// before or after (D008, D094 (g)). Its other premise is presumed as any
// move's is; the bridge and its modal variant weigh as the categorical
// counterpart, with no multiplier. The two schemes are walton.yml's, added to
// the shipped data with roles of their own for the test.
func TestBridgeEntryPremiseIsNotPresumed(t *testing.T) {
	gs, ad, n := bridgeGame(t, false)
	cfg := gs.Config
	label := func(id string) caes.Label {
		if s := gs.Adapter.Graph.Statements[id]; s != nil {
			return s.Label
		}
		return caes.Undecided
	}
	const entry, other, concl = "may_be_worth_performing(carbon_tax,cleaner_air)", "matters_to_decision(carbon_tax,cleaner_air)", "should_be_performed(carbon_tax)"
	consW := n.Schemes["AS:consequences"].BaseWeight

	r, err := gs.ExecuteTurn(nascq.ProposedAction{Type: "support", SchemeID: "AS:bridge_possible_good",
		VariableBindings: map[string]string{"A": "carbon_tax", "G": "cleaner_air"}})
	if err != nil || !r.Accepted {
		t.Fatalf("the bridge: %v %+v", err, r)
	}
	if label(other) != caes.In {
		t.Errorf("the bridge's other premise is %v, not presumed", label(other))
	}
	if label(entry) == caes.In {
		t.Fatal("the entry premise was presumed")
	}
	if label(concl) == caes.In {
		t.Fatal("the bridge reached the recommendation without a modal argument")
	}
	if w := gs.MoveHistory[len(gs.MoveHistory)-1].Weight; w != consW {
		t.Errorf("the bridge weighs %v, not the consequence scheme's %v", w, consW)
	}
	// The weight shown before the move (the terminal player, the moderator)
	// is the weight the move takes.
	for _, id := range []string{"AS:bridge_possible_good", "AS:consequences_modal"} {
		if w := gs.MoveWeight(nascq.ProposedAction{SchemeID: id}); w != consW {
			t.Errorf("%s is shown at %v, not %v", id, w, consW)
		}
	}

	// Later, a modal argument for the entry premise: the bridge then holds.
	r, err = gs.ExecuteTurn(nascq.ProposedAction{Type: "support", SchemeID: "AS:consequences_modal", TargetStmtID: entry,
		VariableBindings: map[string]string{"A": "carbon_tax", "G": "cleaner_air"}})
	if err != nil || !r.Accepted {
		t.Fatalf("the modal argument: %v %+v", err, r)
	}
	if w := gs.MoveHistory[len(gs.MoveHistory)-1].Weight; w != consW {
		t.Errorf("the modal argument weighs %v, not the consequence scheme's %v", w, consW)
	}
	if label(entry) != caes.In || label(concl) != caes.In {
		t.Fatalf("with the modal argument: the entry premise %v, the recommendation %v", label(entry), label(concl))
	}
	if errs := gs.CheckInvariants(); len(errs) > 0 {
		t.Fatal(errs)
	}

	// Control: an ordinary scheme's premises are all presumed.
	gs2 := NewGame(cfg, ad, n, []string{"Alice", "Bob"})
	play(t, gs2, nascq.ProposedAction{Type: "support", SchemeID: "AS:consequences",
		VariableBindings: map[string]string{"A": "carbon_tax", "G": "cleaner_air"}})
	for _, id := range []string{"brings_about(carbon_tax,cleaner_air)", "good(cleaner_air)"} {
		if s := gs2.Adapter.Graph.Statements[id]; s == nil || s.Label != caes.In {
			t.Errorf("the consequence argument's premise %s is not presumed", id)
		}
	}
}

// Other cannot establish the bridges' exception by a move about it either
// (D095 (3)): the refusal is the question's, however the exception is
// argued. Arguing against the exception with Other stays open.
func TestOtherRefusedAtTheBridgesException(t *testing.T) {
	gs, _, _ := bridgeGame(t, true)
	play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:bridge_possible_good",
		VariableBindings: map[string]string{"A": "carbon_tax", "G": "cleaner_air"}})
	const exc = "worth_finding_out_first(carbon_tax,cleaner_air)"
	if gs.Adapter.Graph.Statements[exc] == nil {
		t.Fatalf("the bridge's exception is not on the board")
	}
	r, err := gs.ExecuteTurn(nascq.ProposedAction{Type: "support", SchemeID: nascq.OtherSchemeID, TargetStmtID: exc,
		VariableBindings: map[string]string{"P1": "it could be found out"}})
	if err != nil || r.Accepted || !strings.Contains(gs.LastRejection, `S5: `+exc+` is the exception CQ "CQ:worth_finding_out_first" asks about`) {
		t.Fatalf("Other for the exception: %v %+v %q", err, r, gs.LastRejection)
	}
	// Nor by a move of its own that names the exception (Other's text reuses
	// a statement of the board), nor by presuming it as a premise (the 5a
	// check's E-b): each would make it out without the question's answer.
	for _, b := range []map[string]string{
		{"P1": "it could be investigated", "C": exc},
		{"P1": exc, "C": "the tax can wait"},
	} {
		r, err := gs.ExecuteTurn(nascq.ProposedAction{Type: "support", SchemeID: nascq.OtherSchemeID, VariableBindings: b})
		if err != nil || r.Accepted || !strings.Contains(gs.LastRejection, `the exception CQ "CQ:worth_finding_out_first" asks about`) {
			t.Errorf("Other with %v: %v %+v %q", b, err, r, gs.LastRejection)
		}
	}
	if s := gs.Adapter.Graph.Statements[exc]; s == nil || s.Label == caes.In {
		t.Fatalf("the exception is %+v", s)
	}
	play(t, gs, nascq.ProposedAction{Type: "attack", SchemeID: nascq.OtherSchemeID, TargetStmtID: exc,
		VariableBindings: map[string]string{"P1": "the decision cannot wait"}})
}

// Under another audience's weights (ClaimUnder), a bridge and its modal
// variant weigh as their categorical counterpart, as they do in play.
func TestClaimUnderWeighsRolesAsTheirChain(t *testing.T) {
	gs, _, _ := bridgeGame(t, false)
	gs.Config.Objective = ObjectiveClaim
	const claim, entry = "should_be_performed(carbon_tax)", "may_be_worth_performing(carbon_tax,cleaner_air)"
	ac := map[string]string{"A": "carbon_tax", "G": "cleaner_air"}
	play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:bridge_possible_good", VariableBindings: ac})
	if gs.Claim != claim {
		t.Fatalf("the claim is %q", gs.Claim)
	}
	if r, err := gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"}); err != nil || !r.Accepted {
		t.Fatal(err)
	}
	play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:consequences_modal", TargetStmtID: entry, VariableBindings: ac})
	play(t, gs, nascq.ProposedAction{Type: "attack", SchemeID: "AS:negative_consequences", TargetStmtID: claim,
		VariableBindings: map[string]string{"G": "higher_prices"}})
	for w, want := range map[float64]string{0.9: "in", 0.3: "out"} {
		if got := gs.ClaimUnder(map[string]float64{"AS:consequences": w}); got != want {
			t.Errorf("with positive consequences at %v the claim is %s, want %s", w, got, want)
		}
	}
}
