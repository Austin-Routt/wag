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
	// Other data: the change is named all the same, not taken for the data's;
	// A9.13c (the check of F1, F2 and F4, W2 and N3): with the other data's
	// remedy too, since data without the question bounds nothing by its form.
	e.DataHash = strings.Repeat("0", 64)
	for _, inspect := range []bool{false, true} {
		_, err := s.load(&e, inspect)
		if !errors.Is(err, errExceptionForm) {
			t.Fatalf("with other data: %v", err)
		}
		for _, want := range []string{"6916369 or earlier", "--data", "the data of that version"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("A9.13c, inspect %v: %q lacks %q", inspect, err, want)
			}
		}
	}
}

// A9.13b (docs/c5-item-f.md): a game a build of 6916369 recorded, in which
// Other concluded the bridges' exception as a direct move before any bridge,
// is refused by S5's second layer, and the loader names the change.
func TestA913bOtherConcludingTheExceptionRecorded(t *testing.T) {
	b, err := os.ReadFile("testdata/f1-before-other/game.export.json")
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
		for _, want := range []string{"move 1 (Ann)", "S5: an argument of Other cannot conclude worth_finding_out_first(tax_meat,demand_falls)", "6916369 or earlier"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("inspect %v: %q lacks %q", inspect, err, want)
			}
		}
	}
}
