package caes_test

import (
	"fmt"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
)

// Public stipulation and scheme defaults have different origins even when
// their formulas are identical. This checks both imported and raw API state.
func TestExplicitStipulationPromotesQuestionedDefault(t *testing.T) {
	for _, before := range []bool{false, true} {
		for _, defaultID := range []string{"p", " p ", "¬¬p"} {
			for _, addedID := range []string{"p", " p ", "¬¬p"} {
				t.Run(fmt.Sprintf("evaluated=%v/default=%q/added=%q", before, defaultID, addedID), func(t *testing.T) {
					ag := readNegationGraph(t, "statements:\n  p: P\n  other: Other\ndefaults: [p, other]\nquestioned: [p, other]\n")
					delete(ag.Defaults, "p")
					ag.Defaults[defaultID] = true
					if before {
						if err := ag.Infer(); err != nil {
							t.Fatal(err)
						}
						if ag.GroundedLabelling()[ag.Statements["p"]] == caes.In {
							t.Fatal("unanswered default already accepted")
						}
					}
					p := ag.Statements["p"]
					ag.AddAssumption(addedID)
					if len(ag.Defaults) != 1 || !ag.Defaults["other"] {
						t.Fatalf("promotion changed unrelated origin or retained an alias: %v", ag.Defaults)
					}
					if err := ag.Infer(); err != nil {
						t.Fatal(err)
					}
					labels := ag.GroundedLabelling()
					if labels[ag.Statements["p"]] != caes.In || labels[ag.Statements["other"]] == caes.In {
						t.Fatalf("stipulation/default labels: p=%v other=%v", labels[ag.Statements["p"]], labels[ag.Statements["other"]])
					}
					if ag.Statements["p"] != p || !ag.IsQuestioned("p") {
						t.Fatal("promotion must retain the statement object and question history")
					}
				})
			}
		}
	}
}

func TestGeneratedDefaultKeepsOriginUntilStipulated(t *testing.T) {
	ag := readNegationGraph(t, `language:
  p/0: P
  d/0: D
  q/0: Q
argument_schemes:
  - id: conditional
    premises: [p]
    assumptions: [d]
    conclusions: [q]
assumptions: [p]
questioned: [d]
`)
	if err := ag.Infer(); err != nil {
		t.Fatal(err)
	}
	if !ag.Defaults["d"] || ag.GroundedLabelling()[ag.Statements["q"]] == caes.In {
		t.Fatal("scheme-generated default was accidentally stipulated")
	}
	ag.AddAssumption("¬¬d")
	for i := 0; i < 2; i++ {
		if err := ag.Infer(); err != nil {
			t.Fatal(err)
		}
		if ag.Defaults["d"] || ag.GroundedLabelling()[ag.Statements["q"]] != caes.In {
			t.Fatal("reinstantiating the scheme demoted an explicit stipulation")
		}
	}
}
