package server

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/game"
)

// Follow-ups to the replay-identity patch (verify/r-critical-review-2026-11-23,
// report A). They go through the HTTP API, so that they also run on the patch
// as received, where each of them fails.

func builtinData() Data {
	return Data{YAML: data.CoreYAML, Schemes: data.SchemesJSON, CQs: data.CQsJSON, Edges: data.EdgesJSON, Contexts: data.ContextsJSON}
}

// windows is the data as a Windows checkout has it (Git's default turns LF
// into CRLF).
func windows(d Data) Data {
	conv := func(b []byte) []byte {
		return bytes.ReplaceAll(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")), []byte("\n"), []byte("\r\n"))
	}
	return Data{YAML: conv(d.YAML), Schemes: conv(d.Schemes), CQs: conv(d.CQs), Edges: conv(d.Edges), Contexts: conv(d.Contexts)}
}

// rawHash is the fingerprint as it was before it read CRLF as LF: the games
// saved by a build from a Windows checkout recorded it.
func rawHash(d Data) string {
	h := sha256.New()
	for _, b := range [][]byte{d.YAML, d.Schemes, d.CQs, d.Edges} {
		h.Write(b)
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func loadHTTP(t *testing.T, ts *httptest.Server, e *Export, query string) (int, map[string]any) {
	t.Helper()
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	code := call(t, ts, "POST", "/api/load"+query, string(b), &out)
	return code, out
}

func weightsOf(t *testing.T, st map[string]any) map[string]float64 {
	t.Helper()
	cfg, _ := st["config"].(map[string]any)
	ws, _ := cfg["weights"].(map[string]any)
	out := map[string]float64{}
	for k, v := range ws {
		out[k], _ = v.(float64)
	}
	return out
}

var twoHumans = []Seat{{Name: "Ann", Kind: "human"}, {Name: "Bob", Kind: "human"}}

// The same data in a Windows checkout is the same data: the YAML and JSON
// readers read CRLF as LF, and so does its fingerprint. A game saved on one
// machine loads on the other, and so does one saved by a Windows build
// before the fingerprint read CRLF as LF.
func TestFollowupLineEndingsDoNotChangeTheData(t *testing.T) {
	win := windows(builtinData())
	if bytes.Equal(win.YAML, data.CoreYAML) || rawHash(win) == rawHash(builtinData()) {
		t.Fatal("fixture: no line ending changed")
	}
	sw, err := New(win, "")
	if err != nil {
		t.Fatal(err)
	}
	tw := httptest.NewServer(sw.Handler())
	defer tw.Close()
	sl, tl := newTestServerWith(t)
	meta := func(ts *httptest.Server) string {
		resp, err := http.Get(ts.URL + "/api/meta")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return string(b)
	}
	if meta(tw) != meta(tl) {
		t.Fatal("the schemes read from CRLF files differ from the LF ones")
	}
	fromWindows := reviewExport(t, sw, NewGameRequest{Mode: "classic", Context: "scientific", Seats: twoHumans})
	fromLinux := reviewExport(t, sl, NewGameRequest{Mode: "classic", Context: "scientific", Seats: twoHumans})
	before := *fromWindows
	before.DataHash = rawHash(win)
	for _, tc := range []struct {
		name string
		e    *Export
		on   *httptest.Server
	}{
		{"saved on Windows, loaded on Linux", fromWindows, tl},
		{"saved on Linux, loaded on Windows", fromLinux, tw},
		{"saved on Windows before 24 November, loaded on Linux", &before, tl},
		{"saved on Windows before 24 November, loaded on Windows", &before, tw},
	} {
		code, out := loadHTTP(t, tc.on, tc.e, "")
		if code != http.StatusOK || out["inspect"] != nil {
			t.Errorf("%s: %d %v", tc.name, code, out["error"])
		}
	}
}

// An export from the builds of 13 October before 17:37 UTC records the
// request's context but neither the weights nor the data: it loads with the
// preset as it is, as before the patch.
func TestFollowupThirteenOctoberExportsLoad(t *testing.T) {
	s, ts := newTestServerWith(t)
	e := reviewExport(t, s, NewGameRequest{Mode: "classic", Context: "scientific", Seats: twoHumans})
	preset := maps.Clone(e.Config.Weights)
	e.Config.Weights, e.DataHash = nil, ""
	code, out := loadHTTP(t, ts, e, "")
	if code != http.StatusOK {
		t.Fatalf("%d %v", code, out["error"])
	}
	if w := weightsOf(t, out); !maps.Equal(w, preset) {
		t.Fatalf("loaded with %v, not the preset %v", w, preset)
	}
}

// A preset whose weights changed since, by value only, or which is gone: the
// game is replayed with the weights it recorded.
func TestFollowupRecordedWeightsOutlastThePreset(t *testing.T) {
	for _, gone := range []bool{false, true} {
		s, ts := newTestServerWith(t)
		e := reviewExport(t, s, NewGameRequest{Mode: "classic", Context: "scientific", Seats: twoHumans})
		played := maps.Clone(e.Config.Weights)
		if gone {
			var rest []game.Context
			for _, c := range s.contexts {
				if c.ID != "scientific" {
					rest = append(rest, c)
				}
			}
			s.contexts = rest
		} else {
			w := maps.Clone(played)
			w["AS:analogy"] = 0.8
			s.AddContexts([]game.Context{{ID: "scientific", Weights: w}})
		}
		code, out := loadHTTP(t, ts, e, "")
		if code != http.StatusOK {
			t.Fatalf("preset gone %v: %d %v", gone, code, out["error"])
		}
		if w := weightsOf(t, out); !maps.Equal(w, played) {
			t.Fatalf("preset gone %v: loaded with %v, not %v", gone, w, played)
		}
	}
}

// The recorded weights are checked as a request's are.
func TestFollowupRecordedWeightsAreChecked(t *testing.T) {
	s, ts := newTestServerWith(t)
	e := reviewExport(t, s, NewGameRequest{Mode: "classic", Context: "scientific", Seats: twoHumans})
	e.Config.Weights["AS:analogy"] = 2
	code, out := loadHTTP(t, ts, e, "")
	if msg, _ := out["error"].(string); code != http.StatusBadRequest || !strings.Contains(msg, "AS:analogy weighs 2") {
		t.Fatalf("%d %v", code, out)
	}
}

// A game played with other data is refused, with an offer to open it
// read-only; open so, it shows the game, takes notes and previews but no
// move, and its export and trace keep the data it was played with.
func TestFollowupOtherDataOpensReadOnly(t *testing.T) {
	s, ts := newTestServerWith(t)
	e := reviewExport(t, s, NewGameRequest{Mode: "classic", Seats: twoHumans})
	const other = "0123456789abcdef"
	e.DataHash = other
	n := len(s.sessions)
	code, out := loadHTTP(t, ts, e, "")
	if code != http.StatusBadRequest || out["inspect"] != true || !strings.Contains(out["error"].(string), "read-only") || len(s.sessions) != n {
		t.Fatalf("refusal: %d %v (%d sessions, not %d)", code, out, len(s.sessions), n)
	}
	code, st := loadHTTP(t, ts, e, "?inspect=1")
	if code != http.StatusOK {
		t.Fatalf("inspect: %d %v", code, st["error"])
	}
	in, _ := st["inspect"].(map[string]any)
	if in["played_with"] != other || in["replayed_with"] != s.data.Hash() || st["can_move"] != false {
		t.Fatalf("inspect: %v, can_move %v", in, st["can_move"])
	}
	id := st["id"].(string)
	for _, action := range []string{"move", "reply", "ai", "moderate", "decide"} {
		var res map[string]any
		code := call(t, ts, "POST", "/api/games/"+id+"/"+action, `{"action": {"type": "abstain"}, "reply": "{}", "approve": true, "statement": "x", "decision": "accepted"}`, &res)
		if code != http.StatusConflict || !strings.Contains(res["error"].(string), "read-only") {
			t.Errorf("%s: %d %v", action, code, res)
		}
	}
	if code := call(t, ts, "POST", "/api/games/"+id+"/note", map[string]string{"text": "looks different", "by": "tester"}, nil); code != http.StatusOK {
		t.Errorf("note: %d", code)
	}
	if code := call(t, ts, "POST", "/api/games/"+id+"/preview", map[string]any{"action": map[string]string{"type": "abstain"}}, nil); code != http.StatusOK {
		t.Errorf("preview: %d", code)
	}
	var ex map[string]any
	if code := call(t, ts, "GET", "/api/games/"+id+"/export", nil, &ex); code != http.StatusOK || ex["data_sha256"] != other || ex["inspected_with_data_sha256"] != s.data.Hash() {
		t.Fatalf("export: %d %v %v", code, ex["data_sha256"], ex["inspected_with_data_sha256"])
	}
	again, _ := json.Marshal(ex)
	var e2 Export
	if err := json.Unmarshal(again, &e2); err != nil {
		t.Fatal(err)
	}
	if code, out := loadHTTP(t, ts, &e2, ""); code != http.StatusBadRequest || out["inspect"] != true {
		t.Fatalf("its export loaded as if played with this data: %d %v", code, out)
	}
	resp, err := http.Get(ts.URL + "/api/games/" + id + "/traces")
	if err != nil {
		t.Fatal(err)
	}
	first, _ := bufio.NewReader(resp.Body).ReadString('\n')
	resp.Body.Close()
	var start map[string]any
	if err := json.Unmarshal([]byte(first), &start); err != nil || start["data_sha256"] != other || start["inspected_with_data_sha256"] != s.data.Hash() {
		t.Fatalf("trace start: %v %v", err, start)
	}
	// With the server's own data, inspect changes nothing: the game goes on.
	e.DataHash = s.data.Hash()
	code, st = loadHTTP(t, ts, e, "?inspect=1")
	if code != http.StatusOK || st["inspect"] != nil || st["can_move"] != true {
		t.Fatalf("same data: %d %v %v", code, st["inspect"], st["can_move"])
	}
}

// A game recorded before 2026-11-27 with a value that names nothing (a field
// typed into and cleared on the board was sent as "") no longer replays: S2
// refuses it since step C5.2's item F. The refusal says so and names the
// remedy, as a retired bridge's does (the independent check of step 7, W-2).
func TestNamelessBindingOfAnOldRecord(t *testing.T) {
	for _, v := range []string{"", "  ", "?"} {
		s, ts := newTestServerWith(t)
		e := reviewExport(t, s, NewGameRequest{Mode: "classic", Seats: twoHumans})
		if len(e.History) == 0 || len(e.History[0].ArgIDs) == 0 {
			t.Fatalf("the recorded game has no played move: %+v", e.History)
		}
		e.History[0].Action.VariableBindings["C"] = v
		code, out := loadHTTP(t, ts, e, "")
		msg, _ := out["error"].(string)
		if code == http.StatusOK || !strings.Contains(msg, "names nothing") || !strings.Contains(msg, "a WAG build from before that date can") {
			t.Fatalf("C=%q: %d %q", v, code, msg)
		}
	}
}
