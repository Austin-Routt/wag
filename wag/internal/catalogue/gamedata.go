package catalogue

// The game's data in the catalogue (step C5.1, docs/c5-game.md): the fields
// of the NASCQ that no text defines (plan §6), recorded in the catalogue,
// which the game's files are generated from since C5.1's switch. They are
// read back from the generated files, so that the check sees the round trip
// keep them.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v2"
)

// GameDir holds the game's data that no single entry holds: the questions,
// which several schemes share, the order of the edges, which is behaviour
// (the order of the offered moves), and core.yml's template.
const GameDir = "game"

// The files of GameDir (and GameLanguageFile).
const (
	GameQuestionsFile = "questions.yml"
	GameOrderFile     = "order.yml"
	GameCoreTemplate  = "core.yml.in"
)

// coreMarker marks, in core.yml's template, where a game scheme's block goes.
const coreMarker = "  # @scheme "

// GameQuestion is a question of the NASCQ (critical_questions.json): its id
// and its record, as written there.
type GameQuestion struct {
	ID     string        `yaml:"id"`
	Record yaml.MapSlice `yaml:"record"`
}

// GameEdge is an edge of the NASCQ (edges.json): a question asked of a
// scheme, or a scheme that backs the answer to a question.
type GameEdge struct {
	CQ     string `yaml:"cq"`
	Scheme string `yaml:"scheme"`
	// Across names the decided decision that admits a backup edge whose
	// scheme answers a question of another semantic type (step C5.3,
	// D009): the lint takes it, and the catalogue check that the decision
	// is decided. Ask edges have none.
	Across string `yaml:"across,omitempty"`
}

// GameOrder is the order of the NASCQ's schemes and edges. The order of the
// edges fixes the order of the offered moves, and so the bots' moves and
// the golden games; the locus and the pattern of an ask edge are the game
// part of the scheme's entry's question.
type GameOrder struct {
	Schemes []string   `yaml:"schemes"`
	Ask     []GameEdge `yaml:"ask_edges"`
	Backup  []GameEdge `yaml:"backup_edges"`
}

// GameData is what GameDir holds.
type GameData struct {
	Questions []GameQuestion
	Order     GameOrder
	// CoreTemplate is core.yml with each game scheme's block replaced by
	// coreMarker and its id, and its language's sentences by
	// languageMarker: its comments and the schemes it keeps for old games
	// stay as written.
	CoreTemplate string
	// Language is game/language.yml: the game's sentences that differ from
	// walton.yml's, with their reasons, and the negations' sentences. The
	// game's files do not give it back, so --catalogue-gen keeps it as
	// written.
	Language GameLanguage
}

// gameRaw is the game's files read in order (Parse).
type gameRaw struct {
	schemes   []GameQuestion // reused shape: an id and its record
	questions []GameQuestion
	ask       []GameEdge
	backup    []GameEdge
	core      map[string]yaml.MapSlice // core.yml's schemes, by id
	coreText  string
}

// parseGameRaw reads the game's files keeping their order.
func parseGameRaw(f Files) (*gameRaw, error) {
	g := &gameRaw{coreText: string(lfOf(f.Core))} // a template is made of LF lines
	recs := func(b []byte, what string) ([]GameQuestion, error) {
		v, err := decodeOrderedJSON(b)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", what, err)
		}
		m, ok := v.(yaml.MapSlice)
		if !ok {
			return nil, fmt.Errorf("%s: not an object", what)
		}
		var out []GameQuestion
		for _, it := range m {
			r, ok := it.Value.(yaml.MapSlice)
			if !ok {
				return nil, fmt.Errorf("%s: %v is not an object", what, it.Key)
			}
			out = append(out, GameQuestion{ID: it.Key.(string), Record: r})
		}
		return out, nil
	}
	var err error
	if g.schemes, err = recs(f.Schemes, "nascq/schemes.json"); err != nil {
		return nil, err
	}
	if g.questions, err = recs(f.CQs, "nascq/critical_questions.json"); err != nil {
		return nil, err
	}
	var edges struct {
		Ask    []map[string]string `json:"ask_edges"`
		Backup []map[string]string `json:"backup_edges"`
	}
	if err := json.Unmarshal(f.Edges, &edges); err != nil {
		return nil, fmt.Errorf("nascq/edges.json: %v", err)
	}
	for _, e := range edges.Ask {
		g.ask = append(g.ask, GameEdge{CQ: e["cq"], Scheme: e["scheme"]})
	}
	for _, e := range edges.Backup {
		g.backup = append(g.backup, GameEdge{CQ: e["cq"], Scheme: e["scheme"], Across: e["across"]})
	}
	var core struct {
		Schemes []yaml.MapSlice `yaml:"argument_schemes"`
	}
	if err := yaml.Unmarshal(f.Core, &core); err != nil {
		return nil, fmt.Errorf("core.yml: %v", err)
	}
	g.core = map[string]yaml.MapSlice{}
	for _, s := range core.Schemes {
		if id, ok := value(s, "id").(string); ok {
			g.core[id] = s
		}
	}
	return g, nil
}

// value returns the value of a key of a mapping, or nil.
func value(m yaml.MapSlice, key string) interface{} {
	for _, it := range m {
		if k, ok := it.Key.(string); ok && k == key {
			return it.Value
		}
	}
	return nil
}

// gameData is GameDir's content as the game's files give it.
func (g *gameRaw) gameData(gameIDs map[string]bool) (*GameData, error) {
	d := &GameData{}
	for _, q := range g.questions {
		var rec yaml.MapSlice
		for _, it := range q.Record {
			if k, _ := it.Key.(string); k != "question_template_en" && k != "caes_mapping" {
				rec = append(rec, it) // the text and the effect are made (rules (b), (c))
			}
		}
		d.Questions = append(d.Questions, GameQuestion{ID: q.ID, Record: rec})
	}
	for _, s := range g.schemes {
		d.Order.Schemes = append(d.Order.Schemes, s.ID)
	}
	d.Order.Ask = g.ask
	d.Order.Backup = g.backup
	t, err := coreTemplate(languageTemplate(stripCoreHeader(g.coreText)), gameIDs)
	if err != nil {
		return nil, err
	}
	d.CoreTemplate = t
	return d, nil
}

// coreTemplate replaces, in core.yml's text, the block of each scheme the
// NASCQ plays (by its Carneades id) with a marker: the "  - id:" line and
// the lines indented under it. Comments, the language and the schemes no
// game scheme plays stay as written.
func coreTemplate(text string, ids map[string]bool) (string, error) {
	lines := strings.SplitAfter(text, "\n")
	var b strings.Builder
	found := map[string]bool{}
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		id, ok := strings.CutPrefix(strings.TrimRight(l, "\n"), "  - id: ")
		if !ok || !ids[strings.TrimSpace(id)] {
			b.WriteString(l)
			continue
		}
		id = strings.TrimSpace(id)
		found[id] = true
		b.WriteString(coreMarker + id + "\n")
		// The block is the lines indented under it, but the comments that
		// end it: the generator writes none in a block, so those are the
		// template's own, written under the marker.
		end := i
		for end+1 < len(lines) && strings.HasPrefix(lines[end+1], "    ") {
			end++
		}
		for end > i && comment(lines[end]) {
			end--
		}
		i = end
	}
	for id := range ids {
		if !found[id] {
			return "", fmt.Errorf("core.yml has no scheme %s", id)
		}
	}
	return b.String(), nil
}

// comment tells whether a line holds only a comment.
func comment(l string) bool { return strings.HasPrefix(strings.TrimSpace(l), "#") }

// gameCore is the part of core.yml's block of a scheme the entry records:
// all but its id, and, where its formulas are walton.yml's (the same, in
// the same order), all but its formulas too.
func gameCore(block yaml.MapSlice, w Scheme) yaml.MapSlice {
	same := true
	for _, r := range []struct {
		key  string
		have []string
	}{{"variables", w.Variables}, {"premises", w.Premises}, {"assumptions", w.Assumptions},
		{"exceptions", w.Exceptions}, {"conclusions", w.Conclusions}} {
		if strings.Join(stringList(value(block, r.key)), "\x00") != strings.Join(r.have, "\x00") {
			same = false
		}
	}
	var out yaml.MapSlice
	for _, it := range block {
		k := it.Key.(string)
		if k == "id" || (same && in(coreFormulaKeys, k)) {
			continue
		}
		out = append(out, it)
	}
	return out
}

// stringList reads a YAML list of strings.
func stringList(v interface{}) []string {
	l, _ := v.([]interface{})
	var out []string
	for _, x := range l {
		out = append(out, fmt.Sprint(x))
	}
	return out
}

// decodeOrderedJSON decodes JSON keeping the order of object keys: objects
// become yaml.MapSlice, arrays []interface{}, whole numbers int, the others
// float64, so that a record can be kept in a YAML file as written.
func decodeOrderedJSON(b []byte) (interface{}, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	v, err := readJSONValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("data after the value")
	}
	return v, nil
}

func readJSONValue(dec *json.Decoder) (interface{}, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			m := yaml.MapSlice{}
			for dec.More() {
				k, err := dec.Token()
				if err != nil {
					return nil, err
				}
				v, err := readJSONValue(dec)
				if err != nil {
					return nil, err
				}
				m = append(m, yaml.MapItem{Key: k.(string), Value: v})
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			return m, nil
		case '[':
			a := []interface{}{}
			for dec.More() {
				v, err := readJSONValue(dec)
				if err != nil {
					return nil, err
				}
				a = append(a, v)
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			return a, nil
		}
		return nil, fmt.Errorf("unexpected %v", t)
	case json.Number:
		if i, err := t.Int64(); err == nil {
			return int(i), nil
		}
		return t.Float64()
	default:
		return t, nil
	}
}

// gameHeader heads GameDir's YAML files.
const gameHeader = `# The game's data in the catalogue (step C5.1, docs/c5-game.md), with the
# entries' game parts: wag-cli --catalogue-gen generates the game's files
# (wag/data/nascq/*.json, wag/data/schemes/core.yml) from them. Change it only
# as a decision says. Comments are not kept: --catalogue-gen rewrites this file.
`

// GameFiles is GameDir's content, by file name, as the game's files give it.
func (g *Generator) GameFiles() (map[string]string, error) {
	r := g.Src.gameRaw
	if r == nil {
		return nil, fmt.Errorf("the game's files are not read")
	}
	ids := map[string]bool{}
	for _, s := range g.Src.Game {
		ids[s.CarneadesID] = true
	}
	d, err := r.gameData(ids)
	if err != nil {
		return nil, err
	}
	return d.Files()
}

// Files writes GameData as GameDir's files.
func (d *GameData) Files() (map[string]string, error) {
	q, err := yaml.Marshal(struct {
		Questions []GameQuestion `yaml:"questions"`
	}{d.Questions})
	if err != nil {
		return nil, err
	}
	o, err := yaml.Marshal(d.Order)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		GameQuestionsFile: gameHeader + string(q),
		GameOrderFile:     gameHeader + string(o),
		GameCoreTemplate:  d.CoreTemplate,
	}, nil
}
