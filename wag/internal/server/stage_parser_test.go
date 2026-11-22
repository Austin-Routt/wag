package server

import (
	"github.com/yourorg/wag/internal/annotate"
	"github.com/yourorg/wag/internal/llm"
	"testing"
)

func TestStageMalformedFinalReplyNeverPlaysEarlierExample(t *testing.T) {
	prefix := `An example: {"action":{"type":"abstain"}}. Final answer: `
	for _, last := range []string{`{"action":null}`, `{"action":5}`, `{"action":{}}`, `{"action":{"type":true}}`, `{"action":{"type":"support"}`, `{"type":null}`, `{"action":{"type":"support"},}`, `[]`} {
		t.Run(last, func(t *testing.T) {
			if a, err := ParseReply(prefix + last); err == nil {
				t.Fatalf("played earlier %q despite malformed final answer %s", a.Type, last)
			}
		})
	}
}

func TestStageServerEnforcesWholeAnnotatorProfile(t *testing.T) {
	srv, ts := newTestServerWith(t)
	if err := srv.SetStore(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	p := llm.Profile{Name: "ann", Provider: "mock", Params: llm.Params{Model: "same"}}
	srv.AddModel(llm.WithProvider(p, &llm.Mock{}))
	var made struct {
		Status annotate.Status `json:"status"`
	}
	if code := call(t, ts, "POST", "/api/annotate/runs", map[string]any{"dataset": "kialo-13292-batch2", "annotators": []string{"ann"}, "count": 1}, &made); code != 200 {
		t.Fatalf("new run: %d", code)
	}
	p.System = "Changed instruction with same provider and model"
	srv.AddModel(llm.WithProvider(p, &llm.Mock{}))
	r, notes, err := srv.run(made.Status.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) == 0 || len(r.Status().Missing) == 0 {
		t.Fatalf("server silently reattached drifted profile: notes=%v missing=%v", notes, r.Status().Missing)
	}
}

func TestStageFinalReplyStillAcceptsValidEnvelope(t *testing.T) {
	reply := `An invalid example: {"action":null}. The answer is:
` + "```json\n" + `{"action":{"type":"support","scheme_id":"expert_opinion","variable_bindings":{"A":"a [source] with a \"quote\" and } brace"},"evidence":null},"summary":"Considered [s1]","alternatives":[{"action":{"type":"abstain"}}]}` + "\n```"
	r, err := ParseReplyFull(reply)
	if err != nil {
		t.Fatal(err)
	}
	if r.Action.Type != "support" || len(r.Action.Evidence) != 0 || len(r.Alternatives) != 1 {
		t.Fatalf("valid final reply changed: %+v", r)
	}
}
