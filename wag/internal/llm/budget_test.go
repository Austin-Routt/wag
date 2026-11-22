package llm

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// budgetClient is a client whose provider counts the calls that reach it and
// answers with the given usage.
func budgetClient(p Profile, b *Budget, usage Usage, fail int) (*Client, *int64) {
	var calls int64
	c := WithProvider(p, &Mock{Fn: func(req Request) (*Response, error) {
		n := atomic.AddInt64(&calls, 1)
		if int(n) <= fail {
			return nil, &Error{Provider: "mock", Kind: KindServer, Message: "busy"}
		}
		return &Response{Text: "ok", Usage: usage}, nil
	}})
	c.Sleep, c.Budget = NoWait, b
	return c, &calls
}

func ask(c *Client, text string) error {
	_, err := c.Complete(context.Background(), c.Request("sys", []Message{{Role: "user", Content: text}}), nil)
	return err
}

func refusedByBudget(t *testing.T, err error) {
	t.Helper()
	var be *BudgetError
	e, _ := err.(*Error)
	if e == nil || e.Kind != KindBudget || e.Retryable() || !errors.As(err, &be) {
		t.Fatalf("want a budget refusal, got %v", err)
	}
}

// Every attempt is a call, retries included, and the call that would pass
// the limit is refused before it reaches the provider.
func TestBudgetCountsRetriesAsCalls(t *testing.T) {
	b := &Budget{MaxCalls: 2}
	c, calls := budgetClient(Profile{Name: "p", Provider: "mock"}, b, Usage{}, 2)
	refusedByBudget(t, ask(c, "x"))
	if *calls != 2 || b.Spent().Calls != 2 {
		t.Fatalf("provider saw %d calls, budget counted %d; want 2 and 2", *calls, b.Spent().Calls)
	}
}

// The preflight uses the worst case: the input's bytes and the output limit.
// A call that could pass the token limit is refused, however little it
// would really use; what is charged afterwards is what was reported.
func TestBudgetTokensPreflightAndCharge(t *testing.T) {
	p := Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 100}}
	b := &Budget{MaxTokens: 250}
	c, calls := budgetClient(p, b, Usage{InputTokens: intp(5), OutputTokens: intp(10)}, 0)
	if err := ask(c, "hello"); err != nil { // worst 3+5+100 = 108
		t.Fatal(err)
	}
	if got := b.Spent().Tokens; got != 15 {
		t.Fatalf("charged %d tokens, want the 15 reported", got)
	}
	if err := ask(c, "hello"); err != nil { // 15 spent, worst 108: 123 <= 250
		t.Fatal(err)
	}
	refusedByBudget(t, ask(c, strings.Repeat("x", 200))) // 30 spent, worst 303
	if *calls != 2 {
		t.Fatalf("the refused call reached the provider (%d calls)", *calls)
	}
}

// A provider that reports nothing is charged the worst case.
func TestBudgetUnreportedUsageIsWorstCase(t *testing.T) {
	p := Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 50}}
	b := &Budget{MaxTokens: 1000}
	c, _ := budgetClient(p, b, Usage{}, 0)
	if err := ask(c, "abc"); err != nil {
		t.Fatal(err)
	}
	if got := b.Spent().Tokens; got != 3+3+50 {
		t.Fatalf("charged %d, want the worst case 56", got)
	}
}

// A cost limit needs a price, and refuses the call whose worst-case cost
// would pass it; the reported usage is priced afterwards.
func TestBudgetCost(t *testing.T) {
	p := Profile{Name: "p", Provider: "mock", Params: Params{Model: "m", MaxOutputTokens: 1000}}
	c, calls := budgetClient(p, &Budget{MaxCost: 1}, Usage{}, 0)
	refusedByBudget(t, ask(c, "x"))
	if *calls != 0 {
		t.Fatal("a call without a price reached the provider")
	}
	// $1 per million input and $10 per million output tokens: the worst case
	// of a call is about (4 + 1000*10)/1e6 = $0.01.
	b := &Budget{MaxCost: 0.015, Prices: map[string]Price{"m": {Input: 1, Output: 10}}}
	c, calls = budgetClient(p, b, Usage{InputTokens: intp(4), OutputTokens: intp(500)}, 0)
	if err := ask(c, "x"); err != nil {
		t.Fatal(err)
	}
	if got := b.Spent().Cost; got < 0.005 || got > 0.00501 {
		t.Fatalf("charged %.6f, want 0.005004", got)
	}
	// 0.005004 spent and a worst case of 0.010004 would pass 0.015.
	refusedByBudget(t, ask(c, "x"))
	if *calls != 1 {
		t.Fatalf("the refused call reached the provider (%d calls)", *calls)
	}
}

// Gemini reports its reasoning apart from the answer, so a token or cost
// limit needs a fixed thinking budget, which counts in the worst case.
func TestBudgetGeminiReasoning(t *testing.T) {
	p := Profile{Name: "g", Provider: "gemini", Params: Params{MaxOutputTokens: 100}}
	c, _ := budgetClient(p, &Budget{MaxTokens: 10000}, Usage{}, 0)
	refusedByBudget(t, ask(c, "x"))
	p.Params.ThinkingBudget = intp(0)
	c, _ = budgetClient(p, &Budget{MaxTokens: 10000}, Usage{}, 0)
	if err := ask(c, "x"); err != nil {
		t.Fatal(err)
	}
	p.Params.ThinkingBudget = intp(500)
	b := &Budget{MaxTokens: 10000}
	c, _ = budgetClient(p, b, Usage{}, 0)
	if err := ask(c, "x"); err != nil {
		t.Fatal(err)
	}
	if got := b.Spent().Tokens; got != 3+1+100+500 {
		t.Fatalf("charged %d, want the worst case with reasoning, 604", got)
	}
	p.Params.MaxOutputTokens = 0
	c, _ = budgetClient(p, &Budget{MaxTokens: 10000}, Usage{}, 0)
	refusedByBudget(t, ask(c, "x")) // no output limit: the worst case is unknown
}

// Concurrent calls hold their worst cases while under way, so together they
// cannot pass the limit.
func TestBudgetConcurrentReservations(t *testing.T) {
	p := Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 97}}
	b := &Budget{MaxTokens: 1000} // ten worst cases of 100
	var wg sync.WaitGroup
	var passed int64
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if held, err := b.Reserve(p, Request{Messages: []Message{{Content: "abc"}}, Params: p.Params}); err == nil {
				atomic.AddInt64(&passed, 1)
				_ = held // never charged: they stay under way
			}
		}()
	}
	wg.Wait()
	if passed != 10 {
		t.Fatalf("%d calls reserved at once, want 10", passed)
	}
}

// No budget, or one without limits, changes nothing.
func TestNoBudget(t *testing.T) {
	c, calls := budgetClient(Profile{Name: "p", Provider: "mock"}, nil, Usage{}, 0)
	if err := ask(c, "x"); err != nil || *calls != 1 {
		t.Fatal(err)
	}
	c, _ = budgetClient(Profile{Name: "p", Provider: "mock"}, &Budget{}, Usage{}, 0)
	if err := ask(c, "x"); err != nil {
		t.Fatal(err)
	}
}

// Refused reports that a limit was reached, whichever check refused.
func TestBudgetRefused(t *testing.T) {
	b := &Budget{MaxCalls: 1}
	c, _ := budgetClient(Profile{Name: "p", Provider: "mock"}, b, Usage{}, 0)
	if err := ask(c, "x"); err != nil || b.Refused() {
		t.Fatalf("first call: %v, refused %v", err, b.Refused())
	}
	refusedByBudget(t, ask(c, "x"))
	if !b.Refused() {
		t.Fatal("a refusal was not recorded")
	}
	nb := &Budget{MaxTokens: 10} // no max_output_tokens: refused before the lock
	c, _ = budgetClient(Profile{Name: "p", Provider: "mock"}, nb, Usage{}, 0)
	refusedByBudget(t, ask(c, "x"))
	if !nb.Refused() || (*Budget)(nil).Refused() {
		t.Fatal("Refused is wrong")
	}
}
