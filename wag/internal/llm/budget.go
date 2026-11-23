package llm

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

// Budget bounds client attempts for one invocation, retries included. Token
// and price-table cost reservations are estimates, not provider billing ceilings:
// hidden framing and provider behavior are not measured here. A zero limit is
// disabled. Configuration must remain immutable while the budget is in use.
// Pending reservations are shared safely by concurrent clients. No state is
// restored from an earlier invocation or a resumed annotation directory.
type Budget struct {
	MaxCalls  int
	MaxTokens int
	MaxCost   float64
	// Prices are per million tokens, by profile name, then model, then "*".
	Prices map[string]Price

	mu      sync.Mutex
	spent   Spent
	pending Spent
	refused int
	halted  string // malformed usage or an estimate overrun stops later attempts
}

// Price is the caller's declared rate per million tokens. It is not verified
// against a provider's invoice. Explicit zero rates are allowed.
type Price struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

// Spent records calls and accounted usage/cost; absent or invalid usage uses
// conservative reservation fallbacks. The attempt trace retains raw usage.
type Spent struct {
	Calls  int     `json:"calls"`
	Tokens int     `json:"tokens"`
	Cost   float64 `json:"cost"`
}

type BudgetError struct {
	Profile string
	Reason  string
}

func (e *BudgetError) Error() string {
	return fmt.Sprintf("budget (D046): %s: %s", e.Profile, e.Reason)
}

func budgetError(p Profile, why string) error {
	be := &BudgetError{Profile: p.Name, Reason: why}
	return &Error{Provider: p.Provider, Kind: KindBudget, Message: be.Error(), Err: be}
}

func finiteNonnegative(n float64) bool { return !math.IsNaN(n) && !math.IsInf(n, 0) && n >= 0 }

// Check also protects native callers, which need not use the CLI price parser.
// Native Price values have explicit Go fields; JSON field presence is checked
// separately by the CLI.
func (b *Budget) Check() error {
	if b == nil {
		return nil
	}
	if b.MaxCalls < 0 || b.MaxTokens < 0 || !finiteNonnegative(b.MaxCost) {
		return fmt.Errorf("budget limits must be finite and nonnegative")
	}
	for key, p := range b.Prices {
		if strings.TrimSpace(key) == "" || !finiteNonnegative(p.Input) || !finiteNonnegative(p.Output) {
			return fmt.Errorf("price %q needs finite nonnegative input and output rates", key)
		}
	}
	return nil
}

func (b *Budget) Limited() bool {
	return b != nil && (b.MaxCalls > 0 || b.MaxTokens > 0 || b.MaxCost > 0)
}

func (b *Budget) Refused() bool {
	if b == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.refused > 0
}

func (b *Budget) Spent() Spent {
	if b == nil {
		return Spent{}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.spent
}

func (b *Budget) price(p Profile) (Price, bool) {
	for _, k := range []string{p.Name, p.Params.Model, "*"} {
		if pr, ok := b.Prices[k]; ok && k != "" {
			return pr, true
		}
	}
	return Price{}, false
}

const maxBudgetInt = int(^uint(0) >> 1)

func addCount(a, b int) (int, bool) {
	if a < 0 || b < 0 || a > maxBudgetInt-b {
		return maxBudgetInt, false
	}
	return a + b, true
}

// estimate counts supplied text and schema bytes as a token proxy, plus the
// requested output allowance. Provider framing/hidden additions remain unknown.
// Extra can replace those fields, so it is unsupported for estimated budgets.
func estimate(p Profile, req Request) (in, out int, err error) {
	if len(req.Extra) != 0 {
		return 0, 0, fmt.Errorf("Extra is unsupported with token/cost estimates because it can replace the reserved request; use a call-only budget for custom provider fields")
	}
	in = len(req.System)
	for _, m := range req.Messages {
		var ok bool
		if in, ok = addCount(in, len(m.Content)); !ok {
			return 0, 0, fmt.Errorf("input estimate overflows token accounting")
		}
	}
	if req.Output == "schema" {
		var ok bool
		if in, ok = addCount(in, len(req.Schema)); !ok {
			return 0, 0, fmt.Errorf("schema estimate overflows token accounting")
		}
	}
	out = req.MaxOutputTokens
	if out <= 0 {
		return 0, 0, fmt.Errorf("token/cost estimates need max_output_tokens")
	}
	if p.Provider == "gemini" {
		if req.ThinkingBudget == nil || *req.ThinkingBudget < 0 || *req.ThinkingBudget > math.MaxInt32 {
			return 0, 0, fmt.Errorf("Gemini estimates need thinking_budget in [0, %d]", math.MaxInt32)
		}
		var ok bool
		if out, ok = addCount(out, *req.ThinkingBudget); !ok {
			return 0, 0, fmt.Errorf("reasoning estimate overflows token accounting")
		}
	}
	if _, ok := addCount(in, out); !ok {
		return 0, 0, fmt.Errorf("request estimate overflows token accounting")
	}
	return in, out, nil
}

func priced(in, out int, p Price) float64 {
	return float64(in)/1e6*p.Input + float64(out)/1e6*p.Output
}

// Reserve refuses an attempt before dispatch if its invocation call count or
// estimated reservation would exceed a configured limit.
func (b *Budget) Reserve(p Profile, req Request) (Spent, error) {
	if b == nil {
		return Spent{}, nil
	}
	refuse := func(why string) (Spent, error) {
		b.mu.Lock()
		b.refused++
		b.mu.Unlock()
		return Spent{}, budgetError(p, why)
	}
	if err := b.Check(); err != nil {
		return refuse(err.Error())
	}
	if !b.Limited() {
		return Spent{}, nil
	}
	w := Spent{Calls: 1}
	if b.MaxTokens > 0 || b.MaxCost > 0 {
		in, out, err := estimate(p, req)
		if err != nil {
			return refuse(err.Error())
		}
		w.Tokens = in + out // estimate checked this sum
		if b.MaxCost > 0 {
			pr, ok := b.price(p)
			if !ok {
				return refuse("no price for this profile/model, so estimated cost is unknown")
			}
			w.Cost = priced(in, out, pr)
			if !finiteNonnegative(w.Cost) {
				return refuse("estimated cost overflows accounting")
			}
		}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	calls, callsOK := addCount(b.spent.Calls, b.pending.Calls)
	tokens, tokensOK := addCount(b.spent.Tokens, b.pending.Tokens)
	totalCalls, callSumOK := addCount(calls, w.Calls)
	totalTokens, tokenSumOK := addCount(tokens, w.Tokens)
	cost := b.spent.Cost + b.pending.Cost + w.Cost
	why := ""
	switch {
	case b.halted != "":
		why = "invocation stopped after an accounting violation: " + b.halted
	case !callsOK || !tokensOK || !callSumOK || !tokenSumOK || !finiteNonnegative(cost):
		why = "invocation accounting would overflow"
	case b.MaxCalls > 0 && totalCalls > b.MaxCalls:
		why = fmt.Sprintf("the invocation's %d calls are used", b.MaxCalls)
	case b.MaxTokens > 0 && totalTokens > b.MaxTokens:
		why = fmt.Sprintf("the estimated %d tokens would exceed the invocation's %d-token allowance", w.Tokens, b.MaxTokens)
	case b.MaxCost > 0 && cost > b.MaxCost:
		why = fmt.Sprintf("the estimated cost %.6f would exceed the invocation's %.6f allowance", w.Cost, b.MaxCost)
	}
	if why != "" {
		b.refused++
		return Spent{}, budgetError(p, why)
	}
	b.pending.Calls += w.Calls
	b.pending.Tokens += w.Tokens
	b.pending.Cost += w.Cost
	return w, nil
}

// reportedUsage uses reported components where coherent and reservation
// components where absent. Gemini output excludes separately reported reasoning.
// A reported total with a reported input can supply the full output count.
func reportedUsage(p Profile, req Request, u Usage, in, out int) (int, int, error) {
	for _, count := range []*int{u.InputTokens, u.OutputTokens, u.ReasoningTokens, u.CachedTokens, u.TotalTokens} {
		if count != nil && *count < 0 {
			return 0, 0, fmt.Errorf("negative usage count")
		}
	}
	if u.InputTokens != nil {
		in = *u.InputTokens
	}
	if p.Provider == "gemini" {
		answer, reasoning := req.MaxOutputTokens, *req.ThinkingBudget
		if u.OutputTokens != nil {
			answer = *u.OutputTokens
		}
		if u.ReasoningTokens != nil {
			reasoning = *u.ReasoningTokens
		}
		var ok bool
		if out, ok = addCount(answer, reasoning); !ok {
			return 0, 0, fmt.Errorf("output/reasoning usage overflows accounting")
		}
	} else {
		if u.OutputTokens != nil {
			out = *u.OutputTokens
		}
		if u.ReasoningTokens != nil {
			if u.OutputTokens != nil && *u.ReasoningTokens > out {
				return 0, 0, fmt.Errorf("reasoning usage exceeds inclusive output usage")
			}
			out = max(out, *u.ReasoningTokens)
		}
	}
	if u.TotalTokens != nil {
		minimumOut := 0
		if u.OutputTokens != nil {
			minimumOut = *u.OutputTokens
		}
		if p.Provider == "gemini" && u.ReasoningTokens != nil {
			var ok bool
			if minimumOut, ok = addCount(minimumOut, *u.ReasoningTokens); !ok {
				return 0, 0, fmt.Errorf("reported output overflows accounting")
			}
		} else if u.ReasoningTokens != nil {
			// Inclusive reasoning is a known part of output even when the
			// provider omitted the aggregate output field.
			minimumOut = max(minimumOut, *u.ReasoningTokens)
		}
		minimum := minimumOut
		if u.InputTokens != nil {
			var ok bool
			if minimum, ok = addCount(*u.InputTokens, minimumOut); !ok {
				return 0, 0, fmt.Errorf("reported usage overflows accounting")
			}
		}
		if *u.TotalTokens < minimum {
			return 0, 0, fmt.Errorf("total usage is below reported components")
		}
		if u.InputTokens != nil {
			inferredOut := *u.TotalTokens - *u.InputTokens
			completeOut := u.OutputTokens != nil && (p.Provider != "gemini" || u.ReasoningTokens != nil)
			if completeOut && inferredOut != out {
				return 0, 0, fmt.Errorf("total usage disagrees with reported components")
			}
			out = inferredOut
		}
	}
	if _, ok := addCount(in, out); !ok {
		return 0, 0, fmt.Errorf("reported usage overflows accounting")
	}
	return in, out, nil
}

// Charge must be called exactly once for a successful reservation. Failed calls
// retain their estimate; malformed usage retains at least that estimate and any
// larger nonnegative observed components. An accounting violation stops future
// attempts, but cannot undo already-dispatched requests or guarantee billing.
func (b *Budget) Charge(p Profile, req Request, w Spent, resp *Response) error {
	if !b.Limited() {
		return nil
	}
	used := w
	why := ""
	if resp != nil && w.Tokens > 0 {
		in, out, err := estimate(p, req)
		if err != nil {
			why = err.Error()
		} else {
			u := resp.Usage
			gotIn, gotOut, usageErr := reportedUsage(p, req, u, in, out)
			if usageErr != nil {
				why = usageErr.Error()
				gotIn, gotOut = in, out
				if u.InputTokens != nil && *u.InputTokens > gotIn {
					gotIn = *u.InputTokens
				}
				observedOut := 0
				if u.OutputTokens != nil && *u.OutputTokens > 0 {
					observedOut = *u.OutputTokens
				}
				if u.ReasoningTokens != nil && *u.ReasoningTokens > 0 {
					if p.Provider == "gemini" {
						observedOut, _ = addCount(observedOut, *u.ReasoningTokens)
					} else {
						observedOut = max(observedOut, *u.ReasoningTokens)
					}
				}
				gotOut = max(gotOut, observedOut)
			}
			used.Tokens, _ = addCount(gotIn, gotOut) // saturates on invalid/overflowing observations
			if u.TotalTokens != nil && *u.TotalTokens > used.Tokens {
				used.Tokens = *u.TotalTokens
			}
			if b.MaxCost > 0 {
				pr, _ := b.price(p)
				used.Cost = priced(gotIn, gotOut, pr)
				if u.TotalTokens != nil && *u.TotalTokens > 0 && (usageErr != nil || u.InputTokens == nil) {
					used.Cost = math.Max(used.Cost, float64(*u.TotalTokens)/1e6*math.Max(pr.Input, pr.Output))
				}
				if !finiteNonnegative(used.Cost) {
					used.Cost = math.MaxFloat64
					why = "reported cost overflows accounting"
				}
			}
			if used.Tokens > w.Tokens || used.Cost > w.Cost {
				if why == "" {
					why = "accounted usage exceeded its reserved estimate; observed usage is retained"
				}
			}
		}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pending.Calls -= w.Calls
	b.pending.Tokens -= w.Tokens
	b.pending.Cost = math.Max(0, b.pending.Cost-w.Cost)
	var ok bool
	if b.spent.Calls, ok = addCount(b.spent.Calls, used.Calls); !ok {
		why = "call accounting overflow"
	}
	if b.spent.Tokens, ok = addCount(b.spent.Tokens, used.Tokens); !ok {
		why = "token accounting overflow"
	}
	b.spent.Cost += used.Cost
	if !finiteNonnegative(b.spent.Cost) {
		b.spent.Cost = math.MaxFloat64
		why = "cost accounting overflow"
	}
	if why != "" {
		b.halted = why
		b.refused++
		return budgetError(p, why)
	}
	return nil
}
