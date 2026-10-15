package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Replay answers with recorded attempts (from a trace file), in order, and
// checks that each request is the one that was recorded: replaying a game
// with it shows whether the game sends the model exactly what it sent then
// (the prompt is deterministic) and does exactly what it did with the
// answers (the game is).
type Replay struct {
	Provider string // the recorded provider's name
	Recorded []Attempt

	mu         sync.Mutex
	next       int
	Mismatches []string
}

func (r *Replay) Name() string { return r.Provider }

func (r *Replay) Complete(ctx context.Context, req Request) (*Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.next >= len(r.Recorded) {
		return nil, &Error{Provider: r.Provider, Kind: KindBadReply, Message: fmt.Sprintf("the trace has no call %d", r.next+1)}
	}
	a := r.Recorded[r.next]
	r.next++
	if d := RequestDiff(a.Request, req); d != "" {
		r.Mismatches = append(r.Mismatches, fmt.Sprintf("call %d: %s", r.next, d))
	}
	if a.Error != nil {
		e := *a.Error
		return nil, &e
	}
	if a.Response == nil {
		return nil, &Error{Provider: r.Provider, Kind: KindEmpty, Message: "the trace has neither an answer nor an error"}
	}
	resp := *a.Response
	return &resp, nil
}

// Left is how many recorded calls were not asked for.
func (r *Replay) Left() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Recorded) - r.next
}

// RequestDiff says how two requests differ ("" if they do not).
func RequestDiff(a, b Request) string {
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if bytes.Equal(ja, jb) {
		return ""
	}
	switch {
	case a.System != b.System:
		return "the system prompt differs"
	case len(a.Messages) != len(b.Messages):
		return fmt.Sprintf("%d messages, not %d", len(b.Messages), len(a.Messages))
	}
	for i := range a.Messages {
		if a.Messages[i] != b.Messages[i] {
			x, y := a.Messages[i].Content, b.Messages[i].Content
			n := 0
			for n < len(x) && n < len(y) && x[n] == y[n] {
				n++
			}
			return fmt.Sprintf("message %d (%s) differs from character %d: %q / %q", i+1, a.Messages[i].Role, n, excerpt(x, n), excerpt(y, n))
		}
	}
	return "the parameters differ"
}

func excerpt(s string, at int) string {
	end := min(len(s), at+60)
	return s[at:end]
}
