package game

import (
	"os"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
)

// Every answer the shipped NASCQ offers to every critical question
// concludes what the question's locus and the answer's polarity call for
// (EdgeCases), and docs/scheme-registry/nascq-edges.md (wag-cli --edges) is
// the table of the shipped data.
func TestEveryEdgeCase(t *testing.T) {
	cases, err := EdgeCases(func() (*GameState, error) { return newShippedGame(t, data.EdgesJSON, "Pro", "Con"), nil })
	if err != nil {
		t.Fatal(err)
	}
	offered := 0
	for _, c := range cases {
		if !c.Offered {
			continue
		}
		offered++
		if c.Error != "" {
			t.Errorf("%s %s by %s (%s): %s", c.Scheme, c.CQ, c.Backer, c.Polarity, c.Error)
		}
	}
	if offered < 100 {
		t.Fatalf("only %d answers offered", offered)
	}
	doc, err := os.ReadFile("../../../docs/scheme-registry/nascq-edges.md")
	if err != nil {
		t.Skip("the docs are not here:", err)
	}
	if !strings.Contains(string(doc), EdgeTable(cases)) {
		t.Error("docs/scheme-registry/nascq-edges.md is not the table of the shipped data: run wag-cli --edges > ../docs/scheme-registry/nascq-edges.md")
	}
}
