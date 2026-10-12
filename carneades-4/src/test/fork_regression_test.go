// Regression tests for defects fixed in the kialo-research fork.
// Each test fails on upstream d19d543.
package test

import (
	"math"
	"strings"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/validation"
)

func stmt(ag *caes.ArgGraph, id string) *caes.Statement {
	s := caes.NewStatement()
	s.Id = id
	ag.Statements[id] = s
	return s
}

// A declared, fully matching scheme argument must validate without a panic
// (upstream read the last premise as assumption index -1).
func TestForkValidateSchemeArgumentNoPanic(t *testing.T) {
	ag := caes.NewArgGraph()
	p, q, c := stmt(ag, "p"), stmt(ag, "q"), stmt(ag, "c")
	sc := &caes.Scheme{Id: "s", Weight: caes.LinkedWeighingFunction, Premises: []string{"p", "q"}, Conclusions: []string{"c"}}
	a := caes.NewArgument()
	a.Id, a.Scheme, a.Conclusion = "a", sc, c
	a.Premises = []caes.Premise{{Stmt: p}, {Stmt: q}}
	ag.Arguments["a"] = a
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("validation panicked: %v", r)
		}
	}()
	if probs := validation.Validate(ag); len(probs) != 0 {
		t.Fatalf("unexpected problems: %v", probs)
	}
}

// A faulty scheme argument must be reported even when a basic-scheme argument
// exists (upstream returned at the first basic-scheme argument, in map order).
func TestForkValidateDoesNotStopAtBasicScheme(t *testing.T) {
	for trial := 0; trial < 50; trial++ {
		ag := caes.NewArgGraph()
		p, c, d := stmt(ag, "p"), stmt(ag, "c"), stmt(ag, "d")
		for i := 0; i < 20; i++ { // many basic-scheme arguments
			b := caes.NewArgument()
			b.Id, b.Scheme, b.Conclusion = "b"+string(rune('a'+i)), caes.BasicSchemes["linked"], d
			b.Premises = []caes.Premise{{Stmt: p}}
			ag.Arguments[b.Id] = b
		}
		sc := &caes.Scheme{Id: "s", Weight: caes.LinkedWeighingFunction, Premises: []string{"p", "q"}, Conclusions: []string{"c"}}
		a := caes.NewArgument()
		a.Id, a.Scheme, a.Conclusion = "bad", sc, c
		a.Premises = []caes.Premise{{Stmt: p}} // one premise missing
		ag.Arguments["bad"] = a
		found := false
		for _, pr := range validation.Validate(ag) {
			if pr.Id == "bad" {
				found = true
			}
		}
		if !found {
			t.Fatalf("trial %d: faulty argument not reported", trial)
		}
	}
}

// Arguments without a scheme (as produced by the CAF importer) must not crash validation.
func TestForkValidateNilScheme(t *testing.T) {
	ag := caes.NewArgGraph()
	p, c := stmt(ag, "p"), stmt(ag, "c")
	a := caes.NewArgument()
	a.Id, a.Scheme, a.Conclusion = "a", nil, c
	a.Premises = []caes.Premise{{Stmt: p}}
	ag.Arguments["a"] = a
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("validation panicked on a nil scheme: %v", r)
		}
	}()
	validation.Validate(ag)
}

func TestForkInconsistentReported(t *testing.T) {
	src := "statements:\n  p: P\n  np: not P\nissues:\n  i1:\n    positions: [p, np]\nassumptions: [p, np]\n"
	ag, err := yaml.Import(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	ag.GroundedLabelling() // initialises the assumption index
	if !ag.Inconsistent() {
		t.Fatal("two assumed positions of one issue were not reported as inconsistent")
	}
}

func TestForkCumulativeNoPremises(t *testing.T) {
	a := caes.NewArgument()
	w := caes.CumulativeWeighingFunction(a, caes.NewLabelling())
	if math.IsNaN(w) || w != 1.0 {
		t.Fatalf("cumulative weight of a premise-less argument = %v, want 1.0", w)
	}
}
