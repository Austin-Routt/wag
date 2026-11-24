package annotate

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/yourorg/wag/internal/llm"
)

// An annotator or a reviewer whose replies keep failing the check is
// stopped (the reviewer's rehearsal of 24 November 2026): after
// badRepliesStop claims in a row whose replies cannot be used, more calls
// would most likely fail the same way.

// shapeRun is a run with one annotator over n claims; bad says which claims
// it answers with a reply that is not an annotation (the gold otherwise).
func shapeRun(t *testing.T, n int, bad func(item string) bool) (*Run, []string, func() map[string]int) {
	t.Helper()
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	items := d.IDs()[1 : 1+n]
	var mu sync.Mutex
	calls := map[string]int{}
	p := llm.Profile{Name: "solo", Provider: "mock", Params: llm.Params{Model: "m1"}}
	c := llm.WithProvider(p, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		id := lastClaim(req)
		mu.Lock()
		calls[id]++
		mu.Unlock()
		if bad(id) {
			return &llm.Response{Text: "Scheme: practical reasoning. Premises: the goal, the means.", Model: "m1"}, nil
		}
		b, _ := json.Marshal(d.Gold[id])
		return &llm.Response{Text: string(b), Model: "m1"}, nil
	}})
	c.Sleep = llm.NoWait
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{p}, Items: items}, map[string]*llm.Client{"solo": c})
	if err != nil {
		t.Fatal(err)
	}
	return r, items, func() map[string]int {
		mu.Lock()
		defer mu.Unlock()
		out := map[string]int{}
		for k, v := range calls {
			out[k] = v
		}
		return out
	}
}

func TestAnAnnotatorWhoseRepliesKeepFailingIsStopped(t *testing.T) {
	r, items, calls := shapeRun(t, 6, func(string) bool { return true })
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	got := calls()
	for i, id := range items {
		want := 0
		if i < badRepliesStop {
			want = 2 // the reply is sent back once, with the reason
		}
		if got[id] != want {
			t.Fatalf("%s was asked %d times, not %d (%v)", id, got[id], want, got)
		}
	}
	st := r.Status()
	p := st.Progress[0]
	if p.Failed != badRepliesStop || p.Done != 0 || !strings.Contains(p.LastError, "in a row") || !strings.Contains(p.LastError, "calls.jsonl") {
		t.Fatalf("progress: %+v", p)
	}
	// The claims after the stop are still to do, not failed.
	if len(r.failures["solo"]) != badRepliesStop || r.failures["solo"][items[badRepliesStop]] != "" {
		t.Fatalf("failures: %v", r.failures["solo"])
	}
	// Started again, it is asked again: the person chose to.
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if got := calls(); got[items[0]] != 4 || got[items[badRepliesStop]] != 0 {
		t.Fatalf("started again: %v", got)
	}
}

// A reply that can be used ends the row: the annotator goes on.
func TestAGoodReplyEndsTheRowOfBadOnes(t *testing.T) {
	var bad map[string]bool
	r, items, calls := shapeRun(t, 6, func(id string) bool { return bad[id] })
	bad = map[string]bool{}
	for i, id := range items {
		if i%3 != 2 { // bad, bad, good, bad, bad, good
			bad[id] = true
		}
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if got := calls(); len(got) != len(items) {
		t.Fatalf("asked about %d claims, not %d: %v", len(got), len(items), got)
	}
	if p := r.Status().Progress[0]; p.Failed != 4 || p.Done != 2 || strings.Contains(p.LastError, "in a row") {
		t.Fatalf("progress: %+v", p)
	}
}

func TestAReviewerWhoseRepliesKeepFailingIsStopped(t *testing.T) {
	r, calls := reviewRun(t, 5, "Field scheme: A is right.")
	err := r.Review(context.Background(), nil, false)
	if err == nil || !strings.Contains(err.Error(), "in a row") {
		t.Fatalf("%v", err)
	}
	if *calls != badRepliesStop {
		t.Fatalf("%d calls", *calls)
	}
	st := r.Status()
	if p := st.Progress[1]; !strings.Contains(p.LastError, "in a row") || st.Reviewed != badRepliesStop {
		t.Fatalf("progress: %+v; %d reviewed", p, st.Reviewed)
	}
	for i, id := range r.Settings.Items {
		if rv := r.reviews[id]; (i < badRepliesStop) != (rv != nil && rv.Error != "") {
			t.Fatalf("%s: %+v", id, rv)
		}
	}
}

// A review that can be used ends the row.
func TestAGoodReviewEndsTheRowOfBadOnes(t *testing.T) {
	r, calls := reviewRun(t, 5, "not a review", "not a review", goodReview, "not a review", "not a review", goodReview)
	if err := r.Review(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	if *calls != 5 || r.Status().Reviewed != 5 {
		t.Fatalf("%d calls, %d reviewed", *calls, r.Status().Reviewed)
	}
}

// A call that fails otherwise (a provider's error, after its retries) ends
// the row too: the replies around it are not counted together.
func TestAFailedCallEndsTheRowOfBadReplies(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	items := d.IDs()[1:7]
	kind := map[string]string{items[0]: "bad", items[1]: "bad", items[2]: "server", items[3]: "bad", items[4]: "bad", items[5]: "good"}
	var mu sync.Mutex
	asked := map[string]bool{}
	p := llm.Profile{Name: "solo", Provider: "mock", Params: llm.Params{Model: "m1"}}
	c := llm.WithProvider(p, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		id := lastClaim(req)
		mu.Lock()
		asked[id] = true
		mu.Unlock()
		switch kind[id] {
		case "bad":
			return &llm.Response{Text: "Scheme: practical reasoning.", Model: "m1"}, nil
		case "server":
			return nil, &llm.Error{Provider: "mock", Kind: llm.KindServer, Message: "overloaded"}
		}
		b, _ := json.Marshal(d.Gold[id])
		return &llm.Response{Text: string(b), Model: "m1"}, nil
	}})
	c.Sleep = llm.NoWait
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{p}, Items: items}, map[string]*llm.Client{"solo": c})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if len(asked) != len(items) {
		t.Fatalf("asked about %d claims, not %d", len(asked), len(items))
	}
	if p := r.Status().Progress[0]; p.Failed != 5 || p.Done != 1 || strings.Contains(p.LastError, "in a row") {
		t.Fatalf("progress: %+v", p)
	}
}

// A reviewer that was stopped, and is started again and succeeds, no longer
// says it is stopped (the review of the repairs, finding 1).
func TestAReviewerStartedAgainNoLongerSaysItIsStopped(t *testing.T) {
	r, _ := reviewRun(t, 4, "Field scheme: A is right.", "Field scheme: A is right.", "Field scheme: A is right.", goodReview)
	if err := r.Review(context.Background(), nil, false); err == nil || !strings.Contains(err.Error(), "in a row") {
		t.Fatalf("%v", err)
	}
	if err := r.Review(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	st := r.Status()
	if p := st.Progress[1]; p.LastError != "" || st.Reviewed != 4 {
		t.Fatalf("progress: %+v; %d reviewed", p, st.Reviewed)
	}
}

// A reviewer's call that fails otherwise ends the row too (finding 3).
func TestAFailedReviewCallEndsTheRowOfBadReviews(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	items := d.IDs()[1:7]
	kind := map[string]string{items[0]: "bad", items[1]: "bad", items[2]: "server", items[3]: "bad", items[4]: "bad", items[5]: "good"}
	cp := &copier{d: d}
	pa := llm.Profile{Name: "solo", Provider: "mock", Params: llm.Params{Model: "m1"}}
	pr := llm.Profile{Name: "judge", Provider: "mock", Params: llm.Params{Model: "m2"}}
	var mu sync.Mutex
	asked := map[string]bool{}
	judge := llm.WithProvider(pr, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		id := lastClaim(req)
		mu.Lock()
		asked[id] = true
		mu.Unlock()
		switch kind[id] {
		case "bad":
			return &llm.Response{Text: "Field scheme: A is right."}, nil
		case "server":
			return nil, &llm.Error{Provider: "mock", Kind: llm.KindServer, Message: "overloaded"}
		}
		return &llm.Response{Text: goodReview}, nil
	}})
	judge.Sleep = llm.NoWait
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{pa}, Reviewer: &pr, Items: items},
		map[string]*llm.Client{"solo": llm.WithProvider(pa, cp.provider()), "judge": judge})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err := r.Review(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	if len(asked) != len(items) || r.Status().Reviewed != len(items) {
		t.Fatalf("asked about %d claims, %d reviewed", len(asked), r.Status().Reviewed)
	}
}
