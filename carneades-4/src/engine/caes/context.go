package caes

import (
	"context"
	"errors"
	"github.com/carneades/carneades-4/src/engine/terms"
	"sort"
)

// Inconsistent reports contradictory accepted context, including pending
// issue schemes. CheckContext additionally reports malformed input and limits.
// Neither method changes the graph.
func (ag *ArgGraph) Inconsistent() bool {
	return errors.Is(ag.CheckContext(), ErrInconsistentContext)
}

// MAXISSUESTEPS bounds candidate/pattern visits while materializing issues.
// It is a finite safety limit on work. Ground position lookup avoids the
// former quadratic scan; open patterns and enumerations still scan. Graphs
// exceeding this quota are refused, even if an unbounded evaluation would
// eventually finish.
const MAXISSUESTEPS = 50000000

type issueWork struct {
	ctx       context.Context
	remaining int
}

func (w *issueWork) step() error {
	if w == nil {
		return nil
	}
	if w.ctx.Err() != nil {
		return NewRunError(ErrCancelled, "cancelled during issue construction")
	}
	if w.remaining <= 0 {
		return NewRunError(ErrLimitReached, "issue matching exhausted its candidate-visit budget")
	}
	w.remaining--
	return nil
}

// CheckContext validates the raw assumptions and the complete issue closure,
// including unaccepted statement identities that can connect issue patterns.
// Default origins and questioning are preserved in the private graph.
func (ag *ArgGraph) CheckContext() error { return ag.checkContext(context.Background()) }
func (ag *ArgGraph) checkContext(ctx context.Context) error {
	probe := ag.copyForInference()
	return probe.resolveContext(ctx)
}

// resolveContext is for a private graph: apply issue closure before obtaining
// effective assumptions, so defeated defaults never seed argument generation.
func (ag *ArgGraph) resolveContext(ctx context.Context) error {
	if err := ag.normalizeGraph(); err != nil {
		return err
	}
	for _, a := range ag.Assumptions {
		if ctx.Err() != nil {
			return NewRunError(ErrCancelled, "cancelled during context validation")
		}
		if t, ok := terms.ReadFormula(a); !ok || t == nil {
			return NewRunError(ErrInvalidInput, "cannot parse assumption %q", a)
		}
	}
	if err := ag.materializeIssues(ctx); err != nil {
		return err
	}
	if ag.inconsistentPositions() {
		return NewRunError(ErrInconsistentContext, "the accepted context contains two positions of one issue, or a statement and its negation")
	}
	return nil
}

func (ag *ArgGraph) materializeIssues(ctx context.Context) error {
	// Include default and questioned identities too. Their status is calculated
	// by Assumed after the complete issue structure exists.
	for _, a := range ag.Assumptions {
		n := terms.Normalize(a)
		if _, ok := ag.Statements[n]; !ok {
			text := n
			if t, ok := terms.ReadFormula(n); ok && ag.Theory != nil {
				text = ag.Theory.Language.Apply(t)
			}
			ag.Statements[n] = &Statement{Id: n, Text: text}
		}
	}
	if ag.Theory == nil {
		return nil
	}
	names := []string{}
	for n := range ag.Theory.IssueSchemes {
		names = append(names, n)
	}
	sort.Strings(names)
	work := &issueWork{ctx: ctx, remaining: MAXISSUESTEPS}
	for _, name := range names {
		if err := work.step(); err != nil {
			return err
		}
		patterns := ag.Theory.IssueSchemes[name]
		if patterns == nil {
			return NewRunError(ErrInvalidInput, "issue scheme %s has no patterns", name)
		}
		if err := ag.makeIssueWithWork(name, *patterns, work); err != nil {
			if errors.Is(err, ErrLimitReached) || errors.Is(err, ErrCancelled) {
				return err
			}
			return NewRunError(ErrInvalidInput, "issue scheme %s: %v", name, err)
		}
	}
	return nil
}
