// Package llm is WAG's interface to language models: one request and
// response format for every provider, adapters for the providers (Gemini
// through Google's GenAI SDK, Claude through Anthropic's SDK, any
// OpenAI-compatible endpoint: OpenAI, DeepSeek, LM Studio, llama.cpp,
// SGLang, vLLM, Ollama; an external command for other harnesses), and a
// Client that spaces, times out and retries calls and reports every attempt
// to the caller, which traces it.
//
// The game does not import a provider's types: a seat sends the
// conversation (Request) and gets text back (Response), and the game's own
// parser and validator decide what the text means.
//
// API keys are read from environment variables, which a Profile names
// (APIKeyEnv); they are never part of a profile, a request or a trace.
package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// Message is one turn of a conversation.
type Message struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// Params are the generation parameters. A nil or zero value leaves the
// provider's default, and is not sent.
type Params struct {
	Model           string   `json:"model"`
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"top_p,omitempty"`
	TopK            *int     `json:"top_k,omitempty"`
	MaxOutputTokens int      `json:"max_output_tokens,omitempty"`
	Seed            *int64   `json:"seed,omitempty"`
	// Effort is how much the model reasons before it answers, in the
	// provider's terms: Claude's effort (low, medium, high, xhigh, max),
	// Gemini's thinking level (minimal, low, medium, high), OpenAI's
	// reasoning_effort.
	Effort string `json:"effort,omitempty"`
	// ThinkingBudget is Gemini 2.5's thinking budget in tokens (0 turns
	// thinking off where the model allows it, -1 lets the model decide).
	ThinkingBudget *int `json:"thinking_budget,omitempty"`
	// Output is "text" (the default), "json" (the provider's JSON mode) or
	// "schema" (JSON constrained to Schema, where the provider supports it).
	Output string `json:"output,omitempty"`
	// Schema is the JSON Schema for Output "schema".
	Schema json.RawMessage `json:"schema,omitempty"`
	// Extra are fields added to the provider's request body as they are: an
	// engine's own options (llama.cpp's grammar, a chat template's
	// arguments). Gemini merges them into its request (HTTPOptions.ExtraBody).
	Extra map[string]any `json:"extra,omitempty"`
}

// Request is one call: the system prompt, the whole conversation so far,
// and the parameters.
type Request struct {
	System   string    `json:"system,omitempty"`
	Messages []Message `json:"messages"`
	Params
}

// Usage counts tokens, as the provider reports them. A nil count was not
// reported (unknown), which is not zero.
type Usage struct {
	InputTokens     *int `json:"input_tokens,omitempty"`
	OutputTokens    *int `json:"output_tokens,omitempty"`    // the answer's tokens
	ReasoningTokens *int `json:"reasoning_tokens,omitempty"` // tokens spent reasoning, when reported apart
	CachedTokens    *int `json:"cached_tokens,omitempty"`    // input tokens served from a cache
	TotalTokens     *int `json:"total_tokens,omitempty"`
}

// Response is what the model answered.
type Response struct {
	// Text is the answer: the text parts of the reply, without a reasoning
	// block that some engines put before it (<think>…</think>), which is not
	// kept.
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason,omitempty"` // in the provider's words: STOP, end_turn, stop, MAX_TOKENS…
	Model        string `json:"model,omitempty"`         // the model (version) that served the request, when reported
	ResponseID   string `json:"response_id,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
	Usage        Usage  `json:"usage"`
	// Meta are other diagnostics of the provider: rate-limit headers, a
	// system fingerprint, a block reason, safety categories.
	Meta map[string]any `json:"meta,omitempty"`
}

// Provider is a model service.
type Provider interface {
	// Name is the provider's name: gemini, anthropic, openai…
	Name() string
	// Complete sends the request and returns the answer. Errors are *Error.
	Complete(ctx context.Context, req Request) (*Response, error)
}

// Error kinds.
const (
	KindAuth       = "auth"            // the key is missing, wrong or not allowed
	KindRateLimit  = "rate_limit"      // too many requests, or a quota is used up
	KindInvalid    = "invalid_request" // the provider rejects the request
	KindNotFound   = "not_found"       // no such model
	KindServer     = "server"          // the provider failed
	KindTimeout    = "timeout"
	KindCanceled   = "canceled"
	KindNetwork    = "network"
	KindBlocked    = "blocked" // the provider declined to answer (safety, refusal)
	KindEmpty      = "empty"   // an answer without text
	KindBadReply   = "bad_reply"
	KindNoProvider = "config"
)

// Error is a failed call.
type Error struct {
	Provider   string        `json:"provider"`
	Kind       string        `json:"kind"`
	Status     int           `json:"status,omitempty"` // HTTP status, if any
	Message    string        `json:"message"`
	RetryAfter time.Duration `json:"retry_after,omitempty"` // what the provider asked us to wait
	RequestID  string        `json:"request_id,omitempty"`
	Err        error         `json:"-"`
}

func (e *Error) Error() string {
	s := e.Provider + ": " + e.Kind
	if e.Status != 0 {
		s += fmt.Sprintf(" (HTTP %d)", e.Status)
	}
	if e.Message != "" {
		s += ": " + e.Message
	}
	return s
}

func (e *Error) Unwrap() error { return e.Err }

// Retryable reports whether the same request may succeed later.
func (e *Error) Retryable() bool {
	switch e.Kind {
	case KindRateLimit:
		return !e.Exhausted()
	case KindServer, KindTimeout, KindNetwork:
		return true
	}
	return false
}

// MaxRetryWait is the longest wait a retry is worth. Per-minute limits clear
// within a minute; a provider that asks for longer has a quota used up (a
// daily one, say), and the next call would fail too.
const MaxRetryWait = 2 * time.Minute

// Exhausted reports whether the error is a quota used up for longer than
// MaxRetryWait: the provider asks to wait that long before the next call.
func (e *Error) Exhausted() bool {
	return e.Kind == KindRateLimit && e.RetryAfter > MaxRetryWait
}

// AsError returns err as an *Error of the provider: as it is if it is one,
// else classified from the context (timeout, canceled) or as a network
// error.
func AsError(provider string, err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	kind := KindNetwork
	var ne net.Error
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		kind = KindTimeout
	case errors.Is(err, context.Canceled):
		kind = KindCanceled
	case errors.As(err, &ne) && ne.Timeout():
		kind = KindTimeout
	}
	return &Error{Provider: provider, Kind: kind, Message: err.Error(), Err: err}
}

// kindOfStatus classifies an HTTP status.
func kindOfStatus(status int) string {
	switch {
	case status == 401 || status == 403:
		return KindAuth
	case status == 404:
		return KindNotFound
	case status == 408:
		return KindTimeout
	case status == 429:
		return KindRateLimit
	case status >= 500:
		return KindServer
	default:
		return KindInvalid
	}
}

// intp returns a pointer to n.
func intp(n int) *int { return &n }

// stripThink removes a reasoning block that some engines (Qwen, DeepSeek
// R1 distills on llama.cpp or LM Studio) put at the start of the answer
// text. It reports whether there was one.
func stripThink(s string) (string, bool) {
	t := strings.TrimLeft(s, " \t\r\n")
	if !strings.HasPrefix(t, "<think>") {
		return s, false
	}
	end := strings.Index(t, "</think>")
	if end < 0 {
		return "", true // the answer never came
	}
	return strings.TrimLeft(t[end+len("</think>"):], " \t\r\n"), true
}
