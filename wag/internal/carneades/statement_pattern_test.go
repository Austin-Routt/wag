package carneades

import "testing"

// TestStatementPatternMatchesTheForm: a statement matches a question's
// exception pattern by its form (F1 of the external review of exchange 3,
// docs/c5-item-f.md A9.8 to A9.12): placeholders vary, constants, arity,
// nesting and repeated placeholders do not.
func TestStatementPatternMatchesTheForm(t *testing.T) {
	var a Adapter
	for _, c := range []struct {
		name, stmt, pattern string
		want                bool
	}{
		{"ground", "worth(a,g)", "worth({A},{G})", true},
		{"whitespace", "worth ( a, g )", "worth ({A}, {G})", true},
		{"constant", "worth(a,now)", "worth({A},now)", true},
		{"wrong_constant", "worth(a,later)", "worth({A},now)", false},
		{"wrong_predicate", "worth_more(a,g)", "worth({A},{G})", false},
		{"wrong_arity", "worth(a)", "worth({A},{G})", false},
		{"negative", "¬worth(a,g)", "worth({A},{G})", false},
		{"repeated_same", "same(a,a)", "same({X},{X})", true},
		{"repeated_different", "same(a,b)", "same({X},{X})", false},
		{"nested", "worth(act(a),effect(g))", "worth({A},effect({G}))", true},
		{"different_nesting", "worth(act(a),value(g))", "worth({A},effect({G}))", false},
		{"nonground_statement", "worth(A,g)", "worth({A},{G})", false},
		{"unmarked_pattern_variable", "worth(a,g)", "worth(A,{G})", false},
		{"generated_name_collision", "worth(a,a)", "worth(PatternVar0,{G})", false},
		{"quoted_generated_name_constant", "worth(\"PatternVar0\",a)", "worth(\"PatternVar0\",{G})", true},
		// The fast path skips parsing for an unrelated predicate; a space before
		// "(" or a quote must not take it (the check of the patch, 01-i).
		{"space_before_paren", "worth (a,g)", "worth({A},{G})", true},
		{"quoted_atom", "worth('tax meat',g)", "worth({A},{G})", true},
		{"negated_pattern", "¬worth(a,g)", "¬worth({A},{G})", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := a.MatchesStatementPattern(c.stmt, c.pattern); got != c.want {
				t.Errorf("match %q %q = %v, want %v", c.stmt, c.pattern, got, c.want)
			}
		})
	}
}
