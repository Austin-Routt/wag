package server

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/game"
	"gopkg.in/yaml.v2"
)

// A model player's move declares actions and records a comparison
// (D094 (b), D089); anything else in those fields is refused by name.
func TestReplyDeclaresActions(t *testing.T) {
	a, err := ParseReply(`{"type":"support","scheme_id":"AS:consequences","bindings":{"A":"tax_fish","G":"x"},
		"declare":[{"id":" tax_fish ","actor":"governments","aliases":["fish tax",""]},{"id":"switch_crops","actor":"farmers"}],"comparison":" Cheaper "}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Declare) != 2 || a.Declare[0].ID != "tax_fish" || a.Declare[0].Actor != "governments" || len(a.Declare[0].Aliases) != 1 ||
		a.Declare[0].Aliases[0] != "fish tax" || a.Declare[1].Actor != "farmers" || a.Comparison != "cheaper" {
		t.Fatalf("%+v", a)
	}
	for _, c := range []struct{ reply, want string }{
		{`{"type":"support","declare":"tax_fish"}`, `"declare" must be a list`},
		{`{"type":"support","declare":["tax_fish"]}`, "each action must be an object with its id and actor"},
		{`{"type":"support","declare":[3]}`, "each action must be an object with its id and actor"},
		{`{"type":"support","declare":[{"id":3}]}`, `an action's "id" must be text`},
		{`{"type":"support","declare":[{"id":"x","aliases":"y"}]}`, "aliases must be a list"},
		{`{"type":"support","declare":[{"id":"x","aliases":[1]}]}`, "each alias must be text"},
		{`{"type":"support","declare":[{"id":"x","owner":"y"}]}`, `not "owner"`},
		{`{"type":"support","comparison":1}`, `"comparison" must be text`},
	} {
		if _, err := ParseReply(c.reply); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", c.reply, err, c.want)
		}
	}
}

// A new game has a profile: the case record's, the one given, or the
// default; the request recorded with the game holds it, so a load replays
// with it. A claim about another actor's action is refused.
func TestNewGameProfile(t *testing.T) {
	cs, err := os.ReadFile("../../../docs/kialo-batch2/case.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewGameRequest{Mode: "classic", Case: cs}.withProfile()
	if err != nil || r.Profile == nil || r.Profile.Agent != "governments" || r.Case != nil {
		t.Fatalf("from the case record: %v %+v", err, r.Profile)
	}
	cfg, err := r.config(nil)
	if err != nil || cfg.Profile == nil || cfg.Profile.Agent != "governments" {
		t.Fatalf("the config: %v %+v", err, cfg.Profile)
	}
	d, err := NewGameRequest{Mode: "classic"}.withProfile()
	if err != nil || d.Profile == nil || d.Profile.Agent != game.DefaultAgent {
		t.Fatalf("the default: %v %+v", err, d.Profile)
	}
	if _, err := (NewGameRequest{Case: cs, Profile: game.DefaultProfile()}).withProfile(); err == nil {
		t.Error("a profile and a case record were both taken")
	}
	bad := NewGameRequest{Mode: "debate", Claim: "should_be_performed(x)", Profile: &game.Profile{Agent: "a",
		Actors: []game.ProfileActor{{ID: "a"}, {ID: "b"}}, Actions: []game.ProfileAction{{ID: "x", Actor: "b"}}, Situation: "now"}}
	if _, err := bad.config(nil); err == nil || !strings.Contains(err.Error(), "b's action, not the agent's") {
		t.Errorf("a claim about another actor's action: %v", err)
	}
	broken := NewGameRequest{Mode: "classic", Profile: &game.Profile{Agent: "nobody"}}
	if _, err := broken.config(nil); err == nil || !strings.Contains(err.Error(), "none of the actors") {
		t.Errorf("a profile whose agent is no actor: %v", err)
	}
	// A request recorded before step C5.2 (none) replays without the rules.
	var old NewGameRequest
	if err := json.Unmarshal([]byte(`{"mode":"classic"}`), &old); err != nil {
		t.Fatal(err)
	}
	if cfg, err := old.config(nil); err != nil || cfg.Profile != nil {
		t.Errorf("an old request: %v %+v", err, cfg.Profile)
	}
}

// The fixes of the independent check of plan version 5a
// (wag-workspace verify/c5/c52/CHECK-5a.md), each with the case that showed
// the fault.

// A discussion imported from Kialo is a game with a profile (E-e): its
// players declare the actions they argue about.
func TestKialoImportHasAProfile(t *testing.T) {
	ts := newTestServer(t)
	body := `{"id":"9","displayTitle":"Q?","thesis_claim_id":"9.1","claims":{"9.1":{"id":"9.1","text":"Yes."},"9.2":{"id":"9.2","text":"Because."}},
	 "edges":[{"location_id":"9.3","from_claim_id":"9.1","to_claim_id":"9.2","relation":"support","vote_vector":[0,0,0,0,1],"active":true}]}`
	var r struct{ State State }
	if code := call(t, ts, "POST", "/api/kialo/import", body, &r); code != 200 {
		t.Fatalf("import: %d", code)
	}
	if r.State.Profile == nil || r.State.Profile.Agent != game.DefaultAgent {
		t.Fatalf("the imported game's profile: %+v", r.State.Profile)
	}
	mr := post(t, ts, "/api/games/"+r.State.ID+"/move", map[string]interface{}{"action": map[string]interface{}{
		"type": "support", "scheme_id": "AS:consequences", "variable_bindings": map[string]string{"A": "anything at all", "G": "x"}}})
	if mr.Result["accepted"] == true || !strings.Contains(fmt.Sprint(mr.Result["error"]), "P1:") {
		t.Errorf("an undeclared action in an imported game: %v %v", mr.Result["accepted"], mr.Result["error"])
	}
}

// "case": null is no case record (N-g); a profile given without a situation
// has the present one (N-f).
func TestNewGameNullCaseAndSituation(t *testing.T) {
	ts := newTestServer(t)
	var st State
	if code := call(t, ts, "POST", "/api/games", `{"mode":"classic","case":null}`, &st); code != 200 || st.Profile == nil {
		t.Fatalf(`"case": null: %d %+v`, code, st.Profile)
	}
	st = State{}
	if code := call(t, ts, "POST", "/api/games", `{"mode":"classic","profile":{"agent":"a","actors":[{"id":"a"}]}}`, &st); code != 200 ||
		st.Profile == nil || st.Profile.Situation != game.DefaultProfile().Situation {
		t.Fatalf("a profile without a situation: %d %+v", code, st.Profile)
	}
}

// The preview validates a move as a turn does (W-d): an agent place left to
// the game, as the board leaves it, previews as the move plays.
func TestPreviewAppliesTheProfile(t *testing.T) {
	var schemes map[string]json.RawMessage
	if err := json.Unmarshal(data.SchemesJSON, &schemes); err != nil {
		t.Fatal(err)
	}
	schemes["AS:pr_agent"] = json.RawMessage(`{"carneades_id":"pr_agent","name":"Practical Reasoning of an Agent","semantic_type":"action",
		"base_weight":0.7,"variables":{"A":{"type":"action"},"G":{"type":"goal"},"P":{"type":"agent"}},
		"premise_templates_en":["{P} has the goal {G}","Doing {A} brings about {G}"],"conclusion_template_en":"{A} should be done"}`)
	var edges map[string][]json.RawMessage
	if err := json.Unmarshal(data.EdgesJSON, &edges); err != nil {
		t.Fatal(err)
	}
	edges["ask_edges"] = append(edges["ask_edges"], json.RawMessage(
		`{"cq":"CQ:likelihood","scheme":"AS:pr_agent","locus":"premise","pattern":"brings_about({A},{G})"}`))
	sb, _ := json.Marshal(schemes)
	eb, _ := json.Marshal(edges)
	var core yaml.MapSlice
	if err := yaml.Unmarshal(data.CoreYAML, &core); err != nil {
		t.Fatal(err)
	}
	for i, kv := range core {
		switch kv.Key {
		case "argument_schemes":
			core[i].Value = append(kv.Value.([]interface{}), yaml.MapSlice{{Key: "id", Value: "pr_agent"},
				{Key: "variables", Value: []string{"A", "G", "P"}},
				{Key: "premises", Value: []string{"goal_of_agent(P,G)", "brings_about(A,G)"}},
				{Key: "conclusions", Value: []string{"should_be_performed(A)"}}})
		case "language":
			core[i].Value = append(kv.Value.(yaml.MapSlice), yaml.MapItem{Key: "goal_of_agent/2", Value: "%s has the goal %s."})
		}
	}
	yb, err := yaml.Marshal(core)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(Data{YAML: yb, Schemes: sb, CQs: data.CQsJSON, Edges: eb, Contexts: data.ContextsJSON}, "")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	var st State
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "debate", Claim: "should_be_performed(tax_meat)",
		Seats: []Seat{{Name: "Ann", Kind: "human"}, {Name: "Bob", Kind: "human"}}}, &st); code != 200 {
		t.Fatalf("new game: %d", code)
	}
	move := map[string]interface{}{"type": "support", "scheme_id": "AS:pr_agent",
		"variable_bindings": map[string]string{"A": "tax_meat", "G": "lower_emissions"}}
	var pv Preview
	call(t, ts, "POST", "/api/games/"+st.ID+"/preview", map[string]interface{}{"action": move}, &pv)
	if pv.Error != "" {
		t.Fatalf("the preview refuses a move the game accepts: %q", pv.Error)
	}
	if mr := post(t, ts, "/api/games/"+st.ID+"/move", map[string]interface{}{"action": move}); mr.Result["accepted"] != true {
		t.Fatalf("the move: %v", mr.Result["error"])
	}
}

// A model player's prompt names the other actors' declared actions too
// (N-e), and a considered move keeps the parts of it that read (N-h).
func TestPromptOtherActorsAndAlternatives(t *testing.T) {
	s, _ := newTestServerWith(t)
	r := NewGameRequest{Mode: "classic", Seats: []Seat{{Name: "Ann", Kind: "human"}, {Name: "Bob", Kind: "human"}},
		Profile: &game.Profile{Agent: "governments", Actors: []game.ProfileActor{{ID: "governments"}, {ID: "farmers"}},
			Actions: []game.ProfileAction{{ID: "tax_meat", Actor: "governments"}, {ID: "switch_crops", Actor: "farmers"}}, Situation: "now"}}
	cfg, err := r.config(s.contexts)
	if err != nil {
		t.Fatal(err)
	}
	gs, err := s.newState(cfg, []string{"Ann", "Bob"})
	if err != nil {
		t.Fatal(err)
	}
	ss := s.add(gs, r.Seats, "t", "new", 1)
	if p := ss.Prompt(); !strings.Contains(p, "Other actors' actions, which a recommendation to the agent does not argue: switch_crops (farmers's).") {
		t.Errorf("the prompt does not list other actors' actions:\n%s", p)
	}
	alts := alternatives([]interface{}{map[string]interface{}{"why_not": "w", "action": map[string]interface{}{
		"type": "support", "scheme_id": "better_means", "comparison": "Cheaper", "sources": []interface{}{"a study"},
		"declare": "not a list"}}})
	if len(alts) != 1 || alts[0].Action.SchemeID != "AS:better_means" || alts[0].Action.Comparison != "cheaper" ||
		len(alts[0].Action.Evidence) != 1 {
		t.Errorf("a considered move with a declaration that does not read: %+v", alts)
	}
}
