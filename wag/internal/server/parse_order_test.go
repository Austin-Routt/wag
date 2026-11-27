package server

import (
	"strings"
	"testing"
)

// A reply with two faulty parts of one object is refused for the same part
// each time, the first by name: a replay compares the refusal's reason, and
// the next prompt quotes it. Found while writing E1.5 (docs/e1-evidence.md):
// the offer's parts, and before it a move's variable bindings and an
// action's declaration, were read in the map's order, which Go varies from
// one reading to the next. Each case is read 64 times.
func TestAFaultyReplyIsRefusedForTheSamePartEachTime(t *testing.T) {
	cases := []struct {
		name string
		read func() error
		want string
	}{
		{"two bindings that are not phrases", func() error {
			_, _, err := moveOf(map[string]interface{}{"type": "support", "scheme_id": "AS:consequences",
				"variable_bindings": map[string]interface{}{"G": map[string]interface{}{}, "A": []interface{}{}}})
			return err
		}, "the binding of A must be a phrase"},
		{"an action's id and actor not text", func() error {
			_, err := declarationsOf(map[string]interface{}{"declare": []interface{}{map[string]interface{}{"id": 5.0, "actor": 6.0}}})
			return err
		}, `an action's "actor" must be text`},
		{"an action with two parts it has not", func() error {
			_, err := declarationsOf(map[string]interface{}{"declare": []interface{}{map[string]interface{}{"zeta": "x", "beta": "y"}}})
			return err
		}, `not "beta"`},
		{"an offer's two parts not text", func() error {
			_, err := offerOf(map[string]interface{}{"offer": map[string]interface{}{"warrant": 1.0, "claim_id": 2.0}})
			return err
		}, `"claim_id" must be text`},
		{"an offer with two parts it has not", func() error {
			_, err := offerOf(map[string]interface{}{"offer": map[string]interface{}{"zeta": "x", "alpha": "y"}})
			return err
		}, `no part "alpha"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			seen := map[string]bool{}
			for i := 0; i < 64; i++ {
				err := c.read()
				if err == nil {
					t.Fatal("read")
				}
				seen[err.Error()] = true
			}
			if len(seen) != 1 {
				t.Fatalf("refused for different parts: %v", seen)
			}
			for reason := range seen {
				if !strings.Contains(reason, c.want) {
					t.Fatalf("%q does not name the first part: %s", reason, c.want)
				}
			}
		})
	}
}
