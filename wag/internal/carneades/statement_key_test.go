package carneades

import "testing"

// K3 (docs/c5-item-f.md, section K; the independent check of F1, F2 and F4,
// E1 and E2): the adapter makes a statement under the key the engine gives it
// (terms.Normalize: the formula read with the negation identity), so a second
// spelling finds the first statement; a text that is no formula keeps the
// name it had.
func TestK3AStatementUnderTheEnginesKey(t *testing.T) {
	a := NewAdapter()
	for _, c := range []struct{ first, second, key string }{
		{"¬ p", "¬p", "¬p"},
		{"¬(p)", "¬p", "¬p"},
		{"(q)", "q", "q"},
		{"r( a )", "r(a)", "r(a)"},
		{"s(¬¬x)", "s(x)", "s(x)"},
		{"¬¬t", "t", "t"},
	} {
		s1 := a.findOrCreateStatement(c.first)
		s2 := a.findOrCreateStatement(c.second)
		if s1 != s2 || s1.Id != c.key || a.Graph.Statements[c.key] != s1 {
			t.Errorf("K3 %q then %q: %q and %q, want one statement %q", c.first, c.second, s1.Id, s2.Id, c.key)
		}
	}
	// No formula: named as before (a negated text by itself, another text by
	// its hash).
	if s := a.findOrCreateStatement("¬the tax is good"); s.Id != "¬the tax is good" {
		t.Errorf("K3: the negated text named %q", s.Id)
	}
	if s := a.findOrCreateStatement("the tax is good"); s.Id != generateStatementIDFromText("the tax is good") {
		t.Errorf("K3: the text named %q", s.Id)
	}
}
