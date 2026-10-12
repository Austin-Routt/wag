package carneades

import (
	"fmt"
	"time"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
)

// Every argument the adapter creates records its role in Metadata[RoleKey].
const (
	RoleKey = "wag_role"

	// RoleMain is an argument a player put forward: a support move's main
	// argument (one per scheme conclusion) or an attack move's argument.
	RoleMain = "main"
	// RoleException is an exception argument created by InstantiateScheme:
	// it argues the main arguments' undercutter from one of the scheme's
	// exceptions. It is part of the scheme, not a move, and is never a target.
	RoleException = "exception"
	// RolePresumption is a premise-less argument that makes a move's premise
	// hold until it is attacked (see AddPresumption).
	RolePresumption = "presumption"
)

// ArgumentRole returns the role recorded for an argument: RoleMain,
// RoleException, RolePresumption, or "" for an argument the adapter did not
// create (for example one loaded from YAML).
func ArgumentRole(arg *caes.Argument) string {
	if arg == nil || arg.Metadata == nil {
		return ""
	}
	r, _ := arg.Metadata[RoleKey].(string)
	return r
}

// IsAttackable reports whether players may target the argument with a
// critical question: everything except exception and presumption arguments.
func IsAttackable(arg *caes.Argument) bool {
	switch ArgumentRole(arg) {
	case RoleException, RolePresumption:
		return false
	}
	return arg != nil && !(arg.Conclusion != nil && arg.Conclusion.IsUndercutter && ArgumentRole(arg) == "")
}

// PresumptionSchemeID is the scheme ID of presumption arguments. It maps to no
// NASCQ scheme, so presumption arguments are never offered as targets.
const PresumptionSchemeID = "wag_presumption"

// AddPresumption makes a statement hold until it is attacked, on behalf of a
// player: a premise-less argument for the statement, owned by the player, with
// a constant weight (review defect #1).
//
// Why not an assumption: Carneades labels an assumption In unconditionally
// and makes the other positions of its issue Out, so an assumed premise can
// never be undermined. A presumption argument can be outweighed: an
// undermining argument for the premise's negation with a greater weight wins
// the issue under PE, an equal weight leaves both positions Out, and the
// presumer can reinstate the premise by backing it with a stronger argument.
// The Kialo translator uses the same device (presumption weight 0.01).
//
// One presumption per (statement, owner): calling it again returns the
// existing argument's ID.
func (a *Adapter) AddPresumption(ownerID, stmtID string, weight float64) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	stmt := a.Graph.Statements[terms.Normalize(stmtID)]
	if stmt == nil {
		stmt = a.Graph.Statements[stmtID]
	}
	if stmt == nil {
		return "", fmt.Errorf("statement %q not found", stmtID)
	}
	for _, arg := range stmt.Args {
		if ArgumentRole(arg) == RolePresumption && arg.Metadata["owner"] == ownerID {
			return arg.Id, nil
		}
	}
	id := a.nextArgID()
	arg := &caes.Argument{
		Id:         id,
		Metadata:   caes.Metadata{"owner": ownerID, RoleKey: RolePresumption, "created_at": time.Now().UTC().Format(time.RFC3339)},
		Scheme:     &caes.Scheme{Id: PresumptionSchemeID, Weight: caes.ConstantWeighingFunction(weight)},
		Conclusion: stmt,
	}
	a.Graph.Arguments[id] = arg
	stmt.Args = append(stmt.Args, arg)
	return id, nil
}

// AssumptionCount returns the number of entries in the graph's assumption list.
func (a *Adapter) AssumptionCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.Graph.Assumptions)
}

// RetractAssumptionsSince removes every assumption added after the list had n
// entries and returns their IDs. InstantiateScheme assumes a scheme's declared
// assumptions (e.g. possible(A) of practical reasoning); the game retracts them
// and presumes them instead, so that critical questions can undermine them.
// Carneades rebuilds its internal assumption map from this list at the start of
// every labelling, so truncating the list is sufficient.
func (a *Adapter) RetractAssumptionsSince(n int) []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if n < 0 || n >= len(a.Graph.Assumptions) {
		return nil
	}
	out := append([]string{}, a.Graph.Assumptions[n:]...)
	a.Graph.Assumptions = a.Graph.Assumptions[:n]
	return out
}
