package server

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/kialo"
	"github.com/yourorg/wag/internal/nascq"
)

func reviewExport(t *testing.T, s *Server, req NewGameRequest) *Export {
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
	ss := s.add(gs, req.Seats, "identity regression", "new", 42)
	ss.Request = &req
	_, err = gs.ExecuteTurn(nascq.ProposedAction{Type: "support", SchemeID: "AS:analogy", VariableBindings: map[string]string{"C": "case one", "A": "claim one"}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(ss.export(s.contexts))
	if err != nil {
		t.Fatal(err)
	}
	var e Export
	if err = json.Unmarshal(b, &e); err != nil {
		t.Fatal(err)
	}
	return &e
}

// Successful move execution must not hide a known scheme-data mismatch.
func TestReviewLoadRejectsKnownDataMismatch(t *testing.T) {
	s, _ := newTestServerWith(t)
	e := reviewExport(t, s, NewGameRequest{Mode: "classic", Seats: []Seat{{Name: "Ann", Kind: "human"}, {Name: "Bob", Kind: "human"}}})
	e.DataHash = "different-recorded-scheme-data"
	before := len(s.sessions)
	_, err := s.load(e, false)
	if err == nil || !strings.Contains(err.Error(), "scheme data") {
		t.Fatalf("known hash mismatch was not refused: %v", err)
	}
	if len(s.sessions) != before {
		t.Fatal("failed replay registered a session")
	}
}

// Data.Hash intentionally covers scheme files, not a mutable context preset.
// The export records the weights in effect, and the game is replayed with
// them, as it was played, whether the preset lost a weight, changed one or
// is gone (revised: the review's patch refused the load, which kept the
// experiment only by asking for the old preset).
func TestReviewLoadKeepsRecordedContextWeights(t *testing.T) {
	changed := map[string]map[string]float64{"fewer weights": {"AS:analogy": 0.8}, "one value": nil, "no preset": nil}
	for name, preset := range changed {
		t.Run(name, func(t *testing.T) {
			s, _ := newTestServerWith(t)
			e := reviewExport(t, s, NewGameRequest{Mode: "classic", Context: "scientific", Seats: []Seat{{Name: "Ann", Kind: "human"}, {Name: "Bob", Kind: "human"}}})
			played := e.Config.Weights
			if played["AS:analogy"] != 0.45 || len(played) != 10 {
				t.Fatalf("fixture: the scientific preset is %v", played)
			}
			switch name {
			case "one value":
				preset = maps.Clone(played)
				preset["AS:analogy"] = 0.8
				fallthrough
			case "fewer weights":
				s.AddContexts([]game.Context{{ID: "scientific", Weights: preset}})
			case "no preset":
				s.contexts = slices.DeleteFunc(slices.Clone(s.contexts), func(c game.Context) bool { return c.ID == "scientific" })
			}
			if e.DataHash != s.data.Hash() {
				t.Fatal("fixture must isolate un-hashed context drift")
			}
			ss, err := s.load(e, false)
			if err != nil {
				t.Fatal(err)
			}
			if c := ss.GS.Config; c.Context != "scientific" || !maps.Equal(c.Weights, played) {
				t.Fatalf("replayed with %s %v, not the recorded weights", c.Context, c.Weights)
			}
			if w := ss.GS.NASCQ.Schemes["AS:analogy"].BaseWeight; w != 0.45 {
				t.Fatalf("AS:analogy weighs %v in the loaded game", w)
			}
			if ss.Request.Context != "scientific" || ss.inspectedWith != "" {
				t.Fatalf("request %+v, read-only %q", ss.Request, ss.inspectedWith)
			}
		})
	}
}

func TestReviewLoadSameContextAndLegacyHashStillWork(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		s, _ := newTestServerWith(t)
		e := reviewExport(t, s, NewGameRequest{Mode: "classic", Context: "scientific", Weights: map[string]float64{"AS:analogy": 0.6}, Seats: []Seat{{Name: "Ann", Kind: "human"}, {Name: "Bob", Kind: "human"}}})
		if legacy {
			e.DataHash = ""
		}
		ss, err := s.load(e, false)
		if err != nil {
			t.Fatalf("legacy=%v: %v", legacy, err)
		}
		if ss.GS.Config.Weights["AS:analogy"] != 0.6 {
			t.Fatal("recorded override changed")
		}
	}
}

// A display name must not claim the reserved identity used by the history
// format for moderator decisions.
func TestReviewPlayerNamedModeratorReplaysAsPlayer(t *testing.T) {
	s, _ := newTestServerWith(t)
	e := reviewExport(t, s, NewGameRequest{Mode: "classic", Seats: []Seat{{Name: game.ModeratorID, Kind: "human"}, {Name: "Bob", Kind: "human"}}})
	if len(e.History) != 1 {
		t.Fatalf("fixture has %d moves", len(e.History))
	}
	if e.History[0].PlayerID == game.ModeratorID {
		t.Error("ordinary player received the reserved moderator identity")
	}
	if _, err := s.load(e, false); err != nil {
		t.Fatalf("player display name broke round trip: %v", err)
	}
}

func TestReviewKialoSourceSurvivesRepeatedExportLoad(t *testing.T) {
	s, _ := newTestServerWith(t)
	s.kialoDir = t.TempDir()
	graph := []byte(`{"id":"9","displayTitle":"Synthetic discussion","thesis_claim_id":"9.1","claims":{"9.1":{"id":"9.1","text":"A claim"},"9.2":{"id":"9.2","text":"A reason"}},"edges":[{"location_id":"9.3","from_claim_id":"9.1","to_claim_id":"9.2","relation":"support","active":true}]}`)
	if err := os.MkdirAll(filepath.Join(s.kialoDir, "9"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.kialoDir, "9", "graph.json"), graph, 0600); err != nil {
		t.Fatal(err)
	}
	gs, rep, err := s.kialoGame(graph, 0, kialo.KeysByClaim)
	if err != nil {
		t.Fatal(err)
	}
	ss := s.add(gs, []Seat{{Name: kialo.ProPlayer, Kind: "human"}, {Name: kialo.ConPlayer, Kind: "human"}}, "synthetic", "kialo:9", 42)
	ss.Import = rep
	wantArgs := len(gs.Adapter.Graph.Arguments)
	if wantArgs == 0 {
		t.Fatal("fixture must contain imported arguments")
	}
	for generation := 1; generation <= 2; generation++ {
		b, err := json.Marshal(ss.export(s.contexts))
		if err != nil {
			t.Fatal(err)
		}
		var e Export
		if err := json.Unmarshal(b, &e); err != nil {
			t.Fatal(err)
		}
		ss, err = s.load(&e, false)
		if err != nil {
			t.Fatalf("generation %d: %v", generation, err)
		}
		if ss.Source != "kialo:9" {
			t.Errorf("generation %d lost source: %q", generation, ss.Source)
		}
		if len(ss.GS.Adapter.Graph.Arguments) != wantArgs {
			t.Errorf("generation %d changed imported argument count", generation)
		}
	}
}
