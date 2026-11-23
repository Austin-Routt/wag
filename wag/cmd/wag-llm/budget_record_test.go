package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/wag/internal/llm"
)

const recordReply = `{"instances": [{"primary": true, "scheme": {"walton_id": "none", "fit": "none"}}]}`

// keyedServer is a local OpenAI-compatible endpoint (no live provider), with
// a dummy key: it counts its calls and answers an annotation with usage.
func keyedServer(t *testing.T) (*httptest.Server, *int64) {
	t.Helper()
	var hits int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"r","model":"gpt-x","choices":[{"message":{"content":"`+strings.ReplaceAll(recordReply, `"`, `\"`)+
			`"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("OPENAI_API_KEY", "dummy-test-key")
	return srv, &hits
}

// annotateFixture writes a dataset of n claims and two profiles: k, keyed
// (the local server above), and l, a mock that reports no usage.
func annotateFixture(t *testing.T, srvURL string, n int) (prof, data, out string) {
	t.Helper()
	dir := t.TempDir()
	data = filepath.Join(dir, "d")
	if err := os.MkdirAll(data, 0o755); err != nil {
		t.Fatal(err)
	}
	var items []string
	for i := 0; i < n; i++ {
		items = append(items, fmt.Sprintf(`{"kialo_claim_id": "1.%d", "text": "Claim %d."}`, i+2, i+2))
	}
	if err := os.WriteFile(filepath.Join(data, "items.json"), []byte(`{"discussion": {"thesis_claim_id": "1.1"}, "items": [`+strings.Join(items, ",")+`]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	q := strings.ReplaceAll(recordReply, `"`, `\"`)
	prof = filepath.Join(dir, "p.json")
	if err := os.WriteFile(prof, []byte(`[{"name": "k", "provider": "openai", "model": "gpt-x", "base_url": "`+srvURL+`", "retries": -1},
		{"name": "l", "provider": "mock", "model": "m", "replies": ["`+q+`"]}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	return prof, data, filepath.Join(dir, "run")
}

func budgetLines(t *testing.T, dir string) []budgetRecord {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "budget.jsonl"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	var out []budgetRecord
	for _, l := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if l == "" {
			continue
		}
		d := json.NewDecoder(strings.NewReader(l))
		d.DisallowUnknownFields()
		var r budgetRecord
		if err := d.Decode(&r); err != nil {
			t.Fatalf("%s: %v", l, err)
		}
		out = append(out, r)
	}
	return out
}

// stderrOf runs f with os.Stderr captured.
func stderrOf(t *testing.T, f func()) string {
	t.Helper()
	tmp, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stderr
	os.Stderr = tmp
	func() {
		defer func() { os.Stderr = old }()
		f()
	}()
	tmp.Close()
	b, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// D046 as the owner amended it (2026-11-23): with no flag a keyed run has
// no limit, and says so once; --unlimited says it explicitly and silences
// the note; either way the calls and the reported usage are counted and
// recorded in the run's folder, one line per invocation.
func TestAnnotateUnlimitedByDefault(t *testing.T) {
	srv, hits := keyedServer(t)
	prof, data, out := annotateFixture(t, srv.URL, 3)
	var err error
	stderr := stderrOf(t, func() {
		err = annotateCmd(context.Background(), []string{"--models", prof, "--dataset", data, "--out", out})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr, "with no limit, D046's default") || !strings.Contains(stderr, "profile k ") || strings.Contains(stderr, "profile l") {
		t.Fatalf("the note names the keyed profile, and only it:\n%s", stderr)
	}
	recs := budgetLines(t, out)
	if len(recs) != 1 || !recs[0].Unlimited || !recs[0].Default || recs[0].Command != "annotate" || recs[0].Refused ||
		recs[0].Spent.Calls != 6 || recs[0].Spent.Tokens != 3*15 || recs[0].Spent.Unreported != 3 || *hits != 3 {
		t.Fatalf("record %+v, keyed calls %d; want unlimited by default, 6 calls (3 keyed with 15 tokens each, 3 mock without usage)", recs, *hits)
	}
	if _, err := time.Parse(time.RFC3339, recs[0].Started); err != nil || recs[0].Ended < recs[0].Started {
		t.Fatalf("times %q, %q: %v", recs[0].Started, recs[0].Ended, err)
	}
	// --unlimited: no note, and the record says it was explicit. The run is
	// finished, so nothing is called.
	stderr = stderrOf(t, func() {
		err = annotateCmd(context.Background(), []string{"--models", prof, "--out", out, "--unlimited"})
	})
	if err != nil || strings.Contains(stderr, "D046's default") || *hits != 3 {
		t.Fatalf("err %v, keyed calls %d, stderr:\n%s", err, *hits, stderr)
	}
	// A contradiction is refused before anything runs or is recorded.
	if err := annotateCmd(context.Background(), []string{"--models", prof, "--out", out, "--unlimited", "--max-calls", "1"}); err == nil || !strings.Contains(err.Error(), "--unlimited") {
		t.Fatalf("--unlimited with a limit: %v", err)
	}
	recs = budgetLines(t, out)
	if len(recs) != 2 || !recs[1].Unlimited || recs[1].Default || recs[1].Spent.Calls != 0 {
		t.Fatalf("records %+v", recs)
	}
	readme, _ := os.ReadFile(filepath.Join(out, "README.txt"))
	if !strings.Contains(string(readme), "budget.jsonl") {
		t.Fatal("the run's README does not list budget.jsonl")
	}
}

// A continued run calls the annotators saved in it, not this invocation's
// --annotators: the note names them, and a limit applies to every client.
func TestAnnotateContinueUsesSavedProfiles(t *testing.T) {
	srv, hits := keyedServer(t)
	prof, data, out := annotateFixture(t, srv.URL, 4)
	// A keyed reviewer, r, saved in the run too.
	raw, err := os.ReadFile(prof)
	if err != nil {
		t.Fatal(err)
	}
	withReviewer := strings.Replace(string(raw), `[{"name": "k"`, `[{"name": "r", "provider": "openai", "model": "gpt-x", "base_url": "`+srv.URL+`", "retries": -1},
		{"name": "k"`, 1)
	if err := os.WriteFile(prof, []byte(withReviewer), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := annotateCmd(context.Background(), []string{"--models", prof, "--dataset", data, "--out", out, "--reviewer", "r", "--max-calls", "2"}); err != nil {
		t.Fatal(err)
	}
	before := *hits
	stderr := stderrOf(t, func() {
		err = annotateCmd(context.Background(), []string{"--models", prof, "--out", out, "--annotators", "l"})
	})
	if err != nil || !strings.Contains(stderr, "profiles k, r use an API key") || *hits == before {
		t.Fatalf("err %v, keyed calls %d -> %d; the note must name the saved keyed annotator and reviewer:\n%s", err, before, *hits, stderr)
	}
	recs := budgetLines(t, out)
	if len(recs) != 2 || recs[0].Unlimited || recs[0].MaxCalls != 2 || recs[0].Spent.Calls != 2 || !recs[0].Refused ||
		!recs[1].Unlimited || !recs[1].Default || recs[1].Spent.Calls == 0 {
		t.Fatalf("records %+v", recs)
	}
}

// play records its invocation's budget too, and its calls are the trace's.
func TestPlayRecordsItsBudget(t *testing.T) {
	dir := t.TempDir()
	prof := filepath.Join(dir, "profiles.json")
	if err := os.WriteFile(prof, []byte(`{"profiles":[{"name":"scripted","provider":"mock","model":"mock-1","replies":["{\"action\":{\"type\":\"abstain\"}}"]}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out")
	var err error
	stderr := stderrOf(t, func() {
		err = play(context.Background(), []string{"--models", prof, "--pro", "scripted", "--con", "greedy", "--out", out, "--games", "1", "--max-rounds", "2"})
	})
	if err != nil {
		t.Fatal(err)
	}
	recs := budgetLines(t, out)
	sum, _ := os.ReadFile(filepath.Join(out, "summary.jsonl"))
	var s map[string]any
	if err := json.Unmarshal(sum, &s); err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 || !recs[0].Unlimited || !recs[0].Default || recs[0].Command != "play" || recs[0].Spent.Calls == 0 ||
		float64(recs[0].Spent.Calls) != s["calls"].(float64) {
		t.Fatalf("records %+v; summary %s", recs, sum)
	}
	if strings.Contains(stderr, "D046's default") {
		t.Fatalf("a mock run needs no note:\n%s", stderr)
	}
}

func flagsFrom(t *testing.T, args ...string) budgetFlags {
	t.Helper()
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	f := addBudgetFlags(fs)
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	return f
}

// The note is for a keyed profile run with no limit by default; not with
// --unlimited, not with a limit, not for mock or local profiles.
func TestBudgetNote(t *testing.T) {
	keyed := []llm.Profile{
		{Name: "a", Provider: "anthropic"},
		{Name: "c", Provider: "openai-compatible", BaseURL: "http://localhost:1/v1", APIKeyEnv: "UNUSED_BUDGET_TEST_KEY"},
		{Name: "o", Provider: "ollama", BaseURL: "https://example.invalid/v1", APIKey: "not-a-real-key"},
	}
	local := []llm.Profile{{Name: "m", Provider: "mock"}, {Name: "l", Provider: "ollama"}, {Name: "x", Provider: "command"}}
	for _, tc := range []struct {
		args []string
		used []llm.Profile
		want string
	}{
		{nil, append(append([]llm.Profile{}, local...), keyed...), "profiles a, c, o"},
		{nil, keyed[:1], "profile a "},
		{nil, local, ""},
		{[]string{"--unlimited"}, keyed, ""},
		{[]string{"--max-calls", "3"}, keyed, ""},
		{[]string{"--max-tokens", "300"}, keyed, ""},
	} {
		f := flagsFrom(t, tc.args...)
		b, err := f.budget()
		if err != nil {
			t.Fatal(err)
		}
		var w bytes.Buffer
		f.note(&w, b, tc.used)
		if got := w.String(); (tc.want == "") != (got == "") || !strings.Contains(got, tc.want) {
			t.Errorf("%v: note %q, want one with %q", tc.args, got, tc.want)
		}
	}
}

// The record names the price table by its file name and SHA-256, not its
// folder, and is not written where the invocation made no folder.
func TestBudgetRecordPrices(t *testing.T) {
	dir := t.TempDir()
	prices := filepath.Join(dir, "rates.json")
	raw := []byte(`{"*":{"input":1,"output":2}}`)
	if err := os.WriteFile(prices, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	f := flagsFrom(t, "--prices", prices)
	b, err := f.budget()
	if err != nil || b.Limited() {
		t.Fatalf("prices alone are no limit: %v", err)
	}
	var w bytes.Buffer
	f.finish(&w, dir, "play", b, time.Now())
	sum := sha256.Sum256(raw)
	recs := budgetLines(t, dir)
	if len(recs) != 1 || recs[0].Prices != "rates.json" || recs[0].PricesSHA256 != hex.EncodeToString(sum[:]) || !recs[0].Unlimited || !recs[0].Default {
		t.Fatalf("records %+v", recs)
	}
	if !strings.Contains(w.String(), "no limit") {
		t.Fatalf("report %q", w.String())
	}
	gone := filepath.Join(dir, "absent")
	w.Reset()
	f.finish(&w, gone, "play", b, time.Now())
	if _, err := os.Stat(gone); !os.IsNotExist(err) || strings.Contains(w.String(), "not written") {
		t.Fatalf("finish without a folder: %v, %q", err, w.String())
	}
	// Invalid prices are refused even without a cost limit.
	if err := os.WriteFile(prices, []byte(`{"*":{"input":-1,"output":2}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := flagsFrom(t, "--prices", prices).budget(); err == nil {
		t.Fatal("a negative rate was accepted without --max-cost")
	}
}

// play with a keyed profile (a local server, a dummy key): the note names it,
// and the record counts its calls and the tokens it reported.
func TestPlayNotesKeyedProfiles(t *testing.T) {
	var hits int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"r","model":"gpt-x","choices":[{"message":{"content":"{\"action\":{\"type\":\"abstain\"}}"},"finish_reason":"stop"}],`+
			`"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("OPENAI_API_KEY", "dummy-test-key")
	dir := t.TempDir()
	prof := filepath.Join(dir, "p.json")
	// u uses a key too, but has no seat: the note does not name it.
	if err := os.WriteFile(prof, []byte(`[{"name": "k", "provider": "openai", "model": "gpt-x", "base_url": "`+srv.URL+`", "retries": -1},
		{"name": "u", "provider": "openai", "model": "gpt-y", "base_url": "`+srv.URL+`", "retries": -1}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out")
	var err error
	stderr := stderrOf(t, func() {
		err = play(context.Background(), []string{"--models", prof, "--pro", "k", "--con", "greedy", "--out", out, "--games", "1", "--max-rounds", "1"})
	})
	if strings.Contains(stderr, "profiles") || strings.Contains(stderr, " u ") {
		t.Fatalf("the note names a profile with no seat:\n%s", stderr)
	}
	if err != nil {
		t.Fatal(err)
	}
	recs := budgetLines(t, out)
	if !strings.Contains(stderr, "profile k uses an API key") || len(recs) != 1 || recs[0].Spent.Calls == 0 ||
		int64(recs[0].Spent.Calls) != atomic.LoadInt64(&hits) || recs[0].Spent.Tokens != 15*recs[0].Spent.Calls || !recs[0].Default {
		t.Fatalf("records %+v, server calls %d, stderr:\n%s", recs, hits, stderr)
	}
}

// The record keeps every limit, and an accounting violation as a halt, not
// as a refused call.
func TestBudgetRecordLimitsAndHalt(t *testing.T) {
	dir := t.TempDir()
	prices := filepath.Join(dir, "prices.json")
	if err := os.WriteFile(prices, []byte(`{"*":{"input":1,"output":1}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	f := flagsFrom(t, "--max-tokens", "1000", "--max-cost", "1", "--prices", prices)
	b, err := f.budget()
	if err != nil {
		t.Fatal(err)
	}
	// The provider reports far more input than the estimate: an overrun.
	c := llm.WithProvider(llm.Profile{Name: "p", Provider: "mock", Params: llm.Params{MaxOutputTokens: 10}},
		&llm.Mock{Fn: func(llm.Request) (*llm.Response, error) {
			n := 50000
			return &llm.Response{Text: "ok", Usage: llm.Usage{InputTokens: &n}}, nil
		}})
	c.Sleep, c.Budget = llm.NoWait, b
	if _, err := c.Complete(context.Background(), c.Request("s", []llm.Message{{Role: "user", Content: "x"}}), nil); err == nil {
		t.Fatal("the overrun was not reported")
	}
	var w bytes.Buffer
	f.finish(&w, dir, "play", b, time.Now())
	recs := budgetLines(t, dir)
	if len(recs) != 1 || recs[0].Unlimited || recs[0].Default || recs[0].MaxTokens != 1000 || recs[0].MaxCost != 1 || recs[0].MaxCalls != 0 ||
		recs[0].Refused || recs[0].Halted == "" || recs[0].Spent.Calls != 1 {
		t.Fatalf("records %+v", recs)
	}
	if !strings.Contains(w.String(), "stopped:") {
		t.Fatalf("the report does not say the invocation stopped: %q", w.String())
	}
	// An empty price table is named as such.
	empty := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(empty, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := flagsFrom(t, "--max-cost", "1", "--prices", empty).budget(); err == nil || !strings.Contains(err.Error(), "has none") {
		t.Fatalf("an empty price table: %v", err)
	}
}
