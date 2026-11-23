package evidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests added on integration, after the independent check of the
// evidence-theory review's patch (the private workspace's
// verify/r-evidence-review-2026-11-22, report E): each covers a rule that a
// mutation could remove without any test failing, or a fix made then.

func codes(issues []Issue) map[string]bool {
	out := map[string]bool{}
	for _, i := range issues {
		out[i.Code] = true
	}
	return out
}

func writePayload(t *testing.T, root, name, text string) Payload {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256([]byte(text))
	return Payload{name, hex.EncodeToString(h[:]), int64(len(text)), "text/plain; charset=utf-8"}
}

// A bundle nested deeper than encoding/json's limit is refused before the
// recursive duplicate-key scan, which overflowed the stack on Go 1.24.
// Brackets inside strings do not count.
func TestDecodeRefusesDeepNesting(t *testing.T) {
	for _, raw := range [][]byte{bytes.Repeat([]byte("["), 5_000_000), []byte(`{"a":` + strings.Repeat(`{"a":`, 20000))} {
		if _, err := Decode(raw); err == nil || !strings.Contains(err.Error(), "deeper than") {
			t.Fatalf("deep JSON: %v", err)
		}
	}
	quoted := []byte(`{"format": "` + strings.Repeat("[", 20000) + `"}`)
	if _, err := Decode(quoted); err != nil && strings.Contains(err.Error(), "deeper than") {
		t.Fatalf("brackets inside a string counted as nesting: %v", err)
	}
	if err := nesting([]byte(`{"a": "\"[[[", "b": [[]]}`), 3); err != nil {
		t.Fatalf("an escaped quote ended the string early: %v", err)
	}
}

// Every rule refuses the case it exists for, with its own code.
func TestRuleCodes(t *testing.T) {
	cases := map[string]struct {
		code   string
		change func(*Bundle, string)
	}{
		"format":                    {"format", func(b *Bundle, _ string) { b.Format = "wag-evidence/9" }},
		"policy":                    {"context", func(b *Bundle, _ string) { b.Context.Policy = "majority" }},
		"pin":                       {"pin", func(b *Bundle, _ string) { b.Context.EngineSHA256 = "ABC" }},
		"undeclared participant":    {"participants", func(b *Bundle, _ string) { b.Context.Participants = []string{"p1", "ghost"} }},
		"a tool as participant":     {"participants", func(b *Bundle, _ string) { b.Context.Participants = []string{"p1", "recorder"} }},
		"partial without omissions": {"capture", func(b *Bundle, _ string) { b.Artifacts[0].CaptureStatus = "partial" }},
		"binding without warrant":   {"binding", func(b *Bundle, _ string) { b.Bindings[0].Warrant = " " }},
		"claim without scope":       {"claim", func(b *Bundle, _ string) { b.Claims[0].Scope = "" }},
		"page out of bounds": {"selector", func(b *Bundle, _ string) {
			b.Representations[0].Pages = 1
			b.Claims[0].Mode, b.Claims[0].Selector = "paraphrase", Selector{Kind: "page", Page: 2}
		}},
		"time out of bounds": {"selector", func(b *Bundle, _ string) {
			b.Representations[0].DurationMS = 1000
			b.Claims[0].Mode, b.Claims[0].Selector = "paraphrase", Selector{Kind: "time", StartMS: 0, EndMS: 2000}
		}},
		"region out of bounds": {"selector", func(b *Bundle, _ string) {
			b.Representations[0].Pages = 1
			b.Claims[0].Mode, b.Claims[0].Selector = "paraphrase", Selector{Kind: "region", Page: 1, X: 0.5, Width: 0.6, Height: 0.5}
		}},
		"derivation cycle": {"derivation_cycle", func(b *Bundle, root string) {
			pa, pb := writePayload(t, root, "a.txt", "a"), writePayload(t, root, "b.txt", "b")
			b.Representations = append(b.Representations,
				Representation{Record: rec("rep-a"), ArtifactID: "artifact-v1", Kind: "normalized", ParentIDs: []string{"rep-b"}, Method: "normalized", Payload: pa},
				Representation{Record: rec("rep-b"), ArtifactID: "artifact-v1", Kind: "normalized", ParentIDs: []string{"rep-a"}, Method: "normalized", Payload: pb})
		}},
		"assessment without rationale": {"assessment", func(b *Bundle, _ string) {
			b.Assessments = append(b.Assessments, Assessment{Record: rec("as1"), BindingID: "reading-use", Dimensions: map[string]string{"relevance": "supported"}, ReviewLevel: "unreviewed", FactCheck: FactCheck{Status: "not_checked"}})
		}},
		"human_checked by a model": {"review_level", func(b *Bundle, _ string) {
			a := Assessment{Record: rec("as1"), BindingID: "reading-use", Dimensions: map[string]string{"relevance": "supported"}, Rationale: "r", ReviewLevel: "human_checked", FactCheck: FactCheck{Status: "not_checked"}}
			a.CreatedBy = "p1"
			b.Assessments = append(b.Assessments, a)
		}},
		"symbolic link inside the root": {"payload", func(b *Bundle, root string) {
			if err := os.Mkdir(filepath.Join(root, "real"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(filepath.Join(root, "report.txt"), filepath.Join(root, "real", "report.txt")); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("real", filepath.Join(root, "link")); err != nil {
				t.Skip("no symbolic links here:", err)
			}
			b.Representations[0].Payload.Path = "link/report.txt"
		}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			b, root := fixture(t)
			c.change(b, root)
			r := Validate(b, root)
			if r.Valid || !codes(r.Errors)[c.code] {
				t.Fatalf("want an error %q, got valid=%v %+v", c.code, r.Valid, r.Errors)
			}
		})
	}
}

// A consent cannot be dated before the binding it concerns, and the log
// cannot go back in time.
func TestEventTimes(t *testing.T) {
	b, root := fixture(t)
	consent(t, b, root, "p1", "reading-use", "consent")
	b.Events[0].CreatedAt = "2026-11-22T11:00:00Z" // the binding is from 12:00
	if r := Validate(b, root); r.Valid || !codes(r.Errors)["event_time"] {
		t.Fatalf("a consent before its binding: %+v", r.Errors)
	}
	b, root = fixture(t)
	consent(t, b, root, "p1", "reading-use", "consent")
	consent(t, b, root, "p2", "reading-use", "consent")
	b.Events[0].CreatedAt, b.Events[1].CreatedAt = "2026-11-22T13:00:00Z", "2026-11-22T12:30:00Z"
	if r := Validate(b, root); r.Valid || !codes(r.Errors)["event_time"] {
		t.Fatalf("a log going back in time: %+v", r.Errors)
	}
	b, root = fixture(t)
	consent(t, b, root, "p1", "reading-use", "consent")
	consent(t, b, root, "p2", "reading-use", "consent")
	b.Events[0].CreatedAt, b.Events[1].CreatedAt = "2026-11-22T12:00:00Z", "2026-11-22T12:00:00Z"
	if r := Validate(b, root); !r.Valid || len(r.AdmittedBindings) != 1 {
		t.Fatalf("simultaneous consents at the binding's time: %+v", r.Errors)
	}
}

// human_checked by a declared human passes, with a warning that the check is
// declared, not attested (D045).
func TestHumanCheckedIsWarned(t *testing.T) {
	b, root := fixture(t)
	b.Actors = append(b.Actors, Actor{"h1", "human", "a declared person"})
	a := Assessment{Record: rec("as1"), BindingID: "reading-use", Dimensions: map[string]string{"relevance": "supported"}, Rationale: "r", ReviewLevel: "human_checked", FactCheck: FactCheck{Status: "not_checked"}}
	a.CreatedBy = "h1"
	b.Assessments = append(b.Assessments, a)
	if r := Validate(b, root); !r.Valid || !codes(r.Warnings)["human_checked"] {
		t.Fatalf("valid=%v warnings %+v errors %+v", r.Valid, r.Warnings, r.Errors)
	}
}

// The fingerprint covers the selected representation's ancestry and the
// participants' records, so changing either makes old consent stale.
func TestFingerprintCoversAncestryAndPlayers(t *testing.T) {
	b, root := fixture(t)
	p := writePayload(t, root, "normalized.txt", "The instrument reported 20 °C.\n")
	b.Representations = append(b.Representations, Representation{Record: rec("normalized-v1"), ArtifactID: "artifact-v1", Kind: "normalized", ParentIDs: []string{"original-v1"}, Method: "whitespace normalized", Payload: p})
	b.Claims[0].RepresentationID = "normalized-v1"
	before := Validate(b, root)
	if !before.Valid {
		t.Fatalf("fixture with a derivative: %+v", before.Errors)
	}
	for name, change := range map[string]func(*Bundle){
		"the parent's method":  func(b *Bundle) { b.Representations[0].Method = "another capture" },
		"a participant's kind": func(b *Bundle) { b.Actors[1].Kind = "human" },
		"a participant's name": func(b *Bundle) { b.Actors[0].Label = "another player" },
	} {
		c := *b
		c.Representations = append([]Representation(nil), b.Representations...)
		c.Actors = append([]Actor(nil), b.Actors...)
		change(&c)
		after := Validate(&c, root)
		if !after.Valid || after.BindingRevisions["reading-use"] == before.BindingRevisions["reading-use"] {
			t.Errorf("%s did not change the binding's revision (%v)", name, after.Errors)
		}
	}
}

// The fixture's binding revision is pinned: a change to the fingerprint (a
// field added to a record, a reordered struct) must be a deliberate version
// change, not a silent one that strands every recorded consent.
func TestGoldenBindingRevision(t *testing.T) {
	b, root := fixture(t)
	const golden = "b5bd335f97c5e0a9326dfb226a681a199dc3788082bb1550568e16f625dc918e"
	if got := Validate(b, root).BindingRevisions["reading-use"]; got != golden {
		t.Fatalf("binding revision %s, want %s (the fingerprint changed: bump wag-binding-go-json and say so)", got, golden)
	}
}

// Reports do not depend on map order.
func TestReportsAreDeterministic(t *testing.T) {
	b, root := fixture(t)
	b.Context.EngineSHA256, b.Context.CatalogueSHA256, b.Context.NASCQSHA256 = "x", "y", "z"
	b.Assessments = append(b.Assessments, Assessment{Record: rec("as1"), BindingID: "reading-use", Dimensions: map[string]string{"q": "?", "r": "?", "s": "?", "t": "?"}, Rationale: "r", ReviewLevel: "unreviewed", FactCheck: FactCheck{Status: "not_checked"}})
	first, _ := json.Marshal(Validate(b, root))
	for i := 0; i < 30; i++ {
		if again, _ := json.Marshal(Validate(b, root)); !bytes.Equal(first, again) {
			t.Fatalf("report %d differs:\n%s\n%s", i, first, again)
		}
	}
}

// Code-point offsets are found by walking the text: multi-byte characters
// count once, and an offset past the end is refused.
func TestCodePoints(t *testing.T) {
	raw := []byte("a°c€d")
	for _, c := range []struct {
		start, end int64
		want       string
		ok         bool
	}{{0, 1, "a", true}, {1, 2, "°", true}, {1, 4, "°c€", true}, {4, 5, "d", true}, {0, 5, "a°c€d", true}, {4, 6, "", false}, {6, 7, "", false}} {
		from, to, ok := codePoints(raw, c.start, c.end)
		if ok != c.ok || (ok && string(raw[from:to]) != c.want) {
			t.Errorf("[%d,%d): got %v %q", c.start, c.end, ok, raw[max(from, 0):max(to, 0)])
		}
	}
}

// A payload larger than declared is a mismatch, found without reading it to
// its end, and its selectors are not checked again.
func TestOversizePayload(t *testing.T) {
	b, root := fixture(t)
	big := bytes.Repeat([]byte("x"), 1<<20)
	if err := os.WriteFile(filepath.Join(root, "report.txt"), big, 0600); err != nil {
		t.Fatal(err)
	}
	r := Validate(b, root)
	if r.Valid || !codes(r.Errors)["payload"] || codes(r.Errors)["selector"] {
		t.Fatalf("want a payload error and no selector error: %+v", r.Errors)
	}
}
