package carneades

import (
	"fmt"
	"os"
	"strings"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	caesYAML "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"

	"github.com/yourorg/wag/internal/language"
)

// Use the actual committed collection and native inference, rather than
// InstantiateScheme (which can bypass the reserved predicate's CHR meaning).
func migratedLibraryGraph(t *testing.T, schemes ...string) *caes.ArgGraph {
	t.Helper()
	path := os.Getenv("WAG_MIGRATION_LIBRARY")
	if path == "" {
		path = "../../../docs/scheme-registry/walton-clean/walton.yml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	g, err := caesYAML.Import(strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	selected := make([]*caes.Scheme, 0, len(schemes))
	for _, id := range schemes {
		for _, scheme := range g.Theory.ArgSchemes {
			if scheme.Id == id {
				selected = append(selected, scheme)
			}
		}
	}
	if len(selected) != len(schemes) {
		t.Fatalf("could not select schemes %v", schemes)
	}
	g.Theory.ArgSchemes = selected
	g.Theory.InitSchemeIndex()
	return g
}

func migrationHasArgument(g *caes.ArgGraph, scheme, conclusion string) bool {
	for _, argument := range g.Arguments {
		if argument.Scheme != nil && argument.Scheme.Id == scheme && argument.Conclusion != nil && argument.Conclusion.Id == conclusion {
			return true
		}
	}
	return false
}

func TestD084NativeInference(t *testing.T) {
	tests := []struct {
		name, scheme, conclusion string
		facts                    []string
		want                     bool
	}{
		{"group positive", "group_2", "is_a(mammal,warm)", []string{"group_quality(animals,warm)", "group_member(mammal,animals)", "group_member_quality(animals,warm,mammal)"}, true},
		{"group wrong member", "group_2", "is_a(mammal,warm)", []string{"group_quality(animals,warm)", "group_member(mammal,animals)", "group_member_quality(animals,warm,reptile)"}, false},
		{"group wrong quality", "group_2", "is_a(mammal,warm)", []string{"group_quality(animals,warm)", "group_member(mammal,animals)", "group_member_quality(animals,cold,mammal)"}, false},
		{"physical positive", "rule_physical_world_1", "means_same(w,d)", []string{"is_a(w,d)", "synonymous"}, true},
		{"physical missing warrant", "rule_physical_world_1", "means_same(w,d)", []string{"is_a(w,d)"}, false},
		{"physical wrong terms", "rule_physical_world_1", "means_same(w,d)", []string{"is_a(other,d)", "synonymous"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := migratedLibraryGraph(t, tt.scheme)
			conclusion, facts := tt.conclusion, tt.facts
			if _, legacy := g.Theory.Language["is/2"]; legacy {
				// Comparative run against the pre-D084 library: the same
				// cases in its own name. Only "group positive" must fail
				// there, and for the reason D084 fixes: GoCHR runs is/2 as
				// an assignment, so group_2's conclusion is never drawn.
				conclusion = strings.ReplaceAll(conclusion, "is_a(", "is(")
				facts = nil
				for _, fact := range tt.facts {
					facts = append(facts, strings.ReplaceAll(fact, "is_a(", "is("))
				}
			}
			for _, fact := range facts {
				g.AddAssumption(fact)
			}
			if err := g.Infer(); err != nil {
				t.Fatal(err)
			}
			if got := migrationHasArgument(g, tt.scheme, conclusion); got != tt.want {
				t.Fatalf("argument from %s for %s: got %v, want %v", tt.scheme, conclusion, got, tt.want)
			}
		})
	}
}

// The conclusion remains means_same(...) on both sides of the migration.
// This catches the reserved-name bridge failure without merely expecting a
// newly renamed conclusion from an old library.
func TestD084NativeTwoSchemeBridge(t *testing.T) {
	g := migratedLibraryGraph(t, "group_2", "rule_physical_world_1")
	for _, fact := range []string{"group_quality(animals,warm)", "group_member(mammal,animals)", "group_member_quality(animals,warm,mammal)", "synonymous"} {
		g.AddAssumption(fact)
	}
	if err := g.Infer(); err != nil {
		t.Fatal(err)
	}
	if !migrationHasArgument(g, "rule_physical_world_1", "means_same(mammal,warm)") {
		t.Fatal("native inference did not connect the classification conclusion to the constitutive-rule premise")
	}
}

func TestD081NativePropertyScope(t *testing.T) {
	tests := []struct {
		name, start, end, property string
		want                       bool
	}{
		{"same property", "clear", "target", "bald", true},
		{"different property", "clear", "target", "rich", false},
		{"unrelated starting case", "other", "target", "bald", false},
		{"unrelated target case", "clear", "other", "bald", false},
		{"missing chain", "", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := migratedLibraryGraph(t, "sorites_slippery_slope")
			chain := fmt.Sprintf("reapplication_sequence(%s,%s,%s)", tt.start, tt.end, tt.property)
			if _, legacy := g.Theory.Language["reapplication_sequence/2"]; legacy {
				// Comparative run against the pre-D081 schema. This is its
				// representable chain fact; property provenance is necessarily
				// lost. The different-property case must fail on that version.
				chain = fmt.Sprintf("reapplication_sequence(%s,%s)", tt.start, tt.end)
			}
			for _, fact := range []string{"beyond_contention(clear,bald)", "all_has(clear,bald,next)"} {
				g.AddAssumption(fact)
			}
			if tt.property != "" { // An empty fixture property means no chain evidence.
				g.AddAssumption(chain)
			}
			if err := g.Infer(); err != nil {
				t.Fatal(err)
			}
			if got := migrationHasArgument(g, "sorites_slippery_slope", "may_have(target,bald)"); got != tt.want {
				t.Fatalf("chain %s from %s to %s licensed may_have(target,bald): got %v, want %v", tt.property, tt.start, tt.end, got, tt.want)
			}
		})
	}
}

// renderAtom reads a ground atom with the library's own template, as the
// catalogue and the game do.
func renderAtom(t *testing.T, g *caes.ArgGraph, atom string) string {
	t.Helper()
	name, args := atom, []string(nil)
	if open := strings.Index(atom, "("); open >= 0 {
		name, args = atom[:open], strings.Split(strings.TrimSuffix(atom[open+1:], ")"), ",")
	}
	tmpl, ok := g.Theory.Language[fmt.Sprintf("%s/%d", name, len(args))]
	if !ok {
		t.Fatalf("no template for %s/%d", name, len(args))
	}
	s, err := language.Render(tmpl, args)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(s)
}

// D081's three approved repairs, with values that differ in every role, so
// that a repair that put a value in the wrong place would show (the private
// workspace's verify/r-stage-review-2026-11-22, report B, checked the roles
// against the book, pp. 312, 318 and 322).
func TestD081RepairRoles(t *testing.T) {
	infer := func(scheme string, facts ...string) *caes.ArgGraph {
		g := migratedLibraryGraph(t, scheme)
		for _, fact := range facts {
			g.AddAssumption(fact)
		}
		if err := g.Infer(); err != nil {
			t.Fatal(err)
		}
		return g
	}

	// p. 318: treatment Y for subject X, so the contrary treatment Y' for the
	// contrary subject X'. WAG's formulas put the treatment first.
	g := infer("rhetorical_oppositions_normative", "treatment_required(admit,resident)", "if_treatment_required(admit,resident,visitor,exclude)")
	if !migrationHasArgument(g, "rhetorical_oppositions_normative", "contrary_treatment_required(exclude,visitor)") {
		t.Fatal("rhetorical oppositions: no argument for the contrary treatment of the contrary subject")
	}
	for atom, want := range map[string]string{
		"if_treatment_required(admit,resident,visitor,exclude)": "If the treatment admit is required for the subject resident, for the contrary subject visitor the contrary treatment exclude is required.",
		"contrary_treatment_required(exclude,visitor)":          "The contrary treatment exclude is required for the subject visitor.",
	} {
		if got := renderAtom(t, g, atom); got != want {
			t.Errorf("%s reads %q, want %q", atom, got, want)
		}
	}
	// The schemes library writes the subject first; such a fact under WAG's
	// formula must not yield the conclusion about the right roles.
	g = infer("rhetorical_oppositions_normative", "treatment_required(admit,resident)", "if_treatment_required(resident,admit,exclude,visitor)")
	if migrationHasArgument(g, "rhetorical_oppositions_normative", "contrary_treatment_required(exclude,visitor)") {
		t.Fatal("rhetorical oppositions: a subject-first conditional yielded the treatment-first conclusion")
	}

	// p. 322: the value V of the sacrifice is the value the conclusion states.
	g = infer("sacrifice", "sacrifice_made(liberty,years)", "great_sacrifice(years,liberty,worth)", "great_sacrifice_made(years,liberty)")
	if !migrationHasArgument(g, "sacrifice", "sacrifice_value(liberty,worth)") {
		t.Fatal("sacrifice: no argument for the value the premise states")
	}
	g = infer("sacrifice", "sacrifice_made(liberty,years)", "great_sacrifice(years,liberty,other)", "great_sacrifice_made(years,liberty)")
	if migrationHasArgument(g, "sacrifice", "sacrifice_value(liberty,worth)") {
		t.Fatal("sacrifice: a value the premise does not state reached the conclusion")
	}

	// p. 312: the speaker's moral justification is the conclusion; with it
	// listed as a premise the scheme had no conclusion and built no argument.
	g = infer("moral_justification_ad_populum_excuse", "group_does(drivers,speeding,policy)", "acceptable_norm(speeding,policy,drivers)", "group_member(drivers)")
	if !migrationHasArgument(g, "moral_justification_ad_populum_excuse", "morally_acceptable_action(speeding,policy)") {
		t.Fatal("moral justification: no argument for the speaker's justification")
	}
}
