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
	"bytes"
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
	// Format is the annotation format its guidelines ask for: 2 (the
	// identity contract, docs/annotation-records.md), or 1 for a dataset
	// whose manifest says wag-annotation/1.
	Format int `json:"format"`
	// Case is its case record (case.json), nil if it has none: the contexts,
	// actors and actions the format-2 checks read.
	Case *Case `json:"-"`

	Discussion json.RawMessage       `json:"-"`
	Items      []Item                `json:"-"`
	Guidelines string                `json:"-"`
	Resources  []Resource            `json:"-"`
	Gold       map[string]Annotation `json:"-"`
	// Hashes are the SHA-256 of every input, by name.
	Hashes map[string]string `json:"-"`
	// GoldFile is the name of the gold among Hashes. The gold scores a run but
	// is never shown to an annotator, so it is not part of a run's identity.
	GoldFile string `json:"-"`
	// annotationFiles identifies files actually supplied to annotators, so a
	// file used both for claims/resources and scoring cannot be excluded.
	annotationFiles map[string]bool
	// lineHashes give, for the hash of an input's bytes, the hashes of the
	// same text with LF and with CRLF line endings: a run made from a
	// Windows checkout pinned the CRLF ones.
	lineHashes map[string][2]string
}

// setHash records an input's hashes.
func (d *Dataset) setHash(name string, b []byte) {
	h := sha(b)
	d.Hashes[name] = h
	if d.lineHashes == nil {
		d.lineHashes = map[string][2]string{}
	}
	lf := bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	d.lineHashes[h] = [2]string{sha(lf), sha(bytes.ReplaceAll(lf, []byte("\n"), []byte("\r\n")))}
}

// sameInput tells whether a pinned hash is the input's, its line endings
// read either way.
func (d *Dataset) sameInput(name, pinned string) bool {
	h := d.Hashes[name]
	if pinned == h {
		return true
	}
	alt, ok := d.lineHashes[h]
	return ok && (pinned == alt[0] || pinned == alt[1])
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
	Format      string   `json:"format"`
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
// gold.json, lexicon.json and case.json. The annotators are also given
// walton.yml and the game's NASCQ files.
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
		if err := checkedJSON(b, &m); err != nil {
			return nil, fmt.Errorf("dataset.json: %v", err)
		}
		if err := canonicalNames(b, jsonFields(manifest{})); err != nil {
			return nil, fmt.Errorf("dataset.json: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("dataset.json: %v", err)
	}
	explicitGuidelines, explicitGold, explicitResources := m.Guidelines != "", m.Gold != "", len(m.Resources) != 0
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
	d := &Dataset{ID: m.ID, Name: m.Name, Description: m.Description, Source: source, Hashes: map[string]string{}, annotationFiles: map[string]bool{}}
	// A manifest that names no format means its gold's format, or format 2,
	// the built-in guidelines', for a dataset without a gold
	// (docs/annotation-records.md §1).
	d.Format = 2
	if m.Format != "" {
		n, err := formatNumber(m.Format)
		if err != nil {
			return nil, fmt.Errorf("dataset.json: %v", err)
		}
		d.Format = n
	}
	b, err := read(m.Items)
	if err != nil {
		return nil, fmt.Errorf("the claims: %v", err)
	}
	d.setHash(m.Items, b)
	d.annotationFiles[m.Items] = true
	var in struct {
		Discussion json.RawMessage   `json:"discussion"`
		Items      []json.RawMessage `json:"items"`
	}
	if err := checkedJSON(b, &in); err != nil {
		return nil, fmt.Errorf("%s: %v", m.Items, err)
	}
	if err := canonicalNames(b, []string{"discussion", "items"}); err != nil {
		return nil, fmt.Errorf("%s: %v", m.Items, err)
	}
	d.Discussion = in.Discussion
	seen := map[string]bool{}
	for i, raw := range in.Items {
		var it struct {
			ID   string `json:"kialo_claim_id"`
			Text string `json:"text"`
		}
		// The same rule as an annotation file's (parseAnnotations): an ID of
		// spaces is no ID, so a run never refuses claims it was given.
		if err := json.Unmarshal(raw, &it); err != nil || strings.TrimSpace(it.ID) == "" {
			return nil, fmt.Errorf("%s: item %d has no kialo_claim_id", m.Items, i+1)
		}
		if err := canonicalNames(raw, []string{"kialo_claim_id", "text"}); err != nil {
			return nil, fmt.Errorf("%s: item %d: %v", m.Items, i+1, err)
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
		if explicitGuidelines || !os.IsNotExist(err) {
			return nil, fmt.Errorf("guidelines %s: %v", m.Guidelines, err)
		}
		g, err = fs.ReadFile(data.Annotation, "annotation/kialo-13292-batch2/guidelines.md")
		if err != nil {
			return nil, err
		}
		m.Guidelines = "guidelines.md (built in)"
	}
	d.Guidelines = string(g)
	d.setHash(m.Guidelines, g)
	d.annotationFiles[m.Guidelines] = true
	// The gold annotation, if there is one.
	if m.Gold == "" {
		m.Gold = "gold.json"
	}
	d.GoldFile = m.Gold // identity of the scoring input, even when it is absent
	if b, err := read(m.Gold); err == nil {
		format, anns, err := parseAnnotationFile(b)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", m.Gold, err)
		}
		if m.Format == "" {
			d.Format = format
		} else if format != d.Format {
			return nil, fmt.Errorf("%s is in format %d, but the dataset's guidelines ask for format %d: name its format in dataset.json (\"format\": \"wag-annotation/%d\"), or migrate the gold (docs/annotation-records.md)",
				m.Gold, format, d.Format, format)
		}
		d.Gold, d.HasGold = anns, true
		d.setHash(m.Gold, b)
	} else if explicitGold || !os.IsNotExist(err) {
		return nil, fmt.Errorf("gold %s: %v", m.Gold, err)
	}
	// The resources: the scheme collection and the game's NASCQ, and the
	// dataset's own (a lexicon, a case record).
	if len(m.Resources) == 0 {
		m.Resources = []string{"lexicon.json", "case.json"}
	}
	add := func(name string, b []byte) {
		d.Resources = append(d.Resources, Resource{Name: name, Text: string(b)})
		d.setHash(name, b)
		d.annotationFiles[name] = true
	}
	hasWalton := false
	for _, r := range m.Resources {
		b, err := read(r)
		if err != nil {
			if (r == "lexicon.json" || r == "case.json") && !explicitResources && os.IsNotExist(err) {
				continue // optional only when the manifest did not request it
			}
			return nil, fmt.Errorf("resource %s: %v", r, err)
		}
		if path.Base(r) == "walton.yml" {
			hasWalton = true
		}
		if path.Base(r) == "case.json" {
			c, err := parseCase(b)
			if err != nil {
				return nil, fmt.Errorf("the case record %s: %v", r, err)
			}
			d.Case = c
		}
		add(path.Base(r), b)
	}
	if !hasWalton {
		b, err := fs.ReadFile(data.Annotation, "annotation/walton.yml")
		if err != nil {
			return nil, err
		}
		d.Resources = append([]Resource{{Name: "walton.yml", Text: string(b)}}, d.Resources...)
		d.setHash("walton.yml", b)
		d.annotationFiles["walton.yml"] = true
	}
	add("schemes.json", data.SchemesJSON)
	add("critical_questions.json", data.CQsJSON)
	add("edges.json", data.EdgesJSON)
	if err := d.checkCase(); err != nil {
		return nil, fmt.Errorf("the case record: %v", err)
	}
	return d, nil
}

// checkCase checks the case record against the rest of the dataset, as
// validate.py does: its thesis is a claim of the dataset, and its actors,
// actions and aliases are atoms of the lexicon, when there is one.
func (d *Dataset) checkCase() error {
	c := d.Case
	if c == nil {
		return nil
	}
	if d.Item(c.Thesis.Claim) == nil {
		return fmt.Errorf("its thesis %q is not a claim of the dataset", c.Thesis.Claim)
	}
	for _, r := range d.Resources {
		if r.Name != "lexicon.json" {
			continue
		}
		var lex map[string]json.RawMessage
		if err := json.Unmarshal([]byte(r.Text), &lex); err != nil {
			return nil // a lexicon of another shape: the atoms are not checked
		}
		atoms := map[string]string{}
		for _, a := range c.Actors {
			atoms[a.ID] = "actor"
		}
		for _, a := range c.Actions {
			atoms[a.ID] = "action"
			for _, al := range a.Aliases {
				atoms[al] = "alias"
			}
		}
		names := make([]string, 0, len(atoms))
		for k := range atoms {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			if _, ok := lex[k]; !ok {
				return fmt.Errorf("the %s %s is not an atom of the lexicon", atoms[k], k)
			}
		}
	}
	return nil
}

// FormatName is the name of the dataset's annotation format.
func (d *Dataset) FormatName() string {
	if d.Format == 2 {
		return Format2
	}
	return Format1
}

// parseAnnotations reads an annotation file ({"annotator": …, "items": […]}).
func parseAnnotations(b []byte) (map[string]Annotation, error) {
	_, m, err := parseAnnotationFile(b)
	return m, err
}

// parseAnnotationFile reads an annotation file and its format: 2 when it
// says "format": "wag-annotation/2", 1 when it names none (or format 1).
func parseAnnotationFile(b []byte) (int, map[string]Annotation, error) {
	var f struct {
		Format *string         `json:"format"`
		Items  json.RawMessage `json:"items"`
	}
	if err := checkedJSON(b, &f); err != nil {
		return 0, nil, err
	}
	if err := canonicalNames(b, []string{"items", "format"}); err != nil {
		return 0, nil, err
	}
	format := 1
	if f.Format != nil {
		n, err := formatNumber(*f.Format)
		if err != nil || *f.Format == "" {
			return 0, nil, fmt.Errorf("annotation file: its format %q is not %s or %s", *f.Format, Format1, Format2)
		}
		format = n
	}
	if len(f.Items) == 0 {
		return 0, nil, fmt.Errorf("annotation file requires an items array (legacy null means empty)")
	}
	var items []Annotation
	if err := json.Unmarshal(f.Items, &items); err != nil {
		return 0, nil, fmt.Errorf("annotation items: %v", err)
	}
	out := map[string]Annotation{}
	for i, it := range items {
		id, ok := it["kialo_claim_id"].(string)
		if !ok || strings.TrimSpace(id) == "" {
			return 0, nil, fmt.Errorf("annotation item %d needs a nonempty textual kialo_claim_id", i+1)
		}
		if _, exists := out[id]; exists {
			return 0, nil, fmt.Errorf("duplicate annotation claim ID %q", id)
		}
		out[id] = it
	}
	return format, out, nil
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
