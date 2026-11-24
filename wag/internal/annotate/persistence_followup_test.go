package annotate

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/wag/internal/llm"
)

// The persistence contract (durable.go) beyond the review's probes
// (annotation_persistence_regression_test.go).

// copier answers each claim with its gold annotation and counts the calls
// per claim.
type copier struct {
	d     *Dataset
	mu    sync.Mutex
	calls map[string]int
	// before, if set, runs before each answer (a test holds a call there).
	before func(item string)
	// fail, if set, answers instead when it returns an error.
	fail func(item string, n int) error
}

func (c *copier) provider() *llm.Mock {
	return &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		id := lastClaim(req)
		c.mu.Lock()
		if c.calls == nil {
			c.calls = map[string]int{}
		}
		c.calls[id]++
		n := c.calls[id]
		c.mu.Unlock()
		if c.before != nil {
			c.before(id)
		}
		if c.fail != nil {
			if err := c.fail(id, n); err != nil {
				return nil, err
			}
		}
		b, _ := json.Marshal(c.d.Gold[id])
		return &llm.Response{Text: string(b), Model: "m1"}, nil
	}}
}

func (c *copier) total() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for _, k := range c.calls {
		n += k
	}
	return n
}

func (c *copier) of(id string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls[id]
}

func soloRun(t *testing.T, n int) (*Run, *copier, []string) {
	t.Helper()
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	items := d.IDs()[1 : 1+n]
	cp := &copier{d: d}
	p := llm.Profile{Name: "solo", Provider: "mock", Params: llm.Params{Model: "m1"}}
	c := llm.WithProvider(p, cp.provider())
	c.Sleep = llm.NoWait
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{p}, Items: items}, map[string]*llm.Client{"solo": c})
	if err != nil {
		t.Fatal(err)
	}
	return r, cp, items
}

func block(t *testing.T, path string) func() {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
}

func savedIDs(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m, err := parseAnnotations(b)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for id := range m {
		out = append(out, id)
	}
	return out
}

// An annotation that cannot be written stops the run after its call; it is
// held as unsaved, not done; no new call is made until the next start has
// written it, and then the run carries on without asking for it again.
func TestWriteFailureKeepsResultsAndResumes(t *testing.T) {
	r, cp, items := soloRun(t, 3)
	unblock := block(t, filepath.Join(r.Dir, "annotations", "solo.json.tmp"))
	err := r.Annotate(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "annotations/solo.json") || cp.total() != 1 {
		t.Fatalf("%v; %d calls", err, cp.total())
	}
	st := r.Status()
	p := st.Progress[0]
	if st.WriteError == "" || p.Done != 0 || p.Unsaved != 1 || p.Failed != 0 || len(st.Unsaved) != 1 || st.Unsaved[0] != "annotations of solo: "+items[0] {
		t.Fatalf("status: %q %+v %q", st.WriteError, p, st.Unsaved)
	}
	if err := r.Annotate(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "cannot write its records") || cp.total() != 1 {
		t.Fatalf("a start that cannot write: %v; %d calls", err, cp.total())
	}
	unblock()
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if cp.total() != 3 || cp.of(items[0]) != 1 {
		t.Fatalf("%d calls, %d for %s", cp.total(), cp.of(items[0]), items[0])
	}
	st = r.Status()
	if st.WriteError != "" || len(st.Unsaved) != 0 || st.Progress[0].Done != 3 || st.Progress[0].Unsaved != 0 {
		t.Fatalf("after: %q %q %+v", st.WriteError, st.Unsaved, st.Progress[0])
	}
	if got := savedIDs(t, r.annotationsPath("solo")); len(got) != 3 {
		t.Fatalf("the file holds %v", got)
	}
}

// When the call log cannot be written, the call's own retries stop: the
// claim is left to do, not failed, and the line is written by the next start.
func TestCallLogFailureStopsRetries(t *testing.T) {
	r, cp, items := soloRun(t, 2)
	cp.fail = func(item string, n int) error {
		if item == items[0] && n == 1 {
			return &llm.Error{Provider: "mock", Kind: llm.KindRateLimit, Message: "slow down", RetryAfter: time.Millisecond}
		}
		return nil
	}
	unblock := block(t, filepath.Join(r.Dir, "calls.jsonl"))
	err := r.Annotate(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "calls.jsonl") || cp.total() != 1 {
		t.Fatalf("%v; %d calls", err, cp.total())
	}
	st := r.Status()
	if p := st.Progress[0]; p.Failed != 0 || p.Done != 0 || p.Calls != 1 || len(st.Unsaved) != 1 || st.Unsaved[0] != "1 line of calls.jsonl" {
		t.Fatalf("status: %+v %q", p, st.Unsaved)
	}
	unblock()
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(r.Dir, "calls.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) != 3 || !strings.Contains(lines[0], `"rate_limit"`) || !strings.Contains(lines[0], `"item":"`+items[0]+`"`) {
		t.Fatalf("%d lines; the first begins %.300s", len(lines), lines[0])
	}
}

// A reply that reached calls.jsonl when its annotation could not be written
// is taken from calls.jsonl when the run is opened again: shown as unsaved,
// written by the next start, never asked again.
func TestReopenTakesAnnotationsFromTheCallLog(t *testing.T) {
	r, cp, items := soloRun(t, 2)
	unblock := block(t, filepath.Join(r.Dir, "annotations", "solo.json.tmp"))
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Fatal("no error")
	}
	unblock()
	// The process ends; the run is opened again.
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	st := again.Status()
	if p := st.Progress[0]; p.Done != 0 || p.Unsaved != 1 || len(st.Unsaved) != 1 || st.Unsaved[0] != "annotations of solo: "+items[0]+" (taken from calls.jsonl)" {
		t.Fatalf("reopened: %+v %q", p, st.Unsaved)
	}
	c := llm.WithProvider(r.Settings.Annotators[0], cp.provider())
	c.Sleep = llm.NoWait
	if err := again.Attach("solo", c); err != nil {
		t.Fatal(err)
	}
	if err := again.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if cp.of(items[0]) != 1 || cp.total() != 2 {
		t.Fatalf("%d calls for %s, %d in all", cp.of(items[0]), items[0], cp.total())
	}
	if got := savedIDs(t, again.annotationsPath("solo")); len(got) != 2 {
		t.Fatalf("the file holds %v", got)
	}
	// What the files hold is not taken again: a third opening finds nothing to take.
	third, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil || len(third.Status().Unsaved) != 0 {
		t.Fatalf("%v %q", err, third.Status().Unsaved)
	}
}

// The same for a review, whose letters the call log records.
func TestReopenTakesReviewsFromTheCallLog(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	items := d.IDs()[1:3]
	cp := &copier{d: d}
	pa := llm.Profile{Name: "solo", Provider: "mock", Params: llm.Params{Model: "m1"}}
	pr := llm.Profile{Name: "judge", Provider: "mock", Params: llm.Params{Model: "m2"}}
	reviews := 0
	reviewer := func() *llm.Mock {
		return &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
			reviews++
			return &llm.Response{Text: `{"fields":{"scheme":{"choice":"A","why":"it fits"}},"confidence":0.8,"needs_human":false}`}, nil
		}}
	}
	clients := map[string]*llm.Client{"solo": llm.WithProvider(pa, cp.provider()), "judge": llm.WithProvider(pr, reviewer())}
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{pa}, Reviewer: &pr, Items: items}, clients)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	unblock := block(t, filepath.Join(r.Dir, "review", "model-judge.json.tmp"))
	if err := r.Review(context.Background(), nil, false); err == nil || reviews != 1 {
		t.Fatalf("%v; %d reviews", err, reviews)
	}
	if st := r.Status(); st.Progress[1].Unsaved != 1 || st.Progress[1].Done != 0 {
		t.Fatalf("reviewer: %+v", st.Progress[1])
	}
	unblock()
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return d, nil })
	if err != nil {
		t.Fatal(err)
	}
	if st := again.Status(); len(st.Unsaved) != 1 || st.Unsaved[0] != "reviews of judge: "+items[0]+" (taken from calls.jsonl)" {
		t.Fatalf("reopened: %q", st.Unsaved)
	}
	if err := again.Flush(); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(again.reviewPath("judge")); !strings.Contains(string(b), `"item": "`+items[0]+`"`) {
		t.Fatalf("Flush did not write the review: %s", b)
	}
	again.Attach("solo", llm.WithProvider(pa, cp.provider()))
	again.Attach("judge", llm.WithProvider(pr, reviewer()))
	if err := again.Review(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	if reviews != 2 {
		t.Fatalf("%d reviews asked, want 2", reviews)
	}
	fourth, err := Open(r.Dir, func(string, string) (*Dataset, error) { return d, nil })
	if err != nil {
		t.Fatal(err)
	}
	if st := fourth.Status(); len(st.Unsaved) != 0 || st.Progress[1].Done != 2 {
		t.Fatalf("saved reviews reopened: %q %+v", st.Unsaved, st.Progress[1])
	}
	b, err := os.ReadFile(again.reviewPath("judge"))
	if err != nil {
		t.Fatal(err)
	}
	var list []*Review
	if err := json.Unmarshal(b, &list); err != nil || len(list) != 2 {
		t.Fatalf("%v: %s", err, b)
	}
	if rv := list[0]; rv.Item != items[0] || rv.Letters["A"] != "solo" || rv.Fields["scheme"].Annotator != "solo" || rv.Confidence != 0.8 {
		t.Fatalf("the review taken from the log: %+v", rv)
	}
}

// Two annotators: once one cannot write, the other finishes its call under
// way and asks nothing new.
func TestOtherAnnotatorsStopAfterAWriteFailure(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	items := d.IDs()[1:4]
	a, b := &copier{d: d}, &copier{d: d}
	pa := llm.Profile{Name: "a", Provider: "mock", Params: llm.Params{Model: "m1"}}
	pb := llm.Profile{Name: "b", Provider: "mock", Params: llm.Params{Model: "m2"}}
	var r *Run
	var flushErr error
	// b's first call waits until a's annotation has failed to be written;
	// a Flush meanwhile is refused: the run is busy.
	b.before = func(string) {
		for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(time.Millisecond) {
			if r.Status().WriteError != "" {
				flushErr = r.Flush()
				return
			}
		}
		t.Error("a's write never failed")
	}
	r, err = New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{pa, pb}, Items: items},
		map[string]*llm.Client{"a": llm.WithProvider(pa, a.provider()), "b": llm.WithProvider(pb, b.provider())})
	if err != nil {
		t.Fatal(err)
	}
	block(t, filepath.Join(r.Dir, "annotations", "a.json.tmp"))
	if err := r.Annotate(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "annotations/a.json") {
		t.Fatal(err)
	}
	if a.total() != 1 || b.total() != 1 {
		t.Fatalf("calls: a %d, b %d", a.total(), b.total())
	}
	if flushErr == nil || !strings.Contains(flushErr.Error(), "annotating") {
		t.Fatalf("Flush while busy: %v", flushErr)
	}
	st := r.Status()
	if st.Progress[0].Unsaved != 1 || st.Progress[1].Done != 1 {
		t.Fatalf("progress: %+v %+v", st.Progress[0], st.Progress[1])
	}
}

// A verdict that cannot be written is not kept: the person gives it again.
func TestVerdictNotSavedIsNotKept(t *testing.T) {
	r, _, items := soloRun(t, 1)
	unblock := block(t, filepath.Join(r.Dir, "review", "human.json.tmp"))
	if err := r.SetVerdict(Verdict{Item: items[0], Choice: "solo"}); err == nil || !strings.Contains(err.Error(), "not saved") || r.Status().Verdicts != 0 {
		t.Fatalf("%v; %d verdicts", err, r.Status().Verdicts)
	}
	unblock()
	if err := r.SetVerdict(Verdict{Item: items[0], Choice: "solo"}); err != nil || r.Status().Verdicts != 1 {
		t.Fatalf("%v; %d verdicts", err, r.Status().Verdicts)
	}
	block(t, filepath.Join(r.Dir, "review", "human.json.tmp"))
	if err := r.SetVerdict(Verdict{Item: items[0], Choice: "none"}); err == nil {
		t.Fatal("no error")
	}
	if v := r.verdicts[items[0]]; v.Choice != "solo" {
		t.Fatalf("the verdict kept is %q, not the one written", v.Choice)
	}
}

// fakeLog is a call log on a disk that fails part way.
type fakeLog struct {
	data                         []byte
	writeN                       int // bytes written before the write fails (-1: none fails)
	syncErr, truncErr, closeErr  error
	truncatedTo                  int64
	truncated, closed, wroteHalf bool
}

func (f *fakeLog) Stat() (os.FileInfo, error) { return fakeInfo{int64(len(f.data))}, nil }
func (f *fakeLog) Write(b []byte) (int, error) {
	if f.writeN >= 0 && f.writeN < len(b) {
		f.data = append(f.data, b[:f.writeN]...)
		f.wroteHalf = true
		return f.writeN, errors.New("no space left on device")
	}
	f.data = append(f.data, b...)
	return len(b), nil
}
func (f *fakeLog) Truncate(n int64) error {
	f.truncated, f.truncatedTo = true, n
	if f.truncErr != nil {
		return f.truncErr
	}
	f.data = f.data[:n]
	return nil
}
func (f *fakeLog) Sync() error  { return f.syncErr }
func (f *fakeLog) Close() error { f.closed = true; return f.closeErr }

type fakeInfo struct{ size int64 }

func (i fakeInfo) Name() string       { return "calls.jsonl" }
func (i fakeInfo) Size() int64        { return i.size }
func (i fakeInfo) Mode() os.FileMode  { return 0o644 }
func (i fakeInfo) ModTime() time.Time { return time.Time{} }
func (i fakeInfo) IsDir() bool        { return false }
func (i fakeInfo) Sys() any           { return nil }

// The log never keeps half a line: a write or a sync that fails is cut off
// at the size the log had, and the caller keeps the lines to write again.
func TestAppendLinesNeverLeavesHalfALine(t *testing.T) {
	old := openAppend
	defer func() { openAppend = old }()
	lines := [][]byte{[]byte(`{"a":1}`), []byte(`{"b":2}`)}
	for _, tc := range []struct {
		name string
		log  *fakeLog
		want string // in the error; "" for none
		data string
	}{
		{"half written", &fakeLog{data: []byte("x\n"), writeN: 5}, "no space", "x\n"},
		{"sync fails", &fakeLog{data: []byte("x\n"), writeN: -1, syncErr: errors.New("I/O error")}, "I/O error", "x\n"},
		{"cannot cut off", &fakeLog{data: []byte("x\n"), writeN: 5, truncErr: errors.New("read-only")}, "could not be cut off", "x\n{\"a\":"},
		{"synced, close fails", &fakeLog{data: []byte("x\n"), writeN: -1, closeErr: errors.New("close")}, "", "x\n{\"a\":1}\n{\"b\":2}\n"},
	} {
		openAppend = func(string) (appendFile, error) { return tc.log, nil }
		err := appendLines("calls.jsonl", lines)
		if tc.want == "" && err != nil || tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
			t.Errorf("%s: %v", tc.name, err)
		}
		if string(tc.log.data) != tc.data || !tc.log.closed {
			t.Errorf("%s: the log holds %q (closed %v)", tc.name, tc.log.data, tc.log.closed)
		}
		if tc.want != "" && (!tc.log.truncated || tc.log.truncatedTo != 2) {
			t.Errorf("%s: cut off %v at %d", tc.name, tc.log.truncated, tc.log.truncatedTo)
		}
	}
	// A real file.
	openAppend = old
	path := filepath.Join(t.TempDir(), "calls.jsonl")
	for i := 0; i < 2; i++ {
		if err := appendLines(path, lines); err != nil {
			t.Fatal(err)
		}
	}
	if b, _ := os.ReadFile(path); string(b) != "{\"a\":1}\n{\"b\":2}\n{\"a\":1}\n{\"b\":2}\n" {
		t.Fatalf("%q", b)
	}
}

// writeFile removes the temporary file it made when it fails, and leaves
// alone what it did not make.
func TestWriteFileCleansUp(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "x.json")
	if err := os.Mkdir(target, 0o755); err != nil { // the rename fails
		t.Fatal(err)
	}
	if err := writeFile(target, []byte("{}")); err == nil {
		t.Fatal("no error")
	}
	if _, err := os.Stat(target + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("the temporary file is left: %v", err)
	}
	other := filepath.Join(dir, "y.json")
	if err := os.Mkdir(other+".tmp", 0o755); err != nil { // not its own
		t.Fatal(err)
	}
	if err := writeFile(other, []byte("{}")); err == nil {
		t.Fatal("no error")
	}
	if fi, err := os.Stat(other + ".tmp"); err != nil || !fi.IsDir() {
		t.Fatalf("what it did not make was removed: %v", err)
	}
}

// run.json written when the run ends: a failure there is returned too, even
// when there was nothing to ask.
func TestRunFileFailureIsReturned(t *testing.T) {
	r, cp, _ := soloRun(t, 1)
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	unblock := block(t, filepath.Join(r.Dir, "run.json.tmp"))
	err := r.Annotate(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "run.json") || cp.total() != 1 {
		t.Fatalf("%v; %d calls", err, cp.total())
	}
	if st := r.Status(); len(st.Unsaved) != 1 || st.Unsaved[0] != "run.json (progress and failures)" || st.Progress[0].Done != 1 {
		t.Fatalf("%q %+v", st.Unsaved, st.Progress[0])
	}
	unblock()
	if err := r.Flush(); err != nil {
		t.Fatal(err)
	}
	if st := r.Status(); st.WriteError != "" || len(st.Unsaved) != 0 {
		t.Fatalf("after Flush: %q %q", st.WriteError, st.Unsaved)
	}
}

// Flush writes what the files lack: an annotation, a review.
func TestFlushWritesWhatTheFilesLack(t *testing.T) {
	r, _, items := soloRun(t, 1)
	unblock := block(t, filepath.Join(r.Dir, "annotations", "solo.json.tmp"))
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Fatal("no error")
	}
	if err := r.Flush(); err == nil {
		t.Fatal("Flush wrote to a blocked file")
	}
	unblock()
	if err := r.Flush(); err != nil {
		t.Fatal(err)
	}
	if got := savedIDs(t, r.annotationsPath("solo")); len(got) != 1 || got[0] != items[0] {
		t.Fatalf("the file holds %v", got)
	}
	if st := r.Status(); st.WriteError != "" || len(st.Unsaved) != 0 || st.Progress[0].Done != 1 {
		t.Fatalf("%q %q %+v", st.WriteError, st.Unsaved, st.Progress[0])
	}
}

// reviewRun is a run with one annotator, its annotations made, and a
// reviewer whose replies come from replies (in turn; the last again).
func reviewRun(t *testing.T, n int, replies ...string) (*Run, *int) {
	t.Helper()
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	cp := &copier{d: d}
	pa := llm.Profile{Name: "solo", Provider: "mock", Params: llm.Params{Model: "m1"}}
	pr := llm.Profile{Name: "judge", Provider: "mock", Params: llm.Params{Model: "m2"}}
	calls := new(int)
	judge := llm.WithProvider(pr, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		reply := replies[min(*calls, len(replies)-1)]
		*calls++
		if reply == "rate limit" {
			return nil, &llm.Error{Provider: "mock", Kind: llm.KindRateLimit, Message: "slow down", RetryAfter: time.Millisecond}
		}
		return &llm.Response{Text: reply}, nil
	}})
	judge.Sleep = llm.NoWait
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{pa}, Reviewer: &pr, Items: d.IDs()[1 : 1+n]},
		map[string]*llm.Client{"solo": llm.WithProvider(pa, cp.provider()), "judge": judge})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err := r.Flush(); err != nil {
		t.Fatal(err)
	}
	return r, calls
}

const goodReview = `{"fields":{"scheme":{"choice":"A","why":"it fits"}},"confidence":0.7}`

// A review that failed is done again; until the new one is written, the
// file holds the failed one, and the new one is unsaved (not done). A run
// opened again takes the new one from calls.jsonl.
func TestARedoneReviewIsUnsavedUntilWritten(t *testing.T) {
	r, calls := reviewRun(t, 1, "not a review", goodReview)
	if err := r.Review(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	if st := r.Status(); st.Progress[1].Done != 1 || r.reviews[r.Settings.Items[0]].Error == "" {
		t.Fatalf("the first review: %+v", st.Progress[1])
	}
	unblock := block(t, filepath.Join(r.Dir, "review", "model-judge.json.tmp"))
	if err := r.Review(context.Background(), nil, false); err == nil || *calls != 2 {
		t.Fatalf("%v; %d calls", err, *calls)
	}
	st := r.Status()
	if p := st.Progress[1]; p.Done != 0 || p.Unsaved != 1 || len(st.Unsaved) != 1 {
		t.Fatalf("the redone review: %+v %q", p, st.Unsaved)
	}
	unblock()
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if rv := again.reviews[r.Settings.Items[0]]; rv == nil || rv.Error != "" || rv.Confidence != 0.7 {
		t.Fatalf("reopened: %+v", rv)
	}
	if st := again.Status(); len(st.Unsaved) != 1 || !strings.HasSuffix(st.Unsaved[0], "(taken from calls.jsonl)") {
		t.Fatalf("reopened: %q", st.Unsaved)
	}
}

// A reply that came with an error was not used, and is not taken from the
// log either: the claim stays failed.
func TestReopenDoesNotTakeAFailedCallsReply(t *testing.T) {
	r, _, items := soloRun(t, 1)
	// The mock answers the gold text and the error together, as a provider
	// that returns a partial response with its error.
	c := llm.WithProvider(r.Settings.Annotators[0], &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		b, _ := json.Marshal(r.Dataset.Gold[lastClaim(req)])
		return &llm.Response{Text: string(b)}, &llm.Error{Provider: "mock", Kind: llm.KindBlocked, Message: "blocked"}
	}})
	c.Sleep = llm.NoWait
	if err := r.Attach("solo", c); err != nil {
		t.Fatal(err)
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if st := r.Status(); st.Progress[0].Failed != 1 {
		t.Fatalf("%+v", st.Progress[0])
	}
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if st := again.Status(); len(st.Unsaved) != 0 || st.Progress[0].Failed != 1 || again.anns["solo"][items[0]] != nil {
		t.Fatalf("reopened: %q %+v", st.Unsaved, st.Progress[0])
	}
}

// A reply that is not an annotation is sent back once; not after a record
// could not be written.
func TestNoSecondAttemptAfterAWriteFailure(t *testing.T) {
	r, _, _ := soloRun(t, 1)
	n := 0
	c := llm.WithProvider(r.Settings.Annotators[0], &llm.Mock{Fn: func(llm.Request) (*llm.Response, error) {
		n++
		return &llm.Response{Text: "I am not sure."}, nil
	}})
	c.Sleep = llm.NoWait
	r.Attach("solo", c)
	block(t, filepath.Join(r.Dir, "calls.jsonl"))
	if err := r.Annotate(context.Background(), nil); err == nil || n != 1 {
		t.Fatalf("%v; %d calls", err, n)
	}
	if st := r.Status(); st.Progress[0].Failed != 0 {
		t.Fatalf("%+v", st.Progress[0])
	}
}

// The reviewer's retries stop once its call cannot be logged, and the claim
// stays to review (no failed review is recorded).
func TestReviewStopsRetriesWhenTheLogFails(t *testing.T) {
	r, calls := reviewRun(t, 2, "rate limit", goodReview)
	if err := os.Remove(filepath.Join(r.Dir, "calls.jsonl")); err != nil {
		t.Fatal(err)
	}
	block(t, filepath.Join(r.Dir, "calls.jsonl"))
	if err := r.Review(context.Background(), nil, false); err == nil || !strings.Contains(err.Error(), "calls.jsonl") || *calls != 1 {
		t.Fatalf("%v; %d calls", err, *calls)
	}
	if st := r.Status(); st.Reviewed != 0 {
		t.Fatalf("%d reviews recorded", st.Reviewed)
	}
}

// The error reported is the first record that could not be written.
func TestTheFirstWriteErrorIsReported(t *testing.T) {
	r, _, _ := soloRun(t, 1)
	block(t, filepath.Join(r.Dir, "annotations", "solo.json.tmp"))
	block(t, filepath.Join(r.Dir, "run.json.tmp"))
	err := r.Annotate(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "annotations/solo.json") {
		t.Fatal(err)
	}
	if st := r.Status(); len(st.Unsaved) != 2 {
		t.Fatalf("%q", st.Unsaved)
	}
}

// A line the run cannot trust is not taken: a review whose letters it does
// not record, a reply for a claim that is not the run's.
func TestReopenIgnoresLinesItCannotTrust(t *testing.T) {
	r, _ := reviewRun(t, 2, goodReview)
	items, d := r.Settings.Items, r.Dataset
	outside := d.IDs()[5]
	path := filepath.Join(r.Dir, "calls.jsonl")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var line map[string]any
	if err := json.Unmarshal([]byte(strings.SplitN(string(b), "\n", 2)[0]), &line); err != nil {
		t.Fatal(err)
	}
	gold, _ := json.Marshal(d.Gold[outside])
	call := line["call"].(map[string]any)
	call["response"].(map[string]any)["text"] = string(gold)
	line["item"] = outside
	notOurs, _ := json.Marshal(line)
	call["response"].(map[string]any)["text"] = goodReview
	line["role"], line["name"], line["item"] = "reviewer", "judge", items[1]
	noLetters, _ := json.Marshal(line)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	f.Write(append(append(append(notOurs, '\n'), noLetters...), '\n'))
	f.Close()
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return d, nil })
	if err != nil {
		t.Fatal(err)
	}
	if st := again.Status(); len(st.Unsaved) != 0 || again.reviews[items[1]] != nil || again.anns["solo"][outside] != nil {
		t.Fatalf("taken: %q", st.Unsaved)
	}
}

// A claim that failed and was then answered, with neither file written,
// is taken from the log as answered, not failed.
func TestARecoveredClaimIsNotFailed(t *testing.T) {
	r, cp, items := soloRun(t, 1)
	cp.fail = func(string, int) error {
		return &llm.Error{Provider: "mock", Kind: llm.KindBlocked, Message: "blocked"}
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if r.Status().Progress[0].Failed != 1 {
		t.Fatal("the claim did not fail")
	}
	cp.fail = nil
	unblockA := block(t, filepath.Join(r.Dir, "annotations", "solo.json.tmp"))
	unblockR := block(t, filepath.Join(r.Dir, "run.json.tmp"))
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Fatal("no error")
	}
	unblockA()
	unblockR()
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if again.failures["solo"][items[0]] != "" || again.anns["solo"][items[0]] == nil {
		t.Fatalf("failure %q", again.failures["solo"][items[0]])
	}
	if err := again.Flush(); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(r.Dir, "run.json")); strings.Contains(string(b), "blocked") {
		t.Fatalf("run.json keeps the failure: %s", b)
	}
}

// A run that cannot be continued takes nothing from the log: what it took
// could never be written.
func TestADriftedRunTakesNothingFromTheLog(t *testing.T) {
	r, _, _ := soloRun(t, 1)
	unblock := block(t, filepath.Join(r.Dir, "annotations", "solo.json.tmp"))
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Fatal("no error")
	}
	unblock()
	drifted := *r.Dataset
	drifted.Hashes = copyHashes(r.Dataset.Hashes)
	changed := ""
	for _, k := range slices.Sorted(maps.Keys(drifted.Hashes)) {
		if !slices.Contains(r.scoringInputs, k) { // a scoring-only input does not make a run drift
			changed = k
			break
		}
	}
	if changed == "" {
		t.Fatal("fixture: the dataset has no model input")
	}
	drifted.Hashes[changed] = "changed"
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return &drifted, nil })
	if err != nil {
		t.Fatal(err)
	}
	if st := again.Status(); st.InputDrift == "" || len(st.Unsaved) != 0 || st.Progress[0].Unsaved != 0 {
		t.Fatalf("drift %q, unsaved %q %+v", st.InputDrift, st.Unsaved, st.Progress[0])
	}
}

// The files win over the log: an annotation in its file is not replaced by
// the reply logged for it (a person may have corrected the file).
func TestReopenKeepsTheFilesAnnotations(t *testing.T) {
	r, _, items := soloRun(t, 2)
	unblock := block(t, filepath.Join(r.Dir, "annotations", "solo.json.tmp"))
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Fatal("no error")
	}
	unblock()
	// The first claim's annotation is in the log only; a corrected copy of
	// it is put in the file by hand, and the second claim stays missing.
	corrected := map[string]any{"kialo_claim_id": items[0], "instances": []any{}, "note": "corrected by hand"}
	b, _ := json.MarshalIndent(map[string]any{"annotator": "solo", "items": []any{corrected}}, "", " ")
	if err := os.WriteFile(r.annotationsPath("solo"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if a := again.anns["solo"][items[0]]; a == nil || a["note"] != "corrected by hand" {
		t.Fatalf("the file's annotation was replaced: %v", a)
	}
}
