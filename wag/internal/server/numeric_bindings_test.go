package server

import (
	"strings"
	"testing"
)

// Numeric bindings are phrases, not a request to perform floating-point
// arithmetic. Preserve their JSON spelling, including values beyond float64.
func TestReplyNumericBindingsKeepExactLexeme(t *testing.T) {
	for _, lexeme := range []string{"3", "9007199254740993", "0.12345678901234567890123456789", "1.2300e+12", "1e400", "-0"} {
		t.Run(lexeme, func(t *testing.T) {
			reply := `{"action":{"type":"support","scheme_id":"AS:x","target_arg_id":null,"variable_bindings":{"N":` + lexeme + `,"O":null,"S":"9007199254740993"}},"confidence":0.75,"alternatives":[{"action":{"type":"support","scheme_id":"AS:x","variable_bindings":{"N":` + lexeme + `}}}]}`
			r, err := ParseReplyFull(reply)
			if err != nil {
				t.Fatal(err)
			}
			if got := r.Action.VariableBindings["N"]; got != lexeme {
				t.Errorf("numeric binding changed: got %q want %q", got, lexeme)
			}
			if r.Action.TargetArgID != "" || r.Action.VariableBindings["S"] != "9007199254740993" {
				t.Errorf("null target or saved string binding changed: %+v", r.Action)
			}
			if _, ok := r.Action.VariableBindings["O"]; ok {
				t.Error("null binding was not absent")
			}
			if r.Confidence == nil || *r.Confidence != 0.75 {
				t.Errorf("numeric confidence compatibility was lost: %v", r.Confidence)
			}
			if len(r.Alternatives) != 1 || r.Alternatives[0].Action.VariableBindings["N"] != lexeme {
				t.Errorf("alternative lost exact numeric binding: %+v", r.Alternatives)
			}
			bare, err := ParseReply(`{"type":"support","scheme_id":"AS:x","bindings":{"N":` + lexeme + `}}`)
			if err != nil || bare.VariableBindings["N"] != lexeme {
				t.Errorf("bare move lost exact numeric binding: %+v %v", bare, err)
			}
		})
	}
	for _, field := range []string{"target_arg_id", "scheme_id", "evidence"} {
		_, err := ParseReply(`{"action":{"type":"support","` + field + `":9007199254740993}}`)
		if err == nil || !strings.Contains(err.Error(), field) || !strings.Contains(err.Error(), "a number") {
			t.Errorf("numeric %s did not retain its named type error: %v", field, err)
		}
	}
}
