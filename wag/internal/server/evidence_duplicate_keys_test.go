package server

import (
	"strings"
	"testing"
)

// A repeated member must not overwrite a malformed citation, action, or binding
// before the normal field checks see it. The raw final object is authoritative.
func TestReplyRejectsDuplicateJSONMembers(t *testing.T) {
	cases := []struct {
		name  string
		reply string
		key   string
	}{
		{"citation overwritten", `{"action":{"type":"support","evidence":[false],"evidence":["Smith 2020"]}}`, "evidence"},
		{"source alias overwritten", `{"type":"support","sources":[123],"sources":["Smith 2020"]}`, "sources"},
		{"action overwritten", `{"action":{"type":"attack"},"action":{"type":"abstain"}}`, "action"},
		{"move type overwritten", `{"type":"attack","type":"support"}`, "type"},
		{"escaped member name", `{"type":"support","evidence":[false],"evide\u006ece":["Smith 2020"]}`, "evidence"},
		{"nested binding", `{"type":"support","variable_bindings":{"A":"first","A":"second"}}`, "A"},
		{"alternative in array", `{"action":{"type":"abstain"},"alternatives":[{"action":{"type":"resign"},"why_not":"first","why_not":"second"}]}`, "why_not"},
		{"unknown nested metadata", `{"action":{"type":"abstain"},"metadata":[{"nested":{"revision":"old","revision":"new"}}]}`, "revision"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := ParseReplyFull(tc.reply)
			if err == nil {
				t.Fatalf("ambiguous JSON accepted as %+v", r)
			}
			if !strings.Contains(err.Error(), "duplicate") || !strings.Contains(err.Error(), `"`+tc.key+`"`) {
				t.Fatalf("error must identify the duplicate member %q: %v", tc.key, err)
			}
		})
	}
}

func TestReplyDuplicateCheckIsScopedToFinalCandidate(t *testing.T) {
	// A malformed earlier example must not contaminate a valid final answer.
	for _, example := range []string{
		`{"type":"support","evidence":[false],"evidence":["overwritten"]}`,
		`{"type":"support","evidence":false,}`,
	} {
		r, err := ParseReplyFull("An example, not my move: " + example + "\nMy move: {\"action\":{\"type\":\"resign\"}}")
		if err != nil || r.Action.Type != "resign" {
			t.Fatalf("valid final reply after %s: result=%+v, err=%v", example, r, err)
		}
	}
	// Conversely, no duplicate-bearing final answer can fall back to an earlier
	// valid move, even when the duplicate has two individually valid values.
	reply := `Example: {"action":{"type":"abstain"}}
My move: {"action":{"type":"support","evidence":["old"],"evidence":["new"]}}`
	if r, err := ParseReplyFull(reply); err == nil {
		t.Fatalf("ambiguous final reply played or fell back: %+v", r)
	}
}

func TestReplyDuplicateCheckPreservesDistinctMembers(t *testing.T) {
	// Binding names are case-sensitive. Equal names in separate objects are
	// also distinct, and quoted JSON-looking content is data.
	reply := `{"action":{"type":"support","variable_bindings":{"A":"first","a":"second"},"evidence":["a quote: {\"x\":1,\"x\":2}"]},"alternatives":[{"action":{"type":"abstain"}}]}`
	r, err := ParseReplyFull(reply)
	if err != nil {
		t.Fatal(err)
	}
	if r.Action.VariableBindings["A"] != "first" || r.Action.VariableBindings["a"] != "second" {
		t.Fatalf("case-distinct bindings lost: %v", r.Action.VariableBindings)
	}
	if len(r.Action.Evidence) != 1 || r.Action.Evidence[0] != `a quote: {"x":1,"x":2}` || len(r.Alternatives) != 1 {
		t.Fatalf("valid nested content changed: %+v", r)
	}
}
