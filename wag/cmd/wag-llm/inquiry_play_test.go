package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/evidence/evidencetest"
)

// stdoutOf runs f and returns what it wrote to standard output.
func stdoutOf(t *testing.T, f func()) string {
	t.Helper()
	tmp, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = tmp
	func() {
		defer func() { os.Stdout = old }()
		f()
	}()
	tmp.Close()
	b, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// A check beyond E1's cases (docs/e1-evidence.md, Results of E1.5 and E1.6),
// its outcome written before it ran: wag-llm play runs an inquiry of two
// model seats with a record set (--mode inquiry --evidence DIR --record-set
// NAME). Its seats are A and B; the summary records the profiles as first
// and second, with no pro or con; the line it prints names the claim's
// label, not a winner; the trace records the record set and the protocol
// wag-protocol/2+evidence, and replays with --evidence; report counts the
// games of each profile in each seat with no wins ("-"). --record-set
// without --evidence is refused, naming --evidence, before the run starts:
// its folder is not made. (The server refuses such a game too, but only
// once the run's folder and summary are made; run 2 of the planted defects
// found the test could not tell them apart.)
func TestPlayAndReportOfAnInquiry(t *testing.T) {
	dir := t.TempDir()
	evidenceDir := filepath.Join(dir, "evidence")
	if err := evidencetest.Write(evidenceDir, "fixture", evidencetest.Set()); err != nil {
		t.Fatal(err)
	}
	prof := filepath.Join(dir, "profiles.json")
	if err := os.WriteFile(prof, []byte(`{"profiles":[{"name":"a","provider":"mock","model":"mock-1","replies":[
		"{\"summary\":\"State the claim.\",\"action\":{\"type\":\"support\",\"scheme_id\":\"AS:consequences\",\"variable_bindings\":{\"A\":\"tax meat\",\"G\":\"lower emissions\"},\"declare\":[{\"id\":\"tax meat\",\"actor\":\"decider\"}]}}",
		"{\"summary\":\"Nothing more.\",\"action\":{\"type\":\"abstain\"}}"]},
		{"name":"b","provider":"mock","model":"mock-1","replies":["{\"summary\":\"Nothing to add.\",\"action\":{\"type\":\"abstain\"}}"]}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	out := filepath.Join(dir, "out")
	none := filepath.Join(dir, "none")
	if err := play(ctx, []string{"--models", prof, "--mode", "inquiry", "--pro", "a", "--con", "b", "--out", none,
		"--record-set", "fixture"}); err == nil || !strings.Contains(err.Error(), "--evidence") {
		t.Fatalf("--record-set without --evidence: %v", err)
	}
	if _, err := os.Stat(none); !os.IsNotExist(err) {
		t.Fatalf("--record-set without --evidence made the run's folder: %v", err)
	}
	var err error
	printed := stdoutOf(t, func() {
		err = play(ctx, []string{"--models", prof, "--mode", "inquiry", "--pro", "a", "--con", "b", "--out", out, "--max-rounds", "4",
			"--evidence", evidenceDir, "--record-set", "fixture"})
	})
	if err != nil {
		t.Fatalf("play: %v", err)
	}
	if !regexp.MustCompile(`game 1: a and b \(an inquiry\): the claim (In|Out|Undecided), `).MatchString(printed) || strings.Contains(printed, "(pro)") {
		t.Fatalf("the line printed: %q", printed)
	}
	sum, _ := os.ReadFile(filepath.Join(out, "summary.jsonl"))
	for _, want := range []string{`"mode":"inquiry"`, `"first":"a"`, `"second":"b"`} {
		if !strings.Contains(string(sum), want) {
			t.Fatalf("the summary has no %s: %s", want, sum)
		}
	}
	if strings.Contains(string(sum), `"pro":`) || strings.Contains(string(sum), `"con":`) {
		t.Fatalf("the summary names sides: %s", sum)
	}
	traces, _ := filepath.Glob(filepath.Join(out, "*.trace.jsonl"))
	if len(traces) != 1 {
		t.Fatalf("traces: %v", traces)
	}
	b, _ := os.ReadFile(traces[0])
	if !strings.Contains(string(b), `"protocol":"wag-protocol/2+evidence"`) || !strings.Contains(string(b), `"record_set":{"name":"fixture"`) {
		t.Fatal("the trace records no record set or not its protocol")
	}
	if err := replay(ctx, []string{traces[0], "--evidence", evidenceDir}); err != nil {
		t.Fatalf("replay: %v", err)
	}
	printed = stdoutOf(t, func() { err = report([]string{out}) })
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	for _, row := range []string{`a in an inquiry, as A +1 +- `, `b in an inquiry, as B +1 +- `} {
		if !regexp.MustCompile(row).MatchString(printed) {
			t.Fatalf("the report has no row %q:\n%s", row, printed)
		}
	}
	if strings.Contains(printed, " as pro") || strings.Contains(printed, " as con") {
		t.Fatalf("the report names sides:\n%s", printed)
	}
}
