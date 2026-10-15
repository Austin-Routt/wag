package llm

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestLive calls real providers, and only when asked: WAG_LIVE_TESTS names
// them (gemini, anthropic, openai, deepseek, lmstudio, …, comma-separated),
// WAG_LIVE_MODEL_<PROVIDER> (e.g. WAG_LIVE_MODEL_GEMINI=gemini-2.5-flash)
// the model, and the provider's usual variable the key. Each sends one
// short request, which may cost credits or use a free tier's quota:
//
//	WAG_LIVE_TESTS=gemini WAG_LIVE_MODEL_GEMINI=gemini-2.5-flash go test ./internal/llm -run TestLive -v
func TestLive(t *testing.T) {
	list := strings.TrimSpace(os.Getenv("WAG_LIVE_TESTS"))
	if list == "" {
		t.Skip("set WAG_LIVE_TESTS to call real providers")
	}
	for _, prov := range strings.Split(list, ",") {
		prov = strings.TrimSpace(prov)
		t.Run(prov, func(t *testing.T) {
			model := os.Getenv("WAG_LIVE_MODEL_" + strings.ToUpper(strings.ReplaceAll(prov, "-", "_")))
			if model == "" {
				t.Skipf("set WAG_LIVE_MODEL_%s", strings.ToUpper(prov))
			}
			c, err := NewClient(Profile{Name: "live", Provider: prov, BaseURL: os.Getenv("WAG_LIVE_BASE_URL_" + strings.ToUpper(prov)),
				Timeout: Duration(2 * time.Minute), Params: Params{Model: model, MaxOutputTokens: 2000, Output: "json"}})
			if err != nil {
				t.Fatal(err)
			}
			var tries []Attempt
			resp, err := c.Complete(context.Background(), c.Request("Answer with one JSON object.",
				[]Message{{Role: "user", Content: `Reply with {"ready": true}`}}), func(a Attempt) { tries = append(tries, a) })
			if err != nil {
				t.Fatalf("%v (%d tries)", err, len(tries))
			}
			if !strings.Contains(resp.Text, "ready") {
				t.Fatalf("answer %q", resp.Text)
			}
			t.Logf("model %s, finish %s, usage in=%v out=%v, latency %v, request %s", resp.Model, resp.FinishReason,
				deref(resp.Usage.InputTokens), deref(resp.Usage.OutputTokens), tries[len(tries)-1].Latency, resp.RequestID)
		})
	}
}

func deref(p *int) any {
	if p == nil {
		return "unknown"
	}
	return *p
}
