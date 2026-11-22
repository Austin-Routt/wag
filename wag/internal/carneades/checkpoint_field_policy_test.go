package carneades

import (
	"reflect"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// This is a change-detection guard, not proof that arbitrary graph mutations
// roll back. It forces review of every new top-level field. The integration
// contract permits only the provisional mutations exercised by game moves.
func TestCheckpointGraphFieldPolicy(t *testing.T) {
	policies := map[string]string{
		"Metadata":         "read-only during a provisional WAG move",
		"Issues":           "membership and positions restored",
		"Statements":       "membership and move-mutated fields restored",
		"Arguments":        "membership and undercutter pointers restored",
		"References":       "read-only during a provisional WAG move",
		"Theory":           "read-only during a provisional WAG move",
		"Assumptions":      "copied and restored",
		"assums":           "private derived cache rebuilt by evaluation",
		"ExpectedLabeling": "read-only during a provisional WAG move",
		"Questioned":       "copied and restored",
		"Defaults":         "deep copied and restored",
	}
	typ := reflect.TypeOf(caes.ArgGraph{})
	if typ.NumField() != len(policies) {
		t.Fatalf("ArgGraph changed: got %d fields, policy has %d; review the checkpoint boundary", typ.NumField(), len(policies))
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if policies[f.Name] == "" {
			t.Errorf("no checkpoint policy for ArgGraph.%s", f.Name)
		}
	}
}
