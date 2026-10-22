package catalogue

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// Tier1 lists the schemes verified first (plan §2): the 13 schemes of the
// batch-2 gold and the close relatives needed for their contrasts, plus
// defeasible modus ponens, the form "other" falls back to.
var Tier1 = []string{
	// practical reasoning
	"practical_reasoning1", "practical_reasoning2", "practical_goal", "end_and_means_1", "practical_necessary_condition",
	// consequences and values
	"positive_consequences", "negative_consequences", "negative_consequences_reasoning", "positive_value", "negative_value",
	// causes and signs
	"cause_to_effect", "cause_to_effect_modal", "correlation_to_cause", "sign",
	// cases and analogy
	"analogy", "practical_analogy_positive", "practical_analogy_negative", "from_example",
	// sources
	"expert_opinion", "position_to_know", "witness_testimony", "popular_opinion",
	// classification and rules
	"verbal_classification", "definition_to_verbal_classification", "rule", "established_rule",
	// alternatives and costs
	"alternative_cognitive", "alternative_normative", "sunk_costs", "waste",
	// slippery slope
	"slippery_slope_base_case", "slippery_slope_inductive_step",
	// the general form of "other"
	"defeasible_modus_ponens",
}

// Generator builds the catalogue's entries from the sources.
type Generator struct {
	Src   *Sources
	Today string // the date of a new entry's history line, YYYY-MM-DD
	By    string // who generated it
}

// Build returns the generated entries: one per walton.yml scheme, in its
// order, then one per compendium scheme that walton.yml lacks.
func (g *Generator) Build() []*Entry {
	var out []*Entry
	for _, w := range g.Src.Walton {
		out = append(out, g.entry(w))
	}
	for _, l := range g.Src.Links {
		if !l.HeadingOnly && len(l.Links) == 0 {
			out = append(out, g.bookOnly(l))
		}
	}
	return out
}

func (g *Generator) entry(w Scheme) *Entry {
	s := g.Src
	e := &Entry{
		ID:   w.ID,
		Name: strings.TrimSpace(w.Meta.Title),
		Tier: 2,
		Walton: &WaltonRef{ID: w.ID, Title: strings.TrimSpace(w.Meta.Title),
			Source: strings.Join(strings.Fields(w.Meta.Source), " ")},
		Formulas: g.formulas(w),
	}
	if e.Name == "" {
		e.Name = w.ID
	}
	if in(Tier1, w.ID) {
		e.Tier = 1
	}
	if gid, ok := s.gameScheme(w.ID); ok {
		e.Game = &GameScheme{ID: gid}
		for _, b := range s.Backups {
			if b.Scheme == gid && !in(e.Game.Backs, b.CQ) {
				e.Game.Backs = append(e.Game.Backs, b.CQ)
			}
		}
	}
	// The compendium entries the scheme formalizes.
	for _, l := range s.Links {
		for _, k := range l.Links {
			if s.migrate(k.ID) != w.ID {
				continue
			}
			ref := BookRef{Entry: l.Number, Title: l.Name, Pages: l.BookPages, Match: k.Match, Note: k.Reason, Status: Generated}
			if l.Parent != nil {
				ref.Parent = *l.Parent
			}
			if l.SharesCQs != nil && len(l.CQLabels) == 0 {
				ref.CQsPrintedUnder = *l.SharesCQs
			}
			e.Book = append(e.Book, ref)
			if k.Match != "exact" {
				e.Discrepancies = append(e.Discrepancies, Discrepancy{Kind: "partial-link", About: l.Number,
					Detail: "walton.yml formalizes compendium " + l.Number + " only in part: " + k.Reason, Status: Generated})
			}
			for _, p := range k.PremiseMap {
				e.PremiseMap = append(e.PremiseMap, PremiseLink{Entry: l.Number, Label: p.Label,
					Formulas: g.targets(e, w, l.Number+" "+p.Label, p.Walton), Note: p.Note, Status: Generated})
			}
			if k.Conclusion != nil {
				e.PremiseMap = append(e.PremiseMap, PremiseLink{Entry: l.Number, Label: "Conclusion",
					Formulas: g.targets(e, w, l.Number+" Conclusion", k.Conclusion.Walton), Note: k.Conclusion.Note, Status: Generated})
			}
			printed, labels := s.printedCQs(l)
			for _, label := range labels {
				id := printed + "/" + label
				var row *CQMapRow
				for i := range k.CQMap {
					if k.CQMap[i].CQ == label {
						row = &k.CQMap[i]
					}
				}
				var ts []Target
				var note string
				if row != nil {
					ts, note = g.targets(e, w, id, row.Walton), row.Note
				}
				if c := findCQ(e.CQs, id); c != nil {
					c.Targets = mergeTargets(c.Targets, ts)
					continue
				}
				e.CQs = append(e.CQs, CQ{ID: id, Book: &BookCQ{Entry: printed, Label: label}, Targets: ts, Note: note, Status: Generated})
			}
		}
	}
	// The game's critical questions, attached to the book question that
	// targets the same formula, or on their own.
	if e.Game != nil {
		for _, a := range s.AskEdges {
			if a.Scheme != e.Game.ID {
				continue
			}
			gc := g.gameCQ(a)
			if c := matchGame(e.CQs, gc); c != nil {
				c.Game = gc
				continue
			}
			e.CQs = append(e.CQs, gameOnly(gc, w))
		}
	}
	e.Families = g.families(w.ID, "")
	e.Contrasts = g.contrasts(w.ID, e.Families)
	if len(e.Book) == 0 {
		e.Discrepancies = append(e.Discrepancies, Discrepancy{Kind: "not-in-book", About: w.ID,
			Detail: "The compendium (ch. 9) lists no scheme this formalizes. " + cites(e.Walton.Source), Status: Generated})
	}
	e.Discrepancies = append(e.Discrepancies, g.gameDiffs(w)...)
	g.work(e)
	return e
}

// bookOnly is the entry of a compendium scheme that walton.yml lacks.
func (g *Generator) bookOnly(l LinkEntry) *Entry {
	e := &Entry{ID: slug(l.Name), Name: l.Name, Tier: 2}
	ref := BookRef{Entry: l.Number, Title: l.Name, Pages: l.BookPages, Match: "none", Status: Generated}
	if l.Parent != nil {
		ref.Parent = *l.Parent
	}
	if l.SharesCQs != nil && len(l.CQLabels) == 0 {
		ref.CQsPrintedUnder = *l.SharesCQs
	}
	e.Book = []BookRef{ref}
	printed, labels := g.Src.printedCQs(l)
	for _, label := range labels {
		e.CQs = append(e.CQs, CQ{ID: printed + "/" + label, Book: &BookCQ{Entry: printed, Label: label}, Status: Generated})
	}
	e.Discrepancies = []Discrepancy{{Kind: "no-formalization", About: l.Number,
		Detail: "No walton.yml scheme formalizes compendium " + l.Number + " (links.json).", Status: Generated}}
	e.Families = g.families("", l.Number)
	e.Contrasts = g.contrasts(e.ID, e.Families)
	g.work(e)
	return e
}

// work sets the work fields still to be written, and the history line.
func (g *Generator) work(e *Entry) {
	e.Summary = Text{Status: Todo}
	e.Class = Classification{Status: Todo}
	e.Specialises = Specialises{Status: Todo}
	e.DialogueTypes = DialogueTypes{Status: Todo}
	e.History = []Event{{Date: g.Today, By: g.By,
		What: "skeleton generated from walton.yml, links.json, families.json and the game's NASCQ"}}
}

func (g *Generator) formulas(w Scheme) *Formulas {
	return &Formulas{Variables: w.Variables, Premises: w.Premises, Assumptions: w.Assumptions, Exceptions: w.Exceptions, Conclusions: w.Conclusions}
}

// targets turns links.json's formulas into the scheme's formulas. A formula
// the scheme lacks, or one links.json puts in another role, is recorded as
// a discrepancy; the target keeps walton.yml's role.
func (g *Generator) targets(e *Entry, w Scheme, about string, fls []FormulaLocus) []Target {
	var out []Target
	for _, fl := range fls {
		claimed := normLocus(fl.Locus)
		fl.Formula = g.Src.edited(w.ID, fl.Formula)
		actual := w.locusOf(fl.Formula)
		switch {
		case actual == "":
			e.Discrepancies = append(e.Discrepancies, Discrepancy{Kind: "link-formula", About: about + " " + fl.Formula,
				Detail: fmt.Sprintf("links.json maps %s to %s (%s), which is not a formula of walton.yml's %s.", about, fl.Formula, fl.Locus, w.ID), Status: Generated})
			continue
		case actual != claimed:
			e.Discrepancies = append(e.Discrepancies, Discrepancy{Kind: "link-formula", About: about + " " + fl.Formula,
				Detail: fmt.Sprintf("For %s, links.json gives %s the role %q; in walton.yml it is a %s.", about, fl.Formula, fl.Locus, actual), Status: Generated})
		}
		out = mergeTargets(out, []Target{{Formula: canonical(w, fl.Formula), Locus: actual}})
	}
	return out
}

// canonical returns walton.yml's own spelling of a formula.
func canonical(w Scheme, formula string) string {
	for _, l := range Loci {
		for _, f := range w.role(l) {
			if compact(f) == compact(formula) {
				return f
			}
		}
	}
	return formula
}

func (g *Generator) gameCQ(a AskEdge) *GameCQ {
	return &GameCQ{ID: a.CQ, Text: g.Src.GameCQs[a.CQ].Template, Locus: a.Locus, Pattern: unbrace(a.Pattern)}
}

// gameOnly is the item of a game question that matches no book question.
func gameOnly(gc *GameCQ, w Scheme) CQ {
	c := CQ{ID: "game/" + gc.ID, Game: gc, Status: Generated}
	if gc.Locus == "argument" {
		c.Undercuts = true
	} else if gc.Pattern != "" {
		if l := w.locusOf(gc.Pattern); l != "" {
			c.Targets = []Target{{Formula: canonical(w, gc.Pattern), Locus: l}}
		}
	} else if gc.Locus == "conclusion" && len(w.Conclusions) > 0 {
		c.Targets = []Target{{Formula: w.Conclusions[0], Locus: "conclusion"}}
	}
	return c
}

// matchGame returns the book question a game question belongs to: the
// first, not yet matched, that targets the formula of the game's pattern.
func matchGame(cs []CQ, gc *GameCQ) *CQ {
	if gc.Pattern == "" {
		return nil
	}
	for i := range cs {
		c := &cs[i]
		if c.Book == nil || c.Game != nil {
			continue
		}
		for _, t := range c.Targets {
			if compact(t.Formula) == compact(gc.Pattern) {
				return c
			}
		}
	}
	return nil
}

func (g *Generator) families(id, compendium string) []FamilyRef {
	var out []FamilyRef
	for _, f := range g.Src.Families {
		for _, m := range f.Members {
			if (id != "" && m.Scheme == id) || (id == "" && m.Scheme == "" && m.Compendium == compendium) {
				out = append(out, FamilyRef{Family: f.ID, Name: f.Name, Relation: m.Relation, Note: m.Note, Status: Generated})
			}
		}
	}
	return out
}

// contrasts lists the close relatives to contrast the scheme with: its
// family's generic scheme, or, for the generic, every member (and every
// member of a family without a generic); and the schemes whose patterns the
// registry finds related. People add others as they draft.
func (g *Generator) contrasts(id string, fams []FamilyRef) []Contrast {
	var out []Contrast
	seen := map[string]bool{id: true}
	add := func(other string) {
		if other != "" && !seen[other] {
			seen[other] = true
			out = append(out, Contrast{Scheme: other, Status: Todo})
		}
	}
	for _, fr := range fams {
		for _, f := range g.Src.Families {
			if f.ID != fr.Family {
				continue
			}
			if f.Generic != "" && f.Generic != id {
				add(f.Generic)
				continue
			}
			for _, m := range f.Members {
				other := m.Scheme
				if other == "" {
					// a compendium scheme without walton.yml's: its own entry,
					// if links.json links it to nothing
					if l, ok := g.Src.link(m.Compendium); ok && len(l.Links) == 0 && !l.HeadingOnly {
						other = slug(l.Name)
					}
				}
				add(other)
			}
		}
	}
	for _, r := range g.Src.Close {
		switch id {
		case r.A:
			add(r.B)
		case r.B:
			add(r.A)
		}
	}
	return out
}

// gameDiffs records where the game's copy of the scheme (core.yml) differs
// from walton.yml.
func (g *Generator) gameDiffs(w Scheme) []Discrepancy {
	c, ok := g.Src.Core[w.ID]
	if !ok {
		return nil
	}
	var out []Discrepancy
	for _, l := range Loci {
		a, b := compactSet(w.role(l)), compactSet(c.role(l))
		for _, f := range w.role(l) {
			if !b[compact(f)] {
				out = append(out, Discrepancy{Kind: "game-differs", About: l + " " + f,
					Detail: fmt.Sprintf("The game's core.yml lacks the %s %s of walton.yml.", l, f), Status: Generated})
			}
		}
		for _, f := range c.role(l) {
			if !a[compact(f)] {
				out = append(out, Discrepancy{Kind: "game-differs", About: l + " " + f,
					Detail: fmt.Sprintf("The game's core.yml has the %s %s, which walton.yml lacks.", l, f), Status: Generated})
			}
		}
	}
	return out
}

func compactSet(fs []string) map[string]bool {
	m := map[string]bool{}
	for _, f := range fs {
		m[compact(f)] = true
	}
	return m
}

func findCQ(cs []CQ, id string) *CQ {
	for i := range cs {
		if cs[i].ID == id {
			return &cs[i]
		}
	}
	return nil
}

func mergeTargets(a, b []Target) []Target {
	for _, t := range b {
		dup := false
		for _, x := range a {
			if compact(x.Formula) == compact(t.Formula) {
				dup = true
			}
		}
		if !dup {
			a = append(a, t)
		}
	}
	return a
}

var nonWord = regexp.MustCompile(`[^a-z0-9]+`)

// slug makes an id from a compendium title: "Precedent Slippery Slope Argument"
// becomes "precedent_slippery_slope_argument".
func slug(name string) string {
	return strings.Trim(nonWord.ReplaceAllString(strings.ToLower(name), "_"), "_")
}

func sortStrings(s []string) { sort.Strings(s) }

// Merge combines an entry read from disk with a freshly generated one. Source
// fields come from the generated entry. A work item comes from the
// generated entry while the item on disk is still generated or todo, and is
// kept as it is once someone has drafted, verified or disputed it; items
// that only people added are kept. The game's part of each critical
// question is source, and is refreshed even in a kept item.
func Merge(old, gen *Entry) *Entry {
	if old == nil {
		return gen
	}
	e := *gen
	if old.Tier != 0 {
		e.Tier = old.Tier
	}
	if len(old.History) > 0 {
		e.History = old.History
	}
	e.Book = mergeItems(old.Book, gen.Book, func(b BookRef) string { return b.Entry }, func(b BookRef) string { return b.Status })
	e.Families = mergeItems(old.Families, gen.Families, func(f FamilyRef) string { return f.Family }, func(f FamilyRef) string { return f.Status })
	e.PremiseMap = mergeItems(old.PremiseMap, gen.PremiseMap, func(p PremiseLink) string { return p.Entry + "/" + p.Label }, func(p PremiseLink) string { return p.Status })
	e.Contrasts = mergeItems(old.Contrasts, gen.Contrasts, func(c Contrast) string { return c.Scheme }, func(c Contrast) string { return c.Status })
	e.Discrepancies = mergeItems(old.Discrepancies, gen.Discrepancies, func(d Discrepancy) string { return d.Kind + "|" + d.About }, func(d Discrepancy) string { return d.Status })
	e.Examples = old.Examples
	if !owned(old.Summary.Status) {
		e.Summary = old.Summary
	}
	if !owned(old.Class.Status) {
		e.Class = old.Class
	}
	if !owned(old.Specialises.Status) {
		e.Specialises = old.Specialises
	}
	if !owned(old.DialogueTypes.Status) {
		e.DialogueTypes = old.DialogueTypes
	}
	e.CQs = mergeCQs(old.CQs, gen.CQs)
	return &e
}

// mergeItems merges a list of work items by key: the generated list, with
// each item replaced by the one on disk once that one is no longer owned
// by the generator, followed by the kept items the generator no longer
// makes.
func mergeItems[T any](old, gen []T, key func(T) string, status func(T) string) []T {
	kept := map[string]T{}
	var order []string
	for _, o := range old {
		if !owned(status(o)) {
			k := key(o)
			if _, dup := kept[k]; !dup {
				order = append(order, k)
			}
			kept[k] = o
		}
	}
	var out []T
	used := map[string]bool{}
	for _, gi := range gen {
		k := key(gi)
		if o, ok := kept[k]; ok {
			out = append(out, o)
			used[k] = true
			continue
		}
		out = append(out, gi)
	}
	for _, k := range order {
		if !used[k] {
			out = append(out, kept[k])
		}
	}
	return out
}

// mergeCQs merges critical questions. A kept question keeps its link to a
// game question; the game question's own data (text, locus, pattern) is
// refreshed from the NASCQ. A game question kept elsewhere is not added
// again on its own.
func mergeCQs(old, gen []CQ) []CQ {
	games := map[string]*GameCQ{}
	var gameOrder []string
	for _, c := range gen {
		if c.Game != nil {
			games[c.Game.ID] = c.Game
			gameOrder = append(gameOrder, c.Game.ID)
		}
	}
	keptGame := map[string]bool{}
	for _, o := range old {
		if !owned(o.Status) && o.Game != nil {
			keptGame[o.Game.ID] = true
		}
	}
	var gen2 []CQ
	for _, c := range gen {
		if c.Game != nil && keptGame[c.Game.ID] {
			if c.Book == nil {
				continue // the game question lives in a kept item
			}
			c.Game = nil // its book question, generated, no longer carries it
		}
		gen2 = append(gen2, c)
	}
	out := mergeItems(old, gen2, func(c CQ) string { return c.ID }, func(c CQ) string { return c.Status })
	present := map[string]bool{}
	for i := range out {
		if out[i].Game != nil {
			if g, ok := games[out[i].Game.ID]; ok {
				out[i].Game = g
				present[g.ID] = true
			}
		}
	}
	// A game question whose generated book item was replaced by a kept item
	// without it stands on its own.
	for _, id := range gameOrder {
		if present[id] {
			continue
		}
		g := games[id]
		c := CQ{ID: "game/" + id, Game: g, Status: Generated, Undercuts: g.Locus == "argument"}
		if g.Pattern != "" && g.Locus != "argument" {
			c.Targets = []Target{{Formula: g.Pattern, Locus: g.Locus}}
		}
		out = append(out, c)
	}
	return out
}

// Header is the comment at the top of every entry file.
const Header = `# A scheme of WAG's catalogue (wag/data/catalogue/README.md).
# Source fields (walton, game, formulas, the game part of each question)
# are regenerated by "wag-cli --catalogue-gen"; do not edit them here.
# Every other item has a status; set it when you draft, verify or dispute.
`

// Marshal writes an entry as YAML, with the header.
func Marshal(e *Entry) ([]byte, error) {
	b, err := yaml.Marshal(e)
	if err != nil {
		return nil, err
	}
	return append([]byte(Header), b...), nil
}

// Unmarshal reads an entry, rejecting unknown keys.
func Unmarshal(b []byte) (*Entry, error) {
	var e Entry
	if err := yaml.UnmarshalStrict(b, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// SchemesDir is where the entries live, inside the catalogue's directory.
const SchemesDir = "schemes"

// ReadDir reads every entry of a catalogue directory, by file name.
func ReadDir(dir string) (map[string]*Entry, map[string]error, error) {
	files, err := filepath.Glob(filepath.Join(dir, SchemesDir, "*.yml"))
	if err != nil {
		return nil, nil, err
	}
	out, bad := map[string]*Entry{}, map[string]error{}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, nil, err
		}
		e, err := Unmarshal(b)
		if err != nil {
			bad[filepath.Base(f)] = err
			continue
		}
		out[filepath.Base(f)] = e
	}
	return out, bad, nil
}

// Result reports what Write did.
type Result struct {
	Created, Updated, Unchanged int
	Orphans                     []string // entry files the sources no longer make, left in place
}

// Write generates the catalogue into dir: it merges every generated entry
// with the file on disk, writes the files that change, and the index.
func (g *Generator) Write(dir string) (Result, error) {
	var r Result
	old, bad, err := ReadDir(dir)
	if err != nil {
		return r, err
	}
	if len(bad) > 0 {
		var names []string
		for n, e := range bad {
			names = append(names, fmt.Sprintf("%s: %v", n, e))
		}
		sort.Strings(names)
		return r, fmt.Errorf("cannot merge into files that do not read: %s", strings.Join(names, "; "))
	}
	if err := os.MkdirAll(filepath.Join(dir, SchemesDir), 0o755); err != nil {
		return r, err
	}
	made := map[string]bool{}
	var entries []*Entry
	for _, gen := range g.Build() {
		name := gen.ID + ".yml"
		made[name] = true
		e := Merge(old[name], gen)
		entries = append(entries, e)
		b, err := Marshal(e)
		if err != nil {
			return r, err
		}
		path := filepath.Join(dir, SchemesDir, name)
		cur, err := os.ReadFile(path)
		switch {
		case err == nil && bytes.Equal(cur, b):
			r.Unchanged++
			continue
		case err == nil:
			r.Updated++
		default:
			r.Created++
		}
		if err := os.WriteFile(path, b, 0o644); err != nil {
			return r, err
		}
	}
	for n := range old {
		if !made[n] {
			r.Orphans = append(r.Orphans, n)
		}
	}
	sort.Strings(r.Orphans)
	idx := Index(entries)
	if err := os.WriteFile(filepath.Join(dir, IndexFile), idx, 0o644); err != nil {
		return r, err
	}
	return r, nil
}

// cites says what walton.yml cites for a scheme.
func cites(source string) string {
	if strings.TrimSpace(source) == "" {
		return "walton.yml cites no source."
	}
	return "walton.yml cites: " + source
}
