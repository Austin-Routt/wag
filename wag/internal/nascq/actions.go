package nascq

import "sort"

// ActionTemplate represents a structurally valid move skeleton that a player
// can fill in with variable bindings. Enumerated by EnumerateActions.
//
// The template has everything EXCEPT the variable bindings — those are
// player-supplied (human input or agent-generated).
type ActionTemplate struct {
	// Action structure
	Type        string // "support", "attack" or "question"
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

	// TargetStmtID is the statement a move about a statement argues for
	// (support) or against (attack); Fixed are the values of the scheme's
	// variables that the statement fixes (see Validate, S9), which the
	// player need not fill in either.
	TargetStmtID string
	Fixed        map[string]string
}

// Action is the move made from the template with the given bindings.
func (t ActionTemplate) Action(bindings map[string]string) ProposedAction {
	return ProposedAction{Type: t.Type, AttackType: t.AttackType, SchemeID: t.SchemeID, AnswersCQID: t.AnswersCQID,
		TargetArgID: t.TargetArgID, TargetPremiseID: t.TargetPremiseID, TargetStmtID: t.TargetStmtID,
		VariableBindings: bindings}
}

// Bound reports whether the game, not the player, binds the variable: the
// answer variable, or one the statement of the move fixes.
func (t ActionTemplate) Bound(v string) bool {
	_, fixed := t.Fixed[v]
	return v == t.AnswerVariable && v != "" || fixed
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
//     d. Generate a question template for each (target, CQ) about a premise
//     or the inference, and an attack and a defence template for each
//     (target, CQ, backup) triple; a template about a premise or an
//     exception carries the statement its CQ's pattern designates, and is
//     left out if the target has no such premise
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
			// exception the target does not have cannot be asked about. A
			// premise locus without a pattern is about any premise that no
			// other question of the scheme is about: one set of templates per
			// premise.
			locus := n.LocusOf(cqID, nascqSchemeID)
			stmtIDs := []string{""}
			switch locus.Kind {
			case "":
				continue // no locus (Lint reports it)
			case LocusConclusion:
				if !n.OwnConclusion(graph, argID, nascqSchemeID) {
					continue
				}
			case LocusPremise:
				if locus.Pattern == "" { // any premise no other question is about
					covered := n.CoveredPremises(graph, argID, nascqSchemeID)
					stmtIDs = nil
					for _, id := range graph.PremiseIDs(argID) {
						if !covered[id] {
							stmtIDs = append(stmtIDs, id)
						}
					}
				} else {
					stmtIDs = []string{graph.PremiseMatching(argID, locus.Pattern)}
				}
			case LocusException:
				stmtIDs = []string{graph.ExceptionMatching(argID, locus.Pattern)}
			}
			for _, stmtID := range stmtIDs {
				if stmtID == "" && (locus.Kind == LocusPremise || locus.Kind == LocusException) {
					continue
				}
				for _, t := range n.answerTemplates(argID, cq, locus, stmtID) {
					t.Fixed = n.answerFixed(t, graph)
					actions = append(actions, t)
				}
			}
		}
	}

	return actions
}

// answerFixed returns the values an answer from the template must give the
// variables of its backing scheme's conclusion (S9, fitAnswer): when the
// scheme concludes the answer's form, the answer fixes them.
func (n *NASCQ) answerFixed(t ActionTemplate, graph GraphQuerier) map[string]string {
	s := n.Schemes[t.SchemeID]
	if t.Type == "question" || s == nil || s.Free || s.AnswerVariable != "" {
		return nil
	}
	va := &ValidatedAction{ProposedAction: ProposedAction{Type: t.Type, TargetArgID: t.TargetArgID, AnswersCQID: t.AnswersCQID},
		Locus: t.Locus, LocusStmtID: t.LocusStmtID}
	stmt, _ := n.concluded(va, graph)
	if stmt == "" {
		return nil
	}
	fixed, err := graph.Fits(s.CarneadesID, stmt, nil)
	if err != nil || len(fixed) == 0 {
		return nil
	}
	return fixed
}

// answerTemplates returns the templates of the moves about one locus of an
// argument: the question alone (for a premise or the inference), and an
// attack and a defence with each scheme that can back an answer.
func (n *NASCQ) answerTemplates(argID string, cq *CQNode, locus Locus, stmtID string) []ActionTemplate {
	var actions []ActionTemplate
	premiseID := ""
	if locus.Kind == LocusPremise {
		premiseID = stmtID
	}

	// A question alone (no scheme) about a premise or the inference (Q1 of
	// Validate).
	if locus.Kind == LocusPremise || locus.Kind == LocusArgument {
		actions = append(actions, ActionTemplate{
			Type:            "question",
			AnswersCQID:     cq.ID,
			CQName:          cq.Name,
			TargetArgID:     argID,
			TargetPremiseID: premiseID,
			Locus:           locus,
			LocusStmtID:     stmtID,
		})
	}

	// Each backup scheme can back up either answer: the negative one (an
	// attack) and the positive one (a defence, a support with a target).
	// The challenge of a premise is a question only: arguments for and
	// against a premise are moves about the statement
	// (EnumerateStatementActions).
	if cq.ID == CQPremiseAcceptable {
		return actions
	}
	attackType := locus.AttackType()
	for _, bs := range n.GetBackupSchemes(cq.ID) {
		for _, typ := range []string{"attack", "support"} {
			tmpl := ActionTemplate{
				Type:                 typ,
				SchemeID:             bs.ID,
				SchemeName:           bs.Name,
				AnswersCQID:          cq.ID,
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
	return actions
}

// EnumerateStatementActions returns the templates of the moves about the
// given statements: for each statement, an argument for it and one against
// it with every scheme that can conclude it, or its opposite (S9 of
// Validate): Other and the schemes with an answer variable conclude any
// statement; another scheme only one its conclusion unifies with, and the
// template carries the values that fixes.
func (n *NASCQ) EnumerateStatementActions(graph GraphQuerier, stmtIDs []string) []ActionTemplate {
	var actions []ActionTemplate
	for _, stmtID := range stmtIDs {
		info := graph.Statement(stmtID)
		if !info.Exists {
			continue
		}
		for _, typ := range []string{"support", "attack"} {
			concl := stmtID
			attackType := ""
			if typ == "attack" {
				concl, attackType = info.Opposite, "rebuttal"
				if info.Premise && !info.Concluded {
					attackType = "undermining"
				}
			}
			for _, id := range sortedSchemeIDs(n.Schemes) {
				s := n.Schemes[id]
				var fixed map[string]string
				if !s.Free && s.AnswerVariable == "" {
					if concl == "" {
						continue
					}
					f, err := graph.Fits(s.CarneadesID, concl, nil)
					if err != nil {
						continue
					}
					fixed = f
				}
				actions = append(actions, ActionTemplate{
					Type:                 typ,
					AttackType:           attackType,
					CaesAttackType:       attackType,
					SchemeID:             s.ID,
					SchemeName:           s.Name,
					Variables:            s.Variables,
					PremiseTemplatesEN:   s.PremiseTemplatesEN,
					ConclusionTemplateEN: s.ConclusionTemplateEN,
					Locus:                Locus{Kind: LocusStatement},
					LocusStmtID:          stmtID,
					AnswerVariable:       s.AnswerVariable,
					TargetStmtID:         stmtID,
					Fixed:                fixed,
				})
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
