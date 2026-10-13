package carneades

import (
	"regexp"
	"strings"
	"testing"
	"unicode"

	"github.com/carneades/carneades-4/src/engine/terms"
)

// legacyTextToAtom is TextToAtom before it kept non-ASCII letters. ASCII input
// with a letter or digit must still convert the same way, so that seeded games
// played in ASCII keep their move digests.
func legacyTextToAtom(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "x_empty"
	}
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(s, "_")
	s = regexp.MustCompile(`_+`).ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if s == "" {
		return "x_empty"
	}
	if unicode.IsDigit(rune(s[0])) {
		s = "x_" + s
	}
	if s == "true" || s == "false" {
		s = "val_" + s
	}
	return s
}

func TestTextToAtomKeepsEveryScript(t *testing.T) {
	for in, want := range map[string]string{
		"税金":              "税金",
		"補助金":             "補助金",
		"Ελλάδα":          "ελλάδα",
		"Steuer für CO₂":  "steuer_für_co", // ₂ is not a decimal digit (nor an identifier rune)
		"naïve plan":      "naïve_plan",
		"٣ أسباب":         "x_٣_أسباب", // starts with an Arabic-Indic digit
		"carbon tax (EU)": "carbon_tax_eu",
		"Steuer € 5":      "steuer_5",
	} {
		if got := TextToAtom(in); got != want {
			t.Errorf("TextToAtom(%q) = %q, want %q", in, got, want)
		}
	}
	if TextToAtom("税金") == TextToAtom("補助金") {
		t.Error("two different words in another script give the same atom")
	}
	if a, b := TextToAtom("€"), TextToAtom("$"); a == b || a == "x_empty" {
		t.Errorf("symbol-only inputs: %q and %q", a, b)
	}
}

// checkAtom fails unless atom parses as an Atom on its own and as the argument
// of a compound term, printing back unchanged.
func checkAtom(t *testing.T, in, atom string) {
	t.Helper()
	tm, ok := terms.ReadString(atom)
	if !ok {
		t.Fatalf("TextToAtom(%q) = %q does not parse", in, atom)
	}
	if _, isAtom := tm.(terms.Atom); !isAtom || tm.String() != atom {
		t.Fatalf("TextToAtom(%q) = %q parses as %T %q", in, atom, tm, tm.String())
	}
	c, ok := terms.ReadString("p(" + atom + ",x)")
	if !ok {
		t.Fatalf("p(%s,x) does not parse", atom)
	}
	if cc, isC := c.(terms.Compound); !isC || len(cc.Args) != 2 || cc.Args[0].String() != atom {
		t.Fatalf("p(%s,x) parses as %q", atom, c.String())
	}
}

func FuzzTextToAtom(f *testing.F) {
	for _, s := range []string{"", " ", "Dr. Marie Curie", "税金", "Ελλάδα", "123", "true", "TRUE", "_x_", "a b", "€", "x(y)", "'q'", "\"q\"",
		"or", "and", "in", "is", "İstanbul", "é", "ǅ", "ß", "😀 smile", "//comment", "/*c*/", "a\nb", "\x00"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		atom := TextToAtom(s)
		checkAtom(t, s, atom)
		if TextToAtom(s) != atom {
			t.Fatal("not deterministic")
		}
		ascii := true
		alnum := false
		for _, r := range s {
			ascii = ascii && r < unicode.MaxASCII
			alnum = alnum || unicode.IsLetter(r) || unicode.IsDigit(r)
		}
		if ascii && alnum {
			if legacy := legacyTextToAtom(s); legacy != atom {
				t.Fatalf("ASCII input %q: %q, previously %q", s, atom, legacy)
			}
		}
	})
}
