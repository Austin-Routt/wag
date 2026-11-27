package language

import "unicode"

// NamesNothing tells whether a value written for a variable names nothing: it
// has no letter and no digit, in any script. A blank value, punctuation alone
// ("?", "--") and invisible characters (a zero-width space, a byte order
// mark) name nothing. The game reads such a value as no binding (S2): an
// answer whose H, A2 or I names nothing would make its exception out without
// saying what outweighs, what does better, or what would settle it (step
// C5.2's item F, G-unbound).
func NamesNothing(v string) bool {
	for _, r := range v {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
