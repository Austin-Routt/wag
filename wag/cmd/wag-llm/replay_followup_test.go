package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/llm"
	"github.com/yourorg/wag/internal/server"
)

// Follow-ups to the replay-identity patch (verify/r-critical-review-2026-11-23,
// report A); each fails on the patch as received, except the legacy trace,
// whose path it left untested.

var scripts = map[string][]string{
	"scripted-a": {`{"summary":"State the claim.","action":{"type":"support","scheme_id":"AS:consequences","variable_bindings":{"A":"tax meat","G":"lower emissions"},"declare":[{"id":"tax meat","actor":"decider"}]}}`,
		`{"summary":"Nothing more.","action":{"type":"abstain"}}`},
	"scripted-b": {`{"summary":"Nothing to add.","action":{"type":"abstain"}}`},
}

// traceOf plays a game between the seats, on a server with the scripted
// profiles, and saves its trace file.
func traceOf(t *testing.T, seats []server.Seat, contextID string) string {
	t.Helper()
	d, err := server.ReadData("")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := server.New(d, "")
	if err != nil {
		t.Fatal(err)
	}
	for name, replies := range scripts {
		c, err := llm.NewClient(llm.Profile{Name: name, Provider: "mock", Params: llm.Params{Model: "mock-1"}, Replies: replies})
		if err != nil {
			t.Fatal(err)
		}
		c.Sleep = llm.NoWait
		srv.AddModel(c)
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	rounds := 4
	req := server.NewGameRequest{Mode: "debate", Seats: seats, Context: contextID, MaxRounds: &rounds, AllowOther: true, Seed: 7}
	var st stateView
	if err := call(ts.URL, "POST", "/api/games", req, &st); err != nil {
		t.Fatal(err)
	}
	g := &game{base: ts.URL, id: st.ID}
	if err := g.run(40); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "game.trace.jsonl")
	if _, err := g.save(path, "traces"); err != nil {
		t.Fatal(err)
	}
	return path
}

// editStart rewrites the game_start event of a trace file into a new one.
func editStart(t *testing.T, path string, edit func(map[string]any)) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")
	var start map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &start); err != nil {
		t.Fatal(err)
	}
	edit(start)
	x, _ := json.Marshal(start)
	lines[0] = string(x)
	out := filepath.Join(t.TempDir(), "edited.trace.jsonl")
	if err := os.WriteFile(out, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	return out
}

// A player's calls are recorded under its ID, which is not its seat's name
// for a seat named after the moderator ("moderator#2") or a second seat of
// the same name ("Ann#2"): the replay finds them by ID.
func TestFollowupReplayFindsCallsByPlayerID(t *testing.T) {
	for name, seats := range map[string][]server.Seat{
		"a model seat named moderator": {{Name: "moderator", Kind: "model", Model: "scripted-a"}, {Name: "Con", Kind: "greedy"}},
		"two model seats named Ann":    {{Name: "Ann", Kind: "model", Model: "scripted-a"}, {Name: "Ann", Kind: "model", Model: "scripted-b"}},
	} {
		trace := traceOf(t, seats, "")
		b, _ := os.ReadFile(trace)
		if !strings.Contains(string(b), `"player":"`+seats[0].Name+`#2"`) && !strings.Contains(string(b), `"player":"`+seats[1].Name+`#2"`) {
			t.Fatalf("%s: fixture: no player has an ID of its own", name)
		}
		if err := replay(context.Background(), []string{trace}); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		// Every seat's replay is checked: a call recorded for the first
		// player that the replay does not make is reported.
		var extra string
		for _, l := range strings.Split(string(b), "\n") {
			if strings.Contains(l, `"type":"call"`) && strings.Contains(l, `"player":"`+seats[0].Name+`"`) {
				extra = l
			}
		}
		if extra == "" {
			continue // the moderator-named seat's calls are under moderator#2
		}
		tampered := filepath.Join(t.TempDir(), "extra.trace.jsonl")
		if err := os.WriteFile(tampered, []byte(strings.TrimRight(string(b), "\n")+"\n"+extra+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := replay(context.Background(), []string{tampered}); err == nil || !strings.Contains(err.Error(), "differs from the trace in 1 ways") {
			t.Errorf("%s: a recorded call not made: %v", name, err)
		}
	}
}

// A trace without a fingerprint (none is made now) replays, with a note.
func TestFollowupLegacyTraceReplaysWithANote(t *testing.T) {
	trace := editStart(t, traceOf(t, []server.Seat{{Name: "Pro", Kind: "model", Model: "scripted-a"}, {Name: "Con", Kind: "greedy"}}, ""),
		func(start map[string]any) { delete(start, "data_sha256") })
	var err error
	note := stderrOf(t, func() { err = replay(context.Background(), []string{trace}) })
	if err != nil || !strings.Contains(note, "no scheme-data fingerprint") {
		t.Fatalf("%v; stderr %q", err, note)
	}
}

// A trace made by a build from a Windows checkout before the fingerprint
// read CRLF as LF recorded the fingerprint of the CRLF bytes: the same data.
func TestFollowupReplayAcceptsAWindowsFingerprint(t *testing.T) {
	conv := func(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\n"), []byte("\r\n")) }
	h := sha256.New()
	for _, b := range [][]byte{data.CoreYAML, data.SchemesJSON, data.CQsJSON, data.EdgesJSON} {
		h.Write(conv(b))
		h.Write([]byte{0})
	}
	windows := hex.EncodeToString(h.Sum(nil))[:16]
	trace := editStart(t, traceOf(t, []server.Seat{{Name: "Pro", Kind: "model", Model: "scripted-a"}, {Name: "Con", Kind: "greedy"}}, ""),
		func(start map[string]any) { start["data_sha256"] = windows })
	if err := replay(context.Background(), []string{trace}); err != nil {
		t.Fatal(err)
	}
}

// A context preset can change without changing the data's fingerprint: the
// replay uses the presets the trace recorded, and refuses a replay whose
// weights would differ when the trace recorded none.
func TestFollowupReplayKeepsTheRecordedPreset(t *testing.T) {
	trace := traceOf(t, []server.Seat{{Name: "Pro", Kind: "model", Model: "scripted-a"}, {Name: "Con", Kind: "greedy"}}, "scientific")
	dir := t.TempDir()
	var f map[string]any
	if err := json.Unmarshal(data.ContextsJSON, &f); err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, c := range f["contexts"].([]any) {
		if c := c.(map[string]any); c["id"] == "scientific" {
			c["weights"].(map[string]any)["AS:consequences"] = 0.9
			changed = true
		}
	}
	if !changed {
		t.Fatal("fixture: no scientific preset")
	}
	contexts, _ := json.Marshal(f)
	for name, b := range map[string][]byte{"schemes/core.yml": data.CoreYAML, "nascq/schemes.json": data.SchemesJSON,
		"nascq/critical_questions.json": data.CQsJSON, "nascq/edges.json": data.EdgesJSON, "contexts.json": contexts} {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := replay(context.Background(), []string{trace, "--data", dir}); err != nil {
		t.Fatalf("with the recorded presets: %v", err)
	}
	unrecorded := editStart(t, trace, func(start map[string]any) { delete(start, "contexts") })
	if err := replay(context.Background(), []string{unrecorded, "--data", dir}); err == nil || !strings.Contains(err.Error(), "weights") {
		t.Fatalf("a changed preset with none recorded: %v", err)
	}
	if err := replay(context.Background(), []string{unrecorded}); err != nil {
		t.Fatalf("the preset unchanged, none recorded: %v", err)
	}
	// Recorded presets are checked as the presets of a data folder are.
	invalid := editStart(t, trace, func(start map[string]any) {
		for _, c := range start["contexts"].([]any) {
			if c := c.(map[string]any); c["id"] == "scientific" {
				c["weights"].(map[string]any)["AS:consequences"] = 2
			}
		}
	})
	if err := replay(context.Background(), []string{invalid}); err == nil || !strings.Contains(err.Error(), `preset "scientific": AS:consequences weighs 2`) {
		t.Fatalf("an invalid recorded preset: %v", err)
	}
}

// The trace records its presets, but they do not give the game's weights (a
// game loaded from an export keeps the weights it was played with): the
// refusal says so, rather than that the trace records no presets.
func TestFollowupReplaySaysWhyTheWeightsDiffer(t *testing.T) {
	trace := traceOf(t, []server.Seat{{Name: "Pro", Kind: "model", Model: "scripted-a"}, {Name: "Con", Kind: "greedy"}}, "scientific")
	edited := editStart(t, trace, func(start map[string]any) {
		start["config"].(map[string]any)["Weights"].(map[string]any)["AS:consequences"] = 0.9
	})
	if err := replay(context.Background(), []string{edited}); err == nil || !strings.Contains(err.Error(), "the presets the trace records do not give them") {
		t.Fatal(err)
	}
}
