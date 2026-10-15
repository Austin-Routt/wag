// Package annotate has language models annotate argumentative claims as
// instances of Walton's argumentation schemes, under written guidelines,
// and measures and reviews what they produce: several annotators on the
// same claims, each claim in its own call, top-down; validation of each
// annotation against the scheme collection; agreement between annotators
// and with a gold annotation; a model's review of their disagreements; a
// person's verdict on each claim. A run keeps everything in one folder.
//
// The dashboard (internal/server) and wag-llm annotate both use it.
package annotate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yourorg/wag/data"
)

// Annotation is one claim's annotation, in the guidelines' output format
// (docs/kialo-batch2/guidelines.md, §4 and §5).
type Annotation = map[string]any

// Item is a claim to annotate: its entry of items.json.
type Item struct {
	ID   string
	Text string
	Raw  json.RawMessage
}

// Resource is a file the annotators are given besides the guidelines.
type Resource struct {
	Name string
	Text string
}

// Dataset is what annotators are given, and what they are scored against.
type Dataset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source"` // "built in", or the folder it was read from
	NItems      int    `json:"items"`
	HasGold     bool   `json:"has_gold"`

	Discussion json.RawMessage       `json:"-"`
	Items      []Item                `json:"-"`
	Guidelines string                `json:"-"`
	Resources  []Resource            `json:"-"`
	Gold       map[string]Annotation `json:"-"`
	// Hashes are the SHA-256 of every input, by name.
	Hashes map[string]string `json:"-"`
}

// manifest is a dataset's dataset.json.
type manifest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Items       string   `json:"items"`
	Guidelines  string   `json:"guidelines"`
	Gold        string   `json:"gold"`
	Resources   []string `json:"resources"`
}

func sha(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// Builtin lists the datasets built into the program (data/annotation).
func Builtin() ([]*Dataset, error) {
	entries, err := fs.ReadDir(data.Annotation, "annotation")
	if err != nil {
		return nil, err
	}
	var out []*Dataset
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		d, err := load(data.Annotation, path.Join("annotation", e.Name()), "built in")
		if err != nil {
			return nil, fmt.Errorf("built-in dataset %s: %v", e.Name(), err)
		}
		out = append(out, d)
	}
	return out, nil
}

// FromDir reads a dataset from a folder: items.json (required), and
// optionally dataset.json, guidelines.md (else the built-in guidelines),
// gold.json and lexicon.json. The annotators are also given walton.yml and
// the game's NASCQ files.
func FromDir(dir string) (*Dataset, error) {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return load(os.DirFS(dir), ".", dir)
}

// load reads the dataset at dir of fsys.
func load(fsys fs.FS, dir, source string) (*Dataset, error) {
	read := func(name string) ([]byte, error) { return fs.ReadFile(fsys, path.Clean(path.Join(dir, name))) }
	var m manifest
	if b, err := read("dataset.json"); err == nil {
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, fmt.Errorf("dataset.json: %v", err)
		}
	}
	if m.Items == "" {
		m.Items = "items.json"
	}
	if m.ID == "" {
		m.ID = path.Base(dir)
		if source != "built in" {
			m.ID = filepath.Base(source)
		}
	}
	if m.Name == "" {
		m.Name = m.ID
	}
	d := &Dataset{ID: m.ID, Name: m.Name, Description: m.Description, Source: source, Hashes: map[string]string{}}
	b, err := read(m.Items)
	if err != nil {
		return nil, fmt.Errorf("the claims: %v", err)
	}
	d.Hashes[m.Items] = sha(b)
	var in struct {
		Discussion json.RawMessage   `json:"discussion"`
		Items      []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(b, &in); err != nil {
		return nil, fmt.Errorf("%s: %v", m.Items, err)
	}
	d.Discussion = in.Discussion
	seen := map[string]bool{}
	for i, raw := range in.Items {
		var it struct {
			ID   string `json:"kialo_claim_id"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(raw, &it); err != nil || it.ID == "" {
			return nil, fmt.Errorf("%s: item %d has no kialo_claim_id", m.Items, i+1)
		}
		if seen[it.ID] {
			return nil, fmt.Errorf("%s: two items are %s", m.Items, it.ID)
		}
		seen[it.ID] = true
		d.Items = append(d.Items, Item{ID: it.ID, Text: it.Text, Raw: raw})
	}
	d.NItems = len(d.Items)
	// The guidelines: the dataset's, else the built-in ones.
	if m.Guidelines == "" {
		m.Guidelines = "guidelines.md"
	}
	g, err := read(m.Guidelines)
	if err != nil {
		g, err = fs.ReadFile(data.Annotation, "annotation/kialo-13292-batch2/guidelines.md")
		if err != nil {
			return nil, err
		}
		m.Guidelines = "guidelines.md (built in)"
	}
	d.Guidelines = string(g)
	d.Hashes[m.Guidelines] = sha(g)
	// The gold annotation, if there is one.
	if m.Gold == "" {
		m.Gold = "gold.json"
	}
	if b, err := read(m.Gold); err == nil {
		anns, err := parseAnnotations(b)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", m.Gold, err)
		}
		d.Gold, d.HasGold = anns, true
		d.Hashes[m.Gold] = sha(b)
	}
	// The resources: the scheme collection and the game's NASCQ, and the
	// dataset's own (a lexicon).
	if len(m.Resources) == 0 {
		m.Resources = []string{"lexicon.json"}
	}
	add := func(name string, b []byte) {
		d.Resources = append(d.Resources, Resource{Name: name, Text: string(b)})
		d.Hashes[name] = sha(b)
	}
	hasWalton := false
	for _, r := range m.Resources {
		b, err := read(r)
		if err != nil {
			if r == "lexicon.json" {
				continue // optional
			}
			return nil, fmt.Errorf("resource %s: %v", r, err)
		}
		if path.Base(r) == "walton.yml" {
			hasWalton = true
		}
		add(path.Base(r), b)
	}
	if !hasWalton {
		b, err := fs.ReadFile(data.Annotation, "annotation/walton.yml")
		if err != nil {
			return nil, err
		}
		d.Resources = append([]Resource{{Name: "walton.yml", Text: string(b)}}, d.Resources...)
		d.Hashes["walton.yml"] = sha(b)
	}
	add("schemes.json", data.SchemesJSON)
	add("critical_questions.json", data.CQsJSON)
	add("edges.json", data.EdgesJSON)
	return d, nil
}

// parseAnnotations reads an annotation file ({"annotator": …, "items": […]}).
func parseAnnotations(b []byte) (map[string]Annotation, error) {
	var f struct {
		Items []Annotation `json:"items"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	out := map[string]Annotation{}
	for _, it := range f.Items {
		if id, _ := it["kialo_claim_id"].(string); id != "" {
			out[id] = it
		}
	}
	return out, nil
}

// WaltonYAML is the scheme collection the dataset gives its annotators.
func (d *Dataset) WaltonYAML() []byte {
	for _, r := range d.Resources {
		if r.Name == "walton.yml" {
			return []byte(r.Text)
		}
	}
	return nil
}

// Item returns the item with the ID, nil if none.
func (d *Dataset) Item(id string) *Item {
	for i := range d.Items {
		if d.Items[i].ID == id {
			return &d.Items[i]
		}
	}
	return nil
}

// IDs returns the items' IDs, in order.
func (d *Dataset) IDs() []string {
	out := make([]string, len(d.Items))
	for i, it := range d.Items {
		out[i] = it.ID
	}
	return out
}

// Datasets lists the built-in datasets and those in dir (one folder each).
func Datasets(dir string) ([]*Dataset, []error) {
	var out []*Dataset
	var errs []error
	b, err := Builtin()
	if err != nil {
		errs = append(errs, err)
	}
	out = append(out, b...)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		d, err := FromDir(filepath.Join(dir, e.Name()))
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %v", e.Name(), err))
			continue
		}
		d.ID = "folder:" + e.Name()
		out = append(out, d)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.HasPrefix(out[j].ID, "folder:") && !strings.HasPrefix(out[i].ID, "folder:")
	})
	return out, errs
}
