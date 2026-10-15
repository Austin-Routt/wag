package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A headless game with a scripted model and a rule bot writes its trace
// file, export and summary; replaying the trace file gives the same game;
// a trace whose recorded answer was changed does not replay.
func TestPlayAndReplay(t *testing.T) {
	dir := t.TempDir()
	prof := filepath.Join(dir, "profiles.json")
	os.WriteFile(prof, []byte(`{"profiles":[{"name":"scripted","provider":"mock","model":"mock-1","replies":[
		"Here is my case. {\"summary\":\"State the claim.\",\"action\":{\"type\":\"support\",\"scheme_id\":\"AS:consequences\",\"variable_bindings\":{\"A\":\"tax meat\",\"G\":\"lower emissions\"}}}",
		"{\"type\":\"support\",\"scheme_id\":\"AS:nonexistent\"}",
		"{\"summary\":\"Nothing more.\",\"action\":{\"type\":\"abstain\"}}"]}]}`), 0o644)
	out := filepath.Join(dir, "out")
	if err := play(context.Background(), []string{"--models", prof, "--pro", "scripted", "--con", "greedy", "--out", out, "--max-rounds", "4"}); err != nil {
		t.Fatal(err)
	}
	traces, _ := filepath.Glob(filepath.Join(out, "*.trace.jsonl"))
	exports, _ := filepath.Glob(filepath.Join(out, "*.export.json"))
	sum, _ := os.ReadFile(filepath.Join(out, "summary.jsonl"))
	if len(traces) != 1 || len(exports) != 1 || !strings.Contains(string(sum), `"rejected":1`) || !strings.Contains(string(sum), `"winner_side"`) {
		t.Fatalf("outputs: %v %v %s", traces, exports, sum)
	}
	if err := replay(context.Background(), []string{traces[0]}); err != nil {
		t.Fatalf("replay: %v", err)
	}
	// The answer recorded in the first call (not the profile's script in
	// game_start, which a replay does not use).
	b, _ := os.ReadFile(traces[0])
	lines := strings.Split(string(b), "\n")
	for i, l := range lines {
		if strings.Contains(l, `"type":"call"`) {
			lines[i] = strings.Replace(l, `\"G\":\"lower emissions\"`, `\"G\":\"higher emissions\"`, 1)
			if lines[i] == l {
				t.Fatal("nothing to tamper with")
			}
			break
		}
	}
	tampered := strings.Join(lines, "\n")
	bad := filepath.Join(dir, "tampered.jsonl")
	os.WriteFile(bad, []byte(tampered), 0o644)
	if err := replay(context.Background(), []string{bad}); err == nil {
		t.Fatal("a changed answer replays as the same game")
	}
	if err := report([]string{out}); err != nil {
		t.Fatal(err)
	}
}

// annotate writes the annotations in the annotators' format, retries a
// reply that is not JSON once, traces every call, and resumes.
func TestAnnotate(t *testing.T) {
	dir := t.TempDir()
	items := filepath.Join(dir, "items.json")
	os.WriteFile(items, []byte(`{"discussion": {"thesis_claim_id": "1.1"}, "items": [
		{"kialo_claim_id": "1.2", "text": "Meat is bad for the climate."},
		{"kialo_claim_id": "1.3", "text": "Taxes hurt the poor."}]}`), 0o644)
	guide := filepath.Join(dir, "guidelines.md")
	os.WriteFile(guide, []byte("# Guidelines\nAnnotate."), 0o644)
	ctxFile := filepath.Join(dir, "walton.yml")
	os.WriteFile(ctxFile, []byte("argument_schemes: []"), 0o644)
	// The mock's script starts again with each run.
	first, second := filepath.Join(dir, "p1.json"), filepath.Join(dir, "p2.json")
	os.WriteFile(first, []byte(`[{"name": "ann", "provider": "mock", "model": "m", "replies": [
		"{\"kialo_claim_id\": \"1.2\", \"instances\": [{\"primary\": true, \"scheme\": {\"walton_id\": \"negative_consequences\", \"fit\": \"reconstructed\"}, \"premises\": [], \"conclusion\": {\"formula\": \"¬should_be_performed(tax_meat)\"}}]}"]}]`), 0o644)
	os.WriteFile(second, []byte(`[{"name": "ann", "provider": "mock", "model": "m", "replies": [
		"I am not sure.",
		"{\"instances\": [{\"primary\": true, \"scheme\": {\"walton_id\": \"none\", \"fit\": \"none\"}}]}"]}]`), 0o644)
	out := filepath.Join(dir, "out")
	args := []string{"--items", items, "--guidelines", guide, "--context", ctxFile, "--out", out}
	if err := annotate(context.Background(), append(args, "--models", first, "--limit", "1")); err != nil {
		t.Fatal(err)
	}
	if err := annotate(context.Background(), append(args, "--models", second)); err != nil { // resumes with 1.3
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(out, "annotator-ann.json"))
	if !strings.Contains(string(b), `"kialo_claim_id": "1.2"`) || !strings.Contains(string(b), `"kialo_claim_id": "1.3"`) {
		t.Fatalf("annotations: %s", b)
	}
	calls, _ := os.ReadFile(filepath.Join(out, "calls.jsonl"))
	if n := strings.Count(string(calls), "\n"); n != 3 {
		t.Fatalf("%d calls traced, want 3 (1.2; 1.3 and its retry)", n)
	}
	run, _ := os.ReadFile(filepath.Join(out, "run.json"))
	if !strings.Contains(string(run), `"annotated": 2`) || !strings.Contains(string(run), "system_prompt_sha256") {
		t.Fatalf("run.json: %s", run)
	}
}
