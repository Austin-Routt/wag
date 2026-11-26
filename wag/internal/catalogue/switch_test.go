package catalogue

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Step C5.1's switch (docs/c5-game.md): wag-cli --catalogue-gen makes the
// game's files from the catalogue first, with walton.yml's formulas, then
// the entries from them.

// regenerate runs --catalogue-gen's order on a catalogue directory: the
// game's files from it (WithGame), the sources parsed with them, then the
// entries. It returns the sources.
func regenerate(t *testing.T, f Files, dir string) *Sources {
	t.Helper()
	f, _, err := WithGame(f, dir)
	if err != nil {
		t.Fatal(err)
	}
	src, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (&Generator{Src: src, Today: "2026-11-26", By: "test"}).Write(dir); err != nil {
		t.Fatal(err)
	}
	return src
}

// gameErrors are a report's errors about the game's data.
func gameErrors(r *Report) []string {
	var out []string
	for _, f := range r.Findings {
		if f.Error && (strings.HasPrefix(f.File, GameDataDir+"/") || strings.HasPrefix(f.File, GameDir) || strings.Contains(f.Msg, "game")) {
			out = append(out, f.String())
		}
	}
	return out
}

// An edit to an entry's game part is what the game's files and the entries
// say after --catalogue-gen. Made from the game's files as they were, the
// entries would lose it, which is why the files come first.
func TestCatalogueGenKeepsAGamePartEdit(t *testing.T) {
	s := sources(t)
	f, err := ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	weight := func(e *Entry) interface{} { return value(e.Game.NASCQ, "base_weight") }
	for _, first := range []bool{true, false} {
		dir := generate(t, s)
		e := readEntry(t, dir, "sign")
		was := weight(e)
		for i, it := range e.Game.NASCQ {
			if it.Key == "base_weight" {
				e.Game.NASCQ[i].Value = 0.42
			}
		}
		writeEntry(t, dir, e)
		if first {
			src := regenerate(t, f, dir)
			if got := weight(readEntry(t, dir, "sign")); got != 0.42 {
				t.Errorf("the entry's weight is %v after --catalogue-gen, not 0.42", got)
			}
			var schemes map[string]struct {
				Weight float64 `json:"base_weight"`
			}
			if err := json.Unmarshal(src.files.Schemes, &schemes); err != nil {
				t.Fatal(err)
			}
			if schemes["AS:sign"].Weight != 0.42 {
				t.Errorf("the game's schemes.json gives AS:sign %v, not 0.42", schemes["AS:sign"].Weight)
			}
			r, err := Check(dir, src, DefaultOptions())
			if err != nil {
				t.Fatal(err)
			}
			if errs := errorsOf(r); len(errs) > 0 {
				t.Errorf("the check after the edit: %v", errs)
			}
		} else {
			if _, err := (&Generator{Src: s, Today: "2026-11-26", By: "test"}).Write(dir); err != nil {
				t.Fatal(err)
			}
			if got := weight(readEntry(t, dir, "sign")); got != was {
				t.Errorf("made from the files as they were, the weight is %v, not %v: the test shows nothing", got, was)
			}
		}
	}
}

// A change to walton.yml's formulas reaches the game where the entry records
// none of its own, and leaves no override behind. Generated from the
// entries' copies, which --catalogue-gen renews only afterwards, the game
// would keep the old formulas and the entry would record them as an
// override: the generator takes walton.yml's.
func TestCatalogueGenTakesWaltonsFormulas(t *testing.T) {
	s := sources(t)
	f, err := ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	const old = "      - expert(W,D)\n      - in_domain(S,D)\n      - asserts(W,S)\n    exceptions:\n      - untrustworthy(W)\n"
	const swapped = "      - in_domain(S,D)\n      - expert(W,D)\n      - asserts(W,S)\n    exceptions:\n      - untrustworthy(W)\n"
	if strings.Count(string(f.Walton), old) != 1 {
		t.Fatal("walton.yml's expert_opinion is not as the test expects")
	}
	f.Walton = []byte(strings.Replace(string(f.Walton), old, swapped, 1))
	want := []string{"in_domain(S,D)", "expert(W,D)", "asserts(W,S)"}

	dir := generate(t, s)
	src := regenerate(t, f, dir)
	e := readEntry(t, dir, "expert_opinion")
	if value(e.Game.Core, "premises") != nil {
		t.Errorf("the entry records an override: %v", e.Game.Core)
	}
	if !reflect.DeepEqual(e.Formulas.Premises, want) || !reflect.DeepEqual(src.Core["expert_opinion"].Premises, want) {
		t.Errorf("premises: the entry's %v, the game's %v; want %v", e.Formulas.Premises, src.Core["expert_opinion"].Premises, want)
	}
	r, err := Check(dir, src, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if errs := gameErrors(r); len(errs) > 0 {
		t.Errorf("the check finds: %v", errs)
	}

	// From the entries' copies instead: the old formulas, as an override.
	dir = generate(t, s)
	entries, _, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	d, err := ReadGameDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	stale, err := GenerateGame(entries, d, GameSources{Migrations: s.Migrations})
	if err != nil {
		t.Fatal(err)
	}
	g := f
	g.Core, g.Schemes, g.CQs, g.Edges = stale.Files[GameCorePath], stale.Files[GameSchemesPath], stale.Files[GameCQsPath], stale.Files[GameEdgesPath]
	src, err = Parse(g)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (&Generator{Src: src, Today: "2026-11-26", By: "test"}).Write(dir); err != nil {
		t.Fatal(err)
	}
	if value(readEntry(t, dir, "expert_opinion").Game.Core, "premises") == nil {
		t.Error("from the entries' copies, no stale override: the test shows nothing")
	}
}

// The check requires the game's files to be what the catalogue makes, byte
// for byte: the files frozen before the switch, with the same values, are
// four errors, and nothing else is.
func TestCheckRequiresTheGeneratedBytes(t *testing.T) {
	f, err := ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	frozen := readGameFiles(t, filepath.Join(root, frozenPreC5))
	f.Core, f.Schemes, f.CQs, f.Edges = frozen[GameCorePath], frozen[GameSchemesPath], frozen[GameCQsPath], frozen[GameEdgesPath]
	src, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Check(filepath.Join(root, filepath.FromSlash(Dir)), src, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]bool{}
	var others []string
	for _, x := range r.Findings {
		if !x.Error {
			continue
		}
		if strings.HasPrefix(x.File, GameDataDir+"/") && strings.Contains(x.Msg, "not what the catalogue makes") {
			files[strings.TrimPrefix(x.File, GameDataDir+"/")] = true
		} else {
			others = append(others, x.String())
		}
	}
	if len(files) != len(GamePaths) || len(others) > 0 {
		t.Errorf("errors on %v (want the %d game files), and others: %v", files, len(GamePaths), others)
	}
}

// A change to walton.yml that --catalogue-gen has not yet carried is seen
// in the game's files too: core.yml is not what the catalogue makes.
func TestCheckSeesAWaltonChangeInTheGame(t *testing.T) {
	f, err := ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	const old = "      - expert(W,D)\n      - in_domain(S,D)\n      - asserts(W,S)\n    exceptions:\n      - untrustworthy(W)\n"
	const swapped = "      - in_domain(S,D)\n      - expert(W,D)\n      - asserts(W,S)\n    exceptions:\n      - untrustworthy(W)\n"
	if strings.Count(string(f.Walton), old) != 1 {
		t.Fatal("walton.yml's expert_opinion is not as the test expects")
	}
	f.Walton = []byte(strings.Replace(string(f.Walton), old, swapped, 1))
	src, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Check(filepath.Join(root, filepath.FromSlash(Dir)), src, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range r.Findings {
		if x.Error && x.File == GameDataDir+"/"+GameCorePath {
			return
		}
	}
	t.Errorf("no error on %s/%s; errors: %v", GameDataDir, GameCorePath, errorsOf(r))
}
