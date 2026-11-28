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
		// K8 (docs/c5-item-f.md; the independent check of F1, F2 and F4, W1): a
		// variable is a constant of its own, named by its name, which a
		// placeholder can stand for and no constant of the pattern equals (the
		// engine's reader gives each occurrence of X its own variable). "_" is
		// an atom for the engine's reader. F1 took ground statements only.
		{"nonground_statement", "worth(A,g)", "worth({A},{G})", true},
		{"variable_against_constant", "worth(A,g)", "worth(a,{G})", false},
		{"variable_repeated", "same(X,X)", "same({X},{X})", true},
		{"variables_different", "same(X,Y)", "same({X},{X})", false},
		{"underscore_is_an_atom", "same(_,_)", "same({X},{X})", true},
		{"bridge_exception_of_a_variable_claim", "worth_finding_out_first(X,demand_falls)", "worth_finding_out_first({A},{G})", true},
		// K8b (the independent check of K, NK1): a variable inside a list is
		// a constant of its own too.
		{"variable_in_a_list_repeated", "same([X],[X])", "same({A},{A})", true},
		{"variables_in_lists_different", "same([X],[Y])", "same({A},{A})", false},
		// K6: the statement as the engine keys it (the negation identity, a
		// comment, unary + and a doubled -); unary - and ! are other
		// statements, as its negation is ("negative", above).
		{"double_negation", "¬¬worth(a,g)", "worth({A},{G})", true},
		{"brackets", "(worth(a,g))", "worth({A},{G})", true},
		{"comment", "worth/*c*/(a,g)", "worth({A},{G})", true},
		{"unary_plus", "+worth(a,g)", "worth({A},{G})", true},
		{"doubled_minus", "--worth(a,g)", "worth({A},{G})", true},
		{"double_negation_inside", "worth(¬¬a,g)", "worth({A},{G})", true},
		{"unary_minus", "-worth(a,g)", "worth({A},{G})", false},
		{"bang", "!worth(a,g)", "worth({A},{G})", false},
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
