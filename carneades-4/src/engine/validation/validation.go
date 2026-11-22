// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Validator for Carneades Argument Evaluation Structures (CAES)
// The validator checks for syntactic and semantic errors in
// CAES source files, represented using YAML, and produces
// an error report.

package validation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
)

// The category of the problem
type Category int

const (
	IMPORT Category = iota
	STATEMENT
	ISSUE
	ARGUMENT
	ASSUMPTION
	EXPECTEDLABELING
	LANGUAGE
	SCHEME  // Argument Scheme
	ISCHEME // Issue Scheme
)

var builtins map[string]bool = map[string]bool{
	"¬/1":  true,
	"=/2":  true,
	"==/2": true,
	"!=":   true,
	"<":    true,
	"<=":   true,
	">":    true,
	">=":   true,
	"=<":   true,
}

func (c Category) String() string {
	switch c {
	case IMPORT:
		return "import"
	case STATEMENT:
		return "statement"
	case ISSUE:
		return "issue"
	case ARGUMENT:
		return "argument"
	case ASSUMPTION:
		return "assumption"
	case EXPECTEDLABELING:
		return "expected labeling"
	case LANGUAGE:
		return "language"
	case SCHEME:
		return "argument scheme"
	case ISCHEME:
		return "issue scheme"
	default:
		return ""
	}

}

// Problem represents an error in some Carneades source file
type Problem struct {
	Category    Category
	Id          string // id of the affected object, if available
	Description string // brief description of the problem, without referencing the category or object id
	Expression  string // the affected part of the object with the problem.
}

// Validate the statements of an argument graph
func validateStatements(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	seen := map[string]*caes.Statement{}
	for k, _ := range ag.Statements {
		n := terms.Normalize(k)
		if old, exists := seen[n]; exists && old != ag.Statements[k] {
			problems = append(problems, Problem{STATEMENT, "", "multiple statement objects normalize to the same formula", n})
		}
		seen[n] = ag.Statements[k]

		// Check that the key is a term
		t, ok := terms.ReadString(k)
		if !ok {
			p := Problem{STATEMENT, "", "key not a term", k}
			problems = append(problems, p)
		} else {
			// Check that the term is a ground atomic formula
			var b terms.Bindings // empty environment
			if !terms.AtomicFormula(t) || !terms.Ground(t, b) {
				p := Problem{STATEMENT, "", "key not a ground atomic formula", k}
				problems = append(problems, p)
			}
		}
	}
	return problems
}

// Validate the issues of an argument graph
func validateIssues(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	for i1, v1 := range ag.Issues {
		// check that every statement is a position of at most one issue,
		// that is, that no position of an issue is also a position of
		// some other issue
		for _, s1 := range v1.Positions {
			for i2, v2 := range ag.Issues {
				if i2 != i1 { // if not the same issue
					for _, s2 := range v2.Positions {
						if s1 == s2 {
							// found s1 to be a position in both i1 and i2
							p := Problem{ISSUE, i1, "statement is a position of two issues", s1.Id}
							problems = append(problems, p)
						}
					}
				}
			}
		}
	}
	return problems
}

// Validate the arguments of an argument graph. An argument that names a
// scheme must instantiate it with one substitution: the scheme's variables
// bound to the argument's parameters, and under those bindings each premise
// an instance of one of the scheme's premises or assumptions (each used once)
// and the conclusion an instance of one of its conclusions.
func validateArguments(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	for id, arg := range ag.Arguments {
		if arg == nil || arg.Scheme == nil || caes.IsBasicScheme(arg.Scheme) {
			// arguments of the basic schemes (linked, convergent, ...) and
			// undercutters without a scheme have no pattern to check
			continue
		}
		if s := arg.Scheme; len(s.Premises) == 0 && len(s.Assumptions) == 0 && len(s.Conclusions) == 0 && len(s.Deletions) == 0 {
			// a scheme that only gives a weighing function (upstream's
			// rebuttal, lex and mcda examples) has no pattern either
			continue
		}
		s := arg.Scheme
		if len(arg.Parameters) == 0 && len(s.Variables) != 0 {
			// no parameters written: the substitution is the one that makes
			// the conclusion, premises and exception links instances of the
			// scheme. Try every matching substitution: a symmetric premise
			// list can have more than one, and its exception links must agree
			// with the same one.
			validInstance := false
			validExceptions := instanceWithoutParameters(s, arg, func(b terms.Bindings) bool {
				validInstance = true
				return carriesExceptions(ag, arg, b)
			})
			if !validInstance {
				p := Problem{ARGUMENT, id, "premises and conclusion do not instantiate the scheme with one substitution", ""}
				problems = append(problems, p)
			} else if !validExceptions {
				p := exceptionProblem(id, s)
				problems = append(problems, p)
			}
			continue
		}
		if len(arg.Parameters) != len(s.Variables) {
			p := Problem{ARGUMENT, id, "number of parameters not the same as declared in the scheme", ""}
			problems = append(problems, p)
			continue
		}
		// the substitution given by the parameters
		var b terms.Bindings
		bound := true
		for i, name := range s.Variables {
			v, ok1 := terms.ReadFormula(name)
			t, ok2 := terms.ReadFormula(arg.Parameters[i])
			if !ok1 || !ok2 || v.Type() != terms.VariableType {
				p := Problem{ARGUMENT, id, "parameter or scheme variable is not a term", name + " = " + arg.Parameters[i]}
				problems = append(problems, p)
				bound = false
				break
			}
			b = terms.AddBinding(v.(terms.Variable), t, b)
		}
		if !bound {
			continue
		}

		// premises: the scheme's premises and assumptions, each used once
		patterns := []terms.Term{}
		for _, ps := range append(append([]string{}, s.Premises...), s.Assumptions...) {
			if t, ok := terms.ReadFormula(ps); ok {
				patterns = append(patterns, t) // patterns of schemes are checked elsewhere
			}
		}
		stmts := []terms.Term{}
		for _, pr := range arg.Premises {
			if pr.Stmt == nil {
				problems = append(problems, Problem{ARGUMENT, id, "premise without a statement", ""})
				continue
			}
			t, ok := terms.ReadFormula(pr.Stmt.Id)
			if !ok {
				problems = append(problems, Problem{ARGUMENT, id, "premise is not a term", pr.Stmt.Id})
				continue
			}
			stmts = append(stmts, t)
		}
		if len(arg.Premises) != len(s.Premises)+len(s.Assumptions) {
			p := Problem{ARGUMENT, id, "number of premises not the same as declared in the scheme, including assumptions", ""}
			problems = append(problems, p)
		} else if len(stmts) == len(patterns) && !matchEach(patterns, stmts, b, make([]bool, len(patterns))) {
			p := Problem{ARGUMENT, id, "premises do not instantiate the scheme's premises and assumptions with the argument's parameters", ""}
			problems = append(problems, p)
		}

		// the scheme's exceptions are critical questions the argument has to
		// carry: a nonnil undercutter alone is not enough. Every exception
		// must have a direct linked argument for that undercutter, under the
		// same substitution (docs/contracts.md, contract 6). Report missing
		// links; never complete or rewrite an authored graph.
		if !carriesExceptions(ag, arg, b) {
			p := exceptionProblem(id, s)
			problems = append(problems, p)
		}

		// conclusion
		if arg.Conclusion == nil {
			problems = append(problems, Problem{ARGUMENT, id, "argument without a conclusion", ""})
			continue
		}
		t3, ok := terms.ReadFormula(arg.Conclusion.Id)
		if !ok {
			p := Problem{ARGUMENT, id, "conclusion is not a term.", arg.Conclusion.Id}
			problems = append(problems, p)
			continue
		}
		found := false
		for _, cs := range s.Conclusions {
			t4, ok := terms.ReadFormula(cs)
			if !ok {
				continue // conclusions of schemes are checked elsewhere
			}
			if _, ok := terms.MatchFormula(t4, t3, b); ok {
				found = true
				break
			}
		}
		if !found {
			p := Problem{ARGUMENT, id, "conclusion does not instantiate a conclusion of the scheme with the argument's parameters.", arg.Conclusion.Id}
			problems = append(problems, p)
		}
	}
	return problems
}

// instanceWithoutParameters reports whether an argument written without
// parameters is an instance of its scheme: one substitution makes its
// conclusion a conclusion of the scheme and its premises the scheme's
// premises and assumptions, each used once. accept checks that the exception
// links use the same substitution, without writing inferred parameters back.
func instanceWithoutParameters(s *caes.Scheme, arg *caes.Argument, accept func(terms.Bindings) bool) bool {
	if arg.Conclusion == nil || len(arg.Premises) != len(s.Premises)+len(s.Assumptions) {
		return false
	}
	c, ok := terms.ReadFormula(arg.Conclusion.Id)
	if !ok {
		return false
	}
	patterns := []terms.Term{}
	for _, ps := range append(append([]string{}, s.Premises...), s.Assumptions...) {
		if t, ok := terms.ReadFormula(ps); ok {
			patterns = append(patterns, t)
		}
	}
	stmts := []terms.Term{}
	for _, pr := range arg.Premises {
		if pr.Stmt == nil {
			return false
		}
		t, ok := terms.ReadFormula(pr.Stmt.Id)
		if !ok {
			return false
		}
		stmts = append(stmts, t)
	}
	// Bind from the premises first. A conclusion ¬X may normalize to p once
	// a premise binds X to ¬p; comparing that conclusion before the bindings
	// exist would reject the valid authored instance.
	return matchEachWith(patterns, stmts, nil, make([]bool, len(patterns)), func(b terms.Bindings) bool {
		for _, cs := range s.Conclusions {
			cp, ok := terms.ReadFormula(cs)
			if !ok {
				continue
			}
			if b2, ok := terms.MatchFormula(cp, c, b); ok && accept(b2) {
				return true
			}
		}
		return false
	})
}

// matchEach reports whether the statements can be paired one to one with the
// patterns so that every pair matches under one extension of the bindings b.
func matchEach(patterns, stmts []terms.Term, b terms.Bindings, used []bool) bool {
	return matchEachWith(patterns, stmts, b, used, func(terms.Bindings) bool { return true })
}

func matchEachWith(patterns, stmts []terms.Term, b terms.Bindings, used []bool, accept func(terms.Bindings) bool) bool {
	if len(stmts) == 0 {
		return accept(b)
	}
	// Ordinarily matching one statement first is order-independent. An
	// unbound negated variable is different: given(X) may bind X=¬p, making
	// the deferred pattern ¬X match p. Try another statement first in that
	// case, without inventing an inverse-negation binding from p alone.
	choices := 1
	for i, p := range patterns {
		if !used[i] && negationNeedsBinding(p, b) {
			choices = len(stmts)
			break
		}
	}
	for j := 0; j < choices; j++ {
		for i, p := range patterns {
			if used[i] {
				continue
			}
			if b2, ok := terms.MatchFormula(p, stmts[j], b); ok {
				remaining := stmts[1:]
				if j != 0 {
					remaining = append(append([]terms.Term(nil), stmts[:j]...), stmts[j+1:]...)
				}
				used[i] = true
				matched := matchEachWith(patterns, remaining, b2, used, accept)
				used[i] = false
				if matched {
					return true
				}
			}
		}
	}
	return false
}

func negationNeedsBinding(t terms.Term, b terms.Bindings) bool {
	switch x := t.(type) {
	case terms.Compound:
		if x.Functor == "¬" && len(x.Args) == 1 {
			if v, ok := x.Args[0].(terms.Variable); ok {
				if _, bound := terms.GetBinding(v, b); !bound {
					return true
				}
			}
		}
		for _, a := range x.Args {
			if negationNeedsBinding(a, b) {
				return true
			}
		}
	case terms.List:
		for _, a := range x {
			if negationNeedsBinding(a, b) {
				return true
			}
		}
	}
	return false
}

func exceptionProblem(id string, s *caes.Scheme) Problem {
	return Problem{ARGUMENT, id,
		"does not carry the scheme's exceptions: each instantiated exception needs a direct, single-premise linked argument for this argument's undercutter; let inference construct the argument, or supply all exception links",
		strings.Join(s.Exceptions, ", ")}
}

// carriesExceptions recognizes the exception structure built by
// InstantiateScheme and preserved by YAML export/import. An inert undercutter,
// a link for another instance, a missing exception, an extra prerequisite or a
// custom weighing function must not silently suppress a scheme's exceptions.
// An exception argument can itself be attacked, as generated arguments can;
// this validates its structure, not the outcome of a particular dispute.
func carriesExceptions(ag *caes.ArgGraph, arg *caes.Argument, b terms.Bindings) bool {
	if len(arg.Scheme.Exceptions) == 0 {
		return true
	}
	u := arg.Undercutter
	if u == nil || ag.Statements[terms.Normalize(u.Id)] != u {
		return false
	}
	for _, e := range arg.Scheme.Exceptions {
		pattern, ok := terms.ReadFormula(e)
		if !ok {
			return false
		}
		// Validation also sees malformed, nonground authored input. Do not
		// follow variable-to-variable bindings (possibly cyclic) while
		// checking an exception: valid argument bindings are ground terms.
		for _, v := range pattern.OccurVars() {
			value, bound := terms.GetBinding(v, b)
			if !bound || !terms.Ground(value, nil) {
				return false
			}
		}
		instance := terms.NormalizeNegation(terms.Substitute(pattern, b))
		if !terms.Ground(instance, nil) {
			return false
		}
		found := false
		for _, link := range u.Args {
			if link == nil || ag.Arguments[link.Id] != link || link.Conclusion != u || len(link.Premises) != 1 {
				continue
			}
			// A nil scheme has the linked weighing function. After YAML
			// import that same default is the built-in linked scheme.
			if link.Scheme != nil && link.Scheme != caes.BasicSchemes["linked"] {
				continue
			}
			p := link.Premises[0].Stmt
			if p != nil && ag.Statements[terms.Normalize(p.Id)] == p && terms.Normalize(p.Id) == terms.Normalize(instance.String()) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Validate the assumptions of an argument graph
func validateAssumptions(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	for _, k := range ag.Assumptions {
		// Check that the key is a term
		t, ok := terms.ReadString(k)
		if !ok {
			p := Problem{ASSUMPTION, "", "not a term", k}
			problems = append(problems, p)
		} else {
			// Check that the term is a ground atomic formula
			var b terms.Bindings // empty environment
			if !terms.AtomicFormula(t) || !terms.Ground(t, b) {
				p := Problem{ASSUMPTION, "", "not a ground atomic formula", k}
				problems = append(problems, p)
			}
			if d := caes.ArithmeticProblem(k); d != "" {
				problems = append(problems, Problem{ASSUMPTION, "", "arithmetic the rule engine would compute; quote a constant such as a date", k})
			}
			if d := caes.ReservedProblem(k); d != "" {
				problems = append(problems, Problem{ASSUMPTION, "", "a reserved name (is, true, false) used as a predicate", k})
			}
		}
		// Check that there is a statement for the assumption
		//		_, ok = ag.Statements[k]
		//		if !ok {
		//			p := Problem{ASSUMPTION, id, "not declared to be a statement", k}
		//			problems = append(problems, p)
		//		}
	}
	return problems
}

// Validate the expected labeling of an argument graph
func validateExpectedLabeling(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	for k, _ := range ag.ExpectedLabeling {
		// Check that the key is a term
		t, ok := terms.ReadString(k)
		if !ok {
			p := Problem{EXPECTEDLABELING, "", "not a term", k}
			problems = append(problems, p)
		} else {
			// Check that the term is a ground atomic formula
			var b terms.Bindings // empty environment
			if !terms.AtomicFormula(t) || !terms.Ground(t, b) {
				p := Problem{EXPECTEDLABELING, "", "not a ground atomic formula", k}
				problems = append(problems, p)
			}
		}
		// Check that there is a statement for the term
		found := false
		for k2, _ := range ag.Statements {
			t2, ok := terms.ReadString(k2)
			if !ok {
				continue
			}
			if t.String() == t2.String() {
				found = true
				break
			}
		}

		if !found {
			p := Problem{EXPECTEDLABELING, "", "not declared to be a statement", k}
			problems = append(problems, p)
		}
	}
	return problems
}

func validateLanguage(l caes.Language) []Problem {
	isPred := func(s string) bool {
		t, ok := terms.ReadString(s)
		return ok && t.Type() == terms.AtomType
	}
	// count the number of verbs in a format string
	verbCount := func(s string) int {
		n := 0
		for _, c := range s {
			if c == '%' {
				n++
			}
		}
		return n
	}
	problems := []Problem{}
	for k, v := range l {
		// check that the key has the form predicate/arity
		l := strings.Split(k, "/")
		if len(l) != 2 {
			p := Problem{LANGUAGE, "", "does not have the form predicate/arity", k}
			problems = append(problems, p)
		} else if !isPred(l[0]) {
			p := Problem{LANGUAGE, k, "not a predicate symbol", l[0]}
			problems = append(problems, p)
		} else if caes.ReservedPredicates[l[0]] {
			// docs/contracts.md, contract 7: in a rule body GoCHR executes
			// is/2 as an assignment
			p := Problem{LANGUAGE, k, "a reserved name, not a predicate symbol", l[0]}
			problems = append(problems, p)
		} else {
			var n int
			_, err := fmt.Sscanf(l[1], "%d", &n)
			if err != nil {
				p := Problem{LANGUAGE, k, "non-integer arity", l[1]}
				problems = append(problems, p)
			} else if verbCount(v) != n {
				p := Problem{LANGUAGE, k, "format string has incorrect number of placeholders (verbs)", l[1]}
				problems = append(problems, p)
			}
		}
	}
	return problems
}

// Validate that each string in a list represents a logical variable
func validateVariables(s *caes.Scheme) []Problem {
	l := s.Variables
	problems := []Problem{}
	for _, v := range l {
		t, ok := terms.ReadString(v)
		if !ok || t.Type() != terms.VariableType {
			p := Problem{SCHEME, s.Id, "not a variable", v}
			problems = append(problems, p)
		}
	}
	return problems
}

// Validate an argumentation scheme s against a lanuage l
func validateScheme(s *caes.Scheme, l caes.Language) []Problem {
	occVars := make(map[string]bool)

	addToMap := func(vars terms.Vars) {
		for _, v := range vars {
			occVars[v.Name] = true
		}
	}

	// Checks if s is declared as a variable in the scheme.
	declaredVariable := func(s2 string) bool {
		for _, v := range s.Variables {
			if s2 == v {
				return true
			}
		}
		return false
	}

	problems := validateVariables(s)

	validateAtom := func(atm string, kind string) {
		t, ok := terms.ReadString(atm)
		if !ok {
			p := Problem{SCHEME, s.Id, "not a term", atm}
			problems = append(problems, p)
		} else {
			var key string
			var varOrBool bool = false
			switch t.Type() {
			case terms.BoolType:
				varOrBool = true
			case terms.AtomType:
				key = t.String() + "/" + "0"
			case terms.CompoundType:
				key = t.(terms.Compound).Functor + "/" + strconv.Itoa(len(t.(terms.Compound).Args))
			case terms.VariableType:
				varOrBool = true
				if !(kind == "premise" || kind == "conclusion") {
					p := Problem{SCHEME, s.Id, fmt.Sprintf("%s may not be a variable", kind), atm}
					problems = append(problems, p)
				}
			default:
				p := Problem{SCHEME, s.Id, "not an atomic formula", atm}
				problems = append(problems, p)
			}
			if kind != "guard" {
				if d := caes.ArithmeticProblem(atm); d != "" {
					problems = append(problems, Problem{SCHEME, s.Id, "arithmetic the rule engine would compute; quote a constant such as a date", atm})
				}
				if d := caes.ReservedProblem(atm); d != "" {
					problems = append(problems, Problem{SCHEME, s.Id, "a reserved name (is, true, false) used as a predicate", atm})
				}
			}
			// Check that the predicate of the atom, with the given arity, has been declared in the language

			if !builtins[key] && key != "¬/1" && !varOrBool {
				// builtin operators, variables and booleans need not be declared
				_, ok := l[key]
				if !ok {
					p := Problem{SCHEME, s.Id, "predicate not declared in the language", key}
					problems = append(problems, p)
				}
			}
			// Check that all variables in the atom have been declared in the scheme
			vars := t.OccurVars()
			if kind == "premise" || kind == "deletion" {
				addToMap(vars)
			}

			for _, v := range vars {
				if !declaredVariable(v.Name) {
					p := Problem{SCHEME, s.Id, "variable not declared in the scheme", v.Name}
					problems = append(problems, p)
				} else if !(kind == "premise" || kind == "deletion") && !occVars[v.Name] {

					p := Problem{SCHEME, s.Id, "variable not used in premises or deletions ", v.Name}
					problems = append(problems, p)
				}
			}
		}
	}

	for _, atm := range s.Premises {

		validateAtom(atm, "premise")
	}
	for _, atm := range s.Deletions {
		validateAtom(atm, "deletion")
	}
	for _, atm := range s.Assumptions {
		validateAtom(atm, "assumption")
	}
	for _, atm := range s.Exceptions {
		validateAtom(atm, "exception")
	}

	for _, atm := range s.Guards {
		validateAtom(atm, "guard")
	}
	for _, atm := range s.Conclusions {
		validateAtom(atm, "conclusion")
	}

	//check if any variables are declared but not used
	for _, v := range s.Variables {
		if !occVars[v] {
			p := Problem{SCHEME, s.Id, "variable declared but not used ", v}
			problems = append(problems, p)
		}
	}

	return problems
}

// Validate the argumentation schemes of a theory
func validateSchemes(theory *caes.Theory) []Problem {
	problems := []Problem{}
	ids := map[string]bool{}
	for _, s := range theory.ArgSchemes {
		if ids[s.Id] {
			p := Problem{SCHEME, s.Id, "duplicate scheme id", ""}
			problems = append(problems, p)
		} else {
			ids[s.Id] = true
		}
		problems = append(problems, validateScheme(s, theory.Language)...)
	}
	return problems
}

func validateIssueSchemes(theory *caes.Theory) []Problem {
	problems := []Problem{}
	l := theory.Language

	validatePattern := func(sid string, t terms.Term) {
		// allow patterns to be variables
		if t.Type() == terms.VariableType {
			return
		}
		var key string
		switch t.Type() {
		case terms.AtomType:
			key = t.String() + "/" + "0"
		case terms.CompoundType:
			key = t.(terms.Compound).Functor + "/" + strconv.Itoa(len(t.(terms.Compound).Args))
		default:
			p := Problem{ISCHEME, sid, "pattern is not an atomic formula", t.String()}
			problems = append(problems, p)
		}

		// Check that the predicate of the atom, with the given arity, has been declared in the language
		if key == "¬/1" {
			// Negation is built-in and doesn't need to be declared
			return
		} else {
			_, ok := l[key]
			if !ok {

				p := Problem{ISCHEME, sid, "predicate of pattern not declared in the language", key}
				problems = append(problems, p)
			}
		}
	}

	for sid, is := range theory.IssueSchemes {
		s := *is
		if len(s) < 2 {
			p := Problem{ISCHEME, sid, "fewer than two patterns", ""}
			problems = append(problems, p)
		}
		// Check that each string in the list of the scheme represents an atom, or
		// has three elements, where the first an last are atoms and the
		// second is "...".  If it is an atom, also check that its predicate
		// is defined in the language.
		if len(s) == 3 && s[1] == "..." {
			for _, i := range []int{0, 2} {
				t, ok := terms.ReadString(s[i])
				if !ok {
					p := Problem{ISCHEME, sid, "pattern is not a term", s[i]}
					problems = append(problems, p)
				} else {
					validatePattern(sid, t)
				}
			}
		} else {
			for i, _ := range s {
				t, ok := terms.ReadString(s[i])
				if !ok {
					p := Problem{ISCHEME, sid, "pattern is not a term", s[i]}
					problems = append(problems, p)
				} else {
					validatePattern(sid, t)
				}
			}
		}
	}
	return problems
}

// Validate a theory of an argument graph
func validateTheory(ag *caes.ArgGraph) []Problem {
	problems := validateLanguage(ag.Theory.Language)
	schemeProblems := validateSchemes(ag.Theory)
	if len(schemeProblems) > 0 {
		problems = append(problems, schemeProblems...)
	}
	issueSchemeProblems := validateIssueSchemes(ag.Theory)
	if len(issueSchemeProblems) > 0 {
		problems = append(problems, issueSchemeProblems...)
	}
	return problems
}

// Validate an argument graph
func Validate(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	problems = append(problems, validateStatements(ag)...)
	problems = append(problems, validateIssues(ag)...)
	problems = append(problems, validateArguments(ag)...)
	problems = append(problems, validateAssumptions(ag)...)
	problems = append(problems, validateExpectedLabeling(ag)...)
	problems = append(problems, validateTheory(ag)...)

	return problems
}
