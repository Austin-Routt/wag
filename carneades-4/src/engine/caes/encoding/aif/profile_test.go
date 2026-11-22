package aif

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/validation"
)

func TestImportedAssumptionsBacklinksAndSchemeProvenance(t *testing.T) {
	for _, tc := range []struct{ kind, wantAssumption string }{{"RA", "p"}, {"CA", ""}} {
		src := `{"nodes":[{"nodeID":"p","type":"I"},{"nodeID":"c","type":"I"},{"nodeID":"r","type":"` + tc.kind + `","scheme":"expert_opinion"}],"edges":[{"edgeID":"e1","fromID":"p","toID":"r"},{"edgeID":"e2","fromID":"r","toID":"c"}]}`
		g, err := Import(strings.NewReader(src))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Join(g.Assumptions, ",") != tc.wantAssumption {
			t.Fatalf("%s: inferred or disputed nodes must not be stipulated: %v", tc.kind, g.Assumptions)
		}
		if problems := validation.Validate(g); len(problems) != 0 {
			t.Fatalf("imported topology is invalid: %v", problems)
		}
		if tc.kind == "RA" {
			a := g.Arguments["r"]
			if a.Scheme.Id != "linked" || a.Metadata["aif_scheme"] != "expert_opinion" {
				t.Fatal("external scheme name must remain provenance without implying a Walton instantiation")
			}
			g.Assumptions = nil
			if g.GroundedLabelling()[g.Statements["c"]] != caes.Out {
				t.Fatal("unaccepted premise must not become unconditional acceptance")
			}
		} else {
			for _, p := range g.Issues["r"].Positions {
				if p.Issue != g.Issues["r"] {
					t.Fatal("both conflict positions need the same Issue backlink")
				}
			}
		}
	}
}

func TestRejectLossyAIFStructures(t *testing.T) {
	base := AIF{Nodes: []Node{{Id: "p", Type: "I"}, {Id: "q", Type: "I"}, {Id: "r", Type: "RA"}}, Edges: []Edge{{Id: "e1", From: "p", To: "r"}, {Id: "e2", From: "r", To: "q"}}}
	for name, mutate := range map[string]func(*AIF){
		"dialogical node":      func(a *AIF) { a.Nodes = append(a.Nodes, Node{Id: "l", Type: "L"}) },
		"duplicate node":       func(a *AIF) { a.Nodes = append(a.Nodes, Node{Id: "p", Type: "I"}) },
		"dangling edge":        func(a *AIF) { a.Edges[0].From = "missing" },
		"duplicate edge":       func(a *AIF) { a.Edges[1].Id = "e1" },
		"direct edge":          func(a *AIF) { a.Edges[0].To = "q" },
		"multiple conclusions": func(a *AIF) { a.Edges = append(a.Edges, Edge{Id: "e3", From: "r", To: "p"}) },
		"missing conclusion":   func(a *AIF) { a.Edges = a.Edges[:1] },
		"missing premise":      func(a *AIF) { a.Edges = a.Edges[1:] },
		"self conflict":        func(a *AIF) { a.Nodes[2].Type = "CA"; a.Edges[1].To = "p" },
	} {
		t.Run(name, func(t *testing.T) {
			a := AIF{Nodes: append([]Node{}, base.Nodes...), Edges: append([]Edge{}, base.Edges...)}
			mutate(&a)
			b, _ := json.Marshal(a)
			if g, err := Import(strings.NewReader(string(b))); err == nil || g != nil {
				t.Fatalf("unsafe structure accepted: %v, %v", g, err)
			}
		})
	}
	_, err := Import(strings.NewReader(`{"nodes":[{"nodeID":"l","type":"YA"}],"edges":[]}`))
	if !errors.Is(err, caes.ErrUnsupported) {
		t.Fatalf("unsupported features need a typed refusal: %v", err)
	}
	for _, bad := range []string{
		`{}`, `{"nodes":[],"edges":[],"locutions":[{}]}`, `{"nodes":[],"edges":[]} {}`,
		`{"nodes":[{"nodeID":"p","type":"I"}],"nodes":[],"edges":[]}`,
		`{"nodes":[{"nodeID":"p","type":"I","type":"RA"}],"edges":[]}`,
		`{"nodes":[],"n\u006fdes":[],"edges":[]}`,
		`{"nodes":[{"nodeID":"p","type":"I"}],"Nodes":[],"edges":[]}`,
		`{"nodes":[{"nodeID":"p","NODEID":"q","type":"I"}],"edges":[]}`,
		`{"nodes":[{"nodeID":"p","type":"I"}],"nodeſ":[],"edges":[]}`,
	} {
		if g, err := Import(strings.NewReader(bad)); err == nil || g != nil {
			t.Fatalf("invalid JSON profile accepted: %s", bad)
		}
	}
}
