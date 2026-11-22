// Copyright © 2026 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL was not distributed
// with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

package caes

import (
	"fmt"
	"strings"

	"github.com/carneades/carneades-4/src/engine/terms"
)

// arithmeticOperators are the operators the rule engine computes when their
// operands are numbers.
var arithmeticOperators = map[string]bool{
	"+": true, "-": true, "*": true, "/": true, "//": true, "**": true, "^": true,
	"mod": true, "div": true, "rem": true,
}

// Arithmetic returns the first subterm of a statement that the rule engine
// would compute: an arithmetic operator applied to a number or a variable,
// such as 1990-05-17, which it computes as 1968, or X+1. A statement carries
// no arithmetic (docs/contracts.md, contract 1); such a statement is invalid
// input. An operator between names, as in the scheme id
// opposition-normative, is not computed, and a negative number is a
// constant.
func Arithmetic(t terms.Term) (terms.Term, bool) {
	switch x := t.(type) {
	case terms.Compound:
		if x.Prio != 0 && arithmeticOperators[x.Functor] {
			if len(x.Args) == 1 && isNumber(x.Args[0]) {
				return nil, false // a signed number
			}
			if computable(x) {
				return x, true
			}
		}
		for _, a := range x.Args {
			if s, ok := Arithmetic(a); ok {
				return s, true
			}
		}
	case terms.List:
		for _, a := range x {
			if s, ok := Arithmetic(a); ok {
				return s, true
			}
		}
	}
	return nil, false
}

// computable reports whether an arithmetic expression has a number or a
// variable among the operands of its arithmetic operators.
func computable(t terms.Term) bool {
	switch x := t.(type) {
	case terms.Int, terms.Float, terms.Variable:
		return true
	case terms.Compound:
		if x.Prio != 0 && arithmeticOperators[x.Functor] {
			for _, a := range x.Args {
				if computable(a) {
					return true
				}
			}
		}
	}
	return false
}

func isNumber(t terms.Term) bool {
	switch t.(type) {
	case terms.Int, terms.Float:
		return true
	}
	return false
}

// ReservedPredicates are the names a statement may not use as its predicate
// (docs/contracts.md, contract 7): in a rule body GoCHR executes is/2 as an
// assignment, and true and false are booleans.
var ReservedPredicates = map[string]bool{"is": true, "true": true, "false": true}

// Unsupported reports the schemes of a theory that are outside the supported
// profile (docs/contracts.md, contract 1): a scheme with deletions or guards
// is a CHR rule, not a propagation rule, and is not run.
func Unsupported(t *Theory) error {
	if t == nil {
		return nil
	}
	var ids []string
	for _, s := range t.ArgSchemes {
		if len(s.Deletions) != 0 || len(s.Guards) != 0 {
			ids = append(ids, s.Id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return NewRunError(ErrUnsupported, "schemes with deletions or guards are outside the supported profile: %s", strings.Join(ids, ", "))
}

// arithmeticProblems lists the statements of a scheme with arithmetic the
// rule engine would compute.
func arithmeticProblems(s *Scheme) []string {
	var out []string
	for _, l := range [][]string{s.Premises, s.Conclusions, s.Assumptions, s.Exceptions} {
		for _, st := range l {
			if p := ArithmeticProblem(st); p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

// ArithmeticProblem says why a statement with arithmetic the rule engine would
// compute is invalid input, or returns "".
func ArithmeticProblem(st string) string {
	t, ok := terms.ReadString(st)
	if !ok {
		return "" // the reader's own error is reported elsewhere
	}
	if a, ok := Arithmetic(t); ok {
		return fmt.Sprintf("%s: arithmetic the rule engine would compute (%s); quote a constant such as a date", st, a)
	}
	return ""
}

// ReservedProblem says why a statement whose predicate is a reserved name is
// invalid (contract 7), or returns "". The validator reports it.
func ReservedProblem(st string) string {
	t, ok := terms.ReadString(st)
	if !ok {
		return ""
	}
	if f, ok := terms.Functor(t); ok && ReservedPredicates[f] {
		return fmt.Sprintf("%s: %s is a reserved name and cannot be a predicate", st, f)
	}
	return ""
}
