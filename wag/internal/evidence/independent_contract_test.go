package evidence

// Independent source-informed contract tests. Their oracle is the record and
// consent contract, not the author's digest implementation or an LLM's judgment
// of evidential truth. In particular, valid records may contain false statements
// and contested interpretations. No runtime game/argument semantics is tested.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func independentRecord(id, actor string) Record {
	return Record{ID: id, CreatedBy: actor, CreatedAt: "2026-11-22T12:00:00Z"}
}

func independentHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func independentFixture(t *testing.T) (*Bundle, string) {
	t.Helper()
	root := t.TempDir()
	body := []byte("Report: café exposure may increase risk.\n")
	if err := os.WriteFile(filepath.Join(root, "report.txt"), body, 0600); err != nil {
		t.Fatal(err)
	}
	quote := strings.TrimSuffix(string(body), "\n")
	b := &Bundle{
		Format: Format,
		Context: Context{ID: "dialogue", Profile: "bounded-cooperative/1", Policy: Policy,
			Participants: []string{"reader-a", "reader-b"}, EngineSHA256: strings.Repeat("a", 64),
			CatalogueSHA256: strings.Repeat("b", 64), NASCQSHA256: strings.Repeat("c", 64)},
		Actors:          []Actor{{ID: "reader-a", Kind: "model", Label: "Reader A"}, {ID: "reader-b", Kind: "human", Label: "Reader B"}},
		Sources:         []Source{{Record: independentRecord("source", "reader-a"), Kind: "document", Label: "Retained report", Reference: "https://example.invalid/report"}},
		Artifacts:       []Artifact{{Record: independentRecord("artifact", "reader-a"), SourceID: "source", Revision: "capture-1", CaptureStatus: "complete", Scope: "supplied text file", Omissions: []string{}}},
		Representations: []Representation{{Record: independentRecord("original", "reader-a"), ArtifactID: "artifact", Kind: "original", ParentIDs: []string{}, Method: "supplied file", Payload: Payload{Path: "report.txt", SHA256: independentHash(body), Bytes: int64(len(body)), MediaType: "text/plain; charset=utf-8"}}},
		Claims:          []Claim{{Record: independentRecord("claim", "reader-a"), RepresentationID: "original", AttributedTo: "source", Mode: "quotation", Text: quote, Scope: "what the retained document says", Selector: Selector{Kind: "text", Start: 0, End: int64(len([]rune(quote))), Quote: quote}}},
		Targets:         []Target{{Record: independentRecord("target", "reader-a"), ArgumentID: "argument", ArgumentRevision: strings.Repeat("d", 64), Kind: "premise", Proposition: "the report says exposure may increase risk"}},
		Bindings:        []Binding{{Record: independentRecord("binding", "reader-a"), ClaimID: "claim", TargetID: "target", Role: "support", Interpretation: "The document reports a possible increase in risk.", Warrant: "The retained quotation supports this source-report premise; no truth of the reported event is asserted."}},
		Assessments:     []Assessment{{Record: independentRecord("assessment", "reader-a"), BindingID: "binding", Dimensions: map[string]string{"authenticity": "unknown", "relevance": "unknown", "probative_force": "unknown"}, Rationale: "The bytes were retained; the claim has not been fact checked.", BasisClaimIDs: []string{}, ReviewLevel: "model_checked", FactCheck: FactCheck{Status: "not_checked", BasisClaimIDs: []string{}}}},
		Dependencies:    []Dependency{}, Events: []Event{},
	}
	if r := Validate(b, root); !r.Valid {
		t.Fatalf("independent baseline invalid: %+v", r.Errors)
	}
	return b, root
}

func independentConsent(t *testing.T, b *Bundle, root, actor, binding, kind string) {
	t.Helper()
	// The validator is used only to obtain the public revision identifier. Tests
	// below compare changes and refusal behavior, never recompute its algorithm.
	r := Validate(b, root)
	if !r.Valid {
		t.Fatalf("cannot extend an invalid baseline log: %+v", r.Errors)
	}
	n := len(b.Events) + 1
	b.Events = append(b.Events, Event{Record: independentRecord("event-"+strings.Repeat("x", n), actor), Sequence: n,
		Type: kind, BindingID: binding, BindingRevision: r.BindingRevisions[binding], Purpose: "consideration"})
}

func independentHasIssue(r Report, code string) bool {
	for _, v := range r.Errors {
		if v.Code == code {
			return true
		}
	}
	return false
}

func independentHasWarning(r Report, code string) bool {
	for _, v := range r.Warnings {
		if v.Code == code {
			return true
		}
	}
	return false
}

func TestIndependentUnknownAssessmentCanBeAdmittedWithoutBecomingTruth(t *testing.T) {
	b, root := independentFixture(t)
	b.Assessments[0].Dimensions["authenticity"] = "challenged"
	b.Assessments[0].Rationale = "The retained bytes are exact, but authorship and interpretation remain disputed."
	independentConsent(t, b, root, "reader-a", "binding", "consent")
	if r := Validate(b, root); !r.Valid || len(r.AdmittedBindings) != 0 {
		t.Fatalf("one party's consent must not be bilateral admission: %+v", r)
	}
	independentConsent(t, b, root, "reader-b", "binding", "consent")
	before, _ := json.Marshal(b)
	r := Validate(b, root)
	after, _ := json.Marshal(b)
	if !r.Valid || !reflect.DeepEqual(r.AdmittedBindings, []string{"binding"}) {
		t.Fatalf("consideration consent must allow a disputed evidential item: %+v", r)
	}
	if string(before) != string(after) || b.Assessments[0].FactCheck.Status != "not_checked" || b.Assessments[0].Dimensions["authenticity"] != "challenged" {
		t.Fatal("validation/admission changed an epistemic assessment")
	}
	if !strings.Contains(r.Scope, "no factual") {
		t.Fatal("the report does not disclose its truth-certification limit")
	}
}

func TestIndependentAdmissionIsScopedToEachBinding(t *testing.T) {
	b, root := independentFixture(t)
	other := b.Bindings[0]
	other.ID = "other-binding"
	other.Role = "context"
	other.Warrant = "This separate use supplies context without supporting the premise."
	b.Bindings = append(b.Bindings, other)
	for _, binding := range []string{"binding", "other-binding"} {
		for _, actor := range []string{"reader-a", "reader-b"} {
			independentConsent(t, b, root, actor, binding, "consent")
		}
	}
	independentConsent(t, b, root, "reader-a", "binding", "withdraw")
	r := Validate(b, root)
	if !r.Valid || !reflect.DeepEqual(r.AdmittedBindings, []string{"other-binding"}) || !reflect.DeepEqual(r.Consents["binding"], []string{"reader-b"}) {
		t.Fatalf("withdrawal must affect only its actor and binding: %+v", r)
	}
	independentConsent(t, b, root, "reader-a", "binding", "consent")
	if r = Validate(b, root); !r.Valid || len(r.AdmittedBindings) != 2 {
		t.Fatalf("withdrawal may be followed by fresh consent: %+v", r)
	}
}

func TestIndependentWrongOrInvalidEventNeverLeaksPartialAdmission(t *testing.T) {
	mutations := map[string]func(*Bundle){
		"nonparticipant": func(b *Bundle) {
			b.Actors = append(b.Actors, Actor{ID: "outsider", Kind: "model", Label: "Another reader"})
			b.Events[1].CreatedBy = "outsider"
		},
		"repeated_actor_consent":  func(b *Bundle) { b.Events[1].CreatedBy = "reader-a" },
		"withdraw_before_consent": func(b *Bundle) { b.Events[0].Type = "withdraw" },
		"truth_purpose":           func(b *Bundle) { b.Events[1].Purpose = "accept_as_true" },
		"missing_prefix":          func(b *Bundle) { b.Events[0].Sequence = 2; b.Events[1].Sequence = 3 },
		"wrong_binding":           func(b *Bundle) { b.Events[1].BindingID = "absent" },
	}
	for name, change := range mutations {
		t.Run(name, func(t *testing.T) {
			b, root := independentFixture(t)
			independentConsent(t, b, root, "reader-a", "binding", "consent")
			independentConsent(t, b, root, "reader-b", "binding", "consent")
			change(b)
			r := Validate(b, root)
			if r.Valid || len(r.AdmittedBindings) != 0 || len(r.Consents) != 0 {
				t.Fatalf("invalid replay exposed usable partial consent: %+v", r)
			}
		})
	}
}

func TestIndependentOldConsentDoesNotCoverChangedUseOrProvenance(t *testing.T) {
	mutations := map[string]func(*Bundle){
		"stronger_interpretation":      func(b *Bundle) { b.Bindings[0].Interpretation = "Exposure certainly causes the outcome." },
		"different_warrant":            func(b *Bundle) { b.Bindings[0].Warrant = "Expert authority establishes the proposition." },
		"different_role":               func(b *Bundle) { b.Bindings[0].Role = "oppose" },
		"different_target_proposition": func(b *Bundle) { b.Targets[0].Proposition = "exposure certainly causes the outcome" },
		"different_argument_revision":  func(b *Bundle) { b.Targets[0].ArgumentRevision = strings.Repeat("e", 64) },
		"different_locus":              func(b *Bundle) { b.Targets[0].Kind = "exception" },
		"source_reidentified":          func(b *Bundle) { b.Sources[0].Reference = "https://example.invalid/other-author" },
		"different_attribution": func(b *Bundle) {
			b.Sources = append(b.Sources, Source{Record: independentRecord("other-source", "reader-a"), Kind: "person", Label: "Different speaker", Reference: "person:other"})
			b.Claims[0].AttributedTo = "other-source"
		},
		"capture_revision":      func(b *Bundle) { b.Artifacts[0].Revision = "capture-2" },
		"capture_scope":         func(b *Bundle) { b.Artifacts[0].Scope = "excerpt only, originally labelled full file" },
		"transformation_method": func(b *Bundle) { b.Representations[0].Method = "manual reconstruction presented as a supplied file" },
		"claim_scope":           func(b *Bundle) { b.Claims[0].Scope = "asserted about a different population" },
		"engine_pin":            func(b *Bundle) { b.Context.EngineSHA256 = strings.Repeat("f", 64) },
		"catalogue_pin":         func(b *Bundle) { b.Context.CatalogueSHA256 = strings.Repeat("f", 64) },
		"nascq_pin":             func(b *Bundle) { b.Context.NASCQSHA256 = strings.Repeat("f", 64) },
		"dialogue_id":           func(b *Bundle) { b.Context.ID = "another-dialogue" },
	}
	for name, change := range mutations {
		t.Run(name, func(t *testing.T) {
			b, root := independentFixture(t)
			independentConsent(t, b, root, "reader-a", "binding", "consent")
			independentConsent(t, b, root, "reader-b", "binding", "consent")
			change(b)
			r := Validate(b, root)
			if r.Valid || !independentHasIssue(r, "stale_binding") || len(r.AdmittedBindings) != 0 {
				t.Fatalf("old consent survived a material change: %+v", r)
			}
		})
	}
}

func TestIndependentChangedRetainedBytesRequireFreshConsent(t *testing.T) {
	b, root := independentFixture(t)
	independentConsent(t, b, root, "reader-a", "binding", "consent")
	independentConsent(t, b, root, "reader-b", "binding", "consent")
	body := []byte("Report: café exposure may increase risk.\nNew paragraph about a different study.\n")
	if err := os.WriteFile(filepath.Join(root, "report.txt"), body, 0600); err != nil {
		t.Fatal(err)
	}
	if r := Validate(b, root); r.Valid || !independentHasIssue(r, "payload") {
		t.Fatalf("changed bytes passed original payload identity: %+v", r)
	}
	b.Representations[0].Payload.SHA256 = independentHash(body)
	b.Representations[0].Payload.Bytes = int64(len(body))
	if r := Validate(b, root); r.Valid || !independentHasIssue(r, "stale_binding") {
		t.Fatalf("updating the hash silently retained old consent: %+v", r)
	}
}

func TestIndependentExactQuotesUseCodePointsAndDoNotCertifyInterpretation(t *testing.T) {
	b, root := independentFixture(t)
	b.Claims[0].Text = "café"
	b.Claims[0].Selector = Selector{Kind: "text", Start: 8, End: 12, Quote: "café"}
	b.Bindings[0].Interpretation = "This is a proposed interpretation, not a fact checked result."
	if r := Validate(b, root); !r.Valid {
		t.Fatalf("valid Unicode code-point range refused: %+v", r.Errors)
	}
	b.Claims[0].Selector.End = 13 // UTF-8 byte width wrongly treated as rune count
	if r := Validate(b, root); r.Valid || !independentHasIssue(r, "selector") {
		t.Fatalf("byte/code-point confusion accepted: %+v", r)
	}
	b.Claims[0].Selector.End = 12
	b.Claims[0].Text = "cafe\u0301"
	b.Claims[0].Selector.Quote = b.Claims[0].Text
	if r := Validate(b, root); r.Valid || !independentHasIssue(r, "selector") {
		t.Fatalf("normalization silently changed exact quote identity: %+v", r)
	}
	b.Claims[0].Mode = "paraphrase"
	b.Claims[0].Text = "A different proposed reading of the selected café passage."
	b.Claims[0].Selector.Quote = "café"
	if r := Validate(b, root); !r.Valid {
		t.Fatalf("explicit paraphrase should be a reviewable record, not certified as correct: %+v", r)
	}
}

func TestIndependentCommonSourceRecordsAreNotIndependenceCertificates(t *testing.T) {
	b, root := independentFixture(t)
	other := b.Artifacts[0]
	other.ID = "mirror-artifact"
	other.Revision = "mirror-1"
	b.Artifacts = append(b.Artifacts, other)
	rep := b.Representations[0]
	rep.ID = "mirror-original"
	rep.ArtifactID = "mirror-artifact"
	b.Representations = append(b.Representations, rep)
	b.Dependencies = append(b.Dependencies, Dependency{Record: independentRecord("dependency", "reader-a"), ArtifactIDs: []string{"artifact", "mirror-artifact"}, Relation: "common_origin", Rationale: "Both mirrors reproduce the same original report."})
	r := Validate(b, root)
	if !r.Valid || len(r.DuplicatePayloads) != 1 || !independentHasWarning(r, "shared_payload") {
		t.Fatalf("copied content lost its duplicate warning: %+v", r)
	}
	b.Dependencies[0].Relation = "independence_claim"
	b.Dependencies[0].Rationale = "An assessor claims independence, despite the duplicate bytes."
	r = Validate(b, root)
	if !r.Valid || !independentHasWarning(r, "independence_claim") || !independentHasWarning(r, "shared_payload") {
		t.Fatalf("contested independence claim became a certificate: %+v", r)
	}
	b.Dependencies[0].ArtifactIDs = []string{"artifact", "artifact"}
	if r = Validate(b, root); r.Valid || !independentHasIssue(r, "dependency") {
		t.Fatalf("one artifact was counted twice as two dependency participants: %+v", r)
	}
}

func TestIndependentUnknownFactCheckCannotMasqueradeAsDetermination(t *testing.T) {
	for _, status := range []string{"supported", "contradicted"} {
		t.Run(status, func(t *testing.T) {
			b, root := independentFixture(t)
			b.Assessments[0].FactCheck = FactCheck{Status: status, Method: "an unspecified inspection", BasisClaimIDs: []string{}}
			if r := Validate(b, root); r.Valid || !independentHasIssue(r, "fact_check") {
				t.Fatalf("determinate finding without basis accepted: %+v", r)
			}
		})
	}
	b, root := independentFixture(t)
	b.Assessments[0].FactCheck = FactCheck{Status: "inconclusive", Method: "No independent evidence was available.", BasisClaimIDs: []string{}}
	if r := Validate(b, root); !r.Valid {
		t.Fatalf("explicitly inconclusive check should remain representable: %+v", r)
	}
	b.Assessments[0].ReviewLevel = "human_checked"
	if r := Validate(b, root); r.Valid || !independentHasIssue(r, "review_level") {
		t.Fatalf("model author acquired human verification: %+v", r)
	}
}

func TestIndependentMissingSourcesAndTargetsCannotBeBound(t *testing.T) {
	for _, field := range []string{"source", "target", "representation", "claim"} {
		t.Run(field, func(t *testing.T) {
			b, root := independentFixture(t)
			switch field {
			case "source":
				b.Claims[0].AttributedTo = "missing-source"
			case "target":
				b.Bindings[0].TargetID = "missing-target"
			case "representation":
				b.Claims[0].RepresentationID = "missing-representation"
			case "claim":
				b.Bindings[0].ClaimID = "missing-claim"
			}
			if r := Validate(b, root); r.Valid || !independentHasIssue(r, "reference") {
				t.Fatalf("missing %s reference accepted: %+v", field, r)
			}
		})
	}
}

func TestIndependentIncompleteCaptureRemainsVisibleWithoutInventedPayload(t *testing.T) {
	b, root := independentFixture(t)
	for _, status := range []string{"partial", "failed", "unsupported", "unattempted"} {
		v := Artifact{Record: independentRecord("capture-"+status, "reader-a"), SourceID: "source", Revision: "attempt-1", CaptureStatus: status, Scope: "remote attachment", Omissions: []string{}}
		if status == "partial" {
			v.Omissions = []string{"All payload bytes unavailable; metadata only."}
		}
		if status == "failed" || status == "unsupported" {
			v.Failure = "No usable retained payload."
		}
		b.Artifacts = append(b.Artifacts, v)
	}
	if r := Validate(b, root); !r.Valid {
		t.Fatalf("unsuccessful captures must remain visible as records: %+v", r)
	}
}

func TestIndependentDecodePreservesOrRejectsUnicodeAndFieldIdentity(t *testing.T) {
	b, _ := independentFixture(t)
	raw, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(raw); err != nil {
		t.Fatalf("explicit baseline JSON did not decode: %v", err)
	}
	for _, replacement := range []string{`"\ud800"`, `"\udfff"`, `"prefix\ud800suffix"`} {
		t.Run(replacement, func(t *testing.T) {
			changed := strings.Replace(string(raw), `"Reader A"`, replacement, 1)
			if _, err := Decode([]byte(changed)); err == nil {
				t.Fatal("lone surrogate was silently replaced during JSON decoding")
			}
		})
	}
	for _, changed := range []string{
		strings.Replace(string(raw), `"format":`, `"Format":`, 1),
		strings.Replace(string(raw), `"format":`, `"format":"duplicate","format":`, 1),
		strings.Replace(string(raw), `"format":`, `"form\u0061t":"duplicate","format":`, 1),
	} {
		if _, err := Decode([]byte(changed)); err == nil {
			t.Fatal("noncanonical or duplicate field identity accepted")
		}
	}
}

func TestIndependentProgrammaticRegionMustHaveFiniteCoordinates(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		b, root := independentFixture(t)
		b.Representations[0].Pages = 1
		b.Claims[0].Mode = "observation"
		b.Claims[0].Text = "A proposed observation from a declared region."
		b.Claims[0].Selector = Selector{Kind: "region", Page: 1, X: value, Y: 0, Width: 0.5, Height: 0.5}
		r := Validate(b, root)
		if r.Valid {
			t.Errorf("non-finite coordinate %v accepted; fingerprint=%q", value, r.BindingRevisions["binding"])
		}
	}
}

func TestIndependentDecodeAcceptsPairedAndLiteralUnicodeEscapes(t *testing.T) {
	b, _ := independentFixture(t)
	raw, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, encoded, want string }{
		{"paired-surrogates", `"\ud83d\ude42"`, "🙂"},
		{"literal-backslash-u", `"\\ud800"`, `\ud800`},
		{"literal-replacement-character", `"\ufffd"`, "�"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := strings.Replace(string(raw), `"Reader A"`, tc.encoded, 1)
			decoded, err := Decode([]byte(input))
			if err != nil {
				t.Fatalf("valid, explicit Unicode form refused: %v", err)
			}
			if decoded.Actors[0].Label != tc.want {
				t.Fatalf("decoded label %q, want %q", decoded.Actors[0].Label, tc.want)
			}
		})
	}
}

func TestIndependentNativeRecordTextCannotNormalizeSilently(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Bundle)
	}{
		{"invalid-utf8-paraphrase", func(b *Bundle) { b.Claims[0].Mode = "paraphrase"; b.Claims[0].Text = string([]byte{'a', 0xff, 'b'}) }},
		{"nul-in-source-label", func(b *Bundle) { b.Sources[0].Label = "source\x00other" }},
		{"invalid-utf8-binding-interpretation", func(b *Bundle) { b.Bindings[0].Interpretation = string([]byte{'a', 0xc0, 0xaf, 'b'}) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, root := independentFixture(t)
			tc.change(b)
			r := Validate(b, root)
			if r.Valid || len(r.BindingRevisions) != 0 || len(r.AdmittedBindings) != 0 {
				t.Fatalf("invalid native text received a usable fingerprint or admission: %+v", r)
			}
		})
	}
}
