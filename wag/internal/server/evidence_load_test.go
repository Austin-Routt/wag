package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/evidence"
)

// E1.6 (docs/e1-evidence.md): a game with a record set loaded from its
// export. Each test's outcome is the plan's, written before the tests ran.

// played is the game of Y1: O2 (A offers the quotation for P, B consents),
// M1 (B argues for P from expert opinion, citing it) and O5 (B withdraws);
// it returns the game, its export's bytes and the cited argument's id.
func played(t *testing.T) (*eg, []byte, string) {
	t.Helper()
	g := newEG(t)
	g.admittedP()
	g.ok("A", passMove)
	mr := g.ok("B", cite(expertFor(premP), "game:b2"))
	arg := mr.Result["arg_ids"].([]any)[0].(string)
	g.ok("A", passMove)
	g.ok("B", withdraw("game:b2", g.rev("game:b2")))
	res, err := direct.Get(g.ts.URL + "/api/games/" + g.id + "/export")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("the export: %d %v", res.StatusCode, err)
	}
	return g, raw, arg
}

// load posts the export to the load endpoint (with query), and returns the
// status, the state and the error.
func load(t *testing.T, g *eg, raw []byte, query string) (int, State, string) {
	t.Helper()
	var out json.RawMessage
	code := call(t, g.ts, "POST", "/api/load"+query, string(raw), &out)
	var st State
	if code == http.StatusOK {
		if err := json.Unmarshal(out, &st); err != nil {
			t.Fatal(err)
		}
		return code, st, ""
	}
	var e struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(out, &e)
	return code, st, e.Error
}

// Y1: a game with O2, M1 and O5, exported and loaded: it replays; the same
// bindings, the revisions equal to the recorded ones, the same admission
// states and references; the records dated as the acts were made.
func TestY1AGameWithARecordSetLoads(t *testing.T) {
	g, raw, arg := played(t)
	before := g.bundle()
	code, st, msg := load(t, g, raw, "")
	if code != http.StatusOK {
		t.Fatalf("Y1: the load: %d %s", code, msg)
	}
	var after evidence.Bundle
	g.srv.mu.Lock()
	after = g.srv.sessions[st.ID].GS.Evidence.Bundle
	g.srv.mu.Unlock()
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("Y1: the loaded game's records differ:\n%+v\n%+v", after.Events, before.Events)
	}
	if b := st.Evidence.Bindings; len(b) != 1 || b[0].Admitted || !b[0].Withdrawn || b[0].Revision != g.rev("game:b2") {
		t.Fatalf("Y1: the bindings %+v", b)
	}
	if a := st.Arguments[arg]; a == nil || !reflect.DeepEqual(a.EvidenceRefs, []RefState{{"game:b2", false}}) || !a.Ungrounded {
		t.Fatalf("Y1: the cited argument %+v", a)
	}
	// The records are dated as the acts were made: each event's time is its
	// act's mark's.
	var marks []string
	g.srv.mu.Lock()
	for _, m := range g.srv.sessions[st.ID].GS.MoveHistory {
		if m.Evidence != nil {
			marks = append(marks, m.Evidence.Time)
		}
	}
	g.srv.mu.Unlock()
	for i, e := range after.Events {
		if e.CreatedAt != marks[i] {
			t.Fatalf("Y1: event %d dated %s, its act %s", i, e.CreatedAt, marks[i])
		}
	}
}

// Y2: Y1 with a payload missing: refused, naming the representation.
func TestY2APayloadMissing(t *testing.T) {
	g, raw, _ := played(t)
	if err := os.Remove(filepath.Join(g.dir, "fixture", "excerpt.txt")); err != nil {
		t.Fatal(err)
	}
	if code, _, msg := load(t, g, raw, ""); code != http.StatusBadRequest || !strings.Contains(msg, "rep-excerpt") {
		t.Fatalf("Y2: %d %s", code, msg)
	}
}

// Y3: Y1 with records.json changed (a claim's text): refused, its SHA-256.
func TestY3TheRecordSetChanged(t *testing.T) {
	g, raw, _ := played(t)
	editSet(t, g.dir, func(m map[string]any) {
		listOf(m, "claims")[1].(map[string]any)["text"] = "The agency found that a meat tax cuts emissions sharply."
	})
	if code, _, msg := load(t, g, raw, ""); code != http.StatusBadRequest || !strings.Contains(msg, "SHA-256") {
		t.Fatalf("Y3: %d %s", code, msg)
	}
}

// Y9, for a load: a server with other pins (another engine): refused, naming
// the pins. And choice 13: no read-only inspection of such a game.
func TestY9OtherPinsAndNoInspection(t *testing.T) {
	g, raw, _ := played(t)
	if code, _, msg := load(t, g, raw, "?inspect=true"); code != http.StatusBadRequest || !strings.Contains(msg, "read-only") {
		t.Fatalf("choice 13: %d %s", code, msg)
	}
	g.srv.mu.Lock()
	g.srv.pins.Engine = strings.Repeat("e", 64)
	g.srv.mu.Unlock()
	if code, _, msg := load(t, g, raw, ""); code != http.StatusBadRequest || !strings.Contains(msg, "pins") || !strings.Contains(msg, "eeeeeeeeeeee") {
		t.Fatalf("Y9: %d %s", code, msg)
	}
}
