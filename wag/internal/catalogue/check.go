package catalogue

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// Finding is something the checker reports about an entry file.
type Finding struct {
	File  string // the entry file, or "" for the catalogue as a whole
	Error bool   // an error fails the check; a warning does not
	Msg   string
}

func (f Finding) String() string {
	level := "warning"
	if f.Error {
		level = "error"
	}
	if f.File == "" {
		return level + ": " + f.Msg
	}
	return fmt.Sprintf("%s: %s: %s", level, f.File, f.Msg)
}

// Options tune the check.
type Options struct {
	// Book is the book's text for the copy check; nil skips it.
	Book *Corpus
	// A text that shares CopyRun words in a row with a book sentence is an
	// error, WarnRun a warning (normalized words: lower case, variables as
	// one word).
	CopyRun, WarnRun int
	// A text of at most ShortText words that is CopySimilar similar to a
	// book sentence is an error, WarnSimilar a warning.
	ShortText                int
	CopySimilar, WarnSimilar float64
}

// DefaultOptions are the thresholds of the copy check.
func DefaultOptions() Options {
	return Options{CopyRun: 8, WarnRun: 6, ShortText: 25, CopySimilar: 0.85, WarnSimilar: 0.7}
}

// Report is the result of a check.
type Report struct {
	Findings []Finding
	Entries  int
	Progress progress
	Tier1    progress
	Copied   bool // whether the copy check ran
	Texts    int  // our texts the copy check compared
}

// Errors counts the errors.
func (r *Report) Errors() int {
	n := 0
	for _, f := range r.Findings {
		if f.Error {
			n++
		}
	}
	return n
}

// Summary is a line on where the catalogue stands.
func (r *Report) Summary() string {
	p := r.Progress
	return fmt.Sprintf("%d entries; %d work items: %d verified, %d drafted, %d disputed, %d generated, %d to write (tier 1: %d of %d verified)",
		r.Entries, p.items, p.verified, p.drafted, p.disputed, p.generated, p.todo, r.Tier1.verified, r.Tier1.items)
}

type checker struct {
	r    *Report
	file string
	src  *Sources
	ids  map[string]bool
}

func (c *checker) err(f string, a ...any) {
	c.r.Findings = append(c.r.Findings, Finding{File: c.file, Error: true, Msg: fmt.Sprintf(f, a...)})
}

func (c *checker) warn(f string, a ...any) {
	c.r.Findings = append(c.r.Findings, Finding{File: c.file, Msg: fmt.Sprintf(f, a...)})
}

// Check checks a catalogue directory against the sources:
//   - every walton.yml scheme and every compendium scheme has its entry,
//     and every entry is made from the sources;
//   - the source fields are the sources' (formulas, English, game);
//   - every formula an item names is one of its scheme's, in that role;
//   - every compendium entry, critical question and family the sources
//     link to the scheme is there, as are the discrepancies the sources
//     show;
//   - every game critical question asked of the scheme is there;
//   - statuses and values follow the vocabularies, and written items have
//     their text;
//   - no text of ours reproduces the book (when given the book's text);
//   - the index is current.
func Check(dir string, src *Sources, opt Options) (*Report, error) {
	r := &Report{}
	entries, bad, err := ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for name, e := range bad {
		r.Findings = append(r.Findings, Finding{File: name, Error: true, Msg: "does not read: " + e.Error()})
	}
	g := &Generator{Src: src}
	gen := map[string]*Entry{}
	var order []string
	for _, e := range g.Build() {
		gen[e.ID] = e
		order = append(order, e.ID)
	}
	byID := map[string]*Entry{}
	ids := map[string]bool{}
	for name, e := range entries {
		ids[e.ID] = true
		if prev, dup := byID[e.ID]; dup {
			r.Findings = append(r.Findings, Finding{File: name, Error: true, Msg: fmt.Sprintf("id %q is also the id of %s.yml", e.ID, prev.ID)})
		}
		byID[e.ID] = e
		if name != e.ID+".yml" {
			r.Findings = append(r.Findings, Finding{File: name, Error: true, Msg: fmt.Sprintf("file name should be %s.yml", e.ID)})
		}
	}
	for _, id := range order {
		if byID[id] == nil {
			r.Findings = append(r.Findings, Finding{Error: true, Msg: fmt.Sprintf("no entry for %s: run wag-cli --catalogue-gen", id)})
		}
	}
	var names []string
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	var all []*Entry
	for _, name := range names {
		e := entries[name]
		all = append(all, e)
		c := &checker{r: r, file: name, src: src, ids: ids}
		want := gen[e.ID]
		if want == nil {
			c.err("the sources make no entry %q (a scheme removed from walton.yml or links.json?)", e.ID)
		} else {
			c.entry(e, want)
		}
		c.statuses(e)
		if opt.Book != nil {
			c.copies(e, opt)
		}
		p := progressOf(e)
		r.Progress.plus(p)
		if e.Tier == 1 {
			r.Tier1.plus(p)
		}
	}
	for _, f := range src.Families {
		for _, m := range f.Members {
			if m.Scheme != "" {
				continue
			}
			if l, ok := src.link(m.Compendium); ok && len(l.Links) > 0 {
				var ids []string
				for _, k := range l.Links {
					ids = append(ids, src.migrate(k.ID))
				}
				r.Findings = append(r.Findings, Finding{Msg: fmt.Sprintf("families.json: family %s lists compendium %s with no scheme, but links.json links it to %s; one of the two needs correcting",
					f.ID, m.Compendium, strings.Join(ids, ", "))})
			}
		}
	}
	r.Entries = len(entries)
	r.Copied = opt.Book != nil
	if opt.Book == nil {
		r.Findings = append(r.Findings, Finding{Msg: "the copy check did not run: give the book's text (--book)"})
	}
	idx, err := os.ReadFile(filepath.Join(dir, IndexFile))
	// A checkout on Windows may turn the line ends into CRLF.
	if err != nil || !bytes.Equal(bytes.ReplaceAll(idx, []byte("\r\n"), []byte("\n")), Index(all)) {
		r.Findings = append(r.Findings, Finding{File: IndexFile, Error: true, Msg: "out of date: run wag-cli --catalogue-gen"})
	}
	return r, nil
}

// entry checks an entry against the one the sources make.
func (c *checker) entry(e, want *Entry) {
	// Source fields.
	if e.Name != want.Name {
		c.err("name %q: walton.yml says %q", e.Name, want.Name)
	}
	if !same(e.Walton, want.Walton) {
		c.err("walton: not walton.yml's; run wag-cli --catalogue-gen")
	}
	if !same(e.Game, want.Game) {
		c.err("game: not the NASCQ's; run wag-cli --catalogue-gen")
	}
	if !same(e.Formulas, want.Formulas) {
		c.err("formulas: not walton.yml's (formulas or their order); run wag-cli --catalogue-gen")
	}
	if e.Tier != 1 && e.Tier != 2 {
		c.err("tier %d: want 1 or 2", e.Tier)
	}
	w, hasW := Scheme{}, false
	if e.Walton != nil {
		w, hasW = c.src.scheme(e.Walton.ID)
	}

	// The compendium entries.
	books := map[string]bool{}
	for _, b := range e.Book {
		books[b.Entry] = true
		l, ok := c.src.link(b.Entry)
		if !ok {
			c.err("book entry %q is not in the compendium (links.json)", b.Entry)
			continue
		}
		var exp *BookRef
		for i := range want.Book {
			if want.Book[i].Entry == b.Entry {
				exp = &want.Book[i]
			}
		}
		switch {
		case exp == nil && owned(b.Status):
			c.err("book entry %s: links.json does not link it to this scheme", b.Entry)
		case exp == nil:
			c.warn("book entry %s: links.json does not link it to this scheme (%s); update links.json if the link is right", b.Entry, b.Status)
		case owned(b.Status) && !same(b, *exp):
			c.err("book entry %s: not as links.json has it (title %q, pages %q, match %q); run wag-cli --catalogue-gen", b.Entry, l.Name, l.BookPages, exp.Match)
		case !in([]string{"exact", "partial", "none"}, b.Match):
			c.err("book entry %s: match %q: want exact, partial or none", b.Entry, b.Match)
		}
	}
	for _, b := range want.Book {
		if !books[b.Entry] {
			c.err("book entry %s (links.json) is missing", b.Entry)
		}
	}

	// Families.
	fams := map[string]bool{}
	for _, f := range e.Families {
		fams[f.Family] = true
		if !c.familyExists(f.Family) {
			c.err("family %q is not in families.json", f.Family)
		}
	}
	for _, f := range want.Families {
		if !fams[f.Family] {
			c.err("family %s (families.json) is missing", f.Family)
		}
		for _, x := range e.Families {
			if x.Family == f.Family && owned(x.Status) && !same(x, f) {
				c.err("family %s: not as families.json has it; run wag-cli --catalogue-gen", f.Family)
			}
		}
	}

	// Premise maps.
	maps := map[string]bool{}
	for _, p := range e.PremiseMap {
		maps[p.Entry+"/"+p.Label] = true
		if !books[p.Entry] {
			c.err("premise map %s %q: %s is not one of the entry's compendium entries", p.Entry, p.Label, p.Entry)
		}
		c.targets(w, hasW, fmt.Sprintf("premise map %s %q", p.Entry, p.Label), p.Formulas)
		if owned(p.Status) {
			stale := true
			for _, x := range want.PremiseMap {
				if same(x, p) {
					stale = false
				}
			}
			if stale {
				c.err("premise map %s %q: not as links.json has it; run wag-cli --catalogue-gen", p.Entry, p.Label)
			}
		}
	}
	for _, p := range want.PremiseMap {
		if !maps[p.Entry+"/"+p.Label] {
			c.err("premise map %s %q (links.json) is missing", p.Entry, p.Label)
		}
	}

	// Critical questions.
	keptGame := map[string]bool{} // game questions that kept items carry
	for _, q := range e.CQs {
		if !owned(q.Status) && q.Game != nil {
			keptGame[q.Game.ID] = true
		}
	}
	seen := map[string]bool{}
	games := map[string]bool{}
	var unlinked []string
	for _, q := range e.CQs {
		where := "question " + q.ID
		if seen[q.ID] {
			c.err("%s: the id is used twice", where)
		}
		seen[q.ID] = true
		if q.Book != nil {
			if l, ok := c.src.link(q.Book.Entry); !ok || !in(l.CQLabels, q.Book.Label) {
				c.err("%s: the compendium prints no %s under %s", where, q.Book.Label, q.Book.Entry)
			} else if q.ID != q.Book.Entry+"/"+q.Book.Label {
				c.err("%s: the id should be %s/%s", where, q.Book.Entry, q.Book.Label)
			}
		}
		if q.Book == nil && q.Game == nil {
			switch {
			case owned(q.Status):
				c.err("%s: neither a book nor a game question", where)
			case !strings.HasPrefix(q.ID, "walton/"):
				c.err("%s: a question that neither the book nor the game asks needs the id walton/<predicate>", where)
			case len(q.Targets) != 1 || (q.Targets[0].Locus != "assumption" && q.Targets[0].Locus != "exception") ||
				"walton/"+predicate(q.Targets[0].Formula) != q.ID:
				c.err("%s: a walton/ question targets one assumption or exception, whose predicate the id names", where)
			}
		}
		c.targets(w, hasW, where, q.Targets)
		if q.Undercuts && len(q.Targets) > 0 {
			c.err("%s: a question that undercuts the inference targets no formula", where)
		}
		if q.Book != nil && len(q.Targets) == 0 && !q.Undercuts && hasW {
			unlinked = append(unlinked, q.ID)
		}
		if q.Game != nil {
			games[q.Game.ID] = true
			var exp *GameCQ
			for _, x := range want.CQs {
				if x.Game != nil && x.Game.ID == q.Game.ID {
					exp = x.Game
				}
			}
			switch {
			case exp == nil:
				c.err("%s: the game does not ask %s of this scheme", where, q.Game.ID)
			case !same(q.Game, exp):
				c.err("%s: the game part is not the NASCQ's; run wag-cli --catalogue-gen", where)
			case exp.Pattern != "" && len(q.Targets) > 0 && !hasTarget(q.Targets, exp.Pattern):
				c.warn("%s: the game's question targets %s, the book's question %s", where, exp.Pattern, formulasOf(q.Targets))
			}
			if exp != nil && hasW && exp.Pattern != "" {
				if l := w.locusOf(exp.Pattern); l != "" && !gameLocusFits(exp.Locus, l) {
					c.warn("%s: the game asks it at the %s, but %s is a %s of walton.yml's scheme", where, exp.Locus, exp.Pattern, l)
				}
			}
		}
		if owned(q.Status) && q.Book != nil {
			for _, x := range want.CQs {
				if x.Game != nil && keptGame[x.Game.ID] {
					x.Game = nil
				}
				if x.ID == q.ID && !same(x, q) {
					c.err("%s: not as links.json and the NASCQ have it; run wag-cli --catalogue-gen", where)
				}
			}
		}
		if hasText(q.Status) && strings.TrimSpace(q.Paraphrase) == "" {
			c.err("%s: %s, but no paraphrase", where, q.Status)
		}
	}
	for _, q := range want.CQs {
		if q.Book != nil && !seen[q.ID] {
			c.err("question %s (compendium) is missing", q.ID)
		}
		if q.Game != nil && !games[q.Game.ID] {
			c.err("game question %s is missing", q.Game.ID)
		}
	}
	if len(unlinked) > 0 {
		c.warn("%d book question(s) target no formula and are not marked as undercutting: %s", len(unlinked), strings.Join(unlinked, ", "))
	}

	// Discrepancies the sources show.
	for _, d := range want.Discrepancies {
		found := false
		for _, x := range e.Discrepancies {
			if x.Kind == d.Kind && x.About == d.About {
				found = true
			}
		}
		if !found {
			c.err("discrepancy %s %q (from the sources) is missing", d.Kind, d.About)
		}
	}
	for _, d := range e.Discrepancies {
		if !in(DiscrepancyKinds, d.Kind) {
			c.err("discrepancy kind %q: want one of %s", d.Kind, strings.Join(DiscrepancyKinds, ", "))
		}
	}

	// Values.
	for _, x := range e.Contrasts {
		if !c.ids[x.Scheme] {
			c.err("contrast with %q: no such entry", x.Scheme)
		}
		if hasText(x.Status) && strings.TrimSpace(x.Text) == "" {
			c.err("contrast with %s: %s, but no text", x.Scheme, x.Status)
		}
	}
	for i, x := range e.Examples {
		if strings.TrimSpace(x.Text) == "" || strings.TrimSpace(x.Coding) == "" {
			c.err("example %d: a text and its coding are both needed", i+1)
		}
	}
	if hasText(e.Summary.Status) && strings.TrimSpace(e.Summary.Text) == "" {
		c.err("summary: %s, but no text", e.Summary.Status)
	}
	if !in(SourceValues, e.Class.Source) || !in(ReasoningValues, e.Class.Reasoning) || !in(EpistemicValues, e.Class.Epistemic) {
		c.err("classification: source %q, reasoning %q, epistemic %q: want source dependent or independent, reasoning practical or epistemic, epistemic discovery or application", e.Class.Source, e.Class.Reasoning, e.Class.Epistemic)
	}
	if e.Class.Epistemic != "" && e.Class.Reasoning != "epistemic" {
		c.err("classification: epistemic %q given for reasoning %q", e.Class.Epistemic, e.Class.Reasoning)
	}
	if e.Class.Source == "dependent" && e.Class.Reasoning != "" {
		c.err("classification: reasoning %q given for a source-dependent scheme; the tree divides only source-independent ones", e.Class.Reasoning)
	}
	if h := e.Class.Heading(); e.Class.Group != "" && !in(Groups[h], e.Class.Group) {
		c.err("classification: group %q under heading %q: want one of %s", e.Class.Group, h, strings.Join(Groups[h], ", "))
	}
	if hasText(e.Class.Status) && (e.Class.Heading() == "" || e.Class.Group == "") {
		c.err("classification: %s, but not placed down to a group", e.Class.Status)
	}
	if e.Specialises.Scheme != "" && !c.ids[e.Specialises.Scheme] {
		c.err("specialises %q: no such entry", e.Specialises.Scheme)
	}
	for _, t := range append(append([]string{}, e.DialogueTypes.Proper...), e.DialogueTypes.Suspect...) {
		if !in(DialogueTypeNames, t) {
			c.err("dialogue type %q: want one of %s", t, strings.Join(DialogueTypeNames, ", "))
		}
	}
}

// targets checks that each formula is the scheme's, in the role given.
func (c *checker) targets(w Scheme, hasW bool, where string, ts []Target) {
	for _, t := range ts {
		if !in(Loci, t.Locus) {
			c.err("%s: locus %q: want one of %s", where, t.Locus, strings.Join(Loci, ", "))
			continue
		}
		if !hasW {
			c.err("%s: %s: the entry has no walton.yml scheme, so no formulas", where, t.Formula)
			continue
		}
		if l := w.locusOf(t.Formula); l != t.Locus {
			if l == "" {
				c.err("%s: %s is not a formula of walton.yml's %s", where, t.Formula, w.ID)
			} else {
				c.err("%s: %s is %s of walton.yml's %s, not %s", where, t.Formula, article(l), w.ID, article(t.Locus))
			}
		}
	}
}

// statuses checks every status string.
func (c *checker) statuses(e *Entry) {
	check := func(where, s string) {
		if _, err := ParseStatus(s); err != nil {
			c.err("%s: %v", where, err)
		}
	}
	for _, x := range e.Book {
		check("book entry "+x.Entry, x.Status)
	}
	for _, x := range e.Families {
		check("family "+x.Family, x.Status)
	}
	for _, x := range e.PremiseMap {
		check("premise map "+x.Entry+" "+x.Label, x.Status)
	}
	for _, x := range e.CQs {
		check("question "+x.ID, x.Status)
	}
	for _, x := range e.Contrasts {
		check("contrast with "+x.Scheme, x.Status)
	}
	for i, x := range e.Examples {
		check(fmt.Sprintf("example %d", i+1), x.Status)
	}
	for _, x := range e.Discrepancies {
		check("discrepancy "+x.Kind+" "+x.About, x.Status)
	}
	check("summary", e.Summary.Status)
	check("classification", e.Class.Status)
	check("specialises", e.Specialises.Status)
	check("dialogue types", e.DialogueTypes.Status)
}

// copies runs the copy check on every text of ours in the entry.
func (c *checker) copies(e *Entry, opt Options) {
	texts := []struct{ where, text string }{
		{"summary", e.Summary.Text}, {"specialises note", e.Specialises.Note},
		{"dialogue types note", e.DialogueTypes.Note}, {"classification note", e.Class.Note},
	}
	for _, q := range e.CQs {
		texts = append(texts, struct{ where, text string }{"question " + q.ID + " paraphrase", q.Paraphrase},
			struct{ where, text string }{"question " + q.ID + " answer", q.Answer})
	}
	for _, x := range e.Contrasts {
		texts = append(texts, struct{ where, text string }{"contrast with " + x.Scheme, x.Text})
	}
	for i, x := range e.Examples {
		texts = append(texts, struct{ where, text string }{fmt.Sprintf("example %d", i+1), x.Text})
	}
	for _, d := range e.Discrepancies {
		texts = append(texts, struct{ where, text string }{"discrepancy " + d.About + " decision", d.Decision})
	}
	for _, t := range texts {
		if strings.TrimSpace(t.text) == "" {
			continue
		}
		c.r.Texts++
		if g := opt.Book.Shares(t.text, opt.CopyRun); g != "" {
			c.err("%s reproduces the book: %d words in a row (%q)", t.where, opt.CopyRun, g)
			continue
		}
		if r, s := opt.Book.Closest(t.text, opt.ShortText); r >= opt.CopySimilar {
			c.err("%s reproduces a book sentence (%.0f%% the same words, in order): %q", t.where, 100*r, s)
			continue
		} else if r >= opt.WarnSimilar {
			c.warn("%s is close to a book sentence (%.0f%%): %q; rephrase it", t.where, 100*r, s)
			continue
		}
		if g := opt.Book.Shares(t.text, opt.WarnRun); g != "" {
			c.warn("%s shares %d words in a row with the book (%q); rephrase it", t.where, opt.WarnRun, g)
		}
	}
}

func (c *checker) familyExists(id string) bool {
	for _, f := range c.src.Families {
		if f.ID == id {
			return true
		}
	}
	return false
}

// article puts "a" or "an" before a role.
func article(role string) string {
	if strings.IndexAny(role[:1], "aeiou") == 0 {
		return "an " + role
	}
	return "a " + role
}

// same compares two values by their YAML, so that an empty list read from a
// file equals a nil one.
func same(a, b any) bool {
	x, err1 := yaml.Marshal(a)
	y, err2 := yaml.Marshal(b)
	return err1 == nil && err2 == nil && bytes.Equal(x, y)
}

// hasText reports whether a status says the item was written.
func hasText(status string) bool {
	st, err := ParseStatus(status)
	return err == nil && (st.State == "drafted" || st.State == "verified")
}

// gameLocusFits reports whether the game's locus fits a formula's role in
// walton.yml. The game presumes a scheme's assumptions as premises, so that
// a question can undermine them; it asks those at the premise.
func gameLocusFits(game, role string) bool {
	return game == role || (game == "premise" && role == "assumption")
}

func hasTarget(ts []Target, formula string) bool {
	for _, t := range ts {
		if compact(t.Formula) == compact(formula) {
			return true
		}
	}
	return false
}

// predicate is a formula's predicate, or the formula itself if it is a
// bare variable.
func predicate(f string) string {
	f = strings.TrimPrefix(strings.TrimSpace(f), "¬")
	if i := strings.IndexByte(f, '('); i >= 0 {
		return f[:i]
	}
	return f
}

func formulasOf(ts []Target) string {
	var fs []string
	for _, t := range ts {
		fs = append(fs, t.Formula)
	}
	return strings.Join(fs, ", ")
}
