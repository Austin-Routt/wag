package server

import (
	"os"
	"path/filepath"
	"testing"
)

// Each folder of data/frozen is the game's data as it was before a change
// of the data's identity, named by its fingerprint, so that --data on it
// replays the games played with it (data/frozen/README.md): its fingerprint
// is its name, and a server starts with it.
func TestFrozenDataIsItsFingerprint(t *testing.T) {
	const frozen = "../../data/frozen"
	dirs, err := os.ReadDir(frozen)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		n++
		data, err := ReadData(filepath.Join(frozen, d.Name()))
		if err != nil {
			t.Fatalf("data/frozen/%s: %v", d.Name(), err)
		}
		if h := data.Hash(); h != d.Name() {
			t.Errorf("data/frozen/%s: its fingerprint is %s", d.Name(), h)
		}
		if _, err := New(data, ""); err != nil {
			t.Errorf("data/frozen/%s: a server does not start with it: %v", d.Name(), err)
		}
	}
	if n == 0 {
		t.Error("no frozen data")
	}
}
