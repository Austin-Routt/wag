package catalogue

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

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

// The generator makes the committed game files, byte for byte, with the
// fingerprint of the values it read in core.yml's header. (Until step C5.2's
// switch it reproduced the files frozen before C5.1's, as trees, and the game
// played both alike, the reference games move for move: C5.1's guarantee,
// kept in the history at 1273555. The 33 reference games play the frozen
// files since, TestReferenceDigests.)
func TestGameGenMakesTheCommittedData(t *testing.T) {
	_, _, out := committedGame(t)
	committed := readGameFiles(t, filepath.Join(root, GameDataDir))
	for _, p := range GamePaths {
		if !bytes.Equal(out.Files[p], bytes.ReplaceAll(committed[p], []byte("\r\n"), []byte("\n"))) {
			t.Errorf("%s is not the committed file", p)
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
		{"an ask edge no entry asks", "0 questions with the game part CQ:correlation", func(t *testing.T, es map[string]*Entry, d *GameData) {
			for i := range es["sign.yml"].CQs {
				if q := es["sign.yml"].CQs[i].Game; q != nil && q.ID == "CQ:correlation" {
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
			gameCQ(t, es["sign.yml"], "CQ:correlation").Locus = "exception"
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
			gameCQ(t, es["practical_reasoning1.yml"], "CQ:side_effect").Pattern = "possible_bring_about(P,A)"
		}},
		{"an unknown locus", "the locus \"premises\" is not one of", func(t *testing.T, es map[string]*Entry, d *GameData) {
			gameCQ(t, es["expert_opinion.yml"], "CQ:expertise").Locus = "premises"
		}},
		{"formulas without variables", "gives formulas but no variables", func(t *testing.T, es map[string]*Entry, d *GameData) {
			es["sign.yml"].Game.Core = without(t, es["sign.yml"].Game.Core, "variables")
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
		// C5.2's rules (docs/c5-game.md, step 4), each wired into the generator.
		{"one id at two formulas", "CQ:trustworthiness is asked at dishonest({W}) (AS:position_to_know) and untrustworthy({W}) (AS:expert_opinion): one id is one question, asked at one formula", func(t *testing.T, es map[string]*Entry, d *GameData) {
			gameCQ(t, es["position_to_know.yml"], "CQ:honesty").ID = "CQ:trustworthiness"
			for i, a := range d.Order.Ask {
				if a == (GameEdge{CQ: "CQ:honesty", Scheme: "AS:position_to_know"}) {
					d.Order.Ask[i].CQ = "CQ:trustworthiness"
				}
			}
			var keep []GameQuestion
			for _, q := range d.Questions {
				if q.ID != "CQ:honesty" {
					keep = append(keep, q)
				}
			}
			d.Questions = keep
		}},
		{"one id with two texts", "CQ:too_slight reads", func(t *testing.T, es map[string]*Entry, d *GameData) {
			for i := range es["negative_consequences.yml"].CQs {
				if q := es["negative_consequences.yml"].CQs[i].Game; q != nil && q.ID == "CQ:too_slight" {
					es["negative_consequences.yml"].CQs[i].Paraphrase = "Is G too slight a harm to count in deciding on A?"
				}
			}
		}},
		{"a question no scheme asks", "CQ:orphan is asked by no game scheme", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Questions = append(d.Questions, GameQuestion{ID: "CQ:orphan", Record: d.Questions[0].Record})
		}},
		{"a record that writes its text", "gives question_template_en, which the generator makes", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Questions[0].Record = append(append(yaml.MapSlice{}, d.Questions[0].Record...), yaml.MapItem{Key: "question_template_en", Value: "Why?"})
		}},
		{"a record without answer words", "CQ:other_goals: its record has no answer_en", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.Questions[0].Record = without(t, d.Questions[0].Record, "answer_en")
		}},
		{"a paraphrase with braces", "its paraphrase has braces", func(t *testing.T, es map[string]*Entry, d *GameData) {
			for i := range es["sign.yml"].CQs {
				if q := es["sign.yml"].CQs[i].Game; q != nil && q.ID == "CQ:correlation" {
					es["sign.yml"].CQs[i].Paraphrase = "How closely does {A} go with {B}?"
				}
			}
		}},
		{"an answer scheme concluding another formula", "AS:better_means concludes realize_more_effectively({P},{G},{A}), not bad_outweighs({A},{G})", func(t *testing.T, es map[string]*Entry, d *GameData) {
			for i, b := range d.Order.Backup {
				if b.Scheme == "AS:better_means" {
					d.Order.Backup[i].CQ = "CQ:bad_outweighs"
				}
			}
		}},
		{"a bridge's entry premise that is not its premise", "AS:bridge_possible_good's entry premise \"may_be_worth_performing({G},{A})\" is none of its premises", func(t *testing.T, es map[string]*Entry, d *GameData) {
			for i, it := range es["bridge_from_possible_good.yml"].Game.NASCQ {
				if it.Key == "entry_premise" {
					es["bridge_from_possible_good.yml"].Game.NASCQ[i].Value = "may_be_worth_performing({G},{A})"
				}
			}
		}},
		{"a sentence written in the template", "the language's block writes", func(t *testing.T, es map[string]*Entry, d *GameData) {
			d.CoreTemplate = strings.Replace(d.CoreTemplate, languageMarker+"\n", languageMarker+"\n  good/1: \"%s is fine.\"\n", 1)
		}},
		{"a predicate without its negation", "good/1 has no negation's sentence", func(t *testing.T, es map[string]*Entry, d *GameData) {
			delete(d.Language.Negations, "good/1")
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
		{"the paraphrase of a question the game does not ask", true, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			for i := range es["sign.yml"].CQs {
				if es["sign.yml"].CQs[i].Game == nil {
					es["sign.yml"].CQs[i].Paraphrase += " more"
				}
			}
		}},
		{"a game question's paraphrase, its text (D009)", false, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			for i := range es["sign.yml"].CQs {
				if es["sign.yml"].CQs[i].Game != nil {
					es["sign.yml"].CQs[i].Paraphrase += " more"
				}
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
		{"a negation's sentence in game/language.yml", false, func(es map[string]*Entry, d *GameData, f map[string]*Formulas) {
			d.Language.Negations["good/1"] = "%s is not good at all."
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
	// walton.yml's templates are read for the predicates the game plays, and
	// for those only.
	for name, c := range map[string]struct {
		key  string
		same bool
	}{"a template the game plays": {"goal_of_agent/2", false}, "a template the game does not play": {"causes/2", true}} {
		es, d := freshGame(t)
		gs := src.GameSources()
		lang := map[string]string{}
		for k, v := range gs.Language {
			lang[k] = v
		}
		lang[c.key] = "%s and %s, reworded."
		gs.Language = lang
		got, err := GenerateGame(es, d, gs)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if (got.Inputs == base.Inputs) != c.same {
			t.Errorf("%s: fingerprint %s (was %s), want the same: %t", name, got.Inputs, base.Inputs, c.same)
		}
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
