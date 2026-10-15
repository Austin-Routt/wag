package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"google.golang.org/genai"
)

// Gemini calls the Gemini Developer API through Google's GenAI SDK.
type Gemini struct {
	client *genai.Client
}

func newGemini(p Profile) (*Gemini, error) {
	key, _ := p.key()
	if key == "" {
		return nil, &Error{Provider: "gemini", Kind: KindAuth,
			Message: "no API key: set " + strings.Join(p.KeyEnv(), " or ") + " in the environment (from Google AI Studio)"}
	}
	cc := &genai.ClientConfig{APIKey: key, Backend: genai.BackendGeminiAPI}
	if p.BaseURL != "" {
		cc.HTTPOptions.BaseURL = p.BaseURL
	}
	// The key is given here, so the SDK does not choose between
	// GEMINI_API_KEY and GOOGLE_API_KEY itself (Profile.KeyEnv says which
	// comes first); the backend is given, so GOOGLE_GENAI_USE_VERTEXAI does
	// not switch it.
	c, err := genai.NewClient(context.Background(), cc)
	if err != nil {
		return nil, &Error{Provider: "gemini", Kind: KindNoProvider, Message: err.Error(), Err: err}
	}
	return &Gemini{client: c}, nil
}

func (g *Gemini) Name() string { return "gemini" }

// geminiConfig maps the request's parameters to the SDK's.
func geminiConfig(req Request) (*genai.GenerateContentConfig, error) {
	cfg := &genai.GenerateContentConfig{}
	if req.System != "" {
		cfg.SystemInstruction = &genai.Content{Parts: []*genai.Part{{Text: req.System}}}
	}
	f32 := func(f float64) *float32 { v := float32(f); return &v }
	if req.Temperature != nil {
		cfg.Temperature = f32(*req.Temperature)
	}
	if req.TopP != nil {
		cfg.TopP = f32(*req.TopP)
	}
	if req.TopK != nil {
		cfg.TopK = f32(float64(*req.TopK))
	}
	if req.MaxOutputTokens > 0 {
		cfg.MaxOutputTokens = int32(min(req.MaxOutputTokens, math.MaxInt32))
	}
	if req.Seed != nil {
		if *req.Seed > math.MaxInt32 || *req.Seed < math.MinInt32 {
			return nil, fmt.Errorf("the seed must fit in 32 bits for Gemini: %d", *req.Seed)
		}
		s := int32(*req.Seed)
		cfg.Seed = &s
	}
	if req.Effort != "" || req.ThinkingBudget != nil {
		// The model's thoughts are not asked for (IncludeThoughts stays
		// false): only how much it thinks.
		tc := &genai.ThinkingConfig{}
		if req.Effort != "" {
			tc.ThinkingLevel = genai.ThinkingLevel(strings.ToUpper(req.Effort))
		}
		if req.ThinkingBudget != nil {
			b := int32(*req.ThinkingBudget)
			tc.ThinkingBudget = &b
		}
		cfg.ThinkingConfig = tc
	}
	switch req.Output {
	case "json":
		cfg.ResponseMIMEType = "application/json"
	case "schema":
		var schema any
		if err := json.Unmarshal(req.Schema, &schema); err != nil {
			return nil, fmt.Errorf("the output schema is not JSON: %v", err)
		}
		cfg.ResponseMIMEType, cfg.ResponseJsonSchema = "application/json", schema
	}
	if len(req.Extra) > 0 {
		cfg.HTTPOptions = &genai.HTTPOptions{ExtraBody: req.Extra}
	}
	return cfg, nil
}

// geminiContents maps the conversation: the assistant is the "model".
func geminiContents(msgs []Message) ([]*genai.Content, error) {
	out := make([]*genai.Content, 0, len(msgs))
	for _, m := range msgs {
		role := genai.RoleUser
		switch m.Role {
		case "user":
		case "assistant":
			role = genai.RoleModel
		default:
			return nil, fmt.Errorf("unknown role %q", m.Role)
		}
		out = append(out, &genai.Content{Role: role, Parts: []*genai.Part{{Text: m.Content}}})
	}
	return out, nil
}

func (g *Gemini) Complete(ctx context.Context, req Request) (*Response, error) {
	cfg, err := geminiConfig(req)
	if err != nil {
		return nil, &Error{Provider: "gemini", Kind: KindInvalid, Message: err.Error(), Err: err}
	}
	contents, err := geminiContents(req.Messages)
	if err != nil {
		return nil, &Error{Provider: "gemini", Kind: KindInvalid, Message: err.Error(), Err: err}
	}
	resp, err := g.client.Models.GenerateContent(ctx, req.Model, contents, cfg)
	if err != nil {
		return nil, geminiError(err)
	}
	out := &Response{Model: resp.ModelVersion, ResponseID: resp.ResponseID}
	if h := resp.SDKHTTPResponse; h != nil {
		for _, k := range []string{"X-Request-Id", "X-Goog-Request-Id"} {
			if v := h.Headers.Get(k); v != "" {
				out.RequestID = v
				break
			}
		}
	}
	if u := resp.UsageMetadata; u != nil {
		out.Usage = Usage{InputTokens: intp(int(u.PromptTokenCount)), OutputTokens: intp(int(u.CandidatesTokenCount)),
			TotalTokens: intp(int(u.TotalTokenCount))}
		if u.ThoughtsTokenCount > 0 {
			out.Usage.ReasoningTokens = intp(int(u.ThoughtsTokenCount))
		}
		if u.CachedContentTokenCount > 0 {
			out.Usage.CachedTokens = intp(int(u.CachedContentTokenCount))
		}
	}
	if pf := resp.PromptFeedback; pf != nil && pf.BlockReason != "" {
		out.Meta = map[string]any{"block_reason": string(pf.BlockReason), "block_message": pf.BlockReasonMessage}
	}
	if len(resp.Candidates) == 0 {
		e := &Error{Provider: "gemini", Kind: KindEmpty, Message: "no candidate in the response", RequestID: out.RequestID}
		if out.Meta != nil {
			e.Kind, e.Message = KindBlocked, fmt.Sprintf("the prompt was blocked (%v)", out.Meta["block_reason"])
		}
		return nil, e
	}
	c := resp.Candidates[0]
	out.FinishReason = string(c.FinishReason)
	var text strings.Builder
	if c.Content != nil {
		for _, p := range c.Content.Parts {
			if p != nil && !p.Thought {
				text.WriteString(p.Text)
			}
		}
	}
	out.Text = text.String()
	if c.FinishMessage != "" {
		if out.Meta == nil {
			out.Meta = map[string]any{}
		}
		out.Meta["finish_message"] = c.FinishMessage
	}
	if strings.TrimSpace(out.Text) == "" {
		kind := KindEmpty
		switch c.FinishReason {
		case genai.FinishReasonSafety, genai.FinishReasonProhibitedContent, genai.FinishReasonBlocklist, genai.FinishReasonSPII, genai.FinishReasonRecitation:
			kind = KindBlocked
		}
		return nil, &Error{Provider: "gemini", Kind: kind, Message: "no text in the answer (finish reason " + out.FinishReason + ")", RequestID: out.RequestID}
	}
	return out, nil
}

// geminiError classifies an SDK error; a rate limit carries the delay the
// API asks for (google.rpc.RetryInfo).
func geminiError(err error) *Error {
	var ae genai.APIError
	if !errors.As(err, &ae) {
		var pae *genai.APIError
		if !errors.As(err, &pae) || pae == nil {
			return AsError("gemini", err)
		}
		ae = *pae
	}
	e := &Error{Provider: "gemini", Kind: kindOfStatus(ae.Code), Status: ae.Code, Message: strings.TrimSpace(ae.Status + " " + ae.Message), Err: err}
	for _, d := range ae.Details {
		if t, _ := d["@type"].(string); strings.HasSuffix(t, "google.rpc.RetryInfo") {
			if s, _ := d["retryDelay"].(string); s != "" {
				if v, perr := time.ParseDuration(s); perr == nil {
					e.RetryAfter = v
				}
			}
		}
	}
	return e
}

// GeminiModels lists the models the key can use, with what they support
// (wag-llm models).
func GeminiModels(ctx context.Context, p Profile) ([]map[string]any, error) {
	g, err := newGemini(p)
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	for m, err := range g.client.Models.All(ctx) {
		if err != nil {
			return out, geminiError(err)
		}
		out = append(out, map[string]any{"name": strings.TrimPrefix(m.Name, "models/"), "display_name": m.DisplayName,
			"input_token_limit": m.InputTokenLimit, "output_token_limit": m.OutputTokenLimit, "thinking": m.Thinking,
			"actions": m.SupportedActions})
	}
	return out, nil
}
