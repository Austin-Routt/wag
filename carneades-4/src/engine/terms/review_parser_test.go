package terms

import "testing"

func TestReviewTruncatedTermsAreRejected(t *testing.T) {
	for _, s := range []string{"p(", "p(a,)", "p([a,b)", "f(g(", "\"unterminated", "p(a) trailing"} {
		if _, ok := ReadString(s); ok {
			t.Errorf("accepted malformed term %q", s)
		}
	}
	for _, s := range []string{"p(a)", "'Joe Smith'", "'foo'('Joe Smith')", "p([a,b])"} {
		if _, ok := ReadString(s); !ok {
			t.Errorf("rejected valid term %q", s)
		}
	}
}
