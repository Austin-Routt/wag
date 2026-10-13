package carneades

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
)

// FreeSchemeID is the scheme ID of arguments of the user-defined scheme
// ("Other"), whose premises and conclusion are free text.
const FreeSchemeID = "other"

// FreeStatementID is the ID a free-text statement gets: the text itself if
// it already is the ID of a statement of the graph (so players can reuse a
// statement on the board), else a hash of the text, lowercased with its
// spaces collapsed and a final period dropped, so that the same sentence
// written twice is one statement. Free text is never parsed as a term: a
// capitalized word would be a Prolog variable.
func (a *Adapter) FreeStatementID(text string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.freeStatementID(text)
}

func (a *Adapter) freeStatementID(text string) string {
	t := strings.TrimSpace(text)
	if _, ok := a.Graph.Statements[t]; ok {
		return t
	}
	norm := strings.TrimSuffix(strings.ToLower(strings.Join(strings.Fields(t), " ")), ".")
	h := sha1.Sum([]byte(norm))
	return "text_" + hex.EncodeToString(h[:8])
}

// freeStatement returns the statement for a piece of free text, creating it
// with the text as its display text.
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) freeStatement(text string) *caes.Statement {
	id := a.freeStatementID(text)
	if s, ok := a.Graph.Statements[id]; ok {
		return s
	}
	s := &caes.Statement{Id: id, Text: strings.TrimSpace(text), Metadata: caes.NewMetadata()}
	a.Graph.Statements[id] = s
	return s
}

// AddFreeArgument adds an argument of the user-defined scheme: premises and
// conclusion in free text (or IDs of statements already on the board),
// owned by owner, with a constant weight. It returns the argument's ID.
func (a *Adapter) AddFreeArgument(owner string, premises []string, conclusion string, weight float64) ([]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if strings.TrimSpace(conclusion) == "" {
		return nil, fmt.Errorf("empty conclusion")
	}
	return a.addFreeArgument(owner, premises, a.freeStatement(conclusion), weight)
}

// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) addFreeArgument(owner string, premises []string, conclusion *caes.Statement, weight float64) ([]string, error) {
	var ps []caes.Premise
	for _, p := range premises {
		if strings.TrimSpace(p) == "" {
			continue
		}
		s := a.freeStatement(p)
		if s == conclusion {
			return nil, fmt.Errorf("the premise %q is the conclusion itself", p)
		}
		ps = append(ps, caes.Premise{Stmt: s})
	}
	if len(ps) == 0 {
		return nil, fmt.Errorf("an argument needs at least one premise")
	}
	id := a.nextArgID()
	arg := &caes.Argument{
		Id:         id,
		Metadata:   caes.Metadata{"owner": owner, RoleKey: RoleMain, "created_at": time.Now().UTC().Format(time.RFC3339)},
		Scheme:     &caes.Scheme{Id: FreeSchemeID, Weight: caes.ConstantWeighingFunction(weight)},
		Premises:   ps,
		Conclusion: conclusion,
	}
	a.Graph.Arguments[id] = arg
	conclusion.Args = append(conclusion.Args, arg)
	return []string{id}, nil
}

// SetArgumentMeta records a value in an argument's metadata (for example the
// player's own wording of it, or the sources they cite).
func (a *Adapter) SetArgumentMeta(argID, key string, value interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if arg := a.Graph.Arguments[argID]; arg != nil {
		if arg.Metadata == nil {
			arg.Metadata = caes.NewMetadata()
		}
		arg.Metadata[key] = value
	}
}

// Statement decisions: the dialectical status a moderator (an audience, a
// judge) gives a statement, after Gordon and Walton (2006): an accepted
// statement holds and a rejected one does not, whatever is argued; an
// undecided one is left to the arguments.
const (
	DecisionAccepted  = "accepted"
	DecisionRejected  = "rejected"
	DecisionUndecided = "undecided"
)

// Decide sets the moderator's decision on a statement. Accepting assumes it
// (Carneades labels an assumption In, and the other positions of its issue
// Out). Rejecting makes it an issue with its opposing position and assumes
// that position, so the statement is Out. Undecided withdraws both. It
// returns the statement's ID.
func (a *Adapter) Decide(stmtID, decision string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.Graph.Statements[stmtID]
	if s == nil {
		return "", fmt.Errorf("statement %q not found", stmtID)
	}
	var opp *caes.Statement
	if s.Issue != nil && len(s.Issue.Positions) == 2 {
		opp = s.Issue.Positions[0]
		if opp == s {
			opp = s.Issue.Positions[1]
		}
	}
	a.unassume(s)
	if opp != nil {
		a.unassume(opp)
	}
	switch decision {
	case DecisionAccepted:
		a.Graph.AddAssumption(terms.Normalize(s.Id))
	case DecisionRejected:
		if opp == nil {
			var err error
			if opp, err = a.opposingPosition(s, caes.PE); err != nil {
				return "", err
			}
		}
		a.Graph.AddAssumption(terms.Normalize(opp.Id))
	case DecisionUndecided:
	default:
		return "", fmt.Errorf("unknown decision %q (accepted, rejected, undecided)", decision)
	}
	return s.Id, nil
}

// Decision returns the moderator's decision on a statement: accepted if it
// is assumed, rejected if its opposing position is, else undecided.
func (a *Adapter) Decision(stmtID string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.Graph.Statements[stmtID]
	if s == nil {
		return DecisionUndecided
	}
	if a.assumed(s) {
		return DecisionAccepted
	}
	if s.Issue != nil {
		for _, p := range s.Issue.Positions {
			if p != s && a.assumed(p) {
				return DecisionRejected
			}
		}
	}
	return DecisionUndecided
}

// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) assumed(s *caes.Statement) bool {
	id := terms.Normalize(s.Id)
	for _, x := range a.Graph.Assumptions {
		if terms.Normalize(x) == id {
			return true
		}
	}
	return false
}

// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) unassume(s *caes.Statement) {
	id := terms.Normalize(s.Id)
	kept := a.Graph.Assumptions[:0]
	for _, x := range a.Graph.Assumptions {
		if terms.Normalize(x) != id {
			kept = append(kept, x)
		}
	}
	a.Graph.Assumptions = kept
}

// AddFreeArgumentOn adds an argument of the user-defined scheme for the
// statement stmtID or, if against, for its opposing position (a rebuttal),
// with premises in free text (or IDs of statements on the board). It returns
// the argument's ID.
func (a *Adapter) AddFreeArgumentOn(owner string, premises []string, stmtID string, against bool, weight float64) ([]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.Graph.Statements[stmtID]
	if s == nil {
		return nil, fmt.Errorf("statement %q not found", stmtID)
	}
	if against {
		var err error
		if s, err = a.opposingPosition(s, caes.PE); err != nil {
			return nil, err
		}
	}
	return a.addFreeArgument(owner, premises, s, weight)
}

// AddFreeStatement returns the ID of the statement for a piece of free text,
// creating it if needed (see FreeStatementID).
func (a *Adapter) AddFreeStatement(text string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.freeStatement(text).Id
}
