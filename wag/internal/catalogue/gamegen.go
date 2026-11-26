package catalogue

// The game's files generated from the catalogue (step C5.1, docs/c5-game.md):
// the NASCQ files and core.yml, from the entries' game parts and formulas
// and GameDir's files. The catalogue is the generator's only input, so the
// game takes what the catalogue records and nothing else.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// GameDataDir is the game's data directory, relative to the repository's
// root; the game's files are under it at their paths below.
const GameDataDir = "wag/data"

// The game's files, by their path under GameDataDir.
const (
	GameCorePath    = "schemes/core.yml"
	GameSchemesPath = "nascq/schemes.json"
	GameCQsPath     = "nascq/critical_questions.json"
	GameEdgesPath   = "nascq/edges.json"
)

// GamePaths are the game's files in the order of their fingerprint
// (server.Data.Hash).
var GamePaths = []string{GameCorePath, GameSchemesPath, GameCQsPath, GameEdgesPath}

// coreFormulaKeys are core.yml's formula keys in the order it writes them.
var coreFormulaKeys = []string{"variables", "conclusions", "premises", "assumptions", "exceptions"}

// Loci of a game question (an ask edge): a premise or an assumption of the
// scheme (an assumption is asked at the premise), an exception, the
// conclusion, or the argument as a whole.
var gameLoci = []string{"premise", "exception", "conclusion", "argument"}

// ReadGameDir reads GameDir's files from a catalogue directory.
func ReadGameDir(dir string) (*GameData, error) {
	read := func(name string) ([]byte, error) {
		b, err := os.ReadFile(filepath.Join(dir, GameDir, name))
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %v", GameDir, name, err)
		}
		return b, nil
	}
	d := &GameData{}
	b, err := read(GameQuestionsFile)
	if err != nil {
		return nil, err
	}
	var q struct {
		Questions []GameQuestion `yaml:"questions"`
	}
	if err := yaml.UnmarshalStrict(b, &q); err != nil {
		return nil, fmt.Errorf("%s/%s: %v", GameDir, GameQuestionsFile, err)
	}
	d.Questions = q.Questions
	if b, err = read(GameOrderFile); err != nil {
		return nil, err
	}
	if err := yaml.UnmarshalStrict(b, &d.Order); err != nil {
		return nil, fmt.Errorf("%s/%s: %v", GameDir, GameOrderFile, err)
	}
	if b, err = read(GameCoreTemplate); err != nil {
		return nil, err
	}
	d.CoreTemplate = string(b)
	return d, nil
}

// gameIn is a game scheme as the generator reads it from its entry.
type gameIn struct {
	id, cid string
	entry   *Entry
	block   yaml.MapSlice // core.yml's block, its id first
	roles   map[string][]string
	vars    map[string]bool
}

// GameGen is the generator's output: the game's files by path, and the
// fingerprint of the values it read.
type GameGen struct {
	Files  map[string][]byte
	Inputs string
}

// GenerateGame makes the game's files from the catalogue's entries (by file
// name, as ReadDir returns them) and GameDir's data. It refuses a catalogue
// whose game parts disagree with one another or break the rules of a game
// question (docs/c5-game.md, C5.1 step 3).
func GenerateGame(entries map[string]*Entry, d *GameData) (*GameGen, error) {
	var errs []string
	fail := func(format string, a ...interface{}) { errs = append(errs, fmt.Sprintf(format, a...)) }

	names := make([]string, 0, len(entries))
	for n := range entries {
		names = append(names, n)
	}
	sort.Strings(names)
	byGame := map[string]*Entry{}
	for _, n := range names {
		e := entries[n]
		if e.Game == nil {
			continue
		}
		if o, ok := byGame[e.Game.ID]; ok {
			fail("%s is the game part of both %s and %s", e.Game.ID, o.ID, e.ID)
			continue
		}
		byGame[e.Game.ID] = e
	}

	// The schemes, in order.yml's order.
	var schemes []*gameIn
	byID, byCID := map[string]*gameIn{}, map[string]*gameIn{}
	for _, gid := range d.Order.Schemes {
		e, ok := byGame[gid]
		switch {
		case !ok:
			fail("%s/%s names %s, which is no entry's game part", GameDir, GameOrderFile, gid)
			continue
		case byID[gid] != nil:
			fail("%s/%s names %s twice", GameDir, GameOrderFile, gid)
			continue
		}
		g, err := schemeIn(gid, e)
		if err != nil {
			fail("%s: %v", e.ID, err)
			continue
		}
		if o := byCID[g.cid]; o != nil {
			fail("%s and %s are both core.yml's %s", o.id, gid, g.cid)
			continue
		}
		schemes = append(schemes, g)
		byID[gid], byCID[g.cid] = g, g
	}
	for _, n := range names {
		if e := entries[n]; e.Game != nil && byID[e.Game.ID] == nil && byGame[e.Game.ID] == e {
			if !in(d.Order.Schemes, e.Game.ID) {
				fail("%s: its game part %s is not in %s/%s", e.ID, e.Game.ID, GameDir, GameOrderFile)
			}
		}
	}

	// The questions, in their order.
	questions := yaml.MapSlice{}
	qids := map[string]bool{}
	for _, q := range d.Questions {
		switch {
		case q.ID == "":
			fail("%s/%s: a question without an id", GameDir, GameQuestionsFile)
		case qids[q.ID]:
			fail("%s/%s: %s twice", GameDir, GameQuestionsFile, q.ID)
		case len(q.Record) == 0:
			fail("%s/%s: %s has no record", GameDir, GameQuestionsFile, q.ID)
		default:
			qids[q.ID] = true
			questions = append(questions, yaml.MapItem{Key: q.ID, Value: q.Record})
		}
	}

	// The ask edges: order.yml's order, each with its entry's locus and
	// pattern, which must be the scheme's.
	var ask []interface{}
	asked := map[GameEdge]bool{}
	for _, a := range d.Order.Ask {
		g := byID[a.Scheme]
		switch {
		case g == nil:
			fail("%s/%s: the ask edge %s of %s: no such game scheme", GameDir, GameOrderFile, a.CQ, a.Scheme)
			continue
		case !qids[a.CQ]:
			fail("%s/%s: the ask edge %s of %s: no such question", GameDir, GameOrderFile, a.CQ, a.Scheme)
			continue
		case asked[a]:
			fail("%s/%s: the ask edge %s of %s twice", GameDir, GameOrderFile, a.CQ, a.Scheme)
			continue
		}
		asked[a] = true
		var found []*GameCQ
		for i := range g.entry.CQs {
			if q := g.entry.CQs[i].Game; q != nil && q.ID == a.CQ {
				found = append(found, q)
			}
		}
		if len(found) != 1 {
			fail("%s: %d questions with the game part %s, not one", g.entry.ID, len(found), a.CQ)
			continue
		}
		q := found[0]
		if err := g.askRule(q.Locus, q.Pattern); err != nil {
			fail("%s: %s: %v", g.entry.ID, a.CQ, err)
			continue
		}
		edge := yaml.MapSlice{{Key: "cq", Value: a.CQ}, {Key: "scheme", Value: a.Scheme}, {Key: "locus", Value: q.Locus}}
		if q.Pattern != "" {
			edge = append(edge, yaml.MapItem{Key: "pattern", Value: g.rebrace(q.Pattern)})
		}
		ask = append(ask, edge)
	}
	for _, g := range schemes {
		for _, c := range g.entry.CQs {
			if c.Game != nil && !asked[GameEdge{CQ: c.Game.ID, Scheme: g.id}] {
				fail("%s: %s asks %s, which %s/%s does not", g.entry.ID, c.ID, c.Game.ID, GameDir, GameOrderFile)
			}
		}
	}

	// The backup edges, in order.yml's order.
	var backup []interface{}
	backed := map[GameEdge]bool{}
	for _, b := range d.Order.Backup {
		switch {
		case byID[b.Scheme] == nil:
			fail("%s/%s: the backup edge %s by %s: no such game scheme", GameDir, GameOrderFile, b.CQ, b.Scheme)
		case !qids[b.CQ]:
			fail("%s/%s: the backup edge %s by %s: no such question", GameDir, GameOrderFile, b.CQ, b.Scheme)
		case backed[b]:
			fail("%s/%s: the backup edge %s by %s twice", GameDir, GameOrderFile, b.CQ, b.Scheme)
		default:
			backed[b] = true
			backup = append(backup, yaml.MapSlice{{Key: "scheme", Value: b.Scheme}, {Key: "cq", Value: b.CQ}})
		}
	}

	// core.yml: the template with each marker replaced by its scheme's block.
	core, used, err := fillCore(d.CoreTemplate, byCID)
	if err != nil {
		fail("%s/%s: %v", GameDir, GameCoreTemplate, err)
	}
	for _, g := range schemes {
		if !used[g.cid] {
			fail("%s/%s has no marker for %s (%s)", GameDir, GameCoreTemplate, g.cid, g.id)
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("the catalogue's game data does not make the game's files:\n  %s", strings.Join(errs, "\n  "))
	}

	nascq := yaml.MapSlice{}
	for _, g := range schemes {
		nascq = append(nascq, yaml.MapItem{Key: g.id, Value: g.entry.Game.NASCQ})
	}
	edges := yaml.MapSlice{{Key: "ask_edges", Value: ask}, {Key: "backup_edges", Value: backup}}
	if ask == nil {
		edges[0].Value = []interface{}{}
	}
	if backup == nil {
		edges[1].Value = []interface{}{}
	}

	// The fingerprint of the values read: the outputs' values, compact, and
	// the filled core.yml. Their layout does not change it.
	h := sha256.New()
	for _, v := range []interface{}{nascq, questions, edges} {
		var b bytes.Buffer
		if err := writeJSON(&b, v, "", true, false); err != nil {
			return nil, err
		}
		h.Write(b.Bytes())
		h.Write([]byte{0})
	}
	h.Write([]byte(core))
	inputs := hex.EncodeToString(h.Sum(nil))[:16]

	out := &GameGen{Files: map[string][]byte{}, Inputs: inputs}
	for path, v := range map[string]interface{}{GameSchemesPath: nascq, GameCQsPath: questions, GameEdgesPath: edges} {
		var b bytes.Buffer
		if err := writeJSON(&b, v, "", false, false); err != nil {
			return nil, fmt.Errorf("%s: %v", path, err)
		}
		b.WriteString("\n")
		out.Files[path] = b.Bytes()
	}
	out.Files[GameCorePath] = []byte(coreHeader(inputs) + core)
	return out, nil
}

// coreHeader heads the generated core.yml.
func coreHeader(inputs string) string {
	return "# Generated by wag-cli --game-gen from the scheme catalogue (wag/data/catalogue:\n" +
		"# the entries' game parts and formulas, and game/), with the NASCQ files beside\n" +
		"# it. Do not edit them: edit the catalogue and generate them again\n" +
		"# (docs/c5-game.md).\n" +
		"# catalogue: " + inputs + "\n" +
		"#\n"
}

// schemeIn reads a game scheme from its entry: its NASCQ record, and its
// core.yml block, whose formulas are the entry's own (walton.yml's) unless
// the entry records others.
func schemeIn(gid string, e *Entry) (*gameIn, error) {
	if len(e.Game.NASCQ) == 0 {
		return nil, fmt.Errorf("the game part %s has no NASCQ record", gid)
	}
	cid, ok := value(e.Game.NASCQ, "carneades_id").(string)
	if !ok || cid == "" {
		return nil, fmt.Errorf("the NASCQ record of %s has no carneades_id", gid)
	}
	g := &gameIn{id: gid, cid: cid, entry: e, roles: map[string][]string{}, vars: map[string]bool{}}
	g.block = yaml.MapSlice{{Key: "id", Value: cid}}
	override := false
	for _, it := range e.Game.Core {
		k, ok := it.Key.(string)
		if !ok || k == "id" {
			return nil, fmt.Errorf("the core block of %s has the key %v", gid, it.Key)
		}
		if in(coreFormulaKeys, k) {
			override = true
		}
		g.block = append(g.block, it)
	}
	if override {
		for _, k := range []string{"variables", "conclusions"} {
			if value(e.Game.Core, k) == nil {
				return nil, fmt.Errorf("the core block of %s gives formulas but no %s", gid, k)
			}
		}
	} else {
		f := e.Formulas
		if f == nil {
			return nil, fmt.Errorf("no formulas for the game's %s", gid)
		}
		for _, r := range []struct {
			key  string
			have []string
		}{{"variables", f.Variables}, {"conclusions", f.Conclusions}, {"premises", f.Premises},
			{"assumptions", f.Assumptions}, {"exceptions", f.Exceptions}} {
			if len(r.have) > 0 {
				l := make([]interface{}, len(r.have))
				for i, x := range r.have {
					l[i] = x
				}
				g.block = append(g.block, yaml.MapItem{Key: r.key, Value: l})
			}
		}
	}
	for _, k := range coreFormulaKeys {
		v := value(g.block, k)
		if v == nil {
			continue
		}
		l, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("the core block of %s: %s is not a list", gid, k)
		}
		for _, x := range l {
			s, ok := x.(string)
			if !ok || s == "" {
				return nil, fmt.Errorf("the core block of %s: %s has %v, not a formula", gid, k, x)
			}
			g.roles[k] = append(g.roles[k], s)
		}
	}
	for _, v := range g.roles["variables"] {
		g.vars[v] = true
	}
	return g, nil
}

// askRule is the rule of a game question's locus and pattern: asked at a
// premise, its pattern is a premise or an assumption of the formulas the
// game plays; at an exception, an exception; of the conclusion or of the
// argument, it has none.
func (g *gameIn) askRule(locus, pattern string) error {
	if !in(gameLoci, locus) {
		return fmt.Errorf("the locus %q is not one of %s", locus, strings.Join(gameLoci, ", "))
	}
	switch locus {
	case "premise":
		if pattern == "" || !(in(g.roles["premises"], pattern) || in(g.roles["assumptions"], pattern)) {
			return fmt.Errorf("asked at a premise, its pattern %q is no premise or assumption of the game's %s", pattern, g.cid)
		}
	case "exception":
		if pattern == "" || !in(g.roles["exceptions"], pattern) {
			return fmt.Errorf("asked at an exception, its pattern %q is no exception of the game's %s", pattern, g.cid)
		}
	default:
		if pattern != "" {
			return fmt.Errorf("asked of the %s, it has a pattern (%q)", locus, pattern)
		}
	}
	return nil
}

var identifier = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

// rebrace writes a pattern as the NASCQ does: the scheme's variables in
// braces.
func (g *gameIn) rebrace(pattern string) string {
	return identifier.ReplaceAllStringFunc(pattern, func(s string) string {
		if g.vars[s] {
			return "{" + s + "}"
		}
		return s
	})
}

// fillCore replaces each marker of core.yml's template with its scheme's
// block. It reports the Carneades ids it filled.
func fillCore(tpl string, byCID map[string]*gameIn) (string, map[string]bool, error) {
	used := map[string]bool{}
	var b strings.Builder
	for _, l := range strings.SplitAfter(tpl, "\n") {
		id, ok := strings.CutPrefix(strings.TrimRight(l, "\r\n"), coreMarker)
		if !ok {
			b.WriteString(l)
			continue
		}
		g := byCID[id]
		switch {
		case g == nil:
			return "", nil, fmt.Errorf("a marker for %s, which no game scheme is", id)
		case used[id]:
			return "", nil, fmt.Errorf("two markers for %s", id)
		}
		used[id] = true
		y, err := yaml.Marshal(g.block[1:])
		if err != nil {
			return "", nil, err
		}
		b.WriteString("  - id: " + id + "\n")
		for _, line := range strings.SplitAfter(strings.TrimRight(string(y), "\n"), "\n") {
			b.WriteString("    " + line)
		}
		b.WriteString("\n")
	}
	return b.String(), used, nil
}

// writeJSON writes v as JSON, objects in their order. Compact, it writes no
// space; otherwise an object has a key a line, indented, but an object of
// scalars in a list, which is one line, and a list of scalars or an object
// of scalars, each on one line when that is short.
func writeJSON(b *bytes.Buffer, v interface{}, indent string, compact, inline bool) error {
	switch x := v.(type) {
	case yaml.MapSlice:
		if len(x) == 0 {
			b.WriteString("{}")
			return nil
		}
		if compact || inline || (scalars(x) && short(x, indent)) {
			b.WriteString("{")
			for i, it := range x {
				if i > 0 {
					b.WriteString(sep(compact))
				}
				if err := writeKey(b, it.Key, compact); err != nil {
					return err
				}
				if err := writeJSON(b, it.Value, indent, compact, true); err != nil {
					return err
				}
			}
			b.WriteString("}")
			return nil
		}
		b.WriteString("{\n")
		for i, it := range x {
			b.WriteString(indent + "  ")
			if err := writeKey(b, it.Key, false); err != nil {
				return err
			}
			if err := writeJSON(b, it.Value, indent+"  ", false, false); err != nil {
				return err
			}
			if i < len(x)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(indent + "}")
		return nil
	case []interface{}:
		if len(x) == 0 {
			b.WriteString("[]")
			return nil
		}
		if compact || inline || (scalars(x) && short(x, indent)) {
			b.WriteString("[")
			for i, e := range x {
				if i > 0 {
					b.WriteString(sep(compact))
				}
				if err := writeJSON(b, e, indent, compact, true); err != nil {
					return err
				}
			}
			b.WriteString("]")
			return nil
		}
		b.WriteString("[\n")
		for i, e := range x {
			b.WriteString(indent + "  ")
			m, isMap := e.(yaml.MapSlice)
			if err := writeJSON(b, e, indent+"  ", false, isMap && scalars(m)); err != nil {
				return err
			}
			if i < len(x)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(indent + "]")
		return nil
	case string, bool, nil, int, int64, uint64, float64:
		var s bytes.Buffer
		enc := json.NewEncoder(&s)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(x); err != nil {
			return err
		}
		b.Write(bytes.TrimRight(s.Bytes(), "\n"))
		return nil
	default:
		return fmt.Errorf("%v (%T) is not a JSON value", v, v)
	}
}

func writeKey(b *bytes.Buffer, k interface{}, compact bool) error {
	s, ok := k.(string)
	if !ok {
		return fmt.Errorf("the key %v is not a string", k)
	}
	if err := writeJSON(b, s, "", true, true); err != nil {
		return err
	}
	b.WriteString(":" + map[bool]string{false: " ", true: ""}[compact])
	return nil
}

func sep(compact bool) string {
	if compact {
		return ","
	}
	return ", "
}

// scalars tells whether an object's values, or a list's items, are all
// scalars.
func scalars(v interface{}) bool {
	var items []interface{}
	switch x := v.(type) {
	case yaml.MapSlice:
		for _, it := range x {
			items = append(items, it.Value)
		}
	case []interface{}:
		items = x
	}
	for _, it := range items {
		switch it.(type) {
		case yaml.MapSlice, []interface{}:
			return false
		}
	}
	return true
}

// short tells whether v fits on one line after indent.
func short(v interface{}, indent string) bool {
	var b bytes.Buffer
	if err := writeJSON(&b, v, "", false, true); err != nil {
		return false
	}
	return len(indent)+b.Len() <= 80
}
