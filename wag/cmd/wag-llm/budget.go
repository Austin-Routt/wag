package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yourorg/wag/internal/llm"
)

// budgetFlags are a run's hard limits (D046): every command that calls a
// model with the owner's key takes them.
type budgetFlags struct {
	calls, tokens *int
	cost          *float64
	prices        *string
}

func addBudgetFlags(fs *flag.FlagSet) budgetFlags {
	return budgetFlags{
		calls:  fs.Int("max-calls", 0, "stop the run before its N+1th call to a model, counting every retry (D046)"),
		tokens: fs.Int("max-tokens", 0, "refuse any call whose worst case would take the run past N tokens (needs max_output_tokens on the profiles; D046)"),
		cost:   fs.Float64("max-cost", 0, "refuse any call whose worst case would take the run past this cost, priced from --prices (D046)"),
		prices: fs.String("prices", "", `a JSON file of prices per million tokens, by profile name or model ("*" for any other): {"m": {"input": 0.3, "output": 2.5}}`),
	}
}

// budget builds the run's budget for the profiles it calls. A run that
// calls a provider with an API key must have a limit (D046); local
// providers and mock ones need none.
func (f budgetFlags) budget(used []llm.Profile) (*llm.Budget, error) {
	b := &llm.Budget{MaxCalls: *f.calls, MaxTokens: *f.tokens, MaxCost: *f.cost}
	if b.MaxCalls < 0 || b.MaxTokens < 0 || b.MaxCost < 0 {
		return nil, fmt.Errorf("a limit cannot be negative")
	}
	if *f.prices != "" {
		raw, err := os.ReadFile(*f.prices)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &b.Prices); err != nil {
			return nil, fmt.Errorf("--prices %s: %v", *f.prices, err)
		}
	}
	if b.MaxCost > 0 && len(b.Prices) == 0 {
		return nil, fmt.Errorf("--max-cost needs --prices")
	}
	if !b.Limited() {
		for _, p := range used {
			if llm.NeedsKey(p.Provider) {
				return nil, fmt.Errorf("profile %s calls %s with an API key, so the run needs a limit (D046): give --max-calls, --max-tokens or --max-cost", p.Name, p.Provider)
			}
		}
		return nil, nil
	}
	return b, nil
}

// reportBudget says what the run spent against its limits.
func reportBudget(w io.Writer, b *llm.Budget) {
	if b == nil {
		return
	}
	s := b.Spent()
	fmt.Fprintf(w, "Budget: %d calls", s.Calls)
	if b.MaxCalls > 0 {
		fmt.Fprintf(w, " of %d", b.MaxCalls)
	}
	if b.MaxTokens > 0 || b.MaxCost > 0 {
		fmt.Fprintf(w, ", %d tokens", s.Tokens)
		if b.MaxTokens > 0 {
			fmt.Fprintf(w, " of %d", b.MaxTokens)
		}
	}
	if b.MaxCost > 0 {
		fmt.Fprintf(w, ", cost %.4f of %.4f", s.Cost, b.MaxCost)
	}
	fmt.Fprintln(w, " (worst cases where a provider reported no usage).")
}
