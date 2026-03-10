package carneades

import (
	"testing"

	"github.com/carneades/carneades-4/src/engine/terms"
)

// TestTextToAtom verifies that TextToAtom produces valid Prolog atoms
// that parse cleanly through terms.ReadString without any warnings.
func TestTextToAtom(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Basic conversions
		{"hello", "hello"},
		{"Hello World", "hello_world"},
		{"climate change", "climate_change"},

		// Special characters → underscores
		{"Dr. Marie Curie", "dr_marie_curie"},
		{"it's a test!", "it_s_a_test"},
		{"cost/benefit", "cost_benefit"},
		{"A & B", "a_b"},

		// Leading digits → prefix
		{"123 test", "x_123_test"},
		{"42", "x_42"},

		// Empty and whitespace
		{"", "x_empty"},
		{"   ", "x_empty"},

		// Prolog reserved words
		{"true", "val_true"},
		{"false", "val_false"},

		// Already valid atoms
		{"reduce_carbon_emissions", "reduce_carbon_emissions"},
		{"my_premise", "my_premise"},

		// Mixed case → lowercase
		{"CamelCase", "camelcase"},
		{"ALL_CAPS", "all_caps"},

		// Multiple special chars → single underscore
		{"a---b...c", "a_b_c"},
	}
	for _, tt := range tests {
		got := TextToAtom(tt.input)
		if got != tt.want {
			t.Errorf("TextToAtom(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestTextToAtomParsesCleanly verifies that every output from TextToAtom
// parses through terms.ReadString as an Atom (not a Variable, not an error).
func TestTextToAtomParsesCleanly(t *testing.T) {
	inputs := []string{
		"climate change",
		"Dr. Marie Curie",
		"reduce carbon emissions",
		"42 reasons",
		"it's complicated!",
		"A & B → C",
		"",
		"true",
		"false",
		"hello",
	}
	for _, input := range inputs {
		atom := TextToAtom(input)
		term, ok := terms.ReadString(atom)
		if !ok {
			t.Errorf("TextToAtom(%q) = %q — failed to parse as Prolog term", input, atom)
			continue
		}
		// Verify it parsed as an Atom, not a Variable.
		// In the Carneades term parser, identifiers starting with A-Z
		// are Variables. Our TextToAtom should always produce lowercase.
		s := term.String()
		if s != atom {
			t.Errorf("TextToAtom(%q) = %q — term.String() = %q (mismatch)",
				input, atom, s)
		}
	}
}

// TestSanitizeBindingMatchesTextToAtom verifies that SanitizeBinding
// is a direct alias for TextToAtom.
func TestSanitizeBindingMatchesTextToAtom(t *testing.T) {
	inputs := []string{"hello world", "Dr. X", "123", "", "true"}
	for _, input := range inputs {
		got := SanitizeBinding(input)
		want := TextToAtom(input)
		if got != want {
			t.Errorf("SanitizeBinding(%q) = %q, TextToAtom(%q) = %q — should match",
				input, got, input, want)
		}
	}
}
