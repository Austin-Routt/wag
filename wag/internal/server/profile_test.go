package server

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/game"
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
