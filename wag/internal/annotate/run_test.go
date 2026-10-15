package annotate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/llm"
)

var claimID = regexp.MustCompile(`"kialo_claim_id": "([^"]+)"`)

// lastClaim finds the claim a prompt asks about (the last one it names).
func lastClaim(req llm.Request) string {
	m := claimID.FindAllStringSubmatch(req.Messages[len(req.Messages)-1].Content, -1)
	if len(m) == 0 {
		m = claimID.FindAllStringSubmatch(req.Messages[0].Content, -1)
	}
	return m[len(m)-1][1]
}

// Two scripted annotators (one copies the gold, one changes its scheme and
// first answers with prose), a scripted reviewer, a verdict: the files, the
// report and the reopened run.
func TestRun(t *testing.T) {
	ds, _ := Builtin()
	d := ds[0]
	items := d.IDs()[1:4]
	gold := func(id string) string { b, _ := json.Marshal(d.Gold[id]); return string(b) }
	copier := &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		return &llm.Response{Text: gold(lastClaim(req))}, nil
	}}
	first := true
	changer := &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		if first {
			first = false
			return &llm.Response{Text: "Let me think about it."}, nil
		}
		var a Annotation
		_ = json.Unmarshal([]byte(gold(lastClaim(req))), &a)
		primary(a)["scheme"].(map[string]any)["walton_id"] = "other"
		b, _ := json.Marshal(a)
		return &llm.Response{Text: "```json\n" + string(b) + "\n```"}, nil
	}}
	reviewer := &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		id := lastClaim(req)
		return &llm.Response{Text: `{"kialo_claim_id": "` + id + `", "fields": {"scheme": {"choice": "B", "why": "the guidelines"}},
			"adjudicated": ` + gold(id) + `, "confidence": 0.8, "needs_human": true, "notes": "check"}`}, nil
	}}
	pa := llm.Profile{Name: "copier", Provider: "mock", Params: llm.Params{Model: "m1"}, APIKey: "secret-key-1"}
	pb := llm.Profile{Name: "changer", Provider: "mock", Params: llm.Params{Model: "m2"}}
	pr := llm.Profile{Name: "judge", Provider: "mock", Params: llm.Params{Model: "m3"}}
	dir := filepath.Join(t.TempDir(), "run1")
	r, err := New(dir, d, Settings{Annotators: []llm.Profile{pa, pb}, Reviewer: &pr, Items: items},
		map[string]*llm.Client{"copier": llm.WithProvider(pa, copier), "changer": llm.WithProvider(pb, changer), "judge": llm.WithProvider(pr, reviewer)})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	st := r.Status()
	for _, p := range st.Progress {
		if p.Role == "annotator" && (p.Done != 3 || p.Failed != 0) {
			t.Fatalf("progress %+v", p)
		}
	}
	if st.Disagreed != 3 || st.Annotated2 != 3 || st.NextItem != "" {
		t.Fatalf("status %+v", st)
	}
	if err := r.Review(context.Background(), nil, true); err != nil {
		t.Fatal(err)
	}
	if err := r.SetVerdict(Verdict{Item: items[0], Choice: "copier", Note: "the gold's reading"}); err != nil {
		t.Fatal(err)
	}
	if err := r.SetVerdict(Verdict{Item: items[0], Choice: "someone"}); err == nil {
		t.Fatal("an unknown choice")
	}
	rep, err := r.Report()
	if err != nil {
		t.Fatal(err)
	}
	if rep.Annotators[0].VsGold.Fields["scheme"].Agreement != 1 || rep.Annotators[1].VsGold.Fields["scheme"].Agreement != 0 ||
		rep.Annotators[0].Valid != 3 || len(rep.Disagreements) != 3 || rep.Reviewer.Reviewed != 3 || rep.Reviewer.NeedsHuman != 3 ||
		rep.Reviewer.VsGold.Fields["scheme"].Agreement != 1 || rep.Verdicts["copier"] != 1 {
		b, _ := json.MarshalIndent(rep, "", " ")
		t.Fatalf("report %s", b)
	}
	// B is whoever the rotation put second: the review names the annotator.
	v, _ := r.Item(items[0])
	if v.Review == nil || v.Review.Fields["scheme"].Annotator != v.Review.Letters["B"] || len(v.Disagree) == 0 || v.Verdict == nil {
		t.Fatalf("item view %+v", v)
	}
	for _, f := range []string{"run.json", "README.txt", "calls.jsonl", "report.json", "annotations/copier.json", "annotations/changer.json",
		"review/model-judge.json", "review/human.json"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("no %s", f)
		}
	}
	calls, _ := os.ReadFile(filepath.Join(dir, "calls.jsonl"))
	if n := strings.Count(string(calls), "\n"); n != 3+4+3 { // copier 3, changer 3 and a retry, reviewer 3
		t.Fatalf("%d calls", n)
	}
	all, _ := os.ReadFile(filepath.Join(dir, "run.json"))
	if strings.Contains(string(all)+string(calls), "secret-key-1") {
		t.Fatal("a key is in the run's files")
	}
	// Reopened: everything is there; continuing needs the clients again.
	r2, err := Open(dir, func(id, dir string) (*Dataset, error) { return d, nil })
	if err != nil {
		t.Fatal(err)
	}
	st2 := r2.Status()
	if st2.Reviewed != 3 || st2.Verdicts != 1 || len(st2.Missing) != 3 || st2.Progress[0].Done != 3 {
		t.Fatalf("reopened %+v", st2)
	}
	// A refused key stops the annotator after one claim.
	bad := &llm.Mock{Fn: func(llm.Request) (*llm.Response, error) {
		return nil, &llm.Error{Provider: "mock", Kind: llm.KindAuth, Message: "bad key"}
	}}
	r3, _ := New(filepath.Join(t.TempDir(), "run3"), d, Settings{Annotators: []llm.Profile{pa}, Items: items},
		map[string]*llm.Client{"copier": llm.WithProvider(pa, bad)})
	_ = r3.Annotate(context.Background(), nil)
	if p := r3.Status().Progress[0]; p.Failed != 1 || !strings.Contains(p.LastError, "bad key") {
		t.Fatalf("refused key: %+v", p)
	}
}
