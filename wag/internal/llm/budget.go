package llm

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

// Budget is one invocation's account of its model calls (D046). It always
// counts: every client attempt, retries included; the tokens providers
// report; their cost where the price table has the profile; and the attempts
// with no reported usage. Limits are optional, and none is the default (the
// owner, 2026-11-23). MaxCalls caps attempts exactly. MaxTokens and MaxCost
// reserve an estimate of each attempt before it is sent and refuse one that
// would pass them: estimates, not provider billing ceilings, since hidden
// framing and provider behavior are not measured here. A zero limit is no
// limit. The configuration must not change while the budget is in use; one
// budget may serve concurrent clients. Nothing is restored from an earlier
// invocation or a resumed annotation directory.
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

// Spent is what an invocation's attempts used. Calls counts every attempt.
// Tokens and Cost are what providers reported, priced from the table; under a
// token or cost limit, a count that was not reported is taken at its estimate
// (the trace keeps the raw usage). Unreported counts the attempts, failed ones
// included, for which no usage was reported: without a token or cost limit
// their tokens are not in Tokens. Unpriced counts the attempts with usage
// that a given price table has no rate for: their cost is not in Cost.
type Spent struct {
	Calls      int     `json:"calls"`
	Tokens     int     `json:"tokens"`
	Cost       float64 `json:"cost"`
	Unreported int     `json:"unreported"`
	Unpriced   int     `json:"unpriced"`
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

// Limited reports whether the budget sets any limit.
func (b *Budget) Limited() bool {
	return b != nil && (b.MaxCalls > 0 || b.MaxTokens > 0 || b.MaxCost > 0)
}

// estimates reports whether attempts are reserved at an estimate: under a
// token or cost limit.
func (b *Budget) estimates() bool {
	return b != nil && (b.MaxTokens > 0 || b.MaxCost > 0)
}

// Halted is the accounting violation that stopped the invocation's later
// attempts ("" if none).
func (b *Budget) Halted() string {
	if b == nil {
		return ""
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.halted
}

// Refused reports whether the budget has stopped the invocation: it refused
// a call, or an accounting violation halted later attempts.
func (b *Budget) Refused() bool {
	if b == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.refused > 0 || b.halted != ""
}

// Refusals is how many attempts the budget refused before they were sent.
func (b *Budget) Refusals() int {
	if b == nil {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.refused
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

// costSlack is the relative tolerance of a cost limit: a sum of float costs
// (0.1+0.1+0.1 is 0.30000000000000004) must not refuse the call that only
// reaches the limit.
const costSlack = 1e-9

func priced(in, out int, p Price) float64 {
	return float64(in)/1e6*p.Input + float64(out)/1e6*p.Output
}

// Reserve is called before each attempt. It holds the attempt (and, under a
// token or cost limit, its estimate) until Charge, and refuses the attempt
// before it is sent if that would pass a limit. Without limits it refuses
// nothing.
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
	w := Spent{Calls: 1}
	if b.estimates() {
		in, out, err := estimate(p, req)
		if err != nil {
			return refuse(err.Error())
		}
		w.Tokens = in + out // estimate checked this sum
		pr, ok := b.price(p)
		if !ok && b.MaxCost > 0 {
			return refuse("no price for this profile/model, so estimated cost is unknown")
		}
		if ok {
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
	case b.MaxCost > 0 && cost > b.MaxCost*(1+costSlack):
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
// components where absent. Gemini counts its thoughts apart from the answer;
// the other providers count reasoning within the completion, except that
// reasoning larger than the completion cannot be part of it, and is counted
// apart (xAI reports it so). A reported total with a reported input supplies
// the output: a total below the parts that are known is a violation, and
// what a total adds to them (reasoning reported apart, Gemini's tool-use
// prompt) is charged as output, so a larger total never credits. Without an
// input count, a total less a complete output is the input. inKnown says
// whether the split between input and output is the provider's.
func reportedUsage(p Profile, req Request, u Usage, in, out int) (int, int, bool, error) {
	for _, count := range []*int{u.InputTokens, u.OutputTokens, u.ReasoningTokens, u.CachedTokens, u.TotalTokens} {
		if count != nil && *count < 0 {
			return 0, 0, false, fmt.Errorf("negative usage count")
		}
	}
	inKnown := u.InputTokens != nil
	if inKnown {
		in = *u.InputTokens
	}
	// known is the output the provider reported in parts.
	known, ok := 0, true
	if p.Provider == "gemini" {
		answer, reasoning := req.MaxOutputTokens, *req.ThinkingBudget
		if u.OutputTokens != nil {
			answer = *u.OutputTokens
			known = *u.OutputTokens
		}
		if u.ReasoningTokens != nil {
			reasoning = *u.ReasoningTokens
			if known, ok = addCount(known, *u.ReasoningTokens); !ok {
				return 0, 0, false, fmt.Errorf("reported output overflows accounting")
			}
		}
		if out, ok = addCount(answer, reasoning); !ok {
			return 0, 0, false, fmt.Errorf("output/reasoning usage overflows accounting")
		}
	} else {
		if u.OutputTokens != nil {
			out = *u.OutputTokens
			known = *u.OutputTokens
		}
		if u.ReasoningTokens != nil {
			if u.OutputTokens != nil && *u.ReasoningTokens > *u.OutputTokens {
				if out, ok = addCount(*u.OutputTokens, *u.ReasoningTokens); !ok {
					return 0, 0, false, fmt.Errorf("output/reasoning usage overflows accounting")
				}
				known = out
			} else {
				out = max(out, *u.ReasoningTokens)
				known = max(known, *u.ReasoningTokens)
			}
		}
	}
	if u.TotalTokens != nil {
		minimum := known
		if u.InputTokens != nil {
			if minimum, ok = addCount(*u.InputTokens, known); !ok {
				return 0, 0, false, fmt.Errorf("reported usage overflows accounting")
			}
		}
		if *u.TotalTokens < minimum {
			return 0, 0, false, fmt.Errorf("total usage is below reported components")
		}
		if u.InputTokens != nil {
			out = *u.TotalTokens - *u.InputTokens
		} else if outComplete(p, u) {
			in = *u.TotalTokens - known
			inKnown = true
		}
	}
	if _, ok := addCount(in, out); !ok {
		return 0, 0, false, fmt.Errorf("reported usage overflows accounting")
	}
	return in, out, inKnown, nil
}

// Charge must be called exactly once for a successful reservation. Without a
// token or cost limit it counts what was reported, and stops nothing. Under
// one, failed calls retain their estimate; malformed usage retains at least
// that estimate and any larger nonnegative observed components. An
// accounting violation stops future attempts, but cannot undo
// already-dispatched requests or guarantee billing.
func (b *Budget) Charge(p Profile, req Request, w Spent, resp *Response) error {
	if b == nil {
		return nil
	}
	if !b.estimates() {
		b.count(p, w, resp)
		return nil
	}
	used := w
	why := ""
	unpriced := false
	if resp == nil {
		// A failed attempt keeps its reservation, priced when it was made.
		if _, ok := b.price(p); !ok && len(b.Prices) > 0 {
			unpriced = true
		}
	} else {
		in, out, err := estimate(p, req)
		if err != nil {
			why = err.Error()
		} else {
			u := resp.Usage
			gotIn, gotOut, inKnown, usageErr := reportedUsage(p, req, u, in, out)
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
					// Reasoning apart from the answer (Gemini's, or larger
					// than the completion) adds to it.
					if p.Provider == "gemini" || *u.ReasoningTokens > observedOut {
						observedOut, _ = addCount(observedOut, *u.ReasoningTokens)
					}
				}
				gotOut = max(gotOut, observedOut)
			}
			used.Tokens, _ = addCount(gotIn, gotOut) // saturates on invalid/overflowing observations
			if u.TotalTokens != nil && *u.TotalTokens > used.Tokens {
				used.Tokens = *u.TotalTokens
			}
			pr, ok := b.price(p)
			if !ok && len(b.Prices) > 0 {
				unpriced = true
			}
			if ok {
				used.Cost = priced(gotIn, gotOut, pr)
				// Without the provider's split, the total is priced at the
				// dearer rate: an upper bound.
				if u.TotalTokens != nil && *u.TotalTokens > 0 && (usageErr != nil || !inKnown) {
					used.Cost = math.Max(used.Cost, float64(*u.TotalTokens)/1e6*math.Max(pr.Input, pr.Output))
				}
				if !finiteNonnegative(used.Cost) {
					used.Cost = math.MaxFloat64
					why = "reported cost overflows accounting"
				}
			}
			// The cost is compared only under a cost limit: without one it is
			// recorded, not enforced.
			if used.Tokens > w.Tokens || (b.MaxCost > 0 && used.Cost > w.Cost) {
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
	if resp == nil || !reportedAny(resp.Usage) {
		b.spent.Unreported++
	}
	if unpriced {
		b.spent.Unpriced++
	}
	if why != "" {
		b.halted = why
		return budgetError(p, why)
	}
	return nil
}

func usable(n *int) bool { return n != nil && *n >= 0 }

// reportedAny says whether a usage reports a count of tokens: input, output,
// reasoning or total.
func reportedAny(u Usage) bool {
	return usable(u.InputTokens) || usable(u.OutputTokens) || usable(u.ReasoningTokens) || usable(u.TotalTokens)
}

// outComplete says whether a usage reports the whole output in parts: the
// completion, and for Gemini its thoughts.
func outComplete(p Profile, u Usage) bool {
	return usable(u.OutputTokens) && (p.Provider != "gemini" || usable(u.ReasoningTokens))
}

// count records an attempt of a budget that estimates nothing (no limit, or
// a call cap alone): the call, and the usage the provider reported, the
// larger of its total and its parts, with reasoning added where it is apart
// from the answer (as in reportedUsage), priced where the table has the
// profile. Without an input count, a total less a complete output is the
// input; without either, the count is priced at the dearer rate. A negative
// count is not usage. It refuses and stops nothing.
func (b *Budget) count(p Profile, w Spent, resp *Response) {
	tokens, cost, unpriced := 0, 0.0, false
	reported := resp != nil && reportedAny(resp.Usage)
	if reported {
		get := func(n *int) int {
			if !usable(n) {
				return 0
			}
			return *n
		}
		u := resp.Usage
		in, out, reasoning, total := get(u.InputTokens), get(u.OutputTokens), get(u.ReasoningTokens), get(u.TotalTokens)
		if p.Provider == "gemini" || reasoning > out {
			out, _ = addCount(out, reasoning)
		}
		inKnown := usable(u.InputTokens)
		if !inKnown && usable(u.TotalTokens) && outComplete(p, u) && total >= out {
			in, inKnown = total-out, true
		}
		tokens, _ = addCount(in, out)
		tokens = max(tokens, total)
		if pr, ok := b.price(p); ok {
			if inKnown {
				cost = priced(in, tokens-in, pr) // a total's excess over the parts is output
			} else {
				cost = float64(tokens) / 1e6 * math.Max(pr.Input, pr.Output)
			}
		} else if len(b.Prices) > 0 {
			unpriced = true
		}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pending.Calls -= w.Calls
	b.spent.Calls, _ = addCount(b.spent.Calls, w.Calls)
	b.spent.Tokens, _ = addCount(b.spent.Tokens, tokens)
	// Finite rates give a finite or infinite cost, never NaN: an overflow
	// is capped.
	b.spent.Cost = math.Min(b.spent.Cost+cost, math.MaxFloat64)
	if !reported {
		b.spent.Unreported++
	}
	if unpriced {
		b.spent.Unpriced++
	}
}
