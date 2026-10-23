// Package registry keeps the argumentation schemes apart: it reduces each
// Carneades scheme to its pattern (its formulas by role, with the variables
// renamed canonically), finds schemes that repeat a pattern under another
// name or differ from another in a single formula, and checks the families
// the data declares (a generic scheme and its specializations, variants and
// subtypes).
//
// The aim (the owner's fifth point, 2026-10-13): any scheme may be added to
// the game, but no pattern twice under two names, and every close relative
// of a scheme declared as such. Predicate names carry the meaning, so two
// schemes are the same pattern only if they use the same predicates in the
// same way; the same shape with other predicates (position to know and
// expert opinion) is a family resemblance, not a duplicate.
package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
)

// Roles of the formulas of a scheme.
const (
	RoleConclusion = "conclusion"
	RolePremise    = "premise"
	RoleAssumption = "assumption"
	RoleException  = "exception"
)

var roles = []string{RoleConclusion, RolePremise, RoleAssumption, RoleException}

// Pattern is a scheme reduced to its structure: for each role, its formulas
// with the variables renamed V1, V2, ... in a canonical order, sorted. Two
// schemes with equal canonical forms are the same pattern, whatever they
// call their variables and however they order their premises.
type Pattern struct {
	ID        string              `json:"id"`
	Formulas  map[string][]string `json:"formulas"`  // role -> canonical formulas
	Canonical string              `json:"canonical"` // the formulas, role by role
	Signature string              `json:"signature"` // a short hash of Canonical
	// Predicates are the predicate/arity of each formula, by role: the
	// shape near-duplicates are compared on.
	Predicates map[string][]string `json:"predicates"`
	// Inference is the canonical form of the conclusion and premises alone
	// with every predicate written P/arity: how the variables flow from the
	// premises to the conclusion, whatever the predicates say. Schemes with
	// the same inference and the same conclusion predicate, differing in one
	// premise predicate, are the same form about another kind of source or
	// thing (expert opinion and position to know).
	Inference string `json:"inference"`
}

// formula is one formula of a scheme before canonical renaming.
type formula struct {
	role string
	term terms.Term
	skel string // the formula with every variable written "_"
}

// PatternOf reduces a Carneades scheme to its pattern. Formulas that do not
// parse are kept as text (and reported by the Carneades validator).
func PatternOf(s *caes.Scheme) Pattern {
	fs, forms, text := canonicalize(s)
	p := Pattern{ID: s.Id, Formulas: forms, Canonical: text, Predicates: map[string][]string{}, Inference: inferenceOf(s)}
	h := sha256.Sum256([]byte(text))
	p.Signature = hex.EncodeToString(h[:6])
	for _, f := range fs {
		p.Predicates[f.role] = append(p.Predicates[f.role], predicate(f.term))
	}
	for _, r := range roles {
		sort.Strings(p.Predicates[r])
	}
	return p
}

// canonicalize parses the scheme's formulas and writes them in canonical
// form: variables renamed V1, V2, ... in the order that gives the smallest
// text.
func canonicalize(s *caes.Scheme) ([]formula, map[string][]string, string) {
	var fs []formula
	for _, r := range roles {
		for _, src := range formulasOf(s, r) {
			t, ok := terms.ReadString(src)
			if !ok {
				t = terms.Atom("unparsed:" + src)
			}
			fs = append(fs, formula{role: r, term: t, skel: r + ":" + skeleton(t)})
		}
	}
	// Formulas sort by skeleton; within a group of equal skeletons the
	// variable numbering depends on their order, so every order of every
	// group is tried (up to a bound) and the smallest result kept.
	sort.SliceStable(fs, func(i, j int) bool { return fs[i].skel < fs[j].skel })
	var groups [][2]int
	for i := 0; i < len(fs); {
		j := i
		for j < len(fs) && fs[j].skel == fs[i].skel {
			j++
		}
		groups = append(groups, [2]int{i, j})
		i = j
	}
	best := ""
	var bestForms map[string][]string
	tries := 0
	var permute func(g int)
	permute = func(g int) {
		if tries > 5040 {
			return
		}
		if g == len(groups) {
			tries++
			forms, text := render(fs)
			if best == "" || text < best {
				best, bestForms = text, forms
			}
			return
		}
		lo, hi := groups[g][0], groups[g][1]
		permuteRange(fs[lo:hi], func() { permute(g + 1) })
	}
	permute(0)
	if bestForms == nil {
		bestForms, best = render(fs)
	}
	return fs, bestForms, best
}

// inferenceOf is the canonical form of the scheme's conclusion and premises
// with the predicates abstracted (see Pattern.Inference).
func inferenceOf(s *caes.Scheme) string {
	abstract := func(src string) string {
		t, ok := terms.ReadString(src)
		if !ok {
			return src
		}
		return termString(abstractTerm(t))
	}
	var concl, prem []string
	for _, f := range s.Conclusions {
		concl = append(concl, abstract(f))
	}
	for _, f := range s.Premises {
		prem = append(prem, abstract(f))
	}
	_, _, text := canonicalize(&caes.Scheme{Id: s.Id, Conclusions: concl, Premises: prem})
	return text
}

// abstractTerm writes a formula's predicate as p<arity>, keeping a negation
// ¬ and the arguments.
func abstractTerm(t terms.Term) terms.Term {
	c, ok := t.(terms.Compound)
	if !ok {
		return t
	}
	if c.Functor == "¬" && len(c.Args) == 1 {
		return terms.Compound{Functor: "¬", Args: []terms.Term{abstractTerm(c.Args[0])}}
	}
	args := make([]terms.Term, len(c.Args))
	copy(args, c.Args)
	return terms.Compound{Functor: fmt.Sprintf("p%d", len(c.Args)), Args: args}
}

func termString(t terms.Term) string { return t.String() }

func formulasOf(s *caes.Scheme, role string) []string {
	switch role {
	case RoleConclusion:
		return s.Conclusions
	case RolePremise:
		return s.Premises
	case RoleAssumption:
		return s.Assumptions
	}
	return s.Exceptions
}

// permuteRange calls f once for every order of xs (Heap's algorithm), and
// leaves xs as it found it.
func permuteRange(xs []formula, f func()) {
	n := len(xs)
	if n <= 1 {
		f()
		return
	}
	c := make([]int, n)
	f()
	for i := 0; i < n; {
		if c[i] < i {
			if i%2 == 0 {
				xs[0], xs[i] = xs[i], xs[0]
			} else {
				xs[c[i]], xs[i] = xs[i], xs[c[i]]
			}
			f()
			c[i]++
			i = 0
		} else {
			c[i] = 0
			i++
		}
	}
	// Heap's algorithm ends on a permutation; restore the skeleton order.
	sort.SliceStable(xs, func(i, j int) bool { return xs[i].skel < xs[j].skel })
}

// render renames the variables V1, V2, ... in order of first occurrence and
// writes the formulas role by role.
func render(fs []formula) (map[string][]string, string) {
	names := map[string]string{}
	forms := map[string][]string{}
	for _, f := range fs {
		forms[f.role] = append(forms[f.role], rename(f.term, names))
	}
	var b strings.Builder
	for _, r := range roles {
		sort.Strings(forms[r])
		fmt.Fprintf(&b, "%s: %s\n", r, strings.Join(forms[r], "; "))
	}
	return forms, b.String()
}

func rename(t terms.Term, names map[string]string) string {
	switch x := t.(type) {
	case terms.Variable:
		n, ok := names[x.Name]
		if !ok {
			n = fmt.Sprintf("V%d", len(names)+1)
			names[x.Name] = n
		}
		return n
	case terms.Compound:
		args := make([]string, len(x.Args))
		for i, a := range x.Args {
			args[i] = rename(a, names)
		}
		return x.Functor + "(" + strings.Join(args, ",") + ")"
	}
	return t.String()
}

func skeleton(t terms.Term) string {
	switch x := t.(type) {
	case terms.Variable:
		return "_"
	case terms.Compound:
		args := make([]string, len(x.Args))
		for i, a := range x.Args {
			args[i] = skeleton(a)
		}
		return x.Functor + "(" + strings.Join(args, ",") + ")"
	}
	return t.String()
}

// predicate is predicate/arity, "¬p/1" for a negation ¬p(...), and "VAR" for
// a bare variable (a scheme that concludes any statement).
func predicate(t terms.Term) string {
	switch x := t.(type) {
	case terms.Variable:
		return "VAR"
	case terms.Compound:
		if x.Functor == "¬" && len(x.Args) == 1 {
			return "¬" + predicate(x.Args[0])
		}
		return fmt.Sprintf("%s/%d", x.Functor, len(x.Args))
	}
	return t.String() + "/0"
}

// Relation is how two schemes' patterns relate.
type Relation struct {
	A, B string
	// Kind is "identical" (the same pattern), "roles" (the same formulas,
	// some in another role, as a premise and an assumption), "one-formula"
	// (the same but for one formula: a variant), "extends" (B has every
	// formula of A and more: B specializes A), or "close" (the same
	// conclusion and at least CloseShare of their formulas in common).
	Kind   string
	Detail string
}

// Compare finds the related pairs among the patterns.
func Compare(ps []Pattern) []Relation {
	var out []Relation
	sort.Slice(ps, func(i, j int) bool { return ps[i].ID < ps[j].ID })
	for i := range ps {
		for j := i + 1; j < len(ps); j++ {
			if r, ok := relate(ps[i], ps[j]); ok {
				out = append(out, r)
			}
		}
	}
	return out
}

func relate(a, b Pattern) (Relation, bool) {
	if a.Canonical == b.Canonical {
		return Relation{A: a.ID, B: b.ID, Kind: "identical", Detail: "the same formulas, up to the names of the variables"}, true
	}
	all := func(p Pattern) []string {
		var out []string
		for _, r := range roles {
			out = append(out, p.Predicates[r]...)
		}
		sort.Strings(out)
		return out
	}
	if strings.Join(all(a), " ") == strings.Join(all(b), " ") && conclusionsEqual(a, b) {
		return Relation{A: a.ID, B: b.ID, Kind: "roles", Detail: "the same predicates, some in another role"}, true
	}
	onlyA, onlyB := diff(a, b)
	if !conclusionsEqual(a, b) {
		return Relation{}, false // a different conclusion is a different scheme
	}
	switch {
	case len(onlyA) == 1 && len(onlyB) == 1:
		return Relation{A: a.ID, B: b.ID, Kind: "one-formula", Detail: fmt.Sprintf("%s instead of %s", onlyB[0], onlyA[0])}, true
	case len(onlyA) == 0 && len(onlyB) > 0:
		return Relation{A: a.ID, B: b.ID, Kind: "extends", Detail: b.ID + " adds " + strings.Join(onlyB, ", ")}, true
	case len(onlyB) == 0 && len(onlyA) > 0:
		return Relation{A: b.ID, B: a.ID, Kind: "extends", Detail: a.ID + " adds " + strings.Join(onlyA, ", ")}, true
	}
	if a.Inference == b.Inference && premiseDiff(a, b) == 1 {
		onlyA, onlyB := diffRole(a, b, RolePremise)
		return Relation{A: a.ID, B: b.ID, Kind: "same-form", Detail: fmt.Sprintf("the same inference with %s instead of %s", strings.Join(onlyB, ", "), strings.Join(onlyA, ", "))}, true
	}
	na, nb := size(a), size(b)
	shared := (na + nb - len(onlyA) - len(onlyB)) / 2
	if union := na + nb - shared; union > 0 && float64(shared)/float64(union) >= CloseShare {
		return Relation{A: a.ID, B: b.ID, Kind: "close", Detail: fmt.Sprintf("%d of %d formulas shared; %s has %s; %s has %s", shared, union,
			a.ID, strings.Join(onlyA, ", "), b.ID, strings.Join(onlyB, ", "))}, true
	}
	return Relation{}, false
}

// Relate tells how two patterns relate, as Compare does for each pair.
func Relate(a, b Pattern) (Relation, bool) { return relate(a, b) }

// Distance is how far apart two patterns are: the role:predicate items of
// each, the conclusion included, that the other lacks. An added or dropped
// formula counts one, a changed formula two (one dropped, one added). It
// also returns the items only a has and only b has. It compares predicates
// only: two patterns with the same predicates in each role are 0 apart even
// when their variables flow differently (their Canonical forms differ).
func Distance(a, b Pattern) (int, []string, []string) {
	onlyA, onlyB := diff(a, b)
	ca, cb := diffRole(a, b, RoleConclusion)
	for _, p := range ca {
		onlyA = append(onlyA, RoleConclusion+" "+p)
	}
	for _, p := range cb {
		onlyB = append(onlyB, RoleConclusion+" "+p)
	}
	return len(onlyA) + len(onlyB), onlyA, onlyB
}

// CloseShare is the share of formulas (shared over all, by role and
// predicate) from which two schemes with the same conclusion are reported
// as close.
var CloseShare = 0.45

func size(p Pattern) int {
	n := 0
	for _, r := range roles[1:] {
		n += len(p.Predicates[r])
	}
	return n
}

// premiseDiff is the number of premise predicates of a that b lacks.
func premiseDiff(a, b Pattern) int {
	onlyA, _ := diffRole(a, b, RolePremise)
	return len(onlyA)
}

func diffRole(a, b Pattern, role string) ([]string, []string) {
	count := map[string]int{}
	for _, p := range a.Predicates[role] {
		count[p]++
	}
	for _, p := range b.Predicates[role] {
		count[p]--
	}
	var onlyA, onlyB []string
	for _, k := range sortedKeys(count) {
		for n := count[k]; n > 0; n-- {
			onlyA = append(onlyA, k)
		}
		for n := count[k]; n < 0; n++ {
			onlyB = append(onlyB, k)
		}
	}
	return onlyA, onlyB
}

func conclusionsEqual(a, b Pattern) bool {
	return strings.Join(a.Predicates[RoleConclusion], " ") == strings.Join(b.Predicates[RoleConclusion], " ")
}

// diff returns the role:predicate items of a not in b and of b not in a
// (multisets).
func diff(a, b Pattern) ([]string, []string) {
	count := map[string]int{}
	for _, r := range roles[1:] {
		for _, p := range a.Predicates[r] {
			count[r+" "+p]++
		}
		for _, p := range b.Predicates[r] {
			count[r+" "+p]--
		}
	}
	var onlyA, onlyB []string
	for _, k := range sortedKeys(count) {
		for n := count[k]; n > 0; n-- {
			onlyA = append(onlyA, k)
		}
		for n := count[k]; n < 0; n++ {
			onlyB = append(onlyB, k)
		}
	}
	return onlyA, onlyB
}

func sortedKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Families is the file that declares scheme families.
type Families struct {
	Note     string   `json:"_note,omitempty"`
	Families []Family `json:"families"`
}

// Family is a generic scheme with its relatives.
type Family struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Basis   []string `json:"basis,omitempty"` // where the grouping comes from (the book, the scheme files)
	Generic string   `json:"generic,omitempty"`
	Members []Member `json:"members"`
}

// Member is a scheme of a family. Relation is one of generic,
// specialization (it adds to the generic scheme or narrows it), variant (a
// sibling that differs in a formula), subtype (the book lists it as one),
// counterpart (the same form with the opposite polarity, as negative to
// positive consequences), context-specific (the generic scheme as it is
// used in one kind of dialogue or domain), related, or alias (the same
// scheme under another name: to be merged). Scheme is "" for a scheme of the
// book that no scheme file formalizes yet (Compendium names it).
type Member struct {
	Scheme     string `json:"scheme"`
	Relation   string `json:"relation"`
	Compendium string `json:"compendium,omitempty"` // the scheme's number in the book's compendium
	Note       string `json:"note,omitempty"`
}

var relations = map[string]bool{"generic": true, "specialization": true, "variant": true, "subtype": true, "counterpart": true,
	"context-specific": true, "related": true, "alias": true}

// LoadFamilies reads a families file.
func LoadFamilies(b []byte) (*Families, error) {
	var f Families
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("families: %w", err)
	}
	return &f, nil
}

// Finding is something the registry check reports.
type Finding struct {
	Error bool   `json:"error"`
	Where string `json:"where"`
	Msg   string `json:"msg"`
}

func (f Finding) String() string {
	kind := "warning"
	if f.Error {
		kind = "ERROR"
	}
	return fmt.Sprintf("%s %s: %s", kind, f.Where, f.Msg)
}

// Check reduces the schemes to patterns, compares them, and checks the
// relations against the families (fam may be nil):
//   - error: two schemes with the same pattern, unless declared aliases;
//   - error: a family names a scheme the file does not have (only when
//     strict: a families file may cover more schemes than one scheme file);
//   - error: an unknown relation, or a family without its generic scheme;
//   - warning: two schemes in a "roles", "one-formula", "extends" or "close" relation
//     that no family declares together;
//   - warning: a scheme no family declares (only when strict).
func Check(schemes []*caes.Scheme, fam *Families, strict bool) ([]Pattern, []Relation, []Finding) {
	var ps []Pattern
	have := map[string]bool{}
	for _, s := range schemes {
		ps = append(ps, PatternOf(s))
		have[s.Id] = true
	}
	rels := Compare(ps)
	var out []Finding
	add := func(isErr bool, where, format string, args ...interface{}) {
		out = append(out, Finding{Error: isErr, Where: where, Msg: fmt.Sprintf(format, args...)})
	}
	together := map[[2]string]string{} // pair -> the relation a family declares
	declared := map[string]bool{}
	if fam != nil {
		for _, f := range fam.Families {
			if f.Generic != "" && !have[f.Generic] && strict {
				add(true, f.ID, "its generic scheme %s is not in the file", f.Generic)
			}
			for _, m := range f.Members {
				if !relations[m.Relation] {
					add(true, f.ID, "%s: unknown relation %q", m.Scheme, m.Relation)
				}
				if m.Scheme == "" {
					continue // a scheme of the book only
				}
				if !have[m.Scheme] {
					if strict {
						add(true, f.ID, "member %s is not in the file", m.Scheme)
					}
					continue
				}
				declared[m.Scheme] = true
				for _, n := range f.Members {
					if n.Scheme != m.Scheme {
						together[pair(m.Scheme, n.Scheme)] = m.Relation
						if m.Relation == "alias" || n.Relation == "alias" {
							together[pair(m.Scheme, n.Scheme)] = "alias"
						}
					}
				}
			}
		}
	}
	for _, r := range rels {
		decl, ok := together[pair(r.A, r.B)]
		switch {
		case r.Kind == "identical" && decl != "alias":
			add(true, r.A+" = "+r.B, "the same pattern under two names (%s): merge them, or declare them aliases", r.Detail)
		case r.Kind != "identical" && !ok:
			add(false, r.A+" ~ "+r.B, "%s (%s), but no family declares them together", r.Kind, r.Detail)
		}
	}
	if strict {
		for _, p := range ps {
			if !declared[p.ID] {
				add(false, p.ID, "in no family")
			}
		}
	}
	return ps, rels, out
}

func pair(a, b string) [2]string {
	if a > b {
		a, b = b, a
	}
	return [2]string{a, b}
}
