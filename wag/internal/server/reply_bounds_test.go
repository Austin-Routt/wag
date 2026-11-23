package server

import (
	"encoding/json"
	"strings"
	"testing"
)

// The duplicate-member scan must not recurse deeper than encoding/json's own
// nesting limit: on Go 1.24 its Token reader has none, and a reply of two
// million '[' overflowed the goroutine stack (a fatal error, not a panic).
// From the evidence-theory review's independent check (report P).
func TestReplyDeepNestingIsRefusedCheaply(t *testing.T) {
	for _, reply := range []string{strings.Repeat("[", 1<<21), `{"a":` + strings.Repeat(`{"a":`, 1<<20)} {
		if _, err := ParseReplyFull(reply); err == nil || !strings.Contains(err.Error(), "max depth") {
			t.Fatalf("deeply nested reply: %v", err)
		}
	}
}

// A truncated final answer keeps encoding/json's message: it is the feedback
// the model gets, and recorded traces replay only if it is unchanged.
func TestReplyTruncatedFinalKeepsItsMessage(t *testing.T) {
	for _, reply := range []string{`{"action":{"type":"support"`, `{"action":{"type":`, `{"action":{"type":"supp`} {
		if _, err := ParseReplyFull("My move: " + reply); err == nil || err.Error() != "final move JSON: unexpected end of JSON input" {
			t.Errorf("%s: %v", reply, err)
		}
	}
}

// A confidence that is not a finite number is no confidence: a reply that
// held one could not be saved in a trace.
func TestReplyNonFiniteConfidenceIsDropped(t *testing.T) {
	for _, c := range []string{`"NaN"`, `"inf"`, `"-Infinity"`, `1e400`} {
		r, err := ParseReplyFull(`{"action":{"type":"abstain"},"confidence":` + c + `}`)
		if err != nil || r.Confidence != nil {
			t.Fatalf("confidence %s: %+v %v", c, r, err)
		}
		if _, err := json.Marshal(r); err != nil {
			t.Fatalf("confidence %s: the reply cannot be saved: %v", c, err)
		}
	}
	r, err := ParseReplyFull(`{"action":{"type":"abstain"},"confidence":"0.7"}`)
	if err != nil || r.Confidence == nil || *r.Confidence != 0.7 {
		t.Fatalf("a finite confidence: %+v %v", r, err)
	}
}
