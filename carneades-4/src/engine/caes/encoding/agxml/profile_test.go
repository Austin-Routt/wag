package agxml

import (
	"strings"
	"testing"
)

const smallAGXML = `<arggraph id="g"><edu id="e1">Premise</edu><edu id="e2">Conclusion</edu><adu id="p" type="pro"/><adu id="c" type="pro"/><edge id="s1" src="e1" trg="p" type="seg"/><edge id="s2" src="e2" trg="c" type="seg"/><edge id="a" src="p" trg="c" type="exa"/></arggraph>`

func TestAGXMLSimpleConversionPreservesTopology(t *testing.T) {
	g, err := Import(strings.NewReader(smallAGXML))
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Assumptions) != 1 || g.Assumptions[0] != "p" {
		t.Fatalf("inferred conclusion was assumed: %v", g.Assumptions)
	}
	if g.Statements["p"].Text != "Premise" || g.Statements["c"].Text != "Conclusion" {
		t.Fatal("segmentation text was lost")
	}
	if len(g.Statements["c"].Args) != 1 || g.Statements["c"].Args[0] != g.Arguments["a"] {
		t.Fatal("conclusion backlink must appear exactly once")
	}
	if g.Arguments["a"].Metadata["note"] != "support by example" {
		t.Fatal("example metadata was lost")
	}
}

func TestAGXMLRebuttalPreservesIssueLinksAndText(t *testing.T) {
	g, err := Import(strings.NewReader(strings.Replace(smallAGXML, `type="exa"`, `type="reb"`, 1)))
	if err != nil {
		t.Fatal(err)
	}
	negative := g.Statements["¬c"]
	if negative == nil || negative.Text != "¬Conclusion" || negative.Issue == nil || g.Statements["c"].Issue != negative.Issue {
		t.Fatal("rebuttal text or issue backlinks were lost")
	}
	if len(negative.Args) != 1 || len(g.Assumptions) != 1 || g.Assumptions[0] != "p" {
		t.Fatal("rebuttal must have one argument backlink and neither issue position is stipulated")
	}
}

func TestAGXMLRejectsEmptyAndMalformedConversion(t *testing.T) {
	for _, bad := range []string{
		`<arggraph><edu id="e">Text without an ADU</edu></arggraph>`,
		`<wrong/>`,
		strings.Replace(smallAGXML, `src="p"`, `from="p"`, 1),
		strings.Replace(smallAGXML, `trg="c" type="exa"`, `trg="missing" type="exa"`, 1),
		strings.Replace(smallAGXML, `type="exa"`, `type="unknown"`, 1),
		strings.Replace(smallAGXML, `</arggraph>`, `<joint id="j"/></arggraph>`, 1),
	} {
		if g, err := Import(strings.NewReader(bad)); err == nil || g != nil {
			t.Fatalf("unsupported/malformed AGXML was accepted: %s", bad)
		}
	}
}
