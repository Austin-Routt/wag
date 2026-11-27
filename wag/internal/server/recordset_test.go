package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/evidence"
)

// E1.2 (docs/e1-evidence.md): the record set. Each test's outcome is the
// plan's, written before the tests ran.

// evidenceServer is a test server whose evidence folder holds the fixture
// (fixture_test.go), and the folder.
func evidenceServer(t *testing.T) (*Server, *httptest.Server, string) {
	t.Helper()
	srv, ts := newTestServerWith(t)
	dir := writeFixture(t, t.TempDir(), "fixture", fixtureSet())
	srv.SetEvidence(dir)
	return srv, ts, dir
}

// evidenceInquiry is the fixture's game: an inquiry of A and B, people,
// with the record set "fixture".
func evidenceInquiry() NewGameRequest {
	req := inquiry(2, nil)
	req.Evidence = "fixture"
	return req
}

// refusedGame asks for a game with req and fails the test unless it is
// refused (400) saying each of says, and no game is made.
func refusedGame(t *testing.T, srv *Server, ts *httptest.Server, req NewGameRequest, says ...string) {
	t.Helper()
	var out map[string]any
	code := call(t, ts, "POST", "/api/games", req, &out)
	msg, _ := out["error"].(string)
	if code != http.StatusBadRequest {
		t.Fatalf("not refused: %d %v", code, out)
	}
	t.Logf("refused: %s", msg)
	for _, s := range says {
		if !strings.Contains(msg, s) {
			t.Fatalf("the refusal %q does not say %q", msg, s)
		}
	}
	srv.mu.Lock()
	n := len(srv.sessions)
	srv.mu.Unlock()
	if n != 0 {
		t.Fatalf("a refused game was made: %d games", n)
	}
}

// editSet writes the fixture's record set, edited, as the record set "fixture".
func editSet(t *testing.T, dir string, edit func(map[string]any)) {
	t.Helper()
	set := fixtureSet()
	edit(set)
	writeFixture(t, dir, "fixture", set)
}

// listOf is a list of the record set's JSON, as maps.
func listOf(set map[string]any, key string) []any { return set[key].([]any) }

// R1: the fixture's game is created; the board lists the sources, the claims
// and the dependency; the trace's start and the export record the name and
// records.json's SHA-256.
func TestR1TheFixturesGame(t *testing.T) {
	_, ts, dir := evidenceServer(t)
	var st State
	if code := call(t, ts, "POST", "/api/games", evidenceInquiry(), &st); code != http.StatusOK {
		t.Fatalf("R1: %d", code)
	}
	e := st.Evidence
	if e == nil || len(e.Sources) != 2 || len(e.Claims) != 3 || len(e.Dependencies) != 1 || e.Name != "fixture" {
		t.Fatalf("R1: the board's evidence: %+v", e)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "fixture", "records.json"))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	want := RecordSet{Name: "fixture", SHA256: hex.EncodeToString(sum[:])}
	var ex struct {
		RecordSet *RecordSet `json:"record_set"`
	}
	if code := call(t, ts, "GET", "/api/games/"+st.ID+"/export", nil, &ex); code != http.StatusOK || ex.RecordSet == nil || *ex.RecordSet != want {
		t.Fatalf("R1: the export's record set: %d %+v, want %+v", code, ex.RecordSet, want)
	}
	res, err := direct.Get(ts.URL + "/api/games/" + st.ID + "/traces")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	line, _, _ := bytes.Cut(body, []byte("\n"))
	var start GameStart
	if err := json.Unmarshal(line, &start); err != nil || start.RecordSet == nil || *start.RecordSet != want {
		t.Fatalf("R1: the trace's start: %v %+v", err, start.RecordSet)
	}
}

// R2: a payload changed by one byte: refused, naming the representation.
func TestR2APayloadChanged(t *testing.T) {
	srv, ts, dir := evidenceServer(t)
	p := filepath.Join(dir, "fixture", "report.txt")
	b, _ := os.ReadFile(p)
	b[0] ^= 1
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	refusedGame(t, srv, ts, evidenceInquiry(), "rep-report")
}

// R3: a record set with a key the game writes, even empty: refused, each.
func TestR3ARecordSetWithTheGamesKeys(t *testing.T) {
	for _, k := range []string{"context", "targets", "bindings", "assessments", "events"} {
		t.Run(k, func(t *testing.T) {
			srv, ts, dir := evidenceServer(t)
			editSet(t, dir, func(m map[string]any) {
				if k == "context" {
					m[k] = map[string]any{}
				} else {
					m[k] = []any{}
				}
			})
			refusedGame(t, srv, ts, evidenceInquiry(), k)
		})
	}
}

// R4: an id, or an actor, beginning with game: is refused.
func TestR4AGameIDInTheRecordSet(t *testing.T) {
	for name, edit := range map[string]func(map[string]any){
		"a source's id": func(m map[string]any) { listOf(m, "sources")[0].(map[string]any)["id"] = "game:src" },
		"an actor":      func(m map[string]any) { listOf(m, "actors")[1].(map[string]any)["id"] = "game:owner" },
	} {
		t.Run(name, func(t *testing.T) {
			srv, ts, dir := evidenceServer(t)
			editSet(t, dir, edit)
			refusedGame(t, srv, ts, evidenceInquiry(), "game:")
		})
	}
}

// R5: a record whose author is not one of the record set's actors.
func TestR5AnAuthorNotOfTheRecordSet(t *testing.T) {
	srv, ts, dir := evidenceServer(t)
	editSet(t, dir, func(m map[string]any) { listOf(m, "claims")[1].(map[string]any)["created_by"] = "someone" })
	refusedGame(t, srv, ts, evidenceInquiry(), "someone")
}

// R6: a payload path outside NAME/, a link below it, the NAME folder a link,
// or the evidence folder a link: refused, each.
func TestR6PathsAndLinks(t *testing.T) {
	t.Run("a payload outside", func(t *testing.T) {
		srv, ts, dir := evidenceServer(t)
		if err := os.WriteFile(filepath.Join(dir, "outside.txt"), []byte(fixtureArticle), 0o644); err != nil {
			t.Fatal(err)
		}
		editSet(t, dir, func(m map[string]any) {
			listOf(m, "representations")[2].(map[string]any)["payload"].(map[string]any)["path"] = "../outside.txt"
		})
		refusedGame(t, srv, ts, evidenceInquiry(), "rep-article")
	})
	t.Run("a link below", func(t *testing.T) {
		srv, ts, dir := evidenceServer(t)
		p := filepath.Join(dir, "fixture", "article.txt")
		other := filepath.Join(t.TempDir(), "article.txt")
		if err := os.WriteFile(other, []byte(fixtureArticle), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(p); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(other, p); err != nil {
			t.Skip("no symbolic links here:", err)
		}
		refusedGame(t, srv, ts, evidenceInquiry(), "rep-article")
	})
	t.Run("records.json a link", func(t *testing.T) {
		srv, ts, dir := evidenceServer(t)
		p := filepath.Join(dir, "fixture", "records.json")
		other := filepath.Join(t.TempDir(), "records.json")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(other, b, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(p); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(other, p); err != nil {
			t.Skip("no symbolic links here:", err)
		}
		refusedGame(t, srv, ts, evidenceInquiry(), "link")
	})
	t.Run("the NAME folder a link", func(t *testing.T) {
		srv, ts, dir := evidenceServer(t)
		if err := os.Symlink(filepath.Join(dir, "fixture"), filepath.Join(dir, "linked")); err != nil {
			t.Skip("no symbolic links here:", err)
		}
		req := evidenceInquiry()
		req.Evidence = "linked"
		refusedGame(t, srv, ts, req, "link")
	})
	t.Run("the evidence folder a link", func(t *testing.T) {
		srv, ts, dir := evidenceServer(t)
		link := filepath.Join(t.TempDir(), "evidence")
		if err := os.Symlink(dir, link); err != nil {
			t.Skip("no symbolic links here:", err)
		}
		srv.SetEvidence(link)
		refusedGame(t, srv, ts, evidenceInquiry(), "link")
	})
}

// R7: a NAME with a separator or "..": refused.
func TestR7ANameThatIsNotPlain(t *testing.T) {
	for _, name := range []string{"fixture/", "a/fixture", `a\fixture`, "..", "../fixture", "fix..ture"} {
		srv, ts, _ := evidenceServer(t)
		req := evidenceInquiry()
		req.Evidence = name
		refusedGame(t, srv, ts, req, "plain name")
	}
}

// R8: a debate with the record set; R9: an inquiry with a rule bot seat; R10:
// a moderated inquiry: refused, the game not created.
func TestR8R9R10GamesThatTakeNoRecordSet(t *testing.T) {
	srv, ts, _ := evidenceServer(t)
	req := evidenceInquiry()
	req.Mode = "debate"
	refusedGame(t, srv, ts, req, "inquiry")
	req = evidenceInquiry()
	req.Seats[1] = Seat{Name: "B", Kind: "greedy"}
	refusedGame(t, srv, ts, req, "greedy")
	req = evidenceInquiry()
	req.Seats[0] = Seat{Name: "A", Kind: "random"}
	refusedGame(t, srv, ts, req, "random")
	req = evidenceInquiry()
	req.Moderated = true
	refusedGame(t, srv, ts, req, "moderated")
}

// R11: a server without an evidence folder (no data folder, no --evidence).
func TestR11AServerWithoutAnEvidenceFolder(t *testing.T) {
	srv, ts := newTestServerWith(t)
	refusedGame(t, srv, ts, evidenceInquiry(), "evidence folder")
}

// R12: the record set's JSON with a duplicate key, an unknown field, or a
// null: refused, each.
func TestR12StrictDecoding(t *testing.T) {
	for name, edit := range map[string]func([]byte) []byte{
		"a duplicate key": func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"format": "wag-evidence/0.2"`), []byte(`"format": "wag-evidence/0.2", "format": "wag-evidence/0.2"`), 1)
		},
		"an unknown field": func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"format": "wag-evidence/0.2"`), []byte(`"format": "wag-evidence/0.2", "extra": 1`), 1)
		},
		"a null": func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"label": "the owner"`), []byte(`"label": null`), 1)
		},
	} {
		t.Run(name, func(t *testing.T) {
			srv, ts, dir := evidenceServer(t)
			p := filepath.Join(dir, "fixture", "records.json")
			b, _ := os.ReadFile(p)
			edited := edit(b)
			if bytes.Equal(edited, b) {
				t.Fatal("the edit changed nothing")
			}
			if err := os.WriteFile(p, edited, 0o644); err != nil {
				t.Fatal(err)
			}
			refusedGame(t, srv, ts, evidenceInquiry(), "fixture")
		})
	}
}

// R14: the fixture's game: the context and the actors of its bundle.
// R15: an inquiry with an llm seat: created, and that player a model
// participant.
func TestR14R15TheGamesContextAndActors(t *testing.T) {
	srv, ts, _ := evidenceServer(t)
	req := evidenceInquiry()
	req.Seats[1] = Seat{Name: "B", Kind: "llm"}
	var st State
	if code := call(t, ts, "POST", "/api/games", req, &st); code != http.StatusOK {
		t.Fatalf("R15: %d", code)
	}
	srv.mu.Lock()
	b := srv.sessions[st.ID].GS.Evidence.Bundle
	pins := srv.pins
	srv.mu.Unlock()
	c := b.Context
	if c.ID != "game:context" || c.Profile != "wag-inquiry/1" || c.Policy != evidence.Policy ||
		strings.Join(c.Participants, ",") != "game:player:A,game:player:B" ||
		c.EngineSHA256 != pins.Engine || c.CatalogueSHA256 != pins.Catalogue || c.NASCQSHA256 != pins.NASCQ {
		t.Fatalf("R14: the context: %+v", c)
	}
	actors := map[string]evidence.Actor{}
	for _, a := range b.Actors {
		actors[a.ID] = a
	}
	if a := actors["game:player:A"]; a.Kind != "human" || a.Label != "A" {
		t.Fatalf("R14: A: %+v", a)
	}
	if a := actors["game:player:B"]; a.Kind != "model" || a.Label != "B" {
		t.Fatalf("R15: B, an llm seat: %+v", a)
	}
	if a := actors["game:wag"]; a.Kind != "tool" {
		t.Fatalf("R14: the game: %+v", a)
	}
}
