package llm

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestBudgetHardeningRejectsExtraForEstimates(t *testing.T) {
	for _, provider := range []string{"openai", "openai-compatible", "gemini"} {
		for _, cost := range []bool{false, true} {
			p := Profile{Name: "p", Provider: provider, Params: Params{MaxOutputTokens: 1, ThinkingBudget: intp(0), Extra: map[string]any{"max_tokens": 100000, "max_completion_tokens": 100000, "messages": []any{}}}}
			b := &Budget{MaxTokens: 10000}
			if cost {
				b = &Budget{MaxCost: 1, Prices: map[string]Price{"*": {Input: 1, Output: 1}}}
			}
			c, calls := budgetClient(p, b, Usage{}, 0)
			refusedByBudget(t, ask(c, "x"))
			if *calls != 0 {
				t.Fatalf("%s Extra reached provider under estimated cap", provider)
			}
		}
	}
	for _, b := range []*Budget{nil, {MaxCalls: 1}} {
		p := Profile{Name: "p", Provider: "mock", Params: Params{Extra: map[string]any{"custom_option": true}}}
		c, calls := budgetClient(p, b, Usage{}, 0)
		if err := ask(c, "x"); err != nil || *calls != 1 {
			t.Fatalf("call-only/unbudgeted compatibility changed: %v, calls=%d", err, *calls)
		}
	}
}

func TestBudgetHardeningIncludesSchemaInEstimate(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","description":"` + strings.Repeat("x", 300) + `"}`)
	p := Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 1, Output: "schema", Schema: schema}}
	c, calls := budgetClient(p, &Budget{MaxTokens: 50}, Usage{}, 0)
	refusedByBudget(t, ask(c, "x"))
	if *calls != 0 {
		t.Fatal("schema bytes were omitted from the estimate")
	}
}

func TestBudgetHardeningInvalidNativeConfiguration(t *testing.T) {
	bad := []*Budget{
		{MaxCalls: -1}, {MaxTokens: -1}, {MaxCost: -1}, {MaxCost: math.NaN()}, {MaxCost: math.Inf(1)},
		{MaxCost: 1, Prices: map[string]Price{"*": {Input: -1, Output: 1}}},
		{MaxCost: 1, Prices: map[string]Price{"*": {Input: 1, Output: math.NaN()}}},
		{MaxCalls: 1, Prices: map[string]Price{"*": {Input: math.Inf(1), Output: 1}}},
	}
	for i, b := range bad {
		c, calls := budgetClient(Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 10}}, b, Usage{}, 0)
		if err := ask(c, "x"); err == nil || *calls != 0 {
			t.Errorf("invalid configuration %d reached provider: err=%v calls=%d", i, err, *calls)
		}
	}
}

func TestBudgetHardeningUsageNeverCreatesCredit(t *testing.T) {
	for _, usage := range []Usage{
		{InputTokens: intp(-10), OutputTokens: intp(1)},
		{InputTokens: intp(20), TotalTokens: intp(5)},
		{InputTokens: intp(4), OutputTokens: intp(10), TotalTokens: intp(13)},
		{InputTokens: intp(4), OutputTokens: intp(1), CachedTokens: intp(-1)},
		{InputTokens: intp(500), OutputTokens: intp(-1)},
	} {
		b := &Budget{MaxTokens: 10000, MaxCost: 1, Prices: map[string]Price{"*": {Input: 1, Output: 1}}}
		c, calls := budgetClient(Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 100}}, b, usage, 0)
		refusedByBudget(t, ask(c, "x"))
		spent := b.Spent()
		if spent.Calls != 1 || spent.Tokens < 104 || spent.Cost+1e-12 < 0.000104 {
			t.Fatalf("invalid usage created credit below reservation: %+v", spent)
		}
		if usage.InputTokens != nil && *usage.InputTokens > 104 && spent.Tokens < *usage.InputTokens {
			t.Fatalf("valid positive observation was lost: %+v", spent)
		}
		refusedByBudget(t, ask(c, "x"))
		if *calls != 1 {
			t.Fatal("accounting violation did not stop subsequent attempts")
		}
	}
}

func TestBudgetHardeningPartialGeminiReasoning(t *testing.T) {
	p := Profile{Name: "g", Provider: "gemini", Params: Params{MaxOutputTokens: 50, ThinkingBudget: intp(100)}}
	b := &Budget{MaxTokens: 1000}
	c, _ := budgetClient(p, b, Usage{InputTokens: intp(4), OutputTokens: intp(10)}, 0)
	if err := ask(c, "x"); err != nil {
		t.Fatal(err)
	}
	if got := b.Spent().Tokens; got != 114 {
		t.Fatalf("unreported reasoning vanished: charged %d, want 114", got)
	}
	b = &Budget{MaxTokens: 1000}
	c, _ = budgetClient(p, b, Usage{InputTokens: intp(4), OutputTokens: intp(10), ReasoningTokens: intp(5), TotalTokens: intp(19)}, 0)
	if err := ask(c, "x"); err != nil || b.Spent().Tokens != 19 {
		t.Fatalf("complete consistent Gemini usage not retained: %v %+v", err, b.Spent())
	}
}

func TestBudgetHardeningTotalBelowKnownReasoning(t *testing.T) {
	b := &Budget{MaxTokens: 1000}
	c, calls := budgetClient(Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 100}}, b,
		Usage{InputTokens: intp(1), ReasoningTokens: intp(10), TotalTokens: intp(2)}, 0)
	refusedByBudget(t, ask(c, "x"))
	if spent := b.Spent(); spent.Tokens < 104 || spent.Calls != 1 || *calls != 1 {
		t.Fatalf("inconsistent total erased known inclusive reasoning: %+v, calls=%d", spent, *calls)
	}
}

func TestBudgetHardeningObservedOverrunIsPreserved(t *testing.T) {
	b := &Budget{MaxTokens: 1000, MaxCost: 1, Prices: map[string]Price{"*": {Input: 1, Output: 1}}}
	c, calls := budgetClient(Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: 100}}, b, Usage{InputTokens: intp(200), OutputTokens: intp(50), TotalTokens: intp(250)}, 0)
	var trace Attempt
	_, err := c.Complete(context.Background(), c.Request("sys", []Message{{Role: "user", Content: "x"}}), func(a Attempt) { trace = a })
	refusedByBudget(t, err)
	if got := b.Spent(); got.Tokens != 250 || math.Abs(got.Cost-0.00025) > 1e-12 || got.Calls != 1 {
		t.Fatalf("observed spend overwritten by estimate: %+v", got)
	}
	if trace.Response == nil || trace.Response.Usage.TotalTokens == nil || *trace.Response.Usage.TotalTokens != 250 || trace.Error == nil {
		t.Fatalf("accounting failure lost observed response/usage: %+v", trace)
	}
	refusedByBudget(t, ask(c, "x"))
	if *calls != 1 {
		t.Fatal("another attempt was sent after estimated reservation was exceeded")
	}
}

func TestBudgetHardeningEstimateOverflowRefused(t *testing.T) {
	p := Profile{Name: "p", Provider: "mock", Params: Params{MaxOutputTokens: int(^uint(0) >> 1)}}
	c, calls := budgetClient(p, &Budget{MaxTokens: 1000}, Usage{}, 0)
	refusedByBudget(t, ask(c, "x"))
	if *calls != 0 {
		t.Fatal("overflowed reservation reached provider")
	}
}
