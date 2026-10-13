package carneades

import (
	"fmt"
	"strings"
	"time"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
)

// Locus kinds, as in package nascq: the part of an argument a critical
// question is about.
const (
	LocusPremise    = "premise"
	LocusException  = "exception"
	LocusConclusion = "conclusion"
	LocusArgument   = "argument"
)

// Answer is an answer to a critical question about an argument (or a direct
// attack or defence of one of its parts), backed by an instance of a scheme.
type Answer struct {
	Owner    string            // the player answering
	SchemeID string            // Carneades scheme of the answering argument
	Bindings map[string]string // values for the scheme's variables (sanitized here)
	Weight   float64           // constant weight of the answering argument

	TargetArgID string // the argument the question is about
	Locus       string // LocusPremise, LocusException, LocusConclusion or LocusArgument
	LocusStmtID string // the premise or exception statement, for those loci
	Positive    bool   // true: answer in the target's favour (defence); false: attack

	// AnswerVariable, if not empty, is the scheme variable that states its
	// conclusion; it is bound to the answer (whatever Bindings says), so the
	// premises are about the answer.
	AnswerVariable string

	Standard caes.Standard // proof standard of an issue the answer creates
}

// AnswerQuestion makes the argument that answers a critical question about an
// argument, and returns its ID and those of its exception arguments.
//
// What the answer concludes depends on the locus and the polarity:
//
//	locus        negative answer (attack)          positive answer (defence)
//	premise P    ¬P: undermines                    P: backs the premise
//	exception E  E: undercuts, through the         ¬E: E becomes an issue, decided
//	             exception argument                by weight under the proof standard
//	conclusion C ¬C: rebuts                        C: supports the conclusion
//	argument     the undercutter: undercuts        ¬undercutter: the undercutter
//	                                               becomes an issue
//
// (¬X stands for the opposing position of X; see opposingPosition.)
//
// The answering argument is an instance of the backup scheme (see
// AddSchemeArgument): Pyon and Ri (2022, p. 789) — "the argument consists of
// an answer to the question as the conclusion and support for the answer as
// the premises".
func (a *Adapter) AnswerQuestion(ans Answer) ([]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	target := a.Graph.Arguments[ans.TargetArgID]
	if target == nil || target.Conclusion == nil {
		return nil, fmt.Errorf("target argument %q not found", ans.TargetArgID)
	}
	var concl *caes.Statement
	var err error
	switch ans.Locus {
	case LocusPremise:
		p := a.Graph.Statements[ans.LocusStmtID]
		if p == nil || !hasPremise(target, p) {
			return nil, fmt.Errorf("%q is not a premise of %s", ans.LocusStmtID, target.Id)
		}
		concl = p
		if !ans.Positive {
			concl, err = a.opposingPosition(p, ans.Standard)
		}
	case LocusException:
		if strings.TrimSpace(ans.LocusStmtID) == "" {
			return nil, fmt.Errorf("no exception named for %s", target.Id)
		}
		e := a.ensureException(target, ans.LocusStmtID)
		concl = e
		if ans.Positive {
			concl, err = a.opposingPosition(e, ans.Standard)
		}
	case LocusConclusion:
		concl = target.Conclusion
		if !ans.Positive {
			concl, err = a.opposingPosition(target.Conclusion, ans.Standard)
		}
	case LocusArgument:
		u := a.ensureUndercutter(target)
		concl = u
		if ans.Positive {
			concl, err = a.opposingPosition(u, ans.Standard)
		}
	default:
		return nil, fmt.Errorf("unknown locus %q", ans.Locus)
	}
	if err != nil {
		return nil, err
	}
	bindings := ans.Bindings
	if ans.AnswerVariable != "" {
		bindings = make(map[string]string, len(ans.Bindings)+1)
		for k, v := range ans.Bindings {
			bindings[k] = v
		}
		bindings[ans.AnswerVariable] = StatementTerm(concl.Id)
	}
	return a.addSchemeArgument(ans.Owner, ans.SchemeID, bindings, ans.AnswerVariable, concl, ans.Weight)
}

// StatementTerm returns a statement ID as a term that can be the value of a
// scheme variable: the ID itself, with the opposing position ¬X written
// not(X) (¬ is not term syntax).
func StatementTerm(id string) string {
	if rest := strings.TrimPrefix(id, "¬"); rest != id {
		return "not(" + StatementTerm(rest) + ")"
	}
	return id
}

// AddSchemeArgument creates an argument of the Carneades scheme schemeID for
// the statement conclusionID (created if needed), owned by owner, with a
// constant weight. Its premises are the scheme's premises and assumptions
// with the bindings substituted, as statements of the graph (their text
// comes from the Language templates). If the scheme has exceptions, the
// argument gets its own undercutter and one exception argument per exception,
// as InstantiateScheme gives a support move's argument. The argument records
// its parameters, so the critical questions asked of it can name its
// premises and exceptions (PremiseMatching, ExceptionMatching).
func (a *Adapter) AddSchemeArgument(owner, schemeID string, bindings map[string]string, conclusionID string, weight float64) ([]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if strings.TrimSpace(conclusionID) == "" {
		return nil, fmt.Errorf("empty conclusion")
	}
	return a.addSchemeArgument(owner, schemeID, bindings, "", a.findOrCreateStatement(conclusionID), weight)
}

// addSchemeArgument is AddSchemeArgument with the conclusion statement given.
// The value of the variable termVar is a term (see StatementTerm), used as it
// is; the other values are sanitized into atoms.
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) addSchemeArgument(owner, schemeID string, bindings map[string]string, termVar string, conclusion *caes.Statement, weight float64) ([]string, error) {
	scheme := a.schemeMap[schemeID]
	if scheme == nil {
		return nil, fmt.Errorf("scheme %q not found in theory", schemeID)
	}
	params := make([]string, len(scheme.Variables))
	var env terms.Bindings
	for i, v := range scheme.Variables {
		val, ok := bindings[v]
		if !ok {
			return nil, fmt.Errorf("missing binding for variable %s of scheme %s", v, schemeID)
		}
		params[i] = val
		if v != termVar {
			params[i] = SanitizeBinding(val)
		}
		t, ok := terms.ReadString(params[i])
		if !ok {
			return nil, fmt.Errorf("binding %q of %s does not parse", params[i], v)
		}
		env = terms.AddBinding(terms.NewVariable(v), t, env)
	}
	instantiate := func(formula string) (*caes.Statement, error) {
		t, ok := terms.ReadString(formula)
		if !ok {
			return nil, fmt.Errorf("formula %q of scheme %s does not parse", formula, schemeID)
		}
		return a.findOrCreateStatement(terms.Substitute(t, env).String()), nil
	}

	var premises []caes.Premise
	for i, f := range append(append([]string{}, scheme.Premises...), scheme.Assumptions...) {
		s, err := instantiate(f)
		if err != nil {
			return nil, err
		}
		role := ""
		if i < len(scheme.Roles) {
			role = scheme.Roles[i]
		}
		premises = append(premises, caes.Premise{Stmt: s, Role: role})
	}
	now := time.Now().UTC().Format(time.RFC3339)
	id := a.nextArgID()
	arg := &caes.Argument{
		Id:         id,
		Metadata:   caes.Metadata{"owner": owner, RoleKey: RoleMain, "created_at": now},
		Scheme:     a.makeWeightedScheme(schemeID, weight),
		Parameters: params,
		Premises:   premises,
		Conclusion: conclusion,
	}
	a.Graph.Arguments[id] = arg
	conclusion.Args = append(conclusion.Args, arg)
	created := []string{id}

	if len(scheme.Exceptions) > 0 {
		a.ensureUndercutter(arg)
		for _, f := range scheme.Exceptions {
			e, err := instantiate(f)
			if err != nil {
				return created, err
			}
			created = append(created, a.addExceptionArgument(arg, e, owner).Id)
		}
	}
	return created, nil
}

// ensureUndercutter returns the argument's undercutter statement, creating
// undercut(<id>) if it has none.
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) ensureUndercutter(arg *caes.Argument) *caes.Statement {
	if arg.Undercutter != nil {
		return arg.Undercutter
	}
	u := a.findOrCreateStatement(fmt.Sprintf("undercut(%s)", arg.Id))
	u.IsUndercutter = true
	arg.Undercutter = u
	return u
}

// ensureException returns the exception statement exceptionID of the
// argument, making sure an exception argument connects it to the argument's
// undercutter (so that the exception, once In, undercuts the argument).
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) ensureException(arg *caes.Argument, exceptionID string) *caes.Statement {
	e := a.findOrCreateStatement(exceptionID)
	u := a.ensureUndercutter(arg)
	for _, x := range u.Args {
		if ArgumentRole(x) == RoleException && len(x.Premises) == 1 && x.Premises[0].Stmt == e {
			return e
		}
	}
	a.addExceptionArgument(arg, e, "")
	return e
}

// addExceptionArgument adds an argument from the exception e to the
// undercutter of arg: structure of the scheme, not a move, so it has no
// owner (instantiatedBy records who made the argument it belongs to). It has
// no scheme, so Carneades weighs it as linked: 1 when e is In.
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) addExceptionArgument(arg *caes.Argument, e *caes.Statement, instantiatedBy string) *caes.Argument {
	u := a.ensureUndercutter(arg)
	x := &caes.Argument{
		Id:         a.nextArgID(),
		Metadata:   caes.Metadata{RoleKey: RoleException, "instantiated_by": instantiatedBy, "created_at": time.Now().UTC().Format(time.RFC3339)},
		Premises:   []caes.Premise{{Stmt: e}},
		Conclusion: u,
	}
	a.Graph.Arguments[x.Id] = x
	u.Args = append(u.Args, x)
	return x
}

func hasPremise(arg *caes.Argument, s *caes.Statement) bool {
	for _, p := range arg.Premises {
		if p.Stmt == s {
			return true
		}
	}
	return false
}

// ExceptionMatching returns the ID of the exception of argument argID that a
// pattern in NASCQ notation, such as "side_effects({A},{S1})", designates:
// the pattern with the argument's own bindings substituted. An argument
// without such bindings matches when exactly one of the premises of its
// exception arguments has the pattern's predicate and arity. Returns "" if
// neither works. The exception statement need not exist yet; an attack on
// it creates it (AnswerQuestion).
func (a *Adapter) ExceptionMatching(argID, pattern string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	arg := a.Graph.Arguments[argID]
	if arg == nil || strings.TrimSpace(pattern) == "" {
		return ""
	}
	inst := pattern
	for v, val := range argumentBindings(arg) {
		inst = strings.ReplaceAll(inst, "{"+v+"}", val)
	}
	if !strings.Contains(inst, "{") {
		if t, ok := terms.ReadString(inst); ok {
			return t.String()
		}
	}
	functor, arity, ok := patternPredicate(pattern)
	if !ok || arg.Undercutter == nil {
		return ""
	}
	match := ""
	for _, x := range arg.Undercutter.Args {
		if ArgumentRole(x) != RoleException || len(x.Premises) != 1 {
			continue
		}
		t, ok := terms.ReadString(x.Premises[0].Stmt.Id)
		if !ok {
			continue
		}
		if f, ok := terms.Functor(t); ok && f == functor && terms.Arity(t) == arity {
			if match != "" && match != x.Premises[0].Stmt.Id {
				return ""
			}
			match = x.Premises[0].Stmt.Id
		}
	}
	return match
}

// QuestionSchemeID is the scheme ID of question arguments (RaiseQuestion). It
// maps to no NASCQ scheme, so question arguments are never targets.
const QuestionSchemeID = "wag_question"

// RaiseQuestion asks a critical question about a premise of an argument, or
// about its inference, without arguing for an answer, and returns the ID of
// the argument that stands for the question.
//
// Asking shifts the burden of proof to the argument's proponent, as asking
// about an assumption does in Carneades (Gordon, Prakken and Walton 2007): the
// question is a premise-less argument for the negative answer, owned by the
// asker, with the weight of a presumption. For a premise P it argues ¬P, so
// P (presumed with the same weight) is tied and Out until the proponent backs
// it; for the inference it argues the undercutter, so the argument is out of
// force until the proponent answers (¬undercut, which any scheme outweighs).
// An exception is not asked about this way: the one who raises it bears its
// burden, so it is attacked with an argument.
func (a *Adapter) RaiseQuestion(asker, targetArgID, locus, locusStmtID string, weight float64, standard caes.Standard) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	target := a.Graph.Arguments[targetArgID]
	if target == nil || target.Conclusion == nil {
		return "", fmt.Errorf("target argument %q not found", targetArgID)
	}
	var concl *caes.Statement
	switch locus {
	case LocusPremise:
		p := a.Graph.Statements[locusStmtID]
		if p == nil || !hasPremise(target, p) {
			return "", fmt.Errorf("%q is not a premise of %s", locusStmtID, target.Id)
		}
		var err error
		if concl, err = a.opposingPosition(p, standard); err != nil {
			return "", err
		}
	case LocusArgument:
		concl = a.ensureUndercutter(target)
	default:
		return "", fmt.Errorf("a question about the %s is asked with an argument (attack), not alone", locus)
	}
	id := a.nextArgID()
	arg := &caes.Argument{
		Id:         id,
		Metadata:   caes.Metadata{"owner": asker, RoleKey: RoleQuestion, "target": targetArgID, "created_at": time.Now().UTC().Format(time.RFC3339)},
		Scheme:     &caes.Scheme{Id: QuestionSchemeID, Weight: caes.ConstantWeighingFunction(weight)},
		Conclusion: concl,
	}
	a.Graph.Arguments[id] = arg
	concl.Args = append(concl.Args, arg)
	return id, nil
}

// MakeIssue makes the statement (created if needed) and its opposing position
// ¬statement the positions of an issue with the given proof standard, unless
// it is already a position of one, and returns the statement's ID. A
// debate's central claim is at issue from the start, so the claim standard
// applies to it even before anyone rebuts it.
func (a *Adapter) MakeIssue(stmtID string, standard caes.Standard) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.findOrCreateStatement(stmtID)
	if s.Issue == nil {
		if _, err := a.opposingPosition(s, standard); err != nil {
			return "", err
		}
	}
	return s.Id, nil
}
