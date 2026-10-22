package catalogue

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// A proposal (docs/proposals/*.yml; docs/process.md §5) is the evidence for
// a decision that changes walton.yml, prepared before the change is made:
// the schemes as they would be, the templates they need, and examples bound
// to their variables. Each example is rendered in English, so that a person
// can judge whether the instantiated scheme says something sensible, as an
// argument, as an answer in the game, and under each critical question.
type Proposal struct {
	Decisions []string            `yaml:"decisions"`
	Title     string              `yaml:"title"`
	Summary   string              `yaml:"summary"`
	Schemes   []Scheme            `yaml:"schemes"`  // as they would be in walton.yml
	Language  map[string]string   `yaml:"language"` // templates added or changed
	Effects   []string            `yaml:"effects,omitempty"`
	Examples  []ProposalExample   `yaml:"examples"`
	Questions map[string][]PQuery `yaml:"questions,omitempty"` // by scheme: the questions after the change
}

// PQuery is a critical question of a proposed scheme: its label (a book
// label, game/…, walton/…), what it asks, and the formula it targets.
type PQuery struct {
	ID     string `yaml:"id"`
	Asks   string `yaml:"asks"`
	Target string `yaml:"target"`
}

// ProposalExample is a text coded with a scheme, its variables bound to
// phrases.
type ProposalExample struct {
	Scheme   string            `yaml:"scheme"`
	Title    string            `yaml:"title"`
	Text     string            `yaml:"text"`
	Kialo    string            `yaml:"kialo,omitempty"`
	Invented bool              `yaml:"invented,omitempty"` // our own text, not a quotation
	Use      string            `yaml:"use"`                // argument, answer or near-miss
	Answers  string            `yaml:"answers,omitempty"`  // for an answer: the question and the statement
	Form     string            `yaml:"form,omitempty"`     // negated, when the scheme's "(false)" form
	Of       string            `yaml:"near_miss_of,omitempty"`
	Bindings map[string]string `yaml:"bindings"`
	Attacks  []Attack          `yaml:"attacks,omitempty"`
	Note     string            `yaml:"note,omitempty"`
}

// Attack is a critical question pressed against an example.
type Attack struct {
	Question string `yaml:"question"`
	Text     string `yaml:"text"`
}

var exampleUses = []string{"argument", "answer", "near-miss"}

// LoadProposal reads a proposal; unknown fields are errors.
func LoadProposal(path string) (*Proposal, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Proposal
	if err := yaml.UnmarshalStrict(b, &p); err != nil {
		return nil, fmt.Errorf("%s: %v", path, err)
	}
	return &p, nil
}

// proposalSchemes is walton.yml's schemes with the proposal's in place,
// and its language with the proposal's templates.
func proposalSchemes(p *Proposal, src *Sources) (map[string]Scheme, map[string]string) {
	schemes := map[string]Scheme{}
	for _, w := range src.Walton {
		schemes[w.ID] = w
	}
	for _, s := range p.Schemes {
		schemes[s.ID] = s
	}
	lang := map[string]string{}
	for k, v := range src.Language {
		lang[k] = v
	}
	for k, v := range p.Language {
		lang[k] = strings.TrimSpace(v)
	}
	return schemes, lang
}

func (s Scheme) formulas() []string {
	var out []string
	for _, l := range [][]string{s.Premises, s.Assumptions, s.Exceptions, s.Conclusions} {
		out = append(out, l...)
	}
	return out
}

// CheckProposal checks a proposal against walton.yml:
//   - each proposed scheme declares exactly the variables its formulas use,
//     and every predicate it uses has a template;
//   - each example names a scheme, binds exactly its variables, and presses
//     only questions the scheme has;
//   - each proposed scheme has the minimum set of examples (docs/process.md
//     §5): two arguments, an answer if it can be one, a near miss, and every
//     question pressed at least once (warnings).
func CheckProposal(p *Proposal, src *Sources) []Finding {
	var out []Finding
	bad := func(f string, a ...any) { out = append(out, Finding{Error: true, Msg: fmt.Sprintf(f, a...)}) }
	warn := func(f string, a ...any) { out = append(out, Finding{Msg: fmt.Sprintf(f, a...)}) }
	schemes, lang := proposalSchemes(p, src)
	if len(p.Decisions) == 0 || p.Title == "" {
		bad("a proposal names its decisions and has a title")
	}
	for _, s := range p.Schemes {
		used := map[string]bool{}
		for _, f := range s.formulas() {
			t, err := parseFormula(f)
			if err != nil {
				bad("%s: %v", s.ID, err)
				continue
			}
			t.walk(func(x *term) {
				if x.variable() {
					used[x.name] = true
					return
				}
				key := fmt.Sprintf("%s/%d", x.name, len(x.args))
				if _, ok := lang[key]; !ok {
					bad("%s: no template for %s", s.ID, key)
				}
			})
		}
		declared := map[string]bool{}
		for _, v := range s.Variables {
			declared[v] = true
			if !used[v] {
				bad("%s: variable %s is declared but used in no formula", s.ID, v)
			}
		}
		for v := range used {
			if !declared[v] {
				bad("%s: variable %s is used but not declared", s.ID, v)
			}
		}
		for _, v := range s.Variables {
			if !inFormulas(v, s.Premises) {
				bad("%s: variable %s occurs in no premise (the Carneades validator requires it)", s.ID, v)
			}
		}
		for _, q := range p.Questions[s.ID] {
			if !in(s.formulas(), q.Target) {
				bad("%s: question %s targets %s, not a formula of the scheme", s.ID, q.ID, q.Target)
			}
		}
	}
	type tally struct{ arguments, answers, nearMisses int }
	count := map[string]*tally{}
	pressed := map[string]map[string]bool{}
	for _, s := range p.Schemes {
		count[s.ID] = &tally{}
		pressed[s.ID] = map[string]bool{}
	}
	for i, x := range p.Examples {
		where := fmt.Sprintf("example %d (%s)", i+1, x.Title)
		s, ok := schemes[x.Scheme]
		if !ok {
			bad("%s: no scheme %q", where, x.Scheme)
			continue
		}
		if x.Text == "" || x.Title == "" {
			bad("%s: a title and a text are needed", where)
		}
		if !in(exampleUses, x.Use) {
			bad("%s: use %q is not one of %s", where, x.Use, strings.Join(exampleUses, ", "))
		}
		if x.Use == "answer" && x.Answers == "" {
			bad("%s: an answer says which question and statement it answers", where)
		}
		if x.Use == "near-miss" && x.Of == "" {
			bad("%s: a near miss names the scheme it resembles (near_miss_of)", where)
		}
		for _, v := range s.Variables {
			if strings.TrimSpace(x.Bindings[v]) == "" {
				bad("%s: variable %s is not bound", where, v)
			}
		}
		for v := range x.Bindings {
			if !in(s.Variables, v) {
				bad("%s: %s is not a variable of %s", where, v, s.ID)
			}
		}
		qs := p.Questions[x.Scheme]
		for _, a := range x.Attacks {
			if questionByID(qs, a.Question) == nil {
				bad("%s: %s is not a question of %s in this proposal", where, a.Question, x.Scheme)
				continue
			}
			if pressed[x.Scheme] != nil {
				pressed[x.Scheme][a.Question] = true
			}
		}
		if c := count[x.Scheme]; c != nil {
			switch x.Use {
			case "argument":
				c.arguments++
			case "answer":
				c.answers++
			}
		}
		if c := count[x.Of]; c != nil && x.Use == "near-miss" {
			c.nearMisses++
		}
	}
	for _, s := range p.Schemes {
		c := count[s.ID]
		if c.arguments < 2 {
			warn("%s: %d example(s) of use as an argument; the minimum is two", s.ID, c.arguments)
		}
		if c.answers == 0 && answerable(s) {
			warn("%s: concludes a bare proposition, so it can be an answer in the game, but no example shows it", s.ID)
		}
		if c.nearMisses == 0 {
			warn("%s: no near miss", s.ID)
		}
		for _, q := range p.Questions[s.ID] {
			if !pressed[s.ID][q.ID] {
				warn("%s: question %s is pressed in no example", s.ID, q.ID)
			}
		}
	}
	return out
}

func questionByID(qs []PQuery, id string) *PQuery {
	for i := range qs {
		if qs[i].ID == id {
			return &qs[i]
		}
	}
	return nil
}

func inFormulas(v string, fs []string) bool {
	for _, f := range fs {
		t, err := parseFormula(f)
		if err != nil {
			continue
		}
		found := false
		t.walk(func(x *term) {
			if x.variable() && x.name == v {
				found = true
			}
		})
		if found {
			return true
		}
	}
	return false
}

// answerable tells whether a scheme concludes a bare proposition, which the
// game binds to an answer.
func answerable(s Scheme) bool {
	for _, c := range s.Conclusions {
		if t, err := parseFormula(c); err == nil && t.variable() {
			return true
		}
	}
	return false
}

func (t *term) walk(f func(*term)) {
	f(t)
	for _, a := range t.args {
		a.walk(f)
	}
}

// ProposalWalton returns walton.yml (raw) with the proposal applied: its
// schemes replace the blocks with the same id or are added at the end, and
// its templates replace the lines with the same key or are added to the
// language. It edits the text, so that everything else stays byte for byte
// (a round trip through YAML would read variables such as Y as booleans).
// The result is for the cleaning's gate (walton-check), not for
// committing: the change is made through walton-clean/tools/edits.py once
// it is decided.
func ProposalWalton(raw []byte, p *Proposal) ([]byte, error) {
	lines := strings.Split(string(raw), "\n")
	top := func(i int) bool { return i < len(lines) && lines[i] != "" && lines[i][0] != ' ' && lines[i][0] != '#' }
	section := func(key string) (int, int) { // the lines of a top-level key's body
		for i := range lines {
			if top(i) && strings.HasPrefix(lines[i], key+":") {
				j := i + 1
				for j < len(lines) && !top(j) {
					j++
				}
				return i + 1, j
			}
		}
		return -1, -1
	}
	keys := make([]string, 0, len(p.Language))
	for k := range p.Language {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		// One line, double-quoted, as walton.yml writes its templates (a
		// marshalled scalar may wrap onto a second line).
		line := "  " + k + ": " + strconv.Quote(strings.TrimSpace(p.Language[k]))
		from, to := section("language")
		if from < 0 {
			return nil, fmt.Errorf("walton.yml has no language")
		}
		set, last := false, from
		for i := from; i < to; i++ {
			if strings.HasPrefix(lines[i], "  "+k+":") {
				lines[i], set = line, true
			}
			if strings.TrimSpace(lines[i]) != "" {
				last = i
			}
		}
		if !set {
			lines = append(lines[:last+1], append([]string{line}, lines[last+1:]...)...)
		}
	}
	for _, s := range p.Schemes {
		b, err := yaml.Marshal([]Scheme{s})
		if err != nil {
			return nil, err
		}
		var block []string
		for _, l := range strings.Split(strings.TrimRight(string(b), "\n"), "\n") {
			block = append(block, "  "+l)
		}
		from, to := section("argument_schemes")
		if from < 0 {
			return nil, fmt.Errorf("walton.yml has no argument_schemes")
		}
		start := -1
		for i := from; i < to; i++ {
			if strings.TrimRight(lines[i], " ") == "  - id: "+s.ID {
				start = i
				break
			}
		}
		if start < 0 {
			end := to
			for end > from && strings.TrimSpace(lines[end-1]) == "" {
				end--
			}
			lines = append(lines[:end], append(block, lines[end:]...)...)
			continue
		}
		end := start + 1
		for end < to && !strings.HasPrefix(lines[end], "  - ") {
			end++
		}
		for end > start+1 && strings.TrimSpace(lines[end-1]) == "" {
			end--
		}
		lines = append(lines[:start], append(block, lines[end:]...)...)
	}
	return []byte(strings.Join(lines, "\n")), nil
}

// ProposalView renders a proposal for review: the proposed schemes, the new
// templates, the questions, and every example in English with its bindings.
func ProposalView(p *Proposal, src *Sources, file string) []byte {
	schemes, lang := proposalSchemes(p, src)
	var b bytes.Buffer
	fmt.Fprintf(&b, "# Proposal %s: %s\n\n", strings.Join(p.Decisions, ", "), p.Title)
	fmt.Fprintf(&b, "Generated by `wag-cli --proposal` from [`%s`](%s); edit that file, not this one. The decisions are in [`decisions.md`](../decisions.md); how proposals work is in [`process.md`](../process.md) §5.\n\n", file, file)
	if p.Summary != "" {
		b.WriteString(strings.TrimSpace(p.Summary) + "\n\n")
	}
	b.WriteString("## The schemes as proposed\n\n")
	for _, s := range p.Schemes {
		fmt.Fprintf(&b, "### `%s`: %s\n\n", s.ID, s.Meta.Title)
		if old, ok := findWalton(src, s.ID); ok {
			b.WriteString("| Role | Now | Proposed |\n|---|---|---|\n")
			for _, r := range []struct {
				name     string
				old, new []string
			}{
				{"variables", old.Variables, s.Variables}, {"premises", old.Premises, s.Premises},
				{"assumptions", old.Assumptions, s.Assumptions}, {"exceptions", old.Exceptions, s.Exceptions},
				{"conclusion", old.Conclusions, s.Conclusions},
			} {
				if len(r.old) == 0 && len(r.new) == 0 {
					continue
				}
				fmt.Fprintf(&b, "| %s | %s | %s |\n", r.name, code(r.old), code(r.new))
			}
		} else {
			b.WriteString("A new scheme.\n\n| Role | Proposed |\n|---|---|\n")
			fmt.Fprintf(&b, "| variables | %s |\n| premises | %s |\n", code(s.Variables), code(s.Premises))
			if len(s.Assumptions) > 0 {
				fmt.Fprintf(&b, "| assumptions | %s |\n", code(s.Assumptions))
			}
			if len(s.Exceptions) > 0 {
				fmt.Fprintf(&b, "| exceptions | %s |\n", code(s.Exceptions))
			}
			fmt.Fprintf(&b, "| conclusion | %s |\n", code(s.Conclusions))
		}
		b.WriteString("\n")
		if qs := p.Questions[s.ID]; len(qs) > 0 {
			b.WriteString("| Question | Asks | Targets |\n|---|---|---|\n")
			for _, q := range qs {
				fmt.Fprintf(&b, "| %s | %s | `%s` (%s) |\n", q.ID, cell(q.Asks), q.Target, roleOf(s, q.Target))
			}
			b.WriteString("\n")
		}
	}
	if len(p.Language) > 0 {
		b.WriteString("**Templates added or changed:**\n\n")
		keys := make([]string, 0, len(p.Language))
		for k := range p.Language {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "- `%s`: “%s”\n", k, strings.TrimSpace(p.Language[k]))
		}
		b.WriteString("\n")
	}
	if len(p.Effects) > 0 {
		b.WriteString("## What else changes if it is decided\n\n")
		for _, e := range p.Effects {
			fmt.Fprintf(&b, "- %s\n", strings.Join(strings.Fields(e), " "))
		}
		b.WriteString("\n")
	}
	b.WriteString("## The examples, rendered\n\n")
	b.WriteString("Each example binds the scheme's variables to phrases; every formula is then read with `walton.yml`'s templates and the proposal's. Read each line as the game would show it, and ask whether it says something true and sensible.\n\n")
	for i, x := range p.Examples {
		s := schemes[x.Scheme]
		fmt.Fprintf(&b, "### %d. %s\n\n", i+1, x.Title)
		var tags []string
		switch x.Use {
		case "answer":
			tags = append(tags, "**as an answer** in the game: "+strings.Join(strings.Fields(x.Answers), " "))
		case "near-miss":
			tags = append(tags, fmt.Sprintf("**a near miss** of `%s`: it is `%s`", x.Of, x.Scheme))
		default:
			tags = append(tags, fmt.Sprintf("**as an argument**, with `%s`", x.Scheme))
		}
		if x.Form == "negated" {
			tags = append(tags, "the negated (\"false\") form")
		}
		switch {
		case x.Kialo != "":
			tags = append(tags, "Kialo "+x.Kialo)
		case x.Invented:
			tags = append(tags, "our own text")
		}
		fmt.Fprintf(&b, "%s.\n\n> %s\n\n", strings.Join(tags, "; "), strings.Join(strings.Fields(x.Text), " "))
		var vars []string
		for _, v := range s.Variables {
			vars = append(vars, fmt.Sprintf("%s = “%s”", v, x.Bindings[v]))
		}
		fmt.Fprintf(&b, "Bindings: %s.\n\n", strings.Join(vars, "; "))
		for _, f := range s.Premises {
			fmt.Fprintf(&b, "- %s\n", EnglishWith(f, lang, x.Bindings))
		}
		for _, f := range s.Assumptions {
			fmt.Fprintf(&b, "- *Presumed:* %s\n", EnglishWith(f, lang, x.Bindings))
		}
		for _, f := range s.Exceptions {
			fmt.Fprintf(&b, "- *Unless:* %s\n", EnglishWith(f, lang, x.Bindings))
		}
		for _, f := range s.Conclusions {
			// Some templates open with "Therefore,"; the label says it.
			fmt.Fprintf(&b, "- **So:** %s\n", sentence(strings.TrimPrefix(EnglishWith(f, lang, x.Bindings), "Therefore, ")))
		}
		if len(x.Attacks) > 0 {
			b.WriteString("\nPressed with its questions:\n\n")
			for _, a := range x.Attacks {
				q := questionByID(p.Questions[x.Scheme], a.Question)
				if q == nil {
					continue
				}
				fmt.Fprintf(&b, "- **%s**, at the %s “%s”: %s\n", a.Question, roleOf(s, q.Target), strings.TrimSuffix(EnglishWith(q.Target, lang, x.Bindings), "."), strings.Join(strings.Fields(a.Text), " "))
			}
		}
		if x.Note != "" {
			fmt.Fprintf(&b, "\n*Note:* %s\n", strings.Join(strings.Fields(x.Note), " "))
		}
		b.WriteString("\n")
	}
	return b.Bytes()
}

func findWalton(src *Sources, id string) (Scheme, bool) {
	for _, w := range src.Walton {
		if w.ID == id {
			return w, true
		}
	}
	return Scheme{}, false
}

func roleOf(s Scheme, f string) string {
	switch {
	case in(s.Premises, f):
		return "premise"
	case in(s.Assumptions, f):
		return "assumption"
	case in(s.Exceptions, f):
		return "exception"
	case in(s.Conclusions, f):
		return "conclusion"
	}
	return "?"
}

func code(l []string) string {
	if len(l) == 0 {
		return "–"
	}
	var out []string
	for _, x := range l {
		out = append(out, "`"+x+"`")
	}
	return strings.Join(out, ", ")
}

// ProposalViewPath is where a proposal's view is written: beside it, .md.
func ProposalViewPath(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path)) + ".md"
}
