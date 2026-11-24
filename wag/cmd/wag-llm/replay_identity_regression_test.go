package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A changed data fingerprint cannot be reported as an exact trace replay,
// even if every recorded move happens to remain legal.
func TestReviewReplayRefusesChangedDataIdentity(t *testing.T) {
	dir := t.TempDir()
	if err := play(context.Background(), []string{"--pro", "greedy", "--con", "random", "--out", dir, "--max-rounds", "2"}); err != nil {
		t.Fatal(err)
	}
	paths, err := filepath.Glob(filepath.Join(dir, "*.trace.jsonl"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("trace: %v %v", paths, err)
	}
	b, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")
	var start map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &start); err != nil {
		t.Fatal(err)
	}
	start["data_sha256"] = "changed-data-identity"
	x, err := json.Marshal(start)
	if err != nil {
		t.Fatal(err)
	}
	lines[0] = string(x)
	bad := filepath.Join(t.TempDir(), "changed.jsonl")
	if err := os.WriteFile(bad, []byte(strings.Join(lines, "\n")), 0600); err != nil {
		t.Fatal(err)
	}
	if err := replay(context.Background(), []string{bad}); err == nil || !strings.Contains(err.Error(), "scheme data") {
		t.Fatalf("known data mismatch did not fail replay: %v", err)
	}
}
