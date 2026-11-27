package catalogue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

// plain turns a tree read keeping order (yaml.MapSlice) or read by YAML or
// JSON into maps keyed by strings, with numbers as float64, so that trees
// read from JSON and from YAML compare. A key twice is a failure.
func plain(t *testing.T, v interface{}) interface{} {
	t.Helper()
	switch x := v.(type) {
	case yaml.MapSlice:
		m := map[string]interface{}{}
		for _, it := range x {
			k, ok := it.Key.(string)
			if !ok {
				t.Fatalf("key %v is not a string", it.Key)
			}
			if _, dup := m[k]; dup {
				t.Fatalf("key %s twice", k)
			}
			m[k] = plain(t, it.Value)
		}
		return m
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v := range x {
			m[fmt.Sprint(k)] = plain(t, v)
		}
		return m
	case map[string]interface{}:
		m := map[string]interface{}{}
		for k, v := range x {
			m[k] = plain(t, v)
		}
		return m
	case []interface{}:
		out := make([]interface{}, len(x))
		for i, v := range x {
			out[i] = plain(t, v)
		}
		return out
	case int:
		return float64(x)
	case int64:
		return float64(x)
	default:
		return v
	}
}

// jsonTree reads JSON as a plain tree.
func jsonTree(t *testing.T, b []byte) interface{} {
	t.Helper()
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	return plain(t, v)
}

// The game's JSON is read keeping the order of keys, whole numbers apart
// from the others, and nothing after the value.
func TestDecodeOrderedJSON(t *testing.T) {
	v, err := decodeOrderedJSON([]byte(`{"b": 1, "a": [2.5, {"z": true, "y": null}], "c": "x", "d": 1.0}`))
	if err != nil {
		t.Fatal(err)
	}
	want := yaml.MapSlice{
		{Key: "b", Value: 1},
		{Key: "a", Value: []interface{}{2.5, yaml.MapSlice{{Key: "z", Value: true}, {Key: "y", Value: nil}}}},
		{Key: "c", Value: "x"},
		{Key: "d", Value: 1.0},
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("got %#v, want %#v", v, want)
	}
	for _, bad := range []string{`{"a": 1} {}`, `{"a": 1`, `[1,]`, `{"a" 1}`, ``} {
		if _, err := decodeOrderedJSON([]byte(bad)); err == nil {
			t.Errorf("%q: no error", bad)
		}
	}
}

// core.yml's template keeps all but the game schemes' blocks, each of
// which becomes a marker; a game scheme core.yml lacks is an error.
func TestCoreTemplate(t *testing.T) {
	text := "language:\n  a/1: \"%s\"\n\nargument_schemes:\n  # A comment kept.\n  - id: one\n    premises:\n      - a(X)\n" +
		"  - id: kept\n    premises: [a(X)]\n\n  # Another.\n  - id: two\n    meta:\n      title: Two\n"
	got, err := coreTemplate(text, map[string]bool{"one": true, "two": true})
	if err != nil {
		t.Fatal(err)
	}
	want := "language:\n  a/1: \"%s\"\n\nargument_schemes:\n  # A comment kept.\n  # @scheme one\n" +
		"  - id: kept\n    premises: [a(X)]\n\n  # Another.\n  # @scheme two\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	for _, ids := range []map[string]bool{{"three": true}, {"one": true, "three": true}} {
		if _, err := coreTemplate(text, ids); err == nil || !strings.Contains(err.Error(), "three") {
			t.Errorf("%v: got %v, want an error naming three", ids, err)
		}
	}
	// An id line written otherwise is not taken for the scheme's: an error,
	// not a block kept twice.
	if _, err := coreTemplate(strings.Replace(text, "- id: one", "- id: one # x", 1), map[string]bool{"one": true}); err == nil {
		t.Error("an id line with a comment: no error")
	}
}

// An entry records core.yml's block but its id, and but its formulas where
// they are walton.yml's, in the same order.
func TestGameCore(t *testing.T) {
	block := yaml.MapSlice{
		{Key: "id", Value: "x"},
		{Key: "meta", Value: yaml.MapSlice{{Key: "title", Value: "X"}}},
		{Key: "variables", Value: []interface{}{"A", "B"}},
		{Key: "conclusions", Value: []interface{}{"c(A)"}},
		{Key: "premises", Value: []interface{}{"p(A)", "q(B)"}},
		{Key: "exceptions", Value: []interface{}{"e(A)"}},
	}
	w := Scheme{ID: "x", Variables: []string{"A", "B"}, Premises: []string{"p(A)", "q(B)"},
		Exceptions: []string{"e(A)"}, Conclusions: []string{"c(A)"}}
	if got, want := gameCore(block, w), block[1:2]; !reflect.DeepEqual(got, want) {
		t.Errorf("the same formulas: got %v, want %v", got, want)
	}
	for name, edit := range map[string]func(*Scheme){
		"premises in another order": func(s *Scheme) { s.Premises = []string{"q(B)", "p(A)"} },
		"another variable":          func(s *Scheme) { s.Variables = []string{"A", "B", "C"} },
		"an assumption":             func(s *Scheme) { s.Assumptions = []string{"a(A)"} },
		"no exception":              func(s *Scheme) { s.Exceptions = nil },
		"another conclusion":        func(s *Scheme) { s.Conclusions = []string{"d(A)"} },
	} {
		v := w
		edit(&v)
		if got, want := gameCore(block, v), block[1:]; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
}

// Each game scheme is in one entry, which records its NASCQ record as the
// NASCQ has it and its core.yml block (with walton.yml's formulas where it
// leaves them out) as core.yml has it.
func TestGameFieldsRecorded(t *testing.T) {
	s := sources(t)
	f, err := ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	nascq := jsonTree(t, f.Schemes).(map[string]interface{})
	var core struct {
		Schemes []yaml.MapSlice `yaml:"argument_schemes"`
	}
	if err := yaml.Unmarshal(f.Core, &core); err != nil {
		t.Fatal(err)
	}
	blocks := map[string]yaml.MapSlice{}
	for _, b := range core.Schemes {
		blocks[fmt.Sprint(value(b, "id"))] = b
	}
	g := &Generator{Src: s, Today: "2026-10-22", By: "test"}
	seen := map[string]string{}
	for _, e := range g.Build() {
		if e.Game == nil {
			continue
		}
		gid := e.Game.ID
		if other, ok := seen[gid]; ok {
			t.Errorf("%s is in %s and %s", gid, other, e.ID)
		}
		seen[gid] = e.ID
		if e.Game.Status != Generated {
			t.Errorf("%s: status %q, want %q", e.ID, e.Game.Status, Generated)
		}
		if got, want := plain(t, e.Game.NASCQ), nascq[gid]; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: the NASCQ record differs:\n got %v\nwant %v", e.ID, got, want)
		}
		cid := s.Game[gid].CarneadesID
		w, ok := s.scheme(cid)
		if !ok {
			t.Fatalf("%s: no walton.yml scheme %s", e.ID, cid)
		}
		rebuilt := yaml.MapSlice{{Key: "id", Value: cid}}
		rebuilt = append(rebuilt, e.Game.Core...)
		if value(e.Game.Core, "premises") == nil && value(e.Game.Core, "conclusions") == nil {
			for _, r := range []struct {
				key  string
				have []string
			}{{"variables", w.Variables}, {"premises", w.Premises}, {"assumptions", w.Assumptions},
				{"exceptions", w.Exceptions}, {"conclusions", w.Conclusions}} {
				if len(r.have) > 0 {
					l := make([]interface{}, len(r.have))
					for i, x := range r.have {
						l[i] = x
					}
					rebuilt = append(rebuilt, yaml.MapItem{Key: r.key, Value: l})
				}
			}
		}
		if got, want := plain(t, rebuilt), plain(t, blocks[cid]); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: core.yml's block differs:\n got %v\nwant %v", e.ID, got, want)
		}
	}
	if len(seen) != len(s.Game) || len(seen) != len(nascq) {
		t.Errorf("%d game schemes in entries, %d in the NASCQ", len(seen), len(s.Game))
	}
}

// GameDir's files hold the NASCQ's questions as it has them, its schemes
// and edges in its order, and core.yml but the game schemes' blocks.
func TestGameFiles(t *testing.T) {
	s := sources(t)
	f, err := ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	files, err := (&Generator{Src: s}).GameFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Errorf("%d files, want 3", len(files))
	}
	var q struct {
		Questions []struct {
			ID     string        `yaml:"id"`
			Record yaml.MapSlice `yaml:"record"`
		} `yaml:"questions"`
	}
	if err := yaml.Unmarshal([]byte(files[GameQuestionsFile]), &q); err != nil {
		t.Fatal(err)
	}
	got := map[string]interface{}{}
	var ids []string
	for _, x := range q.Questions {
		got[x.ID] = plain(t, x.Record)
		ids = append(ids, x.ID)
	}
	// questions.yml is critical_questions.json but for what the generator
	// makes: each question's text and caes_mapping (rules (b) and (c)).
	want := jsonTree(t, f.CQs)
	for _, rec := range want.(map[string]interface{}) {
		delete(rec.(map[string]interface{}), "question_template_en")
		delete(rec.(map[string]interface{}), "caes_mapping")
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("questions.yml is not critical_questions.json without the made text and caes_mapping")
	}
	if want := orderOfKeys(t, f.CQs); !reflect.DeepEqual(ids, want) {
		t.Errorf("questions.yml's order %v, want critical_questions.json's %v", ids, want)
	}
	var o GameOrder
	if err := yaml.UnmarshalStrict([]byte(files[GameOrderFile]), &o); err != nil {
		t.Fatal(err)
	}
	if want := orderOfKeys(t, f.Schemes); !reflect.DeepEqual(o.Schemes, want) {
		t.Errorf("order.yml's schemes %v, want schemes.json's %v", o.Schemes, want)
	}
	var edges struct {
		Ask, Backup []GameEdge
	}
	var raw struct {
		Ask    []map[string]string `json:"ask_edges"`
		Backup []map[string]string `json:"backup_edges"`
	}
	if err := json.Unmarshal(f.Edges, &raw); err != nil {
		t.Fatal(err)
	}
	for _, e := range raw.Ask {
		edges.Ask = append(edges.Ask, GameEdge{CQ: e["cq"], Scheme: e["scheme"]})
	}
	for _, e := range raw.Backup {
		edges.Backup = append(edges.Backup, GameEdge{CQ: e["cq"], Scheme: e["scheme"]})
	}
	if !reflect.DeepEqual(o.Ask, edges.Ask) || !reflect.DeepEqual(o.Backup, edges.Backup) {
		t.Errorf("order.yml's edges are not edges.json's, in its order")
	}
	tpl := files[GameCoreTemplate]
	for id := range s.Game {
		cid := s.Game[id].CarneadesID
		if strings.Count(tpl, coreMarker+cid+"\n") != 1 {
			t.Errorf("core.yml.in: no one marker for %s", cid)
		}
		if strings.Contains(tpl, "- id: "+cid+"\n") {
			t.Errorf("core.yml.in keeps %s's block", cid)
		}
	}
	var rest struct {
		Language map[string]string `yaml:"language"`
		Schemes  []yaml.MapSlice   `yaml:"argument_schemes"`
	}
	if err := yaml.Unmarshal([]byte(tpl), &rest); err != nil {
		t.Fatalf("core.yml.in is not YAML: %v", err)
	}
	var whole struct {
		Language map[string]string `yaml:"language"`
		Schemes  []yaml.MapSlice   `yaml:"argument_schemes"`
	}
	if err := yaml.Unmarshal(f.Core, &whole); err != nil {
		t.Fatal(err)
	}
	// The template's language writes no sentence: its marker stands where
	// core.yml's sentences are made (rule (e)).
	if len(rest.Language) != 0 || strings.Count(tpl, languageMarker+"\n") != 1 || len(whole.Language) == 0 {
		t.Errorf("core.yml.in's language: %d sentences, %d markers; core.yml's: %d sentences",
			len(rest.Language), strings.Count(tpl, languageMarker+"\n"), len(whole.Language))
	}
	if len(rest.Schemes)+len(s.Game) != len(whole.Schemes) {
		t.Errorf("core.yml.in keeps %d schemes; core.yml has %d, the game plays %d", len(rest.Schemes), len(whole.Schemes), len(s.Game))
	}
}

// orderOfKeys lists a JSON object's keys in their order.
func orderOfKeys(t *testing.T, b []byte) []string {
	t.Helper()
	v, err := decodeOrderedJSON(b)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, it := range v.(yaml.MapSlice) {
		out = append(out, it.Key.(string))
	}
	return out
}

// The check requires GameDir's files to be what the generator makes, and
// the entries' game fields to be the NASCQ's.
func TestCheckFindsGameFaults(t *testing.T) {
	s := sources(t)
	for _, c := range []struct {
		name, file string
		edit       func(t *testing.T, dir string)
	}{
		{"order changed", GameDir + "/" + GameOrderFile, func(t *testing.T, dir string) {
			p := filepath.Join(dir, GameDir, GameOrderFile)
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			n := strings.Replace(string(b), "- AS:values\n", "", 1)
			n = strings.Replace(n, "- AS:sign\n", "- AS:sign\n- AS:values\n", 1)
			if n == string(b) {
				t.Fatal("the edit changed nothing")
			}
			if err := os.WriteFile(p, []byte(n), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"questions missing", GameDir + "/" + GameQuestionsFile, func(t *testing.T, dir string) {
			if err := os.Remove(filepath.Join(dir, GameDir, GameQuestionsFile)); err != nil {
				t.Fatal(err)
			}
		}},
		{"template edited", GameDir + "/" + GameCoreTemplate, func(t *testing.T, dir string) {
			p := filepath.Join(dir, GameDir, GameCoreTemplate)
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, append(b, "# a comment\n"...), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"an entry's NASCQ record", "positive_value.yml", func(t *testing.T, dir string) {
			e := readEntry(t, dir, "positive_value")
			for i, it := range e.Game.NASCQ {
				if it.Key == "base_weight" {
					e.Game.NASCQ[i].Value = 0.61
				}
			}
			writeEntry(t, dir, e)
		}},
		{"an entry's core block", "sign.yml", func(t *testing.T, dir string) {
			e := readEntry(t, dir, "sign")
			e.Game.Core = e.Game.Core[1:]
			writeEntry(t, dir, e)
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := generate(t, s)
			c.edit(t, dir)
			r, err := Check(dir, s, DefaultOptions())
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, f := range r.Findings {
				if f.Error && f.File == c.file {
					found = true
				}
			}
			if !found {
				t.Errorf("no error on %s; errors: %v", c.file, errorsOf(r))
			}
			// Generating again puts it right.
			g := &Generator{Src: s, Today: "2026-10-22", By: "test"}
			if _, err := g.Write(dir); err != nil {
				t.Fatal(err)
			}
			if r, err = Check(dir, s, DefaultOptions()); err != nil {
				t.Fatal(err)
			}
			if errs := errorsOf(r); len(errs) > 0 {
				t.Errorf("generated again, the check still finds: %v", errs)
			}
		})
	}
}
