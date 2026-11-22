package yaml_test

import (
	"bytes"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
	caesyaml "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
)

// Invalid UTF-8 in a statement's text or a string value of the metadata comes
// back as U+FFFD, never as a different valid character.
func TestYAMLInvalidUTF8BecomesReplacementCharacter(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"a\xffb", "a\uFFFDb"},
		{"caf\xe9", "caf\uFFFD"},
		{"\xc3(", "\uFFFD("},
		{"ok \u00e9", "ok \u00e9"},
	} {
		g := caes.NewArgGraph()
		g.Metadata["title"] = tc.in
		g.Statements["s"] = &caes.Statement{Id: "s", Text: tc.in, Label: caes.Undecided}
		var out bytes.Buffer
		caesyaml.Export(&out, g)
		got, err := caesyaml.Import(bytes.NewReader(out.Bytes()))
		if err != nil {
			t.Fatalf("%q: reimport: %v", tc.in, err)
		}
		if s := got.Statements["s"]; s == nil || s.Text != tc.want {
			t.Errorf("%q: statement got %#v, want text %q", tc.in, s, tc.want)
		}
		if v, _ := got.Metadata["title"].(string); v != tc.want {
			t.Errorf("%q: metadata got %q, want %q", tc.in, v, tc.want)
		}
	}
}
