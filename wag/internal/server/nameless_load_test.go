package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"
)

// A game recorded before 2026-11-27 with a value or a name that names
// nothing no longer replays. The loader names the build that replays it for
// each such refusal (S2, P1 on a value, P1 on a declaration, the profile's
// check), and for no other (the independent check of step 8, W-3 and N-1).

// exportOf plays the moves, each accepted, in a new game of the request, and
// returns the game's export.
func exportOf(t *testing.T, s *Server, req NewGameRequest, moves ...nascq.ProposedAction) *Export {
	t.Helper()
	cfg, err := req.config(s.contexts)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(req.Seats))
	for i, seat := range req.Seats {
		names[i] = seat.Name
	}
	gs, err := s.newState(cfg, names)
	if err != nil {
		t.Fatal(err)
	}
	ss := s.add(gs, req.Seats, "nameless", "new", 42)
	ss.Request = &req
	for _, m := range moves {
		r, err := gs.ExecuteTurn(m)
		if err != nil || !r.Accepted {
			t.Fatalf("%+v: %+v %v", m, r, err)
		}
	}
	b, err := json.Marshal(ss.export(s.contexts))
	if err != nil {
		t.Fatal(err)
	}
	var e Export
	if err := json.Unmarshal(b, &e); err != nil {
		t.Fatal(err)
	}
	return &e
}

func taxMeat() nascq.ProposedAction {
	return nascq.ProposedAction{Type: "support", SchemeID: "AS:consequences",
		VariableBindings: map[string]string{"A": "tax meat", "G": "lower emissions"},
		Declare:          []nascq.ActionDecl{{ID: "tax meat", Actor: game.DefaultAgent}}}
}

func TestNamelessNamesOfAnOldRecord(t *testing.T) {
	const remedy = "a WAG build of 2026-11-26 or earlier can"
	debate := NewGameRequest{Mode: "debate", Seats: twoHumans, Profile: game.DefaultProfile()}
	for _, c := range []struct {
		name  string
		edit  func(*Export)
		names string // what the message names; "" for no remedy
	}{
		{"an action value ?, P1", func(e *Export) {
			m := &e.History[0].Action
			m.VariableBindings["A"], m.Declare = "?", nil
		}, `A = "?"`},
		{"a declared alias, P1", func(e *Export) {
			e.History[0].Action.Declare[0].Aliases = []string{"—"}
		}, "the declaration of \"tax meat\" names \"—\""},
		{"a profile's alias", func(e *Export) {
			p := game.DefaultProfile()
			p.Actions = []game.ProfileAction{{ID: "ban ads", Actor: game.DefaultAgent, Aliases: []string{"—"}}}
			e.Request.Profile = p
		}, "the action \"ban ads\" has a name, \"—\", that names nothing"},
		{"a value that names nothing beside another refusal", func(e *Export) {
			m := &e.History[0].Action
			m.VariableBindings["A"], m.VariableBindings["Z"], m.Declare = "ban meat", "", nil
		}, ""},
	} {
		s, ts := newTestServerWith(t)
		e := exportOf(t, s, debate, taxMeat())
		c.edit(e)
		code, out := loadHTTP(t, ts, e, "")
		msg, _ := out["error"].(string)
		switch {
		case code == http.StatusOK:
			t.Errorf("%s: loaded", c.name)
		case c.names == "" && strings.Contains(msg, remedy):
			t.Errorf("%s: the remedy is named for another refusal: %q", c.name, msg)
		case c.names != "" && (!strings.Contains(msg, remedy) || !strings.Contains(msg, c.names)):
			t.Errorf("%s: %q, want the remedy and %s", c.name, msg, c.names)
		}
	}
}
