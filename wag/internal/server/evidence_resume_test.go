package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/yourorg/wag/internal/llm"
)

// E1.6 (docs/e1-evidence.md): a game with a record set resumed from its
// folder of the data folder, and citation strings in such a game. Each
// test's outcome is the plan's, written before the tests ran.

// stoppedConsent plays the fixture's game of two model seats, A and B, on
// the data folder store with the evidence folder dir: A argues a1, B
// passes, A offers the quotation for P (game:b2), and B's consent is
// recorded and saved, and the turn stopped before it is played (G). It
// returns the data folder as the stop left it and the consent's revision.
func stoppedConsent(t *testing.T, store, dir string) (string, string) {
	t.Helper()
	var mu sync.Mutex
	var replies []func() string
	srv, ts, _ := journalServer(t, store, "m-1", func(n int, _ llm.Request) (*llm.Response, error) {
		mu.Lock()
		f := replies[n-1]
		mu.Unlock()
		return &llm.Response{Text: f(), FinishReason: "stop", Model: "m-1"}, nil
	})
	srv.SetEvidence(dir)
	rev := func() string {
		srv.mu.Lock()
		defer srv.mu.Unlock()
		for _, ss := range srv.sessions {
			r, _ := ss.GS.Evidence.Revision("game:b2")
			return r
		}
		return ""
	}
	offer := `{"type":"evidence_offer","offer":{"claim_id":"claim-quote","argument_id":"a1","kind":"premise","statement":"` + premP +
		`","role":"support","interpretation":"` + interpX + `","warrant":"` + warrantW + `"}}`
	replies = []func() string{
		func() string { return replyWith(string(claimAction)) },
		func() string { return replyPass },
		func() string { return replyWith(offer) },
		func() string {
			return replyWith(fmt.Sprintf(`{"type":"evidence_consent","binding_id":"game:b2","revision":%q}`, rev()))
		},
	}
	req := evidenceInquiry()
	req.Seats = []Seat{{Name: "A", Kind: "model", Model: "m"}, {Name: "B", Kind: "model", Model: "m"}}
	id := newGame(t, ts, req)
	for i := 0; i < 3; i++ {
		if code, msg := ai(t, ts, id); code != http.StatusOK {
			t.Fatalf("move %d: %d %s", i+1, code, msg)
		}
	}
	want := rev()
	stopped := stopAt(t, srv, 1, store)
	if code, msg := ai(t, ts, id); code == http.StatusOK || !strings.Contains(msg, "stopped after the model's reply") {
		t.Fatalf("the stopped turn: %d %q", code, msg)
	}
	return stopped, want
}

// Y4: a game of two model seats with the record set, stopped after a model's
// consent was recorded and before it was played, resumed: it resumes; the
// consent is played without asking the model; its revision is the recorded
// one.
func TestY4AConsentRecordedBeforeTheStopIsPlayed(t *testing.T) {
	dir := writeFixture(t, t.TempDir(), "fixture", fixtureSet())
	stopped, want := stoppedConsent(t, t.TempDir(), dir)
	srv, ts, p := journalServer(t, stopped, "m-1", script(replyPass))
	srv.SetEvidence(dir)
	code, out := resumeHTTP(t, ts, folder(t, stopped))
	if code != http.StatusOK {
		t.Fatalf("Y4: the resume: %d %v", code, out["error"])
	}
	id, _ := out["id"].(string)
	if code, msg := ai(t, ts, id); code != http.StatusOK || p.calls() != 0 {
		t.Fatalf("Y4: B's turn: %d %q, %d calls", code, msg, p.calls())
	}
	srv.mu.Lock()
	ev := srv.sessions[id].GS.Evidence
	admitted, events := ev.Admitted("game:b2"), ev.Bundle.Events
	srv.mu.Unlock()
	if !admitted || len(events) != 2 || events[1].CreatedBy != "game:player:B" || events[1].BindingRevision != want {
		t.Fatalf("Y4: admitted %v, events %+v (want the revision %s)", admitted, events, want)
	}
}

// Y10: Y4 with a payload changed before the resume: refused, naming the
// representation. Y9, for a resume: a server with other pins: refused,
// naming the pins.
func TestY10Y9AResumeWithAChangedPayloadOrOtherPins(t *testing.T) {
	dir := writeFixture(t, t.TempDir(), "fixture", fixtureSet())
	stopped, _ := stoppedConsent(t, t.TempDir(), dir)
	p := filepath.Join(dir, "fixture", "excerpt.txt")
	orig, _ := os.ReadFile(p)
	if err := os.WriteFile(p, append([]byte{}, append(orig[:len(orig)-1], '!')...), 0o644); err != nil {
		t.Fatal(err)
	}
	srv, ts, _ := journalServer(t, stopped, "m-1", script(replyPass))
	srv.SetEvidence(dir)
	if code, out := resumeHTTP(t, ts, folder(t, stopped)); code == http.StatusOK || !strings.Contains(fmt.Sprint(out["error"]), "rep-excerpt") {
		t.Fatalf("Y10: %d %v", code, out["error"])
	}
	if err := os.WriteFile(p, orig, 0o644); err != nil {
		t.Fatal(err)
	}
	srv2, ts2, _ := journalServer(t, stopped, "m-1", script(replyPass))
	srv2.SetEvidence(dir)
	srv2.mu.Lock()
	srv2.pins.NASCQ = strings.Repeat("9", 64)
	srv2.mu.Unlock()
	if code, out := resumeHTTP(t, ts2, folder(t, stopped)); code == http.StatusOK || !strings.Contains(fmt.Sprint(out["error"]), "pins") {
		t.Fatalf("Y9: %d %v", code, out["error"])
	}
}

// Y8: a new game whose citation strings are a source's reference URL and a
// binding's id: they stay citations.
func TestY8CitationStringsStayStrings(t *testing.T) {
	g := newEG(t)
	g.admittedP()
	g.ok("A", passMove)
	m := expertFor(premP)
	m["evidence"] = []string{"https://example.org/agency-report", "game:b2"}
	mr := g.ok("B", m)
	id := mr.Result["arg_ids"].([]any)[0].(string)
	a := g.st.Arguments[id]
	if !reflect.DeepEqual(a.Evidence, []string{"https://example.org/agency-report", "game:b2"}) || len(a.EvidenceRefs) != 0 || a.Ungrounded {
		t.Fatalf("Y8: the argument's sources %v, cited bindings %+v", a.Evidence, a.EvidenceRefs)
	}
	g.ok("A", passMove)
	g.ok("B", withdraw("game:b2", g.rev("game:b2")))
	if a := g.st.Arguments[id]; a.Ungrounded || len(a.EvidenceRefs) != 0 {
		t.Fatalf("Y8: after the withdrawal: %+v", a)
	}
}
