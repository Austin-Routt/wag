package lkif

import (
	"github.com/carneades/carneades-4/src/engine/caes"
	"strings"
	"testing"
)

func TestLKIFRejectsLossAndDanglingReferences(t *testing.T) {
	graph := `<argument-graph><statements><statement id="p"><s>Premise</s></statement><statement id="c"><s>Conclusion</s></statement></statements><arguments><argument id="a"><conclusion statement="c"/><premises><premise statement="p" polarity="positive" type="ordinary"/></premises></argument></arguments></argument-graph>`
	wrap := func(s string) string { return `<lkif><argument-graphs>` + s + `</argument-graphs></lkif>` }
	if _, err := Import(strings.NewReader(wrap(graph))); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{
		wrap(graph + graph), wrap(strings.Replace(graph, `statement="p"`, `statement="missing"`, 1)),
		wrap(strings.Replace(graph, `polarity="positive"`, `polarity="garbled"`, 1)),
		graph,
	} {
		if g, err := Import(strings.NewReader(bad)); err == nil || g != nil {
			t.Fatalf("unsafe LKIF accepted: %s", bad)
		}
	}
}

func TestLKIFExceptionHasConclusionBacklinkAndDefeatsTarget(t *testing.T) {
	src := `<lkif><argument-graphs><argument-graph><statements><statement id="p" value="true"><s>P</s></statement><statement id="e" value="true"><s>Exception</s></statement><statement id="c"><s>C</s></statement></statements><arguments><argument id="a"><conclusion statement="c"/><premises><premise statement="p" polarity="positive" type="ordinary"/><premise statement="e" polarity="positive" type="exception"/></premises></argument></arguments></argument-graph></argument-graphs></lkif>`
	g, err := Import(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	u := g.Arguments["a"].Undercutter
	if u == nil || len(u.Args) != 1 || u.Args[0] != g.Arguments["a.1"] {
		t.Fatal("exception conclusion backlink is missing")
	}
	if g.GroundedLabelling()[g.Statements["c"]] != caes.Out {
		t.Fatal("proved exception must defeat the target argument")
	}
	g.Assumptions = []string{"p"}
	if g.GroundedLabelling()[g.Statements["c"]] != caes.In {
		t.Fatal("unproved exception must not defeat the target argument")
	}
}
