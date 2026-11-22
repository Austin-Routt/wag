package chr

import (
	"errors"
	"io"
	"os"
	"testing"

	. "github.com/hfried/GoCHR/src/engine/terms"
)

// A program that needs exactly max rule applications is not interrupted.
func TestLimitExactlyMaxIsComplete(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.AddRule("down", nil, []string{"n(X)"}, []string{"X > 0"}, []string{"Y is X - 1", "n(Y)"})
	ok, store, err := rs.Infer([]string{"n(5)"}, 5)
	if err != nil || !ok {
		t.Fatalf("five applications under a limit of five: ok=%v err=%v store=%v", ok, err, store)
	}
	if count(store, "n(0)") != 1 {
		t.Errorf("want n(0) in the store, got %v", store)
	}
}

// One application more than the limit allows is an interruption.
func TestLimitReachedWithWorkLeft(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.AddRule("down", nil, []string{"n(X)"}, []string{"X > 0"}, []string{"Y is X - 1", "n(Y)"})
	ok, _, err := rs.Infer([]string{"n(6)"}, 5)
	if ok || !errors.Is(err, ErrLimit) {
		t.Fatalf("six applications under a limit of five: want ErrLimit, got ok=%v err=%v", ok, err)
	}
	if !errors.Is(rs.Stopped, ErrLimit) {
		t.Errorf("rs.Stopped = %v, want ErrLimit", rs.Stopped)
	}
	// the next run starts afresh
	ok, _, err = rs.Infer([]string{"n(2)"}, 5)
	if !ok || err != nil || rs.Stopped != nil {
		t.Errorf("a second, short run: ok=%v err=%v Stopped=%v", ok, err, rs.Stopped)
	}
}

// A term that grows without end is stopped by MaxTermSize long before the
// rule limit.
func TestMaxTermSizeStopsGrowingTerm(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.MaxTermSize = 50
	rs.AddRule("grow", nil, []string{"t(X)"}, nil, []string{"t(f(X))"})
	ok, store, err := rs.Infer([]string{"t(a)"}, 100000)
	if ok || !errors.Is(err, ErrLimit) {
		t.Fatalf("want ErrLimit, got ok=%v err=%v", ok, err)
	}
	for _, s := range store {
		if len(s) > 2*50+10 {
			t.Errorf("a constraint larger than the limit is in the store: %d characters", len(s))
		}
	}
}

func TestTermSize(t *testing.T) {
	for _, c := range []struct {
		term string
		want int
	}{
		{"a", 1},
		{"p(a)", 2},
		{"p(a, f(b, c))", 5},
		{"p([a, b, c])", 5},
		{"p(X, 1, \"s\")", 4},
	} {
		g, err := parseGoals([]string{c.term})
		if err != nil || len(g) != 1 {
			t.Fatalf("parse %s: %v", c.term, err)
		}
		if got := termSize(*g[0], 1000); got != c.want {
			t.Errorf("termSize(%s) = %d, want %d", c.term, got, c.want)
		}
	}
	g, _ := parseGoals([]string{"p(a, b, c, d, e, f, g)"})
	if got := termSize(*g[0], 3); got > 5 {
		t.Errorf("termSize stops counting past its limit: got %d", got)
	}
}

// A closed Done channel stops the run.
func TestDoneCancels(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	done := make(chan struct{})
	close(done)
	rs.Done = done
	rs.AddRule("up", nil, []string{"n(X)"}, nil, []string{"Y is X + 1", "n(Y)"})
	ok, _, err := rs.Infer([]string{"n(0)"}, 100000)
	if ok || !errors.Is(err, ErrCancelled) {
		t.Fatalf("want ErrCancelled, got ok=%v err=%v", ok, err)
	}
}

// The solver writes nothing to stdout when it stops, unless tracing.
func TestLimitIsSilent(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.AddRule("up", nil, []string{"n(X)"}, nil, []string{"Y is X + 1", "n(Y)"})
	out := captureStdout(t, func() { rs.Infer([]string{"n(0)"}, 100) })
	if out != "" {
		t.Errorf("the solver wrote to stdout: %q", out)
	}
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() { b, _ := io.ReadAll(r); done <- string(b) }()
	f()
	os.Stdout = old
	w.Close()
	return <-done
}
