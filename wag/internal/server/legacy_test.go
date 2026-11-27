package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// The critical review's adverse case 11 (docs/c5-game.md): a legacy record
// (an export that names no scheme data) stays legacy. Continued, its export
// names the data its later moves were played with, and a load checks that
// data as it checks any game's.

// legacyExport reads the export made before 13 October 2026 that the tests
// keep, as JSON to edit.
func legacyExport(t *testing.T) map[string]any {
	t.Helper()
	b, err := os.ReadFile("testdata/user-game-4.json")
	if err != nil {
		t.Fatal(err)
	}
	var e map[string]any
	if err := json.Unmarshal(b, &e); err != nil {
		t.Fatal(err)
	}
	if e["data_sha256"] != nil {
		t.Fatal("the fixture names scheme data")
	}
	return e
}

// loadJSON loads an export given as JSON.
func loadJSON(t *testing.T, ts *httptest.Server, e map[string]any, query string) (int, map[string]any) {
	t.Helper()
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	code := call(t, ts, "POST", "/api/load"+query, string(b), &out)
	return code, out
}

func TestLegacyRecordContinued(t *testing.T) {
	s, ts := newPrevServer(t) // the record was played on data the switch of step C5.2 changed
	here := s.data.Hash()
	e := legacyExport(t)
	n := len(e["history"].([]any))
	code, st := loadJSON(t, ts, e, "")
	if code != http.StatusOK {
		t.Fatalf("load: %d %v", code, st["error"])
	}
	id := st["id"].(string)
	if l, _ := st["legacy"].(map[string]any); l == nil || l["replayed_with"] != here || l["moves"] != float64(n) {
		t.Fatalf("legacy: %v", st["legacy"])
	}
	// Not continued: its export names no data, nor a continuation.
	var ex map[string]any
	call(t, ts, "GET", "/api/games/"+id+"/export", nil, &ex)
	if ex["data_sha256"] != "" || ex["legacy_moves"] != nil || ex["continued_with_data_sha256"] != nil {
		t.Fatalf("export of the record: data %v, legacy_moves %v, continued %v", ex["data_sha256"], ex["legacy_moves"], ex["continued_with_data_sha256"])
	}
	// Continued: one move with this data.
	if code := call(t, ts, "POST", "/api/games/"+id+"/move", map[string]any{"action": map[string]string{"type": "abstain"}}, nil); code != http.StatusOK {
		t.Fatalf("move: %d", code)
	}
	call(t, ts, "GET", "/api/games/"+id+"/export", nil, &ex)
	if ex["data_sha256"] != "" || ex["legacy_replayed_with_data_sha256"] != here || ex["legacy_moves"] != float64(n) || ex["continued_with_data_sha256"] != here {
		t.Fatalf("export of the continued record: data %v, replayed %v, legacy_moves %v, continued %v",
			ex["data_sha256"], ex["legacy_replayed_with_data_sha256"], ex["legacy_moves"], ex["continued_with_data_sha256"])
	}
	// Loaded again with this data: a legacy record still, its continuation
	// named again.
	code, st = loadJSON(t, ts, ex, "")
	if l, _ := st["legacy"].(map[string]any); code != http.StatusOK || l == nil || l["moves"] != float64(n) {
		t.Fatalf("load of the continued record: %d %v", code, st)
	}
	var ex2 map[string]any
	call(t, ts, "GET", "/api/games/"+st["id"].(string)+"/export", nil, &ex2)
	if ex2["legacy_moves"] != float64(n) || ex2["continued_with_data_sha256"] != here {
		t.Fatalf("export again: legacy_moves %v, continued %v", ex2["legacy_moves"], ex2["continued_with_data_sha256"])
	}
	// A legacy record's export records the weights in effect, as a current
	// export does: a load replays it with them.
	ex["config"].(map[string]any)["weights"] = map[string]any{"AS:sign": 0.123}
	code, st = loadJSON(t, ts, ex, "")
	if ws, _ := st["config"].(map[string]any)["weights"].(map[string]any); code != http.StatusOK || ws["AS:sign"] != 0.123 {
		t.Fatalf("a legacy record's recorded weights: %d %v", code, st["config"])
	}
	delete(ex["config"].(map[string]any), "weights")
	// Its continuation played with other data: refused, or opened read-only,
	// as any game played with other data; its export keeps what it names.
	const other = "0123456789abcdef"
	ex["continued_with_data_sha256"] = other
	if code, out := loadJSON(t, ts, ex, ""); code != http.StatusBadRequest || out["inspect"] != true {
		t.Fatalf("continued with other data: %d %v", code, out)
	}
	code, st = loadJSON(t, ts, ex, "?inspect=1")
	if code != http.StatusOK || st["inspect"] == nil || st["legacy"] == nil || st["can_move"] != false {
		t.Fatalf("read-only: %d inspect %v legacy %v can_move %v", code, st["inspect"], st["legacy"], st["can_move"])
	}
	call(t, ts, "GET", "/api/games/"+st["id"].(string)+"/export", nil, &ex2)
	if ex2["data_sha256"] != "" || ex2["continued_with_data_sha256"] != other || ex2["legacy_moves"] != float64(n) || ex2["inspected_with_data_sha256"] != here {
		t.Fatalf("read-only export: data %v, continued %v, legacy_moves %v, inspected %v",
			ex2["data_sha256"], ex2["continued_with_data_sha256"], ex2["legacy_moves"], ex2["inspected_with_data_sha256"])
	}
	// Malformed continuations are refused.
	for name, edit := range map[string]func(map[string]any){
		"legacy_moves past the history":     func(x map[string]any) { x["legacy_moves"] = n + 5 },
		"legacy_moves below zero":           func(x map[string]any) { x["legacy_moves"] = -1 },
		"a continuation without its data":   func(x map[string]any) { delete(x, "continued_with_data_sha256") },
		"data named and a continuation too": func(x map[string]any) { x["data_sha256"] = here },
		"a continuation's data, no position": func(x map[string]any) {
			delete(x, "legacy_moves")
		},
	} {
		x := map[string]any{}
		for k, v := range ex {
			x[k] = v
		}
		x["continued_with_data_sha256"] = here
		edit(x)
		if code, out := loadJSON(t, ts, x, ""); code != http.StatusBadRequest || out["inspect"] == true {
			t.Errorf("%s: %d %v", name, code, out)
		}
	}
}

// Opened with ?inspect=1, an export that names no data is a legacy record
// too (it was not played with other data, so it is not read-only); a legacy
// record's trace says what it was replayed with.
func TestLegacyOnTheReadOnlyPathAndInTraces(t *testing.T) {
	s, ts := newPrevServer(t) // the record was played on data the switch of step C5.2 changed
	here := s.data.Hash()
	code, st := loadJSON(t, ts, legacyExport(t), "?inspect=1")
	if l, _ := st["legacy"].(map[string]any); code != http.StatusOK || l == nil || l["replayed_with"] != here || st["inspect"] != nil {
		t.Fatalf("?inspect=1: %d legacy %v inspect %v", code, st["legacy"], st["inspect"])
	}
	g := s.gameStart(s.sessions[st["id"].(string)])
	if g.LegacyReplayedWith != here || g.DataSHA256 != "" {
		t.Errorf("the legacy record's trace: data %q, legacy replayed with %q", g.DataSHA256, g.LegacyReplayedWith)
	}
	ex := reviewExport(t, s, NewGameRequest{Mode: "classic", Seats: twoHumans})
	ss, err := s.load(ex, false)
	if err != nil {
		t.Fatal(err)
	}
	if g := s.gameStart(ss); g.LegacyReplayedWith != "" || g.DataSHA256 != here {
		t.Errorf("a current game's trace: data %q, legacy %q", g.DataSHA256, g.LegacyReplayedWith)
	}
	if !strings.Contains(string(mustJSON(t, s.gameStart(ss))), `"data_sha256":"`+here+`"`) {
		t.Error("a current game's trace does not name its data")
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
