package llm

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
)

// Anthropic calls the Claude API through Anthropic's SDK (Messages). A
// declined request is served again by a model the API picks for the
// refusal's category (server-side fallback), instead of failing; the
// response's model says which served it.
type Anthropic struct {
	client anthropic.Client
}

func newAnthropic(p Profile) (*Anthropic, error) {
	// The SDK retries on its own by default; WAG retries (Client), so that
	// every attempt is traced.
	opts := []option.RequestOption{option.WithMaxRetries(0)}
	if p.APIKeyEnv != "" { // else the SDK's own: ANTHROPIC_API_KEY, or an `ant auth login` profile
		key, _ := p.key()
		if key == "" {
			return nil, &Error{Provider: "anthropic", Kind: KindAuth, Message: "no API key: set " + p.APIKeyEnv + " in the environment"}
		}
		opts = append(opts, option.WithAPIKey(key))
	}
	if p.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(p.BaseURL))
	}
	return &Anthropic{client: anthropic.NewClient(opts...)}, nil
}

func (a *Anthropic) Name() string { return "anthropic" }

func (a *Anthropic) Complete(ctx context.Context, req Request) (*Response, error) {
	msgs := make([]anthropic.BetaMessageParam, 0, len(req.Messages))
	for _, m := range req.Messages {
		switch m.Role {
		case "user":
			msgs = append(msgs, anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(m.Content)))
		case "assistant":
			msgs = append(msgs, anthropic.BetaMessageParam{Role: anthropic.BetaMessageParamRoleAssistant,
				Content: []anthropic.BetaContentBlockParamUnion{anthropic.NewBetaTextBlock(m.Content)}})
		default:
			return nil, &Error{Provider: "anthropic", Kind: KindInvalid, Message: "unknown role " + m.Role}
		}
	}
	params := anthropic.BetaMessageNewParams{
		Model:     anthropic.Model(req.Model),
		MaxTokens: 16000,
		Messages:  msgs,
		Fallbacks: anthropic.BetaFallbacksParamUnion{OfDefault: constant.ValueOf[constant.Default]()},
		Betas:     []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
	}
	if req.System != "" {
		params.System = []anthropic.BetaTextBlockParam{{Text: req.System}}
	}
	if req.MaxOutputTokens > 0 {
		params.MaxTokens = int64(req.MaxOutputTokens)
	}
	if req.Effort != "" {
		params.OutputConfig = anthropic.BetaOutputConfigParam{Effort: anthropic.BetaOutputConfigEffort(req.Effort)}
	}
	if req.Temperature != nil {
		params.Temperature = anthropic.Float(*req.Temperature)
	}
	if req.TopP != nil {
		params.TopP = anthropic.Float(*req.TopP)
	}
	if req.TopK != nil {
		params.TopK = anthropic.Int(int64(*req.TopK))
	}
	var httpResp *http.Response
	resp, err := a.client.Beta.Messages.New(ctx, params, option.WithResponseInto(&httpResp))
	if err != nil {
		var ae *anthropic.Error
		if errors.As(err, &ae) {
			e := &Error{Provider: "anthropic", Kind: kindOfStatus(ae.StatusCode), Status: ae.StatusCode, RequestID: ae.RequestID,
				Message: truncate(err.Error(), 500), Err: err}
			if ae.StatusCode == 529 { // overloaded
				e.Kind = KindServer
			}
			return nil, e
		}
		return nil, AsError("anthropic", err)
	}
	out := &Response{Model: string(resp.Model), ResponseID: resp.ID, FinishReason: string(resp.StopReason),
		Usage: Usage{InputTokens: intp(int(resp.Usage.InputTokens)), OutputTokens: intp(int(resp.Usage.OutputTokens)),
			CachedTokens: intp(int(resp.Usage.CacheReadInputTokens))}}
	if httpResp != nil {
		out.RequestID = httpResp.Header.Get("Request-Id")
	}
	if resp.StopReason == anthropic.BetaStopReasonRefusal {
		return nil, &Error{Provider: "anthropic", Kind: KindBlocked, RequestID: out.RequestID,
			Message: "Claude declined (" + string(resp.StopDetails.Category) + ")"}
	}
	var text strings.Builder
	for _, b := range resp.Content {
		if t, ok := b.AsAny().(anthropic.BetaTextBlock); ok {
			text.WriteString(t.Text)
		}
	}
	out.Text = text.String()
	if strings.TrimSpace(out.Text) == "" {
		return nil, &Error{Provider: "anthropic", Kind: KindEmpty, RequestID: out.RequestID, Message: "no text in the answer (stop reason " + out.FinishReason + ")"}
	}
	return out, nil
}
