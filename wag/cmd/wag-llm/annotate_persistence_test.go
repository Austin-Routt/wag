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
