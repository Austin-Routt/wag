package server

import (
	"testing"

	"github.com/yourorg/wag/internal/evidence/evidencetest"
)

// The fixture of docs/e1-evidence.md's cases (evidencetest): a synthetic
// record set named "fixture" in a temporary evidence folder.

const (
	fixtureReport  = evidencetest.Report
	fixtureExcerpt = evidencetest.Excerpt
	fixtureArticle = evidencetest.Article
)

// fixtureSet is the fixture's record set, as records.json holds it.
func fixtureSet() map[string]any { return evidencetest.Set() }

// writeFixture writes the record set set as NAME/records.json of the
// evidence folder dir, with the fixture's payloads, and returns dir.
func writeFixture(t *testing.T, dir, name string, set map[string]any) string {
	t.Helper()
	if err := evidencetest.Write(dir, name, set); err != nil {
		t.Fatal(err)
	}
	return dir
}
