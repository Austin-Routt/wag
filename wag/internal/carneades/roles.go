package carneades

import caes "github.com/carneades/carneades-4/src/engine/caes"

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
