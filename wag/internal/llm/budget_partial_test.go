package llm

import (
	"context"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"
)

// An unusable answer (empty, refused, blocked) can still be billed: each
// adapter returns its usage with the error, so the budget counts it and the
// trace keeps it, while the client still fails the call.
func TestBudgetCountsUnusableAnswers(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "k")
	t.Setenv("WAG_TEST_ANTHROPIC_KEY", "sk-ant-test")
	for _, tc := range []struct {
		name, reply string
		profile     func(url string) Profile
		kind        string
		tokens      int
	}{
		{"openai-compatible, reasoning took the whole output",
			`{"id":"r","model":"m","choices":[{"message":{"content":""},"finish_reason":"length"}],"usage":{"prompt_tokens":1000,"completion_tokens":4000,"total_tokens":5000,"completion_tokens_details":{"reasoning_tokens":4000}}}`,
			func(url string) Profile {
				return Profile{Name: "o", Provider: "lmstudio", BaseURL: url + "/v1", Params: Params{Model: "m"}}
			}, KindEmpty, 5000},
		{"openai-compatible, refused",
			`{"id":"r","model":"m","choices":[{"message":{"content":null,"refusal":"no"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1000,"completion_tokens":3,"total_tokens":1003}}`,
			func(url string) Profile {
				return Profile{Name: "o", Provider: "lmstudio", BaseURL: url + "/v1", Params: Params{Model: "m"}}
			}, KindBlocked, 1003},
		{"gemini, thoughts took the whole output",
			`{"candidates":[{"finishReason":"MAX_TOKENS","content":{"role":"model","parts":[{"text":"hmm","thought":true}]}}],"usageMetadata":{"promptTokenCount":1000,"candidatesTokenCount":0,"thoughtsTokenCount":4000,"totalTokenCount":5000}}`,
			func(url string) Profile {
				return Profile{Name: "g", Provider: "gemini", BaseURL: url, Params: Params{Model: "m"}}
			}, KindEmpty, 5000},
		{"gemini, the prompt blocked",
			`{"promptFeedback":{"blockReason":"SAFETY"},"usageMetadata":{"promptTokenCount":1000,"totalTokenCount":1000}}`,
			func(url string) Profile {
				return Profile{Name: "g", Provider: "gemini", BaseURL: url, Params: Params{Model: "m"}}
			}, KindBlocked, 1000},
		{"anthropic, the output limit reached with no text",
			`{"id":"msg_1","type":"message","role":"assistant","model":"claude-test","content":[],"stop_reason":"max_tokens","usage":{"input_tokens":1000,"output_tokens":4000}}`,
			func(url string) Profile {
				return Profile{Name: "c", Provider: "anthropic", BaseURL: url, APIKeyEnv: "WAG_TEST_ANTHROPIC_KEY", Params: Params{Model: "claude-test"}}
			}, KindEmpty, 5000},
	} {
		f := newFake(t, func(w http.ResponseWriter, n int) {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, tc.reply)
		})
		c, err := NewClient(tc.profile(f.URL))
		if err != nil {
			t.Fatal(err)
		}
		c.Sleep = NoWait
		b := &Budget{}
		c.Budget = b
		var trace []Attempt
		resp, err := c.Complete(context.Background(), c.Request("sys", []Message{{Role: "user", Content: "x"}}), func(a Attempt) { trace = append(trace, a) })
		e, _ := err.(*Error)
		if resp != nil || e == nil || e.Kind != tc.kind {
			t.Fatalf("%s: the call must fail as %s: %v, %v", tc.name, tc.kind, resp, err)
		}
		if s := b.Spent(); s.Calls != 1 || s.Tokens != tc.tokens || s.Unreported != 0 {
			t.Errorf("%s: spent %+v; want 1 call and %d reported tokens", tc.name, s, tc.tokens)
		}
		if len(trace) != 1 || trace[0].Error == nil || trace[0].Response == nil || trace[0].Response.Usage.InputTokens == nil {
			t.Errorf("%s: the trace does not keep the usage: %+v", tc.name, trace)
		}
	}
}

// Under a token limit, an unusable answer that used more than its estimate
// is an overrun: it is charged what was reported and stops the invocation.
func TestBudgetUnusableAnswerOverrun(t *testing.T) {
	f := newFake(t, func(w http.ResponseWriter, n int) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"r","model":"m","choices":[{"message":{"content":""},"finish_reason":"length"}],"usage":{"prompt_tokens":1000,"completion_tokens":4000,"total_tokens":5000}}`)
	})
	c, err := NewClient(Profile{Name: "o", Provider: "lmstudio", BaseURL: f.URL + "/v1", Retries: -1, Params: Params{Model: "m", MaxOutputTokens: 100}})
	if err != nil {
		t.Fatal(err)
	}
	b := &Budget{MaxTokens: 100000}
	c.Budget = b
	if err := ask(c, "x"); err == nil {
		t.Fatal("an empty answer succeeded")
	}
	if s := b.Spent(); s.Tokens != 5000 || b.Halted() == "" {
		t.Fatalf("spent %+v, halted %q; want 5000 tokens and a halt", s, b.Halted())
	}
	refusedByBudget(t, ask(c, "x"))
	if len(f.reqs) != 1 {
		t.Fatal("a call was sent after the overrun")
	}
}

// Without an input count, the total less a complete output is the input:
// priced so, nothing halts. Without the output either, the total is priced
// at the dearer rate, an upper bound.
func TestBudgetInputFromTotal(t *testing.T) {
	prices := map[string]Price{"*": {Input: 1, Output: 10}}
	long := strings.Repeat("x", 997) // with "sys", 1000 input bytes
	p := Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 100}}
	b := &Budget{MaxTokens: 100000, MaxCost: 1, Prices: prices}
	c, _ := budgetClient(p, b, Usage{OutputTokens: intp(50), TotalTokens: intp(600)}, 0)
	if err := ask(c, long); err != nil {
		t.Fatalf("a coherent report without input halted the run: %v", err)
	}
	if s := b.Spent(); s.Tokens != 600 || math.Abs(s.Cost-(550*1+50*10)/1e6) > 1e-12 || b.Halted() != "" {
		t.Fatalf("spent %+v, halted %q; want 600 tokens and cost 1050e-6", s, b.Halted())
	}
	for _, tc := range []struct {
		u    Usage
		cost float64
	}{
		{Usage{OutputTokens: intp(50), TotalTokens: intp(600)}, (550*1 + 50*10) / 1e6},
		{Usage{TotalTokens: intp(600)}, 600 * 10 / 1e6},
	} {
		nb := &Budget{Prices: prices}
		c, _ := budgetClient(Profile{Name: "p", Provider: "mock"}, nb, tc.u, 0)
		if err := ask(c, "x"); err != nil {
			t.Fatal(err)
		}
		if s := nb.Spent(); s.Tokens != 600 || math.Abs(s.Cost-tc.cost) > 1e-12 {
			t.Errorf("no limit, %+v: spent %+v; want 600 tokens and cost %g", tc.u, s, tc.cost)
		}
	}
	// The dearer rate is the dearer whichever it is: here the input's.
	ib := &Budget{Prices: map[string]Price{"*": {Input: 10, Output: 1}}}
	c, _ = budgetClient(Profile{Name: "p", Provider: "mock"}, ib, Usage{TotalTokens: intp(600)}, 0)
	if err := ask(c, "x"); err != nil || math.Abs(ib.Spent().Cost-600*10/1e6) > 1e-12 {
		t.Fatalf("%v, spent %+v; want cost 6000e-6", err, ib.Spent())
	}
}

// A cost limit lets through the call that reaches it exactly, whatever the
// float sum says (0.1+0.1+0.1 is 0.30000000000000004), and refuses the next.
func TestBudgetCostLimitReachedExactly(t *testing.T) {
	p := Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 100}}
	b := &Budget{MaxCost: 0.3, Prices: map[string]Price{"*": {Input: 0, Output: 1000}}} // 0.1 a call
	c, calls := budgetClient(p, b, Usage{InputTokens: intp(4), OutputTokens: intp(100)}, 0)
	for i := 0; i < 3; i++ {
		if err := ask(c, "x"); err != nil {
			t.Fatalf("call %d: %v", i+1, err)
		}
	}
	refusedByBudget(t, ask(c, "x"))
	if *calls != 3 {
		t.Fatalf("%d calls reached the provider, want 3", *calls)
	}
}

// An accounting violation stops the invocation without being a refused
// call; the next attempt is refused.
func TestBudgetHaltIsNotARefusal(t *testing.T) {
	b := &Budget{MaxTokens: 100000}
	c, _ := budgetClient(Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 10}}, b, Usage{InputTokens: intp(50000)}, 0)
	refusedByBudget(t, ask(c, "x"))
	if b.Halted() == "" || !b.Refused() || b.Refusals() != 0 {
		t.Fatalf("halted %q, refused %v, refusals %d", b.Halted(), b.Refused(), b.Refusals())
	}
	refusedByBudget(t, ask(c, "x"))
	if b.Refusals() != 1 {
		t.Fatalf("refusals %d, want 1", b.Refusals())
	}
}

// Usage that a given price table has no rate for is counted as unpriced,
// with or without an estimate, and adds no cost.
func TestBudgetUnpriced(t *testing.T) {
	prices := map[string]Price{"a": {Input: 1, Output: 1}}
	for _, b := range []*Budget{{Prices: prices}, {MaxTokens: 1000, Prices: prices}} {
		ca, _ := budgetClient(Profile{Name: "a", Provider: "mock", Params: Params{MaxOutputTokens: 10}}, b, Usage{InputTokens: intp(4), OutputTokens: intp(5)}, 0)
		cb, _ := budgetClient(Profile{Name: "b", Provider: "mock", Params: Params{MaxOutputTokens: 10}}, b, Usage{InputTokens: intp(4), OutputTokens: intp(5)}, 0)
		if err := ask(ca, "x"); err != nil {
			t.Fatal(err)
		}
		if err := ask(cb, "x"); err != nil {
			t.Fatal(err)
		}
		if s := b.Spent(); s.Unpriced != 1 || math.Abs(s.Cost-9e-6) > 1e-12 {
			t.Errorf("limit %d: spent %+v; want 1 unpriced and cost 9e-6", b.MaxTokens, s)
		}
	}
	// A failed attempt under a token limit keeps its unpriced reservation.
	fb := &Budget{MaxTokens: 1000, Prices: prices}
	cf, _ := budgetClient(Profile{Name: "b", Provider: "mock", Retries: -1, Params: Params{MaxOutputTokens: 10}}, fb, Usage{}, 1)
	if err := ask(cf, "x"); err == nil || fb.Spent().Unpriced != 1 || fb.Spent().Cost != 0 {
		t.Fatalf("%v, spent %+v; want 1 unpriced and no cost", err, fb.Spent())
	}
	nb := &Budget{} // no price table: nothing is unpriced
	c, _ := budgetClient(Profile{Name: "b", Provider: "mock"}, nb, Usage{InputTokens: intp(4)}, 0)
	if err := ask(c, "x"); err != nil || nb.Spent().Unpriced != 0 {
		t.Fatalf("%v, %+v", err, nb.Spent())
	}
}

// A cost that overflows is capped, not dropped.
func TestBudgetCountCapsAnInfiniteCost(t *testing.T) {
	b := &Budget{Prices: map[string]Price{"*": {Input: math.MaxFloat64, Output: math.MaxFloat64}}}
	c, _ := budgetClient(Profile{Name: "p", Provider: "mock"}, b, Usage{InputTokens: intp(2000000), OutputTokens: intp(2000000)}, 0)
	if err := ask(c, "x"); err != nil {
		t.Fatal(err)
	}
	if got := b.Spent().Cost; got != math.MaxFloat64 {
		t.Fatalf("cost %g, want the cap", got)
	}
}
