package game

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v2"
)

// The game's data edited by its structure, not by its text, so that the
// tests do not depend on how the files are laid out: step C5 generates them
// from the catalogue (docs/c5-game.md, C5.1 step 1).

// coreEdit parses core.yml keeping the order of its keys, applies edit to
// its top-level mapping and writes it again.
func coreEdit(t *testing.T, core []byte, edit func(doc yaml.MapSlice) yaml.MapSlice) []byte {
	t.Helper()
	var doc yaml.MapSlice
	if err := yaml.Unmarshal(core, &doc); err != nil {
		t.Fatal(err)
	}
	out, err := yaml.Marshal(edit(doc))
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// item returns the value of a key of a mapping and its index, or -1.
func item(m yaml.MapSlice, key string) (interface{}, int) {
	for i, it := range m {
		if k, ok := it.Key.(string); ok && k == key {
			return it.Value, i
		}
	}
	return nil, -1
}

// withTemplate sets the language template of a key core.yml has; it fails
// the test if core.yml has no such key.
func withTemplate(t *testing.T, core []byte, key, template string) []byte {
	t.Helper()
	return coreEdit(t, core, func(doc yaml.MapSlice) yaml.MapSlice {
		v, i := item(doc, "language")
		lang, ok := v.(yaml.MapSlice)
		if i < 0 || !ok {
			t.Fatal("core.yml has no language mapping")
		}
		_, j := item(lang, key)
		if j < 0 {
			t.Fatalf("core.yml has no template %s", key)
		}
		lang[j].Value = template
		doc[i].Value = lang
		return doc
	})
}

// addTemplate adds a language template core.yml does not have.
func addTemplate(t *testing.T, core []byte, key, template string) []byte {
	t.Helper()
	return coreEdit(t, core, func(doc yaml.MapSlice) yaml.MapSlice {
		v, i := item(doc, "language")
		lang, ok := v.(yaml.MapSlice)
		if i < 0 || !ok {
			t.Fatal("core.yml has no language mapping")
		}
		if _, j := item(lang, key); j >= 0 {
			t.Fatalf("core.yml has the template %s already", key)
		}
		doc[i].Value = append(lang, yaml.MapItem{Key: key, Value: template})
		return doc
	})
}

// schemes returns core.yml's argument schemes and their index in doc.
func coreSchemes(t *testing.T, doc yaml.MapSlice) ([]interface{}, int) {
	t.Helper()
	v, i := item(doc, "argument_schemes")
	list, ok := v.([]interface{})
	if i < 0 || !ok {
		t.Fatal("core.yml has no argument_schemes list")
	}
	return list, i
}

// withSchemeField sets a field core.yml's scheme id has.
func withSchemeField(t *testing.T, core []byte, id, field string, value interface{}) []byte {
	t.Helper()
	return coreEdit(t, core, func(doc yaml.MapSlice) yaml.MapSlice {
		list, i := coreSchemes(t, doc)
		for n, s := range list {
			m, ok := s.(yaml.MapSlice)
			if !ok {
				continue
			}
			if v, _ := item(m, "id"); v == id {
				_, j := item(m, field)
				if j < 0 {
					t.Fatalf("core.yml's %s has no %s", id, field)
				}
				m[j].Value = value
				list[n] = m
				doc[i].Value = list
				return doc
			}
		}
		t.Fatalf("core.yml has no scheme %s", id)
		return doc
	})
}

// withScheme appends a scheme to core.yml's argument schemes.
func withScheme(t *testing.T, core []byte, scheme yaml.MapSlice) []byte {
	t.Helper()
	return coreEdit(t, core, func(doc yaml.MapSlice) yaml.MapSlice {
		list, i := coreSchemes(t, doc)
		doc[i].Value = append(list, scheme)
		return doc
	})
}

// withBackupEdge puts an edge, written as JSON, first among edges.json's
// backup edges, keeping the other edges and their order.
func withBackupEdge(t *testing.T, edges []byte, edge string) []byte {
	t.Helper()
	var f map[string][]json.RawMessage
	if err := json.Unmarshal(edges, &f); err != nil {
		t.Fatal(err)
	}
	if _, ok := f["backup_edges"]; !ok {
		t.Fatal("edges.json has no backup_edges")
	}
	if !json.Valid([]byte(edge)) {
		t.Fatalf("not JSON: %s", edge)
	}
	f["backup_edges"] = append([]json.RawMessage{json.RawMessage(edge)}, f["backup_edges"]...)
	out, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
