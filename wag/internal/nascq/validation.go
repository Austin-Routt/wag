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

	// ExceptionMatching returns the ID of the exception of argID that an
	// exception pattern (e.g. "side_effects({A},{S1})") designates, with the
	// argument's own bindings substituted, or "" if none does. The statement
	// need not exist yet.
	ExceptionMatching(argID, pattern string) string

	// PremiseIDs returns the statement IDs of the argument's premises, in
	// order (for questions about any premise, such as those asked of the
	// user-defined scheme).
	PremiseIDs(argID string) []string

	// ArgumentConclusion returns the statement ID of the argument's
	// conclusion, or "" if there is no such argument.
	ArgumentConclusion(argID string) string

	// Statement describes a statement of the graph, for a move about it.
	Statement(stmtID string) StatementInfo

	// Fits returns the values (terms) that the variables of the conclusion
	// of the Carneades scheme schemeID take when an argument of the scheme
	// concludes the statement stmtID, or an error saying why no argument of
	// it can: its conclusion does not unify with the statement, or a value
	// the mover gave (given) differs from the one the statement fixes.
	Fits(schemeID, stmtID string, given map[string]string) (map[string]string, error)
}

// StatementInfo is what a move about a statement needs to know of it.
type StatementInfo struct {
	Exists    bool
	Premise   bool // a premise of some argument
	Concluded bool // the conclusion of some argument made by a move
	// Opposite is the statement an argument against this one concludes: the
	// other position of its issue, or its negation ¬S. It is "" when the
	// statement is one of more than two positions of an issue (argue for
	// one of the others instead).
	Opposite string
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
// A move is a plain support (no target: a new argument), an attack (a
// negative answer to a critical question about the target, or a direct
// rebuttal, undercutter or undermining), a defence (a support with a target:
// a positive answer to a critical question, or a direct backing of a premise
// or of the conclusion), an argument for (support) or against (attack) a
// statement on the board (TargetStmtID: the claim, or any premise,
// conclusion or exception), or a question (a critical question asked without
// an answer, which the target's proponent must answer). The returned
// ValidatedAction carries the locus the move is about and, for a premise,
// exception or statement, its statement.
//
// Checks performed (from WAG plan §13, SRD §7.5):
//
//	S0: Move shape — an attack has a target and answers a CQ or names an
//	    attack type (rebuttal, undercutter, undermining); a defence that answers
//	    a CQ or names a premise has a target; a question names a target and a
//	    CQ (and no scheme: S1, S2, S5 and S6 do not apply to it); a move about
//	    a statement is a support or an attack and names nothing else
//	S1: Scheme existence — the NASCQ scheme ID must exist
//	S2: Variable completeness — all scheme variables must have bindings,
//	    except, on a move with a target, the scheme's answer variable
//	S3: CQ applicability — the target was made with a NASCQ scheme, the CQ is
//	    asked of that scheme and has a locus there (NASCQ.LocusOf), and an
//	    attack type the mover names is the one the locus implies
//	S5: Backup edge validity — if CQ-directed, the scheme must be a valid backup for the CQ
//	S6: Template instantiation — substitute variables into English templates
//	S7: Target exists — the target argument, or the target statement, is in the graph
//	Q1: Question locus — a question alone can be about a premise or the
//	    inference only; an exception or a counterclaim must be argued
//	S8: Locus statement — a premise locus names a premise of the target (the
//	    mover's TargetPremiseID, or the locus pattern with the target's
//	    bindings); an exception locus designates an exception of the target
//	S9: Scheme fit — a direct move or a move about a statement concludes
//	    the statement it is about, or its opposite, so its scheme must be
//	    able to conclude that (see fit)
//
// The checks run in the order S0, S1, the target (S7, S3, Q1, S5, S8),
// S9, S2, S6: S9 fixes some variables that S2 would otherwise ask for.
//
// S4 (semantic types) is not a move check: S5 admits only schemes with an
// explicit backup edge, which the NASCQ author wired, or thesis-independent
// ones, which back every CQ, so a type comparison after S5 could never fail.
// Semantic types are checked where they can fail, on the data (see Lint).
func (n *NASCQ) Validate(action ProposedAction, graph GraphQuerier) (*ValidatedAction, error) {
	// S0: Move shape
	switch {
	case action.TargetStmtID != "":
		if action.Type != "support" && action.Type != "attack" {
			return nil, fmt.Errorf("S0: a move about a statement argues for it (support) or against it (attack), not %q", action.Type)
		}
		if action.TargetArgID != "" || action.AnswersCQID != "" || action.TargetPremiseID != "" {
			return nil, fmt.Errorf("S0: a move about a statement names the statement only (target_stmt_id), " +
				"not an argument, a premise or a critical question")
		}
		action.AttackType = ""
	case action.Type == "support":
		if action.TargetArgID == "" && (action.AnswersCQID != "" || action.TargetPremiseID != "") {
			return nil, fmt.Errorf("S0: a defence needs a target argument")
		}
	case action.Type == "attack":
		if action.TargetArgID == "" {
			return nil, fmt.Errorf("S0: an attack needs a target argument or a target statement")
		}
		if action.AnswersCQID == "" && !attackTypes[action.AttackType] {
			return nil, fmt.Errorf("S0: an attack must answer a critical question or name an attack type "+
				"(rebuttal, undercutter, undermining); got %q", action.AttackType)
		}
	case action.Type == "question":
		if action.TargetArgID == "" || action.AnswersCQID == "" {
			return nil, fmt.Errorf("S0: a question names a target argument and a critical question")
		}
		action.SchemeID, action.VariableBindings, action.AttackType = "", nil, ""
	default:
		return nil, fmt.Errorf("S0: unknown action type %q (want support, attack or question)", action.Type)
	}

	if action.Type == "question" {
		return n.validateTarget(action, &ValidatedAction{}, graph)
	}

	// S1: Scheme existence
	scheme, ok := n.Schemes[action.SchemeID]
	if !ok {
		return nil, fmt.Errorf("S1: unknown scheme %q; available: %v", action.SchemeID, sortedSchemeIDs(n.Schemes))
	}

	va := &ValidatedAction{ProposedAction: action}
	var err error
	switch {
	case action.TargetStmtID != "":
		if va, err = n.validateStatement(action, graph); err != nil {
			return nil, err
		}
	case action.TargetArgID != "":
		if va, err = n.validateTarget(action, va, graph); err != nil {
			return nil, err
		}
	}
	switch {
	case va.Direct():
		if err := n.fit(va, scheme, graph); err != nil {
			return nil, err
		}
	case action.AnswersCQID != "":
		if err := n.fitAnswer(va, scheme, graph); err != nil {
			return nil, err
		}
	}
	if err := n.validateScheme(va, scheme); err != nil {
		return nil, err
	}
	return va, nil
}

// OwnConclusion reports whether the argument concludes what its scheme
// concludes. An argument that answers a critical question concludes the
// answer: through the scheme's answer variable, which is its conclusion, or,
// for a scheme without one that the NASCQ links to the question as a backup,
// in place of the scheme's own conclusion (a consequences argument backing
// "is there a better means?" concludes that another action is better, not
// that it should be done). The scheme's questions about its conclusion are
// not about such an argument's conclusion.
func (n *NASCQ) OwnConclusion(graph GraphQuerier, argID, schemeID string) bool {
	s := n.Schemes[schemeID]
	if s == nil || s.Free || s.AnswerVariable != "" {
		return true
	}
	c := graph.ArgumentConclusion(argID)
	if c == "" {
		return true
	}
	_, err := graph.Fits(s.CarneadesID, c, nil)
	return err == nil
}

// validateStatement performs the checks of a move about a statement: the
// statement is on the board (S7), and an attack on it is a rebuttal, or an
// undermining when it is only a premise.
func (n *NASCQ) validateStatement(action ProposedAction, graph GraphQuerier) (*ValidatedAction, error) {
	if graph == nil {
		return nil, fmt.Errorf("S7: graph querier required for moves about a statement")
	}
	info := graph.Statement(action.TargetStmtID)
	if !info.Exists {
		return nil, fmt.Errorf("S7: statement %q is not on the board", action.TargetStmtID)
	}
	va := &ValidatedAction{Locus: Locus{Kind: LocusStatement}, LocusStmtID: action.TargetStmtID}
	if action.Type == "attack" {
		action.AttackType = "rebuttal"
		if info.Premise && !info.Concluded {
			action.AttackType = "undermining"
		}
		va.CaesAttackType = action.AttackType
	}
	va.ProposedAction = action
	return va, nil
}

// fit performs S9: a direct move, or a move about a statement, concludes the
// statement it is about (support) or its opposite (attack), whatever its
// scheme would conclude, so the scheme must be able to conclude that
// statement. The user-defined scheme and the schemes with an answer variable
// can conclude any statement: the answer variable is bound to it (an expert
// can assert anything). Any other scheme fits if its conclusion unifies with
// the statement; the unification fixes the values of its conclusion's
// variables (va.Fixed), which the mover need not give. Practical reasoning,
// which concludes should_be_performed(A), can thus argue for
// should_be_performed(carbon_tax), with A = carbon_tax, but not for
// possible(carbon_tax).
//
// An answer to a critical question is not checked here: the NASCQ's backup
// edges (S5) say which schemes can answer it.
func (n *NASCQ) fit(va *ValidatedAction, scheme *ASNode, graph GraphQuerier) error {
	if scheme.Free || scheme.AnswerVariable != "" {
		return nil
	}
	stmt, what := n.concluded(va, graph)
	if stmt == "" {
		return fmt.Errorf("S9: %s cannot conclude %s; argue it with Other, or with a scheme whose conclusion is any statement (%s)",
			scheme.Name, what, strings.Join(n.answerSchemes(), ", "))
	}
	fixed, err := graph.Fits(scheme.CarneadesID, stmt, va.VariableBindings)
	if err != nil {
		return fmt.Errorf("S9: %s cannot conclude %s: %v", scheme.Name, what, err)
	}
	va.Fixed = fixed
	return nil
}

// fitAnswer performs S9 for an answer to a critical question backed by a
// scheme without an answer variable. If the scheme's conclusion can be the
// answer (positive consequences answering with "A should be done"), the
// answer fixes its variables, as for a direct move, and the mover's values
// must agree: premises about one action may not conclude that another
// should be done. If it cannot (consequences backing "is there a better
// means?"), the NASCQ's backup edge (S5) stands for the link, and the
// argument concludes the answer in place of what the scheme would conclude.
func (n *NASCQ) fitAnswer(va *ValidatedAction, scheme *ASNode, graph GraphQuerier) error {
	if scheme.Free || scheme.AnswerVariable != "" {
		return nil
	}
	stmt, what := n.concluded(va, graph)
	if stmt == "" {
		return nil
	}
	if _, err := graph.Fits(scheme.CarneadesID, stmt, nil); err != nil {
		return nil // the answer is not of the scheme's form
	}
	fixed, err := graph.Fits(scheme.CarneadesID, stmt, va.VariableBindings)
	if err != nil {
		return fmt.Errorf("S9: as the answer, %s concludes %s: %v", scheme.Name, what, err)
	}
	va.Fixed = fixed
	return nil
}

// concluded returns the statement that the argument of a direct move, or of
// a move about a statement, concludes, and a description of it for messages;
// "" if no scheme without an answer variable can conclude it (an
// undercutter, or the opposite of a statement with several).
func (n *NASCQ) concluded(va *ValidatedAction, graph GraphQuerier) (string, string) {
	var stmt string
	switch va.Locus.Kind {
	case LocusStatement, LocusPremise:
		stmt = va.LocusStmtID
	case LocusConclusion:
		stmt = graph.ArgumentConclusion(va.TargetArgID)
	case LocusException: // an attack argues the exception, a defence its opposite
		if va.Type == "attack" {
			return va.LocusStmtID, va.LocusStmtID
		}
		return graph.Statement(va.LocusStmtID).Opposite, "the opposite of " + va.LocusStmtID
	default: // the inference: the undercutter
		return "", "that the inference of " + va.TargetArgID + " does not hold"
	}
	if va.Type == "support" {
		return stmt, stmt
	}
	return graph.Statement(stmt).Opposite, "the opposite of " + stmt
}

// answerSchemes returns the names of the schemes with an answer variable,
// which can conclude any statement.
func (n *NASCQ) answerSchemes() []string {
	var out []string
	for _, id := range sortedSchemeIDs(n.Schemes) {
		if s := n.Schemes[id]; s.AnswerVariable != "" && !s.Free {
			out = append(out, s.Name)
		}
	}
	return out
}

// validateScheme performs S2 and S6 for a move made with a scheme, and
// completes va.
func (n *NASCQ) validateScheme(va *ValidatedAction, scheme *ASNode) error {
	action := va.ProposedAction
	// S2: Variable completeness. On a move with a target, the scheme's answer
	// variable is bound to the answer, not by the mover, and so are the
	// variables S9 fixed.
	answerVar := ""
	if action.TargetArgID != "" || action.TargetStmtID != "" {
		answerVar = scheme.AnswerVariable
	}
	missing := []string{}
	if scheme.Free { // the user-defined scheme: a conclusion and at least one premise
		if len(FreePremises(action.VariableBindings)) == 0 {
			missing = append(missing, "P1")
		}
		if strings.TrimSpace(action.VariableBindings["C"]) == "" && answerVar == "" {
			missing = append(missing, "C")
		}
	}
	for varName := range scheme.Variables {
		if _, ok := action.VariableBindings[varName]; !ok && varName != answerVar && va.Fixed[varName] == "" && !scheme.Free {
			missing = append(missing, varName)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("S2: missing variable binding(s): %v", missing)
	}

	// S6: Template instantiation
	bindings := action.VariableBindings
	if len(va.Fixed) > 0 {
		bindings = make(map[string]string, len(action.VariableBindings)+len(va.Fixed))
		for k, v := range action.VariableBindings {
			bindings[k] = v
		}
		for k, v := range va.Fixed {
			bindings[k] = v
		}
	}
	va.InstPremises, va.InstConclusion = instantiateTemplates(scheme, bindings)
	if scheme.Free {
		va.InstPremises = FreePremises(action.VariableBindings)
	}
	va.CarneadesSchemeID = scheme.CarneadesID
	va.Weight = scheme.BaseWeight
	va.AnswerVariable = answerVar
	return nil
}

// validateTarget performs the checks of a move with a target (S7, S3, Q1,
// S5, S8) and completes va.
func (n *NASCQ) validateTarget(action ProposedAction, va *ValidatedAction, graph GraphQuerier) (*ValidatedAction, error) {
	// S7: Target argument exists
	if graph == nil {
		return nil, fmt.Errorf("S7: graph querier required for moves with a target")
	}
	if !graph.ArgumentExists(action.TargetArgID) {
		return nil, fmt.Errorf("S7: target argument %q not found in graph", action.TargetArgID)
	}

	// S3: the target was made with a NASCQ scheme, and the CQ, if any, is
	// asked of that scheme. The locus says what the move is about.
	caesSchemeID := graph.ArgumentSchemeID(action.TargetArgID)
	targetScheme := n.FindNASCQSchemeID(caesSchemeID)
	if targetScheme == "" {
		if _, ok := n.Schemes[caesSchemeID]; ok {
			targetScheme = caesSchemeID
		}
	}
	if targetScheme == "" {
		return nil, fmt.Errorf("S3: target %q was not made with a NASCQ scheme, so no critical question applies to it",
			action.TargetArgID)
	}
	var locus Locus
	switch {
	case action.AnswersCQID != "":
		if _, ok := n.Questions[action.AnswersCQID]; !ok {
			return nil, fmt.Errorf("S3: unknown CQ %q", action.AnswersCQID)
		}
		if askCQs := n.askIndex[targetScheme]; !containsStr(askCQs, action.AnswersCQID) {
			return nil, fmt.Errorf("S3: CQ %q is not applicable to scheme %q (ask CQs: %v)",
				action.AnswersCQID, targetScheme, askCQs)
		}
		if locus = n.LocusOf(action.AnswersCQID, targetScheme); locus.Kind == "" {
			return nil, fmt.Errorf("S3: CQ %q has no locus on scheme %q", action.AnswersCQID, targetScheme)
		}
		if locus.Kind == LocusConclusion && !n.OwnConclusion(graph, action.TargetArgID, targetScheme) {
			return nil, fmt.Errorf("S3: %q answers a critical question, so its conclusion is that answer, not what %s concludes: "+
				"CQ %q, about the scheme's conclusion, does not apply to it (argue for or against the statement instead)",
				action.TargetArgID, targetScheme, action.AnswersCQID)
		}
	case action.Type == "attack": // a direct attack names its type
		locus = Locus{Kind: map[string]string{"rebuttal": LocusConclusion, "undercutter": LocusArgument,
			"undermining": LocusPremise}[action.AttackType]}
	case action.TargetPremiseID != "": // a direct defence of a premise
		locus = Locus{Kind: LocusPremise}
	default: // a direct defence of the conclusion
		locus = Locus{Kind: LocusConclusion}
	}
	if action.Type == "attack" {
		at := locus.AttackType()
		if action.AttackType != "" && action.AttackType != at {
			return nil, fmt.Errorf("S3: CQ %q is answered by an attack of type %q, not %q",
				action.AnswersCQID, at, action.AttackType)
		}
		action.AttackType = at
		va.CaesAttackType = at
	} else {
		action.AttackType = ""
	}

	// Q1: a question alone shifts the burden of proof to the proponent, so
	// it can be about a premise (presumed until questioned) or the inference.
	// An exception, or a counterclaim to the conclusion, is the asker's to
	// prove: it is raised by an attack.
	if action.Type == "question" && locus.Kind != LocusPremise && locus.Kind != LocusArgument {
		return nil, fmt.Errorf("Q1: CQ %q is about the %s of %q: raise it with an argument (attack), not a question alone",
			action.AnswersCQID, locus.Kind, action.TargetArgID)
	}

	// S5: Backup edge validity: an explicit backup edge, or a
	// thesis-independent scheme (which backs every CQ).
	if action.AnswersCQID != "" && action.Type != "question" {
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

	// S8: the premise or exception the move is about. The mover may name a
	// premise; otherwise the locus pattern designates it, with the target's
	// own bindings.
	switch locus.Kind {
	case LocusPremise:
		if action.TargetPremiseID != "" {
			if !graph.PremiseExists(action.TargetArgID, action.TargetPremiseID) {
				return nil, fmt.Errorf("S8: premise %q not found in argument %q",
					action.TargetPremiseID, action.TargetArgID)
			}
			if locus.Pattern == "" && action.AnswersCQID != "" && n.CoveredPremises(graph, action.TargetArgID, targetScheme)[action.TargetPremiseID] {
				return nil, fmt.Errorf("S8: a critical question of its own is about premise %q of %q: ask that one",
					action.TargetPremiseID, action.TargetArgID)
			}
		} else if locus.Pattern == "" {
			return nil, fmt.Errorf("S8: name the premise of %q (target_premise_id)", action.TargetArgID)
		} else if action.TargetPremiseID = graph.PremiseMatching(action.TargetArgID, locus.Pattern); action.TargetPremiseID == "" {
			return nil, fmt.Errorf("S8: argument %q has no premise matching %q", action.TargetArgID, locus.Pattern)
		}
		va.LocusStmtID = action.TargetPremiseID
	case LocusException:
		action.TargetPremiseID = ""
		if va.LocusStmtID = graph.ExceptionMatching(action.TargetArgID, locus.Pattern); va.LocusStmtID == "" {
			return nil, fmt.Errorf("S8: argument %q has no exception matching %q", action.TargetArgID, locus.Pattern)
		}
	default:
		action.TargetPremiseID = ""
	}
	va.Locus = locus
	va.ProposedAction = action
	return va, nil
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
