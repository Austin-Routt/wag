package carneades

// yaml.go — YAML-based graph construction, argument removal for rollback,
// and snapshot/restore functionality.
//
// NewAdapterFromYAML loads Carneades YAML files that include schemes, language
// maps, and optionally pre-defined arguments and assumptions.
//
// Snapshot/Restore design:
// ========================
// The original plan proposed YAML Export/Import for snapshots. After source
// code review, this approach has a critical flaw: the YAML Export function
// does NOT export the Theory section (language, schemes, weighing functions).
// See the exporter in carneades-4's encoding/yaml/yaml.go — the Theory export is
// entirely commented out. After import, all argument Scheme pointers are
// relinked to BasicSchemes or whatever the YAML file defines, losing our
// ConstantWeighingFunction overrides.
//
// Instead, we provide two rollback mechanisms:
//
// 1. RemoveArgument (for single-move undo): Removes a specific argument
//    created during a failed turn attempt. This is the primary rollback
//    mechanism during the turn execution pipeline.
//
// 2. Full snapshot via original-YAML + move replay (for full game state
//    rollback). The game layer stores the original YAML bytes and the
//    move history. To restore: reload YAML, replay moves. This is deferred
//    to the game layer (Block 5).
//
// Why not YAML Export/Import for snapshots:
//   - Theory (schemes, language, weighing functions) is NOT exported
//   - ConstantWeighingFunction closures can't be serialized
//   - Package-level globals in the YAML encoder make concurrent use unsafe
//   - Arguments get re-linked to wrong scheme objects on import
//
// Why not encoding/gob or encoding/json for deep copy:
//   - caes.ArgGraph has cyclic pointer references (Statement ↔ Argument,
//     Statement ↔ Issue) that cause infinite recursion in naive serializers
//   - WeighingFunction is a func type, not serializable

import (
	"bytes"
	"fmt"
	"io"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	caesYaml "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
)

// NewAdapterFromYAML loads a Carneades ArgGraph from YAML bytes.
//
// The YAML format supports:
//   - language: predicate/arity → format string mappings
//   - argument_schemes: scheme definitions with variables, premises,
//     conclusions, assumptions, exceptions, and weighing functions
//   - statements: pre-defined statement nodes
//   - arguments: pre-defined argument nodes
//   - issues: pre-defined issue groupings
//   - assumptions: statements assumed to be true
//   - weighing_functions: custom weighing function definitions
//
// After import, Theory.InitSchemeIndex() is called automatically (by the
// YAML Import function in encoding/yaml/yaml.go) to populate the unexported
// schemeIndex map used by InstantiateScheme.
//
// The adapter's local schemeMap cache is then built from Theory.ArgSchemes.
func NewAdapterFromYAML(yamlBytes []byte) (*Adapter, error) {
	ag, err := caesYaml.Import(bytes.NewReader(yamlBytes))
	if err != nil {
		return nil, fmt.Errorf("yaml import failed: %w", err)
	}

	// Defensive: ensure all maps are initialized.
	// NewArgGraph() (called by caesArgMapGraph2caes) already does this,
	// but belt-and-suspenders is cheap.
	ensureMapsInitialized(ag)

	a := &Adapter{Graph: ag}
	a.buildSchemeMap()
	return a, nil
}

// NewAdapterFromYAMLReader loads from an io.Reader instead of bytes.
// Convenience wrapper for file-based loading.
func NewAdapterFromYAMLReader(r io.Reader) (*Adapter, error) {
	ag, err := caesYaml.Import(r)
	if err != nil {
		return nil, fmt.Errorf("yaml import failed: %w", err)
	}
	ensureMapsInitialized(ag)
	a := &Adapter{Graph: ag}
	a.buildSchemeMap()
	return a, nil
}

// ensureMapsInitialized guards against nil map panics.
func ensureMapsInitialized(ag *caes.ArgGraph) {
	if ag.Statements == nil {
		ag.Statements = make(map[string]*caes.Statement)
	}
	if ag.Arguments == nil {
		ag.Arguments = make(map[string]*caes.Argument)
	}
	if ag.Issues == nil {
		ag.Issues = make(map[string]*caes.Issue)
	}
	if ag.Theory == nil {
		ag.Theory = caes.NewTheory()
	}
}

// ---------------------------------------------------------------------------
// Rollback: RemoveArgument
// ---------------------------------------------------------------------------

// RemoveArgument removes an argument from the graph and cleans up all
// bidirectional links. This is the primary rollback mechanism for undoing
// a failed move during the turn execution pipeline.
//
// What it removes:
//   - The argument from Graph.Arguments map
//   - The argument from its conclusion's Args slice (bidirectional link)
//   - If the argument was the only one concluding an undercutter statement,
//     the undercutter link on the target argument is NOT cleared (other
//     arguments may also conclude that undercutter).
//
// What it does NOT remove:
//   - Statements created as premises or conclusions (they may be shared)
//   - Issues created for rebuttals (they may involve other arguments)
//   - The undercutter statement itself (may be referenced by other arguments)
//
// This is safe for single-move rollback because:
//   - AddArgumentDirect/AddRebuttal/AddUndercutter/AddUndermining each
//     create exactly one argument.
//   - AddSupport may create multiple arguments; use RemoveArguments (plural)
//     to undo all of them.
//   - Statements and Issues without arguments are harmless — they just
//     become unsupported and label Out after the next Evaluate().
//
// Usage in the turn pipeline:
//
//	argIDs, err := adapter.AddUndercutter(...)
//	if !adapter.IsAcyclic() {
//	    adapter.RemoveArgument(argIDs[0])
//	    return error("creates cycle")
//	}
func (a *Adapter) RemoveArgument(argID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	arg := a.Graph.Arguments[argID]
	if arg == nil {
		return fmt.Errorf("argument %q not found", argID)
	}

	// Remove from conclusion's Args slice (bidirectional link cleanup).
	if arg.Conclusion != nil {
		a.detachArgumentFromStatement(arg, arg.Conclusion)
	}

	// Remove from graph map.
	delete(a.Graph.Arguments, argID)

	return nil
}

// RemoveArguments removes multiple arguments. Used to undo AddSupport which
// may create more than one argument (when the scheme has multiple conclusions).
func (a *Adapter) RemoveArguments(argIDs []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, argID := range argIDs {
		arg := a.Graph.Arguments[argID]
		if arg == nil {
			return fmt.Errorf("argument %q not found during batch remove", argID)
		}
		if arg.Conclusion != nil {
			a.detachArgumentFromStatement(arg, arg.Conclusion)
		}
		delete(a.Graph.Arguments, argID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Graph inspection helpers
// ---------------------------------------------------------------------------

// StatementLabels returns a map of statement ID → Label for all statements.
// Useful for comparing labels before and after a move (delta scoring).
func (a *Adapter) StatementLabels() map[string]caes.Label {
	a.mu.Lock()
	defer a.mu.Unlock()

	labels := make(map[string]caes.Label, len(a.Graph.Statements))
	for id, s := range a.Graph.Statements {
		labels[id] = s.Label
	}
	return labels
}

// ArgumentOwner returns the owner metadata for an argument, or "" if not set.
func (a *Adapter) ArgumentOwner(argID string) string {
	a.mu.Lock()
	defer a.mu.Unlock()

	arg := a.Graph.Arguments[argID]
	if arg == nil || arg.Metadata == nil {
		return ""
	}
	if owner, ok := arg.Metadata["owner"].(string); ok {
		return owner
	}
	return ""
}

// SchemeIDs returns the list of scheme IDs available in the Theory.
// Useful for diagnostics and for the NASCQ layer to validate scheme references.
func (a *Adapter) SchemeIDs() []string {
	a.mu.Lock()
	defer a.mu.Unlock()

	ids := make([]string, 0, len(a.schemeMap))
	for id := range a.schemeMap {
		ids = append(ids, id)
	}
	return ids
}
