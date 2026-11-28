package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/yourorg/wag/internal/evidence/evidencetest"
	"github.com/yourorg/wag/internal/llm"
	"github.com/yourorg/wag/internal/server"
)

// E1.6 (docs/e1-evidence.md): the trace of a game with a record set,
// replayed. Each test's outcome is the plan's, written before the tests ran.

// evidenceTrace plays the fixture's inquiry of two model seats, A and B, on
// a server with the evidence folder dir (A argues a1, B passes, A offers the
// quotation for a1's premise, B consents, both pass), and saves its trace.
func evidenceTrace(t *testing.T, dir string) string {
	t.Helper()
	d, err := server.ReadData("")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := server.New(d, "")
	if err != nil {
		t.Fatal(err)
	}
	srv.SetEvidence(dir)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	var id string
	rev := func() string {
		res, err := http.Get(ts.URL + "/api/games/" + id)
		if err != nil {
			return err.Error()
		}
		defer res.Body.Close()
		var st struct {
			Evidence struct {
				Bindings []struct {
					ID       string `json:"id"`
					Revision string `json:"revision"`
				} `json:"bindings"`
			} `json:"evidence"`
		}
		_ = json.NewDecoder(res.Body).Decode(&st)
		for _, b := range st.Evidence.Bindings {
			return b.Revision[:12]
		}
		return ""
	}
	replies := []func() string{
		func() string {
			return `{"summary":"The claim.","action":{"type":"support","scheme_id":"AS:consequences","variable_bindings":{"A":"tax meat","G":"lower emissions"},"declare":[{"id":"tax meat","actor":"decider"}]}}`
		},
		func() string { return `{"summary":"Nothing yet.","action":{"type":"abstain"}}` },
		func() string {
			return `{"summary":"Offer the report.","action":{"type":"evidence_offer","offer":{"claim_id":"claim-quote","argument_id":"a1","kind":"premise","statement":"brings_about(tax_meat,lower_emissions)","role":"support","interpretation":"The report finds it.","warrant":"An agency's report is a defeasible indicator."}}}`
		},
		func() string {
			return fmt.Sprintf(`{"summary":"Consent.","action":{"type":"evidence_consent","binding_id":"game:b2","revision":%q}}`, rev())
		},
		func() string { return `{"summary":"Done.","action":{"type":"abstain"}}` },
	}
	var mu sync.Mutex
	n := 0
	c := llm.WithProvider(llm.Profile{Name: "m", Provider: "mock", Params: llm.Params{Model: "mock-1"}}, &llm.Mock{Fn: func(llm.Request) (*llm.Response, error) {
		mu.Lock()
		f := replies[min(n, len(replies)-1)]
		n++
		mu.Unlock()
		return &llm.Response{Text: f(), FinishReason: "stop", Model: "mock-1"}, nil
	}})
	c.Sleep = llm.NoWait
	srv.AddModel(c)
	rounds := 6
	req := server.NewGameRequest{Mode: "inquiry", Seats: []server.Seat{{Name: "A", Kind: "model", Model: "m"}, {Name: "B", Kind: "model", Model: "m"}},
		MaxRounds: &rounds, AllowOther: true, Seed: 7, Evidence: "fixture"}
	var st stateView
	if err := call(ts.URL, "POST", "/api/games", req, &st); err != nil {
		t.Fatal(err)
	}
	id = st.ID
	g := &game{base: ts.URL, id: st.ID}
	if err := g.run(20); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "game.trace.jsonl")
	if _, err := g.save(path, "traces"); err != nil {
		t.Fatal(err)
	}
	return path
}

// Y5: the trace replayed (wag-llm replay --evidence DIR): it matches. Y6:
// replayed without --evidence: refused, naming the flag. Y9, for a replay:
// other pins than this build's: refused, naming the pins.
func TestY5Y6Y9TheReplayOfAGameWithARecordSet(t *testing.T) {
	dir := t.TempDir()
	if err := evidencetest.Write(dir, "fixture", evidencetest.Set()); err != nil {
		t.Fatal(err)
	}
	path := evidenceTrace(t, dir)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"type":"evidence_offer"`) || !strings.Contains(string(b), `"type":"evidence_consent"`) || !strings.Contains(string(b), `"record_set":{"name":"fixture"`) {
		t.Fatal("the trace has no evidence acts or no record set")
	}
	ctx := context.Background()
	if err := replay(ctx, []string{path, "--evidence", dir}); err != nil {
		t.Fatalf("Y5: %v", err)
	}
	if err := replay(ctx, []string{path}); err == nil || !strings.Contains(err.Error(), "--evidence") {
		t.Fatalf("Y6: %v", err)
	}
	other := editStart(t, path, func(m map[string]any) {
		m["pins"].(map[string]any)["engine_sha256"] = strings.Repeat("e", 64)
	})
	if err := replay(ctx, []string{other, "--evidence", dir}); err == nil || !strings.Contains(err.Error(), "pins") {
		t.Fatalf("Y9: %v", err)
	}
	// And a changed record set: its SHA-256.
	set := evidencetest.Set()
	set["claims"].([]any)[1].(map[string]any)["text"] = "The agency found that a meat tax cuts emissions sharply."
	changed := t.TempDir()
	if err := evidencetest.Write(changed, "fixture", set); err != nil {
		t.Fatal(err)
	}
	if err := replay(ctx, []string{path, "--evidence", changed}); err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("the changed record set: %v", err)
	}
}

// beforeC53 is the frozen copy of the data the fixtures of
// internal/server/testdata/e1-before were made on (092b6a6f48a672d2), which
// the switch of step C5.3 changed: they replay with --data on it.
var beforeC53 = filepath.Join("..", "..", "data", "frozen", "092b6a6f48a672d2")

// Y7: a trace made before E1 (citations only, by the build of 4ebff26, a
// game played to its end) replays on this build, with the data it was
// played with.
func TestY7ATraceMadeBeforeE1Replays(t *testing.T) {
	trace := filepath.Join("..", "..", "internal", "server", "testdata", "e1-before", "finished.trace.jsonl")
	if err := replay(context.Background(), []string{trace, "--data", beforeC53}); err != nil {
		t.Fatalf("Y7: %v", err)
	}
}

// X7 (docs/e1-evidence.md, E1.7, B3; the outcome written before the test
// ran): a debate of model seats whose replies carry the evidence fields
// (`{"type":"withdraw"}`, a stray "revision": 2, a "binding_id", an "offer"
// and "evidence_refs"), recorded by the build of 4ebff26, before E1: it
// replays on this build, `withdraw` refused with S0's unknown action type
// and the others played without those fields, as that build did.
func TestX7ADebateWithStrayEvidenceFieldsReplays(t *testing.T) {
	trace := filepath.Join("..", "..", "internal", "server", "testdata", "e1-before", "stray.trace.jsonl")
	b, err := os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`\"revision\":2`, `\"binding_id\":\"game:b1\"`, `\"evidence_refs\":[\"game:b2\"]`, `S0: unknown action type \"withdraw\"`} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("the trace has no %s", want)
		}
	}
	if err := replay(context.Background(), []string{trace, "--data", beforeC53}); err != nil {
		t.Fatalf("X7: %v", err)
	}
}
