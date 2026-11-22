package server

import (
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/annotate"
	"github.com/yourorg/wag/internal/llm"
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

func TestStageProseBracketsAreNotMoveContainers(t *testing.T) {
	move := `{"action":{"type":"support","scheme_id":"AS:x"}}`
	for _, reply := range []string{
		move + "\nThis targets [s3].", move + "\n[1] Smith 2020.", move + "\nSee [the source](https://example.org).",
		"Options: [a) attack, b) support. " + move, `A [12" record] aside. ` + move, move + "\n- [ ] check later",
	} {
		if a, err := ParseReply(reply); err != nil || a.Type != "support" {
			t.Errorf("%q: %v", reply, err)
		}
	}
	for _, reply := range []string{move + ` []`, move + ` [{"type":"abstain"}]`, move + ` ["x"]`, move + ` [{"type":"support"`} {
		if _, err := ParseReply(reply); err == nil {
			t.Errorf("%q: a final JSON array was not refused", reply)
		}
	}
}

// A target, scheme or binding of the wrong type is refused by name, never
// dropped: a defence without its target would be played as a new argument.
func TestWrongTypedMoveFieldsAreRefused(t *testing.T) {
	for reply, field := range map[string]string{
		`{"action":{"type":"support","target_arg_id":7,"answers_cq_id":"CQ:x","scheme_id":"AS:x"}}`:      "target_arg_id",
		`{"action":{"type":"attack","target_stmt_id":["s3"],"scheme_id":"AS:x"}}`:                        "target_stmt_id",
		`{"action":{"type":"support","scheme_id":{"id":"AS:x"}}}`:                                        "scheme_id",
		`{"action":{"type":"question","target_arg_id":"a1","answers_cq_id":true}}`:                       "answers_cq_id",
		`{"action":{"type":"support","scheme_id":"AS:x","variable_bindings":["W","D"]}}`:                 "variable_bindings",
		`{"action":{"type":"support","scheme_id":"AS:x","variable_bindings":{"W":{"name":"a doctor"}}}}`: "variable_bindings",
	} {
		_, err := ParseReply(reply)
		if err == nil || !strings.Contains(err.Error(), field) {
			t.Errorf("%s: got %v, want an error naming %s", reply, err, field)
		}
	}
	// null is absent, and a number is a phrase, as before.
	a, err := ParseReply(`{"action":{"type":"support","target_arg_id":null,"scheme_id":"AS:x","variable_bindings":{"N":3,"W":"a doctor"}}}`)
	if err != nil || a.TargetArgID != "" || a.VariableBindings["N"] != "3" || a.VariableBindings["W"] != "a doctor" {
		t.Fatalf("null target or numeric binding: %+v %v", a, err)
	}
}
