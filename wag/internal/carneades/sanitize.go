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
//   "Dr. Marie Curie"  → dr_marie_curie
//   "climate change"   → climate_change
//   "123 test"         → x_123_test  (atoms can't start with digits)
//   "Steuer für Öl"    → steuer_für_öl
//   "税金"              → 税金
//   ""                 → x_empty
//
// These parse as valid Prolog atoms with zero warnings from the term parser.
//
// The conversion is lossy: case, spacing and punctuation are dropped, so
// "Carbon-Tax" and "carbon tax" name the same thing (which is usually what the
// player means), and Statement.Text shows the atom, not the original words.

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"unicode"
)

// TextToAtom converts arbitrary user text into a valid Prolog atom.
//
// Rules for valid Prolog atoms (matching the carneades-4 term parser, which
// scans identifiers with Go's text/scanner: a letter or underscore, then
// letters, digits and underscores, in any script):
//   - An identifier whose first byte is A-Z is a Variable; anything else is
//     an Atom. Lowercasing prevents the Variable reading.
//   - A digit cannot start an identifier (it would be read as a number).
//   - "true" and "false" are read as booleans.
//
// Conversion steps:
//  1. Trim whitespace and lowercase.
//  2. Keep letters and digits of every script; turn each run of other
//     characters (spaces, punctuation, symbols, underscores) into one "_".
//  3. Strip leading/trailing underscores.
//  4. If the result starts with a digit, prefix with "x_".
//  5. If the result is "true" or "false", prefix with "val_".
//  6. An empty or blank input gives "x_empty". An input with no letter or
//     digit at all (e.g. "€" or "?!") gives "x_" and a hash of the input, so
//     different such inputs stay different.
//
// Letters and digits used to be restricted to ASCII, so every word in another
// script was dropped: "税金" (tax) and "補助金" (subsidy) both became
// "x_empty" and named the same statement. ASCII input converts exactly as
// before.
func TextToAtom(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "x_empty"
	}

	var b strings.Builder
	sep := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if sep && b.Len() > 0 {
				b.WriteByte('_')
			}
			sep = false
			b.WriteRune(r)
		} else {
			sep = true
		}
	}
	out := b.String()
	if out == "" {
		h := sha1.Sum([]byte(s))
		return "x_" + hex.EncodeToString(h[:4])
	}

	// A digit cannot start an atom.
	for _, r := range out {
		if unicode.IsDigit(r) {
			out = "x_" + out
		}
		break
	}

	// "true" and "false" are parsed as Bool(true)/Bool(false), not Atom.
	// This would cause subtle bugs because Bool.String() returns "true"/"false"
	// which look like atoms but behave differently in matching/substitution.
	if out == "true" || out == "false" {
		out = "val_" + out
	}
	return out
}

// SanitizeBinding converts user-supplied text into a form suitable for
// terms.ReadString. This is the improved version of Block 1's sanitizeBinding.
//
// It returns a clean Prolog atom that:
//   - Parses without warnings from the term parser
//   - Is deterministic (same input → same output)
//   - Does NOT start with uppercase (avoids Variable interpretation)
//
// This function replaces the quoteForProlog approach from Block 1.
func SanitizeBinding(s string) string {
	return TextToAtom(s)
}
