package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/nascq"
)

// The new golden set (step C5.2's step 7; Replay-2 of the external review of
// 26 November), on the data the game ships (package data), each file's
// first line pinning the data's fingerprint and the profile, so that a change
// of either is reported as such before any game is compared:
//
//   - the reference games' table, played with the default profile, which
//     every game made since step C5.2 has (TestShippedReferenceDigests);
//   - a scripted game for each new route of the data (TestShippedReferenceRoutes):
//     each answer scheme, the modal variants and their bridges, the new and
//     moved questions, Other and the questions asked alone, on item F's board
//     and profile; its line pins the game's move digest and its final state.
//
// The 33 games on the frozen data (TestReferenceDigests) stay as they were.
// WAG_UPDATE_REFERENCE=1 go test ./internal/game -run '^TestShippedReference'
// rewrites both files; read the diff before committing it.

// shippedFingerprint is the shipped data's fingerprint, as the server computes
// it (server.Data.Hash: the four files, LF line endings, each followed by a
// zero byte).
func shippedFingerprint() string {
	h := sha256.New()
	for _, b := range [][]byte{data.CoreYAML, data.SchemesJSON, data.CQsJSON, data.EdgesJSON} {
		h.Write(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// golden compares the lines got (a header, then one line a game) with the
// file, or writes them with WAG_UPDATE_REFERENCE set.
func golden(t *testing.T, path string, got []string) {
	t.Helper()
	if os.Getenv("WAG_UPDATE_REFERENCE") != "" {
		if err := os.WriteFile(path, []byte(strings.Join(got, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d reference games on the shipped data to %s", len(got)-1, path)
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (WAG_UPDATE_REFERENCE=1 writes it)", err)
	}
	want := strings.Split(strings.TrimRight(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n"), "\n")
	if want[0] != got[0] {
		t.Fatalf("the golden set was made with other data or another profile:\n got  %s\n want %s\n"+
			"(a change of the data's identity rewrites it: WAG_UPDATE_REFERENCE=1, and read the diff)", got[0], want[0])
	}
	if len(want) != len(got) {
		t.Fatalf("%d reference games stored, %d played", len(want)-1, len(got)-1)
	}
	differ := 0
	for i := 1; i < len(got); i++ {
		if got[i] != want[i] {
			differ++
			t.Errorf("game %d differs from the reference\n got  %s\n want %s", i, got[i], want[i])
		}
	}
	if differ == 0 {
		t.Logf("%d reference games on the shipped data reproduced", len(got)-1)
	}
}

func TestShippedReferenceDigests(t *testing.T) {
	prof, err := json.Marshal(DefaultProfile())
	if err != nil {
		t.Fatal(err)
	}
	got := []string{fmt.Sprintf("data %s profile %s", shippedFingerprint(), prof)}
	for _, g := range referenceGames() {
		got = append(got, g.playOn(t, data.CoreYAML, data.SchemesJSON, data.CQsJSON, data.EdgesJSON, DefaultProfile()))
	}
	golden(t, "testdata/reference-digests-shipped.txt", got)
}

// shippedRoutes are the scripted games of the new routes, each named after
// the item F case it plays (docs/c5-item-f.md) or the question it reaches.
func shippedRoutes() []struct {
	name string
	play func(g *fgame)
} {
	other := func(p string) map[string]string { return vb("P1", p) }
	return []struct {
		name string
		play func(g *fgame)
	}{
		{"A1.1 bad_outweighs_good at a consequence", func(g *fgame) {
			c1 := g.c1()
			g.play("Con", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", c1, vb("H", "food_costlier")))
		}},
		{"A1.2 one weighing for consequences and practical reasoning", func(g *fgame) {
			g.c1()
			pr := g.pr()
			g.play("Con", answer("attack", "CQ:bad_outweighs", "AS:bad_outweighs_good", pr, vb("H", "food_costlier")))
		}},
		{"A2.2 good_outweighs_bad for both negative arguments", func(g *fgame) {
			n1 := g.n1()
			g.nm()
			g.play("Pro", answer("attack", "CQ:good_outweighs", "AS:good_outweighs_bad", n1, vb("H", "emissions_fall")))
		}},
		{"A3.1 better_means, sooner", func(g *fgame) {
			pr := g.pr()
			g.play("Con", betterMeans(pr, vb("A2", "subsidy_plants"), "sooner"))
		}},
		{"A4.1 finding_out_first for both bridges", func(g *fgame) {
			g.mPlus("demand_falls")
			bp := g.bPlus("demand_falls")
			g.mMinus()
			g.bMinus()
			g.c1()
			g.play("Con", answer("attack", "CQ:worth_finding_out_first", "AS:finding_out_first", bp, vb("I", "pilot_study")))
		}},
		{"A4.4 an expert at the bridges' question, then CQ:expertise", func(g *fgame) {
			g.mPlus("demand_falls")
			bp := g.bPlus("demand_falls")
			exp := g.play("Con", answer("attack", "CQ:worth_finding_out_first", "AS:expert_opinion", bp, vb("W", "dr_lee", "D", "food_economics")))
			g.play("Pro", ask("CQ:expertise", exp))
		}},
		{"A4.11 a bridge before its modal argument", func(g *fgame) {
			g.bPlus("demand_falls")
			g.mPlus("demand_falls")
		}},
		{"A5.1 sign at CQ:side_effect", func(g *fgame) {
			c1 := g.c1()
			g.play("Con", answer("attack", "CQ:side_effect", "AS:sign", c1, vb("A", "food_banks_report_strain")))
		}},
		{"A5.2 sign at CQ:possibility", func(g *fgame) {
			pr := g.pr()
			g.play("Con", answer("attack", "CQ:possibility", "AS:sign", pr, vb("A", "the_budget_is_frozen")))
		}},
		{"A5.3 sign at CQ:goal_possible", func(g *fgame) {
			pg := g.pg()
			g.play("Con", answer("attack", "CQ:goal_possible", "AS:sign", pg, vb("A", "emissions_rose_last_year")))
		}},
		{"A6.1 consequences at CQ:opposite_consequences", func(g *fgame) {
			n1 := g.n1()
			g.play("Pro", answer("attack", "CQ:opposite_consequences", "AS:consequences", n1, vb("A", "tax_meat", "G", "emissions_fall")))
		}},
		{"A7.1 negative consequences at CQ:side_effect", func(g *fgame) {
			c1 := g.c1()
			g.play("Con", answer("attack", "CQ:side_effect", "AS:negative_consequences", c1, vb("A", "tax_meat", "G", "food_costlier")))
		}},
		{"B1 Other at CQ:too_slight", func(g *fgame) {
			c1 := g.c1()
			g.play("Con", answer("attack", "CQ:too_slight", nascq.OtherSchemeID, c1, other("the fall is tiny")))
		}},
		{"B2 too_slight on both sides", func(g *fgame) {
			c3 := g.c3()
			g.n3()
			g.play("Con", answer("attack", "CQ:too_slight", nascq.OtherSchemeID, c3, other("demand hardly moves")))
		}},
		{"B3 CQ:relevance asked of Other", func(g *fgame) {
			c1 := g.c1()
			o := g.play("Con", answer("attack", "CQ:too_slight", nascq.OtherSchemeID, c1, other("the fall is tiny")))
			g.play("Pro", ask(nascq.CQRelevance, o))
		}},
		{"B6 D074's undercutter, with independent support", func(g *fgame) {
			c1 := g.c1()
			g.play("Con", ask("CQ:likelihood", c1))
			wish := g.play("Pro", support(nascq.OtherSchemeID, vb("P1", "lower emissions would be good", "C", brings1)))
			g.play("Con", answer("attack", nascq.CQRelevance, nascq.OtherSchemeID, wish, other("desirability is no evidence")))
			g.play("Pro", about("support", "AS:expert_opinion", brings1, vb("W", "dr_lee", "D", "climate_science")))
		}},
		{"C2 CQ:likelihood and CQ:evidence answered by one argument", func(g *fgame) {
			c1 := g.c1()
			g.play("Con", ask("CQ:likelihood", c1))
			g.play("Con", ask("CQ:evidence", c1))
			g.play("Pro", about("support", "AS:expert_opinion", brings1, vb("W", "dr_lee", "D", "climate_science")))
		}},
		{"C3 CQ:possibility asked", func(g *fgame) {
			pr := g.pr()
			g.play("Con", ask("CQ:possibility", pr))
		}},
		{"C6 a tie at too_slight", func(g *fgame) {
			c1 := g.c1()
			g.play("Con", answer("attack", "CQ:too_slight", nascq.OtherSchemeID, c1, other("the fall is tiny")))
			g.play("Pro", about("attack", nascq.OtherSchemeID, slight1, other("the fall is large")))
		}},
		{"D2 a direct rebuttal by negative consequences", func(g *fgame) {
			g.c1()
			g.play("Con", about("attack", "AS:negative_consequences", shouldTax, vb("A", "tax_meat", "G", "food_costlier")))
		}},
		{"E3 an alias, and a new action declared", func(g *fgame) {
			g.play("Pro", support("AS:consequences", vb("A", "meat_tax", "G", "emissions_fall")))
			a := support("AS:negative_consequences", vb("A", "ban_meat_ads", "G", "ad_revenue_lost"))
			a.Declare = []nascq.ActionDecl{{ID: "ban_meat_ads", Actor: DefaultAgent, Aliases: []string{"meat_ad_ban"}}}
			g.play("Con", a)
		}},
		{"CQ:modal_likelihood and CQ:modal_evidence asked", func(g *fgame) {
			mp := g.mPlus("demand_falls")
			mm := g.mMinus()
			g.play("Con", ask("CQ:modal_likelihood", mp))
			g.play("Pro", ask("CQ:modal_evidence", mm))
		}},
		{"Other at a modal argument's too_slight and gain_discounted", func(g *fgame) {
			mp := g.mPlus("demand_falls")
			g.play("Con", answer("attack", "CQ:too_slight", nascq.OtherSchemeID, mp, other("demand would barely fall")))
			g.play("Con", answer("attack", "CQ:gain_discounted", nascq.OtherSchemeID, mp, other("the gain goes to those who need it least")))
		}},
		{"Other at CQ:harm_discounted", func(g *fgame) {
			n1 := g.n1()
			g.play("Pro", answer("attack", "CQ:harm_discounted", nascq.OtherSchemeID, n1, other("those who pay can bear it")))
		}},
		{"CQ:correlation asked of sign; Other at position to know's CQ:honesty", func(g *fgame) {
			s := g.play("Pro", support("AS:sign", vb("A", "smoke_seen", "B", "fire_nearby")))
			g.play("Con", ask("CQ:correlation", s))
			p := g.play("Pro", support("AS:position_to_know", vb("W", "farmer_joe", "D", "farming", "S", "prices_rise")))
			g.play("Con", answer("attack", "CQ:honesty", nascq.OtherSchemeID, p, other("he sells the meat")))
		}},
		{"Other at practical reasoning's other goals and best means", func(g *fgame) {
			pr := g.pr()
			g.play("Con", answer("attack", "CQ:other_goals", nascq.OtherSchemeID, pr, other("cheap food matters more")))
			g.play("Con", answer("attack", "CQ:best_means", nascq.OtherSchemeID, pr, other("a subsidy reaches the state sooner")))
		}},
		{"Other at practical reasoning with a goal's conflicting goals and intervening actions", func(g *fgame) {
			pg := g.pg()
			g.play("Con", answer("attack", "CQ:conflicting_goals", nascq.OtherSchemeID, pg, other("growth is a goal too")))
			g.play("Con", answer("attack", "CQ:intervening_actions", nascq.OtherSchemeID, pg, other("imports would rise first")))
		}},
	}
}

func TestShippedReferenceRoutes(t *testing.T) {
	prof, err := json.Marshal(itemfProfile())
	if err != nil {
		t.Fatal(err)
	}
	got := []string{fmt.Sprintf("data %s profile %s", shippedFingerprint(), prof)}
	for _, r := range shippedRoutes() {
		g := newFGame(t)
		r.play(g)
		state := sha256.Sum256([]byte(g.state()))
		got = append(got, fmt.Sprintf("%s: moves=%d digest=%s state=%s", r.name, len(g.gs.MoveHistory),
			g.gs.MoveDigest(), hex.EncodeToString(state[:])))
	}
	golden(t, "testdata/reference-routes-shipped.txt", got)
}
