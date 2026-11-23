package llm

import (
	"context"
	"math"
	"sync"
	"testing"
)

// With no limit (D046's default, the owner, 2026-11-23) nothing is refused,
// and every attempt is still counted: the calls, retries included; the
// tokens providers report (the larger of the total and its parts); the cost
// where the price table has the profile; and the attempts with no reported
// usage, failed ones included, whose tokens are not in the count.
func TestBudgetUnlimitedCounts(t *testing.T) {
	b := &Budget{Prices: map[string]Price{"*": {Input: 1, Output: 2}}}
	c, calls := budgetClient(Profile{Name: "p", Provider: "mock"}, b, Usage{InputTokens: intp(4), OutputTokens: intp(5), TotalTokens: intp(10)}, 1)
	if err := ask(c, "x"); err != nil {
		t.Fatal(err)
	}
	c2, _ := budgetClient(Profile{Name: "q", Provider: "mock"}, b, Usage{}, 0)
	if err := ask(c2, "x"); err != nil {
		t.Fatal(err)
	}
	s := b.Spent()
	// 4 input at 1 and 6 output (the total's excess counts as output) at 2,
	// per million.
	if *calls != 2 || s.Calls != 3 || s.Tokens != 10 || s.Unreported != 2 || b.Refused() || math.Abs(s.Cost-16e-6) > 1e-12 {
		t.Fatalf("spent %+v (provider calls %d); want 3 calls (a retry included), 10 tokens, 2 unreported (the failed attempt and the silent one), cost 16e-6", s, *calls)
	}
	if b.Limited() || b.Halted() != "" {
		t.Fatalf("an unlimited budget is limited or halted: %q", b.Halted())
	}
}

// Malformed usage cannot stop an unlimited run: negative counts are not
// usage, and a count that is reported is kept.
func TestBudgetUnlimitedMalformedUsage(t *testing.T) {
	b := &Budget{}
	c, calls := budgetClient(Profile{Name: "p", Provider: "mock"}, b, Usage{InputTokens: intp(-10), OutputTokens: intp(7)}, 0)
	for i := 0; i < 3; i++ {
		if err := ask(c, "x"); err != nil {
			t.Fatal(err)
		}
	}
	if s := b.Spent(); *calls != 3 || s.Calls != 3 || s.Tokens != 21 || s.Unreported != 0 || b.Refused() {
		t.Fatalf("spent %+v, provider calls %d", s, *calls)
	}
}

// Concurrent unlimited attempts are all counted, and none stays pending.
func TestBudgetUnlimitedConcurrent(t *testing.T) {
	b := &Budget{}
	c, calls := budgetClient(Profile{Name: "p", Provider: "mock"}, b, Usage{InputTokens: intp(1), OutputTokens: intp(2)}, 0)
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ask(c, "x"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	b.mu.Lock()
	pending := b.pending
	b.mu.Unlock()
	if s := b.Spent(); *calls != 40 || s.Calls != 40 || s.Tokens != 120 || pending != (Spent{}) {
		t.Fatalf("spent %+v, pending %+v, provider calls %d", s, pending, *calls)
	}
}

// A call cap alone estimates nothing, and still counts what is reported.
func TestBudgetCallCapCountsReportedUsage(t *testing.T) {
	b := &Budget{MaxCalls: 5}
	c, _ := budgetClient(Profile{Name: "p", Provider: "mock"}, b, Usage{InputTokens: intp(4), OutputTokens: intp(5), TotalTokens: intp(9)}, 0)
	for i := 0; i < 2; i++ {
		if err := ask(c, "x"); err != nil {
			t.Fatal(err)
		}
	}
	if s := b.Spent(); s.Calls != 2 || s.Tokens != 18 || s.Unreported != 0 {
		t.Fatalf("spent %+v; want 2 calls and 18 reported tokens", s)
	}
}

// Under a token limit, an attempt with no reported usage is charged its
// estimate and counted as unreported; with a price table and no cost limit
// the cost is recorded, and never compared with a reservation it did not
// make.
func TestBudgetEstimateModeUnreportedAndPrices(t *testing.T) {
	b := &Budget{MaxTokens: 1000, Prices: map[string]Price{"*": {Input: 1, Output: 2}}}
	c, _ := budgetClient(Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 50}}, b, Usage{}, 0)
	if err := ask(c, "x"); err != nil {
		t.Fatal(err)
	}
	c2, _ := budgetClient(Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 50}}, b, Usage{InputTokens: intp(4), OutputTokens: intp(5)}, 0)
	if err := ask(c2, "x"); err != nil {
		t.Fatal(err)
	}
	// A failed attempt, not retried, keeps its reservation, priced too.
	c3, _ := budgetClient(Profile{Name: "p", Provider: "mock", Retries: -1, Params: Params{MaxOutputTokens: 50}}, b, Usage{}, 1)
	if err := ask(c3, "x"); err == nil {
		t.Fatal("the failed attempt succeeded")
	}
	s := b.Spent()
	// The first and the third are charged their estimate, 3+1 input bytes
	// and 50 output; the second what it reported, 4 and 5.
	want := (2*(4.0*1+50*2) + 4*1 + 5*2) / 1e6
	if s.Calls != 3 || s.Tokens != 2*54+9 || s.Unreported != 2 || math.Abs(s.Cost-want) > 1e-12 || b.Refused() {
		t.Fatalf("spent %+v; want 3 calls, 117 tokens, 2 unreported, cost %g, nothing refused", s, want)
	}
}

// Reasoning reported apart from the completion count (xAI: the total is
// prompt + completion + reasoning) is charged, not treated as a violation;
// the patch as received halted the run here.
func TestBudgetReasoningOutsideCompletion(t *testing.T) {
	p := Profile{Name: "x", Provider: "openai-compatible", Params: Params{MaxOutputTokens: 2000}}
	b := &Budget{MaxTokens: 10000}
	c, calls := budgetClient(p, b, Usage{InputTokens: intp(32), OutputTokens: intp(9), ReasoningTokens: intp(94), TotalTokens: intp(135)}, 0)
	if err := ask(c, "x"); err != nil {
		t.Fatal(err)
	}
	if err := ask(c, "x"); err != nil || *calls != 2 || b.Spent().Tokens != 270 {
		t.Fatalf("err %v, calls %d, spent %+v; want two calls charged 135 each", err, *calls, b.Spent())
	}
	nb := &Budget{MaxTokens: 10000} // the same without a total: both counts are charged
	c, _ = budgetClient(p, nb, Usage{InputTokens: intp(32), OutputTokens: intp(9), ReasoningTokens: intp(94)}, 0)
	if err := ask(c, "x"); err != nil || nb.Spent().Tokens != 135 {
		t.Fatalf("err %v, spent %+v; want 135", err, nb.Spent())
	}
	sb := &Budget{MaxTokens: 10000} // reasoning apart but smaller than the answer: the total tells
	c, _ = budgetClient(p, sb, Usage{InputTokens: intp(32), OutputTokens: intp(100), ReasoningTokens: intp(50), TotalTokens: intp(182)}, 0)
	if err := ask(c, "x"); err != nil || sb.Spent().Tokens != 182 {
		t.Fatalf("err %v, spent %+v; want 182", err, sb.Spent())
	}
	ib := &Budget{MaxTokens: 10000} // reasoning within the completion (OpenAI): not counted twice
	c, _ = budgetClient(p, ib, Usage{InputTokens: intp(32), OutputTokens: intp(100), ReasoningTokens: intp(50), TotalTokens: intp(132)}, 0)
	if err := ask(c, "x"); err != nil || ib.Spent().Tokens != 132 {
		t.Fatalf("err %v, spent %+v; want 132", err, ib.Spent())
	}
	// A total below the parts that are known, reasoning apart included, is
	// still a violation: charged no less than its estimate, and it stops the run.
	vb := &Budget{MaxTokens: 10000}
	c, calls = budgetClient(p, vb, Usage{InputTokens: intp(32), OutputTokens: intp(9), ReasoningTokens: intp(94), TotalTokens: intp(110)}, 0)
	refusedByBudget(t, ask(c, "x"))
	if s := vb.Spent(); s.Tokens < 3+1+2000 || vb.Halted() == "" {
		t.Fatalf("spent %+v, halted %q", s, vb.Halted())
	}
	refusedByBudget(t, ask(c, "x"))
	if *calls != 1 {
		t.Fatal("a violation did not stop later attempts")
	}
}

// Gemini's total can count more than the answer and the thoughts (a tool-use
// prompt): the excess is charged as output, not treated as a violation.
func TestBudgetGeminiTotalAboveParts(t *testing.T) {
	p := Profile{Name: "g", Provider: "gemini", Params: Params{MaxOutputTokens: 100, ThinkingBudget: intp(50)}}
	b := &Budget{MaxTokens: 10000, MaxCost: 1, Prices: map[string]Price{"*": {Input: 1, Output: 2}}}
	c, calls := budgetClient(p, b, Usage{InputTokens: intp(10), OutputTokens: intp(5), ReasoningTokens: intp(3), TotalTokens: intp(25)}, 0)
	for i := 0; i < 2; i++ {
		if err := ask(c, "x"); err != nil {
			t.Fatal(err)
		}
	}
	if s := b.Spent(); *calls != 2 || s.Tokens != 50 || math.Abs(s.Cost-2*(10*1+15*2)/1e6) > 1e-12 {
		t.Fatalf("spent %+v, calls %d; want 50 tokens and cost 80e-6", s, *calls)
	}
}

// The dashboard's clients have no budget: a nil budget refuses and counts
// nothing.
func TestBudgetNilIsNoBudget(t *testing.T) {
	var b *Budget
	c, calls := budgetClient(Profile{Name: "p", Provider: "mock"}, b, Usage{InputTokens: intp(1)}, 0)
	if err := ask(c, "x"); err != nil || *calls != 1 || b.Spent() != (Spent{}) || b.Refused() || b.Halted() != "" || b.Limited() {
		t.Fatalf("a nil budget: %v, %d calls", err, *calls)
	}
	if _, err := c.Complete(context.Background(), c.Request("s", []Message{{Role: "user", Content: "x"}}), nil); err != nil {
		t.Fatal(err)
	}
}

// Reasoning equal to the completion can be all of it (a reasoning model that
// used its whole output on thought): it is not counted twice, with or
// without a total.
func TestBudgetReasoningEqualToCompletion(t *testing.T) {
	p := Profile{Name: "o", Provider: "openai", Params: Params{MaxOutputTokens: 100}}
	for _, u := range []Usage{
		{InputTokens: intp(10), OutputTokens: intp(50), ReasoningTokens: intp(50)},
		{InputTokens: intp(10), OutputTokens: intp(50), ReasoningTokens: intp(50), TotalTokens: intp(60)},
	} {
		for _, b := range []*Budget{{MaxTokens: 1000}, {}} {
			c, _ := budgetClient(p, b, u, 0)
			if err := ask(c, "x"); err != nil || b.Spent().Tokens != 60 {
				t.Fatalf("usage %+v, limit %d: %v, spent %+v; want 60", u, b.MaxTokens, err, b.Spent())
			}
		}
	}
}

// Without an estimate, reasoning is counted as reportedUsage counts it:
// apart for Gemini and where it is larger than the completion (xAI), within
// the completion otherwise.
func TestBudgetCountsReasoningLikeTheLimits(t *testing.T) {
	for _, tc := range []struct {
		provider string
		u        Usage
		want     int
	}{
		{"openai-compatible", Usage{InputTokens: intp(32), OutputTokens: intp(9), ReasoningTokens: intp(94)}, 135},
		{"gemini", Usage{InputTokens: intp(10), OutputTokens: intp(5), ReasoningTokens: intp(3)}, 18},
		{"openai", Usage{InputTokens: intp(32), OutputTokens: intp(100), ReasoningTokens: intp(50)}, 132},
	} {
		b := &Budget{}
		c, _ := budgetClient(Profile{Name: "p", Provider: tc.provider}, b, tc.u, 0)
		if err := ask(c, "x"); err != nil || b.Spent().Tokens != tc.want {
			t.Errorf("%s %+v: %v, spent %+v; want %d", tc.provider, tc.u, err, b.Spent(), tc.want)
		}
	}
}

// Under a token limit alone, a price table records the cost; a cost above
// its reservation stops nothing, since no cost limit was set. Under a cost
// limit the same cost is an overrun.
func TestBudgetCostIsEnforcedOnlyUnderACostLimit(t *testing.T) {
	p := Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 50}}
	prices := map[string]Price{"*": {Input: 10, Output: 1}}
	// The estimate is 4 input and 50 output tokens; the report 50 and 4: no
	// more tokens, but a dearer input.
	u := Usage{InputTokens: intp(50), OutputTokens: intp(4)}
	b := &Budget{MaxTokens: 1000, Prices: prices}
	c, calls := budgetClient(p, b, u, 0)
	for i := 0; i < 2; i++ {
		if err := ask(c, "x"); err != nil {
			t.Fatalf("a recorded cost stopped a run with no cost limit: %v", err)
		}
	}
	if s := b.Spent(); *calls != 2 || math.Abs(s.Cost-2*504e-6) > 1e-12 || b.Halted() != "" {
		t.Fatalf("spent %+v, calls %d, halted %q", s, *calls, b.Halted())
	}
	cb := &Budget{MaxTokens: 1000, MaxCost: 1, Prices: prices}
	c, _ = budgetClient(p, cb, u, 0)
	refusedByBudget(t, ask(c, "x"))
	if cb.Halted() == "" {
		t.Fatal("a cost above its reservation under a cost limit did not stop the run")
	}
}

// When usage is malformed, the fallback keeps reasoning reported apart from
// a smaller completion as well as the completion.
func TestBudgetViolationKeepsReasoningApart(t *testing.T) {
	b := &Budget{MaxTokens: 10000}
	c, _ := budgetClient(Profile{Name: "x", Provider: "openai-compatible", Params: Params{MaxOutputTokens: 100}}, b,
		Usage{InputTokens: intp(-1), OutputTokens: intp(9), ReasoningTokens: intp(94)}, 0)
	refusedByBudget(t, ask(c, "x"))
	// 4 input bytes (the negative count is no observation), and the larger of
	// the 100 reserved and the 9+94 observed output tokens.
	if got := b.Spent().Tokens; got != 4+103 {
		t.Fatalf("charged %d, want 107", got)
	}
}
