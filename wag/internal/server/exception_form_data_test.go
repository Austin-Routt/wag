package server

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

// A9.13c (corrected), A9.13d and A9.13e (docs/c5-item-f.md; WK1 of the
// independent check of K): the bound by the exception's form is the data's.
// A game a build of 6916369 recorded on frozen data, sign arguing the
// exception's form before any bridge, loads on that data with this build only
// where the data has no question that refuses Other; the loader's remedy for a
// game played on other data says which.
func TestA913dTheFormBoundIsTheData(t *testing.T) {
	read := func(name string) *Export {
		b, err := os.ReadFile("testdata/f1-before-other-data/" + name)
		if err != nil {
			t.Fatal(err)
		}
		var e Export
		if err := json.Unmarshal(b, &e); err != nil {
			t.Fatal(err)
		}
		return &e
	}
	on := func(frozen string) *Server {
		d, err := ReadData("../../data/frozen/" + frozen)
		if err != nil {
			t.Fatal(err)
		}
		s, err := New(d, "")
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	own, _ := newTestServerWith(t)
	// A9.13c: on this build's data, the other data's remedy as it holds, and
	// no claim that this build cannot replay the game.
	for _, name := range []string{"092b6a6f48a672d2.export.json", "afd3335a23f12f41.export.json"} {
		for _, inspect := range []bool{false, true} {
			_, err := own.load(read(name), inspect)
			if !errors.Is(err, errExceptionForm) {
				t.Fatalf("%s, inspect %v: %v, want the named change", name, inspect, err)
			}
			msg := err.Error()
			for _, want := range []string{"if that data has no question that refuses Other, this build replays it with --data",
				"if it has one, a WAG build of commit 6916369 or earlier does"} {
				if !strings.Contains(msg, want) {
					t.Errorf("A9.13c %s: %q lacks %q", name, msg, want)
				}
			}
			if strings.Contains(msg, "This build cannot replay the game") {
				t.Errorf("A9.13c %s: %q says this build cannot replay it", name, msg)
			}
		}
	}
	// A9.13d: on the game's own data, which has the question, refused with the
	// same-data remedy.
	_, err := on("092b6a6f48a672d2").load(read("092b6a6f48a672d2.export.json"), false)
	if !errors.Is(err, errExceptionForm) || !strings.Contains(err.Error(), "This build cannot replay the game; a WAG build of commit 6916369 or earlier can") {
		t.Errorf("A9.13d: %v", err)
	}
	// A9.13e: on the game's own data, which has no such question, loaded.
	if _, err := on("afd3335a23f12f41").load(read("afd3335a23f12f41.export.json"), false); err != nil {
		t.Errorf("A9.13e: %v", err)
	}
}
