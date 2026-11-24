package llm

import (
	"context"
	"errors"
	"testing"
	"time"
)

// A gate is asked before each attempt: it can let the first attempt go and
// stop the retry, and a call it stops sends nothing and reports no attempt.
func TestGateStopsTheNextAttempt(t *testing.T) {
	n := 0
	m := &Mock{Fn: func(Request) (*Response, error) {
		n++
		return nil, &Error{Provider: "mock", Kind: KindRateLimit, Message: "slow down", RetryAfter: time.Millisecond}
	}}
	c := WithProvider(Profile{Name: "p", Provider: "mock", Params: Params{Model: "m"}}, m)
	c.Sleep = NoWait
	open := true
	stop := errors.New("stopped")
	ctx := WithGate(context.Background(), func() error {
		if !open {
			return stop
		}
		return nil
	})
	var reported int
	_, err := c.Complete(ctx, c.Request("", []Message{{Role: "user", Content: "x"}}), func(Attempt) { reported++; open = false })
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindCanceled || !errors.Is(err, stop) || n != 1 || reported != 1 {
		t.Fatalf("%v (%d sent, %d reported)", err, n, reported)
	}
	// Closed from the start: nothing is sent, nothing reported.
	n, reported, open = 0, 0, false
	if _, err := c.Complete(ctx, c.Request("", []Message{{Role: "user", Content: "x"}}), func(Attempt) { reported++ }); err == nil || n != 0 || reported != 0 {
		t.Fatalf("%v (%d sent, %d reported)", err, n, reported)
	}
	// Without a gate, the retries go on as before.
	n = 0
	if _, err := c.Complete(context.Background(), c.Request("", []Message{{Role: "user", Content: "x"}}), nil); err == nil || n != 4 {
		t.Fatalf("%v (%d sent)", err, n)
	}
}
