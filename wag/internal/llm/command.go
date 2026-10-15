package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// Command runs an external program for each call: the Request as JSON on
// its standard input; on its standard output either a Response as JSON
// ({"text": …, "finish_reason": …, "model": …, "usage": {…}}) or plain
// text, which is the answer. A non-zero exit is an error, with the end of
// its standard error. Any harness can play a seat this way: Guidance or
// DSPy programs, a constrained-decoding wrapper, a local model without a
// server, a human relay.
type Command struct {
	Args []string
}

func (c *Command) Name() string { return "command" }

func (c *Command) Complete(ctx context.Context, req Request) (*Response, error) {
	in, err := json.Marshal(req)
	if err != nil {
		return nil, &Error{Provider: "command", Kind: KindInvalid, Message: err.Error(), Err: err}
	}
	cmd := exec.CommandContext(ctx, c.Args[0], c.Args[1:]...)
	cmd.Stdin = bytes.NewReader(in)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, AsError("command", ctx.Err())
		}
		return nil, &Error{Provider: "command", Kind: KindServer, Message: fmt.Sprintf("%v: %s", err, tail(stderr.String(), 500)), Err: err}
	}
	out := strings.TrimSpace(stdout.String())
	var r Response
	if strings.HasPrefix(out, "{") && json.Unmarshal([]byte(out), &r) == nil && r.Text != "" {
		return &r, nil
	}
	if out == "" {
		return nil, &Error{Provider: "command", Kind: KindEmpty, Message: "no output"}
	}
	return &Response{Text: out}, nil
}

func tail(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return "…" + s[len(s)-n:]
}

// Mock answers from a script, without a network: each call gets the next
// of Replies (the last one again when they run out), or what Fn returns.
// It keeps the requests it got.
type Mock struct {
	Replies []string
	Fn      func(Request) (*Response, error)

	mu       sync.Mutex
	Requests []Request
}

func (m *Mock) Name() string { return "mock" }

func (m *Mock) Complete(ctx context.Context, req Request) (*Response, error) {
	m.mu.Lock()
	n := len(m.Requests)
	m.Requests = append(m.Requests, req)
	m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, AsError("mock", err)
	}
	if m.Fn != nil {
		return m.Fn(req)
	}
	if len(m.Replies) == 0 {
		return &Response{Text: `{"summary":"I pass.","action":{"type":"abstain"}}`, FinishReason: "stop", Model: "mock"}, nil
	}
	return &Response{Text: m.Replies[min(n, len(m.Replies)-1)], FinishReason: "stop", Model: "mock"}, nil
}
