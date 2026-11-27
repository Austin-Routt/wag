package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/evidence"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// Evidence in the game (docs/e1-evidence.md): a record set given with an
// inquiry, offers, consents and withdrawals as moves, and moves that cite
// the bindings admitted for consideration. Admission is for consideration
// only (D076): it asserts nothing, and sets no label, weight or assumption.

// The evidence acts.
const (
	ActOffer    = "evidence_offer"
	ActConsent  = "evidence_consent"
	ActWithdraw = "evidence_withdraw"
)

// IsEvidenceAct: a move of this type is an evidence act.
func IsEvidenceAct(t string) bool { return t == ActOffer || t == ActConsent || t == ActWithdraw }

// WithoutEvidence is the action without the fields of an evidence act and
// of a move's citations (offer, binding, revision, evidence_refs): what a
// game without a record set reads of it, as before E1 (docs/e1-evidence.md,
// E1.7, B3).
func WithoutEvidence(a nascq.ProposedAction) nascq.ProposedAction {
	a.Offer, a.BindingID, a.Revision, a.EvidenceRefs = nil, "", "", nil
	return a
}

// ArgumentRevisionAlgorithm names how ArgumentRevision is computed.
const ArgumentRevisionAlgorithm = "wag-argument-revision/1"

// Evidence is a game's evidence: its record set, with the context and the
// actors the game writes, and the targets, bindings and events its
// players' acts made (evidence.GameBundle).
type Evidence struct {
	Name   string          // the record set's name in the evidence folder
	SHA256 string          // the SHA-256 of its records.json, as the game took it
	Bundle evidence.Bundle // the game's bundle (wag-evidence/0.2)
	report evidence.Report // the bundle's last check: revisions, consents, admission
}

// EvidenceMark marks an accepted evidence act in the history: what it made
// or named, and when, so that the board reads it as accepted and a replay
// makes it again with the same records and checks the revisions.
type EvidenceMark struct {
	Act             string `json:"act"`
	BindingID       string `json:"binding_id"`
	BindingRevision string `json:"binding_revision"`
	// ArgumentRevision is an offer's: the revision of the argument its
	// target is in (ArgumentRevision).
	ArgumentRevision string `json:"argument_revision,omitempty"`
	// Time is when the act was made: the time of the records it made, which
	// a replay gives them again.
	Time string `json:"time"`
}

// NewEvidence is the evidence of a game of the two players given, with the
// record set set (as evidence.DecodeRecordSet gives it), whose records.json
// is named name and has the SHA-256 sha, and the game's pins. Its records
// are checked (evidence.GameBundle); the caller checks the payloads.
func NewEvidence(name, sha string, set *evidence.Bundle, players []evidence.Participant, engine, catalogue, nascqPin string) (*Evidence, error) {
	b, err := evidence.GameBundle(set, players, engine, catalogue, nascqPin)
	if err != nil {
		return nil, err
	}
	ev := &Evidence{Name: name, SHA256: sha, Bundle: *b}
	ev.report = evidence.ValidateRecords(&ev.Bundle)
	return ev, nil
}

// Admitted: the binding is admitted for consideration (both players'
// consents to its revision stand).
func (ev *Evidence) Admitted(bindingID string) bool {
	for _, id := range ev.report.AdmittedBindings {
		if id == bindingID {
			return true
		}
	}
	return false
}

// Revision is the binding's revision, if the game has the binding.
func (ev *Evidence) Revision(bindingID string) (string, bool) {
	r, ok := ev.report.BindingRevisions[bindingID]
	return r, ok
}

// Consents are the players whose consent to the binding stands.
func (ev *Evidence) Consents(bindingID string) []string {
	var out []string
	for _, a := range ev.report.Consents[bindingID] {
		out = append(out, strings.TrimPrefix(a, evidence.PlayerActor("")))
	}
	return out
}

// Binding is the game's binding of that id.
func (ev *Evidence) Binding(id string) (evidence.Binding, bool) {
	for _, b := range ev.Bundle.Bindings {
		if b.ID == id {
			return b, true
		}
	}
	return evidence.Binding{}, false
}

// Target is the game's target of that id.
func (ev *Evidence) Target(id string) (evidence.Target, bool) {
	for _, t := range ev.Bundle.Targets {
		if t.ID == id {
			return t, true
		}
	}
	return evidence.Target{}, false
}

// clone copies the evidence for a snapshot: the slices an act appends to are
// copied, so that undoing the act undoes its records. An act replaces the
// report, never changes it in place.
func (ev *Evidence) clone() *Evidence {
	if ev == nil {
		return nil
	}
	c := *ev
	c.Bundle.Targets = append([]evidence.Target{}, ev.Bundle.Targets...)
	c.Bundle.Bindings = append([]evidence.Binding{}, ev.Bundle.Bindings...)
	c.Bundle.Events = append([]evidence.Event{}, ev.Bundle.Events...)
	return &c
}

// offerable is the move that made the argument, if an offer may target it:
// a main argument of a support or attack move (not a question's, a
// presumption or an exception argument).
func (gs *GameState) offerable(argID string) (*Move, *caes.Argument, bool) {
	arg := gs.Adapter.Graph.Arguments[argID]
	if arg == nil || carneades.ArgumentRole(arg) != carneades.RoleMain {
		return nil, nil, false
	}
	for i := range gs.MoveHistory {
		m := &gs.MoveHistory[i]
		if m.Rejection != "" || (m.Action.Type != "support" && m.Action.Type != "attack") {
			continue
		}
		for _, id := range m.ArgIDs {
			if id == argID {
				return m, arg, true
			}
		}
	}
	return nil, nil, false
}

// ArgumentRevision is the revision of an argument an offer may target
// (ArgumentRevisionAlgorithm): the SHA-256 of the JSON encoding of the
// algorithm's name, the NASCQ scheme of the move that made it, its premises
// (statement ids, in order), its conclusion, its exceptions (sorted) and the
// move's weight: the weight assigned, not the evaluated weight, which an
// undercutter sets to 0. An argument does not change once made, so neither
// does its revision. ok is false for an argument no offer may target.
func (gs *GameState) ArgumentRevision(argID string) (rev string, ok bool) {
	m, arg, ok := gs.offerable(argID)
	if !ok {
		return "", false
	}
	premises := []string{}
	for _, p := range arg.Premises {
		if p.Stmt != nil {
			premises = append(premises, p.Stmt.Id)
		}
	}
	conclusion := ""
	if arg.Conclusion != nil {
		conclusion = arg.Conclusion.Id
	}
	value := struct {
		Algorithm  string
		Scheme     string
		Premises   []string
		Conclusion string
		Exceptions []string
		Weight     float64
	}{ArgumentRevisionAlgorithm, m.Action.SchemeID, premises, conclusion, gs.exceptionsOf(arg), m.Weight}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", false
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), true
}

// exceptionsOf are the argument's exceptions: the premises of the exception
// arguments for its undercutter, sorted.
func (gs *GameState) exceptionsOf(arg *caes.Argument) []string {
	out := []string{}
	if arg.Undercutter == nil {
		return out
	}
	for _, x := range gs.Adapter.Graph.Arguments {
		if x.Conclusion == arg.Undercutter && carneades.ArgumentRole(x) == carneades.RoleException && len(x.Premises) == 1 && x.Premises[0].Stmt != nil {
			out = append(out, x.Premises[0].Stmt.Id)
		}
	}
	sort.Strings(out)
	return out
}

// statementsOf are the argument's statements of a kind: its premises (its
// scheme's premises), its assumptions (the premises after them), its
// exceptions, its conclusion, or its inference: the argument's undercutter
// statement, undercut(A), which an undercutter of it concludes.
func (gs *GameState) statementsOf(arg *caes.Argument, kind string) []string {
	// An Other argument's scheme lists no premises: all its premises are
	// premises, and it has no assumptions (docs/e1-evidence.md, E1.7, A1).
	n := len(arg.Premises)
	if arg.Scheme != nil && arg.Scheme.Id != carneades.FreeSchemeID && len(arg.Scheme.Premises) <= n {
		n = len(arg.Scheme.Premises)
	}
	var out []string
	switch kind {
	case "premise", "assumption":
		ps := arg.Premises[:n]
		if kind == "assumption" {
			ps = arg.Premises[n:]
		}
		for _, p := range ps {
			if p.Stmt != nil {
				out = append(out, p.Stmt.Id)
			}
		}
	case "exception":
		out = gs.exceptionsOf(arg)
	case "conclusion":
		if arg.Conclusion != nil {
			out = []string{arg.Conclusion.Id}
		}
	case "inference":
		// An argument whose scheme has no exceptions, or an Other argument,
		// has no undercutter statement until it is undercut or asked about:
		// its inference is undercut(A) all the same (E1.7, A2).
		out = []string{carneades.UndercutterID(arg.Id)}
		if arg.Undercutter != nil {
			out = []string{arg.Undercutter.Id}
		}
	}
	return out
}

// OfferStatements are the statements of a kind (premise, assumption,
// exception, conclusion, inference) of an argument an offer may target, as
// an offer names them; nil for an argument no offer may target.
func (gs *GameState) OfferStatements(argID, kind string) []string {
	if _, arg, ok := gs.offerable(argID); ok {
		return gs.statementsOf(arg, kind)
	}
	return nil
}

// ReplayActAt gives the next evidence act the time it was made at (its
// mark's), which its records take instead of the clock's: a replay makes
// the same records, dated as when the act was played.
func (gs *GameState) ReplayActAt(t string) { gs.actTime = t }

// actAt is the time of the evidence act being played: the clock's, or the
// latest time of the game's bindings and events if that is later. The
// contract refuses an event dated before the one before it, or before its
// binding, and a clock behind the last act (a resume on another computer, a
// clock set back) must not refuse a valid act (E1.7, A3).
func (gs *GameState) actAt() string {
	if t := gs.actTime; t != "" {
		gs.actTime = ""
		return t
	}
	at := gs.now().UTC()
	if ev := gs.Evidence; ev != nil {
		var times []string
		for _, b := range ev.Bundle.Bindings {
			times = append(times, b.CreatedAt)
		}
		for _, e := range ev.Bundle.Events {
			times = append(times, e.CreatedAt)
		}
		for _, s := range times {
			if t, err := time.Parse(time.RFC3339Nano, s); err == nil && t.After(at) {
				at = t.UTC()
			}
		}
	}
	return at.Format(time.RFC3339Nano)
}

// evidenceAct plays an evidence act: it checks the act against the record
// set and the live graph, writes the records it makes (their ids from the
// act's place in the history), checks the bundle with the contract's rules
// (evidence.ValidateRecords), and records the act in the history with its
// mark. A refused act is refused as any move is, and leaves no record.
func (gs *GameState) evidenceAct(player *Player, a nascq.ProposedAction) *TurnResult {
	ev := gs.Evidence
	at := gs.actAt()
	if ev == nil {
		return gs.rejectInvalid(player, a, "EV: this game has no record set: it takes no evidence act")
	}
	if len(a.EvidenceRefs) > 0 || len(nonEmpty(a.Evidence)) > 0 {
		return gs.rejectInvalid(player, a, "EV: an evidence act cites nothing: it argues nothing")
	}
	i := len(gs.MoveHistory)
	actor := evidence.PlayerActor(player.ID)
	b := ev.clone().Bundle
	mark := &EvidenceMark{Act: a.Type, Time: at}
	event := func(typ, bindingID, revision string) evidence.Event {
		return evidence.Event{Record: evidence.Record{ID: fmt.Sprintf("%se%d", evidence.GamePrefix, i), CreatedBy: actor, CreatedAt: at},
			Sequence: len(b.Events) + 1, Type: typ, BindingID: bindingID, BindingRevision: revision, Purpose: "consideration"}
	}
	switch a.Type {
	case ActOffer:
		o := a.Offer
		if o == nil || a.BindingID != "" || a.Revision != "" {
			return gs.rejectInvalid(player, a, "EV: an offer gives an offer (a claim, an argument, a kind, a role, an interpretation and a warrant), and no binding or revision")
		}
		found := false
		for _, c := range b.Claims {
			found = found || c.ID == o.ClaimID
		}
		if !found {
			return gs.rejectInvalid(player, a, fmt.Sprintf("EV: the record set has no claim %q", o.ClaimID))
		}
		argRev, ok := gs.ArgumentRevision(o.ArgumentID)
		if !ok {
			return gs.rejectInvalid(player, a, fmt.Sprintf("EV: %q is not an argument a support or attack move put on the board: an offer targets one (not a question's, a presumption or an exception argument)", o.ArgumentID))
		}
		switch o.Kind {
		case "premise", "assumption", "exception", "conclusion", "inference":
		default:
			return gs.rejectInvalid(player, a, fmt.Sprintf("EV: an offer targets an argument's premise, assumption, exception, conclusion or inference, not %q", o.Kind))
		}
		stmts := gs.statementsOf(gs.Adapter.Graph.Arguments[o.ArgumentID], o.Kind)
		stmt := o.Statement
		if stmt == "" && (o.Kind == "conclusion" || o.Kind == "inference") && len(stmts) == 1 {
			stmt = stmts[0]
		}
		found = false
		for _, s := range stmts {
			found = found || s == stmt
		}
		if !found {
			return gs.rejectInvalid(player, a, fmt.Sprintf("EV: %s has no %s %q (its %ss: %s)", o.ArgumentID, o.Kind, stmt, o.Kind, strings.Join(stmts, ", ")))
		}
		switch o.Role {
		case "support", "oppose", "context", "definition":
		default:
			return gs.rejectInvalid(player, a, fmt.Sprintf("EV: an offer's role is support, oppose, context or definition, not %q", o.Role))
		}
		if strings.TrimSpace(o.Interpretation) == "" || strings.TrimSpace(o.Warrant) == "" {
			return gs.rejectInvalid(player, a, "EV: an offer gives its interpretation and its warrant")
		}
		for _, x := range b.Bindings {
			t, _ := ev.Target(x.TargetID)
			if x.ClaimID == o.ClaimID && t.ArgumentID == o.ArgumentID && t.Kind == o.Kind && t.Proposition == stmt &&
				x.Role == o.Role && x.Interpretation == o.Interpretation && x.Warrant == o.Warrant {
				return gs.rejectInvalid(player, a, fmt.Sprintf("EV: this offer is already made: the binding %s", x.ID))
			}
		}
		tid, bid := fmt.Sprintf("%st%d", evidence.GamePrefix, i), fmt.Sprintf("%sb%d", evidence.GamePrefix, i)
		b.Targets = append(b.Targets, evidence.Target{Record: evidence.Record{ID: tid, CreatedBy: evidence.GameActor, CreatedAt: at},
			ArgumentID: o.ArgumentID, ArgumentRevision: argRev, Kind: o.Kind, Proposition: stmt})
		b.Bindings = append(b.Bindings, evidence.Binding{Record: evidence.Record{ID: bid, CreatedBy: actor, CreatedAt: at},
			ClaimID: o.ClaimID, TargetID: tid, Role: o.Role, Interpretation: o.Interpretation, Warrant: o.Warrant})
		r := evidence.ValidateRecords(&b)
		if !r.Valid {
			return gs.rejectInvalid(player, a, "EV: "+evidence.FirstIssue(r.Errors))
		}
		rev := r.BindingRevisions[bid]
		b.Events = append(b.Events, event("consent", bid, rev)) // an offer is its player's consent
		mark.BindingID, mark.BindingRevision, mark.ArgumentRevision = bid, rev, argRev
	case ActConsent, ActWithdraw:
		if a.Offer != nil {
			return gs.rejectInvalid(player, a, "EV: a consent or a withdrawal names a binding and its revision, and offers nothing")
		}
		x, ok := ev.Binding(a.BindingID)
		if !ok {
			return gs.rejectInvalid(player, a, fmt.Sprintf("EV: the game has no binding %q", a.BindingID))
		}
		current, _ := ev.Revision(x.ID)
		rev := a.Revision
		if len(rev) == 12 && strings.HasPrefix(current, rev) {
			rev = current // given by its first 12 digits
		}
		if rev != current {
			return gs.rejectInvalid(player, a, fmt.Sprintf("EV: %q is not the revision of the binding %s, %s: a consent is to the exact revision (in full, or its first 12 digits), and one for another revision is stale", a.Revision, x.ID, current))
		}
		has := false
		for _, p := range ev.Consents(x.ID) {
			has = has || p == player.ID
		}
		typ := "consent"
		if a.Type == ActWithdraw {
			typ = "withdraw"
			if !has {
				return gs.rejectInvalid(player, a, fmt.Sprintf("EV: %s has no consent to the binding %s to withdraw", player.ID, x.ID))
			}
		} else if has {
			return gs.rejectInvalid(player, a, fmt.Sprintf("EV: %s's consent to the binding %s already stands", player.ID, x.ID))
		}
		b.Events = append(b.Events, event(typ, x.ID, rev))
		mark.BindingID, mark.BindingRevision = x.ID, rev
	default:
		return gs.rejectInvalid(player, a, fmt.Sprintf("EV: unknown evidence act %q", a.Type))
	}
	r := evidence.ValidateRecords(&b)
	if !r.Valid {
		return gs.rejectInvalid(player, a, "EV: "+evidence.FirstIssue(r.Errors))
	}
	ev.Bundle, ev.report = b, r
	gs.MoveHistory = append(gs.MoveHistory, Move{TurnNumber: gs.CurrentTurn, PlayerID: player.ID, Action: a, Evidence: mark})
	player.ConsecutiveAbstentions = 0
	gs.roundHadMove = true // an act is a move, not a pass
	gs.endTurn(player)
	return gs.ended(&TurnResult{Accepted: true})
}

// evidenceRefsRefusal tells why the bindings a move cites do not ground it,
// or "" if they do (docs/e1-evidence.md, choice 7). Each is a binding of
// the game, admitted for consideration when the move is made, of role
// support or oppose, and agrees with what the move's argument concludes, C:
// a support binding's target is C, or, at an argument's inference, the
// statement C is the opposite of (C defends the inference); an oppose
// binding's target is the statement C is the opposite of, or, at an
// argument's inference, C itself (an undercutter concludes undercut(A)).
// Context and definition bindings ground no move; a question cites none.
func (gs *GameState) evidenceRefsRefusal(a nascq.ProposedAction, v *nascq.ValidatedAction, argIDs []string) string {
	if len(a.EvidenceRefs) == 0 {
		return ""
	}
	ev := gs.Evidence
	if ev == nil {
		return "EV: this game has no record set: a move cites no binding"
	}
	if v.Type == "question" || len(argIDs) == 0 {
		return "EV: a question cites no binding: it argues nothing"
	}
	arg := gs.Adapter.Graph.Arguments[argIDs[0]]
	if arg == nil || arg.Conclusion == nil {
		return "EV: the move's argument concludes nothing a binding can ground"
	}
	c := arg.Conclusion.Id
	seen := map[string]bool{}
	for _, id := range a.EvidenceRefs {
		if seen[id] {
			return fmt.Sprintf("EV: the binding %s is cited twice", id)
		}
		seen[id] = true
		b, ok := ev.Binding(id)
		if !ok {
			return fmt.Sprintf("EV: the game has no binding %q (a source's reference is cited among the move's evidence strings)", id)
		}
		if !ev.Admitted(id) {
			return fmt.Sprintf("EV: the binding %s is not admitted for consideration: both players consent to its revision first", id)
		}
		t, _ := ev.Target(b.TargetID)
		_, _, _, opposite := gs.Adapter.StatementUse(t.Proposition)
		switch {
		case b.Role == "context" || b.Role == "definition":
			return fmt.Sprintf("EV: the binding %s is a %s binding, which grounds no move", id, b.Role)
		case b.Role == "support" && t.Kind != "inference" && t.Proposition == c:
		case b.Role == "support" && t.Kind == "inference" && opposite == c:
		case b.Role == "oppose" && t.Kind != "inference" && opposite == c:
		case b.Role == "oppose" && t.Kind == "inference" && t.Proposition == c:
		default:
			return fmt.Sprintf("EV: the binding %s (%s, at %s's %s %s) does not agree with the move, which concludes %s", id, b.Role, t.ArgumentID, t.Kind, t.Proposition, c)
		}
	}
	return ""
}
