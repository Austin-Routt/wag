package server

import (
	"net/http/httptest"
	"os"
	"testing"
)

// The frozen copies of the game's data (wag/data/frozen): before step C5.2
// (afd3335a23f12f41) and before step C5.3 (092b6a6f48a672d2). The records the
// tests keep (exports, games and traces made before a switch) replay on the
// copy they were played with, as a server started with --data on those files
// replays them (docs/c5-game.md): they were played with its schemes and
// questions.
const (
	beforeC52 = "afd3335a23f12f41"
	beforeC53 = "092b6a6f48a672d2"
)

// frozenData reads the frozen copy named by its fingerprint.
func frozenData(t *testing.T, fp string) Data {
	t.Helper()
	dir := "../../data/frozen/" + fp + "/"
	read := func(p string) []byte {
		b, err := os.ReadFile(dir + p)
		if err != nil {
			t.Fatalf("the frozen data %s: %v", fp, err)
		}
		return b
	}
	return Data{YAML: read("schemes/core.yml"), Schemes: read("nascq/schemes.json"), CQs: read("nascq/critical_questions.json"),
		Edges: read("nascq/edges.json"), Contexts: read("contexts.json")}
}

// prevData is the game's data before step C5.2.
func prevData(t *testing.T) Data {
	t.Helper()
	return frozenData(t, beforeC52)
}

// newFrozenServer is a test server on the frozen copy named.
func newFrozenServer(t *testing.T, fp string) (*Server, *httptest.Server) {
	t.Helper()
	s, err := New(frozenData(t, fp), "")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return s, ts
}

// newPrevServer is a test server on the frozen pre-C5.2 data.
func newPrevServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	return newFrozenServer(t, beforeC52)
}
