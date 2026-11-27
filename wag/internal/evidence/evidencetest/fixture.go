// Package evidencetest holds the synthetic record set of the evidence plan's
// cases (docs/e1-evidence.md, "The fixture"), for the tests of the packages
// that play it. It is synthetic: no captured text.
package evidencetest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/yourorg/wag/internal/evidence"
)

// The fixture of docs/e1-evidence.md's cases: a synthetic record set named
// "fixture" in a temporary evidence folder. Two sources (an agency's report,
// a news article quoting it), each with an original text representation, the
// report with an excerpt derivative, a common-origin dependency between the
// two artifacts, and three claims: a quotation from the report's excerpt, a
// paraphrase in the article, an observation.

const (
	Report  = "The agency's report: taxing meat lowers emissions. Emissions fell 3% in the first year."
	Excerpt = "Taxing meat lowers emissions."
	Article = "A news article: the agency says a meat tax cuts emissions."
)

// Set is the fixture's record set, as records.json holds it: the bundle
// without the keys the game writes.
func Set() map[string]any {
	at := "2026-11-27T09:00:00Z"
	rec := func(id string) evidence.Record { return evidence.Record{ID: id, CreatedBy: "capture", CreatedAt: at} }
	payload := func(path, text string) evidence.Payload {
		sum := sha256.Sum256([]byte(text))
		return evidence.Payload{Path: path, SHA256: hex.EncodeToString(sum[:]), Bytes: int64(len(text)), MediaType: "text/plain; charset=utf-8"}
	}
	b := evidence.Bundle{
		Format: evidence.FormatGame,
		Actors: []evidence.Actor{{ID: "capture", Kind: "tool", Label: "the owner's capture tool"}, {ID: "owner", Kind: "human", Label: "the owner"}},
		Sources: []evidence.Source{
			{Record: rec("src-report"), Kind: "document", Label: "An agency's report on a meat tax", Reference: "https://example.org/agency-report"},
			{Record: rec("src-article"), Kind: "document", Label: "A news article quoting the report", Reference: "https://example.org/news-article"},
		},
		Artifacts: []evidence.Artifact{
			{Record: rec("art-report"), SourceID: "src-report", Revision: "v1", CaptureStatus: "complete", Scope: "the whole report", Omissions: []string{}},
			{Record: rec("art-article"), SourceID: "src-article", Revision: "v1", CaptureStatus: "complete", Scope: "the whole article", Omissions: []string{}},
		},
		Representations: []evidence.Representation{
			{Record: rec("rep-report"), ArtifactID: "art-report", Kind: "original", ParentIDs: []string{}, Method: "saved as text", Payload: payload("report.txt", Report)},
			{Record: rec("rep-excerpt"), ArtifactID: "art-report", Kind: "excerpt", ParentIDs: []string{"rep-report"}, Method: "cut by hand", Payload: payload("excerpt.txt", Excerpt)},
			{Record: rec("rep-article"), ArtifactID: "art-article", Kind: "original", ParentIDs: []string{}, Method: "saved as text", Payload: payload("article.txt", Article)},
		},
		Claims: []evidence.Claim{
			{Record: rec("claim-quote"), RepresentationID: "rep-excerpt", AttributedTo: "src-report", Mode: "quotation", Text: Excerpt,
				Scope: "the report's finding, as the report states it", Selector: evidence.Selector{Kind: "text", Start: 0, End: int64(len(Excerpt)), Quote: Excerpt}},
			{Record: rec("claim-paraphrase"), RepresentationID: "rep-article", AttributedTo: "src-article", Mode: "paraphrase", Text: "The agency found that a meat tax cuts emissions.",
				Scope: "the article's reading of the report", Selector: evidence.Selector{Kind: "whole"}},
			{Record: rec("claim-observation"), RepresentationID: "rep-report", AttributedTo: "src-report", Mode: "observation", Text: "Emissions fell 3% in the first year.",
				Scope: "one year, one country", Selector: evidence.Selector{Kind: "text", Start: 51, End: int64(len(Report)), Quote: "Emissions fell 3% in the first year."}},
		},
		Dependencies: []evidence.Dependency{
			{Record: rec("dep-origin"), ArtifactIDs: []string{"art-report", "art-article"}, Relation: "common_origin", Rationale: "the article quotes the report"},
		},
	}
	raw, _ := json.Marshal(b)
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	for _, k := range []string{"context", "targets", "bindings", "assessments", "events"} {
		delete(m, k)
	}
	return m
}

// Write writes the record set set as NAME/records.json of the evidence
// folder dir, with the fixture's payloads.
func Write(dir, name string, set map[string]any) error {
	root := filepath.Join(dir, name)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	for file, text := range map[string]string{"report.txt": Report, "excerpt.txt": Excerpt, "article.txt": Article} {
		if err := os.WriteFile(filepath.Join(root, file), []byte(text), 0o644); err != nil {
			return err
		}
	}
	raw, err := json.MarshalIndent(set, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "records.json"), raw, 0o644)
}
