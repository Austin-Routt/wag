package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeServer records the requests it gets and answers with handle.
type fakeServer struct {
	mu     sync.Mutex
	reqs   []*http.Request
	bodies []map[string]any
	*httptest.Server
}

func newFake(t *testing.T, handle func(w http.ResponseWriter, n int)) *fakeServer {
	f := &fakeServer{}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		f.mu.Lock()
		f.reqs = append(f.reqs, r)
		f.bodies = append(f.bodies, m)
		n := len(f.reqs)
		f.mu.Unlock()
		handle(w, n)
	}))
	t.Cleanup(f.Close)
	return f
}

func fp(f float64) *float64 { return &f }

func conversation() []Message {
	return []Message{{Role: "user", Content: "the board"}, {Role: "assistant", Content: "my move"}, {Role: "user", Content: "rejected: try again"}}
}

// The Gemini adapter sends the system prompt, the conversation with the
// model's turns as "model", the parameters and the key (in a header, not the
// URL), and reads the answer without its thought parts, with the usage.
func TestGeminiRequestAndResponse(t *testing.T) {
	f := newFake(t, func(w http.ResponseWriter, n int) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"candidates":[{"content":{"role":"model","parts":[{"text":"hidden","thought":true},{"text":"{\"action\":"},{"text":"{\"type\":\"abstain\"}}"}]},"finishReason":"STOP"}],
			"usageMetadata":{"promptTokenCount":120,"candidatesTokenCount":15,"thoughtsTokenCount":40,"totalTokenCount":175},
			"modelVersion":"gemini-test-001","responseId":"resp-1"}`)
	})
	t.Setenv("GEMINI_API_KEY", "test-key-123")
	t.Setenv("GOOGLE_API_KEY", "")
	c, err := NewClient(Profile{Name: "g", Provider: "gemini", BaseURL: f.URL, Params: Params{Model: "gemini-test", Temperature: fp(0.5), Output: "json", Effort: "low"}})
	if err != nil {
		t.Fatal(err)
	}
	var attempts []Attempt
	resp, err := c.Complete(context.Background(), c.Request("You play WAG.", conversation()), func(a Attempt) { attempts = append(attempts, a) })
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text != `{"action":{"type":"abstain"}}` || resp.FinishReason != "STOP" || resp.Model != "gemini-test-001" || resp.ResponseID != "resp-1" {
		t.Fatalf("response: %+v", resp)
	}
	if *resp.Usage.InputTokens != 120 || *resp.Usage.OutputTokens != 15 || *resp.Usage.ReasoningTokens != 40 || *resp.Usage.TotalTokens != 175 || resp.Usage.CachedTokens != nil {
		t.Fatalf("usage: %+v", resp.Usage)
	}
	r := f.reqs[0]
	if !strings.HasSuffix(r.URL.Path, "/models/gemini-test:generateContent") || strings.Contains(r.URL.String(), "test-key") {
		t.Fatalf("URL %s", r.URL)
	}
	if r.Header.Get("X-Goog-Api-Key") != "test-key-123" {
		t.Fatalf("key header: %v", r.Header)
	}
	body := f.bodies[0]
	contents := body["contents"].([]any)
	roles := []string{}
	for _, c := range contents {
		roles = append(roles, c.(map[string]any)["role"].(string))
	}
	if strings.Join(roles, ",") != "user,model,user" {
		t.Fatalf("roles %v", roles)
	}
	sys, _ := json.Marshal(body["systemInstruction"])
	gc, _ := json.Marshal(body["generationConfig"])
	if !strings.Contains(string(sys), "You play WAG.") || !strings.Contains(string(gc), `"temperature":0.5`) ||
		!strings.Contains(string(gc), `"responseMimeType":"application/json"`) || !strings.Contains(string(gc), `"thinkingLevel":"LOW"`) ||
		strings.Contains(string(gc), "includeThoughts") {
		t.Fatalf("system %s, config %s", sys, gc)
	}
	if len(attempts) != 1 || attempts[0].Response == nil || attempts[0].Request.System != "You play WAG." || len(attempts[0].Request.Messages) != 3 {
		t.Fatalf("attempts: %+v", attempts)
	}
	if b, _ := json.Marshal(attempts[0]); strings.Contains(string(b), "test-key") {
		t.Fatal("the key is in the trace")
	}
}

// A rate limit is retried after the delay the API asks for; the trace has
// both attempts.
func TestGeminiRateLimitRetry(t *testing.T) {
	f := newFake(t, func(w http.ResponseWriter, n int) {
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			w.WriteHeader(429)
			io.WriteString(w, `{"error":{"code":429,"message":"Quota exceeded","status":"RESOURCE_EXHAUSTED",
				"details":[{"@type":"type.googleapis.com/google.rpc.RetryInfo","retryDelay":"37s"}]}}`)
			return
		}
		io.WriteString(w, `{"candidates":[{"content":{"parts":[{"text":"ok"}]},"finishReason":"STOP"}]}`)
	})
	t.Setenv("GEMINI_API_KEY", "k")
	c, err := NewClient(Profile{Name: "g", Provider: "gemini", BaseURL: f.URL, Params: Params{Model: "m"}})
	if err != nil {
		t.Fatal(err)
	}
	var slept []time.Duration
	c.Sleep = func(_ context.Context, d time.Duration) error { slept = append(slept, d); return nil }
	var attempts []Attempt
	resp, err := c.Complete(context.Background(), c.Request("", []Message{{Role: "user", Content: "hi"}}), func(a Attempt) { attempts = append(attempts, a) })
	if err != nil || resp.Text != "ok" {
		t.Fatalf("%v %+v", err, resp)
	}
	if len(attempts) != 2 || attempts[0].Error == nil || attempts[0].Error.Kind != KindRateLimit || attempts[0].Error.RetryAfter != 37*time.Second || attempts[1].N != 2 {
		t.Fatalf("attempts: %+v", attempts)
	}
	if resp.Usage.InputTokens != nil {
		t.Fatal("unreported usage must stay unknown")
	}
	if len(slept) < 2 || slept[1] != 37*time.Second {
		t.Fatalf("waits %v", slept)
	}
}

// A blocked answer, and a missing key, are errors of their kind.
func TestGeminiBlockedAndNoKey(t *testing.T) {
	f := newFake(t, func(w http.ResponseWriter, n int) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"candidates":[{"finishReason":"SAFETY"}]}`)
	})
	t.Setenv("GEMINI_API_KEY", "k")
	c, _ := NewClient(Profile{Name: "g", Provider: "gemini", BaseURL: f.URL, Params: Params{Model: "m"}})
	_, err := c.Complete(context.Background(), c.Request("", []Message{{Role: "user", Content: "hi"}}), nil)
	if e, ok := err.(*Error); !ok || e.Kind != KindBlocked {
		t.Fatalf("%v", err)
	}
	if len(f.reqs) != 1 {
		t.Fatal("a blocked answer is not retried")
	}
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")
	if _, err := NewClient(Profile{Name: "g", Provider: "gemini", Params: Params{Model: "m"}}); err == nil || !strings.Contains(err.Error(), "GEMINI_API_KEY") {
		t.Fatalf("no key: %v", err)
	}
}

// The OpenAI-compatible adapter sends the system prompt as a message, only
// the parameters set, and reads the answer, its usage, and the request ID;
// a reasoning block before the answer is dropped.
func TestOpenAICompatible(t *testing.T) {
	f := newFake(t, func(w http.ResponseWriter, n int) {
		w.Header().Set("X-Request-Id", "req-9")
		io.WriteString(w, `{"id":"chatcmpl-1","model":"qwen3-8b","choices":[{"message":{"content":"<think>long</think>\n{\"action\":{\"type\":\"abstain\"}}","reasoning_content":"x"},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":50,"completion_tokens":9,"total_tokens":59,"completion_tokens_details":{"reasoning_tokens":4}}}`)
	})
	c, err := NewClient(Profile{Name: "local", Provider: "lmstudio", BaseURL: f.URL + "/v1", Params: Params{Model: "qwen3-8b", MaxOutputTokens: 900, Output: "json",
		Extra: map[string]any{"chat_template_kwargs": map[string]any{"enable_thinking": false}}}})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Complete(context.Background(), c.Request("Sys", conversation()), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text != `{"action":{"type":"abstain"}}` || resp.RequestID != "req-9" || resp.Model != "qwen3-8b" || *resp.Usage.ReasoningTokens != 4 || resp.Meta["reasoning_block_removed"] != true {
		t.Fatalf("%+v", resp)
	}
	if f.reqs[0].URL.Path != "/v1/chat/completions" || f.reqs[0].Header.Get("Authorization") != "" {
		t.Fatalf("%s %v", f.reqs[0].URL.Path, f.reqs[0].Header)
	}
	b := f.bodies[0]
	msgs := b["messages"].([]any)
	if len(msgs) != 4 || msgs[0].(map[string]any)["role"] != "system" || msgs[2].(map[string]any)["role"] != "assistant" {
		t.Fatalf("messages %v", msgs)
	}
	if b["max_tokens"] != 900.0 || b["temperature"] != nil || b["response_format"].(map[string]any)["type"] != "json_object" || b["chat_template_kwargs"] == nil {
		t.Fatalf("body %v", b)
	}
}

// OpenAI itself needs a key, and uses max_completion_tokens; an error's
// Retry-After is kept.
func TestOpenAIKeyAndErrors(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	if _, err := NewClient(Profile{Name: "o", Provider: "openai", Params: Params{Model: "gpt-x"}}); err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("%v", err)
	}
	f := newFake(t, func(w http.ResponseWriter, n int) {
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(429)
		io.WriteString(w, `{"error":{"message":"slow down","type":"rate_limit_exceeded"}}`)
	})
	t.Setenv("OPENAI_API_KEY", "sk-test")
	c, err := NewClient(Profile{Name: "o", Provider: "openai", BaseURL: f.URL, Retries: -1, Params: Params{Model: "gpt-x", MaxOutputTokens: 100}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Complete(context.Background(), c.Request("", []Message{{Role: "user", Content: "hi"}}), nil)
	e, ok := err.(*Error)
	if !ok || e.Kind != KindRateLimit || e.RetryAfter != 7*time.Second || !strings.Contains(e.Message, "slow down") {
		t.Fatalf("%v", err)
	}
	if f.reqs[0].Header.Get("Authorization") != "Bearer sk-test" || f.bodies[0]["max_completion_tokens"] != 100.0 || len(f.reqs) != 1 {
		t.Fatalf("%v %v", f.reqs[0].Header, f.bodies[0])
	}
}

// The Anthropic adapter, through the SDK, against a fake Messages API.
func TestAnthropic(t *testing.T) {
	f := newFake(t, func(w http.ResponseWriter, n int) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Request-Id", "req_abc")
		io.WriteString(w, `{"id":"msg_1","type":"message","role":"assistant","model":"claude-test","content":[{"type":"text","text":"ok"}],
			"stop_reason":"end_turn","usage":{"input_tokens":30,"output_tokens":2,"cache_read_input_tokens":0}}`)
	})
	t.Setenv("WAG_TEST_ANTHROPIC_KEY", "sk-ant-test")
	c, err := NewClient(Profile{Name: "c", Provider: "anthropic", BaseURL: f.URL, APIKeyEnv: "WAG_TEST_ANTHROPIC_KEY", Params: Params{Model: "claude-test", Effort: "low"}})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Complete(context.Background(), c.Request("Sys", conversation()), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text != "ok" || resp.RequestID != "req_abc" || resp.FinishReason != "end_turn" || *resp.Usage.InputTokens != 30 {
		t.Fatalf("%+v", resp)
	}
	b := f.bodies[0]
	if b["model"] != "claude-test" || len(b["messages"].([]any)) != 3 || f.reqs[0].Header.Get("X-Api-Key") != "sk-ant-test" {
		t.Fatalf("%v %v", b, f.reqs[0].Header)
	}
}

// The command provider passes the request on standard input.
func TestCommand(t *testing.T) {
	c, err := NewClient(Profile{Name: "x", Provider: "command", Command: []string{"sh", "-c", `grep -q '"content":"hi"' && echo '{"text":"fine","model":"script"}'`}})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Complete(context.Background(), c.Request("", []Message{{Role: "user", Content: "hi"}}), nil)
	if err != nil || resp.Text != "fine" || resp.Model != "script" {
		t.Fatalf("%v %+v", err, resp)
	}
	c, _ = NewClient(Profile{Name: "x", Provider: "command", Retries: -1, Command: []string{"sh", "-c", "echo broken >&2; exit 3"}})
	if _, err := c.Complete(context.Background(), c.Request("", nil), nil); err == nil || !strings.Contains(err.Error(), "broken") {
		t.Fatalf("%v", err)
	}
}

// Not every error is retried; min_interval spaces calls.
func TestClientPolicy(t *testing.T) {
	calls := 0
	m := &Mock{Fn: func(Request) (*Response, error) {
		calls++
		return nil, &Error{Provider: "mock", Kind: KindAuth, Message: "bad key"}
	}}
	c := WithProvider(Profile{Name: "m", Provider: "mock"}, m)
	if _, err := c.Complete(context.Background(), c.Request("", nil), nil); err == nil || calls != 1 {
		t.Fatalf("an auth error is not retried: %v, %d calls", err, calls)
	}
	c = WithProvider(Profile{Name: "m", Provider: "mock", MinInterval: Duration(time.Hour)}, &Mock{})
	var waits []time.Duration
	c.Sleep = func(_ context.Context, d time.Duration) error { waits = append(waits, d); return nil }
	for i := 0; i < 2; i++ {
		if _, err := c.Complete(context.Background(), c.Request("", nil), nil); err != nil {
			t.Fatal(err)
		}
	}
	if waits[len(waits)-1] < 59*time.Minute {
		t.Fatalf("the second call should wait about an hour: %v", waits)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := WithProvider(Profile{Name: "m", Provider: "mock"}, &Mock{}).Complete(ctx, Request{}, nil); err == nil || err.(*Error).Kind != KindCanceled {
		t.Fatalf("canceled: %v", err)
	}
}

func TestProfiles(t *testing.T) {
	ps, err := ParseProfiles([]byte(`{"profiles":[{"name":"flash","provider":"gemini","model":"gemini-2.5-flash","temperature":0.7,"min_interval":"6s","output":"json"},
		{"name":"local","provider":"lmstudio","model":"qwen3-8b","timeout":120}]}`))
	if err != nil || len(ps) != 2 || *ps[0].Temperature != 0.7 || time.Duration(ps[0].MinInterval) != 6*time.Second || time.Duration(ps[1].Timeout) != 2*time.Minute {
		t.Fatalf("%v %+v", err, ps)
	}
	for _, bad := range []string{
		`[{"name":"a","provider":"gemini","model":"m","api_key":"AIza…"}]`,
		`[{"name":"a","provider":"gemini","model":"m","temprature":1}]`,
		`[{"name":"a","provider":"nope","model":"m"}]`,
		`[{"name":"a","provider":"gemini"}]`,
		`[{"name":"a","provider":"openai-compatible","model":"m"}]`,
		`[{"name":"a","provider":"gemini","model":"m"},{"name":"a","provider":"gemini","model":"n"}]`,
	} {
		if _, err := ParseProfiles([]byte(bad)); err == nil {
			t.Errorf("accepted %s", bad)
		}
	}
	p, err := ParseSpec("fast=gemini:gemini-2.5-flash,temperature=0.2,min_interval=4s,seed=7,output=json")
	if err != nil || p.Name != "fast" || p.Model != "gemini-2.5-flash" || *p.Seed != 7 || time.Duration(p.MinInterval) != 4*time.Second {
		t.Fatalf("%v %+v", err, p)
	}
	if p, err := ParseSpec("lmstudio:qwen3-8b,base_url=http://127.0.0.1:1234/v1"); err != nil || p.Name != "lmstudio:qwen3-8b" || p.BaseURL != "http://127.0.0.1:1234/v1" {
		t.Fatalf("%v %+v", err, p)
	}
	if _, err := ParseSpec("gemini:m,api_key=AIza"); err == nil {
		t.Fatal("a key on the command line")
	}
	t.Setenv("GEMINI_API_KEY", "a")
	t.Setenv("GOOGLE_API_KEY", "b")
	if k, v := (&Profile{Provider: "gemini"}).key(); k != "a" || v != "GEMINI_API_KEY" {
		t.Fatalf("GEMINI_API_KEY comes first: %s %s", k, v)
	}
}

func TestStripThink(t *testing.T) {
	for in, want := range map[string]string{"plain": "plain", "<think>a</think>\n\nb": "b", "  <think>a": "", "x <think>y</think>": "x <think>y</think>"} {
		if got, _ := stripThink(in); got != want {
			t.Errorf("%q: %q", in, got)
		}
	}
}

// Gemini answers a wrong key with HTTP 400: it is an auth error.
func TestGeminiWrongKey(t *testing.T) {
	f := newFake(t, func(w http.ResponseWriter, n int) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		io.WriteString(w, `{"error":{"code":400,"message":"API key not valid. Please pass a valid API key.","status":"INVALID_ARGUMENT"}}`)
	})
	c, _ := NewClient(Profile{Name: "g", Provider: "gemini", BaseURL: f.URL, APIKey: "wrong", Params: Params{Model: "m"}})
	_, err := c.Complete(context.Background(), c.Request("", []Message{{Role: "user", Content: "hi"}}), nil)
	if e, ok := err.(*Error); !ok || e.Kind != KindAuth || len(f.reqs) != 1 {
		t.Fatalf("%v, %d requests", err, len(f.reqs))
	}
	if f.reqs[0].Header.Get("X-Goog-Api-Key") != "wrong" {
		t.Fatal("the dashboard's key is not used")
	}
}
