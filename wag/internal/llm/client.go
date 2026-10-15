package llm

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// New makes the provider of a profile. It reads the key from the
// environment; a provider that needs one fails here, before any call, if
// it is missing.
func New(p Profile) (Provider, error) {
	if err := p.Check(); err != nil {
		return nil, err
	}
	switch p.Provider {
	case "gemini":
		return newGemini(p)
	case "anthropic":
		return newAnthropic(p)
	case "command":
		return &Command{Args: p.Command}, nil
	case "mock":
		return &Mock{Replies: p.Replies}, nil
	}
	return newOpenAI(p)
}

// Attempt is one call to a provider, as it is traced: the request, the
// answer or the error, when and how long.
type Attempt struct {
	Profile  string    `json:"profile"`
	Provider string    `json:"provider"`
	N        int       `json:"n"` // 1 for the first try of this request, 2… for the transport retries
	Start    time.Time `json:"start"`
	// Waited is how long the call waited for the profile's min_interval, or
	// for a retry's delay, before it was sent.
	Waited   time.Duration `json:"waited_ns,omitempty"`
	Latency  time.Duration `json:"latency_ns"`
	Request  Request       `json:"request"`
	Response *Response     `json:"response,omitempty"`
	Error    *Error        `json:"error,omitempty"`
}

// Client calls a provider for a profile: it spaces the calls
// (min_interval), bounds each (timeout), retries the ones that may pass
// (rate limits, server errors, timeouts), and reports every attempt.
// It is safe for concurrent use.
type Client struct {
	Profile  Profile
	Provider Provider

	// Sleep waits before a retry or for min_interval (nil: really wait).
	// A replay, which has nothing to wait for, sets it to NoWait.
	Sleep func(ctx context.Context, d time.Duration) error

	mu   sync.Mutex
	last time.Time
}

// NoWait is a Client.Sleep that does not wait.
func NoWait(ctx context.Context, _ time.Duration) error { return ctx.Err() }

// NewClient makes the client of a profile.
func NewClient(p Profile) (*Client, error) {
	prov, err := New(p)
	if err != nil {
		return nil, err
	}
	if p.Output == "schema" && len(p.Schema) == 0 {
		b, err := os.ReadFile(p.SchemaFile)
		if err != nil {
			return nil, fmt.Errorf("profile %s: %v", p.Name, err)
		}
		p.Schema = b
	}
	return &Client{Profile: p, Provider: prov}, nil
}

// WithProvider makes a client of the profile with a provider given (tests,
// or a provider built by hand).
func WithProvider(p Profile, prov Provider) *Client { return &Client{Profile: p, Provider: prov} }

// Request completes a request with the profile's parameters and system
// prompt (the profile's own, if it has one, replaces system).
func (c *Client) Request(system string, msgs []Message) Request {
	if c.Profile.System != "" {
		system = c.Profile.System
	}
	return Request{System: system, Messages: append([]Message(nil), msgs...), Params: c.Profile.Params}
}

func defaultSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Complete sends the request, retrying as the profile says, and calls
// report (if not nil) after every attempt. The error is the last attempt's.
func (c *Client) Complete(ctx context.Context, req Request, report func(Attempt)) (*Response, error) {
	sleep := c.Sleep
	if sleep == nil {
		sleep = defaultSleep
	}
	retries := c.Profile.Retries
	if retries == 0 {
		retries = 3
	}
	if retries < 0 {
		retries = 0
	}
	timeout := time.Duration(c.Profile.Timeout)
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	var delay time.Duration
	for n := 1; ; n++ {
		waited := delay
		if err := sleep(ctx, delay); err != nil {
			return nil, AsError(c.Provider.Name(), err)
		}
		waited += c.space(ctx, sleep)
		if err := ctx.Err(); err != nil {
			return nil, AsError(c.Provider.Name(), err)
		}
		a := Attempt{Profile: c.Profile.Name, Provider: c.Provider.Name(), N: n, Start: time.Now().UTC(), Waited: waited, Request: req}
		cctx, cancel := context.WithTimeout(ctx, timeout)
		start := time.Now()
		resp, err := c.Provider.Complete(cctx, req)
		a.Latency = time.Since(start)
		cancel()
		if err != nil {
			a.Error = AsError(c.Provider.Name(), err)
			if ctx.Err() != nil { // the caller gave up, not the call
				a.Error = AsError(c.Provider.Name(), ctx.Err())
			}
		} else {
			a.Response = resp
		}
		if report != nil {
			report(a)
		}
		if a.Error == nil {
			return resp, nil
		}
		if !a.Error.Retryable() || n > retries || ctx.Err() != nil {
			return nil, a.Error
		}
		delay = a.Error.RetryAfter
		if delay <= 0 {
			delay = time.Duration(1<<uint(n)) * time.Second // 2s, 4s, 8s…
		}
		if delay > 2*time.Minute {
			delay = 2 * time.Minute
		}
	}
}

// space waits until min_interval has passed since the previous call's
// start, and returns how long it waited.
func (c *Client) space(ctx context.Context, sleep func(context.Context, time.Duration) error) time.Duration {
	gap := time.Duration(c.Profile.MinInterval)
	if gap <= 0 {
		return 0
	}
	c.mu.Lock()
	wait := time.Until(c.last.Add(gap))
	if wait < 0 {
		wait = 0
	}
	c.last = time.Now().Add(wait)
	c.mu.Unlock()
	_ = sleep(ctx, wait)
	return wait
}
