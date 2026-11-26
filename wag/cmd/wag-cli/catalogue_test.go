package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/wag/internal/catalogue"
)

// tempRepo copies what wag-cli --catalogue-gen reads (the sources and the
// catalogue) into a temporary repository.
func tempRepo(t *testing.T) string {
	t.Helper()
	const root = "../../.."
	dst := t.TempDir()
	p := catalogue.SourcePaths
	for _, path := range []string{p.Walton, p.Links, p.Migrations, p.Changes, p.Families, p.Schemes, p.CQs, p.Edges, p.Core} {
		copyFile(t, filepath.Join(root, path), filepath.Join(dst, path))
	}
	copyTree(t, filepath.Join(root, catalogue.Dir), filepath.Join(dst, catalogue.Dir))
	return dst
}

func copyFile(t *testing.T, from, to string) {
	t.Helper()
	b, err := os.ReadFile(from)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(to, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func copyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(from, path)
		if err != nil {
			return err
		}
		copyFile(t, path, filepath.Join(to, rel))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// gameFiles reads the game's files of a repository.
func gameFiles(t *testing.T, repo string) map[string][]byte {
	t.Helper()
	return readGameDir(t, filepath.Join(repo, catalogue.GameDataDir))
}

// readGameDir reads the game's files of a directory laid out as wag/data.
func readGameDir(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	for _, p := range catalogue.GamePaths {
		b, err := os.ReadFile(filepath.Join(dir, p))
		if err != nil {
			t.Fatal(err)
		}
		out[p] = b
	}
	return out
}

// setWeight sets the base weight of an entry's game scheme.
func setWeight(t *testing.T, dir, id string, w float64) {
	t.Helper()
	path := filepath.Join(dir, catalogue.SchemesDir, id+".yml")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	e, err := catalogue.Unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for i, it := range e.Game.NASCQ {
		if it.Key == "base_weight" {
			e.Game.NASCQ[i].Value, found = w, true
		}
	}
	if !found {
		t.Fatalf("%s has no base_weight", id)
	}
	if b, err = catalogue.Marshal(e); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// weight reads it back.
func weight(t *testing.T, dir, id string) interface{} {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, catalogue.SchemesDir, id+".yml"))
	if err != nil {
		t.Fatal(err)
	}
	e, err := catalogue.Unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range e.Game.NASCQ {
		if it.Key == "base_weight" {
			return it.Value
		}
	}
	return nil
}

// gameWeight reads a game scheme's base weight in the game's schemes.json.
func gameWeight(t *testing.T, files map[string][]byte, gid string) float64 {
	t.Helper()
	var s map[string]struct {
		W float64 `json:"base_weight"`
	}
	if err := json.Unmarshal(files[catalogue.GameSchemesPath], &s); err != nil {
		t.Fatal(err)
	}
	return s[gid].W
}

func sameFiles(a, b map[string][]byte) bool {
	for p := range a {
		if !bytes.Equal(a[p], b[p]) {
			return false
		}
	}
	return len(a) == len(b)
}

// wag-cli --catalogue-gen writes the game's files from the repository's
// catalogue first, so an edit to an entry's game part is what the game and
// the entry say afterwards; a second run changes nothing.
func TestCatalogueGenWritesTheGameFromTheCatalogue(t *testing.T) {
	repo := tempRepo(t)
	dir := filepath.Join(repo, catalogue.Dir)
	setWeight(t, dir, "sign", 0.42)
	if code := runCatalogueGen(repo, "", "test"); code != 0 {
		t.Fatalf("--catalogue-gen: status %d", code)
	}
	if w := weight(t, dir, "sign"); w != 0.42 {
		t.Errorf("the entry's weight is %v after --catalogue-gen, not 0.42", w)
	}
	files := gameFiles(t, repo)
	if w := gameWeight(t, files, "AS:sign"); w != 0.42 {
		t.Errorf("the game's schemes.json gives AS:sign %v, not 0.42", w)
	}
	entry, err := os.ReadFile(filepath.Join(dir, catalogue.SchemesDir, "sign.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if code := runCatalogueGen(repo, "", "test"); code != 0 {
		t.Fatalf("--catalogue-gen again: status %d", code)
	}
	again, err := os.ReadFile(filepath.Join(dir, catalogue.SchemesDir, "sign.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !sameFiles(files, gameFiles(t, repo)) || !bytes.Equal(entry, again) {
		t.Error("a second run changed the game's files or the entry")
	}
}

// Run on another catalogue, wag-cli --catalogue-gen makes that catalogue's
// entries from the game's files it makes, keeping its edits, and leaves the
// repository's game files alone; --game-gen needs --game-out for it.
func TestCatalogueCommandsOnAnotherCatalogue(t *testing.T) {
	repo := tempRepo(t)
	other := filepath.Join(t.TempDir(), "catalogue")
	copyTree(t, filepath.Join(repo, catalogue.Dir), other)
	setWeight(t, other, "sign", 0.42)
	before := gameFiles(t, repo)
	if code := runCatalogueGen(repo, other, "test"); code != 0 {
		t.Fatalf("--catalogue-gen --catalogue OTHER: status %d", code)
	}
	if w := weight(t, other, "sign"); w != 0.42 {
		t.Errorf("the other catalogue's edit was reverted: %v", w)
	}
	if !sameFiles(before, gameFiles(t, repo)) {
		t.Error("--catalogue-gen on another catalogue changed the repository's game files")
	}
	if code := runGameGen(repo, other, ""); code == 0 {
		t.Error("--game-gen on another catalogue without --game-out: status 0")
	}
	if !sameFiles(before, gameFiles(t, repo)) {
		t.Error("--game-gen on another catalogue changed the repository's game files")
	}
	out := t.TempDir()
	if code := runGameGen(repo, other, out); code != 0 {
		t.Fatalf("--game-gen --game-out: status %d", code)
	}
	if w := gameWeight(t, readGameDir(t, out), "AS:sign"); w != 0.42 {
		t.Errorf("--game-gen --game-out wrote AS:sign's weight %v, not 0.42", w)
	}
	// On the repository's own catalogue, --game-gen writes the repository's
	// files, which are current: they do not change.
	if code := runGameGen(repo, "", ""); code != 0 || !sameFiles(before, gameFiles(t, repo)) {
		t.Errorf("--game-gen on the repository's catalogue: status %d, files changed: %t", code, !sameFiles(before, gameFiles(t, repo)))
	}
}
