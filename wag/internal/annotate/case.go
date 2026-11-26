package annotate

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/yourorg/wag/data"
)

// Annotation formats (docs/annotation-records.md): format 1, the batch-2
// format of 2026-10-13, and format 2, the identity contract (D092).
const (
	Format1 = "wag-annotation/1"
	Format2 = "wag-annotation/2"
)

// formatNumber reads a format's name: 2 for wag-annotation/2, 1 for
// wag-annotation/1 or none.
func formatNumber(name string) (int, error) {
	switch name {
	case Format2:
		return 2, nil
	case Format1, "":
		return 1, nil
	}
	return 0, fmt.Errorf("format %q is not %s or %s", name, Format1, Format2)
}

// Case is a dataset's case record, case.json (docs/annotation-records.md
// §3): the contexts the discussion argues about, who can act in it, and the
// actions it deliberates, each with the one actor who would perform it and
// the other atoms that name it.
type Case struct {
	Format     string `json:"format"`
	ID         string `json:"id"`
	Discussion string `json:"discussion"`
	Title      string `json:"title"`
	Thesis     struct {
		Claim  string `json:"claim"`
		Action string `json:"action"`
	} `json:"thesis"`
	DefaultContext string `json:"default_context"`
	Contexts       []struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"contexts"`
	Actors []struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"actors"`
	Actions []struct {
		ID      string   `json:"id"`
		Actor   string   `json:"actor"`
		Aliases []string `json:"aliases"`
	} `json:"actions"`
	Status string `json:"status"`

	canon    map[string]string // an action or an alias → the action
	actorOf  map[string]string // an action → its actor
	actors   map[string]bool
	contexts map[string]bool
}

// parseCase reads and checks a case record: its ids used once, every
// action's actor declared, no atom both an action and an alias or the alias
// of two actions, its default context and its thesis's action declared.
func parseCase(b []byte) (*Case, error) {
	var c Case
	if err := checkedJSON(b, &c); err != nil {
		return nil, err
	}
	if c.Format != "wag-case/1" {
		return nil, fmt.Errorf("its format is %q, not wag-case/1", c.Format)
	}
	c.canon, c.actorOf, c.actors, c.contexts = map[string]string{}, map[string]string{}, map[string]bool{}, map[string]bool{}
	for _, x := range c.Contexts {
		if x.ID == "" || c.contexts[x.ID] {
			return nil, fmt.Errorf("a context has no id, or one already used (%q)", x.ID)
		}
		c.contexts[x.ID] = true
	}
	if !c.contexts[c.DefaultContext] {
		return nil, fmt.Errorf("its default context %q is not one of its contexts", c.DefaultContext)
	}
	for _, x := range c.Actors {
		if x.ID == "" || c.actors[x.ID] {
			return nil, fmt.Errorf("an actor has no id, or one already used (%q)", x.ID)
		}
		c.actors[x.ID] = true
	}
	for _, x := range c.Actions {
		if x.ID == "" || c.canon[x.ID] != "" {
			return nil, fmt.Errorf("an action has no id, or one already used (%q)", x.ID)
		}
		if c.actors[x.ID] {
			return nil, fmt.Errorf("%q is an actor, so it cannot be an action", x.ID)
		}
		if !c.actors[x.Actor] {
			return nil, fmt.Errorf("the action %s is performed by %q, not an actor of the case", x.ID, x.Actor)
		}
		c.canon[x.ID], c.actorOf[x.ID] = x.ID, x.Actor
	}
	for _, x := range c.Actions {
		if x.Aliases == nil {
			return nil, fmt.Errorf("the action %s has no list of aliases", x.ID)
		}
		for _, al := range x.Aliases {
			if al == "" || c.canon[al] != "" {
				return nil, fmt.Errorf("%q is an action or an alias already, so it cannot be an alias of %s", al, x.ID)
			}
			if c.actors[al] {
				return nil, fmt.Errorf("%q is an actor, so it cannot be an alias of %s", al, x.ID)
			}
			c.canon[al] = x.ID
		}
	}
	if c.actorOf[c.Thesis.Action] == "" {
		return nil, fmt.Errorf("the thesis's action %q is not an action of the case", c.Thesis.Action)
	}
	return &c, nil
}

// Canon is the action an atom names: itself, or the action it is an alias
// of.
func (c *Case) Canon(atom string) string {
	if c != nil {
		if a, ok := c.canon[atom]; ok {
			return a
		}
	}
	return atom
}

// actionPlaces are the places, by predicate and arity, of the formulas of an
// instance of a scheme that hold one of its action variables (typedVars), as
// zero-based argument positions: they cover the places positions.json does
// not type, such as sunk_costs(A,C)'s A, where the scheme says the variable is
// an action elsewhere. Nil for no scheme of the collection.
type actionPlaces map[string]map[int]bool

func schemeActionPlaces(sc waltonScheme) actionPlaces {
	acts := map[string]bool{}
	for v, kind := range typedVars(sc) {
		if kind == "action" {
			acts[v] = true
		}
	}
	out := actionPlaces{}
	for _, fs := range [][]string{sc.Premises, sc.Assumptions, sc.Exceptions, sc.Conclusions} {
		for _, f := range fs {
			name, args := splitFormula(f)
			key := fmt.Sprintf("%s/%d", name, len(args))
			for i, a := range args {
				if acts[a] {
					if out[key] == nil {
						out[key] = map[int]bool{}
					}
					out[key][i] = true
				}
			}
		}
	}
	return out
}

// canonFormula puts each alias in an action's place of a formula back to its
// action: the places positions.json types, by predicate and arity, nested
// terms included, and in an instance of a scheme of the collection, the
// places where the scheme has an action variable (extra). An atom in any other
// place keeps its name, even when it is spelled as an alias: a goal or a
// consequence is not an action. A formula with nothing to put back is
// returned as it is.
func (c *Case) canonFormula(f string, extra actionPlaces) string {
	if c == nil || len(c.canon) == 0 {
		return f
	}
	if out, changed := c.canonTerm(f, extra); changed {
		return out
	}
	return f
}

// canonTerm is canonFormula for one term; changed says whether it put an
// alias back.
func (c *Case) canonTerm(f string, extra actionPlaces) (string, bool) {
	g := strings.TrimSpace(f)
	if strings.HasPrefix(g, "¬") {
		y, changed := c.canonTerm(strings.TrimPrefix(g, "¬"), extra)
		if !changed {
			return f, false
		}
		return "¬" + y, true
	}
	name, args := splitFormula(g)
	if len(args) == 0 {
		return f, false
	}
	key := fmt.Sprintf("%s/%d", name, len(args))
	places := map[int]bool{}
	for _, i := range typedPlaces.Action[key] {
		places[i-1] = true
	}
	for i := range extra[key] {
		places[i] = true
	}
	changed := false
	for i, x := range args {
		switch {
		case strings.Contains(x, "("):
			if y, ch := c.canonTerm(x, extra); ch {
				args[i], changed = y, true
			}
		case places[i]:
			if y := c.Canon(x); y != x {
				args[i], changed = y, true
			}
		}
	}
	if !changed {
		return f, false
	}
	return name + "(" + strings.Join(args, ",") + ")", true
}

// instanceScheme is the scheme of the collection an instance names; false
// without the collection, or when the scheme is not one of its schemes
// (other).
func instanceScheme(inst map[string]any, col *Collection) (waltonScheme, bool) {
	if col == nil {
		return waltonScheme{}, false
	}
	sc, _ := inst["scheme"].(map[string]any)
	id, _ := sc["walton_id"].(string)
	s, ok := col.schemes[id]
	return s, ok
}

// actionVars are the variables of an instance's scheme that stand in an
// action's place; nil without the collection, or when the scheme is not
// one of its schemes (other).
func actionVars(inst map[string]any, col *Collection) map[string]bool {
	s, ok := instanceScheme(inst, col)
	if !ok {
		return nil
	}
	out := map[string]bool{}
	for v, kind := range typedVars(s) {
		if kind == "action" {
			out[v] = true
		}
	}
	return out
}

// Canonical returns a copy of an annotation in which each alias of an action
// is put back to its action where an action stands, so that agreement reads
// two names of one action as one (docs/annotation-records.md §11): in the
// binding of a variable that stands in an action's place in the instance's
// scheme (walton.yml, positions.json), and in an action's place of a
// formula: one positions.json types, or, in an instance's formulas, one where
// the instance's scheme has an action variable, so that a binding and the
// formulas it fills are read alike. An atom anywhere else keeps its name,
// even when it is spelled as an alias: a goal or a consequence written with
// an action's alias is not that action, and reading it as one would make two
// different annotations agree. Without the collection, the bindings are left
// as they are, and formulas are read by positions.json alone.
func (c *Case) Canonical(a Annotation, col *Collection) Annotation {
	if c == nil || len(c.canon) == 0 || a == nil {
		return a
	}
	var walk func(v any, key string, extra actionPlaces) any
	walk = func(v any, key string, extra actionPlaces) any {
		switch x := v.(type) {
		case map[string]any:
			out := make(map[string]any, len(x))
			var acts map[string]bool
			if _, ok := x["bindings"]; ok {
				acts = actionVars(x, col)
				extra = nil
				if s, ok := instanceScheme(x, col); ok {
					extra = schemeActionPlaces(s)
				}
			}
			for k, y := range x {
				if b, ok := y.(map[string]any); ok && k == "bindings" {
					nb := make(map[string]any, len(b))
					for v, val := range b {
						if s, ok := val.(string); ok && acts[v] {
							nb[v] = c.Canon(s)
						} else {
							nb[v] = val
						}
					}
					out[k] = nb
					continue
				}
				out[k] = walk(y, k, extra)
			}
			return out
		case []any:
			out := make([]any, len(x))
			for i, y := range x {
				out[i] = walk(y, key, extra)
			}
			return out
		case string:
			if key == "formula" {
				return c.canonFormula(x, extra)
			}
		}
		return v
	}
	out, _ := walk(map[string]any(a), "", nil).(map[string]any)
	return out
}

// CanonicalSet is Canonical for each annotation of a set.
func (c *Case) CanonicalSet(set map[string]Annotation, col *Collection) map[string]Annotation {
	if c == nil || len(c.canon) == 0 {
		return set
	}
	out := make(map[string]Annotation, len(set))
	for id, a := range set {
		out[id] = c.Canonical(a, col)
	}
	return out
}

// positions are the places of walton.yml's predicates that hold an agent or
// an action (wag/data/annotation/positions.json), keyed name/arity.
type positions struct {
	Agent  map[string][]int `json:"agent"`
	Action map[string][]int `json:"action"`
}

var typedPlaces = func() positions {
	var p positions
	b, err := fs.ReadFile(data.Annotation, "annotation/positions.json")
	if err == nil {
		err = json.Unmarshal(b, &p)
	}
	if err != nil || len(p.Agent) == 0 || len(p.Action) == 0 {
		panic(fmt.Sprintf("wag/data/annotation/positions.json: %v", err))
	}
	return p
}()

// typedVars gives the scheme's variables that stand in an agent's or an
// action's place ("agent", "action").
func typedVars(sc waltonScheme) map[string]string {
	out := map[string]string{}
	for _, fs := range [][]string{sc.Premises, sc.Assumptions, sc.Exceptions, sc.Conclusions} {
		for _, f := range fs {
			name, args := splitFormula(f)
			key := fmt.Sprintf("%s/%d", name, len(args))
			for _, t := range []struct {
				kind  string
				table map[string][]int
			}{{"agent", typedPlaces.Agent}, {"action", typedPlaces.Action}} {
				for _, i := range t.table[key] {
					if i-1 < len(args) {
						if _, ok := out[args[i-1]]; !ok {
							out[args[i-1]] = t.kind
						}
					}
				}
			}
		}
	}
	return out
}

// splitFormula reads "¬p(a,q(b),c)" as p and its top-level arguments, the
// name and each argument trimmed ("p (a)" is p, as formulas are compared
// without their spaces).
func splitFormula(f string) (string, []string) {
	f = strings.TrimPrefix(strings.TrimSpace(f), "¬")
	i := strings.Index(f, "(")
	if i < 0 || !strings.HasSuffix(f, ")") {
		return f, nil
	}
	var args []string
	depth, cur := 0, strings.Builder{}
	for _, r := range f[i+1 : len(f)-1] {
		switch {
		case r == ',' && depth == 0:
			args = append(args, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		case r == '(':
			depth++
		case r == ')':
			depth--
		}
		cur.WriteRune(r)
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		args = append(args, s)
	}
	return strings.TrimSpace(f[:i]), args
}
