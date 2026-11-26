package game

import (
	"github.com/yourorg/wag/data"
	"strings"
	"testing"
)

const indexedThreat = "I hereby assert that in fact I will see to it that %[2]s occurs if you bring about %[1]s."

func TestIndexedTemplateCardAndBoard(t *testing.T) {
	gs := modeGame(t, "classic", "Alice", "Bob")
	gs.Adapter.Graph.Theory.Language["assert_bad/2"] = indexedThreat
	if got, want := FormulaText(gs.Adapter.Graph.Theory.Language, "assert_bad(X,Y)"), "I hereby assert that in fact I will see to it that {Y} occurs if you bring about {X}"; got != want {
		t.Fatalf("card %q != %q", got, want)
	}
	if got, want := gs.Say("assert_bad(act,harm)"), "I hereby assert that in fact I will see to it that ‘harm’ occurs if you bring about ‘act’."; got != want {
		t.Fatalf("board %q != %q", got, want)
	}
}

func TestIndexedTemplateLint(t *testing.T) {
	for _, tc := range []struct {
		template string
		invalid  bool
	}{{"%[1]s is good.", false}, {"%[2]s is good.", true}, {"%v is good.", true}} {
		input := withTemplate(t, data.CoreYAML, "good/1", tc.template)
		issues, err := LintData(input, data.SchemesJSON, data.CQsJSON, data.EdgesJSON)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, issue := range issues {
			if issue.Error && strings.Contains(issue.String(), "good/1") {
				found = true
			}
		}
		if found != tc.invalid {
			t.Errorf("template %q: invalid=%v, want %v", tc.template, found, tc.invalid)
		}
	}
}
