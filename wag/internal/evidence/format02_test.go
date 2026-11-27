package evidence

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// V1 and V2 (docs/e1-evidence.md): in a wag-evidence/0.2 bundle, the bundle
// a game writes, a binding's revision (wag-binding-go-json/3) does not change
// with any record's time, and changes with each part D076 lists; a 0.1
// bundle's still changes with a time, as before.
func TestV1V2ARevisionWithoutTimes(t *testing.T) {
	revision := func(b *Bundle) string {
		t.Helper()
		r := ValidateRecords(b)
		if !r.Valid {
			t.Fatalf("invalid: %+v", r.Errors)
		}
		return r.BindingRevisions["reading-use"]
	}
	fresh := func(format string) *Bundle {
		b, _ := fixture(t)
		b.Format = format
		return b
	}
	base := revision(fresh(FormatGame))
	if old := revision(fresh(Format)); old == base {
		t.Fatal("0.1 and 0.2 give the same revision: the algorithm's name is not hashed")
	}
	later := "2026-12-01T08:00:00Z"
	for name, edit := range map[string]func(*Bundle){
		"the binding's time":        func(b *Bundle) { b.Bindings[0].CreatedAt = later },
		"the target's time":         func(b *Bundle) { b.Targets[0].CreatedAt = later },
		"the claim's time":          func(b *Bundle) { b.Claims[0].CreatedAt = later },
		"the source's time":         func(b *Bundle) { b.Sources[0].CreatedAt = later },
		"the artifact's time":       func(b *Bundle) { b.Artifacts[0].CreatedAt = later },
		"the representation's time": func(b *Bundle) { b.Representations[0].CreatedAt = later },
	} {
		b := fresh(FormatGame)
		edit(b)
		if revision(b) != base {
			t.Errorf("0.2: %s changes the revision", name)
		}
		o := fresh(Format)
		before := revision(o)
		edit(o)
		if revision(o) == before {
			t.Errorf("0.1: %s no longer changes the revision", name)
		}
	}
	for name, edit := range map[string]func(*Bundle){
		"the interpretation":     func(b *Bundle) { b.Bindings[0].Interpretation += " Read narrowly." },
		"the warrant":            func(b *Bundle) { b.Bindings[0].Warrant += " Calibration known." },
		"the role":               func(b *Bundle) { b.Bindings[0].Role = "context" },
		"the target":             func(b *Bundle) { b.Targets[0].Proposition = "The room was warm." },
		"the argument revision":  func(b *Bundle) { b.Targets[0].ArgumentRevision = b.Context.EngineSHA256[:63] + "2" },
		"the claim's scope":      func(b *Bundle) { b.Claims[0].Scope += "; one day" },
		"a participant's label":  func(b *Bundle) { b.Actors[0].Label = "another model" },
		"a pin":                  func(b *Bundle) { b.Context.NASCQSHA256 = b.Context.EngineSHA256[:63] + "3" },
		"the artifact revision":  func(b *Bundle) { b.Artifacts[0].Revision = "v2" },
		"the binding's author":   func(b *Bundle) { b.Bindings[0].CreatedBy = "p1" },
		"the context's profile":  func(b *Bundle) { b.Context.Profile = "evidence-fixture/2" },
		"the source's reference": func(b *Bundle) { b.Sources[0].Reference = "fixture:report-2" },
	} {
		b := fresh(FormatGame)
		edit(b)
		if revision(b) == base {
			t.Errorf("0.2: %s does not change the revision", name)
		}
	}
}

// V3: a records-only check reads no payload: a bundle whose payloads are
// gone has its records and its admission checked; the full check refuses
// it, naming the representation.
func TestV3RecordsOnlyReadsNoPayload(t *testing.T) {
	b, root := fixture(t)
	consent(t, b, root, "p1", "reading-use", "consent")
	consent(t, b, root, "p2", "reading-use", "consent")
	if err := os.Remove(filepath.Join(root, "report.txt")); err != nil {
		t.Fatal(err)
	}
	r := ValidateRecords(b)
	if !r.Valid || len(r.AdmittedBindings) != 1 {
		t.Fatalf("records only: valid %v, admitted %v, errors %+v", r.Valid, r.AdmittedBindings, r.Errors)
	}
	full := Validate(b, root)
	named := false
	for _, e := range full.Errors {
		named = named || (e.Code == "payload" && e.Location == "original-v1")
	}
	if full.Valid || !named {
		t.Fatalf("the full check: valid %v, errors %+v", full.Valid, full.Errors)
	}
}

// V4: on the same bundle with its payloads, the records-only check gives the
// full check's revisions, consents and admission.
func TestV4RecordsOnlyGivesTheFullChecksResult(t *testing.T) {
	for _, format := range []string{Format, FormatGame} {
		b, root := fixture(t)
		b.Format = format
		consent(t, b, root, "p1", "reading-use", "consent")
		full, records := Validate(b, root), ValidateRecords(b)
		if !full.Valid || !records.Valid || !reflect.DeepEqual(full.BindingRevisions, records.BindingRevisions) ||
			!reflect.DeepEqual(full.Consents, records.Consents) || !reflect.DeepEqual(full.AdmittedBindings, records.AdmittedBindings) {
			t.Fatalf("%s: full %+v; records only %+v", format, full, records)
		}
	}
}
