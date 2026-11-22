package yaml_test

import (
	"bytes"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
	caesyaml "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
)

// These expected bytes are authored inputs, not predictions from the exporter.
// The contract is limited to valid UTF-8 scalar string values at the tested
// statement-text and string-metadata locations, with an ordinary statement ID.
func TestReviewYAMLScalarRoundTrip(t *testing.T) {
	cases := []struct{ name, text string }{
		{"ordinary", "ordinary text"},
		{"mixed_quotes", "He said \"it's fine\"."},
		{"kialo_math_backslash", `Math: \(x\)`},
		{"windows_path", `C:\new\file`},
		{"newline", "first\nsecond"},
		{"controls", "tab\tNUL\x00\x01end"},
		{"boolean_looking", "yes"},
		{"numeric_looking", "123"},
		{"unicode", "Crème brûlée ≤ 5; λ → μ"},
		{"empty", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := caes.NewArgGraph()
			g.Metadata["title"] = tc.text
			g.Statements["s"] = &caes.Statement{Id: "s", Text: tc.text, Label: caes.Undecided}
			var out bytes.Buffer
			caesyaml.Export(&out, g)
			got, err := caesyaml.Import(bytes.NewReader(out.Bytes()))
			if err != nil {
				t.Fatalf("reimport: %v\nexported:\n%s", err, out.String())
			}
			if s := got.Statements["s"]; s == nil || s.Text != tc.text {
				t.Errorf("statement got %#v; want text %q", s, tc.text)
			}
			if value, ok := got.Metadata["title"].(string); !ok || value != tc.text {
				t.Errorf("metadata got %#v; want string %q", got.Metadata["title"], tc.text)
			}
		})
	}
}
