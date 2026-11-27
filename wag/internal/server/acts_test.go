package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/evidence"
	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"
)

// E1.3 and E1.4 (docs/e1-evidence.md): offers, consents and withdrawals as
// moves, and moves that cite bindings. Each test's outcome is the plan's,
// written before the tests ran. The fixture's game: A (first) and B, the
// record set "fixture" (evidence_fixture_test.go); A's opening argument a1
// is AS:consequences for "tax meat".

const (
	premP  = "brings_about(tax_meat,lower_emissions)"
	premV  = "good(lower_emissions)"
	excBad = "bad_outweighs(tax_meat,lower_emissions)"
	claimS = "should_be_performed(tax_meat)"
	infA1  = "undercut(a1)"

	interpX  = "The report finds that the tax brings the fall about."
	interpY  = "The report finds only that the two go together."
	warrantW = "An agency's report on its own measurements is a defeasible indicator of what it measured."
)

// eg is a game of the fixture's record set, played through the server.
type eg struct {
	t   *testing.T
	srv *Server
	ts  *httptest.Server
	dir string // the evidence folder
	id  string
	st  State // the board after the last request
}

func newEG(t *testing.T, edit ...func(*NewGameRequest)) *eg {
	t.Helper()
	srv, ts, dir := evidenceServer(t)
	req := evidenceInquiry()
	for _, e := range edit {
		e(&req)
	}
	g := &eg{t: t, srv: srv, ts: ts, dir: dir}
	if code := call(t, ts, "POST", "/api/games", req, &g.st); code != http.StatusOK {
		t.Fatalf("the fixture's game: %d", code)
	}
	g.id = g.st.ID
	return g
}

// move plays a as who, who must be to move.
func (g *eg) move(who string, a any) moveReply {
	g.t.Helper()
	if g.st.Active != who {
		g.t.Fatalf("%s is to move, not %s", g.st.Active, who)
	}
	mr := post(g.t, g.ts, "/api/games/"+g.id+"/move", map[string]any{"action": a})
	g.st = mr.State
	return mr
}

// ok plays a as who; it must be accepted.
func (g *eg) ok(who string, a any) moveReply {
	g.t.Helper()
	mr := g.move(who, a)
	if mr.Result["accepted"] != true {
		g.t.Fatalf("%s's move %v is not accepted: %v", who, a, mr.Result["error"])
	}
	return mr
}

// no plays a as who; it must be refused, saying says.
func (g *eg) no(who string, a any, says string) moveReply {
	g.t.Helper()
	mr := g.move(who, a)
	msg, _ := mr.Result["error"].(string)
	if mr.Result["accepted"] == true || !strings.Contains(msg, says) {
		g.t.Fatalf("%s's move %v: accepted %v, %q; want a refusal saying %q", who, a, mr.Result["accepted"], msg, says)
	}
	return mr
}

// with runs f on the game, under the server's lock.
func (g *eg) with(f func(gs *game.GameState)) {
	g.srv.mu.Lock()
	defer g.srv.mu.Unlock()
	f(g.srv.sessions[g.id].GS)
}

func (g *eg) bundle() evidence.Bundle {
	var b evidence.Bundle
	g.with(func(gs *game.GameState) { b = gs.Evidence.Bundle })
	return b
}

func (g *eg) rev(binding string) string {
	var r string
	g.with(func(gs *game.GameState) { r, _ = gs.Evidence.Revision(binding) })
	if r == "" {
		g.t.Fatalf("no binding %s", binding)
	}
	return r
}

// last is the binding the last move's act made or named.
func (g *eg) last() string {
	var id string
	g.with(func(gs *game.GameState) {
		if m := gs.MoveHistory[len(gs.MoveHistory)-1]; m.Evidence != nil {
			id = m.Evidence.BindingID
		}
	})
	if id == "" {
		g.t.Fatal("the last move is not an evidence act")
	}
	return id
}

func (g *eg) admitted(binding string) bool {
	var ok bool
	g.with(func(gs *game.GameState) { ok = gs.Evidence.Admitted(binding) })
	return ok
}

func (g *eg) consents(binding string) string {
	var c []string
	g.with(func(gs *game.GameState) { c = gs.Evidence.Consents(binding) })
	return strings.Join(c, ",")
}

// opened: A argues a1, B passes; A is to move, in round 2.
func (g *eg) opened() {
	g.t.Helper()
	g.ok("A", claimAction)
	g.ok("B", passMove)
}

// admittedP: opened; A offers the quotation for a1's premise P, support (the
// binding game:b2), B consents: it is admitted. A is to move, in round 3.
func (g *eg) admittedP() {
	g.t.Helper()
	g.opened()
	g.ok("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpX, warrantW))
	g.ok("B", consent("game:b2", g.rev("game:b2")))
}

var passMove = map[string]any{"type": "abstain"}

// undercutA1 is a direct undercutter of a1, which concludes undercut(a1):
// an undercutter that answers no question (a1's scheme has none whose locus
// is its inference; an answer at an exception concludes the exception).
var undercutA1 = map[string]any{"type": "attack", "attack_type": "undercutter", "scheme_id": "AS:expert_opinion", "target_arg_id": "a1",
	"variable_bindings": map[string]string{"W": "a statistician", "D": "evidence"}}

func offerAt(arg, claim, kind, stmt, role, interp, warrant string) map[string]any {
	o := map[string]any{"claim_id": claim, "argument_id": arg, "kind": kind, "role": role, "interpretation": interp, "warrant": warrant}
	if stmt != "" {
		o["statement"] = stmt
	}
	return map[string]any{"type": "evidence_offer", "offer": o}
}

func consent(binding, rev string) map[string]any {
	return map[string]any{"type": "evidence_consent", "binding_id": binding, "revision": rev}
}

func withdraw(binding, rev string) map[string]any {
	return map[string]any{"type": "evidence_withdraw", "binding_id": binding, "revision": rev}
}

// cite is the move a, citing the bindings refs.
func cite(a map[string]any, refs ...string) map[string]any {
	b := map[string]any{}
	for k, v := range a {
		b[k] = v
	}
	b["evidence_refs"] = refs
	return b
}

// expertFor is an argument from expert opinion for the statement.
func expertFor(stmt string) map[string]any {
	return map[string]any{"type": "support", "scheme_id": "AS:expert_opinion", "target_stmt_id": stmt,
		"variable_bindings": map[string]string{"W": "an agency economist", "D": "emissions"}}
}

// undermine is an undermining of a1's premise by expert opinion.
func undermine(premise string) map[string]any {
	return map[string]any{"type": "attack", "attack_type": "undermining", "target_arg_id": "a1", "target_premise_id": premise,
		"scheme_id": "AS:expert_opinion", "variable_bindings": map[string]string{"W": "a farm economist", "D": "emissions"}}
}

// labels are the board's labels of the statements, and the weights and
// force of the arguments.
func (g *eg) labels() map[string]string {
	out := map[string]string{}
	for id, s := range g.st.Statements {
		out[id] = s.Label
	}
	for id, a := range g.st.Arguments {
		out["arg "+id] = strings.Join([]string{jsonOf(a.Weight), jsonOf(a.InForce)}, " ")
	}
	return out
}

func jsonOf(v any) string { b, _ := json.Marshal(v); return string(b) }

// O1: A offers the quotation for a1's premise P, role support: accepted; a
// target (by game:wag: a1, its revision, kind premise, P) and a binding (by
// game:player:A) are recorded, with A's consent to the binding's revision;
// not admitted. O2: then B consents to that revision: admitted.
func TestO1O2AnOfferThenAConsent(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.ok("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpX, warrantW))
	var rev string
	g.with(func(gs *game.GameState) {
		b := gs.Evidence.Bundle
		argRev, _ := gs.ArgumentRevision("a1")
		if len(b.Targets) != 1 || len(b.Bindings) != 1 || len(b.Events) != 1 {
			t.Fatalf("O1: %d targets, %d bindings, %d events", len(b.Targets), len(b.Bindings), len(b.Events))
		}
		tg, bd, ev := b.Targets[0], b.Bindings[0], b.Events[0]
		if tg.ID != "game:t2" || tg.CreatedBy != "game:wag" || tg.ArgumentID != "a1" || tg.ArgumentRevision != argRev || argRev == "" ||
			tg.Kind != "premise" || tg.Proposition != premP {
			t.Fatalf("O1: the target %+v", tg)
		}
		if bd.ID != "game:b2" || bd.CreatedBy != "game:player:A" || bd.ClaimID != "claim-quote" || bd.TargetID != "game:t2" || bd.Role != "support" ||
			bd.Interpretation != interpX || bd.Warrant != warrantW {
			t.Fatalf("O1: the binding %+v", bd)
		}
		rev, _ = gs.Evidence.Revision("game:b2")
		if ev.ID != "game:e2" || ev.CreatedBy != "game:player:A" || ev.Type != "consent" || ev.BindingID != "game:b2" || ev.BindingRevision != rev || len(rev) != 64 {
			t.Fatalf("O1: the event %+v (revision %s)", ev, rev)
		}
		if gs.Evidence.Admitted("game:b2") {
			t.Fatal("O1: admitted on one consent")
		}
	})
	g.ok("B", consent("game:b2", rev))
	if !g.admitted("game:b2") || g.consents("game:b2") != "A,B" {
		t.Fatalf("O2: admitted %v, consents %s", g.admitted("game:b2"), g.consents("game:b2"))
	}
}

// O3: A offers the quotation for P with interpretation X, then with Y; B
// consents naming Y's binding and X's revision: refused, stale.
func TestO3AConsentToAnotherRevision(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.ok("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpX, warrantW))
	g.ok("B", passMove)
	g.ok("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpY, warrantW))
	g.no("B", consent("game:b4", g.rev("game:b2")), "stale")
	if g.admitted("game:b4") || g.consents("game:b4") != "A" {
		t.Fatalf("O3: %s", g.consents("game:b4"))
	}
}

// O4: O2; B consents again: refused, a duplicate consent. O5: O2; B
// withdraws: not admitted; A's consent and the records stay. O7: O5; B
// consents again to the same revision: admitted again.
func TestO4O5O7ConsentsAndWithdrawals(t *testing.T) {
	g := newEG(t)
	g.admittedP()
	rev := g.rev("game:b2")
	g.ok("A", passMove)
	g.no("B", consent("game:b2", rev), "already stands")
	g.ok("B", withdraw("game:b2", rev))
	b := g.bundle()
	if g.admitted("game:b2") || g.consents("game:b2") != "A" || len(b.Targets) != 1 || len(b.Bindings) != 1 || len(b.Events) != 3 {
		t.Fatalf("O5: admitted %v, consents %q, %d targets, %d bindings, %d events", g.admitted("game:b2"), g.consents("game:b2"), len(b.Targets), len(b.Bindings), len(b.Events))
	}
	g.ok("A", passMove)
	g.ok("B", consent("game:b2", rev))
	if !g.admitted("game:b2") {
		t.Fatal("O7: not admitted again")
	}
}

// O6: O1; B withdraws without having consented: refused.
func TestO6AWithdrawalWithoutAConsent(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.ok("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpX, warrantW))
	g.no("B", withdraw("game:b2", g.rev("game:b2")), "no consent")
	if n := len(g.bundle().Events); n != 1 {
		t.Fatalf("O6: %d events", n)
	}
}

// O8: an offer for an argument not on the board, a question's argument, a
// presumption argument (and an exception argument), a statement not in the
// argument, a kind it does not have, or a kind source or annotation:
// refused, each; no target, binding or event is left.
func TestO8OffersAtWhatIsNotATarget(t *testing.T) {
	g := newEG(t)
	g.ok("A", claimAction)
	g.ok("B", map[string]any{"type": "question", "target_arg_id": "a1", "answers_cq_id": "CQ:likelihood"})
	var question, presumption, exception string
	g.with(func(gs *game.GameState) {
		for id, a := range gs.Adapter.Graph.Arguments {
			switch carneades.ArgumentRole(a) {
			case carneades.RoleQuestion:
				question = id
			case carneades.RolePresumption:
				presumption = id
			case carneades.RoleException:
				exception = id
			}
		}
	})
	if question == "" || presumption == "" || exception == "" {
		t.Fatalf("O8: the board has a question's argument %q, a presumption %q, an exception argument %q", question, presumption, exception)
	}
	for name, o := range map[string]map[string]any{
		"an argument not on the board":  offerAt("a99", "claim-quote", "premise", premP, "support", interpX, warrantW),
		"a question's argument":         offerAt(question, "claim-quote", "conclusion", "", "support", interpX, warrantW),
		"a presumption argument":        offerAt(presumption, "claim-quote", "conclusion", "", "support", interpX, warrantW),
		"an exception argument":         offerAt(exception, "claim-quote", "conclusion", "", "support", interpX, warrantW),
		"a statement not in a1":         offerAt("a1", "claim-quote", "premise", "good(nothing)", "support", interpX, warrantW),
		"a kind a1 does not have":       offerAt("a1", "claim-quote", "assumption", premP, "support", interpX, warrantW),
		"a premise given as conclusion": offerAt("a1", "claim-quote", "conclusion", premP, "support", interpX, warrantW),
		"kind source":                   offerAt("a1", "claim-quote", "source", premP, "support", interpX, warrantW),
		"kind annotation":               offerAt("a1", "claim-quote", "annotation", premP, "support", interpX, warrantW),
	} {
		mr := g.move("A", o)
		if mr.Result["accepted"] == true {
			t.Fatalf("O8: %s accepted", name)
		}
		t.Logf("O8, %s: %v", name, mr.Result["error"])
	}
	if b := g.bundle(); len(b.Targets)+len(b.Bindings)+len(b.Events) != 0 {
		t.Fatalf("O8: records left: %d targets, %d bindings, %d events", len(b.Targets), len(b.Bindings), len(b.Events))
	}
}

// O9: an offer of a claim not in the record set, of role source_challenge,
// or without an interpretation or a warrant: refused, each, leaving
// nothing.
func TestO9OffersThatAreNotWhole(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.no("A", offerAt("a1", "claim-x", "premise", premP, "support", interpX, warrantW), "no claim")
	g.no("A", offerAt("a1", "claim-quote", "premise", premP, "source_challenge", interpX, warrantW), "source_challenge")
	g.no("A", offerAt("a1", "claim-quote", "premise", premP, "support", " ", warrantW), "interpretation")
	g.no("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpX, ""), "warrant")
	if b := g.bundle(); len(b.Targets)+len(b.Bindings)+len(b.Events) != 0 {
		t.Fatalf("O9: records left")
	}
}

// O10: O2; then another argument for P, an answer at a1's exception
// bad_outweighs, a question about a1 and a direct undercutter of a1: a1's
// revision is unchanged, and so is the binding's; the admission stands.
func TestO10LaterMovesChangeNoRevision(t *testing.T) {
	g := newEG(t)
	g.admittedP()
	var argRev string
	g.with(func(gs *game.GameState) { argRev, _ = gs.ArgumentRevision("a1") })
	rev := g.rev("game:b2")
	check := func(after string) {
		t.Helper()
		var now string
		g.with(func(gs *game.GameState) { now, _ = gs.ArgumentRevision("a1") })
		if now != argRev || g.rev("game:b2") != rev || !g.admitted("game:b2") {
			t.Fatalf("O10, after %s: a1 %s (was %s), the binding %s (was %s), admitted %v", after, now, argRev, g.rev("game:b2"), rev, g.admitted("game:b2"))
		}
	}
	for name, m := range map[string]any{
		"another argument for P":     expertFor(premP),
		"an answer at bad_outweighs": map[string]any{"type": "attack", "answers_cq_id": "CQ:bad_outweighs", "scheme_id": "AS:bad_outweighs_good", "target_arg_id": "a1", "variable_bindings": map[string]string{"H": "food costlier"}},
		"a question about a1":        map[string]any{"type": "question", "target_arg_id": "a1", "answers_cq_id": "CQ:likelihood"},
		"a direct undercutter of a1": undercutA1,
	} {
		g.ok("A", passMove)
		g.ok("B", m)
		check(name)
	}
}

// O11: each of O1, O2, O5 and a refused act: the history records the act
// with its mark; the board's log shows an accepted act as accepted and a
// refused one as refused; the trace and the export record it; no label,
// weight or assumption changes.
func TestO11TheRecordOfAnAct(t *testing.T) {
	g := newEG(t)
	g.opened()
	before := g.labels()
	step := func(who string, a any, accepted bool, act string) {
		t.Helper()
		mr := g.move(who, a)
		if (mr.Result["accepted"] == true) != accepted {
			t.Fatalf("O11: %v accepted %v: %v", a, mr.Result["accepted"], mr.Result["error"])
		}
		l := g.st.Log[len(g.st.Log)-1]
		if l.Accepted != accepted || (accepted && l.Kind != "evidence") || (!accepted && l.Kind != "rejected") {
			t.Fatalf("O11: the log reads %v as %+v", a, l)
		}
		var mark *game.EvidenceMark
		g.with(func(gs *game.GameState) { mark = gs.MoveHistory[len(gs.MoveHistory)-1].Evidence })
		if accepted && (mark == nil || mark.Act != act || mark.Time == "" || len(mark.BindingRevision) != 64) || !accepted && mark != nil {
			t.Fatalf("O11: the history's mark %+v", mark)
		}
		if after := g.labels(); !reflect.DeepEqual(after, before) {
			t.Fatalf("O11: %v changed the labels", a)
		}
	}
	step("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpX, warrantW), true, game.ActOffer)
	rev := g.rev("game:b2")
	step("B", consent("game:b2", rev), true, game.ActConsent)
	g.ok("A", passMove)
	before = g.labels()
	step("B", withdraw("game:b2", rev), true, game.ActWithdraw)
	step("A", withdraw("game:b2", "0123456789ab"), false, "")
	// The export's history and the trace's decisions record the acts.
	var ex struct {
		History []game.Move `json:"history"`
	}
	if code := call(t, g.ts, "GET", "/api/games/"+g.id+"/export", nil, &ex); code != http.StatusOK {
		t.Fatal(code)
	}
	marks := 0
	for _, m := range ex.History {
		if m.Evidence != nil {
			marks++
		}
	}
	if marks != 3 {
		t.Fatalf("O11: the export's history has %d marked acts", marks)
	}
	res, err := direct.Get(g.ts.URL + "/api/games/" + g.id + "/traces")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var acts int
	dec := json.NewDecoder(res.Body)
	for dec.More() {
		var tr struct {
			Type   string               `json:"type"`
			Action nascq.ProposedAction `json:"action"`
		}
		if err := dec.Decode(&tr); err != nil {
			t.Fatal(err)
		}
		if tr.Type == "decision" && game.IsEvidenceAct(tr.Action.Type) {
			acts++
		}
	}
	if acts != 4 {
		t.Fatalf("O11: the trace has %d evidence acts, not 4 (3 accepted, 1 refused)", acts)
	}
}

// O12: O1, then the same offer again: refused, naming the binding.
func TestO12TheSameOfferTwice(t *testing.T) {
	g := newEG(t)
	g.opened()
	o := offerAt("a1", "claim-quote", "premise", premP, "support", interpX, warrantW)
	g.ok("A", o)
	g.ok("B", passMove)
	g.no("A", o, "game:b2")
}

// O13: A offers the same claim for P with another warrant: another binding,
// another revision; B's consent to the first does not admit it.
func TestO13AnotherWarrantIsAnotherBinding(t *testing.T) {
	g := newEG(t)
	g.admittedP()
	g.ok("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpX, warrantW+" Its methods are published."))
	if g.rev("game:b4") == g.rev("game:b2") || g.admitted("game:b4") || g.consents("game:b4") != "A" {
		t.Fatalf("O13: %s %s, admitted %v", g.rev("game:b4"), g.rev("game:b2"), g.admitted("game:b4"))
	}
}

// O14: a round in which both players make acts only: not over (an act is
// not a pass).
func TestO14ARoundOfActsIsNotARoundOfPasses(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.ok("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpX, warrantW))
	g.ok("B", consent("game:b2", g.rev("game:b2")))
	if g.st.Over || g.st.Round != 3 {
		t.Fatalf("O14: over %v, round %d", g.st.Over, g.st.Round)
	}
}

// O15: B consents by the revision's first 12 digits: admitted, the full
// revision recorded; then by 12 digits of another binding's revision:
// refused.
func TestO15ARevisionsFirst12Digits(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.ok("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpX, warrantW))
	rev := g.rev("game:b2")
	g.ok("B", consent("game:b2", rev[:12]))
	b := g.bundle()
	var mark *game.EvidenceMark
	g.with(func(gs *game.GameState) { mark = gs.MoveHistory[len(gs.MoveHistory)-1].Evidence })
	if len(b.Events) != 2 || mark == nil {
		t.Fatalf("O15: %d events, mark %+v", len(b.Events), mark)
	}
	if !g.admitted("game:b2") || b.Events[1].BindingRevision != rev || mark.BindingRevision != rev {
		t.Fatalf("O15: admitted %v, event %s, mark %+v", g.admitted("game:b2"), b.Events[1].BindingRevision, mark)
	}
	g.ok("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpY, warrantW))
	g.no("B", consent("game:b4", rev[:12]), "stale")
}

// O16: a refused act: the usual penalty; no record.
func TestO16ARefusedAct(t *testing.T) {
	g := newEG(t)
	g.opened()
	var pen, before float64
	g.with(func(gs *game.GameState) { pen, before = gs.Config.Scoring.InvalidActionPenalty, gs.Players[0].Score })
	g.no("A", consent("game:b9", strings.Repeat("0", 64)), "no binding")
	var after float64
	g.with(func(gs *game.GameState) { after = gs.Players[0].Score })
	if pen >= 0 || after != before+pen || len(g.bundle().Events) != 0 {
		t.Fatalf("O16: penalty %v, score %v then %v, %d events", pen, before, after, len(g.bundle().Events))
	}
}

// O17: the argument revision of a1 equals the golden value:
// wag-argument-revision/1 over AS:consequences, P and V in order,
// should_be_performed(tax_meat), the three exceptions sorted, 0.65.
func TestO17TheArgumentRevisionsGoldenValue(t *testing.T) {
	g := newEG(t)
	g.ok("A", claimAction)
	raw, _ := json.Marshal(struct {
		Algorithm  string
		Scheme     string
		Premises   []string
		Conclusion string
		Exceptions []string
		Weight     float64
	}{"wag-argument-revision/1", "AS:consequences", []string{premP, premV}, claimS,
		[]string{excBad, "gain_discounted_by_bearer(tax_meat,lower_emissions)", "too_slight(tax_meat,lower_emissions)"}, 0.65})
	sum := sha256.Sum256(raw)
	want := hex.EncodeToString(sum[:])
	const golden = "92622d17f45a29c9c9f24ff68a54dd55ae38a6cd1b9eccdf275695e2c7001d04"
	var got string
	g.with(func(gs *game.GameState) { got, _ = gs.ArgumentRevision("a1") })
	if got != want || got != golden {
		t.Fatalf("O17: %s; from the plan's parts %s; golden %s", got, want, golden)
	}
}

// O18: the board's preview of an act, and an act tried as a reply's
// alternative: the evidence state after as before.
func TestO18ASimulatedActMakesNoAct(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.ok("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpX, warrantW))
	g.ok("B", passMove)
	before := g.bundle()
	var out map[string]any
	if code := call(t, g.ts, "POST", "/api/games/"+g.id+"/preview", map[string]any{"action": offerAt("a1", "claim-observation", "premise", premP, "support", interpX, warrantW)}, &out); code != http.StatusOK {
		t.Fatalf("O18: the preview: %d %v", code, out)
	}
	if after := g.bundle(); !reflect.DeepEqual(after, before) {
		t.Fatal("O18: the preview made an act")
	}
	g.with(func(gs *game.GameState) {
		alt := nascq.ProposedAction{Type: game.ActConsent, BindingID: "game:b2"}
		alt.Revision, _ = gs.Evidence.Revision("game:b2")
		gs.Simulate(alt) // A's own consent stands: refused, a rejection undone
		r, err := gs.Simulate(nascq.ProposedAction{Type: game.ActOffer, Offer: &nascq.EvidenceOffer{ClaimID: "claim-observation", ArgumentID: "a1", Kind: "premise", Statement: premP, Role: "support", Interpretation: interpX, Warrant: warrantW}})
		if err != nil || !r.Accepted {
			t.Fatalf("O18: the simulated offer: %v %+v", err, r)
		}
	})
	if after := g.bundle(); !reflect.DeepEqual(after, before) {
		t.Fatal("O18: a simulation made an act")
	}
	var n int
	g.with(func(gs *game.GameState) { n = len(gs.MoveHistory) })
	if n != 4 {
		t.Fatalf("O18: %d moves in the history", n)
	}
}

// O19: an offer at a1's inference: its target's proposition is undercut(a1).
func TestO19AnOfferAtTheInference(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.ok("A", offerAt("a1", "claim-observation", "inference", "", "oppose", interpX, warrantW))
	if b := g.bundle(); len(b.Targets) != 1 || b.Targets[0].Kind != "inference" || b.Targets[0].Proposition != infA1 {
		t.Fatalf("O19: %+v", b.Targets)
	}
}

// M1: O2 (admitted, support, P); B argues for P with expert opinion, citing
// the binding: accepted; the argument records the reference.
// M7: then B withdraws: the argument stays, in force, the labels as before;
// its reference is shown withdrawn; the argument is marked no longer
// grounded.
func TestM1M7ACitationThenAWithdrawal(t *testing.T) {
	g := newEG(t)
	g.admittedP()
	g.ok("A", passMove)
	mr := g.ok("B", cite(expertFor(premP), "game:b2"))
	id := mr.Result["arg_ids"].([]any)[0].(string)
	a := g.st.Arguments[id]
	if len(a.EvidenceRefs) != 1 || a.EvidenceRefs[0] != (RefState{"game:b2", true}) || a.Ungrounded {
		t.Fatalf("M1: %+v %v", a.EvidenceRefs, a.Ungrounded)
	}
	g.ok("A", passMove)
	before := g.labels()
	g.ok("B", withdraw("game:b2", g.rev("game:b2")))
	a = g.st.Arguments[id]
	if !a.InForce || !reflect.DeepEqual(g.labels(), before) || a.EvidenceRefs[0] != (RefState{"game:b2", false}) || !a.Ungrounded {
		t.Fatalf("M7: in force %v, refs %+v, ungrounded %v", a.InForce, a.EvidenceRefs, a.Ungrounded)
	}
}

// M2: O1 (offered, not admitted); the same move: refused, not admitted.
func TestM2ABindingNotAdmitted(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.ok("A", offerAt("a1", "claim-quote", "premise", premP, "support", interpX, warrantW))
	g.no("B", cite(expertFor(premP), "game:b2"), "not admitted")
}

// M3: O2; an undermining of P citing the binding: refused, the role does not
// agree. M4: O2; a move for another statement citing it: refused, the target
// does not agree. M10: a move citing an unknown binding id: refused.
func TestM3M4M10CitationsThatDoNotAgree(t *testing.T) {
	g := newEG(t)
	g.admittedP()
	g.ok("A", passMove)
	g.no("B", cite(undermine(premP), "game:b2"), "does not agree")
	g.no("B", cite(expertFor(premV), "game:b2"), "does not agree")
	g.no("B", cite(expertFor(premP), "https://example.org/agency-report"), "no binding")
	g.no("B", cite(expertFor(premP), "game:b9"), "no binding")
}

// M5: Require sources on; expert opinion for P citing neither a string nor a
// binding: refused by the citation rule; with the admitted support binding,
// accepted; with an admitted definition binding for P only, refused.
func TestM5RequireSourcesWithBindings(t *testing.T) {
	g := newEG(t, func(r *NewGameRequest) { r.RequireEvidence = true })
	g.admittedP()
	g.ok("A", offerAt("a1", "claim-paraphrase", "premise", premP, "definition", interpX, warrantW))
	g.ok("B", consent("game:b4", g.rev("game:b4")))
	g.ok("A", passMove)
	g.no("B", expertFor(premP), "E1:")
	g.no("B", cite(expertFor(premP), "game:b4"), "definition")
	g.ok("B", cite(expertFor(premP), "game:b2"))
}

// M6: Require sources on; an undermining of P by expert opinion citing an
// admitted oppose binding at P: accepted; citing an admitted context
// binding at P only, refused.
func TestM6RequireSourcesWithAnOpposeBinding(t *testing.T) {
	g := newEG(t, func(r *NewGameRequest) { r.RequireEvidence = true })
	g.opened()
	g.ok("A", offerAt("a1", "claim-observation", "premise", premP, "oppose", interpY, warrantW))
	g.ok("B", consent("game:b2", g.rev("game:b2")))
	g.ok("A", offerAt("a1", "claim-paraphrase", "premise", premP, "context", interpX, warrantW))
	g.ok("B", consent("game:b4", g.rev("game:b4")))
	g.ok("A", passMove)
	g.no("B", cite(undermine(premP), "game:b4"), "context")
	g.ok("B", cite(undermine(premP), "game:b2"))
}

// M8: M1 citing two admitted bindings, then one withdrawn: one reference
// withdrawn, one admitted, not marked; and M1 citing one binding and a
// string, then the binding withdrawn: not marked (the string grounds it).
func TestM8ArgumentsStillGrounded(t *testing.T) {
	g := newEG(t)
	g.admittedP()
	g.ok("A", offerAt("a1", "claim-paraphrase", "premise", premP, "support", interpX, warrantW))
	g.ok("B", consent("game:b4", g.rev("game:b4")))
	g.ok("A", passMove)
	mr := g.ok("B", cite(expertFor(premP), "game:b2", "game:b4"))
	two := mr.Result["arg_ids"].([]any)[0].(string)
	g.ok("A", passMove)
	withString := cite(map[string]any{"type": "support", "scheme_id": "AS:expert_opinion", "target_stmt_id": premP, "evidence": []string{"https://example.org/agency-report"},
		"variable_bindings": map[string]string{"W": "a second economist", "D": "emissions"}}, "game:b2")
	mr = g.ok("B", withString)
	one := mr.Result["arg_ids"].([]any)[0].(string)
	g.ok("A", passMove)
	g.ok("B", withdraw("game:b2", g.rev("game:b2")))
	a := g.st.Arguments[two]
	if !reflect.DeepEqual(a.EvidenceRefs, []RefState{{"game:b2", false}, {"game:b4", true}}) || a.Ungrounded {
		t.Fatalf("M8: two bindings, one withdrawn: %+v %v", a.EvidenceRefs, a.Ungrounded)
	}
	if a := g.st.Arguments[one]; a.Ungrounded || len(a.Evidence) != 1 {
		t.Fatalf("M8: a binding and a string: %+v %v", a.EvidenceRefs, a.Ungrounded)
	}
}

// M9: M1's moves with a citation string in place of the binding: the same
// labels and the same open questions (paired).
func TestM9ABindingChangesNothingAStringDoesNot(t *testing.T) {
	play := func(m map[string]any) (map[string]string, []QuestionState) {
		g := newEG(t)
		g.admittedP()
		g.ok("A", passMove)
		g.ok("B", m)
		return g.labels(), g.st.Questions
	}
	withString := expertFor(premP)
	withString["evidence"] = []string{"https://example.org/agency-report"}
	l1, q1 := play(cite(expertFor(premP), "game:b2"))
	l2, q2 := play(withString)
	if !reflect.DeepEqual(l1, l2) || !reflect.DeepEqual(q1, q2) {
		t.Fatalf("M9: a binding and a string differ: %v / %v; %v / %v", l1, l2, q1, q2)
	}
}

// M11: an admitted oppose binding at a1's inference; a direct undercutter of
// a1 citing it: accepted; a rebuttal of a1 citing it: refused.
func TestM11AnInferenceBinding(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.ok("A", offerAt("a1", "claim-observation", "inference", "", "oppose", interpY, warrantW))
	g.ok("B", consent("game:b2", g.rev("game:b2")))
	g.ok("A", passMove)
	rebut := map[string]any{"type": "attack", "scheme_id": "AS:negative_consequences", "target_stmt_id": claimS,
		"variable_bindings": map[string]string{"A": "tax meat", "G": "food costlier"}}
	g.no("B", cite(rebut, "game:b2"), "does not agree")
	mr := g.ok("B", cite(undercutA1, "game:b2"))
	if c := g.st.Arguments[mr.Result["arg_ids"].([]any)[0].(string)].Conclusion; c != infA1 {
		t.Fatalf("M11: the undercutter concludes %s, not %s", c, infA1)
	}
	// Beyond the plan's cases, the table's last but one row: a defence of
	// a1's inference (an attack on undercut(a1), which concludes its
	// opposite) is grounded by a support binding at the inference, not by
	// the oppose binding.
	g.ok("A", offerAt("a1", "claim-quote", "inference", "", "support", interpX, warrantW))
	support := g.last()
	g.ok("B", consent(support, g.rev(support)))
	defence := map[string]any{"type": "attack", "scheme_id": nascq.OtherSchemeID, "target_stmt_id": infA1,
		"variable_bindings": map[string]string{"P1": "the report covers this country"}}
	g.no("A", cite(defence, "game:b2"), "does not agree")
	mr = g.ok("A", cite(defence, support))
	if c := g.st.Arguments[mr.Result["arg_ids"].([]any)[0].(string)].Conclusion; c != "¬"+infA1 {
		t.Fatalf("M11: the defence concludes %s", c)
	}
}

// M12: an admitted support binding at a1's exception bad_outweighs; B's
// answer at that exception (CQ:bad_outweighs) citing it: accepted; the same
// answer citing an inference binding: refused. M13: an admitted oppose
// binding at bad_outweighs, after B's answer there; A's defence at the
// exception (concluding its opposite) citing it: accepted.
func TestM12M13BindingsAtAnException(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.ok("A", offerAt("a1", "claim-paraphrase", "exception", excBad, "support", interpX, warrantW))
	g.ok("B", consent("game:b2", g.rev("game:b2")))
	g.ok("A", offerAt("a1", "claim-observation", "inference", "", "support", interpY, warrantW))
	g.ok("B", consent("game:b4", g.rev("game:b4")))
	g.ok("A", offerAt("a1", "claim-observation", "exception", excBad, "oppose", interpY, warrantW))
	g.ok("B", consent("game:b6", g.rev("game:b6")))
	g.ok("A", passMove)
	answer := map[string]any{"type": "attack", "answers_cq_id": "CQ:bad_outweighs", "scheme_id": "AS:bad_outweighs_good", "target_arg_id": "a1",
		"variable_bindings": map[string]string{"H": "food costlier"}}
	g.no("B", cite(answer, "game:b4"), "does not agree")
	g.ok("B", cite(answer, "game:b2"))
	defence := map[string]any{"type": "attack", "scheme_id": nascq.OtherSchemeID, "target_stmt_id": excBad,
		"variable_bindings": map[string]string{"P1": "the harm is small beside the gain"}}
	g.ok("A", cite(defence, "game:b6"))
}

// M14: a binding at a1's value premise V, admitted; an argument for V citing
// it: accepted.
func TestM14ABindingForAValue(t *testing.T) {
	g := newEG(t)
	g.opened()
	g.ok("A", offerAt("a1", "claim-paraphrase", "premise", premV, "support", interpX, warrantW))
	g.ok("B", consent("game:b2", g.rev("game:b2")))
	g.ok("A", passMove)
	g.ok("B", cite(expertFor(premV), "game:b2"))
}
