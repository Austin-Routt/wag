package main

import (
	"flag"
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
	return f.budget(profiles)
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

func TestBudgetHardeningKeyedProfilesNeedCallCap(t *testing.T) {
	for _, p := range []llm.Profile{
		{Name: "remote", Provider: "openai-compatible", BaseURL: "https://example.invalid/v1", APIKeyEnv: "UNUSED_BUDGET_TEST_KEY"},
		{Name: "inline", Provider: "openai-compatible", BaseURL: "https://example.invalid/v1", APIKey: "not-a-real-key"},
		{Name: "keyed", Provider: "anthropic"},
		{Name: "overridden", Provider: "ollama", BaseURL: "https://example.invalid/v1", APIKeyEnv: "UNUSED_BUDGET_TEST_KEY"},
	} {
		for _, args := range [][]string{nil, {"--max-tokens", "1000"}} {
			if _, err := budgetFromArgs(t, args, []llm.Profile{p}); err == nil {
				t.Errorf("keyed profile %s started without a call cap: %v", p.Name, args)
			}
		}
		if _, err := budgetFromArgs(t, []string{"--max-calls", "1"}, []llm.Profile{p}); err != nil {
			t.Errorf("explicit call cap refused for %s: %v", p.Name, err)
		}
	}
	if _, err := budgetFromArgs(t, nil, []llm.Profile{{Name: "local", Provider: "ollama"}, {Name: "mock", Provider: "mock"}}); err != nil {
		t.Fatalf("unkeyed local/mock compatibility changed: %v", err)
	}
}
