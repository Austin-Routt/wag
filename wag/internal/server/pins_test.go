package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/manifest"
)

// E1.1 (docs/e1-evidence.md): a game's trace's start and its export carry
// the server's three pins (P2), which follow the data (P3).

func startAndExportPins(t *testing.T, d Data) (start, export manifest.Pins) {
	t.Helper()
	s, err := New(d, "")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	var st State
	if code := call(t, ts, "POST", "/api/games", NewGameRequest{Mode: "classic", Seats: twoHumans}, &st); code != http.StatusOK {
		t.Fatalf("new game: %d", code)
	}
	get := func(what string) []byte {
		resp, err := direct.Get(ts.URL + "/api/games/" + st.ID + "/" + what)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return b
	}
	var g struct {
		Pins *manifest.Pins `json:"pins"`
	}
	first, _, _ := bytes.Cut(get("traces"), []byte("\n"))
	if err := json.Unmarshal(first, &g); err != nil || g.Pins == nil {
		t.Fatalf("the trace's start has no pins: %v %s", err, first)
	}
	var e struct {
		Pins manifest.Pins `json:"pins"`
	}
	if err := json.Unmarshal(get("export"), &e); err != nil {
		t.Fatal(err)
	}
	return *g.Pins, e.Pins
}

// P2: each pin is 64 hexadecimal digits, the engine's the build's, the
// others the data's; the start and the export agree.
func TestP2AGameRecordsItsPins(t *testing.T) {
	d := builtinData()
	start, export := startAndExportPins(t, d)
	hex64 := regexp.MustCompile(`^[0-9a-f]{64}$`)
	for _, v := range []string{start.Engine, start.Catalogue, start.NASCQ} {
		if !hex64.MatchString(v) {
			t.Fatalf("pin %q", v)
		}
	}
	if start != export || start.Engine != manifest.EngineSHA256 || start != manifest.Of(d.YAML, d.Schemes, d.CQs, d.Edges) {
		t.Fatalf("start %+v, export %+v", start, export)
	}
}

// P3: another scheme file or other NASCQ data changes its pin, and the
// other stays; a CRLF checkout of the same data changes none.
func TestP3PinsFollowTheData(t *testing.T) {
	base, _ := startAndExportPins(t, builtinData())
	if crlf, _ := startAndExportPins(t, windows(builtinData())); crlf != base {
		t.Fatalf("a CRLF checkout changed the pins: %+v, %+v", base, crlf)
	}
	yaml := builtinData()
	yaml.YAML = append([]byte("# another scheme file\n"), yaml.YAML...)
	if p, _ := startAndExportPins(t, yaml); p.Catalogue == base.Catalogue || p.NASCQ != base.NASCQ || p.Engine != base.Engine {
		t.Fatalf("another scheme file: %+v, %+v", base, p)
	}
	edges := builtinData()
	edges.Edges = []byte(strings.Replace(string(edges.Edges), "{", "{ ", 1))
	if p, _ := startAndExportPins(t, edges); p.NASCQ == base.NASCQ || p.Catalogue != base.Catalogue {
		t.Fatalf("other edges: %+v, %+v", base, p)
	}
}
