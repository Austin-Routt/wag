package catalogue

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"
	"gopkg.in/yaml.v2"
)

// frozenPreC5 is the game's data before step C5's switch (wag/data/frozen).
const frozenPreC5 = "wag/data/frozen/e7141ec8dea379bc"

// committedGame generates the game's files from the committed catalogue.
func committedGame(t *testing.T) (map[string]*Entry, *GameData, *GameGen) {
	t.Helper()
	dir := filepath.Join(root, filepath.FromSlash(Dir))
	entries, bad, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(bad) > 0 {
		t.Fatalf("entries that do not read: %v", bad)
	}
	d, err := ReadGameDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out, err := GenerateGame(entries, d, sources(t).GameSources())
	if err != nil {
		t.Fatal(err)
	}
	return entries, d, out
}

// sameGameFiles compares generated files with files on disk as trees: the
// JSON files' objects in their order; core.yml's mappings in any order (its
// readers read them into structures, and an entry keeps a scheme's formulas
// in core.yml's usual order); lists, among them the schemes and the
// formulas, in theirs.
func sameGameFiles(t *testing.T, files map[string][]byte, dir string) []string {
	t.Helper()
	var diffs []string
	for _, p := range GamePaths {
		want, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(p)))
		if err != nil {
			t.Fatal(err)
		}
		var a, b interface{}
		if strings.HasSuffix(p, ".json") {
			if a, err = decodeOrderedJSON(files[p]); err != nil {
				t.Fatalf("generated %s: %v", p, err)
			}
			if b, err = decodeOrderedJSON(want); err != nil {
				t.Fatal(err)
			}
		} else {
			var x, y yaml.MapSlice
			if err := yaml.Unmarshal(files[p], &x); err != nil {
				t.Fatalf("generated %s: %v", p, err)
			}
			if err := yaml.Unmarshal(want, &y); err != nil {
				t.Fatal(err)
			}
			a, b = plain(t, x), plain(t, y)
		}
		if !reflect.DeepEqual(a, b) {
			diffs = append(diffs, p)
		}
	}
	return diffs
}

// The generator reproduces the game's files frozen before the switch, as
// trees: every field, and the order of the schemes, the questions, the
// edges and the formulas (docs/c5-game.md, C5.1 step 5).
func TestGameGenReproducesTheFrozenData(t *testing.T) {
	_, _, out := committedGame(t)
	if diffs := sameGameFiles(t, out.Files, filepath.Join(root, frozenPreC5)); len(diffs) > 0 {
		for _, p := range diffs {
			t.Errorf("%s is not the frozen file's tree", p)
		}
	}
	if len(out.Files) != len(GamePaths) {
		t.Errorf("%d files, want %d", len(out.Files), len(GamePaths))
	}
	if !strings.Contains(string(out.Files[GameCorePath]), "# catalogue: "+out.Inputs+"\n") || len(out.Inputs) != 16 {
		t.Errorf("core.yml's header lacks the inputs' fingerprint %q", out.Inputs)
	}
}

// readGameFiles reads the game's files from a directory laid out as
// wag/data.
func readGameFiles(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	for _, p := range GamePaths {
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(p)))
		if err != nil {
			t.Fatal(err)
		}
		files[p] = b
	}
	return files
}

// playAll plays, on the game's files, every answer the NASCQ allows to every
// question (wag-cli --edges), and the 33 seeded games of the reference
// digests (internal/game, TestReferenceDigests), and reports how each ended.
func playAll(t *testing.T, files map[string][]byte) ([]game.EdgeCase, []string) {
	t.Helper()
	cases, err := game.EdgeCases(func() (*game.GameState, error) {
		ad, err := carneades.NewAdapterFromYAML(files[GameCorePath])
		if err != nil {
			return nil, err
		}
		n, err := nascq.LoadFromBytes(files[GameSchemesPath], files[GameCQsPath], files[GameEdgesPath])
		if err != nil {
			return nil, err
		}
		return game.NewGame(game.DefaultGameConfig(), ad, n, []string{"Pro", "Con"}), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	type ref struct {
		mode, claim string
		turns       int
		p1, p2      string
		s1, s2      int64
	}
	var refs []ref
	for seed := int64(1); seed <= 8; seed++ {
		refs = append(refs, ref{"classic", "", 12, "random", "random", seed, seed + 100})
	}
	for seed := int64(1); seed <= 4; seed++ {
		refs = append(refs,
			ref{"classic", "", 6, "greedy", "random", seed, seed + 100},
			ref{"infinite", "", 10, "random", "random", seed, seed + 100},
			ref{"target", "", 12, "random", "random", seed, seed + 100},
			ref{"cq", "", 12, "random", "random", seed, seed + 100},
			ref{"debate", "should_be_performed(carbon_tax)", 10, "random", "random", seed, seed + 100})
	}
	refs = append(refs,
		ref{"classic", "", 5, "greedy", "greedy", 1, 2},
		ref{"infinite", "", 5, "greedy", "random", 3, 4},
		ref{"target", "", 5, "random", "greedy", 5, 6},
		ref{"cq", "", 5, "greedy", "random", 7, 8},
		ref{"debate", "should_be_performed(carbon_tax)", 5, "greedy", "greedy", 9, 10})
	agent := func(name, kind string, seed int64) game.Agent {
		if kind == "greedy" {
			return game.NewGreedyAgent(name, seed, 2)
		}
		return game.NewRandomAgent(name, seed, 0.15, 0.55)
	}
	var games []string
	for _, g := range refs {
		cfg, err := game.ModeConfig(g.mode)
		if err != nil {
			t.Fatal(err)
		}
		cfg.MaxTurns = g.turns
		if g.claim != "" {
			cfg.Objective, cfg.Claim = game.ObjectiveClaim, g.claim
		}
		r := game.RunGame(cfg, game.DefaultRunnerConfig(), files[GameCorePath], files[GameSchemesPath], files[GameCQsPath],
			files[GameEdgesPath], []game.Agent{agent("Alice", g.p1, g.s1), agent("Bob", g.p2, g.s2)})
		if r.Panicked {
			t.Fatalf("%+v panicked: %s", g, r.PanicMessage)
		}
		games = append(games, fmt.Sprintf("%+v moves=%d rejected=%d abstained=%d alice=%.6f bob=%.6f winner=%s verdict=%q reason=%q kinds=%v digest=%s",
			g, r.MoveCount, r.InvalidCount, r.AbstainCount, r.Scores["Alice"], r.Scores["Bob"], r.Winner, r.Verdict, r.Reason, r.MoveKinds, r.MoveDigest))
	}
	return cases, games
}

// The game plays the generated files as it plays the frozen ones: every
// answer to every question does the same, and the reference games end the
// same, move for move.
func TestGameGenPlaysTheSame(t *testing.T) {
	_, _, out := committedGame(t)
	c1, g1 := playAll(t, readGameFiles(t, filepath.Join(root, frozenPreC5)))
	c2, g2 := playAll(t, out.Files)
	if len(c1) == 0 || len(g1) != 33 {
		t.Fatalf("%d answers, %d games", len(c1), len(g1))
	}
	if !reflect.DeepEqual(c1, c2) {
		t.Errorf("the answers play otherwise on the generated files (%d and %d)", len(c1), len(c2))
	}
	for i := range g1 {
		if g1[i] != g2[i] {
			t.Errorf("game %d differs:\n frozen    %s\n generated %s", i+1, g1[i], g2[i])
		}
	}
}

// freshGame reads the committed catalogue's game data, to edit.
func freshGame(t *testing.T) (map[string]*Entry, *GameData) {
	t.Helper()
	dir := filepath.Join(root, filepath.FromSlash(Dir))
	entries, _, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	d, err := ReadGameDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	return entries, d
}

// gameCQ returns an entry's question with a game part.
func gameCQ(t *testing.T, e *Entry, id string) *GameCQ {
	t.Helper()
	for i := range e.CQs {
		if q := e.CQs[i].Game; q != nil && q.ID == id {
			return q
		}
	}
	t.Fatalf("%s has no game question %s", e.ID, id)
	return nil
}

// without returns a mapping without a key, which it must have.
func without(t *testing.T, m yaml.MapSlice, key string) yaml.MapSlice {
	t.Helper()
	var out yaml.MapSlice
	for _, it := range m {
		if it.Key != key {
			out = append(out, it)
		}
	}
	if len(out) == len(m) {
		t.Fatalf("no key %s", key)
	}
	return out
}

// The generator refuses a catalogue whose game parts disagree with one
// another or break a question's rule, naming the fault.
func TestGameGenFindsFaults(t *testing.T) {
	src := sources(t)
	for _, c := range []struct {
		name, want string
		edit       func(t *testing.T, es map[string]*Entry, d *GameData)
	}{
		{"a carneades_id that is not the entry's scheme", "the game part AS:sign names core.yml's sign_renamed, which is walton.yml's sign_renamed, not this entry's scheme sign", func(t *testing.T, es map[string]*Entry, d *GameData) {
			for i, it := range es["sign.yml"].Game.NASCQ {
				if it.Key == "carneades_id" {
					es["sign.yml"].Game.NASCQ[i].Value = "sign_renamed"
				}
			}
			d.CoreTemplate = strings.Replace(d.CoreTemplate, coreMarker+"sign\n", coreMarker+"sign_renamed\n", 1)
		}},
		{"no NASCQ record", "has no NASCQ record", func(t *testing.T, es map[string]*Entry, d *GameData) {
			es["sign.yml"].Game.NASCQ = nil
		}},
		{"no carneades_id", "has no carneades_id", func(t *testing.T, es map[string]*Entry, d *GameData) {
			es["sign.yml"].Game.NASCQ = without(t, es["sign.yml"].Game.NASCQ, "carneades_id")
		}},
		{"a game part order.yml lacks", "its game part AS:sign is not in game/order.yml", func(t *testing.T, es map[string]*Entry, d *GameData) {
			var keep []string
			for _, s := range d.Order.Schemes {
				if s != "AS:sign" {
					keep = append(keep, s)
				}
			}
			d.Order.Schemes = keep
		}},
		{"a scheme no entry has", "names AS:nothing, which is no entry's game part", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Order.Schemes = append(d.Order.Schemes, "AS:nothing")
		}},
		{"a scheme twice", "names AS:sign twice", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Order.Schemes = append(d.Order.Schemes, "AS:sign")
		}},
		{"two entries, one game part", "is the game part of both", func(t *testing.T, es map[string]*Entry, d *GameData) {
			es["analogy.yml"].Game.ID = "AS:sign"
		}},
		{"two game schemes, one core.yml scheme", "the game part AS:analogy names core.yml's sign, which is walton.yml's sign, not this entry's scheme analogy", func(t *testing.T, es map[string]*Entry, d *GameData) {
			for i, it := range es["analogy.yml"].Game.NASCQ {
				if it.Key == "carneades_id" {
					es["analogy.yml"].Game.NASCQ[i].Value = "sign"
				}
			}
		}},
		{"an ask edge twice", "the ask edge CQ:other_goals of AS:practical_reasoning twice", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Order.Ask = append(d.Order.Ask, d.Order.Ask[0])
		}},
		{"an ask edge of an unknown question", "no such question", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Order.Ask[0].CQ = "CQ:nothing"
		}},
		{"an ask edge of an unknown scheme", "no such game scheme", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Order.Ask[0].Scheme = "AS:nothing"
		}},
		{"an ask edge no entry asks", "0 questions with the game part CQ:likelihood", func(t *testing.T, es map[string]*Entry, d *GameData) {
			for i := range es["sign.yml"].CQs {
				if q := es["sign.yml"].CQs[i].Game; q != nil && q.ID == "CQ:likelihood" {
					es["sign.yml"].CQs[i].Game = nil
				}
			}
		}},
		{"two questions, one game part", "2 questions with the game part CQ:expertise", func(t *testing.T, es map[string]*Entry, d *GameData) {
			e := es["expert_opinion.yml"]
			q := *gameCQ(t, e, "CQ:expertise")
			e.CQs = append(e.CQs, CQ{ID: "walton/expert", Game: &q})
		}},
		{"a question order.yml does not ask", "asks CQ:other_event, which game/order.yml does not", func(t *testing.T, es map[string]*Entry, d *GameData) {
			var keep []GameEdge
			for _, a := range d.Order.Ask {
				if a != (GameEdge{CQ: "CQ:other_event", Scheme: "AS:sign"}) {
					keep = append(keep, a)
				}
			}
			d.Order.Ask = keep
		}},
		{"a premise's pattern at an exception", "asked at an exception, its pattern \"strong_correlation(A,B)\"", func(t *testing.T, es map[string]*Entry, d *GameData) {
			gameCQ(t, es["sign.yml"], "CQ:likelihood").Locus = "exception"
		}},
		{"an exception's pattern at a premise", "asked at a premise, its pattern \"untrustworthy(W)\"", func(t *testing.T, es map[string]*Entry, d *GameData) {
			gameCQ(t, es["expert_opinion.yml"], "CQ:trustworthiness").Locus = "premise"
		}},
		{"a pattern the formulas lack", "no premise or assumption of the game's expert_opinion", func(t *testing.T, es map[string]*Entry, d *GameData) {
			gameCQ(t, es["expert_opinion.yml"], "CQ:expertise").Pattern = "expert(D,W)"
		}},
		{"a premise question without a pattern", "asked at a premise, its pattern \"\"", func(t *testing.T, es map[string]*Entry, d *GameData) {
			gameCQ(t, es["expert_opinion.yml"], "CQ:expertise").Pattern = ""
		}},
		{"a conclusion question with a pattern", "asked of the conclusion, it has a pattern", func(t *testing.T, es map[string]*Entry, d *GameData) {
			gameCQ(t, es["practical_reasoning1.yml"], "CQ:other_goals").Pattern = "possible(A)"
		}},
		{"an unknown locus", "the locus \"premises\" is not one of", func(t *testing.T, es map[string]*Entry, d *GameData) {
			gameCQ(t, es["expert_opinion.yml"], "CQ:expertise").Locus = "premises"
		}},
		{"formulas without variables", "gives formulas but no variables", func(t *testing.T, es map[string]*Entry, d *GameData) {
			es["practical_reasoning1.yml"].Game.Core = without(t, es["practical_reasoning1.yml"].Game.Core, "variables")
		}},
		{"a formula that is not a string", "not a formula", func(t *testing.T, es map[string]*Entry, d *GameData) {
			for i, it := range es["sign.yml"].Game.Core {
				if it.Key == "premises" {
					es["sign.yml"].Game.Core[i].Value = []interface{}{"a(X)", 3}
				}
			}
		}},
		{"a question without a record", "CQ:other_goals has no record", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Questions[0].Record = nil
		}},
		{"a question twice", "CQ:other_goals twice", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Questions = append(d.Questions, d.Questions[0])
		}},
		{"a backup edge of an unknown question", "the backup edge CQ:nothing by", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Order.Backup[0].CQ = "CQ:nothing"
		}},
		{"a backup edge twice", "twice", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Order.Backup = append(d.Order.Backup, d.Order.Backup[0])
		}},
		{"a scheme without its marker", "has no marker for sign (AS:sign)", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.CoreTemplate = strings.Replace(d.CoreTemplate, coreMarker+"sign\n", "", 1)
		}},
		{"a marker for no scheme", "a marker for nothing, which no game scheme is", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.CoreTemplate += coreMarker + "nothing\n"
		}},
		{"two markers", "two markers for sign", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.CoreTemplate += coreMarker + "sign\n"
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			es, d := freshGame(t)
			c.edit(t, es, d)
			_, err := GenerateGame(es, d, src.GameSources())
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("got %v, want an error with %q", err, c.want)
			}
		})
	}
	// A game scheme whose entry records no formulas of its own, and
	// walton.yml lacks: refused.
	es, d := freshGame(t)
	gs := src.GameSources()
	delete(gs.Formulas, "expert_opinion")
	if _, err := GenerateGame(es, d, gs); err == nil || !strings.Contains(err.Error(), "walton.yml has no scheme expert_opinion, whose formulas the game's AS:expert_opinion plays") {
		t.Errorf("walton.yml without expert_opinion: got %v", err)
	}
	// An upstream id that walton.yml's migrations map to the entry's scheme
	// is the entry's: core.yml's block takes it.
	es, d = freshGame(t)
	for i, it := range es["sign.yml"].Game.NASCQ {
		if it.Key == "carneades_id" {
			es["sign.yml"].Game.NASCQ[i].Value = "sign_upstream"
		}
	}
	d.CoreTemplate = strings.Replace(d.CoreTemplate, coreMarker+"sign\n", coreMarker+"sign_upstream\n", 1)
	gs = src.GameSources()
	gs.Migrations = map[string]string{"sign_upstream": "sign"}
	if g, err := GenerateGame(es, d, gs); err != nil || !strings.Contains(string(g.Files[GameCorePath]), "  - id: sign_upstream\n") {
		t.Errorf("a migrated id: %v", err)
	}
	// Unedited, it generates, and a template read with CRLF line endings
	// fills the same markers.
	es, d = freshGame(t)
	if _, err := GenerateGame(es, d, src.GameSources()); err != nil {
		t.Fatal(err)
	}
	d.CoreTemplate = strings.ReplaceAll(d.CoreTemplate, "\n", "\r\n")
	if _, err := GenerateGame(es, d, src.GameSources()); err != nil {
		t.Errorf("a CRLF template: %v", err)
	}
}

// The fingerprint in core.yml's header is that of the values the generator
// reads (core.yml's template, its comments included): drafting an entry's
// other parts leaves it, as it leaves the files; a value read changes both.
func TestGameGenFingerprint(t *testing.T) {
	src := sources(t)
	es, d := freshGame(t)
	base, err := GenerateGame(es, d, src.GameSources())
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name string
		same bool
		edit func(es map[string]*Entry, d *GameData, f map[string]*Formulas)
	}{
		{"an entry's summary", true, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			es["sign.yml"].Summary.Text = "something else"
		}},
		{"a question's paraphrase", true, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			for i := range es["sign.yml"].CQs {
				es["sign.yml"].CQs[i].Paraphrase += " more"
			}
		}},
		{"a question's text in its entry", true, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			for i := range es["sign.yml"].CQs {
				if q := es["sign.yml"].CQs[i].Game; q != nil {
					q.Text += " more"
				}
			}
		}},
		{"an entry with no game part", true, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			for _, e := range es {
				if e.Game == nil && e.Formulas != nil {
					e.Formulas.Premises = append(e.Formulas.Premises, "x(Y)")
					return
				}
			}
		}},
		{"a weight", false, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			for i, it := range es["sign.yml"].Game.NASCQ {
				if it.Key == "base_weight" {
					es["sign.yml"].Game.NASCQ[i].Value = 0.51
				}
			}
		}},
		{"an entry's copy of walton.yml's formulas", true, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			p := es["expert_opinion.yml"].Formulas.Premises
			p[0], p[1] = p[1], p[0]
		}},
		{"a formula of walton.yml the game plays", false, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			w := *f["expert_opinion"]
			w.Premises = append([]string{w.Premises[1], w.Premises[0]}, w.Premises[2:]...)
			f["expert_opinion"] = &w
		}},
		{"a formula of a scheme the game does not play", true, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			w := *f["cause_to_effect"]
			w.Premises = append([]string{"x(Y)"}, w.Premises...)
			f["cause_to_effect"] = &w
		}},
		{"a question's record", false, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			d.Questions[0].Record = append(yaml.MapSlice{}, d.Questions[0].Record...)
			d.Questions[0].Record[0].Value = "Another Name"
		}},
		{"the order of the edges", false, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			d.Order.Ask[0], d.Order.Ask[1] = d.Order.Ask[1], d.Order.Ask[0]
		}},
		{"core.yml's template", false, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			d.CoreTemplate += "# a comment\n"
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			es, d := freshGame(t)
			gs := src.GameSources()
			c.edit(es, d, gs.Formulas)
			got, err := GenerateGame(es, d, gs)
			if err != nil {
				t.Fatal(err)
			}
			sameFiles := true
			for _, p := range GamePaths {
				if !bytes.Equal(got.Files[p], base.Files[p]) {
					sameFiles = false
				}
			}
			if (got.Inputs == base.Inputs) != c.same || sameFiles != c.same {
				t.Errorf("fingerprint %s (was %s), files the same: %t; want the same: %t", got.Inputs, base.Inputs, sameFiles, c.same)
			}
		})
	}
}

// A pattern takes the NASCQ's braces round the scheme's variables only.
func TestRebrace(t *testing.T) {
	g := &gameIn{vars: map[string]bool{"W": true, "S": true, "A": true}}
	for in, want := range map[string]string{
		"based_on_evidence(asserts(W,S))": "based_on_evidence(asserts({W},{S}))",
		"possible(A)":                     "possible({A})",
		"more_on_point(A,C)":              "more_on_point({A},C)",
		"Asserts(WS,S)":                   "Asserts(WS,{S})",
		"":                                "",
	} {
		if got := g.rebrace(in); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}

// The JSON writer keeps the order of keys, writes HTML characters as they
// are, puts an object of scalars in a list on one line, and refuses what
// JSON cannot hold.
func TestWriteJSON(t *testing.T) {
	v := yaml.MapSlice{
		{Key: "z", Value: "a & b <c>"},
		{Key: "list", Value: []interface{}{yaml.MapSlice{{Key: "b", Value: 1}, {Key: "a", Value: true}},
			yaml.MapSlice{{Key: "cq", Value: "CQ:a_long_question"}, {Key: "scheme", Value: "AS:a_long_scheme_name"}, {Key: "pattern", Value: "realize_more_effectively({G},{A})"}}}},
		{Key: "w", Value: 0.7},
		{Key: "n", Value: nil},
		{Key: "short", Value: []interface{}{"x", "y"}},
		{Key: "obj", Value: yaml.MapSlice{{Key: "k", Value: "v"}}},
		{Key: "empty", Value: yaml.MapSlice{}},
	}
	var b bytes.Buffer
	if err := writeJSON(&b, v, "", false, false); err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"z\": \"a & b <c>\",\n  \"list\": [\n    {\"b\": 1, \"a\": true},\n" +
		"    {\"cq\": \"CQ:a_long_question\", \"scheme\": \"AS:a_long_scheme_name\", \"pattern\": \"realize_more_effectively({G},{A})\"}\n" +
		"  ],\n  \"w\": 0.7,\n  \"n\": null,\n" +
		"  \"short\": [\"x\", \"y\"],\n  \"obj\": {\"k\": \"v\"},\n  \"empty\": {}\n}"
	if b.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", b.String(), want)
	}
	back, err := decodeOrderedJSON(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back, v) {
		t.Errorf("read back: %#v", back)
	}
	b.Reset()
	if err := writeJSON(&b, v, "", true, false); err != nil {
		t.Fatal(err)
	}
	if want := `{"z":"a & b <c>","list":[{"b":1,"a":true},{"cq":"CQ:a_long_question","scheme":"AS:a_long_scheme_name","pattern":"realize_more_effectively({G},{A})"}],"w":0.7,"n":null,"short":["x","y"],"obj":{"k":"v"},"empty":{}}`; b.String() != want {
		t.Errorf("compact: got %s, want %s", b.String(), want)
	}
	for _, bad := range []interface{}{map[string]int{"a": 1}, yaml.MapSlice{{Key: 1, Value: "x"}}, struct{}{}} {
		b.Reset()
		if err := writeJSON(&b, bad, "", false, false); err == nil {
			t.Errorf("%#v: no error", bad)
		}
	}
}
