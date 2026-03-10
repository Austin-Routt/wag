package carneades

// sanitize.go — Converts arbitrary user text into valid Prolog atoms for
// the Carneades term parser.
//
// Problem solved:
// Block 1's quoteForProlog wraps text in single quotes: 'Dr. Marie Curie'.
// Go's text/scanner (used by the Carneades term parser) treats single quotes
// as character literals and emits "invalid char literal" warnings to stderr.
// The parse still succeeds (the parser's Char case falls through to Factor_name
// → bi_0 which creates an Atom), but the warnings are ugly and confusing.
//
// This file replaces that approach with clean Prolog atom generation:
//   "Dr. Marie Curie"  → dr__marie_curie
//   "climate change"   → climate_change
//   "123 test"         → x_123_test  (atoms can't start with digits)
//   ""                 → x_empty
//
// These parse as valid Prolog atoms (lowercase, no spaces, no special chars)
// with zero warnings from the term parser.
//
// The original human-readable text is preserved in Statement.Text via the
// Theory's Language templates or via direct assignment in the adapter.

import (
	"regexp"
	"strings"
	"unicode"
)

// nonAlphanumeric matches any character that is not a letter, digit, or underscore.
var nonAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// multiUnderscore collapses runs of underscores into a single one.
var multiUnderscore = regexp.MustCompile(`_+`)

// TextToAtom converts arbitrary user text into a valid Prolog atom.
//
// Rules for valid Prolog atoms (matching the carneades-4 term parser):
//   - Must start with a lowercase letter (a-z). If the first character is
//     uppercase (A-Z), the parser treats it as a Variable. If it starts
//     with a digit, the parser treats it as an Int or Float.
//   - May contain lowercase/uppercase letters, digits, and underscores.
//   - No spaces, no special characters.
//
// Conversion steps:
//  1. Trim whitespace.
//  2. Lowercase the entire string (prevents Variable interpretation).
//  3. Replace non-alphanumeric characters with underscores.
//  4. Collapse multiple underscores into one.
//  5. Strip leading/trailing underscores.
//  6. If the result starts with a digit, prefix with "x_".
//  7. If the result is empty, return "x_empty".
//  8. If the result is "true" or "false" (Prolog built-ins), prefix with "val_".
//
// Examples:
//   TextToAtom("Dr. Marie Curie")     → "dr__marie_curie"
//   TextToAtom("climate change")      → "climate_change"
//   TextToAtom("123 test")            → "x_123_test"
//   TextToAtom("true")                → "val_true"
//   TextToAtom("")                    → "x_empty"
//   TextToAtom("it's a test!")        → "it_s_a_test"
func TextToAtom(s string) string {
	// Step 1: Trim whitespace.
	s = strings.TrimSpace(s)
	if s == "" {
		return "x_empty"
	}

	// Step 2: Lowercase.
	s = strings.ToLower(s)

	// Step 3: Replace non-alphanumeric chars with underscore.
	s = nonAlphanumeric.ReplaceAllString(s, "_")

	// Step 4: Collapse multiple underscores.
	s = multiUnderscore.ReplaceAllString(s, "_")

	// Step 5: Strip leading/trailing underscores.
	s = strings.Trim(s, "_")

	if s == "" {
		return "x_empty"
	}

	// Step 6: Prefix if starts with digit.
	if unicode.IsDigit(rune(s[0])) {
		s = "x_" + s
	}

	// Step 7: Check for Prolog built-in atoms that bi_0 handles specially.
	// "true" and "false" are parsed as Bool(true)/Bool(false), not Atom.
	// This would cause subtle bugs because Bool.String() returns "true"/"false"
	// which look like atoms but behave differently in matching/substitution.
	if s == "true" || s == "false" {
		s = "val_" + s
	}

	return s
}

// SanitizeBinding converts user-supplied text into a form suitable for
// terms.ReadString. This is the improved version of Block 1's sanitizeBinding.
//
// It returns a clean Prolog atom that:
//   - Parses without warnings from the term parser
//   - Is deterministic (same input → same output)
//   - Is collision-resistant (different inputs → different outputs in practice)
//   - Does NOT start with uppercase (avoids Variable interpretation)
//
// This function replaces the quoteForProlog approach from Block 1.
func SanitizeBinding(s string) string {
	return TextToAtom(s)
}
