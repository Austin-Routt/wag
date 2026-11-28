package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/game"
)

// U1, U3 and U5 (docs/c5-item-f.md): the server refuses a new game whose
// claim, or whose claim's opposite, is in the engine's undercutter
// namespace, takes a claim of another form, and names the change when it
// loads a game recorded with such a claim.
func TestU1U3U5TheUndercuttersNamespace(t *testing.T) {
	s, ts := newTestServerWith(t)
	for _, claim := range []string{"undercut(a11)", "undercut( a11 )", "undercut (a11)", "¬undercut(a11)", "¬¬undercut(a11)", "¬(undercut(a11))"} {
		req := NewGameRequest{Mode: "inquiry", Claim: claim, Seats: twoHumans}
		if _, err := req.config(s.contexts); !errors.Is(err, game.ErrReservedClaim) {
			t.Errorf("U1 %s: %v", claim, err)
		}
	}
	var out map[string]any
	req := NewGameRequest{Mode: "inquiry", Claim: "undercut(a11)", Seats: twoHumans}
	if code := call(t, ts, "POST", "/api/games", req, &out); code != http.StatusBadRequest ||
		!strings.Contains(out["error"].(string), "names the engine's undercutter of an argument") {
		t.Errorf("U1 at /api/games: %d %v", code, out)
	}
	for _, claim := range []string{"undercut_cost(a11)", "undercut(a11,b)", "undercut", "worth_finding_out_first(tax_meat,demand_falls)"} {
		req := NewGameRequest{Mode: "inquiry", Claim: claim, Seats: twoHumans}
		if _, err := req.config(s.contexts); err != nil {
			t.Errorf("U3 %s: %v", claim, err)
		}
	}
	// U5: a game recorded with the claim (A9.13's record, its claim changed
	// and its moves left out: the load stops at the claim).
	b, err := os.ReadFile("testdata/f1-before/game.export.json")
	if err != nil {
		t.Fatal(err)
	}
	var e Export
	if err := json.Unmarshal(b, &e); err != nil {
		t.Fatal(err)
	}
	e.Request.Claim, e.History = "undercut(a11)", nil
	for _, inspect := range []bool{false, true} {
		_, err := s.load(&e, inspect)
		if !errors.Is(err, game.ErrReservedClaim) || !strings.Contains(err.Error(), "a WAG build of commit f38d189 or earlier can") {
			t.Errorf("U5, inspect %v: %v", inspect, err)
		}
	}
}
