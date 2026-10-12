package nascq

import "sort"

// ActionTemplate represents a structurally valid move skeleton that a player
// can fill in with variable bindings. Enumerated by EnumerateActions.
//
// The template has everything EXCEPT the variable bindings — those are
// player-supplied (human input or agent-generated).
type ActionTemplate struct {
	// Action structure
	Type        string // "support" or "attack"
	AttackType  string // "rebuttal", "undercutter", "undermining" (attacks only)
	SchemeID    string // NASCQ scheme ID
	SchemeName  string // human-readable name
	AnswersCQID string // CQ being answered (empty for direct support)
	CQName      string // human-readable CQ name
	TargetArgID string // argument being attacked (empty for support)

	// For undermining: which premise to target
	TargetPremiseID string

	// What the player needs to fill in
	Variables          map[string]VariableSpec
	PremiseTemplatesEN []string
	ConclusionTemplateEN string

	// Resolved from CQ mapping (for attacks)
	CaesAttackType string
}

// EnumerateActions returns all structurally valid action templates available
// given the current argument graph. This drives the UI's action wizard and
// the AI agent's decision-making.
//
// The enumeration includes:
//   1. Direct support with every scheme (no target needed)
//   2. For each argument in the graph:
//      a. Find the argument's NASCQ scheme (via reverse index)
//      b. Get the ask-CQs for that scheme
//      c. For each CQ, get backup schemes
//      d. Generate attack templates for each (target, CQ, backup) triple
//
// Variable bindings are NOT included — the player/agent fills those in.
func (n *NASCQ) EnumerateActions(graph GraphQuerier, argIDs []string) []ActionTemplate {
	var actions []ActionTemplate

	// 1. Direct support with every scheme, in sorted scheme-ID order (map
	// iteration order is random; the enumeration order must be reproducible).
	for _, schemeID := range sortedSchemeIDs(n.Schemes) {
		scheme := n.Schemes[schemeID]
		actions = append(actions, ActionTemplate{
			Type:                 "support",
			SchemeID:             scheme.ID,
			SchemeName:           scheme.Name,
			Variables:            scheme.Variables,
			PremiseTemplatesEN:   scheme.PremiseTemplatesEN,
			ConclusionTemplateEN: scheme.ConclusionTemplateEN,
		})
	}

	// 2. CQ-directed attacks on existing arguments.
	for _, argID := range argIDs {
		caesSchemeID := graph.ArgumentSchemeID(argID)
		if caesSchemeID == "" {
			continue
		}
		nascqSchemeID := n.FindNASCQSchemeID(caesSchemeID)
		if nascqSchemeID == "" {
			// Fallback: maybe carneades ID IS the NASCQ ID.
			if _, ok := n.Schemes[caesSchemeID]; ok {
				nascqSchemeID = caesSchemeID
			} else {
				continue
			}
		}

		// Get ask-CQs for this scheme.
		askCQIDs := n.askIndex[nascqSchemeID]
		for _, cqID := range askCQIDs {
			cq := n.Questions[cqID]
			if cq == nil {
				continue
			}

			// Get backup schemes for this CQ.
			backups := n.GetBackupSchemes(cqID)
			for _, bs := range backups {
				// Determine attack type from CQ mapping.
				attackType := cq.CaesMapping.NegativeAnswerType
				if cq.CaesMapping.PositiveAnswerType != "support" {
					attackType = cq.CaesMapping.PositiveAnswerType
				}
				if attackType == "support" {
					attackType = cq.CaesMapping.NegativeAnswerType
				}

				tmpl := ActionTemplate{
					Type:                 "attack",
					AttackType:           attackType,
					SchemeID:             bs.ID,
					SchemeName:           bs.Name,
					AnswersCQID:          cqID,
					CQName:              cq.Name,
					TargetArgID:          argID,
					Variables:            bs.Variables,
					PremiseTemplatesEN:   bs.PremiseTemplatesEN,
					ConclusionTemplateEN: bs.ConclusionTemplateEN,
					CaesAttackType:       attackType,
				}

				// For undermining: enumerate premises of the target argument.
				if attackType == "undermining" {
					// We can't enumerate premise IDs without accessing the
					// graph in detail. Leave TargetPremiseID empty — the
					// agent or UI must fill it in. For random agents, we
					// handle this in the agent code.
				}

				actions = append(actions, tmpl)
			}
		}
	}

	return actions
}

// ArgumentIDs returns a list of all argument IDs accessible through the
// GraphQuerier. This is a helper for EnumerateActions callers.
// Since GraphQuerier is minimal, the game layer provides the ID list.

func sortedSchemeIDs(m map[string]*ASNode) []string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
