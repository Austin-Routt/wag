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

const validFixture = `language:
  p/1: "p %s"
  c/1: "c %s"
argument_schemes:
  - id: s
    variables: [X]
    premises: [p(X)]
    conclusions: [c(X)]
`

// Each part of check 1 fails the import gate on its own; in particular an
// import error without strict-decode errors (the strict engine's own
// refusal) is not passed by the lenient reading.
func TestImportGateNeedsEveryCheck(t *testing.T) {
	if !importPasses(&report{}) {
		t.Fatal("a clean report failed the import gate")
	}
	for name, r := range map[string]report{
		"yaml error":    {YAMLError: "x"},
		"import error":  {ImportError: "x", LenientImport: "read leniently"},
		"import stderr": {ImportStderr: "x"},
		"strict decode": {StrictErrors: []string{"x"}},
	} {
		r := r
		if importPasses(&r) {
			t.Errorf("%s passed the import gate", name)
		}
	}
}

// A report that cannot be written stops the checker with status 2, an
// operational error, which run.sh does not take for a finding (status 1).
func TestReportWriteErrorStopsTheChecker(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "input.yml")
	if err := os.WriteFile(in, []byte(validFixture), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestCheckerProcessHelper$")
	cmd.Env = append(os.Environ(), "WAG_CLEAN_CHECK_HELPER=1", "WAG_CLEAN_CHECK_INPUT="+in,
		"WAG_CLEAN_CHECK_REPORT="+filepath.Join(dir, "absent", "report.json"))
	var log bytes.Buffer
	cmd.Stdout, cmd.Stderr = &log, &log
	err := cmd.Run()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 2 {
		t.Fatalf("unwritable report: got %v, want exit status 2\n%s", err, log.String())
	}
	if !strings.Contains(log.String(), "write report") {
		t.Fatalf("the failure does not say what failed:\n%s", log.String())
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

// The language check reads indexed templates as WAG's renderer does (D060),
// and plain ones as before.
func TestTemplateProblem(t *testing.T) {
	for _, c := range []struct {
		tmpl  string
		arity int
		ok    bool
	}{
		{"%s is good.", 1, true},
		{"%s is good, %s too.", 1, false},
		{"100%% of %s.", 1, true},
		{"%s and %d.", 2, false},
		{"I will see to it that %[2]s happens if you go ahead with %[1]s.", 2, true},
		{"At %[4]s, %[1]s is committed to %[2]s by what was done at %[3]s.", 4, true},
		{"%[2]s, then %s, after %[1]s.", 3, true},
		{"%[2]s, then %s.", 3, false},
		{"%[2]s, then %s.", 2, false},
		{"%[1]s twice: %[1]s.", 2, false},
		{"%[1]s and %[1]s and %[2]s.", 2, true},
		{"%[3]s and %[1]s.", 2, false},
		{"%[0]s and %[1]s.", 1, false},
		{"%[]s and %s.", 2, false},
		{"%[1]d and %[2]s.", 2, false},
		{"%[1]s and %[2]s %", 2, false},
		{"100%% of %[1]s.", 1, true},
	} {
		if got := templateProblem(c.tmpl, c.arity); (got == "") != c.ok {
			t.Errorf("%q, arity %d: %q, want ok=%v", c.tmpl, c.arity, got, c.ok)
		}
	}
}
