package carneades

import (
	"os"
	"strings"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	caesYAML "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
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
			for _, fact := range tt.facts {
				g.AddAssumption(fact)
			}
			if err := g.Infer(); err != nil {
				t.Fatal(err)
			}
			if got := migrationHasArgument(g, tt.scheme, tt.conclusion); got != tt.want {
				t.Fatalf("argument from %s for %s: got %v, want %v", tt.scheme, tt.conclusion, got, tt.want)
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
