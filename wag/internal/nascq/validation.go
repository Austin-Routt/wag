package nascq

import (
	"fmt"
	"sort"
	"strings"
)

// GraphQuerier is a minimal read-only interface that the NASCQ validator
// needs from the argument graph. The Carneades adapter (or a thin wrapper)
// implements this interface.
//
// This keeps the nascq package free of carneades-4 imports while still
// allowing S3, S7, and S8 checks that require graph inspection.
type GraphQuerier interface {
	// ArgumentExists returns true if an argument with the given ID exists.
	ArgumentExists(argID string) bool

	// ArgumentSchemeID returns the Carneades scheme ID (e.g., "practical_reasoning1")
	// for the given argument, or "" if the argument doesn't exist or has no scheme.
	ArgumentSchemeID(argID string) string

	// PremiseExists returns true if the argument with argID has a premise
	// whose statement ID matches premiseStmtID.
	PremiseExists(argID, premiseStmtID string) bool
}

// ---------------------------------------------------------------------------
// Reverse index: Carneades scheme ID → NASCQ scheme ID
// ---------------------------------------------------------------------------

// buildReverseSchemeIndex creates a mapping from CarneadesID → NASCQ ID.
// Called during NASCQ construction so that findNASCQSchemeID is O(1).
//
// Multiple NASCQ schemes can map to the same Carneades ID (many-to-one).
// In that case, the first one found wins. This is acceptable because the
// reverse lookup is used for S3 (CQ applicability), where any matching
// NASCQ scheme is sufficient.
func (n *NASCQ) buildReverseSchemeIndex() {
	n.reverseSchemeIndex = make(map[string]string, len(n.Schemes))
	for nascqID, node := range n.Schemes {
		if node.CarneadesID != "" {
			// Only store the first mapping (don't overwrite).
			if _, exists := n.reverseSchemeIndex[node.CarneadesID]; !exists {
				n.reverseSchemeIndex[nascqID] = node.CarneadesID
				// We index carneadesID → nascqID
			}
		}
	}
	// Rebuild: index is carneadesID → nascqID
	idx := make(map[string]string, len(n.Schemes))
	for nascqID, node := range n.Schemes {
		if node.CarneadesID != "" {
			if _, exists := idx[node.CarneadesID]; !exists {
				idx[node.CarneadesID] = nascqID
			}
		}
	}
	n.reverseSchemeIndex = idx
}

// FindNASCQSchemeID returns the NASCQ scheme ID for a given Carneades scheme ID.
// Returns "" if no mapping exists.
//
// Example: FindNASCQSchemeID("practical_reasoning1") → "AS:practical_reasoning"
func (n *NASCQ) FindNASCQSchemeID(carneadesSchemeID string) string {
	return n.reverseSchemeIndex[carneadesSchemeID]
}

// ---------------------------------------------------------------------------
// Template instantiation
// ---------------------------------------------------------------------------

// instantiateTemplate replaces {VarName} placeholders in a template string
// with the corresponding values from the bindings map.
//
// Example:
//
//	template: "{Agent} has goal {Goal}"
//	bindings: {"Agent": "Alice", "Goal": "reduce emissions"}
//	result:   "Alice has goal reduce emissions"
func instantiateTemplate(template string, bindings map[string]string) string {
	result := template
	for varName, value := range bindings {
		result = strings.ReplaceAll(result, "{"+varName+"}", value)
	}
	return result
}

// instantiateTemplates applies variable bindings to all premise and conclusion
// templates of a scheme, returning the instantiated English strings.
func instantiateTemplates(scheme *ASNode, bindings map[string]string) (premises []string, conclusion string) {
	premises = make([]string, len(scheme.PremiseTemplatesEN))
	for i, t := range scheme.PremiseTemplatesEN {
		premises[i] = instantiateTemplate(t, bindings)
	}
	conclusion = instantiateTemplate(scheme.ConclusionTemplateEN, bindings)
	return
}

// ---------------------------------------------------------------------------
// Validate: S1–S8 structural checks
// ---------------------------------------------------------------------------

// Validate checks a ProposedAction against the NASCQ structural rules.
// If all checks pass, it returns a ValidatedAction with resolved Carneades
// scheme ID, instantiated templates, CAES attack type, and weight.
//
// The graph parameter can be nil for direct support actions (no target),
// but is required for attack actions (S3, S7, S8 need graph inspection).
//
// Checks performed (from WAG plan §13, SRD §7.5):
//
//	S1: Scheme existence — the NASCQ scheme ID must exist
//	S2: Variable completeness — all scheme variables must have bindings
//	S3: CQ applicability — if CQ-directed, the CQ must be askable of the target's scheme
//	S4: Semantic type compatibility — if CQ-directed and thesis-dependent, types must match
//	S5: Backup edge validity — if CQ-directed, the scheme must be a valid backup for the CQ
//	S6: Template instantiation — substitute variables into English templates
//	S7: Target argument exists — if attacking, the target must exist in the graph
//	S8: Target premise exists — for undermining, the specific premise must exist
func (n *NASCQ) Validate(action ProposedAction, graph GraphQuerier) (*ValidatedAction, error) {
	// S1: Scheme existence
	scheme, ok := n.Schemes[action.SchemeID]
	if !ok {
		avail := make([]string, 0, len(n.Schemes))
		for k := range n.Schemes {
			avail = append(avail, k)
		}
		sort.Strings(avail)
		return nil, fmt.Errorf("S1: unknown scheme %q; available: %v",
			action.SchemeID, avail)
	}

	// S2: Variable completeness
	missing := []string{}
	for varName := range scheme.Variables {
		if _, ok := action.VariableBindings[varName]; !ok {
			missing = append(missing, varName)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("S2: missing variable binding(s): %v", missing)
	}

	// S7: Target argument exists (checked early so S3 can use it)
	if action.TargetArgID != "" {
		if graph == nil {
			return nil, fmt.Errorf("S7: graph querier required for attack actions")
		}
		if !graph.ArgumentExists(action.TargetArgID) {
			return nil, fmt.Errorf("S7: target argument %q not found in graph",
				action.TargetArgID)
		}
	}

	// S3: CQ applicability (if CQ-directed)
	if action.AnswersCQID != "" {
		if _, ok := n.Questions[action.AnswersCQID]; !ok {
			return nil, fmt.Errorf("S3: unknown CQ %q", action.AnswersCQID)
		}
		if action.TargetArgID != "" && graph != nil {
			caesSchemeID := graph.ArgumentSchemeID(action.TargetArgID)
			nascqSchemeID := n.FindNASCQSchemeID(caesSchemeID)
			if nascqSchemeID == "" {
				if _, ok := n.Schemes[caesSchemeID]; ok {
					nascqSchemeID = caesSchemeID
				}
			}
			if nascqSchemeID != "" {
				askCQs := n.askIndex[nascqSchemeID]
				if !containsStr(askCQs, action.AnswersCQID) {
					return nil, fmt.Errorf("S3: CQ %q is not applicable to scheme %q (ask CQs: %v)",
						action.AnswersCQID, nascqSchemeID, askCQs)
				}
			}
		}
	}

	// S5: Backup edge validity (checked BEFORE S4 because an explicit
	// backup edge overrides semantic type checking — the NASCQ author
	// already verified compatibility when wiring the edge)
	hasExplicitBackup := false
	if action.AnswersCQID != "" {
		backups := n.GetBackupSchemes(action.AnswersCQID)
		for _, bs := range backups {
			if bs.ID == action.SchemeID {
				hasExplicitBackup = true
				break
			}
		}
		if !hasExplicitBackup {
			return nil, fmt.Errorf("S5: scheme %q cannot back up CQ %q "+
				"(no backup edge and scheme is not thesis-independent)",
				action.SchemeID, action.AnswersCQID)
		}
	}

	// S4: Semantic type compatibility (only if CQ-directed, thesis-dependent,
	// and NO explicit backup edge — an explicit edge or thesis-independence
	// bypasses this check)
	if action.AnswersCQID != "" && !scheme.ThesisIndependent && !hasExplicitBackup {
		cq := n.Questions[action.AnswersCQID]
		if cq != nil && cq.SemanticType != "" && scheme.SemanticType != "" {
			if cq.SemanticType != scheme.SemanticType {
				return nil, fmt.Errorf("S4: semantic mismatch: CQ %q expects type %q, "+
					"but scheme %q provides type %q",
					action.AnswersCQID, cq.SemanticType,
					action.SchemeID, scheme.SemanticType)
			}
		}
	}

	// S6: Template instantiation
	instPremises, instConclusion := instantiateTemplates(scheme, action.VariableBindings)

	// S8: Target premise exists (undermining only)
	if action.AttackType == "undermining" && action.TargetPremiseID != "" {
		if graph == nil {
			return nil, fmt.Errorf("S8: graph querier required for undermining")
		}
		if !graph.PremiseExists(action.TargetArgID, action.TargetPremiseID) {
			return nil, fmt.Errorf("S8: premise %q not found in argument %q",
				action.TargetPremiseID, action.TargetArgID)
		}
	}

	// Resolve CAES attack type from CQ mapping
	caesAttackType := action.AttackType
	if action.AnswersCQID != "" && action.Type == "attack" {
		cq := n.Questions[action.AnswersCQID]
		if cq != nil {
			if cq.CaesMapping.PositiveAnswerType != "support" {
				caesAttackType = cq.CaesMapping.PositiveAnswerType
			} else {
				caesAttackType = cq.CaesMapping.NegativeAnswerType
			}
		}
	}

	return &ValidatedAction{
		ProposedAction:    action,
		CarneadesSchemeID: scheme.CarneadesID,
		InstPremises:      instPremises,
		InstConclusion:    instConclusion,
		CaesAttackType:    caesAttackType,
		Weight:            scheme.BaseWeight,
	}, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func containsStr(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
