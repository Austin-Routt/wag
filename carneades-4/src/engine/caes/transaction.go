package caes

import "github.com/carneades/carneades-4/src/engine/terms"

// copyForInference copies mutable graph topology. Theory, metadata and
// weighing functions are shared read-only; inference does not modify them.
func (ag *ArgGraph) copyForInference() *ArgGraph {
	w := *ag
	ss := map[*Statement]*Statement{}
	aa := map[*Argument]*Argument{}
	ii := map[*Issue]*Issue{}
	var stmt func(*Statement) *Statement
	var arg func(*Argument) *Argument
	var issue func(*Issue) *Issue
	stmt = func(s *Statement) *Statement {
		if s == nil {
			return nil
		}
		if c, ok := ss[s]; ok {
			return c
		}
		c := *s
		ss[s] = &c
		c.Issue = issue(s.Issue)
		c.Args = make([]*Argument, len(s.Args))
		for j, a := range s.Args {
			c.Args[j] = arg(a)
		}
		return &c
	}
	arg = func(a *Argument) *Argument {
		if a == nil {
			return nil
		}
		if c, ok := aa[a]; ok {
			return c
		}
		c := *a
		aa[a] = &c
		c.Parameters = append([]string(nil), a.Parameters...)
		c.Premises = append([]Premise(nil), a.Premises...)
		for j, p := range a.Premises {
			c.Premises[j].Stmt = stmt(p.Stmt)
		}
		c.Conclusion, c.Undercutter = stmt(a.Conclusion), stmt(a.Undercutter)
		return &c
	}
	issue = func(i *Issue) *Issue {
		if i == nil {
			return nil
		}
		if c, ok := ii[i]; ok {
			return c
		}
		c := *i
		ii[i] = &c
		c.Positions = make([]*Statement, len(i.Positions))
		for j, p := range i.Positions {
			c.Positions[j] = stmt(p)
		}
		return &c
	}
	w.Statements = map[string]*Statement{}
	w.Arguments = map[string]*Argument{}
	w.Issues = map[string]*Issue{}
	for k, s := range ag.Statements {
		w.Statements[k] = stmt(s)
	}
	for k, a := range ag.Arguments {
		w.Arguments[k] = arg(a)
	}
	for k, i := range ag.Issues {
		w.Issues[k] = issue(i)
	}
	w.Assumptions = append([]string(nil), ag.Assumptions...)
	w.Questioned = append([]string(nil), ag.Questioned...)
	w.Defaults = map[string]bool{}
	for k, v := range ag.Defaults {
		w.Defaults[k] = v
	}
	w.assums = map[string]bool{}
	for k, v := range ag.assums {
		w.assums[k] = v
	}
	return &w
}

// commitInference preserves the identities of pre-existing graph objects.
// Callers holding a statement, issue or argument pointer continue to refer to
// that same object after a successful inference. No writes occur on error.
func (ag *ArgGraph) commitInference(w *ArgGraph) {
	ss := map[*Statement]*Statement{}
	aa := map[*Argument]*Argument{}
	ii := map[*Issue]*Issue{}
	// A successful private normalization may change a key without changing
	// the user's statement object. normalizeGraph has already refused any
	// collision between distinct objects, so this correspondence is unique.
	originalByKey := map[string]*Statement{}
	for key, stmt := range ag.Statements {
		originalByKey[terms.Normalize(key)] = stmt
	}
	for k, s := range w.Statements {
		d := ag.Statements[k]
		if d == nil {
			d = originalByKey[k]
		}
		if d == nil {
			d = s
		}
		ss[s] = d
	}
	for k, a := range w.Arguments {
		d := ag.Arguments[k]
		if d == nil {
			d = a
		}
		aa[a] = d
	}
	for k, i := range w.Issues {
		d := ag.Issues[k]
		if d == nil {
			d = i
		}
		ii[i] = d
	}
	statement := func(s *Statement) *Statement {
		if d, ok := ss[s]; ok {
			return d
		}
		return s
	}
	argument := func(a *Argument) *Argument {
		if d, ok := aa[a]; ok {
			return d
		}
		return a
	}
	issue := func(i *Issue) *Issue {
		if d, ok := ii[i]; ok {
			return d
		}
		return i
	}
	for s, d := range ss {
		if s == nil {
			continue
		}
		c := *s
		c.Issue = issue(s.Issue)
		c.Args = make([]*Argument, len(s.Args))
		for j, a := range s.Args {
			c.Args[j] = argument(a)
		}
		*d = c
	}
	for a, d := range aa {
		if a == nil {
			continue
		}
		c := *a
		c.Conclusion = statement(a.Conclusion)
		c.Undercutter = statement(a.Undercutter)
		c.Premises = append([]Premise(nil), a.Premises...)
		for j, p := range a.Premises {
			c.Premises[j].Stmt = statement(p.Stmt)
		}
		*d = c
	}
	for i, d := range ii {
		if i == nil {
			continue
		}
		c := *i
		c.Positions = make([]*Statement, len(i.Positions))
		for j, p := range i.Positions {
			c.Positions[j] = statement(p)
		}
		*d = c
	}
	for k, s := range w.Statements {
		w.Statements[k] = statement(s)
	}
	for k, a := range w.Arguments {
		w.Arguments[k] = argument(a)
	}
	for k, i := range w.Issues {
		w.Issues[k] = issue(i)
	}
	*ag = *w
}
