package catalogue

import (
	"strings"
	"testing"
)

// A formula's predicates, outermost first, with their arities; variables
// and constants inside a statement are none.
func TestPredicatesOf(t *testing.T) {
	for f, want := range map[string]string{
		"brings_about(A,G)":               "brings_about/2",
		"¬should_be_performed(A)":         "should_be_performed/1",
		"based_on_evidence(asserts(W,S))": "based_on_evidence/1 asserts/2",
		"more_reliable":                   "more_reliable/0",
		"S":                               "",
		"in_case(A, c1)":                  "in_case/2",
		"f(g(h(X)), k(Y, z))":             "f/2 g/1 h/1 k/2",
	} {
		got, err := predicatesOf(f)
		if err != nil || strings.Join(got, " ") != want {
			t.Errorf("%s: %v %v, want %s", f, got, err, want)
		}
	}
	for _, f := range []string{"", "p(A", "p(A,)", "p(A) q", "(A)"} {
		if _, err := predicatesOf(f); err == nil {
			t.Errorf("%q read", f)
		}
	}
}

const langCore = `language:
  # How statements read.
  # @language

argument_schemes:
  - id: s
    variables: [A, G]
    conclusions:
      - should_be_performed(A)
    premises:
      - brings_about(A,G)
      - good(G)
`

// The language's sentences are walton.yml's, or the recorded differences,
// each with its negation's, in the order the schemes use them (rule (e)).
func TestFillLanguage(t *testing.T) {
	walton := map[string]string{"should_be_performed/1": "Action %s should be performed.", "brings_about/2": "The action %s brings about %s.",
		"good/1": "%s is good.", "bad/1": "%s is bad."}
	l := GameLanguage{
		Differences: []LanguageDifference{{Reason: "the board's", Templates: map[string]string{"should_be_performed/1": "%s should be done."}}},
		Negations: map[string]string{"should_be_performed/1": "%s should not be done.", "brings_about/2": "The action %s does not bring about %s.",
			"good/1": "%s is not good."},
	}
	got, err := fillLanguage(langCore, walton, l)
	if err != nil {
		t.Fatal(err)
	}
	want := `language:
  # How statements read.
  should_be_performed/1: "%s should be done."
  ¬should_be_performed/1: "%s should not be done."
  brings_about/2: "The action %s brings about %s."
  ¬brings_about/2: "The action %s does not bring about %s."
  good/1: "%s is good."
  ¬good/1: "%s is not good."

argument_schemes:
`
	if !strings.HasPrefix(got, want) {
		t.Errorf("got:\n%s", got)
	}
	// Read back, the template is the one written.
	if back := languageTemplate(got); back != langCore {
		t.Errorf("the template read back:\n%s", back)
	}

	for _, c := range []struct {
		name string
		core string
		edit func(l *GameLanguage, w map[string]string)
		want string
	}{
		{"no marker", strings.Replace(langCore, "  # @language\n", "", 1), nil, `no marker "# @language"`},
		{"two markers", strings.Replace(langCore, "  # @language\n", "  # @language\n  # @language\n", 1), nil, "is not once in the language's block"},
		{"a marker outside the block", langCore + "  # @language\n", nil, "is not once in the language's block"},
		{"a sentence in the template", strings.Replace(langCore, "  # @language\n", "  # @language\n  bad/1: \"%s is bad.\"\n", 1), nil,
			`the language's block writes "bad/1: \"%s is bad.\""`},
		{"no sentence", langCore, func(l *GameLanguage, w map[string]string) { delete(w, "good/1") },
			"good/1 has no sentence: walton.yml gives none"},
		{"no negation", langCore, func(l *GameLanguage, w map[string]string) { delete(l.Negations, "good/1") },
			"good/1 has no negation's sentence"},
		{"a difference that is walton's", langCore, func(l *GameLanguage, w map[string]string) {
			l.Differences[0].Templates["good/1"] = "%s is good."
		}, "good/1's sentence is walton.yml's own"},
		{"a difference recorded twice", langCore, func(l *GameLanguage, w map[string]string) {
			l.Differences = append(l.Differences, LanguageDifference{Reason: "again", Templates: map[string]string{"should_be_performed/1": "%s must be done."}})
		}, "should_be_performed/1 is recorded twice"},
		{"no reason", langCore, func(l *GameLanguage, w map[string]string) { l.Differences[0].Reason = " " }, "gives no reason"},
		{"a difference the game does not play", langCore, func(l *GameLanguage, w map[string]string) {
			l.Differences[0].Templates["bad/1"] = "%s is harmful."
		}, "bad/1 is a predicate the game does not play"},
		{"a negation the game does not play", langCore, func(l *GameLanguage, w map[string]string) { l.Negations["bad/1"] = "%s is not bad." },
			"the negation of bad/1, a predicate the game does not play"},
		{"a sentence's places", langCore, func(l *GameLanguage, w map[string]string) { w["brings_about/2"] = "The action %s brings about good." },
			`brings_about/2's sentence "The action %s brings about good." has 1 places, not 2`},
		{"a negation's places", langCore, func(l *GameLanguage, w map[string]string) { l.Negations["good/1"] = "Nothing is good." },
			"good/1's negation's sentence \"Nothing is good.\" has 0 places, not 1"},
	} {
		l2 := GameLanguage{Negations: map[string]string{}}
		for k, v := range l.Negations {
			l2.Negations[k] = v
		}
		for _, d := range l.Differences {
			m := map[string]string{}
			for k, v := range d.Templates {
				m[k] = v
			}
			l2.Differences = append(l2.Differences, LanguageDifference{Reason: d.Reason, Templates: m})
		}
		w2 := map[string]string{}
		for k, v := range walton {
			w2[k] = v
		}
		if c.edit != nil {
			c.edit(&l2, w2)
		}
		if _, err := fillLanguage(c.core, w2, l2); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", c.name, err, c.want)
		}
	}
}

// A sentence with a double quote or a backslash stays one YAML scalar.
func TestYAMLQuoted(t *testing.T) {
	if got := yamlQuoted(`%s says "%s" \ so`); got != `"%s says \"%s\" \\ so"` {
		t.Errorf("%s", got)
	}
}
