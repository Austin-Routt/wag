package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func rec(id string) Record { return Record{id, "recorder", "2026-11-22T12:00:00Z"} }

// This is explicitly synthetic evidence. An authentic retained quotation of an
// instrument report is not proof that the room had the reported temperature.
func fixture(t *testing.T) (*Bundle, string) {
	t.Helper()
	root := t.TempDir()
	text := "The instrument reported 20 °C.\n"
	if err := os.WriteFile(filepath.Join(root, "report.txt"), []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256([]byte(text))
	pin := strings.Repeat("1", 64)
	quote := strings.TrimSuffix(text, "\n")
	b := &Bundle{
		Format:          Format,
		Context:         Context{ID: "synthetic-session", Profile: "evidence-fixture/1", Policy: Policy, Participants: []string{"p1", "p2"}, EngineSHA256: pin, CatalogueSHA256: pin, NASCQSHA256: pin},
		Actors:          []Actor{{"p1", "model", "first model participant"}, {"p2", "model", "second model participant"}, {"recorder", "tool", "synthetic fixture author"}},
		Sources:         []Source{{Record: rec("source"), Kind: "document", Label: "synthetic instrument report", Reference: "fixture:report"}},
		Artifacts:       []Artifact{{Record: rec("artifact-v1"), SourceID: "source", Revision: "v1", CaptureStatus: "complete", Scope: "the complete synthetic text file", Omissions: []string{}}},
		Representations: []Representation{{Record: rec("original-v1"), ArtifactID: "artifact-v1", Kind: "original", ParentIDs: []string{}, Method: "fixture creation", Payload: Payload{"report.txt", hex.EncodeToString(h[:]), int64(len([]byte(text))), "text/plain; charset=utf-8"}}},
		Claims:          []Claim{{Record: rec("reported-reading"), RepresentationID: "original-v1", AttributedTo: "source", Mode: "quotation", Text: quote, Scope: "one reported instrument reading; calibration unknown", Selector: Selector{Kind: "text", End: int64(len([]rune(quote))), Quote: quote}}},
		Targets:         []Target{{Record: rec("temperature-premise"), ArgumentID: "synthetic-argument", ArgumentRevision: pin, Kind: "premise", Proposition: "The room temperature was 20 °C."}},
		Bindings:        []Binding{{Record: rec("reading-use"), ClaimID: "reported-reading", TargetID: "temperature-premise", Role: "support", Interpretation: "Offer the reported reading for the temperature premise.", Warrant: "A calibrated instrument under suitable conditions is a defeasible indicator; those conditions remain contestable."}},
		Assessments:     []Assessment{}, Dependencies: []Dependency{}, Events: []Event{},
	}
	if r := Validate(b, root); !r.Valid {
		t.Fatalf("bad fixture: %+v", r.Errors)
	}
	return b, root
}

func consent(t *testing.T, b *Bundle, root, actor, binding, typ string) {
	t.Helper()
	r := Validate(b, root)
	if !r.Valid {
		t.Fatalf("pre-event bundle invalid: %+v", r.Errors)
	}
	e := Event{Record: rec("event-" + string(rune('a'+len(b.Events)))), Sequence: len(b.Events) + 1, Type: typ, BindingID: binding, BindingRevision: r.BindingRevisions[binding], Purpose: "consideration"}
	e.CreatedBy = actor
	b.Events = append(b.Events, e)
}

func TestAdmissionTruthTableAndWithdrawal(t *testing.T) {
	b, root := fixture(t)
	for step, actor := range []string{"", "p1", "p2", "p1", "p1"} {
		if actor != "" {
			typ := "consent"
			if step == 3 {
				typ = "withdraw"
			}
			consent(t, b, root, actor, "reading-use", typ)
		}
		r := Validate(b, root)
		if !r.Valid {
			t.Fatal(r.Errors)
		}
		want := step == 2 || step == 4
		if (len(r.AdmittedBindings) == 1) != want {
			t.Fatalf("step %d: admission %v, want %v", step, r.AdmittedBindings, want)
		}
		if !reflect.DeepEqual(r, Validate(b, root)) {
			t.Fatal("same event log gave a different report")
		}
	}
}

func TestConsentDoesNotTransferAcrossChangedBinding(t *testing.T) {
	changes := map[string]func(*Bundle){
		"target proposition":        func(b *Bundle) { b.Targets[0].Proposition = "The room temperature was 80 °C." },
		"target argument":           func(b *Bundle) { b.Targets[0].ArgumentID = "another-argument" },
		"target revision":           func(b *Bundle) { b.Targets[0].ArgumentRevision = strings.Repeat("2", 64) },
		"offered interpretation":    func(b *Bundle) { b.Bindings[0].Interpretation = "Offer it as a certain measurement." },
		"inferential warrant":       func(b *Bundle) { b.Bindings[0].Warrant = "A different inference." },
		"support versus opposition": func(b *Bundle) { b.Bindings[0].Role = "oppose" },
		"context":                   func(b *Bundle) { b.Context.ID = "another-session" },
		"profile":                   func(b *Bundle) { b.Context.Profile = "another-profile/1" },
		"catalogue":                 func(b *Bundle) { b.Context.CatalogueSHA256 = strings.Repeat("2", 64) },
		"source attribution":        func(b *Bundle) { b.Sources[0].Label = "another claimed author" },
		"artifact revision":         func(b *Bundle) { b.Artifacts[0].Revision = "v2" },
		"derivation method":         func(b *Bundle) { b.Representations[0].Method = "a different capture method" },
		"claim scope":               func(b *Bundle) { b.Claims[0].Scope = "every reading in every room" },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			b, root := fixture(t)
			consent(t, b, root, "p1", "reading-use", "consent")
			consent(t, b, root, "p2", "reading-use", "consent")
			change(b)
			r := Validate(b, root)
			if r.Valid || len(r.AdmittedBindings) != 0 {
				t.Fatalf("stale consent survived: %+v", r)
			}
			found := false
			for _, e := range r.Errors {
				found = found || e.Code == "stale_binding"
			}
			if !found {
				t.Fatalf("expected stale binding: %+v", r.Errors)
			}
		})
	}
}

func TestRecordAndPayloadFailures(t *testing.T) {
	cases := map[string]func(*Bundle, string){
		"missing target":         func(b *Bundle, _ string) { b.Bindings[0].TargetID = "absent" },
		"missing source":         func(b *Bundle, _ string) { b.Artifacts[0].SourceID = "absent" },
		"missing representation": func(b *Bundle, _ string) { b.Claims[0].RepresentationID = "absent" },
		"duplicate ID":           func(b *Bundle, _ string) { b.Bindings = append(b.Bindings, b.Bindings[0]) },
		"unrecorded author":      func(b *Bundle, _ string) { b.Claims[0].CreatedBy = "absent" },
		"complete without original": func(b *Bundle, _ string) {
			b.Representations = []Representation{}
			b.Claims = []Claim{}
			b.Bindings = []Binding{}
		},
		"complete with omissions":       func(b *Bundle, _ string) { b.Artifacts[0].Omissions = []string{"missing page"} },
		"unknown capture status":        func(b *Bundle, _ string) { b.Artifacts[0].CaptureStatus = "verified" },
		"hash mismatch":                 func(b *Bundle, _ string) { b.Representations[0].Payload.SHA256 = strings.Repeat("0", 64) },
		"length mismatch":               func(b *Bundle, _ string) { b.Representations[0].Payload.Bytes++ },
		"outside root":                  func(b *Bundle, _ string) { b.Representations[0].Payload.Path = "../report.txt" },
		"bad quote":                     func(b *Bundle, _ string) { b.Claims[0].Selector.Quote = "A different statement." },
		"bad bounds":                    func(b *Bundle, _ string) { b.Claims[0].Selector.End = 10000 },
		"byte offsets for Unicode":      func(b *Bundle, _ string) { b.Claims[0].Selector.End = int64(len(b.Claims[0].Text)) },
		"paraphrase declared quotation": func(b *Bundle, _ string) { b.Claims[0].Text = "The room was 20 °C." },
		"derivative without parent":     func(b *Bundle, _ string) { b.Representations[0].Kind = "ocr" },
		"one participant twice":         func(b *Bundle, _ string) { b.Context.Participants = []string{"p1", "p1"} },
		"media bytes changed": func(b *Bundle, root string) {
			if err := os.WriteFile(filepath.Join(root, "report.txt"), []byte("different"), 0600); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			b, root := fixture(t)
			change(b, root)
			if r := Validate(b, root); r.Valid {
				t.Fatalf("accepted %s", name)
			}
		})
	}
}

func TestUncapturedAndPartialAreHonestRecords(t *testing.T) {
	for _, status := range []string{"failed", "unsupported", "unattempted", "partial"} {
		t.Run(status, func(t *testing.T) {
			b, root := fixture(t)
			b.Representations = []Representation{}
			b.Claims = []Claim{}
			b.Targets = []Target{}
			b.Bindings = []Binding{}
			b.Artifacts[0].CaptureStatus = status
			if oneOf(status, "failed", "unsupported") {
				b.Artifacts[0].Failure = "source unavailable"
			}
			if status == "partial" {
				b.Artifacts[0].Omissions = []string{"capture stopped before any body bytes were retained"}
			}
			if r := Validate(b, root); !r.Valid {
				t.Fatalf("honest unavailable record rejected: %+v", r.Errors)
			}
		})
	}
}

func TestAssessmentCannotCreateConsent(t *testing.T) {
	b, root := fixture(t)
	b.Assessments = []Assessment{{Record: Record{"assessment", "p1", "2026-11-22T12:01:00Z"}, BindingID: "reading-use", Dimensions: map[string]string{"authenticity": "supported", "relevance": "supported", "competence": "unknown", "probative_force": "challenged"}, Rationale: "Quotation matches retained text; calibration is unknown.", BasisClaimIDs: []string{"reported-reading"}, ReviewLevel: "model_checked", FactCheck: FactCheck{Status: "not_checked", BasisClaimIDs: []string{}}}}
	r := Validate(b, root)
	if !r.Valid || len(r.AdmittedBindings) != 0 {
		t.Fatalf("assessment became admission: %+v", r)
	}
	b.Assessments[0].ReviewLevel = "human_checked"
	if r := Validate(b, root); r.Valid {
		t.Fatal("model assessment claimed human verification")
	}
}

func TestDecodeStrictnessAndRoundTrip(t *testing.T) {
	b, root := fixture(t)
	raw, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if r := Validate(got, root); !r.Valid {
		t.Fatal(r.Errors)
	}
	cases := map[string][]byte{
		"duplicate":         []byte(`{"format":"wrong","format":"wag-evidence/0.1"}`),
		"escaped duplicate": []byte(`{"format":"wrong","f\u006frmat":"wag-evidence/0.1"}`),
		"unknown":           []byte(`{"format":"wag-evidence/0.1","verified":true}`),
		"wrong case":        []byte(`{"Format":"wag-evidence/0.1"}`),
		"null":              []byte(`null`),
		"nested null":       []byte(`{"actors":null}`),
		"trailing":          append(append([]byte{}, raw...), []byte(` {}`)...),
		"invalid utf8":      []byte{'{', '"', 'x', '"', ':', '"', 0xff, '"', '}'},
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Decode(data); err == nil {
				t.Fatal("accepted malformed/ambiguous record")
			}
		})
	}
}
