package caes

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

// docs/contracts.md, contract 1: inference runs from several goroutines are
// serialized and each gets the result it would get alone (the third review's
// concurrency contracts; run with -race by tools/vv/engine-tests.sh).
func TestConcurrentInference(t *testing.T) {
	build := func(c string) *ArgGraph {
		ag := NewArgGraph()
		ag.Theory.Language = Language{"p/1": "%s is p.", "q/1": "%s is q.", "r/1": "%s is r."}
		ag.Theory.ArgSchemes = []*Scheme{
			{Id: "s", Variables: []string{"X"}, Premises: []string{"p(X)"}, Conclusions: []string{"q(X)"}, Weight: LinkedWeighingFunction},
			{Id: "t", Variables: []string{"X"}, Premises: []string{"q(X)"}, Conclusions: []string{"r(X)"}, Weight: LinkedWeighingFunction},
		}
		ag.Theory.InitSchemeIndex()
		ag.Assumptions = []string{"p(" + c + ")"}
		return ag
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				c := fmt.Sprintf("c%d_%d", i, j)
				ag := build(c)
				if err := ag.Infer(); err != nil {
					t.Errorf("%s: %v", c, err)
					return
				}
				l := ag.GroundedLabelling()
				if st := ag.Statements["r("+c+")"]; st == nil || l[st] != In || len(ag.Arguments) != 2 {
					t.Errorf("%s: r(%s) is not in, or %d arguments instead of 2", c, c, len(ag.Arguments))
				}
			}
		}(i)
	}
	wg.Wait()
}

// A run that waits for the rule engine ends as cancelled when its context is
// done, and leaves the graph as it was.
func TestWaitingRunCanBeCancelled(t *testing.T) {
	if err := acquireEngine(context.Background()); err != nil {
		t.Fatal(err)
	}
	ag := NewArgGraph()
	ag.Theory.Language = Language{"p/1": "%s is p.", "q/1": "%s is q."}
	ag.Theory.ArgSchemes = []*Scheme{{Id: "s", Variables: []string{"X"}, Premises: []string{"p(X)"}, Conclusions: []string{"q(X)"}, Weight: LinkedWeighingFunction}}
	ag.Theory.InitSchemeIndex()
	ag.Assumptions = []string{"p(a)"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ag.InferContext(ctx)
	releaseEngine()
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("a run waiting for the engine with a cancelled context returned %v", err)
	}
	if len(ag.Arguments) != 0 {
		t.Fatalf("the cancelled run changed the graph: %d arguments", len(ag.Arguments))
	}
	if err := ag.Infer(); err != nil || len(ag.Arguments) != 1 {
		t.Fatalf("after the engine is free the run completes: %v, %d arguments", err, len(ag.Arguments))
	}
}
