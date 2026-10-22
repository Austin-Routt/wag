package catalogue

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	Decisions []string `yaml:"decisions"`
	Title     string   `yaml:"title"`
	// Applied records that the proposal has been made in walton.yml (a date
	// and where); its view then says that its "Now" columns, which show
	// walton.yml as it is, equal the proposed ones.
	Applied   string              `yaml:"applied,omitempty"`
	Summary   string              `yaml:"summary"`
	Schemes   []Scheme            `yaml:"schemes"`  // as they would be in walton.yml
	Language  map[string]string   `yaml:"language"` // templates added or changed
	Effects   []string            `yaml:"effects,omitempty"`
	Examples  []ProposalExample   `yaml:"examples"`
	Questions map[string][]PQuery `yaml:"questions,omitempty"` // by scheme: the questions after the change
	// Candidates are sketched beside the proposal, never applied (below).
	Candidates *Candidates `yaml:"candidates,omitempty"`
}

// Candidates are schemes sketched beside a proposal but not proposed. They
// test that the proposal's vocabulary extends to neighbouring patterns, so
// that the proposed schemes need not be reopened when those are added. They
// are checked and rendered like the proposed schemes, with their own
// templates and questions, and never applied to walton.yml.
type Candidates struct {
	Note      string              `yaml:"note"`
	Schemes   []Scheme            `yaml:"schemes"`
	Language  map[string]string   `yaml:"language,omitempty"`
	Questions map[string][]PQuery `yaml:"questions,omitempty"`
}

// Challenge is the game's challenge of a premise (game-rules.md §2): every
// premise can be challenged, so a premise needs no question of its own. An
// example presses it with the premise as its target.
const Challenge = "CQ:premise_acceptable"

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

// Attack is a critical question pressed against an example, or a challenge
// of one of its premises (Question is Challenge, Target the premise).
type Attack struct {
	Question string `yaml:"question"`
	Target   string `yaml:"target,omitempty"`
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

// proposalSchemes is walton.yml's schemes with the proposal's in place and
// its candidates added, and its language with the proposal's and the
// candidates' templates.
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
	if c := p.Candidates; c != nil {
		for _, s := range c.Schemes {
			if _, ok := schemes[s.ID]; !ok {
				schemes[s.ID] = s
			}
		}
		for k, v := range c.Language {
			if _, ok := lang[k]; !ok {
				lang[k] = strings.TrimSpace(v)
			}
		}
	}
	return schemes, lang
}

// questions is a proposed or candidate scheme's questions: a proposed
// scheme's from the proposal's questions, a candidate's from the candidates'.
func (p *Proposal) questions(id string) []PQuery {
	if in(schemeIDs(p.Schemes), id) {
		return p.Questions[id]
	}
	if p.candidate(id) {
		return p.Candidates.Questions[id]
	}
	return nil
}

func (p *Proposal) candidate(id string) bool {
	if p.Candidates == nil {
		return false
	}
	for _, s := range p.Candidates.Schemes {
		if s.ID == id {
			return true
		}
	}
	return false
}

func (s Scheme) formulas() []string {
	var out []string
	for _, l := range [][]string{s.Premises, s.Assumptions, s.Exceptions, s.Conclusions} {
		out = append(out, l...)
	}
	return out
}

// CheckProposal checks a proposal against walton.yml:
//   - each proposed or candidate scheme declares exactly the variables its
//     formulas use, and every predicate it uses has a template;
//   - questions are given only for proposed (or candidate) schemes, under
//     their own heading, with distinct ids;
//   - each question targets a formula of its scheme, and a walton/ question
//     an assumption or exception whose predicate its id names, as in the
//     catalogue (premises are challenged instead); a book label (7/CQ1) is
//     one the compendium prints for that entry, an entry linked to the
//     scheme (or under which its questions are printed);
//   - a candidate is a new scheme, neither in walton.yml nor proposed;
//   - each example names a scheme, binds exactly its variables, presses only
//     questions the scheme has, and challenges only a premise or assumption
//     that none of the scheme's questions targets (the game asks the question
//     instead, game-rules.md §2);
//   - each proposed scheme has the minimum set of examples (docs/process.md
//     §5): two arguments, an answer if it can be one, a near miss, and every
//     question pressed at least once (warnings); a candidate needs one
//     example.
func CheckProposal(p *Proposal, src *Sources) []Finding {
	var out []Finding
	bad := func(f string, a ...any) { out = append(out, Finding{Error: true, Msg: fmt.Sprintf(f, a...)}) }
	warn := func(f string, a ...any) { out = append(out, Finding{Msg: fmt.Sprintf(f, a...)}) }
	schemes, lang := proposalSchemes(p, src)
	if len(p.Decisions) == 0 || p.Title == "" {
		bad("a proposal names its decisions and has a title")
	}
	all := append([]Scheme{}, p.Schemes...)
	if c := p.Candidates; c != nil {
		for _, s := range c.Schemes {
			if _, ok := findWalton(src, s.ID); ok || p.Questions[s.ID] != nil || in(schemeIDs(p.Schemes), s.ID) {
				bad("candidate %s: a candidate is a new scheme, neither in walton.yml nor proposed", s.ID)
			}
			all = append(all, s)
		}
		for k := range c.Language {
			if _, ok := p.Language[k]; ok {
				bad("candidates: template %s is also proposed", k)
			} else if _, ok := src.Language[k]; ok {
				bad("candidates: template %s is already in walton.yml", k)
			}
		}
	}
	for k := range p.Questions {
		if !in(schemeIDs(p.Schemes), k) {
			bad("questions for %s, which is not a proposed scheme", k)
		}
	}
	if c := p.Candidates; c != nil {
		for k := range c.Questions {
			if !in(schemeIDs(c.Schemes), k) {
				bad("candidates: questions for %s, which is not a candidate", k)
			}
		}
	}
	for _, s := range all {
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
		ids := map[string]bool{}
		for _, q := range p.questions(s.ID) {
			if ids[q.ID] {
				bad("%s: question %s is given twice", s.ID, q.ID)
			}
			ids[q.ID] = true
			if m := bookLabelRE.FindStringSubmatch(q.ID); m != nil {
				if l, ok := src.link(m[1]); !ok || !in(l.CQLabels, m[2]) {
					bad("%s: question %s: the compendium prints no %s under %s", s.ID, q.ID, m[2], m[1])
				} else if es := linkedEntries(src, s.ID); len(es) > 0 && !in(es, m[1]) {
					bad("%s: question %s: %s is not a compendium entry of %s (%s)", s.ID, q.ID, m[1], s.ID, strings.Join(es, ", "))
				}
			}
			if !in(s.formulas(), q.Target) {
				bad("%s: question %s targets %s, not a formula of the scheme", s.ID, q.ID, q.Target)
			} else if strings.HasPrefix(q.ID, "walton/") &&
				(!in(s.Assumptions, q.Target) && !in(s.Exceptions, q.Target) || "walton/"+predicate(q.Target) != q.ID) {
				bad("%s: question %s: a walton/ question targets an assumption or exception whose predicate its id names (a premise is challenged, %s)", s.ID, q.ID, Challenge)
			}
		}
	}
	type tally struct{ arguments, answers, nearMisses int }
	exemplified := map[string]bool{} // candidates with an example
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
		qs := p.questions(x.Scheme)
		for _, a := range x.Attacks {
			if a.Question == Challenge {
				switch {
				case !in(s.Premises, a.Target) && !in(s.Assumptions, a.Target):
					bad("%s: a challenge targets a premise or assumption of %s, and %q is not one", where, x.Scheme, a.Target)
				case targetedBy(qs, a.Target) != "":
					bad("%s: %s is targeted by question %s, which the game asks instead of a challenge", where, a.Target, targetedBy(qs, a.Target))
				}
				continue
			}
			if a.Target != "" {
				bad("%s: only a challenge (%s) names its target; %s targets its own formula", where, Challenge, a.Question)
			}
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
		if p.candidate(x.Scheme) {
			exemplified[x.Scheme] = true
		}
	}
	if c := p.Candidates; c != nil {
		for _, s := range c.Schemes {
			if !exemplified[s.ID] {
				warn("candidate %s: no example", s.ID)
			}
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
		for _, q := range p.questions(s.ID) {
			if !pressed[s.ID][q.ID] {
				warn("%s: question %s is pressed in no example", s.ID, q.ID)
			}
		}
	}
	return out
}

// bookLabelRE matches a question labelled as the book prints it: the
// compendium entry and the question's label.
var bookLabelRE = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)*[a-z]?)/(CQ[0-9]+)$`)

// linkedEntries are the compendium entries that links.json links to a
// walton.yml scheme, with the entries their questions are printed under.
func linkedEntries(src *Sources, id string) []string {
	var out []string
	for _, e := range src.Links {
		for _, l := range e.Links {
			if l.ID == id {
				out = append(out, e.Number)
				if e.SharesCQs != nil {
					out = append(out, *e.SharesCQs)
				}
			}
		}
	}
	return out
}

// targetedBy is the id of the first question that targets a formula, or "".
func targetedBy(qs []PQuery, f string) string {
	for _, q := range qs {
		if q.Target == f {
			return q.ID
		}
	}
	return ""
}

func schemeIDs(ss []Scheme) []string {
	var out []string
	for _, s := range ss {
		out = append(out, s.ID)
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
	if p.Applied != "" {
		fmt.Fprintf(&b, "**Applied:** %s. The \"Now\" columns below show `walton.yml` as it is since, so they equal the proposed ones; the formulas before the change are in `walton-clean/CHANGELOG.md`, among the owner's decisions.\n\n", strings.TrimSpace(p.Applied))
	}
	if p.Summary != "" {
		b.WriteString(strings.TrimSpace(p.Summary) + "\n\n")
	}
	b.WriteString("## The schemes as proposed\n\n")
	for _, s := range p.Schemes {
		writeScheme(&b, src, s, p.Questions[s.ID], "A new scheme.")
	}
	writeTemplates(&b, "**Templates added or changed:**", p.Language)
	if c := p.Candidates; c != nil {
		b.WriteString("## Candidates: sketched, not proposed\n\n")
		if c.Note != "" {
			b.WriteString(strings.TrimSpace(c.Note) + "\n\n")
		}
		for _, s := range c.Schemes {
			writeScheme(&b, src, s, c.Questions[s.ID], "A candidate, not proposed: it is not applied to `walton.yml`.")
		}
		writeTemplates(&b, "**The candidates' templates:**", c.Language)
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
		writeExample(&b, "###", i+1, x, schemes[x.Scheme], lang, p.questions(x.Scheme), p.candidate(x.Scheme))
	}
	return b.Bytes()
}

// writeExample renders an example under a heading of the given level: how it
// is used, its text and bindings, its scheme's formulas read in English, the
// questions pressed against it and its note.
func writeExample(b *bytes.Buffer, level string, n int, x ProposalExample, s Scheme, lang map[string]string, qs []PQuery, candidate bool) {
	fmt.Fprintf(b, "%s %d. %s\n\n", level, n, x.Title)
	var tags []string
	switch x.Use {
	case "answer":
		tags = append(tags, "**as an answer** in the game: "+strings.TrimSuffix(strings.Join(strings.Fields(x.Answers), " "), "."))
	case "near-miss":
		tags = append(tags, fmt.Sprintf("**a near miss** of `%s`: it is `%s`", x.Of, x.Scheme))
	default:
		tags = append(tags, fmt.Sprintf("**as an argument**, with `%s`", x.Scheme))
	}
	if candidate {
		tags = append(tags, "a candidate")
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
	fmt.Fprintf(b, "%s.\n\n> %s\n\n", strings.Join(tags, "; "), strings.Join(strings.Fields(x.Text), " "))
	var vars []string
	for _, v := range s.Variables {
		vars = append(vars, fmt.Sprintf("%s = “%s”", v, x.Bindings[v]))
	}
	fmt.Fprintf(b, "Bindings: %s.\n\n", strings.Join(vars, "; "))
	for _, f := range s.Premises {
		fmt.Fprintf(b, "- %s\n", EnglishWith(f, lang, x.Bindings))
	}
	for _, f := range s.Assumptions {
		fmt.Fprintf(b, "- *Presumed:* %s\n", EnglishWith(f, lang, x.Bindings))
	}
	for _, f := range s.Exceptions {
		fmt.Fprintf(b, "- *Unless:* %s\n", EnglishWith(f, lang, x.Bindings))
	}
	for _, f := range s.Conclusions {
		// Some templates open with "Therefore,"; the label says it.
		fmt.Fprintf(b, "- **So:** %s\n", sentence(strings.TrimPrefix(EnglishWith(f, lang, x.Bindings), "Therefore, ")))
	}
	if len(x.Attacks) > 0 {
		b.WriteString("\nPressed with its questions:\n\n")
		for _, a := range x.Attacks {
			if a.Question == Challenge {
				fmt.Fprintf(b, "- **Challenge** (%s), at the premise “%s”: %s\n", Challenge, strings.TrimSuffix(EnglishWith(a.Target, lang, x.Bindings), "."), strings.Join(strings.Fields(a.Text), " "))
				continue
			}
			q := questionByID(qs, a.Question)
			if q == nil {
				continue
			}
			fmt.Fprintf(b, "- **%s**, at the %s “%s”: %s\n", a.Question, roleOf(s, q.Target), strings.TrimSuffix(EnglishWith(q.Target, lang, x.Bindings), "."), strings.Join(strings.Fields(a.Text), " "))
		}
	}
	if x.Note != "" {
		fmt.Fprintf(b, "\n*Note:* %s\n", strings.Join(strings.Fields(x.Note), " "))
	}
	b.WriteString("\n")
}

// writeScheme renders a scheme against walton.yml's (now and proposed, or
// new) and its questions.
func writeScheme(b *bytes.Buffer, src *Sources, s Scheme, qs []PQuery, isNew string) {
	fmt.Fprintf(b, "### `%s`: %s\n\n", s.ID, s.Meta.Title)
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
			fmt.Fprintf(b, "| %s | %s | %s |\n", r.name, code(r.old), code(r.new))
		}
	} else {
		fmt.Fprintf(b, "%s\n\n| Role | Proposed |\n|---|---|\n", isNew)
		fmt.Fprintf(b, "| variables | %s |\n| premises | %s |\n", code(s.Variables), code(s.Premises))
		if len(s.Assumptions) > 0 {
			fmt.Fprintf(b, "| assumptions | %s |\n", code(s.Assumptions))
		}
		if len(s.Exceptions) > 0 {
			fmt.Fprintf(b, "| exceptions | %s |\n", code(s.Exceptions))
		}
		fmt.Fprintf(b, "| conclusion | %s |\n", code(s.Conclusions))
	}
	b.WriteString("\n")
	if len(qs) > 0 {
		b.WriteString("| Question | Asks | Targets |\n|---|---|---|\n")
		for _, q := range qs {
			fmt.Fprintf(b, "| %s | %s | `%s` (%s) |\n", q.ID, cell(q.Asks), q.Target, roleOf(s, q.Target))
		}
		b.WriteString("\n")
	}
}

func writeTemplates(b *bytes.Buffer, title string, lang map[string]string) {
	if len(lang) == 0 {
		return
	}
	b.WriteString(title + "\n\n")
	keys := make([]string, 0, len(lang))
	for k := range lang {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(b, "- `%s`: “%s”\n", k, strings.TrimSpace(lang[k]))
	}
	b.WriteString("\n")
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
