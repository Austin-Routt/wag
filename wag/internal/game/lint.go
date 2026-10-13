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
//     exceptions (the question could never be answered there); a Language
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
			if p, a, ok := predicateOf(f); ok && a > 0 {
				used[fmt.Sprintf("%s/%d", p, a)] = true
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
