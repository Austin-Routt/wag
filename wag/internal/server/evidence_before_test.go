package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/llm"
)

// Y7 (docs/e1-evidence.md): an export and a trace made before E1, with
// citations only, by the build of 4ebff26 (`wag-llm play` with the scripted
// profiles of testdata/e1-before/profiles.json: game.* stopped after three
// moves, finished.trace.jsonl played to its end), load, replay and resume
// as before; the citations, a source's reference URL and binding ids among
// them, stay citations. (The replay is tested in cmd/wag-llm: a game
// stopped early asks a replay for a call its trace does not have, under any
// build.)

// citedStrings are the citation strings of a game's arguments, and whether
// any argument cites a binding.
func citedStrings(st State) ([]string, bool) {
	var out []string
	refs := false
	for _, a := range st.Arguments {
		out = append(out, a.Evidence...)
		refs = refs || len(a.EvidenceRefs) > 0
	}
	return out, refs
}

func TestY7AGameMadeBeforeE1(t *testing.T) {
	export, err := os.ReadFile(filepath.Join("testdata", "e1-before", "game.export.json"))
	if err != nil {
		t.Fatal(err)
	}
	trace, err := os.ReadFile(filepath.Join("testdata", "e1-before", "game.trace.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	check := func(what string, st State) {
		t.Helper()
		cited, refs := citedStrings(st)
		joined := strings.Join(cited, " ")
		if st.Evidence != nil || refs || !strings.Contains(joined, "https://example.org/agency-report") || !strings.Contains(joined, "game:b2") || !strings.Contains(joined, "game:b4") {
			t.Fatalf("Y7, %s: the record set %v, bindings cited %v, citations %v", what, st.Evidence, refs, cited)
		}
	}
	// The load.
	_, ts := newTestServerWith(t)
	var st State
	if code := call(t, ts, "POST", "/api/load", string(export), &st); code != http.StatusOK {
		t.Fatalf("Y7: the load: %d", code)
	}
	check("the load", st)
	var ex map[string]json.RawMessage
	if code := call(t, ts, "GET", "/api/games/"+st.ID+"/export", nil, &ex); code != http.StatusOK {
		t.Fatal(code)
	}
	if _, ok := ex["record_set"]; ok {
		t.Fatal("Y7: the loaded game's export names a record set")
	}
	if _, ok := ex["evidence"]; ok {
		t.Fatal("Y7: the loaded game's export has evidence records")
	}
	// The resume, from a folder of the data folder, with the profiles it was
	// played with.
	store := t.TempDir()
	dir := filepath.Join(store, "games", "2026-11-27_080012_game1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, b := range map[string][]byte{"export.json": export, "trace.jsonl": trace} {
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	srv, ts2 := newTestServerWith(t)
	if err := srv.SetStore(store); err != nil {
		t.Fatal(err)
	}
	ps, err := llm.ProfilesFrom(filepath.Join("testdata", "e1-before", "profiles.json"), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range ps {
		c, err := llm.NewClient(p)
		if err != nil {
			t.Fatal(err)
		}
		c.Sleep = llm.NoWait
		srv.AddModel(c)
	}
	code, out := resumeHTTP(t, ts2, "games/2026-11-27_080012_game1")
	if code != http.StatusOK {
		t.Fatalf("Y7: the resume: %d %v", code, out["error"])
	}
	var st2 State
	if code := call(t, ts2, "GET", "/api/games/"+out["id"].(string), nil, &st2); code != http.StatusOK || len(st2.Log) != 3 || st2.Over {
		t.Fatalf("Y7: the resumed game: %d, %d moves, over %v", code, len(st2.Log), st2.Over)
	}
	check("the resume", st2)
}
