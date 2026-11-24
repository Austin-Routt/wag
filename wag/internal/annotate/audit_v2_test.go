package annotate

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditV2ReviewRejectsWrongClaimAndInvalidDecision(t *testing.T) {
	for name, raw := range map[string]string{
		"empty":                     `{}`,
		"wrong outer claim":         `{"kialo_claim_id":"wrong","fields":{"scheme":{"choice":"A"}}}`,
		"wrong adjudicated claim":   `{"fields":{"scheme":{"choice":"A"}},"adjudicated":{"kialo_claim_id":"wrong","instances":[]}}`,
		"unknown letter":            `{"fields":{"scheme":{"choice":"Z"}}}`,
		"invalid confidence":        `{"fields":{"scheme":{"choice":"A"}},"confidence":3}`,
		"duplicate field spelling":  `{"fields":{"scheme":{"choice":"Z"}},"Fields":{"scheme":{"choice":"A"}}}`,
		"duplicate choice spelling": `{"fields":{"scheme":{"choice":"A","Choice":"Z"}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			r := &Review{Item: "1.1", Letters: map[string]string{"A": "annotator"}}
			if err := readReview(r, raw); err == nil {
				t.Fatalf("invalid review accepted: fields=%v adjudicated=%v confidence=%v", r.Fields, r.Adjudicated, r.Confidence)
			}
		})
	}
	// Existing partial reviews remain legitimate: neither an ID nor every field
	// is required when the supplied decisions are meaningful and correctly bound.
	for _, raw := range []string{`{"fields":{"scheme":{"choice":"A"}}}`, `{"kialo_claim_id":"1.1","fields":{"scheme":{"choice":"none"}}}`} {
		if err := readReview(&Review{Item: "1.1", Letters: map[string]string{"A": "annotator"}}, raw); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAuditV2MalformedReviewNotRecoveredAsSuccess(t *testing.T) {
	r, _ := reviewRun(t, 1, `{"kialo_claim_id":"wrong","fields":{"scheme":{"choice":"A"}}}`)
	if err := r.Review(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	id := r.Settings.Items[0]
	if r.reviews[id].Error == "" {
		t.Error("wrong-claim review counted as success live")
	}
	reopened, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if rv := reopened.reviews[id]; rv != nil && rv.Error == "" {
		t.Error("wrong-claim review retained/recovered as successful")
	}
}

func TestAuditV2ReportWriteFailureIsReturned(t *testing.T) {
	r, _, _ := soloRun(t, 1)
	_ = block(t, filepath.Join(r.Dir, "report.json"))
	if err := r.Annotate(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "report.json") {
		t.Fatalf("report write failure not surfaced: %v", err)
	}
}
