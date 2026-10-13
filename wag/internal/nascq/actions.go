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
	Variables            map[string]VariableSpec
	PremiseTemplatesEN   []string
	ConclusionTemplateEN string

	// Resolved from CQ mapping (for attacks)
	CaesAttackType string

	// For a move with a target: the part of the target it is about (see
	// Locus) and the premise or exception statement for those loci.
	Locus       Locus
	LocusStmtID string

	// AnswerVariable is the scheme's variable bound to the answer (see
	// ASNode.AnswerVariable): the player need not fill it in.
	AnswerVariable string
}

// EnumerateActions returns all structurally valid action templates available
// given the current argument graph. This drives the UI's action wizard and
// the AI agent's decision-making.
//
// The enumeration includes:
//  1. Direct support with every scheme (no target needed)
//  2. For each argument in the graph:
//     a. Find the argument's NASCQ scheme (via reverse index)
//     b. Get the ask-CQs for that scheme
//     c. For each CQ, get backup schemes
//     d. Generate attack templates for each (target, CQ, backup) triple;
//     an undermining template carries the premise its CQ's pattern
//     designates, and is left out if the target has no such premise
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

	// 2. Answers to the critical questions about existing arguments.
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

			// The locus decides what the answers are about; a premise or
			// exception the target does not have cannot be asked about.
			locus := n.LocusOf(cqID, nascqSchemeID)
			stmtID := ""
			switch locus.Kind {
			case "":
				continue // no locus (Lint reports it)
			case LocusPremise:
				if stmtID = graph.PremiseMatching(argID, locus.Pattern); stmtID == "" {
					continue
				}
			case LocusException:
				if stmtID = graph.ExceptionMatching(argID, locus.Pattern); stmtID == "" {
					continue
				}
			}
			premiseID := ""
			if locus.Kind == LocusPremise {
				premiseID = stmtID
			}
			attackType := locus.AttackType()

			// Each backup scheme can back up either answer: the negative one
			// (an attack) and the positive one (a defence, a support with a
			// target).
			for _, bs := range n.GetBackupSchemes(cqID) {
				for _, typ := range []string{"attack", "support"} {
					tmpl := ActionTemplate{
						Type:                 typ,
						SchemeID:             bs.ID,
						SchemeName:           bs.Name,
						AnswersCQID:          cqID,
						CQName:               cq.Name,
						TargetArgID:          argID,
						Variables:            bs.Variables,
						PremiseTemplatesEN:   bs.PremiseTemplatesEN,
						ConclusionTemplateEN: bs.ConclusionTemplateEN,
						TargetPremiseID:      premiseID,
						Locus:                locus,
						LocusStmtID:          stmtID,
						AnswerVariable:       bs.AnswerVariable,
					}
					if typ == "attack" {
						tmpl.AttackType, tmpl.CaesAttackType = attackType, attackType
					}
					actions = append(actions, tmpl)
				}
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
