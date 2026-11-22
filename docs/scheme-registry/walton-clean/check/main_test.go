package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	caesYaml "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
)

const duplicateFixture = `language:
  p/1: "old %s"
  p/1: "last %s"
  c/1: "conclusion %s"
ignored_top_level: value
argument_schemes:
  - id: s
    variables: [X]
    premises: [p(X)]
    conclusions: [c(X)]
    ignored_scheme_field: value
`

func TestLenientProjectionIsDiagnosticOnly(t *testing.T) {
	if _, err := caesYaml.Import(strings.NewReader(duplicateFixture)); err == nil {
		t.Fatal("production importer accepted duplicate/unknown keys")
	}
	ag, err := lenientImport([]byte(duplicateFixture))
	if err != nil {
		t.Fatal(err)
	}
	if ag.Theory.Language["p/1"] != "last %s" || len(ag.Theory.ArgSchemes) != 1 {
		t.Fatal("diagnostic projection lost last-key/known-scheme behavior")
	}
	r, status := runCheckerFixture(t, duplicateFixture)
	if status != 1 || r.Pass || r.Gate["1_import"] || r.ImportError == "" || r.LenientImport == "" {
		t.Fatalf("lenient diagnostic erased strict rejection: status=%d report=%+v", status, r)
	}
}

func TestLenientProjectionKeepsParserDiagnostics(t *testing.T) {
	r, status := runCheckerFixture(t, duplicateFixture+"statements:\n  \"p(\": malformed statement\n")
	if status != 1 || r.ImportStderr == "" || r.ImportError == "" || r.LenientImport == "" {
		t.Fatalf("parser diagnostics were lost: status=%d report=%+v", status, r)
	}
}

func TestCheckerProcessHelper(t *testing.T) {
	if os.Getenv("WAG_CLEAN_CHECK_HELPER") != "1" {
		return
	}
	flag.CommandLine = flag.NewFlagSet("walton-check", flag.ExitOnError)
	os.Args = []string{"walton-check", "-json", os.Getenv("WAG_CLEAN_CHECK_REPORT"), os.Getenv("WAG_CLEAN_CHECK_INPUT")}
	main()
}

func TestOutputErrorsAreNotIgnored(t *testing.T) {
	if err := writeJSON(filepath.Join(t.TempDir(), "absent", "report.json"), report{}, "  "); err == nil {
		t.Fatal("missing output directory did not fail")
	}
	if err := writeJSON(filepath.Join(t.TempDir(), "report.json"), make(chan int), "  "); err == nil {
		t.Fatal("unsupported report value did not fail")
	}
}

func runCheckerFixture(t *testing.T, input string) (report, int) {
	t.Helper()
	dir := t.TempDir()
	in, out := filepath.Join(dir, "input.yml"), filepath.Join(dir, "report.json")
	if err := os.WriteFile(in, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestCheckerProcessHelper$")
	cmd.Env = append(os.Environ(), "WAG_CLEAN_CHECK_HELPER=1", "WAG_CLEAN_CHECK_INPUT="+in, "WAG_CLEAN_CHECK_REPORT="+out)
	var log bytes.Buffer
	cmd.Stdout, cmd.Stderr = &log, &log
	err := cmd.Run()
	status := 0
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			status = exit.ExitCode()
		} else {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("no diagnostic report: %v\n%s", err, log.String())
	}
	var r report
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	return r, status
}
