package annotate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/llm"
)

// Follow-ups to the independent review of the persistence contract
// (verify/r-critical-review-2026-11-23, INTEGRATION.md). Each test fails on
// the commit it reviewed (1504bcd), several in the reviewer's own form.

// A reply with a byte that is not UTF-8 is refused live; the log keeps it
// as it came, so a reopened run refuses it too, and the claim stays failed.
func TestRefusedRepliesAreNotTakenFromTheLog(t *testing.T) {
	r, _, items := soloRun(t, 1)
	n := 0
	c := llm.WithProvider(r.Settings.Annotators[0], &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		n++
		g := map[string]any{}
		for k, v := range r.Dataset.Gold[lastClaim(req)] {
			g[k] = v
		}
		g["note"] = "MARK"
		b, _ := json.Marshal(g)
		b = bytes.Replace(b, []byte(`"MARK"`), []byte("\"caf\xe9\""), 1) // Latin-1 é: not UTF-8
		return &llm.Response{Text: string(b), Model: "m1"}, nil
	}})
	c.Sleep = llm.NoWait
	if err := r.Attach("solo", c); err != nil {
		t.Fatal(err)
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if r.Status().Progress[0].Failed != 1 || n != 2 {
		t.Fatalf("fixture: the live run refuses both replies: %+v", r.Status().Progress[0])
	}
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if a := again.anns["solo"][items[0]]; a != nil || again.failures["solo"][items[0]] == "" {
		t.Fatalf("a refused reply was taken from the log: %v; failure %q", a, again.failures["solo"][items[0]])
	}
	// The same for the reviewer: its refused review stays a failed one.
	bad := "{\"fields\":{\"scheme\":{\"choice\":\"A\",\"why\":\"caf\xe9\"}},\"confidence\":0.7}"
	rr, _ := reviewRun(t, 1, bad)
	if err := rr.Review(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	id := rr.Settings.Items[0]
	if rr.reviews[id].Error == "" {
		t.Fatal("fixture: the live run refuses the review")
	}
	reopened, err := Open(rr.Dir, func(string, string) (*Dataset, error) { return rr.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if rv := reopened.reviews[id]; rv == nil || rv.Error == "" {
		t.Fatalf("a refused review was replaced from the log: %+v", rv)
	}
	// The log holds the reply as it came.
	b, _ := os.ReadFile(filepath.Join(rr.Dir, "calls.jsonl"))
	if !strings.Contains(string(b), `"reply_base64"`) {
		t.Fatal("the reply's bytes are not in the log")
	}
}

// A reply whose bytes are fine but for one outside its JSON object was
// accepted live, and is taken from the log as it came.
func TestALoggedReplyIsReadAsItCame(t *testing.T) {
	r, _, items := soloRun(t, 1)
	c := llm.WithProvider(r.Settings.Annotators[0], &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		b, _ := json.Marshal(r.Dataset.Gold[lastClaim(req)])
		return &llm.Response{Text: "Voil\xe0: " + string(b), Model: "m1"}, nil
	}})
	c.Sleep = llm.NoWait
	r.Attach("solo", c)
	unblock := block(t, filepath.Join(r.Dir, "annotations", "solo.json.tmp"))
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Fatal("no error")
	}
	unblock()
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if again.anns["solo"][items[0]] == nil {
		t.Fatal("an accepted reply was not taken from the log")
	}
}

// Half a line at the end of the log (a crash, or a failed write that could
// not be cut off) is ended before the next line, which stays whole.
func TestAHalfLineEndsBeforeTheNext(t *testing.T) {
	r, cp, items := soloRun(t, 2)
	if err := r.Annotate(context.Background(), items[:1]); err != nil {
		t.Fatal(err)
	}
	f, _ := os.OpenFile(filepath.Join(r.Dir, "calls.jsonl"), os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString(`{"format":"wag-annotation-call/1","role":"annota`) // the crash
	f.Close()
	unblock := block(t, filepath.Join(r.Dir, "annotations", "solo.json.tmp"))
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Fatal("no error")
	}
	unblock()
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if again.anns["solo"][items[1]] == nil {
		t.Fatal("the reply after the half line was not taken")
	}
	if rep, err := again.Report(); err != nil || rep.UnreadableCallLines != 1 {
		t.Fatalf("%v; %d unreadable lines", err, rep.UnreadableCallLines)
	}
	if cp.of(items[1]) != 1 {
		t.Fatalf("%d calls", cp.of(items[1]))
	}
	// appendLines' own failure: the half line it could not cut off.
	old := openAppend
	defer func() { openAppend = old }()
	log := &fakeLog{data: []byte("x\n"), writeN: 5, truncErr: errors.New("read-only")}
	openAppend = func(string) (appendFile, error) { return log, nil }
	lines := [][]byte{[]byte(`{"a":1}`), []byte(`{"b":2}`)}
	if err := appendLines("calls.jsonl", lines); err == nil {
		t.Fatal("no error")
	}
	log.writeN, log.truncErr = -1, nil
	if err := appendLines("calls.jsonl", lines); err != nil {
		t.Fatal(err)
	}
	if string(log.data) != "x\n{\"a\":\n{\"a\":1}\n{\"b\":2}\n" {
		t.Fatalf("%q", log.data)
	}
	// A log that ends with its newline gets none added.
	log.data = []byte("x\n")
	if err := appendLines("calls.jsonl", lines); err != nil || string(log.data) != "x\n{\"a\":1}\n{\"b\":2}\n" {
		t.Fatalf("%v %q", err, log.data)
	}
}

// report.json, written as a run that could not write ends, says what no
// file holds.
func TestReportSaysWhatIsUnsaved(t *testing.T) {
	r, _, items := soloRun(t, 1)
	block(t, filepath.Join(r.Dir, "annotations", "solo.json.tmp"))
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Fatal("no error")
	}
	b, err := os.ReadFile(filepath.Join(r.Dir, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rep Report
	if err := json.Unmarshal(b, &rep); err != nil || len(rep.Unsaved) != 1 || rep.Unsaved[0] != "annotations of solo: "+items[0] {
		t.Fatalf("%v; unsaved %q", err, rep.Unsaved)
	}
}

// One spelling of a field of dataset.json, however it is spelt, is not
// ambiguous: it loads. items.json is read by its names as written, as
// validate.py reads it (case SC12, the independent check of F3, F5 and F6),
// so a lone Items or Text there is not read (TestSC12ItemsNamesAsWritten).
func TestALoneSpellingLoads(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"dataset.json {Name}": {"dataset.json": `{"Name": "My dataset"}`, "items.json": `{"items": [{"kialo_claim_id": "1.2", "text": "Meat is bad for the climate."}]}`},
	} {
		d, err := FromDir(dataset(t, files))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if len(d.Items) != 1 || d.Items[0].Text != "Meat is bad for the climate." {
			t.Errorf("%s: %+v", name, d.Items)
		}
		if name == "dataset.json {Name}" && d.Name != "My dataset" {
			t.Errorf("%s: name %q", name, d.Name)
		}
	}
}

// Annotator b waits to retry a call when a's annotation cannot be written:
// b's call stops at once, and nothing more is sent to b's provider.
func TestAWaitingRetryStopsAtOnce(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	items := d.IDs()[1:2]
	pa := llm.Profile{Name: "a", Provider: "mock", Params: llm.Params{Model: "m1"}}
	pb := llm.Profile{Name: "b", Provider: "mock", Params: llm.Params{Model: "m2"}}
	bTried := make(chan struct{})
	var bCalls atomic.Int32
	var r *Run
	ca := llm.WithProvider(pa, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		<-bTried
		for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline) && r.Status().Progress[1].Calls < 1; time.Sleep(time.Millisecond) {
		}
		g, _ := json.Marshal(d.Gold[lastClaim(req)])
		return &llm.Response{Text: string(g), Model: "m1"}, nil
	}})
	ca.Sleep = llm.NoWait
	cb := llm.WithProvider(pb, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		if bCalls.Add(1) == 1 {
			defer close(bTried)
			return nil, &llm.Error{Provider: "mock", Kind: llm.KindRateLimit, Message: "slow down", RetryAfter: time.Second}
		}
		g, _ := json.Marshal(d.Gold[lastClaim(req)])
		return &llm.Response{Text: string(g), Model: "m2"}, nil
	}})
	// b's wait before its retry ends only when its call is stopped (or after
	// 20 s, as a long Retry-After would).
	cb.Sleep = func(ctx context.Context, dur time.Duration) error {
		if dur > 0 {
			select {
			case <-ctx.Done():
			case <-time.After(20 * time.Second):
			}
		}
		return ctx.Err()
	}
	r, err = New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{pa, pb}, Items: items},
		map[string]*llm.Client{"a": ca, "b": cb})
	if err != nil {
		t.Fatal(err)
	}
	block(t, filepath.Join(r.Dir, "annotations", "a.json.tmp"))
	start := time.Now()
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Fatal("no error")
	}
	if took := time.Since(start); bCalls.Load() != 1 || took > 10*time.Second {
		t.Fatalf("b's provider got %d requests; the run stopped after %v", bCalls.Load(), took)
	}
	if st := r.Status(); st.Progress[1].Failed != 0 {
		t.Fatalf("b's claim was recorded as failed: %+v", st.Progress[1])
	}
}

// A run made from a Windows checkout before 24 November pinned every input,
// and the system prompt, with CRLF line endings: the same files with LF ones
// are the same inputs. A run made with LF files continues on the same files
// with CRLF ones. A changed file is still a change.
func TestLineEndingsDoNotDriftARun(t *testing.T) {
	files := map[string]string{
		"items.json":    "{\n  \"discussion\": {\"thesis_claim_id\": \"1.1\"},\n  \"items\": [{\"kialo_claim_id\": \"1.2\", \"text\": \"A.\"}]\n}\n",
		"guidelines.md": "# Guidelines\n\nCode each claim.\nThen check it.\n",
		"lexicon.json":  "{\n  \"atoms\": []\n}\n",
	}
	crlf := func(s string) string { return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\n", "\r\n") }
	write := func(dir string, conv func(string) string) {
		for name, body := range files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(conv(body)), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	same := func(s string) string { return s }
	newRun := func(dir string) *Run {
		d, err := FromDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		p := llm.Profile{Name: "solo", Provider: "mock", Params: llm.Params{Model: "m1"}}
		r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{p}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	open := func(r *Run, dir string) *Run {
		again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return FromDir(dir) })
		if err != nil {
			t.Fatal(err)
		}
		return again
	}
	// As a Windows build pinned it: rewrite run.json with every input's
	// CRLF hash and the CRLF prompt's.
	dir := t.TempDir()
	write(dir, same)
	r := newRun(dir)
	inputs := map[string][]byte{"walton.yml": nil, "schemes.json": data.SchemesJSON, "critical_questions.json": data.CQsJSON, "edges.json": data.EdgesJSON}
	inputs["walton.yml"], _ = data.Annotation.ReadFile("annotation/walton.yml")
	for name, body := range files {
		inputs[name] = []byte(body)
	}
	b, err := os.ReadFile(filepath.Join(r.Dir, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	var run map[string]any
	if err := json.Unmarshal(b, &run); err != nil {
		t.Fatal(err)
	}
	pinned := run["inputs_sha256"].(map[string]any)
	for name := range pinned {
		body, ok := inputs[name]
		if !ok {
			t.Fatalf("fixture: no bytes for input %s", name)
		}
		pinned[name] = sha([]byte(crlf(string(body))))
	}
	run["annotator_system_prompt_sha256"] = sha([]byte(r.Dataset.annotatorSystem(crlf)))
	b, _ = json.Marshal(run)
	if err := os.WriteFile(filepath.Join(r.Dir, "run.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	if drift := open(r, dir).InputDrift(); drift != "" {
		t.Errorf("made on Windows: %s", drift)
	}
	// Made with LF files, continued on CRLF ones.
	dir2 := t.TempDir()
	write(dir2, same)
	r2 := newRun(dir2)
	write(dir2, crlf)
	if drift := open(r2, dir2).InputDrift(); drift != "" {
		t.Errorf("made on Linux: %s", drift)
	}
	// A changed file is a change.
	if err := os.WriteFile(filepath.Join(dir2, "guidelines.md"), []byte("# Guidelines\n\nCode each claim twice.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if drift := open(r2, dir2).InputDrift(); !strings.Contains(drift, "guidelines.md") || !strings.Contains(drift, "system prompt") {
		t.Errorf("a changed file: %q", drift)
	}
}

// A call log that cannot be read through damages the run: nothing is
// taken from it, and the run is not continued (it could ask again).
func TestAnUnreadableLogDamagesTheRun(t *testing.T) {
	r, _, _ := soloRun(t, 1)
	if err := os.Mkdir(filepath.Join(r.Dir, "calls.jsonl"), 0o755); err != nil { // read: "is a directory"
		t.Fatal(err)
	}
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if !again.Damaged() || !strings.Contains(again.InputDrift(), "calls.jsonl") {
		t.Fatalf("damaged %v, drift %q", again.Damaged(), again.InputDrift())
	}
}

// slowProvider answers like a provider over a network: it gives up if the
// call's context ends while the answer is on its way.
type slowProvider struct {
	answer func(llm.Request) string
	wait   func()
	n      atomic.Int32
}

func (p *slowProvider) Name() string { return "slow" }
func (p *slowProvider) Complete(ctx context.Context, req llm.Request) (*llm.Response, error) {
	p.n.Add(1)
	p.wait()
	select {
	case <-ctx.Done():
		return nil, llm.AsError("slow", ctx.Err())
	case <-time.After(50 * time.Millisecond):
	}
	return &llm.Response{Text: p.answer(req), Model: "m2"}, nil
}

// When a record cannot be written, an attempt already with a provider is not
// cut off: its answer is received, logged and kept.
func TestAnAttemptUnderWayFinishes(t *testing.T) {
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	d := ds[0]
	items := d.IDs()[1:2]
	pa := llm.Profile{Name: "a", Provider: "mock", Params: llm.Params{Model: "m1"}}
	pb := llm.Profile{Name: "b", Provider: "slow", Params: llm.Params{Model: "m2"}}
	var r *Run
	b := &slowProvider{answer: func(req llm.Request) string { g, _ := json.Marshal(d.Gold[lastClaim(req)]); return string(g) }}
	bSent := make(chan struct{})
	b.wait = func() { // b's attempt is sent, then a's write fails
		close(bSent)
		for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline) && r.Status().WriteError == ""; time.Sleep(time.Millisecond) {
		}
	}
	cp := &copier{d: d, before: func(string) { <-bSent }}
	r, err = New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{pa, pb}, Items: items},
		map[string]*llm.Client{"a": llm.WithProvider(pa, cp.provider()), "b": llm.WithProvider(pb, b)})
	if err != nil {
		t.Fatal(err)
	}
	block(t, filepath.Join(r.Dir, "annotations", "a.json.tmp"))
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Fatal("no error")
	}
	if st := r.Status(); b.n.Load() != 1 || st.Progress[1].Done != 1 || st.Progress[1].Failed != 0 {
		t.Fatalf("b: %d requests, %+v", b.n.Load(), st.Progress[1])
	}
}

// A call whose own line cannot be logged, waiting to retry, stops at once.
func TestARetryAfterItsOwnLogFailsStopsAtOnce(t *testing.T) {
	r, cp, _ := soloRun(t, 1)
	cp.fail = func(string, int) error {
		return &llm.Error{Provider: "mock", Kind: llm.KindRateLimit, Message: "slow down", RetryAfter: time.Second}
	}
	c := llm.WithProvider(r.Settings.Annotators[0], cp.provider())
	c.Sleep = func(ctx context.Context, dur time.Duration) error {
		if dur > 0 {
			select {
			case <-ctx.Done():
			case <-time.After(20 * time.Second):
			}
		}
		return ctx.Err()
	}
	r.Attach("solo", c)
	block(t, filepath.Join(r.Dir, "calls.jsonl"))
	start := time.Now()
	if err := r.Annotate(context.Background(), nil); err == nil {
		t.Fatal("no error")
	}
	if took := time.Since(start); cp.total() != 1 || took > 10*time.Second {
		t.Fatalf("%d calls; stopped after %v", cp.total(), took)
	}
}

// A log whose end cannot be read is not written to.
func TestAppendLinesReadsTheEndFirst(t *testing.T) {
	old := openAppend
	defer func() { openAppend = old }()
	log := &fakeLog{data: []byte("x\n"), writeN: -1, readErr: errors.New("I/O error")}
	openAppend = func(string) (appendFile, error) { return log, nil }
	if err := appendLines("calls.jsonl", [][]byte{[]byte(`{"a":1}`)}); err == nil || !strings.Contains(err.Error(), "I/O error") || string(log.data) != "x\n" || !log.closed {
		t.Fatalf("%v; %q", err, log.data)
	}
}

// The refusal of two spellings names them, the field's own one as the base.
func TestTwoSpellingsAreNamed(t *testing.T) {
	for raw, want := range map[string]string{
		`{"gold": "a.json", "GOLD": ""}`:      `member "GOLD" is another spelling of "gold"`,
		`{"Gold": "", "gold": "a.json"}`:      `member "Gold" is another spelling of "gold"`,
		`{"Items": [], "ITEMS": []}`:          `member "Items" is another spelling of "ITEMS"`,
		`{"GOLD": "", "Gold": "", "gold": 1}`: `member "GOLD" is another spelling of "gold"`,
	} {
		err := canonicalNames(json.RawMessage(raw), []string{"gold", "items"})
		if err == nil || err.Error() != want {
			t.Errorf("%s: %v", raw, err)
		}
	}
}

// A call log that cannot be opened (here a link to itself) damages the run.
func TestALogThatCannotBeOpenedDamagesTheRun(t *testing.T) {
	r, _, _ := soloRun(t, 1)
	if err := os.Symlink("calls.jsonl", filepath.Join(r.Dir, "calls.jsonl")); err != nil {
		t.Skip("no symbolic links here:", err)
	}
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if !again.Damaged() || !strings.Contains(again.InputDrift(), "calls.jsonl") {
		t.Fatalf("damaged %v, drift %q", again.Damaged(), again.InputDrift())
	}
}
