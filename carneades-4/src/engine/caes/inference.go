// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Inference of arguments (aka argument construction or generation)
// using Constraint Handling Rules (CHR)

package caes

import (
	"context"
	"errors"
	"strings"

	"github.com/carneades/carneades-4/src/engine/terms"
	chr "github.com/hfried/GoCHR/src/engine/CHR"
)

// maximum number of rule (scheme) applications when deriving arguments
const MAXRULEAPPS = 100000

// MAXTERMSIZE is the largest size, in nodes, of a statement or argument term
// that argument construction may derive (docs/contracts.md, contract 3). A
// scheme whose terms grow on every application reaches it long before
// MAXRULEAPPS, which alone would let such a run exhaust the memory: GoCHR's
// cost per application grows with the size of the terms, so a term growing
// by one node per application needs 0.25 s and 170 MB to reach 500 nodes and
// 1.3 s and 1.2 GB to reach 1,000. The suite of tools/vv gives the same
// result on every check with a limit of 30 as with 500.
const MAXTERMSIZE = 500

// Logical work and storage quotas for the supported propagation profile.
// These also bound unsuccessful joins and broad, shallow stores; they do
// not replace input-size limits or operating-system memory isolation.
// MAXMATCHSTEPS admits the reference four-premise join over 120 facts, which
// exceeds one million visits. Runtime depends on the machine and theory.
// Larger finite searches can still be refused at this explicit work quota.
const MAXMATCHSTEPS = 50000000
const MAXCONSTRAINTS = 100000
const MAXSTORENODES = 2000000

// ArgDesc: Structure describing an argument instantiating an
// argument scheme. Represented in Prolog as argument(Scheme,Values)
type ArgDesc struct {
	Scheme string   // id of the scheme
	Values []string // values instantiating the variables of the scheme
}

// Returns true with an ArgDesc in the second result if the
// string s matches argument(S,P)
func termToArgDesc(s string) (bool, *ArgDesc) {
	if s == "" {
		return false, nil
	}
	t1, ok := terms.ReadFormula(s)
	if !ok {
		return false, nil
	}
	S := terms.NewVariable("S")
	P := terms.NewVariable("P")
	t2 := terms.NewCompound("argument", []terms.Term{S, P})
	var bindings terms.Bindings
	bindings, ok = terms.Match(t2, t1, bindings)
	if !ok {
		return false, nil
	}
	scheme, _ := terms.GetBinding(S, bindings)
	parms, _ := terms.GetBinding(P, bindings)

	// convert the parms from a list of terms to a list of strings
	l := []string{}
	if parms.Type() == terms.ListType {
		for _, t := range parms.(terms.List) {
			l = append(l, t.String())
		}
	}

	return true, &ArgDesc{scheme.String(), l}
}

// Infer: Translate a theory into CHR rules and use
// a CHR engine to construct arguments and add them to the argument graph.
// Does not compute or update labels.  If the theory is syntactically incorrect
// and thus cannot be parsed by the CHR inference engine, an error is returned
// and argument graph is left unchanged. If all goes well, the argument
// graph is updated and nil is returned.
//
// A run that reaches MAXRULEAPPS or MAXTERMSIZE while a scheme could still
// apply returns an error wrapping ErrLimitReached, and leaves the graph
// unchanged (docs/contracts.md, contract 3).
func (ag *ArgGraph) Infer() error {
	return ag.InferContext(context.Background())
}

// InferContext is Infer, stopped with an error wrapping ErrCancelled when ctx
// is done before argument construction ends. The graph is then left
// unchanged.
//
// Runs from several goroutines are serialized: a run waits until no other run
// uses the rule engine (docs/contracts.md, contract 1), and a run whose ctx
// is done while it waits ends as cancelled.
func (ag *ArgGraph) InferContext(ctx context.Context) error {
	if err := acquireEngine(ctx); err != nil {
		return err
	}
	defer releaseEngine()
	work := ag.copyForInference()
	if err := work.inferStaged(ctx); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return NewRunError(ErrCancelled, "cancelled before committing the graph")
	}
	ag.commitInference(work)
	return nil
}

// inferStaged may mutate only a private graph; any failure discards it.
func (ag *ArgGraph) inferStaged(ctx context.Context) error {
	if ag.Theory == nil {
		return NewRunError(ErrInvalidInput, "argument graph has no theory")
	}
	if err := Unsupported(ag.Theory); err != nil {
		return err
	}
	if err := ag.resolveContext(ctx); err != nil {
		return err
	}
	for _, a := range ag.Assumptions {
		if p := ArithmeticProblem(a); p != "" {
			return NewRunError(ErrInvalidInput, "assumption %s", p)
		}
	}
	if len(ag.Theory.ArgSchemes) != 0 {
		// rb := TheoryToSWIRulebase(ag.Theory)
		rb, err := ruleStoreOf(ag.Theory)
		if err != nil {
			return NewRunError(ErrInvalidInput, "the rule engine refused the theory: %v", err)
		}
		rb.Done = ctx.Done()

		// Create an index of the previous arguments constructed
		// to avoid constructing equivalent instanstiations of schemes
		// and to allow the inference engine to construct undercutters
		//
		// An argument is identified by its scheme and its canonical parameters
		// (docs/contracts.md, contract 2): the key is the term in the
		// canonical form of the term reader, so "[a,b]" written here and
		// "[a, b]" printed by the rule engine are one key. Arguments without a
		// scheme, such as the undercutters built for exceptions, are not
		// instances of a scheme and have no key.
		prevArgs := map[string]bool{}
		goalArgs := []string{}
		for _, a := range ag.Arguments {
			if a != nil && a.Scheme != nil {
				k := "argument(" + a.Scheme.Id + ",[" + strings.Join(a.Parameters, ",") + "])"
				if !prevArgs[terms.Normalize(k)] {
					goalArgs = append(goalArgs, k)
				}
				prevArgs[terms.Normalize(k)] = true
			}
		}

		// The goals are initialized with a "go" goal, for use by schemes
		// with no premises, which are translated into CHR rules with a dummy "go"
		// term in their heads.
		goals := []string{"go"}

		// The actual goals in the query with the CHR inference
		// engine consist of the union of the assumptions of the argument graph
		// and the assumptions for each of the previous arguments
		// a questioned default is no fact to reason from (contract 5)
		goals = append(goals, ag.Assumed()...)
		goals = append(goals, goalArgs...)

		success, store, err := rb.Infer(goals, MAXRULEAPPS)
		switch {
		case errors.Is(err, chr.ErrLimit):
			return NewRunError(ErrLimitReached, "argument construction stopped at %s",
				strings.TrimPrefix(err.Error(), chr.ErrLimit.Error()+": "))
		case errors.Is(err, chr.ErrCancelled):
			return NewRunError(ErrCancelled, "argument construction was stopped by the caller")
		case err != nil:
			return NewRunError(ErrInvalidInput, "the rule engine rejected the theory: %v", err)
		}
		if !success {
			return nil
		}
		if ctx.Err() != nil {
			return NewRunError(ErrCancelled, "cancelled before committing constructed arguments")
		}

		// fmt.Printf("store:\n")
		for _, s := range store {
			if ctx.Err() != nil {
				return NewRunError(ErrCancelled, "cancelled while constructing arguments")
			}
			// fmt.Printf("   %s\n", s)
			// If the term does not represent an argument already in the graph
			// then, if the term does represent an argument, use it to
			// add an argument to the argument graph by instantiating the
			// argumentation scheme applied.
			if k := terms.Normalize(s); !prevArgs[k] {
				isArg, a := termToArgDesc(s)
				if isArg {
					ag.InstantiateScheme(a.Scheme, a.Values)
					prevArgs[k] = true
				}
			}
		}
	}
	if len(ag.Theory.ArgSchemes) == 0 && ctx.Err() != nil {
		return NewRunError(ErrCancelled, "cancelled before committing context statements")
	}
	// Once the run is complete: an assumption is a statement of the graph even
	// when no argument uses it, so that it is labelled (in) like any other.
	// (On any other outcome the graph stays as it was.)
	for _, a := range ag.Assumed() {
		if _, ok := ag.Statements[a]; !ok {
			text := a
			if t, ok := terms.ReadFormula(a); ok && ag.Theory != nil {
				text = ag.Theory.Language.Apply(t)
			}
			ag.Statements[a] = &Statement{Id: a, Text: text}
		}
	}
	if err := ag.materializeIssues(ctx); err != nil {
		return err
	}
	if ag.inconsistentPositions() {
		return NewRunError(ErrInconsistentContext, "constructed issues connect contradictory accepted context")
	}
	return nil
}
