package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Claims 2.2 and 2.4 say the same thing; merged, 2.4 would support the con
// of itself. The check reports the claims, imports the discussion, and says
// that the import of before 2026-10-15 refused it.
func TestKialoCheck(t *testing.T) {
	dir := t.TempDir()
	graph := `{"id": "2", "displayTitle": "Same words", "thesis_claim_id": "2.1",
	 "claims": {"2.1": {"id": "2.1", "text": "We should do it."}, "2.2": {"id": "2.2", "text": "It is cheap."},
	  "2.3": {"id": "2.3", "text": "Cheap things break."}, "2.4": {"id": "2.4", "text": "it is  cheap"}},
	 "edges": [{"location_id": "2.11", "from_claim_id": "2.1", "to_claim_id": "2.2", "relation": "support", "active": true},
	  {"location_id": "2.12", "from_claim_id": "2.2", "to_claim_id": "2.3", "relation": "attack", "active": true},
	  {"location_id": "2.13", "from_claim_id": "2.3", "to_claim_id": "2.4", "relation": "support", "active": true}]}`
	if err := os.MkdirAll(filepath.Join(dir, "2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2", "graph.json"), []byte(graph), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	say = &out
	defer func() { say = os.Stdout }()
	files, _ := loadData("")
	if code := runKialoCheck(files, dir); code != 0 {
		t.Fatalf("status %d:\n%s", code, out.String())
	}
	line := strings.Fields(strings.Split(out.String(), "\n")[1])
	if line[0] != "2" || line[3] != "2" || line[4] != "ok" || line[5] != "refused:" || !strings.Contains(out.String(), "1 refused by the import before") {
		t.Fatalf("%s", out.String())
	}
}
