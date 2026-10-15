package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// OpenAI calls an OpenAI-compatible Chat Completions endpoint
// (POST {base}/chat/completions): OpenAI, DeepSeek, and local engines
// (LM Studio, llama.cpp, SGLang, vLLM, Ollama). Compatible servers differ in
// what they accept; WAG sends only the parameters a profile sets, and
// records what the server answered.
type OpenAI struct {
	name      string
	base      string
	key       string
	maxTokens string // the request field for the output limit
	http      *http.Client
}

func newOpenAI(p Profile) (*OpenAI, error) {
	ep := endpoints[p.Provider]
	base := p.BaseURL
	if base == "" {
		base = ep.BaseURL
	}
	key, _ := p.key()
	if ep.NeedsKey && key == "" {
		return nil, &Error{Provider: p.Provider, Kind: KindAuth, Message: "no API key: set " + strings.Join(p.KeyEnv(), " or ") + " in the environment"}
	}
	mt := ep.MaxTokens
	if mt == "" {
		mt = "max_tokens"
	}
	return &OpenAI{name: p.Provider, base: strings.TrimRight(base, "/"), key: key, maxTokens: mt, http: &http.Client{}}, nil
}

func (o *OpenAI) Name() string { return o.name }

// openAIBody builds the request body.
func openAIBody(req Request, maxTokensField string) (map[string]any, error) {
	msgs := make([]map[string]string, 0, len(req.Messages)+1)
	if req.System != "" {
		msgs = append(msgs, map[string]string{"role": "system", "content": req.System})
	}
	for _, m := range req.Messages {
		if m.Role != "user" && m.Role != "assistant" {
			return nil, fmt.Errorf("unknown role %q", m.Role)
		}
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	body := map[string]any{"model": req.Model, "messages": msgs, "stream": false}
	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		body["top_p"] = *req.TopP
	}
	if req.TopK != nil {
		body["top_k"] = *req.TopK
	}
	if req.MaxOutputTokens > 0 {
		body[maxTokensField] = req.MaxOutputTokens
	}
	if req.Seed != nil {
		body["seed"] = *req.Seed
	}
	if req.Effort != "" {
		body["reasoning_effort"] = req.Effort
	}
	switch req.Output {
	case "json":
		body["response_format"] = map[string]any{"type": "json_object"}
	case "schema":
		var schema any
		if err := json.Unmarshal(req.Schema, &schema); err != nil {
			return nil, fmt.Errorf("the output schema is not JSON: %v", err)
		}
		body["response_format"] = map[string]any{"type": "json_schema",
			"json_schema": map[string]any{"name": "wag_reply", "schema": schema}}
	}
	for k, v := range req.Extra {
		body[k] = v
	}
	return body, nil
}

// openAIReply is the part of a Chat Completions answer WAG reads. A
// reasoning field some servers add (reasoning_content) is not read.
type openAIReply struct {
	ID                string `json:"id"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Message struct {
			Content *string `json:"content"`
			Refusal *string `json:"refusal"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens        *int `json:"prompt_tokens"`
		CompletionTokens    *int `json:"completion_tokens"`
		TotalTokens         *int `json:"total_tokens"`
		PromptCacheHit      *int `json:"prompt_cache_hit_tokens"` // DeepSeek
		PromptTokensDetails *struct {
			CachedTokens *int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
		CompletionTokensDetails *struct {
			ReasoningTokens *int `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (o *OpenAI) Complete(ctx context.Context, req Request) (*Response, error) {
	body, err := openAIBody(req, o.maxTokens)
	if err != nil {
		return nil, &Error{Provider: o.name, Kind: KindInvalid, Message: err.Error(), Err: err}
	}
	b, _ := json.Marshal(body)
	hr, err := http.NewRequestWithContext(ctx, http.MethodPost, o.base+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, &Error{Provider: o.name, Kind: KindNoProvider, Message: err.Error(), Err: err}
	}
	hr.Header.Set("Content-Type", "application/json")
	if o.key != "" {
		hr.Header.Set("Authorization", "Bearer "+o.key)
	}
	resp, err := o.http.Do(hr)
	if err != nil {
		return nil, AsError(o.name, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, AsError(o.name, err)
	}
	reqID := resp.Header.Get("X-Request-Id")
	var r openAIReply
	jerr := json.Unmarshal(raw, &r)
	if resp.StatusCode/100 != 2 {
		e := &Error{Provider: o.name, Kind: kindOfStatus(resp.StatusCode), Status: resp.StatusCode, RequestID: reqID}
		if jerr == nil && r.Error != nil {
			e.Message = strings.TrimSpace(r.Error.Type + " " + r.Error.Message)
		} else {
			e.Message = truncate(string(raw), 500)
		}
		if s := resp.Header.Get("Retry-After"); s != "" {
			if n, err := strconv.Atoi(s); err == nil {
				e.RetryAfter = time.Duration(n) * time.Second
			}
		}
		return nil, e
	}
	if jerr != nil {
		return nil, &Error{Provider: o.name, Kind: KindBadReply, Message: "the answer is not JSON: " + truncate(string(raw), 300), RequestID: reqID}
	}
	out := &Response{Model: r.Model, ResponseID: r.ID, RequestID: reqID}
	meta := map[string]any{}
	if r.SystemFingerprint != "" {
		meta["system_fingerprint"] = r.SystemFingerprint
	}
	for _, h := range []string{"X-Ratelimit-Remaining-Requests", "X-Ratelimit-Remaining-Tokens"} {
		if v := resp.Header.Get(h); v != "" {
			meta[strings.ToLower(h)] = v
		}
	}
	if u := r.Usage; u != nil {
		out.Usage = Usage{InputTokens: u.PromptTokens, OutputTokens: u.CompletionTokens, TotalTokens: u.TotalTokens, CachedTokens: u.PromptCacheHit}
		if d := u.PromptTokensDetails; d != nil && d.CachedTokens != nil {
			out.Usage.CachedTokens = d.CachedTokens
		}
		if d := u.CompletionTokensDetails; d != nil {
			out.Usage.ReasoningTokens = d.ReasoningTokens
		}
	}
	if len(r.Choices) == 0 {
		return nil, &Error{Provider: o.name, Kind: KindEmpty, Message: "no choice in the answer", RequestID: reqID}
	}
	c := r.Choices[0]
	out.FinishReason = c.FinishReason
	if c.Message.Refusal != nil && *c.Message.Refusal != "" {
		return nil, &Error{Provider: o.name, Kind: KindBlocked, Message: "the model refused: " + truncate(*c.Message.Refusal, 300), RequestID: reqID}
	}
	if c.Message.Content != nil {
		text, stripped := stripThink(*c.Message.Content)
		out.Text = text
		if stripped {
			meta["reasoning_block_removed"] = true
		}
	}
	if len(meta) > 0 {
		out.Meta = meta
	}
	if strings.TrimSpace(out.Text) == "" {
		return nil, &Error{Provider: o.name, Kind: KindEmpty, Message: "no text in the answer (finish reason " + out.FinishReason + ")", RequestID: reqID}
	}
	return out, nil
}

// OpenAIModels lists an OpenAI-compatible endpoint's models (GET /models).
func OpenAIModels(ctx context.Context, p Profile) ([]map[string]any, error) {
	o, err := newOpenAI(p)
	if err != nil {
		return nil, err
	}
	hr, err := http.NewRequestWithContext(ctx, http.MethodGet, o.base+"/models", nil)
	if err != nil {
		return nil, err
	}
	if o.key != "" {
		hr.Header.Set("Authorization", "Bearer "+o.key)
	}
	resp, err := o.http.Do(hr)
	if err != nil {
		return nil, AsError(o.name, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode/100 != 2 {
		return nil, &Error{Provider: o.name, Kind: kindOfStatus(resp.StatusCode), Status: resp.StatusCode, Message: truncate(string(raw), 300)}
	}
	var list struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, &Error{Provider: o.name, Kind: KindBadReply, Message: err.Error()}
	}
	return list.Data, nil
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
