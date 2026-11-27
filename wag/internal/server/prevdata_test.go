package server

import (
	"net/http/httptest"
	"os"
	"testing"
)

// The game's data before step C5.2 (wag/data/frozen/afd3335a23f12f41). The
// records the tests keep (exports and games made before the switch) replay
// on it, as a server started with --data on those files replays them
// (docs/c5-game.md, C5.2): they were played with its schemes and questions.
func prevData(t *testing.T) Data {
	t.Helper()
	const dir = "../../data/frozen/afd3335a23f12f41/"
	read := func(p string) []byte {
		b, err := os.ReadFile(dir + p)
		if err != nil {
			t.Fatalf("the frozen pre-C5.2 data: %v", err)
		}
		return b
	}
	return Data{YAML: read("schemes/core.yml"), Schemes: read("nascq/schemes.json"), CQs: read("nascq/critical_questions.json"),
		Edges: read("nascq/edges.json"), Contexts: read("contexts.json")}
}

// newPrevServer is a test server on the frozen pre-C5.2 data.
func newPrevServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	s, err := New(prevData(t), "")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return s, ts
}
