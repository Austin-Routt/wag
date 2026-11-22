// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Wrapper for the GoCHR inference engine

package caes

import (
	"context"
	"errors"
	"fmt"
	// "log"
	"strings"

	chr "github.com/hfried/GoCHR/src/engine/CHR"
)

// Translate a theory into a GoCHR rulestore. A scheme whose rule GoCHR
// refuses is left out; RuleStoreOf reports it.
func TheoryToRuleStore(t *Theory) *chr.RuleStore {
	rs, _ := RuleStoreOf(t)
	return rs
}

// RuleStoreOf translates a theory into a GoCHR rule store, one rule per
// scheme with conclusions. The error names every scheme whose rule GoCHR
// refuses, for example for a premise it cannot read (E-PARSE), or that has a
// statement with arithmetic the rule engine would compute (contract 1); those
// schemes are left out of the store.
//
// The rule engine keeps package-level state, so RuleStoreOf waits while an
// inference run or another translation is in progress (docs/contracts.md,
// contract 1). Using the returned store is outside that protection.
func RuleStoreOf(t *Theory) (*chr.RuleStore, error) {
	if err := acquireEngine(context.Background()); err != nil {
		return nil, err
	}
	defer releaseEngine()
	return ruleStoreOf(t)
}

// engine serializes the uses of the rule engine: GoCHR's parser and
// interpreter keep package-level state (a variable counter, trace settings),
// which two runs at the same time would share (docs/contracts.md, contract 1;
// the third review's TestConcurrentImports and race job).
var engine = make(chan struct{}, 1)

// acquireEngine waits until no other run uses the rule engine, or ctx is done.
func acquireEngine(ctx context.Context) error {
	if ctx.Err() != nil {
		return NewRunError(ErrCancelled, "cancelled before using the rule engine")
	}
	select {
	case engine <- struct{}{}:
		if ctx.Err() != nil {
			releaseEngine()
			return NewRunError(ErrCancelled, "cancelled before using the rule engine")
		}
		return nil
	case <-ctx.Done():
		return NewRunError(ErrCancelled, "cancelled while another run used the rule engine")
	}
}

func releaseEngine() { <-engine }

// ruleStoreOf is RuleStoreOf for a caller that holds the engine.
func ruleStoreOf(t *Theory) (*chr.RuleStore, error) {
	// log.Printf("TheoryToRuleStore\n") // DEBUG
	rs := chr.MakeRuleStore()
	var errs []string
	// Argument construction is forward chaining over facts, and facts form a
	// set (docs/contracts.md, contract 2): a fact derived a second time must
	// not trigger the schemes again. Without this, schemes that re-derive a
	// fact, or derive each other's premises, run to MAXRULEAPPS.
	rs.Unique = true
	rs.ReuseConstraints = true
	rs.MaxTermSize = MAXTERMSIZE
	rs.MaxMatchSteps = MAXMATCHSTEPS
	rs.MaxConstraints = MAXCONSTRAINTS
	rs.MaxStoreNodes = MAXSTORENODES
	// ¬S is the explicit negation of S (docs/contracts.md, contract 4): it is
	// matched by rule heads like any statement, also when it reaches a rule
	// through a statement variable, and ¬¬S is read as S.
	rs.ExplicitNegation = true
	for _, s := range t.ArgSchemes {
		// If the scheme has no conclusions, skip the scheme
		// and assume it only defines a weighing function but no rule
		if len(s.Conclusions) > 0 {
			// A "go" term is added to CHR rules for
			// argument schemes with no premises, since CHR requires
			// rules to have at least one term in the head.
			var premises []string
			if len(s.Premises) > 0 {
				premises = s.Premises
			} else if len(s.Deletions) == 0 {
				premises = []string{"go"}
			}
			argTerm := fmt.Sprintf("argument(%s,[%s])", s.Id, strings.Join(s.Variables, ","))
			// A new slice: appending to s.Conclusions could write into the
			// scheme's own backing array.
			conclusions := make([]string, 0, len(s.Conclusions)+len(s.Assumptions)+1)
			conclusions = append(conclusions, s.Conclusions...)
			conclusions = append(conclusions, s.Assumptions...)
			conclusions = append(conclusions, argTerm)
			// Note that the body of the rules includes the assumptions
			// and conclusions of the scheme.
			// fmt.Printf("AddRule id=%s, premises=%s, deletions=%s, guards=%s, conclusions=%s\n", s.Id, premises, s.Deletions, s.Guards, conclusions)
			if probs := arithmeticProblems(s); len(probs) != 0 {
				errs = append(errs, fmt.Sprintf("scheme %s: %s", s.Id, strings.Join(probs, "; ")))
				continue
			}
			if err := rs.AddRule(s.Id, premises, s.Deletions, s.Guards, conclusions); err != nil {
				errs = append(errs, fmt.Sprintf("scheme %s: %v", s.Id, err))
			}
		}
	}
	if len(errs) != 0 {
		return rs, errors.New(strings.Join(errs, "; "))
	}
	return rs, nil
}
