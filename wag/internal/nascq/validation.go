package nascq

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/yourorg/wag/internal/language"
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

	// ArgumentSchemeID returns the recorded NASCQ ID, or a legacy engine
	// scheme ID whose game mapping is unambiguous. Return "" for an absent
	// argument or an identity incompatible with the loaded game profile.
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
	// ExceptionOf names the arguments the statement is an exception of.
	ExceptionOf []string
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
// A shared Carneades identity is ambiguous and is a lint error. Its reverse
// entry is empty so that S3 cannot silently choose another scheme's CQs.
func (n *NASCQ) buildReverseSchemeIndex() {
	// Keep an empty entry after the second mapping, including when a third
	// alias is encountered. Merely deleting it would let the third alias win.
	idx := make(map[string]string, len(n.Schemes))
	for _, nascqID := range sortedSchemeIDs(n.Schemes) {
		node := n.Schemes[nascqID]
		if node.CarneadesID != "" {
			if _, exists := idx[node.CarneadesID]; !exists {
				idx[node.CarneadesID] = nascqID
			} else {
				idx[node.CarneadesID] = ""
			}
		}
	}
	n.reverseSchemeIndex = idx
}

// FindNASCQSchemeID returns the NASCQ scheme ID for a given Carneades scheme ID.
// Returns "" if no unambiguous mapping exists.
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

	// S1: an answer scheme answers a critical question, on its backup edges
	// (D094 (g)); it is not a move of its own.
	if scheme.Role == RoleAnswer && action.AnswersCQID == "" {
		return nil, fmt.Errorf("S1: %s is an answer scheme: it answers a critical question whose formula it concludes, not a move of its own (D094 (g))",
			scheme.Name)
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
// concludes. Every argument made since D075's rule (2026-11-22) does: an
// answer to a critical question concludes it through the scheme's answer
// variable or its own conclusion, never through a free-text bridge. An
// argument from an older game that a bridge let through concludes something
// else, and the scheme's questions about its conclusion are not about it.
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
	// A question that refuses Other (D095 (3)) refuses it however its
	// exception is argued: an argument made with Other that concludes the
	// exception, as a move about it (or against its negation), is refused as
	// an answer made with Other at the question would be (S5).
	if s := n.Schemes[action.SchemeID]; s != nil && s.Free {
		concl := action.TargetStmtID
		if action.Type == "attack" {
			concl = info.Opposite
		}
		if concl != "" {
			if cq := n.RefusingQuestion(concl, graph); cq != "" {
				return nil, fmt.Errorf("S5: %s is the exception CQ %q asks about, and that question refuses Other (D095 (3)): make it out with a scheme that backs the question (%s)",
					concl, cq, strings.Join(n.typedBackups(cq), ", "))
			}
		}
	}
	va.ProposedAction = action
	return va, nil
}

// RefusingQuestion returns the question that refuses Other (D095 (3)) and
// is asked at the exception stmt of an argument on the board, or "".
func (n *NASCQ) RefusingQuestion(stmt string, graph GraphQuerier) string {
	for _, argID := range graph.Statement(stmt).ExceptionOf {
		sid := graph.ArgumentSchemeID(argID)
		for _, cqID := range n.askIndex[sid] {
			cq := n.Questions[cqID]
			if cq == nil || !cq.RefuseOther {
				continue
			}
			if l := n.LocusOf(cqID, sid); l.Kind == LocusException && graph.ExceptionMatching(argID, l.Pattern) == stmt {
				return cqID
			}
		}
	}
	return ""
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
// An answer to a critical question is checked by fitAnswer.
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
// scheme without an answer variable: the answering argument concludes the
// answer, so the scheme's conclusion must be the answer, as for a direct
// move. Its conclusion fixes the values of its variables, and the mover's
// values must agree: positive consequences can answer "would doing A also
// have good consequences?" with "A should be done", about the same A, and
// premises about one action may not conclude that another should be done.
//
// A scheme whose conclusion cannot be the answer cannot back it (its
// premises would support its own conclusion, not the answer: a value
// argument's "V is a reason to keep G" does not show that another goal
// comes first). A free-text explanation on a backup edge does not supply
// that inference. A different conclusion requires a distinct, typed scheme
// whose premises actually establish it. The schemes with an answer
// variable conclude any answer: the answer is in their premises (the expert
// asserts the answer itself).
func (n *NASCQ) fitAnswer(va *ValidatedAction, scheme *ASNode, graph GraphQuerier) error {
	if scheme.Free || scheme.AnswerVariable != "" {
		return nil
	}
	stmt, what := n.concluded(va, graph)
	var err error
	if stmt == "" {
		err = fmt.Errorf("its conclusion is a statement, not an inference's failure")
	} else {
		_, err = graph.Fits(scheme.CarneadesID, stmt, nil)
	}
	if err != nil {
		return fmt.Errorf("S9: %s cannot conclude the answer, %s (%v); answer with Other, or with a scheme whose conclusion is any statement (%s)",
			scheme.Name, what, err, strings.Join(n.answerSchemes(), ", "))
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
	// variables S9 fixed. A value that names nothing (no letter or digit:
	// punctuation, invisible characters) binds nothing, as a blank one: an
	// answer whose H, A2 or I names nothing would make its exception out
	// without saying what outweighs, what does better, or what would settle
	// it (item F, G-unbound). Other's premises and conclusion are free text,
	// and the same holds of them: a premise "?" says nothing (the check of
	// step 8, W-4). Other's premises after the first may be left blank.
	answerVar := ""
	if action.TargetArgID != "" || action.TargetStmtID != "" {
		answerVar = scheme.AnswerVariable
	}
	missing, nameless := []string{}, []string{}
	check := func(v string) {
		switch x := action.VariableBindings[v]; {
		case strings.TrimSpace(x) == "":
			missing = append(missing, v)
		case language.NamesNothing(x):
			nameless = append(nameless, v)
		}
	}
	if scheme.Free { // the user-defined scheme: a conclusion and at least one premise
		if len(FreePremises(action.VariableBindings)) == 0 {
			missing = append(missing, "P1")
		}
		for i := 1; i <= OtherMaxPremises; i++ {
			if v := "P" + strconv.Itoa(i); strings.TrimSpace(action.VariableBindings[v]) != "" {
				check(v)
			}
		}
		if answerVar == "" {
			check("C")
		}
	} else {
		for varName := range scheme.Variables {
			if varName != answerVar && va.Fixed[varName] == "" {
				check(varName)
			}
		}
	}
	if len(missing) > 0 || len(nameless) > 0 {
		var why []string
		if len(missing) > 0 {
			sort.Strings(missing)
			why = append(why, fmt.Sprintf("missing variable binding(s): %v", missing))
		}
		if len(nameless) > 0 {
			sort.Strings(nameless)
			why = append(why, fmt.Sprintf("binding(s) that name nothing (no letter or digit): %v", nameless))
		}
		return fmt.Errorf("S2: %s", strings.Join(why, "; "))
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
	answered := ""
	if action.AnswersCQID != "" {
		answered = va.targetScheme
	}
	va.Weight = n.MoveWeight(action.SchemeID, answered)
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
	// The recorded NASCQ scheme first, then a legacy graph's engine scheme
	// (as in EnumerateActions).
	targetScheme := caesSchemeID
	if _, ok := n.Schemes[targetScheme]; !ok {
		targetScheme = n.FindNASCQSchemeID(caesSchemeID)
	}
	if targetScheme == "" {
		return nil, fmt.Errorf("S3: target %q was not made with a NASCQ scheme, so no critical question applies to it",
			action.TargetArgID)
	}
	va.targetScheme = targetScheme
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
	// thesis-independent scheme (which backs every CQ). A question that
	// refuses Other (D095 (3)) takes no answer made with it, before S9.
	if action.AnswersCQID != "" && action.Type != "question" {
		if cq := n.Questions[action.AnswersCQID]; cq != nil && cq.RefuseOther {
			if s := n.Schemes[action.SchemeID]; s != nil && s.Free {
				return nil, fmt.Errorf("S5: CQ %q refuses Other as an answer (D095 (3)): make the answer out with a scheme that backs it (%s)",
					action.AnswersCQID, strings.Join(n.typedBackups(action.AnswersCQID), ", "))
			}
		}
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

// MoveWeight is the weight of an argument made with a scheme (D094 (g)):
// for an answer scheme, the weight of the scheme of the argument it answers
// (answered: that argument's game scheme, "" when it answers none); for any
// other, WeightOf. A context that sets a scheme's weight sets the weight of
// the modal variants, bridges and answers that weigh as it.
func (n *NASCQ) MoveWeight(schemeID, answered string) float64 {
	s := n.Schemes[schemeID]
	if s == nil {
		return 0
	}
	if s.Role == RoleAnswer {
		if t := n.Schemes[answered]; t != nil {
			return n.WeightOf(t)
		}
	}
	return n.WeightOf(s)
}

// WeightOf is the weight of a scheme's arguments: its own base weight, or,
// for a modal variant or a bridge, the weight of the scheme it weighs as
// (WeighsAs, followed to its end; D094 (g)). Lint refuses a cycle and a
// name that is no scheme; such a chain stops at the last scheme it reaches.
func (n *NASCQ) WeightOf(s *ASNode) float64 {
	w := s.BaseWeight
	for seen := 0; s.WeighsAs != "" && seen < len(n.Schemes); seen++ {
		next := n.Schemes[s.WeighsAs]
		if next == nil {
			break
		}
		s, w = next, next.BaseWeight
	}
	return w
}

// typedBackups names the schemes that back a question other than Other:
// its backup edges' and the thesis-independent ones.
// BackupNames are the names of the schemes that back a question's answer,
// but the user-defined one: the schemes an answer to a question that
// refuses Other is made with (D095 (3)).
func (n *NASCQ) BackupNames(cqID string) []string { return n.typedBackups(cqID) }

func (n *NASCQ) typedBackups(cqID string) []string {
	var out []string
	for _, s := range n.GetBackupSchemes(cqID) {
		if !s.Free {
			out = append(out, s.Name)
		}
	}
	return out
}
