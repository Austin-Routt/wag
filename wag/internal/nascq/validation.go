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

	// PremiseMatching returns the ID of the premise of argID that a CQ's
	// target_premise_pattern (e.g. "possible({A})") designates, with the
	// argument's own variable bindings substituted, or "" if none does.
	PremiseMatching(argID, pattern string) string
}

// attackTypes are the CAES operations an attack can perform.
var attackTypes = map[string]bool{"rebuttal": true, "undercutter": true, "undermining": true}

// AttackTypeOf returns the CAES attack operation that answers the critical
// question: the answer type of its caes_mapping that is not "support". It
// returns "" if neither answer type is an attack type.
func AttackTypeOf(cq *CQNode) string {
	if cq == nil {
		return ""
	}
	if attackTypes[cq.CaesMapping.PositiveAnswerType] {
		return cq.CaesMapping.PositiveAnswerType
	}
	if attackTypes[cq.CaesMapping.NegativeAnswerType] {
		return cq.CaesMapping.NegativeAnswerType
	}
	return ""
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
	// Index carneadesID → nascqID. When several NASCQ schemes map to the same
	// Carneades scheme, the one with the smallest NASCQ ID wins, so the choice
	// does not depend on map iteration order.
	idx := make(map[string]string, len(n.Schemes))
	for _, nascqID := range sortedSchemeIDs(n.Schemes) {
		node := n.Schemes[nascqID]
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

// InstantiateTemplate replaces {VarName} placeholders in a template string
// with the corresponding values from the bindings map; placeholders without
// a binding are left as they are.
func InstantiateTemplate(template string, bindings map[string]string) string {
	return instantiateTemplate(template, bindings)
}

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
	names := make([]string, 0, len(bindings))
	for varName := range bindings {
		names = append(names, varName)
	}
	sort.Strings(names) // a value containing "{X}" must not make the result order-dependent
	for _, varName := range names {
		result = strings.ReplaceAll(result, "{"+varName+"}", bindings[varName])
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
//	S0: Move shape — a support has no target or CQ; an attack has a target and
//	    answers a CQ or names an attack type (rebuttal, undercutter, undermining)
//	S1: Scheme existence — the NASCQ scheme ID must exist
//	S2: Variable completeness — all scheme variables must have bindings
//	S3: CQ applicability — the CQ must be askable of the target's scheme (a
//	    target made with no NASCQ scheme admits no CQ), and an attack type the
//	    mover names must be the one the CQ maps to
//	S5: Backup edge validity — if CQ-directed, the scheme must be a valid backup for the CQ
//	S6: Template instantiation — substitute variables into English templates
//	S7: Target argument exists — if attacking, the target must exist in the graph
//	S8: Target premise — an undermining names a premise of the target, or its
//	    CQ's target_premise_pattern designates one; the resolved premise is
//	    returned in the ValidatedAction
//
// S4 (semantic types) is not a move check: S5 admits only schemes with an
// explicit backup edge, which the NASCQ author wired, or thesis-independent
// ones, which back every CQ, so a type comparison after S5 could never fail.
// Semantic types are checked where they can fail, on the data (see Lint).
func (n *NASCQ) Validate(action ProposedAction, graph GraphQuerier) (*ValidatedAction, error) {
	// S0: Move shape
	switch action.Type {
	case "support":
		if action.TargetArgID != "" || action.AnswersCQID != "" {
			return nil, fmt.Errorf("S0: a support move takes no target or critical question " +
				"(answering a critical question in the target's favour is not implemented)")
		}
	case "attack":
		if action.TargetArgID == "" {
			return nil, fmt.Errorf("S0: an attack needs a target argument")
		}
		if action.AnswersCQID == "" && !attackTypes[action.AttackType] {
			return nil, fmt.Errorf("S0: an attack must answer a critical question or name an attack type "+
				"(rebuttal, undercutter, undermining); got %q", action.AttackType)
		}
	default:
		return nil, fmt.Errorf("S0: unknown action type %q (want support or attack)", action.Type)
	}

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
			if nascqSchemeID == "" {
				return nil, fmt.Errorf("S3: target %q was not made with a NASCQ scheme, so no critical question applies to it",
					action.TargetArgID)
			}
			askCQs := n.askIndex[nascqSchemeID]
			if !containsStr(askCQs, action.AnswersCQID) {
				return nil, fmt.Errorf("S3: CQ %q is not applicable to scheme %q (ask CQs: %v)",
					action.AnswersCQID, nascqSchemeID, askCQs)
			}
		}
	}

	// S5: Backup edge validity: an explicit backup edge, or a
	// thesis-independent scheme (which backs every CQ).
	if action.AnswersCQID != "" {
		backed := false
		for _, bs := range n.GetBackupSchemes(action.AnswersCQID) {
			if bs.ID == action.SchemeID {
				backed = true
				break
			}
		}
		if !backed {
			return nil, fmt.Errorf("S5: scheme %q cannot back up CQ %q "+
				"(no backup edge and scheme is not thesis-independent)",
				action.SchemeID, action.AnswersCQID)
		}
	}

	// S6: Template instantiation
	instPremises, instConclusion := instantiateTemplates(scheme, action.VariableBindings)

	// S3 (continued): the CAES attack type comes from the CQ's mapping. An
	// attack type the mover names must agree with it.
	caesAttackType := ""
	if action.Type == "attack" {
		caesAttackType = action.AttackType
		if action.AnswersCQID != "" {
			mapped := AttackTypeOf(n.Questions[action.AnswersCQID])
			if mapped == "" {
				return nil, fmt.Errorf("S3: CQ %q maps to no attack type", action.AnswersCQID)
			}
			if action.AttackType != "" && action.AttackType != mapped {
				return nil, fmt.Errorf("S3: CQ %q is answered by an attack of type %q, not %q",
					action.AnswersCQID, mapped, action.AttackType)
			}
			caesAttackType = mapped
		}
	}

	// S8: Target premise (undermining only). The resolved premise is returned.
	if caesAttackType == "undermining" {
		if graph == nil {
			return nil, fmt.Errorf("S8: graph querier required for undermining")
		}
		if action.TargetPremiseID != "" {
			if !graph.PremiseExists(action.TargetArgID, action.TargetPremiseID) {
				return nil, fmt.Errorf("S8: premise %q not found in argument %q",
					action.TargetPremiseID, action.TargetArgID)
			}
		} else {
			pattern := ""
			if cq := n.Questions[action.AnswersCQID]; cq != nil {
				pattern = cq.CaesMapping.TargetPremisePattern
			}
			if pattern == "" {
				return nil, fmt.Errorf("S8: name the premise of %q to undermine (target_premise_id)",
					action.TargetArgID)
			}
			action.TargetPremiseID = graph.PremiseMatching(action.TargetArgID, pattern)
			if action.TargetPremiseID == "" {
				return nil, fmt.Errorf("S8: argument %q has no premise matching %q",
					action.TargetArgID, pattern)
			}
		}
	}
	action.AttackType = caesAttackType

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
