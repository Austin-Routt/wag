package yaml_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
	caesyaml "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
)

// Check the actual Export -> Import contract, comparing each string with its
// authored input, not an expected value calculated by another encoder.
func TestConvergenceYAMLUnicodeRoundTrip(t *testing.T) {
	cases := []struct{ name, text string }{
		{"DEL", "before\u007fafter"},
		{"C1_80", "before\u0080after"},
		{"NEL", "before\u0085after"},
		{"C1_9f", "before\u009fafter"},
		{"noncharacter_FFFE", "before\ufffeafter"},
		{"noncharacter_FFFF", "before\uffffafter"},
		{"line_separator", "before\u2028after"},
		{"paragraph_separator", "before\u2029after"},
		{"nonBMP", "before\U0001f600after\U0010ffff"},
	}
	var boundaries strings.Builder
	for r := rune(0); r <= 0x9f; r++ {
		boundaries.WriteRune(r)
	}
	for _, r := range []rune{0xd7ff, 0xe000, 0xfdd0, 0xfdef, 0xfffd, 0xfffe, 0xffff, 0x10000, 0x1fffe, 0x1ffff, 0x10fffd, 0x10fffe, 0x10ffff} {
		boundaries.WriteRune(r)
	}
	cases = append(cases, struct{ name, text string }{"control_and_scalar_boundaries", boundaries.String()})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := caes.NewArgGraph()
			g.Metadata["title"] = tc.text
			g.Statements["s"] = &caes.Statement{Id: "s", Text: tc.text, Label: caes.Undecided}
			var out bytes.Buffer
			caesyaml.Export(&out, g)
			got, err := caesyaml.Import(bytes.NewReader(out.Bytes()))
			if err != nil {
				t.Fatalf("reimport: %v; export %q", err, out.String())
			}
			if s := got.Statements["s"]; s == nil || s.Text != tc.text {
				t.Errorf("statement got %v, want text %q", s, tc.text)
			}
			if value, ok := got.Metadata["title"].(string); !ok || value != tc.text {
				t.Errorf("metadata got %#v; want string %q", got.Metadata["title"], tc.text)
			}
		})
	}
}

// This property check covers all valid Unicode scalar values in 17 bounded
// batches, as actual graph string values. It does not cover identifiers, other
// metadata types, invalid UTF-8 byte sequences or complete graph round trips.
func TestConvergenceYAMLAllUnicodeScalars(t *testing.T) {
	for plane := 0; plane <= 16; plane++ {
		t.Run(fmt.Sprintf("plane_%02x", plane), func(t *testing.T) {
			var b strings.Builder
			for r := rune(plane * 0x10000); r < rune((plane+1)*0x10000); r++ {
				if r >= 0xd800 && r <= 0xdfff {
					continue
				}
				b.WriteRune(r)
			}
			text := b.String()
			g := caes.NewArgGraph()
			g.Metadata["title"] = text
			g.Statements["s"] = &caes.Statement{Id: "s", Text: text, Label: caes.Undecided}
			var out bytes.Buffer
			caesyaml.Export(&out, g)
			got, err := caesyaml.Import(bytes.NewReader(out.Bytes()))
			if err != nil {
				t.Fatalf("reimport: %v", err)
			}
			if s := got.Statements["s"]; s == nil || s.Text != text {
				t.Fatal("statement scalar bytes changed")
			}
			if value, ok := got.Metadata["title"].(string); !ok || value != text {
				t.Fatal("metadata scalar bytes/type changed")
			}
		})
	}
}
