package carneades

import (
	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// Checkpoint is a restorable record of the argument graph's structure: which
// statements, arguments and issues exist, every statement's argument list,
// issue, undercutter flag and text, every argument's undercutter, every
// issue's positions, and the assumption list.
//
// Review defect #5: a move rejected as cyclic was undone by RemoveArguments
// only, which left behind the statements, issues, assumptions and undercutter
// pointers the move had created. Restore puts the graph back exactly.
type Checkpoint struct {
	stmts       map[string]stmtState
	args        map[string]*caes.Statement // argument ID -> its undercutter at checkpoint time
	issues      map[string][]*caes.Statement
	assumptions []string
}

type stmtState struct {
	stmt          *caes.Statement
	args          []*caes.Argument
	issue         *caes.Issue
	isUndercutter bool
	text          string
	label         caes.Label
}

// Checkpoint records the graph's current structure.
func (a *Adapter) Checkpoint() *Checkpoint {
	a.mu.Lock()
	defer a.mu.Unlock()
	cp := &Checkpoint{
		stmts:       make(map[string]stmtState, len(a.Graph.Statements)),
		args:        make(map[string]*caes.Statement, len(a.Graph.Arguments)),
		issues:      make(map[string][]*caes.Statement, len(a.Graph.Issues)),
		assumptions: append([]string{}, a.Graph.Assumptions...),
	}
	for id, s := range a.Graph.Statements {
		cp.stmts[id] = stmtState{stmt: s, args: append([]*caes.Argument{}, s.Args...), issue: s.Issue,
			isUndercutter: s.IsUndercutter, text: s.Text, label: s.Label}
	}
	for id, arg := range a.Graph.Arguments {
		cp.args[id] = arg.Undercutter
	}
	for id, is := range a.Graph.Issues {
		cp.issues[id] = append([]*caes.Statement{}, is.Positions...)
	}
	return cp
}

// Restore puts the graph back to the checkpoint: everything created since is
// removed and everything changed is reset. The statement, argument and issue
// objects that existed at the checkpoint are kept (pointers held elsewhere
// stay valid).
func (a *Adapter) Restore(cp *Checkpoint) {
	if cp == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for id := range a.Graph.Arguments {
		if _, ok := cp.args[id]; !ok {
			delete(a.Graph.Arguments, id)
		}
	}
	a.argSeq = 0 // rescan: argument IDs are free again
	for id, und := range cp.args {
		if arg := a.Graph.Arguments[id]; arg != nil {
			arg.Undercutter = und
		}
	}
	for id := range a.Graph.Statements {
		if _, ok := cp.stmts[id]; !ok {
			delete(a.Graph.Statements, id)
		}
	}
	for id, st := range cp.stmts {
		a.Graph.Statements[id] = st.stmt
		st.stmt.Args = append([]*caes.Argument{}, st.args...)
		st.stmt.Issue = st.issue
		st.stmt.IsUndercutter = st.isUndercutter
		st.stmt.Text = st.text
		st.stmt.Label = st.label
	}
	for id, is := range a.Graph.Issues {
		pos, ok := cp.issues[id]
		if !ok {
			delete(a.Graph.Issues, id)
			continue
		}
		is.Positions = append([]*caes.Statement{}, pos...)
	}
	a.Graph.Assumptions = append([]string{}, cp.assumptions...)
}
