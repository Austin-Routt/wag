package catalogue

// The rules of the game's questions (docs/c5-game.md, C5.2, step 4, applied
// with C5.2's switch): one id is one question, asked at one formula and one
// kind of locus by every scheme that asks it, with one text (the refinement
// of D094; D095 (1) for a question answer schemes back); its record's
// caes_mapping is derived from that locus, and its text from its entries'
// paraphrase (D009); a role scheme's record agrees with its formulas.

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// askedAt is where a scheme asks a question: the ask edge's locus, the
// formula asked (the pattern, or the scheme's conclusion for a question
// about the conclusion; none for one about the inference), with the
// scheme's variables in braces, and the question's text there: the
// scheme's entry's paraphrase with its variables braced.
type askedAt struct {
	Scheme, Locus, Formula, Text string
}

// questionFaults reports the questions that break the rule that one id is
// one question: every scheme that asks it asks it (a) at one formula, so
// that an answer scheme is never offered where it cannot conclude the
// target (D095 (1)), and (b) at one kind of locus, so that its record's
// caes_mapping, derived from that locus, is the effect of every edge
// (D009); and (c) with one text, its entries' one paraphrase. asked maps a
// question id to where each scheme asks it.
func questionFaults(asked map[string][]askedAt) []string {
	var faults []string
	ids := make([]string, 0, len(asked))
	for id := range asked {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		loci, formulas, texts := map[string][]string{}, map[string][]string{}, map[string][]string{}
		for _, e := range asked[id] {
			loci[e.Locus] = append(loci[e.Locus], e.Scheme)
			formulas[e.Formula] = append(formulas[e.Formula], e.Scheme)
			texts[e.Text] = append(texts[e.Text], e.Scheme)
		}
		if len(loci) > 1 {
			faults = append(faults, fmt.Sprintf("%s is asked at %s: one id is one question, asked at one kind of locus, so that its effect is the same on every edge (split it, one id for each)",
				id, describe(loci)))
		}
		if len(formulas) > 1 {
			faults = append(faults, fmt.Sprintf("%s is asked at %s: one id is one question, asked at one formula wherever it is asked (the refinement of D094; D095 (1)): split it, the scheme whose book question named it keeping the id",
				id, describe(formulas)))
		}
		if len(texts) > 1 {
			faults = append(faults, fmt.Sprintf("%s reads %s: one id is one question, with one text: give its entries one paraphrase",
				id, quoted(texts)))
		}
	}
	return faults
}

// describe lists the values of a key (a locus or a formula) with the
// schemes that have each.
func describe(by map[string][]string) string {
	keys := make([]string, 0, len(by))
	for k := range by {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		name := k
		if name == "" {
			name = "no formula"
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", name, strings.Join(by[k], ", ")))
	}
	return strings.Join(parts, " and ")
}

// quoted lists texts, quoted, with the schemes that have each.
func quoted(by map[string][]string) string {
	q := map[string][]string{}
	for k, v := range by {
		q[fmt.Sprintf("%q", k)] = v
	}
	return describe(q)
}

// answerTypes is the caes_mapping a question's locus gives it: what a
// positive answer does, what a negative one does, and what it is about.
func answerTypes(locus string) (positive, negative, target string) {
	switch locus {
	case "premise":
		return "support", "undermining", "premise"
	case "conclusion":
		return "support", "rebuttal", "conclusion"
	default: // an exception, or the inference
		return "support", "undercutter", "argument"
	}
}

// caesMapping is the caes_mapping of a question asked at a locus, as the
// NASCQ writes it.
func caesMapping(locus string) yaml.MapSlice {
	p, n, t := answerTypes(locus)
	return yaml.MapSlice{{Key: "positive_answer_type", Value: p}, {Key: "negative_answer_type", Value: n}, {Key: "target", Value: t}}
}

// The keys of a question's record in GameDir's questions.yml. The record's
// text and caes_mapping are not among them: the generator makes them, from
// the entries' paraphrase and the question's locus.
var questionKeys = []string{"name", "semantic_type", "answer_en", "refuse_other"}

// questionRecord is a question's NASCQ record: GameDir's, with its text
// after its semantic type and its caes_mapping after its answer words, as
// critical_questions.json writes them. It refuses a record that lacks a
// key the game needs or has one it does not know, its own text or
// caes_mapping included.
func questionRecord(id string, rec yaml.MapSlice, text, locus string) (yaml.MapSlice, error) {
	var out yaml.MapSlice
	have := map[string]bool{}
	for _, it := range rec {
		k, _ := it.Key.(string)
		if !in(questionKeys, k) {
			if k == "question_template_en" || k == "caes_mapping" {
				return nil, fmt.Errorf("%s: its record gives %s, which the generator makes (from its entries' paraphrase, from its locus)", id, k)
			}
			return nil, fmt.Errorf("%s: its record has the key %v, not one of %s", id, it.Key, strings.Join(questionKeys, ", "))
		}
		have[k] = true
		out = append(out, it)
		switch k {
		case "semantic_type":
			out = append(out, yaml.MapItem{Key: "question_template_en", Value: text})
		case "answer_en":
			if !answerWords(it.Value) {
				return nil, fmt.Errorf("%s: its answer_en gives no words: an \"against\" and a \"for\", the literal answers that attack and defend the argument", id)
			}
			out = append(out, yaml.MapItem{Key: "caes_mapping", Value: caesMapping(locus)})
		}
	}
	for _, k := range questionKeys[:3] {
		if !have[k] {
			return nil, fmt.Errorf("%s: its record has no %s", id, k)
		}
	}
	return out, nil
}

// answerWords tells whether a record's answer_en gives both literal answers.
func answerWords(v interface{}) bool {
	m, ok := v.(yaml.MapSlice)
	if !ok || len(m) != 2 {
		return false
	}
	for _, k := range []string{"against", "for"} {
		if s, ok := value(m, k).(string); !ok || strings.TrimSpace(s) == "" {
			return false
		}
	}
	return true
}

// paraphraseFault reports why a paraphrase cannot be a game question's
// text, or "": the text is the paraphrase with the scheme's variables in
// braces, so it has none of its own.
func paraphraseFault(p string) string {
	switch {
	case strings.TrimSpace(p) == "":
		return "it has no paraphrase, which the game's text is made from (D009)"
	case strings.ContainsAny(p, "{}"):
		return "its paraphrase has braces: the generator braces the scheme's variables in it"
	}
	return ""
}

// sameFormula tells whether two formulas are one formula, each with its
// own scheme's variables (in braces): the same but for a consistent
// renaming of the variables.
func sameFormula(a, b string) bool {
	ta, tb := formulaTokens.FindAllString(a, -1), formulaTokens.FindAllString(b, -1)
	if len(ta) != len(tb) {
		return false
	}
	ab, ba := map[string]string{}, map[string]string{}
	for i := range ta {
		x, y := ta[i], tb[i]
		vx, vy := strings.HasPrefix(x, "{"), strings.HasPrefix(y, "{")
		switch {
		case vx != vy:
			return false
		case !vx:
			if x != y {
				return false
			}
		default:
			if o, ok := ab[x]; ok && o != y {
				return false
			}
			if o, ok := ba[y]; ok && o != x {
				return false
			}
			ab[x], ba[y] = y, x
		}
	}
	return true
}

// formulaTokens are a braced formula's tokens: a braced variable, a name,
// or a single character.
var formulaTokens = regexp.MustCompile(`\{[^{}]*\}|[A-Za-z_][A-Za-z0-9_]*|\S`)

// The roles a game scheme's NASCQ record can name (decision D094 (g)):
// none, an argument of its own.
var gameRoles = []string{"answer", "modal", "bridge"}

// roleFaults reports where a role scheme's record disagrees with the
// formulas the game plays (rule (d)): an answer scheme concludes the
// formula of each question it backs, asked at an exception or a premise
// (D075); a modal variant's and a bridge's weighs_as is a game scheme; a
// bridge's entry_premise is one of its premises. asked maps a question id
// to where each scheme asks it; backs maps a game scheme to the questions
// it backs.
func roleFaults(schemes []*gameIn, byID map[string]*gameIn, asked map[string][]askedAt, backs map[string][]string) []string {
	var faults []string
	for _, g := range schemes {
		rec := g.entry.Game.NASCQ
		role, _ := value(rec, "role").(string)
		if value(rec, "role") != nil && !in(gameRoles, role) {
			faults = append(faults, fmt.Sprintf("%s: its role %v is not one of %s", g.id, value(rec, "role"), strings.Join(gameRoles, ", ")))
			continue
		}
		switch role {
		case "answer":
			concl := ""
			if c := g.roles["conclusions"]; len(c) == 1 {
				concl = g.rebrace(c[0])
			}
			for _, cq := range backs[g.id] {
				for _, a := range asked[cq] {
					switch {
					case a.Locus != "exception" && a.Locus != "premise":
						faults = append(faults, fmt.Sprintf("%s backs %s, which %s asks at its %s: an answer scheme answers an exception or an assumption (D075)", g.id, cq, a.Scheme, a.Locus))
					case !sameFormula(concl, a.Formula):
						faults = append(faults, fmt.Sprintf("%s concludes %s, not %s, which %s asks %s at: an answer scheme concludes what it answers (D075, D095 (1))", g.id, orNone(concl), a.Formula, a.Scheme, cq))
					}
				}
			}
		case "modal", "bridge":
			w, _ := value(rec, "weighs_as").(string)
			if byID[w] == nil {
				faults = append(faults, fmt.Sprintf("%s weighs as %q, which is no game scheme", g.id, w))
			}
			if role == "bridge" {
				p, _ := value(rec, "entry_premise").(string)
				found := false
				for _, x := range g.roles["premises"] {
					found = found || g.rebrace(x) == p
				}
				if !found {
					faults = append(faults, fmt.Sprintf("%s's entry premise %q is none of its premises", g.id, p))
				}
			}
		}
	}
	return faults
}

func orNone(s string) string {
	if s == "" {
		return "no one formula"
	}
	return s
}
