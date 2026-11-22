package chr

import (
	"fmt"
	"math/big"
	"strings"

	. "github.com/hfried/GoCHR/src/engine/terms"
)

// Propagation-only programs under set semantics
// ---------------------------------------------
//
// When a rule store has set semantics (Unique) and every rule is a pure
// propagation rule (no deleted head, no guard), CHRsolver runs this solver
// instead of the general one. It computes the same fixpoint, the closure of
// the store under the rules, with two guarantees the general solver does not
// give:
//
//   - each rule fires at most once for each tuple of constraints (a real
//     propagation history), whatever the order of the heads and whatever other
//     constraints share a functor with a head. The general solver tracks what
//     it has tried by positions in its per-functor lists; a head that is a
//     variable, or a constraint of another arity under the same functor,
//     upsets that bookkeeping, and a rule then fires on the same constraints
//     again and again until the limit;
//   - a head that is a bare variable matches wherever it stands: heads are
//     matched in an order in which such a variable is bound first when another
//     head binds it, and otherwise it matches any constraint.
//
// By default heads match distinct constraints as in the general solver.
// ReuseConstraints instead permits repeated use of a fact in a conjunction.
// max bounds rule applications; MaxMatchSteps also bounds failed searches.
//
// The solver works in rounds (semi-naive evaluation): a round looks only for
// tuples with at least one constraint added since the previous round began,
// and every rule gets its turn in every round, so no rule can starve the rules
// after it.

// isPropagationProgram reports whether the propagation solver applies to rs.
func isPropagationProgram(rs *RuleStore) bool {
	if !rs.Unique {
		return false
	}
	for _, r := range rs.CHRruleStore {
		if len(r.delHead) != 0 || len(r.guard) != 0 || len(r.keepHead) == 0 {
			return false
		}
	}
	return true
}

// a head of a rule, prepared for matching
type propHead struct {
	pos   int       // position in the rule's keep head (for the history key)
	head  *Compound // the head as written
	varOf Variable  // for a head that is a bare variable: the variable
	isVar bool
}

// headOrder returns the heads of r in an order for matching: compound heads
// first, as written, then each bare-variable head once a preceding head binds
// its variable, then the bare-variable heads nothing binds.
func headOrder(r *chrRule) []propHead {
	var compound, vars []propHead
	for i, h := range r.keepHead {
		if h.Functor == "" && len(h.Args) == 1 && h.Args[0].Type() == VariableType {
			vars = append(vars, propHead{pos: i, head: h, varOf: h.Args[0].(Variable), isVar: true})
		} else {
			compound = append(compound, propHead{pos: i, head: h})
		}
	}
	order := append([]propHead{}, compound...)
	bound := map[string]bool{}
	for _, h := range compound {
		for _, v := range h.head.OccurVars() {
			bound[v.Name] = true
		}
	}
	for len(vars) != 0 {
		progress := false
		rest := vars[:0]
		for _, h := range vars {
			if bound[h.varOf.Name] {
				order = append(order, h)
				progress = true
			} else {
				rest = append(rest, h)
			}
		}
		vars = rest
		if !progress {
			// nothing binds these: each matches any constraint
			for _, h := range vars {
				order = append(order, h)
				bound[h.varOf.Name] = true
			}
			break
		}
	}
	return order
}

// constraints lists the constraints alive in the CHR store whose functor is f,
// or all of them when f is "".
func aliveConstraints(rs *RuleStore, f string) CList {
	var out CList
	add := func(a *argCHR) {
		for _, c := range a.varArg {
			if c != nil && !c.IsDeleted {
				out = append(out, c)
			}
		}
		for _, c := range a.noArg {
			if c != nil && !c.IsDeleted {
				out = append(out, c)
			}
		}
	}
	if f != "" {
		if a, ok := rs.CHRstore[f]; ok {
			add(a)
		}
		return out
	}
	for _, a := range rs.CHRstore {
		add(a)
	}
	return out
}

// matchHead matches one prepared head against a constraint.
func matchHead(h propHead, c *Compound, env Bindings) (Bindings, bool) {
	if !h.isVar {
		return Match(*h.head, *c, env)
	}
	if b, ok := GetBinding(h.varOf, env); ok {
		if a, isAtom := b.(Atom); isAtom {
			return env, c.Functor == string(a) && len(c.Args) == 0
		}
		return env, Equal(b, *c)
	}
	return AddBinding(h.varOf, *c, env), true
}

// candidateView snapshots the store at a rule's turn. Streaming the matches
// must not let a firing add candidates to that same turn; they enter the next
// round, as when all tuples were materialized before firing. At most one list
// per functor and one list of all constraints is cached.
type candidateView struct {
	rs     *RuleStore
	before *big.Int
	lists  map[string]CList
}

func (v *candidateView) list(f string) CList {
	if a, ok := v.lists[f]; ok {
		return a
	}
	a := aliveConstraints(v.rs, f)
	out := a[:0]
	for _, c := range a {
		if c.Id.Cmp(v.before) < 0 {
			out = append(out, c)
		}
	}
	v.lists[f] = out
	return out
}

func (v *candidateView) candidates(h propHead, env Bindings) CList {
	if !h.isVar {
		return v.list(h.head.Functor)
	}
	if b, ok := GetBinding(h.varOf, env); ok {
		switch t := b.(type) {
		case Compound:
			return v.list(t.Functor)
		case Atom:
			return v.list(string(t))
		}
		return nil
	}
	return v.list("")
}

// Each candidate visit counts, even when its bindings fail to join. A rule
// application limit alone cannot bound a search in which no tuple matches.
func matchStep(rs *RuleStore) bool {
	if rs.Stopped != nil || cancelled(rs) {
		return false
	}
	if rs.MaxMatchSteps > 0 && rs.matchSteps >= rs.MaxMatchSteps {
		rs.Stopped = fmt.Errorf("%w: %d match candidate visits", ErrLimit, rs.MaxMatchSteps)
		return false
	}
	rs.matchSteps++
	return true
}

type propMatch struct {
	env Bindings
	key string
}

// visitMatches streams tuples with at least one constraint whose id is at
// least from. It stops immediately when the visitor, cancellation or a budget
// asks it to stop. No array of the potentially combinatorial join is built.
func visitMatches(rs *RuleStore, r *chrRule, order []propHead, from *big.Int, yield func(propMatch) bool) bool {
	view := candidateView{rs: rs, before: new(big.Int).Set(rs.chrCounter), lists: map[string]CList{}}
	n := len(order)
	used := make([]*Compound, n)
	isNew := func(c *Compound) bool { return c.Id.Cmp(from) >= 0 }
	for i := 0; i < n; i++ {
		var walk func(k int, env Bindings) bool
		walk = func(k int, env Bindings) bool {
			if rs.Stopped != nil || cancelled(rs) {
				return false
			}
			if k == n {
				ids := make([]string, len(r.keepHead))
				for j, h := range order {
					ids[h.pos] = used[j].Id.String()
				}
				return yield(propMatch{env: env, key: strings.Join(ids, ",")})
			}
			for _, c := range view.candidates(order[k], env) {
				if !matchStep(rs) {
					return false
				}
				switch {
				case k == i && !isNew(c), k < i && isNew(c):
					continue
				}
				distinct := true
				if !rs.ReuseConstraints {
					for j := 0; j < k; j++ {
						if used[j] == c {
							distinct = false
							break
						}
					}
				}
				if !distinct {
					continue
				}
				env2, ok := matchHead(order[k], c, env)
				if !ok {
					continue
				}
				used[k] = c
				if !walk(k+1, env2) {
					return false
				}
			}
			return true
		}
		if !walk(0, nil) {
			return false
		}
	}
	return true
}

// propagationSolver computes the closure of the store under the rules.
func propagationSolver(rs *RuleStore, max int) {
	orders := make([][]propHead, len(rs.CHRruleStore))
	for i, r := range rs.CHRruleStore {
		orders[i] = headOrder(r)
	}
	history := map[string]bool{}
	apps := 0
	from := big.NewInt(0)
	for rs.Stopped == nil && rs.Result != RFalse {
		if cancelled(rs) {
			return
		}
		next := new(big.Int).Set(rs.chrCounter)
		fired := false
		for i, r := range rs.CHRruleStore {
			if !visitMatches(rs, r, orders[i], from, func(m propMatch) bool {
				key := fmt.Sprintf("%d:%s", r.id, m.key)
				if history[key] {
					return true
				}
				if max != 0 && apps >= max {
					rs.Stopped = fmt.Errorf("%w: %d rule applications", ErrLimit, max)
					return false
				}
				if cancelled(rs) {
					return false
				}
				history[key] = true
				rs.RenameRuleVars = <-Counter
				TraceHeadln(1, 1, "rule ", r.name, " fired ", r.keepHead.String(), " <=> ", r.body.String(), ".")
				fireRule(rs, r, m.env)
				apps++
				fired = true
				return rs.Stopped == nil && rs.Result != RFalse
			}) {
				return
			}
		}
		if !fired {
			return
		}
		from = next
	}
}
