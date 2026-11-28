package language

import "testing"

// The Go half of case EN13 (docs/d086-first-gold.md §6; the independent
// review of D098 and R21, N3): validate.py's R21 fills walton.yml's
// templates as Render does (its go_render), and these are the strings both
// give: %s after %[n]s takes the argument after the n-th, %% is a literal %,
// and a template Render refuses is not filled.
func TestRenderAsR21Fills(t *testing.T) {
	for _, c := range []struct {
		tmpl string
		args []string
		want string
		ok   bool
	}{
		{"%[3]s, from %[1]s then %s.", []string{"A", "B", "C"}, "C, from A then B.", true},
		{"%s is bad, 100%%.", []string{"A"}, "A is bad, 100%.", true},
		{"Performing %s in %s would bring about %s.", []string{"A", "B", "C"}, "Performing A in B would bring about C.", true},
		{"%[0]s is good.", []string{"A"}, "", false},
		{"%s and %s.", []string{"A", "B", "C"}, "", false}, // an argument left out
	} {
		got, err := Render(c.tmpl, c.args)
		if (err == nil) != c.ok || got != c.want {
			t.Errorf("Render(%q, %v) = %q, %v; want %q, ok %v", c.tmpl, c.args, got, err, c.want, c.ok)
		}
	}
}
