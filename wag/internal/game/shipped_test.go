package game

import (
	"sort"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
)

// newDataGame starts a game on the data the game ships (package data).
func newDataGame(t *testing.T, players ...string) *GameState {
	t.Helper()
	ad, err := carneades.NewAdapterFromYAML(data.CoreYAML)
	if err != nil {
		t.Fatal(err)
	}
	n, err := nascq.LoadFromBytes(data.SchemesJSON, data.CQsJSON, data.EdgesJSON)
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultGameConfig()
	cfg.MaxTurns = 12
	return NewGame(cfg, ad, n, players)
}

// The data the game ships since step C5.2 (docs/c5-game.md): seventeen
// schemes and thirty-four questions, generated from the catalogue. The
// mechanics tests play the frozen pre-C5.2 data (prevdata_test.go); this
// test, the golden set on the new data and item F read package data.
func TestShippedDataLoading(t *testing.T) {
	ad, err := carneades.NewAdapterFromYAML(data.CoreYAML)
	if err != nil {
		t.Fatal(err)
	}
	n, err := nascq.LoadFromBytes(data.SchemesJSON, data.CQsJSON, data.EdgesJSON)
	if err != nil {
		t.Fatal(err)
	}
	if issues, err := LintData(data.CoreYAML, data.SchemesJSON, data.CQsJSON, data.EdgesJSON); err != nil || len(issues) > 0 {
		t.Fatalf("lint: %v %v", err, issues)
	}
	stats := n.Statistics()
	if stats.SchemeCount != 17 || stats.QuestionCount != 34 || stats.ThesisIndependentCount != 2 {
		t.Errorf("%d schemes, %d questions, %d thesis-independent; want 17, 34, 2", stats.SchemeCount, stats.QuestionCount, stats.ThesisIndependentCount)
	}
	if errs := n.ValidateBipartite(); len(errs) > 0 {
		t.Errorf("bipartite: %v", errs)
	}
	known := map[string]bool{}
	for _, id := range ad.SchemeIDs() {
		known[id] = true
	}
	roles := map[string][]string{}
	for id, s := range n.Schemes {
		if !known[s.CarneadesID] {
			t.Errorf("%s plays %s, which core.yml lacks", id, s.CarneadesID)
		}
		roles[s.Role] = append(roles[s.Role], id)
		asked := len(n.GetAskCQs(id))
		switch {
		case s.Role == nascq.RoleAnswer && asked > 0:
			t.Errorf("the answer scheme %s is asked %d questions", id, asked)
		case s.Role != nascq.RoleAnswer && asked == 0:
			// An answer scheme records D075's reasoned none for its
			// questions (its entry's summary); any other is asked some.
			t.Errorf("no critical question is asked of %s", id)
		}
	}
	for r, want := range map[string]string{
		"":               "AS:analogy AS:consequences AS:expert_opinion AS:negative_consequences AS:position_to_know AS:practical_reasoning AS:practical_reasoning_goal AS:sign AS:values",
		nascq.RoleAnswer: "AS:bad_outweighs_good AS:better_means AS:finding_out_first AS:good_outweighs_bad",
		nascq.RoleModal:  "AS:consequences_modal AS:negative_consequences_modal",
		nascq.RoleBridge: "AS:bridge_possible_good AS:bridge_possible_harm",
	} {
		sort.Strings(roles[r])
		if got := strings.Join(roles[r], " "); got != want {
			t.Errorf("role %q: %s, want %s", r, got, want)
		}
	}
	// Each answer scheme backs the one question whose formula it concludes.
	for cq, want := range map[string]string{
		"CQ:bad_outweighs": "AS:bad_outweighs_good", "CQ:good_outweighs": "AS:good_outweighs_bad",
		"CQ:other_means": "AS:better_means", "CQ:worth_finding_out_first": "AS:finding_out_first",
	} {
		var got []string
		for _, s := range n.GetBackupSchemes(cq) {
			if s.Role == nascq.RoleAnswer {
				got = append(got, s.ID)
			}
		}
		if len(got) != 1 || got[0] != want {
			t.Errorf("%s is answered by %v, want %s", cq, got, want)
		}
	}
	for id, q := range n.Questions {
		if q.AnswerEN == nil || q.AnswerEN.Against == "" || q.AnswerEN.For == "" {
			t.Errorf("%s has no answer words", id)
		}
		if q.RefuseOther != (id == "CQ:worth_finding_out_first") {
			t.Errorf("%s refuses Other: %v (D095 (3) names only CQ:worth_finding_out_first)", id, q.RefuseOther)
		}
	}
	if s := n.Schemes["AS:bridge_possible_good"]; s.EntryPremise != "may_be_worth_performing({A},{G})" || s.WeighsAs != "AS:consequences_modal" {
		t.Errorf("the good bridge: %+v", s)
	}
	if s := n.Schemes["AS:better_means"]; len(s.Distinct) != 1 || !s.Comparison {
		t.Errorf("better_means: distinct %v, comparison %v", s.Distinct, s.Comparison)
	}
}
