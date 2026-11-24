package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A run whose annotations cannot be written stops after the call, says
// what is not saved, and when it is continued takes the reply from
// calls.jsonl instead of asking again.
func TestAnnotateSaysWhatIsNotSaved(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "discussion")
	os.MkdirAll(data, 0o755)
	os.WriteFile(filepath.Join(data, "items.json"), []byte(`{"discussion": {"thesis_claim_id": "1.1"}, "items": [
		{"kialo_claim_id": "1.2", "text": "Meat is bad for the climate."},
		{"kialo_claim_id": "1.3", "text": "Taxes hurt the poor."}]}`), 0o644)
	prof := filepath.Join(dir, "p.json")
	os.WriteFile(prof, []byte(`[{"name": "ann", "provider": "mock", "model": "m", "replies": [
		"{\"instances\": [{\"primary\": true, \"scheme\": {\"walton_id\": \"none\", \"fit\": \"none\"}}]}"]}]`), 0o644)
	out := filepath.Join(dir, "run")
	blocked := filepath.Join(out, "annotations", "ann.json.tmp")
	if err := os.MkdirAll(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
	var err error
	msg := stderrOf(t, func() {
		err = annotateCmd(context.Background(), []string{"--models", prof, "--dataset", data, "--out", out})
	})
	if err == nil || !strings.Contains(err.Error(), "could not write annotations/ann.json") ||
		!strings.Contains(msg, "Not saved: annotations of ann: 1.2") {
		t.Fatalf("%v; stderr %q", err, msg)
	}
	if err := os.Remove(blocked); err != nil {
		t.Fatal(err)
	}
	if err := annotateCmd(context.Background(), []string{"--models", prof, "--out", out}); err != nil {
		t.Fatal(err)
	}
	calls, _ := os.ReadFile(filepath.Join(out, "calls.jsonl"))
	ann, _ := os.ReadFile(filepath.Join(out, "annotations", "ann.json"))
	if n := strings.Count(string(calls), "\n"); n != 2 || !strings.Contains(string(ann), `"1.2"`) || !strings.Contains(string(ann), `"1.3"`) {
		t.Fatalf("%d calls; %s", n, ann)
	}
}

// An annotator whose replies keep failing the check is stopped, and the
// command says why (the reviewer's rehearsal of 24 November 2026).
func TestAnnotateSaysWhyAnAnnotatorStopped(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "discussion")
	os.MkdirAll(data, 0o755)
	var items []string
	for i := 2; i <= 6; i++ {
		items = append(items, `{"kialo_claim_id": "1.`+string(rune('0'+i))+`", "text": "Claim `+string(rune('0'+i))+`."}`)
	}
	os.WriteFile(filepath.Join(data, "items.json"), []byte(`{"discussion": {"thesis_claim_id": "1.1"}, "items": [`+strings.Join(items, ",")+`]}`), 0o644)
	prof := filepath.Join(dir, "p.json")
	os.WriteFile(prof, []byte(`[{"name": "ann", "provider": "mock", "model": "m", "replies": ["Scheme: practical reasoning."]}]`), 0o644)
	out := filepath.Join(dir, "run")
	var err error
	msg := stderrOf(t, func() {
		err = annotateCmd(context.Background(), []string{"--models", prof, "--dataset", data, "--out", out})
	})
	if err != nil || !strings.Contains(msg, "ann: stopped after 3 claims in a row") {
		t.Fatalf("%v; stderr %q", err, msg)
	}
	calls, _ := os.ReadFile(filepath.Join(out, "calls.jsonl"))
	if n := strings.Count(string(calls), "\n"); n != 6 {
		t.Fatalf("%d calls, not 6 (3 claims, each sent back once)", n)
	}
}

// An annotator that tried every claim says nothing on the way out, even when
// its last claim failed: nothing stopped it (finding 2 of the review of the
// repairs).
func TestAnnotateSaysNothingOfAnAnnotatorThatFinished(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "discussion")
	os.MkdirAll(data, 0o755)
	var items []string
	for i := 2; i <= 6; i++ {
		items = append(items, `{"kialo_claim_id": "1.`+string(rune('0'+i))+`", "text": "Claim `+string(rune('0'+i))+`."}`)
	}
	os.WriteFile(filepath.Join(data, "items.json"), []byte(`{"discussion": {"thesis_claim_id": "1.1"}, "items": [`+strings.Join(items, ",")+`]}`), 0o644)
	good := `"{\"instances\": [{\"primary\": true, \"scheme\": {\"walton_id\": \"none\", \"fit\": \"none\"}}]}"`
	prof := filepath.Join(dir, "p.json")
	os.WriteFile(prof, []byte(`[{"name": "ann", "provider": "mock", "model": "m", "replies": [`+strings.Repeat(good+", ", 4)+`"Scheme: none."]}]`), 0o644)
	out := filepath.Join(dir, "run")
	var err error
	msg := stderrOf(t, func() {
		err = annotateCmd(context.Background(), []string{"--models", prof, "--dataset", data, "--out", out})
	})
	if err != nil || strings.Contains(msg, "ann:") {
		t.Fatalf("%v; stderr %q", err, msg)
	}
	ann, _ := os.ReadFile(filepath.Join(out, "annotations", "ann.json"))
	if n := strings.Count(string(ann), `"kialo_claim_id"`); n != 4 {
		t.Fatalf("%d annotations, not 4: %s", n, ann)
	}
}
