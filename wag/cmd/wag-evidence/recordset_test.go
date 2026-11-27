package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// R16 (docs/e1-evidence.md): wag-evidence --record-set checks a record set
// alone, with a placeholder context of two players. The example's records
// as a record set are valid; with a binding, refused.
func TestR16ARecordSetAlone(t *testing.T) {
	dir := copyExample(t)
	raw, err := os.ReadFile(filepath.Join(dir, "bundle.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	m["format"] = "wag-evidence/0.2"
	bindings := m["bindings"]
	for _, k := range []string{"context", "targets", "bindings", "assessments", "events"} {
		delete(m, k)
	}
	file := filepath.Join(dir, "records.json")
	write := func() {
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write()
	if code, out := run(t, "--record-set", file); code != 0 || !strings.Contains(out, `"valid": true`) {
		t.Fatalf("R16: the record set: %d\n%s", code, out)
	}
	m["bindings"] = bindings
	write()
	if code, out := run(t, "--record-set", file); code != 1 || !strings.Contains(out, `"bindings"`) {
		t.Fatalf("R16: a record set with a binding: %d\n%s", code, out)
	}
}
