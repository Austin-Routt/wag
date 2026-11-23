package main

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelperProcess runs main in a child process, so that its exit status
// can be tested (the pattern of os/exec's own tests).
func TestHelperProcess(t *testing.T) {
	if os.Getenv("WAG_EVIDENCE_HELPER") != "1" {
		return
	}
	args := []string{}
	if a := os.Getenv("WAG_EVIDENCE_ARGS"); a != "" {
		args = strings.Split(a, "\x00")
	}
	os.Args = append([]string{"wag-evidence"}, args...)
	flag.CommandLine = flag.NewFlagSet("wag-evidence", flag.ExitOnError)
	main()
	os.Exit(0)
}

func run(t *testing.T, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
	cmd.Env = append(os.Environ(), "WAG_EVIDENCE_HELPER=1", "WAG_EVIDENCE_ARGS="+strings.Join(args, "\x00"))
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	if err == nil {
		return 0, out.String()
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), out.String()
	}
	t.Fatal(err)
	return -1, ""
}

func copyExample(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range []string{"bundle.json", "report.txt"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", "evidence", f))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, f), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// The exit status says what happened: 0 a valid bundle, 1 an invalid one
// (malformed, tampered, too deep), 2 a usage or file error.
func TestExitStatus(t *testing.T) {
	dir := copyExample(t)
	if code, out := run(t, filepath.Join(dir, "bundle.json")); code != 0 || !strings.Contains(out, `"valid": true`) {
		t.Fatalf("the example: %d\n%s", code, out)
	}
	tampered := copyExample(t)
	if err := os.WriteFile(filepath.Join(tampered, "report.txt"), []byte("The instrument reported 30 °C.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if code, out := run(t, filepath.Join(tampered, "bundle.json")); code != 1 || !strings.Contains(out, `"valid": false`) {
		t.Fatalf("tampered bytes: %d\n%s", code, out)
	}
	deep := filepath.Join(t.TempDir(), "deep.json")
	if err := os.WriteFile(deep, bytes.Repeat([]byte("["), 3_000_000), 0600); err != nil {
		t.Fatal(err)
	}
	if code, out := run(t, deep); code != 1 || !strings.Contains(out, "deeper than") {
		t.Fatalf("deep JSON: %d\n%s", code, out)
	}
	if code, out := run(t); code != 2 || !strings.Contains(out, "usage") {
		t.Fatalf("no argument: %d\n%s", code, out)
	}
	if code, _ := run(t, filepath.Join(dir, "absent.json")); code != 2 {
		t.Fatalf("a missing file: %d", code)
	}
}
