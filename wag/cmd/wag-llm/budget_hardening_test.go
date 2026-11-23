package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/wag/internal/llm"
)

func budgetFromArgs(t *testing.T, args []string, profiles []llm.Profile) (*llm.Budget, error) {
	t.Helper()
	fs := flag.NewFlagSet("budget-test", flag.ContinueOnError)
	f := addBudgetFlags(fs)
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	b, err := f.budget()
	if err == nil {
		f.note(io.Discard, b, profiles)
	}
	return b, err
}

func TestBudgetHardeningCLILimitsAndPrices(t *testing.T) {
	for _, cost := range []string{"NaN", "+Inf", "-Inf", "-1"} {
		if _, err := budgetFromArgs(t, []string{"--max-calls", "1", "--max-cost", cost}, nil); err == nil {
			t.Errorf("invalid cost limit accepted: %s", cost)
		}
	}
	for _, raw := range []string{
		`{"*":{"input":1}}`, `{"*":{"output":1}}`, `{"*":{"input":null,"output":1}}`,
		`{"*":{"input":-1,"output":1}}`, `{"*":{"input":1,"output":-1}}`, `{"*":null}`,
		`{"*":{"input":1,"output":1,"typo":2}}`, `{"*":{"input":-1,"input":1,"output":1}}`,
		`{"*":{"input":1,"output":1},"*":{"input":0,"output":0}}`,
	} {
		file := filepath.Join(t.TempDir(), "prices.json")
		if err := os.WriteFile(file, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := budgetFromArgs(t, []string{"--max-calls", "1", "--max-cost", "1", "--prices", file}, nil); err == nil {
			t.Errorf("invalid price table accepted: %s", raw)
		}
	}
	file := filepath.Join(t.TempDir(), "prices.json")
	if err := os.WriteFile(file, []byte(`{"*":{"input":0,"output":0}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := budgetFromArgs(t, []string{"--max-cost", "1", "--prices", file}, nil); err != nil {
		t.Fatalf("explicit free price rates should remain valid: %v", err)
	}
}

// The patch as received required --max-calls of keyed profiles; the owner
// decided (D046, 2026-11-23) that no limit is the default, so a keyed profile
// starts with no flag, with --unlimited, or with any one limit.
func TestBudgetKeyedProfilesStartWithoutALimit(t *testing.T) {
	for _, p := range []llm.Profile{
		{Name: "remote", Provider: "openai-compatible", BaseURL: "https://example.invalid/v1", APIKeyEnv: "UNUSED_BUDGET_TEST_KEY"},
		{Name: "inline", Provider: "openai-compatible", BaseURL: "https://example.invalid/v1", APIKey: "not-a-real-key"},
		{Name: "keyed", Provider: "anthropic"},
		{Name: "overridden", Provider: "ollama", BaseURL: "https://example.invalid/v1", APIKeyEnv: "UNUSED_BUDGET_TEST_KEY"},
	} {
		if !keyed(p) {
			t.Errorf("profile %s is not seen to use a key", p.Name)
		}
		for _, args := range [][]string{nil, {"--unlimited"}, {"--max-calls", "1"}, {"--max-tokens", "1000"}} {
			b, err := budgetFromArgs(t, args, []llm.Profile{p})
			if err != nil || b == nil {
				t.Errorf("%v refused for %s: %v", args, p.Name, err)
			}
		}
	}
	for _, p := range []llm.Profile{{Name: "local", Provider: "ollama"}, {Name: "mock", Provider: "mock"}} {
		if keyed(p) {
			t.Errorf("profile %s is seen to use a key", p.Name)
		}
	}
}
