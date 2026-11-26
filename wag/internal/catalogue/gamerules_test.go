package catalogue

import (
	"strings"
	"testing"
)

// The rules of a question's locus (C5.2, step 4): a backed id at one
// formula and kind of locus (D095 (1)); every id at one kind of locus.
func TestQuestionLociFaults(t *testing.T) {
	weighing := []askedAt{
		{Scheme: "AS:practical_reasoning", Locus: "exception", Pattern: "bad_outweighs({A},{G})"},
		{Scheme: "AS:consequences", Locus: "exception", Pattern: "bad_outweighs({A},{G})"},
	}
	rebuttal := []askedAt{
		{Scheme: "AS:practical_reasoning", Locus: "conclusion"},
		{Scheme: "AS:consequences", Locus: "conclusion"},
	}
	likelihood := []askedAt{
		{Scheme: "AS:consequences", Locus: "premise", Pattern: "brings_about({A},{G})"},
		{Scheme: "AS:sign", Locus: "premise", Pattern: "strong_correlation({A},{B})"},
	}
	sound := map[string][]askedAt{"CQ:bad_outweighs": weighing, "CQ:side_effect": rebuttal, "CQ:likelihood": likelihood}
	if f := questionLociFaults(sound, map[string]bool{"CQ:bad_outweighs": true, "CQ:side_effect": true}); len(f) > 0 {
		t.Fatalf("sound questions: %v", f)
	}

	// Today's CQ:side_effect: an exception of practical reasoning and the
	// conclusion of positive consequences, backed by a scheme.
	mixed := []askedAt{
		{Scheme: "AS:practical_reasoning", Locus: "exception", Pattern: "side_effects({A},{S1})"},
		{Scheme: "AS:consequences", Locus: "conclusion"},
	}
	f := questionLociFaults(map[string][]askedAt{"CQ:side_effect": mixed}, map[string]bool{"CQ:side_effect": true})
	joined := strings.Join(f, "\n")
	for _, want := range []string{
		"CQ:side_effect is asked at conclusion (AS:consequences) and exception (AS:practical_reasoning): a question is asked at one kind of locus",
		"CQ:side_effect is backed by answer schemes and asked at no formula (AS:consequences) and side_effects({A},{S1}) (AS:practical_reasoning)",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in:\n%s", want, joined)
		}
	}
	// Unbacked, the same mix breaks only the one-locus rule.
	if f := questionLociFaults(map[string][]askedAt{"CQ:side_effect": mixed}, nil); len(f) != 1 || !strings.Contains(f[0], "one kind of locus") {
		t.Errorf("unbacked: %v", f)
	}
	// A backed id at two formulas of one kind of locus.
	two := []askedAt{
		{Scheme: "AS:practical_reasoning", Locus: "exception", Pattern: "bad_outweighs({A},{G})"},
		{Scheme: "AS:negative_consequences", Locus: "exception", Pattern: "good_outweighs({A},{G})"},
	}
	f = questionLociFaults(map[string][]askedAt{"CQ:weighing": two}, map[string]bool{"CQ:weighing": true})
	if len(f) != 1 || !strings.Contains(f[0], "asked at bad_outweighs({A},{G}) (AS:practical_reasoning) and good_outweighs({A},{G}) (AS:negative_consequences)") {
		t.Errorf("a backed id at two formulas: %v", f)
	}
	// Unbacked, two formulas at one kind of locus are allowed.
	if f := questionLociFaults(map[string][]askedAt{"CQ:weighing": two}, nil); len(f) > 0 {
		t.Errorf("unbacked, two formulas: %v", f)
	}
}

// A question's caes_mapping follows from its locus.
func TestAnswerTypes(t *testing.T) {
	for locus, want := range map[string]string{
		"premise":    "support undermining premise",
		"exception":  "support undercutter argument",
		"argument":   "support undercutter argument",
		"conclusion": "support rebuttal conclusion",
	} {
		p, n, tg := answerTypes(locus)
		if got := p + " " + n + " " + tg; got != want {
			t.Errorf("%s: %s, want %s", locus, got, want)
		}
	}
}

// A game question's text is its paraphrase with the variables in braces.
func TestTextFault(t *testing.T) {
	vars := map[string]bool{"A": true, "G": true, "P": true}
	for _, c := range []struct{ text, para, want string }{
		{"Can {P} actually do {A}?", "Can P actually do A?", ""},
		{"Is {G} too slight to count? (A grave harm is not slight.)", "Is G too slight to count? (A grave harm is not slight.)", ""},
		{"Can {P} actually do {A}?", "Can P really do A?", "not the paraphrase"},
		{"Can {Q} actually do {A}?", "Can Q actually do A?", "{Q} is not a variable"},
		{"Can {P} actually do A?", "Can P actually do A?", ""}, // an unbraced variable is the author's to see
	} {
		got := textFault(c.text, c.para, vars)
		if (c.want == "") != (got == "") || !strings.Contains(got, c.want) {
			t.Errorf("%q / %q: %q, want %q", c.text, c.para, got, c.want)
		}
	}
}
