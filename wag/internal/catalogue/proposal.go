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
	// Links are links of links.json that the proposal adds or replaces
	// (ProposedLink); applying the proposal writes them there.
	Links []ProposedLink `yaml:"links,omitempty"`
}

// ProposedLink is a link of links.json (docs/scheme-registry/links.json)
// that a proposal adds or replaces, written with links.json's keys: a
// scheme of the proposal formalizes a compendium entry, how closely, and
// which of its formulas each printed premise, the conclusion and each
// printed question map to. The check reads links.json with the proposal's
// links in place, so that a new scheme can carry the questions the book
// prints for its entry (D008's modal variants and 33's questions), and the
// target contract holds the proposal's questions to the maps it proposes.
// An applied proposal's links are in links.json as proposed.
type ProposedLink struct {
	Entry      string `yaml:"entry"`
	WaltonLink `yaml:",inline"`
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
	AnswerTo *AnswerTo         `yaml:"answer_to,omitempty"` // for an answer: what it answers, checked
	Bindings map[string]string `yaml:"bindings"`
	Attacks  []Attack          `yaml:"attacks,omitempty"`
	Note     string            `yaml:"note,omitempty"`
}

// AnswerTo is what an answer example answers: the argument answered (its
// scheme, and the bindings of the formula the question targets), the
// question, the formula when the question does not name exactly one, and
// whether the answer attacks the argument or defends it. CheckProposal finds
// the formula the question targets, where it is in the argument (its locus),
// and so what the answer must conclude (game-rules.md: an attack argues
// against a premise, an assumption or the conclusion, and for an exception;
// a defence the reverse), and compares it with the example's conclusion,
// both instantiated. The answers field stays the reader's description. The
// case an argument is about (its context) is not compared here: it is the
// records' (the stage 2 identity contract), not the schemes'.
type AnswerTo struct {
	Scheme   string            `yaml:"scheme"`
	Question string            `yaml:"question"`
	Target   string            `yaml:"target,omitempty"`
	Polarity string            `yaml:"polarity"` // attack or defence
	Bindings map[string]string `yaml:"bindings"`
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
	src = proposedLinks(p, src, bad)
	schemes, lang := proposalSchemes(p, src)
	if len(p.Decisions) == 0 || p.Title == "" {
		bad("a proposal names its decisions and has a title")
	}
	// An id given twice would hide one of its schemes (proposalSchemes keeps
	// the last).
	proposed := map[string]bool{}
	for _, s := range p.Schemes {
		if proposed[s.ID] {
			bad("%s is proposed twice", s.ID)
		}
		proposed[s.ID] = true
	}
	if c := p.Candidates; c != nil {
		sketched := map[string]bool{}
		for _, s := range c.Schemes {
			if sketched[s.ID] {
				bad("candidate %s is sketched twice", s.ID)
			}
			sketched[s.ID] = true
		}
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
			if in(schemeIDs(p.Schemes), s.ID) {
				if msg := carriedQuestion(src, s.ID, q.ID); msg != "" {
					bad("%s: question %s: %s", s.ID, q.ID, msg)
				} else if bookLabelRE.MatchString(q.ID) {
					// The target contract: links.json maps the book's question
					// to formulas of walton.yml as it is; a proposal that
					// retargets it must bring links.json along.
					if mapped, why := questionTargets(&Proposal{}, src, s, q.ID, ""); why == "" {
						hit := false
						for _, m := range mapped {
							hit = hit || compact(m.Formula) == compact(q.Target)
						}
						if !hit {
							warn("%s: question %s targets %s, and links.json maps %s of %s to %s: update links.json when the proposal is applied, or target what it maps",
								s.ID, q.ID, q.Target, q.ID, s.ID, lociFormulas(mapped))
						}
					} else {
						// No link maps the label (the review of D008-D074
						// v2, CW1): the proposal's own link must; links.json's
						// must follow when the proposal is applied.
						entry, label, _ := strings.Cut(q.ID, "/")
						if l, ok := proposedLinkOf(p, src, s.ID, entry); ok {
							bad("%s: question %s: the proposal's link of %s to %s maps no %s: map it there, or drop the question", s.ID, q.ID, s.ID, l, label)
						} else {
							warn("%s: question %s: links.json maps no %s of %s: add it to links.json when the proposal is applied, or drop the question", s.ID, q.ID, label, s.ID)
						}
					}
				}
			}
		}
	}
	// A proposed scheme linked to an entry carries in the catalogue every
	// question printed for it (generate.go): the proposal gives each, so
	// that its view shows what the catalogue will carry (the review of
	// D008-D074 v2, CW2).
	for _, s := range p.Schemes {
		given := map[string]bool{}
		for _, q := range p.Questions[s.ID] {
			given[q.ID] = true
		}
		for _, e := range src.Links {
			for _, l := range e.Links {
				if l.ID != s.ID {
					continue
				}
				under, labels := src.printedCQs(e)
				mapped := map[string]bool{}
				for _, r := range l.CQMap {
					if len(r.Walton) > 0 {
						mapped[r.CQ] = true
					}
				}
				for _, label := range labels {
					// A printed question the link maps to no formula is the
					// catalogue's undercutting question: it has no formula to
					// target, so the proposal cannot give it (the check of
					// D008-D074 v3, W3).
					if mapped[label] && !given[under+"/"+label] {
						warn("%s: %s/%s is printed for %s, and the catalogue will carry it for %s: give it among the scheme's questions", s.ID, under, label, e.Number, s.ID)
					}
				}
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
		switch {
		case x.Use == "near-miss" && x.Of == "":
			bad("%s: a near miss names the scheme it resembles (near_miss_of)", where)
		case x.Use != "near-miss" && x.Of != "":
			bad("%s: only a near miss names near_miss_of", where)
		case x.Use == "near-miss":
			if _, ok := schemes[x.Of]; !ok {
				bad("%s: near_miss_of names no scheme %q", where, x.Of)
			} else if x.Of == x.Scheme {
				bad("%s: a near miss of %s is coded with %s itself", where, x.Of, x.Scheme)
			}
		}
		switch {
		case x.Use != "answer" && x.AnswerTo != nil:
			bad("%s: only an answer names answer_to", where)
		case x.Use == "answer" && x.AnswerTo == nil && p.Applied == "":
			bad("%s: an answer names what it answers (answer_to: its scheme, question, polarity and bindings)", where)
		case x.Use == "answer" && x.AnswerTo == nil:
			warn("%s: an answer without answer_to (the proposal was applied before answer_to existed)", where)
		case x.Use == "answer":
			for _, m := range checkAnswer(p, src, schemes, x) {
				bad("%s: %s", where, m)
			}
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

// proposedLinks checks the links a proposal adds or replaces, and returns
// walton.yml's sources with them in place of the scheme's own link to the
// entry, if it has one, other schemes' links kept. Each links a proposed
// scheme, once, to an entry of links.json that is not a heading, exactly or
// partly, with a reason; each formula it maps is the scheme's, at the locus
// the map gives; each question it maps is one the compendium prints for the
// entry (or under the entry its questions are printed under). An applied
// proposal's links are in links.json as proposed.
func proposedLinks(p *Proposal, src *Sources, bad func(string, ...any)) *Sources {
	if len(p.Links) == 0 {
		return src
	}
	proposed := map[string]Scheme{}
	for _, s := range p.Schemes {
		proposed[s.ID] = s
	}
	out := *src
	out.Links = append([]LinkEntry(nil), src.Links...)
	seen := map[string]bool{}
	for _, l := range p.Links {
		where := fmt.Sprintf("the link of %s to %s", l.ID, l.Entry)
		s, ok := proposed[l.ID]
		if !ok {
			bad("%s: %s is not a proposed scheme", where, l.ID)
			continue
		}
		at := -1
		for i, e := range out.Links {
			if e.Number == l.Entry {
				at = i
			}
		}
		if at < 0 {
			bad("%s: links.json has no compendium entry %s", where, l.Entry)
			continue
		}
		e := out.Links[at]
		if e.HeadingOnly {
			bad("%s: %s is a heading, which no scheme formalizes", where, l.Entry)
		}
		if seen[l.Entry+" "+l.ID] {
			bad("%s: %s is linked twice to %s", where, l.ID, l.Entry)
		}
		seen[l.Entry+" "+l.ID] = true
		if l.Match != "exact" && l.Match != "partial" {
			bad("%s: match is exact or partial, not %q", where, l.Match)
		}
		if strings.TrimSpace(l.Reason) == "" {
			bad("%s: a link gives its reason", where)
		}
		maps := func(what string, fls []FormulaLocus) {
			if len(fls) == 0 {
				bad("%s: %s maps to no formula", where, what)
			}
			for _, fl := range fls {
				role := "?"
				for _, locus := range Loci {
					for _, f := range s.role(locus) {
						if compact(f) == compact(fl.Formula) {
							role = locus
						}
					}
				}
				switch {
				case role == "?":
					bad("%s: %s maps to %s, not a formula of %s", where, what, fl.Formula, l.ID)
				case role != fl.Locus:
					bad("%s: %s maps to %s as its %s, but it is %s's %s", where, what, fl.Formula, fl.Locus, l.ID, role)
				}
			}
		}
		// The labels the entry's links give its printed premises (the
		// book's, whichever scheme formalizes it, the one relinked
		// included): a proposed link uses them, and maps at least as many
		// premises as the entry prints (the review of D008-D074 v2, CW3).
		var known []string
		for _, o := range e.Links {
			for _, m := range o.PremiseMap {
				if !in(known, m.Label) {
					known = append(known, m.Label)
				}
			}
		}
		if len(l.PremiseMap) < e.NPremises {
			bad("%s: %s prints %d premise(s), and the link maps %d", where, l.Entry, e.NPremises, len(l.PremiseMap))
		}
		for _, m := range l.PremiseMap {
			if len(known) > 0 && !in(known, m.Label) {
				bad("%s: %s prints no premise %q (its links name %s)", where, l.Entry, m.Label, strings.Join(known, ", "))
			}
			maps("premise "+strconv.Quote(m.Label), m.Walton)
		}
		// Every premise the entry's links name is mapped, and none more
		// often than they map it, so that a label mapped twice cannot stand
		// for one never mapped (the check of D008-D074 v3, N1). The book
		// prints some labels twice (43), so a label's count is its links'.
		most, mine := map[string]int{}, map[string]int{}
		for _, o := range e.Links {
			n := map[string]int{}
			for _, m := range o.PremiseMap {
				n[m.Label]++
			}
			for k, v := range n {
				if v > most[k] {
					most[k] = v
				}
			}
		}
		for _, m := range l.PremiseMap {
			mine[m.Label]++
		}
		for _, label := range known {
			switch {
			case mine[label] == 0:
				bad("%s: %s prints premise %q, which the link does not map", where, l.Entry, label)
			case mine[label] > most[label]:
				bad("%s: the link maps %s's premise %q %d times, and the entry's links map it %d", where, l.Entry, label, mine[label], most[label])
			}
		}
		if l.Conclusion == nil || len(l.Conclusion.Walton) == 0 {
			bad("%s: the link maps no conclusion", where)
		} else {
			maps("the conclusion", l.Conclusion.Walton)
		}
		under, labels := src.printedCQs(e)
		rows := map[string]bool{}
		for _, r := range l.CQMap {
			if !in(labels, r.CQ) {
				bad("%s: the compendium prints no %s under %s", where, r.CQ, under)
			}
			if rows[r.CQ] {
				// The catalogue keeps one row per question (generate.go).
				bad("%s: the link maps %s twice", where, r.CQ)
			}
			rows[r.CQ] = true
			maps(r.CQ, r.Walton)
		}
		e.Links = append([]WaltonLink(nil), e.Links...)
		// links.json names a renamed scheme by its upstream id (treat for
		// threat, id-migrations.json): the proposed link replaces that one.
		had := -1
		for i := range e.Links {
			if src.migrate(e.Links[i].ID) == l.ID {
				had = i
			}
		}
		switch {
		case p.Applied != "" && had < 0:
			bad("%s: the proposal is applied, but links.json does not have its link of %s to %s", where, l.ID, l.Entry)
		case p.Applied != "" && linkKey(withID(e.Links[had], l.ID)) != linkKey(l.WaltonLink):
			bad("%s: the proposal is applied, but links.json does not have its link of %s to %s as proposed", where, l.ID, l.Entry)
		}
		if had < 0 {
			e.Links = append(e.Links, l.WaltonLink)
		} else {
			e.Links[had] = l.WaltonLink
		}
		out.Links[at] = e
	}
	return &out
}

// proposedLinkOf names the entry through which the proposal links a scheme
// to the entry its questions are printed under (the entry itself, or one
// that shares its questions), if it does.
func proposedLinkOf(p *Proposal, src *Sources, scheme, printedUnder string) (string, bool) {
	for _, l := range p.Links {
		if l.ID != scheme {
			continue
		}
		if l.Entry == printedUnder {
			return l.Entry, true
		}
		if e, ok := src.link(l.Entry); ok && e.SharesCQs != nil && *e.SharesCQs == printedUnder {
			return l.Entry, true
		}
	}
	return "", false
}

// withID is a link under another id: links.json's link of a renamed scheme,
// under its upstream id, compared with a proposal's under the cleaned one.
func withID(l WaltonLink, id string) WaltonLink {
	l.ID = id
	return l
}

// linkKey writes what a link says, so that a proposed link can be compared
// with links.json's: its id, match and reason, and each map's label, note
// and formulas with their loci (spacing aside).
func linkKey(l WaltonLink) string {
	var b strings.Builder
	item := func(label, note string, fls []FormulaLocus) {
		fmt.Fprintf(&b, "|%s|%s|", strings.TrimSpace(label), strings.Join(strings.Fields(note), " "))
		for _, fl := range fls {
			fmt.Fprintf(&b, "%s@%s,", compact(fl.Formula), fl.Locus)
		}
	}
	fmt.Fprintf(&b, "%s|%s|%s", l.ID, l.Match, strings.Join(strings.Fields(l.Reason), " "))
	for _, m := range l.PremiseMap {
		item(m.Label, m.Note, m.Walton)
	}
	b.WriteString("|conclusion")
	if l.Conclusion != nil {
		item(l.Conclusion.Label, l.Conclusion.Note, l.Conclusion.Walton)
	}
	b.WriteString("|questions")
	for _, r := range l.CQMap {
		item(r.CQ, r.Note, r.Walton)
	}
	return b.String()
}

// carriedQuestion says why a proposed scheme's question could not be carried
// into the catalogue once the proposal is applied, or "" if it can: the
// catalogue takes a question the book prints for the scheme's own entry (an
// entry links.json links to the scheme, or the one its questions are
// printed under), one the game asks (game/...), or a walton/ question of an
// assumption or exception (checked above). A question borrowed from another
// entry is a derived question (D023), which the catalogue cannot hold yet;
// candidates' sketches, which are never applied, may still borrow.
func carriedQuestion(src *Sources, scheme, id string) string {
	if strings.HasPrefix(id, "walton/") || strings.HasPrefix(id, "game/") {
		return ""
	}
	entry, label, ok := strings.Cut(id, "/")
	l, found := src.link(entry)
	if !ok || !found || !in(l.CQLabels, label) {
		return "neither a book question, a game question (game/…) nor a walton/ question; the catalogue could not carry it (a question derived from another, D023, is not supported yet)"
	}
	es := linkedEntries(src, scheme)
	if len(es) == 0 {
		return fmt.Sprintf("no compendium entry is linked to %s, so %s's question is borrowed: a derived question (D023), which the catalogue cannot carry yet", scheme, entry)
	}
	if !in(es, entry) {
		return fmt.Sprintf("%s is not a compendium entry of %s (%s)", entry, scheme, strings.Join(es, ", "))
	}
	return ""
}

// bookLabelRE matches a question labelled as the book prints it: the
// compendium entry and the question's label.
var bookLabelRE = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)*[a-z]?)/(CQ[0-9]+)$`)

// checkAnswer checks what an answer example says it answers (AnswerTo):
// that the scheme, the question and the formula it targets exist, that the
// bindings bind the formula's variables and only the scheme's, and that the
// example concludes what an answer there must conclude. A scheme whose
// conclusion is a bare proposition is not compared: the game binds it to
// the answer.
func checkAnswer(p *Proposal, src *Sources, schemes map[string]Scheme, x ProposalExample) []string {
	a := x.AnswerTo
	var out []string
	bad := func(f string, args ...any) { out = append(out, fmt.Sprintf(f, args...)) }
	answered, ok := schemes[a.Scheme]
	if !ok {
		return []string{fmt.Sprintf("answer_to: no scheme %q", a.Scheme)}
	}
	if a.Polarity != "attack" && a.Polarity != "defence" {
		bad("answer_to: polarity is attack or defence, not %q", a.Polarity)
	}
	targets, why := questionTargets(p, src, answered, a.Question, a.Target)
	if why != "" {
		return append(out, "answer_to: "+why)
	}
	var at FormulaLocus
	switch {
	case a.Target != "":
		found := false
		for _, t := range targets {
			if compact(t.Formula) == compact(a.Target) {
				at, found = t, true
			}
		}
		if !found {
			return append(out, fmt.Sprintf("answer_to: %s is not what %s targets in %s (%s)", a.Target, a.Question, a.Scheme, lociFormulas(targets)))
		}
	case len(targets) == 1:
		at = targets[0]
	default:
		return append(out, fmt.Sprintf("answer_to: %s targets %d formulas of %s (%s): name one as its target", a.Question, len(targets), a.Scheme, lociFormulas(targets)))
	}
	for _, v := range sortedKeys(a.Bindings) {
		if !in(answered.Variables, v) {
			bad("answer_to: %s is not a variable of %s", v, a.Scheme)
		}
	}
	if at.Formula == "" {
		return out // the argument itself (an undercut): no formula to compare
	}
	tt, err := parseFormula(at.Formula)
	if err != nil {
		return append(out, "answer_to: "+err.Error())
	}
	unbound := map[string]bool{}
	tt.walk(func(t *term) {
		if t.variable() && strings.TrimSpace(a.Bindings[t.name]) == "" && !unbound[t.name] {
			unbound[t.name] = true
			bad("answer_to binds no %s, a variable of %s", t.name, at.Formula)
		}
	})
	own := schemes[x.Scheme]
	if len(out) > 0 || len(own.Conclusions) == 0 {
		return out
	}
	ct, err := parseFormula(own.Conclusions[0])
	if err != nil || ct.variable() {
		return out // reported with the scheme, or a bare proposition the game binds
	}
	// What an answer must conclude: an attack argues against a premise, an
	// assumption or the conclusion, and for an exception; a defence the
	// reverse (game-rules.md, the table of loci).
	want := tt.bound(a.Bindings)
	if (a.Polarity == "attack") != (at.Locus == "exception") {
		want.neg = !want.neg
	}
	got := ct.bound(x.Bindings)
	if x.Form == "negated" {
		got.neg = !got.neg
	}
	if got.canon() != want.canon() {
		bad("it answers %s's %s at its %s (%s), so it must conclude %s; it concludes %s", a.Scheme, a.Question, at.Locus, a.Polarity, want.canon(), got.canon())
	}
	return out
}

// placeholderRE is a variable in an ask edge's pattern: {A}.
var placeholderRE = regexp.MustCompile(`\{([A-Za-z0-9_]+)\}`)

// questionTargets are the formulas a question of a scheme targets, with
// their loci, or why there are none: a proposed or candidate scheme's
// questions are this proposal's; a walton.yml scheme's are its walton/
// questions, the book's (links.json's question maps, under the entry they
// are printed under) and the game's (its ask edges); a challenge targets the
// premise or assumption it names.
func questionTargets(p *Proposal, src *Sources, s Scheme, q, named string) ([]FormulaLocus, string) {
	var out []FormulaLocus
	add := func(f string) { out = append(out, FormulaLocus{Formula: f, Locus: roleOf(s, f)}) }
	switch {
	case q == Challenge:
		for _, f := range append(append([]string{}, s.Premises...), s.Assumptions...) {
			if named != "" && compact(f) == compact(named) {
				add(f)
			}
		}
		if len(out) == 0 {
			return nil, fmt.Sprintf("a challenge targets a premise or assumption of %s, which it names as its target", s.ID)
		}
		return out, ""
	case in(schemeIDs(p.Schemes), s.ID) || p.candidate(s.ID):
		if pq := questionByID(p.questions(s.ID), q); pq != nil {
			add(pq.Target)
		}
	case strings.HasPrefix(q, "walton/"):
		for _, f := range append(append([]string{}, s.Assumptions...), s.Exceptions...) {
			if "walton/"+predicate(f) == q {
				add(f)
			}
		}
	case strings.HasPrefix(q, "CQ:"):
		for as, g := range src.Game {
			if g.CarneadesID != s.ID {
				continue
			}
			for _, e := range src.AskEdges {
				switch {
				case e.CQ != q || e.Scheme != as:
				case e.Pattern != "":
					out = append(out, FormulaLocus{Formula: placeholderRE.ReplaceAllString(e.Pattern, "$1"), Locus: e.Locus})
				case e.Locus == "conclusion":
					for _, c := range s.Conclusions {
						out = append(out, FormulaLocus{Formula: c, Locus: "conclusion"})
					}
				default:
					out = append(out, FormulaLocus{Locus: e.Locus}) // the argument itself
				}
			}
		}
	default:
		if entry, label, ok := strings.Cut(q, "/"); ok {
			for _, e := range src.Links {
				if e.Number != entry && (e.SharesCQs == nil || *e.SharesCQs != entry) {
					continue
				}
				for _, l := range e.Links {
					for _, r := range l.CQMap {
						if l.ID == s.ID && r.CQ == label {
							out = append(out, r.Walton...)
						}
					}
				}
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Sprintf("%s is not a question of %s", q, s.ID)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Formula < out[j].Formula })
	return out, ""
}

func lociFormulas(ls []FormulaLocus) string {
	var out []string
	for _, l := range ls {
		out = append(out, l.Formula)
	}
	return strings.Join(out, ", ")
}

func sortedKeys(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// canon writes a term, its bound variables as their phrases in quotes
// (spaces collapsed, lower case), so that two instances can be compared.
func (t *term) canon() string {
	var b strings.Builder
	if t.neg {
		b.WriteString("¬")
	}
	if t.phrase {
		b.WriteString(strconv.Quote(strings.ToLower(strings.Join(strings.Fields(t.name), " "))))
	} else {
		b.WriteString(t.name)
	}
	if len(t.args) > 0 {
		var args []string
		for _, a := range t.args {
			args = append(args, a.canon())
		}
		b.WriteString("(" + strings.Join(args, ",") + ")")
	}
	return b.String()
}

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
		fmt.Fprintf(&b, "**Applied:** %s. The \"Now\" columns below show `walton.yml` as it is since, so they equal the proposed ones, but for a scheme a later decision has changed, which is marked; the formulas before the change are in `walton-clean/CHANGELOG.md`, among the owner's decisions.\n\n", strings.TrimSpace(p.Applied))
	}
	if p.Summary != "" {
		b.WriteString(strings.TrimSpace(p.Summary) + "\n\n")
	}
	b.WriteString("## The schemes as proposed\n\n")
	for _, s := range p.Schemes {
		writeScheme(&b, src, s, p.Questions[s.ID], "A new scheme.", p.Applied != "")
	}
	writeTemplates(&b, "**Templates added or changed:**", p.Language)
	if len(p.Links) > 0 {
		b.WriteString("## Links to the compendium, as proposed\n\n")
		b.WriteString("Written into `docs/scheme-registry/links.json` when the proposal is applied, in place of the scheme's own link to the entry if it has one. The labels refer to the book's printed text, which is not reproduced here.\n\n")
		for _, l := range p.Links {
			fmt.Fprintf(&b, "- **%s** formalizes %s (%s): %s\n", l.ID, l.Entry, l.Match, strings.Join(strings.Fields(l.Reason), " "))
			row := func(label, note string, fls []FormulaLocus) {
				var fs []string
				for _, fl := range fls {
					fs = append(fs, fmt.Sprintf("`%s` (%s)", fl.Formula, fl.Locus))
				}
				fmt.Fprintf(&b, "  - %s: %s", label, strings.Join(fs, ", "))
				if note = strings.Join(strings.Fields(note), " "); note != "" {
					fmt.Fprintf(&b, "; %s", note)
				}
				b.WriteString("\n")
			}
			for _, m := range l.PremiseMap {
				row(m.Label, m.Note, m.Walton)
			}
			if l.Conclusion != nil {
				row("the conclusion", l.Conclusion.Note, l.Conclusion.Walton)
			}
			for _, r := range l.CQMap {
				row(r.CQ, r.Note, r.Walton)
			}
		}
		b.WriteString("\n")
	}
	if c := p.Candidates; c != nil {
		b.WriteString("## Candidates: sketched, not proposed\n\n")
		if c.Note != "" {
			b.WriteString(strings.TrimSpace(c.Note) + "\n\n")
		}
		for _, s := range c.Schemes {
			writeScheme(&b, src, s, c.Questions[s.ID], "A candidate, not proposed: it is not applied to `walton.yml`.", false)
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
func writeScheme(b *bytes.Buffer, src *Sources, s Scheme, qs []PQuery, isNew string, applied bool) {
	fmt.Fprintf(b, "### `%s`: %s\n\n", s.ID, s.Meta.Title)
	if old, ok := findWalton(src, s.ID); ok {
		if applied && !sameFormulas(old, s) {
			b.WriteString("*Changed since this proposal was applied:* `walton.yml`'s formulas differ from the proposed ones (a later decision; `walton-clean/CHANGELOG.md`). The \"Now\" column shows `walton.yml` as it is.\n\n")
		}
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
		if applied {
			// Not new: a later decision removed it (the check of D008-D074
			// v3, N2).
			b.WriteString("*Removed since this proposal was applied:* `walton.yml` no longer has this scheme (a later decision; `walton-clean/CHANGELOG.md`).\n\n| Role | Proposed |\n|---|---|\n")
		} else {
			fmt.Fprintf(b, "%s\n\n| Role | Proposed |\n|---|---|\n", isNew)
		}
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

// sameFormulas reports whether two versions of a scheme have the same
// variables and formulas, spacing aside, in the same order: the view's two
// columns list them in order, and the applied view's banner says that they
// are equal.
func sameFormulas(a, b Scheme) bool {
	key := func(s Scheme) string {
		var parts []string
		for _, l := range [][]string{s.Variables, s.Premises, s.Assumptions, s.Exceptions, s.Conclusions} {
			var fs []string
			for _, f := range l {
				fs = append(fs, compact(f))
			}
			parts = append(parts, strings.Join(fs, ","))
		}
		return strings.Join(parts, "|")
	}
	return key(a) == key(b)
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
