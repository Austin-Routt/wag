package catalogue

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// term is a parsed formula: an optional negation, a name, and arguments.
// A name starting with a capital letter and without arguments is a
// variable.
type term struct {
	neg  bool
	name string
	args []*term
}

var tokenRE = regexp.MustCompile(`\s*(¬|[A-Za-z_][A-Za-z0-9_]*|\(|\)|,|\S)`)

// parseFormula parses a walton.yml formula such as "¬known(S)" or
// "based_on_evidence(asserts(W,S))".
func parseFormula(s string) (*term, error) {
	var toks []string
	for _, m := range tokenRE.FindAllStringSubmatch(s, -1) {
		toks = append(toks, m[1])
	}
	pos := 0
	var parse func() (*term, error)
	parse = func() (*term, error) {
		t := &term{}
		if pos < len(toks) && toks[pos] == "¬" {
			t.neg = true
			pos++
		}
		if pos >= len(toks) || !isName(toks[pos]) {
			return nil, fmt.Errorf("formula %q: a name expected", s)
		}
		t.name = toks[pos]
		pos++
		if pos < len(toks) && toks[pos] == "(" {
			pos++
			for {
				a, err := parse()
				if err != nil {
					return nil, err
				}
				t.args = append(t.args, a)
				if pos < len(toks) && toks[pos] == "," {
					pos++
					continue
				}
				if pos < len(toks) && toks[pos] == ")" {
					pos++
					break
				}
				return nil, fmt.Errorf("formula %q: , or ) expected", s)
			}
		}
		return t, nil
	}
	t, err := parse()
	if err != nil {
		return nil, err
	}
	if pos != len(toks) {
		return nil, fmt.Errorf("formula %q: unexpected %q", s, toks[pos])
	}
	return t, nil
}

func isName(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return r == '_' || unicode.IsLetter(r)
}

func (t *term) variable() bool {
	r, _ := utf8.DecodeRuneInString(t.name)
	return len(t.args) == 0 && unicode.IsUpper(r)
}

// English reads a formula with walton.yml's language templates. A variable
// reads as itself; a nested formula reads in quotes; a negation without a
// template reads "It is not the case that ...". A formula that does not
// parse reads as itself.
func English(formula string, lang map[string]string) string {
	t, err := parseFormula(formula)
	if err != nil {
		return formula
	}
	return sentence(say(t, lang))
}

// say reads a term, without the final period.
func say(t *term, lang map[string]string) string {
	if t.neg {
		key := fmt.Sprintf("¬%s/%d", t.name, len(t.args))
		if tmpl, ok := lang[key]; ok {
			return fill(tmpl, t.args, lang)
		}
		inner := *t
		inner.neg = false
		return "it is not the case that " + lowerFirst(say(&inner, lang))
	}
	if t.variable() {
		return t.name
	}
	tmpl, ok := lang[fmt.Sprintf("%s/%d", t.name, len(t.args))]
	if !ok {
		var args []string
		for _, a := range t.args {
			args = append(args, argument(a, lang))
		}
		if len(args) == 0 {
			return t.name
		}
		return t.name + "(" + strings.Join(args, ", ") + ")"
	}
	return fill(tmpl, t.args, lang)
}

// fill puts the arguments into a template's %s, in order.
func fill(tmpl string, args []*term, lang map[string]string) string {
	out := strings.TrimSpace(tmpl)
	for _, a := range args {
		i := strings.Index(out, "%s")
		if i < 0 {
			break
		}
		out = out[:i] + argument(a, lang) + out[i+2:]
	}
	return strings.TrimSuffix(strings.TrimSpace(out), ".")
}

// argument reads a term in argument position: a variable or an atom as
// itself, anything else in quotes.
func argument(a *term, lang map[string]string) string {
	if !a.neg && len(a.args) == 0 {
		return a.name
	}
	return "“" + say(a, lang) + "”"
}

// sentence capitalizes and ends a reading with a period.
func sentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	s = string(unicode.ToUpper(r)) + s[n:]
	if !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, "?") && !strings.HasSuffix(s, "!") {
		s += "."
	}
	return s
}

var leadingVar = regexp.MustCompile(`^[A-Z][A-Z0-9]*\b`)

// lowerFirst lowers the first letter, unless the reading starts with a
// variable ("W is ...").
func lowerFirst(s string) string {
	if leadingVar.MatchString(s) {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
