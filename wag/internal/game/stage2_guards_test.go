package game

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
)

// A positive consequence of another action does not establish that it
// realizes the attacked argument's goal more effectively. A prose bridge
// cannot supply that missing comparative inference (D075).
func TestStage2BridgeDoesNotWaiveAnswerFit(t *testing.T) {
	edges := withBackupEdge(t, prevEdges, `{"scheme":"AS:consequences","cq":"CQ:other_means","bridge":"the alternate action has a good consequence"}`)
	gs := newShippedGame(t, edges, "Alice", "Bob")
	gs.Config.MaxTurns = 0
	arg := play(t, gs, prSupportAction())
	a := answer("attack", "CQ:other_means", "AS:consequences", arg,
		map[string]string{"A": "subsidize insulation", "G": "unrelated benefit"})
	r, err := gs.ExecuteTurn(a)
	if err != nil || r.Accepted || !strings.Contains(r.Error, "S9") {
		t.Fatalf("unwarranted answer accepted through bridge: %+v %v", r, err)
	}
	for _, offered := range gs.NASCQ.EnumerateActions(gs.Bridge(), []string{arg}) {
		if offered.SchemeID == "AS:consequences" && offered.AnswersCQID == "CQ:other_means" {
			t.Fatal("menu offers answer solely on a prose bridge")
		}
	}
}

func TestStage2LintRefusesBridgeAndSharedEngineIdentity(t *testing.T) {
	var schemes map[string]map[string]interface{}
	if err := json.Unmarshal(prevSchemes, &schemes); err != nil {
		t.Fatal(err)
	}
	copyBytes, _ := json.Marshal(schemes["AS:sign"])
	var alias map[string]interface{}
	json.Unmarshal(copyBytes, &alias)
	schemes["AS:sign_alias"] = alias
	s, _ := json.Marshal(schemes)
	edges := withBackupEdge(t, prevEdges, `{"scheme":"AS:consequences","cq":"CQ:other_means","bridge":"untyped explanation"}`)
	issues, err := LintData(prevCore, s, prevCQs, edges)
	if err != nil {
		t.Fatal(err)
	}
	bridge, identity := false, false
	for _, issue := range issues {
		if !issue.Error {
			continue
		}
		bridge = bridge || strings.Contains(issue.Msg, "free-text bridge")
		identity = identity || strings.Contains(issue.Msg, "carneades_id") && strings.Contains(issue.Msg, "AS:sign_alias")
	}
	if !bridge || !identity {
		t.Fatalf("missing errors bridge=%v identity=%v: %v", bridge, identity, issues)
	}
	n, err := nascq.LoadFromBytes(s, prevCQs, edges)
	if err != nil {
		t.Fatal(err)
	}
	if got := n.FindNASCQSchemeID("sign"); got != "" {
		t.Fatalf("ambiguous engine identity silently chooses %s", got)
	}
}

func TestStage2RecordsExactGameSchemeOnMainArguments(t *testing.T) {
	gs := newShippedGame(t, prevEdges, "Alice", "Bob")
	act := prSupportAction()
	id := play(t, gs, act)
	arg := gs.Adapter.Graph.Arguments[id]
	if got := arg.Metadata["nascq_scheme_id"]; got != act.SchemeID {
		t.Fatalf("main argument lost game scheme: got %v want %s", got, act.SchemeID)
	}
	for _, a := range gs.Adapter.Graph.Arguments {
		if carneades.ArgumentRole(a) != carneades.RoleMain && a.Metadata["nascq_scheme_id"] != nil {
			t.Errorf("auxiliary role %s falsely attributed a played scheme", carneades.ArgumentRole(a))
		}
	}
}

func TestStage2ViewPreservesRecordedSchemeIdentity(t *testing.T) {
	gs := newShippedGame(t, prevEdges, "Alice", "Bob")
	id := play(t, gs, prSupportAction())
	// An old profile identity must survive even if the currently loaded
	// NASCQ no longer contains that identity. Do not silently relabel it.
	gs.Adapter.Graph.Arguments[id].Metadata["nascq_scheme_id"] = "AS:prior_profile_practical"
	for _, arg := range gs.BuildView().Arguments {
		if arg.ID == id {
			if arg.Scheme != "AS:prior_profile_practical" {
				t.Fatalf("history relabelled as %s", arg.Scheme)
			}
			return
		}
	}
	t.Fatal("argument absent from view")
}

// A displayed historical identity must not silently inherit the current
// profile's questions. Also reject a recorded ID for a different engine
// scheme; a matching string alone is not a compatible argument identity.
func TestStage2HistoricalIdentityCannotAcquireCurrentCQs(t *testing.T) {
	gs := newShippedGame(t, prevEdges, "Alice", "Bob")
	id := play(t, gs, prSupportAction())
	arg := gs.Adapter.Graph.Arguments[id]
	for _, incompatible := range []string{"AS:prior_profile_practical", "AS:sign"} {
		arg.Metadata["nascq_scheme_id"] = incompatible
		for _, offered := range gs.NASCQ.EnumerateActions(gs.Bridge(), []string{id}) {
			if offered.TargetArgID == id && offered.AnswersCQID != "" {
				t.Fatalf("%s inherited current CQ %s", incompatible, offered.AnswersCQID)
			}
		}
		_, err := gs.NASCQ.Validate(nascq.ProposedAction{Type: "question", TargetArgID: id,
			AnswersCQID: "CQ:other_goals"}, gs.Bridge())
		if err == nil || !strings.Contains(err.Error(), "S3") {
			t.Fatalf("incompatible recorded identity allowed current CQ: %v", err)
		}
	}
	delete(arg.Metadata, "nascq_scheme_id")
	if got := gs.CurrentSchemeID(arg); got != "AS:practical_reasoning" {
		t.Fatalf("unambiguous legacy identity lost: %q", got)
	}
}

// Historical games may already contain an answer whose engine scheme does
// not conclude that answer. Preserve the OwnConclusion protection when
// reading those games, even though new prose-bridge moves are now refused.
func TestStage2LegacyBridgedConclusionDoesNotAcquireSchemeCQ(t *testing.T) {
	gs := newShippedGame(t, prevEdges, "Alice", "Bob")
	gs.Config.MaxTurns = 0
	id := play(t, gs, prSupportAction())
	ans := play(t, gs, answer("attack", "CQ:other_means", "AS:position_to_know", id,
		map[string]string{"W": "an energy auditor", "D": "home insulation"}))
	arg := gs.Adapter.Graph.Arguments[ans]
	// This is a legacy-shaped input fixture, not a newly permitted move.
	arg.Scheme = &caes.Scheme{Id: gs.NASCQ.GetScheme("AS:consequences").CarneadesID}
	delete(arg.Metadata, "nascq_scheme_id")
	for _, offered := range gs.NASCQ.EnumerateActions(gs.Bridge(), []string{ans}) {
		if offered.TargetArgID == ans && offered.AnswersCQID == "CQ:side_effect" {
			t.Fatal("legacy answer inherited a CQ about a conclusion it never made")
		}
	}
	_, err := gs.NASCQ.Validate(nascq.ProposedAction{Type: "question", TargetArgID: ans,
		AnswersCQID: "CQ:side_effect"}, gs.Bridge())
	if err == nil || !strings.Contains(err.Error(), "does not apply") {
		t.Fatalf("legacy conclusion bypassed OwnConclusion: %v", err)
	}
}
