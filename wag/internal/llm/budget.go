package llm

import (
	"fmt"
	"sync"
)

// Budget is the hard limit the owner of a run sets on what its calls may
// spend (D046): calls, counting every attempt and so every retry; tokens, as
// providers report them; and cost, from a price table. A zero limit is no
// limit. Before each attempt Reserve refuses a call whose worst case would
// exceed what is left, so a run stops before it overspends, not after.
// After the attempt Charge records what the provider reported, or the worst
// case where it reported nothing. A Budget is safe for concurrent use, and
// one Budget may serve every client of a run.
type Budget struct {
	MaxCalls  int
	MaxTokens int
	MaxCost   float64
	// Prices are per million tokens, by profile name, then by model; the
	// key "*" prices any other. A cost limit needs a price for every
	// profile that calls.
	Prices map[string]Price

	mu      sync.Mutex
	spent   Spent
	pending Spent // the worst cases of the attempts under way
	refused int   // calls refused
}

// Price is what a million tokens cost, in the price table's currency.
type Price struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"` // output and reasoning tokens
}

// Spent is what a run's calls have used.
type Spent struct {
	Calls  int     `json:"calls"`
	Tokens int     `json:"tokens"`
	Cost   float64 `json:"cost"`
}

// BudgetError is why the budget refused a call. Reserve returns it inside an
// *Error of KindBudget, which is not retryable; errors.As finds it.
type BudgetError struct {
	Profile string
	Reason  string
}

func (e *BudgetError) Error() string {
	return fmt.Sprintf("budget (D046): %s: %s", e.Profile, e.Reason)
}

// Limited reports whether the budget sets any limit.
func (b *Budget) Limited() bool {
	return b != nil && (b.MaxCalls > 0 || b.MaxTokens > 0 || b.MaxCost > 0)
}

// Refused reports whether the budget has refused a call: the run has
// reached a limit.
func (b *Budget) Refused() bool {
	if b == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.refused > 0
}

// Spent returns what the run has used so far.
func (b *Budget) Spent() Spent {
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

// worst is the most an attempt of the request can use: its input's bytes,
// which no tokenizer turns into more tokens, and the output the profile
// allows. Gemini reports its reasoning apart from the answer, so a fixed
// thinking budget is added there; other providers count reasoning within
// the output limit.
func worst(p Profile, req Request) (in, out int, err error) {
	in = len(req.System)
	for _, m := range req.Messages {
		in += len(m.Content)
	}
	out = req.MaxOutputTokens
	if out <= 0 {
		return 0, 0, fmt.Errorf("no max_output_tokens, so the worst case of a call is unknown")
	}
	if p.Provider == "gemini" {
		tb := req.ThinkingBudget
		switch {
		case tb != nil && *tb == 0:
		case tb != nil && *tb > 0:
			out += *tb
		default:
			return 0, 0, fmt.Errorf("Gemini may reason without a limit: set thinking_budget (0 turns it off)")
		}
	}
	return in, out, nil
}

// Reserve is called before each attempt. It returns the attempt's worst
// case, held against the limits until Charge is given it, or an *Error of
// KindBudget if that worst case would go beyond a limit.
func (b *Budget) Reserve(p Profile, req Request) (Spent, error) {
	if !b.Limited() {
		return Spent{}, nil
	}
	// refuse builds the refusal; the caller counts it (b.refused) under the
	// lock it holds or takes.
	refuse := func(why string, a ...any) error {
		be := &BudgetError{Profile: p.Name, Reason: fmt.Sprintf(why, a...)}
		return &Error{Provider: p.Provider, Kind: KindBudget, Message: be.Error(), Err: be}
	}
	w := Spent{Calls: 1}
	if b.MaxTokens > 0 || b.MaxCost > 0 {
		in, out, err := worst(p, req)
		if err != nil {
			b.mu.Lock()
			b.refused++
			b.mu.Unlock()
			return Spent{}, refuse("%v", err)
		}
		w.Tokens = in + out
		if b.MaxCost > 0 {
			pr, ok := b.price(p)
			if !ok {
				b.mu.Lock()
				b.refused++
				b.mu.Unlock()
				return Spent{}, refuse("no price for profile %s or model %s, so the cost of its calls is unknown", p.Name, p.Params.Model)
			}
			w.Cost = (float64(in)*pr.Input + float64(out)*pr.Output) / 1e6
		}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	at := Spent{b.spent.Calls + b.pending.Calls, b.spent.Tokens + b.pending.Tokens, b.spent.Cost + b.pending.Cost}
	var why error
	switch {
	case b.MaxCalls > 0 && at.Calls+w.Calls > b.MaxCalls:
		why = refuse("the run's %d calls are used", b.MaxCalls)
	case b.MaxTokens > 0 && at.Tokens+w.Tokens > b.MaxTokens:
		why = refuse("this call could use %d tokens, and %d of the run's %d are left", w.Tokens, b.MaxTokens-at.Tokens, b.MaxTokens)
	case b.MaxCost > 0 && at.Cost+w.Cost > b.MaxCost:
		why = refuse("this call could cost %.4f, and %.4f of the run's %.4f is left", w.Cost, b.MaxCost-at.Cost, b.MaxCost)
	}
	if why != nil {
		b.refused++ // b.mu is held
		return Spent{}, why
	}
	b.pending.Calls += w.Calls
	b.pending.Tokens += w.Tokens
	b.pending.Cost += w.Cost
	return w, nil
}

// Charge records an attempt: it releases the reservation w that Reserve
// returned and adds what the provider reported. A count the provider did not
// report is taken at its worst case. An attempt that failed is charged as a
// call and at its worst case, since a provider may bill a failed call.
func (b *Budget) Charge(p Profile, req Request, w Spent, resp *Response) {
	if !b.Limited() {
		return
	}
	used := w
	if resp != nil && w.Tokens > 0 {
		in, out, _ := worst(p, req)
		u := resp.Usage
		gotIn, gotOut := in, out
		if u.InputTokens != nil {
			gotIn = *u.InputTokens
		}
		if u.OutputTokens != nil {
			gotOut = *u.OutputTokens
			if p.Provider == "gemini" && u.ReasoningTokens != nil {
				gotOut += *u.ReasoningTokens
			}
		}
		if u.TotalTokens != nil && u.InputTokens != nil {
			gotOut = *u.TotalTokens - *u.InputTokens
		}
		used.Tokens = gotIn + gotOut
		if b.MaxCost > 0 {
			pr, _ := b.price(p)
			used.Cost = (float64(gotIn)*pr.Input + float64(gotOut)*pr.Output) / 1e6
		}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pending.Calls -= w.Calls
	b.pending.Tokens -= w.Tokens
	b.pending.Cost -= w.Cost
	b.spent.Calls += used.Calls
	b.spent.Tokens += used.Tokens
	b.spent.Cost += used.Cost
}
