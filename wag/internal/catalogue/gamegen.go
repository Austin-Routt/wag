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

// game returns the game's file at a path of GamePaths, as read.
func (f Files) game(p string) []byte {
	switch p {
	case GameCorePath:
		return f.Core
	case GameSchemesPath:
		return f.Schemes
	case GameCQsPath:
		return f.CQs
	case GameEdgesPath:
		return f.Edges
	}
	return nil
}

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
	d.CoreTemplate = string(lfOf(b)) // as written in a checkout that turns LF into CRLF
	if b, err = read(GameLanguageFile); err != nil {
		return nil, err
	}
	if err := yaml.UnmarshalStrict(b, &d.Language); err != nil {
		return nil, fmt.Errorf("%s/%s: %v", GameDir, GameLanguageFile, err)
	}
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

// GameSources are what the generator reads besides the catalogue:
// walton.yml's formulas by scheme id, which a game scheme plays when its
// entry records none of its own (nil: each entry's own copy), its id
// migrations, which map an upstream Carneades id to walton.yml's, and its
// language's templates, which the game's sentences are unless game/
// records a difference (D094 (f)).
type GameSources struct {
	Formulas   map[string]*Formulas
	Migrations map[string]string
	Language   map[string]string
}

// GameSources are the generator's sources from these sources.
func (s *Sources) GameSources() GameSources {
	return GameSources{Formulas: FormulasOf(s.Walton), Migrations: s.Migrations, Language: s.Language}
}

// GenerateGame makes the game's files from the catalogue's entries (by file
// name, as ReadDir returns them) and GameDir's data, with walton.yml's
// formulas and id migrations (src). It refuses a catalogue whose game parts
// disagree with one another or with their entries, or break the rules of a
// game question (docs/c5-game.md, C5.1 step 3).
func GenerateGame(entries map[string]*Entry, d *GameData, src GameSources) (*GameGen, error) {
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
		f := e.Formulas
		if src.Formulas != nil {
			f = src.Formulas[e.ID]
		}
		g, err := schemeIn(gid, e, f, src.Migrations)
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

	// The questions, in their order; their records are made once their ask
	// edges are read.
	var qorder []GameQuestion
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
			qorder = append(qorder, q)
		}
	}

	// The ask edges: order.yml's order, each with its entry's locus and
	// pattern, which must be the scheme's, and its question's text there,
	// the entry's paraphrase with the scheme's variables braced (rule (c)).
	var ask []interface{}
	asked := map[GameEdge]bool{}
	asks := map[string][]askedAt{}
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
		var found []*CQ
		for i := range g.entry.CQs {
			if q := g.entry.CQs[i].Game; q != nil && q.ID == a.CQ {
				found = append(found, &g.entry.CQs[i])
			}
		}
		if len(found) != 1 {
			fail("%s: %d questions with the game part %s, not one", g.entry.ID, len(found), a.CQ)
			continue
		}
		c := found[0]
		q := c.Game
		if err := g.askRule(q.Locus, q.Pattern); err != nil {
			fail("%s: %s: %v", g.entry.ID, a.CQ, err)
			continue
		}
		if f := paraphraseFault(c.Paraphrase); f != "" {
			fail("%s: %s, the game's %s: %s", g.entry.ID, c.ID, a.CQ, f)
			continue
		}
		formula := ""
		switch q.Locus {
		case "premise", "exception":
			formula = g.rebrace(q.Pattern)
		case "conclusion":
			var cs []string
			for _, x := range g.roles["conclusions"] {
				cs = append(cs, g.rebrace(x))
			}
			formula = strings.Join(cs, "; ")
		}
		asks[a.CQ] = append(asks[a.CQ], askedAt{Scheme: a.Scheme, Locus: q.Locus, Formula: formula, Text: g.rebrace(c.Paraphrase)})
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
	backed := map[[2]string]bool{} // by question and scheme: a mark does not make another edge
	askedQ := map[string]bool{}
	for a := range asked {
		askedQ[a.CQ] = true
	}
	for _, b := range d.Order.Backup {
		switch {
		case byID[b.Scheme] == nil:
			fail("%s/%s: the backup edge %s by %s: no such game scheme", GameDir, GameOrderFile, b.CQ, b.Scheme)
		case !qids[b.CQ]:
			fail("%s/%s: the backup edge %s by %s: no such question", GameDir, GameOrderFile, b.CQ, b.Scheme)
		case !askedQ[b.CQ]:
			fail("%s/%s: the backup edge %s by %s backs a question no scheme asks", GameDir, GameOrderFile, b.CQ, b.Scheme)
		case backed[[2]string{b.CQ, b.Scheme}]:
			fail("%s/%s: the backup edge %s by %s twice", GameDir, GameOrderFile, b.CQ, b.Scheme)
		default:
			backed[[2]string{b.CQ, b.Scheme}] = true
			edge := yaml.MapSlice{{Key: "scheme", Value: b.Scheme}, {Key: "cq", Value: b.CQ}}
			if b.Across != "" {
				edge = append(edge, yaml.MapItem{Key: "across", Value: b.Across})
			}
			backup = append(backup, edge)
		}
	}

	// The questions' records: GameDir's, with the text and the effect the
	// ask edges give (rules (a) to (c)); then the role schemes (rule (d)).
	questions := yaml.MapSlice{}
	for _, q := range qorder {
		at := asks[q.ID]
		if len(at) == 0 {
			fail("%s/%s: %s is asked by no game scheme, so it has no text: its entries' paraphrase", GameDir, GameQuestionsFile, q.ID)
			continue
		}
		rec, err := questionRecord(q.ID, q.Record, at[0].Text, at[0].Locus)
		if err != nil {
			fail("%s/%s: %v", GameDir, GameQuestionsFile, err)
			continue
		}
		questions = append(questions, yaml.MapItem{Key: q.ID, Value: rec})
	}
	errs = append(errs, questionFaults(asks)...)
	backs := map[string][]string{}
	for _, b := range d.Order.Backup {
		if backed[[2]string{b.CQ, b.Scheme}] && !in(backs[b.Scheme], b.CQ) {
			backs[b.Scheme] = append(backs[b.Scheme], b.CQ)
		}
	}
	errs = append(errs, roleFaults(schemes, byID, asks, backs)...)

	// core.yml: the template with each marker replaced by its scheme's
	// block, and the language's marker by its sentences (rule (e)).
	core, used, err := fillCore(d.CoreTemplate, byCID)
	if err != nil {
		fail("%s/%s: %v", GameDir, GameCoreTemplate, err)
	} else {
		for _, g := range schemes {
			if !used[g.cid] {
				fail("%s/%s has no marker for %s (%s)", GameDir, GameCoreTemplate, g.cid, g.id)
			}
		}
		if core, err = fillLanguage(core, src.Language, d.Language); err != nil {
			fail("the game's sentences (%s/%s, %s/%s): %v", GameDir, GameCoreTemplate, GameDir, GameLanguageFile, err)
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

// FormulasOf are walton.yml's formulas by scheme id, as the generator plays
// them.
func FormulasOf(schemes []Scheme) map[string]*Formulas {
	out := map[string]*Formulas{}
	for _, w := range schemes {
		out[w.ID] = &Formulas{Variables: w.Variables, Premises: w.Premises, Assumptions: w.Assumptions,
			Exceptions: w.Exceptions, Conclusions: w.Conclusions}
	}
	return out
}

// GameFromCatalogue generates the game's files from a catalogue directory
// and the repository's walton.yml (under root).
func GameFromCatalogue(root, dir string) (*GameGen, error) {
	f, err := ReadFiles(root)
	if err != nil {
		return nil, err
	}
	_, g, err := WithGame(f, dir)
	return g, err
}

// WithGame returns the source files with the game's replaced by what the
// catalogue in dir makes of its game data, with f's walton.yml, and the
// generator's output. wag-cli --catalogue-gen parses these, so that the
// entries are made from what the catalogue says (step C5.1's switch).
func WithGame(f Files, dir string) (Files, *GameGen, error) {
	var w struct {
		Schemes  []Scheme          `yaml:"argument_schemes"`
		Language map[string]string `yaml:"language"`
	}
	if err := yaml.Unmarshal(f.Walton, &w); err != nil {
		return f, nil, fmt.Errorf("walton.yml: %v", err)
	}
	language := map[string]string{}
	for k, v := range w.Language {
		language[k] = strings.TrimSpace(v) // as Sources reads it
	}
	var mig struct {
		Schemes map[string]string `json:"schemes"`
	}
	if err := json.Unmarshal(f.Migrations, &mig); err != nil {
		return f, nil, fmt.Errorf("id-migrations.json: %v", err)
	}
	entries, bad, err := ReadDir(dir)
	if err != nil {
		return f, nil, err
	}
	if len(bad) > 0 {
		var names []string
		for n, e := range bad {
			names = append(names, n+": "+e.Error())
		}
		sort.Strings(names)
		return f, nil, fmt.Errorf("entries that do not read:\n  %s", strings.Join(names, "\n  "))
	}
	d, err := ReadGameDir(dir)
	if err != nil {
		return f, nil, err
	}
	g, err := GenerateGame(entries, d, GameSources{Formulas: FormulasOf(w.Schemes), Migrations: mig.Schemes, Language: language})
	if err != nil {
		return f, nil, err
	}
	f.Core, f.Schemes, f.CQs, f.Edges = g.Files[GameCorePath], g.Files[GameSchemesPath], g.Files[GameCQsPath], g.Files[GameEdgesPath]
	return f, g, nil
}

// lfOf reads CRLF line endings as LF, as the data's fingerprint does: a
// checkout on Windows may write them.
func lfOf(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }

// WriteGameFiles writes the game's files under dir (laid out as wag/data),
// those that change, and returns their paths.
func WriteGameFiles(dir string, g *GameGen) ([]string, error) {
	var written []string
	for _, p := range GamePaths {
		path := filepath.Join(dir, filepath.FromSlash(p))
		if cur, err := os.ReadFile(path); err == nil && bytes.Equal(lfOf(cur), g.Files[p]) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(path, g.Files[p], 0o644); err != nil {
			return written, err
		}
		written = append(written, p)
	}
	return written, nil
}

// coreHeaderFirst is the first line of the generated core.yml's header.
const coreHeaderFirst = "# Generated by wag-cli --game-gen from the scheme catalogue (wag/data/catalogue:\n"

// stripCoreHeader removes the generator's header from core.yml's text, so
// that the template read from a generated core.yml is the template it was
// made from.
func stripCoreHeader(text string) string {
	if !strings.HasPrefix(text, coreHeaderFirst) {
		return text
	}
	i := strings.Index(text, "\n# catalogue: ")
	if i < 0 {
		return text
	}
	line, _, _ := strings.Cut(text[i+1:], "\n")
	if rest, ok := strings.CutPrefix(text, coreHeader(strings.TrimPrefix(line, "# catalogue: "))); ok {
		return rest
	}
	return text
}

// coreHeader heads the generated core.yml.
func coreHeader(inputs string) string {
	return coreHeaderFirst +
		"# the entries' game parts and formulas, and game/), with the NASCQ files beside\n" +
		"# it. Do not edit them: edit the catalogue and generate them again\n" +
		"# (docs/c5-game.md).\n" +
		"# catalogue: " + inputs + "\n" +
		"#\n"
}

// schemeIn reads a game scheme from its entry: its NASCQ record, and its
// core.yml block, whose formulas are walton.yml's (f) unless the entry
// records others. An entry that records formulas records them all (an
// override is all or nothing: a role it leaves out is empty). The block is
// written with its id, its other keys in the entry's order, then its
// formulas in core.yml's order, whichever way the entry lists them, so that
// one run of the generator is a fixpoint.
func schemeIn(gid string, e *Entry, f *Formulas, migrations map[string]string) (*gameIn, error) {
	if len(e.Game.NASCQ) == 0 {
		return nil, fmt.Errorf("the game part %s has no NASCQ record", gid)
	}
	cid, ok := value(e.Game.NASCQ, "carneades_id").(string)
	if !ok || cid == "" {
		return nil, fmt.Errorf("the NASCQ record of %s has no carneades_id", gid)
	}
	// The game's scheme is the entry's (after walton.yml's id migrations,
	// as the entries are made from the files): another would move the game
	// part to another entry, or to none.
	to := cid
	if m, ok := migrations[cid]; ok {
		to = m
	}
	if to != e.ID {
		return nil, fmt.Errorf("the game part %s names core.yml's %s, which is walton.yml's %s, not this entry's scheme %s", gid, cid, to, e.ID)
	}
	g := &gameIn{id: gid, cid: cid, entry: e, roles: map[string][]string{}, vars: map[string]bool{}}
	g.block = yaml.MapSlice{{Key: "id", Value: cid}}
	formulas := map[string]interface{}{}
	for _, it := range e.Game.Core {
		k, ok := it.Key.(string)
		if !ok || k == "id" {
			return nil, fmt.Errorf("the core block of %s has the key %v", gid, it.Key)
		}
		if in(coreFormulaKeys, k) {
			formulas[k] = it.Value
			continue
		}
		g.block = append(g.block, it)
	}
	if len(formulas) > 0 {
		for _, k := range []string{"variables", "conclusions"} {
			if formulas[k] == nil {
				return nil, fmt.Errorf("the core block of %s gives formulas but no %s", gid, k)
			}
		}
	} else {
		if f == nil {
			return nil, fmt.Errorf("walton.yml has no scheme %s, whose formulas the game's %s plays", e.ID, gid)
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
				formulas[r.key] = l
			}
		}
	}
	for _, k := range coreFormulaKeys {
		v, ok := formulas[k]
		if !ok {
			continue
		}
		g.block = append(g.block, yaml.MapItem{Key: k, Value: v})
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
	// A formula names only the scheme's variables (a capitalized name is a
	// variable, as in Carneades).
	for _, k := range coreFormulaKeys[1:] {
		for _, x := range g.roles[k] {
			for _, t := range identifier.FindAllString(x, -1) {
				if t[0] >= 'A' && t[0] <= 'Z' && !g.vars[t] {
					return nil, fmt.Errorf("the game's %s: %s names %s, which is not one of its variables %v", cid, x, t, g.roles["variables"])
				}
			}
		}
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
	lines := strings.SplitAfter(tpl, "\n")
	for i, l := range lines {
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
		if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "    ") && !comment(lines[i+1]) {
			return "", nil, fmt.Errorf("the marker for %s is followed by lines of a block (%q): the block is the entry's", id, strings.TrimSpace(lines[i+1]))
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
