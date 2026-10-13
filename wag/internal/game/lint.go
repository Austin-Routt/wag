package game

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	"github.com/carneades/carneades-4/src/engine/terms"
)

// LintData loads a Carneades scheme file and the three NASCQ files and
// reports what is wrong with them, alone (nascq.Lint) and together:
//
//   - errors: a NASCQ scheme whose carneades_id is not a scheme of the YAML;
//     NASCQ and Carneades variables that differ (a support move binds the
//     NASCQ variables, and InstantiateScheme needs exactly the Carneades
//     ones); a premise locus that names no premise or assumption of the
//     Carneades scheme, or an exception locus that names none of its
//     exceptions (the question could never be answered there); an
//     answer_variable that is not what the Carneades scheme concludes; a Language
//     template whose number of %s differs from its predicate's arity;
//   - warnings: a predicate used by a scheme with no Language template (its
//     statements show as Prolog terms, and Carneades prints "Functor not
//     defined" for each).
//
// A load failure is returned as the error.
func LintData(yamlBytes, schemesJSON, cqsJSON, edgesJSON []byte) ([]nascq.LintIssue, error) {
	ad, err := carneades.NewAdapterFromYAML(yamlBytes)
	if err != nil {
		return nil, err
	}
	n, err := nascq.LoadFromBytes(schemesJSON, cqsJSON, edgesJSON)
	if err != nil {
		return nil, err
	}
	out := n.Lint()
	add := func(isErr bool, where, format string, args ...interface{}) {
		out = append(out, nascq.LintIssue{Error: isErr, Where: where, Msg: fmt.Sprintf(format, args...)})
	}

	theory := ad.Graph.Theory
	caesSchemes := map[string][]string{} // scheme ID -> its premise and assumption formulas
	caesVariables := map[string][]string{}
	for _, s := range theory.ArgSchemes {
		caesSchemes[s.Id] = append(append([]string{}, s.Premises...), s.Assumptions...)
		caesVariables[s.Id] = s.Variables
	}
	for _, id := range nascq.SortedStringKeys(n.Schemes) {
		ns := n.Schemes[id]
		vars, ok := caesVariables[ns.CarneadesID]
		if !ok {
			if ns.CarneadesID != "" {
				add(true, id, "carneades_id %q is not a scheme of the YAML", ns.CarneadesID)
			}
			continue
		}
		caesVars := map[string]bool{}
		for _, v := range vars {
			caesVars[v] = true
			if _, ok := ns.Variables[v]; !ok {
				add(true, id, "Carneades scheme %s has variable %s, which the NASCQ scheme does not declare", ns.CarneadesID, v)
			}
		}
		for _, v := range nascq.SortedStringKeys(ns.Variables) {
			if !caesVars[v] {
				add(true, id, "variable %s is not a variable of Carneades scheme %s", v, ns.CarneadesID)
			}
		}
	}

	// An answer variable is what the Carneades scheme concludes: the variable
	// itself (S) or a one-place predicate of it (true_in_situation(B)).
	caesConclusions := map[string][]string{}
	for _, s := range theory.ArgSchemes {
		caesConclusions[s.Id] = s.Conclusions
	}
	for _, id := range nascq.SortedStringKeys(n.Schemes) {
		ns := n.Schemes[id]
		concl, known := caesConclusions[ns.CarneadesID]
		if ns.AnswerVariable == "" || !known {
			continue
		}
		isVar := func(t terms.Term) bool {
			v, ok := t.(terms.Variable)
			return ok && v.Name == ns.AnswerVariable
		}
		ok := false
		if len(concl) == 1 {
			if t, parsed := terms.ReadString(concl[0]); parsed {
				c, isCompound := t.(terms.Compound)
				ok = isVar(t) || isCompound && len(c.Args) == 1 && isVar(c.Args[0])
			}
		}
		if !ok {
			add(true, id, "answer_variable %s is not what Carneades scheme %s concludes (%v)", ns.AnswerVariable, ns.CarneadesID, concl)
		}
	}

	// A backup scheme without an answer variable should conclude the answer
	// it backs: otherwise an answering argument concludes the answer in place
	// of what its scheme concludes (its premises support the scheme's
	// conclusion, not the answer).
	for _, b := range n.BackupEdges {
		bs := n.Schemes[b.Scheme]
		if bs == nil || bs.AnswerVariable != "" || bs.Free {
			continue
		}
		bc := caesConclusions[bs.CarneadesID]
		if len(bc) != 1 {
			continue
		}
		for _, e := range n.AskEdges {
			if e.CQ != b.CQ {
				continue
			}
			ts := n.Schemes[e.Scheme]
			if ts == nil {
				continue
			}
			l := n.LocusOf(e.CQ, e.Scheme)
			var answer string
			switch l.Kind {
			case nascq.LocusPremise, nascq.LocusException:
				answer = placeholderVars(l.Pattern)
			case nascq.LocusConclusion:
				if c := caesConclusions[ts.CarneadesID]; len(c) == 1 {
					answer = renameVars(c[0], "T")
				}
			}
			if answer == "" || unifiesEitherWay(bc[0], answer) {
				continue
			}
			add(false, b.Scheme, "backs %s on %s, but concludes %s, not its answer (%s or its opposite): an answering argument concludes the answer instead",
				b.CQ, e.Scheme, bc[0], answer)
		}
	}

	// Every premise or exception locus names a formula of the Carneades scheme.
	caesExceptions := map[string][]string{}
	for _, s := range theory.ArgSchemes {
		caesExceptions[s.Id] = s.Exceptions
	}
	for _, sid := range nascq.SortedStringKeys(n.Schemes) {
		for _, q := range n.GetAskCQs(sid) {
			l := n.LocusOf(q.ID, sid)
			if _, known := caesVariables[n.Schemes[sid].CarneadesID]; !known {
				continue // the unknown carneades_id is reported above
			}
			var formulas []string
			switch l.Kind {
			case nascq.LocusPremise:
				formulas = caesSchemes[n.Schemes[sid].CarneadesID]
			case nascq.LocusException:
				formulas = caesExceptions[n.Schemes[sid].CarneadesID]
			default:
				continue
			}
			pred, arity, ok := predicateOf(l.Pattern)
			if !ok {
				add(true, q.ID, "pattern %q is not a predicate pattern", l.Pattern)
				continue
			}
			found := false
			for _, f := range formulas {
				if p, a, ok := predicateOf(f); ok && p == pred && a == arity {
					found = true
				}
			}
			if !found {
				what := "a premise or assumption"
				if l.Kind == nascq.LocusException {
					what = "an exception"
				}
				add(true, q.ID, "asked of %s: %s locus %s/%d is not %s of Carneades scheme %s",
					sid, l.Kind, pred, arity, what, n.Schemes[sid].CarneadesID)
			}
		}
	}

	used := map[string]bool{}
	for _, s := range theory.ArgSchemes {
		for _, f := range append(append(append(append([]string{}, s.Premises...), s.Assumptions...), s.Exceptions...), s.Conclusions...) {
			// A negation ¬p(...) reads through the "¬p/n" template, or the
			// generic "it is not the case that"; it needs p's.
			neg := strings.HasPrefix(f, "¬")
			if p, a, ok := predicateOf(strings.TrimPrefix(f, "¬")); ok && a > 0 {
				used[fmt.Sprintf("%s/%d", p, a)] = true
				if neg {
					used[fmt.Sprintf("¬%s/%d", p, a)] = true
				}
			}
		}
	}
	for _, spec := range nascq.SortedStringKeys(theory.Language) {
		tmpl := theory.Language[spec]
		i := strings.LastIndex(spec, "/")
		arity, err := strconv.Atoi(spec[i+1:])
		if i < 0 || err != nil {
			add(true, "language", "key %q is not predicate/arity", spec)
			continue
		}
		if c := strings.Count(strings.ReplaceAll(tmpl, "%%", ""), "%s"); c != arity {
			add(true, "language", "%s has %d %%s for %d arguments: %q", spec, c, arity, tmpl)
		}
	}
	missing := []string{}
	for spec := range used {
		if _, ok := theory.Language[spec]; !ok {
			missing = append(missing, spec)
		}
	}
	sort.Strings(missing)
	for _, spec := range missing {
		add(false, "language", "no template for %s", spec)
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].Error && !out[j].Error })
	return out, nil
}

// predicateOf returns the predicate and arity of a formula such as
// "possible(A)" or a pattern such as "possible({A})".
func predicateOf(formula string) (string, int, bool) {
	t, ok := terms.ReadString(placeholderToVar(formula))
	if !ok {
		return "", 0, false
	}
	f, ok := terms.Functor(t)
	if !ok {
		return "", 0, false
	}
	return f, terms.Arity(t), true
}

// placeholderToVar turns NASCQ {Var} placeholders into Prolog variables.
func placeholderToVar(s string) string {
	for _, v := range nascq.Placeholders(s) {
		s = strings.ReplaceAll(s, "{"+v+"}", "V_"+v)
	}
	return s
}

// placeholderVars turns a NASCQ pattern's {X} placeholders into variables
// QX (distinct from any scheme's variable names).
func placeholderVars(pattern string) string {
	var b strings.Builder
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '{' {
			if j := strings.IndexByte(pattern[i:], '}'); j > 0 {
				b.WriteString("Q" + pattern[i+1:i+j])
				i += j
				continue
			}
		}
		b.WriteByte(pattern[i])
	}
	return b.String()
}

// renameVars prefixes the variables of a formula, so that two schemes'
// formulas share none.
func renameVars(formula, prefix string) string {
	t, ok := terms.ReadString(formula)
	if !ok {
		return formula
	}
	var walk func(terms.Term) terms.Term
	walk = func(t terms.Term) terms.Term {
		switch x := t.(type) {
		case terms.Variable:
			return terms.NewVariable(prefix + x.Name)
		case terms.Compound:
			args := make([]terms.Term, len(x.Args))
			for i, a := range x.Args {
				args[i] = walk(a)
			}
			return terms.Compound{Functor: x.Functor, Args: args}
		}
		return t
	}
	return walk(t).String()
}

// unifiesEitherWay reports whether a scheme's conclusion can be the answer
// or its negation.
func unifiesEitherWay(conclusion, answer string) bool {
	c, ok := terms.ReadString(conclusion)
	if !ok {
		return true // reported elsewhere
	}
	neg := "¬" + answer
	if rest, isNeg := strings.CutPrefix(answer, "¬"); isNeg {
		neg = rest
	}
	for _, a := range []string{answer, neg} {
		if t, ok := terms.ReadString(a); ok {
			if _, ok := terms.Unify(c, t, nil); ok {
				return true
			}
		}
	}
	return false
}
