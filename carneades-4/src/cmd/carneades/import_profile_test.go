package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
)

func TestImportProfileRequiresExplicitLegacyOptIn(t *testing.T) {
	for _, format := range []string{"caf", "agxml", "lkif"} {
		if err := checkImportProfile(format, false); !errors.Is(err, caes.ErrUnsupported) || caes.ExitStatus(err) != 2 {
			t.Fatalf("legacy format %s must be refused by default: %v", format, err)
		}
		if err := checkImportProfile(format, true); err != nil {
			t.Fatalf("explicit legacy opt-in rejected: %v", err)
		}
	}
	for _, format := range []string{"yaml", "aif"} {
		if err := checkImportProfile(format, false); err != nil {
			t.Fatalf("supported format rejected: %v", err)
		}
	}
	if err := checkImportProfile("unknown", true); !errors.Is(err, caes.ErrUnsupported) {
		t.Fatalf("opt-in must not admit unknown formats: %v", err)
	}
}

func TestImportProfileCLI(t *testing.T) {
	validCAF := `<caf version="1.3"><statements><statement id="p"/></statements><arguments/></caf>`
	for _, tc := range []struct {
		name   string
		args   []string
		input  string
		status int
	}{
		{"caf check blocked", []string{"check", "-f", "caf"}, validCAF, 2},
		{"caf eval blocked", []string{"eval", "-f", "caf"}, validCAF, 2},
		{"agxml blocked", []string{"check", "-f", "agxml"}, "", 2},
		{"lkif blocked", []string{"check", "-f", "lkif"}, "", 2},
		{"caf opt in", []string{"check", "-f", "caf", "-experimental-import"}, validCAF, 0},
		{"opt in does not bypass profile", []string{"check", "-f", "caf", "-experimental-import"}, strings.Replace(validCAF, "1.3", "4", 1), 2},
		{"aif unsupported status", []string{"check", "-f", "aif"}, `{"nodes":[{"nodeID":"l","type":"L"}],"edges":[]}`, 2},
		{"native malformed premise", []string{"check"}, `statements: {p: P, c: C}
arguments: {a: {premises: [{statement: p}], conclusion: c}}`, 1},
		{"native works", []string{"check"}, "statements: {p: P}\nassumptions: [p]\n", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"-test.run=^TestImportProfileCLIProcess$", "--"}, tc.args...)
			cmd := exec.Command(os.Args[0], args...)
			cmd.Env = append(os.Environ(), "CARNEADES_IMPORT_PROFILE_CHILD=1")
			cmd.Stdin = strings.NewReader(tc.input)
			var stdout, stderr strings.Builder
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			status := 0
			if err != nil {
				var exited *exec.ExitError
				if !errors.As(err, &exited) {
					t.Fatal(err)
				}
				status = exited.ExitCode()
			}
			if status != tc.status || stdout.Len() != 0 {
				t.Fatalf("status=%d want=%d stdout=%q stderr=%q", status, tc.status, stdout.String(), stderr.String())
			}
		})
	}
}

func TestImportProfileCLIProcess(t *testing.T) {
	if os.Getenv("CARNEADES_IMPORT_PROFILE_CHILD") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"carneades"}, os.Args[i+1:]...)
			main()
			os.Exit(0)
		}
	}
	os.Exit(99)
}
