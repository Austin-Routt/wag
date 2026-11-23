package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/yourorg/wag/internal/llm"
)

// Limits apply to one CLI invocation, including all its clients and retries.
// Only the client-attempt count is a hard cap; token/cost checks are estimates.
type budgetFlags struct {
	calls, tokens *int
	cost          *float64
	prices        *string
}

func addBudgetFlags(fs *flag.FlagSet) budgetFlags {
	return budgetFlags{
		calls:  fs.Int("max-calls", 0, "stop this invocation before its N+1th client attempt, counting retries (D046)"),
		tokens: fs.Int("max-tokens", 0, "refuse a call whose estimated reservation exceeds N accounted tokens; not a provider billing ceiling (D046)"),
		cost:   fs.Float64("max-cost", 0, "refuse a call whose estimated price-table cost exceeds this invocation allowance; not a billing ceiling (D046)"),
		prices: fs.String("prices", "", `JSON rates per million tokens, with explicit input/output rates, by profile or model ("*" fallback): {"m":{"input":0.3,"output":2.5}}`),
	}
}

// decodePrices requires complete, unambiguous rates. Null, omitted, duplicated
// or misspelled rate fields must not become a silent zero or overwrite a rate.
func decodePrices(raw []byte) (map[string]llm.Price, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return nil, fmt.Errorf("prices must be an object keyed by profile/model")
	}
	prices := map[string]llm.Price{}
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return nil, err
		}
		name, ok := token.(string)
		if !ok || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("price name must be nonempty text")
		}
		if _, exists := prices[name]; exists {
			return nil, fmt.Errorf("duplicate price name %q", name)
		}
		start, err := d.Token()
		if err != nil || start != json.Delim('{') {
			return nil, fmt.Errorf("price %q needs explicit input and output rates", name)
		}
		rates := map[string]float64{}
		for d.More() {
			token, err := d.Token()
			if err != nil {
				return nil, err
			}
			field, ok := token.(string)
			if !ok || (field != "input" && field != "output") {
				return nil, fmt.Errorf("price %q has an unknown rate field %q", name, field)
			}
			if _, exists := rates[field]; exists {
				return nil, fmt.Errorf("price %q repeats rate %q", name, field)
			}
			token, err = d.Token()
			if err != nil {
				return nil, err
			}
			number, ok := token.(json.Number)
			if !ok {
				return nil, fmt.Errorf("price %q %s must be a finite nonnegative number", name, field)
			}
			value, err := number.Float64()
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
				return nil, fmt.Errorf("price %q %s must be a finite nonnegative number", name, field)
			}
			rates[field] = value
		}
		if _, err := d.Token(); err != nil {
			return nil, err
		}
		if len(rates) != 2 {
			return nil, fmt.Errorf("price %q needs explicit input and output rates", name)
		}
		prices[name] = llm.Price{Input: rates["input"], Output: rates["output"]}
	}
	if _, err := d.Token(); err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing data after prices")
	}
	return prices, nil
}

func needsCallBudget(p llm.Profile) bool {
	if p.Provider == "mock" {
		return false
	}
	return llm.NeedsKey(p.Provider) || strings.TrimSpace(p.APIKeyEnv) != "" || strings.TrimSpace(p.APIKey) != ""
}

// Each invocation starts a new allowance, even when resuming the same saved
// annotation run. Keyed profiles need a real attempt cap; estimated token/cost
// allowances can supplement it but cannot replace it.
func (f budgetFlags) budget(used []llm.Profile) (*llm.Budget, error) {
	b := &llm.Budget{MaxCalls: *f.calls, MaxTokens: *f.tokens, MaxCost: *f.cost}
	if err := b.Check(); err != nil {
		return nil, err
	}
	if *f.prices != "" {
		raw, err := os.ReadFile(*f.prices)
		if err != nil {
			return nil, err
		}
		if b.Prices, err = decodePrices(raw); err != nil {
			return nil, fmt.Errorf("--prices %s: %v", *f.prices, err)
		}
	}
	if err := b.Check(); err != nil {
		return nil, err
	}
	if b.MaxCost > 0 && len(b.Prices) == 0 {
		return nil, fmt.Errorf("--max-cost needs --prices")
	}
	for _, p := range used {
		if needsCallBudget(p) && b.MaxCalls <= 0 {
			return nil, fmt.Errorf("profile %s uses a key-requiring or credential-configured provider, so this invocation needs --max-calls (D046); token/cost estimates alone are insufficient", p.Name)
		}
	}
	if !b.Limited() {
		return nil, nil
	}
	return b, nil
}

func reportBudget(w io.Writer, b *llm.Budget) {
	if b == nil {
		return
	}
	s := b.Spent()
	fmt.Fprintf(w, "Invocation budget: %d client attempts", s.Calls)
	if b.MaxCalls > 0 {
		fmt.Fprintf(w, " of %d", b.MaxCalls)
	}
	if b.MaxTokens > 0 || b.MaxCost > 0 {
		fmt.Fprintf(w, ", %d accounted tokens", s.Tokens)
		if b.MaxTokens > 0 {
			fmt.Fprintf(w, " of %d", b.MaxTokens)
		}
	}
	if b.MaxCost > 0 {
		fmt.Fprintf(w, ", estimated price-table cost %.6f of %.6f", s.Cost, b.MaxCost)
	}
	fmt.Fprintln(w, " (reported usage with conservative fallbacks; token/cost reservations are not billing ceilings; a resumed invocation starts a fresh allowance).")
}
