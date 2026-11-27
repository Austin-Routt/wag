package catalogue

// The game's sentences (rule (e) of docs/c5-game.md, C5.2, step 4; decision
// D094 (f)): core.yml's language is made, not written. A predicate the game
// plays reads as walton.yml's template, or as the one game/language.yml
// records with the reason it differs; its negation reads as the sentence
// game/language.yml writes for it, since walton.yml gives none and core.yml
// has one for every predicate.

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// GameLanguageFile is GameDir's file of the game's sentences.
const GameLanguageFile = "language.yml"

// languageMarker marks, in core.yml's template, where the language's
// templates go: the template writes none of its own.
const languageMarker = "  # @language"

// GameLanguage is game/language.yml: the game's sentences where they differ
// from walton.yml's, each group with its reason (D094 (f): the game's
// templates stay, each difference recorded), and the sentence of each
// predicate's negation. A key is "predicate/arity", as core.yml's.
type GameLanguage struct {
	Differences []LanguageDifference `yaml:"differences"`
	Negations   map[string]string    `yaml:"negations"`
}

// LanguageDifference is a group of the game's sentences that differ from
// walton.yml's, for one reason.
type LanguageDifference struct {
	Reason    string            `yaml:"reason"`
	Templates map[string]string `yaml:"templates"`
}

// languageEntry is a line of core.yml's language: a key and its sentence.
var languageEntry = regexp.MustCompile(`^  ¬?[^\s#][^:]*: `)

// fillLanguage writes the language's templates at the template's marker:
// those of every predicate core.yml's schemes use, kept schemes included,
// in the order they first appear, each followed by its negation's. It
// refuses a predicate with no sentence, a recorded difference that is
// walton.yml's own sentence or has no reason, a sentence for a predicate
// the game does not play, a sentence whose places are not the predicate's,
// and a template that writes sentences of its own.
func fillLanguage(core string, walton map[string]string, l GameLanguage) (string, error) {
	var errs []string
	fail := func(format string, a ...interface{}) { errs = append(errs, fmt.Sprintf(format, a...)) }

	// The marker, once, in the language's block, which writes no sentence.
	lines := strings.SplitAfter(core, "\n")
	marker, inLang := -1, false
	for i, line := range lines {
		t := strings.TrimRight(line, "\r\n")
		switch {
		case t == "language:":
			inLang = true
			continue
		case t != "" && !strings.HasPrefix(t, " ") && !strings.HasPrefix(t, "#"):
			inLang = false
		}
		if t == languageMarker {
			if !inLang || marker >= 0 {
				fail("the marker %q is not once in the language's block", strings.TrimSpace(languageMarker))
			}
			marker = i
		} else if inLang && languageEntry.MatchString(t) {
			fail("the language's block writes %q: its sentences are walton.yml's or %s/%s's", strings.TrimSpace(t), GameDir, GameLanguageFile)
		}
	}
	if marker < 0 {
		fail("no marker %q in the language's block", strings.TrimSpace(languageMarker))
	}

	// The predicates core.yml's schemes use, in order.
	var c struct {
		Schemes []yaml.MapSlice `yaml:"argument_schemes"`
	}
	if err := yaml.Unmarshal([]byte(core), &c); err != nil {
		return "", fmt.Errorf("core.yml: %v", err)
	}
	var played []string
	plays := map[string]bool{}
	for _, s := range c.Schemes {
		for _, it := range s {
			k, _ := it.Key.(string)
			if !in(coreFormulaKeys[1:], k) {
				continue
			}
			for _, f := range stringList(it.Value) {
				ps, err := predicatesOf(f)
				if err != nil {
					fail("%v of %v: %v", f, value(s, "id"), err)
				}
				for _, p := range ps {
					if !plays[p] {
						plays[p] = true
						played = append(played, p)
					}
				}
			}
		}
	}

	// The recorded differences.
	diff := map[string]string{}
	for i, d := range l.Differences {
		if strings.TrimSpace(d.Reason) == "" {
			fail("%s/%s: the differences' group %d gives no reason", GameDir, GameLanguageFile, i+1)
		}
		for _, p := range sortedKeys(d.Templates) {
			t := d.Templates[p]
			switch _, dup := diff[p]; {
			case dup:
				fail("%s/%s: %s is recorded twice: a difference is recorded once", GameDir, GameLanguageFile, p)
			case !plays[p]:
				fail("%s/%s: %s is a predicate the game does not play", GameDir, GameLanguageFile, p)
			case t == walton[p]:
				fail("%s/%s: %s's sentence is walton.yml's own: it is no difference to record", GameDir, GameLanguageFile, p)
			}
			diff[p] = t
		}
	}
	for _, p := range sortedKeys(l.Negations) {
		if !plays[p] {
			fail("%s/%s: the negation of %s, a predicate the game does not play", GameDir, GameLanguageFile, p)
		}
	}

	// The sentences, in order.
	var b strings.Builder
	for _, p := range played {
		t, ok := diff[p]
		if !ok {
			t = walton[p]
		}
		n := l.Negations[p]
		switch {
		case t == "":
			fail("%s has no sentence: walton.yml gives none, and %s/%s records none", p, GameDir, GameLanguageFile)
		case n == "":
			fail("%s has no negation's sentence in %s/%s: core.yml has one for every predicate", p, GameDir, GameLanguageFile)
		}
		arity := p[strings.LastIndex(p, "/")+1:]
		for _, s := range []struct{ what, text string }{{"sentence", t}, {"negation's sentence", n}} {
			if s.text != "" && fmt.Sprint(strings.Count(s.text, "%s")) != arity {
				fail("%s's %s %q has %d places, not %s", p, s.what, s.text, strings.Count(s.text, "%s"), arity)
			}
		}
		fmt.Fprintf(&b, "  %s: %s\n  ¬%s: %s\n", p, yamlQuoted(t), p, yamlQuoted(n))
	}
	if len(errs) > 0 {
		return "", fmt.Errorf("%s", strings.Join(errs, "\n  "))
	}
	lines[marker] = b.String()
	return strings.Join(lines, ""), nil
}

// predicatesOf lists the predicates a formula uses, as "name/arity", the
// outermost first: a statement inside a statement has its own sentence. A
// variable (capitalized) is none, nor is a constant inside a statement.
func predicatesOf(f string) ([]string, error) {
	s := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(f), "¬"))
	var out []string
	var term func(top bool) error
	term = func(top bool) error {
		s = strings.TrimLeft(s, " ")
		m := identifier.FindStringIndex(s)
		if m == nil || m[0] != 0 {
			return fmt.Errorf("a term expected at %q", s)
		}
		name := s[:m[1]]
		s = s[m[1]:]
		if !strings.HasPrefix(s, "(") {
			if top && !(name[0] >= 'A' && name[0] <= 'Z') {
				out = append(out, name+"/0")
			}
			return nil
		}
		at := len(out)
		out = append(out, "")
		s = s[1:]
		n := 0
		for {
			if err := term(false); err != nil {
				return err
			}
			n++
			s = strings.TrimLeft(s, " ")
			switch {
			case strings.HasPrefix(s, ","):
				s = s[1:]
			case strings.HasPrefix(s, ")"):
				s = s[1:]
				out[at] = fmt.Sprintf("%s/%d", name, n)
				return nil
			default:
				return fmt.Errorf("\",\" or \")\" expected at %q", s)
			}
		}
	}
	if err := term(true); err != nil {
		return nil, err
	}
	if strings.TrimSpace(s) != "" {
		return nil, fmt.Errorf("%q after the formula", s)
	}
	return out, nil
}

// recordedLanguage is game/language.yml recorded from core.yml's text: its
// sentences that differ from walton.yml's, in one group whose reason is to
// be given, and its negations' sentences. A catalogue made from the game's
// files gets it, as it gets GameDir's other files.
func recordedLanguage(core string, walton map[string]string) ([]byte, error) {
	l := GameLanguage{Negations: map[string]string{}}
	diff := map[string]string{}
	inLang := false
	for _, line := range strings.Split(core, "\n") {
		switch {
		case line == "language:":
			inLang = true
			continue
		case line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#"):
			inLang = false
		}
		if !inLang || !languageEntry.MatchString(line) {
			continue
		}
		var kv map[string]string
		if err := yaml.Unmarshal([]byte(strings.TrimSpace(line)), &kv); err != nil {
			return nil, fmt.Errorf("core.yml's language: %q: %v", line, err)
		}
		for k, v := range kv {
			if p, neg := strings.CutPrefix(k, "¬"); neg {
				l.Negations[p] = v
			} else if v != walton[k] {
				diff[k] = v
			}
		}
	}
	if len(diff) > 0 {
		l.Differences = []LanguageDifference{{Reason: "Recorded from the game's core.yml: the sentences that differ from walton.yml's. Give each group its reason (D094 (f)).", Templates: diff}}
	}
	b, err := yaml.Marshal(l)
	if err != nil {
		return nil, err
	}
	return append([]byte(languageHeader), b...), nil
}

// languageHeader heads game/language.yml.
const languageHeader = `# The game's sentences (D094 (f); rule (e) of docs/c5-game.md, C5.2): the
# templates of core.yml's language, where they differ from walton.yml's,
# each group with its reason, and the sentence of each predicate's negation,
# which walton.yml does not give. A predicate the game plays and this file
# does not record reads as walton.yml's template. wag-cli --catalogue-gen
# reads this file and does not rewrite it.
`

// yamlQuoted writes a sentence as a YAML double-quoted scalar.
func yamlQuoted(s string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
}

// languageTemplate gives back, from core.yml's text, the template's
// language block: the sentences, which are made, replaced by the marker
// where the first of them is.
func languageTemplate(text string) string {
	lines := strings.SplitAfter(text, "\n")
	var b strings.Builder
	inLang, marked := false, false
	for _, line := range lines {
		t := strings.TrimRight(line, "\r\n")
		switch {
		case t == "language:":
			inLang = true
		case t != "" && !strings.HasPrefix(t, " ") && !strings.HasPrefix(t, "#"):
			inLang = false
		}
		if inLang && languageEntry.MatchString(t) {
			if !marked {
				b.WriteString(languageMarker + "\n")
				marked = true
			}
			continue
		}
		b.WriteString(line)
	}
	return b.String()
}
