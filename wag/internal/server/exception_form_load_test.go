package server

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

// A9.13 (docs/c5-item-f.md): a game a build of 6303759 recorded, in which
// sign argued the bridges' exception before any bridge, no longer replays;
// the loader names the change and the build that replays it, with this
// server's data as with other data.
func TestA913AGameRecordedBeforeTheBoundByForm(t *testing.T) {
	b, err := os.ReadFile("testdata/f1-before/game.export.json")
	if err != nil {
		t.Fatal(err)
	}
	var e Export
	if err := json.Unmarshal(b, &e); err != nil {
		t.Fatal(err)
	}
	s, _ := newTestServerWith(t)
	for _, inspect := range []bool{false, true} {
		_, err := s.load(&e, inspect)
		if !errors.Is(err, errExceptionForm) {
			t.Fatalf("inspect %v: %v, want the named change", inspect, err)
		}
		for _, want := range []string{"move 1 (Ann)", "S5: worth_finding_out_first(tax_meat,demand_falls) is the exception", "6916369 or earlier"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("inspect %v: %q lacks %q", inspect, err, want)
			}
		}
	}
	// Other data: the change is named all the same, not taken for the data's.
	e.DataHash = strings.Repeat("0", 64)
	if _, err := s.load(&e, true); !errors.Is(err, errExceptionForm) {
		t.Fatalf("with other data: %v", err)
	}
}
