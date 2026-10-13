package carneades

import (
	"fmt"
	"strings"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
)

// LocusStatement is the locus of a move about a statement rather than a part
// of an argument: an argument for or against any statement on the board.
const LocusStatement = "statement"

// Fits returns the values (terms) that the variables of the conclusion of
// the Carneades scheme schemeID take when an argument of the scheme
// concludes the statement stmtID: the unifier of its conclusion formula and
// the statement. Practical reasoning (should_be_performed(A)) fits
// should_be_performed(carbon_tax) with A = carbon_tax, and the opposite
// ¬should_be_performed(carbon_tax) only through a scheme that concludes
// ¬should_be_performed(A). It fails if the two do not unify, or if a value
// in given (the mover's bindings) differs from the one the statement fixes,
// as it is or as an atom (SanitizeBinding).
func (a *Adapter) Fits(schemeID, stmtID string, given map[string]string) (map[string]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	scheme := a.schemeMap[schemeID]
	if scheme == nil {
		return nil, fmt.Errorf("scheme %q is not in the theory", schemeID)
	}
	if len(scheme.Conclusions) != 1 {
		return nil, fmt.Errorf("scheme %s has %d conclusions", schemeID, len(scheme.Conclusions))
	}
	c, ok := terms.ReadString(scheme.Conclusions[0])
	if !ok {
		return nil, fmt.Errorf("the conclusion %q of scheme %s does not parse", scheme.Conclusions[0], schemeID)
	}
	s, ok := terms.ReadString(stmtID)
	if !ok {
		return nil, fmt.Errorf("%s is not a formula", stmtID)
	}
	env, ok := terms.Unify(c, s, nil)
	if !ok {
		return nil, fmt.Errorf("its conclusion has the form %s", scheme.Conclusions[0])
	}
	fixed := map[string]string{}
	for _, v := range scheme.Variables {
		t := terms.Substitute(terms.NewVariable(v), env)
		if tv, isVar := t.(terms.Variable); isVar && tv.Name == v {
			continue // not in the conclusion
		}
		fixed[v] = t.String()
		if g := strings.TrimSpace(given[v]); g != "" && g != fixed[v] && SanitizeBinding(g) != fixed[v] {
			return nil, fmt.Errorf("concluding %s makes %s %s, not %q (leave %s out: the statement fixes it)",
				stmtID, v, fixed[v], g, v)
		}
	}
	return fixed, nil
}

// StatementUse says whether the statement stmtID is in the graph, whether it
// is a premise of some argument, whether a move's argument (not a
// presumption, an exception or a question) concludes it, and what an
// argument against it concludes (see oppositeOf).
func (a *Adapter) StatementUse(stmtID string) (exists, premise, concluded bool, opposite string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.Graph.Statements[stmtID]
	if s == nil {
		return false, false, false, ""
	}
	for _, arg := range a.Graph.Arguments {
		if hasPremise(arg, s) {
			premise = true
			break
		}
	}
	for _, arg := range s.Args {
		if r := ArgumentRole(arg); r == RoleMain || r == "" {
			concluded = true
			break
		}
	}
	return true, premise, concluded, oppositeOf(s)
}

// oppositeOf returns the statement that an argument against s concludes,
// without creating it (opposingPosition does, when the move is played): the
// other position of its issue, or ¬s (s for ¬s). It returns "" for one of
// more than two positions of an issue.
func oppositeOf(s *caes.Statement) string {
	if is := s.Issue; is != nil {
		if len(is.Positions) != 2 {
			return ""
		}
		if is.Positions[0] == s {
			return is.Positions[1].Id
		}
		return is.Positions[0].Id
	}
	if rest := strings.TrimPrefix(s.Id, "¬"); rest != s.Id {
		return rest
	}
	return "¬" + s.Id
}

// joinOpposites makes a statement and its negation the two positions of one
// issue when an argument concludes the negation ¬p, or concludes p while ¬p
// exists: otherwise both could be In, as nothing would weigh one against the
// other (a scheme such as negative consequences concludes
// ¬should_be_performed(A) directly). The issue's proof standard is PE; the
// claim's is set by MakeIssue.
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) joinOpposites(s *caes.Statement) error {
	if s == nil || s.Issue != nil || s.IsUndercutter {
		return nil
	}
	if !strings.HasPrefix(s.Id, "¬") {
		if _, exists := a.Graph.Statements["¬"+s.Id]; !exists {
			return nil
		}
	}
	_, err := a.opposingPosition(s, caes.PE)
	return err
}
