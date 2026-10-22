package catalogue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/registry"
	"gopkg.in/yaml.v2"
)

// The source files, as paths relative to the repository's root.
var SourcePaths = struct {
	Walton, Links, Migrations, Changes, Families, Schemes, CQs, Edges, Core string
}{
	Walton:     "docs/scheme-registry/walton-clean/walton.yml",
	Links:      "docs/scheme-registry/links.json",
	Migrations: "docs/scheme-registry/walton-clean/id-migrations.json",
	Changes:    "docs/scheme-registry/walton-clean/changes.json",
	Families:   "wag/data/schemes/families.json",
	Schemes:    "wag/data/nascq/schemes.json",
	CQs:        "wag/data/nascq/critical_questions.json",
	Edges:      "wag/data/nascq/edges.json",
	Core:       "wag/data/schemes/core.yml",
}

// Dir is the catalogue's directory, relative to the repository's root.
const Dir = "wag/data/catalogue"

// FindRoot walks up from dir to the repository's root: the first directory
// that holds the links file.
func FindRoot(dir string) (string, error) {
	d, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(d, filepath.FromSlash(SourcePaths.Links))); err == nil {
			return d, nil
		}
		up := filepath.Dir(d)
		if up == d {
			return "", fmt.Errorf("no repository root above %s (looked for %s)", dir, SourcePaths.Links)
		}
		d = up
	}
}

// Sources are the files the catalogue is generated from and checked against.
type Sources struct {
	Walton   []Scheme          // walton.yml, in file order
	Language map[string]string // walton.yml's English templates, by "predicate/arity"
	Links    []LinkEntry       // links.json, in compendium order
	// Migrations maps an upstream scheme id to its id in the cleaned
	// walton.yml.
	Migrations map[string]string
	// Edits maps, per scheme, a formula of the upstream walton.yml (compacted)
	// to what the cleaning made of it (changes.json): links.json was drafted
	// against both spellings.
	Edits    map[string]map[string]string
	Families []Family
	Game     map[string]NASCQScheme // by NASCQ id ("AS:...")
	GameCQs  map[string]NASCQCQ     // by NASCQ id ("CQ:...")
	AskEdges []AskEdge
	Backups  []BackupEdge
	Core     map[string]Scheme // the game's core.yml, by id
	// Close pairs walton.yml's schemes whose patterns the registry finds
	// related (internal/registry: the same form, one formula apart, one
	// extending the other, close).
	Close []registry.Relation
}

// Scheme is a Carneades scheme as walton.yml and core.yml write it.
type Scheme struct {
	ID   string `yaml:"id"`
	Meta struct {
		Title  string `yaml:"title"`
		Source string `yaml:"source"`
	} `yaml:"meta"`
	Variables   []string `yaml:"variables"`
	Premises    []string `yaml:"premises"`
	Assumptions []string `yaml:"assumptions"`
	Exceptions  []string `yaml:"exceptions"`
	Conclusions []string `yaml:"conclusions"`
}

// role returns the scheme's formulas for a locus.
func (s Scheme) role(locus string) []string {
	switch locus {
	case "premise":
		return s.Premises
	case "assumption":
		return s.Assumptions
	case "exception":
		return s.Exceptions
	case "conclusion":
		return s.Conclusions
	}
	return nil
}

// locusOf returns the role a formula has in the scheme ("" if none).
func (s Scheme) locusOf(formula string) string {
	for _, l := range Loci {
		for _, f := range s.role(l) {
			if compact(f) == compact(formula) {
				return l
			}
		}
	}
	return ""
}

// LinkEntry is an entry of links.json: a compendium entry and the walton.yml
// schemes that formalize it.
type LinkEntry struct {
	Number      string       `json:"number"`
	Name        string       `json:"name"`
	BookPages   string       `json:"book_pages"`
	Parent      *string      `json:"parent"`
	HeadingOnly bool         `json:"heading_only"`
	CQLabels    []string     `json:"cq_labels"`
	SharesCQs   *string      `json:"shares_cqs_printed_under"`
	Links       []WaltonLink `json:"walton_links"`
}

// WaltonLink links a compendium entry to a walton.yml scheme.
type WaltonLink struct {
	ID         string     `json:"id"`
	Match      string     `json:"match"`
	Reason     string     `json:"reason"`
	PremiseMap []MapItem  `json:"premise_map"`
	Conclusion *MapItem   `json:"conclusion_map"`
	CQMap      []CQMapRow `json:"cq_map"`
}

// MapItem maps a printed premise (or the conclusion) to formulas.
type MapItem struct {
	Label  string         `json:"compendium_label"`
	Note   string         `json:"note"`
	Walton []FormulaLocus `json:"walton"`
}

// CQMapRow maps a printed critical question to formulas.
type CQMapRow struct {
	CQ     string         `json:"cq"`
	Note   string         `json:"note"`
	Walton []FormulaLocus `json:"walton"`
}

// FormulaLocus is a formula and its role, as links.json gives them.
type FormulaLocus struct {
	Formula string `json:"formula"`
	Locus   string `json:"locus"`
}

// Family is a family of families.json.
type Family struct {
	ID      string `json:"id"`
	Generic string `json:"generic"`
	Name    string `json:"name"`
	Members []struct {
		Scheme     string `json:"scheme"`
		Relation   string `json:"relation"`
		Compendium string `json:"compendium"`
		Note       string `json:"note"`
	} `json:"members"`
}

// NASCQScheme is a scheme node of the game's NASCQ.
type NASCQScheme struct {
	CarneadesID string `json:"carneades_id"`
	Name        string `json:"name"`
}

// NASCQCQ is a critical-question node of the game's NASCQ.
type NASCQCQ struct {
	Name     string `json:"name"`
	Template string `json:"question_template_en"`
}

// AskEdge says that a game critical question can be asked of a scheme.
type AskEdge struct {
	CQ      string `json:"cq"`
	Scheme  string `json:"scheme"`
	Locus   string `json:"locus"`
	Pattern string `json:"pattern"`
}

// BackupEdge says that a scheme can back an answer to a critical question.
type BackupEdge struct {
	Scheme string `json:"scheme"`
	CQ     string `json:"cq"`
}

// Files are the raw source files.
type Files struct {
	Walton, Links, Migrations, Changes, Families, Schemes, CQs, Edges, Core []byte
}

// ReadFiles reads the source files under the repository's root.
func ReadFiles(root string) (Files, error) {
	var f Files
	for _, x := range []struct {
		dst  *[]byte
		path string
	}{
		{&f.Walton, SourcePaths.Walton}, {&f.Links, SourcePaths.Links}, {&f.Migrations, SourcePaths.Migrations}, {&f.Changes, SourcePaths.Changes},
		{&f.Families, SourcePaths.Families}, {&f.Schemes, SourcePaths.Schemes}, {&f.CQs, SourcePaths.CQs},
		{&f.Edges, SourcePaths.Edges}, {&f.Core, SourcePaths.Core},
	} {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(x.path)))
		if err != nil {
			return f, err
		}
		*x.dst = b
	}
	return f, nil
}

// Parse decodes the source files.
func Parse(f Files) (*Sources, error) {
	s := &Sources{}
	var w struct {
		Language map[string]string `yaml:"language"`
		Schemes  []Scheme          `yaml:"argument_schemes"`
	}
	if err := yaml.Unmarshal(f.Walton, &w); err != nil {
		return nil, fmt.Errorf("walton.yml: %v", err)
	}
	if len(w.Schemes) == 0 {
		return nil, fmt.Errorf("walton.yml: no argument_schemes")
	}
	s.Walton, s.Language = w.Schemes, map[string]string{}
	for k, v := range w.Language {
		s.Language[k] = strings.TrimSpace(v)
	}
	var links struct {
		Entries []LinkEntry `json:"entries"`
	}
	if err := json.Unmarshal(f.Links, &links); err != nil {
		return nil, fmt.Errorf("links.json: %v", err)
	}
	s.Links = links.Entries
	var mig struct {
		Schemes map[string]string `json:"schemes"`
	}
	if err := json.Unmarshal(f.Migrations, &mig); err != nil {
		return nil, fmt.Errorf("id-migrations.json: %v", err)
	}
	s.Migrations = mig.Schemes
	var changes []struct {
		Scheme string `json:"scheme_after"`
		Before string `json:"before"`
		After  string `json:"after"`
	}
	if err := json.Unmarshal(f.Changes, &changes); err != nil {
		return nil, fmt.Errorf("changes.json: %v", err)
	}
	s.Edits = map[string]map[string]string{}
	for _, c := range changes {
		b, okb := listItem(c.Before)
		a, oka := listItem(c.After)
		if !okb || !oka || c.Scheme == "language" {
			continue
		}
		if s.Edits[c.Scheme] == nil {
			s.Edits[c.Scheme] = map[string]string{}
		}
		s.Edits[c.Scheme][compact(b)] = a
	}
	var fam struct {
		Families []Family `json:"families"`
	}
	if err := json.Unmarshal(f.Families, &fam); err != nil {
		return nil, fmt.Errorf("families.json: %v", err)
	}
	s.Families = fam.Families
	if err := json.Unmarshal(f.Schemes, &s.Game); err != nil {
		return nil, fmt.Errorf("nascq/schemes.json: %v", err)
	}
	if err := json.Unmarshal(f.CQs, &s.GameCQs); err != nil {
		return nil, fmt.Errorf("nascq/critical_questions.json: %v", err)
	}
	var edges struct {
		Ask    []AskEdge    `json:"ask_edges"`
		Backup []BackupEdge `json:"backup_edges"`
	}
	if err := json.Unmarshal(f.Edges, &edges); err != nil {
		return nil, fmt.Errorf("nascq/edges.json: %v", err)
	}
	s.AskEdges, s.Backups = edges.Ask, edges.Backup
	var core struct {
		Schemes []Scheme `yaml:"argument_schemes"`
	}
	if err := yaml.Unmarshal(f.Core, &core); err != nil {
		return nil, fmt.Errorf("core.yml: %v", err)
	}
	s.Core = map[string]Scheme{}
	for _, c := range core.Schemes {
		s.Core[c.ID] = c
	}
	ad, err := carneades.NewAdapterFromYAML(f.Walton)
	if err != nil {
		return nil, fmt.Errorf("walton.yml (Carneades): %v", err)
	}
	var ps []registry.Pattern
	for _, x := range ad.Graph.Theory.ArgSchemes {
		ps = append(ps, registry.PatternOf(x))
	}
	s.Close = registry.Compare(ps)
	return s, nil
}

// Load reads and decodes the sources under the repository's root.
func Load(root string) (*Sources, error) {
	f, err := ReadFiles(root)
	if err != nil {
		return nil, err
	}
	return Parse(f)
}

// scheme returns the walton.yml scheme with an id (after migration).
func (s *Sources) scheme(id string) (Scheme, bool) {
	id = s.migrate(id)
	for _, w := range s.Walton {
		if w.ID == id {
			return w, true
		}
	}
	return Scheme{}, false
}

// migrate maps an upstream scheme id to the cleaned walton.yml's.
func (s *Sources) migrate(id string) string {
	if n, ok := s.Migrations[id]; ok {
		return n
	}
	return id
}

// link returns the links.json entry with a compendium number.
func (s *Sources) link(number string) (LinkEntry, bool) {
	for _, e := range s.Links {
		if e.Number == number {
			return e, true
		}
	}
	return LinkEntry{}, false
}

// printedCQs returns the entry under which an entry's critical questions are
// printed, and their labels.
func (s *Sources) printedCQs(e LinkEntry) (string, []string) {
	if len(e.CQLabels) == 0 && e.SharesCQs != nil {
		if p, ok := s.link(*e.SharesCQs); ok {
			return p.Number, p.CQLabels
		}
	}
	return e.Number, e.CQLabels
}

// gameScheme returns the NASCQ scheme that plays a walton.yml scheme.
func (s *Sources) gameScheme(waltonID string) (string, bool) {
	ids := make([]string, 0, len(s.Game))
	for id, g := range s.Game {
		if s.migrate(g.CarneadesID) == waltonID {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return "", false
	}
	sortStrings(ids)
	return ids[0], true
}

// listItem reads a YAML list item ("      - f(X)") as a formula.
func listItem(line string) (string, bool) {
	t := strings.TrimSpace(line)
	if !strings.HasPrefix(t, "- ") {
		return "", false
	}
	return strings.TrimSpace(t[2:]), true
}

// edited follows the cleaning's edits of a formula of a scheme, if any.
func (s *Sources) edited(scheme, formula string) string {
	for i := 0; i < 8; i++ {
		n, ok := s.Edits[scheme][compact(formula)]
		if !ok || compact(n) == compact(formula) {
			break
		}
		formula = n
	}
	return formula
}

// compact drops the spaces of a formula: "f(a, b)" and "f(a,b)" are the
// same formula.
func compact(formula string) string { return strings.Join(strings.Fields(formula), "") }

// normLocus reads a links.json locus, which may carry a note after the
// role ("assumption (misspelled key ...)").
func normLocus(l string) string {
	for _, x := range append(Loci, "argument") {
		if strings.HasPrefix(l, x) {
			return x
		}
	}
	return ""
}

// unbrace turns a NASCQ pattern ("asserts({W},{S})") into a formula
// ("asserts(W,S)").
func unbrace(p string) string { return strings.NewReplacer("{", "", "}", "").Replace(p) }
